package services

import (
	"fmt"
	"sublink/models"
	"sublink/services/sse"
	"sublink/utils"
)

// ApplyAutoTagRules 对节点应用自动标签规则
// 注意：此函数通过 TaskManager 创建任务记录，以便在任务列表中显示
func ApplyAutoTagRules(nodes []models.Node, triggerType string) {
	if len(nodes) == 0 {
		return
	}

	// 获取指定触发类型的启用规则
	rules := models.ListByTriggerType(triggerType)
	if len(rules) == 0 {
		return
	}

	utils.Info("开始应用自动标签规则: 触发类型=%s, 节点数=%d, 规则数=%d", triggerType, len(nodes), len(rules))

	// 使用 TaskManager 创建任务（自动触发）
	tm := GetTaskManager()
	taskName := fmt.Sprintf("自动标签规则 (%s)", triggerType)
	task, _, createErr := tm.CreateTask(models.TaskTypeTagRule, taskName, models.TaskTriggerScheduled, len(nodes))
	var taskID string
	if createErr != nil {
		utils.Error("创建自动标签任务失败: %v，继续执行但不追踪任务", createErr)
	} else {
		taskID = task.ID
	}

	taggedCount := 0
	removedCount := 0
	processedCount := 0
	// 规则名称
	ruleNames := make([]string, 0)
	for _, rule := range rules {
		ruleNames = append(ruleNames, rule.Name)
		// 解析条件
		conditions, err := models.ParseConditions(rule.Conditions)
		if err != nil {
			utils.Error("规则 %s 条件解析失败: %v", rule.Name, err)
			continue
		}

		// 评估每个节点
		matchedNodeIDs := make([]int, 0)
		unmatchedNodeIDs := make([]int, 0)
		for _, node := range nodes {
			if conditions.EvaluateNode(node) {
				matchedNodeIDs = append(matchedNodeIDs, node.ID)
			} else if node.HasTagName(rule.TagName) {
				// 节点不满足条件但有此标签，需要移除
				unmatchedNodeIDs = append(unmatchedNodeIDs, node.ID)
			}
		}

		// 批量打标签 (使用标签名称)
		if len(matchedNodeIDs) > 0 {
			utils.Info("规则 [%s] 匹配 %d 个节点, 打标签: %s", rule.Name, len(matchedNodeIDs), rule.TagName)
			if err := models.BatchAddTagToNodes(matchedNodeIDs, rule.TagName); err != nil {
				utils.Error("批量打标签失败: %v", err)
			} else {
				taggedCount += len(matchedNodeIDs)
			}
		}

		// 批量移除不满足条件的标签
		if len(unmatchedNodeIDs) > 0 {
			utils.Info("规则 [%s] 移除 %d 个不满足条件的节点标签: %s", rule.Name, len(unmatchedNodeIDs), rule.TagName)
			if err := models.BatchRemoveTagFromNodes(unmatchedNodeIDs, rule.TagName); err != nil {
				utils.Error("批量移除标签失败: %v", err)
			} else {
				removedCount += len(unmatchedNodeIDs)
			}
		}

		processedCount += len(matchedNodeIDs) + len(unmatchedNodeIDs)

		// 更新进度（仅SSE广播，不写数据库）
		if taskID != "" {
			tm.UpdateProgress(taskID, processedCount, rule.Name, map[string]interface{}{
				"matched": len(matchedNodeIDs),
				"removed": len(unmatchedNodeIDs),
			})
		}
	}

	// 完成任务
	if taskID != "" {
		message := fmt.Sprintf("自动标签完成: 标记 %d 个节点, 移除 %d 个标签", taggedCount, removedCount)
		tm.CompleteTask(taskID, message, map[string]interface{}{
			"taggedCount":  taggedCount,
			"removedCount": removedCount,
			"totalNodes":   len(nodes),
			"triggerType":  triggerType,
		})
	}

	if taggedCount > 0 || removedCount > 0 {
		utils.Info("自动标签规则应用完成: 共标记 %d 个节点, 移除 %d 个标签", taggedCount, removedCount)
		// 广播事件
		sse.GetSSEBroker().BroadcastEvent("task_update", sse.NotificationPayload{
			Event:   "auto_tag",
			Title:   "自动标签完成",
			Message: fmt.Sprintf("自动标签规则【%s】应用完成，执行规则【%s】: 共标记 %d 个节点", triggerType, ruleNames, taggedCount),
			Data: map[string]interface{}{
				"status":       "success",
				"error":        fmt.Sprintf("自动标签规则【%s】应用完成: 共标记 %d 个节点", triggerType, taggedCount),
				"triggerType":  triggerType,
				"taggedCount":  taggedCount,
				"removedCount": removedCount,
			},
		})
	}
}

