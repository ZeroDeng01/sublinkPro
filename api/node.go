package api

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/services"
	"sublink/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func NodeUpdadte(c *gin.Context) {
	var Node models.Node
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
	Node.Name = oldname
	Node.Link = oldlink
	err := Node.Find()
	if err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	Node.Name = name

	//更新构造节点元数据
	u, err := url.Parse(link)
	if err != nil {
		log.Println(err)
		return
	}
	switch {
	case u.Scheme == "ss":
		ss, err := protocol.DecodeSSURL(link)
		if err != nil {
			log.Println(err)
			return
		}
		if Node.Name == "" {
			Node.Name = ss.Name
		}
		Node.LinkName = ss.Name
		Node.LinkAddress = ss.Server + ":" + strconv.Itoa(ss.Port)
		Node.LinkHost = ss.Server
		Node.LinkPort = strconv.Itoa(ss.Port)
	case u.Scheme == "ssr":
		ssr, err := protocol.DecodeSSRURL(link)
		if err != nil {
			log.Println(err)
			return
		}
		if Node.Name == "" {
			Node.Name = ssr.Qurey.Remarks
		}
		Node.LinkName = ssr.Qurey.Remarks
		Node.LinkAddress = ssr.Server + ":" + strconv.Itoa(ssr.Port)
		Node.LinkHost = ssr.Server
		Node.LinkPort = strconv.Itoa(ssr.Port)
	case u.Scheme == "trojan":
		trojan, err := protocol.DecodeTrojanURL(link)
		if err != nil {
			log.Println(err)
			return
		}

		if Node.Name == "" {
			Node.Name = trojan.Name
		}
		Node.LinkName = trojan.Name
		Node.LinkAddress = trojan.Hostname + ":" + strconv.Itoa(trojan.Port)
		Node.LinkHost = trojan.Hostname
		Node.LinkPort = strconv.Itoa(trojan.Port)
	case u.Scheme == "vmess":
		vmess, err := protocol.DecodeVMESSURL(link)
		if err != nil {
			log.Println(err)
			return
		}
		if Node.Name == "" {
			Node.Name = vmess.Ps
		}
		Node.LinkName = vmess.Ps
		prot := fmt.Sprintf("%v", vmess.Port)
		Node.LinkAddress = vmess.Add + ":" + prot
		Node.LinkHost = vmess.Host
		Node.LinkPort = prot
	case u.Scheme == "vless":
		vless, err := protocol.DecodeVLESSURL(link)
		if err != nil {
			log.Println(err)
			return
		}
		if Node.Name == "" {
			Node.Name = vless.Name
		}
		Node.LinkName = vless.Name
		Node.LinkAddress = vless.Server + ":" + strconv.Itoa(vless.Port)
		Node.LinkHost = vless.Server
		Node.LinkPort = strconv.Itoa(vless.Port)
	case u.Scheme == "hy" || u.Scheme == "hysteria":
		hy, err := protocol.DecodeHYURL(link)
		if err != nil {
			log.Println(err)
			return
		}
		if Node.Name == "" {
			Node.Name = hy.Name
		}
		Node.LinkName = hy.Name
		Node.LinkAddress = hy.Host + ":" + strconv.Itoa(hy.Port)
		Node.LinkHost = hy.Host
		Node.LinkPort = strconv.Itoa(hy.Port)
	case u.Scheme == "hy2" || u.Scheme == "hysteria2":
		hy2, err := protocol.DecodeHY2URL(link)
		if err != nil {
			log.Println(err)
			return
		}
		if Node.Name == "" {
			Node.Name = hy2.Name
		}
		Node.LinkName = hy2.Name
		Node.LinkAddress = hy2.Host + ":" + strconv.Itoa(hy2.Port)
		Node.LinkHost = hy2.Host
		Node.LinkPort = strconv.Itoa(hy2.Port)
	case u.Scheme == "tuic":
		tuic, err := protocol.DecodeTuicURL(link)
		if err != nil {
			log.Println(err)
			return
		}
		if Node.Name == "" {
			Node.Name = tuic.Name
		}
		Node.LinkName = tuic.Name
		Node.LinkAddress = tuic.Host + ":" + strconv.Itoa(tuic.Port)
		Node.LinkHost = tuic.Host
		Node.LinkPort = strconv.Itoa(tuic.Port)
	case u.Scheme == "socks5":
		socks5, err := protocol.DecodeSocks5URL(link)
		if err != nil {
			log.Println(err)
			return
		}
		if Node.Name == "" {
			Node.Name = socks5.Name
		}
		Node.LinkName = socks5.Name
		Node.LinkAddress = socks5.Server + ":" + strconv.Itoa(socks5.Port)
		Node.LinkHost = socks5.Server
		Node.LinkPort = strconv.Itoa(socks5.Port)
	}

	Node.Link = link
	Node.DialerProxyName = dialerProxyName
	Node.Group = group
	err = Node.Update()
	if err != nil {
		utils.FailWithMsg(c, "更新失败")
		return
	}
	utils.OkWithMsg(c, "更新成功")
}

