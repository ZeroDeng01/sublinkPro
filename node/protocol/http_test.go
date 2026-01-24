package protocol

import (
	"testing"
)

func TestDecodeHTTPURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    HTTP
		wantErr bool
	}{
		{
			name: "HTTP with username and password",
			url:  "http://user:pass@example.com:8080#TestNode",
			want: HTTP{
				Name:     "TestNode",
				Server:   "example.com",
				Port:     8080,
				Username: "user",
				Password: "pass",
				TLS:      false,
			},
			wantErr: false,
		},
		{
			name: "HTTPS with username and password",
			url:  "https://user:pass@example.com:443#TestNode",
			want: HTTP{
				Name:     "TestNode",
				Server:   "example.com",
				Port:     443,
				Username: "user",
				Password: "pass",
				TLS:      true,
			},
			wantErr: false,
		},
		{
			name: "HTTPS with skip-cert-verify and sni",
			url:  "https://user:pass@example.com:8443?skip-cert-verify=true&sni=example.com#TestNode",
			want: HTTP{
				Name:           "TestNode",
				Server:         "example.com",
				Port:           8443,
				Username:       "user",
				Password:       "pass",
				TLS:            true,
				SkipCertVerify: true,
				SNI:            "example.com",
			},
			wantErr: false,
		},
		{
			name: "HTTP without authentication",
			url:  "http://example.com:8080#TestNode",
			want: HTTP{
				Name:     "TestNode",
				Server:   "example.com",
				Port:     8080,
				TLS:      false,
			},
			wantErr: false,
		},
		{
			name: "HTTPS without port (default 443)",
			url:  "https://user:pass@example.com#TestNode",
			want: HTTP{
				Name:     "TestNode",
				Server:   "example.com",
				Port:     443,
				Username: "user",
				Password: "pass",
				TLS:      true,
			},
			wantErr: false,
		},
		{
			name: "HTTP without port (default 80)",
			url:  "http://example.com#TestNode",
			want: HTTP{
				Name:   "TestNode",
				Server: "example.com",
				Port:   80,
				TLS:    false,
			},
			wantErr: false,
		},
		{
			name:    "Invalid protocol",
			url:     "ftp://example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeHTTPURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeHTTPURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Name != tt.want.Name {
					t.Errorf("DecodeHTTPURL() Name = %v, want %v", got.Name, tt.want.Name)
				}
				if got.Server != tt.want.Server {
					t.Errorf("DecodeHTTPURL() Server = %v, want %v", got.Server, tt.want.Server)
				}
				if got.Port != tt.want.Port {
					t.Errorf("DecodeHTTPURL() Port = %v, want %v", got.Port, tt.want.Port)
				}
				if got.Username != tt.want.Username {
					t.Errorf("DecodeHTTPURL() Username = %v, want %v", got.Username, tt.want.Username)
				}
				if got.Password != tt.want.Password {
					t.Errorf("DecodeHTTPURL() Password = %v, want %v", got.Password, tt.want.Password)
				}
				if got.TLS != tt.want.TLS {
					t.Errorf("DecodeHTTPURL() TLS = %v, want %v", got.TLS, tt.want.TLS)
				}
				if got.SkipCertVerify != tt.want.SkipCertVerify {
					t.Errorf("DecodeHTTPURL() SkipCertVerify = %v, want %v", got.SkipCertVerify, tt.want.SkipCertVerify)
				}
				if got.SNI != tt.want.SNI {
					t.Errorf("DecodeHTTPURL() SNI = %v, want %v", got.SNI, tt.want.SNI)
				}
			}
		})
	}
}

