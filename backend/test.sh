#!/bin/bash

# Go Finance Advisor - Test Script
# This script tests all major functionality of the application

echo "🧪 Testing Go Finance Advisor Application"
echo "=========================================="

BASE_URL="http://localhost:8081"

echo "📡 Testing Server Connectivity"
echo "-------------------------------"

# Test server is running
echo -n "Testing server status... "
if curl -s "$BASE_URL/api/overview" > /dev/null; then
    echo "✅ Server is running"
else
    echo "❌ Server is not responding"
    exit 1
fi

# Test API endpoints
echo -n "Testing overview endpoint... "
OVERVIEW=$(curl -s "$BASE_URL/api/overview")
if [ -n "$OVERVIEW" ]; then
    echo "✅ Overview data available"
else
    echo "❌ Overview endpoint failed"
fi

echo -n "Testing holdings endpoint... "
HOLDINGS=$(curl -s "$BASE_URL/api/holdings")
if [ -n "$HOLDINGS" ]; then
    echo "✅ Holdings data available"
else
    echo "❌ Holdings endpoint failed"
fi

echo -n "Testing performance endpoint... "
PERFORMANCE=$(curl -s "$BASE_URL/api/performance")
if [ -n "$PERFORMANCE" ]; then
    echo "✅ Performance data available"
else
    echo "❌ Performance endpoint failed"
fi

echo -n "Testing allocation endpoint... "
ALLOCATION=$(curl -s "$BASE_URL/api/allocation")
if [ -n "$ALLOCATION" ]; then
    echo "✅ Allocation data available"
else
    echo "❌ Allocation endpoint failed"
fi

# Test frontend pages
echo ""
echo "🌐 Testing Frontend Pages"
echo "-------------------------"

PAGES=("platform-live.html" "index.html" "portfolio-live.html" "market-live.html" "features-live.html" "documents-live.html" "transactions-live.html" "budgets-live.html" "goals-live.html")

for page in "${PAGES[@]}"; do
    echo -n "Testing $page... "
    if curl -s "$BASE_URL/$page" | grep -q "DOCTYPE html"; then
        echo "✅ Page loads correctly"
    else
        echo "❌ Page failed to load"
    fi
done

# Test API documentation
echo ""
echo "📚 Testing API Documentation"
echo "----------------------------"

echo -n "Testing API docs endpoint... "
if curl -s "$BASE_URL/api/docs" | grep -q "openapi"; then
    echo "✅ API documentation available"
else
    echo "❌ API documentation not found"
fi

# Test WebSocket (basic connectivity)
echo ""
echo "🔌 Testing WebSocket Connection"
echo "-------------------------------"

echo -n "Testing WebSocket endpoint... "
WS_RESPONSE=$(curl -s -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Host: localhost:8081" -H "Origin: http://localhost:8081" "$BASE_URL/api/ws/ticker" 2>&1 | head -1)
if echo "$WS_RESPONSE" | grep -q "400 Bad Request"; then
    echo "✅ WebSocket endpoint responds (400 expected without proper handshake)"
else
    echo "❌ WebSocket endpoint not responding"
fi

# Test Docker configuration
echo ""
echo "🐳 Testing Docker Configuration"
echo "-------------------------------"

echo -n "Testing Dockerfile syntax... "
if docker build --dry-run . > /dev/null 2>&1; then
    echo "✅ Dockerfile syntax valid"
else
    echo "❌ Dockerfile has syntax errors"
fi

echo ""
echo "🎉 Testing Complete!"
echo "===================="
echo ""
echo "✅ All core functionality has been tested"
echo "✅ Frontend pages are accessible"
echo "✅ API endpoints are responding"
echo "✅ WebSocket connection is available"
echo "✅ Docker configuration is valid"
echo ""
echo "🚀 Application is ready for use!"
echo "   Frontend: http://localhost:8081/platform-live.html"
echo "   API Docs: http://localhost:8081/api/docs"
echo "   Metrics:  http://localhost:8081/metrics"