package api

import (
	"strconv"
	"sublink/database"
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// ScriptAdd 添加脚本
func ScriptAdd(c *gin.Context) {
	var data models.Script
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if data.Name == "" || data.Content == "" {
		utils.FailWithMsg(c, "名称和内容不能为空")
		return
	}
	if data.Version == "" {
		data.Version = "0.0.0"
	}

	if data.CheckNameVersion() {
		utils.FailWithMsg(c, "该名称和版本的脚本已存在")
		return
	}

	if err := data.Add(); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "添加成功", data)
}

// ScriptDel 删除脚本
func ScriptDel(c *gin.Context) {
	var data models.Script
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if err := data.Del(); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkWithMsg(c, "删除成功")
}

func GetScriptUsage(c *gin.Context) {
	scriptIDStr := c.Query("id")
	if scriptIDStr == "" {
		utils.FailWithMsg(c, "脚本ID不能为空")
		return
	}

	scriptID, err := strconv.Atoi(scriptIDStr)
	if err != nil || scriptID <= 0 {
		utils.FailWithMsg(c, "脚本ID非法")
		return
	}

	var subScripts []models.SubcriptionScript
	if err := database.DB.Where("script_id = ?", scriptID).Find(&subScripts).Error; err != nil {
		utils.FailWithMsg(c, "获取脚本使用情况失败")
		return
	}

	if len(subScripts) == 0 {
		utils.OkWithData(c, gin.H{
			"subscriptions": []string{},
			"count":         0,
		})
		return
	}

	subIDs := make([]int, 0, len(subScripts))
	for _, subScript := range subScripts {
		subIDs = append(subIDs, subScript.SubcriptionID)
	}

	var subs []models.Subcription
	if err := database.DB.Where("id IN ?", subIDs).Find(&subs).Error; err != nil {
		utils.FailWithMsg(c, "获取脚本使用情况失败")
		return
	}

	usedBy := make([]string, 0, len(subs))
	for _, sub := range subs {
		usedBy = append(usedBy, sub.Name)
	}

	utils.OkWithData(c, gin.H{
		"subscriptions": usedBy,
		"count":         len(usedBy),
	})
}

// ScriptUpdate 更新脚本
func ScriptUpdate(c *gin.Context) {
	var data models.Script
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if data.CheckNameVersion() {
		utils.FailWithMsg(c, "该名称和版本的脚本已存在")
		return
	}
	if err := data.Update(); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "更新成功", data)
}

// ScriptList 获取脚本列表
func ScriptList(c *gin.Context) {
	var data models.Script

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

	// 如果提供了分页参数，返回分页响应
	if page > 0 && pageSize > 0 {
		list, total, err := data.ListPaginated(page, pageSize)
		if err != nil {
			utils.FailWithMsg(c, err.Error())
			return
		}
		totalPages := 0
		if pageSize > 0 {
			totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
		}
		utils.OkDetailed(c, "获取成功", gin.H{
			"items":      list,
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		})
		return
	}

	// 不带分页参数，返回全部（向后兼容）
	list, err := data.List()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "获取成功", list)
}
