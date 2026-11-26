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
		NodesGroup.POST("/add", middlewares.DemoModeRestrict, api.NodeAdd)
		NodesGroup.DELETE("/delete", middlewares.DemoModeRestrict, api.NodeDel)
		NodesGroup.DELETE("/batch-delete", middlewares.DemoModeRestrict, api.NodeBatchDel)
		NodesGroup.POST("/batch-update-group", middlewares.DemoModeRestrict, api.NodeBatchUpdateGroup)
		NodesGroup.POST("/batch-update-dialer-proxy", middlewares.DemoModeRestrict, api.NodeBatchUpdateDialerProxy)
		NodesGroup.POST("/batch-update-source", middlewares.DemoModeRestrict, api.NodeBatchUpdateSource)
		NodesGroup.POST("/batch-update-country", middlewares.DemoModeRestrict, api.NodeBatchUpdateCountry)
		NodesGroup.POST("/batch-fill-country", middlewares.DemoModeRestrict, api.NodeBatchFillCountry)
		NodesGroup.GET("/get", api.NodeGet)
		NodesGroup.GET("/selector", api.NodeSelector)
		NodesGroup.GET("/selector/by-ids", api.NodeSelectorByIDs)
		NodesGroup.GET("/ids", api.NodeGetIDs)
		NodesGroup.POST("/update", middlewares.DemoModeRestrict, api.NodeUpdadte)
		NodesGroup.GET("/groups", api.GetGroups)
		NodesGroup.GET("/group-stats", api.NodeGroupStats)
		NodesGroup.GET("/sources", api.GetSources)
		NodesGroup.GET("/countries", api.GetNodeCountries)
		NodesGroup.GET("/ip-info", middlewares.DemoModeRestrict, api.GetIPDetails)
		NodesGroup.GET("/ip-cache/stats", api.GetIPCacheStats)
		NodesGroup.DELETE("/ip-cache", middlewares.DemoModeRestrict, api.ClearIPCache)
		NodesGroup.GET("/protocols", api.GetNodeProtocols)
		// 节点原始信息相关
		NodesGroup.GET("/protocol-ui-meta", api.GetProtocolUIMeta)
		NodesGroup.GET("/parse-link", api.ParseNodeLinkAPI)
		NodesGroup.GET("/raw-info", api.GetNodeRawInfo)
		NodesGroup.POST("/update-raw", middlewares.DemoModeRestrict, api.UpdateNodeRawInfo)
	}

}
