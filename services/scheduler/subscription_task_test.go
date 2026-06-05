package scheduler

import (
	"testing"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupSubscriptionTaskNodeTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "subscription_task_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.Node{}); err != nil {
		t.Fatalf("auto migrate nodes: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		testutil.CloseDB(t, db)
	})
}

func TestResolveUpdateAfterDetectNodeIDsSkipsWhenChangedOnlyHasNoChanges(t *testing.T) {
	nodeIDs, shouldRun, err := resolveUpdateAfterDetectNodeIDs(1, nil, true)
	if err != nil {
		t.Fatalf("resolve node IDs: %v", err)
	}
	if shouldRun {
		t.Fatalf("expected changed-only update with no changed nodes to skip, got node IDs %v", nodeIDs)
	}
	if len(nodeIDs) != 0 {
		t.Fatalf("expected no node IDs for skipped changed-only update, got %v", nodeIDs)
	}
}

func TestResolveUpdateAfterDetectNodeIDsUsesChangedNodesWhenChangedOnlyHasChanges(t *testing.T) {
	changedNodeIDs := []int{11, 12}
	nodeIDs, shouldRun, err := resolveUpdateAfterDetectNodeIDs(1, changedNodeIDs, true)
	if err != nil {
		t.Fatalf("resolve node IDs: %v", err)
	}
	if !shouldRun {
		t.Fatal("expected changed-only update with changed nodes to run")
	}
	if len(nodeIDs) != len(changedNodeIDs) {
		t.Fatalf("expected %d node IDs, got %d", len(changedNodeIDs), len(nodeIDs))
	}
	for i, want := range changedNodeIDs {
		if nodeIDs[i] != want {
			t.Fatalf("node ID %d = %d, want %d", i, nodeIDs[i], want)
		}
	}
}

func TestResolveUpdateAfterDetectNodeIDsUsesAirportNodesWhenChangedOnlyDisabled(t *testing.T) {
	setupSubscriptionTaskNodeTestDB(t)

	nodes := []models.Node{
		{Name: "节点A", Link: "ss://node-a", LinkHash: "hash-node-a", SourceID: 3},
		{Name: "节点B", Link: "ss://node-b", LinkHash: "hash-node-b", SourceID: 3},
		{Name: "其它机场节点", Link: "ss://other", LinkHash: "hash-other", SourceID: 4},
	}
	for i := range nodes {
		if err := database.DB.Create(&nodes[i]).Error; err != nil {
			t.Fatalf("create node %q: %v", nodes[i].Name, err)
		}
	}

	nodeIDs, shouldRun, err := resolveUpdateAfterDetectNodeIDs(3, nil, false)
	if err != nil {
		t.Fatalf("resolve node IDs: %v", err)
	}
	if !shouldRun {
		t.Fatal("expected full-airport post-update check to run when airport has nodes")
	}

	wantIDs := []int{nodes[0].ID, nodes[1].ID}
	if len(nodeIDs) != len(wantIDs) {
		t.Fatalf("expected %d airport node IDs, got %d: %v", len(wantIDs), len(nodeIDs), nodeIDs)
	}
	for i, want := range wantIDs {
		if nodeIDs[i] != want {
			t.Fatalf("node ID %d = %d, want %d", i, nodeIDs[i], want)
		}
	}
}
