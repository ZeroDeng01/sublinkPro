package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sublink/models"
	"sublink/services/sse"
	"sublink/utils"
	"sync"
	"time"
)

// RunningTask 运行中的任务信息
type RunningTask struct {
	Task      *models.Task
	Context   context.Context
	Cancel    context.CancelFunc
	StartTime time.Time
}

// TaskManager 任务管理器
type TaskManager struct {
	runningTasks map[string]*RunningTask
	mutex        sync.RWMutex
}

var (
	taskManager     *TaskManager
	taskManagerOnce sync.Once
)

// GetTaskManager 获取任务管理器单例
func GetTaskManager() *TaskManager {
	taskManagerOnce.Do(func() {
		taskManager = &TaskManager{
			runningTasks: make(map[string]*RunningTask),
		}
	})
	return taskManager
}

// CreateTask 创建新任务
// 返回任务对象和可取消的 context
func (tm *TaskManager) CreateTask(taskType models.TaskType, name string, trigger models.TaskTrigger, total int) (*models.Task, context.Context, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// 创建任务对象
	now := time.Now()
	task := &models.Task{
		ID:        models.GenerateTaskID(taskType),
		Type:      taskType,
		Name:      name,
		Status:    models.TaskStatusRunning,
		Trigger:   trigger,
		Progress:  0,
		Total:     total,
		StartedAt: &now,
	}

	// 保存到数据库
	if err := task.Create(); err != nil {
		return nil, nil, err
	}

	// 创建可取消的 context
	ctx, cancel := context.WithCancel(context.Background())

	// 注册到运行中任务
	tm.runningTasks[task.ID] = &RunningTask{
		Task:      task,
		Context:   ctx,
		Cancel:    cancel,
		StartTime: now,
	}

	// 广播任务开始
	tm.broadcastProgress(task, "started")

	utils.Info("任务创建成功: ID=%s, Type=%s, Name=%s, Trigger=%s", task.ID, task.Type, task.Name, task.Trigger)

	return task, ctx, nil
}

// UpdateProgress 更新任务进度
// 仅更新内存状态并通过 SSE 广播，不写入数据库
// 数据库同步延迟到任务结束时统一处理
func (tm *TaskManager) UpdateProgress(taskID string, progress int, currentItem string, result interface{}) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	running, ok := tm.runningTasks[taskID]
	if !ok {
		return errors.New("任务不存在或已结束")
	}

	// 仅更新内存中的任务状态
	running.Task.Progress = progress
	running.Task.CurrentItem = currentItem

	// 仅通过 SSE 广播进度（不写数据库）
	tm.broadcastProgressWithResult(running.Task, "progress", result)

	return nil
}

// UpdateTotal 更新任务总数（在任务开始后确定总数时使用）
// 仅更新内存状态并通过 SSE 广播，不写入数据库
func (tm *TaskManager) UpdateTotal(taskID string, total int) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	running, ok := tm.runningTasks[taskID]
	if !ok {
		return errors.New("任务不存在或已结束")
	}

	// 仅更新内存中的任务总数
	running.Task.Total = total

	// 广播 total 更新（不写数据库）
	tm.broadcastProgressWithResult(running.Task, "progress", nil)

	return nil
}

// CompleteTask 完成任务
func (tm *TaskManager) CompleteTask(taskID string, message string, result interface{}) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	running, ok := tm.runningTasks[taskID]
	if !ok {
		return errors.New("任务不存在或已结束")
	}

	// 更新状态
	running.Task.Status = models.TaskStatusCompleted
	running.Task.Message = message
	running.Task.Progress = running.Task.Total
	now := time.Now()
	running.Task.CompletedAt = &now

	// 设置结果
	if result != nil {
		resultJSON, _ := json.Marshal(result)
		running.Task.Result = string(resultJSON)
	}

	// 同步最终状态到数据库（任务结束时一次性写入）
	if err := running.Task.SyncFinalStatus(); err != nil {
		utils.Error("同步任务最终状态失败: %v", err)
	}

	// 广播完成进度（前端用）
	tm.broadcastProgressWithResult(running.Task, "completed", result)

	// 注意：不在此处广播通知事件，由各任务类型自行发送详细的通知
	// 这样可以避免重复通知，且各任务类型可以发送更具体的信息
	// 例如：速测任务发送成功/失败数量，订阅任务发送新增/删除数量

	// 取消 context（确保所有 goroutine 退出）
	running.Cancel()

	// 从运行中任务移除（延迟移除，给 SSE 时间广播）
	go func() {
		time.Sleep(500 * time.Millisecond)
		tm.mutex.Lock()
		delete(tm.runningTasks, taskID)
		tm.mutex.Unlock()
	}()

	utils.Info("任务完成: ID=%s, Message=%s", taskID, message)

	return nil
}

