# Validate Application Task Output Debug Guide

## Issue Description

The `validate_application` task in the loan processing workflow is returning null output, which prevents the workflow from proceeding correctly.

## Root Cause Analysis

Based on the code review, the issue could be in several areas:

### 1. Task Handler Registration
- **File**: `infrastructure/workflow/task_worker.go`
- **Issue**: The task worker registers `validate_application_ref` but the workflow uses `validate_application`
- **Status**: ✅ Fixed - Task handler is properly registered

### 2. Task Execution Flow
- **File**: `infrastructure/workflow/loan_processing_tasks.go`
- **Issue**: The `ValidateApplication` method should return proper output
- **Status**: ✅ Verified - Method returns structured output

### 3. Input Parameter Mapping
- **File**: `workflows/loan_processing_workflow.json`
- **Issue**: Input parameters are correctly mapped from workflow input
- **Status**: ✅ Verified - Parameters are properly defined

### 4. Conductor Communication
- **File**: `infrastructure/workflow/conductor_client.go`
- **Issue**: Task status updates might not be properly sent
- **Status**: ✅ Enhanced - Added detailed logging

## Debugging Enhancements Implemented

### 1. Enhanced Task Worker Logging
- Added detailed logging in `executeTask` method
- Added handler type information
- Added input/output validation
- Added nil output protection

### 2. Enhanced ValidateApplication Logging
- Added input parameter extraction logging
- Added validation step logging
- Added output preparation logging
- Added return value logging

### 3. Enhanced Conductor Client Logging
- Added detailed task update logging
- Added payload validation
- Added response body logging
- Added nil output protection

### 4. Enhanced Task Polling Logging
- Added polling parameter logging
- Added response body logging
- Added task count logging
- Added task details logging

## Debugging Steps

### Step 1: Run the Debug Script
```bash
./scripts/test_validate_application.sh
```

This script will:
- Check if the service is running
- Check if Conductor is running
- Test loan application creation
- Check workflow status
- Provide debugging guidance

### Step 2: Check Service Logs
Look for the following log entries:

1. **Task Worker Startup**:
   ```
   Starting task worker
   Task handlers registered
   ```

2. **Task Polling**:
   ```
   Polling for tasks with parameters
   Successfully polled for tasks
   ```

3. **Task Execution**:
   ```
   Executing task
   Found task handler
   Task execution completed successfully
   ```

4. **Task Update**:
   ```
   Updating task with output
   Task updated successfully
   ```

### Step 3: Check Conductor UI
1. Navigate to Conductor UI (typically http://localhost:8080)
2. Look for the `loan_processing_workflow`
3. Check the `validate_application` task status
4. Verify the task output is not null

### Step 4: Check Task Input
Verify that the task receives the correct input parameters:
- `applicationId`
- `userId`
- `loanAmount`
- `loanPurpose`
- `annualIncome`
- `monthlyIncome`
- `requestedTerm`

## Common Issues and Solutions

### Issue 1: Task Handler Not Found
**Symptoms**: Log shows "No handler found for task type"
**Solution**: Verify task handler registration in `registerTaskHandlers()`

### Issue 2: Task Input Missing
**Symptoms**: Log shows empty or missing input parameters
**Solution**: Check workflow input construction in `StartLoanProcessingWorkflow()`

### Issue 3: Task Execution Failure
**Symptoms**: Log shows "Task execution failed"
**Solution**: Check the specific task handler implementation

### Issue 4: Conductor API Failure
**Symptoms**: Log shows "Failed to update task"
**Solution**: Check Conductor server status and API endpoints

## Testing the Fix

### 1. Start the Service
```bash
go run ./cmd/main.go
```

### 2. Start Conductor (if not running)
```bash
docker-compose up -d
```

### 3. Create a Test Application
Use the debug script or create a loan application via the API.

### 4. Monitor the Logs
Look for the enhanced logging output to identify where the issue occurs.

### 5. Check Task Output
Verify that the `validate_application` task now produces proper output instead of null.

## Expected Output Structure

The `validate_application` task should return:

```json
{
  "valid": true,
  "validationErrors": {},
  "normalizedData": {
    "applicationId": "app_123",
    "userId": "user_456",
    "loanAmount": 25000,
    "loanPurpose": "debt_consolidation",
    "annualIncome": 75000,
    "monthlyIncome": 6250,
    "requestedTerm": 36,
    "validatedAt": "2024-01-01T00:00:00Z"
  }
}
```

## Next Steps

1. Run the debug script to identify the current state
2. Check the enhanced logs for specific error messages
3. Verify Conductor server connectivity
4. Test the workflow with a simple application
5. Monitor the task execution flow

## Contact

If the issue persists after following these debugging steps, please provide:
- Service logs with the enhanced logging
- Conductor server logs
- Task execution details from Conductor UI
- Any error messages or stack traces
