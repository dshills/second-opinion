package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GoogleProvider implements the Provider interface for Google AI (Gemini)
type GoogleProvider struct {
	apiKey      string
	model       string
	temperature float64
	maxTokens   int
	retryConfig RetryConfig
	httpClient  *http.Client
}

// NewGoogleProvider creates a new Google AI provider
func NewGoogleProvider(config Config) (*GoogleProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("the Google API key is required")
	}

	model := config.Model
	if model == "" {
		model = "gemini-1.5-flash"
	}

	temperature := config.Temperature
	if temperature == 0 {
		temperature = 0.3
	}

	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	return &GoogleProvider{
		apiKey:      config.APIKey,
		model:       model,
		temperature: temperature,
		maxTokens:   maxTokens,
		retryConfig: DefaultRetryConfig(),
		httpClient:  SharedHTTPClient,
	}, nil
}

// Analyze sends a prompt to Google AI and returns the response
func (p *GoogleProvider) Analyze(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.model, p.apiKey)

	requestBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{
						"text": prompt,
					},
				},
			},
		},
		"systemInstruction": map[string]any{
			"parts": []map[string]string{
				{
					"text": "You are an expert code reviewer and git analysis assistant. Provide clear, actionable feedback.",
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
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
		// Redact API key from error message if present
		errMsg := string(body)
		if p.apiKey != "" && len(p.apiKey) > 8 {
			errMsg = fmt.Sprintf("Google AI API error (status %d): [response body redacted for security]", resp.StatusCode)
		}
		return "", fmt.Errorf("%s", errMsg)
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Google AI")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

// Name returns the provider name
func (p *GoogleProvider) Name() string {
	return "google"
}
