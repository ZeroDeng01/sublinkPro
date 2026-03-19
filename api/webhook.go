package api

import (
	"encoding/json"
	"strconv"
	"strings"
	"sublink/services/notifications"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type webhookRequest struct {
	Name               string   `json:"name"`
	WebhookUrl         string   `json:"webhookUrl"`
	WebhookMethod      string   `json:"webhookMethod"`
	WebhookContentType string   `json:"webhookContentType"`
	WebhookHeaders     string   `json:"webhookHeaders"`
	WebhookBody        string   `json:"webhookBody"`
	WebhookEnabled     bool     `json:"webhookEnabled"`
	EventKeys          []string `json:"eventKeys"`
}

func ListWebhooks(c *gin.Context) {
	items, err := notifications.ListWebhookConfigs()
	if err != nil {
		utils.FailWithMsg(c, "获取 Webhook 列表失败: "+err.Error())
		return
	}

	utils.OkDetailed(c, "获取成功", gin.H{
		"items":        items,
		"eventOptions": notifications.EventCatalogForChannel(notifications.ChannelWebhook),
	})
}

func CreateWebhook(c *gin.Context) {
	config, err := bindWebhookRequest(c)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	created, err := notifications.CreateWebhookConfig(config)
	if err != nil {
		utils.FailWithMsg(c, "创建 Webhook 失败: "+err.Error())
		return
	}

	utils.OkDetailed(c, "创建成功", created)
}

func UpdateWebhook(c *gin.Context) {
	id, ok := parseWebhookID(c)
	if !ok {
		return
	}

	config, err := bindWebhookRequest(c)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	config.ID = id

	updated, err := notifications.UpdateWebhookConfig(config)
	if err != nil {
		utils.FailWithMsg(c, "更新 Webhook 失败: "+err.Error())
		return
	}

	utils.OkDetailed(c, "更新成功", updated)
}

func DeleteWebhook(c *gin.Context) {
	id, ok := parseWebhookID(c)
	if !ok {
		return
	}

	if err := notifications.DeleteWebhookConfig(id); err != nil {
		utils.FailWithMsg(c, "删除 Webhook 失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "删除成功")
}

func TestWebhookByID(c *gin.Context) {
	id, ok := parseWebhookID(c)
	if !ok {
		return
	}

	config, err := notifications.GetWebhookConfig(id)
	if err != nil {
		utils.FailWithMsg(c, "获取 Webhook 失败: "+err.Error())
		return
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

	now := time.Now()
	config.LastTestAt = &now
	if _, err := notifications.UpdateWebhookConfig(config); err != nil {
		utils.FailWithMsg(c, "更新测试时间失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "测试发送成功")
}

func bindWebhookRequest(c *gin.Context) (*notifications.WebhookConfig, error) {
	var req webhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	if err := normalizeWebhookRequestError(req); err != nil {
		return nil, err
	}

	url := strings.TrimSpace(req.WebhookUrl)

	return &notifications.WebhookConfig{
		Name:        strings.TrimSpace(req.Name),
		URL:         url,
		Method:      req.WebhookMethod,
		ContentType: req.WebhookContentType,
		Headers:     req.WebhookHeaders,
		Body:        req.WebhookBody,
		Enabled:     req.WebhookEnabled,
		EventKeys:   req.EventKeys,
	}, nil
}

func normalizeWebhookRequestError(req webhookRequest) error {
	if req.WebhookHeaders != "" {
		var js map[string]interface{}
		if json.Unmarshal([]byte(req.WebhookHeaders), &js) != nil {
			return &webhookValidationError{message: "Headers 必须是有效的 JSON 格式"}
		}
	}
	if req.WebhookEnabled && strings.TrimSpace(req.WebhookUrl) == "" {
		return &webhookValidationError{message: "启用 Webhook 时需填写 URL"}
	}
	return nil
}

type webhookValidationError struct {
	message string
}

func (e *webhookValidationError) Error() string {
	return e.message
}

func parseWebhookID(c *gin.Context) (uint, bool) {
	idValue, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || idValue == 0 {
		utils.FailWithMsg(c, "Webhook ID 无效")
		return 0, false
	}
	return uint(idValue), true
}
