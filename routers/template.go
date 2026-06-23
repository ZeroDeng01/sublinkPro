package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Templates(r *gin.Engine) {
	TempsGroup := r.Group("/api/v1/template")
	TempsGroup.Use(middlewares.AuthToken)
	{
		TempsGroup.POST("/add", api.AddTemp)
		TempsGroup.POST("/delete", api.DelTemp)
		TempsGroup.GET("/usage", api.GetTemplateUsage)
		TempsGroup.GET("/get", api.GetTempS)
		TempsGroup.POST("/update", api.UpdateTemp)
		TempsGroup.GET("/presets", api.GetACL4SSRPresets)
		TempsGroup.POST("/convert", api.ConvertRules)
		TempsGroup.POST("/ai/edit-sessions/stream", api.StartTemplateAIEditSessionStream)
		TempsGroup.GET("/ai/edit-sessions/:sessionId", api.GetTemplateAIEditSession)
		TempsGroup.POST("/ai/edit-sessions/:sessionId/accept", api.AcceptTemplateAIEditSession)
		TempsGroup.POST("/ai/edit-sessions/:sessionId/discard", api.DiscardTemplateAIEditSession)
		TempsGroup.POST("/ai/generate", api.TemplateAILegacyRemoved)
		TempsGroup.POST("/ai/generate-stream", api.TemplateAILegacyRemoved)
		TempsGroup.POST("/ai/validate", api.TemplateAILegacyRemoved)
		TempsGroup.POST("/ai/apply", api.TemplateAILegacyRemoved)
	}

}
