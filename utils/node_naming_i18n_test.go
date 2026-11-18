package utils

import "testing"

func TestNormalizeNodeNameLanguage(t *testing.T) {
	cases := map[string]string{
		"":        NodeNameLanguageEnglish,
		"en":      NodeNameLanguageEnglish,
		"en-GB":   NodeNameLanguageEnglish,
		"zh":      NodeNameLanguageChinese,
		"zh-CN":   NodeNameLanguageChinese,
		"zh_cn":   NodeNameLanguageChinese,
		"zh-Hans": NodeNameLanguageChinese,
		"ru-RU":   NodeNameLanguageEnglish,
		"zh-Hant": NodeNameLanguageEnglish,
	}
	for input, want := range cases {
		if got := NormalizeNodeNameLanguage(input); got != want {
			t.Errorf("NormalizeNodeNameLanguage(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNodeNameTextFallbacks(t *testing.T) {
	if got := NodeNameText(NodeNameLanguageEnglish, NodeNameTextUngrouped); got != "Ungrouped" {
		t.Errorf("english ungrouped = %q", got)
	}
	if got := NodeNameText(NodeNameLanguageChinese, NodeNameTextUngrouped); got != "未分组" {
		t.Errorf("chinese ungrouped = %q", got)
	}
	if got := NodeNameText("ru-RU", NodeNameTextResidential); got != "Residential" {
		t.Errorf("unknown language should fall back to english, got %q", got)
	}
	if got := NodeNameText(NodeNameLanguageEnglish, "missing-key"); got != "" {
		t.Errorf("missing key should be empty, got %q", got)
	}
}

func TestUnlockStatusText(t *testing.T) {
	cases := map[string]string{
		"available":   NodeNameTextUnlockUnlocked,
		"PARTIAL":     NodeNameTextUnlockPartial,
		"reachable":   NodeNameTextUnlockReachable,
		"restricted":  NodeNameTextUnlockRestricted,
		"unsupported": NodeNameTextUnlockUnsupported,
		"error":       NodeNameTextUnlockError,
		"unknown":     NodeNameTextUnlockUnknown,
		"":            NodeNameTextUnlockUntested,
	}
	for status, wantKey := range cases {
		if got := UnlockStatusTextKey(status); got != wantKey {
			t.Errorf("UnlockStatusTextKey(%q) = %q, want %q", status, got, wantKey)
		}
	}
	if got := UnlockStatusText(NodeNameLanguageChinese, "available"); got != "解锁" {
		t.Errorf("chinese available = %q", got)
	}
}

func TestRenameNodeWithLanguageRendersLocalizedDefaults(t *testing.T) {
	info := NodeInfo{Name: "Node 01", LinkName: "Node 01", Source: "manual"}

	english := RenameNodeWithLanguage("$Group|$Source|$LinkCountry", info, NodeNameLanguageEnglish)
	if english != "Ungrouped|Manual|Unknown" {
		t.Errorf("english defaults = %q", english)
	}

	chinese := RenameNodeWithLanguage("$Group|$Source|$LinkCountry", info, NodeNameLanguageChinese)
	if chinese != "未分组|手动|未知" {
		t.Errorf("chinese defaults = %q", chinese)
	}
}
