package models

import (
	"fmt"
	"strings"
	"sublink/dto"
	"time"

	"gorm.io/gorm"
)

type Subcription struct {
	ID               int
	Name             string
	Config           string    `gorm:"embedded"`
	Nodes            []Node    `gorm:"-" json:"-"`
	SubLogs          []SubLogs `gorm:"foreignKey:SubcriptionID;"` // 一对多关系 约束父表被删除子表记录跟着删除
	CreateDate       string
	NodesWithSort    []NodeWithSort   `gorm:"-" json:"Nodes"`
	Groups           []string         `gorm:"-" json:"-"`      // 内部使用，不返回给前端
	GroupsWithSort   []GroupWithSort  `gorm:"-" json:"Groups"` // 订阅关联的分组列表（带Sort）
	Scripts          []Script         `gorm:"-" json:"-"`      // 内部使用
	ScriptsWithSort  []ScriptWithSort `gorm:"-" json:"Scripts"`
	IPWhitelist      string           `json:"IPWhitelist"`      //IP白名单
	IPBlacklist      string           `json:"IPBlacklist"`      //IP黑名单
	DelayTime        int              `json:"DelayTime"`        // 最大延迟(ms)
	MinSpeed         float64          `json:"MinSpeed"`         // 最小速度(MB/s)
	CountryWhitelist string           `json:"CountryWhitelist"` // 国家白名单（逗号分隔）
	CountryBlacklist string           `json:"CountryBlacklist"` // 国家黑名单（逗号分隔）
	NodeNameRule       string           `json:"NodeNameRule"`       // 节点命名规则模板
	NodeNamePreprocess string           `json:"NodeNamePreprocess"` // 原名预处理规则 (JSON数组)
	CreatedAt        time.Time        `json:"CreatedAt"`
	UpdatedAt        time.Time        `json:"UpdatedAt"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"DeletedAt"`
}

type GroupWithSort struct {
	Name string `json:"Name"`
	Sort int    `json:"Sort"`
}

type ScriptWithSort struct {
	Script
	Sort int `json:"Sort"`
}

type SubcriptionNode struct {
	SubcriptionID int    `gorm:"primaryKey"`
	NodeName      string `gorm:"primaryKey"`
	Sort          int    `gorm:"default:0"`
}

// SubcriptionGroup 订阅与分组关联表
type SubcriptionGroup struct {
	SubcriptionID int    `gorm:"primaryKey"`
	GroupName     string `gorm:"primaryKey"`
	Sort          int    `gorm:"default:0"`
}

