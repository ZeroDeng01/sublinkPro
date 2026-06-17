package models

import (
	"bufio"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
	"sync"
	"time"

	"gorm.io/gorm"
)

const (
	// MinCountryRulePriority 国家规则最小优先级
	MinCountryRulePriority = 0
	// MaxCountryRulePriority 国家规则最大优先级
	MaxCountryRulePriority = 1000
)

// CountryRule 国家规则模型
// 用于从节点名称解析国家代码
type CountryRule struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	CountryCode string    `gorm:"size:10;not null;uniqueIndex" json:"countryCode"` // 国家代码 (如 CN, US, HK)
	CountryName string    `gorm:"size:100;not null" json:"countryName"`            // 国家名称 (如 中国, 美国, 香港)
	Pattern     string    `gorm:"type:text;not null" json:"pattern"`               // 匹配模式
	Enabled     bool      `gorm:"default:true;index" json:"enabled"`               // 是否启用
	Priority    int       `gorm:"default:0;index" json:"priority"`                 // 优先级 (数字越大越优先，默认0)
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`                 // 创建时间
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`                 // 更新时间
}

// TableName 指定表名
func (CountryRule) TableName() string {
	return "country_rules"
}

// countryRuleCache 国家规则缓存
var countryRuleCache *cache.MapCache[int, CountryRule]
var countryRuleCacheMu sync.RWMutex

// compiledRegexCache 正则表达式编译缓存
var compiledRegexCache = make(map[int]*regexp.Regexp)
var regexCacheMu sync.RWMutex

func init() {
	countryRuleCache = cache.NewMapCache(func(r CountryRule) int { return r.ID })
	countryRuleCache.AddIndex("enabled", func(r CountryRule) string {
		if r.Enabled {
			return "true"
		}
		return "false"
	})
	countryRuleCache.AddIndex("countryCode", func(r CountryRule) string { return r.CountryCode })
}

// InitCountryRuleCache 初始化国家规则缓存
func InitCountryRuleCache() error {
	utils.Info("开始加载国家规则到缓存")
	var rules []CountryRule
	if err := database.DB.Find(&rules).Error; err != nil {
		return err
	}

	countryRuleCache.LoadAll(rules)
	utils.Info("国家规则缓存初始化完成，共加载 %d 条规则", countryRuleCache.Count())

	// 预编译正则表达式
	for _, rule := range rules {
		if rule.Enabled {
			compileAndCacheRegex(rule.ID, rule.Pattern)
		}
	}

	cache.Manager.Register("country_rule", countryRuleCache)
	return nil
}

// compileAndCacheRegex 编译并缓存正则表达式
func compileAndCacheRegex(ruleID int, pattern string) {
	regexCacheMu.Lock()
	defer regexCacheMu.Unlock()

	re, err := regexp.Compile(pattern)
	if err != nil {
		utils.Warn("国家规则 ID=%d 正则表达式编译失败: %v", ruleID, err)
		delete(compiledRegexCache, ruleID)
		return
	}
	compiledRegexCache[ruleID] = re
}

// getCachedRegex 获取缓存的正则表达式
func getCachedRegex(ruleID int) (*regexp.Regexp, bool) {
	regexCacheMu.RLock()
	defer regexCacheMu.RUnlock()
	re, ok := compiledRegexCache[ruleID]
	return re, ok
}

// clearRegexCache 清除正则表达式缓存
func clearRegexCache(ruleID int) {
	regexCacheMu.Lock()
	defer regexCacheMu.Unlock()
	delete(compiledRegexCache, ruleID)
}

// GetEnabledCountryRules 获取所有启用的国家规则（按优先级降序、创建时间升序）
func GetEnabledCountryRules() []CountryRule {
	countryRuleCacheMu.RLock()
	defer countryRuleCacheMu.RUnlock()

	rules := countryRuleCache.GetByIndex("enabled", "true")

	// 排序：优先级降序，优先级相同时按创建时间升序
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Priority != rules[j].Priority {
			return rules[i].Priority > rules[j].Priority
		}
		return rules[i].CreatedAt.Before(rules[j].CreatedAt)
	})

	return rules
}

