package models

import (
	"testing"

	"sublink/cache"
	"sublink/database"
	"sublink/internal/testutil"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSubcriptionApplyFiltersDistinguishesHTTPAndHTTPSProtocols(t *testing.T) {
	nodes := []Node{
		{Name: "plain-http", Link: "http://user:pass@example.com:8080#plain-http", Protocol: "http"},
		{Name: "secure-https", Link: "https://user:pass@example.com:8443#secure-https", Protocol: "https"},
	}

	tests := []struct {
		name string
		sub  Subcription
		want []string
	}{
		{
			name: "http whitelist keeps only HTTP nodes",
			sub:  Subcription{ProtocolWhitelist: "http", Nodes: append([]Node(nil), nodes...)},
			want: []string{"plain-http"},
		},
		{
			name: "https whitelist keeps only HTTPS nodes",
			sub:  Subcription{ProtocolWhitelist: "https", Nodes: append([]Node(nil), nodes...)},
			want: []string{"secure-https"},
		},
		{
			name: "http blacklist removes only HTTP nodes",
			sub:  Subcription{ProtocolBlacklist: "http", Nodes: append([]Node(nil), nodes...)},
			want: []string{"secure-https"},
		},
		{
			name: "https blacklist removes only HTTPS nodes",
			sub:  Subcription{ProtocolBlacklist: "https", Nodes: append([]Node(nil), nodes...)},
			want: []string{"plain-http"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sub.ApplyFilters(tt.sub.Nodes)
			if len(got) != len(tt.want) {
				t.Fatalf("filtered nodes = %v, want names %v", nodeNames(got), tt.want)
			}
			for i, node := range got {
				if node.Name != tt.want[i] {
					t.Fatalf("filtered nodes = %v, want names %v", nodeNames(got), tt.want)
				}
			}
		})
	}
}

func nodeNames(nodes []Node) []string {
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.Name)
	}
	return names
}

func resetSubcriptionCacheForTest() {
	subcriptionCache = cache.NewMapCache(func(s Subcription) int { return s.ID })
	subcriptionCache.AddIndex("name", func(s Subcription) string { return s.Name })
}

func setupSubcriptionCopyTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized
	oldSubcriptionCache := subcriptionCache
	oldSubscriptionShareCache := subscriptionShareCache

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "subcription_copy_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&Subcription{},
		&SubcriptionNode{},
		&SubcriptionGroup{},
		&SubcriptionScript{},
		&SubscriptionShare{},
		&SubscriptionChainRule{},
	); err != nil {
		t.Fatalf("auto migrate subscription copy tables: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = true
	resetSubcriptionCacheForTest()
	resetSubscriptionShareCacheForTest()

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		subcriptionCache = oldSubcriptionCache
		subscriptionShareCache = oldSubscriptionShareCache
		testutil.CloseDB(t, db)
	})
}

func TestSubcriptionCopyPreservesUpdateInterval(t *testing.T) {
	setupSubcriptionCopyTestDB(t)

	sub := &Subcription{
		Name:                  "interval-sub",
		Config:                `{"clash":"./template/clash.yaml","surge":"./template/surge.conf"}`,
		RefreshUsageOnRequest: true,
		UpdateInterval:        6,
	}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}

	copySub, err := sub.Copy()
	if err != nil {
		t.Fatalf("copy subscription: %v", err)
	}

	if copySub.UpdateInterval != sub.UpdateInterval {
		t.Fatalf("expected copied UpdateInterval %d, got %d", sub.UpdateInterval, copySub.UpdateInterval)
	}
}
