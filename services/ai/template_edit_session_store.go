package ai

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"sublink/models"
)

const (
	TemplateEditSessionTTL               = 15 * time.Minute
	TemplateEditSessionCleanupInterval   = 5 * time.Minute
	TemplateEditSessionMaxActivePerOwner = 10

	templateEditSessionIDBytes = 16
)

type TemplateEditSessionCreateInput struct {
	OwnerKey           string
	Filename           string
	Category           string
	BaseHash           string
	BaseText           string
	CandidateText      string
	Operations         []TemplateEditOperation
	Validation         ValidationResult
	WarningFingerprint string
	Status             TemplateEditSessionStatus
	LastError          string
}

type TemplateEditSessionStore struct {
	mu                sync.RWMutex
	sessions          map[string]TemplateEditSession
	now               func() time.Time
	ttl               time.Duration
	cleanupInterval   time.Duration
	maxActivePerOwner int
	generateID        func() (string, error)
}

type TemplateEditSessionStoreOption func(*TemplateEditSessionStore)

func NewTemplateEditSessionStore(options ...TemplateEditSessionStoreOption) *TemplateEditSessionStore {
	store := &TemplateEditSessionStore{
		sessions:          make(map[string]TemplateEditSession),
		now:               time.Now,
		ttl:               TemplateEditSessionTTL,
		cleanupInterval:   TemplateEditSessionCleanupInterval,
		maxActivePerOwner: TemplateEditSessionMaxActivePerOwner,
		generateID:        generateTemplateEditSessionID,
	}
	for _, option := range options {
		option(store)
	}
	return store
}

func WithTemplateEditSessionClock(now func() time.Time) TemplateEditSessionStoreOption {
	return func(store *TemplateEditSessionStore) {
		if now != nil {
			store.now = now
		}
	}
}

func WithTemplateEditSessionIDGenerator(generateID func() (string, error)) TemplateEditSessionStoreOption {
	return func(store *TemplateEditSessionStore) {
		if generateID != nil {
			store.generateID = generateID
		}
	}
}

func WithTemplateEditSessionTTL(ttl time.Duration) TemplateEditSessionStoreOption {
	return func(store *TemplateEditSessionStore) {
		if ttl > 0 {
			store.ttl = ttl
		}
	}
}

func TemplateEditSessionOwnerKey(user *models.User) string {
	if user == nil {
		return ""
	}
	if user.ID > 0 {
		return fmt.Sprintf("user:%d", user.ID)
	}
	username := strings.ToLower(strings.TrimSpace(user.Username))
	if username == "" {
		return ""
	}
	return "username:" + username
}

