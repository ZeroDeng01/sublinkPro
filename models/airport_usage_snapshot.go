package models

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"sublink/database"
	"sublink/utils"
)

// 用量历史相关常量
const (
	// AirportUsageRetentionSettingKey 用量历史快照保留天数的系统设置键
	AirportUsageRetentionSettingKey = "airport_usage_retention_days"
	// DefaultAirportUsageRetentionDays 默认保留天数
	DefaultAirportUsageRetentionDays = 180
	// MinAirportUsageRetentionDays 允许配置的最小保留天数
	MinAirportUsageRetentionDays = 1
	// MaxAirportUsageRetentionDays 允许配置的最大保留天数
	MaxAirportUsageRetentionDays = 3650

	// UsageGranularityAuto 自动选择聚合粒度
	UsageGranularityAuto = "auto"
	// UsageGranularityRefresh 每次刷新各占一个图表点
	UsageGranularityRefresh = "refresh"
	// UsageGranularityHour 按小时聚合
	UsageGranularityHour = "hour"
	// UsageGranularityDay 按天聚合
	UsageGranularityDay = "day"

	// DefaultUsageHistoryDays 图表默认展示天数
	DefaultUsageHistoryDays = 30
	// DefaultUsageSparklineDays 列表迷你图默认展示天数
	DefaultUsageSparklineDays = 14
	// maxUsageHistoryPoints 历史图表返回的点数上限，超出时自动降级聚合粒度
	maxUsageHistoryPoints = 3000
	// maxUsageSparklinePoints 迷你图单机场的点数上限，超出时均匀降采样
	maxUsageSparklinePoints = 60
	// maxUsageSparklineAirports 单次批量迷你图请求允许的机场数量上限
	maxUsageSparklineAirports = 500
)

var (
	// usageSnapshotPruneInterval 自动清理保留期外快照的最小间隔
	usageSnapshotPruneInterval = time.Hour

	// lastUsageSnapshotPruneAt 上次自动清理时间（UnixNano）
	lastUsageSnapshotPruneAt atomic.Int64
)

// AirportUsageSnapshot 机场用量历史快照
// 每次成功刷新写入一条记录；刷新失败时额外写入一条 Failed 标记，
// 并沿用最近一次成功快照的数值，保证图表曲线在失败点处保持连续。
type AirportUsageSnapshot struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	AirportID  int       `gorm:"index:idx_airport_usage_snapshot,priority:1" json:"airportId"`
	Upload     int64     `gorm:"default:0" json:"upload"`
	Download   int64     `gorm:"default:0" json:"download"`
	Total      int64     `gorm:"default:0" json:"total"`
	Expire     int64     `gorm:"default:0" json:"expire"`
	Failed     bool      `gorm:"default:false" json:"failed"`
	Error      string    `gorm:"size:512" json:"error,omitempty"`
	CapturedAt time.Time `gorm:"index:idx_airport_usage_snapshot,priority:2" json:"capturedAt"`
}

// TableName 指定表名
func (AirportUsageSnapshot) TableName() string {
	return "airport_usage_snapshots"
}

// UsageHistoryPoint 用量历史图表数据点
type UsageHistoryPoint struct {
	Timestamp  int64     `json:"timestamp"` // 毫秒时间戳，便于前端图表直接使用
	CapturedAt time.Time `json:"capturedAt"`
	Upload     int64     `json:"upload"`
	Download   int64     `json:"download"`
	Used       int64     `json:"used"` // Upload + Download
	Total      int64     `json:"total"`
	Expire     int64     `json:"expire"`
	Failed     bool      `json:"failed"`
	Error      string    `json:"error,omitempty"`
}

