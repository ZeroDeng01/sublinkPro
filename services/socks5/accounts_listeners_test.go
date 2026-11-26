package socks5

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
)

func setupSocks5SettingsDB(t *testing.T) {
	t.Helper()
	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized
	db := testutil.OpenMemoryDB(t, "socks5_settings")
	if err := db.AutoMigrate(&models.SystemSetting{}); err != nil {
		t.Fatalf("auto migrate settings: %v", err)
	}
	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	if err := models.InitSettingCache(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SUBLINK_API_ENCRYPTION_KEY", "socks5-test-key-0123456789abcdef0123456789")
	t.Cleanup(func() {
		defaultManager.Stop()
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		if oldDB != nil {
			_ = models.InitSettingCache()
		}
		testutil.CloseDB(t, db)
	})
}

func testBaseConfig(t *testing.T) Config {
	t.Helper()
	cfg, err := NormalizeConfig(Config{
		ListenAddress: "127.0.0.1", Port: 1080, Username: "proxy", Password: "legacy-secret",
		RequireAuth: true, Selection: "best",
	})
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestSavedListenerControlsAuthValidationAfterEditingDefault(t *testing.T) {
	setupSocks5SettingsDB(t)
	base := testBaseConfig(t)
	base.Enabled = true
	base.ListenAddress = "0.0.0.0"

	saved, err := SaveConfig(base)
	if err != nil {
		t.Fatal(err)
	}
	listeners, err := LoadListeners(saved)
	if err != nil || len(listeners) != 1 || listeners[0].ID != "default" || listeners[0].ListenAddress != base.ListenAddress {
		t.Fatalf("expected the original address in the listener list: %+v, %v", listeners, err)
	}

	base.RequireAuth = false
	if _, err := SaveConfig(base); err == nil {
		t.Fatal("an enabled non-loopback default listener must require authentication")
	}

	listener := listeners[0]
	listener.ListenAddress = "127.0.0.1"
	listener.Port = freePort(t)
	if _, err := UpdateListener(saved, listener.ID, listener); err != nil {
		t.Fatal(err)
	}
	saved, err = SaveConfig(base)
	if err != nil || saved.RequireAuth {
		t.Fatalf("could not disable shared authentication for a loopback listener: %v", err)
	}

	manager := &Manager{}
	t.Cleanup(manager.Stop)
	if err := manager.Apply(saved); err != nil {
		t.Fatal(err)
	}
	statuses := manager.ListenerStatuses()
	if len(statuses) != 1 || !statuses[0].Running || statuses[0].ListenAddress != listener.ListenAddress || statuses[0].Port != listener.Port {
		t.Fatalf("expected the edited listener to be running: %+v", statuses)
	}
	if statuses[0].RequireAuth != nil || statuses[0].EffectiveRequireAuth == nil || *statuses[0].EffectiveRequireAuth {
		t.Fatalf("expected the running listener to inherit disabled authentication: %+v", statuses[0])
	}
}

func authenticateForTest(t *testing.T, server *Server, username, password string) (routingIdentity, error) {
	t.Helper()
	client, serverConn := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = serverConn.Close() }()
	type result struct {
		identity routingIdentity
		err      error
	}
	done := make(chan result, 1)
	go func() {
		identity, err := server.authenticate(serverConn)
		done <- result{identity: identity, err: err}
	}()
	if _, err := client.Write([]byte{0x05, 0x01, 0x02}); err != nil {
		t.Fatal(err)
	}
	response := make([]byte, 2)
	if _, err := io.ReadFull(client, response); err != nil {
		t.Fatal(err)
	}
	request := []byte{0x01, byte(len(username))}
	request = append(request, username...)
	request = append(request, byte(len(password)))
	request = append(request, password...)
	if _, err := client.Write(request); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(client, response); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-done:
		return got.identity, got.err
	case <-time.After(time.Second):
		t.Fatal("authentication did not complete")
		return routingIdentity{}, context.DeadlineExceeded
	}
}

