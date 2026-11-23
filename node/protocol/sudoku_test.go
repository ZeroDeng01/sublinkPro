package protocol

import (
	"encoding/base64"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSudokuOfficialShortLinkAndRawEdit(t *testing.T) {
	const payload = `{"n":"数独节点","h":"2001:db8::1","p":8443,"k":"example-key","a":"ascii","m":1081,"x":true,"hm":"ws","ht":true,"hh":"edge.example","hx":"auto","hy":"proxy"}`
	link := "SUDOKU://" + base64.URLEncoding.EncodeToString([]byte(payload))
	identity, err := ExtractLinkIdentity(link)
	if err != nil || identity.Protocol != "sudoku" || identity.Name != "数独节点" || identity.Address != "[2001:db8::1]:8443" {
		t.Fatalf("unexpected Sudoku identity: %+v, %v", identity, err)
	}
	parsed, err := ParseNodeLink(link)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Fields["CustomTables"] != "" || parsed.Fields["AEADMethod"] != "none" || parsed.Fields["TableType"] != "prefer_ascii" {
		t.Fatal("short-link defaults and editable fields were not resolved")
	}
	updated, err := UpdateNodeLinkFields(link, `{"Name":"renamed","PaddingMin":0,"CustomTables":"xpxvvpvv, vxpvxvvp"}`)
	if err != nil {
		t.Fatal(err)
	}
	target, err := DecodeSudokuURL(updated)
	if err != nil || target.Name != "renamed" || target.LocalPort != 1081 || target.PaddingMin != 0 || target.PaddingMax != 15 ||
		!target.PackedDownlink || !target.HTTPMaskTLS || target.HTTPMaskHost != "edge.example" || target.CustomTables != "xpxvvpvv, vxpvxvvp" {
		t.Fatalf("Sudoku edit lost node settings, decode error: %v", err)
	}
	if !SupportsClientForLink(updated, ClientClash) || !SupportsClientForLink(updated, ClientMihomo) ||
		SupportsClientForLink(updated, ClientV2ray) || SupportsClientForLink(updated, ClientSurge) {
		t.Fatal("unexpected Sudoku client support")
	}
}

func TestSudokuClashRoundTrip(t *testing.T) {
	const document = `
name: original
type: sudoku
server: 2001:db8::2
port: 443
key: example-key
padding-min: 0
padding-max: 0
table-type: up_entropy_down_ascii
custom-tables: [xpxvvpvv, vxpvxvvp]
enable-pure-downlink: false
multiplex: "off"
http-mask: false
http-mask-mode: poll
http-mask-multiplex: auto
path-root: legacy
httpmask:
  mode: ws
  tls: true
  host: front.example:443
  multiplex: "on"
mptcp: true
dialer-proxy: base
additional-field: keep
`
	var imported Proxy
	if err := yaml.Unmarshal([]byte(document), &imported); err != nil {
		t.Fatal(err)
	}
	link, err := EncodeProxyLink(imported)
	if err != nil {
		t.Fatal(err)
	}
	extra, err := EncodeClashExtra(imported)
	if err != nil || strings.TrimSpace(extra) != "additional-field: keep" {
		t.Fatalf("unexpected retained fields: %q, %v", extra, err)
	}
	rebuilt, err := LinkToProxy(Urls{Url: link, ClashExtra: extra}, OutputConfig{Udp: true})
	if err != nil {
		t.Fatal(err)
	}
	if GenerateProxyContentHash(imported) != GenerateProxyContentHash(rebuilt) {
		t.Fatal("Sudoku content hash changed during YAML round trip")
	}
	data, err := yaml.Marshal(rebuilt)
	if err != nil {
		t.Fatal(err)
	}
	var exported map[string]any
	if err := yaml.Unmarshal(data, &exported); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"name": "original", "type": "sudoku", "server": "2001:db8::2", "port": 443, "key": "example-key",
		"padding-min": 0, "padding-max": 0, "table-type": "up_entropy_down_ascii", "aead-method": "chacha20-poly1305",
		"custom-tables": []any{"xpxvvpvv", "vxpvxvvp"}, "enable-pure-downlink": false, "multiplex": "on",
		"httpmask": map[string]any{"mode": "ws", "tls": true, "host": "front.example:443", "path-root": "legacy"},
		"udp":      true, "mptcp": true, "dialer-proxy": "base", "additional-field": "keep",
	}
	if !reflect.DeepEqual(exported, want) {
		t.Fatalf("unexpected Sudoku export:\n%s", data)
	}
}

func TestSudokuRejectsInvalidLinks(t *testing.T) {
	for _, payload := range []string{
		`{"h":"example.org","p":443}`,
		`{"p":443,"k":"key"}`,
		`{"h":"example.org","p":65536,"k":"key"}`,
		`{"h":"example.org","p":443,"k":"key","padding-min":16}`,
		`{"h":"example.org","p":443,"k":"key","hm":"unknown"}`,
		`{"h":"example.org","p":443,"k":"key","a":"invalid"}`,
	} {
		if _, err := DecodeSudokuURL(sudokuURLPrefix + base64.RawURLEncoding.EncodeToString([]byte(payload))); err == nil {
			t.Fatalf("accepted invalid Sudoku payload %s", payload)
		}
	}
	if _, err := DecodeSudokuURL("sudoku://not+base64url"); err == nil {
		t.Fatal("accepted invalid Sudoku encoding")
	}
}

func TestSudokuClashFieldsDoNotCaptureOtherProtocolExtras(t *testing.T) {
	var proxy Proxy
	if err := yaml.Unmarshal([]byte("type: vless\nmultiplex: {enabled: true}\npath-root: custom\n"), &proxy); err != nil {
		t.Fatal(err)
	}
	extra, err := EncodeClashExtra(proxy)
	if err != nil || !strings.Contains(extra, "multiplex: {enabled: true}") || !strings.Contains(extra, "path-root: custom") {
		t.Fatalf("expected both unknown attributes, got %q, %v", extra, err)
	}
}
