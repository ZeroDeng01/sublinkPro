package node

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"sublink/internal/testutil"
	"sublink/models"
)

func TestFetchAirportUsageReleasesConnections(t *testing.T) {
	for _, fallback := range []bool{false, true} {
		name := "head"
		if fallback {
			name = "get-fallback"
		}
		t.Run(name, func(t *testing.T) {
			var requests atomic.Int32
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if fallback && r.Method == http.MethodHead {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.Header().Set("subscription-userinfo", "upload=1; download=2; total=1000; expire=2000000000")
				w.WriteHeader(http.StatusNoContent)
			}))
			connections := testutil.TrackHTTPConnections(server)
			server.Start()
			t.Cleanup(server.Close)
			for range 3 {
				usage, err := FetchAirportUsageInfo(&models.Airport{URL: server.URL, FetchUsageInfo: true})
				if err != nil || usage == nil || usage.Total != 1000 {
					t.Fatalf("unexpected usage: %+v, %v", usage, err)
				}
				connections.WaitForClosed(t)
			}
			wantRequests := int32(3)
			if fallback {
				wantRequests = 6
			}
			if requests.Load() != wantRequests {
				t.Fatalf("requests = %d, want %d", requests.Load(), wantRequests)
			}
		})
	}
}
