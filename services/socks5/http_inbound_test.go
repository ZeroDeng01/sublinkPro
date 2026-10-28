package socks5

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"sublink/models"
)

func startInboundTestServer(t *testing.T, protocol string, cfg Config, dial DialFunc) (*Server, string) {
	t.Helper()
	server, err := newServerForListener(cfg, dial, "test", defaultProfileID, protocol)
	if err != nil {
		t.Fatal(err)
	}
	var lc net.ListenConfig
	listener, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- server.Serve(ctx, listener) }()
	t.Cleanup(func() {
		cancel()
		_ = listener.Close()
		server.Close()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(3 * time.Second):
			t.Error("proxy did not stop")
		}
	})
	return server, listener.Addr().String()
}

func directInboundTestDial(ctx context.Context, _ models.Node, host string, port uint16) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, "tcp", net.JoinHostPort(host, fmt.Sprint(port)))
}

func proxyTestClient(t *testing.T, address, user, password string) *http.Client {
	t.Helper()
	proxy, err := url.Parse("http://" + address)
	if err != nil {
		t.Fatal(err)
	}
	if user != "" {
		proxy.User = url.UserPassword(user, password)
	}
	transport := &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true}
	t.Cleanup(transport.CloseIdleConnections)
	return &http.Client{Transport: transport, Timeout: 5 * time.Second}
}

type inboundTestResponse struct {
	StatusCode int
	Close      bool
}

func doInboundTestRequest(t *testing.T, client *http.Client, method, target string, body io.Reader) (*http.Response, error) {
	t.Helper()
	request, err := http.NewRequestWithContext(context.Background(), method, target, body)
	if err != nil {
		t.Fatal(err)
	}
	return client.Do(request)
}

func readTestResponse(t *testing.T, reader *bufio.Reader, method string) (inboundTestResponse, string) {
	t.Helper()
	response, err := http.ReadResponse(reader, &http.Request{Method: method})
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	return inboundTestResponse{StatusCode: response.StatusCode, Close: response.Close}, string(body)
}

func waitInboundIdle(t *testing.T, server *Server) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if server.Stats().ActiveConnections == 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("sessions not released: %+v", server.Connections())
}

func TestHTTPProxyAuthenticatesAndStripsProxyHeaders(t *testing.T) {
	node := models.Node{ID: 1, Name: "outbound", Link: "test://one"}
	withCandidateNodes(t, []models.Node{node})
	seen := make(chan http.Header, 1)
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen <- r.Header.Clone()
		w.Header().Set("Connection", "X-Origin-Hop")
		w.Header().Set("X-Origin-Hop", "remove")
		w.Header().Set("Proxy-Authenticate", "must-not-leak")
		w.Header().Set("X-End-To-End", "kept")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, "origin-response")
	}))
	defer origin.Close()
	var dials atomic.Int32
	server, address := startInboundTestServer(t, ProtocolHTTP, Config{Username: "proxy", Password: "secret", RequireAuth: true}, func(ctx context.Context, n models.Node, h string, p uint16) (net.Conn, error) {
		dials.Add(1)
		if got, _ := ctx.Value(inboundProtocolKey{}).(string); got != ProtocolHTTP {
			return nil, errors.New("missing HTTP metadata")
		}
		return directInboundTestDial(ctx, n, h, p)
	})
	for _, credentials := range []struct{ user, password string }{{"", ""}, {"proxy", "wrong"}} {
		response, err := doInboundTestRequest(t, proxyTestClient(t, address, credentials.user, credentials.password), http.MethodGet, origin.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		if response.StatusCode != 407 || response.Header.Get("Proxy-Authenticate") == "" {
			t.Fatalf("authentication response: %v", response)
		}
	}
	if dials.Load() != 0 {
		t.Fatal("unauthenticated request dialed outbound")
	}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, origin.URL+"/path?q=one", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer origin-only")
	request.Header.Set("Connection", "X-Private-Hop")
	request.Header.Set("X-Private-Hop", "remove")
	response, err := proxyTestClient(t, address, "proxy", "secret").Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil || response.StatusCode != 201 || string(body) != "origin-response" {
		t.Fatalf("response %v %q %v", response.StatusCode, body, err)
	}
	headers := <-seen
	if headers.Get("Proxy-Authorization") != "" || headers.Get("X-Private-Hop") != "" || headers.Get("Authorization") != "Bearer origin-only" {
		t.Fatalf("origin headers=%v", headers)
	}
	if response.Header.Get("Proxy-Authenticate") != "" || response.Header.Get("X-End-To-End") != "kept" {
		t.Fatalf("response headers=%v", response.Header)
	}
	waitInboundIdle(t, server)
	if got := server.Stats(); got.SuccessfulConnections != 1 || got.FailedConnections != 2 || got.DownloadBytes != int64(len(body)) {
		t.Fatalf("stats=%+v", got)
	}
	if got := server.router.runtime.metrics(node); got.ActiveConnections != 0 || got.SuccessfulConnections != 1 {
		t.Fatalf("runtime=%+v", got)
	}
}

