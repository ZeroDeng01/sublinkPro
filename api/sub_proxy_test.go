package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"

	"github.com/gin-gonic/gin"
)

func setupProxyRelayTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db := testutil.OpenMemoryDB(t, "proxy_relay_test")
	if err := db.AutoMigrate(
		&models.Airport{},
		&models.Subcription{},
		&models.SubscriptionShare{},
		&models.SubcriptionNode{},
		&models.SubcriptionGroup{},
		&models.SubcriptionAirport{},
		&models.SubcriptionScript{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	if err := models.InitAirportCache(); err != nil {
		t.Fatalf("init airport cache: %v", err)
	}
	if err := models.InitSubcriptionCache(); err != nil {
		t.Fatalf("init subcription cache: %v", err)
	}
	if err := models.InitSubscriptionShareCache(); err != nil {
		t.Fatalf("init subscription share cache: %v", err)
	}

	t.Cleanup(func() {
		// 等待在飞的异步访问统计写入完成，避免其在拆库后解引用 nil 的 database.DB。
		models.WaitForPendingAccessRecords()
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		testutil.CloseDB(t, db)
	})
}

func addRelayAirport(t *testing.T, airport models.Airport) models.Airport {
	t.Helper()

	if err := airport.Add(); err != nil {
		t.Fatalf("add airport: %v", err)
	}
	return airport
}

func performRelayRequest(t *testing.T, sub models.Subcription, method string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), method, "/c?token=test", nil)

	serveProxySubscription(ginContext, sub)
	return recorder
}

// performSubcriptionFormRequest 以表单方式调用订阅接口（SubAdd / SubUpdate 使用 c.PostForm）。
func performSubcriptionFormRequest(t *testing.T, handler gin.HandlerFunc, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/",
		strings.NewReader(form.Encode()),
	)
	ginContext.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler(ginContext)
	return recorder
}

func TestServeProxySubscriptionRelaysAirportResponse(t *testing.T) {
	setupProxyRelayTestDB(t)

	var gotUserAgent, gotCustomHeader, gotHWID string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		gotCustomHeader = r.Header.Get("X-Airport-Token")
		gotHWID = r.Header.Get("X-HWID")
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		w.Header().Set("subscription-userinfo", "upload=1; download=2; total=3")
		w.Header().Set("profile-update-interval", "12")
		_, _ = w.Write([]byte("proxies:\n  - name: relayed\n"))
	}))
	defer upstream.Close()

	airport := addRelayAirport(t, models.Airport{
		Name:      "relay-airport",
		URL:       upstream.URL,
		UserAgent: "AirportApp/1.0",
		RequestHeaders: models.AirportRequestHeaders{
			{Key: "X-Airport-Token", Value: "secret"},
		},
		DisableHWID: true,
	})

	sub := models.Subcription{
		Name:      "relay",
		Type:      models.SubcriptionTypeProxy,
		AirportID: airport.ID,
	}

	recorder := performRelayRequest(t, sub, http.MethodGet)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); body != "proxies:\n  - name: relayed\n" {
		t.Fatalf("body = %q", body)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/yaml; charset=utf-8" {
		t.Fatalf("content-type = %q", got)
	}
	if got := recorder.Header().Get("subscription-userinfo"); got != "upload=1; download=2; total=3" {
		t.Fatalf("subscription-userinfo = %q", got)
	}
	if got := recorder.Header().Get("profile-update-interval"); got != "12" {
		t.Fatalf("profile-update-interval = %q", got)
	}
	if gotUserAgent != "AirportApp/1.0" {
		t.Fatalf("upstream user-agent = %q", gotUserAgent)
	}
	if gotCustomHeader != "secret" {
		t.Fatalf("upstream custom header = %q", gotCustomHeader)
	}
	if gotHWID != "" {
		t.Fatalf("disabled X-HWID should not be sent, got %q", gotHWID)
	}
}

func TestServeProxySubscriptionForwardsClientUserAgent(t *testing.T) {
	setupProxyRelayTestDB(t)

	var gotUserAgent string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	airport := addRelayAirport(t, models.Airport{
		Name:        "client-ua-airport",
		URL:         upstream.URL,
		UserAgent:   "clash.meta",
		DisableHWID: true,
	})
	sub := models.Subcription{Name: "relay", Type: models.SubcriptionTypeProxy, AirportID: airport.ID}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/c?token=test", nil)
	ginContext.Request.Header.Set("User-Agent", "Happ/2.0 (Android 14; Pixel 8)")

	serveProxySubscription(ginContext, sub)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if gotUserAgent != "Happ/2.0 (Android 14; Pixel 8)" {
		t.Fatalf("upstream user-agent = %q, want client UA", gotUserAgent)
	}
}

