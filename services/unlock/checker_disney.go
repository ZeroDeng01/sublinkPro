package unlock

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sublink/models"
)

const disneyBrowserBearer = "Bearer ZGlzbmV5JmJyb3dzZXImMS4wLjA.Cu56AgSfBTDag5NiRA81oLHkDZfu5L3CKadnefEAY84"
const disneyGraphAuthorization = "ZGlzbmV5JmJyb3dzZXImMS4wLjA.Cu56AgSfBTDag5NiRA81oLHkDZfu5L3CKadnefEAY84"
const disneyDevicePayload = `{"deviceFamily":"browser","applicationRuntime":"chrome","deviceProfile":"windows","attributes":{}}`
const disneyTokenPayloadTemplate = "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Atoken-exchange&latitude=0&longitude=0&platform=browser&subject_token=DISNEYASSERTION&subject_token_type=urn%3Abamtech%3Aparams%3Aoauth%3Atoken-type%3Adevice"
const disneyGraphPayloadTemplate = `{"query":"mutation refreshToken($input: RefreshTokenInput!) {\n            refreshToken(refreshToken: $input) {\n                activeSession {\n                    sessionId\n                }\n            }\n        }","variables":{"input":{"refreshToken":"ILOVEDISNEY"}}}`

var disneyAssertionPattern = regexp.MustCompile(`(?i)"assertion"\s*:\s*"([^"]+)"`)
var disneyRefreshTokenPattern = regexp.MustCompile(`(?i)"refresh_token"\s*:\s*"([^"]+)"`)
var disneyCountryPattern = regexp.MustCompile(`(?i)"countryCode"\s*:\s*"([^"]+)"`)
var disneySupportedLocationPattern = regexp.MustCompile(`(?i)"inSupportedLocation"\s*:\s*(false|true)`)

type disneyUnlockChecker struct{}

func (disneyUnlockChecker) Key() string { return models.UnlockProviderDisney }

func (disneyUnlockChecker) Aliases() []string {
	return []string{"disney", "disney+", "disneyplus", "disney_plus"}
}

func (disneyUnlockChecker) Meta() models.UnlockProviderMeta {
	return models.UnlockProviderMeta{Value: models.UnlockProviderDisney, Label: "Disney+", Description: "Checks whether the Disney+ entry point is reachable and clearly unrestricted", Category: "streaming"}
}

func (disneyUnlockChecker) RenameVariableMeta() models.UnlockRenameVariableMeta {
	return models.UnlockRenameVariableMeta{Provider: models.UnlockProviderDisney}
}

func (disneyUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	deviceResp, err := fetchUnlockRequest(runtime, http.MethodPost, "https://disney.api.edge.bamgrid.com/devices", map[string]string{
		"Authorization": disneyBrowserBearer,
		"Content-Type":  "application/json; charset=UTF-8",
	}, []byte(disneyDevicePayload), 64*1024)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusError, Reason: err.Error()}
	}
	assertion := extractMatch(disneyAssertionPattern, deviceResp.RawBody)
	if assertion == "" {
		return evaluateDisneyUnlockProbe(deviceResp, nil, nil, nil)
	}
	tokenPayload := strings.ReplaceAll(disneyTokenPayloadTemplate, "DISNEYASSERTION", assertion)
	tokenResp, err := fetchUnlockRequest(runtime, http.MethodPost, "https://disney.api.edge.bamgrid.com/token", map[string]string{
		"Authorization": disneyBrowserBearer,
		"Content-Type":  "application/x-www-form-urlencoded",
	}, []byte(tokenPayload), 128*1024)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusError, Reason: err.Error()}
	}
	refreshToken := extractMatch(disneyRefreshTokenPattern, tokenResp.RawBody)
	if refreshToken == "" {
		return evaluateDisneyUnlockProbe(deviceResp, tokenResp, nil, nil)
	}
	graphPayload := strings.ReplaceAll(disneyGraphPayloadTemplate, "ILOVEDISNEY", refreshToken)
	graphResp, err := fetchUnlockRequest(runtime, http.MethodPost, "https://disney.api.edge.bamgrid.com/graph/v1/device/graphql", map[string]string{
		"Authorization": disneyGraphAuthorization,
		"Content-Type":  "application/json",
	}, []byte(graphPayload), 128*1024)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusError, Reason: err.Error()}
	}
	previewResp, err := fetchUnlockProbe(runtime, "https://disneyplus.com", nil)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusError, Reason: err.Error()}
	}
	return evaluateDisneyUnlockProbe(deviceResp, tokenResp, graphResp, previewResp)
}

func evaluateDisneyUnlockProbe(deviceResp *unlockHTTPResponse, tokenResp *unlockHTTPResponse, graphResp *unlockHTTPResponse, previewResp *unlockHTTPResponse) models.UnlockProviderResult {
	if deviceResp == nil || deviceResp.RawBody == "" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusError, Reason: "network_connection"}
	}
	if strings.Contains(deviceResp.Body, "403 error") {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusRestricted, Reason: "ip_banned"}
	}
	if extractMatch(disneyAssertionPattern, deviceResp.RawBody) == "" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusUnknown, Reason: "assertion_missing"}
	}
	if tokenResp == nil || tokenResp.RawBody == "" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusError, Reason: "network_connection"}
	}
	if strings.Contains(tokenResp.Body, "forbidden-location") || strings.Contains(tokenResp.Body, "403 error") {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusRestricted, Reason: "ip_banned"}
	}
	if extractMatch(disneyRefreshTokenPattern, tokenResp.RawBody) == "" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusUnknown, Reason: "refresh_token_missing"}
	}
	if graphResp == nil || graphResp.RawBody == "" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusError, Reason: "network_connection"}
	}
	region := strings.ToUpper(strings.TrimSpace(extractMatch(disneyCountryPattern, graphResp.RawBody)))
	inSupportedLocation := strings.ToLower(extractMatch(disneySupportedLocationPattern, graphResp.RawBody))
	if region == "" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusRestricted, Reason: "region_missing"}
	}
	if region == "JP" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusAvailable, Region: region}
	}
	if previewResp != nil && (strings.Contains(previewResp.FinalURL, "preview") || strings.Contains(previewResp.FinalURL, "unavailable")) {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusRestricted, Region: region, Reason: "preview_unavailable"}
	}
	if inSupportedLocation == "false" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusPartial, Region: region, Detail: "available_soon"}
	}
	if inSupportedLocation == "true" {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusAvailable, Region: region}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusUnknown, Region: region, Reason: fmt.Sprintf("supported_location_%s", inSupportedLocation)}
}

func extractMatch(pattern *regexp.Regexp, text string) string {
	matches := pattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func init() {
	RegisterUnlockChecker(disneyUnlockChecker{})
}