// SubcriptionScript 订阅与脚本关联表
type SubcriptionScript struct {
	SubcriptionID int `gorm:"primaryKey"`
	ScriptID      int `gorm:"primaryKey"`
	Sort          int `gorm:"default:0"`
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

// 添加分组列表建立关系
func (sub *Subcription) AddGroups(groups []string) error {
	for i, groupName := range groups {
		if groupName == "" {
			continue
		}
		subGroup := SubcriptionGroup{
			SubcriptionID: sub.ID,
			GroupName:     groupName,
			Sort:          i,
		}
		if err := DB.Create(&subGroup).Error; err != nil {
			return err
		}
	}
	return nil
}

// AddScripts 添加脚本关联
func (sub *Subcription) AddScripts(scriptIDs []int) error {
	for i, scriptID := range scriptIDs {
		subScript := SubcriptionScript{
			SubcriptionID: sub.ID,
			ScriptID:      scriptID,
			Sort:          i,
		}
		if err := DB.Create(&subScript).Error; err != nil {
			return err
		}
	}
	return nil
}

// 更新订阅
func (sub *Subcription) Update() error {
	updates := map[string]interface{}{
		"name":                 sub.Name,
		"config":               sub.Config,
		"create_date":          sub.CreateDate,
		"ip_whitelist":         sub.IPWhitelist,
		"ip_blacklist":         sub.IPBlacklist,
		"delay_time":           sub.DelayTime,
		"min_speed":            sub.MinSpeed,
		"country_whitelist":    sub.CountryWhitelist,
		"country_blacklist":    sub.CountryBlacklist,
		"node_name_rule":       sub.NodeNameRule,
		"node_name_preprocess": sub.NodeNamePreprocess,
	}
	return DB.Model(&Subcription{}).Where("id = ? or name = ?", sub.ID, sub.Name).Updates(updates).Error
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

// 更新分组列表
func (sub *Subcription) UpdateGroups(groups []string) error {
	// 先删除旧的关联
	if err := DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionGroup{}).Error; err != nil {
		return err
	}
	// 再添加新的关联
	for i, groupName := range groups {
		if groupName == "" {
			continue
		}
		subGroup := SubcriptionGroup{
			SubcriptionID: sub.ID,
			GroupName:     groupName,
			Sort:          i,
		}
		if err := DB.Create(&subGroup).Error; err != nil {
			return err
		}
	}
	return nil
}

// UpdateScripts 更新脚本关联
func (sub *Subcription) UpdateScripts(scriptIDs []int) error {
	// 先删除旧的关联
	if err := DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionScript{}).Error; err != nil {
		return err
	}
	// 再添加新的关联
	for i, scriptID := range scriptIDs {
		subScript := SubcriptionScript{
			SubcriptionID: sub.ID,
			ScriptID:      scriptID,
			Sort:          i,
		}
		if err := DB.Create(&subScript).Error; err != nil {
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
	// 定义节点排序项结构
	type NodeSortItem struct {
		Node
		Sort    int
		IsGroup bool
	}

	// 获取直接选择的节点及其排序
	var directNodeItems []NodeSortItem
	err := DB.Table("nodes").
		Select("nodes.*, subcription_nodes.sort, 0 as is_group").
		Joins("left join subcription_nodes ON subcription_nodes.node_name = nodes.name").
		Where("subcription_nodes.subcription_id = ?", sub.ID).
		Scan(&directNodeItems).Error
	if err != nil {
		return err
	}

	// 获取分组信息及其排序
	var groups []struct {
		GroupName string
		Sort      int
	}
	err = DB.Table("subcription_groups").
		Select("group_name, sort").
		Where("subcription_id = ?", sub.ID).
		Scan(&groups).Error
	if err != nil {
		return err
	}

	// 获取通过分组动态选择的节点
	groupNodeMap := make(map[string][]Node) // groupName -> nodes
	for _, group := range groups {
		var groupNodes []Node
		err = DB.Table("nodes").
			Where("nodes.`group` = ?", group.GroupName).
			Order("nodes.id ASC").
			Find(&groupNodes).Error
		if err != nil {
			return err
		}
		groupNodeMap[group.GroupName] = groupNodes
	}

	// 创建一个混合列表，包含节点和分组
	type MixedItem struct {
		Node    *Node
		Group   string
		Sort    int
		IsGroup bool
	}

	var mixedItems []MixedItem

	// 添加直接选择的节点
	for _, item := range directNodeItems {
		node := item.Node
		mixedItems = append(mixedItems, MixedItem{
			Node:    &node,
			Sort:    item.Sort,
			IsGroup: false,
		})
	}

	// 添加分组
	for _, group := range groups {
		mixedItems = append(mixedItems, MixedItem{
			Group:   group.GroupName,
			Sort:    group.Sort,
			IsGroup: true,
		})
	}

	// 按排序值排序混合列表
	// 使用简单的冒泡排序
	for i := 0; i < len(mixedItems); i++ {
		for j := i + 1; j < len(mixedItems); j++ {
			if mixedItems[i].Sort > mixedItems[j].Sort {
				mixedItems[i], mixedItems[j] = mixedItems[j], mixedItems[i]
			}
		}
	}

	// 按排序后的顺序构建最终节点列表
	nodeMap := make(map[string]bool) // 用于去重
	sub.Nodes = make([]Node, 0)

	for _, item := range mixedItems {
		if item.IsGroup {
			// 添加分组中的所有节点
			if nodes, exists := groupNodeMap[item.Group]; exists {
				for _, node := range nodes {
					if !nodeMap[node.Name] {
						sub.Nodes = append(sub.Nodes, node)
						nodeMap[node.Name] = true
					}
				}
			}
		} else {
			// 添加单个节点
			if item.Node != nil && !nodeMap[item.Node.Name] {
				sub.Nodes = append(sub.Nodes, *item.Node)
				nodeMap[item.Node.Name] = true
			}
		}
	}

	// 过滤节点
	if sub.DelayTime > 0 || sub.MinSpeed > 0 {
		var filteredNodes []Node
		for _, node := range sub.Nodes {
			// 检查延迟 (DelayTime)
			if sub.DelayTime > 0 {
				// 如果节点没有测速数据(DelayTime=0)或者延迟超过限制，则跳过
				if node.DelayTime <= 0 || node.DelayTime > sub.DelayTime {
					continue
				}
			}

			// 检查速度 (MinSpeed)
			if sub.MinSpeed > 0 {
				// 如果节点没有测速数据(Speed=0)或者速度小于限制，则跳过
				if node.Speed < sub.MinSpeed {
					continue
				}
			}

			filteredNodes = append(filteredNodes, node)
		}
		sub.Nodes = filteredNodes
	}

	// 国家代码过滤
	if sub.CountryWhitelist != "" || sub.CountryBlacklist != "" {
		// 解析白名单和黑名单
		whitelistMap := make(map[string]bool)
		blacklistMap := make(map[string]bool)

		if sub.CountryWhitelist != "" {
			for _, c := range strings.Split(sub.CountryWhitelist, ",") {
				c = strings.TrimSpace(c)
				if c != "" {
					whitelistMap[strings.ToUpper(c)] = true
				}
			}
		}

		if sub.CountryBlacklist != "" {
			for _, c := range strings.Split(sub.CountryBlacklist, ",") {
				c = strings.TrimSpace(c)
				if c != "" {
					blacklistMap[strings.ToUpper(c)] = true
				}
			}
		}

		var filteredNodes []Node
		for _, node := range sub.Nodes {
			country := strings.ToUpper(node.LinkCountry)

			// 黑名单优先：如果在黑名单中则排除
			if len(blacklistMap) > 0 && blacklistMap[country] {
				continue
			}

			// 白名单：如果设置了白名单但节点不在白名单中，则排除
			if len(whitelistMap) > 0 && !whitelistMap[country] {
				continue
			}

			filteredNodes = append(filteredNodes, node)
		}
		sub.Nodes = filteredNodes
	}

	// 获取脚本信息及其排序
	var scriptsWithSort []ScriptWithSort
	err = DB.Table("scripts").
		Select("scripts.*, subcription_scripts.sort").
		Joins("LEFT JOIN subcription_scripts ON subcription_scripts.script_id = scripts.id").
		Where("subcription_scripts.subcription_id = ?", sub.ID).
		Order("subcription_scripts.sort ASC").
		Scan(&scriptsWithSort).Error
	if err != nil {
		return err
	}
	sub.ScriptsWithSort = scriptsWithSort

	return nil
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

		// 查询订阅关联的分组（带Sort字段）
		var groups []SubcriptionGroup
		err = DB.Where("subcription_id = ?", subs[i].ID).
			Order("sort ASC").
			Find(&groups).Error
		if err != nil {
			return nil, err
		}
		subs[i].GroupsWithSort = make([]GroupWithSort, 0, len(groups))
		for _, g := range groups {
			subs[i].GroupsWithSort = append(subs[i].GroupsWithSort, GroupWithSort{
				Name: g.GroupName,
				Sort: g.Sort,
			})
		}

		// 查询订阅关联的脚本（带Sort字段）
		var scriptsWithSort []ScriptWithSort
		err = DB.Table("scripts").
			Select("scripts.*, subcription_scripts.sort").
			Joins("LEFT JOIN subcription_scripts ON subcription_scripts.script_id = scripts.id").
			Where("subcription_scripts.subcription_id = ?", subs[i].ID).
			Order("subcription_scripts.sort ASC").
			Scan(&scriptsWithSort).Error
		if err != nil {
			return nil, err
		}
		subs[i].ScriptsWithSort = scriptsWithSort

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

// 删除订阅（硬删除）
func (sub *Subcription) Del() error {
	// 先删除关联的节点关系
	if err := DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionNode{}).Error; err != nil {
		return err
	}
	// 删除关联的分组关系
	if err := DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionGroup{}).Error; err != nil {
		return err
	}
	// 删除关联的脚本关系
	if err := DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionScript{}).Error; err != nil {
		return err
	}
	// 硬删除订阅本身（Unscoped 绕过软删除）
	return DB.Unscoped().Delete(sub).Error
}

func (sub *Subcription) Sort(subNodeSort dto.SubcriptionNodeSortUpdate) error {
	tx := DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开启事务失败: %w", tx.Error)
	}

	for _, item := range subNodeSort.NodeSort {
		// 判断是节点还是分组
		isGroup := item.IsGroup != nil && *item.IsGroup

		if isGroup {
			// 更新分组排序
			err := tx.Model(&SubcriptionGroup{}).
				Where("subcription_id = ? AND group_name = ?", subNodeSort.ID, item.Name).
				Update("sort", item.Sort).Error

			if err != nil {
				tx.Rollback()
				return fmt.Errorf("更新分组排序失败: %w", err)
			}
		} else {
			// 更新节点排序
			err := tx.Model(&SubcriptionNode{}).
				Where("subcription_id = ? AND node_name = ?", subNodeSort.ID, item.Name).
				Update("sort", item.Sort).Error

			if err != nil {
				tx.Rollback()
				return fmt.Errorf("更新节点排序失败: %w", err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}
	return nil
}
