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
		newTextConditionField("name", "Node name", "Matches the node's effective outbound name as text"),
		newTextConditionField("effective_name", "Effective name", "Matches the effective outbound name computed from the node name pattern"),
		newTextConditionField("link_name", "Original name", "Matches the node's original name as text"),
		newTextConditionField("link_country", "Country/Region", "Matches the node's country/region code"),
		newTextConditionField("protocol", "Protocol", "Matches the node's protocol"),
		newTextConditionField("group", "Group", "Matches the node's group"),
		newTextConditionField("source", "Source", "Matches the node's source"),
		newNumberConditionField("speed", "Speed (MB/s)", "Compares the node's speed test result numerically"),
		newNumberConditionField("delay_time", "Latency (ms)", "Compares the node's latency result numerically"),
		newNumberConditionField("fraud_score", "Fraud score", "Compares the node's fraud score numerically"),
		newEnumConditionField(
			"quality_status",
			"Quality status",
			"Filters by the node's quality status",
			[]NodeConditionValueOption{
				{Value: QualityStatusSuccess, Label: "Excellent"},
				{Value: QualityStatusPartial, Label: "Fair"},
				{Value: QualityStatusFailed, Label: "Poor"},
				{Value: QualityStatusDisabled, Label: "Disabled"},
				{Value: QualityStatusUntested, Label: "Not checked"},
			},
		),
		{
			Value:        "unlock_provider",
			Label:        "Unlock provider",
			DataType:     "enum",
			InputType:    "select",
			Operators:    cloneStringSlice(nodeConditionEnumOperators),
			OptionSource: "unlockProviders",
			Description:  "Matches the provider of the node's primary unlock result",
			Surfaces:     []string{"tag", "chain", "condition-builder"},
		},
		{
			Value:        "unlock_status",
			Label:        "Unlock status",
			DataType:     "enum",
			InputType:    "select",
			Operators:    cloneStringSlice(nodeConditionEnumOperators),
			OptionSource: "unlockStatuses",
			Description:  "Matches the status of the node's primary unlock result",
			Surfaces:     []string{"tag", "chain", "condition-builder"},
		},
		newTextConditionField("unlock_keyword", "Unlock keyword", "Fuzzy-matches the provider, status, region, reason and details of unlock results"),
		newTextConditionField("unlock_result", "Unlock summary", "Matches the node's compact unlock summary as text"),
		newEnumConditionField(
			"ip_type",
			"IP type",
			"Filters by the node's IP type",
			[]NodeConditionValueOption{{Value: "native", Label: "Native IP"}, {Value: "broadcast", Label: "Broadcast IP"}, {Value: "untested", Label: "Not checked"}},
		),
		newEnumConditionField(
			"residential_type",
			"Residential type",
			"Filters by the node's residential type",
			[]NodeConditionValueOption{{Value: "residential", Label: "Residential IP"}, {Value: "datacenter", Label: "Datacenter IP"}, {Value: "untested", Label: "Not checked"}},
		),
		newEnumConditionField(
			"speed_status",
			"Speed test status",
			"Filters by speed test status",
			[]NodeConditionValueOption{{Value: "untested", Label: "Untested"}, {Value: "success", Label: "Success"}, {Value: "timeout", Label: "Timeout"}, {Value: "error", Label: "Failed"}},
		),
		newEnumConditionField(
			"delay_status",
			"Latency status",
			"Filters by latency test status",
			[]NodeConditionValueOption{{Value: "untested", Label: "Untested"}, {Value: "success", Label: "Success"}, {Value: "timeout", Label: "Timeout"}, {Value: "error", Label: "Failed"}},
		),
		newTextConditionField("tags", "Tags", "Matches the node's tags as text"),
		newTextConditionField("link_address", "Address", "Matches the node's address"),
		newTextConditionField("link_host", "Hostname", "Matches the node's hostname"),
		newTextConditionField("link_port", "Port", "Matches the node's port"),
		newTextConditionField("dialer_proxy_name", "Dialer proxy", "Matches the node's dialer proxy name"),
		newTextConditionField("link", "Node link", "Matches the node's full link"),
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
