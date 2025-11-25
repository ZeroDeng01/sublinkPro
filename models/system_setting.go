package models

import "gorm.io/gorm/clause"

type SystemSetting struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

// GetSetting 获取设置
func GetSetting(key string) (string, error) {
	var setting SystemSetting
	err := DB.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

// SetSetting 保存设置
func SetSetting(key string, value string) error {
	setting := SystemSetting{
		Key:   key,
		Value: value,
	}
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&setting).Error
}
