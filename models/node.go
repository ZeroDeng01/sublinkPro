package models

import (
	"log"
	"sort"
	"sync"

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
	DialerProxyName string
	CreateDate      string
	Source          string `gorm:"default:'manual'"`
	SourceID        int
	Group           string
	Speed           float64 `gorm:"default:0"` // 测速结果(MB/s)
	DelayTime       int     `gorm:"default:0"` // 延迟时间(ms)
	LastCheck       string  // 最后检测时间
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
	err := DB.Model(node).Select("Name", "Link", "DialerProxyName", "Group", "LinkName", "LinkAddress", "LinkHost", "LinkPort").Updates(node).Error
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
	err := DB.Model(node).Select("Speed", "DelayTime", "LastCheck").Updates(node).Error
	if err != nil {
		return err
	}

	nodeLock.Lock()
	defer nodeLock.Unlock()

	if cachedNode, ok := nodeCache[node.ID]; ok {
		cachedNode.Speed = node.Speed
		cachedNode.DelayTime = node.DelayTime
		cachedNode.LastCheck = node.LastCheck
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
		DoUpdates: clause.AssignmentColumns([]string{"name", "link_name", "link_address", "link_host", "link_port", "create_date", "source", "source_id", "group"}),
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
