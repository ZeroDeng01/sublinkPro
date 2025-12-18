package protocol

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sublink/cache"
	"sublink/utils"

	"gopkg.in/yaml.v3"
)

type Proxy struct {
	Name               string                 `yaml:"name,omitempty"`               // 节点名称
	Type               string                 `yaml:"type,omitempty"`               // 代理类型 (ss, vmess, trojan, etc.)
	Server             string                 `yaml:"server,omitempty"`             // 服务器地址
	Port               int                    `yaml:"port,omitempty"`               // 服务器端口
	Cipher             string                 `yaml:"cipher,omitempty"`             // 加密方式
	Username           string                 `yaml:"username,omitempty"`           // 用户名 (socks5 等)
	Password           string                 `yaml:"password,omitempty"`           // 密码
	Client_fingerprint string                 `yaml:"client-fingerprint,omitempty"` // 客户端指纹 (uTLS)
	Tfo                bool                   `yaml:"tfo,omitempty"`                // TCP Fast Open
	Udp                bool                   `yaml:"udp,omitempty"`                // 是否启用 UDP
	Skip_cert_verify   bool                   `yaml:"skip-cert-verify,omitempty"`   // 跳过证书验证
	Tls                bool                   `yaml:"tls,omitempty"`                // 是否启用 TLS
	Servername         string                 `yaml:"servername,omitempty"`         // TLS SNI
	Flow               string                 `yaml:"flow,omitempty"`               // 流控 (xtls-rprx-vision 等)
	AlterId            string                 `yaml:"alterId,omitempty"`            // VMess AlterId
	Network            string                 `yaml:"network,omitempty"`            // 传输协议 (ws, grpc, etc.)
	Reality_opts       map[string]interface{} `yaml:"reality-opts,omitempty"`       // Reality 选项
	Ws_opts            map[string]interface{} `yaml:"ws-opts,omitempty"`            // WebSocket 选项
	Grpc_opts          map[string]interface{} `yaml:"grpc-opts,omitempty"`          // gRPC 选项
	Auth_str           string                 `yaml:"auth_str,omitempty"`           // Hysteria 认证字符串
	Auth               string                 `yaml:"auth,omitempty"`               // 认证信息
	Up                 int                    `yaml:"up,omitempty"`                 // 上行带宽限制
	Down               int                    `yaml:"down,omitempty"`               // 下行带宽限制
	Alpn               []string               `yaml:"alpn,omitempty"`               // ALPN
	Sni                string                 `yaml:"sni,omitempty"`                // SNI
	Obfs               string                 `yaml:"obfs,omitempty"`               // 混淆模式 (SSR/Hysteria2)
	Obfs_password      string                 `yaml:"obfs-password,omitempty"`      // 混淆密码
	Protocol           string                 `yaml:"protocol,omitempty"`           // SSR 协议
	Uuid               string                 `yaml:"uuid,omitempty"`               // UUID (VMess/VLESS)
	Peer               string                 `yaml:"peer,omitempty"`               // Peer (Hysteria)
	Congestion_control string                 `yaml:"congestion_control,omitempty"` // 拥塞控制 (Tuic)
	Udp_relay_mode     string                 `yaml:"udp_relay_mode,omitempty"`     // UDP 转发模式 (Tuic)
	Disable_sni        bool                   `yaml:"disable_sni,omitempty"`        // 禁用 SNI (Tuic)
	Dialer_proxy       string                 `yaml:"dialer-proxy,omitempty"`       // 前置代理
}

type ProxyGroup struct {
	Proxies []string `yaml:"proxies"`
}
type Config struct {
	Proxies      []Proxy      `yaml:"proxies"`
	Proxy_groups []ProxyGroup `yaml:"proxy-groups"`
}

// 代理链接的结构体
type Urls struct {
	Url             string
	DialerProxyName string
}

// 删除opts中的空值
func DeleteOpts(opts map[string]interface{}) {
	for k, v := range opts {
		switch v := v.(type) {
		case string:
			if v == "" {
				delete(opts, k)
			}
		case map[string]interface{}:
			DeleteOpts(v)
			if len(v) == 0 {
				delete(opts, k)
			}
		}
	}
}
func convertToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("unexpected type %T", v)
	}
}

