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
		NodesGroup.POST("/add", api.SubSchedulerAdd)
		NodesGroup.DELETE("/delete/:id", api.SubSchedulerDel)
		NodesGroup.GET("/get", api.SubSchedulerGet)
		NodesGroup.PUT("/update", api.SubSchedulerUpdate)
		NodesGroup.POST("/pull", api.PullClashConfigFromURL)
	}
}
