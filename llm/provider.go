package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/dshills/second-opinion/config"
)

// Provider represents an LLM provider interface
type Provider interface {
	// Analyze sends a prompt to the LLM and returns the response
	Analyze(ctx context.Context, prompt string) (string, error)
	// Name returns the provider name
	Name() string
}

// OptimizedProvider extends Provider with optimization capabilities
type OptimizedProvider interface {
	Provider
	// AnalyzeOptimized performs optimized analysis based on content size and task type
	AnalyzeOptimized(ctx context.Context, prompt string, contentSize int, task config.AnalysisTask) (string, error)
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

	case "uncommitted_work":
		stagedOnly := false
		if s, ok := options["staged_only"].(bool); ok {
			stagedOnly = s
		}

		changeType := "uncommitted changes"
		if stagedOnly {
			changeType = "staged changes"
		}

		prompt := fmt.Sprintf(`Analyze these %s in the repository:

%s

Provide:
1. Summary of all changes (files modified, added, deleted)
2. Type and nature of changes (feature, bugfix, refactor, etc.)
3. Completeness and readiness for commit
4. Potential issues or concerns
5. Suggested commit message(s) if changes are ready
6. Recommendations for organizing commits if changes should be split`, changeType, content)
		return prompt

	default:
		return content
	}
}

// OptimizedAnalysisRequest holds optimization parameters for analysis
type OptimizedAnalysisRequest struct {
	Provider    string
	Content     string
	ContentSize int
	Task        config.AnalysisTask
	Config      *config.Config
}

// NewOptimizedProvider creates an optimized provider wrapper
func NewOptimizedProvider(baseProvider Provider, cfg *config.Config) OptimizedProvider {
	return &optimizedProviderWrapper{
		Provider: baseProvider,
		config:   cfg,
	}
}

// optimizedProviderWrapper wraps any provider with optimization capabilities
type optimizedProviderWrapper struct {
	Provider
	config *config.Config
}

// AnalyzeOptimized performs optimized analysis
func (w *optimizedProviderWrapper) AnalyzeOptimized(ctx context.Context, prompt string, contentSize int, task config.AnalysisTask) (string, error) {
	// Get optimized configuration
	maxTokens, temperature, providerConfig := w.config.GetProviderOptimizedConfig(w.Name(), contentSize, task)

	// Check if we need to chunk the content
	fileCount := estimateFileCount(prompt)
	shouldChunk, chunkSize := w.config.ShouldChunkDiff(contentSize, fileCount)

	if shouldChunk {
		return w.analyzeInChunks(ctx, prompt, chunkSize, maxTokens, temperature, providerConfig)
	}

	// For small content, use direct analysis with optimization
	return w.analyzeWithOptimization(ctx, prompt, maxTokens, temperature, providerConfig)
}

// analyzeInChunks processes large content in chunks
func (w *optimizedProviderWrapper) analyzeInChunks(ctx context.Context, prompt string, chunkSize int, maxTokens int, temperature float64, providerConfig map[string]any) (string, error) {
	// Split content into logical chunks
	chunks := w.splitContentIntoChunks(prompt, chunkSize)

	results := make([]string, 0, len(chunks))
	for i, chunk := range chunks {
		chunkPrompt := fmt.Sprintf("Analysis part %d of %d:\n\n%s", i+1, len(chunks), chunk)

		result, err := w.analyzeWithOptimization(ctx, chunkPrompt, maxTokens, temperature, providerConfig)
		if err != nil {
			return "", fmt.Errorf("chunk %d analysis failed: %w", i+1, err)
		}

		results = append(results, fmt.Sprintf("## Part %d Analysis\n%s", i+1, result))
	}

	// Combine results with a summary
	combinedResult := strings.Join(results, "\n\n")
	summaryPrompt := fmt.Sprintf(`Provide a comprehensive summary of the following analysis parts:

%s

Please provide:
1. Overall summary of all changes
2. Key issues and concerns across all parts
3. Unified recommendations`, combinedResult)

	summary, err := w.analyzeWithOptimization(ctx, summaryPrompt, maxTokens, temperature, providerConfig)
	if err != nil {
		// If summary fails, return the combined results
		return combinedResult, nil
	}

	return fmt.Sprintf("%s\n\n## Overall Summary\n%s", combinedResult, summary), nil
}

// analyzeWithOptimization performs analysis with optimized parameters
func (w *optimizedProviderWrapper) analyzeWithOptimization(ctx context.Context, prompt string, maxTokens int, temperature float64, providerConfig map[string]any) (string, error) {
	// For now, delegate to the base provider
	// In the future, we could modify the underlying provider's behavior here
	// TODO: Use maxTokens, temperature, and providerConfig to optimize the analysis
	_ = maxTokens      // Reserved for future optimization
	_ = temperature    // Reserved for future optimization
	_ = providerConfig // Reserved for future optimization
	return w.Analyze(ctx, prompt)
}

// splitContentIntoChunks splits content into logical chunks
func (w *optimizedProviderWrapper) splitContentIntoChunks(content string, chunkSizeBytes int) []string {
	// Simple chunking by size for now
	// TODO: Implement smarter chunking by file boundaries, function boundaries, etc.

	if len(content) <= chunkSizeBytes {
		return []string{content}
	}

	var chunks []string
	for i := 0; i < len(content); i += chunkSizeBytes {
		end := i + chunkSizeBytes
		if end > len(content) {
			end = len(content)
		}

		chunk := content[i:end]

		// Try to break at line boundaries to maintain readability
		if end < len(content) {
			if lastNewline := strings.LastIndex(chunk, "\n"); lastNewline > len(chunk)/2 {
				chunk = content[i : i+lastNewline+1]
				i = i + lastNewline + 1 - chunkSizeBytes // Adjust for next iteration
			}
		}

		chunks = append(chunks, chunk)
	}

	return chunks
}

// estimateFileCount estimates the number of files in a diff
func estimateFileCount(content string) int {
	// Count occurrences of "diff --git" or similar patterns
	count := strings.Count(content, "diff --git")
	if count == 0 {
		// Fallback: count file headers
		count = strings.Count(content, "+++")
	}
	if count == 0 {
		// Default to 1 if no clear file indicators
		count = 1
	}
	return count
}

// GetTaskFromAnalysisType maps analysis types to tasks
func GetTaskFromAnalysisType(analysisType string) config.AnalysisTask {
	switch analysisType {
	case "diff":
		return config.TaskDiffAnalysis
	case "code_review":
		return config.TaskCodeReview
	case "commit":
		return config.TaskCommitAnalysis
	case "uncommitted_work":
		return config.TaskCodeReview
	case "security":
		return config.TaskSecurityReview
	case "architecture":
		return config.TaskArchitectureReview
	default:
		return config.TaskGeneral
	}
}
