#!/bin/bash

# Test runner with clean summary output

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test results tracking
total_tests=0
failed_tests=0
test_names=()
test_results=()
test_coverage=()

# Function to run a test suite
run_test_suite() {
    local name=$1
    local package=$2
    local pattern=$3
    local timeout=$4
    
    echo -ne "${YELLOW}${name}...${NC} "
    
    # Run test and capture output
    if [ -n "$pattern" ]; then
        output=$(go test "$package" -run "$pattern" -timeout "$timeout" -coverprofile="coverage_${total_tests}.out" 2>&1)
    else
        output=$(go test "$package" -timeout "$timeout" -coverprofile="coverage_${total_tests}.out" 2>&1)
    fi
    
    exit_code=$?
    
    # Store test name
    test_names+=("$name")
    
    # Extract coverage
    coverage=$(echo "$output" | grep -o 'coverage: [0-9.]*%' | head -1 | awk '{print $2}')
    test_coverage+=("${coverage:-N/A}")
    
    if [ $exit_code -eq 0 ]; then
        test_results+=("PASS")
        echo -e "${GREEN}✓${NC} ${coverage:-N/A}"
    else
        test_results+=("FAIL")
        failed_tests=$((failed_tests + 1))
        echo -e "${RED}✗${NC}"
        # Store failure details
        echo "$output" | grep -E "(FAIL|Error:|panic:|--- FAIL)" > "test_failure_${total_tests}.log"
    fi
    
    total_tests=$((total_tests + 1))
}

# Header
echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo -e "${BLUE}    Second Opinion Test Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo ""

# Determine test mode
TEST_MODE=${1:-all}

case $TEST_MODE in
    "quick")
        run_test_suite "Unit Tests" "./llm" "TestAnalysisPrompts|TestEnvironmentVariables" "10s"
        ;;
    
    "all")
        run_test_suite "Unit Tests" "./llm" "TestAnalysisPrompts|TestEnvironmentVariables" "10s"
        run_test_suite "Provider Tests" "./llm" "TestProviderConnections" "30s"
        run_test_suite "Model Tests" "./llm" "TestProviderModels" "60s"
        run_test_suite "Integration Tests" "." "" "120s"
        ;;
    
    *)
        echo "Usage: $0 [quick|all]"
        exit 1
        ;;
esac

# Summary
echo ""
echo -e "${BLUE}───────────────────────────────────────${NC}"

# Merge coverage files
if [ $total_tests -gt 0 ]; then
    # Simple merge: concatenate all coverage files
    echo "mode: set" > coverage.out
    for i in $(seq 0 $((total_tests - 1))); do
        if [ -f "coverage_${i}.out" ]; then
            tail -n +2 "coverage_${i}.out" >> coverage.out 2>/dev/null || true
            rm -f "coverage_${i}.out"
        fi
    done
fi

# Calculate total coverage
if [ -f coverage.out ] && [ -s coverage.out ]; then
    total_coverage=$(go tool cover -func=coverage.out 2>/dev/null | grep total | awk '{print $3}')
    if [ -n "$total_coverage" ]; then
        echo -e "Total Coverage: ${GREEN}${total_coverage}${NC}"
    fi
    
    # Generate HTML report
    go tool cover -html=coverage.out -o coverage.html 2>/dev/null
fi

# Test summary
echo ""
if [ $failed_tests -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed! (${total_tests}/${total_tests})${NC}"
else
    echo -e "${RED}❌ ${failed_tests} test(s) failed (${failed_tests}/${total_tests})${NC}"
    echo ""
    echo "Failed tests:"
    for i in $(seq 0 $((total_tests - 1))); do
        if [ "${test_results[$i]}" == "FAIL" ]; then
            echo -e "  ${RED}✗${NC} ${test_names[$i]}"
        fi
    done
    
    # Show failure details
    echo ""
    echo "Failure details:"
    for i in $(seq 0 $((total_tests - 1))); do
        if [ -f "test_failure_${i}.log" ]; then
            echo -e "${YELLOW}$(cat test_failure_${i}.log)${NC}"
            rm -f "test_failure_${i}.log"
        fi
    done
    
    exit 1
fi