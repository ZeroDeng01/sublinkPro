package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const testClashTemplate = `port: 7890
proxies: []
proxy-groups:
  - name: test
    type: select
    proxies: []
`

const testSurgeTemplate = `[General]

[Proxy]

[Proxy Group]
test = select
`

func setupClientsAPITestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized
	oldHook := testGetClientAfterResolveSubscriptionNameHook

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "clients_api_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Subcription{},
		&models.Node{},
		&models.SubcriptionNode{},
		&models.SubcriptionGroup{},
		&models.SubcriptionScript{},
		&models.SubscriptionShare{},
		&models.SubscriptionChainRule{},
		&models.Script{},
	); err != nil {
		t.Fatalf("auto migrate clients api tables: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = true

	if err := models.InitNodeCache(); err != nil {
		t.Fatalf("init node cache: %v", err)
	}
	if err := models.InitSubcriptionCache(); err != nil {
		t.Fatalf("init subcription cache: %v", err)
	}
	if err := models.InitSubscriptionShareCache(); err != nil {
		t.Fatalf("init subscription share cache: %v", err)
	}
	if err := models.InitChainRuleCache(); err != nil {
		t.Fatalf("init chain rule cache: %v", err)
	}
	if err := models.InitScriptCache(); err != nil {
		t.Fatalf("init script cache: %v", err)
	}

	t.Cleanup(func() {
		testGetClientAfterResolveSubscriptionNameHook = oldHook
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		testutil.CloseDB(t, db)
	})
}

func createClientSubscriptionFixture(t *testing.T, clashTemplatePath, surgeTemplatePath, subName, token, linkName string) {
	t.Helper()

	sub := models.Subcription{
		Name:                  subName,
		Config:                `{"clash":"` + clashTemplatePath + `","surge":"` + surgeTemplatePath + `"}`,
		RefreshUsageOnRequest: false,
	}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription %s: %v", subName, err)
	}

	node := models.Node{
		Name:     subName + "-node",
		LinkName: linkName,
		Link:     "ss://YWVzLTEyOC1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#" + linkName,
		Protocol: "ss",
		Source:   "manual",
		Group:    "",
		SourceID: 0,
	}
	if err := node.Add(); err != nil {
		t.Fatalf("add node for %s: %v", subName, err)
	}

	sub.Nodes = []models.Node{node}
	if err := sub.AddNode(); err != nil {
		t.Fatalf("add node relation for %s: %v", subName, err)
	}

	share := models.SubscriptionShare{
		SubscriptionID: sub.ID,
		Token:          token,
		Name:           subName + " share",
		ExpireType:     models.ExpireTypeNever,
		Enabled:        true,
	}
	if err := share.Add(); err != nil {
		t.Fatalf("add share for %s: %v", subName, err)
	}
}

func expireClientSubscriptionShare(t *testing.T, token string) {
	t.Helper()

	share, err := models.GetSubscriptionShareByToken(token)
	if err != nil {
		t.Fatalf("find share %s: %v", token, err)
	}

	expiredAt := time.Now().Add(-time.Hour)
	share.ExpireType = models.ExpireTypeDateTime
	share.ExpireAt = &expiredAt
	if err := share.Update(); err != nil {
		t.Fatalf("expire share %s: %v", token, err)
	}
}

func deleteClientSubscriptionByToken(t *testing.T, token string) {
	t.Helper()

	share, err := models.GetSubscriptionShareByToken(token)
	if err != nil {
		t.Fatalf("find share %s: %v", token, err)
	}

	if err := database.DB.Delete(&models.Subcription{}, share.SubscriptionID).Error; err != nil {
		t.Fatalf("delete subscription %d: %v", share.SubscriptionID, err)
	}
	models.InitSubcriptionCache()
}

func writeTestClashTemplate(t *testing.T) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "clash-template-*.yaml")
	if err != nil {
		t.Fatalf("create clash template: %v", err)
	}
	if _, err := file.WriteString(testClashTemplate); err != nil {
		_ = file.Close()
		t.Fatalf("write clash template: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close clash template: %v", err)
	}

	return file.Name()
}

