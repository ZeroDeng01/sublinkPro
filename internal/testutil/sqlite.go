package testutil

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// UniqueMemoryDSN 为每个测试生成独立的命名内存 SQLite DSN。
func UniqueMemoryDSN(t testing.TB, prefix string) string {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	return fmt.Sprintf("file:%s_%s?mode=memory&cache=shared", prefix, name)
}

// OpenMemoryDB 打开一个独立的命名内存 SQLite 测试库，并关闭 GORM 日志，
// 避免 "record not found" 等正常查询信息污染测试输出。失败时直接终止测试。
func OpenMemoryDB(t testing.TB, prefix string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(UniqueMemoryDSN(t, prefix)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	return db
}

// CloseDB 关闭测试数据库连接并在失败时标记测试错误。
func CloseDB(t testing.TB, db *gorm.DB) {
	t.Helper()
	if db == nil {
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Errorf("get sql db: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		t.Errorf("close test db: %v", err)
	}
}
