package config

import (
	"testing"
)

func TestGetOptimalTokensForDiff(t *testing.T) {
	cfg := &Config{}

	tests := []struct {
		name        string
		diffSizeKB  int
		expectedMin int
		expectedMax int
	}{
		{
			name:        "Very small diff (2KB)",
			diffSizeKB:  2 * 1024,
			expectedMin: 4000,
			expectedMax: 5000,
		},
		{
			name:        "Small diff (10KB)",
			diffSizeKB:  10 * 1024,
			expectedMin: 6000,
			expectedMax: 7000,
		},
		{
			name:        "Medium diff (30KB)",
			diffSizeKB:  30 * 1024,
			expectedMin: 8000,
			expectedMax: 9000,
		},
		{
			name:        "Large diff (100KB)",
			diffSizeKB:  100 * 1024,
			expectedMin: 12000,
			expectedMax: 13000,
		},
		{
			name:        "Very large diff (300KB)",
			diffSizeKB:  300 * 1024,
			expectedMin: 16000,
			expectedMax: 17000,
		},
		{
			name:        "Huge diff (1MB)",
			diffSizeKB:  1024 * 1024,
			expectedMin: 32000,
			expectedMax: 33000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := cfg.GetOptimalTokensForDiff(tt.diffSizeKB)

			if tokens < tt.expectedMin || tokens > tt.expectedMax {
				t.Errorf("GetOptimalTokensForDiff(%d) = %d, expected between %d and %d",
					tt.diffSizeKB, tokens, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestGetOptimalTemperatureForTask(t *testing.T) {
	cfg := &Config{}

	tests := []struct {
		name     string
		task     AnalysisTask
		expected float64
	}{
		{
			name:     "Security review should be very deterministic",
			task:     TaskSecurityReview,
			expected: 0.1,
		},
		{
			name:     "Code review should be mostly deterministic",
			task:     TaskCodeReview,
			expected: 0.2,
		},
		{
			name:     "Commit analysis should be mostly deterministic",
			task:     TaskCommitAnalysis,
			expected: 0.2,
		},
		{
			name:     "Diff analysis should be slightly flexible",
			task:     TaskDiffAnalysis,
			expected: 0.25,
		},
		{
			name:     "Architecture review should allow creativity",
			task:     TaskArchitectureReview,
			expected: 0.3,
		},
		{
			name:     "General tasks should be balanced",
			task:     TaskGeneral,
			expected: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temperature := cfg.GetOptimalTemperatureForTask(tt.task)

			if temperature != tt.expected {
				t.Errorf("GetOptimalTemperatureForTask(%s) = %f, expected %f",
					tt.task, temperature, tt.expected)
			}
		})
	}
}

func TestGetProviderOptimizedConfig(t *testing.T) {
	cfg := &Config{
		Memory: MemoryConfig{
			MaxDiffSizeMB: 10,
			MaxFileCount:  1000,
		},
	}

	tests := []struct {
		name           string
		provider       string
		diffSize       int
		task           AnalysisTask
		expectedTokens int
		expectedTemp   float64
	}{
		{
			name:           "OpenAI with small diff",
			provider:       "openai",
			diffSize:       1024, // 1KB
			task:           TaskCodeReview,
			expectedTokens: 4096,
			expectedTemp:   0.2,
		},
		{
			name:           "Google with large diff",
			provider:       "google",
			diffSize:       100 * 1024, // 100KB
			task:           TaskSecurityReview,
			expectedTokens: 8192, // Should be capped
			expectedTemp:   0.1,
		},
		{
			name:           "Mistral with medium diff",
			provider:       "mistral",
			diffSize:       30 * 1024, // 30KB
			task:           TaskDiffAnalysis,
			expectedTokens: 8192,
			expectedTemp:   0.25,
		},
		{
			name:           "Ollama with large diff",
			provider:       "ollama",
			diffSize:       200 * 1024, // 200KB
			task:           TaskArchitectureReview,
			expectedTokens: 8192, // Should be capped
			expectedTemp:   0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxTokens, temperature, providerConfig := cfg.GetProviderOptimizedConfig(
				tt.provider, tt.diffSize, tt.task)

			if maxTokens != tt.expectedTokens {
				t.Errorf("GetProviderOptimizedConfig maxTokens = %d, expected %d",
					maxTokens, tt.expectedTokens)
			}

			if temperature != tt.expectedTemp {
				t.Errorf("GetProviderOptimizedConfig temperature = %f, expected %f",
					temperature, tt.expectedTemp)
			}

			if providerConfig == nil {
				t.Error("Expected non-nil providerConfig")
			}
		})
	}
}

func TestShouldChunkDiff(t *testing.T) {
	cfg := &Config{
		Memory: MemoryConfig{
			MaxDiffSizeMB: 5,   // 5MB limit
			MaxFileCount:  100, // 100 files limit
			ChunkSizeMB:   1,   // 1MB chunks
		},
	}

	tests := []struct {
		name              string
		diffSizeBytes     int
		fileCount         int
		expectedChunk     bool
		expectedChunkSize int
	}{
		{
			name:              "Small diff should not chunk",
			diffSizeBytes:     1024 * 1024, // 1MB
			fileCount:         10,
			expectedChunk:     false,
			expectedChunkSize: 1024 * 1024, // 1MB
		},
		{
			name:              "Large diff should chunk",
			diffSizeBytes:     10 * 1024 * 1024, // 10MB
			fileCount:         50,
			expectedChunk:     true,
			expectedChunkSize: 1024 * 1024, // 1MB
		},
		{
			name:              "Many files should chunk with smaller chunks",
			diffSizeBytes:     2 * 1024 * 1024, // 2MB
			fileCount:         150,             // > 100 files
			expectedChunk:     true,
			expectedChunkSize: 512 * 1024, // 512KB (halved)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldChunk, chunkSize := cfg.ShouldChunkDiff(tt.diffSizeBytes, tt.fileCount)

			if shouldChunk != tt.expectedChunk {
				t.Errorf("ShouldChunkDiff shouldChunk = %v, expected %v",
					shouldChunk, tt.expectedChunk)
			}

			if chunkSize != tt.expectedChunkSize {
				t.Errorf("ShouldChunkDiff chunkSize = %d, expected %d",
					chunkSize, tt.expectedChunkSize)
			}
		})
	}
}

func TestEstimateTokensForText(t *testing.T) {
	cfg := &Config{}

	tests := []struct {
		name           string
		text           string
		expectedTokens int
	}{
		{
			name:           "Empty text",
			text:           "",
			expectedTokens: 0,
		},
		{
			name:           "Short text",
			text:           "Hello world",
			expectedTokens: 2, // 11 chars / 4 = 2.75 -> 2
		},
		{
			name:           "Medium text",
			text:           "This is a longer piece of text that should estimate to more tokens",
			expectedTokens: 16, // 67 chars / 4 = 16.75 -> 16
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := cfg.EstimateTokensForText(tt.text)

			if tokens != tt.expectedTokens {
				t.Errorf("EstimateTokensForText(%q) = %d, expected %d",
					tt.text, tokens, tt.expectedTokens)
			}
		})
	}
}

func TestGetMemoryOptimizedConfig(t *testing.T) {
	cfg := &Config{
		Memory: MemoryConfig{
			EnableStreaming: true,
		},
	}

	tests := []struct {
		name                 string
		estimatedInputTokens int
		expectedStreaming    bool
		expectedBatchSize    int
	}{
		{
			name:                 "Small input should not force streaming",
			estimatedInputTokens: 2000,
			expectedStreaming:    true, // Already enabled in config
			expectedBatchSize:    1,
		},
		{
			name:                 "Large input should force streaming",
			estimatedInputTokens: 10000,
			expectedStreaming:    true,
			expectedBatchSize:    1, // 16384 / 10000 = 1.6 -> 1
		},
		{
			name:                 "Very large input should use batching",
			estimatedInputTokens: 20000,
			expectedStreaming:    true,
			expectedBatchSize:    1, // max(1, 16384/20000) = 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streaming, batchSize := cfg.GetMemoryOptimizedConfig(tt.estimatedInputTokens)

			if streaming != tt.expectedStreaming {
				t.Errorf("GetMemoryOptimizedConfig streaming = %v, expected %v",
					streaming, tt.expectedStreaming)
			}

			if batchSize != tt.expectedBatchSize {
				t.Errorf("GetMemoryOptimizedConfig batchSize = %d, expected %d",
					batchSize, tt.expectedBatchSize)
			}
		})
	}
}
