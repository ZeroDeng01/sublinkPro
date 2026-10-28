package socks5

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sublink/models"
)

func TestHTTPProxyEarlyResponseDoesNotHangOrReplayUpload(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Link: "one"}})
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	}))
	defer origin.Close()
	server, address := startInboundTestServer(t, ProtocolHTTP, Config{IdleTimeoutSeconds: 2}, directInboundTestDial)
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, origin.URL, strings.NewReader(strings.Repeat("x", 1<<20)))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Expect", "100-continue")
	start := time.Now()
	client := proxyTestClient(t, address, "", "")
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("test client transport is not HTTP")
	}
	transport.ExpectContinueTimeout = time.Second
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, response.Body)
	_ = response.Body.Close()
	if response.StatusCode != http.StatusRequestEntityTooLarge || time.Since(start) > 4*time.Second {
		t.Fatalf("early response stalled: %d %s", response.StatusCode, time.Since(start))
	}
	waitInboundIdle(t, server)
}
func TestHTTPProxyPartialUploadGetsEarlyResponseWithoutIdleTimeout(t *testing.T) {
	for _, body := range []string{"", "denied"} {
		t.Run(fmt.Sprintf("body-%d", len(body)), func(t *testing.T) {
			withCandidateNodes(t, []models.Node{{ID: 1, Link: "one"}})
			stopOrigin := make(chan struct{})
			t.Cleanup(func() { close(stopOrigin) })
			server, address := startInboundTestServer(t, ProtocolHTTP, Config{IdleTimeoutSeconds: 0, MaxConnectionDurationSeconds: 0}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
				conn, peer := net.Pipe()
				go func() {
					defer func() { _ = peer.Close() }()
					_, err := http.ReadRequest(bufio.NewReader(peer))
					if err != nil {
						return
					}
					_, _ = fmt.Fprintf(peer, "HTTP/1.1 413 Request Entity Too Large\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", len(body), body)
					// Keep the receive side open: an immediate net.Pipe close can
					// race the response with an unrelated upstream write failure.
					<-stopOrigin
				}()
				return conn, nil
			})
			var d net.Dialer
			conn, err := d.DialContext(context.Background(), "tcp", address)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = conn.Close() }()
			_ = conn.SetDeadline(time.Now().Add(1500 * time.Millisecond))
			// No Expect header, incomplete upload, no half-close and no gateway idle timer.
			_, _ = io.WriteString(conn, "POST http://example.invalid/upload HTTP/1.1\r\nHost: example.invalid\r\nContent-Length: 1000000\r\n\r\nx")
			response, payload := readTestResponse(t, bufio.NewReader(conn), http.MethodPost)
			if response.StatusCode != 413 || payload != body || !response.Close {
				t.Fatalf("early response=%d %q close=%v", response.StatusCode, payload, response.Close)
			}
			waitInboundIdle(t, server)
		})
	}
}

