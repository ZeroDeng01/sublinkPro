package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"sublink/node/protocol"
	"sublink/utils"
)

// Xray JSON 订阅解析。
//
// 部分面板（Remnawave 家族、Happ AI Mode 等）不返回 Clash YAML 或分享链接列表，
// 而是返回 Xray 客户端配置。常见有两种形式：
//
//	{"outbounds": [...]}                                    单个配置
//	[{"remarks": "Германия", "outbounds": [...]}, ...]      多个配置组成的数组
//
// 解析结果统一转换为 protocol.Proxy，后续的过滤、重命名、去重和入库流程保持不变。

// xrayGenericOutboundTags 是没有实际节点含义的 outbound tag，遇到时回退到 address:port 作为节点名。
var xrayGenericOutboundTags = map[string]bool{
	"proxy":    true,
	"out":      true,
	"outbound": true,
	"node":     true,
	"default":  true,
}

// xrayDefaultRealityFingerprint 是 reality 节点缺失 uTLS 指纹时的兼容默认值。
const xrayDefaultRealityFingerprint = "chrome"

type xrayProfile struct {
	Remarks   string         `json:"remarks"`
	Outbounds []xrayOutbound `json:"outbounds"`
}

type xrayOutbound struct {
	Tag            string              `json:"tag"`
	Protocol       string              `json:"protocol"`
	Settings       json.RawMessage     `json:"settings"`
	StreamSettings *xrayStreamSettings `json:"streamSettings"`
}

type xrayStreamSettings struct {
	Network             string                   `json:"network"`
	Security            string                   `json:"security"`
	TLSSettings         *xrayTLSSettings         `json:"tlsSettings"`
	RealitySettings     *xrayRealitySettings     `json:"realitySettings"`
	WSSettings          *xrayWSSettings          `json:"wsSettings"`
	GRPCSettings        *xrayGRPCSettings        `json:"grpcSettings"`
	XHTTPSettings       *xrayXHTTPSettings       `json:"xhttpSettings"`
	SplitHTTPSettings   *xrayXHTTPSettings       `json:"splithttpSettings"`
	HTTPSettings        *xrayHTTPSettings2       `json:"httpSettings"`
	HTTPUpgradeSettings *xrayHTTPUpgradeSettings `json:"httpupgradeSettings"`
	TCPSettings         *xrayTCPSettings         `json:"tcpSettings"`
}

// xhttpSettings 优先取新命名，兼容旧的 splithttpSettings。
func (s *xrayStreamSettings) xhttp() *xrayXHTTPSettings {
	if s == nil {
		return nil
	}
	if s.XHTTPSettings != nil {
		return s.XHTTPSettings
	}
	return s.SplitHTTPSettings
}

type xrayTLSSettings struct {
	ServerName    string   `json:"serverName"`
	AllowInsecure bool     `json:"allowInsecure"`
	Alpn          []string `json:"alpn"`
	Fingerprint   string   `json:"fingerprint"`
}

type xrayRealitySettings struct {
	ServerName  string   `json:"serverName"`
	ServerNames []string `json:"serverNames"`
	PublicKey   string   `json:"publicKey"`
	ShortID     string   `json:"shortId"`
	Fingerprint string   `json:"fingerprint"`
}

type xrayWSSettings struct {
	Path                string         `json:"path"`
	Host                string         `json:"host"`
	Headers             map[string]any `json:"headers"`
	MaxEarlyData        int            `json:"maxEarlyData"`
	EarlyDataHeaderName string         `json:"earlyDataHeaderName"`
}

type xrayGRPCSettings struct {
	ServiceName string `json:"serviceName"`
	MultiMode   bool   `json:"multiMode"`
}

