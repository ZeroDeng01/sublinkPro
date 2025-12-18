package scheduler

import (
	"sublink/node"
)

// TaskManagerReporter 实现 node.TaskReporter 接口，用于将任务进度报告给 TaskManager
type TaskManagerReporter struct {
	tm     TaskManagerInterface
	taskID string
}

// NewTaskManagerReporter 创建 TaskManagerReporter
func NewTaskManagerReporter(tm TaskManagerInterface, taskID string) *TaskManagerReporter {
	return &TaskManagerReporter{
		tm:     tm,
		taskID: taskID,
	}
}

func (r *TaskManagerReporter) UpdateTotal(total int) {
	r.tm.UpdateTotal(r.taskID, total)
}

func (r *TaskManagerReporter) ReportProgress(current int, currentItem string, result interface{}) {
	r.tm.UpdateProgress(r.taskID, current, currentItem, result)
}

func (r *TaskManagerReporter) ReportComplete(message string, result interface{}) {
	r.tm.CompleteTask(r.taskID, message, result)
}

func (r *TaskManagerReporter) ReportFail(errMsg string) {
	r.tm.FailTask(r.taskID, errMsg)
}

// 确保 TaskManagerReporter 实现了 node.TaskReporter 接口
var _ node.TaskReporter = (*TaskManagerReporter)(nil)
