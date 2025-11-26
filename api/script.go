package api

import (
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
	list, err := data.List()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "获取成功", list)
}
