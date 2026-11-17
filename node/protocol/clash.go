package protocol

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
			return fmt.Errorf("cannot convert port '%s' to integer: %w", strVal, err)
		}
		*fp = FlexPort(intVal)
		return nil
	}

	return fmt.Errorf("cannot parse port value")
}

// MarshalYAML 实现 yaml.Marshaler 接口，始终输出为 int
func (fp FlexPort) MarshalYAML() (any, error) {
	return int(fp), nil
}

// Int 返回端口的 int 值
func (fp FlexPort) Int() int {
	return int(fp)
}

// IsZero lets YAML omitempty suppress absent ports for protocols that use
// alternative port fields such as Mieru port-range.
func (fp FlexPort) IsZero() bool {
	return fp == 0
}

// Mbps 是 Clash 兼容的带宽数值类型。
// 兼容 provider 里把 up/down 之类字段写成 "11 Mbps" 这种字符串的情况，但序列化时仍输出整数。
type Mbps int

var bandwidthRatePattern = regexp.MustCompile(`^(\d+)\s*([KMGTkmgt]?)([Bb])ps$`)

// UnmarshalYAML 支持把带宽字段解析为整数、裸数字字符串或标准速率字符串。
// 这里保留对 Hysteria/Hysteria2 兼容格式的支持，但拒绝非标准后缀，避免把非法带宽静默降级。
func (m *Mbps) UnmarshalYAML(value *yaml.Node) error {
	var intVal int
	if err := value.Decode(&intVal); err == nil {
		*m = Mbps(intVal)
		return nil
	}

	var strVal string
	if err := value.Decode(&strVal); err != nil {
		return fmt.Errorf("cannot parse bandwidth value")
	}
	strVal = strings.TrimSpace(strVal)
	if strVal == "" {
		*m = 0
		return nil
	}
	if directVal, err := strconv.Atoi(strVal); err == nil {
		*m = Mbps(directVal)
		return nil
	}
	matches := bandwidthRatePattern.FindStringSubmatch(strVal)
	if matches == nil {
		return fmt.Errorf("cannot parse bandwidth value %q", strVal)
	}
	base, err := strconv.Atoi(matches[1])
	if err != nil {
		return fmt.Errorf("cannot convert bandwidth value %q to integer: %w", strVal, err)
	}
	switch strings.ToUpper(matches[2]) {
	case "", "M":
		*m = Mbps(base)
		return nil
	case "G":
		*m = Mbps(base * 1000)
		return nil
	case "T":
		*m = Mbps(base * 1000 * 1000)
		return nil
	case "K":
		return fmt.Errorf("bandwidth value %q uses Kbps units; lossless conversion to Mbps is not possible", strVal)
	default:
		return fmt.Errorf("cannot parse bandwidth value %q", strVal)
	}
}

// MarshalYAML 始终把带宽写成整数，避免把内部类型泄漏成字符串。
func (m Mbps) MarshalYAML() (any, error) {
	return int(m), nil
}

