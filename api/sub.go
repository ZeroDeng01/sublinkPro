package api

import (
	"strconv"
	"strings"
	"sublink/dto"
	"sublink/models"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func SubTotal(c *gin.Context) {
	var Sub models.Subcription
	subs, err := Sub.List()
	count := len(subs)
	if err != nil {
		c.JSON(500, gin.H{
			"msg": "取得订阅总数失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"data": count,
		"msg":  "取得订阅总数",
	})
}

// 获取订阅列表
func SubGet(c *gin.Context) {
	var Sub models.Subcription
	Subs, err := Sub.List()
	if err != nil {
		c.JSON(500, gin.H{
			"msg": "node list error",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"data": Subs,
		"msg":  "node get",
	})
}

// 添加节点
func SubAdd(c *gin.Context) {
	var sub models.Subcription
	name := c.PostForm("name")
	config := c.PostForm("config")
	nodes := c.PostForm("nodes")
	groups := c.PostForm("groups") // 新增：分组列表
	ipWhitelist := c.PostForm("IPWhitelist")
	ipBlacklist := c.PostForm("IPBlacklist")
	if name == "" || (nodes == "" && groups == "") {
		c.JSON(400, gin.H{
			"msg": "订阅名称不能为空，且节点或分组至少选择一项",
		})
		return
	}
	if ipWhitelist != "" {
		ok := utils.IpFormatValidation(ipWhitelist)
		if !ok {
			c.JSON(400, gin.H{
				"msg": "IP白名单有误，请检查IP格式",
			})
			return
		}
	}
	if ipBlacklist != "" {
		ok := utils.IpFormatValidation(ipBlacklist)
		if !ok {
			c.JSON(400, gin.H{
				"msg": "IP黑名单有误，请检查IP格式",
			})
			return
		}
	}

	sub.Nodes = []models.Node{}
	if nodes != "" {
		for _, v := range strings.Split(nodes, ",") {
			var node models.Node
			node.Name = v
			err := node.Find()
			if err != nil {
				continue
			}
			sub.Nodes = append(sub.Nodes, node)
		}
	}

	sub.Config = config
	sub.Name = name
	sub.IPWhitelist = ipWhitelist
	sub.IPBlacklist = ipBlacklist
	sub.CreateDate = time.Now().Format("2006-01-02 15:04:05")

	err := sub.Add()
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "添加失败",
		})
		return
	}

	// 添加节点关系
	if len(sub.Nodes) > 0 {
		err = sub.AddNode()
		if err != nil {
			c.JSON(400, gin.H{
				"msg": err.Error(),
			})
			return
		}
	}

	// 添加分组关系
	if groups != "" {
		err = sub.AddGroups(strings.Split(groups, ","))
		if err != nil {
			c.JSON(400, gin.H{
				"msg": err.Error(),
			})
			return
		}
	}

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "添加成功",
	})
}

// 更新节点
func SubUpdate(c *gin.Context) {
	var sub models.Subcription
	name := c.PostForm("name")
	oldname := c.PostForm("oldname")
	config := c.PostForm("config")
	nodes := c.PostForm("nodes")
	groups := c.PostForm("groups") // 新增：分组列表
	ipWhitelist := c.PostForm("IPWhitelist")
	ipBlacklist := c.PostForm("IPBlacklist")
	if name == "" || (nodes == "" && groups == "") {
		c.JSON(400, gin.H{
			"msg": "订阅名称不能为空，且节点或分组至少选择一项",
		})
		return
	}
	if ipWhitelist != "" {
		ok := utils.IpFormatValidation(ipWhitelist)
		if !ok {
			c.JSON(400, gin.H{
				"msg": "IP白名单有误，请检查IP格式",
			})
			return
		}
	}
	if ipBlacklist != "" {
		ok := utils.IpFormatValidation(ipBlacklist)
		if !ok {
			c.JSON(400, gin.H{
				"msg": "IP黑名单有误，请检查IP格式",
			})
			return
		}
	}
	// 查找旧节点
	sub.Name = oldname
	err := sub.Find()
	if err != nil {
		c.JSON(400, gin.H{
			"msg": err.Error(),
		})
		return
	}
	// 更新节点
	sub.Config = config
	sub.Name = name
	sub.CreateDate = time.Now().Format("2006-01-02 15:04:05")
	sub.Nodes = []models.Node{}
	if nodes != "" {
		for _, v := range strings.Split(nodes, ",") {
			var node models.Node
			node.Name = v
			err := node.Find()
			if err != nil {
				continue
			}
			sub.Nodes = append(sub.Nodes, node)
		}
	}
	sub.IPWhitelist = ipWhitelist
	sub.IPBlacklist = ipBlacklist
	err = sub.Update()
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "更新失败",
		})
		return
	}

	// 更新节点关系
	err = sub.UpdateNodes()
	if err != nil {
		c.JSON(400, gin.H{
			"msg": err.Error(),
		})
		return
	}

	// 更新分组关系
	if groups != "" {
		err = sub.UpdateGroups(strings.Split(groups, ","))
	} else {
		err = sub.UpdateGroups([]string{})
	}
	if err != nil {
		c.JSON(400, gin.H{
			"msg": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "更新成功",
	})
}

// 删除节点
func SubDel(c *gin.Context) {
	var sub models.Subcription
	id := c.Query("id")
	if id == "" {
		c.JSON(400, gin.H{
			"msg": "id 不能为空",
		})
		return
	}
	x, _ := strconv.Atoi(id)
	sub.ID = x
	err := sub.Find()
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "查找失败",
		})
		return
	}
	err = sub.Del()
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "删除失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "删除成功",
	})
}

func SubSort(c *gin.Context) {
	var subNodeSort dto.SubcriptionNodeSortUpdate
	err := c.BindJSON(&subNodeSort)
	if err != nil {
		c.JSON(400, gin.H{"msg": "参数错误: " + err.Error()})
		return
	}

	var sub models.Subcription
	sub.ID = subNodeSort.ID
	err = sub.Sort(subNodeSort)

	if err != nil {
		c.JSON(400, gin.H{
			"msg": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "更新排序成功",
	})
}
