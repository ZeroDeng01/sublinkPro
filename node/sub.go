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
	"net/url"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/services/mihomo"
	"sublink/services/sse"
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
	return LoadClashConfigFromURLWithReporter(id, urlStr, subName, downloadWithProxy, proxyLink, userAgent, nil, false, true)
}

// LoadClashConfigFromURLWithReporter 从指定 URL 加载 Clash 配置（带任务报告器）
// reporter: 任务进度报告器，用于TaskManager集成
// fetchUsageInfo: 是否获取用量信息
// skipTLSVerify: 是否跳过TLS证书验证
func LoadClashConfigFromURLWithReporter(id int, urlStr string, subName string, downloadWithProxy bool, proxyLink string, userAgent string, reporter TaskReporter, fetchUsageInfo bool, skipTLSVerify bool) (*UsageInfo, error) {
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

						portInt, atoiErr := strconv.Atoi(portStr)
						if atoiErr != nil {
							return nil, fmt.Errorf("invalid port: %v", atoiErr)
						}

						// 验证端口范围
						if portInt < 0 || portInt > 65535 {
							return nil, fmt.Errorf("port out of range: %d", portInt)
						}

						// 创建 mihomo metadata
						metadata := &constant.Metadata{
							Host:    host,
							DstPort: uint16(portInt),
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

	// 设置 User-Agent
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

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
		sse.GetSSEBroker().BroadcastEvent("sub_update", sse.NotificationPayload{
			Event:   "sub_update",
			Title:   title,
			Message: message,
			Data: map[string]interface{}{
				"id":       id,
				"name":     subName,
				"status":   "failed",
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
		sse.GetSSEBroker().BroadcastEvent("sub_update", sse.NotificationPayload{
			Event:   "sub_update",
			Title:   "订阅更新失败",
			Message: fmt.Sprintf("❌订阅【%s】读取响应失败: %v", subName, err),
			Data: map[string]interface{}{
				"id":     id,
				"name":   subName,
				"status": "failed",
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
		sse.GetSSEBroker().BroadcastEvent("sub_update", sse.NotificationPayload{
			Event:   "sub_update",
			Title:   "订阅更新失败",
			Message: fmt.Sprintf("❌订阅【%s】解析失败或未找到节点", subName),
			Data: map[string]interface{}{
				"id":     id,
				"name":   subName,
				"status": "failed",
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
	skipCount := 0 // 已存在的节点数量（跳过）
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

	// 1. 获取该订阅当前在数据库中的所有节点
	existingNodes, err := models.ListBySourceID(id)
	if err != nil {
		utils.Info("获取订阅【%s】现有节点失败: %v", subName, err)
		existingNodes = []models.Node{} // 确保后续逻辑不会panic
	}

	// 创建现有节点的映射表（以Link为键）
	existingNodeMap := make(map[string]models.Node)
	for _, node := range existingNodes {
		existingNodeMap[node.Link] = node
	}

	utils.Info("📄订阅【%s】获取到订阅数量【%d】，现有节点数量【%d】", subName, len(proxys), len(existingNodes))

	// 更新任务总数（此时已知道需要处理的节点数量）
	reporter.UpdateTotal(len(proxys))

	// 记录本次获取到的节点Link
	currentLinks := make(map[string]bool)

	// 批量收集：新增节点列表（稍后批量写入）
	nodesToAdd := make([]models.Node, 0)

	// 2. 遍历新获取的节点，插入或更新
	for _, proxy := range proxys {
		utils.Info("💾准备存储节点【%s】", proxy.Name)
		var Node models.Node
		var link string
		//var systemNodeName = subName + "_" + strings.TrimSpace(proxy.Name) //系统节点名称
		proxy.Name = strings.TrimSpace(proxy.Name) // 某些机场的节点名称可能包含空格
		proxy.Server = utils.WrapIPv6Host(proxy.Server)
		switch strings.ToLower(proxy.Type) {
		case "ss":
			// ss://method:password@server:port#name
			method := proxy.Cipher
			password := proxy.Password
			server := proxy.Server
			port := int(proxy.Port)
			name := proxy.Name
			encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", method, password)))
			link = fmt.Sprintf("ss://%s@%s:%d#%s", encoded, server, port, name)
		case "ssr":
			// ssr://server:port:protocol:method:obfs:base64(password)/?remarks=base64(remarks)&obfsparam=base64(obfsparam)
			server := proxy.Server
			port := int(proxy.Port)
			protocol := proxy.Protocol
			method := proxy.Cipher
			obfs := proxy.Obfs
			password := base64.StdEncoding.EncodeToString([]byte(proxy.Password))
			remarks := base64.StdEncoding.EncodeToString([]byte(proxy.Name))
			obfsparam := ""
			if proxy.Obfs_password != "" {
				obfsparam = base64.StdEncoding.EncodeToString([]byte(proxy.Obfs_password))
			}
			params := fmt.Sprintf("remarks=%s", remarks)
			if obfsparam != "" {
				params += fmt.Sprintf("&obfsparam=%s", obfsparam)
			}
			data := fmt.Sprintf("%s:%d:%s:%s:%s:%s/?%s", server, port, protocol, method, obfs, password, params)
			link = fmt.Sprintf("ssr://%s", base64.StdEncoding.EncodeToString([]byte(data)))

		case "trojan":
			// trojan://password@server:port?参数#name
			password := proxy.Password
			server := proxy.Server
			port := int(proxy.Port)
			name := proxy.Name
			query := url.Values{}

			// 添加所有Clash配置中的参数
			if proxy.Sni != "" {
				query.Set("sni", proxy.Sni)
			}

			// 处理Peer参数，通常与SNI相同
			if proxy.Peer != "" {
				query.Set("peer", proxy.Peer)
			}

			// 处理跳过证书验证
			if proxy.Skip_cert_verify {
				query.Set("allowInsecure", "1")
			}

			// 处理网络类型
			if proxy.Network != "" {
				query.Set("type", proxy.Network)
			}

			// 处理客户端指纹
			if proxy.Client_fingerprint != "" {
				query.Set("fp", proxy.Client_fingerprint)
			}

			// 处理ALPN
			if len(proxy.Alpn) > 0 {
				query.Set("alpn", strings.Join(proxy.Alpn, ","))
			}

			// 处理Flow
			if proxy.Flow != "" {
				query.Set("flow", proxy.Flow)
			}

			// 处理WebSocket选项
			if len(proxy.Ws_opts) > 0 {
				if path, ok := proxy.Ws_opts["path"].(string); ok && path != "" {
					query.Set("path", path)
				}

				if headers, ok := proxy.Ws_opts["headers"].(map[string]interface{}); ok {
					if host, ok := headers["Host"].(string); ok && host != "" {
						query.Set("host", host)
					}
				}
			}

			// 处理Reality选项
			if len(proxy.Reality_opts) > 0 {
				if publicKey, ok := proxy.Reality_opts["public-key"].(string); ok && publicKey != "" {
					query.Set("pbk", publicKey)
				}

				if shortId, ok := proxy.Reality_opts["short-id"].(string); ok && shortId != "" {
					query.Set("sid", shortId)
				}
			}

			// 构建最终URL
			queryStr := query.Encode()
			if queryStr != "" {
				link = fmt.Sprintf("trojan://%s@%s:%d?%s#%s", password, server, port, queryStr, name)
			} else {
				link = fmt.Sprintf("trojan://%s@%s:%d#%s", password, server, port, name)
			}

		case "vmess":
			// vmess://base64(json)
			vmessObj := map[string]interface{}{
				"v":    "2",
				"ps":   proxy.Name,
				"add":  proxy.Server,
				"port": proxy.Port,
				"id":   proxy.Uuid,
				"scy":  proxy.Cipher,
			}
			if proxy.AlterId != "" {
				aid, _ := strconv.Atoi(proxy.AlterId)
				vmessObj["aid"] = aid
			} else {
				vmessObj["aid"] = 0
			}
			vmessObj["net"] = proxy.Network
			if proxy.Tls {
				vmessObj["tls"] = "tls"
			} else {
				vmessObj["tls"] = "none"
			}
			if len(proxy.Ws_opts) > 0 {
				if path, ok := proxy.Ws_opts["path"].(string); ok {
					vmessObj["path"] = path
				}
				if headers, ok := proxy.Ws_opts["headers"].(map[string]interface{}); ok {
					if host, ok := headers["Host"].(string); ok {
						vmessObj["host"] = host
					}
				}
			}
			jsonData, _ := json.Marshal(vmessObj)
			link = fmt.Sprintf("vmess://%s", base64.StdEncoding.EncodeToString(jsonData))

		case "vless":
			// vless://uuid@server:port?参数#name
			uuid := proxy.Uuid
			server := proxy.Server
			port := int(proxy.Port)
			name := proxy.Name
			query := url.Values{}

			// 处理网络类型
			if proxy.Network != "" {
				query.Set("type", proxy.Network)
			}

			// 处理TLS设置
			if proxy.Tls {
				query.Set("security", "tls")
			} else {
				query.Set("security", "none")
			}

			// 处理SNI设置(servername)
			if proxy.Servername != "" {
				query.Set("sni", proxy.Servername)
			}

			// 处理客户端指纹
			if proxy.Client_fingerprint != "" {
				query.Set("fp", proxy.Client_fingerprint)
			}

			// 处理Flow控制
			if proxy.Flow != "" {
				query.Set("flow", proxy.Flow)
			}

			// 处理跳过证书验证
			if proxy.Skip_cert_verify {
				query.Set("allowInsecure", "1")
			}

			// 处理ALPN
			if len(proxy.Alpn) > 0 {
				query.Set("alpn", strings.Join(proxy.Alpn, ","))
			}

			// 处理WebSocket选项
			if len(proxy.Ws_opts) > 0 {
				if path, ok := proxy.Ws_opts["path"].(string); ok && path != "" {
					query.Set("path", path)
				}
				if headers, ok := proxy.Ws_opts["headers"].(map[string]interface{}); ok {
					if host, ok := headers["Host"].(string); ok && host != "" {
						query.Set("host", host)
					}
				}
			}

			// 处理Reality选项
			if len(proxy.Reality_opts) > 0 {
				if pbk, ok := proxy.Reality_opts["public-key"].(string); ok && pbk != "" {
					query.Set("pbk", pbk)
				}
				if sid, ok := proxy.Reality_opts["short-id"].(string); ok && sid != "" {
					query.Set("sid", sid)
				}
			}

			// 处理GRPC选项
			if len(proxy.Grpc_opts) > 0 {
				query.Set("security", "reality")
				if sn, ok := proxy.Grpc_opts["grpc-service-name"].(string); ok && sn != "" {
					query.Set("serviceName", sn)
				}
				if mode, ok := proxy.Grpc_opts["grpc-mode"].(string); ok && mode == "multi" {
					query.Set("mode", "multi")
				}
			}

			// 构建最终URL
			queryStr := query.Encode()
			if queryStr != "" {
				link = fmt.Sprintf("vless://%s@%s:%d?%s#%s", uuid, server, port, queryStr, name)
			} else {
				link = fmt.Sprintf("vless://%s@%s:%d#%s", uuid, server, port, name)
			}

		case "hysteria":
			// hysteria://server:port?protocol=udp&auth=auth&peer=peer&insecure=1&upmbps=up&downmbps=down&alpn=alpn#name
			server := proxy.Server
			port := int(proxy.Port)
			name := proxy.Name
			query := url.Values{}
			query.Set("protocol", "udp")
			if proxy.Auth_str != "" {
				query.Set("auth", proxy.Auth_str)
			}
			if proxy.Peer != "" {
				query.Set("peer", proxy.Peer)
			}
			if proxy.Skip_cert_verify {
				query.Set("insecure", "1")
			}
			if proxy.Up > 0 {
				query.Set("upmbps", strconv.Itoa(proxy.Up))
			}
			if proxy.Down > 0 {
				query.Set("downmbps", strconv.Itoa(proxy.Down))
			}
			if len(proxy.Alpn) > 0 {
				query.Set("alpn", strings.Join(proxy.Alpn, ","))
			}
			link = fmt.Sprintf("hysteria://%s:%d?%s#%s", server, port, query.Encode(), name)

		case "hysteria2":
			// hysteria2://auth@server:port?sni=sni&insecure=1&obfs=obfs&obfs-password=obfs-password&mport=ports&upmbps=up&downmbps=down&fp=fingerprint#name
			server := proxy.Server
			port := int(proxy.Port)
			auth := proxy.Password
			name := proxy.Name
			query := url.Values{}
			// SNI: 优先使用 Sni，如果为空则使用 Servername
			if proxy.Sni != "" {
				query.Set("sni", proxy.Sni)
			} else if proxy.Servername != "" {
				query.Set("sni", proxy.Servername)
			}
			// 跳过证书验证
			if proxy.Skip_cert_verify {
				query.Set("insecure", "1")
			}
			// 混淆
			if proxy.Obfs != "" {
				query.Set("obfs", proxy.Obfs)
			}
			if proxy.Obfs_password != "" {
				query.Set("obfs-password", proxy.Obfs_password)
			}
			// ALPN
			if len(proxy.Alpn) > 0 {
				query.Set("alpn", strings.Join(proxy.Alpn, ","))
			}
			// 端口跳跃 (ports -> mport)
			if proxy.Ports != "" {
				query.Set("mport", proxy.Ports)
			}
			// 上行带宽
			if proxy.Up > 0 {
				query.Set("upmbps", strconv.Itoa(proxy.Up))
			}
			// 下行带宽
			if proxy.Down > 0 {
				query.Set("downmbps", strconv.Itoa(proxy.Down))
			}
			// 客户端指纹
			if proxy.Client_fingerprint != "" {
				query.Set("fp", proxy.Client_fingerprint)
			}
			link = fmt.Sprintf("hysteria2://%s@%s:%d?%s#%s", auth, server, port, query.Encode(), name)

		case "tuic":
			// tuic://uuid:password@server:port?sni=sni&congestion_control=congestion_control&alpn=alpn#name
			uuid := proxy.Uuid
			password := proxy.Password
			server := proxy.Server
			port := int(proxy.Port)
			name := proxy.Name
			query := url.Values{}
			if proxy.Sni != "" {
				query.Set("sni", proxy.Sni)
			}
			if proxy.Congestion_control != "" {
				query.Set("congestion_control", proxy.Congestion_control)
			}
			if len(proxy.Alpn) > 0 {
				query.Set("alpn", strings.Join(proxy.Alpn, ","))
			}
			if proxy.Udp_relay_mode != "" {
				query.Set("udp_relay_mode", proxy.Udp_relay_mode)
			}
			if proxy.Disable_sni {
				query.Set("disable_sni", "1")
			}
			link = fmt.Sprintf("tuic://%s:%s@%s:%d?%s#%s", uuid, password, server, port, query.Encode(), name)

		case "anytls":
			// anytls://password@server:port?sni=sni&insecure=1&fp=chrome#anytls_name

			password := proxy.Password
			server := proxy.Server
			port := int(proxy.Port)
			name := proxy.Name
			query := url.Values{}
			if proxy.Sni != "" {
				query.Set("sni", proxy.Sni)
			}
			if proxy.Skip_cert_verify {
				query.Set("insecure", "1")
			}
			if proxy.Client_fingerprint != "" {
				query.Set("fp", proxy.Client_fingerprint)
			}

			link = fmt.Sprintf("anytls://%s@%s:%d?%s#%s", password, server, port, query.Encode(), name)

		case "socks5":
			// socks5://username:password@server:port#name
			username := proxy.Username
			password := proxy.Password
			server := proxy.Server
			port := int(proxy.Port)
			name := proxy.Name
			if username != "" && password != "" {
				link = fmt.Sprintf("socks5://%s:%s@%s:%d#%s", username, password, server, port, name)
			} else {
				link = fmt.Sprintf("socks5://%s:%d#%s", server, port, name)
			}

		}
		Node.Link = link
		Node.Name = proxy.Name
		Node.LinkName = proxy.Name
		Node.LinkAddress = proxy.Server + ":" + strconv.Itoa(int(proxy.Port))
		Node.LinkHost = proxy.Server
		Node.LinkPort = strconv.Itoa(int(proxy.Port))
		Node.Source = subName
		Node.SourceID = id
		Node.Group = airport.Group
		Node.Protocol = proxy.Type
		// 记录本次获取到的节点
		currentLinks[link] = true

		// 判断节点是否已存在 - 收集到内存，稍后批量写入
		var nodeStatus string
		if _, exists := existingNodeMap[link]; exists {
			skipCount++
			nodeStatus = "skipped"
			// 已存在的节点跳过，不做任何处理
		} else {
			// 节点不存在，收集到待添加列表
			nodesToAdd = append(nodesToAdd, Node)
			addSuccessCount++
			nodeStatus = "added"
		}

		// 更新进度（通过 reporter 报告）- 基于内存计数，保持实时性
		processedCount++
		reporter.ReportProgress(processedCount, proxy.Name, map[string]interface{}{
			"status":  nodeStatus,
			"added":   addSuccessCount,
			"skipped": skipCount,
		})
	}

	// 3. 收集需要删除的节点ID（本次订阅没有获取到但数据库中存在的节点）
	nodeIDsToDelete := make([]int, 0)
	for link, existingNode := range existingNodeMap {
		if !currentLinks[link] {
			// 该节点不在本次订阅中，需要删除
			nodeIDsToDelete = append(nodeIDsToDelete, existingNode.ID)
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

	utils.Info("✅订阅【%s】节点同步完成，总节点【%d】个，成功处理【%d】个，新增节点【%d】个，已存在节点【%d】个，删除失效【%d】个", subName, len(proxys), addSuccessCount+skipCount, addSuccessCount, skipCount, deleteCount)
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
	reporter.ReportComplete(fmt.Sprintf("订阅更新完成 (新增: %d, 已存在: %d, 删除: %d)", addSuccessCount, skipCount, deleteCount), map[string]interface{}{
		"added":   addSuccessCount,
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

	sse.GetSSEBroker().BroadcastEvent("sub_update", sse.NotificationPayload{
		Event:   "sub_update",
		Title:   "订阅更新完成",
		Message: fmt.Sprintf("✅订阅【%s】节点同步完成，耗时 %s，总节点【%d】个，成功处理【%d】个，新增节点【%d】个，已存在节点【%d】个，删除失效【%d】个%s", subName, durationStr, len(proxys), addSuccessCount+skipCount, addSuccessCount, skipCount, deleteCount, usageText),
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
