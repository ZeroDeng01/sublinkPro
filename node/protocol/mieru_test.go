package protocol

import (
	"encoding/base64"
	"net/url"
	"strings"
	"testing"

	"sublink/cache"

	pb "github.com/enfein/mieru/v3/pkg/appctl/appctlpb"
	"github.com/metacubex/mihomo/adapter"
	"google.golang.org/protobuf/proto"
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
		TrafficPattern: "CCo=",
	}

	encoded := EncodeMieruURL(original)
	u, err := url.Parse(encoded)
	if err != nil {
		t.Fatalf("parse encoded link: %v", err)
	}
	if u.Scheme != "mierus" || u.Port() != "" || u.Query().Get("port") != "2999" ||
		u.Query().Get("protocol") != "TCP" || u.Query().Get("profile") != "mieru-node" {
		t.Fatalf("invalid simple share format: %s", encoded)
	}
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
		TrafficPattern: "CCo=",
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
	assertEqualString(t, "TrafficPattern", "CCo=", proxy.TrafficPattern)
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
		"traffic-pattern: CCo=",
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
    traffic-pattern: CCo=
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
	assertEqualString(t, "TrafficPattern", "CCo=", decoded.TrafficPattern)
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

	if got := GetProtocolFromLink("mierus://user:password@official.example.com?port=2999&profile=raw&protocol=TCP"); got != "mieru" {
		t.Fatalf("simple share link protocol = %q", got)
	}
}

func TestDefaultClientSupportForSurgeCapableProtocol(t *testing.T) {
	for _, client := range []string{ClientClash, ClientMihomo, ClientV2ray, ClientSurge} {
		if !ProtocolSupportsClient("ss", client) {
			t.Fatalf("ss should support %s", client)
		}
	}
}

