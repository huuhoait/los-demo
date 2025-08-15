#!/bin/bash

# Test script for debugging the validate_application task output issue
# This script helps identify why the task output is null

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

# Test data for loan application
APPLICATION_ID="test_app_123"
USER_ID="test_user_456"
LOAN_AMOUNT=25000
LOAN_PURPOSE="debt_consolidation"
ANNUAL_INCOME=75000
MONTHLY_INCOME=6250
REQUESTED_TERM=36

echo -e "${BLUE}=== Validate Application Task Debug Test ===${NC}"
echo "Service URL: ${SERVICE_URL}"
echo "Application ID: ${APPLICATION_ID}"
echo "User ID: ${USER_ID}"
echo "Loan Amount: \$${LOAN_AMOUNT}"
echo "Loan Purpose: ${LOAN_PURPOSE}"
echo "Annual Income: \$${ANNUAL_INCOME}"
echo "Monthly Income: \$${MONTHLY_INCOME}"
echo "Requested Term: ${REQUESTED_TERM} months"
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

# Function to test loan application creation
test_create_application() {
    echo -e "${YELLOW}Testing loan application creation...${NC}"
    
    # Create test payload
    PAYLOAD=$(cat <<EOF
{
    "user": {
        "email": "test@example.com",
        "firstName": "John",
        "lastName": "Doe",
        "phone": "+1234567890",
        "dateOfBirth": "1990-01-01"
    },
    "loanAmount": ${LOAN_AMOUNT},
    "loanPurpose": "${LOAN_PURPOSE}",
    "annualIncome": ${ANNUAL_INCOME},
    "monthlyIncome": ${MONTHLY_INCOME},
    "monthlyDebt": 1500,
    "requestedTerm": ${REQUESTED_TERM},
    "employmentStatus": "full_time"
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
        "${SERVICE_URL}/loans/applications")
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Loan application creation successful${NC}"
        echo "Response:"
        echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
        
        # Extract application ID if available
        if echo "$RESPONSE" | grep -q "id"; then
            EXTRACTED_ID=$(echo "$RESPONSE" | jq -r '.id' 2>/dev/null || echo "unknown")
            echo ""
            echo -e "${BLUE}Extracted Application ID: ${EXTRACTED_ID}${NC}"
        fi
    else
        echo -e "${RED}✗ Loan application creation failed${NC}"
        echo "Response: $RESPONSE"
    fi
    echo ""
}

# Function to check workflow status
check_workflow_status() {
    echo -e "${YELLOW}Checking workflow status...${NC}"
    
    # This would typically be called with a real workflow ID
    # For now, we'll just test the endpoint structure
    WORKFLOW_ID="test_workflow_123"
    
    RESPONSE=$(curl -s -X GET \
        -H "Authorization: Bearer test_token" \
        "${SERVICE_URL}/workflows/${WORKFLOW_ID}/status")
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Workflow status check successful${NC}"
        echo "Response:"
        echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    else
        echo -e "${RED}✗ Workflow status check failed${NC}"
        echo "Response: $RESPONSE"
    fi
    echo ""
}

# Function to check Conductor server status
check_conductor() {
    echo -e "${YELLOW}Checking Conductor server status...${NC}"
    
    CONDUCTOR_URL="http://localhost:8080"
    
    # Check if Conductor is running
    if curl -s "${CONDUCTOR_URL}/api/admin/config" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Conductor server is running${NC}"
        
        # Get Conductor configuration
        CONFIG_RESPONSE=$(curl -s "${CONDUCTOR_URL}/api/admin/config")
        echo "Conductor config:"
        echo "$CONFIG_RESPONSE" | jq '.' 2>/dev/null || echo "$CONFIG_RESPONSE"
    else
        echo -e "${RED}✗ Conductor server is not running${NC}"
        echo "Please start Conductor first:"
        echo "  docker-compose up -d"
    fi
    echo ""
}

# Function to check task worker logs
check_task_worker() {
    echo -e "${YELLOW}Checking task worker status...${NC}"
    
    # This would check if the task worker is running and processing tasks
    echo "To check task worker logs, look for:"
    echo "  - Task execution logs in the service logs"
    echo "  - Conductor task status updates"
    echo "  - Task output in Conductor UI"
    echo ""
}

# Main execution
main() {
    echo -e "${BLUE}Starting validate_application task debugging...${NC}"
    echo ""
    
    # Check service status
    if ! check_service; then
        exit 1
    fi
    
    # Check Conductor status
    check_conductor
    
    # Test application creation
    test_create_application
    
    # Check workflow status
    check_workflow_status
    
    # Check task worker
    check_task_worker
    
    echo -e "${BLUE}=== Debugging Complete ===${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Check the service logs for task execution details"
    echo "2. Verify the task worker is running and polling for tasks"
    echo "3. Check Conductor UI for task status and output"
    echo "4. Look for any error messages in the logs"
    echo ""
    echo "Common issues to check:"
    echo "- Task handler registration in task_worker.go"
    echo "- Input parameter mapping in workflow definition"
    echo "- Task execution flow in loan_processing_tasks.go"
    echo "- Conductor API communication in conductor_client.go"
}

# Run main function
main "$@"
