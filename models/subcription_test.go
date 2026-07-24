package models

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"sublink/cache"
	"sublink/database"
	"sublink/internal/testutil"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSubcriptionApplyFiltersDistinguishesHTTPAndHTTPSProtocols(t *testing.T) {
	nodes := []Node{
		{Name: "plain-http", Link: "http://user:pass@example.com:8080#plain-http", Protocol: "http"},
		{Name: "secure-https", Link: "https://user:pass@example.com:8443#secure-https", Protocol: "https"},
	}

	tests := []struct {
		name string
		sub  Subcription
		want []string
	}{
		{
			name: "http whitelist keeps only HTTP nodes",
			sub:  Subcription{ProtocolWhitelist: "http", Nodes: append([]Node(nil), nodes...)},
			want: []string{"plain-http"},
		},
		{
			name: "https whitelist keeps only HTTPS nodes",
			sub:  Subcription{ProtocolWhitelist: "https", Nodes: append([]Node(nil), nodes...)},
			want: []string{"secure-https"},
		},
		{
			name: "http blacklist removes only HTTP nodes",
			sub:  Subcription{ProtocolBlacklist: "http", Nodes: append([]Node(nil), nodes...)},
			want: []string{"secure-https"},
		},
		{
			name: "https blacklist removes only HTTPS nodes",
			sub:  Subcription{ProtocolBlacklist: "https", Nodes: append([]Node(nil), nodes...)},
			want: []string{"plain-http"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sub.ApplyFilters(tt.sub.Nodes)
			if len(got) != len(tt.want) {
				t.Fatalf("filtered nodes = %v, want names %v", nodeNames(got), tt.want)
			}
			for i, node := range got {
				if node.Name != tt.want[i] {
					t.Fatalf("filtered nodes = %v, want names %v", nodeNames(got), tt.want)
				}
			}
		})
	}
}

func nodeNames(nodes []Node) []string {
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.Name)
	}
	return names
}

func resetSubcriptionCacheForTest() {
	subcriptionCache = cache.NewMapCache(func(s Subcription) int { return s.ID })
	subcriptionCache.AddIndex("name", func(s Subcription) string { return s.Name })
}

func resetSubLogsCacheForTest() {
	subLogsCache = cache.NewMapCache(func(sl SubLogs) int { return sl.ID })
	subLogsCache.AddIndex("subcriptionID", func(sl SubLogs) string { return strconv.Itoa(sl.SubcriptionID) })
	subLogsCache.AddIndex("shareID", func(sl SubLogs) string { return strconv.Itoa(sl.ShareID) })
}

func resetChainRuleCacheForTest() {
	chainRuleCache = cache.NewMapCache(func(r SubscriptionChainRule) int { return r.ID })
	chainRuleCache.AddIndex("subscriptionId", func(r SubscriptionChainRule) string { return strconv.Itoa(r.SubscriptionID) })
}

func setupSubcriptionCopyTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized
	oldNodeCache := nodeCache
	oldAirportCache := airportCache
	oldSubcriptionCache := subcriptionCache
	oldSubLogsCache := subLogsCache
	oldSubscriptionShareCache := subscriptionShareCache
	oldChainRuleCache := chainRuleCache

	db, err := gorm.Open(sqlite.Open(testutil.UniqueMemoryDSN(t, "subcription_copy_test")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&Subcription{},
		&SubLogs{},
		&Node{},
		&Airport{},
		&SubcriptionNode{},
		&SubcriptionGroup{},
		&SubcriptionAirport{},
		&SubcriptionScript{},
		&Script{},
		&SubscriptionShare{},
		&SubscriptionChainRule{},
	); err != nil {
		t.Fatalf("auto migrate subscription copy tables: %v", err)
	}
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable SQLite foreign keys: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = true
	resetNodeCacheForTest()
	resetAirportCacheForTest()
	resetSubcriptionCacheForTest()
	resetSubLogsCacheForTest()
	resetSubscriptionShareCacheForTest()
	resetChainRuleCacheForTest()

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		nodeCache = oldNodeCache
		airportCache = oldAirportCache
		subcriptionCache = oldSubcriptionCache
		subLogsCache = oldSubLogsCache
		subscriptionShareCache = oldSubscriptionShareCache
		chainRuleCache = oldChainRuleCache
		testutil.CloseDB(t, db)
	})
}

