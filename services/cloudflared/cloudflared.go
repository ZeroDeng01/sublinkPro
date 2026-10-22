package cloudflared

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"sublink/models"
	"sublink/utils"
)

const (
	settingEnabledKey = "cloudflared_enabled"
	settingTokenKey   = "cloudflared_tunnel_token_encrypted"
	maxMessageLength  = 2000
)

// Config 描述 Cloudflare Tunnel 的持久化配置。
type Config struct {
	Enabled     bool   `json:"enabled"`
	HasToken    bool   `json:"hasToken"`
	MaskedToken string `json:"maskedToken"`
	RawToken    string `json:"-"`
}

// Status 描述 cloudflared 当前安装、配置和进程状态。
type Status struct {
	Installed    bool   `json:"installed"`
	Path         string `json:"path"`
	Version      string `json:"version"`
	Running      bool   `json:"running"`
	Enabled      bool   `json:"enabled"`
	HasToken     bool   `json:"hasToken"`
	MaskedToken  string `json:"maskedToken"`
	LastMessage  string `json:"lastMessage"`
	LastError    string `json:"lastError"`
	CommandLabel string `json:"commandLabel"`
}

// Manager 托管单个 cloudflared 子进程。
type Manager struct {
	mu          sync.Mutex
	cmd         *exec.Cmd
	cancel      context.CancelFunc
	done        chan struct{}
	running     bool
	stopping    bool
	lastMessage string
	lastError   string
}

var defaultManager = &Manager{}

// DefaultManager 返回全局 cloudflared 管理器。
func DefaultManager() *Manager {
	return defaultManager
}

// LoadConfig 从系统设置中读取 Cloudflare Tunnel 配置。
func LoadConfig() (Config, error) {
	enabled := strings.TrimSpace(settingValue(settingEnabledKey)) == "true"
	encryptedToken := strings.TrimSpace(settingValue(settingTokenKey))
	config := Config{Enabled: enabled, HasToken: encryptedToken != ""}
	if !config.HasToken {
		return config, nil
	}

	token, err := models.DecryptUserAISecret(encryptedToken)
	if err != nil {
		return Config{}, err
	}
	config.RawToken = token
	config.MaskedToken = models.MaskSecret(token)
	return config, nil
}

// SaveEnabled 保存是否允许启动 cloudflared。
func SaveEnabled(enabled bool) error {
	return models.SetSetting(settingEnabledKey, fmt.Sprintf("%t", enabled))
}

// SaveToken 加密保存 Cloudflare Tunnel token。
func SaveToken(token string) error {
	trimmed := normalizeToken(token)
	if trimmed == "" {
		return errors.New("cloudflare tunnel token 不能为空")
	}
	encrypted, err := models.EncryptUserAISecret(trimmed)
	if err != nil {
		return err
	}
	return models.SetSetting(settingTokenKey, encrypted)
}

// ClearToken 清空已保存的 Cloudflare Tunnel token。
func ClearToken() error {
	return models.SetSetting(settingTokenKey, "")
}

// AutoStart 在应用启动时根据页面保存的配置启动 cloudflared。
func AutoStart() {
	config, err := LoadConfig()
	if err != nil {
		utils.Error("[cloudflared] 加载配置失败: %v", err)
		return
	}
	if !config.Enabled || !config.HasToken {
		return
	}
	if err := defaultManager.Start(""); err != nil {
		utils.Error("[cloudflared] 自动启动失败: %v", err)
	}
}

// Status 返回当前 cloudflared 状态。
func (m *Manager) Status() Status {
	path, installed := lookupCloudflared()
	version := ""
	if installed {
		version = cloudflaredVersion(path)
	}
	config, err := LoadConfig()
	status := Status{
		Installed:    installed,
		Path:         path,
		Version:      version,
		CommandLabel: "cloudflared tunnel --no-autoupdate run",
	}
	if err == nil {
		status.Enabled = config.Enabled
		status.HasToken = config.HasToken
		status.MaskedToken = config.MaskedToken
	} else {
		status.LastError = err.Error()
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	status.Running = m.running
	status.LastMessage = m.lastMessage
	if status.LastError == "" {
		status.LastError = m.lastError
	}
	return status
}

// Start 启动 cloudflared。token 非空时会先加密保存并启用自动启动。
func (m *Manager) Start(token string) error {
	return m.start(token, true)
}

func (m *Manager) start(token string, persist bool) error {
	normalized := normalizeToken(token)
	if normalized != "" && persist {
		if err := SaveToken(normalized); err != nil {
			return err
		}
	}
	if persist {
		if err := SaveEnabled(true); err != nil {
			return err
		}
	}

	config, err := LoadConfig()
	if err != nil {
		return err
	}
	if normalized != "" {
		config.RawToken = normalized
		config.HasToken = true
		config.MaskedToken = models.MaskSecret(normalized)
	}

	if !config.HasToken || strings.TrimSpace(config.RawToken) == "" {
		return errors.New("请先填写 Cloudflare Tunnel token")
	}

	path, installed := lookupCloudflared()
	if !installed {
		return errors.New("未找到 cloudflared，请先安装 cloudflared 或使用内置 cloudflared 的 Docker 镜像")
	}

	m.mu.Lock()
	if m.cmd != nil || m.running {
		m.mu.Unlock()
		return errors.New("cloudflared 已在运行")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, path, "tunnel", "--no-autoupdate", "run")
	cmd.Env = cloudflaredEnv(config.RawToken)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		m.mu.Unlock()
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		m.mu.Unlock()
		return err
	}
	if err := cmd.Start(); err != nil {
		cancel()
		m.mu.Unlock()
		return err
	}

	m.cmd = cmd
	m.cancel = cancel
	m.done = make(chan struct{})
	m.running = true
	m.stopping = false
	m.lastMessage = "cloudflared 正在启动"
	m.lastError = ""
	done := m.done
	m.mu.Unlock()

	go m.scanOutput(stdout, false)
	go m.scanOutput(stderr, true)
	go m.wait(cmd, done)
	return nil
}

