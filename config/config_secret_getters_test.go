package config

import (
	"testing"
)

func TestGetAPIEncryptionKeyReturnsGlobalConfigValue(t *testing.T) {
	oldCfg := globalConfig
	oldGetter := secretGetterFunc
	oldSetter := secretSetterFunc
	t.Cleanup(func() {
		globalConfig = oldCfg
		secretGetterFunc = oldGetter
		secretSetterFunc = oldSetter
	})

	t.Setenv(envPrefix+"API_ENCRYPTION_KEY", "")
	globalConfig = &AppConfig{APIEncryptionKey: "global-key-0123456789abcdef0123456789"}
	secretGetterFunc = func(key string) string {
		if key != "api_encryption_key" {
			return ""
		}
		return "db-key-0123456789abcdef0123456789"
	}

	if got := GetAPIEncryptionKey(); got != "global-key-0123456789abcdef0123456789" {
		t.Fatalf("expected global config key, got %q", got)
	}
}

func TestGetAPIEncryptionKeyFallsBackToEnv(t *testing.T) {
	oldCfg := globalConfig
	oldGetter := secretGetterFunc
	oldSetter := secretSetterFunc
	t.Cleanup(func() {
		globalConfig = oldCfg
		secretGetterFunc = oldGetter
		secretSetterFunc = oldSetter
	})

	globalConfig = &AppConfig{}
	t.Setenv(envPrefix+"API_ENCRYPTION_KEY", "env-key-0123456789abcdef0123456789")
	secretGetterFunc = func(key string) string {
		if key != "api_encryption_key" {
			return ""
		}
		return "db-key-0123456789abcdef0123456789"
	}

	if got := GetAPIEncryptionKey(); got != "env-key-0123456789abcdef0123456789" {
		t.Fatalf("expected env key, got %q", got)
	}
}

func TestGetAPIEncryptionKeyFallsBackToSecretGetter(t *testing.T) {
	oldCfg := globalConfig
	oldGetter := secretGetterFunc
	oldSetter := secretSetterFunc
	t.Cleanup(func() {
		globalConfig = oldCfg
		secretGetterFunc = oldGetter
		secretSetterFunc = oldSetter
	})

	globalConfig = &AppConfig{}
	t.Setenv(envPrefix+"API_ENCRYPTION_KEY", "")
	secretGetterFunc = func(key string) string {
		if key != "api_encryption_key" {
			return ""
		}
		return "db-key-0123456789abcdef0123456789"
	}

	if got := GetAPIEncryptionKey(); got != "db-key-0123456789abcdef0123456789" {
		t.Fatalf("expected DB-backed key, got %q", got)
	}
}

func TestGetAPIEncryptionKeyPrefersEnvOverSecretGetter(t *testing.T) {
	oldCfg := globalConfig
	oldGetter := secretGetterFunc
	oldSetter := secretSetterFunc
	t.Cleanup(func() {
		globalConfig = oldCfg
		secretGetterFunc = oldGetter
		secretSetterFunc = oldSetter
	})

	globalConfig = &AppConfig{}
	t.Setenv(envPrefix+"API_ENCRYPTION_KEY", "env-key-0123456789abcdef0123456789")
	secretGetterFunc = func(key string) string {
		if key != "api_encryption_key" {
			return ""
		}
		return "db-key-0123456789abcdef0123456789"
	}

	if got := GetAPIEncryptionKey(); got != "env-key-0123456789abcdef0123456789" {
		t.Fatalf("expected env key to win over DB fallback, got %q", got)
	}
}
