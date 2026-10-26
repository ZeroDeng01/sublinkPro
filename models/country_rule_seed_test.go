package models

import (
	"testing"

	"sublink/database"
)

// extendedRuleByCode 从扩展规则清单中取出指定国家代码的规则
func extendedRuleByCode(t *testing.T, code string) CountryRule {
	t.Helper()
	for _, rule := range extendedCountryRules() {
		if rule.CountryCode == code {
			return rule
		}
	}
	t.Fatalf("未找到扩展国家规则 %s", code)
	return CountryRule{}
}

func TestExtendedCountryRulesMatchNodeNames(t *testing.T) {
	tests := []struct {
		code     string
		nodeName string
	}{
		{"BY", "BY-01"},
		{"BY", "🇧🇾 Minsk"},
		{"BY", "Беларусь 02"},
		{"TJ", "TJ Premium"},
		{"AQ", "Antarctica"},
		{"AL", "AL-01"},
		{"AL", "Албания"},
		{"MD", "MD-01"},
		{"MK", "Macedonia"},
		{"AM", "🇦🇲 Armenia"},
		{"AM", "Армения 01"},
		{"UZ", "UZ-01"},
		{"GE", "GE-01"},
		{"GE", "Грузия"},
		{"BA", "Bosnia"},
		{"IQ", "Ирак"},
		{"KH", "Камбоджа 01"},
		{"AZ", "AZ-01"},
		{"BH", "Bahrain"},
		{"MN", "Монголия"},
		{"CY", "Cyprus-01"},
		{"KG", "Киргизия"},
		{"BN", "Brunei"},
		{"LI", "Liechtenstein"},
		{"LK", "Sri Lanka 01"},
		{"SS", "South Sudan"},
		{"SS", "Южный Судан"},
	}

	for _, tt := range tests {
		rule := extendedRuleByCode(t, tt.code)
		matched, err := TestPattern(rule.Pattern, tt.nodeName)
		if err != nil {
			t.Fatalf("规则 %s 正则编译失败: %v", tt.code, err)
		}
		if !matched {
			t.Errorf("规则 %s 未匹配节点名 %q", tt.code, tt.nodeName)
		}
	}
}

func TestExtendedCountryRuleCodesRequireBoundaries(t *testing.T) {
	tests := []struct {
		code     string
		nodeName string
	}{
		{"AM", "Amsterdam"},
		{"GE", "Germany"},
		{"AL", "Almaty"},
		{"BY", "Standby"},
		{"BA", "Bangkok"},
		{"SS", "Shadowsocks"},
	}

	for _, tt := range tests {
		rule := extendedRuleByCode(t, tt.code)
		matched, err := TestPattern(rule.Pattern, tt.nodeName)
		if err != nil {
			t.Fatalf("规则 %s 正则编译失败: %v", tt.code, err)
		}
		if matched {
			t.Errorf("规则 %s 误匹配节点名 %q", tt.code, tt.nodeName)
		}
	}
}

func TestSeedExtendedCountryRulesIsIdempotent(t *testing.T) {
	setupTestDB(t)

	if err := seedExtendedCountryRules(database.DB); err != nil {
		t.Fatalf("首次补充国家规则失败: %v", err)
	}
	var count int64
	if err := database.DB.Model(&CountryRule{}).Count(&count).Error; err != nil {
		t.Fatalf("统计国家规则失败: %v", err)
	}
	if want := int64(len(extendedCountryRules())); count != want {
		t.Fatalf("国家规则数量 = %d, 期望 %d", count, want)
	}

	if err := seedExtendedCountryRules(database.DB); err != nil {
		t.Fatalf("重复补充国家规则失败: %v", err)
	}
	if err := database.DB.Model(&CountryRule{}).Count(&count).Error; err != nil {
		t.Fatalf("统计国家规则失败: %v", err)
	}
	if count != int64(len(extendedCountryRules())) {
		t.Fatalf("重复执行后国家规则数量 = %d, 期望 %d", count, len(extendedCountryRules()))
	}

	// 用户自定义的同代码规则不能被覆盖
	if err := database.DB.Model(&CountryRule{}).Where("country_code = ?", "BY").Update("pattern", "custom-by").Error; err != nil {
		t.Fatalf("更新自定义规则失败: %v", err)
	}
	if err := seedExtendedCountryRules(database.DB); err != nil {
		t.Fatalf("再次补充国家规则失败: %v", err)
	}
	var by CountryRule
	if err := database.DB.Where("country_code = ?", "BY").First(&by).Error; err != nil {
		t.Fatalf("查询 BY 规则失败: %v", err)
	}
	if by.Pattern != "custom-by" {
		t.Errorf("用户自定义规则被覆盖: %q", by.Pattern)
	}
}