func TestHTTPProxyKeepAliveReauthenticatesAndChangesTargetAndProfile(t *testing.T) {
	previous := listCandidateNodesFunc
	listCandidateNodesFunc = func(cfg Config) ([]models.Node, error) {
		id := 1
		if len(cfg.CandidateGroups) > 0 && cfg.CandidateGroups[0] == "second" {
			id = 2
		}
		return []models.Node{{ID: id, Name: fmt.Sprint(id), Link: fmt.Sprint(id)}}, nil
	}
	t.Cleanup(func() { listCandidateNodesFunc = previous })
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, r.Host+r.URL.Path) }))
	defer origin.Close()
	parsed, _ := url.Parse(origin.URL)
	ids := make(chan int, 3)
	server, address := startInboundTestServer(t, ProtocolHTTP, Config{Username: "proxy", Password: "secret", RequireAuth: true}, func(ctx context.Context, n models.Node, _ string, _ uint16) (net.Conn, error) {
		ids <- n.ID
		var d net.Dialer
		return d.DialContext(ctx, "tcp", parsed.Host)
	})
	profile, err := normalizeRoutingProfile(server.cfg, RoutingProfile{ID: "second", Name: "Second", Enabled: true, CandidateGroups: []string{"second"}})
	if err != nil {
		t.Fatal(err)
	}
	server.profiles.replace(server.cfg, []RoutingProfile{defaultRoutingProfile(server.cfg), profile})
	server.accounts.replace([]Account{{ID: "alice", Username: "alice", Password: "own-secret", ProfileID: "second", Enabled: true}})
	var d net.Dialer
	conn, err := d.DialContext(context.Background(), "tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	reader := bufio.NewReader(conn)
	for i, item := range []struct {
		user, password, host string
		node                 int
	}{{"proxy", "secret", "one.invalid", 1}, {"alice", "own-secret", "two.invalid", 2}, {"proxy@second.account", "secret", "three.invalid", 2}} {
		auth := base64.StdEncoding.EncodeToString([]byte(item.user + ":" + item.password))
		_, err = fmt.Fprintf(conn, "GET http://%s/path%d HTTP/1.1\r\nHost: %s\r\nProxy-Authorization: Basic %s\r\n\r\n", item.host, i, item.host, auth)
		if err != nil {
			t.Fatal(err)
		}
		response, body := readTestResponse(t, reader, http.MethodGet)
		if response.StatusCode != 200 || body != fmt.Sprintf("%s/path%d", item.host, i) {
			t.Fatalf("keepalive request %d: %d %q", i, response.StatusCode, body)
		}
		if got := <-ids; got != item.node {
			t.Fatalf("cross-profile outbound: got %d want %d", got, item.node)
		}
	}
	// Reusing a TCP session does not reuse its authenticated identity.
	_, _ = io.WriteString(conn, "GET http://four.invalid/ HTTP/1.1\r\nHost: four.invalid\r\n\r\n")
	response, _ := readTestResponse(t, reader, http.MethodGet)
	if response.StatusCode != 407 {
		t.Fatalf("missing per-request auth: %d", response.StatusCode)
	}
	_ = conn.Close()
	waitInboundIdle(t, server)
	if len(ids) != 0 {
		t.Fatal("unauthenticated keepalive request reached origin")
	}
	if got := server.Stats(); got.TotalConnections != 1 || got.SuccessfulConnections != 1 || got.FailedConnections != 0 {
		t.Fatalf("TCP counters count requests: %+v", got)
	}
}

