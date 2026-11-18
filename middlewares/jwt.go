package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sublink/config"
	"sublink/models"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// getJwtSecret 获取 JWT 密钥（动态获取，支持配置热更新）
func getJwtSecret() []byte {
	secret := config.GetJwtSecret()
	if secret == "" {
		// 回退到旧的配置读取方式（兼容性）
		secret = models.ReadConfig().JwtSecret
	}
	return []byte(secret)
}

// JwtClaims jwt声明
type JwtClaims struct {
	Username       string `json:"username"`
	CredentialSign string `json:"cs"` // 凭证签名，用于密码/用户名变更后使JWT失效
	jwt.RegisteredClaims
}

// AuthToken 验证token中间件
func AuthToken(c *gin.Context) {
	// 检查api key
	accessKey := c.GetHeader("X-API-Key")

	if accessKey != "" {
		username, bool, err := validApiKey(accessKey, c.ClientIP())
		if err != nil || !bool {
			if err != nil {
				// Keep decryption and lookup details in server logs, never in the response.
				utils.Debug("API key authentication failed: %v", err)
			}
			utils.Forbidden(c, "Invalid API key")
			c.Abort()
			return
		}
		c.Set("username", username)
		c.Next()
		return
	}

	token := c.Request.Header.Get("Authorization")
	if token == "" {
		// EventSource cannot set headers, so the SSE stream is the only endpoint
		// that may authenticate with a query-string token.
		if !isSSETokenRequest(c) {
			utils.Forbidden(c, "Missing token")
			c.Abort()
			return
		}
		token = c.Query("token")
	}
	if token == "" {
		utils.Forbidden(c, "Missing token")
		c.Abort()
		return
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		utils.Forbidden(c, "Malformed token")
		c.Abort()
		return
	}
	// 去掉Bearer前缀
	token = strings.ReplaceAll(token, "Bearer ", "")
	mc, err := ParseToken(token)
	if err != nil {
		utils.Debug("JWT validation failed: %v", err)
		// 对SSE请求特殊处理：发送认证失败事件
		if isSSETokenRequest(c) {
			sendSSEAuthError(c, "Invalid or expired token")
			return
		}
		utils.Forbidden(c, "Invalid or expired token")
		c.Abort()
		return
	}
	// 验证凭证签名，用于检测密码或用户名是否已变更
	if !models.VerifyCredentialSign(mc.Username, mc.CredentialSign) {
		errMsg := "Session expired; please sign in again"
		// 对SSE请求特殊处理：发送认证失败事件
		if isSSETokenRequest(c) {
			sendSSEAuthError(c, errMsg)
			return
		}
		utils.Forbidden(c, errMsg)
		c.Abort()
		return
	}
	c.Set("username", mc.Username)
	c.Next()
}

// isSSETokenRequest reports whether the request targets the SSE stream.
// Query-string tokens are limited to it because EventSource cannot set headers.
func isSSETokenRequest(c *gin.Context) bool {
	return c.Request.Method == http.MethodGet && strings.HasSuffix(c.Request.URL.Path, "/api/sse")
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*JwtClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (i any, err error) {
		return getJwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// sendSSEAuthError 向SSE客户端发送认证失败事件
// 用于在SSE连接认证失败时，先发送一个auth_error事件，让客户端可以正确处理登出
func sendSSEAuthError(c *gin.Context, message string) {
	// 设置SSE响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// 构造认证失败事件数据
	eventData := fmt.Sprintf(`{"message":"%s","code":"AUTH_EXPIRED"}`, message)
	sseMessage := fmt.Sprintf("event: auth_error\ndata: %s\n\n", eventData)

	// 发送事件
	_, _ = c.Writer.WriteString(sseMessage)
	c.Writer.Flush()

	// 终止请求处理
	c.Abort()
}

func validApiKey(apiKey string, clientIP string) (string, bool, error) {

	// 快速格式验证
	parts := strings.Split(apiKey, "_")
	if len(parts) != 3 {
		return "", false, fmt.Errorf("malformed API key")
	}

	encryptionKey := config.GetAPIEncryptionKey()
	if encryptionKey == "" {
		// 回退到旧的配置读取方式（兼容性）
		encryptionKey = models.ReadConfig().APIEncryptionKey
	}

	// 解密用户ID
	userID, err := utils.DecryptUserIDCompact(parts[1], []byte(encryptionKey))
	if err != nil {
		return "", false, fmt.Errorf("failed to decrypt user ID: %w", err)
	}

	// 数据库查询
	keys, err := models.FindValidAccessKeys(userID)
	if err != nil {
		return "", false, fmt.Errorf("failed to look up access key: %w", err)
	}

	// bcrypt验证
	for _, key := range keys {
		if key.VerifyKey(apiKey) {
			// 记录最近使用时间与来源 IP（内部已做节流与异步落库）
			key.TouchLastUsed(clientIP, time.Now())
			return key.Username, true, nil
		}
	}

	return "", false, fmt.Errorf("invalid API key")
}
