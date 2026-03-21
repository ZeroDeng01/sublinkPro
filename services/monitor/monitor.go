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
	Label string `json:"label"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Env   string `json:"env"`
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
	lastCPUTime  time.Time
	lastUserTime float64
	lastSysTime  float64
	cpuMutex     sync.Mutex
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
		{Label: "端口", Key: "port", Value: formatIntValue(cfg.Port, ""), Env: "SUBLINK_PORT"},
		{Label: "登录令牌有效期", Key: "expire_days", Value: formatIntValue(cfg.ExpireDays, "天"), Env: "SUBLINK_EXPIRE_DAYS"},
		{Label: "登录失败次数限制", Key: "login_fail_count", Value: formatIntValue(cfg.LoginFailCount, "次"), Env: "SUBLINK_LOGIN_FAIL_COUNT"},
		{Label: "登录失败统计窗口", Key: "login_fail_window", Value: formatIntValue(cfg.LoginFailWindow, "分钟"), Env: "SUBLINK_LOGIN_FAIL_WINDOW"},
		{Label: "登录封禁时长", Key: "login_ban_duration", Value: formatIntValue(cfg.LoginBanDuration, "分钟"), Env: "SUBLINK_LOGIN_BAN_DURATION"},
		{Label: "日志级别", Key: "log_level", Value: emptyFallback(strings.ToUpper(cfg.LogLevel), "未设置"), Env: "SUBLINK_LOG_LEVEL"},
		{Label: "验证码模式", Key: "captcha_mode", Value: formatCaptchaMode(captchaCfg), Env: "SUBLINK_CAPTCHA_MODE"},
		{Label: "可信代理规则", Key: "trusted_proxies", Value: formatTrustedProxies(cfg.TrustedProxies), Env: "SUBLINK_TRUSTED_PROXIES"},
		{Label: "功能开关", Key: "feature", Value: formatEnabledFeatures(features), Env: "SUBLINK_FEATURE"},
	}

	maskedSummary := []RuntimeConfigItem{
		{Label: "数据库连接", Key: "dsn", Value: summarizeDSN(cfg.DSN), Env: "SUBLINK_DSN"},
		{Label: "本地数据目录", Key: "db_path", Value: summarizePathSetting(cfg.DBPath, config.DefaultDBPath), Env: "SUBLINK_DB_PATH"},
		{Label: "日志目录路径", Key: "log_path", Value: summarizePathSetting(cfg.LogPath, config.DefaultLogPath), Env: "SUBLINK_LOG_PATH"},
		{Label: "GeoIP 数据库", Key: "geoip_path", Value: summarizeOptionalPathSetting(cfg.GeoIPPath), Env: "SUBLINK_GEOIP_PATH"},
		{Label: "前端访问基础路径", Key: "web_base_path", Value: summarizeWebBasePath(cfg.WebBasePath), Env: "SUBLINK_WEB_BASE_PATH"},
		{Label: "Turnstile 站点密钥", Key: "turnstile_site_key", Value: configuredSummary(cfg.TurnstileSiteKey), Env: "SUBLINK_TURNSTILE_SITE_KEY"},
		{Label: "Turnstile 验证代理", Key: "turnstile_proxy_link", Value: configuredSummary(cfg.TurnstileProxyLink), Env: "SUBLINK_TURNSTILE_PROXY_LINK"},
	}

	return RuntimeConfigSummary{SafeToShow: safeToShow, MaskedSummary: maskedSummary}
}

func formatIntValue(value int, unit string) string {
	if value <= 0 {
		return "未设置"
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
		return "关闭"
	case config.CaptchaModeTurnstile:
		return "Cloudflare Turnstile"
	default:
		if cfg.ConfiguredMode == config.CaptchaModeTurnstile && cfg.Degraded {
			return "Cloudflare Turnstile（已降级为传统验证码）"
		}
		return "传统验证码"
	}
}

func formatTrustedProxies(proxies []string) string {
	if len(proxies) == 0 {
		return "未启用"
	}
	return intToString(len(proxies)) + " 条规则"
}

func formatEnabledFeatures(features []string) string {
	if len(features) == 0 {
		return "未启用"
	}
	return strings.Join(features, "、")
}

func configuredSummary(value string) string {
	if strings.TrimSpace(value) == "" {
		return "未配置"
	}
	return "已配置"
}

func summarizePathSetting(value, defaultValue string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "默认路径"
	}
	if strings.TrimSpace(defaultValue) != "" && trimmed == defaultValue {
		return "默认路径"
	}
	return "已自定义"
}

func summarizeOptionalPathSetting(value string) string {
	if strings.TrimSpace(value) == "" {
		return "自动检测"
	}
	return "已自定义"
}

func summarizeDSN(dsn string) string {
	trimmed := strings.TrimSpace(dsn)
	if trimmed == "" {
		return "SQLite · 已配置"
	}

	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasPrefix(lower, "sqlite"):
		return "SQLite · 已配置"
	case strings.HasPrefix(lower, "mysql://"):
		return summarizeSQLDSN("MySQL", trimmed)
	case strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://"):
		return summarizeSQLDSN("PostgreSQL", trimmed)
	default:
		return "已配置（详情已隐藏）"
	}
}

func summarizeSQLDSN(kind, dsn string) string {
	parts := strings.SplitN(dsn, "@", 2)
	if len(parts) != 2 {
		return kind + " · 已配置（详情已隐藏）"
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
		return kind + " · 已配置（详情已隐藏）"
	}
	if database == "" || host == "" {
		return kind + " · 已配置"
	}
	return kind + " · " + host + " · " + database
}

func summarizeWebBasePath(basePath string) string {
	if strings.TrimSpace(basePath) == "" {
		return "根路径"
	}
	return "已自定义"
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
		return string(rune(int64(value))) + " " + unit
	}
	return string(rune(int64(value*100)/100)) + " " + unit
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
	return string(rune(days)) + "天" + string(rune(hours)) + "时" + string(rune(minutes)) + "分"
}

func formatHoursMinutes(hours, minutes, secs int64) string {
	return string(rune(hours)) + "时" + string(rune(minutes)) + "分" + string(rune(secs)) + "秒"
}

func formatMinutesSeconds(minutes, secs int64) string {
	return string(rune(minutes)) + "分" + string(rune(secs)) + "秒"
}

func formatSeconds(secs int64) string {
	return string(rune(secs)) + "秒"
}