func createSubcriptionTestAirport(t *testing.T, name string) Airport {
	t.Helper()
	airport := Airport{
		Name:     name,
		URL:      "https://example.com/" + name,
		CronExpr: "0 0 * * *",
		Enabled:  true,
	}
	if err := airport.Add(); err != nil {
		t.Fatalf("add airport %q: %v", name, err)
	}
	return airport
}

func createSubcriptionTestNode(t *testing.T, node Node) Node {
	t.Helper()
	if node.Link == "" {
		node.Link = "ss://" + node.Name
	}
	if node.Protocol == "" {
		node.Protocol = "ss"
	}
	if err := node.Add(); err != nil {
		t.Fatalf("add node %q: %v", node.Name, err)
	}
	return node
}

func TestSubcriptionCopyPreservesUpdateInterval(t *testing.T) {
	setupSubcriptionCopyTestDB(t)

	sub := &Subcription{
		Name:                  "interval-sub",
		Config:                `{"clash":"./template/clash.yaml","surge":"./template/surge.conf"}`,
		RefreshUsageOnRequest: true,
		UpdateInterval:        6,
	}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}

	copySub, err := sub.Copy()
	if err != nil {
		t.Fatalf("copy subscription: %v", err)
	}

	if copySub.UpdateInterval != sub.UpdateInterval {
		t.Fatalf("expected copied UpdateInterval %d, got %d", sub.UpdateInterval, copySub.UpdateInterval)
	}
}

func TestSubcriptionAirportOnlyResolvesCurrentAirportNodes(t *testing.T) {
	setupSubcriptionCopyTestDB(t)

	airport := createSubcriptionTestAirport(t, "airport-dynamic")
	createSubcriptionTestNode(t, Node{Name: "airport-dynamic-b", LinkName: "airport-dynamic-b", Source: airport.Name, SourceID: airport.ID, SourceSort: 2})
	createSubcriptionTestNode(t, Node{Name: "airport-dynamic-a", LinkName: "airport-dynamic-a", Source: airport.Name, SourceID: airport.ID, SourceSort: 1})

	sub := &Subcription{Name: "airport-only", RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}
	if err := sub.AddAirports([]int{airport.ID}); err != nil {
		t.Fatalf("add airport relation: %v", err)
	}

	if err := sub.GetSub("clash"); err != nil {
		t.Fatalf("get subscription: %v", err)
	}
	if got, want := nodeNames(sub.Nodes), []string{"airport-dynamic-a", "airport-dynamic-b"}; !sameStrings(got, want) {
		t.Fatalf("nodes = %v, want %v", got, want)
	}

	createSubcriptionTestNode(t, Node{Name: "airport-dynamic-c", LinkName: "airport-dynamic-c", Source: airport.Name, SourceID: airport.ID, SourceSort: 3})
	if err := sub.GetSub("clash"); err != nil {
		t.Fatalf("get subscription after adding airport node: %v", err)
	}
	if got, want := nodeNames(sub.Nodes), []string{"airport-dynamic-a", "airport-dynamic-b", "airport-dynamic-c"}; !sameStrings(got, want) {
		t.Fatalf("nodes = %v, want %v", got, want)
	}
}

func TestSubcriptionMixedDirectGroupAirportDedupesAndOrders(t *testing.T) {
	setupSubcriptionCopyTestDB(t)

	airport := createSubcriptionTestAirport(t, "airport-mixed")
	direct := createSubcriptionTestNode(t, Node{Name: "mixed-direct", LinkName: "mixed-direct", Source: "manual"})
	createSubcriptionTestNode(t, Node{Name: "mixed-group-first", LinkName: "mixed-group-first", Source: "manual", Group: "mixed-group"})
	createSubcriptionTestNode(t, Node{Name: "mixed-shared", LinkName: "mixed-shared", Source: "manual", Group: "mixed-group"})
	createSubcriptionTestNode(t, Node{Name: "mixed-airport-only", LinkName: "mixed-airport-only", Source: airport.Name, SourceID: airport.ID, SourceSort: 1})
	createSubcriptionTestNode(t, Node{Name: "mixed-airport-shared", LinkName: "mixed-shared", Source: airport.Name, SourceID: airport.ID, SourceSort: 2})

	sub := &Subcription{Name: "mixed-sub", RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}
	if err := database.DB.Create(&SubcriptionNode{SubcriptionID: sub.ID, NodeID: direct.ID, Sort: 1}).Error; err != nil {
		t.Fatalf("add node relation: %v", err)
	}
	if err := database.DB.Create(&SubcriptionGroup{SubcriptionID: sub.ID, GroupName: "mixed-group", Sort: 0}).Error; err != nil {
		t.Fatalf("add group relation: %v", err)
	}
	if err := database.DB.Create(&SubcriptionAirport{SubcriptionID: sub.ID, AirportID: airport.ID, Sort: 2}).Error; err != nil {
		t.Fatalf("add airport relation: %v", err)
	}

	if err := sub.GetSub("clash"); err != nil {
		t.Fatalf("get subscription: %v", err)
	}
	if got, want := nodeNames(sub.Nodes), []string{"mixed-group-first", "mixed-shared", "mixed-direct", "mixed-airport-only"}; !sameStrings(got, want) {
		t.Fatalf("nodes = %v, want %v", got, want)
	}
}

