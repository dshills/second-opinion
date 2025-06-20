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
	defaultOllamaEndpoint = "http://localhost:11434"
	defaultOllamaModel    = "devstral:latest"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	endpoint    string
	model       string
	temperature float64
	maxTokens   int
	retryConfig RetryConfig
	httpClient  *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config Config) (*OllamaProvider, error) {
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = defaultOllamaEndpoint
	}

	model := config.Model
	if model == "" {
		model = defaultOllamaModel
	}

	temperature := config.Temperature
	if temperature == 0 {
		temperature = 0.3
	}

	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	return &OllamaProvider{
		endpoint:    endpoint,
		model:       model,
		temperature: temperature,
		maxTokens:   maxTokens,
		retryConfig: DefaultRetryConfig(),
		httpClient:  SharedHTTPClient,
	}, nil
}

// Analyze sends a prompt to Ollama and returns the response
func (p *OllamaProvider) Analyze(ctx context.Context, prompt string) (string, error) {
	requestBody := map[string]any{
		"model":  p.model,
		"prompt": prompt,
		"system": "You are an expert code reviewer and git analysis assistant. Provide clear, actionable feedback.",
		"stream": false,
		"options": map[string]any{
			"temperature":    p.temperature,
			"num_predict":    p.maxTokens,
			"top_k":          40,
			"top_p":          0.9,
			"repeat_last_n":  64,
			"repeat_penalty": 1.1,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
		return "", fmt.Errorf("the Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Response string `json:"response"`
		Error    string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("the Ollama error: %s", result.Error)
	}

	return result.Response, nil
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return "ollama"
}
