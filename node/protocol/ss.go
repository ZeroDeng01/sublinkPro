package protocol

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

// ss匹配规则
type Ss struct {
	Param  Param
	Server string
	Port   interface{}
	Name   string
	Type   string
	Plugin *SsPlugin // SS 插件配置
}
type Param struct {
	Cipher   string
	Password string
}

// SsPlugin SS 插件配置
type SsPlugin struct {
	Name string            // 插件名称：obfs, v2ray-plugin, shadow-tls, restls, kcptun 等
	Opts map[string]string // 插件选项键值对
}

// parseSSURL 解析SS URL，返回认证信息、地址、名称和插件参数
// 支持 SIP002 格式：ss://userinfo@host:port/?plugin=xxx#name
func parseSSURL(s string) (auth, addr, name string, plugin *SsPlugin) {
	u, err := url.Parse(s)
	if err != nil {
		log.Println("ss url parse fail.", err)
		return "", "", "", nil
	}
	if u.Scheme != "ss" {
		log.Println("ss url parse fail, not ss url.")
		return "", "", "", nil
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
				return "", "", "", nil
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

// parseSSPlugin 解析 SIP002 格式的 plugin 参数
// 格式: plugin_name;opt1=val1;opt2=val2
// 特殊字符需要反斜杠转义
func parseSSPlugin(pluginStr string) *SsPlugin {
	if pluginStr == "" {
		return nil
	}

	// SIP003 格式：使用分号分隔，第一个是插件名称
	// 需要处理转义字符
	parts := splitPluginOpts(pluginStr)
	if len(parts) == 0 {
		return nil
	}

	plugin := &SsPlugin{
		Name: parts[0],
		Opts: make(map[string]string),
	}

	// 解析剩余的选项
	for i := 1; i < len(parts); i++ {
		opt := parts[i]
		if idx := strings.Index(opt, "="); idx != -1 {
			key := opt[:idx]
			value := opt[idx+1:]
			plugin.Opts[key] = value
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
	if s.Plugin != nil && s.Plugin.Name != "" {
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
func encodeSSPlugin(plugin *SsPlugin) string {
	if plugin == nil || plugin.Name == "" {
		return ""
	}

	var parts []string
	parts = append(parts, escapePluginValue(plugin.Name))

	// 按固定顺序输出常见选项，保证一致性
	orderedKeys := []string{"mode", "host", "path", "tls", "mux", "password", "version", "version-hint", "restls-script"}
	addedKeys := make(map[string]bool)

	for _, key := range orderedKeys {
		if val, ok := plugin.Opts[key]; ok {
			parts = append(parts, escapePluginValue(key)+"="+escapePluginValue(val))
			addedKeys[key] = true
		}
	}

	// 添加其他未在固定顺序中的选项
	for key, val := range plugin.Opts {
		if !addedKeys[key] {
			parts = append(parts, escapePluginValue(key)+"="+escapePluginValue(val))
		}
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
	param = utils.Base64Decode(param)
	// 判断是否为空
	if param == "" || addr == "" {
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
		fmt.Println("Param:", utils.Base64Decode(param))
		fmt.Println("Server", server)
		fmt.Println("Port", port)
		fmt.Println("Name:", name)
		fmt.Println("Cipher:", cipher)
		fmt.Println("Password:", password)
		if plugin != nil {
			fmt.Println("Plugin:", plugin.Name)
			fmt.Println("Plugin Opts:", plugin.Opts)
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
