package scheduler

import (
	"regexp"
	"strings"
	"sublink/models"
	"sublink/utils"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// SchedulerManager 定时任务管理器
type SchedulerManager struct {
	cron  *cron.Cron
	jobs  map[int]cron.EntryID // 存储任务ID和cron EntryID的映射
	mutex sync.RWMutex
}

// 全局定时任务管理器实例
var globalScheduler *SchedulerManager
var once sync.Once

// GetSchedulerManager 获取全局定时任务管理器实例（单例模式）
func GetSchedulerManager() *SchedulerManager {
	once.Do(func() {
		globalScheduler = &SchedulerManager{
			cron: cron.New(cron.WithParser(cron.NewParser(
				cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
			))),
			jobs: make(map[int]cron.EntryID),
		}
	})
	return globalScheduler
}

// Start 启动定时任务管理器
func (sm *SchedulerManager) Start() {
	sm.cron.Start()
	utils.Info("定时任务管理器已启动")
}

// Stop 停止定时任务管理器
func (sm *SchedulerManager) Stop() {
	sm.cron.Stop()
	utils.Info("定时任务管理器已停止")
}

// LoadFromDatabase 从数据库加载所有启用的定时任务
func (sm *SchedulerManager) LoadFromDatabase() error {

	airports, err := models.ListEnabledAirports()
	if err != nil {
		utils.Error("从数据库加载定时任务失败: %v", err)
		return err
	}
	// 添加所有启用的任务
	for _, airport := range airports {
		err := sm.AddJob(airport.ID, airport.CronExpr, func(id int, url string, subName string) {
			ExecuteSubscriptionTask(id, url, subName)
		}, airport.ID, airport.URL, airport.Name)

		if err != nil {
			utils.Error("添加定时任务失败 - ID: %d, Error: %v", airport.ID, err)
		} else {
			utils.Info("成功添加定时任务 - ID: %d, Name: %s, Cron: %s",
				airport.ID, airport.Name, airport.CronExpr)
		}
	}

	// 加载节点检测策略定时任务
	profiles, err := models.ListEnabledNodeCheckProfiles()
	if err != nil {
		utils.Error("从数据库加载节点检测策略定时任务失败: %v", err)
	} else {
		for _, profile := range profiles {
			if profile.CronExpr == "" {
				continue
			}
			if err := sm.AddNodeCheckProfileJob(profile.ID, profile.CronExpr); err != nil {
				utils.Error("添加节点检测定时任务失败 - ID: %d, Error: %v", profile.ID, err)
			} else {
				utils.Info("成功添加节点检测定时任务 - ID: %d, Name: %s, Cron: %s",
					profile.ID, profile.Name, profile.CronExpr)
			}
		}
	}

	// 启动 Host 过期清理任务
	if err := sm.StartHostCleanupTask(); err != nil {
		utils.Error("创建Host过期清理任务失败: %v", err)
	}

	return nil
}

// AddJob 添加定时任务
func (sm *SchedulerManager) AddJob(schedulerID int, cronExpr string, jobFunc func(int, string, string), id int, url string, subName string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 清理Cron表达式
	cleanCronExpr := cleanCronExpression(cronExpr)

	// 如果任务已存在，先删除
	if entryID, exists := sm.jobs[schedulerID]; exists {
		sm.cron.Remove(entryID)
		delete(sm.jobs, schedulerID)
	}

	// 添加新任务
	entryID, err := sm.cron.AddFunc(cleanCronExpr, func() {
		// 记录开始执行时间
		startTime := time.Now()

		// 执行业务逻辑
		jobFunc(id, url, subName)

		// 计算下次运行时间
		nextTime := sm.getNextRunTime(cleanCronExpr)

		// 更新数据库中的运行时间
		sm.updateRunTime(schedulerID, &startTime, nextTime)
	})

	if err != nil {
		utils.Error("添加定时任务失败 - ID: %d, Cron: %s, Error: %v", schedulerID, cleanCronExpr, err)
		return err
	}

	// 存储任务映射
	sm.jobs[schedulerID] = entryID

	// 计算并设置下次运行时间
	nextTime := sm.getNextRunTime(cleanCronExpr)
	sm.updateRunTime(schedulerID, nil, nextTime)

	utils.Info("成功添加定时任务 - ID: %d, Cron: %s, 下次运行: %v", schedulerID, cleanCronExpr, nextTime)

	return nil
}

// RemoveJob 删除定时任务
func (sm *SchedulerManager) RemoveJob(schedulerID int) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if entryID, exists := sm.jobs[schedulerID]; exists {
		sm.cron.Remove(entryID)
		delete(sm.jobs, schedulerID)
		utils.Info("成功删除定时任务 - ID: %d", schedulerID)
	}
}

// UpdateJob 更新定时任务
func (sm *SchedulerManager) UpdateJob(schedulerID int, cronExpr string, enabled bool, url string, subName string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 清理Cron表达式，去除多余空格
	cleanCronExpr := cleanCronExpression(cronExpr)

	// 先删除旧任务
	if entryID, exists := sm.jobs[schedulerID]; exists {
		sm.cron.Remove(entryID)
		delete(sm.jobs, schedulerID)
	}

	// 如果启用，添加新任务
	if enabled {
		entryID, err := sm.cron.AddFunc(cleanCronExpr, func() {
			// 记录开始执行时间
			startTime := time.Now()

			ExecuteSubscriptionTask(schedulerID, url, subName)

			// 计算下次运行时间
			nextTime := sm.getNextRunTime(cleanCronExpr)

			// 更新数据库中的运行时间
			sm.updateRunTime(schedulerID, &startTime, nextTime)
		})

		if err != nil {
			utils.Error("更新定时任务失败 - ID: %d, Cron: %s, Error: %v", schedulerID, cleanCronExpr, err)
			return err
		}

		sm.jobs[schedulerID] = entryID

		// 计算并设置下次运行时间
		nextTime := sm.getNextRunTime(cleanCronExpr)
		sm.updateRunTime(schedulerID, nil, nextTime)

		utils.Info("成功更新定时任务 - ID: %d, Cron: %s, 下次运行: %v", schedulerID, cleanCronExpr, nextTime)
	} else {
		// 如果禁用，清除下次运行时间
		sm.updateRunTime(schedulerID, nil, nil)
		utils.Info("任务已禁用 - ID: %d", schedulerID)
	}

	return nil
}

