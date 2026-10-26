package scheduler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sublink/models"
	"testing"
)

// completingSubscriptionTaskManager 模拟生产 CompleteTask 主动取消 context 的生命周期。
type completingSubscriptionTaskManager struct {
	ctx       context.Context
	cancel    context.CancelFunc
	completed bool
}

func (m *completingSubscriptionTaskManager) UpdateTotal(string, int) error { return nil }
func (m *completingSubscriptionTaskManager) UpdateProgress(string, int, string, any) error {
	return nil
}
func (m *completingSubscriptionTaskManager) CompleteTask(string, string, any) error {
	m.completed = true
	m.cancel()
	return nil
}
func (m *completingSubscriptionTaskManager) FailTask(string, string) error { m.cancel(); return nil }
func (m *completingSubscriptionTaskManager) CreateTask(models.TaskType, string, models.TaskTrigger, int) (*models.Task, context.Context, error) {
	return &models.Task{ID: "subscription-regression"}, m.ctx, nil
}

func TestNodeCheckPreupdateCompletionCancellationStillAppliesTagsUsageAndChecks(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		name := "successful-completion"
		if cancelled {
			name = "user-cancelled"
		}
		t.Run(name, func(t *testing.T) {
			profile := setupNodeCheckJobTest(t)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("subscription-userinfo", "upload=1; download=2; total=100; expire=2000000000")
				_, _ = w.Write([]byte("proxies:\n  - {name: updated-node, type: ss, server: 127.0.0.1, port: 443, cipher: aes-128-gcm, password: synthetic}\n"))
			}))
			t.Cleanup(server.Close)
			airport := models.Airport{Name: "preupdate-real-loader", URL: server.URL, FetchUsageInfo: true}
			if err := airport.Add(); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			if cancelled {
				cancel()
			}
			manager := &completingSubscriptionTaskManager{ctx: ctx, cancel: cancel}
			oldGetter, oldTags := taskManagerGetter, autoTagRulesApplier
			t.Cleanup(func() { taskManagerGetter = oldGetter; autoTagRulesApplier = oldTags })
			taskManagerGetter = func() TaskManagerInterface { return manager }
			tagged, checked := false, false
			autoTagRulesApplier = func(nodes []models.Node, source string) {
				if source != "subscription_update" || len(nodes) != 1 {
					t.Errorf("tag boundary=%s,%v", source, nodes)
				}
				tagged = true
			}
			runCheckJobNodes = func(_ *models.NodeCheckProfile, nodes []models.Node, _ models.TaskTrigger, _ string) error {
				if !tagged || len(nodes) != 1 || nodes[0].SourceID != airport.ID {
					t.Errorf("check before successful tagging/scope: %v,%v", tagged, nodes)
				}
				checked = true
				return nil
			}
			job := models.NodeCheckJob{ProfileID: profile.ID, Scope: "filtered", AirportIDs: []int{airport.ID}, UpdateBeforeCheck: true}
			err := executeNodeCheckJob(job, models.TaskTriggerScheduled, nil)
			if cancelled {
				if err == nil || checked || tagged {
					t.Fatalf("cancelled update continued: %v,checked=%v,tagged=%v", err, checked, tagged)
				}
				return
			}
			if err != nil || !manager.completed || ctx.Err() != context.Canceled || !checked {
				t.Fatalf("completed update stopped chain: err=%v,completed=%v,ctx=%v,checked=%v", err, manager.completed, ctx.Err(), checked)
			}
			saved, err := models.GetAirportByID(airport.ID)
			if err != nil || saved.UsageTotal != 100 || saved.UsageDownload != 2 {
				t.Fatalf("usage not saved: %+v,%v", saved, err)
			}
		})
	}
}
