package services

import (
	"fmt"
	"log"
	"sublink/models"
	"sublink/services/sse"
	"time"
)

// ApplyAutoTagRules 对节点应用自动标签规则
func ApplyAutoTagRules(nodes []models.Node, triggerType string) {
	if len(nodes) == 0 {
		return
	}

	// 获取指定触发类型的启用规则
	rules := models.ListByTriggerType(triggerType)
	if len(rules) == 0 {
		return
	}

	log.Printf("开始应用自动标签规则: 触发类型=%s, 节点数=%d, 规则数=%d", triggerType, len(nodes), len(rules))

	taggedCount := 0
	// 规则名称
	ruleNames := make([]string, 0)
	for _, rule := range rules {
		ruleNames = append(ruleNames, rule.Name)
		// 解析条件
		conditions, err := models.ParseConditions(rule.Conditions)
		if err != nil {
			log.Printf("规则 %s 条件解析失败: %v", rule.Name, err)
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
			log.Printf("规则 [%s] 匹配 %d 个节点, 打标签: %s", rule.Name, len(matchedNodeIDs), rule.TagName)
			if err := models.BatchAddTagToNodes(matchedNodeIDs, rule.TagName); err != nil {
				log.Printf("批量打标签失败: %v", err)
			} else {
				taggedCount += len(matchedNodeIDs)
			}
		}

		// 批量移除不满足条件的标签
		if len(unmatchedNodeIDs) > 0 {
			log.Printf("规则 [%s] 移除 %d 个不满足条件的节点标签: %s", rule.Name, len(unmatchedNodeIDs), rule.TagName)
			if err := models.BatchRemoveTagFromNodes(unmatchedNodeIDs, rule.TagName); err != nil {
				log.Printf("批量移除标签失败: %v", err)
			}
		}
	}

	if taggedCount > 0 {
		log.Printf("自动标签规则应用完成: 共标记 %d 个节点", taggedCount)
		// 广播事件
		sse.GetSSEBroker().BroadcastEvent("task_update", sse.NotificationPayload{
			Event:   "auto_tag",
			Title:   "自动标签完成",
			Message: fmt.Sprintf("自动标签规则【%s】应用完成，执行规则【%s】: 共标记 %d 个节点", triggerType, ruleNames, taggedCount),
			Data: map[string]interface{}{
				"status":      "success",
				"error":       fmt.Sprintf("自动标签规则【%s】应用完成: 共标记 %d 个节点", triggerType, taggedCount),
				"triggerType": triggerType,
				"taggedCount": taggedCount,
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

	// 生成唯一任务ID
	taskID := fmt.Sprintf("tag_rule_%d_%d", ruleID, time.Now().UnixNano())
	startTime := time.Now()
	startTimeMs := startTime.UnixMilli()

	// 广播任务开始事件
	sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
		TaskID:    taskID,
		TaskType:  "tag_rule",
		TaskName:  rule.Name,
		Status:    "started",
		Current:   0,
		Total:     totalNodes,
		Message:   fmt.Sprintf("开始执行标签规则: %s", rule.Name),
		StartTime: startTimeMs,
	})

	// 解析条件
	conditions, err := models.ParseConditions(rule.Conditions)
	if err != nil {
		// 广播错误事件
		sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
			TaskID:    taskID,
			TaskType:  "tag_rule",
			TaskName:  rule.Name,
			Status:    "error",
			Current:   0,
			Total:     totalNodes,
			Message:   fmt.Sprintf("规则条件解析失败: %v", err),
			StartTime: startTimeMs,
		})
		return err
	}

	// 评估并打标签 (使用标签名称)
	matchedCount := 0
	removedCount := 0
	// 计算进度广播间隔（避免大量节点时消息阻塞）
	broadcastInterval := 1
	if totalNodes > 500 {
		broadcastInterval = 100
	} else if totalNodes > 50 {
		broadcastInterval = 50
	}
	for i, n := range nodes {
		matched := conditions.EvaluateNode(n)
		resultStatus := "skipped"

		if matched {
			if err := n.AddTagByName(rule.TagName); err == nil {
				matchedCount++
				resultStatus = "tagged"
			} else {
				resultStatus = "error"
			}
		} else if n.HasTagName(rule.TagName) {
			// 不满足条件但有此标签，移除它
			if err := n.RemoveTagByName(rule.TagName); err == nil {
				removedCount++
				resultStatus = "untagged"
			} else {
				resultStatus = "error"
			}
		}

		// 广播进度更新（降低频率防止消息阻塞）
		currentProgress := i + 1
		shouldBroadcast := currentProgress%broadcastInterval == 0 || currentProgress == totalNodes
		if shouldBroadcast {
			sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
				TaskID:      taskID,
				TaskType:    "tag_rule",
				TaskName:    rule.Name,
				Status:      "progress",
				Current:     currentProgress,
				Total:       totalNodes,
				CurrentItem: n.Name,
				Result: map[string]interface{}{
					"status":  resultStatus,
					"matched": matched,
				},
				StartTime: startTimeMs,
			})
		}
	}

	// 广播任务完成事件
	sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
		TaskID:   taskID,
		TaskType: "tag_rule",
		TaskName: rule.Name,
		Status:   "completed",
		Current:  totalNodes,
		Total:    totalNodes,
		Message:  fmt.Sprintf("规则执行完成: 匹配 %d 个节点, 移除 %d 个节点标签", matchedCount, removedCount),
		Result: map[string]interface{}{
			"matchedCount": matchedCount,
			"removedCount": removedCount,
			"totalCount":   totalNodes,
		},
		StartTime: startTimeMs,
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

	log.Printf("手动触发规则 [%s] 完成: 匹配 %d 个节点, 移除 %d 个节点标签", rule.Name, matchedCount, removedCount)
	return nil
}
