package protocol

import (
	"encoding/json"
	"fmt"
	"math"
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
		FieldMeta{Name: "Query.Type", Label: "传输层", Type: "string", Group: "transport", Options: []string{"tcp", "ws", "httpupgrade", "grpc"}},
		FieldMeta{Name: "Query.Path", Label: "路径", Type: "string", Group: "transport", Placeholder: "/ws", Advanced: true},
		FieldMeta{Name: "Query.Host", Label: "Host", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.Headers", Label: "Headers", Type: "string", Group: "transport", Multiline: true, Advanced: true},
		FieldMeta{Name: "Query.MaxEarlyData", Label: "Early Data", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.EarlyDataHeader", Label: "Early Data Header", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.HttpUpgrade", Label: "HTTP Upgrade", Type: "int", Group: "transport", Advanced: true, Options: []string{"0", "1"}},
		FieldMeta{Name: "Query.HttpUpgradeFastOpen", Label: "HTTP Upgrade Fast Open", Type: "int", Group: "transport", Advanced: true, Options: []string{"0", "1"}},
		FieldMeta{Name: "Query.ServiceName", Label: "gRPC Service Name", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.Mode", Label: "gRPC Mode", Type: "string", Group: "transport", Advanced: true, Options: []string{"gun"}},
		FieldMeta{Name: "Query.GrpcUserAgent", Label: "gRPC User Agent", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.PingInterval", Label: "gRPC Ping Interval", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.MaxConnections", Label: "gRPC Max Connections", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.MinStreams", Label: "gRPC Min Streams", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.MaxStreams", Label: "gRPC Max Streams", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.Flow", Label: "Flow", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Query.Security", Label: "安全类型", Type: "string", Group: "tls", Options: []string{"tls", "reality"}, Advanced: true},
		FieldMeta{Name: "Query.Sni", Label: "SNI", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Query.Peer", Label: "Peer", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Query.Alpn", Label: "ALPN", Type: "string", Group: "tls", Advanced: true, Multiline: true},
		FieldMeta{Name: "Query.Fp", Label: "指纹", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Query.Fingerprint", Label: "证书指纹", Type: "string", Group: "tls", Advanced: true},
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
	Port     any         `json:"port"`
	Query    TrojanQuery `json:"query,omitempty"`
	Name     string      `json:"name"`
	Type     string      `json:"type"`
}
type TrojanQuery struct {
	Peer                string   `json:"peer,omitempty"`
	Type                string   `json:"type,omitempty"`
	Path                string   `json:"path,omitempty"`
	Security            string   `json:"security,omitempty"`
	Fp                  string   `json:"fp,omitempty"`
	Fingerprint         string   `json:"fingerprint,omitempty"`
	AllowInsecure       int      `json:"allowInsecure,omitempty"`
	Alpn                []string `json:"alpn,omitempty"`
	Sni                 string   `json:"sni,omitempty"`
	Host                string   `json:"host,omitempty"`
	Headers             string   `json:"headers,omitempty"`
	MaxEarlyData        int      `json:"maxEarlyData,omitempty"`
	EarlyDataHeader     string   `json:"earlyDataHeader,omitempty"`
	HttpUpgrade         int      `json:"httpUpgrade,omitempty"`
	HttpUpgradeFastOpen int      `json:"httpUpgradeFastOpen,omitempty"`
	ServiceName         string   `json:"serviceName,omitempty"`
	Mode                string   `json:"mode,omitempty"`
	GrpcUserAgent       string   `json:"grpcUserAgent,omitempty"`
	PingInterval        int      `json:"pingInterval,omitempty"`
	MaxConnections      int      `json:"maxConnections,omitempty"`
	MinStreams          int      `json:"minStreams,omitempty"`
	MaxStreams          int      `json:"maxStreams,omitempty"`
	Flow                string   `json:"flow,omitempty"`
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
		Host:   formatURLHostPort(t.Hostname, utils.GetPortString(t.Port)),
	}
	q := u.Query()
	isHTTPUpgrade := t.Query.Type == "httpupgrade" || (t.Query.Type == "ws" && t.Query.HttpUpgrade == 1 && t.Query.HttpUpgradeFastOpen == 1)
	isGRPC := t.Query.Type == "grpc"
	transportType := t.Query.Type
	if isHTTPUpgrade {
		transportType = "httpupgrade"
	}
	q.Set("peer", t.Query.Peer)
	q.Set("allowInsecure", fmt.Sprintf("%d", t.Query.AllowInsecure))
	q.Set("sni", t.Query.Sni)
	q.Set("type", transportType)
	if isHTTPUpgrade {
		q.Set("path", trojanHTTPUpgradePath(t.Query.Path, t.Query.MaxEarlyData))
	} else {
		q.Set("path", t.Query.Path)
	}
	q.Set("security", t.Query.Security)
	q.Set("fp", t.Query.Fp)
	q.Set("pcs", t.Query.Fingerprint)
	// alpn 参数支持
	if len(t.Query.Alpn) > 0 {
		q.Set("alpn", strings.Join(t.Query.Alpn, ","))
	}
	q.Set("host", t.Query.Host)
	q.Set("headers", t.Query.Headers)
	if t.Query.MaxEarlyData != 0 && !isHTTPUpgrade {
		q.Set("ed", strconv.Itoa(t.Query.MaxEarlyData))
	}
	q.Set("eh", t.Query.EarlyDataHeader)
	if t.Query.HttpUpgrade == 1 && !isHTTPUpgrade {
		q.Set("httpUpgrade", "1")
	}
	if t.Query.HttpUpgradeFastOpen == 1 && !isHTTPUpgrade {
		q.Set("httpUpgradeFastOpen", "1")
	}
	if isGRPC {
		q.Set("serviceName", t.Query.ServiceName)
		if t.Query.ServiceName != "" {
			mode := t.Query.Mode
			if mode == "" {
				mode = "gun"
			}
			q.Set("mode", mode)
		}
		q.Set("grpcUserAgent", t.Query.GrpcUserAgent)
		if t.Query.PingInterval != 0 {
			q.Set("pingInterval", strconv.Itoa(t.Query.PingInterval))
		}
		if t.Query.MaxStreams != 0 {
			q.Set("maxStreams", strconv.Itoa(t.Query.MaxStreams))
		} else {
			if t.Query.MaxConnections != 0 {
				q.Set("maxConnections", strconv.Itoa(t.Query.MaxConnections))
			}
			if t.Query.MinStreams != 0 {
				q.Set("minStreams", strconv.Itoa(t.Query.MinStreams))
			}
		}
	} else {
		q.Set("serviceName", t.Query.ServiceName)
		q.Set("mode", t.Query.Mode)
		q.Set("grpcUserAgent", t.Query.GrpcUserAgent)
		if t.Query.PingInterval != 0 {
			q.Set("pingInterval", strconv.Itoa(t.Query.PingInterval))
		}
		if t.Query.MaxConnections != 0 {
			q.Set("maxConnections", strconv.Itoa(t.Query.MaxConnections))
		}
		if t.Query.MinStreams != 0 {
			q.Set("minStreams", strconv.Itoa(t.Query.MinStreams))
		}
		if t.Query.MaxStreams != 0 {
			q.Set("maxStreams", strconv.Itoa(t.Query.MaxStreams))
		}
	}
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
	types := normalizeTrojanNetwork(u.Query().Get("type"))
	path := u.Query().Get("path")
	security := u.Query().Get("security")
	fp := u.Query().Get("fp")
	fingerprint := sanitizeCertificateFingerprint(u.Query().Get("pcs"))
	alpns := u.Query().Get("alpn")
	alpn := strings.Split(alpns, ",")
	if alpns == "" {
		alpn = nil
	}
	host := u.Query().Get("host")
	headers := u.Query().Get("headers")
	maxEarlyData := trojanURLInt(u.Query().Get("ed"))
	earlyDataHeader := u.Query().Get("eh")
	httpUpgrade := trojanURLBool(u.Query().Get("httpUpgrade"))
	httpUpgradeFastOpen := trojanURLBool(u.Query().Get("httpUpgradeFastOpen"))
	serviceName := u.Query().Get("serviceName")
	mode := u.Query().Get("mode")
	if types == "httpupgrade" {
		if pathEarlyData := trojanHTTPUpgradeEarlyData(path); pathEarlyData != 0 {
			maxEarlyData = pathEarlyData
		}
		path = trojanHTTPUpgradePathWithoutEarlyData(path)
		httpUpgrade = 1
		httpUpgradeFastOpen = 1
	}
	grpcUserAgent := u.Query().Get("grpcUserAgent")
	pingInterval := trojanURLInt(u.Query().Get("pingInterval"))
	maxConnections := trojanURLInt(u.Query().Get("maxConnections"))
	minStreams := trojanURLInt(u.Query().Get("minStreams"))
	maxStreams := trojanURLInt(u.Query().Get("maxStreams"))
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
		fmt.Println("fingerprint:", fingerprint)
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
			Peer:                peer,
			Type:                types,
			Path:                path,
			Security:            security,
			Fp:                  fp,
			Fingerprint:         fingerprint,
			AllowInsecure:       insecureVal,
			Alpn:                alpn,
			Sni:                 sni,
			Host:                host,
			Headers:             headers,
			MaxEarlyData:        maxEarlyData,
			EarlyDataHeader:     earlyDataHeader,
			HttpUpgrade:         httpUpgrade,
			HttpUpgradeFastOpen: httpUpgradeFastOpen,
			ServiceName:         serviceName,
			Mode:                mode,
			GrpcUserAgent:       grpcUserAgent,
			PingInterval:        pingInterval,
			MaxConnections:      maxConnections,
			MinStreams:          minStreams,
			MaxStreams:          maxStreams,
			Flow:                flow,
		},
		Name: name,
		Type: "trojan",
	}, nil
}

