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
		geoipGroup.GET("/config", api.GetGeoIPConfig)
		geoipGroup.PUT("/config", api.SaveGeoIPConfig)
		geoipGroup.GET("/status", api.GetGeoIPStatus)
		geoipGroup.POST("/download", api.DownloadGeoIP)
		geoipGroup.POST("/stop", api.StopGeoIPDownload)
	}
}
