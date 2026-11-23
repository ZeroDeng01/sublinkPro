package models

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"sublink/database"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Use dedicated disposable databases: these tests create and delete their own
// share rows, leaving unrelated rows untouched. DSNs are never printed.
func TestShareDeviceConcurrentAdmissionSQLDatabases(t *testing.T) {
	for _, dialect := range []string{"mysql", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			dsn := os.Getenv("SUBLINK_TEST_" + map[string]string{"mysql": "MYSQL", "postgres": "POSTGRES"}[dialect] + "_DSN")
			if dsn == "" {
				t.Skip("dedicated database DSN not configured")
			}
			driver := mysql.Open(dsn)
			if dialect == "postgres" {
				driver = postgres.Open(dsn)
			}
			db, err := gorm.Open(driver, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			if err != nil {
				t.Fatal("cannot open dedicated test database")
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			sqlDB.SetMaxOpenConns(24)
			sqlDB.SetMaxIdleConns(24)
			t.Cleanup(func() { _ = sqlDB.Close() })
			if err := db.AutoMigrate(&SubscriptionShare{}, &ShareDevice{}); err != nil {
				t.Fatal(err)
			}
			oldDB, oldDialect := database.DB, database.Dialect
			database.DB, database.Dialect = db, dialect
			t.Cleanup(func() { database.DB, database.Dialect = oldDB, oldDialect })
			for iteration := range 10 {
				token, err := GenerateToken()
				if err != nil {
					t.Fatal(err)
				}
				share := SubscriptionShare{Token: token, Enabled: true, MaxDevices: 1}
				if err := db.Create(&share).Error; err != nil {
					t.Fatal(err)
				}
				start := make(chan struct{})
				results := make(chan error, 24)
				var workers sync.WaitGroup
				for i := range 24 {
					workers.Add(1)
					go func(i int) {
						defer workers.Done()
						<-start
						results <- CheckShareDevice(share.ID, token, ShareDeviceIdentity{HWID: fmt.Sprintf("device-%d", i)}, true)
					}(i)
				}
				close(start)
				workers.Wait()
				close(results)
				admitted, denied := 0, 0
				for err := range results {
					var accessError *ShareAccessError
					if err == nil {
						admitted++
					} else if errors.As(err, &accessError) && accessError.Code == "device_limit_exceeded" {
						denied++
					} else {
						t.Errorf("unexpected admission failure: %v", err)
					}
				}
				var count int64
				if err := db.Model(&ShareDevice{}).Where("share_id = ?", share.ID).Count(&count).Error; err != nil {
					t.Error(err)
				}
				if err := db.Where("share_id = ?", share.ID).Delete(&ShareDevice{}).Error; err != nil {
					t.Error(err)
				}
				if err := db.Delete(&share).Error; err != nil {
					t.Error(err)
				}
				if admitted != 1 || denied != 23 || count != 1 {
					t.Fatalf("round %d: admitted=%d denied=%d persisted=%d", iteration+1, admitted, denied, count)
				}
			}
			t.Log("10 rounds × 24 concurrent requests: one admission, 23 limit denials, zero database errors per round")
		})
	}
}
