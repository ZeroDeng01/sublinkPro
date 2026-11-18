package protocol

import (
	"bytes"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// clashTextEdit 表示模板文本里的一次行级替换：把 [start, end) 行换成 lines。
type clashTextEdit struct {
	start int
	end   int
	lines []string
}

// mergeClashTemplateText 在模板原文上重写顶层 proxies 与代理组的 proxies。
// 未改动的字节（注释、空行、缩进、流式写法、键顺序）全部原样保留，因此生成的
// 文件与模板逐字一致，只有节点列表和代理组引用会被替换。
// 返回 false 表示模板结构不受支持，调用方应退回节点树重编码。
func mergeClashTemplateText(data []byte, root *yaml.Node, proxyNodes []*yaml.Node, proxyNames []string, customGroupNodes []*yaml.Node) ([]byte, bool) {
	// 统一按 LF 处理行，最后再还原模板原本的换行符（Windows 模板常见 CRLF）
	text := string(data)
	eol := "\n"
	switch {
	case strings.Contains(text, "\r\n"):
		eol = "\r\n"
		text = strings.ReplaceAll(text, "\r\n", "\n")
	case strings.Contains(text, "\r"):
		// 只有 CR 换行的模板无法按行对齐，交给节点树重编码
		return nil, false
	}
	lines := strings.Split(text, "\n")

	// 流式根映射（整份模板写在一行里）无法逐行替换
	if root.Style&yaml.FlowStyle != 0 {
		return nil, false
	}

	proxiesKey, proxiesValue, ok := clashMappingKeyValue(root, "proxies")
	if !ok {
		return nil, false
	}
	defaultIndent := clashTemplateItemIndent(lines, root)

	proxyEdit, ok := clashSequenceAppendEdit(lines, proxiesKey, proxiesValue, proxyNodes, defaultIndent)
	if !ok {
		return nil, false
	}
	edits := []clashTextEdit{proxyEdit}

	groupsKey, groupsValue, hasGroups := clashMappingKeyValue(root, "proxy-groups")
	if hasGroups {
		groupsNode := resolveYAMLAlias(groupsValue)
		if groupsNode == nil || groupsNode.Kind != yaml.SequenceNode || groupsNode.Style&yaml.FlowStyle != 0 {
			return nil, false
		}
		for _, group := range groupsNode.Content {
			edit, changed, ok := clashProxyGroupEdit(lines, group, proxyNames)
			if !ok {
				return nil, false
			}
			if changed {
				edits = append(edits, edit)
			}
		}
	}

	if len(customGroupNodes) > 0 {
		if !hasGroups {
			return nil, false
		}
		edit, ok := clashSequenceAppendEdit(lines, groupsKey, groupsValue, customGroupNodes, defaultIndent)
		if !ok {
			return nil, false
		}
		edits = append(edits, edit)
	}

	merged, ok := applyClashTextEdits(lines, edits)
	if !ok {
		return nil, false
	}
	joined := strings.Join(merged, "\n")
	if eol != "\n" {
		joined = strings.ReplaceAll(joined, "\n", eol)
	}
	return []byte(joined), true
}

// clashSequenceAppendEdit 在模板的序列字段后面追加元素。
// 空值（~ / null / []）会就地展开成块序列，已有块序列则在最后一项后追加。
func clashSequenceAppendEdit(lines []string, keyNode, valueNode *yaml.Node, itemNodes []*yaml.Node, defaultIndent int) (clashTextEdit, bool) {
	keyLine := keyNode.Line - 1
	if keyLine < 0 || keyLine >= len(lines) {
		return clashTextEdit{}, false
	}
	if len(itemNodes) == 0 {
		return clashTextEdit{start: keyLine, end: keyLine}, true
	}
	if valueNode.Kind == yaml.AliasNode {
		// 别名指向别处定义的列表，逐行替换会改到错误的段落
		return clashTextEdit{}, false
	}
	resolved := resolveYAMLAlias(valueNode)
	if resolved == nil {
		return clashTextEdit{}, false
	}

	switch {
	case resolved.Kind == yaml.ScalarNode && resolved.Tag == "!!null":
		return clashInlineSequenceEdit(lines, keyLine, keyNode, resolved, itemNodes, defaultIndent)
	case resolved.Kind == yaml.SequenceNode && resolved.Style&yaml.FlowStyle != 0:
		if len(resolved.Content) != 0 {
			return clashTextEdit{}, false
		}
		return clashInlineSequenceEdit(lines, keyLine, keyNode, resolved, itemNodes, defaultIndent)
	case resolved.Kind == yaml.SequenceNode:
		if len(resolved.Content) == 0 {
			return clashInlineSequenceEdit(lines, keyLine, keyNode, resolved, itemNodes, defaultIndent)
		}
		itemIndent := clashBlockItemIndent(lines, resolved, defaultIndent)
		itemLines, ok := clashBlockItemLines(itemNodes, itemIndent)
		if !ok {
			return clashTextEdit{}, false
		}
		end := clashBlockSequenceEnd(lines, resolved)
		return clashTextEdit{start: end, end: end, lines: itemLines}, true
	}

	return clashTextEdit{}, false
}

// clashInlineSequenceEdit 把键行上的内联值（~ / null / []）改成块序列。
func clashInlineSequenceEdit(lines []string, keyLine int, keyNode, valueNode *yaml.Node, itemNodes []*yaml.Node, indent int) (clashTextEdit, bool) {
	itemLines, ok := clashBlockItemLines(itemNodes, indent)
	if !ok {
		return clashTextEdit{}, false
	}
	rewritten, ok := clashRewriteKeyLineValue(lines[keyLine], keyNode, "")
	if !ok {
		return clashTextEdit{}, false
	}
	end, ok := clashInlineValueEnd(lines, keyNode, valueNode, keyLine)
	if !ok {
		return clashTextEdit{}, false
	}
	replacement := append([]string{rewritten}, itemLines...)
	return clashTextEdit{start: keyLine, end: end, lines: replacement}, true
}

// clashInlineValueEnd 返回键行内联值需要替换到的行下标（右端不含）。
// 值写在下一行（如 proxies: 换行 ~）时一并替换，无法安全判断时 ok=false。
func clashInlineValueEnd(lines []string, keyNode, valueNode *yaml.Node, keyLine int) (int, bool) {
	if valueNode == nil || valueNode.Line != keyNode.Line+1 || keyLine+1 >= len(lines) {
		return keyLine + 1, true
	}
	trimmed := strings.TrimSpace(lines[keyLine+1])
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return keyLine + 1, true
	}
	switch clashInlineValueToken(trimmed) {
	case "~", "null", "Null", "NULL", "[]":
		return keyLine + 2, true
	}
	if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") {
		// 值写成了下一行的流式集合，逐行替换不安全
		return 0, false
	}
	return keyLine + 1, true
}

