package api

import (
	"context"
	"encoding/json"
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
	"sublink/node/protocol"
	"sublink/utils"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
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

	db := testutil.OpenMemoryDB(t, "clients_api_test")

	if err := db.AutoMigrate(
		&models.Subcription{},
		&models.Node{},
		&models.SubcriptionNode{},
		&models.SubcriptionGroup{},
		&models.SubcriptionAirport{},
		&models.SubcriptionScript{},
		&models.SubscriptionShare{},
		&models.SubscriptionChainRule{},
		&models.Script{},
		&models.SystemSetting{},
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
	if err := models.InitSettingCache(); err != nil {
		t.Fatalf("init setting cache: %v", err)
	}

	t.Cleanup(func() {
		// 等待在飞的异步访问统计写入完成，避免其在拆库后解引用 nil 的 database.DB。
		models.WaitForPendingAccessRecords()
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

	share.ExpireType = models.ExpireTypeDateTime
	share.ExpireAt = new(time.Now().Add(-time.Hour))
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
	if err := models.InitSubcriptionCache(); err != nil {
		t.Fatalf("refresh subscription cache: %v", err)
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

func writeTestSurgeTemplateWithManagedConfig(t *testing.T) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "surge-managed-template-*.conf")
	if err != nil {
		t.Fatalf("create surge managed template: %v", err)
	}
	content := "#!MANAGED-CONFIG https://old.example/sub interval=86400 strict=false\n" + testSurgeTemplate
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		t.Fatalf("write surge managed template: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close surge managed template: %v", err)
	}

	return file.Name()
}

func saveSubStoreSettings(t *testing.T, baseURL string, allowedTargets []string) {
	t.Helper()
	if err := models.SetSetting("substore_enabled", "true"); err != nil {
		t.Fatalf("set substore enabled: %v", err)
	}
	if err := models.SetSetting("substore_base_url", baseURL); err != nil {
		t.Fatalf("set substore base url: %v", err)
	}
	if err := models.SetSetting("substore_timeout_seconds", "1"); err != nil {
		t.Fatalf("set substore timeout: %v", err)
	}
	if err := models.SetSetting("substore_allowed_targets", strings.Join(allowedTargets, ",")); err != nil {
		t.Fatalf("set substore targets: %v", err)
	}
	if err := models.SetSetting("substore_max_response_bytes", "1048576"); err != nil {
		t.Fatalf("set substore max response bytes: %v", err)
	}
}

func performClientRequest(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), method, path, nil)
	GetClient(ginContext)
	return recorder
}

