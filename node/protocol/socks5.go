package protocol

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

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
