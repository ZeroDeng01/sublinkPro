package services

import (
	"sublink/models"
	"sublink/services/telegram"
)

// telegramServicesWrapper 实现 telegram.ServicesWrapper 接口
// 用于从 Telegram 回调中调用服务层
type telegramServicesWrapper struct{}

func (w *telegramServicesWrapper) RunSpeedTestOnNodes(nodes []models.Node) {
	RunSpeedTestOnNodes(nodes)
}

func (w *telegramServicesWrapper) ExecuteScheduledSpeedTest() {
	ExecuteNodeSpeedTestTask()
}

func (w *telegramServicesWrapper) ExecuteSubscriptionTaskWithTrigger(id int, url string, subName string, trigger models.TaskTrigger) {
	ExecuteSubscriptionTaskWithTrigger(id, url, subName, trigger)
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

// InitTelegramWrapper 初始化 Telegram 服务包装器
// 在 Telegram 初始化后调用
func InitTelegramWrapper() {
	telegram.SetServicesWrapper(&telegramServicesWrapper{})
}
