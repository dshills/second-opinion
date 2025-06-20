package main

import (
	"context"
	"strings"
	"testing"

	"github.com/dshills/second-opinion/config"
	"github.com/dshills/second-opinion/llm"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Load configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// Force default provider to ollama for tests to avoid API rate limits
	cfg.DefaultProvider = "ollama"

	// Initialize default provider for tests
	apiKey, model, endpoint := cfg.GetProviderConfig(cfg.DefaultProvider)
	defaultConfig := llm.Config{
		Provider:    cfg.DefaultProvider,
		APIKey:      apiKey,
		Model:       model,
		Endpoint:    endpoint,
		Temperature: cfg.Temperature,
		MaxTokens:   cfg.MaxTokens,
	}

	defaultProvider, err := llm.NewProvider(defaultConfig)
	if err == nil {
		llmProviders[cfg.DefaultProvider] = defaultProvider
	}

	// Run tests
	m.Run()
}

// TestHandleGitDiff tests the git diff analysis handler
func TestHandleGitDiff(t *testing.T) {
	if !isProviderConfigured(cfg.DefaultProvider) {
		t.Skip("Default provider not configured")
	}

	// Get provider model for error reporting
	_, model, _ := cfg.GetProviderConfig(cfg.DefaultProvider)

	testCases := []struct {
		name     string
		args     map[string]any
		wantErr  bool
		checkFor []string
	}{
		{
			name: "Basic diff analysis",
			args: map[string]any{
				"diff_content": `diff --git a/main.go b/main.go
index abc123..def456 100644
--- a/main.go
+++ b/main.go
@@ -10,6 +10,7 @@ import (

 func main() {
+	// New comment
 	fmt.Println("Hello, World!")
 }`,
			},
			wantErr:  false,
			checkFor: []string{"main.go", "comment"},
		},
		{
			name: "Diff with summary",
			args: map[string]any{
				"diff_content": `diff --git a/test.go b/test.go
+func NewFeature() {}`,
				"summarize": true,
			},
			wantErr:  false,
			checkFor: []string{"summary", "feature"},
		},
		{
			name:    "Missing diff content",
			args:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "analyze_git_diff",
					Arguments: tc.args,
				},
			}

			result, err := handleGitDiff(context.Background(), req)

			if tc.wantErr {
				if err == nil && !result.IsError {
					t.Errorf("[Provider: %s, Model: %s] Expected error but got none", cfg.DefaultProvider, model)
				}
				return
			}

			if err != nil {
				t.Fatalf("[Provider: %s, Model: %s] Unexpected error: %v", cfg.DefaultProvider, model, err)
			}

			if result.IsError {
				t.Fatalf("[Provider: %s, Model: %s] Handler returned error: %v", cfg.DefaultProvider, model, result.Content)
			}

			// Check response content
			response := ""
			if len(result.Content) > 0 {
				if textContent, ok := result.Content[0].(mcp.TextContent); ok {
					response = textContent.Text
				}
			}

			if response == "" {
				t.Errorf("[Provider: %s, Model: %s] Empty response", cfg.DefaultProvider, model)
			}

			// Check for expected content
			for _, check := range tc.checkFor {
				if !strings.Contains(strings.ToLower(response), strings.ToLower(check)) {
					t.Errorf("[Provider: %s, Model: %s] Response doesn't contain '%s': %s", cfg.DefaultProvider, model, check, response)
				}
			}
		})
	}
}

