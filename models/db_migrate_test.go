package models

import (
	"testing"

	"sublink/internal/testutil"
)

func TestIsFreshDatabaseDetectsEmptySchema(t *testing.T) {
	db := testutil.OpenMemoryDB(t, "fresh_database_test")
	t.Cleanup(func() {
		testutil.CloseDB(t, db)
	})

	if !isFreshDatabase(db) {
		t.Fatal("empty database should be reported as fresh")
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("auto migrate users: %v", err)
	}
	// 用户表已存在说明不是全新安装，默认模板不应再被补写。
	if isFreshDatabase(db) {
		t.Fatal("database with an existing users table should not be reported as fresh")
	}
}
