package ai

import (
	"errors"
	"testing"
)

func TestApplyTemplateEditOperationsSuccess(t *testing.T) {
	tests := []struct {
		name string
		base string
		ops  []TemplateEditOperation
		want string
	}{
		{
			name: "replace",
			base: "rules:\n  - MATCH,DIRECT\n",
			ops: []TemplateEditOperation{
				{
					Op:        TemplateEditOperationReplace,
					OldString: "  - MATCH,DIRECT",
					NewString: "  - GEOIP,CN,DIRECT\n  - MATCH,Proxy",
				},
			},
			want: "rules:\n  - GEOIP,CN,DIRECT\n  - MATCH,Proxy\n",
		},
		{
			name: "insert before",
			base: "rules:\n  - MATCH,DIRECT\n",
			ops: []TemplateEditOperation{
				{
					Op:        TemplateEditOperationInsert,
					Anchor:    "  - MATCH,DIRECT",
					NewString: "  - DOMAIN-SUFFIX,example.com,Proxy\n",
					Position:  TemplateEditInsertBefore,
				},
			},
			want: "rules:\n  - DOMAIN-SUFFIX,example.com,Proxy\n  - MATCH,DIRECT\n",
		},
		{
			name: "insert after",
			base: "rules:\n  - MATCH,DIRECT\n",
			ops: []TemplateEditOperation{
				{
					Op:        TemplateEditOperationInsert,
					Anchor:    "rules:\n",
					NewString: "  - DOMAIN-SUFFIX,example.com,Proxy\n",
					Position:  TemplateEditInsertAfter,
				},
			},
			want: "rules:\n  - DOMAIN-SUFFIX,example.com,Proxy\n  - MATCH,DIRECT\n",
		},
		{
			name: "delete",
			base: "rules:\n  - DOMAIN-SUFFIX,ads.example,REJECT\n  - MATCH,DIRECT\n",
			ops: []TemplateEditOperation{
				{
					Op:        TemplateEditOperationDelete,
					OldString: "  - DOMAIN-SUFFIX,ads.example,REJECT\n",
				},
			},
			want: "rules:\n  - MATCH,DIRECT\n",
		},
		{
			name: "replace all matches",
			base: "rules:\n  - MATCH,DIRECT\n  - MATCH,DIRECT\n",
			ops: []TemplateEditOperation{
				{
					Op:        TemplateEditOperationReplace,
					OldString: "  - MATCH,DIRECT",
					NewString: "  - MATCH,Proxy",
					Match:     TemplateEditMatchAll,
				},
			},
			want: "rules:\n  - MATCH,Proxy\n  - MATCH,Proxy\n",
		},
		{
			name: "delete all matches",
			base: "proxy-groups:\n  - name: Auto\n    proxies:\n      - 🇺🇸 美国节点\n      - 🇭🇰 香港节点\n  - name: Fallback\n    proxies:\n      - 🇺🇸 美国节点\n",
			ops: []TemplateEditOperation{
				{
					Op:        TemplateEditOperationDelete,
					OldString: "      - 🇺🇸 美国节点\n",
					Match:     TemplateEditMatchAll,
				},
			},
			want: "proxy-groups:\n  - name: Auto\n    proxies:\n      - 🇭🇰 香港节点\n  - name: Fallback\n    proxies:\n",
		},
		{
			name: "multi operation success",
			base: "rules:\n  - MATCH,DIRECT\n",
			ops: []TemplateEditOperation{
				{
					Op:        TemplateEditOperationInsert,
					Anchor:    "  - MATCH,DIRECT",
					NewString: "  - DOMAIN-SUFFIX,example.com,Proxy\n",
					Position:  TemplateEditInsertBefore,
				},
				{
					Op:        TemplateEditOperationReplace,
					OldString: "  - MATCH,DIRECT",
					NewString: "  - MATCH,Proxy",
				},
			},
			want: "rules:\n  - DOMAIN-SUFFIX,example.com,Proxy\n  - MATCH,Proxy\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, result, err := ApplyTemplateEditOperations(tt.base, tt.ops)
			if err != nil {
				t.Fatalf("apply operations: %v", err)
			}
			if !result.Success {
				t.Fatalf("expected success result, got %+v", result)
			}
			if result.OperationsApplied != len(tt.ops) {
				t.Fatalf("expected %d applied operations, got %d", len(tt.ops), result.OperationsApplied)
			}
			if got != tt.want {
				t.Fatalf("expected candidate %q, got %q", tt.want, got)
			}
		})
	}
}

