# Pre-qualification Workflow Implementation

This document describes the implementation of the pre-qualification workflow for the loan service using Netflix Conductor.

## Overview

The pre-qualification workflow is a comprehensive system that evaluates loan applicants based on various criteria including income, debt-to-income ratio, employment status, and other risk factors. The workflow is implemented using Netflix Conductor as the workflow engine and follows a task-based architecture.

## Architecture

### Components

1. **PreQualificationWorkflowService** - Main service that orchestrates the workflow
2. **PreQualificationTaskHandler** - Handles individual workflow tasks
3. **TaskWorker** - Polls Conductor for tasks and executes them
4. **ConductorClientImpl** - HTTP client for Netflix Conductor
5. **Workflow Orchestrator** - Coordinates workflow execution

### Workflow Flow

```
User Request → Start Workflow → Execute Tasks → Return Result
                ↓
        [Validate Input] → [Calculate DTI] → [Assess Risk] → [Generate Terms] → [Finalize]
```

## Workflow Tasks

### 1. Validate Pre-qualify Input (`validate_prequalify_input`)

- Validates required fields (userId, loanAmount, annualIncome, employmentStatus)
- Checks field limits and constraints
- Returns validation status and any errors

**Input:**
```json
{
  "userId": "string",
  "loanAmount": "number",
  "annualIncome": "number",
  "monthlyDebt": "number",
  "employmentStatus": "string"
}
```

**Output:**
```json
{
  "valid": "boolean",
  "validationErrors": "object"
}
```

### 2. Calculate DTI Ratio (`calculate_dti_ratio`)

- Calculates debt-to-income ratio
- Computes monthly income from annual income
- Returns DTI ratio and monthly income

**Input:**
```json
{
  "annualIncome": "number",
  "monthlyDebt": "number"
}
```

**Output:**
```json
{
  "dtiRatio": "number",
  "monthlyIncome": "number"
}
```

### 3. Assess Pre-qualify Risk (`assess_prequalify_risk`)

- Evaluates risk based on DTI ratio, income, and employment
- Determines risk level (LOW, MEDIUM, HIGH)
- Calculates base interest rate
- Identifies risk factors

**Input:**
```json
{
  "loanAmount": "number",
  "annualIncome": "number",
  "employmentStatus": "string",
  "dtiRatio": "number"
}
```

**Output:**
```json
{
  "riskLevel": "string",
  "riskFactors": "array",
  "baseInterestRate": "number"
}
```

### 4. Generate Pre-qualify Terms (`generate_prequalify_terms`)

- Determines qualification status
- Calculates maximum loan amount
- Generates interest rate range
- Recommends loan terms
- Creates user message

**Input:**
```json
{
  "loanAmount": "number",
  "annualIncome": "number",
  "employmentStatus": "string",
  "dtiRatio": "number",
  "riskAssessment": "object"
}
```

**Output:**
```json
{
  "qualified": "boolean",
  "maxLoanAmount": "number",
  "interestRateRange": "object",
  "recommendedTerms": "array",
  "message": "string"
}
```

### 5. Finalize Pre-qualification (`finalize_prequalification`)

- Consolidates all results
- Generates pre-qualification ID
- Returns final pre-qualification result

**Input:**
```json
{
  "userId": "string",
  "qualified": "boolean",
  "maxLoanAmount": "number",
  "interestRateRange": "object",
  "recommendedTerms": "array",
  "dtiRatio": "number",
  "message": "string"
}
```

**Output:**
```json
{
  "qualified": "boolean",
  "maxLoanAmount": "number",
  "minInterestRate": "number",
  "maxInterestRate": "number",
  "recommendedTerms": "array",
  "dtiRatio": "number",
  "message": "string",
  "prequalificationId": "string"
}
```

## Configuration

The workflow is configured through `config/prequalification_workflow.yaml`:

