package socks5

import (
	"strconv"
	"testing"
	"time"

	"sublink/models"
)

func TestNodeRuntimeTracksActiveLifecycleExactlyOnce(t *testing.T) {
	table := newNodeRuntimeTable()
	now := time.Unix(100, 0)
	table.now = func() time.Time { return now }
	node := models.Node{ID: 1, Name: "node", Link: "a"}

	release := table.start(node)
	state := table.metrics(node)
	if state.ActiveConnections != 1 || state.SuccessfulConnections != 1 || !state.LastSelectedAt.Equal(now) {
		t.Fatalf("unexpected active state: %+v", state)
	}
	release()
	release()
	state = table.metrics(node)
	if state.ActiveConnections != 0 || state.SuccessfulConnections != 1 {
		t.Fatalf("release was not exact-once: %+v", state)
	}
}

func TestNodeRuntimeRecordsFailuresAndRetainsActiveEntries(t *testing.T) {
	table := newNodeRuntimeTable()
	activeNode := models.Node{ID: 1, Link: "a"}
	idleNode := models.Node{ID: 2, Link: "b"}
	release := table.start(activeNode)
	table.recordFailure(idleNode)
	table.retain(nil)
	if _, ok := table.states[adapterKey(activeNode)]; !ok {
		t.Fatal("retain removed active runtime state")
	}
	if _, ok := table.states[adapterKey(idleNode)]; ok {
		t.Fatal("retain kept idle runtime state outside candidate pool")
	}
	release()
	table.retain(nil)
	if len(table.states) != 0 {
		t.Fatalf("retain did not remove released state: %+v", table.states)
	}
}

func TestNodeRuntimeTableRemainsBoundedWhenAllEntriesAreActive(t *testing.T) {
	table := newNodeRuntimeTable()
	releases := make([]func(), 0, maxNodeRuntimeEntries)
	for id := 1; id <= maxNodeRuntimeEntries; id++ {
		releases = append(releases, table.start(models.Node{ID: id, Link: strconv.Itoa(id)}))
	}
	extraNode := models.Node{ID: maxNodeRuntimeEntries + 1, Link: "extra"}
	extraRelease := table.start(extraNode)
	if len(table.states) != maxNodeRuntimeEntries {
		t.Fatalf("runtime table exceeded limit: %d", len(table.states))
	}
	if state := table.metrics(extraNode); state.ActiveConnections != 0 {
		t.Fatalf("overflow node unexpectedly tracked: %+v", state)
	}
	extraRelease()
	for _, release := range releases {
		release()
	}
}

func TestNodeRuntimeResetPreservesActiveConnections(t *testing.T) {
	table := newNodeRuntimeTable()
	node := models.Node{ID: 1, Link: "active"}
	idle := models.Node{ID: 2, Link: "idle"}
	release := table.start(node)
	table.recordFailure(node)
	table.recordFailure(idle)
	table.reset()

	state := table.metrics(node)
	if state.ActiveConnections != 1 || state.SuccessfulConnections != 0 || state.FailedConnections != 0 || !state.LastSelectedAt.IsZero() {
		t.Fatalf("active state was not reset safely: %+v", state)
	}
	if _, ok := table.states[adapterKey(idle)]; ok {
		t.Fatal("idle runtime state was not removed by reset")
	}
	release()
}
