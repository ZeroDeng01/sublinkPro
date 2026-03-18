package testutil

import (
	"fmt"
	"strings"
	"testing"

	"gorm.io/gorm"
)

// UniqueMemoryDSN 为每个测试生成独立的命名内存 SQLite DSN。
func UniqueMemoryDSN(t testing.TB, prefix string) string {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	return fmt.Sprintf("file:%s_%s?mode=memory&cache=shared", prefix, name)
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
