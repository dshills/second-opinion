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
	cfg.Google.Model = getEnv("GOOGLE_MODEL", "gemini-1.5-flash")

	cfg.Ollama.Endpoint = getEnv("OLLAMA_ENDPOINT", "http://localhost:11434")
	cfg.Ollama.Model = getEnv("OLLAMA_MODEL", "llama3.2")

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
		// Return config for default provider
		return c.GetProviderConfig(c.DefaultProvider)
	}
}

// getEnv gets an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
