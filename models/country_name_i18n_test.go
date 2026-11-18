package models

import (
	"testing"

	"sublink/utils"
)

func TestGetLocalizedCountryNameRendersDefaultsPerLanguage(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	// Default seeded style rule: Chinese country name with an English keyword set.
	createTestRule(t, "HK", "香港", "(?i)香港|hong kong|hk", 0)

	if got := GetLocalizedCountryName("HK", utils.NodeNameLanguageEnglish); got != "Hong Kong" {
		t.Errorf("english default name = %q, want %q", got, "Hong Kong")
	}
	if got := GetLocalizedCountryName("HK", utils.NodeNameLanguageChinese); got != "香港" {
		t.Errorf("chinese default name = %q, want %q", got, "香港")
	}
}

func TestGetLocalizedCountryNameKeepsCustomNames(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	createTestRule(t, "HK", "HK Premium", "(?i)香港|hong kong|hk", 0)

	if got := GetLocalizedCountryName("HK", utils.NodeNameLanguageEnglish); got != "HK Premium" {
		t.Errorf("custom name should be preserved, got %q", got)
	}
	if got := GetLocalizedCountryName("HK", utils.NodeNameLanguageChinese); got != "HK Premium" {
		t.Errorf("custom name should be preserved, got %q", got)
	}
}

func TestGetLocalizedCountryNameUnknownCode(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)

	if got := GetLocalizedCountryName("ZZ", utils.NodeNameLanguageEnglish); got != "" {
		t.Errorf("unknown code should resolve to empty, got %q", got)
	}
}
