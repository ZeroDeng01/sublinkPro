package scheduler

import (
	"sublink/models"
	"sublink/utils"
)

// StartHostCleanupTask 启动 Host 过期清理定时任务
// 每10分钟执行一次，静默清理过期的 Host
func (sm *SchedulerManager) StartHostCleanupTask() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	const hostCleanupCron = "*/10 * * * *" // 每10分钟执行一次

	// 如果任务已存在，先删除
	if entryID, exists := sm.jobs[JobIDHostCleanup]; exists {
		sm.cron.Remove(entryID)
		delete(sm.jobs, JobIDHostCleanup)
	}

	// 添加清理任务
	entryID, err := sm.cron.AddFunc(hostCleanupCron, func() {
		ExecuteHostCleanupTask()
	})

	if err != nil {
		utils.Error("Failed to add Host expiry cleanup task - Cron: %s, Error: %v", hostCleanupCron, err)
		return err
	}

	sm.jobs[JobIDHostCleanup] = entryID
	utils.Info("Added Host expiry cleanup task - Cron: %s", hostCleanupCron)
	return nil
}

// ExecuteHostCleanupTask 执行 Host 过期清理任务
func ExecuteHostCleanupTask() {
	deleted, err := models.CleanExpiredHosts()
	if err != nil {
		utils.Error("Host expiry cleanup task failed: %v", err)
		return
	}
	if deleted > 0 {
		utils.Debug("Host expiry cleanup finished; %d records deleted", deleted)
	}
}
