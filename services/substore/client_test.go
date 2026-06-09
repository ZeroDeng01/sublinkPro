package substore

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientConvertPostsParseRequest(t *testing.T) {
	var gotRequest parseRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != parsePath {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"par_res":"loon output"}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, Timeout: time.Second, MaxResponseBytes: 1024, AllowedTargets: []string{"loon"}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	converted, err := client.Convert(context.Background(), "proxies:\n  - name: test\n", "loon")
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if gotRequest.Client != "Loon" {
		t.Fatalf("expected Loon target, got %q", gotRequest.Client)
	}
	if !strings.Contains(gotRequest.Data, "proxies:") {
		t.Fatalf("expected bridge yaml in request, got %q", gotRequest.Data)
	}
	if converted.Body != "loon output" {
		t.Fatalf("unexpected converted body: %q", converted.Body)
	}
	if converted.ContentType != "text/plain; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", converted.ContentType)
	}
}

func TestClientConvertMapsQuanxToQX(t *testing.T) {
	var target string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req parseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		target = req.Client
		_, _ = w.Write([]byte(`{"status":"success","data":{"par_res":"qx output"}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, AllowedTargets: []string{"quanx"}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if _, err := client.Convert(context.Background(), "proxies: []", "quanx"); err != nil {
		t.Fatalf("convert: %v", err)
	}
	if target != "QX" {
		t.Fatalf("expected QX target, got %q", target)
	}
}

func TestClientConvertAllowsV2rayURIWhenURIIsAllowed(t *testing.T) {
	var target string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req parseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		target = req.Client
		_, _ = w.Write([]byte(`{"status":"success","data":{"par_res":"uri output"}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, AllowedTargets: []string{"uri"}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if _, err := client.Convert(context.Background(), "proxies: []", "v2ray-uri"); err != nil {
		t.Fatalf("convert v2ray-uri with uri allowlist: %v", err)
	}
	if target != "URI" {
		t.Fatalf("expected URI target, got %q", target)
	}
}

func TestClientConvertRejectsDisabledTarget(t *testing.T) {
	client, err := NewClient(Config{BaseURL: "http://127.0.0.1:1", AllowedTargets: []string{"loon"}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, err = client.Convert(context.Background(), "proxies: []", "stash")
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled target error, got %v", err)
	}
}

func TestClientConvertReportsSidecarFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway"))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, AllowedTargets: []string{"loon"}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, err = client.Convert(context.Background(), "proxies: []", "loon")
	if err == nil || !strings.Contains(err.Error(), "HTTP 502") {
		t.Fatalf("expected HTTP status error, got %v", err)
	}
}

func TestClientConvertRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","data":{"par_res":"too large"}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, MaxResponseBytes: 10, AllowedTargets: []string{"loon"}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, err = client.Convert(context.Background(), "proxies: []", "loon")
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected oversized response error, got %v", err)
	}
}
