package models

import (
	"encoding/json"
	"fmt"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 等待执行
	TaskStatusRunning   TaskStatus = "running"   // 正在执行
	TaskStatusCompleted TaskStatus = "completed" // 已完成
	TaskStatusCancelled TaskStatus = "cancelled" // 已取消
	TaskStatusError     TaskStatus = "error"     // 执行错误
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeSpeedTest TaskType = "speed_test" // 节点测速
	TaskTypeSubUpdate TaskType = "sub_update" // 订阅更新
	TaskTypeTagRule   TaskType = "tag_rule"   // 标签规则
)

// TaskTrigger 任务触发方式
type TaskTrigger string

const (
	TaskTriggerManual    TaskTrigger = "manual"    // 手动触发
	TaskTriggerScheduled TaskTrigger = "scheduled" // 定时触发
)

// Task 任务模型
type Task struct {
	ID          string      `gorm:"primaryKey;size:64" json:"id"`
	Type        TaskType    `gorm:"size:32;index" json:"type"`
	Name        string      `gorm:"size:128" json:"name"`
	Status      TaskStatus  `gorm:"size:20;index" json:"status"`
	Trigger     TaskTrigger `gorm:"size:20;index" json:"trigger"`
	Progress    int         `json:"progress"`                    // 当前进度
	Total       int         `json:"total"`                       // 总数
	CurrentItem string      `gorm:"size:256" json:"currentItem"` // 当前处理项
	Message     string      `gorm:"size:512" json:"message"`     // 状态消息
	Result      string      `gorm:"type:text" json:"result"`     // JSON 格式的结果数据
	StartedAt   *time.Time  `json:"startedAt"`
	CompletedAt *time.Time  `json:"completedAt"`
	CreatedAt   time.Time   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TaskFilter 任务过滤条件
type TaskFilter struct {
	Status  string // 状态过滤
	Type    string // 类型过滤
	Trigger string // 触发方式过滤
}

// taskCache 任务缓存（仅缓存最近的任务）
var taskCache *cache.MapCache[string, Task]

func init() {
	taskCache = cache.NewMapCache(func(t Task) string { return t.ID })
	taskCache.AddIndex("status", func(t Task) string { return string(t.Status) })
	taskCache.AddIndex("type", func(t Task) string { return string(t.Type) })
}

// InitTaskCache 初始化任务缓存
func InitTaskCache() error {
	utils.Info("开始加载任务到缓存")
	var tasks []Task
	// 只加载最近7天的任务到缓存
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := database.DB.Where("created_at > ?", sevenDaysAgo).Find(&tasks).Error; err != nil {
		return err
	}

	taskCache.LoadAll(tasks)
	utils.Info("任务缓存初始化完成，共加载 %d 个任务", taskCache.Count())

	cache.Manager.Register("task", taskCache)
	return nil
}

// GenerateTaskID 生成任务ID
// 格式: {type}_{timestamp}_{random}
func GenerateTaskID(taskType TaskType) string {
	timestamp := time.Now().UnixNano()
	random := fmt.Sprintf("%06d", time.Now().Nanosecond()%1000000)
	return fmt.Sprintf("%s_%d_%s", taskType, timestamp, random)
}

// Create 创建任务 (Write-Through)
func (t *Task) Create() error {
	if t.ID == "" {
		t.ID = GenerateTaskID(t.Type)
	}
	if t.Status == "" {
		t.Status = TaskStatusPending
	}
	err := database.DB.Create(t).Error
	if err != nil {
		return err
	}
	taskCache.Set(t.ID, *t)
	return nil
}

// Update 更新任务 (Write-Through)
// 明确指定需要更新的字段以确保 Status 等字段能正确更新
func (t *Task) Update() error {
	err := database.DB.Model(t).Select(
		"Status", "Progress", "Total", "CurrentItem", "Message", "Result", "CompletedAt", "UpdatedAt",
	).Updates(t).Error
	if err != nil {
		return err
	}
	taskCache.Set(t.ID, *t)
	return nil
}

// SyncFinalStatus 同步任务最终状态到数据库
// 仅在任务结束时调用（完成/失败/取消），一次性同步所有终态字段
// 包括：Status, Progress, Total, CurrentItem, Message, Result, CompletedAt
func (t *Task) SyncFinalStatus() error {
	err := database.DB.Model(t).Select(
		"Status", "Progress", "Total", "CurrentItem", "Message", "Result", "CompletedAt", "UpdatedAt",
	).Updates(t).Error
	if err != nil {
		return err
	}
	taskCache.Set(t.ID, *t)
	return nil
}

// UpdateStatus 更新任务状态
func (t *Task) UpdateStatus(status TaskStatus, message string) error {
	t.Status = status
	t.Message = message
	if status == TaskStatusCompleted || status == TaskStatusCancelled || status == TaskStatusError {
		now := time.Now()
		t.CompletedAt = &now
	}
	err := database.DB.Model(t).Select("Status", "Message", "CompletedAt", "UpdatedAt").Updates(t).Error
	if err != nil {
		return err
	}
	taskCache.Set(t.ID, *t)
	return nil
}

// SetResult 设置任务结果
func (t *Task) SetResult(result interface{}) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}
	t.Result = string(resultJSON)
	return database.DB.Model(t).Select("Result", "UpdatedAt").Updates(t).Error
}

