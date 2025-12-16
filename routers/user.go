package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func User(r *gin.Engine) {
	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/login", api.UserLogin)
		authGroup.POST("/remember-login", api.RememberLogin) // 记住密码令牌登录
		authGroup.DELETE("/logout", api.UserOut)
		authGroup.GET("/captcha", api.GetCaptcha)
	}
	userGroup := r.Group("/api/v1/users")
	userGroup.Use(middlewares.AuthToken)
	{
		userGroup.GET("/me", api.UserMe)
		userGroup.GET("/page", api.UserPages)
		userGroup.POST("/update", middlewares.DemoModeRestrict, api.UserSet)
		// 演示模式下禁止修改用户资料和密码
		userGroup.POST("/update-profile", middlewares.DemoModeRestrict, api.UserUpdateProfile)
		userGroup.POST("/change-password", middlewares.DemoModeRestrict, api.UserChangePassword)

	}
}
