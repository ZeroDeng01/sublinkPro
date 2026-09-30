package socks5

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"sublink/models"
	"sublink/utils"

	"github.com/metacubex/mihomo/constant"
)

const (
	settingEnabled                      = "socks5_enabled"
	settingListenAddress                = "socks5_listen_address"
	settingPort                         = "socks5_port"
	settingUsername                     = "socks5_username"
	settingPassword                     = "socks5_password"
	settingNodeID                       = "socks5_node_id"
	settingSelection                    = "socks5_selection"
	settingRequireAuth                  = "socks5_require_auth"
	settingMaxAttempts                  = "socks5_max_attempts"
	settingDialTimeoutSeconds           = "socks5_dial_timeout_seconds"
	settingFailureCooldownSeconds       = "socks5_failure_cooldown_seconds"
	settingSpecificFallback             = "socks5_specific_fallback"
	settingMaxConnections               = "socks5_max_connections"
	settingMaxConnectionsPerClient      = "socks5_max_connections_per_client"
	settingIdleTimeoutSeconds           = "socks5_idle_timeout_seconds"
	settingMaxConnectionDurationSeconds = "socks5_max_connection_duration_seconds"
	settingHealthCheckEnabled           = "socks5_health_check_enabled"
	settingHealthCheckIntervalSeconds   = "socks5_health_check_interval_seconds"
	settingHealthCheckTimeoutSeconds    = "socks5_health_check_timeout_seconds"
	settingCandidateGroups              = "socks5_candidate_groups"
	settingCandidateSources             = "socks5_candidate_sources"
	settingCandidateProtocols           = "socks5_candidate_protocols"
	settingCandidateCountries           = "socks5_candidate_countries"
	settingStickySessionEnabled         = "socks5_sticky_session_enabled"
	settingStickySessionMode            = "socks5_sticky_session_mode"
	settingStickySessionTTLSeconds      = "socks5_sticky_session_ttl_seconds"

	defaultListenAddress                = "127.0.0.1"
	defaultPort                         = 1080
	defaultSelection                    = "best"
	defaultMaxAttempts                  = 1
	defaultDialTimeoutSeconds           = 30
	defaultFailureCooldownSeconds       = 0
	defaultMaxConnections               = 256
	defaultMaxConnectionsPerClient      = 32
	defaultIdleTimeoutSeconds           = 600
	defaultMaxConnectionDurationSeconds = 0
	defaultHealthCheckIntervalSeconds   = 60
	defaultHealthCheckTimeoutSeconds    = 5
	defaultStickySessionMode            = "client_ip"
	defaultStickySessionTTLSeconds      = 1800
	handshakeTimeout                    = 15 * time.Second
)

// Config controls the SOCKS5 gateway.
type Config struct {
	Enabled                      bool
	ListenAddress                string
	Port                         int
	Username                     string
	Password                     string
	NodeID                       int
	Selection                    string
	RequireAuth                  bool
	MaxAttempts                  int
	DialTimeoutSeconds           int
	FailureCooldownSeconds       int
	SpecificFallback             bool
	MaxConnections               int
	MaxConnectionsPerClient      int
	IdleTimeoutSeconds           int
	MaxConnectionDurationSeconds int
	HealthCheckEnabled           bool
	HealthCheckIntervalSeconds   int
	HealthCheckTimeoutSeconds    int
	CandidateGroups              []string
	CandidateSources             []string
	CandidateProtocols           []string
	CandidateCountries           []string
	StickySessionEnabled         bool
	StickySessionMode            string
	StickySessionTTLSeconds      int
	ClearPassword                bool
}

// PublicConfig is safe to return from the settings API.
type PublicConfig struct {
	Enabled                      bool     `json:"enabled"`
	ListenAddress                string   `json:"listenAddress"`
	Port                         int      `json:"port"`
	Username                     string   `json:"username"`
	HasPassword                  bool     `json:"hasPassword"`
	MaskedPassword               string   `json:"maskedPassword,omitempty"`
	NodeID                       int      `json:"nodeId"`
	Selection                    string   `json:"selection"`
	RequireAuth                  bool     `json:"requireAuth"`
	MaxAttempts                  int      `json:"maxAttempts"`
	DialTimeoutSeconds           int      `json:"dialTimeoutSeconds"`
	FailureCooldownSeconds       int      `json:"failureCooldownSeconds"`
	SpecificFallback             bool     `json:"specificFallback"`
	MaxConnections               int      `json:"maxConnections"`
	MaxConnectionsPerClient      int      `json:"maxConnectionsPerClient"`
	IdleTimeoutSeconds           int      `json:"idleTimeoutSeconds"`
	MaxConnectionDurationSeconds int      `json:"maxConnectionDurationSeconds"`
	HealthCheckEnabled           bool     `json:"healthCheckEnabled"`
	HealthCheckIntervalSeconds   int      `json:"healthCheckIntervalSeconds"`
	HealthCheckTimeoutSeconds    int      `json:"healthCheckTimeoutSeconds"`
	CandidateGroups              []string `json:"candidateGroups"`
	CandidateSources             []string `json:"candidateSources"`
	CandidateProtocols           []string `json:"candidateProtocols"`
	CandidateCountries           []string `json:"candidateCountries"`
	StickySessionEnabled         bool     `json:"stickySessionEnabled"`
	StickySessionMode            string   `json:"stickySessionMode"`
	StickySessionTTLSeconds      int      `json:"stickySessionTtlSeconds"`
	Running                      bool     `json:"running"`
	BoundAddress                 string   `json:"boundAddress,omitempty"`
}

