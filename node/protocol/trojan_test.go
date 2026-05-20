package protocol

import (
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
