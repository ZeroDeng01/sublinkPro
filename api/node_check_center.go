package api

import (
	"fmt"
	"strconv"
	"strings"
	"sublink/database"
	"sublink/models"
	"sublink/services/scheduler"
	"sublink/utils"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

// nodeCheckConfigMu 串行化方案与任务配置，防止引用检查与删除交错。
var nodeCheckConfigMu sync.Mutex

type nodeCheckJobView struct {
	models.NodeCheckJob
	Legacy      bool       `json:"legacy"`
	Running     bool       `json:"running"`
	NextRunTime *time.Time `json:"nextRunTime"`
}

func validateCheckProfileConfig(p models.NodeCheckProfile) error {
	if p.Mode != "tcp" && p.Mode != "mihomo" {
		return fmt.Errorf("invalid check mode")
	}
	if p.LatencyConcurrency < 0 || p.SpeedConcurrency < 0 {
		return fmt.Errorf("concurrency cannot be negative")
	}
	if p.Enabled && strings.TrimSpace(p.CronExpr) == "" {
		return fmt.Errorf("cron is required when scheduled checks are enabled")
	}
	if strings.TrimSpace(p.CronExpr) != "" {
		if _, err := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(p.CronExpr); err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
	}
	return nil
}

// legacyCheckJob 映射而非复制旧调度，避免升级后双跑；ID 负值与独立任务隔离。
func legacyCheckJob(p models.NodeCheckProfile) nodeCheckJobView {
	scope := "all"
	if p.Groups != "" || p.Tags != "" {
		scope = "filtered"
	}
	return nodeCheckJobView{NodeCheckJob: models.NodeCheckJob{ID: -p.ID, Name: p.Name, ProfileID: p.ID, Enabled: p.Enabled, Trigger: "scheduled", CronExpr: p.CronExpr, Scope: scope, AirportIDs: []int{}, Groups: p.GetGroups(), Tags: p.GetTags(), LastRunTime: p.LastRunTime}, Legacy: true, NextRunTime: p.NextRunTime}
}

func ListNodeCheckJobs(c *gin.Context) {
	jobs, err := models.ListNodeCheckJobs()
	if err != nil {
		utils.FailWithMsg(c, "Failed to fetch automatic tasks")
		return
	}
	views := make([]nodeCheckJobView, 0, len(jobs))
	for _, job := range jobs {
		views = append(views, nodeCheckJobView{NodeCheckJob: job, Running: scheduler.NodeCheckJobRunning(job.ID), NextRunTime: scheduler.GetSchedulerManager().NodeCheckJobNextRun(job.ID)})
	}
	profiles, _ := (&models.NodeCheckProfile{}).List()
	for _, p := range profiles {
		if p.Enabled || p.CronExpr != "" {
			views = append(views, legacyCheckJob(p))
		}
	}
	utils.OkDetailed(c, "Fetched successfully", gin.H{"items": views, "timezone": time.Now().Format("MST -07:00")})
}

func SaveNodeCheckJob(c *gin.Context) {
	nodeCheckConfigMu.Lock()
	defer nodeCheckConfigMu.Unlock()
	var job models.NodeCheckJob
	if err := c.ShouldBindJSON(&job); err != nil {
		utils.FailWithMsg(c, "Invalid automatic task parameters")
		return
	}
	id := 0
	if c.Param("id") != "" {
		var err error
		id, err = strconv.Atoi(c.Param("id"))
		if err != nil || id == 0 {
			utils.FailWithMsg(c, "Invalid task ID")
			return
		}
	}
	if id < 0 {
		saveLegacyCheckJob(c, id, job)
		return
	}
	var old models.NodeCheckJob
	if id > 0 {
		var err error
		old, err = models.GetNodeCheckJob(id)
		if err != nil {
			utils.FailWithMsg(c, "Task not found")
			return
		}
	}
	job.ID = id
	job.LastRunTime = nil
	job.LastError = ""
	job.CreatedAt = time.Time{}
	job.UpdatedAt = time.Time{}
	if err := scheduler.ValidateNodeCheckJob(&job); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	if err := job.Save(); err != nil {
		utils.FailWithMsg(c, "Failed to save automatic task")
		return
	}
	if err := scheduler.GetSchedulerManager().UpdateNodeCheckJobSchedule(job); err != nil {
		var rollbackErr error
		if id > 0 {
			rollbackErr = old.Save()
		} else {
			rollbackErr = job.Delete()
		}
		if rollbackErr != nil {
			utils.Error("Failed to roll back check task: %v", rollbackErr)
		}
		utils.FailWithMsg(c, "Failed to register automatic task: "+err.Error())
		return
	}
	utils.OkDetailed(c, "Saved successfully", job)
}

func saveLegacyCheckJob(c *gin.Context, id int, job models.NodeCheckJob) {
	p, err := models.GetNodeCheckProfileByID(-id)
	if err != nil {
		utils.FailWithMsg(c, "Legacy task not found")
		return
	}
	// 兼容任务始终绑定原方案；更换方案或事件编排请新建独立任务。
	if job.ProfileID != p.ID || job.Trigger != "scheduled" || (job.Scope != "all" && job.Scope != "filtered") || len(job.AirportIDs) > 0 || job.UpdateBeforeCheck {
		utils.FailWithMsg(c, "Legacy tasks only support the original schedule and group/tag scope; create a new standalone task to use advanced orchestration")
		return
	}
	// 暂停并清空 Cron 仅免除调度校验，不能绕过范围校验而把空 filtered 扩成全库。
	validation := job
	validation.Name = p.Name
	if !job.Enabled && strings.TrimSpace(job.CronExpr) == "" {
		validation.Trigger = "airport_update"
	}
	if err := validation.Validate(); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	old := *p
	p.Enabled = job.Enabled
	p.CronExpr = job.CronExpr
	p.SetGroups(job.Groups)
	p.SetTags(job.Tags)
	if err := p.Update(); err != nil {
		utils.FailWithMsg(c, "Failed to save legacy task")
		return
	}
	if err := scheduler.GetSchedulerManager().UpdateNodeCheckProfileJob(p.ID, p.CronExpr, p.Enabled); err != nil {
		if rollbackErr := old.Update(); rollbackErr != nil {
			utils.Error("Failed to roll back compatibility task: %v", rollbackErr)
		}
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkDetailed(c, "Saved successfully", legacyCheckJob(*p))
}

func DeleteNodeCheckJob(c *gin.Context) {
	nodeCheckConfigMu.Lock()
	defer nodeCheckConfigMu.Unlock()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id == 0 {
		utils.FailWithMsg(c, "Invalid task ID")
		return
	}
	if id < 0 {
		p, err := models.GetNodeCheckProfileByID(-id)
		if err != nil {
			utils.FailWithMsg(c, "Legacy task not found")
			return
		}
		p.Enabled = false
		p.CronExpr = ""
		if err := p.Update(); err != nil {
			utils.FailWithMsg(c, "Failed to delete legacy task")
			return
		}
		scheduler.GetSchedulerManager().RemoveNodeCheckProfileJob(p.ID)
	} else {
		job, err := models.GetNodeCheckJob(id)
		if err != nil {
			utils.FailWithMsg(c, "Task not found")
			return
		}
		if scheduler.NodeCheckJobRunning(id) {
			utils.FailWithMsg(c, "Task is running; delete it after it finishes")
			return
		}
		if err := job.Delete(); err != nil {
			utils.FailWithMsg(c, "Failed to delete task")
			return
		}
		job.Enabled = false
		if err := scheduler.GetSchedulerManager().UpdateNodeCheckJobSchedule(job); err != nil {
			utils.FailWithMsg(c, err.Error())
			return
		}
	}
	utils.OkWithMsg(c, "Deleted successfully")
}

func RunNodeCheckJob(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id == 0 {
		utils.FailWithMsg(c, "Invalid task ID")
		return
	}
	if id < 0 {
		if _, err := models.GetNodeCheckProfileByID(-id); err != nil {
			utils.FailWithMsg(c, "Legacy task not found")
			return
		}
		go scheduler.ExecuteNodeCheckWithProfile(-id, nil, models.TaskTriggerManual)
	} else if err := scheduler.StartNodeCheckJob(id, models.TaskTriggerManual, 0); err != nil {
		utils.FailWithMsg(c, err.Error())
		return
	}
	utils.OkWithMsg(c, "Node check task started")
}

// hasNodeCheckProfileDependents 包含暂停任务和机场的引用，禁止删除后遗留悬空方案。
func hasNodeCheckProfileDependents(id int) (bool, error) {
	jobs, err := models.CountNodeCheckJobsForProfile(id)
	if err != nil {
		return false, err
	}
	var airports int64
	err = database.DB.Model(&models.Airport{}).Where("update_after_detect_profile_id = ?", id).Count(&airports).Error
	return jobs > 0 || airports > 0, err
}
