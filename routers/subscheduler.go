package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func SubScheduler(r *gin.Engine) {
	NodesGroup := r.Group("/api/v1/sub_scheduler")
	NodesGroup.Use(middlewares.AuthToken)
	{
		// 演示模式下禁止修改订阅调度
		NodesGroup.POST("/add", middlewares.DemoModeRestrict, api.SubSchedulerAdd)
		NodesGroup.DELETE("/delete/:id", middlewares.DemoModeRestrict, api.SubSchedulerDel)
		NodesGroup.GET("/get", api.SubSchedulerGet)
		NodesGroup.PUT("/update", middlewares.DemoModeRestrict, api.SubSchedulerUpdate)
		NodesGroup.POST("/pull", middlewares.DemoModeRestrict, api.PullClashConfigFromURL)
	}
}
