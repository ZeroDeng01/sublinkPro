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

	sparklines, err := GetAirportUsageSparklines([]int{first.ID, second.ID, first.ID}, 14, time.Local)
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

	sparklines, err := getAirportUsageSparklines([]int{withBaseline.ID, noBaseline.ID}, 14, now, time.Local)
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

// TestGetAirportUsageSparklinesIgnoresComponentFlaps 回归：面板偶发交换或缺失上下行字段时，
// used 总量保持单调，旧的分量差值算法会把最后一段重新累计一遍，导致「今日」超过实际消耗。
func TestGetAirportUsageSparklinesIgnoresComponentFlaps(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	airport := newUsageTestAirport(t, "sparkline-component-flap")

	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.Local)
	gib := int64(1) << 30
	snapshots := []AirportUsageSnapshot{
		// 昨天基线：合计 20 GiB。
		{AirportID: airport.ID, Upload: 8 * gib, Download: 12 * gib, Total: 100 * gib, CapturedAt: now.AddDate(0, 0, -1)},
		// 面板把上传/下载互换：used 总量不变，不应该产生消耗。
		{AirportID: airport.ID, Upload: 12 * gib, Download: 8 * gib, Total: 100 * gib, CapturedAt: now.Add(-3 * time.Hour)},
		// 上传字段偶发缺失：used 短暂回落。
		{AirportID: airport.ID, Upload: 0, Download: 12 * gib, Total: 100 * gib, CapturedAt: now.Add(-2 * time.Hour)},
		// 下一次刷新恢复：回落段应被忽略，而不是把近全量重新计一遍。
		{AirportID: airport.ID, Upload: 8 * gib, Download: 12 * gib, Total: 100 * gib, CapturedAt: now.Add(-2*time.Hour + 10*time.Minute)},
		{AirportID: airport.ID, Upload: 8 * gib, Download: 13 * gib, Total: 100 * gib, CapturedAt: now.Add(-time.Hour)},
	}
	if err := database.DB.Create(&snapshots).Error; err != nil {
		t.Fatalf("create sparkline snapshots: %v", err)
	}

	sparklines, err := getAirportUsageSparklines([]int{airport.ID}, 14, now, time.Local)
	if err != nil {
		t.Fatalf("get usage sparklines: %v", err)
	}
	if len(sparklines) != 1 {
		t.Fatalf("sparkline count = %d, want 1", len(sparklines))
	}
	// 实际消耗 = 最终 21 GiB − 基线 20 GiB。
	if got := sparklines[0].TodayUsed; got != gib {
		t.Fatalf("today used = %d, want %d (component flaps must not inflate today)", got, gib)
	}
}

// TestGetAirportUsageSparklinesIgnoresTransientDip 一次成功刷新回报了 0 值时，
// 恢复后不能把近全量计数重新计入当天。
func TestGetAirportUsageSparklinesIgnoresTransientDip(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	airport := newUsageTestAirport(t, "sparkline-transient-dip")

	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.Local)
	gib := int64(1) << 30
	snapshots := []AirportUsageSnapshot{
		{AirportID: airport.ID, Upload: 5 * gib, Download: 10 * gib, Total: 100 * gib, CapturedAt: now.AddDate(0, 0, -1)},
		{AirportID: airport.ID, Upload: 5 * gib, Download: 12 * gib, Total: 100 * gib, CapturedAt: now.Add(-5 * time.Hour)},
		{AirportID: airport.ID, Upload: 0, Download: 0, Total: 100 * gib, CapturedAt: now.Add(-4 * time.Hour)},
		{AirportID: airport.ID, Upload: 5 * gib, Download: 13 * gib, Total: 100 * gib, CapturedAt: now.Add(-3 * time.Hour)},
		{AirportID: airport.ID, Upload: 5 * gib, Download: 15 * gib, Total: 100 * gib, CapturedAt: now.Add(-time.Hour)},
	}
	if err := database.DB.Create(&snapshots).Error; err != nil {
		t.Fatalf("create sparkline snapshots: %v", err)
	}

	sparklines, err := getAirportUsageSparklines([]int{airport.ID}, 14, now, time.Local)
	if err != nil {
		t.Fatalf("get usage sparklines: %v", err)
	}
	// 实际消耗 = 最终 20 GiB − 基线 15 GiB。
	want := 5 * gib
	if got := sparklines[0].TodayUsed; got != want {
		t.Fatalf("today used = %d, want %d (transient dip must be ignored)", got, want)
	}
}

// TestGetAirportUsageSparklinesUsesClientDayBoundary 客户端时区的自然日边界：
// 服务器时区更早的消耗不能被算进客户端「今日」。
func TestGetAirportUsageSparklinesUsesClientDayBoundary(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	airport := newUsageTestAirport(t, "sparkline-client-day")

	clientLoc := time.FixedZone("client", 5*60*60)
	clientDayStart := time.Date(2026, 11, 18, 0, 0, 0, 0, clientLoc)
	now := clientDayStart.Add(12 * time.Hour)
	// 快照按服务器本地时区的文本存储，写入时换成 time.Local 才能与查询边界正确比较。
	snapshots := []AirportUsageSnapshot{
		// 客户端自然日之前的快照：只作为基线。
		{AirportID: airport.ID, Upload: 5, Download: 10, Total: 1000, CapturedAt: clientDayStart.Add(-3 * time.Hour).In(time.Local)},
		{AirportID: airport.ID, Upload: 10, Download: 20, Total: 1000, CapturedAt: clientDayStart.Add(-time.Hour).In(time.Local)},
		// 客户端当天：+4 和 +8。
		{AirportID: airport.ID, Upload: 12, Download: 22, Total: 1000, CapturedAt: clientDayStart.Add(time.Hour).In(time.Local)},
		{AirportID: airport.ID, Upload: 15, Download: 27, Total: 1000, CapturedAt: clientDayStart.Add(2 * time.Hour).In(time.Local)},
	}
	if err := database.DB.Create(&snapshots).Error; err != nil {
		t.Fatalf("create sparkline snapshots: %v", err)
	}

	sparklines, err := getAirportUsageSparklines([]int{airport.ID}, 14, now, clientLoc)
	if err != nil {
		t.Fatalf("get usage sparklines: %v", err)
	}
	if got := sparklines[0].TodayUsed; got != 12 {
		t.Fatalf("today used = %d, want 12 (client timezone day boundary)", got)
	}
}

