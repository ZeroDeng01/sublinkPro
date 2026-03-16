package database

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sublink/config"
	"sublink/utils"
	"time"

	"github.com/glebarez/sqlite"
	mysqlcfg "github.com/go-sql-driver/mysql"
	gmysql "gorm.io/driver/mysql"
	gpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DialectSQLite   = "sqlite"
	DialectMySQL    = "mysql"
	DialectPostgres = "postgres"

	defaultSQLiteFilename = "sublink.db"
)

var DB *gorm.DB

// Dialect 表示当前已初始化的数据库方言
var Dialect = DialectSQLite

// IsInitialized 标记数据库是否已初始化迁移
var IsInitialized bool

type dbConnectionInfo struct {
	Dialect            string
	Dialector          gorm.Dialector
	BootstrapDialector gorm.Dialector
	DBName             string
	RedactedDSN        string
}

// isDemoMode 判断是否为演示模式（避免循环导入）
func isDemoMode() bool {
	val := os.Getenv("SUBLINK_DEMO_MODE")
	return strings.EqualFold(val, "true") || val == "1"
}

// Init 初始化数据库连接，支持 sqlite/mysql/postgres。
func Init() error {
	// 演示模式使用内存数据库
	if isDemoMode() {
		return initMemorySqlite()
	}

	rawDSN := strings.TrimSpace(config.GetDSN())
	if rawDSN == "" {
		rawDSN = defaultSQLiteDSN(config.GetDBPath())
	}

	db, info, err := openDatabase(rawDSN, newGormConfig())
	if err != nil {
		DB = nil
		utils.Error("连接数据库失败: %v", err)
		return err
	}

	configureConnectionPool(db, info.Dialect)

	DB = db
	Dialect = info.Dialect
	utils.Info("数据库已初始化: driver=%s, dsn=%s", info.Dialect, info.RedactedDSN)
	return nil
}

// InitSqlite 为旧调用方保留兼容入口。
func InitSqlite() error {
	return Init()
}

func IsSQLite() bool {
	return Dialect == DialectSQLite
}

func IsMySQL() bool {
	return Dialect == DialectMySQL
}

func IsPostgres() bool {
	return Dialect == DialectPostgres
}

func defaultSQLiteDSN(dbPath string) string {
	dbFile := filepath.Join(dbPath, defaultSQLiteFilename)
	return "sqlite://" + dbFile
}

func openDatabase(rawDSN string, gormConfig *gorm.Config) (*gorm.DB, *dbConnectionInfo, error) {
	info, err := buildConnectionInfo(rawDSN)
	if err != nil {
		return nil, nil, err
	}

	db, err := gorm.Open(info.Dialector, gormConfig)
	if err != nil {
		if info.BootstrapDialector != nil && isMissingDatabaseError(info.Dialect, err) {
			if createErr := createDatabase(info, gormConfig); createErr != nil {
				return nil, nil, fmt.Errorf("%w; 自动建库失败: %v", err, createErr)
			}
			db, err = gorm.Open(info.Dialector, gormConfig)
			if err == nil {
				return db, info, nil
			}
		}
		return nil, nil, err
	}

	return db, info, nil
}

func buildConnectionInfo(rawDSN string) (*dbConnectionInfo, error) {
	rawDSN = strings.TrimSpace(rawDSN)
	if rawDSN == "" {
		return nil, fmt.Errorf("数据库 DSN 不能为空")
	}

	switch {
	case isMySQLDSN(rawDSN):
		return buildMySQLConnectionInfo(rawDSN)
	case isPostgresDSN(rawDSN):
		return buildPostgresConnectionInfo(rawDSN)
	case isSQLiteDSN(rawDSN):
		return buildSQLiteConnectionInfo(rawDSN)
	default:
		return nil, fmt.Errorf("不支持的数据库 DSN: %s", rawDSN)
	}
}

func isMySQLDSN(rawDSN string) bool {
	return strings.HasPrefix(strings.ToLower(rawDSN), "mysql://")
}

func isPostgresDSN(rawDSN string) bool {
	lower := strings.ToLower(rawDSN)
	return strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://")
}

func isSQLiteDSN(rawDSN string) bool {
	lower := strings.ToLower(rawDSN)
	return strings.HasPrefix(lower, "sqlite://") ||
		strings.HasPrefix(lower, "file:") ||
		strings.HasPrefix(rawDSN, "/") ||
		strings.HasPrefix(rawDSN, "./") ||
		strings.HasPrefix(rawDSN, "../")
}

func buildSQLiteConnectionInfo(rawDSN string) (*dbConnectionInfo, error) {
	driverDSN := strings.TrimSpace(rawDSN)
	if strings.HasPrefix(strings.ToLower(driverDSN), "sqlite://") {
		driverDSN = driverDSN[len("sqlite://"):]
	}
	driverDSN = mergeQueryParams(driverDSN, map[string]string{
		"_busy_timeout": "5000",
		"_journal_mode": "WAL",
		"_synchronous":  "NORMAL",
		"_cache_size":   "-64000",
		"_foreign_keys": "ON",
	})

	if err := ensureSQLiteDir(driverDSN); err != nil {
		return nil, err
	}

	return &dbConnectionInfo{
		Dialect:     DialectSQLite,
		Dialector:   sqlite.Open(driverDSN),
		DBName:      baseSQLiteName(driverDSN),
		RedactedDSN: driverDSN,
	}, nil
}

