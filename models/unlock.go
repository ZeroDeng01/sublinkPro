package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"sublink/utils"
)

const (
	UnlockProviderNetflix   = "netflix"
	UnlockProviderDisney    = "disney"
	UnlockProviderYouTube   = "youtube_premium"
	UnlockProviderOpenAI    = "openai"
	UnlockProviderGemini    = "gemini"
	UnlockProviderClaude    = "claude"
	UnlockStatusUntested    = "untested"
	UnlockStatusAvailable   = "available"
	UnlockStatusPartial     = "partial"
	UnlockStatusRestricted  = "restricted"
	UnlockStatusReachable   = "reachable"
	UnlockStatusUnsupported = "unsupported"
	UnlockStatusUnknown     = "unknown"
	UnlockStatusError       = "error"
)

type UnlockProviderResult struct {
	Provider string `json:"provider"`
	Status   string `json:"status"`
	Region   string `json:"region,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

type UnlockSummary struct {
	Providers []UnlockProviderResult `json:"providers"`
	UpdatedAt string                 `json:"updatedAt,omitempty"`
}

type UnlockAggregate struct {
	Enabled   bool                            `json:"enabled"`
	Providers []string                        `json:"providers,omitempty"`
	Counts    map[string]map[string]int       `json:"counts,omitempty"`
	Samples   map[string]UnlockProviderResult `json:"samples,omitempty"`
}

type UnlockProviderMeta struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

type UnlockRenameVariableMeta struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Provider    string `json:"provider,omitempty"`
}

type UnlockStatusMeta struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	ShortLabel  string `json:"shortLabel,omitempty"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Severity    string `json:"severity,omitempty"`
}

type UnlockFilterRule struct {
	Provider string `json:"provider,omitempty"`
	Status   string `json:"status,omitempty"`
	Keyword  string `json:"keyword,omitempty"`
}

const (
	UnlockRuleModeAny = "or"
	UnlockRuleModeAll = "and"
)

func NormalizeUnlockProvider(provider string) string {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return strings.Trim(normalized, "_")
}

func NormalizeUnlockProviders(providers []string) []string {
	seen := make(map[string]struct{}, len(providers))
	normalized := make([]string, 0, len(providers))
	for _, provider := range providers {
		key := NormalizeUnlockProvider(provider)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}
	return normalized
}

func NormalizeUnlockFilterRules(rules []UnlockFilterRule) []UnlockFilterRule {
	normalized := make([]UnlockFilterRule, 0, len(rules))
	for _, rule := range rules {
		provider := NormalizeUnlockProvider(rule.Provider)
		status := strings.ToLower(strings.TrimSpace(rule.Status))
		keyword := strings.TrimSpace(rule.Keyword)
		if provider == "" && status == "" && keyword == "" {
			continue
		}
		normalized = append(normalized, UnlockFilterRule{
			Provider: provider,
			Status:   status,
			Keyword:  keyword,
		})
	}
	return normalized
}

func BuildUnlockFilterRulesJSON(rules []UnlockFilterRule) string {
	normalized := NormalizeUnlockFilterRules(rules)
	if len(normalized) == 0 {
		return ""
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return ""
	}
	return string(payload)
}

func ParseUnlockFilterRules(raw string) []UnlockFilterRule {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var rules []UnlockFilterRule
	if err := json.Unmarshal([]byte(raw), &rules); err != nil {
		return nil
	}
	return NormalizeUnlockFilterRules(rules)
}

func BuildUnlockSummaryJSON(summary UnlockSummary) string {
	if len(summary.Providers) == 0 {
		return ""
	}
	payload, err := json.Marshal(summary)
	if err != nil {
		return ""
	}
	return string(payload)
}

func ParseUnlockSummary(raw string) UnlockSummary {
	if strings.TrimSpace(raw) == "" {
		return UnlockSummary{}
	}
	var summary UnlockSummary
	if err := json.Unmarshal([]byte(raw), &summary); err != nil {
		return UnlockSummary{}
	}
	return summary
}

func BuildUnlockAggregate(summaries []UnlockSummary, providers []string) UnlockAggregate {
	agg := UnlockAggregate{
		Enabled:   len(summaries) > 0,
		Providers: NormalizeUnlockProviders(providers),
		Counts:    make(map[string]map[string]int),
		Samples:   make(map[string]UnlockProviderResult),
	}
	for _, summary := range summaries {
		for _, result := range summary.Providers {
			provider := NormalizeUnlockProvider(result.Provider)
			if provider == "" {
				continue
			}
			if _, exists := agg.Counts[provider]; !exists {
				agg.Counts[provider] = make(map[string]int)
			}
			status := strings.TrimSpace(result.Status)
			if status == "" {
				status = UnlockStatusUnknown
			}
			agg.Counts[provider][status]++
			if _, exists := agg.Samples[provider]; !exists {
				agg.Samples[provider] = result
			}
		}
	}
	if len(agg.Counts) == 0 {
		agg.Enabled = false
	}
	return agg
}