func defaultConfig() Config {
	return Config{ListenAddress: defaultListenAddress, Port: defaultPort, Selection: defaultSelection, RequireAuth: true, MaxAttempts: defaultMaxAttempts, DialTimeoutSeconds: defaultDialTimeoutSeconds, FailureCooldownSeconds: defaultFailureCooldownSeconds, MaxConnections: defaultMaxConnections, MaxConnectionsPerClient: defaultMaxConnectionsPerClient, IdleTimeoutSeconds: defaultIdleTimeoutSeconds, MaxConnectionDurationSeconds: defaultMaxConnectionDurationSeconds, HealthCheckIntervalSeconds: defaultHealthCheckIntervalSeconds, HealthCheckTimeoutSeconds: defaultHealthCheckTimeoutSeconds, StickySessionMode: defaultStickySessionMode, StickySessionTTLSeconds: defaultStickySessionTTLSeconds}
}

func LoadConfig() (Config, error) {
	cfg := defaultConfig()
	if value, err := models.GetSetting(settingEnabled); err == nil {
		cfg.Enabled = value == "true"
	}
	if value, err := models.GetSetting(settingListenAddress); err == nil && strings.TrimSpace(value) != "" {
		cfg.ListenAddress = strings.TrimSpace(value)
	}
	if value, err := models.GetSetting(settingPort); err == nil && strings.TrimSpace(value) != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return cfg, fmt.Errorf("invalid SOCKS5 port: %w", parseErr)
		}
		cfg.Port = parsed
	}
	if value, err := models.GetSetting(settingUsername); err == nil {
		cfg.Username = value
	}
	if value, err := models.GetSetting(settingPassword); err == nil && strings.TrimSpace(value) != "" {
		password, decryptErr := models.DecryptUserAISecret(value)
		if decryptErr != nil {
			return cfg, fmt.Errorf("decrypt SOCKS5 password: %w", decryptErr)
		}
		cfg.Password = password
	}
	if value, err := models.GetSetting(settingNodeID); err == nil && strings.TrimSpace(value) != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return cfg, fmt.Errorf("invalid SOCKS5 node ID: %w", parseErr)
		}
		cfg.NodeID = parsed
	}
	if value, err := models.GetSetting(settingSelection); err == nil && strings.TrimSpace(value) != "" {
		cfg.Selection = strings.ToLower(strings.TrimSpace(value))
	}
	if value, err := models.GetSetting(settingRequireAuth); err == nil && strings.TrimSpace(value) != "" {
		cfg.RequireAuth = value != "false"
	}
	if value, err := models.GetSetting(settingMaxAttempts); err == nil && strings.TrimSpace(value) != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return cfg, fmt.Errorf("invalid SOCKS5 max attempts: %w", parseErr)
		}
		cfg.MaxAttempts = parsed
	}
	if value, err := models.GetSetting(settingDialTimeoutSeconds); err == nil && strings.TrimSpace(value) != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return cfg, fmt.Errorf("invalid SOCKS5 dial timeout: %w", parseErr)
		}
		cfg.DialTimeoutSeconds = parsed
	}
	if value, err := models.GetSetting(settingFailureCooldownSeconds); err == nil && strings.TrimSpace(value) != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return cfg, fmt.Errorf("invalid SOCKS5 failure cooldown: %w", parseErr)
		}
		cfg.FailureCooldownSeconds = parsed
	}
	if value, err := models.GetSetting(settingSpecificFallback); err == nil && strings.TrimSpace(value) != "" {
		cfg.SpecificFallback = value == "true"
	}
	if value, err := models.GetSetting(settingHealthCheckEnabled); err == nil && strings.TrimSpace(value) != "" {
		cfg.HealthCheckEnabled = value == "true"
	}
	if value, err := models.GetSetting(settingStickySessionEnabled); err == nil && strings.TrimSpace(value) != "" {
		cfg.StickySessionEnabled = value == "true"
	}
	if value, err := models.GetSetting(settingStickySessionMode); err == nil && strings.TrimSpace(value) != "" {
		cfg.StickySessionMode = strings.ToLower(strings.TrimSpace(value))
	}
	for key, target := range map[string]*int{settingMaxConnections: &cfg.MaxConnections, settingMaxConnectionsPerClient: &cfg.MaxConnectionsPerClient, settingIdleTimeoutSeconds: &cfg.IdleTimeoutSeconds, settingMaxConnectionDurationSeconds: &cfg.MaxConnectionDurationSeconds, settingHealthCheckIntervalSeconds: &cfg.HealthCheckIntervalSeconds, settingHealthCheckTimeoutSeconds: &cfg.HealthCheckTimeoutSeconds, settingStickySessionTTLSeconds: &cfg.StickySessionTTLSeconds} {
		if value, err := models.GetSetting(key); err == nil && strings.TrimSpace(value) != "" {
			parsed, parseErr := strconv.Atoi(value)
			if parseErr != nil {
				return cfg, fmt.Errorf("invalid SOCKS5 setting %s: %w", key, parseErr)
			}
			*target = parsed
		}
	}
	for key, target := range map[string]*[]string{
		settingCandidateGroups:    &cfg.CandidateGroups,
		settingCandidateSources:   &cfg.CandidateSources,
		settingCandidateProtocols: &cfg.CandidateProtocols,
		settingCandidateCountries: &cfg.CandidateCountries,
	} {
		values, loadErr := loadStringListSetting(key)
		if loadErr != nil {
			return cfg, loadErr
		}
		*target = values
	}
	return NormalizeConfig(cfg)
}

