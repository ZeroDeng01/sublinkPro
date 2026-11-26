package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func ProxyGateway(r *gin.Engine) {
	group := r.Group("/api/v1/proxy-gateway")
	group.Use(middlewares.AuthToken)
	{
		group.GET("/settings", api.GetSocks5Settings)
		group.POST("/settings", middlewares.DemoModeRestrict, api.UpdateSocks5Settings)
		group.POST("/start", middlewares.DemoModeRestrict, api.StartProxyGateway)
		group.POST("/stop", middlewares.DemoModeRestrict, api.StopSocks5)
		group.GET("/status", api.GetSocks5Status)
		group.GET("/node-stats", api.GetSocks5RoutingSnapshot)
		group.POST("/stats/reset", middlewares.DemoModeRestrict, api.ResetProxyGatewayStats)
		group.POST("/node-stats/reset", middlewares.DemoModeRestrict, api.ResetSocks5RuntimeStats)
		group.GET("/routing-profiles", api.ListSocks5RoutingProfiles)
		group.POST("/routing-profiles", middlewares.DemoModeRestrict, api.CreateSocks5RoutingProfile)
		group.PUT("/routing-profiles/:id", middlewares.DemoModeRestrict, api.UpdateSocks5RoutingProfile)
		group.DELETE("/routing-profiles/:id", middlewares.DemoModeRestrict, api.DeleteSocks5RoutingProfile)
		group.GET("/listeners", api.ListSocks5Listeners)
		group.POST("/listeners", middlewares.DemoModeRestrict, api.CreateSocks5Listener)
		group.PUT("/listeners/:id", middlewares.DemoModeRestrict, api.UpdateSocks5Listener)
		group.DELETE("/listeners/:id", middlewares.DemoModeRestrict, api.DeleteSocks5Listener)
		group.GET("/accounts", api.ListSocks5Accounts)
		group.POST("/accounts", middlewares.DemoModeRestrict, api.CreateSocks5Account)
		group.PUT("/accounts/:id", middlewares.DemoModeRestrict, api.UpdateSocks5Account)
		group.DELETE("/accounts/:id", middlewares.DemoModeRestrict, api.DeleteSocks5Account)
		group.GET("/health", api.GetProxyGatewayHealth)
		group.POST("/health/probe", middlewares.DemoModeRestrict, api.ProbeSocks5Health)
		group.GET("/connections", api.GetSocks5Connections)
		group.DELETE("/connections", middlewares.DemoModeRestrict, api.DeleteSocks5Connections)
		group.DELETE("/connections/:id", middlewares.DemoModeRestrict, api.DeleteSocks5Connection)
	}
}
