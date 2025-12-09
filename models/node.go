package models

import (
	"log"
	"sort"
	"strings"
	"sync"
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
}

var (
	nodeCache = make(map[int]Node)
	nodeLock  sync.RWMutex
)

// InitNodeCache 初始化节点缓存
func InitNodeCache() error {
	log.Printf("加载节点列表到缓存")
	var nodes []Node
	if err := DB.Find(&nodes).Error; err != nil {
		return err
	}

	nodeLock.Lock()
	defer nodeLock.Unlock()

	// 清空旧缓存
	nodeCache = make(map[int]Node)

	for _, n := range nodes {
		nodeCache[n.ID] = n
		log.Printf("加载节点【%s】到缓存成功", n.Name)
	}
	log.Printf("节点缓存初始化完成，共加载 %d 个节点", len(nodes))
	return nil
}

// Add 添加节点
func (node *Node) Add() error {
	err := DB.Create(node).Error
	if err != nil {
		return err
	}
	// 更新缓存
	nodeLock.Lock()
	nodeCache[node.ID] = *node
	nodeLock.Unlock()
	return nil
}

// 更新节点
func (node *Node) Update() error {
	if node.Name == "" {
		node.Name = node.LinkName
	}
	node.UpdatedAt = time.Now()
	err := DB.Model(node).Select("Name", "Link", "DialerProxyName", "Group", "LinkName", "LinkAddress", "LinkHost", "LinkPort", "LinkCountry", "UpdatedAt").Updates(node).Error
	if err != nil {
		return err
	}
	// 更新缓存：先获取完整节点信息，或者只更新变动字段。
	// 为简单起见，这里假设 Update 调用者已经设置了 node 的 ID。
	// 但 Updates 只更新了部分字段，内存中的 node 可能不完整。
	// 最稳妥的方式是重新从 DB 读取一次，或者只更新缓存中的对应字段。
	// 这里选择更新缓存中的对应字段。

	nodeLock.Lock()
	defer nodeLock.Unlock()

	if cachedNode, ok := nodeCache[node.ID]; ok {
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
		nodeCache[node.ID] = cachedNode
	} else {
		// 如果缓存中没有，可能是新加的或者缓存未同步，尝试从 DB 读
		var fullNode Node
		if err := DB.First(&fullNode, node.ID).Error; err == nil {
			nodeCache[node.ID] = fullNode
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

	nodeLock.Lock()
	defer nodeLock.Unlock()

	if cachedNode, ok := nodeCache[node.ID]; ok {
		cachedNode.Speed = node.Speed
		cachedNode.DelayTime = node.DelayTime
		cachedNode.LastCheck = node.LastCheck
		cachedNode.LinkCountry = node.LinkCountry
		nodeCache[node.ID] = cachedNode
	}
	return nil
}

// 查找节点是否重复
func (node *Node) Find() error {
	// 优先查缓存
	nodeLock.RLock()
	for _, n := range nodeCache {
		if n.Link == node.Link || n.Name == node.Name {
			*node = n
			nodeLock.RUnlock()
			return nil
		}
	}
	nodeLock.RUnlock()

	// 缓存未命中，查 DB
	err := DB.Where("link = ? or name = ?", node.Link, node.Name).First(node).Error
	if err != nil {
		return err
	}

	// 更新缓存
	nodeLock.Lock()
	nodeCache[node.ID] = *node
	nodeLock.Unlock()

	return nil
}

// GetByID 根据ID查找节点
func (node *Node) GetByID() error {
	nodeLock.RLock()
	if cachedNode, ok := nodeCache[node.ID]; ok {
		*node = cachedNode
		nodeLock.RUnlock()
		return nil
	}
	nodeLock.RUnlock()

	err := DB.First(node, node.ID).Error
	if err != nil {
		return err
	}

	nodeLock.Lock()
	nodeCache[node.ID] = *node
	nodeLock.Unlock()

	return nil
}

// 节点列表
func (node *Node) List() ([]Node, error) {
	nodeLock.RLock()
	defer nodeLock.RUnlock()

	nodes := make([]Node, 0, len(nodeCache))
	for _, n := range nodeCache {
		nodes = append(nodes, n)
	}

	// 按 ID 升序排序
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})

	return nodes, nil
}

