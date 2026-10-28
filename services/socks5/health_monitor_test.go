package socks5

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"sublink/models"

	"github.com/metacubex/mihomo/constant"
)

func TestNormalizeConfigHealthCheckDefaultsAndValidation(t *testing.T) {
	cfg, err := NormalizeConfig(Config{ListenAddress: "127.0.0.1", Selection: "best"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HealthCheckEnabled || cfg.HealthCheckIntervalSeconds != 60 || cfg.HealthCheckTimeoutSeconds != 5 {
		t.Fatalf("unexpected health defaults: %+v", cfg)
	}
	if _, err := NormalizeConfig(Config{HealthCheckIntervalSeconds: 9}); err == nil {
		t.Fatal("expected health interval validation error")
	}
	if _, err := NormalizeConfig(Config{HealthCheckTimeoutSeconds: 31}); err == nil {
		t.Fatal("expected health timeout validation error")
	}
}

func TestNodeHealthUsesExponentialCooldownAndResetsOnSuccess(t *testing.T) {
	health := newNodeHealth()
	now := time.Unix(100, 0)
	health.now = func() time.Time { return now }
	node := models.Node{ID: 1, Name: "node", Link: "test://node"}
	health.recordFailureReason(node, 10*time.Second, "first")
	if got := health.cooldownUntil[adapterKey(node)]; !got.Equal(now.Add(10 * time.Second)) {
		t.Fatalf("first cooldown = %v", got)
	}
	now = now.Add(time.Second)
	health.recordFailureReason(node, 10*time.Second, "second")
	if got := health.cooldownUntil[adapterKey(node)]; !got.Equal(now.Add(20 * time.Second)) {
		t.Fatalf("second cooldown = %v", got)
	}
	snapshot := health.snapshot()
	if len(snapshot) != 1 || snapshot[0].ConsecutiveFailures != 2 || snapshot[0].LastError != "second" {
		t.Fatalf("failure snapshot = %+v", snapshot)
	}
	health.recordProbeSuccess(node, 42)
	snapshot = health.snapshot()
	if snapshot[0].Status != "healthy" || snapshot[0].LatencyMs != 42 || snapshot[0].ConsecutiveFailures != 0 || snapshot[0].CooldownUntil != nil {
		t.Fatalf("success snapshot = %+v", snapshot[0])
	}
}

func TestServerHealthSweepRecordsNodeResults(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Name: "good", Link: "test://good"}, {ID: 2, Name: "bad", Link: "test://bad"}})
	server, err := NewServer(Config{ListenAddress: "127.0.0.1", Selection: "best", FailureCooldownSeconds: 10}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		return nil, errors.New("unused")
	})
	if err != nil {
		t.Fatal(err)
	}
	server.healthProbe = func(_ context.Context, node models.Node, _ time.Duration) (int, error) {
		if node.ID == 2 {
			return 0, errors.New("probe failed")
		}
		return 25, nil
	}
	server.runHealthSweep(context.Background())
	snapshot := server.HealthSnapshot()
	if snapshot.HealthyNodes != 1 || snapshot.UnhealthyNodes != 1 || len(snapshot.Nodes) != 2 {
		t.Fatalf("health snapshot = %+v", snapshot)
	}
	if snapshot.LastSweepAt == nil || snapshot.Nodes[0].Status != "unhealthy" && snapshot.Nodes[1].Status != "unhealthy" {
		t.Fatalf("missing sweep state: %+v", snapshot)
	}
}

func TestHealthProbeDoesNotOverlap(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Link: "test://node"}})
	server, err := NewServer(Config{ListenAddress: "127.0.0.1", Selection: "best"}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		return nil, errors.New("unused")
	})
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	server.healthProbe = func(_ context.Context, _ models.Node, _ time.Duration) (int, error) {
		close(started)
		<-release
		return 1, nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server.setHealthContext(ctx)
	startedProbe, available := server.triggerHealthProbe()
	if !startedProbe || !available {
		t.Fatal("first probe did not start")
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("probe worker did not start")
	}
	startedProbe, available = server.triggerHealthProbe()
	if startedProbe || !available {
		t.Fatal("overlapping probe was allowed or reported unavailable")
	}
	close(release)
	deadline := time.Now().Add(time.Second)
	for server.HealthSnapshot().Running && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if server.HealthSnapshot().Running {
		t.Fatal("probe did not finish")
	}
}

