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
		// 演示模式下禁止修改脚本
		ScriptGroup.POST("/add", middlewares.DemoModeRestrict, api.ScriptAdd)
		ScriptGroup.DELETE("/delete", middlewares.DemoModeRestrict, api.ScriptDel)
		ScriptGroup.POST("/update", middlewares.DemoModeRestrict, api.ScriptUpdate)
		ScriptGroup.GET("/list", api.ScriptList)
	}
}
