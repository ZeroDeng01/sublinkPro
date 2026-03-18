package notifications

import "slices"

type Channel string

const (
	ChannelWebhook  Channel = "webhook"
	ChannelTelegram Channel = "telegram"
	ChannelInApp    Channel = "in_app"
)

type EventDefinition struct {
	Key            string    `json:"key"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Category       string    `json:"category"`
	CategoryName   string    `json:"categoryName"`
	Severity       string    `json:"severity"`
	Channels       []Channel `json:"channels"`
	DefaultEnabled bool      `json:"defaultEnabled"`
}

type Payload struct {
	Event        string      `json:"event"`
	EventName    string      `json:"eventName,omitempty"`
	Category     string      `json:"category,omitempty"`
	CategoryName string      `json:"categoryName,omitempty"`
	Severity     string      `json:"severity,omitempty"`
	Title        string      `json:"title"`
	Message      string      `json:"message"`
	Data         interface{} `json:"data"`
	Time         string      `json:"time"`
}

var eventCatalog = []EventDefinition{
	{
		Key:            "subscription.sync_succeeded",
		Name:           "订阅更新成功",
		Description:    "机场订阅同步成功并生成节点结果时触发。",
		Category:       "subscription",
		CategoryName:   "订阅同步",
		Severity:       "success",
		Channels:       []Channel{ChannelWebhook, ChannelTelegram, ChannelInApp},
		DefaultEnabled: true,
	},
	{
		Key:            "subscription.sync_failed",
		Name:           "订阅更新失败",
		Description:    "机场订阅拉取、解析或写入失败时触发。",
		Category:       "subscription",
		CategoryName:   "订阅同步",
		Severity:       "error",
		Channels:       []Channel{ChannelWebhook, ChannelTelegram, ChannelInApp},
		DefaultEnabled: true,
	},
	{
		Key:            "task.speed_test_completed",
		Name:           "节点测速完成",
		Description:    "测速任务完成并生成统计结果时触发。",
		Category:       "task",
		CategoryName:   "任务执行",
		Severity:       "success",
		Channels:       []Channel{ChannelWebhook, ChannelTelegram, ChannelInApp},
		DefaultEnabled: true,
	},
	{
		Key:            "task.tag_rule_completed",
		Name:           "标签规则执行完成",
		Description:    "手动执行标签规则完成时触发。",
		Category:       "automation",
		CategoryName:   "自动化",
		Severity:       "success",
		Channels:       []Channel{ChannelWebhook, ChannelTelegram, ChannelInApp},
		DefaultEnabled: true,
	},
	{
		Key:            "task.auto_tag_completed",
		Name:           "自动标签完成",
		Description:    "测速或订阅同步后自动应用标签规则时触发。",
		Category:       "automation",
		CategoryName:   "自动化",
		Severity:       "success",
		Channels:       []Channel{ChannelWebhook, ChannelTelegram, ChannelInApp},
		DefaultEnabled: true,
	},
	{
		Key:            "security.user_login",
		Name:           "用户登录",
		Description:    "后台用户成功登录时触发。",
		Category:       "security",
		CategoryName:   "安全审计",
		Severity:       "info",
		Channels:       []Channel{ChannelWebhook, ChannelTelegram, ChannelInApp},
		DefaultEnabled: true,
	},
}

func EventCatalog() []EventDefinition {
	cloned := make([]EventDefinition, len(eventCatalog))
	copy(cloned, eventCatalog)
	return cloned
}

func EventCatalogForChannel(channel Channel) []EventDefinition {
	filtered := make([]EventDefinition, 0, len(eventCatalog))
	for _, event := range eventCatalog {
		if slices.Contains(event.Channels, channel) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func GetEventDefinition(eventKey string) (EventDefinition, bool) {
	for _, event := range eventCatalog {
		if event.Key == eventKey {
			return event, true
		}
	}
	return EventDefinition{}, false
}

func DefaultEventKeys(channel Channel) []string {
	keys := make([]string, 0, len(eventCatalog))
	for _, event := range eventCatalog {
		if event.DefaultEnabled && slices.Contains(event.Channels, channel) {
			keys = append(keys, event.Key)
		}
	}
	return keys
}

func NormalizeEventKeys(channel Channel, keys []string) []string {
	if len(keys) == 0 {
		return []string{}
	}

	selected := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		selected[key] = struct{}{}
	}

	normalized := make([]string, 0, len(keys))
	for _, event := range eventCatalog {
		if !slices.Contains(event.Channels, channel) {
			continue
		}
		if _, ok := selected[event.Key]; ok {
			normalized = append(normalized, event.Key)
		}
	}

	return normalized
}

func IsEventEnabled(selectedKeys []string, eventKey string) bool {
	return slices.Contains(selectedKeys, eventKey)
}

func FillPayloadMeta(eventKey string, payload Payload) Payload {
	if payload.Time == "" {
		payload.Time = nowString()
	}

	if payload.Event == "" {
		payload.Event = eventKey
	}

	if event, ok := GetEventDefinition(eventKey); ok {
		if payload.EventName == "" {
			payload.EventName = event.Name
		}
		if payload.Category == "" {
			payload.Category = event.Category
		}
		if payload.CategoryName == "" {
			payload.CategoryName = event.CategoryName
		}
		if payload.Severity == "" {
			payload.Severity = event.Severity
		}
	}

	if payload.Severity == "" {
		payload.Severity = "info"
	}

	return payload
}
