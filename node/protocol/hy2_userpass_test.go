package protocol

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

var hy2URLUserpassCases = []struct {
	name         string
	urlSuffix    string
	wantPassword string
	wantAuth     string
}{
	{
		name:         "plain_password",
		urlSuffix:    "test-password@example.invalid:443",
		wantPassword: "test-password",
		wantAuth:     "test-password",
	},
	{
		name:         "raw_userpass",
		urlSuffix:    "user:pass@example.invalid:443",
		wantPassword: "user:pass",
		wantAuth:     "user:pass",
	},
	{
		name:         "raw_password_with_colons",
		urlSuffix:    "user:pass:tail@example.invalid:443",
		wantPassword: "user:pass:tail",
		wantAuth:     "user:pass:tail",
	},
	{
		name:         "whole_userpass_encoded",
		urlSuffix:    "user%3Apass@example.invalid:443",
		wantPassword: "user:pass",
		wantAuth:     "user:pass",
	},
	{
		name:         "separately_encoded_components",
		urlSuffix:    "user%3Aname%40mail:p%40ss%3Aword%2F%3F%23@example.invalid:443",
		wantPassword: "user:name@mail:p@ss:word/?#",
		wantAuth:     "user:name@mail:p@ss:word/?#",
	},
	{
		name:         "reserved_characters_and_plus",
		urlSuffix:    "user+name:p+ass!$&'()*+,;=%40%2F%3F%23%5B%5D@example.invalid:443",
		wantPassword: "user+name:p+ass!$&'()*+,;=@/?#[]",
		wantAuth:     "user+name:p+ass!$&'()*+,;=@/?#[]",
	},
	{
		name:         "literal_percent_sequences_without_separator",
		urlSuffix:    "user%253Apass%2540%252F%253F%2523%2525%25@example.invalid:443",
		wantPassword: "user%3Apass%40%2F%3F%23%25%",
		wantAuth:     "user%3Apass%40%2F%3F%23%25%",
	},
	{
		name:         "literal_percent_sequences_in_both_components",
		urlSuffix:    "user%253A%25:pass%2540%252F%253F%2523%2525%25@example.invalid:443",
		wantPassword: "user%3A%:pass%40%2F%3F%23%25%",
		wantAuth:     "user%3A%:pass%40%2F%3F%23%25%",
	},
	{
		name:         "empty_password_keeps_separator",
		urlSuffix:    "user:@example.invalid:443",
		wantPassword: "user:",
		wantAuth:     "user:",
	},
	{
		name:         "encoded_empty_password_keeps_separator",
		urlSuffix:    "user%3A@example.invalid:443",
		wantPassword: "user:",
		wantAuth:     "user:",
	},
	{
		name:         "empty_username",
		urlSuffix:    ":pass@example.invalid:443",
		wantPassword: ":pass",
		wantAuth:     ":pass",
	},
	{
		name:         "empty_components_keep_separator",
		urlSuffix:    ":@example.invalid:443",
		wantPassword: ":",
		wantAuth:     ":",
	},
	{
		name:         "no_userinfo",
		urlSuffix:    "example.invalid:443",
		wantPassword: "",
		wantAuth:     "",
	},
	{
		name:         "empty_userinfo_without_separator",
		urlSuffix:    "@example.invalid:443",
		wantPassword: "",
		wantAuth:     "",
	},
	{
		name:         "explicit_auth_with_plain_password",
		urlSuffix:    "test-password@example.invalid:443?auth=query-auth",
		wantPassword: "test-password",
		wantAuth:     "query-auth",
	},
	{
		name:         "explicit_auth_with_raw_userpass",
		urlSuffix:    "user:pass@example.invalid:443?auth=query%3Aauth",
		wantPassword: "user:pass",
		wantAuth:     "query:auth",
	},
	{
		name:         "explicit_auth_with_encoded_userpass",
		urlSuffix:    "user%3Apass@example.invalid:443?auth=query%3Aauth",
		wantPassword: "user:pass",
		wantAuth:     "query:auth",
	},
	{
		name:         "empty_auth_falls_back_to_userpass",
		urlSuffix:    "user:pass@example.invalid:443?auth=",
		wantPassword: "user:pass",
		wantAuth:     "user:pass",
	},
	{
		name:         "empty_auth_keeps_empty_password_separator",
		urlSuffix:    "user:@example.invalid:443?auth=",
		wantPassword: "user:",
		wantAuth:     "user:",
	},
	{
		name:         "auth_query_decoded_once",
		urlSuffix:    "user:pass@example.invalid:443?auth=query%253A%2540%252F%25%2B+value",
		wantPassword: "user:pass",
		wantAuth:     "query%3A%40%2F%+ value",
	},
	{
		name:         "auth_without_userinfo",
		urlSuffix:    "example.invalid:443?auth=query%3Aauth",
		wantPassword: "",
		wantAuth:     "query:auth",
	},
	{
		name:         "empty_auth_without_userinfo",
		urlSuffix:    "example.invalid:443?auth=",
		wantPassword: "",
		wantAuth:     "",
	},
}

