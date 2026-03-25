package unlock

import (
	"fmt"
	"sublink/models"
)

type youTubePremiumUnlockChecker struct{}

func (youTubePremiumUnlockChecker) Key() string { return models.UnlockProviderYouTube }

func (youTubePremiumUnlockChecker) Aliases() []string {
	return []string{"youtube", "youtube_premium", "ytpremium", "youtubepremium"}
}

func (youTubePremiumUnlockChecker) Meta() models.UnlockProviderMeta {
	return models.UnlockProviderMeta{Value: models.UnlockProviderYouTube, Label: "YouTube Premium", Description: "检测 YouTube Premium 是否属于支持地区", Category: "streaming"}
}

func (youTubePremiumUnlockChecker) RenameVariableMeta() models.UnlockRenameVariableMeta {
	return models.UnlockRenameVariableMeta{Provider: models.UnlockProviderYouTube}
}

func (youTubePremiumUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	resp, err := fetchUnlockProbe(runtime, "https://www.youtube.com/premium", nil)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderYouTube, Status: models.UnlockStatusError, Region: runtime.LandingCountry, Reason: err.Error()}
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		if containsAny(resp.Body, []string{"youtube premium is not available in your country"}) {
			return models.UnlockProviderResult{Provider: models.UnlockProviderYouTube, Status: models.UnlockStatusUnsupported, Region: runtime.LandingCountry, Reason: "unsupported_country"}
		}
		return models.UnlockProviderResult{Provider: models.UnlockProviderYouTube, Status: models.UnlockStatusReachable, Region: runtime.LandingCountry}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderYouTube, Status: models.UnlockStatusUnknown, Region: runtime.LandingCountry, Reason: fmt.Sprintf("status_%d", resp.StatusCode)}
}

func init() {
	RegisterUnlockChecker(youTubePremiumUnlockChecker{})
}
