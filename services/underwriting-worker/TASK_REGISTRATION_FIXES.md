# Task Registration Fixes: "unknown task type: credit_check"

## Problem Analysis

The error "unknown task type: credit_check" occurs when the underwriting worker tries to poll for tasks from Conductor, but Conductor doesn't recognize the task type because the task definitions haven't been properly registered.

## Root Cause

The issue is in the task registration process:

1. **Task definitions not registered**: The task definitions need to be registered with Conductor before workers can poll for them
2. **Registration timing**: Task definitions must be registered before starting the polling process
3. **Conductor connection issues**: If Conductor is not accessible, task registration fails
4. **Missing fallback mechanism**: No graceful handling when task registration fails

## Solutions Implemented

### 1. **Enhanced Task Registration Process**

**File: `infrastructure/workflow/tasks/underwriting_task_worker.go`**

- Added comprehensive task definition registration
- Implemented fallback mechanism for registration failures
- Enhanced error handling and logging
- Added registration validation

**Key Improvements:**
```go
// Register workflow and task definitions with real Conductor
if !w.useMockConductor {
    w.logger.Info("Registering task definitions with Conductor")
    if err := w.registerWorkflowDefinitions(); err != nil {
        w.logger.Error("Failed to register workflow definitions", zap.Error(err))
        // Try to register just the task definitions as a fallback
        if err := w.registerTaskDefinitionsOnly(); err != nil {
            w.logger.Error("Failed to register task definitions as fallback", zap.Error(err))
            return fmt.Errorf("failed to register task definitions: %w", err)
        }
    }
}
```

### 2. **Fallback Task Registration**

**New method: `registerTaskDefinitionsOnly`**

- Registers only task definitions when workflow registration fails
- Provides detailed registration status reporting
- Ensures minimum required tasks are registered
- Adds propagation delay for Conductor

**Features:**
- Individual task registration with error tracking
- Registration success/failure reporting
- Minimum registration threshold validation
- Propagation delay for Conductor

### 3. **Enhanced Error Handling**

**Improved error messages and logging:**

- Detailed error messages for each registration step
- Registration success/failure counts
- Individual task registration status
- Connection and timeout handling

### 4. **Debugging Tools**

**File: `debug_task_registration.go`**

- Comprehensive Conductor connection testing
- Task definition creation and registration testing
- Workflow definition testing
- Polling mechanism testing

**Test Scenarios:**
- HTTP Conductor client creation
- Task definition creation
- Task definition registration
- Workflow definition registration
- Task polling testing
- Workflow execution testing

## How to Fix the Issue

### Step 1: Run the Diagnostic Tool

```bash
cd underwriting-worker
go run debug_task_registration.go
```

This will test the entire task registration process and identify specific issues.

### Step 2: Check Conductor Status

Ensure Conductor is running and accessible:

```bash
# Check if Conductor is running
curl http://localhost:8082/health

# Check if Conductor API is accessible
curl http://localhost:8082/api/metadata/taskdefs
```

### Step 3: Start the Underwriting Worker

```bash
./underwriting-worker-service
```

The worker now has enhanced error handling and will provide detailed logs for task registration.

### Step 4: Monitor Registration Logs

Look for registration success/failure messages:

```bash
# Check for task registration patterns
grep -i "registered task definition" logs/underwriting-worker.log
grep -i "failed to register" logs/underwriting-worker.log
grep -i "task registration summary" logs/underwriting-worker.log
```

## Expected Behavior

### 1. **Successful Registration**
```
INFO    Registered task definition    task_name=credit_check
INFO    Registered task definition    task_name=income_verification
INFO    Task registration summary     successful=5 total=5 failed=0
INFO    Successfully registered workflow definition    name=underwriting_workflow
```

### 2. **Partial Registration (Fallback)**
```
WARN    Very few task definitions registered successfully, this may cause issues
ERROR   Failed to register task definition    task_name=credit_check error=connection refused
INFO    Task registration summary     successful=3 total=5 failed=2
```

