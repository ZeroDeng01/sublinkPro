package models

import "testing"

func TestEncryptUserAISecretUsesVersionedAEADFormat(t *testing.T) {
	setupUserMFATestDB(t)

	encrypted, err := EncryptUserAISecret("sk-test-1234567890")
	if err != nil {
		t.Fatalf("encrypt ai secret: %v", err)
	}
	if encrypted == "" || encrypted[:3] != "v1:" {
		t.Fatalf("expected versioned encrypted payload, got %s", encrypted)
	}
	decrypted, err := DecryptUserAISecret(encrypted)
	if err != nil {
		t.Fatalf("decrypt ai secret: %v", err)
	}
	if decrypted != "sk-test-1234567890" {
		t.Fatalf("unexpected decrypted value: %s", decrypted)
	}
}

func TestUserGetAISettingsMasksStoredKey(t *testing.T) {
	setupUserMFATestDB(t)

	user := &User{Username: "ai-user", Password: "123456", Role: "admin", Nickname: "AI"}
	if err := user.Create(); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := user.UpdateAISettings(UserAISettings{Enabled: true, BaseURL: "https://api.example.com/v1", Model: "gpt-test", RawAPIKey: "sk-test-1234567890", Temperature: 0.2, MaxTokens: 500}); err != nil {
		t.Fatalf("update ai settings: %v", err)
	}
	settings, err := user.GetAISettings()
	if err != nil {
		t.Fatalf("get ai settings: %v", err)
	}
	if !settings.HasKey {
		t.Fatal("expected hasKey to be true")
	}
	if settings.MaskedKey == "" || settings.MaskedKey == "sk-test-1234567890" {
		t.Fatalf("expected masked key, got %q", settings.MaskedKey)
	}
	if settings.RawAPIKey != "sk-test-1234567890" {
		t.Fatalf("expected decrypted raw key, got %q", settings.RawAPIKey)
	}
}
