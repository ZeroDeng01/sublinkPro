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
	base := newProtocolSpec("socks5", []string{"socks5://"}, "SOCKS5", "#116ea4", "S", Socks5{}, "Name", DecodeSocks5URL, EncodeSocks5URL, func(s Socks5) LinkIdentity {
		return buildIdentity("socks5", s.Name, s.Server, utils.GetPortString(s.Port))
	},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic"},
		FieldMeta{Name: "Server", Label: "服务器地址", Type: "string", Group: "basic"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic"},
		FieldMeta{Name: "Username", Label: "用户名", Type: "string", Group: "auth", Advanced: true},
		FieldMeta{Name: "Password", Label: "密码", Type: "string", Group: "auth", Secret: true, Advanced: true},
	)
	MustRegisterProtocol(newProxyProtocolSpec(base, func(link Urls, _ OutputConfig) (Proxy, error) {
		return buildSocks5Proxy(link)
	}, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "socks5")
	}, ConvertProxyToSocks5, EncodeSocks5URL))
}

type Socks5 struct {
	Name     string
	Server   string
	Port     interface{}
	Username string
	Password string
}

func DecodeSocks5URL(s string) (Socks5, error) {
	if !strings.Contains(s, "socks5://") {
		return Socks5{}, fmt.Errorf("非socks协议: %s", s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return Socks5{}, fmt.Errorf("url parse error: %v", err)
	}
	var socks5 Socks5
	name := u.Fragment
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		fmt.Println("Socks5 SplitHostPort error", err)
		return Socks5{}, err
	}
	rawPort := port
	if rawPort == "" {
		rawPort = "443"
	}
	socks5.Server = host
	socks5.Port, err = strconv.Atoi(rawPort)
	if err != nil {
		fmt.Println("Socks5 Port conversion failed:", err)
		return Socks5{}, err
	}
	socks5.Password, _ = u.User.Password()
	socks5.Username = u.User.Username()
	if name == "" {
		socks5.Name = u.Host
	} else {
		socks5.Name = name
	}
	return socks5, nil
}

// EncodeSocks5URL socks5 编码
func EncodeSocks5URL(s Socks5) string {
	u := url.URL{
		Scheme:   "socks5",
		Host:     fmt.Sprintf("%s:%s", s.Server, utils.GetPortString(s.Port)),
		Fragment: s.Name,
	}
	if s.Username != "" {
		if s.Password != "" {
			u.User = url.UserPassword(s.Username, s.Password)
		} else {
			u.User = url.User(s.Username)
		}
	}
	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if s.Name == "" {
		u.Fragment = fmt.Sprintf("%s:%s", s.Server, utils.GetPortString(s.Port))
	}
	return u.String()
}

// ConvertProxyToSocks5 将 Proxy 结构体转换为 Socks5 结构体
// 用于从 Clash 格式的代理配置生成 Socks5 链接
func ConvertProxyToSocks5(proxy Proxy) Socks5 {
	return Socks5{
		Name:     proxy.Name,
		Server:   proxy.Server,
		Port:     int(proxy.Port),
		Username: proxy.Username,
		Password: proxy.Password,
	}
}

func buildSocks5Proxy(link Urls) (Proxy, error) {
	socks5, err := DecodeSocks5URL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	return Proxy{Name: socks5.Name, Type: "socks5", Server: socks5.Server, Port: FlexPort(utils.GetPortInt(socks5.Port)), Username: socks5.Username, Password: socks5.Password, Dialer_proxy: link.DialerProxyName}, nil
}
