#!/bin/bash

# Container Management Tests for UPM
# Tests Docker container discovery, inspection, and stats

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
API_BASE="http://localhost:6081/api/v1"
ADMIN_USER="admin"

# Read password from .env file or use dev bypass
if [ -f "../.env" ]; then
    ADMIN_PASSWORD=$(grep "^ADMIN_PASSWORD=" ../.env | cut -d'=' -f2 | tr -d '"' | tr -d "'")
    # If it's hashed, we need the unencrypted version
    if [[ "$ADMIN_PASSWORD" == *"$2a$"* ]]; then
        # Extract the unencrypted password from the comment
        ADMIN_PASSWORD=$(grep "unencrypted:" ../.env | cut -d':' -f2 | tr -d ' ')
    fi
else
    # Use development bypass when no .env file exists
    ADMIN_PASSWORD="devtest"  # development bypass
fi

echo -e "${PURPLE}🐳 Starting UPM Container Tests${NC}"
echo "=================================="

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
        exit 1
    fi
}

# Function to make API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local token=$4

    if [ -n "$token" ]; then
        if [ -n "$data" ]; then
            curl -s -X $method "$API_BASE$endpoint" \
                -H "Content-Type: application/json" \
                -H "Authorization: Bearer $token" \
                -d "$data"
        else
            curl -s -X $method "$API_BASE$endpoint" \
                -H "Authorization: Bearer $token"
        fi
    else
        if [ -n "$data" ]; then
            curl -s -X $method "$API_BASE$endpoint" \
                -H "Content-Type: application/json" \
                -d "$data"
        else
            curl -s -X $method "$API_BASE$endpoint"
        fi
    fi
}

# Test 1: Health Check
echo -e "\n${YELLOW}1. Testing API Health Check${NC}"
health_response=$(curl -s "http://localhost:6081/health")
if echo "$health_response" | grep -q "ok"; then
    print_result 0 "API is healthy"
else
    print_result 1 "API health check failed"
fi

# Test 2: Authentication
echo -e "\n${YELLOW}2. Testing Authentication${NC}"
login_data='{"username":"'$ADMIN_USER'","password":"'$ADMIN_PASSWORD'"}'
auth_response=$(api_call "POST" "/auth/login" "$login_data")
token=$(echo "$auth_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$token" ]; then
    print_result 0 "Authentication successful"
    echo "Token: ${token:0:20}..."
else
    print_result 1 "Authentication failed"
    echo "Response: $auth_response"
    exit 1
fi

# Test 3: Get All Containers
echo -e "\n${YELLOW}3. Testing Get All Containers${NC}"
containers_response=$(api_call "GET" "/containers" "" "$token")
if echo "$containers_response" | grep -q "data"; then
    print_result 0 "Containers retrieved successfully"
    container_count=$(echo "$containers_response" | jq '.data | length' 2>/dev/null || echo "0")
    echo "Found $container_count containers"

    # Extract first container ID for further tests
    first_container_id=$(echo "$containers_response" | jq -r '.data[0].id' 2>/dev/null || echo "")
    if [ -n "$first_container_id" ] && [ "$first_container_id" != "null" ]; then
        echo "First container ID: $first_container_id"
    fi
else
    print_result 1 "Failed to retrieve containers"
    echo "Response: $containers_response"
    exit 1
fi

# Test 4: Get Specific Container
echo -e "\n${YELLOW}4. Testing Get Specific Container${NC}"
if [ -n "$first_container_id" ] && [ "$first_container_id" != "null" ]; then
    container_response=$(api_call "GET" "/containers/$first_container_id" "" "$token")
    if echo "$container_response" | grep -q "id"; then
        print_result 0 "Specific container retrieved successfully"

        # Extract container details
        container_name=$(echo "$container_response" | jq -r '.name' 2>/dev/null || echo "unknown")
        container_image=$(echo "$container_response" | jq -r '.image' 2>/dev/null || echo "unknown")
        container_status=$(echo "$container_response" | jq -r '.status' 2>/dev/null || echo "unknown")

        echo "Container: $container_name"
        echo "Image: $container_image"
        echo "Status: $container_status"
    else
        print_result 1 "Failed to retrieve specific container"
        echo "Response: $container_response"
    fi
