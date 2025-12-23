package protocol

import (
	"strings"
	"testing"
)

// TestSocks5EncodeDecode 测试 Socks5 编解码完整性
func TestSocks5EncodeDecode(t *testing.T) {
	original := Socks5{
		Name:     "测试节点-Socks5",
		Server:   "example.com",
		Port:     1080,
		Username: "testuser",
		Password: "testpass",
	}

	// 编码
	encoded := EncodeSocks5URL(original)
	if !strings.HasPrefix(encoded, "socks5://") {
		t.Errorf("编码后应以 socks5:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeSocks5URL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Server", original.Server, decoded.Server)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Username", original.Username, decoded.Username)
	assertEqualString(t, "Password", original.Password, decoded.Password)
	assertEqualString(t, "Name", original.Name, decoded.Name)

	t.Logf("✓ Socks5 编解码测试通过，名称: %s", decoded.Name)
}

// TestSocks5NameModification 测试 Socks5 名称修改
func TestSocks5NameModification(t *testing.T) {
	original := Socks5{
		Name:     "原始名称",
		Server:   "example.com",
		Port:     1080,
		Username: "user",
		Password: "pass",
	}

	newName := "新名称-Socks5-测试"
	encoded := EncodeSocks5URL(original)
	decoded, _ := DecodeSocks5URL(encoded)
	decoded.Name = newName
	reEncoded := EncodeSocks5URL(decoded)
	final, _ := DecodeSocks5URL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Server, final.Server)
	assertEqualString(t, "用户名(不变)", original.Username, final.Username)
	assertEqualString(t, "密码(不变)", original.Password, final.Password)

	t.Logf("✓ Socks5 名称修改测试通过: %s -> %s", original.Name, final.Name)
}

// TestSocks5WithoutAuth 测试无认证的 Socks5
func TestSocks5WithoutAuth(t *testing.T) {
	original := Socks5{
		Name:   "测试节点-无认证",
		Server: "example.com",
		Port:   1080,
	}

	encoded := EncodeSocks5URL(original)
	decoded, err := DecodeSocks5URL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Server", original.Server, decoded.Server)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Name", original.Name, decoded.Name)

	t.Logf("✓ Socks5 无认证测试通过，名称: %s", decoded.Name)
}
