package models

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/node/protocol"
	"sublink/utils"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Node struct {
	ID              int    `gorm:"primaryKey"`
	Link            string `gorm:"uniqueIndex:idx_link_id"` //出站代理原始连接
	Name            string //系统内节点名称
	LinkName        string //节点原始名称
	Protocol        string `gorm:"index"` //协议类型 (vmess, vless, trojan, ss 等)
	LinkAddress     string //节点原始地址
	LinkHost        string //节点原始Host
	LinkPort        string //节点原始端口
	LinkCountry     string //节点所属国家、落地IP国家
	LandingIP       string //落地IP地址
	DialerProxyName string
	Source          string `gorm:"default:'manual'"`
	SourceID        int
	Group           string
	Speed           float64   `gorm:"default:0"`          // 测速结果(MB/s)
	DelayTime       int       `gorm:"default:0"`          // 延迟时间(ms)
	SpeedStatus     string    `gorm:"default:'untested'"` // 速度测试状态: untested, success, timeout, error
	DelayStatus     string    `gorm:"default:'untested'"` // 延迟测试状态: untested, success, timeout, error
	LatencyCheckAt  string    // 延迟测试时间
	SpeedCheckAt    string    // 测速时间
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"CreatedAt"` // 创建时间
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"UpdatedAt"` // 更新时间
	Tags            string    // 标签ID，逗号分隔，如 "1,3,5"
}

// nodeCache 使用新的泛型缓存，支持二级索引
var nodeCache *cache.MapCache[int, Node]

func init() {
	// 初始化节点缓存，主键为 ID
	nodeCache = cache.NewMapCache(func(n Node) int { return n.ID })
	// 添加二级索引
	nodeCache.AddIndex("group", func(n Node) string { return n.Group })
	nodeCache.AddIndex("source", func(n Node) string { return n.Source })
	nodeCache.AddIndex("country", func(n Node) string { return n.LinkCountry })
	nodeCache.AddIndex("protocol", func(n Node) string { return n.Protocol })
	nodeCache.AddIndex("sourceID", func(n Node) string { return fmt.Sprintf("%d", n.SourceID) })
	nodeCache.AddIndex("name", func(n Node) string { return n.Name })
}

// InitNodeCache 初始化节点缓存
func InitNodeCache() error {
	utils.Info("加载节点列表到缓存")
	var nodes []Node
	if err := database.DB.Find(&nodes).Error; err != nil {
		return err
	}

	// 使用批量加载方式初始化缓存
	nodeCache.LoadAll(nodes)
	utils.Info("节点缓存初始化完成，共加载 %d 个节点", nodeCache.Count())

	// 注册到缓存管理器
	cache.Manager.Register("node", nodeCache)
	return nil
}

// UpdateNodeCache 更新节点缓存（供外部包使用）
func UpdateNodeCache(id int, node Node) {
	nodeCache.Set(id, node)
}

// Add 添加节点
func (node *Node) Add() error {
	// Write-Through: 先写数据库
	err := database.DB.Create(node).Error
	if err != nil {
		return err
	}
	// 再更新缓存
	nodeCache.Set(node.ID, *node)
	return nil
}

// Update 更新节点
func (node *Node) Update() error {
	if node.Name == "" {
		node.Name = node.LinkName
	}
	node.UpdatedAt = time.Now()
	// Write-Through: 先写数据库
	err := database.DB.Model(node).Select("Name", "Link", "DialerProxyName", "Group", "LinkName", "LinkAddress", "LinkHost", "LinkPort", "LinkCountry", "UpdatedAt").Updates(node).Error
	if err != nil {
		return err
	}
	// 更新缓存：获取完整节点后更新
	if cachedNode, ok := nodeCache.Get(node.ID); ok {
		cachedNode.Name = node.Name
		cachedNode.Link = node.Link
		cachedNode.DialerProxyName = node.DialerProxyName
		cachedNode.Group = node.Group
		cachedNode.LinkName = node.LinkName
		cachedNode.LinkAddress = node.LinkAddress
		cachedNode.LinkHost = node.LinkHost
		cachedNode.LinkPort = node.LinkPort
		cachedNode.LinkCountry = node.LinkCountry
		cachedNode.UpdatedAt = node.UpdatedAt
		nodeCache.Set(node.ID, cachedNode)
	} else {
		// 缓存未命中，从 DB 读取完整数据
		var fullNode Node
		if err := database.DB.First(&fullNode, node.ID).Error; err == nil {
			nodeCache.Set(node.ID, fullNode)
		}
	}
	return nil
}

// UpdateSpeed 更新节点测速结果
func (node *Node) UpdateSpeed() error {
	err := database.DB.Model(node).Select("Speed", "SpeedStatus", "LinkCountry", "LandingIP", "DelayTime", "DelayStatus", "LatencyCheckAt", "SpeedCheckAt").Updates(node).Error
	if err != nil {
		return err
	}

	if cachedNode, ok := nodeCache.Get(node.ID); ok {
		cachedNode.Speed = node.Speed
		cachedNode.SpeedStatus = node.SpeedStatus
		cachedNode.DelayTime = node.DelayTime
		cachedNode.DelayStatus = node.DelayStatus
		cachedNode.LatencyCheckAt = node.LatencyCheckAt
		cachedNode.SpeedCheckAt = node.SpeedCheckAt
		cachedNode.LinkCountry = node.LinkCountry
		cachedNode.LandingIP = node.LandingIP
		nodeCache.Set(node.ID, cachedNode)
	}
	return nil
}

