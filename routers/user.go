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
		authGroup.POST("/mfa/verify-totp", api.VerifyTOTPLogin)
		authGroup.POST("/mfa/verify-recovery-code", api.VerifyRecoveryCodeLogin)
		authGroup.POST("/mfa/reset", middlewares.DemoModeRestrict, api.ResetTOTP)
		authGroup.DELETE("/logout", api.UserOut)
		authGroup.GET("/captcha", api.GetCaptcha)
	}
	userGroup := r.Group("/api/v1/users")
	userGroup.Use(middlewares.AuthToken)
	{
		userGroup.GET("/me", api.UserMe)
		userGroup.GET("/ai-settings", api.UserGetAISettings)
		userGroup.POST("/ai-settings/models", middlewares.DemoModeRestrict, api.UserListAIModels)
		userGroup.GET("/mfa", api.GetMFAStatus)
		userGroup.POST("/ai-settings", middlewares.DemoModeRestrict, api.UserUpdateAISettings)
		userGroup.POST("/ai-settings/test", middlewares.DemoModeRestrict, api.UserTestAISettings)
		userGroup.POST("/mfa/reauth", middlewares.DemoModeRestrict, api.ReauthMFA)
		userGroup.POST("/mfa/totp/begin", middlewares.DemoModeRestrict, api.BeginTOTPEnrollment)
		userGroup.POST("/mfa/totp/confirm", middlewares.DemoModeRestrict, api.ConfirmTOTPEnrollment)
		userGroup.POST("/mfa/totp/disable", middlewares.DemoModeRestrict, api.DisableTOTP)
		userGroup.POST("/mfa/recovery-codes/regenerate", middlewares.DemoModeRestrict, api.RegenerateRecoveryCodes)
		userGroup.GET("/page", api.UserPages)
		userGroup.POST("/update", middlewares.DemoModeRestrict, api.UserSet)
		// 演示模式下禁止修改用户资料和密码
		userGroup.POST("/update-profile", middlewares.DemoModeRestrict, api.UserUpdateProfile)
		userGroup.POST("/change-password", middlewares.DemoModeRestrict, api.UserChangePassword)

	}
}
