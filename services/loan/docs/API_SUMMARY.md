# Loan Service API Documentation Summary

## Overview

The Loan Service API provides comprehensive loan origination capabilities with Netflix Conductor workflow integration, internationalization support (Vietnamese and English), and robust error handling.

**Base URL**: `http://localhost:8080/v1`
**API Version**: 1.0
**Authentication**: Bearer Token (JWT)

## API Endpoints

### 🔍 Health & Status

#### Health Check
- **GET** `/health`
- **Description**: Check the health status of the loan service
- **Response**: Service health information
- **Authentication**: None required

### 📋 Loan Applications

#### Create Application
- **POST** `/loans/applications`
- **Description**: Create a new loan application with automatic user creation
- **Request Body**: `CreateApplicationRequest` (includes complete user information)
- **Response**: Created `LoanApplication`
- **Authentication**: Required
- **Features**: 
  - Automatically creates new users or finds existing ones by email
  - Comprehensive user data collection (personal, address, employment, banking)
  - Data validation and privacy protection

#### Get Applications by User
- **GET** `/loans/applications`
- **Description**: Retrieve all applications for the current user
- **Response**: Array of `LoanApplication`
- **Authentication**: Required

#### Get Application by ID
- **GET** `/loans/applications/{id}`
- **Description**: Retrieve a specific application by ID
- **Path Parameters**: `id` (string)
- **Response**: `LoanApplication`
- **Authentication**: Required

#### Update Application
- **PUT** `/loans/applications/{id}`
- **Description**: Update an existing application
- **Path Parameters**: `id` (string)
- **Request Body**: `UpdateApplicationRequest`
- **Response**: Updated `LoanApplication`
- **Authentication**: Required

#### Submit Application
- **POST** `/loans/applications/{id}/submit`
- **Description**: Submit a draft application for processing
- **Path Parameters**: `id` (string)
- **Response**: Updated `LoanApplication`
- **Authentication**: Required

### 🎯 Pre-qualification

#### Pre-qualify Loan
- **POST** `/loans/prequalify`
- **Description**: Check if a user qualifies for a loan
- **Request Body**: `PreQualifyRequest`
- **Response**: `PreQualifyResult`
- **Authentication**: Required
- **Workflow**: Triggers prequalification workflow

**Request Body Example**:
```json
{
  "loan_amount": 25000,
  "annual_income": 75000,
  "monthly_debt_payments": 1500,
  "employment_status": "full_time"
}
```

**Response Example**:
```json
{
  "success": true,
  "data": {
    "qualified": true,
    "max_loan_amount": 50000,
    "min_interest_rate": 8.5,
    "max_interest_rate": 12.0,
    "recommended_terms": [36, 48, 60],
    "dti_ratio": 0.24,
    "message": "You are pre-qualified for a loan"
  }
}
```

### 💰 Loan Offers

#### Generate Offer
- **POST** `/loans/applications/{id}/offer`
- **Description**: Generate a loan offer for an application
- **Path Parameters**: `id` (string)
- **Response**: `LoanOffer`
- **Authentication**: Required

#### Accept Offer
- **POST** `/loans/applications/{id}/accept-offer`
- **Description**: Accept a loan offer
- **Path Parameters**: `id` (string)
- **Request Body**: `AcceptOfferRequest`
- **Response**: Success confirmation
- **Authentication**: Required

### 🔄 Workflow Management

#### Get Workflow Status
- **GET** `/workflows/{id}/status`
- **Description**: Retrieve the current status of a workflow execution
- **Path Parameters**: `id` (string)
- **Response**: `WorkflowStatus`
- **Authentication**: Required

#### Pause Workflow
- **POST** `/workflows/{id}/pause`
- **Description**: Pause a running workflow execution
- **Path Parameters**: `id` (string)
- **Response**: Success confirmation
- **Authentication**: Required

#### Resume Workflow
- **POST** `/workflows/{id}/resume`
- **Description**: Resume a paused workflow execution
- **Path Parameters**: `id` (string)
- **Response**: Success confirmation
- **Authentication**: Required

#### Terminate Workflow
- **POST** `/workflows/{id}/terminate`
- **Description**: Terminate a running workflow execution
- **Path Parameters**: `id` (string)
- **Request Body**: Termination reason
- **Response**: Success confirmation
- **Authentication**: Required

