package protocol

import (
	"strings"
	"testing"
)

// TestBuildSSSurgeLineCipherWhitelist 验证 Surge 导出只保留官方支持列表内的 SS 加密方式。
func TestBuildSSSurgeLineCipherWhitelist(t *testing.T) {
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
		"chacha20-poly1305",
		"plain",
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
