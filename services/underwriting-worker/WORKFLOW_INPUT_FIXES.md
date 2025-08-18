# Workflow Input Fixes: "user ID is required and must be a non-empty string"

## Problem Analysis

The error "user ID is required and must be a non-empty string" occurs when the income verification task is executed but the `userId` parameter is missing from the workflow input. This indicates that the workflow is being started without the required input parameters.

## Root Cause

The issue is in the workflow input parameters:

1. **Missing userId parameter**: The workflow is being started without providing the `userId` parameter
2. **Workflow configuration issue**: The workflow definition expects both `applicationId` and `userId` but only one is being provided
3. **Input validation**: The task properly validates input parameters and fails when required parameters are missing

## Error Details

From the log:
```
Task result JSON payload: {
  "taskId": "f9cf3170-7bfe-11f0-a413-1ec36ef668ea",
  "referenceTaskName": "income_verification",
  "workflowInstanceId": "f814fe4c-7bfe-11f0-a413-1ec36ef668ea",
  "status": "FAILED",
  "outputData": {
    "error": "user ID is required and must be a non-empty string",
    "failed_at": "2025-08-18T06:45:49Z",
    "processing_time": "277.709µs",
    "task_type": "income_verification",
    "workflow_id": "f814fe4c-7bfe-11f0-a413-1ec36ef668ea"
  },
  "reasonForIncompletion": "user ID is required and must be a non-empty string",
  "workerId": "worker-3"
}
```

## Solutions Implemented

### 1. **Enhanced Input Validation and Logging**

**File: `infrastructure/workflow/tasks/income_verification_task.go`**

- Added comprehensive input validation with detailed error messages
- Enhanced logging to show all input keys received
- Added type checking for input parameters
- Improved error context for debugging

**Key Improvements:**
```go
// Log all input keys for debugging
inputKeys := make([]string, 0, len(input))
for key := range input {
    inputKeys = append(inputKeys, key)
}
logger.Info("Input keys received", zap.Strings("input_keys", inputKeys))

// Enhanced error messages
if !ok || userID == "" {
    logger.Error("Invalid or missing userId", 
        zap.Any("input", input),
        zap.String("userId_type", fmt.Sprintf("%T", input["userId"])),
        zap.Any("userId_value", input["userId"]),
        zap.String("application_id", applicationID))
    return nil, fmt.Errorf("user ID is required and must be a non-empty string - received: %v", input["userId"])
}
```

### 2. **Debugging Tools**

**File: `debug_workflow_input.go`**

- Comprehensive workflow input testing
- Different input scenarios testing
- Task execution testing with various inputs
- Workflow start testing

**Test Scenarios:**
- Valid input with both parameters
- Missing userId
- Missing applicationId
- Empty parameters
- Nil input
- Wrong data types

## How to Fix the Issue

### Step 1: Run the Diagnostic Tool

```bash
cd underwriting-worker
go run debug_workflow_input.go
```

This will test various input scenarios and help identify the specific issue.

### Step 2: Check Workflow Input

Ensure the workflow is being started with both required parameters:

```bash
# Correct workflow start
curl -X POST http://localhost:8082/api/workflow/underwriting_workflow \
  -H "Content-Type: application/json" \
  -d '{
    "applicationId": "test-app-123",
    "userId": "test-user-123"
  }'
```

### Step 3: Check Workflow Definition

Verify the workflow definition includes both parameters:

```json
{
  "name": "underwriting_workflow",
  "inputParameters": ["applicationId", "userId"],
  "tasks": [
    {
      "name": "income_verification",
      "inputParameters": {
        "applicationId": "${workflow.input.applicationId}",
        "userId": "${workflow.input.userId}"
      }
    }
  ]
}
```

### Step 4: Monitor Task Execution

Look for detailed input validation logs:

```bash
# Check for input validation messages
grep -i "input keys received" logs/underwriting-worker.log
grep -i "invalid or missing" logs/underwriting-worker.log
grep -i "user ID is required" logs/underwriting-worker.log
```

## Expected Behavior

### 1. **Successful Task Execution**
```
INFO    Input keys received    input_keys=["applicationId","userId"]
INFO    Validated input parameters    application_id=test-app-123 user_id=test-user-123
INFO    Income verification completed    application_id=test-app-123
```

### 2. **Missing Parameter Error**
```
INFO    Input keys received    input_keys=["applicationId"]
ERROR   Invalid or missing userId    input={"applicationId":"test-app-123"} userId_type=<nil> userId_value=<nil>
ERROR   user ID is required and must be a non-empty string - received: <nil>
```

