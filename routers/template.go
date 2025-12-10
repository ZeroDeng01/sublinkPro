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
		TempsGroup.GET("/get", api.GetTempS)
		TempsGroup.POST("/update", api.UpdateTemp)
		TempsGroup.GET("/presets", api.GetACL4SSRPresets)
		TempsGroup.POST("/convert", api.ConvertRules)
	}

}
