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
		SettingsGroup.GET("/webhooks", api.ListWebhooks)
		SettingsGroup.POST("/webhooks", middlewares.DemoModeRestrict, api.CreateWebhook)
		SettingsGroup.PUT("/webhooks/:id", middlewares.DemoModeRestrict, api.UpdateWebhook)
		SettingsGroup.DELETE("/webhooks/:id", middlewares.DemoModeRestrict, api.DeleteWebhook)
		SettingsGroup.POST("/webhooks/:id/test", middlewares.DemoModeRestrict, api.TestWebhookByID)
		SettingsGroup.GET("/base-templates", api.GetBaseTemplates)
		SettingsGroup.POST("/base-templates", middlewares.DemoModeRestrict, api.UpdateBaseTemplate)

		// 系统域名配置
		SettingsGroup.GET("/system-domain", api.GetSystemDomain)
		SettingsGroup.POST("/system-domain", middlewares.DemoModeRestrict, api.UpdateSystemDomain)

		// Telegram 机器人设置
		SettingsGroup.GET("/telegram", api.GetTelegramConfig)
		SettingsGroup.POST("/telegram", middlewares.DemoModeRestrict, api.UpdateTelegramConfig)
		SettingsGroup.POST("/telegram/test", middlewares.DemoModeRestrict, api.TestTelegramConnection)
		SettingsGroup.GET("/telegram/status", api.GetTelegramStatus)
		SettingsGroup.POST("/telegram/reconnect", middlewares.DemoModeRestrict, api.ReconnectTelegram)

		// 节点去重配置
		SettingsGroup.GET("/node-dedup", api.GetNodeDedupConfig)
		SettingsGroup.POST("/node-dedup", middlewares.DemoModeRestrict, api.UpdateNodeDedupConfig)

		// 全局节点处理规则配置
		SettingsGroup.GET("/global-node-processing", api.GetGlobalNodeProcessingConfig)
		SettingsGroup.POST("/global-node-processing", middlewares.DemoModeRestrict, api.UpdateGlobalNodeProcessingConfig)

		// AI 助手设置
		SettingsGroup.GET("/ai-assistant", api.UserGetAISettings)
		SettingsGroup.POST("/ai-assistant/models", middlewares.DemoModeRestrict, api.UserListAIModels)
		SettingsGroup.POST("/ai-assistant", middlewares.DemoModeRestrict, api.UserUpdateAISettings)
		SettingsGroup.POST("/ai-assistant/test", middlewares.DemoModeRestrict, api.UserTestAISettings)

		// Cloudflare Tunnel 设置
		SettingsGroup.GET("/cloudflared", api.GetCloudflaredStatus)
		SettingsGroup.POST("/cloudflared", middlewares.DemoModeRestrict, api.UpdateCloudflaredConfig)
		SettingsGroup.POST("/cloudflared/start", middlewares.DemoModeRestrict, api.StartCloudflared)
		SettingsGroup.POST("/cloudflared/stop", middlewares.DemoModeRestrict, api.StopCloudflared)
		SettingsGroup.DELETE("/cloudflared/token", middlewares.DemoModeRestrict, api.RemoveCloudflaredToken)

		// Sub-Store sidecar 设置
		SettingsGroup.GET("/substore", api.GetSubStoreSettings)
		SettingsGroup.POST("/substore", middlewares.DemoModeRestrict, api.UpdateSubStoreSettings)
		SettingsGroup.POST("/substore/test", middlewares.DemoModeRestrict, api.TestSubStoreSettings)

		// 数据库迁移
		SettingsGroup.POST("/database-migration/import", middlewares.DemoModeRestrict, api.ImportDatabaseMigration)
	}
}
