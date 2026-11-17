package unlock

import (
	"fmt"
	"regexp"
	"strings"
	"sublink/models"
)

var youtubePremiumRegionPattern = regexp.MustCompile(`(?i)"INNERTUBE_CONTEXT_GL"\s*:\s*"([^"]+)"`)

const youtubePremiumProbeBodyLimit = 1024 * 1024
const youtubePremiumCookie = "YSC=FSCWhKo2Zgw; VISITOR_PRIVACY_METADATA=CgJERRIEEgAgYQ%3D%3D; PREF=f7=4000; __Secure-YEC=CgtRWTBGTFExeV9Iayjele2yBjIKCgJERRIEEgAgYQ%3D%3D; SOCS=CAISOAgDEitib3FfaWRlbnRpZnlmcm9udGVuZHVpc2VydmVyXzIwMjQwNTI2LjAxX3AwGgV6aC1DTiACGgYIgMnpsgY; VISITOR_INFO1_LIVE=Di84mAIbgKY; __Secure-BUCKET=CGQ"

type youTubePremiumUnlockChecker struct{}

func (youTubePremiumUnlockChecker) Key() string { return models.UnlockProviderYouTube }

func (youTubePremiumUnlockChecker) Aliases() []string {
	return []string{"youtube", "youtube_premium", "ytpremium", "youtubepremium"}
}

func (youTubePremiumUnlockChecker) Meta() models.UnlockProviderMeta {
	return models.UnlockProviderMeta{Value: models.UnlockProviderYouTube, Label: "YouTube Premium", Description: "Checks whether YouTube Premium is available in the exit region", Category: "streaming"}
}

func (youTubePremiumUnlockChecker) RenameVariableMeta() models.UnlockRenameVariableMeta {
	return models.UnlockRenameVariableMeta{Provider: models.UnlockProviderYouTube}
}

func (youTubePremiumUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	resp, err := fetchUnlockProbeWithBodyLimit(runtime, "https://www.youtube.com/premium", map[string]string{"Cookie": youtubePremiumCookie}, youtubePremiumProbeBodyLimit)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderYouTube, Status: models.UnlockStatusError, Region: runtime.LandingCountry, Reason: err.Error()}
	}
	return evaluateYouTubePremiumUnlockProbe(runtime, resp)
}

func evaluateYouTubePremiumUnlockProbe(runtime UnlockRuntime, resp *unlockHTTPResponse) models.UnlockProviderResult {
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		if strings.Contains(resp.Body, "www.google.cn") {
			return models.UnlockProviderResult{Provider: models.UnlockProviderYouTube, Status: models.UnlockStatusRestricted, Region: "CN", Reason: "google_cn"}
		}
		if strings.Contains(resp.Body, "premium is not available in your country") {
			return models.UnlockProviderResult{Provider: models.UnlockProviderYouTube, Status: models.UnlockStatusUnsupported, Region: runtime.LandingCountry, Reason: "unsupported_country"}
		}
		region := extractYouTubePremiumRegion(resp.RawBody)
		if strings.Contains(resp.Body, "ad-free") {
			return models.UnlockProviderResult{Provider: models.UnlockProviderYouTube, Status: models.UnlockStatusAvailable, Region: region}
		}
		return models.UnlockProviderResult{Provider: models.UnlockProviderYouTube, Status: models.UnlockStatusUnknown, Region: region, Reason: "page_marker_missing"}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderYouTube, Status: models.UnlockStatusUnknown, Region: runtime.LandingCountry, Reason: fmt.Sprintf("status_%d", resp.StatusCode)}
}

func extractYouTubePremiumRegion(body string) string {
	matches := youtubePremiumRegionPattern.FindStringSubmatch(body)
	if len(matches) < 2 {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(matches[1]))
}

func init() {
	RegisterUnlockChecker(youTubePremiumUnlockChecker{})
}
