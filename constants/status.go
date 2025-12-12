package constants

// ==================== 节点测试状态常量 ====================
// 集中管理所有状态值，便于统一维护
// 前端 utils.js 中的 NODE_STATUS 需要与此保持同步

// 状态值常量
const (
	StatusUntested = "untested" // 未测试
	StatusSuccess  = "success"  // 成功
	StatusTimeout  = "timeout"  // 超时（连接超时）
	StatusError    = "error"    // 错误（连接失败、解析错误等）
)

// StatusLabels 状态显示文本（中文）
// 用于后端日志或API响应
var StatusLabels = map[string]string{
	StatusUntested: "未测速",
	StatusSuccess:  "成功",
	StatusTimeout:  "超时",
	StatusError:    "失败",
}

// GetStatusLabel 获取状态的显示文本
func GetStatusLabel(status string) string {
	if label, ok := StatusLabels[status]; ok {
		return label
	}
	return status
}

// IsValidStatus 检查状态值是否有效
func IsValidStatus(status string) bool {
	switch status {
	case StatusUntested, StatusSuccess, StatusTimeout, StatusError, "":
		return true
	default:
		return false
	}
}

// AllStatuses 获取所有有效状态值列表
func AllStatuses() []string {
	return []string{StatusUntested, StatusSuccess, StatusTimeout, StatusError}
}
