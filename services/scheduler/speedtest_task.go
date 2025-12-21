package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"sublink/constants"
	"sublink/models"
	"sublink/node"
	"sublink/services/geoip"
	"sublink/services/mihomo"
	"sublink/services/sse"
	"sublink/utils"
	"sync"
	"sync/atomic"
	"time"
)

// StartNodeSpeedTestTask 启动节点测速定时任务
func (sm *SchedulerManager) StartNodeSpeedTestTask(cronExpr string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 清理Cron表达式
	cleanCronExpr := cleanCronExpression(cronExpr)

	// 如果任务已存在，先删除
	if entryID, exists := sm.jobs[JobIDSpeedTest]; exists {
		sm.cron.Remove(entryID)
		delete(sm.jobs, JobIDSpeedTest)
	}

	// 添加新任务
	entryID, err := sm.cron.AddFunc(cleanCronExpr, func() {
		ExecuteNodeSpeedTestTask()
	})

	if err != nil {
		utils.Error("添加节点测速任务失败 - Cron: %s, Error: %v", cleanCronExpr, err)
		return err
	}

	// 存储任务映射
	sm.jobs[JobIDSpeedTest] = entryID
	utils.Info("成功添加节点测速任务 - JobID: %d", JobIDSpeedTest)
	utils.Info("成功添加节点测速任务 - Cron: %s", cleanCronExpr)
	return nil
}

// StopNodeSpeedTestTask 停止节点测速定时任务
func (sm *SchedulerManager) StopNodeSpeedTestTask() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if entryID, exists := sm.jobs[JobIDSpeedTest]; exists {
		sm.cron.Remove(entryID)
		delete(sm.jobs, JobIDSpeedTest)
		utils.Info("成功停止节点测速任务")
	}
}

// ExecuteNodeSpeedTestTask 执行节点测速任务
func ExecuteNodeSpeedTestTask() {
	utils.Info("开始执行节点测速任务...")

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
		utils.Debug("根据分组测速: %v", groups)

		// 在分组基础上按标签继续筛选
		if err == nil && speedTestTagsStr != "" {
			tags := strings.Split(speedTestTagsStr, ",")
			nodes = models.FilterNodesByTags(nodes, tags)
			utils.Debug("在分组基础上按标签过滤: %v, 剩余节点: %d", tags, len(nodes))
		}
	} else if speedTestTagsStr != "" {
		tags := strings.Split(speedTestTagsStr, ",")
		nodes, err = new(models.Node).ListByTags(tags)
		utils.Debug("根据标签测速: %v", tags)
	} else {
		nodes, err = new(models.Node).List()
		utils.Debug("全量测速")
	}

	if err != nil {
		utils.Error("获取节点列表失败: %v", err)
		return
	}

	// 使用 TaskManager 创建任务（定时触发）
	RunSpeedTestOnNodesWithTrigger(nodes, models.TaskTriggerScheduled)
}

