package protocol

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestVlessEncodeDecode 测试 VLESS 编解码完整性
func TestVlessEncodeDecode(t *testing.T) {
	original := VLESS{
		Name:   "测试节点-VLESS",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security:            "tls",
			Encryption:          "none",
			Type:                "ws",
			Host:                "cdn.example.com",
			Path:                "/vless",
			Sni:                 "sni.example.com",
			Fp:                  "chrome",
			Fingerprint:         "16dac3717024eb319093d1c95290c14adc850e2814b2208d11c7b7a436923859",
			Alpn:                []string{"h2", "http/1.1"},
			HttpUpgrade:         1,
			HttpUpgradeFastOpen: 1,
		},
	}

	// 编码
	encoded := EncodeVLESSURL(original)
	if !strings.HasPrefix(encoded, "vless://") {
		t.Errorf("编码后应以 vless:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeVLESSURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Server", original.Server, decoded.Server)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Uuid", original.Uuid, decoded.Uuid)
	assertEqualString(t, "Name", original.Name, decoded.Name)
	assertEqualString(t, "Query.Type", original.Query.Type, decoded.Query.Type)
	assertEqualString(t, "Query.Sni", original.Query.Sni, decoded.Query.Sni)
	assertEqualString(t, "Query.Fingerprint", original.Query.Fingerprint, decoded.Query.Fingerprint)
	assertEqualString(t, "Query.Path", original.Query.Path, decoded.Query.Path)
	assertEqualInt(t, "Query.HttpUpgrade", original.Query.HttpUpgrade, decoded.Query.HttpUpgrade)
	assertEqualInt(t, "Query.HttpUpgradeFastOpen", original.Query.HttpUpgradeFastOpen, decoded.Query.HttpUpgradeFastOpen)

	t.Logf("✓ VLESS 编解码测试通过，名称: %s", decoded.Name)
}

func TestVlessCertificateFingerprintAliases(t *testing.T) {
	const fingerprint = "16dac3717024eb319093d1c95290c14adc850e2814b2208d11c7b7a436923859"
	url := "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=tcp&fp=random&hpkp=" + fingerprint + "#证书指纹节点"

	decoded, err := DecodeVLESSURL(url)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	assertEqualString(t, "ClientFingerprint", "random", decoded.Query.Fp)
	assertEqualString(t, "CertificateFingerprint", fingerprint, decoded.Query.Fingerprint)

	proxy, err := buildVLESSProxy(Urls{Url: url}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildVLESSProxy 失败: %v", err)
	}
	assertEqualString(t, "ProxyClientFingerprint", "random", proxy.Client_fingerprint)
	assertEqualString(t, "ProxyCertificateFingerprint", fingerprint, proxy.Fingerprint)

	reEncoded := EncodeVLESSURL(decoded)
	assertContains(t, "EncodedCertificateFingerprint", reEncoded, "pcs="+fingerprint)
}

func TestVlessCertificateFingerprintPrefersValidPCS(t *testing.T) {
	const pcs = "16dac3717024eb319093d1c95290c14adc850e2814b2208d11c7b7a436923859"
	const hpkp = "78443c6c21bd739a9d3d07ecdba487aa1fe75aed671e484151a61f24d114b39d"
	url := "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=tcp&pcs=" + pcs + "&hpkp=" + hpkp + "#证书指纹优先级"

	decoded, err := DecodeVLESSURL(url)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	assertEqualString(t, "PreferPCS", pcs, decoded.Query.Fingerprint)
}

func TestVlessCertificateFingerprintFallsBackWhenPCSInvalid(t *testing.T) {
	const hpkp = "78443c6c21bd739a9d3d07ecdba487aa1fe75aed671e484151a61f24d114b39d"
	url := "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=tcp&pcs=invalid&hpkp=" + hpkp + "#证书指纹兼容"

	decoded, err := DecodeVLESSURL(url)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	assertEqualString(t, "FallbackHPKP", hpkp, decoded.Query.Fingerprint)
}

// TestVlessNameModification 测试 VLESS 名称修改
func TestVlessNameModification(t *testing.T) {
	original := VLESS{
		Name:   "原始名称",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security: "tls",
			Type:     "tcp",
		},
	}

	newName := "新名称-VLESS-测试"
	encoded := EncodeVLESSURL(original)
	decoded, _ := DecodeVLESSURL(encoded)
	decoded.Name = newName
	reEncoded := EncodeVLESSURL(decoded)
	final, _ := DecodeVLESSURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Server, final.Server)
	assertEqualString(t, "UUID(不变)", original.Uuid, final.Uuid)
	assertEqualIntInterface(t, "端口(不变)", original.Port, final.Port)

	t.Logf("✓ VLESS 名称修改测试通过: %s -> %s", original.Name, final.Name)
}

