package api

import (
	"fmt"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// parseUsageHistoryParams 解析用量历史查询参数。
func parseUsageHistoryParams(c *gin.Context) (days int, granularity string) {
	if raw := strings.TrimSpace(c.Query("days")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			days = value
		}
	}
	return days, strings.TrimSpace(c.Query("granularity"))
}

// parseUsageTZOffset 解析客户端时区偏移（分钟，与浏览器 Date.getTimezoneOffset() 一致），
// 让“今日”与按天统计跟随客户端本地自然日。返回 nil 表示使用服务器本地时区。
func parseUsageTZOffset(c *gin.Context) (*time.Location, error) {
	raw := strings.TrimSpace(c.Query("tzOffset"))
	if raw == "" {
		return nil, nil
	}
	minutes, err := strconv.Atoi(raw)
	if err != nil || minutes < -14*60 || minutes > 14*60 {
		return nil, fmt.Errorf("invalid tzOffset: %s", raw)
	}
	return time.FixedZone("client", -minutes*60), nil
}

// parseAirportIDsQuery 解析逗号分隔的机场 ID 列表。
func parseAirportIDsQuery(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	seen := make(map[int]struct{}, len(parts))
	ids := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.Atoi(part)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("invalid airport ID: %s", part)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

// GetAirportUsageHistory 获取单个机场的用量历史。
func GetAirportUsageHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.FailWithMsg(c, "invalid airport id")
		return
	}

	days, granularity := parseUsageHistoryParams(c)
	history, err := models.GetAirportUsageHistory(id, days, granularity)
	if err != nil {
		utils.FailWithMsg(c, "failed to load usage history: "+err.Error())
		return
	}
	utils.OkDetailed(c, "success", history)
}

// GetAirportUsageSparklines 批量获取机场用量迷你图数据。
func GetAirportUsageSparklines(c *gin.Context) {
	ids, err := parseAirportIDsQuery(c.Query("ids"))
	if err != nil {
		utils.FailWithMsg(c, "invalid params: "+err.Error())
		return
	}
	if len(ids) == 0 {
		utils.FailWithMsg(c, "invalid params: ids is required")
		return
	}

	days, _ := parseUsageHistoryParams(c)
	loc, err := parseUsageTZOffset(c)
	if err != nil {
		utils.FailWithMsg(c, "invalid params: "+err.Error())
		return
	}
	sparklines, err := models.GetAirportUsageSparklines(ids, days, loc)
	if err != nil {
		utils.FailWithMsg(c, "failed to load usage sparklines: "+err.Error())
		return
	}
	utils.OkDetailed(c, "success", gin.H{"items": sparklines})
}

// GetAirportDailyUsage 获取按天统计的流量消耗。
// airportId 省略时汇总全部机场；当天的消耗即报告中的 todayUsed。
func GetAirportDailyUsage(c *gin.Context) {
	days, _ := parseUsageHistoryParams(c)
	loc, err := parseUsageTZOffset(c)
	if err != nil {
		utils.FailWithMsg(c, "invalid params: "+err.Error())
		return
	}

	airportID := 0
	if raw := strings.TrimSpace(c.Query("airportId")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			utils.FailWithMsg(c, "invalid airportId")
			return
		}
		airportID = value
	}

	report, err := models.GetAirportDailyUsage(days, airportID, loc)
	if err != nil {
		utils.FailWithMsg(c, "failed to load daily usage: "+err.Error())
		return
	}
	utils.OkDetailed(c, "success", report)
}
