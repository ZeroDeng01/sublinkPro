package protocol

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

type HY struct {
	Host     string
	Port     interface{}
	Insecure int
	Peer     string
	Auth     string
	UpMbps   int
	DownMbps int
	ALPN     []string
	Name     string
}

// 开发者测试 CallHy 调用
func CallHy() {
	hy := HY{
		Host:     "qq.com",
		Port:     11926,
		Insecure: 1,
		Peer:     "youku.com",
		Auth:     "",
		UpMbps:   11,
		DownMbps: 55,
		// ALPN:     "h3",
	}
	fmt.Println(EncodeHYURL(hy))
}

// hy 编码
func EncodeHYURL(hy HY) string {
	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if hy.Name == "" {
		hy.Name = fmt.Sprintf("%s:%s", hy.Host, utils.GetPortString(hy.Port))
	}
	u := url.URL{
		Scheme:   "hysteria",
		Host:     fmt.Sprintf("%s:%s", hy.Host, utils.GetPortString(hy.Port)),
		Fragment: hy.Name,
	}
	q := u.Query()
	q.Set("insecure", strconv.Itoa(hy.Insecure))
	q.Set("peer", hy.Peer)
	q.Set("auth", hy.Auth)
	q.Set("upmbps", strconv.Itoa(hy.UpMbps))
	q.Set("downmbps", strconv.Itoa(hy.DownMbps))
	// q.Set("alpn", hy.ALPN)
	// 检查query是否有空值，有的话删除
	for k, v := range q {
		if v[0] == "" {
			delete(q, k)
			// fmt.Printf("k: %v, v: %v\n", k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// hy 解码
func DecodeHYURL(s string) (HY, error) {
	u, err := url.Parse(s)
	if err != nil {
		return HY{}, fmt.Errorf("失败的URL: %s", s)
	}
	if u.Scheme != "hy" && u.Scheme != "hysteria" {
		return HY{}, fmt.Errorf("非hy协议: %s", s)
	}
	server := u.Hostname()
	rawPort := u.Port()
	if rawPort == "" {
		rawPort = "443"
	}
	port, _ := strconv.Atoi(rawPort)
	insecure, _ := strconv.Atoi(u.Query().Get("insecure"))
	auth := u.Query().Get("auth")
	upMbps, _ := strconv.Atoi(u.Query().Get("upmbps"))
	downMbps, _ := strconv.Atoi(u.Query().Get("downmbps"))
	alpns := u.Query().Get("alpn")
	alpn := strings.Split(alpns, ",")
	if alpns == "" {
		alpn = nil
	}
	// 如果没有设置 Name，则使用 Fragment 作为 Name
	name := u.Fragment
	if name == "" {
		name = server + ":" + u.Port()
	}
	if utils.CheckEnvironment() {
		fmt.Println("server:", server)
		fmt.Println("port:", port)
		fmt.Println("insecure:", insecure)
		fmt.Println("auth:", auth)
		fmt.Println("upMbps:", upMbps)
		fmt.Println("downMbps:", downMbps)
		fmt.Println("alpn:", alpn)
		fmt.Println("name:", name)
	}
	return HY{
		Host:     server,
		Port:     port,
		Insecure: insecure,
		Auth:     auth,
		UpMbps:   upMbps,
		DownMbps: downMbps,
		ALPN:     alpn,
		Name:     name,
	}, nil
}
