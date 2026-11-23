package node

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sublink/database"
	"sublink/models"
	"sublink/node/protocol"
	"sync/atomic"
	"testing"
)

type airportSyncTestReporter struct {
	NoOpTaskReporter
	message string
	result  any
}

func (r *airportSyncTestReporter) ReportComplete(message string, result any) {
	r.message, r.result = message, result
}

func newAirportSyncTestAirport(t *testing.T, name, url string) *models.Airport {
	t.Helper()
	airport := &models.Airport{Name: name, URL: url, NodeNameIntraUniquify: true}
	if err := airport.Add(); err != nil {
		t.Fatal(err)
	}
	return airport
}

func mustLoadAirportSyncTest(t *testing.T, airport *models.Airport, reporter TaskReporter) []int {
	t.Helper()
	ids, _, err := LoadClashConfigFromURLWithReporter(context.Background(), airport.ID, airport.URL, airport.Name, false, "", "", nil, reporter, false, false)
	if err != nil {
		t.Fatal(err)
	}
	return ids
}

func fetchAirportSyncTestRows(t *testing.T, airportID int) []models.Node {
	t.Helper()
	var nodes []models.Node
	if err := database.DB.Where("source_id = ?", airportID).Order("id").Find(&nodes).Error; err != nil {
		t.Fatal(err)
	}
	return nodes
}

func TestAirportSyncPreservesLegacyRowsAcrossReorderedYAML(t *testing.T) {
	setupSubscriptionCountryBackfillTestDB(t)
	if err := models.SetSetting("cross_airport_dedup_enabled", "false"); err != nil {
		t.Fatal(err)
	}
	var revised atomic.Bool
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "proxies:\n")
		if revised.Load() {
			_, _ = io.WriteString(w, "\n")
		}
		for i := 1; i <= 142; i++ {
			position, hostVersion := i, 1
			if revised.Load() {
				position = 143 - i
				if position > 136 {
					hostVersion = 2
				}
			}
			_, _ = fmt.Fprintf(w, "  - {name: node-%03d, type: trojan, server: host-%d-%d.example.com, port: 443, password: fixture}\n", position, hostVersion, position)
		}
	}))
	t.Cleanup(upstream.Close)
	airport := newAirportSyncTestAirport(t, "legacy", upstream.URL)
	if ids := mustLoadAirportSyncTest(t, airport, nil); len(ids) != 142 {
		t.Fatalf("initial insert yielded %d IDs", len(ids))
	}
	old := fetchAirportSyncTestRows(t, airport.ID)
	// Simulate pre-upgrade hashes that differ from the canonical hash for every row.
	if err := database.DB.Exec("UPDATE nodes SET content_hash = link_hash WHERE source_id = ?", airport.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := models.UpdateNodeFields(old[0].ID, map[string]any{
		"name": "custom remark", "name_mode": models.NodeNameModeRemark, "tags": "1,2", "delay_time": 125, "speed": 3.5,
	}); err != nil {
		t.Fatal(err)
	}
	binding := models.SubcriptionNode{SubcriptionID: 20, NodeID: old[0].ID}
	if err := database.DB.Create(&binding).Error; err != nil {
		t.Fatal(err)
	}
	// A stale entry must not drive reconciliation or survive a successful refresh.
	shadow := old[1]
	shadow.Link, shadow.LinkName = old[2].Link, old[2].LinkName
	models.UpdateNodeCache(shadow.ID, shadow)

	revised.Store(true)
	reporter := &airportSyncTestReporter{}
	mustLoadAirportSyncTest(t, airport, reporter)
	nodes := fetchAirportSyncTestRows(t, airport.ID)
	if len(nodes) != 142 {
		t.Fatalf("database contains %d nodes after upgrade, want 142", len(nodes))
	}
	for i := range 136 {
		if nodes[i].ID != old[i].ID || nodes[i].ContentHash != old[i].ContentHash || nodes[i].SourceSort != 142-i {
			t.Fatalf("original row %d was not reconciled in place", old[i].ID)
		}
	}
	if nodes[0].Name != "custom remark" || nodes[0].Tags != "1,2" || nodes[0].DelayTime != 125 || nodes[0].Speed != 3.5 {
		t.Fatal("the surviving node lost its metadata")
	}
	if err := database.DB.Where("subcription_id = ? AND node_id = ?", binding.SubcriptionID, binding.NodeID).First(&models.SubcriptionNode{}).Error; err != nil {
		t.Fatalf("direct subscription binding was lost: %v", err)
	}
	for _, row := range nodes {
		cached, ok := models.GetNodeByID(row.ID)
		if !ok || cached.Link != row.Link || cached.ContentHash != row.ContentHash {
			t.Fatalf("cached node ID %d does not match persisted row", row.ID)
		}
	}
	if stats, ok := reporter.result.(map[string]any); !ok || stats["added"] != 6 || stats["deleted"] != 6 {
		t.Fatalf("sync result = %v, want six inserts and six removals", reporter.result)
	}
	if ids := mustLoadAirportSyncTest(t, airport, nil); len(ids) != 0 {
		t.Fatalf("unchanged snapshot reported changed IDs: %v", ids)
	}
	storedAirport, err := models.GetAirportByID(airport.ID)
	if err != nil || storedAirport.SuccessCount != 142 {
		t.Fatalf("refresh count not based on the 142 resolved nodes: airport=%v err=%v", storedAirport, err)
	}
}

