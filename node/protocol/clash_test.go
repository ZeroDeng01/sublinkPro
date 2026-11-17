package protocol

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestProxyBandwidthYAMLUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		yamlDoc  string
		wantUp   int
		wantDown int
		wantErr  string
	}{
		{name: "integer", yamlDoc: "up: 30\n", wantUp: 30},
		{name: "bare string", yamlDoc: "up: \"30\"\n", wantUp: 30},
		{name: "provider bare string pair", yamlDoc: "up: \"500\"\ndown: \"1000\"\n", wantUp: 500, wantDown: 1000},
		{name: "mbps string", yamlDoc: "up: \"30 Mbps\"\n", wantUp: 30},
		{name: "invalid suffix", yamlDoc: "up: \"30 bananas\"\n", wantErr: "cannot parse bandwidth value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var proxy Proxy
			err := yaml.Unmarshal([]byte(tt.yamlDoc), &proxy)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("yaml.Unmarshal error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("yaml.Unmarshal failed: %v", err)
			}
			if got := int(proxy.Up); got != tt.wantUp {
				t.Fatalf("proxy.Up = %d, want %d", got, tt.wantUp)
			}
			if tt.wantDown != 0 {
				if got := int(proxy.Down); got != tt.wantDown {
					t.Fatalf("proxy.Down = %d, want %d", got, tt.wantDown)
				}
			}
		})
	}
}

// TestLinkToProxy_SS 测试 SS 链接转换为 Proxy 结构体
func TestLinkToProxy_SS(t *testing.T) {
	link := Urls{
		Url:             "ss://YWVzLTI1Ni1nY206dGVzdC1wYXNzd29yZA@example.com:8388#测试节点-SS",
		DialerProxyName: "",
	}
	config := OutputConfig{
		Udp:  true,
		Cert: true,
	}

	proxy, err := LinkToProxy(link, config)
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualString(t, "Type", "ss", proxy.Type)
	assertEqualString(t, "Server", "example.com", proxy.Server)
	assertEqualFlexPort(t, "Port", 8388, proxy.Port)
	assertEqualBool(t, "Udp", true, proxy.Udp)

	t.Logf("✓ SS LinkToProxy 测试通过，名称: %s", proxy.Name)
}

// TestLinkToProxy_VMess 测试 VMess 链接转换为 Proxy 结构体
func TestLinkToProxy_VMess(t *testing.T) {
	// 创建一个 VMess 节点并编码
	vmess := Vmess{
		Add:  "example.com",
		Port: "443",
		Id:   "12345678-1234-1234-1234-123456789abc",
		Net:  "ws",
		Path: "/vmess",
		Tls:  "tls",
		Ps:   "测试节点-VMess",
		V:    "2",
	}
	encoded := EncodeVmessURL(vmess)

	link := Urls{Url: encoded}
	config := OutputConfig{Udp: true, Cert: true}

	proxy, err := LinkToProxy(link, config)
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualString(t, "Type", "vmess", proxy.Type)
	assertEqualString(t, "Server", "example.com", proxy.Server)
	assertEqualFlexPort(t, "Port", 443, proxy.Port)
	assertEqualString(t, "Uuid", vmess.Id, proxy.Uuid)

	t.Logf("✓ VMess LinkToProxy 测试通过，名称: %s", proxy.Name)
}

// TestLinkToProxy_VLESS 测试 VLESS 链接转换为 Proxy 结构体
func TestLinkToProxy_VLESS(t *testing.T) {
	vless := VLESS{
		Name:   "测试节点-VLESS",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security: "tls",
			Type:     "ws",
			Path:     "/vless",
		},
	}
	encoded := EncodeVLESSURL(vless)

	link := Urls{Url: encoded}
	config := OutputConfig{Udp: true, Cert: true}

	proxy, err := LinkToProxy(link, config)
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualString(t, "Type", "vless", proxy.Type)
	assertEqualString(t, "Server", "example.com", proxy.Server)
	assertEqualFlexPort(t, "Port", 443, proxy.Port)
	assertEqualString(t, "Uuid", vless.Uuid, proxy.Uuid)

	t.Logf("✓ VLESS LinkToProxy 测试通过，名称: %s", proxy.Name)
}

func TestLinkToProxy_VLESSXHTTP(t *testing.T) {
	vless := VLESS{
		Name:   "测试节点-VLESS-XHTTP",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security:   "tls",
			Encryption: "mlkem768x25519plus.native.0rtt.test-key",
			Type:       "xhttp",
			Host:       "cdn.example.com",
			Path:       "/xhttp",
			Mode:       "stream-up",
			Extra:      `{"headers":{"User-Agent":"curl/8.0"},"noGRPCHeader":true,"downloadSettings":{"path":"/download"}}`,
		},
	}
	encoded := EncodeVLESSURL(vless)

	link := Urls{Url: encoded}
	config := OutputConfig{Udp: true, Cert: true}

	proxy, err := LinkToProxy(link, config)
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualString(t, "Type", "vless", proxy.Type)
	assertEqualString(t, "Network", "xhttp", proxy.Network)
	assertEqualString(t, "Encryption", vless.Query.Encryption, proxy.Encryption)
	assertEqualString(t, "Server", "example.com", proxy.Server)
	assertEqualFlexPort(t, "Port", 443, proxy.Port)
	assertEqualString(t, "Uuid", vless.Uuid, proxy.Uuid)
	assertEqualString(t, "XHTTPPath", "/xhttp", mustString(t, "XHTTPPath", proxy.XHTTP_opts["path"]))
	assertEqualString(t, "XHTTPHost", "cdn.example.com", mustString(t, "XHTTPHost", proxy.XHTTP_opts["host"]))
	assertEqualString(t, "XHTTPMode", "stream-up", mustString(t, "XHTTPMode", proxy.XHTTP_opts["mode"]))
	headers := mustMap(t, "XHTTPHeader headers", proxy.XHTTP_opts["headers"])
	assertEqualString(t, "XHTTPHeader", "curl/8.0", mustString(t, "XHTTPHeader User-Agent", headers["User-Agent"]))
	downloadSettings := mustMap(t, "XHTTPDownloadPath download-settings", proxy.XHTTP_opts["download-settings"])
	assertEqualString(t, "XHTTPDownloadPath", "/download", mustString(t, "XHTTPDownloadPath path", downloadSettings["path"]))

	data, err := yaml.Marshal(proxy)
	if err != nil {
		t.Fatalf("YAML 编码失败: %v", err)
	}
	if !strings.Contains(string(data), "encryption: "+vless.Query.Encryption) {
		t.Fatalf("VLESS encryption 应出现在 YAML 输出中: %s", string(data))
	}

	t.Logf("✓ VLESS XHTTP LinkToProxy 测试通过，名称: %s", proxy.Name)
}