// clashInlineValueToken 取出一行里的第一个标量记号。
func clashInlineValueToken(line string) string {
	if index := strings.IndexAny(line, " \t"); index >= 0 {
		return line[:index]
	}
	return line
}

// clashProxyGroupEdit 计算单个代理组在模板文本里的改动。
// 返回 changed=false 表示该组保持模板原样（relay、include-all、已有固定列表等）。
func clashProxyGroupEdit(lines []string, group *yaml.Node, proxyNames []string) (clashTextEdit, bool, bool) {
	group = resolveYAMLAlias(group)
	if group == nil || group.Kind != yaml.MappingNode || group.Style&yaml.FlowStyle != 0 {
		return clashTextEdit{}, false, false
	}
	var fields map[string]any
	if err := group.Decode(&fields); err != nil {
		return clashTextEdit{}, false, false
	}
	if groupType, _ := fields["type"].(string); strings.TrimSpace(groupType) == "relay" {
		return clashTextEdit{}, false, true
	}
	// 自动匹配组不能补全成员，但显式写下的 __ALL_PROXIES__ 仍要就地展开。
	preserve := shouldPreserveProxyGroup(fields)

	keyNode, valueNode, found := clashMappingKeyValue(group, "proxies")
	if !found {
		if preserve {
			return clashTextEdit{}, false, true
		}
		groupIndent := group.Column - 1
		itemLines, ok := clashNameItemLines(clashFillNames(proxyNames), groupIndent+2)
		if !ok {
			return clashTextEdit{}, false, false
		}
		end := clashBlockMappingEnd(lines, group)
		replacement := append([]string{strings.Repeat(" ", groupIndent) + "proxies:"}, itemLines...)
		return clashTextEdit{start: end, end: end, lines: replacement}, true, true
	}

	return clashProxyNamesEdit(lines, keyNode, valueNode, proxyNames, !preserve)
}

