package socks5

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"sublink/models"
)

const (
	healthProbeURL          = "https://cp.cloudflare.com/generate_204"
	healthProbeConcurrency  = 4
	maxHealthErrorLength    = 240
	healthSnapshotNodeLimit = 200
)

type HealthProbeFunc func(ctx context.Context, node models.Node, timeout time.Duration) (int, error)

type NodeHealthSnapshot struct {
	ListenerID          string     `json:"listenerId,omitempty"`
	NodeID              int        `json:"nodeId"`
	NodeName            string     `json:"nodeName"`
	Status              string     `json:"status"`
	LatencyMs           int        `json:"latencyMs"`
	LastCheckedAt       *time.Time `json:"lastCheckedAt,omitempty"`
	LastSuccessAt       *time.Time `json:"lastSuccessAt,omitempty"`
	LastFailureAt       *time.Time `json:"lastFailureAt,omitempty"`
	ConsecutiveFailures int        `json:"consecutiveFailures"`
	CooldownUntil       *time.Time `json:"cooldownUntil,omitempty"`
	LastError           string     `json:"lastError,omitempty"`
}

type HealthSummary struct {
	Running        bool       `json:"running"`
	LastSweepAt    *time.Time `json:"lastSweepAt,omitempty"`
	NextSweepAt    *time.Time `json:"nextSweepAt,omitempty"`
	HealthyNodes   int        `json:"healthyNodes"`
	UnhealthyNodes int        `json:"unhealthyNodes"`
	CheckingNodes  int        `json:"checkingNodes"`
	UnknownNodes   int        `json:"unknownNodes"`
	TotalNodes     int        `json:"totalNodes"`
}

type HealthSnapshot struct {
	HealthSummary
	Truncated bool                 `json:"truncated"`
	Nodes     []NodeHealthSnapshot `json:"nodes"`
}

func (summary *HealthSummary) count(status string) {
	summary.TotalNodes++
	switch status {
	case "healthy":
		summary.HealthyNodes++
	case "unhealthy":
		summary.UnhealthyNodes++
	case "checking":
		summary.CheckingNodes++
	default:
		summary.UnknownNodes++
	}
}

func monitoringHealthRank(state nodeHealthState, now time.Time) int {
	if state.ExcludeFromRouting || state.CooldownUntil.After(now) {
		return 0
	}
	switch state.Status {
	case "healthy":
		return 4
	case "checking":
		return 3
	case "unhealthy":
		return 1
	default:
		return 2
	}
}

func preferMonitoringHealth(candidate, current nodeHealthState, now time.Time) bool {
	// A reported failure is more informative than a listener with no observation yet.
	if candidate.NodeID == 0 || current.NodeID == 0 {
		return candidate.NodeID != 0
	}
	next, previous := monitoringHealthRank(candidate, now), monitoringHealthRank(current, now)
	if next != previous {
		return next > previous
	}
	return candidate.LastCheckedAt.After(current.LastCheckedAt)
}

type listenerHealthState struct {
	state      nodeHealthState
	listenerID string
}

func (m *Manager) collectHealth() (HealthSummary, map[string]listenerHealthState) {
	var summary HealthSummary
	nodes := make(map[string]listenerHealthState)
	for _, runtime := range m.runtimeSnapshot() {
		schedule := runtime.server.probeSchedule()
		summary.Running = summary.Running || schedule.Running
		if schedule.LastSweepAt != nil && (summary.LastSweepAt == nil || schedule.LastSweepAt.After(*summary.LastSweepAt)) {
			summary.LastSweepAt = schedule.LastSweepAt
		}
		if schedule.NextSweepAt != nil && (summary.NextSweepAt == nil || schedule.NextSweepAt.Before(*summary.NextSweepAt)) {
			summary.NextSweepAt = schedule.NextSweepAt
		}
		health := runtime.server.router.health
		now := health.now()
		health.mu.Lock()
		for key, state := range health.states {
			previous, found := nodes[key]
			if !found || preferMonitoringHealth(*state, previous.state, now) {
				nodes[key] = listenerHealthState{state: *state, listenerID: runtime.spec.ID}
			}
		}
		health.mu.Unlock()
	}
	for _, node := range nodes {
		summary.count(node.state.Status)
	}
	return summary, nodes
}

