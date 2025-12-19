package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
func VerifyTurnstile(token, secretKey, remoteIP string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("turnstile token 为空")
	}
	if secretKey == "" {
		return false, fmt.Errorf("turnstile secret key 未配置")
	}

	// 构建请求数据
	formData := url.Values{}
	formData.Set("secret", secretKey)
	formData.Set("response", token)
	if remoteIP != "" {
		formData.Set("remoteip", remoteIP)
	}

	// 创建 HTTP 客户端（带超时）
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送 POST 请求
	resp, err := client.PostForm(TurnstileVerifyURL, formData)
	if err != nil {
		Error("Turnstile 验证请求失败: %v", err)
		return false, fmt.Errorf("验证请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Error("读取 Turnstile 响应失败: %v", err)
		return false, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var result TurnstileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		Error("解析 Turnstile 响应失败: %v, body: %s", err, string(body))
		return false, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查验证结果
	if !result.Success {
		Warn("Turnstile 验证失败, 错误代码: %v", result.ErrorCodes)
		return false, nil
	}

	Debug("Turnstile 验证成功, hostname: %s, challenge_ts: %s", result.Hostname, result.ChallengeTs)
	return true, nil
}
