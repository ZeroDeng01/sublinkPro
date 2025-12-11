package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Total(r *gin.Engine) {
	TotalGroup := r.Group("/api/v1/total")
	TotalGroup.Use(middlewares.AuthToken)
	{
		TotalGroup.GET("/sub", api.SubTotal)
		TotalGroup.GET("/node", api.NodesTotal)
		TotalGroup.GET("/fastest-speed", api.FastestSpeedNode)
		TotalGroup.GET("/lowest-delay", api.LowestDelayNode)
		TotalGroup.GET("/country-stats", api.NodeCountryStats)
		TotalGroup.GET("/protocol-stats", api.NodeProtocolStats)
		TotalGroup.GET("/system-stats", api.GetSystemStats)
		TotalGroup.GET("/tag-stats", api.NodeTagStats)
		TotalGroup.GET("/group-stats", api.NodeGroupStats)
		TotalGroup.GET("/source-stats", api.NodeSourceStats)
	}

}
