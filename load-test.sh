#!/bin/bash

# Set variables
SERVER="http://localhost:8080"
SUPER_KEY="super_admin_key"
TOTAL_USERS=5
JOBS_PER_USER=10
DATA_ITEMS_PER_USER=10

echo "Load Testing Multi-User Cron Server API"
echo "======================================="

# Create multiple users
echo -e "\nCreating $TOTAL_USERS users..."
for i in $(seq 1 $TOTAL_USERS); do
  user="load_user_$i"
  curl -s -X POST "${SERVER}/admin/${SUPER_KEY}/users" -d "{\"user\":\"$user\"}" &
done
wait

# Get all users
echo -e "\nVerifying all users were created..."
curl -s "${SERVER}/admin/${SUPER_KEY}/users" | jq

# Create multiple jobs for each user
echo -e "\nCreating $JOBS_PER_USER jobs for each user (total: $((TOTAL_USERS * JOBS_PER_USER)))..."
for i in $(seq 1 $TOTAL_USERS); do
  user="load_user_$i"
  for j in $(seq 1 $JOBS_PER_USER); do
    # Alternate between active and inactive jobs
    active="true"
    if [ $((j % 2)) -eq 0 ]; then
      active="false"
    fi
    
    # Use different cron schedules
    minute=$((j * 5 % 60))
    curl -s -X POST "${SERVER}/cron/${user}/jobs" \
      -d "{\"id\":\"job_${j}\",\"cron\":\"0 ${minute} * * * *\",\"url\":\"https://example.com/${user}/job_${j}\",\"active\":${active}}" &
  done
done
wait

# Store multiple data items for each user
echo -e "\nStoring $DATA_ITEMS_PER_USER data items for each user (total: $((TOTAL_USERS * DATA_ITEMS_PER_USER)))..."
for i in $(seq 1 $TOTAL_USERS); do
  user="load_user_$i"
  for j in $(seq 1 $DATA_ITEMS_PER_USER); do
    curl -s -X PUT "${SERVER}/data/${user}/data_key_${j}" \
      -d "{\"value\":\"data_value_${j}\",\"timestamp\":\"$(date +%s)\",\"metadata\":{\"source\":\"load_test\",\"index\":${j}}}" &
  done
done
wait

# Test concurrent updates to a single job
echo -e "\nTesting concurrent updates to a single job..."
user="load_user_1"
job_id="job_1"
for i in $(seq 1 10); do
  updated_url="https://example.com/updated_${i}_$(date +%s)"
  curl -s -X PUT "${SERVER}/cron/${user}/job/${job_id}" \
    -d "{\"cron\":\"0 ${i} * * * *\",\"url\":\"${updated_url}\",\"active\":true}" &
done
wait

# Check the final state of the job
echo -e "\nChecking final state of the concurrently updated job..."
curl -s "${SERVER}/cron/${user}/job/${job_id}" | jq

# Test concurrent updates to user data
echo -e "\nTesting concurrent updates to user data..."
for i in $(seq 1 10); do
  curl -s -X PUT "${SERVER}/data/${user}/concurrent_test" \
    -d "{\"counter\":${i},\"timestamp\":\"$(date +%s)\"}" &
done
wait

# Check final state of user data
echo -e "\nChecking final state of concurrently updated data..."
curl -s "${SERVER}/data/${user}/concurrent_test" | jq

# Get job status
echo -e "\nGetting job status for a user..."
curl -s "${SERVER}/status/${user}" | jq

# Verify data keys
echo -e "\nVerifying data keys for a user..."
curl -s "${SERVER}/data/${user}/keys" | jq

# Get configuration
echo -e "\nGetting full configuration..."
curl -s "${SERVER}/admin/${SUPER_KEY}/config" | jq | head -20
echo "... (truncated for brevity)"

# Cleanup - Delete all load test users
echo -e "\nCleaning up - Deleting all load test users..."
for i in $(seq 1 $TOTAL_USERS); do
  user="load_user_$i"
  curl -s -X DELETE "${SERVER}/admin/${SUPER_KEY}/users/${user}" &
done
wait

echo -e "\nLoad test completed!"