func TestLinkToProxy_VLESSPreservesTopLevelECH(t *testing.T) {
	vless := VLESS{
		Name:   "测试节点-VLESS-ECH",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security: "tls",
			Type:     "ws",
			Path:     "/vless",
			Ech:      "BASE64_ECH_CONFIG",
		},
	}

	proxy, err := LinkToProxy(Urls{Url: EncodeVLESSURL(vless)}, OutputConfig{})
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualBool(t, "ECHEnable", true, mustBool(t, "ECHEnable", proxy.ECH_opts["enable"]))
	assertEqualString(t, "ECHConfig", vless.Query.Ech, mustString(t, "ECHConfig", proxy.ECH_opts["config"]))
	if len(proxy.XHTTP_opts) != 0 {
		t.Fatalf("顶层 ECH 不应被静默写入 xhttp-opts, 实际: %#v", proxy.XHTTP_opts)
	}
}

func TestLinkToProxy_VLESSDNSStyleECHUsesBestEffortECHOpts(t *testing.T) {
	vless := VLESS{
		Name:   "测试节点-VLESS-ECH-DNS",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security: "tls",
			Type:     "ws",
			Path:     "/vless",
			Ech:      "encryptedsni.com+https://dns.alidns.com/dns-query",
		},
	}

	proxy, err := LinkToProxy(Urls{Url: EncodeVLESSURL(vless)}, OutputConfig{})
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualBool(t, "ECHEnable", true, mustBool(t, "ECHEnable", proxy.ECH_opts["enable"]))
	assertEqualString(t, "ECHQueryServerName", "encryptedsni.com", mustString(t, "ECHQueryServerName", proxy.ECH_opts["query-server-name"]))
	if _, exists := proxy.ECH_opts["config"]; exists {
		t.Fatalf("DNS 风格 ech 不应被错误写入 config: %#v", proxy.ECH_opts)
	}
}

func TestProxyYAMLMapsTopLevelECHToECHOpts(t *testing.T) {
	proxy := Proxy{
		Name:       "测试节点-VLESS-ECH-YAML",
		Type:       "vless",
		Server:     "example.com",
		Port:       443,
		Uuid:       "12345678-1234-1234-1234-123456789abc",
		Network:    "xhttp",
		Tls:        true,
		Servername: "example.com",
		ECH_opts: map[string]any{
			"enable": true,
			"config": "BASE64_ECH_CONFIG",
		},
		XHTTP_opts: map[string]any{
			"download-settings": map[string]any{
				"ech-opts": map[string]any{
					"config":            "base64-ech",
					"query-server-name": "dns.example.com",
				},
			},
		},
	}

	data, err := yaml.Marshal(proxy)
	if err != nil {
		t.Fatalf("YAML 编码失败: %v", err)
	}

	encoded := string(data)
	if !strings.Contains(encoded, "ech-opts") {
		t.Fatalf("顶层 ECH 应映射为 ech-opts: %s", encoded)
	}
	if !strings.Contains(encoded, "config: BASE64_ECH_CONFIG") {
		t.Fatalf("顶层 ECH config 应出现在 YAML 输出中: %s", encoded)
	}
	if !strings.Contains(encoded, "query-server-name") {
		t.Fatalf("ech-opts 子字段应保留在 YAML 输出中: %s", encoded)
	}
}

// TestLinkToProxy_Trojan 测试 Trojan 链接转换为 Proxy 结构体
func TestLinkToProxy_Trojan(t *testing.T) {
	trojan := Trojan{
		Name:     "测试节点-Trojan",
		Password: "test-password",
		Hostname: "example.com",
		Port:     443,
		Query: TrojanQuery{
			Security: "tls",
			Sni:      "sni.example.com",
		},
	}
	encoded := EncodeTrojanURL(trojan)

	link := Urls{Url: encoded}
	config := OutputConfig{Udp: true, Cert: true}

	proxy, err := LinkToProxy(link, config)
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualString(t, "Type", "trojan", proxy.Type)
	assertEqualString(t, "Server", "example.com", proxy.Server)
	assertEqualFlexPort(t, "Port", 443, proxy.Port)
	assertEqualString(t, "Password", trojan.Password, proxy.Password)

	t.Logf("✓ Trojan LinkToProxy 测试通过，名称: %s", proxy.Name)
}

// TestLinkToProxy_HY2 测试 Hysteria2 链接转换为 Proxy 结构体
func TestLinkToProxy_HY2(t *testing.T) {
	hy2 := HY2{
		Name:     "测试节点-HY2",
		Host:     "example.com",
		Port:     443,
		Password: "test-password",
		Sni:      "sni.example.com",
	}
	encoded := EncodeHY2URL(hy2)

	link := Urls{Url: encoded}
	config := OutputConfig{Udp: true, Cert: true}

	proxy, err := LinkToProxy(link, config)
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualString(t, "Type", "hysteria2", proxy.Type)
	assertEqualString(t, "Server", "example.com", proxy.Server)
	assertEqualFlexPort(t, "Port", 443, proxy.Port)
	assertEqualString(t, "Password", hy2.Password, proxy.Password)

	t.Logf("✓ Hysteria2 LinkToProxy 测试通过，名称: %s", proxy.Name)
}

// TestLinkToProxy_TUIC 测试 TUIC 链接转换为 Proxy 结构体
func TestLinkToProxy_TUIC(t *testing.T) {
	tuic := Tuic{
		Name:     "测试节点-TUIC",
		Host:     "example.com",
		Port:     443,
		Uuid:     "12345678-1234-1234-1234-123456789abc",
		Password: "test-password",
	}
	encoded := EncodeTuicURL(tuic)

	link := Urls{Url: encoded}
	config := OutputConfig{Udp: true, Cert: true}

	proxy, err := LinkToProxy(link, config)
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualString(t, "Type", "tuic", proxy.Type)
	assertEqualString(t, "Server", "example.com", proxy.Server)
	assertEqualFlexPort(t, "Port", 443, proxy.Port)

	t.Logf("✓ TUIC LinkToProxy 测试通过，名称: %s", proxy.Name)
}

// TestLinkToProxy_Socks5 测试 Socks5 链接转换为 Proxy 结构体
func TestLinkToProxy_Socks5(t *testing.T) {
	socks5 := Socks5{
		Name:     "测试节点-Socks5",
		Server:   "example.com",
		Port:     1080,
		Username: "user",
		Password: "pass",
	}
	encoded := EncodeSocks5URL(socks5)

	link := Urls{Url: encoded}
	config := OutputConfig{}

	proxy, err := LinkToProxy(link, config)
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualString(t, "Type", "socks5", proxy.Type)
	assertEqualString(t, "Server", "example.com", proxy.Server)
	assertEqualFlexPort(t, "Port", 1080, proxy.Port)
	assertEqualString(t, "Username", socks5.Username, proxy.Username)

	t.Logf("✓ Socks5 LinkToProxy 测试通过，名称: %s", proxy.Name)
}

