package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// TcpPing 测试TCP连接延迟
func TcpPing(host string, port int, timeout time.Duration) (int, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	duration := time.Since(start)
	return int(duration.Milliseconds()), nil
}

// ResolveIP 使用指定的 DNS 服务器解析域名 IP
func ResolveIP(host string) (string, error) {
	// 如果已经是 IP，直接返回
	if net.ParseIP(host) != nil {
		return host, nil
	}

	// 使用阿里 DNS
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * 5,
			}
			return d.DialContext(ctx, "udp", "223.5.5.5:53")
		},
	}

	ips, err := resolver.LookupHost(context.Background(), host)
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("no ip found for host: %s", host)
	}
	return ips[0], nil
}

// ParseNodeLink 解析节点链接获取 Host 和 Port
func ParseNodeLink(link string) (string, int, error) {
	if strings.HasPrefix(link, "vmess://") {
		return parseVmess(link)
	}
	if strings.HasPrefix(link, "vless://") {
		return parseVless(link)
	}
	if strings.HasPrefix(link, "trojan://") {
		return parseTrojan(link)
	}
	if strings.HasPrefix(link, "ss://") {
		return parseSS(link)
	}
	if strings.HasPrefix(link, "ssr://") {
		return parseSSR(link)
	}
	if strings.HasPrefix(link, "hy2://") || strings.HasPrefix(link, "hysteria2://") {
		return parseHy2(link)
	}
	if strings.HasPrefix(link, "socks5://") {
		return parseSocks5(link)
	}

	// 通用 URL 解析
	u, err := url.Parse(link)
	if err != nil {
		return "", 0, err
	}

	host := u.Hostname()
	portStr := u.Port()
	if portStr == "" {
		return "", 0, errors.New("no port found")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func parseVmess(link string) (string, int, error) {
	b64 := strings.TrimPrefix(link, "vmess://")
	// 处理可能的 padding
	if i := len(b64) % 4; i != 0 {
		b64 += strings.Repeat("=", 4-i)
	}
	decoded, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		// 尝试 URL safe base64
		decoded, err = base64.URLEncoding.DecodeString(b64)
		if err != nil {
			return "", 0, err
		}
	}
	var vmess map[string]interface{}
	if err := json.Unmarshal(decoded, &vmess); err != nil {
		return "", 0, err
	}

	add, ok := vmess["add"].(string)
	if !ok {
		return "", 0, errors.New("invalid vmess address")
	}

	// port 可能是 string 或 float64 (json number)
	var port int
	if p, ok := vmess["port"].(float64); ok {
		port = int(p)
	} else if pStr, ok := vmess["port"].(string); ok {
		p, err := strconv.Atoi(pStr)
		if err != nil {
			return "", 0, err
		}
		port = p
	} else {
		return "", 0, errors.New("invalid vmess port")
	}

	return add, port, nil
}

func parseVless(link string) (string, int, error) {
	// vless://uuid@host:port?params
	u, err := url.Parse(link)
	if err != nil {
		return "", 0, err
	}
	host := u.Hostname()
	portStr := u.Port()
	if portStr == "" {
		return "", 0, errors.New("vless port not found")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func parseTrojan(link string) (string, int, error) {
	// trojan://password@host:port?params
	u, err := url.Parse(link)
	if err != nil {
		return "", 0, err
	}
	host := u.Hostname()
	portStr := u.Port()
	if portStr == "" {
		return "", 0, errors.New("trojan port not found")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func parseSS(link string) (string, int, error) {
	// ss://method:pass@host:port
	// ss://base64(method:pass@host:port)
	// ss://base64(method:pass)@host:port

	u, err := url.Parse(link)
	if err != nil {
		return "", 0, err
	}

	// 如果 User 存在，说明是 ss://method:pass@host:port 格式
	if u.User != nil {
		host := u.Hostname()
		portStr := u.Port()
		if portStr != "" {
			port, err := strconv.Atoi(portStr)
			if err == nil {
				return host, port, nil
			}
		}
	}

	// 尝试处理 base64 格式
	// 可能是 ss://BASE64
	// 或者 ss://BASE64#name
	// url.Parse 会把 BASE64 当作 Host (如果没 @) 或者 User (如果有 @)

	// 获取可能包含 base64 的部分
	var b64 string
	if u.User != nil {
		// ss://BASE64@host:port 这种情况，User 是 BASE64
		// 但通常 ss://BASE64 这种格式，url.Parse 会解析为 Host=BASE64 (scheme://host)
		// 除非有 @
		b64 = u.User.Username()
	} else {
		b64 = u.Hostname()
	}

	// 尝试解码
	decoded := Base64Decode(b64)
	// 解码后可能是 method:pass@host:port
	if strings.Contains(decoded, "@") {
		parts := strings.Split(decoded, "@")
		if len(parts) == 2 {
			hostPort := parts[1]
			host, portStr, err := net.SplitHostPort(hostPort)
			if err == nil {
				port, err := strconv.Atoi(portStr)
				if err == nil {
					return host, port, nil
				}
			}
		}
	}

	// 如果上面没成功，可能已经是 host:port 格式了 (在 u.Hostname() 里)
	host := u.Hostname()
	portStr := u.Port()
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		return host, port, err
	}

	return "", 0, errors.New("unsupported ss link format")
}

func parseSSR(link string) (string, int, error) {
	// ssr://base64(server:port:protocol:method:obfs:base64pass/?params)
	b64 := strings.TrimPrefix(link, "ssr://")
	decoded := Base64Decode(b64)

	parts := strings.Split(decoded, ":")
	if len(parts) < 6 {
		return "", 0, errors.New("invalid ssr link")
	}

	host := parts[0]
	portStr := parts[1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}

	return host, port, nil
}

func parseHy2(link string) (string, int, error) {
	// hy2://password@host:port?params
	u, err := url.Parse(link)
	if err != nil {
		return "", 0, err
	}
	host := u.Hostname()
	portStr := u.Port()
	if portStr == "" {
		return "", 0, errors.New("hy2 port not found")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func parseSocks5(link string) (string, int, error) {
	// socks5://user:pass@host:port
	u, err := url.Parse(link)
	if err != nil {
		return "", 0, err
	}
	host := u.Hostname()
	portStr := u.Port()
	if portStr == "" {
		return "", 0, errors.New("socks5 port not found")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}
