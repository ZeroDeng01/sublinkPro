package scheduler

import (
	"context"
	"fmt"
	"sublink/models"
	"sublink/node"
	"sublink/services/notifications"
	"sublink/utils"
	"time"
)

// ExecuteSubscriptionTask 执行订阅任务的具体业务逻辑
func ExecuteSubscriptionTask(id int, url string, subName string) {
	ExecuteSubscriptionTaskWithTrigger(id, url, subName, models.TaskTriggerScheduled)
}

// ExecuteSubscriptionTaskWithTrigger 执行订阅任务（带触发类型）
func ExecuteSubscriptionTaskWithTrigger(id int, url string, subName string, trigger models.TaskTrigger) {
	if err := executeSubscriptionTask(id, url, subName, trigger, true); err != nil {
		utils.Warn("Subscription task failed: %v", err)
	}
}

// executeSubscriptionTask 提供可等待、可抑制检测联动的前置更新边界。
func executeSubscriptionTask(id int, url string, subName string, trigger models.TaskTrigger, followups bool) (runErr error) {
	utils.Info("Running automatic subscription task - ID: %d, Name: %s, URL: %s, Trigger: %s", id, subName, url, trigger)

	// 获取最新的机场配置，以便使用最新的代理设置
	var downloadWithProxy bool
	var proxyLink string
	var userAgent string
	var requestHeaders models.AirportRequestHeaders
	var fetchUsageInfo bool
	var skipTLSVerify bool

	airport, err := models.GetAirportByID(id)
	if err != nil {
		utils.Warn("Failed to load airport config ID: %d; using defaults: %v", id, err)
	} else {
		downloadWithProxy = airport.DownloadWithProxy
		proxyLink = airport.ProxyLink
		userAgent = airport.EffectiveUserAgent()
		requestHeaders = airport.EffectiveRequestHeaders()
		fetchUsageInfo = airport.FetchUsageInfo
		skipTLSVerify = airport.SkipTLSVerify
	}

	// 创建 TaskManager 任务和报告器
	tm := getTaskManager()
	task, ctx, createErr := tm.CreateTask(models.TaskTypeSubUpdate, subName, trigger, 0)

	var reporter node.TaskReporter
	if createErr != nil {
		utils.Warn("Failed to create subscription update task: %v; falling back to degraded mode", createErr)
		reporter = nil             // 使用 nil，将在 sub.go 中降级为 NoOpTaskReporter
		ctx = context.Background() // 降级情况下使用 Background context
	} else {
		reporter = NewTaskManagerReporter(tm, task.ID)

		// 添加全局 panic 保护
		defer func() {
			if r := recover(); r != nil {
				utils.Error("Subscription task panicked: %v, task ID: %s, subscription: %s, URL: %s", r, task.ID, subName, url)
				runErr = fmt.Errorf("task crashed: %v", r)
				reporter.ReportFail(runErr.Error())
			}
		}()

		// 在调用前快速检查任务是否已被取消
		select {
		case <-ctx.Done():
			utils.Info("Task was cancelled before execution: %s", subName)
			return ctx.Err()
		default:
		}
	}

	// 兜底超时：上游 HTTP/代理层卡死时，任务不会长期停留在 running。
	// 成功路径下 ReportComplete 也会主动取消 context，两者不冲突。
	const subscriptionUpdateTimeout = 10 * time.Minute
	ctx, cancel := context.WithTimeout(ctx, subscriptionUpdateTimeout)
	defer cancel()

	changedNodeIDs, usageInfo, err := node.LoadClashConfigFromURLWithReporter(ctx, id, url, subName, downloadWithProxy, proxyLink, userAgent, requestHeaders, reporter, fetchUsageInfo, skipTLSVerify)

	// 用量与节点解析彼此独立：即使节点解析失败，也要把已获取到的用量落库并记录失败点，
	// 便于历史图表保留数据和红色失败标记。
	persistSubscriptionUsage(airport, fetchUsageInfo, usageInfo, err)

	if err != nil {
		// 仅在失败时发送通知，成功通知由 node/sub.go 中的 scheduleClashToNodeLinks 发送
		// 这样可以避免重复通知，且成功通知包含更详细的节点统计信息
		if reporter != nil {
			reporter.ReportFail(err.Error())
		}
		notifications.Publish("subscription.sync_failed", notifications.Payload{
			Title:   "Subscription update failed",
			Message: fmt.Sprintf("Subscription [%s] update failed: %v", subName, err),
			Data: map[string]any{
				"id":     id,
				"name":   subName,
				"status": "error",
			},
		})
		return err
	}

	// 加载器已返回真实失败/取消错误；成功 ReportComplete 也会主动取消任务 context，
	// 此处不能再据 ctx.Err() 否定成功，否则用量、标签及后续检测都会被跳过。

	// 标签属于节点更新边界，完成后才解析按标签筛选的自动检测范围。
	applySubscriptionUpdateTags(id)

	if !followups {
		return nil
	}
	go RunNodeCheckJobsAfterAirportUpdate(id)

	// 订阅更新成功后，如机场开启“更新后检测”，则立即按指定策略补做一次机场内节点检测。
	if airport != nil && airport.UpdateAfterDetect && airport.UpdateAfterDetectProfileID > 0 {
		profileID := airport.UpdateAfterDetectProfileID
		changedOnly := airport.UpdateAfterDetectChangedOnly
		go func(airportID int, airportName string, nodeCheckProfileID int, changedNodeIDs []int, changedOnly bool) {
			nodeIDs, shouldRun, listErr := resolveUpdateAfterDetectNodeIDs(airportID, changedNodeIDs, changedOnly)
			if listErr != nil {
				utils.Warn("Failed to load airport nodes; skipping post-update check - ID: %d, Error: %v", airportID, listErr)
				return
			}
			if !shouldRun {
				if changedOnly {
					utils.Info("Airport [%s] only checks changed/new nodes, but this update produced none; skipping post-update check", airportName)
				} else {
					utils.Warn("Airport [%s] has post-update checks enabled but no nodes to check", airportName)
				}
				return
			}

			if changedOnly {
				utils.Info("Airport [%s] subscription updated; checking only %d changed/new nodes, profile ID: %d", airportName, len(nodeIDs), nodeCheckProfileID)
			} else {
				utils.Info("Airport [%s] subscription updated; running node check profile ID: %d", airportName, nodeCheckProfileID)
			}

			ExecuteNodeCheckWithProfile(nodeCheckProfileID, nodeIDs, models.TaskTriggerAirportUpdate)
		}(id, subName, profileID, changedNodeIDs, changedOnly)
	}
	return nil
}