// clashFillNames 返回填充空代理组时使用的节点名列表，空订阅回退到 DIRECT。
func clashFillNames(names []string) []string {
	if len(names) == 0 {
		return []string{"DIRECT"}
	}
	return names
}

// clashProxyNamesEdit 用节点名替换单个代理组的 proxies 内容。
// fillEmpty=false 时只展开 __ALL_PROXIES__，缺失或空列表保持模板原样（自动匹配组）。
func clashProxyNamesEdit(lines []string, keyNode, valueNode *yaml.Node, names []string, fillEmpty bool) (clashTextEdit, bool, bool) {
	keyLine := keyNode.Line - 1
	if keyLine < 0 || keyLine >= len(lines) {
		return clashTextEdit{}, false, false
	}
	if valueNode.Kind == yaml.AliasNode {
		// 别名指向别处定义的列表，逐行替换会改到错误的段落
		return clashTextEdit{}, false, false
	}
	resolved := resolveYAMLAlias(valueNode)
	if resolved == nil {
		return clashTextEdit{}, false, false
	}
	itemIndent := clashLineIndent(lines[keyLine]) + 2

	if resolved.Kind == yaml.ScalarNode {
		if resolved.Tag != "!!null" {
			return clashTextEdit{}, false, false
		}
		if !fillEmpty {
			return clashTextEdit{}, false, true
		}
		itemLines, ok := clashNameItemLines(clashFillNames(names), itemIndent)
		if !ok {
			return clashTextEdit{}, false, false
		}
		rewritten, ok := clashRewriteKeyLineValue(lines[keyLine], keyNode, "")
		if !ok {
			return clashTextEdit{}, false, false
		}
		end, ok := clashInlineValueEnd(lines, keyNode, resolved, keyLine)
		if !ok {
			return clashTextEdit{}, false, false
		}
		replacement := append([]string{rewritten}, itemLines...)
		return clashTextEdit{start: keyLine, end: end, lines: replacement}, true, true
	}

	if resolved.Kind != yaml.SequenceNode {
		return clashTextEdit{}, false, false
	}

	if resolved.Style&yaml.FlowStyle != 0 {
		if len(resolved.Content) == 0 {
			if !fillEmpty {
				return clashTextEdit{}, false, true
			}
			itemLines, ok := clashNameItemLines(clashFillNames(names), itemIndent)
			if !ok {
				return clashTextEdit{}, false, false
			}
			rewritten, ok := clashRewriteKeyLineValue(lines[keyLine], keyNode, "")
			if !ok {
				return clashTextEdit{}, false, false
			}
			end, ok := clashInlineValueEnd(lines, keyNode, resolved, keyLine)
			if !ok {
				return clashTextEdit{}, false, false
			}
			replacement := append([]string{rewritten}, itemLines...)
			return clashTextEdit{start: keyLine, end: end, lines: replacement}, true, true
		}
		hasPlaceholder := false
		for _, item := range resolved.Content {
			if resolvedItem := resolveYAMLAlias(item); resolvedItem != nil && resolvedItem.Kind == yaml.ScalarNode && resolvedItem.Value == clashAllProxiesPlaceholder {
				hasPlaceholder = true
				break
			}
		}
		if !hasPlaceholder {
			return clashTextEdit{}, false, true
		}
		flowText, ok := clashFlowSequenceText(resolved.Content, names)
		if !ok {
			return clashTextEdit{}, false, false
		}
		line := lines[keyLine]
		startColumn, endColumn, ok := clashFlowSequenceSpan(line, resolved.Column-1)
		if !ok {
			return clashTextEdit{}, false, false
		}
		rewritten := line[:startColumn] + flowText + line[endColumn:]
		return clashTextEdit{start: keyLine, end: keyLine + 1, lines: []string{rewritten}}, true, true
	}

	// 块序列：只有 __ALL_PROXIES__ 这一项需要就地展开，其余条目和注释保持原样
	for _, item := range resolved.Content {
		resolvedItem := resolveYAMLAlias(item)
		if resolvedItem == nil || resolvedItem.Kind != yaml.ScalarNode || resolvedItem.Value != clashAllProxiesPlaceholder {
			continue
		}
		itemLine := item.Line - 1
		if itemLine < 0 || itemLine >= len(lines) {
			return clashTextEdit{}, false, false
		}
		dashIndent, ok := clashLineDashIndent(lines[itemLine])
		if !ok {
			return clashTextEdit{}, false, false
		}
		itemLines, ok := clashNameItemLines(names, dashIndent)
		if !ok {
			return clashTextEdit{}, false, false
		}
		return clashTextEdit{start: itemLine, end: itemLine + 1, lines: itemLines}, true, true
	}

	return clashTextEdit{}, false, true
}

