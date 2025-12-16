package models

import (
	"sort"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
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

// scriptCache 使用新的泛型缓存
var scriptCache *cache.MapCache[int, Script]

func init() {
	scriptCache = cache.NewMapCache(func(s Script) int { return s.ID })
	scriptCache.AddIndex("name", func(s Script) string { return s.Name })
}

// InitScriptCache 初始化脚本缓存
func InitScriptCache() error {
	utils.Info("开始加载脚本到缓存")
	var scripts []Script
	if err := database.DB.Find(&scripts).Error; err != nil {
		return err
	}

	scriptCache.LoadAll(scripts)
	utils.Info("脚本缓存初始化完成，共加载 %d 个脚本", scriptCache.Count())

	cache.Manager.Register("script", scriptCache)
	return nil
}

// Add 添加脚本 (Write-Through)
func (s *Script) Add() error {
	err := database.DB.Create(s).Error
	if err != nil {
		return err
	}
	scriptCache.Set(s.ID, *s)
	return nil
}

// Update 更新脚本 (Write-Through)
func (s *Script) Update() error {
	err := database.DB.Model(s).Updates(s).Error
	if err != nil {
		return err
	}
	// 从DB读取完整数据后更新缓存
	var updated Script
	if err := database.DB.First(&updated, s.ID).Error; err == nil {
		scriptCache.Set(s.ID, updated)
	}
	return nil
}

// Del 删除脚本 (Write-Through)
func (s *Script) Del() error {
	err := database.DB.Delete(s).Error
	if err != nil {
		return err
	}
	scriptCache.Delete(s.ID)
	return nil
}

// List 获取脚本列表
func (s *Script) List() ([]Script, error) {
	scripts := scriptCache.GetAllSorted(func(a, b Script) bool {
		return a.ID < b.ID
	})
	return scripts, nil
}

// ListPaginated 分页获取脚本列表
func (s *Script) ListPaginated(page, pageSize int) ([]Script, int64, error) {
	allScripts := scriptCache.GetAllSorted(func(a, b Script) bool {
		return a.ID < b.ID
	})
	total := int64(len(allScripts))

	if page <= 0 || pageSize <= 0 {
		return allScripts, total, nil
	}

	offset := (page - 1) * pageSize
	if offset >= len(allScripts) {
		return []Script{}, total, nil
	}

	end := offset + pageSize
	if end > len(allScripts) {
		end = len(allScripts)
	}

	return allScripts[offset:end], total, nil
}

// CheckNameVersion 检查名称和版本是否重复
func (s *Script) CheckNameVersion() bool {
	// 使用缓存查询
	scripts := scriptCache.GetByIndex("name", s.Name)
	for _, script := range scripts {
		if script.Version == s.Version && script.ID != s.ID {
			return true
		}
	}
	return false
}

// GetScriptByID 根据ID获取脚本
func GetScriptByID(id int) (*Script, error) {
	if script, ok := scriptCache.Get(id); ok {
		return &script, nil
	}
	var script Script
	if err := database.DB.First(&script, id).Error; err != nil {
		return nil, err
	}
	scriptCache.Set(script.ID, script)
	return &script, nil
}

// Ensure sort is used
var _ = sort.Slice
