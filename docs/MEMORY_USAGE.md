# Memory Usage Guidelines

## Overview

The Second Opinion MCP server implements several memory management features to handle large Git diffs efficiently. This document outlines the configuration options and best practices for managing memory usage.

## Configuration

Memory management can be configured through either `.second-opinion.json` in your home directory or environment variables.

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `max_diff_size_mb` | 10 MB | Maximum size of diffs to process |
| `max_file_count` | 1000 | Maximum number of files in a single diff |
| `max_line_length` | 1000 | Maximum characters per line before truncation |
| `enable_streaming` | true | Enable streaming mode for large diffs |
| `chunk_size_mb` | 1 MB | Size of chunks when streaming |

### JSON Configuration

Add a `memory` section to your `.second-opinion.json`:

```json
{
  "memory": {
    "max_diff_size_mb": 20,
    "max_file_count": 2000,
    "max_line_length": 1500,
    "enable_streaming": true,
    "chunk_size_mb": 2
  }
}
```

### Environment Variables

You can also configure memory settings using environment variables:

- `MAX_DIFF_SIZE_MB` - Maximum diff size in megabytes
- `MAX_FILE_COUNT` - Maximum number of files
- `MAX_LINE_LENGTH` - Maximum line length
- `ENABLE_STREAMING` - Enable streaming ("true" or "1")
- `CHUNK_SIZE_MB` - Chunk size for streaming

Example:
```bash
export MAX_DIFF_SIZE_MB=50
export MAX_FILE_COUNT=5000
export ENABLE_STREAMING=true
```

## Memory Safety Features

### 1. Pre-flight Size Checks

Before loading a diff, the system performs a size estimation using `git diff --numstat`. This prevents loading diffs that would exceed memory limits.

### 2. Streaming Mode

When `enable_streaming` is true, large diffs are processed in chunks rather than loaded entirely into memory. This allows processing of diffs larger than available RAM.

### 3. Line Truncation

Lines longer than `max_line_length` are automatically truncated with "..." appended. This prevents memory issues from files with extremely long lines (e.g., minified JavaScript).

### 4. File Count Limits

Diffs with more files than `max_file_count` are truncated. This prevents memory exhaustion from repositories with thousands of changed files.

## Warning Messages

When limits are exceeded, the system provides clear warning messages:

- **Size limit exceeded**: "Diff truncated at XMB limit"
- **File count exceeded**: "Truncated at X files limit"
- **Pre-flight check failure**: "diff too large: estimated XKB exceeds limit of YKB"

## Performance Considerations

### Small Repositories (< 10MB diffs)

The default settings work well for most small to medium repositories:
- 10MB diff size limit
- 1000 file limit
- Streaming enabled

### Large Repositories (10-100MB diffs)

For larger repositories, consider:
```json
{
  "memory": {
    "max_diff_size_mb": 50,
    "max_file_count": 5000,
    "enable_streaming": true,
    "chunk_size_mb": 5
  }
}
```

### Very Large Repositories (> 100MB diffs)

For very large monorepos:
```json
{
  "memory": {
    "max_diff_size_mb": 200,
    "max_file_count": 10000,
    "max_line_length": 500,
    "enable_streaming": true,
    "chunk_size_mb": 10
  }
}
```

## Troubleshooting

### Out of Memory Errors

If you encounter OOM errors:
1. Reduce `max_diff_size_mb`
2. Ensure `enable_streaming` is true
3. Reduce `chunk_size_mb` for systems with limited RAM

### Truncated Diffs

If important changes are being truncated:
1. Increase `max_diff_size_mb` 
2. Increase `max_file_count`
3. Consider analyzing specific subdirectories instead of the entire repository

### Slow Performance

If diff processing is slow:
1. Increase `chunk_size_mb` (uses more memory but faster)
2. Ensure your system has sufficient RAM for the configured limits
3. Consider using SSDs for better Git performance

## Best Practices

1. **Monitor Memory Usage**: Use system monitoring tools to observe memory usage during large diff operations.

2. **Incremental Analysis**: For very large changes, consider analyzing commits individually rather than entire branch differences.

3. **Optimize Git**: Keep your Git repository optimized with regular garbage collection:
   ```bash
   git gc --aggressive
   ```

4. **Use Shallow Clones**: For CI/CD environments, use shallow clones to reduce repository size.

5. **Configure Per-Project**: Different projects may need different limits. Consider project-specific configurations.

## Examples

### Analyzing a Large PR

When analyzing a pull request with many changes:
```bash
# Set temporary higher limits
export MAX_DIFF_SIZE_MB=100
export MAX_FILE_COUNT=5000

# Use the analyze_commit tool
# The system will automatically handle memory limits
```

### CI/CD Configuration

For automated code review in CI/CD:
```yaml
env:
  MAX_DIFF_SIZE_MB: 50
  ENABLE_STREAMING: true
  CHUNK_SIZE_MB: 5
```

## Technical Details

The memory management system uses:
- **Streaming**: `io.Reader` based processing for large files
- **Buffering**: Configurable buffer sizes for optimal performance
- **Line-by-line processing**: Prevents loading entire files into memory
- **Early termination**: Stops processing when limits are reached

This ensures that even repositories with gigabytes of changes can be analyzed without exhausting system memory.