// NodeFilter 节点过滤参数
type NodeFilter struct {
	Search    string   // 搜索关键词（匹配节点名称或链接）
	Group     string   // 分组过滤
	Source    string   // 来源过滤
	MaxDelay  int      // 最大延迟(ms)，只显示延迟在此值以下的节点
	MinSpeed  float64  // 最低速度(MB/s)，只显示速度在此值以上的节点
	Countries []string // 国家代码过滤
	SortBy    string   // 排序字段: "delay" 或 "speed"
	SortOrder string   // 排序顺序: "asc" 或 "desc"
}

// ListWithFilters 根据过滤条件获取节点列表
func (node *Node) ListWithFilters(filter NodeFilter) ([]Node, error) {
	nodeLock.RLock()
	defer nodeLock.RUnlock()

	// 预处理搜索关键词
	searchLower := strings.ToLower(filter.Search)

	// 创建国家代码映射，加速查找
	countryMap := make(map[string]bool)
	for _, c := range filter.Countries {
		countryMap[c] = true
	}

	// 过滤节点
	nodes := make([]Node, 0, len(nodeCache))
	for _, n := range nodeCache {
		// 搜索过滤
		if searchLower != "" {
			nameLower := strings.ToLower(n.Name)
			linkLower := strings.ToLower(n.Link)
			if !strings.Contains(nameLower, searchLower) && !strings.Contains(linkLower, searchLower) {
				continue
			}
		}

		// 分组过滤
		if filter.Group != "" {
			if filter.Group == "未分组" {
				if n.Group != "" {
					continue
				}
			} else {
				groupLower := strings.ToLower(n.Group)
				filterGroupLower := strings.ToLower(filter.Group)
				if !strings.Contains(groupLower, filterGroupLower) {
					continue
				}
			}
		}

		// 来源过滤
		if filter.Source != "" {
			if filter.Source == "手动添加" {
				if n.Source != "" && n.Source != "manual" {
					continue
				}
			} else {
				sourceLower := strings.ToLower(n.Source)
				filterSourceLower := strings.ToLower(filter.Source)
				if !strings.Contains(sourceLower, filterSourceLower) {
					continue
				}
			}
		}

		// 最大延迟过滤 - 只过滤延迟大于0的节点
		if filter.MaxDelay > 0 {
			if n.DelayTime <= 0 || n.DelayTime > filter.MaxDelay {
				continue
			}
		}

		// 最低速度过滤
		if filter.MinSpeed > 0 {
			if n.Speed <= filter.MinSpeed {
				continue
			}
		}

		// 国家代码过滤
		if len(countryMap) > 0 {
			if n.LinkCountry == "" || !countryMap[n.LinkCountry] {
				continue
			}
		}

		nodes = append(nodes, n)
	}

	// 排序
	if filter.SortBy != "" {
		sort.Slice(nodes, func(i, j int) bool {
			switch filter.SortBy {
			case "delay":
				// 没有有效延迟的节点始终放最后
				aValid := nodes[i].DelayTime > 0
				bValid := nodes[j].DelayTime > 0
				if !aValid && !bValid {
					return nodes[i].ID < nodes[j].ID // 都无效时按ID排序
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
				// 没有有效速度的节点始终放最后
				aValid := nodes[i].Speed > 0
				bValid := nodes[j].Speed > 0
				if !aValid && !bValid {
					return nodes[i].ID < nodes[j].ID // 都无效时按ID排序
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
		// 默认按 ID 升序排序
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].ID < nodes[j].ID
		})
	}

	return nodes, nil
}

// ListByGroups 根据分组获取节点列表
func (node *Node) ListByGroups(groups []string) ([]Node, error) {
	nodeLock.RLock()
	defer nodeLock.RUnlock()

	var nodes []Node
	groupMap := make(map[string]bool)
	for _, g := range groups {
		groupMap[g] = true
	}

	for _, n := range nodeCache {
		if groupMap[n.Group] {
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

// 删除节点
func (node *Node) Del() error {
	// 先清除节点与订阅的关联关系（通过节点名称）
	if err := DB.Exec("DELETE FROM subcription_nodes WHERE node_name = ?", node.Name).Error; err != nil {
		return err
	}
	// 再删除节点本身
	err := DB.Delete(node).Error
	if err != nil {
		return err
	}

	nodeLock.Lock()
	delete(nodeCache, node.ID)
	nodeLock.Unlock()

	return nil
}

// UpsertNode 插入或更新节点
func (node *Node) UpsertNode() error {
	err := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "link"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "link_name", "link_address", "link_host", "link_port", "link_country", "source", "source_id", "group"}),
	}).Create(node).Error
	if err != nil {
		return err
	}

	// Upsert 后 ID 可能变了（如果是插入），或者 ID 没变（如果是更新）
	// 最简单的是重新根据 Link 查一次，或者直接重新加载所有缓存（开销大）
	// 这里尝试查询更新后的节点
	var updatedNode Node
	if err := DB.Where("link = ?", node.Link).First(&updatedNode).Error; err == nil {
		nodeLock.Lock()
		nodeCache[updatedNode.ID] = updatedNode
		nodeLock.Unlock()
		*node = updatedNode // 更新传入的 node 对象
	}

	return nil
}

// DeleteAutoSubscriptionNodes 删除订阅节点
func DeleteAutoSubscriptionNodes(sourceId int) error {
	// 先获取要删除的节点名称列表，用于清理订阅关联
	nodeLock.RLock()
	nodeNames := make([]string, 0)
	for _, n := range nodeCache {
		if n.SourceID == sourceId {
			nodeNames = append(nodeNames, n.Name)
		}
	}
	nodeLock.RUnlock()

	// 清除节点与订阅的关联关系
	if len(nodeNames) > 0 {
		if err := DB.Exec("DELETE FROM subcription_nodes WHERE node_name IN ?", nodeNames).Error; err != nil {
			return err
		}
	}

	// 删除节点
	err := DB.Where("source_id = ?", sourceId).Delete(&Node{}).Error
	if err != nil {
		return err
	}

	nodeLock.Lock()
	defer nodeLock.Unlock()

	// 遍历删除缓存中对应的节点
	for id, n := range nodeCache {
		if n.SourceID == sourceId {
			delete(nodeCache, id)
		}
	}
	return nil
}

// BatchDel 批量删除节点
func BatchDel(ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	// 获取节点名称列表，用于删除订阅关联
	nodeLock.RLock()
	nodeNames := make([]string, 0, len(ids))
	for _, id := range ids {
		if n, ok := nodeCache[id]; ok {
			nodeNames = append(nodeNames, n.Name)
		}
	}
	nodeLock.RUnlock()

	// 先清除节点与订阅的关联关系
	if len(nodeNames) > 0 {
		if err := DB.Exec("DELETE FROM subcription_nodes WHERE node_name IN ?", nodeNames).Error; err != nil {
			return err
		}
	}

	// 批量删除节点
	if err := DB.Where("id IN ?", ids).Delete(&Node{}).Error; err != nil {
		return err
	}

	// 更新缓存
	nodeLock.Lock()
	defer nodeLock.Unlock()
	for _, id := range ids {
		delete(nodeCache, id)
	}

	return nil
}

// GetAllGroups 获取所有分组
func (node *Node) GetAllGroups() ([]string, error) {
	nodeLock.RLock()
	defer nodeLock.RUnlock()

	groupMap := make(map[string]bool)
	for _, n := range nodeCache {
		if n.Group != "" {
			groupMap[n.Group] = true
		}
	}

	groups := make([]string, 0, len(groupMap))
	for g := range groupMap {
		groups = append(groups, g)
	}
	return groups, nil
}

// GetAllSources 获取所有来源
func (node *Node) GetAllSources() ([]string, error) {
	nodeLock.RLock()
	defer nodeLock.RUnlock()

	sourceMap := make(map[string]bool)
	for _, n := range nodeCache {
		if n.Source != "" {
			sourceMap[n.Source] = true
		}
	}

	sources := make([]string, 0, len(sourceMap))
	for s := range sourceMap {
		sources = append(sources, s)
	}
	return sources, nil
}

// GetBestProxyNode 获取最佳代理节点（延迟最低且速度大于0）
func GetBestProxyNode() (*Node, error) {
	nodeLock.RLock()
	defer nodeLock.RUnlock()

	var bestNode *Node
	for _, n := range nodeCache {
		if n.DelayTime > 0 && n.Speed > 0 {
			if bestNode == nil || n.DelayTime < bestNode.DelayTime {
				nodeCopy := n
				bestNode = &nodeCopy
			}
		}
	}

	if bestNode != nil {
		return bestNode, nil
	}

	// 缓存中没有符合条件的节点，从数据库查询
	var nodes []Node
	if err := DB.Where("delay_time > 0 AND speed > 0").Order("delay_time ASC").Limit(1).Find(&nodes).Error; err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nil, nil
	}

	return &nodes[0], nil
}

