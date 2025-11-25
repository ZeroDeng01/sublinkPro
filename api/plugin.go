package api

import (
	"sublink/utils"

	"github.com/eun1e/sublinkE-plugins"
	"github.com/gin-gonic/gin"
)

// PluginListResponse 插件列表响应
type PluginListResponse struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	FilePath    string                 `json:"filePath"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// PluginConfigRequest 插件配置请求
type PluginConfigRequest struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

// GetPlugins 获取所有插件列表
func GetPlugins(c *gin.Context) {
	manager := plugins.GetManager()
	allPlugins := manager.GetAllPlugins()

	var response []PluginListResponse
	for _, plugin := range allPlugins {
		response = append(response, PluginListResponse{
			Name:        plugin.Name,
			Version:     plugin.Version,
			Description: plugin.Description,
			Enabled:     plugin.Enabled,
			FilePath:    plugin.FilePath,
			Config:      plugin.Config,
		})
	}

	utils.OkDetailed(c, "获取插件列表成功", response)
}

// EnablePlugin 启用插件
func EnablePlugin(c *gin.Context) {
	pluginName := c.Param("name")
	if pluginName == "" {
		utils.FailWithMsg(c, "插件名称不能为空")
		return
	}

	manager := plugins.GetManager()
	if err := manager.EnablePlugin(pluginName); err != nil {
		utils.FailWithMsg(c, "启用插件失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "插件启用成功")
}

// DisablePlugin 禁用插件
func DisablePlugin(c *gin.Context) {
	pluginName := c.Param("name")
	if pluginName == "" {
		utils.FailWithMsg(c, "插件名称不能为空")
		return
	}

	manager := plugins.GetManager()
	if err := manager.DisablePlugin(pluginName); err != nil {
		utils.FailWithMsg(c, "禁用插件失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "插件禁用成功")
}

// UpdatePluginConfig 更新插件配置
func UpdatePluginConfig(c *gin.Context) {
	var req PluginConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请求参数错误: "+err.Error())
		return
	}

	manager := plugins.GetManager()
	if err := manager.UpdatePluginConfig(req.Name, req.Config); err != nil {
		utils.FailWithMsg(c, "更新插件配置失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "插件配置更新成功")
}

// GetPluginConfig 获取插件配置
func GetPluginConfig(c *gin.Context) {
	pluginName := c.Param("name")
	if pluginName == "" {
		utils.FailWithMsg(c, "插件名称不能为空")
		return
	}

	manager := plugins.GetManager()
	plugin, exists := manager.GetPlugin(pluginName)
	if !exists {
		utils.FailWithMsg(c, "插件不存在")
		return
	}

	utils.OkDetailed(c, "获取插件配置成功", plugin.Config)
}

// ReloadPlugins 重新加载插件
func ReloadPlugins(c *gin.Context) {
	manager := plugins.GetManager()

	// 关闭所有插件
	manager.Shutdown()

	// 重新加载插件
	if err := manager.LoadPlugins(); err != nil {
		utils.FailWithMsg(c, "重新加载插件失败: "+err.Error())
		return
	}

	utils.OkWithMsg(c, "插件重新加载成功")
}
