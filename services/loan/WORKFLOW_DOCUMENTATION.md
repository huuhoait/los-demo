# Loan Service Workflow Documentation

## Overview
The Loan Service implements a comprehensive workflow system using Netflix Conductor for orchestrating loan processing, from initial application through funding and activation.

## Workflow Architecture

### 🔄 Three Main Workflows

#### 1. **Pre-qualification Workflow** (`prequalification_workflow`)
- **Purpose**: Quick assessment of loan eligibility
- **Trigger**: User submits pre-qualification request
- **Input**: Basic financial information
- **Output**: Qualification status and terms

#### 2. **Loan Processing Workflow** (`loan_processing_workflow`)
- **Purpose**: Complete loan application processing
- **Trigger**: User creates loan application
- **Input**: Full application details
- **Output**: Processed application with decision

#### 3. **Underwriting Workflow** (`underwriting_workflow`)
- **Purpose**: Detailed risk assessment and decision making
- **Trigger**: Application reaches underwriting state
- **Input**: Complete application + risk data
- **Output**: Approval/denial decision

## 📊 State Machine Flow

```
┌─────────────┐    ┌─────────────────┐    ┌──────────────────────┐
│  initiated  │───▶│ pre_qualified   │───▶│ documents_submitted  │
└─────────────┘    └─────────────────┘    └──────────────────────┘
                                                      │
┌─────────────┐    ┌─────────────────┐    ┌──────────────────────┐
│   closed    │◀───│     active      │◀───│ identity_verified    │
└─────────────┘    └─────────────────┘    └──────────────────────┘
                           ▲                          │
┌─────────────┐    ┌─────────────────┐    ┌──────────────────────┐
│   funded    │───▶│documents_signed │    │   underwriting       │◀─┐
└─────────────┘    └─────────────────┘    └──────────────────────┘  │
       ▲                   ▲                          │              │
       │                   │               ┌──────────▼──────────┐   │
       │                   │               │                     │   │
       │                   │               ▼                     ▼   │
       │                   │    ┌─────────────────┐    ┌─────────────┴─┐
       │                   └────│    approved     │    │ manual_review  │
       └────────────────────────└─────────────────┘    └───────────────┘
                                          │                     │
                                          ▼                     ▼
                                   ┌─────────────────┐   ┌─────────────────┐
                                   │     denied      │   │     denied      │
                                   └─────────────────┘   └─────────────────┘
```

## 🎯 State Transitions & Workflow Tasks

### State Transition Mapping

| From State | To State | Workflow Task | Description |
|------------|----------|---------------|-------------|
| `initiated` | `pre_qualified` | `prequalification_task` | Initial qualification check |
| `pre_qualified` | `documents_submitted` | `document_collection_task` | Collect required documents |
| `documents_submitted` | `identity_verified` | `identity_verification_task` | Verify customer identity |
| `identity_verified` | `underwriting` | `automated_underwriting_task` | Run automated risk assessment |
| `underwriting` | `approved` | `approval_task` | Automatic approval |
| `underwriting` | `denied` | `denial_task` | Automatic denial |
| `underwriting` | `manual_review` | `manual_review_task` | Flag for manual review |
| `manual_review` | `approved` | `manual_approval_task` | Manual approval |
| `manual_review` | `denied` | `manual_denial_task` | Manual denial |
| `approved` | `documents_signed` | `document_signing_task` | Generate and sign loan docs |
| `documents_signed` | `funded` | `funding_task` | Disburse loan funds |
| `funded` | `active` | `loan_activation_task` | Activate loan account |

## 🔧 Workflow Implementation

### Netflix Conductor Integration

#### Interface Definition
```go
type ConductorClient interface {
    StartWorkflow(ctx context.Context, workflowName string, version int, input map[string]interface{}) (*WorkflowExecution, error)
    GetWorkflowStatus(ctx context.Context, workflowID string) (*WorkflowStatus, error)
    TerminateWorkflow(ctx context.Context, workflowID string, reason string) error
    PauseWorkflow(ctx context.Context, workflowID string) error
    ResumeWorkflow(ctx context.Context, workflowID string) error
    UpdateTask(ctx context.Context, taskID string, status string, output map[string]interface{}) error
}
```

