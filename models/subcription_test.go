package models

import (
	"testing"

	"sublink/cache"
	"sublink/database"
	"sublink/internal/testutil"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

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
