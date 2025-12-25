package api

import (
	"errors"
	"strconv"
	"sublink/dto"
	"sublink/models"
	"sublink/node"
	"sublink/services/scheduler"
	"sublink/utils"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// validateCron 验证5字段Cron表达式
func validateCron(expr string) bool {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(expr)
	return err == nil
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
		enabled := enabledStr == "true"
		filter.Enabled = &enabled
	}

	// 分页查询（带筛选）
	if page > 0 && pageSize > 0 {
		airports, total, err := new(models.Airport).ListWithFilter(page, pageSize, filter)
		if err != nil {
			utils.FailWithMsg(c, "获取机场列表失败: "+err.Error())
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
		utils.OkDetailed(c, "获取成功", gin.H{
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
		utils.FailWithMsg(c, "获取机场列表失败: "+err.Error())
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

	utils.OkDetailed(c, "获取成功", result)
}

// AirportGet 获取单个机场详情
func AirportGet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	airport, err := models.GetAirportByID(id)
	if err != nil {
		utils.FailWithMsg(c, "机场不存在")
		return
	}

	// 填充节点数量
	nodes, err := models.ListNodesByAirportID(airport.ID)
	if err == nil {
		airport.NodeCount = len(nodes)
	}

	utils.OkDetailed(c, "获取成功", airport)
}

// AirportAdd 添加机场
func AirportAdd(c *gin.Context) {
	var req dto.AirportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误: "+err.Error())
		return
	}

	if !validateCron(req.CronExpr) {
		utils.FailWithMsg(c, "Cron表达式格式错误")
		return
	}

	airport := models.Airport{
		Name:              req.Name,
		URL:               req.URL,
		CronExpr:          req.CronExpr,
		Enabled:           req.Enabled,
		Group:             req.Group,
		DownloadWithProxy: req.DownloadWithProxy,
		ProxyLink:         req.ProxyLink,
		UserAgent:         req.UserAgent,
		FetchUsageInfo:    req.FetchUsageInfo,
		SkipTLSVerify:     req.SkipTLSVerify,
		Remark:            req.Remark,
		Logo:              req.Logo,
	}

	// 检查是否重复
	if err := airport.Find(); err == nil {
		utils.FailWithMsg(c, "机场已存在（名称或URL重复）")
		return
	}

	if err := airport.Add(); err != nil {
		utils.FailWithMsg(c, "添加失败: "+err.Error())
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

	utils.OkWithMsg(c, "添加成功")
}

// AirportUpdate 更新机场
func AirportUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	var req dto.AirportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误: "+err.Error())
		return
	}

	if !validateCron(req.CronExpr) {
		utils.FailWithMsg(c, "Cron表达式格式错误")
		return
	}

	// 检查是否存在
	existing, err := models.GetAirportByID(id)
	if err != nil {
		utils.FailWithMsg(c, "机场不存在")
		return
	}

	// 检查名称/URL是否与其他机场冲突
	checkAirport := models.Airport{Name: req.Name, URL: req.URL}
	if err := checkAirport.Find(); err == nil && checkAirport.ID != id {
		utils.FailWithMsg(c, "机场已存在（名称或URL与其他机场重复）")
		return
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.FailWithMsg(c, "更新失败")
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
	existing.FetchUsageInfo = req.FetchUsageInfo
	existing.SkipTLSVerify = req.SkipTLSVerify
	existing.Remark = req.Remark
	existing.Logo = req.Logo

	if err := existing.Update(); err != nil {
		utils.FailWithMsg(c, "更新失败: "+err.Error())
		return
	}

	// 同步更新关联节点的来源名称和分组
	if err := models.UpdateNodesByAirportID(id, req.Name, req.Group); err != nil {
		// 记录错误但不阻断流程
		utils.Warn("更新关联节点失败: %v", err)
	}

	// 更新定时任务
	sch := scheduler.GetSchedulerManager()
	_ = sch.UpdateJob(id, req.CronExpr, req.Enabled, req.URL, req.Name)

	utils.OkWithMsg(c, "更新成功")
}

// AirportDelete 删除机场
func AirportDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	// 检查是否需要同时删除关联节点
	deleteNodes := c.Query("deleteNodes") == "true"
	if deleteNodes {
		if err := models.DeleteAirportNodes(id); err != nil {
			utils.FailWithMsg(c, "删除关联节点失败")
			return
		}
	}

	airport := &models.Airport{}
	airport.ID = id
	if err := airport.Del(); err != nil {
		utils.FailWithMsg(c, "删除失败")
		return
	}

	// 删除定时任务
	sch := scheduler.GetSchedulerManager()
	sch.RemoveJob(id)

	utils.OkWithMsg(c, "删除成功")
}

// AirportPull 手动拉取机场订阅
func AirportPull(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	airport, err := models.GetAirportByID(id)
	if err != nil {
		utils.FailWithMsg(c, "机场不存在")
		return
	}

	// 异步执行拉取任务
	go scheduler.ExecuteSubscriptionTaskWithTrigger(airport.ID, airport.URL, airport.Name, models.TaskTriggerManual)

	utils.OkWithMsg(c, "任务已提交，请稍后刷新查看结果")
}

// AirportRefreshUsage 仅刷新机场的用量信息，不更新订阅/节点
func AirportRefreshUsage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	airport, err := models.GetAirportByID(id)
	if err != nil {
		utils.FailWithMsg(c, "机场不存在")
		return
	}

	if !airport.FetchUsageInfo {
		utils.FailWithMsg(c, "该机场未开启用量信息获取")
		return
	}

	// 同步获取用量信息
	usageInfo, err := node.UpdateAirportUsageInfo(id)
	if err != nil {
		utils.FailWithMsg(c, "获取用量信息失败: "+err.Error())
		return
	}

	// 返回用量信息
	utils.OkDetailed(c, "用量信息已更新", map[string]interface{}{
		"upload":   usageInfo.Upload,
		"download": usageInfo.Download,
		"total":    usageInfo.Total,
		"expire":   usageInfo.Expire,
	})
}
