package api

import (
	"encoding/json"
	"strings"
	"sublink/models"
	"sublink/services/notifications"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// GetWebhookConfig 获取Webhook配置
func GetWebhookConfig(c *gin.Context) {
	config, err := notifications.LoadWebhookConfig()
	if err != nil {
		utils.FailWithMsg(c, "获取 Webhook 配置失败: "+err.Error())
		return
	}

	utils.OkDetailed(c, "获取成功", gin.H{
		"webhookUrl":         config.URL,
		"webhookMethod":      config.Method,
		"webhookContentType": config.ContentType,
		"webhookHeaders":     config.Headers,
		"webhookBody":        config.Body,
		"webhookEnabled":     config.Enabled,
		"eventKeys":          config.EventKeys,
		"eventOptions":       notifications.EventCatalogForChannel(notifications.ChannelWebhook),
	})
}

// UpdateWebhookConfig 更新Webhook配置
func UpdateWebhookConfig(c *gin.Context) {
	var req struct {
		WebhookUrl         string   `json:"webhookUrl"`
		WebhookMethod      string   `json:"webhookMethod"`
		WebhookContentType string   `json:"webhookContentType"`
		WebhookHeaders     string   `json:"webhookHeaders"`
		WebhookBody        string   `json:"webhookBody"`
		WebhookEnabled     bool     `json:"webhookEnabled"`
		EventKeys          []string `json:"eventKeys"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	// 验证 Headers 是否为有效的 JSON
	if req.WebhookHeaders != "" {
		var js map[string]interface{}
		if json.Unmarshal([]byte(req.WebhookHeaders), &js) != nil {
			utils.FailWithMsg(c, "Headers 必须是有效的 JSON 格式")
			return
		}
	}

	config := &notifications.WebhookConfig{
		URL:         strings.TrimSpace(req.WebhookUrl),
		Method:      req.WebhookMethod,
		ContentType: req.WebhookContentType,
		Headers:     req.WebhookHeaders,
		Body:        req.WebhookBody,
		Enabled:     req.WebhookEnabled,
		EventKeys:   req.EventKeys,
	}

	if err := notifications.SaveWebhookConfig(config); err != nil {
		utils.FailWithMsg(c, "保存 Webhook 配置失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "保存成功")
}

// TestWebhookConfig 测试Webhook配置
func TestWebhookConfig(c *gin.Context) {
	var req struct {
		WebhookUrl         string `json:"webhookUrl"`
		WebhookMethod      string `json:"webhookMethod"`
		WebhookContentType string `json:"webhookContentType"`
		WebhookHeaders     string `json:"webhookHeaders"`
		WebhookBody        string `json:"webhookBody"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	if req.WebhookHeaders != "" {
		var headers map[string]interface{}
		if err := json.Unmarshal([]byte(req.WebhookHeaders), &headers); err != nil {
			utils.FailWithMsg(c, "Headers 必须是有效的 JSON 格式")
			return
		}
	}

	config := &notifications.WebhookConfig{
		URL:         strings.TrimSpace(req.WebhookUrl),
		Method:      req.WebhookMethod,
		ContentType: req.WebhookContentType,
		Headers:     req.WebhookHeaders,
		Body:        req.WebhookBody,
	}

	payload := notifications.Payload{
		Event:        "test.webhook",
		EventName:    "Webhook 测试",
		Category:     "system",
		CategoryName: "系统测试",
		Severity:     "info",
		Title:        "Sublink Pro Webhook 测试",
		Message:      "这是一条Sublink Pro测试消息，用于验证 Webhook 配置是否正确。",
		Data: map[string]interface{}{
			"test": true,
		},
	}

	if err := notifications.SendWebhook(config, payload); err != nil {
		utils.FailWithMsg(c, "测试失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "测试发送成功")
}

// GetBaseTemplates 获取基础模板配置
func GetBaseTemplates(c *gin.Context) {
	clashTemplate, _ := models.GetSetting("base_template_clash")
	surgeTemplate, _ := models.GetSetting("base_template_surge")

	utils.OkDetailed(c, "获取成功", gin.H{
		"clashTemplate": clashTemplate,
		"surgeTemplate": surgeTemplate,
	})
}

// UpdateBaseTemplate 更新基础模板配置
func UpdateBaseTemplate(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required,oneof=clash surge"`
		Content  string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误：category 必须为 clash 或 surge")
		return
	}

	key := "base_template_" + req.Category
	if err := models.SetSetting(key, req.Content); err != nil {
		utils.FailWithMsg(c, "保存模板失败: "+err.Error())
		return
	}

	categoryName := "Clash"
	if req.Category == "surge" {
		categoryName = "Surge"
	}
	utils.OkWithMsg(c, categoryName+" 基础模板保存成功")
}

// GetSystemDomain 获取系统域名配置
func GetSystemDomain(c *gin.Context) {
	domain, _ := models.GetSetting("system_domain")
	utils.OkWithData(c, gin.H{"systemDomain": domain})
}

// UpdateSystemDomain 更新系统域名配置
func UpdateSystemDomain(c *gin.Context) {
	var req struct {
		SystemDomain string `json:"systemDomain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if err := models.SetSetting("system_domain", req.SystemDomain); err != nil {
		utils.FailWithMsg(c, "保存失败: "+err.Error())
		return
	}
	utils.OkWithMsg(c, "保存成功")
}

// GetNodeDedupConfig 获取节点去重配置
func GetNodeDedupConfig(c *gin.Context) {
	crossAirportDedup, _ := models.GetSetting("cross_airport_dedup_enabled")
	utils.OkDetailed(c, "获取成功", gin.H{
		"crossAirportDedupEnabled": crossAirportDedup != "false",
	})
}

// UpdateNodeDedupConfig 更新节点去重配置
func UpdateNodeDedupConfig(c *gin.Context) {
	var req struct {
		CrossAirportDedupEnabled *bool `json:"crossAirportDedupEnabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if req.CrossAirportDedupEnabled == nil {
		utils.FailWithMsg(c, "缺少必填字段 crossAirportDedupEnabled")
		return
	}
	value := "true"
	if !*req.CrossAirportDedupEnabled {
		value = "false"
	}
	if err := models.SetSetting("cross_airport_dedup_enabled", value); err != nil {
		utils.FailWithMsg(c, "保存失败: "+err.Error())
		return
	}
	utils.OkWithMsg(c, "保存成功")
}
