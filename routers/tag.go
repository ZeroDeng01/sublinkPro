package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Tag(r *gin.Engine) {
	tagGroup := r.Group("/api/v1/tags")
	tagGroup.Use(middlewares.AuthToken)
	{
		// 标签管理
		tagGroup.GET("/list", api.TagGet)
		tagGroup.GET("/groups", api.TagGroupList)
		tagGroup.POST("/add", middlewares.DemoModeRestrict, api.TagAdd)
		tagGroup.POST("/update", middlewares.DemoModeRestrict, api.TagUpdate)
		tagGroup.DELETE("/delete", middlewares.DemoModeRestrict, api.TagDelete)

		// 规则管理
		tagGroup.GET("/rules", api.TagRuleGet)
		tagGroup.POST("/rules/add", middlewares.DemoModeRestrict, api.TagRuleAdd)
		tagGroup.POST("/rules/update", middlewares.DemoModeRestrict, api.TagRuleUpdate)
		tagGroup.DELETE("/rules/delete", middlewares.DemoModeRestrict, api.TagRuleDelete)
		tagGroup.POST("/rules/trigger", middlewares.DemoModeRestrict, api.TagRuleTrigger)

		// 节点标签操作
		tagGroup.POST("/node/add", middlewares.DemoModeRestrict, api.NodeAddTag)
		tagGroup.POST("/node/remove", middlewares.DemoModeRestrict, api.NodeRemoveTag)
		tagGroup.POST("/node/batch-add", middlewares.DemoModeRestrict, api.NodeBatchAddTag)
		tagGroup.POST("/node/batch-set", middlewares.DemoModeRestrict, api.NodeBatchSetTags)
		tagGroup.POST("/node/batch-remove", middlewares.DemoModeRestrict, api.NodeBatchRemoveTags)
		tagGroup.GET("/node/tags", api.GetNodeTags)
	}
}
