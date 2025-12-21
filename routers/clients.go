package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Clients(r *gin.Engine) {
	ClientsGroup := r.Group("/c")
	ClientsGroup.Use(middlewares.GetIp)
	{
		// ClientsGroup.GET("/v2ray/:subname", api.GetV2ray)
		// ClientsGroup.GET("/clash/:subname", api.GetClash)
		// ClientsGroup.GET("/surge/:subname", api.GetSurge)
		ClientsGroup.GET("/", api.GetClient)
		ClientsGroup.HEAD("/", api.GetClient)
	}

}
