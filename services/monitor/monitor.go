package monitor

import (
	"runtime"
	"strconv"
	"strings"
	"sublink/config"
	"sync"
	"time"
)

type RuntimeConfigItem struct {
	Label       string         `json:"label"`
	Key         string         `json:"key"`
	Value       string         `json:"value"`
	ValueKey    string         `json:"value_key,omitempty"`
	ValueParams map[string]any `json:"value_params,omitempty"`
	Env         string         `json:"env"`
}

type RuntimeConfigSummary struct {
	SafeToShow    []RuntimeConfigItem `json:"safe_to_show"`
	MaskedSummary []RuntimeConfigItem `json:"masked_summary"`
}

// SystemStats 系统监控统计信息
type SystemStats struct {
	// 进程内存统计 (Go runtime)
	HeapAlloc   uint64  `json:"heap_alloc"`   // 堆内存已分配 (bytes)
	HeapSys     uint64  `json:"heap_sys"`     // 堆内存从OS获取 (bytes)
	HeapInuse   uint64  `json:"heap_inuse"`   // 堆内存使用中 (bytes)
	StackInuse  uint64  `json:"stack_inuse"`  // 栈内存使用中 (bytes)
	Sys         uint64  `json:"sys"`          // 从OS获取的总内存 (bytes)
	TotalAlloc  uint64  `json:"total_alloc"`  // 累计分配的内存 (bytes)
	MemoryUsage float64 `json:"memory_usage"` // 内存使用率估算 (%)

	// CPU 统计
	NumCPU     int     `json:"num_cpu"`    // 逻辑CPU数量
	GOMAXPROCS int     `json:"gomaxprocs"` // Go可用的最大处理器数
	CPUUsage   float64 `json:"cpu_usage"`  // 进程CPU使用率 (%)

	// Goroutine/线程统计
	NumGoroutine int   `json:"num_goroutine"` // 当前Goroutine数量
	NumCgoCall   int64 `json:"num_cgo_call"`  // CGO调用次数

	// GC 统计
	NumGC        uint32  `json:"num_gc"`         // GC循环次数
	LastGCTime   int64   `json:"last_gc_time"`   // 上次GC时间 (unix毫秒)
	PauseTotalNs uint64  `json:"pause_total_ns"` // GC暂停总时间 (纳秒)
	GCCPUFrac    float64 `json:"gc_cpu_frac"`    // GC使用的CPU比例

	// 应用运行时间
	StartTime int64 `json:"start_time"` // 启动时间 (unix秒)
	Uptime    int64 `json:"uptime"`     // 运行时间 (秒)

	// Go版本信息
	GoVersion string `json:"go_version"` // Go版本
	GOARCH    string `json:"goarch"`     // 目标架构
	GOOS      string `json:"goos"`       // 目标操作系统

	RuntimeConfig RuntimeConfigSummary `json:"runtime_config"`
}

var (
	startTime time.Time
	once      sync.Once

	// CPU使用率计算相关
	lastCPUTime time.Time
	cpuMutex    sync.Mutex
)

// init 初始化启动时间
func init() {
	once.Do(func() {
		startTime = time.Now()
	})
}

// GetSystemStats 获取系统监控统计信息
// 使用Go runtime包实现跨平台兼容
func GetSystemStats() SystemStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 计算CPU使用率
	cpuUsage := calculateCPUUsage()

	// 计算内存使用率 (基于堆内存与系统分配的比例)
	memoryUsage := 0.0
	if memStats.Sys > 0 {
		memoryUsage = float64(memStats.HeapInuse+memStats.StackInuse) / float64(memStats.Sys) * 100
	}

	// 上次GC时间转换
	lastGCTime := int64(0)
	if memStats.LastGC > 0 {
		lastGCTime = int64(memStats.LastGC / 1e6) // 转换为毫秒
	}

	return SystemStats{
		// 进程内存统计
		HeapAlloc:   memStats.HeapAlloc,
		HeapSys:     memStats.HeapSys,
		HeapInuse:   memStats.HeapInuse,
		StackInuse:  memStats.StackInuse,
		Sys:         memStats.Sys,
		TotalAlloc:  memStats.TotalAlloc,
		MemoryUsage: memoryUsage,

		// CPU统计
		NumCPU:     runtime.NumCPU(),
		GOMAXPROCS: runtime.GOMAXPROCS(0),
		CPUUsage:   cpuUsage,

		// Goroutine/线程统计
		NumGoroutine: runtime.NumGoroutine(),
		NumCgoCall:   runtime.NumCgoCall(),

		// GC统计
		NumGC:        memStats.NumGC,
		LastGCTime:   lastGCTime,
		PauseTotalNs: memStats.PauseTotalNs,
		GCCPUFrac:    memStats.GCCPUFraction,

		// 运行时间
		StartTime: startTime.Unix(),
		Uptime:    int64(time.Since(startTime).Seconds()),

		// Go版本信息
		GoVersion: runtime.Version(),
		GOARCH:    runtime.GOARCH,
		GOOS:      runtime.GOOS,

		RuntimeConfig: buildRuntimeConfigSummary(),
	}
}

