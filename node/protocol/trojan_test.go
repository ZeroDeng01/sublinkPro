package protocol

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// TestTrojanEncodeDecode 测试 Trojan 编解码完整性
func TestTrojanEncodeDecode(t *testing.T) {
	original := Trojan{
		Name:     "测试节点-Trojan",
		Password: "test-password-12345",
		Hostname: "example.com",
		Port:     443,
		Query: TrojanQuery{
			Security:    "tls",
			Type:        "ws",
			Host:        "cdn.example.com",
			Path:        "/trojan",
			Sni:         "sni.example.com",
			Fp:          "chrome",
			Fingerprint: "16dac3717024eb319093d1c95290c14adc850e2814b2208d11c7b7a436923859",
		},
	}

	// 编码
	encoded := EncodeTrojanURL(original)
	if !strings.HasPrefix(encoded, "trojan://") {
		t.Errorf("编码后应以 trojan:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeTrojanURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Hostname", original.Hostname, decoded.Hostname)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Password", original.Password, decoded.Password)
	assertEqualString(t, "Name", original.Name, decoded.Name)
	assertEqualString(t, "Query.Sni", original.Query.Sni, decoded.Query.Sni)
	assertEqualString(t, "Query.Fp", original.Query.Fp, decoded.Query.Fp)
	assertEqualString(t, "Query.Fingerprint", original.Query.Fingerprint, decoded.Query.Fingerprint)

	proxy, err := buildTrojanProxy(Urls{Url: encoded}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildTrojanProxy 失败: %v", err)
	}
	assertEqualString(t, "ProxyClientFingerprint", original.Query.Fp, proxy.Client_fingerprint)
	assertEqualString(t, "ProxyFingerprint", original.Query.Fingerprint, proxy.Fingerprint)

	t.Logf("✓ Trojan 编解码测试通过，名称: %s", decoded.Name)
}

func TestTrojanRejectsInvalidCertificateFingerprint(t *testing.T) {
	link := "trojan://password@example.com:443?security=tls&pcs=abc%2Cskip-cert-verify%3Dtrue#bad-fingerprint"

	decoded, err := DecodeTrojanURL(link)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	assertEqualString(t, "InvalidFingerprintDropped", "", decoded.Query.Fingerprint)

	proxy, err := buildTrojanProxy(Urls{Url: link}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildTrojanProxy 失败: %v", err)
	}
	assertEqualString(t, "ProxyInvalidFingerprintDropped", "", proxy.Fingerprint)
}

// TestTrojanNameModification 测试 Trojan 名称修改
func TestTrojanNameModification(t *testing.T) {
	original := Trojan{
		Name:     "原始名称",
		Password: "test-password",
		Hostname: "example.com",
		Port:     443,
		Query: TrojanQuery{
			Security: "tls",
			Type:     "tcp",
		},
	}

	newName := "新名称-Trojan-测试"
	encoded := EncodeTrojanURL(original)
	decoded, _ := DecodeTrojanURL(encoded)
	decoded.Name = newName
	reEncoded := EncodeTrojanURL(decoded)
	final, _ := DecodeTrojanURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Hostname, final.Hostname)
	assertEqualString(t, "密码(不变)", original.Password, final.Password)

	t.Logf("✓ Trojan 名称修改测试通过: %s -> %s", original.Name, final.Name)
}

