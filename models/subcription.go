package models

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/dto"
	"sublink/utils"
	"time"

	"gorm.io/gorm"
)

// subcriptionCache 使用新的泛型缓存
var subcriptionCache *cache.MapCache[int, Subcription]

func init() {
	subcriptionCache = cache.NewMapCache(func(s Subcription) int { return s.ID })
	subcriptionCache.AddIndex("name", func(s Subcription) string { return s.Name })
}

type Subcription struct {
	ID                 int
	Name               string
	Config             string    `gorm:"embedded"`
	Nodes              []Node    `gorm:"-" json:"-"`
	SubLogs            []SubLogs `gorm:"foreignKey:SubcriptionID;"` // 一对多关系 约束父表被删除子表记录跟着删除
	CreateDate         string
	NodesWithSort      []NodeWithSort   `gorm:"-" json:"Nodes"`
	Groups             []string         `gorm:"-" json:"-"`      // 内部使用，不返回给前端
	GroupsWithSort     []GroupWithSort  `gorm:"-" json:"Groups"` // 订阅关联的分组列表（带Sort）
	Scripts            []Script         `gorm:"-" json:"-"`      // 内部使用
	ScriptsWithSort    []ScriptWithSort `gorm:"-" json:"Scripts"`
	IPWhitelist        string           `json:"IPWhitelist"`        //IP白名单
	IPBlacklist        string           `json:"IPBlacklist"`        //IP黑名单
	DelayTime          int              `json:"DelayTime"`          // 最大延迟(ms)
	MinSpeed           float64          `json:"MinSpeed"`           // 最小速度(MB/s)
	CountryWhitelist   string           `json:"CountryWhitelist"`   // 国家白名单（逗号分隔）
	CountryBlacklist   string           `json:"CountryBlacklist"`   // 国家黑名单（逗号分隔）
	NodeNameRule       string           `json:"NodeNameRule"`       // 节点命名规则模板
	NodeNamePreprocess string           `json:"NodeNamePreprocess"` // 原名预处理规则 (JSON数组)
	NodeNameWhitelist  string           `json:"NodeNameWhitelist"`  // 节点名称白名单 (JSON数组)
	NodeNameBlacklist  string           `json:"NodeNameBlacklist"`  // 节点名称黑名单 (JSON数组)
	TagWhitelist       string           `json:"TagWhitelist"`       // 标签白名单（逗号分隔）
	TagBlacklist       string           `json:"TagBlacklist"`       // 标签黑名单（逗号分隔）
	CreatedAt          time.Time        `json:"CreatedAt"`
	UpdatedAt          time.Time        `json:"UpdatedAt"`
	DeletedAt          gorm.DeletedAt   `gorm:"index" json:"DeletedAt"`
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

// InitSubcriptionCache 初始化订阅缓存
func InitSubcriptionCache() error {
	log.Printf("开始加载订阅到缓存")
	var subs []Subcription
	if err := database.DB.Find(&subs).Error; err != nil {
		return err
	}

	subcriptionCache.LoadAll(subs)
	log.Printf("订阅缓存初始化完成，共加载 %d 个订阅", subcriptionCache.Count())

	cache.Manager.Register("subcription", subcriptionCache)
	return nil
}

// Add 添加订阅 (Write-Through)
func (sub *Subcription) Add() error {
	err := database.DB.Create(sub).Error
	if err != nil {
		return err
	}
	subcriptionCache.Set(sub.ID, *sub)
	return nil
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
		if err := database.DB.Create(&subNode).Error; err != nil {
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
		if err := database.DB.Create(&subGroup).Error; err != nil {
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
		if err := database.DB.Create(&subScript).Error; err != nil {
			return err
		}
	}
	return nil
}

// 更新订阅 (Write-Through)
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
		"node_name_whitelist":  sub.NodeNameWhitelist,
		"node_name_blacklist":  sub.NodeNameBlacklist,
		"tag_whitelist":        sub.TagWhitelist,
		"tag_blacklist":        sub.TagBlacklist,
	}
	err := database.DB.Model(&Subcription{}).Where("id = ? or name = ?", sub.ID, sub.Name).Updates(updates).Error
	if err != nil {
		return err
	}
	// 更新缓存：从数据库读取完整数据后更新
	var updated Subcription
	if err := database.DB.First(&updated, sub.ID).Error; err == nil {
		subcriptionCache.Set(sub.ID, updated)
	}
	return nil
}

