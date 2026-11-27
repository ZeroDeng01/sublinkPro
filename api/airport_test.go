package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"testing"
	"testing/synctest"
	"time"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"

	"github.com/gin-gonic/gin"
)

func setupAirportAPITestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db := testutil.OpenMemoryDB(t, "airport_api_test")
	if err := db.AutoMigrate(&models.Airport{}, &models.Node{}); err != nil {
		t.Fatalf("auto migrate airports and nodes: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	if err := models.InitAirportCache(); err != nil {
		t.Fatalf("init airport cache: %v", err)
	}

	t.Cleanup(func() {
		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		if oldDB != nil {
			_ = models.InitAirportCache()
		}
		testutil.CloseDB(t, db)
	})
}

func performAirportJSONRequest(t *testing.T, handler gin.HandlerFunc, method, path string, body any, params ...gin.Param) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var requestBody []byte
	var err error
	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
	}

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequestWithContext(context.Background(), method, path, bytes.NewReader(requestBody))
	ginContext.Request.Header.Set("Content-Type", "application/json")
	ginContext.Params = params

	handler(ginContext)
	return recorder
}

func decodeAirportAPIResponse(t *testing.T, recorder *httptest.ResponseRecorder) apiJSONResponse {
	t.Helper()
	return decodeAPIResponse(t, recorder)
}

func TestAirportAddPersistsUpdateAfterDetectSettings(t *testing.T) {
	setupAirportAPITestDB(t)

	addBody := map[string]any{
		"name":                         "机场A",
		"url":                          "https://example.com/subscription",
		"cronExpr":                     "0 */12 * * *",
		"enabled":                      false,
		"group":                        "默认组",
		"downloadWithProxy":            false,
		"proxyLink":                    "",
		"userAgent":                    "",
		"requestHeaders":               []map[string]string{},
		"fetchUsageInfo":               false,
		"skipTLSVerify":                false,
		"updateAfterDetect":            true,
		"updateAfterDetectProfileId":   7,
		"updateAfterDetectChangedOnly": true,
		"remark":                       "",
		"logo":                         "",
		"nodeNameWhitelist":            "",
		"nodeNameBlacklist":            "",
		"protocolWhitelist":            "",
		"protocolBlacklist":            "",
		"nodeNamePreprocess":           "",
		"deduplicationRule":            "",
		"nodeNameUniquify":             false,
		"nodeNamePrefix":               "",
		"nodeNameIntraUniquify":        false,
	}

	addRecorder := performAirportJSONRequest(t, AirportAdd, http.MethodPost, "/api/v1/airports", addBody)
	if addRecorder.Code != http.StatusOK {
		t.Fatalf("add airport status = %d, body = %s", addRecorder.Code, addRecorder.Body.String())
	}
	addResponse := decodeAirportAPIResponse(t, addRecorder)
	if addResponse.Code != 200 {
		t.Fatalf("add airport response code = %d, msg = %s", addResponse.Code, addResponse.Msg)
	}

	var stored models.Airport
	if err := database.DB.Where("name = ?", "机场A").First(&stored).Error; err != nil {
		t.Fatalf("load added airport: %v", err)
	}
	if !stored.UpdateAfterDetect {
		t.Fatal("expected update_after_detect to be stored as true after add")
	}
	if stored.UpdateAfterDetectProfileID != 7 {
		t.Fatalf("expected update_after_detect_profile_id = 7 after add, got %d", stored.UpdateAfterDetectProfileID)
	}
	if !stored.UpdateAfterDetectChangedOnly {
		t.Fatal("expected update_after_detect_changed_only to be stored as true after add")
	}
}