// clashBlockItemLines 把 YAML 节点序列渲染成指定缩进的块序列行。
func clashBlockItemLines(nodes []*yaml.Node, indent int) ([]string, bool) {
	if len(nodes) == 0 {
		return nil, true
	}
	cloned := make([]*yaml.Node, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			return nil, false
		}
		clone := cloneYAMLNode(node)
		cloned = append(cloned, &clone)
	}
	document := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{{Kind: yaml.SequenceNode, Tag: "!!seq", Content: cloned}},
	}
	encoded, ok := clashEncodeNode(document)
	if !ok {
		return nil, false
	}
	encoded = strings.TrimRight(encoded, "\n")
	if encoded == "" {
		return nil, true
	}
	rawLines := strings.Split(encoded, "\n")
	prefix := strings.Repeat(" ", indent)
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, prefix+line)
	}
	return lines, true
}

// clashNameItemLines 把节点名渲染成块序列的条目行。
func clashNameItemLines(names []string, indent int) ([]string, bool) {
	prefix := strings.Repeat(" ", indent)
	lines := make([]string, 0, len(names))
	for _, name := range names {
		encoded, ok := clashEncodeScalar(name)
		if !ok {
			return nil, false
		}
		lines = append(lines, prefix+"- "+encoded)
	}
	return lines, true
}

// clashFlowSequenceText 重建流式代理组列表，占位符展开为全部节点。
func clashFlowSequenceText(items []*yaml.Node, names []string) (string, bool) {
	parts := make([]string, 0, len(items)+len(names))
	for _, item := range items {
		resolved := resolveYAMLAlias(item)
		if resolved == nil || resolved.Kind != yaml.ScalarNode {
			return "", false
		}
		if resolved.Value == clashAllProxiesPlaceholder {
			for _, name := range names {
				encoded, ok := clashEncodeScalar(name)
				if !ok {
					return "", false
				}
				parts = append(parts, encoded)
			}
			continue
		}
		encoded, ok := clashEncodeScalar(resolved.Value)
		if !ok {
			return "", false
		}
		parts = append(parts, encoded)
	}
	return "[" + strings.Join(parts, ", ") + "]", true
}

