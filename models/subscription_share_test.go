package models

import (
	"strconv"
	"testing"
	"time"

	"sublink/cache"
	"sublink/database"
	"sublink/internal/testutil"

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

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "subscription_share_test")), &gorm.Config{})
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
		// 等待在飞的异步访问统计写入完成，避免其在拆库/重置缓存后并发读写全局状态。
		WaitForPendingAccessRecords()
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		resetSubscriptionShareCacheForTest()
		testutil.CloseDB(t, db)
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

func TestSubscriptionShareRecordAccessAsyncUpdatesEventually(t *testing.T) {
	setupSubscriptionShareTestDB(t)

	share := &SubscriptionShare{
		SubscriptionID: 1,
		Name:           "record-access-async",
		ExpireType:     ExpireTypeNever,
	}

	if err := share.Add(); err != nil {
		t.Fatalf("add share: %v", err)
	}

	share.RecordAccessAsync()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var stored SubscriptionShare
		if err := database.DB.First(&stored, share.ID).Error; err != nil {
			t.Fatalf("reload share: %v", err)
		}
		if stored.AccessCount == 1 && stored.LastAccessAt != nil && !stored.LastAccessAt.IsZero() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	var stored SubscriptionShare
	if err := database.DB.First(&stored, share.ID).Error; err != nil {
		t.Fatalf("reload share after wait: %v", err)
	}
	t.Fatalf("expected async access to be recorded, got count=%d last_access_at=%v", stored.AccessCount, stored.LastAccessAt)
}

func TestGetSharesBySubscriptionIDPaginatedSortsNamesNaturally(t *testing.T) {
	setupSubscriptionShareTestDB(t)

	names := []string{"测试sss-1", "测试sss-11", "测试sss-2", "测试sss-10", "测试sss-3"}
	for _, name := range names {
		share := &SubscriptionShare{
			SubscriptionID: 1,
			Name:           name,
			ExpireType:     ExpireTypeNever,
		}
		if err := share.Add(); err != nil {
			t.Fatalf("add share %q: %v", name, err)
		}
	}

	shares, total, err := GetSharesBySubscriptionIDPaginated(1, 1, 10, "", "", "name", "asc")
	if err != nil {
		t.Fatalf("get shares sorted asc: %v", err)
	}
	if total != len(names) {
		t.Fatalf("expected total=%d, got %d", len(names), total)
	}
	assertShareNames(t, shares, []string{"测试sss-1", "测试sss-2", "测试sss-3", "测试sss-10", "测试sss-11"})

	shares, _, err = GetSharesBySubscriptionIDPaginated(1, 1, 10, "", "", "name", "desc")
	if err != nil {
		t.Fatalf("get shares sorted desc: %v", err)
	}
	assertShareNames(t, shares, []string{"测试sss-11", "测试sss-10", "测试sss-3", "测试sss-2", "测试sss-1"})
}

func TestGetSharesBySubscriptionIDPaginatedSortsBeforePagination(t *testing.T) {
	setupSubscriptionShareTestDB(t)

	names := []string{"测试sss-1", "测试sss-11", "测试sss-2", "测试sss-10", "测试sss-3"}
	for _, name := range names {
		share := &SubscriptionShare{
			SubscriptionID: 1,
			Name:           name,
			ExpireType:     ExpireTypeNever,
		}
		if err := share.Add(); err != nil {
			t.Fatalf("add share %q: %v", name, err)
		}
	}

	firstPage, total, err := GetSharesBySubscriptionIDPaginated(1, 1, 2, "", "", "name", "asc")
	if err != nil {
		t.Fatalf("get first sorted page: %v", err)
	}
	if total != len(names) {
		t.Fatalf("expected total=%d, got %d", len(names), total)
	}
	assertShareNames(t, firstPage, []string{"测试sss-1", "测试sss-2"})

	secondPage, _, err := GetSharesBySubscriptionIDPaginated(1, 2, 2, "", "", "name", "asc")
	if err != nil {
		t.Fatalf("get second sorted page: %v", err)
	}
	assertShareNames(t, secondPage, []string{"测试sss-3", "测试sss-10"})
}

func assertShareNames(t *testing.T, shares []SubscriptionShare, expected []string) {
	t.Helper()
	if len(shares) != len(expected) {
		t.Fatalf("expected %d shares, got %d", len(expected), len(shares))
	}
	for index, share := range shares {
		if share.Name != expected[index] {
			t.Fatalf("share[%d]: expected %q, got %q", index, expected[index], share.Name)
		}
	}
}
