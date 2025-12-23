package models

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/dto"
	"sublink/node/protocol"
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
	ID                    int
	Name                  string
	Config                string    `gorm:"embedded"`
	Nodes                 []Node    `gorm:"-" json:"-"`
	SubLogs               []SubLogs `gorm:"foreignKey:SubcriptionID;"` // 一对多关系 约束父表被删除子表记录跟着删除
	CreateDate            string
	NodesWithSort         []NodeWithSort   `gorm:"-" json:"Nodes"`
	Groups                []string         `gorm:"-" json:"-"`      // 内部使用，不返回给前端
	GroupsWithSort        []GroupWithSort  `gorm:"-" json:"Groups"` // 订阅关联的分组列表（带Sort）
	Scripts               []Script         `gorm:"-" json:"-"`      // 内部使用
	ScriptsWithSort       []ScriptWithSort `gorm:"-" json:"Scripts"`
	IPWhitelist           string           `json:"IPWhitelist"`                               //IP白名单
	IPBlacklist           string           `json:"IPBlacklist"`                               //IP黑名单
	DelayTime             int              `json:"DelayTime"`                                 // 最大延迟(ms)
	MinSpeed              float64          `json:"MinSpeed"`                                  // 最小速度(MB/s)
	CountryWhitelist      string           `json:"CountryWhitelist"`                          // 国家白名单（逗号分隔）
	CountryBlacklist      string           `json:"CountryBlacklist"`                          // 国家黑名单（逗号分隔）
	NodeNameRule          string           `json:"NodeNameRule"`                              // 节点命名规则模板
	NodeNamePreprocess    string           `json:"NodeNamePreprocess"`                        // 原名预处理规则 (JSON数组)
	NodeNameWhitelist     string           `json:"NodeNameWhitelist"`                         // 节点名称白名单 (JSON数组)
	NodeNameBlacklist     string           `json:"NodeNameBlacklist"`                         // 节点名称黑名单 (JSON数组)
	TagWhitelist          string           `json:"TagWhitelist"`                              // 标签白名单（逗号分隔）
	TagBlacklist          string           `json:"TagBlacklist"`                              // 标签黑名单（逗号分隔）
	ProtocolWhitelist     string           `json:"ProtocolWhitelist"`                         // 协议白名单（逗号分隔）
	ProtocolBlacklist     string           `json:"ProtocolBlacklist"`                         // 协议黑名单（逗号分隔）
	DeduplicationRule     string           `json:"DeduplicationRule"`                         // 去重规则配置(JSON)
	RefreshUsageOnRequest bool             `gorm:"default:true" json:"RefreshUsageOnRequest"` // 获取订阅时是否实时刷新用量信息
	CreatedAt             time.Time        `json:"CreatedAt"`
	UpdatedAt             time.Time        `json:"UpdatedAt"`
	DeletedAt             gorm.DeletedAt   `gorm:"index" json:"DeletedAt"`
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
	SubcriptionID int `gorm:"primaryKey"`
	NodeID        int `gorm:"primaryKey"` // 使用节点 ID 关联
	Sort          int `gorm:"default:0"`
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
	utils.Info("开始加载订阅到缓存")
	var subs []Subcription
	if err := database.DB.Find(&subs).Error; err != nil {
		return err
	}

	subcriptionCache.LoadAll(subs)
	utils.Info("订阅缓存初始化完成，共加载 %d 个订阅", subcriptionCache.Count())

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

