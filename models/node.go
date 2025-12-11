package models

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sublink/cache"
	"time"

	"gorm.io/gorm/clause"
)

type Node struct {
	ID              int    `gorm:"primaryKey"`
	Link            string `gorm:"uniqueIndex:idx_link_id"` //出站代理原始连接
	Name            string //系统内节点名称
	LinkName        string //节点原始名称
	LinkAddress     string //节点原始地址
	LinkHost        string //节点原始Host
	LinkPort        string //节点原始端口
	LinkCountry     string //节点所属国家、落地IP国家
	DialerProxyName string
	Source          string `gorm:"default:'manual'"`
	SourceID        int
	Group           string
	Speed           float64   `gorm:"default:0"` // 测速结果(MB/s)
	DelayTime       int       `gorm:"default:0"` // 延迟时间(ms)
	LastCheck       string    // 最后检测时间
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
	nodeCache.AddIndex("sourceID", func(n Node) string { return fmt.Sprintf("%d", n.SourceID) })
	nodeCache.AddIndex("name", func(n Node) string { return n.Name })
}

// InitNodeCache 初始化节点缓存
func InitNodeCache() error {
	log.Printf("加载节点列表到缓存")
	var nodes []Node
	if err := DB.Find(&nodes).Error; err != nil {
		return err
	}

	// 使用批量加载方式初始化缓存
	nodeCache.LoadAll(nodes)
	log.Printf("节点缓存初始化完成，共加载 %d 个节点", nodeCache.Count())

	// 注册到缓存管理器
	cache.Manager.Register("node", nodeCache)
	return nil
}

// Add 添加节点
func (node *Node) Add() error {
	// Write-Through: 先写数据库
	err := DB.Create(node).Error
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
	err := DB.Model(node).Select("Name", "Link", "DialerProxyName", "Group", "LinkName", "LinkAddress", "LinkHost", "LinkPort", "LinkCountry", "UpdatedAt").Updates(node).Error
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
		if err := DB.First(&fullNode, node.ID).Error; err == nil {
			nodeCache.Set(node.ID, fullNode)
		}
	}
	return nil
}

// UpdateSpeed 更新节点测速结果
func (node *Node) UpdateSpeed() error {
	err := DB.Model(node).Select("Speed", "LinkCountry", "DelayTime", "LastCheck").Updates(node).Error
	if err != nil {
		return err
	}

	if cachedNode, ok := nodeCache.Get(node.ID); ok {
		cachedNode.Speed = node.Speed
		cachedNode.DelayTime = node.DelayTime
		cachedNode.LastCheck = node.LastCheck
		cachedNode.LinkCountry = node.LinkCountry
		nodeCache.Set(node.ID, cachedNode)
	}
	return nil
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
	err := DB.Where("link = ? or name = ?", node.Link, node.Name).First(node).Error
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
	err := DB.First(node, node.ID).Error
	if err != nil {
		return err
	}

	// 更新缓存
	nodeCache.Set(node.ID, *node)
	return nil
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
	Search    string   // 搜索关键词（匹配节点名称或链接）
	Group     string   // 分组过滤
	Source    string   // 来源过滤
	MaxDelay  int      // 最大延迟(ms)，只显示延迟在此值以下的节点
	MinSpeed  float64  // 最低速度(MB/s)，只显示速度在此值以上的节点
	Countries []string // 国家代码过滤
	Tags      []string // 标签过滤（匹配任一标签的节点）
	SortBy    string   // 排序字段: "delay" 或 "speed"
	SortOrder string   // 排序顺序: "asc" 或 "desc"
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
				groupLower := strings.ToLower(n.Group)
				filterGroupLower := strings.ToLower(filter.Group)
				if !strings.Contains(groupLower, filterGroupLower) {
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
				sourceLower := strings.ToLower(n.Source)
				filterSourceLower := strings.ToLower(filter.Source)
				if !strings.Contains(sourceLower, filterSourceLower) {
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
func (node *Node) ListByGroups(groups []string) ([]Node, error) {
	groupMap := make(map[string]bool)
	for _, g := range groups {
		groupMap[g] = true
	}

	// 使用缓存的 Filter 方法
	nodes := nodeCache.Filter(func(n Node) bool {
		return groupMap[n.Group]
	})
	return nodes, nil
}

// ListByTags 根据标签获取节点列表（匹配任意标签）
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

	// 使用缓存的 Filter 方法
	nodes := nodeCache.Filter(func(n Node) bool {
		nodeTags := n.GetTagNames()
		for _, nt := range nodeTags {
			if tagMap[nt] {
				return true
			}
		}
		return false
	})
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
	if err := DB.Exec("DELETE FROM subcription_nodes WHERE node_name = ?", node.Name).Error; err != nil {
		return err
	}
	// Write-Through: 先删除数据库
	err := DB.Delete(node).Error
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
	err := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "link"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "link_name", "link_address", "link_host", "link_port", "link_country", "source", "source_id", "group"}),
	}).Create(node).Error
	if err != nil {
		return err
	}

	// 查询更新后的节点并更新缓存
	var updatedNode Node
	if err := DB.Where("link = ?", node.Link).First(&updatedNode).Error; err == nil {
		nodeCache.Set(updatedNode.ID, updatedNode)
		*node = updatedNode
	}
	return nil
}

