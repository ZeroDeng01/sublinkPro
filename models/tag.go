package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"time"

	"gorm.io/gorm"
)

// Tag 标签模型 - 使用Name作为主键
type Tag struct {
	Name        string    `gorm:"primaryKey;size:100" json:"name"` // 标签名称（主键）
	GroupName   string    `gorm:"size:100;index" json:"groupName"` // 标签组（同组标签互斥）
	Color       string    `gorm:"default:'#1976d2'" json:"color"`  // 标签颜色(HEX)
	Description string    `json:"description"`                     // 标签描述
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TagRule 自动标签规则
type TagRule struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	TagName     string    `gorm:"index;size:100" json:"tagName"` // 关联的标签名称
	Name        string    `json:"name"`                          // 规则名称
	Enabled     bool      `gorm:"default:true" json:"enabled"`   // 是否启用
	TriggerType string    `json:"triggerType"`                   // 触发类型: subscription_update, speed_test
	Conditions  string    `gorm:"type:text" json:"conditions"`   // JSON条件表达式
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TagConditions 条件表达式结构
type TagConditions struct {
	Logic      string         `json:"logic"`      // "and" 或 "or"
	Conditions []TagCondition `json:"conditions"` // 条件列表
}

// TagCondition 单个条件
type TagCondition struct {
	Field    string      `json:"field"`    // 字段名
	Operator string      `json:"operator"` // 操作符
	Value    interface{} `json:"value"`    // 比较值
}

// tagCache 标签缓存 - 使用name作为主键
var tagCache *cache.MapCache[string, Tag]

// tagRuleCache 标签规则缓存
var tagRuleCache *cache.MapCache[int, TagRule]

func init() {
	tagCache = cache.NewMapCache(func(t Tag) string { return t.Name })
	tagRuleCache = cache.NewMapCache(func(r TagRule) int { return r.ID })
}

// InitTagCache 初始化标签缓存
func InitTagCache() error {
	utils.Info("开始加载标签到缓存")

	var tags []Tag
	if err := database.DB.Find(&tags).Error; err != nil {
		return err
	}

	tagCache.LoadAll(tags)
	utils.Info("标签缓存初始化完成，共加载 %d 个标签", tagCache.Count())

	cache.Manager.Register("tag", tagCache)
	return nil
}

// InitTagRuleCache 初始化标签规则缓存
func InitTagRuleCache() error {
	utils.Info("开始加载标签规则到缓存")

	tagRuleCache.AddIndex("tagName", func(r TagRule) string { return r.TagName })
	tagRuleCache.AddIndex("triggerType", func(r TagRule) string { return r.TriggerType })

	var rules []TagRule
	if err := database.DB.Find(&rules).Error; err != nil {
		return err
	}

	tagRuleCache.LoadAll(rules)
	utils.Info("标签规则缓存初始化完成，共加载 %d 个规则", tagRuleCache.Count())

	cache.Manager.Register("tagRule", tagRuleCache)
	return nil
}

// ========== Tag CRUD ==========

// Add 添加标签
func (t *Tag) Add() error {
	if t.Name == "" {
		return fmt.Errorf("标签名称不能为空")
	}
	if err := database.DB.Create(t).Error; err != nil {
		return err
	}
	tagCache.Set(t.Name, *t)
	return nil
}

// Update 更新标签（可更新颜色、描述和标签组，不能修改名称）
func (t *Tag) Update() error {
	t.UpdatedAt = time.Now()
	if err := database.DB.Model(t).Where("name = ?", t.Name).Updates(map[string]interface{}{
		"group_name":  t.GroupName,
		"color":       t.Color,
		"description": t.Description,
		"updated_at":  t.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	tagCache.Set(t.Name, *t)
	return nil
}

// Delete 删除标签
func (t *Tag) Delete() error {
	// 删除关联规则
	if err := database.DB.Where("tag_name = ?", t.Name).Delete(&TagRule{}).Error; err != nil {
		return err
	}
	// 清除节点上的此标签
	ClearTagFromAllNodes(t.Name)
	// 删除标签
	if err := database.DB.Where("name = ?", t.Name).Delete(&Tag{}).Error; err != nil {
		return err
	}
	// 清除规则缓存中相关规则
	rules := tagRuleCache.GetByIndex("tagName", t.Name)
	for _, r := range rules {
		tagRuleCache.Delete(r.ID)
	}
	tagCache.Delete(t.Name)
	return nil
}

// GetByName 根据名称获取标签
func (t *Tag) GetByName(name string) error {
	if cached, ok := tagCache.Get(name); ok {
		*t = cached
		return nil
	}
	if err := database.DB.Where("name = ?", name).First(t).Error; err != nil {
		return err
	}
	tagCache.Set(t.Name, *t)
	return nil
}

// Exists 检查标签是否存在
func TagExists(name string) bool {
	if _, ok := tagCache.Get(name); ok {
		return true
	}
	var count int64
	database.DB.Model(&Tag{}).Where("name = ?", name).Count(&count)
	return count > 0
}

// List 获取所有标签
func (t *Tag) List() ([]Tag, error) {
	if tagCache.Count() > 0 {
		return tagCache.GetAll(), nil
	}
	var tags []Tag
	if err := database.DB.Find(&tags).Error; err != nil {
		return nil, err
	}
	for _, tag := range tags {
		tagCache.Set(tag.Name, tag)
	}
	return tags, nil
}

// GetTagsInSameGroup 获取与指定标签同组的其他标签名称列表
// 用于实现标签互斥逻辑
func GetTagsInSameGroup(tagName string) []string {
	tag, ok := tagCache.Get(tagName)
	if !ok || tag.GroupName == "" {
		return nil
	}

	// 从缓存获取同组所有标签
	allTags := tagCache.GetAll()
	sameGroupTags := make([]string, 0)
	for _, t := range allTags {
		if t.GroupName == tag.GroupName && t.Name != tagName {
			sameGroupTags = append(sameGroupTags, t.Name)
		}
	}
	return sameGroupTags
}

// GetExistingGroups 获取所有已存在的标签组名称（用于前端自动补全）
func GetExistingGroups() []string {
	allTags := tagCache.GetAll()
	groupSet := make(map[string]bool)
	for _, t := range allTags {
		if t.GroupName != "" {
			groupSet[t.GroupName] = true
		}
	}

	groups := make([]string, 0, len(groupSet))
	for g := range groupSet {
		groups = append(groups, g)
	}
	return groups
}

// ========== TagRule CRUD ==========

// Add 添加规则
func (r *TagRule) Add() error {
	if err := database.DB.Create(r).Error; err != nil {
		return err
	}
	tagRuleCache.Set(r.ID, *r)
	return nil
}

// Update 更新规则
func (r *TagRule) Update() error {
	r.UpdatedAt = time.Now()
	if err := database.DB.Save(r).Error; err != nil {
		return err
	}
	tagRuleCache.Set(r.ID, *r)
	return nil
}

// Delete 删除规则
func (r *TagRule) Delete() error {
	if err := database.DB.Delete(r).Error; err != nil {
		return err
	}
	tagRuleCache.Delete(r.ID)
	return nil
}

// GetByID 根据ID获取规则
func (r *TagRule) GetByID(id int) error {
	if cached, ok := tagRuleCache.Get(id); ok {
		*r = cached
		return nil
	}
	if err := database.DB.First(r, id).Error; err != nil {
		return err
	}
	tagRuleCache.Set(r.ID, *r)
	return nil
}

// List 获取所有规则
func (r *TagRule) List() ([]TagRule, error) {
	if tagRuleCache.Count() > 0 {
		return tagRuleCache.GetAll(), nil
	}
	var rules []TagRule
	if err := database.DB.Find(&rules).Error; err != nil {
		return nil, err
	}
	for _, rule := range rules {
		tagRuleCache.Set(rule.ID, rule)
	}
	return rules, nil
}

// ListByTriggerType 根据触发类型获取启用的规则
func ListByTriggerType(triggerType string) []TagRule {
	rules := tagRuleCache.GetByIndex("triggerType", triggerType)
	enabledRules := make([]TagRule, 0)
	for _, r := range rules {
		if r.Enabled {
			enabledRules = append(enabledRules, r)
		}
	}
	return enabledRules
}

// ========== 条件评估 ==========

// ParseConditions 解析JSON条件表达式
func ParseConditions(conditionsJSON string) (*TagConditions, error) {
	if conditionsJSON == "" {
		return nil, fmt.Errorf("empty conditions")
	}
	var conditions TagConditions
	if err := json.Unmarshal([]byte(conditionsJSON), &conditions); err != nil {
		return nil, err
	}
	return &conditions, nil
}

// EvaluateNode 对节点评估条件
func (tc *TagConditions) EvaluateNode(node Node) bool {
	if len(tc.Conditions) == 0 {
		return false
	}

	results := make([]bool, len(tc.Conditions))
	for i, cond := range tc.Conditions {
		results[i] = evaluateCondition(node, cond)
	}

	// 根据逻辑运算符合并结果
	if tc.Logic == "or" {
		for _, r := range results {
			if r {
				return true
			}
		}
		return false
	}

	// 默认 AND 逻辑
	for _, r := range results {
		if !r {
			return false
		}
	}
	return true
}

// evaluateCondition 评估单个条件
func evaluateCondition(node Node, cond TagCondition) bool {
	fieldValue := getNodeFieldValue(node, cond.Field)
	compareValue := cond.Value

	switch cond.Operator {
	case "equals":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", compareValue)
	case "not_equals":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", compareValue)
	case "contains":
		return strings.Contains(strings.ToLower(fmt.Sprintf("%v", fieldValue)), strings.ToLower(fmt.Sprintf("%v", compareValue)))
	case "not_contains":
		return !strings.Contains(strings.ToLower(fmt.Sprintf("%v", fieldValue)), strings.ToLower(fmt.Sprintf("%v", compareValue)))
	case "regex":
		pattern := fmt.Sprintf("%v", compareValue)
		re, err := regexp.Compile(pattern)
		if err != nil {
			utils.Error("正则表达式编译失败: %s, error: %v", pattern, err)
			return false
		}
		return re.MatchString(fmt.Sprintf("%v", fieldValue))
	case "greater_than":
		return compareNumeric(fieldValue, compareValue) > 0
	case "less_than":
		return compareNumeric(fieldValue, compareValue) < 0
	case "greater_or_equal":
		return compareNumeric(fieldValue, compareValue) >= 0
	case "less_or_equal":
		return compareNumeric(fieldValue, compareValue) <= 0
	default:
		return false
	}
}

// getNodeFieldValue 获取节点字段值
func getNodeFieldValue(node Node, field string) interface{} {
	switch field {
	case "name":
		return node.Name
	case "link_name":
		return node.LinkName
	case "link_address":
		return node.LinkAddress
	case "link_host":
		return node.LinkHost
	case "link_port":
		return node.LinkPort
	case "link_country":
		return node.LinkCountry
	case "protocol":
		return node.Protocol
	case "source":
		return node.Source
	case "group":
		return node.Group
	case "speed":
		return node.Speed
	case "speed_status":
		return node.SpeedStatus
	case "delay_time":
		return node.DelayTime
	case "delay_status":
		return node.DelayStatus
	case "dialer_proxy_name":
		return node.DialerProxyName
	case "link":
		return node.Link
	case "tags":
		return node.Tags
	default:
		return ""
	}
}

// compareNumeric 数值比较，返回 -1, 0, 1
func compareNumeric(a, b interface{}) int {
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)
	if aFloat > bFloat {
		return 1
	} else if aFloat < bFloat {
		return -1
	}
	return 0
}

// toFloat64 转换为float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

// ========== 节点标签操作 (使用标签名称) ==========

// GetTagNames 获取节点的标签名称列表
func (n *Node) GetTagNames() []string {
	if n.Tags == "" {
		return []string{}
	}
	parts := strings.Split(n.Tags, ",")
	names := make([]string, 0, len(parts))
	seen := make(map[string]bool)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" && !seen[p] {
			names = append(names, p)
			seen[p] = true
		}
	}
	return names
}

