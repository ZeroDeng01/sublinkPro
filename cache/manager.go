package cache

import (
	"sublink/utils"
	"sync"
)

// CacheManager 全局缓存管理器
// 用于统一管理和初始化所有模块的缓存
type CacheManager struct {
	caches       map[string]interface{}
	initializers map[string]func() error
	lock         sync.RWMutex
}

// Manager 全局缓存管理器实例
var Manager = &CacheManager{
	caches:       make(map[string]interface{}),
	initializers: make(map[string]func() error),
}

// Register 注册模块缓存
// name: 模块名称
// cache: 缓存实例
func (m *CacheManager) Register(name string, cache interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.caches[name] = cache
	utils.Info("缓存模块注册成功: %s", name)
}

// RegisterWithInit 注册模块缓存及其初始化函数
// name: 模块名称
// cache: 缓存实例
// initFunc: 初始化函数（从数据库加载数据）
func (m *CacheManager) RegisterWithInit(name string, cache interface{}, initFunc func() error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.caches[name] = cache
	m.initializers[name] = initFunc
	utils.Info("缓存模块注册成功（含初始化器）: %s", name)
}

// Get 获取指定模块的缓存
func (m *CacheManager) Get(name string) (interface{}, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	cache, exists := m.caches[name]
	return cache, exists
}

// InitAll 初始化所有已注册的缓存
func (m *CacheManager) InitAll() error {
	m.lock.RLock()
	defer m.lock.RUnlock()

	utils.Info("开始初始化所有缓存模块...")
	for name, initFunc := range m.initializers {
		utils.Info("正在初始化缓存: %s", name)
		if err := initFunc(); err != nil {
			utils.Error("缓存初始化失败 [%s]: %v", name, err)
			return err
		}
		utils.Info("缓存初始化成功: %s", name)
	}
	utils.Info("所有缓存模块初始化完成")
	return nil
}

// Stats 获取所有缓存的统计信息
func (m *CacheManager) Stats() map[string]int {
	m.lock.RLock()
	defer m.lock.RUnlock()

	stats := make(map[string]int)
	for name, cache := range m.caches {
		// 尝试获取缓存大小
		if counter, ok := cache.(interface{ Count() int }); ok {
			stats[name] = counter.Count()
		}
	}
	return stats
}

// List 列出所有已注册的缓存模块名称
func (m *CacheManager) List() []string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	names := make([]string, 0, len(m.caches))
	for name := range m.caches {
		names = append(names, name)
	}
	return names
}
