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
		SubcriptionGroup.POST("/add", api.SubAdd)
		SubcriptionGroup.DELETE("/delete", api.SubDel)
		SubcriptionGroup.GET("/get", api.SubGet)
		SubcriptionGroup.POST("/update", api.SubUpdate)
		SubcriptionGroup.POST("/sort", api.SubSort)
		SubcriptionGroup.POST("/preview", api.PreviewSubscriptionNodes)  // 节点预览接口
		SubcriptionGroup.GET("/protocol-meta", api.GetProtocolMeta)      // 协议元数据接口
		SubcriptionGroup.GET("/node-fields-meta", api.GetNodeFieldsMeta) // 节点字段元数据接口
	}

}
