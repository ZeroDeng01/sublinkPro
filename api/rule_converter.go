package api

import (
	"bufio"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sublink/models"
	"sublink/utils"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ConvertRulesRequest 规则转换请求
type ConvertRulesRequest struct {
	RuleSource string `json:"ruleSource"` // 远程 ACL 配置 URL
	Category   string `json:"category"`   // clash / surge
	Expand     bool   `json:"expand"`     // 是否展开规则
	Template   string `json:"template"`   // 当前模板内容
	UseProxy   bool   `json:"useProxy"`   // 是否使用代理
	ProxyLink  string `json:"proxyLink"`  // 代理节点链接（可选）
}

// ConvertRulesResponse 规则转换响应
type ConvertRulesResponse struct {
	Content string `json:"content"` // 转换后的完整模板内容
}

// ACLRuleset ACL 规则集定义
type ACLRuleset struct {
	Group   string // 目标代理组
	RuleURL string // 规则 URL 或内联规则
}

// ACLProxyGroup ACL 代理组定义
type ACLProxyGroup struct {
	Name      string   // 组名
	Type      string   // 类型: select, url-test, fallback, load-balance
	Proxies   []string // 代理列表
	URL       string   // 测速 URL (url-test 类型)
	Interval  int      // 测速间隔
	Tolerance int      // 容差 (url-test 类型)
}

// ConvertRules 规则转换 API
func ConvertRules(c *gin.Context) {
	var req ConvertRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误: "+err.Error())
		return
	}

	if req.RuleSource == "" {
		utils.FailWithMsg(c, "请提供远程规则配置地址")
		return
	}

	if req.Category == "" {
		req.Category = "clash"
	}

	// 检测模板类型与选择的类别是否匹配
	templateType := detectTemplateType(req.Template)
	if templateType != "" && templateType != req.Category {
		utils.FailWithMsg(c, fmt.Sprintf("模板内容与选择的类别不匹配：检测到 %s 格式的模板，但选择的类别是 %s", templateType, req.Category))
		return
	}

	// 如果模板为空，自动补全默认内容
	if strings.TrimSpace(req.Template) == "" {
		req.Template = getDefaultTemplate(req.Category)
	}

	// 获取远程 ACL 配置
	aclContent, err := fetchRemoteContent(req.RuleSource, req.UseProxy, req.ProxyLink)
	if err != nil {
		utils.FailWithMsg(c, "获取远程配置失败: "+err.Error())
		return
	}

	// 解析 ACL 配置
	rulesets, proxyGroups := parseACLConfig(aclContent)

	// 根据类型生成配置
	var proxyGroupsStr, rulesStr string
	if req.Category == "surge" {
		proxyGroupsStr = generateSurgeProxyGroups(proxyGroups)
		rulesStr, err = generateSurgeRules(rulesets, req.Expand, req.UseProxy, req.ProxyLink)
	} else {
		proxyGroupsStr = generateClashProxyGroups(proxyGroups)
		rulesStr, err = generateClashRules(rulesets, req.Expand, req.UseProxy, req.ProxyLink)
	}

	if err != nil {
		utils.FailWithMsg(c, "生成规则失败: "+err.Error())
		return
	}

	// 合并到模板内容
	finalContent := mergeToTemplate(req.Template, proxyGroupsStr, rulesStr, req.Category)

	utils.OkDetailed(c, "ok", ConvertRulesResponse{
		Content: finalContent,
	})
}

// fetchRemoteContent 获取远程内容
// 支持使用代理节点下载
func fetchRemoteContent(url string, useProxy bool, proxyLink string) (string, error) {
	data, err := utils.FetchWithProxy(url, useProxy, proxyLink, 30*time.Second, "")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// parseACLConfig 解析 ACL 配置
func parseACLConfig(content string) ([]ACLRuleset, []ACLProxyGroup) {
	var rulesets []ACLRuleset
	var proxyGroups []ACLProxyGroup

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过注释和空行
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析 ruleset=
		if strings.HasPrefix(line, "ruleset=") {
			parts := strings.SplitN(line[8:], ",", 2)
			if len(parts) == 2 {
				rulesets = append(rulesets, ACLRuleset{
					Group:   strings.TrimSpace(parts[0]),
					RuleURL: strings.TrimSpace(parts[1]),
				})
			}
		}

		// 解析 custom_proxy_group=
		if strings.HasPrefix(line, "custom_proxy_group=") {
			pg := parseProxyGroup(line[19:])
			if pg.Name != "" {
				proxyGroups = append(proxyGroups, pg)
			}
		}
	}

	return rulesets, proxyGroups
}

