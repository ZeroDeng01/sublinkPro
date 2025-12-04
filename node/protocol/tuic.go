package protocol

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

type Tuic struct {
	Name               string
	Password           string
	Host               string
	Port               int
	Uuid               string
	Congestion_control string
	Alpn               []string
	Sni                string
	Udp_relay_mode     string
	Disable_sni        int
}

// Tuic 解码
func DecodeTuicURL(s string) (Tuic, error) {
	u, err := url.Parse(s)
	if err != nil {
		return Tuic{}, fmt.Errorf("解析失败的URL: %s", s)
	}
	if u.Scheme != "tuic" {
		return Tuic{}, fmt.Errorf("非tuic协议: %s", s)
	}

	uuid := u.User.Username()
	password, _ := u.User.Password()
	// log.Println(password)
	// password = Base64Decode2(password)
	server := u.Hostname()
	port, _ := strconv.Atoi(u.Port())
	Congestioncontrol := u.Query().Get("Congestion_control")
	alpns := u.Query().Get("alpn")
	alpn := strings.Split(alpns, ",")
	if alpns == "" {
		alpn = nil
	}
	sni := u.Query().Get("sni")
	Udprelay_mode := u.Query().Get("Udp_relay_mode")
	Disablesni, _ := strconv.Atoi(u.Query().Get("Disable_sni"))
	name := u.Fragment
	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if name == "" {
		name = server + ":" + u.Port()
	}
	if utils.CheckEnvironment() {
		fmt.Println("password:", password)
		fmt.Println("server:", server)
		fmt.Println("port:", port)
		fmt.Println("insecure:", Congestioncontrol)
		fmt.Println("uuid:", uuid)
		fmt.Println("Udprelay_mode:", Udprelay_mode)
		fmt.Println("alpn:", alpn)
		fmt.Println("sni:", sni)
		fmt.Println("Disablesni:", Disablesni)
		fmt.Println("name:", name)
	}
	return Tuic{
		Name:               name,
		Password:           password,
		Host:               server,
		Port:               port,
		Uuid:               uuid,
		Congestion_control: Congestioncontrol,
		Alpn:               alpn,
		Sni:                sni,
		Udp_relay_mode:     Udprelay_mode,
		Disable_sni:        Disablesni,
	}, nil
}
