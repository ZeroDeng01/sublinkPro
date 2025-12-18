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
		utils.Error("添加Host过期清理任务失败 - Cron: %s, Error: %v", hostCleanupCron, err)
		return err
	}

	sm.jobs[JobIDHostCleanup] = entryID
	utils.Info("成功添加Host过期清理任务 - Cron: %s", hostCleanupCron)
	return nil
}

// ExecuteHostCleanupTask 执行 Host 过期清理任务
func ExecuteHostCleanupTask() {
	deleted, err := models.CleanExpiredHosts()
	if err != nil {
		utils.Error("Host过期清理任务执行失败: %v", err)
		return
	}
	if deleted > 0 {
		utils.Debug("Host过期清理完成，删除 %d 条记录", deleted)
	}
}
