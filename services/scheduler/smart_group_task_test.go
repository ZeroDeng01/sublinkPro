package scheduler

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"sublink/database"
	"sublink/internal/testutil"
	"sublink/models"

	"github.com/robfig/cron/v3"
)

func setupSmartGroupCheckTest(t *testing.T) models.NodeCheckProfile {
	t.Helper()
	oldDB, oldDialect, oldInitialized := database.DB, database.Dialect, database.IsInitialized
	db := testutil.OpenMemoryDB(t, "smart_group_check")
	if err := db.AutoMigrate(&models.SmartGroup{}, &models.Node{}, &models.NodeCheckProfile{}); err != nil {
		t.Fatal(err)
	}
	database.DB, database.Dialect, database.IsInitialized = db, database.DialectSQLite, true
	if err := models.InitNodeCache(); err != nil {
		t.Fatal(err)
	}
	if err := models.InitNodeCheckProfileCache(); err != nil {
		t.Fatal(err)
	}
	oldExecute := executeSmartGroupNodes
	t.Cleanup(func() {
		waitSmartGroupChecksIdle(t)
		executeSmartGroupNodes = oldExecute
		database.DB, database.Dialect, database.IsInitialized = oldDB, oldDialect, oldInitialized
		if oldDB != nil {
			_ = models.InitNodeCache()
			_ = models.InitNodeCheckProfileCache()
		}
		testutil.CloseDB(t, db)
	})
	profile := models.NodeCheckProfile{Name: "candidate-check", Mode: "tcp", Groups: "unrelated", Enabled: false}
	if err := profile.Add(); err != nil {
		t.Fatal(err)
	}
	return profile
}

