package substore

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/models"
	"time"
)

const (
	settingEnabled          = "substore_enabled"
	settingBaseURL          = "substore_base_url"
	settingTimeoutSeconds   = "substore_timeout_seconds"
	settingAllowedTargets   = "substore_allowed_targets"
	settingMaxResponseBytes = "substore_max_response_bytes"
	defaultTimeoutSeconds   = 10
	defaultMaxResponseBytes = 8 << 20
)

type FieldValue[T any] struct {
	Value  T      `json:"value"`
	Source string `json:"source"`
}

type Settings struct {
	Enabled          bool     `json:"enabled"`
	BaseURL          string   `json:"baseUrl"`
	TimeoutSeconds   int      `json:"timeoutSeconds"`
	AllowedTargets   []string `json:"allowedTargets"`
	MaxResponseBytes int64    `json:"maxResponseBytes"`
}

type EffectiveSettings struct {
	Configured       bool                 `json:"configured"`
	SupportedTargets []string             `json:"supportedTargets"`
	Enabled          FieldValue[bool]     `json:"enabled"`
	BaseURL          FieldValue[string]   `json:"baseUrl"`
	TimeoutSeconds   FieldValue[int]      `json:"timeoutSeconds"`
	AllowedTargets   FieldValue[[]string] `json:"allowedTargets"`
	MaxResponseBytes FieldValue[int64]    `json:"maxResponseBytes"`
}

func SupportedTargets() []string {
	return []string{"loon", "egern", "stash", "surfboard", "shadowrocket", "quanx", "sing-box", "uri", "json"}
}

func LoadSettings() Settings {
	settings := Settings{
		Enabled:          getBoolSetting(settingEnabled, false),
		BaseURL:          normalizeBaseURL(getStringSetting(settingBaseURL)),
		TimeoutSeconds:   getIntSetting(settingTimeoutSeconds, defaultTimeoutSeconds),
		AllowedTargets:   normalizeTargets(strings.Split(getStringSetting(settingAllowedTargets), ",")),
		MaxResponseBytes: getInt64Setting(settingMaxResponseBytes, defaultMaxResponseBytes),
	}
	if settings.TimeoutSeconds <= 0 {
		settings.TimeoutSeconds = defaultTimeoutSeconds
	}
	if settings.MaxResponseBytes <= 0 {
		settings.MaxResponseBytes = defaultMaxResponseBytes
	}
	return settings
}

func SaveSettings(settings Settings) error {
	settings.BaseURL = normalizeBaseURL(settings.BaseURL)
	settings.AllowedTargets = normalizeTargets(settings.AllowedTargets)
	if _, err := ConfigFromSettings(settings); err != nil {
		return err
	}
	if err := models.SetSetting(settingEnabled, strconv.FormatBool(settings.Enabled)); err != nil {
		return err
	}
	if err := models.SetSetting(settingBaseURL, settings.BaseURL); err != nil {
		return err
	}
	if err := models.SetSetting(settingTimeoutSeconds, strconv.Itoa(settings.TimeoutSeconds)); err != nil {
		return err
	}
	if err := models.SetSetting(settingAllowedTargets, strings.Join(settings.AllowedTargets, ",")); err != nil {
		return err
	}
	return models.SetSetting(settingMaxResponseBytes, strconv.FormatInt(settings.MaxResponseBytes, 10))
}

func ConfigFromSettings(settings Settings) (Config, error) {
	settings.BaseURL = normalizeBaseURL(settings.BaseURL)
	settings.AllowedTargets = normalizeTargets(settings.AllowedTargets)
	if settings.TimeoutSeconds <= 0 {
		settings.TimeoutSeconds = defaultTimeoutSeconds
	}
	if settings.MaxResponseBytes <= 0 {
		settings.MaxResponseBytes = defaultMaxResponseBytes
	}
	if err := validateSettings(settings); err != nil {
		return Config{}, err
	}
	return Config{
		BaseURL:          settings.BaseURL,
		Timeout:          time.Duration(settings.TimeoutSeconds) * time.Second,
		MaxResponseBytes: settings.MaxResponseBytes,
		AllowedTargets:   settings.AllowedTargets,
	}, nil
}

func EffectiveConfig() (Config, EffectiveSettings) {
	effective := ResolveEffectiveSettings()
	if !effective.Configured {
		return Config{}, effective
	}
	cfg := Config{
		BaseURL:          effective.BaseURL.Value,
		Timeout:          time.Duration(effective.TimeoutSeconds.Value) * time.Second,
		MaxResponseBytes: effective.MaxResponseBytes.Value,
		AllowedTargets:   effective.AllowedTargets.Value,
	}
	return cfg, effective
}

func ResolveEffectiveSettings() EffectiveSettings {
	db := LoadSettings()
	effective := EffectiveSettings{SupportedTargets: SupportedTargets()}

	effective.Enabled = FieldValue[bool]{Value: db.Enabled, Source: "database"}
	effective.BaseURL = FieldValue[string]{Value: db.BaseURL, Source: "database"}
	effective.TimeoutSeconds = FieldValue[int]{Value: db.TimeoutSeconds, Source: "database"}
	effective.AllowedTargets = FieldValue[[]string]{Value: db.AllowedTargets, Source: "database"}
	effective.MaxResponseBytes = FieldValue[int64]{Value: db.MaxResponseBytes, Source: "database"}
	effective.Configured = effective.Enabled.Value && effective.BaseURL.Value != ""
	return effective
}

func normalizeBaseURL(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}

func normalizeTargets(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, ok := ResolveTarget(value); !ok {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func validateSettings(settings Settings) error {
	if settings.Enabled {
		if settings.BaseURL == "" {
			return fmt.Errorf("Sub-Store base URL is required when enabled")
		}
		parsed, err := url.Parse(settings.BaseURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return fmt.Errorf("Sub-Store base URL must be an http or https URL")
		}
	}
	if settings.TimeoutSeconds <= 0 {
		return fmt.Errorf("Sub-Store timeout must be greater than 0")
	}
	if settings.MaxResponseBytes <= 0 {
		return fmt.Errorf("Sub-Store max response bytes must be greater than 0")
	}
	return nil
}

func getStringSetting(key string) string {
	value, _ := models.GetSetting(key)
	return value
}

func getBoolSetting(key string, fallback bool) bool {
	value := strings.TrimSpace(getStringSetting(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getIntSetting(key string, fallback int) int {
	value := strings.TrimSpace(getStringSetting(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func getInt64Setting(key string, fallback int64) int64 {
	value := strings.TrimSpace(getStringSetting(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