func TestAirportSyncRollsBackWhenRemovingStaleRowsFails(t *testing.T) {
	setupSubscriptionCountryBackfillTestDB(t)
	airport := newAirportSyncTestAirport(t, "rollback", "https://example.test/rollback")
	old := protocol.Proxy{Type: "trojan", Name: "old", Server: "old.example.com", Port: 443, Password: "fixture"}
	kept := protocol.Proxy{Type: "trojan", Name: "kept", Server: "kept.example.com", Port: 443, Password: "fixture"}
	if _, err := scheduleClashToNodeLinks(context.Background(), airport.ID, []protocol.Proxy{old, kept}, airport.Name, nil, nil, false); err != nil {
		t.Fatal(err)
	}
	before := fetchAirportSyncTestRows(t, airport.ID)
	if err := database.DB.Create(&models.SubcriptionNode{SubcriptionID: 30, NodeID: before[0].ID}).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Exec("CREATE TRIGGER reject_sync_delete BEFORE DELETE ON nodes BEGIN SELECT RAISE(ABORT, 'fixture rejects delete'); END").Error; err != nil {
		t.Fatal(err)
	}
	newNode := old
	newNode.Name, newNode.Server = "new", "new.example.com"
	kept.Name = "rename"
	rep := &airportSyncTestReporter{}
	if ids, err := scheduleClashToNodeLinks(context.Background(), airport.ID, []protocol.Proxy{kept, newNode}, airport.Name, rep, nil, false); err == nil || len(ids) != 0 || rep.result != nil {
		t.Fatalf("failed refresh reported success: IDs=%v err=%v result=%v", ids, err, rep.result)
	}
	if after := fetchAirportSyncTestRows(t, airport.ID); !slices.Equal(before, after) {
		t.Fatal("the failed transaction changed persisted nodes")
	}
	if err := database.DB.Where("subcription_id = ?", 30).First(&models.SubcriptionNode{}).Error; err != nil {
		t.Fatalf("association delete was not rolled back: %v", err)
	}
	for _, row := range before {
		cached, ok := models.GetNodeByID(row.ID)
		if !ok || cached.Link != row.Link {
			t.Fatalf("failed transaction changed cache for node %d", row.ID)
		}
	}
}

