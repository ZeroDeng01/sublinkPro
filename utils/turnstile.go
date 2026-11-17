package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TurnstileVerifyURL Cloudflare Turnstile 验证 API 地址
const TurnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// TurnstileResponse Turnstile API 响应结构
type TurnstileResponse struct {
	Success     bool     `json:"success"`      // 验证是否成功
	ChallengeTs string   `json:"challenge_ts"` // 挑战时间戳
	Hostname    string   `json:"hostname"`     // 主机名
	ErrorCodes  []string `json:"error-codes"`  // 错误代码列表
	Action      string   `json:"action"`       // 操作标识
	Cdata       string   `json:"cdata"`        // 自定义数据
}

// VerifyTurnstile 验证 Cloudflare Turnstile 令牌
// token: 前端传递的 cf-turnstile-response
// secretKey: Turnstile Secret Key
// remoteIP: 用户 IP 地址（可选）
// proxyLink: 代理节点链接（mihomo 格式，为空则直连）
func VerifyTurnstile(token, secretKey, remoteIP, proxyLink string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("turnstile token is empty")
	}
	if secretKey == "" {
		return false, fmt.Errorf("turnstile secret key is not configured")
	}

	// 构建请求数据
	formData := url.Values{}
	formData.Set("secret", secretKey)
	formData.Set("response", token)
	if remoteIP != "" {
		formData.Set("remoteip", remoteIP)
	}

	// 创建 HTTP 客户端（支持代理）
	var client *http.Client
	var err error
	useProxy := proxyLink != ""

	client, usedProxy, err := CreateProxyHTTPClient(useProxy, proxyLink, 30*time.Second)
	if err != nil {
		Error("Failed to create Turnstile HTTP client: %v", err)
		return false, fmt.Errorf("failed to create HTTP client: %v", err)
	}

	if usedProxy != "" {
		Debug("Turnstile verification uses proxy: %s", usedProxy)
	}

	// 发送 POST 请求
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, TurnstileVerifyURL, strings.NewReader(formData.Encode()))
	if err != nil {
		Error("Failed to create Turnstile verification request: %v", err)
		return false, fmt.Errorf("failed to create verification request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		Error("Turnstile verification request failed: %v", err)
		return false, fmt.Errorf("verification request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Error("Failed to read Turnstile response: %v", err)
		return false, fmt.Errorf("failed to read response: %v", err)
	}

	// 解析响应
	var result TurnstileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		Error("Failed to parse Turnstile response: %v, body: %s", err, string(body))
		return false, fmt.Errorf("failed to parse response: %v", err)
	}

	// 检查验证结果
	if !result.Success {
		Warn("Turnstile verification failed, error code: %v", result.ErrorCodes)
		return false, nil
	}

	Debug("Turnstile verification succeeded, hostname: %s, challenge_ts: %s", result.Hostname, result.ChallengeTs)
	return true, nil
}