func TestHTTPConnectPreservesBufferedPayloadAndMonitoring(t *testing.T) {
	node := models.Node{ID: 1, Name: "tunnel", Link: "test://tunnel"}
	withCandidateNodes(t, []models.Node{node})
	server, address := startInboundTestServer(t, ProtocolMixed, Config{}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		proxy, peer := net.Pipe()
		go func() { defer func() { _ = peer.Close() }(); _, _ = io.Copy(peer, peer) }()
		return proxy, nil
	})
	var d net.Dialer
	conn, err := d.DialContext(context.Background(), "tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	// Send tunnel data in the same write as the CONNECT header.
	payload := "buffered-first-payload"
	_, err = io.WriteString(conn, "CONNECT example.invalid:443 HTTP/1.1\r\nHost: example.invalid:443\r\n\r\n"+payload)
	if err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(conn)
	status, err := reader.ReadString('\n')
	if err != nil || !strings.Contains(status, "200") {
		t.Fatalf("CONNECT status=%q err=%v", status, err)
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if line == "\r\n" {
			break
		}
		if strings.HasPrefix(strings.ToLower(line), "content-length:") || strings.HasPrefix(strings.ToLower(line), "transfer-encoding:") {
			t.Fatalf("invalid CONNECT header %q", line)
		}
	}
	data := make([]byte, len(payload))
	if _, err := io.ReadFull(reader, data); err != nil || string(data) != payload {
		t.Fatalf("buffered payload %q %v", data, err)
	}
	snapshots := server.Connections()
	if len(snapshots) != 1 || snapshots[0].Protocol != ProtocolHTTP || snapshots[0].Method != "CONNECT" || snapshots[0].Target != "example.invalid:443" || snapshots[0].NodeName != "tunnel" {
		t.Fatalf("monitor=%+v", snapshots)
	}
	if !server.registry.close(snapshots[0].ID) {
		t.Fatal("admin disconnect did not find HTTP session")
	}
	waitInboundIdle(t, server)
	if got := server.Stats(); got.UploadBytes != int64(len(payload)) || got.DownloadBytes != int64(len(payload)) {
		t.Fatalf("tunnel bytes doubled or lost: %+v", got)
	}
	if got := server.router.runtime.metrics(node); got.ActiveConnections != 0 {
		t.Fatalf("CONNECT runtime not released: %+v", got)
	}
}

