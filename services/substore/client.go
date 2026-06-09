package substore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const parsePath = "/api/proxy/parse"

var ErrNotConfigured = errors.New("sub-store sidecar is not configured")

type Config struct {
	BaseURL          string
	Timeout          time.Duration
	MaxResponseBytes int64
	AllowedTargets   []string
}

type Client struct {
	baseURL          string
	httpClient       *http.Client
	maxResponseBytes int64
	allowedTargets   map[string]struct{}
}

type ConvertedSubscription struct {
	Body        string
	ContentType string
}

type parseRequest struct {
	Data   string `json:"data"`
	Client string `json:"client"`
}

type parseResponse struct {
	Status string `json:"status"`
	Data   struct {
		ParRes string `json:"par_res"`
	} `json:"data"`
	Error *parseError `json:"error"`
}

type parseError struct {
	Code    string `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
	Details any    `json:"details"`
}

func NewClient(cfg Config) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, ErrNotConfigured
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.MaxResponseBytes <= 0 {
		cfg.MaxResponseBytes = 8 << 20
	}

	allowedTargets := make(map[string]struct{}, len(cfg.AllowedTargets))
	for _, target := range cfg.AllowedTargets {
		target = normalizeTargetKey(target)
		if target != "" {
			allowedTargets[target] = struct{}{}
		}
	}

	return &Client{
		baseURL:          baseURL,
		httpClient:       &http.Client{Timeout: cfg.Timeout},
		maxResponseBytes: cfg.MaxResponseBytes,
		allowedTargets:   allowedTargets,
	}, nil
}

func (c *Client) Convert(ctx context.Context, sourceYAML, target string) (*ConvertedSubscription, error) {
	if c == nil || c.baseURL == "" {
		return nil, ErrNotConfigured
	}
	targetName, ok := ResolveTarget(target)
	if !ok {
		return nil, fmt.Errorf("unsupported subscription conversion target: %s", target)
	}
	if _, allowed := c.allowedTargets[canonicalTargetKey(target)]; !allowed {
		return nil, fmt.Errorf("subscription conversion target is disabled: %s", target)
	}

	payload, err := json.Marshal(parseRequest{Data: sourceYAML, Client: targetName})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+parsePath, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, c.maxResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > c.maxResponseBytes {
		return nil, fmt.Errorf("sub-store response exceeds %d bytes", c.maxResponseBytes)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("sub-store returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed parseResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode sub-store response: %w", err)
	}
	if parsed.Status != "success" {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return nil, fmt.Errorf("sub-store conversion failed: %s", parsed.Error.Message)
		}
		return nil, fmt.Errorf("sub-store conversion failed with status %q", parsed.Status)
	}
	if strings.TrimSpace(parsed.Data.ParRes) == "" {
		return nil, errors.New("sub-store returned empty conversion result")
	}

	return &ConvertedSubscription{
		Body:        parsed.Data.ParRes,
		ContentType: contentTypeForTarget(target),
	}, nil
}

func ResolveTarget(target string) (string, bool) {
	switch canonicalTargetKey(target) {
	case "stash":
		return "Stash", true
	case "mihomo", "clashmeta", "clash-meta":
		return "ClashMeta", true
	case "egern":
		return "Egern", true
	case "surfboard":
		return "Surfboard", true
	case "loon":
		return "Loon", true
	case "shadowrocket":
		return "Shadowrocket", true
	case "quanx", "qx", "quantumultx", "quantumult-x":
		return "QX", true
	case "singbox", "sing-box":
		return "sing-box", true
	case "json":
		return "JSON", true
	case "uri":
		return "URI", true
	}
	return "", false
}

func IsSupportedTarget(target string) bool {
	_, ok := ResolveTarget(target)
	return ok
}

func normalizeTargetKey(target string) string {
	return strings.ToLower(strings.TrimSpace(target))
}

func canonicalTargetKey(target string) string {
	switch normalizeTargetKey(target) {
	case "v2ray-uri", "v2rayuri":
		return "uri"
	case "singbox":
		return "sing-box"
	}
	return normalizeTargetKey(target)
}

func contentTypeForTarget(target string) string {
	switch canonicalTargetKey(target) {
	case "json", "sing-box":
		return "application/json; charset=utf-8"
	default:
		return "text/plain; charset=utf-8"
	}
}
