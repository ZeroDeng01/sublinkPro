package node

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sublink/internal/yamlutil"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/services/mihomo"
	"sublink/services/notifications"
	"sublink/utils"
	"time"

	"gopkg.in/yaml.v3"
)

// TaskReporter 任务报告接口，用于解耦任务管理
// 由 scheduler 传入实现，避免 node 包导入 services 包导致的循环依赖
type TaskReporter interface {
	// UpdateTotal 更新任务总数（在解析完订阅后调用）
	UpdateTotal(total int)
	// ReportProgress 报告任务进度
	ReportProgress(current int, currentItem string, result any)
	// ReportComplete 报告任务完成
	ReportComplete(message string, result any)
	// ReportFail 报告任务失败
	ReportFail(errMsg string)
}

// NoOpTaskReporter 空实现，当没有传入reporter时使用
type NoOpTaskReporter struct{}

func (n *NoOpTaskReporter) UpdateTotal(total int)                                      {}
func (n *NoOpTaskReporter) ReportProgress(current int, currentItem string, result any) {}
func (n *NoOpTaskReporter) ReportComplete(message string, result any)                  {}
func (n *NoOpTaskReporter) ReportFail(errMsg string)                                   {}

const (
	// providerResponseSizeLimit 限制单个 provider 响应体大小，兼顾大型机场节点列表与内存上界。
	providerResponseSizeLimit int64 = 16 << 20
	// selectedProviderCountLimit 限制单次订阅展开的 provider 数量，避免异常配置触发无界外部请求。
	selectedProviderCountLimit = 64
)

// UsageInfo 订阅用量信息（从 subscription-userinfo header 解析）
type UsageInfo struct {
	Upload   int64 // 已上传流量（字节）
	Download int64 // 已下载流量（字节）
	Total    int64 // 总流量配额（字节）
	Expire   int64 // 订阅过期时间（Unix时间戳）
}

// ParseSubscriptionUserInfo 解析 subscription-userinfo header
// 格式: upload=189594657; download=39476274625; total=108447924224; expire=1768890123
func ParseSubscriptionUserInfo(headerValue string) *UsageInfo {
	if headerValue == "" {
		return nil
	}

	info := &UsageInfo{}
	// 按分号分割各个字段
	parts := strings.Split(headerValue, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// 按等号分割键值对
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "upload":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				info.Upload = v
			}
		case "download":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				info.Download = v
			}
		case "total":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				info.Total = v
			}
		case "expire":
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				info.Expire = v
			}
		}
	}

	// 如果所有字段都为0，则认为解析失败
	if info.Upload == 0 && info.Download == 0 && info.Total == 0 && info.Expire == 0 {
		return nil
	}

	return info
}

// FailedUsageInfo 返回表示用量信息获取失败的特殊值
// 使用 -1 作为 Total 字段的标记，表示开启了获取但机场不支持
func FailedUsageInfo() *UsageInfo {
	return &UsageInfo{
		Upload:   0,
		Download: 0,
		Total:    -1, // -1 表示获取失败
		Expire:   0,
	}
}

type ClashConfig struct {
	Proxies            []protocol.Proxy              `yaml:"proxies"`
	ProxyProviders     map[string]ClashProxyProvider `yaml:"proxy-providers"`
	ProxyProviderOrder []string                      `yaml:"-"`
	ProxyGroups        []ClashProxyGroup             `yaml:"proxy-groups"`
	incompleteLinks    bool
}

type ClashProxyProvider struct {
	Type    string           `yaml:"type"`
	URL     string           `yaml:"url"`
	Proxies []protocol.Proxy `yaml:"proxies"`
}

type ClashProxyGroup struct {
	Use                 []string `yaml:"use"`
	IncludeAll          bool     `yaml:"include-all"`
	IncludeAllProviders bool     `yaml:"include-all-providers"`
}

func (g *ClashProxyGroup) UnmarshalYAML(value *yaml.Node) error {
	type groupAlias ClashProxyGroup
	var decoded groupAlias
	if err := yamlutil.DecodeWithStringBools(value, &decoded); err != nil {
		return err
	}
	*g = ClashProxyGroup(decoded)
	return nil
}

type namedClashProxyProvider struct {
	Name     string
	Provider ClashProxyProvider
}

func (c *ClashConfig) UnmarshalYAML(value *yaml.Node) error {
	type rawClashConfig struct {
		Proxies        []protocol.Proxy              `yaml:"proxies"`
		ProxyProviders map[string]ClashProxyProvider `yaml:"proxy-providers"`
		ProxyGroups    []ClashProxyGroup             `yaml:"proxy-groups"`
	}

	var raw rawClashConfig
	if err := value.Decode(&raw); err != nil {
		return err
	}

	c.Proxies = raw.Proxies
	c.ProxyProviders = raw.ProxyProviders
	c.ProxyGroups = raw.ProxyGroups
	c.ProxyProviderOrder = orderedClashProxyProviderNames(value)
	return nil
}

