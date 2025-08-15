# Loan Service API Documentation

## Overview

The Loan Service API provides a comprehensive loan origination system with internationalization support for Vietnamese and English, featuring Netflix Conductor workflow integration.

## API Information

- **Title**: Loan Service API
- **Version**: 1.0
- **Base URL**: `http://localhost:8081/v1`
- **Description**: A comprehensive loan origination service with internationalization support for Vietnamese and English, featuring Netflix Conductor workflow integration.

## Authentication

The API uses Bearer token authentication. Include the JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Internationalization

The API supports multiple languages through the `X-Language` header:

- `en` - English (default)
- `vi` - Vietnamese

Example:
```
X-Language: vi
```

## API Endpoints

### Health Check

#### GET /health
Check the health status of the loan service.

**Response:**
```json
{
  "success": true,
  "data": {
    "service": "loan-service",
    "status": "healthy",
    "timestamp": {
      "unix": {
        "seconds": {
          "value": "1692032400"
        }
      }
    },
    "version": "v1.0.0"
  },
  "metadata": {
    "request_id": "",
    "service": "loan-service",
    "timestamp": "2025-08-14T07:52:25Z",
    "version": "v1"
  }
}
```

### Applications

#### POST /loans/applications
Create a new loan application.

**Request Body:**
```json
{
  "loan_amount": 25000,
  "loan_purpose": "debt_consolidation",
  "requested_term_months": 60,
  "annual_income": 75000,
  "monthly_income": 6250,
  "employment_status": "full_time",
  "monthly_debt_payments": 1500
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "user_id": "user123",
    "application_number": "LOAN123456",
    "loan_amount": 25000,
    "loan_purpose": "debt_consolidation",
    "requested_term_months": 60,
    "annual_income": 75000,
    "monthly_income": 6250,
    "employment_status": "full_time",
    "monthly_debt_payments": 1500,
    "current_state": "initiated",
    "status": "draft",
    "created_at": "2025-08-14T12:00:00Z",
    "updated_at": "2025-08-14T12:00:00Z"
  },
  "message": "Application created successfully",
  "metadata": {
    "request_id": "",
    "service": "loan-service",
    "timestamp": "2025-08-14T07:52:25Z",
    "version": "v1"
  }
}
```

#### GET /loans/applications
Get all loan applications for the current user.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "user_id": "user123",
      "application_number": "LOAN123456",
      "loan_amount": 25000,
      "loan_purpose": "debt_consolidation",
      "requested_term_months": 60,
      "annual_income": 75000,
      "monthly_income": 6250,
      "employment_status": "full_time",
      "monthly_debt_payments": 1500,
      "current_state": "initiated",
      "status": "draft",
      "created_at": "2025-08-14T12:00:00Z",
      "updated_at": "2025-08-14T12:00:00Z"
    }
  ],
  "metadata": {
    "request_id": "",
    "service": "loan-service",
    "timestamp": "2025-08-14T07:52:25Z",
    "version": "v1"
  }
}
```

#### GET /loans/applications/{id}
Get a specific loan application by ID.

**Parameters:**
- `id` (path, required): Application ID

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "user_id": "user123",
    "application_number": "LOAN123456",
    "loan_amount": 25000,
    "loan_purpose": "debt_consolidation",
    "requested_term_months": 60,
    "annual_income": 75000,
    "monthly_income": 6250,
    "employment_status": "full_time",
    "monthly_debt_payments": 1500,
    "current_state": "initiated",
    "status": "draft",
    "created_at": "2025-08-14T12:00:00Z",
    "updated_at": "2025-08-14T12:00:00Z"
  },
  "metadata": {
    "request_id": "",
    "service": "loan-service",
    "timestamp": "2025-08-14T07:52:25Z",
    "version": "v1"
  }
}
```

### Pre-qualification

#### POST /loans/prequalify
Perform loan pre-qualification.

**Request Body:**
```json
{
  "loan_amount": 25000,
  "annual_income": 75000,
  "monthly_debt_payments": 1500,
  "employment_status": "full_time"
}
```

**Response:**
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
  },
  "message": "Pre-qualification completed",
  "metadata": {
    "request_id": "",
    "service": "loan-service",
    "timestamp": "2025-08-14T07:52:25Z",
    "version": "v1"
  }
}
```

### Admin

#### GET /loans/stats
Get application statistics (admin only).

**Query Parameters:**
- `status` (optional): Filter by status
- `state` (optional): Filter by state
- `days` (optional): Number of days to look back (default: 30)

**Response:**
```json
{
  "success": true,
  "data": {
    "total_applications": 150,
    "pending_review": 25,
    "approved": 80,
    "denied": 30,
    "funded": 15,
    "period_days": 30,
    "filters": {
      "status": "",
      "state": ""
    }
  },
  "metadata": {
    "request_id": "",
    "service": "loan-service",
    "timestamp": "2025-08-14T07:52:25Z",
    "version": "v1"
  }
}
```

## Data Models

### LoanApplication
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
  "risk_score": "integer (optional)",
  "workflow_id": "string (optional)",
  "correlation_id": "string (optional)",
  "created_at": "string (date-time)",
  "updated_at": "string (date-time)"
}
```

