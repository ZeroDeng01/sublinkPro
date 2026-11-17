package unlock

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sublink/models"
)

const netflixProbeBodyLimit = 512 * 1024

var netflixPageRegionPattern = regexp.MustCompile(`(?s)"id":"([A-Za-z]{2})".*"countryName":"[^"]*"`)

type netflixUnlockChecker struct{}

func (netflixUnlockChecker) Key() string { return models.UnlockProviderNetflix }

func (netflixUnlockChecker) Aliases() []string { return []string{"netflix"} }

func (netflixUnlockChecker) Meta() models.UnlockProviderMeta {
	return models.UnlockProviderMeta{Value: models.UnlockProviderNetflix, Label: "Netflix", Description: "Checks whether the full catalog or only Originals is supported", Category: "streaming"}
}

func (netflixUnlockChecker) RenameVariableMeta() models.UnlockRenameVariableMeta {
	return models.UnlockRenameVariableMeta{Provider: models.UnlockProviderNetflix}
}

func (netflixUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	primary, err := fetchUnlockProbeWithBodyLimit(runtime, "https://www.netflix.com/title/81280792", netflixHeaders(), netflixProbeBodyLimit)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusError, Reason: err.Error()}
	}
	fallback, fallbackErr := fetchUnlockProbeWithBodyLimit(runtime, "https://www.netflix.com/title/70143836", netflixHeaders(), netflixProbeBodyLimit)
	if fallbackErr != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusError, Reason: fallbackErr.Error()}
	}
	return evaluateNetflixUnlockProbe(primary, fallback)
}

func evaluateNetflixUnlockProbe(primary *unlockHTTPResponse, fallback *unlockHTTPResponse) models.UnlockProviderResult {
	if strings.Contains(primary.Body, "nsez-403") || strings.Contains(primary.FinalURL, "nsez-403") || strings.Contains(fallback.Body, "nsez-403") || strings.Contains(fallback.FinalURL, "nsez-403") {
		return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusRestricted, Reason: "nsez_403"}
	}
	if primary.StatusCode >= http.StatusBadRequest || fallback.StatusCode >= http.StatusBadRequest {
		return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusRestricted, Reason: fmt.Sprintf("status_%d_%d", primary.StatusCode, fallback.StatusCode)}
	}
	primaryBlocked := strings.Contains(primary.Body, "oh no!")
	fallbackBlocked := strings.Contains(fallback.Body, "oh no!")
	region := extractNetflixRegion(primary)
	if primaryBlocked && fallbackBlocked {
		return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusPartial, Region: region, Detail: "originals_only"}
	}
	if !primaryBlocked || !fallbackBlocked {
		return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusAvailable, Region: region}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusUnknown, Region: region, Reason: "page_marker_missing"}
}

func netflixHeaders() map[string]string {
	return map[string]string{"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"}
}

func extractNetflixRegion(resp *unlockHTTPResponse) string {
	if matches := netflixPageRegionPattern.FindStringSubmatch(resp.RawBody); len(matches) >= 2 {
		return strings.ToUpper(strings.TrimSpace(matches[1]))
	}
	parsed, err := url.Parse(resp.FinalURL)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) > 0 {
		candidate := strings.ToUpper(strings.TrimSpace(parts[0]))
		if len(candidate) == 2 {
			return candidate
		}
	}
	return ""
}

func init() {
	RegisterUnlockChecker(netflixUnlockChecker{})
}
