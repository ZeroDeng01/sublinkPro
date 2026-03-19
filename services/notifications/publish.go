package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sublink/services/sse"
	"sublink/utils"
	"sync"
	"time"
)

const webhookDispatchConcurrency = 4

var telegramSender func(eventKey string, payload Payload)

func RegisterTelegramSender(sender func(eventKey string, payload Payload)) {
	telegramSender = sender
}

func Publish(eventKey string, payload Payload) {
	enrichedPayload := FillPayloadMeta(eventKey, payload)

	sse.GetSSEBroker().BroadcastJSONEvent("notification", enrichedPayload)
	go TriggerWebhook(eventKey, enrichedPayload)
	go TriggerTelegram(eventKey, enrichedPayload)
}

func TriggerWebhook(eventKey string, payload Payload) {
	configs, err := ListWebhookConfigs()
	if err != nil {
		utils.Warn("加载 Webhook 配置失败: %v", err)
		return
	}
	if len(configs) == 0 {
		return
	}

	sem := make(chan struct{}, webhookDispatchConcurrency)
	var wg sync.WaitGroup

	for _, config := range configs {
		if !config.Enabled || strings.TrimSpace(config.URL) == "" {
			continue
		}
		if !IsEventEnabled(config.EventKeys, eventKey) {
			continue
		}

		cfg := config
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			if err := SendWebhook(&cfg, payload); err != nil {
				name := strings.TrimSpace(cfg.Name)
				if name == "" {
					name = fmt.Sprintf("#%d", cfg.ID)
				}
				utils.Warn("发送 Webhook[%s] 通知失败: %v", name, err)
			}
		}()
	}

	wg.Wait()
}

func SendWebhook(config *WebhookConfig, payload Payload) error {
	enrichedPayload := FillPayloadMeta(payload.Event, payload)
	data := map[string]interface{}{
		"event":        enrichedPayload.Event,
		"eventName":    enrichedPayload.EventName,
		"category":     enrichedPayload.Category,
		"categoryName": enrichedPayload.CategoryName,
		"severity":     enrichedPayload.Severity,
		"title":        enrichedPayload.Title,
		"message":      enrichedPayload.Message,
		"time":         enrichedPayload.Time,
		"data":         enrichedPayload.Data,
	}

	urlStr := replaceTemplateVars(config.URL, enrichedPayload, url.QueryEscape)
	method := NormalizeWebhookMethod(config.Method)
	bodyStr := buildWebhookBody(config, enrichedPayload, data, method)

	var body io.Reader
	if bodyStr != "" {
		body = bytes.NewBufferString(bodyStr)
	}

	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		utils.Error("创建 Webhook 请求失败: %v", err)
		return err
	}

	req.Header.Set("Content-Type", normalizeWebhookContentType(config.ContentType))
	req.Header.Set("User-Agent", "Sublink-Webhook/2.0")

	if strings.TrimSpace(config.Headers) != "" {
		var headers map[string]interface{}
		if err := json.Unmarshal([]byte(config.Headers), &headers); err != nil {
			return fmt.Errorf("Headers 不是有效的 JSON: %w", err)
		}
		for key, value := range headers {
			req.Header.Set(key, fmt.Sprint(value))
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		utils.Error("发送 Webhook 失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		utils.Warn("Webhook 发送失败，状态码: %d", resp.StatusCode)
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	utils.Info("Webhook sent successfully to %s", urlStr)
	return nil
}

func TriggerTelegram(eventKey string, payload Payload) {
	if telegramSender == nil {
		return
	}

	eventKeys, err := LoadTelegramEventKeys()
	if err != nil {
		utils.Warn("加载 Telegram 事件配置失败: %v", err)
		return
	}
	if !IsEventEnabled(eventKeys, eventKey) {
		return
	}

	telegramSender(eventKey, FillPayloadMeta(eventKey, payload))
}

func buildWebhookBody(config *WebhookConfig, payload Payload, data map[string]interface{}, method string) string {
	customBody := config.Body
	if strings.TrimSpace(customBody) == "" {
		if method == http.MethodGet {
			return ""
		}
		jsonBytes, _ := json.Marshal(data)
		return string(jsonBytes)
	}

	var escapeFunc func(string) string
	contentType := strings.ToLower(config.ContentType)
	switch {
	case strings.Contains(contentType, "application/json"):
		escapeFunc = func(value string) string {
			jsonBytes, _ := json.Marshal(value)
			if len(jsonBytes) >= 2 {
				return string(jsonBytes[1 : len(jsonBytes)-1])
			}
			return string(jsonBytes)
		}
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		escapeFunc = url.QueryEscape
	default:
		escapeFunc = func(value string) string { return value }
	}

	body := replaceTemplateVars(customBody, payload, escapeFunc)
	if strings.Contains(body, "{{json .}}") {
		jsonBytes, _ := json.Marshal(data)
		body = strings.ReplaceAll(body, "{{json .}}", string(jsonBytes))
	}
	return body
}

func replaceTemplateVars(template string, payload Payload, replacer func(string) string) string {
	replaced := strings.ReplaceAll(template, "{{event}}", replacer(payload.Event))
	replaced = strings.ReplaceAll(replaced, "{{event_name}}", replacer(payload.EventName))
	replaced = strings.ReplaceAll(replaced, "{{eventName}}", replacer(payload.EventName))
	replaced = strings.ReplaceAll(replaced, "{{category}}", replacer(payload.Category))
	replaced = strings.ReplaceAll(replaced, "{{category_name}}", replacer(payload.CategoryName))
	replaced = strings.ReplaceAll(replaced, "{{categoryName}}", replacer(payload.CategoryName))
	replaced = strings.ReplaceAll(replaced, "{{severity}}", replacer(payload.Severity))
	replaced = strings.ReplaceAll(replaced, "{{title}}", replacer(payload.Title))
	replaced = strings.ReplaceAll(replaced, "{{message}}", replacer(payload.Message))
	replaced = strings.ReplaceAll(replaced, "{{time}}", replacer(payload.Time))
	return replaced
}
