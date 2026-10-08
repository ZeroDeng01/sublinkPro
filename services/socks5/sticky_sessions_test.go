package socks5

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"sublink/models"
)

func TestStickySessionTablePromotesAndExpiresLease(t *testing.T) {
	table := newStickySessionTable()
	now := time.Unix(100, 0)
	table.now = func() time.Time { return now }
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}}
	table.bind("client_ip:127.0.0.1", nodes[1], 30*time.Second)

	promoted, hit := table.promote("client_ip:127.0.0.1", nodes)
	if !hit || !reflect.DeepEqual([]int{promoted[0].ID, promoted[1].ID}, []int{2, 1}) {
		t.Fatalf("sticky lease was not promoted: hit=%v nodes=%+v", hit, promoted)
	}

	now = now.Add(31 * time.Second)
	promoted, hit = table.promote("client_ip:127.0.0.1", nodes)
	if hit || !reflect.DeepEqual(promoted, nodes) {
		t.Fatalf("expired sticky lease was reused: hit=%v nodes=%+v", hit, promoted)
	}
}

func TestStickySessionRefreshAtCapacityDoesNotEvictAnotherLease(t *testing.T) {
	table := newStickySessionTable()
	now := time.Unix(100, 0)
	table.now = func() time.Time { return now }
	node := models.Node{ID: 1, Link: "a"}
	for index := 0; index < maxStickySessionLeases; index++ {
		table.leases[fmt.Sprintf("client_ip:%d", index)] = stickySessionLease{nodeKey: adapterKey(node), expiresAt: now.Add(time.Hour)}
	}
	table.bind("client_ip:0", node, 2*time.Hour)
	if len(table.leases) != maxStickySessionLeases {
		t.Fatalf("refresh changed lease count: %d", len(table.leases))
	}
	if _, ok := table.leases["client_ip:1"]; !ok {
		t.Fatal("refresh evicted another lease")
	}
}

func TestStickySessionLeaseInvalidatesWhenNodeLinkChanges(t *testing.T) {
	table := newStickySessionTable()
	table.bind("client_ip:127.0.0.1", models.Node{ID: 1, Link: "a"}, time.Minute)
	nodes := []models.Node{{ID: 1, Link: "changed"}, {ID: 2, Link: "b"}}
	promoted, hit := table.promote("client_ip:127.0.0.1", nodes)
	if hit || promoted[0].ID != 1 {
		t.Fatalf("link-changed lease was reused: hit=%v nodes=%+v", hit, promoted)
	}
	if _, exists := table.leases["client_ip:127.0.0.1"]; exists {
		t.Fatal("link-changed lease was not removed")
	}
}

func TestNodeRouterStickyRoundRobinDoesNotAdvanceOnLeaseHit(t *testing.T) {
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}, {ID: 3, Link: "c"}}
	withCandidateNodes(t, nodes)
	router := newNodeRouter()
	cfg := Config{Selection: "round_robin", MaxAttempts: 3, StickySessionEnabled: true, StickySessionMode: "client_ip", StickySessionTTLSeconds: 1800}
	key := "client_ip:127.0.0.1"

	first, err := router.candidatesFor(cfg, key)
	if err != nil {
		t.Fatal(err)
	}
	router.sticky.bind(key, first[0], 30*time.Minute)
	second, err := router.candidatesFor(cfg, key)
	if err != nil {
		t.Fatal(err)
	}
	third, err := router.candidatesFor(cfg, "client_ip:127.0.0.2")
	if err != nil {
		t.Fatal(err)
	}
	if first[0].ID != 1 || second[0].ID != 1 || third[0].ID != 2 {
		t.Fatalf("unexpected sticky rotation: first=%v second=%v third=%v", first, second, third)
	}
}

func TestNodeRouterStickyLeaseRespectsHealthExclusion(t *testing.T) {
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}}
	withCandidateNodes(t, nodes)
	router := newNodeRouter()
	cfg := Config{Selection: "round_robin", MaxAttempts: 2, StickySessionEnabled: true, StickySessionMode: "client_ip", StickySessionTTLSeconds: 1800}
	key := "client_ip:127.0.0.1"
	router.sticky.bind(key, nodes[0], 30*time.Minute)
	router.health.recordProbeFailureReason(nodes[0], 0, "probe failed")

	candidates, err := router.candidatesFor(cfg, key)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) == 0 || candidates[0].ID != 2 {
		t.Fatalf("excluded sticky node remained first: %+v", candidates)
	}
}

func TestStickySessionKeyModes(t *testing.T) {
	cfg := Config{StickySessionEnabled: true, StickySessionMode: "client_ip", Selection: "best"}
	if got := stickySessionKey(cfg, "192.0.2.10", "alice"); got != "client_ip:192.0.2.10" {
		t.Fatalf("client IP key = %q", got)
	}
	cfg.StickySessionMode = "username"
	if got := stickySessionKey(cfg, "192.0.2.10", "alice"); got != "username:alice" {
		t.Fatalf("username key = %q", got)
	}
	cfg.Selection = "specific"
	if got := stickySessionKey(cfg, "192.0.2.10", "alice"); got != "" {
		t.Fatalf("specific selection should not create sticky key: %q", got)
	}
}
