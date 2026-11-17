package api

import (
	"strconv"
	"sublink/services/notifications"
	"sublink/services/telegram"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// GetTelegramConfig 获取 Telegram 配置
func GetTelegramConfig(c *gin.Context) {
	config, err := telegram.LoadConfig()
	if err != nil {
		utils.FailWithMsg(c, "Failed to fetch configuration: "+err.Error())
		return
	}

	// 获取状态
	status := telegram.GetStatus()

	utils.OkDetailed(c, "Fetched successfully", gin.H{
		"enabled":      config.Enabled,
		"botToken":     config.BotToken,
		"chatId":       config.ChatID,
		"useProxy":     config.UseProxy,
		"proxyLink":    config.ProxyLink,
		"eventKeys":    config.EventKeys,
		"eventOptions": notifications.EventCatalogForChannel(notifications.ChannelTelegram),
		"connected":    status["connected"],
		"lastError":    status["error"],
		"botUsername":  status["botUsername"],
		"botId":        status["botId"],
	})
}

// UpdateTelegramConfig 更新 Telegram 配置
func UpdateTelegramConfig(c *gin.Context) {
	var req struct {
		Enabled   bool     `json:"enabled"`
		BotToken  string   `json:"botToken"`
		ChatId    int64    `json:"chatId"`
		UseProxy  bool     `json:"useProxy"`
		ProxyLink string   `json:"proxyLink"`
		EventKeys []string `json:"eventKeys"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}

	config := &telegram.Config{
		Enabled:   req.Enabled,
		BotToken:  req.BotToken,
		ChatID:    req.ChatId,
		UseProxy:  req.UseProxy,
		ProxyLink: req.ProxyLink,
		EventKeys: req.EventKeys,
	}

	// 保存配置
	if err := telegram.SaveConfig(config); err != nil {
		utils.FailWithMsg(c, "Failed to save configuration: "+err.Error())
		return
	}

	// 如果启用且有 Token，启动机器人
	if config.Enabled && config.BotToken != "" {
		telegram.StopBot()
		if err := telegram.StartBot(config); err != nil {
			utils.OkDetailed(c, "Configuration saved, but starting the bot failed: "+err.Error(), gin.H{
				"saved":     true,
				"connected": false,
				"error":     err.Error(),
			})
			return
		}
	} else {
		telegram.StopBot()
	}

	utils.OkWithMsg(c, "Saved successfully")
}

// TestTelegramConnection 测试 Telegram 连接
func TestTelegramConnection(c *gin.Context) {
	var req struct {
		BotToken  string `json:"botToken" binding:"required"`
		UseProxy  bool   `json:"useProxy"`
		ProxyLink string `json:"proxyLink"`
		ChatId    int64  `json:"chatId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Provide a Bot Token")
		return
	}

	config := &telegram.Config{
		Enabled:   true,
		BotToken:  req.BotToken,
		ChatID:    req.ChatId,
		UseProxy:  req.UseProxy,
		ProxyLink: req.ProxyLink,
	}

	// 临时启动测试
	testBot, err := telegram.CreateTestBot(config)
	if err != nil {
		utils.FailWithMsg(c, "Connection failed: "+err.Error())
		return
	}

	// 如果有 ChatID，发送测试消息
	if req.ChatId != 0 {
		if err := testBot.SendMessage(req.ChatId, "✅ Sublink Pro connection test succeeded!", ""); err != nil {
			utils.OkDetailed(c, "Connected, but failed to send the test message", gin.H{
				"connected":   true,
				"messageSent": false,
				"error":       err.Error(),
			})
			return
		}
		utils.OkDetailed(c, "Connected; test message sent", gin.H{
			"connected":   true,
			"messageSent": true,
		})
		return
	}

	utils.OkDetailed(c, "Connected successfully", gin.H{
		"connected":   true,
		"messageSent": false,
	})
}

// GetTelegramStatus 获取 Telegram 连接状态
func GetTelegramStatus(c *gin.Context) {
	status := telegram.GetStatus()
	utils.OkWithData(c, status)
}

// ReconnectTelegram 重新连接 Telegram
func ReconnectTelegram(c *gin.Context) {
	if err := telegram.Reconnect(); err != nil {
		utils.FailWithMsg(c, "Reconnect failed: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "Reconnected successfully")
}

// SendTelegramMessage 发送 Telegram 消息（用于测试）
func SendTelegramMessage(c *gin.Context) {
	var req struct {
		Message string `json:"message" binding:"required"`
		ChatId  string `json:"chatId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Provide message content")
		return
	}

	bot := telegram.GetBot()
	if bot == nil {
		utils.FailWithMsg(c, "Bot is not running")
		return
	}

	chatID := bot.ChatID
	if req.ChatId != "" {
		parsed, err := strconv.ParseInt(req.ChatId, 10, 64)
		if err == nil {
			chatID = parsed
		}
	}

	if chatID == 0 {
		utils.FailWithMsg(c, "Chat ID is not configured")
		return
	}

	if err := bot.SendMessage(chatID, req.Message, "Markdown"); err != nil {
		utils.FailWithMsg(c, "Send failed: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "Sent successfully")
}