// AirportUsageHistory 单个机场的用量历史
type AirportUsageHistory struct {
	AirportID       int                 `json:"airportId"`
	AirportName     string              `json:"airportName"`
	Days            int                 `json:"days"`
	Granularity     string              `json:"granularity"`
	From            time.Time           `json:"from"`
	To              time.Time           `json:"to"`
	RetentionDays   int                 `json:"retentionDays"`
	CurrentUpload   int64               `json:"currentUpload"`
	CurrentDownload int64               `json:"currentDownload"`
	CurrentUsed     int64               `json:"currentUsed"`
	CurrentTotal    int64               `json:"currentTotal"`
	CurrentExpire   int64               `json:"currentExpire"`
	Points          []UsageHistoryPoint `json:"points"`
}

// AirportUsageSparkline 机场用量迷你图
type AirportUsageSparkline struct {
	AirportID int                 `json:"airportId"`
	Points    []UsageHistoryPoint `json:"points"`
	Failed    bool                `json:"failed"`    // 最近一次刷新是否失败
	UpdatedAt int64               `json:"updatedAt"` // 最近一次快照时间（毫秒）
	TodayUsed int64               `json:"todayUsed"` // 本地时区当天已消耗流量（快照差值累计）
}

// usageValues 可继承的用量数值
type usageValues struct {
	upload   int64
	download int64
	total    int64
	expire   int64
}

// usageBucket 单个聚合时间桶
type usageBucket struct {
	key       time.Time
	sampledAt time.Time
	values    usageValues
	hasValue  bool
	failed    bool
	errorText string
}

// GetAirportUsageRetentionDays 返回配置的保留天数，非法值回落到默认值。
func GetAirportUsageRetentionDays() int {
	raw, err := GetSetting(AirportUsageRetentionSettingKey)
	if err != nil {
		return DefaultAirportUsageRetentionDays
	}
	days, convErr := strconv.Atoi(strings.TrimSpace(raw))
	if convErr != nil || days < MinAirportUsageRetentionDays || days > MaxAirportUsageRetentionDays {
		return DefaultAirportUsageRetentionDays
	}
	return days
}

// SetAirportUsageRetentionDays 保存保留天数并立即清理超出保留期的历史。
func SetAirportUsageRetentionDays(days int) error {
	if days < MinAirportUsageRetentionDays || days > MaxAirportUsageRetentionDays {
		return fmt.Errorf("retention days must be between %d and %d", MinAirportUsageRetentionDays, MaxAirportUsageRetentionDays)
	}
	if err := SetSetting(AirportUsageRetentionSettingKey, strconv.Itoa(days)); err != nil {
		return err
	}
	return PruneAirportUsageSnapshots(days)
}

