#!/bin/bash

# Create User Example Script
# This script demonstrates how to create a new user and get a JWT token

API_BASE="http://localhost:8080/api/v1"

echo "🚀 Creating a new user..."

# Create user
USER_RESPONSE=$(curl -s -X POST "$API_BASE/users" \
  -H "Content-Type: application/json" \
  -d '{
    "risk_tolerance": "moderate"
  }')

echo "User created: $USER_RESPONSE"

# Extract user ID (assuming jq is available)
USER_ID=$(echo $USER_RESPONSE | jq -r '.id')

if [ "$USER_ID" != "null" ] && [ "$USER_ID" != "" ]; then
  echo "✅ User created successfully with ID: $USER_ID"
  
  echo "🔐 Logging in to get JWT token..."
  
  # Login to get token
  LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
    -H "Content-Type: application/json" \
    -d "{
      \"user_id\": $USER_ID
    }")
  
  echo "Login response: $LOGIN_RESPONSE"
  
  # Extract token
  TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
  
  if [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
    echo "✅ Login successful!"
    echo "🎫 JWT Token: $TOKEN"
    echo ""
    echo "💡 You can now use this token for authenticated requests:"
    echo "   Authorization: Bearer $TOKEN"
    echo ""
    echo "📝 Example: Get user profile"
    echo "   curl -H \"Authorization: Bearer $TOKEN\" $API_BASE/users/$USER_ID"
  else
    echo "❌ Login failed"
  fi
else
  echo "❌ User creation failed"
fi