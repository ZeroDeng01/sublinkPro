package node

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

const xraySingleConfigFixture = `{
  "outbounds": [
    {
      "tag": "Германия Reality",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "de.example.com",
            "port": 443,
            "users": [{"id": "11111111-1111-1111-1111-111111111111", "flow": "xtls-rprx-vision", "encryption": "none"}]
          }
        ]
      },
      "streamSettings": {
        "network": "tcp",
        "security": "reality",
        "realitySettings": {"serverName": "www.microsoft.com", "publicKey": "REALITYPUBKEY", "shortId": "0a1b", "fingerprint": "chrome"}
      }
    },
    {
      "tag": "WS TLS",
      "protocol": "vless",
      "settings": {
        "vnext": [
          {
            "address": "ws.example.com",
            "port": "8443",
            "users": [{"id": "22222222-2222-2222-2222-222222222222", "encryption": "none"}]
          }
        ]
      },
      "streamSettings": {
        "network": "ws",
        "security": "tls",
        "tlsSettings": {"serverName": "cdn.example.com", "allowInsecure": true, "alpn": ["h2", "http/1.1"], "fingerprint": "safari"},
        "wsSettings": {"path": "/ws", "host": "cdn.example.com", "maxEarlyData": 2048, "earlyDataHeaderName": "Sec-WebSocket-Protocol"}
      }
    },
    {
      "tag": "VMess WS",
      "protocol": "vmess",
      "settings": {
        "vnext": [
          {
            "address": "vmess.example.com",
            "port": 8080,
            "users": [{"id": "33333333-3333-3333-3333-333333333333", "alterId": 0, "security": "aes-128-gcm"}]
          }
        ]
      },
      "streamSettings": {
        "network": "grpc",
        "security": "tls",
        "tlsSettings": {"serverName": "vmess.example.com"},
        "grpcSettings": {"serviceName": "grpcsvc"}
      }
    },
    {
      "tag": "Trojan Reality",
      "protocol": "trojan",
      "settings": {"servers": [{"address": "trojan.example.com", "port": 443, "password": "secret-pass"}]},
      "streamSettings": {
        "network": "ws",
        "security": "reality",
        "realitySettings": {"serverName": "www.apple.com", "publicKey": "TROJANPUBKEY", "shortId": "ff00"},
        "wsSettings": {"path": "/tr", "headers": {"Host": ["trojan-cdn.example.com"]}}
      }
    },
    {
      "tag": "SS Node",
      "protocol": "shadowsocks",
      "settings": {"servers": [{"address": "ss.example.com", "port": 8388, "method": "chacha20-ietf-poly1305", "password": "ss-pass"}]}
    },
    {"tag": "direct", "protocol": "freedom"},
    {"tag": "block", "protocol": "blackhole"},
    {"tag": "dns-out", "protocol": "dns"}
  ]
}`