func writeTestSurgeTemplate(t *testing.T) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "surge-template-*.conf")
	if err != nil {
		t.Fatalf("create surge template: %v", err)
	}
	if _, err := file.WriteString(testSurgeTemplate); err != nil {
		_ = file.Close()
		t.Fatalf("write surge template: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close surge template: %v", err)
	}

	return file.Name()
}

func performClientRequest(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(method, path, nil)
	GetClient(context)
	return recorder
}

func TestGetClientConcurrentRequestsKeepSubscriptionScoped(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "alpha-sub", "alpha-token", "Alpha Node")
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "beta-sub", "beta-token", "Beta Node")

	startSecond := make(chan struct{})
	allowFirstToContinue := make(chan struct{})

	var hookMu sync.Mutex
	callCount := 0
	testGetClientAfterResolveSubscriptionNameHook = func(c *gin.Context) {
		hookMu.Lock()
		callCount++
		currentCall := callCount
		hookMu.Unlock()

		switch currentCall {
		case 1:
			close(startSecond)
			<-allowFirstToContinue
		case 2:
			close(allowFirstToContinue)
		}
	}

	results := make(chan struct {
		body               string
		contentDisposition string
	}, 2)

	go func() {
		recorder := performClientRequest(t, http.MethodGet, "/c/?token=alpha-token&client=v2ray")
		results <- struct {
			body               string
			contentDisposition string
		}{
			body:               recorder.Body.String(),
			contentDisposition: recorder.Header().Get("Content-Disposition"),
		}
	}()

	<-startSecond

	go func() {
		recorder := performClientRequest(t, http.MethodGet, "/c/?token=beta-token&client=v2ray")
		results <- struct {
			body               string
			contentDisposition string
		}{
			body:               recorder.Body.String(),
			contentDisposition: recorder.Header().Get("Content-Disposition"),
		}
	}()

	first := <-results
	second := <-results

	responses := []struct {
		body               string
		contentDisposition string
	}{first, second}

	var alphaBodyCount, betaBodyCount int
	var alphaHeaderCount, betaHeaderCount int
	for _, response := range responses {
		decoded := strings.TrimSpace(utils.Base64Decode(response.body))
		switch {
		case strings.Contains(decoded, "#Alpha Node"):
			alphaBodyCount++
		case strings.Contains(decoded, "#Beta Node"):
			betaBodyCount++
		default:
			t.Fatalf("unexpected decoded response body: %q", decoded)
		}

		switch {
		case strings.Contains(response.contentDisposition, "alpha-sub.txt"):
			alphaHeaderCount++
		case strings.Contains(response.contentDisposition, "beta-sub.txt"):
			betaHeaderCount++
		default:
			t.Fatalf("unexpected content disposition: %q", response.contentDisposition)
		}
	}

	if alphaBodyCount != 1 || betaBodyCount != 1 {
		t.Fatalf("expected one response body per subscription, got alpha=%d beta=%d", alphaBodyCount, betaBodyCount)
	}
	if alphaHeaderCount != 1 || betaHeaderCount != 1 {
		t.Fatalf("expected one content-disposition per subscription, got alpha=%d beta=%d", alphaHeaderCount, betaHeaderCount)
	}
	if callCount != 2 {
		t.Fatalf("expected two hook invocations, got %d", callCount)
	}
}

func TestSubscriptionHandlersRequireRequestScopedSubscriptionName(t *testing.T) {
	setupClientsAPITestDB(t)

	tests := []struct {
		name    string
		path    string
		handler gin.HandlerFunc
	}{
		{name: "v2ray", path: "/c/?client=v2ray", handler: GetV2ray},
		{name: "clash", path: "/c/?client=clash", handler: GetClash},
		{name: "surge", path: "/c/?client=surge", handler: GetSurge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = httptest.NewRequest(http.MethodGet, tt.path, nil)

			tt.handler(context)

			if body := recorder.Body.String(); body != "订阅名为空" {
				t.Fatalf("expected missing subscription name error, got %q", body)
			}
		})
	}
}