// TestLinkToProxy_AnyTLS 测试 AnyTLS 链接转换为 Proxy 结构体
func TestLinkToProxy_AnyTLS(t *testing.T) {
	anytls := AnyTLS{
		Name:     "测试节点-AnyTLS",
		Server:   "example.com",
		Port:     443,
		Password: "test-password",
		SNI:      "sni.example.com",
	}
	encoded := EncodeAnyTLSURL(anytls)

	link := Urls{Url: encoded}
	config := OutputConfig{}

	proxy, err := LinkToProxy(link, config)
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualString(t, "Type", "anytls", proxy.Type)
	assertEqualString(t, "Server", "example.com", proxy.Server)
	assertEqualFlexPort(t, "Port", 443, proxy.Port)
	assertEqualString(t, "Password", anytls.Password, proxy.Password)

	t.Logf("✓ AnyTLS LinkToProxy 测试通过，名称: %s", proxy.Name)
}

// TestLinkToProxy_SSR 测试 SSR 链接转换为 Proxy 结构体
func TestLinkToProxy_SSR(t *testing.T) {
	ssr := Ssr{
		Server:   "example.com",
		Port:     8388,
		Method:   "aes-256-cfb",
		Password: "test-password",
		Protocol: "origin",
		Obfs:     "plain",
		Qurey: Ssrquery{
			Remarks: "测试节点-SSR",
		},
	}
	encoded := EncodeSSRURL(ssr)

	link := Urls{Url: encoded}
	config := OutputConfig{Udp: true, Cert: true}

	proxy, err := LinkToProxy(link, config)
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualString(t, "Type", "ssr", proxy.Type)
	assertEqualString(t, "Server", "example.com", proxy.Server)
	assertEqualFlexPort(t, "Port", 8388, proxy.Port)
	assertEqualString(t, "Cipher", ssr.Method, proxy.Cipher)
	assertEqualString(t, "Password", ssr.Password, proxy.Password)
	assertEqualString(t, "Protocol", ssr.Protocol, proxy.Protocol)

	t.Logf("✓ SSR LinkToProxy 测试通过，名称: %s", proxy.Name)
}

// TestLinkToProxy_Hysteria 测试 Hysteria 链接转换为 Proxy 结构体
func TestLinkToProxy_Hysteria(t *testing.T) {
	hy := HY{
		Name:     "测试节点-HY",
		Host:     "example.com",
		Port:     443,
		Auth:     "auth-token",
		Peer:     "sni.example.com",
		Protocol: "udp",
		Insecure: 1,
		UpMbps:   50,
		DownMbps: 100,
		ALPN:     []string{"h3"},
	}

	link := Urls{Url: EncodeHYURL(hy)}
	proxy, err := LinkToProxy(link, OutputConfig{Cert: false})
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualString(t, "Type", "hysteria", proxy.Type)
	assertEqualString(t, "Server", hy.Host, proxy.Server)
	assertEqualString(t, "Auth", hy.Auth, proxy.Auth_str)
	assertEqualString(t, "Peer", hy.Peer, proxy.Peer)
	assertEqualBool(t, "Udp", true, proxy.Udp)
	assertEqualBool(t, "SkipCertVerify", true, proxy.Skip_cert_verify)
}

// TestLinkToProxy_HTTP 测试 HTTP/HTTPS 链接转换为 Proxy 结构体
func TestLinkToProxy_HTTP(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		tls      bool
		skipCert bool
		username string
		password string
		port     int
		sni      string
	}{
		{
			name:     "HTTP",
			url:      "http://user:pass@example.com:8080#HTTP节点",
			tls:      false,
			skipCert: false,
			username: "user",
			password: "pass",
			port:     8080,
			sni:      "",
		},
		{
			name:     "HTTPS",
			url:      "https://user:pass@example.com:8443?skip-cert-verify=true&sni=sni.example.com#HTTPS节点",
			tls:      true,
			skipCert: true,
			username: "user",
			password: "pass",
			port:     8443,
			sni:      "sni.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, err := LinkToProxy(Urls{Url: tt.url}, OutputConfig{})
			if err != nil {
				t.Fatalf("LinkToProxy 失败: %v", err)
			}

			assertEqualString(t, "Type", "http", proxy.Type)
			assertEqualString(t, "Server", "example.com", proxy.Server)
			assertEqualFlexPort(t, "Port", tt.port, proxy.Port)
			assertEqualString(t, "Username", tt.username, proxy.Username)
			assertEqualString(t, "Password", tt.password, proxy.Password)
			assertEqualBool(t, "Tls", tt.tls, proxy.Tls)
			assertEqualBool(t, "SkipCertVerify", tt.skipCert, proxy.Skip_cert_verify)
			assertEqualString(t, "Sni", tt.sni, proxy.Sni)
		})
	}
}

// TestLinkToProxy_WireGuard 测试 WireGuard 链接转换为 Proxy 结构体
func TestLinkToProxy_WireGuard(t *testing.T) {
	wg := WireGuard{
		Name:         "测试节点-WireGuard",
		Server:       "162.159.192.127",
		Port:         7152,
		PrivateKey:   "OOrigZsSjw2YaY4urjbbU4/BNOZKXqW6EYNm8XKLtkU=",
		PublicKey:    "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=",
		PreSharedKey: "31aIhAPwktDGpH4JDhA8GNvjFXEf/a6+UaQRyOAiyfM=",
		IP:           "172.16.0.2",
		IPv6:         "2606:4700:110:82ce:bdeb:e72d:572a:e280",
		MTU:          1280,
		Reserved:     []int{1, 2, 3},
	}

	proxy, err := LinkToProxy(Urls{Url: EncodeWireGuardURL(wg)}, OutputConfig{})
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	assertEqualString(t, "Type", "wireguard", proxy.Type)
	assertEqualString(t, "Server", wg.Server, proxy.Server)
	assertEqualFlexPort(t, "Port", 7152, proxy.Port)
	assertEqualString(t, "PrivateKey", wg.PrivateKey, proxy.Private_key)
	assertEqualString(t, "PublicKey", wg.PublicKey, proxy.Public_key)
	assertEqualString(t, "PreSharedKey", wg.PreSharedKey, proxy.Pre_shared_key)
	assertEqualString(t, "IP", wg.IP, proxy.Ip)
	assertEqualString(t, "IPv6", wg.IPv6, proxy.Ipv6)
	assertEqualInt(t, "MTU", wg.MTU, proxy.Mtu)
	assertEqualBool(t, "Udp", true, proxy.Udp)
}

// TestLinkToProxy_UnsupportedScheme 测试不支持的协议
func TestLinkToProxy_UnsupportedScheme(t *testing.T) {
	link := Urls{Url: "unknown://example.com:443"}
	config := OutputConfig{}

	_, err := LinkToProxy(link, config)
	if err == nil {
		t.Error("应该返回错误，因为协议不支持")
	}

	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("错误信息应该包含 'unsupported', 实际: %s", err.Error())
	}

	t.Log("✓ 不支持协议测试通过")
}

