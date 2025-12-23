package services

import (
	"sublink/models"
	"sublink/services/scheduler"
	"sublink/services/telegram"
)

// telegramServicesWrapper 实现 telegram.ServicesWrapper 接口
// 用于从 Telegram 回调中调用服务层
type telegramServicesWrapper struct{}

func (w *telegramServicesWrapper) ExecuteSubscriptionTaskWithTrigger(id int, url string, subName string, trigger models.TaskTrigger) {
	scheduler.ExecuteSubscriptionTaskWithTrigger(id, url, subName, trigger)
}

func (w *telegramServicesWrapper) ApplyAutoTagRules(nodes []models.Node, triggerSource string) {
	ApplyAutoTagRules(nodes, triggerSource)
}

func (w *telegramServicesWrapper) CancelTask(taskID string) error {
	return CancelTask(taskID)
}

func (w *telegramServicesWrapper) GetRunningTasks() []models.Task {
	return GetTaskManager().GetRunningTasksInfo()
}

// GetNodeCheckProfiles 获取所有节点检测策略
func (w *telegramServicesWrapper) GetNodeCheckProfiles() ([]models.NodeCheckProfile, error) {
	var profile models.NodeCheckProfile
	return profile.List()
}

// ExecuteNodeCheckWithProfile 使用指定策略执行节点检测
func (w *telegramServicesWrapper) ExecuteNodeCheckWithProfile(profileID int, nodeIDs []int) {
	scheduler.ExecuteNodeCheckWithProfile(profileID, nodeIDs)
}

// ToggleProfileEnabled 开关策略的定时执行
func (w *telegramServicesWrapper) ToggleProfileEnabled(profileID int) (bool, error) {
	profile, err := models.GetNodeCheckProfileByID(profileID)
	if err != nil {
		return false, err
	}

	// 切换启用状态
	newEnabled := !profile.Enabled
	profile.Enabled = newEnabled

	if err := profile.Update(); err != nil {
		return false, err
	}

	// 更新调度器任务
	sch := scheduler.GetSchedulerManager()
	if err := sch.UpdateNodeCheckProfileJob(profile.ID, profile.CronExpr, newEnabled); err != nil {
		return newEnabled, err
	}

	return newEnabled, nil
}

// InitTelegramWrapper 初始化 Telegram 服务包装器
// 在 Telegram 初始化后调用
func InitTelegramWrapper() {
	telegram.SetServicesWrapper(&telegramServicesWrapper{})
}
