package protocol

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sublink/cache"
	"sublink/utils"

	"gopkg.in/yaml.v3"
)

// FlexPort 是一个可以从 int 或 string 类型解析的端口类型
// 用于处理订阅源返回的 port 字段可能是 int 或 string 的情况
type FlexPort int

// UnmarshalYAML 实现 yaml.Unmarshaler 接口，支持从 int 或 string 解析
func (fp *FlexPort) UnmarshalYAML(value *yaml.Node) error {
	var intVal int
	if err := value.Decode(&intVal); err == nil {
		*fp = FlexPort(intVal)
		return nil
	}

	var strVal string
	if err := value.Decode(&strVal); err == nil {
		if strVal == "" {
			*fp = 0
			return nil
		}
		intVal, err := strconv.Atoi(strVal)
		if err != nil {
			return fmt.Errorf("无法将端口 '%s' 转换为整数: %w", strVal, err)
		}
		*fp = FlexPort(intVal)
		return nil
	}

	return fmt.Errorf("无法解析端口值")
}

// MarshalYAML 实现 yaml.Marshaler 接口，始终输出为 int
func (fp FlexPort) MarshalYAML() (interface{}, error) {
	return int(fp), nil
}

// Int 返回端口的 int 值
func (fp FlexPort) Int() int {
	return int(fp)
}

type Proxy struct {
	Name                  string                 `yaml:"name,omitempty"`                  // 节点名称
	Type                  string                 `yaml:"type,omitempty"`                  // 代理类型 (ss, vmess, trojan, etc.)
	Server                string                 `yaml:"server,omitempty"`                // 服务器地址
	Port                  FlexPort               `yaml:"port,omitempty"`                  // 服务器端口
	Ports                 string                 `yaml:"ports,omitempty"`                 // hysteria2端口跳跃
	Cipher                string                 `yaml:"cipher,omitempty"`                // 加密方式
	Username              string                 `yaml:"username,omitempty"`              // 用户名 (socks5 等)
	Password              string                 `yaml:"password,omitempty"`              // 密码
	Client_fingerprint    string                 `yaml:"client-fingerprint,omitempty"`    // 客户端指纹 (uTLS)
	Tfo                   bool                   `yaml:"tfo,omitempty"`                   // TCP Fast Open
	Udp                   bool                   `yaml:"udp,omitempty"`                   // 是否启用 UDP
	Skip_cert_verify      bool                   `yaml:"skip-cert-verify,omitempty"`      // 跳过证书验证
	Tls                   bool                   `yaml:"tls,omitempty"`                   // 是否启用 TLS
	Servername            string                 `yaml:"servername,omitempty"`            // TLS SNI
	Flow                  string                 `yaml:"flow,omitempty"`                  // 流控 (xtls-rprx-vision 等)
	AlterId               string                 `yaml:"alterId,omitempty"`               // VMess AlterId
	Network               string                 `yaml:"network,omitempty"`               // 传输协议 (ws, grpc, etc.)
	Reality_opts          map[string]interface{} `yaml:"reality-opts,omitempty"`          // Reality 选项
	Ws_opts               map[string]interface{} `yaml:"ws-opts,omitempty"`               // WebSocket 选项
	Grpc_opts             map[string]interface{} `yaml:"grpc-opts,omitempty"`             // gRPC 选项
	Auth_str              string                 `yaml:"auth-str,omitempty"`              // Hysteria 认证字符串
	Auth                  string                 `yaml:"auth,omitempty"`                  // 认证信息
	Up                    int                    `yaml:"up,omitempty"`                    // 上行带宽限制
	Down                  int                    `yaml:"down,omitempty"`                  // 下行带宽限制
	Up_Speed              int                    `yaml:"up-speed,omitempty"`              // 上行带宽限制兼容stash
	Down_Speed            int                    `yaml:"down-speed,omitempty"`            // 下行带宽限制兼容stash
	Alpn                  []string               `yaml:"alpn,omitempty"`                  // ALPN
	Sni                   string                 `yaml:"sni,omitempty"`                   // SNI
	Obfs                  string                 `yaml:"obfs,omitempty"`                  // 混淆模式 (SSR/Hysteria2)
	Obfs_password         string                 `yaml:"obfs-password,omitempty"`         // 混淆密码
	Protocol              string                 `yaml:"protocol,omitempty"`              // SSR 协议
	Uuid                  string                 `yaml:"uuid,omitempty"`                  // UUID (VMess/VLESS)
	Peer                  string                 `yaml:"peer,omitempty"`                  // Peer (Hysteria)
	Congestion_controller string                 `yaml:"congestion-controller,omitempty"` // 拥塞控制 (Tuic)
	Udp_relay_mode        string                 `yaml:"udp-relay-mode,omitempty"`        // UDP 转发模式 (Tuic)
	Disable_sni           bool                   `yaml:"disable-sni,omitempty"`           // 禁用 SNI (Tuic)
	Dialer_proxy          string                 `yaml:"dialer-proxy,omitempty"`          // 前置代理
	// SS 插件字段
	Plugin      string                 `yaml:"plugin,omitempty"`      // SS 插件名称
	Plugin_opts map[string]interface{} `yaml:"plugin-opts,omitempty"` // SS 插件选项
	// WireGuard 特有字段
	Private_key    string   `yaml:"private-key,omitempty"`    // WireGuard 私钥
	Public_key     string   `yaml:"public-key,omitempty"`     // WireGuard 公钥
	Pre_shared_key string   `yaml:"pre-shared-key,omitempty"` // WireGuard 预共享密钥（可选）
	Ip             string   `yaml:"ip,omitempty"`             // WireGuard 客户端 IPv4
	Ipv6           string   `yaml:"ipv6,omitempty"`           // WireGuard 客户端 IPv6
	Mtu            int      `yaml:"mtu,omitempty"`            // MTU 值
	Reserved       []int    `yaml:"reserved,omitempty"`       // 保留字段
	Allowed_ips    []string `yaml:"allowed-ips,omitempty"`    // 允许的 IP 段
	Version        int      `yaml:"version,omitempty"`        // 版本
	Token          string   `yaml:"token,omitempty"`          // Tuic 令牌v4
	// VLESS 特有字段
	Packet_encoding string                 `yaml:"packet-encoding,omitempty"` // VLESS packet-encoding (xudp/packetaddr)
	H2_opts         map[string]interface{} `yaml:"h2-opts,omitempty"`         // HTTP/2 传输层选项
	Http_opts       map[string]interface{} `yaml:"http-opts,omitempty"`       // HTTP 传输层选项
	XHTTP_opts      map[string]interface{} `yaml:"xhttp-opts,omitempty"`
}