// ListBySourceID 根据订阅ID查询节点列表
func ListBySourceID(sourceID int) ([]Node, error) {
	nodeLock.RLock()
	defer nodeLock.RUnlock()

	var nodes []Node
	for _, n := range nodeCache {
		if n.SourceID == sourceID {
			nodes = append(nodes, n)
		}
	}

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
	// 更新数据库中的节点
	updateFields := map[string]interface{}{
		"source": sourceName,
		"group":  group,
	}
	if err := DB.Model(&Node{}).Where("source_id = ?", sourceID).Updates(updateFields).Error; err != nil {
		return err
	}

	// 更新缓存中的节点
	nodeLock.Lock()
	defer nodeLock.Unlock()

	for id, n := range nodeCache {
		if n.SourceID == sourceID {
			n.Source = sourceName
			n.Group = group
			nodeCache[id] = n
		}
	}

	return nil
}

// GetFastestSpeedNode 获取最快速度节点
func GetFastestSpeedNode() *Node {
	nodeLock.RLock()
	defer nodeLock.RUnlock()

	var fastest *Node
	for _, n := range nodeCache {
		if n.Speed > 0 {
			if fastest == nil || n.Speed > fastest.Speed {
				nodeCopy := n
				fastest = &nodeCopy
			}
		}
	}

	return fastest
}

