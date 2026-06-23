package ai

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestBuildTemplateEditPromptOperationOnly(t *testing.T) {
	messages := BuildTemplateEditPrompt(GenerateRequest{
		Filename:    "clash.yaml",
		Category:    "clash",
		CurrentText: "proxies: []\nproxy-groups:\n  - name: Auto\n    proxies:\n      - __ALL_PROXIES__\nrules:\n  - MATCH,DIRECT\n",
		UserPrompt:  "Route final traffic through Proxy",
	})

	if len(messages) != 2 {
		t.Fatalf("expected system and user messages, got %d", len(messages))
	}
	systemPrompt := messages[0].Content
	for _, want := range []string{
		"operation-only edit plan",
		"operations: array of replace, insert, and delete operations",
		`{"op":"replace"`,
		`{"op":"insert"`,
		`{"op":"delete"`,
		`"match":"unique|all"`,
		`"match":"all"`,
		"all/every/全部/所有",
		`Never include match for insert operations`,
		"AI_EDIT_AMBIGUOUS_TARGET",
		"Never return a full rewritten template",
		"Line numbers are context only",
	} {
		if !strings.Contains(systemPrompt, want) {
			t.Fatalf("expected system prompt to contain %q\n%s", want, systemPrompt)
		}
	}
	legacyOutputField := "candidate" + "Text"
	for _, oldInstruction := range []string{
		"Return JSON with keys: summary, warnings, " + legacyOutputField,
		"keep " + legacyOutputField + " equal to the original template",
	} {
		if strings.Contains(systemPrompt, oldInstruction) {
			t.Fatalf("system prompt still contains legacy instruction %q", oldInstruction)
		}
	}

	var payload templateEditPromptPayload
	if err := json.Unmarshal([]byte(messages[1].Content), &payload); err != nil {
		t.Fatalf("unmarshal prompt payload: %v", err)
	}
	if payload.Context.CurrentTemplate == "" {
		t.Fatal("expected small template to include currentTemplate")
	}
	if !payload.Metadata.FullTemplateIncluded {
		t.Fatal("expected small template metadata to mark full template included")
	}
	if len(payload.Context.Windows) == 0 {
		t.Fatal("expected line-numbered context windows")
	}
	if !strings.Contains(payload.Context.Windows[0].NumberedText, "L1 |") {
		t.Fatalf("expected numbered context lines, got %q", payload.Context.Windows[0].NumberedText)
	}
}

func TestParseTemplateEditModelOutputRejectsLegacyCandidateTextAtRoot(t *testing.T) {
	legacyOutputField := "candidate" + "Text"
	_, err := ParseTemplateEditModelOutput(`{"summary":"legacy","warnings":[],"` + legacyOutputField + `":"rules:\n  - MATCH,DIRECT\n"}`)
	assertTemplateEditErrorCode(t, err, TemplateEditLegacyOutput)
	var editErr *TemplateEditError
	if !errors.As(err, &editErr) || editErr.Code != TemplateEditLegacyOutput {
		t.Fatalf("expected root legacy model output to return %s, got %v", TemplateEditLegacyOutput, err)
	}
}

func TestBuildTemplateEditPromptOmitsFullLargeTemplate(t *testing.T) {
	largeTemplate := buildLargeTemplateForPromptTest()
	messages := BuildTemplateEditPrompt(GenerateRequest{
		Filename:    "large.yaml",
		Category:    "clash",
		CurrentText: largeTemplate,
		UserPrompt:  "Add a final proxy rule",
	})

	if len(messages) != 2 {
		t.Fatalf("expected system and user messages, got %d", len(messages))
	}
	if strings.Contains(messages[1].Content, `"currentTemplate"`) {
		t.Fatalf("large template payload must omit currentTemplate: %s", messages[1].Content)
	}
	if strings.Contains(messages[1].Content, "UNIQUE_MIDDLE_SENTINEL_SHOULD_NOT_BE_SENT") {
		t.Fatal("large template payload included middle content outside selected context windows")
	}

	var payload templateEditPromptPayload
	if err := json.Unmarshal([]byte(messages[1].Content), &payload); err != nil {
		t.Fatalf("unmarshal prompt payload: %v", err)
	}
	if payload.Metadata.FullTemplateIncluded {
		t.Fatal("expected large template metadata to mark full template omitted")
	}
	if payload.Metadata.TemplateLength <= templateEditFullTemplateMaxChars {
		t.Fatalf("test fixture must exceed full template threshold, got %d", payload.Metadata.TemplateLength)
	}
	if len(payload.Context.Windows) == 0 {
		t.Fatal("expected compact context windows for large template")
	}
	totalContextChars := 0
	for _, window := range payload.Context.Windows {
		if !strings.Contains(window.NumberedText, "L") || !strings.Contains(window.NumberedText, " | ") {
			t.Fatalf("expected line-numbered window, got %q", window.NumberedText)
		}
		totalContextChars += len(window.NumberedText)
	}
	if totalContextChars > templateEditContextMaxChars {
		t.Fatalf("expected context windows <= %d chars, got %d", templateEditContextMaxChars, totalContextChars)
	}
}