func baseSQLiteName(driverDSN string) string {
	base := driverDSN
	if idx := strings.Index(base, "?"); idx >= 0 {
		base = base[:idx]
	}
	return filepath.Base(base)
}

func ensureSQLiteDir(driverDSN string) error {
	base := driverDSN
	if idx := strings.Index(base, "?"); idx >= 0 {
		base = base[:idx]
	}

	if base == "" || strings.HasPrefix(base, "file::memory:") || strings.HasPrefix(base, ":memory:") {
		return nil
	}

	if strings.HasPrefix(base, "file:") {
		return nil
	}

	dir := filepath.Dir(base)
	if dir == "." || dir == "" {
		return nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建 SQLite 目录失败: %w", err)
	}
	return nil
}

func buildMySQLConnectionInfo(rawDSN string) (*dbConnectionInfo, error) {
	stripped := strings.TrimSpace(rawDSN[len("mysql://"):])
	if stripped == "" {
		return nil, fmt.Errorf("MySQL DSN 不能为空")
	}

	var (
		cfg *mysqlcfg.Config
		err error
	)

	if !strings.Contains(stripped, "@tcp(") && !strings.Contains(stripped, "@unix(") {
		parsed, err := url.Parse(rawDSN)
		if err != nil {
			return nil, fmt.Errorf("解析 MySQL DSN 失败: %w", err)
		}

		if parsed.Host == "" {
			return nil, fmt.Errorf("MySQL DSN 缺少主机地址")
		}

		dbName := strings.TrimPrefix(parsed.Path, "/")
		if dbName == "" {
			return nil, fmt.Errorf("MySQL DSN 缺少数据库名")
		}

		cfg = mysqlcfg.NewConfig()
		cfg.User = parsed.User.Username()
		password, _ := parsed.User.Password()
		cfg.Passwd = password
		cfg.Net = "tcp"
		cfg.Addr = parsed.Host
		cfg.DBName = dbName

		params := parsed.Query()
		if parseTimeValue := params.Get("parseTime"); parseTimeValue != "" {
			parseTime, err := strconv.ParseBool(parseTimeValue)
			if err != nil {
				return nil, fmt.Errorf("解析 MySQL DSN 参数 parseTime 失败: %w", err)
			}
			cfg.ParseTime = parseTime
			params.Del("parseTime")
		} else {
			cfg.ParseTime = true
		}

		if locValue := params.Get("loc"); locValue != "" {
			location, err := time.LoadLocation(locValue)
			if err != nil {
				return nil, fmt.Errorf("解析 MySQL DSN 参数 loc 失败: %w", err)
			}
			cfg.Loc = location
			params.Del("loc")
		} else {
			cfg.Loc = time.Local
		}

		if params.Get("charset") == "" {
			params.Set("charset", "utf8mb4")
		}

		cfg.Params = make(map[string]string, len(params))
		for key, values := range params {
			if len(values) == 0 {
				continue
			}
			cfg.Params[key] = values[len(values)-1]
		}

	} else {
		cfg, err = mysqlcfg.ParseDSN(stripped)
		if err != nil {
			return nil, fmt.Errorf("解析 MySQL DSN 失败: %w", err)
		}
		if cfg.DBName == "" {
			return nil, fmt.Errorf("MySQL DSN 缺少数据库名")
		}
		if !strings.Contains(stripped, "parseTime=") {
			cfg.ParseTime = true
		}
		if !strings.Contains(stripped, "loc=") {
			cfg.Loc = time.Local
		}
		if !strings.Contains(stripped, "charset=") {
			if cfg.Params == nil {
				cfg.Params = make(map[string]string)
			}
			cfg.Params["charset"] = "utf8mb4"
		}
	}

	driverDSN := cfg.FormatDSN()
	adminCfg := *cfg
	adminCfg.DBName = ""

	return &dbConnectionInfo{
		Dialect:            DialectMySQL,
		Dialector:          gmysql.Open(driverDSN),
		BootstrapDialector: gmysql.Open(adminCfg.FormatDSN()),
		DBName:             cfg.DBName,
		RedactedDSN:        "mysql://" + redactMySQLDriverDSN(driverDSN),
	}, nil
}

