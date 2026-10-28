package protocol

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestOpenVPNEncodeDecodeRoundTrip(t *testing.T) {
	tranWindow := 0
	original := OpenVPN{
		Name:               "openvpn-node",
		Server:             "vpn.example.com",
		Port:               1242,
		Proto:              "tcp",
		Dev:                "tun",
		Cipher:             "AES-128-CBC",
		DataCiphers:        []string{"AES-128-CBC", "AES-256-GCM"},
		DataCipherFallback: "AES-128-CBC",
		Auth:               "SHA1",
		CompLZO:            "no",
		CA:                 "-----BEGIN CERTIFICATE-----\nCA\n-----END CERTIFICATE-----\n",
		Cert:               "-----BEGIN CERTIFICATE-----\nCERT\n-----END CERTIFICATE-----\n",
		Key:                "-----BEGIN PRIVATE KEY-----\nKEY\n-----END PRIVATE KEY-----\n",
		TLSAuth:            "-----BEGIN OpenVPN Static key V1-----\nAUTH\n-----END OpenVPN Static key V1-----\n",
		KeyDirection:       "1",
		TLSCrypt:           "tls-crypt\nvalue",
		TLSCryptV2:         "tls-crypt-v2\nvalue",
		Username:           "user@example.com",
		Password:           "p@ss&word",
		PeerInfo:           map[string]string{"IV_VER": "3.git", "UV_ID": "jp"},
		Ping:               10,
		PingRestart:        60,
		TranWindow:         &tranWindow,
		HandshakeTimeout:   30,
		MTU:                1400,
		UDP:                true,
		IPStack:            map[string]any{"mode": "gvisor", "congestion-controller": "bbr"},
		RemoteDNSResolve:   true,
		DNS:                []string{"1.1.1.1", "8.8.8.8"},
		TFO:                true,
		MPTCP:              true,
		InterfaceName:      "eth0",
		RoutingMark:        1234,
		IPVersion:          "ipv4-prefer",
		DialerProxy:        "front-proxy",
	}

	encoded := EncodeOpenVPNURL(original)
	decoded, err := DecodeOpenVPNURL(encoded)
	if err != nil {
		t.Fatalf("DecodeOpenVPNURL failed: %v", err)
	}
	if !reflect.DeepEqual(original, decoded) {
		t.Fatalf("OpenVPN round trip mismatch:\noriginal: %#v\ndecoded:  %#v", original, decoded)
	}
	if !strings.Contains(encoded, "ca=-----BEGIN+CERTIFICATE-----%0ACA%0A") {
		t.Fatalf("PEM content was not URL encoded: %s", encoded)
	}
}

