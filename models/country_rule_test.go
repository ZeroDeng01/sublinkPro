package models

import (
	"strings"
	"testing"

	"sublink/cache"
	"sublink/database"
	"sublink/internal/testutil"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// resetCountryRuleCacheForTest 重置国家规则缓存（用于测试隔离）
func resetCountryRuleCacheForTest() {
	countryRuleCache = cache.NewMapCache(func(r CountryRule) int { return r.ID })
	countryRuleCache.AddIndex("enabled", func(r CountryRule) string {
		if r.Enabled {
			return "true"
		}
		return "false"
	})
	countryRuleCache.AddIndex("countryCode", func(r CountryRule) string { return r.CountryCode })
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) {
	t.Helper()

	// 保存原有状态
	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	// 创建独立的测试数据库
	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "country_rule_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	// 运行迁移
	if err := db.AutoMigrate(&CountryRule{}); err != nil {
		t.Fatalf("auto migrate country_rules: %v", err)
	}

	// 设置测试环境
	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	resetCountryRuleCacheForTest()

	// 初始化国家规则缓存
	if err := InitCountryRuleCache(); err != nil {
		t.Fatalf("init country rule cache: %v", err)
	}

	// 注册清理函数
	t.Cleanup(func() {
		testutil.CloseDB(t, db)
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
	})
}

// cleanupTestDB 清理测试数据库（已废弃，使用 t.Cleanup）
func cleanupTestDB(t *testing.T) {
	// 保留空函数以兼容现有测试，实际清理在 setupTestDB 的 t.Cleanup 中
}

// createTestRule 创建测试规则
func createTestRule(t *testing.T, code string, name string, pattern string, priority int) *CountryRule {
	t.Helper()
	rule := &CountryRule{
		CountryCode: code,
		CountryName: name,
		Pattern:     pattern,
		Priority:    priority,
		Enabled:     true,
	}
	if err := rule.Add(); err != nil {
		t.Fatalf("create test rule: %v", err)
	}
	return rule
}

func TestCountryRulePatternMatching(t *testing.T) {
	tests := []struct {
		pattern  string
		nodeName string
		expected bool
	}{
		{"香港|HK|Hong Kong|🇭🇰", "香港节点01", true},
		{"香港|HK|Hong Kong|🇭🇰", "HK-001", true},
		{"香港|HK|Hong Kong|🇭🇰", "Hong Kong Server", true},
		{"香港|HK|Hong Kong|🇭🇰", "🇭🇰 节点", true},
		{"香港|HK|Hong Kong|🇭🇰", "台湾节点", false},
		{"美国|US|USA|United States|🇺🇸", "美国节点01", true},
		{"美国|US|USA|United States|🇺🇸", "US-001", true},
		{"美国|US|USA|United States|🇺🇸", "United States Server", true},
		{"美国|US|USA|United States|🇺🇸", "🇺🇸 节点", true},
		{"美国|US|USA|United States|🇺🇸", "日本节点", false},
		// 测试复杂正则表达式（不区分大小写、带空格）
		{"(?i)(香港|HK|Hong\\s*Kong)", "hong kong", true},
		{"(?i)(香港|HK|Hong\\s*Kong)", "HongKong", true},
		{"(?i)(香港|HK|Hong\\s*Kong)", "Hong   Kong", true},
		{"(?i)(美国|USA?|United\\s*States)", "usa", true},
		{"(?i)(美国|USA?|United\\s*States)", "us", true},
		{"(?i)(美国|USA?|United\\s*States)", "United States", true},
	}

	for _, tt := range tests {
		matched, err := TestPattern(tt.pattern, tt.nodeName)
		if err != nil {
			t.Errorf("TestPattern(%q, %q) returned error: %v", tt.pattern, tt.nodeName, err)
			continue
		}
		if matched != tt.expected {
			t.Errorf("TestPattern(%q, %q) = %v, expected %v", tt.pattern, tt.nodeName, matched, tt.expected)
		}
	}
}

func TestCountryRuleValidation(t *testing.T) {
	tests := []struct {
		name        string
		rule        CountryRule
		expectError bool
	}{
		{
			name: "valid rule",
			rule: CountryRule{
				CountryCode: "US",
				CountryName: "美国",
				Pattern:     "美国|US",
				Priority:    0,
				Enabled:     true,
			},
			expectError: false,
		},
		{
			name: "empty country code",
			rule: CountryRule{
				CountryCode: "",
				CountryName: "美国",
				Pattern:     "美国|US",
				Priority:    0,
				Enabled:     true,
			},
			expectError: true,
		},
		{
			name: "empty country name",
			rule: CountryRule{
				CountryCode: "US",
				CountryName: "",
				Pattern:     "美国|US",
				Priority:    0,
				Enabled:     true,
			},
			expectError: true,
		},
		{
			name: "empty pattern",
			rule: CountryRule{
				CountryCode: "US",
				CountryName: "美国",
				Pattern:     "",
				Priority:    0,
				Enabled:     true,
			},
			expectError: true,
		},
		{
			name: "invalid regex pattern",
			rule: CountryRule{
				CountryCode: "US",
				CountryName: "美国",
				Pattern:     "[invalid(regex",
				Priority:    0,
				Enabled:     true,
			},
			expectError: true,
		},
		{
			name: "priority too high",
			rule: CountryRule{
				CountryCode: "US",
				CountryName: "美国",
				Pattern:     "美国|US",
				Priority:    1001,
				Enabled:     true,
			},
			expectError: true,
		},
		{
			name: "priority negative",
			rule: CountryRule{
				CountryCode: "US",
				CountryName: "美国",
				Pattern:     "美国|US",
				Priority:    -1,
				Enabled:     true,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.validate()
			if tt.expectError && err == nil {
				t.Errorf("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestCountryRuleAdd_DisabledRule(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// 创建禁用的规则
	rule := &CountryRule{
		CountryCode: "TEST",
		CountryName: "测试",
		Pattern:     "test",
		Priority:    50,
		Enabled:     false,
	}

	if err := rule.Add(); err != nil {
		t.Fatalf("Failed to add disabled rule: %v", err)
	}

	// 从数据库验证
	var dbRule CountryRule
	if err := database.DB.Where("country_code = ?", "TEST").First(&dbRule).Error; err != nil {
		t.Fatalf("Failed to query rule: %v", err)
	}

	if dbRule.Enabled {
		t.Errorf("Expected rule to be disabled, but it is enabled")
	}

	// 从缓存验证
	var cacheRule CountryRule
	cacheRule.ID = rule.ID
	if err := cacheRule.GetByID(); err != nil {
		t.Fatalf("Failed to get rule from cache: %v", err)
	}

	if cacheRule.Enabled {
		t.Errorf("Expected cached rule to be disabled, but it is enabled")
	}
}

func TestCountryRuleNormalize(t *testing.T) {
	rule := CountryRule{
		CountryCode: " us ",
		CountryName: " 美国 ",
		Pattern:     " 美国|US ",
		Priority:    0,
		Enabled:     true,
	}

	rule.normalize()

	if rule.CountryCode != "US" {
		t.Errorf("Expected CountryCode to be normalized to 'US', got %q", rule.CountryCode)
	}
	if rule.CountryName != "美国" {
		t.Errorf("Expected CountryName to be trimmed, got %q", rule.CountryName)
	}
	if rule.Pattern != "美国|US" {
		t.Errorf("Expected Pattern to be trimmed, got %q", rule.Pattern)
	}
}

// ========== 文本导出功能测试 ==========

func TestGetCountryNameByCode(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// 添加测试规则
	rules := []CountryRule{
		{
			CountryCode: "CN",
			CountryName: "中国",
			Pattern:     "(?i)(中国|china|cn)",
			Priority:    100,
			Enabled:     true,
		},
		{
			CountryCode: "HK",
			CountryName: "香港",
			Pattern:     "(?i)(香港|hong kong|hk)",
			Priority:    90,
			Enabled:     true,
		},
		{
			CountryCode: "US",
			CountryName: "美国",
			Pattern:     "(?i)(美国|united states|usa|us)",
			Priority:    80,
			Enabled:     true,
		},
	}

	for _, rule := range rules {
		if err := rule.Add(); err != nil {
			t.Fatalf("Failed to add rule %s: %v", rule.CountryCode, err)
		}
	}

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "有匹配的国家代码 - HK",
			code:     "HK",
			expected: "香港",
		},
		{
			name:     "有匹配的国家代码 - CN",
			code:     "CN",
			expected: "中国",
		},
		{
			name:     "有匹配的国家代码 - US",
			code:     "US",
			expected: "美国",
		},
		{
			name:     "小写国家代码会自动转大写 - hk",
			code:     "hk",
			expected: "香港",
		},
		{
			name:     "混合大小写 - Hk",
			code:     "Hk",
			expected: "香港",
		},
		{
			name:     "没有匹配的国家代码 - XX",
			code:     "XX",
			expected: "",
		},
		{
			name:     "空字符串",
			code:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCountryNameByCode(tt.code)
			if result != tt.expected {
				t.Errorf("GetCountryNameByCode(%q) = %q, expected %q", tt.code, result, tt.expected)
			}
		})
	}
}

// ========== 文本导出功能测试 ==========

func TestExportCountryRulesToText_Empty(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	text := ExportCountryRulesToText()

	// 验证包含注释头
	if !strings.Contains(text, "# 国家规则配置") {
		t.Error("Expected text to contain header comment")
	}
	if !strings.Contains(text, "# 格式：国家代码 国家名称 优先级 启用状态 正则表达式") {
		t.Error("Expected text to contain format comment")
	}
}

func TestExportCountryRulesToText_SingleRule(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// 创建单条规则
	createTestRule(t, "CN", "中国", "中国|China|CN", 100)

	text := ExportCountryRulesToText()

	// 验证包含规则
	if !strings.Contains(text, "CN 中国 100 true 中国|China|CN") {
		t.Errorf("Expected text to contain rule, got: %s", text)
	}
}

func TestExportCountryRulesToText_MultipleRules(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// 创建多条规则
	createTestRule(t, "CN", "中国", "中国|China|CN", 100)
	createTestRule(t, "US", "美国", "美国|USA?|US", 90)
	createTestRule(t, "JP", "日本", "日本|Japan|JP", 80)

	text := ExportCountryRulesToText()

	// 验证所有规则都存在
	if !strings.Contains(text, "CN 中国 100 true") {
		t.Error("Expected text to contain CN rule")
	}
	if !strings.Contains(text, "US 美国 90 true") {
		t.Error("Expected text to contain US rule")
	}
	if !strings.Contains(text, "JP 日本 80 true") {
		t.Error("Expected text to contain JP rule")
	}

	// 验证排序（按优先级降序）
	lines := strings.Split(text, "\n")
	var ruleLines []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			ruleLines = append(ruleLines, line)
		}
	}

	// CN(100) 应该在 US(90) 之前，US(90) 应该在 JP(80) 之前
	cnIndex, usIndex, jpIndex := -1, -1, -1
	for i, line := range ruleLines {
		if strings.HasPrefix(line, "CN ") {
			cnIndex = i
		} else if strings.HasPrefix(line, "US ") {
			usIndex = i
		} else if strings.HasPrefix(line, "JP ") {
			jpIndex = i
		}
	}

	if cnIndex == -1 || usIndex == -1 || jpIndex == -1 {
		t.Error("Not all rules found in export")
	}
	if cnIndex >= usIndex || usIndex >= jpIndex {
		t.Errorf("Rules not sorted by priority: CN at %d, US at %d, JP at %d", cnIndex, usIndex, jpIndex)
	}
}

func TestExportCountryRulesToText_SpecialCharacters(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// 创建包含特殊字符的规则
	createTestRule(t, "HK", "香港🇭🇰", "(?i)(香港|HK|Hong\\s*Kong)", 100)

	text := ExportCountryRulesToText()

	// 验证特殊字符被正确导出
	if !strings.Contains(text, "HK 香港🇭🇰 100 true (?i)(香港|HK|Hong\\s*Kong)") {
		t.Errorf("Expected text to contain rule with special characters, got: %s", text)
	}
}

func TestExportCountryRulesToText_DisabledRule(t *testing.T) {

	setupTestDB(t)
	defer cleanupTestDB(t)

	// 使用同步功能创建规则（包含禁用规则）
	text := `CN 中国 100 true 中国|China|CN
TEST 测试 50 false test`

	added, _, _, err := SyncCountryRulesFromText(text)
	if err != nil {
		t.Fatalf("Failed to sync rules: %v", err)
	}

	if added != 2 {
		t.Errorf("Expected 2 added, got %d", added)
	}

	// 验证数据库中的禁用状态
	var testRule CountryRule
	if err := database.DB.Where("country_code = ?", "TEST").First(&testRule).Error; err != nil {
		t.Fatalf("Failed to find TEST rule: %v", err)
	}

	if testRule.Enabled {
		t.Error("Expected TEST rule to be disabled, but it is enabled")
	}

	// 导出并验证
	exportedText := ExportCountryRulesToText()

	// 验证包含两条规则
	lines := strings.Split(exportedText, "\n")
	ruleCount := 0
	hasDisabledRule := false
	hasCNRule := false
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			ruleCount++
			if strings.Contains(line, "TEST 测试 50 false test") {
				hasDisabledRule = true
			}
			if strings.Contains(line, "CN 中国 100 true") {
				hasCNRule = true
			}
		}
	}

	if ruleCount != 2 {
		t.Errorf("Expected 2 rules in export, got %d. Text: %s", ruleCount, exportedText)
	}

	if !hasCNRule {
		t.Errorf("Expected text to contain CN rule, got: %s", exportedText)
	}

	if !hasDisabledRule {
		t.Errorf("Expected text to contain disabled TEST rule, got: %s", exportedText)
	}
}

