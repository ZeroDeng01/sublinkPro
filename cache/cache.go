package cache

import (
	"sort"
	"sync"
)

// EntityCache 实体缓存接口
type EntityCache[K comparable, V any] interface {
	Get(key K) (V, bool)
	GetAll() []V
	Set(key K, value V)
	Delete(key K)
	Count() int
	Clear()
}

// MapCache 基于 map 的高性能泛型缓存实现
// 支持主键查询 O(1) 和二级索引查询
type MapCache[K comparable, V any] struct {
	data     map[K]V
	indexes  map[string]*secondaryIndex[K] // field -> index
	lock     sync.RWMutex
	getKey   func(V) K
	indexers map[string]func(V) string
}

// secondaryIndex 二级索引结构
type secondaryIndex[K comparable] struct {
	data map[string][]K // value -> primary keys
}

// NewMapCache 创建新的 MapCache 实例
// getKey: 从实体中提取主键的函数
func NewMapCache[K comparable, V any](getKey func(V) K) *MapCache[K, V] {
	return &MapCache[K, V]{
		data:     make(map[K]V),
		indexes:  make(map[string]*secondaryIndex[K]),
		getKey:   getKey,
		indexers: make(map[string]func(V) string),
	}
}

// AddIndex 添加二级索引
// field: 索引名称
// getField: 从实体中提取索引字段值的函数
func (c *MapCache[K, V]) AddIndex(field string, getField func(V) string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.indexers[field] = getField
	c.indexes[field] = &secondaryIndex[K]{
		data: make(map[string][]K),
	}

	// 为现有数据建立索引
	for key, value := range c.data {
		fieldValue := getField(value)
		c.indexes[field].data[fieldValue] = append(c.indexes[field].data[fieldValue], key)
	}
}

// Get 根据主键获取实体 O(1)
func (c *MapCache[K, V]) Get(key K) (V, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	value, ok := c.data[key]
	return value, ok
}

// GetAll 获取所有实体
func (c *MapCache[K, V]) GetAll() []V {
	c.lock.RLock()
	defer c.lock.RUnlock()

	result := make([]V, 0, len(c.data))
	for _, v := range c.data {
		result = append(result, v)
	}
	return result
}

// GetAllSorted 获取所有实体并按主键排序
func (c *MapCache[K, V]) GetAllSorted(less func(a, b V) bool) []V {
	result := c.GetAll()
	sort.Slice(result, func(i, j int) bool {
		return less(result[i], result[j])
	})
	return result
}

// Set 设置实体到缓存
func (c *MapCache[K, V]) Set(key K, value V) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 如果已存在，先从索引中移除旧值
	if oldValue, exists := c.data[key]; exists {
		c.removeFromIndexes(key, oldValue)
	}

	// 设置新值
	c.data[key] = value

	// 添加到索引
	c.addToIndexes(key, value)
}

// Delete 从缓存中删除实体
func (c *MapCache[K, V]) Delete(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if value, exists := c.data[key]; exists {
		c.removeFromIndexes(key, value)
		delete(c.data, key)
	}
}

// Count 返回缓存中实体数量
func (c *MapCache[K, V]) Count() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.data)
}

// Clear 清空缓存
func (c *MapCache[K, V]) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.data = make(map[K]V)
	for field := range c.indexes {
		c.indexes[field].data = make(map[string][]K)
	}
}

// LoadAll 批量加载数据到缓存
func (c *MapCache[K, V]) LoadAll(items []V) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 清空旧数据
	c.data = make(map[K]V, len(items))
	for field := range c.indexes {
		c.indexes[field].data = make(map[string][]K)
	}

	// 加载新数据
	for _, item := range items {
		key := c.getKey(item)
		c.data[key] = item
		c.addToIndexes(key, item)
	}
}

// GetByIndex 根据二级索引查询实体列表
func (c *MapCache[K, V]) GetByIndex(field, value string) []V {
	c.lock.RLock()
	defer c.lock.RUnlock()

	index, exists := c.indexes[field]
	if !exists {
		return nil
	}

	keys, exists := index.data[value]
	if !exists {
		return nil
	}

	result := make([]V, 0, len(keys))
	for _, key := range keys {
		if v, ok := c.data[key]; ok {
			result = append(result, v)
		}
	}
	return result
}

// GetDistinctIndexValues 获取索引字段的所有不同值
func (c *MapCache[K, V]) GetDistinctIndexValues(field string) []string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	index, exists := c.indexes[field]
	if !exists {
		return nil
	}

	result := make([]string, 0, len(index.data))
	for value := range index.data {
		if value != "" { // 排除空值
			result = append(result, value)
		}
	}
	return result
}

// Filter 根据自定义条件过滤实体
func (c *MapCache[K, V]) Filter(predicate func(V) bool) []V {
	c.lock.RLock()
	defer c.lock.RUnlock()

	result := make([]V, 0)
	for _, v := range c.data {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// FilterSorted 根据自定义条件过滤实体并排序
// 注意：由于 map 迭代顺序是随机的，涉及去重等顺序敏感操作时应使用此方法
func (c *MapCache[K, V]) FilterSorted(predicate func(V) bool, less func(a, b V) bool) []V {
	result := c.Filter(predicate)
	if len(result) > 1 {
		sort.Slice(result, func(i, j int) bool {
			return less(result[i], result[j])
		})
	}
	return result
}

// FilterWithLimit 根据条件过滤并限制返回数量
func (c *MapCache[K, V]) FilterWithLimit(predicate func(V) bool, limit int) []V {
	c.lock.RLock()
	defer c.lock.RUnlock()

	result := make([]V, 0, limit)
	for _, v := range c.data {
		if predicate(v) {
			result = append(result, v)
			if len(result) >= limit {
				break
			}
		}
	}
	return result
}

// addToIndexes 将实体添加到所有索引（调用者需持有写锁）
func (c *MapCache[K, V]) addToIndexes(key K, value V) {
	for field, getField := range c.indexers {
		fieldValue := getField(value)
		c.indexes[field].data[fieldValue] = append(c.indexes[field].data[fieldValue], key)
	}
}

// removeFromIndexes 从所有索引中移除实体（调用者需持有写锁）
func (c *MapCache[K, V]) removeFromIndexes(key K, value V) {
	for field, getField := range c.indexers {
		fieldValue := getField(value)
		keys := c.indexes[field].data[fieldValue]

		// 从切片中移除 key
		for i, k := range keys {
			if k == key {
				c.indexes[field].data[fieldValue] = append(keys[:i], keys[i+1:]...)
				break
			}
		}

		// 如果切片为空，删除该索引条目
		if len(c.indexes[field].data[fieldValue]) == 0 {
			delete(c.indexes[field].data, fieldValue)
		}
	}
}

// HasIndex 检查是否存在指定索引
func (c *MapCache[K, V]) HasIndex(field string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, exists := c.indexes[field]
	return exists
}

// IndexCount 获取索引中不同值的数量
func (c *MapCache[K, V]) IndexCount(field string) int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	index, exists := c.indexes[field]
	if !exists {
		return 0
	}
	return len(index.data)
}
