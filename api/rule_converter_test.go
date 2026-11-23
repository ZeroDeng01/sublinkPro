package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateSurgeProxyGroupsPreservesFilter(t *testing.T) {
	for _, filter := range []string{
		"(香港|HK|Hong Kong)",
		"(?i)(AWS|Azure|Oracle)",
		"(?:AWS|Amazon)(?:[ _-]?Cloud)?",
		"(?<![a-z])BCE(?![a-z])",
		"(?i)(AWS|Amazon(?:[ _-]?Web)?|(?<![a-z])BCE(?![a-z]))",
	} {
		group := ACLProxyGroup{Name: "Cloud", Type: "select", Filter: filter}
		want := "[Proxy Group]\nCloud = select, policy-regex-filter=" + filter + ", include-all-proxies=1"
		if got := generateSurgeProxyGroups([]ACLProxyGroup{group}, false); got != want {
			t.Errorf("filter %q:\ngot  %s\nwant %s", filter, got, want)
		}
	}
}

func TestConvertRulesSurgeProxyGroupFilters(t *testing.T) {
	const nested = "(?:AWS|Amazon)(?:[ _-]?Cloud)?"
	const assertion = "(?<![a-z])BCE(?![a-z])"
	const flagged = "(?i)(AWS|Azure|Oracle)"
	const complexFilter = "(?i)(AWS|Amazon(?:[ _-]?Web)?|(?<![a-z])BCE(?![a-z]))"
	config := strings.Join([]string{
		"custom_proxy_group=Region`select`(香港|HK|Hong Kong)",
		"custom_proxy_group=Cloud`select`" + complexFilter + "`[]REJECT",
		"custom_proxy_group=Speed`url-test`" + nested + "`https://example.org/ping`90,,25",
		"custom_proxy_group=Backup`fallback`[]Cloud`" + assertion,
		"custom_proxy_group=Balance`load-balance`" + flagged + "`" + assertion,
		"custom_proxy_group=Manual`select`[]DIRECT`[]REJECT`[](US|EU)",
		"custom_proxy_group=All`select`.*",
		"ruleset=Manual,[]FINAL",
	}, "\n")
	for _, includeAll := range []bool{false, true} {
		allGroup := "All = select"
		if includeAll {
			allGroup += ", include-all-proxies=1"
		}
		want := strings.Join([]string{
			"[General]", "loglevel = notify", "", "[Proxy Group]",
			"Region = select, policy-regex-filter=(香港|HK|Hong Kong), include-all-proxies=1",
			"Cloud = select, REJECT, policy-regex-filter=" + complexFilter + ", include-all-proxies=1",
			"Speed = url-test, url=https://example.org/ping, interval=90, timeout=5, tolerance=25, policy-regex-filter=" + nested + ", include-all-proxies=1",
			"Backup = fallback, Cloud, url=http://www.gstatic.com/generate_204, interval=300, timeout=5, tolerance=150, policy-regex-filter=" + assertion + ", include-all-proxies=1",
			"Balance = load-balance, policy-regex-filter=(?:" + flagged + ")|(?:" + assertion + "), include-all-proxies=1",
			"Manual = select, DIRECT, REJECT, (US|EU)", allGroup,
			"", "[Rule]", "FINAL,Manual",
		}, "\n")
		request := ConvertRulesRequest{Category: "surge", Template: "[General]\nloglevel = notify", EnableIncludeAll: includeAll}
		if got := convertACLForTest(t, config, request); got != want {
			t.Errorf("include-all=%t:\ngot:\n%s\nwant:\n%s", includeAll, got, want)
		}
	}
}

func TestConvertRulesClashProxyGroupFilters(t *testing.T) {
	config := "custom_proxy_group=Region`select`(香港|HK)`(日本|JP)`[]REJECT\nruleset=Region,[]FINAL"
	request := ConvertRulesRequest{Category: "clash", Template: "proxies: []"}
	want := strings.Join([]string{
		"proxies: []", "", "proxy-groups:", "  - name: Region", "    type: select",
		"    include-all: true", "    filter: (香港|HK|日本|JP)", "    proxies:", "      - REJECT",
		"", "rules:", "  - MATCH,Region",
	}, "\n")
	if got := convertACLForTest(t, config, request); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func convertACLForTest(t *testing.T, config string, request ConvertRulesRequest) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(writer, config)
	}))
	defer server.Close()
	request.RuleSource = server.URL
	response := decodeAPIResponse(t, performJSONRequest(t, ConvertRules, http.MethodPost, request))
	var result ConvertRulesResponse
	if err := json.Unmarshal(response.Data, &result); err != nil || response.Code != http.StatusOK {
		t.Fatalf("convert rules: code %d, message %q, error %v", response.Code, response.Msg, err)
	}
	return result.Content
}