// TriggerTagRule 手动触发指定规则
func TriggerTagRule(ruleID int) error {
	var rule models.TagRule
	if err := rule.GetByID(ruleID); err != nil {
		return err
	}

	// 获取所有节点
	var node models.Node
	nodes, err := node.List()
	if err != nil {
		return err
	}

	totalNodes := len(nodes)
	if totalNodes == 0 {
		return nil
	}

	// 使用 TaskManager 创建任务
	tm := GetTaskManager()
	task, _, createErr := tm.CreateTask(models.TaskTypeTagRule, rule.Name, models.TaskTriggerManual, totalNodes)
	if createErr != nil {
		utils.Error("创建标签规则任务失败: %v", createErr)
		return createErr
	}
	taskID := task.ID

	// 解析条件
	conditions, err := models.ParseConditions(rule.Conditions)
	if err != nil {
		// 使用 TaskManager 报告失败
		tm.FailTask(taskID, fmt.Sprintf("规则条件解析失败: %v", err))
		return err
	}

	// 评估节点并收集需要操作的节点ID（稍后批量写入）
	matchedNodeIDs := make([]int, 0)
	unmatchedNodeIDs := make([]int, 0)

	for i, n := range nodes {
		matched := conditions.EvaluateNode(n)
		resultStatus := "skipped"

		if matched && !n.HasTagName(rule.TagName) {
			// 匹配条件但没有此标签，需要添加
			matchedNodeIDs = append(matchedNodeIDs, n.ID)
			resultStatus = "tagged"
		} else if !matched && n.HasTagName(rule.TagName) {
			// 不满足条件但有此标签，需要移除
			unmatchedNodeIDs = append(unmatchedNodeIDs, n.ID)
			resultStatus = "untagged"
		}

		// 更新进度（TaskManager 内置节流策略）- 基于内存计数，保持实时性
		currentProgress := i + 1
		tm.UpdateProgress(taskID, currentProgress, n.Name, map[string]interface{}{
			"status":  resultStatus,
			"matched": matched,
		})
	}

	// 批量写入数据库（一次性操作，减少数据库I/O）
	matchedCount := 0
	removedCount := 0

	if len(matchedNodeIDs) > 0 {
		if err := models.BatchAddTagToNodes(matchedNodeIDs, rule.TagName); err != nil {
			utils.Error("❌批量添加标签失败：%v", err)
		} else {
			matchedCount = len(matchedNodeIDs)
			utils.Info("✅批量添加标签到 %d 个节点", matchedCount)
		}
	}

	if len(unmatchedNodeIDs) > 0 {
		if err := models.BatchRemoveTagFromNodes(unmatchedNodeIDs, rule.TagName); err != nil {
			utils.Error("❌批量移除标签失败：%v", err)
		} else {
			removedCount = len(unmatchedNodeIDs)
			utils.Info("✅批量移除 %d 个节点的标签", removedCount)
		}
	}

	// 使用 TaskManager 完成任务
	tm.CompleteTask(taskID, fmt.Sprintf("规则执行完成: 匹配 %d 个节点, 移除 %d 个节点标签", matchedCount, removedCount), map[string]interface{}{
		"matchedCount": matchedCount,
		"removedCount": removedCount,
		"totalCount":   totalNodes,
	})

	// 广播通知消息，让用户在通知中心看到完成通知
	sse.GetSSEBroker().BroadcastEvent("task_update", sse.NotificationPayload{
		Event:   "tag_rule",
		Title:   "标签规则执行完成",
		Message: fmt.Sprintf("规则【%s】执行完成: 匹配 %d 个节点, 移除 %d 个节点标签", rule.Name, matchedCount, removedCount),
		Data: map[string]interface{}{
			"status":       "success",
			"matchedCount": matchedCount,
			"removedCount": removedCount,
			"totalCount":   totalNodes,
		},
	})

	utils.Info("手动触发规则 [%s] 完成: 匹配 %d 个节点, 移除 %d 个节点标签", rule.Name, matchedCount, removedCount)
	return nil
}
