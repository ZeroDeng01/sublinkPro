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

type HealthSnapshot struct {
	Running        bool                 `json:"running"`
	LastSweepAt    *time.Time           `json:"lastSweepAt,omitempty"`
	NextSweepAt    *time.Time           `json:"nextSweepAt,omitempty"`
	HealthyNodes   int                  `json:"healthyNodes"`
	UnhealthyNodes int                  `json:"unhealthyNodes"`
	CheckingNodes  int                  `json:"checkingNodes"`
	UnknownNodes   int                  `json:"unknownNodes"`
	TotalNodes     int                  `json:"totalNodes"`
	Truncated      bool                 `json:"truncated"`
	Nodes          []NodeHealthSnapshot `json:"nodes"`
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
		cooldownUntil := state.CooldownUntil
		if !cooldownUntil.IsZero() && !cooldownUntil.After(now) {
			cooldownUntil = time.Time{}
		}
		result = append(result, NodeHealthSnapshot{
			NodeID: state.NodeID, NodeName: state.NodeName, Status: state.Status, LatencyMs: state.LatencyMs,
			LastCheckedAt: timePointer(state.LastCheckedAt), LastSuccessAt: timePointer(state.LastSuccessAt),
			LastFailureAt: timePointer(state.LastFailureAt), ConsecutiveFailures: state.ConsecutiveFailures,
			CooldownUntil: timePointer(cooldownUntil), LastError: state.LastError,
		})
	}
	h.mu.Unlock()
	sort.Slice(result, func(i, j int) bool {
		if result[i].NodeName == result[j].NodeName {
			return result[i].NodeID < result[j].NodeID
		}
		return result[i].NodeName < result[j].NodeName
	})
	return result
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
	go s.runHealthSweep(ctx)
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

func (s *Server) runHealthSweep(ctx context.Context) {
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

	nodes := s.allProfileCandidateNodes()
	if len(nodes) == 0 {
		return
	}
	s.router.health.retain(nodes)
	s.router.runtime.retain(nodes)
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

func (s *Server) HealthSnapshot() HealthSnapshot {
	s.healthSweep.mu.Lock()
	running := s.healthSweep.running
	lastSweepAt := timePointer(s.healthSweep.lastSweepAt)
	nextSweepAt := timePointer(s.healthSweep.nextSweepAt)
	s.healthSweep.mu.Unlock()
	nodes := s.router.health.snapshot()
	result := HealthSnapshot{Running: running, LastSweepAt: lastSweepAt, NextSweepAt: nextSweepAt, TotalNodes: len(nodes)}
	for _, node := range nodes {
		switch node.Status {
		case "healthy":
			result.HealthyNodes++
		case "unhealthy":
			result.UnhealthyNodes++
		case "checking":
			result.CheckingNodes++
		default:
			result.UnknownNodes++
		}
	}
	if len(nodes) > healthSnapshotNodeLimit {
		result.Nodes = nodes[:healthSnapshotNodeLimit]
		result.Truncated = true
	} else {
		result.Nodes = nodes
	}
	return result
}
