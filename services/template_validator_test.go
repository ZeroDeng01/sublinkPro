package services

import "testing"

func TestValidateTemplateCandidateRejectsRemovedProtectedToken(t *testing.T) {
	result := ValidateTemplateCandidate(TemplateValidationInput{
		Category:      "clash",
		OriginalText:  "proxy-groups:\n  - name: Test\n    proxies:\n      - __ALL_PROXIES__\n",
		CandidateText: "proxy-groups:\n  - name: Test\n    proxies:\n      - DIRECT\n",
	})
	if result.Valid {
		t.Fatal("expected validation to fail when __ALL_PROXIES__ is removed")
	}
}

func TestValidateTemplateCandidateRejectsCategoryMismatch(t *testing.T) {
	result := ValidateTemplateCandidate(TemplateValidationInput{
		Category:      "clash",
		OriginalText:  "port: 7890\nproxies: []\n",
		CandidateText: "[General]\nloglevel = notify\n[Proxy]\nDIRECT = direct\n",
	})
	if result.Valid {
		t.Fatal("expected validation to fail on category mismatch")
	}
}

func TestValidateTemplateCandidateAcceptsBasicSurgeTemplate(t *testing.T) {
	result := ValidateTemplateCandidate(TemplateValidationInput{
		Category:      "surge",
		OriginalText:  "[General]\n[Proxy]\nDIRECT = direct\n",
		CandidateText: "[General]\n[Proxy]\nDIRECT = direct\n[Rule]\nFINAL,DIRECT\n",
	})
	if !result.Valid {
		t.Fatalf("expected surge template to validate, got errors: %#v", result.Errors)
	}
}
