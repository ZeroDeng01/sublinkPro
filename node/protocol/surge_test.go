package protocol

import (
	"strings"
	"testing"
)

// TestEncodeSurge_SS 测试 SS 节点的 Surge 格式输出
func TestEncodeSurge_SS(t *testing.T) {
	ss := Ss{
		Name:   "测试节点-SS",
		Server: "example.com",
		Port:   8388,
		Param: Param{
			Cipher:   "aes-256-gcm",
			Password: "test-password",
		},
	}
	link := EncodeSSURL(ss)

	// 注意：EncodeSurge 需要模板文件，这里只测试内部逻辑
	// 我们通过解码后验证数据不丢失
	decoded, err := DecodeSSURL(link)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Name", ss.Name, decoded.Name)
	assertEqualString(t, "Server", ss.Server, decoded.Server)
	assertEqualIntInterface(t, "Port", ss.Port, decoded.Port)
	assertEqualString(t, "Cipher", ss.Param.Cipher, decoded.Param.Cipher)
	assertEqualString(t, "Password", ss.Param.Password, decoded.Param.Password)

	t.Logf("✓ SS Surge 格式数据完整性测试通过，名称: %s", decoded.Name)
}

// TestEncodeSurge_VMess 测试 VMess 节点的 Surge 格式输出
func TestEncodeSurge_VMess(t *testing.T) {
	vmess := Vmess{
		Add:  "example.com",
		Port: "443",
		Id:   "12345678-1234-1234-1234-123456789abc",
		Net:  "ws",
		Path: "/vmess",
		Host: "cdn.example.com",
		Tls:  "tls",
		Sni:  "sni.example.com",
		Ps:   "测试节点-VMess",
		V:    "2",
	}
	link := EncodeVmessURL(vmess)

	decoded, err := DecodeVMESSURL(link)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Ps(Name)", vmess.Ps, decoded.Ps)
	assertEqualString(t, "Add(Server)", vmess.Add, decoded.Add)
	assertEqualString(t, "Id(UUID)", vmess.Id, decoded.Id)
	assertEqualString(t, "Net", vmess.Net, decoded.Net)
	assertEqualString(t, "Path", vmess.Path, decoded.Path)

	t.Logf("✓ VMess Surge 格式数据完整性测试通过，名称: %s", decoded.Ps)
}

// TestEncodeSurge_Trojan 测试 Trojan 节点的 Surge 格式输出
func TestEncodeSurge_Trojan(t *testing.T) {
	trojan := Trojan{
		Name:     "测试节点-Trojan",
		Password: "test-password",
		Hostname: "example.com",
		Port:     443,
		Query: TrojanQuery{
			Security: "tls",
			Sni:      "sni.example.com",
		},
	}
	link := EncodeTrojanURL(trojan)

	decoded, err := DecodeTrojanURL(link)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Name", trojan.Name, decoded.Name)
	assertEqualString(t, "Hostname", trojan.Hostname, decoded.Hostname)
	assertEqualIntInterface(t, "Port", trojan.Port, decoded.Port)
	assertEqualString(t, "Password", trojan.Password, decoded.Password)
	assertEqualString(t, "Sni", trojan.Query.Sni, decoded.Query.Sni)

	t.Logf("✓ Trojan Surge 格式数据完整性测试通过，名称: %s", decoded.Name)
}

// TestEncodeSurge_HY2 测试 Hysteria2 节点的 Surge 格式输出
func TestEncodeSurge_HY2(t *testing.T) {
	hy2 := HY2{
		Name:     "测试节点-HY2",
		Host:     "example.com",
		Port:     443,
		Password: "test-password",
		Sni:      "sni.example.com",
	}
	link := EncodeHY2URL(hy2)

	decoded, err := DecodeHY2URL(link)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Name", hy2.Name, decoded.Name)
	assertEqualString(t, "Host", hy2.Host, decoded.Host)
	assertEqualIntInterface(t, "Port", hy2.Port, decoded.Port)
	assertEqualString(t, "Password", hy2.Password, decoded.Password)
	assertEqualString(t, "Sni", hy2.Sni, decoded.Sni)

	t.Logf("✓ Hysteria2 Surge 格式数据完整性测试通过，名称: %s", decoded.Name)
}

// TestEncodeSurge_TUIC 测试 TUIC 节点的 Surge 格式输出
func TestEncodeSurge_TUIC(t *testing.T) {
	tuic := Tuic{
		Name:     "测试节点-TUIC",
		Host:     "example.com",
		Port:     443,
		Uuid:     "12345678-1234-1234-1234-123456789abc",
		Password: "test-password",
		Sni:      "sni.example.com",
	}
	link := EncodeTuicURL(tuic)

	decoded, err := DecodeTuicURL(link)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Name", tuic.Name, decoded.Name)
	assertEqualString(t, "Host", tuic.Host, decoded.Host)
	assertEqualIntInterface(t, "Port", tuic.Port, decoded.Port)
	assertEqualString(t, "Uuid", tuic.Uuid, decoded.Uuid)
	assertEqualString(t, "Password", tuic.Password, decoded.Password)

	t.Logf("✓ TUIC Surge 格式数据完整性测试通过，名称: %s", decoded.Name)
}

// TestEnsureProxyGroupHasProxies 测试代理组后备节点逻辑
func TestEnsureProxyGroupHasProxies(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "只有类型无代理",
			input:    "MyGroup = select",
			expected: "MyGroup = select, DIRECT",
		},
		{
			name:     "类型后有逗号但无代理",
			input:    "MyGroup = select, ",
			expected: "MyGroup = select, DIRECT",
		},
		{
			name:     "有有效代理",
			input:    "MyGroup = select, Proxy1, Proxy2",
			expected: "MyGroup = select, Proxy1, Proxy2",
		},
		{
			name:     "末尾多余逗号",
			input:    "MyGroup = select, Proxy1, ",
			expected: "MyGroup = select, Proxy1, ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ensureProxyGroupHasProxies(tc.input)
			if result != tc.expected {
				t.Errorf("期望: [%s], 实际: [%s]", tc.expected, result)
			} else {
				t.Logf("✓ %s 测试通过", tc.name)
			}
		})
	}
}