func orderedClashProxyProviderNames(value *yaml.Node) []string {
	if value.Kind == yaml.DocumentNode && len(value.Content) > 0 {
		value = value.Content[0]
	}
	if value.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i+1 < len(value.Content); i += 2 {
		if value.Content[i].Value != "proxy-providers" {
			continue
		}
		providersNode := value.Content[i+1]
		if providersNode.Kind != yaml.MappingNode {
			return nil
		}

		providerNames := make([]string, 0, len(providersNode.Content)/2)
		for j := 0; j+1 < len(providersNode.Content); j += 2 {
			providerNames = append(providerNames, strings.TrimSpace(providersNode.Content[j].Value))
		}
		return providerNames
	}

	return nil
}

// isTLSError 检测是否为 TLS 证书相关错误
func isTLSError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "x509:") ||
		strings.Contains(errStr, "certificate") ||
		strings.Contains(errStr, "tls:") ||
		strings.Contains(errStr, "TLS")
}

// LoadClashConfigFromURL 从指定 URL 加载 Clash 配置
// 支持 YAML 格式和 Base64 编码的订阅链接
// id: 订阅ID
// url: 订阅链接
// subName: 订阅名称
// downloadWithProxy: 是否使用代理下载
// proxyLink: 代理链接 (可选)
// userAgent: 请求的 User-Agent (可选，默认 Clash)
func LoadClashConfigFromURL(id int, urlStr string, subName string, downloadWithProxy bool, proxyLink string, userAgent string) (*UsageInfo, error) {
	_, usageInfo, err := LoadClashConfigFromURLWithReporter(context.TODO(), id, urlStr, subName, downloadWithProxy, proxyLink, userAgent, nil, nil, false, true)
	return usageInfo, err
}

func applyRequestHeaders(req *http.Request, userAgent string, requestHeaders models.AirportRequestHeaders) {
	resolvedUserAgent := strings.TrimSpace(userAgent)
	if resolvedUserAgent == "" {
		resolvedUserAgent = "clash.meta"
	}
	req.Header.Set("User-Agent", resolvedUserAgent)

	for _, header := range requestHeaders {
		if header.Key == "" || strings.EqualFold(header.Key, "User-Agent") {
			continue
		}
		req.Header.Add(header.Key, header.Value)
	}
}

func expandClashProxyProviders(ctx context.Context, client *http.Client, rootSubscriptionURL string, config *ClashConfig, userAgent string, requestHeaders models.AirportRequestHeaders) error {
	providers := selectClashProxyProviders(config)
	if len(providers) == 0 {
		return nil
	}
	if len(providers) > selectedProviderCountLimit {
		return fmt.Errorf("proxy-providers 数量过多: %d，最多允许 %d 个", len(providers), selectedProviderCountLimit)
	}

	rootHost := normalizedURLHost(rootSubscriptionURL)
	providerErrors := make([]string, 0)
	for _, item := range providers {
		if len(item.Provider.Proxies) > 0 {
			config.Proxies = append(config.Proxies, item.Provider.Proxies...)
			continue
		}

		// 机场订阅导入只展开远端 http provider，不实现本地 file、缓存或 health-check 等 mihomo 运行时语义。
		if !strings.EqualFold(strings.TrimSpace(item.Provider.Type), "http") {
			continue
		}

		providerURL := strings.TrimSpace(item.Provider.URL)
		if providerURL == "" {
			providerErrors = append(providerErrors, fmt.Sprintf("%s: 缺少 URL", item.Name))
			continue
		}

		providerProxies, err := fetchClashProxyProvider(ctx, client, item.Name, providerURL, rootHost, userAgent, requestHeaders)
		config.Proxies = append(config.Proxies, providerProxies...)
		if err != nil {
			utils.Warn("proxy-provider【%s】获取或解析失败: %v", item.Name, err)
			providerErrors = append(providerErrors, fmt.Sprintf("%s: %v", item.Name, err))
		}
	}

	if len(providerErrors) > 0 {
		return fmt.Errorf("proxy-providers 解析失败: %s", strings.Join(providerErrors, "; "))
	}
	return nil
}

