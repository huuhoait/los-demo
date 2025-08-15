# Netflix Conductor Workflow Definitions for Loan Service

This directory contains complete Netflix Conductor workflow and task definitions for the loan processing system.

## 📁 Directory Structure

```
workflows/
├── deploy.sh                           # Deployment script
├── README.md                           # This file
├── prequalification_workflow.json     # Pre-qualification workflow
├── loan_processing_workflow.json      # Main loan processing workflow  
├── underwriting_workflow.json         # Underwriting workflow with decision engine
└── tasks/
    ├── prequalification_tasks.json    # Pre-qualification task definitions
    ├── loan_processing_tasks.json     # Loan processing task definitions
    └── underwriting_tasks.json        # Underwriting task definitions
```

## 🔄 Workflow Overview

### 1. Pre-qualification Workflow (`prequalification_workflow`)
- **Purpose**: Quick loan eligibility assessment
- **Duration**: ~30 seconds
- **Tasks**: 5 sequential tasks
- **Input**: Basic financial information
- **Output**: Qualification result with terms

**Task Flow:**
```
validate_prequalify_input → calculate_dti_ratio → assess_prequalify_risk → 
generate_prequalify_terms → finalize_prequalification
```

### 2. Loan Processing Workflow (`loan_processing_workflow`)
- **Purpose**: Complete loan application processing
- **Duration**: 1-3 days (includes human tasks)
- **Tasks**: 8 tasks with sub-workflow
- **Input**: Full loan application
- **Output**: Final loan decision

**Task Flow:**
```
validate_application → update_state_to_prequalified → document_collection → 
update_state_to_documents_submitted → identity_verification → 
update_state_to_identity_verified → trigger_underwriting → finalize_loan_decision
```

### 3. Underwriting Workflow (`underwriting_workflow`)
- **Purpose**: Risk assessment and decision making
- **Duration**: 15 minutes - 2 days (depending on risk level)
- **Tasks**: Complex decision tree with manual review paths
- **Input**: Verified application data
- **Output**: Approval/denial with terms

**Task Flow:**
```
update_state_to_underwriting → credit_check → income_verification → 
calculate_risk_score → decision_engine → [auto_approve|auto_deny|manual_review]
```

## 🚀 Deployment Instructions

### Prerequisites
- Netflix Conductor server running
- curl command available
- Proper network access to Conductor server

### Quick Deployment
```bash
# Set Conductor server URL (default: http://localhost:8080)
export CONDUCTOR_SERVER=http://your-conductor-server:8080

# Deploy all workflows and tasks
./deploy.sh
```

### Manual Deployment

#### 1. Deploy Task Definitions
```bash
# Deploy pre-qualification tasks
curl -X POST $CONDUCTOR_SERVER/api/metadata/taskdefs \
  -H "Content-Type: application/json" \
  -d @tasks/prequalification_tasks.json

# Deploy loan processing tasks  
curl -X POST $CONDUCTOR_SERVER/api/metadata/taskdefs \
  -H "Content-Type: application/json" \
  -d @tasks/loan_processing_tasks.json

# Deploy underwriting tasks
curl -X POST $CONDUCTOR_SERVER/api/metadata/taskdefs \
  -H "Content-Type: application/json" \
  -d @tasks/underwriting_tasks.json
```

#### 2. Deploy Workflow Definitions
```bash
# Deploy pre-qualification workflow
curl -X PUT $CONDUCTOR_SERVER/api/metadata/workflow \
  -H "Content-Type: application/json" \
  -d @prequalification_workflow.json

# Deploy loan processing workflow
curl -X PUT $CONDUCTOR_SERVER/api/metadata/workflow \
  -H "Content-Type: application/json" \
  -d @loan_processing_workflow.json

# Deploy underwriting workflow  
curl -X PUT $CONDUCTOR_SERVER/api/metadata/workflow \
  -H "Content-Type: application/json" \
  -d @underwriting_workflow.json
```

## 🎯 Task Definitions Summary

### Pre-qualification Tasks (5 tasks)
- `validate_prequalify_input`: Input validation (30s timeout)
- `calculate_dti_ratio`: DTI calculation (15s timeout)  
- `assess_prequalify_risk`: Risk assessment (60s timeout)
- `generate_prequalify_terms`: Terms generation (45s timeout)
- `finalize_prequalification`: Result finalization (30s timeout)

