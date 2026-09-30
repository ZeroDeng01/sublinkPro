package services

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	backupservice "sublink/services/backup"
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