// GetByID 根据ID获取任务
func (t *Task) GetByID(id string) error {
	if cached, ok := taskCache.Get(id); ok {
		*t = cached
		return nil
	}
	return database.DB.Where("id = ?", id).First(t).Error
}

// ListTasks 获取任务列表
func ListTasks(filter TaskFilter, page, pageSize int) ([]Task, int64, error) {
	var tasks []Task
	var total int64

	query := database.DB.Model(&Task{})

	// 应用过滤条件
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Trigger != "" {
		query = query.Where("trigger = ?", filter.Trigger)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，按创建时间倒序
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetRunningTasks 获取运行中的任务
func GetRunningTasks() ([]Task, error) {
	tasks := taskCache.GetByIndex("status", string(TaskStatusRunning))
	if len(tasks) > 0 {
		return tasks, nil
	}
	// 缓存中没有，从数据库查询
	var dbTasks []Task
	if err := database.DB.Where("status = ?", TaskStatusRunning).Find(&dbTasks).Error; err != nil {
		return nil, err
	}
	return dbTasks, nil
}

// GetPendingTasks 获取等待中的任务
func GetPendingTasks() ([]Task, error) {
	tasks := taskCache.GetByIndex("status", string(TaskStatusPending))
	if len(tasks) > 0 {
		return tasks, nil
	}
	var dbTasks []Task
	if err := database.DB.Where("status = ?", TaskStatusPending).Find(&dbTasks).Error; err != nil {
		return nil, err
	}
	return dbTasks, nil
}

// CleanupOldTasks 清理旧任务
func CleanupOldTasks(before time.Time) (int64, error) {
	result := database.DB.Where("created_at < ? AND status IN ?", before, []TaskStatus{TaskStatusCompleted, TaskStatusCancelled, TaskStatusError}).Delete(&Task{})
	if result.Error != nil {
		return 0, result.Error
	}

	// 只有当清理范围与缓存范围（7天内）有交集时才刷新缓存
	// 缓存只保存最近7天的任务，如果清理的是7天之前的任务，则无需刷新
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if result.RowsAffected > 0 && before.After(sevenDaysAgo) {
		go func() {
			_ = InitTaskCache()
		}()
	}
	return result.RowsAffected, nil
}

// GetTaskStats 获取任务统计
func GetTaskStats() map[string]int64 {
	stats := make(map[string]int64)

	// 从数据库获取统计
	var results []struct {
		Status TaskStatus
		Count  int64
	}
	database.DB.Model(&Task{}).Select("status, count(*) as count").Group("status").Scan(&results)

	for _, r := range results {
		stats[string(r.Status)] = r.Count
	}

	// 确保所有状态都有值
	for _, status := range []TaskStatus{TaskStatusPending, TaskStatusRunning, TaskStatusCompleted, TaskStatusCancelled, TaskStatusError} {
		if _, ok := stats[string(status)]; !ok {
			stats[string(status)] = 0
		}
	}

	return stats
}

// MarkRunningTasksAsError 将所有运行中的任务标记为错误（用于服务重启时）
func MarkRunningTasksAsError() error {
	return database.DB.Model(&Task{}).
		Where("status = ?", TaskStatusRunning).
		Updates(map[string]interface{}{
			"status":  TaskStatusError,
			"message": "服务重启，任务被中断",
		}).Error
}