```yaml
# Key configuration sections
workflow:
  engine: "netflix_conductor"
  conductor:
    base_url: "http://localhost:8080"

prequalification:
  dti:
    max_ratio: 0.43
  income:
    minimum_annual: 25000
  loan:
    maximum_amount: 50000
```

## Usage

### Starting the Workflow

```go
// Create the workflow service
workflowService := NewPreQualificationWorkflowService(
    workflowOrchestrator,
    logger,
    localizer,
)

// Process a pre-qualification request
result, err := workflowService.ProcessPreQualificationWorkflow(
    ctx,
    userID,
    prequalifyRequest,
)
```

### Starting the Task Worker

```go
// Create and start the task worker
worker := NewTaskWorker(conductorClient, logger, localizer)
go worker.Start(ctx)
```

## Business Rules

### Qualification Criteria

1. **DTI Ratio**: Must be ≤ 43%
2. **Income**: Minimum annual income of $25,000
3. **Employment**: Cannot be unemployed
4. **Risk Level**: High risk + high DTI = automatic rejection

### Interest Rate Calculation

- Base rate: 8.0%
- DTI adjustments: -0.5% to +3.0%
- Income adjustments: -1.0% to +1.0%
- Employment adjustments: 0% to +4.0%

### Loan Amount Limits

- Minimum: $5,000
- Maximum: $50,000
- Based on income and DTI ratio
- Cannot exceed 80% of annual income

## Error Handling

The workflow includes comprehensive error handling:

- **Validation Errors**: Field-level validation with specific error messages
- **Workflow Failures**: Automatic retry with exponential backoff
- **Task Failures**: Individual task failure handling
- **Timeout Handling**: Configurable timeouts for workflow and tasks

## Monitoring and Logging

- Structured logging with correlation IDs
- Workflow execution metrics
- Task performance monitoring
- Error tracking and alerting

## Testing

### Unit Tests

```bash
go test ./infrastructure/workflow -v
go test ./application -v
```

### Integration Tests

```bash
# Start Conductor locally
docker-compose up -d

# Run integration tests
go test ./... -tags=integration
```

## Deployment

### Prerequisites

1. Netflix Conductor server running
2. Go 1.19+ installed
3. Configuration files in place

### Steps

1. **Build the service:**
   ```bash
   go build -o loan-service ./cmd/main.go
   ```

2. **Deploy workflows to Conductor:**
   ```bash
   ./workflows/deploy.sh
   ```

3. **Start the task worker:**
   ```bash
   ./workflows/start-worker.sh
   ```

4. **Start the loan service:**
   ```bash
   ./loan-service
   ```

## Troubleshooting

### Common Issues

1. **Workflow not starting**: Check Conductor connectivity and workflow definition
2. **Tasks not executing**: Verify task worker is running and can reach Conductor
3. **Validation failures**: Check input data format and business rules
4. **Timeout errors**: Adjust timeout configuration values

### Debug Mode

Enable debug logging by setting the log level to DEBUG in the configuration:

```yaml
logging:
  level: "DEBUG"
```

### Health Checks

The service provides health check endpoints:

- `GET /health` - Service health
- `GET /workflows/{id}/status` - Workflow status

## Performance Considerations

- **Task Polling**: Configurable polling interval (default: 5 seconds)
- **Concurrent Execution**: Support for multiple concurrent task executions
- **Caching**: Consider implementing result caching for repeated requests
- **Database Optimization**: Ensure proper indexing for workflow state queries

## Security

- Input validation and sanitization
- User authentication and authorization
- Audit logging for all workflow operations
- Secure communication with Conductor (HTTPS recommended)

## Future Enhancements

1. **Machine Learning Integration**: Risk scoring using ML models
2. **Real-time Updates**: WebSocket-based status updates
3. **Advanced Analytics**: Workflow performance metrics and insights
4. **Multi-tenant Support**: Isolated workflows per organization
5. **API Versioning**: Support for multiple API versions
