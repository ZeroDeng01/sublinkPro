package models

import (
	"strconv"
	"testing"
	"time"

	"sublink/database"
	"sublink/internal/testutil"
)

func setupUsageSnapshotTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db := testutil.OpenMemoryDB(t, "usage_snapshot_test")
	if err := db.AutoMigrate(
		&Airport{},
		&AirportUsageSnapshot{},
		&Subcription{},
		&SubLogs{},
		&SubcriptionAirport{},
		&SubcriptionNode{},
		&SubcriptionGroup{},
		&Node{},
		&SystemSetting{},
	); err != nil {
		t.Fatalf("auto migrate usage snapshot tables: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false

	resetAirportCacheForTest()
	resetSubcriptionCacheForTest()
	lastUsageSnapshotPruneAt.Store(0)
	if err := InitSettingCache(); err != nil {
		t.Fatalf("init setting cache: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Exec("DELETE FROM system_settings").Error
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		lastUsageSnapshotPruneAt.Store(0)
		resetAirportCacheForTest()
		resetSubcriptionCacheForTest()
		if oldDB != nil {
			_ = InitSettingCache()
		}
		testutil.CloseDB(t, db)
	})
}

func newUsageTestAirport(t *testing.T, name string) *Airport {
	t.Helper()
	airport := &Airport{
		Name:           name,
		URL:            "https://example.com/" + name,
		CronExpr:       "0 */12 * * *",
		FetchUsageInfo: true,
	}
	if err := airport.Add(); err != nil {
		t.Fatalf("add airport %s: %v", name, err)
	}
	return airport
}

func countUsageSnapshots(t *testing.T, airportID int) int64 {
	t.Helper()
	var count int64
	if err := database.DB.Model(&AirportUsageSnapshot{}).Where("airport_id = ?", airportID).Count(&count).Error; err != nil {
		t.Fatalf("count usage snapshots: %v", err)
	}
	return count
}

func TestUpdateUsageInfoRecordsEveryRefresh(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	airport := newUsageTestAirport(t, "snapshot-every-refresh")

	if err := airport.UpdateUsageInfo(100, 200, 1000, 0); err != nil {
		t.Fatalf("update usage info: %v", err)
	}
	if got := countUsageSnapshots(t, airport.ID); got != 1 {
		t.Fatalf("snapshot count after first update = %d, want 1", got)
	}

	// 数值未变化也必须落库：每个请求都要在历史中留下一个点。
	if err := airport.UpdateUsageInfo(100, 200, 1000, 0); err != nil {
		t.Fatalf("repeat update usage info: %v", err)
	}
	if got := countUsageSnapshots(t, airport.ID); got != 2 {
		t.Fatalf("snapshot count after identical update = %d, want 2", got)
	}

	if err := airport.UpdateUsageInfo(150, 250, 1000, 0); err != nil {
		t.Fatalf("update changed usage info: %v", err)
	}
	if got := countUsageSnapshots(t, airport.ID); got != 3 {
		t.Fatalf("snapshot count after changed update = %d, want 3", got)
	}
}

func TestRecordUsageFailureWritesEveryAttempt(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	airport := newUsageTestAirport(t, "snapshot-failure-every-attempt")

	if err := airport.UpdateUsageInfo(10, 20, 100, 0); err != nil {
		t.Fatalf("update usage info: %v", err)
	}

	// 相同原因的重复失败同样逐次记录，红色标记数量与真实请求次数一致。
	airport.RecordUsageFailure("fetch failed")
	airport.RecordUsageFailure("fetch failed")

	if got := countUsageSnapshots(t, airport.ID); got != 3 {
		t.Fatalf("snapshot count after repeated failures = %d, want 3", got)
	}
}

func TestUpdateUsageInfoFailureCarriesForwardValues(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	airport := newUsageTestAirport(t, "snapshot-failure")

	if err := airport.UpdateUsageInfo(300, 400, 5000, 1700000000); err != nil {
		t.Fatalf("update usage info: %v", err)
	}

	if err := airport.UpdateUsageInfoWithStatus(0, 0, -1, 0, true, "节点解析失败"); err != nil {
		t.Fatalf("record usage failure: %v", err)
	}

	var failure AirportUsageSnapshot
	if err := database.DB.Where("airport_id = ? AND failed = ?", airport.ID, true).
		Order("captured_at DESC").First(&failure).Error; err != nil {
		t.Fatalf("load failure snapshot: %v", err)
	}
	if failure.Upload != 300 || failure.Download != 400 || failure.Total != 5000 || failure.Expire != 1700000000 {
		t.Fatalf("failure snapshot did not carry forward values: %+v", failure)
	}
	if failure.Error != "节点解析失败" {
		t.Fatalf("failure snapshot error = %q", failure.Error)
	}
}

func TestPruneAirportUsageSnapshotsRemovesExpiredRows(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	airport := newUsageTestAirport(t, "snapshot-prune")

	oldSnapshot := AirportUsageSnapshot{
		AirportID:  airport.ID,
		Upload:     1,
		Download:   2,
		CapturedAt: time.Now().AddDate(0, 0, -30),
	}
	recentSnapshot := AirportUsageSnapshot{
		AirportID:  airport.ID,
		Upload:     3,
		Download:   4,
		CapturedAt: time.Now().AddDate(0, 0, -1),
	}
	if err := database.DB.Create(&oldSnapshot).Error; err != nil {
		t.Fatalf("create old snapshot: %v", err)
	}
	if err := database.DB.Create(&recentSnapshot).Error; err != nil {
		t.Fatalf("create recent snapshot: %v", err)
	}

	if err := PruneAirportUsageSnapshots(7); err != nil {
		t.Fatalf("prune usage snapshots: %v", err)
	}
	if got := countUsageSnapshots(t, airport.ID); got != 1 {
		t.Fatalf("snapshot count after prune = %d, want 1", got)
	}
}

func TestSetAirportUsageRetentionDaysValidatesAndPrunes(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	airport := newUsageTestAirport(t, "retention-setting")

	if err := SetAirportUsageRetentionDays(0); err == nil {
		t.Fatal("expected error for retention days 0")
	}
	if err := SetAirportUsageRetentionDays(MaxAirportUsageRetentionDays + 1); err == nil {
		t.Fatal("expected error for retention days above max")
	}

	oldSnapshot := AirportUsageSnapshot{AirportID: airport.ID, CapturedAt: time.Now().AddDate(0, 0, -10)}
	if err := database.DB.Create(&oldSnapshot).Error; err != nil {
		t.Fatalf("create old snapshot: %v", err)
	}

	if err := SetAirportUsageRetentionDays(3); err != nil {
		t.Fatalf("set retention days: %v", err)
	}
	if got := GetAirportUsageRetentionDays(); got != 3 {
		t.Fatalf("retention days = %d, want 3", got)
	}
	if got := countUsageSnapshots(t, airport.ID); got != 0 {
		t.Fatalf("snapshot count after retention prune = %d, want 0", got)
	}
}

func TestGetAirportUsageHistoryMarksFailuresAndCarriesValues(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	airport := newUsageTestAirport(t, "history-airport")

	dayAt := func(offset int) time.Time {
		value := time.Now().In(time.Local).AddDate(0, 0, offset)
		return time.Date(value.Year(), value.Month(), value.Day(), 12, 0, 0, 0, time.Local)
	}
	snapshots := []AirportUsageSnapshot{
		{AirportID: airport.ID, Upload: 40, Download: 60, Total: 1000, CapturedAt: dayAt(-2)},
		{AirportID: airport.ID, Upload: 40, Download: 60, Total: 1000, Failed: true, Error: "refresh failed", CapturedAt: dayAt(-1)},
		{AirportID: airport.ID, Upload: 100, Download: 200, Total: 1000, CapturedAt: dayAt(0)},
	}
	if err := database.DB.Create(&snapshots).Error; err != nil {
		t.Fatalf("create history snapshots: %v", err)
	}

	history, err := GetAirportUsageHistory(airport.ID, 30, UsageGranularityDay)
	if err != nil {
		t.Fatalf("get airport usage history: %v", err)
	}
	if len(history.Points) != 3 {
		t.Fatalf("history points = %d, want 3", len(history.Points))
	}
	if history.Points[0].Used != 100 {
		t.Fatalf("first point used = %d, want 100", history.Points[0].Used)
	}
	if !history.Points[1].Failed {
		t.Fatal("expected middle point to be marked as failed")
	}
	if history.Points[1].Used != 100 {
		t.Fatalf("failed point used = %d, want carried forward 100", history.Points[1].Used)
	}
	if history.Points[2].Used != 300 {
		t.Fatalf("last point used = %d, want 300", history.Points[2].Used)
	}
}

func TestGetAirportUsageSparklinesReturnsPointsForEachAirport(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	first := newUsageTestAirport(t, "sparkline-one")
	second := newUsageTestAirport(t, "sparkline-two")

	snapshots := []AirportUsageSnapshot{
		{AirportID: first.ID, Upload: 10, Download: 20, Total: 100, CapturedAt: time.Now().Add(-2 * time.Hour)},
		{AirportID: first.ID, Upload: 20, Download: 30, Total: 100, CapturedAt: time.Now().Add(-1 * time.Hour)},
		{AirportID: second.ID, Upload: 5, Download: 5, Total: 50, CapturedAt: time.Now().Add(-1 * time.Hour)},
	}
	if err := database.DB.Create(&snapshots).Error; err != nil {
		t.Fatalf("create sparkline snapshots: %v", err)
	}

	sparklines, err := GetAirportUsageSparklines([]int{first.ID, second.ID, first.ID}, 14)
	if err != nil {
		t.Fatalf("get usage sparklines: %v", err)
	}
	if len(sparklines) != 2 {
		t.Fatalf("sparkline count = %d, want 2", len(sparklines))
	}
	for _, sparkline := range sparklines {
		if len(sparkline.Points) == 0 {
			t.Fatalf("airport %d sparkline has no points", sparkline.AirportID)
		}
		if sparkline.UpdatedAt <= 0 {
			t.Fatalf("airport %d sparkline updatedAt = %d", sparkline.AirportID, sparkline.UpdatedAt)
		}
	}
}

func TestGetAirportUsageSparklinesReportsTodayUsed(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	withBaseline := newUsageTestAirport(t, "sparkline-today-baseline")
	noBaseline := newUsageTestAirport(t, "sparkline-today-no-baseline")

	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.Local)
	snapshots := []AirportUsageSnapshot{
		// 当天零点前的最后一条成功快照作为基线
		{AirportID: withBaseline.ID, Upload: 100, Download: 200, Total: 1000, CapturedAt: now.AddDate(0, 0, -1)},
		{AirportID: withBaseline.ID, Upload: 100, Download: 250, Total: 1000, CapturedAt: now.Add(-2 * time.Hour)},
		{AirportID: withBaseline.ID, Upload: 130, Download: 260, Total: 1000, CapturedAt: now.Add(-time.Hour)},
		// 失败快照不参与今天消耗的计算
		{AirportID: withBaseline.ID, Upload: 999, Download: 999, Total: 1000, Failed: true, CapturedAt: now.Add(-30 * time.Minute)},
		// 没有历史基线：当天第一条成功快照只作为基线，从第二条开始累计
		{AirportID: noBaseline.ID, Upload: 1, Download: 1, Total: 10, CapturedAt: now.Add(-3 * time.Hour)},
		{AirportID: noBaseline.ID, Upload: 4, Download: 3, Total: 10, CapturedAt: now.Add(-time.Hour)},
	}
	if err := database.DB.Create(&snapshots).Error; err != nil {
		t.Fatalf("create sparkline snapshots: %v", err)
	}

	sparklines, err := getAirportUsageSparklines([]int{withBaseline.ID, noBaseline.ID}, 14, now)
	if err != nil {
		t.Fatalf("get usage sparklines: %v", err)
	}

	todayUsed := make(map[int]int64, len(sparklines))
	for _, sparkline := range sparklines {
		todayUsed[sparkline.AirportID] = sparkline.TodayUsed
	}
	// 昨天基线(100,200) -> 10:00(100,250) 消耗 50，11:00(130,260) 再消耗 40
	if got := todayUsed[withBaseline.ID]; got != 90 {
		t.Fatalf("today used with baseline = %d, want 90", got)
	}
	// (1,1) -> (4,3)：上行 +3、下行 +2
	if got := todayUsed[noBaseline.ID]; got != 5 {
		t.Fatalf("today used without baseline = %d, want 5", got)
	}
}

