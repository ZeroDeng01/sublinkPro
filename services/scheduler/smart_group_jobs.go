package scheduler

import (
	"sublink/models"
	"sublink/utils"
	"time"

	"github.com/robfig/cron/v3"
)

// UpdateSmartGroupJob replaces a schedule only after a new valid entry is prepared.
func (sm *SchedulerManager) UpdateSmartGroupJob(group models.SmartGroup) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	if sm.smartGroupJobs == nil {
		sm.smartGroupJobs = make(map[int]cron.EntryID)
	}
	old, exists := sm.smartGroupJobs[group.ID]
	if !group.ScheduleEnabled || models.IsDemoMode() {
		if exists {
			sm.cron.Remove(old)
			delete(sm.smartGroupJobs, group.ID)
		}
		return nil
	}
	entry, err := sm.cron.AddFunc(group.CronExpr, func() { RunScheduledSmartGroupCheck(group.ID) })
	if err != nil {
		return err
	}
	if exists {
		sm.cron.Remove(old)
	}
	sm.smartGroupJobs[group.ID] = entry
	return nil
}

func (sm *SchedulerManager) RemoveSmartGroupJob(id int) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	if entry, exists := sm.smartGroupJobs[id]; exists {
		sm.cron.Remove(entry)
		delete(sm.smartGroupJobs, id)
	}
}

func (sm *SchedulerManager) LoadSmartGroupJobs() error {
	groups, err := models.ListSmartGroups()
	if err != nil {
		return err
	}
	for _, group := range groups {
		if !group.ScheduleEnabled {
			continue
		}
		if err := ValidateSmartGroupCheckProfile(group, group.CheckProfileID); err != nil {
			utils.Warn("智能分组 [%s] 自动检测策略无效: %v", group.Name, err)
			continue
		}
		if err := sm.UpdateSmartGroupJob(group); err != nil {
			utils.Warn("注册智能分组 [%s] 定时检测失败: %v", group.Name, err)
		}
	}
	return nil
}

func (sm *SchedulerManager) SmartGroupCheckInfo(group models.SmartGroup) models.SmartGroup {
	group.CheckRunning = SmartGroupCheckRunning(group.ID)
	group.CheckTimezone = time.Now().Format("MST -07:00")
	group.NextCheckAt = nil
	sm.mutex.RLock()
	entryID, exists := sm.smartGroupJobs[group.ID]
	sm.mutex.RUnlock()
	if exists && group.ScheduleEnabled {
		next := sm.cron.Entry(entryID).Next
		if !next.IsZero() {
			group.NextCheckAt = &next
		} else {
			group.NextCheckAt = sm.getNextRunTime(group.CronExpr)
		}
	}
	return group
}
