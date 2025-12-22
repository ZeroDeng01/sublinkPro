package api

import (
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订阅ID"})
		return
	}

	rules := models.GetChainRulesBySubscriptionID(subID)

	// 调试日志：记录返回的规则
	utils.Debug("[ChainRule] 获取订阅 %d 的规则，共 %d 条", subID, len(rules))
	for _, r := range rules {
		utils.Debug("[ChainRule] 规则 ID=%d, Name=%s, ChainConfig=%s", r.ID, r.Name, r.ChainConfig)
	}

	c.JSON(http.StatusOK, gin.H{"data": rules})
}

// CreateChainRule 创建链式代理规则
func CreateChainRule(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订阅ID"})
		return
	}

	var rule models.SubscriptionChainRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求数据格式错误: " + err.Error()})
		return
	}

	rule.SubscriptionID = subID

	// 设置默认排序值（最后一个）
	existingRules := models.GetChainRulesBySubscriptionID(subID)
	rule.Sort = len(existingRules)

	if err := rule.Add(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建规则失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// UpdateChainRule 更新链式代理规则
func UpdateChainRule(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订阅ID"})
		return
	}

	ruleIDStr := c.Param("ruleId")
	ruleID, err := strconv.Atoi(ruleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	// 获取现有规则
	var existingRule models.SubscriptionChainRule
	if err := existingRule.GetByID(ruleID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
		return
	}

	// 验证规则属于该订阅
	if existingRule.SubscriptionID != subID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此规则"})
		return
	}

	// 绑定更新数据
	var updateData models.SubscriptionChainRule
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求数据格式错误: " + err.Error()})
		return
	}

	// 更新字段
	existingRule.Name = updateData.Name
	existingRule.Enabled = updateData.Enabled
	existingRule.ChainConfig = updateData.ChainConfig
	existingRule.TargetConfig = updateData.TargetConfig

	// 调试日志：记录更新的数据
	utils.Debug("[ChainRule] 更新规则 ID=%d, 名称=%s", existingRule.ID, existingRule.Name)
	utils.Debug("[ChainRule] ChainConfig: %s", existingRule.ChainConfig)
	utils.Debug("[ChainRule] TargetConfig: %s", existingRule.TargetConfig)

	if err := existingRule.Update(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新规则失败: " + err.Error()})
		return
	}

	utils.Debug("[ChainRule] 规则更新成功，返回数据: ID=%d", existingRule.ID)
	c.JSON(http.StatusOK, gin.H{"data": existingRule})
}

