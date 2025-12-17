package api

import (
	"strconv"
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// HostAdd 添加 Host
func HostAdd(c *gin.Context) {
	var data models.Host
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if data.Hostname == "" || data.IP == "" {
		utils.FailWithMsg(c, "域名和IP不能为空")
		return
	}

	if err := data.Add(); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "添加成功", data)
}

// HostUpdate 更新 Host
func HostUpdate(c *gin.Context) {
	var data models.Host
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if data.ID == 0 {
		utils.FailWithMsg(c, "ID不能为空")
		return
	}
	if data.Hostname == "" || data.IP == "" {
		utils.FailWithMsg(c, "域名和IP不能为空")
		return
	}

	if err := data.Update(); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "更新成功", data)
}

// HostDelete 删除单个 Host
func HostDelete(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		utils.FailWithMsg(c, "ID不能为空")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "ID格式错误")
		return
	}

	host := &models.Host{ID: id}
	if err := host.Delete(); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkWithMsg(c, "删除成功")
}

// HostBatchDelete 批量删除 Host
func HostBatchDelete(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "请选择要删除的记录")
		return
	}

	if err := models.BatchDeleteHosts(req.IDs); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkWithMsg(c, "批量删除成功")
}

// HostList 获取 Host 列表
func HostList(c *gin.Context) {
	var data models.Host
	list, err := data.List()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "获取成功", list)
}

// HostExport 导出所有 Host 为文本格式
func HostExport(c *gin.Context) {
	text := models.ExportHostsToText()
	utils.OkDetailed(c, "导出成功", gin.H{
		"text": text,
	})
}

// HostSync 从文本全量同步 Host
func HostSync(c *gin.Context) {
	var req struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	added, updated, deleted, err := models.SyncHostsFromText(req.Text)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	utils.OkDetailed(c, "同步成功", gin.H{
		"added":   added,
		"updated": updated,
		"deleted": deleted,
	})
}
