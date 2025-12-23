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
		Name:     "测试节点-Hysteria2",
		Host:     "example.com",
		Port:     443,
		Password: "test-hy2-password",
		Sni:      "sni.example.com",
		Insecure: 1,
		Obfs:     "salamander",
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
	assertEqualString(t, "Name", original.Name, decoded.Name)

	t.Logf("✓ Hysteria2 编解码测试通过，名称: %s", decoded.Name)
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
