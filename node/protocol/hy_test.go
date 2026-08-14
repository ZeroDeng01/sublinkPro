package protocol

import (
	"strings"
	"testing"
)

// TestHYEncodeDecode 测试 Hysteria 编解码完整性
func TestHYEncodeDecode(t *testing.T) {
	original := HY{
		Name:     "测试节点-Hysteria",
		Host:     "example.com",
		Port:     443,
		Auth:     "test-auth-string",
		Peer:     "sni.example.com",
		Protocol: "udp",
		Insecure: 1,
		UpMbps:   100,
		DownMbps: 100,
	}

	// 编码
	encoded := EncodeHYURL(original)
	if !strings.HasPrefix(encoded, "hysteria://") && !strings.HasPrefix(encoded, "hy://") {
		t.Errorf("编码后应以 hysteria:// 或 hy:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeHYURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Host", original.Host, decoded.Host)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Peer", original.Peer, decoded.Peer)
	assertEqualString(t, "Auth", original.Auth, decoded.Auth)
	assertEqualString(t, "Protocol", original.Protocol, decoded.Protocol)
	assertEqualString(t, "Name", original.Name, decoded.Name)

	t.Logf("✓ Hysteria 编解码测试通过，名称: %s", decoded.Name)
}

// TestHYNameModification 测试 Hysteria 名称修改
func TestHYNameModification(t *testing.T) {
	original := HY{
		Name: "原始名称",
		Host: "example.com",
		Port: 443,
		Auth: "test-auth",
	}

	newName := "新名称-Hysteria-测试"
	encoded := EncodeHYURL(original)
	decoded, _ := DecodeHYURL(encoded)
	decoded.Name = newName
	reEncoded := EncodeHYURL(decoded)
	final, _ := DecodeHYURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Host, final.Host)

	t.Logf("✓ Hysteria 名称修改测试通过: %s -> %s", original.Name, final.Name)
}

// TestHY2EncodeDecode 测试 Hysteria2 编解码完整性
func TestHY2EncodeDecode(t *testing.T) {
	original := HY2{
		Name:              "测试节点-Hysteria2",
		Host:              "example.com",
		Port:              443,
		Password:          "test-hy2-password",
		Sni:               "sni.example.com",
		Insecure:          1,
		Obfs:              "salamander",
		ClientFingerprint: "chrome",
		Fingerprint:       "16dac3717024eb319093d1c95290c14adc850e2814b2208d11c7b7a436923859",
	}

	// 编码
	encoded := EncodeHY2URL(original)
	if !strings.HasPrefix(encoded, "hysteria2://") && !strings.HasPrefix(encoded, "hy2://") {
		t.Errorf("编码后应以 hysteria2:// 或 hy2:// 开头, 实际: %s", encoded)
	}

	// 解码
	decoded, err := DecodeHY2URL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	// 验证关键字段
	assertEqualString(t, "Host", original.Host, decoded.Host)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Password", original.Password, decoded.Password)
	assertEqualString(t, "ClientFingerprint", original.ClientFingerprint, decoded.ClientFingerprint)
	assertEqualString(t, "Fingerprint", original.Fingerprint, decoded.Fingerprint)
	assertEqualString(t, "Name", original.Name, decoded.Name)

	proxy, err := buildHY2Proxy(Urls{Url: encoded}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildHY2Proxy 失败: %v", err)
	}
	assertEqualString(t, "ProxyClientFingerprint", original.ClientFingerprint, proxy.Client_fingerprint)
	assertEqualString(t, "ProxyFingerprint", original.Fingerprint, proxy.Fingerprint)

	t.Logf("✓ Hysteria2 编解码测试通过，名称: %s", decoded.Name)
}

func TestHY2RejectsInvalidCertificateFingerprint(t *testing.T) {
	link := "hy2://password@example.com:443/?sni=example.com&pinSHA256=abc%2Cskip-cert-verify%3Dtrue#bad-fingerprint"

	decoded, err := DecodeHY2URL(link)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	assertEqualString(t, "InvalidFingerprintDropped", "", decoded.Fingerprint)

	proxy, err := buildHY2Proxy(Urls{Url: link}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildHY2Proxy 失败: %v", err)
	}
	assertEqualString(t, "ProxyInvalidFingerprintDropped", "", proxy.Fingerprint)
}

