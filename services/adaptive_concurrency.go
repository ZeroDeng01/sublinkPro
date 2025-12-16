package services

import (
	"fmt"
	"runtime"
	"sublink/utils"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// 动态并发控制器配置常量
// 修改这些值可以调整自适应并发的行为
// ============================================================

// --- 延迟测试并发配置 ---
// 降低并发上限以避免资源竞争导致的虚假超时
const (
	// LatencyMinConcurrencyPerCPU 延迟测试每CPU核心最小并发数
	LatencyMinConcurrencyPerCPU = 2
	// LatencyMinConcurrencyFloor 延迟测试并发数下限（绝对值）
	LatencyMinConcurrencyFloor = 4
	// LatencyMaxConcurrencyPerCPU 延迟测试每CPU核心最大并发数（降低以减少资源竞争）
	LatencyMaxConcurrencyPerCPU = 20 // 默认20
	// LatencyMaxConcurrencyCeiling 延迟测试并发数上限（绝对值）
	LatencyMaxConcurrencyCeiling = 500 // 原为500
	// LatencyInitialConcurrencyPerCPU 延迟测试每CPU核心初始并发数
	LatencyInitialConcurrencyPerCPU = 5 // 原为8
)

// --- 速度测试并发配置 ---
// 速度测试需要低并发以确保带宽测量准确性
const (
	// SpeedMinConcurrency 速度测试最小并发数
	SpeedMinConcurrency = 1
	// SpeedMaxConcurrencyPerCPU 速度测试每CPU核心最大并发数（降低以避免带宽竞争）
	SpeedMaxConcurrencyPerCPU = 4 // 默认4
	// SpeedMaxConcurrencyCeiling 速度测试并发数上限（绝对值）
	SpeedMaxConcurrencyCeiling = 32 // 原为32
	// SpeedInitialConcurrencyDivisor 速度测试初始并发数计算除数
	SpeedInitialConcurrencyDivisor = 2
	// SpeedInitialConcurrencyFloor 速度测试初始并发数下限
	SpeedInitialConcurrencyFloor = 1 // 原为2，改为1实现准串行
)

// --- 资源阈值配置 ---
const (
	// TargetCPUUsageLatency 延迟测试目标CPU使用率上限
	TargetCPUUsageLatency = 0.80
	// TargetMemoryUsageLatency 延迟测试目标内存使用率上限
	TargetMemoryUsageLatency = 0.80
	// TargetCPUUsageSpeed 速度测试目标CPU使用率上限
	TargetCPUUsageSpeed = 0.80
	// TargetMemoryUsageSpeed 速度测试目标内存使用率上限
	TargetMemoryUsageSpeed = 0.80
	// MemoryEmergencyThreshold 内存紧急阈值（触发紧急降低）
	MemoryEmergencyThreshold = 0.90
	// CPUHighThresholdOffset CPU过高阈值偏移量(相对于target)
	CPUHighThresholdOffset = 0.10
)

// --- 调整幅度配置 ---
const (
	// IncreaseMultiplier 提升并发时的乘数
	IncreaseMultiplier = 1.5
	// DecreaseCPUMultiplier CPU过高时降低并发的乘数
	DecreaseCPUMultiplier = 0.8
	// DecreaseMemoryMultiplier 内存偏高时降低并发的乘数
	DecreaseMemoryMultiplier = 0.85
	// DecreaseEmergencyMultiplier 紧急降低并发的乘数
	DecreaseEmergencyMultiplier = 0.5
	// MinAdjustmentDelta 最小调整幅度（忽略小于此值的调整）
	MinAdjustmentDelta = 2
)

// --- 调整间隔配置 ---
const (
	// AdjustInterval 两次调整之间的最小间隔（缩短以更快响应负载变化）
	AdjustInterval = 500 * time.Millisecond // 默认1秒
	// LatencyAdjustCheckInterval 延迟测试每N个任务检查一次调整
	LatencyAdjustCheckInterval = 5 // 默认10
	// SpeedAdjustCheckInterval 速度测试每N个任务检查一次调整
	SpeedAdjustCheckInterval = 5 // 默认5
)

// --- CPU估算配置 ---
const (
	// IdealGoroutinesPerCPU 每CPU核心理想goroutine数（用于估算CPU使用率）
	IdealGoroutinesPerCPU = 50
	// GoroutineSaturationWeight goroutine饱和度权重
	GoroutineSaturationWeight = 0.7
	// GCPressureWeight GC压力权重
	GCPressureWeight = 0.3
)

// --- Goroutine负载阈值 ---
const (
	// GoroutineHighThreshold 高负载阈值（每CPU），超过此值降低并发
	GoroutineHighThreshold = 30
	// GoroutineCriticalThreshold 临界负载阈值（每CPU），超过此值紧急降低并发
	GoroutineCriticalThreshold = 50
)

// --- 任务启动平滑配置 ---
const (
	// LatencyTaskStartInterval 延迟测试任务启动间隔（避免瞬时大量连接）
	LatencyTaskStartInterval = 10 * time.Millisecond
	// SpeedTaskStartInterval 速度测试任务启动间隔
	SpeedTaskStartInterval = 50 * time.Millisecond
)

// AdaptiveType 自适应控制类型
type AdaptiveType int

const (
	// AdaptiveTypeLatency 延迟测试 - 高并发，I/O密集
	AdaptiveTypeLatency AdaptiveType = iota
	// AdaptiveTypeSpeed 速度测试 - 低并发，带宽密集
	AdaptiveTypeSpeed
)

// SystemMetrics 系统指标快照
type SystemMetrics struct {
	CPUUsage     float64   // CPU使用率 (0-1)
	MemoryUsage  float64   // 内存使用率 (0-1)
	GoroutineNum int       // 当前goroutine数量
	Timestamp    time.Time // 采集时间
}

// GetSystemMetrics 获取当前系统指标
func GetSystemMetrics() SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 计算内存使用率 (使用已分配堆内存 / 系统分配给进程的总内存)
	// 注意：这是进程级别的内存使用，而非系统级别
	memoryUsage := float64(memStats.Alloc) / float64(memStats.Sys)
	if memoryUsage > 1 {
		memoryUsage = 1
	}

	return SystemMetrics{
		// 注意：Go标准库没有直接获取CPU使用率的方法
		// 我们通过goroutine数量和调度器状态间接推断
		CPUUsage:     estimateCPUUsage(),
		MemoryUsage:  memoryUsage,
		GoroutineNum: runtime.NumGoroutine(),
		Timestamp:    time.Now(),
	}
}

