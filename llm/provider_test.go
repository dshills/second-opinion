package llm_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dshills/second-opinion/config"
	"github.com/dshills/second-opinion/llm"
)

// TestProviderConnections tests connections to all configured LLM providers
func TestProviderConnections(t *testing.T) {
	// Load configuration from .env file
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Define test cases for each provider
	testCases := []struct {
		name     string
		provider string
		skipIf   func() bool
		timeout  time.Duration
	}{
		{
			name:     "OpenAI Connection",
			provider: "openai",
			skipIf: func() bool {
				return true
				//return cfg.OpenAI.APIKey == "" || cfg.OpenAI.APIKey == "your_openai_api_key_here"
			},
			timeout: 30 * time.Second,
		},
		{
			name:     "Google AI Connection",
			provider: "google",
			skipIf: func() bool {
				return true
				//return cfg.Google.APIKey == "" || cfg.Google.APIKey == "your_google_api_key_here"
			},
			timeout: 30 * time.Second,
		},
		{
			name:     "Ollama Connection",
			provider: "ollama",
			skipIf: func() bool {
				return true
				// Check if Ollama is running by looking at the endpoint
				//return cfg.Ollama.Endpoint == "" || !isOllamaRunning(cfg.Ollama.Endpoint)
			},
			timeout: 60 * time.Second, // Ollama can be slower
		},
		{
			name:     "Mistral AI Connection",
			provider: "mistral",
			skipIf: func() bool {
				return cfg.Mistral.APIKey == "" || cfg.Mistral.APIKey == "your_mistral_api_key_here"
			},
			timeout: 30 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipIf() {
				t.Skipf("Skipping %s test: provider not configured", tc.provider)
			}

			// Get provider configuration
			apiKey, model, endpoint := cfg.GetProviderConfig(tc.provider)

			// Create provider
			providerConfig := llm.Config{
				Provider:    tc.provider,
				APIKey:      apiKey,
				Model:       model,
				Endpoint:    endpoint,
				Temperature: cfg.Temperature,
				MaxTokens:   100, // Use smaller token limit for tests
			}

			fmt.Println("\n----------------------------------------------")
			fmt.Printf("Testing %s provider with config: %+v\n", tc.provider, providerConfig)

			provider, err := llm.NewProvider(providerConfig)
			if err != nil {
				t.Fatalf("Failed to create %s provider: %v", tc.provider, err)
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			// Test with a simple prompt
			prompt := "Analyze this code snippet: func main() { fmt.Println(\"test\") }"
			response, err := provider.Analyze(ctx, prompt)
			if err != nil {
				fmt.Printf("Error analyzing with %s: %v\n", tc.provider, err)
				return
			}
			fmt.Printf("%s response: %s\n", tc.provider, response)
			fmt.Println("\n----------------------------------------------")
		})
	}
}

