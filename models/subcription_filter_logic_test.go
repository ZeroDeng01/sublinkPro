package models

import "testing"

func TestNormalizeFilterLogic(t *testing.T) {
	tests := map[string]string{
		"":     FilterLogicAnd,
		"and":  FilterLogicAnd,
		"AND":  FilterLogicAnd,
		" or ": FilterLogicOr,
		"OR":   FilterLogicOr,
		"xor":  FilterLogicAnd,
	}

	for input, want := range tests {
		if got := NormalizeFilterLogic(input); got != want {
			t.Errorf("NormalizeFilterLogic(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestApplyFiltersAndLogicRequiresAllGroups(t *testing.T) {
	nodes := filterLogicTestNodes()
	sub := Subcription{
		CountryWhitelist:  "US",
		ProtocolWhitelist: "vmess",
	}

	assertFilteredNodeNames(t, sub.ApplyFilters(nodes), []string{"us-vmess"})
}

func TestApplyFiltersOrLogicRequiresAnyGroup(t *testing.T) {
	nodes := filterLogicTestNodes()
	sub := Subcription{
		FilterLogic:       FilterLogicOr,
		CountryWhitelist:  "US",
		ProtocolWhitelist: "ss",
	}

	// us-vmess matches the country group, us-ss matches both groups,
	// jp-vmess matches neither and is filtered out.
	assertFilteredNodeNames(t, sub.ApplyFilters(nodes), []string{"us-vmess", "us-ss"})
}

func TestApplyFiltersOrLogicKeepsBlacklistAbsolute(t *testing.T) {
	nodes := filterLogicTestNodes()
	sub := Subcription{
		FilterLogic:       FilterLogicOr,
		CountryWhitelist:  "US",
		ProtocolWhitelist: "vmess",
		CountryBlacklist:  "US",
	}

	// Blacklists win over OR: US nodes are excluded even though they match,
	// while jp-vmess survives through the protocol group.
	assertFilteredNodeNames(t, sub.ApplyFilters(nodes), []string{"jp-vmess"})
}

func TestApplyFiltersOrLogicIncludesThresholdGroups(t *testing.T) {
	nodes := []Node{
		{Name: "us-slow", LinkCountry: "US", Protocol: "vmess", DelayTime: 500},
		{Name: "jp-fast", LinkCountry: "JP", Protocol: "ss", DelayTime: 100},
		{Name: "jp-slow", LinkCountry: "JP", Protocol: "ss", DelayTime: 500},
	}
	sub := Subcription{
		FilterLogic:      FilterLogicOr,
		CountryWhitelist: "US",
		DelayTime:        200,
	}

	// us-slow passes through the country group, jp-fast through the delay group,
	// jp-slow fails both.
	assertFilteredNodeNames(t, sub.ApplyFilters(nodes), []string{"us-slow", "jp-fast"})
}

func filterLogicTestNodes() []Node {
	return []Node{
		{Name: "us-vmess", LinkCountry: "US", Protocol: "vmess", DelayTime: 100},
		{Name: "us-ss", LinkCountry: "US", Protocol: "ss", DelayTime: 100},
		{Name: "jp-vmess", LinkCountry: "JP", Protocol: "vmess", DelayTime: 100},
	}
}

func assertFilteredNodeNames(t *testing.T, got []Node, want []string) {
	t.Helper()

	gotNames := nodeNames(got)
	if len(gotNames) != len(want) {
		t.Fatalf("filtered nodes = %v, want %v", gotNames, want)
	}
	for i := range want {
		if gotNames[i] != want[i] {
			t.Fatalf("filtered nodes = %v, want %v", gotNames, want)
		}
	}
}
