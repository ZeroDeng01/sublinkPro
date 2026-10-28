package unlock

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sublink/models"
)

func TestEvaluateGeminiUnlockProbe(t *testing.T) {
	runtime := UnlockRuntime{LandingCountry: "JP"}

	tests := []struct {
		name       string
		response   *unlockHTTPResponse
		wantStatus string
		wantReason string
	}{
		{
			name:       "available marker (escaped in inline JS)",
			response:   &unlockHTTPResponse{StatusCode: http.StatusOK, Body: `prefix \"45631641\":true suffix`},
			wantStatus: models.UnlockStatusAvailable,
		},
		{
			name:       "available marker (unescaped JSON)",
			response:   &unlockHTTPResponse{StatusCode: http.StatusOK, Body: `prefix "45631641":true suffix`},
			wantStatus: models.UnlockStatusAvailable,
		},
		{
			name:       "restricted body",
			response:   &unlockHTTPResponse{StatusCode: http.StatusOK, Body: "gemini isn't available in your country"},
			wantStatus: models.UnlockStatusRestricted,
			wantReason: "region_blocked",
		},
		{
			name:       "successful response without marker",
			response:   &unlockHTTPResponse{StatusCode: http.StatusOK, Body: "generic google shell"},
			wantStatus: models.UnlockStatusRestricted,
			wantReason: "gemini_marker_missing",
		},
		{
			name:       "forbidden",
			response:   &unlockHTTPResponse{StatusCode: http.StatusForbidden},
			wantStatus: models.UnlockStatusRestricted,
			wantReason: "status_403",
		},
		{
			name:       "server error",
			response:   &unlockHTTPResponse{StatusCode: http.StatusInternalServerError},
			wantStatus: models.UnlockStatusUnknown,
			wantReason: "status_500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateGeminiUnlockProbe(runtime, tt.response)
			if result.Provider != models.UnlockProviderGemini {
				t.Fatalf("provider = %q, want %q", result.Provider, models.UnlockProviderGemini)
			}
			if result.Region != runtime.LandingCountry {
				t.Fatalf("region = %q, want %q", result.Region, runtime.LandingCountry)
			}
			if result.Status != tt.wantStatus {
				t.Fatalf("status = %q, want %q", result.Status, tt.wantStatus)
			}
			if result.Reason != tt.wantReason {
				t.Fatalf("reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestFetchUnlockProbeWithBodyLimit(t *testing.T) {
	marker := "45631641,null,true"
	body := strings.Repeat("a", 40*1024) + marker
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	runtime := UnlockRuntime{Client: server.Client(), Timeout: time.Second}

	defaultResp, err := fetchUnlockProbe(runtime, server.URL, nil)
	if err != nil {
		t.Fatalf("fetchUnlockProbe returned error: %v", err)
	}
	if strings.Contains(defaultResp.Body, marker) {
		t.Fatalf("default probe unexpectedly read marker beyond 32 KiB limit")
	}

	largeResp, err := fetchUnlockProbeWithBodyLimit(runtime, server.URL, nil, 64*1024)
	if err != nil {
		t.Fatalf("fetchUnlockProbeWithBodyLimit returned error: %v", err)
	}
	if !strings.Contains(largeResp.Body, marker) {
		t.Fatalf("larger probe did not read marker beyond default limit")
	}
}