// TestVlessSpecialCharacters 测试 VLESS 特殊字符
func TestVlessSpecialCharacters(t *testing.T) {
	specialNames := []string{
		"节点 with spaces",
		"节点-with-dashes",
		"节点_with_underscores",
		"节点中文测试",
		"Node🚀Emoji",
		"Node (parentheses)",
	}

	for _, name := range specialNames {
		t.Run(name, func(t *testing.T) {
			original := VLESS{
				Name:   name,
				Uuid:   "12345678-1234-1234-1234-123456789abc",
				Server: "example.com",
				Port:   443,
				Query: VLESSQuery{
					Security: "tls",
					Type:     "tcp",
				},
			}

			encoded := EncodeVLESSURL(original)
			decoded, err := DecodeVLESSURL(encoded)
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}

			assertEqualString(t, "特殊字符名称", name, decoded.Name)
			t.Logf("✓ 特殊字符测试通过: %s", name)
		})
	}
}

// TestVlessV2rayFormat 测试 v2ray 格式 VLESS 链接解析（明文URL，非base64）
func TestVlessV2rayFormat(t *testing.T) {
	// 典型的v2ray格式VLESS链接
	testCases := []struct {
		name     string
		url      string
		expected VLESSQuery
	}{
		{
			name: "WebSocket传输层",
			url:  "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=ws&host=cdn.example.com&path=%2Fvless&sni=example.com&fp=chrome#测试节点",
			expected: VLESSQuery{
				Security:   "tls",
				Encryption: "none",
				Type:       "ws",
				Host:       "cdn.example.com",
				Path:       "/vless",
				Sni:        "example.com",
				Fp:         "chrome",
			},
		},
		{
			name: "Reality配置",
			url:  "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=reality&type=tcp&flow=xtls-rprx-vision&pbk=testpublickey&sid=testshortid&sni=example.com&fp=chrome#Reality节点",
			expected: VLESSQuery{
				Security: "reality",
				Type:     "tcp",
				Flow:     "xtls-rprx-vision",
				Pbk:      "testpublickey",
				Sid:      "testshortid",
				Sni:      "example.com",
				Fp:       "chrome",
			},
		},
		{
			name: "gRPC传输层",
			url:  "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=grpc&serviceName=mygrpc&mode=gun#gRPC节点",
			expected: VLESSQuery{
				Security:    "tls",
				Type:        "grpc",
				ServiceName: "mygrpc",
				Mode:        "gun",
			},
		},
		{
			name: "H2传输层",
			url:  "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=h2&host=example.com&path=%2Fh2path#H2节点",
			expected: VLESSQuery{
				Security: "tls",
				Type:     "h2",
				Host:     "example.com",
				Path:     "/h2path",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoded, err := DecodeVLESSURL(tc.url)
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}

			assertEqualString(t, "Security", tc.expected.Security, decoded.Query.Security)
			assertEqualString(t, "Type", tc.expected.Type, decoded.Query.Type)
			if tc.expected.Host != "" {
				assertEqualString(t, "Host", tc.expected.Host, decoded.Query.Host)
			}
			if tc.expected.Path != "" {
				assertEqualString(t, "Path", tc.expected.Path, decoded.Query.Path)
			}
			if tc.expected.Flow != "" {
				assertEqualString(t, "Flow", tc.expected.Flow, decoded.Query.Flow)
			}
			if tc.expected.Pbk != "" {
				assertEqualString(t, "Pbk", tc.expected.Pbk, decoded.Query.Pbk)
			}
			if tc.expected.ServiceName != "" {
				assertEqualString(t, "ServiceName", tc.expected.ServiceName, decoded.Query.ServiceName)
			}

			t.Logf("✓ %s 测试通过", tc.name)
		})
	}
}

// TestVlessPacketEncoding 测试 packet-encoding 参数
func TestVlessPacketEncoding(t *testing.T) {
	url := "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=tcp&packetEncoding=xudp#xudp节点"
	decoded, err := DecodeVLESSURL(url)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "PacketEncoding", "xudp", decoded.Query.PacketEncoding)
	t.Logf("✓ packet-encoding 测试通过")
}

func TestVlessTopLevelECHRoundTrip(t *testing.T) {
	original := VLESS{
		Name:   "ECH节点",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security:   "tls",
			Encryption: "none",
			Type:       "ws",
			Host:       "cdn.example.com",
			Path:       "/vless",
			Sni:        "example.com",
			Ech:        "encryptedsni.com+https://dns.alidns.com/dns-query",
		},
	}

	encoded := EncodeVLESSURL(original)
	assertContains(t, "EncodedECH", encoded, "ech=encryptedsni.com%2Bhttps%3A%2F%2Fdns.alidns.com%2Fdns-query")

	decoded, err := DecodeVLESSURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "ECH", original.Query.Ech, decoded.Query.Ech)
}

