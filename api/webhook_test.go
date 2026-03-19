package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type testAPIResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func setupWebhookAPITestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "webhook_api_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.SystemSetting{}, &models.Webhook{}); err != nil {
		t.Fatalf("auto migrate webhook tables: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	if err := models.InitSettingCache(); err != nil {
		t.Fatalf("init setting cache: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Exec("DELETE FROM webhooks").Error
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

func performWebhookJSONRequest(t *testing.T, handler gin.HandlerFunc, method string, path string, body interface{}) *httptest.ResponseRecorder {
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
	context.Request = httptest.NewRequest(method, path, bytes.NewReader(requestBody))
	context.Request.Header.Set("Content-Type", "application/json")
	if strings.Contains(path, "/webhooks/") {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		for i, part := range parts {
			if part == "webhooks" && i+1 < len(parts) {
				context.Params = gin.Params{{Key: "id", Value: parts[i+1]}}
				break
			}
		}
	}

	handler(context)
	return recorder
}

func decodeWebhookAPIResponse(t *testing.T, recorder *httptest.ResponseRecorder) testAPIResponse {
	t.Helper()
	var response testAPIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal api response: %v", err)
	}
	return response
}

func TestListWebhooksReturnsEventOptions(t *testing.T) {
	setupWebhookAPITestDB(t)

	recorder := performWebhookJSONRequest(t, ListWebhooks, http.MethodGet, "/api/v1/settings/webhooks", nil)
	response := decodeWebhookAPIResponse(t, recorder)
	if response.Code != 200 {
		t.Fatalf("expected response code 200, got %d", response.Code)
	}

	var data struct {
		Items        []map[string]interface{} `json:"items"`
		EventOptions []map[string]interface{} `json:"eventOptions"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatalf("unmarshal webhook list data: %v", err)
	}
	if len(data.Items) != 0 {
		t.Fatalf("expected empty list by default, got %d items", len(data.Items))
	}
	if len(data.EventOptions) == 0 {
		t.Fatalf("expected event options to be returned")
	}
}

func TestCreateAndUpdateWebhookPersistData(t *testing.T) {
	setupWebhookAPITestDB(t)

	createRecorder := performWebhookJSONRequest(t, CreateWebhook, http.MethodPost, "/api/v1/settings/webhooks", map[string]interface{}{
		"name":               "错误告警",
		"webhookUrl":         "https://example.com/hook",
		"webhookMethod":      "PUT",
		"webhookContentType": "text/plain",
		"webhookHeaders":     `{"X-Test":"1"}`,
		"webhookBody":        "hello {{message}}",
		"webhookEnabled":     true,
		"eventKeys":          []string{"task.speed_test_completed", "subscription.sync_failed"},
	})
	createResponse := decodeWebhookAPIResponse(t, createRecorder)
	if createResponse.Code != 200 {
		t.Fatalf("expected create response code 200, got %d", createResponse.Code)
	}

	var created struct {
		ID             uint     `json:"id"`
		Name           string   `json:"name"`
		WebhookMethod  string   `json:"webhookMethod"`
		WebhookEnabled bool     `json:"webhookEnabled"`
		EventKeys      []string `json:"eventKeys"`
	}
	if err := json.Unmarshal(createResponse.Data, &created); err != nil {
		t.Fatalf("unmarshal created webhook: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected created webhook to have id")
	}
	if created.Name != "错误告警" {
		t.Fatalf("unexpected webhook name: %s", created.Name)
	}

	updateRecorder := performWebhookJSONRequest(t, UpdateWebhook, http.MethodPut, "/api/v1/settings/webhooks/1", map[string]interface{}{
		"name":               "订阅失败告警",
		"webhookUrl":         "https://example.com/hook-2",
		"webhookMethod":      "POST",
		"webhookContentType": "application/json",
		"webhookHeaders":     `{"Authorization":"Bearer token"}`,
		"webhookBody":        `{"title":"{{title}}"}`,
		"webhookEnabled":     false,
		"eventKeys":          []string{"subscription.sync_failed"},
	})
	updateResponse := decodeWebhookAPIResponse(t, updateRecorder)
	if updateResponse.Code != 200 {
		t.Fatalf("expected update response code 200, got %d", updateResponse.Code)
	}

	webhook, err := models.GetWebhookByID(created.ID)
	if err != nil {
		t.Fatalf("reload webhook: %v", err)
	}
	if webhook.Name != "订阅失败告警" {
		t.Fatalf("unexpected updated name: %s", webhook.Name)
	}
	if webhook.URL != "https://example.com/hook-2" {
		t.Fatalf("unexpected updated url: %s", webhook.URL)
	}
	if webhook.Enabled {
		t.Fatalf("expected webhook to be disabled after update")
	}
}

func TestCreateWebhookRejectsInvalidHeaderJSON(t *testing.T) {
	setupWebhookAPITestDB(t)

	recorder := performWebhookJSONRequest(t, CreateWebhook, http.MethodPost, "/api/v1/settings/webhooks", map[string]interface{}{
		"webhookUrl":     "https://example.com/hook",
		"webhookHeaders": "{invalid",
	})
	response := decodeWebhookAPIResponse(t, recorder)
	if response.Code != 500 {
		t.Fatalf("expected business error code 500, got %d", response.Code)
	}
	if !strings.Contains(response.Msg, "Headers") {
		t.Fatalf("expected error message to mention headers, got %s", response.Msg)
	}
}

func TestTestWebhookByIDSendsConfiguredRequest(t *testing.T) {
	setupWebhookAPITestDB(t)

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

	webhook := &models.Webhook{
		Name:        "测试 Webhook",
		URL:         server.URL,
		Method:      "PUT",
		ContentType: "text/plain",
		Headers:     `{"X-Test":"1"}`,
		Body:        "{{title}}|{{event}}|{{severity}}",
		Enabled:     true,
		EventKeys:   `[]`,
	}
	if err := webhook.Add(); err != nil {
		t.Fatalf("seed webhook: %v", err)
	}

	recorder := performWebhookJSONRequest(t, TestWebhookByID, http.MethodPost, "/api/v1/settings/webhooks/1/test", nil)
	response := decodeWebhookAPIResponse(t, recorder)
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

func TestDeleteWebhookRemovesOnlyTarget(t *testing.T) {
	setupWebhookAPITestDB(t)

	first := &models.Webhook{Name: "one", URL: "https://example.com/1", Method: "POST", ContentType: "application/json", EventKeys: `[]`}
	second := &models.Webhook{Name: "two", URL: "https://example.com/2", Method: "POST", ContentType: "application/json", EventKeys: `[]`}
	if err := first.Add(); err != nil {
		t.Fatalf("add first webhook: %v", err)
	}
	if err := second.Add(); err != nil {
		t.Fatalf("add second webhook: %v", err)
	}

	recorder := performWebhookJSONRequest(t, DeleteWebhook, http.MethodDelete, "/api/v1/settings/webhooks/1", nil)
	response := decodeWebhookAPIResponse(t, recorder)
	if response.Code != 200 {
		t.Fatalf("expected response code 200, got %d", response.Code)
	}

	items, err := models.ListWebhooks()
	if err != nil {
		t.Fatalf("list webhooks: %v", err)
	}
	if len(items) != 1 || items[0].Name != "two" {
		t.Fatalf("unexpected remaining webhooks: %+v", items)
	}
}
