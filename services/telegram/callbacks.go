package telegram

import (
	"fmt"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/utils"
	"time"
)

// HandleCallbackQuery 处理回调查询
func HandleCallbackQuery(bot *TelegramBot, callback *CallbackQuery) error {
	data := callback.Data
	parts := strings.SplitN(data, ":", 2)
	action := parts[0]
	param := ""
	if len(parts) > 1 {
		param = parts[1]
	}

	utils.Debug("处理回调: action=%s, param=%s", action, param)

	switch action {
	// 导航回调
	case "start":
		return handleStartCallback(bot, callback)
	case "help":
		return handleHelpCallback(bot, callback)
	case "stats":
		return handleStatsCallback(bot, callback)
	case "monitor":
		return handleMonitorCallback(bot, callback)
	case "nodes":
		return handleNodesCallback(bot, callback)
	case "tasks":
		return handleTasksCallback(bot, callback)
	case "subscriptions":
		return handleSubscriptionsCallback(bot, callback)
	case "tags":
		return handleTagsCallback(bot, callback, param)
	case "airports":
		return handleAirportsCallback(bot, callback)
	case "airport_detail":
		return handleAirportDetailCallback(bot, callback, param)
	case "cancel":
		return handleCancelCallback(bot, callback)

	// 操作回调
	case "speedtest":
		return handleSpeedTestCallback(bot, callback, param)
	case "sub_pull":
		return handleSubPullCallback(bot, callback, param)
	case "airport_pull":
		return handleAirportPullCallback(bot, callback, param)
	case "sub_link":
		return handleSubLinkCallback(bot, callback, param)
	case "task_cancel":
		return handleTaskCancelCallback(bot, callback, param)

	default:
		utils.Debug("未知回调: %s", data)
		return nil
	}
}

// handleStartCallback 处理 start 回调
func handleStartCallback(bot *TelegramBot, callback *CallbackQuery) error {
	handler := GetHandler("start")
	if handler == nil {
		return nil
	}
	return handler.Handle(bot, callback.Message)
}

// handleHelpCallback 处理 help 回调
func handleHelpCallback(bot *TelegramBot, callback *CallbackQuery) error {
	handler := GetHandler("help")
	if handler == nil {
		return nil
	}
	return handler.Handle(bot, callback.Message)
}

// handleStatsCallback 处理 stats 回调
func handleStatsCallback(bot *TelegramBot, callback *CallbackQuery) error {
	handler := GetHandler("stats")
	if handler == nil {
		return nil
	}
	return handler.Handle(bot, callback.Message)
}

// handleMonitorCallback 处理 monitor 回调
func handleMonitorCallback(bot *TelegramBot, callback *CallbackQuery) error {
	handler := GetHandler("monitor")
	if handler == nil {
		return nil
	}
	return handler.Handle(bot, callback.Message)
}

// handleNodesCallback 处理 nodes 回调
func handleNodesCallback(bot *TelegramBot, callback *CallbackQuery) error {
	handler := GetHandler("nodes")
	if handler == nil {
		return nil
	}
	return handler.Handle(bot, callback.Message)
}

// handleTasksCallback 处理 tasks 回调
func handleTasksCallback(bot *TelegramBot, callback *CallbackQuery) error {
	handler := GetHandler("tasks")
	if handler == nil {
		return nil
	}
	return handler.Handle(bot, callback.Message)
}

// handleSubscriptionsCallback 处理 subscriptions 回调
func handleSubscriptionsCallback(bot *TelegramBot, callback *CallbackQuery) error {
	handler := GetHandler("subscriptions")
	if handler == nil {
		return nil
	}
	return handler.Handle(bot, callback.Message)
}

// handleTagsCallback 处理 tags 回调
func handleTagsCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	if param == "apply_all" {
		if err := ApplyAllTagRules(); err != nil {
			return bot.SendMessage(callback.Message.Chat.ID, "❌ 执行标签规则失败: "+err.Error(), "")
		}
		return bot.SendMessage(callback.Message.Chat.ID, "✅ 已开始执行标签规则", "")
	}

	handler := GetHandler("tags")
	if handler == nil {
		return nil
	}
	return handler.Handle(bot, callback.Message)
}

// handleCancelCallback 处理取消回调
func handleCancelCallback(bot *TelegramBot, callback *CallbackQuery) error {
	return bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID, "✅ 已取消", "", nil)
}

// handleSpeedTestCallback 处理测速回调
func handleSpeedTestCallback(bot *TelegramBot, callback *CallbackQuery, scope string) error {
	if scope == "" {
		handler := GetHandler("speedtest")
		if handler == nil {
			return nil
		}
		return handler.Handle(bot, callback.Message)
	}

	if err := RunSpeedTest(scope); err != nil {
		return bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
			"❌ 启动测速失败: "+err.Error(), "", nil)
	}

	scopeText := "定时测速配置"
	if scope == "untested" {
		scopeText = "未测速节点"
	}

	return bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
		fmt.Sprintf("✅ 已开始测速 (%s)\n\n测速完成后将会收到通知", scopeText), "", nil)
}

