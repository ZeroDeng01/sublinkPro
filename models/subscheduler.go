package models

import (
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
}

// Add 添加订阅调度
func (ss *SubScheduler) Add() error {
	return DB.Create(ss).Error
}

// Update 更新订阅调度
func (ss *SubScheduler) Update() error {
	return DB.Model(ss).Select("Name", "URL", "CronExpr", "Enabled", "LastRunTime", "NextRunTime", "SuccessCount", "Group", "DownloadWithProxy", "ProxyLink").Updates(ss).Error
}

// 查找节点是否重复
func (ss *SubScheduler) Find() error {
	return DB.Where("url = ? or name = ?", ss.URL, ss.Name).First(ss).Error
}

// List 获取所有订阅调度
func (ss *SubScheduler) List() ([]SubScheduler, error) {
	var schedulers []SubScheduler
	err := DB.Find(&schedulers).Error
	if err != nil {
		return nil, err
	}
	return schedulers, nil
}

// ListEnabled 获取所有启用的订阅调度
func ListEnabled() ([]SubScheduler, error) {
	var schedulers []SubScheduler
	err := DB.Where("enabled = 1").Find(&schedulers).Error
	if err != nil {
		return nil, err
	}
	return schedulers, nil
}

// Del 删除订阅调度
func (ss *SubScheduler) Del() error {
	return DB.Delete(ss).Error
}

// UpdateRunTime 更新运行时间
func (ss *SubScheduler) UpdateRunTime(lastRun, nextRun *time.Time) error {
	return DB.Model(ss).Select("LastRunTime", "NextRunTime").Updates(map[string]interface{}{
		"LastRunTime": lastRun,
		"NextRunTime": nextRun,
	}).Error
}

// GetByID 根据ID获取订阅调度
func (ss *SubScheduler) GetByID(id int) error {
	return DB.Where("id = ?", id).First(ss).Error
}
