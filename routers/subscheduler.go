package routers

import (
	"sublink/api"

	"github.com/gin-gonic/gin"
)

func SubScheduler(r *gin.Engine) {
	NodesGroup := r.Group("/api/v1/sub_scheduler")
	{
		NodesGroup.POST("/add", api.SubSchedulerAdd)
		NodesGroup.DELETE("/delete/:id", api.SubSchedulerDel)
		NodesGroup.GET("/get", api.SubSchedulerGet)
		NodesGroup.PUT("/update", api.SubSchedulerUpdate)
		NodesGroup.POST("/pull", api.PullClashConfigFromURL)
	}
}
