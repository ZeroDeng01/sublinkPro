package models

import (
	"sublink/cache"
	"sublink/database"
	"sublink/utils"

	"gorm.io/gorm/clause"
)

type SystemSetting struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

// settingCache 使用新的泛型缓存，主键为 Key
var settingCache *cache.MapCache[string, SystemSetting]

func init() {
	settingCache = cache.NewMapCache(func(s SystemSetting) string { return s.Key })
}

// InitSettingCache 初始化设置缓存
func InitSettingCache() error {
	utils.Info("开始加载系统设置到缓存")
	var settings []SystemSetting
	if err := database.DB.Find(&settings).Error; err != nil {
		return err
	}

	// 使用批量加载方式初始化缓存
	settingCache.LoadAll(settings)
	utils.Info("系统设置缓存初始化完成，共加载 %d 个设置", settingCache.Count())

	// 注册到缓存管理器
	cache.Manager.Register("system_setting", settingCache)
	return nil
}

// GetSetting 获取设置
func GetSetting(key string) (string, error) {
	// 先从缓存读取
	if setting, ok := settingCache.Get(key); ok {
		return setting.Value, nil
	}

	// 缓存不存在，从数据库读取
	var setting SystemSetting
	err := database.DB.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return "", err
	}

	// 更新缓存
	settingCache.Set(key, setting)
	return setting.Value, nil
}

// SetSetting 保存设置 (Write-Through)
func SetSetting(key string, value string) error {
	setting := SystemSetting{
		Key:   key,
		Value: value,
	}
	// Write-Through: 先写数据库
	err := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&setting).Error

	if err != nil {
		return err
	}

	// 再更新缓存
	settingCache.Set(key, setting)
	return nil
}

// GetAllSettings 获取所有设置
func GetAllSettings() map[string]string {
	result := make(map[string]string)
	allSettings := settingCache.GetAll()
	for _, s := range allSettings {
		result[s.Key] = s.Value
	}
	return result
}