// ========== 文本解析功能测试 ==========

func TestParseCountryRuleText_ValidFormat(t *testing.T) {
	text := `CN 中国 100 true 中国|China|CN
US 美国 90 true 美国|USA?|US`

	rules, errors := parseCountryRuleText(text)

	if len(errors) != 0 {
		t.Errorf("Expected no errors, got: %v", errors)
	}
	if len(rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(rules))
	}

	// 验证第一条规则
	if rules[0].CountryCode != "CN" {
		t.Errorf("Expected CountryCode 'CN', got %q", rules[0].CountryCode)
	}
	if rules[0].CountryName != "中国" {
		t.Errorf("Expected CountryName '中国', got %q", rules[0].CountryName)
	}
	if rules[0].Priority != 100 {
		t.Errorf("Expected Priority 100, got %d", rules[0].Priority)
	}
	if !rules[0].Enabled {
		t.Error("Expected Enabled to be true")
	}
	if rules[0].Pattern != "中国|China|CN" {
		t.Errorf("Expected Pattern '中国|China|CN', got %q", rules[0].Pattern)
	}
}

func TestParseCountryRuleText_SkipComments(t *testing.T) {
	text := `# 这是注释
CN 中国 100 true 中国|China|CN
# 另一条注释
US 美国 90 true 美国|USA?|US`

	rules, errors := parseCountryRuleText(text)

	if len(errors) != 0 {
		t.Errorf("Expected no errors, got: %v", errors)
	}
	if len(rules) != 2 {
		t.Errorf("Expected 2 rules (comments should be skipped), got %d", len(rules))
	}
}

