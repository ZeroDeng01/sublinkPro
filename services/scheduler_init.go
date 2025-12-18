package services

import (
	"context"
	"sublink/models"
	"sublink/services/scheduler"
)

// InitSchedulerDependencies 初始化 scheduler 包的依赖
// 必须在使用 scheduler 功能前调用（通常在 main.go 中）
func InitSchedulerDependencies() {
	// 创建自适应并发控制器的工厂函数
	factory := func(adaptiveType int, totalTasks int) scheduler.AdaptiveConcurrencyController {
		var at AdaptiveType
		if adaptiveType == scheduler.AdaptiveTypeLatency {
			at = AdaptiveTypeLatency
		} else {
			at = AdaptiveTypeSpeed
		}
		return NewAdaptiveConcurrencyController(at, totalTasks)
	}

	// 创建 TaskManager 获取函数
	tmGetter := func() scheduler.TaskManagerInterface {
		return &taskManagerAdapter{tm: GetTaskManager()}
	}

	// 注入依赖
	scheduler.InjectDependencies(
		factory,
		tmGetter,
		ApplyAutoTagRules,
		LatencyAdjustCheckInterval,
		SpeedAdjustCheckInterval,
	)
}

// taskManagerAdapter 适配 TaskManager 到 scheduler.TaskManagerInterface
type taskManagerAdapter struct {
	tm *TaskManager
}

func (a *taskManagerAdapter) UpdateTotal(taskID string, total int) error {
	return a.tm.UpdateTotal(taskID, total)
}

func (a *taskManagerAdapter) UpdateProgress(taskID string, progress int, currentItem string, result interface{}) error {
	return a.tm.UpdateProgress(taskID, progress, currentItem, result)
}

func (a *taskManagerAdapter) CompleteTask(taskID string, message string, result interface{}) error {
	return a.tm.CompleteTask(taskID, message, result)
}

func (a *taskManagerAdapter) FailTask(taskID string, errMsg string) error {
	return a.tm.FailTask(taskID, errMsg)
}

func (a *taskManagerAdapter) CreateTask(taskType models.TaskType, name string, trigger models.TaskTrigger, total int) (*models.Task, context.Context, error) {
	return a.tm.CreateTask(taskType, name, trigger, total)
}
