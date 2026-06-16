package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func CountryRule(r *gin.Engine) {
	countryRuleGroup := r.Group("/api/v1/country-rules")
	countryRuleGroup.Use(middlewares.AuthToken)
	{
		// 查询
		countryRuleGroup.GET("", api.ListCountryRules)

		// 修改操作
		countryRuleGroup.POST("", middlewares.DemoModeRestrict, api.CreateCountryRule)
		countryRuleGroup.PUT("/:id", middlewares.DemoModeRestrict, api.UpdateCountryRule)
		countryRuleGroup.DELETE("/:id", middlewares.DemoModeRestrict, api.DeleteCountryRule)

		// 测试
		countryRuleGroup.POST("/test", api.TestCountryRule)
		countryRuleGroup.POST("/batch-test", api.BatchTestCountryRules)

		// 批量操作
		countryRuleGroup.POST("/batch", middlewares.DemoModeRestrict, api.BatchCountryRules)

		// 文本模式
		countryRuleGroup.GET("/export", api.ExportCountryRules)
		countryRuleGroup.POST("/sync", middlewares.DemoModeRestrict, api.SyncCountryRules)
	}
}
