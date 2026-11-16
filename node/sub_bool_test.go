package node

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseClashConfigDataBooleanStringsKeepAllNodes(t *testing.T) {
	var source strings.Builder
	source.WriteString("proxies:\n")
	for i := range 115 {
		udp := "true"
		if i == 95 {
			udp = "'true'"
		}
		fmt.Fprintf(&source, "  - name: fixture-%d\n    type: ss\n    server: example.invalid\n    port: 443\n    cipher: aes-128-gcm\n    password: 'false'\n    udp: %s\n", i, udp)
	}
	source.WriteString("proxy-groups:\n  - name: all\n    type: select\n    include-all: 'true'\n    include-all-providers: \"false\"\n")
	config, yamlErr, providerErr := parseClashConfigData(context.Background(), http.DefaultClient, "https://example.invalid/root", []byte(source.String()), "", nil)
	if yamlErr != nil || providerErr != nil || len(config.Proxies) != 115 {
		t.Fatalf("got nodes=%d YAML error=%v provider error=%v", len(config.Proxies), yamlErr, providerErr)
	}
	if !config.Proxies[95].Udp || config.Proxies[95].Password != "false" || !config.ProxyGroups[0].IncludeAll || config.ProxyGroups[0].IncludeAllProviders {
		t.Fatal("incorrect decoded boolean/string values")
	}
}

func TestParseClashConfigDataProviderBooleanStrings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("proxies:\n  - name: fixture\n    type: trojan\n    server: example.invalid\n    port: 443\n    password: fixture\n    skip-cert-verify: 'true'\n    udp: \"false\"\n"))
	}))
	defer server.Close()
	source := "proxy-providers:\n  fixture:\n    type: http\n    url: " + server.URL + "/provider\nproxy-groups:\n  - include-all-providers: 'true'\n"
	config, yamlErr, providerErr := parseClashConfigData(context.Background(), server.Client(), server.URL+"/root", []byte(source), "", nil)
	if yamlErr != nil || providerErr != nil || len(config.Proxies) != 1 {
		t.Fatalf("got nodes=%d YAML error=%v provider error=%v", len(config.Proxies), yamlErr, providerErr)
	}
	if !config.Proxies[0].Skip_cert_verify || config.Proxies[0].Udp {
		t.Fatal("incorrect provider booleans")
	}
}

func TestParseClashConfigDataInvalidBooleanRemainsAnError(t *testing.T) {
	for _, source := range []string{
		"proxies:\n  - udp: 'invalid'\n",
		"proxies:\n  - udp: 'true'\nproxy-groups:\n  - include-all: 'invalid'\n",
	} {
		_, yamlErr, _ := parseClashConfigData(context.Background(), http.DefaultClient, "https://example.invalid/root", []byte(source), "", nil)
		if yamlErr == nil {
			t.Fatal("invalid boolean should not be accepted")
		}
	}
}
