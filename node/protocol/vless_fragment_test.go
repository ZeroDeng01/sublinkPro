package protocol

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const fragmentLink = "vless://b831381d-6324-4d53-ad4f-8cda48b30811@168.113.131.162:1443" +
	"?encryption=none&security=reality&type=tcp&sni=168.113.131.162&fp=firefox" +
	"&pbk=2V705nBMsH_w0dWqRNfD_bjOJhp5Opc1ELgrfiPFggQ&sid=9ecdbcce4f70a126&flow=xtls-rprx-vision" +
	"&fm=%7B%22tcp%22%3A%5B%7B%22type%22%3A%22fragment%22%2C%22settings%22%3A%7B%22delay%22%3A%220%22%2C%22length%22%3A%2210-20%22%2C%22packets%22%3A%22tlshello%22%7D%7D%5D%7D" +
	"#frag"

func TestVLESSFragmentReachesTheConfig(t *testing.T) {
	proxy, err := buildVLESSProxy(Urls{Url: fragmentLink}, OutputConfig{})
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

	encoded, err := yaml.Marshal(proxy)
	if err != nil {
		t.Fatalf("yaml: %v", err)
	}
	var rendered map[string]any
	if err := yaml.Unmarshal(encoded, &rendered); err != nil {
		t.Fatalf("yaml parse: %v", err)
	}
	fragment, ok := rendered["fragment"].(map[string]any)
	if !ok {
		t.Fatalf("no fragment in:\n%s", encoded)
	}
	if fragment["packets"] != "tlshello" {
		t.Fatalf("rendered packets = %v", fragment["packets"])
	}
}

func TestVLESSFragmentSurvivesTheLink(t *testing.T) {
	decoded, err := DecodeVLESSURL(fragmentLink)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.Query.Fm == "" {
		t.Fatal("fm was dropped while decoding")
	}
	rebuilt := EncodeVLESSURL(decoded)
	if !strings.Contains(rebuilt, "fm=") {
		t.Fatalf("fm left out of %s", rebuilt)
	}
	again, err := DecodeVLESSURL(rebuilt)
	if err != nil {
		t.Fatalf("decode rebuilt: %v", err)
	}
	if again.Query.Fm != decoded.Query.Fm {
		t.Fatalf("fm changed:\n%s\n%s", decoded.Query.Fm, again.Query.Fm)
	}
}

func TestVLESSWithoutFragmentStaysClean(t *testing.T) {
	plain := strings.Split(fragmentLink, "&fm=")[0] + "#plain"
	proxy, err := buildVLESSProxy(Urls{Url: plain}, OutputConfig{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if proxy.Fragment_opts != nil {
		t.Fatalf("fragment set without fm: %v", proxy.Fragment_opts)
	}
	encoded, err := yaml.Marshal(proxy)
	if err != nil {
		t.Fatalf("yaml: %v", err)
	}
	if strings.Contains(string(encoded), "fragment") {
		t.Fatalf("fragment leaked into:\n%s", encoded)
	}
}

func TestVLESSFragmentOtherModesAreDropped(t *testing.T) {
	link := strings.Replace(fragmentLink, "%22tlshello%22", "%221-3%22", 1)
	proxy, err := buildVLESSProxy(Urls{Url: link}, OutputConfig{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if proxy.Fragment_opts != nil {
		t.Fatalf("unsupported mode kept: %v", proxy.Fragment_opts)
	}
}

func TestClashFragmentComesBackAsFm(t *testing.T) {
	query := buildVLESSFragmentQuery(map[string]any{
		"enabled": true, "packets": "tlshello", "length": "10-20", "delay": "0",
	})
	settings, ok := parseVLESSFinalMaskFragment(query)
	if !ok {
		t.Fatalf("not parseable: %s", query)
	}
	if settings.Length != "10-20" || settings.Delay != "0" || settings.Packets != "tlshello" {
		t.Fatalf("round trip lost settings: %+v", settings)
	}
}
