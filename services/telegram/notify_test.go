package telegram

import (
	"strings"
	"sublink/services/notifications"
	"testing"
)

func TestFormatSpeedTestNotificationIncludesCountsAndTraffic(t *testing.T) {
	message := formatSpeedTestNotification(notifications.Payload{
		Message: "测速已完成",
		Data: map[string]interface{}{
			"success_count":    3,
			"fail_count":       1,
			"total_traffic_mb": 12.5,
		},
	})

	if !strings.Contains(message, "成功: 3") {
		t.Fatalf("expected success count in message, got %s", message)
	}
	if !strings.Contains(message, "失败: 1") {
		t.Fatalf("expected fail count in message, got %s", message)
	}
	if !strings.Contains(message, "流量: 12.50 MB") {
		t.Fatalf("expected traffic amount in message, got %s", message)
	}
}

func TestFormatSubUpdateNotificationUsesStatusIcon(t *testing.T) {
	message := formatSubUpdateNotification(notifications.Payload{
		Message: "同步失败",
		Data: map[string]interface{}{
			"name":   "测试订阅",
			"status": "error",
		},
	})

	if !strings.HasPrefix(message, "❌") {
		t.Fatalf("expected error icon prefix, got %s", message)
	}
	if !strings.Contains(message, "测试订阅") {
		t.Fatalf("expected subscription name in message, got %s", message)
	}
}

func TestFormatGenericNotificationReturnsEmptyWhenNoContent(t *testing.T) {
	message := formatGenericNotification("security.user_login", notifications.Payload{})
	if message != "" {
		t.Fatalf("expected empty message, got %s", message)
	}
}
