package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"sublink/constants"
)

// TagGroupTagsFn 标签组标签查询函数类型
// 参数：标签组名称
// 返回：该标签组下所有标签名称列表
type TagGroupTagsFn func(groupName string) []string

// tagGroupTagsFunc 标签组标签查询回调（由外部设置，避免循环依赖）
var tagGroupTagsFunc TagGroupTagsFn

// SetTagGroupTagsFunc 设置标签组标签查询函数（应在应用初始化时调用）
func SetTagGroupTagsFunc(fn TagGroupTagsFn) {
	tagGroupTagsFunc = fn
}

// tagGroupRegex 匹配 $TagGroup(xxx) 格式的正则
var tagGroupRegex = regexp.MustCompile(`\$TagGroup\(([^)]+)\)`)
var unlockRegex = regexp.MustCompile(`\$Unlock\(([^)]+)\)`)

var protocolLabelFromLinkFunc func(string) string
var renameNodeLinkFunc func(string, string) string

func SetProtocolLinkFuncs(labelFunc func(string) string, renameFunc func(string, string) string) {
	protocolLabelFromLinkFunc = labelFunc
	renameNodeLinkFunc = renameFunc
}

type renameUnlockSummary struct {
	Providers []renameUnlockProvider `json:"providers"`
}

type renameUnlockProvider struct {
	Provider string `json:"provider"`
	Status   string `json:"status"`
	Region   string `json:"region,omitempty"`
}

// replaceTagGroupVariables 替换规则中的 $TagGroup(xxx) 变量
// rule: 包含 $TagGroup(xxx) 变量的规则字符串
// nodeTags: 节点的标签列表（逗号分隔）
// 返回替换后的字符串
func replaceTagGroupVariables(rule string, nodeTags string) string {
	if tagGroupTagsFunc == nil {
		// 回调未设置，移除所有 $TagGroup(xxx) 变量
		return tagGroupRegex.ReplaceAllString(rule, "")
	}

	// 将节点标签转为集合便于快速查找
	nodeTagSet := make(map[string]bool)
	if nodeTags != "" {
		for _, t := range strings.Split(nodeTags, ",") {
			nodeTagSet[strings.TrimSpace(t)] = true
		}
	}

	return tagGroupRegex.ReplaceAllStringFunc(rule, func(match string) string {
		// 提取标签组名称
		submatches := tagGroupRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return ""
		}
		groupName := submatches[1]

		// 获取该标签组下的所有标签
		groupTags := tagGroupTagsFunc(groupName)
		if len(groupTags) == 0 {
			return ""
		}

		// 查找节点拥有的该组中的标签（由于互斥，最多一个）
		for _, tag := range groupTags {
			if nodeTagSet[tag] {
				return tag
			}
		}
		return ""
	})
}

// NodeInfo 节点信息结构体，用于重命名
type NodeInfo struct {
	Name          string  // 系统节点备注名称
	LinkName      string  // 节点原始名称（来自订阅源）
	LinkCountry   string  // 落地IP国家代码
	Speed         float64 // 速度 (MB/s)
	SpeedStatus   string  // 速度状态: untested/success/timeout/error
	DelayTime     int     // 延迟 (ms)
	DelayStatus   string  // 延迟状态: untested/success/timeout/error
	Group         string  // 分组
	Source        string  // 来源（手动添加/订阅名称）
	Index         int     // 序号 (从1开始)
	Protocol      string  // 协议类型
	Tags          string  // 节点标签（逗号分隔）
	IsBroadcast   bool    // IP来源：true=广播 false=原生
	IsResidential bool    // 是否住宅IP
	FraudScore    int     // 欺诈评分（0-100，-1=未检测）
	QualityStatus string
	QualityFamily string
	UnlockRaw     string
	UnlockSummary string
	UnlockStatus  string
	UnlockLabel   string
	UnlockRegion  string
}

const (
	qualityStatusUntested = "untested"
	qualityStatusSuccess  = "success"
	qualityStatusPartial  = "partial"
	qualityStatusFailed   = "failed"
	qualityStatusDisabled = "disabled"
	qualityFamilyIPv6     = "ipv6"
)

// FormatFraudScoreIcon 根据欺诈评分返回对应图标
// fraudScore: 欺诈评分（0-100，-1=未检测）
func FormatFraudScoreIcon(fraudScore int, qualityStatus string) string {
	if qualityStatus == qualityStatusPartial {
		return "ℹ️"
	}
	if qualityStatus != "" && qualityStatus != qualityStatusSuccess {
		return "⛔️"
	}
	if fraudScore < 0 {
		return "⛔️"
	}
	switch {
	case fraudScore <= 10:
		return "⚪"
	case fraudScore <= 30:
		return "🟢"
	case fraudScore <= 50:
		return "🟡"
	case fraudScore <= 70:
		return "🟠"
	case fraudScore <= 89:
		return "🔴"
	default:
		return "⚫"
	}
}

