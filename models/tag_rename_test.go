package models

import (
	"strings"
	"testing"

	"sublink/cache"
	"sublink/database"
	"sublink/internal/testutil"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTagRenameTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized
	oldNodeCache := nodeCache
	oldSubcriptionCache := subcriptionCache
	oldProfileCache := nodeCheckProfileCache

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "tag_rename_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&Tag{},
		&TagRule{},
		&Node{},
		&Subcription{},
		&NodeCheckProfile{},
		&NodeCheckJob{},
	); err != nil {
		t.Fatalf("auto migrate tag rename tables: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = true
	resetTagCachesForTest()
	resetNodeCacheForTest()
	resetSubcriptionCacheForTest()
	nodeCheckProfileCache = cache.NewMapCache(func(p NodeCheckProfile) int { return p.ID })
	nodeCheckProfileCache.AddIndex("name", func(p NodeCheckProfile) string { return p.Name })

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		nodeCache = oldNodeCache
		subcriptionCache = oldSubcriptionCache
		nodeCheckProfileCache = oldProfileCache
		resetTagCachesForTest()
		testutil.CloseDB(t, db)
	})
}

func TestTagRenameCascadesToReferences(t *testing.T) {
	setupTagRenameTestDB(t)

	oldTag := &Tag{Name: "youtube", GroupName: "ads", Color: "#ff0000", Description: "clean"}
	if err := oldTag.Add(); err != nil {
		t.Fatalf("add tag: %v", err)
	}
	// 前缀相同的标签不能被一起改掉
	if err := (&Tag{Name: "youtube-ads", GroupName: "ads"}).Add(); err != nil {
		t.Fatalf("add sibling tag: %v", err)
	}

	rule := &TagRule{TagName: "youtube", Name: "rule", Enabled: true, TriggerType: "speed_test"}
	if err := rule.Add(); err != nil {
		t.Fatalf("add rule: %v", err)
	}

	node := Node{Link: "ss://rename-a", LinkHash: "hash-a", Name: "node-a", Tags: "youtube,premium"}
	if err := database.DB.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}
	nodeCache.Set(node.ID, node)

	sibling := Node{Link: "ss://rename-b", LinkHash: "hash-b", Name: "node-b", Tags: "youtube-ads"}
	if err := database.DB.Create(&sibling).Error; err != nil {
		t.Fatalf("create sibling node: %v", err)
	}
	nodeCache.Set(sibling.ID, sibling)

	sub := &Subcription{Name: "rename-sub", TagWhitelist: "youtube, premium", TagBlacklist: "youtube-ads"}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}

	profile := &NodeCheckProfile{Name: "rename-profile", Tags: "youtube,other"}
	if err := profile.Add(); err != nil {
		t.Fatalf("add profile: %v", err)
	}

	job := &NodeCheckJob{
		Name:      "rename-job",
		ProfileID: profile.ID,
		Trigger:   "scheduled",
		Scope:     "filtered",
		Tags:      []string{"youtube", "other"},
	}
	if err := database.DB.Create(job).Error; err != nil {
		t.Fatalf("create job: %v", err)
	}

	if err := oldTag.Rename("yt"); err != nil {
		t.Fatalf("rename tag: %v", err)
	}

	if oldTag.Name != "yt" {
		t.Fatalf("expected tag name %q, got %q", "yt", oldTag.Name)
	}
	if TagExists("youtube") {
		t.Fatal("expected the old tag name to disappear")
	}
	renamed, ok := tagCache.Get("yt")
	if !ok {
		t.Fatal("expected renamed tag in cache")
	}
	if renamed.Color != "#ff0000" || renamed.GroupName != "ads" || renamed.Description != "clean" {
		t.Fatalf("expected colour/group/description preserved, got %+v", renamed)
	}

	var storedRule TagRule
	if err := database.DB.First(&storedRule, rule.ID).Error; err != nil {
		t.Fatalf("reload rule: %v", err)
	}
	if storedRule.TagName != "yt" {
		t.Fatalf("expected rule tag %q, got %q", "yt", storedRule.TagName)
	}
	if cached, ok := tagRuleCache.Get(rule.ID); !ok || cached.TagName != "yt" {
		t.Fatalf("expected cached rule tag %q, got %+v", "yt", cached)
	}

	var storedNode Node
	if err := database.DB.First(&storedNode, node.ID).Error; err != nil {
		t.Fatalf("reload node: %v", err)
	}
	if storedNode.Tags != "yt,premium" {
		t.Fatalf("expected node tags %q, got %q", "yt,premium", storedNode.Tags)
	}
	if cached, ok := nodeCache.Get(node.ID); !ok || cached.Tags != "yt,premium" {
		t.Fatalf("expected cached node tags %q, got %+v", "yt,premium", cached)
	}

	var storedSibling Node
	if err := database.DB.First(&storedSibling, sibling.ID).Error; err != nil {
		t.Fatalf("reload sibling node: %v", err)
	}
	if storedSibling.Tags != "youtube-ads" {
		t.Fatalf("expected sibling tags untouched, got %q", storedSibling.Tags)
	}

	var storedSub Subcription
	if err := database.DB.First(&storedSub, sub.ID).Error; err != nil {
		t.Fatalf("reload subscription: %v", err)
	}
	if storedSub.TagWhitelist != "yt, premium" {
		t.Fatalf("expected whitelist %q, got %q", "yt, premium", storedSub.TagWhitelist)
	}
	if storedSub.TagBlacklist != "youtube-ads" {
		t.Fatalf("expected blacklist untouched, got %q", storedSub.TagBlacklist)
	}
	if cached, ok := subcriptionCache.Get(sub.ID); !ok || cached.TagWhitelist != "yt, premium" {
		t.Fatalf("expected cached whitelist %q, got %+v", "yt, premium", cached)
	}

	var storedProfile NodeCheckProfile
	if err := database.DB.First(&storedProfile, profile.ID).Error; err != nil {
		t.Fatalf("reload profile: %v", err)
	}
	if storedProfile.Tags != "yt,other" {
		t.Fatalf("expected profile tags %q, got %q", "yt,other", storedProfile.Tags)
	}

	var storedJob NodeCheckJob
	if err := database.DB.First(&storedJob, job.ID).Error; err != nil {
		t.Fatalf("reload job: %v", err)
	}
	if strings.Join(storedJob.Tags, ",") != "yt,other" {
		t.Fatalf("expected job tags %q, got %v", "yt,other", storedJob.Tags)
	}
}

