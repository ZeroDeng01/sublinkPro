package scheduler

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"sublink/database"
	"sublink/models"
	"sublink/utils"
)

var (
	ErrSmartGroupCheckRunning = errors.New("该智能分组正在检测，请等待当前任务完成")
	ErrSmartGroupNoCandidates = errors.New("当前条件没有匹配的候选节点")
	smartGroupCheckMu         sync.Mutex
	smartGroupFollowups       sync.WaitGroup                   // tracks queued workers through the release/re-reserve gap
	smartGroupChecks          = make(map[int]map[int]struct{}) // group -> airport update IDs queued during an active check
	executeSmartGroupNodes    = ExecuteNodeCheckWithProfileNamed
)

func ValidateSmartGroupCheckProfile(group models.SmartGroup, profileID int) error {
	profile, err := models.GetNodeCheckProfileByID(profileID)
	if err != nil {
		return fmt.Errorf("节点检测策略不存在")
	}
	if group.MinSpeed > 0 && profile.Mode == "tcp" {
		return fmt.Errorf("最低速度大于零时必须选择包含下载测速的 mihomo 策略")
	}
	return nil
}

func SmartGroupCheckRunning(id int) bool {
	smartGroupCheckMu.Lock()
	defer smartGroupCheckMu.Unlock()
	_, exists := smartGroupChecks[id]
	return exists
}

func reserveSmartGroupCheck(id int, updateAirports []int) bool {
	smartGroupCheckMu.Lock()
	defer smartGroupCheckMu.Unlock()
	if pending, exists := smartGroupChecks[id]; exists {
		for _, airportID := range updateAirports {
			pending[airportID] = struct{}{}
		}
		return false
	}
	smartGroupChecks[id] = make(map[int]struct{})
	return true
}

func finishSmartGroupChecks(groups []models.SmartGroup) {
	smartGroupCheckMu.Lock()
	pendingGroups := make(map[int]map[int]struct{})
	airports := make(map[int]struct{})
	for _, group := range groups {
		if pending := smartGroupChecks[group.ID]; len(pending) > 0 {
			pendingGroups[group.ID] = pending
			for airportID := range pending {
				airports[airportID] = struct{}{}
			}
		}
		delete(smartGroupChecks, group.ID)
	}
	if len(pendingGroups) > 0 {
		smartGroupFollowups.Add(1)
	}
	smartGroupCheckMu.Unlock()
	if len(pendingGroups) > 0 {
		ids := make([]int, 0, len(airports))
		for airportID := range airports {
			ids = append(ids, airportID)
		}
		// Release a whole batch together and preserve same-profile deduplication on
		// follow-up checks. Re-read the latest candidates and configuration.
		go func() {
			defer smartGroupFollowups.Done()
			runSmartGroupAirportChecks(ids, pendingGroups)
		}()
	}
}