func NormalizeConfig(cfg Config) (Config, error) {
	if strings.TrimSpace(cfg.ListenAddress) == "" {
		cfg.ListenAddress = defaultListenAddress
	}
	cfg.ListenAddress = strings.TrimSpace(cfg.ListenAddress)
	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return cfg, errors.New("SOCKS5 port must be between 1 and 65535")
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Selection)) {
	case "", "best":
		cfg.Selection = defaultSelection
	case "random", "round_robin", "smart", "specific":
		cfg.Selection = strings.ToLower(strings.TrimSpace(cfg.Selection))
	default:
		return cfg, errors.New("SOCKS5 selection must be best, random, round_robin, smart, or specific")
	}
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = defaultMaxAttempts
	}
	if cfg.MaxAttempts < 1 || cfg.MaxAttempts > 5 {
		return cfg, errors.New("SOCKS5 max attempts must be between 1 and 5")
	}
	if cfg.DialTimeoutSeconds == 0 {
		cfg.DialTimeoutSeconds = defaultDialTimeoutSeconds
	}
	if cfg.DialTimeoutSeconds < 1 || cfg.DialTimeoutSeconds > 120 {
		return cfg, errors.New("SOCKS5 dial timeout must be between 1 and 120 seconds")
	}
	if cfg.FailureCooldownSeconds < 0 || cfg.FailureCooldownSeconds > 3600 {
		return cfg, errors.New("SOCKS5 failure cooldown must be between 0 and 3600 seconds")
	}
	if cfg.Enabled && !cfg.RequireAuth && !isLoopbackListenAddress(cfg.ListenAddress) {
		return cfg, errors.New("SOCKS5 authentication can only be disabled on a loopback listen address")
	}
	if cfg.Enabled && cfg.RequireAuth && (strings.TrimSpace(cfg.Username) == "" || cfg.Password == "") {
		return cfg, errors.New("SOCKS5 username and password are required when authentication is enabled")
	}
	if len(cfg.Username) > 255 || len(cfg.Password) > 255 {
		return cfg, errors.New("SOCKS5 username and password must be at most 255 bytes")
	}
	if cfg.Selection == "specific" && cfg.NodeID <= 0 {
		return cfg, errors.New("a node ID is required for specific SOCKS5 selection")
	}
	if cfg.Selection == "specific" && cfg.SpecificFallback && cfg.MaxAttempts < 2 {
		return cfg, errors.New("SOCKS5 max attempts must be at least 2 when specific-node fallback is enabled")
	}
	if cfg.MaxConnections == 0 {
		cfg.MaxConnections = defaultMaxConnections
	}
	if cfg.MaxConnections < 1 || cfg.MaxConnections > 10000 {
		return cfg, errors.New("SOCKS5 max connections must be between 1 and 10000")
	}
	if cfg.MaxConnectionsPerClient == 0 {
		cfg.MaxConnectionsPerClient = defaultMaxConnectionsPerClient
	}
	if cfg.MaxConnectionsPerClient < 1 || cfg.MaxConnectionsPerClient > cfg.MaxConnections {
		return cfg, errors.New("SOCKS5 per-client connection limit is invalid")
	}
	if cfg.IdleTimeoutSeconds < 0 || cfg.IdleTimeoutSeconds > 86400 {
		return cfg, errors.New("SOCKS5 idle timeout must be between 0 and 86400 seconds")
	}
	if cfg.MaxConnectionDurationSeconds < 0 || cfg.MaxConnectionDurationSeconds > 604800 {
		return cfg, errors.New("SOCKS5 max connection duration must be between 0 and 604800 seconds")
	}
	if cfg.HealthCheckIntervalSeconds == 0 {
		cfg.HealthCheckIntervalSeconds = defaultHealthCheckIntervalSeconds
	}
	if cfg.HealthCheckIntervalSeconds < 10 || cfg.HealthCheckIntervalSeconds > 3600 {
		return cfg, errors.New("SOCKS5 health check interval must be between 10 and 3600 seconds")
	}
	if cfg.HealthCheckTimeoutSeconds == 0 {
		cfg.HealthCheckTimeoutSeconds = defaultHealthCheckTimeoutSeconds
	}
	if cfg.HealthCheckTimeoutSeconds < 1 || cfg.HealthCheckTimeoutSeconds > 30 {
		return cfg, errors.New("SOCKS5 health check timeout must be between 1 and 30 seconds")
	}
	switch strings.ToLower(strings.TrimSpace(cfg.StickySessionMode)) {
	case "", "client_ip":
		cfg.StickySessionMode = defaultStickySessionMode
	case "username":
		cfg.StickySessionMode = "username"
	default:
		return cfg, errors.New("SOCKS5 sticky session mode must be client_ip or username")
	}
	if cfg.StickySessionTTLSeconds == 0 {
		cfg.StickySessionTTLSeconds = defaultStickySessionTTLSeconds
	}
	if cfg.StickySessionTTLSeconds < 60 || cfg.StickySessionTTLSeconds > 604800 {
		return cfg, errors.New("SOCKS5 sticky session TTL must be between 60 and 604800 seconds")
	}
	if cfg.StickySessionEnabled && cfg.StickySessionMode == "username" && !cfg.RequireAuth {
		return cfg, errors.New("SOCKS5 username sticky sessions require authentication")
	}
	cfg.CandidateGroups = normalizeStringList(cfg.CandidateGroups, nil)
	cfg.CandidateSources = normalizeStringList(cfg.CandidateSources, nil)
	cfg.CandidateProtocols = normalizeStringList(cfg.CandidateProtocols, strings.ToLower)
	cfg.CandidateCountries = normalizeStringList(cfg.CandidateCountries, strings.ToUpper)
	for name, values := range map[string][]string{
		"groups": cfg.CandidateGroups, "sources": cfg.CandidateSources,
		"protocols": cfg.CandidateProtocols, "countries": cfg.CandidateCountries,
	} {
		if len(values) > 256 {
			return cfg, fmt.Errorf("SOCKS5 candidate %s must contain at most 256 values", name)
		}
		for _, value := range values {
			if len(value) > 255 {
				return cfg, fmt.Errorf("SOCKS5 candidate %s values must be at most 255 bytes", name)
			}
		}
	}
	return cfg, nil
}

