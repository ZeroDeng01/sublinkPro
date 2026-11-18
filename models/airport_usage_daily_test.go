package models

import (
	"testing"
	"time"

	"sublink/database"
)

// addUsageSnapshot 直接写入一条用量快照，便于构造跨天的历史数据。
func addUsageSnapshot(t *testing.T, airportID int, uploaded, downloaded int64, capturedAt time.Time, failed bool) {
	t.Helper()
	snapshot := AirportUsageSnapshot{
		AirportID:  airportID,
		Upload:     uploaded,
		Download:   downloaded,
		Failed:     failed,
		CapturedAt: capturedAt,
	}
	if err := database.DB.Create(&snapshot).Error; err != nil {
		t.Fatalf("create usage snapshot: %v", err)
	}
}

// dailyPointByDate 取出指定日期的统计点，不存在时直接失败。
func dailyPointByDate(t *testing.T, report *AirportUsageDailyReport, date string) UsageDailyPoint {
	t.Helper()
	for _, point := range report.Points {
		if point.Date == date {
			return point
		}
	}
	t.Fatalf("date %s not found in report (points=%d)", date, len(report.Points))
	return UsageDailyPoint{}
}

func TestGetAirportDailyUsageComputesDailyDeltas(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.Local)
	airportA := newUsageTestAirport(t, "daily-delta-a")
	airportB := newUsageTestAirport(t, "daily-delta-b")

	// A：窗口前的基线（from = 11-16 00:00），用于计算第一天之前的差值。
	addUsageSnapshot(t, airportA.ID, 50, 60, time.Date(2026, 11, 15, 9, 0, 0, 0, time.Local), false)
	addUsageSnapshot(t, airportA.ID, 150, 260, time.Date(2026, 11, 16, 8, 0, 0, 0, time.Local), false)
	addUsageSnapshot(t, airportA.ID, 200, 320, time.Date(2026, 11, 17, 9, 0, 0, 0, time.Local), false)

	// B：独立计数，同一天发生消耗，验证多机场汇总。
	addUsageSnapshot(t, airportB.ID, 10, 10, time.Date(2026, 11, 15, 10, 0, 0, 0, time.Local), false)
	addUsageSnapshot(t, airportB.ID, 40, 20, time.Date(2026, 11, 17, 7, 0, 0, 0, time.Local), false)

	report, err := getAirportDailyUsage(2, 0, now, time.Local)
	if err != nil {
		t.Fatalf("daily usage: %v", err)
	}
	if len(report.Points) != 2 {
		t.Fatalf("points = %d, want 2 (%+v)", len(report.Points), report.Points)
	}

	if point := dailyPointByDate(t, report, "2026-11-16"); point.Used != 300 {
		t.Fatalf("11-16 used = %d, want 300", point.Used)
	}
	today := dailyPointByDate(t, report, "2026-11-17")
	if today.Upload != 80 || today.Download != 70 || today.Used != 150 {
		t.Fatalf("11-17 point = %+v, want upload=80 download=70 used=150", today)
	}
	if report.TodayUsed != 150 {
		t.Fatalf("todayUsed = %d, want 150", report.TodayUsed)
	}
	if report.RangeUsed != 450 {
		t.Fatalf("rangeUsed = %d, want 450", report.RangeUsed)
	}
}

func TestGetAirportDailyUsageHandlesCounterReset(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.Local)
	airport := newUsageTestAirport(t, "daily-counter-reset")

	addUsageSnapshot(t, airport.ID, 1000, 2000, time.Date(2026, 11, 15, 9, 0, 0, 0, time.Local), false)
	addUsageSnapshot(t, airport.ID, 1000, 2000, time.Date(2026, 11, 16, 9, 0, 0, 0, time.Local), false)
	// 面板重置/换套餐：累计值变小，按新值全量计入当天，而不是出现负数。
	addUsageSnapshot(t, airport.ID, 100, 200, time.Date(2026, 11, 17, 9, 0, 0, 0, time.Local), false)

	report, err := getAirportDailyUsage(2, airport.ID, now, time.Local)
	if err != nil {
		t.Fatalf("daily usage: %v", err)
	}
	if point := dailyPointByDate(t, report, "2026-11-16"); point.Used != 0 {
		t.Fatalf("11-16 used = %d, want 0", point.Used)
	}
	if point := dailyPointByDate(t, report, "2026-11-17"); point.Used != 300 {
		t.Fatalf("11-17 used = %d, want 300 (reset value counted in full)", point.Used)
	}
}

func TestGetAirportDailyUsageSkipsFailedSnapshots(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.Local)
	airport := newUsageTestAirport(t, "daily-failed-skipped")

	addUsageSnapshot(t, airport.ID, 100, 100, time.Date(2026, 11, 15, 9, 0, 0, 0, time.Local), false)
	// 失败快照会沿用上一次的数值，不能被当成一次消耗。
	addUsageSnapshot(t, airport.ID, 999999, 999999, time.Date(2026, 11, 16, 9, 0, 0, 0, time.Local), true)
	addUsageSnapshot(t, airport.ID, 300, 300, time.Date(2026, 11, 17, 9, 0, 0, 0, time.Local), false)

	report, err := getAirportDailyUsage(2, airport.ID, now, time.Local)
	if err != nil {
		t.Fatalf("daily usage: %v", err)
	}
	if point := dailyPointByDate(t, report, "2026-11-16"); point.Used != 0 {
		t.Fatalf("11-16 used = %d, want 0 (failed snapshots ignored)", point.Used)
	}
	if point := dailyPointByDate(t, report, "2026-11-17"); point.Used != 400 {
		t.Fatalf("11-17 used = %d, want 400", point.Used)
	}
}

func TestGetAirportDailyUsageFiltersSingleAirport(t *testing.T) {
	setupUsageSnapshotTestDB(t)
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.Local)
	airportA := newUsageTestAirport(t, "daily-filter-a")
	airportB := newUsageTestAirport(t, "daily-filter-b")

	addUsageSnapshot(t, airportA.ID, 0, 0, time.Date(2026, 11, 16, 9, 0, 0, 0, time.Local), false)
	addUsageSnapshot(t, airportA.ID, 100, 10, time.Date(2026, 11, 17, 9, 0, 0, 0, time.Local), false)
	addUsageSnapshot(t, airportB.ID, 0, 0, time.Date(2026, 11, 16, 9, 0, 0, 0, time.Local), false)
	addUsageSnapshot(t, airportB.ID, 500, 500, time.Date(2026, 11, 17, 9, 0, 0, 0, time.Local), false)

	report, err := getAirportDailyUsage(2, airportA.ID, now, time.Local)
	if err != nil {
		t.Fatalf("daily usage: %v", err)
	}
	if report.AirportID != airportA.ID {
		t.Fatalf("airportID = %d, want %d", report.AirportID, airportA.ID)
	}
	if report.TodayUsed != 110 {
		t.Fatalf("todayUsed = %d, want 110 (only airport A)", report.TodayUsed)
	}
}