// estimateCPUUsage 估算CPU使用率
// 基于多种因素综合估算，避免单一指标导致的误判
func estimateCPUUsage() float64 {
	numCPU := runtime.NumCPU()
	numGoroutine := runtime.NumGoroutine()

	// 获取内存分配信息，高内存分配通常意味着高CPU活动
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 综合多个指标计算CPU压力
	// 1. Goroutine饱和度：活跃goroutine与理想并发数的比值
	idealGoroutines := numCPU * IdealGoroutinesPerCPU
	goroutineSaturation := float64(numGoroutine) / float64(idealGoroutines)
	if goroutineSaturation > 1 {
		goroutineSaturation = 1
	}

	// 2. GC压力：基于最近GC暂停时间估算
	gcPressure := 0.0
	if memStats.NumGC > 0 && memStats.PauseTotalNs > 0 {
		avgPauseNs := float64(memStats.PauseTotalNs) / float64(memStats.NumGC)
		gcPressure = avgPauseNs / 1000000.0
		if gcPressure > 1 {
			gcPressure = 1
		}
	}

	// 综合评估
	cpuUsage := goroutineSaturation*GoroutineSaturationWeight + gcPressure*GCPressureWeight

	return cpuUsage
}

// AdaptiveConcurrencyController 自适应并发控制器
type AdaptiveConcurrencyController struct {
	// 类型与配置
	adaptiveType AdaptiveType
	totalTasks   int // 总任务数

	// 并发限制
	minConcurrency int // 最小并发数
	maxConcurrency int // 最大并发数

	// 目标指标
	targetCPUUsage    float64 // 目标CPU使用率上限
	targetMemoryUsage float64 // 目标内存使用率上限
	minSuccessRate    float64 // 最低成功率阈值

	// 动态状态
	currentConcurrency int32 // 当前并发数 (atomic)
	semaphore          chan struct{}

	// 动态并发控制（用于真正的运行时调整）
	activeCount int32      // 当前活跃任务数 (atomic)
	cond        *sync.Cond // 条件变量，用于动态等待

	// 性能统计
	successCount int32 // 成功数 (atomic)
	failCount    int32 // 失败数 (atomic)
	totalLatency int64 // 累计延迟 (atomic, ms)

	// 调整控制
	lastAdjustTime time.Time
	adjustInterval time.Duration // 最小调整间隔
	mu             sync.Mutex

	// 日志记录
	adjustmentLog []concurrencyAdjustment
}

