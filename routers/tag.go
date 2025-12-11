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
		tagGroup.POST("/add", api.TagAdd)
		tagGroup.POST("/update", api.TagUpdate)
		tagGroup.DELETE("/delete", api.TagDelete)

		// 规则管理
		tagGroup.GET("/rules", api.TagRuleGet)
		tagGroup.POST("/rules/add", api.TagRuleAdd)
		tagGroup.POST("/rules/update", api.TagRuleUpdate)
		tagGroup.DELETE("/rules/delete", api.TagRuleDelete)
		tagGroup.POST("/rules/trigger", api.TagRuleTrigger)

		// 节点标签操作
		tagGroup.POST("/node/add", api.NodeAddTag)
		tagGroup.POST("/node/remove", api.NodeRemoveTag)
		tagGroup.POST("/node/batch-add", api.NodeBatchAddTag)
		tagGroup.POST("/node/batch-set", api.NodeBatchSetTags)
		tagGroup.POST("/node/batch-remove", api.NodeBatchRemoveTags)
		tagGroup.GET("/node/tags", api.GetNodeTags)
	}
}