// TestHY2NameModification 测试 Hysteria2 名称修改
func TestHY2NameModification(t *testing.T) {
	original := HY2{
		Name:     "原始名称",
		Host:     "example.com",
		Port:     443,
		Password: "test-password",
	}

	newName := "新名称-Hysteria2-测试"
	encoded := EncodeHY2URL(original)
	decoded, _ := DecodeHY2URL(encoded)
	decoded.Name = newName
	reEncoded := EncodeHY2URL(decoded)
	final, _ := DecodeHY2URL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Host, final.Host)
	assertEqualString(t, "密码(不变)", original.Password, final.Password)

	t.Logf("✓ Hysteria2 名称修改测试通过: %s -> %s", original.Name, final.Name)
}

// TestHY2IPv6RawUpdatePreservesBracketedAuthority 覆盖原始信息编辑器保存 IPv6 HY2 节点的回归场景。
// DecodeHY2URL 会把 Host 规范化为不带方括号的 IPv6；再次编码时必须恢复 URL authority 所需的方括号。
func TestHY2IPv6RawUpdatePreservesBracketedAuthority(t *testing.T) {
	original := "hy2://test-password@[2001:db8::3]:22000?insecure=1&sni=example.com#ipv6-hy2-test"

	updated, err := UpdateNodeLinkFields(original, `{"Host":"2001:db8::3","Sni":"example.com"}`)
	if err != nil {
		t.Fatalf("更新 HY2 IPv6 原始字段失败: %v", err)
	}
	if !strings.Contains(updated, "@[2001:db8::3]:22000") {
		t.Fatalf("IPv6 authority 应保留方括号，实际链接: %s", updated)
	}

	decoded, err := DecodeHY2URL(updated)
	if err != nil {
		t.Fatalf("回写后的 HY2 IPv6 链接应可解析: %v", err)
	}
	assertEqualString(t, "Host", "2001:db8::3", decoded.Host)
	assertEqualIntInterface(t, "Port", 22000, decoded.Port)

	identity, err := ExtractLinkIdentity(updated)
	if err != nil {
		t.Fatalf("提取 HY2 IPv6 身份信息失败: %v", err)
	}
	assertEqualString(t, "Address", "[2001:db8::3]:22000", identity.Address)
}

func TestHY2SurgeLinePortHopping(t *testing.T) {
	link := EncodeHY2URL(HY2{
		Name:     "Surge-HY2-Hop",
		Host:     "example.com",
		Port:     443,
		Password: "pw",
		MPort:    "10000-20000",
	})

	line, name, err := buildHY2SurgeLine(link, OutputConfig{})
	if err != nil {
		t.Fatalf("buildHY2SurgeLine 失败: %v", err)
	}
	assertEqualString(t, "SurgeName", "Surge-HY2-Hop", name)
	assertContains(t, "SurgeType", line, "Surge-HY2-Hop = hysteria2, example.com, 443")
	assertContains(t, "PortHopping", line, "port-hopping=10000-20000")
}

func TestHY2SurgeLineNoPortHoppingWhenMPortEmpty(t *testing.T) {
	link := EncodeHY2URL(HY2{
		Name:     "Surge-HY2-Plain",
		Host:     "example.com",
		Port:     443,
		Password: "pw",
	})

	line, _, err := buildHY2SurgeLine(link, OutputConfig{})
	if err != nil {
		t.Fatalf("buildHY2SurgeLine 失败: %v", err)
	}
	assertContains(t, "SurgeType", line, "Surge-HY2-Plain = hysteria2, example.com, 443")
	if strings.Contains(line, "port-hopping") {
		t.Errorf("无 mport 时不应输出 port-hopping, 实际: %s", line)
	}
}

