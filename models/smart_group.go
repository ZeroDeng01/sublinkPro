package models

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sublink/database"
	"sync"
	"time"

	countryRegexp "github.com/dlclark/regexp2/v2/compat"
	"github.com/robfig/cron/v3"
)

// SmartGroup is an independent, dynamic view of nodes. It never modifies Node.Group.
type SmartGroup struct {
	ID                      int        `gorm:"primaryKey" json:"id"`
	Name                    string     `gorm:"size:191;uniqueIndex;not null" json:"name"`
	Countries               string     `gorm:"size:191;not null" json:"countries"` // comma-separated ISO 3166-1 alpha-2 codes
	Keyword                 string     `gorm:"size:191" json:"keyword"`
	SourceGroups            []string   `gorm:"serializer:json;type:text" json:"sourceGroups"`
	MaxDelay                int        `json:"maxDelay"`
	MinSpeed                float64    `json:"minSpeed"`
	MaxAgeHours             int        `json:"maxAgeHours"` // 0 disables the freshness check
	CheckProfileID          int        `json:"checkProfileId"`
	ScheduleEnabled         bool       `json:"scheduleEnabled"`
	CronExpr                string     `gorm:"size:191" json:"cronExpr"`
	CheckAfterAirportUpdate bool       `json:"checkAfterAirportUpdate"`
	LastCheckAt             *time.Time `json:"lastCheckAt"` // last trigger, not proof every candidate succeeded
	NextCheckAt             *time.Time `gorm:"-" json:"nextCheckAt"`
	CheckRunning            bool       `gorm:"-" json:"checkRunning"`
	CheckTimezone           string     `gorm:"-" json:"checkTimezone"`
	CreatedAt               time.Time  `json:"createdAt"`
	UpdatedAt               time.Time  `json:"updatedAt"`
}

func (g *SmartGroup) Validate() error {
	g.Name = strings.TrimSpace(g.Name)
	if g.Name == "" || len(g.Name) > 191 {
		return fmt.Errorf("智能分组名称不能为空且不能超过191字符")
	}
	seen := map[string]bool{}
	countries := make([]string, 0)
	for _, part := range strings.Split(g.Countries, ",") {
		code := strings.ToUpper(strings.TrimSpace(part))
		if code == "" {
			continue
		}
		if len(code) != 2 || code[0] < 'A' || code[0] > 'Z' || code[1] < 'A' || code[1] > 'Z' {
			return fmt.Errorf("国家必须是两位 ISO 代码")
		}
		if !seen[code] {
			countries = append(countries, code)
			seen[code] = true
		}
	}
	if len(countries) == 0 || len(countries) > 32 {
		return fmt.Errorf("请选择1至32个国家")
	}
	if g.MaxDelay < 0 || g.MinSpeed < 0 || g.MaxAgeHours < 0 || g.MaxAgeHours > 8760 {
		return fmt.Errorf("检测阈值不能为负，结果有效期不能超过一年")
	}
	g.Countries = strings.Join(countries, ",")
	g.Keyword = strings.TrimSpace(g.Keyword)
	if len(g.Keyword) > 191 {
		return fmt.Errorf("关键词不能超过191字符")
	}
	groups := make([]string, 0, len(g.SourceGroups))
	seenGroups := make(map[string]bool)
	for _, source := range g.SourceGroups {
		source = strings.TrimSpace(source)
		if source == "" || len(source) > 191 {
			return fmt.Errorf("原分组名称不能为空且不能超过191字符")
		}
		key := strings.ToLower(source)
		if !seenGroups[key] {
			groups = append(groups, source)
			seenGroups[key] = true
		}
	}
	if len(groups) > 64 {
		return fmt.Errorf("最多选择64个原分组")
	}
	g.SourceGroups = groups
	g.CronExpr = strings.Join(strings.Fields(g.CronExpr), " ")
	if g.CheckProfileID < 0 || ((g.ScheduleEnabled || g.CheckAfterAirportUpdate) && g.CheckProfileID <= 0) {
		return fmt.Errorf("启用自动检测时请选择节点检测策略")
	}
	if g.ScheduleEnabled && g.CronExpr == "" {
		return fmt.Errorf("启用定时检测时必须填写 Cron 表达式")
	}
	if len(g.CronExpr) > 191 {
		return fmt.Errorf("cron 表达式过长")
	}
	if g.CronExpr != "" {
		// UI and schedules intentionally use five fields in the server's local timezone.
		if len(strings.Fields(g.CronExpr)) != 5 {
			return fmt.Errorf("cron 必须是五字段表达式（分 时 日 月 星期）")
		}
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if _, err := parser.Parse(g.CronExpr); err != nil {
			return fmt.Errorf("cron 表达式无效: %w", err)
		}
	}
	return nil
}

