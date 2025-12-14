package services

import (
	"fmt"
	"log"
	"math"
	"regexp"
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
	ExecuteSubscriptionTaskWithTrigger(id, url, subName, models.TaskTriggerScheduled)
}

// ExecuteSubscriptionTaskWithTrigger 执行订阅任务（带触发类型）
func ExecuteSubscriptionTaskWithTrigger(id int, url string, subName string, trigger models.TaskTrigger) {
	log.Printf("执行自动获取订阅任务 - ID: %d, Name: %s, URL: %s, Trigger: %s", id, subName, url, trigger)

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

	// 创建 TaskManager 任务和报告器
	tm := GetTaskManager()
	task, _, createErr := tm.CreateTask(models.TaskTypeSubUpdate, subName, trigger, 0)

	var reporter node.TaskReporter
	if createErr != nil {
		log.Printf("创建订阅更新任务失败: %v，将使用降级模式", createErr)
		reporter = nil // 使用 nil，将在 sub.go 中降级为 NoOpTaskReporter
	} else {
		reporter = &TaskManagerReporter{
			tm:     tm,
			taskID: task.ID,
		}
	}

	err := node.LoadClashConfigFromURLWithReporter(id, url, subName, downloadWithProxy, proxyLink, userAgent, reporter)
	if err != nil {
		// 仅在失败时发送通知，成功通知由 node/sub.go 中的 scheduleClashToNodeLinks 发送
		// 这样可以避免重复通知，且成功通知包含更详细的节点统计信息
		if reporter != nil {
			reporter.ReportFail(err.Error())
		}
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

// TaskManagerReporter 实现 node.TaskReporter 接口，用于将任务进度报告给 TaskManager
type TaskManagerReporter struct {
	tm     *TaskManager
	taskID string
}

func (r *TaskManagerReporter) UpdateTotal(total int) {
	r.tm.UpdateTotal(r.taskID, total)
}

func (r *TaskManagerReporter) ReportProgress(current int, currentItem string, result interface{}) {
	r.tm.UpdateProgress(r.taskID, current, currentItem, result)
}

func (r *TaskManagerReporter) ReportComplete(message string, result interface{}) {
	r.tm.CompleteTask(r.taskID, message, result)
}

func (r *TaskManagerReporter) ReportFail(errMsg string) {
	r.tm.FailTask(r.taskID, errMsg)
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

	// 使用 TaskManager 创建任务（定时触发）
	RunSpeedTestOnNodesWithTrigger(nodes, models.TaskTriggerScheduled)
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

	// 使用 TaskManager 创建任务（手动触发）
	RunSpeedTestOnNodesWithTrigger(nodes, models.TaskTriggerManual)
}

// RunSpeedTestOnNodes 对指定节点列表执行测速（向后兼容，默认手动触发）
// 采用两阶段测试策略：阶段一并发测延迟，阶段二低并发测速度
func RunSpeedTestOnNodes(nodes []models.Node) {
	RunSpeedTestOnNodesWithTrigger(nodes, models.TaskTriggerManual)
}

// RunSpeedTestOnNodesWithTrigger 对指定节点列表执行测速（带触发类型）
// 采用两阶段测试策略：阶段一并发测延迟，阶段二低并发测速度
// 支持通过 TaskManager 进行任务取消
func RunSpeedTestOnNodesWithTrigger(nodes []models.Node, trigger models.TaskTrigger) {
	if len(nodes) == 0 {
		log.Println("没有要测速的节点")
		return
	}

	totalNodes := len(nodes)
	log.Printf("开始执行节点测速，总节点数: %d, 触发类型: %s", totalNodes, trigger)

	// 使用 TaskManager 创建任务
	tm := GetTaskManager()
	task, ctx, err := tm.CreateTask(models.TaskTypeSpeedTest, "节点测速", trigger, totalNodes)
	if err != nil {
		log.Printf("创建测速任务失败: %v", err)
		return
	}
	taskID := task.ID

	// 确保任务结束时清理
	defer func() {
		if r := recover(); r != nil {
			log.Printf("测速任务执行过程中发生严重错误: %v", r)
			tm.FailTask(taskID, fmt.Sprintf("任务执行异常: %v", r))
		}
	}()

	// 检查是否已被取消
	if ctx.Err() != nil {
		log.Printf("任务已被取消: %s", taskID)
		return
	}

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

	// 获取落地IP查询接口URL
	landingIPUrl, _ := models.GetSetting("speed_test_landing_ip_url")
	if landingIPUrl == "" {
		landingIPUrl = "https://api.ipify.org" // 默认使用ipify
	}

	// 获取流量统计开关设置
	trafficByGroupStr, _ := models.GetSetting("speed_test_traffic_by_group")
	trafficByGroup := trafficByGroupStr != "false" // 默认开启
	trafficBySourceStr, _ := models.GetSetting("speed_test_traffic_by_source")
	trafficBySource := trafficBySourceStr != "false" // 默认开启
	trafficByNodeStr, _ := models.GetSetting("speed_test_traffic_by_node")
	trafficByNode := trafficByNodeStr == "true" // 默认关闭

	// 获取延迟采样次数
	latencySamplesStr, _ := models.GetSetting("speed_test_latency_samples")
	latencySamples := 3 // 默认3次采样
	if latencySamplesStr != "" {
		if s, err := strconv.Atoi(latencySamplesStr); err == nil && s > 0 {
			latencySamples = s
		}
	}

	// 获取是否包含握手时间（默认true，测量完整连接时间；false则预热后测量纯RTT）
	includeHandshakeStr, _ := models.GetSetting("speed_test_include_handshake")
	includeHandshake := includeHandshakeStr != "false" // 默认包含握手时间

	// 获取速度记录模式（average=平均速度, peak=峰值速度）
	speedRecordMode, _ := models.GetSetting("speed_test_speed_record_mode")
	if speedRecordMode == "" {
		speedRecordMode = "average"
	}

	// 获取峰值采样间隔（毫秒）
	peakSampleIntervalStr, _ := models.GetSetting("speed_test_peak_sample_interval")
	peakSampleInterval := 100 // 默认100ms
	if peakSampleIntervalStr != "" {
		if v, err := strconv.Atoi(peakSampleIntervalStr); err == nil && v >= 50 && v <= 200 {
			peakSampleInterval = v
		}
	}

	// 获取延迟测试并发数
	const maxConcurrency = 1000
	latencyConcurrency := 10 // 默认延迟测试并发数
	latencyConcurrencyStr, _ := models.GetSetting("speed_test_latency_concurrency")
	if latencyConcurrencyStr != "" {
		if c, err := strconv.Atoi(latencyConcurrencyStr); err == nil && c >= 0 {
			latencyConcurrency = c
		}
	}
	// 向后兼容：如果新配置为空，尝试读取旧配置
	if latencyConcurrencyStr == "" {
		oldConcurrencyStr, _ := models.GetSetting("speed_test_concurrency")
		if oldConcurrencyStr != "" {
			if c, err := strconv.Atoi(oldConcurrencyStr); err == nil && c >= 0 {
				latencyConcurrency = c
			}
		}
	}

	// 标记是否使用动态并发
	useAdaptiveLatency := latencyConcurrency == 0
	var latencyController *AdaptiveConcurrencyController
	if useAdaptiveLatency {
		latencyController = NewAdaptiveConcurrencyController(AdaptiveTypeLatency, totalNodes)
		latencyConcurrency = latencyController.GetCurrentConcurrency()
	} else {
		if latencyConcurrency > maxConcurrency {
			log.Printf("警告: 延迟并发数 %d 超过最大限制，已调整为 %d", latencyConcurrency, maxConcurrency)
			latencyConcurrency = maxConcurrency
		}
	}

	// 获取速度测试并发数
	speedConcurrency := 1 // 默认速度测试并发数为1（串行）
	speedConcurrencyStr, _ := models.GetSetting("speed_test_speed_concurrency")
	if speedConcurrencyStr != "" {
		if c, err := strconv.Atoi(speedConcurrencyStr); err == nil && c >= 0 {
			speedConcurrency = c
		}
	}

	// 标记是否使用动态并发
	useAdaptiveSpeed := speedConcurrency == 0
	var speedController *AdaptiveConcurrencyController
	if useAdaptiveSpeed {
		speedController = NewAdaptiveConcurrencyController(AdaptiveTypeSpeed, totalNodes)
		speedConcurrency = speedController.GetCurrentConcurrency()
	} else {
		// 硬性并发上限：速度测试不应超过8以避免带宽竞争
		const maxSpeedConcurrency = 32
		if speedConcurrency > maxSpeedConcurrency {
			log.Printf("警告: 速度并发数 %d 超过安全上限，已调整为 %d", speedConcurrency, maxSpeedConcurrency)
			speedConcurrency = maxSpeedConcurrency
		}
	}

	// 结果统计
	var successCount, failCount int32
	var completedCount int32
	var cancelled bool
	var mu sync.Mutex

	// 流量统计累加器（内存累计，测速结束时写入数据库）
	type trafficAccumulator struct {
		totalBytes   int64
		groupBytes   map[string]int64 // 按分组统计（可选）
		sourceBytes  map[string]int64 // 按来源统计（可选）
		nodeBytes    map[int]int64    // 按节点统计（可选，nodeID -> bytes）
		enableGroup  bool
		enableSource bool
		enableNode   bool
		mutex        sync.Mutex
	}
	trafficAcc := &trafficAccumulator{
		groupBytes:   make(map[string]int64),
		sourceBytes:  make(map[string]int64),
		nodeBytes:    make(map[int]int64),
		enableGroup:  trafficByGroup,
		enableSource: trafficBySource,
		enableNode:   trafficByNode,
	}

	// 节点结果存储（用于阶段二）
	type nodeResult struct {
		node    models.Node
		latency int
		err     error
	}
	nodeResults := make([]nodeResult, len(nodes))

	// 批量收集：测速结果列表（任务完成后批量写入数据库）
	speedTestResults := make([]models.SpeedTestResult, 0, len(nodes))

	// ========== 阶段一：延迟测试 ==========
	log.Printf("阶段一：开始延迟测试，并发数: %d（动态: %v），采样次数: %d", latencyConcurrency, useAdaptiveLatency, latencySamples)

	// 固定并发模式的 semaphore（仅在非动态模式下使用）
	var latencySem chan struct{}
	if !useAdaptiveLatency {
		latencySem = make(chan struct{}, latencyConcurrency)
	}
	var latencyWg sync.WaitGroup

	for i, node := range nodes {
		// 检查任务是否被取消
		select {
		case <-ctx.Done():
			mu.Lock()
			cancelled = true
			mu.Unlock()
			log.Printf("任务被取消，停止新的延迟测试")
			break
		default:
		}

		if cancelled {
			break
		}

		latencyWg.Add(1)

		// 根据是否使用动态并发选择不同的获取方式
		if useAdaptiveLatency && latencyController != nil {
			// 使用带延迟的获取方式，平滑任务启动
			latencyController.AcquireWithDelay()
		} else {
			latencySem <- struct{}{}
		}

		go func(idx int, n models.Node) {
			defer latencyWg.Done()
			defer func() {
				if useAdaptiveLatency && latencyController != nil {
					latencyController.ReleaseDynamic()
				} else {
					<-latencySem
				}
			}()

			// 在 goroutine 内再次检查取消状态
			select {
			case <-ctx.Done():
				return
			default:
			}

			// 使用多次采样测量延迟
			// TCP模式下需要检测IP（因为没有速度测试阶段），mihomo模式在速度阶段检测
			detectIPInLatency := detectCountry && speedTestMode == "tcp"
			latency, landingIP, err := mihomo.MihomoDelayWithSamples(n.Link, latencyTestUrl, speedTestTimeout, latencySamples, includeHandshake, detectIPInLatency, landingIPUrl)

			mu.Lock()
			defer mu.Unlock()

			// 再次检查取消状态
			if cancelled {
				return
			}

			nodeResults[idx] = nodeResult{node: n, latency: latency, err: err}
			currentCompleted := int(completedCount) + 1
			completedCount++

			// 向自适应控制器报告结果
			if useAdaptiveLatency && latencyController != nil {
				if err != nil {
					latencyController.ReportFailure()
				} else {
					latencyController.ReportSuccess(latency)
				}
				// 每N个任务检查一次是否需要调整
				if currentCompleted%LatencyAdjustCheckInterval == 0 {
					latencyController.MaybeAdjust()
				}
			}

			// TCP模式下收集结果（稍后批量写入）
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

					// TCP模式下处理落地IP检测结果
					if landingIP != "" {
						n.LandingIP = landingIP
						countryCode, geoErr := geoip.GetCountryISOCode(landingIP)
						if geoErr == nil && countryCode != "" {
							n.LinkCountry = countryCode
							log.Printf("节点 [%s] 落地IP: %s, 国家: %s", n.Name, landingIP, countryCode)
						}
					}
				}
				n.LatencyCheckAt = time.Now().Format("2006-01-02 15:04:05")
				// 收集结果到批量更新列表（不再立即写数据库）
				speedTestResults = append(speedTestResults, models.SpeedTestResult{
					NodeID:         n.ID,
					Speed:          n.Speed,
					SpeedStatus:    n.SpeedStatus,
					DelayTime:      n.DelayTime,
					DelayStatus:    n.DelayStatus,
					LatencyCheckAt: n.LatencyCheckAt,
					SpeedCheckAt:   "",
					LinkCountry:    n.LinkCountry,
					LandingIP:      n.LandingIP,
				})
			}

			// 更新任务进度
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
			tm.UpdateProgress(taskID, progressCurrent, n.Name, map[string]interface{}{
				"status":  resultStatus,
				"phase":   "latency",
				"latency": latency,
			})

			// 同时更新任务的 Total（如果是 mihomo 模式）
			if speedTestMode != "tcp" && idx == 0 {
				tm.UpdateTotal(taskID, progressTotal)
			}
		}(i, node)
	}
	latencyWg.Wait()
	log.Printf("阶段一完成：延迟测试结束")

	// 检查是否被取消
	if cancelled || ctx.Err() != nil {
		log.Printf("任务被取消，跳过阶段二 (已完成: %d/%d)", completedCount, totalNodes)
		tm.UpdateProgress(taskID, int(completedCount), "已取消", nil)
		// 任务已被 CancelTask 标记为取消，无需再次更新
		goto applyTags
	}

	// ========== 阶段二：速度测试（仅 mihomo 模式）==========
	if speedTestMode != "tcp" {
		log.Printf("阶段二：开始速度测试，并发数: %d（动态: %v）", speedConcurrency, useAdaptiveSpeed)

		// 重置进度计数器用于阶段二
		completedCount = 0

		// 固定并发模式的 semaphore（仅在非动态模式下使用）
		var speedSem chan struct{}
		if !useAdaptiveSpeed {
			speedSem = make(chan struct{}, speedConcurrency)
		}
		var speedWg sync.WaitGroup

		for i := range nodeResults {
			// 检查任务是否被取消
			select {
			case <-ctx.Done():
				mu.Lock()
				cancelled = true
				mu.Unlock()
				log.Printf("任务被取消，停止新的速度测试")
				break
			default:
			}

			if cancelled {
				break
			}

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
				// 收集结果到批量更新列表（不再立即写数据库）
				speedTestResults = append(speedTestResults, models.SpeedTestResult{
					NodeID:         nr.node.ID,
					Speed:          nr.node.Speed,
					SpeedStatus:    nr.node.SpeedStatus,
					DelayTime:      nr.node.DelayTime,
					DelayStatus:    nr.node.DelayStatus,
					LatencyCheckAt: nr.node.LatencyCheckAt,
					SpeedCheckAt:   "",
					LinkCountry:    nr.node.LinkCountry,
					LandingIP:      nr.node.LandingIP,
				})
				mu.Unlock()
				continue
			}

			speedWg.Add(1)

			// 根据是否使用动态并发选择不同的获取方式
			if useAdaptiveSpeed && speedController != nil {
				// 使用带延迟的获取方式，平滑任务启动
				speedController.AcquireWithDelay()
			} else {
				speedSem <- struct{}{}
			}

			go func(result *nodeResult) {
				defer speedWg.Done()
				defer func() {
					if useAdaptiveSpeed && speedController != nil {
						speedController.ReleaseDynamic()
					} else {
						<-speedSem
					}
				}()

				// 在 goroutine 内检查取消状态
				select {
				case <-ctx.Done():
					return
				default:
				}

				// 速度测试（延迟已在阶段一获取，同时可选检测落地IP）
				speed, _, bytesDownloaded, landingIP, err := mihomo.MihomoSpeedTest(result.node.Link, speedTestUrl, speedTestTimeout, detectCountry, landingIPUrl, speedRecordMode, peakSampleInterval)

				mu.Lock()
				defer mu.Unlock()

				// 再次检查取消状态
				if cancelled {
					return
				}

				// 累计流量统计（仅速度测试阶段，根据开关控制）
				if bytesDownloaded > 0 {
					trafficAcc.mutex.Lock()
					trafficAcc.totalBytes += bytesDownloaded

					// 按分组统计（可选）
					if trafficAcc.enableGroup {
						group := result.node.Group
						if group == "" {
							group = "未分组"
						}
						trafficAcc.groupBytes[group] += bytesDownloaded
					}

					// 按来源统计（可选）
					if trafficAcc.enableSource {
						source := result.node.Source
						if source == "" || source == "manual" {
							source = "手动添加"
						}
						trafficAcc.sourceBytes[source] += bytesDownloaded
					}

					// 按节点统计（可选）
					if trafficAcc.enableNode {
						trafficAcc.nodeBytes[result.node.ID] += bytesDownloaded
					}
					trafficAcc.mutex.Unlock()
				}

				currentCompleted := int(completedCount) + 1
				completedCount++

				// 向自适应控制器报告结果
				if useAdaptiveSpeed && speedController != nil {
					if err != nil {
						speedController.ReportFailure()
					} else {
						speedController.ReportSuccess(int(speed * 100)) // 速度转换为整数以便统计
					}
					// 每N个任务检查一次是否需要调整（速度测试调整更频繁）
					if currentCompleted%SpeedAdjustCheckInterval == 0 {
						speedController.MaybeAdjust()
					}
				}

				var resultStatus string
				var resultData map[string]interface{}

				if err != nil {
					failCount++
					log.Printf("节点 [%s] 速度测试失败: %v (延迟: %d ms, 已下载: %s)", result.node.Name, err, result.latency, formatBytes(bytesDownloaded))
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
					log.Printf("节点 [%s] 测速成功: 速度 %.2f MB/s, 延迟 %d ms, 流量消耗: %s", result.node.Name, speed, result.latency, formatBytes(bytesDownloaded))
					result.node.Speed = speed
					result.node.SpeedStatus = constants.StatusSuccess
					result.node.DelayTime = result.latency
					result.node.DelayStatus = constants.StatusSuccess
					resultStatus = "success"
					resultData = map[string]interface{}{
						"speed":   speed,
						"latency": result.latency,
					}

					// 处理落地IP检测结果（已由MihomoSpeedTest内部完成）
					if landingIP != "" {
						result.node.LandingIP = landingIP
						countryCode, geoErr := geoip.GetCountryISOCode(landingIP)
						if geoErr == nil && countryCode != "" {
							result.node.LinkCountry = countryCode
							log.Printf("节点 [%s] 落地IP: %s, 国家: %s", result.node.Name, landingIP, countryCode)
						}
					}
				}

				result.node.LatencyCheckAt = time.Now().Format("2006-01-02 15:04:05")
				result.node.SpeedCheckAt = time.Now().Format("2006-01-02 15:04:05")
				// 收集结果到批量更新列表（不再立即写数据库）
				speedTestResults = append(speedTestResults, models.SpeedTestResult{
					NodeID:         result.node.ID,
					Speed:          result.node.Speed,
					SpeedStatus:    result.node.SpeedStatus,
					DelayTime:      result.node.DelayTime,
					DelayStatus:    result.node.DelayStatus,
					LatencyCheckAt: result.node.LatencyCheckAt,
					SpeedCheckAt:   result.node.SpeedCheckAt,
					LinkCountry:    result.node.LinkCountry,
					LandingIP:      result.node.LandingIP,
				})

				// 获取当前流量统计（用于实时显示）
				trafficAcc.mutex.Lock()
				currentTrafficTotal := trafficAcc.totalBytes
				currentTrafficFormatted := formatBytes(currentTrafficTotal)
				trafficAcc.mutex.Unlock()

				// 更新任务进度（速度测试占后50%）
				tm.UpdateProgress(taskID, totalNodes+currentCompleted, result.node.Name, map[string]interface{}{
					"status":  resultStatus,
					"phase":   "speed",
					"speed":   result.node.Speed,
					"latency": result.latency,
					"data":    resultData,
					"traffic": map[string]interface{}{
						"totalBytes":     currentTrafficTotal,
						"totalFormatted": currentTrafficFormatted,
					},
				})
			}(nr)
		}
		speedWg.Wait()
		log.Printf("阶段二完成：速度测试结束")
	}

	// 检查最终是否被取消
	if cancelled || ctx.Err() != nil {
		log.Printf("任务被取消")
		goto applyTags
	}

	// 批量写入所有测速结果到数据库（一次性操作，减少数据库I/O）
	if len(speedTestResults) > 0 {
		if err := models.BatchUpdateSpeedResults(speedTestResults); err != nil {
			log.Printf("❌批量更新测速结果失败：%v", err)
		} else {
			log.Printf("✅批量更新 %d 个节点测速结果成功", len(speedTestResults))
		}
	}

	// 完成任务
	{
		// 格式化流量统计数据
		trafficAcc.mutex.Lock()
		// 构建流量统计对象
		trafficData := map[string]interface{}{
			"totalBytes":     trafficAcc.totalBytes,
			"totalFormatted": formatBytes(trafficAcc.totalBytes),
		}

		// 按分组统计（仅开关开启时包含）
		if trafficAcc.enableGroup && len(trafficAcc.groupBytes) > 0 {
			formattedGroupStats := make(map[string]map[string]interface{})
			for group, bytes := range trafficAcc.groupBytes {
				formattedGroupStats[group] = map[string]interface{}{
					"bytes":     bytes,
					"formatted": formatBytes(bytes),
				}
			}
			trafficData["byGroup"] = formattedGroupStats
		}

		// 按来源统计（仅开关开启时包含）
		if trafficAcc.enableSource && len(trafficAcc.sourceBytes) > 0 {
			formattedSourceStats := make(map[string]map[string]interface{})
			for source, bytes := range trafficAcc.sourceBytes {
				formattedSourceStats[source] = map[string]interface{}{
					"bytes":     bytes,
					"formatted": formatBytes(bytes),
				}
			}
			trafficData["bySource"] = formattedSourceStats
		}

		// 按节点统计（仅开关开启时包含，只存nodeID和bytes减少数据量）
		if trafficAcc.enableNode && len(trafficAcc.nodeBytes) > 0 {
			trafficData["byNode"] = trafficAcc.nodeBytes
		}

		trafficTotal := trafficAcc.totalBytes
		trafficAcc.mutex.Unlock()

		resultData := map[string]interface{}{
			"success": successCount,
			"fail":    failCount,
			"total":   totalNodes,
			"traffic": trafficData,
		}
		tm.CompleteTask(taskID, fmt.Sprintf("测速完成 (成功: %d, 失败: %d, 流量: %s)", successCount, failCount, formatBytes(trafficTotal)), resultData)

		// 广播测速完成通知（让用户在通知中心看到）
		sse.GetSSEBroker().BroadcastEvent("task_update", sse.NotificationPayload{
			Event:   "speed_test",
			Title:   "节点测速完成",
			Message: fmt.Sprintf("测速完成: 成功 %d 个, 失败 %d 个, 消耗流量 %s", successCount, failCount, formatBytes(trafficTotal)),
			Data: map[string]interface{}{
				"status":  "success",
				"success": successCount,
				"fail":    failCount,
				"total":   totalNodes,
				"traffic": formatBytes(trafficTotal),
			},
		})
	}

applyTags:
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

// formatBytes 格式化字节数为人类可读格式
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	if bytes < 0 {
		return "N/A"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	// B, KB, MB, GB, TB
	units := []string{"B", "KB", "MB", "GB", "TB"}
	if exp >= len(units)-1 {
		exp = len(units) - 2
	}

	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp+1])
}
