package api

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sublink/utils"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// ConvertRulesRequest 规则转换请求
type ConvertRulesRequest struct {
	RuleSource string `json:"ruleSource"` // 远程 ACL 配置 URL
	Category   string `json:"category"`   // clash / surge
	Expand     bool   `json:"expand"`     // 是否展开规则
	Template   string `json:"template"`   // 当前模板内容
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
	Name     string   // 组名
	Type     string   // 类型: select, url-test, fallback, load-balance
	Proxies  []string // 代理列表
	URL      string   // 测速 URL (url-test 类型)
	Interval int      // 测速间隔
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
	aclContent, err := fetchRemoteContent(req.RuleSource)
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
		rulesStr, err = generateSurgeRules(rulesets, req.Expand)
	} else {
		proxyGroupsStr = generateClashProxyGroups(proxyGroups)
		rulesStr, err = generateClashRules(rulesets, req.Expand)
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
func fetchRemoteContent(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
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
// 格式: name`type`proxy1`proxy2`...`url`interval
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

		// 检测数字 (interval)
		if matched, _ := regexp.MatchString(`^\d+$`, part); matched {
			fmt.Sscanf(part, "%d", &pg.Interval)
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
	var result []map[string]interface{}

	for _, g := range groups {
		group := map[string]interface{}{
			"name":    g.Name,
			"type":    g.Type,
			"proxies": g.Proxies,
		}

		if g.Type == "url-test" || g.Type == "fallback" {
			if g.URL != "" {
				group["url"] = g.URL
			} else {
				group["url"] = "http://www.gstatic.com/generate_204"
			}
			if g.Interval > 0 {
				group["interval"] = g.Interval
			} else {
				group["interval"] = 300
			}
		}

		result = append(result, group)
	}

	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(map[string]interface{}{
		"proxy-groups": result,
	})
	if err != nil {
		return ""
	}
	return string(yamlBytes)
}

// generateClashRules 生成 Clash 格式的规则
func generateClashRules(rulesets []ACLRuleset, expand bool) (string, error) {
	var rules []string

	if expand {
		// 并发获取所有规则列表
		rules = expandRulesParallel(rulesets)
	} else {
		// 生成 RULE-SET 引用
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
				// 远程规则，使用 RULE-SET
				rules = append(rules, fmt.Sprintf("RULE-SET,%s,%s", rs.RuleURL, rs.Group))
			}
		}
	}

	// 转换为 YAML 格式
	yamlBytes, err := yaml.Marshal(map[string]interface{}{
		"rules": rules,
	})
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}

// expandRulesParallel 并发展开规则
func expandRulesParallel(rulesets []ACLRuleset) []string {
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
				content, err := fetchRemoteContent(ruleset.RuleURL)
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
		proxies := strings.Join(g.Proxies, ",")
		line := fmt.Sprintf("%s = %s,%s", g.Name, g.Type, proxies)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// generateSurgeRules 生成 Surge 格式的规则
func generateSurgeRules(rulesets []ACLRuleset, expand bool) (string, error) {
	var lines []string
	lines = append(lines, "[Rule]")

	if expand {
		// 展开规则
		rules := expandRulesParallel(rulesets)
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
func mergeClashTemplate(template, proxyGroups, rules string) string {
	// 解析现有模板
	var templateData map[string]interface{}
	if err := yaml.Unmarshal([]byte(template), &templateData); err != nil {
		// 模板解析失败，直接追加
		return template + "\n\n# ========== 转换生成的配置 ==========\n\n" + proxyGroups + "\n" + rules
	}

	// 解析生成的代理组
	var proxyGroupsData map[string]interface{}
	if err := yaml.Unmarshal([]byte(proxyGroups), &proxyGroupsData); err == nil {
		if pg, ok := proxyGroupsData["proxy-groups"]; ok {
			templateData["proxy-groups"] = pg
		}
	}

	// 解析生成的规则
	var rulesData map[string]interface{}
	if err := yaml.Unmarshal([]byte(rules), &rulesData); err == nil {
		if r, ok := rulesData["rules"]; ok {
			templateData["rules"] = r
		}
	}

	// 重新生成 YAML
	result, err := yaml.Marshal(templateData)
	if err != nil {
		return template + "\n\n# ========== 转换生成的配置 ==========\n\n" + proxyGroups + "\n" + rules
	}
	return string(result)
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
func getDefaultTemplate(category string) string {
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
