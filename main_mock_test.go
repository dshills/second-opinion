package main

import (
	"context"
	"strings"
	"testing"

	"github.com/dshills/second-opinion/config"
	"github.com/dshills/second-opinion/llm"
	"github.com/mark3labs/mcp-go/mcp"
)

// MockProvider for testing
type MockProvider struct {
	name     string
	response string
	err      error
}

func (m *MockProvider) Analyze(ctx context.Context, prompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.response != "" {
		return m.response, nil
	}
	// Return a simple mock response
	return "Mock analysis complete. The code appears to be correct.", nil
}

func (m *MockProvider) Name() string {
	return m.name
}

// TestHandlersWithMock tests handlers using mock provider
func TestHandlersWithMock(t *testing.T) {
	// Save original state
	originalProviders := llmProviders
	originalCfg := cfg

	// Setup mock environment
	llmProviders = make(map[string]llm.Provider)
	cfg = &config.Config{
		DefaultProvider: "mock",
		Temperature:     0.3,
		MaxTokens:       4096,
	}

	// Add mock provider
	mockProvider := &MockProvider{
		name:     "mock",
		response: "Mock analysis: Code review complete. Found potential issue with division by zero.",
	}
	llmProviders["mock"] = mockProvider

	// Restore original state after test
	defer func() {
		llmProviders = originalProviders
		cfg = originalCfg
	}()

	t.Run("TestHandleGitDiff", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "analyze_git_diff",
				Arguments: map[string]any{
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
			},
		}

		result, err := handleGitDiff(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Fatalf("Handler returned error: %v", result.Content)
		}

		response := getTextResponseMock(result)
		if response == "" {
			t.Error("Empty response")
		}

		if !strings.Contains(response, "Mock analysis") {
			t.Errorf("Response doesn't contain mock prefix: %s", response)
		}
	})

	t.Run("TestHandleCodeReview", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "review_code",
				Arguments: map[string]any{
					"code": `func divide(a, b int) int {
	return a / b
}`,
					"language": "go",
					"focus":    "security",
				},
			},
		}

		result, err := handleCodeReview(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		response := getTextResponseMock(result)
		if response == "" {
			t.Error("Empty response")
		}

		if !strings.Contains(response, "division") {
			t.Errorf("Response doesn't mention division issue: %s", response)
		}
	})

	t.Run("TestHandleCommitAnalysis", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "analyze_commit",
				Arguments: map[string]any{
					"repo_path": ".",
				},
			},
		}

		// This should work with git commands even with mock provider
		result, err := handleCommitAnalysis(context.Background(), req)
		if err != nil {
			// It's ok if this fails due to git not being available in test env
			t.Logf("Git operation failed (expected in some test environments): %v", err)
			return
		}

		if result.IsError {
			// Git errors are acceptable in test environment
			t.Logf("Git command error (expected in some test environments): %v", result.Content)
			return
		}

		response := getTextResponseMock(result)
		if response == "" {
			t.Error("Empty response")
		}
	})
}

// TestErrorCases tests error handling
func TestErrorCases(t *testing.T) {
	// Save original state
	originalProviders := llmProviders
	originalCfg := cfg

	// Setup mock environment
	llmProviders = make(map[string]llm.Provider)
	cfg = &config.Config{
		DefaultProvider: "mock",
	}

	// Restore original state after test
	defer func() {
		llmProviders = originalProviders
		cfg = originalCfg
	}()

	t.Run("MissingProvider", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "analyze_git_diff",
				Arguments: map[string]any{
					"diff_content": "test diff",
					"provider":     "nonexistent",
				},
			},
		}

		result, err := handleGitDiff(context.Background(), req)
		if err == nil && !result.IsError {
			t.Error("Expected error for missing provider")
		}
	})

	t.Run("MissingDiffContent", func(t *testing.T) {
		// Add mock provider
		llmProviders["mock"] = &MockProvider{name: "mock"}

		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "analyze_git_diff",
				Arguments: map[string]any{},
			},
		}

		result, err := handleGitDiff(context.Background(), req)
		if err == nil && !result.IsError {
			t.Error("Expected error for missing diff content")
		}
	})

	t.Run("MissingCode", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "review_code",
				Arguments: map[string]any{},
			},
		}

		result, err := handleCodeReview(context.Background(), req)
		if err == nil && !result.IsError {
			t.Error("Expected error for missing code")
		}
	})
}

// Helper to get text response from result
func getTextResponseMock(result *mcp.CallToolResult) string {
	if result == nil || result.Content == nil || len(result.Content) == 0 {
		return ""
	}

	if textContent, ok := result.Content[0].(mcp.TextContent); ok {
		return textContent.Text
	}

	return ""
}
