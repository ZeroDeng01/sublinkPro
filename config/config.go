package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"sync"

	"gopkg.in/yaml.v3"
)

// 环境变量前缀
const envPrefix = "SUBLINK_"

// 默认配置值
const (
	DefaultPort             = 8000
	DefaultExpireDays       = 14
	DefaultLoginFailCount   = 5
	DefaultLoginFailWindow  = 1
	DefaultLoginBanDuration = 10
	DefaultDBPath           = "./db"
	DefaultLogPath          = "./logs"
	DefaultLogLevel         = "info"
	DefaultConfigFile       = "config.yaml"
	DefaultCaptchaMode      = CaptchaModeTraditional // 默认传统验证码
)

// 验证码模式常量
const (
	CaptchaModeDisabled    = 1 // 关闭验证码
	CaptchaModeTraditional = 2 // 传统图形验证码（默认）
	CaptchaModeTurnstile   = 3 // Cloudflare Turnstile
)

// AppConfig 应用配置结构
type AppConfig struct {
	Port               int    `yaml:"port"`                 // 服务端口
	JwtSecret          string `yaml:"jwt_secret"`           // JWT密钥
	APIEncryptionKey   string `yaml:"api_encryption_key"`   // API加密密钥
	ExpireDays         int    `yaml:"expire_days"`          // Token过期天数
	LoginFailCount     int    `yaml:"login_fail_count"`     // 登录失败次数限制
	LoginFailWindow    int    `yaml:"login_fail_window"`    // 登录失败窗口时间(分钟)
	LoginBanDuration   int    `yaml:"login_ban_duration"`   // 登录失败封禁时间(分钟)
	DBPath             string `yaml:"db_path"`              // 数据库目录
	LogPath            string `yaml:"log_path"`             // 日志目录
	LogLevel           string `yaml:"log_level"`            // 日志等级 (debug/info/warn/error/fatal)
	GeoIPPath          string `yaml:"geoip_path"`           // GeoIP数据库路径
	CaptchaMode        int    `yaml:"captcha_mode"`         // 验证码模式 (1=关闭, 2=传统, 3=Turnstile)
	TurnstileSiteKey   string `yaml:"turnstile_site_key"`   // Cloudflare Turnstile Site Key
	TurnstileSecretKey string `yaml:"turnstile_secret_key"` // Cloudflare Turnstile Secret Key
	TurnstileProxyLink string `yaml:"turnstile_proxy_link"` // Turnstile 验证代理链接（mihomo 格式）
}

// CommandLineConfig 命令行配置（仅存储用户指定的值）
type CommandLineConfig struct {
	Port       int
	DBPath     string
	LogPath    string
	LogLevel   string
	ConfigFile string
}

var (
	globalConfig     *AppConfig                    // 全局配置实例
	cmdConfig        *CommandLineConfig            // 命令行配置
	configMutex      sync.RWMutex                  // 读写锁保护配置
	secretGetterFunc func(key string) string       // 从数据库获取敏感配置的函数
	secretSetterFunc func(key, value string) error // 向数据库写入敏感配置的函数
	initialized      bool
)

// SetCommandLineConfig 设置命令行配置（在 main 函数中调用）
func SetCommandLineConfig(cfg *CommandLineConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	cmdConfig = cfg
}

// SetSecretAccessors 设置敏感配置的访问函数
// 这些函数由 models 包提供，用于从数据库读写敏感配置
func SetSecretAccessors(getter func(key string) string, setter func(key, value string) error) {
	configMutex.Lock()
	defer configMutex.Unlock()
	secretGetterFunc = getter
	secretSetterFunc = setter
}

// GetDBPath 获取数据库路径（在初始化前可用）
func GetDBPath() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	// 命令行参数优先
	if cmdConfig != nil && cmdConfig.DBPath != "" {
		return cmdConfig.DBPath
	}
	// 环境变量次之
	if envPath := os.Getenv(envPrefix + "DB_PATH"); envPath != "" {
		return envPath
	}
	// 默认值
	return DefaultDBPath
}

