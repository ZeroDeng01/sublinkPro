package notifications

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sublink/database"
	"sublink/models"
	"sublink/services/sse"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupNotificationTestDB(t *testing.T) {
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

func drainSSENotifier() {
	broker := sse.GetSSEBroker()
	for {
		select {
		case <-broker.Notifier:
		default:
			return
		}
	}
}

func TestEventCatalogForChannelFiltersByChannel(t *testing.T) {
	events := EventCatalogForChannel(ChannelWebhook)
	if len(events) == 0 {
		t.Fatalf("expected webhook events to be available")
	}

	for _, event := range events {
		found := false
		for _, channel := range event.Channels {
			if channel == ChannelWebhook {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("event %s missing webhook channel", event.Key)
		}
	}
}

func TestNormalizeEventKeysKeepsCatalogOrder(t *testing.T) {
	got := NormalizeEventKeys(ChannelWebhook, []string{
		"task.speed_test_completed",
		"subscription.sync_succeeded",
		"does.not.exist",
		"subscription.sync_failed",
	})

	want := []string{
		"subscription.sync_succeeded",
		"subscription.sync_failed",
		"task.speed_test_completed",
	}

	if len(got) != len(want) {
		t.Fatalf("len(got)=%d, want=%d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%s, want=%s", i, got[i], want[i])
		}
	}
}

func TestFillPayloadMetaUsesCatalogDefaults(t *testing.T) {
	payload := FillPayloadMeta("security.user_login", Payload{
		Title: "登录成功",
	})

	if payload.Event != "security.user_login" {
		t.Fatalf("expected event to be filled, got %s", payload.Event)
	}
	if payload.EventName != "用户登录" {
		t.Fatalf("expected event name 用户登录, got %s", payload.EventName)
	}
	if payload.Category != "security" {
		t.Fatalf("expected category security, got %s", payload.Category)
	}
	if payload.CategoryName != "安全审计" {
		t.Fatalf("expected category name 安全审计, got %s", payload.CategoryName)
	}
	if payload.Severity != "info" {
		t.Fatalf("expected severity info, got %s", payload.Severity)
	}
	if payload.Time == "" {
		t.Fatalf("expected time to be populated")
	}
}

func TestLoadWebhookConfigReturnsDefaults(t *testing.T) {
	setupNotificationTestDB(t)

	config, err := LoadWebhookConfig()
	if err != nil {
		t.Fatalf("load webhook config: %v", err)
	}

	if config.Method != http.MethodPost {
		t.Fatalf("expected default method POST, got %s", config.Method)
	}
	if config.ContentType != "application/json" {
		t.Fatalf("expected default content type application/json, got %s", config.ContentType)
	}
	if config.Enabled {
		t.Fatalf("expected webhook to be disabled by default")
	}

	wantEventKeys := DefaultEventKeys(ChannelWebhook)
	if len(config.EventKeys) != len(wantEventKeys) {
		t.Fatalf("len(eventKeys)=%d, want=%d", len(config.EventKeys), len(wantEventKeys))
	}
}

func TestLoadTelegramEventKeysReturnsDefaults(t *testing.T) {
	setupNotificationTestDB(t)

	keys, err := LoadTelegramEventKeys()
	if err != nil {
		t.Fatalf("load telegram event keys: %v", err)
	}

	want := DefaultEventKeys(ChannelTelegram)
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("telegram default event keys = %#v, want %#v", keys, want)
	}
}

func TestSaveWebhookConfigRoundTripNormalizesMethodAndEvents(t *testing.T) {
	setupNotificationTestDB(t)

	err := SaveWebhookConfig(&WebhookConfig{
		URL:         "https://example.com/hook",
		Method:      "put",
		ContentType: "text/plain",
		Headers:     `{"X-Test":"true"}`,
		Body:        "hello {{message}}",
		Enabled:     true,
		EventKeys: []string{
			"task.speed_test_completed",
			"subscription.sync_failed",
		},
	})
	if err != nil {
		t.Fatalf("save webhook config: %v", err)
	}

	config, err := LoadWebhookConfig()
	if err != nil {
		t.Fatalf("reload webhook config: %v", err)
	}

	if config.Method != http.MethodPut {
		t.Fatalf("expected method PUT, got %s", config.Method)
	}
	if config.ContentType != "text/plain" {
		t.Fatalf("expected content type text/plain, got %s", config.ContentType)
	}
	if !config.Enabled {
		t.Fatalf("expected webhook to be enabled")
	}

	wantEventKeys := []string{
		"subscription.sync_failed",
		"task.speed_test_completed",
	}
	if len(config.EventKeys) != len(wantEventKeys) {
		t.Fatalf("len(eventKeys)=%d, want=%d", len(config.EventKeys), len(wantEventKeys))
	}
	for i := range wantEventKeys {
		if config.EventKeys[i] != wantEventKeys[i] {
			t.Fatalf("eventKeys[%d]=%s, want=%s", i, config.EventKeys[i], wantEventKeys[i])
		}
	}
}