// GetLowestDelayNode 获取最低延迟节点
func GetLowestDelayNode() *Node {
	nodeLock.RLock()
	defer nodeLock.RUnlock()

	var lowest *Node
	for _, n := range nodeCache {
		if n.DelayTime > 0 {
			if lowest == nil || n.DelayTime < lowest.DelayTime {
				nodeCopy := n
				lowest = &nodeCopy
			}
		}
	}

	return lowest
}

// GetAllCountries 获取所有唯一的国家代码
func GetAllCountries() []string {
	nodeLock.RLock()
	defer nodeLock.RUnlock()

	countryMap := make(map[string]bool)
	for _, n := range nodeCache {
		if n.LinkCountry != "" {
			countryMap[n.LinkCountry] = true
		}
	}

	countries := make([]string, 0, len(countryMap))
	for c := range countryMap {
		countries = append(countries, c)
	}
	return countries
}

// GetNodeCountryStats 获取按国家统计的节点数量
func GetNodeCountryStats() map[string]int {
	nodeLock.RLock()
	defer nodeLock.RUnlock()

	stats := make(map[string]int)
	for _, n := range nodeCache {
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
	nodeLock.RLock()
	defer nodeLock.RUnlock()

	stats := make(map[string]int)
	for _, n := range nodeCache {
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
