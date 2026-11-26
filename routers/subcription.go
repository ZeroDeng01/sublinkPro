package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Subcription(r *gin.Engine) {
	SubcriptionGroup := r.Group("/api/v1/subcription")
	SubcriptionGroup.Use(middlewares.AuthToken)
	{
		SubcriptionGroup.POST("/add", middlewares.DemoModeRestrict, api.SubAdd)
		SubcriptionGroup.DELETE("/delete", middlewares.DemoModeRestrict, api.SubDel)
		SubcriptionGroup.GET("/get", api.SubGet)
		SubcriptionGroup.POST("/update", middlewares.DemoModeRestrict, api.SubUpdate)
		SubcriptionGroup.POST("/sort", middlewares.DemoModeRestrict, api.SubSort)
		SubcriptionGroup.POST("/batch-sort", middlewares.DemoModeRestrict, api.SubBatchSort) // 批量排序接口
		SubcriptionGroup.POST("/copy", middlewares.DemoModeRestrict, api.SubCopy)            // 复制订阅接口
		SubcriptionGroup.POST("/preview", api.PreviewSubscriptionNodes)                      // 节点预览接口
		SubcriptionGroup.GET("/protocol-meta", api.GetProtocolMeta)                          // 协议元数据接口
		SubcriptionGroup.GET("/node-fields-meta", api.GetNodeFieldsMeta)                     // 节点字段元数据接口

		// 链式代理规则相关接口
		SubcriptionGroup.GET("/:id/chain-rules", api.GetChainRules)                                                // 获取规则列表
		SubcriptionGroup.POST("/:id/chain-rules", middlewares.DemoModeRestrict, api.CreateChainRule)               // 创建规则
		SubcriptionGroup.PUT("/:id/chain-rules/sort", middlewares.DemoModeRestrict, api.SortChainRules)            // 批量排序（必须在 :ruleId 路由前定义）
		SubcriptionGroup.PUT("/:id/chain-rules/:ruleId", middlewares.DemoModeRestrict, api.UpdateChainRule)        // 更新规则
		SubcriptionGroup.DELETE("/:id/chain-rules/:ruleId", middlewares.DemoModeRestrict, api.DeleteChainRule)     // 删除规则
		SubcriptionGroup.PUT("/:id/chain-rules/:ruleId/toggle", middlewares.DemoModeRestrict, api.ToggleChainRule) // 切换启用状态
		SubcriptionGroup.GET("/:id/chain-options", api.GetChainOptions)                                            // 获取可用选项
		SubcriptionGroup.GET("/:id/chain-rules/preview", api.PreviewChainLinks)                                    // 预览链路（整体）
	}

}