// 更新节点列表建立多对多关系（使用节点名称）
func (sub *Subcription) UpdateNodes() error {
	// 先删除旧的关联
	if err := database.DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionNode{}).Error; err != nil {
		return err
	}
	// 再添加新的关联
	for i, node := range sub.Nodes {
		subNode := SubcriptionNode{
			SubcriptionID: sub.ID,
			NodeName:      node.Name,
			Sort:          i, // 按添加顺序设置排序
		}
		if err := database.DB.Create(&subNode).Error; err != nil {
			return err
		}
	}
	return nil
}

// 更新分组列表
func (sub *Subcription) UpdateGroups(groups []string) error {
	// 先删除旧的关联
	if err := database.DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionGroup{}).Error; err != nil {
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
		if err := database.DB.Create(&subGroup).Error; err != nil {
			return err
		}
	}
	return nil
}

// UpdateScripts 更新脚本关联
func (sub *Subcription) UpdateScripts(scriptIDs []int) error {
	// 先删除旧的关联
	if err := database.DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionScript{}).Error; err != nil {
		return err
	}
	// 再添加新的关联
	for i, scriptID := range scriptIDs {
		subScript := SubcriptionScript{
			SubcriptionID: sub.ID,
			ScriptID:      scriptID,
			Sort:          i,
		}
		if err := database.DB.Create(&subScript).Error; err != nil {
			return err
		}
	}
	return nil
}

// 查找订阅（优先从缓存查找）
func (sub *Subcription) Find() error {
	// 优先从缓存查找
	if sub.ID > 0 {
		if cached, ok := subcriptionCache.Get(sub.ID); ok {
			*sub = cached
			return nil
		}
	}
	if sub.Name != "" {
		subs := subcriptionCache.GetByIndex("name", sub.Name)
		if len(subs) > 0 {
			*sub = subs[0]
			return nil
		}
	}
	// 缓存未命中，查数据库
	err := database.DB.Where("id = ? or name = ?", sub.ID, sub.Name).First(sub).Error
	if err != nil {
		return err
	}
	// 更新缓存
	subcriptionCache.Set(sub.ID, *sub)
	return nil
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
	err := database.DB.Table("nodes").
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
	err = database.DB.Table("subcription_groups").
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
		err = database.DB.Table("nodes").
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

	// 标签过滤（在节点名称过滤之前）
	if sub.TagWhitelist != "" || sub.TagBlacklist != "" {
		// 解析白名单和黑名单标签
		whitelistTags := make(map[string]bool)
		blacklistTags := make(map[string]bool)

		if sub.TagWhitelist != "" {
			for _, t := range strings.Split(sub.TagWhitelist, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					whitelistTags[t] = true
				}
			}
		}

		if sub.TagBlacklist != "" {
			for _, t := range strings.Split(sub.TagBlacklist, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					blacklistTags[t] = true
				}
			}
		}

		var filteredNodes []Node
		for _, node := range sub.Nodes {
			nodeTags := node.GetTagNames()

			// 黑名单优先：如果节点任一标签在黑名单中，则排除
			if len(blacklistTags) > 0 {
				isBlacklisted := false
				for _, nt := range nodeTags {
					if blacklistTags[nt] {
						isBlacklisted = true
						break
					}
				}
				if isBlacklisted {
					continue
				}
			}

			// 白名单：如果设置了白名单，节点必须包含至少一个白名单标签
			if len(whitelistTags) > 0 {
				isWhitelisted := false
				for _, nt := range nodeTags {
					if whitelistTags[nt] {
						isWhitelisted = true
						break
					}
				}
				if !isWhitelisted {
					continue
				}
			}

			filteredNodes = append(filteredNodes, node)
		}
		sub.Nodes = filteredNodes
	}

	// 节点名称过滤 (基于原始节点名称 LinkName)
	// 使用 HasActiveNodeNameFilter 检查是否有有效规则，避免空数组 "[]" 导致过滤异常
	hasWhitelistRules := utils.HasActiveNodeNameFilter(sub.NodeNameWhitelist)
	hasBlacklistRules := utils.HasActiveNodeNameFilter(sub.NodeNameBlacklist)

	if hasWhitelistRules || hasBlacklistRules {
		var filteredNodes []Node
		for _, node := range sub.Nodes {
			// 黑名单优先：如果匹配黑名单则排除
			if hasBlacklistRules && utils.MatchesNodeNameFilter(sub.NodeNameBlacklist, node.LinkName) {
				continue
			}

			// 白名单：如果设置了白名单但不匹配白名单，则排除
			if hasWhitelistRules && !utils.MatchesNodeNameFilter(sub.NodeNameWhitelist, node.LinkName) {
				continue
			}

			filteredNodes = append(filteredNodes, node)
		}
		sub.Nodes = filteredNodes
	}

	// 获取脚本信息及其排序
	var scriptsWithSort []ScriptWithSort
	err = database.DB.Table("scripts").
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

