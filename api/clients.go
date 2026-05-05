package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sublink/models"
	"sublink/node"
	"sublink/node/protocol"
	"sublink/utils"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const subscriptionNameContextKey = "resolvedSubscriptionName"

type clientResponseMode int

const (
	clientResponseNormal clientResponseMode = iota
	clientResponseSyntheticFallback
)

type preparedClientResponse struct {
	ClientType   string
	Mode         clientResponseMode
	Subscription models.Subcription
	SubName      string
	ShareID      int
	FallbackName string
}

type resolvedPreparedResponse struct {
	Subscription models.Subcription
	SubName      string
}

const syntheticClashTemplate = `port: 7890
proxies: []
proxy-groups:
  - name: 节点选择
    type: select
    proxies: []
`

const syntheticSurgeTemplate = `[General]

[Proxy]

[Proxy Group]
节点选择 = select
`

var testGetClientAfterResolveSubscriptionNameHook func(*gin.Context)

var (
	syntheticTemplateOnce sync.Once
	syntheticClashPath    string
	syntheticSurgePath    string
	syntheticTemplateErr  error
)

func setResolvedSubscriptionName(c *gin.Context, subName string) {
	c.Set(subscriptionNameContextKey, subName)
}

func getResolvedSubscriptionName(c *gin.Context) (string, bool) {
	value, ok := c.Get(subscriptionNameContextKey)
	if !ok {
		return "", false
	}

	subName, ok := value.(string)
	if !ok || subName == "" {
		return "", false
	}

	return subName, true
}

func resolvedSubscriptionNameOrWriteError(c *gin.Context) (string, bool) {
	subName, ok := getResolvedSubscriptionName(c)
	if !ok {
		c.Writer.WriteString("订阅名为空")
		return "", false
	}

	return subName, true
}

func GetClient(c *gin.Context) {
	// 获取协议头
	token := c.Query("token")
	if token == "" {
		utils.Warn("token为空")
		c.Writer.WriteString("Not Found")
		return
	}
	clientType := resolveSubscriptionClient(c)
	prepared, ok := prepareClientResponse(c, clientType, strings.ToLower(token))
	if !ok {
		return
	}
	setResolvedSubscriptionName(c, prepared.SubName)
	if testGetClientAfterResolveSubscriptionNameHook != nil {
		testGetClientAfterResolveSubscriptionNameHook(c)
	}
	c.Set("shareID", prepared.ShareID)
	dispatchPreparedClientResponse(c, prepared)
}

func prepareClientResponse(c *gin.Context, clientType, token string) (preparedClientResponse, bool) {
	share, err := models.GetSubscriptionShareByToken(token)
	if err != nil {
		utils.Warn("无效的分享token: %s", token)
		return buildSyntheticFallbackResponse(clientType, "无效的分享链接"), true
	}

	if share.IsExpired() {
		utils.Warn("分享链接已过期: %s", token)
		var expiredSub models.Subcription
		expiredSub.ID = share.SubscriptionID
		if err := expiredSub.Find(); err != nil {
			utils.Warn("过期分享关联订阅不存在: %d", share.SubscriptionID)
			return buildSyntheticFallbackResponse(clientType, "订阅不存在"), true
		}
		return buildPreparedExpiredShareResponse(expiredSub, clientType, "订阅已过期", share.ID)
	}

	var sub models.Subcription
	sub.ID = share.SubscriptionID
	if err := sub.Find(); err != nil {
		utils.Warn("订阅不存在: %d", share.SubscriptionID)
		return buildSyntheticFallbackResponse(clientType, "订阅不存在"), true
	}

	// IP 黑白名单检查
	if sub.IPBlacklist != "" && utils.IsIpInCidr(c.ClientIP(), sub.IPBlacklist) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"msg": "IP受限(IP已被加入黑名单)",
		})
		return preparedClientResponse{}, false
	}
	if sub.IPWhitelist != "" && !utils.IsIpInCidr(c.ClientIP(), sub.IPWhitelist) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"msg": "IP受限(您的IP不在允许访问列表)",
		})
		return preparedClientResponse{}, false
	}

	// 更新访问统计
	share.RecordAccess()
	prepared, ok := buildPreparedResponseFromSubscription(sub, clientType, share.ID)
	if !ok {
		return preparedClientResponse{}, false
	}
	return prepared, true
}

