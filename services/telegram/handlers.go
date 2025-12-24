package telegram

import (
	"fmt"
	"sort"
	"strings"
	"sublink/models"
	"sublink/services/monitor"
	"sublink/utils"
	"sync"
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
func (h *StartHandler) Description() string { return "🚀 开始使用" }

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
		{NewInlineButton("📊 统计", "stats"), NewInlineButton("🖥️ 监控", "monitor")},
		{NewInlineButton("⚡ 检测策略", "profiles"), NewInlineButton("📋 订阅", "subscriptions")},
		{NewInlineButton("❓ 帮助", "help")},
	}

	return bot.SendMessageWithKeyboard(message.Chat.ID, text, "Markdown", keyboard)
}

// ============ HelpHandler ============

type HelpHandler struct{}

func (h *HelpHandler) Command() string     { return "help" }
func (h *HelpHandler) Description() string { return "❓ 帮助信息" }

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
func (h *StatsHandler) Description() string { return "📊 仪表盘统计" }

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
	text.WriteString("📊 *仪表盘统计*\n\n")

	// 基础统计
	text.WriteString(fmt.Sprintf("📋 订阅: *%d*\n", subCount))
	text.WriteString(fmt.Sprintf("📦 节点: *%d* / %d\n\n", available, total))

	// 最快速度
	if fastestNode != nil && fastestNode.Speed > 0 {
		text.WriteString(fmt.Sprintf("🚀 最快速度: *%.2f MB/s*\n", fastestNode.Speed))
		text.WriteString(fmt.Sprintf("   └ %s\n\n", truncateName(fastestNode.Name, 25)))
	}

	// 最低延迟
	if lowestDelayNode != nil && lowestDelayNode.DelayTime > 0 {
		text.WriteString(fmt.Sprintf("⚡ 最低延迟: *%d ms*\n", lowestDelayNode.DelayTime))
		text.WriteString(fmt.Sprintf("   └ %s\n\n", truncateName(lowestDelayNode.Name, 25)))
	}

	// 国家分布
	if len(countryStats) > 0 {
		text.WriteString("🌍 *国家分布*\n")
		sortedCountries := sortMapByValue(countryStats)
		for i, kv := range sortedCountries {
			prefix := "├"
			if i == len(sortedCountries)-1 {
				prefix = "└"
			}
			flag := getCountryFlag(kv.Key)
			text.WriteString(fmt.Sprintf("%s %s %s: %d\n", prefix, flag, kv.Key, kv.Value))
		}
		text.WriteString("\n")
	}

	// 协议分布
	if len(protocolStats) > 0 {
		text.WriteString("📡 *协议分布*\n")
		sortedProtocols := sortMapByValue(protocolStats)
		for i, kv := range sortedProtocols {
			prefix := "├"
			if i == len(sortedProtocols)-1 {
				prefix = "└"
			}
			text.WriteString(fmt.Sprintf("%s %s: %d\n", prefix, kv.Key, kv.Value))
		}
		text.WriteString("\n")
	}

	// 标签分布
	tagStats := models.GetNodeTagStats()
	if len(tagStats) > 0 {
		text.WriteString("🏷️ *标签分布*\n")
		// 排序标签统计
		sort.Slice(tagStats, func(i, j int) bool {
			return tagStats[i].Count > tagStats[j].Count
		})

		for i, ts := range tagStats {
			prefix := "├"
			if i == len(tagStats)-1 {
				prefix = "└"
			}
			text.WriteString(fmt.Sprintf("%s %s: %d\n", prefix, ts.Name, ts.Count))
		}
	}

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("🔄 刷新", "stats")},
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

// ============ MonitorHandler ============

type MonitorHandler struct{}

func (h *MonitorHandler) Command() string     { return "monitor" }
func (h *MonitorHandler) Description() string { return "🖥️ 系统监控" }

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
		{NewInlineButton("🔄 刷新", "monitor"), NewInlineButton("📊 统计", "stats")},
	}

	return bot.SendMessageWithKeyboard(message.Chat.ID, text, "Markdown", keyboard)
}

// ============ ProfilesHandler ============

type ProfilesHandler struct{}

func (h *ProfilesHandler) Command() string     { return "profiles" }
func (h *ProfilesHandler) Description() string { return "⚡ 检测策略" }