// xrayXHTTPSettings 同时容纳标准 Xray 写法（扩展参数放在 extra 内）
// 和 Happ 等客户端的扁平写法（xPaddingBytes 等直接挂在 xhttpSettings 下）。
type xrayXHTTPSettings struct {
	Path             string         `json:"path"`
	Host             string         `json:"host"`
	Mode             string         `json:"mode"`
	Headers          map[string]any `json:"headers"`
	Extra            map[string]any `json:"extra"`
	XPaddingBytes    any            `json:"xPaddingBytes"`
	NoGRPCHeader     bool           `json:"noGRPCHeader"`
	DownloadSettings map[string]any `json:"downloadSettings"`
}

// xrayHTTPSettings2 是 Xray 的 HTTP/2 传输层配置（httpSettings）。
type xrayHTTPSettings2 struct {
	Host []string `json:"host"`
	Path string   `json:"path"`
}

type xrayHTTPUpgradeSettings struct {
	Host string `json:"host"`
	Path string `json:"path"`
}

type xrayTCPSettings struct {
	Header struct {
		Type    string `json:"type"`
		Request struct {
			Path    []string       `json:"path"`
			Headers map[string]any `json:"headers"`
		} `json:"request"`
	} `json:"header"`
}

type xrayOutboundSettings struct {
	Vnext   []xrayVnextServer    `json:"vnext"`
	Servers []xrayServerSettings `json:"servers"`
}

type xrayVnextServer struct {
	Address string          `json:"address"`
	Port    any             `json:"port"`
	Users   []xrayVnextUser `json:"users"`
}

type xrayVnextUser struct {
	ID         string `json:"id"`
	AlterID    any    `json:"alterId"`
	Security   string `json:"security"`
	Flow       string `json:"flow"`
	Encryption string `json:"encryption"`
}

type xrayServerSettings struct {
	Address  string           `json:"address"`
	Port     any              `json:"port"`
	Password string           `json:"password"`
	Method   string           `json:"method"`
	Flow     string           `json:"flow"`
	Users    []xrayServerUser `json:"users"`
}

type xrayServerUser struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

// parseXraySubscriptionProxies 尝试把 Xray JSON 订阅解析为通用 Proxy 列表。
// 非 JSON 内容、解析失败或没有可用出站时返回 nil，调用方可以继续尝试其他格式。
func parseXraySubscriptionProxies(data []byte) []protocol.Proxy {
	profiles, ok := decodeXrayProfiles(data)
	if !ok || len(profiles) == 0 {
		return nil
	}

	// 多配置订阅里各配置通常是同一批节点的不同身份，名称完全重复会让 Clash 拒绝整份配置。
	multiProfile := len(profiles) > 1
	proxies := make([]protocol.Proxy, 0, len(profiles))
	seen := make(map[string]struct{})
	skipped := make(map[string]int)
	for _, profile := range profiles {
		namePrefix := ""
		if multiProfile {
			namePrefix = strings.TrimSpace(profile.Remarks)
		}
		for _, outbound := range profile.Outbounds {
			proxy, ok := xrayOutboundToProxy(outbound, namePrefix)
			if !ok {
				if name := strings.ToLower(strings.TrimSpace(outbound.Protocol)); name != "" {
					skipped[name]++
				}
				continue
			}
			if hash := protocol.GenerateProxyContentHash(proxy); hash != "" {
				if _, exists := seen[hash]; exists {
					continue
				}
				seen[hash] = struct{}{}
			}
			proxies = append(proxies, proxy)
		}
	}

	if len(proxies) > 0 {
		utils.Info("Xray JSON subscription parsed: %d configs, %d nodes", len(profiles), len(proxies))
	}
	if len(skipped) > 0 {
		names := make([]string, 0, len(skipped))
		for name, count := range skipped {
			names = append(names, fmt.Sprintf("%s×%d", name, count))
		}
		sort.Strings(names)
		utils.Debug("Xray JSON subscription skipped an unsupported outbound: %s", strings.Join(names, ", "))
	}
	return proxies
}