type ProxyGroup struct {
	Proxies []string `yaml:"proxies"`
}
type Config struct {
	Proxies      []Proxy      `yaml:"proxies"`
	Proxy_groups []ProxyGroup `yaml:"proxy-groups"`
}

// 代理链接的结构体
type Urls struct {
	Url             string
	DialerProxyName string
}

// 删除opts中的空值
func DeleteOpts(opts map[string]interface{}) {
	for k, v := range opts {
		switch v := v.(type) {
		case string:
			if v == "" {
				delete(opts, k)
			}
		case map[string]interface{}:
			DeleteOpts(v)
			if len(v) == 0 {
				delete(opts, k)
			}
		}
	}
}
func convertToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("unexpected type %T", v)
	}
}

func isTruthyConfigValue(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			return true
		}
	}

	return false
}

// shouldPreserveProxyGroup 判断代理组是否应保留模板原始语义，而不是在服务端展开成固定节点列表。
func shouldPreserveProxyGroup(proxyGroup map[string]interface{}) bool {
	for _, field := range []string{"include-all", "include-all-proxies", "include-all-providers"} {
		if isTruthyConfigValue(proxyGroup[field]) {
			return true
		}
	}

	// Provider 组和自动匹配组由客户端在运行时解析，不能在服务端展开为固定节点列表。
	for _, field := range []string{"use", "filter", "exclude-filter", "exclude-type", "expected-status"} {
		if _, exists := proxyGroup[field]; exists {
			return true
		}
	}

	return false
}

// convertSSPluginOpts 将 SsPlugin 转换为 Clash 格式的 plugin-opts
// 根据不同插件类型生成对应的配置
func convertSSPluginOpts(plugin SsPlugin) map[string]interface{} {
	if plugin.Name == "" {
		return nil
	}

	opts := make(map[string]interface{})

	// 从结构体字段读取值
	if plugin.Mode != "" {
		opts["mode"] = plugin.Mode
	}
	if plugin.Host != "" {
		opts["host"] = plugin.Host
	}
	if plugin.Path != "" {
		opts["path"] = plugin.Path
	}
	if plugin.Tls {
		opts["tls"] = true
	}
	if plugin.Mux {
		opts["mux"] = true
	}
	if plugin.Password != "" {
		opts["password"] = plugin.Password
	}
	if plugin.Version > 0 {
		opts["version"] = plugin.Version
	}

	if len(opts) == 0 {
		return nil
	}
	return opts
}