// GetLogPath 获取日志路径（在初始化前可用）
func GetLogPath() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	// 命令行参数优先
	if cmdConfig != nil && cmdConfig.LogPath != "" {
		return cmdConfig.LogPath
	}
	// 环境变量次之
	if envPath := os.Getenv(envPrefix + "LOG_PATH"); envPath != "" {
		return envPath
	}
	// 默认值
	return DefaultLogPath
}

// GetGeoIPPath 获取 GeoIP 数据库路径（在初始化前可用）
func GetGeoIPPath() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	// 如果已加载配置，从配置中获取
	if globalConfig != nil && globalConfig.GeoIPPath != "" {
		return globalConfig.GeoIPPath
	}
	// 命令行参数优先（暂不支持命令行设置 GeoIPPath）
	// 环境变量次之
	if geoipPath := os.Getenv(envPrefix + "GEOIP_PATH"); geoipPath != "" {
		return geoipPath
	}
	// 默认使用 DBPath + /GeoLite2-City.mmdb
	dbPath := DefaultDBPath
	if cmdConfig != nil && cmdConfig.DBPath != "" {
		dbPath = cmdConfig.DBPath
	} else if envPath := os.Getenv(envPrefix + "DB_PATH"); envPath != "" {
		dbPath = envPath
	}
	return dbPath + "/GeoLite2-City.mmdb"
}

// GetConfigFilePath 获取配置文件完整路径
func GetConfigFilePath() string {
	// 先获取 dbPath（避免锁重入）
	dbPath := getDBPathInternal()
	configFile := DefaultConfigFile

	configMutex.RLock()
	// 命令行指定的配置文件
	if cmdConfig != nil && cmdConfig.ConfigFile != "" {
		configFile = cmdConfig.ConfigFile
	}
	configMutex.RUnlock()

	return dbPath + "/" + configFile
}

// getDBPathInternal 内部获取数据库路径（不加锁）
func getDBPathInternal() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	// 命令行参数优先
	if cmdConfig != nil && cmdConfig.DBPath != "" {
		return cmdConfig.DBPath
	}
	// 环境变量次之
	if envPath := os.Getenv(envPrefix + "DB_PATH"); envPath != "" {
		return envPath
	}
	// 默认值
	return DefaultDBPath
}

// Load 加载配置（数据库初始化后调用，以便读取敏感配置）
func Load() *AppConfig {
	// 先获取配置路径（避免在锁内调用其他需要锁的函数）
	configPath := GetConfigFilePath()

	configMutex.Lock()
	defer configMutex.Unlock()

	cfg := &AppConfig{}

	// 第一步：应用默认值
	applyDefaults(cfg)

	// 第二步：从配置文件加载
	loadFromFileInternal(cfg, configPath)

	// 第三步：从环境变量加载（覆盖配置文件）
	loadFromEnvInternal(cfg)

	// 第四步：从命令行加载（覆盖环境变量）
	loadFromCmdLineInternal(cfg)

	// 第五步：处理敏感配置（JWT Secret、API加密密钥）
	handleSecretsInternal(cfg)

	globalConfig = cfg
	initialized = true

	log.Printf("配置加载完成: Port=%d, ExpireDays=%d, DBPath=%s, LogPath=%s",
		cfg.Port, cfg.ExpireDays, cfg.DBPath, cfg.LogPath)

	return cfg
}

// Get 获取当前配置（线程安全）
func Get() *AppConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		// 返回默认配置
		cfg := &AppConfig{}
		applyDefaults(cfg)
		return cfg
	}
	return globalConfig
}

// GetJwtSecret 获取JWT密钥
func GetJwtSecret() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig != nil && globalConfig.JwtSecret != "" {
		return globalConfig.JwtSecret
	}
	// 如果配置尚未加载，尝试从环境变量获取
	if secret := os.Getenv(envPrefix + "JWT_SECRET"); secret != "" {
		return secret
	}
	return ""
}

