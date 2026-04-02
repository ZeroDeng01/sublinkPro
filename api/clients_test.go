package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

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