func TestSubcriptionCopyPreservesAirportRelations(t *testing.T) {
	setupSubcriptionCopyTestDB(t)

	airport := createSubcriptionTestAirport(t, "airport-copy")
	sub := &Subcription{Name: "airport-copy-sub", RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}
	if err := database.DB.Create(&SubcriptionAirport{SubcriptionID: sub.ID, AirportID: airport.ID, Sort: 7}).Error; err != nil {
		t.Fatalf("add airport relation: %v", err)
	}

	copySub, err := sub.Copy()
	if err != nil {
		t.Fatalf("copy subscription: %v", err)
	}

	var relation SubcriptionAirport
	if err := database.DB.Where("subcription_id = ? AND airport_id = ?", copySub.ID, airport.ID).First(&relation).Error; err != nil {
		t.Fatalf("find copied airport relation: %v", err)
	}
	if relation.Sort != 7 {
		t.Fatalf("copied airport sort = %d, want 7", relation.Sort)
	}
}

func TestSubcriptionDeleteCleansAirportRelations(t *testing.T) {
	setupSubcriptionCopyTestDB(t)

	airport := createSubcriptionTestAirport(t, "airport-delete")
	sub := &Subcription{Name: "airport-delete-sub", RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}
	if err := sub.AddAirports([]int{airport.ID}); err != nil {
		t.Fatalf("add airport relation: %v", err)
	}
	if err := sub.Del(); err != nil {
		t.Fatalf("delete subscription: %v", err)
	}

	var count int64
	if err := database.DB.Model(&SubcriptionAirport{}).Where("subcription_id = ?", sub.ID).Count(&count).Error; err != nil {
		t.Fatalf("count airport relations: %v", err)
	}
	if count != 0 {
		t.Fatalf("airport relation count = %d, want 0", count)
	}
}

