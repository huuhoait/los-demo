#!/bin/bash

# Deploy Conductor Workflows and Tasks
# This script deploys all workflow definitions and task definitions to Netflix Conductor

set -e

# Configuration
CONDUCTOR_SERVER=${CONDUCTOR_SERVER:-"http://localhost:8082"}
CONDUCTOR_API_BASE="${CONDUCTOR_SERVER}/api"
WORKFLOWS_DIR="./workflows"
TASKS_DIR="./workflows/tasks"

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

check_conductor_health() {
    print_status "Checking Conductor server health..."
    
    if curl -f -s "${CONDUCTOR_SERVER}/health" > /dev/null; then
        print_success "Conductor server is healthy!"
        return 0
    else
        print_error "Conductor server is not responding!"
        return 1
    fi
}

wait_for_conductor() {
    print_status "Waiting for Conductor server to be ready..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if check_conductor_health; then
            return 0
        fi
        
        print_status "Attempt $attempt/$max_attempts - Waiting 10 seconds..."
        sleep 10
        attempt=$((attempt + 1))
    done
    
    print_error "Conductor server did not become ready within expected time!"
    return 1
}

deploy_tasks() {
    print_status "Deploying task definitions..."
    
    local task_files=(
        "prequalification_tasks.json"
        "loan_processing_tasks.json"
        "underwriting_tasks.json"
    )
    
    for task_file in "${task_files[@]}"; do
        local task_path="${TASKS_DIR}/${task_file}"
        
        if [ -f "$task_path" ]; then
            print_status "Deploying tasks from $task_file..."
            
            if curl -f -s -X POST "${CONDUCTOR_API_BASE}/metadata/taskdefs" \
                -H "Content-Type: application/json" \
                -d @"$task_path" > /dev/null; then
                print_success "Tasks from $task_file deployed successfully!"
            else
                print_error "Failed to deploy tasks from $task_file!"
                return 1
            fi
        else
            print_warning "Task file $task_file not found, skipping..."
        fi
    done
    
    print_success "All task definitions deployed!"
}

deploy_workflows() {
    print_status "Deploying workflow definitions..."
    
    local workflow_files=(
        "prequalification_workflow.json"
        "loan_processing_workflow.json"
        "underwriting_workflow.json"
    )
    
    for workflow_file in "${workflow_files[@]}"; do
        local workflow_path="${WORKFLOWS_DIR}/${workflow_file}"
        
        if [ -f "$workflow_path" ]; then
            print_status "Deploying workflow from $workflow_file..."
            
            if curl -f -s -X PUT "${CONDUCTOR_API_BASE}/metadata/workflow" \
                -H "Content-Type: application/json" \
                -d @"$workflow_path" > /dev/null; then
                print_success "Workflow from $workflow_file deployed successfully!"
            else
                print_error "Failed to deploy workflow from $workflow_file!"
                return 1
            fi
        else
            print_warning "Workflow file $workflow_file not found, skipping..."
        fi
    done
    
    print_success "All workflow definitions deployed!"
}

verify_deployment() {
    print_status "Verifying deployment..."
    
    # Check if tasks are deployed
    local task_count=$(curl -s "${CONDUCTOR_API_BASE}/metadata/taskdefs" | jq 'length' 2>/dev/null || echo "0")
    print_status "Deployed task definitions: $task_count"
    
    # Check if workflows are deployed
    local workflow_count=$(curl -s "${CONDUCTOR_API_BASE}/metadata/workflow" | jq 'length' 2>/dev/null || echo "0")
    print_status "Deployed workflow definitions: $workflow_count"
    
    if [ "$task_count" -gt 0 ] && [ "$workflow_count" -gt 0 ]; then
        print_success "Deployment verification successful!"
        return 0
    else
        print_error "Deployment verification failed!"
        return 1
    fi
}

show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -s, --server URL    Conductor server URL (default: http://localhost:8080)"
    echo "  -h, --help         Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Deploy to localhost:8080"
    echo "  $0 -s http://conductor:8080          # Deploy to Docker container"
    echo "  $0 --server http://prod:8080         # Deploy to production server"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--server)
            CONDUCTOR_SERVER="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    print_status "Starting Conductor deployment..."
    print_status "Target server: $CONDUCTOR_SERVER"
    
    # Wait for Conductor to be ready
    if ! wait_for_conductor; then
        exit 1
    fi
    
    # Deploy tasks first
    if ! deploy_tasks; then
        print_error "Task deployment failed!"
        exit 1
    fi
    
    # Deploy workflows
    if ! deploy_workflows; then
        print_error "Workflow deployment failed!"
        exit 1
    fi
    
    # Verify deployment
    if ! verify_deployment; then
        print_error "Deployment verification failed!"
        exit 1
    fi
    
    print_success "Conductor deployment completed successfully!"
    print_status "You can now access:"
    print_status "  - Conductor Server: $CONDUCTOR_SERVER"
    print_status "  - Conductor UI: ${CONDUCTOR_SERVER%:8080}:3000"
}

# Run main function
main "$@"
