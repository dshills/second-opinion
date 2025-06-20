# TODO - Second Opinion MCP Server

## Recently Completed ✅

### Phase 1: High Priority Security & Stability Fixes (Completed)
1. **Resource Leak - File Handle** - Added `defer f.Close()` in `config/config.go`
2. **Race Condition - Global Map Access** - Added `sync.RWMutex` to protect `llmProviders` map
3. **Missing HTTP Timeouts** - Added 30-second timeouts to all LLM provider HTTP clients
4. **Command Injection Prevention** - Added `validation.go` with secure path and SHA validation
5. **API Key Security** - Google provider now redacts sensitive info from error messages
6. **HTTP Response Body Handling** - Added proper body draining before closing connections
7. **Config Type Assignment** - Fixed incorrect ".env" assignment to ".second-opinion.json"
8. **Temperature Zero Override** - Fixed OpenAI provider to respect explicit temperature=0

### Phase 2: Error Handling & Context Support (Completed)
9. **Fixed Ignored Git Command Errors** - `getRepoInfo` now shows warnings for failures
10. **Fixed Incomplete Error Handling** - `getCommitInfo` properly handles both command failures
11. **Fixed Staged Changes Fallback** - Errors are now logged appropriately
12. **Added Context Cancellation** - All git commands now use `exec.CommandContext`
13. **Added Error Handling Tests** - Created `handlers_test.go` with comprehensive tests

### Phase 3: Memory Optimization (Completed)
14. **Implemented Memory-Safe Diff Handling** - Created `memory.go` with streaming support
15. **Added Size Limits** - Pre-flight checks prevent loading oversized diffs
16. **Line Truncation** - Long lines are truncated to prevent memory issues
17. **Configuration Support** - Added `MemoryConfig` in config with JSON and env var support
18. **Comprehensive Tests** - Created `memory_test.go` with truncation and streaming tests
19. **Documentation** - Created `docs/MEMORY_USAGE.md` with guidelines and best practices

### Phase 4: Reliability & Performance (Completed)
20. **Retry Logic with Exponential Backoff** - Created `retry.go` with smart retry mechanism
21. **Rate Limit Handling** - Automatic retry for 429 errors with backoff
22. **Connection Pooling** - Created `http_client.go` with SharedHTTPClient
23. **HTTP/2 Support** - Enabled for better performance across all providers
24. **Network Error Resilience** - Retries on transient network failures
25. **Comprehensive Retry Tests** - 27 test cases covering all retry scenarios

## Outstanding Issues

### Performance Issues (High Priority)

#### 1. Add Progress Indicators
- **Issue**: Long-running LLM calls have no feedback
- **Fix**: Add progress callbacks or status updates
- **Impact**: Better user experience

### Code Quality Issues (Medium Priority)

#### 3. Missing Input Validation for Config
- **File**: `config/config.go`, Lines 95-109
- **Issue**: No validation for Temperature and MaxTokens ranges
- **Fix**: Add bounds checking (e.g., Temperature: 0.0-2.0, MaxTokens: 1-100000)

#### 4. Improve Error Messages
- **Issue**: Some errors lack context (which file, which operation)
- **Fix**: Wrap errors with more descriptive messages
- **Example**: Instead of "permission denied", show "cannot read /path/to/file: permission denied"

## Feature Enhancements

### Reliability Improvements

#### 5. Add Request/Response Logging
- **Description**: Optional debug logging for API calls
- **Implementation**: Log requests/responses when DEBUG env var is set
- **Priority**: Medium

### Performance Optimizations

#### 6. Implement Caching Layer
- **Description**: Cache analysis results with TTL
- **Use cases**: 
  - Same diff analyzed multiple times
  - Repeated code reviews
- **Priority**: Low

### Observability

#### 7. Add Metrics Collection
- **Metrics to track**:
  - API call latency by provider
  - Error rates by type
  - Token usage per request
- **Implementation**: Prometheus metrics or StatsD
- **Priority**: Medium

#### 8. Structured Logging
- **Current**: Using `log.Printf`
- **Target**: Use `slog` for structured JSON logs
- **Benefits**: Better log aggregation and searching
- **Priority**: Low

### Testing Improvements

#### 9. Add Integration Test Suite
- **Coverage needed**:
  - Git operations with test repos
  - Mock LLM providers for predictable tests
  - Error injection tests
  - Concurrent request handling

#### 10. Add Benchmarks
- **Areas to benchmark**:
  - Large diff processing
  - Concurrent provider access
  - Memory usage under load

#### 11. Add Fuzz Testing
- **Target areas**:
  - Input validation functions
  - Git command construction
  - Path handling

## New Features

### 12. Additional Analysis Tools
- **`analyze_pull_request`** - Full PR analysis with file-by-file breakdown
- **`suggest_commit_message`** - AI-generated commit messages from changes
- **`analyze_branch_diff`** - Compare and summarize branch differences
- **`code_smell_detection`** - Identify problematic patterns

### 13. Enhanced Configuration
- **YAML Config Support** - Alternative to JSON
- **Project-Specific Config** - `.second-opinion.yml` in repo root
- **Provider Profiles** - Quick switching between configurations
- **Config Validation CLI** - `second-opinion validate-config`

### 14. Security Enhancements
- **Encrypted API Keys** - Store keys encrypted at rest
- **Audit Logging** - Track all operations with user/timestamp
- **Sandboxed Execution** - Run git commands in restricted environment
- **Secret Scanning** - Prevent accidental secret commits

### 15. Developer Experience
- **CLI Mode** - Direct command-line usage without MCP
- **Web UI** - Simple web interface for testing
- **Plugin System** - Extensible analysis modules
- **Custom Prompts** - User-defined analysis templates

## Architecture Improvements

### 16. Modular Provider System
- **Current**: Providers hardcoded in main package
- **Target**: Plugin-based provider loading
- **Benefits**: Easier to add new providers

### 17. Streaming Support
- **Current**: All responses buffered
- **Target**: Stream LLM responses as they arrive
- **Benefits**: Faster perceived performance

### 18. Multi-Model Support
- **Feature**: Use different models for different tasks
- **Example**: Fast model for diffs, powerful model for security review

---

## Next Sprint Priority

### Sprint 1: Performance & Reliability (Completed ✅)
1. ✅ Fix all error handling issues
2. ✅ Add context cancellation support
3. ✅ Memory limits for large diffs
4. ✅ Add retry logic with backoff
5. ✅ Connection pooling for HTTP clients

### Sprint 2: Developer Experience (Current Focus)
1. **Config validation and defaults**
2. **Better error messages with context**
3. **Progress indicators for long operations**
4. **Debug logging mode**

### Sprint 3: New Features
1. **PR analysis tool**
2. **Commit message generator**
3. **Branch diff analyzer**

### Sprint 4: Production Readiness
1. **Metrics and monitoring**
2. **Comprehensive test suite**
3. **Performance benchmarks**
4. **Documentation updates**