package api

import (
	"net/url"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func NodeUpdadte(c *gin.Context) {
	var Node models.Node
	name := c.PostForm("name")
	oldname := c.PostForm("oldname")
	oldlink := c.PostForm("oldlink")
	link := c.PostForm("link")
	dialerProxyName := c.PostForm("dialerProxyName")
	group := c.PostForm("group")
	if name == "" || link == "" {
		utils.FailWithMsg(c, "节点名称 or 备注不能为空")
		return
	}
	// 查找旧节点
	Node.Name = oldname
	Node.Link = oldlink
	err := Node.Find()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	Node.Name = name

	//更新构造节点元数据
	u, err := url.Parse(link)
	if err != nil {
		utils.Error("解析节点链接失败: %v", err)
		return
	}
	switch {
	case u.Scheme == "ss":
		ss, err := protocol.DecodeSSURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = ss.Name
		}
		Node.LinkName = ss.Name
		Node.LinkAddress = ss.Server + ":" + utils.GetPortString(ss.Port)
		Node.LinkHost = ss.Server
		Node.LinkPort = utils.GetPortString(ss.Port)
	case u.Scheme == "ssr":
		ssr, err := protocol.DecodeSSRURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = ssr.Qurey.Remarks
		}
		Node.LinkName = ssr.Qurey.Remarks
		Node.LinkAddress = ssr.Server + ":" + utils.GetPortString(ssr.Port)
		Node.LinkHost = ssr.Server
		Node.LinkPort = utils.GetPortString(ssr.Port)
	case u.Scheme == "trojan":
		trojan, err := protocol.DecodeTrojanURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}

		if Node.Name == "" {
			Node.Name = trojan.Name
		}
		Node.LinkName = trojan.Name
		Node.LinkAddress = trojan.Hostname + ":" + utils.GetPortString(trojan.Port)
		Node.LinkHost = trojan.Hostname
		Node.LinkPort = utils.GetPortString(trojan.Port)
	case u.Scheme == "vmess":
		vmess, err := protocol.DecodeVMESSURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = vmess.Ps
		}
		Node.LinkName = vmess.Ps
		prot := utils.GetPortString(vmess.Port)
		Node.LinkAddress = vmess.Add + ":" + prot
		Node.LinkHost = vmess.Host
		Node.LinkPort = prot
	case u.Scheme == "vless":
		vless, err := protocol.DecodeVLESSURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = vless.Name
		}
		Node.LinkName = vless.Name
		Node.LinkAddress = vless.Server + ":" + utils.GetPortString(vless.Port)
		Node.LinkHost = vless.Server
		Node.LinkPort = utils.GetPortString(vless.Port)
	case u.Scheme == "hy" || u.Scheme == "hysteria":
		hy, err := protocol.DecodeHYURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = hy.Name
		}
		Node.LinkName = hy.Name
		Node.LinkAddress = hy.Host + ":" + utils.GetPortString(hy.Port)
		Node.LinkHost = hy.Host
		Node.LinkPort = utils.GetPortString(hy.Port)
	case u.Scheme == "hy2" || u.Scheme == "hysteria2":
		hy2, err := protocol.DecodeHY2URL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = hy2.Name
		}
		Node.LinkName = hy2.Name
		Node.LinkAddress = hy2.Host + ":" + utils.GetPortString(hy2.Port)
		Node.LinkHost = hy2.Host
		Node.LinkPort = utils.GetPortString(hy2.Port)
	case u.Scheme == "tuic":
		tuic, err := protocol.DecodeTuicURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = tuic.Name
		}
		Node.LinkName = tuic.Name
		Node.LinkAddress = tuic.Host + ":" + utils.GetPortString(tuic.Port)
		Node.LinkHost = tuic.Host
		Node.LinkPort = utils.GetPortString(tuic.Port)
	case u.Scheme == "socks5":
		socks5, err := protocol.DecodeSocks5URL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = socks5.Name
		}
		Node.LinkName = socks5.Name
		Node.LinkAddress = socks5.Server + ":" + utils.GetPortString(socks5.Port)
		Node.LinkHost = socks5.Server
		Node.LinkPort = utils.GetPortString(socks5.Port)
	}

	Node.Link = link
	Node.DialerProxyName = dialerProxyName
	Node.Group = group
	Node.Protocol = protocol.GetProtocolFromLink(link)
	err = Node.Update()
	if err != nil {
		utils.FailWithMsg(c, "更新失败")
		return
	}

	// 处理标签
	tags := c.PostForm("tags")
	if tags != "" {
		tagNames := strings.Split(tags, ",")
		// 过滤空字符串
		var validTagNames []string
		for _, t := range tagNames {
			t = strings.TrimSpace(t)
			if t != "" {
				validTagNames = append(validTagNames, t)
			}
		}
		_ = Node.SetTagNames(validTagNames)
	} else {
		// 如果 tags 参数为空，清除标签
		_ = Node.SetTagNames([]string{})
	}

	utils.OkWithMsg(c, "更新成功")
}

