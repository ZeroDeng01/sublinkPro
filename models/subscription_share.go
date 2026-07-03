package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"sync"
	"time"

	"gorm.io/gorm"
)

// 过期类型常量
const (
	ExpireTypeNever    = 0 // 永不过期
	ExpireTypeDays     = 1 // 按天数过期
	ExpireTypeDateTime = 2 // 指定时间过期
)

// SubscriptionShare 订阅分享表
type SubscriptionShare struct {
	ID             int        `gorm:"primaryKey" json:"id"`
	SubscriptionID int        `gorm:"index" json:"subscription_id"`     // 关联订阅ID
	Token          string     `gorm:"uniqueIndex;size:64" json:"token"` // 分享token（支持自定义或自动生成）
	Name           string     `gorm:"size:100" json:"name"`             // 分享名称/备注
	ExpireType     int        `gorm:"default:0" json:"expire_type"`     // 过期类型
	ExpireDays     int        `gorm:"default:0" json:"expire_days"`     // 过期天数
	ExpireAt       *time.Time `json:"expire_at"`                        // 过期时间
	IsLegacy       bool       `gorm:"default:false" json:"is_legacy"`   // 是否为迁移的老链接
	Enabled        bool       `gorm:"default:true" json:"enabled"`      // 是否启用
	AccessCount    int        `gorm:"default:0" json:"access_count"`    // 访问次数
	LastAccessAt   *time.Time `json:"last_access_at"`                   // 最后访问时间
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
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

func normalizeOptionalTime(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	return new(*value)
}

func (s *SubscriptionShare) normalizeOptionalFields() {
	if s.ExpireType != ExpireTypeDateTime {
		s.ExpireAt = nil
	} else {
		s.ExpireAt = normalizeOptionalTime(s.ExpireAt)
	}

	if s.AccessCount <= 0 {
		s.LastAccessAt = nil
	} else {
		s.LastAccessAt = normalizeOptionalTime(s.LastAccessAt)
	}
}

// CreateDefaultShareForSubscription 为订阅创建默认分享链接
// 创建一个永不过期、启用状态的默认分享链接，标记为 IsLegacy=true
func CreateDefaultShareForSubscription(subscriptionID int) error {
	share := &SubscriptionShare{
		SubscriptionID: subscriptionID,
		Name:           "默认分享链接",
		ExpireType:     ExpireTypeNever,
		IsLegacy:       true,
		Enabled:        true,
	}
	return share.Add()
}

// Add 添加分享 (Write-Through)
func (s *SubscriptionShare) Add() error {
	s.normalizeOptionalFields()

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
		return fmt.Errorf("token 已被使用，请更换")
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
	s.normalizeOptionalFields()

	// 检查 token 唯一性（排除自己）
	if IsTokenExists(s.Token, s.ID) {
		return fmt.Errorf("token 已被使用，请更换")
	}

	err := database.DB.Model(s).Updates(map[string]any{
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
func GetSharesBySubscriptionID(subID int, keyword ...string) []SubscriptionShare {
	shares := subscriptionShareCache.GetByIndex("subscriptionID", strconv.Itoa(subID))

	// 如果提供了搜索关键词，进行过滤
	// 规则：名称使用模糊搜索，token使用精确匹配
	if len(keyword) > 0 && keyword[0] != "" {
		kw := keyword[0]
		kwLower := strings.ToLower(kw)
		filtered := make([]SubscriptionShare, 0)
		for _, share := range shares {
			// 名称：模糊搜索（不区分大小写）
			nameMatch := strings.Contains(strings.ToLower(share.Name), kwLower)
			// Token：精确匹配（区分大小写）
			tokenMatch := share.Token == kw

			if nameMatch || tokenMatch {
				filtered = append(filtered, share)
			}
		}
		return filtered
	}

	return shares
}

// GetShareIDsByIP 根据IP地址获取访问过的分享ID列表
func GetShareIDsByIP(ip string) []int {
	logs := subLogsCache.GetAll()
	shareIDMap := make(map[int]bool)

	for _, log := range logs {
		if log.IP == ip && log.ShareID > 0 {
			shareIDMap[log.ShareID] = true
		}
	}

	shareIDs := make([]int, 0, len(shareIDMap))
	for id := range shareIDMap {
		shareIDs = append(shareIDs, id)
	}
	return shareIDs
}

// GetSharesBySubscriptionIDPaginated 获取订阅的分享列表（分页，支持搜索、IP筛选、排序）
func GetSharesBySubscriptionIDPaginated(subID, page, pageSize int, keyword, ipFilter, sortBy, sortOrder string) ([]SubscriptionShare, int, error) {
	// 先从缓存获取该订阅的所有分享
	allShares := subscriptionShareCache.GetByIndex("subscriptionID", strconv.Itoa(subID))

	// IP过滤优先级最高
	if ipFilter != "" {
		shareIDs := GetShareIDsByIP(ipFilter)
		shareIDSet := make(map[int]bool)
		for _, id := range shareIDs {
			shareIDSet[id] = true
		}

		filtered := make([]SubscriptionShare, 0)
		for _, share := range allShares {
			if shareIDSet[share.ID] {
				filtered = append(filtered, share)
			}
		}
		allShares = filtered
	}

	// 如果提供了搜索关键词，进行过滤
	// 规则：名称使用模糊搜索，token使用精确匹配
	if keyword != "" {
		kwLower := strings.ToLower(keyword)
		filtered := make([]SubscriptionShare, 0)
		for _, share := range allShares {
			// 名称：模糊搜索（不区分大小写）
			nameMatch := strings.Contains(strings.ToLower(share.Name), kwLower)
			// Token：精确匹配（区分大小写）
			tokenMatch := share.Token == keyword

			if nameMatch || tokenMatch {
				filtered = append(filtered, share)
			}
		}
		allShares = filtered
	}

	// 排序
	if sortBy == "access_count" {
		sort.Slice(allShares, func(i, j int) bool {
			if sortOrder == "asc" {
				return allShares[i].AccessCount < allShares[j].AccessCount
			}
			return allShares[i].AccessCount > allShares[j].AccessCount
		})
	}

	total := len(allShares)

	// 计算分页
	offset := (page - 1) * pageSize
	if offset >= total {
		return []SubscriptionShare{}, total, nil
	}

	end := offset + pageSize
	if end > total {
		end = total
	}

	return allShares[offset:end], total, nil
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
		if s.ExpireAt == nil || s.ExpireAt.IsZero() {
			return false
		}
		return time.Now().After(*s.ExpireAt)
	default:
		return false
	}
}

// accessRecordWG 跟踪在飞的异步访问统计写入，便于测试拆库前与服务优雅关闭时等待其完成。
var accessRecordWG sync.WaitGroup

// WaitForPendingAccessRecords 阻塞直到所有异步访问统计写入完成。
// 测试在重置/关闭 database.DB 前调用；服务优雅关闭时亦可调用以避免丢失访问计数。
func WaitForPendingAccessRecords() {
	accessRecordWG.Wait()
}

// RecordAccess 记录一次访问，并使用数据库原子自增避免并发访问丢失计数。
func (s *SubscriptionShare) RecordAccess() {
	if s == nil || s.ID <= 0 {
		return
	}

	// 只读取一次全局句柄：异步 goroutine 可能与测试/关闭流程并发，
	// 若全局 database.DB 中途被重置为 nil，直接跳过，避免空指针解引用。
	db := database.DB
	if db == nil {
		return
	}

	if err := db.Model(&SubscriptionShare{}).
		Where("id = ?", s.ID).
		Updates(map[string]any{
			"access_count":   gorm.Expr("access_count + ?", 1),
			"last_access_at": new(time.Now()),
		}).Error; err != nil {
		utils.Warn("记录订阅分享访问失败: %v", err)
		return
	}

	var updated SubscriptionShare
	if err := db.First(&updated, s.ID).Error; err != nil {
		utils.Warn("刷新订阅分享访问缓存失败: %v", err)
		return
	}
	*s = updated
	subscriptionShareCache.Set(updated.ID, updated)
}

// RecordAccessAsync 异步记录一次访问，避免订阅拉取热路径等待数据库写入。
func (s *SubscriptionShare) RecordAccessAsync() {
	if s == nil || s.ID <= 0 {
		return
	}

	shareID := s.ID
	accessRecordWG.Add(1)
	go func() {
		defer accessRecordWG.Done()
		share := SubscriptionShare{ID: shareID}
		share.RecordAccess()
	}()
}

// List 获取所有分享列表
func (s *SubscriptionShare) List() ([]SubscriptionShare, error) {
	return subscriptionShareCache.GetAll(), nil
}
