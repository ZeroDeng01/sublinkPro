package services

import (
	"log"
	"regexp"
	"strings"
	"sublink/models"
	"sublink/node"
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
	log.Println("定时任务管理器已启动")
}

// Stop 停止定时任务管理器
func (sm *SchedulerManager) Stop() {
	sm.cron.Stop()
	log.Println("定时任务管理器已停止")
}

// LoadFromDatabase 从数据库加载所有启用的定时任务
func (sm *SchedulerManager) LoadFromDatabase() error {

	schedulers, err := models.ListEnabled()
	if err != nil {
		log.Printf("从数据库加载定时任务失败: %v", err)
		return err
	}
	// 添加所有启用的任务
	for _, scheduler := range schedulers {
		err := sm.AddJob(scheduler.ID, scheduler.CronExpr, func(id int, url string, subName string) {
			ExecuteSubscriptionTask(id, url, subName)
		}, scheduler.ID, scheduler.URL, scheduler.Name)

		if err != nil {
			log.Printf("添加定时任务失败 - ID: %d, Error: %v", scheduler.ID, err)
		} else {
			log.Printf("成功添加定时任务 - ID: %d, Name: %s, Cron: %s",
				scheduler.ID, scheduler.Name, scheduler.CronExpr)
		}
	}

	speedTestEnable, err := models.GetSetting("speed_test_enabled")
	if err != nil {
		log.Printf("从数据库加载测速定时任务失败: %v", err)
		return err
	}
	if speedTestEnable == "true" {
		speedTestCron, err := models.GetSetting("speed_test_cron")
		if err != nil {
			log.Printf("从数据库加载测速定时任务失败: %v", err)
		}
		err = sm.StartNodeSpeedTestTask(speedTestCron)
		if err != nil {
			log.Printf("创建测速定时任务失败: %v", err)
		}

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
		log.Printf("添加定时任务失败 - ID: %d, Cron: %s, Error: %v", schedulerID, cleanCronExpr, err)
		return err
	}

	// 存储任务映射
	sm.jobs[schedulerID] = entryID

	// 计算并设置下次运行时间
	nextTime := sm.getNextRunTime(cleanCronExpr)
	sm.updateRunTime(schedulerID, nil, nextTime)

	log.Printf("成功添加定时任务 - ID: %d, Cron: %s, 下次运行: %v", schedulerID, cleanCronExpr, nextTime)

	return nil
}

// RemoveJob 删除定时任务
func (sm *SchedulerManager) RemoveJob(schedulerID int) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if entryID, exists := sm.jobs[schedulerID]; exists {
		sm.cron.Remove(entryID)
		delete(sm.jobs, schedulerID)
		log.Printf("成功删除定时任务 - ID: %d", schedulerID)
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
			log.Printf("更新定时任务失败 - ID: %d, Cron: %s, Error: %v", schedulerID, cleanCronExpr, err)
			return err
		}

		sm.jobs[schedulerID] = entryID

		// 计算并设置下次运行时间
		nextTime := sm.getNextRunTime(cleanCronExpr)
		sm.updateRunTime(schedulerID, nil, nextTime)

		log.Printf("成功更新定时任务 - ID: %d, Cron: %s, 下次运行: %v", schedulerID, cleanCronExpr, nextTime)
	} else {
		// 如果禁用，清除下次运行时间
		sm.updateRunTime(schedulerID, nil, nil)
		log.Printf("任务已禁用 - ID: %d", schedulerID)
	}

	return nil
}

// ExecuteSubscriptionTask 执行订阅任务的具体业务逻辑
func ExecuteSubscriptionTask(id int, url string, subName string) {

	log.Printf("执行自动获取订阅任务 - ID: %d, Name: %s, URL: %s", id, subName, url)
	node.LoadClashConfigFromURL(id, url, subName)
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
		log.Printf("解析Cron表达式失败: %s, Error: %v", cleanCronExpr, err)
		return nil
	}

	nextTime := schedule.Next(time.Now())
	return &nextTime
}

// updateRunTime 更新数据库中的运行时间
func (sm *SchedulerManager) updateRunTime(schedulerID int, lastRun, nextRun *time.Time) {
	go func() {
		var subS models.SubScheduler
		err := subS.GetByID(schedulerID)
		if err != nil {
			log.Printf("获取订阅调度失败 - ID: %d, Error: %v", schedulerID, err)
			return
		}

		err = subS.UpdateRunTime(lastRun, nextRun)
		if err != nil {
			log.Printf("更新运行时间失败 - ID: %d, Error: %v", schedulerID, err)
		}
	}()
}

