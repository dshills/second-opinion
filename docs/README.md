# Documentation Index

This directory contains comprehensive documentation for the Second Opinion MCP Server.

## 📚 Available Documentation

### Core Documentation
- [**Main README**](../README.md) - Project overview, installation, and basic usage
- [**CLAUDE.md**](../CLAUDE.md) - Project instructions for Claude Code (development guide)

### User Guides
- [**Memory Usage Guidelines**](MEMORY_USAGE.md) - Memory management and configuration

### Development
- [**Feature Comparison**](development/FEATURE_COMPARISON.md) - Comparison with other code review tools
- [**Roadmap**](development/ROADMAP.md) - Future development plans and vision
- [**TODO List**](development/TODO.md) - Completed tasks and ongoing work

### Troubleshooting
- [**Bug Reports**](troubleshooting/BUG_REPORT.md) - Known issues and security audit
- [**Ollama Diagnostics**](troubleshooting/OLLAMA_DIAGNOSTICS.md) - Ollama-specific setup and issues

## 🔧 Development Resources

### Quick Links
- **Testing**: Use `make test` for full test suite
- **Linting**: Use `make lint` to check code quality
- **Building**: Use `make build` to create binaries
- **Coverage**: Use `make test-coverage` for detailed coverage reports

### Project Structure
```
second-opinion/
├── config/          # Configuration management
├── llm/            # LLM provider implementations
├── scripts/        # Build and test scripts
├── docs/           # Documentation (this directory)
│   ├── development/    # Development-related docs
│   └── troubleshooting/ # Issue resolution guides
└── handlers.go     # Main MCP tool handlers
```

## 📖 Getting Started

1. **New Users**: Start with the [Main README](../README.md)
2. **Developers**: Read [CLAUDE.md](../CLAUDE.md) for development setup
3. **Issues**: Check [troubleshooting docs](troubleshooting/) for common problems
4. **Contributing**: Review the [roadmap](development/ROADMAP.md) and [TODO list](development/TODO.md)