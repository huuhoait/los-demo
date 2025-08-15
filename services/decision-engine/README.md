# Decision Engine Service

A high-performance, clean architecture decision engine service for loan processing and risk assessment.

## Overview

The Decision Engine Service is a microservice that provides intelligent loan decision-making capabilities using comprehensive risk assessment algorithms. Built with clean architecture principles, it ensures maintainability, testability, and scalability.

## Architecture

The service follows clean architecture with clear separation of concerns:

```
decision-engine/
├── domain/           # Business entities and interfaces
├── application/      # Business logic and use cases
├── infrastructure/   # External integrations (database, APIs)
├── interfaces/       # HTTP handlers and API controllers
└── main.go          # Application entry point
```

### Key Components

- **Domain Layer**: Core business models, validation rules, and interfaces
- **Application Layer**: Decision-making logic, risk assessment, and business rules
- **Infrastructure Layer**: Database repositories and external service integrations
- **Interface Layer**: REST API endpoints and HTTP request handling

## Features

### Core Decision Making
- Comprehensive loan application evaluation
- Risk assessment with multiple scoring factors
- Interest rate calculation based on risk profile
- Configurable decision rules and thresholds

### Risk Assessment
- Credit score analysis and categorization
- Debt-to-income ratio calculations
- Employment stability evaluation
- Payment history assessment
- Risk factor identification and mitigation analysis

### Data Management
- Persistent decision storage with audit trail
- Decision history tracking per customer
- Statistical reporting and analytics
- Request validation and error handling

### API Endpoints

#### Decision Management
- `POST /api/v1/decisions` - Make loan decision
- `GET /api/v1/decisions/:applicationId` - Get specific decision
- `POST /api/v1/decisions/validate` - Validate decision request
- `GET /api/v1/decisions/rules` - Get decision rules
- `GET /api/v1/decisions/statistics` - Get decision statistics

#### Customer Management
- `GET /api/v1/customers/:customerId/decisions` - Get customer decision history

#### System
- `GET /health` - Health check endpoint

## Configuration

Environment variables:

```bash
PORT=8080                    # Server port
DATABASE_URL=postgres://...  # PostgreSQL connection string
LOG_LEVEL=info              # Logging level (debug, info, warn, error)
ENVIRONMENT=development     # Environment (development, production)

# Credit Bureau Configuration
EXPERIAN_ENDPOINT=https://api.experian.com
EQUIFAX_ENDPOINT=https://api.equifax.com
TRANSUNION_ENDPOINT=https://api.transunion.com
```

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 12+
- Git

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd decision-engine
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up environment variables:
```bash
export DATABASE_URL="postgres://user:password@localhost/decision_engine?sslmode=disable"
export PORT=8080
export LOG_LEVEL=info
export ENVIRONMENT=development
```

4. Run the service:
```bash
go run main.go
```

The service will start on `http://localhost:8080`

### Using Docker

1. Build the Docker image:
```bash
docker build -t decision-engine .
```

2. Run with Docker Compose:
```bash
docker-compose up -d
```

## API Usage Examples

### Make a Decision

```bash
curl -X POST http://localhost:8080/api/v1/decisions \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APP-001",
    "customer_id": "CUST-001",
    "loan_amount": 15000,
    "loan_purpose": "DEBT_CONSOLIDATION",
    "loan_term_months": 36,
    "annual_income": 75000,
    "monthly_income": 6250,
    "credit_score": 720,
    "employment_type": "FULL_TIME"
  }'
```

### Get Decision

```bash
curl http://localhost:8080/api/v1/decisions/APP-001
```

### Get Customer History

```bash
curl http://localhost:8080/api/v1/customers/CUST-001/decisions?limit=5
```

### Health Check

```bash
curl http://localhost:8080/health
```

## Decision Logic

### Risk Assessment Categories

1. **Credit Risk** (35% weight)
   - Credit score evaluation
   - Payment history analysis
   - Credit utilization assessment

2. **Debt Risk** (25% weight)
   - Debt-to-income ratio calculation
   - Existing debt obligations
   - Monthly payment capacity

3. **Income Risk** (20% weight)
   - Income stability assessment
   - Loan-to-income ratio analysis
   - Income verification consistency

4. **Employment Risk** (15% weight)
   - Employment type evaluation
   - Job stability indicators
   - Income source reliability

5. **Collateral Risk** (5% weight)
   - Loan security assessment
   - Asset valuation (if applicable)

### Decision Outcomes

- **APPROVED**: Low risk, standard terms
- **CONDITIONAL**: Medium risk, modified terms
- **DECLINED**: High risk, unacceptable risk profile

### Interest Rate Calculation

Interest rates are calculated based on:
- Base rate + risk premium
- Credit score adjustments
- Market conditions
- Loan term considerations

## Database Schema

### decision_requests
- Stores original loan application data
- Indexes on customer_id, application_id, created_at

### decisions  
- Stores decision outcomes and analysis
- JSON fields for risk assessment and applied rules
- Foreign key relationship to decision_requests

## Monitoring and Observability

### Logging
- Structured logging with Zap
- Request/response logging
- Error tracking with context
- Performance metrics

### Health Checks
- Database connectivity validation
- Service health status
- Dependency health monitoring

### Metrics
- Decision processing times
- Approval/decline rates
- Risk score distributions
- API response times

## Development

### Code Structure
- Clean architecture separation
- Dependency injection
- Interface-based design
- Comprehensive error handling

### Testing Strategy
```bash
# Run unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...
```

### Code Quality
- Go fmt for formatting
- Go vet for static analysis
- Golint for style checking
- Race condition detection

## Deployment

### Production Checklist
- [ ] Database migrations applied
- [ ] Environment variables configured
- [ ] SSL/TLS certificates installed
- [ ] Monitoring and alerting configured
- [ ] Load balancer configured
- [ ] Backup strategy implemented

### Scaling Considerations
- Stateless service design
- Database connection pooling
- Horizontal scaling support
- Caching strategies for performance

## Security

- Input validation and sanitization
- SQL injection prevention
- Sensitive data masking in logs
- Rate limiting and DDoS protection
- Authentication and authorization (to be implemented)

## Contributing

1. Follow clean architecture principles
2. Write comprehensive tests
3. Update documentation
4. Follow Go coding standards
5. Use conventional commit messages

## Support

For questions and support:
- Create GitHub issues for bugs
- Submit feature requests via issues
- Check documentation for API details
- Review code comments for implementation details

## License

[MIT License](LICENSE)
