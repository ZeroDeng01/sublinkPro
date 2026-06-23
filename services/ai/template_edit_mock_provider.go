package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	templateEditMockModel = "template-ai-edit-mock"
)

func streamTemplateEditMockOutput(ctx context.Context, messages []Message, onEvent func(ResponsesEvent) error) (string, string, map[string]any, error) {
	content, err := buildTemplateEditMockModelOutput(messages)
	if err != nil {
		return "", "", nil, err
	}
	delay := templateEditMockStreamDelay()
	if onEvent != nil {
		for _, chunk := range splitTemplateEditMockChunks(content) {
			if delay > 0 {
				select {
				case <-ctx.Done():
					return "", "", nil, ctx.Err()
				case <-time.After(delay):
				}
			} else if err := ctx.Err(); err != nil {
				return "", "", nil, err
			}
			payload, err := json.Marshal(map[string]string{"delta": chunk})
			if err != nil {
				return "", "", nil, err
			}
			if err := onEvent(ResponsesEvent{Event: "response.output_text.delta", Data: payload}); err != nil {
				return "", "", nil, err
			}
		}
	}
	return content, "mock", map[string]any{"provider": templateEditMockModel, "mock": true}, nil
}

func buildTemplateEditMockModelOutput(messages []Message) (string, error) {
	payload, err := templateEditMockPromptPayload(messages)
	if err != nil {
		return "", err
	}
	fixture, err := templateEditMockFixture(strings.TrimSpace(payload.UserPrompt))
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(fixture)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func templateEditMockPromptPayload(messages []Message) (templateEditPromptPayload, error) {
	for index := len(messages) - 1; index >= 0; index-- {
		if strings.TrimSpace(messages[index].Role) != "user" {
			continue
		}
		var payload templateEditPromptPayload
		if err := json.Unmarshal([]byte(messages[index].Content), &payload); err != nil {
			return templateEditPromptPayload{}, NewTemplateEditError(TemplateEditInvalidOperation, "mock template AI edit prompt payload is invalid")
		}
		return payload, nil
	}
	return templateEditPromptPayload{}, NewTemplateEditError(TemplateEditInvalidOperation, "mock template AI edit prompt payload is missing")
}

func templateEditMockFixture(prompt string) (TemplateEditModelOutput, error) {
	switch prompt {
	case "QA_REPLACE_DNS_COMMENT":
		return TemplateEditModelOutput{
			Summary:  "Replace deterministic QA DNS comment",
			Warnings: []string{},
			Operations: []TemplateEditOperation{{
				Op:          TemplateEditOperationReplace,
				OldString:   "# QA_DNS_COMMENT: use default resolver",
				NewString:   "# QA_DNS_COMMENT: use deterministic resolver",
				Description: "Replace the QA DNS comment without changing DNS behavior",
			}},
		}, nil
	case "QA_INSERT_PROXY_GROUP":
		return TemplateEditModelOutput{
			Summary:  "Insert deterministic QA proxy group",
			Warnings: []string{},
			Operations: []TemplateEditOperation{{
				Op:       TemplateEditOperationInsert,
				Anchor:   "proxy-groups:\n",
				Position: TemplateEditInsertAfter,
				NewString: "  - name: QA Inserted\n" +
					"    type: select\n" +
					"    proxies:\n" +
					"      - DIRECT\n",
				Description: "Insert a QA proxy group after the proxy-groups header",
			}},
		}, nil
	case "QA_DELETE_TEST_COMMENT":
		return TemplateEditModelOutput{
			Summary:  "Delete deterministic QA test comment",
			Warnings: []string{},
			Operations: []TemplateEditOperation{{
				Op:          TemplateEditOperationDelete,
				OldString:   "# QA_DELETE_ME: temporary test comment\n",
				Description: "Delete only the temporary QA comment",
			}},
		}, nil
	case "QA_DUPLICATE_MATCH":
		return TemplateEditModelOutput{
			Summary:  "Trigger duplicate-match patch ambiguity",
			Warnings: []string{},
			Operations: []TemplateEditOperation{{
				Op:          TemplateEditOperationReplace,
				OldString:   "  - DOMAIN-SUFFIX,duplicate.example,DIRECT",
				NewString:   "  - DOMAIN-SUFFIX,duplicate.example,Proxy",
				Description: "Intentionally target duplicated text for PATCH_AMBIGUOUS_MATCH",
			}},
		}, nil
	case "QA_DELETE_ALL_DUPLICATE_MATCHES":
		return TemplateEditModelOutput{
			Summary:  "Delete every deterministic duplicate QA rule",
			Warnings: []string{},
			Operations: []TemplateEditOperation{{
				Op:          TemplateEditOperationDelete,
				OldString:   "  - DOMAIN-SUFFIX,duplicate.example,DIRECT\n",
				Match:       TemplateEditMatchAll,
				Description: "Delete every exact duplicate QA rule because the prompt asks for all matches",
			}},
		}, nil
	case "QA_REMOVE_PROTECTED_TOKEN":
		return TemplateEditModelOutput{
			Summary:  "Trigger protected-token validation failure",
			Warnings: []string{},
			Operations: []TemplateEditOperation{{
				Op:          TemplateEditOperationDelete,
				OldString:   "      - __ALL_PROXIES__\n",
				Description: "Intentionally remove a protected token for backend validation coverage",
			}},
		}, nil
	case "QA_RULESOURCE_WARNING":
		return TemplateEditModelOutput{
			Summary:  "Replace QA rule-source marker and keep warning path real",
			Warnings: []string{},
			Operations: []TemplateEditOperation{{
				Op:          TemplateEditOperationReplace,
				OldString:   "# QA_RULESOURCE_MARKER: original",
				NewString:   "# QA_RULESOURCE_MARKER: changed",
				Description: "Change a harmless marker while backend rule-source validation raises the warning",
			}},
		}, nil
	default:
		return TemplateEditModelOutput{}, NewTemplateEditError(TemplateEditInvalidOperation, fmt.Sprintf("unsupported mock template AI edit prompt: %s", prompt))
	}
}

func splitTemplateEditMockChunks(content string) []string {
	const chunkSize = 48
	if len(content) <= chunkSize {
		return []string{content}
	}
	chunks := make([]string, 0, len(content)/chunkSize+1)
	for len(content) > chunkSize {
		chunks = append(chunks, content[:chunkSize])
		content = content[chunkSize:]
	}
	if content != "" {
		chunks = append(chunks, content)
	}
	return chunks
}
