package unlock

import (
	"fmt"
	"strings"
	"sublink/models"
)

var openAIComplianceHeaders = map[string]string{
	"Accept":        "*/*",
	"Authorization": "Bearer null",
	"Content-Type":  "application/json",
	"Origin":        "https://platform.openai.com",
	"Referer":       "https://platform.openai.com/",
}

type openAIUnlockChecker struct{}

func (openAIUnlockChecker) Key() string { return models.UnlockProviderOpenAI }

func (openAIUnlockChecker) Aliases() []string { return []string{"openai", "chatgpt"} }

func (openAIUnlockChecker) Meta() models.UnlockProviderMeta {
	return models.UnlockProviderMeta{Value: models.UnlockProviderOpenAI, Label: "OpenAI", Description: "Checks OpenAI / ChatGPT regional availability", Category: "ai"}
}

func (openAIUnlockChecker) RenameVariableMeta() models.UnlockRenameVariableMeta {
	return models.UnlockRenameVariableMeta{Provider: models.UnlockProviderOpenAI}
}

func (openAIUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	complianceResp, err := fetchUnlockProbe(runtime, "https://api.openai.com/compliance/cookie_requirements", openAIComplianceHeaders)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderOpenAI, Status: models.UnlockStatusError, Region: runtime.LandingCountry, Reason: err.Error()}
	}
	iosResp, err := fetchUnlockProbe(runtime, "https://ios.chat.openai.com/", map[string]string{"Accept": "*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"})
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderOpenAI, Status: models.UnlockStatusError, Region: runtime.LandingCountry, Reason: err.Error()}
	}
	return evaluateOpenAIUnlockProbe(runtime, complianceResp, iosResp)
}

func evaluateOpenAIUnlockProbe(runtime UnlockRuntime, complianceResp *unlockHTTPResponse, iosResp *unlockHTTPResponse) models.UnlockProviderResult {
	complianceBlocked := strings.Contains(complianceResp.Body, "unsupported_country")
	iosBlocked := strings.Contains(iosResp.Body, "vpn")
	if !complianceBlocked && !iosBlocked {
		return models.UnlockProviderResult{Provider: models.UnlockProviderOpenAI, Status: models.UnlockStatusAvailable, Region: runtime.LandingCountry}
	}
	if complianceBlocked && iosBlocked {
		return models.UnlockProviderResult{Provider: models.UnlockProviderOpenAI, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "unsupported_country"}
	}
	if !complianceBlocked && iosBlocked {
		return models.UnlockProviderResult{Provider: models.UnlockProviderOpenAI, Status: models.UnlockStatusPartial, Region: runtime.LandingCountry, Detail: "web_only"}
	}
	if complianceBlocked && !iosBlocked {
		return models.UnlockProviderResult{Provider: models.UnlockProviderOpenAI, Status: models.UnlockStatusPartial, Region: runtime.LandingCountry, Detail: "mobile_app_only"}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderOpenAI, Status: models.UnlockStatusUnknown, Region: runtime.LandingCountry, Reason: fmt.Sprintf("status_%d_%d", complianceResp.StatusCode, iosResp.StatusCode)}
}

func init() {
	RegisterUnlockChecker(openAIUnlockChecker{})
}
