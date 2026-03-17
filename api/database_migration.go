package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"sublink/config"
	"sublink/models"
	"sublink/services"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

func ImportDatabaseMigration(c *gin.Context) {
	usernameValue, exists := c.Get("username")
	if !exists {
		utils.Forbidden(c, "未获取到当前用户")
		return
	}

	currentUser := &models.User{Username: usernameValue.(string)}
	if err := currentUser.Find(); err != nil {
		utils.Forbidden(c, "当前用户不存在")
		return
	}
	if !strings.EqualFold(currentUser.Role, "admin") {
		utils.Forbidden(c, "仅管理员可执行数据库迁移")
		return
	}

	uploadedFile, err := c.FormFile("file")
	if err != nil {
		utils.FailWithMsg(c, "请上传 SQLite 数据库文件或 backup.zip")
		return
	}

	fileExt := strings.ToLower(filepath.Ext(uploadedFile.Filename))
	tempFile, err := createDatabaseMigrationUploadFile(fileExt)
	if err != nil {
		utils.FailWithMsg(c, "创建迁移临时文件失败")
		return
	}
	tempPath := tempFile.Name()

	src, err := uploadedFile.Open()
	if err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		utils.FailWithMsg(c, "读取上传文件失败: "+err.Error())
		return
	}

	if _, err := io.Copy(tempFile, src); err != nil {
		_ = src.Close()
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		utils.FailWithMsg(c, "保存迁移文件失败: "+err.Error())
		return
	}
	if err := src.Close(); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		utils.FailWithMsg(c, "关闭上传文件失败: "+err.Error())
		return
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		utils.FailWithMsg(c, "保存迁移文件失败: "+err.Error())
		return
	}

	options := services.DatabaseMigrationOptions{
		IncludeSubLogs:    parseFormBool(c.PostForm("includeSubLogs")),
		IncludeAccessKeys: true,
	}
	if raw := c.PostForm("includeAccessKeys"); raw != "" {
		options.IncludeAccessKeys = parseFormBool(raw)
	}

	taskName := fmt.Sprintf("数据库迁移: %s", uploadedFile.Filename)
	task, ctx, err := services.GetTaskManager().CreateTask(models.TaskTypeDatabaseMigration, taskName, models.TaskTriggerManual, 1)
	if err != nil {
		_ = os.Remove(tempPath)
		utils.FailWithMsg(c, "创建迁移任务失败: "+err.Error())
		return
	}

	go services.RunDatabaseMigrationTask(ctx, task.ID, tempPath, uploadedFile.Filename, options)

	utils.OkDetailed(c, "迁移任务已启动", gin.H{
		"taskId": task.ID,
	})
}

func parseFormBool(raw string) bool {
	raw = strings.TrimSpace(strings.ToLower(raw))
	return raw == "1" || raw == "true" || raw == "yes" || raw == "on"
}

func createDatabaseMigrationUploadFile(fileExt string) (*os.File, error) {
	tempRoot := filepath.Join(config.GetDBPath(), ".tmp", "database-migration")
	if err := os.MkdirAll(tempRoot, 0755); err != nil {
		return nil, err
	}
	return os.CreateTemp(tempRoot, "upload-*"+fileExt)
}
