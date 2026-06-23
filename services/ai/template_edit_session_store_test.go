package ai

import (
	"encoding/hex"
	"errors"
	"testing"
	"time"
	"unicode/utf8"

	"sublink/models"
)

func TestTemplateEditSessionStoreCreateGetAndOwnerLookup(t *testing.T) {
	now := time.Date(2026, 6, 24, 10, 0, 0, 0, time.UTC)
	store := NewTemplateEditSessionStore(WithTemplateEditSessionClock(func() time.Time { return now }))
	ownerKey := TemplateEditSessionOwnerKey(&models.User{ID: 42, Username: "Admin"})

	created, err := store.Create(TemplateEditSessionCreateInput{
		OwnerKey:           ownerKey,
		Filename:           "clash.yaml",
		Category:           "clash",
		BaseHash:           "base-hash",
		BaseText:           "proxies: []\n",
		CandidateText:      "proxies: []\nrules: []\n",
		Operations:         []TemplateEditOperation{{Op: TemplateEditOperationReplace, OldString: "a", NewString: "b"}},
		Validation:         ValidationResult{Valid: true, Warnings: []string{"warning"}},
		WarningFingerprint: "fingerprint",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if created.OwnerKey != ownerKey {
		t.Fatalf("expected owner key %q, got %q", ownerKey, created.OwnerKey)
	}
	if created.Status != TemplateEditSessionCreated {
		t.Fatalf("expected created status, got %q", created.Status)
	}
	if created.CreatedAt != now {
		t.Fatalf("expected createdAt %v, got %v", now, created.CreatedAt)
	}
	if created.ExpiresAt != now.Add(TemplateEditSessionTTL) {
		t.Fatalf("expected expiresAt %v, got %v", now.Add(TemplateEditSessionTTL), created.ExpiresAt)
	}

	fetched, err := store.GetForOwner(ownerKey, created.SessionID)
	if err != nil {
		t.Fatalf("get owner session: %v", err)
	}
	if fetched.Filename != "clash.yaml" || fetched.BaseHash != "base-hash" || fetched.CandidateText == "" {
		t.Fatalf("unexpected fetched session: %#v", fetched)
	}
	fetched.Operations[0].NewString = "mutated"
	fetched.Validation.Warnings[0] = "mutated"

	again, err := store.Get(created.SessionID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if again.Operations[0].NewString != "b" {
		t.Fatalf("session operations were mutated through returned copy")
	}
	if again.Validation.Warnings[0] != "warning" {
		t.Fatalf("session validation was mutated through returned copy")
	}

	_, err = store.GetForOwner("username:other", created.SessionID)
	assertTemplateEditErrorCode(t, err, TemplateEditSessionNotFound)
}

func TestTemplateEditSessionExpires(t *testing.T) {
	now := time.Date(2026, 6, 24, 10, 0, 0, 0, time.UTC)
	store := NewTemplateEditSessionStore(WithTemplateEditSessionClock(func() time.Time { return now }))
	created := createTemplateEditSessionForTest(t, store, "user:1")

	now = now.Add(TemplateEditSessionTTL - time.Nanosecond)
	if _, err := store.GetForOwner("user:1", created.SessionID); err != nil {
		t.Fatalf("session should exist before ttl: %v", err)
	}

	now = now.Add(time.Nanosecond)
	_, err := store.GetForOwner("user:1", created.SessionID)
	assertTemplateEditErrorCode(t, err, TemplateEditSessionExpiredError)

	_, err = store.GetForOwner("user:1", created.SessionID)
	assertTemplateEditErrorCode(t, err, TemplateEditSessionNotFound)
}

func TestTemplateEditSessionLimit(t *testing.T) {
	now := time.Date(2026, 6, 24, 10, 0, 0, 0, time.UTC)
	store := NewTemplateEditSessionStore(WithTemplateEditSessionClock(func() time.Time { return now }))

	for i := 0; i < TemplateEditSessionMaxActivePerOwner; i++ {
		createTemplateEditSessionForTest(t, store, "user:limit")
	}
	_, err := store.Create(TemplateEditSessionCreateInput{OwnerKey: "user:limit", Filename: "overflow.yaml", BaseHash: "hash", BaseText: "base"})
	assertTemplateEditErrorCode(t, err, TemplateEditSessionLimit)

	if _, err := store.Create(TemplateEditSessionCreateInput{OwnerKey: "user:other", Filename: "other.yaml", BaseHash: "hash", BaseText: "base"}); err != nil {
		t.Fatalf("different owner should not share limit: %v", err)
	}

	now = now.Add(TemplateEditSessionTTL + time.Nanosecond)
	if _, err := store.Create(TemplateEditSessionCreateInput{OwnerKey: "user:limit", Filename: "new.yaml", BaseHash: "hash", BaseText: "base"}); err != nil {
		t.Fatalf("expired sessions should not count against limit: %v", err)
	}
}

func TestTemplateEditSessionDiscard(t *testing.T) {
	store := NewTemplateEditSessionStore()
	created := createTemplateEditSessionForTest(t, store, "user:discard")

	discarded, err := store.DiscardForOwner("user:discard", created.SessionID)
	if err != nil {
		t.Fatalf("discard session: %v", err)
	}
	if discarded.Status != TemplateEditSessionDiscarded {
		t.Fatalf("expected discarded status, got %q", discarded.Status)
	}

	_, err = store.UpdateForOwner("user:discard", created.SessionID, func(session *TemplateEditSession) error {
		session.Status = TemplateEditSessionPreviewReady
		return nil
	})
	assertTemplateEditErrorCode(t, err, TemplateEditInvalidOperation)
}

func TestTemplateEditSessionExpireMethod(t *testing.T) {
	store := NewTemplateEditSessionStore()
	created := createTemplateEditSessionForTest(t, store, "user:expire")

	expired, err := store.ExpireForOwner("user:expire", created.SessionID)
	if err != nil {
		t.Fatalf("expire session: %v", err)
	}
	if expired.Status != TemplateEditSessionExpired {
		t.Fatalf("expected expired status, got %q", expired.Status)
	}

	_, err = store.GetForOwner("user:expire", created.SessionID)
	assertTemplateEditErrorCode(t, err, TemplateEditSessionExpiredError)
}

func TestTemplateEditSessionRejectsIllegalStatusTransition(t *testing.T) {
	store := NewTemplateEditSessionStore()
	created := createTemplateEditSessionForTest(t, store, "user:transition")

	_, err := store.UpdateForOwner("user:transition", created.SessionID, func(session *TemplateEditSession) error {
		session.Status = TemplateEditSessionAccepted
		return nil
	})
	assertTemplateEditErrorCode(t, err, TemplateEditInvalidOperation)

	updated, err := store.UpdateForOwner("user:transition", created.SessionID, func(session *TemplateEditSession) error {
		session.Status = TemplateEditSessionStreaming
		session.LastError = ""
		return nil
	})
	if err != nil {
		t.Fatalf("legal transition failed: %v", err)
	}
	if updated.Status != TemplateEditSessionStreaming {
		t.Fatalf("expected streaming status, got %q", updated.Status)
	}
}

func TestTemplateEditSessionIDIsUnguessableLengthSurrogate(t *testing.T) {
	store := NewTemplateEditSessionStore()
	seen := map[string]bool{}
	for i := 0; i < 32; i++ {
		session := createTemplateEditSessionForTest(t, store, "user:ids")
		if session.SessionID == "" {
			t.Fatalf("session id is empty")
		}
		decoded, err := hex.DecodeString(session.SessionID)
		if err != nil {
			t.Fatalf("session id should be hex: %v", err)
		}
		if len(decoded) < templateEditSessionIDBytes {
			t.Fatalf("session id has %d random bytes, want at least %d", len(decoded), templateEditSessionIDBytes)
		}
		if utf8.RuneCountInString(session.SessionID) < 32 {
			t.Fatalf("session id length %d is too short", len(session.SessionID))
		}
		if seen[session.SessionID] {
			t.Fatalf("duplicate session id generated: %s", session.SessionID)
		}
		seen[session.SessionID] = true
		if _, err := store.DiscardForOwner("user:ids", session.SessionID); err != nil {
			t.Fatalf("discard session after id check: %v", err)
		}
	}
}

func TestTemplateEditSessionOwnerKeyUsesStableUserIdentity(t *testing.T) {
	if got := TemplateEditSessionOwnerKey(&models.User{ID: 7, Username: "Admin"}); got != "user:7" {
		t.Fatalf("expected id owner key, got %q", got)
	}
	if got := TemplateEditSessionOwnerKey(&models.User{Username: " Admin "}); got != "username:admin" {
		t.Fatalf("expected username owner key, got %q", got)
	}
	if got := TemplateEditSessionOwnerKey(nil); got != "" {
		t.Fatalf("expected empty nil owner key, got %q", got)
	}
}

func createTemplateEditSessionForTest(t *testing.T, store *TemplateEditSessionStore, ownerKey string) TemplateEditSession {
	t.Helper()
	session, err := store.Create(TemplateEditSessionCreateInput{
		OwnerKey: ownerKey,
		Filename: "template.yaml",
		Category: "clash",
		BaseHash: "hash",
		BaseText: "proxies: []\n",
	})
	if err != nil {
		var editErr *TemplateEditError
		if errors.As(err, &editErr) {
			t.Fatalf("create session returned code %s: %v", editErr.Code, err)
		}
		t.Fatalf("create session: %v", err)
	}
	return session
}