// PruneAirportUsageSnapshots 删除早于保留期的用量快照。
func PruneAirportUsageSnapshots(retentionDays int) error {
	if database.DB == nil || retentionDays <= 0 {
		return nil
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return database.DB.Where("captured_at < ?", cutoff).Delete(&AirportUsageSnapshot{}).Error
}

// CountAirportUsageSnapshots 统计当前保存的用量快照数量。
func CountAirportUsageSnapshots() (int64, error) {
	if database.DB == nil {
		return 0, nil
	}
	var count int64
	err := database.DB.Model(&AirportUsageSnapshot{}).Count(&count).Error
	return count, err
}

// RecordAirportUsageSnapshot 记录一次用量快照。
// 每次刷新尝试都会写入一条记录，即使数值与上一条完全相同，保证每个请求都在历史中留痕。
// failed=true 时写入失败标记；当 total<0（没有可用数值）时沿用上一条快照的数值，
// 保证历史曲线在失败点处保持连续。
func RecordAirportUsageSnapshot(airportID int, upload, download, total, expire int64, failed bool, reason string) {
	if database.DB == nil || airportID <= 0 {
		return
	}

	now := time.Now()
	reason = strings.TrimSpace(reason)
	if len(reason) > 512 {
		reason = reason[:512]
	}

	var last AirportUsageSnapshot
	hasLast := database.DB.Where("airport_id = ?", airportID).
		Order("captured_at DESC").Order("id DESC").First(&last).Error == nil

	if failed && total < 0 {
		if hasLast {
			upload, download, total, expire = last.Upload, last.Download, last.Total, last.Expire
		} else {
			upload, download, total, expire = 0, 0, 0, 0
		}
	}

	snapshot := AirportUsageSnapshot{
		AirportID:  airportID,
		Upload:     upload,
		Download:   download,
		Total:      total,
		Expire:     expire,
		Failed:     failed,
		Error:      reason,
		CapturedAt: now,
	}
	if err := database.DB.Create(&snapshot).Error; err != nil {
		utils.Warn("Failed to record airport usage snapshot ID %d: %v", airportID, err)
		return
	}
	maybePruneUsageSnapshots()
}

// RecordAirportUsageFailure 记录一次没有可用数值的刷新失败。
func RecordAirportUsageFailure(airportID int, reason string) {
	RecordAirportUsageSnapshot(airportID, 0, 0, -1, 0, true, reason)
}

// maybePruneUsageSnapshots 按保留期清理历史，最多每小时执行一次。
func maybePruneUsageSnapshots() {
	now := time.Now().UnixNano()
	previous := lastUsageSnapshotPruneAt.Load()
	if previous != 0 && now-previous < int64(usageSnapshotPruneInterval) {
		return
	}
	if !lastUsageSnapshotPruneAt.CompareAndSwap(previous, now) {
		return
	}
	if err := PruneAirportUsageSnapshots(GetAirportUsageRetentionDays()); err != nil {
		utils.Warn("Failed to clean up airport usage history: %v", err)
	}
}

// NormalizeUsageHistoryDays 归一化查询天数：0 表示全部保留期，超出保留期按保留期处理。
func NormalizeUsageHistoryDays(days, retentionDays int) int {
	if retentionDays <= 0 {
		retentionDays = DefaultAirportUsageRetentionDays
	}
	if days <= 0 || days > retentionDays {
		return retentionDays
	}
	return days
}

// resolveUsageGranularity 选择聚合粒度：显式请求按小时/按天聚合，
// 其余情况（auto / refresh）都按刷新记录逐点绘制，点数过多时再由调用方降采样。
func resolveUsageGranularity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case UsageGranularityHour:
		return UsageGranularityHour
	case UsageGranularityDay:
		return UsageGranularityDay
	}
	return UsageGranularityRefresh
}

// GetAirportUsageHistory 构建单个机场的用量历史。
func GetAirportUsageHistory(airportID, days int, granularity string) (*AirportUsageHistory, error) {
	airport, err := GetAirportByID(airportID)
	if err != nil {
		return nil, err
	}

	retentionDays := GetAirportUsageRetentionDays()
	days = NormalizeUsageHistoryDays(days, retentionDays)
	requestedGranularity := granularity
	to := time.Now()
	from := to.AddDate(0, 0, -days)

	snapshotMap, err := loadAirportUsageSnapshots([]int{airportID}, from)
	if err != nil {
		return nil, err
	}
	granularity = resolveUsageGranularity(requestedGranularity)
	points := buildUsagePoints(snapshotMap[airportID], granularity)
	if granularity == UsageGranularityRefresh {
		points = downsampleUsagePoints(points, maxUsageHistoryPoints)
	}

	history := &AirportUsageHistory{
		AirportID:       airport.ID,
		AirportName:     airport.Name,
		Days:            days,
		Granularity:     granularity,
		From:            from,
		To:              to,
		RetentionDays:   retentionDays,
		CurrentUpload:   airport.UsageUpload,
		CurrentDownload: airport.UsageDownload,
		CurrentUsed:     airport.UsageUpload + airport.UsageDownload,
		CurrentTotal:    airport.UsageTotal,
		CurrentExpire:   airport.UsageExpire,
		Points:          points,
	}
	return history, nil
}

// GetAirportUsageSparklines 批量构建机场用量迷你图数据。
func GetAirportUsageSparklines(airportIDs []int, days int) ([]AirportUsageSparkline, error) {
	return getAirportUsageSparklines(airportIDs, days, time.Now())
}

