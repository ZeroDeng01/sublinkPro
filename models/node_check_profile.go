package models

import (
	"strconv"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"time"
)

// NodeCheckProfile 节点检测策略模型
// 用于管理节点检测配置，支持多策略和定时执行
type NodeCheckProfile struct {
	ID       int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"not null;uniqueIndex" json:"name"` // 策略名称（唯一）
	Enabled  bool   `gorm:"default:false" json:"enabled"`     // 是否启用定时检测
	CronExpr string `json:"cronExpr"`                         // Cron 表达式

	// 检测模式参数
	Mode       string `gorm:"default:'tcp'" json:"mode"` // 检测模式：tcp / mihomo
	TestURL    string `json:"testUrl"`                   // 检测URL（下载测速或延迟检测）
	LatencyURL string `json:"latencyUrl"`                // 延迟检测URL（仅mihomo模式）
	Timeout    int    `gorm:"default:5" json:"timeout"`  // 超时时间(秒)

	// 范围过滤（逗号分隔）
	Groups string `json:"groups"` // 检测分组
	Tags   string `json:"tags"`   // 检测标签

	// 并发控制
	LatencyConcurrency int `gorm:"default:0" json:"latencyConcurrency"` // 延迟检测并发(0=自动)
	SpeedConcurrency   int `gorm:"default:0" json:"speedConcurrency"`   // 速度检测并发(0=自动)

	// 高级选项
	DetectCountry      bool   `gorm:"default:false" json:"detectCountry"`       // 检测落地IP国家
	LandingIPURL       string `json:"landingIpUrl"`                             // IP查询接口URL
	IncludeHandshake   bool   `gorm:"default:true" json:"includeHandshake"`     // 延迟包含握手时间
	SpeedRecordMode    string `gorm:"default:'average'" json:"speedRecordMode"` // 速度记录模式：average/peak
	PeakSampleInterval int    `gorm:"default:100" json:"peakSampleInterval"`    // 峰值采样间隔(ms)

	// 流量统计开关
	TrafficByGroup  bool `gorm:"default:true" json:"trafficByGroup"`
	TrafficBySource bool `gorm:"default:true" json:"trafficBySource"`
	TrafficByNode   bool `gorm:"default:false" json:"trafficByNode"`

	// 执行时间记录
	LastRunTime *time.Time `gorm:"type:datetime" json:"lastRunTime"` // 上次执行时间
	NextRunTime *time.Time `gorm:"type:datetime" json:"nextRunTime"` // 下次执行时间

	// 时间戳
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (NodeCheckProfile) TableName() string {
	return "node_check_profiles"
}

// nodeCheckProfileCache 使用泛型缓存
var nodeCheckProfileCache *cache.MapCache[int, NodeCheckProfile]

func init() {
	nodeCheckProfileCache = cache.NewMapCache(func(p NodeCheckProfile) int { return p.ID })
	nodeCheckProfileCache.AddIndex("enabled", func(p NodeCheckProfile) string { return strconv.FormatBool(p.Enabled) })
	nodeCheckProfileCache.AddIndex("name", func(p NodeCheckProfile) string { return p.Name })
}

// InitNodeCheckProfileCache 初始化节点检测策略缓存
func InitNodeCheckProfileCache() error {
	utils.Info("开始加载节点检测策略到缓存")
	var profiles []NodeCheckProfile
	if err := database.DB.Find(&profiles).Error; err != nil {
		return err
	}

	nodeCheckProfileCache.LoadAll(profiles)
	utils.Info("节点检测策略缓存初始化完成，共加载 %d 个策略", nodeCheckProfileCache.Count())

	cache.Manager.Register("node_check_profile", nodeCheckProfileCache)
	return nil
}

// Add 添加策略 (Write-Through)
func (p *NodeCheckProfile) Add() error {
	err := database.DB.Create(p).Error
	if err != nil {
		return err
	}
	nodeCheckProfileCache.Set(p.ID, *p)
	return nil
}

// Update 更新策略 (Write-Through)
func (p *NodeCheckProfile) Update() error {
	err := database.DB.Model(p).Select(
		"Name", "Enabled", "CronExpr",
		"Mode", "TestURL", "LatencyURL", "Timeout",
		"Groups", "Tags",
		"LatencyConcurrency", "SpeedConcurrency",
		"DetectCountry", "LandingIPURL", "IncludeHandshake",
		"SpeedRecordMode", "PeakSampleInterval",
		"TrafficByGroup", "TrafficBySource", "TrafficByNode",
	).Updates(p).Error
	if err != nil {
		return err
	}
	// 从DB读取完整数据后更新缓存
	var updated NodeCheckProfile
	if err := database.DB.First(&updated, p.ID).Error; err == nil {
		nodeCheckProfileCache.Set(p.ID, updated)
	}
	return nil
}

// Del 删除策略 (Write-Through)
func (p *NodeCheckProfile) Del() error {
	err := database.DB.Delete(p).Error
	if err != nil {
		return err
	}
	nodeCheckProfileCache.Delete(p.ID)
	return nil
}

// GetByID 根据ID获取策略
func (p *NodeCheckProfile) GetByID(id int) error {
	if cached, ok := nodeCheckProfileCache.Get(id); ok {
		*p = cached
		return nil
	}
	return database.DB.Where("id = ?", id).First(p).Error
}

// GetNodeCheckProfileByID 根据ID获取策略（便捷函数）
func GetNodeCheckProfileByID(id int) (*NodeCheckProfile, error) {
	if cached, ok := nodeCheckProfileCache.Get(id); ok {
		return &cached, nil
	}
	var profile NodeCheckProfile
	if err := database.DB.Where("id = ?", id).First(&profile).Error; err != nil {
		return nil, err
	}
	nodeCheckProfileCache.Set(profile.ID, profile)
	return &profile, nil
}

// List 获取所有策略
func (p *NodeCheckProfile) List() ([]NodeCheckProfile, error) {
	profiles := nodeCheckProfileCache.GetAllSorted(func(x, y NodeCheckProfile) bool {
		return x.ID < y.ID
	})
	return profiles, nil
}

// ListEnabledNodeCheckProfiles 获取所有启用定时的策略
func ListEnabledNodeCheckProfiles() ([]NodeCheckProfile, error) {
	return nodeCheckProfileCache.GetByIndex("enabled", "true"), nil
}

// UpdateRunTime 更新运行时间 (Write-Through)
func (p *NodeCheckProfile) UpdateRunTime(lastRun, nextRun *time.Time) error {
	err := database.DB.Model(p).Select("LastRunTime", "NextRunTime").Updates(map[string]interface{}{
		"LastRunTime": lastRun,
		"NextRunTime": nextRun,
	}).Error
	if err != nil {
		return err
	}
	// 更新缓存
	if cached, ok := nodeCheckProfileCache.Get(p.ID); ok {
		cached.LastRunTime = lastRun
		cached.NextRunTime = nextRun
		nodeCheckProfileCache.Set(p.ID, cached)
	}
	return nil
}

// UpdateLastRunTime 只更新上次执行时间（不改变下次执行时间）
func (p *NodeCheckProfile) UpdateLastRunTime(lastRun *time.Time) error {
	err := database.DB.Model(p).Update("LastRunTime", lastRun).Error
	if err != nil {
		return err
	}
	// 更新缓存
	if cached, ok := nodeCheckProfileCache.Get(p.ID); ok {
		cached.LastRunTime = lastRun
		nodeCheckProfileCache.Set(p.ID, cached)
	}
	return nil
}

// FindByName 根据名称查找策略
func FindNodeCheckProfileByName(name string) (*NodeCheckProfile, error) {
	results := nodeCheckProfileCache.GetByIndex("name", name)
	if len(results) > 0 {
		return &results[0], nil
	}
	var profile NodeCheckProfile
	if err := database.DB.Where("name = ?", name).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

// GetGroups 获取分组列表（从逗号分隔字符串解析）
func (p *NodeCheckProfile) GetGroups() []string {
	if p.Groups == "" {
		return []string{}
	}
	return strings.Split(p.Groups, ",")
}

// GetTags 获取标签列表（从逗号分隔字符串解析）
func (p *NodeCheckProfile) GetTags() []string {
	if p.Tags == "" {
		return []string{}
	}
	return strings.Split(p.Tags, ",")
}

// SetGroups 设置分组列表（转换为逗号分隔字符串）
func (p *NodeCheckProfile) SetGroups(groups []string) {
	p.Groups = strings.Join(groups, ",")
}

// SetTags 设置标签列表（转换为逗号分隔字符串）
func (p *NodeCheckProfile) SetTags(tags []string) {
	p.Tags = strings.Join(tags, ",")
}
