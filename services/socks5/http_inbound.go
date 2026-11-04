package socks5

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Keep inbound type out of the public DialFunc contract, including existing probes.
type inboundProtocolKey struct{}

// bufferedConn must preserve both Mixed's lookahead and CONNECT's pre-read bytes.
type bufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) { return c.reader.Read(p) }

func isHTTPMethodByte(b byte) bool {
	return b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", rune(b))
}

// oneConnListener lends an already-admitted client to net/http without opening
// another port or bypassing registry limits. The second Accept waits for the
// HTTP connection (or hijacked CONNECT handler) to finish, so our shared outer
// lifecycle cannot release a session while its handler is still running.
type oneConnListener struct {
	conn      net.Conn
	done      <-chan struct{}
	delivered bool
}

func (l *oneConnListener) Accept() (net.Conn, error) {
	if !l.delivered {
		l.delivered = true
		return l.conn, nil
	}
	<-l.done
	return nil, net.ErrClosed
}
func (l *oneConnListener) Close() error   { return nil } // accepted clients belong to the session
func (l *oneConnListener) Addr() net.Addr { return l.conn.LocalAddr() }

func (s *Server) serveHTTPConn(ctx context.Context, session *connectionInfo, client net.Conn, established func()) {
	session.setInbound(ProtocolHTTP, "")
	_ = client.SetDeadline(time.Time{})
	done := make(chan struct{})
	var once sync.Once
	complete := func() { once.Do(func() { close(done) }) }
	stopLimits := make(chan struct{})
	defer close(stopLimits)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stopLimits:
				return
			case now := <-ticker.C:
				session.mu.Lock()
				last, start := session.LastActivity, session.StartedAt
				session.mu.Unlock()
				if s.cfg.IdleTimeoutSeconds > 0 && now.Sub(last) >= time.Duration(s.cfg.IdleTimeoutSeconds)*time.Second ||
					s.cfg.MaxConnectionDurationSeconds > 0 && now.Sub(start) >= time.Duration(s.cfg.MaxConnectionDurationSeconds)*time.Second {
					session.close()
					return
				}
			}
		}
	}()
	server := &http.Server{
		ReadHeaderTimeout: handshakeTimeout, MaxHeaderBytes: 64 << 10,
		// Use the gateway's sliding idle/duration limits, including active uploads.
		IdleTimeout: time.Duration(s.cfg.IdleTimeoutSeconds) * time.Second,
		ErrorLog:    log.New(io.Discard, "", 0),
		BaseContext: func(net.Listener) context.Context { return ctx },
		ConnState: func(_ net.Conn, state http.ConnState) {
			if state == http.StateClosed {
				complete()
			}
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.handleHTTPRequest(ctx, session, w, r, established, complete)
		}),
	}
	_ = server.Serve(&oneConnListener{conn: &activityConn{Conn: client, touch: session.touch}, done: done})
	_ = server.Close()
}

func (s *Server) httpIdentity(r *http.Request) (routingIdentity, error) {
	identity := defaultRoutingIdentity()
	identity.ProfileID = s.defaultProfileID
	if !s.cfg.RequireAuth {
		return identity, nil
	}
	values := r.Header.Values("Proxy-Authorization")
	if len(values) != 1 || len(values[0]) > 1024 {
		return routingIdentity{}, errors.New("proxy authentication required")
	}
	fields := strings.Fields(values[0])
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Basic") {
		return routingIdentity{}, errors.New("proxy authentication required")
	}
	decoded, err := base64.StdEncoding.DecodeString(fields[1])
	if err != nil || len(decoded) > 511 {
		return routingIdentity{}, errors.New("invalid proxy credentials")
	}
	username, password, ok := strings.Cut(string(decoded), ":")
	if !ok || username == "" || len(username) > 255 || password == "" || len(password) > 255 {
		return routingIdentity{}, errors.New("invalid proxy credentials")
	}
	return s.authenticateCredentials(username, []byte(password))
}

func rejectHTTPProxy(w http.ResponseWriter, status int, message string) {
	// Do not drain an untrusted upload or allow credential retries on this TCP session.
	w.Header().Set("Connection", "close")
	if status == http.StatusProxyAuthRequired {
		w.Header().Set("Proxy-Authenticate", `Basic realm="SublinkPro proxy", charset="UTF-8"`)
	}
	http.Error(w, message, status)
}

// CONNECT requires an explicit port. Absolute http targets default to port 80.
func proxyTarget(authority string, defaultPort uint16) (string, uint16, error) {
	if authority == "" || strings.ContainsAny(authority, " \t\r\n/\\?#@%") {
		return "", 0, errors.New("invalid proxy target")
	}
	parsed, err := url.Parse("http://" + authority)
	if err != nil || parsed.User != nil || parsed.Host != authority || parsed.Hostname() == "" {
		return "", 0, errors.New("invalid proxy target")
	}
	host := parsed.Hostname()
	if strings.Contains(host, ":") && (net.ParseIP(host) == nil || !strings.HasPrefix(authority, "[")) {
		return "", 0, errors.New("invalid proxy target")
	}
	port := defaultPort
	text := parsed.Port()
	if text != "" {
		value, parseErr := strconv.ParseUint(text, 10, 16)
		if parseErr != nil || value == 0 {
			return "", 0, errors.New("invalid proxy port")
		}
		port = uint16(value)
	} else if strings.HasSuffix(authority, ":") || defaultPort == 0 {
		return "", 0, errors.New("proxy port required")
	}
	return host, port, nil
}