// LinkToProxy 将单个节点链接转换为 Proxy 结构体
// 支持 ss, ssr, trojan, vmess, vless, hysteria, hysteria2, tuic, anytls, socks5, http, https 等协议
func LinkToProxy(link Urls, config OutputConfig) (Proxy, error) {
	protocol := detectProtocol(link.Url)
	if protocol == nil {
		return Proxy{}, fmt.Errorf("unsupported scheme: %s", strings.ToLower(strings.Split(link.Url, "://")[0]))
	}
	proxyCapable, ok := protocol.(ProxyCapable)
	if !ok {
		return Proxy{}, fmt.Errorf("protocol %s does not support proxy export", protocol.Name())
	}
	return proxyCapable.ToProxy(link, config)
}

// EncodeClash 用于生成 Clash 配置文件
// 输入: 节点链接列表, SQL配置
// 输出: Clash 配置文件的 YAML 字节流
func EncodeClash(urls []Urls, config OutputConfig) ([]byte, error) {
	// 传入urls，解析urls，生成proxys
	// yamlfile 为模板文件
	var proxys []Proxy

	for _, link := range urls {
		proxy, err := LinkToProxy(link, config)
		if err != nil {
			utils.Error("链接转换失败: %s", err.Error())
			continue
		}
		proxys = append(proxys, proxy)
	}

	// 根据配置执行 Host 替换
	if config.ReplaceServerWithHost && len(config.HostMap) > 0 {
		for i := range proxys {
			if ip, exists := config.HostMap[proxys[i].Server]; exists {
				proxys[i].Server = ip
			}
		}
	}

	// 生成Clash配置文件
	return DecodeClash(proxys, config.Clash, config.CustomProxyGroups)
}

