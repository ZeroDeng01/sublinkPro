package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sublink/models"
	"sublink/services"
)

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
	Summary       string                 `json:"summary"`
	Warnings      []string               `json:"warnings"`
	CandidateText string                 `json:"candidateText"`
	RevisionHash  string                 `json:"revisionHash"`
	Validation    ValidationResult       `json:"validation"`
	FinishReason  string                 `json:"finishReason,omitempty"`
	Usage         map[string]interface{} `json:"usage,omitempty"`
}

type AssistantModelOutput struct {
	Summary       string   `json:"summary"`
	Warnings      []string `json:"warnings"`
	CandidateText string   `json:"candidateText"`
}

func BuildRevisionHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func BuildPrompt(req GenerateRequest) []Message {
	systemPrompt := strings.TrimSpace(`You are a template editing assistant for a subscription management system.
You must only edit the provided template content.
Return JSON only.
Do not change the template dialect.
Preserve unrelated content, preserve comments where possible, and prefer minimal edits.
Important repository-specific rules:
- clash templates use keys like proxies:, proxy-groups:, rules:, rule-providers:
- surge templates use [Proxy], [Proxy Group], [Rule]
- __ALL_PROXIES__ is a reserved system placeholder and must be preserved unless the user explicitly asks to remove it.
- include-all, include-all-proxies, include-all-providers, use, filter, exclude-filter, exclude-type, expected-status, policy-regex-filter are system-significant fields and must not be removed accidentally.
- Do not convert clash syntax to surge or surge syntax to clash.
- If the request is unsafe or ambiguous, keep candidateText equal to the original template and explain in warnings.
Return JSON with keys: summary, warnings, candidateText.`)
	userPayload := map[string]interface{}{
		"filename":         req.Filename,
		"category":         req.Category,
		"ruleSource":       req.RuleSource,
		"useProxy":         req.UseProxy,
		"proxyLink":        req.ProxyLink,
		"enableIncludeAll": req.EnableIncludeAll,
		"userPrompt":       req.UserPrompt,
		"currentTemplate":  req.CurrentText,
	}
	payloadBytes, _ := json.Marshal(userPayload)
	return []Message{{Role: "system", Content: systemPrompt}, {Role: "user", Content: string(payloadBytes)}}
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
		return nil, fmt.Errorf("当前用户未启用 AI 助手")
	}
	if settings.BaseURL == "" || settings.Model == "" || settings.RawAPIKey == "" {
		return nil, fmt.Errorf("AI 设置不完整，请先配置 Base URL、模型和 API Key")
	}
	client, err := NewClient(ClientConfig{
		BaseURL:      settings.BaseURL,
		APIKey:       settings.RawAPIKey,
		Model:        settings.Model,
		Temperature:  settings.Temperature,
		MaxTokens:    settings.MaxTokens,
		ExtraHeaders: settings.ExtraHeaders,
	})
	if err != nil {
		return nil, err
	}
	content, finishReason, usage, err := client.StreamResponses(ctx, BuildPrompt(req), func(event ResponsesEvent) error {
		if onEvent == nil {
			return nil
		}
		return onEvent(event)
	})
	if err != nil {
		return nil, err
	}
	var output AssistantModelOutput
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		return nil, fmt.Errorf("AI 返回格式无效: %w", err)
	}
	if strings.TrimSpace(output.CandidateText) == "" {
		output.CandidateText = req.CurrentText
	}
	validation := services.ValidateTemplateCandidate(services.TemplateValidationInput{
		Category:      req.Category,
		OriginalText:  req.CurrentText,
		CandidateText: output.CandidateText,
		RuleSource:    req.RuleSource,
	})
	response := &GenerateResponse{
		Summary:       strings.TrimSpace(output.Summary),
		Warnings:      append(output.Warnings, validation.Warnings...),
		CandidateText: output.CandidateText,
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