// parseProxyGroup 解析代理组定义
// 格式: name`type`proxy1`proxy2`...`url`interval,,tolerance
func parseProxyGroup(line string) ACLProxyGroup {
	parts := strings.Split(line, "`")
	if len(parts) < 2 {
		return ACLProxyGroup{}
	}

	pg := ACLProxyGroup{
		Name:    parts[0],
		Type:    parts[1],
		Proxies: make([]string, 0),
	}

	for i := 2; i < len(parts); i++ {
		part := parts[i]

		// 检测测速 URL
		if strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://") {
			pg.URL = part
			continue
		}

		// 检测数字格式 interval,,tolerance 或 interval
		if matched, _ := regexp.MatchString(`^\d+`, part); matched {
			// 检查是否有 ,, 分隔符 (interval,,tolerance)
			if strings.Contains(part, ",") {
				numParts := strings.Split(part, ",")
				if len(numParts) >= 1 && numParts[0] != "" {
					fmt.Sscanf(numParts[0], "%d", &pg.Interval)
				}
				// tolerance 在最后一个非空元素
				for j := len(numParts) - 1; j >= 0; j-- {
					if numParts[j] != "" && j > 0 {
						fmt.Sscanf(numParts[j], "%d", &pg.Tolerance)
						break
					}
				}
			} else {
				fmt.Sscanf(part, "%d", &pg.Interval)
			}
			continue
		}

		// 代理名称，去掉 [] 前缀
		proxyName := part
		if strings.HasPrefix(part, "[]") {
			proxyName = part[2:]
		}

		// 跳过通配符
		if proxyName == ".*" || proxyName == "" {
			continue
		}

		pg.Proxies = append(pg.Proxies, proxyName)
	}

	return pg
}

// generateClashProxyGroups 生成 Clash 格式的代理组
func generateClashProxyGroups(groups []ACLProxyGroup) string {
	var lines []string
	lines = append(lines, "proxy-groups:")

	for _, g := range groups {
		lines = append(lines, fmt.Sprintf("  - name: %s", g.Name))
		lines = append(lines, fmt.Sprintf("    type: %s", g.Type))

		if g.Type == "url-test" || g.Type == "fallback" {
			url := g.URL
			if url == "" {
				url = "http://www.gstatic.com/generate_204"
			}
			lines = append(lines, fmt.Sprintf("    url: %s", url))

			interval := g.Interval
			if interval <= 0 {
				interval = 300
			}
			lines = append(lines, fmt.Sprintf("    interval: %d", interval))

			if g.Tolerance > 0 {
				lines = append(lines, fmt.Sprintf("    tolerance: %d", g.Tolerance))
			}
		}

		lines = append(lines, "    proxies:")
		for _, proxy := range g.Proxies {
			lines = append(lines, fmt.Sprintf("      - %s", proxy))
		}
	}

	return strings.Join(lines, "\n")
}

