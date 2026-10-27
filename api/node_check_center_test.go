package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sublink/database"
	"sublink/models"
	"sublink/services/scheduler"
	"testing"
	"time"
)

func checkCenterRequest(t *testing.T, r *gin.Engine, method, path string, body any) map[string]json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), method, path, bytes.NewReader(raw)))
	var response map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("response=%s: %v", recorder.Body.String(), err)
	}
	return response
}

func TestNodeCheckCenterJobsCRUDAndProfileReferenceProtection(t *testing.T) {
	setupPreviewAPITestDB(t)
	profile := models.NodeCheckProfile{Name: "center-plan", Mode: "tcp"}
	if err := profile.Add(); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.POST("/jobs", SaveNodeCheckJob)
	r.PUT("/jobs/:id", SaveNodeCheckJob)
	r.DELETE("/profiles/:id", DeleteNodeCheckProfile)
	response := checkCenterRequest(t, r, http.MethodPost, "/jobs", map[string]any{"name": "center-job", "profileId": profile.ID, "enabled": false, "scope": "all", "trigger": "scheduled", "cronExpr": "30 5 * * *", "lastError": "forged"})
	if string(response["code"]) != "200" {
		t.Fatalf("create=%s", response)
	}
	var job models.NodeCheckJob
	if err := json.Unmarshal(response["data"], &job); err != nil {
		t.Fatal(err)
	}
	if job.ID <= 0 || job.LastError != "" || job.LastRunTime != nil {
		t.Fatalf("untrusted runtime persisted: %+v", job)
	}
	now := time.Now().UTC().Truncate(time.Second)
	if err := database.DB.Model(&job).Updates(map[string]any{"last_run_time": now, "last_error": "keep"}).Error; err != nil {
		t.Fatal(err)
	}
	job.Name = "edited"
	response = checkCenterRequest(t, r, http.MethodPut, "/jobs/"+strconv.Itoa(job.ID), job)
	if string(response["code"]) != "200" {
		t.Fatalf("update=%s", response)
	}
	saved, err := models.GetNodeCheckJob(job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.LastRunTime == nil || !saved.LastRunTime.Equal(now) || saved.LastError != "keep" {
		t.Fatalf("configuration edit clobbered runtime: %+v", saved)
	}
	response = checkCenterRequest(t, r, http.MethodDelete, "/profiles/"+strconv.Itoa(profile.ID), nil)
	if string(response["code"]) == "200" {
		t.Fatal("referenced profile deleted")
	}
	if _, err := models.GetNodeCheckProfileByID(profile.ID); err != nil {
		t.Fatal(err)
	}
	for _, patch := range []map[string]any{{"scope": "filtered"}, {"cronExpr": "*/0 * * * *"}, {"profileId": 999}} {
		payload := map[string]any{"name": "invalid", "profileId": profile.ID, "scope": "all", "trigger": "scheduled", "cronExpr": "30 5 * * *"}
		for key, value := range patch {
			payload[key] = value
		}
		response = checkCenterRequest(t, r, http.MethodPost, "/jobs", payload)
		if string(response["code"]) == "200" {
			t.Fatalf("invalid job accepted: %v", payload)
		}
	}
	var count int64
	if err := database.DB.Model(&models.NodeCheckJob{}).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("invalid jobs persisted: %d,%v", count, err)
	}
}

