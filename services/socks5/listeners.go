package socks5

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"sublink/models"
)

const (
	settingListeners = "socks5_listeners"
	maxListeners     = 64
)

var (
	listenerIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,31}$`)
	listenersMu       sync.Mutex
)

// ListenerSpec describes one SOCKS5 inbound endpoint. Authentication credentials
// remain shared with the legacy gateway and the independent account store.
type ListenerSpec struct {
	ID               string `json:"id"`
	Enabled          bool   `json:"enabled"`
	ListenAddress    string `json:"listenAddress"`
	Port             int    `json:"port"`
	DefaultProfileID string `json:"defaultProfileId"`
	RequireAuth      *bool  `json:"requireAuth,omitempty"`
}

type ListenerStatus struct {
	ListenerSpec
	Running      bool   `json:"running"`
	BoundAddress string `json:"boundAddress,omitempty"`
}

func legacyListener(base Config) ListenerSpec {
	return ListenerSpec{
		ID: defaultProfileID, Enabled: true, ListenAddress: base.ListenAddress,
		Port: base.Port, DefaultProfileID: defaultProfileID,
	}
}

func normalizeListener(base Config, input ListenerSpec) (ListenerSpec, error) {
	input.ID = strings.ToLower(strings.TrimSpace(input.ID))
	input.ListenAddress = strings.TrimSpace(input.ListenAddress)
	input.DefaultProfileID = strings.ToLower(strings.TrimSpace(input.DefaultProfileID))
	if !listenerIDPattern.MatchString(input.ID) {
		return input, errors.New("SOCKS5 listener ID must start with a letter or digit and contain only lowercase letters, digits, underscores, or hyphens")
	}
	if input.ListenAddress == "" {
		input.ListenAddress = defaultListenAddress
	}
	if input.Port < 1 || input.Port > 65535 {
		return input, errors.New("SOCKS5 listener port must be between 1 and 65535")
	}
	if input.DefaultProfileID == "" {
		input.DefaultProfileID = defaultProfileID
	}
	profiles, err := newRoutingProfileStore(base)
	if err != nil {
		return input, err
	}
	if _, _, ok := profiles.resolve(input.DefaultProfileID); !ok {
		return input, errors.New("SOCKS5 listener default routing profile was not found or is disabled")
	}
	requireAuth := base.RequireAuth
	if input.RequireAuth != nil {
		requireAuth = *input.RequireAuth
	}
	if input.Enabled && !requireAuth && !isLoopbackListenAddress(input.ListenAddress) {
		return input, errors.New("SOCKS5 authentication can only be disabled on a loopback listen address")
	}
	return input, nil
}

func normalizeListeners(base Config, listeners []ListenerSpec) ([]ListenerSpec, error) {
	if len(listeners) > maxListeners {
		return nil, fmt.Errorf("SOCKS5 listeners exceed the limit of %d", maxListeners)
	}
	result := make([]ListenerSpec, 0, len(listeners))
	seenIDs := make(map[string]struct{}, len(listeners))
	seenEndpoints := make(map[string]struct{}, len(listeners))
	for _, item := range listeners {
		normalized, err := normalizeListener(base, item)
		if err != nil {
			return nil, fmt.Errorf("invalid SOCKS5 listener %q: %w", item.ID, err)
		}
		if _, exists := seenIDs[normalized.ID]; exists {
			return nil, fmt.Errorf("duplicate SOCKS5 listener ID %q", normalized.ID)
		}
		endpoint := strings.ToLower(net.JoinHostPort(normalized.ListenAddress, strconv.Itoa(normalized.Port)))
		if normalized.Enabled {
			if _, exists := seenEndpoints[endpoint]; exists {
				return nil, fmt.Errorf("duplicate SOCKS5 listener endpoint %q", endpoint)
			}
			seenEndpoints[endpoint] = struct{}{}
		}
		seenIDs[normalized.ID] = struct{}{}
		result = append(result, normalized)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func LoadListeners(base Config) ([]ListenerSpec, error) {
	value, err := models.GetSetting(settingListeners)
	if err != nil || strings.TrimSpace(value) == "" {
		return []ListenerSpec{legacyListener(base)}, nil
	}
	var listeners []ListenerSpec
	if err := json.Unmarshal([]byte(value), &listeners); err != nil {
		return nil, fmt.Errorf("invalid SOCKS5 listeners: %w", err)
	}
	return normalizeListeners(base, listeners)
}

func SaveListeners(base Config, listeners []ListenerSpec) ([]ListenerSpec, error) {
	normalized, err := normalizeListeners(base, listeners)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}
	if err := models.SetSetting(settingListeners, string(encoded)); err != nil {
		return nil, err
	}
	return normalized, nil
}

func ListListeners(base Config) ([]ListenerSpec, error) { return LoadListeners(base) }

func CreateListener(base Config, input ListenerSpec) (ListenerSpec, error) {
	listenersMu.Lock()
	defer listenersMu.Unlock()
	listeners, err := LoadListeners(base)
	if err != nil {
		return input, err
	}
	if len(listeners) >= maxListeners {
		return input, fmt.Errorf("SOCKS5 listeners are limited to %d", maxListeners)
	}
	normalized, err := normalizeListener(base, input)
	if err != nil {
		return input, err
	}
	listeners = append(listeners, normalized)
	if _, err := SaveListeners(base, listeners); err != nil {
		return input, err
	}
	return normalized, nil
}

func UpdateListener(base Config, id string, input ListenerSpec) (ListenerSpec, error) {
	listenersMu.Lock()
	defer listenersMu.Unlock()
	id = strings.ToLower(strings.TrimSpace(id))
	listeners, err := LoadListeners(base)
	if err != nil {
		return input, err
	}
	input.ID = id
	normalized, err := normalizeListener(base, input)
	if err != nil {
		return input, err
	}
	found := false
	for i := range listeners {
		if listeners[i].ID == id {
			listeners[i] = normalized
			found = true
			break
		}
	}
	if !found {
		return input, errors.New("SOCKS5 listener was not found")
	}
	if _, err := SaveListeners(base, listeners); err != nil {
		return input, err
	}
	return normalized, nil
}

func DeleteListener(base Config, id string) error {
	listenersMu.Lock()
	defer listenersMu.Unlock()
	id = strings.ToLower(strings.TrimSpace(id))
	listeners, err := LoadListeners(base)
	if err != nil {
		return err
	}
	result := make([]ListenerSpec, 0, len(listeners))
	found := false
	for _, listener := range listeners {
		if listener.ID == id {
			found = true
			continue
		}
		result = append(result, listener)
	}
	if !found {
		return errors.New("SOCKS5 listener was not found")
	}
	_, err = SaveListeners(base, result)
	return err
}

func RoutingProfileReferencedByListener(base Config, profileID string) (bool, error) {
	listeners, err := LoadListeners(base)
	if err != nil {
		return false, err
	}
	profileID = strings.ToLower(strings.TrimSpace(profileID))
	for _, listener := range listeners {
		if listener.DefaultProfileID == profileID {
			return true, nil
		}
	}
	return false, nil
}

func listenerConfig(base Config, spec ListenerSpec) (Config, error) {
	cfg := base
	cfg.ListenAddress = spec.ListenAddress
	cfg.Port = spec.Port
	if spec.RequireAuth != nil {
		cfg.RequireAuth = *spec.RequireAuth
	}
	return NormalizeConfig(cfg)
}