// TestHY2SurgeLinePortZeroUsesMportFirstPort 覆盖相邻 codex 审核指出的 P1：
// 原始链接携带 :0（实际端口由 mport 提供）时，Surge 主端口必须取 mport 首段作占位，
// 否则输出 "port = 0" 会被 Surge 判为 "The value of port is invalid"。
func TestHY2SurgeLinePortZeroUsesMportFirstPort(t *testing.T) {
	line, name, err := buildHY2SurgeLine("hy2://pw@example.com:0?mport=10000-20000,30000#Surge-Zero", OutputConfig{})
	if err != nil {
		t.Fatalf("buildHY2SurgeLine 失败: %v", err)
	}
	assertEqualString(t, "SurgeName", "Surge-Zero", name)
	assertContains(t, "SurgeMainPort", line, " = hysteria2, example.com, 10000,")
	assertContains(t, "PortHopping", line, "port-hopping=10000-20000;30000")
}

// TestHY2SurgeLineMportSeparatorToSemicolon 覆盖审核指出的 P1：
// mport 多段以逗号分隔（10000-20000,30000），Surge 的 port-hopping 要求分号；
// 直接透传逗号会让 Surge 将后续段解析成独立无效参数。
func TestHY2SurgeLineMportSeparatorToSemicolon(t *testing.T) {
	line, _, err := buildHY2SurgeLine("hy2://pw@example.com:443?mport=20000-30000,40000#Surge-Sep", OutputConfig{})
	if err != nil {
		t.Fatalf("buildHY2SurgeLine 失败: %v", err)
	}
	if strings.Contains(line, ",40000") {
		t.Errorf("mport 多段逗号分隔应转换为分号, 实际: %s", line)
	}
	assertContains(t, "PortHopping", line, "port-hopping=20000-30000;40000")
}

func TestHY2SurgeLineMportNormalizesMihomoSyntax(t *testing.T) {
	testCases := []struct {
		name string
		link string
		want string
	}{
		{
			name: "端口段空白",
			link: "hy2://pw@example.com:443?mport=%2010000-20000%2C%2030000%20#Surge-Space",
			want: "port-hopping=10000-20000;30000",
		},
		{
			name: "斜杠分隔",
			link: "hy2://pw@example.com:443?mport=10000-20000%2F30000#Surge-Slash",
			want: "port-hopping=10000-20000;30000",
		},
		{
			name: "范围端点空白",
			link: "hy2://pw@example.com:443?mport=10000%20-%2020000%2F%2030000#Surge-Range-Space",
			want: "port-hopping=10000-20000;30000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			line, _, err := buildHY2SurgeLine(tc.link, OutputConfig{})
			if err != nil {
				t.Fatalf("buildHY2SurgeLine 失败: %v", err)
			}
			assertContains(t, "PortHopping", line, tc.want)
		})
	}
}

func TestHY2SurgeLineMportRejectsSemicolonInput(t *testing.T) {
	line, _, err := buildHY2SurgeLine("hy2://pw@example.com:443?mport=10000-20000%3B30000#Surge-Semicolon", OutputConfig{})
	if err != nil {
		t.Fatalf("buildHY2SurgeLine 失败: %v", err)
	}
	if strings.Contains(line, "port-hopping") {
		t.Errorf("mihomo mport 源输入不接受分号, 实际: %s", line)
	}
}

// TestHY2SurgeLinePortZeroNoMportFallback443 覆盖主端口无效且无端口跳跃时回退默认端口 443。
func TestHY2SurgeLinePortZeroNoMportFallback443(t *testing.T) {
	line, _, err := buildHY2SurgeLine("hy2://pw@example.com:0#Surge-Fallback", OutputConfig{})
	if err != nil {
		t.Fatalf("buildHY2SurgeLine 失败: %v", err)
	}
	assertContains(t, "SurgeMainPort", line, " = hysteria2, example.com, 443,")
	if strings.Contains(line, "port-hopping") {
		t.Errorf("无 mport 时不应输出 port-hopping, 实际: %s", line)
	}
}