func TestVlessXHTTPURLMapping(t *testing.T) {
	extra := map[string]any{
		"headers": map[string]any{
			"User-Agent": "curl/8.0",
		},
		"noGRPCHeader":  true,
		"xPaddingBytes": "10-20",
		"downloadSettings": map[string]any{
			"path":              "/download",
			"host":              "dl.example.com",
			"tls":               true,
			"server":            "dl-backend.example.com",
			"port":              float64(8443),
			"clientFingerprint": "chrome",
		},
	}
	extraBytes, err := json.Marshal(extra)
	if err != nil {
		t.Fatalf("extra 编码失败: %v", err)
	}

	original := VLESS{
		Name:   "XHTTP节点",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security:   "tls",
			Encryption: "none",
			Type:       "xhttp",
			Host:       "cdn.example.com",
			Path:       "/xhttp",
			Mode:       "stream-up",
			Sni:        "example.com",
			Extra:      string(extraBytes),
		},
	}

	encoded := EncodeVLESSURL(original)
	assertContains(t, "EncodedType", encoded, "type=xhttp")
	assertContains(t, "EncodedMode", encoded, "mode=stream-up")
	assertContains(t, "EncodedExtra", encoded, "extra=")

	decoded, err := DecodeVLESSURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Type", "xhttp", decoded.Query.Type)
	assertEqualString(t, "Host", original.Query.Host, decoded.Query.Host)
	assertEqualString(t, "Path", original.Query.Path, decoded.Query.Path)
	assertEqualString(t, "Mode", original.Query.Mode, decoded.Query.Mode)

	decodedExtra := parseVLESSXHTTPExtra(decoded.Query.Extra)
	if decodedExtra == nil {
		t.Fatal("decoded extra 不应为空")
	}
	headers := mustMap(t, "DecodedHeader headers", decodedExtra["headers"])
	assertEqualString(t, "DecodedHeader", "curl/8.0", mustString(t, "DecodedHeader User-Agent", headers["User-Agent"]))
	assertEqualString(t, "DecodedPadding", "10-20", mustString(t, "DecodedPadding", decodedExtra["x-padding-bytes"]))
	downloadSettings := mustMap(t, "DecodedDownloadPath download-settings", decodedExtra["download-settings"])
	assertEqualString(t, "DecodedDownloadPath", "/download", mustString(t, "DecodedDownloadPath path", downloadSettings["path"]))
	assertEqualString(t, "DecodedDownloadFingerprint", "chrome", mustString(t, "DecodedDownloadFingerprint", downloadSettings["client-fingerprint"]))
}

func TestConvertProxyToVlessXHTTP(t *testing.T) {
	proxy := Proxy{
		Name:       "XHTTP节点",
		Type:       "vless",
		Server:     "example.com",
		Port:       443,
		Uuid:       "12345678-1234-1234-1234-123456789abc",
		Network:    "xhttp",
		Tls:        true,
		Encryption: "mlkem768x25519plus.native.0rtt.test-key",
		XHTTP_opts: map[string]any{
			"path": "/xhttp",
			"host": "cdn.example.com",
			"mode": "packet-up",
			"headers": map[string]any{
				"User-Agent": "curl/8.0",
			},
			"no-grpc-header": true,
			"download-settings": map[string]any{
				"path":               "/download",
				"client-fingerprint": "chrome",
			},
		},
	}

	vless := ConvertProxyToVless(proxy)
	assertEqualString(t, "Type", "xhttp", vless.Query.Type)
	assertEqualString(t, "Host", "cdn.example.com", vless.Query.Host)
	assertEqualString(t, "Path", "/xhttp", vless.Query.Path)
	assertEqualString(t, "Mode", "packet-up", vless.Query.Mode)
	assertEqualString(t, "Encryption", "mlkem768x25519plus.native.0rtt.test-key", vless.Query.Encryption)

	extra := parseVLESSXHTTPExtra(vless.Query.Extra)
	if extra == nil {
		t.Fatal("extra 不应为空")
	}
	var rawExtra map[string]any
	if err := json.Unmarshal([]byte(vless.Query.Extra), &rawExtra); err != nil {
		t.Fatalf("extra 解析失败: %v", err)
	}
	rawHeaders := mustMap(t, "ExtraHeader headers", rawExtra["headers"])
	assertEqualString(t, "ExtraHeader", "curl/8.0", mustString(t, "ExtraHeader User-Agent", rawHeaders["User-Agent"]))
	rawDownloadSettings := mustMap(t, "ExtraDownloadPath downloadSettings", rawExtra["downloadSettings"])
	assertEqualString(t, "ExtraDownloadPath", "/download", mustString(t, "ExtraDownloadPath path", rawDownloadSettings["path"]))
	assertEqualString(t, "ExtraDownloadFingerprint", "chrome", mustString(t, "ExtraDownloadFingerprint", rawDownloadSettings["clientFingerprint"]))

	encoded := EncodeVLESSURL(vless)
	assertContains(t, "EncodedType", encoded, "type=xhttp")
	assertContains(t, "EncodedEncryption", encoded, "encryption=mlkem768x25519plus.native.0rtt.test-key")
}

