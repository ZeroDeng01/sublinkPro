package mihomo

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/utils"
	"time"

	"github.com/metacubex/mihomo/component/resolver"
	"github.com/miekg/dns"
)

// HostInfo 包含从节点link解析的主机信息
type HostInfo struct {
	Server string // 代理服务器地址（域名或IP）
	IsIP   bool   // 是否为IP地址（不需要DNS解析）
}

// DNS服务器预设列表（用于前端下拉选择）
var DNSPresets = []struct {
	Label string `json:"label"`
	Value string `json:"value"`
}{
	{"AliDNS (DoH)", "https://dns.alidns.com/dns-query"},
	{"Tencent DNSPod (DoH)", "https://doh.pub/dns-query"},
	{"Cloudflare (DoH)", "https://cloudflare-dns.com/dns-query"},
	{"Google (DoH)", "https://dns.google/dns-query"},
	{"Cloudflare IP (DoH)", "https://1.1.1.1/dns-query"},
	{"Google IP(DoH)", "https://8.8.8.8/dns-query"},
	{"AliDNS", "223.5.5.5"},
	{"Tencent DNS", "119.29.29.29"},
	{"Cloudflare", "1.1.1.1"},
	{"Google", "8.8.8.8"},
}

// 默认DNS服务器
const DefaultDNSServer = "https://dns.alidns.com/dns-query"

// GetProxyServerFromLink 从节点link解析代理服务器地址
func GetProxyServerFromLink(nodeLink string) HostInfo {
	if nodeLink == "" {
		return HostInfo{}
	}

	outputConfig := protocol.OutputConfig{
		Udp:  true,
		Cert: true,
	}

	proxyStruct, err := protocol.LinkToProxy(protocol.Urls{Url: nodeLink}, outputConfig)
	if err != nil {
		utils.Warn("Failed to parse node link: %v", err)
		return HostInfo{}
	}

	server := proxyStruct.Server
	if server == "" {
		return HostInfo{}
	}

	isIP := net.ParseIP(server) != nil

	return HostInfo{
		Server: server,
		IsIP:   isIP,
	}
}

// ResolveProxyHost 解析代理服务器域名，返回第一个可用IP和解析来源
// 优先级：Host缓存 > mihomo resolver > 用户配置的DNS > 系统DNS
// 这样可以避免测速时重复DNS解析
func ResolveProxyHost(host string) (string, string) {
	if host == "" {
		return "", ""
	}

	// 如果已经是IP地址，直接返回
	if net.ParseIP(host) != nil {
		return host, "IP address"
	}

	// 1. 先检查Host缓存是否已有（避免重复解析）
	if cachedHost, err := models.GetHostByHostname(host); err == nil && cachedHost != nil {
		utils.Debug("[DNS] %s -> %s (host cache)", host, cachedHost.IP)
		return cachedHost.IP, "host cache"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 增加超时时间以容纳代理连接
	defer cancel()

	// 2. 使用mihomo的resolver
	r := resolver.ProxyServerHostResolver
	if r == nil {
		r = resolver.DefaultResolver
	}

	if r != nil {
		ips, err := r.LookupIP(ctx, host)
		if err == nil && len(ips) > 0 {
			ip := ips[0].String()
			utils.Debug("[DNS] %s -> %s (mihomo resolver)", host, ip)
			return ip, "mihomo resolver"
		}
		// 如果是默认resolver失败，这很正常，继续尝试其他方式
		// utils.Debug("[DNS] mihomo resolver失败: %s, %v", host, err)
	}

	// 3. 使用用户配置的DNS服务器
	dnsServer, _ := models.GetSetting("dns_server")

	// 如果配置了DNS服务器，尝试解析
	if dnsServer != "" {
		// 获取代理配置
		useProxyStr, _ := models.GetSetting("dns_use_proxy")
		useProxy := useProxyStr == "true"

		var proxyLink string
		if useProxy {
			strategy, _ := models.GetSetting("dns_proxy_strategy")
			if strategy == "manual" {
				nodeIDStr, _ := models.GetSetting("dns_proxy_node_id")
				nodeID, _ := strconv.Atoi(nodeIDStr)
				if nodeID > 0 {
					node := &models.Node{ID: nodeID}
					if err := node.GetByID(); err == nil {
						proxyLink = node.Link
						utils.Debug("[DNS] Using manually specified proxy node: %s", node.Name)
					}
				}
			} else {
				// 自动选择最佳代理
				if utils.GetBestProxyNodeFunc != nil {
					link, name, err := utils.GetBestProxyNodeFunc()
					if err == nil && link != "" {
						proxyLink = link
						utils.Debug("[DNS] Auto-selected best proxy node: %s", name)
					}
				}
			}

			if proxyLink == "" {
				utils.Warn("[DNS] Proxy is enabled but no usable proxy node was found; trying direct DNS")
				useProxy = false
			}
		}

		if ip := resolveWithCustomDNS(ctx, host, dnsServer, useProxy, proxyLink); ip != "" {
			proxyInfo := ""
			if useProxy {
				proxyInfo = " (via proxy)"
			}
			utils.Info("[DNS] %s -> %s (server: %s%s)", host, ip, dnsServer, proxyInfo)
			return ip, dnsServer
		}
	}

	// 4. Fallback: 使用系统DNS
	// 如果用户没有配置DNS服务器，或者配置的DNS解析失败，则回退到系统DNS
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		utils.Warn("[DNS] All resolution methods failed: %s", host)
		return "", ""
	}

	if len(addrs) > 0 {
		ip := addrs[0].IP.String()
		utils.Info("[DNS] %s -> %s (system DNS)", host, ip)
		return ip, "system DNS"
	}

	return "", ""
}

