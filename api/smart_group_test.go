package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"sublink/database"
	"sublink/models"
	"sublink/services/scheduler"
)

func TestSmartGroupPreviewAndMembers(t *testing.T) {
	setupPreviewAPITestDB(t)
	group := models.SmartGroup{Name: "preview-smart", Countries: "GB,DE", MaxDelay: 400, MaxAgeHours: 72}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	for _, item := range []struct{ name, source, country string }{
		{"smart-gb", "source-one", "GB"}, {"smart-de", "source-two", "DE"}, {"smart-us", "source-three", "US"},
	} {
		node := models.Node{Name: item.name, LinkName: item.name, Link: "ss://" + item.name, Protocol: "ss", Group: item.source, LinkCountry: item.country,
			DelayStatus: "success", SpeedStatus: "success", DelayTime: 100, Speed: 3, LatencyCheckAt: now, SpeedCheckAt: now}
		if err := node.Add(); err != nil {
			t.Fatal(err)
		}
	}
	result, err := previewFormSubscription(PreviewRequest{SmartGroupIDs: []int{group.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Nodes) != 2 {
		t.Fatalf("preview nodes = %d, want 2", len(result.Nodes))
	}
	router := gin.New()
	router.GET("/:id/members", SmartGroupMembers)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/"+strconv.Itoa(group.ID)+"/members", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("members status = %d: %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Data struct {
			Count          int `json:"count"`
			CandidateCount int `json:"candidateCount"`
			Nodes          []struct {
				Group string `json:"group"`
			} `json:"nodes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Count != 2 || response.Data.CandidateCount != 2 || response.Data.Nodes[0].Group != "source-one" || response.Data.Nodes[1].Group != "source-two" {
		t.Fatalf("members response = %+v", response.Data)
	}
}

func TestDeleteSmartGroupRejectsSubscribedGroup(t *testing.T) {
	setupPreviewAPITestDB(t)
	group := models.SmartGroup{Name: "used", Countries: "GB"}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	sub := models.Subcription{Name: "reference", SmartGroupIDs: strconv.Itoa(group.ID)}
	if err := database.DB.Create(&sub).Error; err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.DELETE("/:id", DeleteSmartGroup)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/"+strconv.Itoa(group.ID), nil))
	if recorder.Code != http.StatusConflict {
		t.Fatalf("delete status = %d: %s", recorder.Code, recorder.Body.String())
	}
	if _, err := models.GetSmartGroup(group.ID); err != nil {
		t.Fatalf("referenced group removed: %v", err)
	}
}

func TestCreateAndUpdateSmartGroup(t *testing.T) {
	setupPreviewAPITestDB(t)
	router := gin.New()
	router.POST("/", CreateSmartGroup)
	router.PUT("/:id", UpdateSmartGroup)
	create := httptest.NewRecorder()
	router.ServeHTTP(create, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", bytes.NewBufferString(`{"name":"Europe","countries":"gb,FR","keyword":"  fast ","sourceGroups":[" airport-a ","AIRPORT-A"],"maxDelay":500}`)))
	if create.Code != http.StatusOK {
		t.Fatalf("create status = %d: %s", create.Code, create.Body.String())
	}
	var response struct {
		Data models.SmartGroup `json:"data"`
	}
	if err := json.Unmarshal(create.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.Countries != "GB,FR" || response.Data.MaxAgeHours != 72 || response.Data.Keyword != "fast" || len(response.Data.SourceGroups) != 1 || response.Data.SourceGroups[0] != "airport-a" {
		t.Fatalf("created group = %+v", response.Data)
	}
	update := httptest.NewRecorder()
	router.ServeHTTP(update, httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/"+strconv.Itoa(response.Data.ID), bytes.NewBufferString(`{"name":"Europe","countries":"DE","keyword":"","sourceGroups":[],"maxDelay":0,"minSpeed":0,"maxAgeHours":0}`)))
	if update.Code != http.StatusOK {
		t.Fatalf("update status = %d: %s", update.Code, update.Body.String())
	}
	saved, err := models.GetSmartGroup(response.Data.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Countries != "DE" || saved.MaxAgeHours != 0 || saved.MaxDelay != 0 || saved.Keyword != "" || len(saved.SourceGroups) != 0 {
		t.Fatalf("updated group = %+v", saved)
	}
}

func TestBatchFillCountryUpdatesSmartGroupCandidates(t *testing.T) {
	oldDB := database.DB
	t.Cleanup(func() {
		if oldDB != nil {
			_ = models.InitCountryRuleCache()
		}
	})
	setupPreviewAPITestDB(t)
	if err := database.DB.AutoMigrate(&models.CountryRule{}); err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Create(&models.CountryRule{CountryCode: "PH", CountryName: "Philippines", Pattern: "Philippines", Enabled: true}).Error; err != nil {
		t.Fatal(err)
	}
	if err := models.InitCountryRuleCache(); err != nil {
		t.Fatal(err)
	}
	node := models.Node{Name: "Philippines 01", LinkName: "Philippines 01", Link: "ss://ph-01", Protocol: "ss"}
	if err := node.Add(); err != nil {
		t.Fatal(err)
	}
	group := models.SmartGroup{Name: "PH", Countries: "PH"}
	before, err := group.CandidateIDs()
	if err != nil || len(before) != 1 || before[0] != node.ID {
		t.Fatalf("name-inferred candidates before fill = %v, %v", before, err)
	}
	if country, source := models.SmartGroupCountryForDisplay(node); country != "PH" || source != "name" {
		t.Fatalf("name fallback = %q from %q", country, source)
	}
	router := gin.New()
	router.POST("/fill", NodeBatchFillCountry)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/fill", bytes.NewBufferString(`{"onlyEmpty":true}`)))
	if recorder.Code != http.StatusOK {
		t.Fatalf("fill country status = %d: %s", recorder.Code, recorder.Body.String())
	}
	after, err := group.CandidateIDs()
	if err != nil || len(after) != 1 || after[0] != node.ID {
		t.Fatalf("candidates after fill = %v, %v", after, err)
	}
	var stored models.Node
	if err := database.DB.First(&stored, node.ID).Error; err != nil || stored.LinkCountry != "PH" {
		t.Fatalf("persisted country after fill = %q, %v", stored.LinkCountry, err)
	}
	if country, source := models.SmartGroupCountryForDisplay(stored); country != "PH" || source != "stored" {
		t.Fatalf("stored country = %q from %q", country, source)
	}
}

func TestSmartGroupMembersListsPagedCandidatesWithReasons(t *testing.T) {
	setupPreviewAPITestDB(t)
	group := models.SmartGroup{Name: "Europe", Countries: "GB", MaxAgeHours: 72}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	airport := models.Airport{Name: "Original Airport", URL: "https://example.test/airport", CronExpr: "0 0 * * *", Enabled: true}
	if err := airport.Add(); err != nil {
		t.Fatal(err)
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	for _, item := range []struct {
		name, status, country string
		delay, sourceID       int
	}{
		{"first-untested", "untested", "GB", 0, airport.ID},
		{"second-healthy", "success", "GB", 90, 0},
		{"third-other-country", "success", "DE", 70, 0},
	} {
		node := models.Node{Name: item.name, LinkName: item.name, Link: "ss://" + item.name, Protocol: "ss", LinkCountry: item.country, SourceID: item.sourceID,
			DelayStatus: item.status, DelayTime: item.delay, LatencyCheckAt: now}
		if err := node.Add(); err != nil {
			t.Fatal(err)
		}
	}
	router := gin.New()
	router.GET("/:id/members", SmartGroupMembers)
	for _, tt := range []struct {
		query, name, reason, airportName string
		airportID                        int
		included                         bool
	}{
		{"?page=1&pageSize=1", "first-untested", "delayUnusable", "Original Airport", airport.ID, false},
		{"?page=2&pageSize=1", "second-healthy", "", "", 0, true},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/"+strconv.Itoa(group.ID)+"/members"+tt.query, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("candidate page %s status %d: %s", tt.query, response.Code, response.Body.String())
		}
		var data struct {
			Data struct {
				Count, CandidateCount, Page int
				CandidateNodes              []struct {
					Name, Reason, MatchSource, AirportName string
					AirportID                              int  `json:"airportId"`
					Included                               bool `json:"includedBySmartGroup"`
				} `json:"candidateNodes"`
				StatusCounts struct {
					DelayUnusable int `json:"delayUnusable"`
				} `json:"statusCounts"`
			} `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil {
			t.Fatal(err)
		}
		if data.Data.Count != 1 || data.Data.CandidateCount != 2 || data.Data.StatusCounts.DelayUnusable != 1 ||
			len(data.Data.CandidateNodes) != 1 || data.Data.CandidateNodes[0].Name != tt.name || data.Data.CandidateNodes[0].Reason != tt.reason || data.Data.CandidateNodes[0].MatchSource != "landingCountry" ||
			data.Data.CandidateNodes[0].AirportName != tt.airportName || data.Data.CandidateNodes[0].Included != tt.included {
			t.Fatalf("unexpected candidate page %s: %+v", tt.query, data.Data)
		}
		if data.Data.CandidateNodes[0].AirportID != tt.airportID {
			t.Fatalf("wrong airport ID for %s: %+v", tt.query, data.Data.CandidateNodes[0])
		}
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/"+strconv.Itoa(group.ID)+"/members?pageSize=101", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid page size status = %d", response.Code)
	}
}

func TestSmartGroupCheckRejectsNodeOutsideCandidates(t *testing.T) {
	setupPreviewAPITestDB(t)
	group := models.SmartGroup{Name: "GB", Countries: "GB"}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	var outsideNodeID int
	for _, country := range []string{"GB", "US"} {
		node := models.Node{Name: country, LinkName: country, Link: "ss://" + country, Protocol: "ss", LinkCountry: country}
		if err := node.Add(); err != nil {
			t.Fatal(err)
		}
		if country == "US" {
			outsideNodeID = node.ID
		}
	}
	if err := database.DB.AutoMigrate(&models.NodeCheckProfile{}); err != nil {
		t.Fatal(err)
	}
	profile := models.NodeCheckProfile{Name: "test-profile", Mode: "tcp"}
	if err := profile.Add(); err != nil {
		t.Fatal(err)
	}
	// A valid profile must not allow retesting a node outside this group's candidates.
	router := gin.New()
	router.POST("/:id/check", CheckSmartGroupCandidates)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/"+strconv.Itoa(group.ID)+"/check", bytes.NewBufferString(`{"profileId":`+strconv.Itoa(profile.ID)+`,"nodeId":`+strconv.Itoa(outsideNodeID)+`}`)))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("non-candidate node status = %d: %s", response.Code, response.Body.String())
	}
}

