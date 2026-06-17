package protocol

import (
	"strings"
	"testing"
)

// ============================================================================
// 边界情况和协议特性综合测试
// ============================================================================

// TestEmptyNameFallback 测试空名称时的后备逻辑
func TestEmptyNameFallback(t *testing.T) {
	testCases := []struct {
		name     string
		protocol string
		encode   func() string
		decode   func(string) (string, error)
	}{
		{
			name:     "VMess空名称后备",
			protocol: "vmess",
			encode: func() string {
				v := Vmess{Add: "example.com", Port: "443", Id: "88888888-9999-7777-5555-777777777777", V: "2"}
				return EncodeVmessURL(v)
			},
			decode: func(s string) (string, error) {
				v, err := DecodeVMESSURL(s)
				return v.Ps, err
			},
		},
		{
			name:     "VLESS空名称后备",
			protocol: "vless",
			encode: func() string {
				v := VLESS{Server: "example.com", Port: 443, Uuid: "88888888-9999-7777-5555-777777777777"}
				return EncodeVLESSURL(v)
			},
			decode: func(s string) (string, error) {
				v, err := DecodeVLESSURL(s)
				return v.Name, err
			},
		},
		{
			name:     "Trojan空名称后备",
			protocol: "trojan",
			encode: func() string {
				t := Trojan{Hostname: "example.com", Port: 443, Password: "pass"}
				return EncodeTrojanURL(t)
			},
			decode: func(s string) (string, error) {
				t, err := DecodeTrojanURL(s)
				return t.Name, err
			},
		},
		{
			name:     "SS空名称后备",
			protocol: "ss",
			encode: func() string {
				s := Ss{Server: "example.com", Port: 8388, Param: Param{Cipher: "aes-256-gcm", Password: "pass"}}
				return EncodeSSURL(s)
			},
			decode: func(s string) (string, error) {
				ss, err := DecodeSSURL(s)
				return ss.Name, err
			},
		},
		{
			name:     "HTTP空名称后备",
			protocol: "http",
			encode: func() string {
				h := HTTP{Server: "example.com", Port: 8080, Username: "user", Password: "pass", TLS: false}
				return EncodeHTTPURL(h)
			},
			decode: func(s string) (string, error) {
				h, err := DecodeHTTPURL(s)
				return h.Name, err
			},
		},
		{
			name:     "HTTPS空名称后备",
			protocol: "https",
			encode: func() string {
				h := HTTP{Server: "example.com", Port: 443, Username: "user", Password: "pass", TLS: true}
				return EncodeHTTPURL(h)
			},
			decode: func(s string) (string, error) {
				h, err := DecodeHTTPURL(s)
				return h.Name, err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := tc.encode()
			name, err := tc.decode(encoded)
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}

			// 空名称应该后备为 server:port 格式
			if name == "" {
				t.Error("名称不应为空，应使用后备值")
			}
			if !strings.Contains(name, ":") {
				t.Errorf("后备名称应包含端口分隔符，实际: %s", name)
			}
			t.Logf("✓ %s 空名称后备测试通过: %s", tc.protocol, name)
		})
	}
}

// TestIPv6Address 测试 IPv6 地址处理
func TestIPv6Address(t *testing.T) {
	ipv6Cases := []struct {
		protocol string
		server   string
	}{
		{"vless", "[2001:db8::1]"},
		{"trojan", "[2001:db8::1]"},
		{"ss", "[2001:db8::1]"},
	}

	for _, tc := range ipv6Cases {
		t.Run(tc.protocol+"_ipv6", func(t *testing.T) {
			var encoded string
			expectedServer := "2001:db8::1"
			switch tc.protocol {
			case "vless":
				v := VLESS{Name: "IPv6测试", Server: tc.server, Port: 443, Uuid: "88888888-9999-7777-5555-777777777777"}
				encoded = EncodeVLESSURL(v)
				if !strings.Contains(encoded, tc.server) {
					t.Fatalf("编码结果应保留 IPv6 方括号格式: %s", encoded)
				}
				decoded, err := DecodeVLESSURL(encoded)
				if err != nil {
					t.Fatalf("解码失败: %v", err)
				}
				assertEqualString(t, "Server", expectedServer, decoded.Server)
			case "trojan":
				tr := Trojan{Name: "IPv6测试", Hostname: tc.server, Port: 443, Password: "pass"}
				encoded = EncodeTrojanURL(tr)
				if !strings.Contains(encoded, tc.server) {
					t.Fatalf("编码结果应保留 IPv6 方括号格式: %s", encoded)
				}
				decoded, err := DecodeTrojanURL(encoded)
				if err != nil {
					t.Fatalf("解码失败: %v", err)
				}
				assertEqualString(t, "Hostname", expectedServer, decoded.Hostname)
			case "ss":
				ss := Ss{Name: "IPv6测试", Server: tc.server, Port: 8388, Param: Param{Cipher: "aes-256-gcm", Password: "pass"}}
				encoded = EncodeSSURL(ss)
				if !strings.Contains(encoded, tc.server) {
					t.Fatalf("编码结果应保留 IPv6 方括号格式: %s", encoded)
				}
				decoded, err := DecodeSSURL(encoded)
				if err != nil {
					t.Fatalf("解码失败: %v", err)
				}
				assertEqualString(t, "Server", expectedServer, decoded.Server)
			}

			if !strings.Contains(encoded, "://") {
				t.Errorf("编码失败: %s", encoded)
			}
			t.Logf("✓ %s IPv6 编码测试通过", tc.protocol)
		})
	}
}

