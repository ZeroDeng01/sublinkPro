package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"sublink/utils"
)

var (
	unlockStatusMetaResolver       = defaultUnlockStatusMetaResolver
	unlockProviderMetaResolver     = defaultUnlockProviderMetaResolver
	unlockRenameVariableResolver   = defaultUnlockRenameVariableResolver
	unlockStatusNormalizerResolver = defaultUnlockStatusNormalizer
)

const (
	UnlockProviderNetflix   = "netflix"
	UnlockProviderDisney    = "disney"
	UnlockProviderYouTube   = "youtube_premium"
	UnlockProviderOpenAI    = "openai"
	UnlockProviderGemini    = "gemini"
	UnlockProviderClaude    = "claude"
	UnlockProviderBahamut   = "bahamut"
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

func RegisterUnlockStatusMetaResolver(resolver func() []UnlockStatusMeta) {
	if resolver == nil {
		unlockStatusMetaResolver = defaultUnlockStatusMetaResolver
		return
	}
	unlockStatusMetaResolver = resolver
}

func RegisterUnlockProviderMetaResolver(resolver func(string) UnlockProviderMeta) {
	if resolver == nil {
		unlockProviderMetaResolver = defaultUnlockProviderMetaResolver
		return
	}
	unlockProviderMetaResolver = resolver
}

func RegisterUnlockRenameVariableResolver(resolver func([]string) []UnlockRenameVariableMeta) {
	if resolver == nil {
		unlockRenameVariableResolver = defaultUnlockRenameVariableResolver
		return
	}
	unlockRenameVariableResolver = resolver
}

func RegisterUnlockStatusNormalizer(normalizer func(string) string) {
	if normalizer == nil {
		unlockStatusNormalizerResolver = defaultUnlockStatusNormalizer
		return
	}
	unlockStatusNormalizerResolver = normalizer
}

func NormalizeUnlockStatus(status string) string {
	return unlockStatusNormalizerResolver(status)
}

func defaultUnlockStatusNormalizer(status string) string {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case "", "all":
		return ""
	case UnlockStatusUntested, UnlockStatusAvailable, UnlockStatusPartial, UnlockStatusRestricted,
		UnlockStatusReachable, UnlockStatusUnsupported, UnlockStatusUnknown, UnlockStatusError:
		return normalized
	default:
		return ""
	}
}

