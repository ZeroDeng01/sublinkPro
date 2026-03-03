package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func GroupSort(r *gin.Engine) {
	groupSortGroup := r.Group("/api/v1/group-sort")
	groupSortGroup.Use(middlewares.AuthToken)
	{
		groupSortGroup.GET("/groups", api.GroupSortGroups)
		groupSortGroup.GET("/detail", api.GroupSortDetail)
		groupSortGroup.POST("/save", api.GroupSortSave)
	}
}
