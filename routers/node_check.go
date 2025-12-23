package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

// NodeCheck 注册节点检测策略相关路由
func NodeCheck(r *gin.Engine) {
	group := r.Group("/api/v1/node-check")
	group.Use(middlewares.AuthToken)
	{
		// 策略管理
		group.GET("/profiles", api.ListNodeCheckProfiles)
		group.GET("/profiles/:id", api.GetNodeCheckProfile)
		group.POST("/profiles", middlewares.DemoModeRestrict, api.CreateNodeCheckProfile)
		group.PUT("/profiles/:id", middlewares.DemoModeRestrict, api.UpdateNodeCheckProfile)
		group.DELETE("/profiles/:id", middlewares.DemoModeRestrict, api.DeleteNodeCheckProfile)
		group.POST("/profiles/:id/run", middlewares.DemoModeRestrict, api.RunNodeCheckWithProfile)

		// 执行检测
		group.POST("/run", middlewares.DemoModeRestrict, api.RunNodeCheck)
	}
}
