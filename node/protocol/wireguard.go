package protocol

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sublink/utils"
)

// WireGuard 结构体，存储 WireGuard 节点信息
type WireGuard struct {
	Name       string      `json:"name"`       // 节点名称
	Server     string      `json:"server"`     // 服务器地址
	Port       interface{} `json:"port"`       // 端口
	PrivateKey string      `json:"privateKey"` // 客户端私钥
	PublicKey  string      `json:"publicKey"`  // 服务端公钥
	IP         string      `json:"ip"`         // 客户端 IPv4 地址
	IPv6       string      `json:"ipv6"`       // 客户端 IPv6 地址（可选）
	MTU        int         `json:"mtu"`        // MTU 值（可选，默认 1280）
	Reserved   []int       `json:"reserved"`   // 保留字段（可选，用于 WARP）
	DNS        string      `json:"dns"`        // DNS 服务器（可选）
}

// DecodeWireGuardURL 解析 WireGuard URL
// 格式: wireguard://PrivateKey@Server:Port/?publickey=xxx&address=xxx&mtu=xxx#Name
func DecodeWireGuardURL(s string) (WireGuard, error) {
	u, err := url.Parse(s)
	if err != nil {
		return WireGuard{}, fmt.Errorf("解析失败的URL: %s", s)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "wireguard" && scheme != "wg" {
		return WireGuard{}, fmt.Errorf("非WireGuard协议: %s", s)
	}

	// 提取私钥（URL 用户部分）
	privateKey := u.User.Username()
	if privateKey == "" {
		return WireGuard{}, fmt.Errorf("缺少私钥: %s", s)
	}

	// 服务器和端口
	server := u.Hostname()
	rawPort := u.Port()
	if rawPort == "" {
		rawPort = "51820" // WireGuard 默认端口
	}
	port, _ := strconv.Atoi(rawPort)

	// 查询参数
	publicKey := u.Query().Get("publickey")
	if publicKey == "" {
		publicKey = u.Query().Get("public-key")
	}
	if publicKey == "" {
		return WireGuard{}, fmt.Errorf("缺少公钥: %s", s)
	}

	// 解析地址（可能包含 IPv4 和 IPv6，逗号分隔）
	address := u.Query().Get("address")
	var ip, ipv6 string
	if address != "" {
		addresses := strings.Split(address, ",")
		for _, addr := range addresses {
			addr = strings.TrimSpace(addr)
			// 移除 CIDR 后缀
			addrOnly := strings.Split(addr, "/")[0]
			if strings.Contains(addrOnly, ":") {
				// IPv6
				ipv6 = addrOnly
			} else {
				// IPv4
				ip = addrOnly
			}
		}
	}

	// MTU
	mtu := 0
	if mtuStr := u.Query().Get("mtu"); mtuStr != "" {
		mtu, _ = strconv.Atoi(mtuStr)
	}

	// Reserved 保留字段
	var reserved []int
	if reservedStr := u.Query().Get("reserved"); reservedStr != "" {
		parts := strings.Split(reservedStr, ",")
		for _, p := range parts {
			if v, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				reserved = append(reserved, v)
			}
		}
	}

	// DNS
	dns := u.Query().Get("dns")

	// 节点名称（Fragment）
	name := u.Fragment
	if name == "" {
		name = fmt.Sprintf("%s:%d", server, port)
	}

	if utils.CheckEnvironment() {
		fmt.Println("WireGuard解析结果:")
		fmt.Println("  name:", name)
		fmt.Println("  server:", server)
		fmt.Println("  port:", port)
		fmt.Println("  privateKey:", privateKey)
		fmt.Println("  publicKey:", publicKey)
		fmt.Println("  ip:", ip)
		fmt.Println("  ipv6:", ipv6)
		fmt.Println("  mtu:", mtu)
		fmt.Println("  reserved:", reserved)
	}

	return WireGuard{
		Name:       name,
		Server:     server,
		Port:       port,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		IP:         ip,
		IPv6:       ipv6,
		MTU:        mtu,
		Reserved:   reserved,
		DNS:        dns,
	}, nil
}

