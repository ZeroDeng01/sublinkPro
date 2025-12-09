package node

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

type ClashConfig struct {
	Proxies []protocol.Proxy `yaml:"proxies"`
}

// LoadClashConfigFromURL 从指定 URL 加载 Clash 配置
// 支持 YAML 格式和 Base64 编码的订阅链接
// id: 订阅ID
// url: 订阅链接
// subName: 订阅名称
// downloadWithProxy: 是否使用代理下载
// proxyLink: 代理链接 (可选)
func LoadClashConfigFromURL(id int, urlStr string, subName string, downloadWithProxy bool, proxyLink string) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if downloadWithProxy {
		var proxyNodeLink string

		if proxyLink != "" {
			// 使用指定的代理链接
			proxyNodeLink = proxyLink
			log.Printf("使用指定代理下载订阅")
		} else {
			// 如果没有指定代理，尝试自动选择最佳代理
			// 获取最近测速成功的节点（延迟最低且速度大于0）
			if bestNode, err := models.GetBestProxyNode(); err == nil && bestNode != nil {
				log.Printf("自动选择最佳代理节点: %s 节点延迟：%dms  节点速度：%2fMB/s", bestNode.Name, bestNode.DelayTime, bestNode.Speed)
				proxyNodeLink = bestNode.Link
			}
		}

		if proxyNodeLink != "" {
			// 使用 mihomo 内核创建代理适配器
			proxyAdapter, err := mihomo.GetMihomoAdapter(proxyNodeLink)
			if err != nil {
				log.Printf("创建 mihomo 代理适配器失败: %v，将直接下载", err)
			} else {
				log.Printf("使用 mihomo 内核代理下载订阅")
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
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
			}
		} else {
			log.Println("未找到可用代理，将直接下载")
		}
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		log.Printf("URL %s，获取Clash配置失败:  %v", urlStr, err)
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("URL %s，读取Clash配置失败:  %v", urlStr, err)
		return err
	}
	var config ClashConfig
	// 尝试解析 YAML
	errYaml := yaml.Unmarshal(data, &config)

	// 如果 YAML 解析失败或没有代理节点，尝试 Base64 解码
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
				proxy, errP := protocol.LinkToProxy(protocol.Urls{Url: line}, utils.SqlConfig{})
				if errP == nil {
					config.Proxies = append(config.Proxies, proxy)
				}
			}
		}
	}

	if len(config.Proxies) == 0 {
		log.Printf("URL %s，解析失败或未找到节点 (YAML error: %v)", urlStr, errYaml)
		return fmt.Errorf("解析失败 or 未找到节点")
	}

	return scheduleClashToNodeLinks(id, config.Proxies, subName)
}