func TestTrojanWebSocketOptionsRoundTrip(t *testing.T) {
	original := Trojan{
		Name:     "trojan-ws-options",
		Password: "password",
		Hostname: "example.com",
		Port:     443,
		Query: TrojanQuery{
			Type:                "ws",
			Path:                "/websocket",
			Host:                "cdn.example.com",
			Headers:             `{"Origin":"https://example.com","User-Agent":"SublinkPro"}`,
			MaxEarlyData:        2048,
			EarlyDataHeader:     "Sec-WebSocket-Protocol",
			HttpUpgrade:         1,
			HttpUpgradeFastOpen: 1,
		},
	}

	encoded := EncodeTrojanURL(original)
	const expectedURL = "trojan://password@example.com:443?type=httpupgrade&path=%2Fwebsocket%3Fed%3D2048&host=cdn.example.com&headers=%7B%22Origin%22%3A%22https%3A%2F%2Fexample.com%22%2C%22User-Agent%22%3A%22SublinkPro%22%7D&eh=Sec-WebSocket-Protocol#trojan-ws-options"
	assertEquivalentTrojanURL(t, encoded, expectedURL)
	decoded, err := DecodeTrojanURL(encoded)
	if err != nil {
		t.Fatalf("DecodeTrojanURL failed: %v", err)
	}
	if decoded.Query.Type != "httpupgrade" || decoded.Query.Path != "/websocket" || decoded.Query.MaxEarlyData != 2048 || decoded.Query.EarlyDataHeader != "Sec-WebSocket-Protocol" {
		t.Fatalf("unexpected canonical HTTPUpgrade query: %#v", decoded.Query)
	}

	proxy, err := buildTrojanProxy(Urls{Url: encoded}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildTrojanProxy failed: %v", err)
	}
	if proxy.Network != "ws" {
		t.Fatalf("network = %q, want ws", proxy.Network)
	}
	if proxy.Ws_opts["path"] != "/websocket" || proxy.Ws_opts["max-early-data"] != 2048 {
		t.Fatalf("unexpected ws opts: %#v", proxy.Ws_opts)
	}
	headers, ok := proxy.Ws_opts["headers"].(map[string]any)
	if proxy.Ws_opts["early-data-header-name"] != "Sec-WebSocket-Protocol" || proxy.Ws_opts["v2ray-http-upgrade"] != true || proxy.Ws_opts["v2ray-http-upgrade-fast-open"] != true || !ok || headers["Host"] != "cdn.example.com" || headers["Origin"] != "https://example.com" || headers["User-Agent"] != "SublinkPro" {
		t.Fatalf("unexpected canonical HTTPUpgrade fields: %#v", proxy.Ws_opts)
	}

	restored := ConvertProxyToTrojan(proxy)
	if restored.Query.Type != "httpupgrade" || restored.Query.Path != "/websocket" || restored.Query.MaxEarlyData != 2048 || restored.Query.EarlyDataHeader != "Sec-WebSocket-Protocol" {
		t.Fatalf("WebSocket proxy options did not canonicalize: got %#v", restored.Query)
	}
}

func TestTrojanGRPCOptionsRoundTrip(t *testing.T) {
	original := Trojan{
		Name:     "trojan-grpc-options",
		Password: "password",
		Hostname: "example.com",
		Port:     443,
		Query: TrojanQuery{
			Type:           "grpc",
			ServiceName:    "trojan-service",
			Mode:           "gun",
			GrpcUserAgent:  "SublinkPro/1.0",
			PingInterval:   30,
			MaxConnections: 8,
			MinStreams:     2,
		},
	}

	encoded := EncodeTrojanURL(original)
	const expectedURL = "trojan://password@example.com:443?type=grpc&serviceName=trojan-service&mode=gun&grpcUserAgent=SublinkPro%2F1.0&pingInterval=30&maxConnections=8&minStreams=2#trojan-grpc-options"
	assertEquivalentTrojanURL(t, encoded, expectedURL)
	decoded, err := DecodeTrojanURL(encoded)
	if err != nil {
		t.Fatalf("DecodeTrojanURL failed: %v", err)
	}
	if decoded.Query.Type != "grpc" || decoded.Query.ServiceName != "trojan-service" || decoded.Query.Mode != "gun" || decoded.Query.GrpcUserAgent != "SublinkPro/1.0" || decoded.Query.PingInterval != 30 || decoded.Query.MaxConnections != 8 || decoded.Query.MinStreams != 2 {
		t.Fatalf("unexpected canonical gRPC query: %#v", decoded.Query)
	}

	proxy, err := buildTrojanProxy(Urls{Url: encoded}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildTrojanProxy failed: %v", err)
	}
	if proxy.Network != "grpc" || proxy.Grpc_opts["grpc-service-name"] != "trojan-service" || proxy.Grpc_opts["grpc-mode"] != nil {
		t.Fatalf("unexpected gRPC opts: %#v", proxy.Grpc_opts)
	}
	if proxy.Grpc_opts["grpc-user-agent"] != "SublinkPro/1.0" || proxy.Grpc_opts["ping-interval"] != 30 || proxy.Grpc_opts["max-connections"] != 8 || proxy.Grpc_opts["min-streams"] != 2 {
		t.Fatalf("canonical gRPC proxy must preserve additive fields: %#v", proxy.Grpc_opts)
	}
	if len(proxy.Ws_opts) != 0 {
		t.Fatalf("gRPC transport must not emit ws opts: %#v", proxy.Ws_opts)
	}

	restored := ConvertProxyToTrojan(proxy)
	if restored.Query.Type != "grpc" || restored.Query.ServiceName != "trojan-service" || restored.Query.Mode != "gun" || restored.Query.GrpcUserAgent != "SublinkPro/1.0" || restored.Query.PingInterval != 30 || restored.Query.MaxConnections != 8 || restored.Query.MinStreams != 2 {
		t.Fatalf("gRPC proxy options did not canonicalize: got %#v", restored.Query)
	}
}

