package api

import (
	"errors"
	"net/http"
	"strconv"
	"sublink/database"
	"sublink/models"
	"sublink/services/scheduler"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func smartGroupID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的智能分组ID"})
		return 0, false
	}
	return id, true
}

func smartGroupByID(c *gin.Context) (models.SmartGroup, bool) {
	id, ok := smartGroupID(c)
	if !ok {
		return models.SmartGroup{}, false
	}
	group, err := models.GetSmartGroup(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "智能分组不存在"})
		return models.SmartGroup{}, false
	}
	return group, true
}

func ListSmartGroups(c *gin.Context) {
	groups, err := models.ListSmartGroups()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "获取智能分组失败"})
		return
	}
	for index := range groups {
		groups[index] = scheduler.GetSchedulerManager().SmartGroupCheckInfo(groups[index])
	}
	c.JSON(200, gin.H{"code": 200, "data": groups})
}

func CreateSmartGroup(c *gin.Context) {
	nodeCheckConfigMu.Lock()
	defer nodeCheckConfigMu.Unlock()
	var group models.SmartGroup
	if err := c.ShouldBindBodyWith(&group, binding.JSON); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "无效的分组配置"})
		return
	}
	group.ID = 0
	group.LastCheckAt = nil
	group.MaxAgeHours = maxAgeHoursFromRequest(c, group.MaxAgeHours)
	if err := group.Validate(); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	if group.CheckProfileID > 0 {
		if err := scheduler.ValidateSmartGroupCheckProfile(group, group.CheckProfileID); err != nil {
			c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
			return
		}
	}
	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "名称已存在或保存失败"})
		return
	}
	if err := scheduler.GetSchedulerManager().UpdateSmartGroupJob(group); err != nil {
		_ = database.DB.Delete(&group).Error
		c.JSON(500, gin.H{"code": 500, "msg": "注册定时检测失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": scheduler.GetSchedulerManager().SmartGroupCheckInfo(group)})
}

// New groups default to 72 hours; explicitly passing zero disables expiration.
func maxAgeHoursFromRequest(c *gin.Context, current int) int {
	var fields map[string]any
	if err := c.ShouldBindBodyWith(&fields, binding.JSON); err == nil {
		if _, exists := fields["maxAgeHours"]; !exists {
			return 72
		}
	}
	return current
}

var smartGroupEditableFields = []string{"Name", "Countries", "Keyword", "SourceGroups", "MaxDelay", "MinSpeed", "MaxAgeHours", "CheckProfileID", "ScheduleEnabled", "CronExpr", "CheckAfterAirportUpdate"}

func UpdateSmartGroup(c *gin.Context) {
	nodeCheckConfigMu.Lock()
	defer nodeCheckConfigMu.Unlock()
	existing, ok := smartGroupByID(c)
	if !ok {
		return
	}
	var group models.SmartGroup
	if err := c.ShouldBindBodyWith(&group, binding.JSON); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "无效的分组配置"})
		return
	}
	group.ID = existing.ID
	// Legacy clients omit the new fields; preserve their saved automation settings.
	var fields map[string]any
	if err := c.ShouldBindBodyWith(&fields, binding.JSON); err == nil {
		if _, ok := fields["checkProfileId"]; !ok {
			group.CheckProfileID = existing.CheckProfileID
		}
		if _, ok := fields["scheduleEnabled"]; !ok {
			group.ScheduleEnabled = existing.ScheduleEnabled
		}
		if _, ok := fields["cronExpr"]; !ok {
			group.CronExpr = existing.CronExpr
		}
		if _, ok := fields["checkAfterAirportUpdate"]; !ok {
			group.CheckAfterAirportUpdate = existing.CheckAfterAirportUpdate
		}
	}
	if err := group.Validate(); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	if group.CheckProfileID > 0 {
		if err := scheduler.ValidateSmartGroupCheckProfile(group, group.CheckProfileID); err != nil {
			c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
			return
		}
	}
	if err := database.DB.Model(&models.SmartGroup{}).Where("id = ?", existing.ID).Select(smartGroupEditableFields).Updates(&group).Error; err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "名称已存在或保存失败"})
		return
	}
	if err := scheduler.GetSchedulerManager().UpdateSmartGroupJob(group); err != nil {
		_ = database.DB.Model(&models.SmartGroup{}).Where("id = ?", existing.ID).Select(smartGroupEditableFields).Updates(&existing).Error
		c.JSON(500, gin.H{"code": 500, "msg": "注册定时检测失败"})
		return
	}
	updated, err := models.GetSmartGroup(existing.ID)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "读取更新结果失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": scheduler.GetSchedulerManager().SmartGroupCheckInfo(updated)})
}