func TestServeProxySubscriptionSendsDeviceHeadersFromAirport(t *testing.T) {
	setupProxyRelayTestDB(t)

	var gotHWID, gotDeviceOS string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHWID = r.Header.Get("X-HWID")
		gotDeviceOS = r.Header.Get("X-Device-OS")
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	airport := addRelayAirport(t, models.Airport{
		Name:          "device-airport",
		URL:           upstream.URL,
		HWID:          "abcdef0123456789abcdef0123456789",
		DeviceProfile: models.AirportDeviceProfileAndroid,
		DeviceOS:      "Android",
		OSVersion:     "14",
	})

	sub := models.Subcription{Name: "relay", Type: models.SubcriptionTypeProxy, AirportID: airport.ID}
	if recorder := performRelayRequest(t, sub, http.MethodGet); recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if gotHWID != airport.HWID {
		t.Fatalf("upstream X-HWID = %q, want %q", gotHWID, airport.HWID)
	}
	if gotDeviceOS != "Android" {
		t.Fatalf("upstream X-Device-OS = %q, want Android", gotDeviceOS)
	}
}

func TestServeProxySubscriptionHeadSkipsBody(t *testing.T) {
	setupProxyRelayTestDB(t)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("payload"))
	}))
	defer upstream.Close()

	airport := addRelayAirport(t, models.Airport{Name: "head-airport", URL: upstream.URL, DisableHWID: true})
	sub := models.Subcription{Name: "relay", Type: models.SubcriptionTypeProxy, AirportID: airport.ID}

	recorder := performRelayRequest(t, sub, http.MethodHead)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("HEAD body = %q, want empty", recorder.Body.String())
	}
}

func TestServeProxySubscriptionUpstreamFailureReturnsBadGateway(t *testing.T) {
	setupProxyRelayTestDB(t)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer upstream.Close()

	airport := addRelayAirport(t, models.Airport{Name: "broken-airport", URL: upstream.URL, DisableHWID: true})
	sub := models.Subcription{Name: "relay", Type: models.SubcriptionTypeProxy, AirportID: airport.ID}

	recorder := performRelayRequest(t, sub, http.MethodGet)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadGateway)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "Failed to fetch upstream subscription") {
		t.Fatalf("body = %q", body)
	}
}

func TestServeProxySubscriptionWithoutAirportReturnsBadGateway(t *testing.T) {
	setupProxyRelayTestDB(t)

	sub := models.Subcription{Name: "relay", Type: models.SubcriptionTypeProxy}
	recorder := performRelayRequest(t, sub, http.MethodGet)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadGateway)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "Subscription is not linked to an airport") {
		t.Fatalf("body = %q", body)
	}
}

func TestServeProxySubscriptionMissingAirportReturnsBadGateway(t *testing.T) {
	setupProxyRelayTestDB(t)

	sub := models.Subcription{Name: "relay", Type: models.SubcriptionTypeProxy, AirportID: 987654}
	recorder := performRelayRequest(t, sub, http.MethodGet)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadGateway)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "The airport linked to this subscription does not exist") {
		t.Fatalf("body = %q", body)
	}
}

func TestServeProxySubscriptionInvalidAirportURLReturnsBadGateway(t *testing.T) {
	setupProxyRelayTestDB(t)

	airport := addRelayAirport(t, models.Airport{Name: "bad-url-airport", URL: "ftp://example.com/sub", DisableHWID: true})
	sub := models.Subcription{Name: "relay", Type: models.SubcriptionTypeProxy, AirportID: airport.ID}

	recorder := performRelayRequest(t, sub, http.MethodGet)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadGateway)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "Invalid airport subscription URL") {
		t.Fatalf("body = %q", body)
	}
}

func TestAirportDeleteBlockedByProxySubscription(t *testing.T) {
	setupProxyRelayTestDB(t)

	airport := addRelayAirport(t, models.Airport{Name: "used-airport", URL: "https://example.com/sub", DisableHWID: true})
	sub := models.Subcription{Name: "relay", Type: models.SubcriptionTypeProxy, AirportID: airport.ID}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}

	recorder := performAirportJSONRequest(
		t,
		AirportDelete,
		http.MethodDelete,
		"/api/v1/airports/"+strconv.Itoa(airport.ID),
		nil,
		gin.Param{Key: "id", Value: strconv.Itoa(airport.ID)},
	)
	response := decodeAirportAPIResponse(t, recorder)

	if response.Code == 200 {
		t.Fatal("expected airport delete to be rejected while proxy subscriptions reference it")
	}
	if response.I18nKey != "backend.airports.inUseByProxySubscriptions" {
		t.Fatalf("unexpected delete error key: %q (msg %q)", response.I18nKey, response.Msg)
	}
}

