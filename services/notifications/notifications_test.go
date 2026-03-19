package notifications

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sublink/database"
	"sublink/internal/testutil"
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

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "notification_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.SystemSetting{}, &models.Webhook{}); err != nil {
		t.Fatalf("auto migrate test tables: %v", err)
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
	got := NormalizeEventKeys(ChannelWebhook, []string{"task.speed_test_completed", "subscription.sync_succeeded", "does.not.exist", "subscription.sync_failed"})
	want := []string{"subscription.sync_succeeded", "subscription.sync_failed", "task.speed_test_completed"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestFillPayloadMetaUsesCatalogDefaults(t *testing.T) {
	payload := FillPayloadMeta("security.user_login", Payload{Title: "登录成功"})
	if payload.EventName != "用户登录" || payload.CategoryName != "安全审计" || payload.Severity != "info" || payload.Time == "" {
		t.Fatalf("unexpected payload meta: %+v", payload)
	}
}

func TestListWebhookConfigsReturnsEmptyByDefault(t *testing.T) {
	setupNotificationTestDB(t)
	configs, err := ListWebhookConfigs()
	if err != nil {
		t.Fatalf("list webhook configs: %v", err)
	}
	if len(configs) != 0 {
		t.Fatalf("expected empty webhook list, got %d", len(configs))
	}
}

func TestWebhookConfigCRUDNormalizesMethodAndEvents(t *testing.T) {
	setupNotificationTestDB(t)

	created, err := CreateWebhookConfig(&WebhookConfig{
		Name:        "订阅告警",
		URL:         "https://example.com/hook",
		Method:      "put",
		ContentType: "text/plain",
		Headers:     `{"X-Test":"true"}`,
		Body:        "hello {{message}}",
		Enabled:     true,
		EventKeys:   []string{"task.speed_test_completed", "subscription.sync_failed"},
	})
	if err != nil {
		t.Fatalf("create webhook config: %v", err)
	}
	if created.ID == 0 || created.Method != http.MethodPut {
		t.Fatalf("unexpected created webhook: %+v", created)
	}

	loaded, err := GetWebhookConfig(created.ID)
	if err != nil {
		t.Fatalf("get webhook config: %v", err)
	}
	if loaded.ContentType != "text/plain" || !loaded.Enabled {
		t.Fatalf("unexpected loaded webhook: %+v", loaded)
	}
	wantEventKeys := []string{"subscription.sync_failed", "task.speed_test_completed"}
	if !reflect.DeepEqual(loaded.EventKeys, wantEventKeys) {
		t.Fatalf("event keys = %#v, want %#v", loaded.EventKeys, wantEventKeys)
	}

	loaded.Name = "失败通知"
	loaded.Enabled = false
	loaded.EventKeys = []string{"security.user_login"}
	updated, err := UpdateWebhookConfig(loaded)
	if err != nil {
		t.Fatalf("update webhook config: %v", err)
	}
	if updated.Name != "失败通知" || updated.Enabled {
		t.Fatalf("unexpected updated webhook: %+v", updated)
	}

	if err := DeleteWebhookConfig(created.ID); err != nil {
		t.Fatalf("delete webhook config: %v", err)
	}
	configs, err := ListWebhookConfigs()
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(configs) != 0 {
		t.Fatalf("expected empty list after delete, got %d", len(configs))
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

func TestSaveTelegramEventKeysNormalizesCatalogOrder(t *testing.T) {
	setupNotificationTestDB(t)
	if err := SaveTelegramEventKeys([]string{"security.user_login", "subscription.sync_failed", "invalid.event"}); err != nil {
		t.Fatalf("save telegram event keys: %v", err)
	}
	keys, err := LoadTelegramEventKeys()
	if err != nil {
		t.Fatalf("reload telegram event keys: %v", err)
	}
	want := []string{"subscription.sync_failed", "security.user_login"}
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("telegram event keys = %#v, want %#v", keys, want)
	}
}

func TestSendWebhookSupportsPUTAndTemplates(t *testing.T) {
	var gotMethod, gotContentType, gotHeader string
	var gotBody map[string]interface{}

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

	err := SendWebhook(&WebhookConfig{URL: server.URL + "/hooks/{{event}}", Method: http.MethodPut, ContentType: "application/json", Headers: `{"X-Test":"abc"}`, Body: `{"title":"{{title}}","severity":"{{severity}}","payload":{{json .}}}`}, Payload{Event: "task.speed_test_completed", EventName: "节点测速完成", Category: "task", CategoryName: "任务执行", Severity: "success", Title: "测速结束", Message: "完成", Time: "2026-03-18 10:00:00", Data: map[string]interface{}{"success": 3}})
	if err != nil {
		t.Fatalf("send webhook: %v", err)
	}
	if gotMethod != http.MethodPut || gotContentType != "application/json" || gotHeader != "abc" {
		t.Fatalf("unexpected request metadata: method=%s contentType=%s header=%s", gotMethod, gotContentType, gotHeader)
	}
	payload, ok := gotBody["payload"].(map[string]interface{})
	if !ok || payload["event"] != "task.speed_test_completed" {
		t.Fatalf("unexpected payload: %#v", gotBody)
	}
}

func TestSendWebhookGETOmitsBodyAndInterpolatesURL(t *testing.T) {
	var gotMethod, gotPath, gotBody string
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

	err := SendWebhook(&WebhookConfig{URL: server.URL + "/{{title}}?event={{event}}&severity={{severity}}", Method: http.MethodGet}, Payload{Event: "security.user_login", Severity: "info", Title: "登录成功", Message: "管理员已登录", Time: "2026-03-18 10:00:00"})
	if err != nil {
		t.Fatalf("send webhook GET: %v", err)
	}
	if gotMethod != http.MethodGet || !strings.Contains(gotPath, "event=security.user_login") || gotBody != "" {
		t.Fatalf("unexpected GET webhook request: method=%s path=%s body=%q", gotMethod, gotPath, gotBody)
	}
}

func TestSendWebhookRejectsInvalidHeaderJSON(t *testing.T) {
	err := SendWebhook(&WebhookConfig{URL: "https://example.com/hook", Method: http.MethodPost, Headers: "{invalid"}, Payload{Event: "security.user_login", Title: "登录成功"})
	if err == nil {
		t.Fatalf("expected invalid header JSON to return error")
	}
}

func TestTriggerWebhookFansOutToMatchingTargets(t *testing.T) {
	setupNotificationTestDB(t)

	var mutex sync.Mutex
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		requestCount++
		mutex.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := CreateWebhookConfig(&WebhookConfig{Name: "first", URL: server.URL + "/1", Method: http.MethodPost, Enabled: true, EventKeys: []string{"task.speed_test_completed"}})
	if err != nil {
		t.Fatalf("create first webhook: %v", err)
	}
	_, err = CreateWebhookConfig(&WebhookConfig{Name: "second", URL: server.URL + "/2", Method: http.MethodPost, Enabled: true, EventKeys: []string{"task.speed_test_completed", "subscription.sync_failed"}})
	if err != nil {
		t.Fatalf("create second webhook: %v", err)
	}
	_, err = CreateWebhookConfig(&WebhookConfig{Name: "disabled", URL: server.URL + "/3", Method: http.MethodPost, Enabled: false, EventKeys: []string{"task.speed_test_completed"}})
	if err != nil {
		t.Fatalf("create third webhook: %v", err)
	}

	TriggerWebhook("subscription.sync_failed", Payload{Event: "subscription.sync_failed", Title: "订阅更新失败", Message: "失败"})
	mutex.Lock()
	gotBefore := requestCount
	mutex.Unlock()
	if gotBefore != 1 {
		t.Fatalf("expected one matching request for subscription failure, got %d", gotBefore)
	}

	TriggerWebhook("task.speed_test_completed", Payload{Event: "task.speed_test_completed", Title: "测速完成", Message: "成功"})
	mutex.Lock()
	gotAfter := requestCount
	mutex.Unlock()
	if gotAfter != 3 {
		t.Fatalf("expected total three requests after fan-out, got %d", gotAfter)
	}
}

func TestTriggerWebhookFailureDoesNotBlockOtherTargets(t *testing.T) {
	setupNotificationTestDB(t)

	var successCount int
	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		successCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer goodServer.Close()

	_, err := CreateWebhookConfig(&WebhookConfig{Name: "bad", URL: "http://127.0.0.1:1/unreachable", Method: http.MethodPost, Enabled: true, EventKeys: []string{"task.speed_test_completed"}})
	if err != nil {
		t.Fatalf("create bad webhook: %v", err)
	}
	_, err = CreateWebhookConfig(&WebhookConfig{Name: "good", URL: goodServer.URL, Method: http.MethodPost, Enabled: true, EventKeys: []string{"task.speed_test_completed"}})
	if err != nil {
		t.Fatalf("create good webhook: %v", err)
	}

	TriggerWebhook("task.speed_test_completed", Payload{Event: "task.speed_test_completed", Title: "测速完成", Message: "成功"})
	if successCount != 1 {
		t.Fatalf("expected successful webhook to still receive request, got %d", successCount)
	}
}

func TestTriggerTelegramRespectsSelectedEvents(t *testing.T) {
	setupNotificationTestDB(t)
	oldSender := telegramSender
	defer func() { telegramSender = oldSender }()
	if err := SaveTelegramEventKeys([]string{"security.user_login"}); err != nil {
		t.Fatalf("save telegram event keys: %v", err)
	}
	calls := make(chan Payload, 1)
	RegisterTelegramSender(func(eventKey string, payload Payload) { calls <- payload })
	TriggerTelegram("task.speed_test_completed", Payload{Event: "task.speed_test_completed", Title: "测速完成"})
	select {
	case payload := <-calls:
		t.Fatalf("expected no telegram callback for filtered event, got %+v", payload)
	default:
	}
	TriggerTelegram("security.user_login", Payload{Event: "security.user_login", Title: "登录成功"})
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
	if err := SaveTelegramEventKeys([]string{"security.user_login"}); err != nil {
		t.Fatalf("save telegram event keys: %v", err)
	}
	telegramCalls := make(chan Payload, 1)
	RegisterTelegramSender(func(eventKey string, payload Payload) { telegramCalls <- payload })
	Publish("security.user_login", Payload{Title: "登录成功", Message: "管理员已登录", Time: "2026-03-18 10:00:00"})
	select {
	case raw := <-sse.GetSSEBroker().Notifier:
		message := string(raw)
		if !strings.HasPrefix(message, "event: notification\n") || !strings.Contains(message, `"event":"security.user_login"`) {
			t.Fatalf("unexpected SSE event: %s", message)
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