func sanitizeHealthError(message string) string {
	message = strings.TrimSpace(message)
	runes := []rune(message)
	if len(runes) > maxHealthErrorLength {
		message = string(runes[:maxHealthErrorLength])
	}
	return message
}

func (s *Server) sanitizeProbeError(node models.Node, err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	for _, secret := range []string{node.Link, s.cfg.Password} {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "[redacted]")
		}
	}
	return sanitizeHealthError(message)
}

func timePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	copyValue := value
	return &copyValue
}

func (h *nodeHealth) snapshot() []NodeHealthSnapshot {
	h.mu.Lock()
	now := h.now()
	result := make([]NodeHealthSnapshot, 0, len(h.states))
	for _, state := range h.states {
		result = append(result, state.snapshot(now))
	}
	h.mu.Unlock()
	sortHealthSnapshots(result)
	return result
}

func (h nodeHealthState) snapshot(now time.Time) NodeHealthSnapshot {
	result := NodeHealthSnapshot{
		NodeID: h.NodeID, NodeName: h.NodeName, Status: h.Status, LatencyMs: h.LatencyMs,
		LastCheckedAt: timePointer(h.LastCheckedAt), LastSuccessAt: timePointer(h.LastSuccessAt),
		LastFailureAt: timePointer(h.LastFailureAt), ConsecutiveFailures: h.ConsecutiveFailures, LastError: h.LastError,
	}
	if h.CooldownUntil.After(now) {
		result.CooldownUntil = timePointer(h.CooldownUntil)
	}
	return result
}

func sortHealthSnapshots(nodes []NodeHealthSnapshot) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].NodeName == nodes[j].NodeName {
			return nodes[i].NodeID < nodes[j].NodeID
		}
		return nodes[i].NodeName < nodes[j].NodeName
	})
}

type healthSweepState struct {
	mu          sync.Mutex
	running     bool
	stopping    bool
	lastSweepAt time.Time
	nextSweepAt time.Time
	ctx         context.Context
	cancel      context.CancelFunc
	done        chan struct{}
}

func (s *Server) setHealthContext(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	s.healthSweep.mu.Lock()
	if s.healthSweep.cancel != nil {
		s.healthSweep.cancel()
	}
	s.healthSweep.ctx = ctx
	s.healthSweep.cancel = cancel
	s.healthSweep.stopping = false
	s.healthSweep.mu.Unlock()
}

func (s *Server) startHealthLoop(ctx context.Context) {
	s.setHealthContext(ctx)
	s.healthSweep.mu.Lock()
	ctx = s.healthSweep.ctx
	s.healthSweep.mu.Unlock()
	if !s.cfg.HealthCheckEnabled {
		return
	}
	_, _ = s.triggerHealthProbe()
	go func() {
		ticker := time.NewTicker(time.Duration(s.cfg.HealthCheckIntervalSeconds) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, _ = s.triggerHealthProbe()
			}
		}
	}()
}

func (s *Server) triggerHealthProbe() (started bool, available bool) {
	return s.triggerHealthProbeFor(nil, false)
}

func (s *Server) triggerHealthProbeFor(nodes []models.Node, partial bool) (started bool, available bool) {
	s.healthSweep.mu.Lock()
	if s.healthSweep.stopping || s.healthSweep.ctx == nil || s.healthSweep.ctx.Err() != nil {
		s.healthSweep.mu.Unlock()
		return false, false
	}
	if s.healthSweep.running {
		s.healthSweep.mu.Unlock()
		return false, true
	}
	ctx := s.healthSweep.ctx
	s.healthSweep.running = true
	s.healthSweep.nextSweepAt = time.Time{}
	s.healthSweep.done = make(chan struct{})
	s.healthSweep.mu.Unlock()
	go s.runHealthSweep(ctx, nodes, partial)
	return true, true
}

