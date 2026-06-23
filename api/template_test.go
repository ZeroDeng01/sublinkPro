package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sublink/cache"
	"sublink/config"
	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
	"sublink/services/ai"

	"github.com/gin-gonic/gin"
)

func setupTemplateAPITestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized
	oldBaseTemplateDir := baseTemplateDir
	oldConfig := *config.Get()
	oldSessionStore := templateEditSessions

	db := testutil.OpenMemoryDB(t, "template_api_test")
	if err := db.AutoMigrate(&models.Template{}, &models.User{}, &models.SystemSetting{}); err != nil {
		t.Fatalf("auto migrate templates: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	config.UpdateConfig(func(cfg *config.AppConfig) {
		cfg.APIEncryptionKey = "test-api-encryption-key-0123456789abcd"
		cfg.JwtSecret = "test-jwt-secret"
	})
	if err := models.InitTemplateCache(); err != nil {
		t.Fatalf("init template cache: %v", err)
	}
	if err := models.InitUserCache(); err != nil {
		t.Fatalf("init user cache: %v", err)
	}
	if err := models.InitSettingCache(); err != nil {
		t.Fatalf("init setting cache: %v", err)
	}
	cache.InvalidateAllTemplateContent()
	templateEditSessions = ai.NewTemplateEditSessionStore()

	t.Cleanup(func() {
		cache.InvalidateAllTemplateContent()
		baseTemplateDir = oldBaseTemplateDir
		templateEditSessions = oldSessionStore
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		config.UpdateConfig(func(cfg *config.AppConfig) { *cfg = oldConfig })
		if oldDB != nil {
			_ = models.InitTemplateCache()
			_ = models.InitUserCache()
			_ = models.InitSettingCache()
		}
		testutil.CloseDB(t, db)
	})
}

func TestGetTempSInfersSurgeCategoryWithoutMetadata(t *testing.T) {
	setupTemplateAPITestDB(t)

	templateDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(templateDir, "surge.conf"), []byte("[General]\n"), 0600); err != nil {
		t.Fatalf("write surge template: %v", err)
	}
	baseTemplateDir = templateDir

	recorder := performJSONRequest(t, GetTempS, http.MethodGet, nil)
	response := decodeAPIResponse(t, recorder)
	if response.Code != 200 {
		t.Fatalf("expected response code 200, got %d", response.Code)
	}

	var templates []Temp
	if err := json.Unmarshal(response.Data, &templates); err != nil {
		t.Fatalf("unmarshal template list: %v", err)
	}
	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}
	if templates[0].File != "surge.conf" {
		t.Fatalf("expected surge.conf, got %q", templates[0].File)
	}
	if templates[0].Category != "surge" {
		t.Fatalf("expected inferred category surge, got %q", templates[0].Category)
	}
}

