package models

import (
	"testing"

	"sublink/database"
	"sublink/internal/testutil"
)

func setupShareTokenRotationTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db := testutil.OpenMemoryDB(t, "share_token_rotation")
	if err := db.AutoMigrate(&Subcription{}, &SubscriptionShare{}); err != nil {
		t.Fatalf("auto migrate share tables: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = true

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		testutil.CloseDB(t, db)
	})
}

func TestRotatePredictableShareTokens(t *testing.T) {
	setupShareTokenRotationTestDB(t)

	sub := &Subcription{Name: "legacy-sub"}
	if err := database.DB.Create(sub).Error; err != nil {
		t.Fatalf("create subscription: %v", err)
	}

	predictable := md5Hash(sub.Name)
	legacy := &SubscriptionShare{
		SubscriptionID: sub.ID,
		Token:          predictable,
		Name:           "Default share link",
		ExpireType:     ExpireTypeNever,
		IsLegacy:       true,
		Enabled:        true,
	}
	if err := database.DB.Create(legacy).Error; err != nil {
		t.Fatalf("create legacy share: %v", err)
	}

	custom := &SubscriptionShare{
		SubscriptionID: sub.ID,
		Token:          "custom-token",
		Name:           "custom share",
		Enabled:        true,
	}
	if err := database.DB.Create(custom).Error; err != nil {
		t.Fatalf("create custom share: %v", err)
	}

	if err := rotatePredictableShareTokens(database.DB); err != nil {
		t.Fatalf("rotate predictable share tokens: %v", err)
	}

	var rotated SubscriptionShare
	if err := database.DB.First(&rotated, legacy.ID).Error; err != nil {
		t.Fatalf("reload legacy share: %v", err)
	}
	if rotated.Token == predictable {
		t.Fatal("expected the predictable token to be replaced")
	}
	if len(rotated.Token) != 32 {
		t.Fatalf("expected a 32-character hex token, got %q", rotated.Token)
	}

	var untouched SubscriptionShare
	if err := database.DB.First(&untouched, custom.ID).Error; err != nil {
		t.Fatalf("reload custom share: %v", err)
	}
	if untouched.Token != "custom-token" {
		t.Fatalf("expected the custom token to stay untouched, got %q", untouched.Token)
	}

	// A second run must be a no-op: the rotated token no longer matches md5(name).
	if err := rotatePredictableShareTokens(database.DB); err != nil {
		t.Fatalf("second rotation run: %v", err)
	}
	var again SubscriptionShare
	if err := database.DB.First(&again, legacy.ID).Error; err != nil {
		t.Fatalf("reload rotated share: %v", err)
	}
	if again.Token != rotated.Token {
		t.Fatalf("expected the rotated token to stay stable, got %q -> %q", rotated.Token, again.Token)
	}
}
