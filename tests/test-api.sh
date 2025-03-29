#!/bin/bash

# Set variables
SERVER="http://localhost:8080"
SUPER_KEY="super_admin_key"
USER="test_user"

echo "Testing Multi-User Cron Server API"
echo "=================================="

# Test health endpoint
echo -e "\nTesting health endpoint..."
curl -s "${SERVER}/health"

# Create a new user
echo -e "\n\nCreating a new user: ${USER}..."
curl -s -X POST "${SERVER}/admin/${SUPER_KEY}/users" -d "{\"user\":\"${USER}\"}"

# Get all users
echo -e "\n\nGetting all users..."
curl -s "${SERVER}/admin/${SUPER_KEY}/users" | jq

# Create a new cron job
echo -e "\n\nCreating a new cron job..."
curl -s -X POST "${SERVER}/cron/${USER}/jobs" \
  -d "{\"id\":\"job1\",\"cron\":\"0 */5 * * * *\",\"url\":\"https://example.com\",\"active\":true}"

# Get all jobs
echo -e "\n\nGetting all jobs for user ${USER}..."
curl -s "${SERVER}/cron/${USER}/jobs" | jq

# Store data
echo -e "\n\nStoring data..."
curl -s -X PUT "${SERVER}/data/${USER}/settings" \
  -d "{\"theme\":\"dark\",\"notifications\":true}"

# Get data
echo -e "\n\nGetting data..."
curl -s "${SERVER}/data/${USER}/settings" | jq

# Get job status
echo -e "\n\nGetting job status..."
curl -s "${SERVER}/status/${USER}" | jq

# Create another job that's inactive
echo -e "\n\nCreating an inactive job..."
curl -s -X POST "${SERVER}/cron/${USER}/jobs" \
  -d "{\"id\":\"job2\",\"cron\":\"0 0 * * * *\",\"url\":\"https://example.com/backup\",\"active\":false}"

# Update an existing job
echo -e "\n\nUpdating existing job..."
curl -s -X PUT "${SERVER}/cron/${USER}/job/job1" \
  -d "{\"cron\":\"0 */10 * * * *\",\"url\":\"https://example.com/updated\",\"active\":true}"

# Get specific job after update
echo -e "\n\nGetting specific job after update..."
curl -s "${SERVER}/cron/${USER}/job/job1" | jq

# Store nested data
echo -e "\n\nStoring nested data..."
curl -s -X PUT "${SERVER}/data/${USER}/user_profile" \
  -d "{\"name\":\"John Doe\",\"contact\":{\"email\":\"john@example.com\",\"phone\":\"555-1234\"},\"preferences\":{\"language\":\"en\",\"timezone\":\"UTC\"}}"

# Get nested data
echo -e "\n\nGetting nested data..."
curl -s "${SERVER}/data/${USER}/user_profile" | jq

# List data keys
echo -e "\n\nListing data keys..."
curl -s "${SERVER}/data/${USER}/keys" | jq

# Delete a job
echo -e "\n\nDeleting a job..."
curl -s -X DELETE "${SERVER}/cron/${USER}/job/job2"

# Create another user
echo -e "\n\nCreating another user..."
curl -s -X POST "${SERVER}/admin/${SUPER_KEY}/users" -d "{\"user\":\"user2\"}"

# Get all users (should show both users)
echo -e "\n\nGetting all users (should show both)..."
curl -s "${SERVER}/admin/${SUPER_KEY}/users" | jq

# Test job activation/deactivation
echo -e "\n\nTesting job activation/deactivation..."

# First deactivate the job
echo "Deactivating job..."
curl -s "${SERVER}/cron/${USER}/job/job1/off" | jq

# Wait a moment
sleep 1

# Check job status (should be inactive)
echo -e "\nChecking job status after deactivation..."
jobStatus=$(curl -s "${SERVER}/cron/${USER}/job/job1" | jq -r '.active')
echo "Job active status: $jobStatus"

# Now activate the job
echo -e "\nActivating job..."
curl -s "${SERVER}/cron/${USER}/job/job1/on" | jq

# Wait a moment
sleep 1

# Check job status again (should be active)
echo -e "\nChecking job status after activation..."
jobStatus=$(curl -s "${SERVER}/cron/${USER}/job/job1" | jq -r '.active')
echo "Job active status: $jobStatus"

# Get full configuration
echo -e "\n\nGetting full configuration..."
curl -s "${SERVER}/admin/${SUPER_KEY}/config" | jq

# Delete a user
echo -e "\n\nDeleting a user..."
curl -s -X DELETE "${SERVER}/admin/${SUPER_KEY}/users/user2"

# Get all users after deletion
echo -e "\n\nGetting all users after deletion..."
curl -s "${SERVER}/admin/${SUPER_KEY}/users" | jq

echo -e "\n\nTests completed!"