func TestTemplateEditSessionStreamPreviewReady(t *testing.T) {
	setupTemplateAPITestDB(t)
	baseText := "proxies: []\nproxy-groups: []\nrules:\n  - MATCH,DIRECT\n"
	templateDir := t.TempDir()
	writeTemplateFileForTest(t, templateDir, "clash.yaml", baseText)
	baseTemplateDir = templateDir

	server := newTemplateAIChatServer(t, `{"summary":"Route final traffic","warnings":[],"operations":[{"op":"replace","oldString":"  - MATCH,DIRECT","newString":"  - MATCH,Proxy","description":"route final traffic"}]}`)
	defer server.Close()
	createTemplateAIUser(t, server.URL)

	recorder := performTemplateHandlerRequest(t, StartTemplateAIEditSessionStream, http.MethodPost, "/api/v1/template/ai/edit-sessions/stream", TemplateAIGenerateRequest{
		Filename:    "clash.yaml",
		Category:    "clash",
		CurrentText: baseText,
		UserPrompt:  "Route final traffic through Proxy",
	}, "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	events := parseTemplateSSEEvents(t, recorder.Body.String())
	for _, forbidden := range events["template.final"] {
		if forbidden != nil {
			t.Fatal("stream must not emit legacy template.final")
		}
	}
	for _, eventName := range []string{"template.edit.session.created", "template.edit.model.delta", "template.edit.operations.ready", "template.edit.preview.validating", "template.edit.preview.ready", "template.edit.completed"} {
		if len(events[eventName]) == 0 {
			t.Fatalf("expected SSE event %s in stream %v", eventName, eventNames(events))
		}
	}
	ready := events["template.edit.preview.ready"][0]
	sessionID, _ := ready["sessionId"].(string)
	if sessionID == "" {
		t.Fatalf("expected sessionId in preview payload: %#v", ready)
	}
	if ready["baseHash"] != ai.BuildRevisionHash(baseText) {
		t.Fatalf("unexpected baseHash %#v", ready["baseHash"])
	}
	if ready["candidateText"] != "proxies: []\nproxy-groups: []\nrules:\n  - MATCH,Proxy\n" {
		t.Fatalf("unexpected candidate text: %#v", ready["candidateText"])
	}
	if ready["candidateHash"] == "" || ready["operations"] == nil || ready["validation"] == nil || ready["expiresAt"] == "" {
		t.Fatalf("preview payload missing required fields: %#v", ready)
	}

	getRecorder := performTemplateHandlerRequest(t, GetTemplateAIEditSession, http.MethodGet, "/api/v1/template/ai/edit-sessions/"+sessionID, nil, sessionID)
	response := decodeAPIResponse(t, getRecorder)
	if response.Code != 200 {
		t.Fatalf("expected get session response 200, got %d: %s", response.Code, getRecorder.Body.String())
	}
	var sessionPayload map[string]any
	if err := json.Unmarshal(response.Data, &sessionPayload); err != nil {
		t.Fatalf("unmarshal session payload: %v", err)
	}
	if sessionPayload["candidateText"] != ready["candidateText"] {
		t.Fatalf("GET session returned different candidate: %#v", sessionPayload)
	}
}

func TestTemplateEditSessionStreamUsesMockProviderFixtures(t *testing.T) {
	setupTemplateAPITestDB(t)
	baseText := templateEditMockQAContentForTest()
	templateDir := t.TempDir()
	writeTemplateFileForTest(t, templateDir, "qa-mock.yaml", baseText)
	baseTemplateDir = templateDir
	createTemplateAIUser(t, ai.TemplateEditMockBaseURL)

	tests := []struct {
		name         string
		prompt       string
		ruleSource   string
		wantEvent    string
		wantCode     string
		wantContains string
		wantAbsent   string
		wantWarning  bool
	}{
		{name: "replace", prompt: "QA_REPLACE_DNS_COMMENT", wantEvent: "template.edit.preview.ready", wantContains: "# QA_DNS_COMMENT: use deterministic resolver"},
		{name: "insert", prompt: "QA_INSERT_PROXY_GROUP", wantEvent: "template.edit.preview.ready", wantContains: "name: QA Inserted"},
		{name: "delete", prompt: "QA_DELETE_TEST_COMMENT", wantEvent: "template.edit.preview.ready", wantAbsent: "# QA_DELETE_ME: temporary test comment"},
		{name: "duplicate ambiguity", prompt: "QA_DUPLICATE_MATCH", wantEvent: "template.edit.error", wantCode: string(ai.PatchAmbiguousMatch)},
		{name: "delete all duplicate matches", prompt: "QA_DELETE_ALL_DUPLICATE_MATCHES", wantEvent: "template.edit.preview.ready", wantAbsent: "  - DOMAIN-SUFFIX,duplicate.example,DIRECT"},
		{name: "protected token validation", prompt: "QA_REMOVE_PROTECTED_TOKEN", wantEvent: "template.edit.error", wantCode: string(ai.TemplateEditValidationFailed)},
		{name: "rule-source warning", prompt: "QA_RULESOURCE_WARNING", ruleSource: "https://example.test/rules.list", wantEvent: "template.edit.preview.ready", wantContains: "# QA_RULESOURCE_MARKER: changed", wantWarning: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := performTemplateHandlerRequest(t, StartTemplateAIEditSessionStream, http.MethodPost, "/api/v1/template/ai/edit-sessions/stream", TemplateAIGenerateRequest{
				Filename:    "qa-mock.yaml",
				Category:    "clash",
				CurrentText: baseText,
				UserPrompt:  tt.prompt,
				RuleSource:  tt.ruleSource,
			}, "")
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected HTTP 200, got %d: %s", recorder.Code, recorder.Body.String())
			}
			events := parseTemplateSSEEvents(t, recorder.Body.String())
			payloads := events[tt.wantEvent]
			if len(payloads) == 0 {
				t.Fatalf("expected event %s in stream %v", tt.wantEvent, eventNames(events))
			}
			payload := payloads[len(payloads)-1]
			if tt.wantCode != "" {
				if payload["code"] != tt.wantCode {
					t.Fatalf("expected code %s, got %#v", tt.wantCode, payload)
				}
				return
			}
			candidateText, _ := payload["candidateText"].(string)
			if tt.wantContains != "" && !strings.Contains(candidateText, tt.wantContains) {
				t.Fatalf("candidate missing %q: %#v", tt.wantContains, payload)
			}
			if tt.wantAbsent != "" && strings.Contains(candidateText, tt.wantAbsent) {
				t.Fatalf("candidate still contains %q: %s", tt.wantAbsent, candidateText)
			}
			if tt.wantWarning && len(events["template.edit.warning"]) == 0 {
				t.Fatalf("expected real warning event, got %v", eventNames(events))
			}
		})
	}
}

