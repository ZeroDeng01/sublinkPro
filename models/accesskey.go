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

	"golang.org/x/crypto/bcrypt"
)

type AccessKey struct {
	ID            int        `gorm:"primaryKey"`
	UserID        int        `gorm:"not null;index"`
	Username      string     `gorm:"not null;index"`
	AccessKeyHash string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	CreatedAt     time.Time  `gorm:""`
	ExpiredAt     *time.Time `gorm:"index"`
	Description   string     `gorm:"type:varchar(255)"`
}

// accessKeyCache 使用新的泛型缓存
var accessKeyCache *cache.MapCache[int, AccessKey]

func init() {
	accessKeyCache = cache.NewMapCache(func(ak AccessKey) int { return ak.ID })
	accessKeyCache.AddIndex("userID", func(ak AccessKey) string { return strconv.Itoa(ak.UserID) })
}

// InitAccessKeyCache 初始化AccessKey缓存
func InitAccessKeyCache() error {
	utils.Info("开始加载AccessKey到缓存")
	var accessKeys []AccessKey
	if err := database.DB.Find(&accessKeys).Error; err != nil {
		return err
	}

	accessKeyCache.LoadAll(accessKeys)
	utils.Info("AccessKey缓存初始化完成，共加载 %d 个AccessKey", accessKeyCache.Count())

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
		return fmt.Errorf("获取 AccessKey 信息失败: %w", err)
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
	encryptionKey := ""
	cfg := ReadConfig()
	encryptionKey = cfg.APIEncryptionKey

	encryptedID, err := utils.EncryptUserIDCompact(accessKey.UserID, []byte(encryptionKey))
	if err != nil {
		utils.Error("加密用户ID失败: %v", err)
		return "", fmt.Errorf("加密用户ID失败: %w", err)
	}
	randomBytes := make([]byte, 18)
	_, err = rand.Read(randomBytes)
	if err != nil {
		utils.Error("生成随机数据失败: %v", err)
		return "", fmt.Errorf("生成随机数据失败: %w", err)
	}

	randomHex := hex.EncodeToString(randomBytes)

	apiKey := fmt.Sprintf("subE_%s_%s", encryptedID, randomHex)

	hashedKey, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		utils.Error("哈希API密钥失败: %v", err)
		return "", fmt.Errorf("哈希API密钥失败: %w", err)
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
	utils.Info("开始清理过期的 AccessKey")

	// 查找所有过期的 AccessKey
	var expiredKeys []AccessKey
	err := database.DB.Where("expired_at IS NOT NULL AND expired_at < ?", time.Now()).Find(&expiredKeys).Error
	if err != nil {
		utils.Error("查询过期 AccessKey 失败: %v", err)
		return fmt.Errorf("查询过期 AccessKey 失败: %w", err)
	}

	utils.Info("发现 %d 个过期的 AccessKey，准备清理", len(expiredKeys))

	// 批量删除过期的 AccessKey
	for _, key := range expiredKeys {
		err := key.Delete()
		if err != nil {
			utils.Error("删除过期 AccessKey 失败，ID: %d, 错误: %v", key.ID, err)
			continue
		}
		utils.Info("成功删除过期 AccessKey，ID: %d, Username: %s", key.ID, key.Username)
	}

	utils.Info("过期 AccessKey 清理完成，共处理 %d 个", len(expiredKeys))
	return nil
}

// StartAccessKeyCleanupScheduler 启动 AccessKey 清理定时任务
func StartAccessKeyCleanupScheduler() {
	utils.Info("启动 AccessKey 清理定时任务")
	// 每小时执行一次清理
	ticker := time.NewTicker(1 * time.Hour)

	go func() {
		defer ticker.Stop()
		for range ticker.C {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						utils.Error("AccessKey 清理任务异常: %v", r)
					}
				}()
				CleanupExpiredAccessKeys()
			}()
		}
	}()

	// 启动时立即执行一次清理
	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Error("AccessKey 初始清理任务异常: %v", r)
			}
		}()
		CleanupExpiredAccessKeys()
	}()
}