// Add 添加国家规则 (Write-Through)
func (r *CountryRule) Add() error {
	// 验证数据
	if err := r.validate(); err != nil {
		return err
	}

	// 标准化数据
	r.normalize()

	// 检查是否存在重复的国家代码
	var existing CountryRule
	if err := database.DB.Where("country_code = ?", r.CountryCode).First(&existing).Error; err == nil {
		return fmt.Errorf("国家代码 %s 已存在", r.CountryCode)
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	// 写入数据库 - 使用 map 避免 GORM 跳过零值（特别是 Enabled=false 的情况）
	// GORM v2 会跳过结构体的零值字段，即使使用 Select() 也无效
	if err := database.DB.Model(&CountryRule{}).Create(map[string]any{
		"CountryCode": r.CountryCode,
		"CountryName": r.CountryName,
		"Pattern":     r.Pattern,
		"Priority":    r.Priority,
		"Enabled":     r.Enabled,
	}).Error; err != nil {
		return err
	}

	// 从数据库读取完整记录（包含自动生成的ID和时间戳）
	if err := database.DB.Where("country_code = ?", r.CountryCode).First(r).Error; err != nil {
		return err
	}

	// 更新缓存
	countryRuleCacheMu.Lock()
	countryRuleCache.Set(r.ID, *r)
	countryRuleCacheMu.Unlock()

	// 预编译正则表达式
	if r.Enabled {
		compileAndCacheRegex(r.ID, r.Pattern)
	}

	return nil
}

// Update 更新国家规则 (Write-Through)
func (r *CountryRule) Update() error {
	// 验证数据
	if err := r.validate(); err != nil {
		return err
	}

	// 标准化数据
	r.normalize()

	// 更新数据库
	result := database.DB.Model(r).Select(
		"CountryCode", "CountryName", "Pattern", "Enabled", "Priority",
	).Updates(r)
	if result.Error != nil {
		return result.Error
	}

	// 检查是否影响了行（记录是否存在）
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	// 从数据库读取完整数据后更新缓存
	var updated CountryRule
	if err := database.DB.First(&updated, r.ID).Error; err == nil {
		countryRuleCacheMu.Lock()
		countryRuleCache.Set(r.ID, updated)
		countryRuleCacheMu.Unlock()
	}

	// 更新正则表达式缓存
	if r.Enabled {
		compileAndCacheRegex(r.ID, r.Pattern)
	} else {
		clearRegexCache(r.ID)
	}

	return nil
}

// Delete 删除国家规则 (Write-Through)
func (r *CountryRule) Delete() error {
	// 删除数据库记录
	if err := database.DB.Delete(r).Error; err != nil {
		return err
	}

	// 更新缓存
	countryRuleCacheMu.Lock()
	countryRuleCache.Delete(r.ID)
	countryRuleCacheMu.Unlock()

	// 清除正则表达式缓存
	clearRegexCache(r.ID)

	return nil
}

// List 获取所有国家规则
func (r *CountryRule) List() ([]CountryRule, error) {
	countryRuleCacheMu.RLock()
	defer countryRuleCacheMu.RUnlock()

	rules := countryRuleCache.GetAllSorted(func(a, b CountryRule) bool {
		// 按优先级降序，优先级相同时按ID升序
		if a.Priority != b.Priority {
			return a.Priority > b.Priority
		}
		return a.ID < b.ID
	})

	return rules, nil
}

// GetByID 根据ID获取国家规则
func (r *CountryRule) GetByID() error {
	countryRuleCacheMu.RLock()
	cached, ok := countryRuleCache.Get(r.ID)
	countryRuleCacheMu.RUnlock()

	if ok {
		*r = cached
		return nil
	}

	// 缓存未命中，从数据库查询
	if err := database.DB.First(r, r.ID).Error; err != nil {
		return err
	}

	// 更新缓存
	countryRuleCacheMu.Lock()
	countryRuleCache.Set(r.ID, *r)
	countryRuleCacheMu.Unlock()

	return nil
}

// validate 验证规则数据
func (r *CountryRule) validate() error {
	// 验证国家代码
	if strings.TrimSpace(r.CountryCode) == "" {
		return fmt.Errorf("国家代码不能为空")
	}
	if len(r.CountryCode) < 2 || len(r.CountryCode) > 10 {
		return fmt.Errorf("国家代码长度必须在2-10个字符之间")
	}

	// 验证国家名称
	if strings.TrimSpace(r.CountryName) == "" {
		return fmt.Errorf("国家名称不能为空")
	}

	// 验证匹配模式
	if strings.TrimSpace(r.Pattern) == "" {
		return fmt.Errorf("匹配模式不能为空")
	}

	// 验证正则表达式
	if _, err := regexp.Compile(r.Pattern); err != nil {
		return fmt.Errorf("正则表达式编译失败: %v", err)
	}

	// 验证优先级
	if r.Priority < MinCountryRulePriority || r.Priority > MaxCountryRulePriority {
		return fmt.Errorf("优先级必须在%d-%d之间", MinCountryRulePriority, MaxCountryRulePriority)
	}

	return nil
}

// normalize 标准化规则数据
func (r *CountryRule) normalize() {
	// 国家代码转大写
	r.CountryCode = strings.ToUpper(strings.TrimSpace(r.CountryCode))
	// 国家名称去除首尾空格
	r.CountryName = strings.TrimSpace(r.CountryName)
	// 匹配模式去除首尾空格
	r.Pattern = strings.TrimSpace(r.Pattern)
}

// MatchesNodeName 检查节点名称是否匹配此规则
func (r *CountryRule) MatchesNodeName(nodeName string) bool {
	if !r.Enabled || nodeName == "" {
		return false
	}

	// 统一使用正则表达式匹配
	// 使用缓存的正则表达式
	if re, ok := getCachedRegex(r.ID); ok {
		return re.MatchString(nodeName)
	}
	// 缓存未命中，尝试即时编译
	re, err := regexp.Compile(r.Pattern)
	if err != nil {
		return false
	}
	return re.MatchString(nodeName)
}

// ParseCountryFromNodeName 从节点名称解析国家代码
// 按优先级匹配规则，返回第一个匹配的国家代码
func ParseCountryFromNodeName(nodeName string) string {
	if nodeName == "" {
		return ""
	}

	rules := GetEnabledCountryRules()
	for _, rule := range rules {
		if rule.MatchesNodeName(nodeName) {
			return rule.CountryCode
		}
	}

	return ""
}

// TestPattern 测试匹配模式（静态方法，用于API测试）
func TestPattern(pattern, testName string) (bool, error) {
	if pattern == "" || testName == "" {
		return false, nil
	}

	// 统一使用正则表达式匹配
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Errorf("正则表达式编译失败: %v", err)
	}
	return re.MatchString(testName), nil
}

