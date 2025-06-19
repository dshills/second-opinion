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
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.IsError {
				t.Fatalf("Handler returned error: %v", result.Content)
			}

			// Check response content
			response := ""
			if len(result.Content) > 0 {
				if textContent, ok := result.Content[0].(mcp.TextContent); ok {
					response = textContent.Text
				}
			}

			if response == "" {
				t.Error("Empty response")
			}

			// Check for expected content
			for _, check := range tc.checkFor {
				if !strings.Contains(strings.ToLower(response), strings.ToLower(check)) {
					t.Errorf("Response doesn't contain '%s': %s", check, response)
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
			checkFor: []string{"security", "division"},
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
			checkFor: []string{"SQL", "injection"},
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
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check response
			response := getTextResponse(result)
			if response == "" {
				t.Error("Empty response")
			}

			for _, check := range tc.checkFor {
				if !strings.Contains(strings.ToLower(response), strings.ToLower(check)) {
					t.Errorf("Response doesn't contain '%s'", check)
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
				t.Fatalf("Error with provider %s: %v", provider, err)
			}

			response := getTextResponse(result)
			if response == "" {
				t.Errorf("Empty response from %s", provider)
			}

			t.Logf("%s response length: %d", provider, len(response))
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
		{"google", "gemini-1.0-pro"},
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
				t.Fatalf("Error with %s/%s: %v", tc.provider, tc.model, err)
			}

			response := getTextResponse(result)
			if response == "" {
				t.Errorf("Empty response from %s/%s", tc.provider, tc.model)
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