// generateClashRules 生成 Clash 格式的规则
func generateClashRules(rulesets []ACLRuleset, expand bool, useProxy bool, proxyLink string) (string, error) {
	var rules []string
	var providers []string // rule-providers
	providerIndex := make(map[string]bool)

	if expand {
		// 并发获取所有规则列表
		rules = expandRulesParallel(rulesets, useProxy, proxyLink)
	} else {
		// 生成 RULE-SET 引用 + rule-providers
		for _, rs := range rulesets {
			if strings.HasPrefix(rs.RuleURL, "[]") {
				// 内联规则
				rule := rs.RuleURL[2:] // 去掉 []
				if rule == "GEOIP,CN" {
					rules = append(rules, fmt.Sprintf("GEOIP,CN,%s", rs.Group))
				} else if rule == "FINAL" {
					rules = append(rules, fmt.Sprintf("MATCH,%s", rs.Group))
				} else if strings.HasPrefix(rule, "GEOIP,") {
					geo := strings.TrimPrefix(rule, "GEOIP,")
					rules = append(rules, fmt.Sprintf("GEOIP,%s,%s", geo, rs.Group))
				} else {
					rules = append(rules, fmt.Sprintf("%s,%s", rule, rs.Group))
				}
			} else if strings.HasPrefix(rs.RuleURL, "http") {
				// 远程规则，解析出名称和类型
				providerName, behavior := parseProviderInfo(rs.RuleURL)

				// 生成两个 provider: Domain 和 IP-CIDR
				domainName := providerName + " (Domain)"
				ipcidrName := providerName + " (IP-CIDR)"

				// 添加 Domain 规则
				rules = append(rules, fmt.Sprintf("RULE-SET,%s,%s", domainName, rs.Group))
				// 添加 IP-CIDR 规则
				rules = append(rules, fmt.Sprintf("RULE-SET,%s,%s,no-resolve", ipcidrName, rs.Group))

				// 添加 provider 定义（避免重复）
				if !providerIndex[domainName] {
					providerIndex[domainName] = true
					providers = append(providers, generateProvider(domainName, rs.RuleURL, "domain", behavior))
				}
				if !providerIndex[ipcidrName] {
					providerIndex[ipcidrName] = true
					providers = append(providers, generateProvider(ipcidrName, rs.RuleURL, "ipcidr", behavior))
				}
			}
		}
	}

	// 生成 rules 部分
	var lines []string
	lines = append(lines, "rules:")
	for _, rule := range rules {
		lines = append(lines, fmt.Sprintf("  - %s", rule))
	}

	// 如果有 providers，添加 rule-providers 部分
	if len(providers) > 0 {
		lines = append(lines, "")
		lines = append(lines, "rule-providers:")
		for _, p := range providers {
			lines = append(lines, p)
		}
	}

	return strings.Join(lines, "\n"), nil
}

// parseProviderInfo 从 URL 解析 provider 名称和行为类型
func parseProviderInfo(url string) (name string, behavior string) {
	// 从 URL 提取文件名
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	// 去掉 .list 扩展名
	name = strings.TrimSuffix(filename, ".list")

	// 默认行为类型
	behavior = "classical"

	return name, behavior
}

// generateProvider 生成单个 provider 的 YAML
func generateProvider(name, url, ruleType, behavior string) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("  %s:", name))
	lines = append(lines, "    type: http")
	lines = append(lines, fmt.Sprintf("    behavior: %s", ruleType))
	lines = append(lines, fmt.Sprintf("    url: %s", url))
	lines = append(lines, "    path: ./providers/"+strings.ReplaceAll(name, " ", "_")+".yaml")
	lines = append(lines, "    interval: 86400")
	return strings.Join(lines, "\n")
}

// expandRulesParallel 并发展开规则
func expandRulesParallel(rulesets []ACLRuleset, useProxy bool, proxyLink string) []string {
	type ruleResult struct {
		index int
		rules []string
	}

	results := make(chan ruleResult, len(rulesets))
	var wg sync.WaitGroup

	for i, rs := range rulesets {
		wg.Add(1)
		go func(idx int, ruleset ACLRuleset) {
			defer wg.Done()

			var rules []string
			if strings.HasPrefix(ruleset.RuleURL, "[]") {
				// 内联规则
				rule := ruleset.RuleURL[2:]
				if rule == "GEOIP,CN" {
					rules = append(rules, fmt.Sprintf("GEOIP,CN,%s", ruleset.Group))
				} else if rule == "FINAL" {
					rules = append(rules, fmt.Sprintf("MATCH,%s", ruleset.Group))
				} else if strings.HasPrefix(rule, "GEOIP,") {
					geo := strings.TrimPrefix(rule, "GEOIP,")
					rules = append(rules, fmt.Sprintf("GEOIP,%s,%s", geo, ruleset.Group))
				} else {
					rules = append(rules, fmt.Sprintf("%s,%s", rule, ruleset.Group))
				}
			} else if strings.HasPrefix(ruleset.RuleURL, "http") {
				// 获取远程规则
				content, err := fetchRemoteContent(ruleset.RuleURL, useProxy, proxyLink)
				if err != nil {
					log.Printf("获取规则失败 %s: %v", ruleset.RuleURL, err)
					results <- ruleResult{idx, rules}
					return
				}
				rules = parseRuleList(content, ruleset.Group)
			}
			results <- ruleResult{idx, rules}
		}(i, rs)
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果并按原顺序排序
	orderedResults := make([][]string, len(rulesets))
	for r := range results {
		orderedResults[r.index] = r.rules
	}

	// 合并结果
	var allRules []string
	for _, rules := range orderedResults {
		allRules = append(allRules, rules...)
	}

	return allRules
}

// parseRuleList 解析规则列表文件
func parseRuleList(content string, group string) []string {
	var rules []string
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过注释和空行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 添加目标组
		rules = append(rules, fmt.Sprintf("%s,%s", line, group))
	}

	return rules
}

