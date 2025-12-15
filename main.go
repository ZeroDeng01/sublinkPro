package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"sublink/config"
	"sublink/database"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/routers"
	"sublink/services"
	"sublink/services/geoip"
	"sublink/services/mihomo"
	"sublink/services/sse"
	"sublink/services/telegram"
	"sublink/settings"
	"sublink/utils"

	"github.com/gin-gonic/gin"
	"github.com/metacubex/mihomo/constant"
)

//go:embed template
var Template embed.FS

//go:embed VERSION
var versionFile embed.FS

var version string

func Templateinit() {
	// 设置template路径
	// 检查目录是否创建
	subFS, err := fs.Sub(Template, "template")
	if err != nil {
		log.Println(err)
		return // 如果出错，直接返回
	}
	entries, err := fs.ReadDir(subFS, ".")
	if err != nil {
		log.Println(err)
		return // 如果出错，直接返回
	}
	// 创建template目录
	_, err = os.Stat("./template")
	if os.IsNotExist(err) {
		err = os.Mkdir("./template", 0666)
		if err != nil {
			log.Println(err)
			return
		}
	}
	// 写入默认模板
	for _, entry := range entries {
		_, err := os.Stat("./template/" + entry.Name())
		//如果文件不存在则写入默认模板
		if os.IsNotExist(err) {
			data, err := fs.ReadFile(subFS, entry.Name())
			if err != nil {
				log.Println(err)
				continue
			}
			err = os.WriteFile("./template/"+entry.Name(), data, 0666)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func main() {
	// 定义命令行参数
	var (
		showVersion bool
		port        int
		dbPath      string
		logPath     string
		configFile  string
	)

	// 全局参数
	flag.BoolVar(&showVersion, "version", false, "显示版本号")
	flag.BoolVar(&showVersion, "v", false, "显示版本号 (简写)")
	flag.IntVar(&port, "port", 0, "服务端口 (覆盖配置文件和环境变量)")
	flag.IntVar(&port, "p", 0, "服务端口 (简写)")
	flag.StringVar(&dbPath, "db", "", "数据库目录路径")
	flag.StringVar(&logPath, "log", "", "日志目录路径")
	flag.StringVar(&configFile, "config", "", "配置文件名 (相对于数据库目录)")
	flag.StringVar(&configFile, "c", "", "配置文件名 (简写)")

	// 获取版本号
	version = "dev"
	versionData, err := versionFile.ReadFile("VERSION")
	if err == nil {
		version = strings.TrimSpace(string(versionData))
	}

	// 处理子命令
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "setting":
			// 用户设置子命令
			settingCmd := flag.NewFlagSet("setting", flag.ExitOnError)
			var username, password string
			settingCmd.StringVar(&username, "username", "", "设置账号")
			settingCmd.StringVar(&password, "password", "", "设置密码")
			settingCmd.Parse(os.Args[2:])

			// 初始化数据库目录和数据库
			initDatabase(dbPath, logPath, configFile, port)

			fmt.Println(username, password)
			settings.ResetUser(username, password)
			return

		case "run":
			// 运行子命令
			runCmd := flag.NewFlagSet("run", flag.ExitOnError)
			runCmd.IntVar(&port, "port", 0, "服务端口")
			runCmd.IntVar(&port, "p", 0, "服务端口 (简写)")
			runCmd.StringVar(&dbPath, "db", "", "数据库目录路径")
			runCmd.StringVar(&logPath, "log", "", "日志目录路径")
			runCmd.StringVar(&configFile, "config", "", "配置文件名")
			runCmd.StringVar(&configFile, "c", "", "配置文件名 (简写)")
			runCmd.Parse(os.Args[2:])

			initDatabase(dbPath, logPath, configFile, port)
			Run()
			return

		case "version", "-version", "--version", "-v":
			fmt.Println(version)
			return

		case "help", "-help", "--help", "-h":
			printHelp()
			return
		}
	}

	// 解析全局参数
	flag.Parse()

	if showVersion {
		fmt.Println(version)
		return
	}

	// 默认运行模式
	initDatabase(dbPath, logPath, configFile, port)
	Run()
}

// initDatabase 初始化数据库和配置
func initDatabase(dbPath, logPath, configFile string, port int) {
	// 设置命令行配置
	cmdCfg := &config.CommandLineConfig{
		Port:       port,
		DBPath:     dbPath,
		LogPath:    logPath,
		ConfigFile: configFile,
	}
	config.SetCommandLineConfig(cmdCfg)

	// 确保目录存在
	ensureDir(config.GetDBPath())
	ensureDir(config.GetLogPath())

	// 初始化旧配置文件（向后兼容）
	models.ConfigInit()

	// 初始化数据库
	database.InitSqlite()

	// 执行数据库迁移
	models.RunMigrations()

	// 初始化敏感配置访问器
	models.InitSecretAccessors()

	// 迁移旧配置中的敏感数据到数据库
	config.MigrateFromOldConfig()

	// 加载完整配置
	config.Load()
}

// ensureDir 确保目录存在
func ensureDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Printf("创建目录失败 %s: %v", path, err)
		}
	}
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println(`SublinkPro - 代理订阅管理与转换工具

使用方法:
  sublinkpro [命令] [选项]

命令:
  run           启动服务
  setting       用户设置
  version       显示版本号
  help          显示帮助信息

全局选项:
  -p, --port    服务端口 (默认: 8000)
  -db           数据库目录路径 (默认: ./db)
  -log          日志目录路径 (默认: ./logs)
  -c, --config  配置文件名 (默认: config.yaml)
  -v, --version 显示版本号

环境变量:
  SUBLINK_PORT              服务端口
  SUBLINK_DB_PATH           数据库目录路径
  SUBLINK_LOG_PATH          日志目录路径
  SUBLINK_JWT_SECRET        JWT签名密钥 (可选，自动生成)
  SUBLINK_API_ENCRYPTION_KEY API加密密钥 (可选，自动生成)
  SUBLINK_EXPIRE_DAYS       Token过期天数 (默认: 14)
  SUBLINK_LOGIN_FAIL_COUNT  登录失败次数限制 (默认: 5)
  SUBLINK_LOGIN_FAIL_WINDOW  登录失败窗口时间(分钟) (默认: 1)
  SUBLINK_LOGIN_BAN_DURATION 登录封禁时间(分钟) (默认: 10)
  SUBLINK_ADMIN_PASSWORD    初始管理员密码 (首次启动时设置)

配置优先级:
  命令行参数 > 环境变量 > 配置文件 > 数据库 > 默认值

示例:
  sublinkpro                       # 使用默认配置启动
  sublinkpro run -p 9000           # 指定端口启动
  sublinkpro run --db /data/db     # 指定数据库目录
  sublinkpro setting -username admin -password newpass  # 重置用户
`)
}