type Proxy struct {
	Name                  string         `yaml:"name,omitempty"`                        // 节点名称
	Type                  string         `yaml:"type,omitempty"`                        // 代理类型 (ss, vmess, trojan, etc.)
	Server                string         `yaml:"server,omitempty"`                      // 服务器地址
	Port                  FlexPort       `yaml:"port,omitempty"`                        // 服务器端口
	PortRange             string         `yaml:"port-range,omitempty"`                  // 端口范围 (Mieru)
	Ports                 string         `yaml:"ports,omitempty"`                       // hysteria2端口跳跃
	HopInterval           int            `yaml:"hop-interval,omitempty"`                // hysteria2端口跳跃间隔 (秒)
	Transport             string         `yaml:"transport,omitempty"`                   // 传输协议 (Mieru TCP/UDP)
	Cipher                string         `yaml:"cipher,omitempty"`                      // 加密方式
	Username              string         `yaml:"username,omitempty"`                    // 用户名 (socks5 等)
	Password              string         `yaml:"password,omitempty"`                    // 密码
	Multiplexing          string         `yaml:"multiplexing,omitempty"`                // Mieru 多路复用级别
	TrafficPattern        string         `yaml:"traffic-pattern,omitempty"`             // Mieru 流量模式
	Client_fingerprint    string         `yaml:"client-fingerprint,omitempty"`          // 客户端指纹 (uTLS)
	Fingerprint           string         `yaml:"fingerprint,omitempty"`                 // 服务端证书 SHA-256 指纹
	Tfo                   bool           `yaml:"tfo,omitempty"`                         // TCP Fast Open
	Udp                   bool           `yaml:"udp,omitempty"`                         // 是否启用 UDP
	Skip_cert_verify      bool           `yaml:"skip-cert-verify,omitempty"`            // 跳过证书验证
	Tls                   bool           `yaml:"tls,omitempty"`                         // 是否启用 TLS
	Servername            string         `yaml:"servername,omitempty"`                  // TLS SNI
	Flow                  string         `yaml:"flow,omitempty"`                        // 流控 (xtls-rprx-vision 等)
	AlterId               string         `yaml:"alterId,omitempty"`                     // VMess AlterId
	Network               string         `yaml:"network,omitempty"`                     // 传输协议 (ws, grpc, etc.)
	Reality_opts          map[string]any `yaml:"reality-opts,omitempty"`                // Reality 选项
	Ws_opts               map[string]any `yaml:"ws-opts,omitempty"`                     // WebSocket 选项
	Grpc_opts             map[string]any `yaml:"grpc-opts,omitempty"`                   // gRPC 选项
	Auth_str              string         `yaml:"auth-str,omitempty"`                    // Hysteria 认证字符串
	Auth                  string         `yaml:"auth,omitempty"`                        // 认证信息
	Up                    Mbps           `yaml:"up,omitempty"`                          // 上行带宽限制
	Down                  Mbps           `yaml:"down,omitempty"`                        // 下行带宽限制
	AnyTLSIdleCheck       int            `yaml:"idle-session-check-interval,omitempty"` // AnyTLS 空闲会话检查间隔
	AnyTLSIdleTimeout     int            `yaml:"idle-session-timeout,omitempty"`        // AnyTLS 空闲会话超时时间
	AnyTLSMinIdle         int            `yaml:"min-idle-session,omitempty"`            // AnyTLS 最小空闲会话数
	Up_Speed              Mbps           `yaml:"up-speed,omitempty"`                    // 上行带宽限制兼容stash
	Down_Speed            Mbps           `yaml:"down-speed,omitempty"`                  // 下行带宽限制兼容stash
	Alpn                  []string       `yaml:"alpn,omitempty"`                        // ALPN
	Sni                   string         `yaml:"sni,omitempty"`                         // SNI
	Obfs                  string         `yaml:"obfs,omitempty"`                        // 混淆模式 (SSR/Hysteria2)
	Obfs_password         string         `yaml:"obfs-password,omitempty"`               // 混淆密码
	Psk                   string         `yaml:"psk,omitempty"`                         // Snell 预共享密钥
	Obfs_opts             map[string]any `yaml:"obfs-opts,omitempty"`                   // Snell 混淆选项 (mode/host)
	Protocol              string         `yaml:"protocol,omitempty"`                    // SSR 协议
	ProtocolParam         string         `yaml:"protocol-param,omitempty"`              // SSR 协议参数
	ObfsParam             string         `yaml:"obfs-param,omitempty"`                  // SSR 混淆参数
	Uuid                  string         `yaml:"uuid,omitempty"`                        // UUID (VMess/VLESS)
	Peer                  string         `yaml:"peer,omitempty"`                        // Peer (Hysteria)
	Congestion_controller string         `yaml:"congestion-controller,omitempty"`       // 拥塞控制 (Tuic)
	Udp_relay_mode        string         `yaml:"udp-relay-mode,omitempty"`              // UDP 转发模式 (Tuic)
	Disable_sni           bool           `yaml:"disable-sni,omitempty"`                 // 禁用 SNI (Tuic)
	Dialer_proxy          string         `yaml:"dialer-proxy,omitempty"`                // 前置代理
	// 通用 BasicOption 字段（mihomo 所有出站共用的连接层选项）
	Mptcp          bool   `yaml:"mptcp,omitempty"`          // 是否启用 MPTCP
	Interface_name string `yaml:"interface-name,omitempty"` // 绑定的本地网卡名称
	Routing_mark   int    `yaml:"routing-mark,omitempty"`   // Linux fwmark 路由标记
	Ip_version     string `yaml:"ip-version,omitempty"`     // IP 版本偏好 (dual/ipv4/ipv6/ipv4-prefer/ipv6-prefer)
	// SS 插件字段
	Plugin      string         `yaml:"plugin,omitempty"`      // SS 插件名称
	Plugin_opts map[string]any `yaml:"plugin-opts,omitempty"` // SS 插件选项
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
	// OpenVPN 特有字段
	Proto                string            `yaml:"proto,omitempty"`                 // OpenVPN 传输协议 (tcp/udp)
	Dev                  string            `yaml:"dev,omitempty"`                   // OpenVPN 设备类型（当前 Mihomo 支持 tun）
	Data_ciphers         []string          `yaml:"data-ciphers,omitempty"`          // OpenVPN 数据通道加密算法列表
	Data_cipher_fallback string            `yaml:"data-ciphers-fallback,omitempty"` // OpenVPN 数据通道回退加密算法
	Comp_lzo             string            `yaml:"comp-lzo,omitempty"`              // OpenVPN LZO 压缩模式
	Ca                   string            `yaml:"ca,omitempty"`                    // OpenVPN CA 证书
	Cert                 string            `yaml:"cert,omitempty"`                  // OpenVPN 客户端证书
	Key                  string            `yaml:"key,omitempty"`                   // OpenVPN 客户端私钥
	Tls_auth             string            `yaml:"tls-auth,omitempty"`              // OpenVPN tls-auth 静态密钥
	Key_direction        string            `yaml:"key-direction,omitempty"`         // OpenVPN 静态密钥方向
	Tls_crypt            string            `yaml:"tls-crypt,omitempty"`             // OpenVPN tls-crypt 静态密钥
	Tls_crypt_v2         string            `yaml:"tls-crypt-v2,omitempty"`          // OpenVPN tls-crypt-v2 客户端密钥
	Peer_info            map[string]string `yaml:"peer-info,omitempty"`             // OpenVPN Peer Info
	Ping                 int               `yaml:"ping,omitempty"`                  // OpenVPN keepalive ping 间隔
	Ping_restart         int               `yaml:"ping-restart,omitempty"`          // OpenVPN ping 超时重连时间
	Tran_window          *int              `yaml:"tran-window,omitempty"`           // OpenVPN 数据通道密钥转换窗口
	Handshake_timeout    int               `yaml:"handshake-timeout,omitempty"`     // OpenVPN 握手超时时间
	Ip_stack             map[string]any    `yaml:"ip-stack,omitempty"`              // OpenVPN 用户态 IP 栈配置
	Remote_dns_resolve   bool              `yaml:"remote-dns-resolve,omitempty"`    // 通过远端解析 DNS
	Dns                  []string          `yaml:"dns,omitempty"`                   // OpenVPN DNS 服务器
	// VLESS 特有字段
	Encryption      string         `yaml:"encryption,omitempty"`
	Packet_encoding string         `yaml:"packet-encoding,omitempty"` // VLESS packet-encoding (xudp/packetaddr)
	H2_opts         map[string]any `yaml:"h2-opts,omitempty"`         // HTTP/2 传输层选项
	Http_opts       map[string]any `yaml:"http-opts,omitempty"`       // HTTP 传输层选项
	XHTTP_opts      map[string]any `yaml:"xhttp-opts,omitempty"`
	ECH_opts        map[string]any `yaml:"ech-opts,omitempty"`
	Fragment_opts   map[string]any `yaml:"fragment,omitempty"` // TLS ClientHello 分片（对应 Xray finalmask 的 tcp fragment）
	ClashExtra      yaml.Node      `yaml:"-"`                  // 导入时未被模型识别的 Clash 顶层属性
}

