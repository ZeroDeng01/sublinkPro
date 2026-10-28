package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	socks5service "sublink/services/socks5"
)

func performSocks5JSONRequest(t *testing.T, username, method string, payload any, handler func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequestWithContext(context.Background(), method, "/", bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("username", username)
	handler(ctx)
	return recorder
}

func performSocks5JSONRequestWithParams(t *testing.T, username, method string, payload any, params gin.Params, handler func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequestWithContext(context.Background(), method, "/", bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = params
	ctx.Set("username", username)
	handler(ctx)
	return recorder
}

func TestGetSocks5SettingsRequiresAdmin(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()
	recorder := performBackupRequest(t, "member", GetSocks5Settings)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestUpdateSocks5SettingsPersistsAndMasksPassword(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()
	t.Setenv("SUBLINK_API_ENCRYPTION_KEY", "socks5-test-key-0123456789abcdef0123456789")
	recorder := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"enabled":                      false,
		"listenAddress":                "127.0.0.1",
		"port":                         1080,
		"username":                     "proxy-user",
		"password":                     "proxy-secret",
		"selection":                    "smart",
		"requireAuth":                  true,
		"maxAttempts":                  4,
		"dialTimeoutSeconds":           12,
		"failureCooldownSeconds":       45,
		"specificFallback":             true,
		"maxConnections":               100,
		"maxConnectionsPerClient":      8,
		"idleTimeoutSeconds":           90,
		"maxConnectionDurationSeconds": 3600,
		"healthCheckEnabled":           true,
		"healthCheckIntervalSeconds":   120,
		"healthCheckTimeoutSeconds":    8,
		"candidateGroups":              []string{"premium", "backup"},
		"candidateSources":             []string{"airport-a"},
		"candidateProtocols":           []string{"VLESS", "trojan"},
		"candidateCountries":           []string{"jp", "US"},
		"stickySessionEnabled":         true,
		"stickySessionMode":            "client_ip",
		"stickySessionTtlSeconds":      900,
	}, UpdateSocks5Settings)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if strings.Contains(body, "proxy-secret") || strings.Contains(body, `"password"`) {
		t.Fatalf("response leaked password: %s", body)
	}
	if !strings.Contains(body, `"hasPassword":true`) {
		t.Fatalf("expected masked password metadata: %s", body)
	}
	if !strings.Contains(body, `"candidateGroups":["premium","backup"]`) || !strings.Contains(body, `"candidateCountries":["JP","US"]`) {
		t.Fatalf("response missing normalized candidate pool: %s", body)
	}
	loaded, err := socks5service.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Username != "proxy-user" || loaded.Password != "proxy-secret" || loaded.Enabled || !loaded.RequireAuth {
		t.Fatalf("unexpected stored config: %+v", loaded)
	}
	if loaded.Selection != "smart" || loaded.MaxAttempts != 4 || loaded.DialTimeoutSeconds != 12 || loaded.FailureCooldownSeconds != 45 || !loaded.SpecificFallback {
		t.Fatalf("unexpected phase-two settings: %+v", loaded)
	}
	if loaded.MaxConnections != 100 || loaded.MaxConnectionsPerClient != 8 || loaded.IdleTimeoutSeconds != 90 || loaded.MaxConnectionDurationSeconds != 3600 {
		t.Fatalf("unexpected connection governance settings: %+v", loaded)
	}
	if !loaded.HealthCheckEnabled || loaded.HealthCheckIntervalSeconds != 120 || loaded.HealthCheckTimeoutSeconds != 8 {
		t.Fatalf("unexpected health check settings: %+v", loaded)
	}
	if len(loaded.CandidateGroups) != 2 || loaded.CandidateGroups[0] != "premium" || loaded.CandidateSources[0] != "airport-a" || loaded.CandidateProtocols[0] != "vless" || loaded.CandidateCountries[0] != "JP" {
		t.Fatalf("unexpected candidate pool settings: %+v", loaded)
	}
	if !loaded.StickySessionEnabled || loaded.StickySessionMode != "client_ip" || loaded.StickySessionTTLSeconds != 900 {
		t.Fatalf("unexpected sticky session settings: %+v", loaded)
	}
	if _, err := socks5service.SaveConfig(socks5service.Config{
		Enabled:       false,
		ListenAddress: "127.0.0.1",
		Port:          1080,
		Username:      "proxy-user",
		Selection:     "best",
		RequireAuth:   false,
		ClearPassword: true,
	}); err != nil {
		t.Fatalf("clear SOCKS5 password: %v", err)
	}
	cleared, err := socks5service.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cleared.Password != "" || cleared.RequireAuth {
		t.Fatalf("expected cleared password and disabled auth, got %+v", cleared)
	}
}

func TestUpdateSocks5SettingsPreservesPhaseTwoFieldsForLegacyRequest(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()
	if _, err := socks5service.SaveConfig(socks5service.Config{
		ListenAddress:              "127.0.0.1",
		Port:                       1080,
		Selection:                  "round_robin",
		RequireAuth:                false,
		MaxAttempts:                5,
		DialTimeoutSeconds:         20,
		FailureCooldownSeconds:     60,
		SpecificFallback:           true,
		HealthCheckEnabled:         true,
		HealthCheckIntervalSeconds: 180,
		HealthCheckTimeoutSeconds:  9,
		CandidateGroups:            []string{"premium"},
		CandidateSources:           []string{"airport-a"},
		CandidateProtocols:         []string{"vless"},
		CandidateCountries:         []string{"JP"},
		StickySessionEnabled:       true,
		StickySessionMode:          "client_ip",
		StickySessionTTLSeconds:    1200,
	}); err != nil {
		t.Fatal(err)
	}
	recorder := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"enabled":       false,
		"listenAddress": "127.0.0.1",
		"port":          1081,
		"username":      "legacy",
		"selection":     "best",
		"requireAuth":   false,
	}, UpdateSocks5Settings)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	loaded, err := socks5service.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.MaxAttempts != 5 || loaded.DialTimeoutSeconds != 20 || loaded.FailureCooldownSeconds != 60 || !loaded.SpecificFallback {
		t.Fatalf("legacy request overwrote phase-two fields: %+v", loaded)
	}
	if !loaded.HealthCheckEnabled || loaded.HealthCheckIntervalSeconds != 180 || loaded.HealthCheckTimeoutSeconds != 9 {
		t.Fatalf("legacy request overwrote health check fields: %+v", loaded)
	}
	if !reflect.DeepEqual(loaded.CandidateGroups, []string{"premium"}) || !reflect.DeepEqual(loaded.CandidateSources, []string{"airport-a"}) || !reflect.DeepEqual(loaded.CandidateProtocols, []string{"vless"}) || !reflect.DeepEqual(loaded.CandidateCountries, []string{"JP"}) {
		t.Fatalf("legacy request overwrote candidate pool fields: %+v", loaded)
	}
	if !loaded.StickySessionEnabled || loaded.StickySessionMode != "client_ip" || loaded.StickySessionTTLSeconds != 1200 {
		t.Fatalf("legacy request overwrote sticky session fields: %+v", loaded)
	}

	clearRecorder := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"enabled":              false,
		"listenAddress":        "127.0.0.1",
		"port":                 1081,
		"username":             "legacy",
		"selection":            "best",
		"requireAuth":          false,
		"candidateGroups":      []string{},
		"candidateSources":     []string{},
		"candidateProtocols":   []string{},
		"candidateCountries":   []string{},
		"stickySessionEnabled": false,
	}, UpdateSocks5Settings)
	if clearRecorder.Code != http.StatusOK {
		t.Fatalf("clear status = %d, body = %s", clearRecorder.Code, clearRecorder.Body.String())
	}
	cleared, err := socks5service.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.CandidateGroups) != 0 || len(cleared.CandidateSources) != 0 || len(cleared.CandidateProtocols) != 0 || len(cleared.CandidateCountries) != 0 {
		t.Fatalf("explicit empty candidate pool did not clear settings: %+v", cleared)
	}
	if cleared.StickySessionEnabled {
		t.Fatalf("explicit false did not disable sticky sessions: %+v", cleared)
	}
}

