package protocol

import (
	"testing"
	"sublink/utils"
)

func TestProtocolDemoEncodeDecodeRoundTrip(t *testing.T) {
	original := ProtocolDemo{
		Name:     "demo-node",
		Server:   "demo.example.com",
		Port:     443,
		Token:    "demo-token",
		Mode:     "ws",
		TLS:      true,
		Insecure: true,
	}

	encoded := EncodeProtocolDemoURL(original)
	decoded, err := DecodeProtocolDemoURL(encoded)
	if err != nil {
		t.Fatalf("DecodeProtocolDemoURL failed: %v", err)
	}

	if decoded.Name != original.Name {
		t.Fatalf("name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Server != original.Server {
		t.Fatalf("server = %q, want %q", decoded.Server, original.Server)
	}
	if decoded.Token != original.Token {
		t.Fatalf("token = %q, want %q", decoded.Token, original.Token)
	}
	if decoded.Mode != original.Mode {
		t.Fatalf("mode = %q, want %q", decoded.Mode, original.Mode)
	}
	if !decoded.TLS {
		t.Fatalf("tls = false, want true")
	}
	if !decoded.Insecure {
		t.Fatalf("insecure = false, want true")
	}
	if got := utils.GetPortInt(decoded.Port); got != 443 {
		t.Fatalf("port = %d, want 443", got)
	}
}

func TestProtocolDemoRegistryMetadata(t *testing.T) {
	meta := GetProtocolMeta("protocoldemo")
	if meta == nil {
		t.Fatal("GetProtocolMeta(protocoldemo) returned nil")
	}
	if meta.Label != "ProtocolDemo" {
		t.Fatalf("label = %q, want %q", meta.Label, "ProtocolDemo")
	}
	fieldNames := map[string]bool{}
	for _, field := range meta.Fields {
		fieldNames[field.Name] = true
	}
	for _, name := range []string{"Name", "Server", "Port", "Token", "Mode", "TLS", "Insecure"} {
		if !fieldNames[name] {
			t.Fatalf("expected field %q in protocol meta", name)
		}
	}

	link := EncodeProtocolDemoURL(ProtocolDemo{Server: "demo.example.com", Port: 8443, Token: "x", Mode: "grpc"})
	if got := GetProtocolFromLink(link); got != "protocoldemo" {
		t.Fatalf("protocol = %q, want %q", got, "protocoldemo")
	}
}

func TestProtocolDemoProxyRoundTrip(t *testing.T) {
	link := EncodeProtocolDemoURL(ProtocolDemo{
		Name:     "demo-node",
		Server:   "demo.example.com",
		Port:     9443,
		Token:    "demo-secret",
		Mode:     "grpc",
		TLS:      true,
		Insecure: true,
	})

	proxy, err := LinkToProxy(Urls{Url: link, DialerProxyName: "proxy-a"}, OutputConfig{Udp: true, Cert: false})
	if err != nil {
		t.Fatalf("LinkToProxy failed: %v", err)
	}
	if proxy.Type != "protocoldemo" {
		t.Fatalf("proxy type = %q, want %q", proxy.Type, "protocoldemo")
	}
	if proxy.Password != "demo-secret" {
		t.Fatalf("proxy password = %q, want %q", proxy.Password, "demo-secret")
	}
	if proxy.Network != "grpc" {
		t.Fatalf("proxy network = %q, want %q", proxy.Network, "grpc")
	}

	reEncoded, err := EncodeProxyLink(proxy)
	if err != nil {
		t.Fatalf("EncodeProxyLink failed: %v", err)
	}
	decoded, err := DecodeProtocolDemoURL(reEncoded)
	if err != nil {
		t.Fatalf("DecodeProtocolDemoURL(reEncoded) failed: %v", err)
	}
	if decoded.Name != "demo-node" {
		t.Fatalf("decoded name = %q, want %q", decoded.Name, "demo-node")
	}
	if decoded.Token != "demo-secret" {
		t.Fatalf("decoded token = %q, want %q", decoded.Token, "demo-secret")
	}
}
