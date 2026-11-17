package telegram

import (
	"fmt"
	"sort"
	"strings"
	"sublink/models"
	"sublink/services/monitor"
	"sublink/utils"
	"sync"
	"time"
)

// CommandHandler 命令处理器接口
type CommandHandler interface {
	Command() string
	Description() string
	Handle(bot *TelegramBot, message *Message) error
}

// 命令处理器注册表
var (
	handlers     = make(map[string]CommandHandler)
	handlerMutex sync.RWMutex
)

// RegisterHandler 注册命令处理器
func RegisterHandler(cmd string, handler CommandHandler) {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()
	handlers[cmd] = handler
}

// GetHandler 获取命令处理器
func GetHandler(cmd string) CommandHandler {
	handlerMutex.RLock()
	defer handlerMutex.RUnlock()
	return handlers[cmd]
}

// GetAllHandlers 获取所有处理器
func GetAllHandlers() map[string]CommandHandler {
	handlerMutex.RLock()
	defer handlerMutex.RUnlock()
	result := make(map[string]CommandHandler)
	for k, v := range handlers {
		result[k] = v
	}
	return result
}

func init() {
	// 注册所有命令处理器
	RegisterHandler("start", &StartHandler{})
	RegisterHandler("help", &HelpHandler{})
	RegisterHandler("stats", &StatsHandler{})
	RegisterHandler("monitor", &MonitorHandler{})
	RegisterHandler("profiles", &ProfilesHandler{})
	RegisterHandler("subscriptions", &SubscriptionsHandler{})
	RegisterHandler("nodes", &NodesHandler{})
	RegisterHandler("tags", &TagsHandler{})
	RegisterHandler("tasks", &TasksHandler{})
	RegisterHandler("airports", &AirportsHandler{})
}

// ============ StartHandler ============

type StartHandler struct{}

func (h *StartHandler) Command() string     { return "start" }
func (h *StartHandler) Description() string { return "🚀 Get started" }

func (h *StartHandler) Handle(bot *TelegramBot, message *Message) error {
	text := `🚀 *欢迎使用 Sublink Pro 机器人*

您可以通过此机器人远程管理您的 Sublink Pro 系统。

*可用功能：*
• 📊 查看仪表盘统计数据
• 🖥️ 查看系统监控信息
• ⚡ 节点检测策略管理
• 📋 管理订阅和节点
• 🏷️ 执行标签规则
• 📝 查看和管理任务

使用 /help 查看详细命令列表`

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("📊 Stats", "stats"), NewInlineButton("🖥️ Monitor", "monitor")},
		{NewInlineButton("⚡ Check profiles", "profiles"), NewInlineButton("📋 Subscriptions", "subscriptions")},
		{NewInlineButton("❓ Help", "help")},
	}

	return bot.SendMessageWithKeyboard(message.Chat.ID, text, "Markdown", keyboard)
}

// ============ HelpHandler ============

type HelpHandler struct{}

func (h *HelpHandler) Command() string     { return "help" }
func (h *HelpHandler) Description() string { return "❓ Help" }

func (h *HelpHandler) Handle(bot *TelegramBot, message *Message) error {
	text := `❓ *命令帮助*

/start - 🚀 开始使用
/help - ❓ 帮助信息
/stats - 📊 仪表盘统计
/monitor - 🖥️ 系统监控
/profiles - ⚡ 检测策略
/subscriptions - 📋 订阅管理
/nodes - 🌐 节点信息
/tags - 🏷️ 标签规则
/tasks - 📝 任务管理

💡 *提示*：您也可以点击消息中的按钮进行快捷操作`

	return bot.SendMessage(message.Chat.ID, text, "Markdown")
}

// ============ StatsHandler ============

type StatsHandler struct{}

func (h *StatsHandler) Command() string     { return "stats" }
func (h *StatsHandler) Description() string { return "📊 Dashboard stats" }

