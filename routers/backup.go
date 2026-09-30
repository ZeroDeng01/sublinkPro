package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Backup(r *gin.Engine) {
	backupGroup := r.Group("/api/v1/backup")
	backupGroup.Use(middlewares.AuthToken)
	{
		// 系统备份和 WebDAV 操作均包含敏感数据，演示模式下全部禁用。
		backupGroup.GET("/download", middlewares.DemoModeRestrict, api.Backup)
		backupGroup.GET("/webdav", middlewares.DemoModeRestrict, api.GetWebDAVBackupSettings)
		backupGroup.POST("/webdav", middlewares.DemoModeRestrict, api.UpdateWebDAVBackupSettings)
		backupGroup.POST("/webdav/test", middlewares.DemoModeRestrict, api.TestWebDAVBackup)
		backupGroup.POST("/webdav/upload", middlewares.DemoModeRestrict, api.UploadWebDAVBackup)
		backupGroup.GET("/webdav/files", middlewares.DemoModeRestrict, api.ListWebDAVBackups)
		backupGroup.POST("/webdav/restore", middlewares.DemoModeRestrict, api.RestoreWebDAVBackup)
	}
}
