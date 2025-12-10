package utils

import (
	"regexp"
	"strings"
)

// IsRegexProxyPattern 检测是否是正则代理模式
// 格式: (选项1|选项2|选项3)
func IsRegexProxyPattern(proxy string) bool {
	proxy = strings.TrimSpace(proxy)
	if len(proxy) < 3 {
		return false
	}
	// 检查是否以 ( 开头，以 ) 结尾，且包含 |
	return strings.HasPrefix(proxy, "(") && strings.HasSuffix(proxy, ")") && strings.Contains(proxy, "|")
}

// MatchNodesByRegexPattern 根据正则模式匹配节点名称
// pattern: 格式如 (网易|音乐|解锁|Music|NetEase)
// nodeNames: 所有可用的节点名称列表
// 返回匹配的节点名称列表
func MatchNodesByRegexPattern(pattern string, nodeNames []string) []string {
	if !IsRegexProxyPattern(pattern) {
		return nil
	}

	// 移除首尾括号，获取选项列表
	inner := pattern[1 : len(pattern)-1]
	options := strings.Split(inner, "|")

	var matched []string
	for _, nodeName := range nodeNames {
		nodeNameLower := strings.ToLower(nodeName)
		for _, option := range options {
			optionLower := strings.ToLower(strings.TrimSpace(option))
			if optionLower != "" && strings.Contains(nodeNameLower, optionLower) {
				matched = append(matched, nodeName)
				break // 只要匹配一个选项就添加这个节点
			}
		}
	}

	return matched
}

// ProcessProxyListWithRegex 处理代理列表中的正则模式
// proxies: 代理列表（可能包含正则模式）
// nodeNames: 所有可用的节点名称
// 返回处理后的代理列表
func ProcessProxyListWithRegex(proxies []string, nodeNames []string) []string {
	var result []string

	for _, proxy := range proxies {
		if IsRegexProxyPattern(proxy) {
			// 展开正则模式为匹配的节点
			matchedNodes := MatchNodesByRegexPattern(proxy, nodeNames)
			result = append(result, matchedNodes...)
		} else {
			// 保留原始代理
			result = append(result, proxy)
		}
	}

	return result
}

// ProcessProxyListWithRegexInterface 处理 interface{} 类型的代理列表
func ProcessProxyListWithRegexInterface(proxies []interface{}, nodeNames []string) []interface{} {
	var result []interface{}

	for _, proxy := range proxies {
		proxyStr, ok := proxy.(string)
		if !ok {
			result = append(result, proxy)
			continue
		}

		if IsRegexProxyPattern(proxyStr) {
			// 展开正则模式为匹配的节点
			matchedNodes := MatchNodesByRegexPattern(proxyStr, nodeNames)
			for _, node := range matchedNodes {
				result = append(result, node)
			}
		} else {
			// 保留原始代理
			result = append(result, proxy)
		}
	}

	return result
}

// ExtractKeywordsFromRegexPattern 从正则模式中提取关键词
func ExtractKeywordsFromRegexPattern(pattern string) []string {
	if !IsRegexProxyPattern(pattern) {
		return nil
	}

	inner := pattern[1 : len(pattern)-1]
	options := strings.Split(inner, "|")

	var keywords []string
	for _, opt := range options {
		opt = strings.TrimSpace(opt)
		if opt != "" {
			keywords = append(keywords, opt)
		}
	}

	return keywords
}

// MatchNodeNameByKeywords 检查节点名称是否匹配任意关键词
func MatchNodeNameByKeywords(nodeName string, keywords []string) bool {
	nodeNameLower := strings.ToLower(nodeName)
	for _, keyword := range keywords {
		if strings.Contains(nodeNameLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// CompileRegexPattern 将简单模式编译为正则表达式
func CompileRegexPattern(pattern string) (*regexp.Regexp, error) {
	if !IsRegexProxyPattern(pattern) {
		return nil, nil
	}

	// 转换为正则表达式格式
	inner := pattern[1 : len(pattern)-1]
	regexPattern := "(?i)(" + inner + ")" // 不区分大小写

	return regexp.Compile(regexPattern)
}
