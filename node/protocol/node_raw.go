package protocol

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ParsedNodeInfo 解析后的节点原始信息
type ParsedNodeInfo struct {
	Protocol string                 `json:"protocol"` // 协议类型
	Fields   map[string]interface{} `json:"fields"`   // 字段键值对
}

// ParseNodeLink 解析节点链接，返回协议类型和所有字段的键值对
func ParseNodeLink(link string) (*ParsedNodeInfo, error) {
	if link == "" {
		return nil, fmt.Errorf("链接不能为空")
	}

	linkLower := strings.ToLower(link)
	var protoObj interface{}
	var protocol string
	var err error

	switch {
	case strings.HasPrefix(linkLower, "vmess://"):
		protocol = "vmess"
		protoObj, err = DecodeVMESSURL(link)
	case strings.HasPrefix(linkLower, "vless://"):
		protocol = "vless"
		protoObj, err = DecodeVLESSURL(link)
	case strings.HasPrefix(linkLower, "trojan://"):
		protocol = "trojan"
		protoObj, err = DecodeTrojanURL(link)
	case strings.HasPrefix(linkLower, "ss://"):
		protocol = "ss"
		protoObj, err = DecodeSSURL(link)
	case strings.HasPrefix(linkLower, "ssr://"):
		protocol = "ssr"
		protoObj, err = DecodeSSRURL(link)
	case strings.HasPrefix(linkLower, "hysteria://"), strings.HasPrefix(linkLower, "hy://"):
		protocol = "hysteria"
		protoObj, err = DecodeHYURL(link)
	case strings.HasPrefix(linkLower, "hysteria2://"), strings.HasPrefix(linkLower, "hy2://"):
		protocol = "hysteria2"
		protoObj, err = DecodeHY2URL(link)
	case strings.HasPrefix(linkLower, "tuic://"):
		protocol = "tuic"
		protoObj, err = DecodeTuicURL(link)
	case strings.HasPrefix(linkLower, "anytls://"):
		protocol = "anytls"
		protoObj, err = DecodeAnyTLSURL(link)
	case strings.HasPrefix(linkLower, "socks5://"):
		protocol = "socks5"
		protoObj, err = DecodeSocks5URL(link)
	default:
		return nil, fmt.Errorf("不支持的协议类型")
	}

	if err != nil {
		return nil, fmt.Errorf("解析链接失败: %w", err)
	}

	// 使用反射提取所有字段值
	fields := extractFieldValues(protoObj)

	return &ParsedNodeInfo{
		Protocol: protocol,
		Fields:   fields,
	}, nil
}

// extractFieldValues 使用反射提取结构体所有字段值
func extractFieldValues(v interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	extractFieldValuesRecursive(reflect.ValueOf(v), "", fields)
	return fields
}

// extractFieldValuesRecursive 递归提取字段值
func extractFieldValuesRecursive(v reflect.Value, prefix string, fields map[string]interface{}) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		fieldValue := v.Field(i)
		kind := fieldValue.Kind()

		switch kind {
		case reflect.String:
			fields[fieldName] = fieldValue.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fields[fieldName] = fieldValue.Int()
		case reflect.Bool:
			fields[fieldName] = fieldValue.Bool()
		case reflect.Interface:
			if !fieldValue.IsNil() {
				fields[fieldName] = fieldValue.Interface()
			}
		case reflect.Slice:
			// 处理切片类型（如 ALPN []string）
			if fieldValue.Len() > 0 {
				slice := make([]interface{}, fieldValue.Len())
				for j := 0; j < fieldValue.Len(); j++ {
					slice[j] = fieldValue.Index(j).Interface()
				}
				fields[fieldName] = slice
			}
		case reflect.Struct:
			// 递归处理嵌套结构体
			extractFieldValuesRecursive(fieldValue, fieldName, fields)
		}
	}
}

// UpdateNodeLinkFields 根据字段值更新节点链接
// 返回更新后的链接
func UpdateNodeLinkFields(link string, fieldsJSON string) (string, error) {
	if link == "" {
		return "", fmt.Errorf("链接不能为空")
	}

	var fields map[string]interface{}
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return "", fmt.Errorf("解析字段JSON失败: %w", err)
	}

	linkLower := strings.ToLower(link)

	switch {
	case strings.HasPrefix(linkLower, "vmess://"):
		return updateVmessFields(link, fields)
	case strings.HasPrefix(linkLower, "vless://"):
		return updateVlessFields(link, fields)
	case strings.HasPrefix(linkLower, "trojan://"):
		return updateTrojanFields(link, fields)
	case strings.HasPrefix(linkLower, "ss://"):
		return updateSSFields(link, fields)
	case strings.HasPrefix(linkLower, "ssr://"):
		return updateSSRFields(link, fields)
	case strings.HasPrefix(linkLower, "hysteria://"), strings.HasPrefix(linkLower, "hy://"):
		return updateHysteriaFields(link, fields)
	case strings.HasPrefix(linkLower, "hysteria2://"), strings.HasPrefix(linkLower, "hy2://"):
		return updateHysteria2Fields(link, fields)
	case strings.HasPrefix(linkLower, "tuic://"):
		return updateTuicFields(link, fields)
	case strings.HasPrefix(linkLower, "anytls://"):
		return updateAnyTLSFields(link, fields)
	case strings.HasPrefix(linkLower, "socks5://"):
		return updateSocks5Fields(link, fields)
	default:
		return "", fmt.Errorf("不支持的协议类型")
	}
}

