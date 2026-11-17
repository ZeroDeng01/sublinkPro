package api

import (
	"encoding/json"
	"errors"
	"strconv"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// GetProtocolUIMeta 获取协议 UI 元数据（包含颜色、图标等）
// GET /api/v1/nodes/protocol-ui-meta
func GetProtocolUIMeta(c *gin.Context) {
	metas := protocol.GetAllProtocolMeta()
	utils.OkWithData(c, metas)
}

// ParseNodeLinkAPI 解析节点链接
// GET /api/v1/nodes/parse-link?link=xxx
func ParseNodeLinkAPI(c *gin.Context) {
	link := c.Query("link")
	if link == "" {
		utils.FailWithMsg(c, "Link is required")
		return
	}

	info, err := protocol.ParseNodeLink(link)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	utils.OkWithData(c, info)
}

// UpdateNodeRawRequest 更新节点原始信息请求
type UpdateNodeRawRequest struct {
	NodeID int            `json:"nodeId"` // 节点 ID
	Fields map[string]any `json:"fields"` // 要更新的字段
}

// UpdateNodeRawInfo 更新节点原始信息
// POST /api/v1/nodes/update-raw
func UpdateNodeRawInfo(c *gin.Context) {
	var req UpdateNodeRawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid request parameters")
		return
	}

	if req.NodeID <= 0 {
		utils.FailWithMsg(c, "Invalid node ID")
		return
	}

	// 获取节点
	var node models.Node
	node.ID = req.NodeID
	if err := node.GetByID(); err != nil {
		utils.FailWithMsg(c, "Node not found")
		return
	}

	// 将字段转为 JSON
	fieldsJSON, err := json.Marshal(req.Fields)
	if err != nil {
		utils.FailWithMsg(c, "Failed to serialize fields")
		return
	}

	// 更新链接
	newLink, err := protocol.UpdateNodeLinkFields(node.Link, string(fieldsJSON))
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	conflictNode, conflict, err := models.FindNodeLinkConflict(newLink, req.NodeID)
	if err != nil {
		utils.FailWithMsg(c, "Failed to check node conflicts")
		return
	}
	if conflict {
		utils.FailWithMsg(c, "A node with the same connection already exists: "+conflictNode.Name)
		return
	}

	// 解析新链接以获取可能更新的名称
	newInfo, err := protocol.ParseNodeLink(newLink)
	if err != nil {
		utils.FailWithMsg(c, "Failed to parse the new link")
		return
	}

	// 获取节点名称（不同协议名称字段不同）
	newLinkName := protocol.ExtractNodeNameFromFields(newInfo.Protocol, newInfo.Fields)

	// 更新数据库
	updates := map[string]any{
		"link": newLink,
	}
	if newLinkName != "" {
		updates["link_name"] = newLinkName
		// link 模式或历史“备注=原始名称”的节点继续同步备注；remark 模式下保留用户自定义备注。
		if node.ShouldSyncNameFromLink() {
			updates["name"] = newLinkName
		}
	}

	if err := models.UpdateNodeFields(req.NodeID, updates); err != nil {
		if errors.Is(err, models.ErrNodeNameExists) {
			utils.FailWithMsg(c, "This node remark already exists; choose another one")
			return
		}
		utils.FailWithMsg(c, "Failed to update the database")
		return
	}

	// 更新缓存
	node.Link = newLink
	if newLinkName != "" {
		node.LinkName = newLinkName
		if _, ok := updates["name"]; ok {
			node.Name = newLinkName
		}
	}
	node.NameMode = models.NormalizeNodeNameMode(node.NameMode)
	models.UpdateNodeCache(req.NodeID, node)

	utils.OkWithData(c, gin.H{
		"link":     newLink,
		"linkName": newLinkName,
	})
}

// GetNodeRawInfo 获取节点原始信息
// GET /api/v1/nodes/raw-info?id=xxx
func GetNodeRawInfo(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		utils.FailWithMsg(c, "Node ID is required")
		return
	}

	nodeID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "Invalid node ID format")
		return
	}

	// 获取节点
	var node models.Node
	node.ID = nodeID
	if err := node.GetByID(); err != nil {
		utils.FailWithMsg(c, "Node not found")
		return
	}

	// 解析链接
	info, err := protocol.ParseNodeLink(node.Link)
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	utils.OkWithData(c, info)
}
