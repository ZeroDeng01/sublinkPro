package socks5

import (
	"errors"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"sublink/models"
)

const candidateUngroupedValue = "__ungrouped__"

type nodeHealthState struct {
	NodeID              int
	NodeName            string
	Status              string
	LatencyMs           int
	LastCheckedAt       time.Time
	LastSuccessAt       time.Time
	LastFailureAt       time.Time
	ConsecutiveFailures int
	CooldownUntil       time.Time
	LastError           string
	ExcludeFromRouting  bool
}

type nodeHealth struct {
	mu            sync.Mutex
	cooldownUntil map[string]time.Time
	states        map[string]*nodeHealthState
	now           func() time.Time
}

func newNodeHealth() *nodeHealth {
	return &nodeHealth{cooldownUntil: make(map[string]time.Time), states: make(map[string]*nodeHealthState), now: time.Now}
}

func (h *nodeHealth) stateLocked(node models.Node) *nodeHealthState {
	key := adapterKey(node)
	state := h.states[key]
	if state == nil {
		state = &nodeHealthState{NodeID: node.ID, NodeName: node.EffectiveName(), Status: "unknown"}
		h.states[key] = state
	}
	state.NodeID = node.ID
	state.NodeName = node.EffectiveName()
	return state
}

func (h *nodeHealth) markChecking(node models.Node) {
	h.mu.Lock()
	state := h.stateLocked(node)
	state.Status = "checking"
	h.mu.Unlock()
}

func (h *nodeHealth) retain(nodes []models.Node) {
	allowed := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		allowed[adapterKey(node)] = struct{}{}
	}
	h.mu.Lock()
	for key := range h.states {
		if _, ok := allowed[key]; !ok {
			delete(h.states, key)
			delete(h.cooldownUntil, key)
		}
	}
	h.mu.Unlock()
}

func (h *nodeHealth) recordFailure(node models.Node, cooldown time.Duration) {
	h.recordFailureReason(node, cooldown, "")
}

func (h *nodeHealth) recordFailureReason(node models.Node, cooldown time.Duration, reason string) {
	h.recordFailureState(node, cooldown, reason, false)
}

func (h *nodeHealth) recordProbeFailureReason(node models.Node, cooldown time.Duration, reason string) {
	h.recordFailureState(node, cooldown, reason, true)
}

func (h *nodeHealth) recordFailureState(node models.Node, cooldown time.Duration, reason string, excludeFromRouting bool) {
	h.mu.Lock()
	now := h.now()
	state := h.stateLocked(node)
	state.Status = "unhealthy"
	state.LastCheckedAt = now
	state.LastFailureAt = now
	state.ConsecutiveFailures++
	state.LastError = sanitizeHealthError(reason)
	if excludeFromRouting {
		state.ExcludeFromRouting = true
	}
	if cooldown > 0 {
		backoff := cooldown
		for i := 1; i < state.ConsecutiveFailures && backoff < time.Hour; i++ {
			backoff *= 2
		}
		if backoff > time.Hour {
			backoff = time.Hour
		}
		state.CooldownUntil = now.Add(backoff)
		h.cooldownUntil[adapterKey(node)] = state.CooldownUntil
	}
	h.mu.Unlock()
}

func (h *nodeHealth) recordSuccess(node models.Node) {
	h.mu.Lock()
	now := h.now()
	state := h.stateLocked(node)
	state.Status = "healthy"
	state.LastCheckedAt = now
	state.LastSuccessAt = now
	state.ConsecutiveFailures = 0
	state.CooldownUntil = time.Time{}
	state.LastError = ""
	state.ExcludeFromRouting = false
	delete(h.cooldownUntil, adapterKey(node))
	h.mu.Unlock()
}

func (h *nodeHealth) recordProbeSuccess(node models.Node, latency int) {
	h.mu.Lock()
	now := h.now()
	state := h.stateLocked(node)
	state.Status = "healthy"
	state.LatencyMs = latency
	state.LastCheckedAt = now
	state.LastSuccessAt = now
	state.ConsecutiveFailures = 0
	state.CooldownUntil = time.Time{}
	state.LastError = ""
	state.ExcludeFromRouting = false
	delete(h.cooldownUntil, adapterKey(node))
	h.mu.Unlock()
}

func (h *nodeHealth) routingMetrics(node models.Node) (int, int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	state := h.states[adapterKey(node)]
	if state == nil {
		return 0, 0
	}
	return state.LatencyMs, state.ConsecutiveFailures
}

type rankedNode struct {
	node       models.Node
	status     string
	latencyMs  int
	excluded   bool
	originalAt int
}

func (h *nodeHealth) filter(nodes []models.Node) []models.Node {
	return h.rank(nodes, false)
}

