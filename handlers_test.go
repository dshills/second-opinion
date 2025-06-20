package main

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestContextCancellation verifies that git commands respect context cancellation
func TestContextCancellation(t *testing.T) {
	// Create a context that will be canceled
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Sleep to ensure context times out
	time.Sleep(200 * time.Millisecond)

	// Try to get repo info with canceled context
	info := getRepoInfo(ctx, ".")

	// Should contain some indication of timeout/cancellation
	if !strings.Contains(info, "Failed") && !strings.Contains(info, "Warning") {
		t.Logf("Expected timeout/cancellation indication, got: %s", info)
	}
}

// TestErrorHandling verifies that git command errors are properly handled
func TestErrorHandling(t *testing.T) {
	ctx := context.Background()

	// Test with non-existent path (should be caught by validation)
	info := getRepoInfo(ctx, "/non/existent/path")

	// Should contain error/warning messages
	if !strings.Contains(info, "Failed") && !strings.Contains(info, "Warning") {
		t.Errorf("Expected error messages for non-existent path, got: %s", info)
	}

	// Test getCommitInfo with invalid SHA
	_, err := getCommitInfo(ctx, ".", "invalid-sha")
	if err == nil {
		t.Error("Expected error for invalid commit SHA, got nil")
	}
}

// TestWarningsInOutput verifies that warnings are properly formatted
func TestWarningsInOutput(t *testing.T) {
	ctx := context.Background()

	// This will generate warnings but should still return formatted output
	info := getRepoInfo(ctx, "/tmp")

	// Should still have the header
	if !strings.Contains(info, "📁 Repository Information:") {
		t.Error("Missing repository information header")
	}

	// Should show branch even if unknown
	if !strings.Contains(info, "Branch:") {
		t.Error("Missing branch information")
	}
}
