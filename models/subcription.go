package models

import (
	"fmt"
	"sublink/dto"
)

type Subcription struct {
	ID            int
	Name          string
	Config        string    `gorm:"embedded"`
	Nodes         []Node    `gorm:"-" json:"-"`
	SubLogs       []SubLogs `gorm:"foreignKey:SubcriptionID;"` // 一对多关系 约束父表被删除子表记录跟着删除
	CreateDate    string
	NodesWithSort []NodeWithSort `gorm:"-" json:"Nodes"`
}

type SubcriptionNode struct {
	SubcriptionID int    `gorm:"primaryKey"`
	NodeName      string `gorm:"primaryKey"`
	Sort          int    `gorm:"default:0"`
}

type NodeWithSort struct {
	Node
	Sort int `json:"Sort"`
}

// Add 添加订阅
func (sub *Subcription) Add() error {
	return DB.Create(sub).Error
}

// 添加节点列表建立多对多关系（使用节点名称）
func (sub *Subcription) AddNode() error {
	// 手动插入中间表记录，使用节点名称
	for i, node := range sub.Nodes {
		subNode := SubcriptionNode{
			SubcriptionID: sub.ID,
			NodeName:      node.Name,
			Sort:          i, // 按添加顺序设置排序
		}
		if err := DB.Create(&subNode).Error; err != nil {
			return err
		}
	}
	return nil
}

// 更新订阅
func (sub *Subcription) Update() error {
	return DB.Where("id = ? or name = ?", sub.ID, sub.Name).Updates(sub).Error
}

// 更新节点列表建立多对多关系（使用节点名称）
func (sub *Subcription) UpdateNodes() error {
	// 先删除旧的关联
	if err := DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionNode{}).Error; err != nil {
		return err
	}
	// 再添加新的关联
	for i, node := range sub.Nodes {
		subNode := SubcriptionNode{
			SubcriptionID: sub.ID,
			NodeName:      node.Name,
			Sort:          i, // 按添加顺序设置排序
		}
		if err := DB.Create(&subNode).Error; err != nil {
			return err
		}
	}
	return nil
}

// 查找订阅
func (sub *Subcription) Find() error {
	return DB.Where("id = ? or name = ?", sub.ID, sub.Name).First(sub).Error
}

// 读取订阅
func (sub *Subcription) GetSub() error {
	// err := DB.Find(sub).Error
	// if err != nil {
	// 	return err
	// }
	return DB.Table("nodes").
		Joins("left join subcription_nodes ON subcription_nodes.node_name = nodes.name").
		Where("subcription_nodes.subcription_id = ?", sub.ID).
		Order("subcription_nodes.sort ASC").Find(&sub.Nodes).Error
}

// 订阅列表

func (sub *Subcription) List() ([]Subcription, error) {
	var subs []Subcription
	// 先查所有订阅
	err := DB.Find(&subs).Error
	if err != nil {
		return nil, err
	}

	for i := range subs {
		// 查询订阅对应的节点和sort字段，按sort和node.id排序
		var nodesWithSort []NodeWithSort
		err := DB.Table("nodes").
			Select("nodes.*, subcription_nodes.sort").
			Joins("LEFT JOIN subcription_nodes ON subcription_nodes.node_name = nodes.name").
			Where("subcription_nodes.subcription_id = ?", subs[i].ID).
			Order("subcription_nodes.sort ASC").
			Scan(&nodesWithSort).Error
		if err != nil {
			return nil, err
		}
		subs[i].NodesWithSort = nodesWithSort

		// 查询日志
		err = DB.Model(&subs[i]).Association("SubLogs").Find(&subs[i].SubLogs)
		if err != nil {
			return nil, err
		}
	}

	return subs, nil
}

func (sub *Subcription) IPlogUpdate() error {
	return DB.Model(sub).Association("SubLogs").Replace(&sub.SubLogs)
}

// 删除订阅
func (sub *Subcription) Del() error {
	err := DB.Model(sub).Association("Nodes").Clear()
	if err != nil {
		return err
	}
	return DB.Delete(sub).Error
}

func (sub *Subcription) Sort(subNodeSort dto.SubcriptionNodeSortUpdate) error {
	tx := DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开启事务失败: %w", tx.Error)
	}
	for _, item := range subNodeSort.NodeSort {
		err := tx.Model(&SubcriptionNode{}).
			Where("subcription_id = ? AND node_name = ?", subNodeSort.ID, item.Name).
			Update("sort", item.Sort).Error

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("更新节点排序失败: %w", err)
		}
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}
	return nil
}