// decodeXrayProfiles 识别单配置或配置数组两种 JSON 结构。
func decodeXrayProfiles(data []byte) ([]xrayProfile, bool) {
	trimmed := bytes.TrimSpace(bytes.TrimPrefix(data, []byte("\xEF\xBB\xBF")))
	if len(trimmed) == 0 {
		return nil, false
	}

	switch trimmed[0] {
	case '[':
		var profiles []xrayProfile
		if err := json.Unmarshal(trimmed, &profiles); err == nil {
			return profiles, true
		}
		// 流式响应可能在传输中被截断，尽量保留已经完整到达的配置对象。
		if salvaged := decodeCompleteXrayProfiles(trimmed); len(salvaged) > 0 {
			utils.Warn("Xray JSON subscription response was incomplete; recovered %d complete configs", len(salvaged))
			return salvaged, true
		}
		// 首个配置本身也未接收完整时，退一步保留其中已经完整的出站。
		if outbounds := salvageXrayOutbounds(trimmed); len(outbounds) > 0 {
			utils.Warn("Xray JSON subscription response was incomplete; recovered %d outbounds from the first config", len(outbounds))
			return []xrayProfile{{Outbounds: outbounds}}, true
		}
		return nil, false
	case '{':
		var profile xrayProfile
		if err := json.Unmarshal(trimmed, &profile); err == nil {
			return []xrayProfile{profile}, true
		}
		if outbounds := salvageXrayOutbounds(trimmed); len(outbounds) > 0 {
			utils.Warn("Xray JSON subscription response was incomplete; recovered %d outbounds", len(outbounds))
			return []xrayProfile{{Outbounds: outbounds}}, true
		}
		return nil, false
	default:
		return nil, false
	}
}

// salvageXrayOutbounds 从被截断的配置中恢复 outbounds 数组里已经完整的出站对象。
// 流式响应通常只截断配置本身，其中绝大多数出站对象是完整的。
func salvageXrayOutbounds(data []byte) []xrayOutbound {
	keyIndex := bytes.Index(data, []byte(`"outbounds"`))
	if keyIndex < 0 {
		return nil
	}
	arrayOffset := bytes.IndexByte(data[keyIndex:], '[')
	if arrayOffset < 0 {
		return nil
	}
	return decodeCompleteXrayOutbounds(data[keyIndex+arrayOffset:])
}

// scanCompleteJSONObjects 按大括号配对扫描已完整到达的顶层 JSON 对象，
// 字符串和转义序列会被跳过，被截断的尾部对象不会回调。
func scanCompleteJSONObjects(data []byte, onObject func([]byte)) {
	depth := 0
	inString := false
	escaped := false
	start := -1
	for i := 0; i < len(data); i++ {
		ch := data[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case ch == '\\':
				escaped = true
			case ch == '"':
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
			if depth == 1 {
				start = i
			}
		case '}':
			if depth == 0 {
				continue
			}
			depth--
			if depth == 0 && start >= 0 {
				onObject(data[start : i+1])
				start = -1
			}
		}
	}
}

// decodeCompleteXrayProfiles 解码扫描到的完整配置对象。
func decodeCompleteXrayProfiles(data []byte) []xrayProfile {
	profiles := make([]xrayProfile, 0)
	scanCompleteJSONObjects(data, func(object []byte) {
		var profile xrayProfile
		if err := json.Unmarshal(object, &profile); err == nil {
			profiles = append(profiles, profile)
		}
	})
	return profiles
}

// decodeCompleteXrayOutbounds 解码扫描到的完整出站对象，没有 protocol 的片段直接丢弃。
func decodeCompleteXrayOutbounds(data []byte) []xrayOutbound {
	outbounds := make([]xrayOutbound, 0)
	scanCompleteJSONObjects(data, func(object []byte) {
		var outbound xrayOutbound
		if err := json.Unmarshal(object, &outbound); err == nil && strings.TrimSpace(outbound.Protocol) != "" {
			outbounds = append(outbounds, outbound)
		}
	})
	return outbounds
}

