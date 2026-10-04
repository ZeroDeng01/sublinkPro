package socks5

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"

	"sublink/models"
)

func TestRoutingSnapshotCombinesHealthLoadAndScore(t *testing.T) {
	nodes := []models.Node{
		{ID: 1, Name: "JP primary", Link: "test://one", Group: "premium", Source: "airport-a", Protocol: "vless", LinkCountry: "JP", DelayTime: 80},
		{ID: 2, Name: "US backup", Link: "test://two", Group: "backup", Source: "airport-b", Protocol: "trojan", LinkCountry: "US", DelayTime: 150},
		{ID: 3, Name: "Unknown", Link: "test://three"},
	}
	withCandidateNodes(t, nodes)
	server, err := NewServer(Config{Selection: "smart", MaxAttempts: 3}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1000, 0)
	server.router.health.now = func() time.Time { return now }
	server.router.runtime.now = func() time.Time { return now }
	server.router.health.recordProbeSuccess(nodes[0], 20)
	releaseOne := server.router.runtime.start(nodes[0])
	releaseTwo := server.router.runtime.start(nodes[0])
	defer releaseOne()
	defer releaseTwo()
	server.router.health.recordFailureReason(nodes[1], time.Minute, "dial failed")
	server.router.runtime.recordFailure(nodes[1])

	snapshot, err := server.RoutingSnapshot(RoutingSnapshotQuery{SortBy: "activeConnections", SortOrder: "desc", PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Total != 3 || snapshot.Page != 1 || snapshot.PageSize != 2 || snapshot.TotalPages != 2 || len(snapshot.Items) != 2 {
		t.Fatalf("unexpected page: %+v", snapshot)
	}
	if snapshot.Items[0].NodeID != 1 || snapshot.Items[0].LatencySource != "probe" || snapshot.Items[0].LatencyMs != 20 || snapshot.Items[0].ActiveConnections != 2 || snapshot.Items[0].SmartScore != 40 {
		t.Fatalf("unexpected active node snapshot: %+v", snapshot.Items[0])
	}
	if snapshot.Summary.TotalNodes != 3 || snapshot.Summary.ActiveNodes != 1 || snapshot.Summary.ActiveConnections != 2 || snapshot.Summary.SuccessfulConnections != 2 || snapshot.Summary.FailedConnections != 1 || snapshot.Summary.CoolingNodes != 1 {
		t.Fatalf("unexpected summary: %+v", snapshot.Summary)
	}

	filtered, err := server.RoutingSnapshot(RoutingSnapshotQuery{Keyword: "airport-b", Status: "cooling", SortBy: "smartScore", PageSize: 25})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered.Items) != 1 || filtered.Items[0].NodeID != 2 || filtered.Items[0].LatencySource != "stored" || filtered.Items[0].SmartScore != 250 {
		t.Fatalf("unexpected filtered snapshot: %+v", filtered)
	}

	fallback, err := server.RoutingSnapshot(RoutingSnapshotQuery{Keyword: "Unknown", PageSize: 25})
	if err != nil {
		t.Fatal(err)
	}
	if len(fallback.Items) != 1 || fallback.Items[0].LatencySource != "fallback" || fallback.Items[0].LatencyMs != 500 {
		t.Fatalf("unexpected fallback latency snapshot: %+v", fallback)
	}
}

func TestRoutingSnapshotSortsAndClampsPagination(t *testing.T) {
	nodes := []models.Node{{ID: 1, Name: "charlie", Link: "a"}, {ID: 2, Name: "alpha", Link: "b"}, {ID: 3, Name: "bravo", Link: "c"}}
	withCandidateNodes(t, nodes)
	server, err := NewServer(Config{}, func(context.Context, models.Node, string, uint16) (net.Conn, error) { return nil, nil })
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := server.RoutingSnapshot(RoutingSnapshotQuery{SortBy: "nodeName", SortOrder: "asc", Page: 99, PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Page != 2 || snapshot.TotalPages != 2 || !reflect.DeepEqual([]int{snapshot.Items[0].NodeID}, []int{1}) {
		t.Fatalf("unexpected sorted last page: %+v", snapshot)
	}
}
