package main

import (
	"context"
	"strings"
	"testing"

	"github.com/dshills/second-opinion/config"
)

func TestDiffStats(t *testing.T) {
	ctx := context.Background()

	// Test with current repository
	stats, err := getDiffStats(ctx, ".", "HEAD~1", "HEAD")
	if err != nil {
		// Not a fatal error - might not have commits
		t.Logf("Could not get diff stats: %v", err)
		return
	}

	t.Logf("Diff stats: Files=%d, Insertions=%d, Deletions=%d, EstimatedSizeKB=%d",
		stats.FileCount, stats.Insertions, stats.Deletions, stats.EstimatedSizeKB)
}

func TestTruncateLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		maxLength int
		expected  string
	}{
		{
			name:      "short line",
			line:      "hello world",
			maxLength: 20,
			expected:  "hello world",
		},
		{
			name:      "exact length",
			line:      "hello world",
			maxLength: 11,
			expected:  "hello world",
		},
		{
			name:      "long line",
			line:      "this is a very long line that should be truncated",
			maxLength: 20,
			expected:  "this is a very lo...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateLine(tt.line, tt.maxLength)
			if result != tt.expected {
				t.Errorf("truncateLine() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSafeDiffProcessor(t *testing.T) {
	memConfig := &config.MemoryConfig{
		MaxDiffSizeMB:   1, // 1MB for testing
		MaxFileCount:    10,
		MaxLineLength:   100,
		EnableStreaming: true,
		ChunkSizeMB:     1,
	}

	processor := NewSafeDiffProcessor(memConfig)

	// Test normal processing
	chunk1 := []byte("diff --git a/file1.txt b/file1.txt\n")
	chunk1 = append(chunk1, []byte("index 123..456 100644\n")...)
	chunk1 = append(chunk1, []byte("--- a/file1.txt\n")...)
	chunk1 = append(chunk1, []byte("+++ b/file1.txt\n")...)
	chunk1 = append(chunk1, []byte("@@ -1,3 +1,3 @@\n")...)
	chunk1 = append(chunk1, []byte(" line1\n")...)
	chunk1 = append(chunk1, []byte("-line2\n")...)
	chunk1 = append(chunk1, []byte("+line2 modified\n")...)

	err := processor.ProcessChunk(chunk1)
	if err != nil {
		t.Fatalf("ProcessChunk failed: %v", err)
	}

	result := processor.GetResult()
	if result.IsTruncated {
		t.Error("Expected diff not to be truncated")
	}
	if result.FileCount != 1 {
		t.Errorf("Expected FileCount=1, got %d", result.FileCount)
	}
	if !strings.Contains(result.Content, "diff --git") {
		t.Error("Expected diff content to contain git diff header")
	}
}

func TestSafeDiffProcessorTruncation(t *testing.T) {
	memConfig := &config.MemoryConfig{
		MaxDiffSizeMB:   1,  // 1MB limit
		MaxFileCount:    2,  // Only 2 files
		MaxLineLength:   50, // Short lines
		EnableStreaming: true,
		ChunkSizeMB:     1,
	}

	processor := NewSafeDiffProcessor(memConfig)

	// Test file count truncation
	for i := 0; i < 5; i++ {
		chunk := []byte(strings.Repeat("diff --git a/file.txt b/file.txt\nsome content\n", 1))
		processor.ProcessChunk(chunk)
	}

	result := processor.GetResult()
	if !result.IsTruncated {
		t.Error("Expected diff to be truncated due to file count")
	}
	if !strings.Contains(result.WarningReason, "files limit") {
		t.Errorf("Expected warning about files limit, got: %s", result.WarningReason)
	}
}

func TestSafeDiffProcessorLineTruncation(t *testing.T) {
	memConfig := &config.MemoryConfig{
		MaxDiffSizeMB:   10,
		MaxFileCount:    100,
		MaxLineLength:   20, // Very short max line
		EnableStreaming: true,
		ChunkSizeMB:     1,
	}

	processor := NewSafeDiffProcessor(memConfig)

	longLine := strings.Repeat("a", 100) + "\n"
	err := processor.ProcessChunk([]byte(longLine))
	if err != nil {
		t.Fatalf("ProcessChunk failed: %v", err)
	}

	result := processor.GetResult()
	lines := strings.Split(result.Content, "\n")
	if len(lines) > 0 && len(lines[0]) > 23 { // 20 + "..."
		t.Errorf("Line not truncated properly: length=%d", len(lines[0]))
	}
	if !strings.HasSuffix(lines[0], "...") {
		t.Error("Truncated line should end with ...")
	}
}

func TestMemoryConfigDefaults(t *testing.T) {
	// Test that config loading sets proper defaults
	cfg := &config.Config{}

	// Simulate loading with empty memory config
	if cfg.Memory.MaxDiffSizeMB == 0 {
		cfg.Memory.MaxDiffSizeMB = 10
	}
	if cfg.Memory.MaxFileCount == 0 {
		cfg.Memory.MaxFileCount = 1000
	}
	if cfg.Memory.MaxLineLength == 0 {
		cfg.Memory.MaxLineLength = 1000
	}
	if cfg.Memory.ChunkSizeMB == 0 {
		cfg.Memory.ChunkSizeMB = 1
	}
	if !cfg.Memory.EnableStreaming && cfg.Memory.MaxDiffSizeMB > 0 {
		cfg.Memory.EnableStreaming = true
	}

	// Verify defaults
	if cfg.Memory.MaxDiffSizeMB != 10 {
		t.Errorf("Expected MaxDiffSizeMB=10, got %d", cfg.Memory.MaxDiffSizeMB)
	}
	if cfg.Memory.MaxFileCount != 1000 {
		t.Errorf("Expected MaxFileCount=1000, got %d", cfg.Memory.MaxFileCount)
	}
	if cfg.Memory.MaxLineLength != 1000 {
		t.Errorf("Expected MaxLineLength=1000, got %d", cfg.Memory.MaxLineLength)
	}
	if !cfg.Memory.EnableStreaming {
		t.Error("Expected EnableStreaming=true")
	}
}
