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
		// 演示模式下禁止修改系统设置
		SettingsGroup.POST("/webhook", middlewares.DemoModeRestrict, api.UpdateWebhookConfig)
		SettingsGroup.POST("/webhook/test", middlewares.DemoModeRestrict, api.TestWebhookConfig)
		SettingsGroup.GET("/base-templates", api.GetBaseTemplates)
		SettingsGroup.POST("/base-templates", middlewares.DemoModeRestrict, api.UpdateBaseTemplate)

		// Telegram 机器人设置
		SettingsGroup.GET("/telegram", api.GetTelegramConfig)
		SettingsGroup.POST("/telegram", middlewares.DemoModeRestrict, api.UpdateTelegramConfig)
		SettingsGroup.POST("/telegram/test", middlewares.DemoModeRestrict, api.TestTelegramConnection)
		SettingsGroup.GET("/telegram/status", api.GetTelegramStatus)
		SettingsGroup.POST("/telegram/reconnect", middlewares.DemoModeRestrict, api.ReconnectTelegram)
	}
}