// xrayOutboundToProxy 把单个 Xray outbound 转换为通用 Proxy。
// namePrefix 非空时会加到节点名前，用于区分多配置订阅里不同身份的同名节点。
// 返回 false 表示该出站不是可导入的代理（freedom、blackhole、kcp 传输等）。
func xrayOutboundToProxy(outbound xrayOutbound, namePrefix string) (protocol.Proxy, bool) {
	protocolName := strings.ToLower(strings.TrimSpace(outbound.Protocol))
	switch protocolName {
	case "vless", "vmess", "trojan", "shadowsocks", "ss", "socks", "http":
	default:
		return protocol.Proxy{}, false
	}

	stream := outbound.StreamSettings
	if stream == nil {
		stream = &xrayStreamSettings{}
	}

	var settings xrayOutboundSettings
	if len(outbound.Settings) > 0 {
		if err := json.Unmarshal(outbound.Settings, &settings); err != nil {
			return protocol.Proxy{}, false
		}
	}

	proxy := protocol.Proxy{
		Name: xrayNodeName(outbound.Tag),
		Udp:  true,
	}

	switch protocolName {
	case "vless", "vmess":
		if len(settings.Vnext) == 0 || len(settings.Vnext[0].Users) == 0 {
			return protocol.Proxy{}, false
		}
		vnext := settings.Vnext[0]
		user := vnext.Users[0]
		proxy.Server = strings.TrimSpace(vnext.Address)
		proxy.Port = protocol.FlexPort(xrayIntValue(vnext.Port))
		proxy.Uuid = strings.TrimSpace(user.ID)
		if protocolName == "vless" {
			proxy.Type = "vless"
			proxy.Encryption = strings.TrimSpace(user.Encryption)
			proxy.Flow = strings.TrimSpace(user.Flow)
		} else {
			proxy.Type = "vmess"
			proxy.Cipher = strings.TrimSpace(user.Security)
			if proxy.Cipher == "" {
				proxy.Cipher = "auto"
			}
			proxy.AlterId = strconv.Itoa(xrayIntValue(user.AlterID))
		}
	case "trojan", "shadowsocks", "ss", "socks", "http":
		if len(settings.Servers) == 0 {
			return protocol.Proxy{}, false
		}
		server := settings.Servers[0]
		proxy.Server = strings.TrimSpace(server.Address)
		proxy.Port = protocol.FlexPort(xrayIntValue(server.Port))
		proxy.Password = server.Password
		switch protocolName {
		case "trojan":
			proxy.Type = "trojan"
			proxy.Flow = strings.TrimSpace(server.Flow)
		case "shadowsocks", "ss":
			proxy.Type = "ss"
			proxy.Cipher = strings.TrimSpace(server.Method)
		case "socks":
			proxy.Type = "socks5"
			if len(server.Users) > 0 {
				proxy.Username = server.Users[0].User
				proxy.Password = server.Users[0].Pass
			}
		case "http":
			proxy.Type = "http"
			if len(server.Users) > 0 {
				proxy.Username = server.Users[0].User
				proxy.Password = server.Users[0].Pass
			}
		}
	}

	if proxy.Server == "" || proxy.Port <= 0 {
		return protocol.Proxy{}, false
	}
	if proxy.Name == "" {
		proxy.Name = fmt.Sprintf("%s:%d", proxy.Server, int(proxy.Port))
	}
	if namePrefix != "" {
		proxy.Name = namePrefix + " - " + proxy.Name
	}

	applyXrayTLSSettings(&proxy, stream)
	if !applyXrayTransport(&proxy, stream) {
		utils.Debug("Xray node [%s] uses a transport that is not supported yet; skipped (network=%s)", proxy.Name, stream.Network)
		return protocol.Proxy{}, false
	}
	return proxy, true
}

// xrayNodeName 返回可用作节点名的 tag，通用占位 tag 视为空。
func xrayNodeName(tag string) string {
	name := strings.TrimSpace(tag)
	if xrayGenericOutboundTags[strings.ToLower(name)] {
		return ""
	}
	return name
}