// GetAPIEncryptionKey 获取API加密密钥
func GetAPIEncryptionKey() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig != nil && globalConfig.APIEncryptionKey != "" {
		return globalConfig.APIEncryptionKey
	}
	// 如果配置尚未加载，尝试从环境变量获取
	if key := os.Getenv(envPrefix + "API_ENCRYPTION_KEY"); key != "" {
		return key
	}
	return ""
}

// CaptchaConfig 验证码配置信息
type CaptchaConfig struct {
	Mode             int    // 当前验证码模式（经过降级处理后的实际模式）
	ConfiguredMode   int    // 用户配置的原始模式
	TurnstileSiteKey string // Turnstile Site Key
	Degraded         bool   // 是否已降级
}

// GetCaptchaConfig 获取验证码配置（含降级逻辑）
// 当配置为 Turnstile 但未设置密钥时，自动降级为传统验证码
func GetCaptchaConfig() CaptchaConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()

	cfg := CaptchaConfig{
		Mode:           CaptchaModeTraditional, // 默认传统验证码
		ConfiguredMode: CaptchaModeTraditional,
		Degraded:       false,
	}

	if globalConfig == nil {
		return cfg
	}

	cfg.ConfiguredMode = globalConfig.CaptchaMode
	if cfg.ConfiguredMode == 0 {
		cfg.ConfiguredMode = CaptchaModeTraditional
	}

	switch cfg.ConfiguredMode {
	case CaptchaModeDisabled:
		// 关闭验证码
		cfg.Mode = CaptchaModeDisabled
	case CaptchaModeTurnstile:
		// Turnstile 模式，检查是否配置了密钥
		if globalConfig.TurnstileSiteKey != "" && globalConfig.TurnstileSecretKey != "" {
			cfg.Mode = CaptchaModeTurnstile
			cfg.TurnstileSiteKey = globalConfig.TurnstileSiteKey
		} else {
			// 降级为传统验证码
			cfg.Mode = CaptchaModeTraditional
			cfg.Degraded = true
			log.Printf("警告: Turnstile 配置不完整，降级为传统验证码")
		}
	default:
		// 传统验证码
		cfg.Mode = CaptchaModeTraditional
	}

	return cfg
}

// GetTurnstileSecretKey 获取 Turnstile Secret Key（仅供后端使用）
func GetTurnstileSecretKey() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig != nil {
		return globalConfig.TurnstileSecretKey
	}
	return ""
}

// GetTurnstileProxyLink 获取 Turnstile 代理链接（填写即使用代理，未填写则直连）
func GetTurnstileProxyLink() string {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig != nil {
		return globalConfig.TurnstileProxyLink
	}
	return ""
}

// Reload 重新加载配置
func Reload() *AppConfig {
	return Load()
}

// applyDefaults 应用默认值
func applyDefaults(cfg *AppConfig) {
	cfg.Port = DefaultPort
	cfg.ExpireDays = DefaultExpireDays
	cfg.LoginFailCount = DefaultLoginFailCount
	cfg.LoginFailWindow = DefaultLoginFailWindow
	cfg.LoginBanDuration = DefaultLoginBanDuration
	cfg.DBPath = DefaultDBPath
	cfg.LogPath = DefaultLogPath
	cfg.LogLevel = DefaultLogLevel
	cfg.GeoIPPath = "" // 默认为空，运行时通过 GetGeoIPPath() 计算
	cfg.CaptchaMode = DefaultCaptchaMode
}

