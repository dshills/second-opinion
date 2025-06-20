package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// TestOllamaDiagnostic provides detailed diagnostics for Ollama issues
func TestOllamaDiagnostic(t *testing.T) {
	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	model := os.Getenv("OLLAMA_MODEL")

	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2"
	}

	t.Logf("=== Ollama Diagnostic Test ===")
	t.Logf("Endpoint: %s", endpoint)
	t.Logf("Model: %s", model)
	t.Logf("==============================")

	// Test 1: DNS Resolution
	t.Run("DNS Resolution", func(t *testing.T) {
		if strings.Contains(endpoint, "localhost") || strings.Contains(endpoint, "127.0.0.1") {
			t.Skip("Skipping DNS test for localhost")
		}

		// Extract hostname from endpoint
		hostname := strings.TrimPrefix(endpoint, "http://")
		hostname = strings.TrimPrefix(hostname, "https://")
		if idx := strings.Index(hostname, ":"); idx > 0 {
			hostname = hostname[:idx]
		}

		t.Logf("Testing DNS resolution for: %s", hostname)

		// Simple connectivity test using HTTP client
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		resp, err := client.Get(endpoint)
		if err != nil {
			t.Errorf("DNS/Network error: %v", err)
			return
		}
		resp.Body.Close()

		t.Logf("✓ Successfully resolved and connected to %s", hostname)
	})

	// Test 2: Model Loading Time
	t.Run("Model Loading Performance", func(t *testing.T) {
		client := &http.Client{
			Timeout: 120 * time.Second, // Allow 2 minutes for model loading
		}

		// Small prompt to test response time
		reqBody := map[string]interface{}{
			"model":  model,
			"prompt": "Hi",
			"stream": false,
			"options": map[string]interface{}{
				"temperature": 0.1,
				"num_predict": 5, // Limit response length
			},
		}

		jsonBody, _ := json.Marshal(reqBody)

		// First request (may need to load model)
		t.Log("Testing first request (may load model)...")
		start := time.Now()

		resp, err := client.Post(
			endpoint+"/api/generate",
			"application/json",
			strings.NewReader(string(jsonBody)),
		)

		firstDuration := time.Since(start)

		if err != nil {
			t.Errorf("First request failed: %v", err)
			return
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		t.Logf("First request took: %v", firstDuration)

		if firstDuration > 30*time.Second {
			t.Logf("⚠️  First request was slow - model may have been loading")
		}

		// Second request (model should be loaded)
		t.Log("Testing second request (model should be loaded)...")
		start = time.Now()

		resp2, err := client.Post(
			endpoint+"/api/generate",
			"application/json",
			strings.NewReader(string(jsonBody)),
		)

		secondDuration := time.Since(start)

		if err != nil {
			t.Errorf("Second request failed: %v", err)
			return
		}
		resp2.Body.Close()

		t.Logf("Second request took: %v", secondDuration)

		if secondDuration > 5*time.Second {
			t.Logf("⚠️  Second request still slow - possible performance issue")
		} else {
			t.Logf("✓ Model response time is good")
		}
	})

	// Test 3: Streaming vs Non-streaming
	t.Run("Streaming Mode Comparison", func(t *testing.T) {
		prompt := "Count from 1 to 5"

		// Test non-streaming
		t.Log("Testing non-streaming mode...")
		reqBody := map[string]interface{}{
			"model":  model,
			"prompt": prompt,
			"stream": false,
		}

		jsonBody, _ := json.Marshal(reqBody)
		client := &http.Client{Timeout: 30 * time.Second}

		start := time.Now()
		resp, err := client.Post(
			endpoint+"/api/generate",
			"application/json",
			strings.NewReader(string(jsonBody)),
		)
		nonStreamDuration := time.Since(start)

		if err != nil {
			t.Errorf("Non-streaming request failed: %v", err)
		} else {
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			resp.Body.Close()

			t.Logf("Non-streaming took: %v", nonStreamDuration)
			if response, ok := result["response"].(string); ok {
				t.Logf("Response length: %d characters", len(response))
			}
		}

		// Note: We're not testing streaming mode here as it requires different handling
		t.Log("Note: Streaming mode test skipped (requires different client handling)")
	})

	// Test 4: Error Recovery
	t.Run("Error Recovery", func(t *testing.T) {
		// Test with invalid model
		t.Log("Testing error handling with invalid model...")

		reqBody := map[string]interface{}{
			"model":  "definitely-not-a-real-model",
			"prompt": "test",
			"stream": false,
		}

		jsonBody, _ := json.Marshal(reqBody)
		client := &http.Client{Timeout: 10 * time.Second}

		resp, err := client.Post(
			endpoint+"/api/generate",
			"application/json",
			strings.NewReader(string(jsonBody)),
		)

		if err != nil {
			t.Logf("Network error with invalid model: %v", err)
		} else {
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			resp.Body.Close()

			if errMsg, ok := result["error"].(string); ok {
				t.Logf("✓ Got expected error: %s", errMsg)
			} else {
				t.Logf("Response: %+v", result)
			}
		}
	})

	// Test 5: Memory/Context Limits
	t.Run("Context Size Handling", func(t *testing.T) {
		// Create a large prompt
		largePrompt := "Please summarize the following text:\n" +
			strings.Repeat("This is a test sentence that repeats. ", 100)

		t.Logf("Testing with large prompt (%d characters)", len(largePrompt))

		reqBody := map[string]interface{}{
			"model":  model,
			"prompt": largePrompt,
			"stream": false,
			"options": map[string]interface{}{
				"num_predict": 50, // Limit response
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		client := &http.Client{Timeout: 60 * time.Second}

		start := time.Now()
		resp, err := client.Post(
			endpoint+"/api/generate",
			"application/json",
			strings.NewReader(string(jsonBody)),
		)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Large prompt request failed: %v", err)
		} else {
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			resp.Body.Close()

			t.Logf("Large prompt took: %v", duration)

			if response, ok := result["response"].(string); ok {
				t.Logf("✓ Successfully processed large prompt, response: %d chars", len(response))
			}
		}
	})
}

// TestOllamaProviderRealWorld tests the provider in real-world scenarios
func TestOllamaProviderRealWorld(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real-world test in short mode")
	}

	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	model := os.Getenv("OLLAMA_MODEL")

	if endpoint == "" || model == "" {
		t.Skip("OLLAMA_ENDPOINT and OLLAMA_MODEL required for real-world test")
	}

	provider, err := NewOllamaProvider(Config{
		Provider:    "ollama",
		Endpoint:    endpoint,
		Model:       model,
		Temperature: 0.3,
	})

	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	testCases := []struct {
		name        string
		prompt      string
		timeout     time.Duration
		minResponse int
		maxResponse int
	}{
		{
			name:        "Very short prompt",
			prompt:      "Hi",
			timeout:     10 * time.Second,
			minResponse: 1,
			maxResponse: 100,
		},
		{
			name: "Code review prompt",
			prompt: `Review this code for issues:
function add(a, b) {
  return a + b
}`,
			timeout:     30 * time.Second,
			minResponse: 10,
			maxResponse: 1000,
		},
		{
			name: "Git diff analysis",
			prompt: `Analyze this git diff:
diff --git a/main.go b/main.go
index 123..456 100644
--- a/main.go
+++ b/main.go
@@ -10,7 +10,7 @@ import (
 
 func main() {
-    fmt.Println("Hello")
+    fmt.Println("Hello, World!")
 }`,
			timeout:     30 * time.Second,
			minResponse: 20,
			maxResponse: 500,
		},
		{
			name:        "Complex reasoning",
			prompt:      "Explain the advantages and disadvantages of microservices architecture in 2 sentences.",
			timeout:     45 * time.Second,
			minResponse: 50,
			maxResponse: 500,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			start := time.Now()
			result, err := provider.Analyze(ctx, tc.prompt)
			duration := time.Since(start)

			if err != nil {
				t.Errorf("Analysis failed: %v", err)
				return
			}

			responseLen := len(result)
			t.Logf("Response time: %v, Length: %d chars", duration, responseLen)

			if responseLen < tc.minResponse {
				t.Errorf("Response too short: expected at least %d chars, got %d",
					tc.minResponse, responseLen)
			}

			if responseLen > tc.maxResponse {
				t.Logf("Warning: Response longer than expected: %d > %d chars",
					responseLen, tc.maxResponse)
			}

			// Log first 200 chars of response
			preview := result
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			t.Logf("Response preview: %s", preview)
		})
	}
}