// TestLinkToProxy_HostReplacement 测试 Host 替换功能
func TestLinkToProxy_HostReplacement(t *testing.T) {
	ss := Ss{
		Name:   "测试节点",
		Server: "original.example.com",
		Port:   8388,
		Param: Param{
			Cipher:   "aes-256-gcm",
			Password: "password",
		},
	}
	encoded := EncodeSSURL(ss)

	link := Urls{Url: encoded}
	config := OutputConfig{
		ReplaceServerWithHost: true,
		HostMap: map[string]string{
			"original.example.com": "1.2.3.4",
		},
	}

	// 注意：LinkToProxy 本身不做替换，EncodeClash 中才做替换
	proxy, err := LinkToProxy(link, config)
	if err != nil {
		t.Fatalf("LinkToProxy 失败: %v", err)
	}

	// LinkToProxy 返回原始服务器地址
	assertEqualString(t, "Server", "original.example.com", proxy.Server)

	t.Log("✓ Host 替换配置测试通过")
}

// TestEncodeClash 使用真实模板验证 Clash 配置输出
func TestEncodeClash(t *testing.T) {
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "clash-template.yaml")
	template := "proxies: []\nproxy-groups:\n  - name: Proxy\n    type: select\n    proxies: []\n"
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	ss := Ss{
		Name:   "Clash-SS",
		Server: "original.example.com",
		Port:   8388,
		Param: Param{
			Cipher:   "aes-256-gcm",
			Password: "password",
		},
	}

	data, err := EncodeClash([]Urls{{Url: EncodeSSURL(ss)}}, OutputConfig{
		Clash:                 templatePath,
		Udp:                   true,
		ReplaceServerWithHost: true,
		HostMap: map[string]string{
			"original.example.com": "1.2.3.4",
		},
	})
	if err != nil {
		t.Fatalf("EncodeClash 失败: %v", err)
	}

	output := string(data)
	assertContains(t, "Clash代理名", output, "name: Clash-SS")
	assertContains(t, "Clash服务地址", output, "server: 1.2.3.4")
	assertContains(t, "Clash代理组", output, "- Clash-SS")
}

func TestEncodeClashPreservesProviderGroupsAndEmoji(t *testing.T) {
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "clash-provider-template.yaml")
	template := `proxies: []
proxy-groups:
  - name: 🛠️ 手选单一节点
    type: select
    use: [Airport-A, Airport-B]
  - name: 🇭🇰 香港全自动
    type: url-test
    use: [Airport-A, Airport-B]
    filter: '(?i)(🇭🇰|香港|Hong Kong|\bHK\b)'
    url: https://cp.cloudflare.com/generate_204
    interval: 300
    tolerance: 80
`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	ss := Ss{
		Name:   "Hong Kong-01",
		Server: "provider.example.com",
		Port:   8388,
		Param: Param{
			Cipher:   "aes-256-gcm",
			Password: "password",
		},
	}

	data, err := EncodeClash([]Urls{{Url: EncodeSSURL(ss)}}, OutputConfig{
		Clash: templatePath,
		Udp:   true,
	})
	if err != nil {
		t.Fatalf("EncodeClash 失败: %v", err)
	}

	output := string(data)
	assertContains(t, "provider组use字段", output, "Airport-A")

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("解析输出 YAML 失败: %v", err)
	}

	rawGroups, ok := config["proxy-groups"].([]any)
	if !ok || len(rawGroups) != 2 {
		t.Fatalf("proxy-groups 解析失败: %#v", config["proxy-groups"])
	}

	firstGroup, ok := rawGroups[0].(map[string]any)
	if !ok {
		t.Fatalf("第一个代理组类型错误: %#v", rawGroups[0])
	}
	if firstGroup["name"] != "🛠️ 手选单一节点" {
		t.Fatalf("第一个代理组名称错误: %#v", firstGroup["name"])
	}
	if _, exists := firstGroup["proxies"]; exists {
		t.Fatalf("use 组不应被强行注入 proxies: %#v", firstGroup)
	}
	if use, ok := firstGroup["use"].([]any); !ok || len(use) != 2 {
		t.Fatalf("use 组 providers 丢失: %#v", firstGroup["use"])
	}

	secondGroup, ok := rawGroups[1].(map[string]any)
	if !ok {
		t.Fatalf("第二个代理组类型错误: %#v", rawGroups[1])
	}
	if secondGroup["name"] != "🇭🇰 香港全自动" {
		t.Fatalf("第二个代理组名称错误: %#v", secondGroup["name"])
	}
	if _, exists := secondGroup["proxies"]; exists {
		t.Fatalf("带 filter 的 provider 组不应被强行注入 proxies: %#v", secondGroup)
	}
	if secondGroup["filter"] != "(?i)(🇭🇰|香港|Hong Kong|\\bHK\\b)" {
		t.Fatalf("filter 被意外改写: %#v", secondGroup["filter"])
	}
}

func TestEncodeClashPreservesMihomoDynamicGroups(t *testing.T) {
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "clash-mihomo-dynamic-template.yaml")
	template := `proxies: []
proxy-groups:
  - name: ♻️ 自动选择
    type: url-test
    url: http://www.gstatic.com/generate_204
    interval: 300
    tolerance: 50
    include-all-proxies: true
  - name: ⚖️ 负载均衡
    type: load-balance
    url: https://www.gstatic.com/generate_204
    interval: 300
    strategy: round-robin
    include-all-proxies: true
    filter: '(BWG|bwg)'
  - name: 📦 Provider全集
    type: select
    include-all-providers: true
`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	ss := Ss{
		Name:   "BWG-HK-01",
		Server: "provider.example.com",
		Port:   8388,
		Param: Param{
			Cipher:   "aes-256-gcm",
			Password: "password",
		},
	}

	data, err := EncodeClash([]Urls{{Url: EncodeSSURL(ss)}}, OutputConfig{
		Clash: templatePath,
		Udp:   true,
	})
	if err != nil {
		t.Fatalf("EncodeClash 失败: %v", err)
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("解析输出 YAML 失败: %v", err)
	}

	rawGroups, ok := config["proxy-groups"].([]any)
	if !ok || len(rawGroups) != 3 {
		t.Fatalf("proxy-groups 解析失败: %#v", config["proxy-groups"])
	}

	firstGroup, ok := rawGroups[0].(map[string]any)
	if !ok {
		t.Fatalf("第一个代理组类型错误: %#v", rawGroups[0])
	}
	if _, exists := firstGroup["proxies"]; exists {
		t.Fatalf("include-all-proxies 组不应被强行注入 proxies: %#v", firstGroup)
	}
	if includeAllProxies, ok := firstGroup["include-all-proxies"].(bool); !ok || !includeAllProxies {
		t.Fatalf("include-all-proxies 丢失: %#v", firstGroup["include-all-proxies"])
	}

	secondGroup, ok := rawGroups[1].(map[string]any)
	if !ok {
		t.Fatalf("第二个代理组类型错误: %#v", rawGroups[1])
	}
	if _, exists := secondGroup["proxies"]; exists {
		t.Fatalf("动态负载均衡组不应被强行注入 proxies: %#v", secondGroup)
	}
	if includeAllProxies, ok := secondGroup["include-all-proxies"].(bool); !ok || !includeAllProxies {
		t.Fatalf("load-balance 组 include-all-proxies 丢失: %#v", secondGroup["include-all-proxies"])
	}

	thirdGroup, ok := rawGroups[2].(map[string]any)
	if !ok {
		t.Fatalf("第三个代理组类型错误: %#v", rawGroups[2])
	}
	if _, exists := thirdGroup["proxies"]; exists {
		t.Fatalf("include-all-providers 组不应被强行注入 proxies: %#v", thirdGroup)
	}
	if includeAllProviders, ok := thirdGroup["include-all-providers"].(bool); !ok || !includeAllProviders {
		t.Fatalf("include-all-providers 丢失: %#v", thirdGroup["include-all-providers"])
	}
}