func sanitizeCertificateFingerprint(fingerprint string) string {
	fingerprint = strings.ReplaceAll(fingerprint, ":", "")
	fingerprint = strings.ReplaceAll(fingerprint, "-", "")
	if len(fingerprint) != 64 {
		return ""
	}
	for _, c := range fingerprint {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			continue
		}
		return ""
	}
	fingerprint = strings.ToLower(fingerprint)
	return fingerprint
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
	ClashExtra      string
}

// 删除opts中的空值
func DeleteOpts(opts map[string]any) {
	for k, v := range opts {
		switch v := v.(type) {
		case string:
			if v == "" {
				delete(opts, k)
			}
		case map[string]any:
			DeleteOpts(v)
			if len(v) == 0 {
				delete(opts, k)
			}
		}
	}
}
func convertToInt(value any) (int, error) {
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

func isTruthyConfigValue(value any) bool {
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
func shouldPreserveProxyGroup(proxyGroup map[string]any) bool {
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
func convertSSPluginOpts(plugin SsPlugin) map[string]any {
	if plugin.Name == "" {
		return nil
	}

	opts := make(map[string]any)

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
	proxy, err := proxyCapable.ToProxy(link, config)
	if err != nil {
		return Proxy{}, err
	}
	if err := ApplyClashExtra(&proxy, link.ClashExtra); err != nil {
		return Proxy{}, err
	}
	return proxy, nil
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
			utils.Error("link conversion failed: %s", err.Error())
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

// clashAllProxiesPlaceholder 是代理组内展开全部节点的占位符，与 subconverter 行为一致。
const clashAllProxiesPlaceholder = "__ALL_PROXIES__"

// DecodeClash 用于解析 Clash 配置文件并合并新节点
// proxys: 新增的节点列表
// yamlfile: 模板文件路径或 URL
// customGroups: 自定义代理组列表（可选，由链式代理规则生成）
//
// 模板优先按原文拼接：只在文本上重写顶层 proxies、代理组的 proxies 与追加自定义组，
// 其余字节（注释、空行、缩进、键顺序、流式写法）与模板完全一致。
// 结构过于复杂（流式代理组、锚点、缺少 proxies 键等）时退回节点树重编码。
func DecodeClash(proxys []Proxy, yamlfile string, customGroups ...[]CustomProxyGroup) ([]byte, error) {
	data, err := readClashTemplate(yamlfile)
	if err != nil {
		return nil, err
	}

	// 解析 YAML 文件
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		utils.Error("error: %v", err)
		return nil, err
	}
	root, err := clashTemplateRoot(&document)
	if err != nil {
		utils.Error("error: %v", err)
		return nil, err
	}

	// 添加新代理：节点写进模板的 proxies 序列，模板其余内容保持原样
	proxiesNodes := make([]*yaml.Node, 0, len(proxys))
	proxiesNameList := make([]string, 0, len(proxys))
	for _, proxy := range proxys {
		node := &yaml.Node{}
		if err := node.Encode(proxy); err != nil {
			utils.Error("error: %v", err)
			return nil, err
		}
		proxiesNodes = append(proxiesNodes, node)
		proxiesNameList = append(proxiesNameList, proxy.Name)
	}

	// 自定义代理组（链式代理规则生成）统一追加在模板组之后
	customGroupNodes := make([]*yaml.Node, 0)
	if len(customGroups) > 0 {
		for _, customGroup := range customGroups[0] {
			node := &yaml.Node{}
			if err := node.Encode(buildCustomProxyGroup(customGroup)); err != nil {
				utils.Error("error: %v", err)
				return nil, err
			}
			customGroupNodes = append(customGroupNodes, node)
		}
	}

	// 优先按模板原文拼接，未改动的字节原样保留
	if merged, ok := mergeClashTemplateText(data, root, proxiesNodes, proxiesNameList, customGroupNodes); ok {
		return merged, nil
	}
	return encodeClashTemplateNodes(&document, root, proxiesNodes, proxiesNameList, customGroupNodes)
}

// readClashTemplate 读取模板内容，本地文件优先走模板缓存。
func readClashTemplate(yamlfile string) ([]byte, error) {
	if strings.Contains(yamlfile, "://") {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, yamlfile, nil)
		if err != nil {
			utils.Error("http.Get error: %v", err)
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			utils.Error("http.Get error: %v", err)
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			utils.Error("error: %v", err)
			return nil, err
		}
		return data, nil
	}
	// 优先从缓存读取模板内容（本地文件使用缓存）
	filename := filepath.Base(yamlfile)
	if cached, ok := cache.GetTemplateContent(filename); ok {
		return []byte(cached), nil
	}
	data, err := os.ReadFile(yamlfile)
	if err != nil {
		utils.Error("error: %v", err)
		return nil, err
	}
	cache.SetTemplateContent(filename, string(data))
	return data, nil
}

// encodeClashTemplateNodes 是文本拼接失败时的后备路径：在 yaml.Node 树上原地修改后整体序列化。
// 注释和键顺序仍然保留，但空行会被 YAML 编码器归一化。
func encodeClashTemplateNodes(document, root *yaml.Node, proxiesNodes []*yaml.Node, proxiesNameList []string, customGroupNodes []*yaml.Node) ([]byte, error) {
	proxiesNode := clashTemplateSequence(root, "proxies")
	proxiesNode.Content = append(proxiesNode.Content, proxiesNodes...)
	if len(proxiesNodes) > 0 {
		// proxies: [] / proxies: ~ 追加节点后展开为块格式，避免输出成一行超长 YAML
		proxiesNode.Style &^= yaml.FlowStyle
	}
	// 往ProxyGroup中插入代理列表
	groupsNode := clashTemplateSequence(root, "proxy-groups")
	templateGroupCount := len(groupsNode.Content)

	// 插入自定义代理组（在模板组之后），模板组本身保持原样
	if len(customGroupNodes) > 0 {
		groupsNode.Content = append(groupsNode.Content, customGroupNodes...)
		groupsNode.Style &^= yaml.FlowStyle
	}

	for index := 0; index < templateGroupCount; index++ {
		mergeClashProxyGroup(groupsNode.Content[index], proxiesNameList)
	}

	// 将修改后的内容写回：保留模板里的注释、键顺序与写法
	astralSubstitutions := substituteClashAstralCharacters(document)
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(document); err != nil {
		utils.Error("error: %v", err)
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		utils.Error("error: %v", err)
		return nil, err
	}
	return restoreClashAstralCharacters(output.Bytes(), astralSubstitutions), nil
}

// clashTemplateRoot 返回模板的顶层映射节点，空模板按空映射处理。
func clashTemplateRoot(document *yaml.Node) (*yaml.Node, error) {
	if len(document.Content) == 0 {
		mapping := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		document.Kind = yaml.DocumentNode
		document.Content = []*yaml.Node{mapping}
		return mapping, nil
	}
	mapping := yamlMappingNode(document)
	if mapping == nil {
		return nil, fmt.Errorf("clash template must be a YAML mapping")
	}
	return mapping, nil
}

// clashTemplateSequence 返回顶层键对应的序列节点，缺失或写成空值时原地创建。
func clashTemplateSequence(root *yaml.Node, key string) *yaml.Node {
	value := yamlMappingValue(root, key)
	if value == nil {
		value = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		yamlMappingSet(root, key, value)
		return value
	}
	if resolved := resolveYAMLAlias(value); resolved != nil && resolved.Kind == yaml.SequenceNode {
		return resolved
	}
	if value.Kind != yaml.AliasNode {
		// 就地改造成空序列，保留节点上的注释
		value.Kind = yaml.SequenceNode
		value.Tag = "!!seq"
		value.Value = ""
		value.Style = 0
		value.Content = nil
		return value
	}
	empty := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	yamlMappingSet(root, key, empty)
	return empty
}

// mergeClashProxyGroup 把节点名合并进单个代理组，组本身的字段顺序和注释保持不变。
func mergeClashProxyGroup(group *yaml.Node, proxiesNameList []string) {
	group = resolveYAMLAlias(group)
	if group == nil || group.Kind != yaml.MappingNode {
		return
	}
	if typeValue := yamlMappingValue(group, "type"); typeValue != nil {
		var groupType string
		if err := typeValue.Decode(&groupType); err == nil && groupType == "relay" {
			return
		}
	}
	var fields map[string]any
	if err := group.Decode(&fields); err != nil {
		utils.Error("failed to parse proxy groups: %v", err)
		return
	}
	// include-all、use/filter 等自动匹配组保持模板原意，不补全成员列表，
	// 但显式写下的 __ALL_PROXIES__ 仍要就地展开，否则占位符会原样发给客户端。
	preserve := shouldPreserveProxyGroup(fields)

	proxiesNode := resolveYAMLAlias(yamlMappingValue(group, "proxies"))
	switch {
	case proxiesNode == nil:
		if preserve {
			return
		}
		proxiesNode = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		yamlMappingSet(group, "proxies", proxiesNode)
	case proxiesNode.Kind == yaml.ScalarNode && proxiesNode.Tag == "!!null":
		if preserve {
			return
		}
		// 模板写成 proxies: 后面留空时按空列表处理，与旧版行为一致
		proxiesNode.Kind = yaml.SequenceNode
		proxiesNode.Tag = "!!seq"
		proxiesNode.Value = ""
		proxiesNode.Style = 0
		proxiesNode.Content = nil
	case proxiesNode.Kind != yaml.SequenceNode:
		return
	}

	// 占位符 __ALL_PROXIES__ 展开为全部节点，保留占位符前后的组引用
	for index, item := range proxiesNode.Content {
		resolved := resolveYAMLAlias(item)
		if resolved == nil || resolved.Kind != yaml.ScalarNode || resolved.Value != clashAllProxiesPlaceholder {
			continue
		}
		merged := make([]*yaml.Node, 0, len(proxiesNode.Content)-1+len(proxiesNameList))
		merged = append(merged, proxiesNode.Content[:index]...)
		merged = append(merged, yamlStringNodes(proxiesNameList)...)
		merged = append(merged, proxiesNode.Content[index+1:]...)
		proxiesNode.Content = merged
		proxiesNode.Style &^= yaml.FlowStyle
		return
	}

	if preserve {
		return
	}
	// 模板里已列出组引用的保持不变，只有空列表才追加所有节点
	if len(proxiesNode.Content) > 0 {
		return
	}
	names := proxiesNameList
	if len(names) == 0 {
		names = []string{"DIRECT"}
	}
	proxiesNode.Content = append(proxiesNode.Content, yamlStringNodes(names)...)
	proxiesNode.Style &^= yaml.FlowStyle
}

// buildCustomProxyGroup 把链式代理规则生成的自定义代理组转换成 Clash 组配置。
func buildCustomProxyGroup(customGroup CustomProxyGroup) map[string]any {
	groupMap := map[string]any{
		"name":    customGroup.Name,
		"type":    customGroup.Type,
		"proxies": customGroup.Proxies,
	}
	interval := customGroup.Interval
	if interval <= 0 {
		interval = 300 // 默认 300 秒
	}
	switch customGroup.Type {
	case "url-test", "fallback":
		// url-test 和 fallback 类型需要 url、interval、tolerance
		groupMap["url"] = defaultString(customGroup.URL, "http://www.gstatic.com/generate_204")
		groupMap["interval"] = interval
		tolerance := customGroup.Tolerance
		if tolerance <= 0 {
			tolerance = 50 // 默认 50 毫秒
		}
		groupMap["tolerance"] = tolerance
	case "load-balance":
		// load-balance 类型需要 url、interval、strategy
		groupMap["url"] = defaultString(customGroup.URL, "http://www.gstatic.com/generate_204")
		groupMap["interval"] = interval
		groupMap["strategy"] = defaultString(customGroup.Strategy, "consistent-hashing") // 默认一致性哈希
	}
	return groupMap
}

// clashAstralPlaceholderStart 是替换 BMP 之外字符时使用的私有区起点。
// yaml.v3 会把 U+10000 及以上的字符（emoji、国旗等）转义成 \UXXXXXXXX，
// 先用 BMP 内的私有区占位符替换，编码后再还原，保证输出与模板保持一致。
const clashAstralPlaceholderStart = 0xE000

// clashAstralSubstitution 记录一个占位符与原始字符的对应关系。
type clashAstralSubstitution struct {
	placeholder string
	original    string
}

// substituteClashAstralCharacters 把文档中 BMP 之外的字符替换成占位符，返回还原所需的映射。
// 私有区不足以分配占位符时返回 nil，此时保持 yaml.v3 的 \U 转义输出。
func substituteClashAstralCharacters(document *yaml.Node) []clashAstralSubstitution {
	used := make(map[rune]struct{})
	astral := make(map[rune]struct{})
	walkClashScalarNodes(document, func(node *yaml.Node) {
		for _, char := range node.Value {
			used[char] = struct{}{}
			if char > 0xFFFF {
				astral[char] = struct{}{}
			}
		}
	})
	if len(astral) == 0 {
		return nil
	}

	originals := make([]rune, 0, len(astral))
	for char := range astral {
		originals = append(originals, char)
	}
	sort.Slice(originals, func(left, right int) bool { return originals[left] < originals[right] })

	placeholders := make(map[rune]rune, len(originals))
	substitutions := make([]clashAstralSubstitution, 0, len(originals))
	next := rune(clashAstralPlaceholderStart)
	for _, original := range originals {
		for next <= 0xF8FF {
			if _, taken := used[next]; !taken {
				break
			}
			next++
		}
		if next > 0xF8FF {
			return nil
		}
		used[next] = struct{}{}
		placeholders[original] = next
		substitutions = append(substitutions, clashAstralSubstitution{
			placeholder: string(next),
			original:    string(original),
		})
		next++
	}

	walkClashScalarNodes(document, func(node *yaml.Node) {
		if !strings.ContainsFunc(node.Value, func(char rune) bool { return char > 0xFFFF }) {
			return
		}
		node.Value = strings.Map(func(char rune) rune {
			if placeholder, ok := placeholders[char]; ok {
				return placeholder
			}
			return char
		}, node.Value)
	})
	return substitutions
}

// restoreClashAstralCharacters 把编码结果里的占位符还原成原始字符。
func restoreClashAstralCharacters(data []byte, substitutions []clashAstralSubstitution) []byte {
	for _, substitution := range substitutions {
		data = bytes.ReplaceAll(data, []byte(substitution.placeholder), []byte(substitution.original))
	}
	return data
}

// walkClashScalarNodes 遍历文档中的所有标量节点。
func walkClashScalarNodes(node *yaml.Node, visit func(*yaml.Node)) {
	if node == nil {
		return
	}
	if node.Kind == yaml.ScalarNode {
		visit(node)
	}
	for _, child := range node.Content {
		walkClashScalarNodes(child, visit)
	}
}

// defaultString 在值为空时回退到默认值。
func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// yamlMappingValue 返回映射节点中键对应的值节点。
func yamlMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1]
		}
	}
	return nil
}

// yamlMappingSet 写入键值对：键已存在时只替换值节点，保持模板原有的键顺序。
func yamlMappingSet(mapping *yaml.Node, key string, value *yaml.Node) {
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			mapping.Content[index+1] = value
			return
		}
	}
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
}

// resolveYAMLAlias 跳过锚点别名，返回真正承载数据的节点。
func resolveYAMLAlias(node *yaml.Node) *yaml.Node {
	for node != nil && node.Kind == yaml.AliasNode {
		node = node.Alias
	}
	return node
}

// yamlStringNodes 把字符串列表转换成 YAML 标量节点。
func yamlStringNodes(values []string) []*yaml.Node {
	nodes := make([]*yaml.Node, 0, len(values))
	for _, value := range values {
		nodes = append(nodes, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value})
	}
	return nodes
}
