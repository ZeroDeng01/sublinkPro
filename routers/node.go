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
		NodesGroup.POST("/batch-update-source", api.NodeBatchUpdateSource)
		NodesGroup.GET("/get", api.NodeGet)
		NodesGroup.GET("/ids", api.NodeGetIDs)
		NodesGroup.POST("/update", api.NodeUpdadte)
		NodesGroup.POST("/update-link-name", api.UpdateNodeLinkNameAPI)
		NodesGroup.GET("/groups", api.GetGroups)
		NodesGroup.GET("/sources", api.GetSources)
		NodesGroup.GET("/countries", api.GetNodeCountries)
		NodesGroup.GET("/speed-test/config", api.GetSpeedTestConfig)
		// 演示模式下禁止修改测速配置和执行测速
		NodesGroup.POST("/speed-test/config", middlewares.DemoModeRestrict, api.UpdateSpeedTestConfig)
		NodesGroup.POST("/speed-test/run", middlewares.DemoModeRestrict, api.RunSpeedTest)
		NodesGroup.GET("/ip-info", api.GetIPDetails)
		NodesGroup.GET("/ip-cache/stats", api.GetIPCacheStats)
		NodesGroup.DELETE("/ip-cache", api.ClearIPCache)
	}

}