// HasTagName 检查节点是否有指定标签
func (n *Node) HasTagName(tagName string) bool {
	for _, name := range n.GetTagNames() {
		if name == tagName {
			return true
		}
	}
	return false
}

// AddTagByName 添加标签到节点（按名称）
// 如果标签属于某个组，会自动移除该节点上同组的其他标签（互斥逻辑）
func (n *Node) AddTagByName(tagName string) error {
	if n.HasTagName(tagName) {
		return nil // 已有此标签
	}

	// 获取同组的互斥标签
	sameGroupTags := GetTagsInSameGroup(tagName)
	tagNames := n.GetTagNames()

	// 过滤掉同组的互斥标签
	newTagNames := make([]string, 0, len(tagNames))
	for _, name := range tagNames {
		isSameGroup := false
		for _, sgt := range sameGroupTags {
			if name == sgt {
				isSameGroup = true
				break
			}
		}
		if !isSameGroup {
			newTagNames = append(newTagNames, name)
		}
	}

	// 添加新标签
	newTagNames = append(newTagNames, tagName)
	return n.SetTagNames(newTagNames)
}

// RemoveTagByName 从节点移除标签（按名称）
func (n *Node) RemoveTagByName(tagName string) error {
	tagNames := n.GetTagNames()
	newTagNames := make([]string, 0, len(tagNames))
	for _, name := range tagNames {
		if name != tagName {
			newTagNames = append(newTagNames, name)
		}
	}
	return n.SetTagNames(newTagNames)
}

