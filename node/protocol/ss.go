package protocol

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

func init() {
	base := newProtocolSpec("ss", []string{"ss://"}, "SS", "#2e7d32", "S", Ss{}, "Name", DecodeSSURL, EncodeSSURL, func(s Ss) LinkIdentity {
		return buildIdentity("ss", s.Name, s.Server, utils.GetPortString(s.Port))
	},
		FieldMeta{Name: "Name", Label: "节点名称", Type: "string", Group: "basic", Placeholder: "例如：SS-01"},
		FieldMeta{Name: "Server", Label: "服务器地址", Type: "string", Group: "basic", Placeholder: "example.com"},
		FieldMeta{Name: "Port", Label: "端口", Type: "int", Group: "basic", Placeholder: "8388"},
		FieldMeta{Name: "Param.Cipher", Label: "加密方式", Type: "string", Group: "transport", Options: []string{"aes-128-gcm", "aes-256-gcm", "chacha20-ietf-poly1305", "2022-blake3-aes-128-gcm", "2022-blake3-aes-256-gcm"}},
		FieldMeta{Name: "Param.Password", Label: "密码", Type: "string", Group: "auth", Secret: true, Placeholder: "password"},
		FieldMeta{Name: "Plugin.Name", Label: "插件类型", Type: "string", Group: "transport", Advanced: true, Options: []string{"", "obfs", "obfs-local", "simple-obfs", "v2ray-plugin", "shadow-tls", "restls", "kcptun"}},
		FieldMeta{Name: "Plugin.Mode", Label: "插件模式", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Plugin.Host", Label: "插件 Host", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Plugin.Path", Label: "插件路径", Type: "string", Group: "transport", Advanced: true},
		FieldMeta{Name: "Plugin.Tls", Label: "插件 TLS", Type: "bool", Group: "tls", Advanced: true},
		FieldMeta{Name: "Plugin.Mux", Label: "插件 Mux", Type: "bool", Group: "advanced", Advanced: true},
		FieldMeta{Name: "Plugin.Password", Label: "插件密码", Type: "string", Group: "auth", Secret: true, Advanced: true},
		FieldMeta{Name: "Plugin.Version", Label: "插件版本", Type: "int", Group: "advanced", Advanced: true},
	)
	MustRegisterProtocol(newProxySurgeProtocolSpec(base, buildSSProxy, func(proxy Proxy) bool {
		return proxyTypeMatches(proxy, "ss")
	}, ConvertProxyToSs, EncodeSSURL, buildSSSurgeLine))
}

// ss匹配规则
type Ss struct {
	Param  Param       `json:"param"`
	Server string      `json:"server"`
	Port   interface{} `json:"port"`
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Plugin SsPlugin    `json:"plugin"` // SS 插件配置
}
type Param struct {
	Cipher   string `json:"cipher"`
	Password string `json:"password"`
}

// SsPlugin SS 插件配置
type SsPlugin struct {
	Name     string `json:"name"`     // 插件名称：obfs, v2ray-plugin, shadow-tls, restls, kcptun 等
	Mode     string `json:"mode"`     // 插件模式：http, tls, websocket 等
	Host     string `json:"host"`     // 混淆主机名
	Path     string `json:"path"`     // 路径 (v2ray-plugin)
	Tls      bool   `json:"tls"`      // 是否启用 TLS
	Mux      bool   `json:"mux"`      // 是否启用多路复用
	Password string `json:"password"` // 插件密码 (shadow-tls, restls)
	Version  int    `json:"version"`  // 插件版本 (shadow-tls)
}