func (h *StatsHandler) Handle(bot *TelegramBot, message *Message) error {
	// 获取节点统计（与 Web 端 NodesTotal API 完全一致）
	var node models.Node
	nodes, _ := node.List()
	total := len(nodes)

	// 可用节点：Speed > 0 且 DelayTime > 0（与 Web 端定义一致）
	available := 0
	for _, n := range nodes {
		if n.Speed > 0 && n.DelayTime > 0 {
			available++
		}
	}

	// 获取订阅数量
	var sub models.Subcription
	subs, _ := sub.List()
	subCount := len(subs)

	// 获取最快速度节点和最低延迟节点
	fastestNode := models.GetFastestSpeedNode()
	lowestDelayNode := models.GetLowestDelayNode()

	// 获取统计数据
	countryStats := models.GetNodeCountryStats()
	protocolStats := models.GetNodeProtocolStats()

	// 构建消息
	var text strings.Builder
	text.WriteString("📊 *Dashboard stats*\n\n")

	// 基础统计
	fmt.Fprintf(&text, "📋 Subscriptions: *%d*\n", subCount)
	fmt.Fprintf(&text, "📦 Nodes: *%d* / %d\n\n", available, total)

	// 最快速度
	if fastestNode != nil && fastestNode.Speed > 0 {
		fmt.Fprintf(&text, "🚀 Top speed: *%.2f MB/s*\n", fastestNode.Speed)
		fmt.Fprintf(&text, "   └ %s\n\n", truncateName(fastestNode.Name, 25))
	}

	// 最低延迟
	if lowestDelayNode != nil && lowestDelayNode.DelayTime > 0 {
		fmt.Fprintf(&text, "⚡ Lowest latency: *%d ms*\n", lowestDelayNode.DelayTime)
		fmt.Fprintf(&text, "   └ %s\n\n", truncateName(lowestDelayNode.Name, 25))
	}

	// 国家分布
	if len(countryStats) > 0 {
		text.WriteString("🌍 *Countries*\n")
		sortedCountries := sortMapByValue(countryStats)
		for i, kv := range sortedCountries {
			prefix := "├"
			if i == len(sortedCountries)-1 {
				prefix = "└"
			}
			flag := getCountryFlag(kv.Key)
			fmt.Fprintf(&text, "%s %s %s: %d\n", prefix, flag, kv.Key, kv.Value)
		}
		text.WriteString("\n")
	}

	// 协议分布
	if len(protocolStats) > 0 {
		text.WriteString("📡 *Protocols*\n")
		sortedProtocols := sortMapByValue(protocolStats)
		for i, kv := range sortedProtocols {
			prefix := "├"
			if i == len(sortedProtocols)-1 {
				prefix = "└"
			}
			fmt.Fprintf(&text, "%s %s: %d\n", prefix, kv.Key, kv.Value)
		}
		text.WriteString("\n")
	}

	// 标签分布
	tagStats := models.GetNodeTagStats()
	if len(tagStats) > 0 {
		text.WriteString("🏷️ *Tags*\n")
		// 排序标签统计
		sort.Slice(tagStats, func(i, j int) bool {
			return tagStats[i].Count > tagStats[j].Count
		})

		for i, ts := range tagStats {
			prefix := "├"
			if i == len(tagStats)-1 {
				prefix = "└"
			}
			fmt.Fprintf(&text, "%s %s: %d\n", prefix, ts.Name, ts.Count)
		}
	}

	// 机场流量概览
	buildAirportUsageOverview(&text)

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("🔄 Refresh", "stats")},
	}

	return bot.SendMessageWithKeyboard(message.Chat.ID, text.String(), "Markdown", keyboard)
}

// truncateName 截断名称
func truncateName(name string, maxLen int) string {
	runes := []rune(name)
	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + "..."
	}
	return name
}

// getCountryFlag 获取国家对应的国旗 Emoji
func getCountryFlag(countryCode string) string {
	countryCode = strings.ToUpper(countryCode)
	if len(countryCode) != 2 {
		return "🏳️"
	}
	// 特殊处理
	if countryCode == "UK" {
		countryCode = "GB"
	}

	// 转换逻辑：A=0x1F1E6
	const regionalIndicatorBase = 0x1F1E6
	first := rune(regionalIndicatorBase + int(countryCode[0]) - 'A')
	second := rune(regionalIndicatorBase + int(countryCode[1]) - 'A')
	return string(first) + string(second)
}

// KeyValue 用于排序
type KeyValue struct {
	Key   string
	Value int
}

// sortMapByValue 按值排序 map
func sortMapByValue(m map[string]int) []KeyValue {
	var kvs []KeyValue
	for k, v := range m {
		kvs = append(kvs, KeyValue{k, v})
	}
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Value > kvs[j].Value
	})
	return kvs
}

