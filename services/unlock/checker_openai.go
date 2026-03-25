package unlock

import (
	"fmt"
	"net/http"
	"sublink/models"
)

type openAIUnlockChecker struct{}

func (openAIUnlockChecker) Key() string { return models.UnlockProviderOpenAI }

func (openAIUnlockChecker) Aliases() []string { return []string{"openai", "chatgpt"} }

func (openAIUnlockChecker) Meta() models.UnlockProviderMeta {
	return models.UnlockProviderMeta{Value: models.UnlockProviderOpenAI, Label: "OpenAI", Description: "检测 OpenAI / ChatGPT 服务地区可访问性", Category: "ai"}
}

func (openAIUnlockChecker) RenameVariableMeta() models.UnlockRenameVariableMeta {
	return models.UnlockRenameVariableMeta{Provider: models.UnlockProviderOpenAI}
}

func (openAIUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	resp, err := fetchUnlockProbe(runtime, "https://chatgpt.com/", nil)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderOpenAI, Status: models.UnlockStatusError, Region: runtime.LandingCountry, Reason: err.Error()}
	}
	if containsAny(resp.Body, []string{"not available in your country", "unsupported country"}) {
		return models.UnlockProviderResult{Provider: models.UnlockProviderOpenAI, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "unsupported_country"}
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return models.UnlockProviderResult{Provider: models.UnlockProviderOpenAI, Status: models.UnlockStatusReachable, Region: runtime.LandingCountry}
	}
	if resp.StatusCode == http.StatusForbidden {
		return models.UnlockProviderResult{Provider: models.UnlockProviderOpenAI, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "status_403"}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderOpenAI, Status: models.UnlockStatusUnknown, Region: runtime.LandingCountry, Reason: fmt.Sprintf("status_%d", resp.StatusCode)}
}

func init() {
	RegisterUnlockChecker(openAIUnlockChecker{})
}
