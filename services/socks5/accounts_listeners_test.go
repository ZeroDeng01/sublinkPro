package socks5

import (
	"context"
	"encoding/json"
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
	failed := append(append([]ListenerSpec{}, listeners...), ListenerSpec{ID: "occupied", Enabled: true, ListenAddress: "127.0.0.1", Port: occupiedPort, DefaultProfileID: "default"})
	if err := manager.ApplyListeners(base, failed); err == nil {
		t.Fatal("expected occupied listener apply to fail")
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
