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
	Security      string   `json:"security"`
	Alpn          []string `json:"alpn"`
	Sni           string   `json:"sni"`
	Fp            string   `json:"fp"`
	Sid           string   `json:"sid"`
	Pbk           string   `json:"pbk"`
	Flow          string   `json:"flow"`
	Encryption    string   `json:"encryption"`
	Type          string   `json:"type"`
	HeaderType    string   `json:"headerType"`
	Path          string   `json:"path"`
	Host          string   `json:"host"`
	ServiceName   string   `json:"serviceName,omitempty"`
	Mode          string   `json:"mode,omitempty"`
	AllowInsecure int      `json:"allowInsecure,omitempty"` // 跳过证书验证
	// 新增：packet-encoding参数（xudp/packetaddr）
	PacketEncoding string `json:"packetEncoding,omitempty"`
	// 新增：ws传输层参数
	MaxEarlyData        int    `json:"maxEarlyData,omitempty"`        // Early Data首包长度阈值
	EarlyDataHeader     string `json:"earlyDataHeader,omitempty"`     // Early Data头名称
	HttpUpgrade         int    `json:"httpUpgrade,omitempty"`         // v2ray-http-upgrade (0/1)
	HttpUpgradeFastOpen int    `json:"httpUpgradeFastOpen,omitempty"` // v2ray-http-upgrade-fast-open (0/1)
	// 新增：http传输层参数
	Method string `json:"method,omitempty"` // HTTP请求方法
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
// 输出v2ray格式的VLESS链接（明文URL格式）
func EncodeVLESSURL(v VLESS) string {
	u := url.URL{
		Scheme: "vless",
		User:   url.User(v.Uuid),
		Host:   fmt.Sprintf("%s:%s", v.Server, utils.GetPortString(v.Port)),
	}
	q := u.Query()

	// 基本参数
	q.Set("encryption", v.Query.Encryption)
	q.Set("security", v.Query.Security)
	q.Set("type", v.Query.Type)

	// TLS相关参数
	q.Set("sni", v.Query.Sni)
	q.Set("fp", v.Query.Fp)
	if len(v.Query.Alpn) > 0 {
		q.Set("alpn", strings.Join(v.Query.Alpn, ","))
	}

	// Reality参数
	q.Set("pbk", v.Query.Pbk)
	q.Set("sid", v.Query.Sid)

	// VLESS特有参数
	q.Set("flow", v.Query.Flow)
	q.Set("headerType", v.Query.HeaderType)
	if v.Query.PacketEncoding != "" {
		q.Set("packetEncoding", v.Query.PacketEncoding)
	}

	// 传输层通用参数
	q.Set("path", v.Query.Path)
	q.Set("host", v.Query.Host)

	// gRPC参数
	if v.Query.ServiceName != "" {
		q.Set("serviceName", v.Query.ServiceName)
	}
	if v.Query.Mode != "" {
		q.Set("mode", v.Query.Mode)
	}

	// ws传输层参数
	if v.Query.MaxEarlyData > 0 {
		q.Set("ed", strconv.Itoa(v.Query.MaxEarlyData))
	}
	if v.Query.EarlyDataHeader != "" {
		q.Set("eh", v.Query.EarlyDataHeader)
	}

	// http传输层参数
	if v.Query.Method != "" {
		q.Set("method", v.Query.Method)
	}

	// 跳过证书验证
	if v.Query.AllowInsecure == 1 {
		q.Set("allowInsecure", "1")
	}

	// 检查query是否有空值，有的话删除
	for k, val := range q {
		if val[0] == "" {
			delete(q, k)
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
// v2ray格式的VLESS链接是明文URL，不需要base64解码
// 格式: vless://UUID@server:port?参数#名称
func DecodeVLESSURL(s string) (VLESS, error) {
	if !strings.HasPrefix(s, "vless://") {
		return VLESS{}, fmt.Errorf("非vless协议: %s", s)
	}

	// 直接解析URL（v2ray格式是明文URL，不需要base64解码）
	u, err := url.Parse(s)
	if err != nil {
		return VLESS{}, fmt.Errorf("url parse error: %v", err)
	}

	uuid := u.User.Username()
	if !utils.IsUUID(uuid) {
		utils.Error("❌节点解析错误：%v  【节点：%s】", "UUID格式错误", s)
		return VLESS{}, fmt.Errorf("uuid格式错误:%s", uuid)
	}

	// 处理服务器地址（支持IPv6格式[::1]）
	hostname := utils.UnwrapIPv6Host(u.Hostname())

	// 处理端口
	rawPort := u.Port()
	if rawPort == "" {
		security := u.Query().Get("security")
		if security == "none" || security == "" {
			rawPort = "80"
		} else {
			rawPort = "443"
		}
	}
	port, _ := strconv.Atoi(rawPort)

	// 解析基本参数
	encryption := u.Query().Get("encryption")
	security := u.Query().Get("security")
	types := u.Query().Get("type")
	flow := u.Query().Get("flow")
	headerType := u.Query().Get("headerType")
	pbk := u.Query().Get("pbk")
	sid := u.Query().Get("sid")
	fp := u.Query().Get("fp")
	sni := u.Query().Get("sni")
	path := u.Query().Get("path")
	host := u.Query().Get("host")
	serviceName := u.Query().Get("serviceName")
	mode := u.Query().Get("mode")

	// 解析 alpn 参数（逗号分隔）
	alpns := u.Query().Get("alpn")
	var alpn []string
	if alpns != "" {
		alpn = strings.Split(alpns, ",")
	}

	// 解析 allowInsecure 参数
	allowInsecure := 0
	insecureStr := u.Query().Get("allowInsecure")
	if insecureStr == "1" || insecureStr == "true" {
		allowInsecure = 1
	}

	// 解析 packet-encoding 参数（packetEncoding 或 packet_encoding）
	packetEncoding := u.Query().Get("packetEncoding")
	if packetEncoding == "" {
		packetEncoding = u.Query().Get("packet_encoding")
	}

	// 解析 ws 传输层参数
	maxEarlyData := 0
	if ed := u.Query().Get("ed"); ed != "" {
		maxEarlyData, _ = strconv.Atoi(ed)
	}
	earlyDataHeader := u.Query().Get("eh")

	// 解析 v2ray-http-upgrade 参数
	httpUpgrade := 0
	if hup := u.Query().Get("httpUpgrade"); hup == "1" || hup == "true" {
		httpUpgrade = 1
	}
	httpUpgradeFastOpen := 0
	if hupfo := u.Query().Get("httpUpgradeFastOpen"); hupfo == "1" || hupfo == "true" {
		httpUpgradeFastOpen = 1
	}

	// 解析 http 传输层参数
	method := u.Query().Get("method")

	// 解析名称（URL fragment）
	name := u.Fragment
	if name == "" {
		name = hostname + ":" + rawPort
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
		fmt.Println("packetEncoding:", packetEncoding)
		fmt.Println("maxEarlyData:", maxEarlyData)
		fmt.Println("earlyDataHeader:", earlyDataHeader)
		fmt.Println("httpUpgrade:", httpUpgrade)
		fmt.Println("method:", method)
		fmt.Println("name:", name)
	}

	return VLESS{
		Name:   name,
		Uuid:   uuid,
		Server: hostname,
		Port:   port,
		Query: VLESSQuery{
			Security:            security,
			Alpn:                alpn,
			Sni:                 sni,
			Fp:                  fp,
			Sid:                 sid,
			Pbk:                 pbk,
			Flow:                flow,
			Encryption:          encryption,
			Type:                types,
			HeaderType:          headerType,
			Path:                path,
			Host:                host,
			ServiceName:         serviceName,
			Mode:                mode,
			AllowInsecure:       allowInsecure,
			PacketEncoding:      packetEncoding,
			MaxEarlyData:        maxEarlyData,
			EarlyDataHeader:     earlyDataHeader,
			HttpUpgrade:         httpUpgrade,
			HttpUpgradeFastOpen: httpUpgradeFastOpen,
			Method:              method,
		},
	}, nil
}

// ConvertProxyToVless 将 Proxy 结构体转换为 VLESS 结构体
// 用于从 Clash 格式的代理配置生成 VLESS 链接
func ConvertProxyToVless(proxy Proxy) VLESS {
	vless := VLESS{
		Name:   proxy.Name,
		Uuid:   proxy.Uuid,
		Server: proxy.Server,
		Port:   int(proxy.Port),
		Query: VLESSQuery{
			Sni:            proxy.Servername,
			Fp:             proxy.Client_fingerprint,
			Flow:           proxy.Flow,
			Alpn:           proxy.Alpn,
			Type:           proxy.Network,
			PacketEncoding: proxy.Packet_encoding,
		},
	}

	// 处理跳过证书验证
	if proxy.Skip_cert_verify {
		vless.Query.AllowInsecure = 1
	}

	// 处理 security 参数（TLS/Reality/none）
	if len(proxy.Reality_opts) > 0 {
		vless.Query.Security = "reality"
		if pbk, ok := proxy.Reality_opts["public-key"].(string); ok {
			vless.Query.Pbk = pbk
		}
		if sid, ok := proxy.Reality_opts["short-id"].(string); ok {
			vless.Query.Sid = sid
		}
	} else if proxy.Tls {
		vless.Query.Security = "tls"
	} else {
		vless.Query.Security = "none"
	}

	// 处理 ws_opts
	if len(proxy.Ws_opts) > 0 {
		if path, ok := proxy.Ws_opts["path"].(string); ok {
			vless.Query.Path = path
		}
		if headers, ok := proxy.Ws_opts["headers"].(map[string]interface{}); ok {
			if host, ok := headers["Host"].(string); ok {
				vless.Query.Host = host
			}
		}
		if ed, ok := proxy.Ws_opts["max-early-data"].(int); ok {
			vless.Query.MaxEarlyData = ed
		}
		if edh, ok := proxy.Ws_opts["early-data-header-name"].(string); ok {
			vless.Query.EarlyDataHeader = edh
		}
		if hup, ok := proxy.Ws_opts["v2ray-http-upgrade"].(bool); ok && hup {
			vless.Query.HttpUpgrade = 1
		}
		if hupfo, ok := proxy.Ws_opts["v2ray-http-upgrade-fast-open"].(bool); ok && hupfo {
			vless.Query.HttpUpgradeFastOpen = 1
		}
	}

	// 处理 h2_opts
	if len(proxy.H2_opts) > 0 {
		if path, ok := proxy.H2_opts["path"].(string); ok {
			vless.Query.Path = path
		}
		if hosts, ok := proxy.H2_opts["host"].([]string); ok && len(hosts) > 0 {
			vless.Query.Host = hosts[0]
		}
		if host, ok := proxy.H2_opts["host"].([]interface{}); ok && len(host) > 0 {
			if h, ok := host[0].(string); ok {
				vless.Query.Host = h
			}
		}
	}

	// 处理 http_opts
	if len(proxy.Http_opts) > 0 {
		if method, ok := proxy.Http_opts["method"].(string); ok {
			vless.Query.Method = method
		}
		if paths, ok := proxy.Http_opts["path"].([]string); ok && len(paths) > 0 {
			vless.Query.Path = paths[0]
		}
		if paths, ok := proxy.Http_opts["path"].([]interface{}); ok && len(paths) > 0 {
			if p, ok := paths[0].(string); ok {
				vless.Query.Path = p
			}
		}
		if headers, ok := proxy.Http_opts["headers"].(map[string]interface{}); ok {
			if hosts, ok := headers["Host"].([]interface{}); ok && len(hosts) > 0 {
				if h, ok := hosts[0].(string); ok {
					vless.Query.Host = h
				}
			}
		}
	}

	// 处理 grpc_opts
	if len(proxy.Grpc_opts) > 0 {
		if sn, ok := proxy.Grpc_opts["grpc-service-name"].(string); ok {
			vless.Query.ServiceName = sn
		}
		if mode, ok := proxy.Grpc_opts["grpc-mode"].(string); ok && mode == "multi" {
			vless.Query.Mode = "multi"
		}
	}

	return vless
}
