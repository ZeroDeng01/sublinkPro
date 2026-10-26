package models

import "testing"

func seedCountryStatsNodes() {
	resetNodeCacheForTest()
	for _, n := range []Node{
		{ID: 1, Name: "us-1", LinkName: "us-1", Link: "ss://us-1", LinkCountry: "US", Group: "Asia", Protocol: "ss"},
		{ID: 2, Name: "us-2", LinkName: "us-2", Link: "ss://us-2", LinkCountry: "US", Group: "Asia", Protocol: "ss"},
		{ID: 3, Name: "de-1", LinkName: "de-1", Link: "ss://de-1", LinkCountry: "DE", Group: "Asia", Protocol: "vmess"},
		{ID: 4, Name: "unknown-1", LinkName: "unknown-1", Link: "ss://unknown-1", LinkCountry: "", Group: "Asia", Protocol: "ss"},
		{ID: 5, Name: "sg-1", LinkName: "sg-1", Link: "ss://sg-1", LinkCountry: "SG", Group: "Europe", Protocol: "ss"},
	} {
		nodeCache.Set(n.ID, n)
	}
}

func assertCountryStats(t *testing.T, got []NodeCountryStat, want []NodeCountryStat) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("stats len = %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("stats[%d] = %+v, want %+v (full: %v)", i, got[i], want[i], got)
		}
	}
}

// TestGetFilteredCountryStatsAppliesFilters 验证国家统计与列表使用同一套过滤条件，
// 空国家代码归为未知地区并始终排在最后。
func TestGetFilteredCountryStatsAppliesFilters(t *testing.T) {
	seedCountryStatsNodes()
	t.Cleanup(resetNodeCacheForTest)

	var node Node

	stats, total := node.GetFilteredCountryStats(NodeFilter{Group: "Asia"})
	if total != 4 {
		t.Fatalf("group filter total = %d, want 4", total)
	}
	assertCountryStats(t, stats, []NodeCountryStat{
		{Country: "US", Count: 2},
		{Country: "DE", Count: 1},
		{Country: "", Count: 1},
	})

	stats, total = node.GetFilteredCountryStats(NodeFilter{Countries: []string{"US"}})
	if total != 2 {
		t.Fatalf("country filter total = %d, want 2", total)
	}
	assertCountryStats(t, stats, []NodeCountryStat{{Country: "US", Count: 2}})

	stats, total = node.GetFilteredCountryStats(NodeFilter{Protocol: "vmess"})
	if total != 1 {
		t.Fatalf("protocol filter total = %d, want 1", total)
	}
	assertCountryStats(t, stats, []NodeCountryStat{{Country: "DE", Count: 1}})

	stats, total = node.GetFilteredCountryStats(NodeFilter{Search: "sg-1"})
	if total != 1 {
		t.Fatalf("search filter total = %d, want 1", total)
	}
	assertCountryStats(t, stats, []NodeCountryStat{{Country: "SG", Count: 1}})

	stats, total = node.GetFilteredCountryStats(NodeFilter{})
	if total != 5 {
		t.Fatalf("unfiltered total = %d, want 5", total)
	}
	assertCountryStats(t, stats, []NodeCountryStat{
		{Country: "US", Count: 2},
		{Country: "DE", Count: 1},
		{Country: "SG", Count: 1},
		{Country: "", Count: 1},
	})
}
