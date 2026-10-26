package api

import (
	"strconv"
	"sublink/models"
	"sublink/services/scheduler"
	"sublink/services/unlock"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

func GetNodeCheckMeta(c *gin.Context) {
	providerValues := unlock.ListRegisteredUnlockProviders()
	providerOptions := make([]models.UnlockProviderMeta, 0, len(providerValues))
	for _, provider := range providerValues {
		providerOptions = append(providerOptions, models.GetUnlockProviderMeta(provider))
	}

	utils.OkDetailed(c, "获取成功", gin.H{
		"unlockProviders":       providerOptions,
		"unlockStatuses":        models.GetUnlockStatusMetas(),
		"unlockRenameVariables": models.BuildUnlockRenameVariables(providerValues),
		"conditionFields":       models.GetNodeConditionFields(),
	})
}

// ListNodeCheckProfiles 获取节点检测策略列表
// GET /api/v1/node-check/profiles
func ListNodeCheckProfiles(c *gin.Context) {
	var profile models.NodeCheckProfile
	profiles, err := profile.List()
	if err != nil {
		utils.FailWithMsg(c, "获取策略列表失败")
		return
	}
	utils.OkDetailed(c, "获取成功", profiles)
}

// GetNodeCheckProfile 获取单个策略
// GET /api/v1/node-check/profiles/:id
func GetNodeCheckProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.FailWithMsg(c, "无效的策略ID")
		return
	}

	profile, err := models.GetNodeCheckProfileByID(id)
	if err != nil {
		utils.FailWithMsg(c, "策略不存在")
		return
	}
	utils.OkDetailed(c, "获取成功", profile)
}

