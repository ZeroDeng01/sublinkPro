package api

import (
	"strconv"
	"strings"
	"sublink/dto"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func SubTotal(c *gin.Context) {
	var Sub models.Subcription
	subs, err := Sub.List()
	count := len(subs)
	if err != nil {
		utils.FailWithMsg(c, "取得订阅总数失败")
		return
	}
	utils.OkDetailed(c, "取得订阅总数", count)
}

// 获取订阅列表
func SubGet(c *gin.Context) {
	var Sub models.Subcription

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
		subs, total, err := Sub.ListPaginated(page, pageSize)
		if err != nil {
			utils.FailWithMsg(c, "获取订阅列表失败")
			return
		}
		totalPages := 0
		if pageSize > 0 {
			totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
		}
		utils.OkDetailed(c, "获取成功", gin.H{
			"items":      subs,
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		})
		return
	}

	// 不带分页参数，返回全部（向后兼容）
	Subs, err := Sub.List()
	if err != nil {
		utils.FailWithMsg(c, "node list error")
		return
	}
	utils.OkDetailed(c, "node get", Subs)
}

// 添加节点
func SubAdd(c *gin.Context) {
	var sub models.Subcription
	name := c.PostForm("name")
	config := c.PostForm("config")
	nodeIds := c.PostForm("nodeIds") // 改为接收节点ID列表
	groups := c.PostForm("groups")   // 新增：分组列表
	scripts := c.PostForm("scripts") // 新增：脚本列表
	ipWhitelist := c.PostForm("IPWhitelist")
	ipBlacklist := c.PostForm("IPBlacklist")
	delayTimeStr := c.PostForm("DelayTime")
	delayTime, _ := strconv.Atoi(delayTimeStr)
	minSpeedStr := c.PostForm("MinSpeed")
	minSpeed, _ := strconv.ParseFloat(minSpeedStr, 64)
	countryWhitelist := c.PostForm("CountryWhitelist")
	countryBlacklist := c.PostForm("CountryBlacklist")
	nodeNameRule := c.PostForm("NodeNameRule")
	nodeNamePreprocess := c.PostForm("NodeNamePreprocess")
	nodeNameWhitelist := c.PostForm("NodeNameWhitelist")
	nodeNameBlacklist := c.PostForm("NodeNameBlacklist")
	tagWhitelist := c.PostForm("TagWhitelist")
	tagBlacklist := c.PostForm("TagBlacklist")
	protocolWhitelist := c.PostForm("ProtocolWhitelist")
	protocolBlacklist := c.PostForm("ProtocolBlacklist")
	deduplicationRule := c.PostForm("DeduplicationRule")
	refreshUsageOnRequestStr := c.PostForm("RefreshUsageOnRequest")
	refreshUsageOnRequest := refreshUsageOnRequestStr != "false" // 默认为 true

	if name == "" || (nodeIds == "" && groups == "") {
		utils.FailWithMsg(c, "订阅名称不能为空，且节点或分组至少选择一项")
		return
	}
	if ipWhitelist != "" {
		ok := utils.IpFormatValidation(ipWhitelist)
		if !ok {
			utils.FailWithMsg(c, "IP白名单有误，请检查IP格式")
			return
		}
	}
	if ipBlacklist != "" {
		ok := utils.IpFormatValidation(ipBlacklist)
		if !ok {
			utils.FailWithMsg(c, "IP黑名单有误，请检查IP格式")
			return
		}
	}

	// 检查订阅名称是否重复
	var checkSub models.Subcription
	checkSub.Name = name
	if err := checkSub.Find(); err == nil {
		utils.FailWithMsg(c, "订阅名称不能重复")
		return
	}

	sub.Nodes = []models.Node{}
	if nodeIds != "" {
		nodeNameSet := make(map[string]bool) // 按节点名称去重（Clash等客户端不支持重名节点）
		for _, v := range strings.Split(nodeIds, ",") {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			id, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			var node models.Node
			node.ID = id
			err = node.GetByID() // 直接用ID获取节点
			if err != nil {
				continue
			}
			// 按节点名称去重，同名节点只保留第一个
			if !nodeNameSet[node.Name] {
				nodeNameSet[node.Name] = true
				sub.Nodes = append(sub.Nodes, node)
			}
		}
	}

	sub.Config = config
	sub.Name = name
	sub.IPWhitelist = ipWhitelist
	sub.IPBlacklist = ipBlacklist
	sub.DelayTime = delayTime
	sub.MinSpeed = minSpeed
	sub.CountryWhitelist = countryWhitelist
	sub.CountryBlacklist = countryBlacklist
	sub.NodeNameRule = nodeNameRule
	sub.NodeNamePreprocess = nodeNamePreprocess
	sub.NodeNameWhitelist = nodeNameWhitelist
	sub.NodeNameBlacklist = nodeNameBlacklist
	sub.TagWhitelist = tagWhitelist
	sub.TagBlacklist = tagBlacklist
	sub.ProtocolWhitelist = protocolWhitelist
	sub.ProtocolBlacklist = protocolBlacklist
	sub.DeduplicationRule = deduplicationRule
	sub.RefreshUsageOnRequest = refreshUsageOnRequest
	sub.CreateDate = time.Now().Format("2006-01-02 15:04:05")

	err := sub.Add()
	if err != nil {
		utils.FailWithMsg(c, "添加失败")
		return
	}

	// 添加节点关系
	if len(sub.Nodes) > 0 {
		err = sub.AddNode()
		if err != nil {
			utils.FailWithMsg(c, err.Error())
			return
		}
	}

	// 添加分组关系
	if groups != "" {
		err = sub.AddGroups(strings.Split(groups, ","))
		if err != nil {
			utils.FailWithMsg(c, err.Error())
			return
		}
	}

	// 添加脚本关系
	if scripts != "" {
		scriptIDs := make([]int, 0)
		for _, s := range strings.Split(scripts, ",") {
			id, err := strconv.Atoi(s)
			if err == nil {
				scriptIDs = append(scriptIDs, id)
			}
		}
		if len(scriptIDs) > 0 {
			err = sub.AddScripts(scriptIDs)
			if err != nil {
				utils.FailWithMsg(c, err.Error())
				return
			}
		}
	}

	utils.OkWithMsg(c, "添加成功")
}