// getAirportUsageSparklines 是 GetAirportUsageSparklines 的可注入当前时间的实现，便于测试跨天边界。
func getAirportUsageSparklines(airportIDs []int, days int, now time.Time) ([]AirportUsageSparkline, error) {
	ids := uniquePositiveIDs(airportIDs)
	if len(ids) == 0 {
		return []AirportUsageSparkline{}, nil
	}
	if len(ids) > maxUsageSparklineAirports {
		ids = ids[:maxUsageSparklineAirports]
	}

	if days <= 0 {
		days = DefaultUsageSparklineDays
	}
	retentionDays := GetAirportUsageRetentionDays()
	days = NormalizeUsageHistoryDays(days, retentionDays)
	from := now.AddDate(0, 0, -days)

	snapshotMap, err := loadAirportUsageSnapshots(ids, from)
	if err != nil {
		return nil, err
	}
	todayUsedMap, err := loadUsageTodayUsed(ids, now)
	if err != nil {
		return nil, err
	}

	sparklines := make([]AirportUsageSparkline, 0, len(ids))
	for _, airportID := range ids {
		snapshots := snapshotMap[airportID]
		points := downsampleUsagePoints(buildUsagePoints(snapshots, UsageGranularityRefresh), maxUsageSparklinePoints)
		sparkline := AirportUsageSparkline{
			AirportID: airportID,
			Points:    points,
			TodayUsed: todayUsedMap[airportID],
		}
		if len(snapshots) > 0 {
			last := snapshots[len(snapshots)-1]
			sparkline.Failed = last.Failed
			sparkline.UpdatedAt = last.CapturedAt.UnixMilli()
		}
		sparklines = append(sparklines, sparkline)
	}
	return sparklines, nil
}

// UsageDailyPoint 单日消耗的流量，由相邻两次快照的累计值差值得出。
type UsageDailyPoint struct {
	Date     string `json:"date"` // 本地日期 YYYY-MM-DD
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
	Used     int64  `json:"used"`
}

// AirportUsageDailyReport 按天统计的流量消耗。
type AirportUsageDailyReport struct {
	Days          int               `json:"days"`
	From          time.Time         `json:"from"`
	To            time.Time         `json:"to"`
	RetentionDays int               `json:"retentionDays"`
	AirportID     int               `json:"airportId"` // 0 表示汇总全部机场
	Points        []UsageDailyPoint `json:"points"`
	TodayUsed     int64             `json:"todayUsed"`
	RangeUsed     int64             `json:"rangeUsed"`
}

// GetAirportDailyUsage 按天汇总实际消耗的流量。
// 快照里存的是面板返回的累计值，所以用相邻两次成功快照的差值作为该时段消耗；
// 计数器下降（面板重置、换套餐）时把新值整段计入当天，避免出现负数。
// airportID <= 0 表示汇总全部机场。
func GetAirportDailyUsage(days, airportID int) (*AirportUsageDailyReport, error) {
	return getAirportDailyUsage(days, airportID, time.Now())
}