func TestServeConnStopsHandshakeWhenContextIsCancelled(t *testing.T) {
	server, err := NewServer(Config{ListenAddress: "127.0.0.1", Selection: "best", RequireAuth: false}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		return nil, errors.New("unused")
	})
	if err != nil {
		t.Fatal(err)
	}
	client, serverConn := net.Pipe()
	defer func() { _ = client.Close(); _ = serverConn.Close() }()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		server.serveConn(ctx, serverConn)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handshake did not stop after context cancellation")
	}
}

func TestSanitizeProbeErrorRedactsSecretsAndPreservesUTF8(t *testing.T) {
	server := &Server{cfg: Config{Password: "secret-password"}}
	node := models.Node{Link: "socks5://user:pass@example.com"}
	message := server.sanitizeProbeError(node, errors.New(node.Link+" "+server.cfg.Password+" "+strings.Repeat("错", 300)))
	if strings.Contains(message, node.Link) || strings.Contains(message, server.cfg.Password) {
		t.Fatalf("health error leaked secret: %s", message)
	}
	if !utf8.ValidString(message) || len([]rune(message)) > maxHealthErrorLength {
		t.Fatalf("invalid bounded health error: %q", message)
	}
}

func TestServerCloseWaitsForHandshakeConnections(t *testing.T) {
	server, err := NewServer(Config{ListenAddress: "127.0.0.1", Selection: "best", RequireAuth: false}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		return nil, errors.New("unused")
	})
	if err != nil {
		t.Fatal(err)
	}
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	serveDone := make(chan struct{})
	go func() {
		_ = server.Serve(ctx, listener)
		close(serveDone)
	}()
	var dialer net.Dialer
	client, err := dialer.DialContext(context.Background(), "tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	deadline := time.Now().Add(time.Second)
	for server.Stats().ActiveConnections == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if server.Stats().ActiveConnections != 1 {
		t.Fatal("handshake connection was not registered")
	}
	cancel()
	_ = listener.Close()
	closeDone := make(chan struct{})
	go func() {
		server.Close()
		close(closeDone)
	}()
	select {
	case <-closeDone:
	case <-time.After(time.Second):
		t.Fatal("server close did not wait for and finish handshake connection")
	}
	select {
	case <-serveDone:
	case <-time.After(time.Second):
		t.Fatal("serve loop did not stop")
	}
	if server.Stats().ActiveConnections != 0 {
		t.Fatalf("active connections after close = %d", server.Stats().ActiveConnections)
	}
}

func TestHealthSnapshotBoundsNodeDetails(t *testing.T) {
	server, err := NewServer(Config{ListenAddress: "127.0.0.1", Selection: "best"}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		return nil, errors.New("unused")
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= healthSnapshotNodeLimit+5; i++ {
		server.router.health.recordProbeSuccess(models.Node{ID: i, Name: fmt.Sprintf("node-%03d", i), Link: fmt.Sprintf("test://%d", i)}, i)
	}
	snapshot := server.HealthSnapshot()
	if snapshot.TotalNodes != healthSnapshotNodeLimit+5 || len(snapshot.Nodes) != healthSnapshotNodeLimit || !snapshot.Truncated {
		t.Fatalf("unbounded health snapshot: total=%d returned=%d truncated=%v", snapshot.TotalNodes, len(snapshot.Nodes), snapshot.Truncated)
	}
}

func TestDefaultHealthProbeReleasesLeaseAfterPanic(t *testing.T) {
	proxy := &fakeProxy{}
	server, err := NewServer(Config{ListenAddress: "127.0.0.1", Selection: "best"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	server.pool = newAdapterPool(func(models.Node) (constant.Proxy, error) { return proxy, nil })
	if _, err := server.defaultHealthProbe(context.Background(), models.Node{ID: 1, Link: "test://panic"}, time.Second); err == nil {
		t.Fatal("expected recovered URLTest panic")
	}
	if proxy.closed.Load() != 1 {
		t.Fatalf("probe lease was not released after panic: close count=%d", proxy.closed.Load())
	}
}
