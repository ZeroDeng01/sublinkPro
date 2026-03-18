package models

import (
	"os"
	"path/filepath"
	"testing"

	"sublink/cache"
	"sublink/database"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func resetTemplateCacheForTest() {
	templateCache = cache.NewMapCache(func(t Template) int { return t.ID })
	templateCache.AddIndex("name", func(t Template) string { return t.Name })
}

func setupTemplateTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&Template{}); err != nil {
		t.Fatalf("auto migrate templates: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	resetTemplateCacheForTest()

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		resetTemplateCacheForTest()
	})
}

func TestInferTemplateCategory(t *testing.T) {
	tests := map[string]string{
		"clash.yaml": "clash",
		"surge.conf": "surge",
		"SURGE.CONF": "surge",
		"rules.txt":  "clash",
	}

	for fileName, want := range tests {
		if got := InferTemplateCategory(fileName); got != want {
			t.Fatalf("InferTemplateCategory(%q)=%q, want %q", fileName, got, want)
		}
	}
}

func TestMigrateTemplatesFromFilesCreatesExpectedCategories(t *testing.T) {
	setupTemplateTestDB(t)

	templateDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(templateDir, "clash.yaml"), []byte("proxies: []\n"), 0600); err != nil {
		t.Fatalf("write clash template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "surge.conf"), []byte("[General]\n"), 0600); err != nil {
		t.Fatalf("write surge template: %v", err)
	}

	if err := MigrateTemplatesFromFiles(templateDir); err != nil {
		t.Fatalf("migrate templates: %v", err)
	}

	var templates []Template
	if err := database.DB.Order("name asc").Find(&templates).Error; err != nil {
		t.Fatalf("query templates: %v", err)
	}
	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}

	got := map[string]string{}
	for _, tmpl := range templates {
		got[tmpl.Name] = tmpl.Category
	}
	if got["clash.yaml"] != "clash" {
		t.Fatalf("expected clash.yaml category clash, got %q", got["clash.yaml"])
	}
	if got["surge.conf"] != "surge" {
		t.Fatalf("expected surge.conf category surge, got %q", got["surge.conf"])
	}
}

func TestMigrateTemplatesFromFilesRepairsInvalidCategory(t *testing.T) {
	setupTemplateTestDB(t)

	templateDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(templateDir, "surge.conf"), []byte("[General]\n"), 0600); err != nil {
		t.Fatalf("write surge template: %v", err)
	}

	existing := Template{
		Name:     "surge.conf",
		Category: "unknown",
	}
	if err := database.DB.Create(&existing).Error; err != nil {
		t.Fatalf("create existing template: %v", err)
	}

	if err := MigrateTemplatesFromFiles(templateDir); err != nil {
		t.Fatalf("migrate templates: %v", err)
	}

	var repaired Template
	if err := database.DB.Where("name = ?", "surge.conf").First(&repaired).Error; err != nil {
		t.Fatalf("query repaired template: %v", err)
	}
	if repaired.Category != "surge" {
		t.Fatalf("expected repaired category surge, got %q", repaired.Category)
	}
}
