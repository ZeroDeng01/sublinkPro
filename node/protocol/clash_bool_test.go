package protocol

import (
	"fmt"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestProxyYAMLBooleanFields(t *testing.T) {
	fields := []struct {
		key string
		get func(Proxy) bool
	}{
		{"udp", func(p Proxy) bool { return p.Udp }},
		{"tls", func(p Proxy) bool { return p.Tls }},
		{"tfo", func(p Proxy) bool { return p.Tfo }},
		{"skip-cert-verify", func(p Proxy) bool { return p.Skip_cert_verify }},
		{"disable-sni", func(p Proxy) bool { return p.Disable_sni }},
		{"mptcp", func(p Proxy) bool { return p.Mptcp }},
		{"remote-dns-resolve", func(p Proxy) bool { return p.Remote_dns_resolve }},
	}
	values := []struct {
		yaml string
		want bool
	}{
		{"true", true}, {"false", false},
		{"'true'", true}, {"'false'", false},
		{`"true"`, true}, {`"false"`, false},
		{`"yes"`, true}, // Preserve yaml.v3's existing YAML 1.1 compatibility.
	}
	for _, field := range fields {
		for _, value := range values {
			t.Run(field.key+"/"+value.yaml, func(t *testing.T) {
				t.Parallel()
				var proxy Proxy
				if err := yaml.Unmarshal([]byte(field.key+": "+value.yaml+"\n"), &proxy); err != nil {
					t.Fatal(err)
				}
				if got := field.get(proxy); got != value.want {
					t.Fatalf("%s = %v, want %v", field.key, got, value.want)
				}
			})
		}
	}
}

func TestProxyYAMLBooleanStringsPreserveOtherStringsAndInput(t *testing.T) {
	for _, source := range []string{
		"password: &flag 'true'\nudp: *flag\n",
		"defaults: &options {udp: 'true'}\n<<: *options\npassword: 'true'\n",
		"udp: 'true'\npassword: 'true'\n",
	} {
		t.Run(source, func(t *testing.T) {
			source += "name: 'false'\nfuture-option: 'true'\nws-opts:\n  headers:\n    udp: 'false'\n"
			var document yaml.Node
			if err := yaml.Unmarshal([]byte(source), &document); err != nil {
				t.Fatal(err)
			}
			before, err := yaml.Marshal(&document)
			if err != nil {
				t.Fatal(err)
			}
			var proxy Proxy
			if err := document.Decode(&proxy); err != nil {
				t.Fatal(err)
			}
			if !proxy.Udp || proxy.Password != "true" || proxy.Name != "false" {
				t.Fatalf("incorrect typed values: udp=%v password=%q name=%q", proxy.Udp, proxy.Password, proxy.Name)
			}
			headers, ok := proxy.Ws_opts["headers"].(map[string]any)
			if !ok || headers["udp"] != "false" {
				t.Fatal("untyped nested string was converted")
			}
			after, err := yaml.Marshal(&document)
			if err != nil || string(after) != string(before) {
				t.Fatalf("source YAML node was mutated: %v", err)
			}
			encoded, err := yaml.Marshal(proxy)
			if err != nil {
				t.Fatal(err)
			}
			var exported map[string]any
			if err := yaml.Unmarshal(encoded, &exported); err != nil {
				t.Fatal(err)
			}
			if exported["udp"] != true || exported["future-option"] != "true" || exported["password"] != "true" {
				t.Fatalf("export changed field types: %s", encoded)
			}
		})
	}
}

func TestProxyYAMLBooleanMergeSequenceAndDefaults(t *testing.T) {
	var proxy Proxy
	source := "first: &first {udp: 'true'}\nsecond: &second {tls: 'false'}\n<<: [*first, *second]\n"
	if err := yaml.Unmarshal([]byte(source), &proxy); err != nil {
		t.Fatal(err)
	}
	if !proxy.Udp || proxy.Tls {
		t.Fatal("incorrect booleans from merged aliases")
	}
	proxy = Proxy{Udp: true, Tls: true}
	if err := yaml.Unmarshal([]byte("udp: 'false'\n"), &proxy); err != nil {
		t.Fatal(err)
	}
	if proxy.Udp || proxy.Tls {
		t.Fatal("false/omitted fields retained stale values")
	}
}

func BenchmarkProxyYAMLBooleanStrings(b *testing.B) {
	for _, value := range []string{"true", "'true'"} {
		b.Run(value, func(b *testing.B) {
			data := []byte("type: ss\nserver: example.invalid\nport: 443\npassword: fixture\nudp: " + value + "\n")
			b.ReportAllocs()
			for b.Loop() {
				var proxy Proxy
				if err := yaml.Unmarshal(data, &proxy); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestProxyYAMLBooleanMergeAliasesKeepDecoderLimits(t *testing.T) {
	var source strings.Builder
	source.WriteString("options0: &options0 {udp: 'true'}\n")
	for i := 1; i <= 25; i++ {
		fmt.Fprintf(&source, "options%d: &options%d {<<: [*options%d, *options%d]}\n", i, i, i-1, i-1)
	}
	source.WriteString("<<: *options25\n")
	var proxy Proxy
	if err := yaml.Unmarshal([]byte(source.String()), &proxy); err == nil || !strings.Contains(err.Error(), "excessive aliasing") {
		t.Fatalf("expected yaml.v3's alias limit, got %v", err)
	}
}

func TestProxyYAMLBooleanInvalidValuesStillFail(t *testing.T) {
	for _, source := range []string{
		"udp: 'not-a-bool'\n", "udp: '1'\n", "udp: 1\n", "udp: [true]\n",
		"udp: 'true'\nudp: false\n", "<<: &loop {<<: *loop}\nudp: 'true'\n",
	} {
		t.Run(fmt.Sprintf("%q", source), func(t *testing.T) {
			var proxy Proxy
			if err := yaml.Unmarshal([]byte(source), &proxy); err == nil {
				t.Fatal("expected invalid input to remain an error")
			}
		})
	}
	var proxy Proxy
	if err := yaml.Unmarshal([]byte("udp: 'not-a-bool'\n"), &proxy); err == nil || !strings.Contains(err.Error(), "line 1") {
		t.Fatalf("lost source line in error: %v", err)
	}
}