// clashProxyNamesFromBody 从 Clash YAML 输出中提取最终 proxy.name。
// 这些名称是客户端真实引用的结果，比只测内部命名函数更接近导出行为。
func clashProxyNamesFromBody(t *testing.T, body []byte) []string {
	t.Helper()

	var config struct {
		Proxies []struct {
			Name string `yaml:"name"`
		} `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(body, &config); err != nil {
		t.Fatalf("parse clash output: %v\n%s", err, string(body))
	}

	names := make([]string, 0, len(config.Proxies))
	for _, proxy := range config.Proxies {
		names = append(names, proxy.Name)
	}
	return names
}

// surgeProxyNamesFromBody 从 Surge [Proxy] section 提取代理名称。
// 只读取等号左侧名称，避免把 [Proxy Group] 中的组名误判为节点代理。
func surgeProxyNamesFromBody(body string) []string {
	lines := strings.Split(body, "\n")
	inProxySection := false
	names := []string{}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			inProxySection = trimmedLine == "[Proxy]"
			continue
		}
		if !inProxySection || trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) != 2 {
			continue
		}
		names = append(names, strings.TrimSpace(parts[0]))
	}

	return names
}

func testClientProxyNameNodes(splitFirst bool) []models.Node {
	firstLink := "ss://YWVzLTEyOC1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#first"
	if splitFirst {
		firstLink = strings.Join([]string{
			"ss://YWVzLTEyOC1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#first-a",
			"ss://YWVzLTEyOC1nY206cGFzc0BleGFtcGxlLm9yZzo0NDM=#first-b",
		}, ",")
	}

	return []models.Node{
		{ID: 1, Name: "first", LinkName: "first", Link: firstLink, Protocol: "ss", Source: "source-a"},
		{
			ID:       2,
			Name:     "second",
			LinkName: "second",
			Link:     "ss://YWVzLTEyOC1nY206cGFzc0BleGFtcGxlLm5ldDo0NDM=#second",
			Protocol: "ss",
			Source:   "source-a",
		},
	}
}

func renderPreparedClientProxyNames(t *testing.T, clientType string, nodeNameRule string, nodes []models.Node) []string {
	t.Helper()
	setupClientsAPITestDB(t)
	utils.SetProtocolLinkFuncs(protocol.GetProtocolLabelFromLink, protocol.RenameNodeLink)
	t.Cleanup(func() {
		utils.SetProtocolLinkFuncs(nil, nil)
	})

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/c/?client="+clientType, nil)

	templatePath := writeTestClashTemplate(t)
	config := `{"clash":"` + templatePath + `"}`
	prepared := preparedClientResponse{
		ClientType: clientType,
		Mode:       clientResponseNormal,
		SubName:    "duplicate-name-sub",
		Subscription: models.Subcription{
			Name:                  "duplicate-name-sub",
			Config:                config,
			RefreshUsageOnRequest: false,
			NodeNameRule:          nodeNameRule,
			Nodes:                 nodes,
		},
	}
	if clientType == "surge" {
		templatePath = writeTestSurgeTemplate(t)
		prepared.Subscription.Config = `{"surge":"` + templatePath + `"}`
		renderPreparedSurge(ginContext, prepared)
		return surgeProxyNamesFromBody(recorder.Body.String())
	}

	renderPreparedClash(ginContext, prepared)
	return clashProxyNamesFromBody(t, recorder.Body.Bytes())
}

func TestRenderPreparedV2raySkipsProtocolUnsupportedLinks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/c/?client=v2ray", nil)

	prepared := preparedClientResponse{
		ClientType: "v2ray",
		Mode:       clientResponseNormal,
		SubName:    "mieru-skip",
		Subscription: models.Subcription{
			Name:                  "mieru-skip",
			RefreshUsageOnRequest: false,
			Nodes: []models.Node{
				{
					Name:     "mieru-node",
					LinkName: "mieru-node",
					Link:     "mieru://user:password@mieru.example.com:2999?transport=TCP#m",
					Protocol: "mieru",
				},
				{
					Name:     "ss-node",
					LinkName: "ss-node",
					Link:     "ss://YWVzLTEyOC1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node",
					Protocol: "ss",
				},
				{
					Name:     "official-mieru-node",
					LinkName: "official-mieru-node",
					Link:     "mierus://official.example.com:2999?profile=raw#official",
					Protocol: "mieru",
				},
				{
					Name:     "unknown-node",
					LinkName: "unknown-node",
					Link:     "unknown-protocol://example.com/raw#unknown",
					Protocol: "other",
				},
				{
					Name:     "mixed-node",
					LinkName: "mixed-node",
					Link:     "mieru://user:password@range.example.com?portRange=2090-2099&transport=UDP#range,ss://YWVzLTEyOC1nY206cGFzczJAZXhhbXBsZS5jb206NDQ0#ss-two",
					Protocol: "ss",
				},
			},
		},
	}

	renderPreparedV2ray(ginContext, prepared)
	decoded := utils.Base64Decode(recorder.Body.String())
	if strings.Contains(decoded, "mieru://") || strings.Contains(decoded, "mierus://") {
		t.Fatalf("v2ray output leaked mieru link: %s", decoded)
	}
	if !strings.Contains(decoded, "ss://") {
		t.Fatalf("v2ray output should keep supported raw links: %s", decoded)
	}
	if !strings.Contains(decoded, "unknown-protocol://example.com/raw#unknown") {
		t.Fatalf("v2ray output should preserve unregistered raw links: %s", decoded)
	}
}

func TestRenderPreparedV2rayUsesDuplicateIndexVariable(t *testing.T) {
	utils.SetProtocolLinkFuncs(protocol.GetProtocolLabelFromLink, protocol.RenameNodeLink)
	t.Cleanup(func() {
		utils.SetProtocolLinkFuncs(nil, nil)
	})

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/c/?client=v2ray", nil)

	renderPreparedV2ray(ginContext, preparedClientResponse{
		ClientType: "v2ray",
		Mode:       clientResponseNormal,
		SubName:    "duplicate-name-sub",
		Subscription: models.Subcription{
			Name:                  "duplicate-name-sub",
			RefreshUsageOnRequest: false,
			NodeNameRule:          "$Source$DuplicateIndex",
			Nodes:                 testClientProxyNameNodes(false),
		},
	})

	decoded := utils.Base64Decode(recorder.Body.String())
	for _, want := range []string{"#source-a", "#source-a1"} {
		if !strings.Contains(decoded, want) {
			t.Fatalf("v2ray output missing renamed node %q: %s", want, decoded)
		}
	}
}

func TestShouldSkipV2rayLinkUsesProtocolClientSupport(t *testing.T) {
	for _, tc := range []struct {
		name string
		link string
		skip bool
	}{
		{name: "mieru unsupported", link: "mieru://user:password@mieru.example.com:2999?transport=TCP#m", skip: true},
		{name: "mierus support alias unsupported", link: "mierus://official.example.com:2999?profile=raw#m", skip: true},
		{name: "ss default supported", link: "ss://YWVzLTEyOC1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node", skip: false},
		{name: "unknown raw preserved", link: "unknown-protocol://example.com/raw#unknown", skip: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldSkipV2rayLink(tc.link); got != tc.skip {
				t.Fatalf("shouldSkipV2rayLink() = %v, want %v", got, tc.skip)
			}
			if got := !protocol.SupportsClientForLink(tc.link, protocol.ClientV2ray); got != tc.skip {
				t.Fatalf("protocol support decision = %v, want %v", got, tc.skip)
			}
		})
	}
}

func TestResolveClashDialerProxyNormalizesStoredNameWithSubscriptionRule(t *testing.T) {
	frontNode := models.Node{
		ID:          1,
		Name:        "【亚洲】 日本02 | Vless",
		LinkName:    "【亚洲】 日本02 | Vless",
		Link:        "vless://12345678-1234-1234-1234-123456789abc@example.com:443?security=tls#%E3%80%90%E4%BA%9A%E6%B4%B2%E3%80%91%20%E6%97%A5%E6%9C%AC02%20%7C%20Vless",
		LinkCountry: "JP",
		Speed:       1.5,
	}
	targetNode := models.Node{
		ID:              2,
		Name:            "德国03[HY2]",
		LinkName:        "德国03[HY2]",
		Link:            "hysteria2://password@example.net:443#%E5%BE%B7%E5%9B%BD03%5BHY2%5D",
		DialerProxyName: "【亚洲】 日本02 | Vless",
	}
	sub := models.Subcription{
		NodeNameRule: "$Flag $LinkCountry - $LinkName - $Speed",
		Nodes:        []models.Node{frontNode, targetNode},
	}

	nodeNameMap := buildClashNodeNameMap(sub)
	dialerProxyNameMap := buildDialerProxyNameMap(sub.Nodes, nodeNameMap)
	got := resolveClashDialerProxy(targetNode, nodeNameMap[targetNode.ID], nil, nil, dialerProxyNameMap)
	want := "🇯🇵 JP - 【亚洲】 日本02 | Vless - 1.50MB/s"

	if got != want {
		t.Fatalf("expected stored dialer-proxy to normalize to subscription final name, got %q want %q", got, want)
	}
}

func TestBuildClashNodeNameMapKeepsDuplicateNamesWithoutDuplicateIndexVariable(t *testing.T) {
	sub := models.Subcription{
		NodeNameRule: "$Source",
		Nodes: []models.Node{
			{ID: 1, Name: "first", LinkName: "first", Source: "source-a"},
			{ID: 2, Name: "second", LinkName: "second", Source: "source-a"},
			{ID: 3, Name: "third", LinkName: "third", Source: "source-a"},
		},
	}

	nodeNameMap := buildClashNodeNameMap(sub)

	if got := nodeNameMap[1]; got != "source-a" {
		t.Fatalf("first name = %q, want source-a", got)
	}
	if got := nodeNameMap[2]; got != "source-a" {
		t.Fatalf("second name = %q, want source-a", got)
	}
	if got := nodeNameMap[3]; got != "source-a" {
		t.Fatalf("third name = %q, want source-a", got)
	}
}

func TestBuildClashNodeNameMapUsesDuplicateIndexVariable(t *testing.T) {
	sub := models.Subcription{
		NodeNameRule: "$Source$DuplicateIndex",
		Nodes: []models.Node{
			{ID: 1, Name: "first", LinkName: "first", Source: "source-a"},
			{ID: 2, Name: "second", LinkName: "second", Source: "source-a"},
			{ID: 3, Name: "third", LinkName: "third", Source: "source-a"},
		},
	}

	nodeNameMap := buildClashNodeNameMap(sub)

	if got := nodeNameMap[1]; got != "source-a" {
		t.Fatalf("first name = %q, want source-a", got)
	}
	if got := nodeNameMap[2]; got != "source-a1" {
		t.Fatalf("second name = %q, want source-a1", got)
	}
	if got := nodeNameMap[3]; got != "source-a2" {
		t.Fatalf("third name = %q, want source-a2", got)
	}
}

// TestBuildClashNodeNameMapKeepsDialerProxyAliasesAligned 确认变量编号后的名称仍能解析 dialer-proxy。
// 链式代理或手工 dialer-proxy 可能保存旧名称，导出时必须映射到最终唯一 proxy.name。
func TestBuildClashNodeNameMapKeepsDialerProxyAliasesAligned(t *testing.T) {
	nodes := []models.Node{
		{ID: 1, Name: "front-a", LinkName: "front-a", Source: "source-a"},
		{ID: 2, Name: "front-b", LinkName: "front-b", Source: "source-a"},
		{ID: 3, Name: "target", LinkName: "target", DialerProxyName: "front-b"},
	}
	sub := models.Subcription{
		NodeNameRule: "$Source$DuplicateIndex",
		Nodes:        nodes,
	}

	nodeNameMap := buildClashNodeNameMap(sub)
	dialerProxyNameMap := buildDialerProxyNameMap(nodes, nodeNameMap)
	got := resolveClashDialerProxy(nodes[2], nodeNameMap[nodes[2].ID], nil, nil, dialerProxyNameMap)

	if got != "source-a1" {
		t.Fatalf("expected dialer-proxy alias to resolve to unique final name, got %q", got)
	}
}

func TestResolveClashDialerProxyPrefersChainTargetOverStoredName(t *testing.T) {
	targetNode := models.Node{
		ID:              2,
		Name:            "target-node",
		LinkName:        "target-node",
		DialerProxyName: "stale-front-name",
	}
	targetNodeDialerMap := map[int]string{targetNode.ID: "renamed-front-name"}

	got := resolveClashDialerProxy(targetNode, "target-node", nil, targetNodeDialerMap, nil)

	if got != "renamed-front-name" {
		t.Fatalf("expected chain target dialer to override stored dialer-proxy, got %q", got)
	}
}

func TestBuildRenamedNodeLinkUsesQualityStatusFields(t *testing.T) {
	utils.SetProtocolLinkFuncs(protocol.GetProtocolLabelFromLink, protocol.RenameNodeLink)
	t.Cleanup(func() {
		utils.SetProtocolLinkFuncs(nil, nil)
	})

	node := models.Node{
		Name:          "origin-name",
		LinkName:      "origin-name",
		Link:          "ss://YWVzLTEyOC1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#origin-name",
		LinkCountry:   "US",
		Source:        "[BK]",
		Speed:         1.25,
		SpeedStatus:   "success",
		IsResidential: false,
		IsBroadcast:   false,
		FraudScore:    8,
		QualityStatus: "success",
	}

	got := buildRenamedNodeLink(node, node.LinkName, "$Source$Flag$LinkCountry|$Residential|$IpType|$FraudScoreIcon$Speed-$Index", node.Link, 7)

	if !strings.Contains(got, "%5BBK%5D%F0%9F%87%BA%F0%9F%87%B8US%7C%E6%9C%BA%E6%88%BFIP%7C%E5%8E%9F%E7%94%9FIP%7C%E2%9A%AA1.25MB/s-7") {
		t.Fatalf("expected surge renamed link to include quality fields, got %q", got)
	}
}

func TestBuildDialerProxyNameMapIgnoresAmbiguousAliases(t *testing.T) {
	nodes := []models.Node{
		{ID: 1, Name: "same-old-name", LinkName: "same-old-name"},
		{ID: 2, Name: "same-old-name", LinkName: "same-old-name"},
	}
	nodeNameMap := map[int]string{1: "final-a", 2: "final-b"}

	dialerProxyNameMap := buildDialerProxyNameMap(nodes, nodeNameMap)
	got := normalizeDialerProxyName("same-old-name", dialerProxyNameMap)

	if got != "same-old-name" {
		t.Fatalf("expected ambiguous alias to remain unchanged, got %q", got)
	}
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
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, tt.path, nil)

			tt.handler(ginContext)

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
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodHead, "/c/?token=head-token&client=v2ray", nil)

	GetClient(ginContext)

	if body := recorder.Body.String(); body != "" {
		t.Fatalf("expected empty HEAD body, got %q", body)
	}
	if got := recorder.Header().Get("subscription-userinfo"); got == "" {
		t.Fatal("expected subscription-userinfo header to be set")
	}
	if value, ok := ginContext.Get("subname"); !ok || value != "head-sub" {
		t.Fatalf("expected request-scoped subname to be preserved, got %v ok=%v", value, ok)
	}
}

func TestRenderPreparedClashSetsProfileUpdateIntervalHeader(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)

	tests := []struct {
		name           string
		updateInterval int
		wantHeader     string
	}{
		{name: "custom hours", updateInterval: 6, wantHeader: "6"},
		{name: "default hours", updateInterval: 0, wantHeader: "24"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			recorder := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/c/?client=clash", nil)

			renderPreparedClash(ginContext, preparedClientResponse{
				ClientType: "clash",
				Mode:       clientResponseNormal,
				SubName:    "interval-sub",
				Subscription: models.Subcription{
					Name:                  "interval-sub",
					Config:                `{"clash":"` + clashTemplatePath + `"}`,
					RefreshUsageOnRequest: false,
					UpdateInterval:        tt.updateInterval,
				},
			})

			if got := recorder.Header().Get("profile-update-interval"); got != tt.wantHeader {
				t.Fatalf("expected profile-update-interval %q, got %q", tt.wantHeader, got)
			}
		})
	}
}

func TestRenderPreparedClientKeepsDuplicateProxyNamesWithoutDuplicateIndexVariable(t *testing.T) {
	tests := []struct {
		name       string
		clientType string
		splitFirst bool
		wantNames  []string
	}{
		{name: "clash duplicate", clientType: "clash", wantNames: []string{"source-a", "source-a"}},
		{name: "clash split", clientType: "clash", splitFirst: true, wantNames: []string{"source-a", "source-a", "source-a"}},
		{name: "surge duplicate", clientType: "surge", wantNames: []string{"source-a", "source-a"}},
		{name: "surge split", clientType: "surge", splitFirst: true, wantNames: []string{"source-a", "source-a", "source-a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names := renderPreparedClientProxyNames(t, tt.clientType, "$Source", testClientProxyNameNodes(tt.splitFirst))
			if strings.Join(names, ",") != strings.Join(tt.wantNames, ",") {
				t.Fatalf("proxy names = %v, want %v", names, tt.wantNames)
			}
		})
	}
}

func TestRenderPreparedClientUsesDuplicateIndexVariableForDuplicateProxyNames(t *testing.T) {
	tests := []struct {
		name       string
		clientType string
		splitFirst bool
		wantNames  []string
	}{
		{name: "clash duplicate", clientType: "clash", wantNames: []string{"source-a", "source-a1"}},
		{name: "clash split", clientType: "clash", splitFirst: true, wantNames: []string{"source-a", "source-a1", "source-a2"}},
		{name: "surge duplicate", clientType: "surge", wantNames: []string{"source-a", "source-a1"}},
		{name: "surge split", clientType: "surge", splitFirst: true, wantNames: []string{"source-a", "source-a1", "source-a2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names := renderPreparedClientProxyNames(t, tt.clientType, "$Source$DuplicateIndex", testClientProxyNameNodes(tt.splitFirst))
			if strings.Join(names, ",") != strings.Join(tt.wantNames, ",") {
				t.Fatalf("proxy names = %v, want %v", names, tt.wantNames)
			}
		})
	}
}

func TestRenderPreparedSurgeUsesSubscriptionUpdateIntervalSeconds(t *testing.T) {
	setupClientsAPITestDB(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)

	tests := []struct {
		name           string
		updateInterval int
		wantInterval   string
	}{
		{name: "custom hours converted to seconds", updateInterval: 6, wantInterval: "interval=21600"},
		{name: "default seconds", updateInterval: 0, wantInterval: "interval=86400"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			recorder := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/c/?client=surge", nil)
			ginContext.Request.Host = "example.test"

			renderPreparedSurge(ginContext, preparedClientResponse{
				ClientType: "surge",
				Mode:       clientResponseNormal,
				SubName:    "interval-sub",
				Subscription: models.Subcription{
					Name:                  "interval-sub",
					Config:                `{"surge":"` + surgeTemplatePath + `"}`,
					RefreshUsageOnRequest: false,
					UpdateInterval:        tt.updateInterval,
				},
			})

			body := recorder.Body.String()
			if !strings.Contains(body, tt.wantInterval) {
				t.Fatalf("expected surge managed config to contain %q, got %q", tt.wantInterval, body)
			}
		})
	}
}

func TestRenderPreparedSurgeRewritesExistingManagedConfigInterval(t *testing.T) {
	setupClientsAPITestDB(t)
	surgeTemplatePath := writeTestSurgeTemplateWithManagedConfig(t)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/c/?client=surge", nil)

	renderPreparedSurge(ginContext, preparedClientResponse{
		ClientType: "surge",
		Mode:       clientResponseNormal,
		SubName:    "managed-interval-sub",
		Subscription: models.Subcription{
			Name:                  "managed-interval-sub",
			Config:                `{"surge":"` + surgeTemplatePath + `"}`,
			RefreshUsageOnRequest: false,
			UpdateInterval:        6,
		},
	})

	body := recorder.Body.String()
	if !strings.Contains(body, "#!MANAGED-CONFIG https://old.example/sub interval=21600 strict=false") {
		t.Fatalf("expected existing managed config interval to be rewritten, got %q", body)
	}
	if strings.Contains(body, "interval=86400 strict=false") {
		t.Fatalf("expected old managed config interval to be replaced, got %q", body)
	}
}

func TestParseSubscriptionUpdateIntervalBoundsValues(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want int
	}{
		{name: "empty", raw: "", want: 0},
		{name: "non numeric", raw: "abc", want: 0},
		{name: "negative", raw: "-1", want: 0},
		{name: "valid", raw: "6", want: 6},
		{name: "max valid", raw: "8760", want: maxSubscriptionUpdateIntervalHours},
		{name: "too large", raw: "999999999", want: maxSubscriptionUpdateIntervalHours},
		{name: "larger than uint64", raw: "999999999999999999999999999999", want: maxSubscriptionUpdateIntervalHours},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSubscriptionUpdateInterval(tt.raw); got != tt.want {
				t.Fatalf("parseSubscriptionUpdateInterval(%q)=%d, want %d", tt.raw, got, tt.want)
			}
		})
	}
}

func TestGetClientInvalidShareReturnsSyntheticV2rayNode(t *testing.T) {
	setupClientsAPITestDB(t)

	recorder := performClientRequest(t, http.MethodGet, "/c/?token=missing-token&client=v2ray")
	decoded := strings.TrimSpace(utils.Base64Decode(recorder.Body.String()))

	if !strings.Contains(decoded, "ss://") || !strings.Contains(decoded, "#无效的分享链接") {
		t.Fatalf("expected synthetic invalid-share node in v2ray payload, got %q", decoded)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("expected v2ray content type, got %q", got)
	}
	if got := recorder.Header().Get("subscription-userinfo"); got == "" {
		t.Fatal("expected subscription-userinfo header for invalid v2ray share")
	}
}

func TestGetClientExpandedTargetRequiresConfiguredSubStore(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "expanded-sub", "expanded-token", "Expanded Node")

	recorder := performClientRequest(t, http.MethodGet, "/c/?token=expanded-token&client=loon")

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when Sub-Store is not configured, got %d body=%q", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "Sub-Store sidecar is not configured") {
		t.Fatalf("expected Sub-Store configuration error, got %q", recorder.Body.String())
	}
}

func TestGetClientExpandedTargetUsesDatabaseSubStoreSettings(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "database-substore-sub", "database-substore-token", "Database SubStore Node")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/proxy/parse" {
			t.Fatalf("unexpected Sub-Store path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"success","data":{"par_res":"database settings converted output"}}`))
	}))
	t.Cleanup(server.Close)

	if err := models.SetSetting("substore_enabled", "true"); err != nil {
		t.Fatalf("set substore enabled: %v", err)
	}
	if err := models.SetSetting("substore_base_url", server.URL); err != nil {
		t.Fatalf("set substore base url: %v", err)
	}
	if err := models.SetSetting("substore_allowed_targets", "loon"); err != nil {
		t.Fatalf("set substore targets: %v", err)
	}

	recorder := performClientRequest(t, http.MethodGet, "/c/?token=database-substore-token&client=loon")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%q", recorder.Code, recorder.Body.String())
	}
	if recorder.Body.String() != "database settings converted output" {
		t.Fatalf("unexpected converted body: %q", recorder.Body.String())
	}
}