// SpeedTestResult 测速结果结构（用于批量更新）
type SpeedTestResult struct {
	NodeID         int
	Speed          float64
	SpeedStatus    string
	DelayTime      int
	DelayStatus    string
	LatencyCheckAt string
	SpeedCheckAt   string
	LinkCountry    string
	LandingIP      string
}

// BatchAddNodes 批量添加节点（高效 + 容错）
// 优化策略：
// 1. 分块处理避免 SQLite 变量限制
// 2. 使用 ON CONFLICT DO NOTHING 跳过已存在的节点
// 3. 批量插入失败时，降级到逐条插入以保证容错性
// 4. 单条失败只记录日志，不影响其他节点
func BatchAddNodes(nodes []Node) error {
	if len(nodes) == 0 {
		return nil
	}

	// 分块处理
	chunks := chunkNodes(nodes, database.BatchSize)
	insertedCount := 0

	for chunkIdx, chunk := range chunks {
		// 尝试批量插入
		result := database.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "link"}},
			DoNothing: true,
		}).Create(&chunk)

		if result.Error != nil {
			// 批量插入失败，降级到逐条插入
			utils.Warn("分块 %d 批量插入失败，降级到逐条插入: %v", chunkIdx, result.Error)
			individualInserted := fallbackToIndividualNodeInsert(chunk)
			insertedCount += individualInserted
		} else {
			insertedCount += int(result.RowsAffected)
			// 批量更新缓存（只更新成功插入的，有ID的节点）
			for i := range chunk {
				if chunk[i].ID > 0 {
					nodeCache.Set(chunk[i].ID, chunk[i])
				}
			}
		}
	}

	utils.Info("批量添加节点完成: 尝试 %d 个，实际插入 %d 个（跳过已存在）", len(nodes), insertedCount)
	return nil
}

// fallbackToIndividualNodeInsert 降级到逐条插入节点（容错）
func fallbackToIndividualNodeInsert(nodes []Node) int {
	insertedCount := 0
	for i := range nodes {
		// 使用 ON CONFLICT DO NOTHING 跳过已存在的节点
		result := database.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "link"}},
			DoNothing: true,
		}).Create(&nodes[i])

		if result.Error != nil {
			utils.Error("节点 [%s] 插入失败: %v", nodes[i].Name, result.Error)
			continue
		}

		if result.RowsAffected > 0 {
			insertedCount++
			// 更新缓存
			if nodes[i].ID > 0 {
				nodeCache.Set(nodes[i].ID, nodes[i])
			}
		}
	}
	return insertedCount
}

// BatchUpdateSpeedResults 批量更新测速结果（高效 + 容错）
// 优化策略：
// 1. 分块处理避免 SQLite 变量限制和长时间锁定
// 2. 每块使用 CASE WHEN 批量更新（一条 SQL 更新多条记录）
// 3. 批量更新失败时，降级到逐条更新以保证容错性
// 4. 单条失败只记录日志，不影响其他记录
func BatchUpdateSpeedResults(results []SpeedTestResult) error {
	if len(results) == 0 {
		return nil
	}

	chunks := chunkSpeedResults(results, database.BatchSize)
	successCount := 0
	totalAttempts := len(results)

	for chunkIdx, chunk := range chunks {
		// 尝试使用 CASE WHEN 批量更新
		batchSuccess, batchErr := tryBatchUpdateWithCaseWhen(chunk)
		if batchErr == nil {
			successCount += batchSuccess
			// 批量更新成功，批量更新缓存
			batchUpdateNodeCache(chunk)
		} else {
			// 批量更新失败，降级到逐条更新
			utils.Warn("分块 %d 批量更新失败，降级到逐条更新: %v", chunkIdx, batchErr)
			individualSuccess := fallbackToIndividualSpeedUpdate(chunk)
			successCount += individualSuccess
		}
	}

	utils.Info("批量更新测速结果完成: 尝试 %d 个，成功 %d 个，分 %d 块处理", totalAttempts, successCount, len(chunks))
	return nil
}

// speedResultField 定义测速结果字段的映射关系
type speedResultField struct {
	column    string                         // 数据库列名
	valueFunc func(r SpeedTestResult) string // 获取值的函数
}

