package scheduler

import (
	"fmt"
	"sublink/models"
	"sublink/node"
	"sublink/services/sse"
	"sublink/utils"
)

// ExecuteSubscriptionTask 执行订阅任务的具体业务逻辑
func ExecuteSubscriptionTask(id int, url string, subName string) {
	ExecuteSubscriptionTaskWithTrigger(id, url, subName, models.TaskTriggerScheduled)
}

// ExecuteSubscriptionTaskWithTrigger 执行订阅任务（带触发类型）
func ExecuteSubscriptionTaskWithTrigger(id int, url string, subName string, trigger models.TaskTrigger) {
	utils.Info("执行自动获取订阅任务 - ID: %d, Name: %s, URL: %s, Trigger: %s", id, subName, url, trigger)

	// 获取最新的机场配置，以便使用最新的代理设置
	var downloadWithProxy bool
	var proxyLink string
	var userAgent string
	var fetchUsageInfo bool
	var skipTLSVerify bool

	airport, err := models.GetAirportByID(id)
	if err != nil {
		utils.Warn("获取机场配置失败 ID: %d, 使用默认设置: %v", id, err)
	} else {
		downloadWithProxy = airport.DownloadWithProxy
		proxyLink = airport.ProxyLink
		userAgent = airport.UserAgent
		fetchUsageInfo = airport.FetchUsageInfo
		skipTLSVerify = airport.SkipTLSVerify
	}

	// 创建 TaskManager 任务和报告器
	tm := getTaskManager()
	task, _, createErr := tm.CreateTask(models.TaskTypeSubUpdate, subName, trigger, 0)

	var reporter node.TaskReporter
	if createErr != nil {
		utils.Warn("创建订阅更新任务失败: %v，将使用降级模式", createErr)
		reporter = nil // 使用 nil，将在 sub.go 中降级为 NoOpTaskReporter
	} else {
		reporter = NewTaskManagerReporter(tm, task.ID)
	}

	usageInfo, err := node.LoadClashConfigFromURLWithReporter(id, url, subName, downloadWithProxy, proxyLink, userAgent, reporter, fetchUsageInfo, skipTLSVerify)
	if err != nil {
		// 仅在失败时发送通知，成功通知由 node/sub.go 中的 scheduleClashToNodeLinks 发送
		// 这样可以避免重复通知，且成功通知包含更详细的节点统计信息
		if reporter != nil {
			reporter.ReportFail(err.Error())
		}
		sse.GetSSEBroker().BroadcastEvent("task_update", sse.NotificationPayload{
			Event:   "sub_update",
			Title:   "订阅更新失败",
			Message: fmt.Sprintf("订阅 [%s] 更新失败: %v", subName, err),
			Data: map[string]interface{}{
				"id":     id,
				"name":   subName,
				"status": "error",
			},
		})
		return
	}

	// 更新用量信息（如果开启了获取用量信息且成功获取到）
	if fetchUsageInfo && usageInfo != nil && airport != nil {
		if updateErr := airport.UpdateUsageInfo(usageInfo.Upload, usageInfo.Download, usageInfo.Total, usageInfo.Expire); updateErr != nil {
			utils.Warn("更新机场用量信息失败 ID: %d: %v", id, updateErr)
		} else {
			utils.Info("成功更新机场 [%s] 用量信息", subName)
		}
	}

	// 订阅更新成功后，应用自动标签规则
	go func() {
		updatedNodes, err := models.ListBySourceID(id)
		if err == nil && len(updatedNodes) > 0 {
			applyAutoTagRules(updatedNodes, "subscription_update")
		}
	}()
}
