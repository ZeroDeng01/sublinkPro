package api

import (
	"strings"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// currentUsernameFromContext 从认证中间件写入的 Gin 上下文中读取当前用户名。
// 上下文缺失或类型异常通常表示认证链路未正确通过，按未授权请求处理。
func currentUsernameFromContext(c *gin.Context) (string, bool) {
	usernameValue, exists := c.Get("username")
	if !exists {
		utils.ForbiddenI18n(c, "未获取到当前用户", "backend.auth.currentUser.missing", nil)
		return "", false
	}
	username, ok := usernameValue.(string)
	if !ok || strings.TrimSpace(username) == "" {
		utils.ForbiddenI18n(c, "当前用户身份无效", "backend.auth.currentUser.invalid", nil)
		return "", false
	}
	return username, true
}
