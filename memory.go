package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/dshills/second-opinion/config"
)

const (
	// DefaultMaxDiffSize is the default maximum diff size in bytes (10MB)
	DefaultMaxDiffSize = 10 * 1024 * 1024
	// DefaultMaxFileCount is the default maximum number of files in a diff
	DefaultMaxFileCount = 1000
	// DefaultMaxLineLength is the default maximum line length before truncation
	DefaultMaxLineLength = 1000
	// DefaultChunkSize is the default chunk size for streaming (1MB)
	DefaultChunkSize = 1024 * 1024
)

// DiffStats holds statistics about a diff
type DiffStats struct {
	FileCount       int
	Insertions      int
	Deletions       int
	EstimatedSizeKB int64
}

// TruncatedDiff represents a potentially truncated diff
type TruncatedDiff struct {
	Content       string
	IsTruncated   bool
	TotalSizeKB   int64
	FileCount     int
	TruncatedAt   string
	WarningReason string
}

// getDiffStats gets statistics about a diff without loading the full content
func getDiffStats(ctx context.Context, repoPath string, args ...string) (*DiffStats, error) {
	// Build command arguments
	cmdArgs := []string{"-C", repoPath, "diff", "--numstat"}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get diff stats: %w", err)
	}

	stats := &DiffStats{}
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse numstat format: added deleted filename
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			stats.FileCount++

			// Handle binary files (shown as "-")
			if added, err := strconv.Atoi(parts[0]); err == nil {
				stats.Insertions += added
			}
			if deleted, err := strconv.Atoi(parts[1]); err == nil {
				stats.Deletions += deleted
			}
		}
	}

	// Estimate size: assume average line length of 50 bytes
	stats.EstimatedSizeKB = int64(stats.Insertions+stats.Deletions) * 50 / 1024

	return stats, nil
}

// checkDiffSize checks if a diff is within acceptable size limits
func checkDiffSize(ctx context.Context, repoPath string, memConfig *config.MemoryConfig, args ...string) error {
	stats, err := getDiffStats(ctx, repoPath, args...)
	if err != nil {
		return err
	}

	maxSizeKB := int64(memConfig.MaxDiffSizeMB * 1024)
	if stats.EstimatedSizeKB > maxSizeKB {
		return fmt.Errorf("diff too large: estimated %dKB exceeds limit of %dKB",
			stats.EstimatedSizeKB, maxSizeKB)
	}

	if stats.FileCount > memConfig.MaxFileCount {
		return fmt.Errorf("too many files changed: %d exceeds limit of %d",
			stats.FileCount, memConfig.MaxFileCount)
	}

	return nil
}

// streamCommand runs a command and processes output in chunks
func streamCommand(ctx context.Context, processor func([]byte) error, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Read in chunks
	buf := make([]byte, DefaultChunkSize)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			if procErr := processor(buf[:n]); procErr != nil {
				cmd.Process.Kill()
				return procErr
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read error: %w", err)
		}
	}

	return cmd.Wait()
}

// truncateLine truncates a line if it exceeds maxLength
func truncateLine(line string, maxLength int) string {
	if len(line) <= maxLength {
		return line
	}
	return line[:maxLength-3] + "..."
}

// SafeDiffProcessor processes diffs with memory safety
type SafeDiffProcessor struct {
	memConfig   *config.MemoryConfig
	buffer      *bytes.Buffer
	lineBuffer  []byte
	bytesRead   int64
	linesRead   int
	filesRead   int
	isTruncated bool
	truncateMsg string
}

// NewSafeDiffProcessor creates a new safe diff processor
func NewSafeDiffProcessor(memConfig *config.MemoryConfig) *SafeDiffProcessor {
	return &SafeDiffProcessor{
		memConfig:  memConfig,
		buffer:     &bytes.Buffer{},
		lineBuffer: make([]byte, 0, memConfig.MaxLineLength*2),
	}
}

// ProcessChunk processes a chunk of diff output
func (p *SafeDiffProcessor) ProcessChunk(chunk []byte) error {
	if p.isTruncated {
		return nil // Already truncated, ignore rest
	}

	// Check total size limit
	maxBytes := int64(p.memConfig.MaxDiffSizeMB * 1024 * 1024)
	if p.bytesRead+int64(len(chunk)) > maxBytes {
		p.isTruncated = true
		p.truncateMsg = fmt.Sprintf("Diff truncated at %dMB limit", p.memConfig.MaxDiffSizeMB)
		return nil
	}

	p.bytesRead += int64(len(chunk))

	// Process line by line
	for _, b := range chunk {
		if b == '\n' {
			// Process complete line
			line := string(p.lineBuffer)

			// Count files
			if strings.HasPrefix(line, "diff --git") {
				p.filesRead++
				if p.filesRead > p.memConfig.MaxFileCount {
					p.isTruncated = true
					p.truncateMsg = fmt.Sprintf("Truncated at %d files limit", p.memConfig.MaxFileCount)
					return nil
				}
			}

			// Truncate long lines
			line = truncateLine(line, p.memConfig.MaxLineLength)

			// Write to buffer
			p.buffer.WriteString(line)
			p.buffer.WriteByte('\n')
			p.linesRead++

			// Reset line buffer
			p.lineBuffer = p.lineBuffer[:0]
		} else {
			p.lineBuffer = append(p.lineBuffer, b)
		}
	}

	return nil
}

// GetResult returns the processed diff result
func (p *SafeDiffProcessor) GetResult() *TruncatedDiff {
	// Handle any remaining line
	if len(p.lineBuffer) > 0 {
		line := truncateLine(string(p.lineBuffer), p.memConfig.MaxLineLength)
		p.buffer.WriteString(line)
		p.buffer.WriteByte('\n')
	}

	return &TruncatedDiff{
		Content:       p.buffer.String(),
		IsTruncated:   p.isTruncated,
		TotalSizeKB:   p.bytesRead / 1024,
		FileCount:     p.filesRead,
		TruncatedAt:   p.truncateMsg,
		WarningReason: p.truncateMsg,
	}
}

// getGitDiffSafe safely retrieves a git diff with memory limits
func getGitDiffSafe(ctx context.Context, repoPath string, memConfig *config.MemoryConfig, args ...string) (*TruncatedDiff, error) {
	// First check if diff is within limits
	if err := checkDiffSize(ctx, repoPath, memConfig, args...); err != nil {
		// Get stats for the warning
		stats, _ := getDiffStats(ctx, repoPath, args...)
		return &TruncatedDiff{
			Content:       "",
			IsTruncated:   true,
			TotalSizeKB:   stats.EstimatedSizeKB,
			FileCount:     stats.FileCount,
			WarningReason: err.Error(),
		}, nil
	}

	processor := NewSafeDiffProcessor(memConfig)

	// Build command arguments
	cmdArgs := []string{"-C", repoPath, "diff"}
	cmdArgs = append(cmdArgs, args...)

	// If streaming is enabled, use streaming approach
	if memConfig.EnableStreaming {
		err := streamCommand(ctx, processor.ProcessChunk, "git", cmdArgs...)
		if err != nil && !processor.isTruncated {
			return nil, fmt.Errorf("git diff failed: %w", err)
		}
	} else {
		// Fall back to regular execution with size limits
		cmd := exec.CommandContext(ctx, "git", cmdArgs...)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("git diff failed: %w", err)
		}

		if err := processor.ProcessChunk(output); err != nil {
			return nil, err
		}
	}

	return processor.GetResult(), nil
}
