package models

import (
	"log"
	"sync"

	"gorm.io/gorm/clause"
)

type SystemSetting struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

var (
	settingCache = make(map[string]string)
	settingLock  sync.RWMutex
)

// InitSettingCache 初始化设置缓存
func InitSettingCache() error {
	log.Printf("开始加载系统设置到缓存成功")
	var settings []SystemSetting
	if err := DB.Find(&settings).Error; err != nil {
		return err
	}

	settingLock.Lock()
	defer settingLock.Unlock()

	for _, s := range settings {
		settingCache[s.Key] = s.Value
		log.Printf("加载系统设置【%s】到缓存成功", s.Key)
	}
	log.Printf("开始加载系统设置到缓存结束")

	return nil
}

// GetSetting 获取设置
func GetSetting(key string) (string, error) {
	// 先从缓存读取
	settingLock.RLock()
	if val, ok := settingCache[key]; ok {
		settingLock.RUnlock()
		return val, nil
	}
	settingLock.RUnlock()

	// 缓存不存在，从数据库读取
	var setting SystemSetting
	err := DB.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return "", err
	}

	// 更新缓存
	settingLock.Lock()
	settingCache[key] = setting.Value
	settingLock.Unlock()

	return setting.Value, nil
}

// SetSetting 保存设置
func SetSetting(key string, value string) error {
	setting := SystemSetting{
		Key:   key,
		Value: value,
	}
	err := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&setting).Error

	if err != nil {
		return err
	}

	// 更新缓存
	settingLock.Lock()
	settingCache[key] = value
	settingLock.Unlock()

	return nil
}