// resolveWithCustomDNS 使用自定义DNS服务器解析域名
// 支持格式: DoH(https://...), DoT(tls://...), 普通DNS(IP地址)
func resolveWithCustomDNS(ctx context.Context, host, dnsServer string, useProxy bool, proxyLink string) string {
	if dnsServer == "" {
		return ""
	}

	// 判断DNS服务器类型
	if strings.HasPrefix(dnsServer, "https://") {
		return resolveWithDoH(ctx, host, dnsServer, useProxy, proxyLink)
	} else if strings.HasPrefix(dnsServer, "tls://") {
		// DoT暂不实现，fallback
		return ""
	} else {
		return resolveWithUDP(ctx, host, dnsServer, useProxy, proxyLink)
	}
}

// resolveWithDoH 使用DoH服务器解析域名 (RFC 8484)
func resolveWithDoH(ctx context.Context, host, dohServer string, useProxy bool, proxyLink string) string {
	// 构造 DNS 消息
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(host), dns.TypeA)
	m.RecursionDesired = true
	data, err := m.Pack()
	if err != nil {
		utils.Warn("[DNS] Failed to pack DNS message: %v", err)
		return ""
	}

	// 创建带代理的 Client
	client, _, err := utils.CreateProxyHTTPClient(useProxy, proxyLink, 5*time.Second)
	if err != nil {
		utils.Warn("[DNS] Failed to create DoH proxy client: %v", err)
		return ""
	}

	// 发送 POST 请求 (RFC 8484)
	req, err := http.NewRequestWithContext(ctx, "POST", dohServer, bytes.NewReader(data))
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	resp, err := client.Do(req)
	if err != nil {
		utils.Warn("[DNS] DoH request failed: %s, %v", dohServer, err)
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		utils.Warn("[DNS] DoH response status is not 200: %s, %d", dohServer, resp.StatusCode)
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	// 解析响应
	rm := new(dns.Msg)
	if err := rm.Unpack(body); err != nil {
		utils.Warn("[DNS] Failed to parse DoH response: %v", err)
		return ""
	}

	for _, answer := range rm.Answer {
		if a, ok := answer.(*dns.A); ok {
			return a.A.String()
		}
		if aaaa, ok := answer.(*dns.AAAA); ok {
			return aaaa.AAAA.String()
		}
	}

	utils.Warn("[DNS] DoH returned no valid records: %s", dohServer)
	return ""
}

// resolveWithUDP 使用普通UDP DNS解析
func resolveWithUDP(ctx context.Context, host, dnsServer string, useProxy bool, proxyLink string) string {
	// 确保有端口
	if !strings.Contains(dnsServer, ":") {
		dnsServer = dnsServer + ":53"
	}

	// 注意：根据设计要求，UDP DNS 不经过代理，始终直连。
	// 只有 DoH 支持代理。
	d := net.Dialer{Timeout: 3 * time.Second}
	dialFunc := d.DialContext

	// 使用自定义resolver
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialFunc(ctx, "udp", dnsServer)
		},
	}

	addrs, err := r.LookupIPAddr(ctx, host)
	if err != nil {
		utils.Debug("UDP DNS resolution failed: %s -> %s, %v", host, dnsServer, err)
		return ""
	}

	if len(addrs) > 0 {
		return addrs[0].IP.String()
	}

	return ""
}

// GetDNSPresets 获取DNS预设列表（供API调用）
func GetDNSPresets() []map[string]string {
	result := make([]map[string]string, len(DNSPresets))
	for i, p := range DNSPresets {
		result[i] = map[string]string{
			"label": p.Label,
			"value": p.Value,
		}
	}
	return result
}