func TestMixedListenerAcceptsSOCKS5AndHTTP(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Name: "mixed", Link: "test://mixed"}})
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "http-ok") }))
	defer origin.Close()
	server, address := startInboundTestServer(t, ProtocolMixed, Config{}, directInboundTestDial)
	response, err := doInboundTestRequest(t, proxyTestClient(t, address, "", ""), http.MethodGet, origin.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if string(body) != "http-ok" {
		t.Fatalf("HTTP body %q", body)
	}
	var d net.Dialer
	conn, err := d.DialContext(context.Background(), "tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	// Fragment the discriminator so lookahead cannot eat the SOCKS greeting.
	_, _ = conn.Write([]byte{5})
	_, _ = conn.Write([]byte{1, 0})
	method := make([]byte, 2)
	if _, err := io.ReadFull(conn, method); err != nil || method[1] != 0 {
		t.Fatalf("SOCKS method=%x %v", method, err)
	}
	target, _ := url.Parse(origin.URL)
	host, port, _ := net.SplitHostPort(target.Host)
	portValue := 0
	_, _ = fmt.Sscan(port, &portValue)
	request := append([]byte{5, 1, 0, 3, byte(len(host))}, []byte(host)...)
	request = append(request, byte(portValue>>8), byte(portValue))
	_, _ = conn.Write(request)
	reply := make([]byte, 10)
	if _, err := io.ReadFull(conn, reply); err != nil || reply[1] != 0 {
		t.Fatalf("SOCKS reply=%x %v", reply, err)
	}
	_, _ = fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", target.Host)
	_, bodyString := readTestResponse(t, bufio.NewReader(conn), http.MethodGet)
	if bodyString != "http-ok" {
		t.Fatalf("SOCKS tunneled body %q", bodyString)
	}
	_ = conn.Close()
	waitInboundIdle(t, server)
	if got := server.Stats(); got.SuccessfulConnections != 2 {
		t.Fatalf("mixed stats=%+v", got)
	}
}

func TestHTTPProxyRejectsInvalidTargetsAndUpgradeBeforeDial(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Link: "test://one"}})
	var dials atomic.Int32
	_, address := startInboundTestServer(t, ProtocolHTTP, Config{}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		dials.Add(1)
		return nil, errors.New("must not dial")
	})
	for _, tc := range []struct {
		request string
		status  int
	}{
		{"GET /relative HTTP/1.1\r\nHost: example.invalid\r\n\r\n", 400},
		{"GET https://example.invalid/ HTTP/1.1\r\nHost: example.invalid\r\n\r\n", 400},
		{"CONNECT example.invalid HTTP/1.1\r\nHost: example.invalid\r\n\r\n", 400},
		{"CONNECT example.invalid:0 HTTP/1.1\r\nHost: example.invalid:0\r\n\r\n", 400},
		{"CONNECT example.invalid:443 HTTP/1.1\r\nHost: example.invalid\r\nContent-Length: 1\r\n\r\nx", 400},
		{"GET http://example.invalid/ HTTP/1.1\r\nHost: example.invalid\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n", 501},
	} {
		var d net.Dialer
		conn, err := d.DialContext(context.Background(), "tcp", address)
		if err != nil {
			t.Fatal(err)
		}
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		_, _ = io.WriteString(conn, tc.request)
		response, _ := readTestResponse(t, bufio.NewReader(conn), http.MethodGet)
		_ = conn.Close()
		if response.StatusCode != tc.status {
			t.Fatalf("got %d want %d", response.StatusCode, tc.status)
		}
	}
	if dials.Load() != 0 {
		t.Fatalf("rejected request dialed %d times", dials.Load())
	}
}

func TestHTTPProxyRetriesOnlyDialNotSentPost(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Link: "one"}, {ID: 2, Link: "two"}})
	var attempts atomic.Int32
	bodySeen := make(chan string, 1)
	_, address := startInboundTestServer(t, ProtocolHTTP, Config{MaxAttempts: 2}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		attempts.Add(1)
		conn, peer := net.Pipe()
		go func() {
			defer func() { _ = peer.Close() }()
			request, err := http.ReadRequest(bufio.NewReader(peer))
			if err != nil {
				bodySeen <- "read-error"
				return
			}
			body, _ := io.ReadAll(request.Body)
			_ = request.Body.Close()
			bodySeen <- string(body)
		}()
		return conn, nil
	})
	response, err := doInboundTestRequest(t, proxyTestClient(t, address, "", ""), http.MethodPost, "http://example.invalid/charge", strings.NewReader("one-payment"))
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != 502 || <-bodySeen != "one-payment" || attempts.Load() != 1 {
		t.Fatalf("POST replayed or wrong failure: status=%d attempts=%d", response.StatusCode, attempts.Load())
	}
}

func TestHTTPProxyDialRetryAndUpload(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Link: "one"}, {ID: 2, Link: "two"}})
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.Copy(w, r.Body) }))
	defer origin.Close()
	var attempts atomic.Int32
	server, address := startInboundTestServer(t, ProtocolHTTP, Config{MaxAttempts: 2}, func(ctx context.Context, node models.Node, host string, port uint16) (net.Conn, error) {
		if attempts.Add(1) == 1 {
			return nil, errors.New("first outbound unavailable")
		}
		return directInboundTestDial(ctx, node, host, port)
	})
	response, err := doInboundTestRequest(t, proxyTestClient(t, address, "", ""), http.MethodPost, origin.URL, strings.NewReader("upload-bytes"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil || string(body) != "upload-bytes" || attempts.Load() != 2 {
		t.Fatalf("retry response=%q err=%v attempts=%d", body, err, attempts.Load())
	}
	waitInboundIdle(t, server)
	if got := server.Stats(); got.UploadBytes != 12 || got.DownloadBytes != 12 {
		t.Fatalf("HTTP body bytes=%+v", got)
	}
}

func TestHTTPProxyIdleAndDurationCloseSessions(t *testing.T) {
	for _, tc := range []struct {
		name           string
		idle, duration int
	}{{"idle", 1, 0}, {"duration", 0, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			withCandidateNodes(t, []models.Node{{ID: 1, Link: "one"}})
			server, address := startInboundTestServer(t, ProtocolHTTP, Config{IdleTimeoutSeconds: tc.idle, MaxConnectionDurationSeconds: tc.duration}, directInboundTestDial)
			var d net.Dialer
			conn, err := d.DialContext(context.Background(), "tcp", address)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = conn.Close() }()
			// Deliberately stop partway through the header; the limits also cover handshakes.
			_, _ = io.WriteString(conn, "GET http://example.invalid/ HTTP/1.1\r\n")
			_ = conn.SetReadDeadline(time.Now().Add(4 * time.Second))
			_, err = conn.Read(make([]byte, 1))
			if err == nil {
				t.Fatal("incomplete header should close")
			}
			waitInboundIdle(t, server)
		})
	}
}

