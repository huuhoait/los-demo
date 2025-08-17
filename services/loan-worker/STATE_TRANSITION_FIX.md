# State Transition Fix for Finalize Loan Decision Task

## Issue Description

The error `"task execution failed: invalid state transition from pre_qualified to identity_verified"` indicates that the workflow is trying to execute the `finalize_loan_decision` task when the application is in an invalid state.

## Root Cause

The `finalize_loan_decision` task was being called when the application was in `pre_qualified` state, but according to the workflow state machine, this task should only be executed after the application has gone through the proper sequence:

1. `initiated` → `pre_qualified` (after validation)
2. `pre_qualified` → `documents_submitted` (after document collection)
3. `documents_submitted` → `identity_verified` (after identity verification)
4. `identity_verified` → underwriting workflow
5. `finalize_loan_decision` (should only be called after proper workflow steps)

## State Transition Rules

The valid state transitions are defined in `domain/models.go`:

```go
func (app *LoanApplication) CanTransitionTo(newState ApplicationState) bool {
    validTransitions := map[ApplicationState][]ApplicationState{
        StateInitiated:          {StatePreQualified},
        StatePreQualified:       {StateDocumentsSubmitted},
        StateDocumentsSubmitted: {StateIdentityVerified},
        StateIdentityVerified:   {StateApproved, StateDenied, StateManualReview},
        StateManualReview:       {StateApproved, StateDenied},
        StateApproved:           {StateDocumentsSigned},
        StateDocumentsSigned:    {StateFunded},
        StateFunded:             {StateActive},
        StateActive:             {StateClosed},
    }
    // ... validation logic
}
```

## Fixes Implemented

### 1. Enhanced Task Validation

Updated `finalize_loan_decision_task.go` to include state validation:

```go
// Validate that the task is being called in the correct workflow sequence
// This task should only be called after identity verification and underwriting
if currentState != "" {
    validStatesForFinalization := []string{
        "identity_verified",
        "approved", 
        "denied",
        "manual_review",
    }
    
    // ... validation logic
}
```

### 2. Workflow Configuration Update

Updated `loan-api/workflows/loan_processing_workflow.json` to pass the `currentState` parameter to the `finalize_loan_decision` task:

```json
{
  "name": "finalize_loan_decision",
  "taskReferenceName": "finalize_loan_decision_ref",
  "inputParameters": {
    "applicationId": "${workflow.input.applicationId}",
    "currentState": "${workflow.input.currentState}",
    "underwritingResult": "${trigger_underwriting_ref.output}",
    "finalState": "${trigger_underwriting_ref.output.finalState}",
    "decision": "${trigger_underwriting_ref.output.decision}",
    "interestRate": "${trigger_underwriting_ref.output.interestRate}"
  }
}
```

### 3. Workflow Execution Order Validation

Added `ValidateWorkflowExecutionOrder` method in `conductor.go` to ensure tasks are executed in the correct sequence:

```go
func (o *LoanWorkflowOrchestrator) ValidateWorkflowExecutionOrder(ctx context.Context, applicationID string, currentState, targetState domain.ApplicationState, taskName string) error {
    // Define valid task execution states
    validTaskStates := map[string][]domain.ApplicationState{
        "finalize_loan_decision": {
            domain.StateIdentityVerified,
            domain.StateApproved,
            domain.StateDenied,
            domain.StateManualReview,
            domain.StateDocumentsSigned,
            domain.StateFunded,
            domain.StateActive,
        },
        // ... other task validations
    }
    // ... validation logic
}
```

### 4. Enhanced Logging

Added detailed logging in `task_worker.go` to help debug workflow execution issues:

```go
// Special logging for finalize_loan_decision task to help debug state transition issues
if task.ReferenceTaskName == "finalize_loan_decision_ref" {
    if currentState, exists := task.Input["currentState"]; exists {
        logger.Info("Finalize loan decision task input validation",
            zap.String("current_state", fmt.Sprintf("%v", currentState)),
            zap.String("final_state", fmt.Sprintf("%v", task.Input["finalState"])),
            zap.String("decision", fmt.Sprintf("%v", task.Input["decision"])))
    }
}
```

## How to Use

### 1. Ensure Proper Workflow Sequence

When calling the `finalize_loan_decision` task, make sure the application has gone through the required states:

```bash
# Correct sequence:
initiated → pre_qualified → documents_submitted → identity_verified → finalize_loan_decision

# Incorrect sequence (will fail):
initiated → pre_qualified → finalize_loan_decision  # Missing intermediate steps
```

### 2. Pass Required Parameters

Always pass the `currentState` parameter when calling the `finalize_loan_decision` task:

```json
{
  "currentState": "identity_verified",
  "finalState": "approved",
  "decision": "approved",
  "interestRate": 5.5
}
```

### 3. Monitor Logs

Check the logs for detailed information about task execution and state validation:

```bash
# Look for these log messages:
"Finalize loan decision task input validation"
"Invalid application state for loan decision finalization"
"Workflow execution order validation passed"
```

## Testing

To test the fix:

1. **Valid Workflow**: Start with an application in `initiated` state and follow the complete workflow sequence
2. **Invalid Workflow**: Try to call `finalize_loan_decision` from `pre_qualified` state - it should now fail with a clear error message
3. **State Validation**: Verify that the task only executes when the application is in an appropriate state

## Prevention

To prevent similar issues in the future:

1. Always validate state transitions before executing tasks
2. Use the `ValidateWorkflowExecutionOrder` method for critical tasks
3. Ensure workflow configurations pass all required parameters
4. Monitor logs for state transition validation messages
5. Follow the defined state machine rules strictly

## Related Files

- `loan-worker/infrastructure/workflow/tasks/finalize_loan_decision_task.go` - Enhanced task validation
- `loan-api/workflows/loan_processing_workflow.json` - Updated workflow configuration
- `loan-worker/infrastructure/workflow/conductor.go` - Added workflow validation
- `loan-worker/infrastructure/workflow/task_worker.go` - Enhanced logging
- `loan-worker/domain/models.go` - State transition rules
