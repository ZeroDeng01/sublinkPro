package socks5

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"sublink/models"
)

func TestNormalizeConfigDefaultsAndValidation(t *testing.T) {
	cfg, err := NormalizeConfig(Config{Enabled: true, Username: "user", Password: "pass", RequireAuth: true})
	if err != nil {
		t.Fatalf("normalize config: %v", err)
	}
	if cfg.ListenAddress != defaultListenAddress || cfg.Port != defaultPort || cfg.Selection != defaultSelection || !cfg.RequireAuth {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if _, err := NormalizeConfig(Config{Enabled: true, RequireAuth: true, Username: "user"}); err == nil {
		t.Fatal("expected missing password validation error")
	}
	pool, err := NormalizeConfig(Config{
		CandidateGroups:    []string{" premium ", "PREMIUM", "backup"},
		CandidateSources:   []string{" manual ", "manual"},
		CandidateProtocols: []string{" VLESS ", "Trojan"},
		CandidateCountries: []string{" jp ", "US"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pool.CandidateGroups) != 2 || pool.CandidateGroups[0] != "premium" || pool.CandidateProtocols[0] != "vless" || pool.CandidateCountries[0] != "JP" {
		t.Fatalf("candidate pool was not normalized: %+v", pool)
	}
	if pool.StickySessionMode != "client_ip" || pool.StickySessionTTLSeconds != 1800 {
		t.Fatalf("unexpected sticky defaults: %+v", pool)
	}
	if _, err := NormalizeConfig(Config{StickySessionEnabled: true, StickySessionMode: "username", StickySessionTTLSeconds: 300, RequireAuth: false}); err == nil {
		t.Fatal("expected username sticky sessions to require authentication")
	}
	if _, err := NormalizeConfig(Config{StickySessionMode: "invalid", StickySessionTTLSeconds: 300}); err == nil {
		t.Fatal("expected sticky mode validation error")
	}
	if _, err := NormalizeConfig(Config{StickySessionTTLSeconds: 59}); err == nil {
		t.Fatal("expected sticky TTL validation error")
	}
}

func TestServerHandlesAuthenticatedConnectAndPumpsTraffic(t *testing.T) {
	previous := listCandidateNodesFunc
	listCandidateNodesFunc = func(Config) ([]models.Node, error) {
		return []models.Node{{ID: 1, Link: "socks5://upstream:1080", Name: "test"}}, nil
	}
	t.Cleanup(func() { listCandidateNodesFunc = previous })

	client, serverConn := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = serverConn.Close() }()

	upstream, upstreamPeer := net.Pipe()
	defer func() { _ = upstreamPeer.Close() }()
	defer func() { _ = upstream.Close() }()

	server, err := NewServer(Config{Enabled: true, Username: "alice", Password: "secret", RequireAuth: true}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		return upstream, nil
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	done := make(chan struct{})
	go func() {
		server.serveConn(context.Background(), serverConn)
		close(done)
	}()

	if _, err := client.Write([]byte{0x05, 0x01, 0x02}); err != nil {
		t.Fatalf("write greeting: %v", err)
	}
	response := make([]byte, 2)
	if _, err := io.ReadFull(client, response); err != nil {
		t.Fatalf("read method response: %v", err)
	}
	if string(response) != string([]byte{0x05, 0x02}) {
		t.Fatalf("unexpected method response: %x", response)
	}
	if _, err := client.Write([]byte{0x01, 0x05, 'a', 'l', 'i', 'c', 'e', 0x06, 's', 'e', 'c', 'r', 'e', 't'}); err != nil {
		t.Fatalf("write auth: %v", err)
	}
	if _, err := io.ReadFull(client, response); err != nil {
		t.Fatalf("read auth response: %v", err)
	}
	if string(response) != string([]byte{0x01, 0x00}) {
		t.Fatalf("unexpected auth response: %x", response)
	}
	request := []byte{0x05, 0x01, 0x00, 0x03, 0x0b}
	request = append(request, []byte("example.com")...)
	request = append(request, 0x01, 0xbb)
	if _, err := client.Write(request); err != nil {
		t.Fatalf("write connect request: %v", err)
	}
	reply := make([]byte, 10)
	if _, err := io.ReadFull(client, reply); err != nil {
		t.Fatalf("read connect reply: %v", err)
	}
	if reply[1] != 0x00 {
		t.Fatalf("connect failed: %x", reply)
	}

	payload := []byte("hello through proxy")
	if _, err := client.Write(payload); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	_ = upstreamPeer.SetReadDeadline(time.Now().Add(time.Second))
	received := make([]byte, len(payload))
	if _, err := io.ReadFull(upstreamPeer, received); err != nil {
		t.Fatalf("read upstream payload: %v", err)
	}
	if string(received) != string(payload) {
		t.Fatalf("unexpected payload: %q", received)
	}
	if _, err := upstreamPeer.Write([]byte("pong")); err != nil {
		t.Fatalf("write upstream response: %v", err)
	}
	_ = client.SetReadDeadline(time.Now().Add(time.Second))
	responsePayload := make([]byte, 4)
	if _, err := io.ReadFull(client, responsePayload); err != nil {
		t.Fatalf("read client response: %v", err)
	}
	if string(responsePayload) != "pong" {
		t.Fatalf("unexpected client response: %q", responsePayload)
	}
	_ = client.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("server connection did not exit")
	}
}

func TestServerRuntimeStatsFollowConnectedSessionLifecycle(t *testing.T) {
	node := models.Node{ID: 1, Link: "test://node", Name: "node"}
	withCandidateNodes(t, []models.Node{node})
	upstream, upstreamPeer := net.Pipe()
	defer func() { _ = upstreamPeer.Close() }()
	server, err := NewServer(Config{Selection: "smart", RequireAuth: false, MaxAttempts: 1}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		return upstream, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	client, serverConn := net.Pipe()
	done := make(chan struct{})
	go func() {
		server.serveConn(context.Background(), serverConn)
		close(done)
	}()

	if _, err := client.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		t.Fatal(err)
	}
	method := make([]byte, 2)
	if _, err := io.ReadFull(client, method); err != nil {
		t.Fatal(err)
	}
	request := []byte{0x05, 0x01, 0x00, 0x03, 0x0b}
	request = append(request, []byte("example.com")...)
	request = append(request, 0x00, 0x50)
	if _, err := client.Write(request); err != nil {
		t.Fatal(err)
	}
	reply := make([]byte, 10)
	if _, err := io.ReadFull(client, reply); err != nil {
		t.Fatal(err)
	}
	if reply[1] != 0x00 {
		t.Fatalf("unexpected reply: %x", reply)
	}
	state := server.router.runtime.metrics(node)
	if state.ActiveConnections != 1 || state.SuccessfulConnections != 1 {
		t.Fatalf("unexpected connected runtime state: %+v", state)
	}

	_ = client.Close()
	_ = upstreamPeer.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("session did not close")
	}
	state = server.router.runtime.metrics(node)
	if state.ActiveConnections != 0 || state.SuccessfulConnections != 1 {
		t.Fatalf("runtime state was not released: %+v", state)
	}
}

func TestServerRejectsUnsupportedCommand(t *testing.T) {
	client, serverConn := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = serverConn.Close() }()
	server, err := NewServer(Config{Enabled: true, Username: "alice", Password: "secret", RequireAuth: true}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		t.Fatal("dial should not be called")
		return nil, nil
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	done := make(chan struct{})
	go func() { server.serveConn(context.Background(), serverConn); close(done) }()
	_, _ = client.Write([]byte{0x05, 0x01, 0x02})
	response := make([]byte, 2)
	_, _ = io.ReadFull(client, response)
	_, _ = client.Write([]byte{0x01, 0x05, 'a', 'l', 'i', 'c', 'e', 0x06, 's', 'e', 'c', 'r', 'e', 't'})
	_, _ = io.ReadFull(client, response)
	_, _ = client.Write([]byte{0x05, 0x02, 0x00, 0x01})
	reply := make([]byte, 10)
	if _, err := io.ReadFull(client, reply); err != nil {
		t.Fatalf("read reject reply: %v", err)
	}
	if reply[1] != 0x07 {
		t.Fatalf("expected command-not-supported reply, got %x", reply)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("server connection did not exit")
	}
}

func TestServerRoutesProfileUsernameThroughProfileCandidatePool(t *testing.T) {
	previous := listCandidateNodesFunc
	selectedCountry := make(chan string, 1)
	listCandidateNodesFunc = func(cfg Config) ([]models.Node, error) {
		country := "default"
		if len(cfg.CandidateCountries) > 0 {
			country = cfg.CandidateCountries[0]
		}
		selectedCountry <- country
		return []models.Node{{ID: 1, Name: country + " node", Link: "test://" + country}}, nil
	}
	t.Cleanup(func() { listCandidateNodesFunc = previous })

	upstream, upstreamPeer := net.Pipe()
	defer func() { _ = upstreamPeer.Close() }()
	server, err := NewServer(Config{
		Enabled: true, ListenAddress: "127.0.0.1", Username: "proxy", Password: "secret", RequireAuth: true,
		Selection: "best", MaxAttempts: 1,
	}, func(context.Context, models.Node, string, uint16) (net.Conn, error) { return upstream, nil })
	if err != nil {
		t.Fatal(err)
	}
	profile, err := normalizeRoutingProfile(server.cfg, RoutingProfile{
		ID: "japan", Name: "Japan", Enabled: true, Selection: "smart", MaxAttempts: 2,
		CandidateCountries: []string{"JP"}, StickySessionEnabled: true, StickySessionTTLSeconds: 900,
	})
	if err != nil {
		t.Fatal(err)
	}
	server.profiles.replace(server.cfg, []RoutingProfile{defaultRoutingProfile(server.cfg), profile})

	client, serverConn := net.Pipe()
	done := make(chan struct{})
	go func() {
		server.serveConn(context.Background(), serverConn)
		close(done)
	}()
	if _, err := client.Write([]byte{0x05, 0x01, 0x02}); err != nil {
		t.Fatal(err)
	}
	method := make([]byte, 2)
	if _, err := io.ReadFull(client, method); err != nil {
		t.Fatal(err)
	}
	username := []byte("proxy@japan.user01")
	password := []byte("secret")
	auth := []byte{0x01, byte(len(username))}
	auth = append(auth, username...)
	auth = append(auth, byte(len(password)))
	auth = append(auth, password...)
	if _, err := client.Write(auth); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(client, method); err != nil || method[1] != 0x00 {
		t.Fatalf("profile auth failed: response=%x err=%v", method, err)
	}
	request := []byte{0x05, 0x01, 0x00, 0x03, 0x0b}
	request = append(request, []byte("example.com")...)
	request = append(request, 0x00, 0x50)
	if _, err := client.Write(request); err != nil {
		t.Fatal(err)
	}
	reply := make([]byte, 10)
	if _, err := io.ReadFull(client, reply); err != nil || reply[1] != 0x00 {
		t.Fatalf("profile connect failed: reply=%x err=%v", reply, err)
	}
	select {
	case country := <-selectedCountry:
		if country != "JP" {
			t.Fatalf("profile candidate pool country = %q", country)
		}
	case <-time.After(time.Second):
		t.Fatal("profile candidate selection was not observed")
	}
	connections := server.Connections()
	if len(connections) != 1 || connections[0].ProfileID != "japan" || connections[0].ProfileName != "Japan" || connections[0].Account != "user01" {
		t.Fatalf("profile connection metadata = %+v", connections)
	}
	_ = client.Close()
	_ = upstreamPeer.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("profile session did not close")
	}
}