// getAirportDailyUsage 是 GetAirportDailyUsage 的可注入当前时间的实现，便于测试跨天边界。
func getAirportDailyUsage(days, airportID int, now time.Time) (*AirportUsageDailyReport, error) {
	if airportID > 0 {
		if _, err := GetAirportByID(airportID); err != nil {
			return nil, err
		}
	}

	retentionDays := GetAirportUsageRetentionDays()
	days = NormalizeUsageHistoryDays(days, retentionDays)
	// 按自然日对齐：返回的正好是最近 days 天的完整日桶（含今天）
	from := startOfUsageDay(now).AddDate(0, 0, -(days - 1))

	snapshots, err := loadUsageSnapshotsForDaily(from, now, airportID)
	if err != nil {
		return nil, err
	}
	var baselineIDs []int
	if airportID > 0 {
		baselineIDs = []int{airportID}
	}
	baseline, err := loadUsageBaseline(from, baselineIDs)
	if err != nil {
		return nil, err
	}

	daily := make(map[string]*UsageDailyPoint)
	lastValues := make(map[int]usageValues, len(baseline))
	for id, values := range baseline {
		lastValues[id] = values
	}

	for _, snapshot := range snapshots {
		if snapshot.Failed {
			continue
		}
		current := usageValues{
			upload:   snapshot.Upload,
			download: snapshot.Download,
			total:    snapshot.Total,
			expire:   snapshot.Expire,
		}
		previous, seen := lastValues[snapshot.AirportID]
		lastValues[snapshot.AirportID] = current
		if !seen {
			// 窗口内首次见到该机场：没有基线就无法计算差值，从下一次快照开始累计。
			continue
		}

		uploadDelta, downloadDelta := usageDelta(current, previous)
		if uploadDelta == 0 && downloadDelta == 0 {
			continue
		}

		key := snapshot.CapturedAt.In(time.Local).Format("2006-01-02")
		point, ok := daily[key]
		if !ok {
			point = &UsageDailyPoint{Date: key}
			daily[key] = point
		}
		point.Upload += uploadDelta
		point.Download += downloadDelta
		point.Used += uploadDelta + downloadDelta
	}

	endDay := startOfUsageDay(now)
	points := make([]UsageDailyPoint, 0, days+1)
	for day := startOfUsageDay(from); !day.After(endDay); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		if point, ok := daily[key]; ok {
			points = append(points, *point)
			continue
		}
		points = append(points, UsageDailyPoint{Date: key})
	}

	report := &AirportUsageDailyReport{
		Days:          days,
		From:          from,
		To:            now,
		RetentionDays: retentionDays,
		AirportID:     airportID,
		Points:        points,
	}
	for _, point := range points {
		report.RangeUsed += point.Used
	}
	if len(points) > 0 {
		report.TodayUsed = points[len(points)-1].Used
	}
	return report, nil
}

// usageDelta 计算相邻两次快照的消耗增量。
// 计数器下降（面板重置、换套餐）时把新值整段计入，避免出现负数。
func usageDelta(current, previous usageValues) (upload, download int64) {
	upload = current.upload - previous.upload
	if upload < 0 {
		upload = current.upload
	}
	download = current.download - previous.download
	if download < 0 {
		download = current.download
	}
	return upload, download
}

// startOfUsageDay 返回本地时区的当天零点。
func startOfUsageDay(value time.Time) time.Time {
	local := value.In(time.Local)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.Local)
}

// loadUsageSnapshotsForDaily 加载时间窗口内的快照，按机场与时间升序排列。
func loadUsageSnapshotsForDaily(from, to time.Time, airportID int) ([]AirportUsageSnapshot, error) {
	query := database.DB.Where("captured_at >= ? AND captured_at <= ?", from, to)
	if airportID > 0 {
		query = query.Where("airport_id = ?", airportID)
	}
	var snapshots []AirportUsageSnapshot
	if err := query.Order("airport_id ASC").Order("captured_at ASC").Order("id ASC").Find(&snapshots).Error; err != nil {
		return nil, err
	}
	return snapshots, nil
}

// loadUsageBaseline 取每个机场在窗口开始前的最后一条成功快照，作为差值基线。
// airportIDs 为空表示不限定机场。
func loadUsageBaseline(from time.Time, airportIDs []int) (map[int]usageValues, error) {
	baseline := make(map[int]usageValues)

	idQuery := database.DB.Model(&AirportUsageSnapshot{}).Where("captured_at < ?", from)
	if len(airportIDs) > 0 {
		idQuery = idQuery.Where("airport_id IN ?", airportIDs)
	}
	var ids []int
	if err := idQuery.Distinct("airport_id").Pluck("airport_id", &ids).Error; err != nil {
		return nil, err
	}

	for _, id := range ids {
		var snapshot AirportUsageSnapshot
		err := database.DB.Where("airport_id = ? AND captured_at < ? AND failed = ?", id, from, false).
			Order("captured_at DESC").Order("id DESC").First(&snapshot).Error
		if err != nil {
			continue
		}
		baseline[id] = usageValues{
			upload:   snapshot.Upload,
			download: snapshot.Download,
			total:    snapshot.Total,
			expire:   snapshot.Expire,
		}
	}
	return baseline, nil
}

