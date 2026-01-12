package protocol

import (
	"strings"
	"testing"
)

// TestVlessEncodeDecode 测试 VLESS 编解码完整性
func TestVlessEncodeDecode(t *testing.T) {
	original := VLESS{
		Name:   "测试节点-VLESS",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security:   "tls",
			Encryption: "none",
			Type:       "ws",
			Host:       "cdn.example.com",
			Path:       "/vless",
			Sni:        "sni.example.com",
			Fp:         "chrome",
			Alpn:       []string{"h2", "http/1.1"},
		},
	}

	// 编码
	encoded := EncodeVLESSURL(original)
	if !strings.HasPrefix(encoded, "vless://") {
		t.Errorf("编码后应以 vless:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeVLESSURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Server", original.Server, decoded.Server)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Uuid", original.Uuid, decoded.Uuid)
	assertEqualString(t, "Name", original.Name, decoded.Name)
	assertEqualString(t, "Query.Type", original.Query.Type, decoded.Query.Type)
	assertEqualString(t, "Query.Sni", original.Query.Sni, decoded.Query.Sni)
	assertEqualString(t, "Query.Path", original.Query.Path, decoded.Query.Path)

	t.Logf("✓ VLESS 编解码测试通过，名称: %s", decoded.Name)
}

// TestVlessNameModification 测试 VLESS 名称修改
func TestVlessNameModification(t *testing.T) {
	original := VLESS{
		Name:   "原始名称",
		Uuid:   "12345678-1234-1234-1234-123456789abc",
		Server: "example.com",
		Port:   443,
		Query: VLESSQuery{
			Security: "tls",
			Type:     "tcp",
		},
	}

	newName := "新名称-VLESS-测试"
	encoded := EncodeVLESSURL(original)
	decoded, _ := DecodeVLESSURL(encoded)
	decoded.Name = newName
	reEncoded := EncodeVLESSURL(decoded)
	final, _ := DecodeVLESSURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Server, final.Server)
	assertEqualString(t, "UUID(不变)", original.Uuid, final.Uuid)
	assertEqualIntInterface(t, "端口(不变)", original.Port, final.Port)

	t.Logf("✓ VLESS 名称修改测试通过: %s -> %s", original.Name, final.Name)
}

// TestVlessSpecialCharacters 测试 VLESS 特殊字符
func TestVlessSpecialCharacters(t *testing.T) {
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
			original := VLESS{
				Name:   name,
				Uuid:   "12345678-1234-1234-1234-123456789abc",
				Server: "example.com",
				Port:   443,
				Query: VLESSQuery{
					Security: "tls",
					Type:     "tcp",
				},
			}

			encoded := EncodeVLESSURL(original)
			decoded, err := DecodeVLESSURL(encoded)
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}

			assertEqualString(t, "特殊字符名称", name, decoded.Name)
			t.Logf("✓ 特殊字符测试通过: %s", name)
		})
	}
}

// TestVlessV2rayFormat 测试 v2ray 格式 VLESS 链接解析（明文URL，非base64）
func TestVlessV2rayFormat(t *testing.T) {
	// 典型的v2ray格式VLESS链接
	testCases := []struct {
		name     string
		url      string
		expected VLESSQuery
	}{
		{
			name: "WebSocket传输层",
			url:  "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=ws&host=cdn.example.com&path=%2Fvless&sni=example.com&fp=chrome#测试节点",
			expected: VLESSQuery{
				Security:   "tls",
				Encryption: "none",
				Type:       "ws",
				Host:       "cdn.example.com",
				Path:       "/vless",
				Sni:        "example.com",
				Fp:         "chrome",
			},
		},
		{
			name: "Reality配置",
			url:  "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=reality&type=tcp&flow=xtls-rprx-vision&pbk=testpublickey&sid=testshortid&sni=example.com&fp=chrome#Reality节点",
			expected: VLESSQuery{
				Security: "reality",
				Type:     "tcp",
				Flow:     "xtls-rprx-vision",
				Pbk:      "testpublickey",
				Sid:      "testshortid",
				Sni:      "example.com",
				Fp:       "chrome",
			},
		},
		{
			name: "gRPC传输层",
			url:  "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=grpc&serviceName=mygrpc&mode=gun#gRPC节点",
			expected: VLESSQuery{
				Security:    "tls",
				Type:        "grpc",
				ServiceName: "mygrpc",
				Mode:        "gun",
			},
		},
		{
			name: "H2传输层",
			url:  "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=h2&host=example.com&path=%2Fh2path#H2节点",
			expected: VLESSQuery{
				Security: "tls",
				Type:     "h2",
				Host:     "example.com",
				Path:     "/h2path",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoded, err := DecodeVLESSURL(tc.url)
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}

			assertEqualString(t, "Security", tc.expected.Security, decoded.Query.Security)
			assertEqualString(t, "Type", tc.expected.Type, decoded.Query.Type)
			if tc.expected.Host != "" {
				assertEqualString(t, "Host", tc.expected.Host, decoded.Query.Host)
			}
			if tc.expected.Path != "" {
				assertEqualString(t, "Path", tc.expected.Path, decoded.Query.Path)
			}
			if tc.expected.Flow != "" {
				assertEqualString(t, "Flow", tc.expected.Flow, decoded.Query.Flow)
			}
			if tc.expected.Pbk != "" {
				assertEqualString(t, "Pbk", tc.expected.Pbk, decoded.Query.Pbk)
			}
			if tc.expected.ServiceName != "" {
				assertEqualString(t, "ServiceName", tc.expected.ServiceName, decoded.Query.ServiceName)
			}

			t.Logf("✓ %s 测试通过", tc.name)
		})
	}
}

// TestVlessPacketEncoding 测试 packet-encoding 参数
func TestVlessPacketEncoding(t *testing.T) {
	url := "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=tcp&packetEncoding=xudp#xudp节点"
	decoded, err := DecodeVLESSURL(url)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "PacketEncoding", "xudp", decoded.Query.PacketEncoding)
	t.Logf("✓ packet-encoding 测试通过")
}
