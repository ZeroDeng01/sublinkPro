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
		accessKeyGroup.POST("/add", api.GenerateAccessKey)
		accessKeyGroup.DELETE("/delete/:accessKeyId", api.DeleteAccessKey)
		accessKeyGroup.GET("/get/:userId", api.GetAccessKey)
	}
}