func TestEncodeHTTPURL(t *testing.T) {
	tests := []struct {
		name string
		http HTTP
		want string
	}{
		{
			name: "HTTP with username and password",
			http: HTTP{
				Name:     "TestNode",
				Server:   "example.com",
				Port:     8080,
				Username: "user",
				Password: "pass",
				TLS:      false,
			},
			want: "http://user:pass@example.com:8080#TestNode",
		},
		{
			name: "HTTPS with username and password",
			http: HTTP{
				Name:     "TestNode",
				Server:   "example.com",
				Port:     443,
				Username: "user",
				Password: "pass",
				TLS:      true,
			},
			want: "https://user:pass@example.com:443#TestNode",
		},
		{
			name: "HTTPS with skip-cert-verify and sni",
			http: HTTP{
				Name:           "TestNode",
				Server:         "example.com",
				Port:           8443,
				Username:       "user",
				Password:       "pass",
				TLS:            true,
				SkipCertVerify: true,
				SNI:            "example.com",
			},
			want: "https://user:pass@example.com:8443?skip-cert-verify=true&sni=example.com#TestNode",
		},
		{
			name: "HTTP without authentication",
			http: HTTP{
				Name:   "TestNode",
				Server: "example.com",
				Port:   8080,
				TLS:    false,
			},
			want: "http://example.com:8080#TestNode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeHTTPURL(tt.http)
			if got != tt.want {
				t.Errorf("EncodeHTTPURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "HTTP with authentication",
			url:  "http://user:pass@example.com:8080#TestNode",
		},
		{
			name: "HTTPS with authentication",
			url:  "https://user:pass@example.com:443#TestNode",
		},
		{
			name: "HTTPS with skip-cert-verify and sni",
			url:  "https://user:pass@example.com:8443?skip-cert-verify=true&sni=example.com#TestNode",
		},
		{
			name: "HTTP without authentication",
			url:  "http://example.com:8080#TestNode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := DecodeHTTPURL(tt.url)
			if err != nil {
				t.Errorf("DecodeHTTPURL() error = %v", err)
				return
			}
			encoded := EncodeHTTPURL(decoded)
			if encoded != tt.url {
				t.Errorf("Round trip failed: original = %v, encoded = %v", tt.url, encoded)
			}
		})
	}
}

func TestConvertProxyToHTTP(t *testing.T) {
	tests := []struct {
		name  string
		proxy Proxy
		want  HTTP
	}{
		{
			name: "HTTP proxy",
			proxy: Proxy{
				Name:     "TestNode",
				Type:     "http",
				Server:   "example.com",
				Port:     8080,
				Username: "user",
				Password: "pass",
				Tls:      false,
			},
			want: HTTP{
				Name:     "TestNode",
				Server:   "example.com",
				Port:     8080,
				Username: "user",
				Password: "pass",
				TLS:      false,
			},
		},
		{
			name: "HTTPS proxy",
			proxy: Proxy{
				Name:             "TestNode",
				Type:             "http",
				Server:           "example.com",
				Port:             443,
				Username:         "user",
				Password:         "pass",
				Tls:              true,
				Skip_cert_verify: true,
				Sni:              "example.com",
			},
			want: HTTP{
				Name:           "TestNode",
				Server:         "example.com",
				Port:           443,
				Username:       "user",
				Password:       "pass",
				TLS:            true,
				SkipCertVerify: true,
				SNI:            "example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertProxyToHTTP(tt.proxy)
			if got.Name != tt.want.Name {
				t.Errorf("ConvertProxyToHTTP() Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Server != tt.want.Server {
				t.Errorf("ConvertProxyToHTTP() Server = %v, want %v", got.Server, tt.want.Server)
			}
			if got.Port != tt.want.Port {
				t.Errorf("ConvertProxyToHTTP() Port = %v, want %v", got.Port, tt.want.Port)
			}
			if got.Username != tt.want.Username {
				t.Errorf("ConvertProxyToHTTP() Username = %v, want %v", got.Username, tt.want.Username)
			}
			if got.Password != tt.want.Password {
				t.Errorf("ConvertProxyToHTTP() Password = %v, want %v", got.Password, tt.want.Password)
			}
			if got.TLS != tt.want.TLS {
				t.Errorf("ConvertProxyToHTTP() TLS = %v, want %v", got.TLS, tt.want.TLS)
			}
			if got.SkipCertVerify != tt.want.SkipCertVerify {
				t.Errorf("ConvertProxyToHTTP() SkipCertVerify = %v, want %v", got.SkipCertVerify, tt.want.SkipCertVerify)
			}
			if got.SNI != tt.want.SNI {
				t.Errorf("ConvertProxyToHTTP() SNI = %v, want %v", got.SNI, tt.want.SNI)
			}
		})
	}
}

func TestIsHTTPLink(t *testing.T) {
	tests := []struct {
		name     string
		link     string
		expected bool
	}{
		{
			name:     "HTTP代理节点",
			link:     "http://user:pass@example.com:8080#TestNode",
			expected: true,
		},
		{
			name:     "HTTPS代理节点",
			link:     "https://user:pass@example.com:443#TestNode",
			expected: true,
		},
		{
			name:     "HTTPS代理节点带参数",
			link:     "https://user:pass@example.com:8443?skip-cert-verify=true&sni=example.com#TestNode",
			expected: true,
		},
		{
			name:     "HTTP代理节点无认证",
			link:     "http://example.com:8080#TestNode",
			expected: true,
		},
		{
			name:     "订阅转换链接",
			link:     "http://example.com/sub.txt",
			expected: false,
		},
		{
			name:     "HTTPS订阅转换链接",
			link:     "https://example.com/sub.txt",
			expected: false,
		},
		{
			name:     "VMess节点",
			link:     "vmess://eyJhZGUiOiIxMjcuMC4wLjEiLCJhaWQiOiIwIiwiYWxwbiI6IiIsImZpbmdlcnByaW50IjoiIiwiaG9zdCI6IiIsImlkIjoiOGM4YzJiYmYtYzVjZS00NzU5LWI2NWMtOGI3MjEzZmNhZjY2IiwibmV0Ijoid3MiLCJwYXRoIjoiLyIsInBvcnQiOiI4MCIsInBzIjoiVGVzdCIsInNjeSI6ImF1dG8iLCJzbmkiOiIiLCJ0bHMiOiIiLCJ0eXBlIjoiIiwidiI6IjIifQ==",
			expected: false,
		},
		{
			name:     "SS节点",
			link:     "ss://YWVzLTI1Ni1nY206dGVzdEBleGFtcGxlLmNvbTo4Mzg4I1Rlc3ROb2Rl",
			expected: false,
		},
		{
			name:     "无效HTTP链接",
			link:     "http://",
			expected: false,
		},
		{
			name:     "无效HTTPS链接",
			link:     "https://",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHTTPLink(tt.link)
			if result != tt.expected {
				t.Errorf("IsHTTPLink() = %v, want %v", result, tt.expected)
			}
		})
	}
}