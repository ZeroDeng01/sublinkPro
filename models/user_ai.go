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
	"strings"
	"sublink/config"
	"sublink/database"
)

const userAISecretVersion = "v1"

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
	Configured      bool              `json:"configured"`
	RawAPIKey       string            `json:"-"`
	ExtraHeadersRaw string            `json:"-"`
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

func (user *User) GetAISettings() (UserAISettings, error) {
	settings := UserAISettings{
		Enabled:      user.AIEnabled,
		BaseURL:      strings.TrimSpace(user.AIBaseURL),
		Model:        strings.TrimSpace(user.AIModel),
		HasKey:       strings.TrimSpace(user.AIAPIKeyEncrypted) != "",
		Temperature:  user.AITemperature,
		MaxTokens:    user.AIMaxTokens,
		ProviderType: "openai_compatible",
	}
	if settings.MaxTokens <= 0 {
		settings.MaxTokens = 1200
	}
	if settings.Temperature == 0 {
		settings.Temperature = 0.2
	}
	if settings.HasKey {
		key, err := DecryptUserAISecret(user.AIAPIKeyEncrypted)
		if err != nil {
			return UserAISettings{}, err
		}
		settings.RawAPIKey = key
		settings.MaskedKey = MaskSecret(key)
	}
	settings.ExtraHeadersRaw = strings.TrimSpace(user.AIExtraHeaders)
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

func (user *User) UpdateAISettings(input UserAISettings) error {
	updates := map[string]interface{}{
		"ai_enabled":       input.Enabled,
		"ai_base_url":      strings.TrimSpace(input.BaseURL),
		"ai_model":         strings.TrimSpace(input.Model),
		"ai_temperature":   input.Temperature,
		"ai_max_tokens":    input.MaxTokens,
		"ai_extra_headers": strings.TrimSpace(input.ExtraHeadersRaw),
	}
	if input.RawAPIKey != "" {
		encrypted, err := EncryptUserAISecret(strings.TrimSpace(input.RawAPIKey))
		if err != nil {
			return err
		}
		updates["ai_api_key_encrypted"] = encrypted
	}
	if err := database.DB.Model(&User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return err
	}
	var updated User
	if err := database.DB.First(&updated, user.ID).Error; err == nil {
		*user = updated
		userCache.Set(user.ID, *user)
	}
	return nil
}
