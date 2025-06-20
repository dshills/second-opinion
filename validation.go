package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// gitSHARegex validates git commit SHA (abbreviated or full)
	gitSHARegex = regexp.MustCompile(`^[a-fA-F0-9]{4,40}$`)

	// headRefRegex validates HEAD references
	headRefRegex = regexp.MustCompile(`^HEAD(~\d+)?(\^\d*)?$`)
)

// validateRepoPath validates and cleans a repository path
func validateRepoPath(path string) (string, error) {
	if path == "" || path == "." {
		// Default to current directory
		return ".", nil
	}

	// Clean the path
	cleanPath := filepath.Clean(path)

	// Convert to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Ensure the path is within or is the current working directory
	if !strings.HasPrefix(absPath, cwd) && absPath != cwd {
		return "", fmt.Errorf("path must be within the current working directory")
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("path does not exist: %w", err)
	}

	// Check if it's a git repository
	gitDir := filepath.Join(absPath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return "", fmt.Errorf("not a git repository (no .git directory found)")
	}

	return cleanPath, nil
}

// validateCommitSHA validates a git commit reference
func validateCommitSHA(sha string) error {
	if sha == "" || sha == "HEAD" {
		// Empty or HEAD is valid
		return nil
	}

	// Check if it's a HEAD reference (HEAD~1, HEAD^, etc.)
	if strings.HasPrefix(sha, "HEAD") {
		if headRefRegex.MatchString(sha) {
			return nil
		}
		return fmt.Errorf("invalid HEAD reference format")
	}

	// Check if it's a valid SHA
	if !gitSHARegex.MatchString(sha) {
		return fmt.Errorf("invalid git commit SHA format")
	}

	return nil
}
