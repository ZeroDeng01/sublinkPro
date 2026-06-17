package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"

	"github.com/gin-gonic/gin"
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

	db := testutil.OpenMemoryDB(t, "setting_api_test")
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

func performJSONRequest(t *testing.T, handler gin.HandlerFunc, method string, body any) *httptest.ResponseRecorder {
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
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), method, "/", bytes.NewReader(requestBody))
	ginContext.Request.Header.Set("Content-Type", "application/json")

	handler(ginContext)
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
