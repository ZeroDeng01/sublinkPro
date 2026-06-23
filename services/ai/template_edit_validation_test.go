package ai

import "testing"

func TestTemplateEditValidationErrorBlocksAccept(t *testing.T) {
	baseText := "proxies: []\nproxy-groups:\n  - name: Auto\n    proxies:\n      - __ALL_PROXIES__\nrules: []\n"
	state, err := BuildTemplateEditPreviewCandidate(TemplateEditPreviewInput{
		Category: "clash",
		BaseText: baseText,
		BaseHash: BuildRevisionHash(baseText),
		Operations: []TemplateEditOperation{{
			Op:        TemplateEditOperationDelete,
			OldString: "      - __ALL_PROXIES__\n",
		}},
	})
	assertTemplateEditErrorCode(t, err, TemplateEditValidationFailed)
	if state.Status != TemplateEditSessionFailed {
		t.Fatalf("expected failed status, got %q", state.Status)
	}
	if state.Patch.Success != true {
		t.Fatalf("expected patch to apply before validation blocks preview: %#v", state.Patch)
	}
	if state.Validation.Valid {
		t.Fatalf("expected validation to be invalid: %#v", state.Validation)
	}
	if len(state.Validation.Errors) == 0 {
		t.Fatalf("expected validation errors for removed protected token")
	}
	if len(state.Validation.ProtectedTokensFound) == 0 {
		t.Fatalf("expected protected tokens to be reported")
	}
	if state.WarningFingerprint != "" {
		t.Fatalf("validation failures should not produce warning fingerprint, got %q", state.WarningFingerprint)
	}

	err = ValidateTemplateEditAccept(TemplateEditAcceptValidationInput{Validation: state.Validation})
	assertTemplateEditErrorCode(t, err, TemplateEditValidationFailed)
}

func TestTemplateEditWarningsDoNotBlockAccept(t *testing.T) {
	baseText := "proxies: []\nproxy-groups: []\nrules: []\n"
	state, err := BuildTemplateEditPreviewCandidate(TemplateEditPreviewInput{
		Category:      "clash",
		BaseText:      baseText,
		RuleSource:    "https://example.test/rules.list",
		Subscriptions: []string{"Default Subscription"},
		Operations: []TemplateEditOperation{{
			Op:        TemplateEditOperationInsert,
			Anchor:    "rules: []\n",
			NewString: "# AI inserted rule note\n",
			Position:  TemplateEditInsertAfter,
		}},
	})
	if err != nil {
		t.Fatalf("build preview candidate: %v", err)
	}
	if state.Status != TemplateEditSessionPreviewReady {
		t.Fatalf("expected preview_ready status, got %q", state.Status)
	}
	if !state.Validation.Valid {
		t.Fatalf("expected valid candidate, got %#v", state.Validation)
	}
	if len(state.Validation.Warnings) == 0 {
		t.Fatalf("expected rule source warning")
	}
	if state.WarningFingerprint == "" {
		t.Fatalf("expected warning fingerprint metadata")
	}
	if got := state.Validation.Subscriptions; len(got) != 1 || got[0] != "Default Subscription" {
		t.Fatalf("expected subscriptions to be preserved, got %#v", got)
	}

	if err := ValidateTemplateEditAccept(TemplateEditAcceptValidationInput{Validation: state.Validation}); err != nil {
		t.Fatalf("warning-only validation should be accepted without confirmation: %v", err)
	}
}

func TestTemplateEditWarningFingerprint(t *testing.T) {
	candidateHash := BuildRevisionHash("candidate")
	fingerprint := BuildTemplateEditWarningFingerprint([]string{"beta", "alpha"}, candidateHash)
	if fingerprint != "4756fbd4a0cbcb93966154d7be2f684cb4cb7cfb5c27689d62f1622cf3ca15e4" {
		t.Fatalf("unexpected warning fingerprint %q", fingerprint)
	}
	if fingerprint != BuildTemplateEditWarningFingerprint([]string{"alpha", "beta"}, candidateHash) {
		t.Fatalf("warning fingerprint must be independent of warning order")
	}
	if fingerprint == BuildTemplateEditWarningFingerprint([]string{"alpha", "beta"}, BuildRevisionHash("different candidate")) {
		t.Fatalf("warning fingerprint must include candidate hash")
	}
	if fingerprint == BuildTemplateEditWarningFingerprint([]string{"alpha", "gamma"}, candidateHash) {
		t.Fatalf("warning fingerprint must include sorted warning contents")
	}
}
