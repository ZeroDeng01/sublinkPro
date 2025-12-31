package routers

import (
	"sublink/config"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

func Version(r *gin.Engine, version string) {

	r.GET("/api/v1/version", func(c *gin.Context) {
		// 返回版本号和启用的功能列表
		utils.OkDetailed(c, "获取版本成功", gin.H{
			"version":  version,
			"features": config.GetEnabledFeatures(),
		})
	})
}
