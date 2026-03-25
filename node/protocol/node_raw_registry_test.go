package protocol

import "testing"

func TestExtractNodeNameFromFields(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		fields   map[string]interface{}
		want     string
	}{
		{name: "vmess uses Ps", protocol: "vmess", fields: map[string]interface{}{"Ps": "VMess 节点"}, want: "VMess 节点"},
		{name: "ssr uses remarks", protocol: "ssr", fields: map[string]interface{}{"Qurey.Remarks": "SSR 节点"}, want: "SSR 节点"},
		{name: "vless uses Name", protocol: "vless", fields: map[string]interface{}{"Name": "VLESS 节点"}, want: "VLESS 节点"},
		{name: "unknown protocol returns empty", protocol: "unknown", fields: map[string]interface{}{"Name": "ignored"}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractNodeNameFromFields(tt.protocol, tt.fields)
			if got != tt.want {
				t.Fatalf("ExtractNodeNameFromFields(%q) = %q, want %q", tt.protocol, got, tt.want)
			}
		})
	}
}

func TestParseNodeLinkDetectsAliasesViaRegistry(t *testing.T) {
	hy2Link := EncodeHY2URL(HY2{
		Password: "test-pass",
		Host:     "example.com",
		Port:     443,
		Name:     "hy2-node",
	})

	parsed, err := ParseNodeLink(hy2Link)
	if err != nil {
		t.Fatalf("ParseNodeLink(hy2) failed: %v", err)
	}
	if parsed.Protocol != "hysteria2" {
		t.Fatalf("ParseNodeLink(hy2) protocol = %q, want %q", parsed.Protocol, "hysteria2")
	}

	wgLink := EncodeWireGuardURL(WireGuard{
		Name:       "wg-node",
		Server:     "1.2.3.4",
		Port:       51820,
		PrivateKey: "private-key",
		PublicKey:  "public-key",
	})
	wgLink = "wg://" + wgLink[len("wireguard://"):]
	parsed, err = ParseNodeLink(wgLink)
	if err != nil {
		t.Fatalf("ParseNodeLink(wg) failed: %v", err)
	}
	if parsed.Protocol != "wireguard" {
		t.Fatalf("ParseNodeLink(wg) protocol = %q, want %q", parsed.Protocol, "wireguard")
	}
}

func TestUpdateNodeLinkFieldsViaRegistry(t *testing.T) {
	link := EncodeHTTPURL(HTTP{
		Name:     "old-name",
		Server:   "example.com",
		Port:     8080,
		Username: "user",
		Password: "pass",
	})

	updated, err := UpdateNodeLinkFields(link, `{"Name":"new-name","Server":"new.example.com","Port":9090}`)
	if err != nil {
		t.Fatalf("UpdateNodeLinkFields failed: %v", err)
	}

	parsed, err := ParseNodeLink(updated)
	if err != nil {
		t.Fatalf("ParseNodeLink(updated) failed: %v", err)
	}

	if got := ExtractNodeNameFromFields(parsed.Protocol, parsed.Fields); got != "new-name" {
		t.Fatalf("updated name = %q, want %q", got, "new-name")
	}
	if got := parsed.Fields["Server"]; got != "new.example.com" {
		t.Fatalf("updated server = %#v, want %q", got, "new.example.com")
	}
	if got := parsed.Fields["Port"]; got != int64(9090) && got != 9090 {
		t.Fatalf("updated port = %#v, want 9090", got)
	}
}

func TestExtractLinkIdentity(t *testing.T) {
	tests := []struct {
		name         string
		link         string
		wantProtocol string
		wantName     string
		wantHost     string
		wantPort     string
	}{
		{
			name:         "vmess identity preserves current host semantics",
			link:         EncodeVmessURL(Vmess{Add: "vmess-server.example", Host: "ws-host.example", Port: 443, Id: "88888888-9999-7777-5555-777777777777", Ps: "vmess-node", V: "2"}),
			wantProtocol: "vmess",
			wantName:     "vmess-node",
			wantHost:     "ws-host.example",
			wantPort:     "443",
		},
		{
			name:         "http identity remains http",
			link:         EncodeHTTPURL(HTTP{Name: "http-node", Server: "proxy.example", Port: 8080}),
			wantProtocol: "http",
			wantName:     "http-node",
			wantHost:     "proxy.example",
			wantPort:     "8080",
		},
		{
			name:         "https identity remains https",
			link:         EncodeHTTPURL(HTTP{Name: "https-node", Server: "secure.example", Port: 443, TLS: true}),
			wantProtocol: "https",
			wantName:     "https-node",
			wantHost:     "secure.example",
			wantPort:     "443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity, err := ExtractLinkIdentity(tt.link)
			if err != nil {
				t.Fatalf("ExtractLinkIdentity failed: %v", err)
			}
			if identity.Protocol != tt.wantProtocol {
				t.Fatalf("protocol = %q, want %q", identity.Protocol, tt.wantProtocol)
			}
			if identity.Name != tt.wantName {
				t.Fatalf("name = %q, want %q", identity.Name, tt.wantName)
			}
			if identity.Host != tt.wantHost {
				t.Fatalf("host = %q, want %q", identity.Host, tt.wantHost)
			}
			if identity.Port != tt.wantPort {
				t.Fatalf("port = %q, want %q", identity.Port, tt.wantPort)
			}
		})
	}
}
