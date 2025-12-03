package routers

import (
	"sublink/api"

	"github.com/gin-gonic/gin"
)

func User(r *gin.Engine) {
	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/login", api.UserLogin)
		authGroup.DELETE("/logout", api.UserOut)
		authGroup.GET("/captcha", api.GetCaptcha)
	}
	userGroup := r.Group("/api/v1/users")
	{
		userGroup.GET("/me", api.UserMe)
		userGroup.GET("/page", api.UserPages)
		userGroup.POST("/update", api.UserSet)
		userGroup.POST("/update-profile", api.UserUpdateProfile)
		userGroup.POST("/change-password", api.UserChangePassword)

	}
}
