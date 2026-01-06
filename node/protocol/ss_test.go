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

// TestSSWithObfsPlugin 测试带 obfs 插件的 SS 解码
func TestSSWithObfsPlugin(t *testing.T) {
	// SIP002 格式带 obfs 插件的 SS 链接
	url := "ss://YWVzLTEyOC1nY206dGVzdA@192.168.1.1:8388/?plugin=obfs-local%3Bobfs%3Dhttp%3Bobfs-host%3Dbing.com#ObfsTest"

	ss, err := DecodeSSURL(url)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Server", "192.168.1.1", ss.Server)
	assertEqualIntInterface(t, "Port", 8388, ss.Port)
	assertEqualString(t, "Name", "ObfsTest", ss.Name)

	if ss.Plugin == nil {
		t.Fatal("Plugin 应该不为空")
	}
	assertEqualString(t, "Plugin.Name", "obfs-local", ss.Plugin.Name)
	assertEqualString(t, "Plugin.Opts[obfs]", "http", ss.Plugin.Opts["obfs"])
	assertEqualString(t, "Plugin.Opts[obfs-host]", "bing.com", ss.Plugin.Opts["obfs-host"])

	t.Logf("✓ SS obfs 插件解码测试通过")
}

// TestSSWithV2rayPlugin 测试带 v2ray-plugin 的 SS 解码
func TestSSWithV2rayPlugin(t *testing.T) {
	// v2ray-plugin 格式
	url := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ@example.com:443/?plugin=v2ray-plugin%3Bmode%3Dwebsocket%3Bhost%3Dexample.com%3Bpath%3D%2Fws%3Btls#V2RayTest"

	ss, err := DecodeSSURL(url)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Server", "example.com", ss.Server)
	assertEqualString(t, "Name", "V2RayTest", ss.Name)

	if ss.Plugin == nil {
		t.Fatal("Plugin 应该不为空")
	}
	assertEqualString(t, "Plugin.Name", "v2ray-plugin", ss.Plugin.Name)
	assertEqualString(t, "Plugin.Opts[mode]", "websocket", ss.Plugin.Opts["mode"])
	assertEqualString(t, "Plugin.Opts[host]", "example.com", ss.Plugin.Opts["host"])
	assertEqualString(t, "Plugin.Opts[path]", "/ws", ss.Plugin.Opts["path"])

	t.Logf("✓ SS v2ray-plugin 解码测试通过")
}

// TestSSWithShadowTLS 测试带 shadow-tls 插件的 SS 解码
func TestSSWithShadowTLS(t *testing.T) {
	url := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ@server.com:443/?plugin=shadow-tls%3Bhost%3Dcloud.tencent.com%3Bpassword%3Dsecret%3Bversion%3D2#ShadowTLSTest"

	ss, err := DecodeSSURL(url)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	if ss.Plugin == nil {
		t.Fatal("Plugin 应该不为空")
	}
	assertEqualString(t, "Plugin.Name", "shadow-tls", ss.Plugin.Name)
	assertEqualString(t, "Plugin.Opts[host]", "cloud.tencent.com", ss.Plugin.Opts["host"])
	assertEqualString(t, "Plugin.Opts[password]", "secret", ss.Plugin.Opts["password"])
	assertEqualString(t, "Plugin.Opts[version]", "2", ss.Plugin.Opts["version"])

	t.Logf("✓ SS shadow-tls 插件解码测试通过")
}

// TestSSPluginEncodeDecode 测试带插件的 SS 编解码完整性
func TestSSPluginEncodeDecode(t *testing.T) {
	original := Ss{
		Name:   "插件测试节点",
		Server: "example.com",
		Port:   8388,
		Param: Param{
			Cipher:   "aes-256-gcm",
			Password: "test-password",
		},
		Plugin: &SsPlugin{
			Name: "obfs-local",
			Opts: map[string]string{
				"mode": "http",
				"host": "www.bing.com",
			},
		},
	}

	// 编码
	encoded := EncodeSSURL(original)
	if !strings.Contains(encoded, "plugin=") {
		t.Errorf("编码后应包含 plugin 参数, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeSSURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证基本字段
	assertEqualString(t, "Server", original.Server, decoded.Server)
	assertEqualString(t, "Name", original.Name, decoded.Name)

	// 验证插件
	if decoded.Plugin == nil {
		t.Fatal("解码后 Plugin 应该不为空")
	}
	assertEqualString(t, "Plugin.Name", original.Plugin.Name, decoded.Plugin.Name)

	t.Logf("✓ SS 插件编解码完整性测试通过")
}

// TestSSWithoutPlugin 测试无插件的 SS 链接保持兼容
func TestSSWithoutPlugin(t *testing.T) {
	url := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ@server.com:8388#NoPlugin"

	ss, err := DecodeSSURL(url)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Server", "server.com", ss.Server)
	assertEqualString(t, "Name", "NoPlugin", ss.Name)

	if ss.Plugin != nil {
		t.Error("无插件的链接 Plugin 应该为 nil")
	}

	t.Logf("✓ SS 无插件兼容性测试通过")
}