// parseSSURL 解析SS URL，返回认证信息、地址、名称和插件参数
// 支持 SIP002 格式：ss://userinfo@host:port/?plugin=xxx#name
func parseSSURL(s string) (auth, addr, name string, plugin SsPlugin) {
	u, err := url.Parse(s)
	if err != nil {
		log.Println("ss url parse fail.", err)
		return "", "", "", SsPlugin{}
	}
	if u.Scheme != "ss" {
		log.Println("ss url parse fail, not ss url.")
		return "", "", "", SsPlugin{}
	}
	// 处理url全编码的情况（整个链接base64编码）
	if u.User == nil {
		// 截取ss://后的字符串，处理可能存在的#标签
		raw := s[5:]
		// 先分离可能的#标签
		hashIndex := strings.LastIndex(raw, "#")
		if hashIndex != -1 {
			name = raw[hashIndex+1:]
			raw = raw[:hashIndex]
		}
		decoded := utils.Base64Decode(raw)
		if decoded != "" {
			s = "ss://" + decoded
			if name != "" {
				s += "#" + name
			}
			u, err = url.Parse(s)
			if err != nil {
				return "", "", "", SsPlugin{}
			}
		}
	}

	if u.User != nil {
		auth = u.User.String()
	}
	if u.Host != "" {
		addr = u.Host
	}
	if u.Fragment != "" {
		name = u.Fragment
	}

	// 解析 plugin 查询参数 (SIP002 格式)
	pluginStr := u.Query().Get("plugin")
	if pluginStr != "" {
		plugin = parseSSPlugin(pluginStr)
	}

	return auth, addr, name, plugin
}

func decodeSSAuth(auth string) (string, error) {
	if auth == "" {
		return "", fmt.Errorf("missing SS auth")
	}

	if decoded := utils.Base64Decode(auth); decoded != auth && strings.Contains(decoded, ":") {
		return decoded, nil
	}

	unescaped, err := url.PathUnescape(auth)
	if err != nil {
		return "", fmt.Errorf("unescape SS auth: %w", err)
	}
	if strings.Contains(unescaped, ":") {
		return unescaped, nil
	}

	return "", fmt.Errorf("invalid SS auth")
}

// parseSSPlugin 解析 SIP002 格式的 plugin 参数
// 格式: plugin_name;opt1=val1;opt2=val2
// 特殊字符需要反斜杠转义
func parseSSPlugin(pluginStr string) SsPlugin {
	if pluginStr == "" {
		return SsPlugin{}
	}

	// SIP003 格式：使用分号分隔，第一个是插件名称
	parts := splitPluginOpts(pluginStr)
	if len(parts) == 0 {
		return SsPlugin{}
	}

	plugin := SsPlugin{
		Name: parts[0],
	}

	// 解析剩余的选项到结构化字段
	for i := 1; i < len(parts); i++ {
		opt := parts[i]
		if idx := strings.Index(opt, "="); idx != -1 {
			key := opt[:idx]
			value := opt[idx+1:]
			switch key {
			case "mode", "obfs":
				plugin.Mode = value
			case "host", "obfs-host":
				plugin.Host = value
			case "path":
				plugin.Path = value
			case "tls":
				plugin.Tls = value == "true" || value == "1" || value == ""
			case "mux":
				plugin.Mux = value == "true" || value == "1"
			case "password":
				plugin.Password = value
			case "version":
				if v, err := strconv.Atoi(value); err == nil {
					plugin.Version = v
				}
			}
		} else if opt == "tls" {
			// 无值的 tls 参数表示启用
			plugin.Tls = true
		}
	}

	return plugin
}

// splitPluginOpts 按分号分隔插件选项，处理反斜杠转义

