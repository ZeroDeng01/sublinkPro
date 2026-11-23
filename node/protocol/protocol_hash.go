package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sort"

	"gopkg.in/yaml.v3"
)

// HashFieldConfig 哈希字段配置
type HashFieldConfig struct {
	// IgnoredFields 忽略的字段（不参与哈希计算）
	IgnoredFields map[string]bool
}

type proxyHashNormalizer interface {
	normalizeProxyForHash(Proxy) Proxy
}

// defaultHashConfig 默认哈希配置（集中管理，便于扩展）
// Type 字段参与哈希计算，可以自动区分不同协议
var defaultHashConfig = HashFieldConfig{
	IgnoredFields: map[string]bool{
		"Name":               true, // 节点名称不影响节点功能
		"Dialer_proxy":       true, // 前置代理是用户配置
		"Udp":                true, // UDP 支持是用户可配置选项
		"Tfo":                true, // TCP Fast Open 是用户可配置选项
		"Client_fingerprint": true, // uTLS 指纹，某些机场每次返回随机值
	},
}

// GenerateProxyContentHash 生成 Proxy 的内容哈希
// 使用 SHA256 算法，返回 64 字符的十六进制字符串
// 所有非忽略的非空字段都参与哈希计算，按字段名排序保证一致性
func GenerateProxyContentHash(proxy Proxy) string {
	if normalizer, ok := getProtocolByName(proxy.Type).(proxyHashNormalizer); ok {
		proxy = normalizer.normalizeProxyForHash(proxy)
	}
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
func extractProxyFields(proxy Proxy, ignoredFields map[string]bool) map[string]any {
	result := make(map[string]any)
	v := reflect.ValueOf(proxy)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name

		// 跳过忽略的字段
		if ignoredFields[fieldName] {
			continue
		}
		if fieldName == "ClashExtra" {
			// Hash retained values, without yaml.Node source positions, comments or style.
			// Keep the historical null value when there are no extra attributes.
			result[fieldName] = hashClashExtra(&proxy.ClashExtra)
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

func hashClashExtra(extra *yaml.Node) any {
	mapping := yamlMappingNode(extra)
	if mapping == nil || len(mapping.Content) == 0 {
		return nil
	}
	var value any
	if err := mapping.Decode(&value); err == nil {
		if encoded, err := json.Marshal(normalizeHashYAMLValue(value)); err == nil {
			return json.RawMessage(encoded)
		}
	}
	// Retained YAML can contain valid constructs Go cannot decode into JSON data,
	// such as sequence keys. Keep those imports and hash their structure instead.
	return hashYAMLStructure(mapping)
}

func hashYAMLStructure(node *yaml.Node) any {
	switch node.Kind {
	case yaml.MappingNode:
		entries := make(map[string]any, len(node.Content)/2)
		for i := 0; i+1 < len(node.Content); i += 2 {
			key, _ := json.Marshal(hashYAMLStructure(node.Content[i]))
			entries[string(key)] = hashYAMLStructure(node.Content[i+1])
		}
		return entries
	case yaml.SequenceNode:
		values := make([]any, len(node.Content))
		for i, child := range node.Content {
			values[i] = hashYAMLStructure(child)
		}
		return values
	default:
		return [3]any{node.Kind, node.Tag, node.Value}
	}
}

func normalizeHashYAMLValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			value[key] = normalizeHashYAMLValue(child)
		}
	case map[any]any:
		// JSON requires string keys. Preserve YAML key types in a deterministic encoding.
		entries := make(map[string]any, len(value))
		for key, child := range value {
			encoded, err := json.Marshal(key)
			if err != nil {
				return value // Let the hash encoder reject unrepresentable data.
			}
			entries[string(encoded)] = normalizeHashYAMLValue(child)
		}
		return entries
	case []any:
		for i, child := range value {
			value[i] = normalizeHashYAMLValue(child)
		}
	}
	return value
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
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	default:
		return false
	}
}

// normalizeValue 规范化字段值，确保序列化结果一致
func normalizeValue(v reflect.Value) any {
	// 如果是指针或接口，先解引用
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Bool:
		return v.Bool()
	case reflect.Slice, reflect.Array:
		return normalizeSlice(v)
	case reflect.Map:
		return normalizeMap(v)
	case reflect.Struct:
		// 如果是 struct 类型，递归处理其字段
		return normalizeStruct(v)
	default:
		// 对于其他类型，尝试转换为基础类型
		// 处理类似 FlexPort（底层是 int）的自定义类型
		underlyingKind := v.Type().Kind()
		if underlyingKind == reflect.Int || underlyingKind == reflect.Int64 {
			return v.Int()
		}
		// 如果无法处理，返回 nil 避免序列化不稳定
		return nil
	}
}

// normalizeSlice 规范化切片类型
func normalizeSlice(v reflect.Value) any {
	if v.Len() == 0 {
		return nil
	}

	elemKind := v.Type().Elem().Kind()

	// 对于字符串切片，排序后返回
	if elemKind == reflect.String {
		strs := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			strs[i] = v.Index(i).String()
		}
		sort.Strings(strs)
		return strs
	}

	// 对于 int 切片，直接返回
	if elemKind == reflect.Int || elemKind == reflect.Int64 {
		ints := make([]int64, v.Len())
		for i := 0; i < v.Len(); i++ {
			ints[i] = v.Index(i).Int()
		}
		return ints
	}

	// 其他类型切片，递归处理每个元素
	result := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = normalizeValue(v.Index(i))
	}
	return result
}

// normalizeStruct 规范化 struct 类型
func normalizeStruct(v reflect.Value) map[string]any {
	result := make(map[string]any)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		// 跳过未导出字段
		if field.PkgPath != "" {
			continue
		}
		fieldValue := v.Field(i)
		if !isZeroValue(fieldValue) {
			normalized := normalizeValue(fieldValue)
			if normalized != nil {
				result[field.Name] = normalized
			}
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// normalizeMap 规范化 map 类型，递归处理嵌套结构
func normalizeMap(v reflect.Value) map[string]any {
	if v.Len() == 0 {
		return nil
	}

	result := make(map[string]any)

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
		if !mapValue.IsValid() {
			continue
		}

		// 使用统一的 normalizeValue 处理
		normalized := normalizeValue(mapValue)
		if normalized != nil {
			result[key] = normalized
		}
	}

	if len(result) == 0 {
		return nil
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
func NormalizeProxyForHash(proxy Proxy) map[string]any {
	return extractProxyFields(proxy, defaultHashConfig.IgnoredFields)
}
