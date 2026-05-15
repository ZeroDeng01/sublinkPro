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

func resetNodeCacheForTest() {
	nodeCache = cache.NewMapCache(func(n Node) int { return n.ID })
	nodeCache.AddIndex("group", func(n Node) string { return n.Group })
	nodeCache.AddIndex("source", func(n Node) string { return n.Source })
	nodeCache.AddIndex("country", func(n Node) string { return n.LinkCountry })
	nodeCache.AddIndex("protocol", func(n Node) string { return n.Protocol })
	nodeCache.AddIndex("sourceID", func(n Node) string { return strconv.Itoa(n.SourceID) })
	nodeCache.AddIndex("name", func(n Node) string { return n.Name })
	nodeCache.AddIndex("contentHash", func(n Node) string { return n.ContentHash })
}

func setupNodeInfoBatchTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "node_info_batch_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&Node{}); err != nil {
		t.Fatalf("auto migrate nodes: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	resetNodeCacheForTest()

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		resetNodeCacheForTest()
		testutil.CloseDB(t, db)
	})
}

func createNodeForBatchUpdateTest(t *testing.T, node Node) Node {
	t.Helper()
	node.NormalizeNameModeDefaults()
	node.syncLinkHash()
	if err := database.DB.Create(&node).Error; err != nil {
		t.Fatalf("create node %q: %v", node.Name, err)
	}
	nodeCache.Set(node.ID, node)
	return node
}

func TestBatchUpdateNodeInfoBulkUpdatesFieldsAndPreservesRemarkNames(t *testing.T) {
	setupNodeInfoBatchTestDB(t)

	originalTime := time.Now().Add(-time.Hour).Truncate(time.Millisecond)
	linkModeNode := createNodeForBatchUpdateTest(t, Node{
		Name:       "旧名称A",
		LinkName:   "旧名称A",
		NameMode:   NodeNameModeLink,
		Link:       "ss://old-a",
		Source:     "机场A",
		SourceSort: 1,
		CreatedAt:  originalTime,
		UpdatedAt:  originalTime,
	})
	remarkNode := createNodeForBatchUpdateTest(t, Node{
		Name:       "我的备注",
		LinkName:   "旧名称B",
		NameMode:   NodeNameModeRemark,
		Link:       "ss://old-b",
		Source:     "机场A",
		SourceSort: 2,
		CreatedAt:  originalTime,
		UpdatedAt:  originalTime,
	})

	updates := []NodeInfoUpdate{
		BuildNodeInfoUpdate(linkModeNode, "新名称A 'quoted'", "ss://new-a?name='quoted'", 11),
		BuildNodeInfoUpdate(remarkNode, "新名称B", "trojan://new-b?password='secret'", 12),
	}
	count, err := BatchUpdateNodeInfo(updates)
	if err != nil {
		t.Fatalf("BatchUpdateNodeInfo() error = %v", err)
	}
	if count != len(updates) {
		t.Fatalf("BatchUpdateNodeInfo() count = %d, want %d", count, len(updates))
	}

	var storedLinkMode Node
	if err := database.DB.First(&storedLinkMode, linkModeNode.ID).Error; err != nil {
		t.Fatalf("reload link mode node: %v", err)
	}
	if storedLinkMode.Name != "新名称A 'quoted'" {
		t.Fatalf("link mode name = %q, want synced upstream name", storedLinkMode.Name)
	}
	if storedLinkMode.LinkName != "新名称A 'quoted'" || storedLinkMode.Link != "ss://new-a?name='quoted'" {
		t.Fatalf("link mode link fields not updated: LinkName=%q Link=%q", storedLinkMode.LinkName, storedLinkMode.Link)
	}
	if storedLinkMode.LinkHash != hashNodeLink("ss://new-a?name='quoted'") {
		t.Fatalf("link mode link_hash = %q, want hash of updated link", storedLinkMode.LinkHash)
	}
	if storedLinkMode.SourceSort != 11 {
		t.Fatalf("link mode source_sort = %d, want 11", storedLinkMode.SourceSort)
	}
	if !storedLinkMode.UpdatedAt.After(originalTime) {
		t.Fatalf("link mode updated_at = %v, want after %v", storedLinkMode.UpdatedAt, originalTime)
	}

	var storedRemark Node
	if err := database.DB.First(&storedRemark, remarkNode.ID).Error; err != nil {
		t.Fatalf("reload remark node: %v", err)
	}
	if storedRemark.Name != "我的备注" {
		t.Fatalf("remark name = %q, want custom remark preserved", storedRemark.Name)
	}
	if storedRemark.LinkName != "新名称B" || storedRemark.Link != "trojan://new-b?password='secret'" {
		t.Fatalf("remark link fields not updated: LinkName=%q Link=%q", storedRemark.LinkName, storedRemark.Link)
	}
	if storedRemark.LinkHash != hashNodeLink("trojan://new-b?password='secret'") {
		t.Fatalf("remark link_hash = %q, want hash of updated link", storedRemark.LinkHash)
	}
	if storedRemark.SourceSort != 12 {
		t.Fatalf("remark source_sort = %d, want 12", storedRemark.SourceSort)
	}
	if !storedRemark.UpdatedAt.After(originalTime) {
		t.Fatalf("remark updated_at = %v, want after %v", storedRemark.UpdatedAt, originalTime)
	}

	if cached, ok := nodeCache.Get(linkModeNode.ID); !ok || cached.Name != storedLinkMode.Name || cached.LinkHash != storedLinkMode.LinkHash || cached.SourceSort != 11 || !cached.UpdatedAt.Equal(storedLinkMode.UpdatedAt) {
		t.Fatalf("link mode cache not updated after successful bulk update: %#v, ok=%v", cached, ok)
	}
	if cached, ok := nodeCache.Get(remarkNode.ID); !ok || cached.Name != "我的备注" || cached.LinkName != "新名称B" || cached.LinkHash != storedRemark.LinkHash || cached.SourceSort != 12 || !cached.UpdatedAt.Equal(storedRemark.UpdatedAt) {
		t.Fatalf("remark cache not updated after successful bulk update: %#v, ok=%v", cached, ok)
	}
}