func TestUpdateSocks5SettingsRejectsInvalidRoutingLimits(t *testing.T) {
	setupBackupAPITestDB(t)
	recorder := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"enabled":       false,
		"listenAddress": "127.0.0.1",
		"port":          1080,
		"selection":     "best",
		"requireAuth":   false,
		"maxAttempts":   6,
	}, UpdateSocks5Settings)
	if !strings.Contains(recorder.Body.String(), `"code":500`) || !strings.Contains(recorder.Body.String(), "max attempts") {
		t.Fatalf("invalid routing limits were accepted: %s", recorder.Body.String())
	}
}

func TestGetSocks5StatusAndConnectionsRequireAdmin(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()
	statusRecorder := performSocks5JSONRequest(t, "admin", http.MethodGet, nil, GetSocks5Status)
	if statusRecorder.Code != http.StatusOK {
		t.Fatalf("status endpoint code = %d, body = %s", statusRecorder.Code, statusRecorder.Body.String())
	}
	if !strings.Contains(statusRecorder.Body.String(), `"activeConnections":0`) || !strings.Contains(statusRecorder.Body.String(), `"health"`) {
		t.Fatalf("status response missing gateway stats or health: %s", statusRecorder.Body.String())
	}
	connectionsRecorder := performSocks5JSONRequest(t, "admin", http.MethodGet, nil, GetSocks5Connections)
	if connectionsRecorder.Code != http.StatusOK {
		t.Fatalf("connections endpoint code = %d, body = %s", connectionsRecorder.Code, connectionsRecorder.Body.String())
	}
	if !strings.Contains(connectionsRecorder.Body.String(), `"data":[]`) {
		t.Fatalf("connections response should be empty: %s", connectionsRecorder.Body.String())
	}
	for _, handler := range []func(*gin.Context){GetSocks5Status, GetSocks5Connections} {
		recorder := performSocks5JSONRequest(t, "member", http.MethodGet, nil, handler)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("non-admin status = %d, want %d", recorder.Code, http.StatusForbidden)
		}
	}
}