// Stop 停止当前 cloudflared 子进程，并关闭自动启动。
func (m *Manager) Stop() error {
	return m.stop(true)
}

// Shutdown 停止当前 cloudflared 子进程，但保留自动启动配置。
func (m *Manager) Shutdown() error {
	return m.stop(false)
}

func (m *Manager) stop(disableAutostart bool) error {
	if disableAutostart {
		if err := SaveEnabled(false); err != nil {
			return err
		}
	}

	m.mu.Lock()
	if m.cmd == nil || !m.running {
		m.lastMessage = "cloudflared 未运行"
		m.mu.Unlock()
		return nil
	}
	cmd := m.cmd
	done := m.done
	cancel := m.cancel
	m.stopping = true
	m.lastMessage = "正在停止 cloudflared"
	m.mu.Unlock()

	if cmd.Process != nil {
		if err := cmd.Process.Signal(os.Interrupt); err != nil && cancel != nil {
			cancel()
		}
	} else if cancel != nil {
		cancel()
	}

	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		if cancel != nil {
			cancel()
		}
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		<-done
		return nil
	}
}

func (m *Manager) wait(cmd *exec.Cmd, done chan struct{}) {
	err := cmd.Wait()
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cmd == cmd {
		m.cmd = nil
		m.cancel = nil
		m.running = false
		if err != nil && !m.stopping {
			m.lastError = m.redact(truncateMessage(err.Error()))
			m.lastMessage = "cloudflared 异常退出"
		} else {
			m.lastMessage = "cloudflared 已停止"
		}
		m.stopping = false
	}
	close(done)
}

func (m *Manager) scanOutput(reader io.Reader, isError bool) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024), 64*1024)
	for scanner.Scan() {
		message := m.redact(truncateMessage(scanner.Text()))
		if message == "" {
			continue
		}
		m.mu.Lock()
		if isError {
			m.lastError = message
		} else {
			m.lastMessage = message
		}
		m.mu.Unlock()
	}
}

func (m *Manager) redact(message string) string {
	redacted := message
	config, err := LoadConfig()
	if err == nil && strings.TrimSpace(config.RawToken) != "" {
		redacted = strings.ReplaceAll(redacted, config.RawToken, "[REDACTED]")
	}
	for _, marker := range []string{"TUNNEL_TOKEN="} {
		redacted = redactAssignment(redacted, marker)
	}
	return redacted
}

func cloudflaredEnv(token string) []string {
	env := []string{"TUNNEL_TOKEN=" + token}
	for _, key := range []string{"PATH", "SSL_CERT_FILE", "SSL_CERT_DIR"} {
		if value, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+value)
		}
	}
	return env
}

func redactAssignment(message, marker string) string {
	idx := strings.Index(message, marker)
	if idx == -1 {
		return message
	}
	start := idx + len(marker)
	end := start
	for end < len(message) && !strings.ContainsRune(" \t\r\n'\"", rune(message[end])) {
		end++
	}
	return message[:start] + "[REDACTED]" + message[end:]
}

func lookupCloudflared() (string, bool) {
	if info, err := os.Stat("/usr/local/bin/cloudflared"); err == nil && !info.IsDir() {
		return "/usr/local/bin/cloudflared", true
	}
	path, err := exec.LookPath("cloudflared")
	if err != nil {
		return "", false
	}
	return path, true
}

func cloudflaredVersion(path string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	output, err := exec.CommandContext(ctx, path, "version").CombinedOutput()
	if err != nil {
		return ""
	}
	return parseVersionOutput(string(output))
}

func parseVersionOutput(output string) string {
	line := strings.TrimSpace(output)
	if idx := strings.IndexAny(line, "\r\n"); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}
	fields := strings.Fields(line)
	for i, field := range fields {
		if strings.EqualFold(field, "version") && i+1 < len(fields) {
			return fields[i+1]
		}
	}
	return line
}

func settingValue(key string) string {
	value, _ := models.GetSetting(key)
	return strings.TrimSpace(value)
}

func normalizeToken(token string) string {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return ""
	}
	parts := strings.Fields(trimmed)
	if len(parts) > 0 {
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return trimmed
}

func truncateMessage(message string) string {
	trimmed := strings.TrimSpace(message)
	if len(trimmed) <= maxMessageLength {
		return trimmed
	}
	return trimmed[:maxMessageLength] + "..."
}
