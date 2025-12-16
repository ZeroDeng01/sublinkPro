package database

import (
	"os"
	"sublink/config"
	"sublink/utils"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// IsInitialized 标记数据库是否已初始化迁移
var IsInitialized bool

func InitSqlite() {
	// 获取数据库路径
	dbPath := config.GetDBPath()

	// 检查目录是否创建
	_, err := os.Stat(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dbPath, os.ModePerm)
		}
	}

	// SQLite 连接字符串带高性能参数
	// _busy_timeout: 锁等待超时毫秒数，避免立即返回 SQLITE_BUSY 错误
	// _journal_mode: WAL 模式允许并发读写
	// _synchronous: NORMAL 模式平衡性能和数据安全
	// _cache_size: 负数表示 KB，-64000 = 64MB 缓存
	// _foreign_keys: 启用外键约束
	dsn := dbPath + "/sublink.db?_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-64000&_foreign_keys=ON"

	// 配置 GORM，减少日志噪音
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		utils.Error("连接数据库失败: %v", err)
		return
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		utils.Error("获取底层数据库连接失败: %v", err)
	} else {
		// SQLite 推荐设置
		// MaxIdleConns: 保持的空闲连接数，减少连接开销
		// MaxOpenConns: 最大打开连接数，SQLite 推荐设为 1 以避免并发写入问题
		//               但由于我们使用 WAL 模式，可以适当放宽
		// ConnMaxLifetime: 连接最大生命周期
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		utils.Info("数据库连接池配置完成: MaxIdle=10, MaxOpen=100, MaxLifetime=1h")
	}

	DB = db
	utils.Info("数据库已初始化: %s (WAL模式)", dsn)
}
