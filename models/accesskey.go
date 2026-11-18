package models

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// AccessKeyDescriptionMaxLength 描述的最大长度（与 access_keys.description 的列长度一致）
const AccessKeyDescriptionMaxLength = 255

// accessKeyLastUsedWriteInterval 最近使用记录的落库间隔，避免每个 API 请求都写库
const accessKeyLastUsedWriteInterval = 5 * time.Minute

var (
	// ErrAccessKeyDescriptionEmpty 描述为空
	ErrAccessKeyDescriptionEmpty = errors.New("description is required")
	// ErrAccessKeyDescriptionTooLong 描述超过最大长度
	ErrAccessKeyDescriptionTooLong = fmt.Errorf("description cannot exceed %d characters", AccessKeyDescriptionMaxLength)
)

type AccessKey struct {
	ID            int        `gorm:"primaryKey"`
	UserID        int        `gorm:"not null;index"`
	Username      string     `gorm:"size:100;not null;index"`
	AccessKeyHash string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	CreatedAt     time.Time  `gorm:"autoCreateTime"`
	ExpiredAt     *time.Time `gorm:"index"`
	Description   string     `gorm:"type:varchar(255)"`
	LastUsedAt    *time.Time
	LastUsedIP    string `gorm:"type:varchar(64)"`
}

// ValidateAccessKeyDescription 按字符数校验描述长度，避免多字节字符被按字节数误判
func ValidateAccessKeyDescription(description string) error {
	if utf8.RuneCountInString(description) > AccessKeyDescriptionMaxLength {
		return ErrAccessKeyDescriptionTooLong
	}
	return nil
}

// NormalizeAccessKeyDescription 去除首尾空白并校验长度，返回可直接存储的描述
func NormalizeAccessKeyDescription(description string) (string, error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return "", ErrAccessKeyDescriptionEmpty
	}
	if err := ValidateAccessKeyDescription(description); err != nil {
		return "", err
	}
	return description, nil
}

// accessKeyCache 使用新的泛型缓存
var accessKeyCache *cache.MapCache[int, AccessKey]

func init() {
	accessKeyCache = cache.NewMapCache(func(ak AccessKey) int { return ak.ID })
	accessKeyCache.AddIndex("userID", func(ak AccessKey) string { return strconv.Itoa(ak.UserID) })
}

// InitAccessKeyCache 初始化AccessKey缓存
func InitAccessKeyCache() error {
	utils.Info("Loading access keys into cache")
	var accessKeys []AccessKey
	if err := database.DB.Find(&accessKeys).Error; err != nil {
		return err
	}

	accessKeyCache.LoadAll(accessKeys)
	utils.Info("Access key cache initialized with %d keys", accessKeyCache.Count())

	cache.Manager.Register("accesskey", accessKeyCache)
	return nil
}

// Generate 保存 AccessKey (Write-Through)
func (accessKey *AccessKey) Generate() error {
	err := database.DB.Create(accessKey).Error
	if err != nil {
		return err
	}
	accessKeyCache.Set(accessKey.ID, *accessKey)
	return nil
}

// Update 更新描述与过期时间 (Write-Through)
func (accessKey *AccessKey) Update(description string, expiredAt *time.Time) error {
	var existing AccessKey
	if err := database.DB.First(&existing, accessKey.ID).Error; err != nil {
		return err
	}
	err := database.DB.Model(&AccessKey{}).Where("id = ?", accessKey.ID).
		Updates(map[string]any{"description": description, "expired_at": expiredAt}).Error
	if err != nil {
		return err
	}
	existing.Description = description
	existing.ExpiredAt = expiredAt
	accessKeyCache.Set(existing.ID, existing)
	*accessKey = existing
	return nil
}

// accessKeyLastUsedWrites 记录各密钥上次写入最近使用信息的时间，用于节流
var accessKeyLastUsedWrites sync.Map

// TouchLastUsed 记录密钥最近一次使用的时间与来源 IP。
// 写入按密钥节流并放到后台执行，避免高频 API 调用把 SQLite 写成瓶颈。
func (accessKey *AccessKey) TouchLastUsed(ip string, now time.Time) {
	if last, ok := accessKeyLastUsedWrites.Load(accessKey.ID); ok {
		if lastWrite, ok := last.(time.Time); ok && now.Sub(lastWrite) < accessKeyLastUsedWriteInterval {
			return
		}
	}
	accessKeyLastUsedWrites.Store(accessKey.ID, now)

	go func(id int, clientIP string, at time.Time) {
		err := database.DB.Model(&AccessKey{}).Where("id = ?", id).
			Updates(map[string]any{"last_used_at": at, "last_used_ip": clientIP}).Error
		if err != nil {
			utils.Error("Failed to update access key last-used record, ID: %d, error: %v", id, err)
			return
		}
		if cached, ok := accessKeyCache.Get(id); ok {
			cached.LastUsedAt = &at
			cached.LastUsedIP = clientIP
			accessKeyCache.Set(id, cached)
		}
	}(accessKey.ID, ip, now)
}

