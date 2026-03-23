package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sublink/utils"
	"time"
)

type ClientConfig struct {
	BaseURL      string
	APIKey       string
	Model        string
	Temperature  float64
	MaxTokens    int
	ExtraHeaders map[string]string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponsesEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data,omitempty"`
}

type responseCompletedEventPayload struct {
	Response struct {
		Status string                 `json:"status,omitempty"`
		Usage  map[string]interface{} `json:"usage,omitempty"`
		Output []struct {
			Type    string `json:"type,omitempty"`
			Content []struct {
				Type string `json:"type,omitempty"`
				Text string `json:"text,omitempty"`
			} `json:"content,omitempty"`
		} `json:"output,omitempty"`
	} `json:"response,omitempty"`
}

type responsesRequest struct {
	Model       string                   `json:"model"`
	Input       []map[string]interface{} `json:"input"`
	Temperature float64                  `json:"temperature,omitempty"`
	MaxTokens   int                      `json:"max_output_tokens,omitempty"`
	Stream      bool                     `json:"stream"`
}

type chatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage map[string]interface{} `json:"usage"`
}

type Client struct {
	baseURL      string
	apiKey       string
	model        string
	temperature  float64
	maxTokens    int
	extraHeaders map[string]string
	httpClient   *http.Client
}

type TestResult struct {
	Message      string                 `json:"message"`
	Model        string                 `json:"model"`
	BaseURL      string                 `json:"baseUrl"`
	LatencyMs    int64                  `json:"latencyMs"`
	Usage        map[string]interface{} `json:"usage,omitempty"`
	FinishReason string                 `json:"finishReason,omitempty"`
}

type modelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func NormalizeBaseURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("AI Base URL 无效")
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("AI Base URL 必须为 http 或 https")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("AI Base URL 缺少主机")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return parsed.String(), nil
}

func NewClient(cfg ClientConfig) (*Client, error) {
	baseURL, err := NormalizeBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("AI Base URL 不能为空")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, fmt.Errorf("AI 模型不能为空")
	}
	if cfg.Temperature < 0 || cfg.Temperature > 2 {
		return nil, fmt.Errorf("temperature 必须在 0 到 2 之间")
	}
	if cfg.MaxTokens < 0 {
		return nil, fmt.Errorf("max_tokens 不能小于 0")
	}
	return &Client{
		baseURL:      baseURL,
		apiKey:       strings.TrimSpace(cfg.APIKey),
		model:        strings.TrimSpace(cfg.Model),
		temperature:  cfg.Temperature,
		maxTokens:    cfg.MaxTokens,
		extraHeaders: cfg.ExtraHeaders,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}, nil
}

func (c *Client) streamingHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 45 * time.Second
	transport.IdleConnTimeout = 90 * time.Second
	return &http.Client{Transport: transport}
}

func (c *Client) endpointURL() (string, error) {
	parsed, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}
	parsed.Path = path.Join(parsed.Path, "chat/completions")
	return parsed.String(), nil
}

func (c *Client) responsesEndpointURL() (string, error) {
	parsed, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}
	parsed.Path = path.Join(parsed.Path, "responses")
	return parsed.String(), nil
}

func buildModelsEndpointURL(baseURL string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	parsed.Path = path.Join(parsed.Path, "models")
	return parsed.String(), nil
}

func redactHeaderValue(key string, value string) string {
	lowerKey := strings.ToLower(strings.TrimSpace(key))
	if value == "" {
		return value
	}
	if lowerKey == "authorization" || strings.Contains(lowerKey, "token") || strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "key") {
		return "[REDACTED]"
	}
	return value
}

func redactHeaders(headers http.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	redacted := make(map[string]string, len(headers))
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}
		redacted[key] = redactHeaderValue(key, strings.Join(values, ","))
	}
	return redacted
}

func debugLogAIRequest(method string, endpoint string, headers http.Header, body []byte) {
	utils.Debug("AI upstream request: method=%s url=%s headers=%s body=%s", method, endpoint, mustMarshalLogJSON(redactHeaders(headers)), string(body))
}

func debugLogAIResponse(endpoint string, statusCode int, body []byte) {
	utils.Debug("AI upstream response: url=%s status=%d body=%s", endpoint, statusCode, string(body))
}

func debugLogAIStreamEvent(eventName string, data string) {
	utils.Debug("AI upstream stream event: event=%s data=%s", eventName, data)
}

func mustMarshalLogJSON(value interface{}) string {
	if value == nil {
		return "null"
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}

func (c *Client) newJSONRequest(ctx context.Context, method string, endpoint string, payload interface{}) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	for key, value := range c.extraHeaders {
		if strings.TrimSpace(key) == "" || value == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	debugLogAIRequest(method, endpoint, req.Header, body)
	return req, nil
}

func (c *Client) CreateChatCompletion(ctx context.Context, messages []Message) (string, string, map[string]interface{}, error) {
	endpoint, err := c.endpointURL()
	if err != nil {
		return "", "", nil, err
	}
	payload := chatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
		Stream:      false,
	}
	req, err := c.newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return "", "", nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", nil, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", nil, err
	}
	debugLogAIResponse(endpoint, resp.StatusCode, responseBody)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", nil, fmt.Errorf("AI 服务返回 HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	var parsed chatCompletionResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return "", "", nil, fmt.Errorf("AI 响应解析失败: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", "", nil, fmt.Errorf("AI 响应缺少 choices")
	}
	utils.Debug("AI upstream parsed usage: url=%s finish_reason=%s usage=%s", endpoint, parsed.Choices[0].FinishReason, mustMarshalLogJSON(parsed.Usage))
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return "", parsed.Choices[0].FinishReason, parsed.Usage, fmt.Errorf("AI 响应内容为空")
	}
	return content, parsed.Choices[0].FinishReason, parsed.Usage, nil
}