// SetTagNames 设置节点标签（去重）
func (n *Node) SetTagNames(tagNames []string) error {
	// 去重
	seen := make(map[string]bool)
	uniqueNames := make([]string, 0, len(tagNames))
	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name != "" && !seen[name] {
			uniqueNames = append(uniqueNames, name)
			seen[name] = true
		}
	}
	n.Tags = strings.Join(uniqueNames, ",")

	// 更新数据库
	if err := database.DB.Model(n).Update("tags", n.Tags).Error; err != nil {
		return err
	}
	// 更新缓存
	if cachedNode, ok := nodeCache.Get(n.ID); ok {
		cachedNode.Tags = n.Tags
		nodeCache.Set(n.ID, cachedNode)
	}
	return nil
}

// ClearTagFromAllNodes 从所有节点清除指定标签（按名称）- 优化版本使用事务批量更新
func ClearTagFromAllNodes(tagName string) {
	allNodes := nodeCache.GetAll()
	if len(allNodes) == 0 {
		return
	}

	// 按结果标签分组节点ID
	resultGroups := make(map[string][]int)

	for _, node := range allNodes {
		if !node.HasTagName(tagName) {
			continue
		}
		// 计算移除标签后的新标签列表
		currentTags := node.GetTagNames()
		newTags := make([]string, 0, len(currentTags))
		for _, t := range currentTags {
			if t != tagName {
				newTags = append(newTags, t)
			}
		}
		newTagsString := strings.Join(newTags, ",")
		resultGroups[newTagsString] = append(resultGroups[newTagsString], node.ID)
	}

	if len(resultGroups) == 0 {
		return
	}

	// 使用事务批量更新
	_ = database.WithTransaction(func(tx *gorm.DB) error {
		for newTagsString, ids := range resultGroups {
			if err := tx.Model(&Node{}).Where("id IN ?", ids).Update("tags", newTagsString).Error; err != nil {
				utils.Error("清除标签 %s 失败: %v", tagName, err)
				return err
			}
			// 更新缓存
			for _, id := range ids {
				if cachedNode, ok := nodeCache.Get(id); ok {
					cachedNode.Tags = newTagsString
					nodeCache.Set(id, cachedNode)
				}
			}
		}
		return nil
	})
}

