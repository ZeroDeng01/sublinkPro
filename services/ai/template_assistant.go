package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sublink/models"
	"sublink/services"
	"unicode/utf8"
)

const (
	templateEditFullTemplateMaxChars = 4000
	templateEditContextMaxChars      = 16000
	templateEditPromptWindowRadius   = 4
	templateEditPromptBoundaryLines  = 18
	templateEditPromptMaxLineChars   = 320
)

var templateEditProtectedTokens = []string{
	"__ALL_PROXIES__",
	"include-all",
	"include-all-proxies",
	"include-all-providers",
	"use",
	"filter",
	"exclude-filter",
	"exclude-type",
	"expected-status",
	"policy-regex-filter",
}

type GenerateRequest struct {
	Filename         string `json:"filename"`
	Category         string `json:"category"`
	CurrentText      string `json:"currentText"`
	UserPrompt       string `json:"userPrompt"`
	RuleSource       string `json:"ruleSource"`
	UseProxy         bool   `json:"useProxy"`
	ProxyLink        string `json:"proxyLink"`
	EnableIncludeAll bool   `json:"enableIncludeAll"`
}

type ValidationResult struct {
	Valid                bool     `json:"valid"`
	Errors               []string `json:"errors"`
	Warnings             []string `json:"warnings"`
	DetectedType         string   `json:"detectedType"`
	ProtectedTokensFound []string `json:"protectedTokensFound"`
	Subscriptions        []string `json:"subscriptions,omitempty"`
}

type GenerateResponse struct {
	Summary       string                  `json:"summary"`
	Warnings      []string                `json:"warnings"`
	CandidateText string                  `json:"candidateText"`
	Operations    []TemplateEditOperation `json:"operations"`
	PatchResult   PatchResult             `json:"patchResult"`
	RevisionHash  string                  `json:"revisionHash"`
	Validation    ValidationResult        `json:"validation"`
	FinishReason  string                  `json:"finishReason,omitempty"`
	Usage         map[string]any          `json:"usage,omitempty"`
}

type AssistantModelOutput = TemplateEditModelOutput

type templateEditPromptPayload struct {
	Metadata        templateEditPromptMetadata            `json:"metadata"`
	UserPrompt      string                                `json:"userPrompt"`
	CategoryRules   []string                              `json:"categoryRules"`
	ProtectedTokens []templateEditProtectedTokenInventory `json:"protectedTokens"`
	Context         templateEditPromptContext             `json:"context"`
}

type templateEditPromptMetadata struct {
	Filename               string `json:"filename"`
	Category               string `json:"category"`
	RuleSourceConfigured   bool   `json:"ruleSourceConfigured"`
	UseProxy               bool   `json:"useProxy"`
	ProxyLinkProvided      bool   `json:"proxyLinkProvided"`
	EnableIncludeAll       bool   `json:"enableIncludeAll"`
	TemplateLength         int    `json:"templateLength"`
	FullTemplateIncluded   bool   `json:"fullTemplateIncluded"`
	ContextWindowCharLimit int    `json:"contextWindowCharLimit"`
}

type templateEditProtectedTokenInventory struct {
	Token   string `json:"token"`
	Present bool   `json:"present"`
	Count   int    `json:"count"`
	Lines   []int  `json:"lines,omitempty"`
}

type templateEditPromptContext struct {
	CurrentTemplate   string                            `json:"currentTemplate,omitempty"`
	LineNumberingNote string                            `json:"lineNumberingNote"`
	Windows           []templateEditPromptContextWindow `json:"contextWindows"`
}

type templateEditPromptContextWindow struct {
	Title        string `json:"title"`
	Reason       string `json:"reason"`
	StartLine    int    `json:"startLine"`
	EndLine      int    `json:"endLine"`
	NumberedText string `json:"numberedText"`
}

type templateEditPromptLineRange struct {
	Start  int
	End    int
	Reason string
	Title  string
}

func BuildRevisionHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func BuildPrompt(req GenerateRequest) []Message {
	return BuildTemplateEditPrompt(req)
}

