package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sublink/api"
	"sublink/cache"
	"sublink/config"
	"sublink/database"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/routers"
	"sublink/services"
	"sublink/services/geoip"
	"sublink/services/mihomo"
	"sublink/services/scheduler"
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

// getContentType 根据文件扩展名返回对应的 MIME 类型
func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeTypes := map[string]string{
		".html":        "text/html; charset=utf-8",
		".css":         "text/css; charset=utf-8",
		".js":          "application/javascript; charset=utf-8",
		".json":        "application/json; charset=utf-8",
		".webmanifest": "application/manifest+json; charset=utf-8",
		".png":         "image/png",
		".jpg":         "image/jpeg",
		".jpeg":        "image/jpeg",
		".gif":         "image/gif",
		".svg":         "image/svg+xml",
		".ico":         "image/x-icon",
		".woff":        "font/woff",
		".woff2":       "font/woff2",
		".ttf":         "font/ttf",
		".eot":         "application/vnd.ms-fontobject",
	}
	if contentType, ok := mimeTypes[ext]; ok {
		return contentType
	}
	return "application/octet-stream"
}

// injectConfigToHTML 在 HTML 的 <head> 中注入前端配置脚本
// 配置通过 window.__SUBLINK_CONFIG__ 暴露给前端
func injectConfigToHTML(html string, basePath string) string {
	// 构建配置脚本
	configScript := fmt.Sprintf(`<script>window.__SUBLINK_CONFIG__={basePath:"%s"}</script>`, basePath)
	// 在 <head> 标签后注入
	return strings.Replace(html, "<head>", "<head>\n  "+configScript, 1)
}

// injectBasePathToManifest 修改 PWA manifest 中的 start_url
// 确保从主屏幕启动时导航到正确的路径
func injectBasePathToManifest(manifest string, basePath string) string {
	if basePath == "" {
		return manifest
	}
	// 将 "start_url": "/" 替换为 "start_url": "/basePath/"
	startURL := fmt.Sprintf(`"start_url":"%s/"`, basePath)
	result := strings.Replace(manifest, `"start_url":"/"`, startURL, 1)
	// 也处理可能的空格情况
	result = strings.Replace(result, `"start_url": "/"`, startURL, 1)
	return result
}

// isRootStaticFile 检查路径是否为根路径下的必要静态文件
// 这些文件需要在根路径下才能让 PWA 和其他功能正常工作
func isRootStaticFile(path string) bool {
	// PWA 相关文件（精确匹配）
	// 注：favicon.ico 在 /images/ 目录下，通过 /images StaticFS 挂载
	rootFiles := []string{
		"/sw.js",
		"/registerSW.js",
		"/manifest.webmanifest",
		"/icon-192.png",
		"/icon-512.png",
	}
	for _, f := range rootFiles {
		if path == f {
			return true
		}
	}
	// workbox 运行时文件（前缀匹配）
	if strings.HasPrefix(path, "/workbox-") && strings.HasSuffix(path, ".js") {
		return true
	}
	return false
}

func Templateinit() {
	// 设置template路径
	// 检查目录是否创建
	subFS, err := fs.Sub(Template, "template")
	if err != nil {
		utils.Error("加载内嵌模板失败: %v", err)
		return
	}
	entries, err := fs.ReadDir(subFS, ".")
	if err != nil {
		utils.Error("读取模板目录失败: %v", err)
		return
	}
	// 创建template目录
	_, err = os.Stat("./template")
	if os.IsNotExist(err) {
		err = os.Mkdir("./template", 0666)
		if err != nil {
			utils.Error("创建模板目录失败: %v", err)
			return
		}
	}
	// 写入默认模板
	for _, entry := range entries {
		_, err := os.Stat("./template/" + entry.Name())
		// 如果文件不存在则写入默认模板
		if os.IsNotExist(err) {
			data, err := fs.ReadFile(subFS, entry.Name())
			if err != nil {
				utils.Error("读取模板文件失败: %v", err)
				continue
			}
			err = os.WriteFile("./template/"+entry.Name(), data, 0666)
			if err != nil {
				utils.Error("写入模板文件失败: %v", err)
			}
		}
	}
}