// TestHandleCodeReview tests the code review handler
func TestHandleCodeReview(t *testing.T) {
	if !isProviderConfigured(cfg.DefaultProvider) {
		t.Skip("Default provider not configured")
	}

	// Get provider model for error reporting
	_, model, _ := cfg.GetProviderConfig(cfg.DefaultProvider)

	testCases := []struct {
		name     string
		args     map[string]any
		wantErr  bool
		checkFor []string
	}{
		{
			name: "Go code security review",
			args: map[string]any{
				"code": `func divide(a, b int) int {
	return a / b
}`,
				"language": "go",
				"focus":    "security",
			},
			wantErr:  false,
			checkFor: []string{"divide", "zero"},
		},
		{
			name: "Python code all aspects",
			args: map[string]any{
				"code": `def process_user_input(user_data):
    query = "SELECT * FROM users WHERE name = '" + user_data + "'"
    return execute_query(query)`,
				"language": "python",
				"focus":    "all",
			},
			wantErr:  false,
			checkFor: []string{"query", "user_data"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "review_code",
					Arguments: tc.args,
				},
			}

			result, err := handleCodeReview(context.Background(), req)

			if tc.wantErr {
				if err == nil && !result.IsError {
					t.Errorf("[Provider: %s, Model: %s] Expected error but got none", cfg.DefaultProvider, model)
				}
				return
			}

			if err != nil {
				t.Fatalf("[Provider: %s, Model: %s] Unexpected error: %v", cfg.DefaultProvider, model, err)
			}

			// Check response
			response := getTextResponse(result)
			if response == "" {
				t.Errorf("[Provider: %s, Model: %s] Empty response", cfg.DefaultProvider, model)
			}

			for _, check := range tc.checkFor {
				if !strings.Contains(strings.ToLower(response), strings.ToLower(check)) {
					t.Errorf("[Provider: %s, Model: %s] Response doesn't contain '%s'", cfg.DefaultProvider, model, check)
				}
			}
		})
	}
}

// TestProviderOverride tests using different providers in handlers
func TestProviderOverride(t *testing.T) {
	// Test with each configured provider
	providers := []string{"openai", "google", "mistral"}

	for _, provider := range providers {
		t.Run("Override with "+provider, func(t *testing.T) {
			if !isProviderConfigured(provider) {
				t.Skipf("%s not configured", provider)
			}

			// Get provider model for error reporting
			_, model, _ := cfg.GetProviderConfig(provider)

			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "analyze_git_diff",
					Arguments: map[string]any{
						"diff_content": "diff --git a/test.go b/test.go\n+// Test",
						"provider":     provider,
					},
				},
			}

			result, err := handleGitDiff(context.Background(), req)
			if err != nil {
				t.Fatalf("[Provider: %s, Model: %s] Error: %v", provider, model, err)
			}

			response := getTextResponse(result)
			if response == "" {
				t.Errorf("[Provider: %s, Model: %s] Empty response", provider, model)
			}

			t.Logf("[Provider: %s, Model: %s] Response length: %d", provider, model, len(response))
		})
	}
}

// TestModelOverride tests using different models
func TestModelOverride(t *testing.T) {
	testCases := []struct {
		provider string
		model    string
	}{
		{"openai", "gpt-3.5-turbo"},
		{"openai", "gpt-4o-mini"},
		{"google", "gemini-2.0-flash-exp"},
	}

	for _, tc := range testCases {
		t.Run(tc.provider+"/"+tc.model, func(t *testing.T) {
			if !isProviderConfigured(tc.provider) {
				t.Skipf("%s not configured", tc.provider)
			}

			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "review_code",
					Arguments: map[string]any{
						"code":     "func main() { fmt.Println(1/0) }",
						"language": "go",
						"provider": tc.provider,
						"model":    tc.model,
					},
				},
			}

			result, err := handleCodeReview(context.Background(), req)
			if err != nil {
				t.Fatalf("[Provider: %s, Model: %s] Error: %v", tc.provider, tc.model, err)
			}

			response := getTextResponse(result)
			if response == "" {
				t.Errorf("[Provider: %s, Model: %s] Empty response", tc.provider, tc.model)
			}
		})
	}
}

// Helper functions

func isProviderConfigured(provider string) bool {
	apiKey, _, _ := cfg.GetProviderConfig(provider)
	switch provider {
	case "openai":
		return apiKey != "" && apiKey != "your_openai_api_key_here"
	case "google":
		return apiKey != "" && apiKey != "your_google_api_key_here"
	case "mistral":
		return apiKey != "" && apiKey != "your_mistral_api_key_here"
	case "ollama":
		// For Ollama, just check if it's the provider
		return true
	default:
		return false
	}
}

func getTextResponse(result *mcp.CallToolResult) string {
	if result == nil || result.Content == nil || len(result.Content) == 0 {
		return ""
	}

	if textContent, ok := result.Content[0].(mcp.TextContent); ok {
		return textContent.Text
	}

	return ""
}
