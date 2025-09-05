#!/bin/bash

# Test script for nginx integration
echo "Testing nginx integration..."

# Test 1: Check if containers are running
echo "1. Checking if containers are running..."
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(nginx|backend|frontend)"

# Test 2: Test nginx configuration
echo -e "\n2. Testing nginx configuration..."
docker exec undecided-proxy-manager_nginx_1 nginx -t

# Test 3: Test nginx reload
echo -e "\n3. Testing nginx reload..."
docker exec undecided-proxy-manager_nginx_1 nginx -s reload

# Test 4: Check nginx status
echo -e "\n4. Checking nginx status..."
docker exec undecided-proxy-manager_nginx_1 nginx -s reload && echo "Nginx reload successful" || echo "Nginx reload failed"

# Test 5: Test API endpoints
echo -e "\n5. Testing API endpoints..."
echo "Testing nginx reload endpoint..."
curl -X POST http://localhost:6081/api/v1/nginx/reload -H "Content-Type: application/json" || echo "API test failed"

echo -e "\n6. Testing nginx config test endpoint..."
curl -X POST http://localhost:6081/api/v1/nginx/test -H "Content-Type: application/json" || echo "API test failed"

echo -e "\nNginx integration test completed!"