// BatchAddTagToNodes 批量给节点添加标签（按名称）- 优化版本使用事务批量更新
func BatchAddTagToNodes(nodeIDs []int, tagName string) error {
	if len(nodeIDs) == 0 {
		return nil
	}

	// 获取同组的互斥标签
	sameGroupTags := GetTagsInSameGroup(tagName)
	sameGroupSet := make(map[string]bool)
	for _, t := range sameGroupTags {
		sameGroupSet[t] = true
	}

	// 按结果标签分组节点ID
	resultGroups := make(map[string][]int)

	for _, nodeID := range nodeIDs {
		node, ok := nodeCache.Get(nodeID)
		if !ok {
			var dbNode Node
			dbNode.ID = nodeID
			if err := dbNode.GetByID(); err != nil {
				continue
			}
			node = dbNode
		}

		// 已有此标签则跳过
		if node.HasTagName(tagName) {
			continue
		}

		// 计算新标签列表（移除同组互斥标签，添加新标签）
		currentTags := node.GetTagNames()
		newTags := make([]string, 0, len(currentTags)+1)
		for _, t := range currentTags {
			if !sameGroupSet[t] {
				newTags = append(newTags, t)
			}
		}
		newTags = append(newTags, tagName)
		newTagsString := strings.Join(newTags, ",")

		resultGroups[newTagsString] = append(resultGroups[newTagsString], nodeID)
	}

	if len(resultGroups) == 0 {
		return nil
	}

	// 使用事务批量更新
	return database.WithTransaction(func(tx *gorm.DB) error {
		for newTagsString, ids := range resultGroups {
			if err := tx.Model(&Node{}).Where("id IN ?", ids).Update("tags", newTagsString).Error; err != nil {
				utils.Error("批量添加标签 %s 失败: %v", tagName, err)
				return err
			}
			// 更新缓存
			for _, id := range ids {
				if cachedNode, ok := nodeCache.Get(id); ok {
					cachedNode.Tags = newTagsString
					nodeCache.Set(id, cachedNode)
				}
			}
		}
		return nil
	})
}

