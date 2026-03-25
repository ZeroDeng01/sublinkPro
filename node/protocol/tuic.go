package protocol

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

func init() {
	base := newProtocolSpec("tuic", []string{"tuic://"}, "TUIC", "#0277bd", "T", Tuic{}, "Name", DecodeTuicURL, EncodeTuicURL, func(t Tuic) LinkIdentity {
		return buildIdentity("tuic", t.Name, t.Host, utils.GetPortString(t.Port))
	},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic"},
		FieldMeta{Name: "Host", Label: "服务器地址", Type: "string", Group: "basic"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic"},
		FieldMeta{Name: "Uuid", Label: "UUID", Type: "string", Group: "auth", Secret: true, Advanced: true},
		FieldMeta{Name: "Password", Label: "密码", Type: "string", Group: "auth", Secret: true, Advanced: true},
		FieldMeta{Name: "Token", Label: "Token", Type: "string", Group: "auth", Secret: true, Advanced: true},
		FieldMeta{Name: "Version", Label: "版本", Type: "int", Group: "transport", Options: []string{"4", "5"}},
		FieldMeta{Name: "Congestion_control", Label: "拥塞控制", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Udp_relay_mode", Label: "UDP Relay Mode", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Disable_sni", Label: "禁用 SNI", Type: "int", Group: "tls", Advanced: true, Options: []string{"0", "1"}},
		FieldMeta{Name: "Tls", Label: "启用 TLS", Type: "bool", Group: "tls"},
		FieldMeta{Name: "Sni", Label: "SNI", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Alpn", Label: "ALPN", Type: "string", Group: "tls", Advanced: true, Multiline: true},
		FieldMeta{Name: "ClientFingerprint", Label: "指纹", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Insecure", Label: "跳过证书校验", Type: "int", Group: "tls", Advanced: true, Options: []string{"0", "1"}},
	)
	MustRegisterProtocol(newProxySurgeProtocolSpec(base, buildTuicProxy, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "tuic")
	}, ConvertProxyToTuic, EncodeTuicURL, buildTuicSurgeLine))
}

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

func buildTuicProxy(link Urls, config OutputConfig) (Proxy, error) {
	tuic, err := DecodeTuicURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	if tuic.Name == "" {
		tuic.Name = fmt.Sprintf("%s:%s", tuic.Host, utils.GetPortString(tuic.Port))
	}
	disableSNI := tuic.Disable_sni == 1
	skipCert := config.Cert || tuic.Insecure == 1
	return Proxy{Name: tuic.Name, Type: "tuic", Server: tuic.Host, Port: FlexPort(utils.GetPortInt(tuic.Port)), Password: tuic.Password, Uuid: tuic.Uuid, Congestion_controller: tuic.Congestion_control, Alpn: tuic.Alpn, Udp_relay_mode: tuic.Udp_relay_mode, Disable_sni: disableSNI, Sni: tuic.Sni, Tls: tuic.Tls, Client_fingerprint: tuic.ClientFingerprint, Udp: true, Skip_cert_verify: skipCert, Dialer_proxy: link.DialerProxyName, Version: tuic.Version, Token: tuic.Token}, nil
}

func buildTuicSurgeLine(link string, config OutputConfig) (string, string, error) {
	tuic, err := DecodeTuicURL(link)
	if err != nil {
		return "", "", err
	}
	server := replaceSurgeHost(tuic.Host, config)
	skipCert := config.Cert || tuic.Insecure == 1
	line := fmt.Sprintf("%s = tuic, %s, %d, token=%s, udp-relay=%t, skip-cert-verify=%t", tuic.Name, server, utils.GetPortInt(tuic.Port), tuic.Token, true, skipCert)
	if tuic.Version == 5 {
		line = fmt.Sprintf("%s = tuic, %s, %d, uuid=%s, password=%s, udp-relay=%t, skip-cert-verify=%t", tuic.Name, server, utils.GetPortInt(tuic.Port), tuic.Uuid, tuic.Password, true, skipCert)
	}
	return line, tuic.Name, nil
}
