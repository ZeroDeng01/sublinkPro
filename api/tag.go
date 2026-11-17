package api

import (
	"strconv"
	"strings"
	"sublink/models"
	"sublink/services"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

// ========== Tag API ==========

// TagGet 获取标签列表
func TagGet(c *gin.Context) {
	var tag models.Tag
	tags, err := tag.List()
	if err != nil {
		utils.FailWithMsg(c, "Failed to fetch tag list")
		return
	}
	utils.OkWithData(c, tags)
}

// TagAdd 添加标签
func TagAdd(c *gin.Context) {
	var tag models.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	tag.Name = strings.TrimSpace(tag.Name)
	// 名称会进入以逗号分隔的节点标签列表，新建时就拦下非法名称
	if err := models.ValidateTagName(tag.Name); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	// 检查标签是否已存在
	if models.TagExists(tag.Name) {
		utils.FailWithMsg(c, "Tag name already exists")
		return
	}
	if tag.Color == "" {
		tag.Color = "#1976d2"
	}
	if err := tag.Add(); err != nil {
		utils.FailWithMsg(c, "Failed to add tag: "+err.Error())
		return
	}
	utils.OkWithData(c, tag)
}

// TagUpdate 更新标签；newName 与 name 不同时连同所有引用一起重命名
func TagUpdate(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		NewName     string `json:"newName"`
		GroupName   string `json:"groupName"`
		Color       string `json:"color"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if req.Name == "" {
		utils.FailWithMsg(c, "Tag name is required")
		return
	}
	var tag models.Tag
	if err := tag.GetByName(req.Name); err != nil {
		utils.FailWithCode(c, 404, "Tag not found")
		return
	}

	// 先校验新名称，避免颜色/描述已写入后才发现重命名不可用
	newName := strings.TrimSpace(req.NewName)
	renaming := newName != "" && newName != tag.Name
	if renaming {
		if err := models.ValidateTagRename(newName); err != nil {
			utils.FailWithMsg(c, err.Error())
			return
		}
	}

	tag.GroupName = req.GroupName
	if req.Color != "" {
		tag.Color = req.Color
	}
	tag.Description = req.Description
	if err := tag.Update(); err != nil {
		utils.FailWithMsg(c, "Failed to update tag: "+err.Error())
		return
	}

	if renaming {
		oldName := tag.Name
		if err := tag.Rename(newName); err != nil {
			utils.FailWithMsg(c, "Failed to rename tag: "+err.Error())
			return
		}
		utils.Info("Tag renamed: %s -> %s", oldName, newName)
	}
	utils.OkWithData(c, tag)
}

// TagDelete 删除标签
func TagDelete(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		utils.FailWithMsg(c, "Tag name is required")
		return
	}
	var tag models.Tag
	if err := tag.GetByName(name); err != nil {
		utils.FailWithCode(c, 404, "Tag not found")
		return
	}
	if err := tag.Delete(); err != nil {
		utils.FailWithMsg(c, "Failed to delete tag: "+err.Error())
		return
	}
	utils.Ok(c)
}

// TagGroupList 获取所有标签组名称（用于前端自动补全）
func TagGroupList(c *gin.Context) {
	groups := models.GetExistingGroups()
	utils.OkWithData(c, groups)
}

// ========== TagRule API ==========

// TagRuleGet 获取规则列表
func TagRuleGet(c *gin.Context) {
	var rule models.TagRule
	rules, err := rule.List()
	if err != nil {
		utils.FailWithMsg(c, "Failed to fetch rule list")
		return
	}
	utils.OkWithData(c, rules)
}

// TagRuleAdd 添加规则
func TagRuleAdd(c *gin.Context) {
	var rule models.TagRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if rule.Name == "" {
		utils.FailWithMsg(c, "Rule name is required")
		return
	}
	if rule.TagName == "" {
		utils.FailWithMsg(c, "Associated tag is required")
		return
	}
	// 验证标签是否存在
	if !models.TagExists(rule.TagName) {
		utils.FailWithMsg(c, "Associated tag not found")
		return
	}
	// 验证条件格式
	if rule.Conditions != "" {
		if _, err := models.ParseConditions(rule.Conditions); err != nil {
			utils.FailWithMsg(c, "Invalid condition format: "+err.Error())
			return
		}
	}
	if err := rule.Add(); err != nil {
		utils.FailWithMsg(c, "Failed to add rule: "+err.Error())
		return
	}
	utils.OkWithData(c, rule)
}

// TagRuleUpdate 更新规则
func TagRuleUpdate(c *gin.Context) {
	var req struct {
		ID          int    `json:"id"`
		TagName     string `json:"tagName"`
		Name        string `json:"name"`
		Enabled     bool   `json:"enabled"`
		TriggerType string `json:"triggerType"`
		Conditions  string `json:"conditions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if req.ID == 0 {
		utils.FailWithMsg(c, "Rule ID is required")
		return
	}
	var rule models.TagRule
	if err := rule.GetByID(req.ID); err != nil {
		utils.FailWithMsg(c, "Rule not found")
		return
	}
	rule.TagName = req.TagName
	rule.Name = req.Name
	rule.Enabled = req.Enabled
	rule.TriggerType = req.TriggerType
	rule.Conditions = req.Conditions
	// 验证标签是否存在
	if rule.TagName != "" && !models.TagExists(rule.TagName) {
		utils.FailWithMsg(c, "Associated tag not found")
		return
	}
	// 验证条件格式
	if rule.Conditions != "" {
		if _, err := models.ParseConditions(rule.Conditions); err != nil {
			utils.FailWithMsg(c, "Invalid condition format: "+err.Error())
			return
		}
	}
	if err := rule.Update(); err != nil {
		utils.FailWithMsg(c, "Failed to update rule: "+err.Error())
		return
	}
	utils.OkWithData(c, rule)
}