func TestAccountCRUDEncryptsPasswordsAndAuthenticatesIndependently(t *testing.T) {
	setupSocks5SettingsDB(t)
	base := testBaseConfig(t)
	created, err := CreateAccount(base, Account{ID: "alice", Username: "alice", Password: "alice-secret", ProfileID: "default", Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if !created.HasPassword || created.MaskedPassword == "" {
		t.Fatalf("unexpected public account: %+v", created)
	}
	stored, err := models.GetSetting(settingAccounts)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stored, "alice-secret") {
		t.Fatalf("stored account leaked plaintext password: %s", stored)
	}
	var encoded []map[string]any
	if err := json.Unmarshal([]byte(stored), &encoded); err != nil || len(encoded) != 1 {
		t.Fatalf("invalid stored account JSON: %v %s", err, stored)
	}

	server, err := NewServer(base, func(context.Context, models.Node, string, uint16) (net.Conn, error) { return nil, nil })
	if err != nil {
		t.Fatal(err)
	}
	identity, err := authenticateForTest(t, server, "alice", "alice-secret")
	if err != nil || identity.ProfileID != "default" || identity.Account != "alice" {
		t.Fatalf("independent account identity=%+v err=%v", identity, err)
	}
	if _, err := authenticateForTest(t, server, "alice", "legacy-secret"); err == nil {
		t.Fatal("independent account accepted the legacy password")
	}
	legacy, err := authenticateForTest(t, server, "proxy", "legacy-secret")
	if err != nil || legacy.ProfileID != "default" {
		t.Fatalf("legacy credentials stopped working: %+v err=%v", legacy, err)
	}

	updated, err := UpdateAccount(base, "alice", Account{Username: "alice", ProfileID: "default", Enabled: false}, false)
	if err != nil || updated.Enabled {
		t.Fatalf("disable account: %+v err=%v", updated, err)
	}
	if err := DeleteAccount(base, "alice"); err != nil {
		t.Fatal(err)
	}
	accounts, err := ListAccounts(base)
	if err != nil || len(accounts) != 0 {
		t.Fatalf("accounts after delete=%+v err=%v", accounts, err)
	}
}

func TestAccountValidationRejectsLegacyAndExtendedUsernames(t *testing.T) {
	setupSocks5SettingsDB(t)
	base := testBaseConfig(t)
	for _, username := range []string{"proxy", "alice@japan"} {
		if _, err := CreateAccount(base, Account{ID: strings.ReplaceAll(username, "@", "-"), Username: username, Password: "secret", Enabled: true}); err == nil {
			t.Fatalf("username %q should be rejected", username)
		}
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		_ = listener.Close()
		t.Fatalf("unexpected listener address type %T", listener.Addr())
	}
	port := addr.Port
	_ = listener.Close()
	return port
}

func TestManagerRunsMultipleListenersAndPreservesOldGenerationOnBindFailure(t *testing.T) {
	setupSocks5SettingsDB(t)
	base := testBaseConfig(t)
	base.Enabled = true
	firstPort := freePort(t)
	secondPort := freePort(t)
	manager := &Manager{}
	listeners := []ListenerSpec{
		{ID: "first", Enabled: true, ListenAddress: "127.0.0.1", Port: firstPort, DefaultProfileID: "default"},
		{ID: "second", Enabled: true, ListenAddress: "127.0.0.1", Port: secondPort, DefaultProfileID: "default"},
	}
	if err := manager.ApplyListeners(base, listeners); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(manager.Stop)
	statuses := manager.ListenerStatuses()
	if len(statuses) != 2 || !statuses[0].Running || !statuses[1].Running {
		t.Fatalf("unexpected listener statuses: %+v", statuses)
	}
	for _, port := range []int{firstPort, secondPort} {
		dialer := net.Dialer{Timeout: time.Second}
		conn, err := dialer.DialContext(context.Background(), "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if err != nil {
			t.Fatalf("dial listener %d: %v", port, err)
		}
		_ = conn.Close()
	}

	var listenConfig net.ListenConfig
	occupied, err := listenConfig.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = occupied.Close() }()
	occupiedAddr, ok := occupied.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("unexpected occupied listener address type %T", occupied.Addr())
	}
	occupiedPort := occupiedAddr.Port
	failed := append(append([]ListenerSpec{}, listeners...), ListenerSpec{ID: "zz-occupied", Enabled: true, ListenAddress: "127.0.0.1", Port: occupiedPort, DefaultProfileID: "default"})
	failed[0].Protocol = ProtocolHTTP
	if err := manager.ApplyListeners(base, failed); err == nil {
		t.Fatal("expected occupied listener apply to fail")
	}
	if statuses = manager.ListenerStatuses(); len(statuses) != 2 || statuses[0].Protocol != ProtocolSOCKS5 || !statuses[0].Running {
		t.Fatalf("failed apply did not restore the original listener: %+v", statuses)
	}
	for _, port := range []int{firstPort, secondPort} {
		dialer := net.Dialer{Timeout: time.Second}
		conn, err := dialer.DialContext(context.Background(), "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if err != nil {
			t.Fatalf("old listener %d was lost after failed apply: %v", port, err)
		}
		_ = conn.Close()
	}
}

