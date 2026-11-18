package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"sublink/services/notifications"
	"sublink/utils"
)

// escapeMd 转义 Telegram Markdown 特殊字符，避免用户内容中的特殊字符被错误解析
func escapeMd(text string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`_`, `\_`,
		`*`, `\*`,
		"`", "\\`",
		`[`, `\[`,
		`]`, `\]`,
	)
	return r.Replace(text)
}

// SendNotification 发送通知到 Telegram
func SendNotification(eventKey string, payload notifications.Payload) {
	bot := GetBot()
	if bot == nil || !bot.IsConnected() {
		return
	}

	// 构建通知消息
	var text string

	switch eventKey {
	case "task.speed_test_completed":
		text = formatSpeedTestNotification(payload)
	case "subscription.sync_succeeded", "subscription.sync_failed":
		text = formatSubUpdateNotification(payload)
	case "task.tag_rule_completed":
		text = formatTagRuleNotification(payload)
	case "task.auto_tag_completed":
		text = formatAutoTagNotification(payload)
	case "task_complete":
		text = formatTaskCompleteNotification(payload)
	case "task_error":
		text = formatTaskErrorNotification(payload)
	default:
		// 通用格式
		text = formatGenericNotification(eventKey, payload)
	}

	if text == "" {
		return
	}

	if err := bot.SendMessage(bot.ChatID, text, "Markdown"); err != nil {
		utils.Warn("Failed to send Telegram notification: %v", err)
	}
}

// formatSpeedTestNotification 格式化测速完成通知
func formatSpeedTestNotification(payload notifications.Payload) string {
	data, ok := payload.Data.(map[string]any)
	if !ok {
		return fmt.Sprintf("⚡ *Speed test completed*\n\n%s", escapeMd(payload.Message))
	}

	successCount := getIntFromData(data, "success_count")
	if successCount == 0 {
		successCount = getIntFromData(data, "success")
	}
	failCount := getIntFromData(data, "fail_count")
	if failCount == 0 {
		failCount = getIntFromData(data, "fail")
	}
	totalTraffic := getFloatFromData(data, "total_traffic_mb")

	return fmt.Sprintf(`⚡ *Speed test completed*

%s

*Results*
├ ✅ Success: %d
├ ❌ Failed: %d
└ 📊 Traffic: %.2f MB`, escapeMd(payload.Message), successCount, failCount, totalTraffic)
}

// formatSubUpdateNotification 格式化订阅更新通知
func formatSubUpdateNotification(payload notifications.Payload) string {
	data, ok := payload.Data.(map[string]any)
	if !ok {
		return fmt.Sprintf("📋 *Subscription update*\n\n%s", escapeMd(payload.Message))
	}

	status := getStringFromData(data, "status")
	name := getStringFromData(data, "name")

	icon := "📋"
	switch status {
	case "error":
		icon = "❌"
	case "success":
		icon = "✅"
	}

	return fmt.Sprintf(`%s *Subscription update*

*Subscription*: %s
%s`, icon, escapeMd(name), escapeMd(payload.Message))
}

// formatTagRuleNotification 格式化标签规则通知
func formatTagRuleNotification(payload notifications.Payload) string {
	return fmt.Sprintf("🏷️ *Tag rule completed*\n\n%s", escapeMd(payload.Message))
}

func formatAutoTagNotification(payload notifications.Payload) string {
	return fmt.Sprintf("🏷️ *Auto tagging completed*\n\n%s", escapeMd(payload.Message))
}

// formatTaskCompleteNotification 格式化任务完成通知
func formatTaskCompleteNotification(payload notifications.Payload) string {
	return fmt.Sprintf("✅ *Task completed*\n\n*%s*\n%s", escapeMd(payload.Title), escapeMd(payload.Message))
}

// formatTaskErrorNotification 格式化任务错误通知
func formatTaskErrorNotification(payload notifications.Payload) string {
	return fmt.Sprintf("❌ *Task failed*\n\n*%s*\n%s", escapeMd(payload.Title), escapeMd(payload.Message))
}

// formatGenericNotification 格式化通用通知
func formatGenericNotification(event string, payload notifications.Payload) string {
	if payload.Title == "" && payload.Message == "" {
		return ""
	}

	if payload.Title != "" {
		return fmt.Sprintf("🔔 *%s*\n\n%s", escapeMd(payload.Title), escapeMd(payload.Message))
	}

	return fmt.Sprintf("🔔 %s", escapeMd(payload.Message))
}

// Helper functions

func getIntFromData(data map[string]any, key string) int {
	if v, ok := data[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int32:
			return int(val)
		case int64:
			return int(val)
		case uint:
			return int(val)
		case uint32:
			return int(val)
		case uint64:
			return int(val)
		case float32:
			return int(val)
		case float64:
			return int(val)
		case string:
			if parsed, err := strconv.Atoi(val); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func getFloatFromData(data map[string]any, key string) float64 {
	if v, ok := data[key]; ok {
		switch val := v.(type) {
		case float32:
			return float64(val)
		case float64:
			return val
		case int:
			return float64(val)
		case int32:
			return float64(val)
		case int64:
			return float64(val)
		case uint:
			return float64(val)
		case uint32:
			return float64(val)
		case uint64:
			return float64(val)
		case string:
			if parsed, err := strconv.ParseFloat(val, 64); err == nil {
				return parsed
			}
		}
	}
	return 0
}

func getStringFromData(data map[string]any, key string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
