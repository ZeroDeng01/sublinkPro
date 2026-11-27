package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

func setupPreviewAPITestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db := testutil.OpenMemoryDB(t, "preview_api_test")
	if err := db.AutoMigrate(&models.Airport{}, &models.Node{}, &models.NodeCheckProfile{}, &models.NodeCheckJob{}); err != nil {
		t.Fatalf("auto migrate preview api tables: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = true
	if err := models.InitAirportCache(); err != nil {
		t.Fatalf("init airport cache: %v", err)
	}
	if err := models.InitNodeCache(); err != nil {
		t.Fatalf("init node cache: %v", err)
	}

	if err := models.InitNodeCheckProfileCache(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		if oldDB != nil {
			_ = models.InitAirportCache()
			_ = models.InitNodeCache()
			_ = models.InitNodeCheckProfileCache()
		}
		testutil.CloseDB(t, db)
	})
}

func TestSubscriptionServiceMarksRejectInvalidRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const invalid = `$Name$ServiceMarks({"services":[]})`
	for _, tt := range []struct {
		name, body, contentType string
		handler                 gin.HandlerFunc
	}{
		{"add", url.Values{"NodeNameRule": {invalid}}.Encode(), "application/x-www-form-urlencoded", SubAdd},
		{"update", url.Values{"NodeNameRule": {invalid}}.Encode(), "application/x-www-form-urlencoded", SubUpdate},
		{"preview", `{"NodeNameRule":` + strconv.Quote(invalid) + `}`, "application/json", PreviewSubscriptionNodes},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)
			ctx.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(tt.body))
			ctx.Request.Header.Set("Content-Type", tt.contentType)
			tt.handler(ctx)
			var payload utils.Response
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Code != 400 || payload.I18nKey != "subscriptions.rename.serviceMarks.errors.services" || payload.I18nParams["block"] != float64(1) {
				t.Fatalf("unexpected response: %s", response.Body.String())
			}
		})
	}
}

func TestPreviewSubscriptionResponsesIncludeRenderedNames(t *testing.T) {
	setupClientsAPITestDB(t)
	gin.SetMode(gin.TestMode)
	const rule = `$Name$ServiceMarks({"prefix":" · ","services":[{"provider":"openai","label":"GPT","acceptedStatuses":["partial"],"region":true}]})`
	node := models.Node{
		Name: "preview-mark", LinkName: "preview-mark", Link: "ss://preview-mark", Source: "manual",
		UnlockSummary: `{"providers":[{"provider":"openai","status":"partial","region":"JP"}]}`,
	}
	if err := node.Add(); err != nil {
		t.Fatal(err)
	}
	sub := models.Subcription{Name: "preview-service-marks", NodeNameRule: rule, Nodes: []models.Node{node}}
	if err := sub.Add(); err != nil {
		t.Fatal(err)
	}
	if err := sub.AddNode(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		request PreviewRequest
	}{
		{"form", PreviewRequest{NodeIDs: []int{node.ID}, NodeNameRule: rule}},
		{"saved", PreviewRequest{SubscriptionID: sub.ID}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(string(body)))
			ctx.Request.Header.Set("Content-Type", "application/json")
			PreviewSubscriptionNodes(ctx)
			var response struct {
				Code int
				Data struct{ Nodes []map[string]json.RawMessage }
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil || response.Code != http.StatusOK || len(response.Data.Nodes) != 1 {
				t.Fatalf("unexpected preview response: %s (%v)", recorder.Body.String(), err)
			}
			fields := response.Data.Nodes[0]
			if string(fields["PreviewName"]) != `"preview-mark · GPT-JP"` || string(fields["EffectiveName"]) != `"preview-mark"` || string(fields["OriginalName"]) != `"preview-mark"` {
				t.Fatalf("unexpected preview names: %s", recorder.Body.String())
			}
			for _, field := range []string{"PreviewLink", "CountryFlag", "Protocol"} {
				if _, exists := fields[field]; !exists {
					t.Errorf("response is missing %s", field)
				}
			}
		})
	}
}

func TestPreviewFormSubscriptionIncludesAirportIDs(t *testing.T) {
	setupPreviewAPITestDB(t)

	airport := models.Airport{
		Name:     "preview-airport",
		URL:      "https://example.com/preview-airport",
		CronExpr: "0 0 * * *",
		Enabled:  true,
	}
	if err := airport.Add(); err != nil {
		t.Fatalf("add airport: %v", err)
	}

	first := models.Node{Name: "preview-airport-b", LinkName: "preview-airport-b", Link: "ss://preview-airport-b", Protocol: "ss", Source: airport.Name, SourceID: airport.ID, SourceSort: 2}
	second := models.Node{Name: "preview-airport-a", LinkName: "preview-airport-a", Link: "ss://preview-airport-a", Protocol: "ss", Source: airport.Name, SourceID: airport.ID, SourceSort: 1}
	if err := first.Add(); err != nil {
		t.Fatalf("add first airport node: %v", err)
	}
	if err := second.Add(); err != nil {
		t.Fatalf("add second airport node: %v", err)
	}

	result, err := previewFormSubscription(PreviewRequest{AirportIDs: []int{airport.ID}})
	if err != nil {
		t.Fatalf("preview form subscription: %v", err)
	}

	if got, want := len(result.Nodes), 2; got != want {
		t.Fatalf("preview node count = %d, want %d", got, want)
	}
	if got, want := result.Nodes[0].PreviewName, "preview-airport-a"; got != want {
		t.Fatalf("first preview node = %q, want %q", got, want)
	}
	if got, want := result.Nodes[1].PreviewName, "preview-airport-b"; got != want {
		t.Fatalf("second preview node = %q, want %q", got, want)
	}
}