func (c *Client) StreamResponses(ctx context.Context, messages []Message, onEvent func(ResponsesEvent) error) (string, string, map[string]interface{}, error) {
	endpoint, err := c.responsesEndpointURL()
	if err != nil {
		return "", "", nil, err
	}
	input := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		input = append(input, map[string]interface{}{
			"role": msg.Role,
			"content": []map[string]string{{
				"type": "input_text",
				"text": msg.Content,
			}},
		})
	}
	payload := responsesRequest{
		Model:       c.model,
		Input:       input,
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
		Stream:      true,
	}
	req, err := c.newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return "", "", nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.streamingHTTPClient().Do(req)
	if err != nil {
		return "", "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		debugLogAIResponse(endpoint, resp.StatusCode, responseBody)
		return "", "", nil, fmt.Errorf("AI /responses 返回 HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	utils.Debug("AI upstream stream connected: url=%s status=%d", endpoint, resp.StatusCode)

	reader := bufio.NewReader(resp.Body)
	var eventName string
	var dataLines []string
	var builder strings.Builder
	finishReason := ""
	var usage map[string]interface{}

	flushEvent := func() error {
		if eventName == "" && len(dataLines) == 0 {
			return nil
		}
		dataText := strings.Join(dataLines, "\n")
		trimmedData := strings.TrimSpace(dataText)
		if trimmedData == "[DONE]" {
			eventName = "done"
		}
		debugLogAIStreamEvent(eventName, trimmedData)
		if eventName == "response.output_text.delta" && trimmedData != "" {
			var payload struct {
				Delta string `json:"delta"`
			}
			if err := json.Unmarshal([]byte(trimmedData), &payload); err == nil {
				builder.WriteString(payload.Delta)
			}
		}
		if eventName == "response.completed" && trimmedData != "" {
			var payload responseCompletedEventPayload
			if err := json.Unmarshal([]byte(trimmedData), &payload); err == nil {
				if strings.TrimSpace(payload.Response.Status) != "" {
					finishReason = strings.TrimSpace(payload.Response.Status)
				}
				if len(payload.Response.Usage) > 0 {
					usage = payload.Response.Usage
				}
				var completedTextBuilder strings.Builder
				for _, outputItem := range payload.Response.Output {
					for _, contentItem := range outputItem.Content {
						if contentItem.Type == "output_text" && strings.TrimSpace(contentItem.Text) != "" {
							completedTextBuilder.WriteString(contentItem.Text)
						}
					}
				}
				if completedText := strings.TrimSpace(completedTextBuilder.String()); completedText != "" {
					builder.Reset()
					builder.WriteString(completedText)
				}
				utils.Debug("AI upstream stream completed: url=%s finish_reason=%s usage=%s", endpoint, finishReason, mustMarshalLogJSON(usage))
			}
		}
		if onEvent != nil {
			if err := onEvent(ResponsesEvent{Event: eventName, Data: json.RawMessage([]byte(trimmedData))}); err != nil {
				return err
			}
		}
		eventName = ""
		dataLines = nil
		return nil
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return builder.String(), finishReason, usage, err
		}
		trimmedLine := strings.TrimRight(line, "\r\n")
		if trimmedLine == "" {
			if err := flushEvent(); err != nil {
				return builder.String(), finishReason, usage, err
			}
		} else if strings.HasPrefix(trimmedLine, "event:") {
			eventName = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "event:"))
		} else if strings.HasPrefix(trimmedLine, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(trimmedLine, "data:")))
		}
		if err == io.EOF {
			break
		}
	}
	if err := flushEvent(); err != nil {
		return builder.String(), finishReason, usage, err
	}
	content := strings.TrimSpace(builder.String())
	if content == "" {
		return "", finishReason, usage, fmt.Errorf("AI /responses 流未返回有效内容")
	}
	return content, finishReason, usage, nil
}

func (c *Client) TestConnection(ctx context.Context) (*TestResult, error) {
	start := time.Now()
	content, finishReason, usage, err := c.CreateChatCompletion(ctx, []Message{{
		Role:    "system",
		Content: "Reply with exactly OK.",
	}, {
		Role:    "user",
		Content: "OK",
	}})
	if err != nil {
		return nil, err
	}
	return &TestResult{
		Message:      content,
		Model:        c.model,
		BaseURL:      c.baseURL,
		LatencyMs:    time.Since(start).Milliseconds(),
		Usage:        usage,
		FinishReason: finishReason,
	}, nil
}

func DiscoverModels(ctx context.Context, cfg ClientConfig) ([]string, error) {
	baseURL, err := NormalizeBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("AI Base URL 不能为空")
	}
	endpoint, err := buildModelsEndpointURL(baseURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.APIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(cfg.APIKey))
	}
	for key, value := range cfg.ExtraHeaders {
		if strings.TrimSpace(key) == "" || value == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	resp, err := (&http.Client{Timeout: 45 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("AI /models 返回 HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	var parsed modelsResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return nil, fmt.Errorf("AI 模型列表解析失败: %w", err)
	}
	seen := make(map[string]struct{}, len(parsed.Data))
	models := make([]string, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		modelID := strings.TrimSpace(item.ID)
		if modelID == "" {
			continue
		}
		if _, exists := seen[modelID]; exists {
			continue
		}
		seen[modelID] = struct{}{}
		models = append(models, modelID)
	}
	return models, nil
}
