package models

import (
	"encoding/json"
	"fmt"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"time"
)

// SubscriptionChainRule 订阅链式代理规则
type SubscriptionChainRule struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SubscriptionID int       `gorm:"index;not null" json:"subscriptionId"` // 所属订阅ID
	Name           string    `gorm:"size:100" json:"name"`                 // 规则名称
	Sort           int       `gorm:"default:0" json:"sort"`                // 排序顺序
	Enabled        bool      `gorm:"default:true" json:"enabled"`          // 是否启用
	ChainConfig    string    `gorm:"type:text" json:"chainConfig"`         // 代理链配置 JSON
	TargetConfig   string    `gorm:"type:text" json:"targetConfig"`        // 目标节点条件 JSON
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// ChainProxyItem 代理链单项配置
type ChainProxyItem struct {
	Type           string         `json:"type"`                     // template_group, custom_group, dynamic_node, specified_node
	GroupName      string         `json:"groupName,omitempty"`      // 代理组名称
	GroupType      string         `json:"groupType,omitempty"`      // select, url-test (仅 custom_group)
	URLTestConfig  *URLTestConfig `json:"urlTestConfig,omitempty"`  // URL 测速配置 (仅 custom_group + url-test)
	NodeConditions *TagConditions `json:"nodeConditions,omitempty"` // 节点匹配条件
	SelectMode     string         `json:"selectMode,omitempty"`     // first, random, fastest (仅 dynamic_node)
	NodeID         int            `json:"nodeId,omitempty"`         // 指定节点ID (仅 specified_node)
}

// URLTestConfig URL 测速配置
type URLTestConfig struct {
	URL       string `json:"url"`
	Interval  int    `json:"interval"`
	Tolerance int    `json:"tolerance"`
}

// TargetConfig 目标节点条件配置
type TargetConfig struct {
	Type       string         `json:"type"`                 // all, conditions, specified_node
	Conditions *TagConditions `json:"conditions,omitempty"` // 条件表达式
	NodeID     int            `json:"nodeId,omitempty"`     // 指定节点ID (仅 specified_node)
}

// CustomProxyGroup 自定义代理组（用于生成到 Clash 配置）
type CustomProxyGroup struct {
	Name          string         `json:"name"`
	Type          string         `json:"type"` // select, url-test
	Proxies       []string       `json:"proxies"`
	URLTestConfig *URLTestConfig `json:"urlTestConfig,omitempty"`
}

// chainRuleCache 链式代理规则缓存
var chainRuleCache *cache.MapCache[int, SubscriptionChainRule]

func init() {
	chainRuleCache = cache.NewMapCache(func(r SubscriptionChainRule) int { return r.ID })
}

// InitChainRuleCache 初始化链式代理规则缓存
func InitChainRuleCache() error {
	utils.Info("开始加载链式代理规则到缓存")

	// 添加 SubscriptionID 二级索引
	chainRuleCache.AddIndex("subscriptionId", func(r SubscriptionChainRule) string {
		return fmt.Sprintf("%d", r.SubscriptionID)
	})

	var rules []SubscriptionChainRule
	if err := database.DB.Find(&rules).Error; err != nil {
		return err
	}

	chainRuleCache.LoadAll(rules)
	utils.Info("链式代理规则缓存初始化完成，共加载 %d 条规则", chainRuleCache.Count())

	cache.Manager.Register("chainRule", chainRuleCache)
	return nil
}

// ========== CRUD 操作 ==========

// Add 添加规则 (Write-Through)
func (r *SubscriptionChainRule) Add() error {
	if err := database.DB.Create(r).Error; err != nil {
		return err
	}
	chainRuleCache.Set(r.ID, *r)
	return nil
}

// Update 更新规则 (Write-Through)
func (r *SubscriptionChainRule) Update() error {
	r.UpdatedAt = time.Now()
	if err := database.DB.Save(r).Error; err != nil {
		return err
	}
	chainRuleCache.Set(r.ID, *r)
	return nil
}

// Delete 删除规则 (Write-Through)
func (r *SubscriptionChainRule) Delete() error {
	if err := database.DB.Delete(r).Error; err != nil {
		return err
	}
	chainRuleCache.Delete(r.ID)
	return nil
}

// GetByID 根据 ID 获取规则
func (r *SubscriptionChainRule) GetByID(id int) error {
	if cached, ok := chainRuleCache.Get(id); ok {
		*r = cached
		return nil
	}
	if err := database.DB.First(r, id).Error; err != nil {
		return err
	}
	chainRuleCache.Set(r.ID, *r)
	return nil
}

// GetBySubscriptionID 获取订阅的所有规则（按 Sort 排序）
func GetChainRulesBySubscriptionID(subscriptionID int) []SubscriptionChainRule {
	rules := chainRuleCache.GetByIndex("subscriptionId", fmt.Sprintf("%d", subscriptionID))
	// 按 Sort 排序
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Sort > rules[j].Sort {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}
	return rules
}

