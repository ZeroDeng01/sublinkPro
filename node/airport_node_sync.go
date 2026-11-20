package node

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/node/protocol"
	"sublink/utils"
	"sync"
)

// Serialize reconciliation and cache publication so concurrent airport updates
// use the latest committed identities, including for cross-airport deduplication.
var airportNodeSyncMu sync.Mutex

type airportNodeSyncResult struct {
	added, updated, skipped, deleted, successCount int
	changedIDs                                     []int
	incomplete                                     bool
}

type syncNodeNames map[string]map[string]bool

func (names syncNodeNames) add(hash, name string) {
	if names[hash] == nil {
		names[hash] = make(map[string]bool)
	}
	names[hash][name] = true
}

func syncAirportNodes(ctx context.Context, airport *models.Airport, subName string, proxies []protocol.Proxy, reporter TaskReporter, incomplete bool) (airportNodeSyncResult, error) {
	result := airportNodeSyncResult{incomplete: incomplete}
	// Use pre-numbering names to identify deliberately distinct names sharing a config.
	originalNames := make([]string, len(proxies))
	for i, proxy := range proxies {
		originalNames[i] = strings.TrimSpace(proxy.Name)
	}
	proxies = applyAirportIntraNodeUniquify(airport, proxies)
	requestedNames := make(syncNodeNames)
	candidates := make([]models.Node, 0, len(proxies))
	for index, proxy := range proxies {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		node, err := buildAirportSyncNode(airport, subName, proxy, index+1)
		if err != nil {
			utils.Warn("节点【%s】无法转换，保留可能受影响的旧节点: %v", proxy.Name, err)
			result.incomplete = true
			continue
		}
		requestedNames.add(node.ContentHash, originalNames[index])
		candidates = append(candidates, node)
	}
	if len(candidates) == 0 && result.incomplete {
		return result, fmt.Errorf("订阅获取或转换不完整，未找到可导入的节点")
	}

	airportNodeSyncMu.Lock()
	defer airportNodeSyncMu.Unlock()
	existing, err := models.ListAirportSyncNodes(ctx, airport.ID)
	if err != nil {
		return result, err
	}
	existingLinks, existingGroups := indexExistingSyncNodes(existing, candidates, requestedNames)
	var unmatchedLinks []models.Node
	for _, candidate := range candidates {
		if existingLinks[candidate.Link].ID == 0 {
			unmatchedLinks = append(unmatchedLinks, candidate)
		}
	}
	crossDedup, _ := models.GetSetting("cross_airport_dedup_enabled")
	foreignLinks, foreignHashes, err := models.FindAirportNodeConflicts(ctx, airport.ID, unmatchedLinks, crossDedup != "false")
	if err != nil {
		return result, err
	}

	// For an ordinary repeated config, prefer a copy whose link is already saved.
	firstByHash := make(map[string]int, len(requestedNames))
	for i, candidate := range candidates {
		first, exists := firstByHash[candidate.ContentHash]
		if !exists || (existingLinks[candidates[first].Link].ID == 0 && existingLinks[candidate.Link].ID != 0) {
			firstByHash[candidate.ContentHash] = i
		}
	}

	selected := make([]bool, len(candidates))
	selectedLinks := make(map[string]bool, len(candidates))
	reserved := make(map[int]bool)
	for i, candidate := range candidates {
		multi := len(requestedNames[candidate.ContentHash]) > 1 || len(existingGroups[candidate.ContentHash]) > 1
		if selectedLinks[candidate.Link] || (!multi && firstByHash[candidate.ContentHash] != i) {
			continue
		}
		selectedLinks[candidate.Link] = true
		selected[i] = true
		if old := existingLinks[candidate.Link]; old.ID != 0 {
			reserved[old.ID] = true
		}
	}

	change := models.AirportNodeSync{}
	matched := make(map[int]bool, len(existing))
	reporter.UpdateTotal(len(proxies))
	for i, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		status := "skipped"
		if !selected[i] {
			result.skipped++
		} else {
			multi := len(requestedNames[candidate.ContentHash]) > 1 || len(existingGroups[candidate.ContentHash]) > 1
			old := existingLinks[candidate.Link]
			if old.ID == 0 {
				old = findExistingSyncNode(existingGroups[candidate.ContentHash], candidate.LinkName, multi, reserved, matched)
			}
			switch {
			case old.ID != 0:
				matched[old.ID] = true
				result.skipped++
				if !foreignLinks[candidate.Link] {
					if update, changed := buildAirportSyncUpdate(airport, old, candidate); changed || old.ContentHash != candidate.ContentHash {
						change.Update = append(change.Update, update)
						status = "updated"
						if changed {
							result.changedIDs = append(result.changedIDs, old.ID)
						}
					}
				}
			case foreignLinks[candidate.Link] || (foreignHashes[candidate.ContentHash] && crossDedup != "false" && !multi):
				result.skipped++
			default:
				change.Add = append(change.Add, candidate)
				status = "added"
			}
		}
		reporter.ReportProgress(candidate.SourceSort, candidate.LinkName, map[string]any{
			"status": status, "added": len(change.Add), "skipped": result.skipped,
		})
	}
	if !result.incomplete {
		for _, old := range existing {
			if !matched[old.ID] {
				change.Delete = append(change.Delete, old.ID)
			}
		}
	}
	change.SuccessCount = len(matched) + len(change.Add)
	addedIDs, err := models.ApplyAirportNodeSync(ctx, airport.ID, change)
	if err != nil {
		return airportNodeSyncResult{}, err
	}
	result.added, result.updated, result.deleted = len(addedIDs), len(change.Update), len(change.Delete)
	result.successCount = change.SuccessCount
	result.changedIDs = append(addedIDs, result.changedIDs...)
	return result, nil
}

