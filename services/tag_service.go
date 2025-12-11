package services

import (
	"log"
	"sublink/models"
	"sublink/services/sse"
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
	for _, rule := range rules {
		// 解析条件
		conditions, err := models.ParseConditions(rule.Conditions)
		if err != nil {
			log.Printf("规则 %s 条件解析失败: %v", rule.Name, err)
			continue
		}

		// 评估每个节点
		matchedNodeIDs := make([]int, 0)
		for _, node := range nodes {
			if conditions.EvaluateNode(node) {
				matchedNodeIDs = append(matchedNodeIDs, node.ID)
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
	}

	if taggedCount > 0 {
		log.Printf("自动标签规则应用完成: 共标记 %d 个节点", taggedCount)
		// 广播事件
		sse.GetSSEBroker().BroadcastEvent("task_update", sse.NotificationPayload{
			Event:   "auto_tag",
			Title:   "自动标签完成",
			Message: "自动标签规则执行完成",
			Data: map[string]interface{}{
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

	// 解析条件
	conditions, err := models.ParseConditions(rule.Conditions)
	if err != nil {
		return err
	}

	// 评估并打标签 (使用标签名称)
	matchedCount := 0
	for _, n := range nodes {
		if conditions.EvaluateNode(n) {
			if err := n.AddTagByName(rule.TagName); err == nil {
				matchedCount++
			}
		}
	}

	log.Printf("手动触发规则 [%s] 完成: 匹配 %d 个节点", rule.Name, matchedCount)
	return nil
}