func TestGetClientExpandedTargetUsesMihomoBridgeAndSubStore(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "expanded-sub", "expanded-token", "Expanded Node")

	var received struct {
		Data   string `json:"data"`
		Client string `json:"client"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/proxy/parse" {
			t.Fatalf("unexpected Sub-Store path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode Sub-Store request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"par_res":"loon converted output"}}`))
	}))
	defer server.Close()
	saveSubStoreSettings(t, server.URL, []string{"loon"})

	recorder := performClientRequest(t, http.MethodGet, "/c/?token=expanded-token&client=loon")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 from converted subscription, got %d body=%q", recorder.Code, recorder.Body.String())
	}
	if recorder.Body.String() != "loon converted output" {
		t.Fatalf("unexpected converted body: %q", recorder.Body.String())
	}
	if received.Client != "Loon" {
		t.Fatalf("expected Sub-Store target Loon, got %q", received.Client)
	}
	if !strings.Contains(received.Data, "proxies:") || !strings.Contains(received.Data, "name: Expanded Node") {
		t.Fatalf("expected generated mihomo YAML bridge, got %q", received.Data)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("unexpected converted content type: %q", got)
	}
	if !strings.Contains(recorder.Header().Get("Content-Disposition"), "expanded-sub.conf") {
		t.Fatalf("unexpected content disposition: %q", recorder.Header().Get("Content-Disposition"))
	}
}

