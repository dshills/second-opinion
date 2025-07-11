package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewOpenAIProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		expectModel string
		expectTemp  float64
		expectMax   int
	}{
		{
			name: "Valid config with all fields",
			config: Config{
				APIKey:      "test-key",
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   2048,
			},
			expectError: false,
			expectModel: "gpt-4",
			expectTemp:  0.7,
			expectMax:   2048,
		},
		{
			name: "Missing API key",
			config: Config{
				Model: "gpt-4",
			},
			expectError: true,
		},
		{
			name: "Default values",
			config: Config{
				APIKey: "test-key",
			},
			expectError: false,
			expectModel: "gpt-4o-mini",
			expectTemp:  0,
			expectMax:   4096,
		},
		{
			name: "Zero temperature explicitly set",
			config: Config{
				APIKey:      "test-key",
				Temperature: 0,
			},
			expectError: false,
			expectModel: "gpt-4o-mini",
			expectTemp:  0,
			expectMax:   4096,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOpenAIProvider(tt.config)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if provider.model != tt.expectModel {
				t.Errorf("model = %s, want %s", provider.model, tt.expectModel)
			}
			if provider.temperature != tt.expectTemp {
				t.Errorf("temperature = %f, want %f", provider.temperature, tt.expectTemp)
			}
			if provider.maxTokens != tt.expectMax {
				t.Errorf("maxTokens = %d, want %d", provider.maxTokens, tt.expectMax)
			}
		})
	}
}

func TestOpenAIProvider_isNewGenerationModel(t *testing.T) {
	tests := []struct {
		model    string
		expected bool
	}{
		{"gpt-4", false},
		{"gpt-4o-mini", false},
		{"o3-mini", true},
		{"O3-MINI", true},
		{"o4", true},
		{"O4", true},
		{"gpt-3.5-turbo", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			p := &OpenAIProvider{model: tt.model}
			if got := p.isNewGenerationModel(); got != tt.expected {
				t.Errorf("isNewGenerationModel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOpenAIProvider_supportsCustomTemperature(t *testing.T) {
	tests := []struct {
		model    string
		expected bool
	}{
		{"gpt-4", true},
		{"gpt-4o-mini", true},
		{"o3-mini", false},
		{"O3", false},
		{"o4", false},
		{"O4-preview", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			p := &OpenAIProvider{model: tt.model}
			if got := p.supportsCustomTemperature(); got != tt.expected {
				t.Errorf("supportsCustomTemperature() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOpenAIProvider_Analyze(t *testing.T) {
	tests := []struct {
		name         string
		model        string
		temperature  float64
		maxTokens    int
		serverStatus int
		serverResp   string
		expectError  bool
		expectResult string
	}{
		{
			name:         "Successful analysis with standard model",
			model:        "gpt-4",
			temperature:  0.7,
			maxTokens:    2048,
			serverStatus: http.StatusOK,
			serverResp: `{
				"choices": [{
					"message": {
						"content": "This is a test response"
					}
				}]
			}`,
			expectError:  false,
			expectResult: "This is a test response",
		},
		{
			name:         "Successful analysis with o3 model",
			model:        "o3-mini",
			temperature:  0.7, // Should be ignored
			maxTokens:    2048,
			serverStatus: http.StatusOK,
			serverResp: `{
				"choices": [{
					"message": {
						"content": "O3 model response"
					}
				}]
			}`,
			expectError:  false,
			expectResult: "O3 model response",
		},
		{
			name:         "API error",
			model:        "gpt-4",
			temperature:  0.7,
			maxTokens:    2048,
			serverStatus: http.StatusBadRequest,
			serverResp:   `{"error": {"message": "Invalid request"}}`,
			expectError:  true,
		},
		{
			name:         "Empty choices",
			model:        "gpt-4",
			temperature:  0.7,
			maxTokens:    2048,
			serverStatus: http.StatusOK,
			serverResp:   `{"choices": []}`,
			expectError:  true,
		},
		{
			name:         "Invalid JSON response",
			model:        "gpt-4",
			temperature:  0.7,
			maxTokens:    2048,
			serverStatus: http.StatusOK,
			serverResp:   `invalid json`,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request headers
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Content-Type = %s, want application/json", r.Header.Get("Content-Type"))
				}
				if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
					t.Error("Missing or invalid Authorization header")
				}

				// Verify request body
				var reqBody map[string]any
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Fatalf("Failed to decode request body: %v", err)
				}

				// Check model
				if reqBody["model"] != tt.model {
					t.Errorf("model = %s, want %s", reqBody["model"], tt.model)
				}

				// Check temperature for non-o3/o4 models
				if !strings.Contains(strings.ToLower(tt.model), "o3") && !strings.Contains(strings.ToLower(tt.model), "o4") {
					if temp, ok := reqBody["temperature"].(float64); !ok || temp != tt.temperature {
						t.Errorf("temperature = %v, want %f", reqBody["temperature"], tt.temperature)
					}
				} else {
					// o3/o4 models should not have temperature set
					if _, ok := reqBody["temperature"]; ok {
						t.Error("temperature should not be set for o3/o4 models")
					}
				}

				// Check tokens parameter
				if strings.Contains(strings.ToLower(tt.model), "o3") || strings.Contains(strings.ToLower(tt.model), "o4") {
					if tokens, ok := reqBody["max_completion_tokens"].(float64); !ok || int(tokens) != tt.maxTokens {
						t.Errorf("max_completion_tokens = %v, want %d", reqBody["max_completion_tokens"], tt.maxTokens)
					}
					if _, ok := reqBody["max_tokens"]; ok {
						t.Error("max_tokens should not be set for o3/o4 models")
					}
				} else {
					if tokens, ok := reqBody["max_tokens"].(float64); !ok || int(tokens) != tt.maxTokens {
						t.Errorf("max_tokens = %v, want %d", reqBody["max_tokens"], tt.maxTokens)
					}
					if _, ok := reqBody["max_completion_tokens"]; ok {
						t.Error("max_completion_tokens should not be set for non-o3/o4 models")
					}
				}

				// Send response
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			// Override OpenAI URL for testing
			originalURL := OpenAIURL
			defer func() {
				// This won't work because OpenAIURL is a const, but we'll handle it differently
				_ = originalURL
			}()

			// Create provider with test server
			provider := &OpenAIProvider{
				apiKey:      "test-key",
				model:       tt.model,
				temperature: tt.temperature,
				maxTokens:   tt.maxTokens,
				retryConfig: RetryConfig{
					MaxRetries:      1,
					BaseDelay:       10 * time.Millisecond,
					MaxDelay:        100 * time.Millisecond,
					BackoffMultiple: 2,
				},
				httpClient: &http.Client{},
			}

			// Create a custom HTTP client that redirects to our test server
			provider.httpClient = &http.Client{
				Transport: &testTransport{
					testServer: server,
				},
			}

			ctx := context.Background()
			result, err := provider.Analyze(ctx, "Test prompt")

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result != tt.expectResult {
					t.Errorf("result = %s, want %s", result, tt.expectResult)
				}
			}
		})
	}
}

func TestOpenAIProvider_Name(t *testing.T) {
	provider := &OpenAIProvider{}
	if name := provider.Name(); name != "openai" {
		t.Errorf("Name() = %s, want openai", name)
	}
}

// testTransport redirects requests to the test server
type testTransport struct {
	testServer *httptest.Server
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace the URL with our test server URL
	testURL := t.testServer.URL
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(testURL, "http://")
	return http.DefaultTransport.RoundTrip(req)
}