func TestListenerProtocolNormalizationAndSafety(t *testing.T) {
	base := Config{}
	for _, value := range []string{"", ProtocolSOCKS5, " HTTP ", ProtocolMixed} {
		spec, err := normalizeListener(base, ListenerSpec{ID: "test", Protocol: value, Port: 1080})
		if err != nil {
			t.Fatal(err)
		}
		want := strings.ToLower(strings.TrimSpace(value))
		if want == "" {
			want = ProtocolSOCKS5
		}
		if spec.Protocol != want {
			t.Fatalf("protocol %q became %q", value, spec.Protocol)
		}
	}
	if legacyListener(base).Protocol != ProtocolSOCKS5 {
		t.Fatal("legacy default changed")
	}
	if _, err := normalizeListener(base, ListenerSpec{ID: "test", Protocol: "https", Port: 1080}); err == nil {
		t.Fatal("unknown protocol accepted")
	}
	noAuth := false
	for _, protocol := range []string{ProtocolSOCKS5, ProtocolHTTP, ProtocolMixed} {
		if _, err := normalizeListener(base, ListenerSpec{ID: "test", Protocol: protocol, Enabled: true, ListenAddress: "0.0.0.0", Port: 1080, RequireAuth: &noAuth}); err == nil {
			t.Fatalf("open %s proxy allowed", protocol)
		}
	}
}

func TestHTTPProxyRejectsMalformedAndDuplicateAuthentication(t *testing.T) {
	server, err := NewServer(Config{Username: "proxy", Password: "secret", RequireAuth: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, values := range [][]string{nil, {"Basic ???"}, {"Bearer token"}, {"Basic " + base64.StdEncoding.EncodeToString([]byte("proxy:wrong"))}, {"Basic " + strings.Repeat("A", 1100)}, {"Basic abc", "Basic def"}} {
		request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.invalid/", nil)
		for _, value := range values {
			request.Header.Add("Proxy-Authorization", value)
		}
		if _, err := server.httpIdentity(request); err == nil {
			t.Fatalf("invalid credentials accepted: %d values", len(values))
		}
	}
}

func TestHTTPProxyResponseHopHeadersAndLargeHeaderLimit(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Link: "one"}})
	var dials atomic.Int32
	_, address := startInboundTestServer(t, ProtocolHTTP, Config{}, func(context.Context, models.Node, string, uint16) (net.Conn, error) {
		dials.Add(1)
		conn, peer := net.Pipe()
		go func() {
			defer func() { _ = peer.Close() }()
			request, err := http.ReadRequest(bufio.NewReader(peer))
			if err != nil {
				return
			}
			_ = request.Body.Close()
			_, _ = io.WriteString(peer, "HTTP/1.1 200 OK\r\nContent-Length: 2\r\nConnection: X-Origin-Hop\r\nX-Origin-Hop: remove\r\nProxy-Authenticate: remove\r\nX-End: kept\r\n\r\nok")
		}()
		return conn, nil
	})
	response, err := doInboundTestRequest(t, proxyTestClient(t, address, "", ""), http.MethodGet, "http://example.invalid/", nil)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil || string(body) != "ok" || response.Header.Get("X-Origin-Hop") != "" || response.Header.Get("Proxy-Authenticate") != "" || response.Header.Get("X-End") != "kept" {
		t.Fatalf("response=%v body=%q err=%v", response.Header, body, err)
	}
	var d net.Dialer
	conn, err := d.DialContext(context.Background(), "tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, _ = io.WriteString(conn, "GET http://example.invalid/ HTTP/1.1\r\nHost: example.invalid\r\nX-Large: "+strings.Repeat("x", 80<<10)+"\r\n\r\n")
	result, readErr := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: http.MethodGet})
	if readErr == nil {
		if result.StatusCode != 431 {
			t.Fatalf("large header status=%d", result.StatusCode)
		}
		_ = result.Body.Close()
	} else if netErr, ok := readErr.(net.Error); ok && netErr.Timeout() {
		t.Fatal("large-header rejection timed out")
	}
	// Some TCP stacks reset a rejected connection with unread oversized input.
	if dials.Load() != 1 {
		t.Fatalf("oversized header reached outbound: %d", dials.Load())
	}
}