func TestGetClientHeadRequestUsesRequestScopedSubscriptionName(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "head-sub", "head-token", "Head Node")

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodHead, "/c/?token=head-token&client=v2ray", nil)

	GetClient(context)

	if body := recorder.Body.String(); body != "" {
		t.Fatalf("expected empty HEAD body, got %q", body)
	}
	if got := recorder.Header().Get("subscription-userinfo"); got == "" {
		t.Fatal("expected subscription-userinfo header to be set")
	}
	if value, ok := context.Get("subname"); !ok || value != "head-sub" {
		t.Fatalf("expected request-scoped subname to be preserved, got %v ok=%v", value, ok)
	}
}

func TestGetClientInvalidShareReturnsSyntheticV2rayNode(t *testing.T) {
	setupClientsAPITestDB(t)

	recorder := performClientRequest(t, http.MethodGet, "/c/?token=missing-token&client=v2ray")
	decoded := strings.TrimSpace(utils.Base64Decode(recorder.Body.String()))

	if !strings.Contains(decoded, "http://placeholder.invalid:80#无效的分享链接") {
		t.Fatalf("expected synthetic invalid-share node in v2ray payload, got %q", decoded)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("expected v2ray content type, got %q", got)
	}
	if got := recorder.Header().Get("subscription-userinfo"); got == "" {
		t.Fatal("expected subscription-userinfo header for invalid v2ray share")
	}
}

func TestGetClientExpiredShareReturnsSyntheticClashNode(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "expired-sub", "expired-token", "Expired Node")
	expireClientSubscriptionShare(t, "expired-token")

	recorder := performClientRequest(t, http.MethodGet, "/c/?token=expired-token&client=clash")
	body := recorder.Body.String()

	if !strings.Contains(body, "name: 订阅已过期") {
		t.Fatalf("expected synthetic expired-share clash node, got %q", body)
	}
	if !strings.Contains(body, "type: http") {
		t.Fatalf("expected synthetic expired-share clash node to use http type, got %q", body)
	}
	if !strings.Contains(body, "server: placeholder.invalid") {
		t.Fatalf("expected synthetic expired-share clash node to use placeholder host, got %q", body)
	}
	if !strings.Contains(body, "- 订阅已过期") {
		t.Fatalf("expected clash proxy group to reference synthetic node, got %q", body)
	}
	if !strings.Contains(recorder.Header().Get("Content-Disposition"), "expired-sub.yaml") {
		t.Fatalf("expected expired share to keep original subscription envelope filename, got %q", recorder.Header().Get("Content-Disposition"))
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("expected clash content type, got %q", got)
	}
}

func TestGetClientInvalidShareHeadReturnsHeadersOnly(t *testing.T) {
	setupClientsAPITestDB(t)

	recorder := performClientRequest(t, http.MethodHead, "/c/?token=missing-token&client=surge")

	if body := recorder.Body.String(); body != "" {
		t.Fatalf("expected empty HEAD body for invalid share, got %q", body)
	}
	if got := recorder.Header().Get("Content-Disposition"); !strings.Contains(got, ".conf") {
		t.Fatalf("expected surge content disposition for invalid share HEAD, got %q", got)
	}
	if got := recorder.Header().Get("subscription-userinfo"); got == "" {
		t.Fatal("expected subscription-userinfo header for invalid share HEAD")
	}
}

func TestGetClientInvalidShareReturnsSyntheticSurgeNode(t *testing.T) {
	setupClientsAPITestDB(t)

	recorder := performClientRequest(t, http.MethodGet, "/c/?token=missing-token&client=surge")
	body := recorder.Body.String()

	if !strings.Contains(body, "[Proxy]") {
		t.Fatalf("expected surge proxy section, got %q", body)
	}
	if !strings.Contains(body, "无效的分享链接 = http, placeholder.invalid, 80") {
		t.Fatalf("expected synthetic invalid-share surge node, got %q", body)
	}
	if !strings.Contains(body, "节点选择 = select, 无效的分享链接") {
		t.Fatalf("expected surge proxy group to reference synthetic node, got %q", body)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("expected surge content type, got %q", got)
	}
}

