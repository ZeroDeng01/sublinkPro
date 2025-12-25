package protocol

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

type HY2 struct {
	Password     string
	Host         string
	Port         interface{}
	MPort        string
	Insecure     int
	Peer         string
	Auth         string
	UpMbps       int
	DownMbps     int
	ALPN         []string
	Name         string
	Sni          string
	Obfs         string
	ObfsPassword string
}

// 开发者测试 CallHy 调用
func CallHy2() {
	hy2 := HY2{
		Password: "asdasd",
		Host:     "qq.com",
		Port:     11926,
		Insecure: 1,
		Peer:     "youku.com",
		Auth:     "",
		UpMbps:   11,
		DownMbps: 55,
		// ALPN:     "h3",
	}
	fmt.Println(EncodeHY2URL(hy2))
}

// hy2 编码
func EncodeHY2URL(hy2 HY2) string {
	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if hy2.Name == "" {
		hy2.Name = fmt.Sprintf("%s:%s", hy2.Host, utils.GetPortString(hy2.Port))
	}
	u := url.URL{
		Scheme:   "hy2",
		User:     url.User(hy2.Password),
		Host:     fmt.Sprintf("%s:%s", hy2.Host, utils.GetPortString(hy2.Port)),
		Fragment: hy2.Name,
	}
	q := u.Query()
	q.Set("insecure", strconv.Itoa(hy2.Insecure))
	q.Set("peer", hy2.Peer)
	q.Set("auth", hy2.Auth)
	q.Set("mport", hy2.MPort)
	q.Set("upmbps", strconv.Itoa(hy2.UpMbps))
	q.Set("downmbps", strconv.Itoa(hy2.DownMbps))
	q.Set("sni", hy2.Sni)
	q.Set("obfs", hy2.Obfs)
	q.Set("obfs-password", hy2.ObfsPassword)
	// 检查query是否有空值，有的话删除
	for k, v := range q {
		if v[0] == "" || v[0] == "0" {
			delete(q, k)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// hy2 解码
func DecodeHY2URL(s string) (HY2, error) {
	u, err := url.Parse(s)
	if err != nil {
		return HY2{}, fmt.Errorf("解析失败的URL: %s,错误:%s", s, err)
	}
	if u.Scheme != "hy2" && u.Scheme != "hysteria2" {
		return HY2{}, fmt.Errorf("非hy2协议: %s", s)
	}
	password := u.User.Username()
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
	mport := u.Query().Get("mport")
	alpn := strings.Split(alpns, ",")
	if alpns == "" {
		alpn = nil
	}
	sni := u.Query().Get("sni")
	obfs := u.Query().Get("obfs")
	obfsPassword := u.Query().Get("obfs-password")
	name := u.Fragment
	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if name == "" {
		name = server + ":" + u.Port()
	}
	if utils.CheckEnvironment() {
		fmt.Println("password:", password)
		fmt.Println("server:", server)
		fmt.Println("port:", port)
		fmt.Println("mport:", mport)
		fmt.Println("insecure:", insecure)
		fmt.Println("auth:", auth)
		fmt.Println("upMbps:", upMbps)
		fmt.Println("downMbps:", downMbps)
		fmt.Println("alpn:", alpn)
		fmt.Println("sni:", sni)
		fmt.Println("obfs:", obfs)
		fmt.Println("obfsPassword:", obfsPassword)
		fmt.Println("name:", name)
	}
	return HY2{
		Password:     password,
		Host:         server,
		Port:         port,
		MPort:        mport,
		Insecure:     insecure,
		Auth:         auth,
		UpMbps:       upMbps,
		DownMbps:     downMbps,
		ALPN:         alpn,
		Name:         name,
		Sni:          sni,
		Obfs:         obfs,
		ObfsPassword: obfsPassword,
	}, nil
}