// buildAirportUsageOverview 构建机场流量概览区块
func buildAirportUsageOverview(text *strings.Builder) {
	var airport models.Airport
	airports, err := airport.List()
	if err != nil || len(airports) == 0 {
		return
	}

	// 筛选开启用量获取且有有效数据的机场
	var airportsWithUsage []models.Airport
	for _, a := range airports {
		if a.FetchUsageInfo && a.UsageTotal > 0 {
			airportsWithUsage = append(airportsWithUsage, a)
		}
	}

	if len(airportsWithUsage) == 0 {
		return
	}

	// 全局流量汇总
	var totalUsed, totalQuota int64
	for _, a := range airportsWithUsage {
		totalUsed += a.UsageUpload + a.UsageDownload
		totalQuota += a.UsageTotal
	}

	var globalPercent float64
	if totalQuota > 0 {
		globalPercent = float64(totalUsed) / float64(totalQuota) * 100
		if globalPercent > 100 {
			globalPercent = 100
		}
	}

	// 最近到期机场
	now := time.Now().Unix()
	var nearestExpireAirport *models.Airport
	for i := range airportsWithUsage {
		a := &airportsWithUsage[i]
		if a.UsageExpire > now {
			if nearestExpireAirport == nil || a.UsageExpire < nearestExpireAirport.UsageExpire {
				nearestExpireAirport = a
			}
		}
	}

	// 低流量机场（剩余 < 10%）
	var lowUsageAirports []models.Airport
	for _, a := range airportsWithUsage {
		used := a.UsageUpload + a.UsageDownload
		remaining := a.UsageTotal - used
		if float64(remaining)/float64(a.UsageTotal) < 0.1 {
			lowUsageAirports = append(lowUsageAirports, a)
		}
	}

	// 构建输出
	text.WriteString("\n✈️ *Airport traffic overview*\n")
	fmt.Fprintf(text, "├ Airports: %d\n", len(airportsWithUsage))
	fmt.Fprintf(text, "├ Global usage: %s / %s (%.1f%%)\n",
		formatBytesLocal(totalUsed), formatBytesLocal(totalQuota), globalPercent)

	if nearestExpireAirport != nil {
		fmt.Fprintf(text, "├ Next expiry: %s\n", truncateName(nearestExpireAirport.Name, 15))
		fmt.Fprintf(text, "│    └ %s\n", formatExpireTimeLocal(nearestExpireAirport.UsageExpire))
	}

	if len(lowUsageAirports) > 0 {
		fmt.Fprintf(text, "└ ⚠️ Low traffic: %d\n", len(lowUsageAirports))
		for i, a := range lowUsageAirports {
			if i >= 3 { // 最多显示3个
				fmt.Fprintf(text, "     └ ...and %d more\n", len(lowUsageAirports)-3)
				break
			}
			fmt.Fprintf(text, "     %s %s\n", "├", truncateName(a.Name, 20))
		}
	} else {
		text.WriteString("└ ✓ All airports have sufficient traffic\n")
	}
}

// formatBytesLocal 格式化字节数为可读格式
func formatBytesLocal(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	if bytes < 0 {
		return "N/A"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	if exp >= len(units)-1 {
		exp = len(units) - 2
	}

	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp+1])
}

// formatExpireTimeLocal 格式化到期时间
func formatExpireTimeLocal(timestamp int64) string {
	if timestamp <= 0 {
		return "Unknown"
	}
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04")
}

// ============ MonitorHandler ============

type MonitorHandler struct{}

func (h *MonitorHandler) Command() string     { return "monitor" }
func (h *MonitorHandler) Description() string { return "🖥️ System monitor" }

func (h *MonitorHandler) Handle(bot *TelegramBot, message *Message) error {
	stats := monitor.GetSystemStats()

	// 转换字节为 MB
	heapAllocMB := float64(stats.HeapAlloc) / 1024 / 1024
	sysMB := float64(stats.Sys) / 1024 / 1024

	text := fmt.Sprintf(`🖥️ *系统监控*

*内存使用*
├ 堆分配: %.2f MB
├ 系统总: %.2f MB
└ GC 次数: %d

*运行状态*
├ Goroutines: %d
├ CPU 核心: %d
└ 运行时间: %d 秒`,
		heapAllocMB,
		sysMB,
		stats.NumGC,
		stats.NumGoroutine,
		stats.NumCPU,
		stats.Uptime)

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("🔄 Refresh", "monitor"), NewInlineButton("📊 Stats", "stats")},
	}

	return bot.SendMessageWithKeyboard(message.Chat.ID, text, "Markdown", keyboard)
}