func ListSmartGroups() ([]SmartGroup, error) {
	var groups []SmartGroup
	err := database.DB.Order("id ASC").Find(&groups).Error
	return groups, err
}

func GetSmartGroup(id int) (SmartGroup, error) {
	var group SmartGroup
	err := database.DB.First(&group, id).Error
	return group, err
}

// SmartGroupMemberStats explains why matching candidates are not usable.
// Each candidate is assigned to its first failing criterion, so the counts add up.
type SmartGroupMemberStats struct {
	CandidateCount int `json:"candidateCount"`
	DelayUnusable  int `json:"delayUnusable"`
	DelayOverLimit int `json:"delayOverLimit"`
	DelayStale     int `json:"delayStale"`
	SpeedUnusable  int `json:"speedUnusable"`
	SpeedBelowMin  int `json:"speedBelowMin"`
	SpeedStale     int `json:"speedStale"`
}

// inferSmartGroupCountry is used only to label nodes with no stored landing country.
// It never writes the inferred value back to the node.
func inferSmartGroupCountry(node Node, rules []CountryRule) string {
	if code := strings.ToUpper(strings.TrimSpace(node.LinkCountry)); code != "" {
		return code
	}
	patterns := smartGroupCountryRulePatterns(rules)
	for _, name := range []string{node.LinkName, node.Name} {
		if name == "" {
			continue
		}
		for _, rule := range patterns {
			if rule.pattern.MatchString(name) {
				return strings.ToUpper(rule.code)
			}
		}
	}
	return ""
}

// smartGroupCountryPatterns recognizes exact regional flags and standalone ISO tokens.
func smartGroupCountryPatterns(countries map[string]bool) []*regexp.Regexp {
	patterns := make([]*regexp.Regexp, 0, len(countries))
	for code := range countries {
		if len(code) != 2 || code[0] < 'A' || code[0] > 'Z' || code[1] < 'A' || code[1] > 'Z' {
			continue
		}
		flag := string([]rune{rune(0x1F1E6 + int(code[0]-'A')), rune(0x1F1E6 + int(code[1]-'A'))})
		patterns = append(patterns, regexp.MustCompile(regexp.QuoteMeta(flag)+`|(?i)(?:^|[^A-Za-z])`+code+`(?:$|[^A-Za-z])`))
	}
	return patterns
}

// Smart-group matching must not reuse the unrestricted country-rule cache:
// legacy defaults contain bare TR/ES/RO alternatives matching trojan/vmess.
// Cache guarded copies separately by rule ID and pattern; never rewrite saved rules.
type smartGroupCountryRule struct {
	code    string
	pattern *countryRegexp.Regexp
}

type smartGroupCachedCountryPattern struct {
	source  string
	pattern *countryRegexp.Regexp
}

var smartGroupCountryPatternCache = struct {
	sync.Mutex
	entries map[int]smartGroupCachedCountryPattern
}{entries: make(map[int]smartGroupCachedCountryPattern)}

