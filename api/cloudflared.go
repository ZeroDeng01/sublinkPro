package api

import (
	"strings"
	"sublink/services/cloudflared"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

func requireSessionAuthForCloudflared(c *gin.Context) bool {
	if strings.TrimSpace(c.GetHeader("X-API-Key")) == "" {
		return true
	}
	utils.Forbidden(c, "Cloudflare Tunnel settings are only accessible from a login session")
	return false
}

// GetCloudflaredStatus 获取 Cloudflare Tunnel 配置和运行状态。
func GetCloudflaredStatus(c *gin.Context) {
	if !requireSessionAuthForCloudflared(c) {
		return
	}
	utils.OkDetailed(c, "Cloudflare Tunnel status fetched successfully", cloudflared.DefaultManager().Status())
}

// UpdateCloudflaredConfig 保存 Cloudflare Tunnel 基础配置。
func UpdateCloudflaredConfig(c *gin.Context) {
	if !requireSessionAuthForCloudflared(c) {
		return
	}

	var req struct {
		Enabled bool   `json:"enabled"`
		Token   string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Token) != "" {
		if err := cloudflared.SaveToken(req.Token); err != nil {
			utils.FailWithMsg(c, "Failed to save token: "+err.Error())
			return
		}
	}
	if err := cloudflared.SaveEnabled(req.Enabled); err != nil {
		utils.FailWithMsg(c, "Failed to save configuration: "+err.Error())
		return
	}
	utils.OkDetailed(c, "Cloudflare Tunnel configuration saved", cloudflared.DefaultManager().Status())
}

// StartCloudflared 启动 cloudflared 子进程。
func StartCloudflared(c *gin.Context) {
	if !requireSessionAuthForCloudflared(c) {
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters: "+err.Error())
		return
	}
	if err := cloudflared.DefaultManager().Start(req.Token); err != nil {
		utils.FailWithMsg(c, "Failed to start cloudflared: "+err.Error())
		return
	}
	utils.OkDetailed(c, "cloudflared started", cloudflared.DefaultManager().Status())
}

// StopCloudflared 停止 cloudflared 子进程并关闭自动启动。
func StopCloudflared(c *gin.Context) {
	if !requireSessionAuthForCloudflared(c) {
		return
	}
	if err := cloudflared.DefaultManager().Stop(); err != nil {
		utils.FailWithMsg(c, "Failed to stop cloudflared: "+err.Error())
		return
	}
	utils.OkDetailed(c, "cloudflared stopped", cloudflared.DefaultManager().Status())
}

// RemoveCloudflaredToken 清空 Cloudflare Tunnel token。
func RemoveCloudflaredToken(c *gin.Context) {
	if !requireSessionAuthForCloudflared(c) {
		return
	}
	if err := cloudflared.DefaultManager().Stop(); err != nil {
		utils.FailWithMsg(c, "Failed to stop cloudflared: "+err.Error())
		return
	}
	if err := cloudflared.ClearToken(); err != nil {
		utils.FailWithMsg(c, "Failed to clear token: "+err.Error())
		return
	}
	utils.OkDetailed(c, "Cloudflare Tunnel token cleared", cloudflared.DefaultManager().Status())
}
