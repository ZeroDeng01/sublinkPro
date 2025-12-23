package protocol

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

type VLESS struct {
	Name   string      `json:"name"`
	Uuid   string      `json:"uuid"`
	Server string      `json:"server"`
	Port   interface{} `json:"port"`
	Query  VLESSQuery  `json:"query"`
}
type VLESSQuery struct {
	Security    string   `json:"security"`
	Alpn        []string `json:"alpn"`
	Sni         string   `json:"sni"`
	Fp          string   `json:"fp"`
	Sid         string   `json:"sid"`
	Pbk         string   `json:"pbk"`
	Flow        string   `json:"flow"`
	Encryption  string   `json:"encryption"`
	Type        string   `json:"type"`
	HeaderType  string   `json:"headerType"`
	Path        string   `json:"path"`
	Host        string   `json:"host"`
	ServiceName string   `json:"serviceName,omitempty"`
	Mode        string   `json:"mode,omitempty"`
}

func CallVLESS() {
	vless := VLESS{
		Name:   "Sharon-香港",
		Uuid:   "6adb4f43-9813-45f4-abf8-772be7db08sd",
		Server: "ss.com",
		Port:   443,
		Query: VLESSQuery{
			Security: "reality",
			// Alpn:       "",
			Sni:        "ss.com",
			Fp:         "chrome",
			Sid:        "",
			Pbk:        "g-oxbqigzCaXqARxuyD2_vbTYeMD9zn8wnTo02S69QM",
			Flow:       "xtls-rprx-vision",
			Encryption: "none",
			Type:       "tcp",
			HeaderType: "none",
			Path:       "",
			Host:       "",
		},
	}
	fmt.Println(EncodeVLESSURL(vless))
}

// vless编码
func EncodeVLESSURL(v VLESS) string {
	/*
		base64(username@host:port?encryption=none&security=auto&type=tcp)
	*/
	u := url.URL{
		Scheme: "vless",
		User:   url.User(v.Uuid),
		Host:   fmt.Sprintf("%s:%s", v.Server, utils.GetPortString(v.Port)),
	}
	q := u.Query()
	q.Set("security", v.Query.Security)
	// q.Set("alpn", v.Query.Alpn)
	q.Set("sni", v.Query.Sni)
	q.Set("fp", v.Query.Fp)
	q.Set("sid", v.Query.Sid)
	q.Set("pbk", v.Query.Pbk)
	q.Set("flow", v.Query.Flow)
	q.Set("encryption", v.Query.Encryption)
	q.Set("type", v.Query.Type)
	q.Set("headerType", v.Query.HeaderType)
	q.Set("path", v.Query.Path)
	q.Set("host", v.Query.Host)
	// 检查query是否有空值，有的话删除
	for k, v := range q {
		if v[0] == "" {
			delete(q, k)
			// fmt.Printf("k: %v, v: %v\n", k, v)
		}
	}
	u.RawQuery = q.Encode()
	// 如果没有name则用服务器加端口
	if v.Name == "" {
		u.Fragment = v.Server + ":" + utils.GetPortString(v.Port)
	} else {
		u.Fragment = v.Name
	}
	return u.String()
}

// vless解码
func DecodeVLESSURL(s string) (VLESS, error) {
	/*
		base64(username@host:port?encryption=none&security=auto&type=tcp)
	*/
	// 解析base64然后重新url编码
	if !strings.Contains(s, "vless://") {
		return VLESS{}, fmt.Errorf("非vless协议: %s", s)
	}
	s = "vless://" + utils.Base64Decode(strings.Split(s, "://")[1])
	// 解析url
	u, err := url.Parse(s)
	if err != nil {
		return VLESS{}, fmt.Errorf("url parse error: %v", err)
	}
	uuid := u.User.Username()
	if !utils.IsUUID(uuid) {
		utils.Error("❌节点解析错误：%v  【节点：%s】", "UUID格式错误", s)
		return VLESS{}, fmt.Errorf("uuid格式错误:%s", uuid)
	}
	hostname := u.Hostname()
	rawPort := u.Port()
	if rawPort == "" {
		if u.Query().Get("security") != "" {
			rawPort = "443"
		}
		if u.Query().Get("security") == "none" {
			rawPort = "80"
		}
	}
	port, _ := strconv.Atoi(rawPort)
	encryption := u.Query().Get("encryption")
	security := u.Query().Get("security")
	types := u.Query().Get("type")
	flow := u.Query().Get("flow")
	headerType := u.Query().Get("headerType")
	pbk := u.Query().Get("pbk")
	sid := u.Query().Get("sid")
	fp := u.Query().Get("fp")
	alpns := u.Query().Get("alpn")
	alpn := strings.Split(alpns, ",")
	if alpns == "" {
		alpn = nil
	}
	sni := u.Query().Get("sni")
	path := u.Query().Get("path")
	host := u.Query().Get("host")
	serviceName := u.Query().Get("serviceName")
	mode := u.Query().Get("mode")
	// 如果没有设置name,则使用hostname:port
	name := u.Fragment
	if name == "" {
		name = hostname + ":" + u.Port()
	}
	if utils.CheckEnvironment() {
		fmt.Println("uuid:", uuid)
		fmt.Println("hostname:", hostname)
		fmt.Println("port:", port)
		fmt.Println("encryption:", encryption)
		fmt.Println("security:", security)
		fmt.Println("type:", types)
		fmt.Println("flow:", flow)
		fmt.Println("headerType:", headerType)
		fmt.Println("pbk:", pbk)
		fmt.Println("sid:", sid)
		fmt.Println("fp:", fp)
		fmt.Println("alpn:", alpn)
		fmt.Println("sni:", sni)
		fmt.Println("path:", path)
		fmt.Println("host:", host)
		fmt.Println("serviceName:", serviceName)
		fmt.Println("mode:", mode)
		fmt.Println("name:", name)
	}
	return VLESS{
		Name:   name,
		Uuid:   uuid,
		Server: hostname,
		Port:   port,
		Query: VLESSQuery{
			Security:    security,
			Alpn:        alpn,
			Sni:         sni,
			Fp:          fp,
			Sid:         sid,
			Pbk:         pbk,
			Flow:        flow,
			Encryption:  encryption,
			Type:        types,
			HeaderType:  headerType,
			Path:        path,
			Host:        host,
			ServiceName: serviceName,
			Mode:        mode,
		},
	}, nil
}
