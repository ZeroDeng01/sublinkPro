package protocol

import (
	"strings"
	"testing"
)

// TestAnyTLSEncodeDecode 测试 AnyTLS 编解码完整性
func TestAnyTLSEncodeDecode(t *testing.T) {
	original := AnyTLS{
		Name:              "测试节点-AnyTLS",
		Server:            "example.com",
		Port:              443,
		Password:          "test-anytls-password",
		SkipCertVerify:    true,
		SNI:               "sni.example.com",
		ClientFingerprint: "chrome",
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
	assertEqualString(t, "Name", original.Name, decoded.Name)
	assertEqualBool(t, "SkipCertVerify", original.SkipCertVerify, decoded.SkipCertVerify)
	assertEqualString(t, "ClientFingerprint", original.ClientFingerprint, decoded.ClientFingerprint)

	t.Logf("✓ AnyTLS 编解码测试通过，名称: %s", decoded.Name)
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