// TestRawUpdateHostAuthorityAcrossProtocols 覆盖原始信息编辑器保存不同主机形态时的通用 URL authority 回归场景。
// Raw 预览会把 IPv6 Host/Server 规范化为不带方括号；编码回链接时 IPv6 必须恢复 [IPv6]:port，域名、IPv4、单标签主机也不能被破坏。
func TestRawUpdateHostAuthorityAcrossProtocols(t *testing.T) {
	hostCases := []struct {
		name string
		host string
	}{
		{name: "IPv6", host: "2001:db8::3"},
		{name: "IPv4", host: "192.0.2.10"},
		{name: "Domain", host: "example.com"},
		{name: "SingleLabel", host: "nas"},
	}

	testCases := []struct {
		name      string
		fieldName string
		port      string
		prefix    string
		makeLink  func(host string) string
	}{
		{name: "VMess", fieldName: "Add", port: "443", makeLink: func(host string) string {
			return EncodeVmessURL(Vmess{Add: host, Port: 443, Id: "88888888-9999-7777-5555-777777777777", Ps: "raw-vmess-test", V: "2"})
		}},
		{name: "VLESS", fieldName: "Server", port: "443", prefix: "@", makeLink: func(host string) string {
			return "vless://88888888-9999-7777-5555-777777777777@" + expectedTestHostPort(host, "443") + "?encryption=none#raw-vless-test"
		}},
		{name: "Trojan", fieldName: "Hostname", port: "443", prefix: "@", makeLink: func(host string) string {
			return "trojan://test-password@" + expectedTestHostPort(host, "443") + "#raw-trojan-test"
		}},
		{name: "SS", fieldName: "Server", port: "8388", prefix: "@", makeLink: func(host string) string {
			return EncodeSSURL(Ss{Param: Param{Cipher: "aes-256-gcm", Password: "test-password"}, Server: host, Port: 8388, Name: "raw-ss-test"})
		}},
		{name: "SSR", fieldName: "Server", port: "8388", makeLink: func(host string) string {
			return EncodeSSRURL(Ssr{Server: host, Port: 8388, Protocol: "origin", Method: "aes-256-cfb", Obfs: "plain", Password: "test-password", Qurey: Ssrquery{Remarks: "raw-ssr-test"}})
		}},
		{name: "Hysteria", fieldName: "Host", port: "22000", prefix: "//", makeLink: func(host string) string {
			return "hysteria://" + expectedTestHostPort(host, "22000") + "?auth=test-auth&insecure=1#raw-hysteria-test"
		}},
		{name: "Hysteria2", fieldName: "Host", port: "22000", prefix: "@", makeLink: func(host string) string {
			return "hy2://test-password@" + expectedTestHostPort(host, "22000") + "?insecure=1&sni=example.com#raw-hy2-test"
		}},
		{name: "TUIC", fieldName: "Host", port: "443", prefix: "@", makeLink: func(host string) string {
			return "tuic://88888888-9999-7777-5555-777777777777:test-password@" + expectedTestHostPort(host, "443") + "?version=5#raw-tuic-test"
		}},
		{name: "AnyTLS", fieldName: "Server", port: "443", prefix: "@", makeLink: func(host string) string {
			return "anytls://test-password@" + expectedTestHostPort(host, "443") + "#raw-anytls-test"
		}},
		{name: "SOCKS5", fieldName: "Server", port: "1080", prefix: "@", makeLink: func(host string) string {
			return "socks5://test-user:test-password@" + expectedTestHostPort(host, "1080") + "#raw-socks5-test"
		}},
		{name: "HTTP", fieldName: "Server", port: "8080", prefix: "@", makeLink: func(host string) string {
			return "http://test-user:test-password@" + expectedTestHostPort(host, "8080") + "#raw-http-test"
		}},
		{name: "HTTPS", fieldName: "Server", port: "8443", prefix: "@", makeLink: func(host string) string {
			return "https://test-user:test-password@" + expectedTestHostPort(host, "8443") + "#raw-https-test"
		}},
		{name: "WireGuard", fieldName: "Server", port: "51820", prefix: "@", makeLink: func(host string) string {
			return "wireguard://test-private-key@" + expectedTestHostPort(host, "51820") + "?publickey=test-public-key&address=192.0.2.2%2F32#raw-wireguard-test"
		}},
		{name: "Mieru", fieldName: "Server", port: "2999", prefix: "@", makeLink: func(host string) string {
			return "mieru://test-user:test-password@" + expectedTestHostPort(host, "2999") + "?transport=TCP#raw-mieru-test"
		}},
	}

	for _, tc := range testCases {
		for _, hc := range hostCases {
			t.Run(tc.name+"_"+hc.name, func(t *testing.T) {
				updated, err := UpdateNodeLinkFields(tc.makeLink(hc.host), `{"`+tc.fieldName+`":"`+hc.host+`"}`)
				if err != nil {
					t.Fatalf("更新原始字段失败: %v", err)
				}
				if tc.prefix != "" {
					want := tc.prefix + expectedTestHostPort(hc.host, tc.port)
					if !strings.Contains(updated, want) {
						t.Fatalf("authority 不符合预期，期望包含 %q，实际链接: %s", want, updated)
					}
				}
				info, err := ParseNodeLink(updated)
				if err != nil {
					t.Fatalf("回写后的链接应可解析: %v", err)
				}
				if got := info.Fields[tc.fieldName]; got != hc.host {
					t.Fatalf("回写后字段 %s = %v，期望 %s", tc.fieldName, got, hc.host)
				}
			})
		}
	}
}

