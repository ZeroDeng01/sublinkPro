package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Total(r *gin.Engine) {
	TotalGroup := r.Group("/api/v1/total")
	TotalGroup.Use(middlewares.AuthToken)
	{
		TotalGroup.GET("/sub", api.SubTotal)
		TotalGroup.GET("/node", api.NodesTotal)
	}

}