func TestNodeCheckCenterLegacyScheduleMappedWithoutCopyAndPlanEditPreservesIt(t *testing.T) {
	setupPreviewAPITestDB(t)
	p := models.NodeCheckProfile{Name: "legacy-plan", Mode: "tcp", CronExpr: "30 5 * * *", Groups: "original", Tags: "tag", Enabled: false}
	if err := p.Add(); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.GET("/jobs", ListNodeCheckJobs)
	r.PUT("/profiles/:id", UpdateNodeCheckProfile)
	r.DELETE("/jobs/:id", DeleteNodeCheckJob)
	response := checkCenterRequest(t, r, http.MethodGet, "/jobs", nil)
	var data struct {
		Items []nodeCheckJobView `json:"items"`
	}
	if err := json.Unmarshal(response["data"], &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Items) != 1 || data.Items[0].ID != -p.ID || !data.Items[0].Legacy {
		t.Fatalf("legacy projection=%+v", data)
	}
	var count int64
	if err := database.DB.Model(&models.NodeCheckJob{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("legacy task was copied: %d,%v", count, err)
	}
	response = checkCenterRequest(t, r, http.MethodPut, "/profiles/"+strconv.Itoa(p.ID), map[string]any{"name": p.Name, "mode": "tcp", "detectCountry": true})
	if string(response["code"]) != "200" {
		t.Fatalf("plan update=%s", response)
	}
	saved, err := models.GetNodeCheckProfileByID(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.CronExpr != p.CronExpr || saved.Groups != "original" || saved.Tags != "tag" || !saved.DetectCountry {
		t.Fatalf("legacy schedule lost: %+v", saved)
	}
	response = checkCenterRequest(t, r, http.MethodDelete, "/jobs/"+strconv.Itoa(-p.ID), nil)
	if string(response["code"]) != "200" {
		t.Fatalf("legacy delete=%s", response)
	}
	saved, err = models.GetNodeCheckProfileByID(p.ID)
	if err != nil || saved.CronExpr != "" || saved.Enabled {
		t.Fatalf("deleting legacy job deleted plan or kept cron: %+v,%v", saved, err)
	}
	scheduler.GetSchedulerManager().RemoveNodeCheckProfileJob(p.ID)
}

func TestNodeCheckCenterEmptyExplicitNodesNeverLaunchAll(t *testing.T) {
	setupPreviewAPITestDB(t)
	p := models.NodeCheckProfile{Name: "single-node-plan", Mode: "tcp"}
	if err := p.Add(); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.POST("/run", RunNodeCheck)
	for _, payload := range []map[string]any{{"profileId": p.ID, "scope": "nodes", "nodeIds": []int{}}, {"profileId": p.ID, "scope": "nodes", "nodeIds": []int{999}}, {"profileId": p.ID, "scope": "filtered"}, {"profileId": p.ID, "scope": "filtered", "tags": []string{" "}}, {"profileId": p.ID, "scope": "filtered", "tags": []string{""}}} {
		response := checkCenterRequest(t, r, http.MethodPost, "/run", payload)
		if string(response["code"]) == "200" {
			t.Fatalf("unsafe request accepted: %v", payload)
		}
	}
}

func TestDeleteSmartGroupRejectsPausedDetectionJobReference(t *testing.T) {
	setupPreviewAPITestDB(t)
	group := models.SmartGroup{Name: "job-referenced", Countries: "GB"}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	job := models.NodeCheckJob{Name: "paused", ProfileID: 1, Scope: "smart_group", SmartGroupID: group.ID, Enabled: false}
	if err := job.Save(); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.DELETE("/groups/:id", DeleteSmartGroup)
	response := checkCenterRequest(t, r, http.MethodDelete, "/groups/"+strconv.Itoa(group.ID), nil)
	if string(response["code"]) != "409" {
		t.Fatalf("referenced group deletion=%s", response)
	}
	if _, err := models.GetSmartGroup(group.ID); err != nil {
		t.Fatal(err)
	}
}

func TestNodeCheckCenterPausedLegacyJobStillValidatesScope(t *testing.T) {
	setupPreviewAPITestDB(t)
	p := models.NodeCheckProfile{Name: "paused-legacy", Mode: "tcp", Groups: "keep"}
	if err := p.Add(); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.PUT("/jobs/:id", SaveNodeCheckJob)
	payload := map[string]any{"name": p.Name, "profileId": p.ID, "enabled": false, "trigger": "scheduled", "cronExpr": "", "scope": "filtered", "groups": []string{}, "tags": []string{}}
	response := checkCenterRequest(t, r, http.MethodPut, "/jobs/"+strconv.Itoa(-p.ID), payload)
	if string(response["code"]) == "200" {
		t.Fatal("paused empty filtered scope accepted")
	}
	saved, err := models.GetNodeCheckProfileByID(p.ID)
	if err != nil || saved.Groups != "keep" {
		t.Fatalf("scope expanded: %+v,%v", saved, err)
	}
	payload["groups"] = []string{"keep"}
	response = checkCenterRequest(t, r, http.MethodPut, "/jobs/"+strconv.Itoa(-p.ID), payload)
	if string(response["code"]) != "200" {
		t.Fatalf("valid paused empty cron rejected: %s", response)
	}
}
