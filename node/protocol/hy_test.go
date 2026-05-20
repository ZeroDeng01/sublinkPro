package protocol

import (
	"strings"
	"testing"
)

// TestHYEncodeDecode 测试 Hysteria 编解码完整性
func TestHYEncodeDecode(t *testing.T) {
	original := HY{
		Name:     "测试节点-Hysteria",
		Host:     "example.com",
		Port:     443,
		Auth:     "test-auth-string",
		Peer:     "sni.example.com",
		Protocol: "udp",
		Insecure: 1,
		UpMbps:   100,
		DownMbps: 100,
	}

	// 编码
	encoded := EncodeHYURL(original)
	if !strings.HasPrefix(encoded, "hysteria://") && !strings.HasPrefix(encoded, "hy://") {
		t.Errorf("编码后应以 hysteria:// 或 hy:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeHYURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Host", original.Host, decoded.Host)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Peer", original.Peer, decoded.Peer)
	assertEqualString(t, "Auth", original.Auth, decoded.Auth)
	assertEqualString(t, "Protocol", original.Protocol, decoded.Protocol)
	assertEqualString(t, "Name", original.Name, decoded.Name)

	t.Logf("✓ Hysteria 编解码测试通过，名称: %s", decoded.Name)
}

// TestHYNameModification 测试 Hysteria 名称修改
func TestHYNameModification(t *testing.T) {
	original := HY{
		Name: "原始名称",
		Host: "example.com",
		Port: 443,
		Auth: "test-auth",
	}

	newName := "新名称-Hysteria-测试"
	encoded := EncodeHYURL(original)
	decoded, _ := DecodeHYURL(encoded)
	decoded.Name = newName
	reEncoded := EncodeHYURL(decoded)
	final, _ := DecodeHYURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Host, final.Host)

	t.Logf("✓ Hysteria 名称修改测试通过: %s -> %s", original.Name, final.Name)
}

// TestHY2EncodeDecode 测试 Hysteria2 编解码完整性
func TestHY2EncodeDecode(t *testing.T) {
	original := HY2{
		Name:              "测试节点-Hysteria2",
		Host:              "example.com",
		Port:              443,
		Password:          "test-hy2-password",
		Sni:               "sni.example.com",
		Insecure:          1,
		Obfs:              "salamander",
		ClientFingerprint: "chrome",
		Fingerprint:       "16dac3717024eb319093d1c95290c14adc850e2814b2208d11c7b7a436923859",
	}

	// 编码
	encoded := EncodeHY2URL(original)
	if !strings.HasPrefix(encoded, "hysteria2://") && !strings.HasPrefix(encoded, "hy2://") {
		t.Errorf("编码后应以 hysteria2:// 或 hy2:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeHY2URL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Host", original.Host, decoded.Host)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Password", original.Password, decoded.Password)
	assertEqualString(t, "ClientFingerprint", original.ClientFingerprint, decoded.ClientFingerprint)
	assertEqualString(t, "Fingerprint", original.Fingerprint, decoded.Fingerprint)
	assertEqualString(t, "Name", original.Name, decoded.Name)

	proxy, err := buildHY2Proxy(Urls{Url: encoded}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildHY2Proxy 失败: %v", err)
	}
	assertEqualString(t, "ProxyClientFingerprint", original.ClientFingerprint, proxy.Client_fingerprint)
	assertEqualString(t, "ProxyFingerprint", original.Fingerprint, proxy.Fingerprint)

	t.Logf("✓ Hysteria2 编解码测试通过，名称: %s", decoded.Name)
}

func TestHY2RejectsInvalidCertificateFingerprint(t *testing.T) {
	link := "hy2://password@example.com:443/?sni=example.com&pinSHA256=abc%2Cskip-cert-verify%3Dtrue#bad-fingerprint"

	decoded, err := DecodeHY2URL(link)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	assertEqualString(t, "InvalidFingerprintDropped", "", decoded.Fingerprint)

	proxy, err := buildHY2Proxy(Urls{Url: link}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildHY2Proxy 失败: %v", err)
	}
	assertEqualString(t, "ProxyInvalidFingerprintDropped", "", proxy.Fingerprint)
}

// TestHY2NameModification 测试 Hysteria2 名称修改
func TestHY2NameModification(t *testing.T) {
	original := HY2{
		Name:     "原始名称",
		Host:     "example.com",
		Port:     443,
		Password: "test-password",
	}

	newName := "新名称-Hysteria2-测试"
	encoded := EncodeHY2URL(original)
	decoded, _ := DecodeHY2URL(encoded)
	decoded.Name = newName
	reEncoded := EncodeHY2URL(decoded)
	final, _ := DecodeHY2URL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Host, final.Host)
	assertEqualString(t, "密码(不变)", original.Password, final.Password)

	t.Logf("✓ Hysteria2 名称修改测试通过: %s -> %s", original.Name, final.Name)
}

// TestHY2IPv6RawUpdatePreservesBracketedAuthority 覆盖原始信息编辑器保存 IPv6 HY2 节点的回归场景。
// DecodeHY2URL 会把 Host 规范化为不带方括号的 IPv6；再次编码时必须恢复 URL authority 所需的方括号。
func TestHY2IPv6RawUpdatePreservesBracketedAuthority(t *testing.T) {
	original := "hy2://test-password@[2001:db8::3]:22000?insecure=1&sni=example.com#ipv6-hy2-test"

	updated, err := UpdateNodeLinkFields(original, `{"Host":"2001:db8::3","Sni":"example.com"}`)
	if err != nil {
		t.Fatalf("更新 HY2 IPv6 原始字段失败: %v", err)
	}
	if !strings.Contains(updated, "@[2001:db8::3]:22000") {
		t.Fatalf("IPv6 authority 应保留方括号，实际链接: %s", updated)
	}

	decoded, err := DecodeHY2URL(updated)
	if err != nil {
		t.Fatalf("回写后的 HY2 IPv6 链接应可解析: %v", err)
	}
	assertEqualString(t, "Host", "2001:db8::3", decoded.Host)
	assertEqualIntInterface(t, "Port", 22000, decoded.Port)

	identity, err := ExtractLinkIdentity(updated)
	if err != nil {
		t.Fatalf("提取 HY2 IPv6 身份信息失败: %v", err)
	}
	assertEqualString(t, "Address", "[2001:db8::3]:22000", identity.Address)
}
