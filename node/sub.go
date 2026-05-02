package node

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/services/mihomo"
	"sublink/services/notifications"
	"sublink/utils"
	"time"

	"github.com/metacubex/mihomo/constant"
	"gopkg.in/yaml.v3"
)

// TaskReporter 任务报告接口，用于解耦任务管理
// 由 scheduler 传入实现，避免 node 包导入 services 包导致的循环依赖
type TaskReporter interface {
	// UpdateTotal 更新任务总数（在解析完订阅后调用）
	UpdateTotal(total int)
	// ReportProgress 报告任务进度
	ReportProgress(current int, currentItem string, result interface{})
	// ReportComplete 报告任务完成
	ReportComplete(message string, result interface{})
	// ReportFail 报告任务失败
	ReportFail(errMsg string)
}

// NoOpTaskReporter 空实现，当没有传入reporter时使用
type NoOpTaskReporter struct{}

func (n *NoOpTaskReporter) UpdateTotal(total int)                                              {}
func (n *NoOpTaskReporter) ReportProgress(current int, currentItem string, result interface{}) {}
func (n *NoOpTaskReporter) ReportComplete(message string, result interface{})                  {}
func (n *NoOpTaskReporter) ReportFail(errMsg string)                                           {}

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
	Proxies []protocol.Proxy `yaml:"proxies"`
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
	return LoadClashConfigFromURLWithReporter(id, urlStr, subName, downloadWithProxy, proxyLink, userAgent, nil, nil, false, true)
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

// LoadClashConfigFromURLWithReporter 从指定 URL 加载 Clash 配置（带任务报告器）
// reporter: 任务进度报告器，用于TaskManager集成
// fetchUsageInfo: 是否获取用量信息
// skipTLSVerify: 是否跳过TLS证书验证
func LoadClashConfigFromURLWithReporter(id int, urlStr string, subName string, downloadWithProxy bool, proxyLink string, userAgent string, requestHeaders models.AirportRequestHeaders, reporter TaskReporter, fetchUsageInfo bool, skipTLSVerify bool) (*UsageInfo, error) {
	// 创建 HTTP 客户端，配置 TLS
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
		},
	}

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
				client.Transport = &http.Transport{
					DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
						// 解析地址获取主机和端口
						host, portStr, splitErr := net.SplitHostPort(addr)
						if splitErr != nil {
							return nil, fmt.Errorf("split host port error: %v", splitErr)
						}

						portUint, parseErr := strconv.ParseUint(portStr, 10, 16)
						if parseErr != nil {
							return nil, fmt.Errorf("invalid port: %v", parseErr)
						}

						// 创建 mihomo metadata
						metadata := &constant.Metadata{
							Host:    host,
							DstPort: uint16(portUint),
							Type:    constant.HTTP,
						}

						// 使用 mihomo adapter 建立连接
						return proxyAdapter.DialContext(ctx, metadata)
					},
					TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
				}
			}
		} else {
			utils.Warn("未找到可用代理，将直接下载")
		}
	}

	// 创建请求并设置 User-Agent
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		utils.Error("URL %s，创建请求失败:  %v", urlStr, err)
		return nil, err
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
			Data: map[string]interface{}{
				"id":       id,
				"name":     subName,
				"status":   "error",
				"error":    err.Error(),
				"tlsError": isTLSError(err),
			},
		})
		return nil, err
	}
	defer resp.Body.Close()

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
			Data: map[string]interface{}{
				"id":     id,
				"name":   subName,
				"status": "error",
				"error":  err.Error(),
			},
		})
		return nil, err
	}
	var config ClashConfig
	// 尝试解析 YAML
	errYaml := yaml.Unmarshal(data, &config)

	// 如果 YAML 解析失败或没有代理节点，尝试 Base64 解码 兼容base64订阅
	if errYaml != nil || len(config.Proxies) == 0 {
		// 尝试标准 Base64 解码
		decodedBytes, errB64 := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
		if errB64 != nil {
			// 尝试 Raw Base64 (无填充) 解码
			decodedBytes, errB64 = base64.RawStdEncoding.DecodeString(strings.TrimSpace(string(data)))
		}

		if errB64 == nil {
			// Base64 解码成功，按行解析
			lines := strings.Split(string(decodedBytes), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				proxy, errP := protocol.LinkToProxy(protocol.Urls{Url: line}, protocol.OutputConfig{})
				if errP == nil {
					config.Proxies = append(config.Proxies, proxy)
				}
			}
		}
		// 兼容非base64的v2ray配置文件
		if len(config.Proxies) == 0 {
			// Base64 解码成功，按行解析
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				proxy, errP := protocol.LinkToProxy(protocol.Urls{Url: line}, protocol.OutputConfig{})
				if errP == nil {
					config.Proxies = append(config.Proxies, proxy)
				}
			}
		}
	}

	if len(config.Proxies) == 0 {
		utils.Error("URL %s，解析失败或未找到节点 (YAML error: %v)", urlStr, errYaml)
		// 发送解析失败通知
		notifications.Publish("subscription.sync_failed", notifications.Payload{
			Title:   "订阅更新失败",
			Message: fmt.Sprintf("❌订阅【%s】解析失败或未找到节点", subName),
			Data: map[string]interface{}{
				"id":     id,
				"name":   subName,
				"status": "error",
				"error":  "解析失败或未找到节点",
			},
		})
		return nil, fmt.Errorf("解析失败 or 未找到节点")
	}

	err = scheduleClashToNodeLinks(id, config.Proxies, subName, reporter, usageInfo)
	return usageInfo, err
}

