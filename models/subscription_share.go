package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"time"
)

// 过期类型常量
const (
	ExpireTypeNever    = 0 // 永不过期
	ExpireTypeDays     = 1 // 按天数过期
	ExpireTypeDateTime = 2 // 指定时间过期
)

// SubscriptionShare 订阅分享表
type SubscriptionShare struct {
	ID             int       `gorm:"primaryKey" json:"id"`
	SubscriptionID int       `gorm:"index" json:"subscription_id"`        // 关联订阅ID
	Token          string    `gorm:"uniqueIndex;size:64" json:"token"`    // 分享token（支持自定义或自动生成）
	Name           string    `gorm:"size:100" json:"name"`                // 分享名称/备注
	ExpireType     int       `gorm:"default:0" json:"expire_type"`        // 过期类型
	ExpireDays     int       `gorm:"default:0" json:"expire_days"`        // 过期天数
	ExpireAt       time.Time `gorm:"type:datetime" json:"expire_at"`      // 过期时间
	IsLegacy       bool      `gorm:"default:false" json:"is_legacy"`      // 是否为迁移的老链接
	Enabled        bool      `gorm:"default:true" json:"enabled"`         // 是否启用
	AccessCount    int       `gorm:"default:0" json:"access_count"`       // 访问次数
	LastAccessAt   time.Time `gorm:"type:datetime" json:"last_access_at"` // 最后访问时间
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// subscriptionShareCache 使用泛型缓存
var subscriptionShareCache *cache.MapCache[int, SubscriptionShare]

func init() {
	subscriptionShareCache = cache.NewMapCache(func(s SubscriptionShare) int { return s.ID })
	subscriptionShareCache.AddIndex("token", func(s SubscriptionShare) string { return s.Token })
	subscriptionShareCache.AddIndex("subscriptionID", func(s SubscriptionShare) string { return strconv.Itoa(s.SubscriptionID) })
}

// InitSubscriptionShareCache 初始化订阅分享缓存
func InitSubscriptionShareCache() error {
	utils.Info("开始加载订阅分享到缓存")
	var shares []SubscriptionShare
	if err := database.DB.Find(&shares).Error; err != nil {
		return err
	}

	subscriptionShareCache.LoadAll(shares)
	utils.Info("订阅分享缓存初始化完成，共加载 %d 条记录", subscriptionShareCache.Count())

	cache.Manager.Register("subscription_shares", subscriptionShareCache)
	return nil
}

// GenerateToken 生成随机 token
func GenerateToken() (string, error) {
	bytes := make([]byte, 16) // 16字节 = 32个十六进制字符
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsTokenExists 检查 token 是否已存在
func IsTokenExists(token string, excludeID int) bool {
	shares := subscriptionShareCache.GetByIndex("token", token)
	for _, s := range shares {
		if s.ID != excludeID {
			return true
		}
	}
	return false
}

// Add 添加分享 (Write-Through)
func (s *SubscriptionShare) Add() error {
	// 如果没有提供 token，自动生成
	if s.Token == "" {
		token, err := GenerateToken()
		if err != nil {
			return err
		}
		s.Token = token
	}

	// 检查 token 唯一性
	if IsTokenExists(s.Token, 0) {
		return fmt.Errorf("Token 已被使用，请更换")
	}

	err := database.DB.Create(s).Error
	if err != nil {
		return err
	}
	subscriptionShareCache.Set(s.ID, *s)
	return nil
}

// Update 更新分享 (Write-Through)
func (s *SubscriptionShare) Update() error {
	// 检查 token 唯一性（排除自己）
	if IsTokenExists(s.Token, s.ID) {
		return fmt.Errorf("Token 已被使用，请更换")
	}

	err := database.DB.Model(s).Updates(map[string]interface{}{
		"name":        s.Name,
		"token":       s.Token,
		"expire_type": s.ExpireType,
		"expire_days": s.ExpireDays,
		"expire_at":   s.ExpireAt,
		"enabled":     s.Enabled,
	}).Error
	if err != nil {
		return err
	}

	// 更新缓存
	var updated SubscriptionShare
	if err := database.DB.First(&updated, s.ID).Error; err == nil {
		subscriptionShareCache.Set(updated.ID, updated)
	}
	return nil
}

// Delete 删除分享 (Write-Through)
func (s *SubscriptionShare) Delete() error {
	err := database.DB.Delete(s).Error
	if err != nil {
		return err
	}
	subscriptionShareCache.Delete(s.ID)
	return nil
}

// Find 根据 ID 查找
func (s *SubscriptionShare) Find() error {
	if cached, ok := subscriptionShareCache.Get(s.ID); ok {
		*s = cached
		return nil
	}
	return database.DB.First(s, s.ID).Error
}

// GetByToken 根据 token 查找分享
func GetSubscriptionShareByToken(token string) (*SubscriptionShare, error) {
	shares := subscriptionShareCache.GetByIndex("token", token)
	if len(shares) > 0 {
		return &shares[0], nil
	}
	// 缓存未命中，从数据库查
	var share SubscriptionShare
	if err := database.DB.Where("token = ?", token).First(&share).Error; err != nil {
		return nil, err
	}
	subscriptionShareCache.Set(share.ID, share)
	return &share, nil
}

// GetSharesBySubscriptionID 获取订阅的所有分享
func GetSharesBySubscriptionID(subID int) []SubscriptionShare {
	return subscriptionShareCache.GetByIndex("subscriptionID", strconv.Itoa(subID))
}

// GetDefaultShareForSubscription 获取订阅的默认分享链接
// 优先返回老的迁移链接（IsLegacy=true），否则返回第一个启用的分享
func GetDefaultShareForSubscription(subID int) (*SubscriptionShare, error) {
	shares := GetSharesBySubscriptionID(subID)
	if len(shares) == 0 {
		return nil, fmt.Errorf("该订阅没有分享链接")
	}

	// 优先找 legacy 链接
	for _, s := range shares {
		if s.IsLegacy && s.Enabled {
			return &s, nil
		}
	}

	// 否则返回第一个启用的
	for _, s := range shares {
		if s.Enabled {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("该订阅没有可用的分享链接")
}

// IsExpired 检查分享是否已过期
func (s *SubscriptionShare) IsExpired() bool {
	if !s.Enabled {
		return true
	}

	switch s.ExpireType {
	case ExpireTypeNever:
		return false
	case ExpireTypeDays:
		if s.ExpireDays <= 0 {
			return false
		}
		expireTime := s.CreatedAt.AddDate(0, 0, s.ExpireDays)
		return time.Now().After(expireTime)
	case ExpireTypeDateTime:
		if s.ExpireAt.IsZero() {
			return false
		}
		return time.Now().After(s.ExpireAt)
	default:
		return false
	}
}

// RecordAccess 记录一次访问
func (s *SubscriptionShare) RecordAccess() {
	s.AccessCount++
	s.LastAccessAt = time.Now()
	database.DB.Model(s).Updates(map[string]interface{}{
		"access_count":   s.AccessCount,
		"last_access_at": s.LastAccessAt,
	})
	subscriptionShareCache.Set(s.ID, *s)
}

// List 获取所有分享列表
func (s *SubscriptionShare) List() ([]SubscriptionShare, error) {
	return subscriptionShareCache.GetAll(), nil
}
