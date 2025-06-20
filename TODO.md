# TODO - Second Opinion MCP Server

## Recently Completed ✅

### High Priority Security & Stability Fixes (Completed)
1. **Resource Leak - File Handle** - Added `defer f.Close()` in `config/config.go`
2. **Race Condition - Global Map Access** - Added `sync.RWMutex` to protect `llmProviders` map
3. **Missing HTTP Timeouts** - Added 30-second timeouts to all LLM provider HTTP clients
4. **Command Injection Prevention** - Added `validation.go` with secure path and SHA validation
5. **API Key Security** - Google provider now redacts sensitive info from error messages
6. **HTTP Response Body Handling** - Added proper body draining before closing connections
7. **Config Type Assignment** - Fixed incorrect ".env" assignment to ".second-opinion.json"
8. **Temperature Zero Override** - Fixed OpenAI provider to respect explicit temperature=0

## Outstanding Issues

### Error Handling Issues (High Priority)

#### 1. Ignored Git Command Errors
- **File**: `handlers.go`, Lines 191-203
- **Issue**: All git command errors are ignored in `getRepoInfo()`
- **Fix**: Add proper error handling and return meaningful messages

#### 2. Incomplete Error Handling in Commit Diff
- **File**: `handlers.go`, Lines 173-179
- **Issue**: Second command error ignored when getting commit diff
- **Fix**: Handle both command failures properly

#### 3. Ignored Error in Staged Changes Fallback
- **File**: `handlers.go`, Line 314
- **Issue**: Error ignored when getting staged changes as fallback
- **Fix**: Log or handle the error appropriately

### Performance Issues (Medium Priority)

#### 4. Memory Usage with Large Diffs
- **File**: `handlers.go`, Line 181
- **Issue**: Large diffs loaded entirely into memory
- **Fix**: Consider streaming large diffs or adding size limits

#### 5. Missing Context Cancellation
- **File**: All `exec.Command` calls
- **Issue**: Commands don't respect context cancellation
- **Fix**: Use `exec.CommandContext` instead of `exec.Command`

### Code Quality Issues (Medium Priority)

#### 6. Missing Input Validation for Config
- **File**: `config/config.go`, Lines 95-109
- **Issue**: No validation for Temperature and MaxTokens ranges
- **Fix**: Add bounds checking (e.g., Temperature: 0.0-2.0, MaxTokens: positive integer)

## Feature Enhancements

### Reliability Improvements

#### 7. Add Retry Logic
- **Files**: All LLM provider files
- **Description**: Implement exponential backoff for transient API failures
- **Priority**: Medium

#### 8. Add Rate Limiting
- **Description**: Implement rate limiting per provider to avoid API limits
- **Priority**: Medium

#### 9. Add Caching
- **Description**: Cache analysis results for repeated queries
- **Priority**: Low

### Observability

#### 10. Add Metrics and Monitoring
- **Description**: Add prometheus metrics for API usage, errors, and performance
- **Priority**: Medium

#### 11. Structured Logging
- **Description**: Replace log.Printf with structured logging (e.g., slog)
- **Priority**: Low

### Documentation

#### 12. API Documentation
- Create comprehensive API documentation for all tools
- Document tool parameters and return values
- Add usage examples

#### 13. Integration Examples
- Add example configurations for each LLM provider
- Create sample scripts showing tool usage
- Document common workflows

### Testing

#### 14. Integration Tests
- Test actual git operations with test repositories
- Test LLM provider integrations with mock servers
- Add tests for error scenarios

#### 15. Benchmarks
- Benchmark large diff handling
- Benchmark concurrent provider access
- Memory usage profiling

## New Features

### 16. Additional Analysis Tools
- `analyze_pull_request` - Analyze PR changes and provide feedback
- `suggest_commit_message` - Generate commit messages from staged changes
- `analyze_branch_diff` - Compare branches and summarize differences

### 17. Configuration Improvements
- Support for `.second-opinion.yaml` config format
- Per-project configuration overrides
- Environment-specific provider selection

### 18. Enhanced Security
- Support for encrypted API keys in config
- Audit logging for all operations
- Sandboxed git command execution

---

## Next Priority Order

1. **Fix remaining error handling issues** (prevents silent failures)
2. **Add context cancellation** (improves responsiveness)
3. **Implement memory limits for large diffs** (prevents OOM)
4. **Add config validation** (prevents invalid configurations)
5. **Add retry logic** (improves reliability)
6. **Create documentation** (improves usability)
7. **Add integration tests** (ensures quality)
8. **Implement new features** (extends functionality)