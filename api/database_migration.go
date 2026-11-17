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
	username, ok := currentUsernameFromContext(c)
	if !ok {
		return
	}

	currentUser := &models.User{Username: username}
	if err := currentUser.Find(); err != nil {
		utils.Forbidden(c, "Current user not found")
		return
	}
	if !strings.EqualFold(currentUser.Role, "admin") {
		utils.Forbidden(c, "Only administrators can run database migrations")
		return
	}

	uploadedFile, err := c.FormFile("file")
	if err != nil {
		utils.FailWithMsg(c, "Upload a SQLite database file or backup.zip")
		return
	}

	fileExt := strings.ToLower(filepath.Ext(uploadedFile.Filename))
	tempFile, err := createDatabaseMigrationUploadFile(fileExt)
	if err != nil {
		utils.FailWithMsg(c, "Failed to create migration temp file")
		return
	}
	tempPath := tempFile.Name()

	src, err := uploadedFile.Open()
	if err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		utils.FailWithMsg(c, "Failed to read uploaded file: "+err.Error())
		return
	}

	if _, err := io.Copy(tempFile, src); err != nil {
		_ = src.Close()
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		utils.FailWithMsg(c, "Failed to save migration file: "+err.Error())
		return
	}
	if err := src.Close(); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		utils.FailWithMsg(c, "Failed to close uploaded file: "+err.Error())
		return
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		utils.FailWithMsg(c, "Failed to save migration file: "+err.Error())
		return
	}

	options := services.DatabaseMigrationOptions{
		IncludeSubLogs:    parseFormBool(c.PostForm("includeSubLogs")),
		IncludeAccessKeys: true,
	}
	if raw := c.PostForm("includeAccessKeys"); raw != "" {
		options.IncludeAccessKeys = parseFormBool(raw)
	}

	taskName := fmt.Sprintf("Database migration: %s", uploadedFile.Filename)
	task, ctx, err := services.GetTaskManager().CreateTask(models.TaskTypeDatabaseMigration, taskName, models.TaskTriggerManual, 1)
	if err != nil {
		_ = os.Remove(tempPath)
		utils.FailWithMsg(c, "Failed to create migration task: "+err.Error())
		return
	}

	go services.RunDatabaseMigrationTask(ctx, task.ID, tempPath, uploadedFile.Filename, options)

	utils.OkDetailed(c, "Migration task started", gin.H{
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
