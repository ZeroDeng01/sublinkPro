package protocol

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const defaultOpenVPNPort = 1194

func init() {
	base := newProtocolSpec("openvpn", []string{"openvpn://"}, "OpenVPN", "#ea7e20", "O", OpenVPN{}, "Name", DecodeOpenVPNURL, EncodeOpenVPNURL, func(o OpenVPN) LinkIdentity {
		return buildIdentity("openvpn", o.Name, o.Server, strconv.Itoa(o.Port))
	},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic"},
		FieldMeta{Name: "Server", Label: "服务器地址", Type: "string", Group: "basic"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic"},
		FieldMeta{Name: "Proto", Label: "传输协议", Type: "string", Group: "transport", Options: []string{"udp", "tcp"}},
		FieldMeta{Name: "Dev", Label: "设备类型", Type: "string", Group: "transport", Options: []string{"tun"}, Advanced: true},
		FieldMeta{Name: "Cipher", Label: "加密算法", Type: "string", Group: "transport"},
		FieldMeta{Name: "DataCiphers", Label: "数据通道加密算法", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "DataCipherFallback", Label: "回退加密算法", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Auth", Label: "认证摘要算法", Type: "string", Group: "auth"},
		FieldMeta{Name: "CompLZO", Label: "LZO 压缩", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Username", Label: "用户名", Type: "string", Group: "auth"},
		FieldMeta{Name: "Password", Label: "密码", Type: "string", Group: "auth", Secret: true},
		FieldMeta{Name: "CA", Label: "CA 证书", Type: "string", Group: "tls", Multiline: true},
		FieldMeta{Name: "Cert", Label: "客户端证书", Type: "string", Group: "tls", Multiline: true},
		FieldMeta{Name: "Key", Label: "客户端私钥", Type: "string", Group: "tls", Secret: true, Multiline: true},
		FieldMeta{Name: "TLSAuth", Label: "TLS Auth 密钥", Type: "string", Group: "tls", Secret: true, Multiline: true, Advanced: true},
		FieldMeta{Name: "KeyDirection", Label: "密钥方向", Type: "string", Group: "tls", Options: []string{"0", "1"}, Advanced: true},
		FieldMeta{Name: "TLSCrypt", Label: "TLS Crypt 密钥", Type: "string", Group: "tls", Secret: true, Multiline: true, Advanced: true},
		FieldMeta{Name: "TLSCryptV2", Label: "TLS Crypt V2 密钥", Type: "string", Group: "tls", Secret: true, Multiline: true, Advanced: true},
		FieldMeta{Name: "PeerInfo", Label: "Peer Info", Type: "string", Group: "advanced", Advanced: true},
		FieldMeta{Name: "Ping", Label: "Ping 间隔", Type: "int", Group: "advanced", Advanced: true},
		FieldMeta{Name: "PingRestart", Label: "Ping 重连超时", Type: "int", Group: "advanced", Advanced: true},
		FieldMeta{Name: "TranWindow", Label: "密钥转换窗口", Type: "int", Group: "advanced", Advanced: true},
		FieldMeta{Name: "HandshakeTimeout", Label: "握手超时", Type: "int", Group: "advanced", Advanced: true},
		FieldMeta{Name: "MTU", Label: "MTU", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "UDP", Label: "允许 UDP", Type: "bool", Group: "transport"},
		FieldMeta{Name: "IPStack", Label: "IP Stack", Type: "string", Group: "advanced", Advanced: true},
		FieldMeta{Name: "RemoteDNSResolve", Label: "远端 DNS 解析", Type: "bool", Group: "advanced", Advanced: true},
		FieldMeta{Name: "DNS", Label: "DNS 服务器", Type: "string", Group: "advanced", Advanced: true},
		FieldMeta{Name: "TFO", Label: "TCP Fast Open", Type: "bool", Group: "advanced", Advanced: true},
		FieldMeta{Name: "MPTCP", Label: "MPTCP", Type: "bool", Group: "advanced", Advanced: true},
		FieldMeta{Name: "InterfaceName", Label: "本地网卡", Type: "string", Group: "advanced", Advanced: true},
		FieldMeta{Name: "RoutingMark", Label: "路由标记", Type: "int", Group: "advanced", Advanced: true},
		FieldMeta{Name: "IPVersion", Label: "IP 版本偏好", Type: "string", Group: "advanced", Advanced: true},
		FieldMeta{Name: "DialerProxy", Label: "前置代理", Type: "string", Group: "advanced", Advanced: true},
	).WithClientSupport(ClientClash, ClientMihomo)

	MustRegisterProtocol(newProxyProtocolSpec(base, func(link Urls, _ OutputConfig) (Proxy, error) {
		return buildOpenVPNProxy(link)
	}, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "openvpn")
	}, ConvertProxyToOpenVPN, EncodeOpenVPNURL))
}

// OpenVPN stores SublinkPro's internal editable representation of a Mihomo
// OpenVPN outbound. The openvpn:// URL is an internal round-trip format, not
// an official OpenVPN share-link standard.
type OpenVPN struct {
	Name               string
	Server             string
	Port               int
	Proto              string
	Dev                string
	Cipher             string
	DataCiphers        []string
	DataCipherFallback string
	Auth               string
	CompLZO            string
	CA                 string
	Cert               string
	Key                string
	TLSAuth            string
	KeyDirection       string
	TLSCrypt           string
	TLSCryptV2         string
	Username           string
	Password           string
	PeerInfo           map[string]string
	Ping               int
	PingRestart        int
	TranWindow         *int
	HandshakeTimeout   int
	MTU                int
	UDP                bool
	IPStack            map[string]any
	RemoteDNSResolve   bool
	DNS                []string
	TFO                bool
	MPTCP              bool
	InterfaceName      string
	RoutingMark        int
	IPVersion          string
	DialerProxy        string
}

func DecodeOpenVPNURL(s string) (OpenVPN, error) {
	u, err := url.Parse(s)
	if err != nil {
		return OpenVPN{}, fmt.Errorf("OpenVPN URL 解析失败: %w", err)
	}
	if strings.ToLower(u.Scheme) != "openvpn" {
		return OpenVPN{}, fmt.Errorf("非OpenVPN协议")
	}

	server := u.Hostname()
	if server == "" {
		return OpenVPN{}, fmt.Errorf("OpenVPN 缺少服务器地址")
	}

	port := defaultOpenVPNPort
	if rawPort := u.Port(); rawPort != "" {
		port, err = strconv.Atoi(rawPort)
		if err != nil || port <= 0 || port > 65535 {
			return OpenVPN{}, fmt.Errorf("OpenVPN 端口无效: %q", rawPort)
		}
	}

	q := u.Query()
	peerInfo := map[string]string(nil)
	if raw := q.Get("peer-info"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &peerInfo); err != nil {
			return OpenVPN{}, fmt.Errorf("OpenVPN peer-info 解析失败: %w", err)
		}
	}
	ipStack := map[string]any(nil)
	if raw := q.Get("ip-stack"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &ipStack); err != nil {
			return OpenVPN{}, fmt.Errorf("OpenVPN ip-stack 解析失败: %w", err)
		}
	}

	ping, err := openVPNQueryInt(q, "ping")
	if err != nil {
		return OpenVPN{}, err
	}
	pingRestart, err := openVPNQueryInt(q, "ping-restart")
	if err != nil {
		return OpenVPN{}, err
	}
	handshakeTimeout, err := openVPNQueryInt(q, "handshake-timeout")
	if err != nil {
		return OpenVPN{}, err
	}
	mtu, err := openVPNQueryInt(q, "mtu")
	if err != nil {
		return OpenVPN{}, err
	}
	routingMark, err := openVPNQueryInt(q, "routing-mark")
	if err != nil {
		return OpenVPN{}, err
	}
	tranWindow, err := openVPNOptionalQueryInt(q, "tran-window")
	if err != nil {
		return OpenVPN{}, err
	}

	udp, err := openVPNQueryBool(q, "udp")
	if err != nil {
		return OpenVPN{}, err
	}
	remoteDNSResolve, err := openVPNQueryBool(q, "remote-dns-resolve")
	if err != nil {
		return OpenVPN{}, err
	}
	tfo, err := openVPNQueryBool(q, "tfo")
	if err != nil {
		return OpenVPN{}, err
	}
	mptcp, err := openVPNQueryBool(q, "mptcp")
	if err != nil {
		return OpenVPN{}, err
	}

	name := u.Fragment
	if name == "" {
		name = formatURLHostPort(server, strconv.Itoa(port))
	}

	return OpenVPN{
		Name:               name,
		Server:             server,
		Port:               port,
		Proto:              q.Get("proto"),
		Dev:                q.Get("dev"),
		Cipher:             q.Get("cipher"),
		DataCiphers:        append([]string(nil), q["data-ciphers"]...),
		DataCipherFallback: q.Get("data-ciphers-fallback"),
		Auth:               q.Get("auth"),
		CompLZO:            q.Get("comp-lzo"),
		CA:                 q.Get("ca"),
		Cert:               q.Get("cert"),
		Key:                q.Get("key"),
		TLSAuth:            q.Get("tls-auth"),
		KeyDirection:       q.Get("key-direction"),
		TLSCrypt:           q.Get("tls-crypt"),
		TLSCryptV2:         q.Get("tls-crypt-v2"),
		Username:           q.Get("username"),
		Password:           q.Get("password"),
		PeerInfo:           peerInfo,
		Ping:               ping,
		PingRestart:        pingRestart,
		TranWindow:         tranWindow,
		HandshakeTimeout:   handshakeTimeout,
		MTU:                mtu,
		UDP:                udp,
		IPStack:            ipStack,
		RemoteDNSResolve:   remoteDNSResolve,
		DNS:                append([]string(nil), q["dns"]...),
		TFO:                tfo,
		MPTCP:              mptcp,
		InterfaceName:      q.Get("interface-name"),
		RoutingMark:        routingMark,
		IPVersion:          q.Get("ip-version"),
		DialerProxy:        q.Get("dialer-proxy"),
	}, nil
}

