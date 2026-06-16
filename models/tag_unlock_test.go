package models

import (
	"testing"
)

// TestEvaluateUnlockCondition_SingleProvider 测试单个provider的匹配
func TestEvaluateUnlockCondition_SingleProvider(t *testing.T) {
	node := Node{
		UnlockSummary: `{"providers":[{"provider":"OpenAI","status":"available","region":"US"}]}`,
	}

	tests := []struct {
		name      string
		condition TagCondition
		expected  bool
	}{
		{
			name: "匹配provider-equals",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "OpenAI",
			},
			expected: true,
		},
		{
			name: "匹配provider-小写",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "openai",
			},
			expected: true,
		},
		{
			name: "不匹配provider",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "Netflix",
			},
			expected: false,
		},
		{
			name: "匹配provider-contains",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "contains",
				Value:    "Open",
			},
			expected: true,
		},
		{
			name: "匹配status",
			condition: TagCondition{
				Field:    "unlock_status",
				Operator: "equals",
				Value:    "available",
			},
			expected: true,
		},
		{
			name: "不匹配status",
			condition: TagCondition{
				Field:    "unlock_status",
				Operator: "equals",
				Value:    "restricted",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateCondition(node, tt.condition)
			if result != tt.expected {
				t.Errorf("evaluateCondition() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestEvaluateUnlockCondition_MultipleProviders 测试多个providers的匹配
func TestEvaluateUnlockCondition_MultipleProviders(t *testing.T) {
	node := Node{
		UnlockSummary: `{"providers":[
			{"provider":"Netflix","status":"available","region":"US"},
			{"provider":"Disney+","status":"restricted","region":"CN"},
			{"provider":"OpenAI","status":"available","region":"US"}
		]}`,
	}

	tests := []struct {
		name      string
		condition TagCondition
		expected  bool
	}{
		{
			name: "匹配第一个provider",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "Netflix",
			},
			expected: true,
		},
		{
			name: "匹配最后一个provider",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "OpenAI",
			},
			expected: true,
		},
		{
			name: "匹配中间的provider",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "Disney+",
			},
			expected: true,
		},
		{
			name: "不存在的provider",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "Hulu",
			},
			expected: false,
		},
		{
			name: "匹配任意available状态",
			condition: TagCondition{
				Field:    "unlock_status",
				Operator: "equals",
				Value:    "available",
			},
			expected: true,
		},
		{
			name: "匹配任意restricted状态",
			condition: TagCondition{
				Field:    "unlock_status",
				Operator: "equals",
				Value:    "restricted",
			},
			expected: true,
		},
		{
			name: "不存在的状态",
			condition: TagCondition{
				Field:    "unlock_status",
				Operator: "equals",
				Value:    "unsupported",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateCondition(node, tt.condition)
			if result != tt.expected {
				t.Errorf("evaluateCondition() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestEvaluateUnlockCondition_CombinedConditions 测试组合条件（AND逻辑）
// 注意：当前实现是独立匹配，即provider和status条件独立评估
// 只要节点中存在匹配的provider AND 存在匹配的status（可以是不同的provider），就认为匹配
func TestEvaluateUnlockCondition_CombinedConditions(t *testing.T) {
	tests := []struct {
		name          string
		unlockSummary string
		conditions    TagConditions
		expected      bool
		reason        string
	}{
		{
			name: "OpenAI可用-应该匹配",
			unlockSummary: `{"providers":[
				{"provider":"Netflix","status":"restricted","region":"CN"},
				{"provider":"OpenAI","status":"available","region":"US"}
			]}`,
			conditions: TagConditions{
				Logic: "and",
				Conditions: []TagCondition{
					{Field: "unlock_provider", Operator: "equals", Value: "OpenAI"},
					{Field: "unlock_status", Operator: "equals", Value: "available"},
				},
			},
			expected: true,
			reason:   "节点有OpenAI且有available状态",
		},
		{
			name: "Netflix受限-应该匹配",
			unlockSummary: `{"providers":[
				{"provider":"Netflix","status":"restricted","region":"CN"},
				{"provider":"OpenAI","status":"available","region":"US"}
			]}`,
			conditions: TagConditions{
				Logic: "and",
				Conditions: []TagCondition{
					{Field: "unlock_provider", Operator: "equals", Value: "Netflix"},
					{Field: "unlock_status", Operator: "equals", Value: "restricted"},
				},
			},
			expected: true,
			reason:   "节点有Netflix且有restricted状态",
		},
		{
			name: "OpenAI受限-也会匹配（独立匹配逻辑）",
			unlockSummary: `{"providers":[
				{"provider":"Netflix","status":"restricted","region":"CN"},
				{"provider":"OpenAI","status":"available","region":"US"}
			]}`,
			conditions: TagConditions{
				Logic: "and",
				Conditions: []TagCondition{
					{Field: "unlock_provider", Operator: "equals", Value: "OpenAI"},
					{Field: "unlock_status", Operator: "equals", Value: "restricted"},
				},
			},
			expected: true,
			reason:   "节点有OpenAI（满足第一个条件）且有restricted状态（Netflix，满足第二个条件）",
		},
		{
			name: "Disney可用-不应该匹配",
			unlockSummary: `{"providers":[
				{"provider":"Netflix","status":"restricted","region":"CN"},
				{"provider":"OpenAI","status":"available","region":"US"}
			]}`,
			conditions: TagConditions{
				Logic: "and",
				Conditions: []TagCondition{
					{Field: "unlock_provider", Operator: "equals", Value: "Disney+"},
					{Field: "unlock_status", Operator: "equals", Value: "available"},
				},
			},
			expected: false,
			reason:   "节点没有Disney+",
		},
		{
			name: "OpenAI不支持-不应该匹配",
			unlockSummary: `{"providers":[
				{"provider":"Netflix","status":"restricted","region":"CN"},
				{"provider":"OpenAI","status":"available","region":"US"}
			]}`,
			conditions: TagConditions{
				Logic: "and",
				Conditions: []TagCondition{
					{Field: "unlock_provider", Operator: "equals", Value: "OpenAI"},
					{Field: "unlock_status", Operator: "equals", Value: "unsupported"},
				},
			},
			expected: false,
			reason:   "节点没有unsupported状态",
		},
		{
			name: "只有OpenAI可用-严格匹配",
			unlockSummary: `{"providers":[
				{"provider":"OpenAI","status":"available","region":"US"}
			]}`,
			conditions: TagConditions{
				Logic: "and",
				Conditions: []TagCondition{
					{Field: "unlock_provider", Operator: "equals", Value: "OpenAI"},
					{Field: "unlock_status", Operator: "equals", Value: "available"},
				},
			},
			expected: true,
			reason:   "只有一个provider时，行为符合预期",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := Node{UnlockSummary: tt.unlockSummary}
			result := tt.conditions.EvaluateNode(node)
			if result != tt.expected {
				t.Errorf("EvaluateNode() = %v, expected %v, reason: %s", result, tt.expected, tt.reason)
			}
		})
	}
}

// TestEvaluateUnlockCondition_ORLogic 测试OR逻辑
func TestEvaluateUnlockCondition_ORLogic(t *testing.T) {
	node := Node{
		UnlockSummary: `{"providers":[
			{"provider":"OpenAI","status":"available","region":"US"}
		]}`,
	}

	tests := []struct {
		name       string
		conditions TagConditions
		expected   bool
	}{
		{
			name: "匹配第一个条件",
			conditions: TagConditions{
				Logic: "or",
				Conditions: []TagCondition{
					{Field: "unlock_provider", Operator: "equals", Value: "OpenAI"},
					{Field: "unlock_provider", Operator: "equals", Value: "Netflix"},
				},
			},
			expected: true,
		},
		{
			name: "匹配第二个条件",
			conditions: TagConditions{
				Logic: "or",
				Conditions: []TagCondition{
					{Field: "unlock_provider", Operator: "equals", Value: "Netflix"},
					{Field: "unlock_provider", Operator: "equals", Value: "OpenAI"},
				},
			},
			expected: true,
		},
		{
			name: "都不匹配",
			conditions: TagConditions{
				Logic: "or",
				Conditions: []TagCondition{
					{Field: "unlock_provider", Operator: "equals", Value: "Netflix"},
					{Field: "unlock_provider", Operator: "equals", Value: "Disney+"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.conditions.EvaluateNode(node)
			if result != tt.expected {
				t.Errorf("EvaluateNode() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestEvaluateUnlockCondition_EmptyUnlockSummary 测试空的UnlockSummary
func TestEvaluateUnlockCondition_EmptyUnlockSummary(t *testing.T) {
	tests := []struct {
		name          string
		unlockSummary string
		condition     TagCondition
		expected      bool
	}{
		{
			name:          "空字符串",
			unlockSummary: "",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "OpenAI",
			},
			expected: false,
		},
		{
			name:          "空JSON对象",
			unlockSummary: "{}",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "OpenAI",
			},
			expected: false,
		},
		{
			name:          "空providers数组",
			unlockSummary: `{"providers":[]}`,
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "OpenAI",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := Node{UnlockSummary: tt.unlockSummary}
			result := evaluateCondition(node, tt.condition)
			if result != tt.expected {
				t.Errorf("evaluateCondition() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestEvaluateUnlockCondition_NotEqualsOperator 测试not_equals操作符
func TestEvaluateUnlockCondition_NotEqualsOperator(t *testing.T) {
	node := Node{
		UnlockSummary: `{"providers":[
			{"provider":"Netflix","status":"available","region":"US"},
			{"provider":"OpenAI","status":"available","region":"US"}
		]}`,
	}

	tests := []struct {
		name      string
		condition TagCondition
		expected  bool
	}{
		{
			name: "不等于Netflix-应该匹配OpenAI",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "not_equals",
				Value:    "Netflix",
			},
			expected: true,
		},
		{
			name: "不等于Disney-应该匹配",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "not_equals",
				Value:    "Disney+",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateCondition(node, tt.condition)
			if result != tt.expected {
				t.Errorf("evaluateCondition() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestEvaluateUnlockCondition_ContainsOperator 测试contains操作符
func TestEvaluateUnlockCondition_ContainsOperator(t *testing.T) {
	node := Node{
		UnlockSummary: `{"providers":[
			{"provider":"Netflix","status":"available","region":"US"}
		]}`,
	}

	tests := []struct {
		name      string
		condition TagCondition
		expected  bool
	}{
		{
			name: "包含Net",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "contains",
				Value:    "Net",
			},
			expected: true,
		},
		{
			name: "包含flix",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "contains",
				Value:    "flix",
			},
			expected: true,
		},
		{
			name: "不包含Disney",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "contains",
				Value:    "Disney",
			},
			expected: false,
		},
		{
			name: "包含avail",
			condition: TagCondition{
				Field:    "unlock_status",
				Operator: "contains",
				Value:    "avail",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateCondition(node, tt.condition)
			if result != tt.expected {
				t.Errorf("evaluateCondition() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestEvaluateUnlockCondition_NotContainsOperator 测试not_contains操作符
func TestEvaluateUnlockCondition_NotContainsOperator(t *testing.T) {
	node := Node{
		UnlockSummary: `{"providers":[
			{"provider":"Netflix","status":"available","region":"US"}
		]}`,
	}

	tests := []struct {
		name      string
		condition TagCondition
		expected  bool
	}{
		{
			name: "不包含Disney-应该匹配",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "not_contains",
				Value:    "Disney",
			},
			expected: true,
		},
		{
			name: "不包含Net-不应该匹配",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "not_contains",
				Value:    "Net",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateCondition(node, tt.condition)
			if result != tt.expected {
				t.Errorf("evaluateCondition() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestEvaluateUnlockCondition_CaseSensitivity 测试大小写不敏感
func TestEvaluateUnlockCondition_CaseSensitivity(t *testing.T) {
	node := Node{
		UnlockSummary: `{"providers":[
			{"provider":"OpenAI","status":"Available","region":"US"}
		]}`,
	}

	tests := []struct {
		name      string
		condition TagCondition
		expected  bool
	}{
		{
			name: "小写provider",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "openai",
			},
			expected: true,
		},
		{
			name: "大写provider",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "OPENAI",
			},
			expected: true,
		},
		{
			name: "混合大小写provider",
			condition: TagCondition{
				Field:    "unlock_provider",
				Operator: "equals",
				Value:    "OpEnAi",
			},
			expected: true,
		},
		{
			name: "小写status",
			condition: TagCondition{
				Field:    "unlock_status",
				Operator: "equals",
				Value:    "available",
			},
			expected: true,
		},
		{
			name: "大写status",
			condition: TagCondition{
				Field:    "unlock_status",
				Operator: "equals",
				Value:    "AVAILABLE",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateCondition(node, tt.condition)
			if result != tt.expected {
				t.Errorf("evaluateCondition() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestEvaluateUnlockCondition_RealWorldScenario 测试真实场景
func TestEvaluateUnlockCondition_RealWorldScenario(t *testing.T) {
	// 模拟用户在issue中提到的场景
	node := Node{
		UnlockSummary: `{"providers":[
			{"provider":"Netflix","status":"available","region":"US"},
			{"provider":"Disney+","status":"available","region":"US"},
			{"provider":"OpenAI","status":"available","region":"US"}
		]}`,
	}

	// 用户创建的规则：OpenAI解锁
	conditions := TagConditions{
		Logic: "and",
		Conditions: []TagCondition{
			{Field: "unlock_provider", Operator: "equals", Value: "openai"},
			{Field: "unlock_status", Operator: "equals", Value: "available"},
		},
	}

	result := conditions.EvaluateNode(node)
	if !result {
		t.Errorf("真实场景测试失败：OpenAI解锁的节点应该被匹配到")
	}
}