func buildRuntimeConfigSummary() RuntimeConfigSummary {
	cfg := config.Get()
	captchaCfg := config.GetCaptchaConfig()
	features := config.GetEnabledFeatures()

	safeToShow := []RuntimeConfigItem{
		{Label: "Port", Key: "port", Value: formatIntValue(cfg.Port, ""), ValueKey: intValueKey(cfg.Port, "plain"), ValueParams: intValueParams(cfg.Port), Env: "SUBLINK_PORT"},
		{Label: "Login token TTL", Key: "expire_days", Value: formatIntValue(cfg.ExpireDays, "d"), ValueKey: intValueKey(cfg.ExpireDays, "days"), ValueParams: intValueParams(cfg.ExpireDays), Env: "SUBLINK_EXPIRE_DAYS"},
		{Label: "Login failure limit", Key: "login_fail_count", Value: formatIntValue(cfg.LoginFailCount, "attempts"), ValueKey: intValueKey(cfg.LoginFailCount, "times"), ValueParams: intValueParams(cfg.LoginFailCount), Env: "SUBLINK_LOGIN_FAIL_COUNT"},
		{Label: "Login failure window", Key: "login_fail_window", Value: formatIntValue(cfg.LoginFailWindow, "minutes"), ValueKey: intValueKey(cfg.LoginFailWindow, "minutes"), ValueParams: intValueParams(cfg.LoginFailWindow), Env: "SUBLINK_LOGIN_FAIL_WINDOW"},
		{Label: "Login ban duration", Key: "login_ban_duration", Value: formatIntValue(cfg.LoginBanDuration, "minutes"), ValueKey: intValueKey(cfg.LoginBanDuration, "minutes"), ValueParams: intValueParams(cfg.LoginBanDuration), Env: "SUBLINK_LOGIN_BAN_DURATION"},
		{Label: "Log level", Key: "log_level", Value: emptyFallback(strings.ToUpper(cfg.LogLevel), "Not set"), ValueKey: nonEmptyValueKey(cfg.LogLevel, "literal", "unset"), ValueParams: literalValueParams(strings.ToUpper(cfg.LogLevel)), Env: "SUBLINK_LOG_LEVEL"},
		{Label: "Captcha mode", Key: "captcha_mode", Value: formatCaptchaMode(captchaCfg), ValueKey: captchaModeValueKey(captchaCfg), Env: "SUBLINK_CAPTCHA_MODE"},
		{Label: "Trusted proxy rules", Key: "trusted_proxies", Value: formatTrustedProxies(cfg.TrustedProxies), ValueKey: trustedProxiesValueKey(cfg.TrustedProxies), ValueParams: countValueParams(len(cfg.TrustedProxies)), Env: "SUBLINK_TRUSTED_PROXIES"},
		{Label: "Feature toggles", Key: "feature", Value: formatEnabledFeatures(features), ValueKey: enabledFeaturesValueKey(features), ValueParams: listValueParams(features), Env: "SUBLINK_FEATURE"},
	}

	maskedSummary := []RuntimeConfigItem{
		{Label: "Database connection", Key: "dsn", Value: summarizeDSN(cfg.DSN), ValueKey: dsnSummaryValueKey(cfg.DSN), ValueParams: dsnSummaryValueParams(cfg.DSN), Env: "SUBLINK_DSN"},
		{Label: "Local data directory", Key: "db_path", Value: summarizePathSetting(cfg.DBPath, config.DefaultDBPath), ValueKey: pathSettingValueKey(cfg.DBPath, config.DefaultDBPath), Env: "SUBLINK_DB_PATH"},
		{Label: "Log directory", Key: "log_path", Value: summarizePathSetting(cfg.LogPath, config.DefaultLogPath), ValueKey: pathSettingValueKey(cfg.LogPath, config.DefaultLogPath), Env: "SUBLINK_LOG_PATH"},
		{Label: "GeoIP database", Key: "geoip_path", Value: summarizeOptionalPathSetting(cfg.GeoIPPath), ValueKey: optionalPathSettingValueKey(cfg.GeoIPPath), Env: "SUBLINK_GEOIP_PATH"},
		{Label: "Frontend base path", Key: "web_base_path", Value: summarizeWebBasePath(cfg.WebBasePath), ValueKey: webBasePathValueKey(cfg.WebBasePath), Env: "SUBLINK_WEB_BASE_PATH"},
		{Label: "Turnstile site key", Key: "turnstile_site_key", Value: configuredSummary(cfg.TurnstileSiteKey), ValueKey: configuredSummaryValueKey(cfg.TurnstileSiteKey), Env: "SUBLINK_TURNSTILE_SITE_KEY"},
		{Label: "Turnstile verify proxy", Key: "turnstile_proxy_link", Value: configuredSummary(cfg.TurnstileProxyLink), ValueKey: configuredSummaryValueKey(cfg.TurnstileProxyLink), Env: "SUBLINK_TURNSTILE_PROXY_LINK"},
	}

	return RuntimeConfigSummary{SafeToShow: safeToShow, MaskedSummary: maskedSummary}
}