// generateSurgeProxyGroups 生成 Surge 格式的代理组
func generateSurgeProxyGroups(groups []ACLProxyGroup) string {
	var lines []string
	lines = append(lines, "[Proxy Group]")

	for _, g := range groups {
		// 如果代理列表为空，使用 DIRECT 作为后备
		proxies := g.Proxies
		if len(proxies) == 0 {
			proxies = []string{"DIRECT"}
		}
		proxiesStr := strings.Join(proxies, ", ")
		var line string

		if g.Type == "url-test" || g.Type == "fallback" {
			// url-test 和 fallback 类型需要添加测速参数
			// 格式: name = url-test, proxy1, proxy2, url = xxx, interval = xxx, timeout = 5, tolerance = xxx
			url := g.URL
			if url == "" {
				url = "http://www.gstatic.com/generate_204"
			}
			interval := g.Interval
			if interval <= 0 {
				interval = 600
			}
			tolerance := g.Tolerance
			if tolerance <= 0 {
				tolerance = 200
			}
			line = fmt.Sprintf("%s = %s, %s, url = %s, interval = %d, timeout = 5, tolerance = %d",
				g.Name, g.Type, proxiesStr, url, interval, tolerance)
		} else {
			// select, load-balance 等类型
			line = fmt.Sprintf("%s = %s, %s", g.Name, g.Type, proxiesStr)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// generateSurgeRules 生成 Surge 格式的规则
func generateSurgeRules(rulesets []ACLRuleset, expand bool, useProxy bool, proxyLink string) (string, error) {
	var lines []string
	lines = append(lines, "[Rule]")

	if expand {
		// 展开规则
		rules := expandRulesParallel(rulesets, useProxy, proxyLink)
		for _, rule := range rules {
			// 转换 Clash 格式到 Surge 格式
			// MATCH -> FINAL
			if strings.HasPrefix(rule, "MATCH,") {
				rule = "FINAL," + strings.TrimPrefix(rule, "MATCH,")
			}
			lines = append(lines, rule)
		}
	} else {
		// 生成 RULE-SET 引用
		for _, rs := range rulesets {
			if strings.HasPrefix(rs.RuleURL, "[]") {
				rule := rs.RuleURL[2:]
				if rule == "GEOIP,CN" {
					lines = append(lines, fmt.Sprintf("GEOIP,CN,%s", rs.Group))
				} else if rule == "FINAL" {
					lines = append(lines, fmt.Sprintf("FINAL,%s", rs.Group))
				} else {
					lines = append(lines, fmt.Sprintf("%s,%s", rule, rs.Group))
				}
			} else if strings.HasPrefix(rs.RuleURL, "http") {
				lines = append(lines, fmt.Sprintf("RULE-SET,%s,%s,update-interval=86400", rs.RuleURL, rs.Group))
			}
		}
	}

	return strings.Join(lines, "\n"), nil
}

// mergeToTemplate 将生成的代理组和规则合并到模板内容中
func mergeToTemplate(template, proxyGroups, rules, category string) string {
	if category == "surge" {
		return mergeSurgeTemplate(template, proxyGroups, rules)
	}
	return mergeClashTemplate(template, proxyGroups, rules)
}

// mergeClashTemplate 合并 Clash 模板
// 使用字符串替换方式，避免 yaml.Marshal 转义 emoji
func mergeClashTemplate(template, proxyGroups, rules string) string {
	if strings.TrimSpace(template) == "" {
		// 模板为空，直接返回生成的内容
		return proxyGroups + "\n\n" + rules
	}

	lines := strings.Split(template, "\n")
	var result []string
	skipSection := ""
	sectionsToReplace := map[string]bool{
		"proxy-groups:": true,
		"rules:":        true,
	}

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 检查是否进入需要替换的 section
		if sectionsToReplace[trimmedLine] {
			skipSection = trimmedLine
			continue
		}

		// 如果当前在需要跳过的 section 中
		if skipSection != "" {
			// 检查是否到了新的顶级 key（不以空格开头且以 : 结尾）
			if trimmedLine != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				// 检查下一行是否是列表或嵌套内容
				if strings.HasSuffix(trimmedLine, ":") || (i+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i+1]), "-")) {
					skipSection = ""
					result = append(result, line)
					continue
				}
				skipSection = ""
				result = append(result, line)
				continue
			}
			// 仍在需要跳过的 section 中，跳过此行
			continue
		}

		result = append(result, line)
	}

	// 组合结果
	resultStr := strings.Join(result, "\n")
	resultStr = strings.TrimRight(resultStr, "\n")

	// 添加生成的代理组和规则
	resultStr += "\n\n" + proxyGroups + "\n\n" + rules

	return resultStr
}