// GetEnabledChainRulesBySubscriptionID 获取订阅的已启用规则（按 Sort 排序）
func GetEnabledChainRulesBySubscriptionID(subscriptionID int) []SubscriptionChainRule {
	allRules := GetChainRulesBySubscriptionID(subscriptionID)
	enabled := make([]SubscriptionChainRule, 0)
	for _, r := range allRules {
		if r.Enabled {
			enabled = append(enabled, r)
		}
	}
	return enabled
}

// UpdateChainRulesSort 批量更新规则排序
func UpdateChainRulesSort(ruleIDs []int) error {
	for i, id := range ruleIDs {
		if err := database.DB.Model(&SubscriptionChainRule{}).Where("id = ?", id).Update("sort", i).Error; err != nil {
			return err
		}
		// 更新缓存
		if cached, ok := chainRuleCache.Get(id); ok {
			cached.Sort = i
			chainRuleCache.Set(id, cached)
		}
	}
	return nil
}

// DeleteChainRulesBySubscriptionID 删除订阅的所有链式代理规则
func DeleteChainRulesBySubscriptionID(subscriptionID int) error {
	rules := GetChainRulesBySubscriptionID(subscriptionID)
	for _, r := range rules {
		chainRuleCache.Delete(r.ID)
	}
	return database.DB.Where("subscription_id = ?", subscriptionID).Delete(&SubscriptionChainRule{}).Error
}

// ========== 规则解析和应用 ==========

// ParseChainConfig 解析代理链配置
func (r *SubscriptionChainRule) ParseChainConfig() ([]ChainProxyItem, error) {
	if r.ChainConfig == "" {
		return nil, nil
	}
	var items []ChainProxyItem
	if err := json.Unmarshal([]byte(r.ChainConfig), &items); err != nil {
		return nil, err
	}
	return items, nil
}

// ParseTargetConfig 解析目标节点条件
func (r *SubscriptionChainRule) ParseTargetConfig() (*TargetConfig, error) {
	if r.TargetConfig == "" {
		return &TargetConfig{Type: "all"}, nil
	}
	var target TargetConfig
	if err := json.Unmarshal([]byte(r.TargetConfig), &target); err != nil {
		return nil, err
	}
	return &target, nil
}

// MatchTargetCondition 判断节点是否匹配目标条件
func (r *SubscriptionChainRule) MatchTargetCondition(node Node) bool {
	target, err := r.ParseTargetConfig()
	if err != nil {
		utils.Error("解析目标条件失败: %v", err)
		return false
	}

	switch target.Type {
	case "all":
		return true
	case "specified_node":
		return node.ID == target.NodeID
	case "conditions":
		if target.Conditions == nil {
			return false
		}
		return target.Conditions.EvaluateNode(node)
	default:
		return false
	}
}

// ResolveProxyName 解析规则应生成的代理名称（入口代理名称）
// 返回值：(代理名称, 自定义代理组列表, 错误)
func (r *SubscriptionChainRule) ResolveProxyName(allNodes []Node, nodeNameMap map[int]string) (string, []CustomProxyGroup, error) {
	items, err := r.ParseChainConfig()
	if err != nil {
		return "", nil, err
	}

	if len(items) == 0 {
		return "", nil, nil
	}

	var customGroups []CustomProxyGroup

	// 获取入口代理（第一个配置项）
	entryItem := items[0]
	var entryProxyName string

	switch entryItem.Type {
	case "template_group":
		// 模板代理组：直接使用组名
		if entryItem.GroupName == "" {
			return "", nil, fmt.Errorf("模板代理组名称不能为空")
		}
		entryProxyName = entryItem.GroupName

	case "custom_group":
		// 自定义代理组：需要生成代理组配置
		entryProxyName = entryItem.GroupName
		proxies := r.getMatchingNodeNames(allNodes, entryItem.NodeConditions, nodeNameMap)
		if len(proxies) == 0 {
			return "", nil, fmt.Errorf("自定义代理组 %s 没有匹配的节点", entryItem.GroupName)
		}
		group := CustomProxyGroup{
			Name:          entryItem.GroupName,
			Type:          entryItem.GroupType,
			Proxies:       proxies,
			URLTestConfig: entryItem.URLTestConfig,
		}
		if group.Type == "" {
			group.Type = "select"
		}
		customGroups = append(customGroups, group)

	case "dynamic_node":
		// 动态条件节点：获取第一个匹配的节点名称
		proxyName := r.getFirstMatchingNodeName(allNodes, entryItem.NodeConditions, entryItem.SelectMode, nodeNameMap)
		if proxyName == "" {
			return "", nil, fmt.Errorf("动态条件没有匹配的节点")
		}
		entryProxyName = proxyName

	case "specified_node":
		// 指定节点：获取节点的最终名称
		if name, ok := nodeNameMap[entryItem.NodeID]; ok {
			entryProxyName = name
		} else {
			return "", nil, fmt.Errorf("指定的节点 ID %d 不存在", entryItem.NodeID)
		}

	default:
		return "", nil, fmt.Errorf("未知的代理类型: %s", entryItem.Type)
	}

	// 处理中间链路的 dialer-proxy 设置（如果有多级链）
	// 注意：多级链的处理需要在 GetClash 中统一处理，这里只返回入口代理名称
	// 后续版本可以扩展支持多级链

	return entryProxyName, customGroups, nil
}