func expectedTestHostPort(host, port string) string {
	if strings.Contains(host, ":") {
		return "[" + host + "]:" + port
	}
	return host + ":" + port
}

// TestUnicodeInPassword 测试密码中的特殊字符
func TestUnicodeInPassword(t *testing.T) {
	specialPasswords := []string{
		"password123",
		"pass@word#123",
		"密码测试",
		"パスワード",
		"pass/word?test=1",
	}

	for _, pwd := range specialPasswords {
		t.Run("Trojan_"+pwd[:min(10, len(pwd))], func(t *testing.T) {
			original := Trojan{
				Name:     "测试节点",
				Hostname: "example.com",
				Port:     443,
				Password: pwd,
			}

			encoded := EncodeTrojanURL(original)
			decoded, err := DecodeTrojanURL(encoded)
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}

			if decoded.Password != pwd {
				t.Errorf("密码不匹配: 期望 [%s], 实际 [%s]", pwd, decoded.Password)
			} else {
				t.Logf("✓ 密码特殊字符测试通过: %s", pwd)
			}
		})
	}
}

// TestPortBoundary 测试端口边界值
func TestPortBoundary(t *testing.T) {
	ports := []int{1, 80, 443, 8080, 8388, 65535}

	for _, port := range ports {
		t.Run("VLESS_port_"+string(rune('0'+port%10)), func(t *testing.T) {
			original := VLESS{
				Name:   "端口测试",
				Server: "example.com",
				Port:   port,
				Uuid:   "88888888-9999-7777-5555-777777777777",
			}

			encoded := EncodeVLESSURL(original)
			decoded, err := DecodeVLESSURL(encoded)
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}

			assertEqualIntInterface(t, "Port", port, decoded.Port)
			t.Logf("✓ 端口 %d 测试通过", port)
		})
	}
}

// TestSSRBase64Password 测试 SSR 密码 Base64 编码
func TestSSRBase64Password(t *testing.T) {
	original := Ssr{
		Server:   "example.com",
		Port:     8388,
		Method:   "aes-256-cfb",
		Password: "test-password",
		Protocol: "origin",
		Obfs:     "plain",
		Qurey: Ssrquery{
			Remarks: "SSR密码测试",
		},
	}

	encoded := EncodeSSRURL(original)
	decoded, err := DecodeSSRURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Password", original.Password, decoded.Password)
	t.Log("✓ SSR 密码编解码测试通过")
}

