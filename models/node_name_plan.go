package models

import (
	"regexp"
	"strings"
	"sublink/utils"
)

const duplicateIndexVariable = "$DuplicateIndex"

var nodeOutputLinkSeparator = regexp.MustCompile(`,\s*[a-zA-Z][a-zA-Z0-9+.-]*://`)

// NodeNamePlan records the final rendered names for subscription output.
type NodeNamePlan struct {
	NodeNames   []string
	SplitNames  [][]string
	NodeNameMap map[int]string
}

type nodeNameOccurrence struct {
	nodeIndex int
	linkIndex int
	protocol  string
	baseName  string
}

// BuildNodeNamePlan renders node names in output order and applies duplicate indexes only when requested.
func BuildNodeNamePlan(nodes []Node, nodeNamePreprocess string, nodeNameRule string, protocolFromLink func(string) string) NodeNamePlan {
	if protocolFromLink == nil {
		protocolFromLink = func(string) string { return "" }
	}

	plan := NodeNamePlan{
		NodeNames:   make([]string, len(nodes)),
		SplitNames:  make([][]string, len(nodes)),
		NodeNameMap: make(map[int]string, len(nodes)),
	}
	rule, _ := compileNodeNameRule(nodeNameRule)
	var occurrences []nodeNameOccurrence
	var nameTotals map[string]int
	var nodeInfos []utils.NodeInfo
	if rule.duplicateIndex {
		occurrences = make([]nodeNameOccurrence, 0, len(nodes))
		nameTotals = make(map[string]int, len(nodes))
		nodeInfos = make([]utils.NodeInfo, len(nodes))
	}

	for nodeIndex, node := range nodes {
		links := SplitNodeOutputLinks(node.Link)
		plan.SplitNames[nodeIndex] = make([]string, len(links))
		if rule.empty() {
			name := node.EffectiveName()
			for linkIndex := range links {
				plan.setName(nodeIndex, linkIndex, node.ID, name)
			}
			continue
		}

		processedLinkName := utils.PreprocessNodeName(nodeNamePreprocess, node.LinkName)
		info := BuildNodeRenameInfo(node, processedLinkName, "", nodeIndex+1, 0)
		if rule.duplicateIndex {
			nodeInfos[nodeIndex] = info
		}
		for linkIndex, link := range links {
			info.Protocol = protocolFromLink(link)
			name := rule.render(info)
			if !rule.duplicateIndex {
				plan.setName(nodeIndex, linkIndex, node.ID, name)
				continue
			}
			nameTotals[name]++
			occurrences = append(occurrences, nodeNameOccurrence{
				nodeIndex: nodeIndex,
				linkIndex: linkIndex,
				protocol:  info.Protocol,
				baseName:  name,
			})
		}
	}

	if !rule.duplicateIndex {
		return plan
	}
	nameIndexes := make(map[string]int, len(nameTotals))
	for _, occurrence := range occurrences {
		finalName := occurrence.baseName
		if nameTotals[occurrence.baseName] > 1 {
			duplicateIndex := nameIndexes[occurrence.baseName]
			nameIndexes[occurrence.baseName]++
			if duplicateIndex > 0 {
				info := nodeInfos[occurrence.nodeIndex]
				info.Protocol = occurrence.protocol
				info.DuplicateIndex = duplicateIndex
				finalName = rule.render(info)
			}
		}
		plan.setName(occurrence.nodeIndex, occurrence.linkIndex, nodes[occurrence.nodeIndex].ID, finalName)
	}

	return plan
}

// BuildNodeExportName renders one node name with the current rename variables.
func BuildNodeExportName(node Node, processedLinkName string, nodeNameRule string, protocolName string, index int, duplicateIndex int) string {
	rule, _ := compileNodeNameRule(nodeNameRule)
	if rule.empty() {
		return node.EffectiveName()
	}
	return rule.render(BuildNodeRenameInfo(node, processedLinkName, protocolName, index, duplicateIndex))
}

// NodeNameAt returns the planned primary output name for a node occurrence.
func (plan NodeNamePlan) NodeNameAt(index int, nodeID int) string {
	if index >= 0 && index < len(plan.NodeNames) && plan.NodeNames[index] != "" {
		return plan.NodeNames[index]
	}
	return plan.NodeNameMap[nodeID]
}

// SplitNamesAt returns the planned output names for a comma-split node occurrence.
func (plan NodeNamePlan) SplitNamesAt(index int) []string {
	if index < 0 || index >= len(plan.SplitNames) {
		return nil
	}
	return plan.SplitNames[index]
}

func (plan NodeNamePlan) setName(nodeIndex int, linkIndex int, nodeID int, name string) {
	if nodeIndex < 0 || nodeIndex >= len(plan.SplitNames) || linkIndex < 0 || linkIndex >= len(plan.SplitNames[nodeIndex]) {
		return
	}
	plan.SplitNames[nodeIndex][linkIndex] = name
	if linkIndex == 0 {
		plan.NodeNames[nodeIndex] = name
		if _, exists := plan.NodeNameMap[nodeID]; !exists {
			plan.NodeNameMap[nodeID] = name
		}
	}
}

// SplitNodeOutputLinks separates bundled links while preserving commas within each URI.
func SplitNodeOutputLinks(link string) []string {
	if !strings.Contains(link, ",") {
		return []string{link}
	}
	var links []string
	start := 0
	for _, separator := range nodeOutputLinkSeparator.FindAllStringIndex(link, -1) {
		links = append(links, link[start:separator[0]])
		start = separator[0] + 1
	}
	return append(links, link[start:])
}
