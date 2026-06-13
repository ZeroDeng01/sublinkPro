package scheduler

import (
	"context"
	"fmt"
	"sublink/models"
	"sublink/node"
	"sublink/services/notifications"
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
	var requestHeaders models.AirportRequestHeaders
	var fetchUsageInfo bool
	var skipTLSVerify bool

	airport, err := models.GetAirportByID(id)
	if err != nil {
		utils.Warn("获取机场配置失败 ID: %d, 使用默认设置: %v", id, err)
	} else {
		downloadWithProxy = airport.DownloadWithProxy
		proxyLink = airport.ProxyLink
		userAgent = airport.UserAgent
		requestHeaders = airport.RequestHeaders
		fetchUsageInfo = airport.FetchUsageInfo
		skipTLSVerify = airport.SkipTLSVerify
	}

	// 创建 TaskManager 任务和报告器
	tm := getTaskManager()
	task, ctx, createErr := tm.CreateTask(models.TaskTypeSubUpdate, subName, trigger, 0)

	var reporter node.TaskReporter
	if createErr != nil {
		utils.Warn("创建订阅更新任务失败: %v，将使用降级模式", createErr)
		reporter = nil             // 使用 nil，将在 sub.go 中降级为 NoOpTaskReporter
		ctx = context.Background() // 降级情况下使用 Background context
	} else {
		reporter = NewTaskManagerReporter(tm, task.ID)

		// 添加全局 panic 保护
		defer func() {
			if r := recover(); r != nil {
				utils.Error("订阅任务执行发生 panic: %v, 任务ID: %s, 订阅名称: %s, URL: %s", r, task.ID, subName, url)
				reporter.ReportFail(fmt.Sprintf("任务异常崩溃: %v", r))
			}
		}()

		// 在调用前快速检查任务是否已被取消
		select {
		case <-ctx.Done():
			utils.Info("任务在执行前已被取消: %s", subName)
			return
		default:
		}
	}

	changedNodeIDs, usageInfo, err := node.LoadClashConfigFromURLWithReporter(ctx, id, url, subName, downloadWithProxy, proxyLink, userAgent, requestHeaders, reporter, fetchUsageInfo, skipTLSVerify)
	if err != nil {
		// 仅在失败时发送通知，成功通知由 node/sub.go 中的 scheduleClashToNodeLinks 发送
		// 这样可以避免重复通知，且成功通知包含更详细的节点统计信息
		if reporter != nil {
			reporter.ReportFail(err.Error())
		}
		notifications.Publish("subscription.sync_failed", notifications.Payload{
			Title:   "订阅更新失败",
			Message: fmt.Sprintf("订阅 [%s] 更新失败: %v", subName, err),
			Data: map[string]any{
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

	// 订阅更新成功后，如机场开启“更新后检测”，则立即按指定策略补做一次机场内节点检测。
	if airport != nil && airport.UpdateAfterDetect && airport.UpdateAfterDetectProfileID > 0 {
		profileID := airport.UpdateAfterDetectProfileID
		changedOnly := airport.UpdateAfterDetectChangedOnly
		go func(airportID int, airportName string, nodeCheckProfileID int, changedNodeIDs []int, changedOnly bool) {
			nodeIDs, shouldRun, listErr := resolveUpdateAfterDetectNodeIDs(airportID, changedNodeIDs, changedOnly)
			if listErr != nil {
				utils.Warn("获取机场节点失败，跳过更新后检测 - ID: %d, Error: %v", airportID, listErr)
				return
			}
			if !shouldRun {
				if changedOnly {
					utils.Info("机场 [%s] 开启了仅检测变化/新增节点，但本次更新没有变化/新增节点，跳过更新后检测", airportName)
				} else {
					utils.Warn("机场 [%s] 更新后检测已启用，但没有可检测节点", airportName)
				}
				return
			}

			if changedOnly {
				utils.Info("机场 [%s] 订阅更新完成，仅检测 %d 个变化/新增节点，策略 ID: %d", airportName, len(nodeIDs), nodeCheckProfileID)
			} else {
				utils.Info("机场 [%s] 订阅更新完成，立即执行节点检测策略 ID: %d", airportName, nodeCheckProfileID)
			}

			ExecuteNodeCheckWithProfile(nodeCheckProfileID, nodeIDs, models.TaskTriggerAirportUpdate)
		}(id, subName, profileID, changedNodeIDs, changedOnly)
	}
}

// resolveUpdateAfterDetectNodeIDs 解析更新后检测的目标节点；仅检测变化节点时，空变更代表跳过检测而不是回退到全机场。
func resolveUpdateAfterDetectNodeIDs(airportID int, changedNodeIDs []int, changedOnly bool) ([]int, bool, error) {
	if changedOnly {
		if len(changedNodeIDs) == 0 {
			return nil, false, nil
		}
		return changedNodeIDs, true, nil
	}

	updatedNodes, err := models.ListBySourceID(airportID)
	if err != nil {
		return nil, false, err
	}
	if len(updatedNodes) == 0 {
		return nil, false, nil
	}

	nodeIDs := make([]int, 0, len(updatedNodes))
	for _, n := range updatedNodes {
		nodeIDs = append(nodeIDs, n.ID)
	}
	return nodeIDs, true, nil
}