func TestSmartGroupAutomationAPIAndReadonlyState(t *testing.T) {
	setupPreviewAPITestDB(t)
	profile := models.NodeCheckProfile{Name: "automation", Mode: "mihomo"}
	if err := profile.Add(); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.POST("/", CreateSmartGroup)
	router.PUT("/:id", UpdateSmartGroup)
	router.DELETE("/:id", DeleteSmartGroup)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRecorder()
		router.ServeHTTP(r, httptest.NewRequestWithContext(context.Background(), method, path, bytes.NewBufferString(body)))
		return r
	}
	body := `{"name":"scheduled","countries":"GB","minSpeed":1,"checkProfileId":` + strconv.Itoa(profile.ID) + `,"scheduleEnabled":true,"cronExpr":"0 3 * * *","checkAfterAirportUpdate":true,"lastCheckAt":"2000-01-01T00:00:00Z","checkRunning":true}`
	r := request(http.MethodPost, "/", body)
	if r.Code != http.StatusOK {
		t.Fatalf("create: %d %s", r.Code, r.Body.String())
	}
	var response struct {
		Data models.SmartGroup `json:"data"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	group := response.Data
	t.Cleanup(func() { scheduler.GetSchedulerManager().RemoveSmartGroupJob(group.ID) })
	if group.LastCheckAt != nil || group.CheckRunning || group.NextCheckAt == nil || group.CheckTimezone == "" {
		t.Fatalf("runtime fields = %+v", group)
	}
	path := "/" + strconv.Itoa(group.ID)
	r = request(http.MethodPut, path, `{"name":"scheduled","countries":"GB","minSpeed":1,"checkProfileId":999999,"scheduleEnabled":true,"cronExpr":"0 4 * * *"}`)
	if r.Code != http.StatusBadRequest {
		t.Fatalf("invalid update: %d %s", r.Code, r.Body.String())
	}
	unchanged, err := models.GetSmartGroup(group.ID)
	if err != nil || unchanged.CronExpr != "0 3 * * *" || unchanged.CheckProfileID != profile.ID {
		t.Fatalf("invalid update changed config: %+v %v", unchanged, err)
	}
	triggered := time.Now().Add(-time.Hour)
	if err := database.DB.Model(&group).Update("last_check_at", triggered).Error; err != nil {
		t.Fatal(err)
	}
	r = request(http.MethodPut, path, `{"name":"legacy-edit","countries":"DE","minSpeed":1,"lastCheckAt":null}`)
	if r.Code != http.StatusOK {
		t.Fatalf("legacy update: %d %s", r.Code, r.Body.String())
	}
	saved, err := models.GetSmartGroup(group.ID)
	if err != nil || saved.CheckProfileID != profile.ID || !saved.ScheduleEnabled || !saved.CheckAfterAirportUpdate || saved.CronExpr != "0 3 * * *" || saved.LastCheckAt == nil || !saved.LastCheckAt.Equal(triggered) {
		t.Fatalf("legacy settings/state lost: %+v %v", saved, err)
	}
	r = request(http.MethodPut, path, `{"name":"disabled","countries":"DE","checkProfileId":0,"scheduleEnabled":false,"checkAfterAirportUpdate":false,"cronExpr":""}`)
	if r.Code != http.StatusOK {
		t.Fatalf("disable: %d %s", r.Code, r.Body.String())
	}
	saved, err = models.GetSmartGroup(group.ID)
	if err != nil {
		t.Fatal(err)
	}
	info := scheduler.GetSchedulerManager().SmartGroupCheckInfo(saved)
	if info.NextCheckAt != nil || saved.CheckProfileID != 0 || saved.ScheduleEnabled || saved.CheckAfterAirportUpdate {
		t.Fatalf("disable persisted = %+v", info)
	}
	r = request(http.MethodDelete, path, "")
	if r.Code != http.StatusOK {
		t.Fatalf("delete: %d %s", r.Code, r.Body.String())
	}
}

func TestSmartGroupAutomationAPIRejectsInvalidProfilesAndCron(t *testing.T) {
	setupPreviewAPITestDB(t)
	profile := models.NodeCheckProfile{Name: "tcp-only", Mode: "tcp"}
	if err := profile.Add(); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.POST("/", CreateSmartGroup)
	for _, body := range []string{
		`{"name":"missing","countries":"GB","scheduleEnabled":true,"cronExpr":"0 3 * * *"}`,
		`{"name":"unknown","countries":"GB","checkProfileId":999999,"checkAfterAirportUpdate":true}`,
		`{"name":"bad-cron","countries":"GB","checkProfileId":` + strconv.Itoa(profile.ID) + `,"scheduleEnabled":true,"cronExpr":"*/0 * * * *"}`,
		`{"name":"tcp-speed","countries":"GB","minSpeed":1,"checkProfileId":` + strconv.Itoa(profile.ID) + `,"checkAfterAirportUpdate":true}`,
	} {
		r := httptest.NewRecorder()
		router.ServeHTTP(r, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", bytes.NewBufferString(body)))
		if r.Code != http.StatusBadRequest {
			t.Fatalf("invalid create: %d %s (%s)", r.Code, r.Body.String(), body)
		}
	}
	groups, err := models.ListSmartGroups()
	if err != nil || len(groups) != 0 {
		t.Fatalf("invalid group saved: %+v %v", groups, err)
	}
}

func TestProfileSmartGroupReferenceProtection(t *testing.T) {
	setupPreviewAPITestDB(t)
	profile := models.NodeCheckProfile{Name: "referenced", Mode: "mihomo"}
	if err := profile.Add(); err != nil {
		t.Fatal(err)
	}
	group := models.SmartGroup{Name: "dependent", Countries: "GB", MinSpeed: 1, CheckProfileID: profile.ID, CheckAfterAirportUpdate: true}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.PUT("/:id", UpdateNodeCheckProfile)
	router.DELETE("/:id", DeleteNodeCheckProfile)
	path := "/" + strconv.Itoa(profile.ID)
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		r := httptest.NewRecorder()
		router.ServeHTTP(r, httptest.NewRequestWithContext(context.Background(), method, path, bytes.NewBufferString(`{"name":"referenced","mode":"tcp"}`)))
		var response struct {
			Code int `json:"code"`
		}
		if err := json.Unmarshal(r.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.Code == 200 {
			t.Fatalf("profile mutation accepted: %s %s", method, r.Body.String())
		}
		saved, err := models.GetNodeCheckProfileByID(profile.ID)
		if err != nil || saved.Mode != "mihomo" {
			t.Fatalf("profile mutated: %+v %v", saved, err)
		}
	}
	// Even disabled automation keeps a saved default-profile reference protected.
	if err := database.DB.Model(&group).Update("check_after_airport_update", false).Error; err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRecorder()
	router.ServeHTTP(r, httptest.NewRequestWithContext(context.Background(), http.MethodDelete, path, nil))
	if _, err := models.GetNodeCheckProfileByID(profile.ID); err != nil {
		t.Fatal("default profile reference removed")
	}
	if err := database.DB.Model(&group).Update("check_profile_id", 0).Error; err != nil {
		t.Fatal(err)
	}
	r = httptest.NewRecorder()
	router.ServeHTTP(r, httptest.NewRequestWithContext(context.Background(), http.MethodDelete, path, nil))
	if _, err := models.GetNodeCheckProfileByID(profile.ID); err == nil {
		t.Fatalf("unreferenced profile not deleted: %s", r.Body.String())
	}
}
