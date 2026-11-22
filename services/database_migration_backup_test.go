package services

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
	backupservice "sublink/services/backup"
	socks5service "sublink/services/socks5"

	"gorm.io/gorm"
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

func TestDatabaseMigrationRestoresShareDeviceState(t *testing.T) {
	oldDialect := database.Dialect
	database.Dialect = database.DialectSQLite
	t.Cleanup(func() { database.Dialect = oldDialect })
	for _, format := range []string{"sqlite", "backup_zip"} {
		for _, scenario := range []string{"empty-target", "overwrite-target", "legacy-no-devices", "legacy-no-shares", "missing-shares-with-devices"} {
			t.Run(format+"/"+scenario, func(t *testing.T) {
				sourcePath := filepath.Join(t.TempDir(), "source.db")
				source, err := database.OpenMigrationSourceSQLite(sourcePath)
				if err != nil {
					t.Fatal(err)
				}
				if err := source.AutoMigrate(&models.Subcription{}); err != nil {
					t.Fatal(err)
				}
				if err := insertRecords(source, []models.Subcription{{ID: 3, Name: "restored-subscription"}}); err != nil {
					t.Fatal(err)
				}
				hasShares := scenario != "legacy-no-shares" && scenario != "missing-shares-with-devices"
				hasDevices := scenario == "empty-target" || scenario == "overwrite-target"
				if hasShares {
					if err := source.AutoMigrate(&models.SubscriptionShare{}); err != nil {
						t.Fatal(err)
					}
					limit := 0
					if hasDevices {
						limit = 2
					}
					if err := insertRecords(source, []models.SubscriptionShare{{ID: 7, SubscriptionID: 3, Token: "restored-token", Name: "restored-share", Enabled: true, MaxDevices: limit, DeviceRevision: 12}}); err != nil {
						t.Fatal(err)
					}
				}
				created := time.Date(2026, 10, 1, 2, 3, 4, 0, time.UTC)
				lastAccess := created.Add(time.Hour)
				want := []models.ShareDevice{
					{ID: 11, ShareID: 7, HWIDHash: strings.Repeat("a", 64), Name: "phone", OS: "Android", Model: "Pixel", CreatedAt: created, LastAccessAt: lastAccess},
					{ID: 12, ShareID: 7, HWIDHash: strings.Repeat("b", 64), Name: "revoked-laptop", OS: "Windows", Model: "PC", Revoked: true, CreatedAt: created, LastAccessAt: lastAccess},
				}
				if hasDevices {
					if err := source.AutoMigrate(&models.ShareDevice{}); err != nil {
						t.Fatal(err)
					}
					seed := append(append([]models.ShareDevice{}, want...), models.ShareDevice{ID: 13, ShareID: 999, HWIDHash: strings.Repeat("c", 64)})
					if err := insertRecords(source, seed); err != nil {
						t.Fatal(err)
					}
				} else if scenario == "missing-shares-with-devices" {
					if err := source.AutoMigrate(&models.ShareDevice{}); err != nil {
						t.Fatal(err)
					}
					// The target's next ID after clearing share 7 is 8. An orphan
					// source binding must not attach to that generated legacy share.
					if err := insertRecords(source, []models.ShareDevice{{ID: 11, ShareID: 8, HWIDHash: strings.Repeat("c", 64)}}); err != nil {
						t.Fatal(err)
					}
				}
				testutil.CloseDB(t, source)
				source = openShareRestoreSource(t, sourcePath, format)
				target := testutil.OpenMemoryDB(t, "share_restore_target")
				t.Cleanup(func() { testutil.CloseDB(t, source); testutil.CloseDB(t, target) })
				if err := target.AutoMigrate(&models.Subcription{}, &models.SubscriptionShare{}, &models.ShareDevice{}); err != nil {
					t.Fatal(err)
				}
				if scenario != "empty-target" {
					if err := insertRecords(target, []models.SubscriptionShare{{ID: 7, SubscriptionID: 3, Token: "stale-token", Enabled: true}}); err != nil {
						t.Fatal(err)
					}
					if err := insertRecords(target, []models.ShareDevice{{ID: 11, ShareID: 7, HWIDHash: strings.Repeat("d", 64), Name: "stale-target-device"}}); err != nil {
						t.Fatal(err)
					}
				}
				state := &databaseMigrationState{
					source: source, sourceHasShares: hasShares,
					importedSubIDs: make(map[int]struct{}), importedShareIDs: make(map[int]struct{}),
					result: &DatabaseMigrationResult{Imported: make(map[string]int)},
				}
				if err := target.Transaction(func(tx *gorm.DB) error {
					state.tx = tx
					if err := clearTargetBusinessData(tx); err != nil {
						return err
					}
					if err := importSubscriptions(state); err != nil {
						return err
					}
					return importSubscriptionShares(state)
				}); err != nil {
					t.Fatal(err)
				}
				var actual []models.ShareDevice
				if err := target.Order("id").Find(&actual).Error; err != nil {
					t.Fatal(err)
				}
				if !hasDevices {
					if len(actual) != 0 || state.result.Imported["share_devices"] != 0 {
						t.Fatalf("old backup retained target devices: %+v", actual)
					}
				} else {
					if !reflect.DeepEqual(actual, want) {
						t.Fatalf("restored device state = %+v, want %+v", actual, want)
					}
					if state.result.Imported["share_devices"] != 2 || len(state.result.Warnings) != 1 {
						t.Fatalf("unexpected device import result: %+v", state.result)
					}
				}
				var shares []models.SubscriptionShare
				if err := target.Find(&shares).Error; err != nil {
					t.Fatal(err)
				}
				if len(shares) != 1 || shares[0].Token == "stale-token" {
					t.Fatalf("restored shares = %+v", shares)
				}
				if hasShares && (shares[0].ID != 7 || shares[0].Token != "restored-token" || shares[0].DeviceRevision != 12) {
					t.Fatalf("share policy or token changed during restore: %+v", shares[0])
				}
				if hasDevices {
					if err := target.Create(&models.ShareDevice{ShareID: 7, HWIDHash: strings.Repeat("e", 64)}).Error; err != nil {
						t.Fatalf("register device after restoring explicit IDs: %v", err)
					}
				}
			})
		}
	}
}

// WebDAV downloads use the same ZIP extraction and import path as uploaded backups.
func openShareRestoreSource(t *testing.T, sourcePath, format string) *gorm.DB {
	t.Helper()
	if format == "backup_zip" {
		zipPath := filepath.Join(t.TempDir(), "backup.zip")
		archive, err := os.Create(zipPath)
		if err != nil {
			t.Fatal(err)
		}
		writer := zip.NewWriter(archive)
		entry, err := writer.Create("db/sublink.db")
		if err != nil {
			t.Fatal(err)
		}
		source, err := os.Open(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		_, copyErr := io.Copy(entry, source)
		closeErr := source.Close()
		if copyErr != nil || closeErr != nil {
			t.Fatalf("copy SQLite backup: %v, close: %v", copyErr, closeErr)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		if err := archive.Close(); err != nil {
			t.Fatal(err)
		}
		extracted, err := extractMigrationZipTo(zipPath, t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		sourcePath, err = findSQLiteDatabaseFile(filepath.Join(extracted, "db"))
		if err != nil {
			t.Fatal(err)
		}
	}
	db, err := database.OpenMigrationSourceSQLite(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	return db
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