func TestAirportSyncRollsBackWhenAnInsertFails(t *testing.T) {
	setupSubscriptionCountryBackfillTestDB(t)
	airport := newAirportSyncTestAirport(t, "insert failure", "https://example.test/insert-failure")
	proxy := protocol.Proxy{Type: "trojan", Name: "old", Server: "old.example.com", Port: 443, Password: "fixture"}
	refresh := func(proxies []protocol.Proxy, reporter TaskReporter) ([]int, error) {
		return scheduleClashToNodeLinks(context.Background(), airport.ID, proxies, airport.Name, reporter, nil, false)
	}
	if _, err := refresh([]protocol.Proxy{proxy}, nil); err != nil {
		t.Fatal(err)
	}
	before := fetchAirportSyncTestRows(t, airport.ID)
	if err := database.DB.Exec("CREATE TRIGGER reject_sync_insert BEFORE INSERT ON nodes WHEN NEW.link_name = 'node-100' BEGIN SELECT RAISE(ABORT, 'fixture rejects insert'); END").Error; err != nil {
		t.Fatal(err)
	}
	proxies := make([]protocol.Proxy, 101)
	for i := range proxies {
		proxies[i] = proxy
		proxies[i].Name = fmt.Sprintf("node-%d", i)
		proxies[i].Server = proxies[i].Name + ".example.com"
	}
	reporter := &airportSyncTestReporter{}
	if ids, err := refresh(proxies, reporter); err == nil || len(ids) != 0 || reporter.result != nil {
		t.Fatalf("failed second chunk reported success: IDs=%v err=%v result=%v", ids, err, reporter.result)
	}
	if rows := fetchAirportSyncTestRows(t, airport.ID); !slices.Equal(before, rows) {
		t.Fatal("a failed insert left committed replacements or deleted the old node")
	}
	if nodes, err := models.ListBySourceID(airport.ID); err != nil || len(nodes) != 1 || nodes[0].ID != before[0].ID {
		t.Fatalf("failed insert changed the cache: count=%d err=%v", len(nodes), err)
	}
	if err := database.DB.Exec("DROP TRIGGER reject_sync_insert").Error; err != nil {
		t.Fatal(err)
	}
	if ids, err := refresh(proxies, nil); err != nil || len(ids) != 101 {
		t.Fatalf("retry after removing the failure: count=%d error=%v", len(ids), err)
	}
}

func TestAirportSyncRetainsNodesFromIncompleteSnapshots(t *testing.T) {
	const header = "proxies:\n"
	const a = "  - {name: A, type: trojan, server: a.example.com, port: 443, password: fixture}\n"
	const b = "  - {name: B, type: trojan, server: b.example.com, port: 443, password: fixture}\n"
	for _, mode := range []string{"provider", "unsupported YAML entry", "invalid URI"} {
		t.Run(mode, func(t *testing.T) {
			setupSubscriptionCountryBackfillTestDB(t)
			var incomplete atomic.Bool
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch mode {
				case "provider":
					switch r.URL.Path {
					case "/a":
						_, _ = io.WriteString(w, header+a)
					case "/b":
						if incomplete.Load() {
							http.Error(w, "unavailable", http.StatusServiceUnavailable)
							return
						}
						_, _ = io.WriteString(w, header+b)
					default:
						_, _ = fmt.Fprintf(w, "proxy-providers:\n  a: {type: http, url: http://%s/a}\n  b: {type: http, url: http://%s/b}\n", r.Host, r.Host)
					}
				case "unsupported YAML entry":
					second := b
					if incomplete.Load() {
						second = strings.Replace(b, "type: trojan", "type: unsupported", 1)
					}
					_, _ = io.WriteString(w, header+a+second)
				case "invalid URI":
					second := "trojan://fixture@b.example.com:443#B"
					if incomplete.Load() {
						second = "unsupported://bad"
					}
					_, _ = io.WriteString(w, "trojan://fixture@a.example.com:443#A\n"+second)
				}
			}))
			t.Cleanup(upstream.Close)
			airport := newAirportSyncTestAirport(t, "partial", upstream.URL)
			mustLoadAirportSyncTest(t, airport, nil)
			before := fetchAirportSyncTestRows(t, airport.ID)
			if len(before) != 2 {
				t.Fatalf("initial import has %d nodes", len(before))
			}
			incomplete.Store(true)
			reporter := &airportSyncTestReporter{}
			mustLoadAirportSyncTest(t, airport, reporter)
			after := fetchAirportSyncTestRows(t, airport.ID)
			if len(after) != 2 || before[1].ID != after[1].ID || !strings.Contains(reporter.message, "不完整") {
				t.Fatalf("incomplete refresh affected previous nodes: count=%d message=%q", len(after), reporter.message)
			}
		})
	}
}