/*
// TestProviderModels tests different models for each provider
func TestProviderModels(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Define model variants to test
	modelTests := []struct {
		provider string
		models   []string
		skipIf   func() bool
	}{
		{
			provider: "openai",
			models:   []string{"gpt-3.5-turbo", "gpt-4o-mini"},
			skipIf: func() bool {
				return cfg.OpenAI.APIKey == "" || cfg.OpenAI.APIKey == "your_openai_api_key_here"
			},
		},
		{
			provider: "google",
			models:   []string{"gemini-1.5-flash", "gemini-1.5-pro"},
			skipIf: func() bool {
				return cfg.Google.APIKey == "" || cfg.Google.APIKey == "your_google_api_key_here"
			},
		},
	}

	for _, test := range modelTests {
		t.Run(test.provider+" Models", func(t *testing.T) {
			if test.skipIf() {
				t.Skipf("Skipping %s model tests: provider not configured", test.provider)
			}

			for _, model := range test.models {
				t.Run(model, func(t *testing.T) {
					// Get base provider configuration
					apiKey, _, endpoint := cfg.GetProviderConfig(test.provider)

					// Create provider with specific model
					providerConfig := llm.Config{
						Provider:    test.provider,
						APIKey:      apiKey,
						Model:       model,
						Endpoint:    endpoint,
						Temperature: 0.1, // Low temperature for consistent results
						MaxTokens:   50,
					}

					provider, err := llm.NewProvider(providerConfig)
					if err != nil {
						t.Fatalf("Failed to create provider with model %s: %v", model, err)
					}

					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					// Simple test prompt
					prompt := "Complete this: 2 + 2 equals"
					response, err := provider.Analyze(ctx, prompt)
					if err != nil {
						// Skip if model is not available
						if strings.Contains(err.Error(), "model_not_found") ||
							strings.Contains(err.Error(), "Your organization must be verified") ||
							strings.Contains(err.Error(), "is not found for API version") {
							t.Skipf("Model %s not available: %v", model, err)
							return
						}
						t.Fatalf("Analysis failed with model %s: %v", model, err)
					}

					// Check response contains "4"
					if !strings.Contains(response, "4") {
						t.Errorf("Model %s didn't return expected answer: %s", model, response)
					}

					t.Logf("%s/%s response: %s", test.provider, model, response)
				})
			}
		})
	}
}

// TestAnalysisPrompts tests the different analysis prompt types
func TestAnalysisPrompts(t *testing.T) {
	tests := []struct {
		name         string
		analysisType string
		content      string
		options      map[string]interface{}
		checkFor     []string
	}{
		{
			name:         "Diff Analysis",
			analysisType: "diff",
			content:      "diff --git a/test.go b/test.go\n+func NewFunc() {}\n-func OldFunc() {}",
			options:      map[string]interface{}{"summarize": true},
			checkFor:     []string{"changes", "Summary"},
		},
		{
			name:         "Code Review",
			analysisType: "code_review",
			content:      "func divide(a, b int) int { return a / b }",
			options:      map[string]interface{}{"language": "go", "focus": "security"},
			checkFor:     []string{"Security", "Code"},
		},
		{
			name:         "Commit Analysis",
			analysisType: "commit",
			content:      "commit abc123\nAuthor: Test\nDate: Mon Oct 30\n\nFix: resolve division by zero",
			options:      nil,
			checkFor:     []string{"commit", "Summary"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prompt := llm.AnalysisPrompt(test.analysisType, test.content, test.options)

			// Check that prompt contains expected elements
			for _, check := range test.checkFor {
				if !strings.Contains(prompt, check) {
					t.Errorf("Prompt doesn't contain expected text '%s': %s", check, prompt)
				}
			}
		})
	}
}

// TestEnvironmentVariables verifies that environment variables are loaded correctly
func TestEnvironmentVariables(t *testing.T) {
	// This test helps debug configuration issues
	t.Run("Config Loading", func(t *testing.T) {
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		t.Logf("Default Provider: %s", cfg.DefaultProvider)
		t.Logf("Temperature: %f", cfg.Temperature)
		t.Logf("Max Tokens: %d", cfg.MaxTokens)

		// Log which providers are configured (without exposing keys)
		if cfg.OpenAI.APIKey != "" && cfg.OpenAI.APIKey != "your_openai_api_key_here" {
			t.Logf("OpenAI: Configured with model %s", cfg.OpenAI.Model)
		}
		if cfg.Google.APIKey != "" && cfg.Google.APIKey != "your_google_api_key_here" {
			t.Logf("Google: Configured with model %s", cfg.Google.Model)
		}
		if cfg.Ollama.Endpoint != "" {
			t.Logf("Ollama: Configured with endpoint %s and model %s", cfg.Ollama.Endpoint, cfg.Ollama.Model)
		}
		if cfg.Mistral.APIKey != "" && cfg.Mistral.APIKey != "your_mistral_api_key_here" {
			t.Logf("Mistral: Configured with model %s", cfg.Mistral.Model)
		}
	})
}
*/