func (s *Server) stopHealthChecks() {
	s.healthSweep.mu.Lock()
	s.healthSweep.stopping = true
	if s.healthSweep.cancel != nil {
		s.healthSweep.cancel()
	}
	done := s.healthSweep.done
	s.healthSweep.mu.Unlock()
	if done != nil {
		timer := time.NewTimer(5 * time.Second)
		select {
		case <-done:
			if !timer.Stop() {
				<-timer.C
			}
		case <-timer.C:
		}
	}
}

func (s *Server) runHealthSweep(ctx context.Context, nodes []models.Node, partial bool) {
	defer func() {
		now := time.Now()
		s.healthSweep.mu.Lock()
		s.healthSweep.running = false
		if s.healthSweep.done != nil {
			close(s.healthSweep.done)
			s.healthSweep.done = nil
		}
		s.healthSweep.lastSweepAt = now
		if s.cfg.HealthCheckEnabled && ctx.Err() == nil {
			s.healthSweep.nextSweepAt = now.Add(time.Duration(s.cfg.HealthCheckIntervalSeconds) * time.Second)
		}
		s.healthSweep.mu.Unlock()
	}()

	if !partial {
		nodes = s.allProfileCandidateNodes()
	}
	if len(nodes) == 0 {
		return
	}
	if !partial {
		s.router.health.retain(nodes)
		s.router.runtime.retain(nodes)
	}
	jobs := make(chan models.Node)
	var wg sync.WaitGroup
	workers := healthProbeConcurrency
	if len(nodes) < workers {
		workers = len(nodes)
	}
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for node := range jobs {
				if ctx.Err() != nil {
					return
				}
				s.router.health.markChecking(node)
				probeCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.HealthCheckTimeoutSeconds)*time.Second)
				latency, probeErr := s.healthProbe(probeCtx, node, time.Duration(s.cfg.HealthCheckTimeoutSeconds)*time.Second)
				cancel()
				if probeErr != nil {
					s.router.health.recordProbeFailureReason(node, time.Duration(s.cfg.FailureCooldownSeconds)*time.Second, s.sanitizeProbeError(node, probeErr))
					continue
				}
				s.router.health.recordProbeSuccess(node, latency)
			}
		}()
	}
	for _, node := range nodes {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return
		case jobs <- node:
		}
	}
	close(jobs)
	wg.Wait()
}

func (s *Server) defaultHealthProbe(ctx context.Context, node models.Node, _ time.Duration) (latency int, err error) {
	if s.pool == nil {
		return 0, errors.New("SOCKS5 health probe is unavailable for the configured dialer")
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			latency = 0
			err = errors.New("SOCKS5 health probe failed unexpectedly")
		}
	}()
	proxy, lease, err := s.pool.acquire(node)
	if err != nil {
		return 0, err
	}
	failed := true
	defer func() { lease.release(failed) }()
	delay, err := proxy.URLTest(ctx, healthProbeURL, nil)
	if err != nil {
		return 0, err
	}
	failed = false
	return int(delay), nil
}

func (s *Server) probeSchedule() HealthSummary {
	s.healthSweep.mu.Lock()
	defer s.healthSweep.mu.Unlock()
	return HealthSummary{
		Running: s.healthSweep.running, LastSweepAt: timePointer(s.healthSweep.lastSweepAt),
		NextSweepAt: timePointer(s.healthSweep.nextSweepAt),
	}
}

func (s *Server) HealthSummary() HealthSummary {
	result := s.probeSchedule()
	health := s.router.health
	health.mu.Lock()
	defer health.mu.Unlock()
	for _, state := range health.states {
		result.count(state.Status)
	}
	return result
}

func (s *Server) HealthSnapshot() HealthSnapshot {
	result := HealthSnapshot{HealthSummary: s.probeSchedule()}
	nodes := s.router.health.snapshot()
	for _, node := range nodes {
		result.count(node.Status)
	}
	if len(nodes) > healthSnapshotNodeLimit {
		result.Nodes = nodes[:healthSnapshotNodeLimit]
		result.Truncated = true
	} else {
		result.Nodes = nodes
	}
	return result
}
