#!/bin/bash

BASE_URL="http://localhost:8080"

echo "1. Testing health endpoint:"
curl -s "$BASE_URL/health"
echo "" && echo ""

echo "2. Creating team:"
curl -s -X POST "$BASE_URL/team/add" \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true}
    ]
  }'
echo "" && echo ""

echo "3. Getting team:"
curl -s "$BASE_URL/team/get?team_name=backend"
echo "" && echo ""

echo "4. Creating PR:"
curl -s -X POST "$BASE_URL/pullRequest/create" \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1",
    "pull_request_name": "Add new feature",
    "author_id": "u1"
  }'
echo "" && echo ""

echo "5. Getting reviews for u2:"
curl -s "$BASE_URL/users/getReview?user_id=u2"
echo "" && echo ""

echo "6. Merging PR:"
curl -s -X POST "$BASE_URL/pullRequest/merge" \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1"
  }'
echo "" && echo ""

echo "7. Deactivating user:"
curl -s -X POST "$BASE_URL/users/setIsActive" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "u2",
    "is_active": false
  }'
echo "" && echo ""
