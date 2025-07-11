package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	OpenAIURL      = "https://api.openai.com/v1/chat/completions"
	openAIProvider = "openai"
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

	// Use temperature as provided - OpenAI supports 0 temperature for deterministic output
	temperature := config.Temperature

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

// isNewGenerationModel checks if the model is o3/o4 series that requires max_completion_tokens and has temperature restrictions
func (p *OpenAIProvider) isNewGenerationModel() bool {
	modelLower := strings.ToLower(p.model)
	return strings.Contains(modelLower, "o3") || strings.Contains(modelLower, "o4")
}

// supportsCustomTemperature checks if the model supports custom temperature values
func (p *OpenAIProvider) supportsCustomTemperature() bool {
	return !p.isNewGenerationModel() // o3/o4 models only support default temperature of 1.0
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
	}

	// Set temperature only for models that support custom values
	if p.supportsCustomTemperature() {
		requestBody["temperature"] = p.temperature
	}
	// o3/o4 models use default temperature of 1.0 (no need to set explicitly)

	// Use max_completion_tokens for o3/o4 models, max_tokens for others
	if p.isNewGenerationModel() {
		requestBody["max_completion_tokens"] = p.maxTokens
	} else {
		requestBody["max_tokens"] = p.maxTokens
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
	return openAIProvider
}
