package services

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"sublink/internal/testutil"
	"sublink/models"
	backupservice "sublink/services/backup"
	socks5service "sublink/services/socks5"
)

func TestDatabaseMigrationPreservesWebDAVSettings(t *testing.T) {
	if !shouldPreserveTargetSetting("api_encryption_key") {
		t.Fatal("API encryption key should be preserved during restore")
	}
	for _, key := range backupservice.PreservedSettingKeys() {
		if !shouldPreserveTargetSetting(key) {
			t.Fatalf("WebDAV setting %q should be preserved during restore", key)
		}
	}
}

func TestExtractMigrationZipCleansDirectoryOnFailure(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "invalid.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("../escape.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("escape")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	tempRoot := t.TempDir()
	if _, err := extractMigrationZipTo(zipPath, tempRoot); err == nil {
		t.Fatal("expected invalid path error")
	}
	entries, err := os.ReadDir(tempRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("temporary extraction directory was not cleaned: %#v", entries)
	}
}

func TestDatabaseMigrationKeepsTargetSocks5Configuration(t *testing.T) {
	for _, keepTarget := range []bool{true, false} {
		name := "configured-target"
		if !keepTarget {
			name = "fresh-target"
		}
		t.Run(name, func(t *testing.T) {
			source := testutil.OpenMemoryDB(t, "socks5_migration_source")
			target := testutil.OpenMemoryDB(t, "socks5_migration_target")
			t.Cleanup(func() { testutil.CloseDB(t, source); testutil.CloseDB(t, target) })
			if err := source.AutoMigrate(&models.SystemSetting{}); err != nil {
				t.Fatal(err)
			}
			if err := target.AutoMigrate(&models.SystemSetting{}); err != nil {
				t.Fatal(err)
			}
			keys := append(socks5service.PreservedSettingKeys(), "api_encryption_key")
			for _, key := range keys {
				if err := source.Create(&models.SystemSetting{Key: key, Value: "source:" + key}).Error; err != nil {
					t.Fatal(err)
				}
				if keepTarget {
					if err := target.Create(&models.SystemSetting{Key: key, Value: "target:" + key}).Error; err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := source.Create(&models.SystemSetting{Key: "test_business_setting", Value: "import-me"}).Error; err != nil {
				t.Fatal(err)
			}
			preserved, err := loadPreservedTargetSettings(target)
			if err != nil {
				t.Fatal(err)
			}
			if err := target.Where("1 = 1").Delete(&models.SystemSetting{}).Error; err != nil {
				t.Fatal(err)
			}
			state := &databaseMigrationState{tx: target, source: source, preservedSettings: preserved, importedSettings: map[string]string{}, result: &DatabaseMigrationResult{Imported: map[string]int{}}}
			if err := importSystemSettings(state); err != nil {
				t.Fatal(err)
			}
			var imported []models.SystemSetting
			if err := target.Find(&imported).Error; err != nil {
				t.Fatal(err)
			}
			actual := map[string]string{}
			for _, setting := range imported {
				actual[setting.Key] = setting.Value
			}
			for _, key := range keys {
				if keepTarget {
					if actual[key] != "target:"+key {
						t.Errorf("target setting %s was replaced: %q", key, actual[key])
					}
				} else if _, exists := actual[key]; exists {
					t.Errorf("source gateway setting %s must not enable/configure a fresh target", key)
				}
			}
			if actual["test_business_setting"] != "import-me" {
				t.Fatal("ordinary business settings should still be imported")
			}
		})
	}
}