// clashEncodeScalar 把单个字符串编码成一行 YAML 标量，保留 emoji 等字符。
func clashEncodeScalar(value string) (string, bool) {
	document := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}},
	}
	encoded, ok := clashEncodeNode(document)
	if !ok {
		return "", false
	}
	encoded = strings.TrimRight(encoded, "\n")
	if encoded == "" || strings.Contains(encoded, "\n") {
		return "", false
	}
	return encoded, true
}

// clashEncodeNode 编码 YAML 节点并还原 BMP 之外的字符。
func clashEncodeNode(document *yaml.Node) (string, bool) {
	substitutions := substituteClashAstralCharacters(document)
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(document); err != nil {
		return "", false
	}
	if err := encoder.Close(); err != nil {
		return "", false
	}
	return string(restoreClashAstralCharacters(output.Bytes(), substitutions)), true
}

// clashRewriteKeyLineValue 把键行上的内联值改成 replacement（空字符串表示删除），
// 保留原有缩进与行尾注释。
func clashRewriteKeyLineValue(line string, keyNode *yaml.Node, replacement string) (string, bool) {
	keyStart := keyNode.Column - 1
	if keyStart < 0 || keyStart >= len(line) {
		return "", false
	}
	colon := keyStart + len(keyNode.Value)
	if colon >= len(line) || line[colon] != ':' {
		return "", false
	}
	rest := line[colon+1:]
	comment := ""
	if index := strings.Index(rest, "#"); index >= 0 {
		comment = rest[index:]
	}
	rewritten := line[:colon+1]
	if replacement != "" {
		rewritten += " " + replacement
	}
	if comment != "" {
		rewritten += " " + comment
	}
	return rewritten, true
}

// clashMappingKeyValue 返回映射中键值对的两个节点，键不存在时 ok=false。
func clashMappingKeyValue(mapping *yaml.Node, key string) (keyNode, valueNode *yaml.Node, ok bool) {
	mapping = resolveYAMLAlias(mapping)
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil, nil, false
	}
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index], mapping.Content[index+1], true
		}
	}
	return nil, nil, false
}

// clashTemplateItemIndent 取模板惯用的序列缩进，优先参考 proxy-groups。
func clashTemplateItemIndent(lines []string, root *yaml.Node) int {
	if groupsValue := yamlMappingValue(root, "proxy-groups"); groupsValue != nil {
		if groupsNode := resolveYAMLAlias(groupsValue); groupsNode != nil &&
			groupsNode.Kind == yaml.SequenceNode && groupsNode.Style&yaml.FlowStyle == 0 &&
			len(groupsNode.Content) > 0 {
			line := groupsNode.Content[0].Line - 1
			if line < 0 || line >= len(lines) {
				return 2
			}
			if indent, ok := clashLineDashIndent(lines[line]); ok {
				return indent
			}
		}
	}
	return 2
}

// clashBlockItemIndent 返回块序列条目的缩进。
func clashBlockItemIndent(lines []string, sequence *yaml.Node, fallback int) int {
	if len(sequence.Content) == 0 {
		return fallback
	}
	itemLine := sequence.Content[0].Line - 1
	if itemLine < 0 || itemLine >= len(lines) {
		return fallback
	}
	if indent, ok := clashLineDashIndent(lines[itemLine]); ok {
		return indent
	}
	return fallback
}

// clashBlockSequenceEnd 返回块序列最后一行的下一行下标。
func clashBlockSequenceEnd(lines []string, sequence *yaml.Node) int {
	if len(sequence.Content) == 0 {
		return sequence.Line - 1
	}
	firstLine := sequence.Content[0].Line - 1
	if firstLine < 0 || firstLine >= len(lines) {
		return len(lines)
	}
	itemIndent, ok := clashLineDashIndent(lines[firstLine])
	if !ok {
		itemIndent = sequence.Column - 1
	}
	last := firstLine
	for index := firstLine + 1; index < len(lines); index++ {
		trimmed := strings.TrimSpace(lines[index])
		if trimmed == "" {
			continue
		}
		indent := clashLineIndent(lines[index])
		if strings.HasPrefix(trimmed, "#") {
			// 注释后面跟着更外层的内容时，它属于后面的键，不能算进本序列
			if indent >= itemIndent && !clashCommentLeadsNextKey(lines, index+1, itemIndent) {
				last = index
				continue
			}
			break
		}
		if indent < itemIndent {
			break
		}
		// 与条目同缩进却不是 "-" 开头，说明已经进入下一个键（零缩进序列很常见）
		if indent == itemIndent && !strings.HasPrefix(trimmed, "-") {
			break
		}
		last = index
	}
	return last + 1
}