// TestHY2SurgeLineMportRejectsInjection 覆盖相邻 codex 复核提出的 P2：
// mport 白名单校验——URL 解码出的控制字符/换行等非法字符混入 mport 时丢弃整段，
// 既不输出 port-hopping 也不把注入内容拼进生成的 Surge profile。
func TestHY2SurgeLineMportRejectsInjection(t *testing.T) {
	line, _, err := buildHY2SurgeLine("hy2://pw@example.com:443?mport=443%0A[Rule]%0AFINAL%2CDIRECT#Surge-Inject", OutputConfig{})
	if err != nil {
		t.Fatalf("buildHY2SurgeLine 失败: %v", err)
	}
	if strings.Contains(line, "port-hopping") {
		t.Errorf("非法 mport 不应输出 port-hopping, 实际: %s", line)
	}
	if strings.Contains(line, "[Rule]") || strings.Contains(line, "\n") {
		t.Errorf("非法 mport 不应注入 Surge 行, 实际: %s", line)
	}
}

// TestHY2SurgeLineMportRangeOrderAndBound 覆盖 codex 三轮复核提出的 P2：
// mport 语义校验——降序范围、端口 0、超 65535 等端口段 Surge 判 "port-hopping is invalid"，
// 应整段丢弃不输出，避免生成无法加载的 profile。
func TestHY2SurgeLineMportRangeOrderAndBound(t *testing.T) {
	badMports := []string{"5000-4000", "0", "70000", "10000-20000,0", "10000-20000,300000"}
	for _, mport := range badMports {
		link := "hy2://pw@example.com:443?mport=" + mport + "#Surge-BadRange"
		line, _, err := buildHY2SurgeLine(link, OutputConfig{})
		if err != nil {
			t.Fatalf("buildHY2SurgeLine 失败 (%s): %v", mport, err)
		}
		if strings.Contains(line, "port-hopping") {
			t.Errorf("非法范围 mport=%q 不应输出 port-hopping, 实际: %s", mport, line)
		}
	}
}

func TestHY2SurgeLineInvalidMainPortFallback(t *testing.T) {
	testCases := []struct {
		name        string
		link        string
		wantPort    string
		wantHopping string
	}{
		{
			name:        "超界主端口取 mport 首段",
			link:        "hy2://pw@example.com:70000?mport=10000-20000,30000#Surge-HighPort",
			wantPort:    " = hysteria2, example.com, 10000,",
			wantHopping: "port-hopping=10000-20000;30000",
		},
		{
			name:     "超界主端口无 mport 回退 443",
			link:     "hy2://pw@example.com:70000#Surge-HighPortFallback",
			wantPort: " = hysteria2, example.com, 443,",
		},
		{
			name:     "零端口加非法 mport 回退 443",
			link:     "hy2://pw@example.com:0?mport=70000#Surge-InvalidMportFallback",
			wantPort: " = hysteria2, example.com, 443,",
		},
		{
			name:        "缺省主端口沿用协议默认 443",
			link:        "hy2://pw@example.com?mport=10000-20000#Surge-MissingPort",
			wantPort:    " = hysteria2, example.com, 443,",
			wantHopping: "port-hopping=10000-20000",
		},
		{
			name: "Encode 过滤零值 mport 后回退 443",
			link: EncodeHY2URL(HY2{
				Name:     "Surge-Encoded-Zero",
				Host:     "example.com",
				Port:     0,
				Password: "pw",
				MPort:    "0",
			}),
			wantPort: " = hysteria2, example.com, 443,",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			line, _, err := buildHY2SurgeLine(tc.link, OutputConfig{})
			if err != nil {
				t.Fatalf("buildHY2SurgeLine 失败: %v", err)
			}
			assertContains(t, "SurgeMainPort", line, tc.wantPort)
			if tc.wantHopping != "" {
				assertContains(t, "PortHopping", line, tc.wantHopping)
			} else if strings.Contains(line, "port-hopping") {
				t.Errorf("无合法 mport 时不应输出 port-hopping, 实际: %s", line)
			}
		})
	}
}