func normalizeStringList(values []string, transform func(string) string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if transform != nil {
			value = transform(value)
		}
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func loadStringListSetting(key string) ([]string, error) {
	value, err := models.GetSetting(key)
	if err != nil || strings.TrimSpace(value) == "" {
		return []string{}, nil
	}
	var values []string
	if err := json.Unmarshal([]byte(value), &values); err != nil {
		return nil, fmt.Errorf("invalid SOCKS5 setting %s: %w", key, err)
	}
	return values, nil
}

func encodeStringListSetting(values []string) string {
	encoded, _ := json.Marshal(values)
	return string(encoded)
}

func SaveConfig(input Config) (Config, error) {
	cfg, err := NormalizeConfig(input)
	if err != nil {
		return cfg, err
	}
	accounts, err := loadAccounts(cfg)
	if err != nil {
		return cfg, err
	}
	for _, account := range accounts {
		if strings.EqualFold(account.Username, strings.TrimSpace(cfg.Username)) {
			return cfg, errors.New("SOCKS5 gateway username conflicts with an independent account")
		}
	}
	values := map[string]string{
		settingEnabled:                      strconv.FormatBool(cfg.Enabled),
		settingListenAddress:                cfg.ListenAddress,
		settingPort:                         strconv.Itoa(cfg.Port),
		settingUsername:                     strings.TrimSpace(cfg.Username),
		settingNodeID:                       strconv.Itoa(cfg.NodeID),
		settingSelection:                    cfg.Selection,
		settingRequireAuth:                  strconv.FormatBool(cfg.RequireAuth),
		settingMaxAttempts:                  strconv.Itoa(cfg.MaxAttempts),
		settingDialTimeoutSeconds:           strconv.Itoa(cfg.DialTimeoutSeconds),
		settingFailureCooldownSeconds:       strconv.Itoa(cfg.FailureCooldownSeconds),
		settingSpecificFallback:             strconv.FormatBool(cfg.SpecificFallback),
		settingMaxConnections:               strconv.Itoa(cfg.MaxConnections),
		settingMaxConnectionsPerClient:      strconv.Itoa(cfg.MaxConnectionsPerClient),
		settingIdleTimeoutSeconds:           strconv.Itoa(cfg.IdleTimeoutSeconds),
		settingMaxConnectionDurationSeconds: strconv.Itoa(cfg.MaxConnectionDurationSeconds),
		settingHealthCheckEnabled:           strconv.FormatBool(cfg.HealthCheckEnabled),
		settingHealthCheckIntervalSeconds:   strconv.Itoa(cfg.HealthCheckIntervalSeconds),
		settingHealthCheckTimeoutSeconds:    strconv.Itoa(cfg.HealthCheckTimeoutSeconds),
		settingCandidateGroups:              encodeStringListSetting(cfg.CandidateGroups),
		settingCandidateSources:             encodeStringListSetting(cfg.CandidateSources),
		settingCandidateProtocols:           encodeStringListSetting(cfg.CandidateProtocols),
		settingCandidateCountries:           encodeStringListSetting(cfg.CandidateCountries),
		settingStickySessionEnabled:         strconv.FormatBool(cfg.StickySessionEnabled),
		settingStickySessionMode:            cfg.StickySessionMode,
		settingStickySessionTTLSeconds:      strconv.Itoa(cfg.StickySessionTTLSeconds),
	}
	if cfg.ClearPassword {
		values[settingPassword] = ""
	} else if cfg.Password != "" {
		encrypted, encryptErr := models.EncryptUserAISecret(cfg.Password)
		if encryptErr != nil {
			return cfg, fmt.Errorf("encrypt SOCKS5 password: %w", encryptErr)
		}
		values[settingPassword] = encrypted
	} else if existing, getErr := models.GetSetting(settingPassword); getErr != nil || existing == "" {
		values[settingPassword] = ""
	}
	for key, value := range values {
		if err := models.SetSetting(key, value); err != nil {
			return cfg, err
		}
	}
	return LoadConfig()
}

func ToPublicConfig(cfg Config, running bool, boundAddress string) PublicConfig {
	masked := ""
	if cfg.Password != "" {
		masked = "••••••••"
	}
	return PublicConfig{
		Enabled: cfg.Enabled, ListenAddress: cfg.ListenAddress, Port: cfg.Port, Username: cfg.Username,
		HasPassword: cfg.Password != "", MaskedPassword: masked, NodeID: cfg.NodeID, Selection: cfg.Selection,
		RequireAuth: cfg.RequireAuth, MaxAttempts: cfg.MaxAttempts, DialTimeoutSeconds: cfg.DialTimeoutSeconds,
		FailureCooldownSeconds: cfg.FailureCooldownSeconds, SpecificFallback: cfg.SpecificFallback,
		MaxConnections: cfg.MaxConnections, MaxConnectionsPerClient: cfg.MaxConnectionsPerClient,
		IdleTimeoutSeconds: cfg.IdleTimeoutSeconds, MaxConnectionDurationSeconds: cfg.MaxConnectionDurationSeconds,
		HealthCheckEnabled: cfg.HealthCheckEnabled, HealthCheckIntervalSeconds: cfg.HealthCheckIntervalSeconds,
		HealthCheckTimeoutSeconds: cfg.HealthCheckTimeoutSeconds,
		CandidateGroups:           append([]string{}, cfg.CandidateGroups...), CandidateSources: append([]string{}, cfg.CandidateSources...),
		CandidateProtocols: append([]string{}, cfg.CandidateProtocols...), CandidateCountries: append([]string{}, cfg.CandidateCountries...),
		StickySessionEnabled: cfg.StickySessionEnabled, StickySessionMode: cfg.StickySessionMode,
		StickySessionTTLSeconds: cfg.StickySessionTTLSeconds,
		Running:                 running, BoundAddress: boundAddress,
	}
}

// DialFunc allows tests and future routing implementations to replace mihomo dialing.
type DialFunc func(ctx context.Context, node models.Node, host string, port uint16) (net.Conn, error)

// Server serves SOCKS5 CONNECT requests over a supplied listener.
type Server struct {
	cfg              Config
	listenerID       string
	defaultProfileID string
	dial             DialFunc
	pool             *adapterPool
	router           *nodeRouter
	profiles         *routingProfileStore
	accounts         *accountStore
	registry         *sessionRegistry
	healthProbe      HealthProbeFunc
	healthSweep      healthSweepState
	connMu           sync.Mutex
	activeConns      int
	connZero         chan struct{}
	closing          bool
	closeOnce        sync.Once
	closed           chan struct{}
}

func NewServer(cfg Config, dial DialFunc) (*Server, error) {
	return newServerForListener(cfg, dial, "", defaultProfileID)
}

func newServerForListener(cfg Config, dial DialFunc, listenerID, defaultProfile string) (*Server, error) {
	normalized, err := NormalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	profiles, err := newRoutingProfileStore(normalized)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(defaultProfile) == "" {
		defaultProfile = defaultProfileID
	}
	if _, _, ok := profiles.resolve(defaultProfile); !ok {
		return nil, errors.New("SOCKS5 listener default routing profile was not found or is disabled")
	}
	accounts, err := newAccountStore(normalized)
	if err != nil {
		return nil, err
	}
	registry := newSessionRegistry(normalized.MaxConnections, normalized.MaxConnectionsPerClient)
	registry.setListenerID(listenerID)
	server := &Server{cfg: normalized, listenerID: listenerID, defaultProfileID: defaultProfile, dial: dial, router: newNodeRouter(), profiles: profiles, accounts: accounts, registry: registry, closed: make(chan struct{})}
	if server.dial == nil {
		server.pool = newAdapterPool(nil)
		server.dial = server.dialWithPool
	}
	server.healthProbe = server.defaultHealthProbe
	return server, nil
}

func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	s.startHealthLoop(ctx)
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		s.connMu.Lock()
		if s.closing {
			s.connMu.Unlock()
			_ = conn.Close()
			continue
		}
		if s.activeConns == 0 {
			s.connZero = make(chan struct{})
		}
		s.activeConns++
		s.connMu.Unlock()
		go func() {
			defer s.finishConnection()
			s.serveConn(ctx, conn)
		}()
	}
}

