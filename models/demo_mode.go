package models

import (
	"os"
	"strings"
	"sync"
)

var (
	demoMode     bool
	demoModeOnce sync.Once
)

// IsDemoMode 返回当前是否处于演示模式
// 通过环境变量 SUBLINK_DEMO_MODE 控制
func IsDemoMode() bool {
	demoModeOnce.Do(func() {
		val := os.Getenv("SUBLINK_DEMO_MODE")
		demoMode = strings.EqualFold(val, "true") || val == "1"
	})
	return demoMode
}
