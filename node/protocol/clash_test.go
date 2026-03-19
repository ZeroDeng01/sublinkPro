package protocol

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

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

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("解析输出 YAML 失败: %v", err)
	}

	rawGroups, ok := config["proxy-groups"].([]interface{})
	if !ok || len(rawGroups) != 2 {
		t.Fatalf("proxy-groups 解析失败: %#v", config["proxy-groups"])
	}

	firstGroup, ok := rawGroups[0].(map[string]interface{})
	if !ok {
		t.Fatalf("第一个代理组类型错误: %#v", rawGroups[0])
	}
	if firstGroup["name"] != "🛠️ 手选单一节点" {
		t.Fatalf("第一个代理组名称错误: %#v", firstGroup["name"])
	}
	if _, exists := firstGroup["proxies"]; exists {
		t.Fatalf("use 组不应被强行注入 proxies: %#v", firstGroup)
	}
	if use, ok := firstGroup["use"].([]interface{}); !ok || len(use) != 2 {
		t.Fatalf("use 组 providers 丢失: %#v", firstGroup["use"])
	}

	secondGroup, ok := rawGroups[1].(map[string]interface{})
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

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("解析输出 YAML 失败: %v", err)
	}

	rawGroups, ok := config["proxy-groups"].([]interface{})
	if !ok || len(rawGroups) != 3 {
		t.Fatalf("proxy-groups 解析失败: %#v", config["proxy-groups"])
	}

	firstGroup, ok := rawGroups[0].(map[string]interface{})
	if !ok {
		t.Fatalf("第一个代理组类型错误: %#v", rawGroups[0])
	}
	if _, exists := firstGroup["proxies"]; exists {
		t.Fatalf("include-all-proxies 组不应被强行注入 proxies: %#v", firstGroup)
	}
	if includeAllProxies, ok := firstGroup["include-all-proxies"].(bool); !ok || !includeAllProxies {
		t.Fatalf("include-all-proxies 丢失: %#v", firstGroup["include-all-proxies"])
	}

	secondGroup, ok := rawGroups[1].(map[string]interface{})
	if !ok {
		t.Fatalf("第二个代理组类型错误: %#v", rawGroups[1])
	}
	if _, exists := secondGroup["proxies"]; exists {
		t.Fatalf("动态负载均衡组不应被强行注入 proxies: %#v", secondGroup)
	}
	if includeAllProxies, ok := secondGroup["include-all-proxies"].(bool); !ok || !includeAllProxies {
		t.Fatalf("load-balance 组 include-all-proxies 丢失: %#v", secondGroup["include-all-proxies"])
	}

	thirdGroup, ok := rawGroups[2].(map[string]interface{})
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