// 获取节点列表
func NodeGet(c *gin.Context) {
	var Node models.Node

	// 解析过滤参数
	filter := models.NodeFilter{
		Search:      c.Query("search"),
		Group:       c.Query("group"),
		Source:      c.Query("source"),
		Protocol:    c.Query("protocol"),
		SpeedStatus: c.Query("speedStatus"),
		DelayStatus: c.Query("delayStatus"),
		SortBy:      c.Query("sortBy"),
		SortOrder:   c.Query("sortOrder"),
	}

	// 安全解析数值参数
	if maxDelayStr := c.Query("maxDelay"); maxDelayStr != "" {
		if maxDelay, err := strconv.Atoi(maxDelayStr); err == nil && maxDelay > 0 {
			filter.MaxDelay = maxDelay
		}
	}

	if minSpeedStr := c.Query("minSpeed"); minSpeedStr != "" {
		if minSpeed, err := strconv.ParseFloat(minSpeedStr, 64); err == nil && minSpeed > 0 {
			filter.MinSpeed = minSpeed
		}
	}

	// 解析国家代码数组
	filter.Countries = c.QueryArray("countries[]")

	// 解析标签数组
	filter.Tags = c.QueryArray("tags[]")

	// 验证排序字段（白名单）
	if filter.SortBy != "" && filter.SortBy != "delay" && filter.SortBy != "speed" {
		filter.SortBy = "" // 无效排序字段，忽略
	}

	// 验证排序顺序
	if filter.SortOrder != "" && filter.SortOrder != "asc" && filter.SortOrder != "desc" {
		filter.SortOrder = "asc" // 默认升序
	}

	// 解析分页参数
	page := 0
	pageSize := 0
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// 如果提供了分页参数，返回分页响应
	if page > 0 && pageSize > 0 {
		nodes, total, err := Node.ListWithFiltersPaginated(filter, page, pageSize)
		if err != nil {
			utils.FailWithMsg(c, "node list error")
			return
		}
		totalPages := 0
		if pageSize > 0 {
			totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
		}
		utils.OkDetailed(c, "node get", gin.H{
			"items":      nodes,
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		})
		return
	}

	// 不带分页参数，返回全部（向后兼容）
	nodes, err := Node.ListWithFilters(filter)
	if err != nil {
		utils.FailWithMsg(c, "node list error")
		return
	}
	utils.OkDetailed(c, "node get", nodes)
}

// NodeGetIDs 获取符合过滤条件的所有节点ID（用于全选操作）
func NodeGetIDs(c *gin.Context) {
	var Node models.Node

	// 解析过滤参数
	filter := models.NodeFilter{
		Search:      c.Query("search"),
		Group:       c.Query("group"),
		Source:      c.Query("source"),
		Protocol:    c.Query("protocol"),
		SpeedStatus: c.Query("speedStatus"),
		DelayStatus: c.Query("delayStatus"),
		SortBy:      c.Query("sortBy"),
		SortOrder:   c.Query("sortOrder"),
	}

	// 安全解析数值参数
	if maxDelayStr := c.Query("maxDelay"); maxDelayStr != "" {
		if maxDelay, err := strconv.Atoi(maxDelayStr); err == nil && maxDelay > 0 {
			filter.MaxDelay = maxDelay
		}
	}

	if minSpeedStr := c.Query("minSpeed"); minSpeedStr != "" {
		if minSpeed, err := strconv.ParseFloat(minSpeedStr, 64); err == nil && minSpeed > 0 {
			filter.MinSpeed = minSpeed
		}
	}

	// 解析国家代码数组
	filter.Countries = c.QueryArray("countries[]")

	// 解析标签数组
	filter.Tags = c.QueryArray("tags[]")

	ids, err := Node.GetFilteredNodeIDs(filter)
	if err != nil {
		utils.FailWithMsg(c, "get node ids error")
		return
	}
	utils.OkDetailed(c, "node ids get", ids)
}