### 📊 Administration

#### Get Application Statistics
- **GET** `/loans/stats`
- **Description**: Retrieve application statistics (admin only)
- **Query Parameters**: 
  - `status` (optional): Filter by status
  - `state` (optional): Filter by state
  - `days` (optional): Number of days to look back
- **Response**: Statistics object
- **Authentication**: Required

#### State Transition
- **POST** `/loans/applications/{id}/transition`
- **Description**: Transition application state (admin only)
- **Path Parameters**: `id` (string)
- **Request Body**: State transition details
- **Response**: Success confirmation
- **Authentication**: Required

## Data Models

### Core Entities

#### CreateApplicationRequest
```json
{
  "user": {
    "first_name": "string",
    "last_name": "string",
    "email": "string (email format)",
    "phone_number": "string",
    "date_of_birth": "string (date-time, must be 18+)",
    "ssn": "string (9 digits)",
    "address": {
      "street_address": "string",
      "city": "string",
      "state": "string",
      "zip_code": "string",
      "country": "string",
      "residence_type": "string (own|rent|family|other)",
      "time_at_address_months": "integer (>=0)"
    },
    "employment_info": {
      "employer_name": "string",
      "job_title": "string",
      "time_employed_months": "integer (>=0)",
      "work_phone": "string",
      "work_email": "string (email format, optional)"
    },
    "banking_info": {
      "bank_name": "string",
      "account_type": "string (checking|savings)",
      "account_number": "string",
      "routing_number": "string (9 digits)"
    }
  },
  "loan_amount": "number (5000-50000)",
  "loan_purpose": "string",
  "requested_term_months": "integer (12-84)",
  "annual_income": "number (>=0)",
  "monthly_income": "number (>=0)",
  "employment_status": "string",
  "monthly_debt_payments": "number (>=0)"
}
```

#### LoanApplication
```json
{
  "id": "string",
  "user_id": "string",
  "application_number": "string",
  "loan_amount": "number",
  "loan_purpose": "string",
  "requested_term_months": "integer",
  "annual_income": "number",
  "monthly_income": "number",
  "employment_status": "string",
  "monthly_debt_payments": "number",
  "current_state": "string",
  "status": "string",
  "risk_score": "integer",
  "workflow_id": "string",
  "created_at": "string",
  "updated_at": "string"
}
```

#### PreQualifyRequest
```json
{
  "loan_amount": "number (5000-50000)",
  "annual_income": "number (min: 25000)",
  "monthly_debt_payments": "number (min: 0)",
  "employment_status": "string"
}
```

#### PreQualifyResult
```json
{
  "qualified": "boolean",
  "max_loan_amount": "number",
  "min_interest_rate": "number",
  "max_interest_rate": "number",
  "recommended_terms": "array[integer]",
  "dti_ratio": "number",
  "message": "string"
}
```

### Enums

#### ApplicationState
- `initiated`
- `pre_qualified`
- `documents_submitted`
- `identity_verified`
- `underwriting`
- `manual_review`
- `approved`
- `denied`
- `documents_signed`
- `funded`
- `active`
- `closed`

#### ApplicationStatus
- `draft`
- `submitted`
- `under_review`
- `approved`
- `denied`
- `funded`
- `active`
- `closed`

#### EmploymentStatus
- `full_time`
- `part_time`
- `self_employed`
- `retired`
- `unemployed`
- `student`

#### LoanPurpose
- `debt_consolidation`
- `home_improvement`
- `medical`
- `vacation`
- `wedding`
- `major_purchase`
- `other`

#### ResidenceType
- `own`
- `rent`
- `family`
- `other`

#### AccountType
- `checking`
- `savings`

## User Management

### Features
- **Automatic User Creation**: New users are automatically created when submitting loan applications
- **Existing User Detection**: System checks for existing users by email address
- **Complete User Profile**: Collects comprehensive user information in a single request
- **Data Validation**: Validates all user fields including age verification (18+)
- **Privacy Protection**: Sensitive data like SSN and account numbers are masked in responses

### User Information Collected
- **Personal Details**: First name, last name, email, phone, date of birth, SSN
- **Address**: Complete address with residence type and time at address
- **Employment**: Employer details, job title, tenure, work contact information
- **Banking**: Bank details for loan disbursement

