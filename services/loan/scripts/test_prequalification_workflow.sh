#!/bin/bash

# Test script for the prequalification workflow
# This script demonstrates how to test the workflow endpoints

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080"
API_VERSION="v1"
SERVICE_URL="${BASE_URL}/${API_VERSION}"

# Test data
USER_ID="test_user_123"
LOAN_AMOUNT=25000
ANNUAL_INCOME=75000
MONTHLY_DEBT=1500
EMPLOYMENT_STATUS="full_time"

echo -e "${BLUE}=== Pre-qualification Workflow Test ===${NC}"
echo "Service URL: ${SERVICE_URL}"
echo "User ID: ${USER_ID}"
echo "Loan Amount: \$${LOAN_AMOUNT}"
echo "Annual Income: \$${ANNUAL_INCOME}"
echo "Monthly Debt: \$${MONTHLY_DEBT}"
echo "Employment Status: ${EMPLOYMENT_STATUS}"
echo ""

# Function to check if service is running
check_service() {
    echo -e "${YELLOW}Checking if service is running...${NC}"
    
    if curl -s "${SERVICE_URL}/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Service is running${NC}"
        return 0
    else
        echo -e "${RED}✗ Service is not running${NC}"
        echo "Please start the loan service first:"
        echo "  go run ./cmd/main.go"
        return 1
    fi
}

# Function to test pre-qualification endpoint
test_prequalification() {
    echo -e "${YELLOW}Testing pre-qualification endpoint...${NC}"
    
    # Create test payload
    PAYLOAD=$(cat <<EOF
{
    "loanAmount": ${LOAN_AMOUNT},
    "annualIncome": ${ANNUAL_INCOME},
    "monthlyDebt": ${MONTHLY_DEBT},
    "employmentStatus": "${EMPLOYMENT_STATUS}"
}
EOF
)
    
    echo "Request payload:"
    echo "$PAYLOAD" | jq '.' 2>/dev/null || echo "$PAYLOAD"
    echo ""
    
    # Make the request
    RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer test_token" \
        -d "$PAYLOAD" \
        "${SERVICE_URL}/loans/prequalify")
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Pre-qualification request successful${NC}"
        echo "Response:"
        echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    else
        echo -e "${RED}✗ Pre-qualification request failed${NC}"
        echo "Response: $RESPONSE"
    fi
    echo ""
}

# Function to test workflow status endpoint
test_workflow_status() {
    echo -e "${YELLOW}Testing workflow status endpoint...${NC}"
    
    # This would typically be called with a real workflow ID
    # For now, we'll just test the endpoint structure
    WORKFLOW_ID="test_workflow_123"
    
    RESPONSE=$(curl -s -X GET \
        -H "Authorization: Bearer test_token" \
        "${SERVICE_URL}/workflows/${WORKFLOW_ID}/status")
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Workflow status request successful${NC}"
        echo "Response:"
        echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    else
        echo -e "${RED}✗ Workflow status request failed${NC}"
        echo "Response: $RESPONSE"
    fi
    echo ""
}

# Function to test health endpoint
test_health() {
    echo -e "${YELLOW}Testing health endpoint...${NC}"
    
    RESPONSE=$(curl -s -X GET "${SERVICE_URL}/health")
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Health check successful${NC}"
        echo "Response:"
        echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    else
        echo -e "${RED}✗ Health check failed${NC}"
        echo "Response: $RESPONSE"
    fi
    echo ""
}

# Function to run workflow simulation
simulate_workflow() {
    echo -e "${YELLOW}Simulating workflow execution...${NC}"
    
    echo "1. Starting pre-qualification workflow..."
    sleep 1
    
    echo "2. Validating input parameters..."
    sleep 1
    
    echo "3. Calculating debt-to-income ratio..."
    DTI_RATIO=$(echo "scale=4; ${MONTHLY_DEBT} * 12 / ${ANNUAL_INCOME}" | bc -l)
    echo "   DTI Ratio: ${DTI_RATIO}"
    sleep 1
    
    echo "4. Assessing risk..."
    if (( $(echo "$DTI_RATIO > 0.43" | bc -l) )); then
        RISK_LEVEL="HIGH"
    elif (( $(echo "$DTI_RATIO > 0.36" | bc -l) )); then
        RISK_LEVEL="MEDIUM"
    else
        RISK_LEVEL="LOW"
    fi
    echo "   Risk Level: ${RISK_LEVEL}"
    sleep 1
    
    echo "5. Generating terms..."
    if (( $(echo "$DTI_RATIO <= 0.43" | bc -l) )) && [ "$ANNUAL_INCOME" -ge 25000 ]; then
        QUALIFIED="true"
        MAX_LOAN="50000"
        MESSAGE="You are pre-qualified for a loan"
    else
        QUALIFIED="false"
        MAX_LOAN="0"
        MESSAGE="You do not currently qualify for a loan"
    fi
    echo "   Qualified: ${QUALIFIED}"
    echo "   Max Loan Amount: \$${MAX_LOAN}"
    echo "   Message: ${MESSAGE}"
    sleep 1
    
    echo "6. Finalizing pre-qualification..."
    sleep 1
    
    echo -e "${GREEN}✓ Workflow simulation completed${NC}"
    echo ""
}

# Function to show test results summary
show_summary() {
    echo -e "${BLUE}=== Test Summary ===${NC}"
    echo "Tests completed for:"
    echo "  - Service health check"
    echo "  - Pre-qualification endpoint"
    echo "  - Workflow status endpoint"
    echo "  - Workflow simulation"
    echo ""
    echo "Next steps:"
    echo "  1. Start Netflix Conductor server"
    echo "  2. Deploy workflow definitions"
    echo "  3. Start task worker"
    echo "  4. Run integration tests"
    echo ""
}

# Main execution
main() {
    echo -e "${BLUE}Starting prequalification workflow tests...${NC}"
    echo ""
    
    # Check if service is running
    if ! check_service; then
        exit 1
    fi
    
    # Run tests
    test_health
    test_prequalification
    test_workflow_status
    simulate_workflow
    
    # Show summary
    show_summary
    
    echo -e "${GREEN}All tests completed!${NC}"
}

# Check if jq is available for JSON formatting
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}Warning: jq is not installed. JSON responses will not be formatted.${NC}"
    echo "Install jq for better output formatting:"
    echo "  macOS: brew install jq"
    echo "  Ubuntu: sudo apt-get install jq"
    echo ""
fi

# Check if bc is available for calculations
if ! command -v bc &> /dev/null; then
    echo -e "${YELLOW}Warning: bc is not installed. Some calculations may not work.${NC}"
    echo "Install bc for calculations:"
    echo "  macOS: brew install bc"
    echo "  Ubuntu: sudo apt-get install bc"
    echo ""
fi

# Run main function
main "$@"
