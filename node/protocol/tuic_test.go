package protocol

import (
	"strings"
	"testing"
)

// TestTuicEncodeDecode 测试 TUIC 编解码完整性
func TestTuicEncodeDecode(t *testing.T) {
	original := Tuic{
		Name:               "测试节点-TUIC",
		Host:               "example.com",
		Port:               443,
		Uuid:               "12345678-1234-1234-1234-123456789abc",
		Password:           "test-tuic-password",
		Congestion_control: "bbr",
		Alpn:               []string{"h3"},
		Sni:                "sni.example.com",
		Udp_relay_mode:     "native",
		Disable_sni:        0,
	}

	// 编码
	encoded := EncodeTuicURL(original)
	if !strings.HasPrefix(encoded, "tuic://") {
		t.Errorf("编码后应以 tuic:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeTuicURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Host", original.Host, decoded.Host)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Uuid", original.Uuid, decoded.Uuid)
	assertEqualString(t, "Password", original.Password, decoded.Password)
	assertEqualString(t, "Name", original.Name, decoded.Name)
	assertEqualString(t, "Sni", original.Sni, decoded.Sni)

	t.Logf("✓ TUIC 编解码测试通过，名称: %s", decoded.Name)
}

// TestTuicNameModification 测试 TUIC 名称修改
func TestTuicNameModification(t *testing.T) {
	original := Tuic{
		Name:     "原始名称",
		Host:     "example.com",
		Port:     443,
		Uuid:     "12345678-1234-1234-1234-123456789abc",
		Password: "test-password",
	}

	newName := "新名称-TUIC-测试"
	encoded := EncodeTuicURL(original)
	decoded, _ := DecodeTuicURL(encoded)
	decoded.Name = newName
	reEncoded := EncodeTuicURL(decoded)
	final, _ := DecodeTuicURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Host, final.Host)
	assertEqualString(t, "UUID(不变)", original.Uuid, final.Uuid)
	assertEqualString(t, "密码(不变)", original.Password, final.Password)

	t.Logf("✓ TUIC 名称修改测试通过: %s -> %s", original.Name, final.Name)
}