func (s *Server) finishConnection() {
	s.connMu.Lock()
	if s.activeConns > 0 {
		s.activeConns--
		if s.activeConns == 0 && s.connZero != nil {
			close(s.connZero)
			s.connZero = nil
		}
	}
	s.connMu.Unlock()
}

func (s *Server) serveConn(ctx context.Context, client net.Conn) {
	session, connCtx, ok := s.registry.register(ctx, client)
	if !ok {
		_ = client.Close()
		return
	}
	ctx = connCtx
	watchDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			session.close()
		case <-watchDone:
		}
	}()
	established := false
	releaseRuntime := func() {}
	defer func() {
		releaseRuntime()
		close(watchDone)
		if !established {
			s.registry.recordFailure()
		}
		s.registry.remove(session.ID)
		_ = client.Close()
	}()
	_ = client.SetDeadline(time.Now().Add(handshakeTimeout))
	identity := defaultRoutingIdentity()
	identity.ProfileID = s.defaultProfileID
	if s.cfg.RequireAuth {
		var authErr error
		identity, authErr = s.authenticate(client)
		if authErr != nil {
			return
		}
	} else if err := negotiateNoAuth(client); err != nil {
		return
	}
	profile, routingCfg, ok := s.profiles.resolve(identity.ProfileID)
	if !ok {
		_ = writeReply(client, 0x01)
		return
	}
	identity.ProfileName = profile.Name
	session.setRoutingIdentity(identity)
	host, port, err := readConnectRequest(client)
	if err != nil {
		return
	}
	session.mu.Lock()
	session.Target = net.JoinHostPort(host, strconv.Itoa(int(port)))
	session.mu.Unlock()
	affinityKey := routingProfileStickyKey(profile, routingCfg, identity, session.ClientAddress)
	nodes, err := s.router.candidatesForScope(routingCfg, affinityKey, profile.ID)
	if err != nil {
		_ = writeReply(client, 0x01)
		return
	}
	var upstream net.Conn
	for _, node := range nodes {
		dialCtx, cancel := context.WithTimeout(ctx, time.Duration(routingCfg.DialTimeoutSeconds)*time.Second)
		upstream, err = s.dial(dialCtx, node, host, port)
		cancel()
		if err == nil {
			s.router.health.recordSuccess(node)
			releaseRuntime = s.router.runtime.start(node)
			s.router.sticky.bind(affinityKey, node, time.Duration(routingCfg.StickySessionTTLSeconds)*time.Second)
			session.setUpstream(upstream, node)
			break
		}
		s.router.runtime.recordFailure(node)
		s.router.health.recordFailureReason(node, time.Duration(routingCfg.FailureCooldownSeconds)*time.Second, s.sanitizeProbeError(node, err))
	}
	if upstream == nil {
		s.router.sticky.remove(affinityKey)
		_ = writeReply(client, 0x01)
		return
	}
	defer func() { _ = upstream.Close() }()
	_ = client.SetDeadline(time.Time{})
	if err := writeReply(client, 0x00); err != nil {
		return
	}
	established = true
	s.registry.recordSuccess()
	pumpWithSession(ctx, session, client, upstream, s.cfg.IdleTimeoutSeconds, s.cfg.MaxConnectionDurationSeconds)
}

