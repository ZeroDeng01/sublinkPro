package models

import (
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"sublink/utils"
)

const serviceMarksStart = "$ServiceMarks("

// Match complete legacy calls first so their arguments stay opaque to the mark parser.
var nodeNameCallStart = regexp.MustCompile(`\$(TagGroup|Unlock)\([^)]+\)|\$ServiceMarks\([ \t\r\n]*\{`)

type nodeNameRule struct {
	legacy         string
	parts          []nodeNamePart
	duplicateIndex bool
}

func (rule nodeNameRule) empty() bool {
	return rule.legacy == "" && len(rule.parts) == 0
}

type nodeNamePart struct {
	text  string
	marks *serviceMarks
}

type serviceMarks struct {
	Prefix    string        `json:"prefix"`
	Separator string        `json:"separator"`
	Suffix    string        `json:"suffix"`
	Services  []serviceMark `json:"services"`
}

type serviceMark struct {
	Provider         string   `json:"provider"`
	Label            string   `json:"label"`
	AcceptedStatuses []string `json:"acceptedStatuses"`
	Region           bool     `json:"region"`
}

// InvalidServiceMarks identifies the block and, when applicable, the service to fix.
type InvalidServiceMarks struct {
	Block    int
	Reason   string
	Provider string
}

// ValidateNodeNameRule validates the structured mark blocks in a subscription template.
func ValidateNodeNameRule(rule string) *InvalidServiceMarks {
	_, invalid := compileNodeNameRule(rule)
	return invalid
}

func findServiceMarksStart(rule string) int {
	if !strings.Contains(rule, serviceMarksStart) {
		return -1
	}
	for offset := 0; ; {
		match := nodeNameCallStart.FindStringIndex(rule[offset:])
		if match == nil {
			return -1
		}
		if index := offset + match[0]; strings.HasPrefix(rule[index:], serviceMarksStart) {
			return index
		}
		offset += match[1]
	}
}

func compileNodeNameRule(rule string) (nodeNameRule, *InvalidServiceMarks) {
	start := findServiceMarksStart(rule)
	if start < 0 {
		return nodeNameRule{legacy: rule, duplicateIndex: strings.Contains(rule, duplicateIndexVariable)}, nil
	}

	var compiled nodeNameRule
	for block := 1; ; block++ {
		before := rule
		if start >= 0 {
			before = rule[:start]
		}
		if before != "" {
			compiled.parts = append(compiled.parts, nodeNamePart{text: before})
			compiled.duplicateIndex = compiled.duplicateIndex || strings.Contains(before, duplicateIndexVariable)
		}
		if start < 0 {
			return compiled, nil
		}

		remaining := strings.TrimLeft(rule[start+len(serviceMarksStart):], " \t\r\n")
		marks := serviceMarks{Separator: " · "}
		decoder := json.NewDecoder(strings.NewReader(remaining))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&marks); err != nil {
			return nodeNameRule{}, &InvalidServiceMarks{Block: block, Reason: "syntax"}
		}
		end := strings.TrimLeft(remaining[decoder.InputOffset():], " \t\r\n")
		if !strings.HasPrefix(end, ")") {
			return nodeNameRule{}, &InvalidServiceMarks{Block: block, Reason: "syntax"}
		}
		if invalid := marks.validate(block); invalid != nil {
			return nodeNameRule{}, invalid
		}
		compiled.parts = append(compiled.parts, nodeNamePart{marks: &marks})
		rule = end[1:]
		start = findServiceMarksStart(rule)
	}
}

func isInvalidMarkText(text string) bool {
	return strings.ContainsAny(text, ",=\u2028\u2029") || strings.ContainsFunc(text, unicode.IsControl)
}

func (marks *serviceMarks) validate(block int) *InvalidServiceMarks {
	if len(marks.Services) == 0 {
		return &InvalidServiceMarks{Block: block, Reason: "services"}
	}
	if isInvalidMarkText(marks.Prefix) || isInvalidMarkText(marks.Separator) || isInvalidMarkText(marks.Suffix) {
		return &InvalidServiceMarks{Block: block, Reason: "text"}
	}
	seen := make(map[string]bool, len(marks.Services))
	for i := range marks.Services {
		mark := &marks.Services[i]
		invalid := InvalidServiceMarks{Block: block, Provider: strings.TrimSpace(mark.Provider)}
		mark.Provider = NormalizeUnlockProvider(mark.Provider)
		if mark.Provider == "" {
			invalid.Reason = "provider"
			return &invalid
		}
		if seen[mark.Provider] {
			invalid.Reason = "duplicate"
			return &invalid
		}
		seen[mark.Provider] = true
		if isInvalidMarkText(mark.Label) {
			invalid.Reason = "text"
			return &invalid
		}
		mark.Label = strings.TrimSpace(mark.Label)
		if mark.Label == "" {
			invalid.Reason = "label"
			return &invalid
		}
		if len(mark.AcceptedStatuses) == 0 {
			invalid.Reason = "statuses"
			return &invalid
		}
		for index, status := range mark.AcceptedStatuses {
			normalized := NormalizeUnlockStatus(status)
			if normalized == "" {
				invalid.Reason = "statuses"
				return &invalid
			}
			mark.AcceptedStatuses[index] = normalized
		}
	}
	return nil
}

func (rule nodeNameRule) render(info utils.NodeInfo) string {
	if len(rule.parts) == 0 {
		return utils.RenameNode(rule.legacy, info)
	}
	replace := utils.NewNodeNameReplacer(info)
	var name strings.Builder
	for _, part := range rule.parts {
		if part.marks == nil {
			name.WriteString(replace(part.text))
			continue
		}
		part.marks.writeTo(&name, info.LookupUnlockResult)
	}
	result := strings.TrimSpace(name.String())
	if result == "" {
		return info.Name
	}
	return result
}

func (marks *serviceMarks) writeTo(name *strings.Builder, lookup utils.RenameUnlockLookup) {
	if lookup == nil {
		return
	}
	wroteMark := false
	for _, mark := range marks.Services {
		status, region, ok := lookup(mark.Provider)
		if !ok || !slices.Contains(mark.AcceptedStatuses, NormalizeUnlockStatus(status)) {
			continue
		}
		if wroteMark {
			name.WriteString(marks.Separator)
		} else {
			name.WriteString(marks.Prefix)
			wroteMark = true
		}
		name.WriteString(mark.Label)
		if region = strings.TrimSpace(region); mark.Region && region != "" && !isInvalidMarkText(region) {
			name.WriteByte('-')
			name.WriteString(region)
		}
	}
	if wroteMark {
		name.WriteString(marks.Suffix)
	}
}