func valueKey(name string) string {
	return "components.systemMonitor.config.values." + name
}

func intValueKey(value int, unit string) string {
	if value <= 0 {
		return valueKey("unset")
	}
	return valueKey(unit)
}

func intValueParams(value int) map[string]any {
	if value <= 0 {
		return nil
	}
	return map[string]any{"count": value}
}

func countValueParams(count int) map[string]any {
	return map[string]any{"count": count}
}

func literalValueParams(value string) map[string]any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return map[string]any{"value": value}
}

func listValueParams(values []string) map[string]any {
	if len(values) == 0 {
		return nil
	}
	return map[string]any{"value": strings.Join(values, ", ")}
}

func nonEmptyValueKey(value, nonEmptyKey, emptyKey string) string {
	if strings.TrimSpace(value) == "" {
		return valueKey(emptyKey)
	}
	return valueKey(nonEmptyKey)
}

func captchaModeValueKey(cfg config.CaptchaConfig) string {
	switch cfg.Mode {
	case config.CaptchaModeDisabled:
		return valueKey("captchaDisabled")
	case config.CaptchaModeTurnstile:
		return valueKey("captchaTurnstile")
	default:
		if cfg.ConfiguredMode == config.CaptchaModeTurnstile && cfg.Degraded {
			return valueKey("captchaTurnstileDegraded")
		}
		return valueKey("captchaLegacy")
	}
}

func trustedProxiesValueKey(proxies []string) string {
	if len(proxies) == 0 {
		return valueKey("disabled")
	}
	return valueKey("rules")
}

func enabledFeaturesValueKey(features []string) string {
	if len(features) == 0 {
		return valueKey("disabled")
	}
	return valueKey("literal")
}

func configuredSummaryValueKey(value string) string {
	if strings.TrimSpace(value) == "" {
		return valueKey("notConfigured")
	}
	return valueKey("configured")
}

func pathSettingValueKey(value, defaultValue string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return valueKey("defaultPath")
	}
	if strings.TrimSpace(defaultValue) != "" && trimmed == defaultValue {
		return valueKey("defaultPath")
	}
	return valueKey("customPath")
}

func optionalPathSettingValueKey(value string) string {
	if strings.TrimSpace(value) == "" {
		return valueKey("autoDetect")
	}
	return valueKey("customPath")
}

