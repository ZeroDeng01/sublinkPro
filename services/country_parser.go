package services

import (
	"sublink/models"
	"sublink/utils"
	"sync"
)

// CountryParserService 国家解析服务
// 提供节点名称到国家代码的解析功能
type CountryParserService struct {
	rulesCache       []models.CountryRule
	rulesCacheMu     sync.RWMutex
	cacheInitOnce    sync.Once
	cacheInitialized bool
}

var (
	countryParserService     *CountryParserService
	countryParserServiceOnce sync.Once
)

// GetCountryParserService 获取国家解析服务实例（单例模式）
func GetCountryParserService() *CountryParserService {
	countryParserServiceOnce.Do(func() {
		countryParserService = &CountryParserService{
			rulesCache:       make([]models.CountryRule, 0),
			cacheInitialized: false,
		}
	})
	return countryParserService
}

// ParseCountryFromName 从节点名称解析国家代码
// 参数:
//   - nodeName: 节点名称
//
// 返回:
//   - countryCode: 国家代码（如 CN, US, HK），未匹配时返回空字符串
//   - err: 错误信息
func ParseCountryFromName(nodeName string) (string, error) {
	service := GetCountryParserService()
	return service.parseCountry(nodeName), nil
}

// parseCountry 内部解析方法
func (s *CountryParserService) parseCountry(nodeName string) string {
	if nodeName == "" {
		return ""
	}

	// 确保规则缓存已初始化
	s.ensureRulesLoaded()

	// 读取缓存的规则列表
	s.rulesCacheMu.RLock()
	rules := s.rulesCache
	s.rulesCacheMu.RUnlock()

	// 按优先级顺序匹配规则
	for _, rule := range rules {
		if rule.MatchesNodeName(nodeName) {
			utils.Debug("节点【%s】匹配规则【%s】，国家代码: %s", nodeName, rule.CountryName, rule.CountryCode)
			return rule.CountryCode
		}
	}

	return ""
}

// ensureRulesLoaded 确保规则缓存已加载（使用 sync.Once 保证只初始化一次）
func (s *CountryParserService) ensureRulesLoaded() {
	s.cacheInitOnce.Do(func() {
		s.loadRules()
	})
}

// loadRules 从数据库加载启用的国家规则到缓存
func (s *CountryParserService) loadRules() {
	rules := models.GetEnabledCountryRules()

	s.rulesCacheMu.Lock()
	s.rulesCache = rules
	s.cacheInitialized = true
	s.rulesCacheMu.Unlock()

	utils.Info("国家解析服务: 成功加载 %d 条启用的规则", len(rules))
}

// ReloadRules 重新加载规则（用于规则更新后刷新缓存）
func (s *CountryParserService) ReloadRules() {
	s.loadRules()
}

// ReloadCountryRules 全局便捷函数：重新加载国家规则
func ReloadCountryRules() {
	service := GetCountryParserService()
	service.ReloadRules()
}

// ParseCountryFromNodeName 全局便捷函数：从节点名称解析国家代码
// 此函数为向后兼容保留，实际调用 models.ParseCountryFromNodeName
func ParseCountryFromNodeName(nodeName string) string {
	return models.ParseCountryFromNodeName(nodeName)
}

// ApplyCountryAutoFillToNode 对单个节点应用国家自动填充
// 参数:
//   - node: 节点指针
//   - autoFill: 是否启用自动填充（新节点）
//   - backfill: 是否启用回填（现存节点）
//
// 返回:
//   - filled: 是否填充了国家
func ApplyCountryAutoFillToNode(node *models.Node, autoFill bool, backfill bool) bool {
	if node == nil {
		return false
	}

	// 新节点：国家为空且启用自动填充
	if autoFill && node.LinkCountry == "" {
		if country := models.ParseCountryFromNodeName(node.LinkName); country != "" {
			node.LinkCountry = country
			utils.Debug("🌍 节点【%s】自动填充国家: %s", node.LinkName, country)
			return true
		}
	}

	// 现存节点：国家为空且启用回填
	if backfill && node.ID > 0 && node.LinkCountry == "" {
		if country := models.ParseCountryFromNodeName(node.LinkName); country != "" {
			node.LinkCountry = country
			utils.Debug("🌍 节点【%s】回填国家: %s", node.LinkName, country)
			return true
		}
	}

	return false
}

// BatchApplyCountryAutoFill 批量对节点应用国家自动填充
// 参数:
//   - nodes: 节点列表指针
//   - autoFill: 是否启用自动填充（新节点）
//   - backfill: 是否启用回填（现存节点）
//
// 返回:
//   - filledCount: 填充的节点数量
func BatchApplyCountryAutoFill(nodes *[]models.Node, autoFill bool, backfill bool) int {
	if nodes == nil || len(*nodes) == 0 {
		return 0
	}

	filledCount := 0
	for i := range *nodes {
		if ApplyCountryAutoFillToNode(&(*nodes)[i], autoFill, backfill) {
			filledCount++
		}
	}

	if filledCount > 0 {
		utils.Info("批量国家填充完成: %d/%d 个节点", filledCount, len(*nodes))
	}

	return filledCount
}