// 更新节点
func SubUpdate(c *gin.Context) {
	var sub models.Subcription
	name := c.PostForm("name")
	oldname := c.PostForm("oldname")
	config := c.PostForm("config")
	nodeIds := c.PostForm("nodeIds") // 改为接收节点ID列表
	groups := c.PostForm("groups")   // 新增：分组列表
	scripts := c.PostForm("scripts") // 新增：脚本列表
	ipWhitelist := c.PostForm("IPWhitelist")
	ipBlacklist := c.PostForm("IPBlacklist")
	delayTimeStr := c.PostForm("DelayTime")
	delayTime, _ := strconv.Atoi(delayTimeStr)
	minSpeedStr := c.PostForm("MinSpeed")
	minSpeed, _ := strconv.ParseFloat(minSpeedStr, 64)
	countryWhitelist := c.PostForm("CountryWhitelist")
	countryBlacklist := c.PostForm("CountryBlacklist")
	nodeNameRule := c.PostForm("NodeNameRule")
	nodeNamePreprocess := c.PostForm("NodeNamePreprocess")
	nodeNameWhitelist := c.PostForm("NodeNameWhitelist")
	nodeNameBlacklist := c.PostForm("NodeNameBlacklist")
	tagWhitelist := c.PostForm("TagWhitelist")
	tagBlacklist := c.PostForm("TagBlacklist")
	protocolWhitelist := c.PostForm("ProtocolWhitelist")
	protocolBlacklist := c.PostForm("ProtocolBlacklist")
	deduplicationRule := c.PostForm("DeduplicationRule")
	refreshUsageOnRequestStr := c.PostForm("RefreshUsageOnRequest")
	refreshUsageOnRequest := refreshUsageOnRequestStr != "false" // 默认为 true

	if name == "" || (nodeIds == "" && groups == "") {
		utils.FailWithMsg(c, "订阅名称不能为空，且节点或分组至少选择一项")
		return
	}
	if ipWhitelist != "" {
		ok := utils.IpFormatValidation(ipWhitelist)
		if !ok {
			utils.FailWithMsg(c, "IP白名单有误，请检查IP格式")
			return
		}
	}
	if ipBlacklist != "" {
		ok := utils.IpFormatValidation(ipBlacklist)
		if !ok {
			utils.FailWithMsg(c, "IP黑名单有误，请检查IP格式")
			return
		}
	}

	// 检查订阅名称是否重复
	if name != oldname {
		var checkSub models.Subcription
		checkSub.Name = name
		if err := checkSub.Find(); err == nil {
			utils.FailWithMsg(c, "订阅名称不能重复")
			return
		}
	}

	// 查找旧节点
	sub.Name = oldname
	err := sub.Find()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	// 更新节点
	sub.Config = config
	sub.Name = name
	sub.CreateDate = time.Now().Format("2006-01-02 15:04:05")
	sub.Nodes = []models.Node{}
	if nodeIds != "" {
		nodeNameSet := make(map[string]bool) // 按节点名称去重（Clash等客户端不支持重名节点）
		for _, v := range strings.Split(nodeIds, ",") {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			id, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			var node models.Node
			node.ID = id
			err = node.GetByID() // 直接用ID获取节点
			if err != nil {
				continue
			}
			// 按节点名称去重，同名节点只保留第一个
			if !nodeNameSet[node.Name] {
				nodeNameSet[node.Name] = true
				sub.Nodes = append(sub.Nodes, node)
			}
		}
	}
	sub.IPWhitelist = ipWhitelist
	sub.IPBlacklist = ipBlacklist
	sub.DelayTime = delayTime
	sub.MinSpeed = minSpeed
	sub.CountryWhitelist = countryWhitelist
	sub.CountryBlacklist = countryBlacklist
	sub.NodeNameRule = nodeNameRule
	sub.NodeNamePreprocess = nodeNamePreprocess
	sub.NodeNameWhitelist = nodeNameWhitelist
	sub.NodeNameBlacklist = nodeNameBlacklist
	sub.TagWhitelist = tagWhitelist
	sub.TagBlacklist = tagBlacklist
	sub.ProtocolWhitelist = protocolWhitelist
	sub.ProtocolBlacklist = protocolBlacklist
	sub.DeduplicationRule = deduplicationRule
	sub.RefreshUsageOnRequest = refreshUsageOnRequest
	err = sub.Update()
	if err != nil {
		utils.FailWithMsg(c, "更新失败")
		return
	}

	// 更新节点关系
	err = sub.UpdateNodes()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	// 更新分组关系
	if groups != "" {
		err = sub.UpdateGroups(strings.Split(groups, ","))
	} else {
		err = sub.UpdateGroups([]string{})
	}
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	// 更新脚本关系
	if scripts != "" {
		scriptIDs := make([]int, 0)
		for _, s := range strings.Split(scripts, ",") {
			id, err := strconv.Atoi(s)
			if err == nil {
				scriptIDs = append(scriptIDs, id)
			}
		}
		err = sub.UpdateScripts(scriptIDs)
	} else {
		err = sub.UpdateScripts([]int{})
	}
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	utils.OkWithMsg(c, "更新成功")
}

