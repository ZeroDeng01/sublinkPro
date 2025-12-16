package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"sync"
	"time"
)

// IPInfo IP信息数据模型
type IPInfo struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	IP          string    `gorm:"uniqueIndex;size:64" json:"ip"` // IP地址（支持IPv6）
	Country     string    `json:"country"`                       // 国家名称
	CountryCode string    `json:"countryCode"`                   // 国家代码 (如 CN, US)
	Region      string    `json:"region"`                        // 地区/省份代码
	RegionName  string    `json:"regionName"`                    // 地区/省份名称
	City        string    `json:"city"`                          // 城市
	Zip         string    `json:"zip"`                           // 邮编
	Lat         float64   `json:"lat"`                           // 纬度
	Lon         float64   `json:"lon"`                           // 经度
	Timezone    string    `json:"timezone"`                      // 时区
	ISP         string    `json:"isp"`                           // ISP提供商
	Org         string    `json:"org"`                           // 组织
	AS          string    `json:"as"`                            // AS号
	RawResponse string    `gorm:"type:text" json:"-"`            // 原始JSON响应
	Provider    string    `json:"provider"`                      // 数据提供商
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// 缓存有效期：7天
const ipInfoCacheTTL = 7 * 24 * time.Hour

// ipInfoCache 内存缓存
var ipInfoCache *cache.MapCache[string, IPInfo]

// ipInfoCacheLock 用于API请求去重
var ipInfoRequestLock sync.Map

func init() {
	// 初始化IP信息缓存，主键为 IP
	ipInfoCache = cache.NewMapCache(func(info IPInfo) string { return info.IP })
}

// InitIPInfoCache 初始化IP信息缓存
func InitIPInfoCache() error {
	utils.Info("开始加载IP信息到缓存")

	// 只加载7天内的有效数据到缓存
	var ipInfoList []IPInfo
	cutoffTime := time.Now().Add(-ipInfoCacheTTL)
	if err := database.DB.Where("updated_at > ?", cutoffTime).Find(&ipInfoList).Error; err != nil {
		return err
	}

	ipInfoCache.LoadAll(ipInfoList)
	utils.Info("IP信息缓存初始化完成，共加载 %d 条记录", ipInfoCache.Count())

	// 注册到缓存管理器
	cache.Manager.Register("ipinfo", ipInfoCache)
	return nil
}

// GetIPInfoCount 获取IP信息缓存数量（数据库中的总数）
func GetIPInfoCount() int64 {
	var count int64
	database.DB.Model(&IPInfo{}).Count(&count)
	return count
}

// ClearAllIPInfo 清除所有IP信息缓存
func ClearAllIPInfo() error {
	// 清除数据库
	if err := database.DB.Where("1=1").Delete(&IPInfo{}).Error; err != nil {
		return err
	}

	// 清除内存缓存
	ipInfoCache.Clear()

	utils.Info("已清除所有IP信息缓存")
	return nil
}

// GetIPInfo 获取IP信息（多级缓存）
func GetIPInfo(ip string) (*IPInfo, error) {
	if ip == "" {
		return nil, fmt.Errorf("IP地址不能为空")
	}

	// 1. 检查内存缓存
	if info, ok := ipInfoCache.Get(ip); ok {
		// 检查缓存是否过期
		if time.Since(info.UpdatedAt) < ipInfoCacheTTL {
			return &info, nil
		}
		// 缓存过期，删除旧数据
		ipInfoCache.Delete(ip)
	}

	// 2. 检查数据库
	var dbInfo IPInfo
	if err := database.DB.Where("ip = ?", ip).First(&dbInfo).Error; err == nil {
		// 检查数据是否过期
		if time.Since(dbInfo.UpdatedAt) < ipInfoCacheTTL {
			// 加载到内存缓存
			ipInfoCache.Set(ip, dbInfo)
			return &dbInfo, nil
		}
		// 数据过期，需要刷新
	}

	// 3. 请求去重：防止多个并发请求同时查询同一个IP
	lockChan := make(chan struct{})
	if actual, loaded := ipInfoRequestLock.LoadOrStore(ip, lockChan); loaded {
		// 已有其他请求在处理，等待完成
		<-actual.(chan struct{})
		// 再次检查缓存
		if info, ok := ipInfoCache.Get(ip); ok {
			return &info, nil
		}
		return nil, fmt.Errorf("获取IP信息失败")
	}
	defer func() {
		ipInfoRequestLock.Delete(ip)
		close(lockChan)
	}()

	// 4. 从第三方API获取
	info, err := fetchIPInfoFromAPI(ip)
	if err != nil {
		return nil, err
	}

	// 5. 保存到数据库（更新或插入）
	if dbInfo.ID > 0 {
		// 更新现有记录
		info.ID = dbInfo.ID
		if err := database.DB.Save(info).Error; err != nil {
			utils.Error("更新IP信息失败: %v", err)
		}
	} else {
		// 插入新记录
		if err := database.DB.Create(info).Error; err != nil {
			utils.Error("保存IP信息失败: %v", err)
		}
	}

	// 6. 更新内存缓存
	ipInfoCache.Set(ip, *info)

	return info, nil
}

// ipAPIResponse ip-api.com API响应结构
type ipAPIResponse struct {
	Status      string  `json:"status"`
	Message     string  `json:"message"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
}

// fetchIPInfoFromAPI 从第三方API获取IP信息
func fetchIPInfoFromAPI(ip string) (*IPInfo, error) {
	// 使用 ip-api.com（支持中文，免费）
	url := fmt.Sprintf("http://ip-api.com/json/%s?lang=zh-CN&fields=status,message,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,query", ip)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求IP信息API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取API响应失败: %w", err)
	}

	var apiResp ipAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析API响应失败: %w", err)
	}

	if apiResp.Status != "success" {
		return nil, fmt.Errorf("API返回错误: %s", apiResp.Message)
	}

	// 转换为IPInfo结构
	info := &IPInfo{
		IP:          ip,
		Country:     apiResp.Country,
		CountryCode: apiResp.CountryCode,
		Region:      apiResp.Region,
		RegionName:  apiResp.RegionName,
		City:        apiResp.City,
		Zip:         apiResp.Zip,
		Lat:         apiResp.Lat,
		Lon:         apiResp.Lon,
		Timezone:    apiResp.Timezone,
		ISP:         apiResp.ISP,
		Org:         apiResp.Org,
		AS:          apiResp.AS,
		RawResponse: string(body),
		Provider:    "ip-api.com",
	}

	return info, nil
}
