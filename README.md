# Second Opinion 🔍

An MCP (Model Context Protocol) server that assists Claude Code in reviewing commits and code bases. This tool leverages external LLMs (OpenAI, Google Gemini, Ollama, Mistral) to provide intelligent code review capabilities, git diff analysis, commit quality assessment, and uncommitted work analysis.

## Features

- **Git Diff Analysis**: Analyze git diff output to understand code changes using LLMs
- **Code Review**: Review code for quality, security, and best practices with AI assistance
- **Commit Analysis**: Analyze git commits for quality and adherence to best practices
- **Uncommitted Work Analysis**: Analyze all uncommitted changes or just staged changes
- **Repository Information**: Get information about git repositories
- **Multiple LLM Support**: Works with OpenAI, Google Gemini, Ollama (local), and Mistral AI
- **🚀 Smart Optimization**: Dynamic token allocation and task-specific temperature tuning
- **⚡ Performance Tuning**: Provider-specific optimizations and memory-aware chunking
- **Security**: Input validation, secure path handling, and API key protection
- **Memory Safety**: Configurable memory limits and streaming support for large diffs

## Installation

### Prerequisites
- Go 1.20 or higher
- Git
- Claude Code Desktop app

### Build from Source

1. Clone the repository:
```bash
git clone https://github.com/dshills/second-opinion.git
cd second-opinion
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the server:
```bash
go build -o bin/second-opinion
```

## Configuration

Second Opinion supports two configuration methods, with the following priority order:

1. **JSON Configuration File** (preferred): `~/.second-opinion.json` in your home directory
2. **Environment Variables**: Using `.env` file or system environment variables

### JSON Configuration (Recommended)

Create a `.second-opinion.json` file in your home directory:

```json
{
  "default_provider": "openai",
  "temperature": 0.3,
  "max_tokens": 4096,
  "server_name": "Second Opinion 🔍",
  "server_version": "1.0.0",
  "openai": {
    "api_key": "sk-your-openai-api-key",
    "model": "gpt-4o-mini"
  },
  "google": {
    "api_key": "your-google-api-key",
    "model": "gemini-2.0-flash-exp"
  },
  "ollama": {
    "endpoint": "http://localhost:11434",
    "model": "devstral:latest"
  },
  "mistral": {
    "api_key": "your-mistral-api-key",
    "model": "mistral-small-latest"
  },
  "memory": {
    "max_diff_size_mb": 10,
    "max_file_count": 1000,
    "max_line_length": 1000,
    "enable_streaming": true,
    "chunk_size_mb": 1
  }
}
```

**🚀 Smart Optimization Features:**
- **Dynamic Token Allocation**: Automatically adjusts tokens (4096-32768) based on diff size
- **Task-Specific Temperature**: Optimizes temperature (0.1-0.3) based on analysis type
- **Provider Optimization**: Custom parameters for each LLM provider
- **Memory Management**: Automatic chunking for large diffs and high file counts

### Environment Variables Configuration

If no JSON configuration is found, the server falls back to environment variables:

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit `.env` and configure your LLM providers:

```env
# Set your default provider
DEFAULT_PROVIDER=openai  # or google, ollama, mistral

# Configure each provider with its own API key and preferred model
OPENAI_API_KEY=sk-your-openai-api-key
OPENAI_MODEL=gpt-4o-mini  # or gpt-4o, gpt-4-turbo, gpt-3.5-turbo

GOOGLE_API_KEY=your-google-api-key
GOOGLE_MODEL=gemini-2.0-flash-exp  # or gemini-1.5-flash, gemini-1.5-pro

OLLAMA_ENDPOINT=http://localhost:11434
OLLAMA_MODEL=devstral:latest  # or llama3.2, codellama, mistral, etc.

MISTRAL_API_KEY=your-mistral-api-key
MISTRAL_MODEL=mistral-small-latest  # or mistral-large-latest, codestral-latest

