package models

import "testing"

func TestNodeCheckJobValidatesExplicitScopeAndSchedule(t *testing.T) {
	base := NodeCheckJob{Name: "check", ProfileID: 1, Trigger: "scheduled", CronExpr: "30 5 * * *", Scope: "all"}
	cases := []struct {
		name    string
		change  func(*NodeCheckJob)
		invalid bool
	}{
		{"valid", func(j *NodeCheckJob) {}, false},
		{"empty-filter", func(j *NodeCheckJob) { j.Scope = "filtered" }, true},
		{"hidden-filter-on-all", func(j *NodeCheckJob) { j.Groups = []string{"hidden"} }, true},
		{"unsupported-scope", func(j *NodeCheckJob) { j.Scope = "smart_group" }, true},
		{"unknown-trigger", func(j *NodeCheckJob) { j.Trigger = "unknown" }, true},
		{"six-field-cron", func(j *NodeCheckJob) { j.CronExpr = "0 30 5 * * *" }, true},
		{"zero-step", func(j *NodeCheckJob) { j.CronExpr = "*/0 * * * *" }, true},
		{"preupdate-needs-airports", func(j *NodeCheckJob) { j.Scope = "filtered"; j.Groups = []string{"a"}; j.UpdateBeforeCheck = true }, true},
		{"event-cannot-preupdate", func(j *NodeCheckJob) {
			j.Trigger = "airport_update"
			j.Scope = "filtered"
			j.AirportIDs = []int{1}
			j.UpdateBeforeCheck = true
		}, true},
		{"event-no-cron", func(j *NodeCheckJob) { j.Trigger = "airport_update"; j.CronExpr = "" }, false},
		{"filtered-preupdate", func(j *NodeCheckJob) { j.Scope = "filtered"; j.AirportIDs = []int{1, 1}; j.UpdateBeforeCheck = true }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			job := base
			tc.change(&job)
			err := job.Validate()
			if (err != nil) != tc.invalid {
				t.Fatalf("Validate()=%v invalid=%v", err, tc.invalid)
			}
		})
	}
}