// TestVMESSPortTypes 测试 VMess 端口类型处理
func TestVMESSPortTypes(t *testing.T) {
	// VMess 的 Port 是 interface{} 类型，可能是 string 或 int
	vmessWithStringPort := Vmess{
		Add:  "example.com",
		Port: "443",
		Id:   "88888888-9999-7777-5555-777777777777",
		Ps:   "String端口测试",
		V:    "2",
	}

	encoded := EncodeVmessURL(vmessWithStringPort)
	decoded, err := DecodeVMESSURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	assertEqualString(t, "Server", vmessWithStringPort.Add, decoded.Add)
	t.Log("✓ VMess 端口类型测试通过")
}

// TestTrojanAlpn 测试 Trojan ALPN 处理
func TestTrojanAlpn(t *testing.T) {
	original := Trojan{
		Name:     "ALPN测试",
		Hostname: "example.com",
		Port:     443,
		Password: "password",
		Query: TrojanQuery{
			Security: "tls",
			Alpn:     []string{"h2", "http/1.1"},
		},
	}

	encoded := EncodeTrojanURL(original)
	decoded, err := DecodeTrojanURL(encoded)
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}

	if len(decoded.Query.Alpn) != len(original.Query.Alpn) {
		t.Fatalf("ALPN 数量不匹配: 期望 %d, 实际 %d", len(original.Query.Alpn), len(decoded.Query.Alpn))
	}
	for i, alpn := range original.Query.Alpn {
		assertEqualString(t, "ALPN", alpn, decoded.Query.Alpn[i])
	}
	t.Log("✓ Trojan ALPN 测试通过")
}

// TestSSCipherMethods 测试各种加密方式
func TestSSCipherMethods(t *testing.T) {
	ciphers := []string{
		"aes-256-gcm",
		"aes-128-gcm",
		"chacha20-ietf-poly1305",
		"2022-blake3-aes-256-gcm",
	}

	for _, cipher := range ciphers {
		t.Run("SS_"+cipher, func(t *testing.T) {
			original := Ss{
				Name:   "加密测试-" + cipher,
				Server: "example.com",
				Port:   8388,
				Param: Param{
					Cipher:   cipher,
					Password: "password",
				},
			}

			encoded := EncodeSSURL(original)
			decoded, err := DecodeSSURL(encoded)
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}

			assertEqualString(t, "Cipher", cipher, decoded.Param.Cipher)
			t.Logf("✓ 加密方式 %s 测试通过", cipher)
		})
	}
}