// StartSmartGroupCheck reserves before returning, so manual, cron and update-triggered checks cannot overlap per group.
func StartSmartGroupCheck(groupID, profileID, nodeID int, trigger models.TaskTrigger) (int, error) {
	group, err := models.GetSmartGroup(groupID)
	if err != nil {
		return 0, err
	}
	if err := ValidateSmartGroupCheckProfile(group, profileID); err != nil {
		return 0, err
	}
	ids, err := group.CandidateIDs()
	if err != nil {
		return 0, err
	}
	if nodeID != 0 {
		found := false
		for _, id := range ids {
			if id == nodeID {
				found = true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("节点不属于当前智能分组候选")
		}
		ids = []int{nodeID}
	}
	if len(ids) == 0 {
		return 0, ErrSmartGroupNoCandidates
	}
	if !reserveSmartGroupCheck(groupID, nil) {
		return 0, ErrSmartGroupCheckRunning
	}
	go runSmartGroupBatch(profileID, []models.SmartGroup{group}, ids, trigger)
	return len(ids), nil
}

func RunScheduledSmartGroupCheck(id int) {
	if models.IsDemoMode() {
		return
	}
	group, err := models.GetSmartGroup(id)
	if err != nil || !group.ScheduleEnabled {
		return
	}
	if _, err := StartSmartGroupCheck(id, group.CheckProfileID, 0, models.TaskTriggerScheduled); err != nil {
		utils.Warn("智能分组 [%s] 定时检测跳过: %v", group.Name, err)
	}
}

func runSmartGroupBatch(profileID int, groups []models.SmartGroup, ids []int, trigger models.TaskTrigger) {
	defer func() {
		finishSmartGroupChecks(groups)
		if r := recover(); r != nil {
			utils.Error("智能分组检测异常: %v", r)
		}
	}()
	// Empty explicit IDs mean profile scope in the shared runner; never call it with an empty candidate set.
	if len(ids) == 0 {
		return
	}
	now := time.Now()
	names := make([]string, 0, len(groups))
	for _, group := range groups {
		names = append(names, group.Name)
		if err := database.DB.Model(&models.SmartGroup{}).Where("id = ?", group.ID).Update("last_check_at", now).Error; err != nil {
			utils.Warn("记录智能分组检测触发时间失败: %v", err)
		}
	}
	name := []rune("智能分组: " + strings.Join(names, ", "))
	if len(name) > 40 {
		name = name[:40]
	} // fit Task.Name on all supported databases
	executeSmartGroupNodes(profileID, ids, trigger, string(name))
}

// RunSmartGroupChecksAfterAirportUpdate checks matching nodes from this airport, not the whole group/library.
func RunSmartGroupChecksAfterAirportUpdate(airportID int) {
	runSmartGroupAirportChecks([]int{airportID}, nil)
}

func runSmartGroupAirportChecks(airportIDs []int, pendingGroups map[int]map[int]struct{}) {
	if models.IsDemoMode() {
		return
	}
	groups, err := models.ListSmartGroups()
	if err != nil {
		utils.Warn("加载更新后智能分组失败: %v", err)
		return
	}
	airportNodes := make(map[int]int)
	for _, airportID := range airportIDs {
		nodes, err := models.ListBySourceID(airportID)
		if err != nil {
			utils.Warn("加载机场节点失败，跳过智能分组检测: %v", err)
			return
		}
		for _, node := range nodes {
			airportNodes[node.ID] = airportID
		}
	}
	if len(airportNodes) == 0 {
		return
	}
	type batch struct {
		groups []models.SmartGroup
		nodes  map[int]struct{}
	}
	batches := make(map[int]*batch)
	for _, group := range groups {
		if !group.CheckAfterAirportUpdate || (pendingGroups != nil && pendingGroups[group.ID] == nil) {
			continue
		}
		if err := ValidateSmartGroupCheckProfile(group, group.CheckProfileID); err != nil {
			utils.Warn("智能分组 [%s] 更新后检测跳过: %v", group.Name, err)
			continue
		}
		ids, err := group.CandidateIDs()
		if err != nil {
			utils.Warn("获取智能分组候选失败: %v", err)
			continue
		}
		targets := make([]int, 0)
		for _, id := range ids {
			if airportID, exists := airportNodes[id]; exists {
				if pendingGroups != nil {
					if _, queued := pendingGroups[group.ID][airportID]; !queued {
						continue
					}
				}
				targets = append(targets, id)
			}
		}
		if len(targets) == 0 {
			continue
		}
		queuedAirports := airportIDs
		if pendingGroups != nil {
			queuedAirports = make([]int, 0, len(pendingGroups[group.ID]))
			for airportID := range pendingGroups[group.ID] {
				queuedAirports = append(queuedAirports, airportID)
			}
		}
		if !reserveSmartGroupCheck(group.ID, queuedAirports) {
			continue
		}
		item := batches[group.CheckProfileID]
		if item == nil {
			item = &batch{nodes: make(map[int]struct{})}
			batches[group.CheckProfileID] = item
		}
		item.groups = append(item.groups, group)
		for _, id := range targets {
			item.nodes[id] = struct{}{}
		}
	}
	profileIDs := make([]int, 0, len(batches))
	for id := range batches {
		profileIDs = append(profileIDs, id)
	}
	sort.Ints(profileIDs)
	for _, profileID := range profileIDs {
		item := batches[profileID]
		ids := make([]int, 0, len(item.nodes))
		for id := range item.nodes {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		runSmartGroupBatch(profileID, item.groups, ids, models.TaskTriggerAirportUpdate)
	}
}