func selectClashProxyProviders(config *ClashConfig) []namedClashProxyProvider {
	if config == nil || len(config.ProxyProviders) == 0 {
		return nil
	}

	includeAllProviders := false
	referencedNames := make([]string, 0)
	for _, group := range config.ProxyGroups {
		if group.IncludeAll || group.IncludeAllProviders {
			includeAllProviders = true
		}
		for _, providerName := range group.Use {
			providerName = strings.TrimSpace(providerName)
			if providerName != "" {
				referencedNames = append(referencedNames, providerName)
			}
		}
	}

	if includeAllProviders || len(referencedNames) == 0 {
		referencedNames = orderedClashProviderNames(config)
	}

	providers := make([]namedClashProxyProvider, 0, len(referencedNames))
	seenNames := make(map[string]bool, len(referencedNames))
	seenURLs := make(map[string]bool, len(referencedNames))
	for _, providerName := range referencedNames {
		provider, ok := config.ProxyProviders[providerName]
		if !ok {
			continue
		}

		providerURL := strings.TrimSpace(provider.URL)
		// 同一 URL 的 provider 只拉取一次，保留首次声明，避免重复外呼和重复节点。
		if seenNames[providerName] || (providerURL != "" && seenURLs[providerURL]) {
			continue
		}
		seenNames[providerName] = true
		if providerURL != "" {
			seenURLs[providerURL] = true
		}

		providers = append(providers, namedClashProxyProvider{Name: providerName, Provider: provider})
	}

	return providers
}

func orderedClashProviderNames(config *ClashConfig) []string {
	providerNames := make([]string, 0, len(config.ProxyProviders))
	seen := make(map[string]bool, len(config.ProxyProviders))
	for _, providerName := range config.ProxyProviderOrder {
		if _, ok := config.ProxyProviders[providerName]; ok && !seen[providerName] {
			providerNames = append(providerNames, providerName)
			seen[providerName] = true
		}
	}

	remainingNames := make([]string, 0)
	for providerName := range config.ProxyProviders {
		if !seen[providerName] {
			remainingNames = append(remainingNames, providerName)
		}
	}
	sort.Strings(remainingNames)
	providerNames = append(providerNames, remainingNames...)
	return providerNames
}

func fetchClashProxyProvider(ctx context.Context, client *http.Client, providerName string, providerURL string, rootSubscriptionHost string, userAgent string, requestHeaders models.AirportRequestHeaders) ([]protocol.Proxy, error) {
	parsedProviderURL, err := validateClashProviderURL(providerURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, providerURL, nil)
	if err != nil {
		return nil, err
	}
	if rootSubscriptionHost != "" && strings.EqualFold(parsedProviderURL.Host, rootSubscriptionHost) {
		applyRequestHeaders(req, userAgent, requestHeaders)
	} else {
		applyRequestHeaders(req, userAgent, nil)
	}

	providerClient := newClashProviderHTTPClient(client, rootSubscriptionHost, userAgent, requestHeaders)
	resp, err := providerClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("HTTP 状态码 %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, providerResponseSizeLimit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > providerResponseSizeLimit {
		return nil, fmt.Errorf("provider 响应超过大小限制 %d bytes", providerResponseSizeLimit)
	}

	proxies, err := parseClashProviderPayload(data)
	if err != nil && len(proxies) == 0 {
		return nil, err
	}
	if len(proxies) == 0 {
		return nil, fmt.Errorf("未找到节点")
	}

	utils.Info("proxy-provider【%s】导入节点数量：%d", providerName, len(proxies))
	return proxies, err
}

func validateClashProviderURL(providerURL string) (*url.URL, error) {
	parsedProviderURL, err := url.Parse(providerURL)
	if err != nil {
		return nil, err
	}
	if parsedProviderURL.Host == "" {
		return nil, fmt.Errorf("provider URL 缺少 host")
	}
	if parsedProviderURL.Scheme != "http" && parsedProviderURL.Scheme != "https" {
		return nil, fmt.Errorf("provider URL scheme %q 不受支持", parsedProviderURL.Scheme)
	}
	return parsedProviderURL, nil
}

func normalizedURLHost(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.ToLower(parsedURL.Host)
}

func newClashProviderHTTPClient(client *http.Client, rootSubscriptionHost string, userAgent string, requestHeaders models.AirportRequestHeaders) *http.Client {
	providerClient := *client
	providerClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			return fmt.Errorf("provider URL scheme %q 不受支持", req.URL.Scheme)
		}
		applyRequestHeaders(req, userAgent, nil)
		if rootSubscriptionHost == "" || !strings.EqualFold(req.URL.Host, rootSubscriptionHost) {
			removeCustomAirportHeaders(req, requestHeaders)
		}
		return nil
	}
	return &providerClient
}

func removeCustomAirportHeaders(req *http.Request, requestHeaders models.AirportRequestHeaders) {
	for _, header := range requestHeaders {
		if header.Key == "" || strings.EqualFold(header.Key, "User-Agent") {
			continue
		}
		req.Header.Del(header.Key)
	}
}

func parseClashProviderPayload(data []byte) ([]protocol.Proxy, error) {
	var config ClashConfig
	errYaml := yaml.Unmarshal(data, &config)
	if errYaml == nil && len(config.Proxies) > 0 {
		return config.Proxies, nil
	}

	proxies, incomplete := parseSubscriptionLinkProxies(data)
	if len(proxies) > 0 {
		if incomplete {
			return proxies, fmt.Errorf("部分节点链接无法解析")
		}
		return proxies, nil
	}

	if errYaml != nil {
		return nil, fmt.Errorf("YAML 解析失败且链接 fallback 未找到节点: %w", errYaml)
	}
	return nil, fmt.Errorf("未找到节点")
}