else
    print_result 0 "No containers available for specific test"
fi

# Test 5: Get Container Stats
echo -e "\n${YELLOW}5. Testing Get Container Stats${NC}"
if [ -n "$first_container_id" ] && [ "$first_container_id" != "null" ]; then
    stats_response=$(api_call "GET" "/containers/$first_container_id/stats" "" "$token")
    if echo "$stats_response" | grep -q "stats"; then
        print_result 0 "Container stats retrieved successfully"

        # Extract basic stats info
        cpu_usage=$(echo "$stats_response" | jq -r '.cpu_stats.cpu_usage.total_usage' 2>/dev/null || echo "N/A")
        memory_usage=$(echo "$stats_response" | jq -r '.memory_stats.usage' 2>/dev/null || echo "N/A")

        echo "CPU Usage: $cpu_usage"
        echo "Memory Usage: $memory_usage bytes"
    else
        print_result 1 "Failed to retrieve container stats"
        echo "Response: $stats_response"
    fi
else
    print_result 0 "No containers available for stats test"
fi

# Test 6: Test Container Filtering
echo -e "\n${YELLOW}6. Testing Container Data Structure${NC}"
if [ -n "$first_container_id" ] && [ "$first_container_id" != "null" ]; then
    container_response=$(api_call "GET" "/containers/$first_container_id" "" "$token")

    # Check for required fields
    required_fields=("id" "name" "image" "status" "state" "created")
    missing_fields=()

    for field in "${required_fields[@]}"; do
        if ! echo "$container_response" | jq -e ".$field" > /dev/null 2>&1; then
            missing_fields+=("$field")
        fi
    done

    if [ ${#missing_fields[@]} -eq 0 ]; then
        print_result 0 "Container data structure is complete"
    else
        print_result 1 "Container data missing fields: ${missing_fields[*]}"
    fi
else
    print_result 0 "No containers available for structure test"
fi

# Test 7: Test Container Port Information
echo -e "\n${YELLOW}7. Testing Container Port Information${NC}"
if [ -n "$first_container_id" ] && [ "$first_container_id" != "null" ]; then
    container_response=$(api_call "GET" "/containers/$first_container_id" "" "$token")

    # Check if ports array exists
    if echo "$container_response" | jq -e '.ports' > /dev/null 2>&1; then
        port_count=$(echo "$container_response" | jq '.ports | length' 2>/dev/null || echo "0")
        print_result 0 "Container port information available ($port_count ports)"

        if [ "$port_count" -gt 0 ]; then
            echo "Port mappings:"
            echo "$container_response" | jq -r '.ports[] | "  \(.private_port)/\(.type) -> \(.public_port)"' 2>/dev/null || echo "  No port details available"
        fi
    else
        print_result 1 "Container port information missing"
    fi
else
    print_result 0 "No containers available for port test"
fi

# Test 8: Test Container Labels
echo -e "\n${YELLOW}8. Testing Container Labels${NC}"
if [ -n "$first_container_id" ] && [ "$first_container_id" != "null" ]; then
    container_response=$(api_call "GET" "/containers/$first_container_id" "" "$token")

    # Check if labels exist
    if echo "$container_response" | jq -e '.labels' > /dev/null 2>&1; then
        label_count=$(echo "$container_response" | jq '.labels | length' 2>/dev/null || echo "0")
        print_result 0 "Container labels available ($label_count labels)"

        if [ "$label_count" -gt 0 ]; then
            echo "Labels:"
            echo "$container_response" | jq -r '.labels | to_entries[] | "  \(.key): \(.value)"' 2>/dev/null || echo "  No label details available"
        fi
    else
        print_result 1 "Container labels missing"
    fi
else
    print_result 0 "No containers available for labels test"
fi

# Test 9: Test Container Mounts
echo -e "\n${YELLOW}9. Testing Container Mounts${NC}"
if [ -n "$first_container_id" ] && [ "$first_container_id" != "null" ]; then
    container_response=$(api_call "GET" "/containers/$first_container_id" "" "$token")

    # Check if mounts exist
    if echo "$container_response" | jq -e '.mounts' > /dev/null 2>&1; then
        mount_count=$(echo "$container_response" | jq '.mounts | length' 2>/dev/null || echo "0")
        print_result 0 "Container mounts available ($mount_count mounts)"

        if [ "$mount_count" -gt 0 ]; then
            echo "Mounts:"
            echo "$container_response" | jq -r '.mounts[] | "  \(.source) -> \(.destination) (\(.type))"' 2>/dev/null || echo "  No mount details available"
        fi
    else
        print_result 1 "Container mounts missing"
    fi
else
    print_result 0 "No containers available for mounts test"
fi

# Test 10: Test Error Handling
echo -e "\n${YELLOW}10. Testing Error Handling${NC}"

# Test with invalid container ID
invalid_id="invalid-container-id-12345"
invalid_response=$(api_call "GET" "/containers/$invalid_id" "" "$token")
if echo "$invalid_response" | grep -q "error\|not found\|failed"; then
    print_result 0 "Invalid container ID correctly handled"
else
    print_result 1 "Invalid container ID not handled properly"
    echo "Response: $invalid_response"
fi

# Test with invalid stats request
invalid_stats_response=$(api_call "GET" "/containers/$invalid_id/stats" "" "$token")
if echo "$invalid_stats_response" | grep -q "error\|not found\|failed"; then
    print_result 0 "Invalid stats request correctly handled"
else
    print_result 1 "Invalid stats request not handled properly"
fi

# Test 11: Test Container State Information
echo -e "\n${YELLOW}11. Testing Container State Information${NC}"
if [ -n "$first_container_id" ] && [ "$first_container_id" != "null" ]; then
    container_response=$(api_call "GET" "/containers/$first_container_id" "" "$token")

    # Check for state information
    state_fields=("started_at" "finished_at" "network_mode")
    state_info_available=true

    for field in "${state_fields[@]}"; do
        if ! echo "$container_response" | jq -e ".$field" > /dev/null 2>&1; then
            state_info_available=false
            break
        fi
    done

    if [ "$state_info_available" = true ]; then
        print_result 0 "Container state information available"

        started_at=$(echo "$container_response" | jq -r '.started_at' 2>/dev/null || echo "N/A")
        network_mode=$(echo "$container_response" | jq -r '.network_mode' 2>/dev/null || echo "N/A")

        echo "Started at: $started_at"
        echo "Network mode: $network_mode"
    else
        print_result 1 "Container state information incomplete"
    fi
else
    print_result 0 "No containers available for state test"
fi

# Test 12: Test Container Size Information
echo -e "\n${YELLOW}12. Testing Container Size Information${NC}"
if [ -n "$first_container_id" ] && [ "$first_container_id" != "null" ]; then
    container_response=$(api_call "GET" "/containers/$first_container_id" "" "$token")

    # Check for size information
    if echo "$container_response" | jq -e '.size_rw' > /dev/null 2>&1 && echo "$container_response" | jq -e '.size_root_fs' > /dev/null 2>&1; then
        print_result 0 "Container size information available"

        size_rw=$(echo "$container_response" | jq -r '.size_rw' 2>/dev/null || echo "0")
        size_root_fs=$(echo "$container_response" | jq -r '.size_root_fs' 2>/dev/null || echo "0")

        echo "Read/Write size: $size_rw bytes"
        echo "Root filesystem size: $size_root_fs bytes"
    else
        print_result 1 "Container size information missing"
    fi
else
    print_result 0 "No containers available for size test"
fi

# Summary
echo -e "\n${GREEN}🎉 All Container Tests Completed Successfully!${NC}"
echo "============================================="
echo "✅ API Health Check"
echo "✅ Authentication"
echo "✅ Container Discovery"
echo "✅ Container Details"
echo "✅ Container Stats"
echo "✅ Data Structure Validation"
echo "✅ Port Information"
echo "✅ Label Information"
echo "✅ Mount Information"
echo "✅ Error Handling"
echo "✅ State Information"
echo "✅ Size Information"
echo ""
echo -e "${BLUE}Container management service is fully functional!${NC}"