// ConvertProxyToTrojan 将 Proxy 结构体转换为 Trojan 结构体
// 用于从 Clash 格式的代理配置生成 Trojan 链接
func ConvertProxyToTrojan(proxy Proxy) Trojan {
	query := TrojanQuery{
		Sni:         proxy.Sni,
		Type:        normalizeTrojanNetwork(proxy.Network),
		Fp:          proxy.Client_fingerprint,
		Fingerprint: sanitizeCertificateFingerprint(proxy.Fingerprint),
		Flow:        proxy.Flow,
		Alpn:        proxy.Alpn,
		Peer:        proxy.Peer,
	}
	trojan := Trojan{
		Password: proxy.Password,
		Hostname: proxy.Server,
		Port:     int(proxy.Port),
		Name:     proxy.Name,
		Type:     "trojan",
		Query:    query,
	}

	// 处理跳过证书验证
	if proxy.Skip_cert_verify {
		trojan.Query.AllowInsecure = 1
	}

	switch proxy.Network {
	case "ws":
		populateTrojanWSQuery(&trojan.Query, proxy.Ws_opts)
		if trojan.Query.HttpUpgrade == 1 && trojan.Query.HttpUpgradeFastOpen == 1 {
			trojan.Query.Type = "httpupgrade"
		}
	case "grpc":
		populateTrojanGRPCQuery(&trojan.Query, proxy.Grpc_opts)
		trojan.Query.Mode = "gun"
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
	var wsOpts, grpcOpts map[string]any
	switch trojan.Query.Type {
	case "ws", "httpupgrade":
		wsOpts = buildTrojanWSOpts(trojan.Query)
	case "grpc":
		grpcOpts = buildTrojanGRPCOpts(trojan.Query)
	}
	skipCert := config.Cert || trojan.Query.AllowInsecure == 1
	return Proxy{Name: trojan.Name, Type: "trojan", Server: trojan.Hostname, Port: FlexPort(utils.GetPortInt(trojan.Port)), Password: trojan.Password, Client_fingerprint: trojan.Query.Fp, Fingerprint: trojan.Query.Fingerprint, Sni: trojan.Query.Sni, Network: trojanClashNetwork(trojan.Query.Type), Flow: trojan.Query.Flow, Alpn: trojan.Query.Alpn, Ws_opts: wsOpts, Grpc_opts: grpcOpts, Udp: config.Udp, Skip_cert_verify: skipCert, Dialer_proxy: link.DialerProxyName}, nil
}

func buildTrojanWSOpts(query TrojanQuery) map[string]any {
	opts := map[string]any{}
	if query.Path != "" {
		opts["path"] = query.Path
	}
	if headers := trojanWSHeaders(query.Headers, query.Host); len(headers) > 0 {
		opts["headers"] = headers
	}
	if query.MaxEarlyData != 0 {
		opts["max-early-data"] = query.MaxEarlyData
	}
	if query.EarlyDataHeader != "" {
		opts["early-data-header-name"] = query.EarlyDataHeader
	}
	if query.HttpUpgrade == 1 || query.Type == "httpupgrade" {
		opts["v2ray-http-upgrade"] = true
	}
	if query.HttpUpgradeFastOpen == 1 || query.Type == "httpupgrade" {
		opts["v2ray-http-upgrade-fast-open"] = true
	}
	return opts
}

func buildTrojanGRPCOpts(query TrojanQuery) map[string]any {
	opts := map[string]any{}
	if query.ServiceName != "" {
		opts["grpc-service-name"] = query.ServiceName
	}
	if query.GrpcUserAgent != "" {
		opts["grpc-user-agent"] = query.GrpcUserAgent
	}
	if query.PingInterval != 0 {
		opts["ping-interval"] = query.PingInterval
	}
	if query.MaxStreams != 0 {
		opts["max-streams"] = query.MaxStreams
		return opts
	}
	if query.MaxConnections != 0 {
		opts["max-connections"] = query.MaxConnections
	}
	if query.MinStreams != 0 {
		opts["min-streams"] = query.MinStreams
	}
	return opts
}

func populateTrojanWSQuery(query *TrojanQuery, opts map[string]any) {
	if query == nil || len(opts) == 0 {
		return
	}
	if path, ok := opts["path"].(string); ok {
		query.Path = path
	}
	headers := trojanHeadersFromValue(opts["headers"])
	query.Host, headers = trojanHostAndExtraHeaders(headers)
	if len(headers) > 0 {
		if encoded, err := json.Marshal(headers); err == nil {
			query.Headers = string(encoded)
		}
	}
	query.MaxEarlyData = trojanOptionInt(opts["max-early-data"])
	if header, ok := opts["early-data-header-name"].(string); ok {
		query.EarlyDataHeader = header
	}
	if trojanOptionBool(opts["v2ray-http-upgrade"]) {
		query.HttpUpgrade = 1
	}
	if trojanOptionBool(opts["v2ray-http-upgrade-fast-open"]) {
		query.HttpUpgradeFastOpen = 1
	}
}

func populateTrojanGRPCQuery(query *TrojanQuery, opts map[string]any) {
	if query == nil || len(opts) == 0 {
		return
	}
	if value, ok := opts["grpc-service-name"].(string); ok {
		query.ServiceName = value
	}
	if value, ok := opts["grpc-user-agent"].(string); ok {
		query.GrpcUserAgent = value
	}
	query.PingInterval = trojanOptionInt(opts["ping-interval"])
	query.MaxConnections = trojanOptionInt(opts["max-connections"])
	query.MinStreams = trojanOptionInt(opts["min-streams"])
	query.MaxStreams = trojanOptionInt(opts["max-streams"])
	if query.MaxStreams != 0 {
		query.MaxConnections = 0
		query.MinStreams = 0
	}
}

func normalizeTrojanNetwork(network string) string {
	switch network {
	case "", "tcp", "ws", "httpupgrade", "grpc":
		return network
	default:
		return "tcp"
	}
}

func trojanClashNetwork(network string) string {
	if network == "httpupgrade" {
		return "ws"
	}
	return network
}

func trojanHTTPUpgradeEarlyData(path string) int {
	parsed, err := url.Parse(path)
	if err != nil {
		return 0
	}
	return trojanURLInt(parsed.Query().Get("ed"))
}

func trojanHTTPUpgradePathWithoutEarlyData(path string) string {
	parsed, err := url.Parse(path)
	if err != nil {
		return path
	}
	query := parsed.Query()
	query.Del("ed")
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func trojanHTTPUpgradePath(path string, maxEarlyData int) string {
	parsed, err := url.Parse(path)
	if err != nil {
		return path
	}
	if maxEarlyData != 0 {
		query := parsed.Query()
		query.Set("ed", strconv.Itoa(maxEarlyData))
		parsed.RawQuery = query.Encode()
	}
	return parsed.String()
}

func trojanWSHeaders(encoded, host string) map[string]any {
	headers := map[string]any{}
	if encoded != "" {
		if err := json.Unmarshal([]byte(encoded), &headers); err != nil || headers == nil {
			headers = map[string]any{}
		}
	}
	if host != "" {
		for key := range headers {
			if strings.EqualFold(key, "host") {
				delete(headers, key)
			}
		}
		headers["Host"] = host
	}
	return headers
}

func trojanHeadersFromValue(value any) map[string]any {
	switch headers := value.(type) {
	case map[string]any:
		return headers
	case map[string]string:
		converted := make(map[string]any, len(headers))
		for key, header := range headers {
			converted[key] = header
		}
		return converted
	default:
		return nil
	}
}

func trojanHostAndExtraHeaders(headers map[string]any) (string, map[string]any) {
	if len(headers) == 0 {
		return "", nil
	}
	host := ""
	extra := make(map[string]any, len(headers))
	for key, value := range headers {
		if strings.EqualFold(key, "host") {
			if header, ok := value.(string); ok && header != "" && host == "" {
				host = header
				continue
			}
		}
		extra[key] = value
	}
	return host, extra
}

func trojanURLInt(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func trojanURLBool(value string) int {
	if value == "1" || strings.EqualFold(value, "true") {
		return 1
	}
	return 0
}

func trojanOptionInt(value any) int {
	maxInt := int64(^uint(0) >> 1)
	minInt := -maxInt - 1
	switch number := value.(type) {
	case int:
		return number
	case int8:
		return int(number)
	case int16:
		return int(number)
	case int32:
		return int(number)
	case int64:
		if number >= minInt && number <= maxInt {
			return int(number)
		}
	case uint:
		if uint64(number) <= uint64(maxInt) {
			return int(number)
		}
	case uint8:
		return int(number)
	case uint16:
		return int(number)
	case uint32:
		if uint64(number) <= uint64(maxInt) {
			return int(number)
		}
	case uint64:
		if number <= uint64(maxInt) {
			return int(number)
		}
	case float32:
		if math.Trunc(float64(number)) == float64(number) && float64(number) >= float64(minInt) && float64(number) < float64(maxInt) {
			return int(number)
		}
	case float64:
		if math.Trunc(number) == number && number >= float64(minInt) && number < float64(maxInt) {
			return int(number)
		}
	case string:
		return trojanURLInt(strings.TrimSpace(number))
	case json.Number:
		return trojanURLInt(number.String())
	}
	return 0
}

func trojanOptionBool(value any) bool {
	switch option := value.(type) {
	case bool:
		return option
	case string:
		return option == "1" || strings.EqualFold(option, "true")
	default:
		return trojanOptionInt(value) != 0
	}
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
