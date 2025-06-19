# Second Opinion 🔍

An MCP (Model Context Protocol) server that assists Claude Code in reviewing commits and code bases. This tool leverages external LLMs (OpenAI, Google Gemini, Ollama, Mistral) to provide intelligent code review capabilities, git diff analysis, and commit quality assessment.

## Features

- **Git Diff Analysis**: Analyze git diff output to understand code changes using LLMs
- **Code Review**: Review code for quality, security, and best practices with AI assistance
- **Commit Analysis**: Analyze git commits for quality and adherence to best practices
- **Repository Information**: Get information about git repositories
- **Multiple LLM Support**: Works with OpenAI, Google Gemini, Ollama (local), and Mistral AI

## Installation

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
    "model": "gemini-1.5-flash"
  },
  "ollama": {
    "endpoint": "http://localhost:11434",
    "model": "llama3.2"
  },
  "mistral": {
    "api_key": "your-mistral-api-key",
    "model": "mistral-small-latest"
  }
}
```

### Environment Variables Configuration

If no JSON configuration is found, the server falls back to environment variables:

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit `.env` and configure your LLM providers:

### Multi-Provider Configuration

The server now supports configuring multiple LLM providers simultaneously. You can set a default provider and switch between providers at runtime.

```env
# Set your default provider
DEFAULT_PROVIDER=openai  # or google, ollama, mistral

# Configure each provider with its own API key and preferred model
OPENAI_API_KEY=sk-your-openai-api-key
OPENAI_MODEL=gpt-4o-mini  # or gpt-4o, gpt-4-turbo, gpt-3.5-turbo

GOOGLE_API_KEY=your-google-api-key
GOOGLE_MODEL=gemini-1.5-flash  # or gemini-1.5-pro, gemini-1.0-pro

OLLAMA_ENDPOINT=http://localhost:11434
OLLAMA_MODEL=llama3.2  # or codellama, mistral, etc.

MISTRAL_API_KEY=your-mistral-api-key
MISTRAL_MODEL=mistral-small-latest  # or mistral-large-latest, codestral-latest

# Global settings apply to all providers
LLM_TEMPERATURE=0.3  # Controls randomness (0.0-2.0)
LLM_MAX_TOKENS=4096  # Maximum response length
```

### Runtime Provider Selection

Each tool supports optional `provider` and `model` parameters to override the defaults:

```
# Use default provider
"analyze_git_diff": {"diff_content": "..."}

# Use specific provider
"analyze_git_diff": {"diff_content": "...", "provider": "google"}

# Use specific provider and model
"analyze_git_diff": {"diff_content": "...", "provider": "openai", "model": "gpt-4o"}
```

## Usage

### Running the Server

The server runs over stdio, making it compatible with MCP clients:

```bash
./bin/second-opinion
```

### Configuring with Claude Desktop

Add the following to your Claude Desktop configuration:

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
This will use the configuration from `~/.second-opinion.json` in your home directory.

**Option 2: Using Environment Variables**
```json
{
  "mcpServers": {
    "second-opinion": {
      "command": "/path/to/second-opinion/bin/second-opinion",
      "env": {
        "DEFAULT_PROVIDER": "openai",
        "OPENAI_API_KEY": "your-openai-api-key",
        "GOOGLE_API_KEY": "your-google-api-key"
      }
    }
  }
}
```

**Option 3: Using .env File**
Place a `.env` file in the second-opinion project directory with all provider configurations, then use the simple configuration in Claude Desktop.

## Available Tools

### 1. `analyze_git_diff`
Analyzes git diff output to understand code changes using the configured LLM.

**Parameters:**
- `diff_content` (required): Git diff output to analyze
- `summarize` (optional): Whether to provide a summary of changes
- `provider` (optional): LLM provider to use (overrides default)
- `model` (optional): Model to use (overrides provider default)

**LLM Analysis Includes:**
- Summary of changes (files changed, lines added/removed)
- Type of change (feature, bugfix, refactor, etc.)
- Potential issues or concerns
- Brief summary (if requested)

### 2. `review_code`
Reviews code for quality, security, and best practices using the configured LLM.

**Parameters:**
- `code` (required): Code to review
- `language` (optional): Programming language of the code
- `focus` (optional): Specific focus area - `security`, `performance`, `style`, or `all`
- `provider` (optional): LLM provider to use (overrides default)
- `model` (optional): Model to use (overrides provider default)

**LLM Review Includes:**
- Security vulnerabilities
- Performance concerns
- Code quality and style issues
- Best practice violations
- Suggestions for improvement

### 3. `analyze_commit`
Analyzes a git commit for quality and adherence to best practices using the configured LLM.

**Parameters:**
- `commit_sha` (optional): Git commit SHA to analyze (default: HEAD)
- `repo_path` (optional): Path to the git repository (default: current directory)
- `provider` (optional): LLM provider to use (overrides default)
- `model` (optional): Model to use (overrides provider default)

**LLM Analysis Includes:**
- Summary of commit changes
- Commit message quality assessment
- Best practices adherence
- Improvement suggestions

### 4. `get_repo_info`
Gets information about a git repository (no LLM analysis).

**Parameters:**
- `repo_path` (optional): Path to the git repository (default: current directory)

## Example Usage in Claude

Once configured, you can use these tools in Claude:

### Basic Usage (uses default provider)
```
"Can you analyze the latest commit in my repository?"
"Review this code for security issues: [paste code]"
"What changes are in this diff: [paste git diff]"
```

### Using Specific Providers
```
"Analyze this commit using Google Gemini"
"Review this code with Ollama's codellama model"
"Use GPT-4o to analyze this diff"
```

The tools will automatically use the provider and model specified in the request, or fall back to your configured defaults.

## Development

### Running Tests

The project includes comprehensive tests for LLM provider connections and tool functionality.

#### Prerequisites
1. Create a `.env` file with your API keys:
```bash
cp .env.example .env
# Edit .env with your actual API keys
```

2. For Ollama tests, ensure Ollama is running locally:
```bash
ollama serve
```

#### Test Commands

Using the test script:
```bash
# Run all tests
./test.sh

# Test provider connections only
./test.sh providers

# Test different models
./test.sh models

# Run integration tests
./test.sh integration

# Run quick tests (no API calls)
./test.sh quick
```

Using Go directly:
```bash
# Run all tests
go test ./... -v

# Run specific test suites
go test ./llm -v -run TestProviderConnections
go test ./llm -v -run TestProviderModels
go test . -v -run TestHandleGitDiff

# Run with timeout
go test ./... -v -timeout 5m
```

#### Test Structure

- **Provider Connection Tests** (`llm/provider_test.go`): Tests connections to each configured LLM provider
- **Model Tests**: Tests different models for providers that support multiple models
- **Integration Tests** (`main_test.go`): Tests the MCP tool handlers with real LLM calls
- **Unit Tests**: Tests individual components without external dependencies

The tests will automatically skip providers that aren't configured in your `.env` file.

### Linting
```bash
golangci-lint run
```

## Contributing

Contributions are welcome! Please ensure all tests pass and the linter is happy before submitting a PR.

## License

MIT