func TestApplyTemplateEditOperationsRejectsDuplicateMatch(t *testing.T) {
	_, result, err := ApplyTemplateEditOperations("rules:\n  - MATCH,DIRECT\n  - MATCH,DIRECT\n", []TemplateEditOperation{
		{
			Op:        TemplateEditOperationReplace,
			OldString: "  - MATCH,DIRECT",
			NewString: "  - MATCH,Proxy",
		},
	})
	assertPatchErrorCode(t, err, PatchAmbiguousMatch, 0)
	if result.Success {
		t.Fatalf("expected failed result, got %+v", result)
	}
	if result.ErrorCode != PatchAmbiguousMatch || result.OperationIndex != 0 {
		t.Fatalf("expected result code %s at operation 0, got %+v", PatchAmbiguousMatch, result)
	}
}

func TestApplyTemplateEditOperationsRejectsNoMatch(t *testing.T) {
	candidate, result, err := ApplyTemplateEditOperations("rules:\n  - MATCH,DIRECT\n", []TemplateEditOperation{
		{
			Op:        TemplateEditOperationDelete,
			OldString: "  - DOMAIN-SUFFIX,missing.example,REJECT",
		},
	})
	assertPatchErrorCode(t, err, PatchNoMatch, 0)
	if candidate != "" {
		t.Fatalf("expected no usable candidate on failure, got %q", candidate)
	}
	if result.ErrorCode != PatchNoMatch || result.OperationIndex != 0 {
		t.Fatalf("expected result code %s at operation 0, got %+v", PatchNoMatch, result)
	}

	candidate, result, err = ApplyTemplateEditOperations("rules:\n  - MATCH,DIRECT\n", []TemplateEditOperation{
		{
			Op:        TemplateEditOperationDelete,
			OldString: "  - DOMAIN-SUFFIX,missing.example,REJECT",
			Match:     TemplateEditMatchAll,
		},
	})
	assertPatchErrorCode(t, err, PatchNoMatch, 0)
	if candidate != "" {
		t.Fatalf("expected no usable candidate on all-match failure, got %q", candidate)
	}
	if result.ErrorCode != PatchNoMatch || result.OperationIndex != 0 {
		t.Fatalf("expected all-match result code %s at operation 0, got %+v", PatchNoMatch, result)
	}

	candidate, result, err = ApplyTemplateEditOperations("rules:\n  - MATCH,DIRECT\n", []TemplateEditOperation{
		{
			Op:        TemplateEditOperationReplace,
			OldString: "  - DOMAIN-SUFFIX,missing.example,REJECT",
			NewString: "  - DOMAIN-SUFFIX,missing.example,Proxy",
			Match:     TemplateEditMatchAll,
		},
	})
	assertPatchErrorCode(t, err, PatchNoMatch, 0)
	if candidate != "" {
		t.Fatalf("expected no usable candidate on replace all-match failure, got %q", candidate)
	}
	if result.ErrorCode != PatchNoMatch || result.OperationIndex != 0 {
		t.Fatalf("expected replace all-match result code %s at operation 0, got %+v", PatchNoMatch, result)
	}
}

func TestApplyTemplateEditOperationsRejectsInvalidInsertPosition(t *testing.T) {
	_, _, err := ApplyTemplateEditOperations("rules:\n  - MATCH,DIRECT\n", []TemplateEditOperation{
		{
			Op:        TemplateEditOperationInsert,
			Anchor:    "  - MATCH,DIRECT",
			NewString: "  - GEOIP,CN,DIRECT\n",
			Position:  TemplateEditInsertPosition("during"),
		},
	})
	assertPatchErrorCode(t, err, PatchInvalidInsertPosition, 0)
}

func TestApplyTemplateEditOperationsRejectsAllMatchInsert(t *testing.T) {
	_, result, err := ApplyTemplateEditOperations("rules:\n  - MATCH,DIRECT\n", []TemplateEditOperation{
		{
			Op:        TemplateEditOperationInsert,
			Anchor:    "  - MATCH,DIRECT",
			NewString: "  - GEOIP,CN,DIRECT\n",
			Position:  TemplateEditInsertBefore,
			Match:     TemplateEditMatchAll,
		},
	})
	assertPatchErrorCode(t, err, PatchInvalidMatchMode, 0)
	if result.ErrorCode != PatchInvalidMatchMode || result.OperationIndex != 0 {
		t.Fatalf("expected result code %s at operation 0, got %+v", PatchInvalidMatchMode, result)
	}
}

