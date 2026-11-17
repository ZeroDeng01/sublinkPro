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

	utils.OkDetailed(c, "Fetched successfully", gin.H{
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
		utils.FailWithMsg(c, "Invalid parameters: category must be clash or surge")
		return
	}

	key := "base_template_" + req.Category
	if err := models.SetSetting(key, req.Content); err != nil {
		utils.FailWithMsg(c, "Failed to save template: "+err.Error())
		return
	}

	categoryName := "Clash"
	if req.Category == "surge" {
		categoryName = "Surge"
	}
	utils.OkWithMsg(c, categoryName+" base templates saved")
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
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if err := models.SetSetting("system_domain", req.SystemDomain); err != nil {
		utils.FailWithMsg(c, "Save failed: "+err.Error())
		return
	}
	utils.OkWithMsg(c, "Saved successfully")
}

// GetNodeDedupConfig 获取节点去重配置
func GetNodeDedupConfig(c *gin.Context) {
	crossAirportDedup, _ := models.GetSetting("cross_airport_dedup_enabled")
	utils.OkDetailed(c, "Fetched successfully", gin.H{
		"crossAirportDedupEnabled": crossAirportDedup != "false",
	})
}

// UpdateNodeDedupConfig 更新节点去重配置
func UpdateNodeDedupConfig(c *gin.Context) {
	var req struct {
		CrossAirportDedupEnabled *bool `json:"crossAirportDedupEnabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if req.CrossAirportDedupEnabled == nil {
		utils.FailWithMsg(c, "Missing required field crossAirportDedupEnabled")
		return
	}
	value := "true"
	if !*req.CrossAirportDedupEnabled {
		value = "false"
	}
	if err := models.SetSetting("cross_airport_dedup_enabled", value); err != nil {
		utils.FailWithMsg(c, "Save failed: "+err.Error())
		return
	}
	utils.OkWithMsg(c, "Saved successfully")
}

// GetGlobalNodeProcessingConfig 获取全局节点处理配置
func GetGlobalNodeProcessingConfig(c *gin.Context) {
	whitelist, _ := models.GetSetting("global_node_name_whitelist")
	blacklist, _ := models.GetSetting("global_node_name_blacklist")
	protocolWhitelist, _ := models.GetSetting("global_protocol_whitelist")
	protocolBlacklist, _ := models.GetSetting("global_protocol_blacklist")
	preprocess, _ := models.GetSetting("global_node_name_preprocess")

	utils.OkDetailed(c, "Fetched successfully", gin.H{
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
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}

	// 保存配置
	if err := models.SetSetting("global_node_name_whitelist", req.NodeNameWhitelist); err != nil {
		utils.FailWithMsg(c, "Failed to save global node name whitelist: "+err.Error())
		return
	}
	if err := models.SetSetting("global_node_name_blacklist", req.NodeNameBlacklist); err != nil {
		utils.FailWithMsg(c, "Failed to save global node name blacklist: "+err.Error())
		return
	}
	if err := models.SetSetting("global_protocol_whitelist", req.ProtocolWhitelist); err != nil {
		utils.FailWithMsg(c, "Failed to save global protocol whitelist: "+err.Error())
		return
	}
	if err := models.SetSetting("global_protocol_blacklist", req.ProtocolBlacklist); err != nil {
		utils.FailWithMsg(c, "Failed to save global protocol blacklist: "+err.Error())
		return
	}
	if err := models.SetSetting("global_node_name_preprocess", req.NodeNamePreprocess); err != nil {
		utils.FailWithMsg(c, "Failed to save global node name preprocessing rules: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "Saved successfully")
}

// GetUsageHistoryConfig 获取用量历史配置
func GetUsageHistoryConfig(c *gin.Context) {
	retentionDays := models.GetAirportUsageRetentionDays()
	snapshotCount, _ := models.CountAirportUsageSnapshots()

	utils.OkDetailed(c, "success", gin.H{
		"retentionDays":    retentionDays,
		"minRetentionDays": models.MinAirportUsageRetentionDays,
		"maxRetentionDays": models.MaxAirportUsageRetentionDays,
		"snapshotCount":    snapshotCount,
	})
}

// UpdateUsageHistoryConfig 更新用量历史配置，保存后立即清理超出保留期的历史
func UpdateUsageHistoryConfig(c *gin.Context) {
	var req struct {
		RetentionDays *int `json:"retentionDays"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "invalid payload")
		return
	}
	if req.RetentionDays == nil {
		utils.FailWithMsg(c, "missing required field retentionDays")
		return
	}
	if err := models.SetAirportUsageRetentionDays(*req.RetentionDays); err != nil {
		utils.FailWithMsg(c, "failed to save: "+err.Error())
		return
	}

	utils.OkDetailed(c, "success", gin.H{
		"retentionDays": *req.RetentionDays,
	})
}
