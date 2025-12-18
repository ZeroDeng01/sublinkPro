package protocol

// OutputConfig 订阅输出配置
// 控制 Clash/Surge 等客户端配置的生成参数
type OutputConfig struct {
	Clash                 string            `json:"clash"`                 // Clash 模板路径或 URL
	Surge                 string            `json:"surge"`                 // Surge 模板路径或 URL
	Udp                   bool              `json:"udp"`                   // 是否启用 UDP
	Cert                  bool              `json:"cert"`                  // 是否跳过证书验证
	ReplaceServerWithHost bool              `json:"replaceServerWithHost"` // 是否使用 Host 替换服务器地址
	HostMap               map[string]string `json:"-"`                     // 运行时填充的 Host 映射，不序列化
}
