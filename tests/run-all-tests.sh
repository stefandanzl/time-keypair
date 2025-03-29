#!/bin/bash

# Set colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Ensure scripts are executable
echo -e "${YELLOW}Making sure test scripts are executable...${NC}"
chmod +x test-api.sh load-test.sh error-test.sh

# Track overall success
ALL_PASSED=true

# Function to run a test with proper formatting
run_test() {
  local test_script=$1
  local test_name=$2
  
  echo -e "\n${YELLOW}==========================================${NC}"
  echo -e "${YELLOW}Running $test_name...${NC}"
  echo -e "${YELLOW}==========================================${NC}\n"
  
  if bash ./$test_script; then
    echo -e "\n${GREEN}✓ $test_name completed successfully${NC}\n"
    return 0
  else
    echo -e "\n${RED}✗ $test_name failed${NC}\n"
    ALL_PASSED=false
    return 1
  fi
}

# Run unit tests via go test
echo -e "\n${YELLOW}==========================================${NC}"
echo -e "${YELLOW}Running Unit Tests...${NC}"
echo -e "${YELLOW}==========================================${NC}\n"

cd ..
if go test -v ./...; then
  echo -e "\n${GREEN}✓ Unit Tests completed successfully${NC}\n"
else
  echo -e "\n${RED}✗ Unit Tests failed${NC}\n"
  ALL_PASSED=false
fi
cd - > /dev/null

# Run API tests
run_test "test-api.sh" "API Functionality Tests"

# Run load tests
run_test "load-test.sh" "Load Tests"

# Run error handling tests
run_test "error-test.sh" "Error Handling Tests"

# Final summary
echo -e "\n${YELLOW}==========================================${NC}"
if [ "$ALL_PASSED" = true ]; then
  echo -e "${GREEN}All tests passed successfully!${NC}"
else
  echo -e "${RED}Some tests failed. Please check the output above.${NC}"
  exit 1
fi
echo -e "${YELLOW}==========================================${NC}\n"
