package protocol

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

func init() {
	base := newProtocolSpec("anytls", []string{"anytls://"}, "AnyTLS", "#20a84c", "A", AnyTLS{}, "Name", DecodeAnyTLSURL, EncodeAnyTLSURL, func(a AnyTLS) LinkIdentity {
		return buildIdentity("anytls", a.Name, a.Server, utils.GetPortString(a.Port))
	},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic"},
		FieldMeta{Name: "Server", Label: "服务器地址", Type: "string", Group: "basic"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic"},
		FieldMeta{Name: "Password", Label: "密码", Type: "string", Group: "auth", Secret: true},
		FieldMeta{Name: "SNI", Label: "SNI", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "ALPN", Label: "ALPN", Type: "string", Group: "tls", Advanced: true, Multiline: true},
		FieldMeta{Name: "SkipCertVerify", Label: "跳过证书校验", Type: "bool", Group: "tls", Advanced: true},
		FieldMeta{Name: "ClientFingerprint", Label: "指纹", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Fingerprint", Label: "证书指纹", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "UDP", Label: "UDP", Type: "bool", Group: "transport", Advanced: true},
		FieldMeta{Name: "IdleSessionCheckInterval", Label: "空闲会话检查间隔", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "IdleSessionTimeout", Label: "空闲会话超时时间", Type: "int", Group: "transport", Advanced: true},
		FieldMeta{Name: "MinIdleSession", Label: "最小空闲会话数", Type: "int", Group: "transport", Advanced: true},
	)
	MustRegisterProtocol(newProxySurgeProtocolSpec(base, buildAnyTLSProxy, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "anytls")
	}, ConvertProxyToAnyTLS, EncodeAnyTLSURL, buildAnyTLSSurgeLine))
}

type AnyTLS struct {
	Name                     string
	Server                   string
	Port                     any
	Password                 string
	SkipCertVerify           bool
	SNI                      string
	ALPN                     []string
	ClientFingerprint        string
	Fingerprint              string
	UDP                      bool
	IdleSessionCheckInterval int
	IdleSessionTimeout       int
	MinIdleSession           int
}

// DecodeAnyTLSURL 解析 AnyTLS 链接，并补齐默认端口与基础 TLS 相关字段。
func DecodeAnyTLSURL(s string) (AnyTLS, error) {

	if !strings.Contains(s, "anytls://") {
		return AnyTLS{}, fmt.Errorf("非anytls协议: %s", s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return AnyTLS{}, fmt.Errorf("url parse error: %v", err)
	}
	var anyTLS AnyTLS
	name := u.Fragment
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		fmt.Println("AnyTLS SplitHostPort error", err)
		return AnyTLS{}, err
	}
	anyTLS.Server = host
	rawPort := port
	if rawPort == "" {
		rawPort = "443"
	}
	anyTLS.Port, err = strconv.Atoi(rawPort)
	if err != nil {
		fmt.Println("AnyTLS Port conversion failed:", err)
		return AnyTLS{}, err
	}
	anyTLS.Password = u.User.Username()
	skipCertVerify := u.Query().Get("insecure")
	if skipCertVerify != "" {
		anyTLS.SkipCertVerify, err = strconv.ParseBool(skipCertVerify)
	}
	if err != nil {
		fmt.Println("AnyTLS SkipCertVerify conversion failed:", err)
		return AnyTLS{}, err
	}
	anyTLS.SNI = u.Query().Get("sni")
	anyTLS.ALPN = splitVMessALPN(u.Query().Get("alpn"))
	anyTLS.ClientFingerprint = u.Query().Get("fp")
	anyTLS.Fingerprint = sanitizeCertificateFingerprint(u.Query().Get("fingerprint"))
	anyTLS.UDP, _ = strconv.ParseBool(u.Query().Get("udp"))
	anyTLS.IdleSessionCheckInterval, _ = strconv.Atoi(u.Query().Get("idle-session-check-interval"))
	anyTLS.IdleSessionTimeout, _ = strconv.Atoi(u.Query().Get("idle-session-timeout"))
	anyTLS.MinIdleSession, _ = strconv.Atoi(u.Query().Get("min-idle-session"))

	if name == "" {
		anyTLS.Name = u.Host
	} else {
		anyTLS.Name = name
	}
	return anyTLS, nil
}