// ============ ProfilesHandler ============

type ProfilesHandler struct{}

const profilesPageSize = 8 // 每页显示策略数量

func (h *ProfilesHandler) Command() string     { return "profiles" }
func (h *ProfilesHandler) Description() string { return "⚡ Check profiles" }

func (h *ProfilesHandler) Handle(bot *TelegramBot, message *Message) error {
	return h.HandleWithPage(bot, message, 0)
}

// HandleWithPage 处理带分页的策略列表
func (h *ProfilesHandler) HandleWithPage(bot *TelegramBot, message *Message, page int) error {
	profiles, err := GetNodeCheckProfiles()
	if err != nil {
		return bot.SendMessage(message.Chat.ID, "❌ Failed to load profiles: "+err.Error(), "")
	}

	if len(profiles) == 0 {
		text := "⚡ *Check profiles*\n\nNo check profiles yet; create one in the web UI."
		return bot.SendMessage(message.Chat.ID, text, "Markdown")
	}

	total := len(profiles)
	totalPages := (total + profilesPageSize - 1) / profilesPageSize

	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	start := page * profilesPageSize
	end := start + profilesPageSize
	if end > total {
		end = total
	}

	var text strings.Builder
	if totalPages > 1 {
		fmt.Fprintf(&text, "⚡ *Check profiles* (%d/%d)\n\n", page+1, totalPages)
	} else {
		text.WriteString("⚡ *Check profiles*\n\n")
	}

	var keyboard [][]InlineKeyboardButton

	for i := start; i < end; i++ {
		p := profiles[i]

		// 状态图标
		status := "⏸️"
		if p.Enabled {
			status = "✅"
		}

		// 模式显示（与Web端保持一致）
		mode := "Latency test only"
		if p.Mode == "mihomo" {
			mode = "Latency + speed test"
		}

		fmt.Fprintf(&text, "%s *%s*\n", status, p.Name)
		fmt.Fprintf(&text, "   └ Mode: %s", mode)
		if p.CronExpr != "" {
			fmt.Fprintf(&text, " | Schedule: `%s`", p.CronExpr)
		}
		text.WriteString("\n")

		if p.LastRunTime != nil {
			fmt.Fprintf(&text, "   └ Last run: %s\n", p.LastRunTime.Format("01-02 15:04"))
		}
		text.WriteString("\n")

		// 操作按钮
		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton("🔍 "+truncateName(p.Name, 10), fmt.Sprintf("profile_detail:%d", p.ID)),
			NewInlineButton("▶️ Run", fmt.Sprintf("profile_run:%d", p.ID)),
		})
	}

	// 分页按钮
	if totalPages > 1 {
		var navButtons []InlineKeyboardButton
		if page > 0 {
			navButtons = append(navButtons, NewInlineButton("⬅️ Previous", fmt.Sprintf("profiles_page:%d", page-1)))
		}
		if page < totalPages-1 {
			navButtons = append(navButtons, NewInlineButton("➡️ Next", fmt.Sprintf("profiles_page:%d", page+1)))
		}
		if len(navButtons) > 0 {
			keyboard = append(keyboard, navButtons)
		}
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

	if untestedCount > 0 {
		fmt.Fprintf(&text, "\n📌 *Untested nodes: %d*\n", untestedCount)
		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton("🔍 Test untested nodes", "profile_select_untested"),
		})
	}

	keyboard = append(keyboard, []InlineKeyboardButton{
		NewInlineButton("🔙 Back", "start"),
	})

	return bot.SendMessageWithKeyboard(message.Chat.ID, text.String(), "Markdown", keyboard)
}

// ============ SubscriptionsHandler ============

type SubscriptionsHandler struct{}

const subscriptionsPageSize = 8 // 每页显示订阅数量

