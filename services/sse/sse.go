package sse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sublink/models"
	"sublink/utils"
	"sync"
	"time"
)

// SSEBroker manages Server-Sent Events clients and broadcasting
type SSEBroker struct {
	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte

	// New client connections
	newClients chan chan []byte

	// Closed client connections
	closingClients chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool

	// Mutex to protect the clients map
	mutex sync.Mutex
}

var (
	sseBroker *SSEBroker
	sseOnce   sync.Once
)

// GetSSEBroker returns the singleton instance of the SSEBroker
func GetSSEBroker() *SSEBroker {
	sseOnce.Do(func() {
		sseBroker = &SSEBroker{
			Notifier:       make(chan []byte, 100), // Buffer increased for rapid progress events
			newClients:     make(chan chan []byte),
			closingClients: make(chan chan []byte),
			clients:        make(map[chan []byte]bool),
		}
	})
	return sseBroker
}

// Listen starts the broker to listen for incoming and closing clients
func (broker *SSEBroker) Listen() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case s := <-broker.newClients:
			// A new client has connected.
			// Register their message channel
			broker.mutex.Lock()
			broker.clients[s] = true
			broker.mutex.Unlock()
			utils.Info("Client added. %d registered clients", len(broker.clients))

		case s := <-broker.closingClients:
			// A client has detached and we want to stop sending them messages.
			broker.mutex.Lock()
			delete(broker.clients, s)
			broker.mutex.Unlock()
			utils.Info("Removed client. %d registered clients", len(broker.clients))

		case event := <-broker.Notifier:
			// We got a new event from the outside!
			// Send event to all connected clients
			broker.mutex.Lock()
			for clientMessageChan := range broker.clients {
				select {
				case clientMessageChan <- event:
				default:
					// If the client's channel buffer is full, just skip this message
					// The client will catch up with later messages
					// Don't disconnect - they might just be temporarily slow
				}
			}
			broker.mutex.Unlock()

		case <-ticker.C:
			// Send heartbeat to all clients
			broker.mutex.Lock()
			heartbeatMsg := []byte("event: heartbeat\ndata: ping\n\n")
			for clientMessageChan := range broker.clients {
				select {
				case clientMessageChan <- heartbeatMsg:
				default:
					// Client buffer full, skip heartbeat (they're probably processing messages)
				}
			}
			broker.mutex.Unlock()
		}
	}
}

// AddClient adds a client to the broker
func (broker *SSEBroker) AddClient(clientChan chan []byte) {
	broker.newClients <- clientChan
}

// RemoveClient removes a client from the broker
func (broker *SSEBroker) RemoveClient(clientChan chan []byte) {
	broker.closingClients <- clientChan
}

// Broadcast sends a message to all clients
func (broker *SSEBroker) Broadcast(message string) {
	broker.Notifier <- []byte(message)
}