func TestApplyTemplateEditOperationsRejectsEmptyTargetAndNoop(t *testing.T) {
	tests := []struct {
		name     string
		ops      []TemplateEditOperation
		wantCode PatchErrorCode
	}{
		{
			name: "replace empty target",
			ops: []TemplateEditOperation{
				{Op: TemplateEditOperationReplace, NewString: "new"},
			},
			wantCode: PatchEmptyTarget,
		},
		{
			name: "insert empty anchor",
			ops: []TemplateEditOperation{
				{Op: TemplateEditOperationInsert, NewString: "new", Position: TemplateEditInsertBefore},
			},
			wantCode: PatchEmptyTarget,
		},
		{
			name: "delete empty target",
			ops: []TemplateEditOperation{
				{Op: TemplateEditOperationDelete},
			},
			wantCode: PatchEmptyTarget,
		},
		{
			name: "replace noop",
			ops: []TemplateEditOperation{
				{Op: TemplateEditOperationReplace, OldString: "same", NewString: "same"},
			},
			wantCode: PatchNoop,
		},
		{
			name: "insert noop",
			ops: []TemplateEditOperation{
				{Op: TemplateEditOperationInsert, Anchor: "same", Position: TemplateEditInsertAfter},
			},
			wantCode: PatchNoop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, result, err := ApplyTemplateEditOperations("same", tt.ops)
			assertPatchErrorCode(t, err, tt.wantCode, 0)
			if result.ErrorCode != tt.wantCode || result.OperationIndex != 0 {
				t.Fatalf("expected result code %s at operation 0, got %+v", tt.wantCode, result)
			}
		})
	}
}

func TestApplyTemplateEditOperationsPreservesCRLF(t *testing.T) {
	base := "rules:\r\n  - MATCH,DIRECT\r\n"
	ops := []TemplateEditOperation{
		{
			Op:        TemplateEditOperationReplace,
			OldString: "rules:\n  - MATCH,DIRECT\n",
			NewString: "rules:\r  - GEOIP,CN,DIRECT\r  - MATCH,Proxy\r",
		},
	}

	got, _, err := ApplyTemplateEditOperations(base, ops)
	if err != nil {
		t.Fatalf("apply operations: %v", err)
	}
	want := "rules:\r\n  - GEOIP,CN,DIRECT\r\n  - MATCH,Proxy\r\n"
	if got != want {
		t.Fatalf("expected CRLF candidate %q, got %q", want, got)
	}
}

func TestApplyTemplateEditOperationsAtomicMultiOperationFailure(t *testing.T) {
	candidate, result, err := ApplyTemplateEditOperations("rules:\n  - MATCH,DIRECT\n", []TemplateEditOperation{
		{
			Op:        TemplateEditOperationReplace,
			OldString: "  - MATCH,DIRECT",
			NewString: "  - MATCH,Proxy",
		},
		{
			Op:        TemplateEditOperationDelete,
			OldString: "  - DOMAIN-SUFFIX,missing.example,REJECT",
		},
	})
	assertPatchErrorCode(t, err, PatchNoMatch, 1)
	if candidate != "" {
		t.Fatalf("expected no usable candidate on failure, got %q", candidate)
	}
	if result.Success || result.OperationsApplied != 0 {
		t.Fatalf("expected atomic failure result without partial success, got %+v", result)
	}
	if result.OperationIndex != 1 || result.ErrorCode != PatchNoMatch {
		t.Fatalf("expected no-match result at operation 1, got %+v", result)
	}
}

func assertPatchErrorCode(t *testing.T, err error, wantCode PatchErrorCode, wantOperationIndex int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s, got nil", wantCode)
	}
	var patchErr *PatchError
	if !errors.As(err, &patchErr) {
		t.Fatalf("expected PatchError, got %T: %v", err, err)
	}
	if patchErr.Code != wantCode {
		t.Fatalf("expected code %s, got %s (%v)", wantCode, patchErr.Code, err)
	}
	if patchErr.OperationIndex != wantOperationIndex {
		t.Fatalf("expected operation index %d, got %d (%v)", wantOperationIndex, patchErr.OperationIndex, err)
	}
}