func (h *SubscriptionsHandler) Command() string     { return "subscriptions" }
func (h *SubscriptionsHandler) Description() string { return "📋 Subscriptions" }

func (h *SubscriptionsHandler) Handle(bot *TelegramBot, message *Message) error {
	return h.HandleWithPage(bot, message, 0)
}

// HandleWithPage 处理带分页的订阅列表
func (h *SubscriptionsHandler) HandleWithPage(bot *TelegramBot, message *Message, page int) error {
	var sub models.Subcription
	subs, err := sub.List()
	if err != nil {
		return fmt.Errorf("failed to load subscription list: %v", err)
	}

	if len(subs) == 0 {
		return bot.SendMessage(message.Chat.ID, "📋 No subscriptions", "")
	}

	total := len(subs)
	totalPages := (total + subscriptionsPageSize - 1) / subscriptionsPageSize

	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	start := page * subscriptionsPageSize
	end := start + subscriptionsPageSize
	if end > total {
		end = total
	}

	var text strings.Builder
	if totalPages > 1 {
		fmt.Fprintf(&text, "📋 *Subscriptions* (%d/%d)\n\n", page+1, totalPages)
	} else {
		text.WriteString("📋 *Subscriptions*\n\n")
	}

	var keyboard [][]InlineKeyboardButton

	for i := start; i < end; i++ {
		s := subs[i]

		// 获取节点数和分组数
		nodeCount := len(s.NodesWithSort)
		groupCount := len(s.GroupsWithSort)

		fmt.Fprintf(&text, "*%d. %s*\n", i+1, truncateName(s.Name, 20))
		fmt.Fprintf(&text, "   └ %d nodes, %d groups\n", nodeCount, groupCount)
		if s.CreatedAt.Year() > 2000 {
			fmt.Fprintf(&text, "   └ %s\n", s.CreatedAt.Format("2006-01-02"))
		}
		text.WriteString("\n")

		// 每个订阅一行按钮
		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton("📝 "+truncateName(s.Name, 12), fmt.Sprintf("sub_link:%d", s.ID)),
		})
	}

	// 分页按钮
	if totalPages > 1 {
		var navButtons []InlineKeyboardButton
		if page > 0 {
			navButtons = append(navButtons, NewInlineButton("⬅️ Previous", fmt.Sprintf("subscriptions_page:%d", page-1)))
		}
		if page < totalPages-1 {
			navButtons = append(navButtons, NewInlineButton("➡️ Next", fmt.Sprintf("subscriptions_page:%d", page+1)))
		}
		if len(navButtons) > 0 {
			keyboard = append(keyboard, navButtons)
		}
	}

	keyboard = append(keyboard, []InlineKeyboardButton{
		NewInlineButton("🔙 Back", "start"),
	})

	return bot.SendMessageWithKeyboard(message.Chat.ID, text.String(), "Markdown", keyboard)
}

// ============ NodesHandler ============

type NodesHandler struct{}

func (h *NodesHandler) Command() string     { return "nodes" }
func (h *NodesHandler) Description() string { return "🌐 Nodes" }

func (h *NodesHandler) Handle(bot *TelegramBot, message *Message) error {
	var node models.Node
	nodes, _ := node.List()
	total := len(nodes)

	// 统计在线节点
	onlineCount := 0
	for _, n := range nodes {
		if n.DelayStatus == "success" || n.SpeedStatus == "success" {
			onlineCount++
		}
	}

	// 获取地区分布
	countryStats := models.GetNodeCountryStats()

	// 排序地区统计
	type countryStat struct {
		Country string
		Count   int
	}
	var sortedCountries []countryStat
	for country, count := range countryStats {
		sortedCountries = append(sortedCountries, countryStat{country, count})
	}
	sort.Slice(sortedCountries, func(i, j int) bool {
		return sortedCountries[i].Count > sortedCountries[j].Count
	})

	var countryText strings.Builder
	for i, cs := range sortedCountries {
		if i >= 5 {
			break
		}
		fmt.Fprintf(&countryText, "├ %s: %d\n", cs.Country, cs.Count)
	}

	text := fmt.Sprintf(`🌐 *节点信息*

*节点概览*
├ 总数量: %d
├ 在线: %d
└ 离线: %d

*地区分布（前5）*
%s`, total, onlineCount, total-onlineCount, countryText.String())

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("🔄 Refresh", "nodes"), NewInlineButton("⚡ Check", "profiles")},
	}

	return bot.SendMessageWithKeyboard(message.Chat.ID, text, "Markdown", keyboard)
}

