package api

import (
	"log"
	"net/url"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/node"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func NodeUpdadte(c *gin.Context) {
	var node models.Node
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
	node.Name = oldname
	node.Link = oldlink
	err := node.Find()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	node.Name = name
	node.Link = link
	node.DialerProxyName = dialerProxyName
	node.Group = group
	err = node.Update()
	if err != nil {
		utils.FailWithMsg(c, "更新失败")
		return
	}
	utils.OkWithMsg(c, "更新成功")
}

// 获取节点列表
func NodeGet(c *gin.Context) {
	var Node models.Node
	nodes, err := Node.List()
	if err != nil {
		utils.FailWithMsg(c, "node list error")
		return
	}
	utils.OkDetailed(c, "node get", nodes)
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
	if name == "" {
		u, err := url.Parse(link)
		if err != nil {
			log.Println(err)
			return
		}
		switch {
		case u.Scheme == "ss":
			ss, err := node.DecodeSSURL(link)
			if err != nil {
				log.Println(err)
				return
			}
			Node.Name = ss.Name
		case u.Scheme == "ssr":
			ssr, err := node.DecodeSSRURL(link)
			if err != nil {
				log.Println(err)
				return
			}
			Node.Name = ssr.Qurey.Remarks
		case u.Scheme == "trojan":
			trojan, err := node.DecodeTrojanURL(link)
			if err != nil {
				log.Println(err)
				return
			}
			Node.Name = trojan.Name
		case u.Scheme == "vmess":
			vmess, err := node.DecodeVMESSURL(link)
			if err != nil {
				log.Println(err)
				return
			}
			Node.Name = vmess.Ps
		case u.Scheme == "vless":
			vless, err := node.DecodeVLESSURL(link)
			if err != nil {
				log.Println(err)
				return
			}
			Node.Name = vless.Name
		case u.Scheme == "hy" || u.Scheme == "hysteria":
			hy, err := node.DecodeHYURL(link)
			if err != nil {
				log.Println(err)
				return
			}
			Node.Name = hy.Name
		case u.Scheme == "hy2" || u.Scheme == "hysteria2":
			hy2, err := node.DecodeHY2URL(link)
			if err != nil {
				log.Println(err)
				return
			}
			Node.Name = hy2.Name
		case u.Scheme == "tuic":
			tuic, err := node.DecodeTuicURL(link)
			if err != nil {
				log.Println(err)
				return
			}
			Node.Name = tuic.Name
		}
	}
	Node.Link = link
	Node.DialerProxyName = dialerProxyName
	Node.Group = group
	Node.CreateDate = time.Now().Format("2006-01-02 15:04:05")
	err := Node.Find()
	// 如果找到记录说明重复
	if err == nil {
		Node.Name = name + " " + time.Now().Format("2006-01-02 15:04:05")
	}
	err = Node.Add()
	if err != nil {
		utils.FailWithMsg(c, "添加失败检查一下是否节点重复")
		return
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
	count := len(nodes)
	if err != nil {
		utils.FailWithMsg(c, "获取不到节点统计")
		return
	}
	utils.OkDetailed(c, "取得节点统计", count)
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
