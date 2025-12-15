package middlewares

import (
	"errors"
	"fmt"
	"strings"
	"sublink/config"
	"sublink/models"
	"sublink/utils"

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
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthToken 验证token中间件
func AuthToken(c *gin.Context) {
	// 检查api key
	accessKey := c.GetHeader("X-API-Key")

	if accessKey != "" {
		username, bool, err := validApiKey(accessKey)
		if err != nil || !bool {
			utils.Forbidden(c, err.Error())
			c.Abort()
			return
		}
		c.Set("username", username)
		c.Next()
		return
	}

	token := c.Request.Header.Get("Authorization")
	if token == "" {
		token = c.Query("token")
	}
	if token == "" {
		utils.Forbidden(c, "请求未携带token")
		c.Abort()
		return
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		utils.Forbidden(c, "token格式错误")
		c.Abort()
		return
	}
	// 去掉Bearer前缀
	token = strings.Replace(token, "Bearer ", "", -1)
	mc, err := ParseToken(token)
	if err != nil {
		utils.Forbidden(c, err.Error())
		c.Abort()
		return
	}
	c.Set("username", mc.Username)
	c.Next()
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*JwtClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (i interface{}, err error) {
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

func validApiKey(apiKey string) (string, bool, error) {

	// 快速格式验证
	parts := strings.Split(apiKey, "_")
	if len(parts) != 3 {
		return "", false, fmt.Errorf("API Key格式错误")
	}

	encryptionKey := config.GetAPIEncryptionKey()
	if encryptionKey == "" {
		// 回退到旧的配置读取方式（兼容性）
		encryptionKey = models.ReadConfig().APIEncryptionKey
	}

	// 解密用户ID
	userID, err := utils.DecryptUserIDCompact(parts[1], []byte(encryptionKey))
	if err != nil {
		return "", false, fmt.Errorf("解密用户ID失败: %w", err)
	}

	// 数据库查询
	keys, err := models.FindValidAccessKeys(userID)
	if err != nil {
		return "", false, fmt.Errorf("查询Access Key失败: %w", err)
	}

	// bcrypt验证
	for _, key := range keys {
		if key.VerifyKey(apiKey) {

			return key.Username, true, nil
		}
	}

	return "", false, fmt.Errorf("无效的API Key")
}