// 添加节点
func NodeAdd(c *gin.Context) {
	var Node models.Node
	link := c.PostForm("link")
	name := c.PostForm("name")
	dialerProxyName := c.PostForm("dialerProxyName")
	group := c.PostForm("group")
	if link == "" {
		utils.FailWithMsg(c, "link  不能为空")
		return
	}
	if !strings.Contains(link, "://") {
		utils.FailWithMsg(c, "link 必须包含 ://")
		return
	}
	Node.Name = name
	u, err := url.Parse(link)
	if err != nil {
		utils.Error("解析节点链接失败: %v", err)
		return
	}
	switch {
	case u.Scheme == "ss":
		ss, err := protocol.DecodeSSURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if Node.Name == "" {
			Node.Name = ss.Name
		}
		Node.LinkName = ss.Name
		Node.LinkAddress = ss.Server + ":" + utils.GetPortString(ss.Port)
		Node.LinkHost = ss.Server
		Node.LinkPort = utils.GetPortString(ss.Port)
	case u.Scheme == "ssr":
		ssr, err := protocol.DecodeSSRURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if name == "" {
			Node.Name = ssr.Qurey.Remarks
		}
		Node.LinkName = ssr.Qurey.Remarks
		Node.LinkAddress = ssr.Server + ":" + utils.GetPortString(ssr.Port)
		Node.LinkHost = ssr.Server
		Node.LinkPort = utils.GetPortString(ssr.Port)
	case u.Scheme == "trojan":
		trojan, err := protocol.DecodeTrojanURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if name == "" {
			Node.Name = trojan.Name
		}
		Node.LinkName = trojan.Name
		Node.LinkAddress = trojan.Hostname + ":" + utils.GetPortString(trojan.Port)
		Node.LinkHost = trojan.Hostname
		Node.LinkPort = utils.GetPortString(trojan.Port)
	case u.Scheme == "vmess":
		vmess, err := protocol.DecodeVMESSURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}
		if name == "" {
			Node.Name = vmess.Ps
		}
		Node.LinkName = vmess.Ps
		port := utils.GetPortString(vmess.Port)
		Node.LinkAddress = vmess.Add + ":" + port
		Node.LinkHost = vmess.Host
		Node.LinkPort = port
	case u.Scheme == "vless":
		vless, err := protocol.DecodeVLESSURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}

		if name == "" {
			Node.Name = vless.Name
		}
		Node.LinkName = vless.Name
		Node.LinkAddress = vless.Server + ":" + utils.GetPortString(vless.Port)
		Node.LinkHost = vless.Server
		Node.LinkPort = utils.GetPortString(vless.Port)
	case u.Scheme == "hy" || u.Scheme == "hysteria":
		hy, err := protocol.DecodeHYURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}

		if name == "" {
			Node.Name = hy.Name
		}
		Node.LinkName = hy.Name
		Node.LinkAddress = hy.Host + ":" + utils.GetPortString(hy.Port)
		Node.LinkHost = hy.Host
		Node.LinkPort = utils.GetPortString(hy.Port)
	case u.Scheme == "hy2" || u.Scheme == "hysteria2":
		hy2, err := protocol.DecodeHY2URL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}

		if name == "" {
			Node.Name = hy2.Name
		}
		Node.LinkName = hy2.Name
		Node.LinkAddress = hy2.Host + ":" + utils.GetPortString(hy2.Port)
		Node.LinkHost = hy2.Host
		Node.LinkPort = utils.GetPortString(hy2.Port)
	case u.Scheme == "tuic":
		tuic, err := protocol.DecodeTuicURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}

		if name == "" {
			Node.Name = tuic.Name
		}
		Node.LinkName = tuic.Name
		Node.LinkAddress = tuic.Host + ":" + utils.GetPortString(tuic.Port)
		Node.LinkHost = tuic.Host
		Node.LinkPort = utils.GetPortString(tuic.Port)
	case u.Scheme == "socks5":
		socks5, err := protocol.DecodeSocks5URL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}

		if name == "" {
			Node.Name = socks5.Name
		}
		Node.LinkName = socks5.Name
		Node.LinkAddress = socks5.Server + ":" + utils.GetPortString(socks5.Port)
		Node.LinkHost = socks5.Server
		Node.LinkPort = utils.GetPortString(socks5.Port)
	case u.Scheme == "anytls":
		anytls, err := protocol.DecodeAnyTLSURL(link)
		if err != nil {
			utils.Error("解析节点链接失败: %v", err)
			return
		}

		if name == "" {
			Node.Name = anytls.Name
		}
		Node.LinkName = anytls.Name
		Node.LinkAddress = anytls.Server + ":" + utils.GetPortString(anytls.Port)
		Node.LinkHost = anytls.Server
		Node.LinkPort = utils.GetPortString(anytls.Port)
	}
	Node.Link = link
	Node.DialerProxyName = dialerProxyName
	Node.Group = group
	Node.Protocol = protocol.GetProtocolFromLink(link)
	err = Node.Find()
	// 如果找到记录说明重复
	if err == nil {
		Node.Name = name + " " + time.Now().Format("2006-01-02 15:04:05")
	}
	err = Node.Add()
	if err != nil {
		utils.FailWithMsg(c, "添加失败检查一下是否节点重复")
		return
	}

	// 处理标签
	tags := c.PostForm("tags")
	if tags != "" {
		tagNames := strings.Split(tags, ",")
		// 过滤空字符串
		var validTagNames []string
		for _, t := range tagNames {
			t = strings.TrimSpace(t)
			if t != "" {
				validTagNames = append(validTagNames, t)
			}
		}
		_ = Node.SetTagNames(validTagNames)
	}

	utils.OkWithMsg(c, "添加成功")
}