func TestHTTPProxyStreamsResponseWhileUploadIsIncomplete(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Link: "one"}})
	originErrors := make(chan error, 1)
	server, address := startInboundTestServer(t, ProtocolHTTP, Config{}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		conn, peer := net.Pipe()
		go func() {
			defer func() { _ = peer.Close() }()
			_ = peer.SetDeadline(time.Now().Add(5 * time.Second))
			r, err := http.ReadRequest(bufio.NewReader(peer))
			if err == nil {
				_, err = io.WriteString(peer, "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\nConnection: close\r\n\r\n5\r\nready\r\n")
			}
			if err == nil {
				_, err = io.CopyN(io.Discard, r.Body, 32768)
			}
			if err == nil {
				_, err = io.WriteString(peer, "4\r\nnext\r\n")
			}
			if err == nil {
				_, err = io.CopyN(io.Discard, r.Body, 32768)
			}
			if err == nil {
				_, err = io.WriteString(peer, "4\r\ndone\r\n0\r\n\r\n")
			}
			originErrors <- err
		}()
		return conn, nil
	})
	var d net.Dialer
	conn, err := d.DialContext(context.Background(), "tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, _ = io.WriteString(conn, "POST http://example.invalid/stream HTTP/1.1\r\nHost: example.invalid\r\nContent-Length: 65536\r\n\r\n")
	response, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: http.MethodPost})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	for _, stage := range []struct {
		text string
		send bool
	}{{"ready", true}, {"next", true}, {"done", false}} {
		buf := make([]byte, len(stage.text))
		if _, err := io.ReadFull(response.Body, buf); err != nil || string(buf) != stage.text {
			t.Fatalf("duplex stage %q: %q %v", stage.text, buf, err)
		}
		if stage.send {
			if _, err := io.WriteString(conn, strings.Repeat("x", 32768)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if rest, err := io.ReadAll(response.Body); err != nil || len(rest) != 0 {
		t.Fatalf("response end %q %v", rest, err)
	}
	if err := <-originErrors; err != nil {
		t.Fatal(err)
	}
	if !response.Close {
		t.Fatal("partially uploaded request must not be reused")
	}
	_ = conn.Close()
	waitInboundIdle(t, server)
	if stats := server.Stats(); stats.UploadBytes != 65536 || stats.DownloadBytes != 13 {
		t.Fatalf("duplex statistics: %+v", stats)
	}
}

func TestHTTPProxyCompletedUploadKeepsInboundReusable(t *testing.T) {
	for _, chunked := range []bool{false, true} {
		t.Run(fmt.Sprintf("chunked-%v", chunked), func(t *testing.T) {
			withCandidateNodes(t, []models.Node{{ID: 1, Link: "one"}})
			origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "body read", 400)
					return
				}
				_, _ = w.Write(body)
			}))
			defer origin.Close()
			server, address := startInboundTestServer(t, ProtocolHTTP, Config{}, directInboundTestDial)
			var d net.Dialer
			conn, err := d.DialContext(context.Background(), "tcp", address)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = conn.Close() }()
			_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
			reader := bufio.NewReader(conn)
			for i := 0; i < 2; i++ {
				framing := "Content-Length: 4\r\n\r\ndata"
				if chunked {
					framing = "Transfer-Encoding: chunked\r\n\r\n4\r\ndata\r\n0\r\n\r\n"
				}
				_, _ = fmt.Fprintf(conn, "POST %s HTTP/1.1\r\nHost: ignored.invalid\r\n%s", origin.URL, framing)
				response, body := readTestResponse(t, reader, http.MethodPost)
				if response.StatusCode != 200 || body != "data" || response.Close {
					t.Fatalf("request %d: status=%d body=%q close=%v", i, response.StatusCode, body, response.Close)
				}
			}
			_ = conn.Close()
			waitInboundIdle(t, server)
		})
	}
}

func TestHTTPProxyUpstreamFailureInterruptsIncompleteUpload(t *testing.T) {
	for _, malformed := range []bool{false, true} {
		t.Run(fmt.Sprintf("malformed-%v", malformed), func(t *testing.T) {
			withCandidateNodes(t, []models.Node{{ID: 1, Link: "one"}})
			originErrors := make(chan error, 1)
			stopOrigin := make(chan struct{})
			t.Cleanup(func() { close(stopOrigin) })
			server, address := startInboundTestServer(t, ProtocolHTTP, Config{IdleTimeoutSeconds: 0, MaxConnectionDurationSeconds: 0}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
				conn, peer := net.Pipe()
				go func() {
					defer func() { _ = peer.Close() }()
					_ = peer.SetDeadline(time.Now().Add(5 * time.Second))
					r, err := http.ReadRequest(bufio.NewReader(peer))
					if err == nil {
						_, err = io.CopyN(io.Discard, r.Body, 32768)
					}
					if err == nil && malformed {
						_, err = io.WriteString(peer, "HTTP/1.1 200 OK\r\nBad Header\r\n\r\n")
					}
					originErrors <- err
					if malformed {
						// The gateway must detect invalid headers itself, rather than
						// depending on EOF from the origin to interrupt the upload.
						<-stopOrigin
					}
				}()
				return conn, nil
			})
			var d net.Dialer
			conn, err := d.DialContext(context.Background(), "tcp", address)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = conn.Close() }()
			_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
			_, _ = io.WriteString(conn, "POST http://example.invalid/upload HTTP/1.1\r\nHost: example.invalid\r\nContent-Length: 1000000\r\n\r\n"+strings.Repeat("x", 32768))
			response, _ := readTestResponse(t, bufio.NewReader(conn), http.MethodPost)
			if response.StatusCode != http.StatusBadGateway || !response.Close {
				t.Fatalf("failed origin: status=%d close=%v", response.StatusCode, response.Close)
			}
			if err := <-originErrors; err != nil {
				t.Fatal(err)
			}
			waitInboundIdle(t, server)
		})
	}
}
