package protocol

import (
	"net/url"
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

// TestTuicInsecureParamUsesClientCompatibleName 验证跳过证书校验输出为客户端识别的 allow_insecure 参数。
// 主流客户端（v2rayN、NekoBox 等）的 TUIC 链接均读取 allow_insecure，而非项目历史使用的非标准 insecure。
func TestTuicInsecureParamUsesClientCompatibleName(t *testing.T) {
	encoded := EncodeTuicURL(Tuic{
		Name:     "自签证书节点",
		Host:     "example.com",
		Port:     443,
		Uuid:     "12345678-1234-1234-1234-123456789abc",
		Password: "test-tuic-password",
		Insecure: 1,
	})
	parsed, err := url.Parse(encoded)
	if err != nil {
		t.Fatalf("解析编码结果失败: %v, 链接: %s", err, encoded)
	}
	if got := parsed.Query().Get("allow_insecure"); got != "1" {
		t.Fatalf("allow_insecure = %q, want \"1\" (链接: %s)", got, encoded)
	}
	if got := parsed.Query().Get("insecure"); got != "" {
		t.Fatalf("不应再输出非标准 insecure 参数, 实际: %q (链接: %s)", got, encoded)
	}
}

// TestTuicDecodeAllowInsecure 验证解码兼容客户端使用的 allow_insecure，并保留对历史 insecure 的兼容。
func TestTuicDecodeAllowInsecure(t *testing.T) {
	allowInsecureLink := "tuic://12345678-1234-1234-1234-123456789abc:test-password@example.com:443?allow_insecure=1&sni=sni.example.com#allow_insecure节点"
	decoded, err := DecodeTuicURL(allowInsecureLink)
	if err != nil {
		t.Fatalf("解码 allow_insecure 链接失败: %v", err)
	}
	assertEqualInt(t, "Insecure", 1, decoded.Insecure)

	legacyLink := "tuic://12345678-1234-1234-1234-123456789abc:test-password@example.com:443?insecure=1#历史节点"
	legacy, err := DecodeTuicURL(legacyLink)
	if err != nil {
		t.Fatalf("解码历史 insecure 链接失败: %v", err)
	}
	assertEqualInt(t, "Insecure(legacy)", 1, legacy.Insecure)
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