func parseSubscriptionLinkProxies(data []byte) ([]protocol.Proxy, bool) {
	decodedBytes, errB64 := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
	if errB64 != nil {
		decodedBytes, errB64 = base64.RawStdEncoding.DecodeString(strings.TrimSpace(string(data)))
	}
	if errB64 == nil {
		if proxies, incomplete := parsePlainSubscriptionLinkProxies(decodedBytes); len(proxies) > 0 {
			return proxies, incomplete
		}
	}
	return parsePlainSubscriptionLinkProxies(data)
}

func parsePlainSubscriptionLinkProxies(data []byte) ([]protocol.Proxy, bool) {
	proxies := make([]protocol.Proxy, 0)
	incomplete := false
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		proxy, errP := protocol.LinkToProxy(protocol.Urls{Url: line}, protocol.OutputConfig{})
		if errP == nil {
			proxies = append(proxies, proxy)
		} else {
			incomplete = true
		}
	}
	return proxies, incomplete
}

func parseClashConfigData(ctx context.Context, client *http.Client, rootSubscriptionURL string, data []byte, userAgent string, requestHeaders models.AirportRequestHeaders) (ClashConfig, error, error) {
	var config ClashConfig
	// 尝试解析 YAML
	errYaml := yaml.Unmarshal(data, &config)
	var providerErr error
	if errYaml == nil && len(config.Proxies) == 0 && len(config.ProxyProviders) > 0 {
		providerErr = expandClashProxyProviders(ctx, client, rootSubscriptionURL, &config, userAgent, requestHeaders)
	}

	// 如果 YAML 解析失败或没有代理节点，尝试 Base64/明文链接解析，兼容 V2Ray 订阅。
	if errYaml != nil || len(config.Proxies) == 0 {
		proxies, incomplete := parseSubscriptionLinkProxies(data)
		config.Proxies = append(config.Proxies, proxies...)
		config.incompleteLinks = incomplete
	}

	return config, errYaml, providerErr
}

