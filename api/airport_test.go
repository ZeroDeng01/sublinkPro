package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupAirportAPITestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "airport_api_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.Airport{}); err != nil {
		t.Fatalf("auto migrate airports: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	if err := models.InitAirportCache(); err != nil {
		t.Fatalf("init airport cache: %v", err)
	}

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		if oldDB != nil {
			_ = models.InitAirportCache()
		}
		testutil.CloseDB(t, db)
	})
}

func performAirportJSONRequest(t *testing.T, handler gin.HandlerFunc, method, path string, body any, params ...gin.Param) *httptest.ResponseRecorder {
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
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), method, path, bytes.NewReader(requestBody))
	ginContext.Request.Header.Set("Content-Type", "application/json")
	ginContext.Params = params

	handler(ginContext)
	return recorder
}

func decodeAirportAPIResponse(t *testing.T, recorder *httptest.ResponseRecorder) apiJSONResponse {
	t.Helper()
	return decodeAPIResponse(t, recorder)
}

func TestAirportAddPersistsUpdateAfterDetectSettings(t *testing.T) {
	setupAirportAPITestDB(t)

	addBody := map[string]any{
		"name":                       "机场A",
		"url":                        "https://example.com/subscription",
		"cronExpr":                   "0 */12 * * *",
		"enabled":                    false,
		"group":                      "默认组",
		"downloadWithProxy":          false,
		"proxyLink":                  "",
		"userAgent":                  "",
		"requestHeaders":             []map[string]string{},
		"fetchUsageInfo":             false,
		"skipTLSVerify":              false,
		"updateAfterDetect":          true,
		"updateAfterDetectProfileId": 7,
		"remark":                     "",
		"logo":                       "",
		"nodeNameWhitelist":          "",
		"nodeNameBlacklist":          "",
		"protocolWhitelist":          "",
		"protocolBlacklist":          "",
		"nodeNamePreprocess":         "",
		"deduplicationRule":          "",
		"nodeNameUniquify":           false,
		"nodeNamePrefix":             "",
		"nodeNameIntraUniquify":      false,
	}

	addRecorder := performAirportJSONRequest(t, AirportAdd, http.MethodPost, "/api/v1/airports", addBody)
	if addRecorder.Code != http.StatusOK {
		t.Fatalf("add airport status = %d, body = %s", addRecorder.Code, addRecorder.Body.String())
	}
	addResponse := decodeAirportAPIResponse(t, addRecorder)
	if addResponse.Code != 200 {
		t.Fatalf("add airport response code = %d, msg = %s", addResponse.Code, addResponse.Msg)
	}

	var stored models.Airport
	if err := database.DB.Where("name = ?", "机场A").First(&stored).Error; err != nil {
		t.Fatalf("load added airport: %v", err)
	}
	if !stored.UpdateAfterDetect {
		t.Fatal("expected update_after_detect to be stored as true after add")
	}
	if stored.UpdateAfterDetectProfileID != 7 {
		t.Fatalf("expected update_after_detect_profile_id = 7 after add, got %d", stored.UpdateAfterDetectProfileID)
	}
}
