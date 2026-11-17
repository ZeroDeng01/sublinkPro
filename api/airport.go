package api

import (
	"errors"
	"strconv"
	"strings"
	"sublink/dto"
	"sublink/models"
	"sublink/node"
	"sublink/services/scheduler"
	"sublink/utils"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

// validateCron 验证5字段Cron表达式
func validateCron(expr string) bool {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(expr)
	return err == nil
}

func normalizeAirportRequestHeaders(headers []dto.AirportRequestHeader) (models.AirportRequestHeaders, error) {
	normalized := make(models.AirportRequestHeaders, 0, len(headers))
	for _, header := range headers {
		key := strings.TrimSpace(header.Key)
		value := strings.TrimSpace(header.Value)

		if key == "" && value == "" {
			continue
		}
		if key == "" {
			return nil, errors.New("custom header name cannot be empty")
		}
		if strings.EqualFold(key, "User-Agent") {
			return nil, errors.New("use the dedicated field to set the User-Agent")
		}

		normalized = append(normalized, models.AirportRequestHeader{
			Key:   key,
			Value: value,
		})
	}
	return normalized, nil
}

// resolveAirportHWID 返回本次保存使用的设备标识。
// 未填写时优先保留原值；仅新建（generate=true）才自动生成，避免编辑时把已有设备换掉。
// 禁用 X-HWID 时仍保留原值，方便之后重新开启。
func resolveAirportHWID(requested, previous string, disable, generate bool) string {
	if hwid := strings.TrimSpace(requested); hwid != "" {
		return hwid
	}
	if previous = strings.TrimSpace(previous); previous != "" {
		return previous
	}
	if disable || !generate {
		return ""
	}
	return utils.GenerateHWID()
}

// AirportWithStats 机场数据（包含节点统计）
type AirportWithStats struct {
	models.Airport
	NodeStats models.AirportNodeStats `json:"nodeStats"`
}

// AirportList 获取机场列表（支持分页和筛选）
func AirportList(c *gin.Context) {
	// 解析分页参数
	page := 0
	pageSize := 0
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// 解析筛选参数
	filter := models.AirportFilter{
		Keyword: c.Query("keyword"),
		Group:   c.Query("group"),
	}

	// 解析启用状态筛选
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		filter.Enabled = new(enabledStr == "true")
	}

	// 分页查询（带筛选）
	if page > 0 && pageSize > 0 {
		airports, total, err := new(models.Airport).ListWithFilter(page, pageSize, filter)
		if err != nil {
			utils.FailWithMsg(c, "Failed to fetch airport list: "+err.Error())
			return
		}

		// 填充节点数量和统计信息
		result := make([]AirportWithStats, len(airports))
		for i := range airports {
			nodes, err := models.ListNodesByAirportID(airports[i].ID)
			if err == nil {
				airports[i].NodeCount = len(nodes)
			}
			result[i] = AirportWithStats{
				Airport:   airports[i],
				NodeStats: models.GetAirportNodeStats(airports[i].ID),
			}
		}

		totalPages := 0
		if pageSize > 0 {
			totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
		}
		utils.OkDetailed(c, "Fetched successfully", gin.H{
			"items":      result,
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		})
		return
	}

	// 不带分页，返回全部（但仍支持筛选）
	airports, _, err := new(models.Airport).ListWithFilter(0, 0, filter)
	if err != nil {
		utils.FailWithMsg(c, "Failed to fetch airport list: "+err.Error())
		return
	}

	// 填充节点数量和统计信息
	result := make([]AirportWithStats, len(airports))
	for i := range airports {
		nodes, err := models.ListNodesByAirportID(airports[i].ID)
		if err == nil {
			airports[i].NodeCount = len(nodes)
		}
		result[i] = AirportWithStats{
			Airport:   airports[i],
			NodeStats: models.GetAirportNodeStats(airports[i].ID),
		}
	}

	utils.OkDetailed(c, "Fetched successfully", result)
}

// AirportGet 获取单个机场详情
func AirportGet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}

	airport, err := models.GetAirportByID(id)
	if err != nil {
		utils.FailWithMsg(c, "Airport not found")
		return
	}

	// 填充节点数量
	nodes, err := models.ListNodesByAirportID(airport.ID)
	if err == nil {
		airport.NodeCount = len(nodes)
	}

	utils.OkDetailed(c, "Fetched successfully", airport)
}

