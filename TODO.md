# TODO - Second Opinion MCP Server

## Critical Issues (High Priority)

### 1. Resource Leak - File Handle Not Closed
- **File**: `config/config.go`, Line 61
- **Issue**: File handle `f` is never closed when reading config
- **Fix**: Add `defer f.Close()` after `os.Open()`

### 2. Race Condition - Global Map Access
- **File**: `main.go`, Line 15
- **Issue**: Global `llmProviders` map accessed by multiple goroutines without synchronization
- **Fix**: Use `sync.RWMutex` or `sync.Map` for thread-safe access

### 3. Missing HTTP Timeouts
- **Files**: All files in `llm/` directory (openai.go, google.go, ollama.go, mistral.go)
- **Issue**: HTTP clients have no timeout configuration
- **Fix**: Add timeout to HTTP client initialization

## Security Issues (High Priority)

### 4. Potential Command Injection
- **File**: `handlers.go`, multiple locations
- **Issue**: User-provided `repoPath` and `commitSHA` used in exec.Command without validation
- **Fix**: Validate inputs and use `filepath.Clean()` on paths

### 5. API Key Exposure Risk
- **File**: `llm/google.go`, Line 51
- **Issue**: API keys included in URLs could be exposed in logs
- **Fix**: Ensure keys are not logged in error messages

## Error Handling Issues (Medium Priority)

### 6. Ignored Git Command Errors
- **File**: `handlers.go`, Lines 191-203
- **Issue**: All git command errors are ignored in `getRepoInfo()`
- **Fix**: Add proper error handling and return meaningful messages

### 7. Incomplete Error Handling in Commit Diff
- **File**: `handlers.go`, Lines 173-179
- **Issue**: Second command error ignored when getting commit diff
- **Fix**: Handle both command failures properly

### 8. Ignored Error in Staged Changes Fallback
- **File**: `handlers.go`, Line 314
- **Issue**: Error ignored when getting staged changes as fallback
- **Fix**: Log or handle the error appropriately

## Performance Issues (Medium Priority)

### 9. Memory Usage with Large Diffs
- **File**: `handlers.go`, Line 181
- **Issue**: Large diffs loaded entirely into memory
- **Fix**: Consider streaming large diffs or adding size limits

### 10. Missing Context Cancellation
- **File**: All `exec.Command` calls
- **Issue**: Commands don't respect context cancellation
- **Fix**: Use `exec.CommandContext` instead of `exec.Command`

## Code Quality Issues (Low Priority)

### 11. HTTP Response Body Handling
- **Files**: All LLM provider files
- **Issue**: Response bodies not drained before closing
- **Fix**: Add `io.Copy(io.Discard, resp.Body)` before close

### 12. Incorrect Config Type Assignment
- **File**: `config/config.go`, Line 62
- **Issue**: ConfigType set to ".env" instead of ".second-opinion.json"
- **Fix**: Remove or correct the assignment

### 13. Temperature Zero Override
- **File**: `llm/openai.go`, Lines 35-38
- **Issue**: Valid temperature of 0 gets overridden to 0.3
- **Fix**: Only set default if temperature is not explicitly set

### 14. Missing Input Validation
- **File**: `config/config.go`, Lines 95-109
- **Issue**: No validation for Temperature and MaxTokens ranges
- **Fix**: Add bounds checking for configuration values

## Feature Enhancements

### 15. Add Retry Logic
- **Files**: All LLM provider files
- **Issue**: No retry logic for transient failures
- **Fix**: Implement exponential backoff for API calls

### 16. Add Metrics and Monitoring
- **Issue**: No metrics for API usage, errors, or performance
- **Fix**: Add prometheus metrics or logging

### 17. Add Rate Limiting
- **Issue**: No rate limiting for API calls
- **Fix**: Implement rate limiting per provider

### 18. Add Caching
- **Issue**: No caching of analysis results
- **Fix**: Add optional caching layer for repeated analyses

## Documentation

### 19. Add API Documentation
- Create comprehensive API documentation for all tools

### 20. Add Examples
- Add example usage for each tool
- Create integration examples

## Testing

### 21. Add Integration Tests
- Test actual git operations
- Test LLM provider integrations

### 22. Add Benchmarks
- Benchmark large diff handling
- Benchmark concurrent provider access

---

## Priority Order

1. Fix resource leak (config file handle)
2. Fix race condition (llmProviders map)
3. Add HTTP timeouts
4. Validate command inputs
5. Improve error handling
6. Add context cancellation
7. Optimize memory usage
8. Improve code quality
9. Add features and documentation