type concurrencyAdjustment struct {
	Time        time.Time
	OldValue    int
	NewValue    int
	Reason      string
	CPUUsage    float64
	MemoryUsage float64
	SuccessRate float64
}

// NewAdaptiveConcurrencyController 创建自适应并发控制器
func NewAdaptiveConcurrencyController(adaptiveType AdaptiveType, totalTasks int) *AdaptiveConcurrencyController {
	numCPU := runtime.NumCPU()

	var minC, maxC, initialC int
	var targetCPU, targetMem, minSuccess float64

	switch adaptiveType {
	case AdaptiveTypeLatency:
		// 延迟测试：I/O密集型，可以高并发
		minC = max(numCPU*LatencyMinConcurrencyPerCPU, LatencyMinConcurrencyFloor)
		maxC = min(numCPU*LatencyMaxConcurrencyPerCPU, LatencyMaxConcurrencyCeiling)
		initialC = min(numCPU*LatencyInitialConcurrencyPerCPU, totalTasks)
		if initialC < minC {
			initialC = minC
		}
		if initialC > maxC {
			initialC = maxC
		}
		targetCPU = TargetCPUUsageLatency
		targetMem = TargetMemoryUsageLatency
		minSuccess = 0.60 // 保留但不再用于调整决策

	case AdaptiveTypeSpeed:
		// 速度测试：带宽密集型，需要控制并发避免相互干扰
		minC = SpeedMinConcurrency
		maxC = min(numCPU*SpeedMaxConcurrencyPerCPU, SpeedMaxConcurrencyCeiling)
		initialC = max(numCPU/SpeedInitialConcurrencyDivisor, SpeedInitialConcurrencyFloor)
		if initialC > maxC {
			initialC = maxC
		}
		targetCPU = TargetCPUUsageSpeed
		targetMem = TargetMemoryUsageSpeed
		minSuccess = 0.70

	default:
		minC = 4
		maxC = 50
		initialC = 10
		targetCPU = 0.80
		targetMem = 0.80
		minSuccess = 0.70
	}

	// 确保初始并发不超过任务总数
	if totalTasks > 0 && initialC > totalTasks {
		initialC = totalTasks
	}

	controller := &AdaptiveConcurrencyController{
		adaptiveType:       adaptiveType,
		totalTasks:         totalTasks,
		minConcurrency:     minC,
		maxConcurrency:     maxC,
		targetCPUUsage:     targetCPU,
		targetMemoryUsage:  targetMem,
		minSuccessRate:     minSuccess,
		currentConcurrency: int32(initialC),
		semaphore:          make(chan struct{}, maxC),
		lastAdjustTime:     time.Now(),
		adjustInterval:     AdjustInterval,
		adjustmentLog:      make([]concurrencyAdjustment, 0),
	}
	// 初始化条件变量（用于动态并发控制）
	controller.cond = sync.NewCond(&controller.mu)

	typeName := "延迟测试"
	if adaptiveType == AdaptiveTypeSpeed {
		typeName = "速度测试"
	}

	utils.Info("[动态并发] 初始化%s控制器: 并发=%d, 范围=[%d, %d], CPU核心=%d",
		typeName, initialC, minC, maxC, numCPU)

	return controller
}

// Acquire 获取一个并发槽位
func (acc *AdaptiveConcurrencyController) Acquire() {
	acc.semaphore <- struct{}{}
}

// Release 释放一个并发槽位
func (acc *AdaptiveConcurrencyController) Release() {
	<-acc.semaphore
}

