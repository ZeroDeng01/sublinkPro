package protocol

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

func init() {
	base := newProtocolSpec("trojan", []string{"trojan://"}, "Trojan", "#d32f2f", "T", Trojan{}, "Name", DecodeTrojanURL, EncodeTrojanURL, func(t Trojan) LinkIdentity {
		return buildIdentity("trojan", t.Name, t.Hostname, utils.GetPortString(t.Port))
	},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic"},
		FieldMeta{Name: "Hostname", Label: "服务器地址", Type: "string", Group: "basic"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic"},
		FieldMeta{Name: "Password", Label: "密码", Type: "string", Group: "auth", Secret: true},
		FieldMeta{Name: "Query.Type", Label: "传输层", Type: "string", Group: "transport", Options: []string{"tcp", "ws", "grpc"}},
		FieldMeta{Name: "Query.Path", Label: "路径", Type: "string", Group: "transport", Placeholder: "/ws", Advanced: true},
		FieldMeta{Name: "Query.Host", Label: "Host", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.Flow", Label: "Flow", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.Security", Label: "安全类型", Type: "string", Group: "tls", Options: []string{"tls", "reality"}, Advanced: true},
		FieldMeta{Name: "Query.Sni", Label: "SNI", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Query.Peer", Label: "Peer", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Query.Alpn", Label: "ALPN", Type: "string", Group: "tls", Advanced: true, Multiline: true},
		FieldMeta{Name: "Query.Fp", Label: "指纹", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Query.AllowInsecure", Label: "跳过证书校验", Type: "int", Group: "tls", Advanced: true, Options: []string{"0", "1"}},
		FieldMeta{Name: "Query.Pbk", Label: "Public Key", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Query.Sid", Label: "Short ID", Type: "string", Group: "tls", Advanced: true},
	)
	MustRegisterProtocol(newProxySurgeProtocolSpec(base, buildTrojanProxy, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "trojan")
	}, ConvertProxyToTrojan, EncodeTrojanURL, buildTrojanSurgeLine))
}

type Trojan struct {
	Password string      `json:"password"`
	Hostname string      `json:"hostname"`
	Port     interface{} `json:"port"`
	Query    TrojanQuery `json:"query,omitempty"`
	Name     string      `json:"name"`
	Type     string      `json:"type"`
}
type TrojanQuery struct {
	Peer          string   `json:"peer,omitempty"`
	Type          string   `json:"type,omitempty"`
	Path          string   `json:"path,omitempty"`
	Security      string   `json:"security,omitempty"`
	Fp            string   `json:"fp,omitempty"`
	AllowInsecure int      `json:"allowInsecure,omitempty"`
	Alpn          []string `json:"alpn,omitempty"`
	Sni           string   `json:"sni,omitempty"`
	Host          string   `json:"host,omitempty"`
	Flow          string   `json:"flow,omitempty"`
	// Reality 参数
	Pbk string `json:"pbk,omitempty"` // Reality public-key
	Sid string `json:"sid,omitempty"` // Reality short-id
}

// EncodeTrojanURL 将 Trojan 结构编码为 trojan:// 链接。
// 导出时会自动清理空查询参数，并在节点名缺失时回退为 host:port 形式。
func EncodeTrojanURL(t Trojan) string {
	/*
		trojan://password@hostname:port?peer=example.com&allowInsecure=0&sni=example.com
	*/
	u := url.URL{
		Scheme: "trojan",
		User:   url.User(t.Password),
		Host:   fmt.Sprintf("%s:%s", t.Hostname, utils.GetPortString(t.Port)),
	}
	q := u.Query()
	q.Set("peer", t.Query.Peer)
	q.Set("allowInsecure", fmt.Sprintf("%d", t.Query.AllowInsecure))
	q.Set("sni", t.Query.Sni)
	q.Set("type", t.Query.Type)
	q.Set("path", t.Query.Path)
	q.Set("security", t.Query.Security)
	q.Set("fp", t.Query.Fp)
	// alpn 参数支持
	if len(t.Query.Alpn) > 0 {
		q.Set("alpn", strings.Join(t.Query.Alpn, ","))
	}
	q.Set("host", t.Query.Host)
	q.Set("flow", t.Query.Flow)
	// Reality 参数支持
	q.Set("pbk", t.Query.Pbk)
	q.Set("sid", t.Query.Sid)
	// 检查query是否有空值，有的话删除
	for k, v := range q {
		if v[0] == "" {
			delete(q, k)
		}
	}
	// allowInsecure为0时也删除
	if t.Query.AllowInsecure == 0 {
		delete(q, "allowInsecure")
	}
	// 如果没有设置name,则使用hostname:port
	if t.Name == "" {
		t.Name = t.Hostname + ":" + utils.GetPortString(t.Port)
	}
	u.Fragment = t.Name
	u.RawQuery = q.Encode()
	return u.String()
}

// DecodeTrojanURL 解析 Trojan 链接，并按当前约定补默认端口和默认备注字段。
// 当前实现会提取后续导出链路已使用的查询参数，但不会还原全部扩展 TLS/Reality 字段。
func DecodeTrojanURL(s string) (Trojan, error) {
	/*
		trojan://password@hostname:port?peer=example.com&allowInsecure=0&sni=example.com
	*/
	u, err := url.Parse(s)
	if err != nil {
		return Trojan{}, fmt.Errorf("url格式化失败:%s", s)
	}
	if u.Scheme != "trojan" {
		return Trojan{}, fmt.Errorf("非trojan协议: %s", s)
	}
	password := u.User.Username()
	hostname := u.Hostname()
	rawPort := u.Port()
	if rawPort == "" {
		rawPort = "443"
	}
	port, _ := strconv.Atoi(rawPort)
	peer := u.Query().Get("peer")
	allowInsecure := u.Query().Get("allowInsecure")
	sni := u.Query().Get("sni")
	types := u.Query().Get("type")
	path := u.Query().Get("path")
	security := u.Query().Get("security")
	fp := u.Query().Get("fp")
	alpns := u.Query().Get("alpn")
	alpn := strings.Split(alpns, ",")
	if alpns == "" {
		alpn = nil
	}
	host := u.Query().Get("host")
	flow := u.Query().Get("flow")
	name := u.Fragment
	// 如果没有设置name,则使用hostname:port
	if name == "" {
		name = hostname + ":" + u.Port()
	}
	if utils.CheckEnvironment() {
		fmt.Println("password:", password)
		fmt.Println("password:", u.User.Username())
		fmt.Println("hostname:", hostname)
		fmt.Println("port:", port)
		fmt.Println("peer:", peer)
		fmt.Println("allowInsecure:", allowInsecure)
		fmt.Println("sni:", sni)
		fmt.Println("type:", types)
		fmt.Println("path:", path)
		fmt.Println("security:", security)
		fmt.Println("fp:", fp)
		fmt.Println("alpn:", alpn)
		fmt.Println("host:", host)
		fmt.Println("flow:", flow)
		fmt.Println("name:", name)
	}
	// 解析 allowInsecure 参数
	insecureVal := 0
	if allowInsecure == "1" || allowInsecure == "true" {
		insecureVal = 1
	}
	return Trojan{
		Password: password,
		Hostname: hostname,
		Port:     port,
		Query: TrojanQuery{
			Peer:          peer,
			Type:          types,
			Path:          path,
			Security:      security,
			Fp:            fp,
			AllowInsecure: insecureVal,
			Alpn:          alpn,
			Sni:           sni,
			Host:          host,
			Flow:          flow,
		},
		Name: name,
		Type: "trojan",
	}, nil
}

// ConvertProxyToTrojan 将 Proxy 结构体转换为 Trojan 结构体
// 用于从 Clash 格式的代理配置生成 Trojan 链接
func ConvertProxyToTrojan(proxy Proxy) Trojan {
	trojan := Trojan{
		Password: proxy.Password,
		Hostname: proxy.Server,
		Port:     int(proxy.Port),
		Name:     proxy.Name,
		Type:     "trojan",
		Query: TrojanQuery{
			Sni:  proxy.Sni,
			Type: proxy.Network,
			Fp:   proxy.Client_fingerprint,
			Flow: proxy.Flow,
			Alpn: proxy.Alpn,
			Peer: proxy.Peer,
		},
	}

	// 处理跳过证书验证
	if proxy.Skip_cert_verify {
		trojan.Query.AllowInsecure = 1
	}

	// 处理 ws_opts
	if len(proxy.Ws_opts) > 0 {
		if path, ok := proxy.Ws_opts["path"].(string); ok {
			trojan.Query.Path = path
		}
		if headers, ok := proxy.Ws_opts["headers"].(map[string]interface{}); ok {
			if host, ok := headers["Host"].(string); ok {
				trojan.Query.Host = host
			}
		}
	}

	// 处理 Reality 参数
	if len(proxy.Reality_opts) > 0 {
		if pbk, ok := proxy.Reality_opts["public-key"].(string); ok {
			trojan.Query.Pbk = pbk
		}
		if sid, ok := proxy.Reality_opts["short-id"].(string); ok {
			trojan.Query.Sid = sid
		}
	}

	return trojan
}

// buildTrojanProxy 将 Trojan 链接转换为 Clash Proxy，并合并链接内与输出配置中的证书校验策略。
func buildTrojanProxy(link Urls, config OutputConfig) (Proxy, error) {
	trojan, err := DecodeTrojanURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	if trojan.Name == "" {
		trojan.Name = fmt.Sprintf("%s:%s", trojan.Hostname, utils.GetPortString(trojan.Port))
	}
	wsOpts := map[string]interface{}{"path": trojan.Query.Path, "headers": map[string]interface{}{"Host": trojan.Query.Host}}
	DeleteOpts(wsOpts)
	skipCert := config.Cert || trojan.Query.AllowInsecure == 1
	return Proxy{Name: trojan.Name, Type: "trojan", Server: trojan.Hostname, Port: FlexPort(utils.GetPortInt(trojan.Port)), Password: trojan.Password, Client_fingerprint: trojan.Query.Fp, Sni: trojan.Query.Sni, Network: trojan.Query.Type, Flow: trojan.Query.Flow, Alpn: trojan.Query.Alpn, Ws_opts: wsOpts, Udp: config.Udp, Skip_cert_verify: skipCert, Dialer_proxy: link.DialerProxyName}, nil
}

// buildTrojanSurgeLine 将 Trojan 链接转换为 Surge 节点行。
// Surge 导出仅保留当前实现支持的核心字段，不会完整展开所有 Trojan 扩展参数。
func buildTrojanSurgeLine(link string, config OutputConfig) (string, string, error) {
	trojan, err := DecodeTrojanURL(link)
	if err != nil {
		return "", "", err
	}
	server := replaceSurgeHost(trojan.Hostname, config)
	skipCert := config.Cert || trojan.Query.AllowInsecure == 1
	line := fmt.Sprintf("%s = trojan, %s, %d, password=%s, udp-relay=%t, skip-cert-verify=%t", trojan.Name, server, utils.GetPortInt(trojan.Port), trojan.Password, config.Udp, skipCert)
	if trojan.Query.Sni != "" {
		line = fmt.Sprintf("%s, sni=%s", line, trojan.Query.Sni)
	}
	return line, trojan.Name, nil
}
