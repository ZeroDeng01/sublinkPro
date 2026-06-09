package substore

import (
	"reflect"
	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupSubStoreSettingsTestDB(t *testing.T) {
	t.Helper()
	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "substore_settings_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.SystemSetting{}); err != nil {
		t.Fatalf("auto migrate system settings: %v", err)
	}
	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = true
	if err := models.InitSettingCache(); err != nil {
		t.Fatalf("init setting cache: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Exec("DELETE FROM system_settings").Error
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		if oldDB != nil {
			_ = models.InitSettingCache()
		}
		testutil.CloseDB(t, db)
	})
}

func TestResolveEffectiveSettingsUsesDatabaseValues(t *testing.T) {
	setupSubStoreSettingsTestDB(t)

	if err := SaveSettings(Settings{
		Enabled:          true,
		BaseURL:          "http://substore:3000/",
		TimeoutSeconds:   7,
		AllowedTargets:   []string{"loon", "quanx", "unknown", "loon"},
		MaxResponseBytes: 2048,
	}); err != nil {
		t.Fatalf("save settings: %v", err)
	}

	effective := ResolveEffectiveSettings()
	if !effective.Configured {
		t.Fatal("expected database settings to configure Sub-Store")
	}
	if effective.BaseURL.Value != "http://substore:3000" || effective.BaseURL.Source != "database" {
		t.Fatalf("unexpected base URL field: %+v", effective.BaseURL)
	}
	if effective.TimeoutSeconds.Value != 7 {
		t.Fatalf("unexpected timeout: %+v", effective.TimeoutSeconds)
	}
	wantTargets := []string{"loon", "quanx"}
	if !reflect.DeepEqual(effective.AllowedTargets.Value, wantTargets) {
		t.Fatalf("unexpected targets: got %v want %v", effective.AllowedTargets.Value, wantTargets)
	}
}

func TestEffectiveConfigRequiresEnabledSettings(t *testing.T) {
	setupSubStoreSettingsTestDB(t)

	if err := SaveSettings(Settings{
		Enabled:          false,
		BaseURL:          "http://substore:3000",
		TimeoutSeconds:   7,
		AllowedTargets:   []string{"loon"},
		MaxResponseBytes: 2048,
	}); err != nil {
		t.Fatalf("save settings: %v", err)
	}

	cfg, effective := EffectiveConfig()
	if effective.Configured {
		t.Fatal("expected disabled settings to leave Sub-Store unconfigured")
	}
	if cfg.BaseURL != "" {
		t.Fatalf("expected empty effective config when disabled, got %+v", cfg)
	}
}

func TestResolveEffectiveSettingsKeepsEmptyAllowedTargets(t *testing.T) {
	setupSubStoreSettingsTestDB(t)
	if err := SaveSettings(Settings{
		Enabled:          true,
		BaseURL:          "http://database-substore:3000",
		TimeoutSeconds:   7,
		AllowedTargets:   nil,
		MaxResponseBytes: 2048,
	}); err != nil {
		t.Fatalf("save settings: %v", err)
	}

	effective := ResolveEffectiveSettings()
	if effective.BaseURL.Value != "http://database-substore:3000" || effective.BaseURL.Source != "database" {
		t.Fatalf("expected database base URL, got %+v", effective.BaseURL)
	}
	if len(effective.AllowedTargets.Value) != 0 {
		t.Fatalf("expected no expanded targets when allowlist is empty, got %v", effective.AllowedTargets.Value)
	}
}
