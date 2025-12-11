package api

import (
	"strconv"
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
		utils.FailWithMsg(c, "获取标签列表失败")
		return
	}
	utils.OkWithData(c, tags)
}

// TagAdd 添加标签
func TagAdd(c *gin.Context) {
	var tag models.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if tag.Name == "" {
		utils.FailWithMsg(c, "标签名称不能为空")
		return
	}
	// 检查标签是否已存在
	if models.TagExists(tag.Name) {
		utils.FailWithMsg(c, "标签名称已存在")
		return
	}
	if tag.Color == "" {
		tag.Color = "#1976d2"
	}
	if err := tag.Add(); err != nil {
		utils.FailWithMsg(c, "添加标签失败: "+err.Error())
		return
	}
	utils.OkWithData(c, tag)
}

// TagUpdate 更新标签
func TagUpdate(c *gin.Context) {
	var tag models.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if tag.Name == "" {
		utils.FailWithMsg(c, "标签名称不能为空")
		return
	}
	if err := tag.Update(); err != nil {
		utils.FailWithMsg(c, "更新标签失败: "+err.Error())
		return
	}
	utils.OkWithData(c, tag)
}

// TagDelete 删除标签
func TagDelete(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		utils.FailWithMsg(c, "标签名称不能为空")
		return
	}
	var tag models.Tag
	if err := tag.GetByName(name); err != nil {
		utils.FailWithCode(c, 404, "标签不存在")
		return
	}
	if err := tag.Delete(); err != nil {
		utils.FailWithMsg(c, "删除标签失败: "+err.Error())
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
		utils.FailWithMsg(c, "获取规则列表失败")
		return
	}
	utils.OkWithData(c, rules)
}

// TagRuleAdd 添加规则
func TagRuleAdd(c *gin.Context) {
	var rule models.TagRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if rule.Name == "" {
		utils.FailWithMsg(c, "规则名称不能为空")
		return
	}
	if rule.TagName == "" {
		utils.FailWithMsg(c, "关联标签不能为空")
		return
	}
	// 验证标签是否存在
	if !models.TagExists(rule.TagName) {
		utils.FailWithMsg(c, "关联的标签不存在")
		return
	}
	// 验证条件格式
	if rule.Conditions != "" {
		if _, err := models.ParseConditions(rule.Conditions); err != nil {
			utils.FailWithMsg(c, "条件格式错误: "+err.Error())
			return
		}
	}
	if err := rule.Add(); err != nil {
		utils.FailWithMsg(c, "添加规则失败: "+err.Error())
		return
	}
	utils.OkWithData(c, rule)
}

// TagRuleUpdate 更新规则
func TagRuleUpdate(c *gin.Context) {
	var rule models.TagRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if rule.ID == 0 {
		utils.FailWithMsg(c, "规则ID不能为空")
		return
	}
	// 验证标签是否存在
	if rule.TagName != "" && !models.TagExists(rule.TagName) {
		utils.FailWithMsg(c, "关联的标签不存在")
		return
	}
	// 验证条件格式
	if rule.Conditions != "" {
		if _, err := models.ParseConditions(rule.Conditions); err != nil {
			utils.FailWithMsg(c, "条件格式错误: "+err.Error())
			return
		}
	}
	if err := rule.Update(); err != nil {
		utils.FailWithMsg(c, "更新规则失败: "+err.Error())
		return
	}
	utils.OkWithData(c, rule)
}

// TagRuleDelete 删除规则
func TagRuleDelete(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	var rule models.TagRule
	rule.ID = id
	if err := rule.Delete(); err != nil {
		utils.FailWithMsg(c, "删除规则失败: "+err.Error())
		return
	}
	utils.Ok(c)
}

// TagRuleTrigger 手动触发规则
func TagRuleTrigger(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	go func() {
		if err := services.TriggerTagRule(id); err != nil {
			// 记录错误日志
		}
	}()
	utils.OkWithMsg(c, "规则已开始执行")
}

// ========== Node Tag API ==========

// NodeAddTag 给节点添加标签
func NodeAddTag(c *gin.Context) {
	var req struct {
		NodeID  int    `json:"nodeId"`
		TagName string `json:"tagName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	var node models.Node
	node.ID = req.NodeID
	if err := node.GetByID(); err != nil {
		utils.FailWithCode(c, 404, "节点不存在")
		return
	}
	if err := node.AddTagByName(req.TagName); err != nil {
		utils.FailWithMsg(c, "添加标签失败")
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
		utils.FailWithMsg(c, "参数错误")
		return
	}
	var node models.Node
	node.ID = req.NodeID
	if err := node.GetByID(); err != nil {
		utils.FailWithCode(c, 404, "节点不存在")
		return
	}
	if err := node.RemoveTagByName(req.TagName); err != nil {
		utils.FailWithMsg(c, "移除标签失败")
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
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if err := models.BatchAddTagToNodes(req.NodeIDs, req.TagName); err != nil {
		utils.FailWithMsg(c, "批量添加标签失败")
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
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if err := models.BatchSetTagsForNodes(req.NodeIDs, req.TagNames); err != nil {
		utils.FailWithMsg(c, "批量设置标签失败")
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
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if len(req.TagNames) == 0 {
		utils.FailWithMsg(c, "请选择要删除的标签")
		return
	}
	if err := models.BatchRemoveTagsFromNodes(req.NodeIDs, req.TagNames); err != nil {
		utils.FailWithMsg(c, "批量移除标签失败")
		return
	}
	utils.Ok(c)
}

// GetNodeTags 获取节点的标签
func GetNodeTags(c *gin.Context) {
	nodeIDStr := c.Query("nodeId")
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil || nodeID == 0 {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	var node models.Node
	node.ID = nodeID
	if err := node.GetByID(); err != nil {
		utils.FailWithCode(c, 404, "节点不存在")
		return
	}
	tags := models.GetTagsByNode(node)
	utils.OkWithData(c, tags)
}
