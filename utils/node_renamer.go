package utils

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// NodeInfo 节点信息结构体，用于重命名
type NodeInfo struct {
	Name        string  // 系统节点备注名称
	LinkName    string  // 节点原始名称（来自订阅源）
	LinkCountry string  // 落地IP国家代码
	Speed       float64 // 速度 (MB/s)
	DelayTime   int     // 延迟 (ms)
	Group       string  // 分组
	Source      string  // 来源（手动添加/订阅名称）
	Index       int     // 序号 (从1开始)
	Protocol    string  // 协议类型
}

// ISOToFlag 将国家ISO代码转换为国旗emoji
// isoCode: 两位ISO国家代码 (如 "CN", "US", "HK")
// TW会转换为中国国旗，未知/无效代码返回白旗 🏳️
func ISOToFlag(isoCode string) string {
	if isoCode == "" || len(isoCode) != 2 {
		return "🏳️" // 未知国旗使用白旗
	}

	code := strings.ToUpper(isoCode)

	// TW使用中国国旗
	if code == "TW" {
		code = "CN"
	}

	// 检查是否为有效的字母代码
	for _, c := range code {
		if c < 'A' || c > 'Z' {
			return "🏳️"
		}
	}

	// 将字母转换为区域指示符号 (Regional Indicator Symbol)
	// 'A' 对应 U+1F1E6
	flag := ""
	for _, c := range code {
		flag += string(rune(0x1F1E6 + int(c) - 'A'))
	}

	return flag
}

// RenameNode 根据规则重命名节点
// rule: 命名规则，如 "$LinkCountry - $Name ($Speed)"
// info: 节点信息
// 返回重命名后的名称，如果rule为空则返回原始名称
func RenameNode(rule string, info NodeInfo) string {
	if rule == "" {
		return info.Name
	}

	result := rule

	// 如果国家代码为空，使用"未知"
	linkCountry := info.LinkCountry
	if linkCountry == "" {
		linkCountry = "未知"
	}

	// 替换所有支持的变量
	replacements := map[string]string{
		"$Name":        info.Name,
		"$LinkName":    info.LinkName,
		"$LinkCountry": linkCountry,
		"$Flag":        ISOToFlag(info.LinkCountry),
		"$Speed":       FormatSpeed(info.Speed),
		"$Delay":       FormatDelay(info.DelayTime),
		"$Group":       info.Group,
		"$Source":      info.Source,
		"$Index":       fmt.Sprintf("%d", info.Index),
		"$Protocol":    info.Protocol,
	}

	for variable, value := range replacements {
		result = strings.ReplaceAll(result, variable, value)
	}

	// 清理连续空格和首尾空格
	result = strings.TrimSpace(result)

	// 如果结果为空，返回原始名称
	if result == "" {
		return info.Name
	}

	return result
}

// FormatSpeed 格式化速度显示
// speed: 速度值 (MB/s)
// 返回格式化字符串，如 "1.50MB/s" 或 "N/A"
func FormatSpeed(speed float64) string {
	if speed <= 0 {
		return "N/A"
	}
	return fmt.Sprintf("%.2fMB/s", speed)
}

// FormatDelay 格式化延迟显示
// delay: 延迟值 (ms)
// 返回格式化字符串，如 "100ms" 或 "N/A"
func FormatDelay(delay int) string {
	if delay <= 0 {
		return "N/A"
	}
	return fmt.Sprintf("%dms", delay)
}

// GetProtocolFromLink 从节点链接解析协议类型
// link: 节点链接
// 返回协议名称，如 "VMess", "VLESS", "Trojan" 等
func GetProtocolFromLink(link string) string {
	if link == "" {
		return "未知"
	}

	// 常见协议前缀映射
	protocolPrefixes := map[string]string{
		"ss://":        "SS",
		"ssr://":       "SSR",
		"vmess://":     "VMess",
		"vless://":     "VLESS",
		"trojan://":    "Trojan",
		"hysteria://":  "Hysteria",
		"hysteria2://": "Hysteria2",
		"hy2://":       "Hysteria2",
		"tuic://":      "TUIC",
		"wg://":        "WireGuard",
		"wireguard://": "WireGuard",
		"naive://":     "NaiveProxy",
		"anytls://":    "AnyTLS",
		"socks5://":    "SOCKS5",
	}

	linkLower := strings.ToLower(link)
	for prefix, name := range protocolPrefixes {
		if strings.HasPrefix(linkLower, prefix) {
			return name
		}
	}

	return "其他"
}

