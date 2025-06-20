package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// MemoryConfig holds memory management settings
type MemoryConfig struct {
	MaxDiffSizeMB   int  `json:"max_diff_size_mb"`
	MaxFileCount    int  `json:"max_file_count"`
	MaxLineLength   int  `json:"max_line_length"`
	EnableStreaming bool `json:"enable_streaming"`
	ChunkSizeMB     int  `json:"chunk_size_mb"`
}

// Config holds the application configuration.
type Config struct {
	// Default provider settings
	DefaultProvider string  `json:"default_provider"`
	Temperature     float64 `json:"temperature"`
	MaxTokens       int     `json:"max_tokens"`

	// Provider-specific configurations
	OpenAI struct {
		APIKey string `json:"api_key"`
		Model  string `json:"model"`
	} `json:"openai"`
	Google struct {
		APIKey string `json:"api_key"`
		Model  string `json:"model"`
	} `json:"google"`
	Ollama struct {
		Endpoint string `json:"endpoint"`
		Model    string `json:"model"`
	} `json:"ollama"`
	Mistral struct {
		APIKey string `json:"api_key"`
		Model  string `json:"model"`
	} `json:"mistral"`

	// Server settings
	ServerName    string `json:"server_name"`
	ServerVersion string `json:"server_version"`

	// Memory management settings
	Memory MemoryConfig `json:"memory"`

	ConfigType string
}

func Load() (*Config, error) {
	conf, err := loadFromHome()
	if err == nil {
		conf.ConfigType = ".second-opinion.json"
		return conf, nil
	}
	return loadEnv()
}

func loadFromHome() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	fPath := filepath.Join(homeDir, ".second-opinion.json")
	f, err := os.Open(fPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	conf := Config{ConfigType: ".second-opinion.json"}
	err = json.NewDecoder(f).Decode(&conf)

	// Set memory defaults if not specified in JSON
	if conf.Memory.MaxDiffSizeMB == 0 {
		conf.Memory.MaxDiffSizeMB = 10
	}
	if conf.Memory.MaxFileCount == 0 {
		conf.Memory.MaxFileCount = 1000
	}
	if conf.Memory.MaxLineLength == 0 {
		conf.Memory.MaxLineLength = 1000
	}
	if conf.Memory.ChunkSizeMB == 0 {
		conf.Memory.ChunkSizeMB = 1
	}
	// EnableStreaming defaults to true unless explicitly set to false
	if !conf.Memory.EnableStreaming && conf.Memory.MaxDiffSizeMB > 0 {
		conf.Memory.EnableStreaming = true
	}

	return &conf, err
}