// AirportAdd 添加机场
func AirportAdd(c *gin.Context) {
	var req dto.AirportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters: "+err.Error())
		return
	}

	if !validateCron(req.CronExpr) {
		utils.FailWithMsg(c, "Invalid cron expression format")
		return
	}

	// happ:// 链接先还原为明文订阅地址再保存
	resolvedURL, ok := resolveHappURLValue(c, req.URL)
	if !ok {
		return
	}
	req.URL = resolvedURL

	requestHeaders, err := normalizeAirportRequestHeaders(req.RequestHeaders)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	airport := models.Airport{
		Name:                         req.Name,
		URL:                          req.URL,
		CronExpr:                     req.CronExpr,
		Enabled:                      req.Enabled,
		Group:                        req.Group,
		DownloadWithProxy:            req.DownloadWithProxy,
		ProxyLink:                    req.ProxyLink,
		UserAgent:                    req.UserAgent,
		RequestHeaders:               requestHeaders,
		HWID:                         resolveAirportHWID(req.HWID, "", req.DisableHWID, true),
		DisableHWID:                  req.DisableHWID,
		DeviceProfile:                models.NormalizeAirportDeviceProfile(req.DeviceProfile),
		DeviceOS:                     strings.TrimSpace(req.DeviceOS),
		OSVersion:                    strings.TrimSpace(req.OSVersion),
		DeviceModel:                  strings.TrimSpace(req.DeviceModel),
		AppName:                      strings.TrimSpace(req.AppName),
		FetchUsageInfo:               req.FetchUsageInfo,
		SkipTLSVerify:                req.SkipTLSVerify,
		UpdateAfterDetect:            req.UpdateAfterDetect,
		UpdateAfterDetectProfileID:   req.UpdateAfterDetectProfileID,
		UpdateAfterDetectChangedOnly: req.UpdateAfterDetectChangedOnly,
		Remark:                       req.Remark,
		Logo:                         req.Logo,
		NodeNameWhitelist:            req.NodeNameWhitelist,
		NodeNameBlacklist:            req.NodeNameBlacklist,
		ProtocolWhitelist:            req.ProtocolWhitelist,
		ProtocolBlacklist:            req.ProtocolBlacklist,
		NodeNamePreprocess:           req.NodeNamePreprocess,
		DeduplicationRule:            req.DeduplicationRule,
		NodeNameUniquify:             req.NodeNameUniquify,
		NodeNamePrefix:               req.NodeNamePrefix,
		NodeNameIntraUniquify:        req.NodeNameIntraUniquify,
		AutoFillCountry:              req.AutoFillCountry,
		BackfillExistingCountry:      req.BackfillExistingCountry,
	}

	// 检查是否重复
	if err := airport.Find(); err == nil {
		utils.FailWithMsg(c, "Airport already exists (duplicate name or URL)")
		return
	}

	if err := airport.Add(); err != nil {
		utils.FailWithMsg(c, "Add failed: "+err.Error())
		return
	}

	// 添加定时任务
	if req.Enabled {
		sch := scheduler.GetSchedulerManager()
		_ = sch.AddJob(airport.ID, req.CronExpr, func(id int, url string, name string) {
			scheduler.ExecuteSubscriptionTask(id, url, name)
		}, airport.ID, req.URL, req.Name)
	}

	// 立即执行一次
	if req.Enabled {
		go scheduler.ExecuteSubscriptionTaskWithTrigger(airport.ID, airport.URL, airport.Name, models.TaskTriggerManual)
	}

	utils.OkWithMsg(c, "Added successfully")
}

