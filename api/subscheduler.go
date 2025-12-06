package api

import (
	"errors"
	"strconv"
	"sublink/dto"
	"sublink/models"
	"sublink/node"
	"sublink/services"
	"sublink/utils"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func validateFiveFieldCron(expr string) bool {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(expr)
	return err == nil
}

func SubSchedulerAdd(c *gin.Context) {
	var req dto.SubSchedulerAddRequest
	err := c.BindJSON(&req)
	if err != nil {
		utils.FailWithMsg(c, "参数错误: "+err.Error())
		return
	}
	if !validateFiveFieldCron(req.CronExpr) {
		utils.FailWithMsg(c, "CRON表达式错误 ")
		return
	}

	subS := models.SubScheduler{
		Name:              req.Name,
		URL:               req.URL,
		CronExpr:          req.CronExpr,
		Enabled:           req.Enabled,
		Group:             req.Group,
		DownloadWithProxy: req.DownloadWithProxy,
		ProxyLink:         req.ProxyLink,
	}

	err = subS.Find()
	if err == nil {
		// 找到了，重复
		utils.FailWithMsg(c, "订阅已存在")
		return
	}

	err = subS.Add()
	if err != nil {
		utils.FailWithMsg(c, "添加失败，可能重复或其他错误")
		return
	}

	// 添加定时任务
	if req.Enabled {
		scheduler := services.GetSchedulerManager()
		_ = scheduler.AddJob(subS.ID, req.CronExpr, func(id int, url string, subName string) {
			services.ExecuteSubscriptionTask(id, url, subName)
		}, subS.ID, req.URL, req.Name)
	}

	// 立即执行一次任务
	if req.Enabled {
		go node.LoadClashConfigFromURL(subS.ID, subS.URL, subS.Name, subS.DownloadWithProxy, subS.ProxyLink)
	}

	utils.OkWithMsg(c, "添加成功")
}

func PullClashConfigFromURL(c *gin.Context) {
	var req dto.SubSchedulerAddRequest
	err := c.BindJSON(&req)
	if err != nil {
		utils.FailWithMsg(c, "参数错误: "+err.Error())
		return
	}
	go node.LoadClashConfigFromURL(req.ID, req.URL, req.Name, req.DownloadWithProxy, req.ProxyLink)
	utils.OkWithMsg(c, "任务启动成功")
}

func SubSchedulerDel(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		utils.FailWithMsg(c, "参数错误")
		return
	}
	var subS models.SubScheduler
	ssID, err := strconv.Atoi(id)

	if err != nil {
		utils.FailWithMsg(c, "删除失败")
		return
	}
	subS.ID = ssID
	err = subS.Del()
	if err != nil {
		utils.FailWithMsg(c, "删除失败")
		return
	}

	// 删除定时任务
	scheduler := services.GetSchedulerManager()
	scheduler.RemoveJob(ssID)

	utils.OkWithMsg(c, "删除成功")
}

func SubSchedulerGet(c *gin.Context) {
	subSs, err := new(models.SubScheduler).List()
	if err != nil {
		utils.FailWithMsg(c, "获取订阅任务列表失败: "+err.Error())
		return
	}

	for i := range subSs {
		nodes, err := models.ListBySourceID(subSs[i].ID)
		if err == nil {
			subSs[i].NodeCount = len(nodes)
		}
	}

	utils.OkDetailed(c, "获取成功", subSs)
}

func SubSchedulerUpdate(c *gin.Context) {
	var req dto.SubSchedulerAddRequest
	err := c.BindJSON(&req)
	if err != nil {
		utils.FailWithMsg(c, "参数错误: "+err.Error())
		return
	}
	if req.ID == 0 {
		utils.FailWithMsg(c, "更新失败，ID 不能为空")
		return
	}
	if !validateFiveFieldCron(req.CronExpr) {
		utils.FailWithMsg(c, "CRON表达式错误 ")
		return
	}
	subS := models.SubScheduler{
		Name: req.Name,
		URL:  req.URL,
	}
	err = subS.Find()
	if err == nil && subS.ID != req.ID {
		// 找到了，重复
		utils.FailWithMsg(c, "订阅已存在")
		return
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {

		utils.FailWithMsg(c, "更新失败")
		return
	}

	subS.Name = req.Name
	subS.URL = req.URL
	subS.ID = req.ID
	subS.CronExpr = req.CronExpr
	subS.Enabled = req.Enabled
	subS.Group = req.Group
	subS.DownloadWithProxy = req.DownloadWithProxy
	subS.ProxyLink = req.ProxyLink
	err = subS.Update()

	if err != nil {
		utils.FailWithMsg(c, "更新失败")
		return
	}

	// 更新定时任务
	scheduler := services.GetSchedulerManager()
	_ = scheduler.UpdateJob(req.ID, req.CronExpr, req.Enabled, req.URL, req.Name)

	utils.OkWithMsg(c, "更新成功")

}
