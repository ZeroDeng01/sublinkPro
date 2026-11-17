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
	{Value: models.UnlockStatusAvailable, Label: "Unlocked", ShortLabel: "Unlocked", Description: "Available normally", Color: "success", Severity: "success"},
	{Value: models.UnlockStatusPartial, Label: "Partial", ShortLabel: "Partial", Description: "Only some capabilities are available", Color: "warning", Severity: "warning"},
	{Value: models.UnlockStatusReachable, Label: "Reachable", ShortLabel: "Reachable", Description: "Reachable, but not necessarily fully available", Color: "info", Severity: "info"},
	{Value: models.UnlockStatusRestricted, Label: "Restricted", ShortLabel: "Restricted", Description: "The current region or exit is restricted", Color: "error", Severity: "error"},
	{Value: models.UnlockStatusUnsupported, Label: "Unsupported", ShortLabel: "Unsupported", Description: "The current region is outside the officially supported range", Color: "warning", Severity: "warning"},
	{Value: models.UnlockStatusUnknown, Label: "Unknown", ShortLabel: "Unknown", Description: "The result cannot be determined reliably", Color: "default", Severity: "default"},
	{Value: models.UnlockStatusError, Label: "Error", ShortLabel: "Error", Description: "This check encountered an error", Color: "error", Severity: "error"},
	{Value: models.UnlockStatusUntested, Label: "Untested", ShortLabel: "Untested", Description: "No check has been run yet", Color: "default", Severity: "default"},
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
				meta.Label = providerMeta.Label + " Unlock"
			}
			if meta.Description == "" {
				meta.Description = "Outputs a compact unlock summary for this service, e.g. “Unlocked-US” or “Restricted”"
			}
			return meta
		}
	}
	providerMeta := GetUnlockProviderMeta(provider)
	return models.UnlockRenameVariableMeta{
		Key:         fmt.Sprintf("$Unlock(%s)", provider),
		Label:       providerMeta.Label + " Unlock",
		Description: "Outputs a compact unlock summary for this service, e.g. “Unlocked-US” or “Restricted”",
		Provider:    provider,
	}
}

func init() {
	models.RegisterUnlockStatusMetaResolver(ListUnlockStatusMetas)
	models.RegisterUnlockProviderMetaResolver(GetUnlockProviderMeta)
	models.RegisterUnlockRenameVariableResolver(BuildUnlockRenameVariables)
	models.RegisterUnlockStatusNormalizer(NormalizeUnlockStatus)
}