# Global settings apply to all providers
LLM_TEMPERATURE=0.3  # Controls randomness (0.0-2.0, default: 0.3)
LLM_MAX_TOKENS=4096  # Maximum response length (default: 4096)
```

## Setting up with Claude Code

### 1. Locate Claude Code Configuration

The configuration file location depends on your operating system:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

### 2. Edit Configuration

Open the configuration file and add the Second Opinion server:

**Option 1: Using JSON Configuration (Recommended)**
```json
{
  "mcpServers": {
    "second-opinion": {
      "command": "/path/to/second-opinion/bin/second-opinion"
    }
  }
}
```

Replace `/path/to/second-opinion` with the actual path where you cloned the repository.

**Option 2: Using Environment Variables**
```json
{
  "mcpServers": {
    "second-opinion": {
      "command": "/path/to/second-opinion/bin/second-opinion",
      "env": {
        "DEFAULT_PROVIDER": "openai",
        "OPENAI_API_KEY": "your-openai-api-key",
        "OPENAI_MODEL": "gpt-4o-mini",
        "LLM_TEMPERATURE": "0.3",
        "LLM_MAX_TOKENS": "4096"
      }
    }
  }
}
```

### 3. Restart Claude Code

After saving the configuration, restart Claude Code for the changes to take effect.

### 4. Verify Installation

In Claude Code, you should see "second-opinion" in the MCP servers list. You can test it by asking:

```
"What git repository information can you get from the current directory?"
```

## Available Tools

### 1. `analyze_git_diff` 🚀 **Optimized**
Analyzes git diff output to understand code changes using the configured LLM with automatic optimization.

**Parameters:**
- `diff_content` (required): Git diff output to analyze
- `summarize` (optional): Whether to provide a summary of changes
- `provider` (optional): LLM provider to use (overrides default)
- `model` (optional): Model to use (overrides provider default)

**Smart Optimizations:**
- **Dynamic Token Allocation**: 4096-32768 tokens based on diff size
- **Temperature Tuning**: 0.25 optimized for diff analysis
- **Chunking**: Automatic chunking for large diffs (>10MB or >1000 files)
- **Provider-Specific**: Custom parameters per LLM provider

**Example in Claude Code:**
```
"Analyze this git diff and tell me what changed: [paste diff here]"
```

### 2. `review_code` 🚀 **Optimized**
Reviews code for quality, security, and best practices using the configured LLM with task-specific optimization.

**Parameters:**
- `code` (required): Code to review
- `language` (optional): Programming language of the code
- `focus` (optional): Specific focus area - `security`, `performance`, `style`, or `all`
- `provider` (optional): LLM provider to use (overrides default)
- `model` (optional): Model to use (overrides provider default)

**Smart Optimizations:**
- **Task-Specific Temperature**: 0.1 for security focus (high precision), 0.2 for general code review
- **Dynamic Token Allocation**: Scales with code size for comprehensive analysis
- **Focus-Aware Analysis**: Specialized prompts and parameters per focus area

**Example in Claude Code:**
```
"Review this Python code for security issues: [paste code here]"
```

### 3. `analyze_commit` 🚀 **Optimized**
Analyzes a git commit for quality and adherence to best practices using the configured LLM with commit-specific optimization.

**Parameters:**
- `commit_sha` (optional): Git commit SHA to analyze (default: HEAD)
- `repo_path` (optional): Path to the git repository (default: current directory)
- `provider` (optional): LLM provider to use (overrides default)
- `model` (optional): Model to use (overrides provider default)

**Smart Optimizations:**
- **Commit Analysis Temperature**: 0.2 for consistent, deterministic commit analysis
- **Memory-Safe Diff Processing**: Handles large commits with automatic truncation
- **Combined Analysis**: Includes commit message quality, diff analysis, and best practices

**Example in Claude Code:**
```
"Analyze the latest commit in this repository"
"Analyze commit abc123 and tell me if it follows best practices"
```

### 4. `analyze_uncommitted_work` 🚀 **Optimized**
Analyzes uncommitted changes in a git repository to help prepare for commits with intelligent optimization.

**Parameters:**
- `repo_path` (optional): Path to the git repository (default: current directory)
- `staged_only` (optional): Analyze only staged changes (default: false, analyzes all uncommitted changes)
- `provider` (optional): LLM provider to use (overrides default)
- `model` (optional): Model to use (overrides provider default)

**Smart Optimizations:**
- **Code Review Temperature**: 0.2 for balanced analysis of uncommitted changes
- **Large Changeset Handling**: Automatic chunking for extensive modifications
- **Context-Aware Analysis**: Tailored analysis for staged vs. all uncommitted work

**LLM Analysis Includes:**
- Summary of all changes (files modified, added, deleted)
- Type and nature of changes (feature, bugfix, refactor, etc.)
- Completeness and readiness for commit
- Potential issues or concerns
- Suggested commit message(s) if changes are ready
- Recommendations for organizing commits if changes should be split

**Example in Claude Code:**
```
"Analyze my uncommitted changes and suggest a commit message"
"Review only my staged changes before I commit"
"Should I split my current changes into multiple commits?"
```

### 5. `get_repo_info`
Gets information about a git repository (no LLM analysis).

**Parameters:**
- `repo_path` (optional): Path to the git repository (default: current directory)

**Example in Claude Code:**
```
"Show me information about this git repository"
```

## Security Features

- **Input Validation**: All repository paths and commit SHAs are validated to prevent command injection
- **Path Restrictions**: Repository paths must be within the current working directory
- **API Key Protection**: API keys are never exposed in error messages or logs
- **HTTP Timeouts**: All LLM API calls have 30-second timeouts to prevent hanging
- **Concurrent Access**: Thread-safe provider management for concurrent requests

## Optimization System 🚀

Second Opinion includes a comprehensive optimization system that automatically tunes performance based on content and context:

### Dynamic Token Allocation
- **4096 tokens**: Very small diffs (<5KB)
- **6144 tokens**: Small diffs (5-20KB)  
- **8192 tokens**: Medium diffs (20-50KB)
- **12288 tokens**: Large diffs (50-150KB)
- **16384 tokens**: Very large diffs (150-500KB)
- **32768 tokens**: Huge diffs (>500KB)

### Task-Specific Temperature Settings
- **0.1**: Security reviews (maximum precision)
- **0.2**: Code reviews and commit analysis (mostly deterministic)
- **0.25**: Diff analysis (slightly flexible)
- **0.3**: Architecture reviews (allows creativity)

### Provider-Specific Optimizations
- **OpenAI**: Full token allocation with top_p=0.9
- **Google**: Capped at 8192 tokens with focused sampling (top_k=20, top_p=0.8)
- **Mistral**: Conservative allocation with top_p=0.8
- **Ollama**: Local model optimization with repeat_penalty=1.05

### Memory Management
- **Automatic Chunking**: Large diffs (>10MB or >1000 files) are intelligently split
- **Smart Chunk Sizing**: Adapts chunk size based on file count
- **Memory-Aware Streaming**: Enables streaming for large operations

## Development

### Project Structure
```
second-opinion/
├── main.go              # MCP server setup and tool registration
├── handlers.go          # Tool handler implementations
├── validation.go        # Input validation functions
├── config/              # Configuration loading and optimization
│   ├── config.go        # Main configuration with optimization methods
│   └── optimization_test.go # Comprehensive optimization tests
├── llm/                 # LLM provider implementations
│   ├── provider.go      # Provider interface, prompts, and optimization wrapper
│   ├── openai.go        # OpenAI implementation
│   ├── google.go        # Google Gemini implementation
│   ├── ollama.go        # Ollama implementation with advanced options
│   └── mistral.go       # Mistral implementation with additional parameters
├── CLAUDE.md           # Claude Code specific instructions
└── TODO.md             # Development roadmap
```

### Running Tests
```bash
# Run all tests
go test ./... -v