func TestAirportPriorityListAndUpdate(t *testing.T) {
	setupAirportAPITestDB(t)
	requestOK := func(handler gin.HandlerFunc, method, path string, body any, params ...gin.Param) json.RawMessage {
		t.Helper()
		recorder := performAirportJSONRequest(t, handler, method, path, body, params...)
		response := decodeAirportAPIResponse(t, recorder)
		if recorder.Code != http.StatusOK || response.Code != 200 {
			t.Fatalf("%s %s failed: HTTP %d, response %d: %s", method, path, recorder.Code, response.Code, response.Msg)
		}
		return response.Data
	}
	updateAndWait := func(c *gin.Context) {
		// Let the scheduler finish its writes before the test database is closed.
		synctest.Test(t, func(*testing.T) { AirportUpdate(c) })
	}

	basePath := "/api/v1/airports"
	fixtures := []struct {
		name     string
		group    string
		priority *int32
	}{
		{name: "default", group: "test"},
		{name: "negative", group: "test", priority: new(int32(-1))},
		{name: "first", group: "test", priority: new(int32(10))},
		{name: "second", group: "test", priority: new(int32(10))},
		{name: "other", group: "other", priority: new(int32(20))},
	}
	for _, airport := range fixtures {
		body := map[string]any{
			"name": airport.name, "url": "https://example.com/" + airport.name,
			"cronExpr": "0 */12 * * *", "group": airport.group,
		}
		if airport.priority != nil {
			body["priority"] = *airport.priority
		}
		requestOK(AirportAdd, http.MethodPost, basePath, body)
	}

	sameName := func(airport models.Airport, name string) bool { return airport.Name == name }
	wantAll := []string{"other", "first", "second", "default", "negative"}
	all, err := new(models.Airport).List()
	if err != nil || !slices.EqualFunc(all, wantAll, sameName) {
		t.Fatalf("shared list has unexpected order or error: %v", err)
	}
	page, total, err := new(models.Airport).ListPaginated(2, 2)
	if err != nil || total != 5 || !slices.EqualFunc(page, wantAll[2:4], sameName) {
		t.Fatalf("shared page has unexpected order or total %d: %v", total, err)
	}

	var unpaginated []models.Airport
	if err := json.Unmarshal(requestOK(AirportList, http.MethodGet, basePath, nil), &unpaginated); err != nil {
		t.Fatal(err)
	}
	if !slices.EqualFunc(unpaginated, wantAll, sameName) {
		t.Fatalf("unpaginated response is out of order, want %v", wantAll)
	}
	for _, want := range []struct {
		query string
		names []string
	}{
		{query: "?group=test&page=1&pageSize=2", names: []string{"first", "second"}},
		{query: "?group=test&page=2&pageSize=2", names: []string{"default", "negative"}},
	} {
		var result struct {
			Items []models.Airport `json:"items"`
			Total int              `json:"total"`
		}
		if err := json.Unmarshal(requestOK(AirportList, http.MethodGet, basePath+want.query, nil), &result); err != nil {
			t.Fatal(err)
		}
		if result.Total != 4 || !slices.EqualFunc(result.Items, want.names, sameName) {
			t.Fatalf("page %s has total %d or unexpected order, want total 4 and %v", want.query, result.Total, want.names)
		}
	}

	defaultAirport := unpaginated[3]
	id := strconv.Itoa(defaultAirport.ID)
	path := basePath + "/" + id
	param := gin.Param{Key: "id", Value: id}
	assertSavedAirport := func(want int32) {
		t.Helper()
		var stored models.Airport
		if err := database.DB.First(&stored, defaultAirport.ID).Error; err != nil {
			t.Fatal(err)
		}
		if stored.Priority != want {
			t.Fatalf("stored priority = %d, want %d", stored.Priority, want)
		}
		if stored.NextRunTime != nil {
			t.Fatalf("disabled airport next run = %v, want nil", stored.NextRunTime)
		}
		var detail struct {
			Priority *int32 `json:"priority"`
		}
		if err := json.Unmarshal(requestOK(AirportGet, http.MethodGet, path, nil, param), &detail); err != nil {
			t.Fatal(err)
		}
		if detail.Priority == nil || *detail.Priority != want {
			t.Fatalf("detail response should include priority %d", want)
		}
	}
	assertSavedAirport(0)
	if err := defaultAirport.UpdateRunTime(nil, new(time.Now().Add(time.Hour))); err != nil {
		t.Fatalf("seed previous next run time: %v", err)
	}
	update := map[string]any{
		"name": defaultAirport.Name, "url": defaultAirport.URL,
		"cronExpr": defaultAirport.CronExpr, "group": defaultAirport.Group,
		"priority": 15,
	}
	requestOK(updateAndWait, http.MethodPut, path, update, param)
	assertSavedAirport(15)
	delete(update, "priority")
	requestOK(updateAndWait, http.MethodPut, path, update, param)
	assertSavedAirport(15)
	update["priority"] = 0
	requestOK(updateAndWait, http.MethodPut, path, update, param)
	assertSavedAirport(0)
}
