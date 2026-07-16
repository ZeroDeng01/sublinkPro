package protocol

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSnellEncodeDecodeRoundTrip(t *testing.T) {
	original := Snell{
		Name:          "snell-node",
		Server:        "example.com",
		Port:          44046,
		Psk:           "this-is-a-psk",
		Version:       3,
		UDP:           true,
		ObfsMode:      "http",
		ObfsHost:      "bing.com",
		Tfo:           true,
		Mptcp:         true,
		InterfaceName: "eth0",
		RoutingMark:   255,
		IPVersion:     "ipv4-prefer",
	}

	encoded := EncodeSnellURL(original)
	if !strings.HasPrefix(encoded, "snell://") {
		t.Fatalf("编码后应以 snell:// 开头, 实际: %s", encoded)
	}

	decoded, err := DecodeSnellURL(encoded)
	if err != nil {
		t.Fatalf("DecodeSnellURL failed: %v", err)
	}

	assertEqualString(t, "Name", original.Name, decoded.Name)
	assertEqualString(t, "Server", original.Server, decoded.Server)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Psk", original.Psk, decoded.Psk)
	assertEqualInt(t, "Version", original.Version, decoded.Version)
	assertEqualBool(t, "UDP", original.UDP, decoded.UDP)
	assertEqualString(t, "ObfsMode", original.ObfsMode, decoded.ObfsMode)
	assertEqualString(t, "ObfsHost", original.ObfsHost, decoded.ObfsHost)
	assertEqualBool(t, "Tfo", original.Tfo, decoded.Tfo)
	assertEqualBool(t, "Mptcp", original.Mptcp, decoded.Mptcp)
	assertEqualString(t, "InterfaceName", original.InterfaceName, decoded.InterfaceName)
	assertEqualInt(t, "RoutingMark", original.RoutingMark, decoded.RoutingMark)
	assertEqualString(t, "IPVersion", original.IPVersion, decoded.IPVersion)
}

func TestSnellDecodePskFromUserinfo(t *testing.T) {
	decoded, err := DecodeSnellURL("snell://my-psk@example.com:44046#snell-node")
	if err != nil {
		t.Fatalf("DecodeSnellURL failed: %v", err)
	}
	assertEqualString(t, "Psk", "my-psk", decoded.Psk)
	assertEqualString(t, "Server", "example.com", decoded.Server)
	assertEqualIntInterface(t, "Port", 44046, decoded.Port)
}

func TestSnellLinkToProxyAndYAMLOutput(t *testing.T) {
	link := EncodeSnellURL(Snell{
		Name:     "snell-yaml",
		Server:   "snell.example.com",
		Port:     44046,
		Psk:      "psk-value",
		Version:  3,
		UDP:      true,
		ObfsMode: "tls",
		ObfsHost: "cloudflare.com",
	})

	proxy, err := LinkToProxy(Urls{Url: link, DialerProxyName: "front-proxy"}, OutputConfig{})
	if err != nil {
		t.Fatalf("LinkToProxy failed: %v", err)
	}
	assertEqualString(t, "Type", "snell", proxy.Type)
	assertEqualString(t, "Server", "snell.example.com", proxy.Server)
	assertEqualFlexPort(t, "Port", 44046, proxy.Port)
	assertEqualString(t, "Psk", "psk-value", proxy.Psk)
	assertEqualInt(t, "Version", 3, proxy.Version)
	assertEqualBool(t, "UDP", true, proxy.Udp)
	assertEqualString(t, "DialerProxy", "front-proxy", proxy.Dialer_proxy)

	mode, ok := proxy.Obfs_opts["mode"].(string)
	if !ok || mode != "tls" {
		t.Fatalf("obfs-opts mode = %#v, want tls", proxy.Obfs_opts["mode"])
	}
	host, ok := proxy.Obfs_opts["host"].(string)
	if !ok || host != "cloudflare.com" {
		t.Fatalf("obfs-opts host = %#v, want cloudflare.com", proxy.Obfs_opts["host"])
	}

	data, err := yaml.Marshal(proxy)
	if err != nil {
		t.Fatalf("yaml marshal failed: %v", err)
	}
	encodedYAML := string(data)
	for _, want := range []string{
		"type: snell",
		"server: snell.example.com",
		"port: 44046",
		"psk: psk-value",
		"version: 3",
		"udp: true",
		"obfs-opts:",
		"mode: tls",
		"host: cloudflare.com",
		"dialer-proxy: front-proxy",
	} {
		if !strings.Contains(encodedYAML, want) {
			t.Fatalf("YAML output missing %q: %s", want, encodedYAML)
		}
	}
}

func TestSnellOutputConfigForcesUDP(t *testing.T) {
	link := EncodeSnellURL(Snell{
		Name:    "snell-config",
		Server:  "example.com",
		Port:    44046,
		Psk:     "psk",
		Version: 3,
	})

	proxy, err := buildSnellProxy(Urls{Url: link}, OutputConfig{Udp: true})
	if err != nil {
		t.Fatalf("buildSnellProxy failed: %v", err)
	}
	assertEqualBool(t, "ForcedUDP", true, proxy.Udp)
}