func main() {
	// 定义命令行参数
	var (
		showVersion bool
		port        int
		dsn         string
		dbPath      string
		logPath     string
		logLevel    string
		configFile  string
	)

	// 全局参数
	flag.BoolVar(&showVersion, "version", false, "显示版本号")
	flag.BoolVar(&showVersion, "v", false, "显示版本号 (简写)")
	flag.IntVar(&port, "port", 0, "服务端口 (覆盖配置文件和环境变量)")
	flag.IntVar(&port, "p", 0, "服务端口 (简写)")
	flag.StringVar(&dsn, "dsn", "", "数据库 DSN (支持 sqlite/mysql/postgres)")
	flag.StringVar(&dbPath, "db", "", "本地数据目录 / SQLite 默认数据库目录")
	flag.StringVar(&logPath, "log", "", "日志目录路径")
	flag.StringVar(&logLevel, "log-level", "", "日志等级 (debug/info/warn/error/fatal)")
	flag.StringVar(&configFile, "config", "", "配置文件名 (相对于本地数据目录)")
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
			settingCmd.StringVar(&dsn, "dsn", "", "数据库 DSN (支持 sqlite/mysql/postgres)")
			settingCmd.StringVar(&dbPath, "db", "", "本地数据目录 / SQLite 默认数据库目录")
			settingCmd.StringVar(&logPath, "log", "", "日志目录路径")
			settingCmd.StringVar(&logLevel, "log-level", "", "日志等级 (debug/info/warn/error/fatal)")
			settingCmd.StringVar(&configFile, "config", "", "配置文件名")
			settingCmd.StringVar(&configFile, "c", "", "配置文件名 (简写)")
			settingCmd.Parse(os.Args[2:])

			// 初始化数据库目录和数据库
			if err := initDatabase(dsn, dbPath, logPath, logLevel, configFile, port); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			utils.Info("重置用户: %s", username)
			settings.ResetUser(username, password)
			return

		case "run":
			// 运行子命令
			runCmd := flag.NewFlagSet("run", flag.ExitOnError)
			runCmd.IntVar(&port, "port", 0, "服务端口")
			runCmd.IntVar(&port, "p", 0, "服务端口 (简写)")
			runCmd.StringVar(&dsn, "dsn", "", "数据库 DSN (支持 sqlite/mysql/postgres)")
			runCmd.StringVar(&dbPath, "db", "", "本地数据目录 / SQLite 默认数据库目录")
			runCmd.StringVar(&logPath, "log", "", "日志目录路径")
			runCmd.StringVar(&logLevel, "log-level", "", "日志等级 (debug/info/warn/error/fatal)")
			runCmd.StringVar(&configFile, "config", "", "配置文件名")
			runCmd.StringVar(&configFile, "c", "", "配置文件名 (简写)")
			runCmd.Parse(os.Args[2:])

			if err := initDatabase(dsn, dbPath, logPath, logLevel, configFile, port); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
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
	if err := initDatabase(dsn, dbPath, logPath, logLevel, configFile, port); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	Run()
}

// initDatabase 初始化数据库和配置
func initDatabase(dsn, dbPath, logPath, logLevel, configFile string, port int) error {
	// 设置命令行配置
	cmdCfg := &config.CommandLineConfig{
		Port:       port,
		DSN:        dsn,
		DBPath:     dbPath,
		LogPath:    logPath,
		LogLevel:   logLevel,
		ConfigFile: configFile,
	}
	config.SetCommandLineConfig(cmdCfg)

	// 先确保默认本地数据目录存在，以便初始化配置文件
	ensureDir(config.GetDBPath())

	// 初始化旧配置文件（向后兼容）
	models.ConfigInit()

	// 先加载基础配置，使配置文件中的 dsn/db_path/log_path 能参与启动
	config.LoadBootstrap()

	// 确保运行时需要的本地目录存在
	ensureDir(config.GetDBPath())
	ensureDir(config.GetLogPath())

	// 初始化数据库
	if err := database.Init(); err != nil {
		return err
	}
	if database.DB == nil {
		return fmt.Errorf("数据库初始化失败")
	}

	// 执行数据库迁移
	if err := models.RunMigrations(); err != nil {
		return err
	}

	// 初始化敏感配置访问器
	models.InitSecretAccessors()

	// 迁移旧配置中的敏感数据到数据库
	config.MigrateFromOldConfig()

	// 加载完整配置
	config.Load()
	return nil
}

// ensureDir 确保目录存在
func ensureDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			utils.Error("创建目录失败 %s: %v", path, err)
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
  -p, --port      服务端口 (默认: 8000)
  --dsn           数据库 DSN (支持 sqlite/mysql/postgres)
  -db             本地数据目录 / SQLite 默认数据库目录 (默认: ./db)
  -log            日志目录路径 (默认: ./logs)
  --log-level     日志等级 (debug/info/warn/error/fatal, 默认: info)
  -c, --config    配置文件名 (默认: config.yaml)
  -v, --version   显示版本号

