package protocol

import (
	"reflect"
	"sort"
	"strings"
	"sync"
)

// FieldMeta 字段元数据
type FieldMeta struct {
	Name  string `json:"name"`  // 字段名称
	Label string `json:"label"` // 显示标签
	Type  string `json:"type"`  // 字段类型
}

// ProtocolMeta 协议元数据
type ProtocolMeta struct {
	Name   string      `json:"name"`   // 协议名称（小写）
	Label  string      `json:"label"`  // 显示名称
	Color  string      `json:"color"`  // 主题颜色（用于前端展示）
	Icon   string      `json:"icon"`   // 图标字符（用于前端展示）
	Fields []FieldMeta `json:"fields"` // 可用字段列表
}

// 全局缓存，系统启动时初始化
var (
	protocolMetaCache []ProtocolMeta
	metaOnce          sync.Once
)

// protocolRegistry 协议注册表，包含协议名称、显示标签、URL前缀和结构体实例
type protocolRegistry struct {
	name     string      // 协议名称（小写，存储到数据库）
	label    string      // 显示名称
	color    string      // 主题颜色
	icon     string      // 图标字符
	prefixes []string    // URL 前缀列表（支持多个，如 hy2:// 和 hysteria2://）
	instance interface{} // 结构体实例（用于反射，可为nil表示不支持解析）
}

// registeredProtocols 全局协议注册表
var registeredProtocols = []protocolRegistry{
	{name: "vmess", label: "VMess", color: "#1976d2", icon: "V", prefixes: []string{"vmess://"}, instance: Vmess{}},
	{name: "vless", label: "VLESS", color: "#7b1fa2", icon: "V", prefixes: []string{"vless://"}, instance: VLESS{}},
	{name: "trojan", label: "Trojan", color: "#d32f2f", icon: "T", prefixes: []string{"trojan://"}, instance: Trojan{}},
	{name: "ss", label: "SS", color: "#2e7d32", icon: "S", prefixes: []string{"ss://"}, instance: Ss{}},
	{name: "ssr", label: "SSR", color: "#e64a19", icon: "R", prefixes: []string{"ssr://"}, instance: Ssr{}},
	{name: "hysteria", label: "Hysteria", color: "#f9a825", icon: "H", prefixes: []string{"hysteria://", "hy://"}, instance: HY{}},
	{name: "hysteria2", label: "Hysteria2", color: "#ef6c00", icon: "H", prefixes: []string{"hysteria2://", "hy2://"}, instance: HY2{}},
	{name: "tuic", label: "TUIC", color: "#0277bd", icon: "T", prefixes: []string{"tuic://"}, instance: Tuic{}},
	{name: "wireguard", label: "WireGuard", color: "#88171a", icon: "W", prefixes: []string{"wg://", "wireguard://"}, instance: nil},
	{name: "naiveproxy", label: "NaiveProxy", color: "#5d4037", icon: "N", prefixes: []string{"naive://"}, instance: nil},
	{name: "anytls", label: "AnyTLS", color: "#20a84c", icon: "A", prefixes: []string{"anytls://"}, instance: AnyTLS{}},
	{name: "socks5", label: "SOCKS5", color: "#116ea4", icon: "S", prefixes: []string{"socks5://"}, instance: Socks5{}},
	{name: "socks", label: "SOCKS", color: "#dd4984", icon: "S", prefixes: []string{"socks://"}, instance: nil},
}

// InitProtocolMeta 系统启动时调用，通过反射扫描所有协议结构体
func InitProtocolMeta() {
	metaOnce.Do(func() {
		for _, proto := range registeredProtocols {
			// 确保 fields 为空切片而非 nil，JSON 序列化时输出 [] 而非 null
			fields := []FieldMeta{}
			if proto.instance != nil {
				fields = extractFields(proto.instance)
			}
			meta := ProtocolMeta{
				Name:   proto.name,
				Label:  proto.label,
				Color:  proto.color,
				Icon:   proto.icon,
				Fields: fields,
			}
			protocolMetaCache = append(protocolMetaCache, meta)
		}

		// 按名称排序，保证返回顺序稳定
		sort.Slice(protocolMetaCache, func(i, j int) bool {
			return protocolMetaCache[i].Name < protocolMetaCache[j].Name
		})
	})
}

