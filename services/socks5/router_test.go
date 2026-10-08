package socks5

import (
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"sublink/models"
)

func withCandidateNodes(t *testing.T, nodes []models.Node) {
	t.Helper()
	previous := listCandidateNodesFunc
	listCandidateNodesFunc = func(Config) ([]models.Node, error) {
		copyNodes := append([]models.Node(nil), nodes...)
		return copyNodes, nil
	}
	t.Cleanup(func() { listCandidateNodesFunc = previous })
}

func TestNormalizeConfigPhaseTwoValidation(t *testing.T) {
	cfg, err := NormalizeConfig(Config{ListenAddress: "127.0.0.1", Selection: "round_robin"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxAttempts != 1 || cfg.DialTimeoutSeconds != 30 || cfg.FailureCooldownSeconds != 0 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if _, err := NormalizeConfig(Config{MaxAttempts: 6}); err == nil {
		t.Fatal("expected max attempts validation error")
	}
	if _, err := NormalizeConfig(Config{DialTimeoutSeconds: 121}); err == nil {
		t.Fatal("expected dial timeout validation error")
	}
	if _, err := NormalizeConfig(Config{FailureCooldownSeconds: -1}); err == nil {
		t.Fatal("expected cooldown validation error")
	}
	if _, err := NormalizeConfig(Config{Enabled: true, ListenAddress: "0.0.0.0", RequireAuth: false}); err == nil {
		t.Fatal("expected open proxy protection error")
	}
	if _, err := NormalizeConfig(Config{Selection: "specific", NodeID: 1, SpecificFallback: true, MaxAttempts: 1}); err == nil {
		t.Fatal("expected specific fallback attempt validation error")
	}
	if smart, err := NormalizeConfig(Config{Selection: "smart"}); err != nil || smart.Selection != "smart" {
		t.Fatalf("smart selection was rejected: cfg=%+v err=%v", smart, err)
	}
}

func TestFilterCandidatePoolCombinesFields(t *testing.T) {
	nodes := []models.Node{
		{ID: 1, Link: "a", Group: "premium", Source: "airport-a", Protocol: "vless", LinkCountry: "JP"},
		{ID: 2, Link: "b", Group: "backup", Source: "airport-a", Protocol: "trojan", LinkCountry: "US"},
		{ID: 3, Link: "c", Group: "premium", Source: "airport-b", Protocol: "vless", LinkCountry: "JP"},
		{ID: 4, Link: "d", Group: "free", Source: "airport-a", Protocol: "vless", LinkCountry: "JP"},
		{ID: 5, Link: "e", Group: "premium", Source: "airport-a", Protocol: "ss", LinkCountry: "JP"},
		{ID: 6, Group: "premium", Source: "airport-a", Protocol: "vless", LinkCountry: "JP"},
	}
	filtered := filterCandidatePool(nodes, Config{
		CandidateGroups:    []string{"premium", "backup"},
		CandidateSources:   []string{"airport-a"},
		CandidateProtocols: []string{"VLESS", "trojan"},
		CandidateCountries: []string{"jp", "US"},
	})
	if len(filtered) != 2 || !reflect.DeepEqual([]int{filtered[0].ID, filtered[1].ID}, []int{1, 2}) {
		t.Fatalf("unexpected candidate pool: %+v", filtered)
	}
	manual := filterCandidatePool([]models.Node{{ID: 7, Link: "g"}}, Config{CandidateGroups: []string{candidateUngroupedValue}, CandidateSources: []string{"manual"}})
	if len(manual) != 1 {
		t.Fatalf("ungrouped manual node did not match candidate pool: %+v", manual)
	}
}

func TestBestCandidateNodeStaysInsideFilteredPool(t *testing.T) {
	nodes := []models.Node{
		{ID: 1, Link: "a", Group: "excluded", DelayTime: 5, Speed: 100},
		{ID: 2, Link: "b", Group: "pool", DelayTime: 30, Speed: 20},
		{ID: 3, Link: "c", Group: "pool", DelayTime: 60, Speed: 30},
	}
	filtered := filterCandidatePool(nodes, Config{CandidateGroups: []string{"pool"}})
	best := bestCandidateNode(filtered)
	if best == nil || best.ID != 2 {
		t.Fatalf("best candidate escaped filtered pool: %+v", best)
	}
}

func TestNodeRouterRoundRobin(t *testing.T) {
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}, {ID: 3, Link: "c"}}
	withCandidateNodes(t, nodes)
	router := newNodeRouter()
	cfg := Config{Selection: "round_robin", MaxAttempts: 3, DialTimeoutSeconds: 30}
	first, err := router.candidates(cfg)
	if err != nil {
		t.Fatal(err)
	}
	second, err := router.candidates(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]int{first[0].ID, second[0].ID}, []int{1, 2}) {
		t.Fatalf("unexpected rotation: first=%v second=%v", first, second)
	}
}

