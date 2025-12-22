package models

import (
	"os"
	"path/filepath"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"time"
)

// Template 模板数据模型
type Template struct {
	ID               int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name             string    `gorm:"uniqueIndex" json:"name"`               // 文件名
	Category         string    `gorm:"default:'clash'" json:"category"`       // clash / surge
	RuleSource       string    `gorm:"default:''" json:"ruleSource"`          // 远程规则配置地址
	UseProxy         bool      `gorm:"default:false" json:"useProxy"`         // 是否使用代理下载远程规则
	ProxyLink        string    `gorm:"default:''" json:"proxyLink"`           // 代理节点链接
	EnableIncludeAll bool      `gorm:"default:false" json:"enableIncludeAll"` // 是否启用 include-all 模式
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// templateCache 模板缓存
var templateCache *cache.MapCache[int, Template]

func init() {
	templateCache = cache.NewMapCache(func(t Template) int { return t.ID })
}

// InitTemplateCache 初始化模板缓存
func InitTemplateCache() error {
	utils.Info("开始加载模板到缓存")

	// 添加 Name 二级索引
	templateCache.AddIndex("name", func(t Template) string { return t.Name })

	var templates []Template
	if err := database.DB.Find(&templates).Error; err != nil {
		return err
	}

	templateCache.LoadAll(templates)
	utils.Info("模板缓存初始化完成，共加载 %d 个模板", templateCache.Count())

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
		utils.Warn("模板目录不存在，跳过迁移: %s", templateDir)
		return nil
	}

	files, err := os.ReadDir(templateDir)
	if err != nil {
		return err
	}

	migratedCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()

		// 检查是否已存在
		var existing Template
		if err := database.DB.Where("name = ?", fileName).First(&existing).Error; err == nil {
			// 已存在，跳过
			continue
		}

		// 根据扩展名推断类别
		category := "clash"
		ext := strings.ToLower(filepath.Ext(fileName))
		if ext == ".conf" {
			category = "surge"
		}

		// 创建模板记录
		template := Template{
			Name:       fileName,
			Category:   category,
			RuleSource: "",
		}

		if err := database.DB.Create(&template).Error; err != nil {
			utils.Error("迁移模板失败 %s: %v", fileName, err)
			continue
		}

		migratedCount++
		utils.Info("已迁移模板: %s (类别: %s)", fileName, category)
	}

	if migratedCount > 0 {
		utils.Info("模板迁移完成，共迁移 %d 个模板", migratedCount)
	}

	return nil
}
