package models

import (
	"strconv"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"time"
)

// Airport 机场模型
// 用于管理外部订阅源，支持定时拉取更新
type Airport struct {
	ID                int        `gorm:"primaryKey;autoIncrement" json:"id"`
	Name              string     `json:"name"`                                   // 机场名称（唯一）
	URL               string     `json:"url"`                                    // 订阅地址
	CronExpr          string     `json:"cronExpr"`                               // 定时更新Cron表达式
	Enabled           bool       `json:"enabled"`                                // 是否启用
	SuccessCount      int        `gorm:"default:0" json:"successCount"`          // 成功拉取次数
	LastRunTime       *time.Time `gorm:"type:datetime" json:"lastRunTime"`       // 上次运行时间
	NextRunTime       *time.Time `gorm:"type:datetime" json:"nextRunTime"`       // 下次运行时间
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"createdAt"`        // 创建时间
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`        // 更新时间
	Group             string     `json:"group"`                                  // 导入节点的默认分组
	DownloadWithProxy bool       `gorm:"default:false" json:"downloadWithProxy"` // 是否使用代理下载
	ProxyLink         string     `gorm:"default:''" json:"proxyLink"`            // 代理节点链接
	UserAgent         string     `json:"userAgent"`                              // 自定义User-Agent
	NodeCount         int        `gorm:"-" json:"nodeCount"`                     // 节点数量（非数据库字段）
	// 用量信息相关字段
	FetchUsageInfo bool   `gorm:"default:false" json:"fetchUsageInfo"` // 是否获取用量信息
	UsageUpload    int64  `gorm:"default:0" json:"usageUpload"`        // 已上传流量（字节）
	UsageDownload  int64  `gorm:"default:0" json:"usageDownload"`      // 已下载流量（字节）
	UsageTotal     int64  `gorm:"default:0" json:"usageTotal"`         // 总流量配额（字节）
	UsageExpire    int64  `gorm:"default:0" json:"usageExpire"`        // 订阅过期时间（Unix时间戳）
	SkipTLSVerify  bool   `gorm:"default:false" json:"skipTLSVerify"`  // 是否跳过TLS证书验证
	Remark         string `json:"remark"`                              // 备注信息
	Logo           string `json:"logo"`                                // Logo：URL、icon:图标名、或emoji字符
}

// TableName 指定表名
func (Airport) TableName() string {
	return "airports"
}

// airportCache 使用泛型缓存
var airportCache *cache.MapCache[int, Airport]

func init() {
	airportCache = cache.NewMapCache(func(a Airport) int { return a.ID })
	airportCache.AddIndex("enabled", func(a Airport) string { return strconv.FormatBool(a.Enabled) })
	airportCache.AddIndex("name", func(a Airport) string { return a.Name })
}

// InitAirportCache 初始化机场缓存
func InitAirportCache() error {
	utils.Info("开始加载机场数据到缓存")
	var airports []Airport
	if err := database.DB.Find(&airports).Error; err != nil {
		return err
	}

	airportCache.LoadAll(airports)
	utils.Info("机场缓存初始化完成，共加载 %d 个机场", airportCache.Count())

	cache.Manager.Register("airport", airportCache)
	return nil
}

// Add 添加机场 (Write-Through)
func (a *Airport) Add() error {
	err := database.DB.Create(a).Error
	if err != nil {
		return err
	}
	airportCache.Set(a.ID, *a)
	return nil
}

// Update 更新机场 (Write-Through)
func (a *Airport) Update() error {
	err := database.DB.Model(a).Select(
		"Name", "URL", "CronExpr", "Enabled", "LastRunTime", "NextRunTime",
		"SuccessCount", "Group", "DownloadWithProxy", "ProxyLink", "UserAgent",
		"FetchUsageInfo", "SkipTLSVerify", "Remark", "Logo",
	).Updates(a).Error
	if err != nil {
		return err
	}
	// 从DB读取完整数据后更新缓存
	var updated Airport
	if err := database.DB.First(&updated, a.ID).Error; err == nil {
		airportCache.Set(a.ID, updated)
	}
	return nil
}

// Find 查找机场是否重复（按URL或名称）
func (a *Airport) Find() error {
	// 先查缓存
	results := airportCache.Filter(func(ap Airport) bool {
		return ap.URL == a.URL || ap.Name == a.Name
	})
	if len(results) > 0 {
		*a = results[0]
		return nil
	}
	return database.DB.Where("url = ? or name = ?", a.URL, a.Name).First(a).Error
}

