package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type TemplateAIEditRequest struct {
	Filename         string `json:"filename"`
	Category         string `json:"category"`
	CurrentText      string `json:"currentText"`
	UserPrompt       string `json:"userPrompt"`
	RuleSource       string `json:"ruleSource"`
	UseProxy         bool   `json:"useProxy"`
	ProxyLink        string `json:"proxyLink"`
	EnableIncludeAll bool   `json:"enableIncludeAll"`
}

type TemplateEditOperationType string

const (
	TemplateEditOperationReplace TemplateEditOperationType = "replace"
	TemplateEditOperationInsert  TemplateEditOperationType = "insert"
	TemplateEditOperationDelete  TemplateEditOperationType = "delete"
)

type TemplateEditInsertPosition string

const (
	TemplateEditInsertBefore TemplateEditInsertPosition = "before"
	TemplateEditInsertAfter  TemplateEditInsertPosition = "after"
)

type TemplateEditMatchMode string

const (
	TemplateEditMatchUnique TemplateEditMatchMode = "unique"
	TemplateEditMatchAll    TemplateEditMatchMode = "all"
)

type TemplateEditOperation struct {
	Op          TemplateEditOperationType   `json:"op"`
	OldString   string                      `json:"oldString,omitempty"`
	NewString   string                      `json:"newString,omitempty"`
	Anchor      string                      `json:"anchor,omitempty"`
	Position    TemplateEditInsertPosition  `json:"position,omitempty"`
	Match       TemplateEditMatchMode       `json:"match,omitempty"`
	Description string                      `json:"description,omitempty"`
	fields      templateEditOperationFields `json:"-"`
}

type templateEditOperationFields struct {
	Op          bool
	OldString   bool
	NewString   bool
	Anchor      bool
	Position    bool
	Match       bool
	Description bool
}

type TemplateEditSessionStatus string

const (
	TemplateEditSessionCreated         TemplateEditSessionStatus = "created"
	TemplateEditSessionStreaming       TemplateEditSessionStatus = "streaming"
	TemplateEditSessionOperationsReady TemplateEditSessionStatus = "operations_ready"
	TemplateEditSessionValidating      TemplateEditSessionStatus = "validating"
	TemplateEditSessionPreviewReady    TemplateEditSessionStatus = "preview_ready"
	TemplateEditSessionAccepted        TemplateEditSessionStatus = "accepted"
	TemplateEditSessionDiscarded       TemplateEditSessionStatus = "discarded"
	TemplateEditSessionFailed          TemplateEditSessionStatus = "failed"
	TemplateEditSessionExpired         TemplateEditSessionStatus = "expired"
)

type TemplateEditSession struct {
	SessionID          string                    `json:"sessionId"`
	OwnerKey           string                    `json:"ownerKey,omitempty"`
	Status             TemplateEditSessionStatus `json:"status"`
	Filename           string                    `json:"filename"`
	Category           string                    `json:"category"`
	BaseHash           string                    `json:"baseHash"`
	BaseText           string                    `json:"baseText"`
	CandidateText      string                    `json:"candidateText"`
	Operations         []TemplateEditOperation   `json:"operations"`
	Validation         ValidationResult          `json:"validation"`
	WarningFingerprint string                    `json:"warningFingerprint"`
	CreatedAt          time.Time                 `json:"createdAt"`
	ExpiresAt          time.Time                 `json:"expiresAt"`
	LastError          string                    `json:"lastError,omitempty"`
}

type TemplateEditModelOutput struct {
	Summary    string                  `json:"summary"`
	Warnings   []string                `json:"warnings"`
	Operations []TemplateEditOperation `json:"operations"`
}

type TemplateEditErrorCode string

const (
	TemplateEditInvalidOperation        TemplateEditErrorCode = "AI_EDIT_INVALID_OPERATION"
	TemplateEditEmptyOperations         TemplateEditErrorCode = "AI_EDIT_EMPTY_OPERATIONS"
	TemplateEditMalformedOutput         TemplateEditErrorCode = "AI_EDIT_MALFORMED_OUTPUT"
	TemplateEditLegacyOutput            TemplateEditErrorCode = "AI_EDIT_LEGACY_OUTPUT"
	TemplateEditSessionExpiredError     TemplateEditErrorCode = "AI_EDIT_SESSION_EXPIRED"
	TemplateEditSessionNotFound         TemplateEditErrorCode = "AI_EDIT_SESSION_NOT_FOUND"
	TemplateEditSessionLimit            TemplateEditErrorCode = "AI_EDIT_SESSION_LIMIT"
	TemplateEditValidationFailed        TemplateEditErrorCode = "AI_EDIT_VALIDATION_FAILED"
	TemplateEditMockProviderUnavailable TemplateEditErrorCode = "AI_EDIT_MOCK_PROVIDER_UNAVAILABLE"
)