// BatchRemoveTagFromNodes 批量从节点移除单个标签（按名称）- 优化版本使用事务批量更新
func BatchRemoveTagFromNodes(nodeIDs []int, tagName string) error {
	if len(nodeIDs) == 0 {
		return nil
	}

	// 按结果标签分组节点ID
	resultGroups := make(map[string][]int)

	for _, nodeID := range nodeIDs {
		node, ok := nodeCache.Get(nodeID)
		if !ok {
			var dbNode Node
			dbNode.ID = nodeID
			if err := dbNode.GetByID(); err != nil {
				continue
			}
			node = dbNode
		}

		// 没有此标签则跳过
		if !node.HasTagName(tagName) {
			continue
		}

		// 计算移除标签后的新标签列表
		currentTags := node.GetTagNames()
		newTags := make([]string, 0, len(currentTags))
		for _, t := range currentTags {
			if t != tagName {
				newTags = append(newTags, t)
			}
		}
		newTagsString := strings.Join(newTags, ",")

		resultGroups[newTagsString] = append(resultGroups[newTagsString], nodeID)
	}

	if len(resultGroups) == 0 {
		return nil
	}

	// 使用事务批量更新
	return database.WithTransaction(func(tx *gorm.DB) error {
		for newTagsString, ids := range resultGroups {
			if err := tx.Model(&Node{}).Where("id IN ?", ids).Update("tags", newTagsString).Error; err != nil {
				utils.Error("批量移除标签 %s 失败: %v", tagName, err)
				return err
			}
			// 更新缓存
			for _, id := range ids {
				if cachedNode, ok := nodeCache.Get(id); ok {
					cachedNode.Tags = newTagsString
					nodeCache.Set(id, cachedNode)
				}
			}
		}
		return nil
	})
}

