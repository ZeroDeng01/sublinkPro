package protocol

import (
	"strings"
	"testing"
)

// TestVmessEncodeDecode 测试 VMess 编解码完整性
func TestVmessEncodeDecode(t *testing.T) {
	original := Vmess{
		Add:  "example.com",
		Port: "443",
		Id:   "12345678-1234-1234-1234-123456789abc",
		Aid:  "0",
		Net:  "ws",
		Type: "none",
		Host: "cdn.example.com",
		Path: "/vmess",
		Tls:  "tls",
		Sni:  "sni.example.com",
		Alpn: "h2,http/1.1",
		Fp:   "chrome",
		Scy:  "auto",
		Ps:   "测试节点-VMess",
		V:    "2",
	}

	// 编码
	encoded := EncodeVmessURL(original)
	if !strings.HasPrefix(encoded, "vmess://") {
		t.Errorf("编码后应以 vmess:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeVMESSURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Add", original.Add, decoded.Add)
	assertEqualString(t, "Id", original.Id, decoded.Id)
	assertEqualString(t, "Net", original.Net, decoded.Net)
	assertEqualString(t, "Path", original.Path, decoded.Path)
	assertEqualString(t, "Ps(名称)", original.Ps, decoded.Ps)
	assertEqualString(t, "Sni", original.Sni, decoded.Sni)
	assertEqualString(t, "Alpn", original.Alpn, decoded.Alpn)
	assertEqualString(t, "Fp", original.Fp, decoded.Fp)
	assertEqualString(t, "Scy", original.Scy, decoded.Scy)
	assertEqualString(t, "Tls", original.Tls, decoded.Tls)

	t.Logf("✓ VMess 编解码测试通过，名称: %s", decoded.Ps)
}

func TestVmessBuildProxyPreservesTLSMetadata(t *testing.T) {
	vmess := Vmess{
		Add:  "example.com",
		Port: "443",
		Id:   "12345678-1234-1234-1234-123456789abc",
		Net:  "ws",
		Path: "/vmess",
		Host: "cdn.example.com",
		Tls:  "tls",
		Sni:  "sni.example.com",
		Alpn: "h2,http/1.1",
		Fp:   "chrome",
		Scy:  "auto",
		Ps:   "测试节点-VMess-TLS",
		V:    "2",
	}

	proxy, err := buildVMessProxy(Urls{Url: EncodeVmessURL(vmess)}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildVMessProxy 失败: %v", err)
	}

	assertEqualString(t, "ClientFingerprint", vmess.Fp, proxy.Client_fingerprint)
	assertEqualString(t, "Servername", vmess.Sni, proxy.Servername)
	assertEqualString(t, "ALPN0", "h2", proxy.Alpn[0])
	assertEqualString(t, "ALPN1", "http/1.1", proxy.Alpn[1])
}

// TestVmessNameModification 测试 VMess 名称修改
func TestVmessNameModification(t *testing.T) {
	original := Vmess{
		Add:  "example.com",
		Port: "443",
		Id:   "12345678-1234-1234-1234-123456789abc",
		Net:  "tcp",
		Ps:   "原始名称",
		V:    "2",
	}

	newName := "新名称-VMess-测试"
	encoded := EncodeVmessURL(original)
	decoded, _ := DecodeVMESSURL(encoded)
	decoded.Ps = newName
	reEncoded := EncodeVmessURL(decoded)
	final, _ := DecodeVMESSURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Ps)
	assertEqualString(t, "服务器(不变)", original.Add, final.Add)
	assertEqualString(t, "UUID(不变)", original.Id, final.Id)

	t.Logf("✓ VMess 名称修改测试通过: %s -> %s", original.Ps, final.Ps)
}

// TestVmessSpecialCharacters 测试 VMess 特殊字符
func TestVmessSpecialCharacters(t *testing.T) {
	specialNames := []string{
		"节点 with spaces",
		"节点-with-dashes",
		"节点_with_underscores",
		"节点中文测试",
		"Node🚀Emoji",
		"Node (parentheses)",
	}

	for _, name := range specialNames {
		t.Run(name, func(t *testing.T) {
			original := Vmess{
				Add:  "example.com",
				Port: "443",
				Id:   "12345678-1234-1234-1234-123456789abc",
				Net:  "tcp",
				Ps:   name,
				V:    "2",
			}

			encoded := EncodeVmessURL(original)
			decoded, err := DecodeVMESSURL(encoded)
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}

			assertEqualString(t, "特殊字符名称", name, decoded.Ps)
			t.Logf("✓ 特殊字符测试通过: %s", name)
		})
	}
}