func TestGetAirportUsageHistoryAutoKeepsRefreshPoints(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	airport := newUsageTestAirport(t, "history-refresh")

	base := time.Now().Add(-30 * time.Minute).Truncate(time.Second)
	snapshots := []AirportUsageSnapshot{
		{AirportID: airport.ID, Upload: 10, Download: 20, Total: 1000, CapturedAt: base},
		{AirportID: airport.ID, Upload: 20, Download: 30, Total: 1000, CapturedAt: base.Add(5 * time.Minute)},
		{AirportID: airport.ID, Upload: 30, Download: 40, Total: 1000, CapturedAt: base.Add(10 * time.Minute)},
	}
	if err := database.DB.Create(&snapshots).Error; err != nil {
		t.Fatalf("create refresh snapshots: %v", err)
	}

	history, err := GetAirportUsageHistory(airport.ID, 30, UsageGranularityAuto)
	if err != nil {
		t.Fatalf("get airport usage history: %v", err)
	}
	if history.Granularity != UsageGranularityRefresh {
		t.Fatalf("granularity = %s, want refresh", history.Granularity)
	}
	if len(history.Points) != 3 {
		t.Fatalf("history points = %d, want 3", len(history.Points))
	}
	if history.Points[1].Timestamp == history.Points[0].Timestamp {
		t.Fatal("expected distinct timestamps for refreshes within the same hour")
	}
	if history.Points[2].Used != 70 {
		t.Fatalf("last point used = %d, want 70", history.Points[2].Used)
	}
}