// getMatchingNodeNames 获取所有匹配条件的节点名称列表
func (r *SubscriptionChainRule) getMatchingNodeNames(nodes []Node, conditions *TagConditions, nameMap map[int]string) []string {
	if conditions == nil {
		return nil
	}

	var names []string
	for _, node := range nodes {
		if conditions.EvaluateNode(node) {
			if name, ok := nameMap[node.ID]; ok {
				names = append(names, name)
			}
		}
	}
	return names
}

// getFirstMatchingNodeName 获取第一个匹配条件的节点名称
func (r *SubscriptionChainRule) getFirstMatchingNodeName(nodes []Node, conditions *TagConditions, selectMode string, nameMap map[int]string) string {
	if conditions == nil {
		return ""
	}

	var matchedNodes []Node
	for _, node := range nodes {
		if conditions.EvaluateNode(node) {
			matchedNodes = append(matchedNodes, node)
		}
	}

	if len(matchedNodes) == 0 {
		return ""
	}

	// 根据选择模式选择节点
	var selectedNode Node
	switch selectMode {
	case "random":
		// 简单随机：使用当前时间作为种子
		selectedNode = matchedNodes[time.Now().UnixNano()%int64(len(matchedNodes))]
	case "fastest":
		// 最快节点：选择延迟最低的
		selectedNode = matchedNodes[0]
		for _, n := range matchedNodes[1:] {
			if n.DelayTime > 0 && (selectedNode.DelayTime <= 0 || n.DelayTime < selectedNode.DelayTime) {
				selectedNode = n
			}
		}
	default: // "first"
		selectedNode = matchedNodes[0]
	}

	if name, ok := nameMap[selectedNode.ID]; ok {
		return name
	}
	return ""
}

// ========== 辅助函数 ==========

// ApplyChainRulesToNode 为单个节点应用链式代理规则
// 返回值：(dialer-proxy 名称, 自定义代理组列表)
func ApplyChainRulesToNode(node Node, rules []SubscriptionChainRule, allNodes []Node, nodeNameMap map[int]string) (string, []CustomProxyGroup) {
	// 如果节点已设置 DialerProxyName，保留原值
	if node.DialerProxyName != "" {
		return node.DialerProxyName, nil
	}

	// 按顺序匹配规则
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		if rule.MatchTargetCondition(node) {
			proxyName, customGroups, err := rule.ResolveProxyName(allNodes, nodeNameMap)
			if err != nil {
				utils.Warn("规则 %s 解析失败: %v", rule.Name, err)
				continue
			}
			if proxyName != "" {
				return proxyName, customGroups
			}
		}
	}

	return "", nil
}

// CollectCustomProxyGroups 收集所有规则中的自定义代理组
func CollectCustomProxyGroups(rules []SubscriptionChainRule, allNodes []Node, nodeNameMap map[int]string) []CustomProxyGroup {
	groupMap := make(map[string]CustomProxyGroup)

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		items, err := rule.ParseChainConfig()
		if err != nil {
			continue
		}

		for _, item := range items {
			if item.Type == "custom_group" {
				if _, exists := groupMap[item.GroupName]; exists {
					continue // 已存在同名组，跳过
				}

				// 获取匹配的节点名称
				var proxies []string
				if item.NodeConditions != nil {
					for _, node := range allNodes {
						if item.NodeConditions.EvaluateNode(node) {
							if name, ok := nodeNameMap[node.ID]; ok {
								proxies = append(proxies, name)
							}
						}
					}
				}

				if len(proxies) > 0 {
					group := CustomProxyGroup{
						Name:          item.GroupName,
						Type:          item.GroupType,
						Proxies:       proxies,
						URLTestConfig: item.URLTestConfig,
					}
					if group.Type == "" {
						group.Type = "select"
					}
					groupMap[item.GroupName] = group
				}
			}
		}
	}

	// 转换为切片
	groups := make([]CustomProxyGroup, 0, len(groupMap))
	for _, g := range groupMap {
		groups = append(groups, g)
	}
	return groups
}
