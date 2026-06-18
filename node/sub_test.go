package node

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
	"sublink/node/protocol"
	"testing"

	"github.com/glebarez/sqlite"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

func TestGenerateProxyLinkReconstructsDNSStyleECH(t *testing.T) {
	proxy := protocol.Proxy{
		Name:       "导入节点-ECH-DNS",
		Type:       "vless",
		Server:     "example.com",
		Port:       443,
		Uuid:       "12345678-1234-1234-1234-123456789abc",
		Network:    "ws",
		Tls:        true,
		Servername: "example.com",
		ECH_opts: map[string]any{
			"enable":            true,
			"query-server-name": "encryptedsni.com",
		},
	}

	link := GenerateProxyLink(proxy)
	if link == "" {
		t.Fatal("生成链接失败")
	}

	if !strings.Contains(link, "ech=encryptedsni.com%2Bhttps%3A%2F%2Fdns.alidns.com%2Fdns-query") {
		t.Fatalf("ImportedECH 应包含重建后的 DNS ECH, 实际: %s", link)
	}
	decoded, err := protocol.DecodeVLESSURL(link)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	if decoded.Query.Ech != "encryptedsni.com+https://dns.alidns.com/dns-query" {
		t.Fatalf("RestoredECH 不匹配: 期望 %s, 实际 %s", "encryptedsni.com+https://dns.alidns.com/dns-query", decoded.Query.Ech)
	}
}

func TestGenerateProxyLinkPreservesECHConfig(t *testing.T) {
	proxy := protocol.Proxy{
		Name:       "导入节点-ECH-Config",
		Type:       "vless",
		Server:     "example.com",
		Port:       443,
		Uuid:       "12345678-1234-1234-1234-123456789abc",
		Network:    "ws",
		Tls:        true,
		Servername: "example.com",
		ECH_opts: map[string]any{
			"enable": true,
			"config": "BASE64_ECH_CONFIG",
		},
	}

	link := GenerateProxyLink(proxy)
	if link == "" {
		t.Fatal("生成链接失败")
	}

	if !strings.Contains(link, "ech=BASE64_ECH_CONFIG") {
		t.Fatalf("ImportedECHConfig 应包含 config 形式 ECH, 实际: %s", link)
	}
}

func TestGenerateProxyLinkKeepsNonVLESSUnchanged(t *testing.T) {
	proxy := protocol.Proxy{
		Name:     "trojan-node",
		Type:     "trojan",
		Server:   "example.com",
		Port:     443,
		Password: "secret",
	}

	genericLink := GenerateProxyLink(proxy)
	reconstructedLink := GenerateProxyLink(proxy)
	if genericLink != reconstructedLink {
		t.Fatalf("NonVLESSLink 不匹配: 期望 %s, 实际 %s", genericLink, reconstructedLink)
	}
}

func TestGenerateProxyLinkDoesNotReconstructDisabledECH(t *testing.T) {
	proxy := protocol.Proxy{
		Name:       "导入节点-ECH-Disabled",
		Type:       "vless",
		Server:     "example.com",
		Port:       443,
		Uuid:       "12345678-1234-1234-1234-123456789abc",
		Network:    "ws",
		Tls:        true,
		Servername: "example.com",
		ECH_opts: map[string]any{
			"enable":            false,
			"query-server-name": "encryptedsni.com",
		},
	}

	link := GenerateProxyLink(proxy)
	if link == "" {
		t.Fatal("生成链接失败")
	}
	if strings.Contains(link, "ech=") {
		t.Fatalf("禁用 ECH 时不应重建顶层 ech, 实际: %s", link)
	}
}

