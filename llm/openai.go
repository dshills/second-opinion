package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	OpenAIURL = "https://api.openai.com/v1/chat/completions"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey      string
	model       string
	temperature float64
	maxTokens   int
	retryConfig RetryConfig
	httpClient  *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config Config) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	model := config.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	temperature := config.Temperature
	// Only set default if temperature wasn't explicitly set in config
	if temperature == 0 && config.Temperature == 0 {
		temperature = 0.3
	}

	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	return &OpenAIProvider{
		apiKey:      config.APIKey,
		model:       model,
		temperature: temperature,
		maxTokens:   maxTokens,
		retryConfig: DefaultRetryConfig(),
		httpClient:  SharedHTTPClient,
	}, nil
}

// Analyze sends a prompt to OpenAI and returns the response
func (p *OpenAIProvider) Analyze(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]any{
		"model": p.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are an expert code reviewer and git analysis assistant. Provide clear, actionable feedback.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": p.temperature,
		"max_tokens":  p.maxTokens,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", OpenAIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := RetryableHTTPRequest(ctx, p.httpClient, req, p.retryConfig)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return result.Choices[0].Message.Content, nil
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}