func TestParseXraySubscriptionSingleConfig(t *testing.T) {
	proxies := parseXraySubscriptionProxies([]byte(xraySingleConfigFixture))
	if len(proxies) != 5 {
		t.Fatalf("expected 5 nodes, got %d", len(proxies))
	}

	byName := make(map[string]int, len(proxies))
	for index, proxy := range proxies {
		byName[proxy.Name] = index
	}
	for _, name := range []string{"Германия Reality", "WS TLS", "VMess WS", "Trojan Reality", "SS Node"} {
		if _, ok := byName[name]; !ok {
			t.Fatalf("node %q not found, got %v", name, byName)
		}
	}

	reality := proxies[byName["Германия Reality"]]
	if reality.Type != "vless" || reality.Server != "de.example.com" || reality.Port != 443 {
		t.Fatalf("unexpected reality node: %+v", reality)
	}
	if !reality.Tls || reality.Servername != "www.microsoft.com" || reality.Flow != "xtls-rprx-vision" {
		t.Fatalf("reality TLS fields not mapped: %+v", reality)
	}
	if reality.Client_fingerprint != "chrome" {
		t.Fatalf("reality fingerprint not mapped: %q", reality.Client_fingerprint)
	}
	if reality.Reality_opts["public-key"] != "REALITYPUBKEY" || reality.Reality_opts["short-id"] != "0a1b" {
		t.Fatalf("reality opts not mapped: %+v", reality.Reality_opts)
	}
	if link := GenerateProxyLink(reality); !strings.Contains(link, "security=reality") || !strings.Contains(link, "pbk=REALITYPUBKEY") {
		t.Fatalf("reality link not generated correctly: %s", link)
	}

	wsNode := proxies[byName["WS TLS"]]
	if wsNode.Port != 8443 || wsNode.Network != "ws" || !wsNode.Tls {
		t.Fatalf("ws node not mapped: %+v", wsNode)
	}
	if !wsNode.Skip_cert_verify || wsNode.Servername != "cdn.example.com" {
		t.Fatalf("ws TLS settings not mapped: %+v", wsNode)
	}
	if wsNode.Ws_opts["path"] != "/ws" || wsNode.Ws_opts["max-early-data"] != 2048 {
		t.Fatalf("ws opts not mapped: %+v", wsNode.Ws_opts)
	}
	headers, ok := wsNode.Ws_opts["headers"].(map[string]any)
	if !ok || headers["Host"] != "cdn.example.com" {
		t.Fatalf("ws host header not mapped: %+v", wsNode.Ws_opts["headers"])
	}
	if link := GenerateProxyLink(wsNode); !strings.Contains(link, "type=ws") || !strings.Contains(link, "host=cdn.example.com") {
		t.Fatalf("ws link not generated correctly: %s", link)
	}

	vmess := proxies[byName["VMess WS"]]
	if vmess.Type != "vmess" || vmess.Network != "grpc" || vmess.Cipher != "aes-128-gcm" || vmess.AlterId != "0" {
		t.Fatalf("vmess node not mapped: %+v", vmess)
	}
	if vmess.Grpc_opts["grpc-service-name"] != "grpcsvc" {
		t.Fatalf("vmess grpc opts not mapped: %+v", vmess.Grpc_opts)
	}

	trojan := proxies[byName["Trojan Reality"]]
	if trojan.Type != "trojan" || trojan.Password != "secret-pass" || trojan.Network != "ws" {
		t.Fatalf("trojan node not mapped: %+v", trojan)
	}
	if trojan.Sni != "www.apple.com" || trojan.Reality_opts["public-key"] != "TROJANPUBKEY" {
		t.Fatalf("trojan reality fields not mapped: %+v", trojan)
	}
	trojanHeaders, ok := trojan.Ws_opts["headers"].(map[string]any)
	if !ok || trojanHeaders["Host"] != "trojan-cdn.example.com" {
		t.Fatalf("trojan ws headers not mapped: %+v", trojan.Ws_opts)
	}

	ss := proxies[byName["SS Node"]]
	if ss.Type != "ss" || ss.Cipher != "chacha20-ietf-poly1305" || ss.Password != "ss-pass" {
		t.Fatalf("ss node not mapped: %+v", ss)
	}
	if link := GenerateProxyLink(ss); !strings.HasPrefix(link, "ss://") {
		t.Fatalf("ss link not generated: %s", link)
	}
}

func TestParseXraySubscriptionMultiProfileDeduplicates(t *testing.T) {
	outbound := `{
      "tag": "NL-1",
      "protocol": "vless",
      "settings": {"vnext": [{"address": "nl.example.com", "port": 443, "users": [{"id": "44444444-4444-4444-4444-444444444444", "encryption": "none"}]}]},
      "streamSettings": {"network": "tcp", "security": "tls", "tlsSettings": {"serverName": "nl.example.com"}}
    }`
	payload := `[{"remarks": "Профиль A", "outbounds": [` + outbound + `]}, {"remarks": "Профиль B", "outbounds": [` + outbound + `]}]`

	proxies := parseXraySubscriptionProxies([]byte(payload))
	if len(proxies) != 1 {
		t.Fatalf("duplicate outbounds across profiles must collapse to one node, got %d", len(proxies))
	}
	if proxies[0].Name != "Профиль A - NL-1" {
		t.Fatalf("multi-profile node name must carry profile remarks, got %q", proxies[0].Name)
	}
}

