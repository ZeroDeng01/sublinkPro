package scheduler

import (
	"fmt"
	"slices"
	"strings"
	"sublink/models"
	"sublink/utils"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

var checkJobWorkers sync.WaitGroup

type nodeCheckJobRun struct{ pending map[int]bool }

var checkJobRuns = struct {
	sync.Mutex
	jobs map[int]*nodeCheckJobRun
}{jobs: make(map[int]*nodeCheckJobRun)}

// 可替换的同步边界用于编排测试；实际检测仍使用同一测速执行器。
var runCheckJobNodes = func(profile *models.NodeCheckProfile, nodes []models.Node, trigger models.TaskTrigger, name string) error {
	return RunSpeedTestWithConfig(nodes, trigger, name, SpeedTestConfigFromProfile(profile))
}
var updateCheckJobAirport func(int) error

func init() {
	updateCheckJobAirport = func(id int) error {
		airport, err := models.GetAirportByID(id)
		if err != nil {
			return err
		}
		// 前置更新不再派发更新后检测，避免同一流程重复检测或自触发循环。
		return executeSubscriptionTask(id, airport.URL, airport.Name, models.TaskTriggerScheduled, false)
	}

}

func ValidateNodeCheckJob(job *models.NodeCheckJob) error {
	if err := job.Validate(); err != nil {
		return err
	}
	if _, err := models.GetNodeCheckProfileByID(job.ProfileID); err != nil {
		return fmt.Errorf("检测方案不存在")
	}
	for _, id := range job.AirportIDs {
		if _, err := models.GetAirportByID(id); err != nil {
			return fmt.Errorf("机场 %d 不存在", id)
		}
	}
	return nil
}

// ResolveNodeCheckJobNodes 返回明确目标快照；空结果绝不回退到全库。
func ResolveNodeCheckJobNodes(job models.NodeCheckJob, airportEvents []int) ([]models.Node, error) {
	nodes, err := (&models.Node{}).ListWithFilters(models.NodeFilter{})
	if err != nil {
		return nil, err
	}
	result := make([]models.Node, 0)
	for _, n := range nodes {
		if len(job.AirportIDs) > 0 && !slices.Contains(job.AirportIDs, n.SourceID) {
			continue
		}
		if len(airportEvents) > 0 && !slices.Contains(airportEvents, n.SourceID) {
			continue
		}
		if len(job.Groups) > 0 && !slices.ContainsFunc(job.Groups, func(g string) bool { return strings.EqualFold(g, n.Group) }) {
			continue
		}
		result = append(result, n)
	}
	if len(job.Tags) > 0 {
		result = models.FilterNodesByTags(result, job.Tags)
	}
	return result, nil
}

func NodeCheckJobRunning(id int) bool {
	checkJobRuns.Lock()
	defer checkJobRuns.Unlock()
	return checkJobRuns.jobs[id] != nil
}

// StartNodeCheckJob 同步预留执行槽，防止连点/定时重叠；机场事件在运行期间合并补跑。
func StartNodeCheckJob(id int, trigger models.TaskTrigger, airportID int) error {
	job, err := models.GetNodeCheckJob(id)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}
	if models.IsDemoMode() {
		return fmt.Errorf("演示模式禁止执行检测")
	}
	if trigger == models.TaskTriggerScheduled && job.Trigger != "scheduled" {
		return fmt.Errorf("任务不接受定时触发")
	}
	if trigger != models.TaskTriggerManual && !job.Enabled {
		return fmt.Errorf("任务已暂停")
	}
	if trigger == models.TaskTriggerAirportUpdate && (job.Trigger != "airport_update" || airportID <= 0) {
		return fmt.Errorf("任务不接受此机场事件")
	}
	if err := ValidateNodeCheckJob(&job); err != nil {
		return err
	}
	checkJobRuns.Lock()
	if running := checkJobRuns.jobs[id]; running != nil {
		if trigger == models.TaskTriggerAirportUpdate {
			running.pending[airportID] = true
			checkJobRuns.Unlock()
			return nil
		}
		checkJobRuns.Unlock()
		return fmt.Errorf("任务正在运行，请勿重复启动")
	}
	owner := &nodeCheckJobRun{pending: make(map[int]bool)}
	checkJobRuns.jobs[id] = owner
	checkJobRuns.Unlock()
	events := []int(nil)
	if airportID > 0 {
		events = []int{airportID}
	}
	checkJobWorkers.Add(1)
	go runNodeCheckJobLoop(id, trigger, events, owner)
	return nil
}