// DeleteAutoSubscriptionNodes 删除订阅节点
func DeleteAutoSubscriptionNodes(sourceId int) error {
	// 使用二级索引获取要删除的节点
	nodesToDelete := nodeCache.GetByIndex("sourceID", strconv.Itoa(sourceId))
	nodeNames := make([]string, 0, len(nodesToDelete))
	for _, n := range nodesToDelete {
		nodeNames = append(nodeNames, n.Name)
	}

	// 清除节点与订阅的关联关系
	if len(nodeNames) > 0 {
		if err := DB.Exec("DELETE FROM subcription_nodes WHERE node_name IN ?", nodeNames).Error; err != nil {
			return err
		}
	}

	// Write-Through: 先删除数据库
	err := DB.Where("source_id = ?", sourceId).Delete(&Node{}).Error
	if err != nil {
		return err
	}

	// 再更新缓存
	for _, n := range nodesToDelete {
		nodeCache.Delete(n.ID)
	}
	return nil
}

// BatchDel 批量删除节点
func BatchDel(ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	// 获取节点名称列表
	nodeNames := make([]string, 0, len(ids))
	for _, id := range ids {
		if n, ok := nodeCache.Get(id); ok {
			nodeNames = append(nodeNames, n.Name)
		}
	}

	// 先清除节点与订阅的关联关系
	if len(nodeNames) > 0 {
		if err := DB.Exec("DELETE FROM subcription_nodes WHERE node_name IN ?", nodeNames).Error; err != nil {
			return err
		}
	}

	// Write-Through: 先删除数据库
	if err := DB.Where("id IN ?", ids).Delete(&Node{}).Error; err != nil {
		return err
	}

	// 再更新缓存
	for _, id := range ids {
		nodeCache.Delete(id)
	}
	return nil
}

// BatchUpdateGroup 批量更新节点分组
func BatchUpdateGroup(ids []int, group string) error {
	if len(ids) == 0 {
		return nil
	}

	// Write-Through: 先更新数据库
	if err := DB.Model(&Node{}).Where("id IN ?", ids).Update("group", group).Error; err != nil {
		return err
	}

	// 再更新缓存
	for _, id := range ids {
		if n, ok := nodeCache.Get(id); ok {
			n.Group = group
			nodeCache.Set(id, n)
		}
	}
	return nil
}

// BatchUpdateDialerProxy 批量更新节点前置代理
func BatchUpdateDialerProxy(ids []int, dialerProxyName string) error {
	if len(ids) == 0 {
		return nil
	}

	// Write-Through: 先更新数据库
	if err := DB.Model(&Node{}).Where("id IN ?", ids).Update("dialer_proxy_name", dialerProxyName).Error; err != nil {
		return err
	}

	// 再更新缓存
	for _, id := range ids {
		if n, ok := nodeCache.Get(id); ok {
			n.DialerProxyName = dialerProxyName
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
	if err := DB.Where("delay_time > 0 AND speed > 0").Order("delay_time ASC").Limit(1).Find(&dbNodes).Error; err != nil {
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
	if err := DB.Where("source_id = ?", sourceID).Find(&nodes).Error; err != nil {
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
	if err := DB.Model(&Node{}).Where("source_id = ?", sourceID).Updates(updateFields).Error; err != nil {
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
		protocol := parseProtocolFromLink(n.Link)
		stats[protocol]++
	}
	return stats
}

// parseProtocolFromLink 从节点链接中解析协议类型
func parseProtocolFromLink(link string) string {
	if link == "" {
		return "未知"
	}
	// 常见协议前缀映射
	protocolPrefixes := map[string]string{
		"ss://":        "Shadowsocks",
		"ssr://":       "ShadowsocksR",
		"vmess://":     "VMess",
		"vless://":     "VLESS",
		"trojan://":    "Trojan",
		"hysteria://":  "Hysteria",
		"hysteria2://": "Hysteria2",
		"hy2://":       "Hysteria2",
		"tuic://":      "TUIC",
		"wg://":        "WireGuard",
		"wireguard://": "WireGuard",
		"naive://":     "NaiveProxy",
		"http://":      "HTTP",
		"https://":     "HTTPS",
		"socks://":     "SOCKS",
		"socks5://":    "SOCKS5",
	}

	for prefix, name := range protocolPrefixes {
		if len(link) >= len(prefix) && link[:len(prefix)] == prefix {
			return name
		}
	}
	return "其他"
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
