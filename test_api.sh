#!/bin/bash

# Test script for GoMall Microservice E-commerce System

echo "=================================="
echo "GoMall System Test"
echo "=================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to test an endpoint
test_endpoint() {
    local name=$1
    local url=$2
    local method=$3
    local data=$4
    
    echo "Testing: $name"
    
    if [ "$method" == "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" -H "Content-Type: application/json" -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✓ PASSED${NC} (HTTP $http_code)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
    else
        echo -e "${RED}✗ FAILED${NC} (HTTP $http_code)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo "$body"
    fi
    echo ""
}

echo "1. Health Check"
test_endpoint "API Gateway Health" "http://localhost:8080/health" "GET"

echo "2. User Service Tests"
test_endpoint "Register User" "http://localhost:8080/api/register" "POST" '{"username":"bob","email":"bob@example.com","password":"secret123"}'
test_endpoint "Login User" "http://localhost:8080/api/login" "POST" '{"username":"bob","password":"secret123"}'
test_endpoint "Get User by ID" "http://localhost:8080/api/user?id=1" "GET"
test_endpoint "List All Users" "http://localhost:8080/api/users" "GET"

echo "3. Product Service Tests"
test_endpoint "List All Products" "http://localhost:8080/api/products" "GET"
test_endpoint "Get Product by ID" "http://localhost:8080/api/product?id=1" "GET"
test_endpoint "Create New Product" "http://localhost:8080/api/product/create" "POST" '{"name":"Tablet","description":"Portable tablet","price":499.99,"stock":15,"category":"Electronics"}'

echo "4. Order Service Tests"
test_endpoint "Create Order" "http://localhost:8080/api/order/create" "POST" '{"user_id":1,"items":[{"product_id":1,"quantity":1},{"product_id":2,"quantity":2}]}'
test_endpoint "Get Order by ID" "http://localhost:8080/api/order?id=1" "GET"
test_endpoint "List Orders by User" "http://localhost:8080/api/orders?user_id=1" "GET"
test_endpoint "List All Orders" "http://localhost:8080/api/orders" "GET"

echo "=================================="
echo "Test Results"
echo "=================================="
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
echo "Total: $((TESTS_PASSED + TESTS_FAILED))"
echo "=================================="

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