// EncodeWireGuardURL 将 WireGuard 结构体编码为 URL
func EncodeWireGuardURL(wg WireGuard) string {
	u := url.URL{
		Scheme:   "wireguard",
		User:     url.User(wg.PrivateKey),
		Host:     fmt.Sprintf("%s:%s", wg.Server, utils.GetPortString(wg.Port)),
		Fragment: wg.Name,
	}

	q := u.Query()
	if wg.PublicKey != "" {
		q.Set("publickey", wg.PublicKey)
	}

	// 组装地址
	var addresses []string
	if wg.IP != "" {
		addresses = append(addresses, wg.IP+"/32")
	}
	if wg.IPv6 != "" {
		addresses = append(addresses, wg.IPv6+"/128")
	}
	if len(addresses) > 0 {
		q.Set("address", strings.Join(addresses, ","))
	}

	if wg.MTU > 0 {
		q.Set("mtu", strconv.Itoa(wg.MTU))
	}

	if len(wg.Reserved) > 0 {
		var parts []string
		for _, v := range wg.Reserved {
			parts = append(parts, strconv.Itoa(v))
		}
		q.Set("reserved", strings.Join(parts, ","))
	}

	if wg.DNS != "" {
		q.Set("dns", wg.DNS)
	}

	u.RawQuery = q.Encode()

	// 如果没有名称，使用服务器:端口
	if wg.Name == "" {
		u.Fragment = fmt.Sprintf("%s:%s", wg.Server, utils.GetPortString(wg.Port))
	}

	return u.String()
}

// IsWireGuardConfig 检查输入是否为 WireGuard 标准配置文件格式
func IsWireGuardConfig(input string) bool {
	return strings.Contains(input, "[Interface]") && strings.Contains(input, "[Peer]")
}

// ParseWireGuardConfig 解析标准 WireGuard 配置文件（INI 格式）
// 支持格式:
// [Interface]
// Address = 172.16.0.2/32
// PrivateKey = xxx
// DNS = 1.1.1.1
// MTU = 1280
// [Peer]
// AllowedIPs = 0.0.0.0/0
// Endpoint = server:port
// PublicKey = xxx
func ParseWireGuardConfig(config string) (WireGuard, error) {
	if !IsWireGuardConfig(config) {
		return WireGuard{}, fmt.Errorf("不是有效的 WireGuard 配置文件")
	}

	wg := WireGuard{}
	lines := strings.Split(config, "\n")
	currentSection := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 检测段落
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}

		// 解析键值对
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch currentSection {
		case "interface":
			switch key {
			case "address":
				// 可能包含多个地址，逗号分隔
				addresses := strings.Split(value, ",")
				for _, addr := range addresses {
					addr = strings.TrimSpace(addr)
					addrOnly := strings.Split(addr, "/")[0]
					if strings.Contains(addrOnly, ":") {
						wg.IPv6 = addrOnly
					} else {
						wg.IP = addrOnly
					}
				}
			case "privatekey":
				wg.PrivateKey = value
			case "dns":
				wg.DNS = value
			case "mtu":
				wg.MTU, _ = strconv.Atoi(value)
			}
		case "peer":
			switch key {
			case "endpoint":
				// 格式: server:port
				if idx := strings.LastIndex(value, ":"); idx != -1 {
					wg.Server = value[:idx]
					if port, err := strconv.Atoi(value[idx+1:]); err == nil {
						wg.Port = port
					}
				} else {
					wg.Server = value
					wg.Port = 51820
				}
			case "publickey":
				wg.PublicKey = value
			}
		}
	}

	// 验证必要字段
	if wg.PrivateKey == "" {
		return WireGuard{}, fmt.Errorf("配置文件缺少 PrivateKey")
	}
	if wg.PublicKey == "" {
		return WireGuard{}, fmt.Errorf("配置文件缺少 PublicKey")
	}
	if wg.Server == "" {
		return WireGuard{}, fmt.Errorf("配置文件缺少 Endpoint")
	}

	// 生成默认名称
	if wg.Name == "" {
		wg.Name = fmt.Sprintf("%s:%s", wg.Server, utils.GetPortString(wg.Port))
	}

	return wg, nil
}