func TestTemplateEditSessionStreamRejectsMockProviderInProduction(t *testing.T) {
	setupTemplateAPITestDB(t)
	oldAppEnv := os.Getenv("APP_ENV")
	t.Cleanup(func() { _ = os.Setenv("APP_ENV", oldAppEnv) })
	if err := os.Setenv("APP_ENV", "production"); err != nil {
		t.Fatalf("set APP_ENV: %v", err)
	}
	baseText := templateEditMockQAContentForTest()
	templateDir := t.TempDir()
	writeTemplateFileForTest(t, templateDir, "qa-mock.yaml", baseText)
	baseTemplateDir = templateDir
	createTemplateAIUser(t, ai.TemplateEditMockBaseURL)

	recorder := performTemplateHandlerRequest(t, StartTemplateAIEditSessionStream, http.MethodPost, "/api/v1/template/ai/edit-sessions/stream", TemplateAIGenerateRequest{
		Filename:    "qa-mock.yaml",
		Category:    "clash",
		CurrentText: baseText,
		UserPrompt:  "QA_REPLACE_DNS_COMMENT",
	}, "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected SSE HTTP 200 with structured error event, got %d: %s", recorder.Code, recorder.Body.String())
	}
	events := parseTemplateSSEEvents(t, recorder.Body.String())
	if len(events["template.edit.error"]) == 0 {
		t.Fatalf("expected template.edit.error event, got %v", eventNames(events))
	}
	errorPayload := events["template.edit.error"][0]
	if errorPayload["code"] != string(ai.TemplateEditMockProviderUnavailable) {
		t.Fatalf("expected production mock rejection code, got %#v", errorPayload)
	}
}

func TestTemplateEditSessionRuntimeStoreStartsCleanupOnce(t *testing.T) {
	oldStarter := templateEditSessionCleanupStarter
	defer func() { templateEditSessionCleanupStarter = oldStarter }()

	started := 0
	var startedStore *ai.TemplateEditSessionStore
	templateEditSessionCleanupStarter = func(store *ai.TemplateEditSessionStore) {
		started++
		startedStore = store
	}

	store := newTemplateEditSessionStoreRuntime()
	if started != 1 {
		t.Fatalf("expected cleanup starter to run once, got %d", started)
	}
	if startedStore != store {
		t.Fatalf("cleanup starter received a different store")
	}
}

