package models

type NodeConditionValueOption struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type NodeConditionFieldMeta struct {
	Value        string                     `json:"value"`
	Label        string                     `json:"label"`
	DataType     string                     `json:"dataType,omitempty"`
	InputType    string                     `json:"inputType,omitempty"`
	Operators    []string                   `json:"operators,omitempty"`
	Options      []NodeConditionValueOption `json:"options,omitempty"`
	OptionSource string                     `json:"optionSource,omitempty"`
	Description  string                     `json:"description,omitempty"`
	Surfaces     []string                   `json:"surfaces,omitempty"`
}

var (
	nodeConditionStringOperators = []string{"equals", "not_equals", "contains", "not_contains", "regex"}
	nodeConditionNumberOperators = []string{"equals", "not_equals", "greater_than", "less_than", "greater_or_equal", "less_or_equal"}
	nodeConditionEnumOperators   = []string{"equals", "not_equals"}
)

func GetNodeConditionFields() []NodeConditionFieldMeta {
	fields := []NodeConditionFieldMeta{
		newTextConditionField("name", "节点名称", "按节点备注做文本匹配"),
		newTextConditionField("link_name", "原始名称", "按节点原始名称做文本匹配"),
		newTextConditionField("link_country", "国家/地区", "按节点国家/地区代码匹配"),
		newTextConditionField("protocol", "协议类型", "按节点协议类型匹配"),
		newTextConditionField("group", "分组", "按节点分组匹配"),
		newTextConditionField("source", "来源", "按节点来源匹配"),
		newNumberConditionField("speed", "速度 (MB/s)", "按节点测速结果做数值比较"),
		newNumberConditionField("delay_time", "延迟 (ms)", "按节点延迟结果做数值比较"),
		newNumberConditionField("fraud_score", "欺诈评分", "按节点欺诈评分做数值比较"),
		newEnumConditionField(
			"quality_status",
			"质量状态",
			"按节点质量状态筛选",
			[]NodeConditionValueOption{
				{Value: QualityStatusSuccess, Label: "优质"},
				{Value: QualityStatusPartial, Label: "一般"},
				{Value: QualityStatusFailed, Label: "差"},
				{Value: QualityStatusDisabled, Label: "未启用"},
				{Value: QualityStatusUntested, Label: "未检测"},
			},
		),
		{
			Value:        "unlock_provider",
			Label:        "解锁 Provider",
			DataType:     "enum",
			InputType:    "select",
			Operators:    cloneStringSlice(nodeConditionEnumOperators),
			OptionSource: "unlockProviders",
			Description:  "按节点主解锁结果对应的 Provider 匹配",
			Surfaces:     []string{"tag", "chain", "condition-builder"},
		},
		{
			Value:        "unlock_status",
			Label:        "解锁状态",
			DataType:     "enum",
			InputType:    "select",
			Operators:    cloneStringSlice(nodeConditionEnumOperators),
			OptionSource: "unlockStatuses",
			Description:  "按节点主解锁结果的状态匹配",
			Surfaces:     []string{"tag", "chain", "condition-builder"},
		},
		newTextConditionField("unlock_keyword", "解锁关键词", "在解锁结果的 Provider、状态、地区、原因、细节中做模糊匹配"),
		newTextConditionField("unlock_result", "解锁摘要", "按节点紧凑解锁摘要做文本匹配"),
		newEnumConditionField(
			"ip_type",
			"IP类型",
			"按节点 IP 类型筛选",
			[]NodeConditionValueOption{{Value: "native", Label: "原生IP"}, {Value: "broadcast", Label: "广播IP"}, {Value: "untested", Label: "未检测"}},
		),
		newEnumConditionField(
			"residential_type",
			"住宅属性",
			"按节点住宅属性筛选",
			[]NodeConditionValueOption{{Value: "residential", Label: "住宅IP"}, {Value: "datacenter", Label: "机房IP"}, {Value: "untested", Label: "未检测"}},
		),
		newEnumConditionField(
			"speed_status",
			"测速状态",
			"按测速状态筛选",
			[]NodeConditionValueOption{{Value: "untested", Label: "未测速"}, {Value: "success", Label: "成功"}, {Value: "timeout", Label: "超时"}, {Value: "error", Label: "失败"}},
		),
		newEnumConditionField(
			"delay_status",
			"延迟状态",
			"按延迟测试状态筛选",
			[]NodeConditionValueOption{{Value: "untested", Label: "未测速"}, {Value: "success", Label: "成功"}, {Value: "timeout", Label: "超时"}, {Value: "error", Label: "失败"}},
		),
		newTextConditionField("tags", "标签", "按节点标签做文本匹配"),
		newTextConditionField("link_address", "地址", "按节点地址匹配"),
		newTextConditionField("link_host", "主机名", "按节点主机名匹配"),
		newTextConditionField("link_port", "端口", "按节点端口匹配"),
		newTextConditionField("dialer_proxy_name", "前置代理", "按节点前置代理名称匹配"),
		newTextConditionField("link", "节点链接", "按节点完整链接匹配"),
	}

	return fields
}

func newTextConditionField(value string, label string, description string) NodeConditionFieldMeta {
	return NodeConditionFieldMeta{
		Value:       value,
		Label:       label,
		DataType:    "string",
		InputType:   "text",
		Operators:   cloneStringSlice(nodeConditionStringOperators),
		Description: description,
		Surfaces:    []string{"tag", "chain", "condition-builder"},
	}
}

func newNumberConditionField(value string, label string, description string) NodeConditionFieldMeta {
	return NodeConditionFieldMeta{
		Value:       value,
		Label:       label,
		DataType:    "number",
		InputType:   "text",
		Operators:   cloneStringSlice(nodeConditionNumberOperators),
		Description: description,
		Surfaces:    []string{"tag", "chain", "condition-builder"},
	}
}

func newEnumConditionField(value string, label string, description string, options []NodeConditionValueOption) NodeConditionFieldMeta {
	clonedOptions := make([]NodeConditionValueOption, len(options))
	copy(clonedOptions, options)
	return NodeConditionFieldMeta{
		Value:       value,
		Label:       label,
		DataType:    "enum",
		InputType:   "select",
		Operators:   cloneStringSlice(nodeConditionEnumOperators),
		Options:     clonedOptions,
		Description: description,
		Surfaces:    []string{"tag", "chain", "condition-builder"},
	}
}

func cloneStringSlice(items []string) []string {
	cloned := make([]string, len(items))
	copy(cloned, items)
	return cloned
}
