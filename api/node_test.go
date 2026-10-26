package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
	"sublink/node/protocol"

	"github.com/gin-gonic/gin"
)

// setupNodeUpdateAPITestDB prepares an isolated database and caches for node update handler tests.
func setupNodeUpdateAPITestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized
	db := testutil.OpenMemoryDB(t, "node_update_api_test")
	if err := db.AutoMigrate(&models.Node{}, &models.SystemSetting{}); err != nil {
		t.Fatalf("auto migrate node update tables: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	if err := models.InitNodeCache(); err != nil {
		t.Fatalf("init node cache: %v", err)
	}
	if err := models.InitSettingCache(); err != nil {
		t.Fatalf("init setting cache: %v", err)
	}

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		if oldDB != nil {
			_ = models.InitNodeCache()
			_ = models.InitSettingCache()
		}
		testutil.CloseDB(t, db)
	})
}

// TestNodeUpdatePreservesClashExtraWhenLinkUnchanged verifies that metadata-only
// edits keep retained Clash fields, while replacing the link clears stale fields.
func TestNodeUpdatePreservesClashExtraWhenLinkUnchanged(t *testing.T) {
	for _, tt := range []struct {
		name        string
		updatedLink func(original string) string
		wantExtra   string
	}{
		{
			name:        "metadata-only edit",
			updatedLink: func(original string) string { return original },
			wantExtra:   "smux:\n  enabled: true\n",
		},
		{
			name: "link replacement",
			updatedLink: func(string) string {
				return protocol.EncodeSSURL(protocol.Ss{
					Name:   "replacement-node",
					Server: "replacement.example.com",
					Port:   8443,
					Param: protocol.Param{
						Cipher:   "aes-256-gcm",
						Password: "password",
					},
				})
			},
			wantExtra: "",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			setupNodeUpdateAPITestDB(t)
			originalLink := protocol.EncodeSSURL(protocol.Ss{
				Name:   "source-node",
				Server: "edge.example.com",
				Port:   443,
				Param: protocol.Param{
					Cipher:   "aes-256-gcm",
					Password: "password",
				},
			})
			node := models.Node{
				Name:       "source-node",
				LinkName:   "source-node",
				NameMode:   models.NodeNameModeLink,
				Link:       originalLink,
				ClashExtra: "smux:\n  enabled: true\n",
				Protocol:   "ss",
			}
			if err := node.Add(); err != nil {
				t.Fatalf("add test node: %v", err)
			}

			updatedLink := tt.updatedLink(originalLink)
			form := url.Values{
				"oldname":         {node.Name},
				"oldlink":         {originalLink},
				"link":            {updatedLink},
				"name":            {"renamed-node"},
				"nameMode":        {models.NodeNameModeRemark},
				"dialerProxyName": {""},
				"group":           {"updated-group"},
			}
			recorder := httptest.NewRecorder()
			ginContext, _ := gin.CreateTestContext(recorder)
			ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/nodes/update", strings.NewReader(form.Encode()))
			ginContext.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			NodeUpdadte(ginContext)
			if recorder.Code != http.StatusOK {
				t.Fatalf("update status = %d, body = %s", recorder.Code, recorder.Body.String())
			}

			var stored models.Node
			if err := database.DB.First(&stored, node.ID).Error; err != nil {
				t.Fatalf("reload updated node: %v", err)
			}
			if stored.ClashExtra != tt.wantExtra {
				t.Fatalf("ClashExtra = %q, want %q", stored.ClashExtra, tt.wantExtra)
			}
		})
	}
}

// TestNodeCountryStatsFilteredAppliesQueryFilters verifies that the filtered country
// aggregation endpoint reuses the same query filters as the node list.
func TestNodeCountryStatsFilteredAppliesQueryFilters(t *testing.T) {
	setupNodeUpdateAPITestDB(t)

	for _, item := range []struct {
		name    string
		link    string
		country string
		group   string
	}{
		{name: "us-1", link: "ss://us-1", country: "US", group: "Asia"},
		{name: "us-2", link: "ss://us-2", country: "US", group: "Asia"},
		{name: "de-1", link: "ss://de-1", country: "DE", group: "Asia"},
		{name: "sg-1", link: "ss://sg-1", country: "SG", group: "Europe"},
	} {
		n := models.Node{
			Name:        item.name,
			LinkName:    item.name,
			NameMode:    models.NodeNameModeLink,
			Link:        item.link,
			Protocol:    "ss",
			Group:       item.group,
			LinkCountry: item.country,
		}
		if err := n.Add(); err != nil {
			t.Fatalf("add node %s: %v", item.name, err)
		}
	}

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/nodes/country-stats?group=Asia", nil)

	NodeCountryStatsFiltered(ginContext)

	if recorder.Code != http.StatusOK {
		t.Fatalf("stats status = %d, body = %s", recorder.Code, recorder.Body.String())
	}

	var body struct {
		Code int `json:"code"`
		Data struct {
			Items []models.NodeCountryStat `json:"items"`
			Total int                      `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode stats response: %v, body = %s", err, recorder.Body.String())
	}
	if body.Code != 200 {
		t.Fatalf("stats code = %d, body = %s", body.Code, recorder.Body.String())
	}
	if body.Data.Total != 3 {
		t.Fatalf("stats total = %d, want 3", body.Data.Total)
	}
	want := []models.NodeCountryStat{{Country: "US", Count: 2}, {Country: "DE", Count: 1}}
	if len(body.Data.Items) != len(want) {
		t.Fatalf("stats items = %+v, want %+v", body.Data.Items, want)
	}
	for i := range want {
		if body.Data.Items[i] != want[i] {
			t.Fatalf("stats items[%d] = %+v, want %+v", i, body.Data.Items[i], want[i])
		}
	}
}
