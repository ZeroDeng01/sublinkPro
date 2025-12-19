package api

import (
	"encoding/json"
	"net/http"
	"sublink/database"
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// PreviewRequest 预览请求结构
type PreviewRequest struct {
	Nodes              []string `json:"Nodes"`              // 选中的节点名称列表
	Groups             []string `json:"Groups"`             // 选中的分组列表
	Scripts            []int    `json:"Scripts"`            // 选中的脚本ID列表
	DelayTime          int      `json:"DelayTime"`          // 最大延迟过滤
	MinSpeed           float64  `json:"MinSpeed"`           // 最小速度过滤
	CountryWhitelist   string   `json:"CountryWhitelist"`   // 国家白名单
	CountryBlacklist   string   `json:"CountryBlacklist"`   // 国家黑名单
	TagWhitelist       string   `json:"TagWhitelist"`       // 标签白名单
	TagBlacklist       string   `json:"TagBlacklist"`       // 标签黑名单
	ProtocolWhitelist  string   `json:"ProtocolWhitelist"`  // 协议白名单（逗号分隔）
	ProtocolBlacklist  string   `json:"ProtocolBlacklist"`  // 协议黑名单（逗号分隔）
	NodeNameWhitelist  string   `json:"NodeNameWhitelist"`  // 节点名称白名单
	NodeNameBlacklist  string   `json:"NodeNameBlacklist"`  // 节点名称黑名单
	NodeNamePreprocess string   `json:"NodeNamePreprocess"` // 原名预处理规则
	NodeNameRule       string   `json:"NodeNameRule"`       // 节点命名规则模板
	DeduplicationRule  string   `json:"DeduplicationRule"`  // 去重规则配置
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

	// 验证至少选择了节点或分组
	if len(req.Nodes) == 0 && len(req.Groups) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "请至少选择节点或分组",
		})
		return
	}

	// 构建临时订阅对象（不保存到数据库）
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

	// 获取节点列表
	var allNodes []models.Node
	totalCount := 0

	// 从名称获取节点
	if len(req.Nodes) > 0 {
		for _, nodeName := range req.Nodes {
			if node, ok := models.GetNodeByName(nodeName); ok {
				allNodes = append(allNodes, *node)
			}
		}
	}

	// 从分组获取节点
	if len(req.Groups) > 0 {
		for _, groupName := range req.Groups {
			var groupNodes []models.Node
			node := &models.Node{}
			groupNodes, _ = node.ListByGroups([]string{groupName})
			// 去重添加
			nodeMap := make(map[string]bool)
			for _, n := range allNodes {
				nodeMap[n.Name] = true
			}
			for _, n := range groupNodes {
				if !nodeMap[n.Name] {
					allNodes = append(allNodes, n)
					nodeMap[n.Name] = true
				}
			}
		}
	}

	totalCount = len(allNodes)

	// 应用脚本处理（filterNode 脚本）
	if len(req.Scripts) > 0 && len(allNodes) > 0 {
		// 获取脚本列表
		for _, scriptID := range req.Scripts {
			var script models.Script
			if err := database.DB.Where("id = ?", scriptID).First(&script).Error; err != nil {
				continue // 跳过不存在的脚本
			}

			// 将节点转换为 JSON
			nodesJSON, err := json.Marshal(allNodes)
			if err != nil {
				continue
			}

			// 执行 filterNode 脚本（使用 Content 字段）
			resultJSON, err := utils.RunNodeFilterScript(script.Content, nodesJSON, "preview")
			if err != nil {
				// 脚本执行失败，继续使用原始节点
				continue
			}

			// 解析脚本处理后的节点
			var processedNodes []models.Node
			if err := json.Unmarshal(resultJSON, &processedNodes); err != nil {
				continue
			}

			allNodes = processedNodes
		}
	}

	tempSub.Nodes = allNodes

	// 调用 PreviewSub 方法获取预览结果
	result, err := tempSub.PreviewSub()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "预览处理失败: " + err.Error(),
		})
		return
	}

	// 更新原始节点数
	result.TotalCount = totalCount

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": result,
	})
}