func TestProbeSocks5HealthRequiresRunningAdminGateway(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()
	member := performSocks5JSONRequest(t, "member", http.MethodPost, nil, ProbeSocks5Health)
	if member.Code != http.StatusForbidden {
		t.Fatalf("member status = %d, want %d", member.Code, http.StatusForbidden)
	}
	admin := performSocks5JSONRequest(t, "admin", http.MethodPost, nil, ProbeSocks5Health)
	if admin.Code != http.StatusOK || !strings.Contains(admin.Body.String(), `"code":500`) || !strings.Contains(admin.Body.String(), "healthProbeUnavailable") {
		t.Fatalf("stopped gateway probe response = %s", admin.Body.String())
	}
}

func TestSocks5RoutingEndpointsRequireAdmin(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()

	routing := performSocks5JSONRequest(t, "admin", http.MethodGet, nil, GetSocks5RoutingSnapshot)
	if routing.Code != http.StatusOK || !strings.Contains(routing.Body.String(), `"items":[]`) || !strings.Contains(routing.Body.String(), `"pageSize":25`) {
		t.Fatalf("unexpected stopped routing response: status=%d body=%s", routing.Code, routing.Body.String())
	}
	reset := performSocks5JSONRequest(t, "admin", http.MethodPost, nil, ResetSocks5RuntimeStats)
	if reset.Code != http.StatusOK || !strings.Contains(reset.Body.String(), "runtimeReset") {
		t.Fatalf("unexpected runtime reset response: status=%d body=%s", reset.Code, reset.Body.String())
	}
	for _, handler := range []func(*gin.Context){GetSocks5RoutingSnapshot, ResetSocks5RuntimeStats} {
		recorder := performSocks5JSONRequest(t, "member", http.MethodGet, nil, handler)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("non-admin routing status = %d, want %d", recorder.Code, http.StatusForbidden)
		}
	}
}

