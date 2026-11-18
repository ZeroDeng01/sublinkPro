package models

import (
	"strings"

	"sublink/utils"
)

// NodeNameLanguageSettingKey stores the language used for generated node names.
// Node names are rendered inside client apps, so the value must survive
// independently of the browser that changed it.
const NodeNameLanguageSettingKey = "node_name_language"

func init() {
	utils.SetNodeNameLanguageProvider(GetNodeNameLanguage)
}

// GetNodeNameLanguage returns the configured node naming language, defaulting
// to English when the setting is missing or unreadable.
func GetNodeNameLanguage() string {
	value, err := GetSetting(NodeNameLanguageSettingKey)
	if err != nil || strings.TrimSpace(value) == "" {
		return utils.NodeNameLanguageEnglish
	}
	return utils.NormalizeNodeNameLanguage(value)
}

// SetNodeNameLanguage persists the node naming language.
func SetNodeNameLanguage(language string) error {
	return SetSetting(NodeNameLanguageSettingKey, utils.NormalizeNodeNameLanguage(language))
}