func TestNodeRouterRoundRobinScopesProfilesIndependently(t *testing.T) {
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}, {ID: 3, Link: "c"}}
	withCandidateNodes(t, nodes)
	router := newNodeRouter()
	cfg := Config{Selection: "round_robin", MaxAttempts: 3, DialTimeoutSeconds: 30}

	japanFirst, err := router.candidatesForScope(cfg, "", "japan")
	if err != nil {
		t.Fatal(err)
	}
	japanSecond, err := router.candidatesForScope(cfg, "", "japan")
	if err != nil {
		t.Fatal(err)
	}
	usFirst, err := router.candidatesForScope(cfg, "", "us")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]int{japanFirst[0].ID, japanSecond[0].ID, usFirst[0].ID}, []int{1, 2, 1}) {
		t.Fatalf("profile round robin was not isolated: japan=%v/%v us=%v", japanFirst, japanSecond, usFirst)
	}
}

func TestNodeRouterRoundRobinSkipsProbeExcludedNodesBeforeRotation(t *testing.T) {
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}, {ID: 3, Link: "c"}}
	withCandidateNodes(t, nodes)
	router := newNodeRouter()
	router.health.recordProbeFailureReason(nodes[0], 0, "probe failed")
	cfg := Config{Selection: "round_robin", MaxAttempts: 2, DialTimeoutSeconds: 30}

	first, err := router.candidates(cfg)
	if err != nil {
		t.Fatal(err)
	}
	second, err := router.candidates(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]int{first[0].ID, second[0].ID}, []int{2, 3}) {
		t.Fatalf("excluded node distorted rotation: first=%v second=%v", first, second)
	}
}

func TestNodeRouterSmartP2CPrefersLowerCompositeScore(t *testing.T) {
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}, {ID: 3, Link: "c"}}
	withCandidateNodes(t, nodes)
	router := newNodeRouter()
	router.health.recordProbeSuccess(nodes[0], 20)
	router.health.recordProbeSuccess(nodes[1], 50)
	router.health.recordProbeSuccess(nodes[2], 80)
	releases := make([]func(), 0, 5)
	for range 5 {
		releases = append(releases, router.runtime.start(nodes[0]))
	}
	defer func() {
		for _, release := range releases {
			release()
		}
	}()
	picks := []int{0, 0}
	router.randomIndex = func(_ int) int {
		pick := picks[0]
		picks = picks[1:]
		return pick
	}

	candidates, err := router.candidates(Config{Selection: "smart", MaxAttempts: 3})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]int{candidates[0].ID, candidates[1].ID, candidates[2].ID}, []int{2, 1, 3}) {
		t.Fatalf("unexpected smart order: %+v", candidates)
	}
}

func TestNodeRouterSmartAppliesFailurePenalty(t *testing.T) {
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}}
	withCandidateNodes(t, nodes)
	router := newNodeRouter()
	router.health.recordProbeSuccess(nodes[0], 30)
	router.health.recordProbeSuccess(nodes[1], 40)
	router.health.recordFailureReason(nodes[0], 0, "dial failed")
	router.randomIndex = func(int) int { return 0 }

	candidates, err := router.candidates(Config{Selection: "smart", MaxAttempts: 2})
	if err != nil {
		t.Fatal(err)
	}
	if candidates[0].ID != 2 {
		t.Fatalf("failure penalty was ignored: %+v", candidates)
	}
}

func TestNodeRouterSmartKeepsStickyHitFirstAndScoresFallbacks(t *testing.T) {
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}, {ID: 3, Link: "c"}}
	withCandidateNodes(t, nodes)
	router := newNodeRouter()
	router.health.recordProbeSuccess(nodes[0], 100)
	router.health.recordProbeSuccess(nodes[1], 20)
	router.health.recordProbeSuccess(nodes[2], 50)
	key := "client_ip:127.0.0.1"
	router.sticky.bind(key, nodes[0], time.Minute)

	candidates, err := router.candidatesFor(Config{Selection: "smart", MaxAttempts: 3, StickySessionEnabled: true}, key)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]int{candidates[0].ID, candidates[1].ID, candidates[2].ID}, []int{1, 2, 3}) {
		t.Fatalf("sticky smart order = %+v", candidates)
	}
}

func TestNodeRouterSpecificFallback(t *testing.T) {
	previous := listCandidateNodesFunc
	listCandidateNodesFunc = func(cfg Config) ([]models.Node, error) {
		if cfg.SpecificFallback {
			return []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}}, nil
		}
		return []models.Node{{ID: 1, Link: "a"}}, nil
	}
	t.Cleanup(func() { listCandidateNodesFunc = previous })
	router := newNodeRouter()
	strict, err := router.candidates(Config{Selection: "specific", NodeID: 1, MaxAttempts: 5, DialTimeoutSeconds: 30})
	if err != nil || len(strict) != 1 {
		t.Fatalf("strict candidates=%v err=%v", strict, err)
	}
	fallback, err := router.candidates(Config{Selection: "specific", NodeID: 1, MaxAttempts: 5, DialTimeoutSeconds: 30, SpecificFallback: true})
	if err != nil || len(fallback) != 2 {
		t.Fatalf("fallback candidates=%v err=%v", fallback, err)
	}
}