func TestSubcriptionDeleteCleansSubLogs(t *testing.T) {
	setupSubcriptionCopyTestDB(t)

	sub := &Subcription{Name: "sub-log-delete-sub", RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}
	for _, ip := range []string{"192.0.2.1", "192.0.2.2"} {
		if err := (&SubLogs{IP: ip, SubcriptionID: sub.ID}).Add(); err != nil {
			t.Fatalf("add subscription log %q: %v", ip, err)
		}
	}
	if count := len(GetSubLogsBySubcriptionID(sub.ID)); count != 2 {
		t.Fatalf("cached subscription log count = %d, want 2", count)
	}
	share := &SubscriptionShare{
		SubscriptionID: sub.ID,
		Name:           "delete-share",
		Token:          "delete-share-token",
		ExpireType:     ExpireTypeNever,
		Enabled:        true,
	}
	if err := share.Add(); err != nil {
		t.Fatalf("add subscription share: %v", err)
	}
	if count := len(GetSharesBySubscriptionID(sub.ID)); count != 1 {
		t.Fatalf("cached subscription share count = %d, want 1", count)
	}
	chainRule := &SubscriptionChainRule{SubscriptionID: sub.ID, Name: "delete-chain-rule", Enabled: true}
	if err := chainRule.Add(); err != nil {
		t.Fatalf("add subscription chain rule: %v", err)
	}
	if count := len(GetChainRulesBySubscriptionID(sub.ID)); count != 1 {
		t.Fatalf("cached subscription chain rule count = %d, want 1", count)
	}

	if err := sub.Del(); err != nil {
		t.Fatalf("delete subscription: %v", err)
	}

	var subscriptionCount int64
	if err := database.DB.Unscoped().Model(&Subcription{}).Where("id = ?", sub.ID).Count(&subscriptionCount).Error; err != nil {
		t.Fatalf("count subscriptions: %v", err)
	}
	if subscriptionCount != 0 {
		t.Fatalf("subscription count = %d, want 0", subscriptionCount)
	}

	var subLogCount int64
	if err := database.DB.Model(&SubLogs{}).Where("subcription_id = ?", sub.ID).Count(&subLogCount).Error; err != nil {
		t.Fatalf("count subscription logs: %v", err)
	}
	if subLogCount != 0 {
		t.Fatalf("subscription log count = %d, want 0", subLogCount)
	}
	if count := len(GetSubLogsBySubcriptionID(sub.ID)); count != 0 {
		t.Fatalf("cached subscription log count = %d, want 0", count)
	}

	var subscriptionShareCount int64
	if err := database.DB.Model(&SubscriptionShare{}).Where("subscription_id = ?", sub.ID).Count(&subscriptionShareCount).Error; err != nil {
		t.Fatalf("count subscription shares: %v", err)
	}
	if subscriptionShareCount != 0 {
		t.Fatalf("subscription share count = %d, want 0", subscriptionShareCount)
	}
	if count := len(GetSharesBySubscriptionID(sub.ID)); count != 0 {
		t.Fatalf("cached subscription share count = %d, want 0", count)
	}

	var chainRuleCount int64
	if err := database.DB.Model(&SubscriptionChainRule{}).Where("subscription_id = ?", sub.ID).Count(&chainRuleCount).Error; err != nil {
		t.Fatalf("count subscription chain rules: %v", err)
	}
	if chainRuleCount != 0 {
		t.Fatalf("subscription chain rule count = %d, want 0", chainRuleCount)
	}
	if count := len(GetChainRulesBySubscriptionID(sub.ID)); count != 0 {
		t.Fatalf("cached subscription chain rule count = %d, want 0", count)
	}
}

func TestSubcriptionListIncludesAirportRelationWithSort(t *testing.T) {
	setupSubcriptionCopyTestDB(t)

	airport := createSubcriptionTestAirport(t, "airport-list")
	sub := &Subcription{Name: "airport-list-sub", RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}
	if err := database.DB.Create(&SubcriptionAirport{SubcriptionID: sub.ID, AirportID: airport.ID, Sort: 5}).Error; err != nil {
		t.Fatalf("add airport relation: %v", err)
	}

	subs, err := sub.List()
	if err != nil {
		t.Fatalf("list subscriptions: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("subscription count = %d, want 1", len(subs))
	}
	if len(subs[0].AirportsWithSort) != 1 {
		t.Fatalf("airport relation count = %d, want 1", len(subs[0].AirportsWithSort))
	}
	gotAirport := subs[0].AirportsWithSort[0]
	if gotAirport.ID != airport.ID || gotAirport.Sort != 5 {
		t.Fatalf("airport relation = {id:%d sort:%d}, want {id:%d sort:5}", gotAirport.ID, gotAirport.Sort, airport.ID)
	}

	payload, err := json.Marshal(subs[0].AirportsWithSort)
	if err != nil {
		t.Fatalf("marshal airport relation: %v", err)
	}
	serialized := string(payload)
	for _, forbidden := range []string{"url", "proxyLink", "requestHeaders"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("airport relation leaked %q in JSON: %s", forbidden, serialized)
		}
	}
}

func TestSubcriptionAddAirportsRejectsMissingAirport(t *testing.T) {
	setupSubcriptionCopyTestDB(t)

	sub := &Subcription{Name: "missing-airport-sub", RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatalf("add subscription: %v", err)
	}
	if err := sub.AddAirports([]int{404}); err == nil {
		t.Fatal("expected missing airport relation to be rejected")
	}

	var count int64
	if err := database.DB.Model(&SubcriptionAirport{}).Where("subcription_id = ?", sub.ID).Count(&count).Error; err != nil {
		t.Fatalf("count airport relations: %v", err)
	}
	if count != 0 {
		t.Fatalf("airport relation count = %d, want 0", count)
	}
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
