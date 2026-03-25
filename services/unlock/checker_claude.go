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
	return models.UnlockProviderMeta{Value: models.UnlockProviderClaude, Label: "Claude", Description: "检测 Anthropic Claude 服务地区可访问性", Category: "ai"}
}

func (claudeUnlockChecker) RenameVariableMeta() models.UnlockRenameVariableMeta {
	return models.UnlockRenameVariableMeta{Provider: models.UnlockProviderClaude}
}

func (claudeUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	if runtime.LandingCountry != "" && !isClaudeSupportedCountry(runtime.LandingCountry) {
		return models.UnlockProviderResult{Provider: models.UnlockProviderClaude, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "unsupported_country"}
	}
	resp, err := fetchUnlockProbe(runtime, "https://claude.ai/", nil)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderClaude, Status: models.UnlockStatusError, Region: runtime.LandingCountry, Reason: err.Error()}
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return models.UnlockProviderResult{Provider: models.UnlockProviderClaude, Status: models.UnlockStatusReachable, Region: runtime.LandingCountry}
	}
	if resp.StatusCode == http.StatusForbidden {
		return models.UnlockProviderResult{Provider: models.UnlockProviderClaude, Status: models.UnlockStatusRestricted, Region: runtime.LandingCountry, Reason: "status_403"}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderClaude, Status: models.UnlockStatusUnknown, Region: runtime.LandingCountry, Reason: fmt.Sprintf("status_%d", resp.StatusCode)}
}

var claudeSupportedCountries = map[string]struct{}{
	"AL": {}, "DZ": {}, "AD": {}, "AO": {}, "AG": {}, "AR": {}, "AM": {}, "AU": {}, "AT": {}, "AZ": {}, "BS": {}, "BH": {}, "BD": {}, "BB": {}, "BE": {}, "BZ": {}, "BJ": {}, "BT": {}, "BO": {}, "BA": {}, "BW": {}, "BR": {}, "BN": {}, "BG": {}, "BF": {}, "BI": {}, "CV": {}, "KH": {}, "CM": {}, "CA": {}, "TD": {}, "CL": {}, "CO": {}, "KM": {}, "CG": {}, "CR": {}, "CI": {}, "HR": {}, "CY": {}, "CZ": {}, "DK": {}, "DJ": {}, "DM": {}, "DO": {}, "EC": {}, "EG": {}, "SV": {}, "GQ": {}, "EE": {}, "SZ": {}, "FJ": {}, "FI": {}, "FR": {}, "GA": {}, "GM": {}, "GE": {}, "DE": {}, "GH": {}, "GR": {}, "GD": {}, "GT": {}, "GN": {}, "GW": {}, "GY": {}, "HT": {}, "HN": {}, "HU": {}, "IS": {}, "IN": {}, "ID": {}, "IQ": {}, "IE": {}, "IL": {}, "IT": {}, "JM": {}, "JP": {}, "JO": {}, "KZ": {}, "KE": {}, "KI": {}, "KW": {}, "KG": {}, "LA": {}, "LV": {}, "LB": {}, "LS": {}, "LR": {}, "LI": {}, "LT": {}, "LU": {}, "MG": {}, "MW": {}, "MY": {}, "MV": {}, "MT": {}, "MH": {}, "MR": {}, "MU": {}, "MX": {}, "FM": {}, "MD": {}, "MC": {}, "MN": {}, "ME": {}, "MA": {}, "MZ": {}, "NA": {}, "NR": {}, "NP": {}, "NL": {}, "NZ": {}, "NE": {}, "NG": {}, "MK": {}, "NO": {}, "OM": {}, "PK": {}, "PW": {}, "PS": {}, "PA": {}, "PG": {}, "PY": {}, "PE": {}, "PH": {}, "PL": {}, "PT": {}, "QA": {}, "RO": {}, "RW": {}, "KN": {}, "LC": {}, "VC": {}, "WS": {}, "SM": {}, "ST": {}, "SA": {}, "SN": {}, "RS": {}, "SC": {}, "SL": {}, "SG": {}, "SK": {}, "SI": {}, "SB": {}, "ZA": {}, "KR": {}, "ES": {}, "LK": {}, "SR": {}, "SE": {}, "CH": {}, "TW": {}, "TJ": {}, "TZ": {}, "TH": {}, "TL": {}, "TG": {}, "TO": {}, "TT": {}, "TN": {}, "TR": {}, "TM": {}, "TV": {}, "UG": {}, "UA": {}, "AE": {}, "GB": {}, "US": {}, "UY": {}, "UZ": {}, "VU": {}, "VA": {}, "VN": {}, "ZM": {}, "ZW": {},
}

func isClaudeSupportedCountry(country string) bool {
	_, exists := claudeSupportedCountries[strings.ToUpper(strings.TrimSpace(country))]
	return exists
}

func init() {
	RegisterUnlockChecker(claudeUnlockChecker{})
}
