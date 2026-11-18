package unlock

import (
	"fmt"
	"net/http"
	"strings"
	"sublink/models"
)

type claudeUnlockChecker struct{}

func (claudeUnlockChecker) Key() string { return models.UnlockProviderClaude }

func (claudeUnlockChecker) Aliases() []string { return []string{"claude"} }

func (claudeUnlockChecker) Meta() models.UnlockProviderMeta {
	return models.UnlockProviderMeta{Value: models.UnlockProviderClaude, Label: "Claude", Description: "Checks Anthropic Claude regional availability", Category: "ai"}
}

func (claudeUnlockChecker) RenameVariableMeta() models.UnlockRenameVariableMeta {
	return models.UnlockRenameVariableMeta{Provider: models.UnlockProviderClaude}
}

func (claudeUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	resp, err := fetchUnlockProbe(runtime, "https://claude.ai/", nil)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderClaude, Status: models.UnlockStatusError, Region: runtime.LandingCountry, Reason: err.Error()}
	}
	return evaluateClaudeUnlockProbe(runtime, resp)
}

func evaluateClaudeUnlockProbe(runtime UnlockRuntime, resp *unlockHTTPResponse) models.UnlockProviderResult {
	finalURL := strings.TrimRight(strings.ToLower(strings.TrimSpace(resp.FinalURL)), "/")
	if finalURL == "https://claude.ai" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderClaude, Status: models.UnlockStatusAvailable, Region: runtime.LandingCountry}
	}
	if finalURL == "https://www.anthropic.com/app-unavailable-in-region" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderClaude, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "app_unavailable_in_region"}
	}
	if resp.StatusCode == http.StatusForbidden {
		return models.UnlockProviderResult{Provider: models.UnlockProviderClaude, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "status_403"}
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return models.UnlockProviderResult{Provider: models.UnlockProviderClaude, Status: models.UnlockStatusUnknown, Region: runtime.LandingCountry, Reason: "unexpected_final_url", Detail: resp.FinalURL}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderClaude, Status: models.UnlockStatusUnknown, Region: runtime.LandingCountry, Reason: fmt.Sprintf("status_%d", resp.StatusCode)}
}
func init() {
	RegisterUnlockChecker(claudeUnlockChecker{})
}