// ============ TagsHandler ============

type TagsHandler struct{}

const tagsPageSize = 10 // 每页显示标签规则数量

func (h *TagsHandler) Command() string     { return "tags" }
func (h *TagsHandler) Description() string { return "🏷️ Tag rules" }

func (h *TagsHandler) Handle(bot *TelegramBot, message *Message) error {
	return h.HandleWithPage(bot, message, 0)
}

// HandleWithPage 处理带分页的标签规则列表
func (h *TagsHandler) HandleWithPage(bot *TelegramBot, message *Message, page int) error {
	var tagRule models.TagRule
	rules, err := tagRule.List()
	if err != nil {
		return fmt.Errorf("failed to load tag rules: %v", err)
	}

	if len(rules) == 0 {
		return bot.SendMessage(message.Chat.ID, "🏷️ No tag rules", "")
	}

	total := len(rules)
	totalPages := (total + tagsPageSize - 1) / tagsPageSize

	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	start := page * tagsPageSize
	end := start + tagsPageSize
	if end > total {
		end = total
	}

	var text strings.Builder
	if totalPages > 1 {
		fmt.Fprintf(&text, "🏷️ *Tag rules* (%d/%d)\n\n", page+1, totalPages)
	} else {
		text.WriteString("🏷️ *Tag rules*\n\n")
	}

	var keyboard [][]InlineKeyboardButton

	for i := start; i < end; i++ {
		rule := rules[i]
		status := "✅"
		if !rule.Enabled {
			status = "⏸️"
		}
		fmt.Fprintf(&text, "%s %s → %s\n", status, rule.Name, rule.TagName)

		// 为每个规则添加执行按钮
		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton("▶️ "+truncateName(rule.Name, 15), fmt.Sprintf("tag_run:%d", rule.ID)),
		})
	}

	// 分页按钮
	if totalPages > 1 {
		var navButtons []InlineKeyboardButton
		if page > 0 {
			navButtons = append(navButtons, NewInlineButton("⬅️ Previous", fmt.Sprintf("tags_page:%d", page-1)))
		}
		if page < totalPages-1 {
			navButtons = append(navButtons, NewInlineButton("➡️ Next", fmt.Sprintf("tags_page:%d", page+1)))
		}
		if len(navButtons) > 0 {
			keyboard = append(keyboard, navButtons)
		}
	}

	keyboard = append(keyboard, []InlineKeyboardButton{
		NewInlineButton("🔙 Back", "start"),
	})

	return bot.SendMessageWithKeyboard(message.Chat.ID, text.String(), "Markdown", keyboard)
}

// ============ TasksHandler ============

type TasksHandler struct{}

func (h *TasksHandler) Command() string     { return "tasks" }
func (h *TasksHandler) Description() string { return "📝 Tasks" }

func (h *TasksHandler) Handle(bot *TelegramBot, message *Message) error {
	// 从服务层获取运行中任务（实时进度）
	runningTasks := GetRunningTasksFromService()

	if len(runningTasks) == 0 {
		text := "📝 *Tasks*\n\nNo running tasks"
		keyboard := [][]InlineKeyboardButton{
			{NewInlineButton("🔄 Refresh", "tasks")},
		}
		return bot.SendMessageWithKeyboard(message.Chat.ID, text, "Markdown", keyboard)
	}

	var text strings.Builder
	text.WriteString("📝 *Running tasks*\n\n")

	var keyboard [][]InlineKeyboardButton

	for _, task := range runningTasks {
		// 任务名称
		fmt.Fprintf(&text, "📋 *%s*\n", task.Name)

		// 进度信息
		if task.Total > 0 {
			percent := float64(task.Progress) / float64(task.Total) * 100
			fmt.Fprintf(&text, "├ Progress: %d/%d (%.0f%%)\n", task.Progress, task.Total, percent)
		}

		// 当前处理项
		if task.CurrentItem != "" {
			fmt.Fprintf(&text, "├ Current: %s\n", truncateName(task.CurrentItem, 30))
		}

		// 状态消息
		if task.Message != "" {
			fmt.Fprintf(&text, "└ Status: %s\n", task.Message)
		}
		text.WriteString("\n")

		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton("❌ Cancel "+truncateName(task.Name, 12), fmt.Sprintf("task_cancel:%s", task.ID)),
		})
	}

	keyboard = append(keyboard, []InlineKeyboardButton{
		NewInlineButton("🔄 Refresh", "tasks"),
	})

	return bot.SendMessageWithKeyboard(message.Chat.ID, text.String(), "Markdown", keyboard)
}

