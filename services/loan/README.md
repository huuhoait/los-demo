# Loan Service

A comprehensive loan origination service with internationalization support for Vietnamese and English, featuring Netflix Conductor workflow integration.

## Features

- 🌍 **Internationalization**: Full Vietnamese and English language support
- ⚡ **Workflow Engine**: Netflix Conductor integration for loan processing workflows
- 🏛️ **Clean Architecture**: Domain-driven design with clear separation of concerns
- 📊 **State Management**: Robust state machine for loan application lifecycle
- 🔒 **Validation**: Comprehensive request validation with localized error messages
- 📈 **Pre-qualification**: Real-time loan pre-qualification with DTI calculations
- 💰 **Offer Engine**: Automated loan offer generation with interest rate calculations
- 🏗️ **Microservice Ready**: Designed for microservice architecture

## Architecture

```
loan-service/
├── cmd/                    # Application entry points
├── domain/                 # Domain models and business rules
├── application/           # Application services and use cases
├── infrastructure/        # External integrations (Conductor, DB)
├── interfaces/           # HTTP handlers and middleware
├── pkg/i18n/             # Internationalization package
└── i18n/                 # Translation files (en.toml, vi.toml)
```

## API Endpoints

### Public Endpoints
- `GET /v1/loans/health` - Health check

### Authenticated Endpoints

#### Application Management
- `POST /v1/loans/applications` - Create new loan application
- `GET /v1/loans/applications` - Get user's applications
- `GET /v1/loans/applications/:id` - Get specific application
- `PUT /v1/loans/applications/:id` - Update draft application
- `POST /v1/loans/applications/:id/submit` - Submit application for processing

#### Pre-qualification
- `POST /v1/loans/prequalify` - Check loan pre-qualification

#### Offers
- `POST /v1/loans/applications/:id/offer` - Generate loan offer
- `POST /v1/loans/applications/:id/accept-offer` - Accept loan offer

#### Admin Endpoints
- `POST /v1/loans/applications/:id/transition` - Manual state transition
- `GET /v1/loans/stats` - Application statistics

## Internationalization

### Language Detection
The service automatically detects language from:
1. Query parameter: `?lang=vi` or `?lang=en`
2. HTTP header: `X-Language: vi`
3. Accept-Language header
4. Default fallback to English

### Supported Languages
- **English (en)**: Default language
- **Vietnamese (vi)**: Full translation support

### Example API Calls

```bash
# English (default)
curl -X POST http://localhost:8080/v1/loans/applications \
  -H "Content-Type: application/json" \
  -d '{"loan_amount": 25000, "loan_purpose": "debt_consolidation"}'

# Vietnamese
curl -X POST http://localhost:8080/v1/loans/applications \
  -H "Content-Type: application/json" \
  -H "X-Language: vi" \
  -d '{"loan_amount": 25000, "loan_purpose": "debt_consolidation"}'
```

## Netflix Conductor Integration

### Workflow Types

1. **Loan Processing Workflow** (`loan_processing_workflow`)
   - Main workflow for end-to-end loan processing
   - Handles state transitions and task orchestration

2. **Pre-qualification Workflow** (`prequalification_workflow`)
   - Quick pre-qualification assessment
   - Credit checks and initial validation

3. **Underwriting Workflow** (`underwriting_workflow`)
   - Detailed underwriting process
   - Risk assessment and decision making

### Workflow Features
- Automatic workflow initiation on application creation
- State transition triggers
- Task status monitoring
- Error handling and recovery

## Domain Models

### Application States
- `initiated` - Application created
- `pre_qualified` - Initial qualification completed
- `documents_submitted` - Required documents uploaded
- `identity_verified` - Identity verification completed
- `underwriting` - Under underwriting review
- `manual_review` - Requires manual review
- `approved` - Application approved
- `denied` - Application denied
- `documents_signed` - Loan documents signed
- `funded` - Loan funded
- `active` - Loan is active

### Loan Purposes
- `debt_consolidation` - Debt consolidation
- `home_improvement` - Home improvement
- `medical` - Medical expenses
- `vacation` - Vacation
- `wedding` - Wedding
- `major_purchase` - Major purchase
- `other` - Other purposes

## Running the Service

### Prerequisites
- Go 1.21 or later
- Netflix Conductor (optional for full workflow functionality)

### Development Mode
```bash
# Clone and navigate to the service
cd /path/to/loan-service

# Download dependencies
go mod download

# Run the service
go run cmd/main.go
```

### Environment Variables
```bash
# Server configuration
PORT=8080
HOST=0.0.0.0
READ_TIMEOUT=30
WRITE_TIMEOUT=30

# Database configuration (for production)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=loan_service
DB_SSLMODE=disable

# Conductor configuration
CONDUCTOR_BASE_URL=http://localhost:8080
CONDUCTOR_TIMEOUT=30

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Docker Support
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o loan-service cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/loan-service .
CMD ["./loan-service"]
```

## Example Usage

### Create Application
```bash
curl -X POST http://localhost:8080/v1/loans/applications \
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

### Pre-qualification Check
```bash
curl -X POST http://localhost:8080/v1/loans/prequalify \
  -H "Content-Type: application/json" \
  -H "X-Language: en" \
  -d '{
    "loan_amount": 30000,
    "annual_income": 80000,
    "monthly_debt_payments": 1200,
    "employment_status": "full_time"
  }'
```

### Submit Application
```bash
curl -X POST http://localhost:8080/v1/loans/applications/123e4567-e89b-12d3-a456-426614174000/submit \
  -H "Content-Type: application/json"
```

## Response Format

### Success Response
```json
{
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "loan_amount": 25000,
    "current_state": "initiated",
    "status": "draft"
  },
  "message": "Đơn xin vay đã được tạo thành công",
  "metadata": {
    "request_id": "req_1692032400",
    "timestamp": "2025-08-14T12:00:00Z",
    "version": "v1",
    "service": "loan-service"
  }
}
```

### Error Response
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "LOAN_001",
    "message": "Số tiền vay không hợp lệ",
    "metadata": {
      "min_amount": 5000,
      "max_amount": 50000
    }
  },
  "metadata": {
    "request_id": "req_1692032400",
    "timestamp": "2025-08-14T12:00:00Z",
    "version": "v1",
    "service": "loan-service"
  }
}
```

## Testing

### Unit Tests
```bash
go test ./...
```

### Integration Tests
```bash
go test -tags=integration ./...
```

### Load Testing
```bash
# Using vegeta
echo "POST http://localhost:8080/v1/loans/prequalify" | \
vegeta attack -rate=100 -duration=30s | \
vegeta report
```

## Monitoring

### Health Check
```bash
curl http://localhost:8080/v1/loans/health
```

### Metrics Endpoints
- Application statistics: `GET /v1/loans/stats`
- Workflow status monitoring via Conductor UI

## Production Considerations

1. **Database**: Replace MockRepository with actual PostgreSQL implementation
2. **Authentication**: Implement proper JWT authentication middleware
3. **Conductor**: Deploy Netflix Conductor for workflow management
4. **Monitoring**: Add Prometheus metrics and distributed tracing
5. **Logging**: Configure centralized logging with ELK stack
6. **Security**: Implement rate limiting and input sanitization
7. **Caching**: Add Redis for performance optimization

## Contributing

1. Follow Go coding standards
2. Add tests for new features
3. Update translation files for new messages
4. Document API changes
5. Ensure backwards compatibility

## License

This project is part of the BMAD loan origination system.
