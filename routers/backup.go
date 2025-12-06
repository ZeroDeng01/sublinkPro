package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Backup(r *gin.Engine) {
	BackupGroup := r.Group("/api/v1/backup")
	BackupGroup.Use(middlewares.AuthToken)
	{
		BackupGroup.GET("/download", api.Backup)
	}

}
