package services

import (
	"sublink/internal/testutil"
	"sublink/models"
	"testing"
)

func TestDatabaseMigrationPreservesIndependentCheckJobsAndProfileIDs(t *testing.T) {
	for _, hasJobs := range []bool{true, false} {
		name := "legacy"
		if hasJobs {
			name = "with-jobs"
		}
		t.Run(name, func(t *testing.T) {
			source, target := testutil.OpenMemoryDB(t, "check_source"), testutil.OpenMemoryDB(t, "check_target")
			t.Cleanup(func() { testutil.CloseDB(t, source); testutil.CloseDB(t, target) })
			if err := source.AutoMigrate(&models.NodeCheckProfile{}); err != nil {
				t.Fatal(err)
			}
			if err := target.AutoMigrate(&models.NodeCheckProfile{}, &models.NodeCheckJob{}); err != nil {
				t.Fatal(err)
			}
			profile := models.NodeCheckProfile{ID: 7, Name: "restored-plan", Mode: "tcp"}
			if err := source.Create(&profile).Error; err != nil {
				t.Fatal(err)
			}
			if hasJobs {
				if err := source.AutoMigrate(&models.NodeCheckJob{}); err != nil {
					t.Fatal(err)
				}
				if err := source.Create(&models.NodeCheckJob{ID: 11, Name: "restored-job", ProfileID: 7, Scope: "all", Trigger: "scheduled", CronExpr: "30 5 * * *", Enabled: true}).Error; err != nil {
					t.Fatal(err)
				}
			}
			if err := target.Create(&models.NodeCheckJob{Name: "stale-target-job", ProfileID: 999}).Error; err != nil {
				t.Fatal(err)
			}
			if err := clearTargetBusinessData(target); err != nil {
				t.Fatal(err)
			}
			state := &databaseMigrationState{source: source, tx: target, sourceHasProfiles: true, result: &DatabaseMigrationResult{Imported: make(map[string]int)}}
			if err := importNodeCheckProfiles(state); err != nil {
				t.Fatal(err)
			}
			var jobs []models.NodeCheckJob
			if err := target.Find(&jobs).Error; err != nil {
				t.Fatal(err)
			}
			if !hasJobs {
				if len(jobs) != 0 {
					t.Fatalf("legacy backup retained stale jobs: %+v", jobs)
				}
				return
			}
			if len(jobs) != 1 || jobs[0].ID != 11 || jobs[0].ProfileID != 7 {
				t.Fatalf("job references=%+v", jobs)
			}
		})
	}
}
