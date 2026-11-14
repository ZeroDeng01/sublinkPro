package models

import (
	"encoding/json"
	"testing"

	"sublink/database"
)

func TestGetAirportNodeStatsTestEvidence(t *testing.T) {
	const checkedAt = "2026-11-13 16:36:02"
	cases := []struct {
		name      string
		nodes     []Node
		tested    bool
		delayPass int
		speedPass int
	}{
		{name: "zero nodes"},
		{name: "never tested", nodes: []Node{{DelayStatus: "untested", SpeedStatus: "untested"}}},
		{name: "all timeout with timestamps", nodes: []Node{
			{DelayStatus: "timeout", DelayTime: -1, LatencyCheckAt: checkedAt},
			{DelayStatus: "timeout", DelayTime: -1, LatencyCheckAt: checkedAt},
		}, tested: true},
		{name: "partial success", nodes: []Node{
			{DelayStatus: "success", DelayTime: 42, LatencyCheckAt: checkedAt, SpeedStatus: "success", Speed: 5, SpeedCheckAt: checkedAt},
			{DelayStatus: "timeout", DelayTime: -1, LatencyCheckAt: checkedAt},
		}, tested: true, delayPass: 1, speedPass: 1},
		{name: "latency only", nodes: []Node{{DelayStatus: "success", DelayTime: 42, LatencyCheckAt: checkedAt, SpeedStatus: "untested"}}, tested: true, delayPass: 1},
		{name: "speed failure only", nodes: []Node{{SpeedStatus: "error", Speed: -1, SpeedCheckAt: checkedAt}}, tested: true},
		{name: "legacy terminal status", nodes: []Node{{DelayStatus: "timeout", DelayTime: -1}}, tested: true},
		{name: "timestamp without terminal status", nodes: []Node{{LatencyCheckAt: checkedAt}}, tested: true},
		{name: "new nodes alongside checked nodes", nodes: []Node{
			{DelayStatus: "timeout", LatencyCheckAt: checkedAt},
			{DelayStatus: "untested", SpeedStatus: "untested"},
		}, tested: true},
		{name: "replacement nodes have no history", nodes: []Node{{}, {}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setupNodeInfoBatchTestDB(t)
			for i, node := range tc.nodes {
				node.ID = i + 1
				node.SourceID = 123
				nodeCache.Set(node.ID, node)
			}
			// Exercise the list response's JSON contract, including zero-valued counts.
			data, err := json.Marshal(GetAirportNodeStats(123))
			if err != nil {
				t.Fatal(err)
			}
			var got struct {
				HasTestedNodes bool `json:"hasTestedNodes"`
				DelayPassCount int  `json:"delayPassCount"`
				SpeedPassCount int  `json:"speedPassCount"`
			}
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if got.HasTestedNodes != tc.tested || got.DelayPassCount != tc.delayPass || got.SpeedPassCount != tc.speedPass {
				t.Fatalf("stats = %s; want tested=%v delayPass=%d speedPass=%d", data, tc.tested, tc.delayPass, tc.speedPass)
			}
		})
	}
}

func TestGetAirportNodeStatsUsesCurrentNodes(t *testing.T) {
	setupNodeInfoBatchTestDB(t)
	hasTestedNodes := func(stats AirportNodeStats) bool {
		t.Helper()
		data, err := json.Marshal(stats)
		if err != nil {
			t.Fatal(err)
		}
		var payload struct {
			HasTestedNodes bool `json:"hasTestedNodes"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatal(err)
		}
		return payload.HasTestedNodes
	}
	checked := createNodeForBatchUpdateTest(t, Node{
		SourceID: 123, Link: "ss://checked", Name: "checked",
		DelayStatus: "timeout", DelayTime: -1, LatencyCheckAt: "2026-11-13 16:36:02",
	})
	createNodeForBatchUpdateTest(t, Node{
		SourceID: 456, Link: "ss://other-airport", Name: "other",
		DelayStatus: "success", DelayTime: 42,
	})
	// The database fallback and cache path must expose the same evidence.
	resetNodeCacheForTest()
	if !hasTestedNodes(GetAirportNodeStats(123)) {
		t.Fatal("database fallback lost failed check history")
	}
	nodeCache.Set(checked.ID, checked)
	createNodeForBatchUpdateTest(t, Node{SourceID: 123, Link: "ss://new", Name: "new"})
	if !hasTestedNodes(GetAirportNodeStats(123)) {
		t.Fatal("adding an untested node hid existing check history")
	}
	// Replacing/removing the checked node must not leave airport-level history behind.
	if err := database.DB.Delete(&checked).Error; err != nil {
		t.Fatal(err)
	}
	nodeCache.Delete(checked.ID)
	stats := GetAirportNodeStats(123)
	if hasTestedNodes(stats) || stats.DelayPassCount != 0 || stats.SpeedPassCount != 0 {
		t.Fatalf("removed/other-airport nodes contributed to stats: %+v", stats)
	}
}
