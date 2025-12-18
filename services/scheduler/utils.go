package scheduler

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// formatDuration 格式化时长为人类可读字符串
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f秒", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0f分%.0f秒", d.Minutes(), math.Mod(d.Seconds(), 60))
	}
	return fmt.Sprintf("%.0f时%.0f分", d.Hours(), math.Mod(d.Minutes(), 60))
}

// formatBytes 格式化字节数为人类可读格式
func formatBytes(bytes int64) string {
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

	// B, KB, MB, GB, TB
	units := []string{"B", "KB", "MB", "GB", "TB"}
	if exp >= len(units)-1 {
		exp = len(units) - 2
	}

	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp+1])
}

// formatNodeDisplayItem 格式化节点进度显示项（包含分组和来源信息）
// 格式: 节点名称 · 分组 · 来源（移动端友好的紧凑格式）
func formatNodeDisplayItem(name, group, source string) string {
	parts := []string{name}

	if group != "" {
		parts = append(parts, group)
	}

	// 来源处理：manual 显示为"手动添加"，空时不显示
	if source != "" {
		if source == "manual" {
			parts = append(parts, "手动添加")
		} else {
			parts = append(parts, source)
		}
	}

	return strings.Join(parts, " · ")
}