func TestTrojanGRPCMaxStreamsRoundTrip(t *testing.T) {
	original := Trojan{
		Password: "password",
		Hostname: "example.com",
		Port:     443,
		Query: TrojanQuery{
			Type:           "grpc",
			ServiceName:    "trojan-service",
			Mode:           "gun",
			MaxStreams:     16,
			MaxConnections: 4,
			MinStreams:     2,
		},
	}

	const expectedURL = "trojan://password@example.com:443?type=grpc&serviceName=trojan-service&mode=gun&maxStreams=16#example.com:443"
	assertEquivalentTrojanURL(t, EncodeTrojanURL(original), expectedURL)

	proxy := Proxy{
		Name:     "example.com:443",
		Type:     "trojan",
		Server:   "example.com",
		Port:     443,
		Password: "password",
		Network:  "grpc",
		Grpc_opts: map[string]any{
			"max-streams":     16,
			"max-connections": 4,
			"min-streams":     2,
		},
	}
	restored := ConvertProxyToTrojan(proxy)
	if restored.Query.MaxStreams != 16 {
		t.Fatalf("max streams did not survive direct proxy conversion: got %d", restored.Query.MaxStreams)
	}
	if restored.Query.MaxConnections != 0 || restored.Query.MinStreams != 0 {
		t.Fatalf("max-streams must exclude conflicting pool options: %#v", restored.Query)
	}
}

func TestTrojanGRPCMaxStreamsExcludesPoolOptions(t *testing.T) {
	opts := buildTrojanGRPCOpts(TrojanQuery{MaxConnections: 4, MinStreams: 2, MaxStreams: 16})
	if opts["max-streams"] != 16 || opts["max-connections"] != nil || opts["min-streams"] != nil {
		t.Fatalf("max-streams must exclude conflicting pool options: %#v", opts)
	}
}

func TestTrojanUnsupportedTransportDefaultsToTCP(t *testing.T) {
	decoded, err := DecodeTrojanURL("trojan://password@example.com:443?type=h2#unsupported-network")
	if err != nil {
		t.Fatalf("DecodeTrojanURL failed: %v", err)
	}
	if decoded.Query.Type != "tcp" {
		t.Fatalf("type = %q, want tcp", decoded.Query.Type)
	}

	proxy, err := buildTrojanProxy(Urls{Url: EncodeTrojanURL(decoded)}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildTrojanProxy failed: %v", err)
	}
	if proxy.Network != "tcp" || len(proxy.Ws_opts) != 0 || len(proxy.Grpc_opts) != 0 {
		t.Fatalf("unsupported transport must normalize to plain TCP: %#v", proxy)
	}
}

func TestTrojanCanonicalHTTPUpgradeURL(t *testing.T) {
	const expectedURL = "trojan://test-password@trojan-ws.example.com:443?sni=trojan-ws.example.com&type=httpupgrade&path=%2Fwebsocket%3Fed%3D2560&host=cdn.example.com#trojan-ws-full"

	decoded, err := DecodeTrojanURL(expectedURL)
	if err != nil {
		t.Fatalf("DecodeTrojanURL failed: %v", err)
	}
	if decoded.Query.Type != "httpupgrade" || decoded.Query.Path != "/websocket" || decoded.Query.MaxEarlyData != 2560 {
		t.Fatalf("unexpected HTTPUpgrade query: %#v", decoded.Query)
	}

	proxy, err := buildTrojanProxy(Urls{Url: expectedURL}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildTrojanProxy failed: %v", err)
	}
	if proxy.Network != "ws" || proxy.Ws_opts["path"] != "/websocket" || proxy.Ws_opts["max-early-data"] != 2560 || proxy.Ws_opts["v2ray-http-upgrade"] != true || proxy.Ws_opts["v2ray-http-upgrade-fast-open"] != true {
		t.Fatalf("unexpected HTTPUpgrade proxy: %#v", proxy)
	}
	assertEquivalentTrojanURL(t, EncodeTrojanURL(decoded), expectedURL)
	assertEquivalentTrojanURL(t, EncodeTrojanURL(ConvertProxyToTrojan(proxy)), expectedURL)
}