func (s *Server) handleHTTPRequest(ctx context.Context, session *connectionInfo, w http.ResponseWriter, r *http.Request, established func(), complete func()) {
	identity, err := s.httpIdentity(r)
	if err != nil {
		rejectHTTPProxy(w, http.StatusProxyAuthRequired, "Proxy authentication required")
		return
	}
	session.setInbound(ProtocolHTTP, r.Method)
	if r.ProtoMajor != 1 || len(r.Header.Values("Upgrade")) > 0 || headerHasToken(strings.Join(r.Header.Values("Connection"), ","), "upgrade") {
		rejectHTTPProxy(w, http.StatusNotImplemented, "HTTP upgrades are not supported")
		return
	}
	authority, defaultPort := r.URL.Host, uint16(80)
	if r.Method == http.MethodConnect {
		authority, defaultPort = r.RequestURI, 0
		if r.ContentLength > 0 || len(r.TransferEncoding) > 0 {
			rejectHTTPProxy(w, http.StatusBadRequest, "CONNECT must not have a request body")
			return
		}
	} else if !r.URL.IsAbs() || !strings.EqualFold(r.URL.Scheme, "http") || r.URL.User != nil || r.URL.Fragment != "" {
		rejectHTTPProxy(w, http.StatusBadRequest, "An absolute http:// target is required; use CONNECT for HTTPS")
		return
	}
	host, port, err := proxyTarget(authority, defaultPort)
	if err != nil {
		rejectHTTPProxy(w, http.StatusBadRequest, "Invalid proxy target")
		return
	}
	upstream, release, err := s.dialTarget(r.Context(), session, identity, host, port, ProtocolHTTP)
	if err != nil {
		rejectHTTPProxy(w, http.StatusBadGateway, "No outbound node could connect")
		return
	}
	hijacked := false
	defer func() {
		_ = upstream.Close()
		session.finishRequest()
		release()
		if hijacked {
			complete()
		}
	}()
	if r.Method == http.MethodConnect {
		conn, rw, hijackErr := http.NewResponseController(w).Hijack()
		if hijackErr != nil {
			rejectHTTPProxy(w, http.StatusInternalServerError, "CONNECT is unavailable")
			return
		}
		// net/http no longer owns a hijacked connection; always finish it, even on panic.
		hijacked = true
		defer func() { _ = conn.Close() }()
		_ = conn.SetDeadline(time.Now().Add(handshakeTimeout))
		if _, err := rw.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
			return
		}
		if err := rw.Flush(); err != nil {
			return
		}
		_ = conn.SetDeadline(time.Time{})
		established()
		// The header parser may already hold the first TLS/application bytes.
		pumpWithSession(ctx, session, &bufferedConn{Conn: conn, reader: rw.Reader}, upstream, s.cfg.IdleTimeoutSeconds, s.cfg.MaxConnectionDurationSeconds)
		return
	}
	s.forwardHTTPRequest(session, w, r, upstream, established)
}

func headerHasToken(value, token string) bool {
	for _, v := range strings.Split(value, ",") {
		if strings.EqualFold(strings.TrimSpace(v), token) {
			return true
		}
	}
	return false
}

// Remove all hop-by-hop fields, including extension names listed by Connection.
func stripProxyHeaders(header http.Header) map[string]bool {
	blocked := map[string]bool{}
	for _, value := range header.Values("Connection") {
		for _, key := range strings.Split(value, ",") {
			if key = strings.TrimSpace(key); key != "" {
				blocked[http.CanonicalHeaderKey(key)] = true
			}
		}
	}
	for _, key := range []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authorization", "Proxy-Authenticate", "Proxy-Authentication-Info", "Te", "Trailer", "Transfer-Encoding", "Upgrade"} {
		blocked[key] = true
	}
	for key := range blocked {
		header.Del(key)
	}
	return blocked
}

const (
	uploadReading uint32 = iota
	uploadComplete
	uploadAborted
)

// Transport may close an upload while its Read is blocked on the client. The
// server body holds a mutex across that Read, so Close alone cannot interrupt it.
// Interrupt only incomplete bodies, and never reuse their inbound connection.
type countedRequestBody struct {
	io.ReadCloser
	session       *connectionInfo
	contentLength int64
	read          int64
	readMu        sync.Mutex    // Close waits for accounting as well as the underlying Read
	state         atomic.Uint32 // completion and interruption are mutually exclusive
	closeOnce     sync.Once
	closeErr      error
}

