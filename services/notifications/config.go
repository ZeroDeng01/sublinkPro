package notifications

import (
	"encoding/json"
	"net/http"
	"strings"
	"sublink/models"
	"time"
)

const (
	webhookEventKeysSetting  = "webhook_event_keys"
	telegramEventKeysSetting = "telegram_event_keys"
)

type WebhookConfig struct {
	URL         string   `json:"webhookUrl"`
	Method      string   `json:"webhookMethod"`
	ContentType string   `json:"webhookContentType"`
	Headers     string   `json:"webhookHeaders"`
	Body        string   `json:"webhookBody"`
	Enabled     bool     `json:"webhookEnabled"`
	EventKeys   []string `json:"eventKeys"`
}

func nowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func NormalizeWebhookMethod(method string) string {
	normalized := strings.ToUpper(strings.TrimSpace(method))
	if normalized == "" {
		return http.MethodPost
	}
	return normalized
}

func LoadWebhookConfig() (*WebhookConfig, error) {
	url, _ := models.GetSetting("webhook_url")
	method, _ := models.GetSetting("webhook_method")
	contentType, _ := models.GetSetting("webhook_content_type")
	headers, _ := models.GetSetting("webhook_headers")
	body, _ := models.GetSetting("webhook_body")
	enabledStr, _ := models.GetSetting("webhook_enabled")
	eventKeys, err := loadEventKeys(webhookEventKeysSetting, ChannelWebhook)
	if err != nil {
		return nil, err
	}

	if contentType == "" {
		contentType = "application/json"
	}

	return &WebhookConfig{
		URL:         url,
		Method:      NormalizeWebhookMethod(method),
		ContentType: contentType,
		Headers:     headers,
		Body:        body,
		Enabled:     enabledStr == "true",
		EventKeys:   eventKeys,
	}, nil
}

func SaveWebhookConfig(config *WebhookConfig) error {
	if err := models.SetSetting("webhook_url", config.URL); err != nil {
		return err
	}
	if err := models.SetSetting("webhook_method", NormalizeWebhookMethod(config.Method)); err != nil {
		return err
	}
	if err := models.SetSetting("webhook_content_type", normalizeWebhookContentType(config.ContentType)); err != nil {
		return err
	}
	if err := models.SetSetting("webhook_headers", config.Headers); err != nil {
		return err
	}
	if err := models.SetSetting("webhook_body", config.Body); err != nil {
		return err
	}

	enabledStr := "false"
	if config.Enabled {
		enabledStr = "true"
	}
	if err := models.SetSetting("webhook_enabled", enabledStr); err != nil {
		return err
	}

	return saveEventKeys(webhookEventKeysSetting, ChannelWebhook, config.EventKeys)
}

func LoadTelegramEventKeys() ([]string, error) {
	return loadEventKeys(telegramEventKeysSetting, ChannelTelegram)
}

func SaveTelegramEventKeys(keys []string) error {
	return saveEventKeys(telegramEventKeysSetting, ChannelTelegram, keys)
}

func normalizeWebhookContentType(contentType string) string {
	normalized := strings.TrimSpace(contentType)
	if normalized == "" {
		return "application/json"
	}
	return normalized
}

func loadEventKeys(settingKey string, channel Channel) ([]string, error) {
	rawValue, err := models.GetSetting(settingKey)
	if err != nil || strings.TrimSpace(rawValue) == "" {
		return DefaultEventKeys(channel), nil
	}

	var keys []string
	if err := json.Unmarshal([]byte(rawValue), &keys); err != nil {
		return DefaultEventKeys(channel), nil
	}

	return NormalizeEventKeys(channel, keys), nil
}

func saveEventKeys(settingKey string, channel Channel, keys []string) error {
	normalized := DefaultEventKeys(channel)
	if keys != nil {
		normalized = NormalizeEventKeys(channel, keys)
	}
	value, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	return models.SetSetting(settingKey, string(value))
}
