package cache

import (
	"sync"
	"time"
)

// TemplateContent 模板内容结构
type TemplateContent struct {
	Name     string    // 模板文件名
	Content  string    // 文件内容
	LoadedAt time.Time // 加载时间
}

// templateContentCache 模板内容缓存（文件名 -> 内容）
var templateContentCache *MapCache[string, TemplateContent]
var templateContentOnce sync.Once

// getTemplateContentCache 获取或初始化模板内容缓存
func getTemplateContentCache() *MapCache[string, TemplateContent] {
	templateContentOnce.Do(func() {
		templateContentCache = NewMapCache(func(tc TemplateContent) string { return tc.Name })
	})
	return templateContentCache
}

// GetTemplateContent 获取模板内容（优先从缓存读取）
// 返回内容和是否命中缓存
func GetTemplateContent(filename string) (string, bool) {
	cache := getTemplateContentCache()
	if cached, ok := cache.Get(filename); ok {
		return cached.Content, true
	}
	return "", false
}

// SetTemplateContent 设置模板内容缓存
func SetTemplateContent(filename, content string) {
	cache := getTemplateContentCache()
	cache.Set(filename, TemplateContent{
		Name:     filename,
		Content:  content,
		LoadedAt: time.Now(),
	})
}

// InvalidateTemplateContent 使模板内容缓存失效
func InvalidateTemplateContent(filename string) {
	cache := getTemplateContentCache()
	cache.Delete(filename)
}

// InvalidateAllTemplateContent 清空所有模板内容缓存
func InvalidateAllTemplateContent() {
	cache := getTemplateContentCache()
	cache.Clear()
}

// InitTemplateContentCache 初始化模板内容缓存（注册到管理器）
func InitTemplateContentCache() {
	cache := getTemplateContentCache()
	Manager.Register("templateContent", cache)
}
