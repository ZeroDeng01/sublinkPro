package models

import (
	"os"
	"sublink/config"
	"sublink/utils"

	"gopkg.in/yaml.v3"
)

// Config 配置结构体（保持向后兼容）
type Config struct {
	JwtSecret        string `yaml:"jwt_secret"`         // JWT密钥
	APIEncryptionKey string `yaml:"api_encryption_key"` // API加密密钥
	ExpireDays       int    `yaml:"expire_days"`        // 过期天数
	Port             int    `yaml:"port"`               // 端口号
	LoginFailCount   int    `yaml:"login_fail_count"`   // 登录失败次数限制
	LoginFailWindow  int    `yaml:"login_fail_window"`  // 登录失败窗口时间(分钟)
	LoginBanDuration int    `yaml:"login_ban_duration"` // 登录失败封禁时间(分钟)
}

// 配置文件注释
var configComment = `# SublinkPro 配置文件
# 敏感配置已存储在数据库中，此文件仅保存非敏感配置
# 如需覆盖敏感配置，请使用环境变量：
#   SUBLINK_JWT_SECRET - JWT签名密钥
#   SUBLINK_API_ENCRYPTION_KEY - API加密密钥
#
# 配置优先级：命令行参数 > 环境变量 > 配置文件 > 数据库 > 默认值
`

// InitSecretAccessors 初始化敏感配置访问器
// 在数据库初始化后调用，将 GetSetting/SetSetting 注入到 config 包
func InitSecretAccessors() {
	config.SetSecretAccessors(
		func(key string) string {
			val, _ := GetSetting(key)
			return val
		},
		func(key, value string) error {
			return SetSetting(key, value)
		},
	)
}

// ConfigInit 初始化配置（保持向后兼容）
// 已废弃：请使用 config.Load() 替代
func ConfigInit() {
	dbPath := config.GetDBPath()

	// 确保数据库目录存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dbPath, 0755); err != nil {
			utils.Error("创建数据库目录失败: %v", err)
		}
	}

	configPath := config.GetConfigFilePath()

	// 如果配置文件不存在，创建默认配置文件
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := Config{
			ExpireDays:       config.DefaultExpireDays,
			Port:             config.DefaultPort,
			LoginFailCount:   config.DefaultLoginFailCount,
			LoginFailWindow:  config.DefaultLoginFailWindow,
			LoginBanDuration: config.DefaultLoginBanDuration,
		}

		data, err := yaml.Marshal(&defaultConfig)
		if err != nil {
			utils.Error("生成默认配置文件失败: %v", err)
			return
		}
		data = []byte(configComment + string(data))
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			utils.Error("写入默认配置文件失败: %v", err)
			return
		}
		utils.Info("配置文件不存在，已创建默认配置文件: %s", configPath)
	}
}

// ReadConfig 读取配置（保持向后兼容）
func ReadConfig() Config {
	appCfg := config.Get()

	cfg := Config{
		JwtSecret:        appCfg.JwtSecret,
		APIEncryptionKey: appCfg.APIEncryptionKey,
		ExpireDays:       appCfg.ExpireDays,
		Port:             appCfg.Port,
		LoginFailCount:   appCfg.LoginFailCount,
		LoginFailWindow:  appCfg.LoginFailWindow,
		LoginBanDuration: appCfg.LoginBanDuration,
	}

	// 应用默认值（防止零值）
	if cfg.LoginFailCount == 0 {
		cfg.LoginFailCount = config.DefaultLoginFailCount
	}
	if cfg.LoginFailWindow == 0 {
		cfg.LoginFailWindow = config.DefaultLoginFailWindow
	}
	if cfg.LoginBanDuration == 0 {
		cfg.LoginBanDuration = config.DefaultLoginBanDuration
	}
	if cfg.Port == 0 {
		cfg.Port = config.DefaultPort
	}
	if cfg.ExpireDays == 0 {
		cfg.ExpireDays = config.DefaultExpireDays
	}

	return cfg
}

// SetConfig 设置配置
func SetConfig(newCfg Config) {
	config.UpdateConfig(func(appCfg *config.AppConfig) {
		if newCfg.Port != 0 {
			appCfg.Port = newCfg.Port
		}
		if newCfg.ExpireDays != 0 {
			appCfg.ExpireDays = newCfg.ExpireDays
		}
		if newCfg.LoginFailCount != 0 {
			appCfg.LoginFailCount = newCfg.LoginFailCount
		}
		if newCfg.LoginFailWindow != 0 {
			appCfg.LoginFailWindow = newCfg.LoginFailWindow
		}
		if newCfg.LoginBanDuration != 0 {
			appCfg.LoginBanDuration = newCfg.LoginBanDuration
		}
	})

	// 保存到配置文件
	if err := config.SaveToFile(); err != nil {
		utils.Error("保存配置到文件失败: %v", err)
	}
}