// persistSubscriptionUsage 将本次拉取到的用量落库，并记录刷新状态。
// usageInfo 非空时保存数值；节点解析失败时标记为失败点但保留数值。
func persistSubscriptionUsage(airport *models.Airport, fetchUsageInfo bool, usageInfo *node.UsageInfo, parseErr error) {
	if !fetchUsageInfo || airport == nil {
		return
	}

	if usageInfo == nil {
		if parseErr != nil {
			airport.RecordUsageFailure(parseErr.Error())
		}
		return
	}

	reason := ""
	if parseErr != nil {
		reason = parseErr.Error()
	}
	if updateErr := airport.UpdateUsageInfoWithStatus(
		usageInfo.Upload, usageInfo.Download, usageInfo.Total, usageInfo.Expire, parseErr != nil, reason,
	); updateErr != nil {
		utils.Warn("Failed to update airport usage ID: %d: %v", airport.ID, updateErr)
		return
	}
	utils.Info("Updated usage info for airport [%s]", airport.Name)
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

// applySubscriptionUpdateTags 保证按标签检测不会漏掉本次更新刚分配标签的节点。
func applySubscriptionUpdateTags(id int) {
	nodes, err := models.ListBySourceID(id)
	if err != nil {
		utils.Warn("Failed to load tagged nodes after update: %v", err)
		return
	}
	if len(nodes) > 0 {
		applyAutoTagRules(nodes, "subscription_update")
	}
}
