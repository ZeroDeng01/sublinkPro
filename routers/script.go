package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Script(r *gin.Engine) {
	ScriptGroup := r.Group("/api/v1/script")
	ScriptGroup.Use(middlewares.AuthToken)
	{
		ScriptGroup.POST("/add", api.ScriptAdd)
		ScriptGroup.DELETE("/delete", api.ScriptDel)
		ScriptGroup.POST("/update", api.ScriptUpdate)
		ScriptGroup.GET("/list", api.ScriptList)
	}
}
