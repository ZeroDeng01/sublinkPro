package protocol

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

func init() {
	base := newProtocolSpec("hysteria2", []string{"hysteria2://", "hy2://"}, "Hysteria2", "#ef6c00", "H", HY2{}, "Name", DecodeHY2URL, EncodeHY2URL, func(h HY2) LinkIdentity {
		return buildIdentity("hysteria2", h.Name, h.Host, utils.GetPortString(h.Port))
	},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic"},
		FieldMeta{Name: "Host", Label: "服务器地址", Type: "string", Group: "basic"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic"},
		FieldMeta{Name: "MPort", Label: "端口跳跃", Type: "string", Group: "basic", Advanced: true},
		FieldMeta{Name: "Password", Label: "密码", Type: "string", Group: "auth", Secret: true},
		FieldMeta{Name: "Auth", Label: "认证", Type: "string", Group: "auth", Secret: true, Advanced: true},
		FieldMeta{Name: "UpMbps", Label: "上行 Mbps", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "DownMbps", Label: "下行 Mbps", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "Obfs", Label: "混淆类型", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "ObfsPassword", Label: "混淆密码", Type: "string", Group: "auth", Secret: true, Advanced: true},
		FieldMeta{Name: "Sni", Label: "SNI", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Peer", Label: "Peer", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "ALPN", Label: "ALPN", Type: "string", Group: "tls", Advanced: true, Multiline: true},
		FieldMeta{Name: "ClientFingerprint", Label: "指纹", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Fingerprint", Label: "证书指纹", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Insecure", Label: "跳过证书校验", Type: "int", Group: "tls", Advanced: true, Options: []string{"0", "1"}},
	)
	MustRegisterProtocol(newProxySurgeProtocolSpec(base, buildHY2Proxy, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "hysteria2")
	}, ConvertProxyToHy2, EncodeHY2URL, buildHY2SurgeLine))
}

type HY2 struct {
	Password          string
	Host              string
	Port              any
	MPort             string
	Insecure          int
	Peer              string
	Auth              string
	UpMbps            int
	DownMbps          int
	ALPN              []string
	Name              string
	Sni               string
	Obfs              string
	ObfsPassword      string
	ClientFingerprint string // 客户端指纹
	Fingerprint       string // 服务端证书 SHA-256 指纹
}