// LinkToProxy 将单个节点链接转换为 Proxy 结构体
// 支持 ss, ssr, trojan, vmess, vless, hysteria, hysteria2, tuic, anytls, socks5 等协议
func LinkToProxy(link Urls, sqlconfig utils.SqlConfig) (Proxy, error) {
	Scheme := strings.ToLower(strings.Split(link.Url, "://")[0])
	switch {
	case Scheme == "ss":
		ss, err := DecodeSSURL(link.Url)
		if err != nil {
			return Proxy{}, err
		}
		// 如果没有名字，就用服务器地址作为名字
		if ss.Name == "" {
			ss.Name = fmt.Sprintf("%s:%d", ss.Server, ss.Port)
		}
		return Proxy{
			Name:             ss.Name,
			Type:             "ss",
			Server:           ss.Server,
			Port:             ss.Port,
			Cipher:           ss.Param.Cipher,
			Password:         ss.Param.Password,
			Udp:              sqlconfig.Udp,
			Skip_cert_verify: sqlconfig.Cert,
			Dialer_proxy:     link.DialerProxyName,
		}, nil
	case Scheme == "ssr":
		ssr, err := DecodeSSRURL(link.Url)
		if err != nil {
			return Proxy{}, err
		}
		// 如果没有名字，就用服务器地址作为名字
		if ssr.Qurey.Remarks == "" {
			ssr.Qurey.Remarks = fmt.Sprintf("%s:%d", ssr.Server, ssr.Port)
		}
		return Proxy{
			Name:             ssr.Qurey.Remarks,
			Type:             "ssr",
			Server:           ssr.Server,
			Port:             ssr.Port,
			Cipher:           ssr.Method,
			Password:         ssr.Password,
			Obfs:             ssr.Obfs,
			Obfs_password:    ssr.Qurey.Obfsparam,
			Protocol:         ssr.Protocol,
			Udp:              sqlconfig.Udp,
			Skip_cert_verify: sqlconfig.Cert,
			Dialer_proxy:     link.DialerProxyName,
		}, nil
	case Scheme == "trojan":
		trojan, err := DecodeTrojanURL(link.Url)
		if err != nil {
			return Proxy{}, err
		}
		// 如果没有名字，就用服务器地址作为名字
		if trojan.Name == "" {
			trojan.Name = fmt.Sprintf("%s:%d", trojan.Hostname, trojan.Port)
		}
		ws_opts := map[string]interface{}{
			"path": trojan.Query.Path,
			"headers": map[string]interface{}{
				"Host": trojan.Query.Host,
			},
		}
		DeleteOpts(ws_opts)
		return Proxy{
			Name:               trojan.Name,
			Type:               "trojan",
			Server:             trojan.Hostname,
			Port:               trojan.Port,
			Password:           trojan.Password,
			Client_fingerprint: trojan.Query.Fp,
			Sni:                trojan.Query.Sni,
			Network:            trojan.Query.Type,
			Flow:               trojan.Query.Flow,
			Alpn:               trojan.Query.Alpn,
			Ws_opts:            ws_opts,
			Udp:                sqlconfig.Udp,
			Skip_cert_verify:   sqlconfig.Cert,
			Dialer_proxy:       link.DialerProxyName,
		}, nil
	case Scheme == "vmess":
		vmess, err := DecodeVMESSURL(link.Url)
		if err != nil {
			return Proxy{}, err
		}
		// 如果没有名字，就用服务器地址作为名字
		if vmess.Ps == "" {
			vmess.Ps = fmt.Sprintf("%s:%s", vmess.Add, vmess.Port)
		}
		ws_opts := map[string]interface{}{
			"path": vmess.Path,
			"headers": map[string]interface{}{
				"Host": vmess.Host,
			},
		}
		DeleteOpts(ws_opts)
		tls := false
		if vmess.Tls != "none" && vmess.Tls != "" {
			tls = true
		}
		port, _ := convertToInt(vmess.Port)
		aid, _ := convertToInt(vmess.Aid)
		return Proxy{
			Name:             vmess.Ps,
			Type:             "vmess",
			Server:           vmess.Add,
			Port:             port,
			Cipher:           vmess.Scy,
			Uuid:             vmess.Id,
			AlterId:          strconv.Itoa(aid),
			Network:          vmess.Net,
			Tls:              tls,
			Ws_opts:          ws_opts,
			Udp:              sqlconfig.Udp,
			Skip_cert_verify: sqlconfig.Cert,
			Dialer_proxy:     link.DialerProxyName,
		}, nil
	case Scheme == "vless":
		vless, err := DecodeVLESSURL(link.Url)
		if err != nil {
			return Proxy{}, err
		}
		// 如果没有名字，就用服务器地址作为名字
		if vless.Name == "" {
			vless.Name = fmt.Sprintf("%s:%d", vless.Server, vless.Port)
		}
		ws_opts := map[string]interface{}{
			"path": vless.Query.Path,
			"headers": map[string]interface{}{
				"Host": vless.Query.Host,
			},
		}
		reality_opts := map[string]interface{}{
			"public-key": vless.Query.Pbk,
			"short-id":   vless.Query.Sid,
		}
		grpc_opts := map[string]interface{}{
			"grpc-mode":         "gun",
			"grpc-service-name": vless.Query.ServiceName,
		}
		if vless.Query.Mode == "multi" {
			grpc_opts["grpc-mode"] = "multi"
		}
		DeleteOpts(ws_opts)
		DeleteOpts(reality_opts)
		DeleteOpts(grpc_opts)
		tls := false
		if vless.Query.Security != "" {
			tls = true
		}
		if vless.Query.Security == "none" {
			tls = false
		}
		return Proxy{
			Name:               vless.Name,
			Type:               "vless",
			Server:             vless.Server,
			Port:               vless.Port,
			Servername:         vless.Query.Sni,
			Uuid:               vless.Uuid,
			Client_fingerprint: vless.Query.Fp,
			Network:            vless.Query.Type,
			Flow:               vless.Query.Flow,
			Alpn:               vless.Query.Alpn,
			Ws_opts:            ws_opts,
			Reality_opts:       reality_opts,
			Grpc_opts:          grpc_opts,
			Udp:                sqlconfig.Udp,
			Skip_cert_verify:   sqlconfig.Cert,
			Tls:                tls,
			Dialer_proxy:       link.DialerProxyName,
		}, nil
	case Scheme == "hy" || Scheme == "hysteria":
		hy, err := DecodeHYURL(link.Url)
		if err != nil {
			return Proxy{}, err
		}
		// 如果没有名字，就用服务器地址作为名字
		if hy.Name == "" {
			hy.Name = fmt.Sprintf("%s:%d", hy.Host, hy.Port)
		}
		return Proxy{
			Name:             hy.Name,
			Type:             "hysteria",
			Server:           hy.Host,
			Port:             hy.Port,
			Auth_str:         hy.Auth,
			Up:               hy.UpMbps,
			Down:             hy.DownMbps,
			Alpn:             hy.ALPN,
			Peer:             hy.Peer,
			Udp:              sqlconfig.Udp,
			Skip_cert_verify: sqlconfig.Cert,
			Dialer_proxy:     link.DialerProxyName,
		}, nil
	case Scheme == "hy2" || Scheme == "hysteria2":
		hy2, err := DecodeHY2URL(link.Url)
		if err != nil {
			return Proxy{}, err
		}
		// 如果没有名字，就用服务器地址作为名字
		if hy2.Name == "" {
			hy2.Name = fmt.Sprintf("%s:%d", hy2.Host, hy2.Port)
		}
		return Proxy{
			Name:             hy2.Name,
			Type:             "hysteria2",
			Server:           hy2.Host,
			Port:             hy2.Port,
			Auth_str:         hy2.Auth,
			Sni:              hy2.Sni,
			Alpn:             hy2.ALPN,
			Obfs:             hy2.Obfs,
			Password:         hy2.Password,
			Obfs_password:    hy2.ObfsPassword,
			Udp:              sqlconfig.Udp,
			Skip_cert_verify: sqlconfig.Cert,
			Dialer_proxy:     link.DialerProxyName,
		}, nil
	case Scheme == "tuic":
		tuic, err := DecodeTuicURL(link.Url)
		if err != nil {
			return Proxy{}, err
		}
		// 如果没有名字，就用服务器地址作为名字
		if tuic.Name == "" {
			tuic.Name = fmt.Sprintf("%s:%d", tuic.Host, tuic.Port)
		}
		disable_sni := false
		if tuic.Disable_sni == 1 {
			disable_sni = true
		}
		return Proxy{
			Name:               tuic.Name,
			Type:               "tuic",
			Server:             tuic.Host,
			Port:               tuic.Port,
			Password:           tuic.Password,
			Uuid:               tuic.Uuid,
			Congestion_control: tuic.Congestion_control,
			Alpn:               tuic.Alpn,
			Udp_relay_mode:     tuic.Udp_relay_mode,
			Disable_sni:        disable_sni,
			Sni:                tuic.Sni,
			Udp:                sqlconfig.Udp,
			Skip_cert_verify:   sqlconfig.Cert,
			Dialer_proxy:       link.DialerProxyName,
		}, nil

	case Scheme == "anytls":
		anyTLS, err := DecodeAnyTLSURL(link.Url)
		if err != nil {
			return Proxy{}, err
		}
		return Proxy{
			Name:               anyTLS.Name,
			Type:               "anytls",
			Server:             anyTLS.Server,
			Port:               anyTLS.Port,
			Password:           anyTLS.Password,
			Skip_cert_verify:   anyTLS.SkipCertVerify,
			Sni:                anyTLS.SNI,
			Client_fingerprint: anyTLS.ClientFingerprint,
			Dialer_proxy:       link.DialerProxyName,
		}, nil
	case Scheme == "socks5":
		socks5, err := DecodeSocks5URL(link.Url)
		if err != nil {
			return Proxy{}, err
		}
		return Proxy{
			Name:         socks5.Name,
			Type:         "socks5",
			Server:       socks5.Server,
			Port:         socks5.Port,
			Username:     socks5.Username,
			Password:     socks5.Password,
			Dialer_proxy: link.DialerProxyName,
		}, nil
	default:
		return Proxy{}, fmt.Errorf("unsupported scheme: %s", Scheme)
	}
}