// DeleteChainRule 删除链式代理规则
func DeleteChainRule(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订阅ID"})
		return
	}

	ruleIDStr := c.Param("ruleId")
	ruleID, err := strconv.Atoi(ruleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	// 获取现有规则
	var existingRule models.SubscriptionChainRule
	if err := existingRule.GetByID(ruleID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
		return
	}

	// 验证规则属于该订阅
	if existingRule.SubscriptionID != subID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此规则"})
		return
	}

	if err := existingRule.Delete(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除规则失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// SortChainRules 批量排序链式代理规则
func SortChainRules(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订阅ID"})
		return
	}

	var req struct {
		RuleIDs []int `json:"ruleIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求数据格式错误: " + err.Error()})
		return
	}

	// 验证所有规则都属于该订阅
	for _, ruleID := range req.RuleIDs {
		var rule models.SubscriptionChainRule
		if err := rule.GetByID(ruleID); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在: " + strconv.Itoa(ruleID)})
			return
		}
		if rule.SubscriptionID != subID {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权操作规则: " + strconv.Itoa(ruleID)})
			return
		}
	}

	if err := models.UpdateChainRulesSort(req.RuleIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "排序失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "排序成功"})
}

// GetChainOptions 获取链式代理可用选项
// 返回：模板代理组列表、条件字段列表、节点列表
func GetChainOptions(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订阅ID"})
		return
	}

	// 获取订阅信息
	var sub models.Subcription
	sub.ID = subID
	if err := sub.Find(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订阅不存在"})
		return
	}

	// 获取订阅关联的节点
	if err := sub.GetSub("none"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订阅节点失败: " + err.Error()})
		return
	}

	// 构建节点简要信息列表
	nodeOptions := make([]map[string]interface{}, 0, len(sub.Nodes))
	for _, node := range sub.Nodes {
		nodeOptions = append(nodeOptions, map[string]interface{}{
			"id":          node.ID,
			"name":        node.Name,
			"linkName":    node.LinkName,
			"linkCountry": node.LinkCountry,
			"protocol":    node.Protocol,
			"group":       node.Group,
		})
	}

	// 条件字段列表
	conditionFields := []map[string]string{
		{"value": "name", "label": "节点名称"},
		{"value": "link_name", "label": "原始名称"},
		{"value": "link_country", "label": "国家/地区"},
		{"value": "protocol", "label": "协议类型"},
		{"value": "group", "label": "分组"},
		{"value": "source", "label": "来源"},
		{"value": "speed", "label": "速度 (MB/s)"},
		{"value": "delay_time", "label": "延迟 (ms)"},
		{"value": "speed_status", "label": "测速状态"},
		{"value": "delay_status", "label": "延迟状态"},
		{"value": "tags", "label": "标签"},
		{"value": "link_address", "label": "地址"},
		{"value": "link_host", "label": "主机名"},
		{"value": "link_port", "label": "端口"},
	}

	// 条件操作符列表
	operators := []map[string]string{
		{"value": "equals", "label": "等于"},
		{"value": "not_equals", "label": "不等于"},
		{"value": "contains", "label": "包含"},
		{"value": "not_contains", "label": "不包含"},
		{"value": "regex", "label": "正则匹配"},
		{"value": "greater_than", "label": "大于"},
		{"value": "less_than", "label": "小于"},
		{"value": "greater_or_equal", "label": "大于等于"},
		{"value": "less_or_equal", "label": "小于等于"},
	}

	// 代理组类型
	groupTypes := []map[string]string{
		{"value": "select", "label": "手动选择 (select)"},
		{"value": "url-test", "label": "自动测速 (url-test)"},
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订阅ID"})
		return
	}

	ruleIDStr := c.Param("ruleId")
	ruleID, err := strconv.Atoi(ruleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则ID"})
		return
	}

	// 获取现有规则
	var existingRule models.SubscriptionChainRule
	if err := existingRule.GetByID(ruleID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
		return
	}

	// 验证规则属于该订阅
	if existingRule.SubscriptionID != subID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此规则"})
		return
	}

	// 切换状态
	existingRule.Enabled = !existingRule.Enabled

	if err := existingRule.Update(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新规则失败: " + err.Error()})
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
		resp, err := http.Get(clashTemplate)
		if err != nil {
			return []string{}
		}
		defer resp.Body.Close()
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
	var clashConfig map[string]interface{}
	if err := yaml.Unmarshal([]byte(templateContent), &clashConfig); err != nil {
		return []string{}
	}

	// 提取 proxy-groups 中的 name 字段
	proxyGroups, ok := clashConfig["proxy-groups"].([]interface{})
	if !ok {
		return []string{}
	}

	var groupNames []string
	for _, pg := range proxyGroups {
		if group, ok := pg.(map[string]interface{}); ok {
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

// PreviewChainLinks 预览订阅的整体链式代理配置
func PreviewChainLinks(c *gin.Context) {
	subIDStr := c.Param("id")
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订阅ID"})
		return
	}

	// 获取订阅及其节点
	var sub models.Subcription
	sub.ID = subID
	if err := sub.Find(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订阅不存在"})
		return
	}

	if err := sub.GetSub("none"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订阅节点失败: " + err.Error()})
		return
	}

	// 构建节点名称映射
	nodeNameMap := make(map[int]string)
	nodeInfoMap := make(map[int]models.Node)
	for _, node := range sub.Nodes {
		nodeNameMap[node.ID] = node.Name
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
					if node.Name == targetNode.Name {
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
						previewItem.Nodes = append(previewItem.Nodes, ChainLinkPreviewNode{
							Name:        node.Name,
							Protocol:    node.Protocol,
							LinkCountry: node.LinkCountry,
							DelayTime:   node.DelayTime,
							Speed:       node.Speed,
							Group:       node.Group,
						})
					}
				}
			}

		case "dynamic_node":
			var matchedNodes []ChainLinkPreviewNode
			if item.NodeConditions != nil {
				for _, node := range nodes {
					if item.NodeConditions.EvaluateNode(node) {
						matchedNodes = append(matchedNodes, ChainLinkPreviewNode{
							Name:        node.Name,
							Protocol:    node.Protocol,
							LinkCountry: node.LinkCountry,
							DelayTime:   node.DelayTime,
							Speed:       node.Speed,
							Group:       node.Group,
						})
					}
				}
			}
			previewItem.Nodes = matchedNodes
			if len(matchedNodes) > 0 {
				previewItem.Name = matchedNodes[0].Name + " (动态)"
			} else {
				previewItem.Name = "(无匹配节点)"
			}

		case "specified_node":
			if name, ok := nodeNameMap[item.NodeID]; ok {
				previewItem.Name = name
				if node, exists := nodeInfoMap[item.NodeID]; exists {
					previewItem.Nodes = []ChainLinkPreviewNode{{
						Name:        node.Name,
						Protocol:    node.Protocol,
						LinkCountry: node.LinkCountry,
						DelayTime:   node.DelayTime,
						Speed:       node.Speed,
						Group:       node.Group,
					}}
				}
			} else {
				previewItem.Name = "(节点不存在)"
			}
		}

		data.Links = append(data.Links, previewItem)
		prevProxyName = previewItem.Name
	}

	// 解析目标配置
	targetConfig, _ := rule.ParseTargetConfig()
	data.TargetType = "all"
	data.TargetInfo = "所有节点"
	if targetConfig != nil {
		data.TargetType = targetConfig.Type
		switch targetConfig.Type {
		case "all":
			data.TargetInfo = "所有节点"
			for _, node := range nodes {
				data.TargetNodes = append(data.TargetNodes, ChainLinkPreviewNode{
					Name:        node.Name,
					Protocol:    node.Protocol,
					LinkCountry: node.LinkCountry,
					DelayTime:   node.DelayTime,
					Speed:       node.Speed,
					Group:       node.Group,
				})
			}
		case "conditions":
			data.TargetInfo = "符合条件的节点"
			if targetConfig.Conditions != nil {
				for _, node := range nodes {
					if targetConfig.Conditions.EvaluateNode(node) {
						data.TargetNodes = append(data.TargetNodes, ChainLinkPreviewNode{
							Name:        node.Name,
							Protocol:    node.Protocol,
							LinkCountry: node.LinkCountry,
							DelayTime:   node.DelayTime,
							Speed:       node.Speed,
							Group:       node.Group,
						})
					}
				}
			}
		case "specified_node":
			if name, ok := nodeNameMap[targetConfig.NodeID]; ok {
				data.TargetInfo = name
				if node, exists := nodeInfoMap[targetConfig.NodeID]; exists {
					data.TargetNodes = []ChainLinkPreviewNode{{
						Name:        node.Name,
						Protocol:    node.Protocol,
						LinkCountry: node.LinkCountry,
						DelayTime:   node.DelayTime,
						Speed:       node.Speed,
						Group:       node.Group,
					}}
				}
			} else {
				data.TargetInfo = "(节点不存在)"
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
			NodeName:    node.Name,
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
									ms.EntryProxy = n.Name + " (动态)"
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
