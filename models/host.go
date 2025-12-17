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
	"time"
)

// Host 自定义 Host 映射模型
type Host struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Hostname  string    `json:"hostname" gorm:"size:255;uniqueIndex;not null"` // 域名
	IP        string    `json:"ip" gorm:"size:45;not null"`                    // IP 地址 (支持 IPv6)
	Remark    string    `json:"remark" gorm:"size:255"`                        // 备注
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// hostCache 使用泛型缓存
var hostCache *cache.MapCache[int, Host]

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
	// 从DB读取完整数据后更新缓存
	var updated Host
	if err := database.DB.First(&updated, h.ID).Error; err == nil {
		hostCache.Set(h.ID, updated)
	}
	return nil
}

// Delete 删除 Host (Write-Through)
func (h *Host) Delete() error {
	err := database.DB.Delete(h).Error
	if err != nil {
		return err
	}
	hostCache.Delete(h.ID)
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
