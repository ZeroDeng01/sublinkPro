package models

import (
	"testing"

	regexp "github.com/dlclark/regexp2/v2/compat"
)

// TestRegexp2Lookbehind 验证 fix 分支引入的 regexp2/compat (PCRE) 引擎行为。
// 对照 sublinkpro 生产 v1.2.18 (标准库 regexp / RE2)：
//   - (?<![A-Za-z])HK(?![A-Za-z]) 在 RE2 编译失败：invalid named capture
//   - (?i)香港|HK|Hong Kong 对 "SHK Premium" 误匹配（裸 HK 子模式）
//
// 该测试需要 fix 分支的 models/country_rule.go 已把引擎换为 compat。
func TestRegexp2Lookbehind(t *testing.T) {
	const pattern = `(?<![A-Za-z])HK(?![A-Za-z])`

	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("compile %q: %v (RE2 会失败，compat 应成功)", pattern, err)
	}

	cases := []struct {
		name   string
		expect bool
		why    string
	}{
		{"HK", true, "纯 HK — 应匹配"},
		{"HK01", true, "右侧数字不算字母 — 应匹配"},
		{"HK Premium 01", true, "中间位置两侧非字母 — 应匹配"},
		{"SHK Premium", false, "左侧 S 是字母 — 不应匹配（issue #270 误匹配场景）"},
		{"HKG", false, "右侧 G 是字母 — 不应匹配"},
		{"The HK node", true, "空格两侧非字母 — 应匹配"},
		{"香港 01", false, "不含 HK — 不应匹配"},
	}
	for _, c := range cases {
		matched := re.MatchString(c.name)
		if matched != c.expect {
			t.Errorf("pattern=%q name=%q: got %v want %v (%s)", pattern, c.name, matched, c.expect, c.why)
		}
	}
}

// TestRegexp2InlineFlags 兼容性：旧 (?i) 内联标志与旧裸子模式 format 必须仍可用。
// 同时确认旧 bug 行为仍存在（这才是需要零宽断言的原因）：(?i)香港|HK|Hong Kong 仍会
// 误匹配 "SHK Premium"，因为 HK 是裸子模式。
func TestRegexp2InlineFlags(t *testing.T) {
	oldPatterns := []struct {
		pattern      string
		mSHK         bool
		mHongKongSrv bool
		mHKNode      bool
	}{
		{`(?i)香港|HK|Hong Kong`, true, true, true},
		{`(?i)(香港|HK|Hong\s*Kong)`, true, true, true},
		{`(?i)(美国|USA?|United\s*States)`, false, false, false},
		{`香港|HK|Hong Kong|🇭🇰`, true, true, true},
	}
	for _, p := range oldPatterns {
		re, err := regexp.Compile(p.pattern)
		if err != nil {
			t.Errorf("compile %q: %v (旧格式必须仍能编译)", p.pattern, err)
			continue
		}
		mSHK := re.MatchString("SHK Premium")
		mHK := re.MatchString("Hong Kong Server")
		mNode := re.MatchString("HK-001")
		if mSHK != p.mSHK {
			t.Errorf("pattern=%q vs SHK Premium: got %v want %v", p.pattern, mSHK, p.mSHK)
		}
		if mHK != p.mHongKongSrv {
			t.Errorf("pattern=%q vs Hong Kong Server: got %v want %v", p.pattern, mHK, p.mHongKongSrv)
		}
		if mNode != p.mHKNode {
			t.Errorf("pattern=%q vs HK-001: got %v want %v", p.pattern, mNode, p.mHKNode)
		}
	}
}
