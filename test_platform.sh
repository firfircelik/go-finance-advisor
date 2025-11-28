#!/bin/bash

# Create Demo Account
echo "Creating Demo Account..."
RESPONSE=$(curl -s -X POST http://localhost:8082/api/auth/demo)
TOKEN=$(echo $RESPONSE | jq -r '.data.token')

if [ "$TOKEN" == "null" ]; then
    echo "Failed to create demo account"
    echo $RESPONSE
    exit 1
fi

echo "Token: $TOKEN"

# Verify Overview
echo "Verifying Overview..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8082/api/overview | jq .

# Verify Holdings
echo "Verifying Holdings..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8082/api/holdings | jq .

echo "Platform Verification Completed"
