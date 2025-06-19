# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project named `second-opinion` (module: `github.com/dshills/second-opinion`). The project uses Go 1.24.4 and implements an MCP (Model Context Protocol) server that assists Claude Code in reviewing commits and code bases.

The server provides tools for:
- Analyzing git diffs to understand code changes
- Reviewing code for quality, security, and best practices  
- Analyzing git commits for quality and adherence to standards
- Getting repository information

## Development Commands

### Linting
```bash
# Run all linters configured in .golangci.yml
golangci-lint run

# Run with verbose output
golangci-lint run -v

# Run on specific directories
golangci-lint run ./...

# Auto-fix issues where possible
golangci-lint run --fix
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run a specific test
go test -run TestName ./...

# Run tests with race detector
go test -race ./...
```

### Building
```bash
# Build the project
go build -o bin/second-opinion

# Build with race detector
go build -race -o bin/second-opinion

# Install dependencies
go mod download

# Tidy dependencies
go mod tidy
```

### Running the MCP Server
```bash
# Run directly
go run main.go

# Or run the built binary
./bin/second-opinion
```

## Code Architecture

The project is an MCP server implementation using the `github.com/mark3labs/mcp-go` library. The architecture consists of:

1. **Main Server** (`main.go`): Initializes the MCP server and registers tool handlers
2. **Tool Handlers**: Four main tools are implemented:
   - `analyze_git_diff`: Analyzes git diff output with file counts and change statistics
   - `review_code`: Reviews code for security, style, and performance issues
   - `analyze_commit`: Analyzes commits including diff stats and commit message quality
   - `get_repo_info`: Provides repository status, branch, and recent commit information

3. **Helper Functions**: Internal functions for diff analysis, code review logic, and git command execution

When implementing new features:
1. **Adding Tools**: Define new tools using `mcp.NewTool()` with proper descriptions and parameters
2. **Handler Functions**: Implement handlers with signature `func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)`
3. **Error Handling**: Use `mcp.NewToolResultError()` for errors and `mcp.NewToolResultText()` for successful responses
4. **Testing**: Place test files alongside source files with `_test.go` suffix

## Linting Configuration

The project uses golangci-lint with an extensive configuration (.golangci.yml) that includes:
- **Security**: gosec for security issues
- **Style**: gofmt, goimports for formatting
- **Quality**: staticcheck, revive, ineffassign, and many others
- **Testing**: Test files are included in linting
- Custom exclusions for generated files and test code

Key linting behaviors:
- Enforces struct field alignment for memory efficiency
- Requires explicit type conversions
- Checks for SQL injection vulnerabilities
- Validates error handling patterns
- Enforces consistent import grouping

## Environment Configuration

The project supports multiple environments through .env files:
- `.env` - Default environment variables
- `.env.local`, `.env.*.local` - Local overrides (git-ignored)
- `.env.development`, `.env.test`, `.env.production` - Environment-specific configs