package protocol

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

func init() {
	base := newProtocolSpec("vless", []string{"vless://"}, "VLESS", "#7b1fa2", "V", VLESS{}, "Name", DecodeVLESSURL, EncodeVLESSURL, func(v VLESS) LinkIdentity {
		return buildIdentity("vless", v.Name, v.Server, utils.GetPortString(v.Port))
	},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic", Placeholder: "例如：日本-01"},
		FieldMeta{Name: "Server", Label: "服务器地址", Type: "string", Group: "basic", Placeholder: "example.com"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic", Placeholder: "443"},
		FieldMeta{Name: "Uuid", Label: "UUID", Type: "string", Group: "auth", Placeholder: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"},
		FieldMeta{Name: "Query.Flow", Label: "Flow", Type: "string", Group: "auth", Advanced: true},
		FieldMeta{Name: "Query.Security", Label: "安全类型", Type: "string", Group: "tls", Options: []string{"none", "tls", "reality"}},
		FieldMeta{Name: "Query.Sni", Label: "SNI", Type: "string", Group: "tls", Placeholder: "server.example.com"},
		FieldMeta{Name: "Query.Ech", Label: "ECH", Type: "string", Group: "tls", Advanced: true, Placeholder: "example.com+https://1.1.1.1/dns-query"},
		FieldMeta{Name: "Query.Alpn", Label: "ALPN", Type: "string", Group: "tls", Multiline: true, Advanced: true},
		FieldMeta{Name: "Query.Fp", Label: "指纹", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Query.Fingerprint", Label: "证书指纹", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Query.Sid", Label: "Short ID", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Query.Pbk", Label: "Public Key", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Query.AllowInsecure", Label: "跳过证书校验", Type: "int", Group: "tls", Advanced: true, Options: []string{"0", "1"}},
		FieldMeta{Name: "Query.Type", Label: "Network", Type: "string", Group: "transport", Options: []string{"tcp", "ws", "grpc", "http", "h2", "xhttp", "quic"}},
		FieldMeta{Name: "Query.Path", Label: "路径", Type: "string", Group: "transport", Placeholder: "/ws"},
		FieldMeta{Name: "Query.Host", Label: "Host", Type: "string", Group: "transport", Placeholder: "cdn.example.com"},
		FieldMeta{Name: "Query.HeaderType", Label: "Header Type", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.ServiceName", Label: "gRPC Service Name", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.Mode", Label: "gRPC Mode", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.Extra", Label: "XHTTP Extra", Type: "string", Group: "transport", Multiline: true, Advanced: true},
		FieldMeta{Name: "Query.Encryption", Label: "Encryption", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.PacketEncoding", Label: "Packet Encoding", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.MaxEarlyData", Label: "Early Data", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.EarlyDataHeader", Label: "Early Data Header", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.HttpUpgrade", Label: "HTTP Upgrade", Type: "int", Group: "transport", Advanced: true, Options: []string{"0", "1"}},
		FieldMeta{Name: "Query.HttpUpgradeFastOpen", Label: "HTTP Upgrade Fast Open", Type: "int", Group: "transport", Advanced: true, Options: []string{"0", "1"}},
		FieldMeta{Name: "Query.Method", Label: "HTTP Method", Type: "string", Group: "transport", Advanced: true},
	)
	MustRegisterProtocol(newProxyProtocolSpec(base, buildVLESSProxy, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "vless")
	}, ConvertProxyToVless, EncodeVLESSURL))
}

type VLESS struct {
	Name   string     `json:"name"`
	Uuid   string     `json:"uuid"`
	Server string     `json:"server"`
	Port   any        `json:"port"`
	Query  VLESSQuery `json:"query"`
}
type VLESSQuery struct {
	Security      string   `json:"security"`
	Alpn          []string `json:"alpn"`
	Sni           string   `json:"sni"`
	Ech           string   `json:"ech,omitempty"`
	Fp            string   `json:"fp"`
	Fingerprint   string   `json:"fingerprint,omitempty"`
	Sid           string   `json:"sid"`
	Pbk           string   `json:"pbk"`
	Flow          string   `json:"flow"`
	Encryption    string   `json:"encryption"`
	Type          string   `json:"type"`
	HeaderType    string   `json:"headerType"`
	Path          string   `json:"path"`
	Host          string   `json:"host"`
	ServiceName   string   `json:"serviceName,omitempty"`
	Mode          string   `json:"mode,omitempty"`
	Extra         string   `json:"extra,omitempty"`
	AllowInsecure int      `json:"allowInsecure,omitempty"` // 跳过证书验证
	// 新增：packet-encoding参数（xudp/packetaddr）
	PacketEncoding string `json:"packetEncoding,omitempty"`
	// 新增：ws传输层参数
	MaxEarlyData        int    `json:"maxEarlyData,omitempty"`        // Early Data首包长度阈值
	EarlyDataHeader     string `json:"earlyDataHeader,omitempty"`     // Early Data头名称
	HttpUpgrade         int    `json:"httpUpgrade,omitempty"`         // v2ray-http-upgrade (0/1)
	HttpUpgradeFastOpen int    `json:"httpUpgradeFastOpen,omitempty"` // v2ray-http-upgrade-fast-open (0/1)
	// 新增：http传输层参数
	Method string `json:"method,omitempty"` // HTTP请求方法
}

// buildVLESSProxy 将 VLESS 链接转换为 Clash Proxy，并根据传输层选择唯一一组传输配置输出。
// 输出阶段的证书校验、UDP 与前置代理配置会覆盖或补充链接中的原始字段。
func buildVLESSProxy(link Urls, config OutputConfig) (Proxy, error) {
	vless, err := DecodeVLESSURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	if vless.Name == "" {
		vless.Name = fmt.Sprintf("%s:%s", vless.Server, utils.GetPortString(vless.Port))
	}
	wsOpts := map[string]any{"path": vless.Query.Path, "headers": map[string]any{"Host": vless.Query.Host}}
	if vless.Query.MaxEarlyData > 0 {
		wsOpts["max-early-data"] = vless.Query.MaxEarlyData
	}
	if vless.Query.EarlyDataHeader != "" {
		wsOpts["early-data-header-name"] = vless.Query.EarlyDataHeader
	}
	if vless.Query.HttpUpgrade == 1 {
		wsOpts["v2ray-http-upgrade"] = true
	}
	if vless.Query.HttpUpgradeFastOpen == 1 {
		wsOpts["v2ray-http-upgrade-fast-open"] = true
	}
	h2Opts := map[string]any{}
	if vless.Query.Host != "" {
		h2Opts["host"] = []string{vless.Query.Host}
	}
	if vless.Query.Path != "" {
		h2Opts["path"] = vless.Query.Path
	}
	httpOpts := map[string]any{}
	if vless.Query.Method != "" {
		httpOpts["method"] = vless.Query.Method
	}
	if vless.Query.Path != "" {
		httpOpts["path"] = []string{vless.Query.Path}
	}
	if vless.Query.Host != "" {
		httpOpts["headers"] = map[string]any{"Host": []string{vless.Query.Host}}
	}
	grpcOpts := map[string]any{"grpc-service-name": vless.Query.ServiceName}
	if vless.Query.Mode != "" {
		grpcOpts["grpc-mode"] = vless.Query.Mode
	} else if vless.Query.ServiceName != "" {
		grpcOpts["grpc-mode"] = "gun"
	}
	xhttpOpts := buildVLESSXHTTPOpts(vless.Query)
	applyVLESSXHTTPSkipCertOverride(xhttpOpts, config.Cert)
	realityOpts := map[string]any{"public-key": vless.Query.Pbk, "short-id": vless.Query.Sid}
	DeleteOpts(wsOpts)
	DeleteOpts(h2Opts)
	DeleteOpts(httpOpts)
	DeleteOpts(grpcOpts)
	DeleteOpts(xhttpOpts)
	DeleteOpts(realityOpts)
	tls := vless.Query.Security != "" && vless.Query.Security != "none"
	skipCert := config.Cert || vless.Query.AllowInsecure == 1
	var finalWsOpts, finalH2Opts, finalHttpOpts, finalGrpcOpts, finalXHTTPOpts map[string]any
	switch vless.Query.Type {
	case "ws":
		finalWsOpts = wsOpts
	case "h2":
		finalH2Opts = h2Opts
	case "http":
		finalHttpOpts = httpOpts
	case "grpc":
		finalGrpcOpts = grpcOpts
	case "xhttp":
		finalXHTTPOpts = xhttpOpts
	}
	echOpts := buildVLESSECHOpts(vless.Query.Ech)
	return Proxy{Name: vless.Name, Type: "vless", Server: vless.Server, Port: FlexPort(utils.GetPortInt(vless.Port)), Servername: vless.Query.Sni, Uuid: vless.Uuid, Client_fingerprint: vless.Query.Fp, Fingerprint: vless.Query.Fingerprint, Network: vless.Query.Type, Flow: vless.Query.Flow, Encryption: vless.Query.Encryption, Alpn: vless.Query.Alpn, Packet_encoding: vless.Query.PacketEncoding, Ws_opts: finalWsOpts, H2_opts: finalH2Opts, Http_opts: finalHttpOpts, Grpc_opts: finalGrpcOpts, XHTTP_opts: finalXHTTPOpts, ECH_opts: echOpts, Reality_opts: realityOpts, Udp: config.Udp, Skip_cert_verify: skipCert, Tls: tls, Dialer_proxy: link.DialerProxyName}, nil
}

// EncodeVLESSURL 将 VLESS 结构编码为 v2ray 常见的明文 URL 形式。
// 编码时会按当前字段状态选择性输出扩展参数，并在名称缺失时回退为 server:port。
func EncodeVLESSURL(v VLESS) string {
	u := url.URL{
		Scheme: "vless",
		User:   url.User(v.Uuid),
		Host:   formatURLHostPort(v.Server, utils.GetPortString(v.Port)),
	}
	q := u.Query()

	// 基本参数
	q.Set("encryption", v.Query.Encryption)
	q.Set("security", v.Query.Security)
	q.Set("type", v.Query.Type)

	// TLS相关参数
	q.Set("sni", v.Query.Sni)
	q.Set("ech", v.Query.Ech)
	q.Set("fp", v.Query.Fp)
	q.Set("pcs", v.Query.Fingerprint)
	if len(v.Query.Alpn) > 0 {
		q.Set("alpn", strings.Join(v.Query.Alpn, ","))
	}

	// Reality参数
	q.Set("pbk", v.Query.Pbk)
	q.Set("sid", v.Query.Sid)

	// VLESS特有参数
	q.Set("flow", v.Query.Flow)
	q.Set("headerType", v.Query.HeaderType)
	if v.Query.PacketEncoding != "" {
		q.Set("packetEncoding", v.Query.PacketEncoding)
	}

	// 传输层通用参数
	q.Set("path", v.Query.Path)
	q.Set("host", v.Query.Host)

	// gRPC参数
	if v.Query.ServiceName != "" {
		q.Set("serviceName", v.Query.ServiceName)
	}
	if v.Query.Mode != "" {
		q.Set("mode", v.Query.Mode)
	}
	if v.Query.Extra != "" {
		q.Set("extra", v.Query.Extra)
	}

	// ws传输层参数
	if v.Query.MaxEarlyData > 0 {
		q.Set("ed", strconv.Itoa(v.Query.MaxEarlyData))
	}
	if v.Query.EarlyDataHeader != "" {
		q.Set("eh", v.Query.EarlyDataHeader)
	}
	// 保留 HTTP Upgrade 扩展，避免 Mihomo 配置经 VLESS URI 往返转换后退化为普通 WebSocket。
	if v.Query.HttpUpgrade == 1 {
		q.Set("httpUpgrade", "1")
	}
	if v.Query.HttpUpgradeFastOpen == 1 {
		q.Set("httpUpgradeFastOpen", "1")
	}

	// http传输层参数
	if v.Query.Method != "" {
		q.Set("method", v.Query.Method)
	}

	// 跳过证书验证
	if v.Query.AllowInsecure == 1 {
		q.Set("allowInsecure", "1")
	}

	// 检查query是否有空值，有的话删除
	for k, val := range q {
		if val[0] == "" {
			delete(q, k)
		}
	}
	u.RawQuery = q.Encode()

	// 如果没有name则用服务器加端口
	if v.Name == "" {
		u.Fragment = v.Server + ":" + utils.GetPortString(v.Port)
	} else {
		u.Fragment = v.Name
	}
	return u.String()
}

// DecodeVLESSURL 解析明文 VLESS URL，并兼容当前仓库支持的多类传输层扩展参数。
// 端口默认值会随 security 语义变化，且 packetEncoding 与 packet_encoding 两种写法都会被接受。
func DecodeVLESSURL(s string) (VLESS, error) {
	if !strings.HasPrefix(s, "vless://") {
		return VLESS{}, fmt.Errorf("非vless协议: %s", s)
	}

	// 直接解析URL（v2ray格式是明文URL，不需要base64解码）
	u, err := url.Parse(s)
	if err != nil {
		return VLESS{}, fmt.Errorf("url parse error: %v", err)
	}

	uuid := u.User.Username()
	if !utils.IsUUID(uuid) {
		utils.Error("❌节点解析错误：%v  【节点：%s】", "UUID格式错误", s)
		return VLESS{}, fmt.Errorf("uuid格式错误:%s", uuid)
	}

	// 处理服务器地址（支持IPv6格式[::1]）
	hostname := utils.UnwrapIPv6Host(u.Hostname())

	// 处理端口
	rawPort := u.Port()
	if rawPort == "" {
		security := u.Query().Get("security")
		if security == "none" || security == "" {
			rawPort = "80"
		} else {
			rawPort = "443"
		}
	}
	port, _ := strconv.Atoi(rawPort)

	// 解析基本参数
	encryption := u.Query().Get("encryption")
	security := u.Query().Get("security")
	types := u.Query().Get("type")
	flow := u.Query().Get("flow")
	headerType := u.Query().Get("headerType")
	pbk := u.Query().Get("pbk")
	sid := u.Query().Get("sid")
	fp := u.Query().Get("fp")
	fingerprint := sanitizeCertificateFingerprint(u.Query().Get("pcs"))
	if fingerprint == "" {
		fingerprint = sanitizeCertificateFingerprint(u.Query().Get("hpkp"))
	}
	sni := u.Query().Get("sni")
	ech := u.Query().Get("ech")
	path := u.Query().Get("path")
	host := u.Query().Get("host")
	serviceName := u.Query().Get("serviceName")
	mode := u.Query().Get("mode")
	extra := u.Query().Get("extra")

	// 解析 alpn 参数（逗号分隔）
	alpns := u.Query().Get("alpn")
	var alpn []string
	if alpns != "" {
		alpn = strings.Split(alpns, ",")
	}

	// 解析 allowInsecure 参数
	allowInsecure := 0
	insecureStr := u.Query().Get("allowInsecure")
	if insecureStr == "1" || insecureStr == "true" {
		allowInsecure = 1
	}

	// 解析 packet-encoding 参数（packetEncoding 或 packet_encoding）
	packetEncoding := u.Query().Get("packetEncoding")
	if packetEncoding == "" {
		packetEncoding = u.Query().Get("packet_encoding")
	}

	// 解析 ws 传输层参数
	maxEarlyData := 0
	if ed := u.Query().Get("ed"); ed != "" {
		maxEarlyData, _ = strconv.Atoi(ed)
	}
	earlyDataHeader := u.Query().Get("eh")

	// 解析 v2ray-http-upgrade 参数
	httpUpgrade := 0
	if hup := u.Query().Get("httpUpgrade"); hup == "1" || hup == "true" {
		httpUpgrade = 1
	}
	httpUpgradeFastOpen := 0
	if hupfo := u.Query().Get("httpUpgradeFastOpen"); hupfo == "1" || hupfo == "true" {
		httpUpgradeFastOpen = 1
	}

	// 解析 http 传输层参数
	method := u.Query().Get("method")

	// 解析名称（URL fragment）
	name := u.Fragment
	if name == "" {
		name = hostname + ":" + rawPort
	}

	if utils.CheckEnvironment() {
		fmt.Println("uuid:", uuid)
		fmt.Println("hostname:", hostname)
		fmt.Println("port:", port)
		fmt.Println("encryption:", encryption)
		fmt.Println("security:", security)
		fmt.Println("type:", types)
		fmt.Println("flow:", flow)
		fmt.Println("headerType:", headerType)
		fmt.Println("pbk:", pbk)
		fmt.Println("sid:", sid)
		fmt.Println("fp:", fp)
		fmt.Println("fingerprint:", fingerprint)
		fmt.Println("alpn:", alpn)
		fmt.Println("sni:", sni)
		fmt.Println("ech:", ech)
		fmt.Println("path:", path)
		fmt.Println("host:", host)
		fmt.Println("serviceName:", serviceName)
		fmt.Println("mode:", mode)
		fmt.Println("packetEncoding:", packetEncoding)
		fmt.Println("maxEarlyData:", maxEarlyData)
		fmt.Println("earlyDataHeader:", earlyDataHeader)
		fmt.Println("httpUpgrade:", httpUpgrade)
		fmt.Println("method:", method)
		fmt.Println("name:", name)
	}

	return VLESS{
		Name:   name,
		Uuid:   uuid,
		Server: hostname,
		Port:   port,
		Query: VLESSQuery{
			Security:            security,
			Alpn:                alpn,
			Sni:                 sni,
			Ech:                 ech,
			Fp:                  fp,
			Fingerprint:         fingerprint,
			Sid:                 sid,
			Pbk:                 pbk,
			Flow:                flow,
			Encryption:          encryption,
			Type:                types,
			HeaderType:          headerType,
			Path:                path,
			Host:                host,
			ServiceName:         serviceName,
			Mode:                mode,
			Extra:               extra,
			AllowInsecure:       allowInsecure,
			PacketEncoding:      packetEncoding,
			MaxEarlyData:        maxEarlyData,
			EarlyDataHeader:     earlyDataHeader,
			HttpUpgrade:         httpUpgrade,
			HttpUpgradeFastOpen: httpUpgradeFastOpen,
			Method:              method,
		},
	}, nil
}

// ConvertProxyToVless 将 Clash Proxy 还原为 VLESS 结构，用于重新生成分享链接。
// 该转换会按当前实现折叠多种 transport 配置来源，因此在部分字段上属于有损回写。
func ConvertProxyToVless(proxy Proxy) VLESS {
	vless := VLESS{
		Name:   proxy.Name,
		Uuid:   proxy.Uuid,
		Server: proxy.Server,
		Port:   int(proxy.Port),
		Query: VLESSQuery{
			Sni:            proxy.Servername,
			Ech:            buildVLESSECHQuery(proxy.ECH_opts),
			Fp:             proxy.Client_fingerprint,
			Fingerprint:    sanitizeCertificateFingerprint(proxy.Fingerprint),
			Flow:           proxy.Flow,
			Encryption:     proxy.Encryption,
			Alpn:           proxy.Alpn,
			Type:           proxy.Network,
			PacketEncoding: proxy.Packet_encoding,
		},
	}

	// 处理跳过证书验证
	if proxy.Skip_cert_verify {
		vless.Query.AllowInsecure = 1
	}

	// 处理 security 参数（TLS/Reality/none）
	if len(proxy.Reality_opts) > 0 {
		vless.Query.Security = "reality"
		if pbk, ok := proxy.Reality_opts["public-key"].(string); ok {
			vless.Query.Pbk = pbk
		}
		if sid, ok := proxy.Reality_opts["short-id"].(string); ok {
			vless.Query.Sid = sid
		}
	} else if proxy.Tls {
		vless.Query.Security = "tls"
	} else {
		vless.Query.Security = "none"
	}

	// 处理 ws_opts
	if len(proxy.Ws_opts) > 0 {
		if path, ok := proxy.Ws_opts["path"].(string); ok {
			vless.Query.Path = path
		}
		if headers, ok := proxy.Ws_opts["headers"].(map[string]any); ok {
			if host, ok := headers["Host"].(string); ok {
				vless.Query.Host = host
			}
		}
		if ed, ok := proxy.Ws_opts["max-early-data"].(int); ok {
			vless.Query.MaxEarlyData = ed
		}
		if edh, ok := proxy.Ws_opts["early-data-header-name"].(string); ok {
			vless.Query.EarlyDataHeader = edh
		}
		if hup, ok := proxy.Ws_opts["v2ray-http-upgrade"].(bool); ok && hup {
			vless.Query.HttpUpgrade = 1
		}
		if hupfo, ok := proxy.Ws_opts["v2ray-http-upgrade-fast-open"].(bool); ok && hupfo {
			vless.Query.HttpUpgradeFastOpen = 1
		}
	}

	// 处理 h2_opts
	if len(proxy.H2_opts) > 0 {
		if path, ok := proxy.H2_opts["path"].(string); ok {
			vless.Query.Path = path
		}
		if hosts, ok := proxy.H2_opts["host"].([]string); ok && len(hosts) > 0 {
			vless.Query.Host = hosts[0]
		}
		if host, ok := proxy.H2_opts["host"].([]any); ok && len(host) > 0 {
			if h, ok := host[0].(string); ok {
				vless.Query.Host = h
			}
		}
	}

	// 处理 http_opts
	if len(proxy.Http_opts) > 0 {
		if method, ok := proxy.Http_opts["method"].(string); ok {
			vless.Query.Method = method
		}
		if paths, ok := proxy.Http_opts["path"].([]string); ok && len(paths) > 0 {
			vless.Query.Path = paths[0]
		}
		if paths, ok := proxy.Http_opts["path"].([]any); ok && len(paths) > 0 {
			if p, ok := paths[0].(string); ok {
				vless.Query.Path = p
			}
		}
		if headers, ok := proxy.Http_opts["headers"].(map[string]any); ok {
			if hosts, ok := headers["Host"].([]any); ok && len(hosts) > 0 {
				if h, ok := hosts[0].(string); ok {
					vless.Query.Host = h
				}
			}
		}
	}

	// 处理 grpc_opts
	if len(proxy.Grpc_opts) > 0 {
		if sn, ok := proxy.Grpc_opts["grpc-service-name"].(string); ok {
			vless.Query.ServiceName = sn
		}
		if mode, ok := proxy.Grpc_opts["grpc-mode"].(string); ok && mode == "multi" {
			vless.Query.Mode = "multi"
		}
	}

	if proxy.Network == "xhttp" {
		populateVLESSQueryFromXHTTPOpts(&vless.Query, proxy.XHTTP_opts)
	}

	return vless
}

func buildVLESSXHTTPOpts(query VLESSQuery) map[string]any {
	xhttpOpts := map[string]any{}
	if query.Path != "" {
		xhttpOpts["path"] = query.Path
	}
	if query.Host != "" {
		xhttpOpts["host"] = query.Host
	}
	if query.Mode != "" {
		xhttpOpts["mode"] = query.Mode
	}
	mergeXHTTPExtraMap(xhttpOpts, parseVLESSXHTTPExtra(query.Extra))
	return xhttpOpts
}

func populateVLESSQueryFromXHTTPOpts(query *VLESSQuery, xhttpOpts map[string]any) {
	if query == nil || len(xhttpOpts) == 0 {
		return
	}
	if path, ok := xhttpOpts["path"].(string); ok {
		query.Path = path
	}
	if host, ok := xhttpOpts["host"].(string); ok {
		query.Host = host
	}
	if mode, ok := xhttpOpts["mode"].(string); ok {
		query.Mode = mode
	}
	extra := buildVLESSXHTTPExtra(xhttpOpts)
	if extra != "" {
		query.Extra = extra
	}
}

func parseVLESSXHTTPExtra(extra string) map[string]any {
	extra = strings.TrimSpace(extra)
	if extra == "" {
		return nil
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(extra), &parsed); err != nil {
		return nil
	}
	return normalizeVLESSXHTTPExtra(parsed)
}

func buildVLESSECHOpts(ech string) map[string]any {
	ech = strings.TrimSpace(ech)
	if ech == "" {
		return nil
	}
	opts := map[string]any{"enable": true}
	if strings.Contains(ech, "://") {
		parts := strings.SplitN(ech, "+", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && !strings.Contains(parts[0], "://") {
			opts["query-server-name"] = strings.TrimSpace(parts[0])
		}
		DeleteOpts(opts)
		if len(opts) == 1 {
			return opts
		}
		return opts
	}
	opts["config"] = ech
	DeleteOpts(opts)
	if len(opts) == 0 {
		return nil
	}
	return opts
}

func buildVLESSECHQuery(echOpts map[string]any) string {
	if len(echOpts) == 0 {
		return ""
	}
	config, _ := echOpts["config"].(string)
	return strings.TrimSpace(config)
}

func normalizeVLESSXHTTPExtra(extra map[string]any) map[string]any {
	if len(extra) == 0 {
		return nil
	}
	normalized := map[string]any{}
	if headers, ok := extra["headers"].(map[string]any); ok && len(headers) > 0 {
		normalized["headers"] = headers
	}
	if noGRPCHeader, ok := extra["noGRPCHeader"]; ok {
		normalized["no-grpc-header"] = noGRPCHeader
	}
	if xPaddingBytes, ok := extra["xPaddingBytes"]; ok {
		normalized["x-padding-bytes"] = xPaddingBytes
	}
	if downloadSettings, ok := extra["downloadSettings"].(map[string]any); ok && len(downloadSettings) > 0 {
		if normalizedDownloadSettings := normalizeVLESSXHTTPDownloadSettings(downloadSettings); len(normalizedDownloadSettings) > 0 {
			normalized["download-settings"] = normalizedDownloadSettings
		}
	}
	DeleteOpts(normalized)
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalizeVLESSXHTTPDownloadSettings(settings map[string]any) map[string]any {
	if len(settings) == 0 {
		return nil
	}
	normalized := map[string]any{}
	for key, value := range settings {
		switch key {
		case "path", "host", "headers", "server", "port", "tls", "alpn", "certificate", "servername":
			normalized[key] = value
		case "noGRPCHeader":
			normalized["no-grpc-header"] = value
		case "xPaddingBytes":
			normalized["x-padding-bytes"] = value
		case "echOpts":
			if echOpts, ok := value.(map[string]any); ok && len(echOpts) > 0 {
				if normalizedECHOpts := normalizeVLESSECHOptsMap(echOpts); len(normalizedECHOpts) > 0 {
					normalized["ech-opts"] = normalizedECHOpts
				}
			}
		case "realityOpts":
			normalized["reality-opts"] = value
		case "skipCertVerify":
			normalized["skip-cert-verify"] = value
		case "fingerprint":
			normalized["fingerprint"] = value
		case "privateKey":
			normalized["private-key"] = value
		case "serverName":
			normalized["servername"] = value
		case "clientFingerprint":
			normalized["client-fingerprint"] = value
		}
	}
	DeleteOpts(normalized)
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func buildVLESSXHTTPExtra(xhttpOpts map[string]any) string {
	if len(xhttpOpts) == 0 {
		return ""
	}
	extra := map[string]any{}
	if headers, ok := xhttpOpts["headers"].(map[string]any); ok && len(headers) > 0 {
		extra["headers"] = headers
	}
	if noGRPCHeader, ok := xhttpOpts["no-grpc-header"]; ok {
		extra["noGRPCHeader"] = noGRPCHeader
	}
	if xPaddingBytes, ok := xhttpOpts["x-padding-bytes"]; ok {
		extra["xPaddingBytes"] = xPaddingBytes
	}
	if downloadSettings, ok := xhttpOpts["download-settings"].(map[string]any); ok && len(downloadSettings) > 0 {
		if extraDownloadSettings := buildVLESSXHTTPExtraDownloadSettings(downloadSettings); len(extraDownloadSettings) > 0 {
			extra["downloadSettings"] = extraDownloadSettings
		}
	}
	if len(extra) == 0 {
		return ""
	}
	encoded, err := json.Marshal(extra)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func buildVLESSXHTTPExtraDownloadSettings(settings map[string]any) map[string]any {
	if len(settings) == 0 {
		return nil
	}
	extraSettings := map[string]any{}
	for key, value := range settings {
		switch key {
		case "path", "host", "headers", "server", "port", "tls", "alpn", "certificate", "fingerprint", "servername":
			extraSettings[key] = value
		case "no-grpc-header":
			extraSettings["noGRPCHeader"] = value
		case "x-padding-bytes":
			extraSettings["xPaddingBytes"] = value
		case "ech-opts":
			if echOpts, ok := value.(map[string]any); ok && len(echOpts) > 0 {
				if extraECHOpts := buildVLESSECHExtraOptsMap(echOpts); len(extraECHOpts) > 0 {
					extraSettings["echOpts"] = extraECHOpts
				}
			}
		case "reality-opts":
			extraSettings["realityOpts"] = value
		case "skip-cert-verify":
			extraSettings["skipCertVerify"] = value
		case "private-key":
			extraSettings["privateKey"] = value
		case "client-fingerprint":
			extraSettings["clientFingerprint"] = value
		}
	}
	if len(extraSettings) == 0 {
		return nil
	}
	return extraSettings
}

func normalizeVLESSECHOptsMap(echOpts map[string]any) map[string]any {
	if len(echOpts) == 0 {
		return nil
	}
	normalized := map[string]any{}
	for key, value := range echOpts {
		switch key {
		case "enable", "config":
			normalized[key] = value
		case "queryServerName", "query-server-name":
			normalized["query-server-name"] = value
		}
	}
	DeleteOpts(normalized)
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func buildVLESSECHExtraOptsMap(echOpts map[string]any) map[string]any {
	if len(echOpts) == 0 {
		return nil
	}
	extraECHOpts := map[string]any{}
	for key, value := range echOpts {
		switch key {
		case "enable", "config":
			extraECHOpts[key] = value
		case "query-server-name", "queryServerName":
			extraECHOpts["queryServerName"] = value
		}
	}
	DeleteOpts(extraECHOpts)
	if len(extraECHOpts) == 0 {
		return nil
	}
	return extraECHOpts
}

func mergeXHTTPExtraMap(target map[string]any, extra map[string]any) {
	if len(target) == 0 || len(extra) == 0 {
		for key, value := range extra {
			target[key] = value
		}
		return
	}
	for key, value := range extra {
		target[key] = value
	}
}

func applyVLESSXHTTPSkipCertOverride(xhttpOpts map[string]any, forceSkipCert bool) {
	if !forceSkipCert || len(xhttpOpts) == 0 {
		return
	}
	downloadSettings, ok := xhttpOpts["download-settings"].(map[string]any)
	if !ok {
		return
	}
	downloadSettings["skip-cert-verify"] = true
}
