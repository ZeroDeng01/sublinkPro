package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"sublink/cache"
	"sublink/database"
	"sublink/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTemplateAPITestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized
	oldBaseTemplateDir := baseTemplateDir

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.Template{}); err != nil {
		t.Fatalf("auto migrate templates: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	if err := models.InitTemplateCache(); err != nil {
		t.Fatalf("init template cache: %v", err)
	}
	cache.InvalidateAllTemplateContent()

	t.Cleanup(func() {
		cache.InvalidateAllTemplateContent()
		baseTemplateDir = oldBaseTemplateDir
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		if oldDB != nil {
			_ = models.InitTemplateCache()
		}
	})
}

func TestGetTempSInfersSurgeCategoryWithoutMetadata(t *testing.T) {
	setupTemplateAPITestDB(t)

	templateDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(templateDir, "surge.conf"), []byte("[General]\n"), 0600); err != nil {
		t.Fatalf("write surge template: %v", err)
	}
	baseTemplateDir = templateDir

	recorder := performJSONRequest(t, GetTempS, http.MethodGet, nil)
	response := decodeAPIResponse(t, recorder)
	if response.Code != 200 {
		t.Fatalf("expected response code 200, got %d", response.Code)
	}

	var templates []Temp
	if err := json.Unmarshal(response.Data, &templates); err != nil {
		t.Fatalf("unmarshal template list: %v", err)
	}
	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}
	if templates[0].File != "surge.conf" {
		t.Fatalf("expected surge.conf, got %q", templates[0].File)
	}
	if templates[0].Category != "surge" {
		t.Fatalf("expected inferred category surge, got %q", templates[0].Category)
	}
}
