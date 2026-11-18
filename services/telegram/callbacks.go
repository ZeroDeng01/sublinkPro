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

	utils.Debug("Handling callback: action=%s, param=%s", action, param)

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
	case "subscriptions_page":
		return handleSubscriptionsPageCallback(bot, callback, param)
	case "tags":
		return handleTagsCallback(bot, callback, param)
	case "tags_page":
		return handleTagsPageCallback(bot, callback, param)
	case "tag_run":
		return handleTagRunCallback(bot, callback, param)
	case "airports":
		return handleAirportsCallback(bot, callback)
	case "airports_page":
		return handleAirportsPageCallback(bot, callback, param)
	case "airport_detail":
		return handleAirportDetailCallback(bot, callback, param)
	case "cancel":
		return handleCancelCallback(bot, callback)

	// 检测策略相关回调
	case "profiles":
		return handleProfilesCallback(bot, callback)
	case "profiles_page":
		return handleProfilesPageCallback(bot, callback, param)
	case "profile_detail":
		return handleProfileDetailCallback(bot, callback, param)
	case "profile_run":
		return handleProfileRunCallback(bot, callback, param)
	case "profile_toggle":
		return handleProfileToggleCallback(bot, callback, param)
	case "profile_select_untested":
		return handleProfileSelectUntestedCallback(bot, callback)
	case "profile_run_untested":
		return handleProfileRunUntestedCallback(bot, callback, param)

	// 其他操作回调
	case "sub_link":
		return handleSubLinkCallback(bot, callback, param)
	case "airport_pull":
		return handleAirportPullCallback(bot, callback, param)
	case "task_cancel":
		return handleTaskCancelCallback(bot, callback, param)

	default:
		utils.Debug("Unknown callback: %s", data)
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

// handleSubscriptionsPageCallback 处理订阅分页回调
func handleSubscriptionsPageCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	page, err := strconv.Atoi(param)
	if err != nil {
		page = 0
	}

	handler := GetHandler("subscriptions")
	if handler == nil {
		return nil
	}

	if subsHandler, ok := handler.(*SubscriptionsHandler); ok {
		return subsHandler.HandleWithPage(bot, callback.Message, page)
	}
	return handler.Handle(bot, callback.Message)
}

// handleTagsCallback 处理 tags 回调
func handleTagsCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	if param == "apply_all" {
		if err := ApplyAllTagRules(); err != nil {
			return bot.SendMessage(callback.Message.Chat.ID, "❌ Failed to run tag rule: "+err.Error(), "")
		}
		return bot.SendMessage(callback.Message.Chat.ID, "✅ Tag rule execution started", "")
	}

	handler := GetHandler("tags")
	if handler == nil {
		return nil
	}
	return handler.Handle(bot, callback.Message)
}

// handleTagsPageCallback 处理标签规则分页回调
func handleTagsPageCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	page, err := strconv.Atoi(param)
	if err != nil {
		page = 0
	}

	handler := GetHandler("tags")
	if handler == nil {
		return nil
	}

	if tagsHandler, ok := handler.(*TagsHandler); ok {
		return tagsHandler.HandleWithPage(bot, callback.Message, page)
	}
	return handler.Handle(bot, callback.Message)
}

// handleTagRunCallback 处理执行单个标签规则回调
func handleTagRunCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	ruleID, err := strconv.Atoi(param)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Invalid rule ID", "")
	}

	// 获取规则信息以显示名称
	var rule models.TagRule
	if err := rule.GetByID(ruleID); err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Rule not found", "")
	}

	// 执行规则
	if err := TriggerTagRule(ruleID); err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Failed to run rule: "+err.Error(), "")
	}

	text := fmt.Sprintf("✅ Tag rule execution started\n\n📋 Rule: *%s*\n🏷️ Tag: *%s*\n\nYou will be notified when it finishes", rule.Name, rule.TagName)
	return bot.SendMessage(callback.Message.Chat.ID, text, "Markdown")
}

// handleCancelCallback 处理取消回调
func handleCancelCallback(bot *TelegramBot, callback *CallbackQuery) error {
	return bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID, "✅ Cancelled", "", nil)
}

// ============ 检测策略相关回调 ============

// handleProfilesCallback 处理 profiles 回调
func handleProfilesCallback(bot *TelegramBot, callback *CallbackQuery) error {
	handler := GetHandler("profiles")
	if handler == nil {
		return nil
	}
	return handler.Handle(bot, callback.Message)
}

