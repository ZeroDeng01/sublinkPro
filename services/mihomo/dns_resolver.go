package mihomo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/utils"
	"time"

	"github.com/metacubex/mihomo/component/resolver"
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
	{"阿里DNS (DoH)", "https://dns.alidns.com/dns-query"},
	{"腾讯DNSPod (DoH)", "https://doh.pub/dns-query"},
	{"Cloudflare (DoH)", "https://cloudflare-dns.com/dns-query"},
	{"Google (DoH)", "https://dns.google/dns-query"},
	{"阿里DNS", "223.5.5.5"},
	{"腾讯DNS", "119.29.29.29"},
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
		utils.Debug("解析节点link失败: %v", err)
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

// ResolveProxyHost 解析代理服务器域名，返回第一个可用IP
// 优先级：Host缓存 > mihomo resolver > 用户配置的DNS > 系统DNS
// 这样可以避免测速时重复DNS解析
func ResolveProxyHost(host string) string {
	if host == "" {
		return ""
	}

	// 如果已经是IP地址，直接返回
	if net.ParseIP(host) != nil {
		return host
	}

	// 1. 先检查Host缓存是否已有（避免重复解析）
	if cachedHost, err := models.GetHostByHostname(host); err == nil && cachedHost != nil {
		utils.Debug("[DNS] %s -> %s (Host缓存)", host, cachedHost.IP)
		return cachedHost.IP
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
			return ip
		}
		utils.Debug("[DNS] mihomo resolver失败: %s, %v", host, err)
	}

	// 3. 使用用户配置的DNS服务器
	dnsServer, _ := models.GetSetting("dns_server")
	if dnsServer == "" {
		dnsServer = DefaultDNSServer
	}

	if ip := resolveWithCustomDNS(ctx, host, dnsServer); ip != "" {
		utils.Info("[DNS] %s -> %s (服务器: %s)", host, ip, dnsServer)
		return ip
	}

	// 4. Fallback: 使用系统DNS
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		utils.Warn("[DNS] 所有解析方式失败: %s", host)
		return ""
	}

	if len(addrs) > 0 {
		ip := addrs[0].IP.String()
		utils.Info("[DNS] %s -> %s (系统DNS)", host, ip)
		return ip
	}

	return ""
}

// resolveWithCustomDNS 使用自定义DNS服务器解析域名
// 支持格式: DoH(https://...), DoT(tls://...), 普通DNS(IP地址)
func resolveWithCustomDNS(ctx context.Context, host, dnsServer string) string {
	if dnsServer == "" {
		return ""
	}

	// 判断DNS服务器类型
	if strings.HasPrefix(dnsServer, "https://") {
		return resolveWithDoH(ctx, host, dnsServer)
	} else if strings.HasPrefix(dnsServer, "tls://") {
		// DoT暂不实现，fallback
		return ""
	} else {
		return resolveWithUDP(ctx, host, dnsServer)
	}
}

// resolveWithDoH 使用DoH服务器解析域名
func resolveWithDoH(ctx context.Context, host, dohServer string) string {
	reqURL := fmt.Sprintf("%s?name=%s&type=A", dohServer, url.QueryEscape(host))

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/dns-json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		utils.Debug("DoH请求失败: %s, %v", dohServer, err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var dohResp struct {
		Status int `json:"Status"`
		Answer []struct {
			Type int    `json:"type"`
			Data string `json:"data"`
		} `json:"Answer"`
	}

	if err := json.Unmarshal(body, &dohResp); err != nil {
		return ""
	}

	// A记录(type=1) 或 AAAA记录(type=28)
	for _, ans := range dohResp.Answer {
		if (ans.Type == 1 || ans.Type == 28) && net.ParseIP(ans.Data) != nil {
			return ans.Data
		}
	}

	return ""
}

// resolveWithUDP 使用普通UDP DNS解析
func resolveWithUDP(ctx context.Context, host, dnsServer string) string {
	// 确保有端口
	if !strings.Contains(dnsServer, ":") {
		dnsServer = dnsServer + ":53"
	}

	// 使用自定义resolver
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 3 * time.Second}
			return d.DialContext(ctx, "udp", dnsServer)
		},
	}

	addrs, err := r.LookupIPAddr(ctx, host)
	if err != nil {
		utils.Debug("UDP DNS解析失败: %s -> %s, %v", host, dnsServer, err)
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
