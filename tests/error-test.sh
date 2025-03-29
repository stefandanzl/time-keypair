#!/bin/bash

# Set variables
SERVER="http://localhost:8080"
SUPER_KEY="super_admin_key"
INVALID_KEY="invalid_key"
TEST_USER="error_test_user"

echo "Error Handling Test for Multi-User Cron Server API"
echo "=================================================="

# Setup - Create a test user
echo -e "\nCreating test user..."
curl -s -X POST "${SERVER}/admin/${SUPER_KEY}/users" -d "{\"user\":\"${TEST_USER}\"}"

# Test Case 1: Invalid super admin key
echo -e "\n\n1. Testing invalid super admin key..."
response=$(curl -s -w "%{http_code}" -X GET "${SERVER}/admin/${INVALID_KEY}/users")
status_code=${response: -3}
content=${response:0:${#response}-3}
echo "Status code: $status_code"
echo "Response: $content"

# Test Case 2: Invalid user key
echo -e "\n\n2. Testing invalid user key..."
response=$(curl -s -w "%{http_code}" -X GET "${SERVER}/status/${INVALID_KEY}")
status_code=${response: -3}
content=${response:0:${#response}-3}
echo "Status code: $status_code"
echo "Response: $content"

# Test Case 3: Non-existent job
echo -e "\n\n3. Testing non-existent job..."
response=$(curl -s -w "%{http_code}" -X GET "${SERVER}/cron/${TEST_USER}/job/non_existent_job")
status_code=${response: -3}
content=${response:0:${#response}-3}
echo "Status code: $status_code"
echo "Response: $content"

# Test Case 4: Invalid cron expression
echo -e "\n\n4. Testing invalid cron expression..."
response=$(curl -s -w "%{http_code}" -X POST "${SERVER}/cron/${TEST_USER}/jobs" \
  -d "{\"id\":\"invalid_job\",\"cron\":\"invalid cron\",\"url\":\"https://example.com\",\"active\":true}")
status_code=${response: -3}
content=${response:0:${#response}-3}
echo "Status code: $status_code"
echo "Response: $content"

# Test Case 5: Missing required fields in job
echo -e "\n\n5. Testing missing required fields in job..."
response=$(curl -s -w "%{http_code}" -X POST "${SERVER}/cron/${TEST_USER}/jobs" \
  -d "{\"id\":\"missing_fields_job\"}")
status_code=${response: -3}
content=${response:0:${#response}-3}
echo "Status code: $status_code"
echo "Response: $content"

# Test Case 6: Accessing non-existent data
echo -e "\n\n6. Testing access to non-existent data..."
response=$(curl -s -w "%{http_code}" -X GET "${SERVER}/data/${TEST_USER}/non_existent_data")
status_code=${response: -3}
content=${response:0:${#response}-3}
echo "Status code: $status_code"
echo "Response: $content"

# Test Case 7: Invalid JSON in request body
echo -e "\n\n7. Testing invalid JSON in request body..."
response=$(curl -s -w "%{http_code}" -X PUT "${SERVER}/data/${TEST_USER}/invalid_json" \
  -d "{invalid json}")
status_code=${response: -3}
content=${response:0:${#response}-3}
echo "Status code: $status_code"
echo "Response: $content"

# Test Case 8: Deleting non-existent user
echo -e "\n\n8. Testing deletion of non-existent user..."
response=$(curl -s -w "%{http_code}" -X DELETE "${SERVER}/admin/${SUPER_KEY}/users/non_existent_user")
status_code=${response: -3}
content=${response:0:${#response}-3}
echo "Status code: $status_code"
echo "Response: $content"

# Test Case 9: Deleting non-existent job
echo -e "\n\n9. Testing deletion of non-existent job..."
response=$(curl -s -w "%{http_code}" -X DELETE "${SERVER}/cron/${TEST_USER}/job/non_existent_job")
status_code=${response: -3}
content=${response:0:${#response}-3}
echo "Status code: $status_code"
echo "Response: $content"

# Test Case 10: Method not allowed
echo -e "\n\n10. Testing method not allowed..."
response=$(curl -s -w "%{http_code}" -X DELETE "${SERVER}/data/${TEST_USER}/keys")
status_code=${response: -3}
content=${response:0:${#response}-3}
echo "Status code: $status_code"
echo "Response: $content"

# Cleanup - Delete test user
echo -e "\n\nCleaning up - Deleting test user..."
curl -s -X DELETE "${SERVER}/admin/${SUPER_KEY}/users/${TEST_USER}"

echo -e "\n\nError handling tests completed!"