func TestTemplateAICandidateLegacyRouteRemoved(t *testing.T) {
	setupTemplateAPITestDB(t)
	legacyRoutes := []string{
		"/api/v1/template/ai/generate",
		"/api/v1/template/ai/generate-stream",
		"/api/v1/template/ai/validate",
		"/api/v1/template/ai/apply",
	}
	for _, route := range legacyRoutes {
		t.Run(route, func(t *testing.T) {
			recorder := performTemplateHandlerRequest(t, TemplateAILegacyRemoved, http.MethodPost, route, map[string]any{"candidateText": "full template"}, "")
			if recorder.Code != http.StatusGone {
				t.Fatalf("expected HTTP 410, got %d", recorder.Code)
			}
			var payload map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
				t.Fatalf("unmarshal legacy response: %v", err)
			}
			if payload["code"] != templateEditLegacyRemovedCode {
				t.Fatalf("expected legacy removed code, got %#v", payload)
			}
		})
	}
}

func TestTemplateEditSessionAcceptDoesNotPersist(t *testing.T) {
	setupTemplateAPITestDB(t)
	baseText := "proxies: []\nproxy-groups: []\nrules:\n  - MATCH,DIRECT\n"
	templateDir := t.TempDir()
	writeTemplateFileForTest(t, templateDir, "clash.yaml", baseText)
	baseTemplateDir = templateDir
	user := createTemplateAIUser(t, "http://127.0.0.1")
	session := createPreviewSessionForTest(t, user, "clash.yaml", baseText, "https://example.test/rules.list")
	if len(session.Validation.Warnings) == 0 || session.WarningFingerprint == "" {
		t.Fatalf("test requires warning metadata, got validation=%#v fingerprint=%q", session.Validation, session.WarningFingerprint)
	}

	accepted := performTemplateHandlerRequest(t, AcceptTemplateAIEditSession, http.MethodPost, "/api/v1/template/ai/edit-sessions/"+session.SessionID+"/accept", nil, session.SessionID)
	acceptedResponse := decodeAPIResponse(t, accepted)
	if acceptedResponse.Code != 200 {
		t.Fatalf("expected warning-only accept response 200 without confirmation payload, got %d: %s", acceptedResponse.Code, accepted.Body.String())
	}
	var acceptedPayload map[string]any
	if err := json.Unmarshal(acceptedResponse.Data, &acceptedPayload); err != nil {
		t.Fatalf("unmarshal accept payload: %v", err)
	}
	if acceptedPayload["candidateText"] == baseText || acceptedPayload["candidateText"] == "" {
		t.Fatalf("expected changed candidate text in accept response: %#v", acceptedPayload)
	}
	diskBytes, err := os.ReadFile(filepath.Join(templateDir, "clash.yaml"))
	if err != nil {
		t.Fatalf("read template after accept: %v", err)
	}
	if string(diskBytes) != baseText {
		t.Fatalf("accept must not persist candidate to disk, got %q", string(diskBytes))
	}
}