## Workflow Integration

### Pre-qualification Workflow

The pre-qualification endpoint triggers a comprehensive workflow that includes:

1. **Input Validation** - Validates request parameters
2. **DTI Calculation** - Calculates debt-to-income ratio
3. **Risk Assessment** - Evaluates risk factors
4. **Term Generation** - Generates loan terms and rates
5. **Finalization** - Consolidates results and generates ID

### Workflow States

- **RUNNING**: Workflow is actively executing
- **COMPLETED**: Workflow finished successfully
- **FAILED**: Workflow encountered an error
- **PAUSED**: Workflow is paused
- **TERMINATED**: Workflow was terminated

## Error Handling

### Error Response Format
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "LOAN_001",
    "message": "Invalid loan amount",
    "description": "The loan amount must be between 5000 and 50000",
    "field": "loan_amount"
  }
}
```

### Common Error Codes

- **LOAN_001**: Invalid loan amount
- **LOAN_002**: Invalid loan purpose
- **LOAN_003**: Invalid loan term
- **LOAN_004**: Invalid income information
- **LOAN_005**: Loan amount below minimum
- **LOAN_006**: Loan amount above maximum
- **LOAN_007**: Insufficient income
- **LOAN_011**: Workflow start failed
- **LOAN_012**: Workflow execution failed
- **LOAN_020**: Invalid request format
- **LOAN_022**: Unauthorized access
- **LOAN_023**: Database error

## Authentication

### Bearer Token
All protected endpoints require a valid JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Mock Authentication (Development)
For development purposes, the service includes mock authentication that automatically sets:
- `user_id`: "user123"
- `user_email`: "user@example.com"

## Internationalization

### Language Support
The API supports multiple languages through the `X-Language` header:
- `en`: English (default)
- `vi`: Vietnamese

### Example
```bash
curl -H "X-Language: vi" \
     -H "Authorization: Bearer <token>" \
     http://localhost:8080/v1/loans/applications
```

## Rate Limiting

Currently, no rate limiting is implemented. Consider implementing rate limiting for production use.

## Pagination

Currently, no pagination is implemented for list endpoints. Consider implementing pagination for large datasets.

## Testing

### Swagger UI
Access the interactive API documentation at:
- **Local**: Open `swagger-ui.html` in your browser
- **Service**: `http://localhost:8080/v1/swagger/index.html`

### Test Scripts
Use the provided test scripts to validate the API:
```bash
./scripts/test_prequalification_workflow.sh
```

## Deployment

### Prerequisites
1. Go 1.19+
2. PostgreSQL database
3. Redis cache
4. Netflix Conductor server

### Environment Variables
```bash
# Server
PORT=8080
HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=loan_service

# Conductor
CONDUCTOR_BASE_URL=http://localhost:8082

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Docker Compose
```bash
docker-compose up -d
```

## Monitoring & Health Checks

### Health Endpoint
- **URL**: `/v1/health`
- **Purpose**: Service health monitoring
- **Response**: Service status and version information

### Workflow Monitoring
- Monitor workflow execution through Conductor UI
- Track workflow performance and failures
- Set up alerts for workflow timeouts

## Security Considerations

### Production Security
1. **HTTPS**: Always use HTTPS in production
2. **JWT Validation**: Implement proper JWT validation
3. **Rate Limiting**: Implement API rate limiting
4. **Input Validation**: All inputs are validated
5. **SQL Injection**: Use parameterized queries
6. **CORS**: Configure CORS appropriately

### Data Privacy
1. **PII Handling**: Implement proper PII handling
2. **Audit Logging**: Log all sensitive operations
3. **Data Encryption**: Encrypt sensitive data at rest
4. **Access Control**: Implement role-based access control

## Support & Documentation

### Additional Resources
- **Workflow Documentation**: `docs/PREQUALIFICATION_WORKFLOW_README.md`
- **API Documentation**: `docs/API_DOCUMENTATION.md`
- **Swagger Files**: `docs/swagger.json`, `docs/swagger.yaml`

### Getting Help
1. Check the logs for detailed error information
2. Review the workflow documentation
3. Use the Swagger UI for endpoint testing
4. Check Conductor UI for workflow status

---

**Last Updated**: August 15, 2025
**Version**: 1.0.0
**Maintainer**: Loan Service Team
