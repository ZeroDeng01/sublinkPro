package mihomo

import (
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sublink/internal/testutil"

	"github.com/metacubex/mihomo/constant"
)

func TestGetMihomoAdapterSudoku(t *testing.T) {
	const payload = `{"n":"Sudoku adapter","h":"2001:db8::1","p":443,"k":"example-key","e":"aes-128-gcm","a":"entropy","x":true,"hm":"ws","ht":true,"hx":"on","ts":["xpxvvpvv","vxpvxvvp"]}`
	link := "sudoku://" + base64.RawURLEncoding.EncodeToString([]byte(payload))
	adapter, err := GetMihomoAdapter(link)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = adapter.Close() })
	if adapter.Type() != constant.Sudoku || adapter.Addr() != "[2001:db8::1]:443" || !adapter.SupportUDP() {
		t.Fatalf("unexpected adapter: %v %s", adapter.Type(), adapter.Addr())
	}
}

func TestSpeedTestEOFReleasesConnections(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, strings.Repeat("x", 16*1024))
	}))
	connections := testutil.TrackHTTPConnections(server)
	server.Start()
	t.Cleanup(server.Close)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodConnect {
			http.Error(w, "CONNECT required", http.StatusMethodNotAllowed)
			return
		}
		dialer := &net.Dialer{Timeout: time.Second}
		upstream, err := dialer.DialContext(r.Context(), "tcp", r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer func() { _ = upstream.Close() }()
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "hijacking unavailable", http.StatusInternalServerError)
			return
		}
		client, buffer, err := hijacker.Hijack()
		if err != nil {
			return
		}
		defer func() { _ = client.Close() }()
		if _, err := buffer.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
			return
		}
		if err := buffer.Flush(); err != nil {
			return
		}
		done := make(chan struct{})
		go func() {
			_, _ = io.Copy(upstream, buffer)
			_ = upstream.Close()
			close(done)
		}()
		_, _ = io.Copy(client, upstream)
		_ = client.Close()
		<-done
	}))
	t.Cleanup(proxy.Close)
	for range 3 {
		_, _, downloaded, _, _, err := MihomoSpeedTest(proxy.URL+"#test", server.URL, time.Second, false, "", false, "", "average", 100, "", "")
		if err != nil || downloaded != 16*1024 {
			t.Fatalf("downloaded %d bytes: %v", downloaded, err)
		}
		connections.WaitForClosed(t)
	}
}

func TestQualityAndLandingIPReleaseBorrowedConnections(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ip" {
			_, _ = io.WriteString(w, "1.2.3.4")
			return
		}
		_, _ = io.WriteString(w, `{"ip":"1.2.3.4","isBroadcast":false,"isResidential":true,"fraudScore":0}`)
	}))
	connections := testutil.TrackHTTPConnections(server)
	server.Start()
	t.Cleanup(server.Close)
	proxy := testutil.NewTrackingProxy(t, "")
	for range 3 {
		if got := fetchLandingIPWithAdapter(proxy, server.URL+"/ip"); got != "1.2.3.4" {
			t.Fatalf("landing IP = %q", got)
		}
		connections.WaitForClosed(t)
		quality := FetchQualityWithAdapter(proxy, server.URL+"/quality")
		if quality == nil || quality.IP != "1.2.3.4" {
			t.Fatalf("unexpected quality: %+v", quality)
		}
		connections.WaitForClosed(t)
	}
	if proxy.CloseCount.Load() != 0 {
		t.Fatal("helper closed a borrowed adapter")
	}
}

type closeErrorProxy struct {
	constant.Proxy
	err    error
	closes int
}

func (p *closeErrorProxy) Close() error { p.closes++; return p.err }

func TestDialerProxyClosesBothAdaptersOnceEvenOnError(t *testing.T) {
	mainError, dialerError := errors.New("main close failed"), errors.New("dialer close failed")
	main := &closeErrorProxy{err: mainError}
	dialer := &closeErrorProxy{err: dialerError}
	proxy := &dialerProxy{Proxy: main, dialer: dialer}
	for range 2 {
		err := proxy.Close()
		if !errors.Is(err, mainError) || !errors.Is(err, dialerError) {
			t.Fatalf("close error lost: %v", err)
		}
	}
	if main.closes != 1 || dialer.closes != 1 {
		t.Fatalf("close counts: main=%d dialer=%d", main.closes, dialer.closes)
	}
}
