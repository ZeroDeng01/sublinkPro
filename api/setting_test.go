package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sublink/database"
	"sublink/services/notifications"
	"testing"

	"sublink/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type testAPIResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func setupSettingAPITestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
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
		database.DB = db
		database.Dialect = database.DialectSQLite
		database.IsInitialized = false
		_ = models.InitSettingCache()

		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		if oldDB != nil {
			_ = models.InitSettingCache()
		}
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
	context.Request = httptest.NewRequest(method, "/api/v1/settings/webhook", bytes.NewReader(requestBody))
	context.Request.Header.Set("Content-Type", "application/json")

	handler(context)
	return recorder
}

func decodeAPIResponse(t *testing.T, recorder *httptest.ResponseRecorder) testAPIResponse {
	t.Helper()

	var response testAPIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal api response: %v", err)
	}
	return response
}

func TestGetWebhookConfigReturnsDefaultsAndEventOptions(t *testing.T) {
	setupSettingAPITestDB(t)

	recorder := performJSONRequest(t, GetWebhookConfig, http.MethodGet, nil)
	response := decodeAPIResponse(t, recorder)

	if response.Code != 200 {
		t.Fatalf("expected response code 200, got %d", response.Code)
	}

	var data struct {
		WebhookMethod      string                          `json:"webhookMethod"`
		WebhookContentType string                          `json:"webhookContentType"`
		EventKeys          []string                        `json:"eventKeys"`
		EventOptions       []notifications.EventDefinition `json:"eventOptions"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatalf("unmarshal webhook config data: %v", err)
	}

	if data.WebhookMethod != http.MethodPost {
		t.Fatalf("expected default method POST, got %s", data.WebhookMethod)
	}
	if data.WebhookContentType != "application/json" {
		t.Fatalf("expected default content type application/json, got %s", data.WebhookContentType)
	}
	if len(data.EventKeys) == 0 {
		t.Fatalf("expected default event keys to be returned")
	}
	if len(data.EventOptions) == 0 {
		t.Fatalf("expected event options to be returned")
	}
}

func TestUpdateWebhookConfigPersistsConfigAndEventKeys(t *testing.T) {
	setupSettingAPITestDB(t)

	recorder := performJSONRequest(t, UpdateWebhookConfig, http.MethodPost, map[string]interface{}{
		"webhookUrl":         "https://example.com/hook",
		"webhookMethod":      "PUT",
		"webhookContentType": "text/plain",
		"webhookHeaders":     `{"X-Test":"1"}`,
		"webhookBody":        "hello {{message}}",
		"webhookEnabled":     true,
		"eventKeys": []string{
			"task.speed_test_completed",
			"subscription.sync_failed",
		},
	})
	response := decodeAPIResponse(t, recorder)
	if response.Code != 200 {
		t.Fatalf("expected response code 200, got %d", response.Code)
	}

	config, err := notifications.LoadWebhookConfig()
	if err != nil {
		t.Fatalf("load webhook config: %v", err)
	}

	if config.URL != "https://example.com/hook" {
		t.Fatalf("unexpected webhook URL: %s", config.URL)
	}
	if config.Method != http.MethodPut {
		t.Fatalf("expected method PUT, got %s", config.Method)
	}
	if config.ContentType != "text/plain" {
		t.Fatalf("unexpected content type: %s", config.ContentType)
	}
	if !config.Enabled {
		t.Fatalf("expected webhook to be enabled")
	}
	if len(config.EventKeys) != 2 {
		t.Fatalf("expected 2 selected event keys, got %d", len(config.EventKeys))
	}
}

func TestUpdateWebhookConfigRejectsInvalidHeaderJSON(t *testing.T) {
	setupSettingAPITestDB(t)

	recorder := performJSONRequest(t, UpdateWebhookConfig, http.MethodPost, map[string]interface{}{
		"webhookUrl":     "https://example.com/hook",
		"webhookHeaders": "{invalid",
	})
	response := decodeAPIResponse(t, recorder)

	if response.Code != 500 {
		t.Fatalf("expected business error code 500, got %d", response.Code)
	}
	if !strings.Contains(response.Msg, "Headers") {
		t.Fatalf("expected error message to mention headers, got %s", response.Msg)
	}
}

func TestTestWebhookConfigSendsConfiguredRequest(t *testing.T) {
	setupSettingAPITestDB(t)

	var (
		gotMethod string
		gotBody   string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		gotMethod = r.Method
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	recorder := performJSONRequest(t, TestWebhookConfig, http.MethodPost, map[string]interface{}{
		"webhookUrl":         server.URL,
		"webhookMethod":      "PUT",
		"webhookContentType": "text/plain",
		"webhookHeaders":     `{"X-Test":"1"}`,
		"webhookBody":        "{{title}}|{{event}}|{{severity}}",
	})
	response := decodeAPIResponse(t, recorder)

	if response.Code != 200 {
		t.Fatalf("expected response code 200, got %d", response.Code)
	}
	if gotMethod != http.MethodPut {
		t.Fatalf("expected method PUT, got %s", gotMethod)
	}
	if !strings.Contains(gotBody, "Sublink Pro Webhook 测试|test.webhook|info") {
		t.Fatalf("unexpected request body: %s", gotBody)
	}
}