func waitSmartGroupChecksIdle(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		smartGroupCheckMu.Lock()
		count := len(smartGroupChecks)
		smartGroupCheckMu.Unlock()
		if count == 0 {
			done := make(chan struct{})
			go func() { smartGroupFollowups.Wait(); close(done) }()
			select {
			case <-done:
				return
			case <-time.After(3 * time.Second):
				t.Fatal("queued smart group workers did not finish")
			}
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("smart group checks did not finish")
}

func saveSmartCheckGroup(t *testing.T, profile models.NodeCheckProfile, name string) models.SmartGroup {
	t.Helper()
	group := models.SmartGroup{Name: name, Countries: "GB", CheckProfileID: profile.ID, ScheduleEnabled: true, CronExpr: "0 3 * * *", CheckAfterAirportUpdate: true}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	return group
}

func addSmartCheckNode(t *testing.T, name, country, group string, sourceID int) models.Node {
	t.Helper()
	n := models.Node{Name: name, LinkName: name, Link: "ss://" + name, LinkCountry: country, Group: group, SourceID: sourceID, DelayStatus: "timeout"}
	if err := n.Add(); err != nil {
		t.Fatal(err)
	}
	return n
}

type smartCheckCall struct {
	profile int
	ids     []int
	trigger models.TaskTrigger
	name    string
}

func awaitSmartCheck(t *testing.T, calls <-chan smartCheckCall) smartCheckCall {
	t.Helper()
	select {
	case call := <-calls:
		return call
	case <-time.After(3 * time.Second):
		t.Fatal("no candidate check started")
		return smartCheckCall{}
	}
}

func TestSmartGroupManualChecksFailedCandidatesAndPreventsOverlap(t *testing.T) {
	profile := setupSmartGroupCheckTest(t)
	group := saveSmartCheckGroup(t, profile, "Europe")
	n := addSmartCheckNode(t, "selected", "GB", "source", 1)
	outside := addSmartCheckNode(t, "outside", "US", "source", 1)
	calls := make(chan smartCheckCall, 3)
	release := make(chan struct{})
	defer close(release)
	executeSmartGroupNodes = func(p int, ids []int, trigger models.TaskTrigger, name string) {
		calls <- smartCheckCall{p, ids, trigger, name}
		<-release
	}
	if _, err := StartSmartGroupCheck(group.ID, profile.ID, outside.ID, models.TaskTriggerManual); err == nil {
		t.Fatal("outside candidate accepted")
	}
	count, err := StartSmartGroupCheck(group.ID, profile.ID, 0, models.TaskTriggerManual)
	if err != nil || count != 1 {
		t.Fatalf("start: count=%d err=%v", count, err)
	}
	call := awaitSmartCheck(t, calls)
	if !reflect.DeepEqual(call.ids, []int{n.ID}) || call.trigger != models.TaskTriggerManual || !strings.Contains(call.name, group.Name) {
		t.Fatalf("call = %+v", call)
	}
	if !SmartGroupCheckRunning(group.ID) {
		t.Fatal("missing reservation")
	}
	if _, err := StartSmartGroupCheck(group.ID, profile.ID, n.ID, models.TaskTriggerManual); !errors.Is(err, ErrSmartGroupCheckRunning) {
		t.Fatalf("overlap err = %v", err)
	}
	RunScheduledSmartGroupCheck(group.ID)
	if len(calls) != 0 {
		t.Fatal("cron overlapped manual check")
	}
	saved, err := models.GetSmartGroup(group.ID)
	if err != nil || saved.LastCheckAt == nil {
		t.Fatalf("last trigger not recorded: %+v %v", saved, err)
	}
}

func TestSmartGroupEmptyInvalidAndDisabledChecksNeverRunFallback(t *testing.T) {
	profile := setupSmartGroupCheckTest(t)
	group := saveSmartCheckGroup(t, profile, "empty")
	calls := make(chan smartCheckCall, 5)
	executeSmartGroupNodes = func(p int, ids []int, trigger models.TaskTrigger, name string) {
		calls <- smartCheckCall{p, ids, trigger, name}
	}
	addSmartCheckNode(t, "not-selected", "US", "source", 1)
	if _, err := StartSmartGroupCheck(group.ID, profile.ID, 0, models.TaskTriggerManual); !errors.Is(err, ErrSmartGroupNoCandidates) {
		t.Fatalf("empty err = %v", err)
	}
	RunScheduledSmartGroupCheck(group.ID)
	RunSmartGroupChecksAfterAirportUpdate(1)
	if len(calls) != 0 {
		t.Fatal("empty candidates invoked runner")
	}
	addSmartCheckNode(t, "selected", "GB", "source", 1)
	if err := database.DB.Model(&group).Update("schedule_enabled", false).Error; err != nil {
		t.Fatal(err)
	}
	RunScheduledSmartGroupCheck(group.ID)
	if err := database.DB.Delete(&group).Error; err != nil {
		t.Fatal(err)
	}
	RunScheduledSmartGroupCheck(group.ID)
	RunSmartGroupChecksAfterAirportUpdate(1)
	if len(calls) != 0 {
		t.Fatal("disabled/deleted schedule started")
	}
}

func TestSmartGroupScheduledReReadsCandidatesAndRejectsTCPForSpeed(t *testing.T) {
	profile := setupSmartGroupCheckTest(t)
	group := saveSmartCheckGroup(t, profile, "fresh")
	addSmartCheckNode(t, "original", "GB", "source", 1)
	n := addSmartCheckNode(t, "new-selection", "DE", "source", 2)
	calls := make(chan smartCheckCall, 3)
	executeSmartGroupNodes = func(p int, ids []int, trigger models.TaskTrigger, name string) {
		calls <- smartCheckCall{p, ids, trigger, name}
	}
	if err := database.DB.Model(&group).Update("countries", "DE").Error; err != nil {
		t.Fatal(err)
	}
	RunScheduledSmartGroupCheck(group.ID)
	call := awaitSmartCheck(t, calls)
	if !reflect.DeepEqual(call.ids, []int{n.ID}) || call.trigger != models.TaskTriggerScheduled {
		t.Fatalf("stale scheduled config: %+v", call)
	}
	waitSmartGroupChecksIdle(t)
	group.MinSpeed = 1
	if err := ValidateSmartGroupCheckProfile(group, profile.ID); err == nil {
		t.Fatal("TCP accepted for download requirement")
	}
	if err := ValidateSmartGroupCheckProfile(group, 999999); err == nil {
		t.Fatal("nonexistent profile accepted")
	}
}

func TestSmartGroupAirportIntersectionAndProfileDeduplication(t *testing.T) {
	profile := setupSmartGroupCheckTest(t)
	first := saveSmartCheckGroup(t, profile, "one")
	second := saveSmartCheckGroup(t, profile, "two")
	second.SourceGroups = []string{"allowed"}
	if err := database.DB.Model(&second).Select("SourceGroups").Updates(&second).Error; err != nil {
		t.Fatal(err)
	}
	a := addSmartCheckNode(t, "shared", "GB", "allowed", 1)
	b := addSmartCheckNode(t, "one-only", "GB", "excluded", 1)
	addSmartCheckNode(t, "other-country", "US", "allowed", 1)
	addSmartCheckNode(t, "other-airport", "GB", "allowed", 2)
	calls := make(chan smartCheckCall, 5)
	executeSmartGroupNodes = func(p int, ids []int, trigger models.TaskTrigger, name string) {
		calls <- smartCheckCall{p, ids, trigger, name}
	}
	RunSmartGroupChecksAfterAirportUpdate(1)
	call := awaitSmartCheck(t, calls)
	if !reflect.DeepEqual(call.ids, []int{a.ID, b.ID}) || call.trigger != models.TaskTriggerAirportUpdate || !strings.Contains(call.name, first.Name) || !strings.Contains(call.name, second.Name) {
		t.Fatalf("airport batch = %+v", call)
	}
	RunSmartGroupChecksAfterAirportUpdate(999)
	if len(calls) != 0 {
		t.Fatal("duplicate profile batch or no-match runner")
	}
}

func TestSmartGroupAirportUpdatesQueueAndCoalesceFollowup(t *testing.T) {
	profile := setupSmartGroupCheckTest(t)
	first := saveSmartCheckGroup(t, profile, "first")
	second := saveSmartCheckGroup(t, profile, "second")
	oldNode := addSmartCheckNode(t, "existing", "GB", "source", 1)
	calls := make(chan smartCheckCall, 5)
	release := make(chan struct{})
	released := false
	defer func() {
		if !released {
			close(release)
		}
	}()
	executeSmartGroupNodes = func(p int, ids []int, trigger models.TaskTrigger, name string) {
		calls <- smartCheckCall{p, ids, trigger, name}
		<-release
	}
	done := make(chan struct{})
	go func() { RunSmartGroupChecksAfterAirportUpdate(1); close(done) }()
	initial := awaitSmartCheck(t, calls)
	if !reflect.DeepEqual(initial.ids, []int{oldNode.ID}) {
		t.Fatalf("initial = %+v", initial)
	}
	newNode := addSmartCheckNode(t, "added-during-check", "GB", "source", 1)
	RunSmartGroupChecksAfterAirportUpdate(1)
	RunSmartGroupChecksAfterAirportUpdate(1)
	if len(calls) != 0 {
		t.Fatal("queued update overlapped active check")
	}
	close(release)
	released = true
	followup := awaitSmartCheck(t, calls)
	if !reflect.DeepEqual(followup.ids, []int{oldNode.ID, newNode.ID}) || !strings.Contains(followup.name, first.Name) || !strings.Contains(followup.name, second.Name) {
		t.Fatalf("followup not coalesced/fresh: %+v", followup)
	}
	<-done
	waitSmartGroupChecksIdle(t)
	if len(calls) != 0 {
		t.Fatal("more than one followup")
	}
}

func TestSmartGroupQueuedUpdateHonorsDisabledLinkage(t *testing.T) {
	profile := setupSmartGroupCheckTest(t)
	group := saveSmartCheckGroup(t, profile, "disabled-linkage")
	addSmartCheckNode(t, "selected", "GB", "source", 1)
	if !reserveSmartGroupCheck(group.ID, nil) {
		t.Fatal("reserve")
	}
	RunSmartGroupChecksAfterAirportUpdate(1)
	if err := database.DB.Model(&group).Update("check_after_airport_update", false).Error; err != nil {
		t.Fatal(err)
	}
	calls := make(chan smartCheckCall, 1)
	executeSmartGroupNodes = func(p int, ids []int, trigger models.TaskTrigger, name string) {
		calls <- smartCheckCall{p, ids, trigger, name}
	}
	// Release before the queued pass: an active reservation must not mask a
	// broken disabled-linkage check and make this assertion falsely pass.
	smartGroupCheckMu.Lock()
	pending := smartGroupChecks[group.ID]
	delete(smartGroupChecks, group.ID)
	smartGroupCheckMu.Unlock()
	if len(pending) != 1 {
		t.Fatalf("update was not queued: %+v", pending)
	}
	runSmartGroupAirportChecks([]int{1}, map[int]map[int]struct{}{group.ID: pending})
	if len(calls) != 0 {
		t.Fatal("disabled update linkage ran")
	}
}

func TestSmartGroupCronLifecycleHasSeparateNamespace(t *testing.T) {
	profile := setupSmartGroupCheckTest(t)
	group := saveSmartCheckGroup(t, profile, "scheduled")
	sm := &SchedulerManager{cron: cron.New(), jobs: make(map[int]cron.EntryID), smartGroupJobs: make(map[int]cron.EntryID)}
	airportEntry, err := sm.cron.AddFunc("0 0 * * *", func() {})
	if err != nil {
		t.Fatal(err)
	}
	sm.jobs[group.ID] = airportEntry
	if err := sm.LoadSmartGroupJobs(); err != nil {
		t.Fatal(err)
	}
	first := sm.smartGroupJobs[group.ID]
	if first == 0 || len(sm.cron.Entries()) != 2 {
		t.Fatalf("jobs = %+v", sm.cron.Entries())
	}
	info := sm.SmartGroupCheckInfo(group)
	if info.NextCheckAt == nil || info.CheckTimezone == "" {
		t.Fatalf("missing runtime info: %+v", info)
	}
	group.CronExpr = "bad"
	if err := sm.UpdateSmartGroupJob(group); err == nil {
		t.Fatal("invalid replacement accepted")
	}
	if sm.smartGroupJobs[group.ID] != first {
		t.Fatal("valid old entry discarded")
	}
	group.CronExpr = "0 4 * * *"
	if err := sm.UpdateSmartGroupJob(group); err != nil {
		t.Fatal(err)
	}
	if sm.smartGroupJobs[group.ID] == first || len(sm.cron.Entries()) != 2 {
		t.Fatal("replacement leaked old entry")
	}
	group.ScheduleEnabled = false
	if err := sm.UpdateSmartGroupJob(group); err != nil {
		t.Fatal(err)
	}
	if len(sm.smartGroupJobs) != 0 || len(sm.cron.Entries()) != 1 {
		t.Fatal("disable removed unrelated job or leaked schedule")
	}
	group.ScheduleEnabled = true
	if err := sm.UpdateSmartGroupJob(group); err != nil {
		t.Fatal(err)
	}
	sm.RemoveSmartGroupJob(group.ID)
	if len(sm.smartGroupJobs) != 0 || sm.jobs[group.ID] != airportEntry {
		t.Fatal("delete collided with airport namespace")
	}
}

func TestSmartGroupBatchReleasesReservationOnEmptyOrPanic(t *testing.T) {
	profile := setupSmartGroupCheckTest(t)
	group := saveSmartCheckGroup(t, profile, "release")
	calls := 0
	executeSmartGroupNodes = func(_ int, _ []int, _ models.TaskTrigger, _ string) { calls++; panic("test failure") }
	if !reserveSmartGroupCheck(group.ID, nil) {
		t.Fatal("reserve")
	}
	runSmartGroupBatch(profile.ID, []models.SmartGroup{group}, nil, models.TaskTriggerScheduled)
	if calls != 0 || SmartGroupCheckRunning(group.ID) {
		t.Fatal("empty batch fell back or leaked lock")
	}
	if !reserveSmartGroupCheck(group.ID, nil) {
		t.Fatal("reserve")
	}
	runSmartGroupBatch(profile.ID, []models.SmartGroup{group}, []int{1}, models.TaskTriggerScheduled)
	if calls != 1 || SmartGroupCheckRunning(group.ID) {
		t.Fatal("panic leaked reservation")
	}
}