func TestSocks5RoutingProfileCRUD(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()

	create := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"id": "japan", "name": "Japan", "enabled": true, "selection": "smart",
		"maxAttempts": 3, "dialTimeoutSeconds": 15, "failureCooldownSeconds": 30,
		"candidateCountries": []string{"jp"}, "stickySessionEnabled": true, "stickySessionTtlSeconds": 900,
	}, CreateSocks5RoutingProfile)
	if create.Code != http.StatusOK || !strings.Contains(create.Body.String(), `"id":"japan"`) {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}

	list := performSocks5JSONRequest(t, "admin", http.MethodGet, nil, ListSocks5RoutingProfiles)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"id":"default"`) || !strings.Contains(list.Body.String(), `"id":"japan"`) {
		t.Fatalf("list profiles status=%d body=%s", list.Code, list.Body.String())
	}

	update := performSocks5JSONRequestWithParams(t, "admin", http.MethodPut, map[string]any{
		"name": "Japan Premium", "enabled": true, "selection": "round_robin",
		"maxAttempts": 2, "dialTimeoutSeconds": 20, "candidateCountries": []string{"JP"},
		"stickySessionEnabled": true, "stickySessionTtlSeconds": 1200,
	}, gin.Params{{Key: "id", Value: "japan"}}, UpdateSocks5RoutingProfile)
	if update.Code != http.StatusOK || !strings.Contains(update.Body.String(), `"name":"Japan Premium"`) || !strings.Contains(update.Body.String(), `"selection":"round_robin"`) {
		t.Fatalf("update profile status=%d body=%s", update.Code, update.Body.String())
	}

	remove := performSocks5JSONRequestWithParams(t, "admin", http.MethodDelete, nil, gin.Params{{Key: "id", Value: "japan"}}, DeleteSocks5RoutingProfile)
	if remove.Code != http.StatusOK {
		t.Fatalf("delete profile status=%d body=%s", remove.Code, remove.Body.String())
	}
	list = performSocks5JSONRequest(t, "admin", http.MethodGet, nil, ListSocks5RoutingProfiles)
	if strings.Contains(list.Body.String(), `"id":"japan"`) {
		t.Fatalf("deleted profile still listed: %s", list.Body.String())
	}
}

func TestSocks5AccountCRUDDoesNotExposePassword(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()
	t.Setenv("SUBLINK_API_ENCRYPTION_KEY", "socks5-test-key-0123456789abcdef0123456789")

	create := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"id": "alice", "username": "alice", "password": "alice-secret", "profileId": "default", "enabled": true,
	}, CreateSocks5Account)
	if create.Code != http.StatusOK || !strings.Contains(create.Body.String(), `"id":"alice"`) {
		t.Fatalf("create account status=%d body=%s", create.Code, create.Body.String())
	}
	if strings.Contains(create.Body.String(), "alice-secret") || strings.Contains(create.Body.String(), `"password"`) {
		t.Fatalf("account create leaked password: %s", create.Body.String())
	}

	list := performSocks5JSONRequest(t, "admin", http.MethodGet, nil, ListSocks5Accounts)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"hasPassword":true`) || strings.Contains(list.Body.String(), "alice-secret") {
		t.Fatalf("list accounts status=%d body=%s", list.Code, list.Body.String())
	}

	update := performSocks5JSONRequestWithParams(t, "admin", http.MethodPut, map[string]any{
		"username": "alice-renamed", "profileId": "default", "enabled": true,
	}, gin.Params{{Key: "id", Value: "alice"}}, UpdateSocks5Account)
	if update.Code != http.StatusOK || !strings.Contains(update.Body.String(), `"username":"alice-renamed"`) {
		t.Fatalf("update account status=%d body=%s", update.Code, update.Body.String())
	}

	member := performSocks5JSONRequest(t, "member", http.MethodGet, nil, ListSocks5Accounts)
	if member.Code != http.StatusForbidden {
		t.Fatalf("member list accounts status=%d", member.Code)
	}

	remove := performSocks5JSONRequestWithParams(t, "admin", http.MethodDelete, nil, gin.Params{{Key: "id", Value: "alice"}}, DeleteSocks5Account)
	if remove.Code != http.StatusOK {
		t.Fatalf("delete account status=%d body=%s", remove.Code, remove.Body.String())
	}
}