func TestExpandClashProxyProvidersFetchesReferencedHTTPProviderWithHeaders(t *testing.T) {
	var gotUserAgent string
	var gotCustomHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		gotCustomHeader = r.Header.Get("X-Airport-Token")
		_, _ = w.Write([]byte(`proxies:
  - name: provider-node
    type: trojan
    server: provider.example.com
    port: 443
    password: secret
`))
	}))
	defer server.Close()

	var config ClashConfig
	if err := yaml.Unmarshal([]byte(`proxy-providers:
  remote-a:
    type: http
    url: "`+server.URL+`/remote-a"
  remote-b:
    type: http
    url: "`+server.URL+`/remote-b"
proxy-groups:
  - name: auto
    type: select
    use:
      - remote-a
`), &config); err != nil {
		t.Fatalf("yaml unmarshal failed: %v", err)
	}

	err := expandClashProxyProviders(context.Background(), server.Client(), server.URL+"/root", &config, "SublinkPro-Test", models.AirportRequestHeaders{
		{Key: "X-Airport-Token", Value: "token-1"},
	})
	if err != nil {
		t.Fatalf("expandClashProxyProviders failed: %v", err)
	}
	if len(config.Proxies) != 1 {
		t.Fatalf("proxy count = %d, want 1", len(config.Proxies))
	}
	if config.Proxies[0].Name != "provider-node" {
		t.Fatalf("proxy name = %q, want provider-node", config.Proxies[0].Name)
	}
	if gotUserAgent != "SublinkPro-Test" {
		t.Fatalf("provider User-Agent = %q, want SublinkPro-Test", gotUserAgent)
	}
	if gotCustomHeader != "token-1" {
		t.Fatalf("provider custom header = %q, want token-1", gotCustomHeader)
	}
}

func TestExpandClashProxyProvidersDoesNotForwardCustomHeadersToDifferentHost(t *testing.T) {
	var gotUserAgent string
	var gotCustomHeader string
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		gotCustomHeader = r.Header.Get("X-Airport-Token")
		_, _ = w.Write([]byte(`proxies:
  - name: cross-host-provider-node
    type: trojan
    server: provider.example.com
    port: 443
    password: secret
`))
	}))
	defer providerServer.Close()

	rootServer := httptest.NewServer(http.NotFoundHandler())
	defer rootServer.Close()

	var config ClashConfig
	if err := yaml.Unmarshal([]byte(`proxy-providers:
  remote-a:
    type: http
    url: "`+providerServer.URL+`/remote-a"
`), &config); err != nil {
		t.Fatalf("yaml unmarshal failed: %v", err)
	}

	err := expandClashProxyProviders(context.Background(), providerServer.Client(), rootServer.URL+"/root", &config, "SublinkPro-Test", models.AirportRequestHeaders{
		{Key: "X-Airport-Token", Value: "token-1"},
	})
	if err != nil {
		t.Fatalf("expandClashProxyProviders failed: %v", err)
	}
	if gotUserAgent != "SublinkPro-Test" {
		t.Fatalf("provider User-Agent = %q, want SublinkPro-Test", gotUserAgent)
	}
	if gotCustomHeader != "" {
		t.Fatalf("cross-host provider custom header = %q, want empty", gotCustomHeader)
	}
}

func TestExpandClashProxyProvidersStripsCustomHeadersOnCrossHostRedirect(t *testing.T) {
	var redirectedUserAgent string
	var redirectedCustomHeader string
	redirectTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectedUserAgent = r.Header.Get("User-Agent")
		redirectedCustomHeader = r.Header.Get("X-Airport-Token")
		_, _ = w.Write([]byte(`proxies:
  - name: redirected-provider-node
    type: trojan
    server: redirected.example.com
    port: 443
    password: secret
`))
	}))
	defer redirectTarget.Close()

	rootServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectTarget.URL+"/provider", http.StatusFound)
	}))
	defer rootServer.Close()

	var config ClashConfig
	if err := yaml.Unmarshal([]byte(`proxy-providers:
  remote-a:
    type: http
    url: "`+rootServer.URL+`/remote-a"
`), &config); err != nil {
		t.Fatalf("yaml unmarshal failed: %v", err)
	}

	err := expandClashProxyProviders(context.Background(), rootServer.Client(), rootServer.URL+"/root", &config, "SublinkPro-Test", models.AirportRequestHeaders{
		{Key: "X-Airport-Token", Value: "token-1"},
	})
	if err != nil {
		t.Fatalf("expandClashProxyProviders failed: %v", err)
	}
	if redirectedUserAgent != "SublinkPro-Test" {
		t.Fatalf("redirected User-Agent = %q, want SublinkPro-Test", redirectedUserAgent)
	}
	if redirectedCustomHeader != "" {
		t.Fatalf("redirected custom header = %q, want empty", redirectedCustomHeader)
	}
}