func TestGetClientMissingSubscriptionReturnsSyntheticNode(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "missing-sub", "missing-sub-token", "Missing Node")
	deleteClientSubscriptionByToken(t, "missing-sub-token")

	recorder := performClientRequest(t, http.MethodGet, "/c/?token=missing-sub-token&client=v2ray")
	decoded := strings.TrimSpace(utils.Base64Decode(recorder.Body.String()))

	if !strings.Contains(decoded, "http://placeholder.invalid:80#订阅不存在") {
		t.Fatalf("expected synthetic missing-subscription node in v2ray payload, got %q", decoded)
	}
	if !strings.Contains(recorder.Header().Get("Content-Disposition"), "%E8%AE%A2%E9%98%85%E4%B8%8D%E5%AD%98%E5%9C%A8.txt") {
		t.Fatalf("expected fallback filename for missing subscription, got %q", recorder.Header().Get("Content-Disposition"))
	}
}

func TestGetClientExpiredSharePreservesSurgeEnvelopeWithSyntheticNode(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "expired-surge-sub", "expired-surge-token", "Expired Surge Node")
	expireClientSubscriptionShare(t, "expired-surge-token")

	recorder := performClientRequest(t, http.MethodGet, "/c/?token=expired-surge-token&client=surge")
	body := recorder.Body.String()

	if !strings.Contains(body, "订阅已过期 = http, placeholder.invalid, 80") {
		t.Fatalf("expected synthetic expired-share surge node, got %q", body)
	}
	if !strings.Contains(recorder.Header().Get("Content-Disposition"), "expired-surge-sub.conf") {
		t.Fatalf("expected expired surge share to keep original subscription envelope filename, got %q", recorder.Header().Get("Content-Disposition"))
	}
}

func TestGetClientExpiredSharePreservesV2rayEnvelopeWithSyntheticNode(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "expired-v2ray-sub", "expired-v2ray-token", "Expired V2ray Node")
	expireClientSubscriptionShare(t, "expired-v2ray-token")

	recorder := performClientRequest(t, http.MethodGet, "/c/?token=expired-v2ray-token&client=v2ray")
	decoded := strings.TrimSpace(utils.Base64Decode(recorder.Body.String()))

	if !strings.Contains(decoded, "http://placeholder.invalid:80#订阅已过期") {
		t.Fatalf("expected synthetic expired-share v2ray node, got %q", decoded)
	}
	if !strings.Contains(recorder.Header().Get("Content-Disposition"), "expired-v2ray-sub.txt") {
		t.Fatalf("expected expired v2ray share to keep original subscription envelope filename, got %q", recorder.Header().Get("Content-Disposition"))
	}
}

func TestGetClientConcurrentClashRequestsKeepSubscriptionScoped(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "alpha-sub", "alpha-token", "Alpha Node")
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "beta-sub", "beta-token", "Beta Node")

	startSecond := make(chan struct{})
	allowFirstToContinue := make(chan struct{})

	var hookMu sync.Mutex
	callCount := 0
	testGetClientAfterResolveSubscriptionNameHook = func(c *gin.Context) {
		hookMu.Lock()
		callCount++
		currentCall := callCount
		hookMu.Unlock()

		switch currentCall {
		case 1:
			close(startSecond)
			<-allowFirstToContinue
		case 2:
			close(allowFirstToContinue)
		}
	}

	results := make(chan struct {
		body               string
		contentDisposition string
	}, 2)

	go func() {
		recorder := performClientRequest(t, http.MethodGet, "/c/?token=alpha-token&client=clash")
		results <- struct {
			body               string
			contentDisposition string
		}{
			body:               recorder.Body.String(),
			contentDisposition: recorder.Header().Get("Content-Disposition"),
		}
	}()

	<-startSecond

	go func() {
		recorder := performClientRequest(t, http.MethodGet, "/c/?token=beta-token&client=clash")
		results <- struct {
			body               string
			contentDisposition string
		}{
			body:               recorder.Body.String(),
			contentDisposition: recorder.Header().Get("Content-Disposition"),
		}
	}()

	first := <-results
	second := <-results

	responses := []struct {
		body               string
		contentDisposition string
	}{first, second}

	var alphaBodyCount, betaBodyCount int
	var alphaHeaderCount, betaHeaderCount int
	for _, response := range responses {
		switch {
		case strings.Contains(response.body, "name: Alpha Node"):
			alphaBodyCount++
		case strings.Contains(response.body, "name: Beta Node"):
			betaBodyCount++
		default:
			t.Fatalf("unexpected clash response body: %q", response.body)
		}

		switch {
		case strings.Contains(response.contentDisposition, "alpha-sub.yaml"):
			alphaHeaderCount++
		case strings.Contains(response.contentDisposition, "beta-sub.yaml"):
			betaHeaderCount++
		default:
			t.Fatalf("unexpected content disposition: %q", response.contentDisposition)
		}
	}

	if alphaBodyCount != 1 || betaBodyCount != 1 {
		t.Fatalf("expected one clash response body per subscription, got alpha=%d beta=%d", alphaBodyCount, betaBodyCount)
	}
	if alphaHeaderCount != 1 || betaHeaderCount != 1 {
		t.Fatalf("expected one clash content-disposition per subscription, got alpha=%d beta=%d", alphaHeaderCount, betaHeaderCount)
	}
	if callCount != 2 {
		t.Fatalf("expected two hook invocations, got %d", callCount)
	}
}

