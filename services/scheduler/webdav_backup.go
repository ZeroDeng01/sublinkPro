package scheduler

import (
	"sync"

	"sublink/database"
	"sublink/models"
	backupservice "sublink/services/backup"
	"sublink/utils"
)

var webdavBackupMu sync.Mutex

// StartWebDAVBackupTask loads persisted WebDAV settings and registers the scheduled backup job.
func (sm *SchedulerManager) StartWebDAVBackupTask() error {
	cfg, err := backupservice.LoadConfig()
	if err != nil {
		return err
	}
	return sm.UpdateWebDAVBackupJob(cfg.CronExpr, cfg.ScheduleEnabled && cfg.BaseURL != "")
}

// UpdateWebDAVBackupJob replaces the in-memory WebDAV backup cron job.
func (sm *SchedulerManager) UpdateWebDAVBackupJob(cronExpr string, enabled bool) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if entryID, exists := sm.jobs[JobIDWebDAVBackup]; exists {
		sm.cron.Remove(entryID)
		delete(sm.jobs, JobIDWebDAVBackup)
	}

	if !enabled {
		utils.Info("WebDAV scheduled backup is disabled")
		return nil
	}

	cleanCronExpr := cleanCronExpression(cronExpr)
	if cleanCronExpr == "" {
		return nil
	}

	entryID, err := sm.cron.AddFunc(cleanCronExpr, func() {
		ExecuteWebDAVBackupTask(models.TaskTriggerScheduled)
	})
	if err != nil {
		utils.Error("Failed to add WebDAV scheduled backup - Cron: %s, Error: %v", cleanCronExpr, err)
		return err
	}

	sm.jobs[JobIDWebDAVBackup] = entryID
	utils.Info("Added WebDAV scheduled backup - Cron: %s", cleanCronExpr)
	return nil
}

// ExecuteWebDAVBackupTask creates a system backup and uploads it to WebDAV.
func ExecuteWebDAVBackupTask(trigger models.TaskTrigger) {
	if !webdavBackupMu.TryLock() {
		utils.Warn("WebDAV scheduled backup is still running; skipping this schedule")
		return
	}
	defer webdavBackupMu.Unlock()

	if !database.IsSQLite() {
		utils.Warn("WebDAV scheduled backup skipped: the database is not SQLite")
		return
	}

	cfg, err := backupservice.LoadConfig()
	if err != nil {
		utils.Error("Failed to load WebDAV backup config: %v", err)
		return
	}
	if !cfg.ScheduleEnabled || cfg.BaseURL == "" {
		return
	}

	tm := getTaskManager()
	task, ctx, err := tm.CreateTask(models.TaskTypeWebDAVBackup, "WebDAV scheduled backup", trigger, 2)
	if err != nil {
		utils.Error("Failed to create WebDAV scheduled backup task: %v", err)
		return
	}

	_ = tm.UpdateProgress(task.ID, 1, "Creating and uploading backup", nil)
	remote, err := backupservice.CreateAndUpload(ctx, cfg)
	if err != nil {
		utils.Error("WebDAV scheduled backup failed: %v", err)
		_ = tm.FailTask(task.ID, err.Error())
		return
	}

	_ = tm.CompleteTask(task.ID, "Backup uploaded: "+remote.Name, remote)
	utils.Info("WebDAV scheduled backup completed: %s", remote.Name)
}