// StartNodeSpeedTestTask 启动节点测速定时任务
func (sm *SchedulerManager) StartNodeSpeedTestTask(cronExpr string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 使用 -1 作为测速任务的 ID
	const speedTestTaskID = -100

	// 清理Cron表达式
	cleanCronExpr := cleanCronExpression(cronExpr)

	// 如果任务已存在，先删除
	if entryID, exists := sm.jobs[speedTestTaskID]; exists {
		sm.cron.Remove(entryID)
		delete(sm.jobs, speedTestTaskID)
	}

	// 添加新任务
	entryID, err := sm.cron.AddFunc(cleanCronExpr, func() {
		ExecuteNodeSpeedTestTask()
	})

	if err != nil {
		log.Printf("添加节点测速任务失败 - Cron: %s, Error: %v", cleanCronExpr, err)
		return err
	}

	// 存储任务映射
	sm.jobs[speedTestTaskID] = entryID
	log.Printf("成功添加节点测速任务 - speedTestTaskID: %d", speedTestTaskID)
	log.Printf("成功添加节点测速任务 - Cron: %s", cleanCronExpr)
	return nil
}

// StopNodeSpeedTestTask 停止节点测速定时任务
func (sm *SchedulerManager) StopNodeSpeedTestTask() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	const speedTestTaskID = -1
	if entryID, exists := sm.jobs[speedTestTaskID]; exists {
		sm.cron.Remove(entryID)
		delete(sm.jobs, speedTestTaskID)
		log.Println("成功停止节点测速任务")
	}
}

// ExecuteNodeSpeedTestTask 执行节点测速任务
func ExecuteNodeSpeedTestTask() {
	log.Println("开始执行节点测速任务...")

	// 获取测速分组配置
	speedTestGroupsStr, _ := models.GetSetting("speed_test_groups")
	var nodes []models.Node
	var err error

	if speedTestGroupsStr != "" {
		groups := strings.Split(speedTestGroupsStr, ",")
		nodes, err = new(models.Node).ListByGroups(groups)
		log.Printf("根据分组测速: %v", groups)
	} else {
		nodes, err = new(models.Node).List()
		log.Println("全量测速")
	}

	if err != nil {
		log.Printf("获取节点列表失败: %v", err)
		return
	}

	// 获取测速配置
	speedTestMode, _ := models.GetSetting("speed_test_mode")
	speedTestURL, _ := models.GetSetting("speed_test_url")
	speedTestTimeoutStr, _ := models.GetSetting("speed_test_timeout")

	timeout := 5 * time.Second
	if speedTestTimeoutStr != "" {
		if t, err := time.ParseDuration(speedTestTimeoutStr + "s"); err == nil {
			timeout = t
		}
	}

	var wg sync.WaitGroup
	// 限制并发数，避免同时发起过多连接
	sem := make(chan struct{}, 10)

	for i := range nodes {
		wg.Add(1)
		go func(n *models.Node) {
			defer wg.Done()
			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			if speedTestMode == "mihomo" {
				// Mihomo 真速度测速
				log.Printf("开始真速度测速节点: %s", n.Name)
				speed, latency, err := MihomoSpeedTest(n.Link, speedTestURL, timeout)
				if err != nil {
					n.Speed = 0
					n.DelayTime = -1
					log.Printf("节点真速度测速失败: %s, Error: %v", n.Name, err)
				} else {
					n.Speed = speed
					n.DelayTime = latency
					log.Printf("节点测速完成: %s, 速度: %dMB/s, 延迟: %dms", n.Name, speed, latency)
				}
			} else {
				// TCP Ping 模式
				// 解析链接获取 host 和 port
				host, port, err := utils.ParseNodeLink(n.Link)
				if err != nil {
					// 解析失败，跳过测速，但可以更新 LastCheck
					n.LastCheck = time.Now().Format("2006-01-02 15:04:05")
					n.UpdateSpeed()
					return
				}

				// 解析IP以进行调试
				realIP, err := utils.ResolveIP(host)
				if err != nil {
					log.Printf("解析域名失败: %s, Error: %v", host, err)
					// 解析失败，跳过测速
					n.LastCheck = time.Now().Format("2006-01-02 15:04:05")
					n.UpdateSpeed()
					return
				}

				log.Printf("开始测速节点: %s (%s:%d) [Real IP: %s]", n.Name, host, port, realIP)

				latency, err := utils.TcpPing(realIP, port, 3*time.Second)
				if err != nil {
					// 测速失败（超时或连接错误），DelayTime 设为 -1 表示不可达
					n.DelayTime = -1
					n.Speed = 0
					log.Printf("节点测速失败: %s, Error: %v", n.Name, err)
				} else {
					n.DelayTime = latency
					n.Speed = 0
					log.Printf("节点测速完成: %s, 延迟: %dms", n.Name, latency)
				}
			}
			n.LastCheck = time.Now().Format("2006-01-02 15:04:05")
			n.UpdateSpeed()
		}(&nodes[i])
	}
	wg.Wait()
	log.Println("节点测速任务执行完成")
}