// AirportUpdate 更新机场
func AirportUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}

	var req dto.AirportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters: "+err.Error())
		return
	}

	if !validateCron(req.CronExpr) {
		utils.FailWithMsg(c, "Invalid cron expression format")
		return
	}

	// happ:// 链接先还原为明文订阅地址再保存
	resolvedURL, ok := resolveHappURLValue(c, req.URL)
	if !ok {
		return
	}
	req.URL = resolvedURL

	requestHeaders, err := normalizeAirportRequestHeaders(req.RequestHeaders)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	// 检查是否存在
	existing, err := models.GetAirportByID(id)
	if err != nil {
		utils.FailWithMsg(c, "Airport not found")
		return
	}

	conflict, err := models.HasAirportIdentityConflict(id, req.Name, req.URL)
	if err != nil {
		utils.FailWithMsg(c, "Update failed")
		return
	}
	if conflict {
		utils.FailWithMsg(c, "Airport already exists (name or URL duplicates another airport)")
		return
	}

	// 更新机场
	existing.Name = req.Name
	existing.URL = req.URL
	existing.CronExpr = req.CronExpr
	existing.Enabled = req.Enabled
	existing.Group = req.Group
	existing.DownloadWithProxy = req.DownloadWithProxy
	existing.ProxyLink = req.ProxyLink
	existing.UserAgent = req.UserAgent
	existing.RequestHeaders = requestHeaders
	existing.HWID = resolveAirportHWID(req.HWID, existing.HWID, req.DisableHWID, false)
	existing.DisableHWID = req.DisableHWID
	existing.DeviceProfile = models.NormalizeAirportDeviceProfile(req.DeviceProfile)
	existing.DeviceOS = strings.TrimSpace(req.DeviceOS)
	existing.OSVersion = strings.TrimSpace(req.OSVersion)
	existing.DeviceModel = strings.TrimSpace(req.DeviceModel)
	existing.AppName = strings.TrimSpace(req.AppName)
	existing.FetchUsageInfo = req.FetchUsageInfo
	existing.SkipTLSVerify = req.SkipTLSVerify
	existing.UpdateAfterDetect = req.UpdateAfterDetect
	existing.UpdateAfterDetectProfileID = req.UpdateAfterDetectProfileID
	existing.UpdateAfterDetectChangedOnly = req.UpdateAfterDetectChangedOnly
	existing.Remark = req.Remark
	existing.Logo = req.Logo
	existing.NodeNameWhitelist = req.NodeNameWhitelist
	existing.NodeNameBlacklist = req.NodeNameBlacklist
	existing.ProtocolWhitelist = req.ProtocolWhitelist
	existing.ProtocolBlacklist = req.ProtocolBlacklist
	existing.NodeNamePreprocess = req.NodeNamePreprocess
	existing.DeduplicationRule = req.DeduplicationRule
	existing.NodeNameUniquify = req.NodeNameUniquify
	existing.NodeNamePrefix = req.NodeNamePrefix
	existing.NodeNameIntraUniquify = req.NodeNameIntraUniquify
	existing.AutoFillCountry = req.AutoFillCountry
	existing.BackfillExistingCountry = req.BackfillExistingCountry

	if err := existing.Update(); err != nil {
		utils.FailWithMsg(c, "Update failed: "+err.Error())
		return
	}

	// 同步更新关联节点的来源名称和分组
	if err := models.UpdateNodesByAirportID(id, req.Name, req.Group); err != nil {
		// 记录错误但不阻断流程
		utils.Warn("Failed to update linked nodes: %v", err)
	}

	// 更新定时任务
	sch := scheduler.GetSchedulerManager()
	_ = sch.UpdateJob(id, req.CronExpr, req.Enabled, req.URL, req.Name)

	utils.OkWithMsg(c, "Updated successfully")
}

// AirportBatchUpdate 批量更新机场的调度和分组
func AirportBatchUpdate(c *gin.Context) {
	var req dto.AirportBatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "Select airports to modify")
		return
	}
	if !req.ApplyGroup && !req.ApplySchedule {
		utils.FailWithMsg(c, "Select at least one field to modify")
		return
	}
	if req.ApplySchedule {
		req.CronExpr = strings.TrimSpace(req.CronExpr)
		if req.CronExpr == "" {
			utils.FailWithMsg(c, "Enter a cron expression")
			return
		}
		if !validateCron(req.CronExpr) {
			utils.FailWithMsg(c, "Invalid cron expression format")
			return
		}
	}

	updatedAirports, err := models.BatchUpdateAirports(req.IDs, models.AirportBatchUpdateParams{
		ApplyGroup:    req.ApplyGroup,
		Group:         req.Group,
		ApplySchedule: req.ApplySchedule,
		CronExpr:      req.CronExpr,
	})
	if err != nil {
		utils.FailWithMsg(c, "Batch update failed: "+err.Error())
		return
	}

	if req.ApplySchedule {
		sch := scheduler.GetSchedulerManager()
		for _, airport := range updatedAirports {
			if err := sch.UpdateJob(airport.ID, airport.CronExpr, airport.Enabled, airport.URL, airport.Name); err != nil {
				utils.FailWithMsg(c, "Batch update succeeded, but refreshing the schedule failed: "+err.Error())
				return
			}
		}
	}

	utils.OkDetailed(c, "Batch update succeeded", gin.H{
		"count": len(updatedAirports),
	})
}