func buildAirportSyncNode(airport *models.Airport, subName string, proxy protocol.Proxy, sort int) (models.Node, error) {
	proxy.Name = strings.TrimSpace(proxy.Name)
	proxy.Server = utils.WrapIPv6Host(proxy.Server)
	hash := protocol.GenerateProxyContentHash(proxy)
	link := GenerateProxyLink(proxy)
	if hash == "" || link == "" {
		return models.Node{}, fmt.Errorf("无法生成节点链接或内容哈希")
	}
	extra, err := protocol.EncodeClashExtra(proxy)
	if err != nil {
		return models.Node{}, err
	}
	port := strconv.Itoa(int(proxy.Port))
	node := models.Node{
		Link: link, Name: proxy.Name, LinkName: proxy.Name, NameMode: models.NodeNameModeLink,
		ClashExtra: extra, LinkAddress: proxy.Server + ":" + port, LinkHost: proxy.Server, LinkPort: port,
		Source: subName, SourceID: airport.ID, SourceSort: sort, Group: airport.Group,
		Protocol: protocol.GetProtocolFromLink(link), ContentHash: hash,
	}
	if airport.AutoFillCountry {
		node.LinkCountry = models.ParseCountryFromNodeName(proxy.Name)
	}
	return node, nil
}

func indexExistingSyncNodes(existing, candidates []models.Node, requested syncNodeNames) (map[string]models.Node, map[string]map[string][]models.Node) {
	links := make(map[string]models.Node, len(existing))
	groups := make(map[string]map[string][]models.Node, len(existing))
	candidateHashes := make(map[string]string, len(candidates))
	for _, candidate := range candidates {
		if _, exists := candidateHashes[candidate.Link]; !exists {
			candidateHashes[candidate.Link] = candidate.ContentHash
		}
	}
	for _, old := range existing {
		links[old.Link] = old
		hash := old.ContentHash
		if current := candidateHashes[old.Link]; current != "" {
			hash = current
		} else if len(requested) > 0 && requested[hash] == nil {
			// Older rows may have hashed the YAML AST. Reuse a row for a renamed
			// node only when its saved configuration reconstructs the current hash.
			if proxy, err := protocol.LinkToProxy(protocol.Urls{Url: old.Link, ClashExtra: old.ClashExtra}, protocol.OutputConfig{}); err == nil {
				proxy.Server = utils.WrapIPv6Host(proxy.Server)
				if reconstructed := protocol.GenerateProxyContentHash(proxy); requested[reconstructed] != nil {
					hash = reconstructed
				}
			}
		}
		name := syncExistingNodeName(old)
		if groups[hash] == nil {
			groups[hash] = make(map[string][]models.Node)
		}
		groups[hash][name] = append(groups[hash][name], old)
	}
	return links, groups
}

func findExistingSyncNode(group map[string][]models.Node, name string, multi bool, reserved, matched map[int]bool) models.Node {
	nodes := group[name]
	if !multi {
		// An ordinary group has at most one historical name.
		for _, sameName := range group {
			nodes = sameName
			break
		}
	}
	for _, old := range nodes {
		if !reserved[old.ID] && !matched[old.ID] {
			return old
		}
	}
	return models.Node{}
}

func syncExistingNodeName(node models.Node) string {
	if name := strings.TrimSpace(node.LinkName); name != "" {
		return name
	}
	return strings.TrimSpace(node.Name)
}

func buildAirportSyncUpdate(airport *models.Airport, old, candidate models.Node) (models.NodeInfoUpdate, bool) {
	update := models.BuildNodeInfoUpdate(old, candidate.LinkName, candidate.Link, candidate.SourceSort)
	update.ClashExtra = candidate.ClashExtra
	update.ContentHash = candidate.ContentHash
	update.PreserveName = old.LinkName == candidate.LinkName
	if airport.BackfillExistingCountry && old.LinkCountry == "" {
		update.BackfillCountry = candidate.LinkCountry
		if !airport.AutoFillCountry {
			update.BackfillCountry = models.ParseCountryFromNodeName(candidate.LinkName)
		}
	}
	changed := old.LinkName != candidate.LinkName || old.Link != candidate.Link || old.ClashExtra != candidate.ClashExtra ||
		old.SourceSort != candidate.SourceSort || update.BackfillCountry != ""
	return update, changed
}
