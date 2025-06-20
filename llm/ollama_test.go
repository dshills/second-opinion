package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestOllamaEndpointConnectivity tests basic connectivity to the Ollama endpoint
func TestOllamaEndpointConnectivity(t *testing.T) {
	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	
	t.Logf("Testing connectivity to Ollama endpoint: %s", endpoint)
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	// Test the base endpoint
	resp, err := client.Get(endpoint)
	if err != nil {
		t.Errorf("Failed to connect to Ollama endpoint: %v", err)
		t.Logf("Make sure Ollama is running at %s", endpoint)
		return
	}
	defer resp.Body.Close()
	
	t.Logf("Ollama endpoint response status: %d", resp.StatusCode)
	
	// Test the API endpoint
	apiResp, err := client.Get(endpoint + "/api/tags")
	if err != nil {
		t.Errorf("Failed to connect to Ollama API endpoint: %v", err)
		return
	}
	defer apiResp.Body.Close()
	
	if apiResp.StatusCode != http.StatusOK {
		t.Errorf("Ollama API returned non-OK status: %d", apiResp.StatusCode)
	}
}

// TestOllamaModelAvailability checks if the configured model is available
func TestOllamaModelAvailability(t *testing.T) {
	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3.2"
	}
	
	t.Logf("Checking availability of model: %s", model)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Get(endpoint + "/api/tags")
	if err != nil {
		t.Skipf("Cannot check models - Ollama not accessible: %v", err)
		return
	}
	defer resp.Body.Close()
	
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Errorf("Failed to parse models list: %v", err)
		return
	}
	
	modelFound := false
	availableModels := []string{}
	for _, m := range result.Models {
		availableModels = append(availableModels, m.Name)
		if strings.HasPrefix(m.Name, model) {
			modelFound = true
			t.Logf("Model %s is available", m.Name)
		}
	}
	
	if !modelFound {
		t.Errorf("Model %s not found. Available models: %v", model, availableModels)
	}
}

// TestOllamaProviderInitialization tests creating an Ollama provider
func TestOllamaProviderInitialization(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		wantErr  bool
		expected struct {
			endpoint string
			model    string
			temp     float64
		}
	}{
		{
			name: "default configuration",
			config: Config{
				Provider: "ollama",
			},
			wantErr: false,
			expected: struct {
				endpoint string
				model    string
				temp     float64
			}{
				endpoint: "http://localhost:11434",
				model:    "llama3.2",
				temp:     0.3,
			},
		},
		{
			name: "custom configuration",
			config: Config{
				Provider:    "ollama",
				Endpoint:    "http://ubuntu-ai-1.local:11434",
				Model:       "devstral",
				Temperature: 0.5,
			},
			wantErr: false,
			expected: struct {
				endpoint string
				model    string
				temp     float64
			}{
				endpoint: "http://ubuntu-ai-1.local:11434",
				model:    "devstral",
				temp:     0.5,
			},
		},
		{
			name: "env-based configuration",
			config: Config{
				Provider:    "ollama",
				Endpoint:    os.Getenv("OLLAMA_ENDPOINT"),
				Model:       os.Getenv("OLLAMA_MODEL"),
				Temperature: 0.3,
			},
			wantErr: false,
			expected: struct {
				endpoint string
				model    string
				temp     float64
			}{
				endpoint: os.Getenv("OLLAMA_ENDPOINT"),
				model:    os.Getenv("OLLAMA_MODEL"),
				temp:     0.3,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOllamaProvider(tt.config)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOllamaProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if err == nil {
				if provider.endpoint != tt.expected.endpoint && tt.expected.endpoint != "" {
					t.Errorf("Expected endpoint %s, got %s", tt.expected.endpoint, provider.endpoint)
				}
				if provider.model != tt.expected.model && tt.expected.model != "" {
					t.Errorf("Expected model %s, got %s", tt.expected.model, provider.model)
				}
				if provider.temperature != tt.expected.temp {
					t.Errorf("Expected temperature %f, got %f", tt.expected.temp, provider.temperature)
				}
			}
		})
	}
}

