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
	if err := sub.GetSub(); err != nil {
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
