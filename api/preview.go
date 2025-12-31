package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"sublink/database"
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// PreviewRequest 预览请求结构
type PreviewRequest struct {
	// 如果提供 SubscriptionID，则直接从数据库获取订阅配置，使用 GetSub 逻辑
	SubscriptionID int `json:"SubscriptionID"`

	// 以下字段用于表单预览（未保存的订阅）
	NodeIDs            []int    `json:"NodeIDs"`            // 选中的节点ID列表（带排序）
	NodeSorts          []int    `json:"NodeSorts"`          // 节点对应的排序值
	Groups             []string `json:"Groups"`             // 选中的分组列表
	GroupSorts         []int    `json:"GroupSorts"`         // 分组对应的排序值
	Scripts            []int    `json:"Scripts"`            // 选中的脚本ID列表
	DelayTime          int      `json:"DelayTime"`          // 最大延迟过滤
	MinSpeed           float64  `json:"MinSpeed"`           // 最小速度过滤
	CountryWhitelist   string   `json:"CountryWhitelist"`   // 国家白名单
	CountryBlacklist   string   `json:"CountryBlacklist"`   // 国家黑名单
	TagWhitelist       string   `json:"TagWhitelist"`       // 标签白名单
	TagBlacklist       string   `json:"TagBlacklist"`       // 标签黑名单
	ProtocolWhitelist  string   `json:"ProtocolWhitelist"`  // 协议白名单
	ProtocolBlacklist  string   `json:"ProtocolBlacklist"`  // 协议黑名单
	NodeNameWhitelist  string   `json:"NodeNameWhitelist"`  // 节点名称白名单
	NodeNameBlacklist  string   `json:"NodeNameBlacklist"`  // 节点名称黑名单
	NodeNamePreprocess string   `json:"NodeNamePreprocess"` // 原名预处理规则
	NodeNameRule       string   `json:"NodeNameRule"`       // 节点命名规则模板
	DeduplicationRule  string   `json:"DeduplicationRule"`  // 去重规则配置

	// 兼容旧版本：节点名称列表（已废弃，保留向后兼容）
	Nodes []interface{} `json:"Nodes"` // 可以是节点ID或节点名称
}

// PreviewSubscriptionNodes 预览订阅节点
// 该接口接受订阅的配置参数，在内存中模拟过滤和重命名逻辑，返回预览结果
func PreviewSubscriptionNodes(c *gin.Context) {
	var req PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请求参数解析失败: " + err.Error(),
		})
		return
	}

	var result *models.PreviewResult
	var err error

	// 如果提供了 SubscriptionID，直接从数据库加载并使用 GetSub 逻辑
	if req.SubscriptionID > 0 {
		result, err = previewSavedSubscription(req.SubscriptionID)
	} else {
		result, err = previewFormSubscription(req)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": result,
	})
}

// previewSavedSubscription 预览已保存的订阅
// 使用与实际拉取完全相同的 GetSub 逻辑，确保预览结果与拉取结果一致
func previewSavedSubscription(subID int) (*models.PreviewResult, error) {
	sub, err := models.GetSubcriptionByID(subID)
	if err != nil {
		return nil, err
	}

	// 使用与 GetSub 完全相同的逻辑获取节点
	// GetSub 已包含：节点/分组混合排序、过滤规则、脚本执行
	if err := sub.GetSub("preview"); err != nil {
		return nil, err
	}

	// 直接构建预览结果，不再重复调用 ApplyFilters
	// 因为 GetSub 已经完成了所有过滤和脚本处理
	totalCount := len(sub.Nodes)
	filteredCount := len(sub.Nodes)

	// 计算用量信息
	upload, download, total, expire := sub.CalculateUsageInfo()

	// 构建预览节点列表
	previewNodes := make([]models.PreviewNode, 0, filteredCount)

	for idx, node := range sub.Nodes {
		// 应用预处理规则到 LinkName
		processedLinkName := utils.PreprocessNodeName(sub.NodeNamePreprocess, node.LinkName)

		// 计算预览名称
		previewName := node.Name
		previewLink := node.Link

		if sub.NodeNameRule != "" {
			previewName = utils.RenameNode(sub.NodeNameRule, utils.NodeInfo{
				Name:        node.Name,
				LinkName:    processedLinkName,
				LinkCountry: node.LinkCountry,
				Speed:       node.Speed,
				DelayTime:   node.DelayTime,
				Group:       node.Group,
				Source:      node.Source,
				Index:       idx + 1,
				Protocol:    utils.GetProtocolFromLink(node.Link),
				Tags:        node.Tags,
			})
			previewLink = utils.RenameNodeLink(node.Link, previewName)
		}

		previewNodes = append(previewNodes, models.PreviewNode{
			Node:         node,
			OriginalName: node.LinkName,
			PreviewName:  previewName,
			PreviewLink:  previewLink,
			Protocol:     utils.GetProtocolFromLink(node.Link),
			CountryFlag:  utils.ISOToFlag(node.LinkCountry),
		})
	}

	return &models.PreviewResult{
		Nodes:         previewNodes,
		TotalCount:    totalCount,
		FilteredCount: filteredCount,
		UsageUpload:   upload,
		UsageDownload: download,
		UsageTotal:    total,
		UsageExpire:   expire,
	}, nil
}

