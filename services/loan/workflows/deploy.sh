#!/bin/bash

# Netflix Conductor Workflow Deployment Script
# This script deploys all loan service workflows and task definitions to Conductor

set -e

# Configuration
CONDUCTOR_SERVER="${CONDUCTOR_SERVER:-http://localhost:8082}"
WORKFLOW_DIR="$(dirname "$0")"
TASKS_DIR="${WORKFLOW_DIR}/tasks"

echo "🚀 Deploying Loan Service Workflows to Conductor Server: $CONDUCTOR_SERVER"

# Function to deploy task definitions
deploy_tasks() {
    local task_file=$1
    local task_name=$(basename "$task_file" .json)
    
    echo "📋 Deploying task definitions from: $task_name"
    
    curl -X POST \
        "$CONDUCTOR_SERVER/api/metadata/taskdefs" \
        -H "Content-Type: application/json" \
        -d @"$task_file" \
        --fail --silent --show-error
        
    echo "✅ Task definitions deployed: $task_name"
}

# Function to deploy workflow definitions
deploy_workflow() {
    local workflow_file=$1
    local workflow_name=$(basename "$workflow_file" .json)
    
    echo "🔄 Deploying workflow: $workflow_name"
    
    # Try to deploy the workflow
    response=$(curl -s -X POST \
        "$CONDUCTOR_SERVER/api/metadata/workflow" \
        -H "Content-Type: application/json" \
        -d @"$workflow_file" \
        -w "%{http_code}")
    
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        echo "✅ Workflow deployed: $workflow_name"
    elif [ "$http_code" = "409" ]; then
        echo "ℹ️  Workflow already exists: $workflow_name (skipping)"
    else
        echo "❌ Failed to deploy workflow: $workflow_name (HTTP $http_code)"
        return 1
    fi
}

# Deploy task definitions first
echo "📋 Deploying Task Definitions..."
for task_file in "$TASKS_DIR"/*.json; do
    if [ -f "$task_file" ]; then
        deploy_tasks "$task_file"
    fi
done

echo ""

# Deploy workflow definitions
echo "🔄 Deploying Workflow Definitions..."
for workflow_file in "$WORKFLOW_DIR"/*.json; do
    if [ -f "$workflow_file" ] && [[ ! "$workflow_file" =~ /tasks/ ]]; then
        deploy_workflow "$workflow_file"
    fi
done

echo ""
echo "🎉 All workflows and tasks deployed successfully!"
echo ""
echo "📊 Deployed Components:"
echo "   • Pre-qualification Workflow (prequalification_workflow v1)"
echo "   • Loan Processing Workflow (loan_processing_workflow v1)" 
echo "   • Underwriting Workflow (underwriting_workflow v1)"
echo "   • Pre-qualification Tasks (5 tasks)"
echo "   • Loan Processing Tasks (5 tasks)"
echo "   • Underwriting Tasks (10 tasks)"
echo ""
echo "🔗 Access Conductor UI: $CONDUCTOR_SERVER"
echo "📖 View workflows at: $CONDUCTOR_SERVER/workflow"

# Verify deployment
echo ""
echo "🔍 Verifying deployment..."

# Check if workflows exist
for workflow in "prequalification_workflow" "loan_processing_workflow" "underwriting_workflow"; do
    response=$(curl -s "$CONDUCTOR_SERVER/api/metadata/workflow/$workflow" || echo "ERROR")
    if [[ "$response" == "ERROR" ]] || [[ "$response" == *"NOT_FOUND"* ]]; then
        echo "❌ Failed to verify workflow: $workflow"
    else
        echo "✅ Verified workflow: $workflow"
    fi
done

echo ""
echo "🏁 Deployment complete!"