// TestEncodeClashExpandsPlaceholderInFilterGroup 验证带 filter / include-all-proxies 的
// 自动匹配组里显式写下的 __ALL_PROXIES__ 仍会就地展开，占位符不能原样发给客户端。
func TestEncodeClashExpandsPlaceholderInFilterGroup(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "clash-filter-placeholder-template.yaml")
	template := `proxies: ~
proxy-groups:
  - name: ChatGPT
    type: select
    include-all-proxies: true
    filter: '(?i)ai-chatgpt'
    exclude-filter: '(?i)ai-chatgpt-blocked'
    proxies:
      - DIRECT
      - __ALL_PROXIES__
  - name: Auto
    type: url-test
    url: http://www.gstatic.com/generate_204
    include-all: true
    filter: '(?i)hk'
    proxies: []
rules:
  - MATCH,Auto
`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}
	output := string(data)

	if strings.Contains(output, "__ALL_PROXIES__") {
		t.Fatalf("占位符未被展开: %s", output)
	}
	// 占位符前后的成员与所在位置保持不变
	assertContains(t, "占位符展开", output, "    proxies:\n      - DIRECT\n      - Node-A\n")
	assertContains(t, "filter 保留", output, "    filter: '(?i)ai-chatgpt'\n")
	// 自动匹配组本身不能补全成员
	assertContains(t, "空列表保持原样", output, "    filter: '(?i)hk'\n    proxies: []\n")

	var parsed struct {
		ProxyGroups []struct {
			Name              string   `yaml:"name"`
			Proxies           []string `yaml:"proxies"`
			IncludeAllProxies bool     `yaml:"include-all-proxies"`
			IncludeAll        bool     `yaml:"include-all"`
			Filter            string   `yaml:"filter"`
		} `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("输出不是合法 YAML: %v", err)
	}
	if len(parsed.ProxyGroups) != 2 {
		t.Fatalf("代理组数量错误: %#v", parsed.ProxyGroups)
	}
	chatgpt := parsed.ProxyGroups[0]
	if !chatgpt.IncludeAllProxies || chatgpt.Filter != "(?i)ai-chatgpt" {
		t.Fatalf("自动匹配字段被改写: %#v", chatgpt)
	}
	if len(chatgpt.Proxies) != 2 || chatgpt.Proxies[0] != "DIRECT" || chatgpt.Proxies[1] != "Node-A" {
		t.Fatalf("占位符展开结果错误: %#v", chatgpt.Proxies)
	}
	if auto := parsed.ProxyGroups[1]; !auto.IncludeAll || len(auto.Proxies) != 0 {
		t.Fatalf("include-all 组被改写: %#v", auto)
	}
}

// TestDecodeClashExpandsPlaceholderInFilterGroupFallback 用流式 proxy-groups 强制走
// 节点树重编码路径，确认自动匹配组里的 __ALL_PROXIES__ 同样会被展开。
func TestDecodeClashExpandsPlaceholderInFilterGroupFallback(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "clash-flow-groups-template.yaml")
	template := "proxies: ~\nproxy-groups: [{name: ChatGPT, type: select, include-all-proxies: true, " +
		"filter: '(?i)ai-chatgpt', proxies: [DIRECT, __ALL_PROXIES__]}]\nrules:\n  - MATCH,ChatGPT\n"
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}
	output := string(data)

	if strings.Contains(output, "__ALL_PROXIES__") {
		t.Fatalf("占位符未被展开: %s", output)
	}

	var parsed struct {
		ProxyGroups []struct {
			Proxies           []string `yaml:"proxies"`
			IncludeAllProxies bool     `yaml:"include-all-proxies"`
		} `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("输出不是合法 YAML: %v", err)
	}
	if len(parsed.ProxyGroups) != 1 {
		t.Fatalf("代理组数量错误: %#v", parsed.ProxyGroups)
	}
	group := parsed.ProxyGroups[0]
	if !group.IncludeAllProxies {
		t.Fatalf("include-all-proxies 丢失: %#v", group)
	}
	if len(group.Proxies) != 2 || group.Proxies[0] != "DIRECT" || group.Proxies[1] != "Node-A" {
		t.Fatalf("占位符展开结果错误: %#v", group.Proxies)
	}
}

func TestEncodeClashPreservesTemplateCommentsAndKeyOrder(t *testing.T) {
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "clash-comment-template.yaml")
	template := `# 订阅模板：顶部注释
mixed-port: 7890
allow-lan: true
# 节点列表注释
proxies: ~
# 代理组注释
proxy-groups:
  # 手选组
  - name: 🚀 节点选择
    # 组内注释
    type: select
    proxies: []
rules:
  # 规则注释
  - MATCH,DIRECT
`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	ss := Ss{
		Name:   "Comment-SS",
		Server: "example.com",
		Port:   8388,
		Param: Param{
			Cipher:   "aes-256-gcm",
			Password: "password",
		},
	}

	data, err := EncodeClash([]Urls{{Url: EncodeSSURL(ss)}}, OutputConfig{
		Clash: templatePath,
		Udp:   true,
	})
	if err != nil {
		t.Fatalf("EncodeClash 失败: %v", err)
	}

	output := string(data)
	for _, comment := range []string{
		"# 订阅模板：顶部注释",
		"# 节点列表注释",
		"# 代理组注释",
		"# 手选组",
		"# 组内注释",
		"# 规则注释",
	} {
		assertContains(t, "模板注释", output, comment)
	}
	assertContains(t, "代理组节点", output, "- Comment-SS")

	// 顶层键顺序必须与模板一致，不能被字母序重排
	mixedIndex := strings.Index(output, "mixed-port:")
	allowLanIndex := strings.Index(output, "allow-lan:")
	proxiesIndex := strings.Index(output, "\nproxies:")
	groupsIndex := strings.Index(output, "\nproxy-groups:")
	rulesIndex := strings.Index(output, "\nrules:")
	if mixedIndex < 0 || allowLanIndex < 0 || proxiesIndex < 0 || groupsIndex < 0 || rulesIndex < 0 {
		t.Fatalf("输出缺少模板键: %s", output)
	}
	if mixedIndex > allowLanIndex || allowLanIndex > proxiesIndex || proxiesIndex > groupsIndex || groupsIndex > rulesIndex {
		t.Fatalf("模板键顺序被改写: %s", output)
	}
}

