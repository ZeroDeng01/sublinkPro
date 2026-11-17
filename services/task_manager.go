package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sublink/models"
	"sublink/services/notifications"
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

	// 所有任务统一使用 24 小时超时（全局最大时长）
	// 实际的超时控制由僵尸任务检测器根据任务类型细分管理
	timeout := 24 * time.Hour

	// 创建带超时的可取消 context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	// 注册到运行中任务
	tm.runningTasks[task.ID] = &RunningTask{
		Task:      task,
		Context:   ctx,
		Cancel:    cancel,
		StartTime: now,
	}

	// 广播任务开始
	tm.broadcastProgress(task, "started")

	utils.Info("Task created: ID=%s, Type=%s, Name=%s, Trigger=%s, Timeout=%v", task.ID, task.Type, task.Name, task.Trigger, timeout)

	return task, ctx, nil
}

// UpdateProgress 更新任务进度
// 仅更新内存状态并通过 SSE 广播，不写入数据库
// 数据库同步延迟到任务结束时统一处理
func (tm *TaskManager) UpdateProgress(taskID string, progress int, currentItem string, result any) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	running, ok := tm.runningTasks[taskID]
	if !ok {
		return errors.New("task does not exist or has already finished")
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
		return errors.New("task does not exist or has already finished")
	}

	// 仅更新内存中的任务总数
	running.Task.Total = total

	// 广播 total 更新（不写数据库）
	tm.broadcastProgressWithResult(running.Task, "progress", nil)

	return nil
}

// CompleteTask 完成任务
func (tm *TaskManager) CompleteTask(taskID string, message string, result any) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	running, ok := tm.runningTasks[taskID]
	if !ok {
		return errors.New("task does not exist or has already finished")
	}

	// 更新状态
	running.Task.Status = models.TaskStatusCompleted
	running.Task.Message = message
	running.Task.Progress = running.Task.Total
	running.Task.CompletedAt = new(time.Now())

	// 设置结果
	if result != nil {
		resultJSON, _ := json.Marshal(result)
		running.Task.Result = string(resultJSON)
	}

	// 同步最终状态到数据库（任务结束时一次性写入）
	if err := running.Task.SyncFinalStatus(); err != nil {
		utils.Error("Failed to sync final task state: %v", err)
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

	utils.Info("Task completed: ID=%s, Message=%s", taskID, message)

	return nil
}

