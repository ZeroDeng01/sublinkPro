package middlewares

import (
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// DemoModeRestrict 演示模式限制中间件
// 在演示模式下阻止敏感操作
func DemoModeRestrict(c *gin.Context) {
	if models.IsDemoMode() {
		utils.FailWithMsg(c, "演示模式下无法执行此操作")
		c.Abort()
		return
	}
	c.Next()
}
