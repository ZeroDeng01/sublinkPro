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

// GetGlobalNodeProcessingConfig 获取全局节点处理配置
func GetGlobalNodeProcessingConfig(c *gin.Context) {
	whitelist, _ := models.GetSetting("global_node_name_whitelist")
	blacklist, _ := models.GetSetting("global_node_name_blacklist")
	protocolWhitelist, _ := models.GetSetting("global_protocol_whitelist")
	protocolBlacklist, _ := models.GetSetting("global_protocol_blacklist")
	preprocess, _ := models.GetSetting("global_node_name_preprocess")

	utils.OkDetailed(c, "获取成功", gin.H{
		"nodeNameWhitelist":  whitelist,
		"nodeNameBlacklist":  blacklist,
		"protocolWhitelist":  protocolWhitelist,
		"protocolBlacklist":  protocolBlacklist,
		"nodeNamePreprocess": preprocess,
	})
}

// UpdateGlobalNodeProcessingConfig 更新全局节点处理配置
func UpdateGlobalNodeProcessingConfig(c *gin.Context) {
	var req struct {
		NodeNameWhitelist  string `json:"nodeNameWhitelist"`
		NodeNameBlacklist  string `json:"nodeNameBlacklist"`
		ProtocolWhitelist  string `json:"protocolWhitelist"`
		ProtocolBlacklist  string `json:"protocolBlacklist"`
		NodeNamePreprocess string `json:"nodeNamePreprocess"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	// 保存配置
	if err := models.SetSetting("global_node_name_whitelist", req.NodeNameWhitelist); err != nil {
		utils.FailWithMsg(c, "保存全局节点名称白名单失败: "+err.Error())
		return
	}
	if err := models.SetSetting("global_node_name_blacklist", req.NodeNameBlacklist); err != nil {
		utils.FailWithMsg(c, "保存全局节点名称黑名单失败: "+err.Error())
		return
	}
	if err := models.SetSetting("global_protocol_whitelist", req.ProtocolWhitelist); err != nil {
		utils.FailWithMsg(c, "保存全局协议白名单失败: "+err.Error())
		return
	}
	if err := models.SetSetting("global_protocol_blacklist", req.ProtocolBlacklist); err != nil {
		utils.FailWithMsg(c, "保存全局协议黑名单失败: "+err.Error())
		return
	}
	if err := models.SetSetting("global_node_name_preprocess", req.NodeNamePreprocess); err != nil {
		utils.FailWithMsg(c, "保存全局节点名称预处理规则失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "保存成功")
}
