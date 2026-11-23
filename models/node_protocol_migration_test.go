package models

import (
	"strings"
	"sublink/node/protocol"
	"testing"

	"sublink/database"
	"sublink/internal/testutil"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupNodeProtocolMigrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "node_protocol_migration_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&Node{}); err != nil {
		t.Fatalf("auto migrate node table: %v", err)
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

	return db
}

func TestRepairHTTPHTTPSNodeProtocolFromLinkCorrectsLegacyRows(t *testing.T) {
	db := setupNodeProtocolMigrationTestDB(t)

	nodes := []Node{
		{Name: "legacy-https", Link: "https://user:pass@example.com:8443#legacy-https", Protocol: "http"},
		{Name: "legacy-http", Link: "http://user:pass@example.com:8080#legacy-http", Protocol: "https"},
		{Name: "already-http", Link: "http://user:pass@example.com:8081#already-http", Protocol: "http"},
		{Name: "non-http", Link: "vmess://example", Protocol: "vmess"},
		{Name: "empty-link", Link: "", Protocol: ""},
	}
	for i := range nodes {
		nodes[i].syncLinkHash()
	}
	if err := db.Create(&nodes).Error; err != nil {
		t.Fatalf("seed nodes: %v", err)
	}

	for i := 0; i < 2; i++ {
		if err := repairHTTPHTTPSNodeProtocolFromLink(db); err != nil {
			t.Fatalf("repair run %d: %v", i+1, err)
		}
	}

	wantProtocols := map[string]string{
		"legacy-https": "https",
		"legacy-http":  "http",
		"already-http": "http",
		"non-http":     "vmess",
		"empty-link":   "",
	}
	for name, want := range wantProtocols {
		var node Node
		if err := db.Where("name = ?", name).First(&node).Error; err != nil {
			t.Fatalf("load node %s: %v", name, err)
		}
		if node.Protocol != want {
			t.Fatalf("node %s protocol = %q, want %q", name, node.Protocol, want)
		}
	}
}

func TestUpgradeMieruNodeLinksPreservesExistingRows(t *testing.T) {
	db := setupNodeProtocolMigrationTestDB(t)

	nodes := []Node{
		{
			Link: "mieru://user:pass@[2001:db8::1]?portRange=2000-2003&transport=UDP#node",
			Name: "remark", NameMode: NodeNameModeRemark, Group: "group", Speed: 12,
			ClashExtra: "handshake-mode: HANDSHAKE_NO_WAIT\n",
		},
		{Link: "mierus://user:pass@host.example?port=11211&profile=default&protocol=TCP", Name: "simple", Protocol: "other"},
		{Link: "ss://opaque", Name: "unrelated", Protocol: "ss"},
	}
	for i := range nodes {
		NormalizeNodeForImport(&nodes[i])
	}
	if err := db.Create(&nodes).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := upgradeMieruNodeLinks(db); err != nil {
			t.Fatal(err)
		}
	}

	var after []Node
	if err := db.Order("id").Find(&after).Error; err != nil || len(after) != len(nodes) {
		t.Fatalf("got %d rows, err = %v", len(after), err)
	}
	legacy := after[0]
	if legacy.ID != nodes[0].ID || legacy.Name != "remark" || legacy.Group != "group" || legacy.Speed != 12 ||
		legacy.LinkHash == nodes[0].LinkHash || legacy.ContentHash == "" || legacy.ClashExtra != "" ||
		legacy.Protocol != "mieru" || legacy.LinkPort != "2000-2003" || !strings.HasPrefix(legacy.Link, "mierus://") {
		t.Fatal("migration did not preserve node metadata or regenerate derived fields")
	}
	info, err := protocol.DecodeMieruURL(legacy.Link)
	if err != nil || info.HandshakeMode != "HANDSHAKE_NO_WAIT" {
		t.Fatalf("legacy handshake setting was not migrated, err = %v", err)
	}
	if after[1].Link != nodes[1].Link || after[1].Protocol != "mieru" || after[2].Link != nodes[2].Link || after[2].Protocol != "ss" {
		t.Fatal("migration did not preserve official and unrelated links")
	}
}
