package llm

import (
	"context"
	"fmt"
)

// Provider represents an LLM provider interface
type Provider interface {
	// Analyze sends a prompt to the LLM and returns the response
	Analyze(ctx context.Context, prompt string) (string, error)
	// Name returns the provider name
	Name() string
}

// Config holds configuration for LLM providers
type Config struct {
	Provider    string // openai, google, ollama, mistral
	APIKey      string
	Model       string
	Endpoint    string // For Ollama or custom endpoints
	Temperature float64
	MaxTokens   int
}

// NewProvider creates a new LLM provider based on config
func NewProvider(config Config) (Provider, error) {
	switch config.Provider {
	case "openai":
		return NewOpenAIProvider(config)
	case "google":
		return NewGoogleProvider(config)
	case "ollama":
		return NewOllamaProvider(config)
	case "mistral":
		return NewMistralProvider(config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

// AnalysisPrompt creates a structured prompt for code analysis
func AnalysisPrompt(analysisType, content string, options map[string]any) string {
	switch analysisType {
	case "diff":
		summarize := false
		if s, ok := options["summarize"].(bool); ok {
			summarize = s
		}
		prompt := fmt.Sprintf(`Analyze this git diff and provide:
1. Summary of changes (files changed, lines added/removed)
2. Type of change (feature, bugfix, refactor, etc.)
3. Potential issues or concerns
%s

Git diff:
%s`,
			map[bool]string{true: "4. Brief summary of the overall change", false: ""}[summarize],
			content)
		return prompt

	case "code_review":
		focus := "all"
		if f, ok := options["focus"].(string); ok {
			focus = f
		}
		language := "unknown"
		if l, ok := options["language"].(string); ok {
			language = l
		}

		prompt := fmt.Sprintf(`Review this %s code with focus on %s. Provide:
1. Security issues (if any)
2. Performance concerns (if any)
3. Code quality and style issues
4. Best practice violations
5. Suggestions for improvement

Code:
%s`, language, focus, content)
		return prompt

	case "commit":
		prompt := fmt.Sprintf(`Analyze this git commit information:

%s

Provide:
1. Summary of the commit changes
2. Quality of the commit message
3. Whether the commit follows best practices
4. Suggestions for improvement`, content)
		return prompt

	default:
		return content
	}
}
