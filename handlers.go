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

	// Get or create the appropriate provider
	provider, err := getOrCreateProvider(providerName, modelOverride)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Create prompt for LLM analysis
	prompt := llm.AnalysisPrompt("diff", diffContent, map[string]interface{}{
		"summarize": summarize,
	})

	// Get analysis from LLM
	analysis, err := provider.Analyze(ctx, prompt)
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

	// Get or create the appropriate provider
	provider, err := getOrCreateProvider(providerName, modelOverride)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Create prompt for LLM analysis
	prompt := llm.AnalysisPrompt("code_review", code, map[string]interface{}{
		"language": language,
		"focus":    focus,
	})

	// Get review from LLM
	review, err := provider.Analyze(ctx, prompt)
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

	info := getRepoInfo(repoPath)

	return mcp.NewToolResultText(info), nil
}

func handleCommitAnalysis(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	commitSHA := "HEAD"
	if sha, ok := request.GetArguments()["commit_sha"].(string); ok && sha != "" {
		commitSHA = sha
	}

	repoPath := "."
	if path, ok := request.GetArguments()["repo_path"].(string); ok && path != "" {
		repoPath = path
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

	// Get or create the appropriate provider
	provider, err := getOrCreateProvider(providerName, modelOverride)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get commit information
	commitInfo, err := getCommitInfo(repoPath, commitSHA)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Create prompt for LLM analysis
	prompt := llm.AnalysisPrompt("commit", commitInfo, nil)

	// Get analysis from LLM
	analysis, err := provider.Analyze(ctx, prompt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LLM analysis failed: %v", err)), nil
	}

	return mcp.NewToolResultText(analysis), nil
}

func getCommitInfo(repoPath, commitSHA string) (string, error) {
	var info strings.Builder

	// Get commit info with diff
	cmd := exec.Command("git", "-C", repoPath, "show", "--stat", commitSHA)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit info: %v", err)
	}
	info.WriteString(string(output))
	info.WriteString("\n\n")

	// Get the actual diff
	diffCmd := exec.Command("git", "-C", repoPath, "diff", commitSHA+"^", commitSHA)
	diffOutput, err := diffCmd.Output()
	if err != nil {
		// If this is the first commit, just get the full content
		diffCmd = exec.Command("git", "-C", repoPath, "show", commitSHA)
		diffOutput, _ = diffCmd.Output()
	}
	info.WriteString("Diff:\n")
	info.WriteString(string(diffOutput))

	return info.String(), nil
}

func getRepoInfo(repoPath string) string {
	var info strings.Builder

	// Get current branch
	branchCmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")
	branch, _ := branchCmd.Output()

	// Get remote URL
	remoteCmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	remote, _ := remoteCmd.Output()

	// Get recent commits
	logCmd := exec.Command("git", "-C", repoPath, "log", "--oneline", "-5")
	recentCommits, _ := logCmd.Output()

	// Get status
	statusCmd := exec.Command("git", "-C", repoPath, "status", "--short")
	status, _ := statusCmd.Output()

	info.WriteString("📁 Repository Information:\n\n")
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

	// Get or create the appropriate provider
	provider, err := getOrCreateProvider(providerName, modelOverride)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get uncommitted changes
	diffContent, err := getUncommittedChanges(repoPath, stagedOnly)
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

	// Get analysis from LLM
	analysis, err := provider.Analyze(ctx, prompt)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LLM analysis failed: %v", err)), nil
	}

	return mcp.NewToolResultText(analysis), nil
}

func getUncommittedChanges(repoPath string, stagedOnly bool) (string, error) {
	var info strings.Builder

	// Add header
	if stagedOnly {
		info.WriteString("📋 Staged Changes Analysis\n\n")
	} else {
		info.WriteString("📝 Uncommitted Work Analysis\n\n")
	}

	// Get status summary
	statusCmd := exec.Command("git", "-C", repoPath, "status", "--short")
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

	// Get diff
	var diffCmd *exec.Cmd
	if stagedOnly {
		// Get only staged changes
		diffCmd = exec.Command("git", "-C", repoPath, "diff", "--cached")
	} else {
		// Get all changes (staged and unstaged)
		diffCmd = exec.Command("git", "-C", repoPath, "diff", "HEAD")
	}

	diffOutput, err := diffCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %v", err)
	}

	// If no diff from HEAD, try to get staged changes
	if len(diffOutput) == 0 && !stagedOnly {
		diffCmd = exec.Command("git", "-C", repoPath, "diff", "--cached")
		diffOutput, _ = diffCmd.Output()
	}

	if len(diffOutput) > 0 {
		info.WriteString("Diff:\n")
		info.WriteString(string(diffOutput))
	}

	// Get statistics
	var statCmd *exec.Cmd
	if stagedOnly {
		statCmd = exec.Command("git", "-C", repoPath, "diff", "--cached", "--stat")
	} else {
		statCmd = exec.Command("git", "-C", repoPath, "diff", "HEAD", "--stat")
	}

	statOutput, _ := statCmd.Output()
	if len(statOutput) > 0 {
		info.WriteString("\n\nStatistics:\n")
		info.WriteString(string(statOutput))
	}

	return info.String(), nil
}