// previewFormSubscription 预览表单中的订阅（未保存）
func previewFormSubscription(req PreviewRequest) (*models.PreviewResult, error) {
	// 构建临时订阅对象
	tempSub := &models.Subcription{
		DelayTime:          req.DelayTime,
		MinSpeed:           req.MinSpeed,
		CountryWhitelist:   req.CountryWhitelist,
		CountryBlacklist:   req.CountryBlacklist,
		TagWhitelist:       req.TagWhitelist,
		TagBlacklist:       req.TagBlacklist,
		ProtocolWhitelist:  req.ProtocolWhitelist,
		ProtocolBlacklist:  req.ProtocolBlacklist,
		NodeNameWhitelist:  req.NodeNameWhitelist,
		NodeNameBlacklist:  req.NodeNameBlacklist,
		NodeNamePreprocess: req.NodeNamePreprocess,
		NodeNameRule:       req.NodeNameRule,
		DeduplicationRule:  req.DeduplicationRule,
	}

	// 使用与 GetSub 相同的混合排序逻辑构建节点列表
	allNodes := buildNodesWithMixedSort(req)
	totalCount := len(allNodes)

	// 应用脚本处理（filterNode 脚本）
	if len(req.Scripts) > 0 && len(allNodes) > 0 {
		for _, scriptID := range req.Scripts {
			var script models.Script
			if err := database.DB.Where("id = ?", scriptID).First(&script).Error; err != nil {
				continue // 跳过不存在的脚本
			}

			nodesJSON, err := json.Marshal(allNodes)
			if err != nil {
				continue
			}

			resultJSON, err := utils.RunNodeFilterScript(script.Content, nodesJSON, "preview")
			if err != nil {
				continue
			}

			var processedNodes []models.Node
			if err := json.Unmarshal(resultJSON, &processedNodes); err != nil {
				continue
			}

			allNodes = processedNodes
		}
	}

	tempSub.Nodes = allNodes

	// 调用 PreviewSub 方法获取预览结果（包含过滤和重命名）
	result, err := tempSub.PreviewSub()
	if err != nil {
		return nil, err
	}

	// 更新原始节点数
	result.TotalCount = totalCount

	return result, nil
}

// buildNodesWithMixedSort 使用与 GetSub 相同的混合排序逻辑构建节点列表
func buildNodesWithMixedSort(req PreviewRequest) []models.Node {
	// 定义混合排序项
	type MixedItem struct {
		Node    *models.Node
		Group   string
		Sort    int
		IsGroup bool
	}

	var mixedItems []MixedItem

	// 处理节点ID列表（带排序）
	if len(req.NodeIDs) > 0 {
		for i, nodeID := range req.NodeIDs {
			if node, ok := models.GetNodeByID(nodeID); ok {
				sortVal := i // 默认使用索引作为排序值
				if i < len(req.NodeSorts) {
					sortVal = req.NodeSorts[i]
				}
				mixedItems = append(mixedItems, MixedItem{
					Node:    node,
					Sort:    sortVal,
					IsGroup: false,
				})
			}
		}
	}

	// 兼容旧版本：处理 Nodes 字段（可能是节点ID或节点名称）
	if len(req.Nodes) > 0 && len(req.NodeIDs) == 0 {
		for i, nodeVal := range req.Nodes {
			var node *models.Node
			var ok bool

			// 尝试作为数字ID解析
			switch v := nodeVal.(type) {
			case float64:
				node, ok = models.GetNodeByID(int(v))
			case int:
				node, ok = models.GetNodeByID(v)
			case string:
				node, ok = models.GetNodeByName(v)
			}

			if ok && node != nil {
				mixedItems = append(mixedItems, MixedItem{
					Node:    node,
					Sort:    i, // 旧版本没有排序信息，使用索引
					IsGroup: false,
				})
			}
		}
	}

	// 处理分组（带排序）
	if len(req.Groups) > 0 {
		for i, groupName := range req.Groups {
			sortVal := len(mixedItems) + i // 默认在节点后面
			if i < len(req.GroupSorts) {
				sortVal = req.GroupSorts[i]
			}
			mixedItems = append(mixedItems, MixedItem{
				Group:   groupName,
				Sort:    sortVal,
				IsGroup: true,
			})
		}
	}

	// 验证至少有节点或分组
	if len(mixedItems) == 0 {
		return nil
	}

	// 按排序值排序混合列表
	sort.Slice(mixedItems, func(i, j int) bool {
		return mixedItems[i].Sort < mixedItems[j].Sort
	})

	// 获取分组节点映射
	groupNodeMap := make(map[string][]models.Node)
	for _, item := range mixedItems {
		if item.IsGroup {
			var groupNodes []models.Node
			node := &models.Node{}
			groupNodes, _ = node.ListByGroups([]string{item.Group})
			groupNodeMap[item.Group] = groupNodes
		}
	}

	// 按排序后的顺序构建最终节点列表（与 GetSub 逻辑一致）
	nodeMap := make(map[string]bool) // 用于按名称去重
	var result []models.Node

	for _, item := range mixedItems {
		if item.IsGroup {
			// 添加分组中的所有节点
			if nodes, exists := groupNodeMap[item.Group]; exists {
				for _, node := range nodes {
					if !nodeMap[node.Name] {
						result = append(result, node)
						nodeMap[node.Name] = true
					}
				}
			}
		} else {
			// 添加单个节点
			if item.Node != nil && !nodeMap[item.Node.Name] {
				result = append(result, *item.Node)
				nodeMap[item.Node.Name] = true
			}
		}
	}

	return result
}