func TestParseXraySubscriptionMultiProfileNames(t *testing.T) {
	payload := `[{"remarks": "Великобритания", "outbounds": [
      {"tag": "1-gb", "protocol": "vless", "settings": {"vnext": [{"address": "gb.example.com", "port": 443, "users": [{"id": "aaaaaaaa-1111-1111-1111-111111111111", "encryption": "none"}]}]}, "streamSettings": {"network": "tcp", "security": "tls", "tlsSettings": {"serverName": "gb.example.com"}}}
    ]}, {"remarks": "Германия", "outbounds": [
      {"tag": "1-de", "protocol": "vless", "settings": {"vnext": [{"address": "de.example.com", "port": 443, "users": [{"id": "bbbbbbbb-2222-2222-2222-222222222222", "encryption": "none"}]}]}, "streamSettings": {"network": "tcp", "security": "tls", "tlsSettings": {"serverName": "de.example.com"}}}
    ]}]`

	proxies := parseXraySubscriptionProxies([]byte(payload))
	if len(proxies) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(proxies))
	}
	names := map[string]bool{}
	for _, proxy := range proxies {
		names[proxy.Name] = true
	}
	for _, name := range []string{"Великобритания - 1-gb", "Германия - 1-de"} {
		if !names[name] {
			t.Fatalf("node %q not found, got %v", name, names)
		}
	}
}

func TestParseXraySubscriptionXHTTP(t *testing.T) {
	payload := `{
      "outbounds": [
        {
          "tag": "XHTTP 节点",
          "protocol": "vless",
          "settings": {"vnext": [{"address": "xhttp.example.com", "port": 443, "users": [{"id": "55555555-5555-5555-5555-555555555555", "encryption": "none"}]}]},
          "streamSettings": {
            "network": "xhttp",
            "security": "tls",
            "tlsSettings": {"serverName": "remontnik.example"},
            "xhttpSettings": {
              "mode": "packet-up",
              "path": "/stream/",
              "xPaddingBytes": "100-1000",
              "extra": {"noGRPCHeader": true, "downloadSettings": {"serverName": "dl.example", "skipCertVerify": true}}
            }
          }
        }
      ]
    }`

	proxies := parseXraySubscriptionProxies([]byte(payload))
	if len(proxies) != 1 {
		t.Fatalf("expected 1 node, got %d", len(proxies))
	}
	proxy := proxies[0]
	if proxy.Network != "xhttp" {
		t.Fatalf("xhttp network not mapped: %+v", proxy)
	}
	if proxy.XHTTP_opts["mode"] != "packet-up" || proxy.XHTTP_opts["path"] != "/stream/" {
		t.Fatalf("xhttp opts not mapped: %+v", proxy.XHTTP_opts)
	}
	if proxy.XHTTP_opts["x-padding-bytes"] != "100-1000" {
		t.Fatalf("x-padding-bytes not mapped: %+v", proxy.XHTTP_opts)
	}
	if proxy.XHTTP_opts["no-grpc-header"] != true {
		t.Fatalf("no-grpc-header not mapped: %+v", proxy.XHTTP_opts)
	}
	if _, ok := proxy.XHTTP_opts["download-settings"].(map[string]any); !ok {
		t.Fatalf("download-settings not mapped: %+v", proxy.XHTTP_opts)
	}
	link := GenerateProxyLink(proxy)
	if !strings.Contains(link, "type=xhttp") || !strings.Contains(link, "mode=packet-up") {
		t.Fatalf("xhttp link not generated correctly: %s", link)
	}
}

func TestParseXraySubscriptionSkipsUnsupportedOutbounds(t *testing.T) {
	payload := `{
      "outbounds": [
        {
          "tag": "kcp",
          "protocol": "vless",
          "settings": {"vnext": [{"address": "kcp.example.com", "port": 443, "users": [{"id": "66666666-6666-6666-6666-666666666666", "encryption": "none"}]}]},
          "streamSettings": {"network": "kcp", "security": "none", "kcpSettings": {"mtu": 1350}}
        },
        {
          "tag": "tcp",
          "protocol": "vless",
          "settings": {"vnext": [{"address": "tcp.example.com", "port": 443, "users": [{"id": "77777777-7777-7777-7777-777777777777", "encryption": "none"}]}]},
          "streamSettings": {"network": "raw", "security": "none"}
        }
      ]
    }`

	proxies := parseXraySubscriptionProxies([]byte(payload))
	if len(proxies) != 1 {
		t.Fatalf("expected only the tcp node, got %d", len(proxies))
	}
	if proxies[0].Name != "tcp" || proxies[0].Network != "tcp" {
		t.Fatalf("unexpected node: %+v", proxies[0])
	}
}

func TestParseXraySubscriptionRejectsNonJSON(t *testing.T) {
	for _, payload := range []string{
		"",
		"proxies:\n  - name: hk\n    type: ss\n",
		"dmxlc3M6Ly91dWlkQGV4YW1wbGUuY29tOjQ0Mwc=",
		`{"code": 200, "data": []}`,
	} {
		if proxies := parseXraySubscriptionProxies([]byte(payload)); len(proxies) != 0 {
			t.Fatalf("payload %q must not produce nodes, got %d", payload, len(proxies))
		}
	}
}