// 删除节点
func NodeDel(c *gin.Context) {
	var Node models.Node
	id := c.Query("id")
	if id == "" {
		utils.FailWithMsg(c, "id 不能为空")
		return
	}
	x, _ := strconv.Atoi(id)
	Node.ID = x
	err := Node.Del()
	if err != nil {
		utils.FailWithMsg(c, "删除失败")
		return
	}
	utils.OkWithMsg(c, "删除成功")
}

// 节点统计
func NodesTotal(c *gin.Context) {
	var Node models.Node
	nodes, err := Node.List()
	if err != nil {
		utils.FailWithMsg(c, "获取不到节点统计")
		return
	}

	total := len(nodes)
	available := 0
	for _, n := range nodes {
		if n.Speed > 0 && n.DelayTime > 0 {
			available++
		}
	}

	utils.OkDetailed(c, "取得节点统计", gin.H{
		"total":     total,
		"available": available,
	})
}

// NodeBatchDel 批量删除节点
func NodeBatchDel(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "请选择要删除的节点")
		return
	}
	err := models.BatchDel(req.IDs)
	if err != nil {
		utils.FailWithMsg(c, "批量删除失败")
		return
	}
	utils.OkWithMsg(c, "批量删除成功")
}

// NodeBatchUpdateGroup 批量更新节点分组
func NodeBatchUpdateGroup(c *gin.Context) {
	var req struct {
		IDs   []int  `json:"ids"`
		Group string `json:"group"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "请选择要修改的节点")
		return
	}
	err := models.BatchUpdateGroup(req.IDs, req.Group)
	if err != nil {
		utils.FailWithMsg(c, "批量更新分组失败")
		return
	}
	utils.OkWithMsg(c, "批量更新分组成功")
}

// NodeBatchUpdateDialerProxy 批量更新节点前置代理
func NodeBatchUpdateDialerProxy(c *gin.Context) {
	var req struct {
		IDs             []int  `json:"ids"`
		DialerProxyName string `json:"dialerProxyName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "请选择要修改的节点")
		return
	}
	err := models.BatchUpdateDialerProxy(req.IDs, req.DialerProxyName)
	if err != nil {
		utils.FailWithMsg(c, "批量更新前置代理失败")
		return
	}
	utils.OkWithMsg(c, "批量更新前置代理成功")
}

