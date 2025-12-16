package api

import (
	"strconv"
	"sublink/models"
	"sublink/services/telegram"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// GetTelegramConfig 获取 Telegram 配置
func GetTelegramConfig(c *gin.Context) {
	config, err := telegram.LoadConfig()
	if err != nil {
		utils.FailWithMsg(c, "获取配置失败: "+err.Error())
		return
	}

	// 获取状态
	status := telegram.GetStatus()

	// 获取系统域名
	systemDomain, _ := models.GetSetting("system_domain")

	utils.OkDetailed(c, "获取成功", gin.H{
		"enabled":      config.Enabled,
		"botToken":     config.BotToken,
		"chatId":       config.ChatID,
		"useProxy":     config.UseProxy,
		"proxyLink":    config.ProxyLink,
		"systemDomain": systemDomain,
		"connected":    status["connected"],
		"lastError":    status["error"],
		"botUsername":  status["botUsername"],
		"botId":        status["botId"],
	})
}

// UpdateTelegramConfig 更新 Telegram 配置
func UpdateTelegramConfig(c *gin.Context) {
	var req struct {
		Enabled      bool   `json:"enabled"`
		BotToken     string `json:"botToken"`
		ChatId       int64  `json:"chatId"`
		UseProxy     bool   `json:"useProxy"`
		ProxyLink    string `json:"proxyLink"`
		SystemDomain string `json:"systemDomain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	// 保存系统域名
	if req.SystemDomain != "" {
		if err := models.SetSetting("system_domain", req.SystemDomain); err != nil {
			utils.Error("保存系统域名失败: %v", err)
		}
	}

	config := &telegram.Config{
		Enabled:   req.Enabled,
		BotToken:  req.BotToken,
		ChatID:    req.ChatId,
		UseProxy:  req.UseProxy,
		ProxyLink: req.ProxyLink,
	}

	// 保存配置
	if err := telegram.SaveConfig(config); err != nil {
		utils.FailWithMsg(c, "保存配置失败: "+err.Error())
		return
	}

	// 如果启用且有 Token，启动机器人
	if config.Enabled && config.BotToken != "" {
		telegram.StopBot()
		if err := telegram.StartBot(config); err != nil {
			utils.OkDetailed(c, "配置已保存，但启动机器人失败: "+err.Error(), gin.H{
				"saved":     true,
				"connected": false,
				"error":     err.Error(),
			})
			return
		}
	} else {
		telegram.StopBot()
	}

	utils.OkWithMsg(c, "保存成功")
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
		utils.FailWithMsg(c, "请提供 Bot Token")
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
		utils.FailWithMsg(c, "连接失败: "+err.Error())
		return
	}

	// 如果有 ChatID，发送测试消息
	if req.ChatId != 0 {
		if err := testBot.SendMessage(req.ChatId, "✅ Sublink Pro 连接测试成功！", ""); err != nil {
			utils.OkDetailed(c, "连接成功，但发送测试消息失败", gin.H{
				"connected":   true,
				"messageSent": false,
				"error":       err.Error(),
			})
			return
		}
		utils.OkDetailed(c, "连接成功，测试消息已发送", gin.H{
			"connected":   true,
			"messageSent": true,
		})
		return
	}

	utils.OkDetailed(c, "连接成功", gin.H{
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
		utils.FailWithMsg(c, "重连失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "重连成功")
}

// SendTelegramMessage 发送 Telegram 消息（用于测试）
func SendTelegramMessage(c *gin.Context) {
	var req struct {
		Message string `json:"message" binding:"required"`
		ChatId  string `json:"chatId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请提供消息内容")
		return
	}

	bot := telegram.GetBot()
	if bot == nil {
		utils.FailWithMsg(c, "机器人未启动")
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
		utils.FailWithMsg(c, "未配置 Chat ID")
		return
	}

	if err := bot.SendMessage(chatID, req.Message, "Markdown"); err != nil {
		utils.FailWithMsg(c, "发送失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "发送成功")
}