func TestFetchClashProxyProviderRejectsInvalidURLScheme(t *testing.T) {
	_, err := fetchClashProxyProvider(context.Background(), http.DefaultClient, "bad-scheme", "ftp://example.com/provider.yaml", "example.com", "", nil)
	if err == nil {
		t.Fatal("expected invalid scheme error, got nil")
	}
	if !strings.Contains(err.Error(), "不受支持") {
		t.Fatalf("invalid scheme error = %q, want unsupported scheme message", err.Error())
	}
}

func TestFetchClashProxyProviderRejectsInvalidRedirectScheme(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "ftp://example.com/provider.yaml", http.StatusFound)
	}))
	defer server.Close()

	_, err := fetchClashProxyProvider(context.Background(), server.Client(), "bad-redirect-scheme", server.URL+"/provider", normalizedURLHost(server.URL), "", nil)
	if err == nil {
		t.Fatal("expected invalid redirect scheme error, got nil")
	}
	if !strings.Contains(err.Error(), "ftp") || !strings.Contains(err.Error(), "不受支持") {
		t.Fatalf("invalid redirect scheme error = %q, want unsupported ftp scheme message", err.Error())
	}
}

func TestFetchClashProxyProviderRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chunk := strings.Repeat("x", 1024*1024)
		for i := int64(0); i <= providerResponseSizeLimit/int64(len(chunk)); i++ {
			_, _ = w.Write([]byte(chunk))
		}
	}))
	defer server.Close()

	_, err := fetchClashProxyProvider(context.Background(), server.Client(), "too-large", server.URL+"/provider", normalizedURLHost(server.URL), "", nil)
	if err == nil {
		t.Fatal("expected oversized provider response error, got nil")
	}
	if !strings.Contains(err.Error(), "超过大小限制") {
		t.Fatalf("oversized response error = %q, want size limit message", err.Error())
	}
}

func TestSelectClashProxyProvidersIncludeAllVariantsSelectAllProviders(t *testing.T) {
	for _, tc := range []struct {
		name  string
		group string
	}{
		{name: "include-all", group: "include-all: true"},
		{name: "include-all-providers", group: "include-all-providers: true"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var config ClashConfig
			if err := yaml.Unmarshal([]byte(`proxy-providers:
  provider-a:
    type: http
    url: "http://example.com/a"
  provider-b:
    type: http
    url: "http://example.com/b"
proxy-groups:
  - name: auto
    type: select
    use:
      - provider-a
    `+tc.group+`
`), &config); err != nil {
				t.Fatalf("yaml unmarshal failed: %v", err)
			}

			providers := selectClashProxyProviders(&config)
			gotNames := []string{providers[0].Name, providers[1].Name}
			wantNames := []string{"provider-a", "provider-b"}
			if !reflect.DeepEqual(gotNames, wantNames) {
				t.Fatalf("selected providers = %v, want %v", gotNames, wantNames)
			}
		})
	}
}

func TestExpandClashProxyProvidersRejectsTooManySelectedProviders(t *testing.T) {
	config := ClashConfig{
		ProxyProviders:     make(map[string]ClashProxyProvider, selectedProviderCountLimit+1),
		ProxyProviderOrder: make([]string, 0, selectedProviderCountLimit+1),
	}
	for i := 0; i < selectedProviderCountLimit+1; i++ {
		name := "provider-" + strconv.Itoa(i)
		config.ProxyProviderOrder = append(config.ProxyProviderOrder, name)
		config.ProxyProviders[name] = ClashProxyProvider{Type: "http", URL: "http://example.com/" + name}
	}

	err := expandClashProxyProviders(context.Background(), http.DefaultClient, "http://example.com/root", &config, "", nil)
	if err == nil {
		t.Fatal("expected selected provider count limit error, got nil")
	}
	if !strings.Contains(err.Error(), "数量过多") {
		t.Fatalf("provider count error = %q, want count limit message", err.Error())
	}
}

