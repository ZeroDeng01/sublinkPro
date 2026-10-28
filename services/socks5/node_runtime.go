package socks5

import (
	"sort"
	"sync"
	"time"

	"sublink/models"
)

const maxNodeRuntimeEntries = 10000

type nodeRuntimeState struct {
	NodeID                int
	NodeName              string
	ActiveConnections     int
	SuccessfulConnections uint64
	FailedConnections     uint64
	LastSelectedAt        time.Time
}

type nodeRuntimeTable struct {
	mu     sync.Mutex
	states map[string]*nodeRuntimeState
	now    func() time.Time
}

func newNodeRuntimeTable() *nodeRuntimeTable {
	return &nodeRuntimeTable{states: make(map[string]*nodeRuntimeState), now: time.Now}
}

func (t *nodeRuntimeTable) stateLocked(node models.Node) *nodeRuntimeState {
	key := adapterKey(node)
	state := t.states[key]
	if state == nil {
		if len(t.states) >= maxNodeRuntimeEntries && !t.evictIdleLocked() {
			return nil
		}
		state = &nodeRuntimeState{}
		t.states[key] = state
	}
	state.NodeID = node.ID
	state.NodeName = node.EffectiveName()
	return state
}

func (t *nodeRuntimeTable) start(node models.Node) func() {
	t.mu.Lock()
	state := t.stateLocked(node)
	if state == nil {
		t.mu.Unlock()
		return func() {}
	}
	state.ActiveConnections++
	state.SuccessfulConnections++
	state.LastSelectedAt = t.now()
	t.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			t.mu.Lock()
			if current := t.states[adapterKey(node)]; current != nil && current.ActiveConnections > 0 {
				current.ActiveConnections--
			}
			t.mu.Unlock()
		})
	}
}

func (t *nodeRuntimeTable) recordFailure(node models.Node) {
	t.mu.Lock()
	state := t.stateLocked(node)
	if state != nil {
		state.FailedConnections++
	}
	t.mu.Unlock()
}

func (t *nodeRuntimeTable) metrics(node models.Node) nodeRuntimeState {
	t.mu.Lock()
	defer t.mu.Unlock()
	state := t.states[adapterKey(node)]
	if state == nil {
		return nodeRuntimeState{NodeID: node.ID, NodeName: node.EffectiveName()}
	}
	return *state
}

func (t *nodeRuntimeTable) snapshot(nodes []models.Node) map[string]nodeRuntimeState {
	t.mu.Lock()
	defer t.mu.Unlock()
	result := make(map[string]nodeRuntimeState, len(nodes))
	for _, node := range nodes {
		key := adapterKey(node)
		if state := t.states[key]; state != nil {
			result[key] = *state
		}
	}
	return result
}

func (t *nodeRuntimeTable) reset() {
	t.mu.Lock()
	for key, state := range t.states {
		if state.ActiveConnections == 0 {
			delete(t.states, key)
			continue
		}
		state.SuccessfulConnections = 0
		state.FailedConnections = 0
		state.LastSelectedAt = time.Time{}
	}
	t.mu.Unlock()
}

func (t *nodeRuntimeTable) retain(nodes []models.Node) {
	allowed := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		allowed[adapterKey(node)] = struct{}{}
	}
	t.mu.Lock()
	for key, state := range t.states {
		if _, ok := allowed[key]; !ok && state.ActiveConnections == 0 {
			delete(t.states, key)
		}
	}
	t.mu.Unlock()
}

func (t *nodeRuntimeTable) evictIdleLocked() bool {
	type candidate struct {
		key      string
		selected time.Time
	}
	candidates := make([]candidate, 0, len(t.states))
	for key, state := range t.states {
		if state.ActiveConnections == 0 {
			candidates = append(candidates, candidate{key: key, selected: state.LastSelectedAt})
		}
	}
	if len(candidates) == 0 {
		return false
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].selected.Before(candidates[j].selected) })
	delete(t.states, candidates[0].key)
	return true
}