func TestGetClientConcurrentSurgeRequestsKeepSubscriptionScoped(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "alpha-sub", "alpha-token", "Alpha Node")
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "beta-sub", "beta-token", "Beta Node")

	startSecond := make(chan struct{})
	allowFirstToContinue := make(chan struct{})

	var hookMu sync.Mutex
	callCount := 0
	testGetClientAfterResolveSubscriptionNameHook = func(c *gin.Context) {
		hookMu.Lock()
		callCount++
		currentCall := callCount
		hookMu.Unlock()

		switch currentCall {
		case 1:
			close(startSecond)
			<-allowFirstToContinue
		case 2:
			close(allowFirstToContinue)
		}
	}

	results := make(chan struct {
		body               string
		contentDisposition string
	}, 2)

	go func() {
		recorder := performClientRequest(t, http.MethodGet, "/c/?token=alpha-token&client=surge")
		results <- struct {
			body               string
			contentDisposition string
		}{
			body:               recorder.Body.String(),
			contentDisposition: recorder.Header().Get("Content-Disposition"),
		}
	}()

	<-startSecond

	go func() {
		recorder := performClientRequest(t, http.MethodGet, "/c/?token=beta-token&client=surge")
		results <- struct {
			body               string
			contentDisposition string
		}{
			body:               recorder.Body.String(),
			contentDisposition: recorder.Header().Get("Content-Disposition"),
		}
	}()

	first := <-results
	second := <-results

	responses := []struct {
		body               string
		contentDisposition string
	}{first, second}

	var alphaBodyCount, betaBodyCount int
	var alphaHeaderCount, betaHeaderCount int
	for _, response := range responses {
		switch {
		case strings.Contains(response.body, "Alpha Node = ss"):
			alphaBodyCount++
		case strings.Contains(response.body, "Beta Node = ss"):
			betaBodyCount++
		default:
			t.Fatalf("unexpected surge response body: %q", response.body)
		}

		switch {
		case strings.Contains(response.contentDisposition, "alpha-sub.conf"):
			alphaHeaderCount++
		case strings.Contains(response.contentDisposition, "beta-sub.conf"):
			betaHeaderCount++
		default:
			t.Fatalf("unexpected content disposition: %q", response.contentDisposition)
		}
	}

	if alphaBodyCount != 1 || betaBodyCount != 1 {
		t.Fatalf("expected one surge response body per subscription, got alpha=%d beta=%d", alphaBodyCount, betaBodyCount)
	}
	if alphaHeaderCount != 1 || betaHeaderCount != 1 {
		t.Fatalf("expected one surge content-disposition per subscription, got alpha=%d beta=%d", alphaHeaderCount, betaHeaderCount)
	}
	if callCount != 2 {
		t.Fatalf("expected two hook invocations, got %d", callCount)
	}
}