func TestConvertProxyToVlessPreservesTopLevelECH(t *testing.T) {
	proxy := Proxy{
		Name:       "ECH节点",
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

	vless := ConvertProxyToVless(proxy)
	assertEqualString(t, "ECH", "BASE64_ECH_CONFIG", vless.Query.Ech)

	encoded := EncodeVLESSURL(vless)
	assertContains(t, "EncodedECH", encoded, "ech=BASE64_ECH_CONFIG")
}

func TestVlessTopLevelECHMapsToECHOpts(t *testing.T) {
	original := VLESS{
		Name:   "ECH双路径节点",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security:   "tls",
			Encryption: "none",
			Type:       "xhttp",
			Host:       "cdn.example.com",
			Path:       "/xhttp",
			Sni:        "example.com",
			Ech:        "BASE64_ECH_CONFIG",
			Extra:      `{"downloadSettings":{"echOpts":{"config":"base64-ech","queryServerName":"dns.example.com"}}}`,
		},
	}

	proxy, err := buildVLESSProxy(Urls{Url: EncodeVLESSURL(original)}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildVLESSProxy 失败: %v", err)
	}

	assertEqualString(t, "TopLevelECHConfig", "BASE64_ECH_CONFIG", mustString(t, "TopLevelECHConfig", proxy.ECH_opts["config"]))
	assertEqualBool(t, "TopLevelECHEnable", true, mustBool(t, "TopLevelECHEnable", proxy.ECH_opts["enable"]))
	downloadSettings, ok := proxy.XHTTP_opts["download-settings"].(map[string]any)
	if !ok {
		t.Fatal("download-settings 不应为空")
	}
	xhttpECHOpts, ok := downloadSettings["ech-opts"].(map[string]any)
	if !ok {
		t.Fatal("ech-opts 不应为空")
	}
	assertEqualString(t, "NestedECHConfig", "base64-ech", mustString(t, "NestedECHConfig", xhttpECHOpts["config"]))
	assertEqualString(t, "NestedECHQueryServerName", "dns.example.com", mustString(t, "NestedECHQueryServerName", xhttpECHOpts["query-server-name"]))

	restored := ConvertProxyToVless(proxy)
	assertEqualString(t, "RestoredTopLevelECH", original.Query.Ech, restored.Query.Ech)
	assertContains(t, "RestoredExtra", restored.Query.Extra, "\"echOpts\"")
	assertContains(t, "RestoredExtraQueryServerName", restored.Query.Extra, "\"queryServerName\":\"dns.example.com\"")
}

func TestVlessDNSStyleECHUsesBestEffortECHOpts(t *testing.T) {
	original := VLESS{
		Name:   "ECH-DNS节点",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security: "tls",
			Type:     "ws",
			Sni:      "example.com",
			Ech:      "encryptedsni.com+https://dns.alidns.com/dns-query",
		},
	}

	proxy, err := buildVLESSProxy(Urls{Url: EncodeVLESSURL(original)}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildVLESSProxy 失败: %v", err)
	}

	assertEqualBool(t, "ECHEnable", true, mustBool(t, "ECHEnable", proxy.ECH_opts["enable"]))
	assertEqualString(t, "ECHQueryServerName", "encryptedsni.com", mustString(t, "ECHQueryServerName", proxy.ECH_opts["query-server-name"]))
	if _, exists := proxy.ECH_opts["config"]; exists {
		t.Fatalf("DNS 风格 ech 不应被错误映射为 config: %#v", proxy.ECH_opts)
	}

	restored := ConvertProxyToVless(proxy)
	assertEqualString(t, "RestoredTopLevelECH", "", restored.Query.Ech)
}

func TestLinkToProxy_VLESSXHTTPSkipCertFollowsSubscriptionConfig(t *testing.T) {
	vless := VLESS{
		Name:   "测试节点-VLESS-XHTTP-SkipCert",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security:   "tls",
			Encryption: "none",
			Type:       "xhttp",
			Host:       "cdn.example.com",
			Path:       "/xhttp",
			Mode:       "stream-one",
			Extra:      `{"downloadSettings":{"path":"/download"}}`,
		},
	}

	proxy, err := buildVLESSProxy(Urls{Url: EncodeVLESSURL(vless)}, OutputConfig{Cert: true})
	if err != nil {
		t.Fatalf("buildVLESSProxy 失败: %v", err)
	}

	assertEqualString(t, "Network", "xhttp", proxy.Network)
	assertEqualBool(t, "SkipCertVerify", true, proxy.Skip_cert_verify)
	downloadSettings, ok := proxy.XHTTP_opts["download-settings"].(map[string]any)
	if !ok {
		t.Fatal("download-settings 不应为空")
	}
	assertEqualBool(t, "DownloadSkipCertVerify", true, mustBool(t, "DownloadSkipCertVerify", downloadSettings["skip-cert-verify"]))
}