func webBasePathValueKey(basePath string) string {
	if strings.TrimSpace(basePath) == "" {
		return valueKey("rootPath")
	}
	return valueKey("customPath")
}

func dsnSummaryValueKey(dsn string) string {
	trimmed := strings.TrimSpace(dsn)
	if trimmed == "" {
		return valueKey("databaseConfigured")
	}
	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "sqlite"):
		return valueKey("databaseConfigured")
	case strings.HasPrefix(lower, "mysql://") || strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://"):
		kind, host, database := parseSQLDSNSummary(trimmed)
		if host != "" && database != "" {
			_ = kind
			return valueKey("databaseHost")
		}
		return valueKey("databaseConfigured")
	default:
		return valueKey("configuredHidden")
	}
}

func dsnSummaryValueParams(dsn string) map[string]any {
	trimmed := strings.TrimSpace(dsn)
	if trimmed == "" {
		return map[string]any{"kind": "SQLite"}
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "sqlite") {
		return map[string]any{"kind": "SQLite"}
	}
	if strings.HasPrefix(lower, "mysql://") || strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://") {
		kind, host, database := parseSQLDSNSummary(trimmed)
		params := map[string]any{"kind": kind}
		if host != "" {
			params["host"] = host
		}
		if database != "" {
			params["database"] = database
		}
		return params
	}
	return nil
}

func parseSQLDSNSummary(dsn string) (string, string, string) {
	kind := "PostgreSQL"
	if strings.HasPrefix(strings.ToLower(dsn), "mysql://") {
		kind = "MySQL"
	}

	parts := strings.SplitN(dsn, "@", 2)
	if len(parts) != 2 {
		return kind, "", ""
	}

	hostAndDB := parts[1]
	hostAndDB = strings.TrimPrefix(hostAndDB, "tcp(")
	hostAndDB = strings.TrimPrefix(hostAndDB, "unix(")
	hostAndDB = strings.ReplaceAll(hostAndDB, ")", "")
	hostAndDB = strings.TrimPrefix(hostAndDB, "/")

	segments := strings.SplitN(hostAndDB, "/", 2)
	host := strings.TrimSpace(segments[0])
	database := ""
	if len(segments) == 2 {
		database = strings.SplitN(segments[1], "?", 2)[0]
	}
	return kind, host, database
}

func formatIntValue(value int, unit string) string {
	if value <= 0 {
		return "Not set"
	}
	if unit == "" {
		return intToString(value)
	}
	return intToString(value) + " " + unit
}

func intToString(value int) string {
	return strconv.Itoa(value)
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func formatCaptchaMode(cfg config.CaptchaConfig) string {
	switch cfg.Mode {
	case config.CaptchaModeDisabled:
		return "Off"
	case config.CaptchaModeTurnstile:
		return "Cloudflare Turnstile"
	default:
		if cfg.ConfiguredMode == config.CaptchaModeTurnstile && cfg.Degraded {
			return "Cloudflare Turnstile (fell back to legacy captcha)"
		}
		return "Legacy captcha"
	}
}

func formatTrustedProxies(proxies []string) string {
	if len(proxies) == 0 {
		return "Disabled"
	}
	return intToString(len(proxies)) + " rules"
}

func formatEnabledFeatures(features []string) string {
	if len(features) == 0 {
		return "Disabled"
	}
	return strings.Join(features, "、")
}

func configuredSummary(value string) string {
	if strings.TrimSpace(value) == "" {
		return "Not configured"
	}
	return "Configured"
}

func summarizePathSetting(value, defaultValue string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "Default path"
	}
	if strings.TrimSpace(defaultValue) != "" && trimmed == defaultValue {
		return "Default path"
	}
	return "Customized"
}

func summarizeOptionalPathSetting(value string) string {
	if strings.TrimSpace(value) == "" {
		return "Auto-detect"
	}
	return "Customized"
}

