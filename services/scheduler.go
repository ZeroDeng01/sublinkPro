package services

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sublink/models"
	"sublink/node"
	"sublink/services/sse"
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
	err := node.LoadClashConfigFromURL(id, url, subName)
	if err != nil {
		sse.GetSSEBroker().BroadcastEvent("task_update", sse.NotificationPayload{
			Event:   "sub_update",
			Title:   "订阅更新失败",
			Message: fmt.Sprintf("订阅 [%s] 更新失败: %v", subName, err),
			Data: map[string]interface{}{
				"id":     id,
				"name":   subName,
				"status": "error",
			},
		})
	} else {
		sse.GetSSEBroker().BroadcastEvent("task_update", sse.NotificationPayload{
			Event:   "sub_update",
			Title:   "订阅更新成功",
			Message: fmt.Sprintf("订阅 [%s] 更新成功", subName),
			Data: map[string]interface{}{
				"id":     id,
				"name":   subName,
				"status": "success",
			},
		})
	}
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

	RunSpeedTestOnNodes(nodes)
}

// ExecuteSpecificNodeSpeedTestTask 执行指定节点测速任务
func ExecuteSpecificNodeSpeedTestTask(nodeIDs []int) {
	log.Printf("开始执行指定节点测速任务: %v", nodeIDs)
	if len(nodeIDs) == 0 {
		return
	}

	// 获取指定节点
	var nodes []models.Node
	for _, id := range nodeIDs {
		var n models.Node
		n.ID = id
		if err := n.GetByID(); err == nil {
			nodes = append(nodes, n)
		}
	}

	if len(nodes) == 0 {
		log.Println("未找到指定节点")
		return
	}

	RunSpeedTestOnNodes(nodes)
}

// RunSpeedTestOnNodes 对指定节点列表执行测速
func RunSpeedTestOnNodes(nodes []models.Node) {
	log.Printf("开始执行节点测速，总节点数: %d", len(nodes))

	// 确保函数最后一定会执行日志和通知
	defer func() {
		if r := recover(); r != nil {
			log.Printf("测速任务执行过程中发生严重错误: %v", r)
			sse.GetSSEBroker().BroadcastEvent("task_update", sse.NotificationPayload{
				Event:   "speed_test",
				Title:   "测速任务异常",
				Message: fmt.Sprintf("测速任务执行异常: %v", r),
				Data: map[string]interface{}{
					"status": "error",
				},
			})
		}
	}()

	// 获取测速配置
	speedTestUrl, _ := models.GetSetting("speed_test_url")
	speedTestTimeoutStr, _ := models.GetSetting("speed_test_timeout")
	speedTestTimeout := 5 * time.Second
	if speedTestTimeoutStr != "" {
		if d, err := time.ParseDuration(speedTestTimeoutStr + "s"); err == nil {
			speedTestTimeout = d
		}
	}

	// 获取测速模式
	speedTestMode, _ := models.GetSetting("speed_test_mode")

	// 并发控制
	concurrency := 10 // 默认并发数
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	// 结果统计
	var successCount, failCount int32
	var mu sync.Mutex

	// 总体超时控制
	// overallTimeout := time.Duration(len(nodes)) * speedTestTimeout / time.Duration(concurrency) * 2
	// if overallTimeout < 30*time.Second {
	// 	overallTimeout = 30 * time.Second
	// }
	done := make(chan struct{})

	go func() {
		for _, node := range nodes {
			wg.Add(1)
			sem <- struct{}{}
			go func(n models.Node) {
				defer wg.Done()
				defer func() { <-sem }()

				var speed float64
				var latency int
				var err error

				if speedTestMode == "tcp" {
					// 仅测试延迟
					latency, err = MihomoDelay(n.Link, speedTestUrl, speedTestTimeout)
					speed = 0
				} else {
					// 测试延迟和速度 (默认 mihomo)
					speed, latency, err = MihomoSpeedTest(n.Link, speedTestUrl, speedTestTimeout)
				}

				mu.Lock()
				defer mu.Unlock()

				if err != nil {
					failCount++
					log.Printf("节点 [%s] 测速失败: %v", n.Name, err)
					speed = -1
					latency = -1
					// 更新节点状态为失败
					n.Speed = speed
					n.DelayTime = latency
					n.LastCheck = time.Now().Format("2006-01-02 15:04:05")
					if err := n.UpdateSpeed(); err != nil {
						log.Printf("更新节点 %s 测速结果失败: %v", n.Name, err)
					}
				} else {
					successCount++
					log.Printf("节点 [%s] 测速成功: 速度 %.2f MB/s, 延迟 %d ms", n.Name, speed, latency)
					// 更新节点测速结果
					n.Speed = speed
					n.DelayTime = latency
					n.LastCheck = time.Now().Format("2006-01-02 15:04:05")
					if err := n.UpdateSpeed(); err != nil {
						log.Printf("更新节点 %s 测速结果失败: %v", n.Name, err)
					}
				}
			}(node)
		}
		wg.Wait()
		close(done)
	}()

	<-done
	// 完成
	log.Printf("节点测速任务执行完成 - 成功: %d, 失败: %d", successCount, failCount)
	sse.GetSSEBroker().BroadcastEvent("task_update", sse.NotificationPayload{
		Event:   "speed_test",
		Title:   "节点测速完成",
		Message: fmt.Sprintf("节点测速完成 (成功: %d, 失败: %d)", successCount, failCount),
		Data: map[string]interface{}{
			"status":  "success",
			"success": successCount,
			"fail":    failCount,
			"total":   len(nodes),
		},
	})
}
