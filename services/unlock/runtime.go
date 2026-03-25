package unlock

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/metacubex/mihomo/constant"
)

type UnlockRuntime struct {
	Adapter        constant.Proxy
	Client         *http.Client
	Timeout        time.Duration
	LandingCountry string
}

type unlockHTTPResponse struct {
	StatusCode int
	FinalURL   string
	Body       string
	Header     http.Header
}

func newUnlockRuntime(proxyAdapter constant.Proxy, timeout time.Duration, landingCountry string) UnlockRuntime {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return UnlockRuntime{
		Adapter:        proxyAdapter,
		Client:         createUnlockHTTPClient(proxyAdapter, timeout),
		Timeout:        timeout,
		LandingCountry: strings.ToUpper(strings.TrimSpace(landingCountry)),
	}
}

func createUnlockHTTPClient(proxyAdapter constant.Proxy, timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				h, pStr, splitErr := net.SplitHostPort(addr)
				if splitErr != nil {
					return nil, splitErr
				}
				pUint, parseErr := strconv.ParseUint(pStr, 10, 16)
				if parseErr != nil {
					return nil, parseErr
				}
				md := &constant.Metadata{Host: h, DstPort: uint16(pUint), Type: constant.HTTP}
				return proxyAdapter.DialContext(ctx, md)
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

func fetchUnlockProbe(runtime UnlockRuntime, target string, headers map[string]string) (*unlockHTTPResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), runtime.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	for key, value := range headers {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		req.Header.Set(key, value)
	}

	resp, err := runtime.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
	if err != nil {
		return nil, err
	}
	finalURL := target
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	return &unlockHTTPResponse{StatusCode: resp.StatusCode, FinalURL: finalURL, Body: strings.ToLower(string(body)), Header: resp.Header.Clone()}, nil
}

func containsAny(text string, needles []string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(text, strings.ToLower(strings.TrimSpace(needle))) {
			return true
		}
	}
	return false
}
