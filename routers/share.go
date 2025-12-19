package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

// Share 注册分享管理路由
func Share(r *gin.Engine) {
	shareGroup := r.Group("/api/v1/shares")
	shareGroup.Use(middlewares.AuthToken)
	{
		shareGroup.GET("/get", api.ShareGet)               // 获取订阅的所有分享
		shareGroup.POST("/add", api.ShareAdd)              // 创建新分享
		shareGroup.POST("/update", api.ShareUpdate)        // 更新分享
		shareGroup.DELETE("/delete", api.ShareDelete)      // 删除分享
		shareGroup.POST("/refresh", api.ShareRefreshToken) // 刷新Token
		shareGroup.GET("/logs", api.ShareLogs)             // 获取分享访问日志
	}
}