// NodeBatchUpdateSource 批量更新节点来源
func NodeBatchUpdateSource(c *gin.Context) {
	var req struct {
		IDs    []int  `json:"ids"`
		Source string `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "请选择要修改的节点")
		return
	}
	err := models.BatchUpdateSource(req.IDs, req.Source)
	if err != nil {
		utils.FailWithMsg(c, "批量更新来源失败")
		return
	}
	utils.OkWithMsg(c, "批量更新来源成功")
}

// 获取所有分组列表
func GetGroups(c *gin.Context) {
	var node models.Node
	groups, err := node.GetAllGroups()
	if err != nil {
		utils.FailWithMsg(c, "获取分组列表失败")
		return
	}
	utils.OkDetailed(c, "获取分组列表成功", groups)
}

// GetSources 获取所有来源列表
func GetSources(c *gin.Context) {
	var node models.Node
	sources, err := node.GetAllSources()
	if err != nil {
		utils.FailWithMsg(c, "获取来源列表失败")
		return
	}
	utils.OkDetailed(c, "获取来源列表成功", sources)
}

// FastestSpeedNode 获取最快速度节点
func FastestSpeedNode(c *gin.Context) {
	node := models.GetFastestSpeedNode()
	utils.OkDetailed(c, "获取最快速度节点成功", node)
}

// LowestDelayNode 获取最低延迟节点
func LowestDelayNode(c *gin.Context) {
	node := models.GetLowestDelayNode()
	utils.OkDetailed(c, "获取最低延迟节点成功", node)
}

// GetNodeCountries 获取所有节点的国家代码列表
func GetNodeCountries(c *gin.Context) {
	countries := models.GetAllCountries()
	utils.OkDetailed(c, "获取国家代码成功", countries)
}

// NodeCountryStats 获取按国家统计的节点数量
func NodeCountryStats(c *gin.Context) {
	stats := models.GetNodeCountryStats()
	utils.OkDetailed(c, "获取国家统计成功", stats)
}

// NodeProtocolStats 获取按协议统计的节点数量
func NodeProtocolStats(c *gin.Context) {
	stats := models.GetNodeProtocolStats()
	utils.OkDetailed(c, "获取协议统计成功", stats)
}

// NodeTagStats 获取按标签统计的节点数量
func NodeTagStats(c *gin.Context) {
	stats := models.GetNodeTagStats()
	utils.OkDetailed(c, "获取标签统计成功", stats)
}

// NodeGroupStats 获取按分组统计的节点数量
func NodeGroupStats(c *gin.Context) {
	stats := models.GetNodeGroupStats()
	utils.OkDetailed(c, "获取分组统计成功", stats)
}

// NodeSourceStats 获取按来源统计的节点数量
func NodeSourceStats(c *gin.Context) {
	stats := models.GetNodeSourceStats()
	utils.OkDetailed(c, "获取来源统计成功", stats)
}

// GetIPDetails 获取IP详细信息
// GET /api/v1/nodes/ip-info?ip=xxx.xxx.xxx.xxx
func GetIPDetails(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		utils.FailWithMsg(c, "IP地址不能为空")
		return
	}

	// 调用模型层获取IP信息（多级缓存）
	ipInfo, err := models.GetIPInfo(ip)
	if err != nil {
		utils.FailWithMsg(c, "查询IP信息失败: "+err.Error())
		return
	}

	utils.OkDetailed(c, "获取成功", ipInfo)
}

// GetIPCacheStats 获取IP缓存统计
// GET /api/v1/nodes/ip-cache/stats
func GetIPCacheStats(c *gin.Context) {
	count := models.GetIPInfoCount()
	utils.OkDetailed(c, "获取成功", gin.H{
		"count": count,
	})
}

// ClearIPCache 清除所有IP缓存
// DELETE /api/v1/nodes/ip-cache
func ClearIPCache(c *gin.Context) {
	err := models.ClearAllIPInfo()
	if err != nil {
		utils.FailWithMsg(c, "清除失败: "+err.Error())
		return
	}
	utils.OkWithMsg(c, "IP缓存已清除")
}

// GetNodeProtocols 获取所有使用中的协议类型列表（用于过滤器选项）
// GET /api/v1/nodes/protocols
func GetNodeProtocols(c *gin.Context) {
	protocols := models.GetAllProtocols()
	utils.OkDetailed(c, "获取协议列表成功", protocols)
}
