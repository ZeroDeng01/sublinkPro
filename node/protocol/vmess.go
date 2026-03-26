package protocol

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sublink/utils"
)

func init() {
	base := newProtocolSpec("vmess", []string{"vmess://"}, "VMess", "#1976d2", "V", Vmess{}, "Ps", DecodeVMESSURL, EncodeVmessURL, func(v Vmess) LinkIdentity {
		return buildIdentity("vmess", v.Ps, v.Host, utils.GetPortString(v.Port))
	},
		FieldMeta{Name: "Ps", Label: "节点名称", Type: "string", Group: "basic", Placeholder: "例如：香港-01"},
		FieldMeta{Name: "Add", Label: "服务器地址", Type: "string", Group: "basic", Placeholder: "example.com"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic", Placeholder: "443"},
		FieldMeta{Name: "Id", Label: "UUID", Type: "string", Group: "auth", Placeholder: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"},
		FieldMeta{Name: "Aid", Label: "Alter ID", Type: "int", Group: "auth", Advanced: true},
		FieldMeta{Name: "Scy", Label: "加密方式", Type: "string", Group: "transport", Options: []string{"auto", "aes-128-gcm", "aes-256-gcm", "chacha20-poly1305", "none"}},
		FieldMeta{Name: "Net", Label: "传输层", Type: "string", Group: "transport", Options: []string{"tcp", "ws", "http", "grpc", "h2", "quic"}},
		FieldMeta{Name: "Path", Label: "路径", Type: "string", Group: "transport", Placeholder: "/ws"},
		FieldMeta{Name: "Host", Label: "Host / WS Host", Type: "string", Group: "transport", Placeholder: "cdn.example.com"},
		FieldMeta{Name: "Type", Label: "Header Type", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Tls", Label: "TLS", Type: "string", Group: "tls", Options: []string{"", "tls", "none"}},
		FieldMeta{Name: "Sni", Label: "SNI", Type: "string", Group: "tls", Placeholder: "server.example.com"},
		FieldMeta{Name: "Alpn", Label: "ALPN", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "Fp", Label: "指纹", Type: "string", Group: "tls", Advanced: true},
		FieldMeta{Name: "V", Label: "协议版本", Type: "string", Group: "advanced", Advanced: true},
	)
	MustRegisterProtocol(newProxySurgeProtocolSpec(base, buildVMessProxy, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "vmess")
	}, ConvertProxyToVmess, EncodeVmessURL, buildVMessSurgeLine))
}

type Vmess struct {
	Add  string      `json:"add,omitempty"` // 服务器地址
	Aid  interface{} `json:"aid,omitempty"`
	Alpn string      `json:"alpn,omitempty"`
	Fp   string      `json:"fp,omitempty"`
	Host string      `json:"host,omitempty"`
	Id   string      `json:"id,omitempty"`
	Net  string      `json:"net,omitempty"`
	Path string      `json:"path,omitempty"`
	Port interface{} `json:"port,omitempty"`
	Ps   string      `json:"ps,omitempty"`
	Scy  string      `json:"scy,omitempty"`
	Sni  string      `json:"sni,omitempty"`
	Tls  string      `json:"tls,omitempty"`
	Type string      `json:"type,omitempty"`
	V    string      `json:"v,omitempty"`
}

// EncodeVmessURL 将 VMess 结构编码为标准 vmess:// base64 JSON 链接。
// 当备注或协议版本缺失时，会按当前实现补默认值，保证导出结果可被后续流程复用。
func EncodeVmessURL(v Vmess) string {
	// 如果备注为空，则使用服务器地址+端口
	if v.Ps == "" {
		v.Ps = v.Add + ":" + utils.GetPortString(v.Port)
	}
	// 如果版本为空，则默认为2
	if v.V == "" {
		v.V = "2"
	}
	param, _ := json.Marshal(v)
	return "vmess://" + utils.Base64Encode(string(param))
}

// DecodeVMESSURL 解析 vmess:// 链接并补齐运行时依赖的默认字段。
// 这里会校验 UUID、兼容 IPv6 主机写法，并在缺少 cipher 或备注时填入当前约定的默认值。
func DecodeVMESSURL(s string) (Vmess, error) {
	if !strings.Contains(s, "vmess://") {
		return Vmess{}, fmt.Errorf("非vmess协议:%s", s)
	}
	param := strings.Split(s, "://")[1]
	param = utils.Base64Decode(strings.TrimSpace(param))
	// fmt.Println(param)
	var vmess Vmess
	err := json.Unmarshal([]byte(param), &vmess)
	if err != nil {
		utils.Error("❌节点解析错误：%v  【节点：%s】", err, param)
		return Vmess{}, fmt.Errorf("json格式化失败:%s", param)
	}
	if !utils.IsUUID(vmess.Id) {
		utils.Error("❌节点解析错误：%v  【节点：%s】", "UUID格式错误", param)
		return Vmess{}, fmt.Errorf("uuid格式错误:%s", vmess.Id)
	}
	vmess.Add = utils.UnwrapIPv6Host(vmess.Add)
	if vmess.Scy == "" {
		vmess.Scy = "auto"
	}
	// 如果备注为空，则使用服务器地址+端口
	if vmess.Ps == "" {
		vmess.Ps = vmess.Add + ":" + utils.GetPortString(vmess.Port)
	}
	if utils.CheckEnvironment() {
		fmt.Println("服务器地址", vmess.Add)
		fmt.Println("端口", vmess.Port)
		fmt.Println("path", vmess.Path)
		fmt.Println("uuid", vmess.Id)
		fmt.Println("alterId", vmess.Aid)
		fmt.Println("cipher", vmess.Scy)
		fmt.Println("client-fingerprint", vmess.Fp)
		fmt.Println("network", vmess.Net)
		fmt.Println("tls", vmess.Tls)
		fmt.Println("备注", vmess.Ps)
	}
	return vmess, nil
}

// ConvertProxyToVmess 将 Proxy 结构体转换为 Vmess 结构体
// 用于从 Clash 格式的代理配置生成 VMess 链接
func ConvertProxyToVmess(proxy Proxy) Vmess {
	vmess := Vmess{
		Add:  proxy.Server,
		Port: int(proxy.Port),
		Id:   proxy.Uuid,
		Ps:   proxy.Name,
		Scy:  proxy.Cipher,
		Net:  proxy.Network,
		V:    "2",
	}

	// 处理 alterId
	if proxy.AlterId != "" {
		if aid, err := strconv.Atoi(proxy.AlterId); err == nil {
			vmess.Aid = aid
		}
	} else {
		vmess.Aid = 0
	}

	// 处理 TLS
	if proxy.Tls {
		vmess.Tls = "tls"
	} else {
		vmess.Tls = "none"
	}

	// 处理 ws_opts
	if len(proxy.Ws_opts) > 0 {
		if path, ok := proxy.Ws_opts["path"].(string); ok {
			vmess.Path = path
		}
		if headers, ok := proxy.Ws_opts["headers"].(map[string]interface{}); ok {
			if host, ok := headers["Host"].(string); ok {
				vmess.Host = host
			}
		}
	}

	// 处理 Sni
	if proxy.Sni != "" {
		vmess.Sni = proxy.Sni
	} else if proxy.Servername != "" {
		vmess.Sni = proxy.Servername
	}

	// 处理 alpn
	if len(proxy.Alpn) > 0 {
		vmess.Alpn = strings.Join(proxy.Alpn, ",")
	}

	// 处理客户端指纹
	vmess.Fp = proxy.Client_fingerprint

	return vmess
}

// buildVMessProxy 将 VMess 链接转换为 Clash Proxy，并应用输出阶段的 UDP、证书校验和前置代理覆盖项。
func buildVMessProxy(link Urls, config OutputConfig) (Proxy, error) {
	vmess, err := DecodeVMESSURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	if vmess.Ps == "" {
		vmess.Ps = fmt.Sprintf("%s:%s", vmess.Add, utils.GetPortString(vmess.Port))
	}
	wsOpts := map[string]interface{}{"path": vmess.Path, "headers": map[string]interface{}{"Host": vmess.Host}}
	DeleteOpts(wsOpts)
	tls := vmess.Tls != "none" && vmess.Tls != ""
	port, _ := convertToInt(vmess.Port)
	aid, _ := convertToInt(vmess.Aid)
	return Proxy{Name: vmess.Ps, Type: "vmess", Server: vmess.Add, Port: FlexPort(port), Cipher: vmess.Scy, Uuid: vmess.Id, AlterId: strconv.Itoa(aid), Network: vmess.Net, Tls: tls, Ws_opts: wsOpts, Udp: config.Udp, Skip_cert_verify: config.Cert, Dialer_proxy: link.DialerProxyName}, nil
}

// buildVMessSurgeLine 将 VMess 链接转换为 Surge 节点行。
// Surge 导出只保留当前实现支持的传输层和 TLS 字段，其余能力不会在此处完整保真。
func buildVMessSurgeLine(link string, config OutputConfig) (string, string, error) {
	vmess, err := DecodeVMESSURL(link)
	if err != nil {
		return "", "", err
	}
	tls := vmess.Tls != "none" && vmess.Tls != ""
	port, _ := convertToInt(vmess.Port)
	server := replaceSurgeHost(vmess.Add, config)
	line := fmt.Sprintf("%s = vmess, %s, %d, username=%s , tls=%t, vmess-aead=true,  udp-relay=%t , skip-cert-verify=%t", vmess.Ps, server, port, vmess.Id, tls, config.Udp, config.Cert)
	if vmess.Net == "ws" {
		line = fmt.Sprintf("%s, ws=true,ws-path=%s", line, vmess.Path)
		if vmess.Host != "" && vmess.Host != "none" {
			line = fmt.Sprintf("%s, ws-headers=Host:%s", line, vmess.Host)
		}
	}
	if vmess.Sni != "" {
		line = fmt.Sprintf("%s, sni=%s", line, vmess.Sni)
	}
	return line, vmess.Ps, nil
}