func (h *ProfilesHandler) Handle(bot *TelegramBot, message *Message) error {
	profiles, err := GetNodeCheckProfiles()
	if err != nil {
		return bot.SendMessage(message.Chat.ID, "❌ 获取策略列表失败: "+err.Error(), "")
	}

	if len(profiles) == 0 {
		text := "⚡ *检测策略*\n\n暂无检测策略，请在 Web 端创建。"
		return bot.SendMessage(message.Chat.ID, text, "Markdown")
	}

	var text strings.Builder
	text.WriteString("⚡ *检测策略列表*\n\n")

	var keyboard [][]InlineKeyboardButton

	for i, p := range profiles {
		if i >= 10 {
			text.WriteString(fmt.Sprintf("\n... 还有 %d 个策略", len(profiles)-10))
			break
		}

		// 状态图标
		status := "⏸️"
		if p.Enabled {
			status = "✅"
		}

		// 模式显示（与Web端保持一致）
		mode := "仅延迟测试"
		if p.Mode == "mihomo" {
			mode = "延迟+速度测试"
		}

		text.WriteString(fmt.Sprintf("%s *%s*\n", status, p.Name))
		text.WriteString(fmt.Sprintf("   └ 模式: %s", mode))
		if p.CronExpr != "" {
			text.WriteString(fmt.Sprintf(" | 定时: `%s`", p.CronExpr))
		}
		text.WriteString("\n")

		if p.LastRunTime != nil {
			text.WriteString(fmt.Sprintf("   └ 上次执行: %s\n", p.LastRunTime.Format("01-02 15:04")))
		}
		text.WriteString("\n")

		// 操作按钮
		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton("🔍 "+truncateName(p.Name, 10), fmt.Sprintf("profile_detail:%d", p.ID)),
			NewInlineButton("▶️ 执行", fmt.Sprintf("profile_run:%d", p.ID)),
		})
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
		text.WriteString(fmt.Sprintf("\n📌 *未测速节点: %d*\n", untestedCount))
		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton("🔍 选择策略检测未测速节点", "profile_select_untested"),
		})
	}

	keyboard = append(keyboard, []InlineKeyboardButton{
		NewInlineButton("🔙 返回", "start"),
	})

	return bot.SendMessageWithKeyboard(message.Chat.ID, text.String(), "Markdown", keyboard)
}

// ============ SubscriptionsHandler ============

type SubscriptionsHandler struct{}

func (h *SubscriptionsHandler) Command() string     { return "subscriptions" }
func (h *SubscriptionsHandler) Description() string { return "📋 订阅管理" }

func (h *SubscriptionsHandler) Handle(bot *TelegramBot, message *Message) error {
	// 获取订阅链接列表
	var sub models.Subcription
	subs, err := sub.List()
	if err != nil {
		return fmt.Errorf("获取订阅列表失败: %v", err)
	}

	if len(subs) == 0 {
		return bot.SendMessage(message.Chat.ID, "📋 暂无订阅", "")
	}

	var text strings.Builder
	text.WriteString("📋 *订阅列表*\n\n")

	var keyboard [][]InlineKeyboardButton

	for i, s := range subs {
		if i >= 8 {
			text.WriteString(fmt.Sprintf("\n... 还有 %d 个订阅", len(subs)-8))
			break
		}

		// 获取节点数和分组数
		nodeCount := len(s.NodesWithSort)
		groupCount := len(s.GroupsWithSort)

		text.WriteString(fmt.Sprintf("*%d. %s*\n", i+1, truncateName(s.Name, 20)))
		text.WriteString(fmt.Sprintf("   └ %d 节点, %d 分组\n", nodeCount, groupCount))
		if s.CreatedAt.Year() > 2000 {
			text.WriteString(fmt.Sprintf("   └ %s\n", s.CreatedAt.Format("2006-01-02")))
		}
		text.WriteString("\n")

		// 每个订阅一行按钮
		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton("📝 "+truncateName(s.Name, 12), fmt.Sprintf("sub_link:%d", s.ID)),
		})
	}

	keyboard = append(keyboard, []InlineKeyboardButton{
		NewInlineButton("🔙 返回", "start"),
	})

	return bot.SendMessageWithKeyboard(message.Chat.ID, text.String(), "Markdown", keyboard)
}

// ============ NodesHandler ============

type NodesHandler struct{}