# Run optimization tests specifically
go test ./config -v

# Run specific test suites
go test ./llm -v -run TestProviderConnections

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./...
```

### Linting
```bash
# Install golangci-lint if not already installed
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Auto-fix issues where possible
golangci-lint run --fix
```

### Building
```bash
# Build for current platform
go build -o bin/second-opinion

# Build with race detector (for development)
go build -race -o bin/second-opinion

# Build for different platforms
GOOS=darwin GOARCH=amd64 go build -o bin/second-opinion-darwin-amd64
GOOS=linux GOARCH=amd64 go build -o bin/second-opinion-linux-amd64
GOOS=windows GOARCH=amd64 go build -o bin/second-opinion-windows-amd64.exe
```

## Troubleshooting

### Common Issues

1. **"Provider not configured" error**
   - Ensure you have set up either `~/.second-opinion.json` or environment variables
   - Check that API keys are valid and have appropriate permissions

2. **"Not a git repository" error**
   - Ensure you're running the tool in a directory with a `.git` folder
   - The tool validates that paths are git repositories for security

3. **Timeout errors**
   - Check your internet connection
   - For Ollama, ensure the local server is running: `ollama serve`
   - Consider using a faster model if timeouts persist

4. **Permission denied errors**
   - The tool only allows access to the current working directory and subdirectories
   - Ensure the binary has execute permissions: `chmod +x bin/second-opinion`

### Debug Mode

To see detailed logs, you can run the server directly:
```bash
./bin/second-opinion 2>debug.log
```

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Ensure all tests pass and linting is clean
4. Submit a pull request

See [TODO.md](TODO.md) for planned features and known issues.

## Memory Usage

For large repositories, see [docs/MEMORY_USAGE.md](docs/MEMORY_USAGE.md) for configuration options to handle large diffs efficiently.

## License

MIT