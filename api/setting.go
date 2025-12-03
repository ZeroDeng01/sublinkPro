package api

import (
	"encoding/json"
	"sublink/models"
	"sublink/services/sse"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// GetWebhookConfig 获取Webhook配置
func GetWebhookConfig(c *gin.Context) {
	webhookUrl, _ := models.GetSetting("webhook_url")
	webhookMethod, _ := models.GetSetting("webhook_method")
	if webhookMethod == "" {
		webhookMethod = "POST"
	}
	webhookContentType, _ := models.GetSetting("webhook_content_type")
	if webhookContentType == "" {
		webhookContentType = "application/json"
	}
	webhookHeaders, _ := models.GetSetting("webhook_headers")
	webhookBody, _ := models.GetSetting("webhook_body")
	webhookEnabledStr, _ := models.GetSetting("webhook_enabled")
	webhookEnabled := webhookEnabledStr == "true"

	utils.OkDetailed(c, "获取成功", gin.H{
		"webhookUrl":         webhookUrl,
		"webhookMethod":      webhookMethod,
		"webhookContentType": webhookContentType,
		"webhookHeaders":     webhookHeaders,
		"webhookBody":        webhookBody,
		"webhookEnabled":     webhookEnabled,
	})
}

// UpdateWebhookConfig 更新Webhook配置
func UpdateWebhookConfig(c *gin.Context) {
	var req struct {
		WebhookUrl         string `json:"webhookUrl"`
		WebhookMethod      string `json:"webhookMethod"`
		WebhookContentType string `json:"webhookContentType"`
		WebhookHeaders     string `json:"webhookHeaders"`
		WebhookBody        string `json:"webhookBody"`
		WebhookEnabled     bool   `json:"webhookEnabled"`
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

	if err := models.SetSetting("webhook_url", req.WebhookUrl); err != nil {
		utils.FailWithMsg(c, "保存 URL 失败")
		return
	}
	if err := models.SetSetting("webhook_method", req.WebhookMethod); err != nil {
		utils.FailWithMsg(c, "保存 Method 失败")
		return
	}
	if err := models.SetSetting("webhook_content_type", req.WebhookContentType); err != nil {
		utils.FailWithMsg(c, "保存 Content-Type 失败")
		return
	}
	if err := models.SetSetting("webhook_headers", req.WebhookHeaders); err != nil {
		utils.FailWithMsg(c, "保存 Headers 失败")
		return
	}
	if err := models.SetSetting("webhook_body", req.WebhookBody); err != nil {
		utils.FailWithMsg(c, "保存 Body 失败")
		return
	}
	enabledStr := "false"
	if req.WebhookEnabled {
		enabledStr = "true"
	}
	if err := models.SetSetting("webhook_enabled", enabledStr); err != nil {
		utils.FailWithMsg(c, "保存启用状态失败")
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

	// 构造配置对象
	config := map[string]string{
		"url":         req.WebhookUrl,
		"method":      req.WebhookMethod,
		"contentType": req.WebhookContentType,
		"headers":     req.WebhookHeaders,
		"body":        req.WebhookBody,
	}

	// 构造测试 Payload
	payload := sse.NotificationPayload{
		Event:   "test_webhook",
		Title:   "Sublink Pro Webhook 测试",
		Message: "这是一条Sublink Pro测试消息，用于验证 Webhook 配置是否正确。",
		Data: map[string]interface{}{
			"test": true,
		},
	}

	if err := sse.SendWebhook(config, "test_webhook", payload); err != nil {
		utils.FailWithMsg(c, "测试失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "测试发送成功")
}
