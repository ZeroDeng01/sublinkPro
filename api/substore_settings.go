package api

import (
	"sublink/services/substore"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

type subStoreSettingsRequest struct {
	Enabled          bool     `json:"enabled"`
	BaseURL          string   `json:"baseUrl"`
	TimeoutSeconds   int      `json:"timeoutSeconds"`
	AllowedTargets   []string `json:"allowedTargets"`
	MaxResponseBytes int64    `json:"maxResponseBytes"`
}

func GetSubStoreSettings(c *gin.Context) {
	utils.OkDetailedI18n(c, "获取 Sub-Store 设置成功", substore.ResolveEffectiveSettings(), "settings.subStore.api.loaded", nil)
}

func UpdateSubStoreSettings(c *gin.Context) {
	var req subStoreSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithI18n(c, "参数错误", "settings.subStore.api.invalidRequest", nil)
		return
	}
	settings := substore.Settings{
		Enabled:          req.Enabled,
		BaseURL:          req.BaseURL,
		TimeoutSeconds:   req.TimeoutSeconds,
		AllowedTargets:   req.AllowedTargets,
		MaxResponseBytes: req.MaxResponseBytes,
	}
	if err := substore.SaveSettings(settings); err != nil {
		utils.FailWithI18n(c, "保存 Sub-Store 设置失败: "+err.Error(), "settings.subStore.api.saveFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "Sub-Store 设置已保存", substore.ResolveEffectiveSettings(), "settings.subStore.api.saved", nil)
}

func TestSubStoreSettings(c *gin.Context) {
	var req subStoreSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithI18n(c, "参数错误", "settings.subStore.api.invalidRequest", nil)
		return
	}
	settings := substore.Settings{
		Enabled:          req.Enabled,
		BaseURL:          req.BaseURL,
		TimeoutSeconds:   req.TimeoutSeconds,
		AllowedTargets:   req.AllowedTargets,
		MaxResponseBytes: req.MaxResponseBytes,
	}
	if req.TimeoutSeconds <= 0 && req.MaxResponseBytes <= 0 && req.BaseURL == "" && len(req.AllowedTargets) == 0 {
		settings = substore.LoadSettings()
	}
	cfg, err := substore.ConfigFromSettings(settings)
	if err != nil {
		utils.FailWithI18n(c, "Sub-Store 设置无效: "+err.Error(), "settings.subStore.api.invalidSettings", map[string]any{"message": err.Error()})
		return
	}
	client, err := substore.NewClient(cfg)
	if err != nil {
		utils.FailWithI18n(c, "Sub-Store 未配置: "+err.Error(), "settings.subStore.api.notConfigured", map[string]any{"message": err.Error()})
		return
	}
	target := "loon"
	if len(settings.AllowedTargets) > 0 {
		target = settings.AllowedTargets[0]
	}
	converted, err := client.Convert(c.Request.Context(), "proxies:\n  - {name: substore-test, type: ss, server: 127.0.0.1, port: 8388, cipher: aes-128-gcm, password: pass}\n", target)
	if err != nil {
		utils.FailWithI18n(c, "Sub-Store 连接测试失败: "+err.Error(), "settings.subStore.api.testFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "Sub-Store 连接测试成功", gin.H{
		"target":      target,
		"contentType": converted.ContentType,
		"resultBytes": len(converted.Body),
	}, "settings.subStore.api.testSucceeded", map[string]any{"target": target, "bytes": len(converted.Body)})
}