// handleSubPullCallback 处理订阅拉取回调
func handleSubPullCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	subID, err := strconv.Atoi(param)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ 无效的订阅 ID", "")
	}

	if err := PullSubscription(subID); err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ 更新订阅失败: "+err.Error(), "")
	}

	return bot.SendMessage(callback.Message.Chat.ID, "✅ 已开始更新订阅，完成后将会收到通知", "")
}

// handleTaskCancelCallback 处理任务取消回调
func handleTaskCancelCallback(bot *TelegramBot, callback *CallbackQuery, taskID string) error {
	if err := CancelTask(taskID); err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ 取消任务失败: "+err.Error(), "")
	}

	return bot.SendMessage(callback.Message.Chat.ID, "✅ 已发送取消请求", "")
}

// handleSubLinkCallback 处理订阅链接回调
func handleSubLinkCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	subID, err := strconv.Atoi(param)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ 无效的订阅 ID", "")
	}

	link, err := GetSubscriptionLink(subID)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ 获取链接失败: "+err.Error(), "")
	}

	// 生成不同格式的链接
	// 注意: link 已经包含 ?token=...，所以后续参数使用 &
	linkAuto := link
	linkV2Ray := link + "&client=v2ray"
	linkClash := link + "&client=clash"
	linkSurge := link + "&client=surge"

	// 发送链接（方便复制）
	text := fmt.Sprintf(`📎 *订阅链接*

🤖 *自动识别*
`+"`%s`"+`

🚀 *v2ray*
`+"`%s`"+`

🐱 *clash*
`+"`%s`"+`

⚡ *surge*
`+"`%s`"+`

点击链接可复制`, linkAuto, linkV2Ray, linkClash, linkSurge)

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("🔙 返回订阅列表", "subscriptions")},
	}

	return bot.SendMessageWithKeyboard(callback.Message.Chat.ID, text, "Markdown", keyboard)
}

// handleAirportsCallback 处理 airports 回调
func handleAirportsCallback(bot *TelegramBot, callback *CallbackQuery) error {
	handler := GetHandler("airports")
	if handler == nil {
		return nil
	}
	return handler.Handle(bot, callback.Message)
}

// handleAirportDetailCallback 处理 airport_detail 回调
func handleAirportDetailCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	id, err := strconv.Atoi(param)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ 无效的机场 ID", "")
	}

	airport, err := models.GetAirportByID(id)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ 获取机场失败: "+err.Error(), "")
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("✈️ *机场详情: %s*\n\n", airport.Name))

	// 基础信息
	text.WriteString(fmt.Sprintf("🔗 地址: `%s`\n", airport.URL))
	text.WriteString(fmt.Sprintf("📂 分组: `%s`\n", airport.Group))
	text.WriteString(fmt.Sprintf("⏰ 定时: `%s`\n", airport.CronExpr))

	status := "启用"
	if !airport.Enabled {
		status = "禁用"
	}
	text.WriteString(fmt.Sprintf("🔌 状态: %s\n", status))

	proxyStatus := "否"
	if airport.DownloadWithProxy {
		proxyStatus = "是"
		if airport.ProxyLink != "" {
			proxyStatus += " (指定)"
		} else {
			proxyStatus += " (自动)"
		}
	}
	text.WriteString(fmt.Sprintf("🌐 代理下载: %s\n", proxyStatus))

	if airport.UserAgent != "" {
		text.WriteString(fmt.Sprintf("🕵️ UA: `%s`\n", airport.UserAgent))
	}

	if airport.LastRunTime != nil {
		text.WriteString(fmt.Sprintf("🕒 上次更新: %s\n", airport.LastRunTime.Format("2006-01-02 15:04:05")))
	}

	// 用量信息
	if airport.FetchUsageInfo {
		text.WriteString("\n📊 *用量信息*\n")
		// 注意: 这里假设 models.Airport 结构体中有用量字段，这在之前的文件查看中已确认
		if airport.UsageTotal > 0 {
			text.WriteString(fmt.Sprintf("⬆️ 上传: %s\n", utils.FormatBytes(airport.UsageUpload)))
			text.WriteString(fmt.Sprintf("⬇️ 下载: %s\n", utils.FormatBytes(airport.UsageDownload)))
			text.WriteString(fmt.Sprintf("📦 总量: %s\n", utils.FormatBytes(airport.UsageTotal)))
			if airport.UsageExpire > 0 {
				text.WriteString(fmt.Sprintf("⏳ 过期: %s\n", time.Unix(airport.UsageExpire, 0).Format("2006-01-02 15:04:05")))
			}
		} else if airport.UsageTotal == -1 {
			text.WriteString("⚠️ 获取失败或不支持\n")
		} else {
			text.WriteString("⏳ 暂无数据\n")
		}
	}

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("🔄 立即更新", fmt.Sprintf("airport_pull:%d", id))},
		{NewInlineButton("🔙 返回列表", "airports")},
	}

	return bot.SendMessageWithKeyboard(callback.Message.Chat.ID, text.String(), "Markdown", keyboard)
}

// handleAirportPullCallback 处理 airport_pull 回调
func handleAirportPullCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	id, err := strconv.Atoi(param)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ 无效的机场 ID", "")
	}

	// 复用 PullSubscription 函数
	if err := PullSubscription(id); err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ 启动更新失败: "+err.Error(), "")
	}

	return bot.SendMessage(callback.Message.Chat.ID, "✅ 已开始更新任务，完成后将收到通知", "")
}