### CreateApplicationRequest
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

### PreQualifyRequest
```json
{
  "loan_amount": "number (5000-50000)",
  "annual_income": "number (>=0)",
  "monthly_debt_payments": "number (>=0)",
  "employment_status": "string"
}
```

### PreQualifyResult
```json
{
  "qualified": "boolean",
  "max_loan_amount": "number",
  "min_interest_rate": "number",
  "max_interest_rate": "number",
  "recommended_terms": ["integer"],
  "dti_ratio": "number",
  "message": "string"
}
```

## Enums

### LoanPurpose
- `debt_consolidation`
- `home_improvement`
- `medical`
- `vacation`
- `wedding`
- `major_purchase`
- `other`

### EmploymentStatus
- `full_time`
- `part_time`
- `self_employed`
- `retired`
- `unemployed`
- `student`

### ApplicationState
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

### ApplicationStatus
- `draft`
- `submitted`
- `under_review`
- `approved`
- `denied`
- `funded`
- `active`
- `closed`

### ResidenceType
- `own`
- `rent`
- `family`
- `other`

### AccountType
- `checking`
- `savings`

## User Creation

The loan application creation endpoint now includes comprehensive user information collection:

### Features
- **Automatic User Creation**: New users are automatically created when submitting loan applications
- **Existing User Detection**: If a user with the same email already exists, their existing profile is used
- **Complete User Profile**: Collects personal, address, employment, and banking information
- **Data Validation**: Comprehensive validation of all user fields including age verification (18+)
- **Privacy Protection**: Sensitive data like SSN and account numbers are masked in responses

### User Information Required
- **Personal Details**: First name, last name, email, phone, date of birth, SSN
- **Address**: Complete address with residence type and time at address
- **Employment**: Employer details, job title, tenure, work contact information
- **Banking**: Bank details for loan disbursement

### Example User Creation Flow
1. User submits loan application with complete user information
2. System validates all user data
3. System checks if user exists by email
4. If new user: creates user profile and generates user ID
5. If existing user: uses existing user ID
6. Creates loan application linked to user ID
7. Returns application with user information (sensitive data masked)

## Error Responses

All error responses follow this format:

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "LOAN_001",
    "message": "Invalid loan amount",
    "description": "The loan amount must be between 5000 and 50000",
    "field": "loan_amount",
    "metadata": {
      "min_amount": 5000,
      "max_amount": 50000
    }
  },
  "metadata": {
    "request_id": "",
    "service": "loan-service",
    "timestamp": "2025-08-14T07:52:25Z",
    "version": "v1"
  }
}
```

## Error Codes

- `LOAN_001` - Invalid loan amount
- `LOAN_002` - Invalid loan purpose
- `LOAN_003` - Invalid loan term
- `LOAN_004` - Invalid income information
- `LOAN_005` - Loan amount below minimum
- `LOAN_006` - Loan amount above maximum
- `LOAN_007` - Insufficient income
- `LOAN_020` - Invalid request format
- `LOAN_022` - Unauthorized access
- `LOAN_023` - Database error

## Testing the API

### Using curl

```bash
# Health check
curl -X GET http://localhost:8081/v1/health

# Pre-qualification
curl -X POST http://localhost:8081/v1/loans/prequalify \
  -H "Content-Type: application/json" \
  -H "X-Language: en" \
  -d '{
    "loan_amount": 25000,
    "annual_income": 75000,
    "monthly_debt_payments": 1500,
    "employment_status": "full_time"
  }'

# Create application
curl -X POST http://localhost:8081/v1/loans/applications \
  -H "Content-Type: application/json" \
  -H "X-Language: vi" \
  -d '{
    "loan_amount": 25000,
    "loan_purpose": "debt_consolidation",
    "requested_term_months": 60,
    "annual_income": 75000,
    "monthly_income": 6250,
    "employment_status": "full_time",
    "monthly_debt_payments": 1500
  }'
```

### Using Swagger UI

1. Open `swagger-ui.html` in your browser
2. The Swagger UI will load the API documentation from `http://localhost:8081/swagger/doc.json`
3. You can test all endpoints directly from the UI

## Rate Limiting

Currently, no rate limiting is implemented. In production, consider implementing rate limiting based on user/IP.

## Security Considerations

1. Always use HTTPS in production
2. Implement proper JWT token validation
3. Validate all input data
4. Use environment variables for sensitive configuration
5. Implement proper logging and monitoring

## Support

For API support, contact:
- Email: support@swagger.io
- URL: http://www.swagger.io/support