func TestParseCountryRuleText_SkipEmptyLines(t *testing.T) {
	text := `CN 中国 100 true 中国|China|CN

US 美国 90 true 美国|USA?|US

`

	rules, errors := parseCountryRuleText(text)

	if len(errors) != 0 {
		t.Errorf("Expected no errors, got: %v", errors)
	}
	if len(rules) != 2 {
		t.Errorf("Expected 2 rules (empty lines should be skipped), got %d", len(rules))
	}
}

func TestParseCountryRuleText_InvalidFieldCount(t *testing.T) {
	text := `CN 中国 100`

	rules, errors := parseCountryRuleText(text)

	if len(errors) == 0 {
		t.Error("Expected parsing error for invalid field count")
	}
	if len(rules) != 0 {
		t.Errorf("Expected 0 rules, got %d", len(rules))
	}
	if !strings.Contains(errors[0], "格式错误") {
		t.Errorf("Expected format error message, got: %s", errors[0])
	}
}

func TestParseCountryRuleText_InvalidPriority(t *testing.T) {
	text := `CN 中国 abc true 中国|China|CN`

	_, errors := parseCountryRuleText(text)

	if len(errors) == 0 {
		t.Error("Expected parsing error for invalid priority")
	}
	if !strings.Contains(errors[0], "优先级必须是整数") {
		t.Errorf("Expected priority error message, got: %s", errors[0])
	}
}