// scheduleClashToNodeLinks 将 Clash 代理配置转换为节点链接并保存到数据库
// id: 订阅ID
// proxys: 代理节点列表
// subName: 订阅名称
func scheduleClashToNodeLinks(id int, proxys []protocol.Proxy, subName string) error {
	addSuccessCount := 0
	updateSuccessCount := 0
	processedCount := 0
	totalNodes := len(proxys)

	// 生成唯一任务ID
	taskID := fmt.Sprintf("sub_update_%d_%d", id, time.Now().UnixNano())

	// 广播任务开始事件
	sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
		TaskID:   taskID,
		TaskType: "sub_update",
		TaskName: subName,
		Status:   "started",
		Current:  0,
		Total:    totalNodes,
		Message:  fmt.Sprintf("开始更新订阅 [%s]，共 %d 个节点", subName, totalNodes),
	})

	// 获取订阅的Group信息
	subS := models.SubScheduler{}
	err := subS.GetByID(id)
	if err != nil {
		log.Printf("获取订阅连接 %s 的Group失败:  %v", subName, err)
	}

	// 1. 获取该订阅当前在数据库中的所有节点
	existingNodes, err := models.ListBySourceID(id)
	if err != nil {
		log.Printf("获取订阅【%s】现有节点失败: %v", subName, err)
		existingNodes = []models.Node{} // 确保后续逻辑不会panic
	}

	// 创建现有节点的映射表（以Link为键）
	existingNodeMap := make(map[string]models.Node)
	for _, node := range existingNodes {
		existingNodeMap[node.Link] = node
	}

	log.Printf("📄订阅【%s】获取到订阅数量【%d】，现有节点数量【%d】", subName, len(proxys), len(existingNodes))

	// 记录本次获取到的节点Link
	currentLinks := make(map[string]bool)

	// 2. 遍历新获取的节点，插入或更新
	for _, proxy := range proxys {
		log.Printf("💾准备存储节点【%s】", proxy.Name)
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
			port := proxy.Port
			name := proxy.Name
			encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", method, password)))
			link = fmt.Sprintf("ss://%s@%s:%d#%s", encoded, server, port, name)
		case "ssr":
			// ssr://server:port:protocol:method:obfs:base64(password)/?remarks=base64(remarks)&obfsparam=base64(obfsparam)
			server := proxy.Server
			port := proxy.Port
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
			port := proxy.Port
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
			port := proxy.Port
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
			port := proxy.Port
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
			// hysteria2://auth@server:port?sni=sni&insecure=1&obfs=obfs&obfs-password=obfs-password#name
			server := proxy.Server
			port := proxy.Port
			auth := proxy.Password
			name := proxy.Name
			query := url.Values{}
			if proxy.Sni != "" {
				query.Set("sni", proxy.Sni)
			}
			if proxy.Skip_cert_verify {
				query.Set("insecure", "1")
			}
			if proxy.Obfs != "" {
				query.Set("obfs", proxy.Obfs)
			}
			if proxy.Obfs_password != "" {
				query.Set("obfs-password", proxy.Obfs_password)
			}
			if len(proxy.Alpn) > 0 {
				query.Set("alpn", strings.Join(proxy.Alpn, ","))
			}
			link = fmt.Sprintf("hysteria2://%s@%s:%d?%s#%s", auth, server, port, query.Encode(), name)

		case "tuic":
			// tuic://uuid:password@server:port?sni=sni&congestion_control=congestion_control&alpn=alpn#name
			uuid := proxy.Uuid
			password := proxy.Password
			server := proxy.Server
			port := proxy.Port
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
			port := proxy.Port
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
			port := proxy.Port
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
		Node.LinkAddress = proxy.Server + ":" + strconv.Itoa(proxy.Port)
		Node.LinkHost = proxy.Server
		Node.LinkPort = strconv.Itoa(proxy.Port)
		Node.Source = subName
		Node.SourceID = id
		Node.Group = subS.Group

		// 记录本次获取到的节点
		currentLinks[link] = true

		// 判断节点是否已存在
		var nodeStatus string
		if existingNode, exists := existingNodeMap[link]; exists {
			updateSuccessCount++
			nodeStatus = "skipped"
			log.Printf("⚠️节点【%s】已存在，不进行任何处理", existingNode.Name)
		} else {
			// 节点不存在，插入新节点
			err = Node.Add()
			if err != nil {
				nodeStatus = "failed"
				log.Printf("❌节点插入失败【%s】：%v", proxy.Name, err)
			} else {
				addSuccessCount++
				nodeStatus = "added"
				log.Printf("✅节点插入成功【%s】", proxy.Name)
			}
		}

		// 更新进度并广播 (节流：超过50个节点时每10个节点广播一次)
		processedCount++
		shouldBroadcast := totalNodes <= 50 || processedCount%10 == 0 || processedCount == totalNodes
		if shouldBroadcast {
			sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
				TaskID:      taskID,
				TaskType:    "sub_update",
				TaskName:    subName,
				Status:      "progress",
				Current:     processedCount,
				Total:       totalNodes,
				CurrentItem: proxy.Name,
				Result: map[string]interface{}{
					"status": nodeStatus,
					"added":  addSuccessCount,
					"exists": updateSuccessCount,
				},
				Message: fmt.Sprintf("处理节点 %d/%d", processedCount, totalNodes),
			})
		}
	}

	// 3. 删除本次订阅没有获取到但数据库中存在的节点
	deleteCount := 0
	for link, existingNode := range existingNodeMap {
		if !currentLinks[link] {
			// 该节点不在本次订阅中，需要删除
			nodeToDelete := models.Node{ID: existingNode.ID}
			if err := nodeToDelete.Del(); err != nil {
				log.Printf("❌删除失效节点失败【%s】：%v", existingNode.Name, err)
			} else {
				deleteCount++
				log.Printf("🗑️删除失效节点【%s】", existingNode.Name)
			}
		}
	}

	log.Printf("✅订阅【%s】节点同步完成，总节点【%d】个，成功处理【%d】个，新增节点【%d】个，已存在节点【%d】个，删除失效【%d】个", subName, len(proxys), addSuccessCount+updateSuccessCount, addSuccessCount, updateSuccessCount, deleteCount)
	// 重新查找订阅以获取最新信息
	subS = models.SubScheduler{
		Name: subName,
	}
	err = subS.Find()
	if err != nil {
		log.Printf("获取订阅连接 %s 失败:  %v", subName, err)
		return err
	}
	subS.SuccessCount = addSuccessCount + updateSuccessCount
	// 当前时间
	now := time.Now()
	subS.LastRunTime = &now
	err1 := subS.Update()
	if err1 != nil {
		return err1
	}
	// 广播完成进度
	sse.GetSSEBroker().BroadcastProgress(sse.ProgressPayload{
		TaskID:   taskID,
		TaskType: "sub_update",
		TaskName: subName,
		Status:   "completed",
		Current:  totalNodes,
		Total:    totalNodes,
		Message:  fmt.Sprintf("订阅更新完成 (新增: %d, 已存在: %d, 删除: %d)", addSuccessCount, updateSuccessCount, deleteCount),
		Result: map[string]interface{}{
			"added":   addSuccessCount,
			"exists":  updateSuccessCount,
			"deleted": deleteCount,
		},
	})

	// 触发webhook的完成事件
	sse.GetSSEBroker().BroadcastEvent("sub_update", sse.NotificationPayload{
		Event:   "sub_update",
		Title:   "订阅更新完成",
		Message: fmt.Sprintf("✅订阅【%s】节点同步完成，总节点【%d】个，成功处理【%d】个，新增节点【%d】个，已存在节点【%d】个，删除失效【%d】个", subName, len(proxys), addSuccessCount+updateSuccessCount, addSuccessCount, updateSuccessCount, deleteCount),
		Data: map[string]interface{}{
			"id":      id,
			"name":    subName,
			"status":  "success",
			"success": addSuccessCount + updateSuccessCount,
		},
	})
	return nil

}