func TestOpenVPNEncodeProxyLinkFromClashYAML(t *testing.T) {
	const input = `proxies:
  - name: "example-openvpn-node"
    type: openvpn
    server: vpn.example.com
    port: 1242
    proto: tcp
    udp: true
    cipher: AES-128-CBC
    auth: SHA1
    data-ciphers: [AES-128-CBC]
    ca: |
      -----BEGIN CERTIFICATE-----
      CA-DATA
      -----END CERTIFICATE-----
    cert: |
      -----BEGIN CERTIFICATE-----
      CERT-DATA
      -----END CERTIFICATE-----
    key: |
      -----BEGIN RSA PRIVATE KEY-----
      KEY-DATA
      -----END RSA PRIVATE KEY-----
`
	var config Config
	if err := yaml.Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("yaml unmarshal failed: %v", err)
	}
	if len(config.Proxies) != 1 {
		t.Fatalf("proxy count = %d, want 1", len(config.Proxies))
	}

	link, err := EncodeProxyLink(config.Proxies[0])
	if err != nil {
		t.Fatalf("EncodeProxyLink failed: %v", err)
	}
	if !strings.HasPrefix(link, "openvpn://vpn.example.com:1242?") {
		t.Fatalf("unexpected OpenVPN link: %s", link)
	}

	parsed, err := ParseNodeLink(link)
	if err != nil {
		t.Fatalf("ParseNodeLink failed: %v", err)
	}
	if parsed.Protocol != "openvpn" {
		t.Fatalf("protocol = %q, want openvpn", parsed.Protocol)
	}

	proxy, err := LinkToProxy(Urls{Url: link}, OutputConfig{})
	if err != nil {
		t.Fatalf("LinkToProxy failed: %v", err)
	}
	if proxy.Name != "example-openvpn-node" || proxy.Server != "vpn.example.com" || int(proxy.Port) != 1242 {
		t.Fatalf("unexpected basic proxy fields: %#v", proxy)
	}
	if proxy.Ca != config.Proxies[0].Ca || proxy.Cert != config.Proxies[0].Cert || proxy.Key != config.Proxies[0].Key {
		t.Fatal("OpenVPN PEM fields changed during round trip")
	}
	if !reflect.DeepEqual(proxy.Data_ciphers, []string{"AES-128-CBC"}) {
		t.Fatalf("data-ciphers = %#v", proxy.Data_ciphers)
	}

	data, err := yaml.Marshal(proxy)
	if err != nil {
		t.Fatalf("yaml marshal failed: %v", err)
	}
	output := string(data)
	for _, want := range []string{
		"type: openvpn",
		"server: vpn.example.com",
		"port: 1242",
		"proto: tcp",
		"udp: true",
		"cipher: AES-128-CBC",
		"auth: SHA1",
		"data-ciphers:",
		"ca: |",
		"cert: |",
		"key: |",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("YAML output missing %q:\n%s", want, output)
		}
	}
}

func TestOpenVPNIPv6RenameAndDefaultPort(t *testing.T) {
	link := EncodeOpenVPNURL(OpenVPN{
		Name:   "old-name",
		Server: "2001:db8::1",
		Port:   443,
		CA:     "certificate",
	})
	decoded, err := DecodeOpenVPNURL(link)
	if err != nil {
		t.Fatalf("IPv6 DecodeOpenVPNURL failed: %v", err)
	}
	if decoded.Server != "2001:db8::1" || decoded.Port != 443 {
		t.Fatalf("unexpected IPv6 endpoint: %#v", decoded)
	}

	renamed := RenameNodeLink(link, "new-name")
	decoded, err = DecodeOpenVPNURL(renamed)
	if err != nil {
		t.Fatalf("DecodeOpenVPNURL(renamed) failed: %v", err)
	}
	if decoded.Name != "new-name" {
		t.Fatalf("renamed node name = %q", decoded.Name)
	}

	withoutPort, err := DecodeOpenVPNURL("openvpn://vpn.example.com?ca=test#default-port")
	if err != nil {
		t.Fatalf("default port decode failed: %v", err)
	}
	if withoutPort.Port != defaultOpenVPNPort {
		t.Fatalf("default port = %d, want %d", withoutPort.Port, defaultOpenVPNPort)
	}
}

func TestOpenVPNClientSupportAndMetadata(t *testing.T) {
	for _, client := range []string{ClientClash, ClientMihomo} {
		if !ProtocolSupportsClient("openvpn", client) {
			t.Fatalf("openvpn should support %s", client)
		}
	}
	for _, client := range []string{ClientV2ray, ClientSurge} {
		if ProtocolSupportsClient("openvpn", client) {
			t.Fatalf("openvpn should not support %s", client)
		}
	}

	meta := GetProtocolMeta("openvpn")
	if meta == nil {
		t.Fatal("GetProtocolMeta(openvpn) returned nil")
	}
	fields := make(map[string]FieldMeta, len(meta.Fields))
	for _, field := range meta.Fields {
		fields[field.Name] = field
	}
	for _, name := range []string{"CA", "Cert", "Key", "TLSAuth", "TLSCrypt", "TLSCryptV2"} {
		if !fields[name].Multiline {
			t.Fatalf("field %s should be multiline", name)
		}
	}
	for _, name := range []string{"Password", "Key", "TLSAuth", "TLSCrypt", "TLSCryptV2"} {
		if !fields[name].Secret {
			t.Fatalf("field %s should be secret", name)
		}
	}
}
