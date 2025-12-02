package models

import (
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
	DelayTime       int    `gorm:"default:0"` // 延迟时间(ms)
	LastCheck       string // 最后检测时间
}

// Add 添加节点
func (node *Node) Add() error {
	return DB.Create(node).Error
}

// 更新节点
func (node *Node) Update() error {
	return DB.Model(node).Select("Name", "Link", "DialerProxyName", "Group", "LinkName", "LinkAddress", "LinkHost", "LinkPort").Updates(node).Error
}

// UpdateSpeed 更新节点测速结果
func (node *Node) UpdateSpeed() error {
	return DB.Model(node).Select("Speed", "DelayTime", "LastCheck").Updates(node).Error
}

// 查找节点是否重复
func (node *Node) Find() error {
	return DB.Where("link = ? or name = ?", node.Link, node.Name).First(node).Error
}

// 节点列表
func (node *Node) List() ([]Node, error) {
	var nodes []Node
	err := DB.Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// ListByGroups 根据分组获取节点列表
func (node *Node) ListByGroups(groups []string) ([]Node, error) {
	var nodes []Node
	err := DB.Where("`group` IN ?", groups).Find(&nodes).Error
	if err != nil {
		return nil, err
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
	return DB.Delete(node).Error
}

// UpsertNode 插入或更新节点
func (node *Node) UpsertNode() error {
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "link"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "link_name", "link_address", "link_host", "link_port", "create_date", "source", "source_id", "group"}),
	}).Create(node).Error
}

// DeleteAutoSubscriptionNodes 删除订阅节点
func DeleteAutoSubscriptionNodes(sourceId int) error {
	return DB.Where("source_id = ?", sourceId).Delete(&Node{}).Error
}

// GetAllGroups 获取所有分组
func (node *Node) GetAllGroups() ([]string, error) {
	var groups []string
	err := DB.Model(&Node{}).
		Where("`group` IS NOT NULL AND `group` != ''").
		Distinct("`group`").
		Pluck("`group`", &groups).Error
	return groups, err
}
