package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func AccessKey(r *gin.Engine) {
	accessKeyGroup := r.Group("/api/v1/accesskey")
	accessKeyGroup.Use(middlewares.AuthToken)
	{
		// 演示模式下禁止创建/删除 AccessKey
		accessKeyGroup.POST("/add", middlewares.DemoModeRestrict, api.GenerateAccessKey)
		accessKeyGroup.DELETE("/delete/:accessKeyId", middlewares.DemoModeRestrict, api.DeleteAccessKey)
		accessKeyGroup.GET("/get/:userId", api.GetAccessKey)
	}
}
