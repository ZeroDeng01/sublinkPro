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

func resetIPInfoCacheForTest() {
	ipInfoCache = cache.NewMapCache(func(info IPInfo) string { return info.IP })
}

func setupIPInfoTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "ipinfo_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&IPInfo{}); err != nil {
		t.Fatalf("auto migrate ip_infos: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	resetIPInfoCacheForTest()

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		resetIPInfoCacheForTest()
		testutil.CloseDB(t, db)
	})
}

func TestIPInfoWhitelistUpdatePreservesCreatedAt(t *testing.T) {
	setupIPInfoTestDB(t)

	existing := &IPInfo{
		IP:          "1.1.1.1",
		Country:     "旧国家",
		CountryCode: "OLD",
		City:        "旧城市",
		Provider:    "old-provider",
	}
	if err := database.DB.Create(existing).Error; err != nil {
		t.Fatalf("create existing ip info: %v", err)
	}

	originalCreatedAt := existing.CreatedAt
	if originalCreatedAt.IsZero() {
		t.Fatal("expected created_at to be set on insert")
	}

	time.Sleep(10 * time.Millisecond)

	refreshed := &IPInfo{
		IP:          existing.IP,
		Country:     "新国家",
		CountryCode: "NEW",
		Region:      "new-region",
		RegionName:  "新区",
		City:        "新城市",
		Zip:         "100000",
		Lat:         1.23,
		Lon:         4.56,
		Timezone:    "Asia/Shanghai",
		ISP:         "新ISP",
		Org:         "新组织",
		AS:          "AS123",
		RawResponse: `{"status":"success"}`,
		Provider:    "ip-api.com",
	}

	refreshed.ID = existing.ID
	refreshed.CreatedAt = existing.CreatedAt
	refreshed.UpdatedAt = time.Now()
	if err := database.DB.Model(&IPInfo{}).Where("id = ?", existing.ID).Updates(map[string]interface{}{
		"ip":           refreshed.IP,
		"country":      refreshed.Country,
		"country_code": refreshed.CountryCode,
		"region":       refreshed.Region,
		"region_name":  refreshed.RegionName,
		"city":         refreshed.City,
		"zip":          refreshed.Zip,
		"lat":          refreshed.Lat,
		"lon":          refreshed.Lon,
		"timezone":     refreshed.Timezone,
		"isp":          refreshed.ISP,
		"org":          refreshed.Org,
		"as":           refreshed.AS,
		"raw_response": refreshed.RawResponse,
		"provider":     refreshed.Provider,
		"updated_at":   refreshed.UpdatedAt,
	}).Error; err != nil {
		t.Fatalf("update ip info: %v", err)
	}

	var stored IPInfo
	if err := database.DB.First(&stored, existing.ID).Error; err != nil {
		t.Fatalf("reload ip info: %v", err)
	}

	if !stored.CreatedAt.Equal(originalCreatedAt) {
		t.Fatalf("expected created_at to stay %v, got %v", originalCreatedAt, stored.CreatedAt)
	}
	if stored.Country != refreshed.Country {
		t.Fatalf("expected country %q, got %q", refreshed.Country, stored.Country)
	}
	if !stored.UpdatedAt.After(originalCreatedAt) {
		t.Fatalf("expected updated_at to move forward, created=%v updated=%v", originalCreatedAt, stored.UpdatedAt)
	}
}
