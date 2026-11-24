package protocol

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuildSSSurgeLineCipherWhitelist 验证 Surge 导出会将已确认等价的 SS 加密方式别名
// 重写为 Surge 支持的标准名称，并仅保留白名单内的加密方式。
func TestBuildSSSurgeLineCipherWhitelist(t *testing.T) {
	for alias, canonical := range surgeSSCipherAliases {
		if _, ok := surgeSupportedSSCiphers[canonical]; !ok {
			t.Fatalf("alias %q maps to unsupported canonical cipher %q", alias, canonical)
		}
	}

	supportedCases := []struct {
		name           string
		cipher         string
		wantMethodName string
	}{
		// AEAD 2022
		{name: "2022-aes-128-gcm", cipher: "2022-blake3-aes-128-gcm", wantMethodName: "2022-blake3-aes-128-gcm"},
		{name: "2022-aes-256-gcm", cipher: "2022-blake3-aes-256-gcm", wantMethodName: "2022-blake3-aes-256-gcm"},
		// AEAD
		{name: "aes-128-gcm", cipher: "aes-128-gcm", wantMethodName: "aes-128-gcm"},
		{name: "aes-192-gcm", cipher: "aes-192-gcm", wantMethodName: "aes-192-gcm"},
		{name: "aes-256-gcm", cipher: "aes-256-gcm", wantMethodName: "aes-256-gcm"},
		{name: "uppercase-aes-256-gcm", cipher: "AES-256-GCM", wantMethodName: "aes-256-gcm"},
		{name: "chacha20-ietf-poly1305", cipher: "chacha20-ietf-poly1305", wantMethodName: "chacha20-ietf-poly1305"},
		{name: "xchacha20-ietf-poly1305", cipher: "xchacha20-ietf-poly1305", wantMethodName: "xchacha20-ietf-poly1305"},
		// Stream (legacy)
		{name: "rc4-md5", cipher: "rc4-md5", wantMethodName: "rc4-md5"},
		{name: "aes-256-cfb", cipher: "aes-256-cfb", wantMethodName: "aes-256-cfb"},
		// None
		{name: "none", cipher: "none", wantMethodName: "none"},
		// Confirmed equivalent aliases
		{name: "alias-chacha20-poly1305", cipher: "chacha20-poly1305", wantMethodName: "chacha20-ietf-poly1305"},
		{name: "alias-uppercase-chacha20-poly1305", cipher: "CHACHA20-POLY1305", wantMethodName: "chacha20-ietf-poly1305"},
		{name: "alias-xchacha20-poly1305", cipher: "xchacha20-poly1305", wantMethodName: "xchacha20-ietf-poly1305"},
		{name: "alias-aead_aes_128_gcm", cipher: "aead_aes_128_gcm", wantMethodName: "aes-128-gcm"},
		{name: "alias-uppercase-aead_aes_192_gcm", cipher: "AEAD_AES_192_GCM", wantMethodName: "aes-192-gcm"},
		{name: "alias-aead_aes_256_gcm", cipher: "aead_aes_256_gcm", wantMethodName: "aes-256-gcm"},
		{name: "alias-aead_chacha20_poly1305", cipher: "aead_chacha20_poly1305", wantMethodName: "chacha20-ietf-poly1305"},
		{name: "alias-aead_xchacha20_poly1305", cipher: "aead_xchacha20_poly1305", wantMethodName: "xchacha20-ietf-poly1305"},
		{name: "alias-plain", cipher: "plain", wantMethodName: "none"},
	}

	for _, tc := range supportedCases {
		t.Run("supported/"+tc.name, func(t *testing.T) {
			link := EncodeSSURL(Ss{
				Name:   "ss-" + tc.name,
				Server: "1.2.3.4",
				Port:   8388,
				Param: Param{
					Cipher:   tc.cipher,
					Password: "test-password",
				},
			})
			line, name, err := buildSSSurgeLine(link, OutputConfig{})
			if err != nil {
				t.Fatalf("supported cipher %q should export, got error: %v", tc.cipher, err)
			}
			if name != "ss-"+tc.name {
				t.Fatalf("unexpected node name %q", name)
			}
			wantToken := "encrypt-method=" + tc.wantMethodName
			if !strings.Contains(line, wantToken) {
				t.Fatalf("expected exported line %q to contain %q", line, wantToken)
			}
		})
	}

	unsupportedCiphers := []string{
		"2022-blake3-chacha20-poly1305",
		"2022-blake3-chacha8-poly1305",
		"aes-128-ccm",
		"table",
		"camellia-256-cfb",
	}

	for _, cipher := range unsupportedCiphers {
		t.Run("unsupported/"+cipher, func(t *testing.T) {
			link := EncodeSSURL(Ss{
				Name:   "bad-" + cipher,
				Server: "1.2.3.4",
				Port:   8388,
				Param: Param{
					Cipher:   cipher,
					Password: "test-password",
				},
			})
			_, _, err := buildSSSurgeLine(link, OutputConfig{})
			if err == nil {
				t.Fatalf("unsupported cipher %q should be rejected", cipher)
			}
			if !strings.Contains(err.Error(), "unsupported Shadowsocks cipher for Surge export") {
				t.Fatalf("unexpected error for %q: %v", cipher, err)
			}
		})
	}
}

// TestEncodeSurge_SSCipherAliasAndFiltering 验证端到端 Surge 导出时，别名节点会被重写并保留在动态代理组中，
// 同时真正不被 Surge 支持的加密方式节点会被跳过且不进入代理组。
func TestEncodeSurge_SSCipherAliasAndFiltering(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "surge-ss-cipher-alias-template.conf")
	template := "[Proxy]\n\n[Proxy Group]\nProxy = select\nAuto = url-test, url=http://www.gstatic.com/generate_204, interval=300\n"
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("write template failed: %v", err)
	}

	urls := []string{
		EncodeSSURL(Ss{
			Name:   "SS-Alias",
			Server: "1.2.3.4",
			Port:   8388,
			Param: Param{
				Cipher:   "chacha20-poly1305",
				Password: "test-password",
			},
		}),
		EncodeSSURL(Ss{
			Name:   "SS-Plain",
			Server: "1.2.3.5",
			Port:   8388,
			Param: Param{
				Cipher:   "plain",
				Password: "test-password",
			},
		}),
		EncodeSSURL(Ss{
			Name:   "SS-Unsupported",
			Server: "1.2.3.6",
			Port:   8388,
			Param: Param{
				Cipher:   "2022-blake3-chacha20-poly1305",
				Password: "test-password",
			},
		}),
	}

	output, err := EncodeSurge(urls, OutputConfig{
		Surge: templatePath,
		Udp:   true,
	})
	if err != nil {
		t.Fatalf("EncodeSurge failed: %v", err)
	}

	assertContains(t, "rewritten chacha20-poly1305 node", output, "SS-Alias = ss, 1.2.3.4, 8388, encrypt-method=chacha20-ietf-poly1305, password=test-password, udp-relay=true")
	assertContains(t, "rewritten plain node", output, "SS-Plain = ss, 1.2.3.5, 8388, encrypt-method=none, password=test-password, udp-relay=true")
	if strings.Contains(output, "SS-Unsupported") {
		t.Fatalf("unsupported cipher node should not appear in output:\n%s", output)
	}
	assertContains(t, "dynamic select group", output, "Proxy = select, SS-Alias, SS-Plain")
	assertContains(t, "dynamic url-test group", output, "Auto = url-test, url=http://www.gstatic.com/generate_204, interval=300, SS-Alias, SS-Plain")
}