func EncodeOpenVPNURL(o OpenVPN) string {
	server := strings.TrimSpace(o.Server)
	port := o.Port
	if port <= 0 || port > 65535 {
		port = defaultOpenVPNPort
	}

	u := url.URL{
		Scheme:   "openvpn",
		Host:     formatURLHostPort(server, strconv.Itoa(port)),
		Fragment: o.Name,
	}
	q := u.Query()
	openVPNSetString(q, "proto", o.Proto)
	openVPNSetString(q, "dev", o.Dev)
	openVPNSetString(q, "cipher", o.Cipher)
	for _, cipher := range o.DataCiphers {
		if cipher = strings.TrimSpace(cipher); cipher != "" {
			q.Add("data-ciphers", cipher)
		}
	}
	openVPNSetString(q, "data-ciphers-fallback", o.DataCipherFallback)
	openVPNSetString(q, "auth", o.Auth)
	openVPNSetString(q, "comp-lzo", o.CompLZO)
	openVPNSetString(q, "ca", o.CA)
	openVPNSetString(q, "cert", o.Cert)
	openVPNSetString(q, "key", o.Key)
	openVPNSetString(q, "tls-auth", o.TLSAuth)
	openVPNSetString(q, "key-direction", o.KeyDirection)
	openVPNSetString(q, "tls-crypt", o.TLSCrypt)
	openVPNSetString(q, "tls-crypt-v2", o.TLSCryptV2)
	openVPNSetString(q, "username", o.Username)
	openVPNSetString(q, "password", o.Password)
	openVPNSetJSON(q, "peer-info", o.PeerInfo)
	openVPNSetInt(q, "ping", o.Ping)
	openVPNSetInt(q, "ping-restart", o.PingRestart)
	if o.TranWindow != nil {
		q.Set("tran-window", strconv.Itoa(*o.TranWindow))
	}
	openVPNSetInt(q, "handshake-timeout", o.HandshakeTimeout)
	openVPNSetInt(q, "mtu", o.MTU)
	openVPNSetBool(q, "udp", o.UDP)
	openVPNSetJSON(q, "ip-stack", o.IPStack)
	openVPNSetBool(q, "remote-dns-resolve", o.RemoteDNSResolve)
	for _, dns := range o.DNS {
		if dns = strings.TrimSpace(dns); dns != "" {
			q.Add("dns", dns)
		}
	}
	openVPNSetBool(q, "tfo", o.TFO)
	openVPNSetBool(q, "mptcp", o.MPTCP)
	openVPNSetString(q, "interface-name", o.InterfaceName)
	openVPNSetInt(q, "routing-mark", o.RoutingMark)
	openVPNSetString(q, "ip-version", o.IPVersion)
	openVPNSetString(q, "dialer-proxy", o.DialerProxy)
	u.RawQuery = q.Encode()

	if u.Fragment == "" {
		u.Fragment = formatURLHostPort(strings.Trim(server, "[]"), strconv.Itoa(port))
	}
	return u.String()
}

