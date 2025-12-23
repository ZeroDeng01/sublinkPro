package scheduler

import (
	"context"
	"sublink/models"
)

// ================================================================================
// 依赖注入接口定义
// 这些接口由外部 services 包实现，在程序启动时注入到 scheduler 包
// 这种设计避免了 scheduler -> services -> scheduler 的循环导入
// ================================================================================

// AdaptiveConcurrencyControllerFactory 创建自适应并发控制器的工厂函数类型
// 参数:
//   - adaptiveType: 控制器类型，使用 AdaptiveTypeLatency 或 AdaptiveTypeSpeed
//   - totalTasks: 总任务数量，用于计算合适的初始并发数
type AdaptiveConcurrencyControllerFactory func(adaptiveType int, totalTasks int) AdaptiveConcurrencyController

// AdaptiveConcurrencyController 自适应并发控制器接口
// 用于动态调整测速任务的并发数，根据系统负载和任务成功率自动伸缩
type AdaptiveConcurrencyController interface {
	// GetCurrentConcurrency 获取当前并发数
	GetCurrentConcurrency() int
	// AcquireWithDelay 获取并发槽位（带启动延迟，避免瞬时大量连接）
	AcquireWithDelay()
	// ReleaseDynamic 释放并发槽位
	ReleaseDynamic()
	// ReportSuccess 报告任务成功，latencyMs 为延迟时间（单位：毫秒）
	ReportSuccess(latencyMs int)
	// ReportFailure 报告任务失败
	ReportFailure()
	// MaybeAdjust 检查并可能调整并发数，返回是否进行了调整
	MaybeAdjust() bool
}

// TaskManagerInterface TaskManager 接口
// 用于管理后台任务的生命周期和进度报告
type TaskManagerInterface interface {
	// UpdateTotal 更新任务总数
	UpdateTotal(taskID string, total int) error
	// UpdateProgress 更新任务进度
	UpdateProgress(taskID string, progress int, currentItem string, result interface{}) error
	// CompleteTask 标记任务完成
	CompleteTask(taskID string, message string, result interface{}) error
	// FailTask 标记任务失败
	FailTask(taskID string, errMsg string) error
	// CreateTask 创建新任务
	CreateTask(taskType models.TaskType, name string, trigger models.TaskTrigger, total int) (*models.Task, context.Context, error)
}

// AutoTagRulesApplier 自动标签规则应用函数类型
// 当测速或订阅更新完成后，自动为节点应用匹配的标签规则
type AutoTagRulesApplier func(nodes []models.Node, source string)

// ================================================================================
// 依赖注入存储
// ================================================================================

var (
	// adaptiveConcurrencyFactory 自适应并发控制器工厂函数
	// 由 services.InitSchedulerDependencies() 注入
	adaptiveConcurrencyFactory AdaptiveConcurrencyControllerFactory

	// taskManagerGetter 任务管理器获取函数
	// 由 services.InitSchedulerDependencies() 注入
	taskManagerGetter func() TaskManagerInterface

	// autoTagRulesApplier 自动标签规则应用函数
	// 由 services.InitSchedulerDependencies() 注入
	autoTagRulesApplier AutoTagRulesApplier

	// latencyAdjustInterval 延迟测试并发调整检查间隔
	// 每完成这么多个任务后检查一次是否需要调整并发数
	// 单位：任务个数
	// 默认值：5（即每完成5个延迟测试任务检查一次）
	latencyAdjustInterval = 5

	// speedAdjustInterval 速度测试并发调整检查间隔
	// 每完成这么多个任务后检查一次是否需要调整并发数
	// 单位：任务个数
	// 默认值：3（速度测试检查更频繁，因为对系统资源影响更大）
	speedAdjustInterval = 3
)

// ================================================================================
// 自适应控制器类型常量
// ================================================================================

const (
	// AdaptiveTypeLatency 延迟测试控制器类型
	// 用于延迟测试阶段，特点：高并发、I/O 密集型
	// 初始并发数较高，可达到 CPU 核心数 * 5
	AdaptiveTypeLatency = 0

	// AdaptiveTypeSpeed 速度测试控制器类型
	// 用于速度测试阶段，特点：低并发、带宽密集型
	// 初始并发数较低，避免多个测速任务竞争带宽导致结果不准确
	AdaptiveTypeSpeed = 1
)

// ================================================================================
// 依赖注入函数（由 services 包在初始化时调用）
// ================================================================================

// InjectDependencies 注入 scheduler 包所需的外部依赖
// 必须在使用 scheduler 功能前调用（通常在 main.go 中由 services.InitSchedulerDependencies() 调用）
// 参数:
//   - factory: 自适应并发控制器工厂函数
//   - tmGetter: 任务管理器获取函数
//   - tagApplier: 自动标签规则应用函数
//   - latencyInterval: 延迟测试并发调整检查间隔（单位：任务个数）
//   - speedInterval: 速度测试并发调整检查间隔（单位：任务个数）
func InjectDependencies(
	factory AdaptiveConcurrencyControllerFactory,
	tmGetter func() TaskManagerInterface,
	tagApplier AutoTagRulesApplier,
	latencyInterval, speedInterval int,
) {
	adaptiveConcurrencyFactory = factory
	taskManagerGetter = tmGetter
	autoTagRulesApplier = tagApplier
	latencyAdjustInterval = latencyInterval
	speedAdjustInterval = speedInterval
	// 同步更新导出变量
	latencyAdjustCheckInterval = latencyInterval
	speedAdjustCheckInterval = speedInterval
}

// ================================================================================
// 内部辅助函数（使用注入的依赖）
// ================================================================================

// newAdaptiveConcurrencyController 创建自适应并发控制器
// 如果依赖未注入会触发 panic
func newAdaptiveConcurrencyController(adaptiveType int, totalTasks int) AdaptiveConcurrencyController {
	if adaptiveConcurrencyFactory == nil {
		panic("scheduler: AdaptiveConcurrencyControllerFactory not injected, call InjectDependencies first")
	}
	return adaptiveConcurrencyFactory(adaptiveType, totalTasks)
}

// getTaskManager 获取任务管理器
// 如果依赖未注入会触发 panic
func getTaskManager() TaskManagerInterface {
	if taskManagerGetter == nil {
		panic("scheduler: TaskManager getter not injected, call InjectDependencies first")
	}
	return taskManagerGetter()
}

// applyAutoTagRules 应用自动标签规则
// 如果未注入则静默跳过
func applyAutoTagRules(nodes []models.Node, source string) {
	if autoTagRulesApplier != nil {
		autoTagRulesApplier(nodes, source)
	}
}

// ================================================================================
// 导出的调整间隔变量（供 speedtest_task.go 使用）
// ================================================================================

var (
	// latencyAdjustCheckInterval 延迟测试并发调整检查间隔
	// 单位：任务个数
	// 在 RunSpeedTestWithConfig 中使用，每完成该数量的延迟测试后检查是否需要调整并发
	latencyAdjustCheckInterval = 5

	// speedAdjustCheckInterval 速度测试并发调整检查间隔
	// 单位：任务个数
	// 在 RunSpeedTestWithConfig 中使用，每完成该数量的速度测试后检查是否需要调整并发
	speedAdjustCheckInterval = 3
)