// TestVlessClashWSHeadersRoundTrip 验证 Clash ws-opts.headers 中 Host 以外的请求头（如 User-Agent）经链接往返后不丢失
func TestVlessClashWSHeadersRoundTrip(t *testing.T) {
	proxy := Proxy{
		Name:    "vless-ws-ua",
		Type:    "vless",
		Server:  "example.com",
		Port:    443,
		Uuid:    "12345678-1234-1234-1234-123456789abc",
		Network: "ws",
		Tls:     true,
		Ws_opts: map[string]any{
			"path": "/1234",
			"headers": map[string]any{
				"Host":       "abc.com",
				"User-Agent": "Mozilla/5.0 SublinkPro",
			},
		},
	}

	link := EncodeVLESSURL(ConvertProxyToVless(proxy))
	decoded, err := DecodeVLESSURL(link)
	if err != nil {
		t.Fatalf("DecodeVLESSURL failed: %v", err)
	}
	assertEqualString(t, "Host", "abc.com", decoded.Query.Host)

	out, err := LinkToProxy(Urls{Url: link}, OutputConfig{})
	if err != nil {
		t.Fatalf("LinkToProxy failed: %v", err)
	}
	headers := mustMap(t, "Ws_opts headers", out.Ws_opts["headers"])
	assertEqualString(t, "Host", "abc.com", mustString(t, "Host", headers["Host"]))
	assertEqualString(t, "User-Agent", "Mozilla/5.0 SublinkPro", mustString(t, "User-Agent", headers["User-Agent"]))
}

// Exercise the same YAML -> stored URL -> Proxy path used by imports and probes.
func TestVLESSRealityMLKEMRoundTrip(t *testing.T) {
	for _, state := range []string{"true", "false", ""} {
		t.Run("state_"+state, func(t *testing.T) {
			source := "name: reality\ntype: vless\nserver: example.com\nport: 443\nuuid: 12345678-1234-1234-1234-123456789abc\ntls: true\nclient-fingerprint: chrome\nreality-opts:\n  public-key: test-key\n  short-id: '01234567'\n"
			if state != "" {
				source += "  support-x25519mlkem768: " + state + "\n"
			}
			var original Proxy
			if err := yaml.Unmarshal([]byte(source), &original); err != nil {
				t.Fatal(err)
			}
			link := EncodeVLESSURL(ConvertProxyToVless(original))
			u, err := url.Parse(link)
			if err != nil {
				t.Fatal(err)
			}
			if got := u.Query().Get("support-x25519mlkem768"); got != state {
				t.Fatalf("URL state = %q, want %q", got, state)
			}
			rebuilt, err := buildVLESSProxy(Urls{Url: link}, OutputConfig{})
			if err != nil {
				t.Fatal(err)
			}
			if got, present := rebuilt.Reality_opts["support-x25519mlkem768"]; present != (state != "") || (present && fmt.Sprint(got) != state) {
				t.Fatalf("rebuilt state = %v, present=%v", got, present)
			}
			if rebuilt.Reality_opts["public-key"] != "test-key" || rebuilt.Reality_opts["short-id"] != "01234567" || rebuilt.Client_fingerprint != "chrome" {
				t.Fatalf("other REALITY fields lost: %+v", rebuilt)
			}
			output, err := yaml.Marshal(rebuilt)
			if err != nil {
				t.Fatal(err)
			}
			if state != "" && !strings.Contains(string(output), "support-x25519mlkem768: "+state) {
				t.Fatalf("YAML lost state: %s", output)
			}
		})
	}
}

func TestVLESSRealityMLKEMURLValues(t *testing.T) {
	for _, state := range []string{"true", "false", "1", "0", "", "invalid"} {
		t.Run(state, func(t *testing.T) {
			link := "vless://12345678-1234-1234-1234-123456789abc@example.com:443?security=reality&support-x25519mlkem768=" + state
			decoded, err := DecodeVLESSURL(link)
			if err != nil {
				t.Fatal(err)
			}
			wantPresent := state != "" && state != "invalid"
			if (decoded.Query.SupportX25519MLKEM768 != nil) != wantPresent {
				t.Fatalf("unexpected optional bool: %v", decoded.Query.SupportX25519MLKEM768)
			}
			if wantPresent && *decoded.Query.SupportX25519MLKEM768 != (state == "true" || state == "1") {
				t.Fatal("wrong boolean value")
			}
			for _, security := range []string{"tls", "none"} {
				proxy, err := buildVLESSProxy(Urls{Url: strings.Replace(link, "security=reality", "security="+security, 1)}, OutputConfig{})
				if err != nil {
					t.Fatal(err)
				}
				if len(proxy.Reality_opts) != 0 {
					t.Fatalf("%s acquired REALITY options: %v", security, proxy.Reality_opts)
				}
			}
		})
	}
}