func TestEncodeClashKeepsEmojiLiteral(t *testing.T) {
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "clash-emoji-template.yaml")
	template := `# 顶部注释
mixed-port: 7890
proxies: ~
proxy-groups:
  # 手选组
  - name: 🚀 节点选择
    type: select
    proxies: [__ALL_PROXIES__, DIRECT]
  - name: 🇭🇰 香港节点
    type: select
    proxies: []
rules:
  - MATCH,🚀 节点选择
`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	ss := Ss{
		Name:   "🇯🇵 东京 01",
		Server: "example.com",
		Port:   8388,
		Param: Param{
			Cipher:   "aes-256-gcm",
			Password: "password",
		},
	}

	data, err := EncodeClash([]Urls{{Url: EncodeSSURL(ss)}}, OutputConfig{Clash: templatePath, Udp: true})
	if err != nil {
		t.Fatalf("EncodeClash 失败: %v", err)
	}

	output := string(data)
	for _, literal := range []string{"🚀 节点选择", "🇭🇰 香港节点", "🇯🇵 东京 01"} {
		assertContains(t, "模板 emoji", output, literal)
	}
	if strings.Contains(output, `\U0001F`) {
		t.Fatalf("emoji 不应被转义成 \\U 序列: %s", output)
	}

	// 重新解析后名称必须与模板和节点名一致
	var parsed struct {
		Proxies []struct {
			Name string `yaml:"name"`
		} `yaml:"proxies"`
		ProxyGroups []struct {
			Name string `yaml:"name"`
		} `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("输出不是合法 YAML: %v", err)
	}
	if len(parsed.Proxies) != 1 || parsed.Proxies[0].Name != "🇯🇵 东京 01" {
		t.Fatalf("节点名称丢失 emoji: %#v", parsed.Proxies)
	}
	if len(parsed.ProxyGroups) != 2 || parsed.ProxyGroups[0].Name != "🚀 节点选择" || parsed.ProxyGroups[1].Name != "🇭🇰 香港节点" {
		t.Fatalf("代理组名称丢失 emoji: %#v", parsed.ProxyGroups)
	}
}

func TestEncodeClashKeepsLiteralBackslashSequences(t *testing.T) {
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "clash-backslash-template.yaml")
	template := "proxies: ~\nproxy-groups:\n  - name: 手动切换\n    type: select\n    proxies: []\n"
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	name := `token\U0001F680name`
	ss := Ss{
		Name:   name,
		Server: "example.com",
		Port:   8388,
		Param: Param{
			Cipher:   "aes-256-gcm",
			Password: "password",
		},
	}

	data, err := EncodeClash([]Urls{{Url: EncodeSSURL(ss)}}, OutputConfig{Clash: templatePath, Udp: true})
	if err != nil {
		t.Fatalf("EncodeClash 失败: %v", err)
	}

	var parsed struct {
		Proxies []struct {
			Name string `yaml:"name"`
		} `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("输出不是合法 YAML: %v", err)
	}
	if len(parsed.Proxies) != 1 || parsed.Proxies[0].Name != name {
		t.Fatalf("字面反斜杠序列被改写: %#v", parsed.Proxies)
	}
}