// speedResultFields 测速结果字段映射表（新增字段只需在此处添加）
var speedResultFields = []speedResultField{
	{"speed", func(r SpeedTestResult) string { return fmt.Sprintf("%f", r.Speed) }},
	{"speed_status", func(r SpeedTestResult) string { return fmt.Sprintf("'%s'", escapeSQL(r.SpeedStatus)) }},
	{"delay_time", func(r SpeedTestResult) string { return fmt.Sprintf("%d", r.DelayTime) }},
	{"delay_status", func(r SpeedTestResult) string { return fmt.Sprintf("'%s'", escapeSQL(r.DelayStatus)) }},
	{"latency_check_at", func(r SpeedTestResult) string { return fmt.Sprintf("'%s'", escapeSQL(r.LatencyCheckAt)) }},
	{"speed_check_at", func(r SpeedTestResult) string { return fmt.Sprintf("'%s'", escapeSQL(r.SpeedCheckAt)) }},
	{"link_country", func(r SpeedTestResult) string { return fmt.Sprintf("'%s'", escapeSQL(r.LinkCountry)) }},
	{"landing_ip", func(r SpeedTestResult) string { return fmt.Sprintf("'%s'", escapeSQL(r.LandingIP)) }},
}

// tryBatchUpdateWithCaseWhen 使用 CASE WHEN 批量更新（高效）
// 生成形如: UPDATE nodes SET speed = CASE id WHEN 1 THEN 100.5 WHEN 2 THEN 200.3 END, ... WHERE id IN (1,2)
func tryBatchUpdateWithCaseWhen(chunk []SpeedTestResult) (int, error) {
	if len(chunk) == 0 {
		return 0, nil
	}

	var sb strings.Builder
	sb.WriteString("UPDATE nodes SET ")

	// 遍历字段映射表，生成 CASE WHEN 语句
	for i, field := range speedResultFields {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(field.column)
		sb.WriteString(" = CASE id ")
		for _, r := range chunk {
			sb.WriteString(fmt.Sprintf("WHEN %d THEN %s ", r.NodeID, field.valueFunc(r)))
		}
		sb.WriteString("END")
	}

	// WHERE 子句
	sb.WriteString(" WHERE id IN (")
	for i, r := range chunk {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%d", r.NodeID))
	}
	sb.WriteString(")")

	// 执行 SQL
	result := database.DB.Exec(sb.String())

	if result.Error != nil {
		return 0, result.Error
	}

	return int(result.RowsAffected), nil
}

// escapeSQL 转义 SQL 字符串中的特殊字符，防止 SQL 注入
func escapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// batchUpdateNodeCache 批量更新节点缓存
func batchUpdateNodeCache(chunk []SpeedTestResult) {
	for _, r := range chunk {
		if cachedNode, ok := nodeCache.Get(r.NodeID); ok {
			cachedNode.Speed = r.Speed
			cachedNode.SpeedStatus = r.SpeedStatus
			cachedNode.DelayTime = r.DelayTime
			cachedNode.DelayStatus = r.DelayStatus
			cachedNode.LatencyCheckAt = r.LatencyCheckAt
			cachedNode.SpeedCheckAt = r.SpeedCheckAt
			cachedNode.LinkCountry = r.LinkCountry
			cachedNode.LandingIP = r.LandingIP
			nodeCache.Set(r.NodeID, cachedNode)
		}
	}
}

// fallbackToIndividualSpeedUpdate 降级到逐条更新（容错）
func fallbackToIndividualSpeedUpdate(chunk []SpeedTestResult) int {
	successCount := 0
	for _, r := range chunk {
		err := database.DB.Model(&Node{}).Where("id = ?", r.NodeID).Updates(map[string]interface{}{
			"speed":            r.Speed,
			"speed_status":     r.SpeedStatus,
			"delay_time":       r.DelayTime,
			"delay_status":     r.DelayStatus,
			"latency_check_at": r.LatencyCheckAt,
			"speed_check_at":   r.SpeedCheckAt,
			"link_country":     r.LinkCountry,
			"landing_ip":       r.LandingIP,
		}).Error

		if err != nil {
			utils.Error("节点 ID=%d 更新失败: %v", r.NodeID, err)
			continue
		}
		successCount++

		// 逐条更新缓存
		if cachedNode, ok := nodeCache.Get(r.NodeID); ok {
			cachedNode.Speed = r.Speed
			cachedNode.SpeedStatus = r.SpeedStatus
			cachedNode.DelayTime = r.DelayTime
			cachedNode.DelayStatus = r.DelayStatus
			cachedNode.LatencyCheckAt = r.LatencyCheckAt
			cachedNode.SpeedCheckAt = r.SpeedCheckAt
			cachedNode.LinkCountry = r.LinkCountry
			cachedNode.LandingIP = r.LandingIP
			nodeCache.Set(r.NodeID, cachedNode)
		}
	}
	return successCount
}

// chunkSpeedResults 将测速结果切片分块
func chunkSpeedResults(results []SpeedTestResult, chunkSize int) [][]SpeedTestResult {
	if chunkSize <= 0 {
		chunkSize = database.BatchSize
	}

	var chunks [][]SpeedTestResult
	for i := 0; i < len(results); i += chunkSize {
		end := i + chunkSize
		if end > len(results) {
			end = len(results)
		}
		chunks = append(chunks, results[i:end])
	}
	return chunks
}

// chunkNodes 将节点切片分块
func chunkNodes(nodes []Node, chunkSize int) [][]Node {
	if chunkSize <= 0 {
		chunkSize = database.BatchSize
	}

	var chunks [][]Node
	for i := 0; i < len(nodes); i += chunkSize {
		end := i + chunkSize
		if end > len(nodes) {
			end = len(nodes)
		}
		chunks = append(chunks, nodes[i:end])
	}
	return chunks
}

