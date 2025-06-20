#!/bin/bash

# Test Ollama connectivity and configuration
# Usage: ./scripts/test-ollama.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Load environment variables if not already set
if [ -z "$OLLAMA_ENDPOINT" ] && [ -f .env ]; then
    OLLAMA_ENDPOINT=$(grep '^OLLAMA_ENDPOINT=' .env | cut -d= -f2-)
    OLLAMA_MODEL=$(grep '^OLLAMA_MODEL=' .env | cut -d= -f2-)
fi

ENDPOINT="${OLLAMA_ENDPOINT:-http://localhost:11434}"
MODEL="${OLLAMA_MODEL:-llama3.2}"

echo "🔍 Testing Ollama Configuration"
echo "================================"
echo "Endpoint: $ENDPOINT"
echo "Model: $MODEL"
echo ""

# Test 1: Basic connectivity
echo -n "1. Testing basic connectivity... "
if curl -s -f "$ENDPOINT" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Connected${NC}"
else
    echo -e "${RED}✗ Failed to connect${NC}"
    echo "   Make sure Ollama is running at $ENDPOINT"
    exit 1
fi

# Test 2: API endpoint
echo -n "2. Testing API endpoint... "
if curl -s -f "$ENDPOINT/api/tags" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ API accessible${NC}"
else
    echo -e "${RED}✗ API not accessible${NC}"
    exit 1
fi

# Test 3: List available models
echo -e "\n3. Available models:"
MODELS=$(curl -s "$ENDPOINT/api/tags" | jq -r '.models[].name' 2>/dev/null || echo "Failed to parse")
if [ "$MODELS" = "Failed to parse" ]; then
    echo -e "${RED}   Failed to list models${NC}"
else
    echo "$MODELS" | sed 's/^/   - /'
    
    # Check if configured model exists
    if echo "$MODELS" | grep -q "^$MODEL"; then
        echo -e "\n   ${GREEN}✓ Configured model '$MODEL' is available${NC}"
    else
        echo -e "\n   ${YELLOW}⚠ Configured model '$MODEL' not found${NC}"
        echo "   Available models that might work:"
        echo "$MODELS" | grep -i "${MODEL%%-*}" | sed 's/^/   - /' || echo "   - No similar models found"
    fi
fi

# Test 4: Simple generation test
echo -e "\n4. Testing generation with $MODEL..."
PROMPT="What is 2+2? Reply with just the number."
REQUEST_JSON=$(cat <<EOF
{
    "model": "$MODEL",
    "prompt": "$PROMPT",
    "stream": false,
    "options": {
        "temperature": 0.1
    }
}
EOF
)

echo "   Sending test prompt: \"$PROMPT\""
START_TIME=$(date +%s)

RESPONSE=$(curl -s -X POST "$ENDPOINT/api/generate" \
    -H "Content-Type: application/json" \
    -d "$REQUEST_JSON" \
    --max-time 30 2>&1)

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

if [ $? -eq 0 ]; then
    # Try to parse the response
    if echo "$RESPONSE" | jq -e '.response' > /dev/null 2>&1; then
        ANSWER=$(echo "$RESPONSE" | jq -r '.response')
        echo -e "   ${GREEN}✓ Response received in ${DURATION}s: $ANSWER${NC}"
    elif echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        ERROR=$(echo "$RESPONSE" | jq -r '.error')
        echo -e "   ${RED}✗ Error: $ERROR${NC}"
    else
        echo -e "   ${RED}✗ Invalid response format${NC}"
        echo "   Raw response: $RESPONSE"
    fi
else
    echo -e "   ${RED}✗ Request failed or timed out after 30s${NC}"
    echo "   Error: $RESPONSE"
fi

# Test 5: Run Go tests
echo -e "\n5. Running Ollama-specific Go tests..."
cd "$(dirname "$0")/.."

# Run just the connectivity and basic tests
echo "   Running connectivity tests..."
if go test -v -timeout 10s ./llm -run "TestOllamaEndpointConnectivity|TestOllamaModelAvailability" 2>&1 | grep -E "(PASS|FAIL|SKIP)"; then
    echo -e "   ${GREEN}✓ Connectivity tests completed${NC}"
else
    echo -e "   ${YELLOW}⚠ Some tests may have failed${NC}"
fi

echo -e "\n📊 Summary"
echo "=========="
echo "Endpoint: $ENDPOINT"
echo "Model: $MODEL"
echo ""
echo "Run the full test suite with:"
echo "  go test -v ./llm -run TestOllama"
echo ""
echo "For integration tests with real Ollama:"
echo "  go test -v ./llm -run TestOllamaRealIntegration"