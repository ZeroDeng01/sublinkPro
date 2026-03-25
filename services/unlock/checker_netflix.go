package unlock

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sublink/models"
)

type netflixUnlockChecker struct{}

func (netflixUnlockChecker) Key() string { return models.UnlockProviderNetflix }

func (netflixUnlockChecker) Aliases() []string { return []string{"netflix"} }

func (netflixUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	primary, err := fetchUnlockProbe(runtime, "https://www.netflix.com/title/80018499", nil)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusError, Reason: err.Error()}
	}
	if primary.StatusCode == http.StatusOK {
		if strings.Contains(primary.Body, "nsez-403") || strings.Contains(primary.FinalURL, "nsez-403") {
			return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusRestricted, Reason: "nsez_403"}
		}
		return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusAvailable, Region: extractNetflixRegion(primary.FinalURL)}
	}
	fallback, fallbackErr := fetchUnlockProbe(runtime, "https://www.netflix.com/title/81215567", nil)
	if fallbackErr != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusError, Reason: fallbackErr.Error()}
	}
	if fallback.StatusCode == http.StatusOK {
		return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusPartial, Region: extractNetflixRegion(fallback.FinalURL), Detail: "originals_only"}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderNetflix, Status: models.UnlockStatusRestricted, Reason: fmt.Sprintf("status_%d", primary.StatusCode)}
}

func extractNetflixRegion(rawURL string) string {
	parsed, err := url.Parse(rawURL)
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