func resolveSubscriptionClient(c *gin.Context) string {
	clientIndex := c.Query("client")
	switch clientIndex {
	case "clash", "surge", "v2ray":
		return clientIndex
	}

	for k, v := range c.Request.Header {
		if k != "User-Agent" {
			continue
		}
		for _, userAgent := range v {
			if userAgent == "" {
				fmt.Println("User-Agent为空")
			}
			for _, client := range []string{"clash", "surge"} {
				if strings.Contains(strings.ToLower(userAgent), strings.ToLower(client)) {
					return client
				}
			}
			return "v2ray"
		}
	}

	return "v2ray"
}

func dispatchPreparedClientResponse(c *gin.Context, prepared preparedClientResponse) {
	switch prepared.ClientType {
	case "clash":
		renderPreparedClash(c, prepared)
	case "surge":
		renderPreparedSurge(c, prepared)
	default:
		renderPreparedV2ray(c, prepared)
	}
}

func buildSyntheticFallbackResponse(clientType, message string) preparedClientResponse {
	config, err := buildSyntheticFallbackConfig()
	if err != nil {
		utils.Warn("构造 synthetic fallback 配置失败: %v", err)
	}
	sub := models.Subcription{
		Name:                  message,
		Config:                config,
		Nodes:                 buildSyntheticErrorNodes(message),
		RefreshUsageOnRequest: false,
	}
	return preparedClientResponse{
		ClientType:   clientType,
		Mode:         clientResponseSyntheticFallback,
		Subscription: sub,
		SubName:      sub.Name,
		FallbackName: message,
	}
}

func buildSyntheticErrorNodes(message string) []models.Node {
	httpLink := protocol.EncodeHTTPURL(protocol.HTTP{
		Name:   message,
		Server: "placeholder.invalid",
		Port:   80,
		TLS:    false,
	})

	return []models.Node{{
		ID:       -1,
		Name:     message,
		LinkName: message,
		Link:     httpLink,
		Protocol: "http",
		Source:   "manual",
		SourceID: 0,
	}}
}

func buildSyntheticFallbackConfig() (string, error) {
	clashPath, surgePath, err := getSyntheticTemplatePaths()
	if err != nil {
		return "", err
	}

	config := map[string]string{
		"clash": clashPath,
		"surge": surgePath,
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}

func getSyntheticTemplatePaths() (string, string, error) {
	syntheticTemplateOnce.Do(func() {
		syntheticClashPath, syntheticTemplateErr = writeSyntheticTemplateFile("synthetic-clash-*.yaml", syntheticClashTemplate)
		if syntheticTemplateErr != nil {
			return
		}
		syntheticSurgePath, syntheticTemplateErr = writeSyntheticTemplateFile("synthetic-surge-*.conf", syntheticSurgeTemplate)
	})

	return syntheticClashPath, syntheticSurgePath, syntheticTemplateErr
}

func writeSyntheticTemplateFile(pattern, content string) (string, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return "", err
	}

	return file.Name(), nil
}

func buildPreparedResponseFromSubscription(sub models.Subcription, clientType string, shareID int) (preparedClientResponse, bool) {
	preparedSub := sub
	if err := preparedSub.GetSub(clientType); err != nil {
		return preparedClientResponse{}, false
	}
	return preparedClientResponse{
		ClientType:   clientType,
		Mode:         clientResponseNormal,
		Subscription: preparedSub,
		SubName:      preparedSub.Name,
		ShareID:      shareID,
	}, true
}

func buildPreparedExpiredShareResponse(sub models.Subcription, clientType, message string, shareID int) (preparedClientResponse, bool) {
	prepared, ok := buildPreparedResponseFromSubscription(sub, clientType, shareID)
	if !ok {
		return preparedClientResponse{}, false
	}
	prepared.Mode = clientResponseSyntheticFallback
	prepared.FallbackName = message
	return prepared, true
}

func applyPreparedResponseMode(prepared preparedClientResponse) resolvedPreparedResponse {
	sub := prepared.Subscription
	subName := prepared.SubName

	switch prepared.Mode {
	case clientResponseSyntheticFallback:
		sub.Nodes = buildSyntheticErrorNodes(prepared.FallbackName)
		sub.RefreshUsageOnRequest = false
		if prepared.ShareID == 0 {
			sub.Name = prepared.FallbackName
			subName = prepared.FallbackName
		}
	}

	return resolvedPreparedResponse{
		Subscription: sub,
		SubName:      subName,
	}
}

