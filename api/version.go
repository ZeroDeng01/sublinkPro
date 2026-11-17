package api

import (
	"sublink/config"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// GetVersion 返回当前版本号和启用的功能列表。
func GetVersion(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.OkDetailed(c, "Version fetched", gin.H{
			"version":  version,
			"features": config.GetEnabledFeatures(),
		})
	}
}
