package api

import (
	"errors"
	"strconv"
	"strings"
	"sublink/database"
	"sublink/models"
	"sublink/node"
	"sublink/node/protocol"
	"sublink/services/unlock"
	"sublink/utils"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func normalizeResidentialType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "residential", "datacenter", "untested":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeQualityStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case models.QualityStatusUntested, models.QualityStatusSuccess, models.QualityStatusPartial, models.QualityStatusFailed, models.QualityStatusDisabled:
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeIPType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "native", "broadcast", "untested":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeUnlockStatus(value string) string {
	return unlock.NormalizeUnlockStatus(value)
}

func failNodeNameConflict(c *gin.Context, err error) bool {
	if errors.Is(err, models.ErrNodeNameExists) {
		utils.FailWithMsg(c, "This node remark already exists; choose another one")
		return true
	}
	return false
}

func parseUnlockRulesFromQuery(c *gin.Context) []models.UnlockFilterRule {
	rawRules := strings.TrimSpace(c.Query("unlockRules"))
	rules := models.ParseUnlockFilterRules(rawRules)
	if len(rules) > 0 {
		return rules
	}

	providers := c.QueryArray("unlockRules[][provider]")
	statuses := c.QueryArray("unlockRules[][status]")
	keywords := c.QueryArray("unlockRules[][keyword]")
	maxLen := len(providers)
	if len(statuses) > maxLen {
		maxLen = len(statuses)
	}
	if len(keywords) > maxLen {
		maxLen = len(keywords)
	}
	if maxLen > 0 {
		ruleList := make([]models.UnlockFilterRule, 0, maxLen)
		for i := 0; i < maxLen; i++ {
			rule := models.UnlockFilterRule{}
			if i < len(providers) {
				rule.Provider = models.NormalizeUnlockProvider(providers[i])
			}
			if i < len(statuses) {
				rule.Status = normalizeUnlockStatus(statuses[i])
			}
			if i < len(keywords) {
				rule.Keyword = strings.TrimSpace(keywords[i])
			}
			ruleList = append(ruleList, rule)
		}
		return models.NormalizeUnlockFilterRules(ruleList)
	}

	legacyProvider := models.NormalizeUnlockProvider(c.Query("unlockProvider"))
	legacyStatus := normalizeUnlockStatus(c.Query("unlockStatus"))
	legacyKeyword := strings.TrimSpace(c.Query("unlockKeyword"))
	if legacyProvider != "" || legacyStatus != "" || legacyKeyword != "" {
		return []models.UnlockFilterRule{{Provider: legacyProvider, Status: legacyStatus, Keyword: legacyKeyword}}
	}
	return nil
}

func parseExcludeIDs(c *gin.Context) []int {
	return parseIntList(c, []string{"excludeIds[]", "excludeIds"})
}

func parseSelectedIDs(c *gin.Context) []int {
	return parseIntList(c, []string{"ids[]", "ids", "id"})
}

func parseIntList(c *gin.Context, keys []string) []int {
	values := make([]string, 0)
	for _, key := range keys {
		values = append(values, c.QueryArray(key)...)
	}
	ids := make([]int, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		for _, part := range strings.Split(trimmed, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if id, err := strconv.Atoi(part); err == nil && id > 0 {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func buildNodeFilterFromQuery(c *gin.Context) models.NodeFilter {
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

	if maxFraudScoreStr := c.Query("maxFraudScore"); maxFraudScoreStr != "" {
		if maxFraudScore, err := strconv.Atoi(maxFraudScoreStr); err == nil && maxFraudScore > 0 {
			filter.MaxFraudScore = maxFraudScore
		}
	}

	filter.Countries = c.QueryArray("countries[]")
	filter.Tags = c.QueryArray("tags[]")
	filter.ResidentialType = normalizeResidentialType(c.Query("residentialType"))
	filter.IPType = normalizeIPType(c.Query("ipType"))
	filter.QualityStatus = normalizeQualityStatus(c.Query("qualityStatus"))
	filter.UnlockRuleMode = models.NormalizeUnlockRuleMode(c.Query("unlockRuleMode"))
	filter.UnlockRules = parseUnlockRulesFromQuery(c)
	filter.ExcludeIDs = parseExcludeIDs(c)
	if len(filter.UnlockRules) == 0 {
		filter.UnlockProvider = models.NormalizeUnlockProvider(c.Query("unlockProvider"))
		filter.UnlockStatus = normalizeUnlockStatus(c.Query("unlockStatus"))
		filter.UnlockKeyword = strings.TrimSpace(c.Query("unlockKeyword"))
	}
	if filter.ResidentialType == "" && c.Query("onlyResidential") == "true" {
		filter.ResidentialType = "residential"
	}
	if filter.IPType == "" && c.Query("onlyNative") == "true" {
		filter.IPType = "native"
	}

	// 验证排序字段
	validSortFields := map[string]bool{
		"delay":    true,
		"speed":    true,
		"name":     true,
		"protocol": true,
		"group":    true,
		"source":   true,
		"country":  true,
	}
	if filter.SortBy != "" && !validSortFields[filter.SortBy] {
		filter.SortBy = ""
	}

	if filter.SortOrder != "" && filter.SortOrder != "asc" && filter.SortOrder != "desc" {
		filter.SortOrder = "asc"
	}

	return filter
}

func parsePagination(c *gin.Context) (int, int) {
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
	return page, pageSize
}

func NodeUpdadte(c *gin.Context) {
	var Node models.Node
	name := c.PostForm("name")
	oldname := c.PostForm("oldname")
	oldlink := c.PostForm("oldlink")
	link := c.PostForm("link")
	nameMode, hasNameMode := c.GetPostForm("nameMode")
	if !hasNameMode {
		nameMode, hasNameMode = c.GetPostForm("NameMode")
	}
	dialerProxyName := c.PostForm("dialerProxyName")
	group := c.PostForm("group")
	if link == "" {
		utils.FailWithMsg(c, "Node link is required")
		return
	}
	userProvidedName := strings.TrimSpace(name) != ""
	// 查找旧节点
	Node.Name = oldname
	Node.Link = oldlink
	err := Node.Find()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	storedLink := Node.Link
	oldContentHash := Node.ContentHash
	if hasNameMode {
		Node.NameMode = models.NormalizeNodeNameMode(nameMode)
	} else {
		Node.NameMode = models.NormalizeNodeNameMode(Node.NameMode)
	}

	//更新构造节点元数据
	// 检测是否为 WireGuard 配置文件格式，如果是则转换为 URL 格式
	if protocol.IsWireGuardConfig(link) {
		wg, err := protocol.ParseWireGuardConfig(link)
		if err != nil {
			utils.FailWithMsg(c, "Failed to parse WireGuard configuration: "+err.Error())
			return
		}
		// 转换为 URL 格式
		link = protocol.EncodeWireGuardURL(wg)
	}
	identity, err := protocol.ExtractLinkIdentity(link)
	if err != nil {
		utils.Error("Failed to parse node link: %v", err)
		return
	}
	Node.Name = name
	if strings.TrimSpace(Node.Name) == "" {
		Node.Name = identity.Name
	}
	if userProvidedName {
		if err := models.EnsureNodeNameAvailable(Node.Name, Node.ID); err != nil {
			if !failNodeNameConflict(c, err) {
				utils.FailWithMsg(c, "Failed to check node remark")
			}
			return
		}
	} else {
		Node.Name = models.GenerateUniqueNodeNameWithSource(Node.Name, Node.Source, Node.ID, nil)
	}
	Node.LinkName = identity.Name
	Node.LinkAddress = identity.Address
	Node.LinkHost = identity.Host
	Node.LinkPort = identity.Port

	Node.Link = link
	if link != storedLink {
		// 链接发生变化后，旧链接对应的未建模 Clash 字段不再可靠，必须丢弃以避免错配。
		Node.ClashExtra = ""
	}
	Node.DialerProxyName = dialerProxyName
	Node.Group = group
	Node.Protocol = protocol.GetProtocolFromLink(link)

	// 重新计算 ContentHash
	proxy, proxyErr := protocol.LinkToProxy(protocol.Urls{Url: link}, protocol.OutputConfig{})
	if proxyErr == nil {
		contentHash := protocol.GenerateProxyContentHash(proxy)
		if contentHash != "" {
			Node.ContentHash = contentHash
			// 内容未变化时无需重复校验（避免历史数据存在重复时无法正常改名/改分组）
			if contentHash != oldContentHash {
				// 读取全局配置：是否启用跨机场去重（默认启用）
				crossAirportDedupVal, _ := models.GetSetting("cross_airport_dedup_enabled")
				enableCrossDedup := crossAirportDedupVal != "false"

				// 检查是否与其他节点重复（排除自身）
				var dupNode *models.Node
				var exists bool
				if enableCrossDedup {
					dupNode, exists = models.GetOtherNodeByContentHash(contentHash, Node.ID)
				} else {
					dupNode, exists = models.GetOtherNodeByContentHashAndSourceID(contentHash, Node.SourceID, Node.ID)
				}
				if exists && dupNode != nil {
					// 构建详细的重复信息
					source := dupNode.Source
					if source == "" || source == "manual" {
						source = "Manual"
					}
					group := dupNode.Group
					if group == "" {
						group = "Ungrouped"
					}
					utils.FailWithMsg(c, "Node content already exists; duplicates: [source: "+source+"] [group: "+group+"] [name: "+dupNode.Name+"]")
					return
				}
			}
		}
	}

	err = Node.Update()
	if err != nil {
		if failNodeNameConflict(c, err) {
			return
		}
		utils.FailWithMsg(c, "Update failed")
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

	utils.OkWithMsg(c, "Updated successfully")
}

// 获取节点列表
func NodeGet(c *gin.Context) {
	var Node models.Node
	filter := buildNodeFilterFromQuery(c)
	page, pageSize := parsePagination(c)

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

func NodeSelector(c *gin.Context) {
	var node models.Node
	filter := buildNodeFilterFromQuery(c)
	page, pageSize := parsePagination(c)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}

	nodes, total, err := node.ListWithFiltersPaginated(filter, page, pageSize)
	if err != nil {
		utils.FailWithMsg(c, "node selector list error")
		return
	}
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	utils.OkDetailed(c, "node selector get", gin.H{
		"items":      models.ToNodeSelectorItems(nodes),
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	})
}

func NodeSelectorByIDs(c *gin.Context) {
	ids := parseSelectedIDs(c)
	if len(ids) == 0 {
		utils.OkDetailed(c, "node selector get by ids", []models.NodeSelectorItem{})
		return
	}
	nodes, err := models.GetNodesByIDs(ids)
	if err != nil {
		utils.FailWithMsg(c, "node selector get by ids error")
		return
	}
	items := models.ToNodeSelectorItems(nodes)
	sorted := make([]models.NodeSelectorItem, 0, len(items))
	indexMap := make(map[int]models.NodeSelectorItem, len(items))
	for _, item := range items {
		indexMap[item.ID] = item
	}
	for _, id := range ids {
		if item, ok := indexMap[id]; ok {
			sorted = append(sorted, item)
		}
	}
	utils.OkDetailed(c, "node selector get by ids", sorted)
}

// NodeGetIDs 获取符合过滤条件的所有节点ID（用于全选操作）
func NodeGetIDs(c *gin.Context) {
	var Node models.Node
	filter := buildNodeFilterFromQuery(c)

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
	nameModeInput := c.PostForm("nameMode")
	if nameModeInput == "" {
		nameModeInput = c.PostForm("NameMode")
	}
	nameMode := models.NormalizeNodeNameMode(nameModeInput)
	dialerProxyName := c.PostForm("dialerProxyName")
	group := c.PostForm("group")
	if link == "" {
		utils.FailWithMsg(c, "link is required")
		return
	}
	userProvidedName := strings.TrimSpace(name) != ""

	// 读取全局配置：是否启用跨机场去重（默认启用）
	crossAirportDedupVal, _ := models.GetSetting("cross_airport_dedup_enabled")
	enableCrossDedup := crossAirportDedupVal != "false"
	// 检测是否为 WireGuard 配置文件格式，如果是则转换为 URL 格式
	if protocol.IsWireGuardConfig(link) {
		wg, err := protocol.ParseWireGuardConfig(link)
		if err != nil {
			utils.FailWithMsg(c, "Failed to parse WireGuard configuration: "+err.Error())
			return
		}
		// 转换为 URL 格式
		link = protocol.EncodeWireGuardURL(wg)
	}

	// 检测是否为 Clash YAML 配置格式
	if strings.Contains(link, "proxies:") {
		var clashConfig node.ClashConfig
		if err := yaml.Unmarshal([]byte(link), &clashConfig); err == nil && len(clashConfig.Proxies) > 0 {
			// 成功解析为 Clash YAML 格式，处理每个代理节点
			var addedCount, failedCount int
			for _, proxy := range clashConfig.Proxies {
				proxyLink := node.GenerateProxyLink(proxy)
				if proxyLink == "" {
					failedCount++
					continue
				}
				clashExtra, extraErr := protocol.EncodeClashExtra(proxy)
				if extraErr != nil {
					utils.Warn("Node [%s] failed to serialize unknown Clash attributes; skipped: %v", proxy.Name, extraErr)
					failedCount++
					continue
				}
				// 创建节点并添加；Clash YAML 中的备注由系统生成，重复时自动追加编号。
				var n models.Node
				n.Name = models.GenerateUniqueNodeName(proxy.Name, 0, nil)
				n.NameMode = models.NodeNameModeLink
				n.Link = proxyLink
				n.ClashExtra = clashExtra
				n.LinkName = proxy.Name
				n.LinkHost = proxy.Server
				n.LinkPort = strconv.Itoa(proxy.Port.Int())
				n.LinkAddress = proxy.Server + ":" + n.LinkPort
				n.DialerProxyName = dialerProxyName
				n.Group = group
				n.Protocol = protocol.GetProtocolFromLink(proxyLink)

				// 生成 ContentHash
				contentHash := protocol.GenerateProxyContentHash(proxy)
				if contentHash != "" {
					n.ContentHash = contentHash
					// 检查是否已存在相同内容的节点（跨机场去重关闭时仅校验同来源）
					if enableCrossDedup {
						if _, exists := models.GetNodeByContentHash(contentHash); exists {
							failedCount++
							continue
						}
					} else if _, exists := models.GetNodeByContentHashAndSourceID(contentHash, 0); exists {
						failedCount++
						continue
					}
				}

				if err := n.Add(); err != nil {
					failedCount++
					continue
				}
				addedCount++
			}

			if addedCount == 0 {
				utils.FailWithMsg(c, "Clash YAML parsed successfully but no nodes could be added (all duplicates or unsupported format)")
				return
			}
			utils.OkWithMsg(c, "Clash YAML import finished, added "+strconv.Itoa(addedCount)+" nodes")
			return
		}
	}

	if !strings.Contains(link, "://") {
		utils.FailWithMsg(c, "link must contain :// or be a valid WireGuard/Clash YAML configuration file")
		return
	}
	Node.Name = name
	Node.NameMode = nameMode
	identity, err := protocol.ExtractLinkIdentity(link)
	if err != nil {
		utils.Error("Failed to parse node link: %v", err)
		return
	}
	if strings.TrimSpace(Node.Name) == "" {
		Node.Name = identity.Name
	}
	if userProvidedName {
		if err := models.EnsureNodeNameAvailable(Node.Name, 0); err != nil {
			if !failNodeNameConflict(c, err) {
				utils.FailWithMsg(c, "Failed to check node remark")
			}
			return
		}
	} else {
		Node.Name = models.GenerateUniqueNodeName(Node.Name, 0, nil)
	}
	Node.LinkName = identity.Name
	Node.LinkAddress = identity.Address
	Node.LinkHost = identity.Host
	Node.LinkPort = identity.Port

	Node.Link = link
	Node.DialerProxyName = dialerProxyName
	Node.Group = group
	Node.Protocol = protocol.GetProtocolFromLink(link)

	// 生成 ContentHash（用于全库去重）
	proxy, proxyErr := protocol.LinkToProxy(protocol.Urls{Url: link}, protocol.OutputConfig{})
	if proxyErr == nil {
		contentHash := protocol.GenerateProxyContentHash(proxy)
		if contentHash != "" {
			Node.ContentHash = contentHash
			// 检查是否已存在相同内容的节点（跨机场去重关闭时仅校验同来源）
			var existingNode *models.Node
			var exists bool
			if enableCrossDedup {
				existingNode, exists = models.GetNodeByContentHash(contentHash)
			} else {
				existingNode, exists = models.GetNodeByContentHashAndSourceID(contentHash, 0)
			}
			if exists && existingNode != nil {
				// 重复节点：返回成功响应 + 跳过标记，不终止添加流程
				dupSource := existingNode.Source
				if dupSource == "" || dupSource == "manual" {
					dupSource = "Manual"
				}
				dupGroup := existingNode.Group
				if dupGroup == "" {
					dupGroup = "Ungrouped"
				}
				utils.OkDetailed(c, "Duplicate node skipped", gin.H{
					"skipped": true,
					"duplicateInfo": gin.H{
						"name":         Node.Name,
						"source":       dupSource,
						"group":        dupGroup,
						"existingName": existingNode.Name,
					},
				})
				return
			}
		}
	}

	err = Node.Add()
	if err != nil {
		if failNodeNameConflict(c, err) {
			return
		}
		utils.FailWithMsg(c, "Add failed; check whether the node is a duplicate")
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

	utils.OkWithMsg(c, "Added successfully")
}

// 删除节点
func NodeDel(c *gin.Context) {
	var Node models.Node
	id := c.Query("id")
	if id == "" {
		utils.FailWithMsg(c, "id is required")
		return
	}
	x, _ := strconv.Atoi(id)
	Node.ID = x
	err := Node.Del()
	if err != nil {
		utils.FailWithMsg(c, "Delete failed")
		return
	}
	utils.OkWithMsg(c, "Deleted successfully")
}

// 节点统计
func NodesTotal(c *gin.Context) {
	var Node models.Node
	nodes, err := Node.List()
	if err != nil {
		utils.FailWithMsg(c, "Unable to fetch node statistics")
		return
	}

	total := len(nodes)
	available := 0
	delayPassCount := 0
	speedPassCount := 0
	for _, n := range nodes {
		if n.DelayStatus == "success" && n.DelayTime > 0 {
			delayPassCount++
		}
		if n.SpeedStatus == "success" && n.Speed > 0 {
			speedPassCount++
		}
		if n.Speed > 0 && n.DelayTime > 0 {
			available++
		}
	}

	utils.OkDetailed(c, "Node statistics fetched", gin.H{
		"total":          total,
		"available":      available,
		"delayPassCount": delayPassCount,
		"speedPassCount": speedPassCount,
	})
}

// NodeBatchDel 批量删除节点
func NodeBatchDel(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "Select nodes to delete")
		return
	}
	err := models.BatchDel(req.IDs)
	if err != nil {
		utils.FailWithMsg(c, "Batch delete failed")
		return
	}
	utils.OkWithMsg(c, "Batch delete completed")
}

// NodeBatchUpdateGroup 批量更新节点分组
func NodeBatchUpdateGroup(c *gin.Context) {
	var req struct {
		IDs   []int  `json:"ids"`
		Group string `json:"group"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "Select nodes to modify")
		return
	}
	err := models.BatchUpdateGroup(req.IDs, req.Group)
	if err != nil {
		utils.FailWithMsg(c, "Failed to update groups in batch")
		return
	}
	utils.OkWithMsg(c, "Groups updated in batch")
}

// NodeBatchUpdateDialerProxy 批量更新节点前置代理
func NodeBatchUpdateDialerProxy(c *gin.Context) {
	var req struct {
		IDs             []int  `json:"ids"`
		DialerProxyName string `json:"dialerProxyName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "Select nodes to modify")
		return
	}
	err := models.BatchUpdateDialerProxy(req.IDs, req.DialerProxyName)
	if err != nil {
		utils.FailWithMsg(c, "Failed to update dialer proxies in batch")
		return
	}
	utils.OkWithMsg(c, "Dialer proxies updated in batch")
}

// NodeBatchUpdateSource 批量更新节点来源
func NodeBatchUpdateSource(c *gin.Context) {
	var req struct {
		IDs    []int  `json:"ids"`
		Source string `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "Select nodes to modify")
		return
	}
	err := models.BatchUpdateSource(req.IDs, req.Source)
	if err != nil {
		utils.FailWithMsg(c, "Failed to update sources in batch")
		return
	}
	utils.OkWithMsg(c, "Sources updated in batch")
}

// NodeBatchUpdateCountry 批量更新节点国家代码
func NodeBatchUpdateCountry(c *gin.Context) {
	var req struct {
		IDs     []int  `json:"ids"`
		Country string `json:"country"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "Select nodes to modify")
		return
	}
	// 国家代码转大写，保持一致性
	country := strings.ToUpper(strings.TrimSpace(req.Country))
	err := models.BatchUpdateCountry(req.IDs, country)
	if err != nil {
		utils.FailWithMsg(c, "Failed to update country codes in batch")
		return
	}
	utils.OkWithMsg(c, "Country codes updated in batch")
}

// NodeBatchFillCountry 批量填充节点国家信息
func NodeBatchFillCountry(c *gin.Context) {
	var req struct {
		AirportID int  `json:"airportId"` // 可选,指定机场
		OnlyEmpty bool `json:"onlyEmpty"` // 只填充国家为空的节点
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}

	// 构建查询条件
	query := database.DB.Model(&models.Node{})
	if req.AirportID > 0 {
		query = query.Where("source_id = ?", req.AirportID)
	}
	if req.OnlyEmpty {
		query = query.Where("link_country IS NULL OR link_country = ''")
	}

	var nodes []models.Node
	if err := query.Find(&nodes).Error; err != nil {
		utils.FailWithMsg(c, "Failed to query nodes")
		return
	}

	updated := 0
	for i := range nodes {
		if req.OnlyEmpty && nodes[i].LinkCountry != "" {
			continue
		}
		if country := models.ParseCountryFromNodeName(nodes[i].Name); country != "" {
			nodes[i].LinkCountry = country
			if err := database.DB.Model(&nodes[i]).Update("link_country", country).Error; err == nil {
				updated++
			}
		}
	}

	utils.OkDetailed(c, "Batch fill completed", gin.H{
		"total":   len(nodes),
		"updated": updated,
	})
}

// 获取所有分组列表
func GetGroups(c *gin.Context) {
	var node models.Node
	groups, err := node.GetAllGroups()
	if err != nil {
		utils.FailWithMsg(c, "Failed to fetch group list")
		return
	}
	utils.OkDetailed(c, "Group list fetched", groups)
}

// GetSources 获取所有来源列表
func GetSources(c *gin.Context) {
	var node models.Node
	sources, err := node.GetAllSources()
	if err != nil {
		utils.FailWithMsg(c, "Failed to fetch source list")
		return
	}
	utils.OkDetailed(c, "Source list fetched", sources)
}

// FastestSpeedNode 获取最快速度节点
func FastestSpeedNode(c *gin.Context) {
	node := models.GetFastestSpeedNode()
	utils.OkDetailed(c, "Fastest node fetched", node)
}

// LowestDelayNode 获取最低延迟节点
func LowestDelayNode(c *gin.Context) {
	node := models.GetLowestDelayNode()
	utils.OkDetailed(c, "Lowest-latency node fetched", node)
}

// GetNodeCountries 获取所有节点的国家代码列表
func GetNodeCountries(c *gin.Context) {
	countries := models.GetAllCountries()
	utils.OkDetailed(c, "Country code fetched", countries)
}

// NodeCountryStats 获取按国家统计的节点数量
func NodeCountryStats(c *gin.Context) {
	stats := models.GetNodeCountryStats()
	utils.OkDetailed(c, "Country statistics fetched successfully", stats)
}

// NodeCountryStatsFiltered 获取符合过滤条件的按国家统计节点数量
func NodeCountryStatsFiltered(c *gin.Context) {
	var Node models.Node
	filter := buildNodeFilterFromQuery(c)
	items, total := Node.GetFilteredCountryStats(filter)
	utils.OkDetailed(c, "Country statistics fetched successfully", gin.H{
		"items": items,
		"total": total,
	})
}

func DashboardCountryStats(c *gin.Context) {
	stats := models.GetDashboardCountryStats()
	utils.OkDetailed(c, "Dashboard country statistics fetched", stats)
}

// NodeProtocolStats 获取按协议统计的节点数量
func NodeProtocolStats(c *gin.Context) {
	stats := models.GetNodeProtocolStats()
	utils.OkDetailed(c, "Protocol statistics fetched", stats)
}

// NodeTagStats 获取按标签统计的节点数量
func NodeTagStats(c *gin.Context) {
	stats := models.GetNodeTagStats()
	utils.OkDetailed(c, "Tag statistics fetched", stats)
}

// NodeGroupStats 获取按分组统计的节点数量
func NodeGroupStats(c *gin.Context) {
	stats := models.GetNodeGroupStats()
	utils.OkDetailed(c, "Group statistics fetched", stats)
}

// NodeSourceStats 获取按来源统计的节点数量
func NodeSourceStats(c *gin.Context) {
	stats := models.GetNodeSourceStats()
	utils.OkDetailed(c, "Source statistics fetched", stats)
}

func DashboardGroupedStats(c *gin.Context) {
	stats := models.GetDashboardGroupedStats()
	utils.OkDetailed(c, "Dashboard grouped statistics fetched", stats)
}

func DashboardQualityStats(c *gin.Context) {
	stats := models.GetDashboardQualityStats()
	utils.OkDetailed(c, "Dashboard quality statistics fetched", stats)
}

// GetIPDetails 获取IP详细信息
// GET /api/v1/nodes/ip-info?ip=xxx.xxx.xxx.xxx
func GetIPDetails(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		utils.FailWithMsg(c, "IP address is required")
		return
	}

	// 调用模型层获取IP信息（多级缓存）
	ipInfo, err := models.GetIPInfo(ip)
	if err != nil {
		utils.FailWithMsg(c, "Failed to query IP information: "+err.Error())
		return
	}

	utils.OkDetailed(c, "Fetched successfully", ipInfo)
}

// GetIPCacheStats 获取IP缓存统计
// GET /api/v1/nodes/ip-cache/stats
func GetIPCacheStats(c *gin.Context) {
	count := models.GetIPInfoCount()
	utils.OkDetailed(c, "Fetched successfully", gin.H{
		"count": count,
	})
}

// ClearIPCache 清除所有IP缓存
// DELETE /api/v1/nodes/ip-cache
func ClearIPCache(c *gin.Context) {
	err := models.ClearAllIPInfo()
	if err != nil {
		utils.FailWithMsg(c, "Clear failed: "+err.Error())
		return
	}
	utils.OkWithMsg(c, "IP cache cleared")
}

// GetNodeProtocols 获取所有使用中的协议类型列表（用于过滤器选项）
// GET /api/v1/nodes/protocols
func GetNodeProtocols(c *gin.Context) {
	protocols := models.GetAllProtocols()
	utils.OkDetailed(c, "Protocol list fetched", protocols)
}