### 3. **Empty Parameter Error**
```
INFO    Input keys received    input_keys=["applicationId","userId"]
ERROR   Invalid or missing userId    input={"applicationId":"test-app-123","userId":""} userId_type=string userId_value=""
ERROR   user ID is required and must be a non-empty string - received: ""
```

## Common Issues and Solutions

### 1. **Workflow Started Without userId**

**Issue**: Workflow is started with only `applicationId`
**Solution**: Ensure both parameters are provided
```json
{
  "applicationId": "test-app-123",
  "userId": "test-user-123"  // This was missing
}
```

### 2. **Wrong Parameter Names**

**Issue**: Using different parameter names
**Solution**: Use exact parameter names
```json
{
  "applicationId": "test-app-123",  // Not "application_id"
  "userId": "test-user-123"         // Not "user_id"
}
```

### 3. **Empty String Values**

**Issue**: Parameters provided but with empty values
**Solution**: Ensure non-empty string values
```json
{
  "applicationId": "test-app-123",
  "userId": "test-user-123"  // Not ""
}
```

### 4. **Wrong Data Types**

**Issue**: Parameters provided with wrong data types
**Solution**: Ensure string values
```json
{
  "applicationId": "test-app-123",  // Not 123
  "userId": "test-user-123"         // Not 456
}
```

## Configuration Requirements

### 1. **Workflow Definition**

Ensure workflow definition includes both parameters:

```json
{
  "name": "underwriting_workflow",
  "inputParameters": ["applicationId", "userId"],
  "tasks": [
    {
      "name": "income_verification",
      "inputParameters": {
        "applicationId": "${workflow.input.applicationId}",
        "userId": "${workflow.input.userId}"
      }
    }
  ]
}
```

### 2. **Task Definition**

Ensure task definition includes both input keys:

```json
{
  "name": "income_verification",
  "inputKeys": ["applicationId", "userId"],
  "outputKeys": ["incomeVerification", "incomeAnalysis"]
}
```

## Testing the Fix

### 1. **Unit Tests**

Test task execution with various inputs:
```bash
go run debug_workflow_input.go
```

### 2. **Integration Tests**

Test complete workflow execution:
```bash
# Start underwriting worker
./underwriting-worker-service

# Submit workflow with correct parameters
curl -X POST http://localhost:8082/api/workflow/underwriting_workflow \
  -H "Content-Type: application/json" \
  -d '{
    "applicationId": "test-app-123",
    "userId": "test-user-123"
  }'
```

### 3. **Error Scenario Testing**

Test various error scenarios:
```bash
# Test with missing userId
curl -X POST http://localhost:8082/api/workflow/underwriting_workflow \
  -H "Content-Type: application/json" \
  -d '{
    "applicationId": "test-app-123"
  }'

# Test with empty userId
curl -X POST http://localhost:8082/api/workflow/underwriting_workflow \
  -H "Content-Type: application/json" \
  -d '{
    "applicationId": "test-app-123",
    "userId": ""
  }'
```

## Monitoring and Alerts

### 1. **Key Metrics to Monitor**
- Task execution success rate
- Input validation failures
- Missing parameter errors
- Workflow execution success rate

### 2. **Alert Conditions**
- High input validation failure rates
- Missing userId parameter errors
- Workflow execution failures
- Task execution timeouts

## Troubleshooting Steps

### 1. **Check Workflow Input**
```bash
# Check what parameters are being sent
grep -i "input keys received" logs/underwriting-worker.log
```

### 2. **Check Workflow Definition**
```bash
# Verify workflow definition
curl http://localhost:8082/api/metadata/workflow/underwriting_workflow
```

### 3. **Check Task Definition**
```bash
# Verify task definition
curl http://localhost:8082/api/metadata/taskdefs/income_verification
```

### 4. **Run Diagnostic Tool**
```bash
go run debug_workflow_input.go
```

### 5. **Check Application Code**
Look for where the workflow is being started and ensure both parameters are provided.

## Next Steps

1. **Identify the source** of the workflow start call
2. **Ensure both parameters** are provided when starting the workflow
3. **Update application code** to include userId parameter
4. **Test the fix** with the diagnostic tool
5. **Monitor logs** for successful task execution

## Contact

If the issue persists after implementing these fixes, please provide:
- Complete workflow start request
- Application code that starts the workflow
- Workflow definition from Conductor
- Complete error logs
- Steps to reproduce the issue
