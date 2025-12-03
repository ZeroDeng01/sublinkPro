package node

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/services/sse"
	"sublink/utils"
	"time"

	"gopkg.in/yaml.v3"
)

type ClashConfig struct {
	Proxies []Proxy `yaml:"proxies"`
}

// LoadClashConfigFromURL 从指定 URL 加载 Clash 配置
// 支持 YAML 格式和 Base64 编码的订阅链接
// id: 订阅ID
// url: 订阅链接
// subName: 订阅名称
func LoadClashConfigFromURL(id int, url string, subName string) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("URL %s，获取Clash配置失败:  %v", url, err)
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("URL %s，读取Clash配置失败:  %v", url, err)
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
				proxy, errP := LinkToProxy(Urls{Url: line}, utils.SqlConfig{})
				if errP == nil {
					config.Proxies = append(config.Proxies, proxy)
				}
			}
		}
	}

	if len(config.Proxies) == 0 {
		log.Printf("URL %s，解析失败或未找到节点 (YAML error: %v)", url, errYaml)
		return fmt.Errorf("解析失败 or 未找到节点")
	}

	return scheduleClashToNodeLinks(id, config.Proxies, subName)
}

// scheduleClashToNodeLinks 将 Clash 代理配置转换为节点链接并保存到数据库
// id: 订阅ID
// proxys: 代理节点列表
// subName: 订阅名称
func scheduleClashToNodeLinks(id int, proxys []Proxy, subName string) error {
	successCount := 0
	err := models.DeleteAutoSubscriptionNodes(id)
	if err != nil {
		log.Printf("删除旧的订阅数据失败: %v", err)
	}
	// 获取订阅的Group信息
	subS := models.SubScheduler{}
	err = subS.GetByID(id)
	if err != nil {
		log.Printf("获取订阅连接 %s 的Group失败:  %v", subName, err)
	}
	log.Printf("📄订阅【%s】获取到订阅数量【%d】", subName, len(proxys))
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
		Node.CreateDate = time.Now().Format("2006-01-02 15:04:05")
		// 插入或更新节点，避免设置好的订阅节点丢失
		err = Node.UpsertNode()
		if err != nil {
			log.Printf("❌节点存储失败【%s】：%v", proxy.Name, err)
		} else {
			successCount++
			log.Printf("✅节点存储成功【%s", proxy.Name)
		}
	}
	log.Printf("✅订阅【%s】节点拉取完成，总节点【%d】个，成功存储【%d】个", subName, len(proxys), successCount)
	// 重新查找订阅以获取最新信息
	subS = models.SubScheduler{
		Name: subName,
	}
	err = subS.Find()
	if err != nil {
		log.Printf("获取订阅连接 %s 失败:  %v", subName, err)
		return err
	}
	subS.SuccessCount = successCount
	// 当前时间
	now := time.Now()
	subS.LastRunTime = &now
	err1 := subS.Update()
	if err1 != nil {
		return err1
	}
	sse.GetSSEBroker().BroadcastEvent("sub_update", sse.NotificationPayload{
		Event:   "sub_update",
		Title:   "订阅更新完成",
		Message: fmt.Sprintf("订阅 [%s] 更新完成 (成功: %d)", subName, successCount),
		Data: map[string]interface{}{
			"id":      id,
			"name":    subName,
			"status":  "success",
			"success": successCount,
		},
	})
	return nil

}