// Load loads configuration from environment variables.
func loadEnv() (*Config, error) {
	// Load .env file if it exists
	// Try to load from current directory first, then parent directories
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	cfg := &Config{
		DefaultProvider: getEnv("DEFAULT_PROVIDER", "openai"),
		ServerName:      getEnv("SERVER_NAME", "Second Opinion 🔍"),
		ServerVersion:   getEnv("SERVER_VERSION", "1.0.0"),
	}

	// Load provider-specific configurations
	cfg.OpenAI.APIKey = getEnv("OPENAI_API_KEY", "")
	cfg.OpenAI.Model = getEnv("OPENAI_MODEL", "gpt-4o-mini")

	cfg.Google.APIKey = getEnv("GOOGLE_API_KEY", "")
	cfg.Google.Model = getEnv("GOOGLE_MODEL", "gemini-2.0-flash-exp")

	cfg.Ollama.Endpoint = getEnv("OLLAMA_ENDPOINT", "http://localhost:11434")
	cfg.Ollama.Model = getEnv("OLLAMA_MODEL", "devstral:latest")

	cfg.Mistral.APIKey = getEnv("MISTRAL_API_KEY", "")
	cfg.Mistral.Model = getEnv("MISTRAL_MODEL", "mistral-small-latest")

	// Parse temperature
	if temp := getEnv("LLM_TEMPERATURE", "0.3"); temp != "" {
		if t, err := strconv.ParseFloat(temp, 64); err == nil {
			cfg.Temperature = t
		} else {
			cfg.Temperature = 0.3
		}
	}

	// Parse max tokens
	if tokens := getEnv("LLM_MAX_TOKENS", "4096"); tokens != "" {
		if t, err := strconv.Atoi(tokens); err == nil {
			cfg.MaxTokens = t
		} else {
			cfg.MaxTokens = 4096
		}
	}

	// Set memory defaults
	cfg.Memory.MaxDiffSizeMB = 10
	cfg.Memory.MaxFileCount = 1000
	cfg.Memory.MaxLineLength = 1000
	cfg.Memory.EnableStreaming = true
	cfg.Memory.ChunkSizeMB = 1

	// Override with environment variables if set
	if maxDiff := getEnv("MAX_DIFF_SIZE_MB", ""); maxDiff != "" {
		if v, err := strconv.Atoi(maxDiff); err == nil {
			cfg.Memory.MaxDiffSizeMB = v
		}
	}
	if maxFiles := getEnv("MAX_FILE_COUNT", ""); maxFiles != "" {
		if v, err := strconv.Atoi(maxFiles); err == nil {
			cfg.Memory.MaxFileCount = v
		}
	}
	if maxLine := getEnv("MAX_LINE_LENGTH", ""); maxLine != "" {
		if v, err := strconv.Atoi(maxLine); err == nil {
			cfg.Memory.MaxLineLength = v
		}
	}
	if streaming := getEnv("ENABLE_STREAMING", ""); streaming != "" {
		cfg.Memory.EnableStreaming = streaming == "true" || streaming == "1"
	}
	if chunkSize := getEnv("CHUNK_SIZE_MB", ""); chunkSize != "" {
		if v, err := strconv.Atoi(chunkSize); err == nil {
			cfg.Memory.ChunkSizeMB = v
		}
	}

	return cfg, nil
}

// GetProviderConfig returns the configuration for a specific provider.
func (c *Config) GetProviderConfig(provider string) (apiKey, model, endpoint string) {
	switch provider {
	case "openai":
		return c.OpenAI.APIKey, c.OpenAI.Model, ""
	case "google":
		return c.Google.APIKey, c.Google.Model, ""
	case "ollama":
		return "", c.Ollama.Model, c.Ollama.Endpoint
	case "mistral":
		return c.Mistral.APIKey, c.Mistral.Model, ""
	default:
		// Return config for default provider if different from requested
		if provider != c.DefaultProvider && c.DefaultProvider != "" {
			return c.GetProviderConfig(c.DefaultProvider)
		}
		// Return empty values if provider not found
		return "", "", ""
	}
}

// AnalysisTask defines the type of analysis being performed
type AnalysisTask string

const (
	TaskCodeReview         AnalysisTask = "code_review"
	TaskSecurityReview     AnalysisTask = "security_review"
	TaskCommitAnalysis     AnalysisTask = "commit_analysis"
	TaskArchitectureReview AnalysisTask = "architecture_review"
	TaskDiffAnalysis       AnalysisTask = "diff_analysis"
	TaskGeneral            AnalysisTask = "general"
)

// GetOptimalTokensForDiff returns optimal token count based on diff size
func (c *Config) GetOptimalTokensForDiff(diffSizeBytes int) int {
	// Convert bytes to KB for easier thresholds
	diffSizeKB := diffSizeBytes / 1024

	switch {
	case diffSizeKB < 5: // Very small diffs (< 5KB)
		return 4096
	case diffSizeKB < 20: // Small diffs (5-20KB)
		return 6144
	case diffSizeKB < 50: // Medium diffs (20-50KB)
		return 8192
	case diffSizeKB < 150: // Large diffs (50-150KB)
		return 12288
	case diffSizeKB < 500: // Very large diffs (150-500KB)
		return 16384
	default: // Huge diffs (> 500KB)
		return 32768
	}
}