// loadFromFileInternal 从配置文件加载（内部使用，不获取锁）
func loadFromFileInternal(cfg *AppConfig, configPath string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("配置文件 %s 不存在，使用默认配置", configPath)
		} else {
			log.Printf("读取配置文件失败: %v", err)
		}
		return
	}

	fileCfg := &AppConfig{}
	if err := yaml.Unmarshal(data, fileCfg); err != nil {
		log.Printf("解析配置文件失败: %v", err)
		return
	}

	// 合并非零值
	if fileCfg.Port != 0 {
		cfg.Port = fileCfg.Port
	}
	if fileCfg.ExpireDays != 0 {
		cfg.ExpireDays = fileCfg.ExpireDays
	}
	if fileCfg.LoginFailCount != 0 {
		cfg.LoginFailCount = fileCfg.LoginFailCount
	}
	if fileCfg.LoginFailWindow != 0 {
		cfg.LoginFailWindow = fileCfg.LoginFailWindow
	}
	if fileCfg.LoginBanDuration != 0 {
		cfg.LoginBanDuration = fileCfg.LoginBanDuration
	}
	if fileCfg.DBPath != "" {
		cfg.DBPath = fileCfg.DBPath
	}
	if fileCfg.LogPath != "" {
		cfg.LogPath = fileCfg.LogPath
	}
	if fileCfg.LogLevel != "" {
		cfg.LogLevel = fileCfg.LogLevel
	}
	// 验证码配置
	if fileCfg.CaptchaMode != 0 {
		cfg.CaptchaMode = fileCfg.CaptchaMode
	}
	if fileCfg.TurnstileSiteKey != "" {
		cfg.TurnstileSiteKey = fileCfg.TurnstileSiteKey
	}
	if fileCfg.TurnstileSecretKey != "" {
		cfg.TurnstileSecretKey = fileCfg.TurnstileSecretKey
	}
	// 敏感配置也从文件读取（用于迁移）
	if fileCfg.JwtSecret != "" {
		cfg.JwtSecret = fileCfg.JwtSecret
	}
	if fileCfg.APIEncryptionKey != "" {
		cfg.APIEncryptionKey = fileCfg.APIEncryptionKey
	}
}

// loadFromEnvInternal 从环境变量加载（内部使用，不获取锁）
func loadFromEnvInternal(cfg *AppConfig) {
	if port := os.Getenv(envPrefix + "PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 {
			cfg.Port = p
		}
	}
	if days := os.Getenv(envPrefix + "EXPIRE_DAYS"); days != "" {
		if d, err := strconv.Atoi(days); err == nil && d > 0 {
			cfg.ExpireDays = d
		}
	}
	if count := os.Getenv(envPrefix + "LOGIN_FAIL_COUNT"); count != "" {
		if c, err := strconv.Atoi(count); err == nil && c > 0 {
			cfg.LoginFailCount = c
		}
	}
	if window := os.Getenv(envPrefix + "LOGIN_FAIL_WINDOW"); window != "" {
		if w, err := strconv.Atoi(window); err == nil && w > 0 {
			cfg.LoginFailWindow = w
		}
	}
	if duration := os.Getenv(envPrefix + "LOGIN_BAN_DURATION"); duration != "" {
		if d, err := strconv.Atoi(duration); err == nil && d > 0 {
			cfg.LoginBanDuration = d
		}
	}
	if dbPath := os.Getenv(envPrefix + "DB_PATH"); dbPath != "" {
		cfg.DBPath = dbPath
	}
	if logPath := os.Getenv(envPrefix + "LOG_PATH"); logPath != "" {
		cfg.LogPath = logPath
	}
	if logLevel := os.Getenv(envPrefix + "LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}
	if geoipPath := os.Getenv(envPrefix + "GEOIP_PATH"); geoipPath != "" {
		cfg.GeoIPPath = geoipPath
	}
	// 敏感配置
	if secret := os.Getenv(envPrefix + "JWT_SECRET"); secret != "" {
		cfg.JwtSecret = secret
	}
	if key := os.Getenv(envPrefix + "API_ENCRYPTION_KEY"); key != "" {
		cfg.APIEncryptionKey = key
	}
	// 验证码配置
	if mode := os.Getenv(envPrefix + "CAPTCHA_MODE"); mode != "" {
		if m, err := strconv.Atoi(mode); err == nil && m >= 1 && m <= 3 {
			cfg.CaptchaMode = m
		}
	}
	if siteKey := os.Getenv(envPrefix + "TURNSTILE_SITE_KEY"); siteKey != "" {
		cfg.TurnstileSiteKey = siteKey
	}
	if secretKey := os.Getenv(envPrefix + "TURNSTILE_SECRET_KEY"); secretKey != "" {
		cfg.TurnstileSecretKey = secretKey
	}
	if proxyLink := os.Getenv(envPrefix + "TURNSTILE_PROXY_LINK"); proxyLink != "" {
		cfg.TurnstileProxyLink = proxyLink
	}
}