func TestParseCountryRuleText_PriorityOutOfRange(t *testing.T) {
	text := `CN 中国 1001 true 中国|China|CN`

	_, errors := parseCountryRuleText(text)

	if len(errors) == 0 {
		t.Error("Expected parsing error for priority out of range")
	}
	if !strings.Contains(errors[0], "优先级必须在0-1000之间") {
		t.Errorf("Expected priority range error message, got: %s", errors[0])
	}
}

func TestParseCountryRuleText_InvalidEnabled(t *testing.T) {
	text := `CN 中国 100 yes 中国|China|CN`

	_, errors := parseCountryRuleText(text)

	if len(errors) == 0 {
		t.Error("Expected parsing error for invalid enabled value")
	}
	if !strings.Contains(errors[0], "启用状态必须是true/false或1/0") {
		t.Errorf("Expected enabled error message, got: %s", errors[0])
	}
}

func TestParseCountryRuleText_InvalidRegex(t *testing.T) {
	text := `CN 中国 100 true [invalid(regex`

	_, errors := parseCountryRuleText(text)

	if len(errors) == 0 {
		t.Error("Expected parsing error for invalid regex")
	}
	if !strings.Contains(errors[0], "正则表达式编译失败") {
		t.Errorf("Expected regex error message, got: %s", errors[0])
	}
}