// Find 查找节点是否重复
func (node *Node) Find() error {
	// 优先查缓存
	results := nodeCache.Filter(func(n Node) bool {
		return n.Link == node.Link || n.Name == node.Name
	})
	if len(results) > 0 {
		*node = results[0]
		return nil
	}

	// 缓存未命中，查 DB
	err := database.DB.Where("link = ? or name = ?", node.Link, node.Name).First(node).Error
	if err != nil {
		return err
	}

	// 更新缓存
	nodeCache.Set(node.ID, *node)
	return nil
}

// FindByName 仅通过名称查找节点（精确匹配）
// 用于订阅节点关联场景，避免 Find() 中 link="" 条件导致的错误匹配
func (node *Node) FindByName() error {
	if node.Name == "" {
		return fmt.Errorf("node name is required")
	}

	// 使用缓存二级索引精确查找
	results := nodeCache.GetByIndex("name", node.Name)
	if len(results) > 0 {
		*node = results[0]
		return nil
	}

	// 缓存未命中，查 DB
	err := database.DB.Where("name = ?", node.Name).First(node).Error
	if err != nil {
		return err
	}

	// 更新缓存
	nodeCache.Set(node.ID, *node)
	return nil
}

// GetByID 根据ID查找节点
func (node *Node) GetByID() error {
	if cachedNode, ok := nodeCache.Get(node.ID); ok {
		*node = cachedNode
		return nil
	}

	// 缓存未命中，查 DB
	err := database.DB.First(node, node.ID).Error
	if err != nil {
		return err
	}

	// 更新缓存
	nodeCache.Set(node.ID, *node)
	return nil
}

// GetNodesByIDs 根据ID列表批量获取节点
func GetNodesByIDs(ids []int) ([]Node, error) {
	if len(ids) == 0 {
		return []Node{}, nil
	}

	nodes := make([]Node, 0, len(ids))

	// 优先从缓存获取
	for _, id := range ids {
		if n, found := nodeCache.Get(id); found {
			nodes = append(nodes, n)
		}
	}

	// 如果缓存命中率100%，直接返回
	if len(nodes) == len(ids) {
		return nodes, nil
	}

	// 否则从数据库获取全部（确保数据一致性）
	nodes = make([]Node, 0, len(ids))
	if err := database.DB.Where("id IN ?", ids).Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

// List 节点列表
func (node *Node) List() ([]Node, error) {
	// 使用 GetAllSorted 获取排序的节点列表
	nodes := nodeCache.GetAllSorted(func(a, b Node) bool {
		return a.ID < b.ID
	})
	return nodes, nil
}

type NodeFilter struct {
	Search      string   // 搜索关键词（匹配节点名称或链接）
	Group       string   // 分组过滤
	Source      string   // 来源过滤
	Protocol    string   // 协议类型过滤（如 vmess, vless, trojan 等）
	MaxDelay    int      // 最大延迟(ms)，只显示延迟在此值以下的节点
	MinSpeed    float64  // 最低速度(MB/s)，只显示速度在此值以上的节点
	SpeedStatus string   // 速度状态过滤: untested, success, timeout, error
	DelayStatus string   // 延迟状态过滤: untested, success, timeout, error
	Countries   []string // 国家代码过滤
	Tags        []string // 标签过滤（匹配任一标签的节点）
	SortBy      string   // 排序字段: "delay" 或 "speed"
	SortOrder   string   // 排序顺序: "asc" 或 "desc"
}

// ListWithFilters 根据过滤条件获取节点列表
func (node *Node) ListWithFilters(filter NodeFilter) ([]Node, error) {
	// 预处理搜索关键词
	searchLower := strings.ToLower(filter.Search)

	// 创建国家代码映射，加速查找
	countryMap := make(map[string]bool)
	for _, c := range filter.Countries {
		countryMap[c] = true
	}

	// 创建标签映射，加速查找
	tagMap := make(map[string]bool)
	for _, t := range filter.Tags {
		tagMap[t] = true
	}

	// 使用缓存的 Filter 方法
	nodes := nodeCache.Filter(func(n Node) bool {
		// 搜索过滤
		if searchLower != "" {
			nameLower := strings.ToLower(n.Name)
			linkLower := strings.ToLower(n.Link)
			if !strings.Contains(nameLower, searchLower) && !strings.Contains(linkLower, searchLower) {
				return false
			}
		}

		// 分组过滤
		if filter.Group != "" {
			if filter.Group == "未分组" {
				if n.Group != "" {
					return false
				}
			} else {
				// 精确匹配分组（不区分大小写）
				if !strings.EqualFold(n.Group, filter.Group) {
					return false
				}
			}
		}

		// 来源过滤
		if filter.Source != "" {
			if filter.Source == "手动添加" {
				if n.Source != "" && n.Source != "manual" {
					return false
				}
			} else {
				// 精确匹配来源（不区分大小写）
				if !strings.EqualFold(n.Source, filter.Source) {
					return false
				}
			}
		}

		// 最大延迟过滤
		if filter.MaxDelay > 0 {
			if n.DelayTime <= 0 || n.DelayTime > filter.MaxDelay {
				return false
			}
		}

		// 最低速度过滤
		if filter.MinSpeed > 0 {
			if n.Speed <= filter.MinSpeed {
				return false
			}
		}

		// 国家代码过滤
		if len(countryMap) > 0 {
			if n.LinkCountry == "" || !countryMap[n.LinkCountry] {
				return false
			}
		}

		// 标签过滤：节点需要包含至少一个所选标签
		if len(tagMap) > 0 {
			nodeTags := strings.Split(n.Tags, ",")
			hasMatchingTag := false
			for _, tag := range nodeTags {
				tag = strings.TrimSpace(tag)
				if tag != "" && tagMap[tag] {
					hasMatchingTag = true
					break
				}
			}
			if !hasMatchingTag {
				return false
			}
		}

		// 速度状态过滤
		if filter.SpeedStatus != "" {
			if n.SpeedStatus != filter.SpeedStatus {
				return false
			}
		}

		// 延迟状态过滤
		if filter.DelayStatus != "" {
			if n.DelayStatus != filter.DelayStatus {
				return false
			}
		}

		// 协议类型过滤
		if filter.Protocol != "" {
			if !strings.EqualFold(n.Protocol, filter.Protocol) {
				return false
			}
		}

		return true
	})

	// 排序
	if filter.SortBy != "" {
		sort.Slice(nodes, func(i, j int) bool {
			switch filter.SortBy {
			case "delay":
				aValid := nodes[i].DelayTime > 0
				bValid := nodes[j].DelayTime > 0
				if !aValid && !bValid {
					return nodes[i].ID < nodes[j].ID
				}
				if !aValid {
					return false
				}
				if !bValid {
					return true
				}
				if filter.SortOrder == "desc" {
					return nodes[i].DelayTime > nodes[j].DelayTime
				}
				return nodes[i].DelayTime < nodes[j].DelayTime
			case "speed":
				aValid := nodes[i].Speed > 0
				bValid := nodes[j].Speed > 0
				if !aValid && !bValid {
					return nodes[i].ID < nodes[j].ID
				}
				if !aValid {
					return false
				}
				if !bValid {
					return true
				}
				if filter.SortOrder == "desc" {
					return nodes[i].Speed > nodes[j].Speed
				}
				return nodes[i].Speed < nodes[j].Speed
			default:
				return nodes[i].ID < nodes[j].ID
			}
		})
	} else {
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].ID < nodes[j].ID
		})
	}

	return nodes, nil
}