// loadFromCmdLineInternal 从命令行参数加载（内部使用，不获取锁）
func loadFromCmdLineInternal(cfg *AppConfig) {
	if cmdConfig == nil {
		return
	}
	if cmdConfig.Port > 0 {
		cfg.Port = cmdConfig.Port
	}
	if cmdConfig.DBPath != "" {
		cfg.DBPath = cmdConfig.DBPath
	}
	if cmdConfig.LogPath != "" {
		cfg.LogPath = cmdConfig.LogPath
	}
	if cmdConfig.LogLevel != "" {
		cfg.LogLevel = cmdConfig.LogLevel
	}
}

// handleSecretsInternal 处理敏感配置（内部使用，不获取锁）
// 优先级：环境变量 > 配置文件 > 数据库 > 自动生成
// 如果用户通过环境变量或配置文件设置了值，也会同步到数据库，方便迁移部署和多机部署
func handleSecretsInternal(cfg *AppConfig) {
	// 检查是否有环境变量设置的 JWT Secret
	envJwtSecret := os.Getenv(envPrefix + "JWT_SECRET")

	// 处理 JWT Secret
	if cfg.JwtSecret != "" {
		// 用户配置了 JWT Secret（来自环境变量或配置文件）
		// 同步到数据库，方便迁移部署
		if secretSetterFunc != nil {
			dbSecret := ""
			if secretGetterFunc != nil {
				dbSecret = secretGetterFunc("jwt_secret")
			}
			// 只有当数据库值不同时才更新
			if dbSecret != cfg.JwtSecret {
				if err := secretSetterFunc("jwt_secret", cfg.JwtSecret); err != nil {
					log.Printf("同步 JWT Secret 到数据库失败: %v", err)
				} else {
					if envJwtSecret != "" {
						log.Println("已将环境变量 SUBLINK_JWT_SECRET 同步到数据库")
					} else {
						log.Println("已将配置文件中的 JWT Secret 同步到数据库")
					}
				}
			}
		}
	} else if secretGetterFunc != nil {
		// 从数据库读取
		if dbSecret := secretGetterFunc("jwt_secret"); dbSecret != "" {
			cfg.JwtSecret = dbSecret
		}
	}

	// 如果仍然没有值，自动生成
	if cfg.JwtSecret == "" {
		cfg.JwtSecret = generateRandomKey(32)
		log.Println("JWT Secret 未配置，已自动生成并保存到数据库")
		if secretSetterFunc != nil {
			if err := secretSetterFunc("jwt_secret", cfg.JwtSecret); err != nil {
				log.Printf("保存 JWT Secret 到数据库失败: %v", err)
			}
		}
	}

	// 检查是否有环境变量设置的 API 加密密钥
	envApiKey := os.Getenv(envPrefix + "API_ENCRYPTION_KEY")

	// 处理 API 加密密钥
	if cfg.APIEncryptionKey != "" {
		// 用户配置了 API 加密密钥（来自环境变量或配置文件）
		// 同步到数据库，方便迁移部署
		if secretSetterFunc != nil {
			dbKey := ""
			if secretGetterFunc != nil {
				dbKey = secretGetterFunc("api_encryption_key")
			}
			// 只有当数据库值不同时才更新
			if dbKey != cfg.APIEncryptionKey {
				if err := secretSetterFunc("api_encryption_key", cfg.APIEncryptionKey); err != nil {
					log.Printf("同步 API 加密密钥到数据库失败: %v", err)
				} else {
					if envApiKey != "" {
						log.Println("已将环境变量 SUBLINK_API_ENCRYPTION_KEY 同步到数据库")
					} else {
						log.Println("已将配置文件中的 API 加密密钥同步到数据库")
					}
				}
			}
		}
	} else if secretGetterFunc != nil {
		// 从数据库读取
		if dbKey := secretGetterFunc("api_encryption_key"); dbKey != "" {
			cfg.APIEncryptionKey = dbKey
		}
	}

	// 如果仍然没有值，自动生成
	if cfg.APIEncryptionKey == "" {
		cfg.APIEncryptionKey = generateRandomKey(32)
		log.Println("API 加密密钥未配置，已自动生成")
		if secretSetterFunc != nil {
			if err := secretSetterFunc("api_encryption_key", cfg.APIEncryptionKey); err != nil {
				log.Printf("保存 API 加密密钥到数据库失败: %v", err)
			} else {
				log.Println("API 加密密钥已保存到数据库")
			}
		}
	}
}