func TestSocks5ListenerCRUD(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()

	create := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"id": "lan", "enabled": true, "listenAddress": "127.0.0.1", "port": 1180,
		"defaultProfileId": "default", "requireAuth": true,
	}, CreateSocks5Listener)
	if create.Code != http.StatusOK || !strings.Contains(create.Body.String(), `"id":"lan"`) {
		t.Fatalf("create listener status=%d body=%s", create.Code, create.Body.String())
	}

	list := performSocks5JSONRequest(t, "admin", http.MethodGet, nil, ListSocks5Listeners)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"id":"default"`) || !strings.Contains(list.Body.String(), `"id":"lan"`) {
		t.Fatalf("list listeners status=%d body=%s", list.Code, list.Body.String())
	}

	update := performSocks5JSONRequestWithParams(t, "admin", http.MethodPut, map[string]any{
		"enabled": false, "listenAddress": "127.0.0.1", "port": 1181,
		"defaultProfileId": "default", "requireAuth": false,
	}, gin.Params{{Key: "id", Value: "lan"}}, UpdateSocks5Listener)
	if update.Code != http.StatusOK || !strings.Contains(update.Body.String(), `"port":1181`) || !strings.Contains(update.Body.String(), `"enabled":false`) {
		t.Fatalf("update listener status=%d body=%s", update.Code, update.Body.String())
	}

	remove := performSocks5JSONRequestWithParams(t, "admin", http.MethodDelete, nil, gin.Params{{Key: "id", Value: "lan"}}, DeleteSocks5Listener)
	if remove.Code != http.StatusOK {
		t.Fatalf("delete listener status=%d body=%s", remove.Code, remove.Body.String())
	}
}

func TestSocks5ProfileDeletionRejectsReferencedAccountOrListener(t *testing.T) {
	setupBackupAPITestDB(t)
	socks5service.DefaultManager().Stop()
	t.Setenv("SUBLINK_API_ENCRYPTION_KEY", "socks5-test-key-0123456789abcdef0123456789")
	create := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"id": "japan", "name": "Japan", "enabled": true, "selection": "best",
	}, CreateSocks5RoutingProfile)
	if create.Code != http.StatusOK {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	account := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"id": "alice", "username": "alice", "password": "secret", "profileId": "japan", "enabled": true,
	}, CreateSocks5Account)
	if account.Code != http.StatusOK {
		t.Fatalf("create account status=%d body=%s", account.Code, account.Body.String())
	}
	remove := performSocks5JSONRequestWithParams(t, "admin", http.MethodDelete, nil, gin.Params{{Key: "id", Value: "japan"}}, DeleteSocks5RoutingProfile)
	if remove.Code != http.StatusOK || !strings.Contains(remove.Body.String(), "referenced") {
		t.Fatalf("referenced profile was deleted: status=%d body=%s", remove.Code, remove.Body.String())
	}
	removeAccount := performSocks5JSONRequestWithParams(t, "admin", http.MethodDelete, nil, gin.Params{{Key: "id", Value: "alice"}}, DeleteSocks5Account)
	if removeAccount.Code != http.StatusOK {
		t.Fatalf("delete account status=%d body=%s", removeAccount.Code, removeAccount.Body.String())
	}
	listener := performSocks5JSONRequest(t, "admin", http.MethodPost, map[string]any{
		"id": "japan-listener", "enabled": false, "listenAddress": "127.0.0.1", "port": 1182,
		"defaultProfileId": "japan", "requireAuth": true,
	}, CreateSocks5Listener)
	if listener.Code != http.StatusOK {
		t.Fatalf("create listener status=%d body=%s", listener.Code, listener.Body.String())
	}
	remove = performSocks5JSONRequestWithParams(t, "admin", http.MethodDelete, nil, gin.Params{{Key: "id", Value: "japan"}}, DeleteSocks5RoutingProfile)
	if remove.Code != http.StatusOK || !strings.Contains(remove.Body.String(), "referenced") {
		t.Fatalf("listener-referenced profile was deleted: status=%d body=%s", remove.Code, remove.Body.String())
	}
}