// DecodeClash 用于解析 Clash 配置文件并合并新节点
// proxys: 新增的节点列表
// yamlfile: 模板文件路径或 URL
// customGroups: 自定义代理组列表（可选，由链式代理规则生成）
func DecodeClash(proxys []Proxy, yamlfile string, customGroups ...[]CustomProxyGroup) ([]byte, error) {
	// 读取 YAML 文件
	var data []byte
	var err error
	if strings.Contains(yamlfile, "://") {
		resp, err := http.Get(yamlfile)
		if err != nil {
			utils.Error("http.Get error: %v", err)
			return nil, err
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			utils.Error("error: %v", err)
			return nil, err
		}
	} else {
		// 优先从缓存读取模板内容（本地文件使用缓存）
		filename := filepath.Base(yamlfile)
		if cached, ok := cache.GetTemplateContent(filename); ok {
			data = []byte(cached)
		} else {
			data, err = os.ReadFile(yamlfile)
			if err != nil {
				utils.Error("error: %v", err)
				return nil, err
			}
			// 写入缓存
			cache.SetTemplateContent(filename, string(data))
		}
	}
	// 解析 YAML 文件
	config := make(map[string]interface{})
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		utils.Error("error: %v", err)
		return nil, err
	}

	// 检查 "proxies" 键是否存在于 config 中
	proxies, ok := config["proxies"].([]interface{})
	if !ok {
		// 如果 "proxies" 键不存在，创建一个新的切片
		proxies = []interface{}{}
	}
	// 定义一个代理列表名字
	ProxiesNameList := []string{}
	// 添加新代理
	for _, p := range proxys {
		ProxiesNameList = append(ProxiesNameList, p.Name)
		proxies = append(proxies, p)
	}
	// proxies = append(proxies, newProxy)
	config["proxies"] = proxies
	// 往ProxyGroup中插入代理列表
	proxyGroups, ok := config["proxy-groups"].([]interface{})
	if !ok {
		proxyGroups = []interface{}{}
	}

	// 插入自定义代理组（在模板组之后）
	// 使用 _custom_group 标记来标识自定义代理组，后续循环时跳过节点追加
	if len(customGroups) > 0 && len(customGroups[0]) > 0 {
		for _, cg := range customGroups[0] {
			// 构建代理组 map
			groupMap := map[string]interface{}{
				"name":          cg.Name,
				"type":          cg.Type,
				"proxies":       cg.Proxies,
				"_custom_group": true, // 标记为自定义代理组，不追加所有节点
			}

			// 根据组类型添加相应配置
			switch cg.Type {
			case "url-test", "fallback":
				// url-test 和 fallback 类型需要 url、interval、tolerance
				if cg.URL != "" {
					groupMap["url"] = cg.URL
				} else {
					groupMap["url"] = "http://www.gstatic.com/generate_204"
				}
				if cg.Interval > 0 {
					groupMap["interval"] = cg.Interval
				} else {
					groupMap["interval"] = 300 // 默认 300 秒
				}
				if cg.Tolerance > 0 {
					groupMap["tolerance"] = cg.Tolerance
				} else {
					groupMap["tolerance"] = 50 // 默认 50 毫秒
				}

			case "load-balance":
				// load-balance 类型需要 url、interval、strategy
				if cg.URL != "" {
					groupMap["url"] = cg.URL
				} else {
					groupMap["url"] = "http://www.gstatic.com/generate_204"
				}
				if cg.Interval > 0 {
					groupMap["interval"] = cg.Interval
				} else {
					groupMap["interval"] = 300 // 默认 300 秒
				}
				if cg.Strategy != "" {
					groupMap["strategy"] = cg.Strategy
				} else {
					groupMap["strategy"] = "consistent-hashing" // 默认一致性哈希
				}
			}

			proxyGroups = append(proxyGroups, groupMap)
		}
	}

	for i, pg := range proxyGroups {
		proxyGroup, ok := pg.(map[string]interface{})
		if !ok {
			continue
		}

		// 链式代理不处理
		if proxyGroup["type"] == "relay" {
			continue
		}

		// 自定义代理组（由链式代理规则生成）已有自己的节点列表，跳过节点追加
		if isCustom, ok := proxyGroup["_custom_group"].(bool); ok && isCustom {
			// 删除内部标记，避免输出到配置文件
			delete(proxyGroup, "_custom_group")
			proxyGroups[i] = proxyGroup
			continue
		}

		// include-all、use/filter 等自动匹配组应保持模板原意，不能强行展开为固定节点列表。
		if shouldPreserveProxyGroup(proxyGroup) {
			continue
		}

		// 获取现有的 proxies 列表
		var existingProxies []interface{}
		if proxyGroup["proxies"] != nil {
			existingProxies, _ = proxyGroup["proxies"].([]interface{})
		}

		// 检查是否包含 __ALL_PROXIES__ 占位符（与 subconverter 行为一致）
		// 如果有占位符，将其替换为所有节点
		hasPlaceholder := false
		placeholderIndex := -1
		for idx, proxy := range existingProxies {
			if proxyStr, ok := proxy.(string); ok && proxyStr == "__ALL_PROXIES__" {
				hasPlaceholder = true
				placeholderIndex = idx
				break
			}
		}

		if hasPlaceholder {
			// 构建新的 proxies 列表：占位符之前的元素 + 所有节点
			var newProxies []interface{}
			// 添加占位符之前的元素（组引用如 🔯 故障转移、♻️ 自动选择、DIRECT 等）
			for j := 0; j < placeholderIndex; j++ {
				newProxies = append(newProxies, existingProxies[j])
			}
			// 添加所有节点
			for _, newProxy := range ProxiesNameList {
				newProxies = append(newProxies, newProxy)
			}
			// 添加占位符之后的元素（如果有的话）
			for j := placeholderIndex + 1; j < len(existingProxies); j++ {
				newProxies = append(newProxies, existingProxies[j])
			}
			proxyGroup["proxies"] = newProxies
			proxyGroups[i] = proxyGroup
			continue
		}

		// 原有逻辑：只有当 proxies 列表为空时才追加所有节点
		// 如果已有 proxies（组引用如 🚀 节点选择、DIRECT 等），保持不变
		if len(existingProxies) == 0 {
			// 没有任何 proxies，追加所有节点
			var validProxies []interface{}
			for _, newProxy := range ProxiesNameList {
				validProxies = append(validProxies, newProxy)
			}
			// 如果仍然为空，插入 DIRECT 作为后备
			if len(validProxies) == 0 {
				validProxies = append(validProxies, "DIRECT")
			}
			proxyGroup["proxies"] = validProxies
			proxyGroups[i] = proxyGroup
		}
		// 已有 proxies 的组保持不变
	}

	config["proxy-groups"] = proxyGroups

	// 将修改后的内容写回文件
	newData, err := yaml.Marshal(config)
	if err != nil {
		utils.Error("error: %v", err)
	}
	return newData, nil
}