func smartGroupCachedCountryRegex(rule CountryRule) *countryRegexp.Regexp {
	smartGroupCountryPatternCache.Lock()
	defer smartGroupCountryPatternCache.Unlock()
	if cached, ok := smartGroupCountryPatternCache.entries[rule.ID]; ok && cached.source == rule.Pattern {
		return cached.pattern
	}
	pattern, _ := countryRegexp.Compile(smartGroupBoundCountryAbbreviations(rule.Pattern))
	// Bound growth even if custom rules are repeatedly deleted and recreated.
	if len(smartGroupCountryPatternCache.entries) >= 512 {
		clear(smartGroupCountryPatternCache.entries)
	}
	smartGroupCountryPatternCache.entries[rule.ID] = smartGroupCachedCountryPattern{source: rule.Pattern, pattern: pattern}
	return pattern
}

func smartGroupCountryRulePatterns(rules []CountryRule) []smartGroupCountryRule {
	result := make([]smartGroupCountryRule, 0, len(rules))
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		if pattern := smartGroupCachedCountryRegex(rule); pattern != nil {
			result = append(result, smartGroupCountryRule{code: rule.CountryCode, pattern: pattern})
		}
	}
	return result
}

// Only plain two/three-letter alternatives are guarded (ISO codes and aliases
// such as UK/USA/UAE/LA). Explicit regex branches, flags and country/flag names
// keep their semantics. Lookarounds don't consume separators between tokens.
func smartGroupBoundCountryAbbreviations(pattern string) string {
	// Split only top-level alternatives. Pipes inside escaped literals, character
	// classes or groups belong to explicit custom regex and must stay untouched.
	parts := make([]string, 0)
	start, depth := 0, 0
	inClass, firstClassChar, quoted := false, false, false
	for i := 0; i < len(pattern); i++ {
		if quoted {
			if strings.HasPrefix(pattern[i:], `\E`) {
				quoted = false
				i++
			}
			continue
		}
		if pattern[i] == '\\' {
			if !inClass && strings.HasPrefix(pattern[i:], `\Q`) {
				quoted = true
			}
			firstClassChar = false
			i++ // Skip the escaped byte, including an escaped pipe or bracket.
			continue
		}
		if inClass {
			// A leading ] (also after ^) is a literal class member.
			if pattern[i] == ']' && !firstClassChar {
				inClass = false
			}
			if pattern[i] != '^' || !firstClassChar {
				firstClassChar = false
			}
			continue
		}
		switch pattern[i] {
		case '[':
			inClass, firstClassChar = true, true
		case '(':
			if strings.HasPrefix(pattern[i:], "(?#") {
				if end := strings.IndexByte(pattern[i:], ')'); end >= 0 {
					i += end
					continue
				}
			}
			// Extended-mode comments can contain arbitrary pipes and brackets.
			// Leave such explicitly customized expressions wholly untouched.
			if strings.HasPrefix(pattern[i:], "(?") {
				for j := i + 2; j < len(pattern) && strings.ContainsRune("imsxU-", rune(pattern[j])); j++ {
					if pattern[j] == 'x' {
						return pattern
					}
				}
			}
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case '|':
			if depth == 0 {
				parts = append(parts, pattern[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, pattern[start:])
	for index, part := range parts {
		prefix := ""
		// Preserve leading global inline flags before a plain abbreviation.
		for strings.HasPrefix(part, "(?") {
			end := strings.IndexByte(part, ')')
			if end < 0 {
				break
			}
			flags := part[2:end]
			if flags == "" || strings.Trim(flags, "imsxU-") != "" {
				break
			}
			prefix += part[:end+1]
			part = part[end+1:]
		}
		if len(part) < 2 || len(part) > 3 {
			continue
		}
		ascii := true
		for _, letter := range part {
			if (letter < 'A' || letter > 'Z') && (letter < 'a' || letter > 'z') {
				ascii = false
				break
			}
		}
		if ascii {
			parts[index] = prefix + "(?<![A-Za-z])" + part + "(?![A-Za-z])"
		}
	}
	return strings.Join(parts, "|")
}

func smartGroupSelectedCountryRules(countries map[string]bool) []smartGroupCountryRule {
	selected := make([]CountryRule, 0)
	for _, rule := range GetEnabledCountryRules() {
		if countries[strings.ToUpper(rule.CountryCode)] {
			selected = append(selected, rule)
		}
	}
	return smartGroupCountryRulePatterns(selected)
}

// smartGroupMatchSource defines the OR relation between stored country, country
// clues in the node name, and the optional node-name keyword.
func smartGroupMatchSource(node Node, countries map[string]bool, codePatterns []*regexp.Regexp, rules []smartGroupCountryRule, keyword string) string {
	if countries[strings.ToUpper(strings.TrimSpace(node.LinkCountry))] {
		return "landingCountry"
	}
	names := []string{node.EffectiveName(), node.LinkName, node.Name}
	for _, rule := range rules {
		for _, name := range names {
			if name != "" && rule.pattern.MatchString(name) {
				return "countryName"
			}
		}
	}
	// An exact flag or selected ISO token also works without a configured country rule.
	for _, pattern := range codePatterns {
		for _, name := range names {
			if pattern.MatchString(name) {
				return "countryName"
			}
		}
	}
	if keyword != "" {
		for _, name := range names {
			if strings.Contains(strings.ToLower(name), keyword) {
				return "keyword"
			}
		}
	}
	return ""
}

// MatchSource labels why a candidate matched, even when its actual landing
// country differs from the country suggested by its name.
func (g SmartGroup) MatchSource(node Node) string {
	countries := make(map[string]bool)
	for _, part := range strings.Split(g.Countries, ",") {
		if code := strings.ToUpper(strings.TrimSpace(part)); code != "" {
			countries[code] = true
		}
	}
	if len(countries) == 0 {
		return ""
	}
	return smartGroupMatchSource(node, countries, smartGroupCountryPatterns(countries), smartGroupSelectedCountryRules(countries), strings.ToLower(strings.TrimSpace(g.Keyword)))
}

// CandidateMatcher 为当前页/请求预编译匹配条件，并包含原分组限制。
// 返回空字符串表示非候选，不能再将健康检查的空排除原因误解为入组。
func (g SmartGroup) CandidateMatcher() func(Node) string {
	countries := make(map[string]bool)
	for _, part := range strings.Split(g.Countries, ",") {
		if code := strings.ToUpper(strings.TrimSpace(part)); code != "" {
			countries[code] = true
		}
	}
	groups := make(map[string]bool, len(g.SourceGroups))
	for _, source := range g.SourceGroups {
		groups[strings.ToLower(source)] = true
	}
	patterns, rules := smartGroupCountryPatterns(countries), smartGroupSelectedCountryRules(countries)
	keyword := strings.ToLower(strings.TrimSpace(g.Keyword))
	return func(node Node) string {
		if len(countries) == 0 || (len(groups) > 0 && !groups[strings.ToLower(node.Group)]) {
			return ""
		}
		return smartGroupMatchSource(node, countries, patterns, rules, keyword)
	}
}

// Candidates intentionally ignores current health so failed nodes can be retested.
// Country, name-country clues and an optional keyword are alternatives; source groups remain an AND filter.
func (g SmartGroup) Candidates() ([]Node, error) {
	nodes, err := (&Node{}).ListWithFilters(NodeFilter{})
	if err != nil {
		return nil, err
	}
	countries := make(map[string]bool)
	for _, code := range strings.Split(g.Countries, ",") {
		if code = strings.ToUpper(strings.TrimSpace(code)); code != "" {
			countries[code] = true
		}
	}
	if len(countries) == 0 {
		return []Node{}, nil
	}
	groups := make(map[string]bool, len(g.SourceGroups))
	for _, source := range g.SourceGroups {
		groups[strings.ToLower(source)] = true
	}
	keyword := strings.ToLower(strings.TrimSpace(g.Keyword))
	countryRules := smartGroupSelectedCountryRules(countries)
	codePatterns := smartGroupCountryPatterns(countries)
	candidates := make([]Node, 0)
	for _, node := range nodes {
		if len(groups) > 0 && !groups[strings.ToLower(node.Group)] {
			continue
		}
		if smartGroupMatchSource(node, countries, codePatterns, countryRules, keyword) != "" {
			candidates = append(candidates, node)
		}
	}
	return candidates, nil
}

func (g SmartGroup) CandidateIDs() ([]int, error) {
	nodes, err := g.Candidates()
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.ID)
	}
	sort.Ints(ids)
	return ids, nil
}

// Members evaluates persisted checks dynamically. Zero minimum speed accepts TCP-only latency checks.
func (g SmartGroup) Members() ([]Node, error) {
	members, _, err := g.MembersWithStats()
	return members, err
}

func (g SmartGroup) MembersWithStats() ([]Node, SmartGroupMemberStats, error) {
	_, members, stats, _, err := g.EvaluateMembers()
	return members, stats, err
}

// EvaluateMembers 在同一候选快照和有效期边界下计算成员及统计，供分页 API 复用。
// Candidates 已按 ID 排序，成员为其有序子序列，无需再次排序。
func (g SmartGroup) EvaluateMembers() (candidates, members []Node, stats SmartGroupMemberStats, cutoff time.Time, err error) {
	candidates, err = g.Candidates()
	if err != nil {
		return
	}
	stats.CandidateCount = len(candidates)
	cutoff = time.Now().Add(-time.Duration(g.MaxAgeHours) * time.Hour)
	members = make([]Node, 0, len(candidates))
	for _, node := range candidates {
		switch g.CandidateExclusionReason(node, cutoff) {
		case "delayUnusable":
			stats.DelayUnusable++
		case "delayOverLimit":
			stats.DelayOverLimit++
		case "delayStale":
			stats.DelayStale++
		case "speedUnusable":
			stats.SpeedUnusable++
		case "speedBelowMin":
			stats.SpeedBelowMin++
		case "speedStale":
			stats.SpeedStale++
		default:
			members = append(members, node)
		}
	}
	return
}

// CandidateExclusionReason returns an empty string for a usable node.
// The caller supplies one cutoff for a consistent snapshot of the current result window.
func (g SmartGroup) CandidateExclusionReason(node Node, cutoff time.Time) string {
	if node.DelayStatus != "success" || node.DelayTime <= 0 {
		return "delayUnusable"
	}
	if g.MaxDelay > 0 && node.DelayTime > g.MaxDelay {
		return "delayOverLimit"
	}
	if g.MaxAgeHours > 0 {
		delayAt, err := time.ParseInLocation("2006-01-02 15:04:05", node.LatencyCheckAt, time.Local)
		if err != nil || delayAt.Before(cutoff) {
			return "delayStale"
		}
	}
	if g.MinSpeed > 0 {
		if node.SpeedStatus != "success" || node.Speed <= 0 {
			return "speedUnusable"
		}
		if node.Speed < g.MinSpeed {
			return "speedBelowMin"
		}
		if g.MaxAgeHours > 0 {
			speedAt, err := time.ParseInLocation("2006-01-02 15:04:05", node.SpeedCheckAt, time.Local)
			if err != nil || speedAt.Before(cutoff) {
				return "speedStale"
			}
		}
	}
	return ""
}

// SmartGroupCountryForDisplay labels name-inferred countries without changing the persisted node.
func SmartGroupCountryForDisplay(node Node) (code, source string) {
	if strings.TrimSpace(node.LinkCountry) != "" {
		return strings.ToUpper(strings.TrimSpace(node.LinkCountry)), "stored"
	}
	if code := inferSmartGroupCountry(node, GetEnabledCountryRules()); code != "" {
		return code, "name"
	}
	return "", "unknown"
}

// ParseSmartGroupIDs parses the subscription's compact, ordered ID list.
func ParseSmartGroupIDs(value string) []int {
	seen := map[int]bool{}
	ids := make([]int, 0)
	for _, part := range strings.Split(value, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && id > 0 && !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	return ids
}

func NormalizeSmartGroupIDs(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	for _, part := range strings.Split(value, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || id <= 0 {
			return "", fmt.Errorf("智能分组ID无效")
		}
	}
	ids := ParseSmartGroupIDs(value)
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, err := GetSmartGroup(id); err != nil {
			return "", fmt.Errorf("智能分组 %d 不存在", id)
		}
		parts = append(parts, strconv.Itoa(id))
	}
	return strings.Join(parts, ","), nil
}