// List 获取所有机场
func (a *Airport) List() ([]Airport, error) {
	airports := airportCache.GetAllSorted(func(x, y Airport) bool {
		return x.ID < y.ID
	})
	return airports, nil
}

// ListPaginated 分页获取机场列表
func (a *Airport) ListPaginated(page, pageSize int) ([]Airport, int64, error) {
	allAirports := airportCache.GetAllSorted(func(x, y Airport) bool {
		return x.ID < y.ID
	})
	total := int64(len(allAirports))

	if page <= 0 || pageSize <= 0 {
		return allAirports, total, nil
	}

	offset := (page - 1) * pageSize
	if offset >= len(allAirports) {
		return []Airport{}, total, nil
	}

	end := offset + pageSize
	if end > len(allAirports) {
		end = len(allAirports)
	}

	return allAirports[offset:end], total, nil
}

// AirportFilter 机场筛选条件
type AirportFilter struct {
	Keyword string // 关键字搜索（匹配名称或备注）
	Group   string // 分组筛选
	Enabled *bool  // 启用状态筛选
}

// ListWithFilter 带筛选条件的分页获取机场列表
func (a *Airport) ListWithFilter(page, pageSize int, filter AirportFilter) ([]Airport, int64, error) {
	// 从缓存获取所有数据
	allAirports := airportCache.GetAllSorted(func(x, y Airport) bool {
		return x.ID < y.ID
	})

	// 应用筛选条件
	var filteredAirports []Airport
	for _, ap := range allAirports {
		// 关键字模糊匹配（名称或备注）
		if filter.Keyword != "" && !containsIgnoreCase(ap.Name, filter.Keyword) && !containsIgnoreCase(ap.Remark, filter.Keyword) {
			continue
		}
		// 分组精确匹配
		if filter.Group != "" && ap.Group != filter.Group {
			continue
		}
		// 启用状态匹配
		if filter.Enabled != nil && ap.Enabled != *filter.Enabled {
			continue
		}
		filteredAirports = append(filteredAirports, ap)
	}

	total := int64(len(filteredAirports))

	if page <= 0 || pageSize <= 0 {
		return filteredAirports, total, nil
	}

	offset := (page - 1) * pageSize
	if offset >= len(filteredAirports) {
		return []Airport{}, total, nil
	}

	end := offset + pageSize
	if end > len(filteredAirports) {
		end = len(filteredAirports)
	}

	return filteredAirports[offset:end], total, nil
}

// ListEnabled 获取所有启用的机场
func ListEnabledAirports() ([]Airport, error) {
	return airportCache.GetByIndex("enabled", "true"), nil
}

// Del 删除机场 (Write-Through)
func (a *Airport) Del() error {
	err := database.DB.Delete(a).Error
	if err != nil {
		return err
	}
	airportCache.Delete(a.ID)
	return nil
}

// UpdateRunTime 更新运行时间 (Write-Through)
func (a *Airport) UpdateRunTime(lastRun, nextRun *time.Time) error {
	err := database.DB.Model(a).Select("LastRunTime", "NextRunTime").Updates(map[string]interface{}{
		"LastRunTime": lastRun,
		"NextRunTime": nextRun,
	}).Error
	if err != nil {
		return err
	}
	// 更新缓存
	if cached, ok := airportCache.Get(a.ID); ok {
		cached.LastRunTime = lastRun
		cached.NextRunTime = nextRun
		airportCache.Set(a.ID, cached)
	}
	return nil
}

// GetByID 根据ID获取机场
func (a *Airport) GetByID(id int) error {
	if cached, ok := airportCache.Get(id); ok {
		*a = cached
		return nil
	}
	return database.DB.Where("id = ?", id).First(a).Error
}

// GetAirportByID 根据ID获取机场（便捷函数）
func GetAirportByID(id int) (*Airport, error) {
	if cached, ok := airportCache.Get(id); ok {
		return &cached, nil
	}
	var airport Airport
	if err := database.DB.Where("id = ?", id).First(&airport).Error; err != nil {
		return nil, err
	}
	airportCache.Set(airport.ID, airport)
	return &airport, nil
}

