package api

import (
	"errors"
	"strconv"
	"sublink/dto"
	"sublink/models"
	"sublink/node"
	"sublink/services"

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
		c.JSON(400, gin.H{"msg": "参数错误: " + err.Error()})
		return
	}
	if !validateFiveFieldCron(req.CronExpr) {
		c.JSON(400, gin.H{"msg": "CRON表达式错误 "})
		return
	}

	subS := models.SubScheduler{
		Name:     req.Name,
		URL:      req.URL,
		CronExpr: req.CronExpr,
		Enabled:  req.Enabled,
	}

	err = subS.Find()
	if err == nil {
		// 找到了，重复
		c.JSON(400, gin.H{"msg": "订阅已存在"})
		return
	}

	err = subS.Add()
	if err != nil {
		c.JSON(400, gin.H{"msg": "添加失败，可能重复或其他错误"})
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
		go node.LoadClashConfigFromURL(subS.ID, subS.URL, subS.Name)
	}

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "添加成功",
	})
}

func PullClashConfigFromURL(c *gin.Context) {
	var req dto.SubSchedulerAddRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{"msg": "参数错误: " + err.Error()})
		return
	}
	go node.LoadClashConfigFromURL(req.ID, req.URL, req.Name)
	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "任务启动成功",
	})
}

func SubSchedulerDel(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}
	var subS models.SubScheduler
	ssID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(500, gin.H{"msg": "删除失败"})
		return
	}
	subS.ID = ssID
	err = subS.Del()
	if err != nil {
		c.JSON(500, gin.H{"msg": "删除失败"})
		return
	}

	// 删除定时任务
	scheduler := services.GetSchedulerManager()
	scheduler.RemoveJob(ssID)

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "删除成功",
	})
}

func SubSchedulerGet(c *gin.Context) {
	subSs, err := new(models.SubScheduler).List()
	if err != nil {
		c.JSON(500, gin.H{
			"code": "50000",
			"msg":  "获取订阅任务列表失败: " + err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "获取成功",
		"data": subSs,
	})
}

func SubSchedulerUpdate(c *gin.Context) {
	var req dto.SubSchedulerAddRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{"msg": "参数错误: " + err.Error()})
		return
	}
	if req.ID == 0 {
		c.JSON(400, gin.H{"msg": "更新失败，ID 不能为空"})
		return
	}
	if !validateFiveFieldCron(req.CronExpr) {
		c.JSON(400, gin.H{"msg": "CRON表达式错误 "})
		return
	}
	subS := models.SubScheduler{
		Name: req.Name,
		URL:  req.URL,
	}
	err = subS.Find()
	if err == nil && subS.ID != req.ID {
		// 找到了，重复
		c.JSON(500, gin.H{"msg": "订阅已存在"})
		return
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {

		c.JSON(500, gin.H{"msg": "更新失败"})
		return
	}

	subS.Name = req.Name
	subS.URL = req.URL
	subS.ID = req.ID
	subS.CronExpr = req.CronExpr
	subS.Enabled = req.Enabled
	err = subS.Update()

	if err != nil {
		c.JSON(500, gin.H{"msg": "更新失败"})
		return
	}

	// 更新定时任务
	scheduler := services.GetSchedulerManager()
	_ = scheduler.UpdateJob(req.ID, req.CronExpr, req.Enabled, req.URL, req.Name)

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "更新成功",
	})

}