// TestOllamaSimpleGeneration tests a basic generation request
func TestOllamaSimpleGeneration(t *testing.T) {
	// Use env vars for real test
	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	model := os.Getenv("OLLAMA_MODEL")
	
	if endpoint == "" || model == "" {
		t.Skip("OLLAMA_ENDPOINT and OLLAMA_MODEL must be set for integration test")
	}
	
	provider, err := NewOllamaProvider(Config{
		Provider: "ollama",
		Endpoint: endpoint,
		Model:    model,
	})
	
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Simple test prompt
	result, err := provider.Analyze(ctx, "What is 2 + 2? Reply with just the number.")
	
	if err != nil {
		t.Errorf("Ollama generation failed: %v", err)
		t.Logf("Endpoint: %s, Model: %s", endpoint, model)
		return
	}
	
	t.Logf("Ollama response: %s", result)
	
	// Check if response contains "4"
	if !strings.Contains(result, "4") {
		t.Errorf("Expected response to contain '4', got: %s", result)
	}
}

// TestOllamaWithRetry tests Ollama with retry mechanism
func TestOllamaWithRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		t.Logf("Request attempt %d to %s", attempts, r.URL.Path)
		
		if attempts < 2 {
			// Simulate temporary failure
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "temporary failure"}`))
			return
		}
		
		// Successful response
		response := map[string]interface{}{
			"response": "Test successful after retry",
			"done":     true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	provider, err := NewOllamaProvider(Config{
		Provider: "ollama",
		Endpoint: server.URL,
		Model:    "test-model",
	})
	
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	
	ctx := context.Background()
	result, err := provider.Analyze(ctx, "test prompt")
	
	if err != nil {
		t.Errorf("Expected successful retry, got error: %v", err)
		return
	}
	
	if result != "Test successful after retry" {
		t.Errorf("Unexpected response: %s", result)
	}
	
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

// TestOllamaTimeout tests timeout handling
func TestOllamaTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	provider, err := NewOllamaProvider(Config{
		Provider: "ollama",
		Endpoint: server.URL,
		Model:    "test-model",
	})
	
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	
	// Use a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	_, err = provider.Analyze(ctx, "test prompt")
	
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected context deadline exceeded error, got: %v", err)
	}
}

// TestOllamaErrorHandling tests various error scenarios
func TestOllamaErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectedError  string
	}{
		{
			name: "404 not found",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Model not found"))
			},
			expectedError: "Ollama API error (status 404)",
		},
		{
			name: "invalid JSON response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			},
			expectedError: "failed to parse response",
		},
		{
			name: "error in response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				response := map[string]string{
					"error": "Model is currently loading",
				}
				json.NewEncoder(w).Encode(response)
			},
			expectedError: "Ollama error: Model is currently loading",
		},
		{
			name: "empty response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				response := map[string]string{
					"response": "",
				}
				json.NewEncoder(w).Encode(response)
			},
			expectedError: "", // Should succeed with empty response
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()
			
			provider, err := NewOllamaProvider(Config{
				Provider: "ollama",
				Endpoint: server.URL,
				Model:    "test-model",
			})
			
			if err != nil {
				t.Fatalf("Failed to create provider: %v", err)
			}
			
			ctx := context.Background()
			_, err = provider.Analyze(ctx, "test prompt")
			
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

// TestOllamaLargePrompt tests handling of large prompts
func TestOllamaLargePrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		prompt, ok := req["prompt"].(string)
		if !ok {
			t.Error("No prompt in request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		if len(prompt) < 1000 {
			t.Errorf("Expected large prompt, got %d characters", len(prompt))
		}
		
		response := map[string]interface{}{
			"response": "Processed large prompt successfully",
			"done":     true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	provider, err := NewOllamaProvider(Config{
		Provider: "ollama",
		Endpoint: server.URL,
		Model:    "test-model",
	})
	
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	
	// Create a large prompt
	largePrompt := strings.Repeat("This is a test sentence. ", 100)
	
	ctx := context.Background()
	result, err := provider.Analyze(ctx, largePrompt)
	
	if err != nil {
		t.Errorf("Failed to process large prompt: %v", err)
		return
	}
	
	if result != "Processed large prompt successfully" {
		t.Errorf("Unexpected response: %s", result)
	}
}

// TestOllamaRequestStructure tests that requests are properly formatted
func TestOllamaRequestStructure(t *testing.T) {
	var capturedRequest map[string]interface{}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedRequest); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		response := map[string]interface{}{
			"response": "OK",
			"done":     true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	provider, err := NewOllamaProvider(Config{
		Provider:    "ollama",
		Endpoint:    server.URL,
		Model:       "test-model",
		Temperature: 0.7,
	})
	
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	
	ctx := context.Background()
	_, err = provider.Analyze(ctx, "Test prompt")
	
	if err != nil {
		t.Errorf("Request failed: %v", err)
		return
	}
	
	// Verify request structure
	if capturedRequest["model"] != "test-model" {
		t.Errorf("Expected model 'test-model', got %v", capturedRequest["model"])
	}
	
	if capturedRequest["prompt"] != "Test prompt" {
		t.Errorf("Expected prompt 'Test prompt', got %v", capturedRequest["prompt"])
	}
	
	if capturedRequest["stream"] != false {
		t.Errorf("Expected stream=false, got %v", capturedRequest["stream"])
	}
	
	options, ok := capturedRequest["options"].(map[string]interface{})
	if !ok {
		t.Error("Expected options to be a map")
	} else {
		if options["temperature"] != 0.7 {
			t.Errorf("Expected temperature 0.7, got %v", options["temperature"])
		}
	}
	
	if capturedRequest["system"] == nil {
		t.Error("Expected system prompt to be set")
	}
}

// TestOllamaRealIntegration performs a real integration test if Ollama is available
func TestOllamaRealIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Load from environment
	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	model := os.Getenv("OLLAMA_MODEL")
	
	if endpoint == "" || model == "" {
		t.Skip("OLLAMA_ENDPOINT and OLLAMA_MODEL must be set for integration test")
	}
	
	t.Logf("Running integration test with endpoint: %s, model: %s", endpoint, model)
	
	provider, err := NewOllamaProvider(Config{
		Provider:    "ollama",
		Endpoint:    endpoint,
		Model:       model,
		Temperature: 0.3,
	})
	
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	
	// Test various prompts
	testCases := []struct {
		name   string
		prompt string
		check  func(string) bool
	}{
		{
			name:   "simple math",
			prompt: "What is 10 + 15? Reply with just the number.",
			check: func(response string) bool {
				return strings.Contains(response, "25")
			},
		},
		{
			name:   "code analysis",
			prompt: "What language is this code: `print('Hello, World!')`? Reply with just the language name.",
			check: func(response string) bool {
				return strings.Contains(strings.ToLower(response), "python")
			},
		},
		{
			name:   "git diff analysis",
			prompt: `Analyze this git diff and provide a one-line summary:
diff --git a/test.js b/test.js
index 123..456 100644
--- a/test.js
+++ b/test.js
@@ -1,3 +1,3 @@
 function hello() {
-  console.log("Hello");
+  console.log("Hello, World!");
 }`,
			check: func(response string) bool {
				return len(response) > 10 // Should have some analysis
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			
			start := time.Now()
			result, err := provider.Analyze(ctx, tc.prompt)
			duration := time.Since(start)
			
			if err != nil {
				t.Errorf("Analysis failed: %v", err)
				return
			}
			
			t.Logf("Response (in %v): %s", duration, result)
			
			if !tc.check(result) {
				t.Errorf("Response validation failed for prompt: %s", tc.prompt)
			}
		})
	}
}