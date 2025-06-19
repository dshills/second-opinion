package main

import (
	"fmt"
	"log"

	"github.com/dshills/second-opinion/config"
	"github.com/dshills/second-opinion/llm"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	cfg          *config.Config
	llmProviders = make(map[string]llm.Provider)
)

func main() {
	// Load configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("%+v", cfg)
	log.Printf("Loaded configuration from %s", cfg.ConfigType)
	if cfg.OpenAI.APIKey != "" {
		log.Println("OpenAI Enabled")
	}
	if cfg.Google.APIKey != "" {
		log.Println("Google Enabled")
	}
	if cfg.Ollama.Endpoint != "" {
		log.Println("Ollama Enabled")
	}
	if cfg.Mistral.APIKey != "" {
		log.Println("Mistral Enabled")
	}
	log.Printf("Default provider: %s", cfg.DefaultProvider)

	// Initialize default LLM provider
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
	if err != nil {
		log.Fatalf("Failed to initialize default LLM provider: %v", err)
	}
	llmProviders[cfg.DefaultProvider] = defaultProvider

	s := server.NewMCPServer(
		cfg.ServerName,
		cfg.ServerVersion,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// Git diff analysis tool
	gitDiffTool := mcp.NewTool("analyze_git_diff",
		mcp.WithDescription("Analyze git diff output to understand code changes using LLM"),
		mcp.WithString("diff_content",
			mcp.Required(),
			mcp.Description("Git diff output to analyze"),
		),
		mcp.WithBoolean("summarize",
			mcp.Description("Whether to provide a summary of changes"),
		),
		mcp.WithString("provider",
			mcp.Description("LLM provider to use (openai, google, ollama, mistral)"),
		),
		mcp.WithString("model",
			mcp.Description("Model to use (overrides default for provider)"),
		),
	)
	s.AddTool(gitDiffTool, handleGitDiff)

	// Code review tool
	codeReviewTool := mcp.NewTool("review_code",
		mcp.WithDescription("Review code for quality, security, and best practices using LLM"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("Code to review"),
		),
		mcp.WithString("language",
			mcp.Description("Programming language of the code"),
		),
		mcp.WithString("focus",
			mcp.Description("Specific focus area for review (security, performance, style, etc.)"),
			mcp.Enum("security", "performance", "style", "all"),
		),
		mcp.WithString("provider",
			mcp.Description("LLM provider to use (openai, google, ollama, mistral)"),
		),
		mcp.WithString("model",
			mcp.Description("Model to use (overrides default for provider)"),
		),
	)
	s.AddTool(codeReviewTool, handleCodeReview)

	// Commit analysis tool
	commitAnalysisTool := mcp.NewTool("analyze_commit",
		mcp.WithDescription("Analyze a git commit for quality and adherence to best practices using LLM"),
		mcp.WithString("commit_sha",
			mcp.Description("Git commit SHA to analyze (default: HEAD)"),
		),
		mcp.WithString("repo_path",
			mcp.Description("Path to the git repository (default: current directory)"),
		),
		mcp.WithString("provider",
			mcp.Description("LLM provider to use (openai, google, ollama, mistral)"),
		),
		mcp.WithString("model",
			mcp.Description("Model to use (overrides default for provider)"),
		),
	)
	s.AddTool(commitAnalysisTool, handleCommitAnalysis)

	// Get repository info tool
	repoInfoTool := mcp.NewTool("get_repo_info",
		mcp.WithDescription("Get information about a git repository"),
		mcp.WithString("repo_path",
			mcp.Description("Path to the git repository (default: current directory)"),
		),
	)
	s.AddTool(repoInfoTool, handleRepoInfo)

	// Start the stdio server
	log.Printf("Starting %s with default provider: %s", cfg.ServerName, cfg.DefaultProvider)
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// getOrCreateProvider gets an existing provider or creates a new one with the specified config
func getOrCreateProvider(providerName, modelOverride string) (llm.Provider, error) {
	// Use default provider if not specified
	if providerName == "" {
		providerName = cfg.DefaultProvider
	}

	// Create a cache key that includes both provider and model
	cacheKey := providerName
	if modelOverride != "" {
		cacheKey = fmt.Sprintf("%s:%s", providerName, modelOverride)
	}

	// Check if we already have this provider configured
	if provider, exists := llmProviders[cacheKey]; exists {
		return provider, nil
	}

	// Get provider configuration
	apiKey, model, endpoint := cfg.GetProviderConfig(providerName)

	// Use model override if provided
	if modelOverride != "" {
		model = modelOverride
	}

	// Create new provider
	providerConfig := llm.Config{
		Provider:    providerName,
		APIKey:      apiKey,
		Model:       model,
		Endpoint:    endpoint,
		Temperature: cfg.Temperature,
		MaxTokens:   cfg.MaxTokens,
	}

	provider, err := llm.NewProvider(providerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s provider: %w", providerName, err)
	}

	// Cache the provider
	llmProviders[cacheKey] = provider
	return provider, nil
}
