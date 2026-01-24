package protocol

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

// HTTP HTTP代理结构体
type HTTP struct {
	Name            string
	Server          string
	Port            interface{}
	Username        string
	Password        string
	TLS             bool
	SkipCertVerify  bool
	SNI             string
}

// IsHTTPLink 判断链接是否是HTTP/HTTPS代理节点链接
// 用于区分HTTP/HTTPS代理节点和订阅转换链接
func IsHTTPLink(link string) bool {
	if !strings.HasPrefix(strings.ToLower(link), "http://") && !strings.HasPrefix(strings.ToLower(link), "https://") {
		return false
	}
	
	// 尝试解析为HTTP/HTTPS代理节点
	_, err := DecodeHTTPURL(link)
	if err != nil {
		return false
	}
	
	// 进一步验证：HTTP/HTTPS代理节点必须包含有效的服务器地址和端口
	// 订阅转换链接通常不包含@符号（或者@符号后面不是有效的host:port格式）
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	
	// 检查是否有有效的host和port
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		// 如果没有端口，可能是订阅转换链接
		return false
	}
	
	// 检查host是否为空
	if host == "" {
		return false
	}
	
	// 检查port是否为有效数字
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum <= 0 || portNum > 65535 {
		return false
	}
	
	return true
}

// DecodeHTTPURL 解析HTTP/HTTPS代理URL
// 支持格式:
// - http://username:password@server:port#name
// - https://username:password@server:port?skip-cert-verify=true&sni=example.com#name
func DecodeHTTPURL(s string) (HTTP, error) {
	if !strings.Contains(s, "http://") && !strings.Contains(s, "https://") {
		return HTTP{}, fmt.Errorf("非http/https协议: %s", s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return HTTP{}, fmt.Errorf("url parse error: %v", err)
	}

	var httpProxy HTTP

	// 解析名称
	name := u.Fragment
	if name == "" {
		name = u.Host
	}
	httpProxy.Name = name

	// 解析服务器地址和端口
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		// 如果没有端口，使用默认端口
		if strings.Contains(err.Error(), "missing port") {
			host = u.Host
			if u.Scheme == "https" {
				port = "443"
			} else {
				port = "80"
			}
		} else {
			return HTTP{}, fmt.Errorf("SplitHostPort error: %v", err)
		}
	}
	httpProxy.Server = host

	rawPort := port
	if rawPort == "" {
		if u.Scheme == "https" {
			rawPort = "443"
		} else {
			rawPort = "80"
		}
	}
	httpProxy.Port, err = strconv.Atoi(rawPort)
	if err != nil {
		return HTTP{}, fmt.Errorf("Port conversion failed: %v", err)
	}

	// 解析用户名和密码
	httpProxy.Password, _ = u.User.Password()
	httpProxy.Username = u.User.Username()

	// 解析TLS配置
	httpProxy.TLS = (u.Scheme == "https")

	// 解析查询参数
	query := u.Query()
	if skipCert := query.Get("skip-cert-verify"); skipCert != "" {
		httpProxy.SkipCertVerify = (skipCert == "true" || skipCert == "1")
	}
	if sni := query.Get("sni"); sni != "" {
		httpProxy.SNI = sni
	}

	return httpProxy, nil
}

// EncodeHTTPURL 编码HTTP/HTTPS代理URL
func EncodeHTTPURL(h HTTP) string {
	scheme := "http"
	if h.TLS {
		scheme = "https"
	}

	u := url.URL{
		Scheme:   scheme,
		Host:     fmt.Sprintf("%s:%s", h.Server, utils.GetPortString(h.Port)),
		Fragment: h.Name,
	}

	// 设置用户名和密码
	if h.Username != "" {
		if h.Password != "" {
			u.User = url.UserPassword(h.Username, h.Password)
		} else {
			u.User = url.User(h.Username)
		}
	}

	// 构建查询参数
	query := url.Values{}
	if h.TLS {
		if h.SkipCertVerify {
			query.Set("skip-cert-verify", "true")
		}
		if h.SNI != "" {
			query.Set("sni", h.SNI)
		}
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if h.Name == "" {
		u.Fragment = fmt.Sprintf("%s:%s", h.Server, utils.GetPortString(h.Port))
	}

	return u.String()
}

// ConvertProxyToHTTP 将 Proxy 结构体转换为 HTTP 结构体
// 用于从 Clash 格式的代理配置生成 HTTP/HTTPS 链接
func ConvertProxyToHTTP(proxy Proxy) HTTP {
	return HTTP{
		Name:           proxy.Name,
		Server:         proxy.Server,
		Port:           int(proxy.Port),
		Username:       proxy.Username,
		Password:       proxy.Password,
		TLS:            proxy.Tls,
		SkipCertVerify: proxy.Skip_cert_verify,
		SNI:            proxy.Sni,
	}
}