package models

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"time"

	"gorm.io/gorm"
)

// Template 模板数据模型
type Template struct {
	ID               int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name             string    `gorm:"size:191;uniqueIndex" json:"name"`      // 文件名
	Description      string    `gorm:"size:512" json:"description"`           // 模板描述
	Category         string    `gorm:"default:'clash'" json:"category"`       // clash / surge
	RuleSource       string    `gorm:"type:text" json:"ruleSource"`           // 远程规则配置地址
	UseProxy         bool      `gorm:"default:false" json:"useProxy"`         // 是否使用代理下载远程规则
	ProxyLink        string    `gorm:"type:text" json:"proxyLink"`            // 代理节点链接
	EnableIncludeAll bool      `gorm:"default:false" json:"enableIncludeAll"` // 是否启用 include-all 模式
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// templateCache 模板缓存
var templateCache *cache.MapCache[int, Template]

func init() {
	templateCache = cache.NewMapCache(func(t Template) int { return t.ID })
}

// InferTemplateCategory 根据文件名推断模板类别
func InferTemplateCategory(fileName string) string {
	if strings.EqualFold(filepath.Ext(fileName), ".conf") {
		return "surge"
	}
	return "clash"
}

// InitTemplateCache 初始化模板缓存
func InitTemplateCache() error {
	utils.Info("Loading templates into cache")

	// 添加 Name 二级索引
	templateCache.AddIndex("name", func(t Template) string { return t.Name })

	var templates []Template
	if err := database.DB.Find(&templates).Error; err != nil {
		return err
	}

	templateCache.LoadAll(templates)
	utils.Info("Template cache initialized with %d templates", templateCache.Count())

	cache.Manager.Register("template", templateCache)
	return nil
}

// Add 添加模板
func (t *Template) Add() error {
	if err := database.DB.Create(t).Error; err != nil {
		return err
	}
	templateCache.Set(t.ID, *t)
	return nil
}

// Update 更新模板
func (t *Template) Update() error {
	if err := database.DB.Save(t).Error; err != nil {
		return err
	}
	templateCache.Set(t.ID, *t)
	return nil
}

// Delete 删除模板
func (t *Template) Delete() error {
	if err := database.DB.Delete(t).Error; err != nil {
		return err
	}
	templateCache.Delete(t.ID)
	return nil
}

// FindByName 根据名称查找模板
func (t *Template) FindByName(name string) error {
	// 先从缓存的二级索引查找
	results := templateCache.GetByIndex("name", name)
	if len(results) > 0 {
		*t = results[0]
		return nil
	}

	// 缓存未命中，从数据库查找
	if err := database.DB.Where("name = ?", name).First(t).Error; err != nil {
		return err
	}

	// 更新缓存
	templateCache.Set(t.ID, *t)
	return nil
}

// FindByID 根据ID查找模板
func (t *Template) FindByID(id int) error {
	if cached, ok := templateCache.Get(id); ok {
		*t = cached
		return nil
	}

	if err := database.DB.First(t, id).Error; err != nil {
		return err
	}

	templateCache.Set(t.ID, *t)
	return nil
}

// List 获取所有模板
func (t *Template) List() ([]Template, error) {
	// 尝试从缓存获取
	if templateCache.Count() > 0 {
		return templateCache.GetAll(), nil
	}

	// 从数据库获取
	var templates []Template
	if err := database.DB.Find(&templates).Error; err != nil {
		return nil, err
	}

	// 更新缓存
	for _, tmpl := range templates {
		templateCache.Set(tmpl.ID, tmpl)
	}

	return templates, nil
}

// MigrateTemplatesFromFiles 从文件系统迁移现有模板到数据库
func MigrateTemplatesFromFiles(templateDir string) error {
	// 检查目录是否存在
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		utils.Warn("Template directory does not exist; skipping migration: %s", templateDir)
		return nil
	}

	files, err := os.ReadDir(templateDir)
	if err != nil {
		return err
	}

	migratedCount := 0
	updatedCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		category := InferTemplateCategory(fileName)

		// 检查是否已存在
		var existing Template
		err := database.DB.Where("name = ?", fileName).First(&existing).Error
		if err == nil {
			if existing.Category == "" || (existing.Category != "clash" && existing.Category != "surge") {
				existing.Category = category
				if err := database.DB.Model(&existing).Update("category", category).Error; err != nil {
					utils.Error("Failed to fix template category %s: %v", fileName, err)
					continue
				}
				updatedCount++
				utils.Info("Fixed template category: %s (category: %s)", fileName, category)
			}
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// 创建模板记录
		template := Template{
			Name:       fileName,
			Category:   category,
			RuleSource: "",
		}

		if err := database.DB.Create(&template).Error; err != nil {
			utils.Error("Failed to migrate template %s: %v", fileName, err)
			continue
		}

		migratedCount++
		utils.Info("Migrated template: %s (category: %s)", fileName, category)
	}

	if migratedCount > 0 {
		utils.Info("Template migration completed; %d templates migrated", migratedCount)
	}
	if updatedCount > 0 {
		utils.Info("Template category fix completed; %d templates fixed", updatedCount)
	}

	return nil
}
