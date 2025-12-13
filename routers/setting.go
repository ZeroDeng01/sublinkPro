package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Settings(r *gin.Engine) {
	SettingsGroup := r.Group("/api/v1/settings")
	SettingsGroup.Use(middlewares.AuthToken)
	{
		SettingsGroup.GET("/webhook", api.GetWebhookConfig)
		SettingsGroup.POST("/webhook", api.UpdateWebhookConfig)
		SettingsGroup.POST("/webhook/test", api.TestWebhookConfig)
		SettingsGroup.GET("/base-templates", api.GetBaseTemplates)
		SettingsGroup.POST("/base-templates", api.UpdateBaseTemplate)

		// Telegram 机器人设置
		SettingsGroup.GET("/telegram", api.GetTelegramConfig)
		SettingsGroup.POST("/telegram", api.UpdateTelegramConfig)
		SettingsGroup.POST("/telegram/test", api.TestTelegramConnection)
		SettingsGroup.GET("/telegram/status", api.GetTelegramStatus)
		SettingsGroup.POST("/telegram/reconnect", api.ReconnectTelegram)
	}
}