// TestOllamaStressTest performs stress testing
func TestOllamaStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	model := os.Getenv("OLLAMA_MODEL")

	if endpoint == "" || model == "" {
		t.Skip("OLLAMA_ENDPOINT and OLLAMA_MODEL required for stress test")
	}

	// Only run stress test if explicitly requested
	if os.Getenv("RUN_STRESS_TEST") != "true" {
		t.Skip("Set RUN_STRESS_TEST=true to run stress test")
	}

	provider, err := NewOllamaProvider(Config{
		Provider: "ollama",
		Endpoint: endpoint,
		Model:    model,
	})

	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Test rapid sequential requests
	t.Run("Sequential Requests", func(t *testing.T) {
		prompts := []string{
			"What is 1+1?",
			"What is 2+2?",
			"What is 3+3?",
			"What is 4+4?",
			"What is 5+5?",
		}

		for i, prompt := range prompts {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			start := time.Now()
			result, err := provider.Analyze(ctx, prompt)
			duration := time.Since(start)

			cancel()

			if err != nil {
				t.Errorf("Request %d failed: %v", i+1, err)
			} else {
				t.Logf("Request %d completed in %v: %s", i+1, duration, result)
			}

			// Small delay between requests
			time.Sleep(100 * time.Millisecond)
		}
	})

	// Test context cancellation
	t.Run("Context Cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Use a prompt that would normally take longer
		longPrompt := "Write a detailed essay about quantum computing, covering its history, current applications, and future potential."

		_, err := provider.Analyze(ctx, longPrompt)

		if err == nil {
			t.Error("Expected timeout error, got success")
		} else if !strings.Contains(err.Error(), "context") {
			t.Errorf("Expected context error, got: %v", err)
		} else {
			t.Logf("✓ Context cancellation worked correctly: %v", err)
		}
	})
}