func TestGetAirportUsageHistoryAutoCapsDensePoints(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	airport := newUsageTestAirport(t, "history-dense")

	base := time.Now().Add(-time.Duration(maxUsageHistoryPoints) * time.Minute)
	batch := make([]AirportUsageSnapshot, 0, maxUsageHistoryPoints+1)
	for i := 0; i <= maxUsageHistoryPoints; i++ {
		batch = append(batch, AirportUsageSnapshot{
			AirportID:  airport.ID,
			Upload:     int64(i),
			Total:      1 << 40,
			CapturedAt: base.Add(time.Duration(i) * time.Minute),
		})
	}
	if err := database.DB.CreateInBatches(&batch, 500).Error; err != nil {
		t.Fatalf("create dense snapshots: %v", err)
	}

	history, err := GetAirportUsageHistory(airport.ID, 30, UsageGranularityAuto)
	if err != nil {
		t.Fatalf("get airport usage history: %v", err)
	}
	if history.Granularity != UsageGranularityRefresh {
		t.Fatalf("granularity = %s, want refresh", history.Granularity)
	}
	if len(history.Points) > maxUsageHistoryPoints {
		t.Fatalf("history points = %d, want at most %d", len(history.Points), maxUsageHistoryPoints)
	}
	last := history.Points[len(history.Points)-1]
	if last.Timestamp != batch[len(batch)-1].CapturedAt.UnixMilli() {
		t.Fatalf("last point timestamp = %d, want %d", last.Timestamp, batch[len(batch)-1].CapturedAt.UnixMilli())
	}
}

