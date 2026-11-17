package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTagAPITestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "tag_api_test")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.Tag{},
		&models.TagRule{},
		&models.Node{},
		&models.Subcription{},
		&models.NodeCheckProfile{},
		&models.NodeCheckJob{},
	); err != nil {
		t.Fatalf("auto migrate tag api tables: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	if err := models.InitTagCache(); err != nil {
		t.Fatalf("init tag cache: %v", err)
	}
	if err := models.InitTagRuleCache(); err != nil {
		t.Fatalf("init tag rule cache: %v", err)
	}

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		if oldDB != nil {
			_ = models.InitTagCache()
			_ = models.InitTagRuleCache()
		}
		testutil.CloseDB(t, db)
	})
}

func TestTagUpdateRenamesTagAndReferences(t *testing.T) {
	setupTagAPITestDB(t)

	tag := &models.Tag{Name: "youtube", GroupName: "ads", Color: "#ff0000"}
	if err := tag.Add(); err != nil {
		t.Fatalf("add tag: %v", err)
	}
	rule := &models.TagRule{TagName: "youtube", Name: "rule", Enabled: true, TriggerType: "speed_test"}
	if err := rule.Add(); err != nil {
		t.Fatalf("add rule: %v", err)
	}
	node := models.Node{Link: "ss://api-rename", LinkHash: "api-rename-hash", Name: "api-node", Tags: "youtube"}
	if err := database.DB.Create(&node).Error; err != nil {
		t.Fatalf("create node: %v", err)
	}
	profile := &models.NodeCheckProfile{Name: "api-profile", Tags: "youtube"}
	if err := profile.Add(); err != nil {
		t.Fatalf("add profile: %v", err)
	}

	recorder := performJSONRequest(t, TagUpdate, http.MethodPost, map[string]any{
		"name":        "youtube",
		"newName":     "yt",
		"color":       "#00ff00",
		"description": "renamed",
		"groupName":   "ads",
	})
	response := decodeAPIResponse(t, recorder)
	if response.Code != 200 {
		t.Fatalf("expected success, got code %d msg %q", response.Code, response.Msg)
	}

	var payload models.Tag
	if err := json.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("decode tag payload: %v", err)
	}
	if payload.Name != "yt" || payload.Color != "#00ff00" || payload.Description != "renamed" {
		t.Fatalf("unexpected tag payload: %+v", payload)
	}
	if models.TagExists("youtube") {
		t.Fatal("expected the old tag name to be gone")
	}

	var storedTag models.Tag
	if err := database.DB.First(&storedTag, "name = ?", "yt").Error; err != nil {
		t.Fatalf("reload tag: %v", err)
	}
	if storedTag.GroupName != "ads" {
		t.Fatalf("expected group preserved, got %q", storedTag.GroupName)
	}

	var storedRule models.TagRule
	if err := database.DB.First(&storedRule, rule.ID).Error; err != nil {
		t.Fatalf("reload rule: %v", err)
	}
	if storedRule.TagName != "yt" {
		t.Fatalf("expected rule to follow the rename, got %q", storedRule.TagName)
	}

	var storedNode models.Node
	if err := database.DB.First(&storedNode, node.ID).Error; err != nil {
		t.Fatalf("reload node: %v", err)
	}
	if storedNode.Tags != "yt" {
		t.Fatalf("expected node tags %q, got %q", "yt", storedNode.Tags)
	}

	var storedProfile models.NodeCheckProfile
	if err := database.DB.First(&storedProfile, profile.ID).Error; err != nil {
		t.Fatalf("reload profile: %v", err)
	}
	if storedProfile.Tags != "yt" {
		t.Fatalf("expected profile tags %q, got %q", "yt", storedProfile.Tags)
	}
}

func TestTagUpdateRejectsDuplicateName(t *testing.T) {
	setupTagAPITestDB(t)

	for _, name := range []string{"youtube", "premium"} {
		if err := (&models.Tag{Name: name}).Add(); err != nil {
			t.Fatalf("add tag %q: %v", name, err)
		}
	}

	recorder := performJSONRequest(t, TagUpdate, http.MethodPost, map[string]any{
		"name":    "youtube",
		"newName": "premium",
	})
	response := decodeAPIResponse(t, recorder)
	if response.Code == 200 {
		t.Fatal("expected the rename to fail")
	}
	if !strings.Contains(response.Msg, "already exists") {
		t.Fatalf("expected a duplicate-name error, got %q", response.Msg)
	}

	var stored models.Tag
	if err := database.DB.First(&stored, "name = ?", "youtube").Error; err != nil {
		t.Fatalf("expected the original tag to survive: %v", err)
	}
}

func TestTagUpdateKeepsNameWithoutNewName(t *testing.T) {
	setupTagAPITestDB(t)

	if err := (&models.Tag{Name: "youtube", Color: "#ff0000"}).Add(); err != nil {
		t.Fatalf("add tag: %v", err)
	}

	recorder := performJSONRequest(t, TagUpdate, http.MethodPost, map[string]any{
		"name":        "youtube",
		"color":       "#0000ff",
		"description": "just a colour change",
	})
	response := decodeAPIResponse(t, recorder)
	if response.Code != 200 {
		t.Fatalf("expected success, got code %d msg %q", response.Code, response.Msg)
	}

	var stored models.Tag
	if err := database.DB.First(&stored, "name = ?", "youtube").Error; err != nil {
		t.Fatalf("reload tag: %v", err)
	}
	if stored.Color != "#0000ff" || stored.Description != "just a colour change" {
		t.Fatalf("expected colour and description updated, got %+v", stored)
	}
}

func TestTagUpdateReturnsNotFoundForUnknownTag(t *testing.T) {
	setupTagAPITestDB(t)

	recorder := performJSONRequest(t, TagUpdate, http.MethodPost, map[string]any{"name": "missing"})
	response := decodeAPIResponse(t, recorder)
	if response.Code != 404 {
		t.Fatalf("expected 404, got code %d msg %q", response.Code, response.Msg)
	}
}

func TestTagAddRejectsInvalidNames(t *testing.T) {
	setupTagAPITestDB(t)

	cases := []struct {
		name string
		want string
	}{
		{"   ", "required"},
		{"a,b", "commas"},
		{strings.Repeat("长", models.TagNameMaxLength+1), "exceed"},
	}
	for _, tc := range cases {
		recorder := performJSONRequest(t, TagAdd, http.MethodPost, map[string]any{"name": tc.name})
		response := decodeAPIResponse(t, recorder)
		if response.Code == 200 {
			t.Fatalf("expected adding %q to fail", tc.name)
		}
		if !strings.Contains(response.Msg, tc.want) {
			t.Fatalf("adding %q: expected error containing %q, got %q", tc.name, tc.want, response.Msg)
		}
	}

	recorder := performJSONRequest(t, TagAdd, http.MethodPost, map[string]any{
		"name":  "  youtube  ",
		"color": "#ff0000",
	})
	response := decodeAPIResponse(t, recorder)
	if response.Code != 200 {
		t.Fatalf("expected the trimmed name to be accepted, got code %d msg %q", response.Code, response.Msg)
	}
	var payload models.Tag
	if err := json.Unmarshal(response.Data, &payload); err != nil {
		t.Fatalf("decode tag payload: %v", err)
	}
	if payload.Name != "youtube" {
		t.Fatalf("expected the trimmed name %q, got %q", "youtube", payload.Name)
	}
}