// GetOptimalTemperatureForTask returns optimal temperature based on analysis task
func (c *Config) GetOptimalTemperatureForTask(task AnalysisTask) float64 {
	switch task {
	case TaskSecurityReview:
		return 0.1 // Very deterministic for security issues
	case TaskCodeReview, TaskCommitAnalysis:
		return 0.2 // Mostly deterministic for code analysis
	case TaskDiffAnalysis:
		return 0.25 // Slightly more flexible for diff interpretation
	case TaskArchitectureReview:
		return 0.3 // Allow creativity for architectural suggestions
	case TaskGeneral:
		return 0.3 // Balanced for general queries
	default:
		return 0.2 // Safe default
	}
}

// GetProviderOptimizedConfig returns provider-specific optimized configuration
func (c *Config) GetProviderOptimizedConfig(provider string, diffSize int, task AnalysisTask) (maxTokens int, temperature float64, providerConfig map[string]any) {
	baseTokens := c.GetOptimalTokensForDiff(diffSize)
	baseTemp := c.GetOptimalTemperatureForTask(task)

	// Provider-specific adjustments
	switch provider {
	case "openai":
		// OpenAI has excellent context handling, can use full allocation
		maxTokens = baseTokens
		temperature = baseTemp
		providerConfig = map[string]any{
			"top_p": 0.9,
		}

	case "google":
		// Gemini has massive context window, optimize for quality
		maxTokens = min(baseTokens, 8192) // Gemini works well with moderate token counts
		temperature = baseTemp
		providerConfig = map[string]any{
			"top_k":           20,  // More focused sampling for code
			"top_p":           0.8, // Conservative nucleus sampling
			"candidate_count": 1,
		}

	case "mistral":
		// Mistral is efficient, use moderate allocation
		maxTokens = min(baseTokens, 8192)
		temperature = baseTemp
		providerConfig = map[string]any{
			"top_p":      0.8, // More conservative
			"max_tokens": maxTokens,
		}

	case "ollama":
		// Ollama depends on local model, be more conservative
		maxTokens = min(baseTokens, 8192)
		temperature = baseTemp
		providerConfig = map[string]any{
			"top_k":          20,   // More focused
			"top_p":          0.8,  // Conservative
			"repeat_penalty": 1.05, // Reduce repetition
			"num_predict":    maxTokens,
		}

	default:
		maxTokens = baseTokens
		temperature = baseTemp
		providerConfig = map[string]any{}
	}

	return maxTokens, temperature, providerConfig
}

// ShouldChunkDiff determines if a diff should be chunked based on size and complexity
func (c *Config) ShouldChunkDiff(diffSizeBytes int, fileCount int) (shouldChunk bool, chunkSizeBytes int) {
	maxSizeBytes := c.Memory.MaxDiffSizeMB * 1024 * 1024

	// Check size threshold
	if diffSizeBytes > maxSizeBytes {
		shouldChunk = true
	}

	// Check file count threshold
	if fileCount > c.Memory.MaxFileCount {
		shouldChunk = true
	}

	// Calculate optimal chunk size
	chunkSizeBytes = c.Memory.ChunkSizeMB * 1024 * 1024

	// Adjust chunk size based on file count
	if fileCount > 100 {
		// Smaller chunks for repos with many files
		chunkSizeBytes = chunkSizeBytes / 2
	}

	return shouldChunk, chunkSizeBytes
}

// EstimateTokensForText estimates token count for text (rough approximation)
func (c *Config) EstimateTokensForText(text string) int {
	// Rough estimation: ~4 characters per token for code
	return len(text) / 4
}

// GetMemoryOptimizedConfig returns memory-aware configuration for large operations
func (c *Config) GetMemoryOptimizedConfig(estimatedInputTokens int) (streaming bool, batchSize int) {
	streaming = c.Memory.EnableStreaming

	// Force streaming for large inputs
	if estimatedInputTokens > 8192 {
		streaming = true
	}

	// Calculate batch size based on available "memory budget"
	maxBudget := 16384 // Conservative token budget per request
	if estimatedInputTokens < maxBudget/2 {
		batchSize = 1 // Process all at once
	} else {
		batchSize = max(1, maxBudget/estimatedInputTokens)
	}

	return streaming, batchSize
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getEnv gets an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