func TestParseXraySubscriptionSalvagesTruncatedArray(t *testing.T) {
	payload := `[{"remarks": "A", "outbounds": [
      {"tag": "Salvaged-1", "protocol": "vless", "settings": {"vnext": [{"address": "one.example.com", "port": 443, "users": [{"id": "88888888-8888-8888-8888-888888888888", "encryption": "none"}]}]}, "streamSettings": {"network": "tcp", "security": "tls", "tlsSettings": {"serverName": "one.example.com"}}}
    ]}, {"remarks": "B", "outbounds": [
      {"tag": "Broken-2", "protocol": "vless", "settings": {"vnext": [{"address": "two.exa`

	proxies := parseXraySubscriptionProxies([]byte(payload))
	if len(proxies) != 1 {
		t.Fatalf("expected one salvaged node, got %d", len(proxies))
	}
	if proxies[0].Name != "Salvaged-1" || proxies[0].Server != "one.example.com" {
		t.Fatalf("unexpected salvaged node: %+v", proxies[0])
	}
}

func TestParseXraySubscriptionSalvagesTruncatedSingleConfig(t *testing.T) {
	payload := `{"outbounds": [
      {"tag": "First", "protocol": "vless", "settings": {"vnext": [{"address": "one.example.com", "port": 443, "users": [{"id": "99999999-9999-9999-9999-999999999999", "encryption": "none"}]}]}, "streamSettings": {"network": "tcp", "security": "tls", "tlsSettings": {"serverName": "one.example.com"}}},
      {"tag": "Second", "protocol": "vless", "settings": {"vnext": [{"address": "two.example.com", "port": 443, "users": [{"id": "aaaaaaaa-9999-9999-9999-999999999999", "encryption": "none"}]}]}, "streamSettings": {"network": "tcp", "security": "tls", "tlsSettings": {"serverName": "two.exam`

	proxies := parseXraySubscriptionProxies([]byte(payload))
	if len(proxies) != 1 {
		t.Fatalf("expected one complete outbound, got %d", len(proxies))
	}
	if proxies[0].Name != "First" || proxies[0].Server != "one.example.com" {
		t.Fatalf("unexpected salvaged node: %+v", proxies[0])
	}

	arrayPayload := `[{"remarks": "A", "outbounds": [
      {"tag": "First", "protocol": "vless", "settings": {"vnext": [{"address": "one.example.com", "port": 443, "users": [{"id": "99999999-9999-9999-9999-999999999999", "encryption": "none"}]}]}, "streamSettings": {"network": "tcp", "security": "tls", "tlsSettings": {"serverName": "one.example.com"}}}
    ]}, {"tag": "Broken", "protocol": "vless", "settings": {"vnext": [{"address`
	arrayProxies := parseXraySubscriptionProxies([]byte(arrayPayload))
	if len(arrayProxies) != 1 || arrayProxies[0].Name != "First" {
		t.Fatalf("expected one salvaged node from truncated array, got %+v", arrayProxies)
	}
}

func TestParseClashConfigDataFallsBackToXrayJSON(t *testing.T) {
	config, _, providerErr := parseClashConfigData(context.Background(), &http.Client{}, "", []byte(xraySingleConfigFixture), "", nil)
	if providerErr != nil {
		t.Fatalf("unexpected provider error: %v", providerErr)
	}
	if len(config.Proxies) != 5 {
		t.Fatalf("expected 5 proxies from JSON fallback, got %d", len(config.Proxies))
	}
}

func TestParseClashConfigDataKeepsClashYAML(t *testing.T) {
	payload := `proxies:
  - name: yaml-node
    type: ss
    server: yaml.example.com
    port: 8388
    cipher: aes-128-gcm
    password: yaml-pass
`
	config, errYaml, _ := parseClashConfigData(context.Background(), &http.Client{}, "", []byte(payload), "", nil)
	if errYaml != nil {
		t.Fatalf("unexpected YAML error: %v", errYaml)
	}
	if len(config.Proxies) != 1 || config.Proxies[0].Name != "yaml-node" {
		t.Fatalf("YAML subscription must not be affected by JSON fallback: %+v", config.Proxies)
	}
}