// loadUsageTodayUsed 批量统计指定机场当天的消耗量。
// 统计口径与按天报表一致：相邻两次成功快照的差值计入当天，
// 基线取当天零点前的最后一条成功快照；没有基线时从当天第一条快照开始累计。
func loadUsageTodayUsed(airportIDs []int, now time.Time) (map[int]int64, error) {
	result := make(map[int]int64, len(airportIDs))
	if len(airportIDs) == 0 {
		return result, nil
	}

	todayStart := startOfUsageDay(now)
	var snapshots []AirportUsageSnapshot
	if err := database.DB.
		Where("airport_id IN ? AND captured_at >= ? AND captured_at <= ?", airportIDs, todayStart, now).
		Order("airport_id ASC").Order("captured_at ASC").Order("id ASC").
		Find(&snapshots).Error; err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return result, nil
	}

	baseline, err := loadUsageBaseline(todayStart, airportIDs)
	if err != nil {
		return nil, err
	}
	lastValues := make(map[int]usageValues, len(baseline))
	for id, values := range baseline {
		lastValues[id] = values
	}

	for _, snapshot := range snapshots {
		if snapshot.Failed {
			continue
		}
		current := usageValues{
			upload:   snapshot.Upload,
			download: snapshot.Download,
			total:    snapshot.Total,
			expire:   snapshot.Expire,
		}
		previous, seen := lastValues[snapshot.AirportID]
		lastValues[snapshot.AirportID] = current
		if !seen {
			continue
		}
		uploadDelta, downloadDelta := usageDelta(current, previous)
		result[snapshot.AirportID] += uploadDelta + downloadDelta
	}
	return result, nil
}

// loadAirportUsageSnapshots 批量加载指定时间范围内的快照，按机场分组。
func loadAirportUsageSnapshots(airportIDs []int, from time.Time) (map[int][]AirportUsageSnapshot, error) {
	result := make(map[int][]AirportUsageSnapshot, len(airportIDs))
	if len(airportIDs) == 0 {
		return result, nil
	}
	var snapshots []AirportUsageSnapshot
	if err := database.DB.
		Where("airport_id IN ? AND captured_at >= ?", airportIDs, from).
		Order("captured_at ASC").Order("id ASC").
		Find(&snapshots).Error; err != nil {
		return nil, err
	}
	for _, snapshot := range snapshots {
		result[snapshot.AirportID] = append(result[snapshot.AirportID], snapshot)
	}
	return result, nil
}

// buildUsagePoints 将单个机场的快照聚合为图表点。
func buildUsagePoints(snapshots []AirportUsageSnapshot, granularity string) []UsageHistoryPoint {
	if len(snapshots) == 0 {
		return []UsageHistoryPoint{}
	}
	if granularity == UsageGranularityRefresh {
		return buildRefreshUsagePoints(snapshots)
	}
	buckets := buildUsageBuckets(snapshots, granularity)
	if len(buckets) == 0 {
		return []UsageHistoryPoint{}
	}
	points := make([]UsageHistoryPoint, 0, len(buckets))
	var carry usageValues
	for _, bucket := range buckets {
		if bucket.hasValue {
			carry = bucket.values
		}
		points = append(points, usagePoint(bucket, carry))
	}
	return points
}