func TestTemplateEditSessionAcceptUsesCurrentTextBaseProof(t *testing.T) {
	setupTemplateAPITestDB(t)
	baseText := "proxies: []\nproxy-groups: []\nrules:\n  - MATCH,DIRECT\n"
	templateDir := t.TempDir()
	writeTemplateFileForTest(t, templateDir, "clash.yaml", baseText)
	baseTemplateDir = templateDir
	user := createTemplateAIUser(t, "http://127.0.0.1")
	firstSession := createPreviewSessionForTest(t, user, "clash.yaml", baseText, "")

	firstAccept := performTemplateHandlerRequest(t, AcceptTemplateAIEditSession, http.MethodPost, "/api/v1/template/ai/edit-sessions/"+firstSession.SessionID+"/accept", nil, firstSession.SessionID)
	firstResponse := decodeAPIResponse(t, firstAccept)
	if firstResponse.Code != 200 {
		t.Fatalf("expected first accept to succeed, got %d: %s", firstResponse.Code, firstAccept.Body.String())
	}
	var firstPayload map[string]any
	if err := json.Unmarshal(firstResponse.Data, &firstPayload); err != nil {
		t.Fatalf("unmarshal first accept payload: %v", err)
	}
	currentEditorText, _ := firstPayload["candidateText"].(string)
	if currentEditorText == "" || currentEditorText == baseText {
		t.Fatalf("expected editor text from first accept, got %#v", firstPayload)
	}
	if diskBytes, err := os.ReadFile(filepath.Join(templateDir, "clash.yaml")); err != nil {
		t.Fatalf("read disk after first accept: %v", err)
	} else if string(diskBytes) != baseText {
		t.Fatalf("first accept must not persist to disk, got %q", string(diskBytes))
	}

	secondPreview, err := ai.BuildTemplateEditPreviewCandidate(ai.TemplateEditPreviewInput{Category: "clash", BaseText: currentEditorText, BaseHash: ai.BuildRevisionHash(currentEditorText), Operations: []ai.TemplateEditOperation{{Op: ai.TemplateEditOperationReplace, OldString: "  - MATCH,Proxy", NewString: "  - MATCH,Direct", Description: "second preview"}}})
	if err != nil {
		t.Fatalf("build second preview: %v", err)
	}
	secondSession, err := templateEditSessions.Create(ai.TemplateEditSessionCreateInput{
		OwnerKey:           ai.TemplateEditSessionOwnerKey(user),
		Filename:           "clash.yaml",
		Category:           "clash",
		BaseHash:           secondPreview.BaseHash,
		BaseText:           currentEditorText,
		CandidateText:      secondPreview.CandidateText,
		Operations:         secondPreview.Operations,
		Validation:         secondPreview.Validation,
		WarningFingerprint: secondPreview.WarningFingerprint,
		Status:             ai.TemplateEditSessionPreviewReady,
	})
	if err != nil {
		t.Fatalf("create second session: %v", err)
	}
	secondAccept := performTemplateHandlerRequest(t, AcceptTemplateAIEditSession, http.MethodPost, "/api/v1/template/ai/edit-sessions/"+secondSession.SessionID+"/accept", TemplateEditSessionAcceptRequest{CurrentText: currentEditorText}, secondSession.SessionID)
	secondResponse := decodeAPIResponse(t, secondAccept)
	if secondResponse.Code != 200 {
		t.Fatalf("expected second accept with currentText to succeed, got %d: %s", secondResponse.Code, secondAccept.Body.String())
	}
	if diskBytes, err := os.ReadFile(filepath.Join(templateDir, "clash.yaml")); err != nil {
		t.Fatalf("read disk after second accept: %v", err)
	} else if string(diskBytes) != baseText {
		t.Fatalf("second accept must not persist to disk, got %q", string(diskBytes))
	}
}

func TestTemplateEditSessionStaleBase(t *testing.T) {
	setupTemplateAPITestDB(t)
	baseText := "proxies: []\nproxy-groups: []\nrules:\n  - MATCH,DIRECT\n"
	templateDir := t.TempDir()
	writeTemplateFileForTest(t, templateDir, "clash.yaml", baseText)
	baseTemplateDir = templateDir
	user := createTemplateAIUser(t, "http://127.0.0.1")
	session := createPreviewSessionForTest(t, user, "clash.yaml", baseText, "")
	writeTemplateFileForTest(t, templateDir, "clash.yaml", strings.Replace(baseText, "DIRECT", "REJECT", 1))

	recorder := performTemplateHandlerRequest(t, AcceptTemplateAIEditSession, http.MethodPost, "/api/v1/template/ai/edit-sessions/"+session.SessionID+"/accept", nil, session.SessionID)
	response := decodeAPIResponse(t, recorder)
	if response.Code == 200 {
		t.Fatalf("expected stale base failure: %s", recorder.Body.String())
	}
	assertTemplateAPIErrorCode(t, response, templateEditStaleBaseCode)
}

func performTemplateHandlerRequest(t *testing.T, handler gin.HandlerFunc, method string, target string, body any, sessionID string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	var requestBody []byte
	if body != nil {
		var err error
		requestBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), method, target, bytes.NewReader(requestBody))
	ginContext.Request.Header.Set("Content-Type", "application/json")
	ginContext.Set("username", "admin")
	if sessionID != "" {
		ginContext.Params = gin.Params{{Key: "sessionId", Value: sessionID}}
	}
	handler(ginContext)
	return recorder
}