// TestComputeUsageGrowthIgnoresUnconfirmedSpike 瞬时虚高的取值在没有被下一次刷新确认时不计增长。
func TestComputeUsageGrowthIgnoresUnconfirmedSpike(t *testing.T) {
	base := time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)
	samples := []usageGrowthSample{
		{upload: 10, download: 20},                                      // 起点 30
		{capturedAt: base.Add(time.Hour), upload: 15, download: 21},     // 36
		{capturedAt: base.Add(2 * time.Hour), upload: 44, download: 30}, // 74 瞬时虚高
		{capturedAt: base.Add(3 * time.Hour), upload: 16, download: 22}, // 38 回落
		{capturedAt: base.Add(4 * time.Hour), upload: 16, download: 24}, // 40
	}

	var total int64
	for _, step := range computeUsageGrowth(samples) {
		total += step.upload + step.download
	}
	// 36−30=6；74 被忽略；38−36=2；40−38=2。
	if total != 10 {
		t.Fatalf("growth total = %d, want 10", total)
	}
}

// TestComputeUsageGrowthKeepsPeakOnPersistentMinorDrop 持续回落但仍接近旧高水位时，
// 视为面板异常而不是换套餐，不重置累计。
func TestComputeUsageGrowthKeepsPeakOnPersistentMinorDrop(t *testing.T) {
	base := time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)
	samples := []usageGrowthSample{
		{upload: 30},                   // 起点 30
		{capturedAt: base, upload: 36}, // +6
		{capturedAt: base.Add(8 * time.Hour), upload: 20}, // 55%，且超出异常容忍时间
		{capturedAt: base.Add(20 * time.Hour), upload: 40},
	}

	var total int64
	for _, step := range computeUsageGrowth(samples) {
		total += step.upload + step.download
	}
	// 20 仍接近旧高水位，不重置；恢复后 40−36=4。
	if total != 10 {
		t.Fatalf("growth total = %d, want 10", total)
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

func TestGetAirportDailyUsageBreaksDownByAirport(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	alpha := newUsageTestAirport(t, "daily-alpha")
	beta := newUsageTestAirport(t, "daily-beta")

	now := time.Date(2026, 5, 20, 18, 0, 0, 0, time.Local)
	dayStart := startOfUsageDay(now)
	baselineAt := dayStart.AddDate(0, 0, -1).Add(-time.Hour)
	snapshots := []AirportUsageSnapshot{
		{AirportID: alpha.ID, Upload: 50, Download: 60, Total: 1000, CapturedAt: baselineAt},
		{AirportID: beta.ID, Upload: 10, Download: 20, Total: 1000, CapturedAt: baselineAt},
		{AirportID: alpha.ID, Upload: 150, Download: 260, Total: 1000, CapturedAt: dayStart.Add(2 * time.Hour)},
		{AirportID: beta.ID, Upload: 12, Download: 26, Total: 1000, CapturedAt: dayStart.Add(2 * time.Hour)},
		{AirportID: alpha.ID, Upload: 180, Download: 300, Total: 1000, CapturedAt: dayStart.Add(6 * time.Hour)},
	}
	if err := database.DB.Create(&snapshots).Error; err != nil {
		t.Fatalf("create daily snapshots: %v", err)
	}

	report, err := getAirportDailyUsage(2, 0, now, time.Local)
	if err != nil {
		t.Fatalf("get airport daily usage: %v", err)
	}
	if len(report.Points) != 2 {
		t.Fatalf("daily points = %d, want 2", len(report.Points))
	}

	today := report.Points[len(report.Points)-1]
	if today.Used != 378 {
		t.Fatalf("today used = %d, want 378", today.Used)
	}
	if len(today.Airports) != 2 {
		t.Fatalf("today airport breakdown = %d, want 2", len(today.Airports))
	}
	// 明细按消耗降序：alpha 贡献 370，beta 贡献 8。
	if today.Airports[0].AirportName != "daily-alpha" || today.Airports[0].Used != 370 {
		t.Fatalf("first breakdown entry = %+v, want daily-alpha with 370", today.Airports[0])
	}
	if today.Airports[0].Upload != 130 || today.Airports[0].Download != 240 {
		t.Fatalf("alpha breakdown upload/download = %d/%d, want 130/240", today.Airports[0].Upload, today.Airports[0].Download)
	}
	if today.Airports[1].AirportName != "daily-beta" || today.Airports[1].Used != 8 {
		t.Fatalf("second breakdown entry = %+v, want daily-beta with 8", today.Airports[1])
	}
	if len(report.Points[0].Airports) != 0 {
		t.Fatalf("yesterday breakdown = %d, want 0", len(report.Points[0].Airports))
	}

	// 单机场报告不需要拆分明细。
	single, err := getAirportDailyUsage(2, alpha.ID, now, time.Local)
	if err != nil {
		t.Fatalf("get single airport daily usage: %v", err)
	}
	if last := single.Points[len(single.Points)-1]; len(last.Airports) != 0 {
		t.Fatalf("single airport breakdown = %d, want 0", len(last.Airports))
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