// TestURLEncodingInPath 测试 WebSocket 路径中的特殊字符
func TestURLEncodingInPath(t *testing.T) {
	paths := []string{
		"/ws",
		"/path/to/websocket",
		"/ws?ed=2048",
		"/vmess?test=1&foo=bar",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			original := VLESS{
				Name:   "路径测试",
				Server: "example.com",
				Port:   443,
				Uuid:   "88888888-9999-7777-5555-777777777777",
				Query: VLESSQuery{
					Type: "ws",
					Path: path,
				},
			}

			encoded := EncodeVLESSURL(original)
			decoded, err := DecodeVLESSURL(encoded)
			if err != nil {
				t.Fatalf("解码失败: %v", err)
			}

			assertEqualString(t, "Path", path, decoded.Query.Path)
		})
	}
	t.Log("✓ URL 路径编码测试通过")
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestHTTPProtocolIntegration 测试HTTP协议的完整集成
func TestHTTPProtocolIntegration(t *testing.T) {
	t.Run("HTTP基本功能", func(t *testing.T) {
		original := HTTP{
			Name:     "HTTP测试节点",
			Server:   "example.com",
			Port:     8080,
			Username: "user",
			Password: "pass",
			TLS:      false,
		}

		encoded := EncodeHTTPURL(original)
		decoded, err := DecodeHTTPURL(encoded)
		if err != nil {
			t.Fatalf("解码失败: %v", err)
		}

		assertEqualString(t, "Name", original.Name, decoded.Name)
		assertEqualString(t, "Server", original.Server, decoded.Server)
		assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
		assertEqualString(t, "Username", original.Username, decoded.Username)
		assertEqualString(t, "Password", original.Password, decoded.Password)
		assertEqualBool(t, "TLS", original.TLS, decoded.TLS)
		t.Log("✓ HTTP基本功能测试通过")
	})

	t.Run("HTTPS完整功能", func(t *testing.T) {
		original := HTTP{
			Name:           "HTTPS测试节点",
			Server:         "example.com",
			Port:           8443,
			Username:       "user",
			Password:       "pass",
			TLS:            true,
			SkipCertVerify: true,
			SNI:            "example.com",
		}

		encoded := EncodeHTTPURL(original)
		decoded, err := DecodeHTTPURL(encoded)
		if err != nil {
			t.Fatalf("解码失败: %v", err)
		}

		assertEqualString(t, "Name", original.Name, decoded.Name)
		assertEqualString(t, "Server", original.Server, decoded.Server)
		assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
		assertEqualString(t, "Username", original.Username, decoded.Username)
		assertEqualString(t, "Password", original.Password, decoded.Password)
		assertEqualBool(t, "TLS", original.TLS, decoded.TLS)
		assertEqualBool(t, "SkipCertVerify", original.SkipCertVerify, decoded.SkipCertVerify)
		assertEqualString(t, "SNI", original.SNI, decoded.SNI)
		t.Log("✓ HTTPS完整功能测试通过")
	})

	t.Run("HTTP到Clash Proxy转换", func(t *testing.T) {
		httpNode := HTTP{
			Name:     "HTTP到Clash",
			Server:   "example.com",
			Port:     8080,
			Username: "user",
			Password: "pass",
			TLS:      false,
		}

		proxy := ConvertProxyToHTTP(Proxy{
			Name:     httpNode.Name,
			Type:     "http",
			Server:   httpNode.Server,
			Port:     FlexPort(toInt(httpNode.Port)),
			Username: httpNode.Username,
			Password: httpNode.Password,
			Tls:      httpNode.TLS,
		})

		assertEqualString(t, "Name", httpNode.Name, proxy.Name)
		assertEqualString(t, "Server", httpNode.Server, proxy.Server)
		assertEqualIntInterface(t, "Port", httpNode.Port, proxy.Port)
		assertEqualString(t, "Username", httpNode.Username, proxy.Username)
		assertEqualString(t, "Password", httpNode.Password, proxy.Password)
		assertEqualBool(t, "TLS", httpNode.TLS, proxy.TLS)
		t.Log("✓ HTTP到Clash Proxy转换测试通过")
	})

	t.Run("HTTPS到Clash Proxy转换", func(t *testing.T) {
		httpNode := HTTP{
			Name:           "HTTPS到Clash",
			Server:         "example.com",
			Port:           8443,
			Username:       "user",
			Password:       "pass",
			TLS:            true,
			SkipCertVerify: true,
			SNI:            "example.com",
		}

		proxy := ConvertProxyToHTTP(Proxy{
			Name:             httpNode.Name,
			Type:             "http",
			Server:           httpNode.Server,
			Port:             FlexPort(toInt(httpNode.Port)),
			Username:         httpNode.Username,
			Password:         httpNode.Password,
			Tls:              httpNode.TLS,
			Skip_cert_verify: httpNode.SkipCertVerify,
			Sni:              httpNode.SNI,
		})

		assertEqualString(t, "Name", httpNode.Name, proxy.Name)
		assertEqualString(t, "Server", httpNode.Server, proxy.Server)
		assertEqualIntInterface(t, "Port", httpNode.Port, proxy.Port)
		assertEqualString(t, "Username", httpNode.Username, proxy.Username)
		assertEqualString(t, "Password", httpNode.Password, proxy.Password)
		assertEqualBool(t, "TLS", httpNode.TLS, proxy.TLS)
		assertEqualBool(t, "SkipCertVerify", httpNode.SkipCertVerify, proxy.SkipCertVerify)
		assertEqualString(t, "SNI", httpNode.SNI, proxy.SNI)
		t.Log("✓ HTTPS到Clash Proxy转换测试通过")
	})

	t.Run("HTTP协议元数据", func(t *testing.T) {
		InitProtocolMeta()
		meta := GetProtocolMeta("http")
		if meta == nil {
			t.Fatal("HTTP协议元数据未找到")
			return
		}

		assertEqualString(t, "ProtocolName", "http", meta.Name)
		assertEqualString(t, "ProtocolLabel", "HTTP", meta.Label)
		t.Log("✓ HTTP协议元数据测试通过")
	})

	t.Run("HTTPS协议元数据", func(t *testing.T) {
		InitProtocolMeta()
		meta := GetProtocolMeta("https")
		if meta == nil {
			t.Fatal("HTTPS协议元数据未找到")
			return
		}

		assertEqualString(t, "ProtocolName", "https", meta.Name)
		assertEqualString(t, "ProtocolLabel", "HTTPS", meta.Label)
		t.Log("✓ HTTPS协议元数据测试通过")
	})
}