func TestTrojanHTTPUpgradeProxyUsesCanonicalURL(t *testing.T) {
	const expectedURL = "trojan://test-password@trojan-ws.example.com:443?sni=trojan-ws.example.com&type=httpupgrade&path=%2Fwebsocket%3Fed%3D2560&host=cdn.example.com&headers=%7B%22Origin%22%3A%22https%3A%2F%2Fexample.com%22%7D&eh=Sec-WebSocket-Protocol#trojan-ws-full"
	proxy := Proxy{
		Name:     "trojan-ws-full",
		Type:     "trojan",
		Server:   "trojan-ws.example.com",
		Port:     443,
		Password: "test-password",
		Sni:      "trojan-ws.example.com",
		Network:  "ws",
		Ws_opts: map[string]any{
			"path":                         "/websocket",
			"headers":                      map[string]any{"Host": "cdn.example.com", "Origin": "https://example.com"},
			"max-early-data":               2560,
			"early-data-header-name":       "Sec-WebSocket-Protocol",
			"v2ray-http-upgrade":           true,
			"v2ray-http-upgrade-fast-open": true,
		},
	}

	assertEquivalentTrojanURL(t, EncodeTrojanURL(ConvertProxyToTrojan(proxy)), expectedURL)
}

func TestTrojanCanonicalGRPCURLs(t *testing.T) {
	fixtures := []string{
		"trojan://test-password@trojan-grpc.example.com:443?sni=trojan-grpc.example.com&type=grpc&serviceName=trojan-service&mode=gun#trojan-grpc-pool",
		"trojan://test-password@trojan-grpc.example.com:443?sni=trojan-grpc.example.com&type=grpc&serviceName=trojan-service&mode=gun#trojan-grpc-max-streams",
	}

	for _, expectedURL := range fixtures {
		t.Run(expectedURL, func(t *testing.T) {
			decoded, err := DecodeTrojanURL(expectedURL)
			if err != nil {
				t.Fatalf("DecodeTrojanURL failed: %v", err)
			}
			if decoded.Query.Type != "grpc" || decoded.Query.ServiceName != "trojan-service" || decoded.Query.Mode != "gun" {
				t.Fatalf("unexpected gRPC query: %#v", decoded.Query)
			}

			proxy, err := buildTrojanProxy(Urls{Url: expectedURL}, OutputConfig{})
			if err != nil {
				t.Fatalf("buildTrojanProxy failed: %v", err)
			}
			if proxy.Network != "grpc" || proxy.Grpc_opts["grpc-service-name"] != "trojan-service" || proxy.Grpc_opts["grpc-mode"] != nil {
				t.Fatalf("unexpected gRPC proxy: %#v", proxy)
			}
			assertEquivalentTrojanURL(t, EncodeTrojanURL(decoded), expectedURL)
			assertEquivalentTrojanURL(t, EncodeTrojanURL(ConvertProxyToTrojan(proxy)), expectedURL)
		})
	}
}

func TestTrojanGRPCProxyUsesCanonicalURL(t *testing.T) {
	const expectedURL = "trojan://test-password@trojan-grpc.example.com:443?sni=trojan-grpc.example.com&type=grpc&serviceName=trojan-service&mode=gun&grpcUserAgent=SublinkPro-Test%2F1.0&pingInterval=30&maxConnections=8&minStreams=2#trojan-grpc-pool"
	proxy := Proxy{
		Name:     "trojan-grpc-pool",
		Type:     "trojan",
		Server:   "trojan-grpc.example.com",
		Port:     443,
		Password: "test-password",
		Sni:      "trojan-grpc.example.com",
		Network:  "grpc",
		Grpc_opts: map[string]any{
			"grpc-service-name": "trojan-service",
			"grpc-user-agent":   "SublinkPro-Test/1.0",
			"ping-interval":     30,
			"max-connections":   8,
			"min-streams":       2,
		},
	}

	assertEquivalentTrojanURL(t, EncodeTrojanURL(ConvertProxyToTrojan(proxy)), expectedURL)
}

