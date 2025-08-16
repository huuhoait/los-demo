# Update Application State Task Implementation

## Overview

The `update_application_state` task has been fully implemented to handle loan application state transitions with proper database integration, state validation, and audit trail management.

## Features

### 1. **Database Integration**
- Real database operations for updating application state
- State transition recording with full audit trail
- Repository pattern with interface abstraction to avoid import cycles

### 2. **State Validation**
- Enforces valid state transition rules defined in the domain model
- Prevents invalid state jumps (e.g., from `pre_qualified` directly to `funded`)
- Validates required input parameters

### 3. **Comprehensive State Management**
- Updates both `ApplicationState` and `ApplicationStatus` appropriately
- Handles automated and manual state transitions
- Tracks user who initiated the transition
- Records transition timestamps and reasons

### 4. **Fallback Mode**
- Simulation mode when no database repository is available
- Maintains basic validation even in simulation mode
- Useful for testing and development

## Implementation Details

### Task Handler Structure

```go
type UpdateApplicationStateTaskHandler struct {
    logger         *zap.Logger
    loanRepository LoanRepository
}
```

### Repository Interface

```go
type LoanRepository interface {
    GetApplicationByID(ctx context.Context, id string) (*domain.LoanApplication, error)
    UpdateApplication(ctx context.Context, app *domain.LoanApplication) error
    CreateStateTransition(ctx context.Context, transition *domain.StateTransition) error
}
```

### Input Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `applicationId` | string | Yes | ID of the loan application |
| `toState` | string | Yes | Target application state |
| `fromState` | string | No | Current state (for validation) |
| `reason` | string | No | Reason for state transition |
| `userId` | string | No | ID of user who initiated transition |
| `automated` | boolean | No | Whether transition is automated |

### Output Format

```json
{
  "success": true,
  "updatedAt": "2025-08-16T00:19:22.375Z",
  "previousState": "initiated",
  "newState": "pre_qualified",
  "newStatus": "submitted",
  "transition": {
    "id": "uuid",
    "fromState": "initiated",
    "toState": "pre_qualified",
    "reason": "Pre-qualification completed",
    "automated": true,
    "timestamp": "2025-08-16T00:19:22.375Z",
    "userId": "user-123"
  }
}
```

## Valid State Transitions

Based on the domain model, the following state transitions are allowed:

```
initiated → pre_qualified
pre_qualified → documents_submitted
documents_submitted → identity_verified
identity_verified → underwriting
underwriting → {approved, denied, manual_review}
manual_review → {approved, denied}
approved → documents_signed
documents_signed → funded
funded → active
active → closed
```

## Integration with Task System

### Factory Pattern Integration

The task is registered in the `TaskFactory` with dependency injection support:

```go
// With repository
factory := tasks.NewTaskFactoryWithRepository(logger, loanRepository)

// Without repository (simulation mode)
factory := tasks.NewTaskFactory(logger)
```

### Task Worker Integration

The task worker supports both modes:

```go
// With repository
worker := workflow.NewTaskWorkerWithRepository(conductorClient, logger, localizer, loanRepository)

// Without repository
worker := workflow.NewTaskWorker(conductorClient, logger, localizer)
```

## Error Handling

### Validation Errors
- Missing `applicationId`: "application ID is required"
- Missing `toState`: "target state is required"
- Invalid transition: "invalid state transition from {current} to {target}"

### Database Errors
- Application not found: "application not found: {id}"
- Update failures: "failed to update application: {error}"
- State transition record failures: "failed to create state transition record: {error}"

## Status Mapping

The task automatically updates the application status based on the new state:

| Application State | Application Status |
|-------------------|-------------------|
| `approved` | `approved` |
| `denied` | `denied` |
| `funded` | `funded` |
| `active` | `active` |
| `closed` | `closed` |
| Others | `submitted` or `under_review` |

## Testing

The implementation has been tested with:

1. ✅ Valid state transitions
2. ✅ Invalid state transitions (properly rejected)
3. ✅ Sequential state progression
4. ✅ Missing required fields validation
5. ✅ Simulation mode functionality
6. ✅ State transition audit trail creation
7. ✅ User tracking and automated flag handling

## Usage in Workflows

This task can be used in Netflix Conductor workflows for:

- Prequalification workflows (moving to `pre_qualified`)
- Document collection workflows (moving to `documents_submitted`)
- Identity verification workflows (moving to `identity_verified`)
- Underwriting workflows (moving to `underwriting`, `approved`, `denied`)
- Loan funding workflows (moving to `funded`, `active`)
- Loan closure workflows (moving to `closed`)

## Configuration

The task automatically adapts based on available dependencies:

- **With Database**: Full functionality with persistent state changes
- **Without Database**: Simulation mode with validation only
- **Logging**: Comprehensive logging at all stages
- **Audit Trail**: Complete state transition history when database is available

## Performance Considerations

- Single database transaction per state update
- Efficient state validation using domain model rules
- Minimal overhead in simulation mode
- Proper error handling prevents partial state updates