// ========== Service Wrapper ==========

// ServicesWrapper 服务包装器接口
type ServicesWrapper interface {
	ExecuteSubscriptionTaskWithTrigger(id int, url string, subName string, trigger models.TaskTrigger)
	ApplyAutoTagRules(nodes []models.Node, triggerSource string)
	CancelTask(taskID string) error
	GetRunningTasks() []models.Task
	GetNodeCheckProfiles() ([]models.NodeCheckProfile, error)
	ExecuteNodeCheckWithProfile(profileID int, nodeIDs []int, trigger models.TaskTrigger)
	ToggleProfileEnabled(profileID int) (bool, error)
	TriggerTagRule(ruleID int) error
}

var servicesWrapper ServicesWrapper

// SetServicesWrapper 设置服务包装器（在 main.go 中调用）
func SetServicesWrapper(wrapper ServicesWrapper) {
	servicesWrapper = wrapper
}

// GetRunningTasksFromService 从服务层获取运行中任务
func GetRunningTasksFromService() []models.Task {
	if servicesWrapper != nil {
		return servicesWrapper.GetRunningTasks()
	}
	// 降级到数据库查询
	tasks, _ := models.GetRunningTasks()
	return tasks
}

// GetNodeCheckProfiles 获取节点检测策略列表
func GetNodeCheckProfiles() ([]models.NodeCheckProfile, error) {
	if servicesWrapper != nil {
		return servicesWrapper.GetNodeCheckProfiles()
	}
	var profile models.NodeCheckProfile
	return profile.List()
}

// ========== Helper Functions ==========

// PullSubscription 拉取订阅（机场更新）
func PullSubscription(airportID int) error {
	airport, err := models.GetAirportByID(airportID)
	if err != nil {
		return fmt.Errorf("failed to load airport: %v", err)
	}

	// 通过包装器调用服务层
	if servicesWrapper != nil {
		go servicesWrapper.ExecuteSubscriptionTaskWithTrigger(airport.ID, airport.URL, airport.Name, models.TaskTriggerManual)
	}
	utils.Info("Telegram triggered airport update: %s", airport.Name)

	return nil
}

// ApplyAllTagRules 应用所有标签规则
func ApplyAllTagRules() error {
	var node models.Node
	nodes, err := node.List()
	if err != nil || len(nodes) == 0 {
		return fmt.Errorf("no nodes")
	}

	// 通过包装器调用服务层
	if servicesWrapper != nil {
		go servicesWrapper.ApplyAutoTagRules(nodes, "telegram_manual")
	}
	utils.Info("Telegram triggered tag rule application: %d nodes", len(nodes))

	return nil
}

// CancelTask 取消任务
func CancelTask(taskID string) error {
	if servicesWrapper != nil {
		return servicesWrapper.CancelTask(taskID)
	}
	return fmt.Errorf("service is not initialized")
}

// ExecuteNodeCheckWithProfile 执行节点检测
func ExecuteNodeCheckWithProfile(profileID int, nodeIDs []int, trigger models.TaskTrigger) error {
	if servicesWrapper != nil {
		go servicesWrapper.ExecuteNodeCheckWithProfile(profileID, nodeIDs, trigger)
		return nil
	}
	return fmt.Errorf("service is not initialized")
}

// ToggleProfileEnabled 开关策略定时执行
func ToggleProfileEnabled(profileID int) (bool, error) {
	if servicesWrapper != nil {
		return servicesWrapper.ToggleProfileEnabled(profileID)
	}
	return false, fmt.Errorf("service is not initialized")
}

// TriggerTagRule 执行指定标签规则
func TriggerTagRule(ruleID int) error {
	if servicesWrapper != nil {
		go func() { _ = servicesWrapper.TriggerTagRule(ruleID) }()
		return nil
	}
	return fmt.Errorf("service is not initialized")
}