func TestSubAddProxySubscriptionDropsIPAccessLists(t *testing.T) {
	setupProxyRelayTestDB(t)

	airport := addRelayAirport(t, models.Airport{Name: "ip-airport", URL: "https://example.com/sub", DisableHWID: true})

	recorder := performSubcriptionFormRequest(t, SubAdd, url.Values{
		"name":        {"proxy-ip"},
		"type":        {"proxy"},
		"AirportID":   {strconv.Itoa(airport.ID)},
		"IPWhitelist": {"10.0.0.0/8"},
		"IPBlacklist": {"192.168.0.0/16"},
	})
	if response := decodeAPIResponse(t, recorder); response.Code != 200 {
		t.Fatalf("add proxy subscription failed: %s", response.Msg)
	}

	var stored models.Subcription
	stored.Name = "proxy-ip"
	if err := stored.Find(); err != nil {
		t.Fatalf("find stored subscription: %v", err)
	}
	if stored.IPWhitelist != "" || stored.IPBlacklist != "" {
		t.Fatalf("proxy subscription must not keep IP lists, got whitelist=%q blacklist=%q", stored.IPWhitelist, stored.IPBlacklist)
	}
}

func TestSubUpdateProxySubscriptionClearsIPAccessLists(t *testing.T) {
	setupProxyRelayTestDB(t)

	airport := addRelayAirport(t, models.Airport{Name: "ip-update-airport", URL: "https://example.com/sub", DisableHWID: true})
	sub := models.Subcription{
		Name:        "proxy-update",
		Type:        models.SubcriptionTypeProxy,
		AirportID:   airport.ID,
		IPWhitelist: "10.0.0.0/8",
		IPBlacklist: "192.168.0.0/16",
	}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}

	recorder := performSubcriptionFormRequest(t, SubUpdate, url.Values{
		"oldname":     {"proxy-update"},
		"name":        {"proxy-update"},
		"type":        {"proxy"},
		"AirportID":   {strconv.Itoa(airport.ID)},
		"IPWhitelist": {"10.0.0.0/8"},
		"IPBlacklist": {"192.168.0.0/16"},
	})
	if response := decodeAPIResponse(t, recorder); response.Code != 200 {
		t.Fatalf("update proxy subscription failed: %s", response.Msg)
	}

	var stored models.Subcription
	if err := database.DB.First(&stored, sub.ID).Error; err != nil {
		t.Fatalf("load stored subscription: %v", err)
	}
	if stored.IPWhitelist != "" || stored.IPBlacklist != "" {
		t.Fatalf("proxy subscription must clear IP lists on update, got whitelist=%q blacklist=%q", stored.IPWhitelist, stored.IPBlacklist)
	}
}

func TestGetClientProxySubscriptionSkipsIPAccessLists(t *testing.T) {
	setupProxyRelayTestDB(t)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		_, _ = w.Write([]byte("proxies: []\n"))
	}))
	defer upstream.Close()

	airport := addRelayAirport(t, models.Airport{Name: "ip-relay-airport", URL: upstream.URL, DisableHWID: true})
	// 历史数据可能残留名单，代理订阅在转发时也必须忽略它们。
	sub := models.Subcription{
		Name:        "proxy-legacy-ip",
		Type:        models.SubcriptionTypeProxy,
		AirportID:   airport.ID,
		IPWhitelist: "10.0.0.0/8",
		IPBlacklist: "203.0.113.0/24",
	}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}
	share := models.SubscriptionShare{
		SubscriptionID: sub.ID,
		Token:          "proxy-ip-token",
		Name:           "proxy share",
		ExpireType:     models.ExpireTypeNever,
		Enabled:        true,
	}
	if err := share.Add(); err != nil {
		t.Fatalf("add share: %v", err)
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/c?token=proxy-ip-token", nil)
	ginContext.Request.RemoteAddr = "203.0.113.10:5678" // 命中黑名单且不在白名单

	GetClient(ginContext)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (IP lists must be ignored for proxy subscriptions)", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); body != "proxies: []\n" {
		t.Fatalf("body = %q", body)
	}
}

func TestNormalizeProxySourceURL(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "https", raw: " https://example.com/sub?token=1 ", want: "https://example.com/sub?token=1"},
		{name: "http", raw: "http://example.com/sub", want: "http://example.com/sub"},
		{name: "empty", raw: "   ", wantErr: true},
		{name: "unsupported scheme", raw: "ftp://example.com/sub", wantErr: true},
		{name: "missing host", raw: "https:///sub", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeProxySourceURL(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got = %q, want %q", got, tc.want)
			}
		})
	}
}
