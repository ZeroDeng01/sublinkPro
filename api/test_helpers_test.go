package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type apiJSONResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func setupSettingAPITestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "setting_api_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.SystemSetting{}); err != nil {
		t.Fatalf("auto migrate system_settings: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
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

func performJSONRequest(t *testing.T, handler gin.HandlerFunc, method string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var requestBody []byte
	var err error
	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(method, "/", bytes.NewReader(requestBody))
	context.Request.Header.Set("Content-Type", "application/json")

	handler(context)
	return recorder
}

func decodeAPIResponse(t *testing.T, recorder *httptest.ResponseRecorder) apiJSONResponse {
	t.Helper()

	var response apiJSONResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal api response: %v", err)
	}
	return response
}