// LoadClashConfigFromURLWithReporter 从指定 URL 加载 Clash 配置（带任务报告器）
// reporter: 任务进度报告器，用于TaskManager集成
// fetchUsageInfo: 是否获取用量信息
// skipTLSVerify: 是否跳过TLS证书验证
func LoadClashConfigFromURLWithReporter(ctx context.Context, id int, urlStr string, subName string, downloadWithProxy bool, proxyLink string, userAgent string, requestHeaders models.AirportRequestHeaders, reporter TaskReporter, fetchUsageInfo bool, skipTLSVerify bool) ([]int, *UsageInfo, error) {
	// 检查任务是否已被取消
	select {
	case <-ctx.Done():
		return nil, nil, fmt.Errorf("任务已取消")
	default:
	}

	// 创建 HTTP 客户端，配置 TLS
	client := utils.NewHTTPClientWithAdapter(nil, 30*time.Second, &tls.Config{InsecureSkipVerify: skipTLSVerify})

	if downloadWithProxy {
		var proxyNodeLink string

		if proxyLink != "" {
			// 使用指定的代理链接
			proxyNodeLink = proxyLink
			utils.Info("使用指定代理下载订阅")
		} else {
			// 如果没有指定代理，尝试自动选择最佳代理
			// 获取最近测速成功的节点（延迟最低且速度大于0）
			if bestNode, err := models.GetBestProxyNode(); err == nil && bestNode != nil {
				utils.Info("自动选择最佳代理节点: %s 节点延迟：%dms  节点速度：%2fMB/s", bestNode.Name, bestNode.DelayTime, bestNode.Speed)
				proxyNodeLink = bestNode.Link
			}
		}

		if proxyNodeLink != "" {
			// 使用 mihomo 内核创建代理适配器
			proxyAdapter, err := mihomo.GetMihomoAdapter(proxyNodeLink)
			if err != nil {
				utils.Error("创建 mihomo 代理适配器失败: %v，将直接下载", err)
			} else {
				utils.Info("使用 mihomo 内核代理下载订阅")
				// 创建自定义 Transport，使用 mihomo adapter 进行代理连接
				defer func() { _ = proxyAdapter.Close() }()
				client = utils.NewHTTPClientWithAdapter(proxyAdapter, 30*time.Second, &tls.Config{InsecureSkipVerify: skipTLSVerify})
			}
		} else {
			utils.Warn("未找到可用代理，将直接下载")
		}
	}

	defer utils.CloseHTTPClient(client)

	// 创建请求并设置 User-Agent（使用传入的 context）
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		utils.Error("URL %s，创建请求失败:  %v", urlStr, err)
		return nil, nil, err
	}

	applyRequestHeaders(req, userAgent, requestHeaders)

	resp, err := client.Do(req)
	if err != nil {
		utils.Error("URL %s，获取Clash配置失败:  %v", urlStr, err)
		// 检测是否为 TLS 证书相关错误，给出更明确的提示
		var title, message string
		if isTLSError(err) {
			title = "订阅更新失败 - TLS证书验证错误"
			if skipTLSVerify {
				message = fmt.Sprintf("❌订阅【%s】TLS错误: %v", subName, err)
			} else {
				message = fmt.Sprintf("❌订阅【%s】证书验证失败: %v\n\n💡 提示：请在机场设置中开启\"忽略证书验证\"选项后重试", subName, err)
			}
		} else {
			title = "订阅更新失败"
			message = fmt.Sprintf("❌订阅【%s】请求失败: %v", subName, err)
		}
		// 发送请求失败通知
		notifications.Publish("subscription.sync_failed", notifications.Payload{
			Title:   title,
			Message: message,
			Data: map[string]any{
				"id":       id,
				"name":     subName,
				"status":   "error",
				"error":    err.Error(),
				"tlsError": isTLSError(err),
			},
		})
		return nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, fmt.Errorf("订阅请求返回 HTTP %d", resp.StatusCode)
	}

	// 解析用量信息（仅当开启获取用量信息时）
	var usageInfo *UsageInfo
	if fetchUsageInfo {
		subUserInfo := resp.Header.Get("subscription-userinfo")
		if subUserInfo != "" {
			usageInfo = ParseSubscriptionUserInfo(subUserInfo)
			if usageInfo != nil {
				utils.Info("订阅【%s】获取用量信息成功: 上传=%d, 下载=%d, 总量=%d, 过期=%d",
					subName, usageInfo.Upload, usageInfo.Download, usageInfo.Total, usageInfo.Expire)
			} else {
				// header 存在但解析失败
				utils.Warn("订阅【%s】用量信息 header 解析失败", subName)
				usageInfo = FailedUsageInfo()
			}
		} else {
			// 开启了获取但机场未返回 header
			utils.Warn("订阅【%s】未返回用量信息 header，机场可能不支持", subName)
			usageInfo = FailedUsageInfo()
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Error("URL %s，读取Clash配置失败:  %v", urlStr, err)
		// 发送读取失败通知
		notifications.Publish("subscription.sync_failed", notifications.Payload{
			Title:   "订阅更新失败",
			Message: fmt.Sprintf("❌订阅【%s】读取响应失败: %v", subName, err),
			Data: map[string]any{
				"id":     id,
				"name":   subName,
				"status": "error",
				"error":  err.Error(),
			},
		})
		return nil, nil, err
	}

	// 检查任务是否在解析前被取消
	select {
	case <-ctx.Done():
		return nil, nil, fmt.Errorf("任务已取消")
	default:
	}

	config, errYaml, providerErr := parseClashConfigData(ctx, client, urlStr, data, userAgent, requestHeaders)

	if len(config.Proxies) == 0 {
		parseErr := fmt.Errorf("解析失败 or 未找到节点")
		errorMessage := "解析失败或未找到节点"
		if providerErr != nil {
			parseErr = providerErr
			errorMessage = providerErr.Error()
		}

		utils.Error("URL %s，解析失败或未找到节点 (YAML error: %v, provider error: %v)", urlStr, errYaml, providerErr)
		// 发送解析失败通知
		notifications.Publish("subscription.sync_failed", notifications.Payload{
			Title:   "订阅更新失败",
			Message: fmt.Sprintf("❌订阅【%s】%s", subName, errorMessage),
			Data: map[string]any{
				"id":     id,
				"name":   subName,
				"status": "error",
				"error":  errorMessage,
			},
		})
		return nil, nil, parseErr
	}

	changedNodeIDs, err := scheduleClashToNodeLinks(ctx, id, config.Proxies, subName, reporter, usageInfo, providerErr != nil || config.incompleteLinks)
	return changedNodeIDs, usageInfo, err
}

