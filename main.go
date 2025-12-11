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
	"sublink/database"
	"sublink/models"
	"sublink/routers"
	"sublink/services"
	"sublink/services/geoip"
	"sublink/services/mihomo"
	"sublink/services/sse"
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
	models.ConfigInit()
	config := models.ReadConfig() // 读取配置文件
	var port = config.Port        // 读取端口号
	// 获取版本号
	var IsVersion bool
	version = "dev"
	//读取VERSION文件获取版本
	versionData, err := versionFile.ReadFile("VERSION")
	if err == nil {
		version = strings.TrimSpace(string(versionData))
		fmt.Println("版本信息：", version)
	} else {
		fmt.Println("版本信息获取失败：", err)
	}

	flag.BoolVar(&IsVersion, "version", false, "显示版本号")
	flag.Parse()
	if IsVersion {
		fmt.Println(version)
		return
	}
	// 初始化数据库
	database.InitSqlite()
	// 执行数据库迁移
	models.RunMigrations()
	// 获取命令行参数
	args := os.Args
	// 如果长度小于2则没有接收到任何参数
	if len(args) < 2 {
		Run(port)
		return
	}
	// 命令行参数选择
	settingCmd := flag.NewFlagSet("setting", flag.ExitOnError)
	var username, password string
	settingCmd.StringVar(&username, "username", "", "设置账号")
	settingCmd.StringVar(&password, "password", "", "设置密码")
	settingCmd.IntVar(&port, "port", 8000, "修改端口")
	switch args[1] {
	// 解析setting命令标志
	case "setting":
		settingCmd.Parse(args[2:])
		fmt.Println(username, password)
		settings.ResetUser(username, password)
		return
	case "run":
		settingCmd.Parse(args[2:])
		models.SetConfig(models.Config{
			Port: port,
		}) // 设置端口
		Run(port)
	default:
		return

	}
}

func Run(port int) {
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