func TestBackfillNodeCountries(t *testing.T) {
	setupTestDB(t)
	if err := database.DB.AutoMigrate(&Node{}); err != nil {
		t.Fatalf("迁移 nodes 表失败: %v", err)
	}

	nodes := []Node{
		{LinkHash: "hash-am", Name: "🇦🇲 Armenia 01", LinkCountry: ""},
		{LinkHash: "hash-cn", Name: "Китай 01", LinkCountry: ""},
		{LinkHash: "hash-us", Name: "US-01", LinkCountry: "US"},
		{LinkHash: "hash-unknown", Name: "Unknown Node", LinkCountry: ""},
	}
	for i := range nodes {
		if err := database.DB.Create(&nodes[i]).Error; err != nil {
			t.Fatalf("创建节点失败: %v", err)
		}
	}

	rules := append(extendedCountryRules(), CountryRule{CountryCode: "CN", CountryName: "中国", Pattern: chinaCountryRulePattern})
	filled, err := backfillNodeCountries(database.DB, rules)
	if err != nil {
		t.Fatalf("回填节点国家失败: %v", err)
	}
	if filled != 2 {
		t.Fatalf("回填数量 = %d, 期望 2", filled)
	}

	want := map[string]string{
		"🇦🇲 Armenia 01": "AM",
		"Китай 01":      "CN",
		"US-01":         "US",
		"Unknown Node":  "",
	}
	var got []Node
	if err := database.DB.Find(&got).Error; err != nil {
		t.Fatalf("查询节点失败: %v", err)
	}
	for _, node := range got {
		if node.LinkCountry != want[node.Name] {
			t.Errorf("节点 %q 国家 = %q, 期望 %q", node.Name, node.LinkCountry, want[node.Name])
		}
	}
}

func TestSeedExtendedCountryRulesUpgradesLegacyChinaRule(t *testing.T) {
	setupTestDB(t)

	legacy := CountryRule{
		CountryCode: "CN",
		CountryName: "中国",
		Pattern:     legacyChinaCountryRulePattern,
		Priority:    0,
		Enabled:     true,
	}
	if err := database.DB.Create(&legacy).Error; err != nil {
		t.Fatalf("创建旧版中国大陆规则失败: %v", err)
	}

	if err := seedExtendedCountryRules(database.DB); err != nil {
		t.Fatalf("补充国家规则失败: %v", err)
	}
	var cn CountryRule
	if err := database.DB.Where("country_code = ?", "CN").First(&cn).Error; err != nil {
		t.Fatalf("查询 CN 规则失败: %v", err)
	}
	if cn.Pattern != chinaCountryRulePattern {
		t.Errorf("中国大陆规则未升级: %q", cn.Pattern)
	}

	// 用户修改过的中国大陆规则不能被升级
	if err := database.DB.Model(&CountryRule{}).Where("country_code = ?", "CN").Update("pattern", "custom-cn").Error; err != nil {
		t.Fatalf("更新自定义规则失败: %v", err)
	}
	if err := seedExtendedCountryRules(database.DB); err != nil {
		t.Fatalf("再次补充国家规则失败: %v", err)
	}
	if err := database.DB.Where("country_code = ?", "CN").First(&cn).Error; err != nil {
		t.Fatalf("查询 CN 规则失败: %v", err)
	}
	if cn.Pattern != "custom-cn" {
		t.Errorf("用户自定义的中国大陆规则被覆盖: %q", cn.Pattern)
	}
}
