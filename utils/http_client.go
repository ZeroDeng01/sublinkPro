package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/metacubex/mihomo/constant"
)

// NewHTTPClientWithAdapter creates a private connection pool. The adapter is
// borrowed: its owner must close it after all clients using it have finished.
// A nil adapter uses direct connections. Call CloseHTTPClient when finished.
func NewHTTPClientWithAdapter(proxy constant.Proxy, timeout time.Duration, tlsConfig *tls.Config) *http.Client {
	return &http.Client{Transport: newHTTPTransportWithAdapter(proxy, tlsConfig), Timeout: timeout}
}

// 独立连接池设置空闲超时作为兜底，调用方仍需在生命周期结束时显式关闭。
func newHTTPTransportWithAdapter(proxy constant.Proxy, tlsConfig *tls.Config) *http.Transport {
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		TLSClientConfig:       tlsConfig,
		MaxIdleConns:          100,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	if proxy != nil {
		transport.DialContext = func(ctx context.Context, _, addr string) (net.Conn, error) {
			host, portString, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			port, err := strconv.ParseUint(portString, 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %w", err)
			}
			return proxy.DialContext(ctx, &constant.Metadata{Host: host, DstPort: uint16(port), Type: constant.HTTP})
		}
	}
	return transport
}

// ownedProxyTransport also owns the adapter created for a client's lifetime.
// CloseIdleConnections alone deliberately leaves that adapter usable.
type ownedProxyTransport struct {
	*http.Transport
	adapter constant.Proxy
	once    sync.Once
	err     error
}

func (t *ownedProxyTransport) Close() error {
	t.once.Do(func() {
		t.CloseIdleConnections()
		t.err = t.adapter.Close()
	})
	return t.err
}

// CloseHTTPClient releases a private pool and any adapter owned by that pool.
// Call it after closing response bodies and stopping requests. Clients using
// the shared default transport do not own it, so it is left untouched.
func CloseHTTPClient(client *http.Client) {
	if client == nil || client.Transport == nil {
		return
	}
	if transport, ok := client.Transport.(*ownedProxyTransport); ok {
		_ = transport.Close()
		return
	}
	client.CloseIdleConnections()
}
