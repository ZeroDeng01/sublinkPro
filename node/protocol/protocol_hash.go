package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sort"
)

// HashFieldConfig 哈希字段配置
type HashFieldConfig struct {
	// IgnoredFields 忽略的字段（不参与哈希计算）
	IgnoredFields map[string]bool
}

// defaultHashConfig 默认哈希配置（集中管理，便于扩展）
// Type 字段参与哈希计算，可以自动区分不同协议
var defaultHashConfig = HashFieldConfig{
	IgnoredFields: map[string]bool{
		"Name":         true, // 节点名称不影响节点功能
		"Dialer_proxy": true, // 前置代理是用户配置
		"Udp":          true, // UDP 支持是用户可配置选项
		"Tfo":          true, // TCP Fast Open 是用户可配置选项
	},
}

// GenerateProxyContentHash 生成 Proxy 的内容哈希
// 使用 SHA256 算法，返回 64 字符的十六进制字符串
// 所有非忽略的非空字段都参与哈希计算，按字段名排序保证一致性
func GenerateProxyContentHash(proxy Proxy) string {
	// 使用反射提取所有字段值
	data := extractProxyFields(proxy, defaultHashConfig.IgnoredFields)

	// 序列化为 JSON（map 会自动按 key 排序）
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	// 计算 SHA256
	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:])
}

// extractProxyFields 使用反射提取 Proxy 的非忽略、非空字段
// 返回按字段名排序的 map
func extractProxyFields(proxy Proxy, ignoredFields map[string]bool) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(proxy)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name

		// 跳过忽略的字段
		if ignoredFields[fieldName] {
			continue
		}

		fieldValue := v.Field(i)

		// 跳过零值字段
		if isZeroValue(fieldValue) {
			continue
		}

		// 处理不同类型的字段值
		result[fieldName] = normalizeValue(fieldValue)
	}

	return result
}

// isZeroValue 检查字段是否为零值
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Array:
		return v.Len() == 0
	case reflect.Map:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}

// normalizeValue 规范化字段值，确保序列化结果一致
func normalizeValue(v reflect.Value) interface{} {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Bool:
		return v.Bool()
	case reflect.Slice:
		// 对于字符串切片（如 Alpn），排序后返回
		if v.Type().Elem().Kind() == reflect.String {
			strs := make([]string, v.Len())
			for i := 0; i < v.Len(); i++ {
				strs[i] = v.Index(i).String()
			}
			sort.Strings(strs)
			return strs
		}
		// 对于 int 切片（如 Reserved），直接返回
		if v.Type().Elem().Kind() == reflect.Int {
			ints := make([]int, v.Len())
			for i := 0; i < v.Len(); i++ {
				ints[i] = int(v.Index(i).Int())
			}
			return ints
		}
		return v.Interface()
	case reflect.Map:
		// 对于 map[string]interface{}（如 Ws_opts），规范化处理
		return normalizeMap(v)
	default:
		return v.Interface()
	}
}

// normalizeMap 规范化 map 类型，递归处理嵌套结构
func normalizeMap(v reflect.Value) map[string]interface{} {
	result := make(map[string]interface{})

	// 获取所有 key 并排序
	keys := v.MapKeys()
	sortedKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		if k.Kind() == reflect.String {
			sortedKeys = append(sortedKeys, k.String())
		}
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		mapValue := v.MapIndex(reflect.ValueOf(key))
		if !mapValue.IsValid() || isZeroValue(mapValue) {
			continue
		}

		// 处理 interface{} 类型
		if mapValue.Kind() == reflect.Interface {
			mapValue = mapValue.Elem()
		}

		if !mapValue.IsValid() {
			continue
		}

		// 递归处理嵌套 map
		if mapValue.Kind() == reflect.Map {
			result[key] = normalizeMap(mapValue)
		} else if mapValue.Kind() == reflect.String {
			str := mapValue.String()
			if str != "" {
				result[key] = str
			}
		} else {
			result[key] = mapValue.Interface()
		}
	}

	return result
}

// GetHashIgnoredFields 获取当前忽略的字段列表（用于调试或展示）
func GetHashIgnoredFields() []string {
	fields := make([]string, 0, len(defaultHashConfig.IgnoredFields))
	for field := range defaultHashConfig.IgnoredFields {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return fields
}

// SetHashIgnoredField 动态添加忽略字段（用于测试或特殊场景）
func SetHashIgnoredField(fieldName string, ignored bool) {
	if ignored {
		defaultHashConfig.IgnoredFields[fieldName] = true
	} else {
		delete(defaultHashConfig.IgnoredFields, fieldName)
	}
}

// IsFieldIgnoredForHash 检查字段是否被忽略
func IsFieldIgnoredForHash(fieldName string) bool {
	return defaultHashConfig.IgnoredFields[fieldName]
}

// NormalizeProxyForHash 返回用于哈希计算的规范化数据（用于调试）
func NormalizeProxyForHash(proxy Proxy) map[string]interface{} {
	return extractProxyFields(proxy, defaultHashConfig.IgnoredFields)
}
