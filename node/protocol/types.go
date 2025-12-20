package protocol

// OutputConfig 订阅输出配置
// 控制 Clash/Surge 等客户端配置的生成参数
type OutputConfig struct {
	Clash                 string             `json:"clash"`                 // Clash 模板路径或 URL
	Surge                 string             `json:"surge"`                 // Surge 模板路径或 URL
	Udp                   bool               `json:"udp"`                   // 是否启用 UDP
	Cert                  bool               `json:"cert"`                  // 是否跳过证书验证
	ReplaceServerWithHost bool               `json:"replaceServerWithHost"` // 是否使用 Host 替换服务器地址
	HostMap               map[string]string  `json:"-"`                     // 运行时填充的 Host 映射，不序列化
	CustomProxyGroups     []CustomProxyGroup `json:"-"`                     // 运行时填充的自定义代理组，不序列化
}

// CustomProxyGroup 自定义代理组（由链式代理规则生成）
type CustomProxyGroup struct {
	Name      string   `json:"name"`                // 代理组名称
	Type      string   `json:"type"`                // select, url-test
	Proxies   []string `json:"proxies"`             // 代理节点列表
	URL       string   `json:"url,omitempty"`       // 测速 URL (url-test)
	Interval  int      `json:"interval,omitempty"`  // 测速间隔 (url-test)
	Tolerance int      `json:"tolerance,omitempty"` // 容差 (url-test)
}
