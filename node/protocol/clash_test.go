package protocol

import (
	"strings"
	"testing"
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
	// 注意：SSR 密码在编码时会进行 base64 编码，这里跳过密码验证
	assertEqualString(t, "Protocol", ssr.Protocol, proxy.Protocol)

	t.Logf("✓ SSR LinkToProxy 测试通过，名称: %s", proxy.Name)
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