func TestTemplateEditOperationOutputParsesAndAppliesPatch(t *testing.T) {
	modelOutput := `{"summary":"Update final policy","warnings":[],"operations":[{"op":"replace","oldString":"  - MATCH,DIRECT","newString":"  - MATCH,Proxy","description":"route final traffic"}]}`
	parsed, err := ParseTemplateEditModelOutput(modelOutput)
	if err != nil {
		t.Fatalf("parse model output: %v", err)
	}
	candidate, result, err := ApplyTemplateEditOperations("rules:\n  - MATCH,DIRECT\n", parsed.Operations)
	if err != nil {
		t.Fatalf("apply operations: %v", err)
	}
	if !result.Success || result.OperationsApplied != 1 {
		t.Fatalf("expected one successful operation, got %+v", result)
	}
	if candidate != "rules:\n  - MATCH,Proxy\n" {
		t.Fatalf("unexpected candidate: %q", candidate)
	}
}

func TestTemplateEditMockProviderFixtures(t *testing.T) {
	baseText := strings.Join([]string{
		"# QA_DNS_COMMENT: use default resolver",
		"# QA_DELETE_ME: temporary test comment",
		"# QA_RULESOURCE_MARKER: original",
		"proxies: []",
		"proxy-groups:",
		"  - name: Auto",
		"    type: select",
		"    proxies:",
		"      - __ALL_PROXIES__",
		"rules:",
		"  - DOMAIN-SUFFIX,duplicate.example,DIRECT",
		"  - DOMAIN-SUFFIX,duplicate.example,DIRECT",
		"  - MATCH,DIRECT",
		"",
	}, "\n")

	tests := []struct {
		name         string
		prompt       string
		wantContains string
		wantAbsent   string
		wantErr      string
		ruleSource   string
		wantWarning  bool
	}{
		{name: "replace DNS comment", prompt: "QA_REPLACE_DNS_COMMENT", wantContains: "# QA_DNS_COMMENT: use deterministic resolver"},
		{name: "insert proxy group", prompt: "QA_INSERT_PROXY_GROUP", wantContains: "name: QA Inserted"},
		{name: "delete comment", prompt: "QA_DELETE_TEST_COMMENT", wantContains: "# QA_RULESOURCE_MARKER: original"},
		{name: "duplicate ambiguity", prompt: "QA_DUPLICATE_MATCH", wantErr: string(PatchAmbiguousMatch)},
		{name: "delete all duplicate matches", prompt: "QA_DELETE_ALL_DUPLICATE_MATCHES", wantAbsent: "  - DOMAIN-SUFFIX,duplicate.example,DIRECT"},
		{name: "protected token validation", prompt: "QA_REMOVE_PROTECTED_TOKEN", wantErr: string(TemplateEditValidationFailed)},
		{name: "rule-source warning", prompt: "QA_RULESOURCE_WARNING", wantContains: "# QA_RULESOURCE_MARKER: changed", ruleSource: "https://example.test/rules.list", wantWarning: true},
	}

	client, err := NewClient(ClientConfig{BaseURL: TemplateEditMockBaseURL, Model: "template-ai-edit-mock", RequestType: RequestTypeChatCompletions})
	if err != nil {
		t.Fatalf("new mock client: %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, _, _, err := client.StreamChatCompletions(context.Background(), BuildTemplateEditPrompt(GenerateRequest{Filename: "qa.yaml", Category: "clash", CurrentText: baseText, UserPrompt: tt.prompt}), nil)
			if err != nil {
				t.Fatalf("mock stream: %v", err)
			}
			output, err := ParseTemplateEditModelOutput(content)
			if err != nil {
				t.Fatalf("parse mock output: %v", err)
			}
			preview, err := BuildTemplateEditPreviewCandidate(TemplateEditPreviewInput{Category: "clash", BaseText: baseText, Operations: output.Operations, RuleSource: tt.ruleSource})
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %s, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("build preview: %v", err)
			}
			if !strings.Contains(preview.CandidateText, tt.wantContains) {
				t.Fatalf("candidate missing %q:\n%s", tt.wantContains, preview.CandidateText)
			}
			if tt.wantAbsent != "" && strings.Contains(preview.CandidateText, tt.wantAbsent) {
				t.Fatalf("candidate still contains %q:\n%s", tt.wantAbsent, preview.CandidateText)
			}
			if tt.wantWarning && len(preview.Validation.Warnings) == 0 {
				t.Fatal("expected real validation warning")
			}
		})
	}
}

func TestTemplateEditMockProviderRejectsProductionMode(t *testing.T) {
	oldAppEnv := os.Getenv("APP_ENV")
	t.Cleanup(func() { _ = os.Setenv("APP_ENV", oldAppEnv) })
	if err := os.Setenv("APP_ENV", "production"); err != nil {
		t.Fatalf("set APP_ENV: %v", err)
	}
	_, err := NewClient(ClientConfig{BaseURL: TemplateEditMockBaseURL, Model: "template-ai-edit-mock"})
	assertTemplateEditErrorCode(t, err, TemplateEditMockProviderUnavailable)
}

func buildLargeTemplateForPromptTest() string {
	var builder strings.Builder
	for index := 0; index < 220; index++ {
		if index == 110 {
			builder.WriteString("  - DOMAIN-SUFFIX,UNIQUE_MIDDLE_SENTINEL_SHOULD_NOT_BE_SENT,DIRECT\n")
			continue
		}
		builder.WriteString("  - DOMAIN-SUFFIX,example")
		builder.WriteString(strings.Repeat("x", 24))
		builder.WriteString(",DIRECT\n")
	}
	return builder.String()
}
