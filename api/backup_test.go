package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
	backupservice "sublink/services/backup"

	"github.com/gin-gonic/gin"
)

func setupBackupAPITestDB(t *testing.T) {
	t.Helper()
	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized
	db := testutil.OpenMemoryDB(t, "backup_api_test")
	if err := db.AutoMigrate(&models.User{}, &models.SystemSetting{}); err != nil {
		t.Fatalf("auto migrate backup tables: %v", err)
	}
	if err := db.Create(&models.User{Username: "admin", Role: "admin"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.User{Username: "member", Role: "user"}).Error; err != nil {
		t.Fatal(err)
	}
	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	if err := models.InitUserCache(); err != nil {
		t.Fatal(err)
	}
	if err := models.InitSettingCache(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		if oldDB != nil {
			_ = models.InitUserCache()
			_ = models.InitSettingCache()
		}
		testutil.CloseDB(t, db)
	})
}

func performBackupRequest(t *testing.T, username string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	ctx.Set("username", username)
	handler(ctx)
	return recorder
}

func TestGetWebDAVBackupSettingsRequiresAdmin(t *testing.T) {
	setupBackupAPITestDB(t)
	recorder := performBackupRequest(t, "member", GetWebDAVBackupSettings)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	var response apiJSONResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusForbidden {
		t.Fatalf("body code = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestGetWebDAVBackupSettingsDoesNotExposePassword(t *testing.T) {
	setupBackupAPITestDB(t)

	t.Setenv("SUBLINK_API_ENCRYPTION_KEY", "backup-test-key-0123456789abcdef0123456789")
	if err := backupservice.SaveConfig(backupservice.Config{
		BaseURL:        "https://dav.example.com",
		Username:       "admin",
		Password:       "top-secret-password",
		RemotePath:     "SublinkPro",
		TimeoutSeconds: 60,
	}); err != nil {
		t.Fatalf("save WebDAV config: %v", err)
	}
	recorder := performBackupRequest(t, "admin", GetWebDAVBackupSettings)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	if body == "" || !json.Valid([]byte(body)) || strings.Contains(body, "top-secret-password") || strings.Contains(body, `"password"`) {
		t.Fatalf("unexpected response body: %s", body)
	}
	if !strings.Contains(body, `"hasPassword":true`) {
		t.Fatalf("expected password metadata in response: %s", body)
	}
}

func TestUploadWebDAVBackupRejectsNonSQLiteDatabase(t *testing.T) {
	setupBackupAPITestDB(t)
	database.Dialect = database.DialectPostgres
	recorder := performBackupRequest(t, "admin", UploadWebDAVBackup)
	response := decodeAPIResponse(t, recorder)
	if response.Code != 500 {
		t.Fatalf("body code = %d, want 500", response.Code)
	}
	if !strings.Contains(recorder.Body.String(), "settings.backup.api.sqliteOnly") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestLocalBackupRequiresAdmin(t *testing.T) {
	setupBackupAPITestDB(t)
	recorder := performBackupRequest(t, "member", Backup)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}
