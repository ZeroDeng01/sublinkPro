package geoip

import (
	"fmt"
	"net/netip"
	"os"
	"sublink/config"
	"sublink/utils"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang/v2"
)

var (
	geoIP     *geoip2.Reader
	mu        sync.RWMutex
	dbPath    string    // 当前加载的数据库路径
	available bool      // 数据库是否可用
	dbInfo    *DBInfo   // 数据库信息
	initOnce  sync.Once // 确保只初始化一次
)

// DBInfo 数据库信息
type DBInfo struct {
	Path      string    `json:"path"`      // 文件路径
	Size      int64     `json:"size"`      // 文件大小（字节）
	ModTime   time.Time `json:"modTime"`   // 最后修改时间
	Available bool      `json:"available"` // 是否可用
}

// InitGeoIP 初始化 GeoIP 数据库
// 如果文件不存在，不会阻止系统启动，只是标记为不可用
func InitGeoIP() error {
	var initErr error
	initOnce.Do(func() {
		initErr = loadDatabase()
	})
	return initErr
}

// loadDatabase 加载数据库文件
func loadDatabase() error {
	mu.Lock()
	defer mu.Unlock()

	// 获取配置的路径
	path := config.GetGeoIPPath()
	dbPath = path

	// 检查文件是否存在
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		available = false
		dbInfo = &DBInfo{
			Path:      path,
			Available: false,
		}
		utils.Warn("GeoIP database file not found: %s; related features will be unavailable", path)
		return nil // 不返回错误，允许系统继续启动
	}
	if err != nil {
		available = false
		dbInfo = &DBInfo{
			Path:      path,
			Available: false,
		}
		utils.Error("Failed to check GeoIP database file: %v", err)
		return nil
	}

	// 打开数据库
	reader, err := geoip2.Open(path)
	if err != nil {
		available = false
		dbInfo = &DBInfo{
			Path:      path,
			Size:      fileInfo.Size(),
			ModTime:   fileInfo.ModTime(),
			Available: false,
		}
		utils.Error("Failed to open GeoIP database: %v", err)
		return err
	}

	// 关闭旧的 reader
	if geoIP != nil {
		if err := geoIP.Close(); err != nil {
			utils.Warn("Failed to close old GeoIP database: %v", err)
		}
	}

	geoIP = reader
	available = true
	dbInfo = &DBInfo{
		Path:      path,
		Size:      fileInfo.Size(),
		ModTime:   fileInfo.ModTime(),
		Available: true,
	}

	utils.Info("GeoIP database loaded: %s (%.2f MB)", path, float64(fileInfo.Size())/1024/1024)
	return nil
}

// Reload 重新加载 GeoIP 数据库
func Reload() error {
	// 重置 initOnce 以允许重新初始化
	initOnce = sync.Once{}
	return loadDatabase()
}

// IsAvailable 检查 GeoIP 数据库是否可用
func IsAvailable() bool {
	mu.RLock()
	defer mu.RUnlock()
	return available
}

// GetDBInfo 获取数据库信息
func GetDBInfo() *DBInfo {
	mu.RLock()
	defer mu.RUnlock()

	if dbInfo != nil {
		return dbInfo
	}

	// 返回默认信息
	path := config.GetGeoIPPath()
	info := &DBInfo{
		Path:      path,
		Available: false,
	}

	// 尝试获取文件信息
	if fileInfo, err := os.Stat(path); err == nil {
		info.Size = fileInfo.Size()
		info.ModTime = fileInfo.ModTime()
	}

	return info
}

// GetDBPath 获取当前数据库路径
func GetDBPath() string {
	mu.RLock()
	defer mu.RUnlock()
	if dbPath != "" {
		return dbPath
	}
	return config.GetGeoIPPath()
}

// GetLocation 返回给定 IP 地址的位置信息
func GetLocation(ipStr string) (string, error) {
	mu.RLock()
	defer mu.RUnlock()

	if !available || geoIP == nil {
		return "", fmt.Errorf("geoIP database is unavailable")
	}

	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return "Unknown", nil
	}

	country := ""
	city := ""

	geoCountry, err := geoIP.Country(ip)
	if err != nil {
		utils.Error("Failed to get Country: %v", err)
	}
	if geoCountry.Country.HasData() {
		country = geoCountry.Country.Names.SimplifiedChinese
		isocode := geoCountry.Country.ISOCode
		flag := ISOCodeToFlag(isocode)
		if flag != "" {
			country = fmt.Sprintf("%s%s", flag, country)
		} else {
			country = fmt.Sprintf("(%s)%s", isocode, country)
		}
	}

	getCity, err := geoIP.City(ip)
	if err != nil {
		utils.Error("Failed to get City: %v", err)
	}
	if getCity.City.HasData() {
		city = getCity.City.Names.SimplifiedChinese
	}

	return fmt.Sprintf("%s%s", country, city), nil
}

// ISOCodeToFlag 将 ISO 3166-1 alpha-2 国家代码转换为国旗 emoji
// 示例: "CN" -> 🇨🇳, "US" -> 🇺🇸
func ISOCodeToFlag(isoCode string) string {
	if len(isoCode) != 2 {
		return ""
	}

	// 将每个字母转换为对应的区域指示符号
	// 区域指示符号范围从 U+1F1E6 (A) 到 U+1F1FF (Z)
	flag := ""
	for _, char := range isoCode {
		if char >= 'A' && char <= 'Z' {
			flag += string(0x1F1E6 + (char - 'A'))
		} else if char >= 'a' && char <= 'z' {
			flag += string(0x1F1E6 + (char - 'a'))
		}
	}
	return flag
}

// GetCountryISOCode 返回给定 IP 地址的 ISO 国家代码 (例如 "US", "CN", "JP")
func GetCountryISOCode(ipStr string) (string, error) {
	mu.RLock()
	defer mu.RUnlock()

	if !available || geoIP == nil {
		return "", fmt.Errorf("geoIP database is unavailable")
	}

	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return "", fmt.Errorf("invalid IP address: %s", ipStr)
	}

	geoCountry, err := geoIP.Country(ip)
	if err != nil {
		return "", fmt.Errorf("failed to get country info: %v", err)
	}
	if geoCountry.Country.HasData() {
		return geoCountry.Country.ISOCode, nil
	}
	return "", nil
}

// Close 关闭 GeoIP reader
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if geoIP != nil {
		err := geoIP.Close()
		geoIP = nil
		available = false
		return err
	}
	return nil
}
