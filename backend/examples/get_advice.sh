#!/bin/bash

# Get Investment Advice Example Script
# This script demonstrates the complete flow: create user, add transactions, get advice

API_BASE="http://localhost:8080/api/v1"

echo "🚀 Personal Finance Advisor Demo"
echo "================================"

# Step 1: Create user
echo "📝 Step 1: Creating user..."
USER_RESPONSE=$(curl -s -X POST "$API_BASE/users" \
  -H "Content-Type: application/json" \
  -d '{
    "risk_tolerance": "moderate"
  }')

USER_ID=$(echo $USER_RESPONSE | jq -r '.id')
echo "✅ User created with ID: $USER_ID"

# Step 2: Login
echo "🔐 Step 2: Getting authentication token..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": $USER_ID
  }")

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
echo "✅ Token obtained"

# Step 3: Add some transactions
echo "💰 Step 3: Adding sample transactions..."

# Add income transactions
curl -s -X POST "$API_BASE/users/$USER_ID/transactions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "income",
    "description": "Monthly Salary",
    "amount": 5000.00,
    "category": "work",
    "date": "2024-01-15T00:00:00Z"
  }' > /dev/null

curl -s -X POST "$API_BASE/users/$USER_ID/transactions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "income",
    "description": "Freelance Work",
    "amount": 1500.00,
    "category": "freelance",
    "date": "2024-01-20T00:00:00Z"
  }' > /dev/null

# Add expense transactions
curl -s -X POST "$API_BASE/users/$USER_ID/transactions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "expense",
    "description": "Rent",
    "amount": 1500.00,
    "category": "housing",
    "date": "2024-01-01T00:00:00Z"
  }' > /dev/null

curl -s -X POST "$API_BASE/users/$USER_ID/transactions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "expense",
    "description": "Groceries",
    "amount": 800.00,
    "category": "food",
    "date": "2024-01-10T00:00:00Z"
  }' > /dev/null

curl -s -X POST "$API_BASE/users/$USER_ID/transactions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "expense",
    "description": "Utilities",
    "amount": 200.00,
    "category": "utilities",
    "date": "2024-01-05T00:00:00Z"
  }' > /dev/null

echo "✅ Sample transactions added"
echo "   💵 Income: $5,000 (salary) + $1,500 (freelance) = $6,500"
echo "   💸 Expenses: $1,500 (rent) + $800 (food) + $200 (utilities) = $2,500"
echo "   💰 Net Monthly Savings: $4,000"

# Step 4: Get investment advice
echo "🤖 Step 4: Getting AI investment advice..."
ADVICE_RESPONSE=$(curl -s -X GET "$API_BASE/users/$USER_ID/advice" \
  -H "Authorization: Bearer $TOKEN")

echo "📊 Investment Advice:"
echo "$ADVICE_RESPONSE" | jq .

# Step 5: Try different risk profiles
echo "🎯 Step 5: Testing different risk profiles..."

echo "\n🛡️  Conservative Profile:"
curl -s -X PUT "$API_BASE/users/$USER_ID/risk" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"risk_tolerance": "conservative"}' > /dev/null

CONSERVATIVE_ADVICE=$(curl -s -X GET "$API_BASE/users/$USER_ID/advice" \
  -H "Authorization: Bearer $TOKEN")
echo "$CONSERVATIVE_ADVICE" | jq .

echo "\n🚀 Aggressive Profile:"
curl -s -X PUT "$API_BASE/users/$USER_ID/risk" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"risk_tolerance": "aggressive"}' > /dev/null

AGGRESSIVE_ADVICE=$(curl -s -X GET "$API_BASE/users/$USER_ID/advice" \
  -H "Authorization: Bearer $TOKEN")
echo "$AGGRESSIVE_ADVICE" | jq .

echo "\n🎉 Demo completed!"
echo "\n💡 Key Features Demonstrated:"
echo "   ✅ User creation and authentication"
echo "   ✅ Transaction management"
echo "   ✅ Automatic savings calculation"
echo "   ✅ Risk-based investment advice"
echo "   ✅ Real-time portfolio allocation"
echo "\n🔗 API Documentation: http://localhost:8080/swagger/index.html"