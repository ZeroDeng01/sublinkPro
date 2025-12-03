package routers

import (
	"sublink/api"

	"github.com/gin-gonic/gin"
)

func Settings(r *gin.Engine) {
	SettingsGroup := r.Group("/api/v1/settings")
	{
		SettingsGroup.GET("/webhook", api.GetWebhookConfig)
		SettingsGroup.POST("/webhook", api.UpdateWebhookConfig)
		SettingsGroup.POST("/webhook/test", api.TestWebhookConfig)
	}
}
