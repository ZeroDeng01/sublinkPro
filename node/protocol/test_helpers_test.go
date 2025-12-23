package protocol

import "testing"

// 测试辅助函数 - 用于验证编解码结果

// assertEqualString 验证两个字符串相等
func assertEqualString(t *testing.T, field string, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s 不匹配: 期望 [%s], 实际 [%s]", field, expected, actual)
	}
}

// assertEqualInt 验证两个整数相等
func assertEqualInt(t *testing.T, field string, expected int, actual FlexPort) {
	t.Helper()
	if expected != int(actual) {
		t.Errorf("%s 不匹配: 期望 %d, 实际 %d", field, expected, int(actual))
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
