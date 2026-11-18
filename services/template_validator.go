package services

import (
	"bufio"
	"strings"
	"sublink/node/protocol"

	"gopkg.in/yaml.v3"
)

type TemplateValidationInput struct {
	Category      string
	OriginalText  string
	CandidateText string
	RuleSource    string
}

type TemplateValidationResult struct {
	Valid                bool
	Errors               []string
	Warnings             []string
	DetectedType         string
	ProtectedTokensFound []string
	Subscriptions        []string
}

var protectedTokens = []string{
	"__ALL_PROXIES__",
	"include-all",
	"include-all-proxies",
	"include-all-providers",
	"use",
	"filter",
	"exclude-filter",
	"exclude-type",
	"expected-status",
	"policy-regex-filter",
}

func ValidateTemplateCandidate(input TemplateValidationInput) TemplateValidationResult {
	result := TemplateValidationResult{Valid: true}
	trimmedCategory := strings.TrimSpace(input.Category)
	detectedType := detectTemplateType(input.CandidateText)
	result.DetectedType = detectedType
	if detectedType != "" && trimmedCategory != "" && detectedType != trimmedCategory {
		result.Valid = false
		result.Errors = append(result.Errors, "Template content does not match the selected category")
	}
	if trimmedCategory == "clash" {
		if err := validateClash(input.CandidateText); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
		}
	}
	if trimmedCategory == "surge" {
		if err := validateSurge(input.CandidateText); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
		}
	}
	for _, token := range protectedTokens {
		if strings.Contains(input.OriginalText, token) {
			result.ProtectedTokensFound = append(result.ProtectedTokensFound, token)
			if !strings.Contains(input.CandidateText, token) {
				result.Valid = false
				result.Errors = append(result.Errors, "Candidate template removed protected semantics: "+token)
			}
		}
	}
	if strings.TrimSpace(input.RuleSource) != "" {
		result.Warnings = append(result.Warnings, "This template has a rule source; later rule conversion will overwrite parts of the generated sections")
	}
	return result
}

func validateClash(content string) error {
	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(content), &parsed); err != nil {
		return err
	}
	return nil
}

func validateSurge(content string) error {
	sections := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sections[line] = true
		}
	}
	if !sections["[Proxy]"] {
		return errString("Surge template is missing the [Proxy] section")
	}
	return nil
}

type errString string

func (e errString) Error() string { return string(e) }

func detectTemplateType(template string) string {
	if strings.TrimSpace(template) == "" {
		return ""
	}
	surgePatterns := []string{"[General]", "[Proxy]", "[Proxy Group]", "[Rule]"}
	for _, pattern := range surgePatterns {
		if strings.Contains(template, pattern) {
			return "surge"
		}
	}
	clashPatterns := []string{"port:", "proxies:", "proxy-groups:", "rules:", "socks-port:", "dns:", "mode:"}
	for _, pattern := range clashPatterns {
		if strings.Contains(template, pattern) {
			return "clash"
		}
	}
	return ""
}

func PreviewProtectedGroupBehavior(category string, template string) []string {
	if category != "clash" {
		return nil
	}
	var warnings []string
	if strings.Contains(template, "__ALL_PROXIES__") {
		warnings = append(warnings, "Template contains __ALL_PROXIES__; it will be replaced with the actual node list at runtime")
	}
	if strings.Contains(template, "include-all") || strings.Contains(template, "include-all-proxies") || strings.Contains(template, "include-all-providers") {
		warnings = append(warnings, "Template contains include-all semantics; automatic group matching is preserved at runtime")
	}
	_ = protocol.Config{}
	return warnings
}
