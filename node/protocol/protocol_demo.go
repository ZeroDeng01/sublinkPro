package protocol

import (
	"fmt"
	"net/url"
	"sublink/utils"
)

type ProtocolDemo struct {
	Name     string      `json:"name"`
	Server   string      `json:"server"`
	Port     interface{} `json:"port"`
	Token    string      `json:"token"`
	Mode     string      `json:"mode"`
	TLS      bool        `json:"tls"`
	Insecure bool        `json:"insecure"`
}

func init() {
	base := newProtocolSpec(
		"protocoldemo",
		[]string{"demo://", "protocoldemo://"},
		"ProtocolDemo",
		"#455a64",
		"D",
		ProtocolDemo{},
		"Name",
		DecodeProtocolDemoURL,
		EncodeProtocolDemoURL,
		func(p ProtocolDemo) LinkIdentity {
			return buildIdentity("protocoldemo", p.Name, p.Server, utils.GetPortString(p.Port))
		},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic", Placeholder: "例如：Demo 节点"},
		FieldMeta{Name: "Server", Label: "服务器地址", Type: "string", Group: "basic", Placeholder: "demo.example.com"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic", Placeholder: "443"},
		FieldMeta{Name: "Token", Label: "令牌", Type: "string", Group: "auth", Secret: true, Placeholder: "demo-token"},
		FieldMeta{Name: "Mode", Label: "传输模式", Type: "string", Group: "transport", Options: []string{"tcp", "ws", "grpc"}},
		FieldMeta{Name: "TLS", Label: "启用 TLS", Type: "bool", Group: "tls"},
		FieldMeta{Name: "Insecure", Label: "跳过证书校验", Type: "bool", Group: "tls", Advanced: true},
	)

	MustRegisterProtocol(newProxySurgeProtocolSpec(
		base,
		buildProtocolDemoProxy,
		func(proxy Proxy) bool {
			return proxyTypeMatches(proxy, "protocoldemo")
		},
		ConvertProxyToProtocolDemo,
		EncodeProtocolDemoURL,
		buildProtocolDemoSurgeLine,
	))
}

func DecodeProtocolDemoURL(raw string) (ProtocolDemo, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return ProtocolDemo{}, fmt.Errorf("invalid ProtocolDemo URL: %w", err)
	}
	if u.Scheme != "demo" && u.Scheme != "protocoldemo" {
		return ProtocolDemo{}, fmt.Errorf("invalid ProtocolDemo scheme: %s", u.Scheme)
	}

	query := u.Query()
	name := u.Fragment
	if name == "" {
		name = u.Host
	}

	return ProtocolDemo{
		Name:     name,
		Server:   u.Hostname(),
		Port:     utils.GetPortInt(u.Port()),
		Token:    query.Get("token"),
		Mode:     query.Get("mode"),
		TLS:      query.Get("tls") == "1" || query.Get("tls") == "true",
		Insecure: query.Get("insecure") == "1" || query.Get("insecure") == "true",
	}, nil
}

func EncodeProtocolDemoURL(p ProtocolDemo) string {
	query := url.Values{}
	if p.Token != "" {
		query.Set("token", p.Token)
	}
	if p.Mode != "" {
		query.Set("mode", p.Mode)
	}
	if p.TLS {
		query.Set("tls", "true")
	}
	if p.Insecure {
		query.Set("insecure", "true")
	}

	if p.Name == "" {
		p.Name = fmt.Sprintf("%s:%s", p.Server, utils.GetPortString(p.Port))
	}

	return (&url.URL{
		Scheme:   "demo",
		Host:     fmt.Sprintf("%s:%s", p.Server, utils.GetPortString(p.Port)),
		RawQuery: query.Encode(),
		Fragment: p.Name,
	}).String()
}

func buildProtocolDemoProxy(link Urls, config OutputConfig) (Proxy, error) {
	p, err := DecodeProtocolDemoURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	return Proxy{
		Name:             p.Name,
		Type:             "protocoldemo",
		Server:           p.Server,
		Port:             FlexPort(utils.GetPortInt(p.Port)),
		Password:         p.Token,
		Network:          p.Mode,
		Tls:              p.TLS,
		Skip_cert_verify: config.Cert || p.Insecure,
		Udp:              config.Udp,
		Dialer_proxy:     link.DialerProxyName,
	}, nil
}

func buildProtocolDemoSurgeLine(link string, config OutputConfig) (string, string, error) {
	p, err := DecodeProtocolDemoURL(link)
	if err != nil {
		return "", "", err
	}
	server := replaceSurgeHost(p.Server, config)
	line := fmt.Sprintf("%s = custom, %s, %d, token=%s, mode=%s, tls=%t, skip-cert-verify=%t", p.Name, server, utils.GetPortInt(p.Port), p.Token, p.Mode, p.TLS, config.Cert || p.Insecure)
	return line, p.Name, nil
}

func ConvertProxyToProtocolDemo(proxy Proxy) ProtocolDemo {
	return ProtocolDemo{
		Name:     proxy.Name,
		Server:   proxy.Server,
		Port:     int(proxy.Port),
		Token:    proxy.Password,
		Mode:     proxy.Network,
		TLS:      proxy.Tls,
		Insecure: proxy.Skip_cert_verify,
	}
}
