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
		FieldMeta{Name: "SkipCertVerify", Label: "跳过证书校验", Type: "bool", Group: "tls", Advanced: true},
		FieldMeta{Name: "ClientFingerprint", Label: "指纹", Type: "string", Group: "tls", Advanced: true},
	)
	MustRegisterProtocol(newProxyProtocolSpec(base, buildAnyTLSProxy, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "anytls")
	}, ConvertProxyToAnyTLS, EncodeAnyTLSURL))
}

type AnyTLS struct {
	Name              string
	Server            string
	Port              interface{}
	Password          string
	SkipCertVerify    bool
	SNI               string
	ClientFingerprint string
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
	anyTLS.ClientFingerprint = u.Query().Get("fp")

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
		Host:     fmt.Sprintf("%s:%s", a.Server, utils.GetPortString(a.Port)),
		Fragment: a.Name,
	}
	q := u.Query()
	if a.SkipCertVerify {
		q.Set("insecure", "1")
	}
	if a.SNI != "" {
		q.Set("sni", a.SNI)
	}
	if a.ClientFingerprint != "" {
		q.Set("fp", a.ClientFingerprint)
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
		Name:              proxy.Name,
		Server:            proxy.Server,
		Port:              int(proxy.Port),
		Password:          proxy.Password,
		SkipCertVerify:    proxy.Skip_cert_verify,
		SNI:               proxy.Sni,
		ClientFingerprint: proxy.Client_fingerprint,
	}
}

func buildAnyTLSProxy(link Urls, config OutputConfig) (Proxy, error) {
	anyTLS, err := DecodeAnyTLSURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	skipCert := config.Cert || anyTLS.SkipCertVerify
	return Proxy{Name: anyTLS.Name, Type: "anytls", Server: anyTLS.Server, Port: FlexPort(utils.GetPortInt(anyTLS.Port)), Password: anyTLS.Password, Skip_cert_verify: skipCert, Sni: anyTLS.SNI, Client_fingerprint: anyTLS.ClientFingerprint, Dialer_proxy: link.DialerProxyName}, nil
}