func TestNodeHealthCooldownAndProbe(t *testing.T) {
	health := newNodeHealth()
	now := time.Unix(100, 0)
	health.now = func() time.Time { return now }
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}}
	health.recordFailure(nodes[0], 10*time.Second)
	filtered := health.filter(nodes)
	if len(filtered) != 1 || filtered[0].ID != 2 {
		t.Fatalf("unexpected filtered nodes: %+v", filtered)
	}
	health.recordFailure(nodes[1], 20*time.Second)
	filtered = health.filter(nodes)
	if len(filtered) != 1 || filtered[0].ID != 1 {
		t.Fatalf("expected earliest cooldown probe, got %+v", filtered)
	}
	now = now.Add(11 * time.Second)
	filtered = health.filter(nodes)
	if len(filtered) != 1 || filtered[0].ID != 1 {
		t.Fatalf("expected expired node, got %+v", filtered)
	}
}

func TestNodeHealthPrefersUsableNodesAndFallsBackWhenAllUnhealthy(t *testing.T) {
	health := newNodeHealth()
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}}
	health.recordProbeFailureReason(nodes[0], 0, "probe failed")
	health.recordProbeSuccess(nodes[1], 25)

	filtered := health.filter(nodes)
	if len(filtered) != 1 || filtered[0].ID != 2 {
		t.Fatalf("expected healthy node only, got %+v", filtered)
	}

	health.recordProbeFailureReason(nodes[1], 0, "probe failed")
	filtered = health.filter(nodes)
	if len(filtered) != 2 {
		t.Fatalf("expected fail-open candidates when all nodes are unhealthy, got %+v", filtered)
	}
}

func TestNodeHealthPassiveFailureDoesNotPermanentlyExcludeNode(t *testing.T) {
	health := newNodeHealth()
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}}
	health.recordFailureReason(nodes[0], 0, "dial failed")
	health.recordProbeSuccess(nodes[1], 25)

	filtered := health.filter(nodes)
	if len(filtered) != 2 {
		t.Fatalf("passive failure unexpectedly excluded node: %+v", filtered)
	}
}

func TestNodeRouterBestUsesActiveHealthLatencyBeforeAttemptLimit(t *testing.T) {
	nodes := []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}, {ID: 3, Link: "c"}}
	withCandidateNodes(t, nodes)
	router := newNodeRouter()
	router.health.recordProbeSuccess(nodes[0], 300)
	router.health.recordProbeSuccess(nodes[1], 20)

	candidates, err := router.candidates(Config{Selection: "best", MaxAttempts: 2, DialTimeoutSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]int{candidates[0].ID, candidates[1].ID}, []int{2, 1}) {
		t.Fatalf("active health latency did not order candidates before truncation: %+v", candidates)
	}
}

func TestNodeHealthConnectionSuccessPreservesProbeLatency(t *testing.T) {
	health := newNodeHealth()
	node := models.Node{ID: 1, Link: "a"}
	health.recordProbeSuccess(node, 42)
	health.recordSuccess(node)

	snapshot := health.snapshot()
	if len(snapshot) != 1 || snapshot[0].Status != "healthy" || snapshot[0].LatencyMs != 42 {
		t.Fatalf("connection success cleared active probe latency: %+v", snapshot)
	}
}

func TestServerRetriesBeforeSuccessReply(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Link: "a"}, {ID: 2, Link: "b"}})
	var mu sync.Mutex
	attempts := make([]int, 0, 2)
	upstream, upstreamPeer := net.Pipe()
	defer func() { _ = upstreamPeer.Close() }()
	server, err := NewServer(Config{Selection: "best", RequireAuth: false, MaxAttempts: 2, DialTimeoutSeconds: 1, FailureCooldownSeconds: 30}, func(_ context.Context, node models.Node, _ string, _ uint16) (net.Conn, error) {
		mu.Lock()
		attempts = append(attempts, node.ID)
		mu.Unlock()
		if node.ID == 1 {
			return nil, errors.New("first failed")
		}
		return upstream, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	client, serverConn := net.Pipe()
	defer func() { _ = client.Close() }()
	go server.serveConn(context.Background(), serverConn)
	_, _ = client.Write([]byte{0x05, 0x01, 0x00})
	method := make([]byte, 2)
	_, _ = io.ReadFull(client, method)
	request := []byte{0x05, 0x01, 0x00, 0x03, 0x0b}
	request = append(request, []byte("example.com")...)
	request = append(request, 0x00, 0x50)
	_, _ = client.Write(request)
	reply := make([]byte, 10)
	if _, err := io.ReadFull(client, reply); err != nil {
		t.Fatal(err)
	}
	if reply[1] != 0x00 {
		t.Fatalf("unexpected reply: %x", reply)
	}
	mu.Lock()
	defer mu.Unlock()
	if !reflect.DeepEqual(attempts, []int{1, 2}) {
		t.Fatalf("attempts=%v", attempts)
	}
}