// mergeSurgeTemplate 合并 Surge 模板
func mergeSurgeTemplate(template, proxyGroups, rules string) string {
	lines := strings.Split(template, "\n")
	var result []string

	skipSection := ""
	sectionsToReplace := map[string]bool{
		"[Proxy Group]": true,
		"[Rule]":        true,
	}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 检查是否进入需要替换的 section
		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			if sectionsToReplace[trimmedLine] {
				skipSection = trimmedLine
				continue
			} else {
				skipSection = ""
			}
		}

		// 跳过需要替换的 section 的内容
		if skipSection != "" {
			continue
		}

		result = append(result, line)
	}

	// 添加生成的内容
	resultStr := strings.Join(result, "\n")
	resultStr = strings.TrimRight(resultStr, "\n")
	resultStr += "\n\n" + proxyGroups + "\n\n" + rules

	return resultStr
}

// detectTemplateType 检测模板类型
func detectTemplateType(template string) string {
	if strings.TrimSpace(template) == "" {
		return ""
	}

	// Surge 特征: [General], [Proxy], [Proxy Group], [Rule] sections
	surgePatterns := []string{"[General]", "[Proxy]", "[Proxy Group]", "[Rule]"}
	for _, pattern := range surgePatterns {
		if strings.Contains(template, pattern) {
			return "surge"
		}
	}

	// Clash 特征: YAML 格式，包含 port:, proxies:, proxy-groups:, rules:
	clashPatterns := []string{"port:", "proxies:", "proxy-groups:", "rules:", "socks-port:", "dns:", "mode:"}
	for _, pattern := range clashPatterns {
		if strings.Contains(template, pattern) {
			return "clash"
		}
	}

	return ""
}

// getDefaultTemplate 获取默认模板内容
// 优先从系统设置读取，如果未配置则返回硬编码默认值
func getDefaultTemplate(category string) string {
	settingKey := "base_template_" + category
	template, err := models.GetSetting(settingKey)
	if err == nil && strings.TrimSpace(template) != "" {
		return template
	}

	// 回退到硬编码默认值
	if category == "surge" {
		return `[General]
loglevel = notify
bypass-system = true
skip-proxy = 127.0.0.1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,100.64.0.0/10,localhost,*.local,e.crashlytics.com,captive.apple.com,::ffff:0:0:0:0/1,::ffff:128:0:0:0/1
bypass-tun = 192.168.0.0/16,10.0.0.0/8,172.16.0.0/12
dns-server = 119.29.29.29,223.5.5.5,218.30.19.40,61.134.1.4
external-controller-access = password@0.0.0.0:6170
http-api = password@0.0.0.0:6171
test-timeout = 5
http-api-web-dashboard = true
exclude-simple-hostnames = true
allow-wifi-access = true
http-listen = 0.0.0.0:6152
socks5-listen = 0.0.0.0:6153
wifi-access-http-port = 6152
wifi-access-socks5-port = 6153

[Proxy]
DIRECT = direct

`
	}

	// Clash 默认模板
	return `port: 7890
socks-port: 7891
allow-lan: true
mode: Rule
log-level: info
external-controller: :9090
dns:
  enabled: true
  nameserver:
    - 119.29.29.29
    - 223.5.5.5
  fallback:
    - 8.8.8.8
    - 8.8.4.4
    - tls://1.0.0.1:853
    - tls://dns.google:853
proxies: ~

`
}
