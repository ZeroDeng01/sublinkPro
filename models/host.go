package models

import (
	"bufio"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"sync"
	"time"

	"gorm.io/gorm/clause"
)

// Host 自定义 Host 映射模型
type Host struct {
	ID        int        `json:"id" gorm:"primaryKey"`
	Hostname  string     `json:"hostname" gorm:"size:255;uniqueIndex;not null"` // 域名
	IP        string     `json:"ip" gorm:"size:45;not null"`                    // IP 地址 (支持 IPv6)
	Remark    string     `json:"remark" gorm:"size:255"`                        // 备注
	Source    string     `json:"source" gorm:"size:255;default:'手动添加'"`         // 来源: 手动添加/DNS服务器地址
	ExpireAt  *time.Time `json:"expireAt" gorm:"index"`                         // 过期时间，nil 表示永不过期
	Pinned    bool       `json:"pinned" gorm:"default:false"`                   // 是否固定，固定后不会被过期删除
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// hostCache 使用泛型缓存
var hostCache *cache.MapCache[int, Host]

// Host变更回调函数，用于通知外部模块（如mihomo resolver）同步数据
var (
	hostChangeCallbackMu sync.RWMutex
	hostChangeCallbacks  []func()
)

// RegisterHostChangeCallback 注册Host变更回调函数
// 当Host数据发生变更（增删改）时会调用已注册的回调
func RegisterHostChangeCallback(callback func()) {
	hostChangeCallbackMu.Lock()
	defer hostChangeCallbackMu.Unlock()
	hostChangeCallbacks = append(hostChangeCallbacks, callback)
}

// notifyHostChanged 通知所有注册的回调函数
func notifyHostChanged() {
	hostChangeCallbackMu.RLock()
	callbacks := make([]func(), len(hostChangeCallbacks))
	copy(callbacks, hostChangeCallbacks)
	hostChangeCallbackMu.RUnlock()

	for _, cb := range callbacks {
		cb()
	}
}

func init() {
	hostCache = cache.NewMapCache(func(h Host) int { return h.ID })
	hostCache.AddIndex("hostname", func(h Host) string { return h.Hostname })
}

// InitHostCache 初始化 Host 缓存
func InitHostCache() error {
	utils.Info("开始加载 Host 到缓存")
	var hosts []Host
	if err := database.DB.Find(&hosts).Error; err != nil {
		return err
	}

	hostCache.LoadAll(hosts)
	utils.Info("Host 缓存初始化完成，共加载 %d 条记录", hostCache.Count())

	cache.Manager.Register("host", hostCache)
	return nil
}

// ========== CRUD 方法 ==========

// Add 添加 Host (Write-Through)
func (h *Host) Add() error {
	// 检查 hostname 是否已存在
	if hosts := hostCache.GetByIndex("hostname", h.Hostname); len(hosts) > 0 {
		return fmt.Errorf("hostname '%s' 已存在", h.Hostname)
	}

	err := database.DB.Create(h).Error
	if err != nil {
		return err
	}
	hostCache.Set(h.ID, *h)
	notifyHostChanged() // 通知外部模块同步
	return nil
}

// Update 更新 Host (Write-Through)
func (h *Host) Update() error {
	// 检查 hostname 是否与其他记录冲突
	if hosts := hostCache.GetByIndex("hostname", h.Hostname); len(hosts) > 0 {
		for _, existing := range hosts {
			if existing.ID != h.ID {
				return fmt.Errorf("hostname '%s' 已被其他记录使用", h.Hostname)
			}
		}
	}

	err := database.DB.Model(h).Updates(map[string]interface{}{
		"hostname":   h.Hostname,
		"ip":         h.IP,
		"remark":     h.Remark,
		"updated_at": time.Now(),
	}).Error
	if err != nil {
		return err
	}
	// 从数据库读取完整数据后更新缓存
	var updated Host
	if err := database.DB.First(&updated, h.ID).Error; err == nil {
		hostCache.Set(h.ID, updated)
	}
	notifyHostChanged() // 通知外部模块同步
	return nil
}

// Delete 删除 Host (Write-Through)
func (h *Host) Delete() error {
	err := database.DB.Delete(h).Error
	if err != nil {
		return err
	}
	hostCache.Delete(h.ID)
	notifyHostChanged() // 通知外部模块同步
	return nil
}

// GetByID 根据 ID 获取 Host
func GetHostByID(id int) (*Host, error) {
	if host, ok := hostCache.Get(id); ok {
		return &host, nil
	}
	var host Host
	if err := database.DB.First(&host, id).Error; err != nil {
		return nil, err
	}
	hostCache.Set(host.ID, host)
	return &host, nil
}

// GetByHostname 根据 hostname 获取 Host
func GetHostByHostname(hostname string) (*Host, error) {
	if hosts := hostCache.GetByIndex("hostname", hostname); len(hosts) > 0 {
		return &hosts[0], nil
	}
	return nil, fmt.Errorf("host '%s' 不存在", hostname)
}

// List 获取所有 Host 列表
func (h *Host) List() ([]Host, error) {
	hosts := hostCache.GetAllSorted(func(a, b Host) bool {
		return a.ID < b.ID
	})
	return hosts, nil
}

// ========== 批量操作 ==========

// BatchDelete 批量删除 Host
func BatchDeleteHosts(ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	err := database.DB.Where("id IN ?", ids).Delete(&Host{}).Error
	if err != nil {
		return err
	}

	// 更新缓存
	for _, id := range ids {
		hostCache.Delete(id)
	}
	notifyHostChanged() // 通知外部模块同步
	return nil
}

// ========== 文本导出导入 ==========

// ExportToText 将所有 Host 导出为文本格式
// 格式：hostname IP # 备注（每行一条）
func ExportHostsToText() string {
	hosts := hostCache.GetAllSorted(func(a, b Host) bool {
		return a.ID < b.ID
	})

	var lines []string
	for _, h := range hosts {
		line := fmt.Sprintf("%s %s", h.Hostname, h.IP)
		if h.Remark != "" {
			line += " # " + h.Remark
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// SyncFromText 从文本全量同步 Host 数据
// 解析文本，与数据库同步（新增、修改、删除）
// 返回同步结果统计
func SyncHostsFromText(text string) (added, updated, deleted int, err error) {
	// 解析文本中的 host 条目
	newHosts := parseHostText(text)

	// 获取当前所有 host（以 hostname 为键）
	currentHosts := make(map[string]Host)
	for _, h := range hostCache.GetAll() {
		currentHosts[h.Hostname] = h
	}

	// 记录文本中出现的 hostname
	textHostnames := make(map[string]bool)

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 处理新增和更新
	for _, newHost := range newHosts {
		textHostnames[newHost.Hostname] = true

		if existing, exists := currentHosts[newHost.Hostname]; exists {
			// 检查是否需要更新
			if existing.IP != newHost.IP || existing.Remark != newHost.Remark {
				existing.IP = newHost.IP
				existing.Remark = newHost.Remark
				existing.UpdatedAt = time.Now()
				if err := tx.Model(&existing).Updates(map[string]interface{}{
					"ip":         existing.IP,
					"remark":     existing.Remark,
					"updated_at": existing.UpdatedAt,
				}).Error; err != nil {
					tx.Rollback()
					return 0, 0, 0, err
				}
				hostCache.Set(existing.ID, existing)
				updated++
			}
		} else {
			// 新增
			newHost.CreatedAt = time.Now()
			newHost.UpdatedAt = time.Now()
			if err := tx.Create(&newHost).Error; err != nil {
				tx.Rollback()
				return 0, 0, 0, err
			}
			hostCache.Set(newHost.ID, newHost)
			added++
		}
	}

	// 处理删除（数据库中存在但文本中不存在的）
	for hostname, host := range currentHosts {
		if !textHostnames[hostname] {
			if err := tx.Delete(&host).Error; err != nil {
				tx.Rollback()
				return 0, 0, 0, err
			}
			hostCache.Delete(host.ID)
			deleted++
		}
	}

	if err := tx.Commit().Error; err != nil {
		return 0, 0, 0, err
	}

	// 有变更时通知外部模块同步
	if added > 0 || updated > 0 || deleted > 0 {
		notifyHostChanged()
	}

	return added, updated, deleted, nil
}

// parseHostText 解析 host 文本
// 支持格式：
// - hostname IP
// - hostname IP # 备注
// - 忽略空行和以 # 开头的注释行
func parseHostText(text string) []Host {
	var hosts []Host
	scanner := bufio.NewScanner(strings.NewReader(text))
	// 匹配：hostname IP [# 备注]
	// hostname 可以是域名或通配符域名
	lineRegex := regexp.MustCompile(`^([^\s#]+)\s+([^\s#]+)(?:\s*#\s*(.*))?$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和纯注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := lineRegex.FindStringSubmatch(line)
		if matches == nil {
			continue // 格式不正确的行跳过
		}

		host := Host{
			Hostname: strings.TrimSpace(matches[1]),
			IP:       strings.TrimSpace(matches[2]),
		}
		if len(matches) > 3 && matches[3] != "" {
			host.Remark = strings.TrimSpace(matches[3])
		}
		host.Source = "手动导入"

		// 简单验证
		if host.Hostname != "" && host.IP != "" {
			hosts = append(hosts, host)
		}
	}

	return hosts
}

// GetAllHosts 获取所有 Host（供其他模块调用）
func GetAllHosts() []Host {
	return hostCache.GetAllSorted(func(a, b Host) bool {
		return a.ID < b.ID
	})
}

// GetHostMap 获取 hostname 到 IP 的映射（供其他模块高效查询）
func GetHostMap() map[string]string {
	hostMap := make(map[string]string)
	for _, h := range hostCache.GetAll() {
		hostMap[h.Hostname] = h.IP
	}
	return hostMap
}

// Ensure sort is used
var _ = sort.Slice

// ========== 测速Host持久化 ==========

// HostMappingInfo 用于批量保存Host时传递节点信息
type HostMappingInfo struct {
	Hostname  string // 代理服务器域名（从link解析得到）
	IP        string // DNS解析得到的IP
	NodeName  string // 节点名称
	Group     string // 节点分组
	Source    string // 节点来源
	DNSSource string // DNS来源
}

// BatchUpsertHosts 批量添加或更新Host映射（测速专用）
// 使用 ON CONFLICT 实现高效的 upsert 操作
// mappings: HostMappingInfo 列表（已去重）
// 返回: 成功处理数, 错误
func BatchUpsertHosts(mappings []HostMappingInfo) (int, error) {
	if len(mappings) == 0 {
		return 0, nil
	}

	now := time.Now()
	successCount := 0

	// 计算过期时间（测速自动持久化的Host会设置过期时间）
	expireAt := CalculateExpireTime()

	// 预处理：生成Host记录，跳过无效数据
	hosts := make([]Host, 0, len(mappings))
	for _, m := range mappings {
		if m.Hostname == "" || m.IP == "" {
			continue
		}
		hosts = append(hosts, Host{
			Hostname:  m.Hostname,
			IP:        m.IP,
			Remark:    formatHostRemark(m.NodeName, m.Group, m.Source),
			Source:    m.DNSSource,
			ExpireAt:  expireAt,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	if len(hosts) == 0 {
		return 0, nil
	}

	// 分块处理，避免SQLite变量限制
	chunks := chunkHosts(hosts, database.BatchSize)

	for _, chunk := range chunks {
		// 使用 ON CONFLICT 实现 upsert：已存在则更新 ip/remark/expire_at/updated_at
		result := database.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "hostname"}},
			DoUpdates: clause.AssignmentColumns([]string{"ip", "remark", "source", "expire_at", "updated_at"}),
		}).Create(&chunk)

		if result.Error != nil {
			utils.Warn("批量upsert Host失败: %v，降级到逐条处理", result.Error)
			// 降级到逐条处理
			for _, h := range chunk {
				if err := upsertSingleHost(h); err == nil {
					successCount++
				}
			}
		} else {
			successCount += len(chunk)
			// 更新缓存
			for i := range chunk {
				if chunk[i].ID > 0 {
					hostCache.Set(chunk[i].ID, chunk[i])
				}
			}
		}
	}

	// 有变更时通知外部模块同步
	if successCount > 0 {
		// 重新加载缓存以确保数据一致性（因为upsert可能更新了已存在记录）
		reloadHostCache()
		utils.Info("[Host持久化] 成功处理 %d 条", successCount)
		notifyHostChanged()
	}

	return successCount, nil
}

// upsertSingleHost 单条upsert（降级用）
func upsertSingleHost(h Host) error {
	existingHosts := hostCache.GetByIndex("hostname", h.Hostname)
	if len(existingHosts) > 0 {
		existing := existingHosts[0]
		return database.DB.Model(&existing).Updates(map[string]interface{}{
			"ip":         h.IP,
			"remark":     h.Remark,
			"source":     h.Source,
			"expire_at":  h.ExpireAt,
			"updated_at": h.UpdatedAt,
		}).Error
	}
	return database.DB.Create(&h).Error
}

// reloadHostCache 重新加载Host缓存
func reloadHostCache() {
	var hosts []Host
	if err := database.DB.Find(&hosts).Error; err != nil {
		utils.Error("重新加载Host缓存失败: %v", err)
		return
	}
	hostCache.LoadAll(hosts)
}

// chunkHosts 将Host列表分块
func chunkHosts(hosts []Host, size int) [][]Host {
	if size <= 0 {
		size = 100
	}
	var chunks [][]Host
	for i := 0; i < len(hosts); i += size {
		end := i + size
		if end > len(hosts) {
			end = len(hosts)
		}
		chunks = append(chunks, hosts[i:end])
	}
	return chunks
}

// formatHostRemark 生成友好的备注格式
// 格式: [自动] 节点名称 | 分组:xxx | 来源:xxx
func formatHostRemark(nodeName, group, source string) string {
	parts := []string{"[自动]"}

	if nodeName != "" {
		// 节点名称可能很长，截取前30个字符
		if len(nodeName) > 30 {
			nodeName = nodeName[:30] + "..."
		}
		parts = append(parts, nodeName)
	}

	if group != "" {
		parts = append(parts, "分组:"+group)
	}

	if source != "" {
		parts = append(parts, "来源:"+source)
	}

	return strings.Join(parts, " | ")
}

// ========== 有效期管理 ==========

// SetHostPinned 设置 Host 的固定状态
// pinned=true 时，该 Host 不会被过期清理
func SetHostPinned(id int, pinned bool) error {
	host, ok := hostCache.Get(id)
	if !ok {
		return fmt.Errorf("host ID %d 不存在", id)
	}

	if err := database.DB.Model(&Host{}).Where("id = ?", id).Update("pinned", pinned).Error; err != nil {
		return err
	}

	// 更新缓存
	host.Pinned = pinned
	hostCache.Set(id, host)
	return nil
}

// CleanExpiredHosts 清理过期的 Host
// 删除条件：ExpireAt 不为空 且 ExpireAt < 当前时间 且 Pinned = false
// 返回删除的数量
func CleanExpiredHosts() (int, error) {
	now := time.Now()

	// 先查询要删除的 ID（用于更新缓存）
	var expiredHosts []Host
	if err := database.DB.Where("expire_at IS NOT NULL AND expire_at < ? AND pinned = ?", now, false).Find(&expiredHosts).Error; err != nil {
		return 0, err
	}

	if len(expiredHosts) == 0 {
		return 0, nil
	}

	// 执行删除
	result := database.DB.Where("expire_at IS NOT NULL AND expire_at < ? AND pinned = ?", now, false).Delete(&Host{})
	if result.Error != nil {
		return 0, result.Error
	}

	deletedCount := int(result.RowsAffected)

	// 更新缓存
	for _, h := range expiredHosts {
		hostCache.Delete(h.ID)
	}

	// 有变更时通知外部模块同步
	if deletedCount > 0 {
		utils.Info("[Host清理] 已删除 %d 条过期Host记录", deletedCount)
		notifyHostChanged()
	}

	return deletedCount, nil
}

// GetHostExpireHours 获取 Host 有效期设置（小时）
// 返回 0 表示永不过期
func GetHostExpireHours() int {
	hoursStr, err := GetSetting("host_expire_hours")
	if err != nil || hoursStr == "" {
		return 0
	}
	hours := 0
	fmt.Sscanf(hoursStr, "%d", &hours)
	if hours < 0 {
		return 0
	}
	return hours
}

// CalculateExpireTime 根据有效期设置计算过期时间
// 返回 nil 表示永不过期
func CalculateExpireTime() *time.Time {
	hours := GetHostExpireHours()
	if hours <= 0 {
		return nil
	}
	expireAt := time.Now().Add(time.Duration(hours) * time.Hour)
	return &expireAt
}

// MigrateHostExpireFields 执行 Host 有效期字段迁移
func MigrateHostExpireFields() error {
	return database.RunAutoMigrate("add_host_expire_fields_v1", &Host{})
}