// cleanCronExpression 清理Cron表达式中的多余空格
func cleanCronExpression(cronExpr string) string {
	// 去除首尾空格
	cleaned := strings.TrimSpace(cronExpr)
	// 使用正则表达式将多个连续空格替换为单个空格
	re := regexp.MustCompile(`\s+`)
	cleaned = re.ReplaceAllString(cleaned, " ")
	return cleaned
}

// getNextRunTime 计算下次运行时间
func (sm *SchedulerManager) getNextRunTime(cronExpr string) *time.Time {
	// 清理Cron表达式
	cleanCronExpr := cleanCronExpression(cronExpr)

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cleanCronExpr)
	if err != nil {
		utils.Warn("解析Cron表达式失败: %s, Error: %v", cleanCronExpr, err)
		return nil
	}

	nextTime := schedule.Next(time.Now())
	return &nextTime
}

// updateRunTime 更新数据库中的运行时间
func (sm *SchedulerManager) updateRunTime(schedulerID int, lastRun, nextRun *time.Time) {
	go func() {
		airport, err := models.GetAirportByID(schedulerID)
		if err != nil {
			utils.Error("获取机场失败 - ID: %d, Error: %v", schedulerID, err)
			return
		}

		err = airport.UpdateRunTime(lastRun, nextRun)
		if err != nil {
			utils.Error("更新运行时间失败 - ID: %d, Error: %v", schedulerID, err)
		}
	}()
}

// ============================================================
// 节点检测策略任务管理
// ============================================================

// nodeCheckJobIDOffset 节点检测策略任务ID偏移量，用于区分机场任务
const nodeCheckJobIDOffset = 1000000

// AddNodeCheckProfileJob 添加节点检测策略定时任务
func (sm *SchedulerManager) AddNodeCheckProfileJob(profileID int, cronExpr string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 使用偏移量区分机场任务和节点检测任务
	jobID := nodeCheckJobIDOffset + profileID

	// 清理Cron表达式
	cleanCronExpr := cleanCronExpression(cronExpr)
	if cleanCronExpr == "" {
		return nil
	}

	// 如果任务已存在，先删除
	if entryID, exists := sm.jobs[jobID]; exists {
		sm.cron.Remove(entryID)
		delete(sm.jobs, jobID)
	}

	// 添加新任务
	entryID, err := sm.cron.AddFunc(cleanCronExpr, func() {
		// 记录开始执行时间
		startTime := time.Now()

		// 执行节点检测
		ExecuteNodeCheckWithProfile(profileID, nil)

		// 计算下次运行时间
		nextTime := sm.getNextRunTime(cleanCronExpr)

		// 更新数据库中的运行时间
		sm.updateNodeCheckProfileRunTime(profileID, &startTime, nextTime)
	})

	if err != nil {
		utils.Error("添加节点检测定时任务失败 - ProfileID: %d, Cron: %s, Error: %v", profileID, cleanCronExpr, err)
		return err
	}

	// 存储任务映射
	sm.jobs[jobID] = entryID

	// 计算并设置下次运行时间
	nextTime := sm.getNextRunTime(cleanCronExpr)
	sm.updateNodeCheckProfileRunTime(profileID, nil, nextTime)

	utils.Info("成功添加节点检测定时任务 - ProfileID: %d, Cron: %s, 下次运行: %v", profileID, cleanCronExpr, nextTime)

	return nil
}

// RemoveNodeCheckProfileJob 删除节点检测策略定时任务
func (sm *SchedulerManager) RemoveNodeCheckProfileJob(profileID int) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	jobID := nodeCheckJobIDOffset + profileID

	if entryID, exists := sm.jobs[jobID]; exists {
		sm.cron.Remove(entryID)
		delete(sm.jobs, jobID)
		utils.Info("成功删除节点检测定时任务 - ProfileID: %d", profileID)
	}
}

// UpdateNodeCheckProfileJob 更新节点检测策略定时任务
func (sm *SchedulerManager) UpdateNodeCheckProfileJob(profileID int, cronExpr string, enabled bool) error {
	// 先删除旧任务
	sm.RemoveNodeCheckProfileJob(profileID)

	// 如果启用且有Cron表达式，添加新任务
	if enabled && cronExpr != "" {
		return sm.AddNodeCheckProfileJob(profileID, cronExpr)
	}

	// 如果禁用，清除下次运行时间
	if !enabled {
		sm.updateNodeCheckProfileRunTime(profileID, nil, nil)
	}

	return nil
}

// updateNodeCheckProfileRunTime 更新节点检测策略的运行时间
func (sm *SchedulerManager) updateNodeCheckProfileRunTime(profileID int, lastRun, nextRun *time.Time) {
	go func() {
		profile, err := models.GetNodeCheckProfileByID(profileID)
		if err != nil {
			utils.Error("获取节点检测策略失败 - ID: %d, Error: %v", profileID, err)
			return
		}

		err = profile.UpdateRunTime(lastRun, nextRun)
		if err != nil {
			utils.Error("更新节点检测策略运行时间失败 - ID: %d, Error: %v", profileID, err)
		}
	}()
}