func TestParseCountryRuleText_EmptyCountryCode(t *testing.T) {
	// 使用空字符串作为国家代码（5个字段，第一个为空）
	text := ` ` + ` 中国 100 true 中国|China|CN`

	_, errors := parseCountryRuleText(text)

	if len(errors) == 0 {
		t.Error("Expected parsing error for empty country code")
	}
	// 接受两种可能的错误消息：字段数量错误或国家代码不能为空
	hasError := strings.Contains(errors[0], "国家代码") || strings.Contains(errors[0], "格式错误")
	if !hasError {
		t.Errorf("Expected country code or format error message, got: %s", errors[0])
	}
}

func TestParseCountryRuleText_EmptyCountryName(t *testing.T) {
	// 使用空字符串作为国家名称（5个字段，第二个为空）
	text := `CN ` + ` 100 true 中国|China|CN`

	_, errors := parseCountryRuleText(text)

	if len(errors) == 0 {
		t.Error("Expected parsing error for empty country name")
	}
	// 接受两种可能的错误消息：字段数量错误或国家名称不能为空
	hasError := strings.Contains(errors[0], "国家名称") || strings.Contains(errors[0], "格式错误")
	if !hasError {
		t.Errorf("Expected country name or format error message, got: %s", errors[0])
	}
}

func TestParseCountryRuleText_PatternWithSpaces(t *testing.T) {
	text := `HK 香港 100 true (?i)(香港|HK|Hong\s*Kong)`

	rules, errors := parseCountryRuleText(text)

	if len(errors) != 0 {
		t.Errorf("Expected no errors for pattern with spaces, got: %v", errors)
	}
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
	if rules[0].Pattern != `(?i)(香港|HK|Hong\s*Kong)` {
		t.Errorf("Expected pattern to be preserved, got: %q", rules[0].Pattern)
	}
}

func TestParseCountryRuleText_PatternWithPipe(t *testing.T) {
	text := `CN 中国 100 true 中国|China|CN`

	rules, errors := parseCountryRuleText(text)

	if len(errors) != 0 {
		t.Errorf("Expected no errors for pattern with pipe, got: %v", errors)
	}
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
	if rules[0].Pattern != "中国|China|CN" {
		t.Errorf("Expected pattern with pipe to be preserved, got: %q", rules[0].Pattern)
	}
}

// ========== 文本同步功能测试 ==========

func TestSyncCountryRulesFromText_NewRules(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	text := `CN 中国 100 true 中国|China|CN
US 美国 90 true 美国|USA?|US`

	added, updated, deleted, err := SyncCountryRulesFromText(text)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if added != 2 {
		t.Errorf("Expected 2 added, got %d", added)
	}
	if updated != 0 {
		t.Errorf("Expected 0 updated, got %d", updated)
	}
	if deleted != 0 {
		t.Errorf("Expected 0 deleted, got %d", deleted)
	}

	// 验证规则已添加到数据库
	var rules []CountryRule
	database.DB.Find(&rules)
	if len(rules) != 2 {
		t.Errorf("Expected 2 rules in database, got %d", len(rules))
	}
}

