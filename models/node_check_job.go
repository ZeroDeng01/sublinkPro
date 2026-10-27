package models

import (
	"fmt"
	"strings"
	"sublink/database"
	"time"

	"github.com/robfig/cron/v3"
)

// NodeCheckJob 将检测范围和自动触发从可复用方案中独立出来。
// 旧方案的调度仍由原入口执行，API 将其投影为负 ID 兼容任务，不复制注册。
type NodeCheckJob struct {
	ID                int        `gorm:"primaryKey" json:"id"`
	Name              string     `gorm:"size:191;not null" json:"name"`
	ProfileID         int        `gorm:"index;not null" json:"profileId"`
	Enabled           bool       `json:"enabled"`
	Trigger           string     `gorm:"size:32" json:"trigger"`
	CronExpr          string     `gorm:"size:191" json:"cronExpr"`
	Scope             string     `gorm:"size:32" json:"scope"`
	AirportIDs        []int      `gorm:"serializer:json;type:text" json:"airportIds"`
	Groups            []string   `gorm:"serializer:json;type:text" json:"groups"`
	Tags              []string   `gorm:"serializer:json;type:text" json:"tags"`
	SmartGroupID      int        `json:"smartGroupId"`
	UpdateBeforeCheck bool       `json:"updateBeforeCheck"`
	LastRunTime       *time.Time `json:"lastRunTime"`
	LastError         string     `gorm:"type:text" json:"lastError"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

func (j *NodeCheckJob) Validate() error {
	j.Name = strings.TrimSpace(j.Name)
	if j.Name == "" || len(j.Name) > 191 {
		return fmt.Errorf("任务名称不能为空且不能超过191字节")
	}
	if j.ProfileID <= 0 {
		return fmt.Errorf("请选择检测方案")
	}
	if j.Trigger != "scheduled" && j.Trigger != "airport_update" {
		return fmt.Errorf("无效的触发方式")
	}
	if j.Scope != "all" && j.Scope != "filtered" && j.Scope != "smart_group" {
		return fmt.Errorf("无效的检测范围")
	}
	if len(j.AirportIDs) > 128 || len(j.Groups) > 64 || len(j.Tags) > 64 {
		return fmt.Errorf("检测范围选项过多")
	}
	seen := map[int]bool{}
	ids := make([]int, 0, len(j.AirportIDs))
	for _, id := range j.AirportIDs {
		if id <= 0 {
			return fmt.Errorf("无效的机场ID")
		}
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	j.AirportIDs = ids
	if j.Scope == "all" && (len(j.AirportIDs) > 0 || len(j.Groups) > 0 || len(j.Tags) > 0 || j.SmartGroupID > 0) {
		return fmt.Errorf("全部节点范围不能包含隐藏筛选条件")
	}
	if j.Scope != "smart_group" && j.SmartGroupID != 0 {
		return fmt.Errorf("当前检测范围不能包含智能分组条件")
	}
	if j.Scope == "filtered" && len(ids) == 0 && len(j.Groups) == 0 && len(j.Tags) == 0 {
		return fmt.Errorf("筛选范围不能为空，请明确选择全部节点")
	}
	if j.Scope == "smart_group" && j.SmartGroupID <= 0 {
		return fmt.Errorf("请选择智能分组")
	}
	if j.UpdateBeforeCheck && (j.Trigger != "scheduled" || len(ids) == 0) {
		return fmt.Errorf("前置更新仅适用于定时任务，且必须指定机场")
	}
	j.CronExpr = strings.Join(strings.Fields(j.CronExpr), " ")
	if len(j.CronExpr) > 191 {
		return fmt.Errorf("cron 表达式过长")
	}
	if j.Trigger == "scheduled" {
		if len(strings.Fields(j.CronExpr)) != 5 {
			return fmt.Errorf("cron 必须为五字段表达式")
		}
		if _, err := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(j.CronExpr); err != nil {
			return fmt.Errorf("cron 表达式无效: %w", err)
		}
	}
	for _, list := range [][]string{j.Groups, j.Tags} {
		for _, v := range list {
			if strings.TrimSpace(v) == "" || len(v) > 191 {
				return fmt.Errorf("分组或标签名称无效")
			}
		}
	}
	return nil
}

func ListNodeCheckJobs() ([]NodeCheckJob, error) {
	jobs := make([]NodeCheckJob, 0)
	err := database.DB.Order("id ASC").Find(&jobs).Error
	return jobs, err
}

func GetNodeCheckJob(id int) (NodeCheckJob, error) {
	var job NodeCheckJob
	err := database.DB.First(&job, id).Error
	return job, err
}

// Save 只更新配置列，避免编辑任务覆盖正在执行的运行记录。
func (j *NodeCheckJob) Save() error {
	if j.ID == 0 {
		return database.DB.Create(j).Error
	}
	return database.DB.Model(j).Select("Name", "ProfileID", "Enabled", "Trigger", "CronExpr", "Scope", "AirportIDs", "Groups", "Tags", "SmartGroupID", "UpdateBeforeCheck").Updates(j).Error
}

func (j *NodeCheckJob) Delete() error { return database.DB.Delete(j).Error }

func (j *NodeCheckJob) RecordRun(err error) error {
	message := ""
	if err != nil {
		message = err.Error()
	}
	return database.DB.Model(j).Updates(map[string]any{"last_run_time": time.Now(), "last_error": message}).Error
}

func CountNodeCheckJobsForProfile(id int) (int64, error) {
	var count int64
	err := database.DB.Model(&NodeCheckJob{}).Where("profile_id = ?", id).Count(&count).Error
	return count, err
}