func Run() {
	// 获取配置
	cfg := config.Get()
	port := cfg.Port

	// 打印版本信息
	fmt.Println("版本信息：", version)

	// 初始化gin框架
	r := gin.Default()
	// 初始化日志配置
	utils.Loginit()
	// 初始化模板
	Templateinit()

	// 初始化代理客户端函数
	utils.GetMihomoAdapterFunc = func(nodeLink string) (constant.Proxy, error) {
		return mihomo.GetMihomoAdapter(nodeLink)
	}
	utils.GetBestProxyNodeFunc = func() (string, string, error) {
		node, err := models.GetBestProxyNode()
		if err != nil {
			return "", "", err
		}
		if node == nil {
			return "", "", nil
		}
		return node.Link, node.Name, nil
	}

	// 初始化 GeoIP 数据库
	if err := geoip.InitGeoIP(); err != nil {
		log.Printf("初始化 GeoIP 数据库失败: %v", err)
	}

	// 启动 AccessKey 清理定时任务
	models.StartAccessKeyCleanupScheduler()

	// 启动SSE服务
	go sse.GetSSEBroker().Listen()

	// 初始化并启动定时任务管理器
	scheduler := services.GetSchedulerManager()
	scheduler.Start()

	if err := models.InitNodeCache(); err != nil {
		log.Println("加载节点到缓存失败: %v", err)
	}
	if err := models.InitSettingCache(); err != nil {
		log.Println("加载系统设置到缓存失败: %v", err)
	}
	if err := models.InitUserCache(); err != nil {
		log.Println("加载用户到缓存失败: %v", err)
	}
	if err := models.InitScriptCache(); err != nil {
		log.Println("加载脚本到缓存失败: %v", err)
	}
	if err := models.InitSubSchedulerCache(); err != nil {
		log.Println("加载订阅调度到缓存失败: %v", err)
	}
	if err := models.InitAccessKeyCache(); err != nil {
		log.Println("加载AccessKey到缓存失败: %v", err)
	}
	if err := models.InitSubLogsCache(); err != nil {
		log.Println("加载订阅日志到缓存失败: %v", err)
	}
	if err := models.InitSubcriptionCache(); err != nil {
		log.Println("加载订阅到缓存失败: %v", err)
	}
	if err := models.InitTemplateCache(); err != nil {
		log.Println("加载模板到缓存失败: %v", err)
	}
	if err := models.InitTagCache(); err != nil {
		log.Println("加载标签到缓存失败: %v", err)
	}
	if err := models.InitTagRuleCache(); err != nil {
		log.Println("加载标签规则到缓存失败: %v", err)
	}
	if err := models.InitTaskCache(); err != nil {
		log.Println("加载任务到缓存失败: %v", err)
	}
	if err := models.InitIPInfoCache(); err != nil {
		log.Println("加载IP信息到缓存失败: %v", err)
	}

	// 初始化去重字段元数据缓存（通过反射扫描协议结构体和Node模型）
	protocol.InitProtocolMeta()
	models.InitNodeFieldsMeta()

	// 启动时清理过期的记住密码令牌
	models.CleanAllExpiredTokens()

	// 初始化任务管理器
	services.InitTaskManager()

	// 初始化 Telegram 机器人 (异步)
	go func() {
		log.Println("正在异步初始化 Telegram 机器人...")
		if err := telegram.InitBot(); err != nil {
			log.Printf("初始化 Telegram 机器人失败: %v", err)
		}
	}()

	// 设置 Telegram 服务包装器和 SSE 通知函数
	services.InitTelegramWrapper()
	sse.TelegramNotifier = telegram.SendNotification

	// 从数据库加载定时任务
	err := scheduler.LoadFromDatabase()
	if err != nil {
		log.Printf("加载定时任务失败: %v", err)
	}
	// 安装中间件

	// 设置静态资源路径
	// 生产环境才启用内嵌静态文件服务
	if StaticFiles != nil {
		staticFiles, err := fs.Sub(StaticFiles, "static")
		if err != nil {
			log.Println(err)
		} else {
			// 增加assets目录的静态服务
			assetsFiles, _ := fs.Sub(staticFiles, "assets")
			r.StaticFS("/assets", http.FS(assetsFiles))
			// 增加images目录的静态服务 (public文件夹)
			imagesFiles, _ := fs.Sub(staticFiles, "images")
			r.StaticFS("/images", http.FS(imagesFiles))
			r.GET("/favicon.svg", func(c *gin.Context) {
				c.FileFromFS("favicon.svg", http.FS(staticFiles))
			})
			r.GET("/", func(c *gin.Context) {
				data, err := fs.ReadFile(staticFiles, "index.html")
				if err != nil {
					c.Error(err)
					return
				}
				c.Data(200, "text/html", data)
			})
		}
	}
	// 注册路由
	routers.User(r)
	routers.AccessKey(r)
	routers.Subcription(r)
	routers.Nodes(r)
	routers.Clients(r)
	routers.Total(r)
	routers.Templates(r)
	routers.Version(r, version)
	routers.SubScheduler(r)
	routers.Backup(r)
	routers.Script(r)
	routers.SSE(r)
	routers.Settings(r)
	routers.Tag(r)
	routers.Tasks(r)

	// 处理前端路由 (SPA History Mode)
	// 必须在所有 backend 路由注册之后注册
	r.NoRoute(func(c *gin.Context) {
		// 如果是 API 请求，返回 404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"error": "API route not found"})
			return
		}

		// 否则返回 index.html
		if StaticFiles != nil {
			// 从 embed 文件系统中读取
			staticFiles, err := fs.Sub(StaticFiles, "static")
			if err != nil {
				c.String(404, "Internal Server Error")
				return
			}
			data, err := fs.ReadFile(staticFiles, "index.html")
			if err != nil {
				c.String(404, "Index file not found")
				return
			}
			c.Data(200, "text/html", data)
		} else {
			// 本地开发环境 fallback (假设 static 目录在当前路径)
			c.File("./static/index.html")
		}
	})

	// 启动服务
	r.Run(fmt.Sprintf("0.0.0.0:%d", port))
}