func TestSnellEncodeProxyLinkFromClashYAML(t *testing.T) {
	var config struct {
		Proxies []Proxy `yaml:"proxies"`
	}
	err := yaml.Unmarshal([]byte(`proxies:
  - name: imported-snell
    type: snell
    server: snell.example.com
    port: 44046
    psk: imported-psk
    version: 3
    udp: true
    tfo: true
    mptcp: true
    interface-name: eth0
    routing-mark: 255
    ip-version: ipv4-prefer
    obfs-opts:
      mode: http
      host: bing.com
`), &config)
	if err != nil {
		t.Fatalf("yaml unmarshal failed: %v", err)
	}
	if len(config.Proxies) != 1 {
		t.Fatalf("proxy count = %d, want 1", len(config.Proxies))
	}

	link, err := EncodeProxyLink(config.Proxies[0])
	if err != nil {
		t.Fatalf("EncodeProxyLink failed: %v", err)
	}
	decoded, err := DecodeSnellURL(link)
	if err != nil {
		t.Fatalf("DecodeSnellURL failed: %v", err)
	}
	assertEqualString(t, "Name", "imported-snell", decoded.Name)
	assertEqualString(t, "Server", "snell.example.com", decoded.Server)
	assertEqualIntInterface(t, "Port", 44046, decoded.Port)
	assertEqualString(t, "Psk", "imported-psk", decoded.Psk)
	assertEqualInt(t, "Version", 3, decoded.Version)
	assertEqualBool(t, "UDP", true, decoded.UDP)
	assertEqualString(t, "ObfsMode", "http", decoded.ObfsMode)
	assertEqualString(t, "ObfsHost", "bing.com", decoded.ObfsHost)
	assertEqualBool(t, "Tfo", true, decoded.Tfo)
	assertEqualBool(t, "Mptcp", true, decoded.Mptcp)
	assertEqualString(t, "InterfaceName", "eth0", decoded.InterfaceName)
	assertEqualInt(t, "RoutingMark", 255, decoded.RoutingMark)
	assertEqualString(t, "IPVersion", "ipv4-prefer", decoded.IPVersion)

	// 反向验证：Clash YAML 输出仍保留这些 BasicOption 字段
	rebuilt, err := buildSnellProxy(Urls{Url: link}, OutputConfig{})
	if err != nil {
		t.Fatalf("buildSnellProxy failed: %v", err)
	}
	data, err := yaml.Marshal(rebuilt)
	if err != nil {
		t.Fatalf("yaml marshal failed: %v", err)
	}
	for _, want := range []string{
		"tfo: true",
		"mptcp: true",
		"interface-name: eth0",
		"routing-mark: 255",
		"ip-version: ipv4-prefer",
	} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("Clash YAML output missing %q: %s", want, string(data))
		}
	}
}

func TestSnellSurgeLineIncludesSupportedFields(t *testing.T) {
	link := EncodeSnellURL(Snell{
		Name:     "snell-surge",
		Server:   "example.com",
		Port:     44046,
		Psk:      "psk-value",
		Version:  3,
		ObfsMode: "http",
		ObfsHost: "bing.com",
	})

	line, name, err := buildSnellSurgeLine(link, OutputConfig{})
	if err != nil {
		t.Fatalf("buildSnellSurgeLine failed: %v", err)
	}

	assertEqualString(t, "SurgeName", "snell-surge", name)
	assertContains(t, "SurgeType", line, "snell-surge = snell, example.com, 44046")
	assertContains(t, "SurgePsk", line, "psk=psk-value")
	assertContains(t, "SurgeVersion", line, "version=3")
	assertContains(t, "SurgeObfs", line, "obfs=http")
	assertContains(t, "SurgeObfsHost", line, "obfs-host=bing.com")
}

func TestSnellRegistryMetadata(t *testing.T) {
	meta := GetProtocolMeta("snell")
	if meta == nil {
		t.Fatal("GetProtocolMeta(snell) returned nil")
		return
	}
	assertEqualString(t, "Label", "Snell", meta.Label)

	fieldNames := map[string]bool{}
	for _, field := range meta.Fields {
		fieldNames[field.Name] = true
	}
	for _, name := range []string{"Name", "Server", "Port", "Psk", "Version", "UDP", "ObfsMode", "ObfsHost", "Tfo", "Mptcp", "InterfaceName", "RoutingMark", "IPVersion"} {
		if !fieldNames[name] {
			t.Fatalf("expected field %q in Snell protocol meta", name)
		}
	}

	link := EncodeSnellURL(Snell{Server: "example.com", Port: 44046, Psk: "psk"})
	if got := GetProtocolFromLink(link); got != "snell" {
		t.Fatalf("protocol = %q, want snell", got)
	}
}

func TestSnellClientSupport(t *testing.T) {
	for _, client := range []string{ClientClash, ClientMihomo, ClientSurge} {
		if !ProtocolSupportsClient("snell", client) {
			t.Fatalf("snell should support %s", client)
		}
	}

	if ProtocolSupportsClient("snell", ClientV2ray) {
		t.Fatal("snell should not support v2ray")
	}

	link := EncodeSnellURL(Snell{Server: "example.com", Port: 44046, Psk: "psk"})
	if SupportsClientForLink(link, ClientV2ray) {
		t.Fatal("snell link should not support v2ray")
	}
}

func TestSnellNameModification(t *testing.T) {
	original := Snell{
		Name:    "原始名称",
		Server:  "example.com",
		Port:    44046,
		Psk:     "psk",
		Version: 3,
	}

	newName := "新名称-Snell-测试"
	encoded := EncodeSnellURL(original)
	decoded, _ := DecodeSnellURL(encoded)
	decoded.Name = newName
	reEncoded := EncodeSnellURL(decoded)
	final, _ := DecodeSnellURL(reEncoded)

	assertEqualString(t, "修改后名称", newName, final.Name)
	assertEqualString(t, "服务器(不变)", original.Server, final.Server)
	assertEqualString(t, "PSK(不变)", original.Psk, final.Psk)
	assertEqualInt(t, "版本(不变)", original.Version, final.Version)
}
