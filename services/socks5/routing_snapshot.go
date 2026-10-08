package socks5

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"sublink/models"
)

const (
	defaultRoutingSnapshotPageSize = 25
	maxRoutingSnapshotPageSize     = 100
)

type RoutingSnapshotQuery struct {
	ProfileID string
	Keyword   string
	Status    string
	SortBy    string
	SortOrder string
	Page      int
	PageSize  int
}

type NodeRoutingSnapshot struct {
	ProfileID             string     `json:"profileId"`
	ProfileName           string     `json:"profileName"`
	NodeID                int        `json:"nodeId"`
	NodeName              string     `json:"nodeName"`
	Group                 string     `json:"group,omitempty"`
	Source                string     `json:"source,omitempty"`
	Protocol              string     `json:"protocol,omitempty"`
	Country               string     `json:"country,omitempty"`
	Status                string     `json:"status"`
	LatencyMs             int        `json:"latencyMs"`
	LatencySource         string     `json:"latencySource"`
	ActiveConnections     int        `json:"activeConnections"`
	SuccessfulConnections uint64     `json:"successfulConnections"`
	FailedConnections     uint64     `json:"failedConnections"`
	ConsecutiveFailures   int        `json:"consecutiveFailures"`
	SmartScore            float64    `json:"smartScore"`
	LastSelectedAt        *time.Time `json:"lastSelectedAt,omitempty"`
	CooldownUntil         *time.Time `json:"cooldownUntil,omitempty"`
	LastError             string     `json:"lastError,omitempty"`
	ExcludedFromRouting   bool       `json:"excludedFromRouting"`
}

type RoutingSnapshotSummary struct {
	TotalNodes            int    `json:"totalNodes"`
	HealthyNodes          int    `json:"healthyNodes"`
	UnhealthyNodes        int    `json:"unhealthyNodes"`
	CheckingNodes         int    `json:"checkingNodes"`
	UnknownNodes          int    `json:"unknownNodes"`
	CoolingNodes          int    `json:"coolingNodes"`
	ActiveNodes           int    `json:"activeNodes"`
	ActiveConnections     int    `json:"activeConnections"`
	SuccessfulConnections uint64 `json:"successfulConnections"`
	FailedConnections     uint64 `json:"failedConnections"`
}

type RoutingSnapshotPage struct {
	ProfileID   string                 `json:"profileId"`
	ProfileName string                 `json:"profileName"`
	Items       []NodeRoutingSnapshot  `json:"items"`
	Total       int                    `json:"total"`
	Page        int                    `json:"page"`
	PageSize    int                    `json:"pageSize"`
	TotalPages  int                    `json:"totalPages"`
	Summary     RoutingSnapshotSummary `json:"summary"`
}

func emptyRoutingSnapshot(query RoutingSnapshotQuery) RoutingSnapshotPage {
	query = normalizeRoutingSnapshotQuery(query)
	return RoutingSnapshotPage{Items: []NodeRoutingSnapshot{}, Page: query.Page, PageSize: query.PageSize}
}

func normalizeRoutingSnapshotQuery(query RoutingSnapshotQuery) RoutingSnapshotQuery {
	query.ProfileID = strings.ToLower(strings.TrimSpace(query.ProfileID))
	if query.ProfileID == "" {
		query.ProfileID = defaultProfileID
	}
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.Status = strings.ToLower(strings.TrimSpace(query.Status))
	query.SortBy = strings.TrimSpace(query.SortBy)
	query.SortOrder = strings.ToLower(strings.TrimSpace(query.SortOrder))
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = defaultRoutingSnapshotPageSize
	} else if query.PageSize > maxRoutingSnapshotPageSize {
		query.PageSize = maxRoutingSnapshotPageSize
	}
	if query.SortBy == "" {
		query.SortBy = "smartScore"
	}
	if query.SortOrder != "desc" {
		query.SortOrder = "asc"
	}
	return query
}

func (s *Server) RoutingSnapshot(query RoutingSnapshotQuery) (RoutingSnapshotPage, error) {
	query = normalizeRoutingSnapshotQuery(query)
	profile, nodes, err := s.candidateNodesForProfile(query.ProfileID)
	if err != nil {
		return emptyRoutingSnapshot(query), err
	}
	healthStates := s.routingHealthStates(nodes)
	runtimeStates := s.router.runtime.snapshot(nodes)
	now := s.router.health.now()
	items := make([]NodeRoutingSnapshot, 0, len(nodes))
	summary := RoutingSnapshotSummary{TotalNodes: len(nodes)}
	for _, node := range nodes {
		key := adapterKey(node)
		health := healthStates[key]
		runtime := runtimeStates[key]
		item := buildNodeRoutingSnapshot(profile, node, health, runtime, now)
		items = append(items, item)
		accumulateRoutingSummary(&summary, item)
	}

	items = filterRoutingSnapshots(items, query)
	sortRoutingSnapshots(items, query.SortBy, query.SortOrder)
	total := len(items)
	totalPages := 0
	if total > 0 {
		totalPages = (total + query.PageSize - 1) / query.PageSize
		if query.Page > totalPages {
			query.Page = totalPages
		}
	}
	start := (query.Page - 1) * query.PageSize
	if start > total {
		start = total
	}
	end := start + query.PageSize
	if end > total {
		end = total
	}
	pageItems := append([]NodeRoutingSnapshot(nil), items[start:end]...)
	return RoutingSnapshotPage{
		ProfileID: profile.ID, ProfileName: profile.Name,
		Items: pageItems, Total: total, Page: query.Page, PageSize: query.PageSize,
		TotalPages: totalPages, Summary: summary,
	}, nil
}

