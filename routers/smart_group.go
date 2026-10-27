package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func SmartGroup(r *gin.Engine) {
	group := r.Group("/api/v1/smart-groups")
	group.Use(middlewares.AuthToken)
	group.GET("", api.ListSmartGroups)
	group.POST("", middlewares.DemoModeRestrict, api.CreateSmartGroup)
	group.PUT("/:id", middlewares.DemoModeRestrict, api.UpdateSmartGroup)
	group.DELETE("/:id", middlewares.DemoModeRestrict, api.DeleteSmartGroup)
	group.GET("/:id/members", api.SmartGroupMembers)
	group.POST("/:id/check", middlewares.DemoModeRestrict, api.CheckSmartGroupCandidates)
}