func BuildTemplateEditPrompt(req GenerateRequest) []Message {
	systemPrompt := strings.TrimSpace(`You are a template editing assistant for a subscription management system.
Return valid JSON only. Do not wrap the JSON in Markdown.

You must return an operation-only edit plan. The root JSON object must contain exactly these keys:
- summary: short human-readable summary of the intended edit.
- warnings: array of warning codes or short warning strings.
- operations: array of replace, insert, and delete operations.

Never return a full rewritten template. Never include a root candidateText field or any other full-template output field. The server applies operations with exact matching and owns preview materialization.

Operation schema:
- replace: {"op":"replace","oldString":"exact text to replace","newString":"replacement text","match":"unique|all","description":"why"}
- insert: {"op":"insert","anchor":"exact unique anchor text","position":"before|after","newString":"text to insert","description":"why"}
- delete: {"op":"delete","oldString":"exact text to delete","match":"unique|all","description":"why"}
- match is optional. Omit it for the default "unique" mode.

Examples:
{"summary":"Replace the final rule policy","warnings":[],"operations":[{"op":"replace","oldString":"  - MATCH,DIRECT","newString":"  - MATCH,Proxy","description":"Route unmatched traffic through Proxy"}]}
{"summary":"Insert a domain rule before the final rule","warnings":[],"operations":[{"op":"insert","anchor":"  - MATCH,DIRECT","position":"before","newString":"  - DOMAIN-SUFFIX,example.com,Proxy\n","description":"Add example.com routing rule"}]}
{"summary":"Delete an obsolete reject rule","warnings":[],"operations":[{"op":"delete","oldString":"  - DOMAIN-SUFFIX,ads.example,REJECT\n","description":"Remove obsolete ad rule"}]}
{"summary":"Delete every USA node entry","warnings":[],"operations":[{"op":"delete","oldString":"      - 🇺🇸 美国节点\n","match":"all","description":"Remove every exact USA node entry because the user asked for all"}]}

Safety rules:
- Use only exact text visible in currentTemplate or contextWindows for oldString and anchor.
- Line numbers are context only and must never appear in operations.
- Without an explicit all/every/全部/所有 style user request, use the default unique mode and choose an exact target that appears exactly once.
- For replace and delete only, use "match":"all" when the user explicitly asks to change all/every/全部/所有 exact occurrences of the same visible text.
- Never include match for insert operations. Insert anchors must remain exact unique anchors.
- If an exact unique target, exact unique anchor, or intentional all-match target cannot be inferred from the provided context, return {"summary":"Ambiguous edit target","warnings":["AI_EDIT_AMBIGUOUS_TARGET"],"operations":[]} instead of guessing.
- Do not change the template dialect. Do not convert clash syntax to surge or surge syntax to clash.
- Preserve unrelated content, comments where possible, protected tokens, and system-significant fields unless the user explicitly asks to change them.`)
	payloadBytes, _ := json.MarshalIndent(buildTemplateEditPromptPayload(req), "", "  ")
	return []Message{{Role: "system", Content: systemPrompt}, {Role: "user", Content: string(payloadBytes)}}
}

func buildTemplateEditPromptPayload(req GenerateRequest) templateEditPromptPayload {
	templateLength := templateEditCharCount(req.CurrentText)
	includeFullTemplate := templateLength <= templateEditFullTemplateMaxChars
	return templateEditPromptPayload{
		Metadata: templateEditPromptMetadata{
			Filename:               req.Filename,
			Category:               req.Category,
			RuleSourceConfigured:   strings.TrimSpace(req.RuleSource) != "",
			UseProxy:               req.UseProxy,
			ProxyLinkProvided:      strings.TrimSpace(req.ProxyLink) != "",
			EnableIncludeAll:       req.EnableIncludeAll,
			TemplateLength:         templateLength,
			FullTemplateIncluded:   includeFullTemplate,
			ContextWindowCharLimit: templateEditContextMaxChars,
		},
		UserPrompt:      req.UserPrompt,
		CategoryRules:   templateEditCategoryRules(req.Category),
		ProtectedTokens: buildTemplateEditProtectedTokenInventory(req.CurrentText),
		Context:         buildTemplateEditPromptContext(req.CurrentText, req.UserPrompt, includeFullTemplate),
	}
}

func templateEditCategoryRules(category string) []string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "clash":
		return []string{
			"clash templates use YAML keys such as proxies:, proxy-groups:, rules:, and rule-providers:",
			"preserve system-significant fields such as include-all, use, filter, exclude-filter, exclude-type, expected-status, and policy-regex-filter unless explicitly asked to change them",
			"do not convert clash syntax to surge syntax",
		}
	case "surge":
		return []string{
			"surge templates use sections such as [Proxy], [Proxy Group], and [Rule]",
			"preserve existing section structure and policy names unless explicitly asked to change them",
			"do not convert surge syntax to clash syntax",
		}
	default:
		return []string{
			"preserve the detected template dialect and surrounding structure",
			"prefer minimal exact operations over broad rewrites",
		}
	}
}