// FailTask 任务失败
func (tm *TaskManager) FailTask(taskID string, errMsg string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	running, ok := tm.runningTasks[taskID]
	if !ok {
		return errors.New("task does not exist or has already finished")
	}

	// 更新状态
	running.Task.Status = models.TaskStatusError
	running.Task.Message = errMsg
	running.Task.CompletedAt = new(time.Now())

	// 同步最终状态到数据库（任务结束时一次性写入）
	if err := running.Task.SyncFinalStatus(); err != nil {
		utils.Error("Failed to sync final task state: %v", err)
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

	utils.Warn("Task failed: ID=%s, Error=%s", taskID, errMsg)

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
			return errors.New("task not found")
		}
		if task.Status != models.TaskStatusRunning && task.Status != models.TaskStatusPending {
			return errors.New("task already finished and cannot be cancelled")
		}
		// 更新数据库状态
		task.Status = models.TaskStatusCancelled
		task.Message = "Cancelled by user"
		task.CompletedAt = new(time.Now())
		if err := task.SyncFinalStatus(); err != nil {
			return err
		}
		// 内存中没有执行体，仍需广播，否则前端会一直显示执行中
		tm.broadcastProgress(&task, "cancelled")
		utils.Info("Task cancelled (database state only): ID=%s", taskID)
		return nil
	}

	// 更新状态
	running.Task.Status = models.TaskStatusCancelled
	running.Task.Message = "Cancelled by user"
	running.Task.CompletedAt = new(time.Now())

	// 同步最终状态到数据库（任务结束时一次性写入）
	if err := running.Task.SyncFinalStatus(); err != nil {
		utils.Error("Failed to sync final task state: %v", err)
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

	utils.Info("Task cancelled: ID=%s", taskID)

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
func (tm *TaskManager) broadcastProgressWithResult(task *models.Task, status string, result any) {
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
func (tm *TaskManager) BroadcastEvent(task *models.Task, eventType string, data map[string]any) {
	notifications.Publish(eventType, notifications.Payload{
		Title:   fmt.Sprintf("%s - %s", task.Name, getTaskStatusTitle(task.Status)),
		Message: task.Message,
		Data:    data,
	})
}

// getTaskStatusTitle 获取任务状态标题
func getTaskStatusTitle(status models.TaskStatus) string {
	switch status {
	case models.TaskStatusPending:
		return "Pending"
	case models.TaskStatusRunning:
		return "Running"
	case models.TaskStatusCompleted:
		return "Completed"
	case models.TaskStatusCancelled:
		return "Cancelled"
	case models.TaskStatusError:
		return "Failed"
	default:
		return "Unknown"
	}
}

// InitTaskManager 初始化任务管理器（服务启动时调用）
func InitTaskManager() {
	tm := GetTaskManager()
	utils.Info("Task manager initialized")

	// 将所有运行中的任务标记为错误（防止服务重启后状态不一致）
	if err := models.MarkRunningTasksAsError(); err != nil {
		utils.Error("Failed to mark running tasks as failed: %v", err)
	}

	// 清理过期任务（保留最近30天）
	go func() {
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		if affected, err := models.CleanupOldTasks(thirtyDaysAgo); err != nil {
			utils.Error("Failed to clean up expired tasks: %v", err)
		} else if affected > 0 {
			utils.Info("Cleaned up %d expired tasks", affected)
		}
	}()

	// 启动僵尸任务检测器
	go tm.startZombieTaskDetector()

	_ = tm // 确保初始化
}

// startZombieTaskDetector 启动僵尸任务检测器
func (tm *TaskManager) startZombieTaskDetector() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	utils.Info("Zombie task detector started")

	for range ticker.C {
		tm.detectAndCleanZombieTasks()
	}
}

// taskMaxRuntime 返回任务类型允许的最长运行时间。
// 订阅更新是有界的短任务（HTTP 客户端自带 30 秒超时），不应跑满一天；
// 测速、节点检测等批量任务仍沿用 24 小时上限。
func taskMaxRuntime(taskType models.TaskType) time.Duration {
	switch taskType {
	case models.TaskTypeSubUpdate:
		return 15 * time.Minute
	default:
		return 24 * time.Hour
	}
}

// detectAndCleanZombieTasks 检测并清理僵尸任务：
// 1. 内存中运行时长超过该类型上限的任务；
// 2. 数据库中仍为 running、但内存里已不存在执行体的孤儿任务（协程异常退出、进程重启未清理）。
func (tm *TaskManager) detectAndCleanZombieTasks() {
	tm.mutex.RLock()
	now := time.Now()
	type zombieTask struct {
		ID       string
		Type     models.TaskType
		Name     string
		Duration time.Duration
		Max      time.Duration
	}
	zombieTasks := make([]zombieTask, 0)
	// 记录内存中的任务 ID，供孤儿任务判断使用
	inMemory := make(map[string]struct{}, len(tm.runningTasks))

	for taskID, taskInfo := range tm.runningTasks {
		inMemory[taskID] = struct{}{}
		duration := now.Sub(taskInfo.StartTime)
		maxRuntime := taskMaxRuntime(taskInfo.Task.Type)
		if duration > maxRuntime {
			zombieTasks = append(zombieTasks, zombieTask{
				ID:       taskID,
				Type:     taskInfo.Task.Type,
				Name:     taskInfo.Task.Name,
				Duration: duration,
				Max:      maxRuntime,
			})
		}
	}
	tm.mutex.RUnlock()

	// 清理内存中的超时任务
	for _, zombie := range zombieTasks {
		utils.Warn("Detected zombie task (running over %s): ID=%s, Type=%s, Name=%s, runtime=%v",
			zombie.Max, zombie.ID, zombie.Type, zombie.Name, zombie.Duration)
		if err := tm.FailTask(zombie.ID, fmt.Sprintf("Task timed out (exceeded %s) and was marked as failed automatically", zombie.Max)); err != nil {
			utils.Error("Error while marking zombie task: task_id=%s, error=%v", zombie.ID, err)
		}
	}

	tm.reconcileOrphanedRunningTasks(inMemory)
}

// orphanTaskGrace 孤儿任务的宽限期，避免与刚创建尚未登记到内存的任务重叠。
const orphanTaskGrace = 2 * time.Minute

// reconcileOrphanedRunningTasks 把数据库里残留的 running 任务标记为中断，
// 否则它们会永远卡在“执行中”，既不会结束也无法取消。
func (tm *TaskManager) reconcileOrphanedRunningTasks(inMemory map[string]struct{}) {
	orphans, err := models.ListRunningTasksCreatedBefore(time.Now().Add(-orphanTaskGrace))
	if err != nil {
		utils.Error("Failed to query orphan tasks: %v", err)
		return
	}

	for i := range orphans {
		if _, ok := inMemory[orphans[i].ID]; ok {
			continue
		}
		if err := orphans[i].MarkTaskInterrupted("Task interrupted (the running process exited or the task expired)"); err != nil {
			utils.Error("Failed to mark orphan task: task_id=%s, error=%v", orphans[i].ID, err)
			continue
		}
		utils.Warn("Cleaned up orphan task: ID=%s, Type=%s, Name=%s", orphans[i].ID, orphans[i].Type, orphans[i].Name)
		tm.broadcastProgress(&orphans[i], "error")
	}
}

// CancelTask 取消任务的包装函数
func CancelTask(taskID string) error {
	return GetTaskManager().CancelTask(taskID)
}
