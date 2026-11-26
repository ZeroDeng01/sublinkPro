package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

// GeoIP 注册 GeoIP 相关路由
func GeoIP(r *gin.Engine) {
	geoipGroup := r.Group("/api/v1/geoip")
	geoipGroup.Use(middlewares.AuthToken)
	{
		geoipGroup.GET("/config", middlewares.DemoModeRestrict, api.GetGeoIPConfig)
		geoipGroup.PUT("/config", middlewares.DemoModeRestrict, api.SaveGeoIPConfig)
		geoipGroup.GET("/status", api.GetGeoIPStatus)
		geoipGroup.POST("/download", middlewares.DemoModeRestrict, api.DownloadGeoIP)
		geoipGroup.POST("/stop", middlewares.DemoModeRestrict, api.StopGeoIPDownload)
	}
}