// 订阅列表（从缓存获取，批量加载关联数据解决 N+1）

func (sub *Subcription) List() ([]Subcription, error) {
	// 从缓存获取所有订阅
	subs := subcriptionCache.GetAllSorted(func(a, b Subcription) bool {
		return a.ID < b.ID
	})

	if len(subs) == 0 {
		return subs, nil
	}

	// 批量加载关联数据
	if err := batchLoadSubcriptionRelations(subs); err != nil {
		return nil, err
	}

	return subs, nil
}

// ListPaginated 分页获取订阅列表（从缓存分页，批量加载关联数据）
func (sub *Subcription) ListPaginated(page, pageSize int) ([]Subcription, int64, error) {
	// 从缓存获取所有订阅并排序
	allSubs := subcriptionCache.GetAllSorted(func(a, b Subcription) bool {
		return a.ID < b.ID
	})
	total := int64(len(allSubs))

	// 如果不需要分页，返回全部
	if page <= 0 || pageSize <= 0 {
		if err := batchLoadSubcriptionRelations(allSubs); err != nil {
			return nil, 0, err
		}
		return allSubs, total, nil
	}

	// 分页
	offset := (page - 1) * pageSize
	if offset >= len(allSubs) {
		return []Subcription{}, total, nil
	}

	end := offset + pageSize
	if end > len(allSubs) {
		end = len(allSubs)
	}

	subs := allSubs[offset:end]

	// 批量加载关联数据
	if err := batchLoadSubcriptionRelations(subs); err != nil {
		return nil, 0, err
	}

	return subs, total, nil
}

