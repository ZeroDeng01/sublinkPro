package scheduler

import (
	"errors"
	"github.com/robfig/cron/v3"
	"reflect"
	"strings"
	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"
	"sync"
	"testing"
	"time"
)

func setupNodeCheckJobTest(t *testing.T) models.NodeCheckProfile {
	t.Helper()
	oldDB, oldDialect, oldInitialized := database.DB, database.Dialect, database.IsInitialized
	db := testutil.OpenMemoryDB(t, "node_check_job")
	if err := db.AutoMigrate(&models.Node{}, &models.NodeCheckProfile{}, &models.NodeCheckJob{}, &models.Airport{}); err != nil {
		t.Fatal(err)
	}
	database.DB, database.Dialect, database.IsInitialized = db, database.DialectSQLite, true
	for _, initCache := range []func() error{models.InitNodeCache, models.InitNodeCheckProfileCache, models.InitAirportCache} {
		if err := initCache(); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		checkJobWorkers.Wait()
		database.DB, database.Dialect, database.IsInitialized = oldDB, oldDialect, oldInitialized
		if oldDB != nil {
			_ = models.InitNodeCache()
			_ = models.InitNodeCheckProfileCache()
			_ = models.InitAirportCache()
		}
		testutil.CloseDB(t, db)
	})
	profile := models.NodeCheckProfile{Name: "check-job-plan", Mode: "tcp", Groups: "unrelated"}
	if err := profile.Add(); err != nil {
		t.Fatal(err)
	}
	oldRunner, oldUpdate := runCheckJobNodes, updateCheckJobAirport
	t.Cleanup(func() { checkJobWorkers.Wait(); runCheckJobNodes = oldRunner; updateCheckJobAirport = oldUpdate })
	return profile
}

func TestNodeCheckJobScopeDoesNotFallBackToProfileOrAllNodes(t *testing.T) {
	setupNodeCheckJobTest(t)
	a := addCheckJobNode(t, "job-gb-a", "GB", "a", 1)
	addCheckJobNode(t, "job-us-b", "US", "b", 2)
	job := models.NodeCheckJob{Scope: "filtered", Groups: []string{"a"}}
	nodes, err := ResolveNodeCheckJobNodes(job, []int{1})
	if err != nil || len(nodes) != 1 || nodes[0].ID != a.ID {
		t.Fatalf("candidate airport intersection=%v,%v", nodes, err)
	}
	nodes, err = ResolveNodeCheckJobNodes(job, []int{2})
	if err != nil || len(nodes) != 0 {
		t.Fatalf("empty candidate intersection=%v,%v", nodes, err)
	}
	nodes, err = ResolveNodeCheckJobNodes(models.NodeCheckJob{Scope: "filtered", Groups: []string{"none"}}, nil)
	if err != nil || len(nodes) != 0 {
		t.Fatalf("missing group expanded scope=%v,%v", nodes, err)
	}
	nodes, err = ResolveNodeCheckJobNodes(models.NodeCheckJob{Scope: "all"}, nil)
	if err != nil || len(nodes) != 2 {
		t.Fatalf("all should ignore legacy profile scope: %v,%v", nodes, err)
	}
}

func TestNodeCheckJobPreupdateMustFinishBeforeCheckingAndStopsOnFailure(t *testing.T) {
	p := setupNodeCheckJobTest(t)
	addCheckJobNode(t, "preupdate-node", "GB", "a", 1)
	job := models.NodeCheckJob{ProfileID: p.ID, Scope: "filtered", AirportIDs: []int{1, 2}, UpdateBeforeCheck: true}
	order := []string{}
	updateCheckJobAirport = func(id int) error {
		if id == 1 {
			order = append(order, "update1")
		} else {
			order = append(order, "update2")
		}
		return nil
	}
	runCheckJobNodes = func(_ *models.NodeCheckProfile, nodes []models.Node, _ models.TaskTrigger, _ string) error {
		order = append(order, "check")
		if len(nodes) != 1 {
			t.Fatalf("nodes=%v", nodes)
		}
		return nil
	}
	if err := executeNodeCheckJob(job, models.TaskTriggerManual, nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(order, []string{"update1", "update2", "check"}) {
		t.Fatalf("order=%v", order)
	}
	order = nil
	updateCheckJobAirport = func(int) error { return errors.New("subscription failed") }
	if err := executeNodeCheckJob(job, models.TaskTriggerManual, nil); err == nil || !strings.Contains(err.Error(), "subscription failed") {
		t.Fatalf("update error=%v", err)
	}
	if len(order) != 0 {
		t.Fatalf("check ran after failed update: %v", order)
	}
	job.UpdateBeforeCheck = false
	job.Groups = []string{"none"}
	if err := executeNodeCheckJob(job, models.TaskTriggerManual, nil); err == nil {
		t.Fatal("empty selection should skip")
	}
	if len(order) != 0 {
		t.Fatal("empty scope reached runner")
	}
}

func TestNodeCheckJobCoalescesAirportEventsAndRejectsOverlappingManualRun(t *testing.T) {
	p := setupNodeCheckJobTest(t)
	airport := models.Airport{Name: "job-airport", URL: "https://example.invalid/sub"}
	if err := airport.Add(); err != nil {
		t.Fatal(err)
	}
	addCheckJobNode(t, "queued-job-node", "GB", "a", airport.ID)
	job := models.NodeCheckJob{Name: "events", ProfileID: p.ID, Enabled: true, Trigger: "airport_update", Scope: "all"}
	if err := job.Save(); err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}, 4), make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }); checkJobWorkers.Wait() })
	runCheckJobNodes = func(_ *models.NodeCheckProfile, _ []models.Node, _ models.TaskTrigger, _ string) error {
		entered <- struct{}{}
		<-release
		return nil
	}
	if err := StartNodeCheckJob(job.ID, models.TaskTriggerManual, 0); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("runner not started")
	}
	if err := StartNodeCheckJob(job.ID, models.TaskTriggerManual, 0); err == nil {
		t.Fatal("overlapping manual run accepted")
	}
	for range 3 {
		if err := StartNodeCheckJob(job.ID, models.TaskTriggerAirportUpdate, airport.ID); err != nil {
			t.Fatal(err)
		}
	}
	once.Do(func() { close(release) })
	checkJobWorkers.Wait()
	if len(entered) != 1 {
		t.Fatalf("expected exactly one coalesced rerun, got %d", len(entered))
	}
	if NodeCheckJobRunning(job.ID) {
		t.Fatal("reservation leaked")
	}
	saved, err := models.GetNodeCheckJob(job.ID)
	if err != nil || saved.LastRunTime == nil || saved.LastError != "" {
		t.Fatalf("runtime=%+v,%v", saved, err)
	}
}