// ListWithFiltersPaginated 根据过滤条件获取分页节点列表
func (node *Node) ListWithFiltersPaginated(filter NodeFilter, page, pageSize int) ([]Node, int64, error) {
	// 先获取全部过滤结果
	allNodes, err := node.ListWithFilters(filter)
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(allNodes))

	// 如果不需要分页，返回全部
	if page <= 0 || pageSize <= 0 {
		return allNodes, total, nil
	}

	// 计算分页
	offset := (page - 1) * pageSize
	if offset >= len(allNodes) {
		return []Node{}, total, nil
	}

	end := offset + pageSize
	if end > len(allNodes) {
		end = len(allNodes)
	}

	return allNodes[offset:end], total, nil
}

// GetFilteredNodeIDs 获取符合过滤条件的所有节点ID（用于全选操作）
func (node *Node) GetFilteredNodeIDs(filter NodeFilter) ([]int, error) {
	allNodes, err := node.ListWithFilters(filter)
	if err != nil {
		return nil, err
	}

	ids := make([]int, len(allNodes))
	for i, n := range allNodes {
		ids[i] = n.ID
	}
	return ids, nil
}

// ListByGroups 根据分组获取节点列表
// 返回按节点 ID 排序的结果，确保顺序稳定（用于去重等顺序敏感操作）
func (node *Node) ListByGroups(groups []string) ([]Node, error) {
	groupMap := make(map[string]bool)
	for _, g := range groups {
		groupMap[g] = true
	}

	// 使用 FilterSorted 确保返回顺序稳定
	nodes := nodeCache.FilterSorted(
		func(n Node) bool {
			return groupMap[n.Group]
		},
		func(a, b Node) bool {
			return a.ID < b.ID
		},
	)
	return nodes, nil
}

// ListByTags 根据标签获取节点列表（匹配任意标签）
// 返回按节点 ID 排序的结果，确保顺序稳定
func (node *Node) ListByTags(tags []string) ([]Node, error) {
	tagMap := make(map[string]bool)
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" {
			tagMap[t] = true
		}
	}

	if len(tagMap) == 0 {
		return []Node{}, nil
	}

	// 使用 FilterSorted 确保返回顺序稳定
	nodes := nodeCache.FilterSorted(
		func(n Node) bool {
			nodeTags := n.GetTagNames()
			for _, nt := range nodeTags {
				if tagMap[nt] {
					return true
				}
			}
			return false
		},
		func(a, b Node) bool {
			return a.ID < b.ID
		},
	)
	return nodes, nil
}

