package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/services/notifications"
	"sublink/utils"
	"sync"
	"time"
)

const (
	TelegramAPIBase = "https://api.telegram.org/bot"
)

// TelegramBot Telegram 机器人核心结构
type TelegramBot struct {
	Token     string
	ChatID    int64
	UseProxy  bool
	ProxyLink string

	client        *http.Client
	pollingActive bool
	stopChan      chan struct{}
	mutex         sync.RWMutex
	connected     bool
	lastError     string
	updateOffset  int64
	botUsername   string // 机器人用户名
	botID         int64  // 机器人ID
}

// Config Telegram 配置
type Config struct {
	Enabled   bool
	BotToken  string
	ChatID    int64
	UseProxy  bool
	ProxyLink string
	EventKeys []string
}

// 全局机器人实例
var (
	globalBot *TelegramBot
	botMutex  sync.RWMutex
)

// GetBot 获取全局机器人实例
func GetBot() *TelegramBot {
	botMutex.RLock()
	defer botMutex.RUnlock()
	return globalBot
}

// InitBot 初始化 Telegram 机器人
func InitBot() error {
	// 启动后台监控协程
	go connectionMonitor()
	return nil
}

// connectionMonitor Telegram 状态监控
func connectionMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 首次立即执行
	checkAndStart()

	for range ticker.C {
		checkAndStart()
	}
}

// checkAndStart 检查并启动机器人
func checkAndStart() {
	config, err := LoadConfig()
	if err != nil {
		utils.Error("[Telegram] Failed to load config: %v", err)
		return
	}

	// 如果未启用，确保机器人停止
	if !config.Enabled || config.BotToken == "" {
		if GetBot() != nil {
			StopBot()
			utils.Info("[Telegram] Bot is disabled; stopping")
		}
		return
	}

	// 如果启用但未运行，尝试启动
	if GetBot() == nil {
		utils.Info("[Telegram] Bot is not running; attempting to start...")
		if err := StartBot(config); err != nil {
			utils.Error("[Telegram] Failed to start: %v", err)
		} else {
			utils.Info("[Telegram] Started successfully")
		}
	}
}

// LoadConfig 从数据库加载配置
func LoadConfig() (*Config, error) {
	enabled, _ := models.GetSetting("telegram_enabled")
	botToken, _ := models.GetSetting("telegram_bot_token")
	chatIDStr, _ := models.GetSetting("telegram_chat_id")
	useProxy, _ := models.GetSetting("telegram_use_proxy")
	proxyLink, _ := models.GetSetting("telegram_proxy_link")
	eventKeys, err := notifications.LoadTelegramEventKeys()
	if err != nil {
		return nil, err
	}

	var chatID int64
	if chatIDStr != "" {
		chatID, _ = strconv.ParseInt(chatIDStr, 10, 64)
	}

	return &Config{
		Enabled:   enabled == "true",
		BotToken:  botToken,
		ChatID:    chatID,
		UseProxy:  useProxy == "true",
		ProxyLink: proxyLink,
		EventKeys: eventKeys,
	}, nil
}

// SaveConfig 保存配置到数据库
func SaveConfig(config *Config) error {
	enabledStr := "false"
	if config.Enabled {
		enabledStr = "true"
	}
	useProxyStr := "false"
	if config.UseProxy {
		useProxyStr = "true"
	}

	if err := models.SetSetting("telegram_enabled", enabledStr); err != nil {
		return err
	}
	if err := models.SetSetting("telegram_bot_token", config.BotToken); err != nil {
		return err
	}
	if err := models.SetSetting("telegram_chat_id", strconv.FormatInt(config.ChatID, 10)); err != nil {
		return err
	}
	if err := models.SetSetting("telegram_use_proxy", useProxyStr); err != nil {
		return err
	}
	if err := models.SetSetting("telegram_proxy_link", config.ProxyLink); err != nil {
		return err
	}
	if err := notifications.SaveTelegramEventKeys(config.EventKeys); err != nil {
		return err
	}

	return nil
}