// setFieldValue 使用反射设置字段值
func setFieldValue(v reflect.Value, fieldPath string, value interface{}) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	parts := strings.Split(fieldPath, ".")
	for i, part := range parts {
		if v.Kind() != reflect.Struct {
			return fmt.Errorf("字段路径无效: %s", fieldPath)
		}
		field := v.FieldByName(part)
		if !field.IsValid() {
			return fmt.Errorf("字段不存在: %s", part)
		}
		if !field.CanSet() {
			return fmt.Errorf("字段不可设置: %s", part)
		}
		if i == len(parts)-1 {
			// 最后一级，设置值
			return setValueByKind(field, value)
		}
		v = field
	}
	return nil
}

// setValueByKind 根据字段类型设置值
func setValueByKind(field reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		str, ok := value.(string)
		if !ok {
			str = fmt.Sprintf("%v", value)
		}
		field.SetString(str)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := value.(type) {
		case float64:
			field.SetInt(int64(v))
		case int:
			field.SetInt(int64(v))
		case int64:
			field.SetInt(v)
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				field.SetInt(i)
			}
		}
	case reflect.Bool:
		switch v := value.(type) {
		case bool:
			field.SetBool(v)
		case string:
			field.SetBool(v == "true" || v == "1")
		}
	case reflect.Interface:
		field.Set(reflect.ValueOf(value))
	case reflect.Slice:
		if arr, ok := value.([]interface{}); ok {
			if field.Type().Elem().Kind() == reflect.String {
				strSlice := make([]string, len(arr))
				for i, v := range arr {
					strSlice[i] = fmt.Sprintf("%v", v)
				}
				field.Set(reflect.ValueOf(strSlice))
			}
		}
	}
	return nil
}

// 以下是各协议的字段更新函数

func updateVmessFields(link string, fields map[string]interface{}) (string, error) {
	vmess, err := DecodeVMESSURL(link)
	if err != nil {
		return "", err
	}
	v := reflect.ValueOf(&vmess).Elem()
	for path, val := range fields {
		setFieldValue(v, path, val)
	}
	return EncodeVmessURL(vmess), nil
}

func updateVlessFields(link string, fields map[string]interface{}) (string, error) {
	vless, err := DecodeVLESSURL(link)
	if err != nil {
		return "", err
	}
	v := reflect.ValueOf(&vless).Elem()
	for path, val := range fields {
		setFieldValue(v, path, val)
	}
	return EncodeVLESSURL(vless), nil
}

func updateTrojanFields(link string, fields map[string]interface{}) (string, error) {
	trojan, err := DecodeTrojanURL(link)
	if err != nil {
		return "", err
	}
	v := reflect.ValueOf(&trojan).Elem()
	for path, val := range fields {
		setFieldValue(v, path, val)
	}
	return EncodeTrojanURL(trojan), nil
}

func updateSSFields(link string, fields map[string]interface{}) (string, error) {
	ss, err := DecodeSSURL(link)
	if err != nil {
		return "", err
	}
	v := reflect.ValueOf(&ss).Elem()
	for path, val := range fields {
		setFieldValue(v, path, val)
	}
	return EncodeSSURL(ss), nil
}

func updateSSRFields(link string, fields map[string]interface{}) (string, error) {
	ssr, err := DecodeSSRURL(link)
	if err != nil {
		return "", err
	}
	v := reflect.ValueOf(&ssr).Elem()
	for path, val := range fields {
		setFieldValue(v, path, val)
	}
	return EncodeSSRURL(ssr), nil
}

func updateHysteriaFields(link string, fields map[string]interface{}) (string, error) {
	hy, err := DecodeHYURL(link)
	if err != nil {
		return "", err
	}
	v := reflect.ValueOf(&hy).Elem()
	for path, val := range fields {
		setFieldValue(v, path, val)
	}
	return EncodeHYURL(hy), nil
}

func updateHysteria2Fields(link string, fields map[string]interface{}) (string, error) {
	hy2, err := DecodeHY2URL(link)
	if err != nil {
		return "", err
	}
	v := reflect.ValueOf(&hy2).Elem()
	for path, val := range fields {
		setFieldValue(v, path, val)
	}
	return EncodeHY2URL(hy2), nil
}

func updateTuicFields(link string, fields map[string]interface{}) (string, error) {
	tuic, err := DecodeTuicURL(link)
	if err != nil {
		return "", err
	}
	v := reflect.ValueOf(&tuic).Elem()
	for path, val := range fields {
		setFieldValue(v, path, val)
	}
	return EncodeTuicURL(tuic), nil
}

func updateAnyTLSFields(link string, fields map[string]interface{}) (string, error) {
	anytls, err := DecodeAnyTLSURL(link)
	if err != nil {
		return "", err
	}
	v := reflect.ValueOf(&anytls).Elem()
	for path, val := range fields {
		setFieldValue(v, path, val)
	}
	return EncodeAnyTLSURL(anytls), nil
}

func updateSocks5Fields(link string, fields map[string]interface{}) (string, error) {
	socks5, err := DecodeSocks5URL(link)
	if err != nil {
		return "", err
	}
	v := reflect.ValueOf(&socks5).Elem()
	for path, val := range fields {
		setFieldValue(v, path, val)
	}
	return EncodeSocks5URL(socks5), nil
}