// CreateNodeCheckProfile 创建策略
// POST /api/v1/node-check/profiles
func CreateNodeCheckProfile(c *gin.Context) {
	var req struct {
		Name                string   `json:"name" binding:"required"`
		Enabled             bool     `json:"enabled"`
		CronExpr            string   `json:"cronExpr"`
		Mode                string   `json:"mode"`
		TestURL             string   `json:"testUrl"`
		LatencyURL          string   `json:"latencyUrl"`
		Timeout             int      `json:"timeout"`
		Groups              []string `json:"groups"`
		Tags                []string `json:"tags"`
		LatencyConcurrency  int      `json:"latencyConcurrency"`
		SpeedConcurrency    int      `json:"speedConcurrency"`
		DetectCountry       bool     `json:"detectCountry"`
		LandingIPURL        string   `json:"landingIpUrl"`
		IncludeHandshake    *bool    `json:"includeHandshake"`
		SpeedRecordMode     string   `json:"speedRecordMode"`
		PeakSampleInterval  int      `json:"peakSampleInterval"`
		TrafficByGroup      *bool    `json:"trafficByGroup"`
		TrafficBySource     *bool    `json:"trafficBySource"`
		TrafficByNode       *bool    `json:"trafficByNode"`
		PreserveSpeedResult bool     `json:"preserveSpeedResult"`
		DetectQuality       bool     `json:"detectQuality"`
		QualityCheckURL     string   `json:"qualityCheckUrl"`
		DetectUnlock        bool     `json:"detectUnlock"`
		UnlockProviders     []string `json:"unlockProviders"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误: "+err.Error())
		return
	}

	// 检查名称是否重复
	if existing, _ := models.FindNodeCheckProfileByName(req.Name); existing != nil {
		utils.FailWithMsg(c, "策略名称已存在")
		return
	}

	// 设置默认值
	mode := req.Mode
	if mode == "" {
		mode = "tcp"
	}
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 5
	}
	// speedConcurrency: 0=智能动态模式，>0=固定并发数
	// 负数视为无效，使用默认值0（智能动态）
	speedConcurrency := req.SpeedConcurrency
	if speedConcurrency < 0 {
		speedConcurrency = 0
	}
	speedRecordMode := req.SpeedRecordMode
	if speedRecordMode == "" {
		speedRecordMode = "average"
	}
	peakSampleInterval := req.PeakSampleInterval
	if peakSampleInterval < 50 || peakSampleInterval > 200 {
		peakSampleInterval = 100
	}
	includeHandshake := true
	if req.IncludeHandshake != nil {
		includeHandshake = *req.IncludeHandshake
	}
	trafficByGroup := true
	if req.TrafficByGroup != nil {
		trafficByGroup = *req.TrafficByGroup
	}
	trafficBySource := true
	if req.TrafficBySource != nil {
		trafficBySource = *req.TrafficBySource
	}
	trafficByNode := false
	if req.TrafficByNode != nil {
		trafficByNode = *req.TrafficByNode
	}

	profile := models.NodeCheckProfile{
		Name:                req.Name,
		Enabled:             req.Enabled,
		CronExpr:            req.CronExpr,
		Mode:                mode,
		TestURL:             req.TestURL,
		LatencyURL:          req.LatencyURL,
		Timeout:             timeout,
		LatencyConcurrency:  req.LatencyConcurrency,
		SpeedConcurrency:    speedConcurrency,
		DetectCountry:       req.DetectCountry,
		LandingIPURL:        req.LandingIPURL,
		IncludeHandshake:    includeHandshake,
		SpeedRecordMode:     speedRecordMode,
		PeakSampleInterval:  peakSampleInterval,
		TrafficByGroup:      trafficByGroup,
		TrafficBySource:     trafficBySource,
		TrafficByNode:       trafficByNode,
		PreserveSpeedResult: req.PreserveSpeedResult,
		DetectQuality:       req.DetectQuality,
		QualityCheckURL:     req.QualityCheckURL,
		DetectUnlock:        req.DetectUnlock,
	}
	profile.SetGroups(req.Groups)
	profile.SetTags(req.Tags)
	profile.SetUnlockProviders(models.NormalizeUnlockProviders(req.UnlockProviders))

	if err := validateCheckProfileConfig(profile); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if err := profile.Add(); err != nil {
		utils.FailWithMsg(c, "创建策略失败")
		return
	}

	// 如果启用了定时任务，注册到调度器
	if profile.Enabled && profile.CronExpr != "" {
		sch := scheduler.GetSchedulerManager()
		if err := sch.AddNodeCheckProfileJob(profile.ID, profile.CronExpr); err != nil {
			utils.Warn("注册节点检测定时任务失败: %v", err)
		}
	}

	utils.OkDetailed(c, "创建成功", profile)
}

// UpdateNodeCheckProfile 更新策略
// PUT /api/v1/node-check/profiles/:id
func UpdateNodeCheckProfile(c *gin.Context) {
	nodeCheckConfigMu.Lock()
	defer nodeCheckConfigMu.Unlock()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.FailWithMsg(c, "无效的策略ID")
		return
	}

	var req struct {
		Name                string    `json:"name"`
		Enabled             *bool     `json:"enabled"`
		CronExpr            *string   `json:"cronExpr"`
		Mode                string    `json:"mode"`
		TestURL             string    `json:"testUrl"`
		LatencyURL          string    `json:"latencyUrl"`
		Timeout             int       `json:"timeout"`
		Groups              *[]string `json:"groups"`
		Tags                *[]string `json:"tags"`
		LatencyConcurrency  int       `json:"latencyConcurrency"`
		SpeedConcurrency    int       `json:"speedConcurrency"`
		DetectCountry       bool      `json:"detectCountry"`
		LandingIPURL        string    `json:"landingIpUrl"`
		IncludeHandshake    *bool     `json:"includeHandshake"`
		SpeedRecordMode     string    `json:"speedRecordMode"`
		PeakSampleInterval  int       `json:"peakSampleInterval"`
		TrafficByGroup      *bool     `json:"trafficByGroup"`
		TrafficBySource     *bool     `json:"trafficBySource"`
		TrafficByNode       *bool     `json:"trafficByNode"`
		PreserveSpeedResult *bool     `json:"preserveSpeedResult"`
		DetectQuality       bool      `json:"detectQuality"`
		QualityCheckURL     string    `json:"qualityCheckUrl"`
		DetectUnlock        bool      `json:"detectUnlock"`
		UnlockProviders     []string  `json:"unlockProviders"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	profile, err := models.GetNodeCheckProfileByID(id)
	if err != nil {
		utils.FailWithMsg(c, "策略不存在")
		return
	}

	// 检查名称是否与其他策略重复
	if req.Name != "" && req.Name != profile.Name {
		if existing, _ := models.FindNodeCheckProfileByName(req.Name); existing != nil && existing.ID != id {
			utils.FailWithMsg(c, "策略名称已存在")
			return
		}
		profile.Name = req.Name
	}

	// 更新字段
	if req.Enabled != nil {
		profile.Enabled = *req.Enabled
	}
	if req.CronExpr != nil {
		profile.CronExpr = *req.CronExpr
	}
	if req.Mode != "" {
		profile.Mode = req.Mode
	}
	profile.TestURL = req.TestURL
	profile.LatencyURL = req.LatencyURL
	if req.Timeout > 0 {
		profile.Timeout = req.Timeout
	}
	if req.Groups != nil {
		profile.SetGroups(*req.Groups)
	}
	if req.Tags != nil {
		profile.SetTags(*req.Tags)
	}
	profile.LatencyConcurrency = req.LatencyConcurrency
	// speedConcurrency: 0=智能动态模式，>=0 的值都应保存
	if req.SpeedConcurrency >= 0 {
		profile.SpeedConcurrency = req.SpeedConcurrency
	}
	profile.DetectCountry = req.DetectCountry
	profile.LandingIPURL = req.LandingIPURL
	if req.IncludeHandshake != nil {
		profile.IncludeHandshake = *req.IncludeHandshake
	}
	if req.SpeedRecordMode != "" {
		profile.SpeedRecordMode = req.SpeedRecordMode
	}
	if req.PeakSampleInterval >= 50 && req.PeakSampleInterval <= 200 {
		profile.PeakSampleInterval = req.PeakSampleInterval
	}
	if req.TrafficByGroup != nil {
		profile.TrafficByGroup = *req.TrafficByGroup
	}
	if req.TrafficBySource != nil {
		profile.TrafficBySource = *req.TrafficBySource
	}
	if req.TrafficByNode != nil {
		profile.TrafficByNode = *req.TrafficByNode
	}
	if req.PreserveSpeedResult != nil {
		profile.PreserveSpeedResult = *req.PreserveSpeedResult
	}
	profile.DetectQuality = req.DetectQuality
	profile.QualityCheckURL = req.QualityCheckURL
	profile.DetectUnlock = req.DetectUnlock
	profile.SetUnlockProviders(models.NormalizeUnlockProviders(req.UnlockProviders))

	if err := validateCheckProfileConfig(*profile); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if err := profile.Update(); err != nil {
		utils.FailWithMsg(c, "更新策略失败")
		return
	}

	// 更新调度器任务
	sch := scheduler.GetSchedulerManager()
	if err := sch.UpdateNodeCheckProfileJob(profile.ID, profile.CronExpr, profile.Enabled); err != nil {
		utils.Warn("更新节点检测定时任务失败: %v", err)
	}

	utils.OkDetailed(c, "更新成功", profile)
}

// DeleteNodeCheckProfile 删除策略
// DELETE /api/v1/node-check/profiles/:id
func DeleteNodeCheckProfile(c *gin.Context) {
	nodeCheckConfigMu.Lock()
	defer nodeCheckConfigMu.Unlock()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.FailWithMsg(c, "无效的策略ID")
		return
	}

	profile, err := models.GetNodeCheckProfileByID(id)
	if err != nil {
		utils.FailWithMsg(c, "策略不存在")
		return
	}

	if used, err := hasNodeCheckProfileDependents(id); err != nil {
		utils.FailWithMsg(c, "检查自动任务或机场引用失败")
		return
	} else if used {
		utils.FailWithMsg(c, "此方案被自动任务或机场引用，请先更换引用方案")
		return
	}
	// 从调度器移除任务
	sch := scheduler.GetSchedulerManager()
	sch.RemoveNodeCheckProfileJob(profile.ID)

	if err := profile.Del(); err != nil {
		utils.FailWithMsg(c, "删除策略失败")
		return
	}

	utils.OkWithMsg(c, "删除成功")
}