// FormatSpeedIcon 根据速度和测速状态返回对应图标
func FormatSpeedIcon(speed float64, speedStatus string) string {
	switch speedStatus {
	case constants.StatusTimeout:
		return "⏱️"
	case constants.StatusError:
		return "❌"
	case constants.StatusSuccess:
		if speed >= 5 {
			return "🟢"
		}
		if speed >= 1 {
			return "🟡"
		}
		return "🔴"
	case constants.StatusUntested:
		return "⛔️"
	}

	// 兼容历史数据：状态缺失时根据数值兜底判断
	if speed == -1 {
		return "❌"
	}
	if speed > 0 {
		if speed >= 5 {
			return "🟢"
		}
		if speed >= 1 {
			return "🟡"
		}
		return "🔴"
	}
	return "⛔️"
}

// FormatDelayIcon 根据延迟和测速状态返回对应图标
func FormatDelayIcon(delay int, delayStatus string) string {
	switch delayStatus {
	case constants.StatusTimeout:
		return "⏱️"
	case constants.StatusError:
		return "❌"
	case constants.StatusSuccess:
		if delay < 200 {
			return "🟢"
		}
		if delay < 500 {
			return "🟡"
		}
		return "🔴"
	case constants.StatusUntested:
		return "⛔️"
	}

	// 兼容历史数据：状态缺失时根据数值兜底判断
	if delay == -1 {
		return "⏱️"
	}
	if delay > 0 {
		if delay < 200 {
			return "🟢"
		}
		if delay < 500 {
			return "🟡"
		}
		return "🔴"
	}
	return "⛔️"
}

// PreprocessRule 原名预处理规则结构体
type PreprocessRule struct {
	MatchMode   string `json:"matchMode"`   // 匹配模式: "text" 或 "regex"
	Pattern     string `json:"pattern"`     // 匹配模式字符串
	Replacement string `json:"replacement"` // 替换内容
	Enabled     bool   `json:"enabled"`     // 是否启用
}

// PreprocessNodeName 应用预处理规则处理节点原名
// rulesJSON: JSON格式的预处理规则数组
// linkName: 原始节点名称
// 返回处理后的名称
func PreprocessNodeName(rulesJSON string, linkName string) string {
	if rulesJSON == "" || linkName == "" {
		return linkName
	}

	var rules []PreprocessRule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		return linkName
	}

	result := linkName
	for _, rule := range rules {
		if !rule.Enabled || rule.Pattern == "" {
			continue
		}

		if rule.MatchMode == "regex" {
			// 正则表达式匹配
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				continue // 跳过无效的正则表达式
			}
			result = re.ReplaceAllString(result, rule.Replacement)
		} else {
			// 纯文本匹配 (默认)
			result = strings.ReplaceAll(result, rule.Pattern, rule.Replacement)
		}
	}

	return result
}

// NodeNameFilterRule 节点名称过滤规则结构体
type NodeNameFilterRule struct {
	MatchMode string `json:"matchMode"` // 匹配模式: "text" 或 "regex"
	Pattern   string `json:"pattern"`   // 匹配模式字符串
	Enabled   bool   `json:"enabled"`   // 是否启用
}

// MatchesNodeNameFilter 检查节点名称是否匹配任意过滤规则
// rulesJSON: JSON格式的过滤规则数组
// nodeName: 节点名称
// 返回 true 如果匹配任意一条启用的规则
func MatchesNodeNameFilter(rulesJSON string, nodeName string) bool {
	if rulesJSON == "" || nodeName == "" {
		return false
	}

	var rules []NodeNameFilterRule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		return false
	}

	for _, rule := range rules {
		if !rule.Enabled || rule.Pattern == "" {
			continue
		}

		if rule.MatchMode == "regex" {
			// 正则表达式匹配
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				continue // 跳过无效的正则表达式
			}
			if re.MatchString(nodeName) {
				return true
			}
		} else {
			// 纯文本匹配 (默认) - 检查是否包含关键字
			if strings.Contains(nodeName, rule.Pattern) {
				return true
			}
		}
	}

	return false
}