func TestParseClashConfigDataKeepsTopLevelProxiesWithoutFetchingProviders(t *testing.T) {
	providerFetched := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerFetched = true
		_, _ = w.Write([]byte(`proxies:
  - name: provider-node
    type: trojan
    server: provider.example.com
    port: 443
    password: secret
`))
	}))
	defer server.Close()

	config, errYaml, providerErr := parseClashConfigData(context.Background(), server.Client(), server.URL+"/root", []byte(`proxies:
  - name: inline-node
    type: trojan
    server: inline.example.com
    port: 443
    password: secret
proxy-providers:
  remote-a:
    type: http
    url: "`+server.URL+`/remote-a"
proxy-groups:
  - name: auto
    type: select
    use:
      - remote-a
`), "", nil)
	if errYaml != nil {
		t.Fatalf("yaml unmarshal failed: %v", errYaml)
	}
	if providerErr != nil {
		t.Fatalf("provider expansion should be skipped, got error: %v", providerErr)
	}
	if providerFetched {
		t.Fatal("provider URL was fetched even though top-level proxies exist")
	}
	if len(config.Proxies) != 1 {
		t.Fatalf("proxy count = %d, want 1", len(config.Proxies))
	}
	if config.Proxies[0].Name != "inline-node" {
		t.Fatalf("proxy name = %q, want inline-node", config.Proxies[0].Name)
	}
}

func TestParseClashConfigDataExpandsProviderOnlyConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`proxies:
  - name: provider-only-node
    type: trojan
    server: provider-only.example.com
    port: 443
    password: secret
  - name: provider-hy-node
    type: hysteria
    server: hy.example.com
    port: 443
    auth-str: secret-hy
    up: 11 Mbps
    down: 55 mbps
  - name: provider-hy2-node
    type: hysteria2
    server: hy2.example.com
    port: 443
    password: secret-hy2
    up-speed: 11mbps
    down-speed: 55 Mbps
`))
	}))
	defer server.Close()

	config, errYaml, providerErr := parseClashConfigData(context.Background(), server.Client(), server.URL+"/root", []byte(`proxy-providers:
  remote-a:
    type: http
    url: "`+server.URL+`/remote-a"
proxy-groups:
  - name: auto
    type: select
    use:
      - remote-a
`), "", nil)
	if errYaml != nil {
		t.Fatalf("yaml unmarshal failed: %v", errYaml)
	}
	if providerErr != nil {
		t.Fatalf("provider expansion failed: %v", providerErr)
	}
	if len(config.Proxies) != 3 {
		t.Fatalf("proxy count = %d, want 3", len(config.Proxies))
	}
	if config.Proxies[0].Name != "provider-only-node" {
		t.Fatalf("proxy name = %q, want provider-only-node", config.Proxies[0].Name)
	}
	if got := int(config.Proxies[1].Up); got != 11 {
		t.Fatalf("hysteria up = %d, want 11", got)
	}
	if got := int(config.Proxies[1].Down); got != 55 {
		t.Fatalf("hysteria down = %d, want 55", got)
	}
	if got := int(config.Proxies[2].Up_Speed); got != 11 {
		t.Fatalf("hysteria2 up-speed = %d, want 11", got)
	}
	if got := int(config.Proxies[2].Down_Speed); got != 55 {
		t.Fatalf("hysteria2 down-speed = %d, want 55", got)
	}
}

