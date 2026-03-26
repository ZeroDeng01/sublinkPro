package protocol

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

func init() {
	base := newProtocolSpec("hysteria", []string{"hysteria://", "hy://"}, "Hysteria", "#f9a825", "H", HY{}, "Name", DecodeHYURL, EncodeHYURL, func(h HY) LinkIdentity {
		return buildIdentity("hysteria", h.Name, h.Host, utils.GetPortString(h.Port))
	},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic"},
		FieldMeta{Name: "Host", Label: "服务器地址", Type: "string", Group: "basic"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic"},
		FieldMeta{Name: "Auth", Label: "认证", Type: "string", Group: "auth", Secret: true, Advanced: true},
		FieldMeta{Name: "Protocol", Label: "协议", Type: "string", Group: "transport", Options: []string{"udp", "wechat-video", "faketcp"}},
		FieldMeta{Name: "UpMbps", Label: "上行 Mbps", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "DownMbps", Label: "下行 Mbps", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "Peer", Label: "Peer", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "ALPN", Label: "ALPN", Type: "string", Group: "tls", Advanced: true, Multiline: true},
		FieldMeta{Name: "Insecure", Label: "跳过证书校验", Type: "int", Group: "tls", Advanced: true, Options: []string{"0", "1"}},
	)
	MustRegisterProtocol(newProxyProtocolSpec(base, buildHYProxy, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "hysteria")
	}, ConvertProxyToHy, EncodeHYURL))
}

type HY struct {
	Host     string
	Port     interface{}
	Insecure int
	Peer     string
	Auth     string
	UpMbps   int
	DownMbps int
	Protocol string
	ALPN     []string
	Name     string
}

// EncodeHYURL 将 Hysteria v1 结构编码为 hysteria:// 链接。
// 导出时会省略空值和零值字段，并在名称缺失时回退为 host:port。
func EncodeHYURL(hy HY) string {
	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if hy.Name == "" {
		hy.Name = fmt.Sprintf("%s:%s", hy.Host, utils.GetPortString(hy.Port))
	}
	u := url.URL{
		Scheme:   "hysteria",
		Host:     fmt.Sprintf("%s:%s", hy.Host, utils.GetPortString(hy.Port)),
		Fragment: hy.Name,
	}
	q := u.Query()
	// protocol 参数，Hysteria 基于 UDP/QUIC
	q.Set("protocol", "udp")
	q.Set("insecure", strconv.Itoa(hy.Insecure))
	q.Set("peer", hy.Peer)
	q.Set("auth", hy.Auth)
	q.Set("upmbps", strconv.Itoa(hy.UpMbps))
	q.Set("downmbps", strconv.Itoa(hy.DownMbps))
	q.Set("protocol", hy.Protocol)
	// alpn 参数支持
	if len(hy.ALPN) > 0 {
		q.Set("alpn", strings.Join(hy.ALPN, ","))
	}
	// 检查query是否有空值，有的话删除
	for k, v := range q {
		if v[0] == "" || v[0] == "0" {
			delete(q, k)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// DecodeHYURL 解析 hy:// 与 hysteria:// 两种别名链接，并在缺省端口时按当前约定回退到 443。
func DecodeHYURL(s string) (HY, error) {
	u, err := url.Parse(s)
	if err != nil {
		return HY{}, fmt.Errorf("失败的URL: %s", s)
	}
	if u.Scheme != "hy" && u.Scheme != "hysteria" {
		return HY{}, fmt.Errorf("非hy协议: %s", s)
	}
	server := u.Hostname()
	rawPort := u.Port()
	if rawPort == "" {
		rawPort = "443"
	}
	port, _ := strconv.Atoi(rawPort)
	insecure, _ := strconv.Atoi(u.Query().Get("insecure"))
	peer := u.Query().Get("peer")
	auth := u.Query().Get("auth")
	upMbps, _ := strconv.Atoi(u.Query().Get("upmbps"))
	downMbps, _ := strconv.Atoi(u.Query().Get("downmbps"))
	alpns := u.Query().Get("alpn")
	alpn := strings.Split(alpns, ",")
	protocol := u.Query().Get("protocol")
	if alpns == "" {
		alpn = nil
	}
	// 如果没有设置 Name，则使用 Fragment 作为 Name
	name := u.Fragment
	if name == "" {
		name = server + ":" + u.Port()
	}
	if utils.CheckEnvironment() {
		fmt.Println("server:", server)
		fmt.Println("port:", port)
		fmt.Println("insecure:", insecure)
		fmt.Println("peer:", peer)
		fmt.Println("auth:", auth)
		fmt.Println("upMbps:", upMbps)
		fmt.Println("downMbps:", downMbps)
		fmt.Println("alpn:", alpn)
		fmt.Println("protocol:", protocol)
		fmt.Println("name:", name)
	}
	return HY{
		Host:     server,
		Port:     port,
		Insecure: insecure,
		Peer:     peer,
		Auth:     auth,
		UpMbps:   upMbps,
		DownMbps: downMbps,
		ALPN:     alpn,
		Protocol: protocol,
		Name:     name,
	}, nil
}

// ConvertProxyToHy 将 Proxy 结构体转换为 HY 结构体
// 用于从 Clash 格式的代理配置生成 Hysteria 链接
func ConvertProxyToHy(proxy Proxy) HY {
	hy := HY{
		Host:     proxy.Server,
		Port:     int(proxy.Port),
		Auth:     proxy.Auth_str,
		UpMbps:   proxy.Up,
		DownMbps: proxy.Down,
		ALPN:     proxy.Alpn,
		Peer:     proxy.Peer,
		Protocol: proxy.Protocol,
		Name:     proxy.Name,
	}

	// 处理跳过证书验证
	if proxy.Skip_cert_verify {
		hy.Insecure = 1
	}

	return hy
}

// buildHYProxy 将 Hysteria 链接转换为 Clash Proxy。
// 当前实现固定按 UDP 节点导出，并将输出阶段的证书校验覆盖与原链接设置合并处理。
func buildHYProxy(link Urls, config OutputConfig) (Proxy, error) {
	hy, err := DecodeHYURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	if hy.Name == "" {
		hy.Name = fmt.Sprintf("%s:%s", hy.Host, utils.GetPortString(hy.Port))
	}
	skipCert := config.Cert || hy.Insecure == 1
	return Proxy{Name: hy.Name, Type: "hysteria", Server: hy.Host, Port: FlexPort(utils.GetPortInt(hy.Port)), Auth_str: hy.Auth, Up: hy.UpMbps, Down: hy.DownMbps, Up_Speed: hy.UpMbps, Down_Speed: hy.DownMbps, Alpn: hy.ALPN, Peer: hy.Peer, Protocol: hy.Protocol, Udp: true, Skip_cert_verify: skipCert, Dialer_proxy: link.DialerProxyName}, nil
}
