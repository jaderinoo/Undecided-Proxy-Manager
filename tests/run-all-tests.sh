#!/bin/bash

# UPM Test Suite Runner
# Runs all available tests for the Undecided Proxy Manager

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

echo -e "${PURPLE}🧪 UPM Test Suite Runner${NC}"
echo "=========================="
echo ""

# Function to run a test
run_test() {
    local test_name=$1
    local test_script=$2

    echo -e "${YELLOW}Running: $test_name${NC}"
    echo "----------------------------------------"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # Check if we're in the tests directory
    if [ -f "$test_script" ]; then
        script_path="$test_script"
    elif [ -f "tests/$test_script" ]; then
        script_path="tests/$test_script"
    else
        echo -e "${RED}❌ $test_name FAILED - Script not found: $test_script${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo ""
        return
    fi

    if bash "$script_path"; then
        echo -e "${GREEN}✅ $test_name PASSED${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ $test_name FAILED${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi

    echo ""
}

# Check if we're in the right directory
if [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}❌ Error: Not in UPM root directory${NC}"
    echo "Please run this script from the UPM root directory"
    exit 1
fi

# Check if services are running
echo -e "${BLUE}Checking if UPM services are running...${NC}"
if ! docker ps | grep -q "undecided-proxy-manager_nginx_1\|undecided-proxy-manager_backend_1\|undecided-proxy-manager_frontend_1"; then
    echo -e "${YELLOW}⚠️  UPM services don't appear to be running${NC}"
    echo "Starting services with: ./dev.sh up"
    echo ""

    if [ -f "../dev.sh" ]; then
        echo "Starting development environment..."
        ../dev.sh up
        echo "Waiting for services to start..."
        sleep 10
    else
        echo -e "${RED}❌ dev.sh not found. Please start services manually${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}✅ Services are running${NC}"
echo ""

# Run all tests
echo -e "${PURPLE}Starting Test Suite...${NC}"
echo ""

# Test 1: Nginx Integration
run_test "Nginx Integration Test" "test-nginx.sh"

# Test 2: DNS Service
run_test "DNS Service Test" "test-dns.sh"

# Test 3: Container Management
run_test "Container Management Test" "test-containers.sh"

# Test 4: Proxy Functionality
run_test "Proxy Functionality Test" "test-proxies.sh"

# Summary
echo -e "${PURPLE}Test Suite Summary${NC}"
echo "=================="
echo -e "Total Tests: ${BLUE}$TOTAL_TESTS${NC}"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 All tests passed! UPM is working correctly!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}❌ Some tests failed. Please check the output above.${NC}"
    exit 1
fi
