package node

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sublink/models"
)

// maxRawSubscriptionBytes 限制原样转发的订阅体积，避免异常上游耗尽内存。
const maxRawSubscriptionBytes = 10 << 20

// RawSubscription 原样转发用的上游订阅响应。
type RawSubscription struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// FetchRawSubscription 拉取上游订阅内容，不做任何解析，供代理订阅原样转发。
func FetchRawSubscription(ctx context.Context, sourceURL string, downloadWithProxy bool, proxyLink, userAgent string, requestHeaders models.AirportRequestHeaders, skipTLSVerify bool) (*RawSubscription, error) {
	client := NewSubscriptionHTTPClient(downloadWithProxy, proxyLink, skipTLSVerify)

	// 与定时更新机场订阅保持一致，先还原 happ:// 加密链接。
	resolvedURL, err := resolveHappSubscriptionURL(sourceURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resolvedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create upstream subscription request: %w", err)
	}
	applyRequestHeaders(req, userAgent, requestHeaders)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("upstream subscription returned status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxRawSubscriptionBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream subscription: %w", err)
	}
	if len(body) > maxRawSubscriptionBytes {
		return nil, fmt.Errorf("upstream subscription content exceeds the %d byte limit", maxRawSubscriptionBytes)
	}

	return &RawSubscription{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       body,
	}, nil
}
