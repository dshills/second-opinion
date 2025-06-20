package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/dshills/second-opinion/llm"
	"github.com/mark3labs/mcp-go/mcp"
)

func handleGitDiff(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	diffContent, err := request.RequireString("diff_content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	summarize := false
	if s, ok := request.GetArguments()["summarize"].(bool); ok {
		summarize = s
	}

	// Get provider and model from request
	providerName := ""
	if p, ok := request.GetArguments()["provider"].(string); ok {
		providerName = p
	}

	modelOverride := ""
	if m, ok := request.GetArguments()["model"].(string); ok {
		modelOverride = m
	}

	// Get or create the appropriate optimized provider
	optimizedProvider, err := getOrCreateOptimizedProvider(providerName, modelOverride)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Create prompt for LLM analysis
	prompt := llm.AnalysisPrompt("diff", diffContent, map[string]interface{}{
		"summarize": summarize,
	})

	// Get analysis from LLM using optimization
	contentSize := len(diffContent)
	task := llm.GetTaskFromAnalysisType("diff")
	analysis, err := optimizedProvider.AnalyzeOptimized(ctx, prompt, contentSize, task)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LLM analysis failed: %v", err)), nil
	}

	return mcp.NewToolResultText(analysis), nil
}

func handleCodeReview(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	language := ""
	if lang, ok := request.GetArguments()["language"].(string); ok {
		language = lang
	}

	focus := "all"
	if f, ok := request.GetArguments()["focus"].(string); ok {
		focus = f
	}

	// Get provider and model from request
	providerName := ""
	if p, ok := request.GetArguments()["provider"].(string); ok {
		providerName = p
	}

	modelOverride := ""
	if m, ok := request.GetArguments()["model"].(string); ok {
		modelOverride = m
	}

	// Get or create the appropriate optimized provider
	optimizedProvider, err := getOrCreateOptimizedProvider(providerName, modelOverride)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Create prompt for LLM analysis
	prompt := llm.AnalysisPrompt("code_review", code, map[string]interface{}{
		"language": language,
		"focus":    focus,
	})

	// Get review from LLM using optimization
	contentSize := len(code)
	task := llm.GetTaskFromAnalysisType("code_review")
	// If focus is security, use security-specific task
	if focus == "security" {
		task = llm.GetTaskFromAnalysisType("security")
	}
	review, err := optimizedProvider.AnalyzeOptimized(ctx, prompt, contentSize, task)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LLM review failed: %v", err)), nil
	}

	return mcp.NewToolResultText(review), nil
}

func handleRepoInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := "."
	if path, ok := request.GetArguments()["repo_path"].(string); ok && path != "" {
		repoPath = path
	}

	// Validate repo path
	validPath, err := validateRepoPath(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid repository path: %v", err)), nil
	}

	info := getRepoInfo(ctx, validPath)

	return mcp.NewToolResultText(info), nil
}

func handleCommitAnalysis(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	commitSHA := "HEAD"
	if sha, ok := request.GetArguments()["commit_sha"].(string); ok && sha != "" {
		commitSHA = sha
	}

	// Validate commit SHA
	if err := validateCommitSHA(commitSHA); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid commit SHA: %v", err)), nil
	}

	repoPath := "."
	if path, ok := request.GetArguments()["repo_path"].(string); ok && path != "" {
		repoPath = path
	}

	// Validate repo path
	validPath, err := validateRepoPath(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid repository path: %v", err)), nil
	}

	// Get provider and model from request
	providerName := ""
	if p, ok := request.GetArguments()["provider"].(string); ok {
		providerName = p
	}

	modelOverride := ""
	if m, ok := request.GetArguments()["model"].(string); ok {
		modelOverride = m
	}

	// Get or create the appropriate optimized provider
	optimizedProvider, err := getOrCreateOptimizedProvider(providerName, modelOverride)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get commit information
	commitInfo, err := getCommitInfo(ctx, validPath, commitSHA)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Create prompt for LLM analysis
	prompt := llm.AnalysisPrompt("commit", commitInfo, nil)

	// Get analysis from LLM using optimization
	contentSize := len(commitInfo)
	task := llm.GetTaskFromAnalysisType("commit")
	analysis, err := optimizedProvider.AnalyzeOptimized(ctx, prompt, contentSize, task)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LLM analysis failed: %v", err)), nil
	}

	return mcp.NewToolResultText(analysis), nil
}

