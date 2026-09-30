package backup

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"sublink/models"

	"github.com/robfig/cron/v3"
)

const (
	settingBaseURL           = "backup_webdav_base_url"
	settingUsername          = "backup_webdav_username"
	settingPasswordEncrypted = "backup_webdav_password_encrypted"
	settingRemotePath        = "backup_webdav_remote_path"
	settingTimeoutSeconds    = "backup_webdav_timeout_seconds"
	settingAllowHTTP         = "backup_webdav_allow_insecure_http"
	settingAllowPrivate      = "backup_webdav_allow_private_network"
	settingScheduleEnabled   = "backup_webdav_schedule_enabled"
	settingCronExpr          = "backup_webdav_cron_expr"
	defaultRemotePath        = "SublinkPro"
	defaultTimeoutSeconds    = 60
	maxTimeoutSeconds        = 600
)

var cronSpacePattern = regexp.MustCompile(`\s+`)

// Config contains the effective WebDAV connection settings. Password is never serialized.
type Config struct {
	BaseURL             string `json:"baseUrl"`
	Username            string `json:"username"`
	Password            string `json:"-"`
	RemotePath          string `json:"remotePath"`
	TimeoutSeconds      int    `json:"timeoutSeconds"`
	AllowInsecureHTTP   bool   `json:"allowInsecureHttp"`
	AllowPrivateNetwork bool   `json:"allowPrivateNetwork"`
	ScheduleEnabled     bool   `json:"scheduleEnabled"`
	CronExpr            string `json:"cronExpr"`
}

// PublicConfig is safe to return to clients.
type PublicConfig struct {
	Configured          bool   `json:"configured"`
	BaseURL             string `json:"baseUrl"`
	Username            string `json:"username"`
	HasPassword         bool   `json:"hasPassword"`
	MaskedPassword      string `json:"maskedPassword"`
	RemotePath          string `json:"remotePath"`
	TimeoutSeconds      int    `json:"timeoutSeconds"`
	AllowInsecureHTTP   bool   `json:"allowInsecureHttp"`
	AllowPrivateNetwork bool   `json:"allowPrivateNetwork"`
	ScheduleEnabled     bool   `json:"scheduleEnabled"`
	CronExpr            string `json:"cronExpr"`
}