// StartBot 启动机器人
func StartBot(config *Config) error {
	botMutex.Lock()
	defer botMutex.Unlock()

	// 如果已有机器人在运行，先停止
	if globalBot != nil {
		globalBot.Stop()
	}

	// 创建 HTTP 客户端（可能带代理）
	client, usedProxy, err := utils.CreateProxyHTTPClient(config.UseProxy, config.ProxyLink, 60*time.Second)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %v", err)
	}

	// Telegram 必须通过代理访问（国内用户），如果配置了代理但未能获取则返回错误
	if config.UseProxy && usedProxy == "" {
		return fmt.Errorf("proxy usage is enabled but no proxy link is available; configure a proxy node or make sure a node passes the speed test")
	}

	if config.UseProxy {
		utils.Info("[Telegram] Using proxy connection: %s", usedProxy)
	}

	bot := &TelegramBot{
		Token:     config.BotToken,
		ChatID:    config.ChatID,
		UseProxy:  config.UseProxy,
		ProxyLink: config.ProxyLink,
		client:    client,
		stopChan:  make(chan struct{}),
	}

	// 验证 Token
	if err := bot.validateToken(); err != nil {
		bot.setError(err.Error())
		return fmt.Errorf("failed to verify token: %v", err)
	}

	// 设置命令菜单
	if err := bot.SetCommands(); err != nil {
		utils.Warn("Failed to set command menu: %v", err)
	}

	globalBot = bot

	// 启动长轮询
	go bot.startPolling()

	utils.Info("Telegram bot started")
	return nil
}

// StopBot 停止全局机器人
func StopBot() {
	botMutex.Lock()
	defer botMutex.Unlock()

	if globalBot != nil {
		globalBot.Stop()
		globalBot = nil
	}
}

// validateToken 验证 Token 是否有效
func (b *TelegramBot) validateToken() error {
	resp, err := b.apiRequest("getMe", nil)
	if err != nil {
		return err
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !result.OK {
		return fmt.Errorf("invalid token")
	}

	utils.Info("Telegram bot verified: @%s", result.Result.Username)
	b.mutex.Lock()
	b.botUsername = result.Result.Username
	b.botID = result.Result.ID
	b.mutex.Unlock()
	b.setConnected(true)
	return nil
}

// SetCommands 设置机器人命令菜单
func (b *TelegramBot) SetCommands() error {
	commands := []map[string]string{
		{"command": "start", "description": "🚀 Get started"},
		{"command": "help", "description": "❓ Help"},
		{"command": "stats", "description": "📊 Dashboard stats"},
		{"command": "monitor", "description": "🖥️ System monitor"},
		{"command": "profiles", "description": "⚡ Check profiles"},
		{"command": "subscriptions", "description": "📋 Subscriptions"},
		{"command": "nodes", "description": "🌐 Nodes"},
		{"command": "tags", "description": "🏷️ Tag rules"},
		{"command": "tasks", "description": "📝 Tasks"},
		{"command": "airports", "description": "✈️ Airports"},
	}

	_, err := b.apiRequest("setMyCommands", map[string]any{
		"commands": commands,
	})

	return err
}

// Stop 停止机器人
func (b *TelegramBot) Stop() {
	b.mutex.Lock()
	if b.pollingActive {
		close(b.stopChan)
		b.pollingActive = false
	}
	b.connected = false
	b.mutex.Unlock()
}

// IsConnected 检查是否连接
func (b *TelegramBot) IsConnected() bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.connected
}

// GetLastError 获取最后的错误
func (b *TelegramBot) GetLastError() string {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.lastError
}

// setConnected 设置连接状态
func (b *TelegramBot) setConnected(connected bool) {
	b.mutex.Lock()
	b.connected = connected
	if connected {
		b.lastError = ""
	}
	b.mutex.Unlock()
}

// setError 设置错误
func (b *TelegramBot) setError(err string) {
	b.mutex.Lock()
	b.lastError = err
	b.connected = false
	b.mutex.Unlock()
}

