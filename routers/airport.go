package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

// Airport 注册机场管理相关路由
func Airport(r *gin.Engine) {
	airportGroup := r.Group("/api/v1/airports")
	airportGroup.Use(middlewares.AuthToken)
	{
		// 列表和详情
		airportGroup.GET("", api.AirportList)
		airportGroup.GET("/:id", api.AirportGet)
		// 增删改（演示模式下限制）
		airportGroup.POST("", middlewares.DemoModeRestrict, api.AirportAdd)
		airportGroup.PUT("/:id", middlewares.DemoModeRestrict, api.AirportUpdate)
		airportGroup.DELETE("/:id", middlewares.DemoModeRestrict, api.AirportDelete)
		// 手动拉取
		airportGroup.POST("/:id/pull", middlewares.DemoModeRestrict, api.AirportPull)
		// 刷新用量信息
		airportGroup.POST("/:id/refresh-usage", middlewares.DemoModeRestrict, api.AirportRefreshUsage)
	}
}