func (s *Server) authenticate(conn net.Conn) (routingIdentity, error) {
	invalidIdentity := routingIdentity{}
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil || header[0] != 0x05 {
		return invalidIdentity, errors.New("invalid SOCKS5 greeting")
	}
	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(conn, methods); err != nil {
		return invalidIdentity, err
	}
	offered := false
	for _, method := range methods {
		if method == 0x02 {
			offered = true
			break
		}
	}
	if !offered {
		_, _ = conn.Write([]byte{0x05, 0xff})
		return invalidIdentity, errors.New("username/password authentication not offered")
	}
	if _, err := conn.Write([]byte{0x05, 0x02}); err != nil {
		return invalidIdentity, err
	}
	authHeader := make([]byte, 2)
	if _, err := io.ReadFull(conn, authHeader); err != nil || authHeader[0] != 0x01 {
		return invalidIdentity, errors.New("invalid SOCKS5 auth version")
	}
	username := make([]byte, int(authHeader[1]))
	if _, err := io.ReadFull(conn, username); err != nil {
		return invalidIdentity, err
	}
	length := make([]byte, 1)
	if _, err := io.ReadFull(conn, length); err != nil {
		return invalidIdentity, err
	}
	password := make([]byte, int(length[0]))
	if _, err := io.ReadFull(conn, password); err != nil {
		return invalidIdentity, err
	}

	rawUsername := string(username)
	if account, found, authenticated := s.accounts.authenticate(rawUsername, password); found {
		profile, _, profileFound := s.profiles.resolve(account.ProfileID)
		if !authenticated || !profileFound {
			_, _ = conn.Write([]byte{0x01, 0x01})
			return invalidIdentity, errors.New("invalid SOCKS5 credentials")
		}
		if _, err := conn.Write([]byte{0x01, 0x00}); err != nil {
			return invalidIdentity, err
		}
		return routingIdentity{
			RawUsername: rawUsername, ProfileID: profile.ID, ProfileName: profile.Name,
			Account: account.Username, Extended: true,
		}, nil
	}

	baseUsername := []byte(s.cfg.Username)
	baseMatch := subtle.ConstantTimeCompare(username, baseUsername) == 1
	extendedBaseMatch := len(username) > len(baseUsername)+1 && username[len(baseUsername)] == '@' &&
		subtle.ConstantTimeCompare(username[:len(baseUsername)], baseUsername) == 1
	profileID, account, extended, parsed := parseRoutingUsername(s.cfg.Username, rawUsername)
	if parsed && !extended && profileID == defaultProfileID {
		profileID = s.defaultProfileID
	}
	profile, _, profileFound := s.profiles.resolve(profileID)
	usernameOK := (baseMatch || extendedBaseMatch) && parsed && profileFound
	passwordOK := subtle.ConstantTimeCompare(password, []byte(s.cfg.Password)) == 1
	if !usernameOK || !passwordOK {
		_, _ = conn.Write([]byte{0x01, 0x01})
		return invalidIdentity, errors.New("invalid SOCKS5 credentials")
	}
	if _, err := conn.Write([]byte{0x01, 0x00}); err != nil {
		return invalidIdentity, err
	}
	return routingIdentity{
		RawUsername: rawUsername, ProfileID: profile.ID, ProfileName: profile.Name,
		Account: account, Extended: extended,
	}, nil
}

func negotiateNoAuth(conn net.Conn) error {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil || header[0] != 0x05 {
		return errors.New("invalid SOCKS5 greeting")
	}
	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(conn, methods); err != nil {
		return err
	}
	for _, method := range methods {
		if method == 0x00 {
			_, err := conn.Write([]byte{0x05, 0x00})
			return err
		}
	}
	_, _ = conn.Write([]byte{0x05, 0xff})
	return errors.New("no-auth method not offered")
}

func readConnectRequest(conn net.Conn) (string, uint16, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil || header[0] != 0x05 {
		return "", 0, errors.New("invalid SOCKS5 request")
	}
	if header[1] != 0x01 {
		_ = writeReply(conn, 0x07)
		return "", 0, errors.New("SOCKS5 command is not CONNECT")
	}
	var host string
	switch header[3] {
	case 0x01:
		address := make([]byte, net.IPv4len)
		if _, err := io.ReadFull(conn, address); err != nil {
			return "", 0, err
		}
		host = net.IP(address).String()
	case 0x03:
		length := make([]byte, 1)
		if _, err := io.ReadFull(conn, length); err != nil {
			return "", 0, err
		}
		address := make([]byte, int(length[0]))
		if _, err := io.ReadFull(conn, address); err != nil {
			return "", 0, err
		}
		host = string(address)
	case 0x04:
		address := make([]byte, net.IPv6len)
		if _, err := io.ReadFull(conn, address); err != nil {
			return "", 0, err
		}
		host = net.IP(address).String()
	default:
		_ = writeReply(conn, 0x08)
		return "", 0, errors.New("unsupported SOCKS5 address type")
	}
	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBytes); err != nil {
		return "", 0, err
	}
	return host, uint16(portBytes[0])<<8 | uint16(portBytes[1]), nil
}

