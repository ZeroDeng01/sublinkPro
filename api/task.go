package api

import (
	"strconv"
	"sublink/models"
	"sublink/services"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// GetTasks 获取任务列表
func GetTasks(c *gin.Context) {
	// 解析过滤参数
	filter := models.TaskFilter{
		Status:  c.Query("status"),
		Type:    c.Query("type"),
		Trigger: c.Query("trigger"),
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取任务列表
	tasks, total, err := models.ListTasks(filter, page, pageSize)
	if err != nil {
		utils.FailWithMsg(c, "获取任务列表失败")
		return
	}

	// 获取运行中任务的实时状态
	runningTasks := services.GetTaskManager().GetRunningTasksInfo()

	// 合并实时状态：将内存中的运行任务状态合并到数据库查询结果中
	runningMap := make(map[string]models.Task)
	for _, t := range runningTasks {
		runningMap[t.ID] = t
	}

	// 更新列表中运行任务的实时状态
	for i := range tasks {
		if running, ok := runningMap[tasks[i].ID]; ok {
			tasks[i].Progress = running.Progress
			tasks[i].CurrentItem = running.CurrentItem
			tasks[i].Status = running.Status
		}
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	utils.OkDetailed(c, "获取成功", gin.H{
		"items":      tasks,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	})
}

// GetTask 获取单个任务详情
func GetTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.FailWithMsg(c, "任务ID不能为空")
		return
	}

	var task models.Task
	if err := task.GetByID(taskID); err != nil {
		utils.FailWithMsg(c, "任务不存在")
		return
	}

	// 如果是运行中的任务，获取实时状态
	runningTasks := services.GetTaskManager().GetRunningTasksInfo()
	for _, t := range runningTasks {
		if t.ID == taskID {
			task.Progress = t.Progress
			task.CurrentItem = t.CurrentItem
			task.Status = t.Status
			break
		}
	}

	utils.OkDetailed(c, "获取成功", task)
}

// StopTask 停止任务
func StopTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.FailWithMsg(c, "任务ID不能为空")
		return
	}

	if err := services.GetTaskManager().CancelTask(taskID); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	utils.OkWithMsg(c, "任务已停止")
}

// GetTaskStats 获取任务统计
func GetTaskStats(c *gin.Context) {
	stats := models.GetTaskStats()

	// 添加运行中任务数
	runningTasks := services.GetTaskManager().GetRunningTasks()
	stats["active"] = int64(len(runningTasks))

	utils.OkDetailed(c, "获取成功", stats)
}

// GetRunningTasks 获取运行中的任务
func GetRunningTasks(c *gin.Context) {
	runningTasks := services.GetTaskManager().GetRunningTasksInfo()
	utils.OkDetailed(c, "获取成功", runningTasks)
}

// ClearTaskHistory 清理任务历史
func ClearTaskHistory(c *gin.Context) {
	var req struct {
		Before string `json:"before"` // ISO 时间字符串，可选
		Days   int    `json:"days"`   // 保留最近几天，可选；0表示清理全部
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有提供任何参数，默认清理30天前的任务
		req.Days = 30
	}

	var beforeTime time.Time
	var clearAll bool
	if req.Before != "" {
		parsed, err := time.Parse(time.RFC3339, req.Before)
		if err != nil {
			// 尝试其他格式
			parsed, err = time.Parse("2006-01-02", req.Before)
			if err != nil {
				utils.FailWithMsg(c, "时间格式错误")
				return
			}
		}
		beforeTime = parsed
	} else if req.Days > 0 {
		beforeTime = time.Now().AddDate(0, 0, -req.Days)
	} else {
		// days == 0 表示清理全部已完成/取消/失败的任务
		clearAll = true
		beforeTime = time.Now().Add(time.Hour) // 未来时间，确保清理所有符合条件的任务
	}

	affected, err := models.CleanupOldTasks(beforeTime)
	if err != nil {
		utils.FailWithMsg(c, "清理失败: "+err.Error())
		return
	}

	message := "清理完成"
	if clearAll {
		message = "已清理全部历史记录"
	}

	utils.OkDetailed(c, message, gin.H{
		"affected": affected,
		"clearAll": clearAll,
	})
}
