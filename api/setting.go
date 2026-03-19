package api

import (
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// GetBaseTemplates 获取基础模板配置
func GetBaseTemplates(c *gin.Context) {
	clashTemplate, _ := models.GetSetting("base_template_clash")
	surgeTemplate, _ := models.GetSetting("base_template_surge")

	utils.OkDetailed(c, "获取成功", gin.H{
		"clashTemplate": clashTemplate,
		"surgeTemplate": surgeTemplate,
	})
}

// UpdateBaseTemplate 更新基础模板配置
func UpdateBaseTemplate(c *gin.Context) {
	var req struct {
		Category string `json:"category" binding:"required,oneof=clash surge"`
		Content  string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误：category 必须为 clash 或 surge")
		return
	}

	key := "base_template_" + req.Category
	if err := models.SetSetting(key, req.Content); err != nil {
		utils.FailWithMsg(c, "保存模板失败: "+err.Error())
		return
	}

	categoryName := "Clash"
	if req.Category == "surge" {
		categoryName = "Surge"
	}
	utils.OkWithMsg(c, categoryName+" 基础模板保存成功")
}

// GetSystemDomain 获取系统域名配置
func GetSystemDomain(c *gin.Context) {
	domain, _ := models.GetSetting("system_domain")
	utils.OkWithData(c, gin.H{"systemDomain": domain})
}

// UpdateSystemDomain 更新系统域名配置
func UpdateSystemDomain(c *gin.Context) {
	var req struct {
		SystemDomain string `json:"systemDomain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if err := models.SetSetting("system_domain", req.SystemDomain); err != nil {
		utils.FailWithMsg(c, "保存失败: "+err.Error())
		return
	}
	utils.OkWithMsg(c, "保存成功")
}

// GetNodeDedupConfig 获取节点去重配置
func GetNodeDedupConfig(c *gin.Context) {
	crossAirportDedup, _ := models.GetSetting("cross_airport_dedup_enabled")
	utils.OkDetailed(c, "获取成功", gin.H{
		"crossAirportDedupEnabled": crossAirportDedup != "false",
	})
}

// UpdateNodeDedupConfig 更新节点去重配置
func UpdateNodeDedupConfig(c *gin.Context) {
	var req struct {
		CrossAirportDedupEnabled *bool `json:"crossAirportDedupEnabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if req.CrossAirportDedupEnabled == nil {
		utils.FailWithMsg(c, "缺少必填字段 crossAirportDedupEnabled")
		return
	}
	value := "true"
	if !*req.CrossAirportDedupEnabled {
		value = "false"
	}
	if err := models.SetSetting("cross_airport_dedup_enabled", value); err != nil {
		utils.FailWithMsg(c, "保存失败: "+err.Error())
		return
	}
	utils.OkWithMsg(c, "保存成功")
}