// TryAcquire 尝试获取并发槽位（非阻塞）
func (acc *AdaptiveConcurrencyController) TryAcquire() bool {
	select {
	case acc.semaphore <- struct{}{}:
		return true
	default:
		return false
	}
}

// GetCurrentConcurrency 获取当前并发数
func (acc *AdaptiveConcurrencyController) GetCurrentConcurrency() int {
	return int(atomic.LoadInt32(&acc.currentConcurrency))
}

// GetActiveCount 获取当前活跃任务数
func (acc *AdaptiveConcurrencyController) GetActiveCount() int {
	return int(atomic.LoadInt32(&acc.activeCount))
}

// AcquireDynamic 动态获取并发槽位
// 与 Acquire 不同，此方法会根据 currentConcurrency 的实时值决定是否阻塞
// 当 MaybeAdjust 增加并发数时，等待中的 goroutine 会被唤醒
func (acc *AdaptiveConcurrencyController) AcquireDynamic() {
	acc.mu.Lock()
	defer acc.mu.Unlock()

	// 等待直到当前活跃数小于允许的并发数
	for atomic.LoadInt32(&acc.activeCount) >= atomic.LoadInt32(&acc.currentConcurrency) {
		acc.cond.Wait()
	}

	// 增加活跃计数
	atomic.AddInt32(&acc.activeCount, 1)
}

// ReleaseDynamic 动态释放并发槽位
// 减少活跃计数并广播唤醒等待的 goroutine
func (acc *AdaptiveConcurrencyController) ReleaseDynamic() {
	atomic.AddInt32(&acc.activeCount, -1)

	// 广播唤醒所有等待者（可能有多个因并发提升而可以启动）
	acc.cond.Broadcast()
}

// AcquireWithDelay 获取并发槽位并等待启动间隔
// 用于平滑任务启动，避免瞬时大量连接导致的资源竞争
func (acc *AdaptiveConcurrencyController) AcquireWithDelay() {
	// 先获取槽位
	acc.AcquireDynamic()

	// 启动间隔，延迟测试和速度测试使用不同的间隔
	var delay time.Duration
	if acc.adaptiveType == AdaptiveTypeLatency {
		delay = LatencyTaskStartInterval
	} else {
		delay = SpeedTaskStartInterval
	}
	time.Sleep(delay)
}

// ReportSuccess 报告任务成功
func (acc *AdaptiveConcurrencyController) ReportSuccess(latencyMs int) {
	atomic.AddInt32(&acc.successCount, 1)
	atomic.AddInt64(&acc.totalLatency, int64(latencyMs))
}

// ReportFailure 报告任务失败
func (acc *AdaptiveConcurrencyController) ReportFailure() {
	atomic.AddInt32(&acc.failCount, 1)
}

// getSuccessRate 获取当前成功率
func (acc *AdaptiveConcurrencyController) getSuccessRate() float64 {
	success := atomic.LoadInt32(&acc.successCount)
	fail := atomic.LoadInt32(&acc.failCount)
	total := success + fail
	if total == 0 {
		return 1.0 // 还没有结果时认为100%成功
	}
	return float64(success) / float64(total)
}

