package models

import (
	"time"
)

type Script struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Version   string    `json:"version" gorm:"default:0.0.0"`
	Content   string    `json:"content" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Add 添加脚本
func (s *Script) Add() error {
	return DB.Create(s).Error
}

// Update 更新脚本
func (s *Script) Update() error {
	return DB.Model(s).Updates(s).Error
}

// Del 删除脚本
func (s *Script) Del() error {
	return DB.Delete(s).Error
}

// List 获取脚本列表
func (s *Script) List() ([]Script, error) {
	var scripts []Script
	err := DB.Find(&scripts).Error
	return scripts, err
}

// CheckNameVersion 检查名称和版本是否重复
func (s *Script) CheckNameVersion() bool {
	var count int64
	DB.Model(&Script{}).Where("name = ? AND version = ? AND id != ?", s.Name, s.Version, s.ID).Count(&count)
	return count > 0
}
