package socks5

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"sublink/models"
)

func TestParseRoutingUsernameKeepsLegacyAndSupportsProfileAccount(t *testing.T) {
	tests := []struct {
		supplied           string
		profileID, account string
		extended, ok       bool
	}{
		{supplied: "proxy", profileID: "default", ok: true},
		{supplied: "proxy@japan", profileID: "japan", extended: true, ok: true},
		{supplied: "proxy@japan.user01", profileID: "japan", account: "user01", extended: true, ok: true},
		{supplied: "other@japan.user01", ok: false},
		{supplied: "proxy@bad.profile.extra", profileID: "bad", account: "profile.extra", extended: true, ok: true},
		{supplied: "proxy@BAD", profileID: "bad", extended: true, ok: true},
		{supplied: "proxy@bad.", ok: false},
	}
	for _, test := range tests {
		profileID, account, extended, ok := parseRoutingUsername("proxy", test.supplied)
		if profileID != test.profileID || account != test.account || extended != test.extended || ok != test.ok {
			t.Fatalf("parse %q = profile=%q account=%q extended=%v ok=%v", test.supplied, profileID, account, extended, ok)
		}
	}
}

func TestRoutingProfileAppliesIndependentRoutingSettings(t *testing.T) {
	base, err := NormalizeConfig(Config{
		Enabled: true, ListenAddress: "127.0.0.1", Username: "proxy", Password: "secret", RequireAuth: true,
		Selection: "best", MaxAttempts: 1, DialTimeoutSeconds: 30, StickySessionTTLSeconds: 1800,
	})
	if err != nil {
		t.Fatal(err)
	}
	profile, err := normalizeRoutingProfile(base, RoutingProfile{
		ID: "japan", Name: "Japan", Enabled: true, Selection: "smart", MaxAttempts: 4,
		DialTimeoutSeconds: 12, FailureCooldownSeconds: 45, CandidateCountries: []string{"jp"},
		StickySessionEnabled: true, StickySessionTTLSeconds: 900,
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg := profile.apply(base)
	if cfg.Selection != "smart" || cfg.MaxAttempts != 4 || cfg.DialTimeoutSeconds != 12 || cfg.FailureCooldownSeconds != 45 || cfg.CandidateCountries[0] != "JP" || cfg.StickySessionTTLSeconds != 900 {
		t.Fatalf("unexpected applied profile config: %+v", cfg)
	}
	if cfg.ListenAddress != base.ListenAddress || cfg.Username != base.Username || cfg.Password != base.Password || cfg.MaxConnections != base.MaxConnections {
		t.Fatalf("profile overwrote gateway-level config: %+v", cfg)
	}
}

func TestRoutingProfileStickyKeyUsesAccountAndProfileScope(t *testing.T) {
	profile := RoutingProfile{ID: "japan", Name: "Japan", Enabled: true, StickySessionEnabled: true}
	cfg := Config{Selection: "smart", StickySessionEnabled: true}
	identity := routingIdentity{ProfileID: "japan", Account: "user01", Extended: true}
	if got := routingProfileStickyKey(profile, cfg, identity, "192.0.2.1"); got != "profile:japan:account:user01" {
		t.Fatalf("account sticky key = %q", got)
	}
	identity.Account = ""
	if got := routingProfileStickyKey(profile, cfg, identity, "192.0.2.1"); got != "profile:japan:client_ip:192.0.2.1" {
		t.Fatalf("profile client sticky key = %q", got)
	}
}

func TestAuthenticateResolvesRoutingProfileIdentity(t *testing.T) {
	server, err := NewServer(Config{
		Enabled: true, ListenAddress: "127.0.0.1", Username: "proxy", Password: "secret", RequireAuth: true,
	}, func(context.Context, models.Node, string, uint16) (net.Conn, error) { return nil, nil })
	if err != nil {
		t.Fatal(err)
	}
	base := server.cfg
	custom, err := normalizeRoutingProfile(base, RoutingProfile{ID: "japan", Name: "Japan", Enabled: true, Selection: "smart"})
	if err != nil {
		t.Fatal(err)
	}
	server.profiles.replace(base, []RoutingProfile{defaultRoutingProfile(base), custom})

	client, serverConn := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = serverConn.Close() }()
	type result struct {
		identity routingIdentity
		err      error
	}
	done := make(chan result, 1)
	go func() {
		identity, authErr := server.authenticate(serverConn)
		done <- result{identity: identity, err: authErr}
	}()

	if _, err := client.Write([]byte{0x05, 0x01, 0x02}); err != nil {
		t.Fatal(err)
	}
	response := make([]byte, 2)
	if _, err := io.ReadFull(client, response); err != nil {
		t.Fatal(err)
	}
	username := []byte("proxy@japan.user01")
	password := []byte("secret")
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
		if got.err != nil || got.identity.ProfileID != "japan" || got.identity.ProfileName != "Japan" || got.identity.Account != "user01" {
			t.Fatalf("unexpected identity: %+v err=%v", got.identity, got.err)
		}
	case <-time.After(time.Second):
		t.Fatal("authentication did not complete")
	}
}