// 添加节点列表建立多对多关系（使用节点 ID）
func (sub *Subcription) AddNode() error {
	// 手动插入中间表记录，使用节点 ID
	for i, node := range sub.Nodes {
		subNode := SubcriptionNode{
			SubcriptionID: sub.ID,
			NodeID:        node.ID,
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
		"name":                     sub.Name,
		"config":                   sub.Config,
		"create_date":              sub.CreateDate,
		"ip_whitelist":             sub.IPWhitelist,
		"ip_blacklist":             sub.IPBlacklist,
		"delay_time":               sub.DelayTime,
		"min_speed":                sub.MinSpeed,
		"country_whitelist":        sub.CountryWhitelist,
		"country_blacklist":        sub.CountryBlacklist,
		"node_name_rule":           sub.NodeNameRule,
		"node_name_preprocess":     sub.NodeNamePreprocess,
		"node_name_whitelist":      sub.NodeNameWhitelist,
		"node_name_blacklist":      sub.NodeNameBlacklist,
		"tag_whitelist":            sub.TagWhitelist,
		"tag_blacklist":            sub.TagBlacklist,
		"protocol_whitelist":       sub.ProtocolWhitelist,
		"protocol_blacklist":       sub.ProtocolBlacklist,
		"deduplication_rule":       sub.DeduplicationRule,
		"refresh_usage_on_request": sub.RefreshUsageOnRequest,
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

// 更新节点列表建立多对多关系（使用节点 ID）
func (sub *Subcription) UpdateNodes() error {
	// 先删除旧的关联
	if err := database.DB.Where("subcription_id = ?", sub.ID).Delete(&SubcriptionNode{}).Error; err != nil {
		return err
	}
	// 再添加新的关联
	for i, node := range sub.Nodes {
		subNode := SubcriptionNode{
			SubcriptionID: sub.ID,
			NodeID:        node.ID,
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

// ApplyFilters 对节点列表应用过滤条件
// 抽取共用的过滤逻辑，供 GetSub 和 PreviewSub 调用
// 返回过滤后的节点列表
func (sub *Subcription) ApplyFilters(nodes []Node) []Node {
	result := nodes

	// 1. 延迟和速度过滤
	if sub.DelayTime > 0 || sub.MinSpeed > 0 {
		var filteredNodes []Node
		for _, node := range result {
			if sub.DelayTime > 0 {
				if node.DelayTime <= 0 || node.DelayTime > sub.DelayTime {
					continue
				}
			}
			if sub.MinSpeed > 0 {
				if node.Speed < sub.MinSpeed {
					continue
				}
			}
			filteredNodes = append(filteredNodes, node)
		}
		result = filteredNodes
	}

	// 2. 国家代码过滤
	if sub.CountryWhitelist != "" || sub.CountryBlacklist != "" {
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
		for _, node := range result {
			country := strings.ToUpper(node.LinkCountry)
			// 黑名单优先
			if len(blacklistMap) > 0 && blacklistMap[country] {
				continue
			}
			// 白名单
			if len(whitelistMap) > 0 && !whitelistMap[country] {
				continue
			}
			filteredNodes = append(filteredNodes, node)
		}
		result = filteredNodes
	}

	// 3. 标签过滤
	if sub.TagWhitelist != "" || sub.TagBlacklist != "" {
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
		for _, node := range result {
			nodeTags := node.GetTagNames()

			// 黑名单优先
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

			// 白名单
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
		result = filteredNodes
	}

	// 4. 节点名称过滤
	hasWhitelistRules := utils.HasActiveNodeNameFilter(sub.NodeNameWhitelist)
	hasBlacklistRules := utils.HasActiveNodeNameFilter(sub.NodeNameBlacklist)

	if hasWhitelistRules || hasBlacklistRules {
		var filteredNodes []Node
		for _, node := range result {
			// 黑名单优先
			if hasBlacklistRules && utils.MatchesNodeNameFilter(sub.NodeNameBlacklist, node.LinkName) {
				continue
			}
			// 白名单
			if hasWhitelistRules && !utils.MatchesNodeNameFilter(sub.NodeNameWhitelist, node.LinkName) {
				continue
			}
			filteredNodes = append(filteredNodes, node)
		}
		result = filteredNodes
	}

	// 5. 协议过滤
	if sub.ProtocolWhitelist != "" || sub.ProtocolBlacklist != "" {
		whitelistProtos := make(map[string]bool)
		blacklistProtos := make(map[string]bool)

		if sub.ProtocolWhitelist != "" {
			for _, p := range strings.Split(sub.ProtocolWhitelist, ",") {
				p = strings.TrimSpace(strings.ToLower(p))
				if p != "" {
					whitelistProtos[p] = true
				}
			}
		}

		if sub.ProtocolBlacklist != "" {
			for _, p := range strings.Split(sub.ProtocolBlacklist, ",") {
				p = strings.TrimSpace(strings.ToLower(p))
				if p != "" {
					blacklistProtos[p] = true
				}
			}
		}

		var filteredNodes []Node
		for _, node := range result {
			nodeProto := strings.ToLower(node.Protocol)
			// 黑名单优先
			if len(blacklistProtos) > 0 && blacklistProtos[nodeProto] {
				continue
			}
			// 白名单
			if len(whitelistProtos) > 0 && !whitelistProtos[nodeProto] {
				continue
			}
			filteredNodes = append(filteredNodes, node)
		}
		result = filteredNodes
	}

	// 6. 应用去重规则
	result = sub.ApplyDeduplication(result)

	return result
}

// 读取订阅
func (sub *Subcription) GetSub(clientType string) error {
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
		Joins("left join subcription_nodes ON subcription_nodes.node_id = nodes.id").
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

	// 调用共用的过滤方法
	sub.Nodes = sub.ApplyFilters(sub.Nodes)

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

	// 执行节点过滤脚本
	sub.Nodes = sub.ApplyNodeFilterScripts(sub.Nodes, scriptsWithSort, clientType)

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

	// 按订阅 ID 分组节点 ID
	subNodeIDs := make(map[int][]struct {
		ID   int
		Sort int
	})
	for _, sn := range subNodes {
		subNodeIDs[sn.SubcriptionID] = append(subNodeIDs[sn.SubcriptionID], struct {
			ID   int
			Sort int
		}{sn.NodeID, sn.Sort})
	}

	// 使用节点缓存获取节点详情
	for i := range subs {
		nodeInfos := subNodeIDs[subs[i].ID]
		nodesWithSort := make([]NodeWithSort, 0, len(nodeInfos))
		for _, ni := range nodeInfos {
			if node, ok := nodeCache.Get(ni.ID); ok {
				nodesWithSort = append(nodesWithSort, NodeWithSort{
					Node: node,
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
	// 删除关联的链式代理规则
	if err := DeleteChainRulesBySubscriptionID(sub.ID); err != nil {
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

// Copy 复制订阅及其关联数据（节点、分组、脚本）
// 新订阅名称为：原名称_复制
func (sub *Subcription) Copy() (*Subcription, error) {
	// 创建新订阅对象，复制所有配置字段
	newSub := &Subcription{
		Name:                  sub.Name + "_复制",
		Config:                sub.Config,
		CreateDate:            time.Now().Format("2006-01-02 15:04:05"),
		IPWhitelist:           sub.IPWhitelist,
		IPBlacklist:           sub.IPBlacklist,
		DelayTime:             sub.DelayTime,
		MinSpeed:              sub.MinSpeed,
		CountryWhitelist:      sub.CountryWhitelist,
		CountryBlacklist:      sub.CountryBlacklist,
		NodeNameRule:          sub.NodeNameRule,
		NodeNamePreprocess:    sub.NodeNamePreprocess,
		NodeNameWhitelist:     sub.NodeNameWhitelist,
		NodeNameBlacklist:     sub.NodeNameBlacklist,
		TagWhitelist:          sub.TagWhitelist,
		TagBlacklist:          sub.TagBlacklist,
		ProtocolWhitelist:     sub.ProtocolWhitelist,
		ProtocolBlacklist:     sub.ProtocolBlacklist,
		DeduplicationRule:     sub.DeduplicationRule,
		RefreshUsageOnRequest: sub.RefreshUsageOnRequest,
	}

	// 使用事务确保数据一致性
	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("开启事务失败: %w", tx.Error)
	}

	// 创建新订阅
	if err := tx.Create(newSub).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建订阅失败: %w", err)
	}

	// 复制节点关联
	var nodes []SubcriptionNode
	if err := database.DB.Where("subcription_id = ?", sub.ID).Find(&nodes).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("查询节点关联失败: %w", err)
	}
	for _, node := range nodes {
		newNode := SubcriptionNode{
			SubcriptionID: newSub.ID,
			NodeID:        node.NodeID,
			Sort:          node.Sort,
		}
		if err := tx.Create(&newNode).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("复制节点关联失败: %w", err)
		}
	}

	// 复制分组关联
	var groups []SubcriptionGroup
	if err := database.DB.Where("subcription_id = ?", sub.ID).Find(&groups).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("查询分组关联失败: %w", err)
	}
	for _, group := range groups {
		newGroup := SubcriptionGroup{
			SubcriptionID: newSub.ID,
			GroupName:     group.GroupName,
			Sort:          group.Sort,
		}
		if err := tx.Create(&newGroup).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("复制分组关联失败: %w", err)
		}
	}

	// 复制脚本关联
	var scripts []SubcriptionScript
	if err := database.DB.Where("subcription_id = ?", sub.ID).Find(&scripts).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("查询脚本关联失败: %w", err)
	}
	for _, script := range scripts {
		newScript := SubcriptionScript{
			SubcriptionID: newSub.ID,
			ScriptID:      script.ScriptID,
			Sort:          script.Sort,
		}
		if err := tx.Create(&newScript).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("复制脚本关联失败: %w", err)
		}
	}

	// 复制链式代理规则
	chainRules := GetChainRulesBySubscriptionID(sub.ID)
	for _, rule := range chainRules {
		newRule := SubscriptionChainRule{
			SubscriptionID: newSub.ID,
			Name:           rule.Name,
			Sort:           rule.Sort,
			Enabled:        rule.Enabled,
			ChainConfig:    rule.ChainConfig,
			TargetConfig:   rule.TargetConfig,
		}
		if err := tx.Create(&newRule).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("复制链式代理规则失败: %w", err)
		}
		// 更新缓存
		chainRuleCache.Set(newRule.ID, newRule)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 更新缓存
	subcriptionCache.Set(newSub.ID, *newSub)

	return newSub, nil
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
				Where("subcription_id = ? AND node_id = ?", subNodeSort.ID, item.ID).
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

// BatchSort 批量排序订阅节点
// sortBy: source(来源), name(名称), protocol(协议), delay(延迟), speed(速度), country(地区)
// sortOrder: asc(升序), desc(降序)
func (sub *Subcription) BatchSort(sortBy, sortOrder string) error {
	// 获取订阅的所有节点关联（带节点详情）
	var subNodes []struct {
		SubcriptionNode
		Node Node `gorm:"embedded"`
	}

	err := database.DB.Table("subcription_nodes").
		Select("subcription_nodes.*, nodes.*").
		Joins("LEFT JOIN nodes ON nodes.id = subcription_nodes.node_id").
		Where("subcription_nodes.subcription_id = ?", sub.ID).
		Scan(&subNodes).Error
	if err != nil {
		return fmt.Errorf("获取节点关联失败: %w", err)
	}

	if len(subNodes) == 0 {
		return nil
	}

	// 按指定规则排序
	sort.Slice(subNodes, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "source":
			less = subNodes[i].Node.Source < subNodes[j].Node.Source
		case "name":
			less = subNodes[i].Node.Name < subNodes[j].Node.Name
		case "protocol":
			less = subNodes[i].Node.Protocol < subNodes[j].Node.Protocol
		case "delay":
			// 未测试的节点排最后
			if subNodes[i].Node.DelayTime <= 0 && subNodes[j].Node.DelayTime <= 0 {
				less = subNodes[i].Node.ID < subNodes[j].Node.ID
			} else if subNodes[i].Node.DelayTime <= 0 {
				less = false
			} else if subNodes[j].Node.DelayTime <= 0 {
				less = true
			} else {
				less = subNodes[i].Node.DelayTime < subNodes[j].Node.DelayTime
			}
		case "speed":
			// 未测试的节点排最后
			if subNodes[i].Node.Speed <= 0 && subNodes[j].Node.Speed <= 0 {
				less = subNodes[i].Node.ID < subNodes[j].Node.ID
			} else if subNodes[i].Node.Speed <= 0 {
				less = false
			} else if subNodes[j].Node.Speed <= 0 {
				less = true
			} else {
				less = subNodes[i].Node.Speed < subNodes[j].Node.Speed
			}
		case "country":
			less = subNodes[i].Node.LinkCountry < subNodes[j].Node.LinkCountry
		default:
			less = subNodes[i].Node.ID < subNodes[j].Node.ID
		}

		// 如果是降序，反转结果
		if sortOrder == "desc" {
			return !less
		}
		return less
	})

	// 批量更新排序值
	tx := database.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开启事务失败: %w", tx.Error)
	}

	for idx, sn := range subNodes {
		err := tx.Model(&SubcriptionNode{}).
			Where("subcription_id = ? AND node_id = ?", sub.ID, sn.NodeID).
			Update("sort", idx).Error
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

// PreviewNode 预览节点信息结构体
type PreviewNode struct {
	Node
	OriginalName string `json:"OriginalName"` // 原始名称（处理前）
	PreviewName  string `json:"PreviewName"`  // 预览名称（应用规则后）
	PreviewLink  string `json:"PreviewLink"`  // 预览链接（应用重命名后）
	Protocol     string `json:"Protocol"`     // 协议类型
	CountryFlag  string `json:"CountryFlag"`  // 国旗 emoji
}

// PreviewResult 预览结果结构体
type PreviewResult struct {
	Nodes         []PreviewNode `json:"Nodes"`
	TotalCount    int           `json:"TotalCount"`    // 原始节点总数
	FilteredCount int           `json:"FilteredCount"` // 过滤后节点数
	// 用量信息
	UsageUpload   int64 `json:"UsageUpload"`   // 已上传流量（字节）
	UsageDownload int64 `json:"UsageDownload"` // 已下载流量（字节）
	UsageTotal    int64 `json:"UsageTotal"`    // 总流量配额（字节）
	UsageExpire   int64 `json:"UsageExpire"`   // 最近到期时间（Unix时间戳）
}

// PreviewSub 预览订阅节点
// 该方法调用共用的 ApplyFilters 方法应用过滤逻辑，同时应用重命名规则生成预览信息
// 注意：调用前需要先设置 sub.Nodes 为待预览的节点列表
func (sub *Subcription) PreviewSub() (*PreviewResult, error) {
	// 记录原始节点数
	totalCount := len(sub.Nodes)

	// 计算用量信息
	upload, download, total, expire := sub.CalculateUsageInfo()

	// 调用共用的过滤方法（与 GetSub 共用逻辑）
	sub.Nodes = sub.ApplyFilters(sub.Nodes)

	// 记录过滤后的节点数
	filteredCount := len(sub.Nodes)

	// 构建预览节点列表
	previewNodes := make([]PreviewNode, 0, filteredCount)

	for idx, node := range sub.Nodes {
		// 应用预处理规则到 LinkName
		processedLinkName := utils.PreprocessNodeName(sub.NodeNamePreprocess, node.LinkName)

		// 计算预览名称
		previewName := node.Name
		previewLink := node.Link

		if sub.NodeNameRule != "" {
			previewName = utils.RenameNode(sub.NodeNameRule, utils.NodeInfo{
				Name:        node.Name,
				LinkName:    processedLinkName,
				LinkCountry: node.LinkCountry,
				Speed:       node.Speed,
				DelayTime:   node.DelayTime,
				Group:       node.Group,
				Source:      node.Source,
				Index:       idx + 1,
				Protocol:    utils.GetProtocolFromLink(node.Link),
				Tags:        node.Tags,
			})
			previewLink = utils.RenameNodeLink(node.Link, previewName)
		}

		previewNodes = append(previewNodes, PreviewNode{
			Node:         node,
			OriginalName: node.LinkName,
			PreviewName:  previewName,
			PreviewLink:  previewLink,
			Protocol:     utils.GetProtocolFromLink(node.Link),
			CountryFlag:  utils.ISOToFlag(node.LinkCountry),
		})
	}

	return &PreviewResult{
		Nodes:         previewNodes,
		TotalCount:    totalCount,
		FilteredCount: filteredCount,
		UsageUpload:   upload,
		UsageDownload: download,
		UsageTotal:    total,
		UsageExpire:   expire,
	}, nil
}

// CalculateUsageInfo 计算订阅的用量信息
func (sub *Subcription) CalculateUsageInfo() (upload, download, total, expire int64) {
	airportIDs := make(map[int]bool)
	for _, node := range sub.Nodes {
		if node.Source != "manual" && node.SourceID > 0 {
			airportIDs[node.SourceID] = true
		}
	}

	now := time.Now().Unix()

	for id := range airportIDs {
		airport, err := GetAirportByID(id)
		if err != nil || airport == nil {
			continue
		}
		if !airport.FetchUsageInfo {
			continue
		}
		// 跳过已过期的机场
		if airport.UsageExpire > 0 && airport.UsageExpire < now {
			continue
		}

		// 累加流量（忽略负数）
		if airport.UsageUpload > 0 {
			upload += airport.UsageUpload
		}
		if airport.UsageDownload > 0 {
			download += airport.UsageDownload
		}
		if airport.UsageTotal > 0 {
			total += airport.UsageTotal
		}

		// 获取最近的过期时间
		if airport.UsageExpire > 0 {
			if expire == 0 || airport.UsageExpire < expire {
				expire = airport.UsageExpire
			}
		}
	}

	return
}

// ApplyNodeFilterScripts 执行节点过滤脚本
// 遍历订阅关联的脚本，依次执行 filterNode 函数处理节点列表
// clientType 参数传递给脚本，默认建议使用 "clash"
func (sub *Subcription) ApplyNodeFilterScripts(nodes []Node, scripts []ScriptWithSort, clientType string) []Node {
	if len(scripts) == 0 || len(nodes) == 0 {
		return nodes
	}

	result := nodes
	nodesJSON, err := json.Marshal(result)
	if err != nil {
		utils.Error("序列化节点失败: %v", err)
		return nodes
	}

	for _, script := range scripts {
		resJSON, err := utils.RunNodeFilterScript(script.Content, nodesJSON, clientType)
		if err != nil {
			// filterNode 函数不存在时跳过，不报错（脚本可能只定义了 subMod）
			if strings.Contains(err.Error(), "filterNode function not found") {
				continue
			}
			utils.Error("节点过滤脚本执行失败: %v", err)
			continue
		}
		var newNodes []Node
		if err := json.Unmarshal(resJSON, &newNodes); err != nil {
			utils.Error("反序列化过滤后节点失败: %v", err)
			continue
		}
		result = newNodes
		nodesJSON = resJSON
	}

	return result
}

// ========== 去重规则配置结构 ==========

// DeduplicationConfig 去重规则配置
type DeduplicationConfig struct {
	Mode          string              `json:"mode"`          // 去重模式: none, common, protocol
	CommonFields  []string            `json:"commonFields"`  // 通用字段列表
	ProtocolRules map[string][]string `json:"protocolRules"` // 协议特定规则
}

// ApplyDeduplication 应用去重规则
func (sub *Subcription) ApplyDeduplication(nodes []Node) []Node {
	// 如果没有配置去重规则，直接返回
	if sub.DeduplicationRule == "" {
		return nodes
	}

	// 解析去重配置
	var config DeduplicationConfig
	if err := json.Unmarshal([]byte(sub.DeduplicationRule), &config); err != nil {
		utils.Warn("解析去重规则失败: %v", err)
		return nodes
	}

	// 根据模式应用去重
	switch config.Mode {
	case "common":
		return deduplicateByCommonFields(nodes, config.CommonFields)
	case "protocol":
		return deduplicateByProtocol(nodes, config.ProtocolRules)
	default:
		return nodes
	}
}

// deduplicateByCommonFields 根据通用字段去重
func deduplicateByCommonFields(nodes []Node, fields []string) []Node {
	if len(fields) == 0 {
		return nodes
	}

	seen := make(map[string]bool)
	var result []Node

	for _, node := range nodes {
		// 生成去重Key
		key := generateNodeKey(&node, fields)
		if key == "" {
			// 如果无法生成Key，保留该节点
			result = append(result, node)
			continue
		}

		if !seen[key] {
			seen[key] = true
			result = append(result, node)
		}
	}

	utils.Info("通用字段去重: 原%d个 -> %d个", len(nodes), len(result))
	return result
}

// generateNodeKey 根据指定字段生成节点的去重Key
func generateNodeKey(node *Node, fields []string) string {
	var parts []string
	for _, field := range fields {
		value := node.GetFieldValue(field)
		parts = append(parts, field+":"+value)
	}
	return strings.Join(parts, "|")
}

// deduplicateByProtocol 根据协议特定字段去重
func deduplicateByProtocol(nodes []Node, protocolRules map[string][]string) []Node {
	if len(protocolRules) == 0 {
		return nodes
	}

	seen := make(map[string]bool)
	var result []Node

	for _, node := range nodes {
		// 获取协议类型
		protoType := protocol.GetProtocolFromLink(node.Link)

		// 获取该协议的去重字段
		fields, exists := protocolRules[protoType]
		if !exists || len(fields) == 0 {
			// 没有配置该协议的去重规则，保留节点
			result = append(result, node)
			continue
		}

		// 生成去重Key
		key := generateProtocolKey(node.Link, protoType, fields)
		if key == "" {
			result = append(result, node)
			continue
		}

		// 加上协议类型前缀，避免不同协议间Key冲突
		fullKey := protoType + ":" + key
		if !seen[fullKey] {
			seen[fullKey] = true
			result = append(result, node)
		}
	}

	utils.Info("协议字段去重: 原%d个 -> %d个", len(nodes), len(result))
	return result
}

// generateProtocolKey 根据协议解析结果生成去重Key
func generateProtocolKey(link string, protoType string, fields []string) string {
	var protoObj interface{}
	var err error

	// 根据协议类型解析节点
	switch protoType {
	case "vmess":
		protoObj, err = protocol.DecodeVMESSURL(link)
	case "vless":
		protoObj, err = protocol.DecodeVLESSURL(link)
	case "trojan":
		protoObj, err = protocol.DecodeTrojanURL(link)
	case "ss":
		protoObj, err = protocol.DecodeSSURL(link)
	case "ssr":
		protoObj, err = protocol.DecodeSSRURL(link)
	case "hysteria":
		protoObj, err = protocol.DecodeHYURL(link)
	case "hysteria2":
		protoObj, err = protocol.DecodeHY2URL(link)
	case "tuic":
		protoObj, err = protocol.DecodeTuicURL(link)
	case "anytls":
		protoObj, err = protocol.DecodeAnyTLSURL(link)
	case "socks5":
		protoObj, err = protocol.DecodeSocks5URL(link)
	default:
		return ""
	}

	if err != nil {
		return ""
	}

	// 根据字段列表提取值并生成Key
	var parts []string
	for _, field := range fields {
		value := protocol.GetProtocolFieldValue(protoObj, field)
		parts = append(parts, field+":"+value)
	}

	return strings.Join(parts, "|")
}
