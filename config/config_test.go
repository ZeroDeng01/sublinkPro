package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestApplyDefaultsSetsTrustedProxies(t *testing.T) {
	cfg := &AppConfig{}
	applyDefaults(cfg)

	if !reflect.DeepEqual(cfg.TrustedProxies, DefaultTrustedProxies) {
		t.Fatalf("unexpected default trusted proxies: got %v want %v", cfg.TrustedProxies, DefaultTrustedProxies)
	}

	if len(cfg.TrustedProxies) > 0 && &cfg.TrustedProxies[0] == &DefaultTrustedProxies[0] {
		t.Fatal("default trusted proxies should be copied, not reused")
	}
}

func TestLoadFromEnvInternalTrustedProxies(t *testing.T) {
	t.Setenv(envPrefix+"TRUSTED_PROXIES", "127.0.0.1, ::1, 10.0.0.0/8, ,127.0.0.1")

	cfg := &AppConfig{}
	applyDefaults(cfg)
	loadFromEnvInternal(cfg)

	want := []string{"127.0.0.1", "::1", "10.0.0.0/8"}
	if !reflect.DeepEqual(cfg.TrustedProxies, want) {
		t.Fatalf("unexpected trusted proxies from env: got %v want %v", cfg.TrustedProxies, want)
	}
}

func TestLoadFromFileInternalTrustedProxiesCanDisable(t *testing.T) {
	cfg := &AppConfig{}
	applyDefaults(cfg)

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("trusted_proxies: []\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loadFromFileInternal(cfg, path)

	if cfg.TrustedProxies == nil {
		t.Fatal("trusted proxies should remain an explicit empty slice when disabled in file")
	}
	if len(cfg.TrustedProxies) != 0 {
		t.Fatalf("expected trusted proxies to be disabled, got %v", cfg.TrustedProxies)
	}
}