// handleProfilesPageCallback 处理策略分页回调
func handleProfilesPageCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	page, err := strconv.Atoi(param)
	if err != nil {
		page = 0
	}

	handler := GetHandler("profiles")
	if handler == nil {
		return nil
	}

	if profilesHandler, ok := handler.(*ProfilesHandler); ok {
		return profilesHandler.HandleWithPage(bot, callback.Message, page)
	}
	return handler.Handle(bot, callback.Message)
}

// handleProfileDetailCallback 处理策略详情回调
func handleProfileDetailCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	id, err := strconv.Atoi(param)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Invalid profile ID", "")
	}

	profile, err := models.GetNodeCheckProfileByID(id)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Profile not found", "")
	}

	var text strings.Builder
	fmt.Fprintf(&text, "⚡ *Profile details: %s*\n\n", profile.Name)

	// 基本信息
	status := "❌ Disabled"
	if profile.Enabled {
		status = "✅ Enabled"
	}
	fmt.Fprintf(&text, "🔌 Schedule status: %s\n", status)

	if profile.CronExpr != "" {
		fmt.Fprintf(&text, "⏰ Schedule: `%s`\n", profile.CronExpr)
	}

	// 模式配置（与Web端保持一致）
	mode := "Latency test only"
	if profile.Mode == "mihomo" {
		mode = "Latency + speed test"
	}
	fmt.Fprintf(&text, "📡 Mode: %s\n", mode)
	fmt.Fprintf(&text, "⏱️ Timeout: %d seconds\n", profile.Timeout)

	// URL配置
	if profile.TestURL != "" {
		fmt.Fprintf(&text, "🔗 Speed test URL: `%s`\n", truncateName(profile.TestURL, 35))
	}
	if profile.LatencyURL != "" {
		fmt.Fprintf(&text, "🔗 Latency URL: `%s`\n", truncateName(profile.LatencyURL, 35))
	}

	// 并发配置
	text.WriteString("\n*Concurrency*\n")
	latencyC := "Auto"
	if profile.LatencyConcurrency > 0 {
		latencyC = fmt.Sprintf("%d", profile.LatencyConcurrency)
	}
	fmt.Fprintf(&text, "├ Latency concurrency: %s\n", latencyC)
	fmt.Fprintf(&text, "└ Speed concurrency: %d\n", profile.SpeedConcurrency)

	// 范围过滤
	groups := profile.GetGroups()
	tags := profile.GetTags()
	if len(groups) > 0 || len(tags) > 0 {
		text.WriteString("\n*Check scope*\n")
		if len(groups) > 0 {
			fmt.Fprintf(&text, "├ Group: %s\n", strings.Join(groups, ", "))
		}
		if len(tags) > 0 {
			fmt.Fprintf(&text, "└ Tag: %s\n", strings.Join(tags, ", "))
		}
	} else {
		text.WriteString("\n*Check scope*: All nodes\n")
	}

	// 执行时间
	if profile.LastRunTime != nil {
		fmt.Fprintf(&text, "\n🕒 Last run: %s\n", profile.LastRunTime.Format("2006-01-02 15:04:05"))
	}
	if profile.NextRunTime != nil {
		fmt.Fprintf(&text, "⏳ Next run: %s\n", profile.NextRunTime.Format("2006-01-02 15:04:05"))
	}

	// 操作按钮
	toggleText := "✅ Enable schedule"
	if profile.Enabled {
		toggleText = "⏸️ Disable schedule"
	}

	keyboard := [][]InlineKeyboardButton{
		{
			NewInlineButton("▶️ Run now", fmt.Sprintf("profile_run:%d", id)),
			NewInlineButton(toggleText, fmt.Sprintf("profile_toggle:%d", id)),
		},
		{NewInlineButton("🔙 Back to list", "profiles")},
	}

	return bot.SendMessageWithKeyboard(callback.Message.Chat.ID, text.String(), "Markdown", keyboard)
}

// handleProfileRunCallback 处理策略执行回调
func handleProfileRunCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	id, err := strconv.Atoi(param)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Invalid profile ID", "")
	}

	profile, err := models.GetNodeCheckProfileByID(id)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Profile not found", "")
	}

	if err := ExecuteNodeCheckWithProfile(id, nil, models.TaskTriggerManual); err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Failed to start check: "+err.Error(), "")
	}

	text := fmt.Sprintf("✅ Check task started\n\n📋 Profile: *%s*\n\nYou will be notified when it finishes", profile.Name)
	return bot.SendMessage(callback.Message.Chat.ID, text, "Markdown")
}

