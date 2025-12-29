package protocol

import (
	"testing"
)

// TestGenerateProxyContentHash_Basic 测试基本哈希生成
func TestGenerateProxyContentHash_Basic(t *testing.T) {
	proxy := Proxy{
		Name:   "测试节点",
		Type:   "vmess",
		Server: "example.com",
		Port:   443,
		Uuid:   "12345678-1234-1234-1234-123456789012",
		Cipher: "auto",
	}

	hash := GenerateProxyContentHash(proxy)

	if hash == "" {
		t.Error("哈希值不应为空")
	}
	if len(hash) != 64 {
		t.Errorf("SHA256 哈希应为 64 字符，实际: %d", len(hash))
	}
}

// TestGenerateProxyContentHash_NameIgnored 测试名称不影响哈希
func TestGenerateProxyContentHash_NameIgnored(t *testing.T) {
	proxy1 := Proxy{
		Name:   "节点A",
		Type:   "vmess",
		Server: "example.com",
		Port:   443,
		Uuid:   "12345678-1234-1234-1234-123456789012",
	}

	proxy2 := Proxy{
		Name:   "节点B - 不同名称",
		Type:   "vmess",
		Server: "example.com",
		Port:   443,
		Uuid:   "12345678-1234-1234-1234-123456789012",
	}

	hash1 := GenerateProxyContentHash(proxy1)
	hash2 := GenerateProxyContentHash(proxy2)

	if hash1 != hash2 {
		t.Errorf("名称不同但内容相同的节点哈希应相同\nhash1: %s\nhash2: %s", hash1, hash2)
	}
}

// TestGenerateProxyContentHash_TypeAffectsHash 测试协议类型影响哈希
func TestGenerateProxyContentHash_TypeAffectsHash(t *testing.T) {
	proxy1 := Proxy{
		Type:   "vmess",
		Server: "example.com",
		Port:   443,
	}

	proxy2 := Proxy{
		Type:   "vless",
		Server: "example.com",
		Port:   443,
	}

	hash1 := GenerateProxyContentHash(proxy1)
	hash2 := GenerateProxyContentHash(proxy2)

	if hash1 == hash2 {
		t.Error("不同协议类型的节点哈希应不同")
	}
}

// TestGenerateProxyContentHash_DifferentContent 测试不同内容生成不同哈希
func TestGenerateProxyContentHash_DifferentContent(t *testing.T) {
	proxy1 := Proxy{
		Type:   "vmess",
		Server: "server1.com",
		Port:   443,
		Uuid:   "uuid-1",
	}

	proxy2 := Proxy{
		Type:   "vmess",
		Server: "server2.com",
		Port:   443,
		Uuid:   "uuid-2",
	}

	hash1 := GenerateProxyContentHash(proxy1)
	hash2 := GenerateProxyContentHash(proxy2)

	if hash1 == hash2 {
		t.Error("不同内容的节点哈希应不同")
	}
}

// TestGenerateProxyContentHash_WsOptsOrder 测试 map 字段顺序不影响哈希
func TestGenerateProxyContentHash_WsOptsOrder(t *testing.T) {
	// 虽然 Go map 本身无序，但通过规范化处理应该保证一致性
	proxy1 := Proxy{
		Type:   "vmess",
		Server: "example.com",
		Port:   443,
		Ws_opts: map[string]interface{}{
			"path": "/ws",
			"headers": map[string]interface{}{
				"Host": "example.com",
			},
		},
	}

	proxy2 := Proxy{
		Type:   "vmess",
		Server: "example.com",
		Port:   443,
		Ws_opts: map[string]interface{}{
			"headers": map[string]interface{}{
				"Host": "example.com",
			},
			"path": "/ws",
		},
	}

	hash1 := GenerateProxyContentHash(proxy1)
	hash2 := GenerateProxyContentHash(proxy2)

	if hash1 != hash2 {
		t.Errorf("相同 ws_opts 内容但不同字段顺序的哈希应相同\nhash1: %s\nhash2: %s", hash1, hash2)
	}
}

