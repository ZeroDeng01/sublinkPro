package unlock

import (
	"fmt"
	"net/http"
	"strings"
	"sublink/models"
)

const geminiUnlockProbeBodyLimit = 1024 * 1024

var geminiRestrictedMarkers = []string{
	"gemini isn't available in your country",
	"gemini is not available in your country",
	"gemini isn't available in your region",
	"gemini is not available in your region",
	"not available in your country",
	"not available in your region",
}

type geminiUnlockChecker struct{}

func (geminiUnlockChecker) Key() string { return models.UnlockProviderGemini }

func (geminiUnlockChecker) Aliases() []string { return []string{"gemini"} }

func (geminiUnlockChecker) Meta() models.UnlockProviderMeta {
	return models.UnlockProviderMeta{Value: models.UnlockProviderGemini, Label: "Gemini", Description: "检测 Google Gemini 服务地区可访问性", Category: "ai"}
}

func (geminiUnlockChecker) RenameVariableMeta() models.UnlockRenameVariableMeta {
	return models.UnlockRenameVariableMeta{Provider: models.UnlockProviderGemini}
}

func (geminiUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	if runtime.LandingCountry == "CN" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderGemini, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "workspace_only_region"}
	}
	resp, err := fetchUnlockProbeWithBodyLimit(runtime, "https://gemini.google.com/", nil, geminiUnlockProbeBodyLimit)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderGemini, Status: models.UnlockStatusError, Region: runtime.LandingCountry, Reason: err.Error()}
	}
	return evaluateGeminiUnlockProbe(runtime, resp)
}

func evaluateGeminiUnlockProbe(runtime UnlockRuntime, resp *unlockHTTPResponse) models.UnlockProviderResult {
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		if strings.Contains(resp.Body, "\"45631641\\\":true") {
			return models.UnlockProviderResult{Provider: models.UnlockProviderGemini, Status: models.UnlockStatusAvailable, Region: runtime.LandingCountry}
		}
		if containsAny(resp.Body, geminiRestrictedMarkers) {
			return models.UnlockProviderResult{Provider: models.UnlockProviderGemini, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "region_blocked"}
		}
		return models.UnlockProviderResult{Provider: models.UnlockProviderGemini, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "gemini_marker_missing"}
	}
	if resp.StatusCode == http.StatusForbidden {
		return models.UnlockProviderResult{Provider: models.UnlockProviderGemini, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "status_403"}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderGemini, Status: models.UnlockStatusUnknown, Region: runtime.LandingCountry, Reason: fmt.Sprintf("status_%d", resp.StatusCode)}
}

func init() {
	RegisterUnlockChecker(geminiUnlockChecker{})
}