func TestSaveTelegramEventKeysNormalizesCatalogOrder(t *testing.T) {
	setupNotificationTestDB(t)

	err := SaveTelegramEventKeys([]string{
		"security.user_login",
		"subscription.sync_failed",
		"invalid.event",
	})
	if err != nil {
		t.Fatalf("save telegram event keys: %v", err)
	}

	keys, err := LoadTelegramEventKeys()
	if err != nil {
		t.Fatalf("reload telegram event keys: %v", err)
	}

	want := []string{
		"subscription.sync_failed",
		"security.user_login",
	}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("telegram event keys = %#v, want %#v", keys, want)
	}
}

func TestSendWebhookSupportsPUTAndTemplates(t *testing.T) {
	var (
		gotMethod      string
		gotContentType string
		gotHeader      string
		gotBody        map[string]interface{}
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotHeader = r.Header.Get("X-Test")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if err := json.Unmarshal(body, &gotBody); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := SendWebhook(&WebhookConfig{
		URL:         server.URL + "/hooks/{{event}}",
		Method:      http.MethodPut,
		ContentType: "application/json",
		Headers:     `{"X-Test":"abc"}`,
		Body:        `{"title":"{{title}}","severity":"{{severity}}","payload":{{json .}}}`,
	}, Payload{
		Event:        "task.speed_test_completed",
		EventName:    "节点测速完成",
		Category:     "task",
		CategoryName: "任务执行",
		Severity:     "success",
		Title:        "测速结束",
		Message:      "完成",
		Time:         "2026-03-18 10:00:00",
		Data: map[string]interface{}{
			"success": 3,
		},
	})
	if err != nil {
		t.Fatalf("send webhook: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Fatalf("expected method PUT, got %s", gotMethod)
	}
	if gotContentType != "application/json" {
		t.Fatalf("expected content type application/json, got %s", gotContentType)
	}
	if gotHeader != "abc" {
		t.Fatalf("expected X-Test header abc, got %s", gotHeader)
	}
	if gotBody["title"] != "测速结束" {
		t.Fatalf("unexpected title: %v", gotBody["title"])
	}
	if gotBody["severity"] != "success" {
		t.Fatalf("unexpected severity: %v", gotBody["severity"])
	}

	payload, ok := gotBody["payload"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected nested payload object, got %T", gotBody["payload"])
	}
	if payload["event"] != "task.speed_test_completed" {
		t.Fatalf("unexpected payload.event: %v", payload["event"])
	}
	if payload["eventName"] != "节点测速完成" {
		t.Fatalf("unexpected payload.eventName: %v", payload["eventName"])
	}
}

func TestSendWebhookGETOmitsBodyAndInterpolatesURL(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotBody   string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		gotMethod = r.Method
		gotPath = r.URL.String()
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := SendWebhook(&WebhookConfig{
		URL:    server.URL + "/{{title}}?event={{event}}&severity={{severity}}",
		Method: http.MethodGet,
	}, Payload{
		Event:    "security.user_login",
		Severity: "info",
		Title:    "登录成功",
		Message:  "管理员已登录",
		Time:     "2026-03-18 10:00:00",
	})
	if err != nil {
		t.Fatalf("send webhook GET: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Fatalf("expected method GET, got %s", gotMethod)
	}
	if !strings.Contains(gotPath, "/%E7%99%BB%E5%BD%95%E6%88%90%E5%8A%9F?") {
		t.Fatalf("expected encoded title in URL, got %s", gotPath)
	}
	if !strings.Contains(gotPath, "event=security.user_login") {
		t.Fatalf("expected event query in URL, got %s", gotPath)
	}
	if gotBody != "" {
		t.Fatalf("expected empty body for GET, got %q", gotBody)
	}
}

func TestSendWebhookRejectsInvalidHeaderJSON(t *testing.T) {
	err := SendWebhook(&WebhookConfig{
		URL:     "https://example.com/hook",
		Method:  http.MethodPost,
		Headers: "{invalid",
	}, Payload{
		Event: "security.user_login",
		Title: "登录成功",
	})
	if err == nil {
		t.Fatalf("expected invalid header JSON to return error")
	}
}

func TestTriggerWebhookRespectsSelectedEvents(t *testing.T) {
	setupNotificationTestDB(t)

	var (
		mutex        sync.Mutex
		requestCount int
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		requestCount++
		mutex.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := SaveWebhookConfig(&WebhookConfig{
		URL:       server.URL,
		Method:    http.MethodPost,
		Enabled:   true,
		EventKeys: []string{"task.speed_test_completed"},
	})
	if err != nil {
		t.Fatalf("save webhook config: %v", err)
	}

	TriggerWebhook("subscription.sync_failed", Payload{
		Event:   "subscription.sync_failed",
		Title:   "订阅更新失败",
		Message: "失败",
	})

	mutex.Lock()
	gotBefore := requestCount
	mutex.Unlock()
	if gotBefore != 0 {
		t.Fatalf("expected filtered event to send no request, got %d", gotBefore)
	}

	TriggerWebhook("task.speed_test_completed", Payload{
		Event:   "task.speed_test_completed",
		Title:   "测速完成",
		Message: "成功",
	})

	mutex.Lock()
	gotAfter := requestCount
	mutex.Unlock()
	if gotAfter != 1 {
		t.Fatalf("expected enabled event to send one request, got %d", gotAfter)
	}
}

func TestTriggerTelegramRespectsSelectedEvents(t *testing.T) {
	setupNotificationTestDB(t)

	oldSender := telegramSender
	defer func() {
		telegramSender = oldSender
	}()

	if err := SaveTelegramEventKeys([]string{"security.user_login"}); err != nil {
		t.Fatalf("save telegram event keys: %v", err)
	}

	calls := make(chan Payload, 1)
	RegisterTelegramSender(func(eventKey string, payload Payload) {
		calls <- payload
	})

	TriggerTelegram("task.speed_test_completed", Payload{
		Event: "task.speed_test_completed",
		Title: "测速完成",
	})

	select {
	case payload := <-calls:
		t.Fatalf("expected no telegram callback for filtered event, got %+v", payload)
	default:
	}

	TriggerTelegram("security.user_login", Payload{
		Event: "security.user_login",
		Title: "登录成功",
	})

	select {
	case payload := <-calls:
		if payload.Event != "security.user_login" {
			t.Fatalf("unexpected event: %s", payload.Event)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected telegram callback to be invoked")
	}
}

func TestPublishBroadcastsSSEAndSelectedTelegram(t *testing.T) {
	setupNotificationTestDB(t)
	drainSSENotifier()

	oldSender := telegramSender
	defer func() {
		telegramSender = oldSender
		drainSSENotifier()
	}()

	if err := SaveWebhookConfig(&WebhookConfig{Enabled: false}); err != nil {
		t.Fatalf("disable webhook: %v", err)
	}
	if err := SaveTelegramEventKeys([]string{"security.user_login"}); err != nil {
		t.Fatalf("save telegram event keys: %v", err)
	}

	telegramCalls := make(chan Payload, 1)
	RegisterTelegramSender(func(eventKey string, payload Payload) {
		telegramCalls <- payload
	})

	Publish("security.user_login", Payload{
		Title:   "登录成功",
		Message: "管理员已登录",
		Time:    "2026-03-18 10:00:00",
	})

	select {
	case raw := <-sse.GetSSEBroker().Notifier:
		message := string(raw)
		if !strings.HasPrefix(message, "event: notification\n") {
			t.Fatalf("unexpected SSE event: %s", message)
		}
		if !strings.Contains(message, `"event":"security.user_login"`) {
			t.Fatalf("expected SSE payload to include event key, got %s", message)
		}
		if !strings.Contains(message, `"categoryName":"安全审计"`) {
			t.Fatalf("expected SSE payload to include category name, got %s", message)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected SSE notification to be published")
	}

	select {
	case payload := <-telegramCalls:
		if payload.EventName != "用户登录" {
			t.Fatalf("expected telegram payload event name 用户登录, got %s", payload.EventName)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected telegram sender to be called")
	}
}