func TestVLESSRealityMLKEMRawEditing(t *testing.T) {
	const field = "Query.SupportX25519MLKEM768"
	link := "vless://12345678-1234-1234-1234-123456789abc@example.com:443?security=reality"
	for _, state := range []string{`true`, `false`, `"true"`, `"false"`, `""`, `null`} {
		updated, err := UpdateNodeLinkFields(link, `{"`+field+`":`+state+`}`)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseNodeLink(updated)
		if err != nil {
			t.Fatal(err)
		}
		want := strings.Trim(state, `"`)
		if state == "null" {
			want = ""
		}
		if got, ok := parsed.Fields[field]; !ok || fmt.Sprint(got) != want {
			t.Fatalf("raw %s = %v, present=%v", state, got, ok)
		}
		decoded, err := DecodeVLESSURL(updated)
		if err != nil {
			t.Fatal(err)
		}
		if got := GetProtocolFieldValue(decoded, field); got != want {
			t.Fatalf("field value = %q, want %q", got, want)
		}
		link = updated
	}
	if _, err := UpdateNodeLinkFields(link, `{"`+field+`":"invalid"}`); err == nil {
		t.Fatal("expected invalid boolean error")
	}
}

func TestVLESSSurgeTransports(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		want        []string
	}{
		{"raw", "type=raw", []string{"vless, 1.2.3.4, 443", "username=550e8400-e29b-41d4-a716-446655440000"}},
		{"tls", "security=tls&allowInsecure=1&alpn=h2,http/1.1&sni=sni.example.com", []string{"tls=true", "skip-cert-verify=true", `alpn="h2,http/1.1"`, "sni=sni.example.com"}},
		{"reality", "security=reality&pbk=public-key&sid=0123&flow=xtls-rprx-vision", []string{"reality-public-key=public-key", "reality-short-id=0123", "flow=xtls-rprx-vision"}},
		{"ws", "type=ws&security=tls&path=/ws&host=cdn.example.com&headers=" + url.QueryEscape(`{"User-Agent":"test"}`), []string{"ws=true", "ws-path=/ws", "ws-headers=Host:cdn.example.com|User-Agent:test"}},
		{"upgrade", "type=httpupgrade&path=/up", []string{"ws=true", "ws-http-upgrade=true", "ws-path=/up"}},
		{"upgrade flag", "type=ws&httpUpgrade=1", []string{"ws-http-upgrade=true"}},
		{"grpc", "type=grpc&security=tls&serviceName=svc&mode=multi&authority=grpc.example.com", []string{"grpc=true", "grpc-service-name=svc", "grpc-mode=multi", "grpc-authority=grpc.example.com"}},
		{"xhttp", "type=xhttp&security=reality&pbk=key&path=/x&host=cdn.example.com&mode=stream-one&extra=" + url.QueryEscape(`{"xPaddingBytes":"10-20","noGRPCHeader":false,"xmux":{"maxConcurrency":8},"headers":{"X-Token":"abc"}}`), []string{"xhttp=true", "xhttp-path=/x", "xhttp-host=cdn.example.com", "xhttp-mode=stream-one", "xhttp-padding-bytes=10-20", "xhttp-no-grpc-header=false", "xhttp-xmux-max-concurrency=8", "xhttp-headers=X-Token:abc"}},
		{"encryption", "type=splithttp&encryption=mlkem768x25519plus.native.0rtt.key&flow=xtls-rprx-vision", []string{"xhttp=true", "encryption=mlkem768x25519plus.native.0rtt.key", "flow=xtls-rprx-vision"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			link := "vless://550e8400-e29b-41d4-a716-446655440000@example.com:443?" + tc.query + "#test"
			p, ok := detectProtocol(link).(SurgeCapable)
			if !ok {
				t.Fatal("VLESS must advertise Surge capability")
			}
			line, name, err := p.ToSurgeLine(link, OutputConfig{ReplaceServerWithHost: true, HostMap: map[string]string{"example.com": "1.2.3.4"}})
			if err != nil {
				t.Fatal(err)
			}
			if name != "test" {
				t.Fatalf("group name = %q", name)
			}
			for _, want := range tc.want {
				if !strings.Contains(line, want) {
					t.Errorf("missing %q in %s", want, line)
				}
			}
			if strings.Contains(line, "udp-relay=") || strings.Contains(line, "vmess-aead") {
				t.Fatalf("unexpected option: %s", line)
			}
		})
	}
}