// EncodeClash 用于生成 Clash 配置文件
// 输入: 节点链接列表, SQL配置
// 输出: Clash 配置文件的 YAML 字节流
func EncodeClash(urls []Urls, sqlconfig utils.SqlConfig) ([]byte, error) {
	// 传入urls，解析urls，生成proxys
	// yamlfile 为模板文件
	var proxys []Proxy

	for _, link := range urls {
		proxy, err := LinkToProxy(link, sqlconfig)
		if err != nil {
			utils.Error(err.Error())
			continue
		}
		proxys = append(proxys, proxy)
	}

	// 根据配置执行 Host 替换
	if sqlconfig.ReplaceServerWithHost && len(sqlconfig.HostMap) > 0 {
		for i := range proxys {
			if ip, exists := sqlconfig.HostMap[proxys[i].Server]; exists {
				proxys[i].Server = ip
			}
		}
	}

	// 生成Clash配置文件
	return DecodeClash(proxys, sqlconfig.Clash)
}

// DecodeClash 用于解析 Clash 配置文件并合并新节点
// proxys: 新增的节点列表
// yamlfile: 模板文件路径或 URL
func DecodeClash(proxys []Proxy, yamlfile string) ([]byte, error) {
	// 读取 YAML 文件
	var data []byte
	var err error
	if strings.Contains(yamlfile, "://") {
		resp, err := http.Get(yamlfile)
		if err != nil {
			utils.Error("http.Get error: %v", err)
			return nil, err
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			utils.Error("error: %v", err)
			return nil, err
		}
	} else {
		// 优先从缓存读取模板内容（本地文件使用缓存）
		filename := filepath.Base(yamlfile)
		if cached, ok := cache.GetTemplateContent(filename); ok {
			data = []byte(cached)
		} else {
			data, err = os.ReadFile(yamlfile)
			if err != nil {
				utils.Error("error: %v", err)
				return nil, err
			}
			// 写入缓存
			cache.SetTemplateContent(filename, string(data))
		}
	}
	// 解析 YAML 文件
	config := make(map[interface{}]interface{})
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		utils.Error("error: %v", err)
		return nil, err
	}

	// 检查 "proxies" 键是否存在于 config 中
	proxies, ok := config["proxies"].([]interface{})
	if !ok {
		// 如果 "proxies" 键不存在，创建一个新的切片
		proxies = []interface{}{}
	}
	// 定义一个代理列表名字
	ProxiesNameList := []string{}
	// 添加新代理
	for _, p := range proxys {
		ProxiesNameList = append(ProxiesNameList, p.Name)
		proxies = append(proxies, p)
	}
	// proxies = append(proxies, newProxy)
	config["proxies"] = proxies
	// 往ProxyGroup中插入代理列表
	proxyGroups := config["proxy-groups"].([]interface{})
	for i, pg := range proxyGroups {
		proxyGroup, ok := pg.(map[string]interface{})
		if !ok {
			continue
		}

		// 链式代理不处理
		if proxyGroup["type"] == "relay" {
			continue
		}

		// 如果已有 include-all: true，说明使用 filter 模式，跳过节点插入
		if includeAll, ok := proxyGroup["include-all"].(bool); ok && includeAll {
			continue
		}

		// 获取现有的 proxies 列表
		var existingProxies []interface{}
		if proxyGroup["proxies"] != nil {
			existingProxies, _ = proxyGroup["proxies"].([]interface{})
		}

		// 合并现有代理和新节点
		var validProxies []interface{}
		for _, p := range existingProxies {
			if p != nil {
				validProxies = append(validProxies, p)
			}
		}
		for _, newProxy := range ProxiesNameList {
			validProxies = append(validProxies, newProxy)
		}

		// 如果代理组为空，插入 DIRECT 作为后备
		if len(validProxies) == 0 {
			validProxies = append(validProxies, "DIRECT")
		}

		proxyGroup["proxies"] = validProxies
		proxyGroups[i] = proxyGroup
	}

	config["proxy-groups"] = proxyGroups

	// 将修改后的内容写回文件
	newData, err := yaml.Marshal(config)
	if err != nil {
		utils.Error("error: %v", err)
	}
	return newData, nil
}