func (h *nodeHealth) rank(nodes []models.Node, preferLatency bool) []models.Node {
	if len(nodes) == 0 {
		return nil
	}
	now := h.now()
	h.mu.Lock()
	available := make([]rankedNode, 0, len(nodes))
	var probe models.Node
	var probeAt time.Time
	for index, node := range nodes {
		key := adapterKey(node)
		until, cooling := h.cooldownUntil[key]
		if cooling && until.After(now) {
			if probeAt.IsZero() || until.Before(probeAt) {
				probe, probeAt = node, until
			}
			continue
		}
		delete(h.cooldownUntil, key)
		status := "unknown"
		latency := 0
		if state := h.states[key]; state != nil {
			state.CooldownUntil = time.Time{}
			status = state.Status
			latency = state.LatencyMs
		}
		excluded := false
		if state := h.states[key]; state != nil {
			excluded = state.ExcludeFromRouting
		}
		available = append(available, rankedNode{node: node, status: status, latencyMs: latency, excluded: excluded, originalAt: index})
	}
	h.mu.Unlock()
	if len(available) == 0 {
		// Avoid a permanently dead pool: when every node is cooling down, probe
		// the node whose cooldown expires first.
		return []models.Node{probe}
	}

	hasRoutable := false
	for _, candidate := range available {
		if !candidate.excluded {
			hasRoutable = true
			break
		}
	}
	if hasRoutable {
		filtered := available[:0]
		for _, candidate := range available {
			if !candidate.excluded {
				filtered = append(filtered, candidate)
			}
		}
		available = filtered
	}

	if preferLatency {
		sort.SliceStable(available, func(i, j int) bool {
			leftRank := healthCandidateRank(available[i])
			rightRank := healthCandidateRank(available[j])
			if leftRank != rightRank {
				return leftRank < rightRank
			}
			if leftRank == 0 && available[i].latencyMs != available[j].latencyMs {
				return available[i].latencyMs < available[j].latencyMs
			}
			return available[i].originalAt < available[j].originalAt
		})
	}

	result := make([]models.Node, 0, len(available))
	for _, candidate := range available {
		result = append(result, candidate.node)
	}
	return result
}

func healthCandidateRank(candidate rankedNode) int {
	switch candidate.status {
	case "healthy":
		if candidate.latencyMs > 0 {
			return 0
		}
		return 1
	case "checking":
		if candidate.latencyMs > 0 {
			return 0
		}
		return 2
	case "unknown", "":
		return 2
	default:
		return 3
	}
}

type nodeRouter struct {
	health       *nodeHealth
	sticky       *stickySessionTable
	runtime      *nodeRuntimeTable
	randomIndex  func(int) int
	roundRobinMu sync.Mutex
	roundRobin   map[string]uint64
}

func newNodeRouter() *nodeRouter {
	return &nodeRouter{health: newNodeHealth(), sticky: newStickySessionTable(), runtime: newNodeRuntimeTable(), randomIndex: rand.Intn, roundRobin: make(map[string]uint64)}
}

var listCandidateNodesFunc = listCandidateNodes

func (r *nodeRouter) candidates(cfg Config) ([]models.Node, error) {
	return r.candidatesFor(cfg, "")
}

func (r *nodeRouter) candidatesFor(cfg Config, stickyKey string) ([]models.Node, error) {
	return r.candidatesForScope(cfg, stickyKey, defaultProfileID)
}

func (r *nodeRouter) candidatesForScope(cfg Config, stickyKey, scope string) ([]models.Node, error) {
	nodes, err := listCandidateNodesFunc(cfg)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, errors.New("no proxy nodes are available")
	}

	if cfg.Selection == "best" {
		nodes = r.health.rank(nodes, true)
	} else {
		nodes = r.health.filter(nodes)
	}

	stickyHit := false
	if cfg.StickySessionEnabled && stickyKey != "" && cfg.Selection != "specific" {
		nodes, stickyHit = r.sticky.promote(stickyKey, nodes)
	}

	switch cfg.Selection {
	case "smart":
		nodes = r.smartOrder(nodes, stickyHit, cfg.MaxAttempts)
	case "random":
		start := 0
		if stickyHit {
			start = 1
		}
		rand.Shuffle(len(nodes)-start, func(i, j int) {
			i += start
			j += start
			nodes[i], nodes[j] = nodes[j], nodes[i]
		})
	case "round_robin":
		if !stickyHit {
			start := r.nextRoundRobin(scope, len(nodes))
			nodes = append(append(make([]models.Node, 0, len(nodes)), nodes[start:]...), nodes[:start]...)
		}
	}
	if cfg.MaxAttempts < len(nodes) {
		nodes = nodes[:cfg.MaxAttempts]
	}
	return nodes, nil
}