func runNodeCheckJobLoop(id int, trigger models.TaskTrigger, events []int, owner *nodeCheckJobRun) {
	defer checkJobWorkers.Done()
	defer func() {
		checkJobRuns.Lock()
		if checkJobRuns.jobs[id] == owner {
			delete(checkJobRuns.jobs, id)
		}
		checkJobRuns.Unlock()
		if r := recover(); r != nil {
			utils.Error("自动检测任务 %d 异常: %v", id, r)
			job := models.NodeCheckJob{ID: id}
			if err := job.RecordRun(fmt.Errorf("检测异常: %v", r)); err != nil {
				utils.Warn("记录检测异常失败: %v", err)
			}
		}
	}()
	for {
		job, err := models.GetNodeCheckJob(id)
		if err != nil || (trigger != models.TaskTriggerManual && !job.Enabled) {
			return
		}
		if (trigger == models.TaskTriggerAirportUpdate && job.Trigger != "airport_update") || (trigger == models.TaskTriggerScheduled && job.Trigger != "scheduled") {
			return
		}
		if err = ValidateNodeCheckJob(&job); err == nil {
			err = executeNodeCheckJob(job, trigger, events)
		}
		if saveErr := job.RecordRun(err); saveErr != nil {
			utils.Warn("记录自动检测任务失败: %v", saveErr)
		}
		checkJobRuns.Lock()
		running := checkJobRuns.jobs[id]
		if len(running.pending) == 0 {
			delete(checkJobRuns.jobs, id)
			checkJobRuns.Unlock()
			return
		}
		events = make([]int, 0, len(running.pending))
		for airport := range running.pending {
			events = append(events, airport)
		}
		running.pending = make(map[int]bool)
		checkJobRuns.Unlock()
		trigger = models.TaskTriggerAirportUpdate
	}
}

func executeNodeCheckJob(job models.NodeCheckJob, trigger models.TaskTrigger, events []int) error {
	if job.UpdateBeforeCheck {
		for _, id := range job.AirportIDs {
			if err := updateCheckJobAirport(id); err != nil {
				return fmt.Errorf("前置更新机场 %d 失败: %w", id, err)
			}
		}
	}
	nodes, err := ResolveNodeCheckJobNodes(job, events)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return fmt.Errorf("本次范围内没有候选节点，已跳过检测")
	}
	profile, err := models.GetNodeCheckProfileByID(job.ProfileID)
	if err != nil {
		return err
	}
	// 节点结果持久化后由现有订阅筛选读取，不再次启动检测。
	return runCheckJobNodes(profile, nodes, trigger, job.Name)
}

func RunNodeCheckJobsAfterAirportUpdate(airportID int) {
	if models.IsDemoMode() {
		return
	}
	jobs, err := models.ListNodeCheckJobs()
	if err != nil {
		utils.Warn("加载机场联动检测任务失败: %v", err)
		return
	}
	for _, job := range jobs {
		if job.Enabled && job.Trigger == "airport_update" && (len(job.AirportIDs) == 0 || slices.Contains(job.AirportIDs, airportID)) {
			if err := StartNodeCheckJob(job.ID, models.TaskTriggerAirportUpdate, airportID); err != nil {
				utils.Warn("启动机场联动检测失败: %v", err)
			}
		}
	}
}

// UpdateNodeCheckJobSchedule 先验证新 entry，成功后再替换旧调度。
func (sm *SchedulerManager) UpdateNodeCheckJobSchedule(job models.NodeCheckJob) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	if sm.nodeCheckJobs == nil {
		sm.nodeCheckJobs = make(map[int]cron.EntryID)
	}
	old, exists := sm.nodeCheckJobs[job.ID]
	if !job.Enabled || job.Trigger != "scheduled" || models.IsDemoMode() {
		if exists {
			sm.cron.Remove(old)
			delete(sm.nodeCheckJobs, job.ID)
		}
		return nil
	}
	entry, err := sm.cron.AddFunc(job.CronExpr, func() {
		if err := StartNodeCheckJob(job.ID, models.TaskTriggerScheduled, 0); err != nil {
			utils.Warn("自动检测任务跳过: %v", err)
		}
	})
	if err != nil {
		return err
	}
	if exists {
		sm.cron.Remove(old)
	}
	sm.nodeCheckJobs[job.ID] = entry
	return nil
}

func (sm *SchedulerManager) NodeCheckJobNextRun(id int) *time.Time {
	sm.mutex.RLock()
	entry, ok := sm.nodeCheckJobs[id]
	sm.mutex.RUnlock()
	if !ok {
		return nil
	}
	next := sm.cron.Entry(entry).Next
	if next.IsZero() {
		job, err := models.GetNodeCheckJob(id)
		if err != nil {
			return nil
		}
		return sm.getNextRunTime(job.CronExpr)
	}
	return &next
}

func (sm *SchedulerManager) LoadNodeCheckJobs() error {
	jobs, err := models.ListNodeCheckJobs()
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if err := ValidateNodeCheckJob(&job); err != nil {
			utils.Warn("自动检测任务 [%s] 无效: %v", job.Name, err)
			continue
		}
		if err := sm.UpdateNodeCheckJobSchedule(job); err != nil {
			return err
		}
	}
	return nil
}