func (store *TemplateEditSessionStore) Create(input TemplateEditSessionCreateInput) (TemplateEditSession, error) {
	ownerKey := strings.TrimSpace(input.OwnerKey)
	if ownerKey == "" {
		return TemplateEditSession{}, NewTemplateEditError(TemplateEditInvalidOperation, "template edit session requires an owner")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	now := store.now()
	store.expireExpiredLocked(now)
	if store.activeSessionCountLocked(ownerKey, now) >= store.maxActivePerOwner {
		return TemplateEditSession{}, NewTemplateEditError(TemplateEditSessionLimit, "template edit session limit reached")
	}

	sessionID, err := store.uniqueSessionIDLocked()
	if err != nil {
		return TemplateEditSession{}, err
	}
	status := input.Status
	if status == "" {
		status = TemplateEditSessionCreated
	}
	if !isTemplateEditSessionKnownStatus(status) {
		return TemplateEditSession{}, NewTemplateEditError(TemplateEditInvalidOperation, "template edit session status is invalid")
	}
	session := TemplateEditSession{
		SessionID:          sessionID,
		OwnerKey:           ownerKey,
		Status:             status,
		Filename:           input.Filename,
		Category:           input.Category,
		BaseHash:           input.BaseHash,
		BaseText:           input.BaseText,
		CandidateText:      input.CandidateText,
		Operations:         cloneTemplateEditOperations(input.Operations),
		Validation:         cloneValidationResult(input.Validation),
		WarningFingerprint: input.WarningFingerprint,
		CreatedAt:          now,
		ExpiresAt:          now.Add(store.ttl),
		LastError:          input.LastError,
	}
	store.sessions[sessionID] = cloneTemplateEditSession(session)
	return cloneTemplateEditSession(session), nil
}

func (store *TemplateEditSessionStore) Get(sessionID string) (TemplateEditSession, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.getLocked(sessionID, "", store.now())
}

func (store *TemplateEditSessionStore) GetForOwner(ownerKey string, sessionID string) (TemplateEditSession, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.getLocked(sessionID, strings.TrimSpace(ownerKey), store.now())
}

func (store *TemplateEditSessionStore) UpdateForOwner(ownerKey string, sessionID string, mutate func(*TemplateEditSession) error) (TemplateEditSession, error) {
	if mutate == nil {
		return TemplateEditSession{}, NewTemplateEditError(TemplateEditInvalidOperation, "template edit session update requires a mutation")
	}
	store.mu.Lock()
	defer store.mu.Unlock()

	now := store.now()
	current, err := store.getLocked(sessionID, strings.TrimSpace(ownerKey), now)
	if err != nil {
		return TemplateEditSession{}, err
	}
	next := cloneTemplateEditSession(current)
	if err := mutate(&next); err != nil {
		return TemplateEditSession{}, err
	}
	next.SessionID = current.SessionID
	next.OwnerKey = current.OwnerKey
	next.CreatedAt = current.CreatedAt
	next.ExpiresAt = current.ExpiresAt
	if next.Status == "" {
		next.Status = current.Status
	}
	if !isTemplateEditSessionTransitionAllowed(current.Status, next.Status) {
		return TemplateEditSession{}, NewTemplateEditError(TemplateEditInvalidOperation, "illegal template edit session status transition")
	}
	store.sessions[current.SessionID] = cloneTemplateEditSession(next)
	return cloneTemplateEditSession(next), nil
}

func (store *TemplateEditSessionStore) DiscardForOwner(ownerKey string, sessionID string) (TemplateEditSession, error) {
	return store.UpdateForOwner(ownerKey, sessionID, func(session *TemplateEditSession) error {
		session.Status = TemplateEditSessionDiscarded
		return nil
	})
}

func (store *TemplateEditSessionStore) ExpireForOwner(ownerKey string, sessionID string) (TemplateEditSession, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	now := store.now()
	current, err := store.getLocked(sessionID, strings.TrimSpace(ownerKey), now)
	if err != nil {
		return TemplateEditSession{}, err
	}
	current.Status = TemplateEditSessionExpired
	current.ExpiresAt = now
	store.sessions[current.SessionID] = cloneTemplateEditSession(current)
	return cloneTemplateEditSession(current), nil
}

func (store *TemplateEditSessionStore) ExpireExpiredSessions() int {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.expireExpiredLocked(store.now())
}

func (store *TemplateEditSessionStore) StartCleanup(ctx context.Context) {
	ticker := time.NewTicker(store.cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			store.ExpireExpiredSessions()
		}
	}
}

func (store *TemplateEditSessionStore) getLocked(sessionID string, ownerKey string, now time.Time) (TemplateEditSession, error) {
	session, ok := store.sessions[strings.TrimSpace(sessionID)]
	if !ok {
		return TemplateEditSession{}, NewTemplateEditError(TemplateEditSessionNotFound, "template edit session not found")
	}
	if ownerKey != "" && session.OwnerKey != ownerKey {
		return TemplateEditSession{}, NewTemplateEditError(TemplateEditSessionNotFound, "template edit session not found")
	}
	if !now.Before(session.ExpiresAt) {
		delete(store.sessions, session.SessionID)
		return TemplateEditSession{}, NewTemplateEditError(TemplateEditSessionExpiredError, "template edit session expired")
	}
	if session.Status == TemplateEditSessionExpired {
		delete(store.sessions, session.SessionID)
		return TemplateEditSession{}, NewTemplateEditError(TemplateEditSessionExpiredError, "template edit session expired")
	}
	return cloneTemplateEditSession(session), nil
}