// scheduleClashToNodeLinks 将 Clash 代理配置转换为节点链接并保存到数据库
// id: 订阅ID
// proxys: 代理节点列表
// subName: 订阅名称
// usageInfo: 订阅用量信息 (可选)
func scheduleClashToNodeLinks(id int, proxys []protocol.Proxy, subName string, reporter TaskReporter, usageInfo *UsageInfo) error {
	if reporter == nil {
		reporter = &NoOpTaskReporter{}
	}

	addSuccessCount := 0
	updateCount := 0 // 名称/链接已更新的节点数量
	skipCount := 0   // 已存在的节点数量（跳过）
	processedCount := 0
	startTime := time.Now() // 记录开始时间用于计算耗时

	// 确保任务结束时处理异常
	defer func() {
		if r := recover(); r != nil {
			utils.Error("订阅更新任务执行过程中发生严重错误: %v", r)
			reporter.ReportFail(fmt.Sprintf("任务异常: %v", r))
		}
	}()

	// 获取机场的Group信息
	airport, err := models.GetAirportByID(id)
	if err != nil {
		utils.Error("获取机场 %s 的Group失败:  %v", subName, err)
	}

	// 应用机场节点过滤和重命名规则
	if airport != nil {
		originalCount := len(proxys)
		proxys = applyAirportNodeFilter(airport, proxys)
		if len(proxys) < originalCount {
			utils.Info("📦订阅【%s】过滤后节点数量：%d（原始：%d，过滤掉：%d）", subName, len(proxys), originalCount, originalCount-len(proxys))
		}
		// 应用高级去重规则
		beforeDedup := len(proxys)
		proxys = applyAirportDeduplication(airport, proxys)
		if len(proxys) < beforeDedup {
			utils.Info("🔄订阅【%s】去重后节点数量：%d（去重前：%d，去重掉：%d）", subName, len(proxys), beforeDedup, beforeDedup-len(proxys))
		}
		//节点重命名
		proxys = applyAirportNodeRename(airport, proxys)
		// 节点名称唯一化（添加机场标识前缀，防止多机场节点重名）
		proxys = applyAirportNodeUniquify(airport, proxys)
	}

	// 1. 获取该订阅当前在数据库中的所有节点
	existingNodes, err := models.ListBySourceID(id)
	if err != nil {
		utils.Info("获取订阅【%s】现有节点失败: %v", subName, err)
		existingNodes = []models.Node{} // 确保后续逻辑不会panic
	}

	// 创建现有节点的映射表（以 ID 为键，用于删除判断时遍历）
	existingNodeByID := make(map[int]models.Node)
	for _, node := range existingNodes {
		existingNodeByID[node.ID] = node
	}

	// 预扫描：统计本次拉取中每个 ContentHash 对应的名称集合（trim 后）
	// 用于识别同 hash 多名称的信息节点（如"到期时间"、"剩余流量"），这类节点应放行入库且需要按名称粒度清理
	currentNamesByHash := make(map[string]map[string]bool)
	for _, p := range proxys {
		name := strings.TrimSpace(p.Name)
		p.Server = utils.WrapIPv6Host(p.Server)
		ch := protocol.GenerateProxyContentHash(p)
		if ch == "" {
			continue
		}
		if currentNamesByHash[ch] == nil {
			currentNamesByHash[ch] = make(map[string]bool)
		}
		currentNamesByHash[ch][name] = true
	}

	// 统计数据库中本机场已有节点的 hash→名称集合
	// 解决“历史上是信息节点，但本次拉取只剩一个名称”时无法识别的问题（否则会导致残留无法清理/误更新）。
	existingNamesByHash := make(map[string]map[string]bool)
	for _, node := range existingNodes {
		if node.ContentHash == "" {
			continue
		}
		name := strings.TrimSpace(node.Name)
		if existingNamesByHash[node.ContentHash] == nil {
			existingNamesByHash[node.ContentHash] = make(map[string]bool)
		}
		existingNamesByHash[node.ContentHash][name] = true
	}

	// 信息节点 hash 集合：本次拉取或数据库历史中，同一 hash 对应多个不同名称
	infoNodeHashes := make(map[string]bool)
	for ch, names := range currentNamesByHash {
		if len(names) > 1 {
			infoNodeHashes[ch] = true
		}
	}
	for ch, names := range existingNamesByHash {
		if len(names) > 1 {
			infoNodeHashes[ch] = true
		}
	}

	// 创建现有节点的映射表（以 ContentHash 为键，用于同机场去重判断与更新）
	existingNodeByContentHash := make(map[string]models.Node)
	// 对信息节点 hash，记录本机场已有的所有名称（用于重新拉取时精确匹配）
	existingInfoNodeNames := make(map[string]map[string]models.Node)
	for _, node := range existingNodes {
		if node.ContentHash != "" {
			existingNodeByContentHash[node.ContentHash] = node
			// 如果该 hash 是信息节点，按名称建立索引
			if infoNodeHashes[node.ContentHash] {
				if existingInfoNodeNames[node.ContentHash] == nil {
					existingInfoNodeNames[node.ContentHash] = make(map[string]models.Node)
				}
				existingInfoNodeNames[node.ContentHash][strings.TrimSpace(node.Name)] = node
			}
		}
	}

	// 读取全局配置：是否启用跨机场去重（默认启用）
	crossAirportDedupVal, _ := models.GetSetting("cross_airport_dedup_enabled")
	enableCrossDedup := crossAirportDedupVal != "false"

	var allNodeHashes map[string]bool
	if enableCrossDedup {
		allNodeHashes = models.GetAllNodeContentHashes()
	} else {
		allNodeHashes = models.GetNodeContentHashesBySourceID(id)
	}

	utils.Info("📄订阅【%s】获取到订阅数量【%d】，现有节点数量【%d】，哈希数量【%d】，跨机场去重【%v】", subName, len(proxys), len(existingNodes), len(allNodeHashes), enableCrossDedup)

	// 更新任务总数（此时已知道需要处理的节点数量）
	reporter.UpdateTotal(len(proxys))

	// 记录本次获取到的节点 ContentHash（用于判断需要删除的节点）
	currentHashes := make(map[string]bool)

	// 批量收集：新增节点列表（稍后批量写入）
	nodesToAdd := make([]models.Node, 0)

	// 批量收集：需要更新名称/链接的节点列表
	nodesToUpdate := make([]models.NodeInfoUpdate, 0)

	// 2. 遍历新获取的节点，插入或更新
	for proxyIndex, proxy := range proxys {
		utils.Info("💾准备存储节点【%s】", proxy.Name)
		var Node models.Node

		// 预处理：去除名称空格，处理 IPv6 地址
		proxy.Name = strings.TrimSpace(proxy.Name)
		proxy.Server = utils.WrapIPv6Host(proxy.Server)

		// 计算节点内容哈希（用于全库去重）
		contentHash := protocol.GenerateProxyContentHash(proxy)
		if contentHash == "" {
			utils.Warn("节点【%s】生成内容哈希失败，跳过", proxy.Name)
			continue
		}

		// 使用公共函数生成节点链接
		link := GenerateProxyLink(proxy)
		if link == "" {
			utils.Warn("节点【%s】生成链接失败，跳过", proxy.Name)
			continue
		}

		Node.Link = link
		Node.Name = proxy.Name
		Node.LinkName = proxy.Name
		Node.LinkAddress = proxy.Server + ":" + strconv.Itoa(int(proxy.Port))
		Node.LinkHost = proxy.Server
		Node.LinkPort = strconv.Itoa(int(proxy.Port))
		Node.Source = subName
		Node.SourceID = id
		Node.SourceSort = proxyIndex + 1
		if airport != nil {
			Node.Group = airport.Group
		}
		Node.Protocol = proxy.Type
		Node.ContentHash = contentHash

		// 记录本次获取到的节点 ContentHash
		currentHashes[contentHash] = true

		// 判断节点是否已存在（全库去重：使用 ContentHash 判断）
		var nodeStatus string
		if allNodeHashes[contentHash] {
			skipCount++
			nodeStatus = "skipped"
			// 节点内容已存在 - 优先判断是否为本机场已存在节点
			if _, ok := existingNodeByContentHash[contentHash]; ok {
				// 属于本机场
				if infoNodeHashes[contentHash] {
					// 信息节点：用名称精确匹配（同 hash 对应多个已有节点）
					if existingByName, nameExists := existingInfoNodeNames[contentHash][proxy.Name]; nameExists {
						// 该名称的信息节点已存在，检查链接或顺序是否变化
						if existingByName.Link != link || existingByName.SourceSort != Node.SourceSort {
							nodesToUpdate = append(nodesToUpdate, models.NodeInfoUpdate{
								ID:         existingByName.ID,
								Name:       proxy.Name,
								LinkName:   proxy.Name,
								Link:       link,
								SourceSort: Node.SourceSort,
							})
							updateCount++
							nodeStatus = "updated"
							utils.Info("✏️ 信息节点【%s】链接/顺序已变更，将更新", proxy.Name)
						} else {
							utils.Debug("⏭️ 信息节点【%s】在本机场已存在，跳过", proxy.Name)
						}
					} else {
						// 该名称的信息节点不存在（上游新增了一个信息节点），入库
						nodesToAdd = append(nodesToAdd, Node)
						skipCount--
						addSuccessCount++
						nodeStatus = "added"
						utils.Info("📌 信息节点【%s】为新名称，允许入库", proxy.Name)
					}
				} else {
					// 普通节点：用 hash 匹配，检查名称或链接是否变化
					existingNode := existingNodeByContentHash[contentHash]
					if existingNode.Name != proxy.Name || existingNode.Link != link || existingNode.SourceSort != Node.SourceSort {
						nodesToUpdate = append(nodesToUpdate, models.NodeInfoUpdate{
							ID:         existingNode.ID,
							Name:       proxy.Name,
							LinkName:   proxy.Name,
							Link:       link,
							SourceSort: Node.SourceSort,
						})
						updateCount++
						nodeStatus = "updated"
						utils.Info("✏️ 节点【%s】名称/链接/顺序已变更，将更新 [旧名称: %s]", proxy.Name, existingNode.Name)
					} else {
						utils.Debug("⏭️ 节点【%s】在本机场已存在，跳过", proxy.Name)
					}
				}
			} else if enableCrossDedup {
				// 跨机场去重开启：若全库已存在该内容，则跳过
				if existingNode, exists := models.GetNodeByContentHash(contentHash); exists {
					// 检查是否为同 hash 不同名的信息节点（预扫描已确定）
					if infoNodeHashes[contentHash] {
						nodesToAdd = append(nodesToAdd, Node)
						skipCount--
						addSuccessCount++
						nodeStatus = "added"
						utils.Info("📌 节点【%s】与已有节点配置相同但名称不同（信息节点），允许入库 [已有: %s]", proxy.Name, existingNode.Name)
					} else {
						utils.Warn("⚠️ 节点【%s】与其他机场重复，跳过 [现有节点: %s] [来源: %s] [分组: %s] [SourceID: %d]", proxy.Name, existingNode.Name, existingNode.Source, existingNode.Group, existingNode.SourceID)
					}
				} else {
					// hash存在于allNodeHashes但缓存中找不到，说明是本次拉取中的内部重复
					if infoNodeHashes[contentHash] {
						nodesToAdd = append(nodesToAdd, Node)
						skipCount--
						addSuccessCount++
						nodeStatus = "added"
						allNodeHashes[contentHash] = true
						utils.Info("📌 节点【%s】与本次拉取中其他节点配置相同但名称不同（信息节点），允许入库", proxy.Name)
					} else {
						hashData := protocol.NormalizeProxyForHash(proxy)
						jsonBytes, _ := json.Marshal(hashData)
						utils.Warn("🔄 节点【%s】与本次拉取中的其他节点重复（相同配置），跳过\n    HashData: %s", proxy.Name, string(jsonBytes))
					}
				}
			} else {
				// 跨机场去重关闭时 allNodeHashes 只包含本机场哈希；若此处找不到现有节点，说明是本次拉取内重复
				if infoNodeHashes[contentHash] {
					nodesToAdd = append(nodesToAdd, Node)
					skipCount--
					addSuccessCount++
					nodeStatus = "added"
					allNodeHashes[contentHash] = true
					utils.Info("📌 节点【%s】与本次拉取中其他节点配置相同但名称不同（信息节点），允许入库", proxy.Name)
				} else {
					hashData := protocol.NormalizeProxyForHash(proxy)
					jsonBytes, _ := json.Marshal(hashData)
					utils.Warn("🔄 节点【%s】与本次拉取中的其他节点重复（相同配置），跳过\n    HashData: %s", proxy.Name, string(jsonBytes))
				}
			}
		} else {
			// 节点不存在，收集到待添加列表
			nodesToAdd = append(nodesToAdd, Node)
			addSuccessCount++
			nodeStatus = "added"
			// 将新节点的 hash 加入全库集合，避免本次拉取内的重复
			allNodeHashes[contentHash] = true
		}

		// 更新进度（通过 reporter 报告）
		processedCount++
		reporter.ReportProgress(processedCount, proxy.Name, map[string]interface{}{
			"status":  nodeStatus,
			"added":   addSuccessCount,
			"skipped": skipCount,
		})
	}

	// 3. 收集需要删除的节点ID（本次订阅没有获取到但数据库中存在的节点）
	nodeIDsToDelete := make([]int, 0)
	for nodeID, node := range existingNodeByID {
		// 使用 ContentHash 判断节点是否在本次拉取中
		if !currentHashes[node.ContentHash] {
			nodeIDsToDelete = append(nodeIDsToDelete, nodeID)
			continue
		}

		// 信息节点：hash 仍在，但需要按名称精细判断，避免名称变化/部分移除导致垃圾节点残留（数据膨胀）
		if infoNodeHashes[node.ContentHash] {
			currentNames := currentNamesByHash[node.ContentHash]
			if len(currentNames) == 0 || !currentNames[strings.TrimSpace(node.Name)] {
				nodeIDsToDelete = append(nodeIDsToDelete, nodeID)
			}
		}
	}

	// 4. 批量写入数据库（一次性操作，减少数据库I/O）
	// 批量添加新节点
	if len(nodesToAdd) > 0 {
		if err := models.BatchAddNodes(nodesToAdd); err != nil {
			utils.Error("❌批量添加节点失败：%v", err)
			// 重置计数，因为添加失败
			addSuccessCount = 0
		} else {
			utils.Info("✅批量添加 %d 个节点成功", len(nodesToAdd))
		}
	}

	// 批量更新名称/链接已变更的节点
	actualUpdateCount := 0
	if len(nodesToUpdate) > 0 {
		if cnt, err := models.BatchUpdateNodeInfo(nodesToUpdate); err != nil {
			utils.Error("❌批量更新节点信息失败：%v", err)
		} else {
			actualUpdateCount = cnt
			utils.Info("✏️批量更新 %d 个节点的名称/链接", actualUpdateCount)
		}
	}

	// 批量删除失效节点
	deleteCount := 0
	if len(nodeIDsToDelete) > 0 {
		if err := models.BatchDel(nodeIDsToDelete); err != nil {
			utils.Error("❌批量删除节点失败：%v", err)
		} else {
			deleteCount = len(nodeIDsToDelete)
			utils.Info("🗑️批量删除 %d 个失效节点", deleteCount)
		}
	}

	utils.Info("✅订阅【%s】节点同步完成，总节点【%d】个，成功处理【%d】个，新增节点【%d】个，更新节点【%d】个，已存在节点【%d】个，删除失效【%d】个", subName, len(proxys), addSuccessCount+skipCount, addSuccessCount, actualUpdateCount, skipCount, deleteCount)
	// 重新查找机场以获取最新信息并更新成功次数
	airport, err = models.GetAirportByID(id)
	if err != nil {
		utils.Error("获取机场 %s 失败:  %v", subName, err)
		return err
	}
	airport.SuccessCount = addSuccessCount + skipCount
	// 当前时间
	now := time.Now()
	airport.LastRunTime = &now
	err1 := airport.Update()
	if err1 != nil {
		return err1
	}
	// 通过 reporter 报告任务完成
	reporter.ReportComplete(fmt.Sprintf("订阅更新完成 (新增: %d, 更新: %d, 已存在: %d, 删除: %d)", addSuccessCount, actualUpdateCount, skipCount, deleteCount), map[string]interface{}{
		"added":   addSuccessCount,
		"updated": actualUpdateCount,
		"skipped": skipCount,
		"deleted": deleteCount,
	})

	// 触发webhook的完成事件
	duration := time.Since(startTime)
	durationStr := formatDurationSub(duration)

	// 构建用量信息文本
	var usageText string
	usageData := make(map[string]interface{})
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

	nData := map[string]interface{}{
		"id":       id,
		"name":     subName,
		"status":   "success",
		"success":  addSuccessCount + skipCount,
		"duration": duration.Milliseconds(),
	}
	if len(usageData) > 0 {
		nData["usage"] = usageData
	}

	notifications.Publish("subscription.sync_succeeded", notifications.Payload{
		Title:   "订阅更新完成",
		Message: fmt.Sprintf("✅订阅【%s】节点同步完成，耗时 %s，总节点【%d】个，成功处理【%d】个，新增节点【%d】个，更新节点【%d】个，已存在节点【%d】个，删除失效【%d】个%s", subName, durationStr, len(proxys), addSuccessCount+skipCount, addSuccessCount, actualUpdateCount, skipCount, deleteCount, usageText),
		Data:    nData,
	})
	return nil

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
		nodeProto := strings.ToLower(proxy.Type)

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
		protoType := strings.ToLower(proxy.Type)
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

// GenerateProxyLink 从 Proxy 结构体生成节点链接
func GenerateProxyLink(proxy protocol.Proxy) string {
	proxy.Name = strings.TrimSpace(proxy.Name)
	proxy.Server = utils.WrapIPv6Host(proxy.Server)
	link, err := protocol.EncodeProxyLink(proxy)
	if err != nil {
		return ""
	}
	return link
}

// applyAirportNodeUniquify 应用机场节点名称唯一化
// 在节点名称前添加机场标识前缀，防止多机场间节点名称重复
// 同一机场同一节点每次生成的名字保持一致（使用机场ID生成稳定前缀）
func applyAirportNodeUniquify(airport *models.Airport, proxys []protocol.Proxy) []protocol.Proxy {
	if airport == nil || !airport.NodeNameUniquify {
		return proxys
	}

	// 生成前缀: 使用用户自定义前缀 或 默认的 [A{id}] 格式
	var prefix string
	if airport.NodeNamePrefix != "" {
		prefix = airport.NodeNamePrefix
	} else {
		prefix = fmt.Sprintf("[A%d]", airport.ID)
	}

	// 为每个节点名称添加前缀
	for i := range proxys {
		proxys[i].Name = prefix + proxys[i].Name
	}

	return proxys
}

// parseProtoFromLink 根据协议类型解析链接获取结构体
func parseProtoFromLink(link string, protoType string) (interface{}, error) {
	protoObj, detectedProto, err := protocol.DecodeProtocolObject(link)
	if err != nil {
		return nil, err
	}
	if detectedProto != protoType {
		return nil, fmt.Errorf("unsupported protocol: %s", protoType)
	}
	return protoObj, nil
}
