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
  -d "{\"id\":\"job1\",\"cron\":\"*/5 * * * *\",\"url\":\"https://example.com\",\"active\":true}"

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

# Get configuration
echo -e "\n\nGetting full configuration..."
curl -s "${SERVER}/admin/${SUPER_KEY}/config" | jq

echo -e "\n\nTests completed!"
