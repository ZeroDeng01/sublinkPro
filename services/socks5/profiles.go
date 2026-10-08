package socks5

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"sublink/models"
)

const (
	settingRoutingProfiles = "socks5_routing_profiles"
	defaultProfileID       = "default"
	maxRoutingProfiles     = 64
)

var (
	routingProfileIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,31}$`)
	routingProfilesMu       sync.Mutex
)

type RoutingProfile struct {
	ID                      string   `json:"id"`
	Name                    string   `json:"name"`
	Enabled                 bool     `json:"enabled"`
	Selection               string   `json:"selection"`
	NodeID                  int      `json:"nodeId"`
	MaxAttempts             int      `json:"maxAttempts"`
	DialTimeoutSeconds      int      `json:"dialTimeoutSeconds"`
	FailureCooldownSeconds  int      `json:"failureCooldownSeconds"`
	SpecificFallback        bool     `json:"specificFallback"`
	CandidateGroups         []string `json:"candidateGroups"`
	CandidateSources        []string `json:"candidateSources"`
	CandidateProtocols      []string `json:"candidateProtocols"`
	CandidateCountries      []string `json:"candidateCountries"`
	StickySessionEnabled    bool     `json:"stickySessionEnabled"`
	StickySessionTTLSeconds int      `json:"stickySessionTtlSeconds"`
	IsDefault               bool     `json:"isDefault"`
	CandidateCount          int      `json:"candidateCount"`
}

type routingIdentity struct {
	RawUsername string
	ProfileID   string
	ProfileName string
	Account     string
	Extended    bool
}

func defaultRoutingIdentity() routingIdentity {
	return routingIdentity{ProfileID: defaultProfileID, ProfileName: "Default"}
}

func parseRoutingUsername(baseUsername, supplied string) (profileID, account string, extended bool, ok bool) {
	if supplied == baseUsername {
		return defaultProfileID, "", false, true
	}
	prefix := baseUsername + "@"
	if baseUsername == "" || !strings.HasPrefix(supplied, prefix) {
		return "", "", false, false
	}
	suffix := strings.TrimSpace(strings.TrimPrefix(supplied, prefix))
	parts := strings.SplitN(suffix, ".", 2)
	profileID = strings.ToLower(strings.TrimSpace(parts[0]))
	if !routingProfileIDPattern.MatchString(profileID) {
		return "", "", false, false
	}
	if len(parts) == 2 {
		account = strings.TrimSpace(parts[1])
		if account == "" || len(account) > 128 {
			return "", "", false, false
		}
	}
	return profileID, account, true, true
}

func routingProfileStickyKey(profile RoutingProfile, cfg Config, identity routingIdentity, clientAddress string) string {
	if !cfg.StickySessionEnabled || cfg.Selection == "specific" {
		return ""
	}
	if identity.Account != "" {
		return "profile:" + profile.ID + ":account:" + identity.Account
	}
	if !profile.IsDefault || identity.Extended {
		if clientAddress == "" {
			return ""
		}
		return "profile:" + profile.ID + ":client_ip:" + clientAddress
	}
	return stickySessionKey(cfg, clientAddress, identity.RawUsername)
}

type routingProfileStore struct {
	mu       sync.RWMutex
	base     Config
	profiles map[string]RoutingProfile
}

func defaultRoutingProfile(cfg Config) RoutingProfile {
	return RoutingProfile{
		ID: defaultProfileID, Name: "Default", Enabled: true, Selection: cfg.Selection, NodeID: cfg.NodeID,
		MaxAttempts: cfg.MaxAttempts, DialTimeoutSeconds: cfg.DialTimeoutSeconds,
		FailureCooldownSeconds: cfg.FailureCooldownSeconds, SpecificFallback: cfg.SpecificFallback,
		CandidateGroups: append([]string{}, cfg.CandidateGroups...), CandidateSources: append([]string{}, cfg.CandidateSources...),
		CandidateProtocols: append([]string{}, cfg.CandidateProtocols...), CandidateCountries: append([]string{}, cfg.CandidateCountries...),
		StickySessionEnabled: cfg.StickySessionEnabled, StickySessionTTLSeconds: cfg.StickySessionTTLSeconds, IsDefault: true,
	}
}

func normalizeRoutingProfile(base Config, profile RoutingProfile) (RoutingProfile, error) {
	profile.ID = strings.ToLower(strings.TrimSpace(profile.ID))
	profile.Name = strings.TrimSpace(profile.Name)
	if profile.ID == "" || !routingProfileIDPattern.MatchString(profile.ID) {
		return profile, errors.New("SOCKS5 routing profile ID must start with a letter or digit and contain only lowercase letters, digits, underscores, or hyphens")
	}
	if profile.ID == defaultProfileID {
		return profile, errors.New("the default SOCKS5 routing profile is reserved")
	}
	if profile.Name == "" || len(profile.Name) > 64 {
		return profile, errors.New("SOCKS5 routing profile name must contain 1-64 bytes")
	}
	if profile.Selection == "" {
		profile.Selection = "smart"
	}
	if profile.MaxAttempts == 0 {
		profile.MaxAttempts = 3
	}
	if profile.DialTimeoutSeconds == 0 {
		profile.DialTimeoutSeconds = base.DialTimeoutSeconds
	}
	if profile.StickySessionTTLSeconds == 0 {
		profile.StickySessionTTLSeconds = base.StickySessionTTLSeconds
	}
	cfg := profile.apply(base)
	normalized, err := NormalizeConfig(cfg)
	if err != nil {
		return profile, err
	}
	profile.Selection = normalized.Selection
	profile.NodeID = normalized.NodeID
	profile.MaxAttempts = normalized.MaxAttempts
	profile.DialTimeoutSeconds = normalized.DialTimeoutSeconds
	profile.FailureCooldownSeconds = normalized.FailureCooldownSeconds
	profile.SpecificFallback = normalized.SpecificFallback
	profile.CandidateGroups = append([]string{}, normalized.CandidateGroups...)
	profile.CandidateSources = append([]string{}, normalized.CandidateSources...)
	profile.CandidateProtocols = append([]string{}, normalized.CandidateProtocols...)
	profile.CandidateCountries = append([]string{}, normalized.CandidateCountries...)
	profile.StickySessionTTLSeconds = normalized.StickySessionTTLSeconds
	profile.IsDefault = false
	return profile, nil
}

func (profile RoutingProfile) apply(base Config) Config {
	cfg := base
	cfg.Selection = profile.Selection
	cfg.NodeID = profile.NodeID
	cfg.MaxAttempts = profile.MaxAttempts
	cfg.DialTimeoutSeconds = profile.DialTimeoutSeconds
	cfg.FailureCooldownSeconds = profile.FailureCooldownSeconds
	cfg.SpecificFallback = profile.SpecificFallback
	cfg.CandidateGroups = append([]string{}, profile.CandidateGroups...)
	cfg.CandidateSources = append([]string{}, profile.CandidateSources...)
	cfg.CandidateProtocols = append([]string{}, profile.CandidateProtocols...)
	cfg.CandidateCountries = append([]string{}, profile.CandidateCountries...)
	cfg.StickySessionEnabled = profile.StickySessionEnabled
	cfg.StickySessionTTLSeconds = profile.StickySessionTTLSeconds
	return cfg
}

func loadCustomRoutingProfiles(base Config) ([]RoutingProfile, error) {
	value, err := models.GetSetting(settingRoutingProfiles)
	if err != nil || strings.TrimSpace(value) == "" {
		return []RoutingProfile{}, nil
	}
	var stored []RoutingProfile
	if err := json.Unmarshal([]byte(value), &stored); err != nil {
		return nil, fmt.Errorf("invalid SOCKS5 routing profiles: %w", err)
	}
	if len(stored) > maxRoutingProfiles {
		return nil, fmt.Errorf("SOCKS5 routing profiles exceed the limit of %d", maxRoutingProfiles)
	}
	seen := make(map[string]struct{}, len(stored))
	profiles := make([]RoutingProfile, 0, len(stored))
	for _, profile := range stored {
		normalized, normalizeErr := normalizeRoutingProfile(base, profile)
		if normalizeErr != nil {
			return nil, fmt.Errorf("invalid SOCKS5 routing profile %q: %w", profile.ID, normalizeErr)
		}
		if _, exists := seen[normalized.ID]; exists {
			return nil, fmt.Errorf("duplicate SOCKS5 routing profile ID %q", normalized.ID)
		}
		seen[normalized.ID] = struct{}{}
		profiles = append(profiles, normalized)
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ID < profiles[j].ID })
	return profiles, nil
}

func saveCustomRoutingProfiles(profiles []RoutingProfile) error {
	encoded, err := json.Marshal(profiles)
	if err != nil {
		return err
	}
	return models.SetSetting(settingRoutingProfiles, string(encoded))
}

func loadRoutingProfiles(base Config) ([]RoutingProfile, error) {
	profiles, err := loadCustomRoutingProfiles(base)
	if err != nil {
		return nil, err
	}
	result := make([]RoutingProfile, 0, len(profiles)+1)
	result = append(result, defaultRoutingProfile(base))
	result = append(result, profiles...)
	return result, nil
}

func ListRoutingProfiles(base Config) ([]RoutingProfile, error) {
	profiles, err := loadRoutingProfiles(base)
	if err != nil {
		return nil, err
	}
	var modelNode models.Node
	nodes, listErr := modelNode.ListWithFilters(models.NodeFilter{})
	if listErr != nil {
		return profiles, nil
	}
	for index := range profiles {
		cfg := base
		if !profiles[index].IsDefault {
			cfg = profiles[index].apply(base)
		}
		profiles[index].CandidateCount = countRoutingProfileCandidates(cfg, nodes)
	}
	return profiles, nil
}

func countRoutingProfileCandidates(cfg Config, nodes []models.Node) int {
	filtered := filterCandidatePool(nodes, cfg)
	if cfg.Selection != "specific" {
		return len(filtered)
	}
	specific, ok := models.GetNodeByID(cfg.NodeID)
	if !ok || strings.TrimSpace(specific.Link) == "" {
		return 0
	}
	if !cfg.SpecificFallback {
		return 1
	}
	count := 1
	for _, node := range filtered {
		if node.ID != specific.ID {
			count++
		}
	}
	return count
}

func CreateRoutingProfile(base Config, input RoutingProfile) (RoutingProfile, error) {
	routingProfilesMu.Lock()
	defer routingProfilesMu.Unlock()
	profiles, err := loadCustomRoutingProfiles(base)
	if err != nil {
		return input, err
	}
	if len(profiles) >= maxRoutingProfiles {
		return input, fmt.Errorf("SOCKS5 routing profiles are limited to %d", maxRoutingProfiles)
	}
	normalized, err := normalizeRoutingProfile(base, input)
	if err != nil {
		return input, err
	}
	for _, profile := range profiles {
		if profile.ID == normalized.ID {
			return input, errors.New("SOCKS5 routing profile ID already exists")
		}
	}
	profiles = append(profiles, normalized)
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ID < profiles[j].ID })
	if err := saveCustomRoutingProfiles(profiles); err != nil {
		return input, err
	}
	return normalized, nil
}

func UpdateRoutingProfile(base Config, id string, input RoutingProfile) (RoutingProfile, error) {
	routingProfilesMu.Lock()
	defer routingProfilesMu.Unlock()
	id = strings.ToLower(strings.TrimSpace(id))
	if id == defaultProfileID {
		return input, errors.New("the default SOCKS5 routing profile is managed by gateway settings")
	}
	profiles, err := loadCustomRoutingProfiles(base)
	if err != nil {
		return input, err
	}
	input.ID = id
	normalized, err := normalizeRoutingProfile(base, input)
	if err != nil {
		return input, err
	}
	if !normalized.Enabled {
		accountReferenced, referenceErr := RoutingProfileReferencedByAccount(base, id)
		if referenceErr != nil {
			return input, referenceErr
		}
		listenerReferenced, referenceErr := RoutingProfileReferencedByListener(base, id)
		if referenceErr != nil {
			return input, referenceErr
		}
		if accountReferenced || listenerReferenced {
			return input, errors.New("SOCKS5 routing profile is still referenced and cannot be disabled")
		}
	}
	found := false
	for index := range profiles {
		if profiles[index].ID == id {
			profiles[index] = normalized
			found = true
			break
		}
	}
	if !found {
		return input, errors.New("SOCKS5 routing profile was not found")
	}
	if err := saveCustomRoutingProfiles(profiles); err != nil {
		return input, err
	}
	return normalized, nil
}

func DeleteRoutingProfile(base Config, id string) error {
	routingProfilesMu.Lock()
	defer routingProfilesMu.Unlock()
	id = strings.ToLower(strings.TrimSpace(id))
	if id == defaultProfileID {
		return errors.New("the default SOCKS5 routing profile cannot be deleted")
	}
	profiles, err := loadCustomRoutingProfiles(base)
	if err != nil {
		return err
	}
	result := make([]RoutingProfile, 0, len(profiles))
	found := false
	for _, profile := range profiles {
		if profile.ID == id {
			found = true
			continue
		}
		result = append(result, profile)
	}
	if !found {
		return errors.New("SOCKS5 routing profile was not found")
	}
	referenced, err := RoutingProfileReferencedByAccount(base, id)
	if err != nil {
		return err
	}
	if referenced {
		return errors.New("SOCKS5 routing profile is still referenced by an account")
	}
	referenced, err = RoutingProfileReferencedByListener(base, id)
	if err != nil {
		return err
	}
	if referenced {
		return errors.New("SOCKS5 routing profile is still referenced by a listener")
	}
	return saveCustomRoutingProfiles(result)
}

func newRoutingProfileStore(base Config) (*routingProfileStore, error) {
	profiles, err := loadRoutingProfiles(base)
	if err != nil {
		return nil, err
	}
	store := &routingProfileStore{}
	store.replace(base, profiles)
	return store, nil
}

func (s *routingProfileStore) replace(base Config, profiles []RoutingProfile) {
	mapped := make(map[string]RoutingProfile, len(profiles))
	for _, profile := range profiles {
		mapped[profile.ID] = profile
	}
	s.mu.Lock()
	s.base = base
	s.profiles = mapped
	s.mu.Unlock()
}

func (s *routingProfileStore) resolve(id string) (RoutingProfile, Config, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		id = defaultProfileID
	}
	s.mu.RLock()
	profile, ok := s.profiles[id]
	base := s.base
	s.mu.RUnlock()
	if !ok || !profile.Enabled {
		return RoutingProfile{}, Config{}, false
	}
	if profile.IsDefault {
		return profile, base, true
	}
	return profile, profile.apply(base), true
}

func (s *routingProfileStore) enabled() []RoutingProfile {
	s.mu.RLock()
	profiles := make([]RoutingProfile, 0, len(s.profiles))
	for _, profile := range s.profiles {
		if profile.Enabled {
			profiles = append(profiles, profile)
		}
	}
	s.mu.RUnlock()
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ID < profiles[j].ID })
	return profiles
}

func (s *Server) candidateNodesForProfile(id string) (RoutingProfile, []models.Node, error) {
	profile, cfg, ok := s.profiles.resolve(id)
	if !ok {
		return RoutingProfile{}, nil, errors.New("SOCKS5 routing profile was not found or is disabled")
	}
	nodes, err := listCandidateNodesFunc(cfg)
	if err != nil {
		return profile, nil, err
	}
	return profile, nodes, nil
}

func (s *Server) allProfileCandidateNodes() []models.Node {
	profiles := s.profiles.enabled()
	seen := make(map[string]struct{})
	nodes := make([]models.Node, 0)
	for _, profile := range profiles {
		_, cfg, ok := s.profiles.resolve(profile.ID)
		if !ok {
			continue
		}
		profileNodes, err := listCandidateNodesFunc(cfg)
		if err != nil {
			continue
		}
		for _, node := range profileNodes {
			key := adapterKey(node)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			nodes = append(nodes, node)
		}
	}
	return nodes
}
