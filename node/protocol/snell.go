package protocol

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

func init() {
	base := newProtocolSpec("snell", []string{"snell://"}, "Snell", "#00897b", "S", Snell{}, "Name", DecodeSnellURL, EncodeSnellURL, func(s Snell) LinkIdentity {
		return buildIdentity("snell", s.Name, s.Server, utils.GetPortString(s.Port))
	},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic"},
		FieldMeta{Name: "Server", Label: "服务器地址", Type: "string", Group: "basic"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic"},
		FieldMeta{Name: "Psk", Label: "PSK", Type: "string", Group: "auth", Secret: true},
		FieldMeta{Name: "Version", Label: "版本", Type: "int", Group: "basic", Options: []string{"1", "2", "3"}},
		FieldMeta{Name: "UDP", Label: "UDP", Type: "bool", Group: "transport", Advanced: true},
		FieldMeta{Name: "ObfsMode", Label: "混淆模式", Type: "string", Group: "transport", Advanced: true, Options: []string{"", "http", "tls"}},
		FieldMeta{Name: "ObfsHost", Label: "混淆 Host", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Tfo", Label: "TCP Fast Open", Type: "bool", Group: "transport", Advanced: true},
		FieldMeta{Name: "Mptcp", Label: "MPTCP", Type: "bool", Group: "transport", Advanced: true},
		FieldMeta{Name: "InterfaceName", Label: "绑定网卡", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "RoutingMark", Label: "路由标记", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "IPVersion", Label: "IP 版本", Type: "string", Group: "transport", Advanced: true, Options: []string{"", "dual", "ipv4", "ipv6", "ipv4-prefer", "ipv6-prefer"}},
	).WithClientSupport(ClientClash, ClientMihomo, ClientSurge)
	MustRegisterProtocol(newProxySurgeProtocolSpec(base, buildSnellProxy, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "snell")
	}, ConvertProxyToSnell, EncodeSnellURL, buildSnellSurgeLine))
}

// Snell 保存 SublinkPro 内部可编辑的 Snell URL 结构。
// Snell 官方没有定义通用的分享链接方案，mihomo 以 Clash YAML 字段描述 Snell；
// SublinkPro 仅在原始编辑与 Clash/mihomo、Surge 导入往返时使用内部 snell:// 形式。
type Snell struct {
	Name     string
	Server   string
	Port     any
	Psk      string
	Version  int
	UDP      bool
	ObfsMode string
	ObfsHost string
	// 通用 BasicOption 连接层选项，保证 Clash/mihomo 往返不丢字段
	Tfo           bool
	Mptcp         bool
	InterfaceName string
	RoutingMark   int
	IPVersion     string
}

// DecodeSnellURL 解析内部 snell:// 链接，补齐默认版本并提取混淆参数。
func DecodeSnellURL(s string) (Snell, error) {
	u, err := url.Parse(s)
	if err != nil {
		return Snell{}, fmt.Errorf("url parse error: %v", err)
	}
	if strings.ToLower(u.Scheme) != "snell" {
		return Snell{}, fmt.Errorf("非snell协议: %s", s)
	}

	server := u.Hostname()
	if server == "" {
		return Snell{}, fmt.Errorf("缺少服务器地址: %s", s)
	}

	rawPort := u.Port()
	port := 0
	if rawPort != "" {
		port, err = strconv.Atoi(rawPort)
		if err != nil {
			return Snell{}, fmt.Errorf("Snell port conversion failed: %w", err)
		}
	}

	query := u.Query()
	psk := query.Get("psk")
	if psk == "" {
		// 兼容将 PSK 放在 userinfo 段的写法：snell://psk@server:port
		psk = u.User.Username()
	}

	version := 0
	if rawVersion := query.Get("version"); rawVersion != "" {
		version, _ = strconv.Atoi(rawVersion)
	}

	udp, _ := strconv.ParseBool(query.Get("udp"))

	obfsMode := query.Get("obfs")
	obfsHost := query.Get("obfs-host")

	tfo, _ := strconv.ParseBool(query.Get("tfo"))
	mptcp, _ := strconv.ParseBool(query.Get("mptcp"))
	interfaceName := query.Get("interface-name")
	routingMark := 0
	if rawMark := query.Get("routing-mark"); rawMark != "" {
		routingMark, _ = strconv.Atoi(rawMark)
	}
	ipVersion := query.Get("ip-version")

	name := u.Fragment
	if name == "" {
		if port != 0 {
			name = fmt.Sprintf("%s:%d", server, port)
		} else {
			name = server
		}
	}

	return Snell{
		Name:     name,
		Server:   server,
		Port:     port,
		Psk:      psk,
		Version:  version,
		UDP:      udp,
		ObfsMode: obfsMode,
		ObfsHost: obfsHost,

		Tfo:           tfo,
		Mptcp:         mptcp,
		InterfaceName: interfaceName,
		RoutingMark:   routingMark,
		IPVersion:     ipVersion,
	}, nil
}