func prepareRendererResponse(c *gin.Context, prepared preparedClientResponse) (resolvedPreparedResponse, bool) {
	resolved := applyPreparedResponseMode(prepared)
	sub := resolved.Subscription
	if sub.RefreshUsageOnRequest {
		node.RefreshUsageForSubscriptionNodes(sub.Nodes)
	}
	c.Writer.Header().Set("subscription-userinfo", getSubscriptionUsage(sub.Nodes))
	c.Set("subname", resolved.SubName)
	if c.Request.Method == "HEAD" {
		return resolved, false
	}
	return resolved, true
}

func GetV2ray(c *gin.Context) {
	subName, ok := resolvedSubscriptionNameOrWriteError(c)
	if !ok {
		return
	}
	var sub models.Subcription
	sub.Name = subName
	if err := sub.Find(); err != nil {
		c.Writer.WriteString("找不到这个订阅:" + subName)
		return
	}
	prepared, ok := buildPreparedResponseFromSubscription(sub, "v2ray", 0)
	if !ok {
		c.Writer.WriteString("读取错误")
		return
	}
	renderPreparedV2ray(c, prepared)
}

func renderPreparedV2ray(c *gin.Context, prepared preparedClientResponse) {
	resolved, shouldWriteBody := prepareRendererResponse(c, prepared)
	if !shouldWriteBody {
		return
	}
	sub := resolved.Subscription
	subName := resolved.SubName
	baselist := ""

	for idx, v := range sub.Nodes {
		// 应用预处理规则到 LinkName
		processedLinkName := utils.PreprocessNodeName(sub.NodeNamePreprocess, v.LinkName)
		// 应用重命名规则
		nodeLink := v.Link
		if sub.NodeNameRule != "" {
			newName := utils.RenameNode(sub.NodeNameRule, models.BuildNodeRenameInfo(v, processedLinkName, protocol.GetProtocolFromLink(v.Link), idx+1))
			nodeLink = utils.RenameNodeLink(v.Link, newName)
		}
		switch {
		// 如果包含多条节点
		case strings.Contains(v.Link, ","):
			links := strings.Split(v.Link, ",")
			// 对每个链接应用重命名
			if sub.NodeNameRule != "" {
				for i, link := range links {
					newName := utils.RenameNode(sub.NodeNameRule, models.BuildNodeRenameInfo(v, processedLinkName, protocol.GetProtocolFromLink(link), idx+1))
					links[i] = utils.RenameNodeLink(link, newName)
				}
			}
			baselist += strings.Join(links, "\n") + "\n"
			continue
		//如果是订阅转换（以 http:// 或 https:// 开头，但不是HTTP/HTTPS代理节点）
		case (strings.HasPrefix(v.Link, "http://") || strings.HasPrefix(v.Link, "https://")) && !protocol.IsHTTPLink(v.Link):
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
	filename := fmt.Sprintf("%s.txt", subName)
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
	subName, ok := resolvedSubscriptionNameOrWriteError(c)
	if !ok {
		return
	}
	var sub models.Subcription
	sub.Name = subName
	if err := sub.Find(); err != nil {
		c.Writer.WriteString("找不到这个订阅:" + subName)
		return
	}
	prepared, ok := buildPreparedResponseFromSubscription(sub, "clash", 0)
	if !ok {
		c.Writer.WriteString("读取错误")
		return
	}
	renderPreparedClash(c, prepared)
}

func renderPreparedClash(c *gin.Context, prepared preparedClientResponse) {
	resolved, shouldWriteBody := prepareRendererResponse(c, prepared)
	if !shouldWriteBody {
		return
	}
	sub := resolved.Subscription
	subName := resolved.SubName
	var urls []protocol.Urls

	// 获取链式代理规则
	chainRules := models.GetEnabledChainRulesBySubscriptionID(sub.ID)

	// 构建节点ID到最终名称的映射（用于链式代理规则解析）
	nodeNameMap := make(map[int]string)
	for idx, v := range sub.Nodes {
		// 计算节点最终名称
		processedLinkName := utils.PreprocessNodeName(sub.NodeNamePreprocess, v.LinkName)
		finalName := v.LinkName // 默认使用原始名称
		if sub.NodeNameRule != "" {
			finalName = utils.RenameNode(sub.NodeNameRule, models.BuildNodeRenameInfo(v, processedLinkName, protocol.GetProtocolFromLink(v.Link), idx+1))
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
			newName := utils.RenameNode(sub.NodeNameRule, models.BuildNodeRenameInfo(v, processedLinkName, protocol.GetProtocolFromLink(v.Link), idx+1))
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
					newName := utils.RenameNode(sub.NodeNameRule, models.BuildNodeRenameInfo(v, processedLinkName, protocol.GetProtocolFromLink(link), idx+1))
					renamedLink = utils.RenameNodeLink(link, newName)
				}
				links[i] = renamedLink
				urls = append(urls, protocol.Urls{
					Url:             renamedLink,
					DialerProxyName: dialerProxy,
				})
			}
			continue
		//如果是订阅转换（以 http:// 或 https:// 开头，但不是HTTP/HTTPS代理节点）
		case (strings.HasPrefix(v.Link, "http://") || strings.HasPrefix(v.Link, "https://")) && !protocol.IsHTTPLink(v.Link):
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
	err := json.Unmarshal([]byte(sub.Config), &configs)
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
				cpg.Strategy = g.URLTestConfig.Strategy
			}
			configs.CustomProxyGroups = append(configs.CustomProxyGroups, cpg)
		}
	}

	DecodeClash, err := protocol.EncodeClash(urls, configs)
	if err != nil {
		c.Writer.WriteString(err.Error())
		return
	}
	filename := fmt.Sprintf("%s.yaml", subName)
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
	subName, ok := resolvedSubscriptionNameOrWriteError(c)
	if !ok {
		return
	}
	var sub models.Subcription
	sub.Name = subName
	if err := sub.Find(); err != nil {
		c.Writer.WriteString("找不到这个订阅:" + subName)
		return
	}
	prepared, ok := buildPreparedResponseFromSubscription(sub, "surge", 0)
	if !ok {
		c.Writer.WriteString("读取错误")
		return
	}
	renderPreparedSurge(c, prepared)
}

