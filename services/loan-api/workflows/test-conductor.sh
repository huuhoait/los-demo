#!/bin/bash

# Test Conductor Integration
# This script tests the Conductor server integration with the loan service

set -e

# Configuration
CONDUCTOR_SERVER="http://localhost:8082"
LOAN_SERVICE="http://localhost:8081"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test Conductor server health
test_conductor_health() {
    print_status "Testing Conductor server health..."
    
    if curl -f -s "${CONDUCTOR_SERVER}/health" > /dev/null; then
        print_success "Conductor server is healthy!"
        return 0
    else
        print_error "Conductor server health check failed!"
        return 1
    fi
}

# Test Conductor API endpoints
test_conductor_api() {
    print_status "Testing Conductor API endpoints..."
    
    # Test metadata endpoint
    if curl -f -s "${CONDUCTOR_SERVER}/api/metadata/workflow" > /dev/null; then
        print_success "Conductor metadata API is accessible!"
    else
        print_warning "Conductor metadata API is not accessible!"
    fi
    
    # Test task definitions endpoint
    if curl -f -s "${CONDUCTOR_SERVER}/api/metadata/taskdefs" > /dev/null; then
        print_success "Conductor task definitions API is accessible!"
    else
        print_warning "Conductor task definitions API is not accessible!"
    fi
}

# Test loan service workflow endpoints
test_loan_service_workflows() {
    print_status "Testing loan service workflow endpoints..."
    
    # Test workflow status endpoint
    if curl -f -s "${LOAN_SERVICE}/v1/workflows/test123/status" > /dev/null; then
        print_success "Loan service workflow status endpoint is working!"
    else
        print_warning "Loan service workflow status endpoint is not working!"
    fi
    
    # Test workflow pause endpoint
    if curl -f -s -X POST "${LOAN_SERVICE}/v1/workflows/test123/pause" > /dev/null; then
        print_success "Loan service workflow pause endpoint is working!"
    else
        print_warning "Loan service workflow pause endpoint is not working!"
    fi
    
    # Test workflow resume endpoint
    if curl -f -s -X POST "${LOAN_SERVICE}/v1/workflows/test123/resume" > /dev/null; then
        print_success "Loan service workflow resume endpoint is working!"
    else
        print_warning "Loan service workflow resume endpoint is not working!"
    fi
    
    # Test workflow terminate endpoint
    if curl -f -s -X POST "${LOAN_SERVICE}/v1/workflows/test123/terminate" \
        -H "Content-Type: application/json" \
        -d '{"reason": "Testing"}' > /dev/null; then
        print_success "Loan service workflow terminate endpoint is working!"
    else
        print_warning "Loan service workflow terminate endpoint is not working!"
    fi
}

# Test workflow execution
test_workflow_execution() {
    print_status "Testing workflow execution..."
    
    # Start a pre-qualification workflow
    print_status "Starting pre-qualification workflow..."
    
    local workflow_response=$(curl -s -X POST "${CONDUCTOR_SERVER}/api/workflow" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "prequalification_workflow",
            "version": 1,
            "input": {
                "userId": "test_user_123",
                "loanAmount": 30000,
                "annualIncome": 60000,
                "monthlyDebt": 500,
                "employmentStatus": "full_time"
            }
        }')
    
    if echo "$workflow_response" | grep -q "workflowId"; then
        local workflow_id=$(echo "$workflow_response" | grep -o '"workflowId":"[^"]*"' | cut -d'"' -f4)
        print_success "Pre-qualification workflow started with ID: $workflow_id"
        
        # Check workflow status
        print_status "Checking workflow status..."
        sleep 5
        
        local status_response=$(curl -s "${CONDUCTOR_SERVER}/api/workflow/${workflow_id}")
        if echo "$status_response" | grep -q "status"; then
            local status=$(echo "$status_response" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
            print_success "Workflow status: $status"
        else
            print_warning "Could not retrieve workflow status"
        fi
    else
        print_warning "Could not start pre-qualification workflow"
        echo "Response: $workflow_response"
    fi
}

# Main test execution
main() {
    print_status "Starting Conductor integration tests..."
    
    # Test Conductor server
    if ! test_conductor_health; then
        print_error "Conductor server tests failed!"
        exit 1
    fi
    
    # Test Conductor API
    test_conductor_api
    
    # Test loan service workflow endpoints
    test_loan_service_workflows
    
    # Test workflow execution
    test_workflow_execution
    
    print_success "All Conductor integration tests completed!"
    print_status "You can now:"
    print_status "  - Access Conductor UI at: http://localhost:3000"
    print_status "  - Use workflow endpoints at: http://localhost:8081/v1/workflows"
    print_status "  - Monitor workflows at: http://localhost:8080"
}

# Run main function
main "$@"