func TestVLESSSurgeRejectsUnsupportedSettings(t *testing.T) {
	for _, query := range []string{
		"type=quic", "type=h2", "type=http", "security=xtls", "ech=config",
		"security=reality", "type=ws&security=reality&pbk=key", "type=grpc",
		"type=grpc&security=tls&flow=xtls-rprx-vision", "flow=unknown", "flow=xtls-rprx-vision",
		"type=ws&security=tls&flow=xtls-rprx-vision", "headerType=http",
		"type=xhttp&mode=unknown", "type=xhttp&extra=invalid", "type=ws&headers=invalid",
		"type=xhttp&extra=" + url.QueryEscape(`{"downloadSettings":{"address":"other.example.com"}}`),
		"security=tls&sni=example.com%0Areject=direct",
	} {
		t.Run(query, func(t *testing.T) {
			line, name, err := buildVLESSSurgeLine("vless://550e8400-e29b-41d4-a716-446655440000@example.com:443?"+query, OutputConfig{})
			if err == nil || line != "" || name != "" {
				t.Fatalf("expected rejection, got %q, %q, %v", line, name, err)
			}
		})
	}
}

func TestVLESSSurgeCertificateOverride(t *testing.T) {
	line, _, err := buildVLESSSurgeLine("vless://550e8400-e29b-41d4-a716-446655440000@[2001:db8::1]:443?security=tls", OutputConfig{Cert: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(line, "skip-cert-verify=true") || !strings.Contains(line, "2001:db8::1, 443") {
		t.Fatal(line)
	}
}

// REALITY extensions must survive editing without being confused with transport paths.
func TestVLESSRealityExtensionsRoundTrip(t *testing.T) {
	const base = "vless://12345678-1234-1234-1234-123456789abc@example.com:443?security=reality&type=tcp&path=%2Ftransport"
	for _, query := range []string{"", "&spx=&pqv=", "&spx=%2Fsearch%3Fq%3Da%2Bb%26lang%3D%E4%B8%AD%E6%96%87&pqv=verify_-key"} {
		t.Run(query, func(t *testing.T) {
			link := base + query
			original, err := url.Parse(link)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeVLESSURL(link)
			if err != nil {
				t.Fatal(err)
			}
			if decoded.Query.SpiderX != original.Query().Get("spx") || decoded.Query.Mldsa65Verify != original.Query().Get("pqv") {
				t.Fatalf("lost REALITY extensions: %+v", decoded.Query)
			}
			updated, err := UpdateNodeLinkFields(link, `{"Name":"renamed"}`)
			if err != nil {
				t.Fatal(err)
			}
			rebuilt, err := url.Parse(updated)
			if err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"spx", "pqv", "path"} {
				if rebuilt.Query().Get(key) != original.Query().Get(key) {
					t.Fatalf("%s changed on edit: %s", key, updated)
				}
			}
			parsed, err := ParseNodeLink(updated)
			if err != nil {
				t.Fatal(err)
			}
			for field, want := range map[string]string{"Query.SpiderX": decoded.Query.SpiderX, "Query.Mldsa65Verify": decoded.Query.Mldsa65Verify} {
				if got, ok := parsed.Fields[field]; !ok || fmt.Sprint(got) != want {
					t.Fatalf("%s = %v, present=%v, want %q", field, got, ok, want)
				}
			}
		})
	}
	link, err := UpdateNodeLinkFields(base, `{"Query.SpiderX":"/crawl?x=1&y=2","Query.Mldsa65Verify":"verify_-key"}`)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeVLESSURL(link)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Query.SpiderX != "/crawl?x=1&y=2" || decoded.Query.Mldsa65Verify != "verify_-key" {
		t.Fatalf("extensions not editable: %+v", decoded.Query)
	}
	cleared, err := UpdateNodeLinkFields(link, `{"Query.SpiderX":"","Query.Mldsa65Verify":""}`)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(cleared)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Has("spx") || u.Query().Has("pqv") {
		t.Fatalf("cleared extensions still emitted: %s", cleared)
	}
}

func TestVLESSRealityExtensionsYAMLImport(t *testing.T) {
	var proxy Proxy
	err := yaml.Unmarshal([]byte("name: reality\ntype: vless\nserver: example.com\nport: 443\nuuid: 12345678-1234-1234-1234-123456789abc\nreality-opts:\n  public-key: test-key\n  short-id: '0123'\n  spider-x: /crawl\n  mldsa65-verify: verify_-key\n"), &proxy)
	if err != nil {
		t.Fatal(err)
	}
	link, err := EncodeProxyLink(proxy)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeVLESSURL(link)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Query.Security != "reality" || decoded.Query.SpiderX != "/crawl" || decoded.Query.Mldsa65Verify != "verify_-key" {
		t.Fatalf("lost imported extensions: %+v", decoded.Query)
	}
}

func TestVLESSRealityVerificationUnsupportedTargets(t *testing.T) {
	const base = "vless://12345678-1234-1234-1234-123456789abc@example.com:443?type=tcp&pbk=test-key&sid=0123"
	for _, tc := range []struct {
		name, query string
		reject      bool
	}{
		{"verification", "&security=reality&pqv=verify_-key", true},
		{"empty verification", "&security=reality&pqv=", false},
		{"crawler only", "&security=reality&spx=%2Fcrawl", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LinkToProxy(Urls{Url: base + tc.query}, OutputConfig{})
			if err != nil {
				t.Fatalf("Clash conversion must preserve extensions: %v", err)
			}
			_, _, err = buildVLESSSurgeLine(base+tc.query, OutputConfig{})
			if (err != nil) != tc.reject {
				t.Fatalf("surge error = %v, reject=%v", err, tc.reject)
			}
			if !SupportsClientForLink(base+tc.query, ClientV2ray) {
				t.Fatal("v2ray must keep REALITY extensions")
			}
		})
	}
}