func TestHTTPProxyCarriesRealHTTPSWithoutInterception(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Link: "tls-node"}})
	origin := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "secure:"+r.URL.Path) }))
	defer origin.Close()
	_, address := startInboundTestServer(t, ProtocolHTTP, Config{Username: "proxy", Password: "secret", RequireAuth: true}, directInboundTestDial)
	proxyURL, _ := url.Parse("http://" + address)
	proxyURL.User = url.UserPassword("proxy", "secret")
	originTransport, ok := origin.Client().Transport.(*http.Transport)
	if !ok {
		t.Fatal("test TLS transport missing")
	}
	transport := &http.Transport{Proxy: http.ProxyURL(proxyURL), TLSClientConfig: originTransport.TLSClientConfig.Clone(), DisableKeepAlives: true}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	response, err := doInboundTestRequest(t, client, http.MethodGet, origin.URL+"/private", nil)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil || string(body) != "secure:/private" {
		t.Fatalf("HTTPS tunnel: %q %v", body, err)
	}
}

func TestHTTPConnectHalfCloseDrainsTrailingResponse(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Link: "one"}})
	var lc net.ListenConfig
	origin, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = origin.Close() }()
	originDone := make(chan struct{})
	go func() {
		defer close(originDone)
		conn, err := origin.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		body, _ := io.ReadAll(conn)
		_, _ = io.WriteString(conn, "tail:"+string(body))
	}()
	_, address := startInboundTestServer(t, ProtocolHTTP, Config{}, directInboundTestDial)
	var d net.Dialer
	raw, err := d.DialContext(context.Background(), "tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	conn, ok := raw.(*net.TCPConn)
	if !ok {
		t.Fatal("TCP connection required")
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, _ = fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", origin.Addr(), origin.Addr())
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if line == "\r\n" {
			break
		}
	}
	_, _ = io.WriteString(conn, "complete-request")
	if err := conn.CloseWrite(); err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(reader)
	if err != nil || string(body) != "tail:complete-request" {
		t.Fatalf("trailing bytes lost: %q %v", body, err)
	}
	select {
	case <-originDone:
	case <-time.After(time.Second):
		t.Fatal("origin not closed")
	}
}

func TestHTTPProxyChunkedUploadAndResponseTrailers(t *testing.T) {
	withCandidateNodes(t, []models.Node{{ID: 1, Link: "one"}})
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Trailer", "X-Checksum")
		_, _ = w.Write(body)
		w.Header().Set("X-Checksum", "done")
	}))
	defer origin.Close()
	server, address := startInboundTestServer(t, ProtocolHTTP, Config{}, directInboundTestDial)
	payload := io.NopCloser(strings.NewReader("streamed-payload"))
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, origin.URL, payload)
	if err != nil {
		t.Fatal(err)
	}
	response, err := proxyTestClient(t, address, "", "").Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil || string(body) != "streamed-payload" || response.Trailer.Get("X-Checksum") != "done" {
		t.Fatalf("body=%q trailer=%v err=%v", body, response.Trailer, err)
	}
	waitInboundIdle(t, server)
	if server.Stats().UploadBytes != int64(len(body)) {
		t.Fatal("chunked transfer double counted")
	}
}