type TemplateEditError struct {
	Code    TemplateEditErrorCode `json:"code"`
	Message string                `json:"message"`
}

func (err *TemplateEditError) Error() string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", err.Code, err.Message)
}

func NewTemplateEditError(code TemplateEditErrorCode, message string) *TemplateEditError {
	return &TemplateEditError{Code: code, Message: message}
}

func ParseTemplateEditOperations(modelOutput string) ([]TemplateEditOperation, error) {
	output, err := ParseTemplateEditModelOutput(modelOutput)
	if err != nil {
		return nil, err
	}
	return output.Operations, nil
}

func ParseTemplateEditModelOutput(modelOutput string) (*TemplateEditModelOutput, error) {
	jsonPayload, err := extractTemplateEditJSON(modelOutput)
	if err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonPayload), &raw); err != nil {
		return nil, NewTemplateEditError(TemplateEditMalformedOutput, "AI edit output must be a JSON object")
	}
	for key := range raw {
		switch key {
		case "summary", "warnings", "operations":
		case "candidateText":
			return nil, NewTemplateEditError(TemplateEditLegacyOutput, "AI edit output must use operations instead of candidateText")
		default:
			return nil, NewTemplateEditError(TemplateEditMalformedOutput, "AI edit output contains unsupported field: "+key)
		}
	}
	var output TemplateEditModelOutput
	if err := json.Unmarshal([]byte(jsonPayload), &output); err != nil {
		if IsTemplateEditError(err) {
			return nil, err
		}
		return nil, NewTemplateEditError(TemplateEditMalformedOutput, "AI edit output has invalid JSON field types")
	}
	if len(output.Operations) == 0 {
		return nil, NewTemplateEditError(TemplateEditEmptyOperations, "AI edit output must include at least one operation")
	}
	for index := range output.Operations {
		if err := output.Operations[index].Validate(); err != nil {
			return nil, err
		}
	}
	return &output, nil
}

func IsTemplateEditError(err error) bool {
	_, ok := err.(*TemplateEditError)
	return ok
}

func (op *TemplateEditOperation) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return NewTemplateEditError(TemplateEditMalformedOutput, "operation must be a JSON object")
	}
	fields := templateEditOperationFields{}
	for key := range raw {
		switch key {
		case "op":
			fields.Op = true
		case "oldString":
			fields.OldString = true
		case "newString":
			fields.NewString = true
		case "anchor":
			fields.Anchor = true
		case "position":
			fields.Position = true
		case "match":
			fields.Match = true
		case "description":
			fields.Description = true
		default:
			return NewTemplateEditError(TemplateEditInvalidOperation, "operation contains unsupported field: "+key)
		}
	}
	type operationAlias TemplateEditOperation
	var decoded operationAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return NewTemplateEditError(TemplateEditMalformedOutput, "operation fields must use valid JSON types")
	}
	*op = TemplateEditOperation(decoded)
	op.fields = fields
	return nil
}