func splitPluginOpts(s string) []string {
	var result []string
	var current strings.Builder
	escaped := false

	for _, ch := range s {
		if escaped {
			current.WriteRune(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == ';' {
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteRune(ch)
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// 开发者测试
func CallSSURL() {
	ss := Ss{}
	// ss.Name = "测试"
	ss.Server = "baidu.com"
	ss.Port = 443
	ss.Param.Cipher = "2022-blake3-aes-256-gcm"
	ss.Param.Password = "asdasd"
	fmt.Println(EncodeSSURL(ss))
}

// ss 编码输出
// 支持 SIP002 格式：ss://userinfo@host:port/?plugin=xxx#name
func EncodeSSURL(s Ss) string {
	p := utils.Base64Encode(s.Param.Cipher + ":" + s.Param.Password)
	// 假设备注没有使用服务器加端口命名
	if s.Name == "" {
		s.Name = s.Server + ":" + utils.GetPortString(s.Port)
	}

	// 构建基础 URL
	u := url.URL{
		Scheme:   "ss",
		User:     url.User(p),
		Host:     fmt.Sprintf("%s:%s", s.Server, utils.GetPortString(s.Port)),
		Fragment: s.Name,
	}

	// 如果有插件配置，添加 plugin 查询参数
	if s.Plugin.Name != "" {
		q := u.Query()
		q.Set("plugin", encodeSSPlugin(s.Plugin))
		u.RawQuery = q.Encode()
		// 添加路径斜杠（SIP002 规范要求有查询参数时需要）
		u.Path = "/"
	}

	return u.String()
}

// encodeSSPlugin 将插件配置编码为 SIP002 格式字符串
// 格式: plugin_name;opt1=val1;opt2=val2
// 注意：不同插件使用不同的参数名
// - simple-obfs/obfs-local: 使用 obfs 和 obfs-host
// - v2ray-plugin 等: 使用 mode 和 host
func encodeSSPlugin(plugin SsPlugin) string {
	if plugin.Name == "" {
		return ""
	}

	var parts []string
	parts = append(parts, escapePluginValue(plugin.Name))

	// 根据插件类型选择正确的参数名
	isSimpleObfs := plugin.Name == "simple-obfs" || plugin.Name == "obfs-local"

	// 按结构体字段输出选项，根据插件类型使用不同的参数名
	if plugin.Mode != "" {
		if isSimpleObfs {
			parts = append(parts, "obfs="+escapePluginValue(plugin.Mode))
		} else {
			parts = append(parts, "mode="+escapePluginValue(plugin.Mode))
		}
	}
	if plugin.Host != "" {
		if isSimpleObfs {
			parts = append(parts, "obfs-host="+escapePluginValue(plugin.Host))
		} else {
			parts = append(parts, "host="+escapePluginValue(plugin.Host))
		}
	}
	if plugin.Path != "" {
		parts = append(parts, "path="+escapePluginValue(plugin.Path))
	}
	if plugin.Tls {
		parts = append(parts, "tls")
	}
	if plugin.Mux {
		parts = append(parts, "mux=true")
	}
	if plugin.Password != "" {
		parts = append(parts, "password="+escapePluginValue(plugin.Password))
	}
	if plugin.Version > 0 {
		parts = append(parts, fmt.Sprintf("version=%d", plugin.Version))
	}

	return strings.Join(parts, ";")
}

// escapePluginValue 转义插件选项中的特殊字符
func escapePluginValue(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, "=", "\\=")
	return s
}

// DecodeSSURL 解析 SS 链接
// 支持 SIP002 格式：ss://userinfo@host:port/?plugin=xxx#name
func DecodeSSURL(s string) (Ss, error) {
	// 解析ss链接
	param, addr, name, plugin := parseSSURL(s)
	// base64解码
	param, err := decodeSSAuth(param)
	// 判断是否为空
	if err != nil || param == "" || addr == "" {
		return Ss{}, fmt.Errorf("invalid SS URL")
	}
	// 解析参数
	parts := strings.Split(addr, ":")
	port, _ := strconv.Atoi(parts[len(parts)-1])
	server := strings.Replace(utils.UnwrapIPv6Host(addr), ":"+parts[len(parts)-1], "", -1)
	cipher := strings.Split(param, ":")[0]
	password := strings.Replace(param, cipher+":", "", 1)
	// 如果没有备注则使用服务器加端口命名
	if name == "" {
		name = addr
	}
	// 开发环境输出结果
	if utils.CheckEnvironment() {
		fmt.Println("Param:", param)
		fmt.Println("Server", server)
		fmt.Println("Port", port)
		fmt.Println("Name:", name)
		fmt.Println("Cipher:", cipher)
		fmt.Println("Password:", password)
		if plugin.Name != "" {
			fmt.Println("Plugin:", plugin.Name)
			fmt.Println("Plugin Mode:", plugin.Mode)
			fmt.Println("Plugin Host:", plugin.Host)
			fmt.Println("Plugin Path:", plugin.Path)
			fmt.Println("Plugin Tls:", plugin.Tls)
			fmt.Println("Plugin Mux:", plugin.Mux)
			fmt.Println("Plugin Password:", plugin.Password)
			fmt.Println("Plugin Version:", plugin.Version)
		}

	}
	// 返回结果
	return Ss{
		Param: Param{
			Cipher:   cipher,
			Password: password,
		},
		Server: server,
		Port:   port,
		Name:   name,
		Type:   "ss",
		Plugin: plugin,
	}, nil
}

// ConvertProxyToSs 将 Proxy 结构体转换为 Ss 结构体
// 用于从 Clash 格式的代理配置生成 SS 链接
func ConvertProxyToSs(proxy Proxy) Ss {
	ss := Ss{
		Param: Param{
			Cipher:   proxy.Cipher,
			Password: proxy.Password,
		},
		Server: proxy.Server,
		Port:   int(proxy.Port),
		Name:   proxy.Name,
		Type:   "ss",
	}

	// 处理插件信息
	if proxy.Plugin != "" {
		// 插件名称映射：Clash 中的 "obfs" 在 SS 链接中应该是 "simple-obfs"
		pluginName := proxy.Plugin
		if pluginName == "obfs" {
			pluginName = "simple-obfs"
		}
		ss.Plugin = SsPlugin{
			Name: pluginName,
		}
		if proxy.Plugin_opts != nil {
			if mode, ok := proxy.Plugin_opts["mode"].(string); ok {
				ss.Plugin.Mode = mode
			}
			if host, ok := proxy.Plugin_opts["host"].(string); ok {
				ss.Plugin.Host = host
			}
			if path, ok := proxy.Plugin_opts["path"].(string); ok {
				ss.Plugin.Path = path
			}
			if tls, ok := proxy.Plugin_opts["tls"].(bool); ok {
				ss.Plugin.Tls = tls
			}
			if mux, ok := proxy.Plugin_opts["mux"].(bool); ok {
				ss.Plugin.Mux = mux
			}
			if password, ok := proxy.Plugin_opts["password"].(string); ok {
				ss.Plugin.Password = password
			}
			if version, ok := proxy.Plugin_opts["version"].(int); ok {
				ss.Plugin.Version = version
			}
		}
	}

	return ss
}

func buildSSProxy(link Urls, config OutputConfig) (Proxy, error) {
	ss, err := DecodeSSURL(link.Url)
	if err != nil {
		return Proxy{}, err
	}
	if ss.Name == "" {
		ss.Name = fmt.Sprintf("%s:%s", ss.Server, utils.GetPortString(ss.Port))
	}
	proxy := Proxy{
		Name:             ss.Name,
		Type:             "ss",
		Server:           ss.Server,
		Port:             FlexPort(utils.GetPortInt(ss.Port)),
		Cipher:           ss.Param.Cipher,
		Password:         ss.Param.Password,
		Udp:              config.Udp,
		Skip_cert_verify: config.Cert,
		Dialer_proxy:     link.DialerProxyName,
	}
	if ss.Plugin.Name != "" {
		pluginName := ss.Plugin.Name
		if pluginName == "simple-obfs" || pluginName == "obfs-local" {
			pluginName = "obfs"
		}
		proxy.Plugin = pluginName
		proxy.Plugin_opts = convertSSPluginOpts(ss.Plugin)
	}
	return proxy, nil
}

func buildSSSurgeLine(link string, config OutputConfig) (string, string, error) {
	ss, err := DecodeSSURL(link)
	if err != nil {
		return "", "", err
	}
	server := replaceSurgeHost(ss.Server, config)
	line := fmt.Sprintf("%s = ss, %s, %d, encrypt-method=%s, password=%s, udp-relay=%t", ss.Name, server, utils.GetPortInt(ss.Port), ss.Param.Cipher, ss.Param.Password, config.Udp)
	if ss.Plugin.Name != "" {
		line = appendSurgeSSPlugin(line, ss.Plugin)
	}
	return line, ss.Name, nil
}
