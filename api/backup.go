package api

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"sublink/database"
	"sublink/models"
	"sublink/services"
	backupservice "sublink/services/backup"
	"sublink/services/scheduler"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

type webDAVConfigRequest struct {
	BaseURL             string `json:"baseUrl"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	ClearPassword       bool   `json:"clearPassword"`
	RemotePath          string `json:"remotePath"`
	TimeoutSeconds      int    `json:"timeoutSeconds"`
	AllowInsecureHTTP   bool   `json:"allowInsecureHttp"`
	AllowPrivateNetwork bool   `json:"allowPrivateNetwork"`
	ScheduleEnabled     bool   `json:"scheduleEnabled"`
	CronExpr            string `json:"cronExpr"`
}

type webDAVRestoreRequest struct {
	Filename          string `json:"filename"`
	IncludeSubLogs    bool   `json:"includeSubLogs"`
	IncludeAccessKeys *bool  `json:"includeAccessKeys"`
}

func Backup(c *gin.Context) {
	if !requireBackupAdmin(c) {
		return
	}
	archive, err := backupservice.CreateArchiveFile(c.Request.Context())
	if err != nil {
		utils.FailWithI18n(c, "Failed to create backup: "+err.Error(), "settings.backup.api.archiveFailed", map[string]any{"message": err.Error()})
		return
	}
	defer func() { _ = os.Remove(archive.Path) }()

	file, err := os.Open(archive.Path)
	if err != nil {
		utils.FailWithI18n(c, "Failed to open backup file", "settings.backup.api.archiveFailed", map[string]any{"message": err.Error()})
		return
	}
	defer func() { _ = file.Close() }()

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", `attachment; filename="`+archive.Name+`"`)
	c.Header("Content-Length", fmt.Sprintf("%d", archive.Size))
	if _, err := io.Copy(c.Writer, file); err != nil {
		return
	}
}

func GetWebDAVBackupSettings(c *gin.Context) {
	if !requireBackupAdmin(c) {
		return
	}
	cfg, err := backupservice.LoadConfig()
	if err != nil {
		utils.FailWithI18n(c, "Failed to load WebDAV settings: "+err.Error(), "settings.backup.api.loadFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "WebDAV settings loaded", backupservice.ToPublicConfig(cfg), "settings.backup.api.loaded", nil)
}

func UpdateWebDAVBackupSettings(c *gin.Context) {
	if !requireBackupAdmin(c) {
		return
	}
	var req webDAVConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithI18n(c, "Invalid parameters", "settings.backup.api.invalidRequest", nil)
		return
	}
	cfg, err := backupservice.ResolveConfig(backupservice.ConfigUpdate(req))
	if err != nil {
		utils.FailWithI18n(c, "Invalid WebDAV settings: "+err.Error(), "settings.backup.api.invalidSettings", map[string]any{"message": err.Error()})
		return
	}
	if err := backupservice.SaveConfig(cfg); err != nil {
		utils.FailWithI18n(c, "Failed to save WebDAV settings: "+err.Error(), "settings.backup.api.saveFailed", map[string]any{"message": err.Error()})
		return
	}
	if err := scheduler.GetSchedulerManager().UpdateWebDAVBackupJob(cfg.CronExpr, cfg.ScheduleEnabled && cfg.BaseURL != ""); err != nil {
		utils.FailWithI18n(c, "Failed to save WebDAV settings: "+err.Error(), "settings.backup.api.saveFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "WebDAV settings saved", backupservice.ToPublicConfig(cfg), "settings.backup.api.saved", nil)
}

func TestWebDAVBackup(c *gin.Context) {
	if !requireBackupAdmin(c) {
		return
	}
	cfg, err := webDAVConfigFromRequest(c)
	if err != nil {
		utils.FailWithI18n(c, "Invalid WebDAV settings: "+err.Error(), "settings.backup.api.invalidSettings", map[string]any{"message": err.Error()})
		return
	}
	client, err := backupservice.NewClient(cfg)
	if err != nil {
		utils.FailWithI18n(c, "Invalid WebDAV settings: "+err.Error(), "settings.backup.api.invalidSettings", map[string]any{"message": err.Error()})
		return
	}
	started := time.Now()
	if err := client.Test(c.Request.Context()); err != nil {
		utils.FailWithI18n(c, "WebDAV connection test failed: "+err.Error(), "settings.backup.api.testFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "WebDAV connection test succeeded", gin.H{"latencyMs": time.Since(started).Milliseconds()}, "settings.backup.api.testSucceeded", nil)
}

func UploadWebDAVBackup(c *gin.Context) {
	if !requireBackupAdmin(c) {
		return
	}
	if !database.IsSQLite() {
		utils.FailWithI18n(c, "WebDAV system backup currently supports SQLite only", "settings.backup.api.sqliteOnly", nil)
		return
	}
	cfg, err := backupservice.LoadConfig()
	if err != nil {
		utils.FailWithI18n(c, "Failed to load WebDAV settings: "+err.Error(), "settings.backup.api.loadFailed", map[string]any{"message": err.Error()})
		return
	}
	remote, err := backupservice.CreateAndUpload(c.Request.Context(), cfg)
	if err != nil {
		utils.FailWithI18n(c, "Failed to upload WebDAV backup: "+err.Error(), "settings.backup.api.uploadFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "WebDAV backup uploaded", remote, "settings.backup.api.uploadSucceeded", map[string]any{"name": remote.Name})
}

func ListWebDAVBackups(c *gin.Context) {
	if !requireBackupAdmin(c) {
		return
	}
	client, err := loadWebDAVClient()
	if err != nil {
		utils.FailWithI18n(c, "WebDAV is not configured: "+err.Error(), "settings.backup.api.notConfigured", map[string]any{"message": err.Error()})
		return
	}
	files, err := client.List(c.Request.Context())
	if err != nil {
		utils.FailWithI18n(c, "Failed to load WebDAV backup list: "+err.Error(), "settings.backup.api.listFailed", map[string]any{"message": err.Error()})
		return
	}
	sort.Slice(files, func(i, j int) bool { return files[i].ModifiedAt.After(files[j].ModifiedAt) })
	if len(files) > 200 {
		files = files[:200]
	}
	utils.OkDetailedI18n(c, "WebDAV backup list loaded", files, "settings.backup.api.listed", nil)
}

func RestoreWebDAVBackup(c *gin.Context) {
	if !requireBackupAdmin(c) {
		return
	}
	var req webDAVRestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Filename) == "" {
		utils.FailWithI18n(c, "Invalid parameters", "settings.backup.api.invalidRequest", nil)
		return
	}
	client, err := loadWebDAVClient()
	if err != nil {
		utils.FailWithI18n(c, "WebDAV is not configured: "+err.Error(), "settings.backup.api.notConfigured", map[string]any{"message": err.Error()})
		return
	}
	tempFile, err := createDatabaseMigrationUploadFile(".zip")
	if err != nil {
		utils.FailWithI18n(c, "Failed to create temporary restore file", "settings.backup.api.restoreFailed", map[string]any{"message": err.Error()})
		return
	}
	tempPath := tempFile.Name()
	cleanup := func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}
	if _, err := client.Download(c.Request.Context(), strings.TrimSpace(req.Filename), tempFile); err != nil {
		cleanup()
		utils.FailWithI18n(c, "Failed to download WebDAV backup: "+err.Error(), "settings.backup.api.restoreFailed", map[string]any{"message": err.Error()})
		return
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		utils.FailWithI18n(c, "Failed to save WebDAV backup", "settings.backup.api.restoreFailed", map[string]any{"message": err.Error()})
		return
	}

	includeAccessKeys := true
	if req.IncludeAccessKeys != nil {
		includeAccessKeys = *req.IncludeAccessKeys
	}
	options := services.DatabaseMigrationOptions{IncludeSubLogs: req.IncludeSubLogs, IncludeAccessKeys: includeAccessKeys}
	task, ctx, err := services.GetTaskManager().CreateTask(models.TaskTypeDatabaseMigration, "WebDAV restore: "+req.Filename, models.TaskTriggerManual, 1)
	if err != nil {
		_ = os.Remove(tempPath)
		utils.FailWithI18n(c, "Failed to create restore task: "+err.Error(), "settings.backup.api.restoreFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "WebDAV restore task started", gin.H{"taskId": task.ID}, "settings.backup.api.restoreStarted", nil)
	go services.RunDatabaseMigrationTask(ctx, task.ID, tempPath, req.Filename, options)
}

func webDAVConfigFromRequest(c *gin.Context) (backupservice.Config, error) {
	var req webDAVConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return backupservice.Config{}, err
	}
	return backupservice.ResolveConfig(backupservice.ConfigUpdate(req))
}

func loadWebDAVClient() (*backupservice.Client, error) {
	cfg, err := backupservice.LoadConfig()
	if err != nil {
		return nil, err
	}
	return backupservice.NewClient(cfg)
}

func requireBackupAdmin(c *gin.Context) bool {
	username, ok := currentUsernameFromContext(c)
	if !ok {
		return false
	}
	currentUser := &models.User{Username: username}
	if err := currentUser.Find(); err != nil || !strings.EqualFold(currentUser.Role, "admin") {
		utils.ForbiddenI18n(c, "Only administrators can manage system backups", "settings.backup.api.adminRequired", nil)
		return false
	}
	return true
}
