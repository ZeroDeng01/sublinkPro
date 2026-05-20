package protocol

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestAnyTLSEncodeDecode 测试 AnyTLS 编解码完整性
func TestAnyTLSEncodeDecode(t *testing.T) {
	original := AnyTLS{
		Name:                     "测试节点-AnyTLS",
		Server:                   "example.com",
		Port:                     443,
		Password:                 "test-anytls-password",
		SkipCertVerify:           true,
		SNI:                      "sni.example.com",
		ALPN:                     []string{"h2", "http/1.1"},
		ClientFingerprint:        "chrome",
		Fingerprint:              "16dac3717024eb319093d1c95290c14adc850e2814b2208d11c7b7a436923859",
		UDP:                      true,
		IdleSessionCheckInterval: 30,
		IdleSessionTimeout:       45,
		MinIdleSession:           2,
	}

	// 编码
	encoded := EncodeAnyTLSURL(original)
	if !strings.HasPrefix(encoded, "anytls://") {
		t.Errorf("编码后应以 anytls:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeAnyTLSURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Server", original.Server, decoded.Server)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Password", original.Password, decoded.Password)
	assertEqualString(t, "SNI", original.SNI, decoded.SNI)
	assertEqualString(t, "ALPN0", original.ALPN[0], decoded.ALPN[0])
	assertEqualString(t, "ALPN1", original.ALPN[1], decoded.ALPN[1])
	assertEqualString(t, "Name", original.Name, decoded.Name)
	assertEqualBool(t, "SkipCertVerify", original.SkipCertVerify, decoded.SkipCertVerify)
	assertEqualString(t, "ClientFingerprint", original.ClientFingerprint, decoded.ClientFingerprint)
	assertEqualString(t, "Fingerprint", original.Fingerprint, decoded.Fingerprint)
	assertEqualBool(t, "UDP", original.UDP, decoded.UDP)
	assertEqualInt(t, "IdleSessionCheckInterval", original.IdleSessionCheckInterval, decoded.IdleSessionCheckInterval)
	assertEqualInt(t, "IdleSessionTimeout", original.IdleSessionTimeout, decoded.IdleSessionTimeout)
	assertEqualInt(t, "MinIdleSession", original.MinIdleSession, decoded.MinIdleSession)

	proxy, err := buildAnyTLSProxy(Urls{Url: encoded}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildAnyTLSProxy 失败: %v", err)
	}
	assertEqualString(t, "ProxyFingerprint", original.Fingerprint, proxy.Fingerprint)
	assertEqualString(t, "ProxyALPN0", original.ALPN[0], proxy.Alpn[0])
	assertEqualBool(t, "ProxyUDP", original.UDP, proxy.Udp)
	assertEqualInt(t, "ProxyIdleSessionCheckInterval", original.IdleSessionCheckInterval, proxy.AnyTLSIdleCheck)
	assertEqualInt(t, "ProxyIdleSessionTimeout", original.IdleSessionTimeout, proxy.AnyTLSIdleTimeout)
	assertEqualInt(t, "ProxyMinIdleSession", original.MinIdleSession, proxy.AnyTLSMinIdle)

	t.Logf("✓ AnyTLS 编解码测试通过，名称: %s", decoded.Name)
}

func TestConvertProxyToAnyTLSPreservesMihomoFields(t *testing.T) {
	proxy := Proxy{
		Name:               "AnyTLS-Clash",
		Type:               "anytls",
		Server:             "example.com",
		Port:               FlexPort(443),
		Password:           "password",
		Sni:                "sni.example.com",
		Alpn:               []string{"h2", "http/1.1"},
		Client_fingerprint: "chrome",
		Fingerprint:        "16dac3717024eb319093d1c95290c14adc850e2814b2208d11c7b7a436923859",
		Udp:                true,
		Skip_cert_verify:   true,
		AnyTLSIdleCheck:    30,
		AnyTLSIdleTimeout:  45,
		AnyTLSMinIdle:      2,
	}

	anyTLS := ConvertProxyToAnyTLS(proxy)
	assertEqualString(t, "ALPN0", proxy.Alpn[0], anyTLS.ALPN[0])
	assertEqualBool(t, "UDP", proxy.Udp, anyTLS.UDP)
	assertEqualInt(t, "IdleSessionCheckInterval", proxy.AnyTLSIdleCheck, anyTLS.IdleSessionCheckInterval)
	assertEqualInt(t, "IdleSessionTimeout", proxy.AnyTLSIdleTimeout, anyTLS.IdleSessionTimeout)
	assertEqualInt(t, "MinIdleSession", proxy.AnyTLSMinIdle, anyTLS.MinIdleSession)

	encoded := EncodeAnyTLSURL(anyTLS)
	assertContains(t, "EncodedALPN", encoded, "alpn=h2%2Chttp%2F1.1")
	assertContains(t, "EncodedUDP", encoded, "udp=1")
	assertContains(t, "EncodedIdleCheck", encoded, "idle-session-check-interval=30")
}

func TestAnyTLSProxyYAMLUsesMihomoFieldNames(t *testing.T) {
	proxy := Proxy{
		Name:              "AnyTLS-YAML",
		Type:              "anytls",
		Server:            "example.com",
		Port:              FlexPort(443),
		Password:          "password",
		Alpn:              []string{"h2", "http/1.1"},
		Udp:               true,
		AnyTLSIdleCheck:   30,
		AnyTLSIdleTimeout: 45,
		AnyTLSMinIdle:     2,
	}

	data, err := yaml.Marshal(proxy)
	if err != nil {
		t.Fatalf("序列化 AnyTLS Proxy 失败: %v", err)
	}
	yamlText := string(data)
	assertContains(t, "YAMLALPN", yamlText, "alpn:")
	assertContains(t, "YAMLUDP", yamlText, "udp: true")
	assertContains(t, "YAMLIdleCheck", yamlText, "idle-session-check-interval: 30")
	assertContains(t, "YAMLIdleTimeout", yamlText, "idle-session-timeout: 45")
	assertContains(t, "YAMLMinIdle", yamlText, "min-idle-session: 2")
}

func TestAnyTLSOutputConfigForcesUDPAndSkipCert(t *testing.T) {
	link := EncodeAnyTLSURL(AnyTLS{
		Name:     "AnyTLS-Config",
		Server:   "example.com",
		Port:     443,
		Password: "password",
	})

	proxy, err := buildAnyTLSProxy(Urls{Url: link}, OutputConfig{Udp: true, Cert: true})
	if err != nil {
		t.Fatalf("buildAnyTLSProxy 失败: %v", err)
	}
	assertEqualBool(t, "ForcedUDP", true, proxy.Udp)
	assertEqualBool(t, "ForcedSkipCert", true, proxy.Skip_cert_verify)
}

func TestAnyTLSOwnUDPAndSkipCertPreservedWithoutOutputConfig(t *testing.T) {
	link := EncodeAnyTLSURL(AnyTLS{
		Name:           "AnyTLS-Self",
		Server:         "example.com",
		Port:           443,
		Password:       "password",
		UDP:            true,
		SkipCertVerify: true,
	})

	proxy, err := buildAnyTLSProxy(Urls{Url: link}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildAnyTLSProxy 失败: %v", err)
	}
	assertEqualBool(t, "OwnUDP", true, proxy.Udp)
	assertEqualBool(t, "OwnSkipCert", true, proxy.Skip_cert_verify)
}

// TestAnyTLSNameModification 测试 AnyTLS 名称修改
func TestAnyTLSNameModification(t *testing.T) {
	original := AnyTLS{
		Name:     "原始名称",
		Server:   "example.com",
		Port:     443,
		Password: "test-password",
		SNI:      "example.com",
	}

	newName := "新名称-AnyTLS-测试"
	encoded := EncodeAnyTLSURL(original)
	decoded, _ := DecodeAnyTLSURL(encoded)
	decoded.Name = newName
	reEncoded := EncodeAnyTLSURL(decoded)
	final, _ := DecodeAnyTLSURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Server, final.Server)
	assertEqualString(t, "密码(不变)", original.Password, final.Password)
	assertEqualString(t, "SNI(不变)", original.SNI, final.SNI)

	t.Logf("✓ AnyTLS 名称修改测试通过: %s -> %s", original.Name, final.Name)
}

// TestAnyTLSWithoutOptionalFields 测试无可选字段的 AnyTLS
func TestAnyTLSWithoutOptionalFields(t *testing.T) {
	original := AnyTLS{
		Name:     "测试节点-最小配置",
		Server:   "example.com",
		Port:     443,
		Password: "password",
	}

	encoded := EncodeAnyTLSURL(original)
	decoded, err := DecodeAnyTLSURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Server", original.Server, decoded.Server)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Password", original.Password, decoded.Password)
	assertEqualString(t, "Name", original.Name, decoded.Name)

	t.Logf("✓ AnyTLS 最小配置测试通过，名称: %s", decoded.Name)
}
