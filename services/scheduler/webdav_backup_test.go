package scheduler

import "testing"

func TestUpdateWebDAVBackupJob(t *testing.T) {
	sm := GetSchedulerManager()
	if err := sm.UpdateWebDAVBackupJob("0 3 * * *", true); err != nil {
		t.Fatal(err)
	}
	sm.mutex.RLock()
	_, exists := sm.jobs[JobIDWebDAVBackup]
	sm.mutex.RUnlock()
	if !exists {
		t.Fatal("expected WebDAV backup job to be registered")
	}
	if err := sm.UpdateWebDAVBackupJob("0 3 * * *", false); err != nil {
		t.Fatal(err)
	}
	sm.mutex.RLock()
	_, exists = sm.jobs[JobIDWebDAVBackup]
	sm.mutex.RUnlock()
	if exists {
		t.Fatal("expected WebDAV backup job to be removed")
	}
}
