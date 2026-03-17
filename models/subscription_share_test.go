package models

import (
	"strconv"
	"testing"
	"time"

	"sublink/cache"
	"sublink/database"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func resetSubscriptionShareCacheForTest() {
	subscriptionShareCache = cache.NewMapCache(func(s SubscriptionShare) int { return s.ID })
	subscriptionShareCache.AddIndex("token", func(s SubscriptionShare) string { return s.Token })
	subscriptionShareCache.AddIndex("subscriptionID", func(s SubscriptionShare) string { return strconv.Itoa(s.SubscriptionID) })
}

func setupSubscriptionShareTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&SubscriptionShare{}); err != nil {
		t.Fatalf("auto migrate subscription_shares: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	resetSubscriptionShareCacheForTest()

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		resetSubscriptionShareCacheForTest()
	})
}

func TestSubscriptionShareAddNormalizesOptionalTimestamps(t *testing.T) {
	setupSubscriptionShareTestDB(t)

	zero := time.Time{}
	share := &SubscriptionShare{
		SubscriptionID: 1,
		Name:           "never-expire",
		ExpireType:     ExpireTypeNever,
		ExpireAt:       &zero,
		LastAccessAt:   &zero,
	}

	if err := share.Add(); err != nil {
		t.Fatalf("add share: %v", err)
	}

	var stored SubscriptionShare
	if err := database.DB.First(&stored, share.ID).Error; err != nil {
		t.Fatalf("reload share: %v", err)
	}

	if stored.ExpireAt != nil {
		t.Fatalf("expected expire_at to be nil, got %v", stored.ExpireAt)
	}
	if stored.LastAccessAt != nil {
		t.Fatalf("expected last_access_at to be nil, got %v", stored.LastAccessAt)
	}
}

func TestSubscriptionShareUpdateClearsExpireAtForNonDateTime(t *testing.T) {
	setupSubscriptionShareTestDB(t)

	expireAt := time.Now().Add(24 * time.Hour).Round(time.Second)
	share := &SubscriptionShare{
		SubscriptionID: 1,
		Name:           "datetime-expire",
		ExpireType:     ExpireTypeDateTime,
		ExpireAt:       &expireAt,
	}

	if err := share.Add(); err != nil {
		t.Fatalf("add share: %v", err)
	}

	share.ExpireType = ExpireTypeNever
	share.ExpireAt = &expireAt
	if err := share.Update(); err != nil {
		t.Fatalf("update share: %v", err)
	}

	var stored SubscriptionShare
	if err := database.DB.First(&stored, share.ID).Error; err != nil {
		t.Fatalf("reload share: %v", err)
	}

	if stored.ExpireAt != nil {
		t.Fatalf("expected expire_at to be cleared, got %v", stored.ExpireAt)
	}
}

func TestSubscriptionShareRecordAccessSetsLastAccessAt(t *testing.T) {
	setupSubscriptionShareTestDB(t)

	share := &SubscriptionShare{
		SubscriptionID: 1,
		Name:           "record-access",
		ExpireType:     ExpireTypeNever,
	}

	if err := share.Add(); err != nil {
		t.Fatalf("add share: %v", err)
	}

	share.RecordAccess()

	var stored SubscriptionShare
	if err := database.DB.First(&stored, share.ID).Error; err != nil {
		t.Fatalf("reload share: %v", err)
	}

	if stored.AccessCount != 1 {
		t.Fatalf("expected access_count=1, got %d", stored.AccessCount)
	}
	if stored.LastAccessAt == nil || stored.LastAccessAt.IsZero() {
		t.Fatalf("expected last_access_at to be set, got %v", stored.LastAccessAt)
	}
}
