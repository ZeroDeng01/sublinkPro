package services

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sublink/constants"
	"sublink/models"
	"sublink/node"
	"sublink/services/geoip"
	"sublink/services/mihomo"
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

	// 获取最新的订阅配置，以便使用最新的代理设置
	var subS models.SubScheduler
	var downloadWithProxy bool
	var proxyLink string
	var userAgent string

	if err := subS.GetByID(id); err != nil {
		log.Printf("获取订阅配置失败 ID: %d, 使用默认设置: %v", id, err)
	} else {
		downloadWithProxy = subS.DownloadWithProxy
		proxyLink = subS.ProxyLink
		userAgent = subS.UserAgent
	}

	err := node.LoadClashConfigFromURL(id, url, subName, downloadWithProxy, proxyLink, userAgent)
	if err != nil {
		// 仅在失败时发送通知，成功通知由 node/sub.go 中的 scheduleClashToNodeLinks 发送
		// 这样可以避免重复通知，且成功通知包含更详细的节点统计信息
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
		return
	}

	// 订阅更新成功后，应用自动标签规则
	go func() {
		updatedNodes, err := models.ListBySourceID(id)
		if err == nil && len(updatedNodes) > 0 {
			ApplyAutoTagRules(updatedNodes, "subscription_update")
		}
	}()
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

	// 获取测速分组和标签配置
	speedTestGroupsStr, _ := models.GetSetting("speed_test_groups")
	speedTestTagsStr, _ := models.GetSetting("speed_test_tags")
	var nodes []models.Node
	var err error

	// 分组优先级高于标签：
	// 1. 如果选了分组，先按分组筛选，再按标签过滤
	// 2. 如果只选了标签，直接按标签筛选
	// 3. 都不选则测全部
	if speedTestGroupsStr != "" {
		groups := strings.Split(speedTestGroupsStr, ",")
		nodes, err = new(models.Node).ListByGroups(groups)
		log.Printf("根据分组测速: %v", groups)

		// 在分组基础上按标签继续筛选
		if err == nil && speedTestTagsStr != "" {
			tags := strings.Split(speedTestTagsStr, ",")
			nodes = models.FilterNodesByTags(nodes, tags)
			log.Printf("在分组基础上按标签过滤: %v, 剩余节点: %d", tags, len(nodes))
		}
	} else if speedTestTagsStr != "" {
		tags := strings.Split(speedTestTagsStr, ",")
		nodes, err = new(models.Node).ListByTags(tags)
		log.Printf("根据标签测速: %v", tags)
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
// 采用两阶段测试策略：阶段一并发测延迟，阶段二低并发测速度
func RunSpeedTestOnNodes(nodes []models.Node) {
	log.Printf("开始执行节点测速，总节点数: %d", len(nodes))

	// 生成唯一任务ID
	taskID := fmt.Sprintf("speed_test_%d", time.Now().UnixNano())
	totalNodes := len(nodes)
	startTime := time.Now()
	startTimeMs := startTime.UnixMilli()

	// 广播任务开始事件
	sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
		TaskID:    taskID,
		TaskType:  "speed_test",
		TaskName:  "节点测速",
		Status:    "started",
		Current:   0,
		Total:     totalNodes,
		Message:   fmt.Sprintf("开始测速 %d 个节点", totalNodes),
		StartTime: startTimeMs,
	})

	// 确保函数最后一定会执行日志和通知
	defer func() {
		if r := recover(); r != nil {
			log.Printf("测速任务执行过程中发生严重错误: %v", r)
			// 广播错误进度
			sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
				TaskID:   taskID,
				TaskType: "speed_test",
				TaskName: "节点测速",
				Status:   "error",
				Message:  fmt.Sprintf("测速任务执行异常: %v", r),
			})
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
	latencyTestUrl, _ := models.GetSetting("speed_test_latency_url")
	// 向后兼容：如果未配置延迟URL，使用速度URL
	if latencyTestUrl == "" {
		latencyTestUrl = speedTestUrl
	}
	speedTestTimeoutStr, _ := models.GetSetting("speed_test_timeout")
	speedTestTimeout := 5 * time.Second
	if speedTestTimeoutStr != "" {
		if d, err := time.ParseDuration(speedTestTimeoutStr + "s"); err == nil {
			speedTestTimeout = d
		}
	}

	// 获取测速模式
	speedTestMode, _ := models.GetSetting("speed_test_mode")

	// 获取是否检测落地IP国家
	detectCountryStr, _ := models.GetSetting("speed_test_detect_country")
	detectCountry := detectCountryStr == "true"

	// 获取延迟采样次数
	latencySamplesStr, _ := models.GetSetting("speed_test_latency_samples")
	latencySamples := 3 // 默认3次采样
	if latencySamplesStr != "" {
		if s, err := strconv.Atoi(latencySamplesStr); err == nil && s > 0 {
			latencySamples = s
		}
	}

	// 获取延迟测试并发数
	const maxConcurrency = 1000
	latencyConcurrency := 10 // 默认延迟测试并发数
	latencyConcurrencyStr, _ := models.GetSetting("speed_test_latency_concurrency")
	if latencyConcurrencyStr != "" {
		if c, err := strconv.Atoi(latencyConcurrencyStr); err == nil && c > 0 {
			latencyConcurrency = c
		}
	}
	// 向后兼容：如果新配置为空，尝试读取旧配置
	if latencyConcurrencyStr == "" {
		oldConcurrencyStr, _ := models.GetSetting("speed_test_concurrency")
		if oldConcurrencyStr != "" {
			if c, err := strconv.Atoi(oldConcurrencyStr); err == nil && c > 0 {
				latencyConcurrency = c
			}
		}
	}
	// 自动设置（如果为0）
	if latencyConcurrency <= 0 {
		cpuCount := runtime.NumCPU()
		latencyConcurrency = cpuCount * 2
		if latencyConcurrency < 2 {
			latencyConcurrency = 2
		}
		log.Printf("自动设置延迟测试并发数: %d (基于 %d CPU核心)", latencyConcurrency, cpuCount)
	}
	if latencyConcurrency > maxConcurrency {
		log.Printf("警告: 延迟并发数 %d 超过最大限制，已调整为 %d", latencyConcurrency, maxConcurrency)
		latencyConcurrency = maxConcurrency
	}

	// 获取速度测试并发数
	speedConcurrency := 1 // 默认速度测试并发数为1（串行）
	speedConcurrencyStr, _ := models.GetSetting("speed_test_speed_concurrency")
	if speedConcurrencyStr != "" {
		if c, err := strconv.Atoi(speedConcurrencyStr); err == nil && c > 0 {
			speedConcurrency = c
		}
	}
	if speedConcurrency > 128 {
		log.Printf("警告: 速度并发数 %d 超过建议值，已调整为 128", speedConcurrency)
		speedConcurrency = 128 // 速度测试并发数上限较低，避免带宽竞争
	}

	// 结果统计
	var successCount, failCount int32
	var completedCount int32
	var mu sync.Mutex

	// 节点结果存储（用于阶段二）
	type nodeResult struct {
		node    models.Node
		latency int
		err     error
	}
	nodeResults := make([]nodeResult, len(nodes))

	// ========== 阶段一：延迟测试 ==========
	log.Printf("阶段一：开始延迟测试，并发数: %d，采样次数: %d", latencyConcurrency, latencySamples)
	latencySem := make(chan struct{}, latencyConcurrency)
	var latencyWg sync.WaitGroup

	for i, node := range nodes {
		latencyWg.Add(1)
		latencySem <- struct{}{}
		go func(idx int, n models.Node) {
			defer latencyWg.Done()
			defer func() { <-latencySem }()

			// 使用多次采样测量延迟
			latency, err := mihomo.MihomoDelayWithSamples(n.Link, latencyTestUrl, speedTestTimeout, latencySamples)

			mu.Lock()
			nodeResults[idx] = nodeResult{node: n, latency: latency, err: err}
			currentCompleted := int(completedCount) + 1
			completedCount++

			// TCP模式下直接更新结果
			if speedTestMode == "tcp" {
				if err != nil {
					failCount++
					log.Printf("节点 [%s] 延迟测试失败: %v", n.Name, err)
					n.Speed = -1
					n.SpeedStatus = constants.StatusUntested // TCP模式不测速度
					n.DelayTime = -1
					n.DelayStatus = constants.StatusTimeout
				} else {
					successCount++
					log.Printf("节点 [%s] 延迟测试成功: %d ms", n.Name, latency)
					n.Speed = 0 // TCP模式不测速度
					n.SpeedStatus = constants.StatusUntested
					n.DelayTime = latency
					n.DelayStatus = constants.StatusSuccess
				}
				n.LatencyCheckAt = time.Now().Format("2006-01-02 15:04:05")
				if updateErr := n.UpdateSpeed(); updateErr != nil {
					log.Printf("更新节点 %s 测速结果失败: %v", n.Name, updateErr)
				}
			}
			mu.Unlock()

			// 广播进度
			var resultStatus string
			if err != nil {
				resultStatus = "failed"
			} else {
				resultStatus = "success"
			}
			// TCP模式: 进度即为实际完成数; mihomo模式: 延迟测试占前50%
			progressCurrent := currentCompleted
			progressTotal := totalNodes
			if speedTestMode != "tcp" {
				// mihomo模式下，延迟测试占前半部分
				progressTotal = totalNodes * 2
			}
			sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
				TaskID:      taskID,
				TaskType:    "speed_test",
				TaskName:    "节点测速",
				Status:      "progress",
				Current:     progressCurrent,
				Total:       progressTotal,
				CurrentItem: n.Name,
				Result: map[string]interface{}{
					"status":  resultStatus,
					"phase":   "latency",
					"latency": latency,
				},
				Message:   fmt.Sprintf("【延迟测试】 %d/%d", currentCompleted, totalNodes),
				StartTime: startTimeMs,
			})
		}(i, node)
	}
	latencyWg.Wait()
	log.Printf("阶段一完成：延迟测试结束")

	// ========== 阶段二：速度测试（仅 mihomo 模式）==========
	if speedTestMode != "tcp" {
		log.Printf("阶段二：开始速度测试，并发数: %d", speedConcurrency)

		// 重置进度计数器用于阶段二
		completedCount = 0
		speedSem := make(chan struct{}, speedConcurrency)
		var speedWg sync.WaitGroup

		for i := range nodeResults {
			nr := &nodeResults[i]
			// 跳过延迟测试失败的节点
			if nr.err != nil {
				mu.Lock()
				failCount++
				completedCount++
				nr.node.Speed = -1
				nr.node.SpeedStatus = constants.StatusError // 因延迟失败无法测速
				nr.node.DelayTime = -1
				nr.node.DelayStatus = constants.StatusTimeout
				nr.node.LatencyCheckAt = time.Now().Format("2006-01-02 15:04:05")
				if updateErr := nr.node.UpdateSpeed(); updateErr != nil {
					log.Printf("更新节点 %s 测速结果失败: %v", nr.node.Name, updateErr)
				}
				mu.Unlock()
				continue
			}

			speedWg.Add(1)
			speedSem <- struct{}{}
			go func(result *nodeResult) {
				defer speedWg.Done()
				defer func() { <-speedSem }()

				// 速度测试（延迟已在阶段一获取）
				speed, _, err := mihomo.MihomoSpeedTest(result.node.Link, speedTestUrl, speedTestTimeout)

				mu.Lock()
				defer mu.Unlock()

				currentCompleted := int(completedCount) + 1
				completedCount++

				var resultStatus string
				var resultData map[string]interface{}

				if err != nil {
					failCount++
					log.Printf("节点 [%s] 速度测试失败: %v (延迟: %d ms)", result.node.Name, err, result.latency)
					result.node.Speed = -1
					result.node.SpeedStatus = constants.StatusError
					result.node.DelayTime = result.latency            // 保留延迟测试结果
					result.node.DelayStatus = constants.StatusSuccess // 延迟测试是成功的
					resultStatus = "failed"
					resultData = map[string]interface{}{
						"speed":   -1,
						"latency": result.latency,
						"error":   err.Error(),
					}
				} else {
					successCount++
					log.Printf("节点 [%s] 测速成功: 速度 %.2f MB/s, 延迟 %d ms", result.node.Name, speed, result.latency)
					result.node.Speed = speed
					result.node.SpeedStatus = constants.StatusSuccess
					result.node.DelayTime = result.latency
					result.node.DelayStatus = constants.StatusSuccess
					resultStatus = "success"
					resultData = map[string]interface{}{
						"speed":   speed,
						"latency": result.latency,
					}

					// 如果开启落地IP检测
					if detectCountry {
						landingIP, countryErr := mihomo.FetchLandingIP(result.node.Link, speedTestTimeout)
						if countryErr == nil && landingIP != "" {
							countryCode, geoErr := geoip.GetCountryISOCode(landingIP)
							if geoErr == nil && countryCode != "" {
								result.node.LinkCountry = countryCode
								log.Printf("节点 [%s] 落地IP: %s, 国家: %s", result.node.Name, landingIP, countryCode)
							} else {
								log.Printf("节点 [%s] 获取国家代码失败: %v", result.node.Name, geoErr)
							}
						} else {
							log.Printf("节点 [%s] 获取落地IP失败: %v", result.node.Name, countryErr)
						}
					}
				}

				result.node.LatencyCheckAt = time.Now().Format("2006-01-02 15:04:05")
				result.node.SpeedCheckAt = time.Now().Format("2006-01-02 15:04:05")
				if updateErr := result.node.UpdateSpeed(); updateErr != nil {
					log.Printf("更新节点 %s 测速结果失败: %v", result.node.Name, updateErr)
				}

				// 广播进度（速度测试占后50%）
				sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
					TaskID:      taskID,
					TaskType:    "speed_test",
					TaskName:    "节点测速",
					Status:      "progress",
					Current:     totalNodes + currentCompleted, // 进度从50%开始
					Total:       totalNodes * 2,
					CurrentItem: result.node.Name,
					Result: map[string]interface{}{
						"status":  resultStatus,
						"phase":   "speed",
						"speed":   result.node.Speed,
						"latency": result.latency,
						"data":    resultData,
					},
					Message:   fmt.Sprintf("【速度测试】 %d/%d", currentCompleted, totalNodes),
					StartTime: startTimeMs,
				})
			}(nr)
		}
		speedWg.Wait()
		log.Printf("阶段二完成：速度测试结束")
	}

	// 广播完成进度
	sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
		TaskID:   taskID,
		TaskType: "speed_test",
		TaskName: "节点测速",
		Status:   "completed",
		Current:  totalNodes,
		Total:    totalNodes,
		Message:  fmt.Sprintf("测速完成 (成功: %d, 失败: %d)", successCount, failCount),
		Result: map[string]interface{}{
			"success": successCount,
			"fail":    failCount,
		},
	})

	// 完成 (触发webhook)
	duration := time.Since(startTime)
	durationStr := formatDuration(duration)
	log.Printf("节点测速任务执行完成 - 成功: %d, 失败: %d, 耗时: %s", successCount, failCount, durationStr)
	sse.GetSSEBroker().BroadcastEvent("task_update", sse.NotificationPayload{
		Event:   "speed_test",
		Title:   "节点测速完成",
		Message: fmt.Sprintf("节点测速完成，耗时 %s (成功: %d, 失败: %d)", durationStr, successCount, failCount),
		Data: map[string]interface{}{
			"status":   "success",
			"success":  successCount,
			"fail":     failCount,
			"total":    len(nodes),
			"duration": duration.Milliseconds(),
		},
	})

	// 应用自动标签规则 - 测速完成后触发
	// 重新获取已测速节点的最新数据（包含更新后的速度/延迟值）
	go func() {
		// 收集测速节点的ID
		testedNodeIDs := make([]int, 0, len(nodes))
		for _, n := range nodes {
			testedNodeIDs = append(testedNodeIDs, n.ID)
		}

		// 从数据库/缓存获取最新的节点数据
		updatedNodes, err := models.GetNodesByIDs(testedNodeIDs)
		if err != nil {
			log.Printf("获取测速节点最新数据失败: %v, 使用原始数据", err)
			ApplyAutoTagRules(nodes, "speed_test")
			return
		}

		ApplyAutoTagRules(updatedNodes, "speed_test")
	}()

}

// formatDuration 格式化时长为人类可读字符串
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f秒", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0f分%.0f秒", d.Minutes(), math.Mod(d.Seconds(), 60))
	}
	return fmt.Sprintf("%.0f时%.0f分", d.Hours(), math.Mod(d.Minutes(), 60))
}