// applyXrayTLSSettings 把 tlsSettings / realitySettings 映射到 Clash 字段。
func applyXrayTLSSettings(proxy *protocol.Proxy, stream *xrayStreamSettings) {
	security := strings.ToLower(strings.TrimSpace(stream.Security))
	if security == "" || security == "none" {
		return
	}

	// trojan 本身始终走 TLS，mihomo 侧没有独立 tls 字段，只有 vless/vmess 需要显式开启。
	if proxy.Type == "vless" || proxy.Type == "vmess" {
		proxy.Tls = true
	}

	if security == "reality" {
		reality := stream.RealitySettings
		if reality == nil {
			return
		}
		setXrayServerName(proxy, firstXrayNonEmpty(reality.ServerName, firstXraySliceValue(reality.ServerNames)))
		if fingerprint := strings.TrimSpace(reality.Fingerprint); fingerprint != "" {
			proxy.Client_fingerprint = fingerprint
		} else if proxy.Client_fingerprint == "" {
			proxy.Client_fingerprint = xrayDefaultRealityFingerprint
		}
		opts := map[string]any{}
		if publicKey := strings.TrimSpace(reality.PublicKey); publicKey != "" {
			opts["public-key"] = publicKey
		}
		if shortID := strings.TrimSpace(reality.ShortID); shortID != "" {
			opts["short-id"] = shortID
		}
		// public-key 缺失时 reality-opts 会让 mihomo 拒绝整份配置，直接省略。
		if _, ok := opts["public-key"]; ok {
			proxy.Reality_opts = opts
		}
		return
	}

	tls := stream.TLSSettings
	if tls == nil {
		return
	}
	setXrayServerName(proxy, strings.TrimSpace(tls.ServerName))
	proxy.Skip_cert_verify = tls.AllowInsecure
	if len(tls.Alpn) > 0 {
		proxy.Alpn = append([]string(nil), tls.Alpn...)
	}
	if fingerprint := strings.TrimSpace(tls.Fingerprint); fingerprint != "" {
		proxy.Client_fingerprint = fingerprint
	}
}

// setXrayServerName 写入各协议实际读取的 SNI 字段：trojan 用 sni，其余用 servername。
func setXrayServerName(proxy *protocol.Proxy, serverName string) {
	if serverName == "" {
		return
	}
	if proxy.Type == "trojan" {
		proxy.Sni = serverName
		return
	}
	proxy.Servername = serverName
}

// applyXrayTransport 把传输层配置映射到 Clash 字段，返回 false 表示无法承载。
func applyXrayTransport(proxy *protocol.Proxy, stream *xrayStreamSettings) bool {
	network := normalizeXrayNetwork(stream)

	// ss/socks5/http 的分享链接没有 ws/grpc 等传输层字段，只能保留直连 TCP。
	if proxy.Type == "ss" || proxy.Type == "socks5" || proxy.Type == "http" {
		return network == "tcp"
	}

	switch network {
	case "tcp":
		if stream.TCPSettings != nil && strings.EqualFold(strings.TrimSpace(stream.TCPSettings.Header.Type), "http") {
			if proxy.Type != "vless" {
				return false
			}
			proxy.Network = "http"
			proxy.Http_opts = buildXrayTCPHTTPOpts(stream.TCPSettings)
			return true
		}
		proxy.Network = "tcp"
		return true
	case "ws":
		proxy.Network = "ws"
		proxy.Ws_opts = buildXrayWSOpts(stream.WSSettings)
		return true
	case "httpupgrade":
		proxy.Network = "ws"
		proxy.Ws_opts = buildXrayHTTPUpgradeOpts(stream.HTTPUpgradeSettings)
		return true
	case "grpc":
		proxy.Network = "grpc"
		proxy.Grpc_opts = buildXrayGRPCOpts(stream.GRPCSettings)
		return true
	case "h2":
		if proxy.Type != "vless" {
			return false
		}
		proxy.Network = "h2"
		proxy.H2_opts = buildXrayH2Opts(stream.HTTPSettings)
		return true
	case "xhttp":
		if proxy.Type != "vless" {
			return false
		}
		proxy.Network = "xhttp"
		proxy.XHTTP_opts = buildXrayXHTTPOpts(stream.xhttp())
		return true
	default:
		return false
	}
}

