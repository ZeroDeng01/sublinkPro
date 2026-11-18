package utils

import "strings"

// Node naming output language. Node names are generated on the server and
// rendered inside client apps, so they cannot rely on the web UI translation
// layer and need their own catalog.
const (
	NodeNameLanguageEnglish = "en-US"
	NodeNameLanguageChinese = "zh-CN"
)

// Node naming text keys shared by the rename variables and the quality helpers.
const (
	NodeNameTextUnknown            = "unknown"
	NodeNameTextManual             = "manual"
	NodeNameTextUngrouped          = "ungrouped"
	NodeNameTextResidential        = "residential"
	NodeNameTextDatacenter         = "datacenter"
	NodeNameTextBroadcast          = "broadcast"
	NodeNameTextNative             = "native"
	NodeNameTextNotChecked         = "notChecked"
	NodeNameTextUnlockUnlocked     = "unlockUnlocked"
	NodeNameTextUnlockPartial      = "unlockPartial"
	NodeNameTextUnlockReachable    = "unlockReachable"
	NodeNameTextUnlockRestricted   = "unlockRestricted"
	NodeNameTextUnlockUnsupported  = "unlockUnsupported"
	NodeNameTextUnlockError        = "unlockError"
	NodeNameTextUnlockUnknown      = "unlockUnknown"
	NodeNameTextUnlockUntested     = "unlockUntested"
	NodeNameTextQualityPartial     = "qualityPartial"
	NodeNameTextQualityPartialIPv6 = "qualityPartialIPv6"
	NodeNameTextQualityFailed      = "qualityFailed"
	NodeNameTextQualityDisabled    = "qualityDisabled"
	NodeNameTextOther              = "other"
)

var nodeNameTexts = map[string]map[string]string{
	NodeNameLanguageEnglish: {
		NodeNameTextUnknown:            "Unknown",
		NodeNameTextManual:             "Manual",
		NodeNameTextUngrouped:          "Ungrouped",
		NodeNameTextResidential:        "Residential",
		NodeNameTextDatacenter:         "Datacenter",
		NodeNameTextBroadcast:          "Broadcast",
		NodeNameTextNative:             "Native",
		NodeNameTextNotChecked:         "Not checked",
		NodeNameTextUnlockUnlocked:     "Unlocked",
		NodeNameTextUnlockPartial:      "Partial",
		NodeNameTextUnlockReachable:    "Reachable",
		NodeNameTextUnlockRestricted:   "Restricted",
		NodeNameTextUnlockUnsupported:  "Unsupported",
		NodeNameTextUnlockError:        "Error",
		NodeNameTextUnlockUnknown:      "Unknown",
		NodeNameTextUnlockUntested:     "Untested",
		NodeNameTextQualityPartial:     "Partial result",
		NodeNameTextQualityPartialIPv6: "IPv6 partial result",
		NodeNameTextQualityFailed:      "Check failed",
		NodeNameTextQualityDisabled:    "Disabled",
		NodeNameTextOther:              "Other",
	},
	NodeNameLanguageChinese: {
		NodeNameTextUnknown:            "未知",
		NodeNameTextManual:             "手动",
		NodeNameTextUngrouped:          "未分组",
		NodeNameTextResidential:        "住宅IP",
		NodeNameTextDatacenter:         "机房IP",
		NodeNameTextBroadcast:          "广播IP",
		NodeNameTextNative:             "原生IP",
		NodeNameTextNotChecked:         "未检测",
		NodeNameTextUnlockUnlocked:     "解锁",
		NodeNameTextUnlockPartial:      "部分",
		NodeNameTextUnlockReachable:    "直连",
		NodeNameTextUnlockRestricted:   "受限",
		NodeNameTextUnlockUnsupported:  "不支持",
		NodeNameTextUnlockError:        "异常",
		NodeNameTextUnlockUnknown:      "未知",
		NodeNameTextUnlockUntested:     "未测",
		NodeNameTextQualityPartial:     "部分结果",
		NodeNameTextQualityPartialIPv6: "IPv6部分结果",
		NodeNameTextQualityFailed:      "检测失败",
		NodeNameTextQualityDisabled:    "未启用",
		NodeNameTextOther:              "其他",
	},
}

// nodeNameLanguageProvider is registered by the models layer, which owns the
// persisted setting.
var nodeNameLanguageProvider = func() string { return NodeNameLanguageEnglish }

// SetNodeNameLanguageProvider registers the resolver used for the current node
// naming language. A nil resolver keeps the English default.
func SetNodeNameLanguageProvider(provider func() string) {
	if provider == nil {
		return
	}
	nodeNameLanguageProvider = provider
}

// NormalizeNodeNameLanguage maps arbitrary language tags to a supported node
// naming language, defaulting to English.
func NormalizeNodeNameLanguage(language string) string {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(language), "_", "-"))
	if normalized == "zh" || normalized == "zh-cn" || normalized == "zh-hans" || strings.HasPrefix(normalized, "zh-hans-") {
		return NodeNameLanguageChinese
	}
	return NodeNameLanguageEnglish
}

// CurrentNodeNameLanguage reports the node naming language configured by the
// models layer.
func CurrentNodeNameLanguage() string {
	return NormalizeNodeNameLanguage(nodeNameLanguageProvider())
}

// NodeNameText resolves one node naming text in the requested language,
// falling back to English for unknown languages or keys.
func NodeNameText(language string, key string) string {
	if text, ok := nodeNameTexts[NormalizeNodeNameLanguage(language)][key]; ok {
		return text
	}
	language = NodeNameLanguageEnglish
	return nodeNameTexts[language][key]
}

// UnlockStatusTextKey maps a stored unlock status value to its text key.
func UnlockStatusTextKey(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "available":
		return NodeNameTextUnlockUnlocked
	case "partial":
		return NodeNameTextUnlockPartial
	case "reachable":
		return NodeNameTextUnlockReachable
	case "restricted":
		return NodeNameTextUnlockRestricted
	case "unsupported":
		return NodeNameTextUnlockUnsupported
	case "error":
		return NodeNameTextUnlockError
	case "unknown":
		return NodeNameTextUnlockUnknown
	default:
		return NodeNameTextUnlockUntested
	}
}

// UnlockStatusText resolves the localized label for an unlock status value.
func UnlockStatusText(language string, status string) string {
	return NodeNameText(language, UnlockStatusTextKey(status))
}