func (h *NodesHandler) Command() string     { return "nodes" }
func (h *NodesHandler) Description() string { return "🌐 节点信息" }

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
		countryText.WriteString(fmt.Sprintf("├ %s: %d\n", cs.Country, cs.Count))
	}

	text := fmt.Sprintf(`🌐 *节点信息*

*节点概览*
├ 总数量: %d
├ 在线: %d
└ 离线: %d

*地区分布（前5）*
%s`, total, onlineCount, total-onlineCount, countryText.String())

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("🔄 刷新", "nodes"), NewInlineButton("⚡ 检测", "profiles")},
	}

	return bot.SendMessageWithKeyboard(message.Chat.ID, text, "Markdown", keyboard)
}

// ============ TagsHandler ============

type TagsHandler struct{}

func (h *TagsHandler) Command() string     { return "tags" }
func (h *TagsHandler) Description() string { return "🏷️ 标签规则" }

func (h *TagsHandler) Handle(bot *TelegramBot, message *Message) error {
	// 获取标签规则
	var tagRule models.TagRule
	rules, err := tagRule.List()
	if err != nil {
		return fmt.Errorf("获取标签规则失败: %v", err)
	}

	if len(rules) == 0 {
		return bot.SendMessage(message.Chat.ID, "🏷️ 暂无标签规则", "")
	}

	var text strings.Builder
	text.WriteString("🏷️ *标签规则*\n\n")

	for i, rule := range rules {
		if i >= 10 {
			text.WriteString(fmt.Sprintf("\n... 还有 %d 条规则", len(rules)-10))
			break
		}

		status := "✅"
		if !rule.Enabled {
			status = "⏸️"
		}
		text.WriteString(fmt.Sprintf("%s %s → %s\n", status, rule.Name, rule.TagName))
	}

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("▶️ 执行全部标签规则", "tags:apply_all")},
		{NewInlineButton("🔙 返回", "start")},
	}

	return bot.SendMessageWithKeyboard(message.Chat.ID, text.String(), "Markdown", keyboard)
}

// ============ TasksHandler ============

type TasksHandler struct{}

func (h *TasksHandler) Command() string     { return "tasks" }
func (h *TasksHandler) Description() string { return "📝 任务管理" }

func (h *TasksHandler) Handle(bot *TelegramBot, message *Message) error {
	// 从服务层获取运行中任务（实时进度）
	runningTasks := GetRunningTasksFromService()

	if len(runningTasks) == 0 {
		text := "📝 *任务管理*\n\n暂无正在运行的任务"
		keyboard := [][]InlineKeyboardButton{
			{NewInlineButton("🔄 刷新", "tasks")},
		}
		return bot.SendMessageWithKeyboard(message.Chat.ID, text, "Markdown", keyboard)
	}

	var text strings.Builder
	text.WriteString("📝 *正在运行的任务*\n\n")

	var keyboard [][]InlineKeyboardButton

	for _, task := range runningTasks {
		// 任务名称
		text.WriteString(fmt.Sprintf("📋 *%s*\n", task.Name))

		// 进度信息
		if task.Total > 0 {
			percent := float64(task.Progress) / float64(task.Total) * 100
			text.WriteString(fmt.Sprintf("├ 进度: %d/%d (%.0f%%)\n", task.Progress, task.Total, percent))
		}

		// 当前处理项
		if task.CurrentItem != "" {
			text.WriteString(fmt.Sprintf("├ 当前: %s\n", truncateName(task.CurrentItem, 30)))
		}

		// 状态消息
		if task.Message != "" {
			text.WriteString(fmt.Sprintf("└ 状态: %s\n", task.Message))
		}
		text.WriteString("\n")

		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton("❌ 取消 "+truncateName(task.Name, 12), fmt.Sprintf("task_cancel:%s", task.ID)),
		})
	}

	keyboard = append(keyboard, []InlineKeyboardButton{
		NewInlineButton("🔄 刷新", "tasks"),
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
	ExecuteNodeCheckWithProfile(profileID int, nodeIDs []int)
	ToggleProfileEnabled(profileID int) (bool, error)
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
		return fmt.Errorf("获取机场失败: %v", err)
	}

	// 通过包装器调用服务层
	if servicesWrapper != nil {
		go servicesWrapper.ExecuteSubscriptionTaskWithTrigger(airport.ID, airport.URL, airport.Name, models.TaskTriggerManual)
	}
	utils.Info("Telegram 触发机场更新: %s", airport.Name)

	return nil
}

