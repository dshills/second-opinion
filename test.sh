#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_color() {
    echo -e "${2}${1}${NC}"
}

# Function to check if API key is configured
is_configured() {
    local key=$1
    if [ -n "$key" ] && [ "$key" != "your_openai_api_key_here" ] && [ "$key" != "your_google_api_key_here" ] && [ "$key" != "your_mistral_api_key_here" ]; then
        return 0
    fi
    return 1
}

# Load environment variables
if [ -f .env ]; then
    # Export each line from .env that's not a comment and not SERVER_NAME
    while IFS='=' read -r key value; do
        # Skip comments and empty lines
        if [[ ! "$key" =~ ^#.*$ ]] && [[ -n "$key" ]] && [[ "$key" != "SERVER_NAME" ]]; then
            # Remove quotes if present
            value="${value%\"}"
            value="${value#\"}"
            export "$key=$value"
        fi
    done < .env
    # Manually set SERVER_NAME
    export SERVER_NAME="Second Opinion 🔍"
fi

print_color "🧪 Running Second Opinion Tests" "$BLUE"
print_color "================================" "$BLUE"
echo

# Show configuration status
print_color "📋 Configuration Status:" "$YELLOW"
print_color "------------------------" "$YELLOW"

if is_configured "$OPENAI_API_KEY"; then
    print_color "✅ OpenAI: Configured (Model: $OPENAI_MODEL)" "$GREEN"
else
    print_color "❌ OpenAI: Not configured" "$RED"
fi

if is_configured "$GOOGLE_API_KEY"; then
    print_color "✅ Google: Configured (Model: $GOOGLE_MODEL)" "$GREEN"
else
    print_color "❌ Google: Not configured" "$RED"
fi

if [ -n "$OLLAMA_ENDPOINT" ]; then
    print_color "✅ Ollama: Configured (Endpoint: $OLLAMA_ENDPOINT, Model: $OLLAMA_MODEL)" "$GREEN"
else
    print_color "❌ Ollama: Not configured" "$RED"
fi

if is_configured "$MISTRAL_API_KEY"; then
    print_color "✅ Mistral: Configured (Model: $MISTRAL_MODEL)" "$GREEN"
else
    print_color "❌ Mistral: Not configured" "$RED"
fi

echo
print_color "Default Provider: $DEFAULT_PROVIDER" "$YELLOW"
echo

# Determine which tests to run
TEST_MODE=${1:-all}

# Coverage file
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"

# Function to run tests with coverage and minimal output
run_tests() {
    local package=$1
    local test_pattern=$2
    local timeout=$3
    local description=$4
    
    print_color "$description" "$YELLOW"
    
    # Run tests with coverage, only showing failures
    if [ -n "$test_pattern" ]; then
        output=$(go test "$package" -coverprofile="$COVERAGE_FILE.tmp" -run "$test_pattern" -timeout "$timeout" -count=1 2>&1)
    else
        output=$(go test "$package" -coverprofile="$COVERAGE_FILE.tmp" -timeout "$timeout" -count=1 2>&1)
    fi
    
    exit_code=$?
    
    # Check if tests passed or failed
    if [ $exit_code -eq 0 ]; then
        # Extract coverage percentage
        coverage=$(echo "$output" | grep -o 'coverage: [0-9.]*%' | head -1)
        if [ -n "$coverage" ]; then
            print_color "  ✅ PASS - $coverage" "$GREEN"
        else
            print_color "  ✅ PASS" "$GREEN"
        fi
    else
        print_color "  ❌ FAIL" "$RED"
        # Show only the failure output
        echo "$output" | grep -E "(FAIL|Error:|panic:|--- FAIL)" | sed 's/^/    /'
    fi
    
    # Merge coverage files
    if [ -f "$COVERAGE_FILE.tmp" ]; then
        if [ -f "$COVERAGE_FILE" ]; then
            # Merge with existing coverage
            go run github.com/wadey/gocovmerge@latest "$COVERAGE_FILE" "$COVERAGE_FILE.tmp" > "$COVERAGE_FILE.merged" 2>/dev/null
            mv "$COVERAGE_FILE.merged" "$COVERAGE_FILE"
        else
            mv "$COVERAGE_FILE.tmp" "$COVERAGE_FILE"
        fi
        rm -f "$COVERAGE_FILE.tmp"
    fi
    
    return $exit_code
}

# Initialize coverage file
rm -f "$COVERAGE_FILE" "$COVERAGE_FILE.tmp" "$COVERAGE_FILE.merged"

# Track overall test status
OVERALL_STATUS=0

case $TEST_MODE in
    "quick")
        print_color "⚡ Running Quick Tests (no API calls)..." "$BLUE"
        print_color "----------------------------------------" "$BLUE"
        run_tests "./llm" "TestAnalysisPrompts|TestEnvironmentVariables" "10s" "Unit Tests:" || OVERALL_STATUS=1
        ;;
    
    "providers")
        print_color "🔌 Testing Provider Connections..." "$BLUE"
        print_color "----------------------------------" "$BLUE"
        run_tests "./llm" "TestProviderConnections" "30s" "Provider Tests:" || OVERALL_STATUS=1
        ;;
    
    "models")
        print_color "🤖 Testing Model Variants..." "$BLUE"
        print_color "----------------------------" "$BLUE"
        run_tests "./llm" "TestProviderModels" "60s" "Model Tests:" || OVERALL_STATUS=1
        ;;
    
    "integration")
        print_color "🔧 Running Integration Tests..." "$BLUE"
        print_color "-------------------------------" "$BLUE"
        run_tests "." "" "120s" "Integration Tests:" || OVERALL_STATUS=1
        ;;
    
    "all")
        print_color "🎯 Running All Tests..." "$BLUE"
        print_color "-----------------------" "$BLUE"
        echo
        
        run_tests "./llm" "TestAnalysisPrompts|TestEnvironmentVariables" "10s" "1️⃣ Unit Tests:" || OVERALL_STATUS=1
        echo
        
        run_tests "./llm" "TestProviderConnections" "30s" "2️⃣ Provider Connections:" || OVERALL_STATUS=1
        echo
        
        run_tests "./llm" "TestProviderModels" "60s" "3️⃣ Model Variants:" || OVERALL_STATUS=1
        echo
        
        run_tests "." "" "120s" "4️⃣ Integration Tests:" || OVERALL_STATUS=1
        ;;
    
    *)
        print_color "❓ Usage: $0 [quick|providers|models|integration|all]" "$RED"
        echo "  quick       - Run only unit tests (no API calls)"
        echo "  providers   - Test LLM provider connections"
        echo "  models      - Test different model variants"
        echo "  integration - Run integration tests"
        echo "  all         - Run all tests (default)"
        exit 1
        ;;
esac

echo

# Generate coverage report if tests were run
if [ -f "$COVERAGE_FILE" ]; then
    print_color "📊 Coverage Report:" "$YELLOW"
    print_color "-------------------" "$YELLOW"
    
    # Show total coverage
    total_coverage=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}')
    print_color "Total Coverage: $total_coverage" "$GREEN"
    
    # Generate HTML report
    go tool cover -html="$COVERAGE_FILE" -o "$COVERAGE_HTML" 2>/dev/null
    if [ -f "$COVERAGE_HTML" ]; then
        print_color "HTML Report: $COVERAGE_HTML" "$BLUE"
    fi
    echo
fi

# Final status
if [ $OVERALL_STATUS -eq 0 ]; then
    print_color "✅ All tests passed!" "$GREEN"
else
    print_color "❌ Some tests failed!" "$RED"
    exit $OVERALL_STATUS
fi