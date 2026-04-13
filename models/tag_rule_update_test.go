package models

import (
	"testing"
	"time"

	"sublink/cache"
	"sublink/database"
	"sublink/internal/testutil"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func resetTagCachesForTest() {
	tagCache = cache.NewMapCache(func(t Tag) string { return t.Name })
	tagRuleCache = cache.NewMapCache(func(r TagRule) int { return r.ID })
	tagRuleCache.AddIndex("tagName", func(r TagRule) string { return r.TagName })
	tagRuleCache.AddIndex("triggerType", func(r TagRule) string { return r.TriggerType })
}

func setupTagRuleTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "tag_rule_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&Tag{}, &TagRule{}); err != nil {
		t.Fatalf("auto migrate tags/tag_rules: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	resetTagCachesForTest()

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		resetTagCachesForTest()
		testutil.CloseDB(t, db)
	})
}

func TestTagRuleUpdatePreservesCreatedAt(t *testing.T) {
	setupTagRuleTestDB(t)

	tag := &Tag{Name: "优质"}
	if err := tag.Add(); err != nil {
		t.Fatalf("add tag: %v", err)
	}

	rule := &TagRule{
		TagName:     tag.Name,
		Name:        "原规则",
		Enabled:     true,
		TriggerType: "speed_test",
		Conditions:  `{"logic":"and","conditions":[{"field":"speed","operator":"greater_or_equal","value":"4"}]}`,
	}
	if err := rule.Add(); err != nil {
		t.Fatalf("add rule: %v", err)
	}

	originalCreatedAt := rule.CreatedAt
	if originalCreatedAt.IsZero() {
		t.Fatal("expected created_at to be set on insert")
	}

	time.Sleep(10 * time.Millisecond)

	rule.Name = "更新后规则"
	rule.Enabled = false
	rule.Conditions = `{"logic":"and","conditions":[{"field":"delay_status","operator":"equals","value":"success"}]}`
	if err := rule.Update(); err != nil {
		t.Fatalf("update rule: %v", err)
	}

	var stored TagRule
	if err := database.DB.First(&stored, rule.ID).Error; err != nil {
		t.Fatalf("reload rule: %v", err)
	}

	if !stored.CreatedAt.Equal(originalCreatedAt) {
		t.Fatalf("expected created_at to stay %v, got %v", originalCreatedAt, stored.CreatedAt)
	}
	if stored.Enabled {
		t.Fatal("expected enabled to persist false after update")
	}
	if !stored.UpdatedAt.After(originalCreatedAt) {
		t.Fatalf("expected updated_at to move forward, created=%v updated=%v", originalCreatedAt, stored.UpdatedAt)
	}

	cached, ok := tagRuleCache.Get(rule.ID)
	if !ok {
		t.Fatal("expected updated rule in cache")
	}
	if !cached.CreatedAt.Equal(originalCreatedAt) {
		t.Fatalf("expected cached created_at to stay %v, got %v", originalCreatedAt, cached.CreatedAt)
	}
	if cached.Enabled {
		t.Fatal("expected cached enabled to persist false")
	}
}
