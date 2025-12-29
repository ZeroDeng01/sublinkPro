package protocol

import (
	"strings"
	"testing"
)

// TestWireGuardEncodeDecode 测试 WireGuard URL 编解码完整性
func TestWireGuardEncodeDecode(t *testing.T) {
	original := WireGuard{
		Name:       "测试节点-WireGuard",
		Server:     "162.159.192.127",
		Port:       7152,
		PrivateKey: "OOrigZsSjw2YaY4urjbbU4/BNOZKXqW6EYNm8XKLtkU=",
		PublicKey:  "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=",
		IP:         "172.16.0.2",
		IPv6:       "2606:4700:110:82ce:bdeb:e72d:572a:e280",
		MTU:        1280,
	}

	// 编码
	encoded := EncodeWireGuardURL(original)
	if !strings.HasPrefix(encoded, "wireguard://") {
		t.Errorf("编码后应以 wireguard:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeWireGuardURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Server", original.Server, decoded.Server)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "PrivateKey", original.PrivateKey, decoded.PrivateKey)
	assertEqualString(t, "PublicKey", original.PublicKey, decoded.PublicKey)
	assertEqualString(t, "IP", original.IP, decoded.IP)
	assertEqualString(t, "Name", original.Name, decoded.Name)

	t.Logf("✓ WireGuard 编解码测试通过，名称: %s", decoded.Name)
}

// TestWireGuardDecodeURL 测试用户提供的 WireGuard URL 解析
func TestWireGuardDecodeURL(t *testing.T) {
	// 用户提供的示例 URL
	testURL := "wireguard://OOrigZsSjw2YaY4urjbbU4%2FBNOZKXqW6EYNm8XKLtkU%3D@162.159.192.127:7152/?publickey=bmXOC%2BF1FxEMF9dyiK2H5%2F1SUtzH0JuVo51h2wPfgyo%3D&address=172.16.0.2%2F32%2C2606%3A4700%3A110%3A82ce%3Abdeb%3Ae72d%3A572a%3Ae280%2F128&mtu=1280#162.159.192.127%3A7152"

	decoded, err := DecodeWireGuardURL(testURL)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证解析结果
	assertEqualString(t, "Server", "162.159.192.127", decoded.Server)
	assertEqualIntInterface(t, "Port", 7152, decoded.Port)
	assertEqualString(t, "IP", "172.16.0.2", decoded.IP)

	t.Logf("✓ WireGuard URL 解析测试通过")
	t.Logf("  Server: %s", decoded.Server)
	t.Logf("  Port: %v", decoded.Port)
	t.Logf("  PrivateKey: %s", decoded.PrivateKey)
	t.Logf("  PublicKey: %s", decoded.PublicKey)
	t.Logf("  IP: %s", decoded.IP)
	t.Logf("  IPv6: %s", decoded.IPv6)
	t.Logf("  MTU: %d", decoded.MTU)
}

// TestWireGuardConfigParse 测试标准 WireGuard 配置文件解析
func TestWireGuardConfigParse(t *testing.T) {
	config := `[Interface]
Address = 172.16.0.2/32,2606:4700:110:82ce:bdeb:e72d:572a:e280/128
PrivateKey = OOrigZsSjw2YaY4urjbbU4/BNOZKXqW6EYNm8XKLtkU=
DNS = 1.1.1.1
MTU = 1280

[Peer]
AllowedIPs = 0.0.0.0/0
Endpoint = 162.159.192.127:7152
PublicKey = bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=`

	// 检测是否为配置文件格式
	if !IsWireGuardConfig(config) {
		t.Fatal("应该识别为 WireGuard 配置文件格式")
	}

	// 解析配置
	wg, err := ParseWireGuardConfig(config)
	if err != nil {
		t.Fatalf("解析配置失败: %v", err)
	}

	// 验证解析结果
	assertEqualString(t, "Server", "162.159.192.127", wg.Server)
	assertEqualIntInterface(t, "Port", 7152, wg.Port)
	assertEqualString(t, "PrivateKey", "OOrigZsSjw2YaY4urjbbU4/BNOZKXqW6EYNm8XKLtkU=", wg.PrivateKey)
	assertEqualString(t, "PublicKey", "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo=", wg.PublicKey)
	assertEqualString(t, "IP", "172.16.0.2", wg.IP)
	assertEqualString(t, "DNS", "1.1.1.1", wg.DNS)

	t.Logf("✓ WireGuard 配置文件解析测试通过")
	t.Logf("  Server: %s", wg.Server)
	t.Logf("  Port: %v", wg.Port)
	t.Logf("  IP: %s", wg.IP)
	t.Logf("  IPv6: %s", wg.IPv6)
}

// TestWireGuardNameModification 测试 WireGuard 名称修改
func TestWireGuardNameModification(t *testing.T) {
	original := WireGuard{
		Name:       "原始名称",
		Server:     "example.com",
		Port:       51820,
		PrivateKey: "test-private-key=",
		PublicKey:  "test-public-key=",
		IP:         "10.0.0.1",
	}

	newName := "新名称-WireGuard-测试"
	encoded := EncodeWireGuardURL(original)
	decoded, _ := DecodeWireGuardURL(encoded)
	decoded.Name = newName
	reEncoded := EncodeWireGuardURL(decoded)
	final, _ := DecodeWireGuardURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Server, final.Server)
	assertEqualString(t, "私钥(不变)", original.PrivateKey, final.PrivateKey)
	assertEqualString(t, "公钥(不变)", original.PublicKey, final.PublicKey)

	t.Logf("✓ WireGuard 名称修改测试通过: %s -> %s", original.Name, final.Name)
}

// TestIsWireGuardConfig 测试配置文件格式检测
func TestIsWireGuardConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"有效配置", "[Interface]\nAddress = 10.0.0.1\n[Peer]\nEndpoint = x", true},
		{"仅Interface", "[Interface]\nAddress = 10.0.0.1", false},
		{"仅Peer", "[Peer]\nEndpoint = x", false},
		{"URL格式", "wireguard://xxx@server:port", false},
		{"空字符串", "", false},
	}

	for _, tt := range tests {
		result := IsWireGuardConfig(tt.input)
		if result != tt.expected {
			t.Errorf("%s: 期望 %v, 实际 %v", tt.name, tt.expected, result)
		}
	}

	t.Log("✓ WireGuard 配置文件格式检测测试通过")
}