// ApplyAllTagRules 应用所有标签规则
func ApplyAllTagRules() error {
	var node models.Node
	nodes, err := node.List()
	if err != nil || len(nodes) == 0 {
		return fmt.Errorf("没有节点")
	}

	// 通过包装器调用服务层
	if servicesWrapper != nil {
		go servicesWrapper.ApplyAutoTagRules(nodes, "telegram_manual")
	}
	utils.Info("Telegram 触发标签规则应用: %d 个节点", len(nodes))

	return nil
}

// CancelTask 取消任务
func CancelTask(taskID string) error {
	if servicesWrapper != nil {
		return servicesWrapper.CancelTask(taskID)
	}
	return fmt.Errorf("服务未初始化")
}

// ExecuteNodeCheckWithProfile 执行节点检测
func ExecuteNodeCheckWithProfile(profileID int, nodeIDs []int) error {
	if servicesWrapper != nil {
		go servicesWrapper.ExecuteNodeCheckWithProfile(profileID, nodeIDs)
		return nil
	}
	return fmt.Errorf("服务未初始化")
}

// ToggleProfileEnabled 开关策略定时执行
func ToggleProfileEnabled(profileID int) (bool, error) {
	if servicesWrapper != nil {
		return servicesWrapper.ToggleProfileEnabled(profileID)
	}
	return false, fmt.Errorf("服务未初始化")
}

// GetSubscriptionLink 获取订阅链接
func GetSubscriptionLink(subID int) (string, error) {
	var sub models.Subcription
	sub.ID = subID
	// 使用 Find 方法获取订阅详情（包括 Name）
	if err := sub.Find(); err != nil {
		return "", fmt.Errorf("获取订阅失败: %v", err)
	}

	// 获取系统域名设置
	domain, _ := models.GetSetting("system_domain")
	if domain == "" {
		domain, _ = models.GetSetting("server_addr")
	}
	if domain == "" {
		domain = "http://localhost:8080"
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
		return "", fmt.Errorf("获取分享链接失败: %v", err)
	}

	// 构建基础链接
	link := fmt.Sprintf("%s/c/?token=%s", domain, share.Token)
	return link, nil
}

// ============ AirportsHandler ============

type AirportsHandler struct{}

func (h *AirportsHandler) Command() string     { return "airports" }
func (h *AirportsHandler) Description() string { return "✈️ 机场管理" }

func (h *AirportsHandler) Handle(bot *TelegramBot, message *Message) error {
	var airport models.Airport
	airports, err := airport.List()
	if err != nil {
		return fmt.Errorf("获取机场列表失败: %v", err)
	}

	if len(airports) == 0 {
		return bot.SendMessage(message.Chat.ID, "✈️ 暂无机场", "")
	}

	var text strings.Builder
	text.WriteString("✈️ *机场列表*\n\n")

	var keyboard [][]InlineKeyboardButton

	// 分页显示? 暂时限制前 10 个，类似 SubscriptionsHandler
	for i, ap := range airports {
		if i >= 10 {
			text.WriteString(fmt.Sprintf("\n... 还有 %d 个机场", len(airports)-10))
			break
		}

		status := "✅"
		if !ap.Enabled {
			status = "⏸️"
		}

		// 节点数量
		nodes, _ := models.ListNodesByAirportID(ap.ID)
		nodeCount := len(nodes)

		text.WriteString(fmt.Sprintf("%s *%s*\n", status, truncateName(ap.Name, 20)))
		text.WriteString(fmt.Sprintf("   └ 🔗 %s\n", truncateName(ap.URL, 30)))
		text.WriteString(fmt.Sprintf("   └ 📦 %d 个节点\n", nodeCount))
		if ap.LastRunTime != nil {
			text.WriteString(fmt.Sprintf("   └ 🕒 上次更新: %s\n", ap.LastRunTime.Format("01-02 15:04")))
		}
		text.WriteString("\n")

		// 按钮
		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton("⚙️ 管理 "+truncateName(ap.Name, 10), fmt.Sprintf("airport_detail:%d", ap.ID)),
		})
	}

	keyboard = append(keyboard, []InlineKeyboardButton{
		NewInlineButton("🔙 返回", "start"),
	})

	return bot.SendMessageWithKeyboard(message.Chat.ID, text.String(), "Markdown", keyboard)
}