func (store *TemplateEditSessionStore) uniqueSessionIDLocked() (string, error) {
	for attempt := 0; attempt < 3; attempt++ {
		sessionID, err := store.generateID()
		if err != nil {
			return "", err
		}
		if _, exists := store.sessions[sessionID]; !exists {
			return sessionID, nil
		}
	}
	return "", fmt.Errorf("template edit session id collision")
}

func (store *TemplateEditSessionStore) activeSessionCountLocked(ownerKey string, now time.Time) int {
	count := 0
	for _, session := range store.sessions {
		if session.OwnerKey != ownerKey || !now.Before(session.ExpiresAt) || isTemplateEditSessionTerminal(session.Status) {
			continue
		}
		count++
	}
	return count
}

func (store *TemplateEditSessionStore) expireExpiredLocked(now time.Time) int {
	count := 0
	for sessionID, session := range store.sessions {
		if now.Before(session.ExpiresAt) {
			continue
		}
		delete(store.sessions, sessionID)
		count++
	}
	return count
}

func generateTemplateEditSessionID() (string, error) {
	randomBytes := make([]byte, templateEditSessionIDBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

func isTemplateEditSessionTransitionAllowed(from TemplateEditSessionStatus, to TemplateEditSessionStatus) bool {
	if from == to {
		return true
	}
	if to == TemplateEditSessionExpired {
		return true
	}
	switch from {
	case TemplateEditSessionCreated:
		return to == TemplateEditSessionStreaming || to == TemplateEditSessionFailed || to == TemplateEditSessionDiscarded
	case TemplateEditSessionStreaming:
		return to == TemplateEditSessionOperationsReady || to == TemplateEditSessionFailed || to == TemplateEditSessionDiscarded
	case TemplateEditSessionOperationsReady:
		return to == TemplateEditSessionValidating || to == TemplateEditSessionFailed || to == TemplateEditSessionDiscarded
	case TemplateEditSessionValidating:
		return to == TemplateEditSessionPreviewReady || to == TemplateEditSessionFailed || to == TemplateEditSessionDiscarded
	case TemplateEditSessionPreviewReady:
		return to == TemplateEditSessionAccepted || to == TemplateEditSessionFailed || to == TemplateEditSessionDiscarded
	case TemplateEditSessionAccepted, TemplateEditSessionDiscarded, TemplateEditSessionFailed, TemplateEditSessionExpired:
		return false
	default:
		return false
	}
}

func isTemplateEditSessionTerminal(status TemplateEditSessionStatus) bool {
	switch status {
	case TemplateEditSessionAccepted, TemplateEditSessionDiscarded, TemplateEditSessionFailed, TemplateEditSessionExpired:
		return true
	default:
		return false
	}
}

func isTemplateEditSessionKnownStatus(status TemplateEditSessionStatus) bool {
	switch status {
	case TemplateEditSessionCreated,
		TemplateEditSessionStreaming,
		TemplateEditSessionOperationsReady,
		TemplateEditSessionValidating,
		TemplateEditSessionPreviewReady,
		TemplateEditSessionAccepted,
		TemplateEditSessionDiscarded,
		TemplateEditSessionFailed,
		TemplateEditSessionExpired:
		return true
	default:
		return false
	}
}

func cloneTemplateEditSession(session TemplateEditSession) TemplateEditSession {
	session.Operations = cloneTemplateEditOperations(session.Operations)
	session.Validation = cloneValidationResult(session.Validation)
	return session
}

func cloneTemplateEditOperations(operations []TemplateEditOperation) []TemplateEditOperation {
	if operations == nil {
		return nil
	}
	cloned := make([]TemplateEditOperation, len(operations))
	copy(cloned, operations)
	return cloned
}

func cloneValidationResult(validation ValidationResult) ValidationResult {
	validation.Errors = cloneStringSlice(validation.Errors)
	validation.Warnings = cloneStringSlice(validation.Warnings)
	validation.ProtectedTokensFound = cloneStringSlice(validation.ProtectedTokensFound)
	validation.Subscriptions = cloneStringSlice(validation.Subscriptions)
	return validation
}

func cloneStringSlice(values []string) []string {
	if values == nil {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}