// NotificationPayload defines the standard payload for notifications
type NotificationPayload struct {
	Event   string      `json:"event"`
	Title   string      `json:"title"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Time    string      `json:"time"`
}

// BroadcastEvent sends a JSON message to all clients
// You can use this to send structured data
func (broker *SSEBroker) BroadcastEvent(event string, payload NotificationPayload) {
	// Ensure time is set
	if payload.Time == "" {
		payload.Time = time.Now().Format("2006-01-02 15:04:05")
	}
	// Ensure event is set in payload
	if payload.Event == "" {
		payload.Event = event
	}

	// Trigger Webhook
	go TriggerWebhook(event, payload)

	// Trigger Telegram notification
	go TriggerTelegram(event, payload)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		utils.Error("Error marshaling SSE payload: %v", err)
		return
	}
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, jsonData)
	broker.Notifier <- []byte(msg)
}

// ProgressPayload defines the payload for progress updates
// This is separate from NotificationPayload to avoid webhook triggers
type ProgressPayload struct {
	TaskID      string      `json:"taskId"`              // 任务唯一标识
	TaskType    string      `json:"taskType"`            // 任务类型: speed_test, sub_update
	TaskName    string      `json:"taskName"`            // 任务名称（如订阅名称）
	Status      string      `json:"status"`              // started, progress, completed, error
	Current     int         `json:"current"`             // 当前进度
	Total       int         `json:"total"`               // 总数
	CurrentItem string      `json:"currentItem"`         // 当前处理的项目名称
	Result      interface{} `json:"result"`              // 当前项目的结果
	Message     string      `json:"message"`             // 可选的消息
	Time        string      `json:"time"`                // 时间戳
	StartTime   int64       `json:"startTime,omitempty"` // 任务开始时间戳(毫秒)
}

// BroadcastProgress sends a progress update to all clients
// This does NOT trigger webhooks, unlike BroadcastEvent
func (broker *SSEBroker) BroadcastProgress(payload ProgressPayload) {
	// Ensure time is set
	if payload.Time == "" {
		payload.Time = time.Now().Format("2006-01-02 15:04:05")
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		utils.Error("Error marshaling SSE progress payload: %v", err)
		return
	}
	msg := fmt.Sprintf("event: task_progress\ndata: %s\n\n", jsonData)
	broker.Notifier <- []byte(msg)
}

// TriggerWebhook sends a webhook notification// TriggerWebhook 触发 Webhook 通知
func TriggerWebhook(event string, payload NotificationPayload) {
	// 获取系统设置中的 Webhook 配置
	webhookUrl, _ := models.GetSetting("webhook_url")
	webhookEnabledStr, _ := models.GetSetting("webhook_enabled")

	if webhookUrl == "" || webhookEnabledStr != "true" {
		return
	}

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

	// 构造配置对象
	config := map[string]string{
		"url":         webhookUrl,
		"method":      webhookMethod,
		"contentType": webhookContentType,
		"headers":     webhookHeaders,
		"body":        webhookBody,
	}

	go SendWebhook(config, event, payload)
}

// SendWebhook sends a webhook notification synchronously and returns error
func SendWebhook(config map[string]string, event string, payload NotificationPayload) error {
	// 准备数据用于替换
	data := map[string]interface{}{
		"event":   event,
		"title":   payload.Title,
		"message": payload.Message,
		"time":    payload.Time,
		"data":    payload.Data,
	}

	// 替换 URL 中的变量
	urlStr := config["url"]
	urlStr = strings.ReplaceAll(urlStr, "{{title}}", url.QueryEscape(payload.Title))
	urlStr = strings.ReplaceAll(urlStr, "{{message}}", url.QueryEscape(payload.Message))
	urlStr = strings.ReplaceAll(urlStr, "{{event}}", url.QueryEscape(event))
	urlStr = strings.ReplaceAll(urlStr, "{{time}}", url.QueryEscape(payload.Time))

	// 处理 Body
	bodyStr := config["body"]
	if bodyStr == "" {
		// 默认 Body
		jsonBytes, _ := json.Marshal(data)
		bodyStr = string(jsonBytes)
	} else {
		// Determine escape function based on content type
		var escapeFunc func(string) string
		contentType := strings.ToLower(config["contentType"])

		if strings.Contains(contentType, "application/json") {
			escapeFunc = func(s string) string {
				b, _ := json.Marshal(s)
				// Remove surrounding quotes
				if len(b) >= 2 {
					return string(b[1 : len(b)-1])
				}
				return string(b)
			}
		} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			escapeFunc = url.QueryEscape
		} else {
			escapeFunc = func(s string) string { return s }
		}

		// 简单模板替换
		bodyStr = strings.ReplaceAll(bodyStr, "{{title}}", escapeFunc(payload.Title))
		bodyStr = strings.ReplaceAll(bodyStr, "{{message}}", escapeFunc(payload.Message))
		bodyStr = strings.ReplaceAll(bodyStr, "{{event}}", escapeFunc(event))
		bodyStr = strings.ReplaceAll(bodyStr, "{{time}}", escapeFunc(payload.Time))

		// 支持 {{json .}} 替换为完整 JSON
		if strings.Contains(bodyStr, "{{json .}}") {
			jsonBytes, _ := json.Marshal(data)
			bodyStr = strings.ReplaceAll(bodyStr, "{{json .}}", string(jsonBytes))
		}
	}

	req, err := http.NewRequest(config["method"], urlStr, bytes.NewBuffer([]byte(bodyStr)))
	if err != nil {
		utils.Error("创建 Webhook 请求失败: %v", err)
		return err
	}

	req.Header.Set("Content-Type", config["contentType"])
	req.Header.Set("User-Agent", "Sublink-Webhook/1.0")

	// 处理 Headers
	if config["headers"] != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(config["headers"]), &headers); err == nil {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		utils.Error("发送 Webhook 失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		utils.Warn("Webhook 发送失败，状态码: %d", resp.StatusCode)
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	} else {
		utils.Info("Webhook sent successfully to %s", urlStr)
	}
	return nil
}

// TelegramNotifier 用于发送 Telegram 通知的函数类型
var TelegramNotifier func(event string, payload NotificationPayload)

// TriggerTelegram 触发 Telegram 通知
func TriggerTelegram(event string, payload NotificationPayload) {
	if TelegramNotifier != nil {
		TelegramNotifier(event, payload)
	}
}