func createTemplateAIUser(t *testing.T, baseURL string) *models.User {
	t.Helper()
	user := &models.User{Username: "admin", Password: "123456", Role: "admin", Nickname: "管理员"}
	if err := user.Create(); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := user.UpdateAISettings(models.UserAISettings{Enabled: true, BaseURL: baseURL, Model: "gpt-test", RequestType: models.SystemAIRequestTypeChatCompletions, RawAPIKey: "sk-test-1234567890", Temperature: 0.2, MaxTokens: 512}); err != nil {
		t.Fatalf("update AI settings: %v", err)
	}
	return user
}

func createPreviewSessionForTest(t *testing.T, user *models.User, filename string, baseText string, ruleSource string) ai.TemplateEditSession {
	t.Helper()
	operations := []ai.TemplateEditOperation{{Op: ai.TemplateEditOperationReplace, OldString: "  - MATCH,DIRECT", NewString: "  - MATCH,Proxy", Description: "route final traffic"}}
	preview, err := ai.BuildTemplateEditPreviewCandidate(ai.TemplateEditPreviewInput{Category: "clash", BaseText: baseText, BaseHash: ai.BuildRevisionHash(baseText), Operations: operations, RuleSource: ruleSource})
	if err != nil {
		t.Fatalf("build preview: %v", err)
	}
	session, err := templateEditSessions.Create(ai.TemplateEditSessionCreateInput{
		OwnerKey:           ai.TemplateEditSessionOwnerKey(user),
		Filename:           filename,
		Category:           "clash",
		BaseHash:           preview.BaseHash,
		BaseText:           baseText,
		CandidateText:      preview.CandidateText,
		Operations:         preview.Operations,
		Validation:         preview.Validation,
		WarningFingerprint: preview.WarningFingerprint,
		Status:             ai.TemplateEditSessionPreviewReady,
	})
	if err != nil {
		t.Fatalf("create preview session: %v", err)
	}
	return session
}

func newTemplateAIChatServer(t *testing.T, modelOutput string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected AI path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":" + strconvQuote(modelOutput) + "}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
}

func parseTemplateSSEEvents(t *testing.T, body string) map[string][]map[string]any {
	t.Helper()
	events := map[string][]map[string]any{}
	scanner := bufio.NewScanner(strings.NewReader(body))
	currentEvent := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event: "))
			continue
		}
		if strings.HasPrefix(line, "data: ") {
			var payload map[string]any
			if err := json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data: "))), &payload); err != nil {
				t.Fatalf("unmarshal SSE data %q: %v", line, err)
			}
			events[currentEvent] = append(events[currentEvent], payload)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan SSE body: %v", err)
	}
	return events
}

func eventNames(events map[string][]map[string]any) []string {
	names := make([]string, 0, len(events))
	for name := range events {
		names = append(names, name)
	}
	return names
}

func writeTemplateFileForTest(t *testing.T, dir string, name string, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
		t.Fatalf("write template %s: %v", name, err)
	}
}

func templateEditMockQAContentForTest() string {
	return strings.Join([]string{
		"# QA_DNS_COMMENT: use default resolver",
		"# QA_DELETE_ME: temporary test comment",
		"# QA_RULESOURCE_MARKER: original",
		"proxies: []",
		"proxy-groups:",
		"  - name: Auto",
		"    type: select",
		"    proxies:",
		"      - __ALL_PROXIES__",
		"rules:",
		"  - DOMAIN-SUFFIX,duplicate.example,DIRECT",
		"  - DOMAIN-SUFFIX,duplicate.example,DIRECT",
		"  - MATCH,DIRECT",
		"",
	}, "\n")
}

func assertTemplateAPIErrorCode(t *testing.T, response apiJSONResponse, want string) {
	t.Helper()
	var payload struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("unmarshal error payload: %v", err)
	}
	if payload.Code != want {
		t.Fatalf("expected error code %s, got %q", want, payload.Code)
	}
}

func strconvQuote(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}
