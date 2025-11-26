package routers

import (
	"sublink/api"

	"github.com/gin-gonic/gin"
)

func Script(r *gin.Engine) {
	ScriptGroup := r.Group("/api/v1/script")
	{
		ScriptGroup.POST("/add", api.ScriptAdd)
		ScriptGroup.DELETE("/delete", api.ScriptDel)
		ScriptGroup.POST("/update", api.ScriptUpdate)
		ScriptGroup.GET("/list", api.ScriptList)
	}
}
