package telegram

import (
	"crypto/md5"
	"encoding/hex"
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
	RegisterHandler("speedtest", &SpeedTestHandler{})
	RegisterHandler("subscriptions", &SubscriptionsHandler{})
	RegisterHandler("nodes", &NodesHandler{})
	RegisterHandler("tags", &TagsHandler{})
	RegisterHandler("tasks", &TasksHandler{})
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
• ⚡ 开始节点测速任务
• 📋 管理订阅和节点
• 🏷️ 执行标签规则
• 📝 查看和管理任务

使用 /help 查看详细命令列表`

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("📊 统计", "stats"), NewInlineButton("🖥️ 监控", "monitor")},
		{NewInlineButton("⚡ 测速", "speedtest"), NewInlineButton("📋 订阅", "subscriptions")},
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
/speedtest - ⚡ 开始测速
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

// ============ SpeedTestHandler ============

type SpeedTestHandler struct{}

func (h *SpeedTestHandler) Command() string     { return "speedtest" }
func (h *SpeedTestHandler) Description() string { return "⚡ 开始测速" }

func (h *SpeedTestHandler) Handle(bot *TelegramBot, message *Message) error {
	// 统计未测速节点数
	var node models.Node
	nodes, _ := node.List()
	untestedCount := 0
	for _, n := range nodes {
		if n.DelayStatus == "" || n.DelayStatus == "untested" {
			untestedCount++
		}
	}

	text := fmt.Sprintf(`⚡ *测速任务*

节点总数: %d
未测速: %d

请选择测速方式：`, len(nodes), untestedCount)

	keyboard := [][]InlineKeyboardButton{
		{NewInlineButton("▶️ 执行定时测速", "speedtest:scheduled")},
		{NewInlineButton("⏰ 测试未测速节点", "speedtest:untested")},
	}

	return bot.SendMessageWithKeyboard(message.Chat.ID, text, "Markdown", keyboard)
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
		{NewInlineButton("🔄 刷新", "nodes"), NewInlineButton("⚡ 测速", "speedtest")},
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
		progress := ""
		if task.Total > 0 {
			progress = fmt.Sprintf(" (%d/%d)", task.Progress, task.Total)
		}
		text.WriteString(fmt.Sprintf("• %s%s\n", task.Name, progress))

		keyboard = append(keyboard, []InlineKeyboardButton{
			NewInlineButton("❌ 取消 "+task.Name, fmt.Sprintf("task_cancel:%s", task.ID)),
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
	RunSpeedTestOnNodes(nodes []models.Node)
	ExecuteScheduledSpeedTest()
	ExecuteSubscriptionTaskWithTrigger(id int, url string, subName string, trigger models.TaskTrigger)
	ApplyAutoTagRules(nodes []models.Node, triggerSource string)
	CancelTask(taskID string) error
	GetRunningTasks() []models.Task
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

// ========== Helper Functions ==========

// RunSpeedTest 启动测速任务
func RunSpeedTest(scope string) error {
	switch scope {
	case "scheduled":
		// 执行定时测速配置（与 Web 端绿色按钮一致）
		if servicesWrapper != nil {
			go servicesWrapper.ExecuteScheduledSpeedTest()
		}
		utils.Info("Telegram 触发定时测速任务")
		return nil

	case "untested":
		var node models.Node
		allNodes, err := node.List()
		if err != nil {
			return fmt.Errorf("获取节点失败: %v", err)
		}
		// 筛选未测速节点
		var nodes []models.Node
		for _, n := range allNodes {
			if n.DelayStatus == "" || n.DelayStatus == "untested" {
				nodes = append(nodes, n)
			}
		}
		if len(nodes) == 0 {
			return fmt.Errorf("没有未测速的节点")
		}
		// 通过包装器调用服务层
		if servicesWrapper != nil {
			go servicesWrapper.RunSpeedTestOnNodes(nodes)
		}
		utils.Info("Telegram 触发未测速节点测速: %d 个节点", len(nodes))
		return nil

	default:
		return fmt.Errorf("未知的测速范围: %s", scope)
	}
}

// PullSubscription 拉取订阅
func PullSubscription(subID int) error {
	var sub models.SubScheduler
	if err := sub.GetByID(subID); err != nil {
		return fmt.Errorf("获取订阅失败: %v", err)
	}

	// 通过包装器调用服务层
	if servicesWrapper != nil {
		go servicesWrapper.ExecuteSubscriptionTaskWithTrigger(sub.ID, sub.URL, sub.Name, models.TaskTriggerManual)
	}
	utils.Info("Telegram 触发订阅更新: %s", sub.Name)

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

// GetSubscriptionLink 获取订阅链接
func GetSubscriptionLink(subID int) (string, error) {
	var sub models.Subcription
	sub.ID = subID
	// 使用 Find 方法获取订阅详情（包括 Name）
	// 注意：GetSub 只加载关联数据（节点等），不会加载订阅本身的信息
	if err := sub.Find(); err != nil {
		return "", fmt.Errorf("获取订阅失败: %v", err)
	}

	// 获取系统域名设置
	domain, _ := models.GetSetting("system_domain") // 优先使用 system_domain
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

	// Token 生成规则: MD5(SubscriptionName)
	// 参考 api/clients.go 中的验证逻辑
	h := md5.New()
	h.Write([]byte(sub.Name))
	token := hex.EncodeToString(h.Sum(nil))

	// 构建基础链接
	link := fmt.Sprintf("%s/c/?token=%s", domain, token)
	return link, nil
}
