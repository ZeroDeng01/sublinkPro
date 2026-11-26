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
		TempsGroup.POST("/add", middlewares.DemoModeRestrict, api.AddTemp)
		TempsGroup.POST("/delete", middlewares.DemoModeRestrict, api.DelTemp)
		TempsGroup.GET("/usage", api.GetTemplateUsage)
		TempsGroup.GET("/get", middlewares.DemoModeRestrict, api.GetTempS)
		TempsGroup.POST("/update", middlewares.DemoModeRestrict, api.UpdateTemp)
		TempsGroup.GET("/presets", api.GetACL4SSRPresets)
		TempsGroup.POST("/convert", middlewares.DemoModeRestrict, api.ConvertRules)
		TempsGroup.POST("/ai/edit-sessions/stream", middlewares.DemoModeRestrict, api.StartTemplateAIEditSessionStream)
		TempsGroup.GET("/ai/edit-sessions/:sessionId", middlewares.DemoModeRestrict, api.GetTemplateAIEditSession)
		TempsGroup.POST("/ai/edit-sessions/:sessionId/accept", middlewares.DemoModeRestrict, api.AcceptTemplateAIEditSession)
		TempsGroup.POST("/ai/edit-sessions/:sessionId/discard", middlewares.DemoModeRestrict, api.DiscardTemplateAIEditSession)
		TempsGroup.POST("/ai/generate", api.TemplateAILegacyRemoved)
		TempsGroup.POST("/ai/generate-stream", api.TemplateAILegacyRemoved)
		TempsGroup.POST("/ai/validate", api.TemplateAILegacyRemoved)
		TempsGroup.POST("/ai/apply", api.TemplateAILegacyRemoved)
	}

}