// handleProfileToggleCallback 处理策略开关回调
func handleProfileToggleCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	id, err := strconv.Atoi(param)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Invalid profile ID", "")
	}

	newEnabled, err := ToggleProfileEnabled(id)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Operation failed: "+err.Error(), "")
	}

	status := "Disabled"
	if newEnabled {
		status = "Enabled"
	}

	text := fmt.Sprintf("✅ Schedule %s", status)
	return bot.SendMessage(callback.Message.Chat.ID, text, "")
}

// handleProfileSelectUntestedCallback 处理选择策略检测未测速节点
func handleProfileSelectUntestedCallback(bot *TelegramBot, callback *CallbackQuery) error {
	profiles, err := GetNodeCheckProfiles()
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Failed to load profiles", "")
	}

	if len(profiles) == 0 {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ No check profiles yet; create one in the web UI first", "")
	}

	// 统计未测速节点
	var node models.Node
	nodes, _ := node.List()
	untestedCount := 0
	for _, n := range nodes {
		if n.DelayStatus == "" || n.DelayStatus == "untested" {
			untestedCount++
		}
	}

	if untestedCount == 0 {
		return bot.SendMessage(callback.Message.Chat.ID, "✅ All nodes already have speed test results", "")
	}

	var text strings.Builder
	fmt.Fprintf(&text, "🔍 *Choose a profile to test untested nodes*\n\nThere are *%d* untested nodes\n\nSelect a profile:", untestedCount)

	var keyboard [][]InlineKeyboardButton
	for _, p := range profiles {
		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton(p.Name, fmt.Sprintf("profile_run_untested:%d", p.ID)),
		})
	}
	keyboard = append(keyboard, []InlineKeyboardButton{
		NewInlineButton("🔙 Back", "profiles"),
	})

	return bot.SendMessageWithKeyboard(callback.Message.Chat.ID, text.String(), "Markdown", keyboard)
}

// handleProfileRunUntestedCallback 使用指定策略检测未测速节点
func handleProfileRunUntestedCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	profileID, err := strconv.Atoi(param)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Invalid profile ID", "")
	}

	profile, err := models.GetNodeCheckProfileByID(profileID)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Profile not found", "")
	}

	// 获取未测速节点ID
	var node models.Node
	nodes, _ := node.List()
	var untestedIDs []int
	for _, n := range nodes {
		if n.DelayStatus == "" || n.DelayStatus == "untested" {
			untestedIDs = append(untestedIDs, n.ID)
		}
	}

	if len(untestedIDs) == 0 {
		return bot.SendMessage(callback.Message.Chat.ID, "✅ All nodes already have speed test results", "")
	}

	// 执行检测
	if err := ExecuteNodeCheckWithProfile(profileID, untestedIDs, models.TaskTriggerManual); err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Failed to start check: "+err.Error(), "")
	}

	text := fmt.Sprintf("✅ Untested-node check started\n\n📋 Profile: *%s*\n📦 Nodes: *%d*\n\nYou will be notified when it finishes", profile.Name, len(untestedIDs))
	return bot.SendMessage(callback.Message.Chat.ID, text, "Markdown")
}

// ============ 其他回调 ============

// handleTaskCancelCallback 处理任务取消回调
func handleTaskCancelCallback(bot *TelegramBot, callback *CallbackQuery, taskID string) error {
	if err := CancelTask(taskID); err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Failed to cancel task: "+err.Error(), "")
	}

	return bot.SendMessage(callback.Message.Chat.ID, "✅ Cancellation requested", "")
}

// handleSubLinkCallback 处理订阅链接回调
func handleSubLinkCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	subID, err := strconv.Atoi(param)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Invalid subscription ID", "")
	}

	link, needHint, err := GetSubscriptionLink(subID)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Failed to get link: "+err.Error(), "")
	}

	// 生成不同格式的链接
	// 注意: link 已经包含 ?token=...，所以后续参数使用 &
	linkAuto := link
	linkV2Ray := link + "&client=v2ray"
	linkClash := link + "&client=clash"
	linkSurge := link + "&client=surge"

	// 构建消息
	var text strings.Builder
	text.WriteString("📎 *Subscription links*\n\n")

	// 如果未配置域名，添加提示
	if needHint {
		text.WriteString("⚠️ *Note*: no remote access domain is configured; the links use the local address and may not be reachable externally.\n")
		text.WriteString("Configure a remote access domain under \"User Center → Personal Settings\" in the web UI.\n\n")
	}

	text.WriteString("🤖 *Auto-detect*\n`" + linkAuto + "`\n\n")
	text.WriteString("🚀 *v2ray*\n`" + linkV2Ray + "`\n\n")
	text.WriteString("🐱 *clash*\n`" + linkClash + "`\n\n")
	text.WriteString("⚡ *surge*\n`" + linkSurge + "`\n\n")
	text.WriteString("Tap a link to copy")

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("🔙 Back to subscriptions", "subscriptions")},
	}

	return bot.SendMessageWithKeyboard(callback.Message.Chat.ID, text.String(), "Markdown", keyboard)
}