// MaybeAdjust 根据系统状态可能调整并发数
// 返回是否进行了调整
func (acc *AdaptiveConcurrencyController) MaybeAdjust() bool {
	acc.mu.Lock()
	defer acc.mu.Unlock()

	// 检查调整间隔
	if time.Since(acc.lastAdjustTime) < acc.adjustInterval {
		return false
	}

	metrics := GetSystemMetrics()
	successRate := acc.getSuccessRate()
	currentC := int(atomic.LoadInt32(&acc.currentConcurrency))
	numCPU := runtime.NumCPU()

	// 计算每CPU的goroutine数
	goroutinePerCPU := float64(metrics.GoroutineNum) / float64(numCPU)

	var newC int
	var reason string

	// 紧急降低：内存压力过大
	if metrics.MemoryUsage > MemoryEmergencyThreshold {
		newC = int(float64(currentC) * DecreaseEmergencyMultiplier)
		reason = "内存压力过大"
	} else if goroutinePerCPU > GoroutineCriticalThreshold {
		// 紧急降低：goroutine过多
		newC = int(float64(currentC) * DecreaseEmergencyMultiplier)
		reason = fmt.Sprintf("goroutine过多(%.0f/CPU)", goroutinePerCPU)
	} else if goroutinePerCPU > GoroutineHighThreshold {
		// 适度降低：goroutine偏高
		newC = int(float64(currentC) * DecreaseCPUMultiplier)
		reason = fmt.Sprintf("goroutine偏高(%.0f/CPU)", goroutinePerCPU)
	} else if metrics.CPUUsage > acc.targetCPUUsage+CPUHighThresholdOffset {
		// 降低：CPU使用率过高
		newC = int(float64(currentC) * DecreaseCPUMultiplier)
		reason = "CPU使用率过高"
	} else if metrics.MemoryUsage > acc.targetMemoryUsage {
		// 降低：内存使用率偏高
		newC = int(float64(currentC) * DecreaseMemoryMultiplier)
		reason = "内存使用率偏高"
	} else if goroutinePerCPU < GoroutineHighThreshold/2 &&
		metrics.CPUUsage < acc.targetCPUUsage &&
		metrics.MemoryUsage < acc.targetMemoryUsage &&
		currentC < acc.maxConcurrency {
		// 资源充足，可以提升
		newC = int(float64(currentC) * IncreaseMultiplier)
		reason = "系统资源充足，提升并发"
	} else {
		// 无需调整
		return false
	}

	// 边界检查
	if newC < acc.minConcurrency {
		newC = acc.minConcurrency
	}
	if newC > acc.maxConcurrency {
		newC = acc.maxConcurrency
	}
	if newC > acc.totalTasks {
		newC = acc.totalTasks
	}

	// 如果变化太小，忽略
	if abs(newC-currentC) < MinAdjustmentDelta {
		return false
	}

	// 记录调整
	adjustment := concurrencyAdjustment{
		Time:        time.Now(),
		OldValue:    currentC,
		NewValue:    newC,
		Reason:      reason,
		CPUUsage:    metrics.CPUUsage,
		MemoryUsage: metrics.MemoryUsage,
		SuccessRate: successRate,
	}
	acc.adjustmentLog = append(acc.adjustmentLog, adjustment)

	typeName := "延迟"
	if acc.adaptiveType == AdaptiveTypeSpeed {
		typeName = "速度"
	}

	utils.Info("[动态并发] %s测试调整: %d→%d (%s) [CPU: %.0f%%, 内存: %.0f%%, 成功率: %.0f%%]",
		typeName, currentC, newC, reason,
		metrics.CPUUsage*100, metrics.MemoryUsage*100, successRate*100)

	atomic.StoreInt32(&acc.currentConcurrency, int32(newC))
	acc.lastAdjustTime = time.Now()

	// 广播唤醒等待的 goroutine（当并发数增加时，更多任务可以启动）
	acc.cond.Broadcast()

	return true
}

// GetSummary 获取调整统计摘要
func (acc *AdaptiveConcurrencyController) GetSummary() string {
	if len(acc.adjustmentLog) == 0 {
		return "无动态调整"
	}

	var minC, maxC, sumC int
	minC = acc.maxConcurrency
	maxC = acc.minConcurrency

	for _, adj := range acc.adjustmentLog {
		if adj.NewValue < minC {
			minC = adj.NewValue
		}
		if adj.NewValue > maxC {
			maxC = adj.NewValue
		}
		sumC += adj.NewValue
	}

	avgC := sumC / len(acc.adjustmentLog)
	return fmt.Sprintf("调整次数: %d, 平均并发: %d, 范围: [%d, %d]",
		len(acc.adjustmentLog), avgC, minC, maxC)
}

// GetStats 获取统计信息
func (acc *AdaptiveConcurrencyController) GetStats() (success, fail int32, avgLatency int64) {
	success = atomic.LoadInt32(&acc.successCount)
	fail = atomic.LoadInt32(&acc.failCount)
	total := atomic.LoadInt64(&acc.totalLatency)
	if success > 0 {
		avgLatency = total / int64(success)
	}
	return
}

// 辅助函数
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