func getCommitInfo(ctx context.Context, repoPath, commitSHA string) (string, error) {
	var info strings.Builder

	// Get commit info with diff
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "show", "--stat", commitSHA)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit info: %v", err)
	}
	info.WriteString(string(output))
	info.WriteString("\n\n")

	// Get the actual diff using safe memory-limited approach
	memConfig := &cfg.Memory
	truncatedDiff, err := getGitDiffSafe(ctx, repoPath, memConfig, commitSHA+"^", commitSHA)
	if err != nil {
		// If this is the first commit, try to get the full content
		truncatedDiff, err = getGitDiffSafe(ctx, repoPath, memConfig, commitSHA)
		if err != nil {
			// If both commands fail, return a meaningful error
			return "", fmt.Errorf("failed to get commit diff: %v", err)
		}
		info.WriteString("Diff (first commit):\n")
	} else {
		info.WriteString("Diff:\n")
	}

	// Add warning if truncated
	if truncatedDiff.IsTruncated {
		info.WriteString(fmt.Sprintf("\n⚠️ WARNING: %s\n", truncatedDiff.WarningReason))
		info.WriteString(fmt.Sprintf("Total size: %dKB, Files: %d\n\n", truncatedDiff.TotalSizeKB, truncatedDiff.FileCount))
	}

	info.WriteString(truncatedDiff.Content)

	return info.String(), nil
}

func getRepoInfo(ctx context.Context, repoPath string) string {
	var info strings.Builder
	var warnings []string

	// Get current branch
	branchCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "branch", "--show-current")
	branch, err := branchCmd.Output()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Failed to get current branch: %v", err))
		branch = []byte("unknown")
	}

	// Get remote URL
	remoteCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "remote", "get-url", "origin")
	remote, err := remoteCmd.Output()
	if err != nil {
		// This is common for repos without remotes, so just note it
		remote = []byte("(no remote configured)")
	}

	// Get recent commits
	logCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "log", "--oneline", "-5")
	recentCommits, err := logCmd.Output()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Failed to get commit history: %v", err))
		recentCommits = []byte("(unable to retrieve commit history)")
	}

	// Get status
	statusCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "status", "--short")
	status, err := statusCmd.Output()
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Failed to get repository status: %v", err))
	}

	info.WriteString("📁 Repository Information:\n\n")

	// Add any warnings at the top
	if len(warnings) > 0 {
		info.WriteString("⚠️ Warnings:\n")
		for _, warning := range warnings {
			info.WriteString(fmt.Sprintf("- %s\n", warning))
		}
		info.WriteString("\n")
	}

	info.WriteString(fmt.Sprintf("Branch: %s", branch))
	info.WriteString(fmt.Sprintf("Remote: %s", remote))
	info.WriteString("\nRecent commits:\n")
	info.WriteString(string(recentCommits))

	if len(status) > 0 {
		info.WriteString("\n⚠️ Uncommitted changes:\n")
		info.WriteString(string(status))
	}

	return info.String()
}

