.PHONY: all build test clean install run lint fmt vet deps test-quick test-providers test-models test-integration help

# Variables
BINARY_NAME=second-opinion
BINARY_PATH=./bin/$(BINARY_NAME)
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")
MAIN_PACKAGE=.
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: clean install

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk '/^##/ { \
		helpMessage = substr($$0, 4); \
		getline; \
		if (match($$0, /^[a-zA-Z_-]+:/)) { \
			helpCommand = substr($$0, 0, index($$0, ":")-1); \
			printf "  %-20s %s\n", helpCommand, helpMessage; \
		} \
	}' $(MAKEFILE_LIST)

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build $(LDFLAGS) -o $(BINARY_PATH) $(MAIN_PACKAGE)
	@echo "✅ Build complete: $(BINARY_PATH)"

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) $(MAIN_PACKAGE)
	@echo "✅ Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)"

## run: Run the MCP server
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BINARY_PATH)

## test: Run all tests (shows only failures and coverage)
test:
	@./scripts/test-summary.sh all

## test-quick: Run quick tests (no API calls)
test-quick:
	@./scripts/test-summary.sh quick

## test-providers: Test provider connections
test-providers:
	@echo "Testing provider connections..."
	@go test ./llm -run "TestProviderConnections" -timeout 30s -coverprofile=coverage.out 2>&1 | \
		grep -E "(FAIL:|Error:|panic:|--- FAIL|PASS.*coverage)" || echo "✅ All tests passed"
	@go tool cover -func=coverage.out 2>/dev/null | grep total | awk '{print "Coverage: " $$3}' || true

## test-models: Test model variants
test-models:
	@echo "Testing model variants..."
	@go test ./llm -run "TestProviderModels" -timeout 60s -coverprofile=coverage.out 2>&1 | \
		grep -E "(FAIL:|Error:|panic:|--- FAIL|PASS.*coverage)" || echo "✅ All tests passed"
	@go tool cover -func=coverage.out 2>/dev/null | grep total | awk '{print "Coverage: " $$3}' || true

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	@go test . -timeout 120s -coverprofile=coverage.out 2>&1 | \
		grep -E "(FAIL:|Error:|panic:|--- FAIL|PASS.*coverage)" || echo "✅ All tests passed"
	@go tool cover -func=coverage.out 2>/dev/null | grep total | awk '{print "Coverage: " $$3}' || true

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Install with:"; \
		echo "  brew install golangci-lint"; \
		echo "  or"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✅ Vet complete"

## deps: Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated"

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "✅ All checks passed"

## env-example: Create example .env file
env-example:
	@echo "Creating .env.example..."
	@cp .env .env.example 2>/dev/null || true
	@sed -i.bak 's/sk-proj-.*/your_openai_api_key_here/' .env.example
	@sed -i.bak 's/AIzaSy.*/your_google_api_key_here/' .env.example
	@sed -i.bak 's/dkSTj.*/your_mistral_api_key_here/' .env.example
	@rm -f .env.example.bak
	@echo "✅ Created .env.example"

## mcp-config: Generate MCP configuration example
mcp-config:
	@echo "Creating mcp-config.json example..."
	@echo '{' > mcp-config.json
	@echo '  "mcpServers": {' >> mcp-config.json
	@echo '    "second-opinion": {' >> mcp-config.json
	@echo '      "command": "'$$HOME'/go/bin/second-opinion",' >> mcp-config.json
	@echo '      "env": {' >> mcp-config.json
	@echo '        "DEFAULT_PROVIDER": "openai",' >> mcp-config.json
	@echo '        "OPENAI_API_KEY": "your-api-key-here",' >> mcp-config.json
	@echo '        "OPENAI_MODEL": "gpt-4o-mini"' >> mcp-config.json
	@echo '      }' >> mcp-config.json
	@echo '    }' >> mcp-config.json
	@echo '  }' >> mcp-config.json
	@echo '}' >> mcp-config.json
	@echo "✅ Created mcp-config.json"
	@echo ""
	@echo "Add this configuration to your Claude Desktop config:"
	@echo "  macOS: ~/Library/Application Support/Claude/config.json"
	@echo "  Windows: %APPDATA%\\Claude\\config.json"

## release: Create a new release build
release: clean
	@echo "Building release binaries..."
	@mkdir -p bin/releases
	# macOS AMD64
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/releases/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	# macOS ARM64
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/releases/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	# Linux AMD64
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/releases/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	# Windows AMD64
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/releases/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "✅ Release binaries created in bin/releases/"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

## mod-update: Update all dependencies to latest versions
mod-update:
	@echo "Updating dependencies to latest versions..."
	@go get -u ./...
	@go mod tidy
	@echo "✅ Dependencies updated to latest versions"

## verify: Verify dependencies
verify:
	@echo "Verifying dependencies..."
	@go mod verify
	@echo "✅ Dependencies verified"

# CI/CD targets
## ci: Run CI pipeline (used by GitHub Actions)
ci: deps vet lint test-coverage
	@echo "✅ CI pipeline complete"