func TestBatchUpdateNodeInfoFallsBackPerRowWhenBulkChunkFails(t *testing.T) {
	setupNodeInfoBatchTestDB(t)

	conflictingNode := createNodeForBatchUpdateTest(t, Node{Name: "保留节点", LinkName: "保留节点", NameMode: NodeNameModeLink, Link: "ss://reserved", Source: "机场A"})
	failedNode := createNodeForBatchUpdateTest(t, Node{Name: "旧失败", LinkName: "旧失败", NameMode: NodeNameModeLink, Link: "ss://old-failed", Source: "机场A"})
	successNode := createNodeForBatchUpdateTest(t, Node{Name: "旧成功", LinkName: "旧成功", NameMode: NodeNameModeLink, Link: "ss://old-success", Source: "机场A"})

	updates := []NodeInfoUpdate{
		BuildNodeInfoUpdate(failedNode, "会失败", conflictingNode.Link, 21),
		BuildNodeInfoUpdate(successNode, "会成功", "ss://new-success", 22),
	}
	count, err := BatchUpdateNodeInfo(updates)
	if err != nil {
		t.Fatalf("BatchUpdateNodeInfo() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("BatchUpdateNodeInfo() count = %d, want only fallback success", count)
	}

	var storedFailed Node
	if err := database.DB.First(&storedFailed, failedNode.ID).Error; err != nil {
		t.Fatalf("reload failed node: %v", err)
	}
	if storedFailed.Link != failedNode.Link || storedFailed.LinkHash != failedNode.LinkHash || storedFailed.SourceSort != failedNode.SourceSort {
		t.Fatalf("failed fallback row should stay unchanged, got Link=%q LinkHash=%q SourceSort=%d", storedFailed.Link, storedFailed.LinkHash, storedFailed.SourceSort)
	}
	if cached, ok := nodeCache.Get(failedNode.ID); !ok || cached.Link != failedNode.Link || cached.LinkHash != failedNode.LinkHash {
		t.Fatalf("failed fallback row cache should stay unchanged: %#v, ok=%v", cached, ok)
	}

	var storedSuccess Node
	if err := database.DB.First(&storedSuccess, successNode.ID).Error; err != nil {
		t.Fatalf("reload success node: %v", err)
	}
	if storedSuccess.Name != "会成功" || storedSuccess.LinkName != "会成功" || storedSuccess.Link != "ss://new-success" || storedSuccess.LinkHash != hashNodeLink("ss://new-success") || storedSuccess.SourceSort != 22 {
		t.Fatalf("successful fallback row not updated: %#v", storedSuccess)
	}
	if cached, ok := nodeCache.Get(successNode.ID); !ok || cached.Name != "会成功" || cached.Link != "ss://new-success" || cached.LinkHash != hashNodeLink("ss://new-success") || cached.SourceSort != 22 || !cached.UpdatedAt.Equal(storedSuccess.UpdatedAt) {
		t.Fatalf("successful fallback row cache not updated: %#v, ok=%v", cached, ok)
	}
}