// EncodeAnyTLSURL anytls 编码
func EncodeAnyTLSURL(a AnyTLS) string {
	u := url.URL{
		Scheme:   "anytls",
		User:     url.User(a.Password),
		Host:     formatURLHostPort(a.Server, utils.GetPortString(a.Port)),
		Fragment: a.Name,
	}
	q := u.Query()
	if a.SkipCertVerify {
		q.Set("insecure", "1")
	}
	if a.SNI != "" {
		q.Set("sni", a.SNI)
	}
	if len(a.ALPN) > 0 {
		q.Set("alpn", strings.Join(a.ALPN, ","))
	}
	if a.ClientFingerprint != "" {
		q.Set("fp", a.ClientFingerprint)
	}
	if a.Fingerprint != "" {
		q.Set("fingerprint", a.Fingerprint)
	}
	if a.UDP {
		q.Set("udp", "1")
	}
	if a.IdleSessionCheckInterval > 0 {
		q.Set("idle-session-check-interval", strconv.Itoa(a.IdleSessionCheckInterval))
	}
	if a.IdleSessionTimeout > 0 {
		q.Set("idle-session-timeout", strconv.Itoa(a.IdleSessionTimeout))
	}
	if a.MinIdleSession > 0 {
		q.Set("min-idle-session", strconv.Itoa(a.MinIdleSession))
	}
	u.RawQuery = q.Encode()
	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if a.Name == "" {
		u.Fragment = fmt.Sprintf("%s:%s", a.Server, utils.GetPortString(a.Port))
	}
	return u.String()
}

// ConvertProxyToAnyTLS 将 Proxy 结构体转换为 AnyTLS 结构体
// 用于从 Clash 格式的代理配置生成 AnyTLS 链接
func ConvertProxyToAnyTLS(proxy Proxy) AnyTLS {
	return AnyTLS{
		Name:                     proxy.Name,
		Server:                   proxy.Server,
		Port:                     int(proxy.Port),
		Password:                 proxy.Password,
		SkipCertVerify:           proxy.Skip_cert_verify,
		SNI:                      proxy.Sni,
		ALPN:                     proxy.Alpn,
		ClientFingerprint:        proxy.Client_fingerprint,
		Fingerprint:              sanitizeCertificateFingerprint(proxy.Fingerprint),
		UDP:                      proxy.Udp,
		IdleSessionCheckInterval: proxy.AnyTLSIdleCheck,
		IdleSessionTimeout:       proxy.AnyTLSIdleTimeout,
		MinIdleSession:           proxy.AnyTLSMinIdle,
	}
}

func buildAnyTLSProxy(link Urls, config OutputConfig) (Proxy, error) {
	anyTLS, err := DecodeAnyTLSURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	skipCert := config.Cert || anyTLS.SkipCertVerify
	udp := config.Udp || anyTLS.UDP
	return Proxy{Name: anyTLS.Name, Type: "anytls", Server: anyTLS.Server, Port: FlexPort(utils.GetPortInt(anyTLS.Port)), Password: anyTLS.Password, Skip_cert_verify: skipCert, Sni: anyTLS.SNI, Alpn: anyTLS.ALPN, Client_fingerprint: anyTLS.ClientFingerprint, Fingerprint: anyTLS.Fingerprint, Udp: udp, AnyTLSIdleCheck: anyTLS.IdleSessionCheckInterval, AnyTLSIdleTimeout: anyTLS.IdleSessionTimeout, AnyTLSMinIdle: anyTLS.MinIdleSession, Dialer_proxy: link.DialerProxyName}, nil
}

// buildAnyTLSSurgeLine 将 AnyTLS 链接转换为 Surge 节点行。
// Surge 侧只输出当前可映射的核心字段；mihomo 专属的空闲会话池参数不写入 Surge 配置，避免生成无效参数。
func buildAnyTLSSurgeLine(link string, config OutputConfig) (string, string, error) {
	anyTLS, err := DecodeAnyTLSURL(link)
	if err != nil {
		return "", "", err
	}
	server := replaceSurgeHost(anyTLS.Server, config)
	skipCert := config.Cert || anyTLS.SkipCertVerify
	line := fmt.Sprintf("%s = anytls, %s, %d, password=%s, skip-cert-verify=%t", anyTLS.Name, server, utils.GetPortInt(anyTLS.Port), anyTLS.Password, skipCert)
	if anyTLS.SNI != "" {
		line = fmt.Sprintf("%s, sni=%s", line, anyTLS.SNI)
	}
	if anyTLS.Fingerprint != "" {
		line = fmt.Sprintf("%s, server-cert-fingerprint-sha256=%s", line, anyTLS.Fingerprint)
	}
	return line, anyTLS.Name, nil
}