// scheduleClashToNodeLinks 将 Clash 代理配置转换为节点链接并保存到数据库
// ctx: context 用于任务取消和超时控制
// id: 订阅ID
// proxys: 代理节点列表
// subName: 订阅名称
// usageInfo: 订阅用量信息 (可选)
func scheduleClashToNodeLinks(ctx context.Context, id int, proxys []protocol.Proxy, subName string, reporter TaskReporter, usageInfo *UsageInfo, incomplete bool) (_ []int, runErr error) {
	if reporter == nil {
		reporter = &NoOpTaskReporter{}
	}

	startTime := time.Now()

	defer func() {
		if r := recover(); r != nil {
			runErr = fmt.Errorf("订阅更新任务异常: %v", r)
		}
	}()

	airport, err := models.GetAirportByID(id)
	if err != nil {
		return nil, err
	}

	// 1. 读取全局节点处理配置
	globalWhitelist, _ := models.GetSetting("global_node_name_whitelist")
	globalBlacklist, _ := models.GetSetting("global_node_name_blacklist")
	globalProtocolWhitelist, _ := models.GetSetting("global_protocol_whitelist")
	globalProtocolBlacklist, _ := models.GetSetting("global_protocol_blacklist")
	globalPreprocess, _ := models.GetSetting("global_node_name_preprocess")

	// 2. 构造全局配置的虚拟 Airport 对象
	globalAirport := &models.Airport{
		NodeNameWhitelist:  globalWhitelist,
		NodeNameBlacklist:  globalBlacklist,
		ProtocolWhitelist:  globalProtocolWhitelist,
		ProtocolBlacklist:  globalProtocolBlacklist,
		NodeNamePreprocess: globalPreprocess,
	}

	// 3. 先应用全局过滤规则
	if globalWhitelist != "" || globalBlacklist != "" ||
		globalProtocolWhitelist != "" || globalProtocolBlacklist != "" {
		originalCount := len(proxys)
		proxys = applyAirportNodeFilter(globalAirport, proxys)
		if len(proxys) < originalCount {
			utils.Info("🌐全局过滤后节点数量：%d（原始：%d，过滤掉：%d）",
				len(proxys), originalCount, originalCount-len(proxys))
		}
	}

	// 4. 再应用机场特定过滤规则
	if airport != nil {
		originalCount := len(proxys)
		proxys = applyAirportNodeFilter(airport, proxys)
		if len(proxys) < originalCount {
			utils.Info("📦机场【%s】过滤后节点数量：%d（原始：%d，过滤掉：%d）", subName, len(proxys), originalCount, originalCount-len(proxys))
		}
		// 应用高级去重规则
		beforeDedup := len(proxys)
		proxys = applyAirportDeduplication(airport, proxys)
		if len(proxys) < beforeDedup {
			utils.Info("🔄机场【%s】去重后节点数量：%d（去重前：%d，去重掉：%d）", subName, len(proxys), beforeDedup, beforeDedup-len(proxys))
		}
	}

	// 5. 先应用全局重命名规则
	if globalPreprocess != "" {
		proxys = applyAirportNodeRename(globalAirport, proxys)
		utils.Info("🌐应用全局节点名称预处理规则")
	}

	// 6. 再应用机场特定重命名规则
	if airport != nil {
		proxys = applyAirportNodeRename(airport, proxys)
		// 节点名称唯一化（先添加机场标识前缀，机场内编号在后续去重判断完成后再应用）
		proxys = applyAirportNodeNamePrefix(airport, proxys)
	}

	result, err := syncAirportNodes(ctx, airport, subName, proxys, reporter, incomplete)
	if err != nil {
		return nil, err
	}
	summary := fmt.Sprintf("新增: %d, 更新: %d, 已存在或跳过: %d, 删除: %d", result.added, result.updated, result.skipped, result.deleted)
	if result.incomplete {
		summary += "。本次获取或转换不完整，已保留未匹配的旧节点"
	}
	utils.Info("订阅【%s】节点同步完成，成功处理 %d 个节点，%s", subName, result.successCount, summary)
	reporter.ReportComplete("订阅更新完成 ("+summary+")", map[string]any{
		"added": result.added, "updated": result.updated, "skipped": result.skipped, "deleted": result.deleted,
	})

	// 触发webhook的完成事件
	duration := time.Since(startTime)
	durationStr := formatDurationSub(duration)

	// 构建用量信息文本
	var usageText string
	usageData := make(map[string]any)
	if usageInfo != nil {
		if usageInfo.Total != -1 {
			usageText = fmt.Sprintf("\n📊 用量信息\n⬆️ 上传: %s\n⬇️ 下载: %s\n📦 总量: %s\n⏳ 过期: %s",
				utils.FormatBytes(usageInfo.Upload),
				utils.FormatBytes(usageInfo.Download),
				utils.FormatBytes(usageInfo.Total),
				time.Unix(usageInfo.Expire, 0).Format("2006-01-02 15:04:05"))
			usageData["upload"] = usageInfo.Upload
			usageData["download"] = usageInfo.Download
			usageData["total"] = usageInfo.Total
			usageData["expire"] = usageInfo.Expire
		}
	}

	nData := map[string]any{
		"id":       id,
		"name":     subName,
		"status":   "success",
		"success":  result.successCount,
		"duration": duration.Milliseconds(),
	}
	if len(usageData) > 0 {
		nData["usage"] = usageData
	}

	notifications.Publish("subscription.sync_succeeded", notifications.Payload{
		Title:   "订阅更新完成",
		Message: fmt.Sprintf("✅订阅【%s】节点同步完成，耗时 %s，总节点【%d】个，成功处理【%d】个，%s%s", subName, durationStr, len(proxys), result.successCount, summary, usageText),
		Data:    nData,
	})
	return result.changedIDs, nil
}

