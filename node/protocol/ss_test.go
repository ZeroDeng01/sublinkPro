package protocol

import (
	"strings"
	"testing"
)

// TestSSEncodeDecode 测试 Shadowsocks 编解码完整性
func TestSSEncodeDecode(t *testing.T) {
	original := Ss{
		Name:   "测试节点-SS",
		Server: "example.com",
		Port:   8388,
		Param: Param{
			Cipher:   "aes-256-gcm",
			Password: "test-ss-password",
		},
	}

	// 编码
	encoded := EncodeSSURL(original)
	if !strings.HasPrefix(encoded, "ss://") {
		t.Errorf("编码后应以 ss:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeSSURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Server", original.Server, decoded.Server)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Cipher", original.Param.Cipher, decoded.Param.Cipher)
	assertEqualString(t, "Password", original.Param.Password, decoded.Param.Password)
	assertEqualString(t, "Name", original.Name, decoded.Name)

	t.Logf("✓ SS 编解码测试通过，名称: %s", decoded.Name)
}

// TestSSNameModification 测试 SS 名称修改
func TestSSNameModification(t *testing.T) {
	original := Ss{
		Name:   "原始名称",
		Server: "example.com",
		Port:   8388,
		Param: Param{
			Cipher:   "aes-256-gcm",
			Password: "test-password",
		},
	}

	newName := "新名称-SS-测试"
	encoded := EncodeSSURL(original)
	decoded, _ := DecodeSSURL(encoded)
	decoded.Name = newName
	reEncoded := EncodeSSURL(decoded)
	final, _ := DecodeSSURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Server, final.Server)
	assertEqualString(t, "密码(不变)", original.Param.Password, final.Param.Password)
	assertEqualString(t, "加密方式(不变)", original.Param.Cipher, final.Param.Cipher)

	t.Logf("✓ SS 名称修改测试通过: %s -> %s", original.Name, final.Name)
}

// TestSsrEncodeDecode 测试 ShadowsocksR 编解码完整性
func TestSsrEncodeDecode(t *testing.T) {
	original := Ssr{
		Server:   "example.com",
		Port:     8388,
		Method:   "aes-256-cfb",
		Password: "test-ssr-password",
		Protocol: "origin",
		Obfs:     "plain",
		Qurey: Ssrquery{
			Remarks:   "测试节点-SSR",
			Obfsparam: "",
		},
	}

	// 编码
	encoded := EncodeSSRURL(original)
	if !strings.HasPrefix(encoded, "ssr://") {
		t.Errorf("编码后应以 ssr:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeSSRURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Server", original.Server, decoded.Server)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Method", original.Method, decoded.Method)
	assertEqualString(t, "Remarks(名称)", original.Qurey.Remarks, decoded.Qurey.Remarks)
	assertEqualString(t, "Protocol", original.Protocol, decoded.Protocol)
	assertEqualString(t, "Obfs", original.Obfs, decoded.Obfs)

	t.Logf("✓ SSR 编解码测试通过，名称: %s", decoded.Qurey.Remarks)
}

// TestSsrNameModification 测试 SSR 名称修改
func TestSsrNameModification(t *testing.T) {
	original := Ssr{
		Server:   "example.com",
		Port:     8388,
		Method:   "aes-256-cfb",
		Password: "test-password",
		Protocol: "origin",
		Obfs:     "plain",
		Qurey: Ssrquery{
			Remarks: "原始名称",
		},
	}

	newName := "新名称-SSR-测试"
	encoded := EncodeSSRURL(original)
	decoded, _ := DecodeSSRURL(encoded)
	decoded.Qurey.Remarks = newName
	reEncoded := EncodeSSRURL(decoded)
	final, _ := DecodeSSRURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Qurey.Remarks)
	assertEqualString(t, "服务器(不变)", original.Server, final.Server)

	t.Logf("✓ SSR 名称修改测试通过: %s -> %s", original.Qurey.Remarks, final.Qurey.Remarks)
}
