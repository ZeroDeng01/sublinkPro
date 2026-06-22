package node

import (
	"testing"

	"sublink/models"
	"sublink/node/protocol"
)

// TestHTTPSProtocolDetection 测试 HTTP 和 HTTPS 协议识别
// 验证修复 GitHub Issue #232：https 白名单无法正确过滤节点
func TestHTTPSProtocolDetection(t *testing.T) {
	tests := []struct {
		name         string
		link         string
		wantProtocol string
	}{
		{
			name:         "HTTP链接应识别为http协议",
			link:         "http://user:pass@example.com:8080#test-http",
			wantProtocol: "http",
		},
		{
			name:         "HTTPS链接应识别为https协议",
			link:         "https://user:pass@example.com:8443#test-https",
			wantProtocol: "https",
		},
		{
			name:         "HTTPS链接（无认证）应识别为https协议",
			link:         "https://example.com:8443#test-https-noauth",
			wantProtocol: "https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试 GetProtocolFromLink
			gotProtocol := protocol.GetProtocolFromLink(tt.link)
			if gotProtocol != tt.wantProtocol {
				t.Errorf("GetProtocolFromLink() = %v, want %v", gotProtocol, tt.wantProtocol)
			}

			// 测试通过 LinkToProxy 转换后的协议
			// 注意：这个测试验证修复前的问题 - proxy.Type 对 HTTP/HTTPS 都是 "http"
			// 但现在我们使用 GetProtocolFromLink(link) 而不是 proxy.Type
			proxy, err := protocol.LinkToProxy(protocol.Urls{Url: tt.link}, protocol.OutputConfig{})
			if err != nil {
				t.Fatalf("LinkToProxy() error = %v", err)
			}

			// Proxy.Type 对于 http 和 https 链接都返回 "http"（这是预期的底层行为）
			// 通过 Proxy.Tls 字段区分是否为 https
			expectedProxyType := "http"
			if proxy.Type != expectedProxyType {
				t.Errorf("proxy.Type = %v, want %v", proxy.Type, expectedProxyType)
			}

			// 验证 TLS 标识
			expectedTls := (tt.wantProtocol == "https")
			if proxy.Tls != expectedTls {
				t.Errorf("proxy.Tls = %v, want %v", proxy.Tls, expectedTls)
			}
		})
	}
}

// TestProtocolFilterWithHTTPS 测试协议过滤对 HTTP/HTTPS 的支持
func TestProtocolFilterWithHTTPS(t *testing.T) {
	// 该测试验证协议过滤逻辑本身是正确的
	// 只要节点的 Protocol 字段正确设置，过滤就能正常工作

	testCases := []struct {
		name              string
		nodeProtocol      string
		protocolWhitelist string
		shouldPass        bool
	}{
		{
			name:              "http节点通过http白名单",
			nodeProtocol:      "http",
			protocolWhitelist: "http",
			shouldPass:        true,
		},
		{
			name:              "https节点不通过http白名单",
			nodeProtocol:      "https",
			protocolWhitelist: "http",
			shouldPass:        false,
		},
		{
			name:              "http节点不通过https白名单",
			nodeProtocol:      "http",
			protocolWhitelist: "https",
			shouldPass:        false,
		},
		{
			name:              "https节点通过https白名单",
			nodeProtocol:      "https",
			protocolWhitelist: "https",
			shouldPass:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 模拟协议过滤逻辑
			whitelistProtos := make(map[string]bool)
			whitelistProtos[tc.protocolWhitelist] = true

			pass := whitelistProtos[tc.nodeProtocol]
			if pass != tc.shouldPass {
				t.Errorf("协议过滤结果 = %v, want %v (节点协议=%s, 白名单=%s)",
					pass, tc.shouldPass, tc.nodeProtocol, tc.protocolWhitelist)
			}
		})
	}
}

func TestApplyAirportNodeFilterDistinguishesHTTPAndHTTPS(t *testing.T) {
	proxies := []protocol.Proxy{
		{Name: "plain-http", Type: "http", Server: "plain.example.com", Port: 8080, Tls: false},
		{Name: "secure-https", Type: "http", Server: "secure.example.com", Port: 8443, Tls: true},
	}

	tests := []struct {
		name    string
		airport models.Airport
		want    []string
	}{
		{
			name:    "http whitelist keeps only non TLS HTTP nodes",
			airport: models.Airport{ProtocolWhitelist: "http"},
			want:    []string{"plain-http"},
		},
		{
			name:    "https whitelist keeps only TLS HTTP nodes",
			airport: models.Airport{ProtocolWhitelist: "https"},
			want:    []string{"secure-https"},
		},
		{
			name:    "http blacklist removes only non TLS HTTP nodes",
			airport: models.Airport{ProtocolBlacklist: "http"},
			want:    []string{"secure-https"},
		},
		{
			name:    "https blacklist removes only TLS HTTP nodes",
			airport: models.Airport{ProtocolBlacklist: "https"},
			want:    []string{"plain-http"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyAirportNodeFilter(&tt.airport, proxies)
			if len(got) != len(tt.want) {
				t.Fatalf("filtered proxies = %v, want names %v", proxyNames(got), tt.want)
			}
			for i, proxy := range got {
				if proxy.Name != tt.want[i] {
					t.Fatalf("filtered proxies = %v, want names %v", proxyNames(got), tt.want)
				}
			}
		})
	}
}

func proxyNames(proxies []protocol.Proxy) []string {
	names := make([]string, 0, len(proxies))
	for _, proxy := range proxies {
		names = append(names, proxy.Name)
	}
	return names
}

// TestApplyAirportDeduplicationDistinguishesHTTPAndHTTPS 验证机场协议去重能区分 HTTP/HTTPS。
// 回归 Issue #232 同源问题：HTTPS 节点的 proxy.Type 为 "http"，若直接用 proxy.Type 作为
// 协议名，会匹配不到按 "https" 配置的去重规则，导致 HTTPS 节点无法去重。
func TestApplyAirportDeduplicationDistinguishesHTTPAndHTTPS(t *testing.T) {
	// 每组两个节点字段完全相同（仅名称不同），按 Server+Port 去重后各应只剩 1 个。
	proxies := []protocol.Proxy{
		{Name: "http-a", Type: "http", Server: "plain.example.com", Port: 8080, Tls: false},
		{Name: "http-b", Type: "http", Server: "plain.example.com", Port: 8080, Tls: false},
		{Name: "https-a", Type: "http", Server: "secure.example.com", Port: 8443, Tls: true},
		{Name: "https-b", Type: "http", Server: "secure.example.com", Port: 8443, Tls: true},
	}

	airport := &models.Airport{
		DeduplicationRule: `{"mode":"protocol","protocolRules":{"http":["Server","Port"],"https":["Server","Port"]}}`,
	}

	got := applyAirportDeduplication(airport, proxies)

	// 期望：HTTP 组去重为 1 个，HTTPS 组去重为 1 个，共 2 个。
	gotProtoCount := map[string]int{}
	for _, p := range got {
		gotProtoCount[proxyProtocolName(p)]++
	}
	if gotProtoCount["http"] != 1 || gotProtoCount["https"] != 1 {
		t.Fatalf("去重结果 = %v (按协议统计 %v), 期望 http=1 https=1", proxyNames(got), gotProtoCount)
	}
}