// formatDurationSub 格式化时长为人类可读字符串
func formatDurationSub(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f秒", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0f分%.0f秒", d.Minutes(), math.Mod(d.Seconds(), 60))
	}
	return fmt.Sprintf("%.0f时%.0f分", d.Hours(), math.Mod(d.Minutes(), 60))
}

// applyAirportNodeFilter 应用机场节点过滤规则
// 根据机场配置的白名单/黑名单规则过滤代理节点
func applyAirportNodeFilter(airport *models.Airport, proxys []protocol.Proxy) []protocol.Proxy {
	if airport == nil {
		return proxys
	}

	hasNameWhitelist := utils.HasActiveNodeNameFilter(airport.NodeNameWhitelist)
	hasNameBlacklist := utils.HasActiveNodeNameFilter(airport.NodeNameBlacklist)
	hasProtocolWhitelist := airport.ProtocolWhitelist != ""
	hasProtocolBlacklist := airport.ProtocolBlacklist != ""

	// 如果没有任何过滤规则，直接返回
	if !hasNameWhitelist && !hasNameBlacklist && !hasProtocolWhitelist && !hasProtocolBlacklist {
		return proxys
	}

	// 解析协议白名单和黑名单
	protocolWhitelistMap := make(map[string]bool)
	protocolBlacklistMap := make(map[string]bool)

	if hasProtocolWhitelist {
		for _, p := range strings.Split(airport.ProtocolWhitelist, ",") {
			p = strings.TrimSpace(strings.ToLower(p))
			if p != "" {
				protocolWhitelistMap[p] = true
			}
		}
	}

	if hasProtocolBlacklist {
		for _, p := range strings.Split(airport.ProtocolBlacklist, ",") {
			p = strings.TrimSpace(strings.ToLower(p))
			if p != "" {
				protocolBlacklistMap[p] = true
			}
		}
	}

	// 过滤节点
	result := make([]protocol.Proxy, 0, len(proxys))
	for _, proxy := range proxys {
		nodeName := strings.TrimSpace(proxy.Name)
		nodeProto := proxyProtocolName(proxy)

		// 1. 名称黑名单检查（优先级最高）
		if hasNameBlacklist && utils.MatchesNodeNameFilter(airport.NodeNameBlacklist, nodeName) {
			continue
		}

		// 2. 名称白名单检查
		if hasNameWhitelist && !utils.MatchesNodeNameFilter(airport.NodeNameWhitelist, nodeName) {
			continue
		}

		// 3. 协议黑名单检查
		if len(protocolBlacklistMap) > 0 && protocolBlacklistMap[nodeProto] {
			continue
		}

		// 4. 协议白名单检查
		if len(protocolWhitelistMap) > 0 && !protocolWhitelistMap[nodeProto] {
			continue
		}

		result = append(result, proxy)
	}

	return result
}

func proxyProtocolName(proxy protocol.Proxy) string {
	proxyType := strings.ToLower(strings.TrimSpace(proxy.Type))
	if proxyType == "http" && proxy.Tls {
		return "https"
	}
	return proxyType
}

// applyAirportNodeRename 应用机场节点重命名规则
// 根据机场配置的预处理规则对节点名称进行替换
func applyAirportNodeRename(airport *models.Airport, proxys []protocol.Proxy) []protocol.Proxy {
	if airport == nil || airport.NodeNamePreprocess == "" {
		return proxys
	}

	// 应用预处理规则到每个节点
	for i := range proxys {
		originalName := proxys[i].Name
		processedName := utils.PreprocessNodeName(airport.NodeNamePreprocess, originalName)
		if processedName != originalName {
			proxys[i].Name = processedName
		}
	}

	return proxys
}

// applyAirportDeduplication 应用机场高级去重规则
// 根据机场配置的去重规则对代理节点进行去重
func applyAirportDeduplication(airport *models.Airport, proxys []protocol.Proxy) []protocol.Proxy {
	if airport == nil || airport.DeduplicationRule == "" {
		return proxys
	}

	// 解析去重配置
	var config models.DeduplicationConfig
	if err := json.Unmarshal([]byte(airport.DeduplicationRule), &config); err != nil {
		utils.Warn("解析机场去重规则失败: %v", err)
		return proxys
	}

	// 只有 protocol 模式才进行高级去重
	if config.Mode != "protocol" || len(config.ProtocolRules) == 0 {
		return proxys
	}

	// 按协议字段去重
	seen := make(map[string]bool)
	var result []protocol.Proxy

	for _, proxy := range proxys {
		protoType := proxyProtocolName(proxy)
		fields, exists := config.ProtocolRules[protoType]
		if !exists || len(fields) == 0 {
			// 该协议未配置去重规则，保留节点
			result = append(result, proxy)
			continue
		}

		// 生成去重Key（需传入协议类型用于解析）
		key := generateProxyDeduplicationKey(proxy, protoType, fields)
		if key == "" {
			result = append(result, proxy)
			continue
		}

		// 加上协议类型前缀，避免不同协议间Key冲突
		fullKey := protoType + ":" + key
		if !seen[fullKey] {
			seen[fullKey] = true
			result = append(result, proxy)
		}
	}

	return result
}