func renderPreparedSurge(c *gin.Context, prepared preparedClientResponse) {
	resolved, shouldWriteBody := prepareRendererResponse(c, prepared)
	if !shouldWriteBody {
		return
	}
	sub := resolved.Subscription
	subName := resolved.SubName
	urls := []string{}
	for idx, v := range sub.Nodes {
		// 应用预处理规则到 LinkName
		processedLinkName := utils.PreprocessNodeName(sub.NodeNamePreprocess, v.LinkName)
		// 应用重命名规则
		nodeLink := v.Link
		if sub.NodeNameRule != "" {
			newName := utils.RenameNode(sub.NodeNameRule, utils.NodeInfo{
				Name:          v.Name,
				LinkName:      processedLinkName,
				LinkCountry:   v.LinkCountry,
				Speed:         v.Speed,
				SpeedStatus:   v.SpeedStatus,
				DelayTime:     v.DelayTime,
				DelayStatus:   v.DelayStatus,
				Group:         v.Group,
				Source:        v.Source,
				Index:         idx + 1,
				Protocol:      protocol.GetProtocolFromLink(v.Link),
				Tags:          v.Tags,
				IsBroadcast:   v.IsBroadcast,
				IsResidential: v.IsResidential,
				FraudScore:    v.FraudScore,
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
						Name:          v.Name,
						LinkName:      processedLinkName,
						LinkCountry:   v.LinkCountry,
						Speed:         v.Speed,
						SpeedStatus:   v.SpeedStatus,
						DelayTime:     v.DelayTime,
						DelayStatus:   v.DelayStatus,
						Group:         v.Group,
						Source:        v.Source,
						Index:         idx + 1,
						Protocol:      protocol.GetProtocolFromLink(link),
						Tags:          v.Tags,
						IsBroadcast:   v.IsBroadcast,
						IsResidential: v.IsResidential,
						FraudScore:    v.FraudScore,
					})
					links[i] = utils.RenameNodeLink(link, newName)
				}
			}
			urls = append(urls, links...)
			continue
		//如果是订阅转换（以 http:// 或 https:// 开头，但不是HTTP/HTTPS代理节点）
		case (strings.HasPrefix(v.Link, "http://") || strings.HasPrefix(v.Link, "https://")) && !protocol.IsHTTPLink(v.Link):
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
	err := json.Unmarshal([]byte(sub.Config), &configs)
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
	filename := fmt.Sprintf("%s.conf", subName)
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
	var domain string
	if c.Request.TLS != nil {
		domain = "https://" + host
	} else {
		domain = "http://" + host
	}
	proto := c.Request.Header.Get("X-Forwarded-Proto")
	if proto != "" {
		domain = proto + "://" + host
	}

	systemDomain, _ := models.GetSetting("system_domain")
	if systemDomain != "" {
		domain = systemDomain
	}
	// 否则就插入头部更新信息
	interval := fmt.Sprintf("#!MANAGED-CONFIG %s interval=86400 strict=false", domain+url)
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