// FailTask 任务失败
func (tm *TaskManager) FailTask(taskID string, errMsg string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	running, ok := tm.runningTasks[taskID]
	if !ok {
		return errors.New("任务不存在或已结束")
	}

	// 更新状态
	running.Task.Status = models.TaskStatusError
	running.Task.Message = errMsg
	now := time.Now()
	running.Task.CompletedAt = &now

	// 同步最终状态到数据库（任务结束时一次性写入）
	if err := running.Task.SyncFinalStatus(); err != nil {
		utils.Error("同步任务最终状态失败: %v", err)
	}

	// 广播错误进度（前端用）
	tm.broadcastProgress(running.Task, "error")

	// 注意：不在此处广播通知事件，由各任务类型自行发送详细的错误通知
	// 这样可以避免重复通知，且各任务类型可以提供更具体的错误上下文

	// 取消 context
	running.Cancel()

	// 从运行中任务移除
	go func() {
		time.Sleep(500 * time.Millisecond)
		tm.mutex.Lock()
		delete(tm.runningTasks, taskID)
		tm.mutex.Unlock()
	}()

	utils.Warn("任务失败: ID=%s, Error=%s", taskID, errMsg)

	return nil
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(taskID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	running, ok := tm.runningTasks[taskID]
	if !ok {
		// 检查是否在数据库中存在且仍在运行
		var task models.Task
		if err := task.GetByID(taskID); err != nil {
			return errors.New("任务不存在")
		}
		if task.Status != models.TaskStatusRunning && task.Status != models.TaskStatusPending {
			return errors.New("任务已结束，无法取消")
		}
		// 更新数据库状态
		task.Status = models.TaskStatusCancelled
		task.Message = "用户取消"
		now := time.Now()
		task.CompletedAt = &now
		return task.SyncFinalStatus()
	}

	// 更新状态
	running.Task.Status = models.TaskStatusCancelled
	running.Task.Message = "用户取消"
	now := time.Now()
	running.Task.CompletedAt = &now

	// 同步最终状态到数据库（任务结束时一次性写入）
	if err := running.Task.SyncFinalStatus(); err != nil {
		utils.Error("同步任务最终状态失败: %v", err)
	}

	// 广播取消
	tm.broadcastProgress(running.Task, "cancelled")

	// 取消 context（这会通知所有监听的 goroutine）
	running.Cancel()

	// 从运行中任务移除
	go func() {
		time.Sleep(500 * time.Millisecond)
		tm.mutex.Lock()
		delete(tm.runningTasks, taskID)
		tm.mutex.Unlock()
	}()

	utils.Info("任务已取消: ID=%s", taskID)

	return nil
}

// IsTaskCancelled 检查任务是否已取消
func (tm *TaskManager) IsTaskCancelled(taskID string) bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	running, ok := tm.runningTasks[taskID]
	if !ok {
		return true // 任务不存在，视为已取消
	}

	select {
	case <-running.Context.Done():
		return true
	default:
		return false
	}
}

// GetContext 获取任务的 context
func (tm *TaskManager) GetContext(taskID string) (context.Context, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	running, ok := tm.runningTasks[taskID]
	if !ok {
		return nil, false
	}

	return running.Context, true
}

// GetRunningTasks 获取所有运行中的任务
func (tm *TaskManager) GetRunningTasks() []*RunningTask {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	tasks := make([]*RunningTask, 0, len(tm.runningTasks))
	for _, t := range tm.runningTasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// GetRunningTasksInfo 获取运行中任务的信息（不含 context）
func (tm *TaskManager) GetRunningTasksInfo() []models.Task {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	tasks := make([]models.Task, 0, len(tm.runningTasks))
	for _, t := range tm.runningTasks {
		tasks = append(tasks, *t.Task)
	}
	return tasks
}

// CleanupTask 清理任务（从内存中移除）
func (tm *TaskManager) CleanupTask(taskID string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if running, ok := tm.runningTasks[taskID]; ok {
		// 取消 context
		running.Cancel()
		delete(tm.runningTasks, taskID)
	}
}

// broadcastProgress 广播任务进度到 SSE
func (tm *TaskManager) broadcastProgress(task *models.Task, status string) {
	tm.broadcastProgressWithResult(task, status, nil)
}

// broadcastProgressWithResult 广播任务进度到 SSE（带结果）
func (tm *TaskManager) broadcastProgressWithResult(task *models.Task, status string, result interface{}) {
	startTimeMs := int64(0)
	if task.StartedAt != nil {
		startTimeMs = task.StartedAt.UnixMilli()
	}

	sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
		TaskID:      task.ID,
		TaskType:    string(task.Type),
		TaskName:    task.Name,
		Status:      status,
		Current:     task.Progress,
		Total:       task.Total,
		CurrentItem: task.CurrentItem,
		Result:      result,
		Message:     task.Message,
		StartTime:   startTimeMs,
	})
}

// BroadcastEvent 广播任务事件（用于完成通知等）
func (tm *TaskManager) BroadcastEvent(task *models.Task, eventType string, data map[string]interface{}) {
	sse.GetSSEBroker().BroadcastEvent("task_update", sse.NotificationPayload{
		Event:   eventType,
		Title:   fmt.Sprintf("%s - %s", task.Name, getTaskStatusTitle(task.Status)),
		Message: task.Message,
		Data:    data,
	})
}

// getTaskStatusTitle 获取任务状态标题
func getTaskStatusTitle(status models.TaskStatus) string {
	switch status {
	case models.TaskStatusPending:
		return "等待中"
	case models.TaskStatusRunning:
		return "执行中"
	case models.TaskStatusCompleted:
		return "已完成"
	case models.TaskStatusCancelled:
		return "已取消"
	case models.TaskStatusError:
		return "执行失败"
	default:
		return "未知"
	}
}

// InitTaskManager 初始化任务管理器（服务启动时调用）
func InitTaskManager() {
	tm := GetTaskManager()
	utils.Info("任务管理器初始化完成")

	// 将所有运行中的任务标记为错误（防止服务重启后状态不一致）
	if err := models.MarkRunningTasksAsError(); err != nil {
		utils.Error("标记运行中任务为错误失败: %v", err)
	}

	// 清理过期任务（保留最近30天）
	go func() {
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		if affected, err := models.CleanupOldTasks(thirtyDaysAgo); err != nil {
			utils.Error("清理过期任务失败: %v", err)
		} else if affected > 0 {
			utils.Info("已清理 %d 个过期任务", affected)
		}
	}()

	_ = tm // 确保初始化
}

// CancelTask 取消任务的包装函数
func CancelTask(taskID string) error {
	return GetTaskManager().CancelTask(taskID)
}