func GetUnlockStatusMetas() []UnlockStatusMeta {
	return []UnlockStatusMeta{
		{Value: UnlockStatusAvailable, Label: "解锁", ShortLabel: "解锁", Description: "可正常使用", Color: "success", Severity: "success"},
		{Value: UnlockStatusPartial, Label: "部分", ShortLabel: "部分", Description: "仅部分能力可用", Color: "warning", Severity: "warning"},
		{Value: UnlockStatusReachable, Label: "直连", ShortLabel: "直连", Description: "可以访问，但不代表完整能力可用", Color: "info", Severity: "info"},
		{Value: UnlockStatusRestricted, Label: "受限", ShortLabel: "受限", Description: "当前地区或出口被限制", Color: "error", Severity: "error"},
		{Value: UnlockStatusUnsupported, Label: "不支持", ShortLabel: "不支持", Description: "当前地区不在官方支持范围内", Color: "warning", Severity: "warning"},
		{Value: UnlockStatusUnknown, Label: "未知", ShortLabel: "未知", Description: "当前无法稳定判断结果", Color: "default", Severity: "default"},
		{Value: UnlockStatusError, Label: "异常", ShortLabel: "异常", Description: "本轮检测出现错误", Color: "error", Severity: "error"},
		{Value: UnlockStatusUntested, Label: "未测", ShortLabel: "未测", Description: "尚未执行检测", Color: "default", Severity: "default"},
	}
}

func GetUnlockProviderMeta(provider string) UnlockProviderMeta {
	switch NormalizeUnlockProvider(provider) {
	case UnlockProviderNetflix:
		return UnlockProviderMeta{Value: UnlockProviderNetflix, Label: "Netflix", Description: "检测是否支持完整区服或仅 Originals", Category: "streaming"}
	case UnlockProviderDisney:
		return UnlockProviderMeta{Value: UnlockProviderDisney, Label: "Disney+", Description: "检测 Disney+ 服务入口是否可访问及是否明显受限", Category: "streaming"}
	case UnlockProviderYouTube:
		return UnlockProviderMeta{Value: UnlockProviderYouTube, Label: "YouTube Premium", Description: "检测 YouTube Premium 是否属于支持地区", Category: "streaming"}
	case UnlockProviderOpenAI:
		return UnlockProviderMeta{Value: UnlockProviderOpenAI, Label: "OpenAI", Description: "检测 OpenAI / ChatGPT 服务地区可访问性", Category: "ai"}
	case UnlockProviderGemini:
		return UnlockProviderMeta{Value: UnlockProviderGemini, Label: "Gemini", Description: "检测 Google Gemini 服务地区可访问性", Category: "ai"}
	case UnlockProviderClaude:
		return UnlockProviderMeta{Value: UnlockProviderClaude, Label: "Claude", Description: "检测 Anthropic Claude 服务地区可访问性", Category: "ai"}
	default:
		key := NormalizeUnlockProvider(provider)
		return UnlockProviderMeta{Value: key, Label: provider, Category: "custom"}
	}
}

func BuildUnlockRenameVariables(providers []string) []UnlockRenameVariableMeta {
	normalizedProviders := NormalizeUnlockProviders(providers)
	variables := make([]UnlockRenameVariableMeta, 0, len(normalizedProviders))
	for _, provider := range normalizedProviders {
		meta := GetUnlockProviderMeta(provider)
		variables = append(variables, UnlockRenameVariableMeta{
			Key:         fmt.Sprintf("$Unlock(%s)", provider),
			Label:       meta.Label + " 解锁",
			Description: "输出该服务的紧凑解锁结果摘要，例如“解锁-US”或“受限”",
			Provider:    provider,
		})
	}
	return variables
}

func GetUnlockStatusLabel(status string) string {
	for _, item := range GetUnlockStatusMetas() {
		if item.Value == status {
			if item.Label != "" {
				return item.Label
			}
			return item.ShortLabel
		}
	}
	return status
}

func GetUnlockResult(summary UnlockSummary, provider string) (UnlockProviderResult, bool) {
	target := NormalizeUnlockProvider(provider)
	for _, result := range summary.Providers {
		if target == "" || NormalizeUnlockProvider(result.Provider) == target {
			return result, true
		}
	}
	return UnlockProviderResult{}, false
}

