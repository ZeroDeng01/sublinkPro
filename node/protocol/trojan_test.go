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
			Security: "tls",
			Type:     "ws",
			Host:     "cdn.example.com",
			Path:     "/trojan",
			Sni:      "sni.example.com",
			Fp:       "chrome",
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

	t.Logf("✓ Trojan 编解码测试通过，名称: %s", decoded.Name)
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