// ConfigUpdate is accepted from the settings API. An empty Password preserves the saved password.
type ConfigUpdate struct {
	BaseURL             string `json:"baseUrl"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	ClearPassword       bool   `json:"clearPassword"`
	RemotePath          string `json:"remotePath"`
	TimeoutSeconds      int    `json:"timeoutSeconds"`
	AllowInsecureHTTP   bool   `json:"allowInsecureHttp"`
	AllowPrivateNetwork bool   `json:"allowPrivateNetwork"`
	ScheduleEnabled     bool   `json:"scheduleEnabled"`
	CronExpr            string `json:"cronExpr"`
}

func LoadConfig() (Config, error) {
	password := ""
	encrypted := settingValue(settingPasswordEncrypted)
	if encrypted != "" {
		decrypted, err := models.DecryptUserAISecret(encrypted)
		if err != nil {
			return Config{}, fmt.Errorf("解密 WebDAV 密码失败: %w", err)
		}
		password = decrypted
	}
	cfg := Config{
		BaseURL:             settingValue(settingBaseURL),
		Username:            settingValue(settingUsername),
		Password:            password,
		RemotePath:          settingValue(settingRemotePath),
		TimeoutSeconds:      settingInt(settingTimeoutSeconds, defaultTimeoutSeconds),
		AllowInsecureHTTP:   settingBool(settingAllowHTTP, false),
		AllowPrivateNetwork: settingBool(settingAllowPrivate, false),
		ScheduleEnabled:     settingBool(settingScheduleEnabled, false),
		CronExpr:            settingValue(settingCronExpr),
	}
	if cfg.RemotePath == "" {
		cfg.RemotePath = defaultRemotePath
	}
	return normalizeAndValidateConfig(cfg, false)
}

func ResolveConfig(update ConfigUpdate) (Config, error) {
	password := ""
	if !update.ClearPassword {
		if update.Password != "" {
			password = update.Password
		} else {
			var err error
			password, err = loadStoredPassword()
			if err != nil {
				return Config{}, err
			}
		}
	}
	cfg := Config{
		BaseURL:             update.BaseURL,
		Username:            update.Username,
		Password:            password,
		RemotePath:          update.RemotePath,
		TimeoutSeconds:      update.TimeoutSeconds,
		AllowInsecureHTTP:   update.AllowInsecureHTTP,
		AllowPrivateNetwork: update.AllowPrivateNetwork,
		ScheduleEnabled:     update.ScheduleEnabled,
		CronExpr:            update.CronExpr,
	}
	return normalizeAndValidateConfig(cfg, true)
}

func SaveConfig(cfg Config) error {
	cfg, err := normalizeAndValidateConfig(cfg, true)
	if err != nil {
		return err
	}
	encrypted := ""
	if cfg.Password != "" {
		encrypted, err = models.EncryptUserAISecret(cfg.Password)
		if err != nil {
			return fmt.Errorf("加密 WebDAV 密码失败: %w", err)
		}
	}
	values := []struct{ key, value string }{
		{settingBaseURL, cfg.BaseURL},
		{settingUsername, cfg.Username},
		{settingPasswordEncrypted, encrypted},
		{settingRemotePath, cfg.RemotePath},
		{settingTimeoutSeconds, strconv.Itoa(cfg.TimeoutSeconds)},
		{settingAllowHTTP, strconv.FormatBool(cfg.AllowInsecureHTTP)},
		{settingAllowPrivate, strconv.FormatBool(cfg.AllowPrivateNetwork)},
		{settingScheduleEnabled, strconv.FormatBool(cfg.ScheduleEnabled)},
		{settingCronExpr, cfg.CronExpr},
	}
	for _, item := range values {
		if err := models.SetSetting(item.key, item.value); err != nil {
			return err
		}
	}
	return nil
}

func ToPublicConfig(cfg Config) PublicConfig {
	return PublicConfig{
		Configured:          cfg.BaseURL != "",
		BaseURL:             cfg.BaseURL,
		Username:            cfg.Username,
		HasPassword:         cfg.Password != "",
		MaskedPassword:      models.MaskSecret(cfg.Password),
		RemotePath:          cfg.RemotePath,
		TimeoutSeconds:      cfg.TimeoutSeconds,
		AllowInsecureHTTP:   cfg.AllowInsecureHTTP,
		AllowPrivateNetwork: cfg.AllowPrivateNetwork,
		ScheduleEnabled:     cfg.ScheduleEnabled,
		CronExpr:            cfg.CronExpr,
	}
}

func normalizeAndValidateConfig(cfg Config, requireURL bool) (Config, error) {
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	cfg.Username = strings.TrimSpace(cfg.Username)
	cfg.RemotePath = strings.Trim(strings.TrimSpace(cfg.RemotePath), "/")
	if len(cfg.BaseURL) > 2048 {
		return Config{}, errors.New("WebDAV 地址过长")
	}
	if len(cfg.Username) > 512 || len(cfg.Password) > 4096 {
		return Config{}, errors.New("WebDAV 凭据过长")
	}
	if len(cfg.RemotePath) > 512 {
		return Config{}, errors.New("WebDAV 远程目录过长")
	}
	if cfg.RemotePath == "" {
		cfg.RemotePath = defaultRemotePath
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = defaultTimeoutSeconds
	}
	if cfg.TimeoutSeconds > maxTimeoutSeconds {
		return Config{}, fmt.Errorf("WebDAV 超时时间不能超过 %d 秒", maxTimeoutSeconds)
	}
	if err := validateRemotePath(cfg.RemotePath); err != nil {
		return Config{}, err
	}
	cfg.CronExpr = cleanCronExpression(cfg.CronExpr)
	if cfg.ScheduleEnabled && cfg.CronExpr == "" {
		return Config{}, errors.New("启用定时备份时必须填写 Cron 表达式")
	}
	if cfg.CronExpr != "" {
		if err := validateCronExpression(cfg.CronExpr); err != nil {
			return Config{}, err
		}
	}
	if cfg.BaseURL == "" {
		if requireURL {
			return Config{}, errors.New("WebDAV 地址不能为空")
		}
		return cfg, nil
	}
	parsed, err := url.Parse(cfg.BaseURL)
	if err != nil || parsed.Host == "" {
		return Config{}, errors.New("WebDAV 地址无效")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && scheme != "http" {
		return Config{}, errors.New("WebDAV 地址仅支持 http 或 https")
	}
	if scheme == "http" && !cfg.AllowInsecureHTTP {
		return Config{}, errors.New("HTTP WebDAV 需要显式允许不安全连接")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return Config{}, errors.New("WebDAV 地址不能包含账号、查询参数或锚点")
	}
	return cfg, nil
}

func validateCronExpression(expr string) error {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	if _, err := parser.Parse(expr); err != nil {
		return fmt.Errorf("cron 表达式无效: %w", err)
	}
	return nil
}

func cleanCronExpression(cronExpr string) string {
	cleaned := strings.TrimSpace(cronExpr)
	return cronSpacePattern.ReplaceAllString(cleaned, " ")
}

func validateRemotePath(value string) error {
	if strings.Contains(value, "\\") || strings.ContainsRune(value, '\x00') {
		return errors.New("WebDAV 远程目录包含非法字符")
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return errors.New("WebDAV 远程目录无效")
		}
	}
	return nil
}

func loadStoredPassword() (string, error) {
	encrypted := settingValue(settingPasswordEncrypted)
	if encrypted == "" {
		return "", nil
	}
	password, err := models.DecryptUserAISecret(encrypted)
	if err != nil {
		return "", fmt.Errorf("解密 WebDAV 密码失败: %w", err)
	}
	return password, nil
}

func settingValue(key string) string {
	value, _ := models.GetSetting(key)
	return strings.TrimSpace(value)
}

func settingInt(key string, fallback int) int {
	value, err := strconv.Atoi(settingValue(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func settingBool(key string, fallback bool) bool {
	value := settingValue(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func (cfg Config) Timeout() time.Duration {
	return time.Duration(cfg.TimeoutSeconds) * time.Second
}

func PreservedSettingKeys() []string {
	return []string{
		settingBaseURL,
		settingUsername,
		settingPasswordEncrypted,
		settingRemotePath,
		settingTimeoutSeconds,
		settingAllowHTTP,
		settingAllowPrivate,
		settingScheduleEnabled,
		settingCronExpr,
	}
}
