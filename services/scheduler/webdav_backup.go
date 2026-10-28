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
		utils.Info("WebDAV 定时备份任务已停用")
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
		utils.Error("添加 WebDAV 定时备份任务失败 - Cron: %s, Error: %v", cleanCronExpr, err)
		return err
	}

	sm.jobs[JobIDWebDAVBackup] = entryID
	utils.Info("成功添加 WebDAV 定时备份任务 - Cron: %s", cleanCronExpr)
	return nil
}

// ExecuteWebDAVBackupTask creates a system backup and uploads it to WebDAV.
func ExecuteWebDAVBackupTask(trigger models.TaskTrigger) {
	if !webdavBackupMu.TryLock() {
		utils.Warn("WebDAV 定时备份仍在运行，跳过本次调度")
		return
	}
	defer webdavBackupMu.Unlock()

	if !database.IsSQLite() {
		utils.Warn("WebDAV 定时备份已跳过：当前数据库不是 SQLite")
		return
	}

	cfg, err := backupservice.LoadConfig()
	if err != nil {
		utils.Error("读取 WebDAV 备份配置失败: %v", err)
		return
	}
	if !cfg.ScheduleEnabled || cfg.BaseURL == "" {
		return
	}

	tm := getTaskManager()
	task, ctx, err := tm.CreateTask(models.TaskTypeWebDAVBackup, "WebDAV 定时备份", trigger, 2)
	if err != nil {
		utils.Error("创建 WebDAV 定时备份任务失败: %v", err)
		return
	}

	_ = tm.UpdateProgress(task.ID, 1, "正在创建并上传备份", nil)
	remote, err := backupservice.CreateAndUpload(ctx, cfg)
	if err != nil {
		utils.Error("WebDAV 定时备份失败: %v", err)
		_ = tm.FailTask(task.ID, err.Error())
		return
	}

	_ = tm.CompleteTask(task.ID, "备份已上传: "+remote.Name, remote)
	utils.Info("WebDAV 定时备份完成: %s", remote.Name)
}
