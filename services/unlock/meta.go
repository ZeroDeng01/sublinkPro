package unlock

import (
	"fmt"
	"strings"
	"sublink/models"
)

type UnlockCheckerMeta interface {
	Meta() models.UnlockProviderMeta
}

type UnlockCheckerRenameMeta interface {
	RenameVariableMeta() models.UnlockRenameVariableMeta
}

var unlockStatusMetas = []models.UnlockStatusMeta{
	{Value: models.UnlockStatusAvailable, Label: "解锁", ShortLabel: "解锁", Description: "可正常使用", Color: "success", Severity: "success"},
	{Value: models.UnlockStatusPartial, Label: "部分", ShortLabel: "部分", Description: "仅部分能力可用", Color: "warning", Severity: "warning"},
	{Value: models.UnlockStatusReachable, Label: "直连", ShortLabel: "直连", Description: "可以访问，但不代表完整能力可用", Color: "info", Severity: "info"},
	{Value: models.UnlockStatusRestricted, Label: "受限", ShortLabel: "受限", Description: "当前地区或出口被限制", Color: "error", Severity: "error"},
	{Value: models.UnlockStatusUnsupported, Label: "不支持", ShortLabel: "不支持", Description: "当前地区不在官方支持范围内", Color: "warning", Severity: "warning"},
	{Value: models.UnlockStatusUnknown, Label: "未知", ShortLabel: "未知", Description: "当前无法稳定判断结果", Color: "default", Severity: "default"},
	{Value: models.UnlockStatusError, Label: "异常", ShortLabel: "异常", Description: "本轮检测出现错误", Color: "error", Severity: "error"},
	{Value: models.UnlockStatusUntested, Label: "未测", ShortLabel: "未测", Description: "尚未执行检测", Color: "default", Severity: "default"},
}

var unlockStatusMetaByValue = buildUnlockStatusMetaIndex(unlockStatusMetas)

func buildUnlockStatusMetaIndex(items []models.UnlockStatusMeta) map[string]models.UnlockStatusMeta {
	indexed := make(map[string]models.UnlockStatusMeta, len(items))
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.Value))
		if key == "" {
			continue
		}
		indexed[key] = item
	}
	return indexed
}

func ListUnlockStatusMetas() []models.UnlockStatusMeta {
	items := make([]models.UnlockStatusMeta, len(unlockStatusMetas))
	copy(items, unlockStatusMetas)
	return items
}

func NormalizeUnlockStatus(status string) string {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case "", "all":
		return ""
	}
	if _, exists := unlockStatusMetaByValue[normalized]; exists {
		return normalized
	}
	return ""
}

func GetUnlockStatusMeta(status string) (models.UnlockStatusMeta, bool) {
	meta, exists := unlockStatusMetaByValue[strings.ToLower(strings.TrimSpace(status))]
	return meta, exists
}

func GetUnlockProviderMeta(provider string) models.UnlockProviderMeta {
	normalized := models.NormalizeUnlockProvider(provider)
	if checker, ok := GetUnlockChecker(normalized); ok {
		if providerMetaChecker, ok := checker.(UnlockCheckerMeta); ok {
			meta := providerMetaChecker.Meta()
			if meta.Value == "" {
				meta.Value = normalized
			}
			if meta.Label == "" {
				meta.Label = provider
			}
			if meta.Category == "" {
				meta.Category = "custom"
			}
			return meta
		}
	}
	if normalized == "" {
		normalized = provider
	}
	return models.UnlockProviderMeta{Value: normalized, Label: provider, Category: "custom"}
}

func BuildUnlockRenameVariables(providers []string) []models.UnlockRenameVariableMeta {
	normalizedProviders := ResolveUnlockProviders(providers)
	if len(providers) > 0 && len(normalizedProviders) == 0 {
		normalizedProviders = models.NormalizeUnlockProviders(providers)
	}
	variables := make([]models.UnlockRenameVariableMeta, 0, len(normalizedProviders))
	for _, provider := range normalizedProviders {
		variable := buildUnlockRenameVariable(provider)
		if variable.Key == "" {
			continue
		}
		variables = append(variables, variable)
	}
	return variables
}

func buildUnlockRenameVariable(provider string) models.UnlockRenameVariableMeta {
	if checker, ok := GetUnlockChecker(provider); ok {
		if renameMetaChecker, ok := checker.(UnlockCheckerRenameMeta); ok {
			meta := renameMetaChecker.RenameVariableMeta()
			if meta.Provider == "" {
				meta.Provider = provider
			}
			if meta.Key == "" {
				meta.Key = fmt.Sprintf("$Unlock(%s)", provider)
			}
			if meta.Label == "" {
				providerMeta := GetUnlockProviderMeta(provider)
				meta.Label = providerMeta.Label + " 解锁"
			}
			if meta.Description == "" {
				meta.Description = "输出该服务的紧凑解锁结果摘要，例如“解锁-US”或“受限”"
			}
			return meta
		}
	}
	providerMeta := GetUnlockProviderMeta(provider)
	return models.UnlockRenameVariableMeta{
		Key:         fmt.Sprintf("$Unlock(%s)", provider),
		Label:       providerMeta.Label + " 解锁",
		Description: "输出该服务的紧凑解锁结果摘要，例如“解锁-US”或“受限”",
		Provider:    provider,
	}
}

func init() {
	models.RegisterUnlockStatusMetaResolver(ListUnlockStatusMetas)
	models.RegisterUnlockProviderMetaResolver(GetUnlockProviderMeta)
	models.RegisterUnlockRenameVariableResolver(BuildUnlockRenameVariables)
	models.RegisterUnlockStatusNormalizer(NormalizeUnlockStatus)
}