type stubAddrListener struct {
	addr net.Addr
}

func (s *stubAddrListener) Accept() (net.Conn, error) { return nil, io.EOF }
func (s *stubAddrListener) Close() error              { return nil }
func (s *stubAddrListener) Addr() net.Addr            { return s.addr }

func TestListenerEndpointMatchingAndConflictDetection(t *testing.T) {
	cases := []struct {
		name         string
		existingSpec ListenerSpec
		boundAddr    net.Addr
		targetSpec   ListenerSpec
		wantMatch    bool
		wantConflict bool
	}{
		{
			name:         "0.0.0.0 bound as dual-stack :: matches 0.0.0.0",
			existingSpec: ListenerSpec{ID: "default", ListenAddress: "0.0.0.0", Port: 20000},
			boundAddr:    &net.TCPAddr{IP: net.IPv6unspecified, Port: 20000},
			targetSpec:   ListenerSpec{ID: "default", ListenAddress: "0.0.0.0", Port: 20000},
			wantMatch:    true,
			wantConflict: true,
		},
		{
			name:         "0.0.0.0 bound with nil IP matches 0.0.0.0",
			existingSpec: ListenerSpec{ID: "default", ListenAddress: "0.0.0.0", Port: 20000},
			boundAddr:    &net.TCPAddr{Port: 20000},
			targetSpec:   ListenerSpec{ID: "default", ListenAddress: "0.0.0.0", Port: 20000},
			wantMatch:    true,
			wantConflict: true,
		},
		{
			name:         "bracketed :: matches :: bound addr",
			existingSpec: ListenerSpec{ID: "default", ListenAddress: "::", Port: 20000},
			boundAddr:    &net.TCPAddr{IP: net.IPv6unspecified, Port: 20000},
			targetSpec:   ListenerSpec{ID: "default", ListenAddress: "[::]", Port: 20000},
			wantMatch:    true,
			wantConflict: true,
		},
		{
			name:         "wildcard :: bound addr conflicts with 127.0.0.1 on same port",
			existingSpec: ListenerSpec{ID: "default", ListenAddress: "0.0.0.0", Port: 20000},
			boundAddr:    &net.TCPAddr{IP: net.IPv6unspecified, Port: 20000},
			targetSpec:   ListenerSpec{ID: "default", ListenAddress: "127.0.0.1", Port: 20000},
			wantMatch:    false,
			wantConflict: true,
		},
		{
			name:         "127.0.0.1 bound addr conflicts with 0.0.0.0 on same port",
			existingSpec: ListenerSpec{ID: "default", ListenAddress: "127.0.0.1", Port: 20000},
			boundAddr:    &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 20000},
			targetSpec:   ListenerSpec{ID: "default", ListenAddress: "0.0.0.0", Port: 20000},
			wantMatch:    false,
			wantConflict: true,
		},
		{
			name:         "different port does not match or conflict",
			existingSpec: ListenerSpec{ID: "default", ListenAddress: "127.0.0.1", Port: 20000},
			boundAddr:    &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 20000},
			targetSpec:   ListenerSpec{ID: "default", ListenAddress: "127.0.0.1", Port: 20001},
			wantMatch:    false,
			wantConflict: false,
		},
		{
			name:         "different non-wildcard IP on same port does not conflict",
			existingSpec: ListenerSpec{ID: "default", ListenAddress: "127.0.0.1", Port: 20000},
			boundAddr:    &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 20000},
			targetSpec:   ListenerSpec{ID: "default", ListenAddress: "192.0.2.10", Port: 20000},
			wantMatch:    false,
			wantConflict: false,
		},
		{
			name:         "localhost bound as 127.0.0.1 matches localhost and conflicts with 127.0.0.1",
			existingSpec: ListenerSpec{ID: "default", ListenAddress: "localhost", Port: 20000},
			boundAddr:    &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 20000},
			targetSpec:   ListenerSpec{ID: "default", ListenAddress: "localhost", Port: 20000},
			wantMatch:    true,
			wantConflict: true,
		},
		{
			name:         "localhost bound as 127.0.0.1 also matches explicit 127.0.0.1",
			existingSpec: ListenerSpec{ID: "default", ListenAddress: "localhost", Port: 20000},
			boundAddr:    &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 20000},
			targetSpec:   ListenerSpec{ID: "default", ListenAddress: "127.0.0.1", Port: 20000},
			wantMatch:    true,
			wantConflict: true,
		},
		{
			name:         "127.0.0.1 bound conflicts with localhost target",
			existingSpec: ListenerSpec{ID: "default", ListenAddress: "127.0.0.1", Port: 20000},
			boundAddr:    &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 20000},
			targetSpec:   ListenerSpec{ID: "default", ListenAddress: "localhost", Port: 20000},
			wantMatch:    false,
			wantConflict: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runtime := &listenerRuntime{
				spec:     tc.existingSpec,
				listener: &stubAddrListener{addr: tc.boundAddr},
			}
			if got := listenerMatchesEndpoint(runtime, tc.targetSpec); got != tc.wantMatch {
				t.Fatalf("listenerMatchesEndpoint() = %v, want %v", got, tc.wantMatch)
			}
			if got := listenerConflictsWithEndpoint(runtime, tc.targetSpec); got != tc.wantConflict {
				t.Fatalf("listenerConflictsWithEndpoint() = %v, want %v", got, tc.wantConflict)
			}
		})
	}
}

