package unlock

import (
	"net/http"
	"testing"

	"sublink/models"
)

func TestEvaluateBahamutDeviceID(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		wantDeviceID    string
		wantUnsupported bool
	}{
		{name: "valid device ID", body: `{"deviceid":"abc123","animeSn":0}`, wantDeviceID: "abc123"},
		{name: "empty device ID", body: `{"deviceid":"","animeSn":0}`, wantDeviceID: ""},
		{name: "HTML response - unsupported", body: `<html><body>not available</body></html>`, wantUnsupported: true},
		{name: "garbage JSON - not unsupported", body: `not json at all`, wantDeviceID: "", wantUnsupported: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deviceID, unsupported := evaluateBahamutDeviceID(tt.body)
			if deviceID != tt.wantDeviceID {
				t.Fatalf("deviceID = %q, want %q", deviceID, tt.wantDeviceID)
			}
			if unsupported != tt.wantUnsupported {
				t.Fatalf("unsupported = %v, want %v", unsupported, tt.wantUnsupported)
			}
		})
	}
}

func TestEvaluateBahamutToken(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantAvail bool
	}{
		{name: "token available", body: `{"animeSn":37783}`, wantAvail: true},
		{name: "token unavailable", body: `{"animeSn":0}`, wantAvail: false},
		{name: "garbage JSON", body: `not json`, wantAvail: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := evaluateBahamutToken(tt.body); got != tt.wantAvail {
				t.Fatalf("evaluateBahamutToken() = %v, want %v", got, tt.wantAvail)
			}
		})
	}
}

func TestEvaluateBahamutRegion(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantRegion string
	}{
		{name: "TW region", body: "ip=1.2.3.4\nloc=TW\nvisit_scheme=https\n", wantRegion: "TW"},
		{name: "HK region", body: "ip=1.2.3.4\nloc=HK\nvisit_scheme=https\n", wantRegion: "HK"},
		{name: "lowercase normalized to upper", body: "ip=1.2.3.4\nloc=tw\n", wantRegion: "TW"},
		{name: "no loc field", body: "ip=1.2.3.4\nvisit_scheme=https\n", wantRegion: ""},
		{name: "empty body", body: "", wantRegion: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := evaluateBahamutRegion(tt.body); got != tt.wantRegion {
				t.Fatalf("region = %q, want %q", got, tt.wantRegion)
			}
		})
	}
}