func TestRealityExtensionsClashRoundTrip(t *testing.T) {
	for _, scheme := range []string{"vless", "trojan"} {
		for _, state := range []string{"populated", "empty", "absent"} {
			t.Run(scheme+"/"+state, func(t *testing.T) {
				user := "password"
				if scheme == "vless" {
					user = "12345678-1234-1234-1234-123456789abc"
				}
				link := scheme + "://" + user + "@example.com:443?security=reality&type=tcp&pbk=test-key&sid=0123&fp=chrome&sni=cover.example.com"
				if state == "populated" {
					link += "&spx=%2Fcrawl%3Fq%3Da%2Bb%26lang%3D%E4%B8%AD%E6%96%87&pqv=verify_-key"
				}
				if state == "empty" {
					link += "&spx=&pqv="
				}
				edited, err := UpdateNodeLinkFields(link, `{"Name":"edited"}`)
				if err != nil {
					t.Fatal(err)
				}
				proxy, err := LinkToProxy(Urls{Url: edited}, OutputConfig{})
				if err != nil {
					t.Fatal(err)
				}
				data, err := yaml.Marshal(proxy)
				if err != nil {
					t.Fatal(err)
				}
				var imported Proxy
				if err := yaml.Unmarshal(data, &imported); err != nil {
					t.Fatal(err)
				}
				for key, want := range map[string]string{"public-key": "test-key", "short-id": "0123"} {
					if imported.Reality_opts[key] != want {
						t.Fatalf("%s lost: %+v", key, imported.Reality_opts)
					}
				}
				for key, want := range map[string]string{"spider-x": "/crawl?q=a+b&lang=中文", "mldsa65-verify": "verify_-key"} {
					value, present := imported.Reality_opts[key]
					if present != (state == "populated") || (present && value != want) {
						t.Fatalf("%s=%v present=%v state=%s", key, value, present, state)
					}
				}
				rebuilt, err := EncodeProxyLink(imported)
				if err != nil {
					t.Fatal(err)
				}
				u, err := url.Parse(rebuilt)
				if err != nil {
					t.Fatal(err)
				}
				original, err := url.Parse(edited)
				if err != nil {
					t.Fatal(err)
				}
				for _, key := range []string{"security", "pbk", "sid", "spx", "pqv", "sni", "fp"} {
					if u.Query().Get(key) != original.Query().Get(key) {
						t.Fatalf("%s changed through Clash: %s", key, rebuilt)
					}
				}
				// Removing fields after a YAML import must remove them from the next YAML export.
				cleared, err := UpdateNodeLinkFields(rebuilt, `{"Query.SpiderX":"","Query.Mldsa65Verify":""}`)
				if err != nil {
					t.Fatal(err)
				}
				cleanProxy, err := LinkToProxy(Urls{Url: cleared}, OutputConfig{})
				if err != nil {
					t.Fatal(err)
				}
				if _, ok := cleanProxy.Reality_opts["spider-x"]; ok {
					t.Fatal("cleared crawler reappeared")
				}
				if _, ok := cleanProxy.Reality_opts["mldsa65-verify"]; ok {
					t.Fatal("cleared verifier reappeared")
				}
			})
		}
	}
}

func TestRealityExtensionsDoNotEnableRealityOnTLS(t *testing.T) {
	for _, link := range []string{
		"vless://12345678-1234-1234-1234-123456789abc@example.com:443?security=tls&spx=%2Fcrawl&pqv=verify_-key",
		"trojan://password@example.com:443?security=tls&spx=%2Fcrawl&pqv=verify_-key",
	} {
		proxy, err := LinkToProxy(Urls{Url: link}, OutputConfig{})
		if err != nil {
			t.Fatal(err)
		}
		if len(proxy.Reality_opts) != 0 {
			t.Fatalf("plain TLS acquired REALITY options: %+v", proxy.Reality_opts)
		}
	}
}

func TestTrojanSurgeDoesNotDowngradeReality(t *testing.T) {
	for _, security := range []string{"tls", "reality"} {
		line, _, err := buildTrojanSurgeLine("trojan://password@example.com:443?security="+security+"&pbk=test-key&spx=%2Fcrawl&pqv=verify_-key", OutputConfig{})
		if security == "reality" {
			if err == nil || line != "" {
				t.Fatalf("REALITY silently downgraded to TLS: %q, %v", line, err)
			}
		} else if err != nil || line == "" {
			t.Fatalf("ordinary Trojan TLS rejected: %v", err)
		}
	}
}
