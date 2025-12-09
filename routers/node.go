package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Nodes(r *gin.Engine) {
	NodesGroup := r.Group("/api/v1/nodes")
	NodesGroup.Use(middlewares.AuthToken)
	{
		NodesGroup.POST("/add", api.NodeAdd)
		NodesGroup.DELETE("/delete", api.NodeDel)
		NodesGroup.DELETE("/batch-delete", api.NodeBatchDel)
		NodesGroup.POST("/batch-update-group", api.NodeBatchUpdateGroup)
		NodesGroup.POST("/batch-update-dialer-proxy", api.NodeBatchUpdateDialerProxy)
		NodesGroup.GET("/get", api.NodeGet)
		NodesGroup.POST("/update", api.NodeUpdadte)
		NodesGroup.GET("/groups", api.GetGroups)
		NodesGroup.GET("/sources", api.GetSources)
		NodesGroup.GET("/countries", api.GetNodeCountries)
		NodesGroup.GET("/speed-test/config", api.GetSpeedTestConfig)
		NodesGroup.POST("/speed-test/config", api.UpdateSpeedTestConfig)
		NodesGroup.POST("/speed-test/run", api.RunSpeedTest)
	}

}
