package api

import (
	"encoding/json"
	"strconv"
	"sublink/database"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		utils.FailWithMsg(c, "链接不能为空")
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
	NodeID int                    `json:"nodeId"` // 节点 ID
	Fields map[string]interface{} `json:"fields"` // 要更新的字段
}

// UpdateNodeRawInfo 更新节点原始信息
// POST /api/v1/nodes/update-raw
func UpdateNodeRawInfo(c *gin.Context) {
	var req UpdateNodeRawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "请求参数错误")
		return
	}

	if req.NodeID <= 0 {
		utils.FailWithMsg(c, "节点ID无效")
		return
	}

	// 获取节点
	var node models.Node
	node.ID = req.NodeID
	if err := node.GetByID(); err != nil {
		utils.FailWithMsg(c, "节点不存在")
		return
	}

	// 将字段转为 JSON
	fieldsJSON, err := json.Marshal(req.Fields)
	if err != nil {
		utils.FailWithMsg(c, "字段序列化失败")
		return
	}

	// 更新链接
	newLink, err := protocol.UpdateNodeLinkFields(node.Link, string(fieldsJSON))
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	// 检查新 Link 是否与其他节点冲突
	var existingNode models.Node
	err = database.DB.Where("link = ? AND id != ?", newLink, req.NodeID).First(&existingNode).Error
	if err == nil {
		utils.FailWithMsg(c, "已存在相同连接的节点: "+existingNode.Name)
		return
	} else if err != gorm.ErrRecordNotFound {
		utils.FailWithMsg(c, "检查节点冲突失败")
		return
	}

	// 解析新链接以获取可能更新的名称
	newInfo, err := protocol.ParseNodeLink(newLink)
	if err != nil {
		utils.FailWithMsg(c, "解析新链接失败")
		return
	}

	// 获取节点名称（不同协议名称字段不同）
	newLinkName := getNameFromFields(newInfo.Protocol, newInfo.Fields)

	// 更新数据库
	updates := map[string]interface{}{
		"link": newLink,
	}
	if newLinkName != "" {
		updates["link_name"] = newLinkName
		// 如果原始名称和显示名称一致，同步更新显示名称
		if node.LinkName == node.Name {
			updates["name"] = newLinkName
		}
	}

	err = database.DB.Model(&models.Node{}).Where("id = ?", req.NodeID).Updates(updates).Error
	if err != nil {
		utils.FailWithMsg(c, "更新数据库失败")
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
	models.UpdateNodeCache(req.NodeID, node)

	utils.OkWithData(c, gin.H{
		"link":     newLink,
		"linkName": newLinkName,
	})
}

// getNameFromFields 从字段中提取节点名称
func getNameFromFields(protocol string, fields map[string]interface{}) string {
	// 不同协议的名称字段不同
	switch protocol {
	case "vmess":
		if ps, ok := fields["Ps"].(string); ok {
			return ps
		}
	case "ssr":
		if remarks, ok := fields["Qurey.Remarks"].(string); ok {
			return remarks
		}
	default:
		// 大多数协议使用 Name 字段
		if name, ok := fields["Name"].(string); ok {
			return name
		}
	}
	return ""
}

// GetNodeRawInfo 获取节点原始信息
// GET /api/v1/nodes/raw-info?id=xxx
func GetNodeRawInfo(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		utils.FailWithMsg(c, "节点ID不能为空")
		return
	}

	nodeID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.FailWithMsg(c, "节点ID格式错误")
		return
	}

	// 获取节点
	var node models.Node
	node.ID = nodeID
	if err := node.GetByID(); err != nil {
		utils.FailWithMsg(c, "节点不存在")
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