// 删除节点
func SubDel(c *gin.Context) {
	var sub models.Subcription
	id := c.Query("id")
	if id == "" {
		utils.FailWithMsg(c, "id 不能为空")
		return
	}
	x, _ := strconv.Atoi(id)
	sub.ID = x
	err := sub.Find()
	if err != nil {
		utils.FailWithMsg(c, "查找失败")
		return
	}
	err = sub.Del()
	if err != nil {
		utils.FailWithMsg(c, "删除失败")
		return
	}
	utils.OkWithMsg(c, "删除成功")
}

// SubCopy 复制订阅
func SubCopy(c *gin.Context) {
	var sub models.Subcription
	id := c.Query("id")
	if id == "" {
		utils.FailWithMsg(c, "id 不能为空")
		return
	}
	x, err := strconv.Atoi(id)
	if err != nil {
		utils.FailWithMsg(c, "id 格式错误")
		return
	}
	sub.ID = x
	err = sub.Find()
	if err != nil {
		utils.FailWithMsg(c, "订阅不存在")
		return
	}

	// 执行复制
	newSub, err := sub.Copy()
	if err != nil {
		utils.FailWithMsg(c, "复制失败: "+err.Error())
		return
	}

	utils.OkDetailed(c, "复制成功", newSub)
}

func SubSort(c *gin.Context) {
	var subNodeSort dto.SubcriptionNodeSortUpdate
	err := c.BindJSON(&subNodeSort)
	if err != nil {
		utils.FailWithMsg(c, "参数错误: "+err.Error())
		return
	}

	var sub models.Subcription
	sub.ID = subNodeSort.ID
	err = sub.Sort(subNodeSort)

	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkWithMsg(c, "更新排序成功")
}

// SubBatchSort 批量排序订阅节点
func SubBatchSort(c *gin.Context) {
	var req dto.BatchSortRequest
	if err := c.BindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误: "+err.Error())
		return
	}

	// 验证排序字段
	validSortBy := map[string]bool{
		"source": true, "name": true, "protocol": true,
		"delay": true, "speed": true, "country": true,
	}
	if !validSortBy[req.SortBy] {
		utils.FailWithMsg(c, "无效的排序字段")
		return
	}

	// 验证排序方向
	if req.SortOrder != "asc" && req.SortOrder != "desc" {
		utils.FailWithMsg(c, "无效的排序方向")
		return
	}

	var sub models.Subcription
	sub.ID = req.ID
	if err := sub.Find(); err != nil {
		utils.FailWithMsg(c, "订阅不存在")
		return
	}

	if err := sub.BatchSort(req.SortBy, req.SortOrder); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}

	utils.OkWithMsg(c, "批量排序成功")
}

// GetProtocolMeta 获取协议元数据（协议列表及其可用字段）
func GetProtocolMeta(c *gin.Context) {
	meta := protocol.GetAllProtocolMeta()
	utils.OkDetailed(c, "获取成功", meta)
}

// GetNodeFieldsMeta 获取节点通用字段元数据
func GetNodeFieldsMeta(c *gin.Context) {
	meta := models.GetNodeFieldsMeta()
	utils.OkDetailed(c, "获取成功", meta)
}