func TestMieruOfficialSimpleLink(t *testing.T) {
	link := "mierus://test-user:p%40ss@[2001:db8::1]?handshake-mode=HANDSHAKE_NO_WAIT&mtu=1400&multiplexing=MULTIPLEXING_OFF&port=11211&port=2000-2003&profile=default&protocol=TCP&protocol=UDP"
	proxy, err := LinkToProxy(Urls{Url: link}, OutputConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if proxy.Server != "2001:db8::1" || proxy.Password != "p@ss" || proxy.HandshakeMode != "HANDSHAKE_NO_WAIT" ||
		proxy.MieruMTU != 1400 || proxy.Multiplexing != "MULTIPLEXING_OFF" || proxy.Port != 11211 || proxy.Transport != "TCP" {
		t.Fatal("simple link did not map to its first binding")
	}
	updated, err := UpdateNodeLinkFields(link, `{"Port":12345}`)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(updated)
	if err != nil || u.Scheme != "mierus" ||
		strings.Join(u.Query()["port"], ",") != "12345,2000-2003" || strings.Join(u.Query()["protocol"], ",") != "TCP,UDP" {
		t.Fatalf("raw edit lost an original port/protocol pair, err = %v", err)
	}
}

func TestMieruClashPortRangeCompatibility(t *testing.T) {
	const pattern = "CCoQARoECAEQCiIYCAMQASoIMDAwMTAyMDMqCDA0MDUwNjA3"
	const base = "mierus://test-user:test-password@1.2.3.4?handshake-mode=HANDSHAKE_NO_WAIT&mtu=1400&multiplexing=MULTIPLEXING_HIGH&traffic-pattern=" + pattern
	template := t.Name() + ".yaml"
	cache.SetTemplateContent(template, `proxies: []
proxy-groups:
  - name: Select
    type: select
    proxies: [__ALL_PROXIES__]
  - name: Fixed
    type: select
    proxies: [mixed]
`)
	mixed := base + "&profile=mixed&port=6666&port=9998-9999&port=6489&port=4896&protocol=TCP&protocol=TCP&protocol=UDP&protocol=UDP"
	// The airport link-to-proxy-to-link round trip preserves the unsupported bindings.
	proxy, err := LinkToProxy(Urls{Url: mixed}, OutputConfig{})
	if err != nil {
		t.Fatal(err)
	}
	stored, err := EncodeProxyLink(proxy)
	if err != nil {
		t.Fatal(err)
	}
	data, err := EncodeClash([]Urls{
		{Url: stored},
		{Url: base + "&profile=gap&port=6666&port=6690&protocol=TCP&protocol=TCP"},
		{Url: base + "&profile=single&port=6666&protocol=TCP"},
		{Url: base + "&profile=range&port=9998-9999&protocol=UDP"},
		{Url: base + "&profile=adjacent&port=2092-2094&port=2090&port=2091-2093&protocol=TCP&protocol=TCP&protocol=TCP"},
	}, OutputConfig{Clash: template})
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Proxies []map[string]any `yaml:"proxies"`
		Groups  []ProxyGroup     `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	expected := []struct {
		name      string
		port      int
		portRange string
		transport string
	}{
		{name: "single", port: 6666, transport: "TCP"},
		{name: "range", portRange: "9998-9999", transport: "UDP"},
		{name: "adjacent", portRange: "2090-2094", transport: "TCP"},
	}
	if len(config.Proxies) != len(expected) || len(config.Groups) != 2 ||
		strings.Join(config.Groups[0].Proxies, ",") != "single,range,adjacent" || strings.Join(config.Groups[1].Proxies, ",") != "REJECT" {
		t.Fatalf("unexpected outbounds or proxy group references")
	}
	for i, want := range expected {
		entry := config.Proxies[i]
		port, hasPort := entry["port"].(int)
		portRange, hasRange := entry["port-range"].(string)
		if entry["name"] != want.name || hasPort == hasRange || port != want.port || portRange != want.portRange || entry["transport"] != want.transport {
			t.Fatalf("outbound %d has the wrong port, transport, or name", i)
		}
		if entry["handshake-mode"] != "HANDSHAKE_NO_WAIT" || entry["multiplexing"] != "MULTIPLEXING_HIGH" || entry["traffic-pattern"] != pattern {
			t.Fatalf("outbound %d lost shared settings", i)
		}
		parsed, err := adapter.ParseProxy(entry)
		if err != nil {
			t.Fatalf("Mihomo rejected outbound %d: %v", i, err)
		}
		if err := parsed.Close(); err != nil {
			t.Fatalf("close outbound %d: %v", i, err)
		}
	}
}

func TestMieruOfficialStandardPreservesClientSettings(t *testing.T) {
	profile := &pb.ClientProfile{
		ProfileName: proto.String("active"), User: &pb.User{Name: proto.String("user"), Password: proto.String("pass")},
		Mtu: proto.Int32(1400), HandshakeMode: pb.HandshakeMode_HANDSHAKE_NO_WAIT.Enum(),
		Dialer: &pb.ClientDialer{Host: proto.String("dialer.example.com"), Port: proto.Int32(1080)},
		Servers: []*pb.ServerEndpoint{
			{
				DomainName: proto.String("example.com"),
				PortBindings: []*pb.PortBinding{
					{Port: proto.Int32(11211), Protocol: pb.TransportProtocol_TCP.Enum()},
					{Port: proto.Int32(22322), Protocol: pb.TransportProtocol_UDP.Enum()},
				},
			},
			{DomainName: proto.String("backup.example.com")},
		},
	}
	config := &pb.ClientConfig{
		Profiles:      []*pb.ClientProfile{{ProfileName: proto.String("inactive")}, profile},
		ActiveProfile: proto.String("active"), Socks5Port: proto.Int32(1081),
	}
	link := standardMieruFixture(t, config)
	node, err := DecodeMieruURL(link)
	if err != nil || node.Server != "example.com" || node.portValue() != "11211" || node.HandshakeMode != "HANDSHAKE_NO_WAIT" {
		t.Fatalf("standard link did not use the active profile's first binding for the editor, err = %v", err)
	}
	if _, err := LinkToProxy(Urls{Url: link}, OutputConfig{}); err == nil {
		t.Fatal("standard link with multiple servers was converted to one Clash proxy")
	}
	updated, err := UpdateNodeLinkFields(link, `{"Name":"new-name","Port":12345}`)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeMieruURL(updated)
	if err != nil {
		t.Fatal(err)
	}
	profile.ProfileName = proto.String("new-name")
	profile.Servers[0].PortBindings[0].Port = proto.Int32(12345)
	config.ActiveProfile = proto.String("new-name")
	if !proto.Equal(config, decoded.standardConfig) {
		t.Fatal("editing the standard link changed unrelated client settings")
	}
}

func standardMieruFixture(t *testing.T, config *pb.ClientConfig) string {
	t.Helper()
	data, err := proto.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	return "mieru://" + base64.RawStdEncoding.EncodeToString(data)
}

func TestMieruRejectsLegacyAndInvalidOfficialLinks(t *testing.T) {
	for _, link := range []string{
		"mieru://user:pass@example.com:11211?transport=TCP",
		"mieru://YQ==",
		"mierus://user:pass@example.com?port=11211&profile=default",
		"mierus://user:pass@example.com?port=11211&profile=default&profile=other&protocol=TCP",
		"mierus://user:pass@example.com?port=11211&profile=default&protocol=OTHER",
		"mierus://user:pass@example.com?port=3002-3000&profile=default&protocol=UDP",
	} {
		if _, err := DecodeMieruURL(link); err == nil {
			t.Fatalf("invalid link accepted: %q", link)
		}
	}
}