// RunNodeCheck 执行节点检测
// POST /api/v1/node-check/run
func RunNodeCheck(c *gin.Context) {
	var req struct {
		ProfileID int      `json:"profileId" binding:"required"` // 策略ID（必填）
		Scope     string   `json:"scope"`
		Groups    []string `json:"groups"`
		Tags      []string `json:"tags"`
		NodeIDs   []int    `json:"nodeIds"` // 节点ID列表（可选，空表示按策略范围执行）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误：必须指定检测策略")
		return
	}

	if req.ProfileID <= 0 {
		utils.FailWithMsg(c, "参数错误：必须指定有效的检测策略ID")
		return
	}

	// 验证策略存在
	if _, err := models.GetNodeCheckProfileByID(req.ProfileID); err != nil {
		utils.FailWithMsg(c, "策略不存在")
		return
	}

	if req.Scope == "filtered" {
		scope := models.NodeCheckJob{Name: "manual", ProfileID: req.ProfileID, Trigger: "airport_update", Scope: "filtered", Groups: req.Groups, Tags: req.Tags}
		if err := scope.Validate(); err != nil {
			utils.FailWithMsg(c, err.Error())
			return
		}
	}
	if req.Scope != "" {
		if req.Scope == "nodes" {
			if len(req.NodeIDs) == 0 {
				utils.FailWithMsg(c, "请选择至少一个节点")
				return
			}
		} else {
			if req.Scope != "all" && req.Scope != "filtered" {
				utils.FailWithMsg(c, "无效的检测范围")
				return
			}
			if req.Scope == "filtered" && len(req.Groups) == 0 && len(req.Tags) == 0 {
				utils.FailWithMsg(c, "筛选范围不能为空")
				return
			}
			if req.Scope == "all" {
				req.Groups = nil
				req.Tags = nil
			}
			nodes, err := scheduler.ResolveNodeCheckJobNodes(models.NodeCheckJob{Scope: req.Scope, Groups: req.Groups, Tags: req.Tags}, nil)
			if err != nil || len(nodes) == 0 {
				utils.FailWithMsg(c, "当前范围没有可检测节点")
				return
			}
			profile, err := models.GetNodeCheckProfileByID(req.ProfileID)
			if err != nil {
				utils.FailWithMsg(c, "检测方案已不存在")
				return
			}
			go func() {
				if err := scheduler.RunSpeedTestWithConfig(nodes, models.TaskTriggerManual, profile.Name, scheduler.SpeedTestConfigFromProfile(profile)); err != nil {
					utils.Warn("手动检测失败: %v", err)
				}
			}()
			utils.OkWithMsg(c, "节点检测任务已启动")
			return
		}
	}
	if len(req.NodeIDs) > 0 {
		seen := make(map[int]bool, len(req.NodeIDs))
		ids := make([]int, 0, len(req.NodeIDs))
		for _, id := range req.NodeIDs {
			if id <= 0 {
				utils.FailWithMsg(c, "无效的节点ID")
				return
			}
			if !seen[id] {
				if _, ok := models.GetNodeByID(id); !ok {
					utils.FailWithMsg(c, "选中的节点已不存在，请刷新后重试")
					return
				}
				ids = append(ids, id)
				seen[id] = true
			}
		}
		req.NodeIDs = ids
	}
	// 使用指定策略执行（手动触发）
	go scheduler.ExecuteNodeCheckWithProfile(req.ProfileID, req.NodeIDs, models.TaskTriggerManual)
	utils.OkWithMsg(c, "节点检测任务已启动")
}

// RunNodeCheckWithProfile 使用指定策略执行节点检测
// POST /api/v1/node-check/profiles/:id/run
func RunNodeCheckWithProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.FailWithMsg(c, "无效的策略ID")
		return
	}

	// 验证策略存在
	if _, err := models.GetNodeCheckProfileByID(id); err != nil {
		utils.FailWithMsg(c, "策略不存在")
		return
	}

	go scheduler.ExecuteNodeCheckWithProfile(id, nil, models.TaskTriggerManual)
	utils.OkWithMsg(c, "节点检测任务已启动")
}