func summarizeDSN(dsn string) string {
	trimmed := strings.TrimSpace(dsn)
	if trimmed == "" {
		return "SQLite · Configured"
	}

	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "sqlite"):
		return "SQLite · Configured"
	case strings.HasPrefix(lower, "mysql://"):
		return summarizeSQLDSN("MySQL", trimmed)
	case strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://"):
		return summarizeSQLDSN("PostgreSQL", trimmed)
	default:
		return "Configured (details hidden)"
	}
}

func summarizeSQLDSN(kind, dsn string) string {
	parts := strings.SplitN(dsn, "@", 2)
	if len(parts) != 2 {
		return kind + " · Configured (details hidden)"
	}

	hostAndDB := parts[1]
	hostAndDB = strings.TrimPrefix(hostAndDB, "tcp(")
	hostAndDB = strings.TrimPrefix(hostAndDB, "unix(")
	hostAndDB = strings.ReplaceAll(hostAndDB, ")", "")
	hostAndDB = strings.TrimPrefix(hostAndDB, "/")

	segments := strings.SplitN(hostAndDB, "/", 2)
	host := strings.TrimSpace(segments[0])
	database := ""
	if len(segments) == 2 {
		database = strings.SplitN(segments[1], "?", 2)[0]
	}

	if host == "" && database == "" {
		return kind + " · Configured (details hidden)"
	}
	if database == "" || host == "" {
		return kind + " · Configured"
	}
	return kind + " · " + host + " · " + database
}

func summarizeWebBasePath(basePath string) string {
	if strings.TrimSpace(basePath) == "" {
		return "Root path"
	}
	return "Customized"
}

// calculateCPUUsage 计算进程CPU使用率
// 使用简单的采样方法，跨平台兼容
func calculateCPUUsage() float64 {
	cpuMutex.Lock()
	defer cpuMutex.Unlock()

	now := time.Now()

	// 首次调用时初始化
	if lastCPUTime.IsZero() {
		lastCPUTime = now
		return 0.0
	}

	// 时间间隔太短，返回上次的值
	elapsed := now.Sub(lastCPUTime)
	if elapsed < 100*time.Millisecond {
		return 0.0
	}

	// 使用GC CPU分数作为CPU使用率的估算
	// 这是一个简化的方法，但保证跨平台兼容
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// GCCPUFraction 表示GC使用的CPU时间比例
	// 我们将其与goroutine数量结合估算总CPU使用率
	goroutineLoad := float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 10
	if goroutineLoad > 100 {
		goroutineLoad = 100
	}

	// 简单的CPU使用率估算 (基于活跃goroutine数量)
	// 这不是精确的CPU使用率，但是跨平台兼容的
	cpuUsage := goroutineLoad * 0.5 // 假设平均50%活跃
	if cpuUsage > 100 {
		cpuUsage = 100
	}

	lastCPUTime = now

	return cpuUsage
}

// FormatBytes 将字节转换为人类可读的格式
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return formatBytesValue(float64(bytes), "B")
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return formatBytesValue(float64(bytes)/float64(div), units[exp])
}

func formatBytesValue(value float64, unit string) string {
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10) + " " + unit
	}
	return strconv.FormatFloat(value, 'f', 2, 64) + " " + unit
}

// FormatDuration 将秒数转换为人类可读的时长格式
func FormatDuration(seconds int64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if days > 0 {
		return formatDurationString(days, hours, minutes)
	} else if hours > 0 {
		return formatHoursMinutes(hours, minutes, secs)
	} else if minutes > 0 {
		return formatMinutesSeconds(minutes, secs)
	}
	return formatSeconds(secs)
}

func formatDurationString(days, hours, minutes int64) string {
	return strconv.FormatInt(days, 10) + "d" + strconv.FormatInt(hours, 10) + "h" + strconv.FormatInt(minutes, 10) + "m"
}

func formatHoursMinutes(hours, minutes, secs int64) string {
	return strconv.FormatInt(hours, 10) + "h" + strconv.FormatInt(minutes, 10) + "m" + strconv.FormatInt(secs, 10) + "s"
}

func formatMinutesSeconds(minutes, secs int64) string {
	return strconv.FormatInt(minutes, 10) + "m" + strconv.FormatInt(secs, 10) + "s"
}

func formatSeconds(secs int64) string {
	return strconv.FormatInt(secs, 10) + "s"
}