func handleAnalyzeUncommittedWork(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := "."
	if path, ok := request.GetArguments()["repo_path"].(string); ok && path != "" {
		repoPath = path
	}

	// Validate repo path
	validPath, err := validateRepoPath(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid repository path: %v", err)), nil
	}

	stagedOnly := false
	if staged, ok := request.GetArguments()["staged_only"].(bool); ok {
		stagedOnly = staged
	}

	// Get provider and model from request
	providerName := ""
	if p, ok := request.GetArguments()["provider"].(string); ok {
		providerName = p
	}

	modelOverride := ""
	if m, ok := request.GetArguments()["model"].(string); ok {
		modelOverride = m
	}

	// Get or create the appropriate optimized provider
	optimizedProvider, err := getOrCreateOptimizedProvider(providerName, modelOverride)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get uncommitted changes
	diffContent, err := getUncommittedChanges(ctx, validPath, stagedOnly)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if diffContent == "" {
		return mcp.NewToolResultText("No uncommitted changes found."), nil
	}

	// Create prompt for LLM analysis
	prompt := llm.AnalysisPrompt("uncommitted_work", diffContent, map[string]any{
		"staged_only": stagedOnly,
	})

	// Get analysis from LLM using optimization
	contentSize := len(diffContent)
	task := llm.GetTaskFromAnalysisType("uncommitted_work")
	analysis, err := optimizedProvider.AnalyzeOptimized(ctx, prompt, contentSize, task)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LLM analysis failed: %v", err)), nil
	}

	return mcp.NewToolResultText(analysis), nil
}

func getUncommittedChanges(ctx context.Context, repoPath string, stagedOnly bool) (string, error) {
	var info strings.Builder

	// Add header
	if stagedOnly {
		info.WriteString("📋 Staged Changes Analysis\n\n")
	} else {
		info.WriteString("📝 Uncommitted Work Analysis\n\n")
	}

	// Get status summary
	statusCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "status", "--short")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git status: %v", err)
	}

	if len(statusOutput) == 0 {
		return "", nil
	}

	info.WriteString("Files changed:\n")
	info.WriteString(string(statusOutput))
	info.WriteString("\n")

	// Get diff using safe memory-limited approach
	memConfig := &cfg.Memory
	var truncatedDiff *TruncatedDiff

	if stagedOnly {
		// Get only staged changes
		truncatedDiff, err = getGitDiffSafe(ctx, repoPath, memConfig, "--cached")
	} else {
		// Get all changes (staged and unstaged)
		truncatedDiff, err = getGitDiffSafe(ctx, repoPath, memConfig, "HEAD")
	}

	if err != nil {
		return "", fmt.Errorf("failed to get diff: %v", err)
	}

	// If no diff from HEAD, try to get staged changes
	if truncatedDiff.Content == "" && !stagedOnly {
		stagedDiff, err := getGitDiffSafe(ctx, repoPath, memConfig, "--cached")
		if err != nil {
			// Log the error but continue since we might have unstaged changes
			info.WriteString(fmt.Sprintf("\nNote: Failed to get staged changes: %v\n", err))
		} else {
			truncatedDiff = stagedDiff
		}
	}

	if truncatedDiff.Content != "" {
		// Add warning if truncated
		if truncatedDiff.IsTruncated {
			info.WriteString(fmt.Sprintf("\n⚠️ WARNING: %s\n", truncatedDiff.WarningReason))
			info.WriteString(fmt.Sprintf("Total size: %dKB, Files: %d\n\n", truncatedDiff.TotalSizeKB, truncatedDiff.FileCount))
		}

		info.WriteString("Diff:\n")
		info.WriteString(truncatedDiff.Content)
	}

	// Get statistics
	var statCmd *exec.Cmd
	if stagedOnly {
		statCmd = exec.CommandContext(ctx, "git", "-C", repoPath, "diff", "--cached", "--stat")
	} else {
		statCmd = exec.CommandContext(ctx, "git", "-C", repoPath, "diff", "HEAD", "--stat")
	}

	statOutput, _ := statCmd.Output()
	if len(statOutput) > 0 {
		info.WriteString("\n\nStatistics:\n")
		info.WriteString(string(statOutput))
	}

	return info.String(), nil
}
