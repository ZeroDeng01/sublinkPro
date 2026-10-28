package socks5

import (
	"context"
	"net"
	"testing"
	"time"
)

type testRemoteAddrConn struct {
	net.Conn
	remote net.Addr
}

func (c *testRemoteAddrConn) RemoteAddr() net.Addr { return c.remote }

type testAddr string

func (a testAddr) Network() string { return "tcp" }
func (a testAddr) String() string  { return string(a) }

func TestSessionRegistryEnforcesGlobalAndPerClientLimits(t *testing.T) {
	registry := newSessionRegistry(2, 1)
	clientA0, peerA := net.Pipe()
	clientA := &testRemoteAddrConn{Conn: clientA0, remote: testAddr("10.0.0.1:1001")}
	defer func() { _ = clientA.Close(); _ = peerA.Close() }()
	first, ctxA, ok := registry.register(context.Background(), clientA)
	if !ok || first == nil || ctxA == nil {
		t.Fatal("first session was not registered")
	}
	clientB0, peerB := net.Pipe()
	clientB := &testRemoteAddrConn{Conn: clientB0, remote: testAddr("10.0.0.1:1002")}
	defer func() { _ = clientB.Close(); _ = peerB.Close() }()
	if _, _, ok := registry.register(context.Background(), clientB); ok {
		t.Fatal("per-client limit was not enforced")
	}
	clientC0, peerC := net.Pipe()
	clientC := &testRemoteAddrConn{Conn: clientC0, remote: testAddr("10.0.0.2:1001")}
	defer func() { _ = clientC.Close(); _ = peerC.Close() }()
	second, _, ok := registry.register(context.Background(), clientC)
	if !ok || second == nil {
		t.Fatal("second client should use remaining global slot")
	}
	clientD0, peerD := net.Pipe()
	clientD := &testRemoteAddrConn{Conn: clientD0, remote: testAddr("10.0.0.3:1001")}
	defer func() { _ = clientD.Close(); _ = peerD.Close() }()
	if _, _, ok := registry.register(context.Background(), clientD); ok {
		t.Fatal("global limit was not enforced")
	}
	registry.remove(first.ID)
	registry.remove(second.ID)
}

func TestPumpWithSessionTracksTrafficAndContextCancellation(t *testing.T) {
	registry := newSessionRegistry(1, 1)
	client, clientPeer := net.Pipe()
	defer func() { _ = client.Close(); _ = clientPeer.Close() }()
	upstream, upstreamPeer := net.Pipe()
	defer func() { _ = upstream.Close(); _ = upstreamPeer.Close() }()
	session, registeredCtx, ok := registry.register(context.Background(), client)
	if !ok {
		t.Fatal("session registration failed")
	}
	ctx, cancel := context.WithCancel(registeredCtx)
	defer cancel()
	done := make(chan struct{})
	go func() { pumpWithSession(ctx, session, client, upstream, 0, 0); close(done) }()
	payload := []byte("hello")
	go func() { _, _ = clientPeer.Write(payload) }()
	buf := make([]byte, len(payload))
	if _, err := upstreamPeer.Read(buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != string(payload) {
		t.Fatalf("payload=%q", buf)
	}
	if session.UploadBytes.Load() != int64(len(payload)) {
		t.Fatalf("upload=%d", session.UploadBytes.Load())
	}
	cancel()
	_ = client.Close()
	_ = clientPeer.Close()
	_ = upstream.Close()
	_ = upstreamPeer.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pump did not stop after context/connection close")
	}
	stats := registry.snapshotStats()
	if stats.UploadBytes != int64(len(payload)) || stats.ActiveConnections != 1 {
		t.Fatalf("active stats = %+v", stats)
	}
	registry.remove(session.ID)
	stats = registry.snapshotStats()
	if stats.UploadBytes != int64(len(payload)) || stats.ActiveConnections != 0 {
		t.Fatalf("completed stats = %+v", stats)
	}
}

func TestSessionRegistryRejectsCancelledParent(t *testing.T) {
	registry := newSessionRegistry(1, 1)
	client, peer := net.Pipe()
	defer func() { _ = client.Close(); _ = peer.Close() }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if session, connCtx, ok := registry.register(ctx, client); ok || session != nil || connCtx != nil {
		t.Fatalf("cancelled parent registered session: session=%v ctx=%v ok=%v", session, connCtx, ok)
	}
	if stats := registry.snapshotStats(); stats.TotalConnections != 0 || stats.ActiveConnections != 0 {
		t.Fatalf("cancelled registration changed stats: %+v", stats)
	}
}
