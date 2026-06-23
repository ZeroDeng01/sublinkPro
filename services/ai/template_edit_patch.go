package ai

import (
	"fmt"
	"strings"
)

type PatchErrorCode string

const (
	PatchNoMatch               PatchErrorCode = "PATCH_NO_MATCH"
	PatchAmbiguousMatch        PatchErrorCode = "PATCH_AMBIGUOUS_MATCH"
	PatchEmptyTarget           PatchErrorCode = "PATCH_EMPTY_TARGET"
	PatchNoop                  PatchErrorCode = "PATCH_NOOP"
	PatchInvalidInsertPosition PatchErrorCode = "PATCH_INVALID_INSERT_POSITION"
	PatchInvalidMatchMode      PatchErrorCode = "PATCH_INVALID_MATCH_MODE"
)

type PatchResult struct {
	Success           bool           `json:"success"`
	OperationIndex    int            `json:"operationIndex"`
	OperationsApplied int            `json:"operationsApplied"`
	ErrorCode         PatchErrorCode `json:"errorCode,omitempty"`
	ErrorMessage      string         `json:"errorMessage,omitempty"`
}

type PatchError struct {
	Code           PatchErrorCode `json:"code"`
	OperationIndex int            `json:"operationIndex"`
	Message        string         `json:"message"`
}

func (err *PatchError) Error() string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%s at operation %d: %s", err.Code, err.OperationIndex, err.Message)
}

func ApplyTemplateEditOperations(baseText string, ops []TemplateEditOperation) (candidate string, result PatchResult, err error) {
	newlineStyle := detectTemplateEditNewlineStyle(baseText)
	working := normalizeTemplateEditLineEndings(baseText)

	for index, op := range ops {
		next, patchErr := applyTemplateEditOperation(working, op, index)
		if patchErr != nil {
			return "", PatchResult{
				Success:        false,
				OperationIndex: index,
				ErrorCode:      patchErr.Code,
				ErrorMessage:   patchErr.Message,
			}, patchErr
		}
		working = next
	}

	return restoreTemplateEditLineEndings(working, newlineStyle), PatchResult{
		Success:           true,
		OperationIndex:    -1,
		OperationsApplied: len(ops),
	}, nil
}