// FilterNodesByTags 从已有节点列表中按标签过滤（用于分组+标签组合过滤）
func FilterNodesByTags(nodes []Node, tags []string) []Node {
	tagMap := make(map[string]bool)
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" {
			tagMap[t] = true
		}
	}

	if len(tagMap) == 0 {
		return nodes
	}

	var filtered []Node
	for _, n := range nodes {
		nodeTags := n.GetTagNames()
		for _, nt := range nodeTags {
			if tagMap[nt] {
				filtered = append(filtered, n)
				break
			}
		}
	}
	return filtered
}

// Del 删除节点
func (node *Node) Del() error {
	// 先清除节点与订阅的关联关系
	if err := database.DB.Exec("DELETE FROM subcription_nodes WHERE node_id = ?", node.ID).Error; err != nil {
		return err
	}
	// Write-Through: 先删除数据库
	err := database.DB.Delete(node).Error
	if err != nil {
		return err
	}
	// 再更新缓存
	nodeCache.Delete(node.ID)
	return nil
}

// UpsertNode 插入或更新节点
func (node *Node) UpsertNode() error {
	// Write-Through: 先写数据库
	err := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "link"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "link_name", "link_address", "link_host", "link_port", "link_country", "source", "source_id", "group"}),
	}).Create(node).Error
	if err != nil {
		return err
	}

	// 查询更新后的节点并更新缓存
	var updatedNode Node
	if err := database.DB.Where("link = ?", node.Link).First(&updatedNode).Error; err == nil {
		nodeCache.Set(updatedNode.ID, updatedNode)
		*node = updatedNode
	}
	return nil
}

// DeleteAutoSubscriptionNodes 删除订阅节点
func DeleteAutoSubscriptionNodes(sourceId int) error {
	// 使用二级索引获取要删除的节点
	nodesToDelete := nodeCache.GetByIndex("sourceID", strconv.Itoa(sourceId))
	nodeIDs := make([]int, 0, len(nodesToDelete))
	for _, n := range nodesToDelete {
		nodeIDs = append(nodeIDs, n.ID)
	}

	// 清除节点与订阅的关联关系
	if len(nodeIDs) > 0 {
		if err := database.DB.Exec("DELETE FROM subcription_nodes WHERE node_id IN ?", nodeIDs).Error; err != nil {
			return err
		}
	}

	// Write-Through: 先删除数据库
	err := database.DB.Where("source_id = ?", sourceId).Delete(&Node{}).Error
	if err != nil {
		return err
	}

	// 再更新缓存
	for _, n := range nodesToDelete {
		nodeCache.Delete(n.ID)
	}
	return nil
}

