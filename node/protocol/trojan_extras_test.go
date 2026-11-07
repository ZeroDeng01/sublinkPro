package protocol

import (
	"net/url"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// trojanFragmentFm 是 xray 的 finalmask 分片参数，mihomo 侧对应 fragment 选项。
const trojanFragmentFm = `{"tcp":[{"type":"fragment","settings":{"delay":"0","length":"10-20","packets":"tlshello"}}]}`

// trojanECHConfig 是配置形式的 ECH，会原样写进 ech-opts.config。
const trojanECHConfig = "AEn+DQBFKwAgACABWIHUGj4u+PIggYXcR5JF0gYk3dCRioBW8uJq9H4mKAAIAAEAAQABAANAEnB1YmxpYy5leGFtcGxlLmNvbQAA"

// trojanExtrasURL 拼出同时带 Reality、ECH 与 finalmask 的 trojan 链接。
func trojanExtrasURL(ech, fm string) string {
	query := url.Values{}
	query.Set("security", "reality")
	query.Set("sni", "example.com")
	query.Set("fp", "firefox")
	query.Set("pbk", "2V705nBMsH_w0dWqRNfD_bjOJhp5Opc1ELgrfiPFggQ")
	query.Set("sid", "9ecdbcce4f70a126")
	query.Set("fm", fm)
	query.Set("ech", ech)
	return "trojan://secret@203.0.113.10:443?" + query.Encode() + "#trojan-extras"
}

func TestTrojanExtrasReachTheConfig(t *testing.T) {
	proxy, err := buildTrojanProxy(Urls{Url: trojanExtrasURL("example.com+https://1.1.1.1/dns-query", trojanFragmentFm)}, OutputConfig{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if got := proxy.Fragment_opts["packets"]; got != "tlshello" {
		t.Fatalf("packets = %v", got)
	}
	if got := proxy.Fragment_opts["length"]; got != "10-20" {
		t.Fatalf("length = %v", got)
	}
	if got := proxy.Fragment_opts["delay"]; got != "0" {
		t.Fatalf("delay = %v", got)
	}
	if enabled, ok := proxy.Fragment_opts["enabled"].(bool); !ok || !enabled {
		t.Fatalf("enabled = %v", proxy.Fragment_opts["enabled"])
	}
	if proxy.Reality_opts["public-key"] != "2V705nBMsH_w0dWqRNfD_bjOJhp5Opc1ELgrfiPFggQ" || proxy.Reality_opts["short-id"] != "9ecdbcce4f70a126" {
		t.Fatalf("reality opts = %v", proxy.Reality_opts)
	}
	if proxy.ECH_opts["enable"] != true || proxy.ECH_opts["query-server-name"] != "example.com" {
		t.Fatalf("ech opts = %v", proxy.ECH_opts)
	}

	encoded, err := yaml.Marshal(proxy)
	if err != nil {
		t.Fatalf("yaml: %v", err)
	}
	var rendered map[string]any
	if err := yaml.Unmarshal(encoded, &rendered); err != nil {
		t.Fatalf("yaml parse: %v", err)
	}
	fragment, ok := rendered["fragment"].(map[string]any)
	if !ok || fragment["packets"] != "tlshello" {
		t.Fatalf("no fragment in:\n%s", encoded)
	}
	if _, ok := rendered["reality-opts"].(map[string]any); !ok {
		t.Fatalf("no reality-opts in:\n%s", encoded)
	}
	if _, ok := rendered["ech-opts"].(map[string]any); !ok {
		t.Fatalf("no ech-opts in:\n%s", encoded)
	}
}

func TestTrojanExtrasSurviveTheLink(t *testing.T) {
	const ech = "example.com+https://1.1.1.1/dns-query"
	decoded, err := DecodeTrojanURL(trojanExtrasURL(ech, trojanFragmentFm))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.Query.Fm == "" {
		t.Fatal("fm was dropped while decoding")
	}
	if decoded.Query.Ech != ech {
		t.Fatalf("ech = %q", decoded.Query.Ech)
	}
	rebuilt := EncodeTrojanURL(decoded)
	if !strings.Contains(rebuilt, "fm=") || !strings.Contains(rebuilt, "ech=") {
		t.Fatalf("extras left out of %s", rebuilt)
	}
	again, err := DecodeTrojanURL(rebuilt)
	if err != nil {
		t.Fatalf("decode rebuilt: %v", err)
	}
	if again.Query.Fm != decoded.Query.Fm || again.Query.Ech != decoded.Query.Ech {
		t.Fatalf("extras changed:\n%s\n%s", decoded.Query.Fm, again.Query.Fm)
	}
}

func TestTrojanWithoutExtrasStaysClean(t *testing.T) {
	plain := "trojan://secret@203.0.113.10:443?security=tls&sni=example.com&type=tcp#plain"
	proxy, err := buildTrojanProxy(Urls{Url: plain}, OutputConfig{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(proxy.Fragment_opts) != 0 || len(proxy.ECH_opts) != 0 || len(proxy.Reality_opts) != 0 {
		t.Fatalf("extras set without parameters: fragment=%v ech=%v reality=%v", proxy.Fragment_opts, proxy.ECH_opts, proxy.Reality_opts)
	}
	encoded, err := yaml.Marshal(proxy)
	if err != nil {
		t.Fatalf("yaml: %v", err)
	}
	for _, key := range []string{"fragment", "ech-opts", "reality-opts"} {
		if strings.Contains(string(encoded), key+":") {
			t.Fatalf("%s leaked into:\n%s", key, encoded)
		}
	}
}

func TestTrojanFragmentOtherModesAreDropped(t *testing.T) {
	fm := strings.Replace(trojanFragmentFm, `"tlshello"`, `"1-3"`, 1)
	proxy, err := buildTrojanProxy(Urls{Url: trojanExtrasURL("", fm)}, OutputConfig{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if proxy.Fragment_opts != nil {
		t.Fatalf("unsupported mode kept: %v", proxy.Fragment_opts)
	}
}

func TestTrojanProxyComesBackAsExtras(t *testing.T) {
	proxy, err := buildTrojanProxy(Urls{Url: trojanExtrasURL(trojanECHConfig, trojanFragmentFm)}, OutputConfig{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	restored := ConvertProxyToTrojan(proxy)
	if restored.Query.Pbk != "2V705nBMsH_w0dWqRNfD_bjOJhp5Opc1ELgrfiPFggQ" || restored.Query.Sid != "9ecdbcce4f70a126" || restored.Query.Security != "reality" {
		t.Fatalf("reality lost: %#v", restored.Query)
	}
	if restored.Query.Ech != trojanECHConfig {
		t.Fatalf("ech lost: %q", restored.Query.Ech)
	}
	settings, ok := parseFinalMaskFragment(restored.Query.Fm)
	if !ok || settings.Length != "10-20" || settings.Delay != "0" || settings.Packets != "tlshello" {
		t.Fatalf("fragment lost: %q", restored.Query.Fm)
	}
	encoded := EncodeTrojanURL(restored)
	if !strings.Contains(encoded, "fm=") || !strings.Contains(encoded, "ech=") || !strings.Contains(encoded, "pbk=") {
		t.Fatalf("extras left out of %s", encoded)
	}
}