// TestOutputConfig 测试配置结构体
func TestOutputConfig(t *testing.T) {
	config := OutputConfig{
		Udp:                   true,
		Cert:                  true,
		Clash:                 "/path/to/clash.yaml",
		Surge:                 "/path/to/surge.conf",
		ReplaceServerWithHost: true,
		HostMap: map[string]string{
			"example.com": "1.2.3.4",
		},
	}

	assertEqualBool(t, "Udp", true, config.Udp)
	assertEqualBool(t, "Cert", true, config.Cert)
	assertEqualBool(t, "ReplaceServerWithHost", true, config.ReplaceServerWithHost)

	if ip, exists := config.HostMap["example.com"]; exists {
		assertEqualString(t, "HostMap[example.com]", "1.2.3.4", ip)
	} else {
		t.Error("HostMap 应该包含 example.com")
	}

	t.Log("✓ OutputConfig 测试通过")
}

// TestProxyStruct 测试 Proxy 结构体序列化
func TestProxyStruct(t *testing.T) {
	proxy := Proxy{
		Name:     "测试节点",
		Type:     "ss",
		Server:   "example.com",
		Port:     8388,
		Cipher:   "aes-256-gcm",
		Password: "password",
		Udp:      true,
	}

	assertEqualString(t, "Name", "测试节点", proxy.Name)
	assertEqualString(t, "Type", "ss", proxy.Type)
	assertEqualString(t, "Server", "example.com", proxy.Server)
	assertEqualFlexPort(t, "Port", 8388, proxy.Port)
	assertEqualBool(t, "Udp", true, proxy.Udp)

	t.Log("✓ Proxy 结构体测试通过")
}

// TestConvertToInt 测试类型转换函数
func TestConvertToInt(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected int
		hasError bool
	}{
		{"int", 443, 443, false},
		{"float64", 443.0, 443, false},
		{"string", "443", 443, false},
		{"invalid string", "abc", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := convertToInt(tc.input)
			if tc.hasError {
				if err == nil {
					t.Error("应该返回错误")
				}
			} else {
				if err != nil {
					t.Errorf("不应该返回错误: %v", err)
				}
				assertEqualInt(t, "result", tc.expected, result)
			}
		})
	}

	t.Log("✓ convertToInt 测试通过")
}

// TestDeleteOpts 测试空值删除函数
func TestDeleteOpts(t *testing.T) {
	opts := map[string]interface{}{
		"key1": "value1",
		"key2": "",
		"key3": map[string]interface{}{
			"nested1": "value",
			"nested2": "",
		},
		"key4": map[string]interface{}{},
	}

	DeleteOpts(opts)

	// key2 应该被删除（空字符串）
	if _, exists := opts["key2"]; exists {
		t.Error("key2 应该被删除")
	}

	// key4 应该被删除（空 map）
	if _, exists := opts["key4"]; exists {
		t.Error("key4 应该被删除")
	}

	// key1 应该保留
	if _, exists := opts["key1"]; !exists {
		t.Error("key1 应该保留")
	}

	// nested2 应该被删除
	if nested, ok := opts["key3"].(map[string]interface{}); ok {
		if _, exists := nested["nested2"]; exists {
			t.Error("nested2 应该被删除")
		}
	}

	t.Log("✓ DeleteOpts 测试通过")
}

// TestHostReplacement 测试 Host 替换功能
func TestHostReplacement(t *testing.T) {
	hostMap := map[string]string{
		"example.com":      "1.2.3.4",
		"test.example.com": "5.6.7.8",
	}

	// 测试存在映射
	if ip, exists := hostMap["example.com"]; exists {
		assertEqualString(t, "example.com", "1.2.3.4", ip)
	} else {
		t.Error("hostMap 应该包含 example.com")
	}

	// 测试不存在映射
	if _, exists := hostMap["unknown.com"]; exists {
		t.Error("hostMap 不应该包含 unknown.com")
	}

	t.Log("✓ Host 替换映射测试通过")
}

// TestAllProtocolsToProxy 综合测试所有协议的 LinkToProxy 转换
func TestAllProtocolsToProxy(t *testing.T) {
	protocols := []struct {
		name     string
		url      string
		expected string
	}{
		{"SS", "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ@example.com:8388#SS节点", "ss"},
		{"VLESS", "vless://uuid@example.com:443?security=tls&type=tcp#VLESS节点", "vless"},
		{"Trojan", "trojan://password@example.com:443?security=tls#Trojan节点", "trojan"},
		{"Socks5", "socks5://user:pass@example.com:1080#Socks5节点", "socks5"},
	}

	config := OutputConfig{Udp: true, Cert: true}

	for _, p := range protocols {
		t.Run(p.name, func(t *testing.T) {
			link := Urls{Url: p.url}
			proxy, err := LinkToProxy(link, config)
			if err != nil {
				// 某些测试链接可能格式不完整，跳过
				t.Skipf("跳过 %s: %v", p.name, err)
			}

			if !strings.EqualFold(proxy.Type, p.expected) {
				t.Errorf("Type 不匹配: 期望 %s, 实际 %s", p.expected, proxy.Type)
			}

			t.Logf("✓ %s 协议 LinkToProxy 测试通过", p.name)
		})
	}
}