// startPolling 启动长轮询
func (b *TelegramBot) startPolling() {
	b.mutex.Lock()
	b.pollingActive = true
	b.mutex.Unlock()

	utils.Info("Telegram long polling started")

	retryCount := 0
	maxRetry := 5

	for {
		select {
		case <-b.stopChan:
			utils.Info("Telegram long polling stopped")
			return
		default:
			updates, err := b.getUpdates()
			if err != nil {
				retryCount++
				b.setError(err.Error())
				utils.Warn("Failed to get updates (%d/%d): %v", retryCount, maxRetry, err)

				if retryCount >= maxRetry {
					utils.Warn("Too many Telegram connection failures; retrying in 30 seconds")
					time.Sleep(30 * time.Second)
					retryCount = 0
				} else {
					time.Sleep(time.Duration(retryCount) * time.Second)
				}
				continue
			}

			retryCount = 0
			b.setConnected(true)

			for _, update := range updates {
				go b.handleUpdate(update)
				b.updateOffset = update.UpdateID + 1
			}
		}
	}
}

// getUpdates 获取更新（长轮询）
func (b *TelegramBot) getUpdates() ([]Update, error) {
	params := map[string]any{
		"offset":  b.updateOffset,
		"timeout": 30,
	}

	resp, err := b.apiRequest("getUpdates", params)
	if err != nil {
		return nil, err
	}

	var result struct {
		OK          bool     `json:"ok"`
		Result      []Update `json:"result"`
		Description string   `json:"description"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse update: %v", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("failed to get updates: %s", result.Description)
	}

	return result.Result, nil
}

// handleUpdate 处理更新
func (b *TelegramBot) handleUpdate(update Update) {
	utils.Debug("[Telegram] Received update ID: %d", update.UpdateID)

	// 处理消息
	if update.Message != nil {
		utils.Debug("[Telegram] Message received - ChatID: %d, From: %s, Text: %s",
			update.Message.Chat.ID,
			update.Message.From.Username,
			update.Message.Text)
		b.handleMessage(update.Message)
		return
	}

	// 处理回调
	if update.CallbackQuery != nil {
		utils.Debug("[Telegram] Callback received - Data: %s", update.CallbackQuery.Data)
		b.handleCallback(update.CallbackQuery)
		return
	}
}

// handleMessage 处理消息
func (b *TelegramBot) handleMessage(message *Message) {
	utils.Debug("[Telegram] Handling message - ChatID: %d, configured ChatID: %d", message.Chat.ID, b.ChatID)

	// 验证 Chat ID（如果已配置）
	if b.ChatID != 0 && message.Chat.ID != b.ChatID {
		utils.Debug("[Telegram] Ignoring message from unauthorized chat: %d (expected: %d)", message.Chat.ID, b.ChatID)
		return
	}

	// 如果 Chat ID 未配置，自动绑定第一个发送 /start 的用户
	if b.ChatID == 0 && strings.HasPrefix(message.Text, "/start") {
		b.ChatID = message.Chat.ID
		_ = models.SetSetting("telegram_chat_id", strconv.FormatInt(message.Chat.ID, 10))
		utils.Info("[Telegram] Auto-bound Chat ID: %d", message.Chat.ID)
	}

	// 处理命令
	if message.Text != "" && strings.HasPrefix(message.Text, "/") {
		parts := strings.Fields(message.Text)
		command := strings.TrimPrefix(parts[0], "/")
		command = strings.Split(command, "@")[0] // 移除 @botname

		utils.Debug("[Telegram] Handling command: /%s", command)

		handler := GetHandler(command)
		if handler != nil {
			utils.Debug("[Telegram] Handler found: %s", handler.Description())
			if err := handler.Handle(b, message); err != nil {
				utils.Warn("[Telegram] Command /%s failed: %v", command, err)
				_ = b.SendMessage(message.Chat.ID, "❌ Command failed: "+err.Error(), "")
			} else {
				utils.Debug("[Telegram] Command /%s executed successfully", command)
			}
		} else {
			utils.Debug("[Telegram] No handler found for command: /%s", command)
			_ = b.SendMessage(message.Chat.ID, "❓ Unknown command; use /help for help", "")
		}
	}
}

// handleCallback 处理回调查询
func (b *TelegramBot) handleCallback(callback *CallbackQuery) {
	// 验证 Chat ID
	if b.ChatID != 0 && callback.Message.Chat.ID != b.ChatID {
		return
	}

	if err := HandleCallbackQuery(b, callback); err != nil {
		utils.Warn("Failed to handle callback: %v", err)
	}

	// 应答回调
	_ = b.answerCallback(callback.ID, "")
}

// apiRequest 发送 API 请求
func (b *TelegramBot) apiRequest(method string, params map[string]any) ([]byte, error) {
	url := TelegramAPIBase + b.Token + "/" + method

	var req *http.Request
	var err error

	if params != nil {
		jsonData, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize parameters: %v", err)
		}
		req, err = http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(context.Background(), "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	return body, nil
}

// SendMessage 发送消息
func (b *TelegramBot) SendMessage(chatID int64, text string, parseMode string) error {
	params := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}

	if parseMode != "" {
		params["parse_mode"] = parseMode
	}

	_, err := b.apiRequest("sendMessage", params)
	return err
}

// SendMessageWithKeyboard 发送带键盘的消息
func (b *TelegramBot) SendMessageWithKeyboard(chatID int64, text string, parseMode string, keyboard [][]InlineKeyboardButton) error {
	params := map[string]any{
		"chat_id": chatID,
		"text":    text,
		"reply_markup": map[string]any{
			"inline_keyboard": keyboard,
		},
	}

	if parseMode != "" {
		params["parse_mode"] = parseMode
	}

	_, err := b.apiRequest("sendMessage", params)
	return err
}

// EditMessage 编辑消息
func (b *TelegramBot) EditMessage(chatID int64, messageID int, text string, parseMode string, keyboard [][]InlineKeyboardButton) error {
	params := map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
	}

	if parseMode != "" {
		params["parse_mode"] = parseMode
	}

	if keyboard != nil {
		params["reply_markup"] = map[string]any{
			"inline_keyboard": keyboard,
		}
	}

	_, err := b.apiRequest("editMessageText", params)
	return err
}

// answerCallback 应答回调查询
func (b *TelegramBot) answerCallback(callbackID string, text string) error {
	params := map[string]any{
		"callback_query_id": callbackID,
	}
	if text != "" {
		params["text"] = text
	}

	_, err := b.apiRequest("answerCallbackQuery", params)
	return err
}

// GetStatus 获取机器人状态
func GetStatus() map[string]any {
	bot := GetBot()
	if bot == nil {
		return map[string]any{
			"enabled":     false,
			"connected":   false,
			"error":       "",
			"botUsername": "",
			"botId":       int64(0),
		}
	}

	bot.mutex.RLock()
	username := bot.botUsername
	botID := bot.botID
	bot.mutex.RUnlock()

	return map[string]any{
		"enabled":     true,
		"connected":   bot.IsConnected(),
		"error":       bot.GetLastError(),
		"botUsername": username,
		"botId":       botID,
	}
}

// Reconnect 重连机器人
func Reconnect() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	if !config.Enabled || config.BotToken == "" {
		return fmt.Errorf("bot is not enabled or configured")
	}

	return StartBot(config)
}

// CreateTestBot 创建临时测试机器人（不启动长轮询）
func CreateTestBot(config *Config) (*TelegramBot, error) {
	// 创建 HTTP 客户端（可能带代理）
	client, _, err := utils.CreateProxyHTTPClient(config.UseProxy, config.ProxyLink, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %v", err)
	}

	bot := &TelegramBot{
		Token:     config.BotToken,
		ChatID:    config.ChatID,
		UseProxy:  config.UseProxy,
		ProxyLink: config.ProxyLink,
		client:    client,
	}

	// 验证 Token
	if err := bot.validateToken(); err != nil {
		return nil, fmt.Errorf("failed to verify token: %v", err)
	}

	return bot, nil
}
