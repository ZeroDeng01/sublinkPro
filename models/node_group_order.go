package models

import "sort"

const defaultGroupAirportSortWeight = 999999

// SortNodesByAirport 对同一分组下的节点做“按机场聚合”的稳定排序，避免不同机场节点穿插。
//
// 规则：
// 1) 机场优先级：按 airportSortMap（airportID -> sortWeight，值越小越靠前）；未配置的机场使用默认权重 defaultGroupAirportSortWeight。
// 2) 同权重下：按首次出现顺序（基于输入切片顺序）保持稳定，尽量不改变既有机场块的相对位置。
// 3) 单机场内部：优先按 SourceSort（上游订阅顺序，从 1 开始）排序；SourceSort=0 视为未初始化，保持原顺序并排在最后。
//
// 说明：
// - 输入 nodes 建议已是稳定顺序（例如按 nodes.id ASC），以便“首次出现顺序”可复现。
// - 该函数不会做去重；去重由上层逻辑（订阅展开时的 nodeMap / 去重规则）负责。
func SortNodesByAirport(nodes []Node, airportSortMap map[int]int) []Node {
	if len(nodes) <= 1 {
		return nodes
	}

	// 每个机场的首次出现序号（用于同权重下保持稳定）
	firstRankBySource := make(map[int]int)
	nextRank := 0
	for _, n := range nodes {
		if _, ok := firstRankBySource[n.SourceID]; ok {
			continue
		}
		firstRankBySource[n.SourceID] = nextRank
		nextRank++
	}

	sort.SliceStable(nodes, func(i, j int) bool {
		si := nodes[i].SourceID
		sj := nodes[j].SourceID

		// 先按机场排序权重
		wi, okI := airportSortMap[si]
		wj, okJ := airportSortMap[sj]
		if !okI {
			wi = defaultGroupAirportSortWeight
		}
		if !okJ {
			wj = defaultGroupAirportSortWeight
		}
		if wi != wj {
			return wi < wj
		}

		// 同权重下按首次出现顺序
		ri := firstRankBySource[si]
		rj := firstRankBySource[sj]
		if ri != rj {
			return ri < rj
		}

		// 同机场内部按 SourceSort（上游顺序）排序；未初始化的（0）排最后
		is := nodes[i].SourceSort
		js := nodes[j].SourceSort
		if is == 0 && js == 0 {
			return false
		}
		if is == 0 {
			return false
		}
		if js == 0 {
			return true
		}
		if is != js {
			return is < js
		}

		return nodes[i].ID < nodes[j].ID
	})

	return nodes
}