// TestGenerateProxyContentHash_AlpnOrder 测试 Alpn 切片顺序规范化
func TestGenerateProxyContentHash_AlpnOrder(t *testing.T) {
	proxy1 := Proxy{
		Type:   "trojan",
		Server: "example.com",
		Port:   443,
		Alpn:   []string{"h2", "http/1.1"},
	}

	proxy2 := Proxy{
		Type:   "trojan",
		Server: "example.com",
		Port:   443,
		Alpn:   []string{"http/1.1", "h2"},
	}

	hash1 := GenerateProxyContentHash(proxy1)
	hash2 := GenerateProxyContentHash(proxy2)

	if hash1 != hash2 {
		t.Errorf("相同 Alpn 内容但不同顺序的哈希应相同\nhash1: %s\nhash2: %s", hash1, hash2)
	}
}

// TestGenerateProxyContentHash_DialerProxyIgnored 测试前置代理不影响哈希
func TestGenerateProxyContentHash_DialerProxyIgnored(t *testing.T) {
	proxy1 := Proxy{
		Type:         "vmess",
		Server:       "example.com",
		Port:         443,
		Dialer_proxy: "",
	}

	proxy2 := Proxy{
		Type:         "vmess",
		Server:       "example.com",
		Port:         443,
		Dialer_proxy: "entry-proxy",
	}

	hash1 := GenerateProxyContentHash(proxy1)
	hash2 := GenerateProxyContentHash(proxy2)

	if hash1 != hash2 {
		t.Errorf("前置代理不同的节点哈希应相同\nhash1: %s\nhash2: %s", hash1, hash2)
	}
}

// TestGenerateProxyContentHash_AllProtocols 测试所有协议类型
func TestGenerateProxyContentHash_AllProtocols(t *testing.T) {
	protocols := []string{"ss", "ssr", "vmess", "vless", "trojan", "hysteria", "hysteria2", "tuic", "wireguard", "anytls", "socks5"}

	hashes := make(map[string]string)
	for _, proto := range protocols {
		proxy := Proxy{
			Type:   proto,
			Server: "example.com",
			Port:   443,
		}
		hash := GenerateProxyContentHash(proxy)

		if hash == "" {
			t.Errorf("协议 %s 的哈希值不应为空", proto)
			continue
		}

		// 检查是否与其他协议哈希冲突
		for otherProto, otherHash := range hashes {
			if hash == otherHash {
				t.Errorf("协议 %s 和 %s 的哈希值不应相同", proto, otherProto)
			}
		}
		hashes[proto] = hash
	}
}

// TestGetHashIgnoredFields 测试获取忽略字段列表
func TestGetHashIgnoredFields(t *testing.T) {
	fields := GetHashIgnoredFields()

	if len(fields) == 0 {
		t.Error("忽略字段列表不应为空")
	}

	// 检查必要字段是否在列表中
	requiredIgnored := []string{"Name", "Dialer_proxy", "Udp", "Tfo"}
	fieldMap := make(map[string]bool)
	for _, f := range fields {
		fieldMap[f] = true
	}

	for _, required := range requiredIgnored {
		if !fieldMap[required] {
			t.Errorf("字段 %s 应在忽略列表中", required)
		}
	}
}

// TestNormalizeProxyForHash 测试规范化数据输出
func TestNormalizeProxyForHash(t *testing.T) {
	proxy := Proxy{
		Name:   "测试节点",
		Type:   "vmess",
		Server: "example.com",
		Port:   443,
		Uuid:   "test-uuid",
	}

	data := NormalizeProxyForHash(proxy)

	// Name 应被忽略
	if _, exists := data["Name"]; exists {
		t.Error("Name 字段应被忽略")
	}

	// Type 应存在
	if _, exists := data["Type"]; !exists {
		t.Error("Type 字段应存在")
	}

	// Server 应存在
	if _, exists := data["Server"]; !exists {
		t.Error("Server 字段应存在")
	}
}
