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
		utils.FailWithMsg(c, "Failed to fetch webhook list: "+err.Error())
		return
	}

	utils.OkDetailed(c, "Fetched successfully", gin.H{
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
		utils.FailWithMsg(c, "Failed to create webhook: "+err.Error())
		return
	}

	utils.OkDetailed(c, "Created successfully", created)
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
		utils.FailWithMsg(c, "Failed to update webhook: "+err.Error())
		return
	}

	utils.OkDetailed(c, "Updated successfully", updated)
}

func DeleteWebhook(c *gin.Context) {
	id, ok := parseWebhookID(c)
	if !ok {
		return
	}

	if err := notifications.DeleteWebhookConfig(id); err != nil {
		utils.FailWithMsg(c, "Failed to delete webhook: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "Deleted successfully")
}

func TestWebhookByID(c *gin.Context) {
	id, ok := parseWebhookID(c)
	if !ok {
		return
	}

	config, err := notifications.GetWebhookConfig(id)
	if err != nil {
		utils.FailWithMsg(c, "Failed to fetch webhook: "+err.Error())
		return
	}

	payload := notifications.Payload{
		Event:        "test.webhook",
		EventName:    "Webhook test",
		Category:     "system",
		CategoryName: "System test",
		Severity:     "info",
		Title:        "Sublink Pro webhook test",
		Message:      "This is a Sublink Pro test message used to verify that the webhook configuration is correct.",
		Data: map[string]any{
			"test": true,
		},
	}

	if err := notifications.SendWebhook(config, payload); err != nil {
		utils.FailWithMsg(c, "Test failed: "+err.Error())
		return
	}

	config.LastTestAt = new(time.Now())
	if _, err := notifications.UpdateWebhookConfig(config); err != nil {
		utils.FailWithMsg(c, "Failed to update test time: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "Test message sent")
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
		var js map[string]any
		if json.Unmarshal([]byte(req.WebhookHeaders), &js) != nil {
			return &webhookValidationError{message: "Headers must be valid JSON"}
		}
	}
	if req.WebhookEnabled && strings.TrimSpace(req.WebhookUrl) == "" {
		return &webhookValidationError{message: "URL is required when the webhook is enabled"}
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
		utils.FailWithMsg(c, "Invalid webhook ID")
		return 0, false
	}
	return uint(idValue), true
}