### Loan Processing Tasks (5 tasks)
- `validate_application`: Application validation (60s timeout)
- `update_application_state`: State updates (30s timeout)
- `document_collection`: Document collection - HUMAN task (24h timeout)
- `identity_verification`: Identity verification (120s timeout)
- `finalize_loan_decision`: Decision finalization (45s timeout)

### Underwriting Tasks (10 tasks)
- `credit_check`: Credit bureau check (120s timeout)
- `income_verification`: Income verification (180s timeout)
- `calculate_risk_score`: Risk scoring (90s timeout)
- `auto_approve`: Automatic approval (45s timeout)
- `auto_deny`: Automatic denial (30s timeout)
- `flag_for_manual_review`: Manual review flagging (30s timeout)
- `manual_underwriting_review`: Manual review - HUMAN task (48h timeout)
- `manual_approve`: Manual approval (45s timeout)
- `manual_deny`: Manual denial (30s timeout)
- `process_manual_decision`: Manual decision processing (30s timeout)

## 🔧 Task Types

### SIMPLE Tasks
- Automated processing tasks
- Synchronous execution
- Defined timeout and retry policies
- Used for business logic, calculations, database operations

### HUMAN Tasks  
- Manual intervention required
- Asynchronous execution
- Long timeouts (24-48 hours)
- Used for document collection, manual reviews

### DECISION Tasks
- Conditional branching
- Route workflow based on data
- Multiple execution paths
- Used for risk-based decisions

### SUB_WORKFLOW Tasks
- Nested workflow execution
- Modular workflow design
- Independent execution context
- Used for complex multi-step processes

## 📊 Workflow Execution Examples

### Start Pre-qualification Workflow
```bash
curl -X POST $CONDUCTOR_SERVER/api/workflow \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prequalification_workflow",
    "version": 1,
    "input": {
      "userId": "user123",
      "loanAmount": 30000,
      "annualIncome": 60000,
      "monthlyDebt": 500,
      "employmentStatus": "full_time"
    }
  }'
```

### Start Loan Processing Workflow
```bash
curl -X POST $CONDUCTOR_SERVER/api/workflow \
  -H "Content-Type: application/json" \
  -d '{
    "name": "loan_processing_workflow", 
    "version": 1,
    "input": {
      "applicationId": "app123",
      "userId": "user123",
      "loanAmount": 30000,
      "loanPurpose": "personal",
      "annualIncome": 60000,
      "monthlyIncome": 5000,
      "monthlyDebt": 500,
      "requestedTerm": 48,
      "currentState": "initiated"
    }
  }'
```

## 🔍 Monitoring & Management

### Check Workflow Status
```bash
curl $CONDUCTOR_SERVER/api/workflow/{workflowId}
```

### View Running Workflows
```bash
curl $CONDUCTOR_SERVER/api/workflow/running/{workflowName}
```

### Search Workflows
```bash
curl "$CONDUCTOR_SERVER/api/workflow/search?query=workflowType:loan_processing_workflow"
```

### Pause/Resume Workflow
```bash
# Pause
curl -X PUT $CONDUCTOR_SERVER/api/workflow/{workflowId}/pause

# Resume  
curl -X PUT $CONDUCTOR_SERVER/api/workflow/{workflowId}/resume
```

## 🚨 Error Handling

### Retry Policies
- **FIXED**: Fixed retry intervals
- **EXPONENTIAL_BACKOFF**: Exponential backoff strategy
- **LINEAR_BACKOFF**: Linear backoff strategy

### Timeout Policies
- **TIME_OUT_WF**: Timeout the workflow
- **ALERT_ONLY**: Send alert but continue
- **RETRY**: Retry the task

### Failure Handling
- Task failures trigger retries based on retry policy
- Workflow failures can trigger failure workflows
- Dead letter queues for manual intervention
- Comprehensive logging and alerting

## 🔐 Security Considerations

### Access Control
- Configure Conductor authentication
- Implement task worker authentication
- Secure API endpoints
- Monitor workflow access

### Data Protection
- Encrypt sensitive data in workflow payloads
- Implement data masking in logs
- Secure inter-service communication
- Comply with financial regulations

## 📈 Performance Tuning

### Concurrency Limits
- Pre-qualification: 100-200 concurrent executions
- Loan processing: 50-100 concurrent executions  
- Underwriting: 30-50 concurrent executions
- Credit checks: Rate limited (10/minute)

### Optimization Tips
- Use appropriate timeout values
- Configure proper retry policies
- Monitor task execution times
- Scale workers based on load

---

*These workflow definitions provide a production-ready foundation for loan processing with comprehensive error handling, monitoring, and scalability features.*
