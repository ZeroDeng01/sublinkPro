package protocol

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

type Trojan struct {
	Password string      `json:"password"`
	Hostname string      `json:"hostname"`
	Port     interface{} `json:"port"`
	Query    TrojanQuery `json:"query,omitempty"`
	Name     string      `json:"name"`
	Type     string      `json:"type"`
}
type TrojanQuery struct {
	Peer          string   `json:"peer,omitempty"`
	Type          string   `json:"type,omitempty"`
	Path          string   `json:"path,omitempty"`
	Security      string   `json:"security,omitempty"`
	Fp            string   `json:"fp,omitempty"`
	AllowInsecure int      `json:"allowInsecure,omitempty"`
	Alpn          []string `json:"alpn,omitempty"`
	Sni           string   `json:"sni,omitempty"`
	Host          string   `json:"host,omitempty"`
	Flow          string   `json:"flow,omitempty"`
}

// 开发者测试
func CallTrojan() {
	trojan := Trojan{
		Password: "4cf3ca26cf114871b3d186a361a3de3",
		Hostname: "baidu.com",
		Port:     443,
		Query: TrojanQuery{
			Peer:          "",
			Type:          "tcp",
			Path:          "",
			Security:      "tls",
			Fp:            "",
			AllowInsecure: 0,
			Alpn:          []string{"h2", "http/1.1"},
			Sni:           "baidu.com",
			Host:          "",
			Flow:          "",
		},
	}
	fmt.Println(EncodeTrojanURL(trojan))
}

// trojan 编码
func EncodeTrojanURL(t Trojan) string {
	/*
		trojan://password@hostname:port?peer=example.com&allowInsecure=0&sni=example.com
	*/
	u := url.URL{
		Scheme: "trojan",
		User:   url.User(t.Password),
		Host:   fmt.Sprintf("%s:%s", t.Hostname, utils.GetPortString(t.Port)),
	}
	q := u.Query()
	q.Set("peer", t.Query.Peer)
	q.Set("allowInsecure", fmt.Sprintf("%d", t.Query.AllowInsecure))
	q.Set("sni", t.Query.Sni)
	q.Set("type", t.Query.Type)
	q.Set("path", t.Query.Path)
	q.Set("security", t.Query.Security)
	q.Set("fp", t.Query.Fp)
	// q.Set("alpn", t.Query.Alpn)
	q.Set("host", t.Query.Host)
	q.Set("flow", t.Query.Flow)
	// 检查query是否有空值，有的话删除
	for k, v := range q {
		if v[0] == "" {
			delete(q, k)
			// fmt.Printf("k: %v, v: %v\n", k, v)
		}
	}
	// 如果没有设置name,则使用hostname:port
	if t.Name == "" {
		t.Name = t.Hostname + ":" + utils.GetPortString(t.Port)
	}
	u.Fragment = t.Name
	u.RawQuery = q.Encode()
	return u.String()
}

// trojan 解码
func DecodeTrojanURL(s string) (Trojan, error) {
	/*
		trojan://password@hostname:port?peer=example.com&allowInsecure=0&sni=example.com
	*/
	u, err := url.Parse(s)
	if err != nil {
		return Trojan{}, fmt.Errorf("url格式化失败:%s", s)
	}
	if u.Scheme != "trojan" {
		return Trojan{}, fmt.Errorf("非trojan协议: %s", s)
	}
	password := u.User.Username()
	hostname := u.Hostname()
	rawPort := u.Port()
	if rawPort == "" {
		rawPort = "443"
	}
	port, _ := strconv.Atoi(rawPort)
	peer := u.Query().Get("peer")
	allowInsecure := u.Query().Get("allowInsecure")
	sni := u.Query().Get("sni")
	types := u.Query().Get("type")
	path := u.Query().Get("path")
	security := u.Query().Get("security")
	fp := u.Query().Get("fp")
	alpns := u.Query().Get("alpn")
	alpn := strings.Split(alpns, ",")
	if alpns == "" {
		alpn = nil
	}
	host := u.Query().Get("host")
	flow := u.Query().Get("flow")
	name := u.Fragment
	// 如果没有设置name,则使用hostname:port
	if name == "" {
		name = hostname + ":" + u.Port()
	}
	if utils.CheckEnvironment() {
		fmt.Println("password:", password)
		fmt.Println("password:", u.User.Username())
		fmt.Println("hostname:", hostname)
		fmt.Println("port:", port)
		fmt.Println("peer:", peer)
		fmt.Println("allowInsecure:", allowInsecure)
		fmt.Println("sni:", sni)
		fmt.Println("type:", types)
		fmt.Println("path:", path)
		fmt.Println("security:", security)
		fmt.Println("fp:", fp)
		fmt.Println("alpn:", alpn)
		fmt.Println("host:", host)
		fmt.Println("flow:", flow)
		fmt.Println("name:", name)
	}
	return Trojan{
		Password: password,
		Hostname: hostname,
		Port:     port,
		Query: TrojanQuery{
			Peer:          peer,
			Type:          types,
			Path:          path,
			Security:      security,
			Fp:            fp,
			AllowInsecure: 0,
			Alpn:          alpn,
			Sni:           sni,
			Host:          host,
			Flow:          flow,
		},
		Name: name,
		Type: "trojan",
	}, nil
}