func TestDecodeClashAppendsCustomGroupsAfterTemplate(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "clash-custom-group-template.yaml")
	template := `proxies: ~
proxy-groups:
  # 模板组保持在前
  - name: 手动切换
    type: select
    proxies: [__ALL_PROXIES__]
`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
		[]CustomProxyGroup{
			{Name: "链式-手选", Type: "select", Proxies: []string{"Node-A", "DIRECT"}},
			{Name: "链式-测速", Type: "url-test", Proxies: []string{"Node-A"}},
			{Name: "链式-均衡", Type: "load-balance", Proxies: []string{"Node-A"}, Strategy: "round-robin"},
		},
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}

	var parsed struct {
		ProxyGroups []struct {
			Name      string   `yaml:"name"`
			Type      string   `yaml:"type"`
			Proxies   []string `yaml:"proxies"`
			URL       string   `yaml:"url"`
			Interval  int      `yaml:"interval"`
			Tolerance int      `yaml:"tolerance"`
			Strategy  string   `yaml:"strategy"`
		} `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("输出不是合法 YAML: %v", err)
	}

	want := []string{"手动切换", "链式-手选", "链式-测速", "链式-均衡"}
	if len(parsed.ProxyGroups) != len(want) {
		t.Fatalf("自定义组应追加在模板组之后: %#v", parsed.ProxyGroups)
	}
	for index, name := range want {
		if parsed.ProxyGroups[index].Name != name {
			t.Fatalf("代理组顺序不符: got %s want %s", parsed.ProxyGroups[index].Name, name)
		}
	}
	if len(parsed.ProxyGroups[0].Proxies) != 1 || parsed.ProxyGroups[0].Proxies[0] != "Node-A" {
		t.Fatalf("模板组占位符未展开: %#v", parsed.ProxyGroups[0].Proxies)
	}

	speed := parsed.ProxyGroups[2]
	if speed.URL != "http://www.gstatic.com/generate_204" || speed.Interval != 300 || speed.Tolerance != 50 {
		t.Fatalf("url-test 默认值错误: %#v", speed)
	}
	balance := parsed.ProxyGroups[3]
	if balance.URL != "http://www.gstatic.com/generate_204" || balance.Interval != 300 || balance.Strategy != "round-robin" {
		t.Fatalf("load-balance 配置错误: %#v", balance)
	}
	if strings.Contains(string(data), "_custom_group") {
		t.Fatalf("自定义组不应输出内部标记: %s", data)
	}
}

// TestEncodeClashKeepsTemplateTextVerbatim 验证模板文本按原样保留：
// 注释、空行、缩进、行内分隔空格和代理组写法都不被 YAML 编码器归一化。
func TestEncodeClashKeepsTemplateTextVerbatim(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "clash-verbatim-template.yaml")
	template := `port: 7890
socks-port: 7891
allow-lan: true
#Здесь ставим dns
dns:
  enabled: true
  nameserver:
    - 119.29.29.29
proxies: ~

proxy-groups:
  - name:  节点选择
    type: select
    proxies:
      -  自动选择
      -  手动切换
      - DIRECT
  - name:  手动切换
    type: select
    proxies:
      - __ALL_PROXIES__
  - name:  香港节点
    type: url-test
    url: http://www.gstatic.com/generate_204
    interval: 300
    tolerance: 50
    include-all: true
    filter: (港|HK|hk|Hong Kong)
    proxies:

rules:
  - RULE-SET,LocalAreaNetwork, 节点选择
  - MATCH, 节点选择

rule-providers:
  LocalAreaNetwork:
    type: http
    behavior: classical
    url: https://example.com/LocalAreaNetwork.list
    interval: 86400
`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}
	output := string(data)

	// 注释行与 dns 之间不应凭空多出空行
	assertContains(t, "注释紧贴 dns", output, "#Здесь ставим dns\ndns:\n  enabled: true\n")
	// proxies 之前的内容逐字保留
	prefix := "port: 7890\nsocks-port: 7891\nallow-lan: true\n#Здесь ставим dns\ndns:\n  enabled: true\n" +
		"  nameserver:\n    - 119.29.29.29\nproxies:"
	if !strings.HasPrefix(output, prefix) {
		t.Fatalf("proxies 之前的模板内容被改写:\n%s", output)
	}
	// 节点写进模板 proxies 的块序列，空行与缩进保持模板原样
	assertContains(t, "proxies 块序列", output, "proxies:\n  - name: Node-A\n")
	assertContains(t, "proxies 与 proxy-groups 之间空行", output, "\n\nproxy-groups:\n  - name:  节点选择")
	// 未改动的组内容逐字保留
	assertContains(t, "手选组成员", output, "    proxies:\n      -  自动选择\n      -  手动切换\n      - DIRECT\n")
	assertContains(t, "占位符就地展开", output, "  - name:  手动切换\n    type: select\n    proxies:\n      - Node-A\n")
	assertContains(t, "include-all 组保持原样", output, "    filter: (港|HK|hk|Hong Kong)\n    proxies:\n\nrules:")

	// rules 之后的内容逐字保留，包括空行
	suffix := "\n\nrules:\n  - RULE-SET,LocalAreaNetwork, 节点选择\n  - MATCH, 节点选择\n\n" +
		"rule-providers:\n  LocalAreaNetwork:\n    type: http\n    behavior: classical\n" +
		"    url: https://example.com/LocalAreaNetwork.list\n    interval: 86400\n"
	if !strings.HasSuffix(output, suffix) {
		t.Fatalf("rules 之后的模板内容被改写:\n%s", output)
	}

	var parsed struct {
		Proxies []struct {
			Name string `yaml:"name"`
		} `yaml:"proxies"`
		ProxyGroups []struct {
			Name       string   `yaml:"name"`
			Proxies    []string `yaml:"proxies"`
			IncludeAll bool     `yaml:"include-all"`
		} `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("输出不是合法 YAML: %v", err)
	}
	if len(parsed.Proxies) != 1 || parsed.Proxies[0].Name != "Node-A" {
		t.Fatalf("节点未写入模板: %#v", parsed.Proxies)
	}
	if len(parsed.ProxyGroups) != 3 {
		t.Fatalf("代理组数量错误: %#v", parsed.ProxyGroups)
	}
	if got := parsed.ProxyGroups[1].Proxies; len(got) != 1 || got[0] != "Node-A" {
		t.Fatalf("占位符展开错误: %#v", got)
	}
	if !parsed.ProxyGroups[2].IncludeAll || len(parsed.ProxyGroups[2].Proxies) != 0 {
		t.Fatalf("include-all 组被改写: %#v", parsed.ProxyGroups[2])
	}
}

// TestEncodeClashKeepsGroupKeyLineComment 验证组内 proxies 键行上的注释不会丢失。
func TestEncodeClashKeepsGroupKeyLineComment(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "clash-key-comment-template.yaml")
	template := "proxies: ~\nproxy-groups:\n  - name: 手动切换\n    type: select\n    proxies:  # 手动补充\n"
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}

	assertContains(t, "组内键行注释", string(data), "    proxies: # 手动补充\n      - Node-A\n")
}

// TestDecodeClashAppendsCustomGroupsKeepsBlankLines 验证自定义组追加后模板空行仍然保留。
func TestDecodeClashAppendsCustomGroupsKeepsBlankLines(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "clash-custom-layout-template.yaml")
	template := "proxies: ~\nproxy-groups:\n  - name: 手动切换\n    type: select\n    proxies: [__ALL_PROXIES__]\n\n" +
		"rules:\n  - MATCH,DIRECT\n"
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
		[]CustomProxyGroup{{Name: "链式-手选", Type: "select", Proxies: []string{"Node-A", "DIRECT"}}},
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}
	output := string(data)

	assertContains(t, "自定义组追加位置", output, "\n  - name: 链式-手选\n")
	assertContains(t, "自定义组节点", output, "      - Node-A\n      - DIRECT\n")
	assertContains(t, "模板空行保留", output, "\n\nrules:\n  - MATCH,DIRECT\n")
}