func TestNodeCheckJobQueuedEventsRespectDisabledJob(t *testing.T) {
	p := setupNodeCheckJobTest(t)
	addCheckJobNode(t, "disable-queued-node", "GB", "a", 1)
	job := models.NodeCheckJob{Name: "disable-events", ProfileID: p.ID, Enabled: true, Trigger: "airport_update", Scope: "all"}
	if err := job.Save(); err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}, 3), make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }); checkJobWorkers.Wait() })
	runCheckJobNodes = func(_ *models.NodeCheckProfile, _ []models.Node, _ models.TaskTrigger, _ string) error {
		entered <- struct{}{}
		<-release
		return nil
	}
	if err := StartNodeCheckJob(job.ID, models.TaskTriggerManual, 0); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("runner not started")
	}
	if err := StartNodeCheckJob(job.ID, models.TaskTriggerAirportUpdate, 1); err != nil {
		t.Fatal(err)
	}
	job.Enabled = false
	if err := job.Save(); err != nil {
		t.Fatal(err)
	}
	once.Do(func() { close(release) })
	checkJobWorkers.Wait()
	if len(entered) != 0 {
		t.Fatal("disabled job executed queued check")
	}
}

func TestNodeCheckJobCronFailurePreservesEntryAndLegacyNamespace(t *testing.T) {
	sm := &SchedulerManager{cron: cron.New(), jobs: make(map[int]cron.EntryID)}
	job := models.NodeCheckJob{ID: 1, Enabled: true, Trigger: "scheduled", CronExpr: "0 3 * * *"}
	if err := sm.UpdateNodeCheckJobSchedule(job); err != nil {
		t.Fatal(err)
	}
	old := sm.nodeCheckJobs[1]
	sm.jobs[nodeCheckJobIDOffset+1] = 999
	job.CronExpr = "invalid"
	if err := sm.UpdateNodeCheckJobSchedule(job); err == nil {
		t.Fatal("invalid cron accepted")
	}
	if sm.nodeCheckJobs[1] != old || sm.jobs[nodeCheckJobIDOffset+1] != 999 {
		t.Fatal("schedule replaced or namespace collided")
	}
	job.Enabled = false
	if err := sm.UpdateNodeCheckJobSchedule(job); err != nil {
		t.Fatal(err)
	}
	if len(sm.nodeCheckJobs) != 0 {
		t.Fatal("disabled job retained cron")
	}
}

func TestLegacyCheckScheduleAdvancesNextTimeAfterTrigger(t *testing.T) {
	p := setupNodeCheckJobTest(t)
	p.Enabled = true
	p.CronExpr = "* * * * *"
	if err := p.Update(); err != nil {
		t.Fatal(err)
	}
	sm := &SchedulerManager{cron: cron.New(), jobs: make(map[int]cron.EntryID)}
	if err := sm.UpdateNodeCheckProfileJob(p.ID, p.CronExpr, true); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour)
	if err := p.UpdateNextRunTime(&old); err != nil {
		t.Fatal(err)
	}
	sm.cron.Entry(sm.jobs[nodeCheckJobIDOffset+p.ID]).Job.Run()
	got, err := models.GetNodeCheckProfileByID(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.NextRunTime == nil || !got.NextRunTime.After(time.Now()) {
		t.Fatalf("next time not advanced: %v", got.NextRunTime)
	}
}

func TestSubscriptionUpdateTagsCompleteBeforeDetectionScopeContinues(t *testing.T) {
	setupNodeCheckJobTest(t)
	addCheckJobNode(t, "tag-boundary", "GB", "a", 1)
	old := autoTagRulesApplier
	t.Cleanup(func() { autoTagRulesApplier = old })
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	autoTagRulesApplier = func(_ []models.Node, source string) {
		if source != "subscription_update" {
			t.Errorf("source=%s", source)
		}
		close(entered)
		<-release
	}
	go func() { applySubscriptionUpdateTags(1); close(done) }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("tags not started")
	}
	select {
	case <-done:
		t.Fatal("update boundary returned before tags completed")
	default:
	}
	close(release)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("tags did not complete")
	}
}

func addCheckJobNode(t *testing.T, name, country, group string, sourceID int) models.Node {
	t.Helper()
	n := models.Node{Name: name, LinkName: name, Link: "ss://" + name, LinkCountry: country, Group: group, SourceID: sourceID, DelayStatus: "timeout"}
	if err := n.Add(); err != nil {
		t.Fatal(err)
	}
	return n
}
