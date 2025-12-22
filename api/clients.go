package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sublink/models"
	"sublink/node"
	"sublink/node/protocol"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

var SunName string

// md5加密
func Md5(src string) string {
	m := md5.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}
func GetClient(c *gin.Context) {
	// 获取协议头
	token := c.Query("token")
	ClientIndex := c.Query("client") // 客户端标识
	if token == "" {
		utils.Warn("token为空")
		c.Writer.WriteString("token为空")
		return
	}

	// 从分享表查找 token
	share, err := models.GetSubscriptionShareByToken(strings.ToLower(token))
	if err != nil {
		utils.Warn("无效的分享token: %s", token)
		c.Writer.WriteString("无效的分享链接")
		return
	}

	// 检查是否过期
	if share.IsExpired() {
		utils.Warn("分享链接已过期: %s", token)
		c.Writer.WriteString("分享链接已过期")
		return
	}

	// 获取关联订阅
	var sub models.Subcription
	sub.ID = share.SubscriptionID
	if err := sub.Find(); err != nil {
		utils.Warn("订阅不存在: %d", share.SubscriptionID)
		c.Writer.WriteString("订阅不存在")
		return
	}
	SunName = sub.Name

	// IP 黑白名单检查
	if sub.IPBlacklist != "" && utils.IsIpInCidr(c.ClientIP(), sub.IPBlacklist) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"msg": "IP受限(IP已被加入黑名单)",
		})
		return
	}
	if sub.IPWhitelist != "" && !utils.IsIpInCidr(c.ClientIP(), sub.IPWhitelist) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"msg": "IP受限(您的IP不在允许访问列表)",
		})
		return
	}

	// 更新访问统计
	share.RecordAccess()

	// 保存 ShareID 到上下文，供IP日志记录使用
	c.Set("shareID", share.ID)

	// 判断是否带客户端参数
	switch ClientIndex {
	case "clash":
		GetClash(c)
		return
	case "surge":
		GetSurge(c)
		return
	case "v2ray":
		GetV2ray(c)
		return
	}

	// 自动识别客户端
	ClientList := []string{"clash", "surge"}
	for k, v := range c.Request.Header {
		if k == "User-Agent" {
			for _, UserAgent := range v {
				if UserAgent == "" {
					fmt.Println("User-Agent为空")
				}
				for _, client := range ClientList {
					if strings.Contains(strings.ToLower(UserAgent), strings.ToLower(client)) {
						switch client {
						case "clash":
							GetClash(c)
							return
						case "surge":
							GetSurge(c)
							return
						default:
							fmt.Println("未知客户端")
						}
					}
				}
				GetV2ray(c)
			}
		}
	}
}
func GetV2ray(c *gin.Context) {
	var sub models.Subcription
	if SunName == "" {
		c.Writer.WriteString("订阅名为空")
		return
	}
	// subname := c.Param("subname")
	// subname := SunName
	// subname = node.Base64Decode(subname)
	sub.Name = SunName
	err := sub.Find()
	if err != nil {
		c.Writer.WriteString("找不到这个订阅:" + SunName)
		return
	}
	err = sub.GetSub("v2ray")
	if err != nil {
		c.Writer.WriteString("读取错误")
		return
	}
	baselist := ""

	// 根据配置决定是否实时刷新用量信息
	if sub.RefreshUsageOnRequest {
		node.RefreshUsageForSubscriptionNodes(sub.Nodes)
	}
	c.Writer.Header().Set("subscription-userinfo", getSubscriptionUsage(sub.Nodes))
	// 如果是HEAD请求将不进行订阅内容相关输出
	if c.Request.Method == "HEAD" {
		return
	}

	for idx, v := range sub.Nodes {
		// 应用预处理规则到 LinkName
		processedLinkName := utils.PreprocessNodeName(sub.NodeNamePreprocess, v.LinkName)
		// 应用重命名规则
		nodeLink := v.Link
		if sub.NodeNameRule != "" {
			newName := utils.RenameNode(sub.NodeNameRule, utils.NodeInfo{
				Name:        v.Name,
				LinkName:    processedLinkName,
				LinkCountry: v.LinkCountry,
				Speed:       v.Speed,
				DelayTime:   v.DelayTime,
				Group:       v.Group,
				Source:      v.Source,
				Index:       idx + 1,
				Protocol:    utils.GetProtocolFromLink(v.Link),
				Tags:        v.Tags,
			})
			nodeLink = utils.RenameNodeLink(v.Link, newName)
		}
		switch {
		// 如果包含多条节点
		case strings.Contains(v.Link, ","):
			links := strings.Split(v.Link, ",")
			// 对每个链接应用重命名
			if sub.NodeNameRule != "" {
				for i, link := range links {
					newName := utils.RenameNode(sub.NodeNameRule, utils.NodeInfo{
						Name:        v.Name,
						LinkName:    processedLinkName,
						LinkCountry: v.LinkCountry,
						Speed:       v.Speed,
						DelayTime:   v.DelayTime,
						Group:       v.Group,
						Source:      v.Source,
						Index:       idx + 1,
						Protocol:    utils.GetProtocolFromLink(link),
						Tags:        v.Tags,
					})
					links[i] = utils.RenameNodeLink(link, newName)
				}
			}
			baselist += strings.Join(links, "\n") + "\n"
			continue
		//如果是订阅转换（以 http:// 或 https:// 开头）
		case strings.HasPrefix(v.Link, "http://") || strings.HasPrefix(v.Link, "https://"):
			resp, err := http.Get(v.Link)
			if err != nil {
				utils.Error("Error getting link: %v", err)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			nodes := utils.Base64Decode(string(body))
			baselist += nodes + "\n"
		// 默认
		default:
			baselist += nodeLink + "\n"
		}
	}
	c.Set("subname", SunName)
	filename := fmt.Sprintf("%s.txt", SunName)
	encodedFilename := url.QueryEscape(filename)
	c.Writer.Header().Set("Content-Disposition", "inline; filename*=utf-8''"+encodedFilename)
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 执行脚本
	for _, script := range sub.ScriptsWithSort {
		res, err := utils.RunScript(script.Content, baselist, "v2ray")
		if err != nil {
			utils.Error("Script execution failed: %v", err)
			continue
		}
		baselist = res
	}
	c.Writer.WriteString(utils.Base64Encode(baselist))
}
func GetClash(c *gin.Context) {
	var sub models.Subcription
	// subname := c.Param("subname")
	// subname := node.Base64Decode(SunName)
	sub.Name = SunName
	err := sub.Find()
	if err != nil {
		c.Writer.WriteString("找不到这个订阅:" + SunName)
		return
	}
	err = sub.GetSub("clash")
	if err != nil {
		c.Writer.WriteString("读取错误")
		return
	}
	var urls []protocol.Urls

	// 根据配置决定是否实时刷新用量信息
	if sub.RefreshUsageOnRequest {
		node.RefreshUsageForSubscriptionNodes(sub.Nodes)
	}
	c.Writer.Header().Set("subscription-userinfo", getSubscriptionUsage(sub.Nodes))
	// 如果是HEAD请求将不进行订阅内容相关输出
	if c.Request.Method == "HEAD" {
		return
	}

	// 获取链式代理规则
	chainRules := models.GetEnabledChainRulesBySubscriptionID(sub.ID)

	// 构建节点ID到最终名称的映射（用于链式代理规则解析）
	nodeNameMap := make(map[int]string)
	for idx, v := range sub.Nodes {
		// 计算节点最终名称
		processedLinkName := utils.PreprocessNodeName(sub.NodeNamePreprocess, v.LinkName)
		finalName := v.LinkName // 默认使用原始名称
		if sub.NodeNameRule != "" {
			finalName = utils.RenameNode(sub.NodeNameRule, utils.NodeInfo{
				Name:        v.Name,
				LinkName:    processedLinkName,
				LinkCountry: v.LinkCountry,
				Speed:       v.Speed,
				DelayTime:   v.DelayTime,
				Group:       v.Group,
				Source:      v.Source,
				Index:       idx + 1,
				Protocol:    utils.GetProtocolFromLink(v.Link),
				Tags:        v.Tags,
			})
		}
		nodeNameMap[v.ID] = finalName
	}

	// 收集自定义代理组
	customGroups := models.CollectCustomProxyGroups(chainRules, sub.Nodes, nodeNameMap)

	// ========== 第一阶段：预先收集所有链路的中间节点 dialer-proxy 映射 ==========
	// key: 节点名称, value: 该节点应设置的 dialer-proxy
	chainNodeDialerMap := make(map[string]string)
	// 同时记录每个目标节点应使用的 FinalDialer
	targetNodeDialerMap := make(map[int]string)

	if len(chainRules) > 0 {
		for _, v := range sub.Nodes {
			// 检查该节点是否匹配任何链式规则
			chainResult := models.ApplyChainRulesToNodeV2(v, chainRules, sub.Nodes, nodeNameMap)
			if chainResult != nil && chainResult.FinalDialer != "" {
				// 记录目标节点的 dialer-proxy
				targetNodeDialerMap[v.ID] = chainResult.FinalDialer
				// 收集链路中间节点的 dialer-proxy 映射
				for _, link := range chainResult.Links {
					// 只处理非代理组类型的中间节点（代理组类型的 dialer-proxy 由组本身处理）
					if !link.IsGroup && link.DialerProxy != "" {
						// 如果同一节点在多个规则中作为中间节点，使用最先匹配的
						if _, exists := chainNodeDialerMap[link.ProxyName]; !exists {
							chainNodeDialerMap[link.ProxyName] = link.DialerProxy
						}
					}
				}
				// 收集中间节点自定义代理组内节点的 dialer-proxy 映射
				for memberName, dialerProxy := range chainResult.GroupMemberDialerMap {
					if _, exists := chainNodeDialerMap[memberName]; !exists {
						chainNodeDialerMap[memberName] = dialerProxy
					}
				}
			}
		}
		utils.Debug("[ChainProxy] 收集完成: 目标节点=%d, 中间节点=%d", len(targetNodeDialerMap), len(chainNodeDialerMap))
	}

	// ========== 第二阶段：遍历节点生成配置 ==========
	for idx, v := range sub.Nodes {
		// 应用预处理规则到 LinkName
		processedLinkName := utils.PreprocessNodeName(sub.NodeNamePreprocess, v.LinkName)
		// 应用重命名规则
		nodeLink := v.Link
		if sub.NodeNameRule != "" {
			newName := utils.RenameNode(sub.NodeNameRule, utils.NodeInfo{
				Name:        v.Name,
				LinkName:    processedLinkName,
				LinkCountry: v.LinkCountry,
				Speed:       v.Speed,
				DelayTime:   v.DelayTime,
				Group:       v.Group,
				Source:      v.Source,
				Index:       idx + 1,
				Protocol:    utils.GetProtocolFromLink(v.Link),
				Tags:        v.Tags,
			})
			nodeLink = utils.RenameNodeLink(v.Link, newName)
		}

		// 计算 dialer-proxy（链式代理规则）
		dialerProxy := strings.TrimSpace(v.DialerProxyName)

		// 优先级：中间节点映射 > 目标节点映射 > 节点自身设置
		finalNodeName := nodeNameMap[v.ID]

		// 检查是否作为链路中间节点（最高优先级）
		if chainDialer, exists := chainNodeDialerMap[finalNodeName]; exists {
			dialerProxy = chainDialer
		} else if targetDialer, exists := targetNodeDialerMap[v.ID]; exists && dialerProxy == "" {
			// 作为目标节点
			dialerProxy = targetDialer
		}

		switch {
		// 如果包含多条节点
		case strings.Contains(v.Link, ","):
			links := strings.Split(v.Link, ",")
			for i, link := range links {
				renamedLink := link
				if sub.NodeNameRule != "" {
					newName := utils.RenameNode(sub.NodeNameRule, utils.NodeInfo{
						Name:        v.Name,
						LinkName:    processedLinkName,
						LinkCountry: v.LinkCountry,
						Speed:       v.Speed,
						DelayTime:   v.DelayTime,
						Group:       v.Group,
						Source:      v.Source,
						Index:       idx + 1,
						Protocol:    utils.GetProtocolFromLink(link),
						Tags:        v.Tags,
					})
					renamedLink = utils.RenameNodeLink(link, newName)
				}
				links[i] = renamedLink
				urls = append(urls, protocol.Urls{
					Url:             renamedLink,
					DialerProxyName: dialerProxy,
				})
			}
			continue
		//如果是订阅转换（以 http:// 或 https:// 开头）
		case strings.HasPrefix(v.Link, "http://") || strings.HasPrefix(v.Link, "https://"):
			resp, err := http.Get(v.Link)
			if err != nil {
				utils.Error("获取包含链接失败: %v", err)
				continue
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			nodes := utils.Base64Decode(string(body))
			links := strings.Split(nodes, "\n")
			for _, link := range links {
				urls = append(urls, protocol.Urls{
					Url:             link,
					DialerProxyName: dialerProxy,
				})
			}
		// 默认
		default:
			urls = append(urls, protocol.Urls{
				Url:             nodeLink,
				DialerProxyName: dialerProxy,
			})
		}
	}

	var configs protocol.OutputConfig
	err = json.Unmarshal([]byte(sub.Config), &configs)
	if err != nil {
		c.Writer.WriteString("配置读取错误")
		return
	}

	// 如果启用 Host 替换，填充 HostMap
	if configs.ReplaceServerWithHost {
		configs.HostMap = models.GetHostMap()
	}

	// 添加自定义代理组到配置
	if len(customGroups) > 0 {
		configs.CustomProxyGroups = make([]protocol.CustomProxyGroup, 0, len(customGroups))
		for _, g := range customGroups {
			cpg := protocol.CustomProxyGroup{
				Name:    g.Name,
				Type:    g.Type,
				Proxies: g.Proxies,
			}
			if g.URLTestConfig != nil {
				cpg.URL = g.URLTestConfig.URL
				cpg.Interval = g.URLTestConfig.Interval
				cpg.Tolerance = g.URLTestConfig.Tolerance
			}
			configs.CustomProxyGroups = append(configs.CustomProxyGroups, cpg)
		}
	}

	DecodeClash, err := protocol.EncodeClash(urls, configs)
	if err != nil {
		c.Writer.WriteString(err.Error())
		return
	}
	c.Set("subname", SunName)
	filename := fmt.Sprintf("%s.yaml", SunName)
	encodedFilename := url.QueryEscape(filename)
	c.Writer.Header().Set("Content-Disposition", "inline; filename*=utf-8''"+encodedFilename)
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// 执行脚本
	for _, script := range sub.ScriptsWithSort {
		res, err := utils.RunScript(script.Content, string(DecodeClash), "clash")
		if err != nil {
			utils.Error("Script execution failed: %v", err)
			continue
		}
		DecodeClash = []byte(res)
	}
	c.Writer.WriteString(string(DecodeClash))
}

func GetSurge(c *gin.Context) {
	var sub models.Subcription
	// subname := c.Param("subname")
	// subname := node.Base64Decode(SunName)
	sub.Name = SunName
	err := sub.Find()
	if err != nil {
		c.Writer.WriteString("找不到这个订阅:" + SunName)
		return
	}
	err = sub.GetSub("surge")
	if err != nil {
		c.Writer.WriteString("读取错误")
		return
	}
	urls := []string{}

	// 根据配置决定是否实时刷新用量信息
	if sub.RefreshUsageOnRequest {
		node.RefreshUsageForSubscriptionNodes(sub.Nodes)
	}
	c.Writer.Header().Set("subscription-userinfo", getSubscriptionUsage(sub.Nodes))
	// 如果是HEAD请求将不进行订阅内容相关输出
	if c.Request.Method == "HEAD" {
		return
	}
	for idx, v := range sub.Nodes {
		// 应用预处理规则到 LinkName
		processedLinkName := utils.PreprocessNodeName(sub.NodeNamePreprocess, v.LinkName)
		// 应用重命名规则
		nodeLink := v.Link
		if sub.NodeNameRule != "" {
			newName := utils.RenameNode(sub.NodeNameRule, utils.NodeInfo{
				Name:        v.Name,
				LinkName:    processedLinkName,
				LinkCountry: v.LinkCountry,
				Speed:       v.Speed,
				DelayTime:   v.DelayTime,
				Group:       v.Group,
				Source:      v.Source,
				Index:       idx + 1,
				Protocol:    utils.GetProtocolFromLink(v.Link),
				Tags:        v.Tags,
			})
			nodeLink = utils.RenameNodeLink(v.Link, newName)
		}
		switch {
		// 如果包含多条节点
		case strings.Contains(v.Link, ","):
			links := strings.Split(v.Link, ",")
			for i, link := range links {
				if sub.NodeNameRule != "" {
					newName := utils.RenameNode(sub.NodeNameRule, utils.NodeInfo{
						Name:        v.Name,
						LinkName:    processedLinkName,
						LinkCountry: v.LinkCountry,
						Speed:       v.Speed,
						DelayTime:   v.DelayTime,
						Group:       v.Group,
						Source:      v.Source,
						Index:       idx + 1,
						Protocol:    utils.GetProtocolFromLink(link),
						Tags:        v.Tags,
					})
					links[i] = utils.RenameNodeLink(link, newName)
				}
			}
			urls = append(urls, links...)
			continue
		//如果是订阅转换（以 http:// 或 https:// 开头）
		case strings.HasPrefix(v.Link, "http://") || strings.HasPrefix(v.Link, "https://"):
			resp, err := http.Get(v.Link)
			if err != nil {
				utils.Error("Error getting link: %v", err)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			nodes := utils.Base64Decode(string(body))
			links := strings.Split(nodes, "\n")
			urls = append(urls, links...)
		// 默认
		default:
			urls = append(urls, nodeLink)
		}
	}

	var configs protocol.OutputConfig
	err = json.Unmarshal([]byte(sub.Config), &configs)
	if err != nil {
		c.Writer.WriteString("配置读取错误")
		return
	}

	// 如果启用 Host 替换，填充 HostMap
	if configs.ReplaceServerWithHost {
		configs.HostMap = models.GetHostMap()
	}

	// log.Println("surge路径:", configs)
	DecodeClash, err := protocol.EncodeSurge(urls, configs)
	if err != nil {
		c.Writer.WriteString(err.Error())
		return
	}
	c.Set("subname", SunName)
	filename := fmt.Sprintf("%s.conf", SunName)
	encodedFilename := url.QueryEscape(filename)
	c.Writer.Header().Set("Content-Disposition", "inline; filename*=utf-8''"+encodedFilename)
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	host := c.Request.Host
	url := c.Request.URL.String()
	// 如果包含头部更新信息
	if strings.Contains(DecodeClash, "#!MANAGED-CONFIG") {
		c.Writer.WriteString(DecodeClash)
		return
	}
	// 否则就插入头部更新信息
	interval := fmt.Sprintf("#!MANAGED-CONFIG %s interval=86400 strict=false", host+url)
	// 执行脚本
	for _, script := range sub.ScriptsWithSort {
		res, err := utils.RunScript(script.Content, DecodeClash, "surge")
		if err != nil {
			utils.Error("Script execution failed: %v", err)
			continue
		}
		DecodeClash = res
	}
	c.Writer.WriteString(string(interval + "\n" + DecodeClash))
}

// getSubscriptionUsage 计算订阅的流量使用情况
func getSubscriptionUsage(nodes []models.Node) string {
	airportIDs := make(map[int]bool)
	for _, node := range nodes {
		if node.Source != "manual" && node.SourceID > 0 {
			airportIDs[node.SourceID] = true
		}
	}

	var upload, download, total int64
	var expire int64 = 0
	now := time.Now().Unix()

	utils.Debug("找到机场订阅数量: %d", len(airportIDs))

	for id := range airportIDs {
		airport, err := models.GetAirportByID(id)
		if err != nil {
			utils.Warn("获取机场信息失败 %d: %v", id, err)
			continue
		}
		if airport == nil {
			utils.Warn("机场 %d 数据为空", id)
			continue
		}
		if !airport.FetchUsageInfo {
			utils.Debug("机场 %d 未开启获取流量信息", id)
			continue
		}
		// 跳过已过期的机场
		if airport.UsageExpire > 0 && airport.UsageExpire < now {
			utils.Debug("机场 %d 已过期，跳过统计", id)
			continue
		}

		utils.Debug("机场数据 %d usage: U=%d, D=%d, T=%d, E=%d", id, airport.UsageUpload, airport.UsageDownload, airport.UsageTotal, airport.UsageExpire)

		// 累加流量（忽略负数）
		if airport.UsageUpload > 0 {
			upload += airport.UsageUpload
		}
		if airport.UsageDownload > 0 {
			download += airport.UsageDownload
		}
		if airport.UsageTotal > 0 {
			total += airport.UsageTotal
		}

		// 获取最近的过期时间
		if airport.UsageExpire > 0 {
			if expire == 0 || airport.UsageExpire < expire {
				expire = airport.UsageExpire
			}
		}
	}

	result := fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d", upload, download, total, expire)
	utils.Debug("完成机场用量信息 subscription-userinfo构造: %s", result)
	return result
}