// HasActiveNodeNameFilter 检查规则JSON是否包含至少一条有效的启用规则
// rulesJSON: JSON格式的过滤规则数组
// 返回 true 如果存在至少一条启用且有pattern的规则
func HasActiveNodeNameFilter(rulesJSON string) bool {
	if rulesJSON == "" || rulesJSON == "[]" {
		return false
	}

	var rules []NodeNameFilterRule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		return false
	}

	for _, rule := range rules {
		if rule.Enabled && rule.Pattern != "" {
			return true
		}
	}

	return false
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
	// 如果来源为manual则替换为手动
	linkSource := info.Source
	if linkSource == "manual" {
		linkSource = "手动"
	}

	// 如果分组为空 则返回未分组
	linkGroup := info.Group
	if linkGroup == "" {
		linkGroup = "未分组"
	}

	// 处理标签
	tags := info.Tags
	if tags != "" {
		// 将逗号分隔转换为竖线分隔
		tags = strings.ReplaceAll(tags, ",", "|")
	}

	// 替换所有支持的变量
	// 使用有序切片代替 map，确保长变量名优先替换
	// 这避免了如 $Tag 先于 $Tags 替换导致的问题
	type replacement struct {
		variable string
		value    string
	}
	// 按变量名长度降序排列，长的变量名优先替换
	replacements := []replacement{
		{"$LinkCountry", linkCountry},
		{"$Residential", func() string {
			if info.QualityStatus != qualityStatusSuccess {
				return formatQualityText(info)
			}
			if info.IsResidential {
				return "住宅IP"
			}
			return "机房IP"
		}()},
		{"$SpeedIcon", FormatSpeedIcon(info.Speed, info.SpeedStatus)},
		{"$DelayIcon", FormatDelayIcon(info.DelayTime, info.DelayStatus)},
		{"$FraudScoreIcon", FormatFraudScoreIcon(info.FraudScore, info.QualityStatus)},
		{"$FraudScore", func() string {
			if info.QualityStatus != qualityStatusSuccess {
				return formatQualityText(info)
			}
			return fmt.Sprintf("%d", info.FraudScore)
		}()},
		{"$LinkName", info.LinkName},
		{"$Protocol", info.Protocol},
		{"$UnlockStatus", info.UnlockStatus},
		{"$UnlockLabel", info.UnlockLabel},
		{"$UnlockRegion", info.UnlockRegion},
		{"$Unlock", info.UnlockSummary},
		{"$IpType", func() string {
			if info.QualityStatus != qualityStatusSuccess {
				return formatQualityText(info)
			}
			if info.IsBroadcast {
				return "广播IP"
			}
			return "原生IP"
		}()},
		{"$Source", linkSource},
		{"$Speed", FormatSpeed(info.Speed)},
		{"$Delay", FormatDelay(info.DelayTime)},
		{"$Group", linkGroup},
		{"$Index", fmt.Sprintf("%d", info.Index)},
		{"$Name", info.Name},
		{"$Flag", ISOToFlag(info.LinkCountry)},
		{"$Tags", tags}, // 所有标签（竖线｜分隔）
	}

	result = unlockRegex.ReplaceAllStringFunc(result, func(match string) string {
		submatches := unlockRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return ""
		}
		provider := strings.TrimSpace(submatches[1])
		if provider == "" {
			return ""
		}
		return buildProviderUnlockValue(info.UnlockRaw, provider)
	})

	for _, r := range replacements {
		result = strings.ReplaceAll(result, r.variable, r.value)
	}

	// 处理 $TagGroup(xxx) 变量 - 查找节点在指定标签组中的标签
	result = replaceTagGroupVariables(result, info.Tags)

	// 清理连续空格和首尾空格
	result = strings.TrimSpace(result)

	// 如果结果为空，返回原始名称
	if result == "" {
		return info.Name
	}

	return result
}

func buildProviderUnlockValue(raw string, provider string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "未检测"
	}
	var summary renameUnlockSummary
	if err := json.Unmarshal([]byte(trimmed), &summary); err != nil {
		return "未检测"
	}
	provider = normalizeUnlockProviderKey(provider)
	for _, item := range summary.Providers {
		if normalizeUnlockProviderKey(item.Provider) != provider {
			continue
		}
		parts := []string{formatUnlockStatusLabel(item.Status)}
		if item.Region != "" {
			parts = append(parts, item.Region)
		}
		return strings.Join(parts, "-")
	}
	return "未检测"
}

func normalizeUnlockProviderKey(provider string) string {
	key := strings.ToLower(strings.TrimSpace(provider))
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "-", "_")
	return strings.Trim(key, "_")
}

func formatUnlockStatusLabel(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "available":
		return "解锁"
	case "partial":
		return "部分"
	case "reachable":
		return "直连"
	case "restricted":
		return "受限"
	case "unsupported":
		return "不支持"
	case "error":
		return "异常"
	case "unknown":
		return "未知"
	default:
		return "未测"
	}
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

// GetProtocolFromLink 从节点链接解析协议类型（废弃：请使用 protocol.GetProtocolFromLink）
// 此函数保留用于向后兼容，返回显示名称格式（如 "VMess", "VLESS"）
// 新代码应直接使用 protocol.GetProtocolFromLink() 或 protocol.GetProtocolLabel()
// Deprecated: Use protocol.GetProtocolFromLink instead
func GetProtocolFromLink(link string) string {
	if link == "" {
		return "未知"
	}
	if protocolLabelFromLinkFunc != nil {
		label := protocolLabelFromLinkFunc(link)
		if label != "" {
			return label
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
	if renameNodeLinkFunc != nil {
		renamed := renameNodeLinkFunc(link, newName)
		if renamed != "" {
			return renamed
		}
	}
	return link
}

func formatQualityText(info NodeInfo) string {
	switch info.QualityStatus {
	case qualityStatusPartial:
		if info.QualityFamily == qualityFamilyIPv6 {
			return "IPv6部分结果"
		}
		return "部分结果"
	case qualityStatusFailed:
		return "检测失败"
	case qualityStatusDisabled:
		return "未启用"
	case qualityStatusUntested, "":
		return "未检测"
	default:
		return "未检测"
	}
}