// normalizeXrayNetwork 把 Xray 的传输层名称收敛到 Clash 侧使用的命名。
func normalizeXrayNetwork(stream *xrayStreamSettings) string {
	if stream == nil {
		return "tcp"
	}
	switch strings.ToLower(strings.TrimSpace(stream.Network)) {
	case "", "tcp", "raw":
		return "tcp"
	case "h2", "http":
		return "h2"
	case "splithttp":
		return "xhttp"
	default:
		return strings.ToLower(strings.TrimSpace(stream.Network))
	}
}

func buildXrayTCPHTTPOpts(settings *xrayTCPSettings) map[string]any {
	opts := map[string]any{}
	paths := make([]string, 0, len(settings.Header.Request.Path))
	for _, path := range settings.Header.Request.Path {
		if trimmed := strings.TrimSpace(path); trimmed != "" {
			paths = append(paths, trimmed)
		}
	}
	if len(paths) > 0 {
		opts["path"] = paths
	}
	if hosts := xrayHeaderValues(settings.Header.Request.Headers["Host"]); len(hosts) > 0 {
		opts["headers"] = map[string]any{"Host": hosts}
	}
	if len(opts) == 0 {
		return nil
	}
	return opts
}

func buildXrayWSOpts(settings *xrayWSSettings) map[string]any {
	if settings == nil {
		return nil
	}
	opts := map[string]any{}
	if path := strings.TrimSpace(settings.Path); path != "" {
		opts["path"] = path
	}

	headers := map[string]any{}
	for key, value := range settings.Headers {
		if header := xrayHeaderValue(value); header != "" {
			headers[key] = header
		}
	}
	host := strings.TrimSpace(settings.Host)
	if host == "" {
		host = xrayHostHeader(headers)
	}
	if host != "" {
		for key := range headers {
			if strings.EqualFold(key, "host") {
				delete(headers, key)
			}
		}
		headers["Host"] = host
	}
	if len(headers) > 0 {
		opts["headers"] = headers
	}

	if settings.MaxEarlyData > 0 {
		opts["max-early-data"] = settings.MaxEarlyData
	}
	if name := strings.TrimSpace(settings.EarlyDataHeaderName); name != "" {
		opts["early-data-header-name"] = name
	}

	protocol.DeleteOpts(opts)
	if len(opts) == 0 {
		return nil
	}
	return opts
}

func buildXrayHTTPUpgradeOpts(settings *xrayHTTPUpgradeSettings) map[string]any {
	opts := map[string]any{
		"v2ray-http-upgrade":           true,
		"v2ray-http-upgrade-fast-open": true,
	}
	if settings == nil {
		return opts
	}
	if path := strings.TrimSpace(settings.Path); path != "" {
		opts["path"] = path
	}
	if host := strings.TrimSpace(settings.Host); host != "" {
		opts["headers"] = map[string]any{"Host": host}
	}
	return opts
}

func buildXrayGRPCOpts(settings *xrayGRPCSettings) map[string]any {
	if settings == nil {
		return nil
	}
	opts := map[string]any{}
	if serviceName := strings.TrimSpace(settings.ServiceName); serviceName != "" {
		opts["grpc-service-name"] = serviceName
		opts["grpc-mode"] = "gun"
	}
	if settings.MultiMode {
		opts["grpc-mode"] = "multi"
	}
	protocol.DeleteOpts(opts)
	if len(opts) == 0 {
		return nil
	}
	return opts
}

