package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sublink/config"
)

const userAISecretVersion = "v1"
const defaultSystemAIMaxTokens = 400000

const (
	SystemAIRequestTypeResponses       = "responses"
	SystemAIRequestTypeChatCompletions = "chat_completions"
)

type UserAISettings struct {
	Enabled         bool              `json:"enabled"`
	BaseURL         string            `json:"baseUrl"`
	Model           string            `json:"model"`
	HasKey          bool              `json:"hasKey"`
	MaskedKey       string            `json:"maskedKey"`
	Temperature     float64           `json:"temperature"`
	MaxTokens       int               `json:"maxTokens"`
	ExtraHeaders    map[string]string `json:"extraHeaders,omitempty"`
	ProviderType    string            `json:"providerType"`
	RequestType     string            `json:"requestType"`
	Configured      bool              `json:"configured"`
	RawAPIKey       string            `json:"-"`
	ExtraHeadersRaw string            `json:"-"`
}

const (
	systemAIEnabledKey      = "ai_enabled"
	systemAIBaseURLKey      = "ai_base_url"
	systemAIModelKey        = "ai_model"
	systemAIAPIKeyKey       = "ai_api_key_encrypted"
	systemAITemperatureKey  = "ai_temperature"
	systemAIMaxTokensKey    = "ai_max_tokens"
	systemAIExtraHeadersKey = "ai_extra_headers"
	systemAIRequestTypeKey  = "ai_request_type"
)

func NormalizeSystemAIRequestType(value string) string {
	switch strings.TrimSpace(value) {
	case SystemAIRequestTypeChatCompletions:
		return SystemAIRequestTypeChatCompletions
	default:
		return SystemAIRequestTypeResponses
	}
}

func userAIEncryptionKey() ([]byte, error) {
	keyMaterial := strings.TrimSpace(config.GetAPIEncryptionKey())
	if len(keyMaterial) < 32 {
		return nil, fmt.Errorf("API_ENCRYPTION_KEY 未设置或长度不足，无法安全存储 AI 密钥")
	}
	sum := sha256.Sum256([]byte(keyMaterial))
	return sum[:], nil
}

func EncryptUserAISecret(secret string) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", nil
	}
	key, err := userAIEncryptionKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(secret), nil)
	payload := append(nonce, ciphertext...)
	return userAISecretVersion + ":" + base64.StdEncoding.EncodeToString(payload), nil
}

func DecryptUserAISecret(secret string) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", nil
	}
	parts := strings.SplitN(strings.TrimSpace(secret), ":", 2)
	if len(parts) != 2 || parts[0] != userAISecretVersion {
		return "", fmt.Errorf("不支持的 AI 密钥格式")
	}
	key, err := userAIEncryptionKey()
	if err != nil {
		return "", err
	}
	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < gcm.NonceSize() {
		return "", fmt.Errorf("AI 密钥数据损坏")
	}
	nonce := payload[:gcm.NonceSize()]
	ciphertext := payload[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func MaskSecret(secret string) string {
	trimmed := strings.TrimSpace(secret)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= 8 {
		return strings.Repeat("*", len(trimmed))
	}
	return trimmed[:4] + strings.Repeat("*", len(trimmed)-8) + trimmed[len(trimmed)-4:]
}

func getSystemSettingValue(key string) string {
	value, _ := GetSetting(key)
	return strings.TrimSpace(value)
}

func parseSystemAIFloat(key string, fallback float64) float64 {
	value := getSystemSettingValue(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseSystemAIInt(key string, fallback int) int {
	value := getSystemSettingValue(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func GetSystemAISettings() (UserAISettings, error) {
	encryptedKey := getSystemSettingValue(systemAIAPIKeyKey)
	settings := UserAISettings{
		Enabled:      getSystemSettingValue(systemAIEnabledKey) == "true",
		BaseURL:      getSystemSettingValue(systemAIBaseURLKey),
		Model:        getSystemSettingValue(systemAIModelKey),
		HasKey:       encryptedKey != "",
		Temperature:  parseSystemAIFloat(systemAITemperatureKey, 0.2),
		MaxTokens:    parseSystemAIInt(systemAIMaxTokensKey, defaultSystemAIMaxTokens),
		ProviderType: "openai_compatible",
		RequestType:  NormalizeSystemAIRequestType(getSystemSettingValue(systemAIRequestTypeKey)),
	}
	if settings.MaxTokens <= 0 {
		settings.MaxTokens = defaultSystemAIMaxTokens
	}
	if settings.HasKey {
		key, err := DecryptUserAISecret(encryptedKey)
		if err != nil {
			return UserAISettings{}, err
		}
		settings.RawAPIKey = key
		settings.MaskedKey = MaskSecret(key)
	}
	settings.ExtraHeadersRaw = getSystemSettingValue(systemAIExtraHeadersKey)
	if settings.ExtraHeadersRaw != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(settings.ExtraHeadersRaw), &headers); err != nil {
			return UserAISettings{}, err
		}
		settings.ExtraHeaders = headers
	}
	settings.Configured = settings.BaseURL != "" && settings.Model != "" && settings.HasKey
	return settings, nil
}

func UpdateSystemAISettings(input UserAISettings) error {
	if err := SetSetting(systemAIEnabledKey, strconv.FormatBool(input.Enabled)); err != nil {
		return err
	}
	if err := SetSetting(systemAIBaseURLKey, strings.TrimSpace(input.BaseURL)); err != nil {
		return err
	}
	if err := SetSetting(systemAIModelKey, strings.TrimSpace(input.Model)); err != nil {
		return err
	}
	if err := SetSetting(systemAITemperatureKey, strconv.FormatFloat(input.Temperature, 'f', -1, 64)); err != nil {
		return err
	}
	if err := SetSetting(systemAIMaxTokensKey, strconv.Itoa(input.MaxTokens)); err != nil {
		return err
	}
	if err := SetSetting(systemAIExtraHeadersKey, strings.TrimSpace(input.ExtraHeadersRaw)); err != nil {
		return err
	}
	if err := SetSetting(systemAIRequestTypeKey, NormalizeSystemAIRequestType(input.RequestType)); err != nil {
		return err
	}
	if strings.TrimSpace(input.RawAPIKey) == "" {
		return nil
	}
	encrypted, err := EncryptUserAISecret(strings.TrimSpace(input.RawAPIKey))
	if err != nil {
		return err
	}
	return SetSetting(systemAIAPIKeyKey, encrypted)
}

func (user *User) GetAISettings() (UserAISettings, error) {
	return GetSystemAISettings()
}

func (user *User) UpdateAISettings(input UserAISettings) error {
	return UpdateSystemAISettings(input)
}
