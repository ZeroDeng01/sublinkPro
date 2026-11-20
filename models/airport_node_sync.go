package models

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"sublink/database"

	"gorm.io/gorm"
)

// ListAirportSyncNodes reads the authoritative rows used to plan a refresh.
func ListAirportSyncNodes(ctx context.Context, airportID int) ([]Node, error) {
	var nodes []Node
	err := database.DB.WithContext(ctx).Where("source_id = ?", airportID).Order("id").Find(&nodes).Error
	return nodes, err
}

// FindAirportNodeConflicts checks the global link constraint and, when enabled,
// cross-airport content deduplication against persisted nodes.
func FindAirportNodeConflicts(ctx context.Context, airportID int, candidates []Node, crossDedup bool) (links, hashes map[string]bool, err error) {
	links, hashes = make(map[string]bool), make(map[string]bool)
	for _, chunk := range chunkNodes(candidates, database.BatchSize) {
		linkHashes := make([]string, 0, len(chunk))
		contentHashes := make([]string, 0, len(chunk))
		for _, node := range chunk {
			linkHashes = append(linkHashes, hashNodeLink(node.Link))
			contentHashes = append(contentHashes, node.ContentHash)
		}
		query := database.DB.WithContext(ctx).Select("link", "content_hash").Where("(source_id <> ? OR source_id IS NULL)", airportID)
		if crossDedup {
			query = query.Where("(link_hash IN ? OR content_hash IN ?)", linkHashes, contentHashes)
		} else {
			query = query.Where("link_hash IN ?", linkHashes)
		}
		var conflicts []Node
		if err := query.Find(&conflicts).Error; err != nil {
			return nil, nil, err
		}
		for _, node := range conflicts {
			links[node.Link] = true
			hashes[node.ContentHash] = true
		}
	}
	return links, hashes, nil
}

// AirportNodeSync holds the changes planned from a single subscription snapshot.
type AirportNodeSync struct {
	Add          []Node
	Update       []NodeInfoUpdate
	Delete       []int
	SuccessCount int
}

// ApplyAirportNodeSync commits the node changes and refresh metadata together.
// Cache entries and created IDs come from the persisted rows after a successful commit.
func ApplyAirportNodeSync(ctx context.Context, airportID int, change AirportNodeSync) ([]int, error) {
	now := currentDBTime()
	var plans []nodeInfoUpdatePlan
	if len(change.Add)+len(change.Update) > 0 {
		names := newNodeInfoNameState()
		for _, update := range change.Update {
			plans = append(plans, prepareNodeInfoUpdatePlan(update, names, now))
		}
		for i := range change.Add {
			node := &change.Add[i]
			node.NormalizeNameModeDefaults()
			node.Name = names.uniqueNodeNameWithSource(node.Name, node.Source, 0)
			node.syncLinkHash()
		}
	}

	var persisted []Node
	var insertedIDs []int
	err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, chunk := range chunkNodes(change.Add, database.BatchSize) {
			// A link conflict here means the snapshot's write plan is stale. Roll it back.
			result := tx.Create(&chunk)
			if result.Error != nil {
				return fmt.Errorf("写入订阅节点失败: %w", result.Error)
			}
			if result.RowsAffected != int64(len(chunk)) {
				return fmt.Errorf("写入订阅节点数量不匹配: %d/%d", result.RowsAffected, len(chunk))
			}
		}
		for chunk := range slices.Chunk(plans, database.BatchSize) {
			if _, err := tryBatchUpdateNodeInfoWithCaseWhen(tx, chunk); err != nil {
				return fmt.Errorf("更新订阅节点失败: %w", err)
			}
		}
		for _, ids := range database.ChunkIntSlice(change.Delete, database.BatchSize) {
			count, err := deleteNodeRows(tx, ids)
			if err != nil {
				return fmt.Errorf("删除失效订阅节点失败: %w", err)
			}
			if count != int64(len(ids)) {
				return fmt.Errorf("失效订阅节点数量不匹配: %d/%d", count, len(ids))
			}
		}
		result := tx.Model(&Airport{}).Where("id = ?", airportID).UpdateColumns(map[string]any{
			"success_count": change.SuccessCount,
			"last_run_time": now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			var airport Airport
			if err := tx.Select("id").First(&airport, airportID).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("source_id = ?", airportID).Find(&persisted).Error; err != nil {
			return err
		}
		if len(change.Add) == 0 {
			return nil
		}
		byLink := make(map[string]int, len(persisted))
		for _, node := range persisted {
			byLink[node.Link] = node.ID
		}
		for _, node := range change.Add {
			id, ok := byLink[node.Link]
			if !ok {
				return fmt.Errorf("写入后未找到新增的订阅节点")
			}
			insertedIDs = append(insertedIDs, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	refreshAirportNodeCache(airportID, persisted)
	if cached, ok := airportCache.Get(airportID); ok {
		cached.SuccessCount = change.SuccessCount
		cached.LastRunTime = new(now)
		airportCache.Set(airportID, cached)
	}
	return insertedIDs, nil
}

func refreshAirportNodeCache(airportID int, persisted []Node) {
	byID := make(map[int]Node, len(persisted))
	for _, node := range persisted {
		byID[node.ID] = node
	}
	for _, node := range nodeCache.GetByIndex("sourceID", strconv.Itoa(airportID)) {
		if _, exists := byID[node.ID]; !exists {
			nodeCache.Delete(node.ID)
		}
	}
	for _, node := range persisted {
		if cached, exists := nodeCache.Get(node.ID); !exists || cached != node {
			nodeCache.Set(node.ID, node)
		}
	}
}