func TestTagRenameRejectsInvalidNames(t *testing.T) {
	setupTagRenameTestDB(t)

	tag := &Tag{Name: "youtube"}
	if err := tag.Add(); err != nil {
		t.Fatalf("add tag: %v", err)
	}
	if err := (&Tag{Name: "premium"}).Add(); err != nil {
		t.Fatalf("add tag: %v", err)
	}

	cases := []struct {
		name string
		want string
	}{
		{"premium", "already exists"},
		{"a,b", "commas"},
		{"   ", "required"},
		{strings.Repeat("x", TagNameMaxLength+1), "exceed"},
	}
	for _, tc := range cases {
		err := tag.Rename(tc.name)
		if err == nil {
			t.Fatalf("expected rename to %q to fail", tc.name)
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("rename to %q: expected error containing %q, got %v", tc.name, tc.want, err)
		}
		if tag.Name != "youtube" {
			t.Fatalf("expected tag name unchanged after failed rename, got %q", tag.Name)
		}
	}

	if err := tag.Rename("youtube"); err != nil {
		t.Fatalf("renaming to the same name should be a no-op, got %v", err)
	}
}

func TestReplaceTagInCSV(t *testing.T) {
	cases := []struct {
		value    string
		oldName  string
		newName  string
		want     string
		replaced bool
	}{
		{"youtube,premium", "youtube", "yt", "yt,premium", true},
		{" premium , youtube ", "youtube", "yt", " premium ,yt", true},
		{"youtube-ads", "youtube", "yt", "youtube-ads", false},
		{"premium", "youtube", "yt", "premium", false},
		{"", "youtube", "yt", "", false},
	}
	for _, tc := range cases {
		got, replaced := replaceTagInCSV(tc.value, tc.oldName, tc.newName)
		if got != tc.want || replaced != tc.replaced {
			t.Fatalf("replaceTagInCSV(%q, %q, %q) = (%q, %v), want (%q, %v)",
				tc.value, tc.oldName, tc.newName, got, replaced, tc.want, tc.replaced)
		}
	}
}
