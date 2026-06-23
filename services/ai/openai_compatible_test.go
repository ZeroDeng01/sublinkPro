package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConnectionUsesResponsesEndpointWithHiPromptAndCappedTokens(t *testing.T) {
	t.Helper()

	type requestPayload struct {
		Model       string  `json:"model"`
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"max_output_tokens"`
		Stream      bool    `json:"stream"`
		Input       []struct {
			Role    string `json:"role"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"input"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" {
			t.Fatalf("expected /v1/responses, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
			t.Fatalf("unexpected authorization header: %q", got)
		}
		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Fatalf("unexpected accept header: %q", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		defer func() { _ = r.Body.Close() }()

		var payload requestPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal request payload: %v", err)
		}
		if payload.Model != "gpt-test" {
			t.Fatalf("unexpected model: %q", payload.Model)
		}
		if payload.MaxTokens != connectionTestMaxTokens {
			t.Fatalf("expected capped max tokens %d, got %d", connectionTestMaxTokens, payload.MaxTokens)
		}
		if !payload.Stream {
			t.Fatal("expected stream=true")
		}
		if len(payload.Input) != 1 {
			t.Fatalf("expected 1 input message, got %d", len(payload.Input))
		}
		if payload.Input[0].Role != "user" {
			t.Fatalf("expected user role, got %q", payload.Input[0].Role)
		}
		if len(payload.Input[0].Content) != 1 {
			t.Fatalf("expected 1 content item, got %d", len(payload.Input[0].Content))
		}
		if payload.Input[0].Content[0].Type != "input_text" {
			t.Fatalf("unexpected content type: %q", payload.Input[0].Content[0].Type)
		}
		if payload.Input[0].Content[0].Text != "hi" {
			t.Fatalf("expected hi prompt, got %q", payload.Input[0].Content[0].Text)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: response.output_text.delta\n"))
		_, _ = w.Write([]byte("data: {\"delta\":\"hello\"}\n\n"))
		_, _ = w.Write([]byte("event: response.completed\n"))
		_, _ = w.Write([]byte("data: {\"response\":{\"status\":\"completed\",\"usage\":{\"output_tokens\":5},\"output\":[{\"content\":[{\"type\":\"output_text\",\"text\":\"hello\"}]}]}}\n\n"))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		BaseURL:     server.URL + "/v1",
		APIKey:      "sk-test",
		Model:       "gpt-test",
		Temperature: 0.2,
		MaxTokens:   1200,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	result, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("test connection: %v", err)
	}
	if result.Message != "hello" {
		t.Fatalf("expected hello, got %q", result.Message)
	}
	if result.FinishReason != "completed" {
		t.Fatalf("expected completed finish reason, got %q", result.FinishReason)
	}
	if result.Usage["output_tokens"] != float64(5) {
		t.Fatalf("expected output_tokens=5, got %#v", result.Usage["output_tokens"])
	}
}

func TestConnectionKeepsSmallerConfiguredMaxTokens(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		defer func() { _ = r.Body.Close() }()

		var payload struct {
			MaxTokens int `json:"max_output_tokens"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal request payload: %v", err)
		}
		if payload.MaxTokens != 8 {
			t.Fatalf("expected configured max tokens 8, got %d", payload.MaxTokens)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: response.output_text.delta\n"))
		_, _ = w.Write([]byte("data: {\"delta\":\"ok\"}\n\n"))
		_, _ = w.Write([]byte("event: response.completed\n"))
		_, _ = w.Write([]byte("data: {\"response\":{\"status\":\"completed\"}}\n\n"))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		BaseURL:     server.URL,
		APIKey:      "sk-test",
		Model:       "gpt-test",
		Temperature: 0.2,
		MaxTokens:   8,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if _, err := client.TestConnection(context.Background()); err != nil {
		t.Fatalf("test connection: %v", err)
	}
}

func TestConnectionUsesChatCompletionsWhenConfigured(t *testing.T) {
	t.Helper()

	type requestPayload struct {
		Model         string    `json:"model"`
		Messages      []Message `json:"messages"`
		MaxTokens     int       `json:"max_tokens"`
		Stream        bool      `json:"stream"`
		StreamOptions struct {
			IncludeUsage bool `json:"include_usage"`
		} `json:"stream_options"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("expected /v1/chat/completions, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Fatalf("unexpected accept header: %q", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		defer func() { _ = r.Body.Close() }()

		var payload requestPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal request payload: %v", err)
		}
		if payload.Model != "gpt-test" {
			t.Fatalf("unexpected model: %q", payload.Model)
		}
		if payload.MaxTokens != connectionTestMaxTokens {
			t.Fatalf("expected capped max tokens %d, got %d", connectionTestMaxTokens, payload.MaxTokens)
		}
		if !payload.Stream {
			t.Fatal("expected stream=true")
		}
		if !payload.StreamOptions.IncludeUsage {
			t.Fatal("expected include_usage=true")
		}
		if len(payload.Messages) != 1 || payload.Messages[0].Role != "user" || payload.Messages[0].Content != "hi" {
			t.Fatalf("unexpected messages: %#v", payload.Messages)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\" chat\"},\"finish_reason\":\"stop\"}],\"usage\":{\"completion_tokens\":2}}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		BaseURL:     server.URL + "/v1",
		APIKey:      "sk-test",
		Model:       "gpt-test",
		RequestType: RequestTypeChatCompletions,
		Temperature: 0.2,
		MaxTokens:   1200,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	result, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("test connection: %v", err)
	}
	if result.Message != "hello chat" {
		t.Fatalf("expected hello chat, got %q", result.Message)
	}
	if result.FinishReason != "stop" {
		t.Fatalf("expected stop finish reason, got %q", result.FinishReason)
	}
	if result.Usage["completion_tokens"] != float64(2) {
		t.Fatalf("expected completion_tokens=2, got %#v", result.Usage["completion_tokens"])
	}
}