// TagRuleDelete 删除规则
func TagRuleDelete(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	var rule models.TagRule
	rule.ID = id
	if err := rule.Delete(); err != nil {
		utils.FailWithMsg(c, "Failed to delete rule: "+err.Error())
		return
	}
	utils.Ok(c)
}

// TagRuleTrigger 手动触发规则
func TagRuleTrigger(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	go func() {
		if err := services.TriggerTagRule(id); err != nil {
			utils.Error("Failed to trigger tag rule: %v", err)
		}
	}()
	utils.OkWithMsg(c, "Rule execution started")
}

// ========== Node Tag API ==========

// NodeAddTag 给节点添加标签
func NodeAddTag(c *gin.Context) {
	var req struct {
		NodeID  int    `json:"nodeId"`
		TagName string `json:"tagName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	var node models.Node
	node.ID = req.NodeID
	if err := node.GetByID(); err != nil {
		utils.FailWithCode(c, 404, "Node not found")
		return
	}
	if err := node.AddTagByName(req.TagName); err != nil {
		utils.FailWithMsg(c, "Failed to add tag")
		return
	}
	utils.Ok(c)
}

// NodeRemoveTag 从节点移除标签
func NodeRemoveTag(c *gin.Context) {
	var req struct {
		NodeID  int    `json:"nodeId"`
		TagName string `json:"tagName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	var node models.Node
	node.ID = req.NodeID
	if err := node.GetByID(); err != nil {
		utils.FailWithCode(c, 404, "Node not found")
		return
	}
	if err := node.RemoveTagByName(req.TagName); err != nil {
		utils.FailWithMsg(c, "Failed to remove tag")
		return
	}
	utils.Ok(c)
}

// NodeBatchAddTag 批量给节点添加标签
func NodeBatchAddTag(c *gin.Context) {
	var req struct {
		NodeIDs []int  `json:"nodeIds"`
		TagName string `json:"tagName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if err := models.BatchAddTagToNodes(req.NodeIDs, req.TagName); err != nil {
		utils.FailWithMsg(c, "Failed to add tags in batch")
		return
	}
	utils.Ok(c)
}

// NodeBatchSetTags 批量设置节点标签（覆盖模式）
func NodeBatchSetTags(c *gin.Context) {
	var req struct {
		NodeIDs  []int    `json:"nodeIds"`
		TagNames []string `json:"tagNames"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if err := models.BatchSetTagsForNodes(req.NodeIDs, req.TagNames); err != nil {
		utils.FailWithMsg(c, "Failed to set tags in batch")
		return
	}
	utils.Ok(c)
}

// NodeBatchRemoveTags 批量从节点移除指定标签
func NodeBatchRemoveTags(c *gin.Context) {
	var req struct {
		NodeIDs  []int    `json:"nodeIds"`
		TagNames []string `json:"tagNames"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	if len(req.TagNames) == 0 {
		utils.FailWithMsg(c, "Select tags to delete")
		return
	}
	if err := models.BatchRemoveTagsFromNodes(req.NodeIDs, req.TagNames); err != nil {
		utils.FailWithMsg(c, "Failed to remove tags in batch")
		return
	}
	utils.Ok(c)
}

// GetNodeTags 获取节点的标签
func GetNodeTags(c *gin.Context) {
	nodeIDStr := c.Query("nodeId")
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil || nodeID == 0 {
		utils.FailWithMsg(c, "Invalid parameters")
		return
	}
	var node models.Node
	node.ID = nodeID
	if err := node.GetByID(); err != nil {
		utils.FailWithCode(c, 404, "Node not found")
		return
	}
	tags := models.GetTagsByNode(node)
	utils.OkWithData(c, tags)
}