func TestResolveUsageGranularityDefaultsToRefresh(t *testing.T) {
	for _, value := range []string{"", UsageGranularityAuto, UsageGranularityRefresh, "unexpected"} {
		if got := resolveUsageGranularity(value); got != UsageGranularityRefresh {
			t.Fatalf("granularity %q = %s, want refresh", value, got)
		}
	}
	if got := resolveUsageGranularity(UsageGranularityHour); got != UsageGranularityHour {
		t.Fatalf("granularity hour = %s, want hour", got)
	}
	if got := resolveUsageGranularity(UsageGranularityDay); got != UsageGranularityDay {
		t.Fatalf("granularity day = %s, want day", got)
	}
}

func TestDownsampleUsagePointsKeepsFailures(t *testing.T) {
	points := make([]UsageHistoryPoint, 0, 200)
	for i := 0; i < 200; i++ {
		points = append(points, UsageHistoryPoint{Timestamp: int64(i), Used: int64(i)})
	}
	points[10].Failed = true
	points[10].Error = "boom"

	sampled := downsampleUsagePoints(points, 40)
	if len(sampled) != 40 {
		t.Fatalf("downsampled points = %d, want 40", len(sampled))
	}
	failed := 0
	for _, point := range sampled {
		if point.Failed {
			failed++
		}
	}
	if failed != 1 {
		t.Fatalf("downsampled failed points = %d, want 1", failed)
	}
	if last := sampled[len(sampled)-1]; last.Timestamp != points[len(points)-1].Timestamp {
		t.Fatalf("downsampled last timestamp = %d, want %d", last.Timestamp, points[len(points)-1].Timestamp)
	}
}

func TestNormalizeUsageHistoryParams(t *testing.T) {
	if got := NormalizeUsageHistoryDays(0, 90); got != 90 {
		t.Fatalf("days 0 normalized to %d, want retention 90", got)
	}
	if got := NormalizeUsageHistoryDays(30, 90); got != 30 {
		t.Fatalf("days 30 normalized to %d, want 30", got)
	}
	if got := NormalizeUsageHistoryDays(200, 90); got != 90 {
		t.Fatalf("days above retention normalized to %d, want 90", got)
	}
}

func TestGetAirportUsageRetentionDaysFallsBackOnInvalidValue(t *testing.T) {
	setupUsageSnapshotTestDB(t)

	if err := SetSetting(AirportUsageRetentionSettingKey, "not-a-number"); err != nil {
		t.Fatalf("set invalid retention setting: %v", err)
	}
	if got := GetAirportUsageRetentionDays(); got != DefaultAirportUsageRetentionDays {
		t.Fatalf("retention days = %d, want default %d", got, DefaultAirportUsageRetentionDays)
	}

	if err := SetSetting(AirportUsageRetentionSettingKey, strconv.Itoa(MaxAirportUsageRetentionDays+100)); err != nil {
		t.Fatalf("set out of range retention setting: %v", err)
	}
	if got := GetAirportUsageRetentionDays(); got != DefaultAirportUsageRetentionDays {
		t.Fatalf("retention days = %d, want default %d", got, DefaultAirportUsageRetentionDays)
	}
}