// batchLoadSubcriptionRelations 批量加载订阅的关联数据（解决 N+1 问题）
func batchLoadSubcriptionRelations(subs []Subcription) error {
	if len(subs) == 0 {
		return nil
	}

	// 收集所有订阅 ID
	subIDs := make([]int, len(subs))
	subIDMap := make(map[int]int) // subID -> index in subs
	for i, s := range subs {
		subIDs[i] = s.ID
		subIDMap[s.ID] = i
	}

	// 1. 批量查询所有订阅的节点关联
	var subNodes []SubcriptionNode
	if err := database.DB.Where("subcription_id IN ?", subIDs).Order("sort ASC").Find(&subNodes).Error; err != nil {
		return err
	}

	// 按订阅 ID 分组节点名称
	subNodeNames := make(map[int][]struct {
		Name string
		Sort int
	})
	for _, sn := range subNodes {
		subNodeNames[sn.SubcriptionID] = append(subNodeNames[sn.SubcriptionID], struct {
			Name string
			Sort int
		}{sn.NodeName, sn.Sort})
	}

	// 使用节点缓存获取节点详情
	for i := range subs {
		nodeInfos := subNodeNames[subs[i].ID]
		nodesWithSort := make([]NodeWithSort, 0, len(nodeInfos))
		for _, ni := range nodeInfos {
			if node, ok := GetNodeByName(ni.Name); ok {
				nodesWithSort = append(nodesWithSort, NodeWithSort{
					Node: *node,
					Sort: ni.Sort,
				})
			}
		}
		// 按 Sort 排序
		sort.Slice(nodesWithSort, func(a, b int) bool {
			return nodesWithSort[a].Sort < nodesWithSort[b].Sort
		})
		subs[i].NodesWithSort = nodesWithSort
	}

	// 2. 批量查询所有订阅的分组关联
	var subGroups []SubcriptionGroup
	if err := database.DB.Where("subcription_id IN ?", subIDs).Order("sort ASC").Find(&subGroups).Error; err != nil {
		return err
	}

	// 按订阅 ID 分组
	subGroupMap := make(map[int][]GroupWithSort)
	for _, sg := range subGroups {
		subGroupMap[sg.SubcriptionID] = append(subGroupMap[sg.SubcriptionID], GroupWithSort{
			Name: sg.GroupName,
			Sort: sg.Sort,
		})
	}
	for i := range subs {
		subs[i].GroupsWithSort = subGroupMap[subs[i].ID]
		if subs[i].GroupsWithSort == nil {
			subs[i].GroupsWithSort = []GroupWithSort{}
		}
	}

	// 3. 批量查询所有订阅的脚本关联
	var subScripts []SubcriptionScript
	if err := database.DB.Where("subcription_id IN ?", subIDs).Order("sort ASC").Find(&subScripts).Error; err != nil {
		return err
	}

	// 按订阅 ID 分组脚本 ID
	subScriptIDs := make(map[int][]struct {
		ScriptID int
		Sort     int
	})
	for _, ss := range subScripts {
		subScriptIDs[ss.SubcriptionID] = append(subScriptIDs[ss.SubcriptionID], struct {
			ScriptID int
			Sort     int
		}{ss.ScriptID, ss.Sort})
	}

	// 使用脚本缓存获取脚本详情
	for i := range subs {
		scriptInfos := subScriptIDs[subs[i].ID]
		scriptsWithSort := make([]ScriptWithSort, 0, len(scriptInfos))
		for _, si := range scriptInfos {
			if script, err := GetScriptByID(si.ScriptID); err == nil {
				scriptsWithSort = append(scriptsWithSort, ScriptWithSort{
					Script: *script,
					Sort:   si.Sort,
				})
			}
		}
		// 按 Sort 排序
		sort.Slice(scriptsWithSort, func(a, b int) bool {
			return scriptsWithSort[a].Sort < scriptsWithSort[b].Sort
		})
		subs[i].ScriptsWithSort = scriptsWithSort
	}

	// 4. 批量获取日志（使用缓存）
	for i := range subs {
		subs[i].SubLogs = GetSubLogsBySubcriptionID(subs[i].ID)
	}

	return nil
}

func (sub *Subcription) IPlogUpdate() error {
	return database.DB.Model(sub).Association("SubLogs").Replace(&sub.SubLogs)
}

// 删除订阅（硬删除，Write-Through）
func (sub *Subcription) Del() error {
	// 先删除关联的节点关系
	if err := database.DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionNode{}).Error; err != nil {
		return err
	}
	// 删除关联的分组关系
	if err := database.DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionGroup{}).Error; err != nil {
		return err
	}
	// 删除关联的脚本关系
	if err := database.DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionScript{}).Error; err != nil {
		return err
	}
	// 硬删除订阅本身（Unscoped 绕过软删除）
	err := database.DB.Unscoped().Delete(sub).Error
	if err != nil {
		return err
	}
	// 从缓存中删除
	subcriptionCache.Delete(sub.ID)
	return nil
}

func (sub *Subcription) Sort(subNodeSort dto.SubcriptionNodeSortUpdate) error {
	tx := database.DB.Begin()
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
