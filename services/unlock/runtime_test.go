package unlock

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sublink/internal/testutil"
)

func TestUnlockRunClosesPoolWithoutClosingBorrowedAdapter(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "{}")
	}))
	connections := testutil.TrackHTTPConnections(server)
	server.StartTLS()
	t.Cleanup(server.Close)
	proxy := testutil.NewTrackingProxy(t, server.Listener.Addr().String())
	for range 3 {
		result := CheckUnlockWithAdapter(proxy, time.Second, "US", []string{"netflix"})
		if len(result.Providers) != 1 {
			t.Fatalf("unexpected result: %+v", result)
		}
		connections.WaitForClosed(t)
	}
	if got := proxy.CloseCount.Load(); got != 0 {
		t.Fatalf("borrowed adapter closed %d times", got)
	}
}
