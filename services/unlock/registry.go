package unlock

import (
	"fmt"
	"strings"
	"sublink/models"
	"sync"
)

type UnlockChecker interface {
	Key() string
	Aliases() []string
	Check(runtime UnlockRuntime) models.UnlockProviderResult
}

type unlockRegistry struct {
	mu              sync.RWMutex
	checkers        map[string]UnlockChecker
	canonicalByName map[string]string
	defaultKeys     []string
}

var globalUnlockRegistry = newUnlockRegistry()

func newUnlockRegistry() *unlockRegistry {
	return &unlockRegistry{
		checkers:        make(map[string]UnlockChecker),
		canonicalByName: make(map[string]string),
		defaultKeys:     []string{},
	}
}

func RegisterUnlockChecker(checker UnlockChecker) {
	globalUnlockRegistry.register(checker)
}

func ListRegisteredUnlockProviders() []string {
	return globalUnlockRegistry.listDefaults()
}

func ResolveUnlockProviders(providers []string) []string {
	return globalUnlockRegistry.resolve(providers)
}

func HasRegisteredUnlockProvider(provider string) bool {
	_, ok := globalUnlockRegistry.get(provider)
	return ok
}

func GetUnlockChecker(provider string) (UnlockChecker, bool) {
	return globalUnlockRegistry.get(provider)
}

func (r *unlockRegistry) register(checker UnlockChecker) {
	key := models.NormalizeUnlockProvider(checker.Key())
	if key == "" {
		panic("unlock checker key is empty or invalid")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.checkers[key]; exists {
		panic(fmt.Sprintf("unlock checker already registered: %s", key))
	}
	r.checkers[key] = checker
	r.defaultKeys = append(r.defaultKeys, key)
	r.canonicalByName[key] = key

	for _, alias := range checker.Aliases() {
		normalized := models.NormalizeUnlockProvider(alias)
		if normalized == "" {
			normalized = strings.ToLower(strings.TrimSpace(alias))
		}
		if normalized == "" {
			continue
		}
		if existing, exists := r.canonicalByName[normalized]; exists && existing != key {
			panic(fmt.Sprintf("unlock checker alias conflict: %s", alias))
		}
		r.canonicalByName[normalized] = key
	}
}

func (r *unlockRegistry) get(provider string) (UnlockChecker, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := models.NormalizeUnlockProvider(provider)
	if key == "" {
		key = strings.ToLower(strings.TrimSpace(provider))
	}
	canonical, exists := r.canonicalByName[key]
	if !exists {
		return nil, false
	}
	checker, ok := r.checkers[canonical]
	return checker, ok
}

func (r *unlockRegistry) listDefaults() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	providers := make([]string, len(r.defaultKeys))
	copy(providers, r.defaultKeys)
	return providers
}

func (r *unlockRegistry) resolve(providers []string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	requested := providers
	if len(requested) == 0 {
		requested = r.defaultKeys
	}

	seen := make(map[string]struct{}, len(requested))
	resolved := make([]string, 0, len(requested))
	for _, provider := range requested {
		key := models.NormalizeUnlockProvider(provider)
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(provider))
		}
		canonical, exists := r.canonicalByName[key]
		if !exists {
			continue
		}
		if _, duplicated := seen[canonical]; duplicated {
			continue
		}
		seen[canonical] = struct{}{}
		resolved = append(resolved, canonical)
	}
	return resolved
}