// EncodeSnellURL 将 Snell 结构编码为内部 snell:// 链接，清理空值与零值字段。
func EncodeSnellURL(s Snell) string {
	server := utils.WrapIPv6Host(strings.TrimSpace(s.Server))
	host := server
	if port := strings.TrimSpace(utils.GetPortString(s.Port)); port != "" && port != "0" && port != "<nil>" {
		host = formatURLHostPort(server, port)
	}

	u := url.URL{
		Scheme: "snell",
		Host:   host,
	}

	q := u.Query()
	if s.Psk != "" {
		q.Set("psk", s.Psk)
	}
	if s.Version > 0 {
		q.Set("version", strconv.Itoa(s.Version))
	}
	if s.UDP {
		q.Set("udp", "1")
	}
	if s.ObfsMode != "" {
		q.Set("obfs", s.ObfsMode)
	}
	if s.ObfsHost != "" {
		q.Set("obfs-host", s.ObfsHost)
	}
	if s.Tfo {
		q.Set("tfo", "1")
	}
	if s.Mptcp {
		q.Set("mptcp", "1")
	}
	if s.InterfaceName != "" {
		q.Set("interface-name", s.InterfaceName)
	}
	if s.RoutingMark > 0 {
		q.Set("routing-mark", strconv.Itoa(s.RoutingMark))
	}
	if s.IPVersion != "" {
		q.Set("ip-version", s.IPVersion)
	}
	u.RawQuery = q.Encode()

	if s.Name == "" {
		u.Fragment = fmt.Sprintf("%s:%s", strings.Trim(server, "[]"), utils.GetPortString(s.Port))
	} else {
		u.Fragment = s.Name
	}
	return u.String()
}

// ConvertProxyToSnell 将 Proxy 结构体转换为 Snell 结构体。
// 用于从 Clash 格式的代理配置生成内部 snell:// 链接。
func ConvertProxyToSnell(proxy Proxy) Snell {
	snell := Snell{
		Name:          proxy.Name,
		Server:        proxy.Server,
		Port:          int(proxy.Port),
		Psk:           proxy.Psk,
		Version:       proxy.Version,
		UDP:           proxy.Udp,
		Tfo:           proxy.Tfo,
		Mptcp:         proxy.Mptcp,
		InterfaceName: proxy.Interface_name,
		RoutingMark:   proxy.Routing_mark,
		IPVersion:     proxy.Ip_version,
	}
	if len(proxy.Obfs_opts) > 0 {
		if mode, ok := proxy.Obfs_opts["mode"].(string); ok {
			snell.ObfsMode = mode
		}
		if host, ok := proxy.Obfs_opts["host"].(string); ok {
			snell.ObfsHost = host
		}
	}
	return snell
}

// buildSnellProxy 将内部 snell:// 链接转换为 Clash Proxy。
// UDP 会与输出配置合并，混淆参数按 mihomo 的 obfs-opts 结构输出。
func buildSnellProxy(link Urls, config OutputConfig) (Proxy, error) {
	snell, err := DecodeSnellURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	proxy := Proxy{
		Name:           snell.Name,
		Type:           "snell",
		Server:         snell.Server,
		Port:           FlexPort(utils.GetPortInt(snell.Port)),
		Psk:            snell.Psk,
		Version:        snell.Version,
		Udp:            config.Udp || snell.UDP,
		Tfo:            snell.Tfo,
		Mptcp:          snell.Mptcp,
		Interface_name: snell.InterfaceName,
		Routing_mark:   snell.RoutingMark,
		Ip_version:     snell.IPVersion,
		Dialer_proxy:   link.DialerProxyName,
	}
	if snell.ObfsMode != "" {
		obfsOpts := map[string]any{"mode": snell.ObfsMode}
		if snell.ObfsHost != "" {
			obfsOpts["host"] = snell.ObfsHost
		}
		proxy.Obfs_opts = obfsOpts
	}
	return proxy, nil
}

// buildSnellSurgeLine 将内部 snell:// 链接转换为 Surge 节点行。
// Surge 使用 psk / version / obfs / obfs-host 字段描述 Snell。
func buildSnellSurgeLine(link string, config OutputConfig) (string, string, error) {
	snell, err := DecodeSnellURL(link)
	if err != nil {
		return "", "", err
	}
	server := replaceSurgeHost(snell.Server, config)
	line := fmt.Sprintf("%s = snell, %s, %d, psk=%s", snell.Name, server, utils.GetPortInt(snell.Port), snell.Psk)
	if snell.Version > 0 {
		line = fmt.Sprintf("%s, version=%d", line, snell.Version)
	}
	if snell.ObfsMode != "" {
		line = fmt.Sprintf("%s, obfs=%s", line, snell.ObfsMode)
		if snell.ObfsHost != "" {
			line = fmt.Sprintf("%s, obfs-host=%s", line, snell.ObfsHost)
		}
	}
	if config.Udp || snell.UDP {
		line = fmt.Sprintf("%s, udp-relay=%t", line, true)
	}
	if snell.Tfo {
		line = fmt.Sprintf("%s, tfo=true", line)
	}
	return line, snell.Name, nil
}