// FindValidAccessKeys 查找未过期的 AccessKey
func FindValidAccessKeys(userID int) ([]AccessKey, error) {
	// 使用二级索引获取用户的所有 key
	allKeys := accessKeyCache.GetByIndex("userID", strconv.Itoa(userID))
	now := time.Now()

	validKeys := make([]AccessKey, 0)
	for _, key := range allKeys {
		if key.ExpiredAt == nil || key.ExpiredAt.After(now) {
			validKeys = append(validKeys, key)
		}
	}
	return validKeys, nil
}

// FindValidAccessKeysPaginated 分页查找未过期的 AccessKey
func FindValidAccessKeysPaginated(userID, page, pageSize int) ([]AccessKey, int64, error) {
	validKeys, _ := FindValidAccessKeys(userID)
	total := int64(len(validKeys))

	if page <= 0 || pageSize <= 0 {
		return validKeys, total, nil
	}

	offset := (page - 1) * pageSize
	if offset >= len(validKeys) {
		return []AccessKey{}, total, nil
	}

	end := offset + pageSize
	if end > len(validKeys) {
		end = len(validKeys)
	}

	return validKeys[offset:end], total, nil
}

// Delete 删除 AccessKey (Write-Through)
func (accessKey *AccessKey) Delete() error {
	// 先从数据库获取完整的 AccessKey 信息
	var fullAccessKey AccessKey
	err := database.DB.First(&fullAccessKey, accessKey.ID).Error
	if err != nil {
		return fmt.Errorf("failed to get access key info: %w", err)
	}

	// 删除数据库记录
	err = database.DB.Unscoped().Delete(accessKey).Error
	if err != nil {
		return err
	}

	// 更新缓存
	accessKeyCache.Delete(accessKey.ID)
	return nil
}

// GenerateAPIKey 生成一个新的 API Key,单用户系统直接全随机不编码用户信息
func (accessKey *AccessKey) GenerateAPIKey() (string, error) {
	// 优先使用 config 包获取加密密钥
	cfg := ReadConfig()
	encryptionKey := cfg.APIEncryptionKey

	encryptedID, err := utils.EncryptUserIDCompact(accessKey.UserID, []byte(encryptionKey))
	if err != nil {
		utils.Error("Failed to encrypt user ID: %v", err)
		return "", fmt.Errorf("failed to encrypt user ID: %w", err)
	}
	randomBytes := make([]byte, 18)
	_, err = rand.Read(randomBytes)
	if err != nil {
		utils.Error("Failed to generate random data: %v", err)
		return "", fmt.Errorf("failed to generate random data: %w", err)
	}

	randomHex := hex.EncodeToString(randomBytes)

	apiKey := fmt.Sprintf("sp_%s_%s", encryptedID, randomHex)

	hashedKey, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		utils.Error("Failed to hash API key: %v", err)
		return "", fmt.Errorf("failed to hash API key: %w", err)
	}
	accessKey.AccessKeyHash = string(hashedKey)

	return apiKey, nil
}

// VerifyKey 验证提供的 API Key 是否与存储的哈希匹配
func (accessKey *AccessKey) VerifyKey(providedKey string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(accessKey.AccessKeyHash), []byte(providedKey))
	return err == nil
}

// CleanupExpiredAccessKeys 清理过期的 AccessKey
func CleanupExpiredAccessKeys() error {
	utils.Info("Starting expired access key cleanup")

	// 查找所有过期的 AccessKey
	var expiredKeys []AccessKey
	err := database.DB.Where("expired_at IS NOT NULL AND expired_at < ?", time.Now()).Find(&expiredKeys).Error
	if err != nil {
		utils.Error("Failed to query expired access keys: %v", err)
		return fmt.Errorf("failed to query expired access keys: %w", err)
	}

	utils.Info("Found %d expired access keys to clean up", len(expiredKeys))

	// 批量删除过期的 AccessKey
	for _, key := range expiredKeys {
		err := key.Delete()
		if err != nil {
			utils.Error("Failed to delete expired access key, ID: %d, error: %v", key.ID, err)
			continue
		}
		utils.Info("Deleted expired access key, ID: %d, Username: %s", key.ID, key.Username)
	}

	utils.Info("Expired access key cleanup completed, processed %d", len(expiredKeys))
	return nil
}

// StartAccessKeyCleanupScheduler 启动 AccessKey 清理定时任务
func StartAccessKeyCleanupScheduler() {
	utils.Info("Starting access key cleanup cron job")
	// 每小时执行一次清理
	ticker := time.NewTicker(1 * time.Hour)

	go func() {
		defer ticker.Stop()
		for range ticker.C {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						utils.Error("Access key cleanup task panic: %v", r)
					}
				}()
				if err := CleanupExpiredAccessKeys(); err != nil {
					utils.Error("Access key cleanup failed: %v", err)
				}
			}()
		}
	}()

	// 启动时立即执行一次清理
	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Error("Initial access key cleanup task panic: %v", r)
			}
		}()
		if err := CleanupExpiredAccessKeys(); err != nil {
			utils.Error("Initial access key cleanup failed: %v", err)
		}
	}()
}