func TestGetClientExpandedTargetMapsQuanxToQX(t *testing.T) {
	setupClientsAPITestDB(t)
	clashTemplatePath := writeTestClashTemplate(t)
	surgeTemplatePath := writeTestSurgeTemplate(t)
	createClientSubscriptionFixture(t, clashTemplatePath, surgeTemplatePath, "quanx-sub", "quanx-token", "Quanx Node")

	var target string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Client string `json:"client"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode Sub-Store request: %v", err)
		}
		target = req.Client
		_, _ = w.Write([]byte(`{"status":"success","data":{"par_res":"qx converted output"}}`))
	}))
	defer server.Close()
	saveSubStoreSettings(t, server.URL, []string{"quanx"})

	recorder := performClientRequest(t, http.MethodGet, "/c/?token=quanx-token&client=quanx")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 from converted subscription, got %d body=%q", recorder.Code, recorder.Body.String())
	}
	if target != "QX" {
		t.Fatalf("expected Sub-Store target QX, got %q", target)
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
	if !strings.Contains(body, "type: ss") {
		t.Fatalf("expected synthetic expired-share clash node to use ss type, got %q", body)
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
	if !strings.Contains(body, "无效的分享链接 = ss, placeholder.invalid, 80") {
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

	if !strings.Contains(decoded, "ss://") || !strings.Contains(decoded, "#订阅不存在") {
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

	if !strings.Contains(body, "订阅已过期 = ss, placeholder.invalid, 80") {
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

	if !strings.Contains(decoded, "ss://") || !strings.Contains(decoded, "#订阅已过期") {
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
