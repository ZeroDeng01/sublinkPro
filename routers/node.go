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
		NodesGroup.GET("/groups", api.GetGroups)
		NodesGroup.GET("/sources", api.GetSources)
		NodesGroup.GET("/countries", api.GetNodeCountries)
		NodesGroup.GET("/ip-info", api.GetIPDetails)
		NodesGroup.GET("/ip-cache/stats", api.GetIPCacheStats)
		NodesGroup.DELETE("/ip-cache", api.ClearIPCache)
		NodesGroup.GET("/protocols", api.GetNodeProtocols)
		// 节点原始信息相关
		NodesGroup.GET("/protocol-ui-meta", api.GetProtocolUIMeta)
		NodesGroup.GET("/parse-link", api.ParseNodeLinkAPI)
		NodesGroup.GET("/raw-info", api.GetNodeRawInfo)
		NodesGroup.POST("/update-raw", middlewares.DemoModeRestrict, api.UpdateNodeRawInfo)
	}

}