// generateProxyDeduplicationKey 根据指定字段生成代理的去重Key
// 需要先生成节点链接，再解析获取完整协议结构体，才能正确提取嵌套字段
func generateProxyDeduplicationKey(proxy protocol.Proxy, protoType string, fields []string) string {
	// 生成节点链接
	link := GenerateProxyLink(proxy)
	if link == "" {
		return ""
	}

	// 解析链接获取完整协议结构体
	protoObj, err := parseProtoFromLink(link, protoType)
	if err != nil || protoObj == nil {
		return ""
	}

	// 使用反射获取嵌套字段值
	var parts []string
	for _, field := range fields {
		value := protocol.GetProtocolFieldValue(protoObj, field)
		parts = append(parts, field+":"+value)
	}
	return strings.Join(parts, "|")
}

const defaultImportedECHResolver = "https://dns.alidns.com/dns-query"

// GenerateProxyLink 从 Proxy 结构体生成节点链接
func GenerateProxyLink(proxy protocol.Proxy) string {
	if strings.EqualFold(proxy.Type, "vless") {
		vless := protocol.ConvertProxyToVless(proxy)
		if strings.TrimSpace(vless.Query.Ech) == "" {
			enabled := true
			if rawEnabled, ok := proxy.ECH_opts["enable"].(bool); ok {
				enabled = rawEnabled
			}
			if enabled {
				if queryServerName, ok := proxy.ECH_opts["query-server-name"].(string); ok {
					queryServerName = strings.TrimSpace(queryServerName)
					if queryServerName != "" {
						vless.Query.Ech = queryServerName + "+" + defaultImportedECHResolver
					}
				}
			}
		}

		if strings.TrimSpace(vless.Query.Ech) != "" {
			vless.Name = strings.TrimSpace(vless.Name)
			vless.Server = utils.WrapIPv6Host(vless.Server)
			return protocol.EncodeVLESSURL(vless)
		}
	}

	proxy.Name = strings.TrimSpace(proxy.Name)
	proxy.Server = utils.WrapIPv6Host(proxy.Server)
	link, err := protocol.EncodeProxyLink(proxy)
	if err != nil {
		return ""
	}
	return link
}

// applyAirportNodeNamePrefix 应用机场节点名称前缀唯一化
// 在节点名称前添加机场标识前缀，防止多机场间节点名称重复
// 同一机场同一节点每次生成的名字保持一致（使用机场ID生成稳定前缀）
func applyAirportNodeNamePrefix(airport *models.Airport, proxys []protocol.Proxy) []protocol.Proxy {
	if airport == nil || !airport.NodeNameUniquify {
		return proxys
	}

	// 生成前缀: 使用用户自定义前缀 或 默认的 [A{id}] 格式
	prefix := strings.TrimSpace(airport.NodeNamePrefix)
	if prefix == "" {
		prefix = fmt.Sprintf("[A%d]", airport.ID)
	}

	// 为每个节点名称添加前缀
	for i := range proxys {
		proxys[i].Name = prefix + proxys[i].Name
	}

	return proxys
}

// applyAirportIntraNodeUniquify 应用机场内节点名称唯一化
// 对同一机场拉取结果中重复的节点名称追加顺序编号，避免同机场内出现重名节点
func applyAirportIntraNodeUniquify(airport *models.Airport, proxys []protocol.Proxy) []protocol.Proxy {
	if airport == nil || !airport.NodeNameIntraUniquify {
		return proxys
	}

	nameTotals := make(map[string]int, len(proxys))
	for _, proxy := range proxys {
		nameTotals[proxy.Name]++
	}

	nameIndexes := make(map[string]int, len(nameTotals))
	for i := range proxys {
		name := proxys[i].Name
		if nameTotals[name] <= 1 {
			continue
		}

		nameIndexes[name]++
		proxys[i].Name = fmt.Sprintf("%s-%d", name, nameIndexes[name])
	}

	return proxys
}

// parseProtoFromLink 根据协议类型解析链接获取结构体
func parseProtoFromLink(link string, protoType string) (any, error) {
	protoObj, detectedProto, err := protocol.DecodeProtocolObject(link)
	if err != nil {
		return nil, err
	}
	if detectedProto != protoType {
		return nil, fmt.Errorf("unsupported protocol: %s", protoType)
	}
	return protoObj, nil
}