// buildRefreshUsagePoints 为每条快照生成一个图表点，保留真实的刷新时间。
// 失败点仍沿用最近一次成功快照的数值，保证曲线连续。
func buildRefreshUsagePoints(snapshots []AirportUsageSnapshot) []UsageHistoryPoint {
	points := make([]UsageHistoryPoint, 0, len(snapshots))
	var carry usageValues
	for _, snapshot := range snapshots {
		if !snapshot.Failed {
			carry = usageValues{
				upload:   snapshot.Upload,
				download: snapshot.Download,
				total:    snapshot.Total,
				expire:   snapshot.Expire,
			}
		}
		points = append(points, UsageHistoryPoint{
			Timestamp:  snapshot.CapturedAt.UnixMilli(),
			CapturedAt: snapshot.CapturedAt,
			Upload:     carry.upload,
			Download:   carry.download,
			Used:       carry.upload + carry.download,
			Total:      carry.total,
			Expire:     carry.expire,
			Failed:     snapshot.Failed,
			Error:      snapshot.Error,
		})
	}
	return points
}

// downsampleUsagePoints 均匀降采样图表点：每个窗口保留最后一个点，
// 并保留窗口内出现过的失败标记。
func downsampleUsagePoints(points []UsageHistoryPoint, maxPoints int) []UsageHistoryPoint {
	if maxPoints <= 1 || len(points) <= maxPoints {
		return points
	}
	stride := (len(points) + maxPoints - 1) / maxPoints
	result := make([]UsageHistoryPoint, 0, maxPoints)
	for start := 0; start < len(points); start += stride {
		end := start + stride
		if end > len(points) {
			end = len(points)
		}
		sampled := points[end-1]
		for _, point := range points[start:end] {
			if point.Failed {
				sampled.Failed = true
				if sampled.Error == "" {
					sampled.Error = point.Error
				}
				break
			}
		}
		result = append(result, sampled)
	}
	return result
}

// buildUsageBuckets 按粒度把快照聚合到时间桶，桶内数值取最后一次成功快照。
func buildUsageBuckets(snapshots []AirportUsageSnapshot, granularity string) []usageBucket {
	if len(snapshots) == 0 {
		return nil
	}

	bucketMap := make(map[time.Time]*usageBucket)
	keys := make([]time.Time, 0)
	for _, snapshot := range snapshots {
		key := usageBucketKey(snapshot.CapturedAt, granularity)
		bucket, ok := bucketMap[key]
		if !ok {
			bucket = &usageBucket{key: key}
			bucketMap[key] = bucket
			keys = append(keys, key)
		}
		if bucket.sampledAt.Before(snapshot.CapturedAt) {
			bucket.sampledAt = snapshot.CapturedAt
		}
		if snapshot.Failed {
			bucket.failed = true
			if bucket.errorText == "" {
				bucket.errorText = snapshot.Error
			}
			continue
		}
		bucket.values = usageValues{
			upload:   snapshot.Upload,
			download: snapshot.Download,
			total:    snapshot.Total,
			expire:   snapshot.Expire,
		}
		bucket.hasValue = true
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i].Before(keys[j]) })
	buckets := make([]usageBucket, 0, len(keys))
	for _, key := range keys {
		buckets = append(buckets, *bucketMap[key])
	}
	return buckets
}

// usagePoint 由时间桶和继承数值生成图表点。
func usagePoint(bucket usageBucket, carry usageValues) UsageHistoryPoint {
	values := carry
	if bucket.hasValue {
		values = bucket.values
	}
	return UsageHistoryPoint{
		Timestamp:  bucket.sampledAt.UnixMilli(),
		CapturedAt: bucket.sampledAt,
		Upload:     values.upload,
		Download:   values.download,
		Used:       values.upload + values.download,
		Total:      values.total,
		Expire:     values.expire,
		Failed:     bucket.failed,
		Error:      bucket.errorText,
	}
}

// usageBucketKey 返回快照所属的时间桶起点。
func usageBucketKey(value time.Time, granularity string) time.Time {
	local := value.In(time.Local)
	if granularity == UsageGranularityHour {
		return time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), 0, 0, 0, time.Local)
	}
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.Local)
}

// uniquePositiveIDs 去重并过滤非法 ID。
func uniquePositiveIDs(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
