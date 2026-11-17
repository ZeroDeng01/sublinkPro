package unlock

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"sublink/utils"
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
	RawBody    string
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
	client := utils.NewHTTPClientWithAdapter(proxyAdapter, timeout, &tls.Config{InsecureSkipVerify: true}) // #nosec G402 -- 解锁检测需要兼容目标站点的异常证书配置
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return http.ErrUseLastResponse
		}
		return nil
	}
	return client
}

func fetchUnlockProbe(runtime UnlockRuntime, target string, headers map[string]string) (*unlockHTTPResponse, error) {
	return fetchUnlockProbeWithBodyLimit(runtime, target, headers, 32*1024)
}

func fetchUnlockProbeWithBodyLimit(runtime UnlockRuntime, target string, headers map[string]string, bodyLimit int64) (*unlockHTTPResponse, error) {
	return fetchUnlockRequest(runtime, http.MethodGet, target, headers, nil, bodyLimit)
}

func fetchUnlockRequest(runtime UnlockRuntime, method string, target string, headers map[string]string, body []byte, bodyLimit int64) (*unlockHTTPResponse, error) {
	if bodyLimit <= 0 {
		bodyLimit = 32 * 1024
	}

	ctx, cancel := context.WithTimeout(context.Background(), runtime.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(body))
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
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, bodyLimit))
	if err != nil {
		return nil, err
	}
	bodyText := string(bodyBytes)
	finalURL := target
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	return &unlockHTTPResponse{StatusCode: resp.StatusCode, FinalURL: finalURL, Body: strings.ToLower(bodyText), RawBody: bodyText, Header: resp.Header.Clone()}, nil
}

func containsAny(text string, needles []string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(text, strings.ToLower(strings.TrimSpace(needle))) {
			return true
		}
	}
	return false
}
