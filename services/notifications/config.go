package notifications

import (
	"encoding/json"
	"net/http"
	"strings"
	"sublink/models"
	"time"
)

const telegramEventKeysSetting = "telegram_event_keys"

type WebhookConfig struct {
	ID          uint       `json:"id,omitempty"`
	Name        string     `json:"name"`
	URL         string     `json:"webhookUrl"`
	Method      string     `json:"webhookMethod"`
	ContentType string     `json:"webhookContentType"`
	Headers     string     `json:"webhookHeaders"`
	Body        string     `json:"webhookBody"`
	Enabled     bool       `json:"webhookEnabled"`
	EventKeys   []string   `json:"eventKeys"`
	CreatedAt   time.Time  `json:"createdAt,omitempty"`
	UpdatedAt   time.Time  `json:"updatedAt,omitempty"`
	LastTestAt  *time.Time `json:"lastTestAt,omitempty"`
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

func ListWebhookConfigs() ([]WebhookConfig, error) {
	webhooks, err := models.ListWebhooks()
	if err != nil {
		return nil, err
	}

	configs := make([]WebhookConfig, 0, len(webhooks))
	for _, webhook := range webhooks {
		configs = append(configs, webhookModelToConfig(webhook))
	}
	return configs, nil
}

func GetWebhookConfig(id uint) (*WebhookConfig, error) {
	webhook, err := models.GetWebhookByID(id)
	if err != nil {
		return nil, err
	}
	config := webhookModelToConfig(*webhook)
	return &config, nil
}

func CreateWebhookConfig(config *WebhookConfig) (*WebhookConfig, error) {
	model := webhookConfigToModel(config)
	if err := model.Add(); err != nil {
		return nil, err
	}
	created := webhookModelToConfig(model)
	return &created, nil
}

func UpdateWebhookConfig(config *WebhookConfig) (*WebhookConfig, error) {
	model := webhookConfigToModel(config)
	if err := model.Update(); err != nil {
		return nil, err
	}
	updated, err := models.GetWebhookByID(model.ID)
	if err != nil {
		return nil, err
	}
	result := webhookModelToConfig(*updated)
	return &result, nil
}

func DeleteWebhookConfig(id uint) error {
	return (&models.Webhook{ID: id}).Delete()
}

func LoadTelegramEventKeys() ([]string, error) {
	return loadEventKeys(telegramEventKeysSetting, ChannelTelegram)
}

func SaveTelegramEventKeys(keys []string) error {
	return saveEventKeys(telegramEventKeysSetting, ChannelTelegram, keys)
}

func webhookModelToConfig(webhook models.Webhook) WebhookConfig {
	return WebhookConfig{
		ID:          webhook.ID,
		Name:        webhook.Name,
		URL:         webhook.URL,
		Method:      NormalizeWebhookMethod(webhook.Method),
		ContentType: normalizeWebhookContentType(webhook.ContentType),
		Headers:     webhook.Headers,
		Body:        webhook.Body,
		Enabled:     webhook.Enabled,
		EventKeys:   normalizeWebhookEventKeys(webhook.EventKeys),
		CreatedAt:   webhook.CreatedAt,
		UpdatedAt:   webhook.UpdatedAt,
		LastTestAt:  webhook.LastTestAt,
	}
}

func webhookConfigToModel(config *WebhookConfig) models.Webhook {
	return models.Webhook{
		ID:          config.ID,
		Name:        strings.TrimSpace(config.Name),
		URL:         strings.TrimSpace(config.URL),
		Method:      NormalizeWebhookMethod(config.Method),
		ContentType: normalizeWebhookContentType(config.ContentType),
		Headers:     config.Headers,
		Body:        config.Body,
		Enabled:     config.Enabled,
		EventKeys:   encodeWebhookEventKeys(config.EventKeys),
		LastTestAt:  config.LastTestAt,
	}
}

func normalizeWebhookEventKeys(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return DefaultEventKeys(ChannelWebhook)
	}

	var keys []string
	if err := json.Unmarshal([]byte(raw), &keys); err != nil {
		return DefaultEventKeys(ChannelWebhook)
	}
	return NormalizeEventKeys(ChannelWebhook, keys)
}

func encodeWebhookEventKeys(keys []string) string {
	normalized := DefaultEventKeys(ChannelWebhook)
	if keys != nil {
		normalized = NormalizeEventKeys(ChannelWebhook, keys)
	}
	value, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(value)
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
