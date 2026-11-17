package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sublink/cache"
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// GetChainRules 获取订阅的链式代理规则列表
func GetChainRules(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	rules := models.GetChainRulesBySubscriptionID(subID)

	// 调试日志：记录返回的规则
	utils.Debug("[ChainRule] Loaded %d rules for subscription %d", subID, len(rules))
	for _, r := range rules {
		utils.Debug("[ChainRule] Rule ID=%d, Name=%s, ChainConfig=%s", r.ID, r.Name, r.ChainConfig)
	}

	c.JSON(http.StatusOK, gin.H{"data": rules})
}

// CreateChainRule 创建链式代理规则
func CreateChainRule(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	var rule models.SubscriptionChainRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	rule.SubscriptionID = subID

	// 设置默认排序值（最后一个）
	existingRules := models.GetChainRulesBySubscriptionID(subID)
	rule.Sort = len(existingRules)

	if err := rule.Add(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// UpdateChainRule 更新链式代理规则
func UpdateChainRule(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	ruleIDStr := c.Param("ruleId")
	ruleID, err := strconv.Atoi(ruleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	// 获取现有规则
	var existingRule models.SubscriptionChainRule
	if err := existingRule.GetByID(ruleID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	// 验证规则属于该订阅
	if existingRule.SubscriptionID != subID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to operate this rule"})
		return
	}

	// 绑定更新数据
	var updateData models.SubscriptionChainRule
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// 更新字段
	existingRule.Name = updateData.Name
	existingRule.Enabled = updateData.Enabled
	existingRule.ChainConfig = updateData.ChainConfig
	existingRule.TargetConfig = updateData.TargetConfig

	// 调试日志：记录更新的数据
	utils.Debug("[ChainRule] Updating rule ID=%d, name=%s", existingRule.ID, existingRule.Name)
	utils.Debug("[ChainRule] ChainConfig: %s", existingRule.ChainConfig)
	utils.Debug("[ChainRule] TargetConfig: %s", existingRule.TargetConfig)

	if err := existingRule.Update(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rule: " + err.Error()})
		return
	}

	utils.Debug("[ChainRule] Rule updated; returning ID=%d", existingRule.ID)
	c.JSON(http.StatusOK, gin.H{"data": existingRule})
}

// DeleteChainRule 删除链式代理规则
func DeleteChainRule(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	ruleIDStr := c.Param("ruleId")
	ruleID, err := strconv.Atoi(ruleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	// 获取现有规则
	var existingRule models.SubscriptionChainRule
	if err := existingRule.GetByID(ruleID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	// 验证规则属于该订阅
	if existingRule.SubscriptionID != subID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to operate this rule"})
		return
	}

	if err := existingRule.Delete(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

// SortChainRules 批量排序链式代理规则
func SortChainRules(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	var req struct {
		RuleIDs []int `json:"ruleIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// 验证所有规则都属于该订阅
	for _, ruleID := range req.RuleIDs {
		var rule models.SubscriptionChainRule
		if err := rule.GetByID(ruleID); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found: " + strconv.Itoa(ruleID)})
			return
		}
		if rule.SubscriptionID != subID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to operate rule: " + strconv.Itoa(ruleID)})
			return
		}
	}

	if err := models.UpdateChainRulesSort(req.RuleIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sorting failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sort order saved"})
}

// GetChainOptions 获取链式代理可用选项
// 返回：模板代理组列表、条件字段列表、节点列表
func GetChainOptions(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	// 获取订阅信息
	var sub models.Subcription
	sub.ID = subID
	if err := sub.Find(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// 获取订阅关联的节点
	if err := sub.GetSub("none"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription nodes: " + err.Error()})
		return
	}

	// 构建节点简要信息列表
	nodeOptions := make([]map[string]any, 0, len(sub.Nodes))
	for _, node := range sub.Nodes {
		effectiveName := node.EffectiveName()
		nodeOptions = append(nodeOptions, map[string]any{
			"id":            node.ID,
			"name":          effectiveName,
			"remarkName":    node.Name,
			"linkName":      node.LinkName,
			"nameMode":      models.NormalizeNodeNameMode(node.NameMode),
			"effectiveName": effectiveName,
			"linkCountry":   node.LinkCountry,
			"protocol":      node.Protocol,
			"group":         node.Group,
		})
	}

	// 条件字段列表
	fieldMetas := models.GetNodeConditionFields()
	conditionFields := make([]models.NodeConditionFieldMeta, 0, len(fieldMetas))
	for _, field := range fieldMetas {
		if field.Value == "dialer_proxy_name" || field.Value == "link" {
			continue
		}
		conditionFields = append(conditionFields, field)
	}

	// 条件操作符列表
	operators := []map[string]string{
		{"value": "equals", "label": "Equals"},
		{"value": "not_equals", "label": "Not equals"},
		{"value": "contains", "label": "Contains"},
		{"value": "not_contains", "label": "Does not contain"},
		{"value": "regex", "label": "Matches regex"},
		{"value": "greater_than", "label": "Greater than"},
		{"value": "less_than", "label": "Less than"},
		{"value": "greater_or_equal", "label": "Greater than or equal"},
		{"value": "less_or_equal", "label": "Less than or equal"},
	}

	// 代理组类型
	groupTypes := []map[string]string{
		{"value": "select", "label": "Manual select (select)"},
		{"value": "url-test", "label": "Auto speed test (url-test)"},
		{"value": "fallback", "label": "Fallback (fallback)"},
		{"value": "load-balance", "label": "Load balance (load-balance)"},
	}

	// 从订阅配置中读取模板代理组列表
	templateGroups := parseTemplateProxyGroups(sub.Config)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"nodes":           nodeOptions,
			"conditionFields": conditionFields,
			"operators":       operators,
			"groupTypes":      groupTypes,
			"templateGroups":  templateGroups,
		},
	})
}

// ToggleChainRule 切换链式代理规则启用状态
func ToggleChainRule(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	ruleIDStr := c.Param("ruleId")
	ruleID, err := strconv.Atoi(ruleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	// 获取现有规则
	var existingRule models.SubscriptionChainRule
	if err := existingRule.GetByID(ruleID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	// 验证规则属于该订阅
	if existingRule.SubscriptionID != subID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to operate this rule"})
		return
	}

	// 切换状态
	existingRule.Enabled = !existingRule.Enabled

	if err := existingRule.Update(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": existingRule})
}

// parseTemplateProxyGroups 从订阅配置中解析模板代理组列表
// configStr: 订阅的 Config 字段（JSON 格式）
// 返回: 代理组名称列表
func parseTemplateProxyGroups(configStr string) []string {
	if configStr == "" {
		return []string{}
	}

	// 解析订阅配置 JSON
	var config struct {
		Clash string `json:"clash"`
		Surge string `json:"surge"`
	}
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return []string{}
	}

	// 获取 Clash 模板路径
	clashTemplate := config.Clash
	if clashTemplate == "" {
		return []string{}
	}

	// 读取模板内容
	var templateContent string
	if strings.Contains(clashTemplate, "://") {
		// 远程模板，通过 HTTP 获取
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, clashTemplate, nil)
		if err != nil {
			return []string{}
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return []string{}
		}
		defer func() { _ = resp.Body.Close() }()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return []string{}
		}
		templateContent = string(data)
	} else {
		// 本地模板，优先从缓存读取
		filename := filepath.Base(clashTemplate)
		if cached, ok := cache.GetTemplateContent(filename); ok {
			templateContent = cached
		} else {
			data, err := os.ReadFile(clashTemplate)
			if err != nil {
				return []string{}
			}
			templateContent = string(data)
		}
	}

	// 解析 YAML 获取代理组列表
	var clashConfig map[string]any
	if err := yaml.Unmarshal([]byte(templateContent), &clashConfig); err != nil {
		return []string{}
	}

	// 提取 proxy-groups 中的 name 字段
	proxyGroups, ok := clashConfig["proxy-groups"].([]any)
	if !ok {
		return []string{}
	}

	var groupNames []string
	for _, pg := range proxyGroups {
		if group, ok := pg.(map[string]any); ok {
			if name, ok := group["name"].(string); ok && name != "" {
				groupNames = append(groupNames, name)
			}
		}
	}

	return groupNames
}

// ChainLinkPreviewNode 链路预览中的节点信息
type ChainLinkPreviewNode struct {
	Name        string  `json:"name"`
	Protocol    string  `json:"protocol"`
	LinkCountry string  `json:"linkCountry"`
	DelayTime   int     `json:"delayTime"`
	Speed       float64 `json:"speed"`
	Group       string  `json:"group"`
}

// ChainLinkPreviewItem 链路预览中的单项信息
type ChainLinkPreviewItem struct {
	Type        string                 `json:"type"`        // template_group, custom_group, dynamic_node, specified_node
	Name        string                 `json:"name"`        // 代理名称
	IsGroup     bool                   `json:"isGroup"`     // 是否为代理组
	GroupType   string                 `json:"groupType"`   // select, url-test (仅组类型)
	DialerProxy string                 `json:"dialerProxy"` // 上级 dialer-proxy
	Nodes       []ChainLinkPreviewNode `json:"nodes"`       // 匹配的节点列表（仅动态/自定义组）
}

// ChainPreviewResult 单条规则的预览数据
type ChainPreviewResult struct {
	RuleID         int                    `json:"ruleId"`
	RuleName       string                 `json:"ruleName"`
	Enabled        bool                   `json:"enabled"`
	Sort           int                    `json:"sort"`
	Links          []ChainLinkPreviewItem `json:"links"`          // 链路节点列表
	TargetType     string                 `json:"targetType"`     // all, conditions, specified_node
	TargetInfo     string                 `json:"targetInfo"`     // 目标描述
	TargetNodes    []ChainLinkPreviewNode `json:"targetNodes"`    // 匹配的目标节点列表
	EffectiveNodes int                    `json:"effectiveNodes"` // 实际生效的节点数
	CoveredNodes   int                    `json:"coveredNodes"`   // 被前面规则覆盖的节点数
	FullyCovered   bool                   `json:"fullyCovered"`   // 是否完全被覆盖（即无生效节点）
}

// SubscriptionChainPreviewResult 订阅链式代理整体预览结果
type SubscriptionChainPreviewResult struct {
	SubscriptionName string               `json:"subscriptionName"`
	TotalNodes       int                  `json:"totalNodes"`   // 订阅总节点数
	Rules            []ChainPreviewResult `json:"rules"`        // 所有规则预览
	MatchSummary     []NodeMatchSummary   `json:"matchSummary"` // 节点匹配摘要
}

// NodeMatchSummary 节点匹配摘要
type NodeMatchSummary struct {
	NodeID        int    `json:"nodeId"`
	NodeName      string `json:"nodeName"`
	LinkCountry   string `json:"linkCountry"`
	MatchedRule   string `json:"matchedRule"` // 匹配的规则名称（第一个匹配的）
	MatchedRuleID int    `json:"matchedRuleId"`
	EntryProxy    string `json:"entryProxy"` // 入口代理名称
	Unmatched     bool   `json:"unmatched"`  // 是否未匹配任何规则
}

func buildChainLinkPreviewNode(node models.Node) ChainLinkPreviewNode {
	return ChainLinkPreviewNode{
		Name:        node.EffectiveName(),
		Protocol:    node.Protocol,
		LinkCountry: node.LinkCountry,
		DelayTime:   node.DelayTime,
		Speed:       node.Speed,
		Group:       node.Group,
	}
}

// PreviewChainLinks 预览订阅的整体链式代理配置
func PreviewChainLinks(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	// 获取订阅及其节点
	var sub models.Subcription
	sub.ID = subID
	if err := sub.Find(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	if err := sub.GetSub("none"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription nodes: " + err.Error()})
		return
	}

	// 构建节点名称映射
	nodeNameMap := make(map[int]string)
	nodeInfoMap := make(map[int]models.Node)
	for _, node := range sub.Nodes {
		nodeNameMap[node.ID] = node.EffectiveName()
		nodeInfoMap[node.ID] = node
	}

	// 获取所有规则（按排序）
	rules := models.GetChainRulesBySubscriptionID(subID)

	// 筛选启用的规则
	enabledRules := make([]models.SubscriptionChainRule, 0)
	for _, rule := range rules {
		if rule.Enabled {
			enabledRules = append(enabledRules, rule)
		}
	}

	// 记录已被匹配的节点 ID（用于计算规则覆盖）
	matchedNodeIDs := make(map[int]bool)

	// 构建规则预览数据（考虑覆盖策略）
	var rulesPreview []ChainPreviewResult
	for _, rule := range rules {
		ruleData := buildRulePreviewData(rule, sub.Nodes, nodeNameMap, nodeInfoMap)

		// 只有启用的规则才计算生效节点
		if rule.Enabled {
			effectiveCount := 0
			coveredCount := 0

			for _, targetNode := range ruleData.TargetNodes {
				// 从 nodeInfoMap 查找节点 ID
				var nodeID int
				for id, node := range nodeInfoMap {
					if node.EffectiveName() == targetNode.Name {
						nodeID = id
						break
					}
				}

				if matchedNodeIDs[nodeID] {
					// 该节点已被前面规则覆盖
					coveredCount++
				} else {
					// 该节点实际生效
					effectiveCount++
					matchedNodeIDs[nodeID] = true
				}
			}

			ruleData.EffectiveNodes = effectiveCount
			ruleData.CoveredNodes = coveredCount
			ruleData.FullyCovered = effectiveCount == 0 && len(ruleData.TargetNodes) > 0
		}

		rulesPreview = append(rulesPreview, ruleData)
	}

	// 构建节点匹配摘要
	matchSummary := buildNodeMatchSummary(sub.Nodes, enabledRules, nodeNameMap)

	result := SubscriptionChainPreviewResult{
		SubscriptionName: sub.Name,
		TotalNodes:       len(sub.Nodes),
		Rules:            rulesPreview,
		MatchSummary:     matchSummary,
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// buildRulePreviewData 构建单条规则的预览数据
func buildRulePreviewData(rule models.SubscriptionChainRule, nodes []models.Node, nodeNameMap map[int]string, nodeInfoMap map[int]models.Node) ChainPreviewResult {
	data := ChainPreviewResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Enabled:  rule.Enabled,
		Sort:     rule.Sort,
	}

	// 解析链路配置
	chainItems, err := rule.ParseChainConfig()
	if err != nil {
		return data
	}

	var prevProxyName string
	for i, item := range chainItems {
		previewItem := ChainLinkPreviewItem{
			Type:    item.Type,
			IsGroup: item.Type == "template_group" || item.Type == "custom_group",
		}

		if i > 0 && prevProxyName != "" {
			previewItem.DialerProxy = prevProxyName
		}

		switch item.Type {
		case "template_group":
			previewItem.Name = item.GroupName
			previewItem.GroupType = "select"

		case "custom_group":
			previewItem.Name = item.GroupName
			previewItem.GroupType = item.GroupType
			if previewItem.GroupType == "" {
				previewItem.GroupType = "select"
			}
			if item.NodeConditions != nil {
				for _, node := range nodes {
					if item.NodeConditions.EvaluateNode(node) {
						previewItem.Nodes = append(previewItem.Nodes, buildChainLinkPreviewNode(node))
					}
				}
			}

		case "dynamic_node":
			var matchedNodes []ChainLinkPreviewNode
			if item.NodeConditions != nil {
				for _, node := range nodes {
					if item.NodeConditions.EvaluateNode(node) {
						matchedNodes = append(matchedNodes, buildChainLinkPreviewNode(node))
					}
				}
			}
			previewItem.Nodes = matchedNodes
			if len(matchedNodes) > 0 {
				previewItem.Name = matchedNodes[0].Name + " (dynamic)"
			} else {
				previewItem.Name = "(no matching nodes)"
			}

		case "specified_node":
			if name, ok := nodeNameMap[item.NodeID]; ok {
				previewItem.Name = name
				if node, exists := nodeInfoMap[item.NodeID]; exists {
					previewItem.Nodes = []ChainLinkPreviewNode{buildChainLinkPreviewNode(node)}
				}
			} else {
				previewItem.Name = "(node not found)"
			}
		}

		data.Links = append(data.Links, previewItem)
		prevProxyName = previewItem.Name
	}

	// 解析目标配置
	targetConfig, _ := rule.ParseTargetConfig()
	data.TargetType = "all"
	data.TargetInfo = "All nodes"
	if targetConfig != nil {
		data.TargetType = targetConfig.Type
		switch targetConfig.Type {
		case "all":
			data.TargetInfo = "All nodes"
			for _, node := range nodes {
				data.TargetNodes = append(data.TargetNodes, buildChainLinkPreviewNode(node))
			}
		case "conditions":
			data.TargetInfo = "Matching nodes"
			if targetConfig.Conditions != nil {
				for _, node := range nodes {
					if targetConfig.Conditions.EvaluateNode(node) {
						data.TargetNodes = append(data.TargetNodes, buildChainLinkPreviewNode(node))
					}
				}
			}
		case "specified_node":
			if name, ok := nodeNameMap[targetConfig.NodeID]; ok {
				data.TargetInfo = name
				if node, exists := nodeInfoMap[targetConfig.NodeID]; exists {
					data.TargetNodes = []ChainLinkPreviewNode{buildChainLinkPreviewNode(node)}
				}
			} else {
				data.TargetInfo = "(node not found)"
			}
		}
	}

	return data
}

// buildNodeMatchSummary 构建节点匹配摘要
func buildNodeMatchSummary(nodes []models.Node, rules []models.SubscriptionChainRule, nodeNameMap map[int]string) []NodeMatchSummary {
	var summary []NodeMatchSummary

	for _, node := range nodes {
		ms := NodeMatchSummary{
			NodeID:      node.ID,
			NodeName:    node.EffectiveName(),
			LinkCountry: node.LinkCountry,
			Unmatched:   true,
		}

		// 按规则顺序检查（第一个匹配的生效）
		for _, rule := range rules {
			if rule.MatchTargetCondition(node) {
				ms.Unmatched = false
				ms.MatchedRule = rule.Name
				ms.MatchedRuleID = rule.ID

				// 获取入口代理名称
				chainItems, err := rule.ParseChainConfig()
				if err == nil && len(chainItems) > 0 {
					firstItem := chainItems[0]
					switch firstItem.Type {
					case "template_group", "custom_group":
						ms.EntryProxy = firstItem.GroupName
					case "dynamic_node":
						if firstItem.NodeConditions != nil {
							for _, n := range nodes {
								if firstItem.NodeConditions.EvaluateNode(n) {
									ms.EntryProxy = n.EffectiveName() + " (dynamic)"
									break
								}
							}
						}
					case "specified_node":
						if name, ok := nodeNameMap[firstItem.NodeID]; ok {
							ms.EntryProxy = name
						}
					}
				}
				break // 只取第一个匹配的规则
			}
		}

		summary = append(summary, ms)
	}

	return summary
}