func TestDecodeHY2URLUserpass(test *testing.T) {
	for _, scheme := range []string{"hy2", "hysteria2"} {
		for _, testCase := range hy2URLUserpassCases {
			test.Run(scheme+"/"+testCase.name, func(test *testing.T) {
				decoded, err := DecodeHY2URL(scheme + "://" + testCase.urlSuffix)
				if err != nil {
					test.Fatalf("DecodeHY2URL() error: %v", err)
				}
				if decoded.Password != testCase.wantPassword || decoded.Auth != testCase.wantAuth {
					test.Fatalf("decoded Password/Auth = %q / %q, want %q / %q", decoded.Password, decoded.Auth, testCase.wantPassword, testCase.wantAuth)
				}
				if decoded.Host != "example.invalid" || decoded.Port != 443 {
					test.Fatalf("decoded endpoint = %s:%v, want example.invalid:443", decoded.Host, decoded.Port)
				}
			})
		}
	}
}

func TestHY2URLUserpassRoundTrip(test *testing.T) {
	for _, scheme := range []string{"hy2", "hysteria2"} {
		for _, testCase := range hy2URLUserpassCases {
			test.Run(scheme+"/"+testCase.name, func(test *testing.T) {
				original, err := DecodeHY2URL(scheme + "://" + testCase.urlSuffix)
				if err != nil {
					test.Fatalf("initial DecodeHY2URL() error: %v", err)
				}
				encoded := EncodeHY2URL(original)
				roundTrip, err := DecodeHY2URL(encoded)
				if err != nil {
					test.Fatalf("round-trip DecodeHY2URL() error: %v", err)
				}
				if roundTrip.Password != testCase.wantPassword || roundTrip.Auth != testCase.wantAuth {
					test.Fatalf("round-trip Password/Auth = %q / %q, want %q / %q", roundTrip.Password, roundTrip.Auth, testCase.wantPassword, testCase.wantAuth)
				}
				if !reflect.DeepEqual(roundTrip, original) {
					test.Fatalf("round trip changed HY2 fields: got %#v, want %#v", roundTrip, original)
				}
			})
		}
	}
}

func TestBuildHY2ProxyUserpass(test *testing.T) {
	for _, scheme := range []string{"hy2", "hysteria2"} {
		for _, testCase := range hy2URLUserpassCases {
			test.Run(scheme+"/"+testCase.name, func(test *testing.T) {
				proxy, err := buildHY2Proxy(Urls{Url: scheme + "://" + testCase.urlSuffix}, OutputConfig{})
				if err != nil {
					test.Fatalf("buildHY2Proxy() error: %v", err)
				}
				if proxy.Password != testCase.wantPassword || proxy.Auth != testCase.wantAuth {
					test.Fatalf("proxy Password/Auth = %q / %q, want %q / %q", proxy.Password, proxy.Auth, testCase.wantPassword, testCase.wantAuth)
				}
				if proxy.Type != "hysteria2" || proxy.Server != "example.invalid" || proxy.Port != FlexPort(443) {
					test.Fatalf("proxy type/endpoint = %s / %s:%v, want hysteria2 / example.invalid:443", proxy.Type, proxy.Server, proxy.Port)
				}
				exported, err := yaml.Marshal(proxy)
				if err != nil {
					test.Fatalf("yaml.Marshal() error: %v", err)
				}
				var document map[string]any
				if err := yaml.Unmarshal(exported, &document); err != nil {
					test.Fatalf("yaml.Unmarshal() error: %v", err)
				}
				for fieldName, expectedValue := range map[string]any{"password": testCase.wantPassword, "auth": testCase.wantAuth} {
					actualValue, present := document[fieldName]
					if expectedValue == "" {
						if present {
							test.Errorf("empty %s must retain omitempty behavior, got %#v", fieldName, actualValue)
						}
					} else if !present || actualValue != expectedValue {
						test.Errorf("exported %s = %#v (present=%t), want %#v", fieldName, actualValue, present, expectedValue)
					}
				}
			})
		}
	}
}