func buildPostgresConnectionInfo(rawDSN string) (*dbConnectionInfo, error) {
	parsed, err := url.Parse(rawDSN)
	if err != nil {
		return nil, fmt.Errorf("解析 PostgreSQL DSN 失败: %w", err)
	}

	dbName := strings.TrimPrefix(parsed.Path, "/")
	if dbName == "" {
		return nil, fmt.Errorf("PostgreSQL DSN 缺少数据库名")
	}

	bootstrapURL := *parsed
	bootstrapURL.Path = "/postgres"
	if bootstrapURL.RawPath != "" {
		bootstrapURL.RawPath = "/postgres"
	} else {
		bootstrapURL.RawPath = ""
	}

	return &dbConnectionInfo{
		Dialect:            DialectPostgres,
		Dialector:          gpostgres.Open(rawDSN),
		BootstrapDialector: gpostgres.Open(bootstrapURL.String()),
		DBName:             dbName,
		RedactedDSN:        redactURLDSN(rawDSN),
	}, nil
}

func isMissingDatabaseError(dialect string, err error) bool {
	if err == nil {
		return false
	}

	switch dialect {
	case DialectMySQL:
		var mysqlErr *mysqlcfg.MySQLError
		if errors.As(err, &mysqlErr) {
			return mysqlErr.Number == 1049
		}
		return strings.Contains(err.Error(), "Unknown database")
	case DialectPostgres:
		errText := strings.ToLower(err.Error())
		return strings.Contains(errText, "database") && strings.Contains(errText, "does not exist")
	default:
		return false
	}
}

func createDatabase(info *dbConnectionInfo, gormConfig *gorm.Config) error {
	if info.BootstrapDialector == nil || info.DBName == "" {
		return fmt.Errorf("缺少自动建库所需的连接信息")
	}

	adminDB, err := gorm.Open(info.BootstrapDialector, gormConfig)
	if err != nil {
		return err
	}

	sqlText := createDatabaseSQL(info.Dialect, info.DBName)
	if err := adminDB.Exec(sqlText).Error; err != nil {
		if info.Dialect == DialectPostgres && strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return nil
		}
		return err
	}

	utils.Info("数据库不存在，已自动创建: %s", info.DBName)
	return nil
}

func createDatabaseSQL(dialect, dbName string) string {
	switch dialect {
	case DialectMySQL:
		return fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", quoteIdentifier(dialect, dbName))
	case DialectPostgres:
		return fmt.Sprintf("CREATE DATABASE %s", quoteIdentifier(dialect, dbName))
	default:
		return ""
	}
}

func quoteIdentifier(dialect, value string) string {
	switch dialect {
	case DialectMySQL:
		return "`" + strings.ReplaceAll(value, "`", "``") + "`"
	case DialectPostgres:
		return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
	default:
		return value
	}
}

func mergeQueryParams(rawValue string, defaults map[string]string) string {
	base := rawValue
	rawQuery := ""
	if idx := strings.Index(rawValue, "?"); idx >= 0 {
		base = rawValue[:idx]
		rawQuery = rawValue[idx+1:]
	}

	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return rawValue
	}

	for key, value := range defaults {
		if values.Get(key) == "" {
			values.Set(key, value)
		}
	}

	encoded := values.Encode()
	if encoded == "" {
		return base
	}
	return base + "?" + encoded
}

func redactURLDSN(rawDSN string) string {
	parsed, err := url.Parse(rawDSN)
	if err != nil {
		return rawDSN
	}

	if parsed.User == nil {
		return rawDSN
	}

	username := parsed.User.Username()
	if _, hasPassword := parsed.User.Password(); hasPassword {
		parsed.User = url.UserPassword(username, "***")
	}
	return parsed.String()
}

func redactMySQLDriverDSN(rawDSN string) string {
	re := regexp.MustCompile(`^([^:]+):([^@]*)@`)
	return re.ReplaceAllString(rawDSN, `$1:***@`)
}

func newGormConfig() *gorm.Config {
	return &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	}
}

func configureConnectionPool(db *gorm.DB, dialect string) {
	sqlDB, err := db.DB()
	if err != nil {
		utils.Error("获取底层数据库连接失败: %v", err)
		return
	}

	switch dialect {
	case DialectSQLite:
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		utils.Info("SQLite 连接池配置完成: MaxIdle=10, MaxOpen=100, MaxLifetime=1h")
	default:
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		utils.Info("数据库连接池配置完成: MaxIdle=10, MaxOpen=100, MaxLifetime=1h")
	}
}

// initMemorySqlite 初始化内存数据库（演示模式专用）
func initMemorySqlite() error {
	// 使用 file::memory:?cache=shared 确保多个连接共享同一内存数据库
	dsn := "file::memory:?cache=shared&_foreign_keys=ON"

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dsn), newGormConfig())
	if err != nil {
		utils.Error("连接内存数据库失败: %v", err)
		return err
	}

	// 配置连接池 - 内存数据库需要保持连接活跃
	sqlDB, err := db.DB()
	if err != nil {
		utils.Error("获取底层数据库连接失败: %v", err)
	} else {
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetConnMaxLifetime(0) // 不过期
	}

	DB = db
	Dialect = DialectSQLite
	utils.Info("演示模式：使用内存数据库")
	return nil
}
