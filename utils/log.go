package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// 日志等级常量
const (
	LevelDebug = iota // 调试信息
	LevelInfo         // 一般信息
	LevelWarn         // 警告信息
	LevelError        // 错误信息
	LevelFatal        // 致命错误
)

// 等级名称映射
var levelNames = map[int]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

// 等级解析映射
var levelFromString = map[string]int{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
	"fatal": LevelFatal,
}

// Logger 统一日志工具
type Logger struct {
	level     int          // 当前日志等级
	logger    *log.Logger  // 底层日志实例
	file      *os.File     // 日志文件
	mutex     sync.RWMutex // 读写锁
	logPath   string       // 日志目录路径
	colorMode bool         // 是否启用颜色输出
}

// 全局日志实例
var (
	globalLogger *Logger
	loggerOnce   sync.Once
)

// GetLogger 获取全局日志实例
func GetLogger() *Logger {
	loggerOnce.Do(func() {
		globalLogger = &Logger{
			level:     LevelInfo, // 默认 INFO 等级
			colorMode: true,      // 默认启用颜色
		}
	})
	return globalLogger
}

// InitLogger 初始化日志系统
// logPath: 日志文件目录
// level: 日志等级字符串 (debug/info/warn/error/fatal)
func InitLogger(logPath string, level string) {
	logger := GetLogger()
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	// 设置日志等级
	if lvl, ok := levelFromString[strings.ToLower(level)]; ok {
		logger.level = lvl
	}

	// 确保日志目录存在
	if logPath == "" {
		logPath = "./logs"
	}
	logger.logPath = logPath

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		if err := os.MkdirAll(logPath, 0755); err != nil {
			log.Printf("创建日志目录失败: %v", err)
			return
		}
	}

	// 创建日志文件
	t := time.Now().Format("2006-01-02") + ".log"
	logFilePath := filepath.Join(logPath, t)
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("创建日志文件失败: %v", err)
		return
	}

	// 关闭旧文件
	if logger.file != nil {
		logger.file.Close()
	}
	logger.file = file

	// 设置多路输出（控制台 + 文件）
	mw := io.MultiWriter(os.Stdout, file)
	logger.logger = log.New(mw, "", 0)
}

// SetLevel 设置日志等级
func (l *Logger) SetLevel(level int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.level = level
}

// SetLevelString 通过字符串设置日志等级
func (l *Logger) SetLevelString(level string) {
	if lvl, ok := levelFromString[strings.ToLower(level)]; ok {
		l.SetLevel(lvl)
	}
}

// GetLevel 获取当前日志等级
func (l *Logger) GetLevel() int {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.level
}

// GetLevelString 获取当前日志等级字符串
func (l *Logger) GetLevelString() string {
	return levelNames[l.GetLevel()]
}

// shouldLog 判断是否应该记录该等级的日志
func (l *Logger) shouldLog(level int) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return level >= l.level
}

// formatMessage 格式化日志消息
func (l *Logger) formatMessage(level int, format string, v ...interface{}) string {
	// 获取调用者信息
	_, file, line, ok := runtime.Caller(3)
	caller := "???"
	if ok {
		caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	// 时间戳
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 等级名称
	levelName := levelNames[level]

	// 格式化消息
	var msg string
	if len(v) > 0 {
		msg = fmt.Sprintf(format, v...)
	} else {
		msg = format
	}

	// 组装完整日志行
	return fmt.Sprintf("%s [%s] [%s] %s", timestamp, levelName, caller, msg)
}

// output 输出日志
func (l *Logger) output(level int, format string, v ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	msg := l.formatMessage(level, format, v...)

	l.mutex.RLock()
	logger := l.logger
	l.mutex.RUnlock()

	if logger != nil {
		logger.Println(msg)
	} else {
		// 回退到标准输出
		log.Println(msg)
	}
}

// Debug 输出调试日志
func (l *Logger) Debug(format string, v ...interface{}) {
	l.output(LevelDebug, format, v...)
}

// Info 输出信息日志
func (l *Logger) Info(format string, v ...interface{}) {
	l.output(LevelInfo, format, v...)
}

// Warn 输出警告日志
func (l *Logger) Warn(format string, v ...interface{}) {
	l.output(LevelWarn, format, v...)
}

// Error 输出错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	l.output(LevelError, format, v...)
}

// Fatal 输出致命错误日志并退出程序
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.output(LevelFatal, format, v...)
	os.Exit(1)
}

// ============= 全局便捷函数 =============

// Debug 输出调试日志
func Debug(format string, v ...interface{}) {
	GetLogger().output(LevelDebug, format, v...)
}

// Info 输出信息日志
func Info(format string, v ...interface{}) {
	GetLogger().output(LevelInfo, format, v...)
}

// Warn 输出警告日志
func Warn(format string, v ...interface{}) {
	GetLogger().output(LevelWarn, format, v...)
}

// Error 输出错误日志
func Error(format string, v ...interface{}) {
	GetLogger().output(LevelError, format, v...)
}

// Fatal 输出致命错误日志并退出程序
func Fatal(format string, v ...interface{}) {
	GetLogger().output(LevelFatal, format, v...)
	os.Exit(1) // output normally doesn't exit for Fatal, but Logger.Fatal does. Wait Logger.Fatal calls output then Exit.
}

// SetLogLevel 设置全局日志等级
func SetLogLevel(level string) {
	GetLogger().SetLevelString(level)
}

// GetLogLevel 获取全局日志等级
func GetLogLevel() string {
	return GetLogger().GetLevelString()
}

// Loginit 初始化日志（向后兼容旧接口）
// Deprecated: 请使用 InitLogger 代替
func Loginit() {
	InitLogger("./logs", "info")
}

// ParseLogLevel 解析日志等级字符串为等级值
func ParseLogLevel(level string) int {
	if lvl, ok := levelFromString[strings.ToLower(level)]; ok {
		return lvl
	}
	return LevelInfo
}