环境变量:
  SUBLINK_PORT               服务端口
  SUBLINK_DSN                数据库 DSN (优先于 db_path，支持 sqlite/mysql/postgres)
  SUBLINK_DB_PATH            本地数据目录 / SQLite 默认数据库目录
  SUBLINK_LOG_PATH           日志目录路径
  SUBLINK_LOG_LEVEL          日志等级 (debug/info/warn/error/fatal)
  SUBLINK_JWT_SECRET         JWT签名密钥 (可选，自动生成)
  SUBLINK_API_ENCRYPTION_KEY API加密密钥 (可选，自动生成)
  SUBLINK_EXPIRE_DAYS        Token过期天数 (默认: 14)
  SUBLINK_LOGIN_FAIL_COUNT   登录失败次数限制 (默认: 5)
  SUBLINK_LOGIN_FAIL_WINDOW  登录失败窗口时间(分钟) (默认: 1)
  SUBLINK_LOGIN_BAN_DURATION 登录封禁时间(分钟) (默认: 10)
  SUBLINK_WEB_BASE_PATH      前端访问基础路径 (站点隐藏)
  SUBLINK_ADMIN_PASSWORD     初始管理员密码 (首次启动时设置)

配置优先级:
  命令行参数 > 环境变量 > 配置文件 > 数据库 > 默认值

示例:
  sublinkpro                           # 使用默认配置启动
  sublinkpro run -p 9000               # 指定端口启动
  sublinkpro run --log-level debug     # 开启调试日志
  sublinkpro run --db /data            # 指定本地数据目录
  sublinkpro run --dsn "sqlite:///data/sublink.db"
  sublinkpro run --dsn "mysql://user:pass@tcp(127.0.0.1:3306)/sublink?charset=utf8mb4&parseTime=True&loc=Local"
  sublinkpro run --dsn "postgres://user:pass@127.0.0.1:5432/sublink?sslmode=disable"
  sublinkpro setting -username admin -password newpass  # 重置用户`)
}