// ========== 文本导出导入 ==========

// ExportCountryRulesToText 将所有国家规则导出为文本格式
// 格式：国家代码 国家名称 优先级 启用状态 正则表达式（每行一条，空格分隔）
// 支持注释行（以 # 开头）
func ExportCountryRulesToText() string {
	countryRuleCacheMu.RLock()
	defer countryRuleCacheMu.RUnlock()

	rules := countryRuleCache.GetAllSorted(func(a, b CountryRule) bool {
		// 按优先级降序，优先级相同时按ID升序
		if a.Priority != b.Priority {
			return a.Priority > b.Priority
		}
		return a.ID < b.ID
	})

	var lines []string
	lines = append(lines, "# 国家规则配置")
	lines = append(lines, "# 格式：国家代码 国家名称 优先级 启用状态 正则表达式")
	lines = append(lines, "# 示例：CN 中国 100 true (?i)(中国|china|cn)")
	lines = append(lines, "")

	for _, r := range rules {
		enabled := "true"
		if !r.Enabled {
			enabled = "false"
		}
		line := fmt.Sprintf("%s %s %d %s %s",
			r.CountryCode,
			r.CountryName,
			r.Priority,
			enabled,
			r.Pattern,
		)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// SyncCountryRulesFromText 从文本全量同步国家规则数据
// 解析文本，与数据库同步（新增、修改、删除）
// 返回同步结果统计
func SyncCountryRulesFromText(text string) (added, updated, deleted int, err error) {
	// 解析文本中的规则
	newRules, parseErrors := parseCountryRuleText(text)

	// 如果有解析错误，返回第一个错误
	if len(parseErrors) > 0 {
		return 0, 0, 0, fmt.Errorf("解析错误: %s", strings.Join(parseErrors, "; "))
	}

	// 获取当前所有规则（以 countryCode 为键）
	countryRuleCacheMu.RLock()
	allCurrentRules := countryRuleCache.GetAll()
	countryRuleCacheMu.RUnlock()

	currentRules := make(map[string]CountryRule)
	for _, r := range allCurrentRules {
		currentRules[r.CountryCode] = r
	}

	// 记录文本中出现的 countryCode
	textCountryCodes := make(map[string]bool)

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Track changes for cache updates after commit
	type cacheUpdate struct {
		action string // "set" or "delete"
		rule   CountryRule
	}
	var cacheUpdates []cacheUpdate

	// 处理新增和更新
	for _, newRule := range newRules {
		textCountryCodes[newRule.CountryCode] = true

		if existing, exists := currentRules[newRule.CountryCode]; exists {
			// 检查是否需要更新
			if existing.CountryName != newRule.CountryName ||
				existing.Pattern != newRule.Pattern ||
				existing.Priority != newRule.Priority ||
				existing.Enabled != newRule.Enabled {

				existing.CountryName = newRule.CountryName
				existing.Pattern = newRule.Pattern
				existing.Priority = newRule.Priority
				existing.Enabled = newRule.Enabled
				existing.UpdatedAt = time.Now()

				if err := tx.Model(&existing).Select(
					"CountryName", "Pattern", "Priority", "Enabled", "UpdatedAt",
				).Updates(&existing).Error; err != nil {
					tx.Rollback()
					return 0, 0, 0, err
				}

				cacheUpdates = append(cacheUpdates, cacheUpdate{"set", existing})
				updated++
			}
		} else {
			// 新增
			newRule.CreatedAt = time.Now()
			newRule.UpdatedAt = time.Now()

			// 使用 map 来避免 GORM 跳过零值（特别是 Enabled=false 的情况）
			// GORM v2 会跳过结构体的零值字段，即使使用 Select() 也无效
			if err := tx.Model(&CountryRule{}).Create(map[string]any{
				"CountryCode": newRule.CountryCode,
				"CountryName": newRule.CountryName,
				"Pattern":     newRule.Pattern,
				"Priority":    newRule.Priority,
				"Enabled":     newRule.Enabled,
				"CreatedAt":   newRule.CreatedAt,
				"UpdatedAt":   newRule.UpdatedAt,
			}).Error; err != nil {
				tx.Rollback()
				return 0, 0, 0, err
			}

			// 从数据库读取完整记录（包含自动生成的ID）
			var created CountryRule
			if err := tx.Where("country_code = ?", newRule.CountryCode).First(&created).Error; err != nil {
				tx.Rollback()
				return 0, 0, 0, err
			}

			cacheUpdates = append(cacheUpdates, cacheUpdate{"set", created})
			added++
		}
	}

	// 处理删除（数据库中存在但文本中不存在的）
	for countryCode, rule := range currentRules {
		if !textCountryCodes[countryCode] {
			if err := tx.Delete(&rule).Error; err != nil {
				tx.Rollback()
				return 0, 0, 0, err
			}

			cacheUpdates = append(cacheUpdates, cacheUpdate{"delete", rule})
			deleted++
		}
	}

	if err := tx.Commit().Error; err != nil {
		return 0, 0, 0, err
	}

	// Apply cache updates after successful commit
	countryRuleCacheMu.Lock()
	for _, update := range cacheUpdates {
		if update.action == "set" {
			countryRuleCache.Set(update.rule.ID, update.rule)
			// 更新正则表达式缓存
			if update.rule.Enabled {
				compileAndCacheRegex(update.rule.ID, update.rule.Pattern)
			} else {
				clearRegexCache(update.rule.ID)
			}
		} else {
			countryRuleCache.Delete(update.rule.ID)
			clearRegexCache(update.rule.ID)
		}
	}
	countryRuleCacheMu.Unlock()

	return added, updated, deleted, nil
}

// parseCountryRuleText 解析国家规则文本
// 支持格式：国家代码 国家名称 优先级 启用状态 正则表达式
// 字段之间用空格分隔，正则表达式从第5个字段开始到行尾（可包含空格和特殊字符）
// 忽略空行和以 # 开头的注释行
// 返回规则列表和解析错误列表
func parseCountryRuleText(text string) ([]CountryRule, []string) {
	var rules []CountryRule
	var errors []string

	scanner := bufio.NewScanner(strings.NewReader(text))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 分割字段：只分割前4个字段，第5个字段到行尾都是正则表达式
		parts := strings.SplitN(line, " ", 5)
		if len(parts) != 5 {
			errors = append(errors, fmt.Sprintf("第%d行格式错误：应为5个字段（用空格分隔），实际%d个", lineNum, len(parts)))
			continue
		}

		// 解析字段
		countryCode := strings.TrimSpace(parts[0])
		countryName := strings.TrimSpace(parts[1])
		priorityStr := strings.TrimSpace(parts[2])
		enabledStr := strings.TrimSpace(parts[3])
		pattern := strings.TrimSpace(parts[4])

		// 验证国家代码
		if countryCode == "" {
			errors = append(errors, fmt.Sprintf("第%d行：国家代码不能为空", lineNum))
			continue
		}
		if len(countryCode) < 2 || len(countryCode) > 10 {
			errors = append(errors, fmt.Sprintf("第%d行：国家代码长度必须在2-10个字符之间", lineNum))
			continue
		}

		// 验证国家名称
		if countryName == "" {
			errors = append(errors, fmt.Sprintf("第%d行：国家名称不能为空", lineNum))
			continue
		}

		// 解析优先级
		priority, err := strconv.Atoi(priorityStr)
		if err != nil {
			errors = append(errors, fmt.Sprintf("第%d行：优先级必须是整数", lineNum))
			continue
		}
		if priority < MinCountryRulePriority || priority > MaxCountryRulePriority {
			errors = append(errors, fmt.Sprintf("第%d行：优先级必须在%d-%d之间", lineNum, MinCountryRulePriority, MaxCountryRulePriority))
			continue
		}

		// 解析启用状态
		enabled := true
		if enabledStr == "false" || enabledStr == "0" {
			enabled = false
		} else if enabledStr != "true" && enabledStr != "1" {
			errors = append(errors, fmt.Sprintf("第%d行：启用状态必须是true/false或1/0", lineNum))
			continue
		}

		// 验证匹配模式
		if pattern == "" {
			errors = append(errors, fmt.Sprintf("第%d行：匹配模式不能为空", lineNum))
			continue
		}

		// 验证正则表达式
		if _, err := regexp.Compile(pattern); err != nil {
			errors = append(errors, fmt.Sprintf("第%d行：正则表达式编译失败: %v", lineNum, err))
			continue
		}

		// 创建规则
		rule := CountryRule{
			CountryCode: strings.ToUpper(countryCode),
			CountryName: countryName,
			Pattern:     pattern,
			Priority:    priority,
			Enabled:     enabled,
		}

		rules = append(rules, rule)
	}

	return rules, errors
}
