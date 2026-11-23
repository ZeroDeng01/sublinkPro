package utils

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"sublink/internal/testutil"

	"github.com/metacubex/mihomo/constant"
)

func TestFetchWithProxyReleasesOwnedResources(t *testing.T) {
	for _, scenario := range []string{"success", "status-error", "read-error", "invalid-url", "dial-error"} {
		t.Run(scenario, func(t *testing.T) {
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if scenario == "status-error" {
					w.WriteHeader(http.StatusBadGateway)
				}
				if scenario == "read-error" {
					w.Header().Set("Content-Length", "100")
				}
				_, _ = io.WriteString(w, "body")
			}))
			connections := testutil.TrackHTTPConnections(server)
			server.Start()
			t.Cleanup(server.Close)
			if scenario == "dial-error" {
				server.Close()
			}
			oldFactory := GetMihomoAdapterFunc
			t.Cleanup(func() { GetMihomoAdapterFunc = oldFactory })
			for range 3 {
				proxy := testutil.NewTrackingProxy(t, "")
				GetMihomoAdapterFunc = func(string) (constant.Proxy, error) { return proxy, nil }
				target := server.URL
				if scenario == "invalid-url" {
					target = "http://[invalid"
				}
				_, err := FetchWithProxy(target, true, "test-proxy", time.Second, "")
				if (err == nil) != (scenario == "success") {
					t.Fatalf("unexpected request result: %v", err)
				}
				connections.WaitForClosed(t)
				if got := proxy.CloseCount.Load(); got != 1 {
					t.Fatalf("adapter closed %d times, want 1", got)
				}
			}
		})
	}
}

func TestHTTPClientBorrowsAdapterAndClosesPrivatePool(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "body")
	}))
	connections := testutil.TrackHTTPConnections(server)
	server.Start()
	t.Cleanup(server.Close)
	proxy := testutil.NewTrackingProxy(t, "")
	for range 3 {
		client := NewHTTPClientWithAdapter(proxy, time.Second, nil)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		CloseHTTPClient(client)
		connections.WaitForClosed(t)
	}
	if got := proxy.CloseCount.Load(); got != 0 {
		t.Fatalf("borrowed adapter closed %d times", got)
	}
}

func TestCloseHTTPClientClosesOwnedAdapterOnce(t *testing.T) {
	proxy := testutil.NewTrackingProxy(t, "")
	oldFactory := GetMihomoAdapterFunc
	GetMihomoAdapterFunc = func(string) (constant.Proxy, error) { return proxy, nil }
	t.Cleanup(func() { GetMihomoAdapterFunc = oldFactory })
	client, _, err := CreateProxyHTTPClient(true, "test-proxy", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() { CloseHTTPClient(client) })
	}
	wg.Wait()
	if got := proxy.CloseCount.Load(); got != 1 {
		t.Fatalf("adapter closed %d times, want 1", got)
	}
}