func TestExpandClashProxyProvidersFallsBackToAllHTTPProvidersAndDedupesURL(t *testing.T) {
	requestCounts := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCounts[r.URL.Path]++
		switch r.URL.Path {
		case "/a":
			_, _ = w.Write([]byte(`proxies:
  - name: provider-a
    type: trojan
    server: a.example.com
    port: 443
    password: secret-a
`))
		case "/c":
			_, _ = w.Write([]byte(`proxies:
  - name: provider-c
    type: trojan
    server: c.example.com
    port: 443
    password: secret-c
`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var config ClashConfig
	if err := yaml.Unmarshal([]byte(`proxy-providers:
  provider-a:
    type: http
    url: "`+server.URL+`/a"
  provider-a-duplicate-url:
    type: http
    url: "`+server.URL+`/a"
  provider-file:
    type: file
    path: ./local.yaml
  provider-c:
    type: http
    url: "`+server.URL+`/c"
`), &config); err != nil {
		t.Fatalf("yaml unmarshal failed: %v", err)
	}

	err := expandClashProxyProviders(context.Background(), server.Client(), server.URL+"/root", &config, "", nil)
	if err != nil {
		t.Fatalf("expandClashProxyProviders failed: %v", err)
	}
	if len(config.Proxies) != 2 {
		t.Fatalf("proxy count = %d, want 2", len(config.Proxies))
	}
	gotNames := []string{config.Proxies[0].Name, config.Proxies[1].Name}
	wantNames := []string{"provider-a", "provider-c"}
	if !reflect.DeepEqual(gotNames, wantNames) {
		t.Fatalf("provider order = %v, want %v", gotNames, wantNames)
	}
	if requestCounts["/a"] != 1 {
		t.Fatalf("duplicate provider URL fetched %d times, want 1", requestCounts["/a"])
	}
	if requestCounts["/c"] != 1 {
		t.Fatalf("provider /c fetched %d times, want 1", requestCounts["/c"])
	}
}

func TestExpandClashProxyProvidersReturnsProviderErrorWhenNoNodesParsed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "provider unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	var config ClashConfig
	if err := yaml.Unmarshal([]byte(`proxy-providers:
  broken:
    type: http
    url: "`+server.URL+`/broken"
`), &config); err != nil {
		t.Fatalf("yaml unmarshal failed: %v", err)
	}

	err := expandClashProxyProviders(context.Background(), server.Client(), server.URL+"/root", &config, "", nil)
	if err == nil {
		t.Fatal("expected provider error, got nil")
	}
	if !strings.Contains(err.Error(), "broken") || !strings.Contains(err.Error(), "502") {
		t.Fatalf("provider error = %q, want provider name and HTTP status", err.Error())
	}
}

func TestParseClashProviderPayloadSupportsPlainAndBase64Links(t *testing.T) {
	link := GenerateProxyLink(protocol.Proxy{
		Name:     "plain-provider-node",
		Type:     "trojan",
		Server:   "plain.example.com",
		Port:     443,
		Password: "secret",
	})
	if link == "" {
		t.Fatal("GenerateProxyLink returned empty link")
	}

	plainProxies, err := parseClashProviderPayload([]byte(link + "\n"))
	if err != nil {
		t.Fatalf("plain provider payload parse failed: %v", err)
	}
	if len(plainProxies) != 1 || plainProxies[0].Name != "plain-provider-node" {
		t.Fatalf("plain provider proxies = %+v, want one plain-provider-node", plainProxies)
	}

	base64Proxies, err := parseClashProviderPayload([]byte(base64.StdEncoding.EncodeToString([]byte(link + "\n"))))
	if err != nil {
		t.Fatalf("base64 provider payload parse failed: %v", err)
	}
	if len(base64Proxies) != 1 || base64Proxies[0].Name != "plain-provider-node" {
		t.Fatalf("base64 provider proxies = %+v, want one plain-provider-node", base64Proxies)
	}
}

func TestApplyAirportNodeNamePrefixAddsPrefixOnly(t *testing.T) {
	airport := &models.Airport{
		ID:               27,
		NodeNameUniquify: true,
	}
	proxys := []protocol.Proxy{{Name: "香港节点-01"}, {Name: "香港节点-02"}}

	result := applyAirportNodeNamePrefix(airport, proxys)
	got := []string{result[0].Name, result[1].Name}
	want := []string{"[A27]香港节点-01", "[A27]香港节点-02"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("前缀唯一化结果不匹配: got=%v want=%v", got, want)
	}
}

func TestApplyAirportNodeNamePrefixFallsBackForWhitespacePrefix(t *testing.T) {
	airport := &models.Airport{
		ID:               27,
		NodeNameUniquify: true,
		NodeNamePrefix:   "   ",
	}
	proxys := []protocol.Proxy{{Name: "香港节点-01"}}

	result := applyAirportNodeNamePrefix(airport, proxys)
	if result[0].Name != "[A27]香港节点-01" {
		t.Fatalf("空白前缀应回退到默认前缀，实际: %s", result[0].Name)
	}
}

func TestApplyAirportIntraNodeUniquifyNumbersDuplicateNamesWithinAirport(t *testing.T) {
	airport := &models.Airport{
		NodeNameIntraUniquify: true,
	}
	proxys := []protocol.Proxy{{Name: "[A27]香港节点-01"}, {Name: "[A27]香港节点-01"}, {Name: "[A27]新加坡节点-01"}, {Name: "[A27]香港节点-01"}}

	result := applyAirportIntraNodeUniquify(airport, proxys)
	got := []string{result[0].Name, result[1].Name, result[2].Name, result[3].Name}
	want := []string{"[A27]香港节点-01-1", "[A27]香港节点-01-2", "[A27]新加坡节点-01", "[A27]香港节点-01-3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("机场内编号唯一化结果不匹配: got=%v want=%v", got, want)
	}
}

func TestApplyAirportIntraNodeUniquifyCanNumberWithoutPrefix(t *testing.T) {
	airport := &models.Airport{
		NodeNameIntraUniquify: true,
	}
	proxys := []protocol.Proxy{{Name: "香港节点-01"}, {Name: "香港节点-01"}, {Name: "日本节点-01"}}

	result := applyAirportIntraNodeUniquify(airport, proxys)
	got := []string{result[0].Name, result[1].Name, result[2].Name}
	want := []string{"香港节点-01-1", "香港节点-01-2", "日本节点-01"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("无前缀时机场内编号结果不匹配: got=%v want=%v", got, want)
	}
}

func TestGenerateProxyLinkRoundTripsMieruClashYAML(t *testing.T) {
	var config ClashConfig
	if err := yaml.Unmarshal([]byte(`proxies:
  - name: mieru-import
    type: mieru
    server: mieru.example.com
    port-range: 2090-2099
    transport: TCP
    username: user
    password: password
    multiplexing: MULTIPLEXING_LOW
    traffic-pattern: dGVzdA==
`), &config); err != nil {
		t.Fatalf("yaml unmarshal failed: %v", err)
	}
	if len(config.Proxies) != 1 {
		t.Fatalf("proxy count = %d, want 1", len(config.Proxies))
	}

	link := GenerateProxyLink(config.Proxies[0])
	if link == "" {
		t.Fatal("GenerateProxyLink returned empty link")
	}
	decoded, err := protocol.DecodeMieruURL(link)
	if err != nil {
		t.Fatalf("DecodeMieruURL failed: %v", err)
	}
	if decoded.PortRange != "2090-2099" {
		t.Fatalf("port range = %q, want 2090-2099", decoded.PortRange)
	}
	if decoded.Transport != "TCP" {
		t.Fatalf("transport = %q, want TCP", decoded.Transport)
	}
	if decoded.Multiplexing != "MULTIPLEXING_LOW" {
		t.Fatalf("multiplexing = %q, want MULTIPLEXING_LOW", decoded.Multiplexing)
	}
	if decoded.TrafficPattern != "dGVzdA==" {
		t.Fatalf("traffic pattern = %q, want dGVzdA==", decoded.TrafficPattern)
	}
}

func TestScheduleClashToNodeLinksBackfillsEveryExistingEmptyCountryNode(t *testing.T) {
	setupSubscriptionCountryBackfillTestDB(t)

	airport := &models.Airport{
		Name:                    "回填机场",
		URL:                     "https://example.com/sub.yaml",
		CronExpr:                "0 */12 * * *",
		Enabled:                 true,
		Group:                   "默认组",
		BackfillExistingCountry: true,
	}
	if err := airport.Add(); err != nil {
		t.Fatalf("add airport: %v", err)
	}
	createCountryBackfillRule(t, "HK", "香港", "香港|HK")
	createCountryBackfillRule(t, "JP", "日本", "日本|JP")

	ordinaryProxy := protocol.Proxy{Name: "HK 普通节点", Type: "trojan", Server: "ordinary.example.com", Port: 443, Password: "secret-ordinary"}
	infoHKProxy := protocol.Proxy{Name: "HK 到期时间", Type: "trojan", Server: "info.example.com", Port: 443, Password: "secret-info"}
	infoJPProxy := protocol.Proxy{Name: "JP 剩余流量", Type: "trojan", Server: "info.example.com", Port: 443, Password: "secret-info"}
	keptCountryProxy := protocol.Proxy{Name: "JP 已有国家", Type: "trojan", Server: "kept.example.com", Port: 443, Password: "secret-kept"}

	ordinary := createExistingSubscriptionNode(t, airport.ID, airport.Name, ordinaryProxy, "", 1)
	infoHK := createExistingSubscriptionNode(t, airport.ID, airport.Name, infoHKProxy, "", 2)
	infoJP := createExistingSubscriptionNode(t, airport.ID, airport.Name, infoJPProxy, "", 3)
	keptCountry := createExistingSubscriptionNode(t, airport.ID, airport.Name, keptCountryProxy, "US", 4)

	_, err := scheduleClashToNodeLinks(context.Background(), airport.ID, []protocol.Proxy{
		ordinaryProxy,
		infoHKProxy,
		infoJPProxy,
		keptCountryProxy,
	}, airport.Name, nil, nil)
	if err != nil {
		t.Fatalf("schedule clash nodes: %v", err)
	}

	assertNodeCountry(t, ordinary.ID, "HK")
	assertNodeCountry(t, infoHK.ID, "HK")
	assertNodeCountry(t, infoJP.ID, "JP")
	assertNodeCountry(t, keptCountry.ID, "US")
	assertCachedNodeCountry(t, airport.ID, ordinary.ID, "HK")
	assertCachedNodeCountry(t, airport.ID, infoHK.ID, "HK")
	assertCachedNodeCountry(t, airport.ID, infoJP.ID, "JP")
}

func setupSubscriptionCountryBackfillTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "subscription_country_backfill_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.Airport{}, &models.Node{}, &models.CountryRule{}, &models.SystemSetting{}); err != nil {
		t.Fatalf("auto migrate test db: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	if err := models.InitAirportCache(); err != nil {
		t.Fatalf("init airport cache: %v", err)
	}
	if err := models.InitNodeCache(); err != nil {
		t.Fatalf("init node cache: %v", err)
	}
	if err := models.InitCountryRuleCache(); err != nil {
		t.Fatalf("init country rule cache: %v", err)
	}
	if err := models.InitSettingCache(); err != nil {
		t.Fatalf("init setting cache: %v", err)
	}

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		testutil.CloseDB(t, db)
	})
}