// GetSubscriptionLink 获取订阅链接
// 返回链接和是否需要提示用户配置域名
func GetSubscriptionLink(subID int) (string, bool, error) {
	var sub models.Subcription
	sub.ID = subID
	// 使用 Find 方法获取订阅详情（包括 Name）
	if err := sub.Find(); err != nil {
		return "", false, fmt.Errorf("failed to load subscription: %v", err)
	}

	// 获取系统域名设置
	domain, _ := models.GetSetting("system_domain")
	needHint := false
	if domain == "" {
		// 未配置域名，使用 localhost:8000，并标记需要提示
		domain = "http://localhost:8000"
		needHint = true
	}
	// 确保没有末尾斜杠
	domain = strings.TrimRight(domain, "/")
	// 确保有协议头
	if !strings.HasPrefix(domain, "http") {
		domain = "http://" + domain
	}

	// 从分享表获取默认分享链接
	share, err := models.GetDefaultShareForSubscription(subID)
	if err != nil {
		return "", false, fmt.Errorf("failed to get share links: %v", err)
	}

	// 构建基础链接
	link := fmt.Sprintf("%s/c/?token=%s", domain, share.Token)
	return link, needHint, nil
}

// ============ AirportsHandler ============

type AirportsHandler struct{}

const airportsPageSize = 8 // 每页显示机场数量

func (h *AirportsHandler) Command() string     { return "airports" }
func (h *AirportsHandler) Description() string { return "✈️ Airports" }

func (h *AirportsHandler) Handle(bot *TelegramBot, message *Message) error {
	return h.HandleWithPage(bot, message, 0)
}

// HandleWithPage 处理带分页的机场列表
func (h *AirportsHandler) HandleWithPage(bot *TelegramBot, message *Message, page int) error {
	var airport models.Airport
	airports, err := airport.List()
	if err != nil {
		return fmt.Errorf("failed to load airport list: %v", err)
	}

	if len(airports) == 0 {
		return bot.SendMessage(message.Chat.ID, "✈️ No airports", "")
	}

	total := len(airports)
	totalPages := (total + airportsPageSize - 1) / airportsPageSize

	// 确保页码有效
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	start := page * airportsPageSize
	end := start + airportsPageSize
	if end > total {
		end = total
	}

	var text strings.Builder
	fmt.Fprintf(&text, "✈️ *Airports* (%d/%d)\n\n", page+1, totalPages)

	var keyboard [][]InlineKeyboardButton

	for i := start; i < end; i++ {
		ap := airports[i]

		status := "✅"
		if !ap.Enabled {
			status = "⏸️"
		}

		// 节点数量
		nodes, _ := models.ListNodesByAirportID(ap.ID)
		nodeCount := len(nodes)

		fmt.Fprintf(&text, "%s *%s*\n", status, truncateName(ap.Name, 20))
		fmt.Fprintf(&text, "   └ 🔗 %s\n", truncateName(ap.URL, 30))
		fmt.Fprintf(&text, "   └ 📦 %d nodes\n", nodeCount)
		if ap.LastRunTime != nil {
			fmt.Fprintf(&text, "   └ 🕒 Last updated: %s\n", ap.LastRunTime.Format("01-02 15:04"))
		}
		text.WriteString("\n")

		// 按钮
		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton("⚙️ Manage "+truncateName(ap.Name, 10), fmt.Sprintf("airport_detail:%d", ap.ID)),
		})
	}

	// 分页按钮
	if totalPages > 1 {
		var navButtons []InlineKeyboardButton
		if page > 0 {
			navButtons = append(navButtons, NewInlineButton("⬅️ Previous", fmt.Sprintf("airports_page:%d", page-1)))
		}
		if page < totalPages-1 {
			navButtons = append(navButtons, NewInlineButton("➡️ Next", fmt.Sprintf("airports_page:%d", page+1)))
		}
		if len(navButtons) > 0 {
			keyboard = append(keyboard, navButtons)
		}
	}

	keyboard = append(keyboard, []InlineKeyboardButton{
		NewInlineButton("🔙 Back", "start"),
	})

	return bot.SendMessageWithKeyboard(message.Chat.ID, text.String(), "Markdown", keyboard)
}