func buildTemplateEditProtectedTokenInventory(template string) []templateEditProtectedTokenInventory {
	lines := strings.Split(normalizeTemplateEditLineEndings(template), "\n")
	inventory := make([]templateEditProtectedTokenInventory, 0, len(templateEditProtectedTokens))
	for _, token := range templateEditProtectedTokens {
		item := templateEditProtectedTokenInventory{Token: token}
		for index, line := range lines {
			count := strings.Count(line, token)
			if count == 0 {
				continue
			}
			item.Present = true
			item.Count += count
			item.Lines = append(item.Lines, index+1)
		}
		inventory = append(inventory, item)
	}
	return inventory
}

func buildTemplateEditPromptContext(template string, userPrompt string, includeFullTemplate bool) templateEditPromptContext {
	normalized := normalizeTemplateEditLineEndings(template)
	lines := strings.Split(normalized, "\n")
	if normalized == "" {
		lines = []string{""}
	}
	context := templateEditPromptContext{
		LineNumberingNote: "Line numbers and L<n> prefixes are context only. They are not part of the template and must not appear in operation text.",
		Windows:           buildTemplateEditPromptWindows(lines, userPrompt),
	}
	if includeFullTemplate {
		context.CurrentTemplate = normalized
	}
	return context
}

func buildTemplateEditPromptWindows(lines []string, userPrompt string) []templateEditPromptContextWindow {
	ranges := []templateEditPromptLineRange{
		{Start: 0, End: minInt(len(lines)-1, templateEditPromptBoundaryLines-1), Title: "template-start", Reason: "beginning of template"},
	}
	if len(lines) > templateEditPromptBoundaryLines {
		ranges = append(ranges, templateEditPromptLineRange{Start: maxInt(0, len(lines)-templateEditPromptBoundaryLines), End: len(lines) - 1, Title: "template-end", Reason: "end of template"})
	}
	for _, token := range templateEditProtectedTokens {
		for index, line := range lines {
			if strings.Contains(line, token) {
				ranges = append(ranges, centeredTemplateEditPromptRange(index, len(lines), "protected-token", "protected token "+token))
			}
		}
	}
	for _, keyword := range templateEditPromptKeywords(userPrompt) {
		matches := 0
		for index, line := range lines {
			if strings.Contains(strings.ToLower(line), keyword) {
				ranges = append(ranges, centeredTemplateEditPromptRange(index, len(lines), "prompt-keyword", "line matching user prompt keyword "+keyword))
				matches++
				if matches >= 4 {
					break
				}
			}
		}
	}
	return materializeTemplateEditPromptWindows(lines, mergeTemplateEditPromptRanges(ranges), templateEditContextMaxChars)
}

func centeredTemplateEditPromptRange(center int, lineCount int, title string, reason string) templateEditPromptLineRange {
	return templateEditPromptLineRange{
		Start:  maxInt(0, center-templateEditPromptWindowRadius),
		End:    minInt(lineCount-1, center+templateEditPromptWindowRadius),
		Title:  title,
		Reason: reason,
	}
}

func templateEditPromptKeywords(userPrompt string) []string {
	fields := strings.FieldsFunc(strings.ToLower(userPrompt), func(r rune) bool {
		return !isTemplateEditPromptKeywordRune(r)
	})
	seen := map[string]bool{}
	keywords := make([]string, 0, 8)
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if len([]rune(field)) < 4 || seen[field] {
			continue
		}
		seen[field] = true
		keywords = append(keywords, field)
		if len(keywords) >= 8 {
			break
		}
	}
	return keywords
}

func isTemplateEditPromptKeywordRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || (r >= '\u4e00' && r <= '\u9fff')
}

func mergeTemplateEditPromptRanges(ranges []templateEditPromptLineRange) []templateEditPromptLineRange {
	if len(ranges) == 0 {
		return nil
	}
	sort.SliceStable(ranges, func(i, j int) bool {
		if ranges[i].Start == ranges[j].Start {
			return ranges[i].End < ranges[j].End
		}
		return ranges[i].Start < ranges[j].Start
	})
	merged := []templateEditPromptLineRange{ranges[0]}
	for _, current := range ranges[1:] {
		last := &merged[len(merged)-1]
		if current.Start <= last.End+1 {
			if current.End > last.End {
				last.End = current.End
			}
			if !strings.Contains(last.Reason, current.Reason) {
				last.Reason += "; " + current.Reason
			}
			continue
		}
		merged = append(merged, current)
	}
	return merged
}