// AirportDelete 删除机场
func AirportDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "invalid airport id")
		return
	}

	// 代理订阅直接引用机场；被引用时阻止删除，避免订阅静默失效。
	proxySubCount, err := models.CountProxySubcriptionsByAirport(id)
	if err != nil {
		utils.FailWithMsg(c, "failed to check proxy subscription references")
		return
	}
	if proxySubCount > 0 {
		utils.FailWithI18n(c,
			"This airport is used by "+strconv.Itoa(int(proxySubCount))+" proxy subscription(s). Update or delete those subscriptions first.",
			"backend.airports.inUseByProxySubscriptions",
			map[string]any{"n": proxySubCount},
		)
		return
	}

	// 检查是否需要同时删除关联节点
	deleteNodes := c.Query("deleteNodes") == "true"
	if deleteNodes {
		if err := models.DeleteAirportNodes(id); err != nil {
			utils.FailWithMsg(c, "failed to delete linked nodes")
			return
		}
	}

	airport := &models.Airport{}
	airport.ID = id
	if err := airport.Del(); err != nil {
		utils.FailWithMsg(c, "delete failed")
		return
	}

	// 删除定时任务
	sch := scheduler.GetSchedulerManager()
	sch.RemoveJob(id)

	utils.OkWithMsg(c, "airport deleted")
}

// AirportPull 手动拉取机场订阅
func AirportPull(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}

	airport, err := models.GetAirportByID(id)
	if err != nil {
		utils.FailWithMsg(c, "Airport not found")
		return
	}

	// 异步执行拉取任务
	go scheduler.ExecuteSubscriptionTaskWithTrigger(airport.ID, airport.URL, airport.Name, models.TaskTriggerManual)

	utils.OkWithMsg(c, "Task submitted; refresh later to view the result")
}

// AirportPullAll 批量拉取机场订阅
// 支持两种模式：
// 1. 提供 ids 参数：拉取指定的机场（不论启用状态）
// 2. 不提供 ids 参数：拉取所有已启用的机场
func AirportPullAll(c *gin.Context) {
	var req dto.AirportPullAllRequest
	// 尝试解析请求体，如果没有请求体或解析失败，req.IDs 为 nil/empty
	_ = c.ShouldBindJSON(&req)

	var airportsToProcess []models.Airport

	if len(req.IDs) > 0 {
		// 模式1：拉取指定的机场（用户明确选择，不论启用状态）
		for _, id := range req.IDs {
			airport, err := models.GetAirportByID(id)
			if err != nil {
				// 跳过不存在的机场，继续处理其他机场
				utils.Warn("Airport ID %d does not exist; skipped", id)
				continue
			}
			airportsToProcess = append(airportsToProcess, *airport)
		}
	} else {
		// 模式2：拉取所有已启用的机场
		airportModel := models.Airport{}
		allAirports, err := airportModel.List()
		if err != nil {
			utils.FailWithMsg(c, "Failed to fetch airport list")
			return
		}

		for _, airport := range allAirports {
			if airport.Enabled {
				airportsToProcess = append(airportsToProcess, airport)
			}
		}
	}

	if len(airportsToProcess) == 0 {
		if len(req.IDs) > 0 {
			utils.OkWithMsg(c, "No valid airports found")
		} else {
			utils.OkWithMsg(c, "No enabled airports")
		}
		return
	}

	// 提交拉取任务
	count := 0
	for _, airport := range airportsToProcess {
		go scheduler.ExecuteSubscriptionTaskWithTrigger(airport.ID, airport.URL, airport.Name, models.TaskTriggerManual)
		count++
	}

	utils.OkWithData(c, map[string]any{
		"message": "Batch fetch task submitted",
		"count":   count,
	})
}

// AirportRefreshUsage 仅刷新机场的用量信息，不更新订阅/节点
func AirportRefreshUsage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "invalid airport id")
		return
	}

	airport, err := models.GetAirportByID(id)
	if err != nil {
		utils.FailWithMsg(c, "airport not found")
		return
	}

	if !airport.FetchUsageInfo {
		utils.FailWithMsg(c, "usage info fetching is disabled for this airport")
		return
	}

	// 同步获取用量信息
	usageInfo, err := node.UpdateAirportUsageInfo(id)
	if err != nil {
		utils.FailWithMsg(c, "failed to fetch usage info: "+err.Error())
		return
	}

	// 返回用量信息
	utils.OkDetailed(c, "usage info refreshed", map[string]any{
		"upload":   usageInfo.Upload,
		"download": usageInfo.Download,
		"total":    usageInfo.Total,
		"expire":   usageInfo.Expire,
	})
}
