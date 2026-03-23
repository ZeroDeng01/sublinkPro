package api

import (
	"encoding/json"
	"testing"

	"sublink/config"
	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupUserAPITestDB(t *testing.T) {
	t.Helper()
	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized
	oldCfg := *config.Get()

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "user_ai_api_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.MFALoginChallenge{}); err != nil {
		t.Fatalf("auto migrate users: %v", err)
	}
	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	config.UpdateConfig(func(cfg *config.AppConfig) {
		cfg.APIEncryptionKey = "test-api-encryption-key-0123456789abcd"
		cfg.JwtSecret = "test-jwt-secret"
	})
	if err := models.InitUserCache(); err != nil {
		t.Fatalf("init user cache: %v", err)
	}
	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		config.UpdateConfig(func(cfg *config.AppConfig) { *cfg = oldCfg })
		if oldDB != nil {
			_ = models.InitUserCache()
		}
		testutil.CloseDB(t, db)
	})
}

func TestUserGetAISettingsReturnsMaskedKey(t *testing.T) {
	setupUserAPITestDB(t)
	user := &models.User{Username: "admin", Password: "123456", Role: "admin", Nickname: "管理员"}
	if err := user.Create(); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := user.UpdateAISettings(models.UserAISettings{Enabled: true, BaseURL: "https://api.openai.com/v1", Model: "gpt-4.1-mini", RawAPIKey: "sk-test-1234567890", Temperature: 0.2, MaxTokens: 512}); err != nil {
		t.Fatalf("update ai settings: %v", err)
	}
	recorder := performJSONRequest(t, func(c *gin.Context) {
		c.Set("username", "admin")
		UserGetAISettings(c)
	}, "GET", nil)
	response := decodeAPIResponse(t, recorder)
	if response.Code != 200 {
		t.Fatalf("expected response code 200, got %d", response.Code)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatalf("unmarshal response data: %v", err)
	}
	if data["maskedKey"] == "sk-test-1234567890" {
		t.Fatal("expected API key to be masked")
	}
}