// TestEvaluateBahamutUnlockProbe 覆盖 Check 的所有判定分支。
// 参数顺序：deviceIDResp, lenientResp(37783), strictResp(38832), traceResp。
// 判定语义参考 MediaUnlockTest/checks/BahamutAnime.go：
//   - 入口探针(37783)通过是 Available 的前提
//   - 台湾探针(38832)通过 → Available TW
//   - 38832 不通过 + trace loc → Available loc（港澳台）
//   - 37783 不通过 → Unsupported
func TestEvaluateBahamutUnlockProbe(t *testing.T) {
	runtime := UnlockRuntime{LandingCountry: "US"}
	deviceIDOK := unlockTestResponse(http.StatusOK, "", `{"deviceid":"dev123","animeSn":0}`)
	lenientOK := unlockTestResponse(http.StatusOK, "", `{"animeSn":37783}`)
	strictOK := unlockTestResponse(http.StatusOK, "", `{"animeSn":38832}`)
	strictFail := unlockTestResponse(http.StatusOK, "", `{"animeSn":0}`)

	tests := []struct {
		name       string
		deviceID   *unlockHTTPResponse
		lenient    *unlockHTTPResponse
		strict     *unlockHTTPResponse
		trace      *unlockHTTPResponse
		wantStatus string
		wantRegion string
		wantReason string
		wantDetail string
	}{
		{
			name:       "HTML device ID - unsupported",
			deviceID:   unlockTestResponse(http.StatusOK, "", `<html>blocked</html>`),
			wantStatus: models.UnlockStatusUnsupported,
			wantRegion: "US",
			wantReason: "not_available",
		},
		{
			name:       "lenient fails - unsupported",
			deviceID:   deviceIDOK,
			lenient:    unlockTestResponse(http.StatusOK, "", `{"animeSn":0}`),
			wantStatus: models.UnlockStatusUnsupported,
			wantRegion: "US",
			wantReason: "region_unavailable",
		},
		{
			name:       "lenient + strict pass - TW full unlock",
			deviceID:   deviceIDOK,
			lenient:    lenientOK,
			strict:     strictOK,
			wantStatus: models.UnlockStatusAvailable,
			wantRegion: "TW",
		},
		{
			name:       "lenient pass + strict fail + HK trace - HK available",
			deviceID:   deviceIDOK,
			lenient:    lenientOK,
			strict:     strictFail,
			trace:      unlockTestResponse(http.StatusOK, "", "ip=1.2.3.4\nloc=HK\n"),
			wantStatus: models.UnlockStatusAvailable,
			wantRegion: "HK",
		},
		{
			name:       "lenient pass + strict fail + MO trace - MO available",
			deviceID:   deviceIDOK,
			lenient:    lenientOK,
			strict:     strictFail,
			trace:      unlockTestResponse(http.StatusOK, "", "ip=1.2.3.4\nloc=MO\n"),
			wantStatus: models.UnlockStatusAvailable,
			wantRegion: "MO",
		},
		{
			name:       "lenient pass + strict fail + no trace - partial fallback",
			deviceID:   deviceIDOK,
			lenient:    lenientOK,
			strict:     strictFail,
			trace:      nil,
			wantStatus: models.UnlockStatusPartial,
			wantRegion: "US",
			wantDetail: "primary_token_only",
		},
		{
			name:       "lenient pass + strict fail + invalid trace loc - partial fallback",
			deviceID:   deviceIDOK,
			lenient:    lenientOK,
			strict:     strictFail,
			trace:      unlockTestResponse(http.StatusOK, "", "ip=1.2.3.4\nloc=INVALID\n"),
			wantStatus: models.UnlockStatusPartial,
			wantRegion: "US",
			wantDetail: "primary_token_only",
		},
		{
			name:       "lenient pass + strict fail + malformed trace - partial fallback",
			deviceID:   deviceIDOK,
			lenient:    lenientOK,
			strict:     strictFail,
			trace:      unlockTestResponse(http.StatusOK, "", "garbage response"),
			wantStatus: models.UnlockStatusPartial,
			wantRegion: "US",
			wantDetail: "primary_token_only",
		},
		{
			name:       "empty device ID response - error",
			deviceID:   unlockTestResponse(http.StatusOK, "", ""),
			wantStatus: models.UnlockStatusError,
			wantRegion: "US",
			wantReason: "network_connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateBahamutUnlockProbe(runtime, tt.deviceID, tt.lenient, tt.strict, tt.trace)
			assertUnlockResult(t, result, models.UnlockProviderBahamut, tt.wantStatus, tt.wantRegion, tt.wantReason, tt.wantDetail)
		})
	}
}

func TestBahamutMeta(t *testing.T) {
	checker := bahamutAnimeUnlockChecker{}

	if checker.Key() != models.UnlockProviderBahamut {
		t.Fatalf("Key() = %q, want %q", checker.Key(), models.UnlockProviderBahamut)
	}

	meta := checker.Meta()
	if meta.Value != models.UnlockProviderBahamut {
		t.Fatalf("Meta().Value = %q, want %q", meta.Value, models.UnlockProviderBahamut)
	}
	if meta.Label == "" {
		t.Fatal("Meta().Label is empty")
	}
	if meta.Category != "streaming" {
		t.Fatalf("Meta().Category = %q, want %q", meta.Category, "streaming")
	}

	renameMeta := checker.RenameVariableMeta()
	if renameMeta.Provider != models.UnlockProviderBahamut {
		t.Fatalf("RenameVariableMeta().Provider = %q, want %q", renameMeta.Provider, models.UnlockProviderBahamut)
	}
}
