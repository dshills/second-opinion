package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	endpoint    string
	model       string
	temperature float64
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config Config) (*OllamaProvider, error) {
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	model := config.Model
	if model == "" {
		model = "llama3.2"
	}

	temperature := config.Temperature
	if temperature == 0 {
		temperature = 0.3
	}

	return &OllamaProvider{
		endpoint:    endpoint,
		model:       model,
		temperature: temperature,
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
			"temperature": p.temperature,
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

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
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
