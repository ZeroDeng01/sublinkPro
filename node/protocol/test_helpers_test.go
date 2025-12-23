package protocol

import (
	"strconv"
	"testing"
)

// 测试辅助函数 - 用于验证编解码结果

// assertEqualString 验证两个字符串相等
func assertEqualString(t *testing.T, field string, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s 不匹配: 期望 [%s], 实际 [%s]", field, expected, actual)
	}
}

// assertEqualInt 验证两个整数相等
func assertEqualInt(t *testing.T, field string, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s 不匹配: 期望 %d, 实际 %d", field, expected, actual)
	}
}

// assertEqualFlexPort 验证 FlexPort 类型的端口
func assertEqualFlexPort(t *testing.T, field string, expected int, actual FlexPort) {
	t.Helper()
	if expected != int(actual) {
		t.Errorf("%s 不匹配: 期望 %d, 实际 %d", field, expected, int(actual))
	}
}

// assertEqualIntInterface 验证 interface{} 类型的整数（用于协议结构体的 Port 字段）
func assertEqualIntInterface(t *testing.T, field string, expected, actual interface{}) {
	t.Helper()
	expectedInt := toInt(expected)
	actualInt := toInt(actual)
	if expectedInt != actualInt {
		t.Errorf("%s 不匹配: 期望 %d, 实际 %d", field, expectedInt, actualInt)
	}
}

// toInt 将 interface{} 转换为 int
func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}

// assertEqualBool 验证两个布尔值相等
func assertEqualBool(t *testing.T, field string, expected, actual bool) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s 不匹配: 期望 %v, 实际 %v", field, expected, actual)
	}
}

// assertNotEmpty 验证字符串非空
func assertNotEmpty(t *testing.T, field string, value string) {
	t.Helper()
	if value == "" {
		t.Errorf("%s 不应为空", field)
	}
}

// assertContains 验证字符串包含子串
func assertContains(t *testing.T, field string, str, substr string) {
	t.Helper()
	if len(str) == 0 || len(substr) == 0 {
		t.Errorf("%s 验证失败: 字符串或子串为空", field)
		return
	}
	found := false
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("%s 应包含 [%s], 实际: [%s]", field, substr, str)
	}
}