func createCountryBackfillRule(t *testing.T, code string, name string, pattern string) {
	t.Helper()
	rule := &models.CountryRule{CountryCode: code, CountryName: name, Pattern: pattern, Priority: 100, Enabled: true}
	if err := rule.Add(); err != nil {
		t.Fatalf("add country rule %s: %v", code, err)
	}
}

func createExistingSubscriptionNode(t *testing.T, sourceID int, source string, proxy protocol.Proxy, country string, sourceSort int) models.Node {
	t.Helper()
	link := GenerateProxyLink(proxy)
	if link == "" {
		t.Fatalf("generate link for %s", proxy.Name)
	}
	node := models.Node{
		Name:        proxy.Name,
		LinkName:    proxy.Name,
		NameMode:    models.NodeNameModeLink,
		Link:        link,
		LinkAddress: proxy.Server + ":" + strconv.Itoa(int(proxy.Port)),
		LinkHost:    proxy.Server,
		LinkPort:    strconv.Itoa(int(proxy.Port)),
		LinkCountry: country,
		Protocol:    proxy.Type,
		Source:      source,
		SourceID:    sourceID,
		SourceSort:  sourceSort,
		ContentHash: protocol.GenerateProxyContentHash(proxy),
	}
	if err := node.Add(); err != nil {
		t.Fatalf("add existing node %s: %v", proxy.Name, err)
	}
	return node
}

func assertNodeCountry(t *testing.T, id int, want string) {
	t.Helper()
	var node models.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		t.Fatalf("load node %d: %v", id, err)
	}
	if node.LinkCountry != want {
		t.Fatalf("node %s country = %q, want %q", node.Name, node.LinkCountry, want)
	}
}

func assertCachedNodeCountry(t *testing.T, sourceID int, nodeID int, want string) {
	t.Helper()
	nodes, err := models.ListBySourceID(sourceID)
	if err != nil {
		t.Fatalf("list cached nodes: %v", err)
	}
	for _, node := range nodes {
		if node.ID == nodeID {
			if node.LinkCountry != want {
				t.Fatalf("cached node %s country = %q, want %q", node.Name, node.LinkCountry, want)
			}
			return
		}
	}
	t.Fatalf("cached node %d not found", nodeID)
}