func writeReply(conn net.Conn, code byte) error {
	_, err := conn.Write([]byte{0x05, code, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	return err
}

func (s *Server) dialWithPool(ctx context.Context, node models.Node, host string, port uint16) (net.Conn, error) {
	adapter, lease, err := s.pool.acquire(node)
	if err != nil {
		return nil, err
	}
	conn, err := adapter.DialContext(ctx, &constant.Metadata{NetWork: constant.TCP, Type: constant.SOCKS5, Host: host, DstPort: port})
	if err != nil {
		lease.release(true)
		return nil, err
	}
	return &pooledConn{Conn: conn, lease: lease}, nil
}

func (s *Server) Close() {
	s.closeOnce.Do(func() {
		s.connMu.Lock()
		s.closing = true
		s.connMu.Unlock()
		s.stopHealthChecks()
		if s.registry != nil {
			s.registry.closeAll()
		}
		s.connMu.Lock()
		activeConnections := s.activeConns
		done := s.connZero
		s.connMu.Unlock()
		if activeConnections > 0 && done != nil {
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-done:
				if !timer.Stop() {
					<-timer.C
				}
			case <-timer.C:
			}
		}
		if s.pool != nil {
			s.pool.close()
		}
		close(s.closed)
	})
	<-s.closed
}

func (s *Server) Connections() []ConnectionSnapshot { return s.registry.snapshot() }
func (s *Server) Stats() GatewaySnapshot            { return s.registry.snapshotStats() }

func isLoopbackListenAddress(address string) bool {
	host := strings.TrimSpace(strings.Trim(address, "[]"))
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// Manager owns the optional long-lived listeners.
type listenerRuntime struct {
	spec     ListenerSpec
	cfg      Config
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
	server   *Server
}

type Manager struct {
	mu        sync.Mutex
	cfg       Config
	specs     []ListenerSpec
	dial      DialFunc
	listeners map[string]*listenerRuntime
}

var defaultManager = &Manager{}

func DefaultManager() *Manager { return defaultManager }

func (m *Manager) Apply(cfg Config) error {
	normalized, err := NormalizeConfig(cfg)
	if err != nil {
		return err
	}
	listeners, err := LoadListeners(normalized)
	if err != nil {
		return err
	}
	return m.ApplyListeners(normalized, listeners)
}

func (m *Manager) ApplyListeners(cfg Config, listeners []ListenerSpec) error {
	normalized, err := NormalizeConfig(cfg)
	if err != nil {
		return err
	}
	normalizedListeners, err := normalizeListeners(normalized, listeners)
	if err != nil {
		return err
	}
	m.mu.Lock()
	old := m.listeners
	dial := m.dial
	m.mu.Unlock()

	prepared := make(map[string]*listenerRuntime)
	if normalized.Enabled {
		for _, spec := range normalizedListeners {
			if !spec.Enabled {
				continue
			}
			listenerCfg, cfgErr := listenerConfig(normalized, spec)
			if cfgErr != nil {
				closeListenerRuntimes(prepared)
				return cfgErr
			}
			endpoint := net.JoinHostPort(spec.ListenAddress, strconv.Itoa(spec.Port))
			if existing := old[spec.ID]; existing != nil && existing.listener != nil && existing.listener.Addr().String() == endpoint && configsEqual(existing.cfg, listenerCfg) && listenerSpecsEqual(existing.spec, spec) {
				prepared[spec.ID] = existing
				continue
			}
			server, serverErr := newServerForListener(listenerCfg, dial, spec.ID, spec.DefaultProfileID)
			if serverErr != nil {
				closeNewListenerRuntimes(prepared, old)
				return serverErr
			}
			var listenConfig net.ListenConfig
			listener, listenErr := listenConfig.Listen(context.Background(), "tcp", endpoint)
			if listenErr != nil {
				existing := old[spec.ID]
				if existing != nil && existing.listener != nil && existing.listener.Addr().String() == endpoint {
					closeListenerRuntime(existing)
					listener, listenErr = listenConfig.Listen(context.Background(), "tcp", endpoint)
					if listenErr != nil {
						m.restoreRuntime(spec.ID, existing, dial)
					}
				}
			}
			if listenErr != nil {
				server.Close()
				closeNewListenerRuntimes(prepared, old)
				return fmt.Errorf("listen SOCKS5 %s: %w", spec.ID, listenErr)
			}
			ctx, cancel := context.WithCancel(context.Background())
			prepared[spec.ID] = &listenerRuntime{spec: spec, cfg: listenerCfg, listener: listener, ctx: ctx, cancel: cancel, server: server}
		}
	}

	m.mu.Lock()
	m.cfg = normalized
	m.specs = append([]ListenerSpec{}, normalizedListeners...)
	m.listeners = prepared
	m.mu.Unlock()

	for id, runtime := range old {
		if prepared[id] == runtime {
			continue
		}
		closeListenerRuntime(runtime)
	}
	for id, runtime := range prepared {
		if old[id] == runtime {
			continue
		}
		go m.serveRuntime(id, runtime)
	}
	return nil
}

func configsEqual(a, b Config) bool {
	return reflect.DeepEqual(a, b)
}

func listenerSpecsEqual(a, b ListenerSpec) bool {
	return reflect.DeepEqual(a, b)
}

func (m *Manager) restoreRuntime(id string, previous *listenerRuntime, dial DialFunc) {
	if previous == nil {
		return
	}
	server, err := newServerForListener(previous.cfg, dial, previous.spec.ID, previous.spec.DefaultProfileID)
	if err != nil {
		return
	}
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(context.Background(), "tcp", net.JoinHostPort(previous.spec.ListenAddress, strconv.Itoa(previous.spec.Port)))
	if err != nil {
		server.Close()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	restored := &listenerRuntime{spec: previous.spec, cfg: previous.cfg, listener: listener, ctx: ctx, cancel: cancel, server: server}
	m.mu.Lock()
	if m.listeners == nil {
		m.listeners = make(map[string]*listenerRuntime)
	}
	m.listeners[id] = restored
	m.mu.Unlock()
	go m.serveRuntime(id, restored)
}

func closeListenerRuntime(runtime *listenerRuntime) {
	if runtime == nil {
		return
	}
	if runtime.cancel != nil {
		runtime.cancel()
	}
	if runtime.listener != nil {
		_ = runtime.listener.Close()
	}
	if runtime.server != nil {
		runtime.server.Close()
	}
}

func closeListenerRuntimes(runtimes map[string]*listenerRuntime) {
	for _, runtime := range runtimes {
		closeListenerRuntime(runtime)
	}
}

func closeNewListenerRuntimes(runtimes, old map[string]*listenerRuntime) {
	for id, runtime := range runtimes {
		if old[id] != runtime {
			closeListenerRuntime(runtime)
		}
	}
}

func (m *Manager) serveRuntime(id string, runtime *listenerRuntime) {
	if serveErr := runtime.server.Serve(runtime.ctx, runtime.listener); serveErr != nil {
		utils.Warn("SOCKS5 listener %s stopped: %v", id, serveErr)
	}
	runtime.server.Close()
	m.mu.Lock()
	if m.listeners[id] == runtime {
		delete(m.listeners, id)
	}
	m.mu.Unlock()
}

func (m *Manager) StartFromSettings() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	return m.Apply(cfg)
}

func (m *Manager) Stop() {
	m.mu.Lock()
	runtimes := m.listeners
	m.listeners = nil
	m.mu.Unlock()
	closeListenerRuntimes(runtimes)
}

func (m *Manager) runtimeSnapshot() []*listenerRuntime {
	m.mu.Lock()
	result := make([]*listenerRuntime, 0, len(m.listeners))
	for _, runtime := range m.listeners {
		result = append(result, runtime)
	}
	m.mu.Unlock()
	sort.Slice(result, func(i, j int) bool { return result[i].spec.ID < result[j].spec.ID })
	return result
}

func (m *Manager) GatewaySnapshot() GatewaySnapshot {
	var result GatewaySnapshot
	for _, runtime := range m.runtimeSnapshot() {
		stats := runtime.server.Stats()
		result.ActiveConnections += stats.ActiveConnections
		result.TotalConnections += stats.TotalConnections
		result.SuccessfulConnections += stats.SuccessfulConnections
		result.FailedConnections += stats.FailedConnections
		result.UploadBytes += stats.UploadBytes
		result.DownloadBytes += stats.DownloadBytes
	}
	return result
}

func (m *Manager) Connections() []ConnectionSnapshot {
	result := make([]ConnectionSnapshot, 0)
	for _, runtime := range m.runtimeSnapshot() {
		result = append(result, runtime.server.Connections()...)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartedAt.Before(result[j].StartedAt) })
	return result
}

func (m *Manager) HealthSnapshot() HealthSnapshot {
	runtimes := m.runtimeSnapshot()
	if len(runtimes) == 0 {
		return HealthSnapshot{Nodes: []NodeHealthSnapshot{}}
	}
	return runtimes[0].server.HealthSnapshot()
}

func (m *Manager) ReloadAccounts() error {
	base, err := LoadConfig()
	if err != nil {
		return err
	}
	accounts, err := loadAccounts(base)
	if err != nil {
		return err
	}
	for _, runtime := range m.runtimeSnapshot() {
		runtime.server.accounts.replace(accounts)
	}
	return nil
}

func (m *Manager) ReloadRoutingProfiles() error {
	base, err := LoadConfig()
	if err != nil {
		return err
	}
	profiles, err := ListRoutingProfiles(base)
	if err != nil {
		return err
	}
	for _, runtime := range m.runtimeSnapshot() {
		runtime.server.profiles.replace(base, profiles)
	}
	return nil
}

func (m *Manager) RoutingSnapshot(query RoutingSnapshotQuery) (RoutingSnapshotPage, error) {
	runtimes := m.runtimeSnapshot()
	if len(runtimes) == 0 {
		return emptyRoutingSnapshot(query), nil
	}
	return runtimes[0].server.RoutingSnapshot(query)
}

func (m *Manager) ResetNodeRuntimeStats() {
	for _, runtime := range m.runtimeSnapshot() {
		runtime.server.router.runtime.reset()
	}
}

func (m *Manager) TriggerHealthProbe() (started bool, available bool) {
	runtimes := m.runtimeSnapshot()
	if len(runtimes) == 0 {
		return false, false
	}
	return runtimes[0].server.triggerHealthProbe()
}

func (m *Manager) CloseConnection(id string) bool {
	for _, runtime := range m.runtimeSnapshot() {
		if runtime.server.registry.close(id) {
			return true
		}
	}
	return false
}

func (m *Manager) CloseAllConnections() {
	for _, runtime := range m.runtimeSnapshot() {
		runtime.server.registry.closeAll()
	}
}

func (m *Manager) Status() (bool, string) {
	statuses := m.ListenerStatuses()
	if len(statuses) == 0 {
		return false, ""
	}
	for _, status := range statuses {
		if status.ID == defaultProfileID && status.Running {
			return true, status.BoundAddress
		}
	}
	for _, status := range statuses {
		if status.Running {
			return true, status.BoundAddress
		}
	}
	return false, ""
}

func (m *Manager) ListenerStatuses() []ListenerStatus {
	m.mu.Lock()
	base := m.cfg
	listeners := append([]ListenerSpec{}, m.specs...)
	runtimes := make(map[string]*listenerRuntime, len(m.listeners))
	for id, runtime := range m.listeners {
		runtimes[id] = runtime
	}
	m.mu.Unlock()
	if strings.TrimSpace(base.ListenAddress) == "" {
		if loaded, loadErr := LoadConfig(); loadErr == nil {
			base = loaded
		}
	}
	if len(listeners) == 0 {
		loaded, err := LoadListeners(base)
		if err == nil {
			listeners = loaded
		}
	}
	result := make([]ListenerStatus, 0, len(listeners))
	for _, spec := range listeners {
		status := ListenerStatus{ListenerSpec: spec}
		if runtime := runtimes[spec.ID]; runtime != nil && runtime.listener != nil {
			status.Running = true
			status.BoundAddress = runtime.listener.Addr().String()
		}
		result = append(result, status)
	}
	return result
}

func (m *Manager) SetDialFuncForTest(dial DialFunc) func() {
	m.mu.Lock()
	previous := m.dial
	m.dial = dial
	m.mu.Unlock()
	return func() {
		m.mu.Lock()
		m.dial = previous
		m.mu.Unlock()
	}
}
