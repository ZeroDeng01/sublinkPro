package api

import (
	"testing"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
)

func setupPreviewAPITestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db := testutil.OpenMemoryDB(t, "preview_api_test")
	if err := db.AutoMigrate(&models.Airport{}, &models.Node{}); err != nil {
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

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		if oldDB != nil {
			_ = models.InitAirportCache()
			_ = models.InitNodeCache()
		}
		testutil.CloseDB(t, db)
	})
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