// RenameNodeLink 重命名节点链接
// link: 原始节点链接
// newName: 新名称
// 返回重命名后的链接
func RenameNodeLink(link string, newName string) string {
	if link == "" || newName == "" {
		return link
	}

	// 获取协议scheme
	idx := strings.Index(link, "://")
	if idx == -1 {
		return link
	}
	scheme := strings.ToLower(link[:idx])

	switch scheme {
	case "vmess":
		return renameVmessLink(link, newName)
	case "vless", "trojan", "hy2", "hysteria2", "hysteria", "tuic", "anytls", "socks5":
		return renameFragmentLink(link, newName)
	case "ss":
		return renameSSLink(link, newName)
	case "ssr":
		return renameSSRLink(link, newName)
	default:
		// 尝试使用Fragment方式
		return renameFragmentLink(link, newName)
	}
}

// renameVmessLink VMess协议重命名 (base64 JSON)
func renameVmessLink(link string, newName string) string {
	if !strings.HasPrefix(link, "vmess://") {
		return link
	}

	encoded := strings.TrimPrefix(link, "vmess://")
	decoded := Base64Decode(strings.TrimSpace(encoded))
	if decoded == "" {
		return link
	}

	var vmess map[string]interface{}
	if err := json.Unmarshal([]byte(decoded), &vmess); err != nil {
		return link
	}

	vmess["ps"] = newName

	newJSON, err := json.Marshal(vmess)
	if err != nil {
		return link
	}

	return "vmess://" + Base64Encode(string(newJSON))
}

// renameFragmentLink 使用URL Fragment的协议重命名 (vless, trojan, hy2, tuic等)
func renameFragmentLink(link string, newName string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}
	u.Fragment = newName
	return u.String()
}

// renameSSLink SS协议重命名
func renameSSLink(link string, newName string) string {
	if !strings.HasPrefix(link, "ss://") {
		return link
	}

	// SS链接可能有多种格式:
	// 1. ss://base64(method:password)@host:port#name (SIP002)
	// 2. ss://base64(全部内容)

	u, err := url.Parse(link)
	if err != nil {
		// 尝试解析纯base64格式
		encoded := strings.TrimPrefix(link, "ss://")
		// 分离 #name 部分
		hashIdx := strings.LastIndex(encoded, "#")
		if hashIdx != -1 {
			encoded = encoded[:hashIdx]
		}
		return "ss://" + encoded + "#" + url.PathEscape(newName)
	}
	u.Fragment = newName
	return u.String()
}

// renameSSRLink SSR协议重命名 (需要解码base64)
func renameSSRLink(link string, newName string) string {
	if !strings.HasPrefix(link, "ssr://") {
		return link
	}

	encoded := strings.TrimPrefix(link, "ssr://")
	decoded := Base64Decode(encoded)
	if decoded == "" {
		return link
	}

	// SSR格式: host:port:protocol:method:obfs:base64(password)/?params
	// remarks=base64(name)
	if strings.Contains(decoded, "remarks=") {
		// 替换remarks参数
		parts := strings.Split(decoded, "remarks=")
		if len(parts) >= 2 {
			// 找到remarks的结束位置（下一个&或字符串结束）
			endIdx := strings.Index(parts[1], "&")
			var suffix string
			if endIdx != -1 {
				suffix = parts[1][endIdx:]
			} else {
				suffix = ""
			}
			decoded = parts[0] + "remarks=" + Base64Encode(newName) + suffix
		}
	} else if strings.Contains(decoded, "/?") {
		// 有参数但没有remarks，添加remarks
		decoded = decoded + "&remarks=" + Base64Encode(newName)
	} else {
		// 没有参数，添加参数
		decoded = decoded + "/?remarks=" + Base64Encode(newName)
	}

	return "ssr://" + Base64Encode(decoded)
}