// BatchDel 批量删除节点 - 使用事务保证原子性
func BatchDel(ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	// 使用事务原子删除
	err := database.WithTransaction(func(tx *gorm.DB) error {
		// 先清除节点与订阅的关联关系
		if len(ids) > 0 {
			if err := tx.Exec("DELETE FROM subcription_nodes WHERE node_id IN ?", ids).Error; err != nil {
				return err
			}
		}

		// 删除节点
		if err := tx.Where("id IN ?", ids).Delete(&Node{}).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	// 事务成功后更新缓存
	for _, id := range ids {
		nodeCache.Delete(id)
	}
	return nil
}

// BatchUpdateGroup 批量更新节点分组 - 使用事务保证原子性
func BatchUpdateGroup(ids []int, group string) error {
	if len(ids) == 0 {
		return nil
	}

	// 使用事务更新
	err := database.WithTransaction(func(tx *gorm.DB) error {
		return tx.Model(&Node{}).Where("id IN ?", ids).Update("group", group).Error
	})

	if err != nil {
		return err
	}

	// 事务成功后更新缓存
	for _, id := range ids {
		if n, ok := nodeCache.Get(id); ok {
			n.Group = group
			nodeCache.Set(id, n)
		}
	}
	return nil
}

// BatchUpdateDialerProxy 批量更新节点前置代理 - 使用事务保证原子性
func BatchUpdateDialerProxy(ids []int, dialerProxyName string) error {
	if len(ids) == 0 {
		return nil
	}

	// 使用事务更新
	err := database.WithTransaction(func(tx *gorm.DB) error {
		return tx.Model(&Node{}).Where("id IN ?", ids).Update("dialer_proxy_name", dialerProxyName).Error
	})

	if err != nil {
		return err
	}

	// 事务成功后更新缓存
	for _, id := range ids {
		if n, ok := nodeCache.Get(id); ok {
			n.DialerProxyName = dialerProxyName
			nodeCache.Set(id, n)
		}
	}
	return nil
}

// BatchUpdateSource 批量更新节点来源 - 使用事务保证原子性
func BatchUpdateSource(ids []int, source string) error {
	if len(ids) == 0 {
		return nil
	}

	// 使用事务更新
	err := database.WithTransaction(func(tx *gorm.DB) error {
		return tx.Model(&Node{}).Where("id IN ?", ids).Update("source", source).Error
	})

	if err != nil {
		return err
	}

	// 事务成功后更新缓存
	for _, id := range ids {
		if n, ok := nodeCache.Get(id); ok {
			n.Source = source
			nodeCache.Set(id, n)
		}
	}
	return nil
}

// GetAllGroups 获取所有分组
func (node *Node) GetAllGroups() ([]string, error) {
	// 使用二级索引获取所有不同的分组值
	return nodeCache.GetDistinctIndexValues("group"), nil
}

// GetAllSources 获取所有来源
func (node *Node) GetAllSources() ([]string, error) {
	// 使用二级索引获取所有不同的来源值
	return nodeCache.GetDistinctIndexValues("source"), nil
}

// GetBestProxyNode 获取最佳代理节点（延迟最低且速度大于0）
func GetBestProxyNode() (*Node, error) {
	// 使用缓存的 Filter 方法
	nodes := nodeCache.Filter(func(n Node) bool {
		return n.DelayTime > 0 && n.Speed > 0
	})

	var bestNode *Node
	for _, n := range nodes {
		if bestNode == nil || n.DelayTime < bestNode.DelayTime {
			nodeCopy := n
			bestNode = &nodeCopy
		}
	}

	if bestNode != nil {
		return bestNode, nil
	}

	// 缓存中没有符合条件的节点，从数据库查询
	var dbNodes []Node
	if err := database.DB.Where("delay_time > 0 AND speed > 0").Order("delay_time ASC").Limit(1).Find(&dbNodes).Error; err != nil {
		return nil, err
	}

	if len(dbNodes) == 0 {
		return nil, nil
	}

	return &dbNodes[0], nil
}

// ListBySourceID 根据订阅ID查询节点列表
func ListBySourceID(sourceID int) ([]Node, error) {
	// 使用二级索引查询
	nodes := nodeCache.GetByIndex("sourceID", strconv.Itoa(sourceID))

	// 如果缓存中有数据，直接返回
	if len(nodes) > 0 {
		return nodes, nil
	}

	// 缓存中没有数据，从数据库查询
	if err := database.DB.Where("source_id = ?", sourceID).Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

// UpdateNodesBySourceID 根据订阅ID批量更新节点的来源名称和分组
func UpdateNodesBySourceID(sourceID int, sourceName string, group string) error {
	// Write-Through: 先更新数据库
	updateFields := map[string]interface{}{
		"source": sourceName,
		"group":  group,
	}
	if err := database.DB.Model(&Node{}).Where("source_id = ?", sourceID).Updates(updateFields).Error; err != nil {
		return err
	}

	// 再更新缓存
	nodesToUpdate := nodeCache.GetByIndex("sourceID", strconv.Itoa(sourceID))
	for _, n := range nodesToUpdate {
		n.Source = sourceName
		n.Group = group
		nodeCache.Set(n.ID, n)
	}
	return nil
}

// GetFastestSpeedNode 获取最快速度节点
func GetFastestSpeedNode() *Node {
	nodes := nodeCache.Filter(func(n Node) bool {
		return n.Speed > 0
	})

	var fastest *Node
	for _, n := range nodes {
		if fastest == nil || n.Speed > fastest.Speed {
			nodeCopy := n
			fastest = &nodeCopy
		}
	}
	return fastest
}

// GetLowestDelayNode 获取最低延迟节点
func GetLowestDelayNode() *Node {
	nodes := nodeCache.Filter(func(n Node) bool {
		return n.DelayTime > 0
	})

	var lowest *Node
	for _, n := range nodes {
		if lowest == nil || n.DelayTime < lowest.DelayTime {
			nodeCopy := n
			lowest = &nodeCopy
		}
	}
	return lowest
}

// GetAllCountries 获取所有唯一的国家代码
func GetAllCountries() []string {
	// 使用二级索引获取所有不同的国家值
	return nodeCache.GetDistinctIndexValues("country")
}

// GetNodeCountryStats 获取按国家统计的节点数量
func GetNodeCountryStats() map[string]int {
	stats := make(map[string]int)
	allNodes := nodeCache.GetAll()
	for _, n := range allNodes {
		country := n.LinkCountry
		if country == "" {
			country = "未知"
		}
		stats[country]++
	}
	return stats
}

// GetNodeProtocolStats 获取按协议统计的节点数量
func GetNodeProtocolStats() map[string]int {
	stats := make(map[string]int)
	allNodes := nodeCache.GetAll()
	for _, n := range allNodes {
		// 使用节点存储的协议类型，如果为空则从链接解析
		protoName := n.Protocol
		if protoName == "" {
			protoName = protocol.GetProtocolFromLink(n.Link)
		}
		// 转换为显示名称
		protoLabel := protocol.GetProtocolLabel(protoName)
		stats[protoLabel]++
	}
	return stats
}

// GetAllProtocols 获取所有使用中的协议类型列表（用于过滤器选项）
// 返回标准化的小写协议名称列表
func GetAllProtocols() []string {
	protoSet := make(map[string]bool)
	allNodes := nodeCache.GetAll()
	for _, n := range allNodes {
		protoName := n.Protocol
		if protoName == "" {
			protoName = protocol.GetProtocolFromLink(n.Link)
		}
		if protoName != "" && protoName != "unknown" && protoName != "other" {
			protoSet[protoName] = true
		}
	}

	protocols := make([]string, 0, len(protoSet))
	for p := range protoSet {
		protocols = append(protocols, p)
	}
	return protocols
}

// GetNodeByName 根据节点名称获取节点
func GetNodeByName(name string) (*Node, bool) {
	nodes := nodeCache.GetByIndex("name", name)
	if len(nodes) > 0 {
		return &nodes[0], true
	}
	return nil, false
}

// TagStat 标签统计结构
type TagStat struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Count int    `json:"count"`
}

// GetNodeTagStats 获取按标签统计的节点数量
func GetNodeTagStats() []TagStat {
	allNodes := nodeCache.GetAll()
	tagCounts := make(map[string]int)
	noTagCount := 0

	for _, n := range allNodes {
		tagNames := n.GetTagNames()
		if len(tagNames) == 0 {
			noTagCount++
		} else {
			for _, tagName := range tagNames {
				tagCounts[tagName]++
			}
		}
	}

	// 构建结果，包含标签颜色
	result := make([]TagStat, 0, len(tagCounts)+1)

	// 先添加"无标签"统计
	if noTagCount > 0 {
		result = append(result, TagStat{
			Name:  "无标签",
			Color: "#9e9e9e",
			Count: noTagCount,
		})
	}

	// 添加各标签统计
	for tagName, count := range tagCounts {
		color := "#1976d2" // 默认颜色
		if tag, ok := tagCache.Get(tagName); ok {
			color = tag.Color
		}
		result = append(result, TagStat{
			Name:  tagName,
			Color: color,
			Count: count,
		})
	}

	return result
}

// GetNodeGroupStats 获取按分组统计的节点数量
func GetNodeGroupStats() map[string]int {
	stats := make(map[string]int)
	allNodes := nodeCache.GetAll()
	for _, n := range allNodes {
		group := n.Group
		if group == "" {
			group = "未分组"
		}
		stats[group]++
	}
	return stats
}

// GetNodeSourceStats 获取按来源统计的节点数量
func GetNodeSourceStats() map[string]int {
	stats := make(map[string]int)
	allNodes := nodeCache.GetAll()
	for _, n := range allNodes {
		source := n.Source
		if source == "" || source == "manual" {
			source = "手动添加"
		}
		stats[source]++
	}
	return stats
}

// ========== 节点字段元数据反射 ==========

// NodeFieldMeta 节点字段元数据
type NodeFieldMeta struct {
	Name  string `json:"name"`  // 字段名称
	Label string `json:"label"` // 显示标签
	Type  string `json:"type"`  // 字段类型
}

// 全局缓存
var nodeFieldsMetaCache []NodeFieldMeta

// InitNodeFieldsMeta 系统启动时调用，通过反射扫描Node结构体
func InitNodeFieldsMeta() {
	// 跳过不适合去重的字段
	skipFields := map[string]bool{
		"ID": true, "Link": true, "CreatedAt": true, "UpdatedAt": true,
		"Tags": true, "SpeedCheckAt": true, "LatencyCheckAt": true,
		"Speed": true, "DelayTime": true, "SpeedStatus": true, "DelayStatus": true,
	}

	// 字段中文标签映射
	labelMap := map[string]string{
		"Name":            "备注",
		"LinkName":        "原始名称",
		"LinkAddress":     "完整地址",
		"LinkHost":        "服务器地址",
		"LinkPort":        "端口",
		"LinkCountry":     "国家代码",
		"LandingIP":       "落地IP",
		"DialerProxyName": "前置代理",
		"Source":          "来源",
		"SourceID":        "来源ID",
		"Group":           "分组",
	}

	t := reflect.TypeOf(Node{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// 跳过不适合去重的字段
		if skipFields[field.Name] {
			continue
		}

		kind := field.Type.Kind()
		// 只提取string和int类型字段用于去重
		if kind != reflect.String && kind != reflect.Int {
			continue
		}

		// 获取中文标签
		label := labelMap[field.Name]
		if label == "" {
			label = field.Name
		}

		fieldType := "string"
		if kind == reflect.Int {
			fieldType = "int"
		}

		nodeFieldsMetaCache = append(nodeFieldsMetaCache, NodeFieldMeta{
			Name:  field.Name,
			Label: label,
			Type:  fieldType,
		})
	}

	utils.Info("节点字段元数据初始化完成，共 %d 个字段可用于去重", len(nodeFieldsMetaCache))
}

// GetNodeFieldsMeta 获取缓存的节点字段元数据
func GetNodeFieldsMeta() []NodeFieldMeta {
	return nodeFieldsMetaCache
}

// GetFieldValue 根据字段名获取节点字段值（使用反射）
func (node *Node) GetFieldValue(fieldName string) string {
	v := reflect.ValueOf(*node)
	f := v.FieldByName(fieldName)
	if !f.IsValid() {
		return ""
	}
	switch f.Kind() {
	case reflect.String:
		return f.String()
	case reflect.Int, reflect.Int64:
		return fmt.Sprintf("%d", f.Int())
	default:
		return ""
	}
}