func Run() {
	// 获取配置
	cfg := config.Get()
	port := cfg.Port

	// 初始化日志系统
	utils.InitLogger(cfg.LogPath, cfg.LogLevel)

	// 演示模式启动提示
	if models.IsDemoMode() {
		utils.Info("🎭 ================================================")
		utils.Info("🎭 系统正在演示模式下运行")
		utils.Info("🎭 数据库: 内存模式（重启后数据丢失）")
		utils.Info("🎭 定时任务: 已禁用")
		utils.Info("🎭 敏感操作: 已禁用")
		utils.Info("🎭 ================================================")
	}

	// 打印版本信息
	utils.Info("启动 SublinkPro 版本: %s", version)
	utils.Info("日志等级: %s", utils.GetLogLevel())

	// 初始化gin框架
	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		utils.Warn("设置 Gin trusted proxies 失败: %v", err)
	}
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
		utils.Warn("初始化 GeoIP 数据库失败: %v", err)
	}

	// 如果 GeoIP 数据库不可用，异步尝试自动下载
	if !geoip.IsAvailable() {
		go api.AutoDownloadGeoIP()
	}

	// 启动 AccessKey 清理定时任务
	models.StartAccessKeyCleanupScheduler()

	// 启动SSE服务
	go sse.GetSSEBroker().Listen()

	// 初始化并启动定时任务管理器（演示模式下跳过）
	var sch *scheduler.SchedulerManager
	if !models.IsDemoMode() {
		sch = scheduler.GetSchedulerManager()
		sch.Start()
	}

	if err := models.InitNodeCache(); err != nil {
		utils.Error("加载节点到缓存失败: %v", err)
	}
	if err := models.InitSettingCache(); err != nil {
		utils.Error("加载系统设置到缓存失败: %v", err)
	}
	if err := models.InitUserCache(); err != nil {
		utils.Error("加载用户到缓存失败: %v", err)
	}
	if err := models.InitScriptCache(); err != nil {
		utils.Error("加载脚本到缓存失败: %v", err)
	}
	if err := models.InitAirportCache(); err != nil {
		utils.Error("加载机场到缓存失败: %v", err)
	}
	if err := models.InitGroupAirportSortCache(); err != nil {
		utils.Error("加载分组机场排序到缓存失败: %v", err)
	}
	if err := models.InitAccessKeyCache(); err != nil {
		utils.Error("加载AccessKey到缓存失败: %v", err)
	}
	if err := models.InitNodeCheckProfileCache(); err != nil {
		utils.Error("加载节点检测策略到缓存失败: %v", err)
	}
	if err := models.InitSubLogsCache(); err != nil {
		utils.Error("加载订阅日志到缓存失败: %v", err)
	}
	if err := models.InitSubcriptionCache(); err != nil {
		utils.Error("加载订阅到缓存失败: %v", err)
	}
	if err := models.InitTemplateCache(); err != nil {
		utils.Error("加载模板到缓存失败: %v", err)
	}
	// 初始化模板内容缓存
	cache.InitTemplateContentCache()
	if err := models.InitTagCache(); err != nil {
		utils.Error("加载标签到缓存失败: %v", err)
	}
	if err := models.InitTagRuleCache(); err != nil {
		utils.Error("加载标签规则到缓存失败: %v", err)
	}
	// 设置标签组标签查询回调（用于节点重命名的 $TagGroup(xxx) 变量）
	utils.SetTagGroupTagsFunc(models.GetTagNamesByGroupName)
	if err := models.InitTaskCache(); err != nil {
		utils.Error("加载任务到缓存失败: %v", err)
	}
	if err := models.InitIPInfoCache(); err != nil {
		utils.Error("加载IP信息到缓存失败: %v", err)
	}
	if err := models.InitHostCache(); err != nil {
		utils.Error("加载Host到缓存失败: %v", err)
	}
	if err := models.InitSubscriptionShareCache(); err != nil {
		utils.Error("加载订阅分享到缓存失败: %v", err)
	}
	if err := models.InitChainRuleCache(); err != nil {
		utils.Error("加载链式代理规则到缓存失败: %v", err)
	}

	// 注册Host变更回调：当Host模块数据变更时自动同步到mihomo resolver
	// 这样所有使用代理的功能（测速、订阅导入、Telegram等）都遵循Host设置
	models.RegisterHostChangeCallback(func() {
		if err := mihomo.SyncHostsFromDB(); err != nil {
			utils.Warn("Host变更同步到mihomo失败: %v", err)
		}
	})
	// 首次同步Host配置到mihomo resolver
	if err := mihomo.SyncHostsFromDB(); err != nil {
		utils.Warn("初始化Host同步到mihomo失败: %v", err)
	}

	// 初始化去重字段元数据缓存（通过反射扫描协议结构体和Node模型）
	protocol.InitProtocolMeta()
	models.InitNodeFieldsMeta()

	// 启动时清理过期的记住密码令牌
	models.CleanAllExpiredTokens()

	// 初始化任务管理器
	services.InitTaskManager()

	// 初始化 scheduler 包的依赖（必须在 InitTaskManager 之后）
	services.InitSchedulerDependencies()

	// 初始化 Telegram 机器人 (异步)
	go func() {
		utils.Debug("正在异步初始化 Telegram 机器人...")
		if err := telegram.InitBot(); err != nil {
			utils.Warn("初始化 Telegram 机器人失败: %v", err)
		}
	}()

	// 设置 Telegram 服务包装器和 SSE 通知函数
	services.InitTelegramWrapper()
	sse.TelegramNotifier = telegram.SendNotification

	// 从数据库加载定时任务（演示模式下跳过）
	if !models.IsDemoMode() && sch != nil {
		err := sch.LoadFromDatabase()
		if err != nil {
			utils.Error("加载定时任务失败: %v", err)
		}
	}

	// 演示模式：初始化演示数据
	if models.IsDemoMode() {
		models.InitDemoData()
	}
	// 安装中间件

	// 获取前端基础路径配置
	webBasePath := config.GetWebBasePath()
	if webBasePath != "" {
		utils.Info("前端基础路径: %s", webBasePath)
	}

	// 设置静态资源路径
	// 生产环境才启用内嵌静态文件服务
	if StaticFiles != nil {
		staticFiles, err := fs.Sub(StaticFiles, "static")
		if err != nil {
			utils.Error("加载静态文件失败: %v", err)
		} else {
			// 创建注入配置的 index.html 处理函数
			serveIndexHTML := func(c *gin.Context) {
				data, err := fs.ReadFile(staticFiles, "index.html")
				if err != nil {
					c.Error(err)
					return
				}
				// 注入配置脚本到 HTML
				html := injectConfigToHTML(string(data), webBasePath)
				c.Data(200, "text/html; charset=utf-8", []byte(html))
			}

			// 静态资源始终挂载到根路径
			// 因为 Vite 构建后的 index.html 中资源路径是绝对的（/assets/xxx）
			assetsFiles, _ := fs.Sub(staticFiles, "assets")
			r.StaticFS("/assets", http.FS(assetsFiles))
			imagesFiles, _ := fs.Sub(staticFiles, "images")
			r.StaticFS("/images", http.FS(imagesFiles))

			// basePath 路由处理
			if webBasePath != "" {
				// basePath 入口（带斜杠和不带斜杠）
				r.GET(webBasePath, serveIndexHTML)
				r.GET(webBasePath+"/", serveIndexHTML)
				// 根路径返回 404（站点隐藏）
				r.GET("/", func(c *gin.Context) {
					c.String(404, "Not Found")
				})
			} else {
				// 默认情况：根路径返回 index.html
				r.GET("/", serveIndexHTML)
			}
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
	routers.Backup(r)
	routers.Script(r)
	routers.SSE(r)
	routers.Settings(r)
	routers.Tag(r)
	routers.Tasks(r)
	routers.GeoIP(r)
	routers.Host(r)
	routers.Share(r)
	routers.Airport(r)
	routers.GroupSort(r)
	routers.NodeCheck(r)

	// 处理前端路由 (SPA History Mode) 和静态文件
	// 必须在所有 backend 路由注册之后注册
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 请求直接返回 404
		if strings.HasPrefix(path, "/api/") {
			c.JSON(404, gin.H{"error": "API route not found"})
			return
		}

		// 订阅请求（/c/）由路由处理，这里不应该到达
		// 如果到达这里说明订阅链接无效
		if strings.HasPrefix(path, "/c/") {
			c.JSON(404, gin.H{"error": "Subscription not found"})
			return
		}

		// 如果设置了 webBasePath，检查路径是否在 basePath 下
		// 注意：/assets/ 和 /images/ 由 StaticFS 处理，不会到达这里
		if webBasePath != "" {
			// 不在 basePath 下的请求返回 404（站点隐藏）
			// 排除静态资源目录（它们已由 StaticFS 处理）
			if !strings.HasPrefix(path, webBasePath+"/") && path != webBasePath {
				// 允许访问根路径下的静态文件（sw.js, manifest.webmanifest 等）
				// 这些文件需要在根路径下才能让 PWA 正常工作
				if isRootStaticFile(path) {
					// 继续处理，不返回 404
				} else {
					c.String(404, "Not Found")
					return
				}
			} else {
				// 去掉 basePath 前缀，得到相对路径
				path = strings.TrimPrefix(path, webBasePath)
			}
		}

		// 生产环境：尝试从 embed 文件系统提供静态文件
		if StaticFiles != nil {
			staticFiles, err := fs.Sub(StaticFiles, "static")
			if err != nil {
				c.String(500, "Internal Server Error")
				return
			}

			// 去掉路径开头的斜杠
			filePath := strings.TrimPrefix(path, "/")
			if filePath == "" {
				filePath = "index.html"
			}

			// 尝试读取请求的文件
			data, err := fs.ReadFile(staticFiles, filePath)
			if err == nil {
				// 文件存在，根据扩展名设置 Content-Type
				contentType := getContentType(filePath)
				// Service Worker 特殊处理
				if filePath == "sw.js" {
					c.Header("Service-Worker-Allowed", "/")
				}
				// 如果是 HTML 文件，注入配置
				if strings.HasSuffix(filePath, ".html") {
					html := injectConfigToHTML(string(data), webBasePath)
					c.Data(200, contentType, []byte(html))
				} else if filePath == "manifest.webmanifest" {
					// PWA manifest 特殊处理：修改 start_url
					manifest := injectBasePathToManifest(string(data), webBasePath)
					c.Data(200, contentType, []byte(manifest))
				} else {
					c.Data(200, contentType, data)
				}
				return
			}

			// 文件不存在，返回 index.html (SPA fallback)
			data, err = fs.ReadFile(staticFiles, "index.html")
			if err != nil {
				c.String(404, "Index file not found")
				return
			}
			// 注入配置到 SPA fallback
			html := injectConfigToHTML(string(data), webBasePath)
			c.Data(200, "text/html; charset=utf-8", []byte(html))
		} else {
			// 开发环境 fallback
			c.File("./static/index.html")
		}
	})

	// 启动服务
	r.Run(fmt.Sprintf("0.0.0.0:%d", port))
}