func TestTrojanCanonicalTCPURL(t *testing.T) {
	const expectedURL = "trojan://test-password@trojan-tcp.example.com:443?sni=trojan-tcp.example.com&type=tcp#trojan-tcp"

	proxy, err := buildTrojanProxy(Urls{Url: expectedURL}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildTrojanProxy failed: %v", err)
	}
	if proxy.Network != "tcp" || len(proxy.Ws_opts) != 0 || len(proxy.Grpc_opts) != 0 {
		t.Fatalf("unexpected TCP proxy: %#v", proxy)
	}
	assertEquivalentTrojanURL(t, EncodeTrojanURL(ConvertProxyToTrojan(proxy)), expectedURL)
}

func TestTrojanWebSocketUpgradeWithoutFastOpenRoundTrip(t *testing.T) {
	original := Trojan{
		Password: "password",
		Hostname: "example.com",
		Port:     443,
		Query: TrojanQuery{
			Type:        "ws",
			Path:        "/ws",
			HttpUpgrade: 1,
		},
	}

	encoded := EncodeTrojanURL(original)
	decoded, err := DecodeTrojanURL(encoded)
	if err != nil {
		t.Fatalf("DecodeTrojanURL failed: %v", err)
	}
	if decoded.Query.Type != "ws" || decoded.Query.HttpUpgrade != 1 || decoded.Query.HttpUpgradeFastOpen != 0 {
		t.Fatalf("upgrade-only URL must not enable fast open: %#v", decoded.Query)
	}

	proxy, err := buildTrojanProxy(Urls{Url: encoded}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildTrojanProxy failed: %v", err)
	}
	if proxy.Ws_opts["v2ray-http-upgrade"] != true || proxy.Ws_opts["v2ray-http-upgrade-fast-open"] != nil {
		t.Fatalf("upgrade-only proxy options changed: %#v", proxy.Ws_opts)
	}
}

func TestTrojanWebSocketEmptyHostHeaderRoundTrip(t *testing.T) {
	proxy := Proxy{
		Name:     "empty-host",
		Type:     "trojan",
		Server:   "example.com",
		Port:     443,
		Password: "password",
		Network:  "ws",
		Ws_opts: map[string]any{
			"headers": map[string]any{"Host": "", "Origin": "https://example.com"},
		},
	}

	restored, err := buildTrojanProxy(Urls{Url: EncodeTrojanURL(ConvertProxyToTrojan(proxy))}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildTrojanProxy failed: %v", err)
	}
	headers, ok := restored.Ws_opts["headers"].(map[string]any)
	if !ok || headers["Host"] != "" || headers["Origin"] != "https://example.com" {
		t.Fatalf("empty Host header was not preserved: %#v", restored.Ws_opts)
	}
}

func assertEquivalentTrojanURL(t *testing.T, got, want string) {
	t.Helper()
	gotURL, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse generated URL: %v", err)
	}
	wantURL, err := url.Parse(want)
	if err != nil {
		t.Fatalf("parse expected URL: %v", err)
	}
	if gotURL.Scheme != wantURL.Scheme || gotURL.User.String() != wantURL.User.String() || gotURL.Host != wantURL.Host || gotURL.Fragment != wantURL.Fragment || !reflect.DeepEqual(gotURL.Query(), wantURL.Query()) {
		t.Fatalf("URL mismatch: got %s, want %s", got, want)
	}
}

func TestTrojanTCPDoesNotEmitTransportOptions(t *testing.T) {
	proxy, err := buildTrojanProxy(Urls{Url: EncodeTrojanURL(Trojan{
		Password: "password",
		Hostname: "example.com",
		Port:     443,
		Query: TrojanQuery{
			Type: "tcp",
			Path: "/ignored",
		},
	})}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildTrojanProxy failed: %v", err)
	}
	if len(proxy.Ws_opts) != 0 || len(proxy.Grpc_opts) != 0 {
		t.Fatalf("TCP transport must not emit nested opts: ws=%#v grpc=%#v", proxy.Ws_opts, proxy.Grpc_opts)
	}
}