// BatchSetTagsForNodes 批量设置节点标签（覆盖模式）- 优化版本使用事务
func BatchSetTagsForNodes(nodeIDs []int, tagNames []string) error {
	if len(nodeIDs) == 0 {
		return nil
	}

	// 去重并生成标签字符串
	seen := make(map[string]bool)
	uniqueNames := make([]string, 0, len(tagNames))
	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name != "" && !seen[name] {
			uniqueNames = append(uniqueNames, name)
			seen[name] = true
		}
	}
	tagsString := strings.Join(uniqueNames, ",")

	// 使用事务批量更新
	return database.WithTransaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Node{}).Where("id IN ?", nodeIDs).Update("tags", tagsString).Error; err != nil {
			return err
		}

		// 批量更新缓存
		for _, nodeID := range nodeIDs {
			if cachedNode, ok := nodeCache.Get(nodeID); ok {
				cachedNode.Tags = tagsString
				nodeCache.Set(nodeID, cachedNode)
			}
		}
		return nil
	})
}

// BatchRemoveTagsFromNodes 批量从节点移除指定标签 - 优化版本使用事务按结果分组批量更新
func BatchRemoveTagsFromNodes(nodeIDs []int, tagNames []string) error {
	if len(nodeIDs) == 0 || len(tagNames) == 0 {
		return nil
	}

	// 创建要删除的标签集合
	removeSet := make(map[string]bool)
	for _, t := range tagNames {
		removeSet[t] = true
	}

	// 按结果标签分组节点ID（相同结果的节点可以用一条SQL更新）
	resultGroups := make(map[string][]int) // key: 新标签字符串, value: 节点ID列表

	for _, nodeID := range nodeIDs {
		node, ok := nodeCache.Get(nodeID)
		if !ok {
			// 从数据库获取
			var dbNode Node
			dbNode.ID = nodeID
			if err := dbNode.GetByID(); err != nil {
				continue
			}
			node = dbNode
		}

		// 计算移除标签后的新标签列表
		currentTags := node.GetTagNames()
		newTags := make([]string, 0, len(currentTags))
		for _, t := range currentTags {
			if !removeSet[t] {
				newTags = append(newTags, t)
			}
		}
		newTagsString := strings.Join(newTags, ",")

		// 按结果分组
		resultGroups[newTagsString] = append(resultGroups[newTagsString], nodeID)
	}

	if len(resultGroups) == 0 {
		return nil
	}

	// 使用事务对每个分组执行一次批量更新
	return database.WithTransaction(func(tx *gorm.DB) error {
		for newTagsString, ids := range resultGroups {
			if err := tx.Model(&Node{}).Where("id IN ?", ids).Update("tags", newTagsString).Error; err != nil {
				utils.Error("批量移除标签失败: %v", err)
				return err
			}
			// 更新缓存
			for _, id := range ids {
				if cachedNode, ok := nodeCache.Get(id); ok {
					cachedNode.Tags = newTagsString
					nodeCache.Set(id, cachedNode)
				}
			}
		}
		return nil
	})
}

// GetTagsByNode 获取节点的所有标签对象
func GetTagsByNode(node Node) []Tag {
	tagNames := node.GetTagNames()
	tags := make([]Tag, 0, len(tagNames))
	for _, name := range tagNames {
		if tag, ok := tagCache.Get(name); ok {
			tags = append(tags, tag)
		}
	}
	return tags
}
