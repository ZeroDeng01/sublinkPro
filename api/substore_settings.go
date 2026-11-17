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
	utils.OkDetailedI18n(c, "Sub-Store settings loaded", substore.ResolveEffectiveSettings(), "settings.subStore.api.loaded", nil)
}

func UpdateSubStoreSettings(c *gin.Context) {
	var req subStoreSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithI18n(c, "Invalid parameters", "settings.subStore.api.invalidRequest", nil)
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
		utils.FailWithI18n(c, "Failed to save Sub-Store settings: "+err.Error(), "settings.subStore.api.saveFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "Sub-Store settings saved", substore.ResolveEffectiveSettings(), "settings.subStore.api.saved", nil)
}

func TestSubStoreSettings(c *gin.Context) {
	var req subStoreSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithI18n(c, "Invalid parameters", "settings.subStore.api.invalidRequest", nil)
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
		utils.FailWithI18n(c, "Invalid Sub-Store settings: "+err.Error(), "settings.subStore.api.invalidSettings", map[string]any{"message": err.Error()})
		return
	}
	client, err := substore.NewClient(cfg)
	if err != nil {
		utils.FailWithI18n(c, "Sub-Store is not configured: "+err.Error(), "settings.subStore.api.notConfigured", map[string]any{"message": err.Error()})
		return
	}
	target := "loon"
	if len(settings.AllowedTargets) > 0 {
		target = settings.AllowedTargets[0]
	}
	converted, err := client.Convert(c.Request.Context(), "proxies:\n  - {name: substore-test, type: ss, server: 127.0.0.1, port: 8388, cipher: aes-128-gcm, password: pass}\n", target)
	if err != nil {
		utils.FailWithI18n(c, "Sub-Store connection test failed: "+err.Error(), "settings.subStore.api.testFailed", map[string]any{"message": err.Error()})
		return
	}
	utils.OkDetailedI18n(c, "Sub-Store connection test succeeded", gin.H{
		"target":      target,
		"contentType": converted.ContentType,
		"resultBytes": len(converted.Body),
	}, "settings.subStore.api.testSucceeded", map[string]any{"target": target, "bytes": len(converted.Body)})
}