// extractFields 使用反射提取结构体字段
func extractFields(v interface{}) []FieldMeta {
	var fields []FieldMeta
	t := reflect.TypeOf(v)

	// 递归处理结构体
	extractFieldsRecursive(t, "", &fields)
	return fields
}

// extractFieldsRecursive 递归提取结构体字段
func extractFieldsRecursive(t reflect.Type, prefix string, fields *[]FieldMeta) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过非导出字段
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		// 获取json标签作为Label
		jsonTag := field.Tag.Get("json")
		label := strings.Split(jsonTag, ",")[0]
		if label == "" || label == "-" {
			label = field.Name
		}

		kind := field.Type.Kind()
		switch kind {
		case reflect.String:
			*fields = append(*fields, FieldMeta{
				Name:  fieldName,
				Label: label,
				Type:  "string",
			})
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			*fields = append(*fields, FieldMeta{
				Name:  fieldName,
				Label: label,
				Type:  "int",
			})
		case reflect.Bool:
			*fields = append(*fields, FieldMeta{
				Name:  fieldName,
				Label: label,
				Type:  "bool",
			})
		case reflect.Struct:
			// 递归处理嵌套结构体（如VLESSQuery, TrojanQuery等）
			extractFieldsRecursive(field.Type, fieldName, fields)
		}
		// 跳过数组、切片等复杂类型（如ALPN []string）
	}
}

// GetAllProtocolMeta 获取缓存的协议元数据
func GetAllProtocolMeta() []ProtocolMeta {
	return protocolMetaCache
}

// GetProtocolFieldValue 从解析后的协议对象中获取指定字段的值
// 使用反射动态获取字段值，支持嵌套字段（如 Query.Sni）
func GetProtocolFieldValue(protoObj interface{}, fieldPath string) string {
	if protoObj == nil {
		return ""
	}

	v := reflect.ValueOf(protoObj)
	// 处理指针类型
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	// 处理嵌套字段路径（如 Query.Sni）
	parts := strings.Split(fieldPath, ".")
	for _, part := range parts {
		if v.Kind() != reflect.Struct {
			return ""
		}
		v = v.FieldByName(part)
		if !v.IsValid() {
			return ""
		}
	}

	// 转换为字符串
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strings.TrimSpace(strings.Replace(reflect.ValueOf(v.Int()).String(), " ", "", -1))
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Interface:
		// 处理 interface{} 类型（如 vmess 的 Port）
		if v.IsNil() {
			return ""
		}
		return strings.TrimSpace(reflect.ValueOf(v.Interface()).String())
	default:
		return ""
	}
}

// GetProtocolFromLink 从节点链接解析协议类型
// 返回标准化的协议名称（小写），用于存储到数据库
func GetProtocolFromLink(link string) string {
	if link == "" {
		return "unknown"
	}
	linkLower := strings.ToLower(link)
	for _, proto := range registeredProtocols {
		for _, prefix := range proto.prefixes {
			if strings.HasPrefix(linkLower, prefix) {
				return proto.name
			}
		}
	}
	return "other"
}

// GetProtocolLabel 根据协议名称获取显示标签
func GetProtocolLabel(name string) string {
	for _, proto := range registeredProtocols {
		if proto.name == name {
			return proto.label
		}
	}
	return name
}

// GetAllProtocolNames 获取所有支持的协议名称列表
func GetAllProtocolNames() []string {
	names := make([]string, len(registeredProtocols))
	for i, proto := range registeredProtocols {
		names[i] = proto.name
	}
	return names
}

// GetProtocolMeta 根据协议名称获取完整的协议元数据
func GetProtocolMeta(name string) *ProtocolMeta {
	for i := range protocolMetaCache {
		if protocolMetaCache[i].Name == name {
			return &protocolMetaCache[i]
		}
	}
	return nil
}
