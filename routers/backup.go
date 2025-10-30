package routers

import (
	"sublink/api"

	"github.com/gin-gonic/gin"
)

func Backup(r *gin.Engine) {
	BackupGroup := r.Group("/api/v1/backup")
	{
		BackupGroup.GET("/download", api.Backup)
	}

}