### 3. **Registration Failure**
```
ERROR   Failed to register workflow definitions    error=connection refused
ERROR   Failed to register task definitions as fallback    error=connection refused
ERROR   failed to register task definitions: connection refused
```

## Common Issues and Solutions

### 1. **Conductor Not Running**

**Issue**: `connection refused` or `connection timeout`
**Solution**: Start Conductor server
```bash
# Start Conductor (example with Docker)
docker run -p 8082:8082 conductor:latest
```

### 2. **Wrong Conductor URL**

**Issue**: `invalid conductor URL` or `404 not found`
**Solution**: Check configuration
```yaml
# config/config.yaml
conductor:
  base_url: "http://localhost:8082"  # Verify this URL
```

### 3. **Task Definition Already Exists**

**Issue**: `409 conflict` during registration
**Solution**: This is normal - task already registered
```
INFO    Task definition already exists    task_name=credit_check
```

### 4. **Network Connectivity Issues**

**Issue**: `timeout` or `connection refused`
**Solution**: Check network and firewall settings
```bash
# Test connectivity
telnet localhost 8082
curl -v http://localhost:8082/health
```

## Configuration Requirements

### 1. **Conductor Configuration**

Ensure proper Conductor configuration:

```yaml
# config/config.yaml
conductor:
  base_url: "http://localhost:8082"
  timeout: 30
  retry_attempts: 3
  retry_delay: 1000
  worker_pool_size: 10
  polling_interval_ms: 1000
  update_retry_time_ms: 3000
```

### 2. **Network Configuration**

Ensure network connectivity:
- Conductor server accessible on configured port
- No firewall blocking connections
- Proper DNS resolution

### 3. **Logging Configuration**

Enable debug logging for troubleshooting:

```yaml
logging:
  level: "debug"
  format: "console"
  output: "stdout"
```

## Testing the Fix

### 1. **Unit Tests**

Run existing tests:
```bash
go test ./...
```

### 2. **Integration Tests**

Test with real Conductor:
```bash
# Start Conductor
docker run -p 8082:8082 conductor:latest

# Start underwriting worker
./underwriting-worker-service

# Submit a test workflow
curl -X POST http://localhost:8082/api/workflow/underwriting_workflow \
  -H "Content-Type: application/json" \
  -d '{
    "applicationId": "test-app-123",
    "userId": "test-user-123"
  }'
```

### 3. **Mock Testing**

Test with mock Conductor:
```bash
# Set environment to use mock
export APP_ENV=development

# Start worker (will use mock Conductor)
./underwriting-worker-service
```

## Monitoring and Alerts

### 1. **Key Metrics to Monitor**
- Task registration success rate
- Conductor connection status
- Task polling success rate
- Workflow execution success rate

### 2. **Alert Conditions**
- Task registration failures
- Conductor connection issues
- High task polling failure rates
- Workflow execution failures

## Troubleshooting Steps

### 1. **Check Conductor Status**
```bash
curl http://localhost:8082/health
curl http://localhost:8082/api/metadata/taskdefs
```

### 2. **Check Worker Logs**
```bash
grep -i "task registration" logs/underwriting-worker.log
grep -i "conductor" logs/underwriting-worker.log
grep -i "credit_check" logs/underwriting-worker.log
```

### 3. **Run Diagnostic Tool**
```bash
go run debug_task_registration.go
```

### 4. **Check Configuration**
```bash
# Verify configuration
cat config/config.yaml | grep -A 10 conductor
```

### 5. **Test Network Connectivity**
```bash
telnet localhost 8082
curl -v http://localhost:8082/health
```

## Next Steps

1. **Deploy the fixes** to your environment
2. **Run the diagnostic tool** to verify task registration
3. **Monitor the logs** for registration success
4. **Test workflow execution** with real tasks
5. **Set up monitoring** for ongoing health checks

## Contact

If the issue persists after implementing these fixes, please provide:
- Complete error logs
- Conductor server status
- Network connectivity test results
- Configuration details
- Steps to reproduce the issue