func ConvertProxyToOpenVPN(proxy Proxy) OpenVPN {
	return OpenVPN{
		Name:               proxy.Name,
		Server:             proxy.Server,
		Port:               int(proxy.Port),
		Proto:              proxy.Proto,
		Dev:                proxy.Dev,
		Cipher:             proxy.Cipher,
		DataCiphers:        append([]string(nil), proxy.Data_ciphers...),
		DataCipherFallback: proxy.Data_cipher_fallback,
		Auth:               proxy.Auth,
		CompLZO:            proxy.Comp_lzo,
		CA:                 proxy.Ca,
		Cert:               proxy.Cert,
		Key:                proxy.Key,
		TLSAuth:            proxy.Tls_auth,
		KeyDirection:       proxy.Key_direction,
		TLSCrypt:           proxy.Tls_crypt,
		TLSCryptV2:         proxy.Tls_crypt_v2,
		Username:           proxy.Username,
		Password:           proxy.Password,
		PeerInfo:           cloneOpenVPNStringMap(proxy.Peer_info),
		Ping:               proxy.Ping,
		PingRestart:        proxy.Ping_restart,
		TranWindow:         cloneOpenVPNIntPointer(proxy.Tran_window),
		HandshakeTimeout:   proxy.Handshake_timeout,
		MTU:                proxy.Mtu,
		UDP:                proxy.Udp,
		IPStack:            cloneOpenVPNAnyMap(proxy.Ip_stack),
		RemoteDNSResolve:   proxy.Remote_dns_resolve,
		DNS:                append([]string(nil), proxy.Dns...),
		TFO:                proxy.Tfo,
		MPTCP:              proxy.Mptcp,
		InterfaceName:      proxy.Interface_name,
		RoutingMark:        proxy.Routing_mark,
		IPVersion:          proxy.Ip_version,
		DialerProxy:        proxy.Dialer_proxy,
	}
}

