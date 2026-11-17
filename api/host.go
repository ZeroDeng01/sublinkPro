package api

import (
	"strconv"
	"sublink/models"
	"sublink/services/mihomo"
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
		utils.FailWithMsg(c, "Domain and IP are required")
		return
	}

	if data.Source == "" {
		data.Source = "Manual"
	}

	if err := data.Add(); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "Added successfully", data)
}

// HostUpdate 更新 Host
func HostUpdate(c *gin.Context) {
	var data models.Host
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if data.ID == 0 {
		utils.FailWithMsg(c, "ID is required")
		return
	}
	if data.Hostname == "" || data.IP == "" {
		utils.FailWithMsg(c, "Domain and IP are required")
		return
	}

	if err := data.Update(); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "Updated successfully", data)
}

// HostDelete 删除单个 Host
func HostDelete(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		utils.FailWithMsg(c, "ID is required")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "Invalid ID format")
		return
	}

	host := &models.Host{ID: id}
	if err := host.Delete(); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkWithMsg(c, "Deleted successfully")
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
		utils.FailWithMsg(c, "Select records to delete")
		return
	}

	if err := models.BatchDeleteHosts(req.IDs); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkWithMsg(c, "Batch delete completed")
}

// HostList 获取 Host 列表
func HostList(c *gin.Context) {
	var data models.Host
	list, err := data.List()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "Fetched successfully", list)
}

// HostExport 导出所有 Host 为文本格式
func HostExport(c *gin.Context) {
	text := models.ExportHostsToText()
	utils.OkDetailed(c, "Exported successfully", gin.H{
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

	utils.OkDetailed(c, "Synced successfully", gin.H{
		"added":   added,
		"updated": updated,
		"deleted": deleted,
	})
}

// GetHostSettings 获取 Host 模块设置
func GetHostSettings(c *gin.Context) {
	persistHostStr, _ := models.GetSetting("speed_test_persist_host")
	persistHost := persistHostStr == "true"

	dnsServer, _ := models.GetSetting("dns_server")
	// 注意：dnsServer 为空表示使用系统DNS，但在前端可能需要区分"未设置"和"空字符串"
	// 这里直接返回数据库中的值即可

	dnsUseProxyStr, _ := models.GetSetting("dns_use_proxy")
	dnsUseProxy := dnsUseProxyStr == "true"

	dnsProxyStrategy, _ := models.GetSetting("dns_proxy_strategy")
	if dnsProxyStrategy == "" {
		dnsProxyStrategy = "auto" // 默认自动
	}

	dnsProxyNodeIDStr, _ := models.GetSetting("dns_proxy_node_id")
	dnsProxyNodeID := 0
	if dnsProxyNodeIDStr != "" {
		dnsProxyNodeID, _ = strconv.Atoi(dnsProxyNodeIDStr)
	}

	// 获取有效期设置
	expireHours := models.GetHostExpireHours()

	utils.OkDetailed(c, "Fetched successfully", gin.H{
		"persist_host":       persistHost,
		"dns_server":         dnsServer,
		"dns_use_proxy":      dnsUseProxy,
		"dns_proxy_strategy": dnsProxyStrategy, // auto 或 manual
		"dns_proxy_node_id":  dnsProxyNodeID,
		"dns_presets":        mihomo.GetDNSPresets(),
		"expire_hours":       expireHours,
	})
}

// UpdateHostSettings 更新 Host 模块设置
func UpdateHostSettings(c *gin.Context) {
	var req struct {
		PersistHost      *bool  `json:"persist_host"`
		DNSServer        string `json:"dns_server"`         // 允许为空
		DNSUseProxy      *bool  `json:"dns_use_proxy"`      // 是否使用代理
		DNSProxyStrategy string `json:"dns_proxy_strategy"` // auto 或 manual
		DNSProxyNodeID   int    `json:"dns_proxy_node_id"`  // 代理节点ID
		ExpireHours      *int   `json:"expire_hours"`       // 有效期（小时），0表示永不过期
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}

	if req.PersistHost != nil {
		if err := models.SetSetting("speed_test_persist_host", strconv.FormatBool(*req.PersistHost)); err != nil {
			utils.FailWithMsg(c, "Failed to save persistent Host configuration")
			return
		}
	}

	// 总是保存 dns_server，即使为空
	if err := models.SetSetting("dns_server", req.DNSServer); err != nil {
		utils.FailWithMsg(c, "Failed to save DNS server configuration")
		return
	}

	if req.DNSUseProxy != nil {
		if err := models.SetSetting("dns_use_proxy", strconv.FormatBool(*req.DNSUseProxy)); err != nil {
			utils.FailWithMsg(c, "Failed to save DNS proxy toggle")
			return
		}
	}

	if req.DNSProxyStrategy != "" {
		if err := models.SetSetting("dns_proxy_strategy", req.DNSProxyStrategy); err != nil {
			utils.FailWithMsg(c, "Failed to save DNS proxy strategy")
			return
		}
	}

	if req.DNSProxyNodeID != 0 || req.DNSProxyStrategy == "manual" {
		// 只有在 manual 模式下或确实传了 ID 时才保存
		if err := models.SetSetting("dns_proxy_node_id", strconv.Itoa(req.DNSProxyNodeID)); err != nil {
			utils.FailWithMsg(c, "Failed to save DNS proxy node ID")
			return
		}
	}

	if req.ExpireHours != nil {
		hours := *req.ExpireHours
		if hours < 0 {
			hours = 0
		}
		if err := models.SetSetting("host_expire_hours", strconv.Itoa(hours)); err != nil {
			utils.FailWithMsg(c, "Failed to save validity configuration")
			return
		}
	}

	utils.OkWithMsg(c, "Saved successfully")
}

// HostSetPinned 设置 Host 的固定状态
func HostSetPinned(c *gin.Context) {
	var req struct {
		ID     int  `json:"id"`
		Pinned bool `json:"pinned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if req.ID == 0 {
		utils.FailWithMsg(c, "ID is required")
		return
	}

	if err := models.SetHostPinned(req.ID, req.Pinned); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	utils.OkWithMsg(c, "Settings saved")
}
