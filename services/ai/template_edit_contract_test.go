package ai

import (
	"errors"
	"testing"
)

func TestTemplateEditOperationSchemaAcceptsValidOperations(t *testing.T) {
	tests := []struct {
		name       string
		payload    string
		wantOp     TemplateEditOperationType
		wantOld    string
		wantNew    string
		wantAnchor string
		wantPos    TemplateEditInsertPosition
		wantMatch  TemplateEditMatchMode
	}{
		{
			name:    "replace",
			payload: `{"operations":[{"op":"replace","oldString":"rules:\n  - MATCH,DIRECT","newString":"rules:\n  - GEOIP,CN,DIRECT\n  - MATCH,Proxy","description":"add CN rule"}]}`,
			wantOp:  TemplateEditOperationReplace,
			wantOld: "rules:\n  - MATCH,DIRECT",
			wantNew: "rules:\n  - GEOIP,CN,DIRECT\n  - MATCH,Proxy",
		},
		{
			name:      "replace all matches",
			payload:   `{"operations":[{"op":"replace","oldString":"  - 🇺🇸 美国节点","newString":"  - DIRECT","match":"all"}]}`,
			wantOp:    TemplateEditOperationReplace,
			wantOld:   "  - 🇺🇸 美国节点",
			wantNew:   "  - DIRECT",
			wantMatch: TemplateEditMatchAll,
		},
		{
			name:       "insert before",
			payload:    `{"operations":[{"op":"insert","anchor":"  - MATCH,DIRECT","newString":"  - DOMAIN-SUFFIX,example.com,Proxy\n","position":"before"}]}`,
			wantOp:     TemplateEditOperationInsert,
			wantNew:    "  - DOMAIN-SUFFIX,example.com,Proxy\n",
			wantAnchor: "  - MATCH,DIRECT",
			wantPos:    TemplateEditInsertBefore,
		},
		{
			name:    "delete",
			payload: `{"operations":[{"op":"delete","oldString":"  - DOMAIN-SUFFIX,ads.example,REJECT"}]}`,
			wantOp:  TemplateEditOperationDelete,
			wantOld: "  - DOMAIN-SUFFIX,ads.example,REJECT",
		},
		{
			name:      "delete explicit unique",
			payload:   `{"operations":[{"op":"delete","oldString":"  - DOMAIN-SUFFIX,ads.example,REJECT","match":"unique"}]}`,
			wantOp:    TemplateEditOperationDelete,
			wantOld:   "  - DOMAIN-SUFFIX,ads.example,REJECT",
			wantMatch: TemplateEditMatchUnique,
		},
		{
			name:    "replace explicit empty match",
			payload: `{"operations":[{"op":"replace","oldString":"a","newString":"b","match":""}]}`,
			wantOp:  TemplateEditOperationReplace,
			wantOld: "a",
			wantNew: "b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, err := ParseTemplateEditOperations(tt.payload)
			if err != nil {
				t.Fatalf("parse operations: %v", err)
			}
			if len(ops) != 1 {
				t.Fatalf("expected one operation, got %d", len(ops))
			}
			got := ops[0]
			if got.Op != tt.wantOp {
				t.Fatalf("expected op %q, got %q", tt.wantOp, got.Op)
			}
			if got.OldString != tt.wantOld {
				t.Fatalf("expected oldString %q, got %q", tt.wantOld, got.OldString)
			}
			if got.NewString != tt.wantNew {
				t.Fatalf("expected newString %q, got %q", tt.wantNew, got.NewString)
			}
			if got.Anchor != tt.wantAnchor {
				t.Fatalf("expected anchor %q, got %q", tt.wantAnchor, got.Anchor)
			}
			if got.Position != tt.wantPos {
				t.Fatalf("expected position %q, got %q", tt.wantPos, got.Position)
			}
			if got.Match != tt.wantMatch {
				t.Fatalf("expected match %q, got %q", tt.wantMatch, got.Match)
			}
		})
	}
}

