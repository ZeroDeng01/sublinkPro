package scheduler

import (
	"sublink/models"
	"time"
)

// SpeedTestConfig 测速任务配置（隔离各任务参数，避免并发覆盖）
// 每个检测任务拥有独立的配置实例，完全避免共享状态
type SpeedTestConfig struct {
	// 测速URL配置
	SpeedTestURL   string        // 速度测试URL
	LatencyTestURL string        // 延迟测试URL
	Timeout        time.Duration // 超时时间

	// 模式配置
	Mode          string // 检测模式：tcp / mihomo
	DetectCountry bool   // 是否检测落地IP国家
	LandingIPURL  string // 落地IP查询接口URL

	// 并发配置
	LatencyConcurrency int // 延迟测试并发数(0=自动)
	SpeedConcurrency   int // 速度测试并发数

	// 高级选项
	IncludeHandshake   bool   // 延迟是否包含握手时间
	SpeedRecordMode    string // 速度记录模式：average/peak
	PeakSampleInterval int    // 峰值采样间隔(ms)
	PersistHost        bool   // 是否持久化Host映射

	// 流量统计开关
	TrafficByGroup  bool // 按分组统计流量
	TrafficBySource bool // 按来源统计流量
	TrafficByNode   bool // 按节点统计流量
}

// SpeedTestConfigFromProfile 从策略构建配置（并发安全）
// 每次调用返回独立的配置实例，完全避免全局状态共享
func SpeedTestConfigFromProfile(profile *models.NodeCheckProfile) *SpeedTestConfig {
	// 超时时间转换
	timeout := time.Duration(profile.Timeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// 延迟测试URL（默认使用测速URL）
	latencyURL := profile.LatencyURL
	if latencyURL == "" {
		latencyURL = profile.TestURL
	}

	// 落地IP URL默认值
	landingIPURL := profile.LandingIPURL
	if landingIPURL == "" {
		landingIPURL = "https://api.ipify.org"
	}

	// 速度记录模式默认值
	speedRecordMode := profile.SpeedRecordMode
	if speedRecordMode == "" {
		speedRecordMode = "average"
	}

	// 峰值采样间隔默认值
	peakSampleInterval := profile.PeakSampleInterval
	if peakSampleInterval == 0 {
		peakSampleInterval = 100
	}

	return &SpeedTestConfig{
		SpeedTestURL:       profile.TestURL,
		LatencyTestURL:     latencyURL,
		Timeout:            timeout,
		Mode:               profile.Mode,
		DetectCountry:      profile.DetectCountry,
		LandingIPURL:       landingIPURL,
		LatencyConcurrency: profile.LatencyConcurrency,
		SpeedConcurrency:   profile.SpeedConcurrency,
		IncludeHandshake:   profile.IncludeHandshake,
		SpeedRecordMode:    speedRecordMode,
		PeakSampleInterval: peakSampleInterval,
		PersistHost:        false, // 持久化Host功能暂不开放
		TrafficByGroup:     profile.TrafficByGroup,
		TrafficBySource:    profile.TrafficBySource,
		TrafficByNode:      profile.TrafficByNode,
	}
}