func freeWildcardPort(t *testing.T) int {
	t.Helper()
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(context.Background(), "tcp", "0.0.0.0:0")
	if err != nil {
		t.Fatal(err)
	}
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		_ = listener.Close()
		t.Fatalf("unexpected listener address type %T", listener.Addr())
	}
	port := addr.Port
	_ = listener.Close()
	return port
}

func TestManagerHotReloadsWildcardListenerWithoutAddressInUse(t *testing.T) {
	setupSocks5SettingsDB(t)
	previous := listCandidateNodesFunc
	listCandidateNodesFunc = func(cfg Config) ([]models.Node, error) {
		if cfg.Selection == "specific" {
			return nil, errors.New("configured SOCKS5 node was not found")
		}
		return []models.Node{{ID: 1, Link: "socks5://upstream:1080", Name: "best-node"}}, nil
	}
	t.Cleanup(func() { listCandidateNodesFunc = previous })

	port := freeWildcardPort(t)
	manager := &Manager{}
	t.Cleanup(manager.Stop)

	cfgSpecific := Config{
		Enabled: true, ListenAddress: "0.0.0.0", Port: port, Username: "proxy", Password: "secret",
		RequireAuth: true, Selection: "specific", NodeID: 999999,
	}
	if err := manager.Apply(cfgSpecific); err != nil {
		t.Fatalf("initial apply failed: %v", err)
	}
	initialRuntimes := manager.runtimeSnapshot()
	if len(initialRuntimes) != 1 {
		t.Fatalf("expected 1 listener runtime, got %d", len(initialRuntimes))
	}

	// Re-applying identical wildcard config should reuse the existing runtime.
	if err := manager.Apply(cfgSpecific); err != nil {
		t.Fatalf("idempotent apply failed: %v", err)
	}
	reusedRuntimes := manager.runtimeSnapshot()
	if len(reusedRuntimes) != 1 || reusedRuntimes[0] != initialRuntimes[0] {
		t.Fatalf("expected unchanged wildcard listener runtime to be reused")
	}

	// Switching selection while keeping 0.0.0.0:port must replace the runtime cleanly.
	cfgBest := cfgSpecific
	cfgBest.Selection = "best"
	cfgBest.NodeID = 0
	if err := manager.Apply(cfgBest); err != nil {
		t.Fatalf("hot reload on wildcard listener failed: %v", err)
	}
	reloadedRuntimes := manager.runtimeSnapshot()
	if len(reloadedRuntimes) != 1 || reloadedRuntimes[0].cfg.Selection != "best" {
		t.Fatalf("unexpected reloaded runtime config: %+v", reloadedRuntimes)
	}
	snapshot, err := manager.RoutingSnapshot(RoutingSnapshotQuery{ProfileID: defaultProfileID})
	if err != nil {
		t.Fatalf("routing snapshot after reload failed: %v", err)
	}
	if snapshot.Total != 1 || len(snapshot.Items) != 1 || snapshot.Items[0].NodeName != "best-node" {
		t.Fatalf("unexpected routing snapshot after reload: %+v", snapshot)
	}

	// Switching between wildcard 0.0.0.0 and specific 127.0.0.1 on the same port must also rebind cleanly.
	cfgLoopback := cfgBest
	cfgLoopback.ListenAddress = "127.0.0.1"
	if err := manager.Apply(cfgLoopback); err != nil {
		t.Fatalf("switch wildcard -> loopback failed: %v", err)
	}
	if err := manager.Apply(cfgBest); err != nil {
		t.Fatalf("switch loopback -> wildcard failed: %v", err)
	}
}
