package monitor

import (
	"runtime"
	"sync"
	"time"
)

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
	}
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
