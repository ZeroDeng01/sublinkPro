package socks5

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"sublink/database"
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

func TestGatewayMonitoringAcrossListenersAndIndependentResets(t *testing.T) {
	setupSocks5SettingsDB(t)
	db := database.DB
	if err := db.AutoMigrate(&models.Node{}); err != nil {
		t.Fatal(err)
	}
	node := models.Node{ID: 1, Name: "upstream", Link: "test://upstream", Group: "monitor", LinkHash: "monitor-fixture"}
	if err := db.Create(&node).Error; err != nil {
		t.Fatal(err)
	}
	if err := models.InitNodeCache(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Delete(&node).Error; err != nil {
			t.Error(err)
		}
		if err := models.InitNodeCache(); err != nil {
			t.Error(err)
		}
	})
	base := testBaseConfig(t)
	base.Enabled = true
	base.FailureCooldownSeconds = 60
	base.CandidateGroups = []string{"monitor"}
	if _, err := CreateRoutingProfile(base, RoutingProfile{ID: "shared", Name: "Shared node", Enabled: true, CandidateGroups: []string{"monitor"}}); err != nil {
		t.Fatal(err)
	}
	manager := &Manager{dial: func(_ context.Context, _ models.Node, host string, _ uint16) (net.Conn, error) {
		if host == "failure.invalid" {
			return nil, errors.New("controlled dial failure")
		}
		upstream, peer := net.Pipe()
		go func() { defer func() { _ = peer.Close() }(); _, _ = io.Copy(peer, peer) }()
		return upstream, nil
	}}
	listeners := []ListenerSpec{
		{ID: "a-first", Enabled: true, Protocol: ProtocolHTTP, ListenAddress: "127.0.0.1", Port: freePort(t)},
		{ID: "z-mixed", Enabled: true, Protocol: ProtocolMixed, ListenAddress: "127.0.0.1", Port: freePort(t), DefaultProfileID: "shared"},
	}
	if err := manager.ApplyListeners(base, listeners); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(manager.Stop)

	connect := func(port int, host, credentials, want string) (net.Conn, *bufio.Reader) {
		t.Helper()
		dialer := net.Dialer{Timeout: time.Second}
		conn, err := dialer.DialContext(t.Context(), "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = conn.Close() })
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if _, err := fmt.Fprintf(conn, "CONNECT %s:443 HTTP/1.1\r\nHost: %s:443\r\nProxy-Authorization: Basic %s\r\n\r\n", host, host, credentials); err != nil {
			t.Fatal(err)
		}
		reader := bufio.NewReader(conn)
		line, err := reader.ReadString('\n')
		if err != nil || !strings.Contains(line, want) {
			t.Fatalf("CONNECT response = %q, want %s, error: %v", line, want, err)
		}
		if _, err := textproto.NewReader(reader).ReadMIMEHeader(); err != nil {
			t.Fatal(err)
		}
		return conn, reader
	}
	exchange := func(conn net.Conn, reader *bufio.Reader, payload string) {
		t.Helper()
		if _, err := io.WriteString(conn, payload); err != nil {
			t.Fatal(err)
		}
		result := make([]byte, len(payload))
		if _, err := io.ReadFull(reader, result); err != nil || string(result) != payload {
			t.Fatalf("echo = %q, err = %v", result, err)
		}
	}
	nodeStats := func(listenerID string) NodeRoutingSnapshot {
		t.Helper()
		page, err := manager.RoutingSnapshot(RoutingSnapshotQuery{ListenerID: listenerID, AllProfiles: true, PageSize: 1})
		if err != nil || page.Total != 1 || len(page.Items) != 1 || page.Summary.SuccessfulConnections != page.Items[0].SuccessfulConnections {
			t.Fatalf("unexpected node snapshot: %+v, %v", page, err)
		}
		return page.Items[0]
	}

	const credentials = "cHJveHk6bGVnYWN5LXNlY3JldA==" // proxy:legacy-secret
	second, secondReader := connect(listeners[1].Port, "ok.invalid", credentials, " 200 ")
	exchange(second, secondReader, "before reset")
	if stats := nodeStats(""); stats.ActiveConnections != 1 || stats.SuccessfulConnections != 1 || stats.HealthListenerID != "z-mixed" {
		t.Fatalf("traffic on later listener missing from aggregate: %+v", stats)
	}
	if stats := nodeStats("a-first"); stats.ActiveConnections != 0 || stats.SuccessfulConnections != 0 {
		t.Fatalf("first listener counted traffic from the second: %+v", stats)
	}
	first, firstReader := connect(listeners[0].Port, "ok.invalid", credentials, " 200 ")
	exchange(first, firstReader, "before reset")
	connect(listeners[1].Port, "failure.invalid", credentials, " 502 ")
	connect(listeners[1].Port, "ok.invalid", "invalid", " 407 ")
	server := manager.runtimeSnapshot()[1].server
	waitForInboundActive(t, server, 1)
	server.healthProbe = func(_ context.Context, _ models.Node, _ time.Duration) (int, error) {
		return 0, errors.New("controlled probe failure")
	}
	started, available, err := manager.TriggerHealthProbe("z-mixed", "shared")
	if err != nil || !started || !available {
		t.Fatalf("could not start scoped probe: started %t, available %t, err %v", started, available, err)
	}
	deadline := time.Now().Add(time.Second)
	for server.probeSchedule().Running && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if server.probeSchedule().Running || manager.runtimeSnapshot()[0].server.probeSchedule().LastSweepAt != nil {
		t.Fatal("probe did not finish on just the selected listener")
	}
	if stats := nodeStats(""); stats.ActiveConnections != 2 || stats.SuccessfulConnections != 2 || stats.FailedConnections != 1 {
		t.Fatalf("unexpected merged dial counters: %+v", stats)
	}
	if stats := manager.GatewaySnapshot(); stats.FailedConnections != 2 || stats.SuccessfulConnections != 2 || stats.TotalConnections != 4 {
		t.Fatalf("unexpected client session counters: %+v", stats)
	}
	if health := manager.HealthSnapshot(); health.TotalNodes != 1 || health.HealthyNodes != 1 || health.Nodes[0].ListenerID != "a-first" {
		t.Fatalf("unexpected aggregate health: %+v", health)
	}
	manager.ResetNodeRuntimeStats()
	if stats := nodeStats("z-mixed"); stats.ActiveConnections != 1 || stats.SuccessfulConnections != 0 || stats.FailedConnections != 0 || stats.LastSelectedAt != nil || stats.LastError != "controlled probe failure" || stats.CooldownUntil == nil {
		t.Fatalf("node reset affected active or health state: %+v", stats)
	}
	if stats := manager.GatewaySnapshot(); stats.TotalConnections != 4 || stats.DownloadBytes == 0 {
		t.Fatalf("node reset affected gateway totals: %+v", stats)
	}
	if !manager.ResetGatewayStats() {
		t.Fatal("expected live gateway reset to succeed")
	}
	if stats := manager.GatewaySnapshot(); stats != (GatewaySnapshot{ActiveConnections: 2}) {
		t.Fatalf("gateway reset retained historical totals: %+v", stats)
	}
	exchange(second, secondReader, "after")
	_ = second.Close()
	waitInboundIdle(t, server)
	if stats := manager.GatewaySnapshot(); stats.ActiveConnections != 1 || stats.UploadBytes != 5 || stats.DownloadBytes != 5 {
		t.Fatalf("closing an existing session restored pre-reset bytes: %+v", stats)
	}
}

func waitForInboundActive(t *testing.T, server *Server, count int64) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for server.Stats().ActiveConnections != count && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := server.Stats().ActiveConnections; got != count {
		t.Fatalf("active connections = %d, want %d", got, count)
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
