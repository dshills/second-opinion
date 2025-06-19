#!/bin/bash

# Quiet test runner - shows only failures and coverage

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Running tests (failures only)...${NC}"
echo "================================"

# Test mode
TEST_MODE=${1:-all}

# Run tests based on mode
case $TEST_MODE in
    "quick")
        echo -e "${YELLOW}Quick Tests:${NC}"
        go test ./llm -run "TestAnalysisPrompts|TestEnvironmentVariables" -coverprofile=coverage.out 2>&1 | grep -E "FAIL|PASS.*coverage|Error:|panic:" || echo -e "${GREEN}✅ All tests passed${NC}"
        ;;
    
    "providers")
        echo -e "${YELLOW}Provider Tests:${NC}"
        go test ./llm -run "TestProviderConnections" -timeout 30s -coverprofile=coverage.out 2>&1 | grep -E "FAIL|PASS.*coverage|Error:|panic:" || echo -e "${GREEN}✅ All tests passed${NC}"
        ;;
    
    "models")
        echo -e "${YELLOW}Model Tests:${NC}"
        go test ./llm -run "TestProviderModels" -timeout 60s -coverprofile=coverage.out 2>&1 | grep -E "FAIL|PASS.*coverage|Error:|panic:" || echo -e "${GREEN}✅ All tests passed${NC}"
        ;;
    
    "integration")
        echo -e "${YELLOW}Integration Tests:${NC}"
        go test . -timeout 120s -coverprofile=coverage.out 2>&1 | grep -E "FAIL|PASS.*coverage|Error:|panic:" || echo -e "${GREEN}✅ All tests passed${NC}"
        ;;
    
    "all")
        # Run all test suites and collect coverage
        echo -e "${YELLOW}1. Unit Tests:${NC}"
        go test ./llm -run "TestAnalysisPrompts|TestEnvironmentVariables" -coverprofile=coverage1.out 2>&1 | grep -E "FAIL|PASS.*coverage|Error:|panic:" || echo -e "${GREEN}✅ Passed${NC}"
        
        echo -e "\n${YELLOW}2. Provider Tests:${NC}"
        go test ./llm -run "TestProviderConnections" -timeout 30s -coverprofile=coverage2.out 2>&1 | grep -E "FAIL|PASS.*coverage|Error:|panic:" || echo -e "${GREEN}✅ Passed${NC}"
        
        echo -e "\n${YELLOW}3. Model Tests:${NC}"
        go test ./llm -run "TestProviderModels" -timeout 60s -coverprofile=coverage3.out 2>&1 | grep -E "FAIL|PASS.*coverage|Error:|panic:" || echo -e "${GREEN}✅ Passed${NC}"
        
        echo -e "\n${YELLOW}4. Integration Tests:${NC}"
        go test . -timeout 120s -coverprofile=coverage4.out 2>&1 | grep -E "FAIL|PASS.*coverage|Error:|panic:" || echo -e "${GREEN}✅ Passed${NC}"
        
        # Merge coverage files if gocovmerge is available
        if command -v gocovmerge &> /dev/null; then
            gocovmerge coverage*.out > coverage.out
            rm coverage[1-4].out
        else
            # Use the last coverage file if merge tool not available
            mv coverage4.out coverage.out 2>/dev/null || true
            rm -f coverage[1-3].out
        fi
        ;;
    
    *)
        echo "Usage: $0 [quick|providers|models|integration|all]"
        exit 1
        ;;
esac

# Show coverage summary if coverage file exists
if [ -f coverage.out ]; then
    echo -e "\n${YELLOW}Coverage Summary:${NC}"
    echo "----------------"
    go tool cover -func=coverage.out | grep total | awk '{print "Total: " $3}'
    
    # Generate HTML report
    go tool cover -html=coverage.out -o coverage.html 2>/dev/null
    echo -e "HTML report: ${GREEN}coverage.html${NC}"
fi