// ExecuteSpecificNodeSpeedTestTask 执行指定节点测速任务
func ExecuteSpecificNodeSpeedTestTask(nodeIDs []int) {
	utils.Debug("开始执行指定节点测速任务: %v", nodeIDs)
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
		utils.Warn("未找到指定节点")
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
		utils.Warn("没有要测速的节点")
		return
	}

	totalNodes := len(nodes)
	utils.Info("开始执行节点测速，总节点数: %d, 触发类型: %s", totalNodes, trigger)

	// 使用 TaskManager 创建任务
	tm := getTaskManager()
	task, ctx, err := tm.CreateTask(models.TaskTypeSpeedTest, "节点测速", trigger, totalNodes)
	if err != nil {
		utils.Error("创建测速任务失败: %v", err)
		return
	}
	taskID := task.ID

	// 确保任务结束时清理
	defer func() {
		if r := recover(); r != nil {
			utils.Error("测速任务执行过程中发生严重错误: %v", r)
			tm.FailTask(taskID, fmt.Sprintf("任务执行异常: %v", r))
		}
	}()

	// 检查是否已被取消
	if ctx.Err() != nil {
		utils.Info("任务已被取消: %s", taskID)
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

	// 获取是否包含握手时间（默认true，测量完整连接时间；false则使用 UnifiedDelay 模式排除握手）
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

	// 获取是否持久化Host（测速成功时自动保存域名->IP映射）
	persistHostStr, _ := models.GetSetting("speed_test_persist_host")
	persistHost := persistHostStr == "true"

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
	var latencyController AdaptiveConcurrencyController
	if useAdaptiveLatency {
		latencyController = newAdaptiveConcurrencyController(AdaptiveTypeLatency, totalNodes)
		latencyConcurrency = latencyController.GetCurrentConcurrency()
	} else {
		if latencyConcurrency > maxConcurrency {
			utils.Warn("警告: 延迟并发数 %d 超过最大限制，已调整为 %d", latencyConcurrency, maxConcurrency)
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
	var speedController AdaptiveConcurrencyController
	if useAdaptiveSpeed {
		speedController = newAdaptiveConcurrencyController(AdaptiveTypeSpeed, totalNodes)
		speedConcurrency = speedController.GetCurrentConcurrency()
	} else {
		// 硬性并发上限：速度测试不应超过32以避免带宽竞争
		const maxSpeedConcurrency = 32
		if speedConcurrency > maxSpeedConcurrency {
			utils.Warn("警告: 速度并发数 %d 超过安全上限，已调整为 %d", speedConcurrency, maxSpeedConcurrency)
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

	// 批量收集：Host映射信息（测速成功时收集，任务完成后批量保存）
	hostMappings := make([]models.HostMappingInfo, 0)
	var hostMu sync.Mutex

	// ========== 阶段一：延迟测试 ==========
	utils.Info("阶段一：开始延迟测试，并发数: %d（动态: %v），UnifiedDelay: %v", latencyConcurrency, useAdaptiveLatency, !includeHandshake)

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
			utils.Debug("任务被取消，停止新的延迟测试")
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

			// 使用 Mihomo URLTest 测量延迟
			// TCP模式下需要检测IP（因为没有速度测试阶段），mihomo模式在速度阶段检测
			detectIPInLatency := detectCountry && speedTestMode == "tcp"
			latency, landingIP, err := mihomo.MihomoDelayTest(n.Link, latencyTestUrl, speedTestTimeout, includeHandshake, detectIPInLatency, landingIPUrl)

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
				if currentCompleted%latencyAdjustCheckInterval == 0 {
					latencyController.MaybeAdjust()
				}
			}

			// TCP模式下收集结果（稍后批量写入）
			if speedTestMode == "tcp" {
				if err != nil {
					failCount++
					utils.Debug("节点 [%s] 延迟测试失败: %v", n.Name, err)
					n.Speed = -1
					n.SpeedStatus = constants.StatusUntested // TCP模式不测速度
					n.DelayTime = -1
					n.DelayStatus = constants.StatusTimeout
				} else {
					successCount++
					utils.Debug("节点 [%s] 延迟测试成功: %d ms", n.Name, latency)
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
							utils.Debug("节点 [%s] 落地IP: %s, 国家: %s", n.Name, landingIP, countryCode)
						}
					}

					// 持久化Host：测速成功时从 link 解析服务器地址并解析DNS
					if persistHost {
						hostInfo := mihomo.GetProxyServerFromLink(n.Link)
						// 只处理域名，跳过已是IP的地址
						if hostInfo.Server != "" && !hostInfo.IsIP {
							if resolvedIP, dnsSource := mihomo.ResolveProxyHost(hostInfo.Server); resolvedIP != "" {
								hostMu.Lock()
								hostMappings = append(hostMappings, models.HostMappingInfo{
									Hostname:  hostInfo.Server,
									IP:        resolvedIP,
									NodeName:  n.Name,
									Group:     n.Group,
									Source:    n.Source,
									DNSSource: dnsSource,
								})
								hostMu.Unlock()
							}
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
			// 格式化节点显示项（包含分组和来源信息，方便手机端查看）
			currentItemDisplay := formatNodeDisplayItem(n.Name, n.Group, n.Source)
			tm.UpdateProgress(taskID, progressCurrent, currentItemDisplay, map[string]interface{}{
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
	utils.Info("阶段一完成：延迟测试结束")

	// 检查是否被取消
	if cancelled || ctx.Err() != nil {
		utils.Info("任务被取消，跳过阶段二 (已完成: %d/%d)", completedCount, totalNodes)
		tm.UpdateProgress(taskID, int(completedCount), "已取消", nil)
		// 任务已被 CancelTask 标记为取消，无需再次更新
		goto applyTags
	}

	// ========== 阶段二：速度测试（仅 mihomo 模式）==========
	if speedTestMode != "tcp" {
		utils.Info("阶段二：开始速度测试，并发数: %d（动态: %v）", speedConcurrency, useAdaptiveSpeed)

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
				utils.Debug("任务被取消，停止新的速度测试")
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

				currentCompleted := int(atomic.AddInt32(&completedCount, 1))

				// 向自适应控制器报告结果
				if useAdaptiveSpeed && speedController != nil {
					if err != nil {
						speedController.ReportFailure()
					} else {
						speedController.ReportSuccess(int(speed * 100)) // 速度转换为整数以便统计
					}
					// 每N个任务检查一次是否需要调整（速度测试调整更频繁）
					if currentCompleted%speedAdjustCheckInterval == 0 {
						speedController.MaybeAdjust()
					}
				}

				var resultStatus string
				var resultData map[string]interface{}

				if err != nil {
					atomic.AddInt32(&failCount, 1)
					utils.Debug("节点 [%s] 速度测试失败: %v (延迟: %d ms, 已下载: %s)", result.node.Name, err, result.latency, formatBytes(bytesDownloaded))
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
					atomic.AddInt32(&successCount, 1)
					utils.Debug("节点 [%s] 测速成功: 速度 %.2f MB/s, 延迟 %d ms, 流量消耗: %s", result.node.Name, speed, result.latency, formatBytes(bytesDownloaded))
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
							utils.Debug("节点 [%s] 落地IP: %s, 国家: %s", result.node.Name, landingIP, countryCode)
						}
					}

					// 持久化Host：测速成功时从 link 解析服务器地址并解析DNS
					if persistHost {
						hostInfo := mihomo.GetProxyServerFromLink(result.node.Link)
						// 只处理域名，跳过已是IP的地址
						if hostInfo.Server != "" && !hostInfo.IsIP {
							if resolvedIP, dnsSource := mihomo.ResolveProxyHost(hostInfo.Server); resolvedIP != "" {
								hostMu.Lock()
								hostMappings = append(hostMappings, models.HostMappingInfo{
									Hostname:  hostInfo.Server,
									IP:        resolvedIP,
									NodeName:  result.node.Name,
									Group:     result.node.Group,
									Source:    result.node.Source,
									DNSSource: dnsSource,
								})
								hostMu.Unlock()
							}
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
				// 格式化节点显示项（包含分组和来源信息，方便手机端查看）
				currentItemDisplay := formatNodeDisplayItem(result.node.Name, result.node.Group, result.node.Source)
				tm.UpdateProgress(taskID, totalNodes+currentCompleted, currentItemDisplay, map[string]interface{}{
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
		utils.Info("阶段二完成：速度测试结束")
	}

	// 检查最终是否被取消
	if cancelled || ctx.Err() != nil {
		utils.Info("任务被取消")
		goto applyTags
	}

	// 批量写入所有测速结果到数据库（一次性操作，减少数据库I/O）
	if len(speedTestResults) > 0 {
		if err := models.BatchUpdateSpeedResults(speedTestResults); err != nil {
			utils.Error("批量更新测速结果失败: %v", err)
		} else {
			utils.Debug("批量更新测速结果成功，共 %d 条记录", len(speedTestResults))
		}
	}

	// 批量保存Host映射到数据库（如果开启了持久化）
	if persistHost && len(hostMappings) > 0 {
		// 去重：同一个hostname可能被多个节点使用，只保留第一个
		uniqueHostMappings := make([]models.HostMappingInfo, 0, len(hostMappings))
		seenHostnames := make(map[string]bool)
		for _, m := range hostMappings {
			if !seenHostnames[m.Hostname] {
				seenHostnames[m.Hostname] = true
				uniqueHostMappings = append(uniqueHostMappings, m)
			}
		}
		// 异步保存，不阻塞测速任务完成
		go func(mappings []models.HostMappingInfo) {
			count, err := models.BatchUpsertHosts(mappings)
			if err != nil {
				utils.Error("批量保存Host映射失败: %v", err)
			} else if count > 0 {
				utils.Info("测速Host持久化: 成功处理 %d 条", count)
			}
		}(uniqueHostMappings)
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
		utils.Info("测速任务完成 - 总计: %d, 成功: %d, 失败: %d, 流量: %s", totalNodes, successCount, failCount, formatBytes(trafficTotal))
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
			utils.Warn("获取测速节点最新数据失败: %v, 使用原始数据", err)
			applyAutoTagRules(nodes, "speed_test")
			return
		}

		applyAutoTagRules(updatedNodes, "speed_test")
	}()

	// 测速完成后后台静默刷新机场用量信息（测速会消耗流量）
	go func() {
		node.RefreshUsageForSubscriptionNodes(nodes)
	}()

}
