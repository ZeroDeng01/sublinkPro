package testutil

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/constant"
)

// HTTPConnections observes real keep-alive sockets, including those that remain
// reachable through transport goroutines after the response body is closed.
type HTTPConnections struct{ active atomic.Int64 }

func TrackHTTPConnections(server *httptest.Server) *HTTPConnections {
	connections := &HTTPConnections{}
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			connections.active.Add(1)
		case http.StateClosed, http.StateHijacked:
			connections.active.Add(-1)
		}
	}
	return connections
}

func (c *HTTPConnections) WaitForClosed(t *testing.T) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for c.active.Load() != 0 {
		select {
		case <-deadline.C:
			t.Fatalf("%d HTTP connections remain after operation finished", c.active.Load())
		case <-ticker.C:
		}
	}
}

// TrackingProxy can route public probe URLs to a local test server. It counts
// explicit closes independently of mihomo's finalizer.
type TrackingProxy struct {
	constant.Proxy
	CloseCount atomic.Int32
	Address    string
}

func NewTrackingProxy(t *testing.T, address string) *TrackingProxy {
	t.Helper()
	proxy, err := adapter.ParseProxy(map[string]any{"name": "test-direct", "type": "direct"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = proxy.Close() })
	return &TrackingProxy{Proxy: proxy, Address: address}
}

func (p *TrackingProxy) DialContext(ctx context.Context, metadata *constant.Metadata) (constant.Conn, error) {
	if p.Address != "" {
		host, portString, err := net.SplitHostPort(p.Address)
		if err != nil {
			return nil, err
		}
		port, err := strconv.ParseUint(portString, 10, 16)
		if err != nil {
			return nil, err
		}
		metadata = &constant.Metadata{Host: host, DstPort: uint16(port), Type: constant.HTTP}
	}
	return p.Proxy.DialContext(ctx, metadata)
}

func (p *TrackingProxy) Close() error {
	p.CloseCount.Add(1)
	return p.Proxy.Close()
}