func NormalizeUnlockFilterRules(rules []UnlockFilterRule) []UnlockFilterRule {
	normalized := make([]UnlockFilterRule, 0, len(rules))
	for _, rule := range rules {
		provider := NormalizeUnlockProvider(rule.Provider)
		status := NormalizeUnlockStatus(rule.Status)
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
			status := NormalizeUnlockStatus(result.Status)
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
	return unlockStatusMetaResolver()
}

func GetUnlockProviderMeta(provider string) UnlockProviderMeta {
	return unlockProviderMetaResolver(provider)
}

func BuildUnlockRenameVariables(providers []string) []UnlockRenameVariableMeta {
	return unlockRenameVariableResolver(providers)
}

func defaultUnlockStatusMetaResolver() []UnlockStatusMeta {
	return []UnlockStatusMeta{
		{Value: UnlockStatusAvailable, Label: "Unlocked", ShortLabel: "Unlocked", Description: "Available normally", Color: "success", Severity: "success"},
		{Value: UnlockStatusPartial, Label: "Partial", ShortLabel: "Partial", Description: "Only some capabilities are available", Color: "warning", Severity: "warning"},
		{Value: UnlockStatusReachable, Label: "Reachable", ShortLabel: "Reachable", Description: "Reachable, but not necessarily fully available", Color: "info", Severity: "info"},
		{Value: UnlockStatusRestricted, Label: "Restricted", ShortLabel: "Restricted", Description: "The current region or exit is restricted", Color: "error", Severity: "error"},
		{Value: UnlockStatusUnsupported, Label: "Unsupported", ShortLabel: "Unsupported", Description: "The current region is outside the officially supported range", Color: "warning", Severity: "warning"},
		{Value: UnlockStatusUnknown, Label: "Unknown", ShortLabel: "Unknown", Description: "The result cannot be determined reliably", Color: "default", Severity: "default"},
		{Value: UnlockStatusError, Label: "Error", ShortLabel: "Error", Description: "This check encountered an error", Color: "error", Severity: "error"},
		{Value: UnlockStatusUntested, Label: "Untested", ShortLabel: "Untested", Description: "No check has been run yet", Color: "default", Severity: "default"},
	}
}

func defaultUnlockProviderMetaResolver(provider string) UnlockProviderMeta {
	switch NormalizeUnlockProvider(provider) {
	case UnlockProviderNetflix:
		return UnlockProviderMeta{Value: UnlockProviderNetflix, Label: "Netflix", Description: "Checks whether the full catalog or only Originals is supported", Category: "streaming"}
	case UnlockProviderDisney:
		return UnlockProviderMeta{Value: UnlockProviderDisney, Label: "Disney+", Description: "Checks whether the Disney+ entry point is reachable and clearly unrestricted", Category: "streaming"}
	case UnlockProviderYouTube:
		return UnlockProviderMeta{Value: UnlockProviderYouTube, Label: "YouTube Premium", Description: "Checks whether YouTube Premium is available in the exit region", Category: "streaming"}
	case UnlockProviderOpenAI:
		return UnlockProviderMeta{Value: UnlockProviderOpenAI, Label: "OpenAI", Description: "Checks OpenAI / ChatGPT regional availability", Category: "ai"}
	case UnlockProviderGemini:
		return UnlockProviderMeta{Value: UnlockProviderGemini, Label: "Gemini", Description: "Checks Google Gemini regional availability", Category: "ai"}
	case UnlockProviderClaude:
		return UnlockProviderMeta{Value: UnlockProviderClaude, Label: "Claude", Description: "Checks Anthropic Claude regional availability", Category: "ai"}
	case UnlockProviderBahamut:
		return UnlockProviderMeta{Value: UnlockProviderBahamut, Label: "Bahamut Anime", Description: "Checks Bahamut Anime (ani.gamer.com.tw) regional availability", Category: "streaming"}
	default:
		key := NormalizeUnlockProvider(provider)
		return UnlockProviderMeta{Value: key, Label: provider, Category: "custom"}
	}
}

func defaultUnlockRenameVariableResolver(providers []string) []UnlockRenameVariableMeta {
	normalizedProviders := NormalizeUnlockProviders(providers)
	variables := make([]UnlockRenameVariableMeta, 0, len(normalizedProviders))
	for _, provider := range normalizedProviders {
		meta := GetUnlockProviderMeta(provider)
		variables = append(variables, UnlockRenameVariableMeta{
			Key:         fmt.Sprintf("$Unlock(%s)", provider),
			Label:       meta.Label + " Unlock",
			Description: "Outputs a compact unlock summary for this service, e.g. “Unlocked-US” or “Restricted”",
			Provider:    provider,
		})
	}
	return variables
}

func GetUnlockStatusLabel(status string) string {
	normalizedStatus := NormalizeUnlockStatus(status)
	if normalizedStatus == "" {
		normalizedStatus = strings.TrimSpace(status)
	}
	for _, item := range GetUnlockStatusMetas() {
		if item.Value == normalizedStatus {
			if item.Label != "" {
				return item.Label
			}
			return item.ShortLabel
		}
	}
	return normalizedStatus
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
	status := NormalizeUnlockStatus(summary.Providers[0].Status)
	if status == "" {
		return strings.TrimSpace(summary.Providers[0].Status)
	}
	return status
}

func MatchUnlockSummary(summary UnlockSummary, provider string, status string, keyword string) bool {
	providerKey := NormalizeUnlockProvider(provider)
	statusKey := NormalizeUnlockStatus(status)
	if statusKey == "" {
		statusKey = strings.TrimSpace(status)
	}
	keywordLower := strings.ToLower(strings.TrimSpace(keyword))

	if len(summary.Providers) == 0 {
		return providerKey == "" && statusKey == "" && keywordLower == ""
	}

	for _, result := range summary.Providers {
		if providerKey != "" && NormalizeUnlockProvider(result.Provider) != providerKey {
			continue
		}
		resultStatus := NormalizeUnlockStatus(result.Status)
		if resultStatus == "" {
			resultStatus = strings.TrimSpace(result.Status)
		}
		if statusKey != "" && resultStatus != statusKey {
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
		return "Not checked"
	}
	if provider != "" {
		result, ok := GetUnlockResult(summary, provider)
		if !ok {
			return "Not checked"
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

func BuildNodeRenameInfo(node Node, processedLinkName string, protocol string, index int, duplicateIndex int) utils.NodeInfo {
	summary := ParseUnlockSummary(node.UnlockSummary)
	primaryUnlockStatus := "Not checked"
	primaryUnlockLabel := "Not checked"
	primaryUnlockRegion := ""
	if len(summary.Providers) > 0 {
		result := summary.Providers[0]
		primaryUnlockStatus = result.Status
		primaryUnlockLabel = GetUnlockStatusLabel(result.Status)
		primaryUnlockRegion = result.Region
	}

	// 根据节点的国家代码获取国家名称
	linkCountryName := ""
	if node.LinkCountry != "" {
		linkCountryName = GetCountryNameByCode(node.LinkCountry)
	}

	return utils.NodeInfo{
		Name:            node.EffectiveName(),
		LinkName:        processedLinkName,
		LinkCountry:     node.LinkCountry,
		LinkCountryName: linkCountryName,
		Speed:           node.Speed,
		SpeedStatus:     node.SpeedStatus,
		DelayTime:       node.DelayTime,
		DelayStatus:     node.DelayStatus,
		Group:           node.Group,
		Source:          node.Source,
		Index:           index,
		DuplicateIndex:  duplicateIndex,
		Protocol:        protocol,
		Tags:            node.Tags,
		IsBroadcast:     node.IsBroadcast,
		IsResidential:   node.IsResidential,
		FraudScore:      node.FraudScore,
		QualityStatus:   node.QualityStatus,
		QualityFamily:   node.QualityFamily,
		UnlockRaw:       node.UnlockSummary,
		UnlockSummary:   BuildUnlockRenameValue(node.UnlockSummary, ""),
		UnlockStatus:    primaryUnlockStatus,
		UnlockLabel:     primaryUnlockLabel,
		UnlockRegion:    primaryUnlockRegion,
	}
}