// 获取节点列表
func NodeGet(c *gin.Context) {
	var Node models.Node

	// 解析过滤参数
	filter := models.NodeFilter{
		Search:    c.Query("search"),
		Group:     c.Query("group"),
		Source:    c.Query("source"),
		SortBy:    c.Query("sortBy"),
		SortOrder: c.Query("sortOrder"),
	}

	// 安全解析数值参数
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

	// 解析国家代码数组
	filter.Countries = c.QueryArray("countries[]")

	// 验证排序字段（白名单）
	if filter.SortBy != "" && filter.SortBy != "delay" && filter.SortBy != "speed" {
		filter.SortBy = "" // 无效排序字段，忽略
	}

	// 验证排序顺序
	if filter.SortOrder != "" && filter.SortOrder != "asc" && filter.SortOrder != "desc" {
		filter.SortOrder = "asc" // 默认升序
	}

	nodes, err := Node.ListWithFilters(filter)
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
	u, err := url.Parse(link)
	if err != nil {
		log.Println(err)
		return
	}
	switch {
	case u.Scheme == "ss":
		ss, err := protocol.DecodeSSURL(link)
		if err != nil {
			log.Println(err)
			return
		}
		if Node.Name == "" {
			Node.Name = ss.Name
		}
		Node.LinkName = ss.Name
		Node.LinkAddress = ss.Server + ":" + strconv.Itoa(ss.Port)
		Node.LinkHost = ss.Server
		Node.LinkPort = strconv.Itoa(ss.Port)
	case u.Scheme == "ssr":
		ssr, err := protocol.DecodeSSRURL(link)
		if err != nil {
			log.Println(err)
			return
		}
		if name == "" {
			Node.Name = ssr.Qurey.Remarks
		}
		Node.LinkName = ssr.Qurey.Remarks
		Node.LinkAddress = ssr.Server + ":" + strconv.Itoa(ssr.Port)
		Node.LinkHost = ssr.Server
		Node.LinkPort = strconv.Itoa(ssr.Port)
	case u.Scheme == "trojan":
		trojan, err := protocol.DecodeTrojanURL(link)
		if err != nil {
			log.Println(err)
			return
		}
		if name == "" {
			Node.Name = trojan.Name
		}
		Node.LinkName = trojan.Name
		Node.LinkAddress = trojan.Hostname + ":" + strconv.Itoa(trojan.Port)
		Node.LinkHost = trojan.Hostname
		Node.LinkPort = strconv.Itoa(trojan.Port)
	case u.Scheme == "vmess":
		vmess, err := protocol.DecodeVMESSURL(link)
		if err != nil {
			log.Println(err)
			return
		}
		if name == "" {
			Node.Name = vmess.Ps
		}
		Node.LinkName = vmess.Ps
		Node.LinkAddress = vmess.Add + ":" + vmess.Port.(string)
		Node.LinkHost = vmess.Host
		Node.LinkPort = vmess.Port.(string)
	case u.Scheme == "vless":
		vless, err := protocol.DecodeVLESSURL(link)
		if err != nil {
			log.Println(err)
			return
		}

		if name == "" {
			Node.Name = vless.Name
		}
		Node.LinkName = vless.Name
		Node.LinkAddress = vless.Server + ":" + strconv.Itoa(vless.Port)
		Node.LinkHost = vless.Server
		Node.LinkPort = strconv.Itoa(vless.Port)
	case u.Scheme == "hy" || u.Scheme == "hysteria":
		hy, err := protocol.DecodeHYURL(link)
		if err != nil {
			log.Println(err)
			return
		}

		if name == "" {
			Node.Name = hy.Name
		}
		Node.LinkName = hy.Name
		Node.LinkAddress = hy.Host + ":" + strconv.Itoa(hy.Port)
		Node.LinkHost = hy.Host
		Node.LinkPort = strconv.Itoa(hy.Port)
	case u.Scheme == "hy2" || u.Scheme == "hysteria2":
		hy2, err := protocol.DecodeHY2URL(link)
		if err != nil {
			log.Println(err)
			return
		}

		if name == "" {
			Node.Name = hy2.Name
		}
		Node.LinkName = hy2.Name
		Node.LinkAddress = hy2.Host + ":" + strconv.Itoa(hy2.Port)
		Node.LinkHost = hy2.Host
		Node.LinkPort = strconv.Itoa(hy2.Port)
	case u.Scheme == "tuic":
		tuic, err := protocol.DecodeTuicURL(link)
		if err != nil {
			log.Println(err)
			return
		}

		if name == "" {
			Node.Name = tuic.Name
		}
		Node.LinkName = tuic.Name
		Node.LinkAddress = tuic.Host + ":" + strconv.Itoa(tuic.Port)
		Node.LinkHost = tuic.Host
		Node.LinkPort = strconv.Itoa(tuic.Port)
	case u.Scheme == "socks5":
		socks5, err := protocol.DecodeSocks5URL(link)
		if err != nil {
			log.Println(err)
			return
		}

		if name == "" {
			Node.Name = socks5.Name
		}
		Node.LinkName = socks5.Name
		Node.LinkAddress = socks5.Server + ":" + strconv.Itoa(socks5.Port)
		Node.LinkHost = socks5.Server
		Node.LinkPort = strconv.Itoa(socks5.Port)
	}
	Node.Link = link
	Node.DialerProxyName = dialerProxyName
	Node.Group = group
	err = Node.Find()
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
	if err != nil {
		utils.FailWithMsg(c, "获取不到节点统计")
		return
	}

	total := len(nodes)
	available := 0
	for _, n := range nodes {
		if n.Speed > 0 && n.DelayTime > 0 {
			available++
		}
	}

	utils.OkDetailed(c, "取得节点统计", gin.H{
		"total":     total,
		"available": available,
	})
}