// TestEncodeClashKeepsDefaultTemplateLayout 用仓库自带的默认模板验证空行、注释
// 和固定成员列表都按原文保留，只有节点列表与占位符会被替换。
func TestEncodeClashKeepsDefaultTemplateLayout(t *testing.T) {
	templatePath := filepath.Join("..", "..", "template", "clash.yaml")
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		t.Skipf("默认模板不可用: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}

	output := strings.ReplaceAll(string(data), "\r\n", "\n")
	templateText := strings.ReplaceAll(string(templateBytes), "\r\n", "\n")

	if got, want := strings.Count(output, "\n\n"), strings.Count(templateText, "\n\n"); got != want {
		t.Fatalf("模板空行数量被改写: got %d want %d", got, want)
	}
	for _, line := range strings.Split(templateText, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.Contains(output, line) {
			t.Fatalf("模板注释丢失: %q", line)
		}
	}
	// 固定成员列表的代理组保持模板原文（含 emoji）
	assertContains(t, "手选组成员", output, "    proxies:\n      - ♻️ 自动选择\n      - 🇭🇰 香港节点\n")
	// 占位符就地展开为节点名
	assertContains(t, "占位符展开", output, "  - name: 🚀 手动切换\n    type: select\n    proxies:\n      - Node-A\n")
	// 模板 proxies 的键行与后面的空行保持原样
	assertContains(t, "proxies 段落", output, "proxies:\n  - name: Node-A\n")
}

// TestEncodeClashAppendsToZeroIndentProxies 验证零缩进的序列只追加到列表末尾，
// 不会把后面的 proxy-groups / rules 一起吞掉。
func TestEncodeClashAppendsToZeroIndentProxies(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "clash-zero-indent-template.yaml")
	template := `mixed-port: 7890
proxies:
- name: Template-Node
  type: ss
  server: 127.0.0.1
  port: 8388
proxy-groups:
- name: 节点选择
  type: select
  proxies: []
rules:
- MATCH,DIRECT
`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}
	output := string(data)

	assertContains(t, "节点追加到零缩进列表末尾", output, "  port: 8388\n- name: Node-A\n")
	assertContains(t, "零缩进组成员", output, "- name: 节点选择\n  type: select\n  proxies:\n    - Node-A\n")
	assertContains(t, "rules 段落原样", output, "rules:\n- MATCH,DIRECT\n")

	var parsed struct {
		Proxies []struct {
			Name string `yaml:"name"`
		} `yaml:"proxies"`
		ProxyGroups []struct {
			Name    string   `yaml:"name"`
			Proxies []string `yaml:"proxies"`
		} `yaml:"proxy-groups"`
		Rules []string `yaml:"rules"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("输出不是合法 YAML: %v\n%s", err, output)
	}
	if len(parsed.Proxies) != 2 || parsed.Proxies[1].Name != "Node-A" {
		t.Fatalf("节点未追加到模板 proxies: %#v", parsed.Proxies)
	}
	if len(parsed.ProxyGroups) != 1 || len(parsed.ProxyGroups[0].Proxies) != 1 ||
		parsed.ProxyGroups[0].Proxies[0] != "Node-A" {
		t.Fatalf("代理组列表错误: %#v", parsed.ProxyGroups)
	}
	if len(parsed.Rules) != 1 || parsed.Rules[0] != "MATCH,DIRECT" {
		t.Fatalf("rules 被改写: %#v", parsed.Rules)
	}
}

// TestEncodeClashFillsMissingGroupProxies 验证缺少 proxies 键的组只在组内补上列表。
func TestEncodeClashFillsMissingGroupProxies(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "clash-missing-group-proxies-template.yaml")
	template := "proxies: ~\nproxy-groups:\n- name: 手动切换\n  type: select\n- name: 备用\n  type: select\n  proxies: []\n" +
		"rules:\n- MATCH,DIRECT\n"
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}

	var parsed struct {
		ProxyGroups []struct {
			Name    string   `yaml:"name"`
			Proxies []string `yaml:"proxies"`
		} `yaml:"proxy-groups"`
		Rules []string `yaml:"rules"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("输出不是合法 YAML: %v\n%s", err, string(data))
	}
	if len(parsed.ProxyGroups) != 2 {
		t.Fatalf("代理组数量错误: %#v", parsed.ProxyGroups)
	}
	for index, group := range parsed.ProxyGroups {
		if len(group.Proxies) != 1 || group.Proxies[0] != "Node-A" {
			t.Fatalf("第 %d 组成员错误: %#v", index, group)
		}
	}
	if len(parsed.Rules) != 1 || parsed.Rules[0] != "MATCH,DIRECT" {
		t.Fatalf("rules 被改写: %#v", parsed.Rules)
	}
}

// TestEncodeClashReplacesNullOnNextLine 验证值另起一行的 ~ 会随键行一起被替换。
func TestEncodeClashReplacesNullOnNextLine(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "clash-null-next-line-template.yaml")
	template := "mixed-port: 7890\nproxies:\n  ~\nproxy-groups:\n  - name: 手动切换\n    type: select\n    proxies: []\n"
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}
	output := string(data)

	assertContains(t, "独立行 null 被替换", output, "proxies:\n  - name: Node-A\n")
	if strings.Contains(output, "~") {
		t.Fatalf("模板里的 null 残留:\n%s", output)
	}

	var parsed struct {
		Proxies []struct {
			Name string `yaml:"name"`
		} `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("输出不是合法 YAML: %v\n%s", err, output)
	}
	if len(parsed.Proxies) != 1 || parsed.Proxies[0].Name != "Node-A" {
		t.Fatalf("节点未写入模板: %#v", parsed.Proxies)
	}
}

// TestEncodeClashKeepsCommentBeforeNextSection 验证序列末尾的段落注释仍属于下面的键。
func TestEncodeClashKeepsCommentBeforeNextSection(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "clash-section-comment-template.yaml")
	template := "proxies:\n  - name: Template-Node\n    type: ss\n    server: 127.0.0.1\n    port: 8388\n" +
		"  # Ниже — прокси-группы\nproxy-groups:\n  - name: 手动切换\n    type: select\n    proxies: []\n"
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}
	output := string(data)

	assertContains(t, "段落注释仍贴在 proxy-groups 前", output, "  # Ниже — прокси-группы\nproxy-groups:\n")
	nodeIndex := strings.Index(output, "- name: Node-A")
	commentIndex := strings.Index(output, "# Ниже — прокси-группы")
	if nodeIndex < 0 || commentIndex < 0 || nodeIndex > commentIndex {
		t.Fatalf("节点被插到了段落注释之后:\n%s", output)
	}
}

// TestEncodeClashAliasTemplateFallsBack 验证别名模板走节点树重编码且结果合法。
func TestEncodeClashAliasTemplateFallsBack(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "clash-alias-template.yaml")
	template := "shared: &nodes\n  - name: Template-Node\n    type: ss\n    server: 127.0.0.1\n    port: 8388\n" +
		"proxies: *nodes\nproxy-groups:\n  - name: 手动切换\n    type: select\n    proxies: []\n"
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}

	var parsed struct {
		Proxies []struct {
			Name string `yaml:"name"`
		} `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("输出不是合法 YAML: %v\n%s", err, string(data))
	}
	if len(parsed.Proxies) != 2 || parsed.Proxies[1].Name != "Node-A" {
		t.Fatalf("节点未追加到别名列表: %#v", parsed.Proxies)
	}
}

// TestEncodeClashCROnlyTemplateFallsBack 验证只有 CR 换行的模板不会按错误行号改写。
func TestEncodeClashCROnlyTemplateFallsBack(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "clash-cr-template.yaml")
	template := strings.ReplaceAll(
		"proxies: ~\nproxy-groups:\n  - name: 手动切换\n    type: select\n    proxies: []\n", "\n", "\r")
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("写入模板失败: %v", err)
	}

	data, err := DecodeClash(
		[]Proxy{{Name: "Node-A", Type: "ss", Server: "example.com", Port: FlexPort(8388)}},
		templatePath,
	)
	if err != nil {
		t.Fatalf("DecodeClash 失败: %v", err)
	}

	var parsed struct {
		Proxies []struct {
			Name string `yaml:"name"`
		} `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("输出不是合法 YAML: %v\n%s", err, string(data))
	}
	if len(parsed.Proxies) != 1 || parsed.Proxies[0].Name != "Node-A" {
		t.Fatalf("节点未写入模板: %#v", parsed.Proxies)
	}
}
