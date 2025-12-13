package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

// Tasks 注册任务管理相关路由
func Tasks(r *gin.Engine) {
	tasksGroup := r.Group("/api/v1/tasks")
	tasksGroup.Use(middlewares.AuthToken)
	{
		tasksGroup.GET("", api.GetTasks)                          // 获取任务列表
		tasksGroup.GET("/stats", api.GetTaskStats)                // 获取任务统计
		tasksGroup.GET("/running", api.GetRunningTasks)           // 获取运行中任务
		tasksGroup.GET("/:id", api.GetTask)                       // 获取任务详情
		tasksGroup.GET("/:id/traffic", api.GetTaskTrafficDetails) // 获取任务流量明细
		tasksGroup.POST("/:id/stop", api.StopTask)                // 停止任务
		tasksGroup.DELETE("", api.ClearTaskHistory)               // 清理历史
	}
}
