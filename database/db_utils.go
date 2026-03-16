package database

import (
	"strings"
	"sublink/utils"
	"time"

	"gorm.io/gorm"
)

// WithTransaction executes a function within a database transaction.
// It automatically handles commit on success and rollback on error/panic.
func WithTransaction(fn func(tx *gorm.DB) error) error {
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			utils.Error("事务执行时发生panic，已回滚: %v", r)
			panic(r) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// WithRetry retries a database operation on lock errors.
// maxRetries: maximum number of retry attempts
// delay: initial delay between retries (doubles each retry)
func WithRetry(maxRetries int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		// Check if it's a retryable database lock/deadlock error
		if isLockError(err) {
			if i < maxRetries-1 {
				utils.Warn("数据库锁冲突，第 %d 次重试，等待 %v...", i+1, delay)
				time.Sleep(delay)
				delay *= 2 // exponential backoff
				continue
			}
		}
		// Non-lock error or max retries exceeded
		break
	}
	return err
}

// isLockError checks if the error is a database lock error
func isLockError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// 覆盖 sqlite/mysql/postgres 常见锁冲突与死锁场景
	return strings.Contains(errStr, "database is locked") ||
		strings.Contains(errStr, "SQLITE_BUSY") ||
		strings.Contains(errStr, "database table is locked") ||
		strings.Contains(errStr, "Lock wait timeout exceeded") ||
		strings.Contains(errStr, "deadlock found") ||
		strings.Contains(errStr, "deadlock detected") ||
		strings.Contains(errStr, "could not serialize access") ||
		strings.Contains(errStr, "could not obtain lock")
}

// BatchSize is the recommended batch size for bulk operations
const BatchSize = 100

// ChunkSlice splits a slice into chunks of specified size
func ChunkIntSlice(slice []int, chunkSize int) [][]int {
	if chunkSize <= 0 {
		chunkSize = BatchSize
	}

	var chunks [][]int
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}
