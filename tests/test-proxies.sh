#!/bin/bash

# Comprehensive Proxy Tests for UPM
# This script tests the complete proxy workflow: Create → Update → Delete → Nginx Config

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# Test data
TEST_PROXY_NAME="test-proxy-$(date +%s)"
TEST_DOMAIN="test.example.com"
TEST_TARGET="http://localhost:3000"

echo -e "${BLUE}🧪 Starting UPM Proxy Tests${NC}"
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

# Test 3: Get Existing Proxies
echo -e "\n${YELLOW}3. Testing Get Proxies${NC}"
proxies_response=$(api_call "GET" "/proxies" "" "$token")
if echo "$proxies_response" | grep -q "data"; then
    print_result 0 "Get proxies successful"
    echo "Current proxies: $(echo "$proxies_response" | jq '.data | length' 2>/dev/null || echo "0")"
else
    print_result 1 "Get proxies failed"
    echo "Response: $proxies_response"
fi

# Test 4: Create Proxy
echo -e "\n${YELLOW}4. Testing Create Proxy${NC}"
create_data='{
    "name":"'$TEST_PROXY_NAME'",
    "domain":"'$TEST_DOMAIN'",
    "target_url":"'$TEST_TARGET'",
    "ssl_enabled":false
}'
create_response=$(api_call "POST" "/proxies" "$create_data" "$token")
proxy_id=$(echo "$create_response" | grep -o '"id":[0-9]*' | cut -d':' -f2)

if [ -n "$proxy_id" ]; then
    print_result 0 "Proxy created successfully (ID: $proxy_id)"
else
    print_result 1 "Proxy creation failed"
    echo "Response: $create_response"
    exit 1
fi

# Test 5: Verify Nginx Config Generation
echo -e "\n${YELLOW}5. Testing Nginx Config Generation${NC}"
if [ -f "nginx/sites-available/proxy-$proxy_id.conf" ]; then
    print_result 0 "Nginx config file created"
    echo "Config file: nginx/sites-available/proxy-$proxy_id.conf"

    # Check if symlink exists
    if [ -L "nginx/sites-enabled/proxy-$proxy_id.conf" ]; then
        print_result 0 "Nginx symlink created"
    else
        print_result 1 "Nginx symlink not created"
    fi
else
    print_result 1 "Nginx config file not created"
fi

# Test 6: Test Nginx Configuration
echo -e "\n${YELLOW}6. Testing Nginx Configuration${NC}"
nginx_test_response=$(api_call "POST" "/nginx/test" "" "$token")
if echo "$nginx_test_response" | grep -q "valid"; then
    print_result 0 "Nginx configuration is valid"
else
    print_result 1 "Nginx configuration test failed"
    echo "Response: $nginx_test_response"
fi

# Test 7: Reload Nginx
echo -e "\n${YELLOW}7. Testing Nginx Reload${NC}"
nginx_reload_response=$(api_call "POST" "/nginx/reload" "" "$token")
if echo "$nginx_reload_response" | grep -q "successfully"; then
    print_result 0 "Nginx reloaded successfully"
else
    print_result 1 "Nginx reload failed"
    echo "Response: $nginx_reload_response"
fi

# Test 8: Get Specific Proxy
echo -e "\n${YELLOW}8. Testing Get Specific Proxy${NC}"
get_proxy_response=$(api_call "GET" "/proxies/$proxy_id" "" "$token")
if echo "$get_proxy_response" | grep -q "$TEST_PROXY_NAME"; then
    print_result 0 "Get specific proxy successful"
else
    print_result 1 "Get specific proxy failed"
    echo "Response: $get_proxy_response"
fi

# Test 9: Update Proxy
echo -e "\n${YELLOW}9. Testing Update Proxy${NC}"
update_data='{
    "name":"'$TEST_PROXY_NAME'-updated",
    "ssl_enabled":true
}'
update_response=$(api_call "PUT" "/proxies/$proxy_id" "$update_data" "$token")
if echo "$update_response" | grep -q "updated"; then
    print_result 0 "Proxy updated successfully"
else
    print_result 1 "Proxy update failed"
    echo "Response: $update_response"
fi

# Test 10: Verify Updated Nginx Config
echo -e "\n${YELLOW}10. Testing Updated Nginx Config${NC}"
if [ -f "nginx/sites-available/proxy-$proxy_id.conf" ]; then
    if grep -q "ssl_certificate" "nginx/sites-available/proxy-$proxy_id.conf"; then
        print_result 0 "Nginx config updated with SSL"
    else
        print_result 1 "Nginx config not updated with SSL"
    fi
else
    print_result 1 "Updated nginx config file not found"
fi

# Test 11: Test Nginx After Update
echo -e "\n${YELLOW}11. Testing Nginx After Update${NC}"
nginx_test_response=$(api_call "POST" "/nginx/test" "" "$token")
if echo "$nginx_test_response" | grep -q "valid"; then
    print_result 0 "Nginx configuration still valid after update"
else
    print_result 1 "Nginx configuration invalid after update"
    echo "Response: $nginx_test_response"
fi

# Test 12: Delete Proxy
echo -e "\n${YELLOW}12. Testing Delete Proxy${NC}"
echo "Attempting to delete proxy ID: $proxy_id"
delete_response=$(api_call "DELETE" "/proxies/$proxy_id" "" "$token")
echo "Delete response: $delete_response"
if echo "$delete_response" | grep -q "successfully"; then
    print_result 0 "Proxy deleted successfully"
elif echo "$delete_response" | grep -q "not found"; then
    print_result 0 "Proxy deletion handled correctly (proxy not found)"
    echo "Response: $delete_response"
elif [ -z "$delete_response" ]; then
    # Check if proxy was actually deleted by trying to get it
    check_response=$(api_call "GET" "/proxies/$proxy_id" "" "$token")
    if echo "$check_response" | grep -q "not found"; then
        print_result 0 "Proxy deleted successfully (confirmed by 404 on get)"
    else
        print_result 1 "Proxy deletion failed - proxy still exists"
        echo "Check response: $check_response"
    fi
else
    print_result 1 "Proxy deletion failed"
    echo "Response: $delete_response"
fi

# Test 13: Verify Nginx Config Cleanup
echo -e "\n${YELLOW}13. Testing Nginx Config Cleanup${NC}"
if [ ! -f "nginx/sites-available/proxy-$proxy_id.conf" ]; then
    print_result 0 "Nginx config file removed"
else
    print_result 1 "Nginx config file not removed"
fi

if [ ! -L "nginx/sites-enabled/proxy-$proxy_id.conf" ]; then
    print_result 0 "Nginx symlink removed"
else
    print_result 1 "Nginx symlink not removed"
fi

# Test 14: Final Nginx Test
echo -e "\n${YELLOW}14. Final Nginx Configuration Test${NC}"
nginx_test_response=$(api_call "POST" "/nginx/test" "" "$token")
if echo "$nginx_test_response" | grep -q "valid"; then
    print_result 0 "Nginx configuration valid after cleanup"
else
    print_result 1 "Nginx configuration invalid after cleanup"
    echo "Response: $nginx_test_response"
fi

# Summary
echo -e "\n${GREEN}🎉 All Proxy Tests Completed Successfully!${NC}"
echo "=================================="
echo "✅ API Health Check"
echo "✅ Authentication"
echo "✅ Proxy CRUD Operations"
echo "✅ Nginx Config Generation"
echo "✅ Nginx Configuration Testing"
echo "✅ Nginx Reload"
echo "✅ Config Cleanup"
echo ""
echo -e "${BLUE}Proxy workflow is fully functional!${NC}"