func materializeTemplateEditPromptWindows(lines []string, ranges []templateEditPromptLineRange, maxChars int) []templateEditPromptContextWindow {
	windows := make([]templateEditPromptContextWindow, 0, len(ranges))
	remaining := maxChars
	for index, lineRange := range ranges {
		if remaining <= 0 {
			break
		}
		numbered := buildTemplateEditNumberedText(lines, lineRange.Start, lineRange.End, remaining)
		if numbered == "" {
			break
		}
		remaining -= len(numbered)
		title := lineRange.Title
		if title == "" {
			title = fmt.Sprintf("context-window-%d", index+1)
		}
		windows = append(windows, templateEditPromptContextWindow{
			Title:        title,
			Reason:       lineRange.Reason,
			StartLine:    lineRange.Start + 1,
			EndLine:      minInt(lineRange.End+1, len(lines)),
			NumberedText: numbered,
		})
	}
	return windows
}

func buildTemplateEditNumberedText(lines []string, start int, end int, maxChars int) string {
	var builder strings.Builder
	for lineIndex := start; lineIndex <= end && lineIndex < len(lines); lineIndex++ {
		line := fmt.Sprintf("L%d | %s\n", lineIndex+1, truncateTemplateEditPromptLine(lines[lineIndex]))
		if builder.Len()+len(line) > maxChars {
			break
		}
		builder.WriteString(line)
	}
	return builder.String()
}

func truncateTemplateEditPromptLine(line string) string {
	if templateEditCharCount(line) <= templateEditPromptMaxLineChars {
		return line
	}
	runes := []rune(line)
	return string(runes[:templateEditPromptMaxLineChars]) + " [...truncated]"
}

func templateEditCharCount(input string) int {
	return utf8.RuneCountInString(input)
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func GenerateCandidate(ctx context.Context, user *models.User, req GenerateRequest) (*GenerateResponse, error) {
	return GenerateCandidateStream(ctx, user, req, nil)
}

func GenerateCandidateStream(ctx context.Context, user *models.User, req GenerateRequest, onEvent func(ResponsesEvent) error) (*GenerateResponse, error) {
	settings, err := user.GetAISettings()
	if err != nil {
		return nil, err
	}
	if !settings.Enabled {
		return nil, fmt.Errorf("AI 助手未启用")
	}
	if settings.BaseURL == "" || settings.Model == "" || settings.RawAPIKey == "" {
		return nil, fmt.Errorf("AI 设置不完整，请先配置 Base URL、模型和 API Key")
	}
	client, err := NewClient(ClientConfig{
		BaseURL:      settings.BaseURL,
		APIKey:       settings.RawAPIKey,
		Model:        settings.Model,
		RequestType:  settings.RequestType,
		Temperature:  settings.Temperature,
		MaxTokens:    settings.MaxTokens,
		ExtraHeaders: settings.ExtraHeaders,
	})
	if err != nil {
		return nil, err
	}
	messages := BuildPrompt(req)
	var content string
	var finishReason string
	var usage map[string]any
	if settings.RequestType == models.SystemAIRequestTypeChatCompletions {
		content, finishReason, usage, err = client.StreamChatCompletions(ctx, messages, onEvent)
	} else {
		content, finishReason, usage, err = client.StreamResponses(ctx, messages, func(event ResponsesEvent) error {
			if onEvent == nil {
				return nil
			}
			return onEvent(event)
		})
	}
	if err != nil {
		return nil, err
	}
	output, err := ParseTemplateEditModelOutput(content)
	if err != nil {
		return nil, fmt.Errorf("AI 返回格式无效: %w", err)
	}
	candidateText, patchResult, err := ApplyTemplateEditOperations(req.CurrentText, output.Operations)
	if err != nil {
		return nil, fmt.Errorf("AI 编辑操作无法应用: %w", err)
	}
	validation := services.ValidateTemplateCandidate(services.TemplateValidationInput{
		Category:      req.Category,
		OriginalText:  req.CurrentText,
		CandidateText: candidateText,
		RuleSource:    req.RuleSource,
	})
	response := &GenerateResponse{
		Summary:       strings.TrimSpace(output.Summary),
		Warnings:      append(output.Warnings, validation.Warnings...),
		CandidateText: candidateText,
		Operations:    output.Operations,
		PatchResult:   patchResult,
		RevisionHash:  BuildRevisionHash(req.CurrentText),
		Validation: ValidationResult{
			Valid:                validation.Valid,
			Errors:               validation.Errors,
			Warnings:             validation.Warnings,
			DetectedType:         validation.DetectedType,
			ProtectedTokensFound: validation.ProtectedTokensFound,
		},
		FinishReason: finishReason,
		Usage:        usage,
	}
	return response, nil
}