func (s *Server) routingHealthStates(nodes []models.Node) map[string]nodeHealthState {
	states := make(map[string]nodeHealthState, len(nodes))
	s.router.health.mu.Lock()
	for _, node := range nodes {
		key := adapterKey(node)
		if state := s.router.health.states[key]; state != nil {
			states[key] = *state
		}
	}
	s.router.health.mu.Unlock()
	return states
}

func buildNodeRoutingSnapshot(profile RoutingProfile, node models.Node, health nodeHealthState, runtime nodeRuntimeState, now time.Time) NodeRoutingSnapshot {
	status := health.Status
	if status == "" {
		status = "unknown"
	}
	latency := health.LatencyMs
	latencySource := "probe"
	if latency <= 0 {
		latency = node.DelayTime
		latencySource = "stored"
	}
	if latency <= 0 {
		latency = 500
		latencySource = "fallback"
	}
	cooldownUntil := timePointer(health.CooldownUntil)
	if cooldownUntil != nil && !cooldownUntil.After(now) {
		cooldownUntil = nil
	}
	if cooldownUntil != nil {
		status = "cooling"
	}
	return NodeRoutingSnapshot{
		ProfileID: profile.ID, ProfileName: profile.Name,
		NodeID: node.ID, NodeName: node.EffectiveName(), Group: node.Group, Source: node.Source,
		Protocol: node.Protocol, Country: node.LinkCountry, Status: status, LatencyMs: latency,
		LatencySource: latencySource, ActiveConnections: runtime.ActiveConnections,
		SuccessfulConnections: runtime.SuccessfulConnections, FailedConnections: runtime.FailedConnections,
		ConsecutiveFailures: health.ConsecutiveFailures,
		SmartScore:          float64(latency)*(1+float64(runtime.ActiveConnections)*0.5) + float64(health.ConsecutiveFailures)*100,
		LastSelectedAt:      timePointer(runtime.LastSelectedAt), CooldownUntil: cooldownUntil,
		LastError: health.LastError, ExcludedFromRouting: health.ExcludeFromRouting,
	}
}

func accumulateRoutingSummary(summary *RoutingSnapshotSummary, item NodeRoutingSnapshot) {
	switch item.Status {
	case "healthy":
		summary.HealthyNodes++
	case "unhealthy":
		summary.UnhealthyNodes++
	case "checking":
		summary.CheckingNodes++
	case "cooling":
		summary.CoolingNodes++
	default:
		summary.UnknownNodes++
	}
	if item.ActiveConnections > 0 {
		summary.ActiveNodes++
	}
	summary.ActiveConnections += item.ActiveConnections
	summary.SuccessfulConnections += item.SuccessfulConnections
	summary.FailedConnections += item.FailedConnections
}

func filterRoutingSnapshots(items []NodeRoutingSnapshot, query RoutingSnapshotQuery) []NodeRoutingSnapshot {
	keyword := strings.ToLower(query.Keyword)
	filtered := make([]NodeRoutingSnapshot, 0, len(items))
	for _, item := range items {
		if query.Status != "" && query.Status != "all" && item.Status != query.Status {
			continue
		}
		if keyword != "" {
			haystack := strings.ToLower(strings.Join([]string{
				item.ProfileID, item.ProfileName, item.NodeName, fmt.Sprintf("%d", item.NodeID), item.Group, item.Source, item.Protocol, item.Country,
			}, " "))
			if !strings.Contains(haystack, keyword) {
				continue
			}
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func sortRoutingSnapshots(items []NodeRoutingSnapshot, sortBy, sortOrder string) {
	descending := sortOrder == "desc"
	sort.SliceStable(items, func(i, j int) bool {
		comparison := compareRoutingSnapshots(items[i], items[j], sortBy)
		if comparison == 0 {
			comparison = strings.Compare(strings.ToLower(items[i].NodeName), strings.ToLower(items[j].NodeName))
		}
		if comparison == 0 {
			comparison = compareInt(items[i].NodeID, items[j].NodeID)
		}
		if descending {
			return comparison > 0
		}
		return comparison < 0
	})
}

func compareRoutingSnapshots(left, right NodeRoutingSnapshot, sortBy string) int {
	switch sortBy {
	case "nodeName":
		return strings.Compare(strings.ToLower(left.NodeName), strings.ToLower(right.NodeName))
	case "status":
		return strings.Compare(left.Status, right.Status)
	case "latencyMs":
		return compareInt(left.LatencyMs, right.LatencyMs)
	case "activeConnections":
		return compareInt(left.ActiveConnections, right.ActiveConnections)
	case "successfulConnections":
		return compareUint64(left.SuccessfulConnections, right.SuccessfulConnections)
	case "failedConnections":
		return compareUint64(left.FailedConnections, right.FailedConnections)
	case "consecutiveFailures":
		return compareInt(left.ConsecutiveFailures, right.ConsecutiveFailures)
	case "lastSelectedAt":
		return compareTimes(left.LastSelectedAt, right.LastSelectedAt)
	default:
		if left.SmartScore < right.SmartScore {
			return -1
		}
		if left.SmartScore > right.SmartScore {
			return 1
		}
		return 0
	}
}

func compareInt(left, right int) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func compareUint64(left, right uint64) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func compareTimes(left, right *time.Time) int {
	if left == nil && right == nil {
		return 0
	}
	if left == nil {
		return -1
	}
	if right == nil {
		return 1
	}
	if left.Before(*right) {
		return -1
	}
	if left.After(*right) {
		return 1
	}
	return 0
}