func TestSyncCountryRulesFromText_UpdateExisting(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// 创建初始规则
	createTestRule(t, "CN", "中国", "中国|CN", 100)

	// 同步更新后的规则
	text := `CN 中国 100 true 中国|China|CN|🇨🇳`

	added, updated, deleted, err := SyncCountryRulesFromText(text)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if added != 0 {
		t.Errorf("Expected 0 added, got %d", added)
	}
	if updated != 1 {
		t.Errorf("Expected 1 updated, got %d", updated)
	}
	if deleted != 0 {
		t.Errorf("Expected 0 deleted, got %d", deleted)
	}

	// 验证规则已更新
	var rule CountryRule
	database.DB.Where("country_code = ?", "CN").First(&rule)
	if rule.Pattern != "中国|China|CN|🇨🇳" {
		t.Errorf("Expected pattern to be updated, got: %q", rule.Pattern)
	}
}

func TestSyncCountryRulesFromText_DeleteMissing(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// 创建初始规则
	createTestRule(t, "CN", "中国", "中国|CN", 100)
	createTestRule(t, "US", "美国", "美国|US", 90)

	// 同步只包含 CN 的文本（US 应该被删除）
	text := `CN 中国 100 true 中国|CN`

	added, updated, deleted, err := SyncCountryRulesFromText(text)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if added != 0 {
		t.Errorf("Expected 0 added, got %d", added)
	}
	if updated != 0 {
		t.Errorf("Expected 0 updated, got %d", updated)
	}
	if deleted != 1 {
		t.Errorf("Expected 1 deleted, got %d", deleted)
	}

	// 验证 US 规则已删除
	var count int64
	database.DB.Model(&CountryRule{}).Where("country_code = ?", "US").Count(&count)
	if count != 0 {
		t.Error("Expected US rule to be deleted")
	}
}

func TestSyncCountryRulesFromText_MixedOperations(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// 创建初始规则
	createTestRule(t, "CN", "中国", "中国|CN", 100)
	createTestRule(t, "US", "美国", "美国|US", 90)

	// 同步：保留 CN，更新 US，新增 JP，删除没有的规则
	text := `CN 中国 100 true 中国|CN
US 美国 95 true 美国|USA?|US
JP 日本 80 true 日本|Japan|JP`

	added, updated, deleted, err := SyncCountryRulesFromText(text)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if added != 1 {
		t.Errorf("Expected 1 added (JP), got %d", added)
	}
	if updated != 1 {
		t.Errorf("Expected 1 updated (US), got %d", updated)
	}
	if deleted != 0 {
		t.Errorf("Expected 0 deleted, got %d", deleted)
	}

	// 验证结果
	var rules []CountryRule
	database.DB.Find(&rules)
	if len(rules) != 3 {
		t.Errorf("Expected 3 rules in database, got %d", len(rules))
	}
}

func TestSyncCountryRulesFromText_ParseError(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// 包含解析错误的文本
	text := `CN 中国 abc true 中国|CN`

	_, _, _, err := SyncCountryRulesFromText(text)

	if err == nil {
		t.Error("Expected error for invalid text")
	}
	if !strings.Contains(err.Error(), "解析错误") {
		t.Errorf("Expected parsing error message, got: %v", err)
	}

	// 验证数据库没有被修改
	var count int64
	database.DB.Model(&CountryRule{}).Count(&count)
	if count != 0 {
		t.Error("Expected no rules to be added on parse error")
	}
}

func TestSyncCountryRulesFromText_EmptyText(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// 创建初始规则
	createTestRule(t, "CN", "中国", "中国|CN", 100)

	// 同步空文本（所有规则应该被删除）
	text := `# 空配置
# 所有规则将被删除`

	added, updated, deleted, err := SyncCountryRulesFromText(text)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if added != 0 {
		t.Errorf("Expected 0 added, got %d", added)
	}
	if updated != 0 {
		t.Errorf("Expected 0 updated, got %d", updated)
	}
	if deleted != 1 {
		t.Errorf("Expected 1 deleted, got %d", deleted)
	}

	// 验证所有规则已删除
	var count int64
	database.DB.Model(&CountryRule{}).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 rules in database, got %d", count)
	}
}