func buildOpenVPNProxy(link Urls) (Proxy, error) {
	o, err := DecodeOpenVPNURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}

	dialerProxy := o.DialerProxy
	if link.DialerProxyName != "" {
		dialerProxy = link.DialerProxyName
	}

	return Proxy{
		Name:                 o.Name,
		Type:                 "openvpn",
		Server:               o.Server,
		Port:                 FlexPort(o.Port),
		Proto:                o.Proto,
		Dev:                  o.Dev,
		Cipher:               o.Cipher,
		Data_ciphers:         append([]string(nil), o.DataCiphers...),
		Data_cipher_fallback: o.DataCipherFallback,
		Auth:                 o.Auth,
		Comp_lzo:             o.CompLZO,
		Ca:                   o.CA,
		Cert:                 o.Cert,
		Key:                  o.Key,
		Tls_auth:             o.TLSAuth,
		Key_direction:        o.KeyDirection,
		Tls_crypt:            o.TLSCrypt,
		Tls_crypt_v2:         o.TLSCryptV2,
		Username:             o.Username,
		Password:             o.Password,
		Peer_info:            cloneOpenVPNStringMap(o.PeerInfo),
		Ping:                 o.Ping,
		Ping_restart:         o.PingRestart,
		Tran_window:          cloneOpenVPNIntPointer(o.TranWindow),
		Handshake_timeout:    o.HandshakeTimeout,
		Mtu:                  o.MTU,
		Udp:                  o.UDP,
		Ip_stack:             cloneOpenVPNAnyMap(o.IPStack),
		Remote_dns_resolve:   o.RemoteDNSResolve,
		Dns:                  append([]string(nil), o.DNS...),
		Tfo:                  o.TFO,
		Mptcp:                o.MPTCP,
		Interface_name:       o.InterfaceName,
		Routing_mark:         o.RoutingMark,
		Ip_version:           o.IPVersion,
		Dialer_proxy:         dialerProxy,
	}, nil
}

func openVPNSetString(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

func openVPNSetInt(q url.Values, key string, value int) {
	if value != 0 {
		q.Set(key, strconv.Itoa(value))
	}
}

func openVPNSetBool(q url.Values, key string, value bool) {
	if value {
		q.Set(key, strconv.FormatBool(value))
	}
}

func openVPNSetJSON(q url.Values, key string, value any) {
	if value == nil {
		return
	}
	encoded, err := json.Marshal(value)
	if err == nil && string(encoded) != "{}" && string(encoded) != "null" {
		q.Set(key, string(encoded))
	}
}

func openVPNQueryInt(q url.Values, key string) (int, error) {
	raw := q.Get(key)
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("OpenVPN %s 参数无效: %w", key, err)
	}
	return value, nil
}

func openVPNOptionalQueryInt(q url.Values, key string) (*int, error) {
	if _, exists := q[key]; !exists {
		return nil, nil
	}
	value, err := openVPNQueryInt(q, key)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func openVPNQueryBool(q url.Values, key string) (bool, error) {
	raw := q.Get(key)
	if raw == "" {
		return false, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("OpenVPN %s 参数无效: %w", key, err)
	}
	return value, nil
}

func cloneOpenVPNStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneOpenVPNAnyMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneOpenVPNIntPointer(input *int) *int {
	if input == nil {
		return nil
	}
	value := *input
	return &value
}
