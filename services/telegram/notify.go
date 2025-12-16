package telegram

import (
	"fmt"
	"sublink/services/sse"
	"sublink/utils"
)

// SendNotification 发送通知到 Telegram
// 在 SSE BroadcastEvent 时调用
func SendNotification(event string, payload sse.NotificationPayload) {
	bot := GetBot()
	if bot == nil || !bot.IsConnected() {
		return
	}

	// 构建通知消息
	var text string

	switch event {
	case "speed_test_complete":
		text = formatSpeedTestNotification(payload)
	case "sub_update":
		text = formatSubUpdateNotification(payload)
	case "tag_rule_applied":
		text = formatTagRuleNotification(payload)
	case "task_complete":
		text = formatTaskCompleteNotification(payload)
	case "task_error":
		text = formatTaskErrorNotification(payload)
	default:
		// 通用格式
		text = formatGenericNotification(event, payload)
	}

	if text == "" {
		return
	}

	if err := bot.SendMessage(bot.ChatID, text, "Markdown"); err != nil {
		utils.Warn("发送 Telegram 通知失败: %v", err)
	}
}

// formatSpeedTestNotification 格式化测速完成通知
func formatSpeedTestNotification(payload sse.NotificationPayload) string {
	data, ok := payload.Data.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("⚡ *测速完成*\n\n%s", payload.Message)
	}

	successCount := getIntFromData(data, "success_count")
	failCount := getIntFromData(data, "fail_count")
	totalTraffic := getFloatFromData(data, "total_traffic_mb")

	return fmt.Sprintf(`⚡ *测速任务完成*

%s

*结果统计*
├ ✅ 成功: %d
├ ❌ 失败: %d
└ 📊 流量: %.2f MB`, payload.Message, successCount, failCount, totalTraffic)
}

// formatSubUpdateNotification 格式化订阅更新通知
func formatSubUpdateNotification(payload sse.NotificationPayload) string {
	data, ok := payload.Data.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("📋 *订阅更新*\n\n%s", payload.Message)
	}

	status := getStringFromData(data, "status")
	name := getStringFromData(data, "name")

	icon := "📋"
	if status == "error" {
		icon = "❌"
	} else if status == "success" {
		icon = "✅"
	}

	return fmt.Sprintf(`%s *订阅更新*

*订阅*: %s
%s`, icon, name, payload.Message)
}

// formatTagRuleNotification 格式化标签规则通知
func formatTagRuleNotification(payload sse.NotificationPayload) string {
	return fmt.Sprintf("🏷️ *标签规则执行完成*\n\n%s", payload.Message)
}

// formatTaskCompleteNotification 格式化任务完成通知
func formatTaskCompleteNotification(payload sse.NotificationPayload) string {
	return fmt.Sprintf("✅ *任务完成*\n\n*%s*\n%s", payload.Title, payload.Message)
}

// formatTaskErrorNotification 格式化任务错误通知
func formatTaskErrorNotification(payload sse.NotificationPayload) string {
	return fmt.Sprintf("❌ *任务失败*\n\n*%s*\n%s", payload.Title, payload.Message)
}

// formatGenericNotification 格式化通用通知
func formatGenericNotification(event string, payload sse.NotificationPayload) string {
	if payload.Title == "" && payload.Message == "" {
		return ""
	}

	if payload.Title != "" {
		return fmt.Sprintf("🔔 *%s*\n\n%s", payload.Title, payload.Message)
	}

	return fmt.Sprintf("🔔 %s", payload.Message)
}

// Helper functions

func getIntFromData(data map[string]interface{}, key string) int {
	if v, ok := data[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

func getFloatFromData(data map[string]interface{}, key string) float64 {
	if v, ok := data[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0
}

func getStringFromData(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