func (r *nodeRouter) nextRoundRobin(scope string, count int) int {
	if count <= 0 {
		return 0
	}
	scope = strings.ToLower(strings.TrimSpace(scope))
	if scope == "" {
		scope = defaultProfileID
	}
	r.roundRobinMu.Lock()
	value := r.roundRobin[scope]
	r.roundRobin[scope] = value + 1
	r.roundRobinMu.Unlock()
	return int(value % uint64(count))
}

func (r *nodeRouter) smartOrder(nodes []models.Node, stickyHit bool, maxAttempts int) []models.Node {
	if len(nodes) < 2 {
		return nodes
	}
	if !stickyHit {
		count := len(nodes)
		left := r.randomIndex(count)
		right := r.randomIndex(count - 1)
		if right >= left {
			right++
		}
		if r.smartScore(nodes[right]) < r.smartScore(nodes[left]) {
			left = right
		}
		nodes[0], nodes[left] = nodes[left], nodes[0]
	}
	const fallbackStart = 1
	limit := maxAttempts
	if limit <= 0 || limit > len(nodes) {
		limit = len(nodes)
	}
	for position := fallbackStart; position < limit; position++ {
		best := position
		bestScore := r.smartScore(nodes[best])
		for candidate := position + 1; candidate < len(nodes); candidate++ {
			candidateScore := r.smartScore(nodes[candidate])
			if candidateScore < bestScore || candidateScore == bestScore && nodes[candidate].ID < nodes[best].ID {
				best = candidate
				bestScore = candidateScore
			}
		}
		nodes[position], nodes[best] = nodes[best], nodes[position]
	}
	return nodes
}

func (r *nodeRouter) smartScore(node models.Node) float64 {
	latency, failures := r.health.routingMetrics(node)
	if latency <= 0 {
		latency = node.DelayTime
	}
	if latency <= 0 {
		latency = 500
	}
	runtime := r.runtime.metrics(node)
	return float64(latency)*(1+float64(runtime.ActiveConnections)*0.5) + float64(failures)*100
}

func listCandidateNodes(cfg Config) ([]models.Node, error) {
	var first *models.Node
	if cfg.Selection == "specific" {
		specific, ok := models.GetNodeByID(cfg.NodeID)
		if !ok || strings.TrimSpace(specific.Link) == "" {
			return nil, errors.New("configured SOCKS5 node was not found")
		}
		if !cfg.SpecificFallback {
			return []models.Node{*specific}, nil
		}
		first = specific
	}

	var modelNode models.Node
	all, err := modelNode.ListWithFilters(models.NodeFilter{})
	if err != nil {
		return nil, err
	}
	all = filterCandidatePool(all, cfg)
	if cfg.Selection == "best" {
		first = bestCandidateNode(all)
	}

	available := make([]models.Node, 0, len(all)+1)
	if first != nil {
		available = append(available, *first)
	}
	for _, node := range all {
		if strings.TrimSpace(node.Link) != "" && (first == nil || node.ID != first.ID) {
			available = append(available, node)
		}
	}
	if first == nil {
		// Keep database results deterministic for round-robin and random seeds.
		sort.SliceStable(available, func(i, j int) bool { return available[i].ID < available[j].ID })
	}
	return available, nil
}

func filterCandidatePool(nodes []models.Node, cfg Config) []models.Node {
	filtered := make([]models.Node, 0, len(nodes))
	for _, node := range nodes {
		if strings.TrimSpace(node.Link) == "" {
			continue
		}
		if !matchesCandidateGroup(node.Group, cfg.CandidateGroups) ||
			!matchesCandidateSource(node.Source, cfg.CandidateSources) ||
			!matchesCandidateValue(node.Protocol, cfg.CandidateProtocols) ||
			!matchesCandidateValue(node.LinkCountry, cfg.CandidateCountries) {
			continue
		}
		filtered = append(filtered, node)
	}
	return filtered
}

func matchesCandidateGroup(group string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, candidate := range allowed {
		if (strings.EqualFold(candidate, candidateUngroupedValue) || candidate == "未分组") && strings.TrimSpace(group) == "" {
			return true
		}
		if strings.EqualFold(group, candidate) {
			return true
		}
	}
	return false
}

func matchesCandidateSource(source string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, candidate := range allowed {
		if (candidate == "手动添加" || strings.EqualFold(candidate, "manual")) && (strings.TrimSpace(source) == "" || strings.EqualFold(source, "manual")) {
			return true
		}
		if strings.EqualFold(source, candidate) {
			return true
		}
	}
	return false
}

func matchesCandidateValue(value string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, candidate := range allowed {
		if strings.EqualFold(value, candidate) {
			return true
		}
	}
	return false
}

func bestCandidateNode(nodes []models.Node) *models.Node {
	var best *models.Node
	for index := range nodes {
		node := &nodes[index]
		if node.DelayTime <= 0 || node.Speed <= 0 {
			continue
		}
		if best == nil || node.DelayTime < best.DelayTime {
			best = node
		}
	}
	return best
}
