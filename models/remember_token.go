package models

import (
	"crypto/rand"
	"encoding/hex"
	"sublink/database"
	"sublink/utils"
	"time"

	"gorm.io/gorm"
)

// RememberToken 记住密码令牌模型，支持多设备登录
type RememberToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    int       `gorm:"index;not null"`       // 关联用户ID
	Token     string    `gorm:"size:128;uniqueIndex"` // 令牌 (128位随机字符串)
	ExpiresAt time.Time `gorm:"index"`                // 过期时间
	UserAgent string    `gorm:"size:512"`             // 浏览器信息 (可选，用于识别设备)
	CreatedAt time.Time `gorm:"autoCreateTime"`       // 创建时间
}

// TableName 指定表名
func (RememberToken) TableName() string {
	return "remember_tokens"
}

// GenerateRememberToken 为用户生成新的记住密码令牌
// 每个用户最多保存 10 个令牌，超过时删除最早创建的
func GenerateRememberToken(userID int, userAgent string) (string, error) {
	// 生成64字节的随机令牌
	tokenBytes := make([]byte, 64)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	// 创建令牌记录，30天有效期
	rememberToken := &RememberToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		UserAgent: userAgent,
	}

	if err := database.DB.Create(rememberToken).Error; err != nil {
		return "", err
	}

	// 异步清理：删除过期令牌 + 限制最大数量
	go cleanUserTokens(userID)

	return token, nil
}

// VerifyAndGetUserByToken 验证令牌并返回用户
func VerifyAndGetUserByToken(token string) (*User, error) {
	if token == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var rememberToken RememberToken
	// 先查找 token 是否存在（不检查过期时间）
	err := database.DB.Where("token = ?", token).First(&rememberToken).Error
	if err != nil {
		return nil, err
	}

	// 检查是否过期
	if rememberToken.ExpiresAt.Before(time.Now()) {
		// Token 已过期，删除它
		database.DB.Delete(&rememberToken)
		return nil, gorm.ErrRecordNotFound
	}

	// 获取用户信息
	var user User
	if err := database.DB.First(&user, rememberToken.UserID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteRememberToken 删除指定的令牌 (用户登出时调用)
func DeleteRememberToken(token string) error {
	return database.DB.Where("token = ?", token).Delete(&RememberToken{}).Error
}

// DeleteUserRememberTokens 删除用户的所有令牌 (修改密码时调用)
func DeleteUserRememberTokens(userID int) error {
	return database.DB.Where("user_id = ?", userID).Delete(&RememberToken{}).Error
}

// cleanUserTokens 清理用户的令牌：删除过期的 + 限制最大数量为 10 个
func cleanUserTokens(userID int) {
	const maxTokensPerUser = 10

	// 1. 删除过期的令牌
	result := database.DB.Where("user_id = ? AND expires_at < ?", userID, time.Now()).Delete(&RememberToken{})
	if result.RowsAffected > 0 {
		utils.Info("清理了用户 %d 的 %d 个过期令牌", userID, result.RowsAffected)
	}

	// 2. 检查令牌数量，删除超出限制的（保留最新的 10 个）
	var count int64
	database.DB.Model(&RememberToken{}).Where("user_id = ?", userID).Count(&count)
	if count > maxTokensPerUser {
		// 查找需要删除的令牌（最早创建的）
		var tokensToDelete []RememberToken
		database.DB.Where("user_id = ?", userID).
			Order("created_at ASC").
			Limit(int(count - maxTokensPerUser)).
			Find(&tokensToDelete)

		for _, t := range tokensToDelete {
			database.DB.Delete(&t)
		}
		utils.Info("用户 %d 的令牌数量超过限制，删除了 %d 个最早的令牌", userID, len(tokensToDelete))
	}
}

// CleanAllExpiredTokens 清理所有过期的令牌 (系统启动时调用)
func CleanAllExpiredTokens() {
	result := database.DB.Where("expires_at < ?", time.Now()).Delete(&RememberToken{})
	if result.RowsAffected > 0 {
		utils.Info("清理了 %d 个过期的记住密码令牌", result.RowsAffected)
	}
}
