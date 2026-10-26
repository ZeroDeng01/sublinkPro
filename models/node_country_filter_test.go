package models

import "testing"

func seedCountryFilterNodes() {
	resetNodeCacheForTest()
	for _, n := range []Node{
		{ID: 1, Name: "us-1", LinkName: "us-1", Link: "ss://us-1", LinkCountry: "US"},
		{ID: 2, Name: "de-1", LinkName: "de-1", Link: "ss://de-1", LinkCountry: "DE"},
		{ID: 3, Name: "unknown-1", LinkName: "unknown-1", Link: "ss://unknown-1", LinkCountry: ""},
	} {
		nodeCache.Set(n.ID, n)
	}
}

// TestListWithFiltersUnknownCountry 验证 UnknownCountryFilter 特殊值可以筛选未识别国家的节点，
// 并与国家代码按“或”语义组合。
func TestListWithFiltersUnknownCountry(t *testing.T) {
	seedCountryFilterNodes()
	t.Cleanup(resetNodeCacheForTest)

	var node Node

	nodes, err := node.ListWithFilters(NodeFilter{Countries: []string{UnknownCountryFilter}})
	if err != nil {
		t.Fatalf("unknown filter error = %v", err)
	}
	if len(nodes) != 1 || nodes[0].LinkCountry != "" {
		t.Fatalf("unknown filter = %v, want single node without country", nodes)
	}

	nodes, err = node.ListWithFilters(NodeFilter{Countries: []string{"US", UnknownCountryFilter}})
	if err != nil {
		t.Fatalf("combined filter error = %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("combined filter len = %d, want 2", len(nodes))
	}
	seen := map[string]bool{}
	for _, n := range nodes {
		seen[n.LinkCountry] = true
	}
	if !seen["US"] || !seen[""] {
		t.Fatalf("combined filter countries = %v, want US and unknown", seen)
	}

	nodes, err = node.ListWithFilters(NodeFilter{Countries: []string{"US"}})
	if err != nil {
		t.Fatalf("known filter error = %v", err)
	}
	if len(nodes) != 1 || nodes[0].LinkCountry != "US" {
		t.Fatalf("known filter = %v, want single US node", nodes)
	}
}
