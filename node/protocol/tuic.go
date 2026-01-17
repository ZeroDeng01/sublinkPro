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
	Password           string //v5
	Host               string
	Port               interface{}
	Uuid               string //v5
	Congestion_control string
	Token              string //v4
	Version            int
	Alpn               []string
	Sni                string
	Udp_relay_mode     string
	Disable_sni        int
	Tls                bool   // TLS开关，对应URI中的security参数
	ClientFingerprint  string // 客户端指纹，对应URI中的fp参数
	Insecure           int    // 跳过证书验证，对应URI中的insecure参数
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
	if !utils.IsUUID(uuid) {
		utils.Error("❌节点解析错误：%v  【节点：%s】", "UUID格式错误", s)
		return Tuic{}, fmt.Errorf("uuid格式错误:%s", uuid)
	}
	password, _ := u.User.Password()
	// log.Println(password)
	// password = Base64Decode2(password)
	server := u.Hostname()
	rawPort := u.Port()
	if rawPort == "" {
		rawPort = "443"
	}
	port, _ := strconv.Atoi(rawPort)
	Congestioncontrol := u.Query().Get("congestion_control")
	alpns := u.Query().Get("alpn")
	alpn := strings.Split(alpns, ",")
	if alpns == "" {
		alpn = nil
	}
	sni := u.Query().Get("sni")
	Udprelay_mode := u.Query().Get("udp_relay_mode")
	Disablesni, _ := strconv.Atoi(u.Query().Get("disable_sni"))
	// 解析security参数，判断是否启用TLS
	security := u.Query().Get("security")
	tls := security == "tls" || security == ""
	// 解析fp参数，获取客户端指纹
	clientFingerprint := u.Query().Get("fp")
	// 解析 insecure 参数，跳过证书验证
	insecure, _ := strconv.Atoi(u.Query().Get("insecure"))
	name := u.Fragment
	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if name == "" {
		name = server + ":" + u.Port()
	}
	version := 5 // 默认版本 暂时只考虑支持v5
	token := ""
	if password == "" && uuid == "" {
		token = u.Query().Get("token")
		version = 4
	}
	if utils.CheckEnvironment() {
		fmt.Println("password:", password)
		fmt.Println("server:", server)
		fmt.Println("port:", port)
		fmt.Println("congestion_control:", Congestioncontrol)
		fmt.Println("insecure:", insecure)
		fmt.Println("uuid:", uuid)
		fmt.Println("udprelay_mode:", Udprelay_mode)
		fmt.Println("alpn:", alpn)
		fmt.Println("sni:", sni)
		fmt.Println("disablesni:", Disablesni)
		fmt.Println("name:", name)
		fmt.Println("version:", version)
		fmt.Println("token", token)
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
		Tls:                tls,
		ClientFingerprint:  clientFingerprint,
		Version:            version,
		Token:              token,
		Insecure:           insecure,
	}, nil
}

// EncodeTuicURL tuic 编码
func EncodeTuicURL(t Tuic) string {
	u := url.URL{
		Scheme:   "tuic",
		Host:     fmt.Sprintf("%s:%s", t.Host, utils.GetPortString(t.Port)),
		Fragment: t.Name,
	}
	// 设置用户信息：uuid:password
	if t.Password != "" {
		u.User = url.UserPassword(t.Uuid, t.Password)
	} else {
		u.User = url.User(t.Uuid)
	}
	q := u.Query()
	if t.Congestion_control != "" {
		q.Set("congestion_control", t.Congestion_control)
	}
	if len(t.Alpn) > 0 {
		q.Set("alpn", strings.Join(t.Alpn, ","))
	}
	if t.Sni != "" {
		q.Set("sni", t.Sni)
	}
	if t.Udp_relay_mode != "" {
		q.Set("udp_relay_mode", t.Udp_relay_mode)
	}
	if t.Disable_sni != 0 {
		q.Set("disable_sni", strconv.Itoa(t.Disable_sni))
	}
	// 编码security参数
	if t.Tls {
		q.Set("security", "tls")
	}
	// 编码客户端指纹
	if t.ClientFingerprint != "" {
		q.Set("fp", t.ClientFingerprint)
	}
	if t.Version == 5 {
		q.Set("version", strconv.Itoa(t.Version))
	}
	if t.Password == "" && t.Uuid == "" {
		q.Set("version", "4")
	}
	if t.Token != "" {
		q.Set("token", t.Token)
	}
	if t.Insecure != 0 {
		q.Set("insecure", strconv.Itoa(t.Insecure))
	}

	u.RawQuery = q.Encode()
	// 如果没有设置 Name，则使用 Host:Port 作为 Fragment
	if t.Name == "" {
		u.Fragment = fmt.Sprintf("%s:%s", t.Host, utils.GetPortString(t.Port))
	}
	return u.String()
}

// ConvertProxyToTuic 将 Proxy 结构体转换为 Tuic 结构体
// 用于从 Clash 格式的代理配置生成 TUIC 链接
func ConvertProxyToTuic(proxy Proxy) Tuic {
	tuic := Tuic{
		Name:               proxy.Name,
		Password:           proxy.Password,
		Host:               proxy.Server,
		Port:               int(proxy.Port),
		Uuid:               proxy.Uuid,
		Congestion_control: proxy.Congestion_controller,
		Alpn:               proxy.Alpn,
		Udp_relay_mode:     proxy.Udp_relay_mode,
		Tls:                proxy.Tls,
		ClientFingerprint:  proxy.Client_fingerprint,
		Version:            proxy.Version,
		Token:              proxy.Token,
	}

	// 处理 Sni
	if proxy.Sni != "" {
		tuic.Sni = proxy.Sni
	} else if proxy.Servername != "" {
		tuic.Sni = proxy.Servername
	}

	// 处理 disable_sni
	if proxy.Disable_sni {
		tuic.Disable_sni = 1
	}

	// 处理跳过证书验证
	if proxy.Skip_cert_verify {
		tuic.Insecure = 1
	}

	// 设置默认版本
	if tuic.Version == 0 {
		tuic.Version = 5
	}

	return tuic
}