func (op TemplateEditOperation) Validate() error {
	fields := op.fields
	if !fields.Op && op.Op == "" {
		return NewTemplateEditError(TemplateEditInvalidOperation, "operation is missing op")
	}
	if _, err := op.validateMatchMode(); err != nil {
		return err
	}
	switch op.Op {
	case TemplateEditOperationReplace:
		if !fields.OldString && op.OldString == "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "replace operation requires oldString")
		}
		if op.OldString == "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "replace operation requires non-empty oldString")
		}
		if fields.Anchor || op.Anchor != "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "replace operation forbids anchor")
		}
		if fields.Position || op.Position != "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "replace operation forbids position")
		}
		if op.OldString == op.NewString {
			return NewTemplateEditError(TemplateEditInvalidOperation, "replace operation must change oldString")
		}
	case TemplateEditOperationDelete:
		if !fields.OldString && op.OldString == "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "delete operation requires oldString")
		}
		if op.OldString == "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "delete operation requires non-empty oldString")
		}
		if fields.NewString || op.NewString != "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "delete operation forbids newString")
		}
		if fields.Anchor || op.Anchor != "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "delete operation forbids anchor")
		}
		if fields.Position || op.Position != "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "delete operation forbids position")
		}
	case TemplateEditOperationInsert:
		if op.Match != "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "insert operation forbids match")
		}
		if fields.OldString || op.OldString != "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "insert operation forbids oldString")
		}
		if !fields.Anchor && op.Anchor == "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "insert operation requires anchor")
		}
		if op.Anchor == "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "insert operation requires non-empty anchor")
		}
		if !fields.NewString && op.NewString == "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "insert operation requires newString")
		}
		if op.NewString == "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "insert operation requires non-empty newString")
		}
		if !fields.Position && op.Position == "" {
			return NewTemplateEditError(TemplateEditInvalidOperation, "insert operation requires position")
		}
		if op.Position != TemplateEditInsertBefore && op.Position != TemplateEditInsertAfter {
			return NewTemplateEditError(TemplateEditInvalidOperation, "insert operation requires position before or after")
		}
	default:
		return NewTemplateEditError(TemplateEditInvalidOperation, "unsupported operation: "+string(op.Op))
	}
	return nil
}

func (op TemplateEditOperation) validateMatchMode() (TemplateEditMatchMode, error) {
	if !op.fields.Match && op.Match == "" {
		return TemplateEditMatchUnique, nil
	}
	switch op.Match {
	case "", TemplateEditMatchUnique, TemplateEditMatchAll:
		return op.Match, nil
	default:
		return "", NewTemplateEditError(TemplateEditInvalidOperation, "operation match must be unique or all")
	}
}

func extractTemplateEditJSON(modelOutput string) (string, error) {
	trimmed := strings.TrimSpace(modelOutput)
	if trimmed == "" {
		return "", NewTemplateEditError(TemplateEditMalformedOutput, "AI edit output is empty")
	}
	fences := findTemplateEditFences(trimmed)
	if len(fences) == 0 {
		if strings.Contains(trimmed, "```") {
			return "", NewTemplateEditError(TemplateEditMalformedOutput, "AI edit output contains an incomplete fenced JSON block")
		}
		return trimmed, nil
	}
	if len(fences) != 1 {
		return "", NewTemplateEditError(TemplateEditMalformedOutput, "AI edit output must contain at most one fenced JSON block")
	}
	fence := fences[0]
	if strings.TrimSpace(trimmed[:fence.start]) != "" || strings.TrimSpace(trimmed[fence.end:]) != "" {
		return "", NewTemplateEditError(TemplateEditMalformedOutput, "AI edit output must not include prose outside the JSON block")
	}
	language := strings.ToLower(strings.TrimSpace(fence.language))
	if language != "" && language != "json" {
		return "", NewTemplateEditError(TemplateEditMalformedOutput, "AI edit output fenced block must be json")
	}
	return strings.TrimSpace(fence.content), nil
}

type templateEditFence struct {
	start    int
	end      int
	language string
	content  string
}

func findTemplateEditFences(input string) []templateEditFence {
	var fences []templateEditFence
	searchFrom := 0
	for {
		openRelative := strings.Index(input[searchFrom:], "```")
		if openRelative < 0 {
			return fences
		}
		open := searchFrom + openRelative
		afterOpen := open + len("```")
		lineEndRelative := strings.IndexAny(input[afterOpen:], "\r\n")
		if lineEndRelative < 0 {
			return fences
		}
		lineEnd := afterOpen + lineEndRelative
		contentStart := lineEnd
		if input[lineEnd] == '\r' && lineEnd+1 < len(input) && input[lineEnd+1] == '\n' {
			contentStart += 2
		} else {
			contentStart++
		}
		closeRelative := strings.Index(input[contentStart:], "```")
		if closeRelative < 0 {
			return fences
		}
		close := contentStart + closeRelative
		end := close + len("```")
		fences = append(fences, templateEditFence{
			start:    open,
			end:      end,
			language: input[afterOpen:lineEnd],
			content:  input[contentStart:close],
		})
		searchFrom = end
	}
}