func TestParseTemplateEditOperationsRejectsInvalidSchema(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		wantCode TemplateEditErrorCode
	}{
		{
			name:     "missing op",
			payload:  `{"operations":[{"oldString":"a","newString":"b"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "missing replace oldString",
			payload:  `{"operations":[{"op":"replace","newString":"b"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "replace forbids anchor",
			payload:  `{"operations":[{"op":"replace","oldString":"a","newString":"b","anchor":"a"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "replace forbids position",
			payload:  `{"operations":[{"op":"replace","oldString":"a","newString":"b","position":"before"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "replace no-op",
			payload:  `{"operations":[{"op":"replace","oldString":"same","newString":"same"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "replace invalid match",
			payload:  `{"operations":[{"op":"replace","oldString":"a","newString":"b","match":"some"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "operation unsupported field",
			payload:  `{"operations":[{"op":"replace","oldString":"a","newString":"b","path":"rules.0"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "missing insert anchor",
			payload:  `{"operations":[{"op":"insert","newString":"new","position":"after"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "missing insert newString",
			payload:  `{"operations":[{"op":"insert","anchor":"target","position":"after"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "missing insert position",
			payload:  `{"operations":[{"op":"insert","anchor":"target","newString":"new"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "invalid insert position",
			payload:  `{"operations":[{"op":"insert","anchor":"target","newString":"new","position":"during"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "insert forbids oldString",
			payload:  `{"operations":[{"op":"insert","oldString":"old","anchor":"target","newString":"new","position":"after"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "insert forbids all match",
			payload:  `{"operations":[{"op":"insert","anchor":"target","newString":"new","position":"after","match":"all"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "insert forbids explicit unique match",
			payload:  `{"operations":[{"op":"insert","anchor":"target","newString":"new","position":"after","match":"unique"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "missing delete oldString",
			payload:  `{"operations":[{"op":"delete"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "delete forbids newString",
			payload:  `{"operations":[{"op":"delete","oldString":"old","newString":""}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "delete forbids anchor",
			payload:  `{"operations":[{"op":"delete","oldString":"old","anchor":"old"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "delete forbids position",
			payload:  `{"operations":[{"op":"delete","oldString":"old","position":"after"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "unsupported op",
			payload:  `{"operations":[{"op":"move","oldString":"a","newString":"b"}]}`,
			wantCode: TemplateEditInvalidOperation,
		},
		{
			name:     "empty operations",
			payload:  `{"operations":[]}`,
			wantCode: TemplateEditEmptyOperations,
		},
		{
			name:     "legacy candidateText",
			payload:  `{"summary":"legacy","candidateText":"full template"}`,
			wantCode: TemplateEditLegacyOutput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTemplateEditOperations(tt.payload)
			assertTemplateEditErrorCode(t, err, tt.wantCode)
		})
	}
}

func TestParseTemplateEditOperationsRejectsMalformedOutput(t *testing.T) {
	tests := []string{
		`{"operations":[`,
		`[]`,
		`{"operations":"replace this"}`,
		`{"operations":[{"op":"replace","oldString":false,"newString":"new"}]}`,
		``,
	}

	for _, payload := range tests {
		t.Run(payload, func(t *testing.T) {
			_, err := ParseTemplateEditOperations(payload)
			assertTemplateEditErrorCode(t, err, TemplateEditMalformedOutput)
		})
	}
}

func TestParseTemplateEditOperationsHandlesMarkdownFencedJSON(t *testing.T) {
	validJSONBlock := "```json\n{\"operations\":[{\"op\":\"replace\",\"oldString\":\"a\",\"newString\":\"b\"}]}\n```"
	if _, err := ParseTemplateEditOperations(validJSONBlock); err != nil {
		t.Fatalf("expected fenced json block to parse: %v", err)
	}

	validPlainBlock := "```\n{\"operations\":[{\"op\":\"delete\",\"oldString\":\"obsolete\"}]}\n```"
	if _, err := ParseTemplateEditOperations(validPlainBlock); err != nil {
		t.Fatalf("expected fenced block with empty language to parse: %v", err)
	}
}

func TestParseTemplateEditOperationsRejectsInvalidMarkdownFences(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "prose before block",
			payload: "Here is the JSON:\n```json\n{\"operations\":[{\"op\":\"replace\",\"oldString\":\"a\",\"newString\":\"b\"}]}\n```",
		},
		{
			name:    "prose after block",
			payload: "```json\n{\"operations\":[{\"op\":\"replace\",\"oldString\":\"a\",\"newString\":\"b\"}]}\n```\nDone.",
		},
		{
			name:    "multiple blocks",
			payload: "```json\n{\"operations\":[{\"op\":\"replace\",\"oldString\":\"a\",\"newString\":\"b\"}]}\n```\n```json\n{\"operations\":[{\"op\":\"delete\",\"oldString\":\"c\"}]}\n```",
		},
		{
			name:    "unsupported language",
			payload: "```yaml\noperations: []\n```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTemplateEditOperations(tt.payload)
			assertTemplateEditErrorCode(t, err, TemplateEditMalformedOutput)
		})
	}
}

func assertTemplateEditErrorCode(t *testing.T, err error, want TemplateEditErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s, got nil", want)
	}
	var editErr *TemplateEditError
	if !errors.As(err, &editErr) {
		t.Fatalf("expected TemplateEditError, got %T: %v", err, err)
	}
	if editErr.Code != want {
		t.Fatalf("expected code %s, got %s (%v)", want, editErr.Code, err)
	}
}
