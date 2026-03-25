package unlock

import (
	"fmt"
	"net/http"
	"sublink/models"
)

type geminiUnlockChecker struct{}

func (geminiUnlockChecker) Key() string { return models.UnlockProviderGemini }

func (geminiUnlockChecker) Aliases() []string { return []string{"gemini"} }

func (geminiUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	if runtime.LandingCountry == "CN" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderGemini, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "workspace_only_region"}
	}
	resp, err := fetchUnlockProbe(runtime, "https://gemini.google.com/", nil)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderGemini, Status: models.UnlockStatusError, Region: runtime.LandingCountry, Reason: err.Error()}
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return models.UnlockProviderResult{Provider: models.UnlockProviderGemini, Status: models.UnlockStatusReachable, Region: runtime.LandingCountry}
	}
	if resp.StatusCode == http.StatusForbidden {
		return models.UnlockProviderResult{Provider: models.UnlockProviderGemini, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "status_403"}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderGemini, Status: models.UnlockStatusUnknown, Region: runtime.LandingCountry, Reason: fmt.Sprintf("status_%d", resp.StatusCode)}
}

func init() {
	RegisterUnlockChecker(geminiUnlockChecker{})
}
