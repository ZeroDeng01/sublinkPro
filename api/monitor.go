package api

import (
	"sublink/services/monitor"

	"github.com/gin-gonic/gin"
)

// GetSystemStats 获取系统监控统计信息
// @Summary 获取系统监控数据
// @Description 返回系统内存、CPU、Goroutine等运行时统计信息
// @Tags Total
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/total/system-stats [get]
func GetSystemStats(c *gin.Context) {
	stats := monitor.GetSystemStats()
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": stats,
	})
}