// clashCommentLeadsNextKey 判断从 start 行起注释之后的第一个内容行是否属于外层结构。
func clashCommentLeadsNextKey(lines []string, start, indent int) bool {
	for index := start; index < len(lines); index++ {
		trimmed := strings.TrimSpace(lines[index])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		return clashLineIndent(lines[index]) < indent
	}
	return true
}

// clashBlockMappingEnd 返回块映射最后一行的下一行下标。
func clashBlockMappingEnd(lines []string, mapping *yaml.Node) int {
	start := mapping.Line - 1
	if start < 0 || start >= len(lines) {
		return len(lines)
	}
	indent := mapping.Column - 1
	last := start
	for index := start + 1; index < len(lines); index++ {
		trimmed := strings.TrimSpace(lines[index])
		if trimmed == "" {
			continue
		}
		lineIndent := clashLineIndent(lines[index])
		if strings.HasPrefix(trimmed, "#") {
			if lineIndent >= indent && !clashCommentLeadsNextKey(lines, index+1, indent) {
				last = index
				continue
			}
			break
		}
		if lineIndent < indent {
			break
		}
		last = index
	}
	return last + 1
}

// clashLineDashIndent 返回块序列条目 "-" 的缩进。
func clashLineDashIndent(line string) (int, bool) {
	for index := 0; index < len(line); index++ {
		switch line[index] {
		case ' ', '\t':
			continue
		case '-':
			if index+1 < len(line) && line[index+1] != ' ' && line[index+1] != '\t' {
				return 0, false
			}
			return index, true
		default:
			return 0, false
		}
	}
	return 0, false
}

// clashLineIndent 返回行首空格缩进。
func clashLineIndent(line string) int {
	indent := 0
	for indent < len(line) && line[indent] == ' ' {
		indent++
	}
	return indent
}

// clashFlowSequenceSpan 返回键行内流式序列的起止下标（右端不含）。
func clashFlowSequenceSpan(line string, startColumn int) (int, int, bool) {
	if startColumn < 0 || startColumn >= len(line) || line[startColumn] != '[' {
		return 0, 0, false
	}
	depth := 0
	inSingle := false
	inDouble := false
	for index := startColumn; index < len(line); index++ {
		switch {
		case inSingle:
			if line[index] == '\'' {
				inSingle = false
			}
		case inDouble:
			switch line[index] {
			case '\\':
				index++
			case '"':
				inDouble = false
			}
		case line[index] == '\'':
			inSingle = true
		case line[index] == '"':
			inDouble = true
		case line[index] == '[':
			depth++
		case line[index] == ']':
			depth--
			if depth == 0 {
				return startColumn, index + 1, true
			}
		}
	}
	return 0, 0, false
}

// applyClashTextEdits 按行号应用替换，保持未改动行的原始内容。
func applyClashTextEdits(lines []string, edits []clashTextEdit) ([]string, bool) {
	sorted := make([]clashTextEdit, len(edits))
	copy(sorted, edits)
	sort.SliceStable(sorted, func(left, right int) bool { return sorted[left].start < sorted[right].start })

	result := make([]string, 0, len(lines))
	cursor := 0
	for _, edit := range sorted {
		if edit.start < cursor || edit.end < edit.start || edit.end > len(lines) {
			return nil, false
		}
		result = append(result, lines[cursor:edit.start]...)
		result = append(result, edit.lines...)
		cursor = edit.end
	}
	result = append(result, lines[cursor:]...)
	return result, true
}