func TestAirportSyncPreservesNumberedInformationNodesAndPrunesRemovedNames(t *testing.T) {
	setupSubscriptionCountryBackfillTestDB(t)
	airport := newAirportSyncTestAirport(t, "information", "https://example.test/information")
	first := protocol.Proxy{Type: "trojan", Name: "A", Server: "info.example.com", Port: 443, Password: "fixture"}
	other := first
	other.Name = "B"
	refresh := func(proxies ...protocol.Proxy) []int {
		t.Helper()
		ids, err := scheduleClashToNodeLinks(context.Background(), airport.ID, proxies, airport.Name, nil, nil, false)
		if err != nil {
			t.Fatal(err)
		}
		return ids
	}
	refresh(first, first, other)
	before := fetchAirportSyncTestRows(t, airport.ID)
	if len(before) != 3 || before[0].LinkName != "A-1" || before[1].LinkName != "A-2" {
		t.Fatalf("initial numbered names were not retained: %v", before)
	}
	if ids := refresh(first, first, other); len(ids) != 0 {
		t.Fatalf("identical refresh reported %v changes", ids)
	}
	refresh(first, first)
	after := fetchAirportSyncTestRows(t, airport.ID)
	if len(after) != 2 || after[0].ID != before[0].ID || after[1].ID != before[1].ID {
		t.Fatal("refresh removed numbered nodes still present upstream")
	}
}

func TestAirportSyncRetainsCrossAirportDedupAndGlobalLinkUniqueness(t *testing.T) {
	setupSubscriptionCountryBackfillTestDB(t)
	first := newAirportSyncTestAirport(t, "first", "https://example.test/first")
	second := newAirportSyncTestAirport(t, "second", "https://example.test/second")
	base := protocol.Proxy{Type: "trojan", Name: "original", Server: "shared.example.com", Port: 443, Password: "fixture"}
	renamed, unique := base, base
	renamed.Name = "different name"
	unique.Name, unique.Server = "unique", "unique.example.com"
	refresh := func(airport *models.Airport, proxies ...protocol.Proxy) []int {
		t.Helper()
		ids, err := scheduleClashToNodeLinks(context.Background(), airport.ID, proxies, airport.Name, nil, nil, false)
		if err != nil {
			t.Fatal(err)
		}
		return ids
	}
	refresh(first, base)
	if ids := refresh(second, renamed); len(ids) != 0 {
		t.Fatal("equivalent content was added with cross-airport deduplication enabled")
	}
	if err := models.SetSetting("cross_airport_dedup_enabled", "false"); err != nil {
		t.Fatal(err)
	}
	ids := refresh(second, base, unique)
	nodes := fetchAirportSyncTestRows(t, second.ID)
	if len(ids) != 1 || len(nodes) != 1 || nodes[0].ID != ids[0] || nodes[0].LinkName != "unique" {
		t.Fatal("a skipped link conflict affected the independent insertion")
	}
	if cached, ok := models.GetNodeByID(ids[0]); !ok || cached.Link != nodes[0].Link {
		t.Fatal("the independent insertion has the wrong cached identity")
	}
	refresh(second, renamed, unique)
	if nodes := fetchAirportSyncTestRows(t, second.ID); len(nodes) != 2 {
		t.Fatal("same configuration with a distinct link should be allowed when cross-airport deduplication is disabled")
	}
}

func TestAirportSyncReconcilesLegacyNodeRenamedUpstream(t *testing.T) {
	setupSubscriptionCountryBackfillTestDB(t)
	airport := newAirportSyncTestAirport(t, "rename", "https://example.test/rename")
	proxy := protocol.Proxy{Type: "trojan", Name: "old name", Server: "node.example.com", Port: 443, Password: "fixture"}
	if _, err := scheduleClashToNodeLinks(context.Background(), airport.ID, []protocol.Proxy{proxy}, airport.Name, nil, nil, false); err != nil {
		t.Fatal(err)
	}
	old := fetchAirportSyncTestRows(t, airport.ID)[0]
	if err := database.DB.Model(&models.Node{}).Where("id = ?", old.ID).UpdateColumn("content_hash", old.LinkHash).Error; err != nil {
		t.Fatal(err)
	}
	proxy.Name = "new name"
	ids, err := scheduleClashToNodeLinks(context.Background(), airport.ID, []protocol.Proxy{proxy}, airport.Name, nil, nil, false)
	if err != nil || !slices.Equal(ids, []int{old.ID}) {
		t.Fatalf("legacy rename yielded changes %v, error %v, expected %d", ids, err, old.ID)
	}
	rows := fetchAirportSyncTestRows(t, airport.ID)
	if len(rows) != 1 || rows[0].ID != old.ID || rows[0].LinkName != proxy.Name || rows[0].ContentHash != old.ContentHash {
		t.Fatal("renamed legacy row was not updated in place")
	}
}
