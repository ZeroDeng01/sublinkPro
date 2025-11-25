package routers

import (
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

func Version(r *gin.Engine, version string) {

	r.GET("/api/v1/version", func(c *gin.Context) {
		utils.OkDetailed(c, "获取版本成功", version)
	})
}
