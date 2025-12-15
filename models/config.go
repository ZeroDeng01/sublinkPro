package models

import (
	"fmt"
	"log"
	"os"
	"sublink/utils"

	"gopkg.in/yaml.v3"
)

// type Config struct {
// 	ID    int
// 	Key   string
// 	Value string
// }

// Config 配置结构体
type Config struct {
	JwtSecret        string `yaml:"jwt_secret"`         // JWT密钥
	APIEncryptionKey string `yaml:"api_encryption_key"` // API加密密钥
	ExpireDays       int    `yaml:"expire_days"`        // 过期天数
	Port             int    `yaml:"port"`               // 端口号
	LoginFailCount   int    `yaml:"login_fail_count"`   // 登录失败次数限制
	LoginFailWindow  int    `yaml:"login_fail_window"`  // 登录失败窗口时间(分钟)
	LoginBanDuration int    `yaml:"login_ban_duration"` // 登录失败封禁时间(分钟)
}

var comment string = `# jwt_secret: JWT密钥
# expire_days: token 过期天数
# port: 启动端口
# login_fail_count: 登录失败次数限制 (默认5)
# login_fail_window: 登录失败统计窗口时间(分钟, 默认1)
# login_ban_duration: 登录失败封禁时间(分钟, 默认10)
`

// 初始化配置
func ConfigInit() {
	// 检查配置文件是否存在
	if _, err := os.Stat("./db/config.yaml"); os.IsNotExist(err) {

		// 如果不存在则创建默认配置文件
		defaultConfig := Config{
			JwtSecret:        utils.RandString(31), // 生成随机JWT密钥
			APIEncryptionKey: utils.RandString(31), // 生成随机API加密密钥
			ExpireDays:       14,
			Port:             8000, // 默认端口
			LoginFailCount:   5,    // 默认5次
			LoginFailWindow:  1,    // 默认1分钟
			LoginBanDuration: 10,   // 默认封禁10分钟
		}

		// 生成yaml文件
		data, err := yaml.Marshal(&defaultConfig)
		if err != nil {
			log.Println("生成默认配置文件失败:", err)
			return
		}
		data = []byte(comment + string(data)) // 添加注释
		err = os.WriteFile("./db/config.yaml", data, 0644)
		if err != nil {
			fmt.Println("写入文件失败:", err)
			return
		}
		log.Println("配置文件不存在，已创建默认配置文件")
	}
}

// 读取配置
func ReadConfig() Config {
	file, err := os.ReadFile("./db/config.yaml")
	if err != nil {
		log.Println(err)
	}
	cfg := Config{}
	yaml.Unmarshal(file, &cfg)

	// Set defaults if missing (for existing configs)
	if cfg.LoginFailCount == 0 {
		cfg.LoginFailCount = 5
	}
	if cfg.LoginFailWindow == 0 {
		cfg.LoginFailWindow = 1
	}
	if cfg.LoginBanDuration == 0 {
		cfg.LoginBanDuration = 10
	}

	return cfg
}

// 设置配置
func SetConfig(newCfg Config) {
	oldCfg := ReadConfig() // 读取旧的配置文件
	// 覆盖新的字段
	if newCfg.JwtSecret != "" {
		oldCfg.JwtSecret = newCfg.JwtSecret
	}
	if newCfg.APIEncryptionKey != "" {
		oldCfg.APIEncryptionKey = newCfg.APIEncryptionKey
	}
	if newCfg.ExpireDays != 0 {
		oldCfg.ExpireDays = newCfg.ExpireDays
	}
	if newCfg.Port != 0 {
		oldCfg.Port = newCfg.Port
	}
	if newCfg.LoginFailCount != 0 {
		oldCfg.LoginFailCount = newCfg.LoginFailCount
	}
	if newCfg.LoginFailWindow != 0 {
		oldCfg.LoginFailWindow = newCfg.LoginFailWindow
	}
	if newCfg.LoginBanDuration != 0 {
		oldCfg.LoginBanDuration = newCfg.LoginBanDuration
	}
	// 写入文件
	data, err := yaml.Marshal(&oldCfg)
	if err != nil {
		log.Println(err)
	}
	data = []byte(comment + string(data)) // 添加注释
	os.WriteFile("./db/config.yaml", data, 0644)
}
