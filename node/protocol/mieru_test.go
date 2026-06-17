package protocol

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMieruEncodeDecodeRoundTrip(t *testing.T) {
	original := Mieru{
		Name:           "mieru-node",
		Server:         "example.com",
		Port:           2999,
		Transport:      "TCP",
		Username:       "user",
		Password:       "pass:word",
		Multiplexing:   "MULTIPLEXING_LOW",
		TrafficPattern: "dGVzdA==",
	}

	encoded := EncodeMieruURL(original)
	decoded, err := DecodeMieruURL(encoded)
	if err != nil {
		t.Fatalf("DecodeMieruURL failed: %v", err)
	}

	assertEqualString(t, "Name", original.Name, decoded.Name)
	assertEqualString(t, "Server", original.Server, decoded.Server)
	assertEqualIntInterface(t, "Port", original.Port, decoded.Port)
	assertEqualString(t, "Transport", original.Transport, decoded.Transport)
	assertEqualString(t, "Username", original.Username, decoded.Username)
	assertEqualString(t, "Password", original.Password, decoded.Password)
	assertEqualString(t, "Multiplexing", original.Multiplexing, decoded.Multiplexing)
	assertEqualString(t, "TrafficPattern", original.TrafficPattern, decoded.TrafficPattern)

	withRange := Mieru{
		Name:      "mieru-range",
		Server:    "range.example.com",
		PortRange: "2090-2099",
		Transport: "UDP",
		Username:  "range-user",
		Password:  "range-pass",
	}
	decodedRange, err := DecodeMieruURL(EncodeMieruURL(withRange))
	if err != nil {
		t.Fatalf("DecodeMieruURL(port-range) failed: %v", err)
	}
	assertEqualString(t, "PortRange", withRange.PortRange, decodedRange.PortRange)
	assertEqualIntInterface(t, "Port", 0, decodedRange.Port)
}

func TestMieruLinkToProxyAndYAMLOutput(t *testing.T) {
	link := EncodeMieruURL(Mieru{
		Name:           "mieru-yaml",
		Server:         "mieru.example.com",
		PortRange:      "2090-2099",
		Transport:      "TCP",
		Username:       "user",
		Password:       "password",
		Multiplexing:   "MULTIPLEXING_MIDDLE",
		TrafficPattern: "cGF0dGVybg==",
	})

	proxy, err := LinkToProxy(Urls{Url: link, DialerProxyName: "front-proxy"}, OutputConfig{})
	if err != nil {
		t.Fatalf("LinkToProxy failed: %v", err)
	}
	assertEqualString(t, "Type", "mieru", proxy.Type)
	assertEqualString(t, "Server", "mieru.example.com", proxy.Server)
	assertEqualString(t, "PortRange", "2090-2099", proxy.PortRange)
	assertEqualString(t, "Transport", "TCP", proxy.Transport)
	assertEqualString(t, "Username", "user", proxy.Username)
	assertEqualString(t, "Password", "password", proxy.Password)
	assertEqualString(t, "Multiplexing", "MULTIPLEXING_MIDDLE", proxy.Multiplexing)
	assertEqualString(t, "TrafficPattern", "cGF0dGVybg==", proxy.TrafficPattern)
	assertEqualString(t, "DialerProxy", "front-proxy", proxy.Dialer_proxy)

	data, err := yaml.Marshal(proxy)
	if err != nil {
		t.Fatalf("yaml marshal failed: %v", err)
	}
	encodedYAML := string(data)
	for _, want := range []string{
		"type: mieru",
		"port-range: 2090-2099",
		"transport: TCP",
		"username: user",
		"password: password",
		"multiplexing: MULTIPLEXING_MIDDLE",
		"traffic-pattern: cGF0dGVybg==",
		"dialer-proxy: front-proxy",
	} {
		if !strings.Contains(encodedYAML, want) {
			t.Fatalf("YAML output missing %q: %s", want, encodedYAML)
		}
	}
	if strings.Contains(encodedYAML, "port: 0") {
		t.Fatalf("port-range output must not emit an empty port: %s", encodedYAML)
	}
}

func TestMieruEncodeProxyLinkFromClashYAML(t *testing.T) {
	var config struct {
		Proxies []Proxy `yaml:"proxies"`
	}
	err := yaml.Unmarshal([]byte(`proxies:
  - name: imported-mieru
    type: mieru
    server: mieru.example.com
    port: 2999
    transport: UDP
    username: user
    password: password
    multiplexing: MULTIPLEXING_HIGH
    traffic-pattern: dHJhZmZpYw==
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
	decoded, err := DecodeMieruURL(link)
	if err != nil {
		t.Fatalf("DecodeMieruURL failed: %v", err)
	}
	assertEqualString(t, "Name", "imported-mieru", decoded.Name)
	assertEqualString(t, "Server", "mieru.example.com", decoded.Server)
	assertEqualIntInterface(t, "Port", 2999, decoded.Port)
	assertEqualString(t, "Transport", "UDP", decoded.Transport)
	assertEqualString(t, "Multiplexing", "MULTIPLEXING_HIGH", decoded.Multiplexing)
	assertEqualString(t, "TrafficPattern", "dHJhZmZpYw==", decoded.TrafficPattern)
}

func TestMieruRegistryMetadata(t *testing.T) {
	meta := GetProtocolMeta("mieru")
	if meta == nil {
		t.Fatal("GetProtocolMeta(mieru) returned nil")
		return
	}
	assertEqualString(t, "Label", "Mieru", meta.Label)

	fieldNames := map[string]bool{}
	for _, field := range meta.Fields {
		fieldNames[field.Name] = true
	}
	for _, name := range []string{"Name", "Server", "Port", "PortRange", "Transport", "Username", "Password", "Multiplexing", "TrafficPattern"} {
		if !fieldNames[name] {
			t.Fatalf("expected field %q in Mieru protocol meta", name)
		}
	}

	link := EncodeMieruURL(Mieru{Server: "example.com", Port: 2999, Username: "user", Password: "pass"})
	if got := GetProtocolFromLink(link); got != "mieru" {
		t.Fatalf("protocol = %q, want mieru", got)
	}
}

func TestMieruClientSupport(t *testing.T) {
	for _, client := range []string{ClientClash, ClientMihomo} {
		if !ProtocolSupportsClient("mieru", client) {
			t.Fatalf("mieru should support %s", client)
		}
	}

	for _, client := range []string{ClientV2ray, ClientSurge} {
		if ProtocolSupportsClient("mieru", client) {
			t.Fatalf("mieru should not support %s", client)
		}
	}

	for _, link := range []string{
		"mieru://user:password@mieru.example.com:2999?transport=TCP#m",
		"mierus://official.example.com:2999?profile=raw#m",
	} {
		if SupportsClientForLink(link, ClientV2ray) {
			t.Fatalf("%s should not support v2ray", link)
		}
	}

	if GetProtocolFromLink("mierus://official.example.com:2999?profile=raw#m") != "other" {
		t.Fatal("mierus:// must remain a support-detection alias, not a full Mieru decode alias")
	}
	if _, err := DecodeMieruURL("mierus://official.example.com:2999?profile=raw#m"); err == nil {
		t.Fatal("DecodeMieruURL should reject mierus:// links until official share-link parsing is implemented")
	}
}

func TestDefaultClientSupportForSurgeCapableProtocol(t *testing.T) {
	for _, client := range []string{ClientClash, ClientMihomo, ClientV2ray, ClientSurge} {
		if !ProtocolSupportsClient("ss", client) {
			t.Fatalf("ss should support %s", client)
		}
	}
}