// handleAirportsCallback 处理 airports 回调
func handleAirportsCallback(bot *TelegramBot, callback *CallbackQuery) error {
	handler := GetHandler("airports")
	if handler == nil {
		return nil
	}
	return handler.Handle(bot, callback.Message)
}

// handleAirportsPageCallback 处理机场分页回调
func handleAirportsPageCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	page, err := strconv.Atoi(param)
	if err != nil {
		page = 0
	}

	handler := GetHandler("airports")
	if handler == nil {
		return nil
	}

	// 类型断言获取 AirportsHandler
	if airportsHandler, ok := handler.(*AirportsHandler); ok {
		return airportsHandler.HandleWithPage(bot, callback.Message, page)
	}
	return handler.Handle(bot, callback.Message)
}

// handleAirportDetailCallback 处理 airport_detail 回调
func handleAirportDetailCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	id, err := strconv.Atoi(param)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Invalid airport ID", "")
	}

	airport, err := models.GetAirportByID(id)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Failed to load airport: "+err.Error(), "")
	}

	var text strings.Builder
	fmt.Fprintf(&text, "✈️ *Airport details: %s*\n\n", airport.Name)

	// 基础信息
	fmt.Fprintf(&text, "🔗 URL: `%s`\n", airport.URL)
	fmt.Fprintf(&text, "📂 Group: `%s`\n", airport.Group)
	fmt.Fprintf(&text, "⏰ Schedule: `%s`\n", airport.CronExpr)

	status := "Enabled"
	if !airport.Enabled {
		status = "Disabled"
	}
	fmt.Fprintf(&text, "🔌 Status: %s\n", status)

	proxyStatus := "No"
	if airport.DownloadWithProxy {
		proxyStatus = "Yes"
		if airport.ProxyLink != "" {
			proxyStatus += " (specified)"
		} else {
			proxyStatus += " (auto)"
		}
	}
	fmt.Fprintf(&text, "🌐 Proxy download: %s\n", proxyStatus)

	if airport.UserAgent != "" {
		fmt.Fprintf(&text, "🕵️ UA: `%s`\n", airport.UserAgent)
	}

	if airport.LastRunTime != nil {
		fmt.Fprintf(&text, "🕒 Last updated: %s\n", airport.LastRunTime.Format("2006-01-02 15:04:05"))
	}

	// 用量信息
	if airport.FetchUsageInfo {
		text.WriteString("\n📊 *Usage*\n")
		// 注意: 这里假设 models.Airport 结构体中有用量字段，这在之前的文件查看中已确认
		if airport.UsageTotal > 0 {
			fmt.Fprintf(&text, "⬆️ Upload: %s\n", utils.FormatBytes(airport.UsageUpload))
			fmt.Fprintf(&text, "⬇️ Download: %s\n", utils.FormatBytes(airport.UsageDownload))
			fmt.Fprintf(&text, "📦 Total: %s\n", utils.FormatBytes(airport.UsageTotal))
			if airport.UsageExpire > 0 {
				fmt.Fprintf(&text, "⏳ Expires: %s\n", time.Unix(airport.UsageExpire, 0).Format("2006-01-02 15:04:05"))
			}
		} else if airport.UsageTotal == -1 {
			text.WriteString("⚠️ Unavailable or unsupported\n")
		} else {
			text.WriteString("⏳ No data yet\n")
		}
	}

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("🔄 Update now", fmt.Sprintf("airport_pull:%d", id))},
		{NewInlineButton("🔙 Back to list", "airports")},
	}

	return bot.SendMessageWithKeyboard(callback.Message.Chat.ID, text.String(), "Markdown", keyboard)
}

// handleAirportPullCallback 处理 airport_pull 回调
func handleAirportPullCallback(bot *TelegramBot, callback *CallbackQuery, param string) error {
	id, err := strconv.Atoi(param)
	if err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Invalid airport ID", "")
	}

	// 复用 PullSubscription 函数
	if err := PullSubscription(id); err != nil {
		return bot.SendMessage(callback.Message.Chat.ID, "❌ Failed to start update: "+err.Error(), "")
	}

	return bot.SendMessage(callback.Message.Chat.ID, "✅ Update task started; you will be notified when it finishes", "")
}