func DeleteSmartGroup(c *gin.Context) {
	nodeCheckConfigMu.Lock()
	defer nodeCheckConfigMu.Unlock()
	group, ok := smartGroupByID(c)
	if !ok {
		return
	}
	var jobs int64
	if err := database.DB.Model(&models.NodeCheckJob{}).Where("scope = ? AND smart_group_id = ?", "smart_group", group.ID).Count(&jobs).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "检查自动任务引用失败"})
		return
	}
	if jobs > 0 {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "msg": "该智能分组已被自动检测任务引用，请先更换任务范围"})
		return
	}
	// Do not silently change the output of subscriptions referencing this view.
	var subscriptions []models.Subcription
	if err := database.DB.Select("id", "smart_group_ids").Where("smart_group_ids LIKE ?", "%"+strconv.Itoa(group.ID)+"%").Find(&subscriptions).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "检查订阅引用失败"})
		return
	}
	for _, sub := range subscriptions {
		for _, id := range models.ParseSmartGroupIDs(sub.SmartGroupIDs) {
			if id == group.ID {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "msg": "该智能分组已被订阅引用，请先移除订阅中的选择"})
				return
			}
		}
	}
	if err := database.DB.Delete(&group).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "删除失败"})
		return
	}
	scheduler.GetSchedulerManager().RemoveSmartGroupJob(group.ID)
	c.JSON(200, gin.H{"code": 200, "msg": "删除成功"})
}

func SmartGroupMembers(c *gin.Context) {
	group, ok := smartGroupByID(c)
	if !ok {
		return
	}
	page := 1
	if value := c.Query("page"); value != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil || parsed < 1 {
			c.JSON(400, gin.H{"code": 400, "msg": "页码无效"})
			return
		}
		page = parsed
	}
	pageSize := 30
	if value := c.Query("pageSize"); value != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil || parsed < 1 || parsed > 100 {
			c.JSON(400, gin.H{"code": 400, "msg": "每页数量必须为1至100"})
			return
		}
		pageSize = parsed
	}
	candidates, members, stats, cutoff, err := group.EvaluateMembers()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "查询候选节点失败"})
		return
	}
	airports, err := (&models.Airport{}).List()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "查询来源机场失败"})
		return
	}
	airportNames := make(map[int]string, len(airports))
	for _, airport := range airports {
		airportNames[airport.ID] = airport.Name
	}
	candidateNodes := make([]gin.H, 0, pageSize)
	if page <= (len(candidates)+pageSize-1)/pageSize {
		start := (page - 1) * pageSize
		end := min(start+pageSize, len(candidates))
		for _, node := range candidates[start:end] {
			country, countrySource := models.SmartGroupCountryForDisplay(node)
			reason := group.CandidateExclusionReason(node, cutoff)
			candidateNodes = append(candidateNodes, gin.H{
				"id": node.ID, "name": node.EffectiveName(), "group": node.Group,
				"airportId": node.SourceID, "airportName": airportNames[node.SourceID],
				"country": country, "countrySource": countrySource, "matchSource": group.MatchSource(node),
				"delay": node.DelayTime, "speed": node.Speed,
				"delayStatus": node.DelayStatus, "speedStatus": node.SpeedStatus,
				"reason": reason, "includedBySmartGroup": reason == "",
			})
		}
	}
	ids := make([]int, 0, len(members))
	details := make([]gin.H, 0, min(len(members), 100))
	for _, node := range members {
		ids = append(ids, node.ID)
		if len(details) < 100 {
			country, countrySource := models.SmartGroupCountryForDisplay(node)
			details = append(details, gin.H{"id": node.ID, "name": node.EffectiveName(), "group": node.Group, "country": country, "countrySource": countrySource, "matchSource": group.MatchSource(node), "delay": node.DelayTime, "speed": node.Speed})
		}
	}
	c.JSON(200, gin.H{"code": 200, "data": gin.H{"ids": ids, "count": len(ids), "candidateCount": stats.CandidateCount, "statusCounts": stats, "nodes": details, "candidateNodes": candidateNodes, "page": page, "pageSize": pageSize, "checkInfo": scheduler.GetSchedulerManager().SmartGroupCheckInfo(group)}})
}

func CheckSmartGroupCandidates(c *gin.Context) {
	group, ok := smartGroupByID(c)
	if !ok {
		return
	}
	var req struct {
		ProfileID int `json:"profileId" binding:"required"`
		NodeID    int `json:"nodeId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ProfileID <= 0 {
		c.JSON(400, gin.H{"code": 400, "msg": "请选择节点检测策略"})
		return
	}
	count, err := scheduler.StartSmartGroupCheck(group.ID, req.ProfileID, req.NodeID, models.TaskTriggerManual)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, scheduler.ErrSmartGroupCheckRunning) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"code": status, "msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": gin.H{"count": count}})
}