#### Workflow Orchestrator
The `LoanWorkflowOrchestrator` manages all workflow interactions:
- **StartLoanProcessingWorkflow**: Initiates main loan processing
- **StartPreQualificationWorkflow**: Runs pre-qualification assessment
- **StartUnderwritingWorkflow**: Triggers underwriting process
- **HandleStateTransition**: Manages state changes and task triggers

## 📝 Workflow Data Flow

### Pre-qualification Workflow Input
```json
{
  "userId": "user123",
  "loanAmount": 30000,
  "annualIncome": 60000,
  "monthlyDebt": 500,
  "employmentStatus": "full_time",
  "startTime": "2025-08-14T07:00:00Z"
}
```

### Loan Processing Workflow Input
```json
{
  "applicationId": "app123",
  "userId": "user123",
  "loanAmount": 30000,
  "loanPurpose": "personal",
  "annualIncome": 60000,
  "monthlyIncome": 5000,
  "monthlyDebt": 500,
  "requestedTerm": 48,
  "currentState": "initiated",
  "startTime": "2025-08-14T07:00:00Z"
}
```

### Underwriting Workflow Input
```json
{
  "applicationId": "app123",
  "userId": "user123",
  "loanAmount": 30000,
  "annualIncome": 60000,
  "monthlyIncome": 5000,
  "monthlyDebt": 500,
  "dtiRatio": 0.1,
  "riskScore": 750,
  "startTime": "2025-08-14T07:00:00Z"
}
```

## 🚀 Workflow Execution Examples

### 1. Pre-qualification Flow
```
User Request → Pre-qualification Workflow → Risk Assessment → Result
```

### 2. Full Application Flow
```
Application Creation → Loan Processing Workflow → State Transitions → Final Decision
```

### 3. Underwriting Flow
```
Identity Verified → Underwriting Workflow → Automated/Manual Review → Approval/Denial
```

## 🔍 Monitoring & Status

### Workflow Status Tracking
- **WorkflowID**: Unique identifier for each workflow instance
- **Status**: RUNNING, COMPLETED, FAILED, TERMINATED
- **Tasks**: Individual task status within workflow
- **Correlation**: Link workflows to loan applications

### Task Status Types
- **COMPLETED**: Task finished successfully
- **IN_PROGRESS**: Task currently executing
- **FAILED**: Task encountered error
- **SCHEDULED**: Task waiting to execute

## 🛠️ Development & Testing

### Mock Implementation
For development and testing, the service uses the real Conductor client that:
- Simulates workflow execution
- Provides realistic responses
- Enables offline development
- Supports integration testing

### Production Deployment
To use with real Netflix Conductor:
1. Ensure Conductor service is running and accessible
2. Configure Conductor server endpoints
3. Deploy workflow definitions to Conductor
4. Update authentication and networking

## 📊 Business Rules

### Pre-qualification Rules
- Minimum loan amount: $5,000
- Maximum loan amount: $50,000
- DTI ratio calculation for risk assessment
- Employment status affects interest rates

### Underwriting Rules
- Automated approval for low-risk applications
- Manual review for edge cases
- Risk score thresholds for decision making
- Income verification requirements

### State Transition Rules
- Linear progression through most states
- Branching at underwriting (approve/deny/review)
- No backward transitions (except manual review)
- Terminal states: denied, closed

## 🌐 Internationalization

The workflow system supports bilingual operation:
- **Vietnamese (vi)**: Complete workflow messages in Vietnamese
- **English (en)**: Complete workflow messages in English
- **Dynamic Language**: Language detection via HTTP headers
- **Localized Errors**: All error messages in user's language

## 🔐 Security & Compliance

### Data Protection
- All workflow data encrypted in transit
- Sensitive data masked in logs
- Audit trail for all state changes
- Compliance with financial regulations

### Access Control
- User authentication required
- Role-based access to workflow operations
- Secure API endpoints
- Request ID tracking for audit

---

*This workflow system provides a robust, scalable foundation for loan processing with comprehensive state management, Netflix Conductor integration, and full internationalization support.*