func BuildUnlockSearchText(raw string) string {
	summary := ParseUnlockSummary(raw)
	if len(summary.Providers) == 0 {
		return ""
	}
	parts := make([]string, 0, len(summary.Providers)*4)
	for _, result := range summary.Providers {
		providerMeta := GetUnlockProviderMeta(result.Provider)
		parts = append(parts,
			providerMeta.Label,
			result.Provider,
			GetUnlockStatusLabel(result.Status),
			result.Region,
			result.Reason,
			result.Detail,
		)
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func BuildPrimaryUnlockStatus(raw string) string {
	summary := ParseUnlockSummary(raw)
	if len(summary.Providers) == 0 {
		return UnlockStatusUntested
	}
	return strings.TrimSpace(summary.Providers[0].Status)
}

func MatchUnlockSummary(summary UnlockSummary, provider string, status string, keyword string) bool {
	providerKey := NormalizeUnlockProvider(provider)
	statusKey := strings.TrimSpace(status)
	keywordLower := strings.ToLower(strings.TrimSpace(keyword))

	if len(summary.Providers) == 0 {
		return providerKey == "" && statusKey == "" && keywordLower == ""
	}

	for _, result := range summary.Providers {
		if providerKey != "" && NormalizeUnlockProvider(result.Provider) != providerKey {
			continue
		}
		if statusKey != "" && strings.TrimSpace(result.Status) != statusKey {
			continue
		}
		if keywordLower != "" {
			providerMeta := GetUnlockProviderMeta(result.Provider)
			joined := strings.ToLower(strings.Join([]string{
				providerMeta.Label,
				result.Provider,
				result.Status,
				GetUnlockStatusLabel(result.Status),
				result.Region,
				result.Reason,
				result.Detail,
			}, " "))
			if !strings.Contains(joined, keywordLower) {
				continue
			}
		}
		return true
	}

	return false
}

func MatchUnlockSummaryRules(summary UnlockSummary, rules []UnlockFilterRule) bool {
	return MatchUnlockSummaryRulesWithMode(summary, rules, UnlockRuleModeAny)
}

func NormalizeUnlockRuleMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case UnlockRuleModeAll:
		return UnlockRuleModeAll
	default:
		return UnlockRuleModeAny
	}
}

func MatchUnlockSummaryRulesWithMode(summary UnlockSummary, rules []UnlockFilterRule, mode string) bool {
	normalized := NormalizeUnlockFilterRules(rules)
	if len(normalized) == 0 {
		return true
	}
	mode = NormalizeUnlockRuleMode(mode)
	if mode == UnlockRuleModeAll {
		for _, rule := range normalized {
			if !MatchUnlockSummary(summary, rule.Provider, rule.Status, rule.Keyword) {
				return false
			}
		}
		return true
	}
	for _, rule := range normalized {
		if MatchUnlockSummary(summary, rule.Provider, rule.Status, rule.Keyword) {
			return true
		}
	}
	return false
}

func BuildUnlockRenameValue(raw string, provider string) string {
	summary := ParseUnlockSummary(raw)
	if len(summary.Providers) == 0 {
		return "未检测"
	}
	if provider != "" {
		result, ok := GetUnlockResult(summary, provider)
		if !ok {
			return "未检测"
		}
		parts := []string{GetUnlockStatusLabel(result.Status)}
		if result.Region != "" {
			parts = append(parts, result.Region)
		}
		return strings.Join(parts, "-")
	}
	primary := summary.Providers[0]
	providerMeta := GetUnlockProviderMeta(primary.Provider)
	parts := []string{providerMeta.Label, GetUnlockStatusLabel(primary.Status)}
	if primary.Region != "" {
		parts = append(parts, primary.Region)
	}
	if len(summary.Providers) > 1 {
		parts = append(parts, fmt.Sprintf("+%d", len(summary.Providers)-1))
	}
	return strings.Join(parts, "-")
}

func BuildNodeRenameInfo(node Node, processedLinkName string, protocol string, index int) utils.NodeInfo {
	summary := ParseUnlockSummary(node.UnlockSummary)
	primaryUnlockStatus := "未检测"
	primaryUnlockLabel := "未检测"
	primaryUnlockRegion := ""
	if len(summary.Providers) > 0 {
		result := summary.Providers[0]
		primaryUnlockStatus = result.Status
		primaryUnlockLabel = GetUnlockStatusLabel(result.Status)
		primaryUnlockRegion = result.Region
	}

	return utils.NodeInfo{
		Name:          node.Name,
		LinkName:      processedLinkName,
		LinkCountry:   node.LinkCountry,
		Speed:         node.Speed,
		SpeedStatus:   node.SpeedStatus,
		DelayTime:     node.DelayTime,
		DelayStatus:   node.DelayStatus,
		Group:         node.Group,
		Source:        node.Source,
		Index:         index,
		Protocol:      protocol,
		Tags:          node.Tags,
		IsBroadcast:   node.IsBroadcast,
		IsResidential: node.IsResidential,
		FraudScore:    node.FraudScore,
		QualityStatus: node.QualityStatus,
		QualityFamily: node.QualityFamily,
		UnlockRaw:     node.UnlockSummary,
		UnlockSummary: BuildUnlockRenameValue(node.UnlockSummary, ""),
		UnlockStatus:  primaryUnlockStatus,
		UnlockLabel:   primaryUnlockLabel,
		UnlockRegion:  primaryUnlockRegion,
	}
}