func (b *countedRequestBody) Read(p []byte) (int, error) {
	b.readMu.Lock()
	defer b.readMu.Unlock()
	n, err := b.ReadCloser.Read(p)
	b.session.UploadBytes.Add(int64(n))
	b.read += int64(n)
	if err == io.EOF || b.contentLength >= 0 && b.read >= b.contentLength {
		b.state.CompareAndSwap(uploadReading, uploadComplete)
	}
	if n > 0 {
		b.session.touch()
	}
	return n, err
}

func (b *countedRequestBody) Close() error {
	b.closeOnce.Do(func() {
		b.abort()
		b.readMu.Lock()
		defer b.readMu.Unlock()
		b.closeErr = b.ReadCloser.Close()
	})
	return b.closeErr
}

func (b *countedRequestBody) abort() {
	if b.state.CompareAndSwap(uploadReading, uploadAborted) {
		_ = b.session.client.SetReadDeadline(time.Now())
	}
}

// On an upstream error Transport waits for its upload writer before returning
// from RoundTrip. Wake that writer independently, not only in the handler defer.
// Setting the deadline is nonblocking even while Read holds the server-body lock.
type uploadAbortConn struct {
	net.Conn
	upload *countedRequestBody
}

func (c *uploadAbortConn) Close() error {
	if c.upload != nil {
		c.upload.abort()
	}
	return c.Conn.Close()
}

func (b *countedRequestBody) reusable() bool {
	return b.state.Load() == uploadComplete
}

type proxyBodyWriter struct {
	writer  http.ResponseWriter
	session *connectionInfo
}

func (w proxyBodyWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	w.session.DownloadBytes.Add(int64(n))
	if n > 0 {
		w.session.touch()
	}
	if err == nil {
		err = http.NewResponseController(w.writer).Flush()
	}
	return n, err
}

func (s *Server) forwardHTTPRequest(session *connectionInfo, w http.ResponseWriter, r *http.Request, upstream net.Conn, established func()) {
	// Responses must be allowed to stream before a client finishes its upload.
	// Otherwise net/http drains the body before flushing and can deadlock with
	// Transport's concurrent upload reader when the origin responds early.
	if err := http.NewResponseController(w).EnableFullDuplex(); err != nil {
		rejectHTTPProxy(w, http.StatusInternalServerError, "HTTP streaming is unavailable")
		return
	}
	out := r.Clone(r.Context())
	out.RequestURI = ""
	out.Host = out.URL.Host
	out.Close = true
	out.GetBody = nil
	stripProxyHeaders(out.Header)
	// Do not forward unvalidated request trailers or proxy credentials in trailers.
	out.Trailer = nil
	var upload *countedRequestBody
	if r.Body != nil && r.Body != http.NoBody {
		upload = &countedRequestBody{ReadCloser: r.Body, session: session, contentLength: r.ContentLength}
		out.Body = upload
		// After the response ends, interrupt any unfinished inbound upload before
		// releasing the shared session, including on errors and handler aborts.
		defer func() { _ = upload.Close() }()
	}
	var used atomic.Bool
	transport := &http.Transport{
		Proxy: nil, DisableKeepAlives: true, DisableCompression: true,
		MaxResponseHeaderBytes: 64 << 10,
		ExpectContinueTimeout:  time.Second, ResponseHeaderTimeout: time.Duration(s.cfg.DialTimeoutSeconds) * time.Second,
		DialContext: func(context.Context, string, string) (net.Conn, error) {
			if used.Swap(true) {
				return nil, errors.New("request replay is disabled")
			}
			return &uploadAbortConn{Conn: &activityConn{Conn: upstream, touch: session.touch}, upload: upload}, nil
		},
	}
	defer transport.CloseIdleConnections()
	response, err := transport.RoundTrip(out)
	if err != nil {
		rejectHTTPProxy(w, http.StatusBadGateway, "Upstream HTTP request failed")
		return
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusSwitchingProtocols {
		rejectHTTPProxy(w, http.StatusNotImplemented, "HTTP upgrades are not supported")
		return
	}
	blocked := stripProxyHeaders(response.Header)
	for key, values := range response.Header {
		w.Header()[key] = append([]string(nil), values...)
	}
	for key := range response.Trailer {
		if !blocked[http.CanonicalHeaderKey(key)] && !strings.EqualFold(key, "Content-Length") {
			w.Header().Add("Trailer", key)
		}
	}
	if upload != nil && !upload.reusable() {
		// Unread upload bytes must never become the next keep-alive request. This
		// is decided before headers; a concurrent completion is safe to close too.
		r.Close = true
		w.Header().Set("Connection", "close")
	}
	established()
	w.WriteHeader(response.StatusCode)
	if err := http.NewResponseController(w).Flush(); err != nil {
		panic(http.ErrAbortHandler)
	}
	if _, err := io.Copy(proxyBodyWriter{writer: w, session: session}, response.Body); err != nil {
		// A partial response cannot be replaced with a fresh 502 or replayed on another node.
		panic(http.ErrAbortHandler)
	}
	for key, values := range response.Trailer {
		if !blocked[http.CanonicalHeaderKey(key)] && !strings.EqualFold(key, "Content-Length") {
			w.Header()[http.TrailerPrefix+key] = values
		}
	}
}
