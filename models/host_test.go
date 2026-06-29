package models

import (
	"testing"

	"sublink/cache"
	"sublink/database"
	"sublink/internal/testutil"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func resetHostCacheForTest() {
	hostCache = cache.NewMapCache(func(h Host) int { return h.ID })
	hostCache.AddIndex("hostname", func(h Host) string { return h.Hostname })
}

func setupHostTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "host_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&Host{}); err != nil {
		t.Fatalf("auto migrate hosts: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	resetHostCacheForTest()

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		resetHostCacheForTest()
		testutil.CloseDB(t, db)
	})

	return db
}

func TestHostAddLowercasesHostnameBeforeStorage(t *testing.T) {
	setupHostTestDB(t)

	host := &Host{Hostname: " Example.COM ", IP: "1.1.1.1"}
	if err := host.Add(); err != nil {
		t.Fatalf("add host: %v", err)
	}

	if host.Hostname != "example.com" {
		t.Fatalf("host hostname = %q, want %q", host.Hostname, "example.com")
	}

	loaded, err := GetHostByHostname("EXAMPLE.COM")
	if err != nil {
		t.Fatalf("get host by uppercase hostname: %v", err)
	}
	if loaded.ID != host.ID || loaded.Hostname != "example.com" {
		t.Fatalf("loaded host = %#v, want id %d hostname example.com", loaded, host.ID)
	}
}

func TestNormalizeHostnamesAndDeduplicateKeepsLowestID(t *testing.T) {
	db := setupHostTestDB(t)

	hosts := []Host{
		{Hostname: "Example.COM", IP: "1.1.1.1", Remark: "keep"},
		{Hostname: "example.com", IP: "2.2.2.2", Remark: "delete"},
		{Hostname: " API.Example.COM ", IP: "3.3.3.3", Remark: "trim"},
	}
	if err := db.Create(&hosts).Error; err != nil {
		t.Fatalf("seed hosts: %v", err)
	}

	for i := 0; i < 2; i++ {
		if err := normalizeHostnamesAndDeduplicate(db); err != nil {
			t.Fatalf("normalize run %d: %v", i+1, err)
		}
	}

	var remaining []Host
	if err := db.Order("id ASC").Find(&remaining).Error; err != nil {
		t.Fatalf("load remaining hosts: %v", err)
	}
	if len(remaining) != 2 {
		t.Fatalf("remaining hosts = %d, want 2: %#v", len(remaining), remaining)
	}

	if remaining[0].ID != hosts[0].ID || remaining[0].Hostname != "example.com" || remaining[0].Remark != "keep" {
		t.Fatalf("first remaining host = %#v, want lowest-id example.com", remaining[0])
	}
	if remaining[1].Hostname != "api.example.com" {
		t.Fatalf("second hostname = %q, want api.example.com", remaining[1].Hostname)
	}
}

func TestNormalizeHostsForStorageDeduplicatesByLowercaseHostname(t *testing.T) {
	hosts := NormalizeHostsForStorage([]Host{
		{Hostname: "Example.COM", IP: "1.1.1.1", Remark: "keep"},
		{Hostname: " example.com ", IP: "2.2.2.2", Remark: "drop"},
		{Hostname: "API.Example.COM", IP: "3.3.3.3", Remark: "keep api"},
	})

	if len(hosts) != 2 {
		t.Fatalf("normalized hosts = %d, want 2: %#v", len(hosts), hosts)
	}
	if hosts[0].Hostname != "example.com" || hosts[0].IP != "1.1.1.1" || hosts[0].Remark != "keep" {
		t.Fatalf("first normalized host = %#v, want first example.com record", hosts[0])
	}
	if hosts[1].Hostname != "api.example.com" {
		t.Fatalf("second hostname = %q, want api.example.com", hosts[1].Hostname)
	}
}