func buildXrayH2Opts(settings *xrayHTTPSettings2) map[string]any {
	if settings == nil {
		return nil
	}
	opts := map[string]any{}
	hosts := make([]string, 0, len(settings.Host))
	for _, host := range settings.Host {
		if trimmed := strings.TrimSpace(host); trimmed != "" {
			hosts = append(hosts, trimmed)
		}
	}
	if len(hosts) > 0 {
		opts["host"] = hosts
	}
	if path := strings.TrimSpace(settings.Path); path != "" {
		opts["path"] = path
	}
	if len(opts) == 0 {
		return nil
	}
	return opts
}

// buildXrayXHTTPOpts 生成 mihomo xhttp-opts。
// 只有 mihomo 有对应承载字段的扩展参数会被保留，xmux 等无对应实现的参数按既有规则丢弃。
func buildXrayXHTTPOpts(settings *xrayXHTTPSettings) map[string]any {
	if settings == nil {
		return nil
	}
	opts := map[string]any{}
	if path := strings.TrimSpace(settings.Path); path != "" {
		opts["path"] = path
	}
	if host := strings.TrimSpace(settings.Host); host != "" {
		opts["host"] = host
	}
	if mode := strings.TrimSpace(settings.Mode); mode != "" {
		opts["mode"] = mode
	}
	if headers := xrayHeadersMap(settings.Headers); len(headers) > 0 {
		opts["headers"] = headers
	}
	if settings.XPaddingBytes != nil {
		opts["x-padding-bytes"] = settings.XPaddingBytes
	}
	if settings.NoGRPCHeader {
		opts["no-grpc-header"] = true
	}
	if len(settings.DownloadSettings) > 0 {
		opts["download-settings"] = settings.DownloadSettings
	}

	for key, value := range settings.Extra {
		switch key {
		case "headers":
			if headers := xrayHeadersMapValue(value); len(headers) > 0 {
				opts["headers"] = headers
			}
		case "noGRPCHeader":
			opts["no-grpc-header"] = value
		case "xPaddingBytes":
			opts["x-padding-bytes"] = value
		case "downloadSettings":
			if download, ok := value.(map[string]any); ok && len(download) > 0 {
				opts["download-settings"] = download
			}
		}
	}

	protocol.DeleteOpts(opts)
	if len(opts) == 0 {
		return nil
	}
	return opts
}

// xrayHeadersMap 把 Xray 的 headers 对象转换成 map[string]any（只保留非空字符串值）。
func xrayHeadersMap(headers map[string]any) map[string]any {
	result := map[string]any{}
	for key, value := range headers {
		if header := xrayHeaderValue(value); header != "" {
			result[key] = header
		}
	}
	return result
}

func xrayHeadersMapValue(value any) map[string]any {
	headers, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	return xrayHeadersMap(headers)
}

// xrayHeaderValue 兼容 headers 值写成字符串或字符串数组两种写法。
func xrayHeaderValue(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case []any:
		for _, item := range v {
			if header, ok := item.(string); ok {
				return strings.TrimSpace(header)
			}
		}
	case []string:
		for _, item := range v {
			return strings.TrimSpace(item)
		}
	}
	return ""
}

// xrayHeaderValues 返回 headers 中的全部非空字符串值，用于 Host 之类的多值字段。
func xrayHeaderValues(value any) []any {
	values := make([]any, 0)
	switch v := value.(type) {
	case string:
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			values = append(values, trimmed)
		}
	case []any:
		for _, item := range v {
			if header, ok := item.(string); ok {
				if trimmed := strings.TrimSpace(header); trimmed != "" {
					values = append(values, trimmed)
				}
			}
		}
	case []string:
		for _, item := range v {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				values = append(values, trimmed)
			}
		}
	}
	return values
}

func xrayHostHeader(headers map[string]any) string {
	for key, value := range headers {
		if strings.EqualFold(key, "host") {
			if host, ok := value.(string); ok {
				return host
			}
		}
	}
	return ""
}

func xrayIntValue(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0
		}
		return parsed
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return 0
		}
		return int(parsed)
	default:
		return 0
	}
}

func firstXrayNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstXraySliceValue(values []string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