// generateRandomKey 生成随机密钥
func generateRandomKey(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用备用方案
		log.Printf("生成随机密钥失败: %v，使用备用方案", err)
		return "fallback-key-please-change-" + strconv.FormatInt(int64(os.Getpid()), 16)
	}
	return hex.EncodeToString(bytes)
}

// UpdateConfig 更新配置（用于运行时修改）
func UpdateConfig(updater func(*AppConfig)) {
	configMutex.Lock()
	defer configMutex.Unlock()

	if globalConfig == nil {
		globalConfig = &AppConfig{}
		applyDefaults(globalConfig)
	}
	updater(globalConfig)
}

// SaveToFile 保存配置到文件
func SaveToFile() error {
	configMutex.RLock()
	cfg := globalConfig
	configMutex.RUnlock()

	if cfg == nil {
		return nil
	}

	configPath := GetConfigFilePath()

	// 创建用于保存的配置（不包含敏感信息）
	saveCfg := &AppConfig{
		Port:             cfg.Port,
		ExpireDays:       cfg.ExpireDays,
		LoginFailCount:   cfg.LoginFailCount,
		LoginFailWindow:  cfg.LoginFailWindow,
		LoginBanDuration: cfg.LoginBanDuration,
	}

	// 生成 YAML 内容（包含注释）
	comment := `# SublinkPro 配置文件
# 敏感配置（jwt_secret, api_encryption_key）已存储在数据库中
# 如需覆盖，请使用环境变量 SUBLINK_JWT_SECRET 和 SUBLINK_API_ENCRYPTION_KEY
`
	data, err := yaml.Marshal(saveCfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, []byte(comment+string(data)), 0644)
}

// MigrateFromOldConfig 从旧配置迁移敏感数据到数据库
// 返回 true 表示有数据被迁移
func MigrateFromOldConfig() bool {
	configPath := GetConfigFilePath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}

	oldCfg := &AppConfig{}
	if err := yaml.Unmarshal(data, oldCfg); err != nil {
		return false
	}

	migrated := false

	// 迁移 JWT Secret
	if oldCfg.JwtSecret != "" && secretSetterFunc != nil {
		// 检查数据库中是否已有值
		if secretGetterFunc == nil || secretGetterFunc("jwt_secret") == "" {
			if err := secretSetterFunc("jwt_secret", oldCfg.JwtSecret); err == nil {
				log.Println("已将 JWT Secret 从配置文件迁移到数据库")
				migrated = true
			}
		}
	}

	// 迁移 API 加密密钥
	if oldCfg.APIEncryptionKey != "" && secretSetterFunc != nil {
		if secretGetterFunc == nil || secretGetterFunc("api_encryption_key") == "" {
			if err := secretSetterFunc("api_encryption_key", oldCfg.APIEncryptionKey); err == nil {
				log.Println("已将 API 加密密钥从配置文件迁移到数据库")
				migrated = true
			}
		}
	}

	return migrated
}