// NodeBatchDel 批量删除节点
func NodeBatchDel(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	if len(req.IDs) == 0 {
		utils.FailWithMsg(c, "请选择要删除的节点")
		return
	}
	err := models.BatchDel(req.IDs)
	if err != nil {
		utils.FailWithMsg(c, "批量删除失败")
		return
	}
	utils.OkWithMsg(c, "批量删除成功")
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

// GetSpeedTestConfig 获取测速配置
func GetSpeedTestConfig(c *gin.Context) {
	cron, _ := models.GetSetting("speed_test_cron")
	enabledStr, _ := models.GetSetting("speed_test_enabled")
	enabled := enabledStr == "true"
	mode, _ := models.GetSetting("speed_test_mode")
	if mode == "" {
		mode = "tcp"
	}
	url, _ := models.GetSetting("speed_test_url")
	timeoutStr, _ := models.GetSetting("speed_test_timeout")
	if timeoutStr == "" {
		timeoutStr = "5"
	}
	timeout, _ := strconv.Atoi(timeoutStr)
	groupsStr, _ := models.GetSetting("speed_test_groups")
	var groups []string
	if groupsStr != "" {
		groups = strings.Split(groupsStr, ",")
	} else {
		groups = []string{}
	}
	detectCountryStr, _ := models.GetSetting("speed_test_detect_country")
	detectCountry := detectCountryStr == "true"

	utils.OkDetailed(c, "获取成功", gin.H{
		"cron":           cron,
		"enabled":        enabled,
		"mode":           mode,
		"url":            url,
		"timeout":        timeout,
		"groups":         groups,
		"detect_country": detectCountry,
	})
}

// UpdateSpeedTestConfig 更新测速配置
func UpdateSpeedTestConfig(c *gin.Context) {
	var req struct {
		Cron          string      `json:"cron"`
		Enabled       bool        `json:"enabled"`
		Mode          string      `json:"mode"`
		Url           string      `json:"url"`
		Timeout       interface{} `json:"timeout"`
		Groups        []string    `json:"groups"`
		DetectCountry bool        `json:"detect_country"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailWithMsg(c, "参数错误")
		return
	}

	// 验证 Cron 表达式
	if req.Cron != "" {
		// 这里简单验证一下，或者依赖 scheduler 的验证
		// 暂时跳过严格验证，直接保存
	}

	err := models.SetSetting("speed_test_cron", req.Cron)
	if err != nil {
		utils.FailWithMsg(c, "保存Cron配置失败")
		return
	}
	err = models.SetSetting("speed_test_enabled", strconv.FormatBool(req.Enabled))
	if err != nil {
		utils.FailWithMsg(c, "保存启用状态失败")
		return
	}
	err = models.SetSetting("speed_test_mode", req.Mode)
	if err != nil {
		utils.FailWithMsg(c, "保存模式配置失败")
		return
	}
	err = models.SetSetting("speed_test_url", req.Url)
	if err != nil {
		utils.FailWithMsg(c, "保存URL配置失败")
		return
	}

	groupsStr := strings.Join(req.Groups, ",")
	err = models.SetSetting("speed_test_groups", groupsStr)
	if err != nil {
		utils.FailWithMsg(c, "保存分组配置失败")
		return
	}

	var timeoutStr string
	switch v := req.Timeout.(type) {
	case float64:
		timeoutStr = strconv.Itoa(int(v))
	case string:
		timeoutStr = v
	case int:
		timeoutStr = strconv.Itoa(v)
	default:
		timeoutStr = "5"
	}

	err = models.SetSetting("speed_test_timeout", timeoutStr)
	if err != nil {
		utils.FailWithMsg(c, "保存超时配置失败")
		return
	}

	err = models.SetSetting("speed_test_detect_country", strconv.FormatBool(req.DetectCountry))
	if err != nil {
		utils.FailWithMsg(c, "保存落地IP检测配置失败")
		return
	}

	// 更新定时任务
	scheduler := services.GetSchedulerManager()
	if req.Enabled {
		if req.Cron == "" {
			utils.FailWithMsg(c, "启用时Cron表达式不能为空")
			return
		}
		err = scheduler.StartNodeSpeedTestTask(req.Cron)
		if err != nil {
			utils.FailWithMsg(c, "启动定时任务失败: "+err.Error())
			return
		}
	} else {
		scheduler.StopNodeSpeedTestTask()
	}

	utils.OkWithMsg(c, "保存成功")
}

// RunSpeedTest 手动执行测速
func RunSpeedTest(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids"`
	}
	// 尝试绑定 JSON，如果失败（例如没有 body），则忽略错误继续执行全量测速
	_ = c.ShouldBindJSON(&req)

	if len(req.IDs) > 0 {
		go services.ExecuteSpecificNodeSpeedTestTask(req.IDs)
		utils.OkWithMsg(c, "指定节点测速任务已在后台启动")
	} else {
		go services.ExecuteNodeSpeedTestTask()
		utils.OkWithMsg(c, "测速任务已在后台启动")
	}
}

// FastestSpeedNode 获取最快速度节点
func FastestSpeedNode(c *gin.Context) {
	node := models.GetFastestSpeedNode()
	utils.OkDetailed(c, "获取最快速度节点成功", node)
}

// LowestDelayNode 获取最低延迟节点
func LowestDelayNode(c *gin.Context) {
	node := models.GetLowestDelayNode()
	utils.OkDetailed(c, "获取最低延迟节点成功", node)
}

// GetNodeCountries 获取所有节点的国家代码列表
func GetNodeCountries(c *gin.Context) {
	countries := models.GetAllCountries()
	utils.OkDetailed(c, "获取国家代码成功", countries)
}

// NodeCountryStats 获取按国家统计的节点数量
func NodeCountryStats(c *gin.Context) {
	stats := models.GetNodeCountryStats()
	utils.OkDetailed(c, "获取国家统计成功", stats)
}

// NodeProtocolStats 获取按协议统计的节点数量
func NodeProtocolStats(c *gin.Context) {
	stats := models.GetNodeProtocolStats()
	utils.OkDetailed(c, "获取协议统计成功", stats)
}
