# Underwriting Worker Service

A comprehensive underwriting worker service built with Go, following clean architecture principles and implementing all underwriting workflow tasks for loan processing.

## 📋 Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Underwriting Tasks](#underwriting-tasks)
- [Workflow Integration](#workflow-integration)
- [API Documentation](#api-documentation)
- [Deployment](#deployment)
- [Monitoring](#monitoring)
- [Contributing](#contributing)

## 🚀 Features

### Core Underwriting Capabilities
- **Comprehensive Credit Assessment** - Multi-bureau credit report analysis with risk scoring
- **Advanced Income Verification** - Multiple verification methods with employment validation
- **Sophisticated Risk Assessment** - Machine learning-based risk scoring with detailed factor analysis
- **Intelligent Decision Engine** - Policy-based automated decisions with manual review triggers
- **Flexible Workflow Integration** - Netflix Conductor-based workflow orchestration

### Technical Features
- **Clean Architecture** - Domain-driven design with clear separation of concerns
- **Microservice Ready** - Container-based deployment with Docker support
- **Scalable Processing** - Horizontal scaling with worker pool management
- **Comprehensive Monitoring** - Prometheus metrics with Grafana dashboards
- **Audit Trail** - Complete audit logging for compliance and tracking

## 🏗 Architecture

The service follows clean architecture principles with the following layers:

```
├── cmd/                    # Application entry points
├── domain/                 # Business logic and entities
│   ├── models.go          # Domain models and entities
│   └── interfaces.go      # Repository and service interfaces
├── application/           # Application business logic
│   ├── usecases/         # Business use cases
│   └── services/         # Application services
├── infrastructure/       # External concerns
│   ├── database/         # Database implementations
│   ├── external/         # External service integrations
│   └── workflow/         # Workflow and task implementations
├── interfaces/           # Interface adapters
│   ├── handlers/         # HTTP/gRPC handlers
│   └── middleware/       # Middleware components
└── pkg/                  # Shared packages
    ├── config/           # Configuration management
    ├── logger/           # Logging utilities
    └── errors/           # Error handling
```

## 📋 Prerequisites

- **Go 1.21+** - Programming language
- **PostgreSQL 15+** - Database
- **Netflix Conductor** - Workflow orchestration
- **Redis** - Caching and session storage
- **Docker & Docker Compose** - Containerization (optional)

## 🚀 Quick Start

### 1. Clone and Build

```bash
# Clone the repository
git clone <repository-url>
cd underwriting-worker

# Install dependencies
go mod tidy

# Build the application
go build -o underwriting-worker cmd/main.go
```

### 2. Configuration

Copy and configure the configuration file:

```bash
cp config/config.yaml.example config/config.yaml
# Edit config/config.yaml with your settings
```

### 3. Environment Variables

Set required environment variables:

```bash
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=underwriting_db
export CONDUCTOR_SERVER_URL=http://localhost:8082
export ENCRYPTION_KEY=your-encryption-key
export JWT_SECRET=your-jwt-secret
```

### 4. Run with Docker Compose (Recommended)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f underwriting-worker

# Stop services
docker-compose down
```

### 5. Run Locally

```bash
# Start the underwriting worker
./underwriting-worker

# Or run with go
go run cmd/main.go
```

## ⚙️ Configuration

### Main Configuration (`config/config.yaml`)

```yaml
application:
  name: "underwriting-worker"
  version: "1.0.0"
  environment: "development"

database:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "password"
  database: "underwriting_db"

conductor:
  server_url: "http://localhost:8082"
  worker_pool_size: 10
  polling_interval_ms: 1000

services:
  credit_bureau:
    provider: "experian"
    base_url: "https://api.experian.com"
    api_key: "${CREDIT_BUREAU_API_KEY}"

  risk_scoring:
    provider: "internal"
    base_url: "http://localhost:8084"
    model_version: "v2.1"
```

### Environment Variable Override

All configuration values can be overridden with environment variables:

- `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `CONDUCTOR_SERVER_URL`
- `CREDIT_BUREAU_API_KEY`, `RISK_SCORING_API_KEY`
- `LOG_LEVEL`, `ENCRYPTION_KEY`, `JWT_SECRET`

## 🔄 Underwriting Tasks

The service implements the following underwriting workflow tasks:

### 1. Credit Check Task (`credit_check`)

**Purpose**: Retrieves and analyzes credit reports from credit bureaus

**Input**:
```json
{
  "applicationId": "app-123",
  "userId": "user-456",
  "creditBureau": "experian"
}
```

**Output**:
```json
{
  "success": true,
  "creditScore": 720,
  "creditScoreRange": "good",
  "creditUtilization": 0.25,
  "riskAnalysis": {
    "riskLevel": "medium",
    "riskFactors": [...]
  }
}
```

**Features**:
- Multi-bureau credit report retrieval
- Credit score range classification
- Credit utilization analysis
- Payment history evaluation
- Derogatory items assessment
- Risk factor identification

### 2. Income Verification Task (`income_verification`)

**Purpose**: Verifies applicant income through multiple methods

**Input**:
```json
{
  "applicationId": "app-123",
  "userId": "user-456",
  "verificationMethod": "automated_verification"
}
```

**Output**:
```json
{
  "success": true,
  "incomeVerification": {
    "verificationStatus": "verified",
    "verifiedAnnualIncome": 75000,
    "employmentStable": true
  },
  "incomeAnalysis": {
    "incomeAdequate": true,
    "verificationScore": 85
  }
}
```

**Features**:
- Multiple verification methods (paystub, tax returns, bank statements)
- Employment history validation
- Income stability analysis
- Variance detection between stated and verified income

### 3. Risk Assessment Task (`risk_assessment`)

**Purpose**: Comprehensive risk analysis using multiple data sources

**Input**:
```json
{
  "applicationId": "app-123",
  "userId": "user-456"
}
```

**Output**:
```json
{
  "success": true,
  "riskAssessment": {
    "overallRiskLevel": "medium",
    "riskScore": 35.5,
    "probabilityOfDefault": 0.08,
    "creditRiskScore": 25,
    "incomeRiskScore": 15,
    "debtRiskScore": 20,
    "fraudRiskScore": 5
  },
  "riskFactors": [...],
  "mitigatingFactors": [...]
}
```

**Features**:
- Multi-dimensional risk scoring
- Credit, income, debt, and fraud risk analysis
- Risk factor identification and weighting
- Mitigating factor consideration
- Probability of default calculation

### 4. Underwriting Decision Task (`underwriting_decision`)

**Purpose**: Makes final underwriting decision based on all assessment data

**Input**:
```json
{
  "applicationId": "app-123",
  "userId": "user-456"
}
```

**Output**:
```json
{
  "success": true,
  "underwritingResult": {
    "decision": "approved",
    "approvedAmount": 25000,
    "interestRate": 8.5,
    "apr": 9.0,
    "monthlyPayment": 456.78,
    "automatedDecision": true
  },
  "conditions": [...],
  "decisionReasons": [...]
}
```

**Features**:
- Policy-based decision making
- Interest rate calculation
- Conditional approval handling
- Counter-offer generation
- Manual review triggers

### 5. Application State Update Task (`update_application_state`)

**Purpose**: Updates loan application state throughout the workflow

**Input**:
```json
{
  "applicationId": "app-123",
  "newState": "underwriting_completed",
  "reason": "Automated underwriting process completed"
}
```

**Output**:
```json
{
  "success": true,
  "stateTransition": {
    "previousState": "underwriting_in_progress",
    "currentState": "underwriting_completed"
  }
}
```

### Additional Specialized Tasks

- **Policy Compliance Check** (`policy_compliance_check`)
- **Fraud Detection** (`fraud_detection`)
- **Interest Rate Calculation** (`calculate_interest_rate`)
- **Final Approval Processing** (`final_approval`)
- **Denial Processing** (`process_denial`)
- **Manual Review Assignment** (`assign_manual_review`)
- **Conditional Approval** (`process_conditional_approval`)
- **Counter Offer Generation** (`generate_counter_offer`)

## 🔄 Workflow Integration

### Netflix Conductor Integration

The service integrates with Netflix Conductor for workflow orchestration:

```json
{
  "name": "underwriting_workflow",
  "description": "Complete underwriting workflow",
  "tasks": [
    {
      "name": "credit_check",
      "taskReferenceName": "credit_check_task",
      "type": "SIMPLE"
    },
    {
      "name": "income_verification",
      "taskReferenceName": "income_verification_task",
      "type": "SIMPLE"
    },
    {
      "name": "risk_assessment",
      "taskReferenceName": "risk_assessment_task",
      "type": "SIMPLE"
    },
    {
      "name": "underwriting_decision",
      "taskReferenceName": "decision_task",
      "type": "SIMPLE"
    }
  ]
}
```

### Workflow Execution

1. **Start Workflow**: Triggered by loan application submission
2. **Parallel Processing**: Credit check and income verification run in parallel
3. **Risk Assessment**: Uses results from previous steps
4. **Decision Making**: Final underwriting decision based on all data
5. **State Updates**: Application state updated throughout process

### Error Handling

- **Retry Logic**: Automatic retry for transient failures
- **Circuit Breaker**: Prevents cascade failures
- **Fallback Processing**: Manual review for system failures
- **Audit Trail**: Complete logging of all decisions and errors

## 📊 Monitoring

### Metrics

The service exposes Prometheus metrics:

- **Task Execution Metrics**
  - `underwriting_task_duration_seconds` - Task execution time
  - `underwriting_task_total` - Total tasks processed
  - `underwriting_task_errors_total` - Task error count

- **Business Metrics**
  - `underwriting_decisions_total` - Decisions by type
  - `underwriting_approval_rate` - Approval rate percentage
  - `underwriting_processing_time` - End-to-end processing time

- **System Metrics**
  - `conductor_worker_pool_utilization` - Worker pool usage
  - `database_connection_pool_size` - DB connection usage

### Health Checks

- **Application Health**: `/health` endpoint
- **Database Connectivity**: Validates database connection
- **Conductor Connectivity**: Validates workflow engine connection
- **External Services**: Validates third-party service availability

### Logging

Structured JSON logging with the following levels:
- **INFO**: Normal operation events
- **WARN**: Recoverable errors and unusual conditions
- **ERROR**: Error conditions requiring attention
- **DEBUG**: Detailed debugging information

Log fields include:
- `timestamp`, `level`, `message`
- `application_id`, `user_id`, `task_name`
- `processing_time`, `operation`, `status`

## 🚀 Deployment

### Docker Deployment

```bash
# Build image
docker build -t underwriting-worker:latest .

# Run with Docker Compose
docker-compose up -d
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: underwriting-worker
spec:
  replicas: 3
  selector:
    matchLabels:
      app: underwriting-worker
  template:
    metadata:
      labels:
        app: underwriting-worker
    spec:
      containers:
      - name: underwriting-worker
        image: underwriting-worker:latest
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: CONDUCTOR_SERVER_URL
          value: "http://conductor-service:8080"
```

### Environment-Specific Configurations

- **Development**: Single instance, verbose logging
- **Staging**: Multi-instance, production-like setup
- **Production**: High availability, monitoring, security hardening

## 📚 API Documentation

### Task Input/Output Schemas

Detailed schemas for all task inputs and outputs are available in:
- `docs/api/task-schemas.md`
- `docs/api/workflow-definitions.md`
- `docs/api/error-responses.md`

### Integration Examples

Example integrations available in:
- `examples/workflow-definitions/`
- `examples/task-testing/`
- `examples/api-clients/`

## 🧪 Testing

### Running Tests

```bash
# Unit tests
go test ./... -v

# Integration tests
go test ./... -tags=integration -v

# Load tests
go test ./... -tags=load -v
```

### Test Coverage

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 🔧 Development

### Local Development Setup

1. **Start Dependencies**:
   ```bash
   docker-compose -f docker-compose.dev.yml up -d
   ```

2. **Run Application**:
   ```bash
   go run cmd/main.go
   ```

3. **Live Reload** (with Air):
   ```bash
   air -c .air.toml
   ```

### Code Quality

- **Linting**: `golangci-lint run`
- **Formatting**: `gofmt -s -w .`
- **Security**: `gosec ./...`
- **Dependencies**: `go mod tidy && go mod verify`

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow clean architecture principles
- Write comprehensive tests
- Include documentation updates
- Ensure backward compatibility
- Add appropriate logging and monitoring

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](issues)
- **Discussions**: [GitHub Discussions](discussions)

## 🔄 Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of changes and version history.

---

## Quick Reference

### Common Commands

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f underwriting-worker

# Run tests
go test ./... -v

# Build application
go build -o underwriting-worker cmd/main.go

# Check configuration
./underwriting-worker -config-check
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `CONDUCTOR_SERVER_URL` | Conductor URL | `http://localhost:8082` |
| `LOG_LEVEL` | Log level | `info` |
| `WORKER_POOL_SIZE` | Worker pool size | `10` |

### Task Types

| Task | Purpose | Complexity |
|------|---------|------------|
| `credit_check` | Credit analysis | Medium |
| `income_verification` | Income validation | Medium |
| `risk_assessment` | Risk scoring | High |
| `underwriting_decision` | Final decision | High |
| `update_application_state` | State management | Low |

---

**Built with ❤️ using Go and clean architecture principles**