// EncodeHY2URL 将 Hysteria2 结构编码为 hy2:// 链接。
// 编码时会清理空值与零值字段，并在名称缺失时回退为 host:port。
func EncodeHY2URL(hy2 HY2) string {
	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if hy2.Name == "" {
		hy2.Name = fmt.Sprintf("%s:%s", hy2.Host, utils.GetPortString(hy2.Port))
	}
	u := url.URL{
		Scheme:   "hy2",
		User:     url.User(hy2.Password),
		Host:     formatURLHostPort(hy2.Host, utils.GetPortString(hy2.Port)),
		Fragment: hy2.Name,
	}
	q := u.Query()
	q.Set("insecure", strconv.Itoa(hy2.Insecure))
	q.Set("peer", hy2.Peer)
	q.Set("auth", hy2.Auth)
	q.Set("mport", hy2.MPort)
	q.Set("upmbps", strconv.Itoa(hy2.UpMbps))
	q.Set("downmbps", strconv.Itoa(hy2.DownMbps))
	q.Set("sni", hy2.Sni)
	q.Set("obfs", hy2.Obfs)
	q.Set("obfs-password", hy2.ObfsPassword)
	// 客户端指纹参数
	q.Set("fp", hy2.ClientFingerprint)
	q.Set("pinSHA256", hy2.Fingerprint)
	// alpn 参数支持
	if len(hy2.ALPN) > 0 {
		q.Set("alpn", strings.Join(hy2.ALPN, ","))
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

// DecodeHY2URL 解析 hy2:// 与 hysteria2:// 链接，并兼容 auth 缺省时回退为 password 的写法。
// 当链接未显式给出端口时，解析结果中的端口字段会按当前约定回退到 443。
func DecodeHY2URL(s string) (HY2, error) {
	u, err := url.Parse(s)
	if err != nil {
		return HY2{}, fmt.Errorf("解析失败的URL: %s,错误:%s", s, err)
	}
	if u.Scheme != "hy2" && u.Scheme != "hysteria2" {
		return HY2{}, fmt.Errorf("非hy2协议: %s", s)
	}
	password := u.User.Username()
	server := u.Hostname()
	rawPort := u.Port()
	if rawPort == "" {
		rawPort = "443"
	}
	port, _ := strconv.Atoi(rawPort)
	insecure, _ := strconv.Atoi(u.Query().Get("insecure"))
	auth := u.Query().Get("auth")
	if auth == "" {
		auth = password
	}
	upMbps, _ := strconv.Atoi(u.Query().Get("upmbps"))
	downMbps, _ := strconv.Atoi(u.Query().Get("downmbps"))
	alpns := u.Query().Get("alpn")
	mport := u.Query().Get("mport")
	alpn := strings.Split(alpns, ",")
	if alpns == "" {
		alpn = nil
	}
	sni := u.Query().Get("sni")
	obfs := u.Query().Get("obfs")
	obfsPassword := u.Query().Get("obfs-password")
	clientFingerprint := u.Query().Get("fp")
	fingerprint := sanitizeCertificateFingerprint(u.Query().Get("pinSHA256"))
	name := u.Fragment
	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if name == "" {
		name = server + ":" + u.Port()
	}
	if utils.CheckEnvironment() {
		fmt.Println("password:", password)
		fmt.Println("server:", server)
		fmt.Println("port:", port)
		fmt.Println("mport:", mport)
		fmt.Println("insecure:", insecure)
		fmt.Println("auth:", auth)
		fmt.Println("upMbps:", upMbps)
		fmt.Println("downMbps:", downMbps)
		fmt.Println("alpn:", alpn)
		fmt.Println("sni:", sni)
		fmt.Println("obfs:", obfs)
		fmt.Println("obfsPassword:", obfsPassword)
		fmt.Println("fp:", clientFingerprint)
		fmt.Println("fingerprint:", fingerprint)
		fmt.Println("name:", name)
	}
	return HY2{
		Password:          password,
		Host:              server,
		Port:              port,
		MPort:             mport,
		Insecure:          insecure,
		Auth:              auth,
		UpMbps:            upMbps,
		DownMbps:          downMbps,
		ALPN:              alpn,
		Name:              name,
		Sni:               sni,
		Obfs:              obfs,
		ObfsPassword:      obfsPassword,
		ClientFingerprint: clientFingerprint,
		Fingerprint:       fingerprint,
	}, nil
}

// ConvertProxyToHy2 将 Proxy 结构体转换为 HY2 结构体
// 用于从 Clash 格式的代理配置生成 Hysteria2 链接
func ConvertProxyToHy2(proxy Proxy) HY2 {
	hy2 := HY2{
		Password:          proxy.Password,
		Host:              proxy.Server,
		Port:              int(proxy.Port),
		MPort:             proxy.Ports,
		Auth:              proxy.Auth_str,
		UpMbps:            int(proxy.Up),
		DownMbps:          int(proxy.Down),
		ALPN:              proxy.Alpn,
		Name:              proxy.Name,
		Sni:               proxy.Sni,
		Obfs:              proxy.Obfs,
		ObfsPassword:      proxy.Obfs_password,
		ClientFingerprint: proxy.Client_fingerprint,
		Fingerprint:       sanitizeCertificateFingerprint(proxy.Fingerprint),
		Peer:              proxy.Peer,
	}

	// 处理跳过证书验证
	if proxy.Skip_cert_verify {
		hy2.Insecure = 1
	}

	// 如果SNI为空，尝试使用Servername
	if hy2.Sni == "" && proxy.Servername != "" {
		hy2.Sni = proxy.Servername
	}

	return hy2
}

// buildHY2Proxy 将 Hysteria2 链接转换为 Clash Proxy。
// 当 mport 存在时优先输出端口跳跃配置，否则回退为单端口字段，并合并输出阶段证书校验策略。
func buildHY2Proxy(link Urls, config OutputConfig) (Proxy, error) {
	hy2, err := DecodeHY2URL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	if hy2.Name == "" {
		hy2.Name = fmt.Sprintf("%s:%s", hy2.Host, utils.GetPortString(hy2.Port))
	}
	skipCert := config.Cert || hy2.Insecure == 1
	proxy := Proxy{Name: hy2.Name, Type: "hysteria2", Server: hy2.Host, Auth: hy2.Auth, Sni: hy2.Sni, Alpn: hy2.ALPN, Obfs: hy2.Obfs, Password: hy2.Password, Obfs_password: hy2.ObfsPassword, Client_fingerprint: hy2.ClientFingerprint, Fingerprint: hy2.Fingerprint, Up: Mbps(hy2.UpMbps), Down: Mbps(hy2.DownMbps), Up_Speed: Mbps(hy2.UpMbps), Down_Speed: Mbps(hy2.DownMbps), Udp: true, Skip_cert_verify: skipCert, Dialer_proxy: link.DialerProxyName}
	if hy2.MPort != "" {
		proxy.Ports = hy2.MPort
	} else {
		proxy.Port = FlexPort(utils.GetPortInt(hy2.Port))
	}
	return proxy, nil
}

// buildHY2SurgeLine 将 Hysteria2 链接转换为 Surge 节点行。
// Surge 导出仅使用当前受支持的密码、端口与 TLS 相关字段，属于精简映射而非完整保真。
func buildHY2SurgeLine(link string, config OutputConfig) (string, string, error) {
	hy2, err := DecodeHY2URL(link)
	if err != nil {
		return "", "", err
	}
	server := replaceSurgeHost(hy2.Host, config)
	skipCert := config.Cert || hy2.Insecure == 1
	port := utils.GetPortInt(hy2.Port)
	// 端口跳跃：hy2:// 的 mport 透传为 Surge hysteria2 的 port-hopping。
	// 先统一做白名单校验并规范化分隔符；非法或空串则本行不输出 port-hopping。
	mphop := surgePortHopping(hy2.MPort)
	// 主端口必须为 1-65535：显式 :0 或越界值会被 Surge 判为
	// "The value of port is invalid"。此时若配置了合法端口跳跃则取 mport 首段作占位端口，
	// 否则回退 443（与协议层 hysteria2 无端口时的默认端口一致）。
	if port < 1 || port > 65535 {
		if mphop != "" {
			port = mportFirstPort(mphop)
		}
		if port < 1 || port > 65535 {
			port = 443
		}
	}
	line := fmt.Sprintf("%s = hysteria2, %s, %d, password=%s, udp-relay=%t, skip-cert-verify=%t", hy2.Name, server, port, hy2.Password, true, skipCert)
	if hy2.Sni != "" {
		line = fmt.Sprintf("%s, sni=%s", line, hy2.Sni)
	}
	if hy2.Fingerprint != "" {
		line = fmt.Sprintf("%s, server-cert-fingerprint-sha256=%s", line, hy2.Fingerprint)
	}
	// 端口跳跃已在上方校验并规范化（mphop），非空才输出。
	if mphop != "" {
		line = fmt.Sprintf("%s, port-hopping=%s", line, mphop)
	}
	return line, hy2.Name, nil
}

// surgePortHopping 将 Hysteria2/mihomo 的 mport 端口跳跃串规范化为 Surge 语法。
// mihomo 允许逗号或斜杠分隔多个端口段，并容忍端口段周围空白；这里先统一为
// 规范端口/范围列表，再用 Surge 要求的分号连接。分号不是 mihomo 源语法，
// 含分号或其他非法内容时整段丢弃，避免将未校验文本拼入 Surge profile。
func surgePortHopping(mport string) string {
	mport = strings.TrimSpace(mport)
	if mport == "" || strings.Contains(mport, ";") {
		return ""
	}
	mport = strings.ReplaceAll(mport, "/", ",")

	// 语义校验：每段端口/端点须在 1-65535 且范围 start<=end。Surge 对降序范围、
	// 端口 0 或超界会判 "The value of port-hopping is invalid"，导致整个 profile 无法加载；
	// 任一非法则该 mport 整体视为无效，不输出 port-hopping。
	segments := strings.Split(mport, ",")
	normalized := make([]string, 0, len(segments))
	for _, seg := range segments {
		value, ok := normalizePortRange(seg)
		if !ok {
			return ""
		}
		normalized = append(normalized, value)
	}
	return strings.Join(normalized, ";")
}

// normalizePortRange 校验并规范化单段端口跳跃：`port` 或 `port-port`。
// 端点须在 1-65535 且 start<=end；端点周围空白与前导零会被清理。
func normalizePortRange(seg string) (string, bool) {
	parts := strings.Split(strings.TrimSpace(seg), "-")
	if len(parts) < 1 || len(parts) > 2 {
		return "", false
	}
	lo, ok := portBound(strings.TrimSpace(parts[0]))
	if !ok {
		return "", false
	}
	if len(parts) == 1 {
		return strconv.Itoa(lo), true
	}
	hi, ok := portBound(strings.TrimSpace(parts[1]))
	if !ok || lo > hi {
		return "", false
	}
	return fmt.Sprintf("%d-%d", lo, hi), true
}

// portBound 解析单个端口/端点并在 1-65535 范围内返回。
func portBound(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 65535 {
		return 0, false
	}
	return n, true
}

// mportFirstPort 从已规范化的 Surge port-hopping 中提取首段起始端口，
// 供主端口无效时作占位主端口。解析失败返回 0，由调用方回退默认端口。
func mportFirstPort(portHopping string) int {
	firstSegment := strings.SplitN(portHopping, ";", 2)[0]
	firstPort := strings.SplitN(firstSegment, "-", 2)[0]
	if n, ok := portBound(firstPort); ok {
		return n
	}
	return 0
}