// IncrementSuccessCount 增加成功次数
func (a *Airport) IncrementSuccessCount() error {
	err := database.DB.Model(a).Update("success_count", a.SuccessCount+1).Error
	if err != nil {
		return err
	}
	if cached, ok := airportCache.Get(a.ID); ok {
		cached.SuccessCount++
		airportCache.Set(a.ID, cached)
	}
	return nil
}

// DeleteAirportNodes 删除机场关联的所有节点
func DeleteAirportNodes(airportID int) error {
	return DeleteAutoSubscriptionNodes(airportID)
}

// ListNodesByAirportID 获取机场关联的所有节点
func ListNodesByAirportID(airportID int) ([]Node, error) {
	return ListBySourceID(airportID)
}

// UpdateNodesByAirportID 更新机场关联节点的来源名称和分组
func UpdateNodesByAirportID(airportID int, name string, group string) error {
	return UpdateNodesBySourceID(airportID, name, group)
}

// UpdateUsageInfo 更新用量信息 (Write-Through)
func (a *Airport) UpdateUsageInfo(upload, download, total, expire int64) error {
	err := database.DB.Model(a).Select("UsageUpload", "UsageDownload", "UsageTotal", "UsageExpire").Updates(map[string]interface{}{
		"UsageUpload":   upload,
		"UsageDownload": download,
		"UsageTotal":    total,
		"UsageExpire":   expire,
	}).Error
	if err != nil {
		return err
	}
	// 更新缓存
	if cached, ok := airportCache.Get(a.ID); ok {
		cached.UsageUpload = upload
		cached.UsageDownload = download
		cached.UsageTotal = total
		cached.UsageExpire = expire
		airportCache.Set(a.ID, cached)
	}
	return nil
}

// containsIgnoreCase 忽略大小写的字符串包含检查
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// AirportNodeStats 机场节点测速统计
type AirportNodeStats struct {
	DelayPassCount    int     `json:"delayPassCount"`    // 延迟测试通过数量
	SpeedPassCount    int     `json:"speedPassCount"`    // 速度测试通过数量
	LowestDelayNode   string  `json:"lowestDelayNode"`   // 延迟最低节点名称
	LowestDelayTime   int     `json:"lowestDelayTime"`   // 最低延迟时间(ms)
	LowestDelaySpeed  float64 `json:"lowestDelaySpeed"`  // 最低延迟节点速度
	HighestSpeedNode  string  `json:"highestSpeedNode"`  // 速度最高节点名称
	HighestSpeed      float64 `json:"highestSpeed"`      // 最高速度(MB/s)
	HighestSpeedDelay int     `json:"highestSpeedDelay"` // 最高速度节点延迟
}

// GetAirportNodeStats 获取机场节点测速统计
func GetAirportNodeStats(airportID int) AirportNodeStats {
	nodes, err := ListBySourceID(airportID)
	if err != nil || len(nodes) == 0 {
		return AirportNodeStats{}
	}

	stats := AirportNodeStats{}
	var lowestDelayNode *Node
	var highestSpeedNode *Node

	for i := range nodes {
		node := &nodes[i]

		// 延迟测试通过：DelayStatus 为 success 且 DelayTime > 0
		if node.DelayStatus == "success" && node.DelayTime > 0 {
			stats.DelayPassCount++

			// 寻找延迟最低且速度有效的节点
			if node.Speed > 0 {
				if lowestDelayNode == nil || node.DelayTime < lowestDelayNode.DelayTime {
					lowestDelayNode = node
				}
			}
		}

		// 速度测试通过：SpeedStatus 为 success 且 Speed > 0
		if node.SpeedStatus == "success" && node.Speed > 0 {
			stats.SpeedPassCount++

			// 寻找速度最高的节点
			if highestSpeedNode == nil || node.Speed > highestSpeedNode.Speed {
				highestSpeedNode = node
			}
		}
	}

	// 填充最低延迟节点信息
	if lowestDelayNode != nil {
		stats.LowestDelayNode = lowestDelayNode.Name
		stats.LowestDelayTime = lowestDelayNode.DelayTime
		stats.LowestDelaySpeed = lowestDelayNode.Speed
	}

	// 填充最高速度节点信息
	if highestSpeedNode != nil {
		stats.HighestSpeedNode = highestSpeedNode.Name
		stats.HighestSpeed = highestSpeedNode.Speed
		stats.HighestSpeedDelay = highestSpeedNode.DelayTime
	}

	return stats
}