func applyTemplateEditOperation(candidate string, op TemplateEditOperation, operationIndex int) (string, *PatchError) {
	switch op.Op {
	case TemplateEditOperationReplace:
		matchMode, patchErr := templateEditOperationMatchMode(op, operationIndex)
		if patchErr != nil {
			return "", patchErr
		}
		oldString := normalizeTemplateEditLineEndings(op.OldString)
		newString := normalizeTemplateEditLineEndings(op.NewString)
		if oldString == "" {
			return "", newPatchError(PatchEmptyTarget, operationIndex, "replace operation requires non-empty oldString")
		}
		if oldString == newString {
			return "", newPatchError(PatchNoop, operationIndex, "replace operation oldString and newString are identical")
		}
		if matchMode == TemplateEditMatchAll {
			if !strings.Contains(candidate, oldString) {
				return "", newPatchError(PatchNoMatch, operationIndex, "operation target did not match the current candidate")
			}
			return strings.ReplaceAll(candidate, oldString, newString), nil
		}
		matchIndex, patchErr := findUniqueTemplateEditMatch(candidate, oldString, operationIndex)
		if patchErr != nil {
			return "", patchErr
		}
		return candidate[:matchIndex] + newString + candidate[matchIndex+len(oldString):], nil
	case TemplateEditOperationInsert:
		matchMode, patchErr := templateEditOperationMatchMode(op, operationIndex)
		if patchErr != nil {
			return "", patchErr
		}
		if op.Match != "" || matchMode == TemplateEditMatchAll {
			return "", newPatchError(PatchInvalidMatchMode, operationIndex, "insert operation forbids match")
		}
		anchor := normalizeTemplateEditLineEndings(op.Anchor)
		newString := normalizeTemplateEditLineEndings(op.NewString)
		if anchor == "" {
			return "", newPatchError(PatchEmptyTarget, operationIndex, "insert operation requires non-empty anchor")
		}
		if newString == "" {
			return "", newPatchError(PatchNoop, operationIndex, "insert operation requires non-empty newString")
		}
		if op.Position != TemplateEditInsertBefore && op.Position != TemplateEditInsertAfter {
			return "", newPatchError(PatchInvalidInsertPosition, operationIndex, "insert operation requires position before or after")
		}
		matchIndex, patchErr := findUniqueTemplateEditMatch(candidate, anchor, operationIndex)
		if patchErr != nil {
			return "", patchErr
		}
		switch op.Position {
		case TemplateEditInsertBefore:
			return candidate[:matchIndex] + newString + candidate[matchIndex:], nil
		case TemplateEditInsertAfter:
			insertAt := matchIndex + len(anchor)
			return candidate[:insertAt] + newString + candidate[insertAt:], nil
		default:
			return "", newPatchError(PatchInvalidInsertPosition, operationIndex, "insert operation requires position before or after")
		}
	case TemplateEditOperationDelete:
		matchMode, patchErr := templateEditOperationMatchMode(op, operationIndex)
		if patchErr != nil {
			return "", patchErr
		}
		oldString := normalizeTemplateEditLineEndings(op.OldString)
		if oldString == "" {
			return "", newPatchError(PatchEmptyTarget, operationIndex, "delete operation requires non-empty oldString")
		}
		if matchMode == TemplateEditMatchAll {
			if !strings.Contains(candidate, oldString) {
				return "", newPatchError(PatchNoMatch, operationIndex, "operation target did not match the current candidate")
			}
			return strings.ReplaceAll(candidate, oldString, ""), nil
		}
		matchIndex, patchErr := findUniqueTemplateEditMatch(candidate, oldString, operationIndex)
		if patchErr != nil {
			return "", patchErr
		}
		return candidate[:matchIndex] + candidate[matchIndex+len(oldString):], nil
	default:
		return "", newPatchError(PatchNoop, operationIndex, "operation type must be replace, insert, or delete")
	}
}

func templateEditOperationMatchMode(op TemplateEditOperation, operationIndex int) (TemplateEditMatchMode, *PatchError) {
	switch op.Match {
	case "", TemplateEditMatchUnique:
		return TemplateEditMatchUnique, nil
	case TemplateEditMatchAll:
		return TemplateEditMatchAll, nil
	default:
		return "", newPatchError(PatchInvalidMatchMode, operationIndex, "operation match must be unique or all")
	}
}

func findUniqueTemplateEditMatch(candidate string, target string, operationIndex int) (int, *PatchError) {
	matchIndex := strings.Index(candidate, target)
	if matchIndex < 0 {
		return -1, newPatchError(PatchNoMatch, operationIndex, "operation target did not match the current candidate")
	}
	if strings.Contains(candidate[matchIndex+1:], target) {
		return -1, newPatchError(PatchAmbiguousMatch, operationIndex, "operation target matched more than once")
	}
	return matchIndex, nil
}

func newPatchError(code PatchErrorCode, operationIndex int, message string) *PatchError {
	return &PatchError{Code: code, OperationIndex: operationIndex, Message: message}
}

func normalizeTemplateEditLineEndings(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	return strings.ReplaceAll(input, "\r", "\n")
}

func detectTemplateEditNewlineStyle(input string) string {
	if strings.Contains(input, "\r\n") {
		return "\r\n"
	}
	if strings.Contains(input, "\r") {
		return "\r"
	}
	return "\n"
}

func restoreTemplateEditLineEndings(input string, newlineStyle string) string {
	if newlineStyle == "\n" {
		return input
	}
	return strings.ReplaceAll(input, "\n", newlineStyle)
}
