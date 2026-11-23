package models

import (
	"errors"
	"sync"
	"testing"
	"time"

	"sublink/cache"
	"sublink/database"
	"sublink/internal/testutil"

	"gorm.io/gorm"
)

func setupTaskTestDB(t *testing.T) {
	t.Helper()
	oldDB, oldCache := database.DB, taskCache
	db := testutil.OpenMemoryDB(t, "task")
	if err := db.AutoMigrate(&Task{}); err != nil {
		t.Fatal(err)
	}
	database.DB = db
	taskCache = cache.NewMapCache(func(task Task) string { return task.ID })
	taskCache.AddIndex("status", func(task Task) string { return string(task.Status) })
	taskCache.AddIndex("type", func(task Task) string { return string(task.Type) })
	t.Cleanup(func() {
		database.DB, taskCache = oldDB, oldCache
		cache.Manager.Register("task", oldCache)
		testutil.CloseDB(t, db)
	})
}

func TestInitTaskCacheLoadsAllHistory(t *testing.T) {
	setupTaskTestDB(t)
	for _, task := range []Task{
		{ID: "old-completed", Status: TaskStatusCompleted, CreatedAt: time.Now().AddDate(-1, 0, 0)},
		{ID: "old-running", Status: TaskStatusRunning, CreatedAt: time.Now().AddDate(0, 0, -40)},
		{ID: "recent", Status: TaskStatusCompleted, CreatedAt: time.Now()},
	} {
		if err := database.DB.Create(&task).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := InitTaskCache(); err != nil {
		t.Fatal(err)
	}
	if got := taskCache.Count(); got != 3 {
		t.Fatalf("loaded %d tasks, want all 3 database records", got)
	}
	for _, id := range []string{"old-completed", "old-running", "recent"} {
		if _, ok := taskCache.Get(id); !ok {
			t.Fatalf("task %s was excluded by age", id)
		}
	}
}

func TestOldTasksRemainCachedUntilDatabaseDeletion(t *testing.T) {
	setupTaskTestDB(t)
	task := Task{ID: "old-completed", Status: TaskStatusCompleted, Type: TaskTypeSpeedTest, CreatedAt: time.Now().AddDate(0, 0, -40), Result: "retained result"}
	if err := task.Create(); err != nil {
		t.Fatal(err)
	}
	if err := task.SyncFinalStatus(); err != nil {
		t.Fatal(err)
	}
	if got, ok := taskCache.Get(task.ID); !ok || got.Result != task.Result {
		t.Fatal("old task was evicted before its database row was deleted")
	}
	var loaded Task
	if err := loaded.GetByID(task.ID); err != nil || loaded.Result != task.Result {
		t.Fatalf("old task unreadable: %+v, %v", loaded, err)
	}
}

func TestCleanupOldTasksSynchronouslyRemovesOnlyDeletedRowsFromCache(t *testing.T) {
	setupTaskTestDB(t)
	now := time.Now()
	for _, task := range []Task{
		{ID: "completed", Status: TaskStatusCompleted, CreatedAt: now.AddDate(0, 0, -40)},
		{ID: "cancelled", Status: TaskStatusCancelled, CreatedAt: now.AddDate(0, 0, -40)},
		{ID: "failed", Status: TaskStatusError, CreatedAt: now.AddDate(0, 0, -40)},
		{ID: "running", Status: TaskStatusRunning, CreatedAt: now.AddDate(0, 0, -40)},
		{ID: "pending", Status: TaskStatusPending, CreatedAt: now.AddDate(0, 0, -40)},
		{ID: "recent", Status: TaskStatusCompleted, CreatedAt: now},
	} {
		if err := task.Create(); err != nil {
			t.Fatal(err)
		}
	}
	deleted, err := CleanupOldTasks(now.AddDate(0, 0, -30))
	if err != nil || deleted != 3 {
		t.Fatalf("deleted %d tasks: %v", deleted, err)
	}
	for _, id := range []string{"completed", "cancelled", "failed"} {
		if _, ok := taskCache.Get(id); ok {
			t.Fatalf("deleted task %s remains cached", id)
		}
		var task Task
		if err := task.GetByID(id); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("deleted task %s still readable: %v", id, err)
		}
	}
	for _, id := range []string{"running", "pending", "recent"} {
		if _, ok := taskCache.Get(id); !ok {
			t.Fatalf("undeleted task %s was evicted", id)
		}
	}
	if got := len(taskCache.GetByIndex("status", string(TaskStatusCompleted))); got != 1 {
		t.Fatalf("completed index has %d entries, want only recent task", got)
	}
}

func TestTaskUpdateAfterCleanupDoesNotRestoreDeletedCacheEntry(t *testing.T) {
	for _, update := range []string{"update", "final-status", "status"} {
		t.Run(update, func(t *testing.T) {
			setupTaskTestDB(t)
			task := Task{ID: "deleted", Status: TaskStatusCompleted, CreatedAt: time.Now().AddDate(0, 0, -40)}
			if err := task.Create(); err != nil {
				t.Fatal(err)
			}
			if _, err := CleanupOldTasks(time.Now().AddDate(0, 0, -30)); err != nil {
				t.Fatal(err)
			}
			var err error
			switch update {
			case "update":
				err = task.Update()
			case "final-status":
				err = task.SyncFinalStatus()
			case "status":
				err = task.UpdateStatus(TaskStatusError, "late update")
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := taskCache.Get(task.ID); ok {
				t.Fatal("late update restored deleted task to cache")
			}
		})
	}
}

func TestTaskCleanupAndUpdatesRemainConsistent(t *testing.T) {
	setupTaskTestDB(t)
	task := Task{ID: "concurrent", Status: TaskStatusCompleted, CreatedAt: time.Now().AddDate(0, 0, -40)}
	if err := task.Create(); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			copy := task
			if err := copy.SyncFinalStatus(); err != nil {
				t.Error(err)
			}
		})
		wg.Go(func() {
			if _, err := CleanupOldTasks(time.Now().AddDate(0, 0, -30)); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if _, ok := taskCache.Get(task.ID); ok {
		t.Fatal("deleted task remains cached after concurrent writes")
	}
}

func TestCleanupFailureKeepsDatabaseAndCache(t *testing.T) {
	setupTaskTestDB(t)
	task := Task{ID: "keep-on-error", Status: TaskStatusCompleted, CreatedAt: time.Now().AddDate(0, 0, -40)}
	if err := task.Create(); err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("delete rejected")
	if err := database.DB.Callback().Delete().Before("gorm:delete").Register("test:reject-delete", func(tx *gorm.DB) {
		_ = tx.AddError(wantErr)
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := CleanupOldTasks(time.Now().AddDate(0, 0, -30)); !errors.Is(err, wantErr) {
		t.Fatalf("cleanup error = %v", err)
	}
	if _, ok := taskCache.Get(task.ID); !ok {
		t.Fatal("failed deletion evicted cached task")
	}
	var count int64
	if err := database.DB.Model(&Task{}).Where("id = ?", task.ID).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("database row lost: count=%d error=%v", count, err)
	}
}
