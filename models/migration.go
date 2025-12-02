package models

import (
	"log"
	"time"
)

type Migration struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
}

// RunAutoMigrate 执行自动迁移，如果 migrationID 已存在则跳过
func RunAutoMigrate(migrationID string, dst ...interface{}) error {
	// 确保 Migration 表存在
	if !DB.Migrator().HasTable(&Migration{}) {
		if err := DB.AutoMigrate(&Migration{}); err != nil {
			return err
		}
	}

	var count int64
	DB.Model(&Migration{}).Where("id = ?", migrationID).Count(&count)
	if count > 0 {
		// 已经执行过，跳过
		return nil
	}
	log.Printf("执行数据库升级任务：%s", migrationID)
	// 执行迁移
	if err := DB.AutoMigrate(dst...); err != nil {
		return err
	}

	// 记录迁移
	return DB.Create(&Migration{
		ID:        migrationID,
		CreatedAt: time.Now(),
	}).Error
}

// RunCustomMigration 执行自定义迁移逻辑，如果 migrationID 已存在则跳过
func RunCustomMigration(migrationID string, action func() error) error {
	// 确保 Migration 表存在
	if !DB.Migrator().HasTable(&Migration{}) {
		if err := DB.AutoMigrate(&Migration{}); err != nil {
			return err
		}
	}

	var count int64
	DB.Model(&Migration{}).Where("id = ?", migrationID).Count(&count)
	if count > 0 {
		// 已经执行过，跳过
		return nil
	}
	log.Printf("执行数据库升级任务：%s", migrationID)

	// 执行自定义迁移逻辑
	if err := action(); err != nil {
		return err
	}

	// 记录迁移
	return DB.Create(&Migration{
		ID:        migrationID,
		CreatedAt: time.Now(),
	}).Error
}
