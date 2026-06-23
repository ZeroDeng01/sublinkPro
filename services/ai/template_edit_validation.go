package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"sublink/services"
)

type TemplateEditPreviewInput struct {
	Category      string
	BaseText      string
	BaseHash      string
	CandidateHash string
	Operations    []TemplateEditOperation
	RuleSource    string
	Subscriptions []string
}

type TemplateEditPreviewState struct {
	Status             TemplateEditSessionStatus `json:"status"`
	BaseHash           string                    `json:"baseHash"`
	CandidateHash      string                    `json:"candidateHash"`
	CandidateText      string                    `json:"candidateText"`
	Operations         []TemplateEditOperation   `json:"operations"`
	Patch              PatchResult               `json:"patch"`
	Validation         ValidationResult          `json:"validation"`
	WarningFingerprint string                    `json:"warningFingerprint,omitempty"`
	Error              *TemplateEditError        `json:"error,omitempty"`
}

type TemplateEditAcceptValidationInput struct {
	Validation ValidationResult
}

func BuildTemplateEditPreviewCandidate(input TemplateEditPreviewInput) (TemplateEditPreviewState, error) {
	baseHash := strings.TrimSpace(input.BaseHash)
	if baseHash == "" {
		baseHash = BuildRevisionHash(input.BaseText)
	}
	state := TemplateEditPreviewState{
		Status:     TemplateEditSessionValidating,
		BaseHash:   baseHash,
		Operations: cloneTemplateEditOperations(input.Operations),
	}

	candidate, patchResult, err := ApplyTemplateEditOperations(input.BaseText, input.Operations)
	state.Patch = patchResult
	if err != nil {
		state.Status = TemplateEditSessionFailed
		return state, err
	}

	state.CandidateText = candidate
	state.CandidateHash = BuildRevisionHash(candidate)
	validation := services.ValidateTemplateCandidate(services.TemplateValidationInput{
		Category:      input.Category,
		OriginalText:  input.BaseText,
		CandidateText: candidate,
		RuleSource:    input.RuleSource,
	})
	state.Validation = templateEditValidationResultFromService(validation, input.Subscriptions)
	if !state.Validation.Valid || len(state.Validation.Errors) > 0 {
		state.Status = TemplateEditSessionFailed
		state.Error = NewTemplateEditError(TemplateEditValidationFailed, "template edit candidate failed validation")
		return state, state.Error
	}

	if len(state.Validation.Warnings) > 0 {
		state.WarningFingerprint = BuildTemplateEditWarningFingerprint(state.Validation.Warnings, state.CandidateHash)
	}
	state.Status = TemplateEditSessionPreviewReady
	return state, nil
}

func ValidateTemplateEditAccept(input TemplateEditAcceptValidationInput) error {
	if !input.Validation.Valid || len(input.Validation.Errors) > 0 {
		return NewTemplateEditError(TemplateEditValidationFailed, "template edit candidate failed validation")
	}
	return nil
}

func BuildTemplateEditWarningFingerprint(warnings []string, candidateHash string) string {
	sortedWarnings := cloneStringSlice(warnings)
	sort.Strings(sortedWarnings)
	hash := sha256.New()
	for _, warning := range sortedWarnings {
		hash.Write([]byte(warning))
		hash.Write([]byte{0})
	}
	hash.Write([]byte(strings.TrimSpace(candidateHash)))
	return hex.EncodeToString(hash.Sum(nil))
}

func templateEditValidationResultFromService(result services.TemplateValidationResult, subscriptions []string) ValidationResult {
	return ValidationResult{
		Valid:                result.Valid,
		Errors:               cloneStringSlice(result.Errors),
		Warnings:             cloneStringSlice(result.Warnings),
		DetectedType:         result.DetectedType,
		ProtectedTokensFound: cloneStringSlice(result.ProtectedTokensFound),
		Subscriptions:        cloneStringSlice(subscriptions),
	}
}
