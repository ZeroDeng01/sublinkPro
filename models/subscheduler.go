package models

import (
	"strconv"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"time"
)

type SubScheduler struct {
	ID                int `gorm:"primaryKey;autoIncrement"`
	Name              string
	URL               string
	CronExpr          string
	Enabled           bool
	SuccessCount      int        `gorm:"default:0"`
	LastRunTime       *time.Time `gorm:"type:datetime"`
	NextRunTime       *time.Time `gorm:"type:datetime"`
	CreatedAt         time.Time  `gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime"`
	Group             string
	DownloadWithProxy bool   `gorm:"default:false"`
	ProxyLink         string `gorm:"default:''"`
	UserAgent         string
	NodeCount         int `gorm:"-"`
}

// subSchedulerCache 使用新的泛型缓存
var subSchedulerCache *cache.MapCache[int, SubScheduler]

func init() {
	subSchedulerCache = cache.NewMapCache(func(ss SubScheduler) int { return ss.ID })
	subSchedulerCache.AddIndex("enabled", func(ss SubScheduler) string { return strconv.FormatBool(ss.Enabled) })
}

// InitSubSchedulerCache 初始化订阅调度缓存
func InitSubSchedulerCache() error {
	utils.Info("开始加载订阅调度到缓存")
	var schedulers []SubScheduler
	if err := database.DB.Find(&schedulers).Error; err != nil {
		return err
	}

	subSchedulerCache.LoadAll(schedulers)
	utils.Info("订阅调度缓存初始化完成，共加载 %d 个调度任务", subSchedulerCache.Count())

	cache.Manager.Register("subscheduler", subSchedulerCache)
	return nil
}

// Add 添加订阅调度 (Write-Through)
func (ss *SubScheduler) Add() error {
	err := database.DB.Create(ss).Error
	if err != nil {
		return err
	}
	subSchedulerCache.Set(ss.ID, *ss)
	return nil
}

// Update 更新订阅调度 (Write-Through)
func (ss *SubScheduler) Update() error {
	err := database.DB.Model(ss).Select("Name", "URL", "CronExpr", "Enabled", "LastRunTime", "NextRunTime", "SuccessCount", "Group", "DownloadWithProxy", "ProxyLink", "UserAgent").Updates(ss).Error
	if err != nil {
		return err
	}
	// 从DB读取完整数据后更新缓存
	var updated SubScheduler
	if err := database.DB.First(&updated, ss.ID).Error; err == nil {
		subSchedulerCache.Set(ss.ID, updated)
	}
	return nil
}

// Find 查找订阅调度是否重复
func (ss *SubScheduler) Find() error {
	// 先查缓存
	results := subSchedulerCache.Filter(func(s SubScheduler) bool {
		return s.URL == ss.URL || s.Name == ss.Name
	})
	if len(results) > 0 {
		*ss = results[0]
		return nil
	}
	return database.DB.Where("url = ? or name = ?", ss.URL, ss.Name).First(ss).Error
}

// List 获取所有订阅调度
func (ss *SubScheduler) List() ([]SubScheduler, error) {
	schedulers := subSchedulerCache.GetAllSorted(func(a, b SubScheduler) bool {
		return a.ID < b.ID
	})
	return schedulers, nil
}

// ListPaginated 分页获取订阅调度列表
func (ss *SubScheduler) ListPaginated(page, pageSize int) ([]SubScheduler, int64, error) {
	allSchedulers := subSchedulerCache.GetAllSorted(func(a, b SubScheduler) bool {
		return a.ID < b.ID
	})
	total := int64(len(allSchedulers))

	if page <= 0 || pageSize <= 0 {
		return allSchedulers, total, nil
	}

	offset := (page - 1) * pageSize
	if offset >= len(allSchedulers) {
		return []SubScheduler{}, total, nil
	}

	end := offset + pageSize
	if end > len(allSchedulers) {
		end = len(allSchedulers)
	}

	return allSchedulers[offset:end], total, nil
}

// ListEnabled 获取所有启用的订阅调度
func ListEnabled() ([]SubScheduler, error) {
	// 使用二级索引查询
	return subSchedulerCache.GetByIndex("enabled", "true"), nil
}

// Del 删除订阅调度 (Write-Through)
func (ss *SubScheduler) Del() error {
	err := database.DB.Delete(ss).Error
	if err != nil {
		return err
	}
	subSchedulerCache.Delete(ss.ID)
	return nil
}

// UpdateRunTime 更新运行时间 (Write-Through)
func (ss *SubScheduler) UpdateRunTime(lastRun, nextRun *time.Time) error {
	err := database.DB.Model(ss).Select("LastRunTime", "NextRunTime").Updates(map[string]interface{}{
		"LastRunTime": lastRun,
		"NextRunTime": nextRun,
	}).Error
	if err != nil {
		return err
	}
	// 更新缓存
	if cached, ok := subSchedulerCache.Get(ss.ID); ok {
		cached.LastRunTime = lastRun
		cached.NextRunTime = nextRun
		subSchedulerCache.Set(ss.ID, cached)
	}
	return nil
}

// GetByID 根据ID获取订阅调度
func (ss *SubScheduler) GetByID(id int) error {
	if cached, ok := subSchedulerCache.Get(id); ok {
		*ss = cached
		return nil
	}
	return database.DB.Where("id = ?", id).First(ss).Error
}
