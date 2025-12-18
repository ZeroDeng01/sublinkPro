package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

// Host 注册 Host 相关路由
func Host(r *gin.Engine) {
	hostGroup := r.Group("/api/v1/hosts")
	hostGroup.Use(middlewares.AuthToken)
	{
		// Host 管理
		hostGroup.GET("/list", api.HostList)
		hostGroup.POST("/add", api.HostAdd)
		hostGroup.POST("/update", api.HostUpdate)
		hostGroup.DELETE("/delete", api.HostDelete)
		hostGroup.DELETE("/batch-delete", api.HostBatchDelete)

		// 文本模式
		hostGroup.GET("/export", api.HostExport)
		hostGroup.POST("/sync", api.HostSync)

		// 模块设置
		hostGroup.GET("/settings", api.GetHostSettings)
		hostGroup.POST("/settings", api.UpdateHostSettings)

		// Pin 固定
		hostGroup.POST("/pin", api.HostSetPinned)
	}
}
