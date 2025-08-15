# Authentication Service

A comprehensive authentication service for the Loan Origination System (LOS) built with Go, Gin framework, and PostgreSQL.

## Features

- **JWT Authentication**: Secure access tokens with refresh token rotation
- **HTTP Signature Validation**: RFC-compliant HTTP signature authentication
- **Role-Based Access Control (RBAC)**: Granular permissions system
- **Rate Limiting**: IP and user-based rate limiting
- **Session Management**: Secure session handling with Redis
- **Audit Logging**: Comprehensive authentication event logging
- **Clean Architecture**: Domain-driven design with clear separation of concerns
- **Internationalization**: Multi-language error messages
- **Structured Logging**: Zap-based structured logging

## Architecture

The service follows clean architecture principles:

```
├── domain/          # Business logic and entities
├── application/     # Use cases and business rules
├── infrastructure/  # Data access and external services
├── interfaces/      # HTTP handlers and middleware
├── cmd/            # Application entry point
└── migrations/     # Database schema migrations
```

## API Endpoints

### Authentication
- `POST /v1/auth/login` - User login
- `POST /v1/auth/refresh` - Refresh access token
- `POST /v1/auth/logout` - User logout
- `POST /v1/auth/logout-all` - Logout from all devices
- `GET /v1/auth/me` - Get current user profile
- `GET /v1/auth/health` - Health check

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Redis 6+

### Quick Setup with Environment File

1. Copy the configuration template:
```bash
cp config.env .env
```

2. Edit `.env` with your database and Redis credentials

3. Source the environment and run:
```bash
source .env && go run cmd/main.go
```

### Environment Variables

You can use the provided `config.env` file as a template:

1. Copy the example configuration:
```bash
cp config.env .env
```

2. Edit `.env` with your actual values:
```bash
# Server Configuration
PORT=8080
READ_TIMEOUT=10s
WRITE_TIMEOUT=10s
IDLE_TIMEOUT=60s

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=los_auth
DB_USER=postgres
DB_PASSWORD=your_actual_password
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_SSL_MODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SIGNING_KEY=your-secret-key-change-in-production
JWT_ISSUER=los-auth-service
JWT_TTL=15m

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
```

3. Source the environment file before running:
```bash
source .env
```

**Note**: Make sure to change the default values, especially:
- `DB_PASSWORD` - Use a strong database password
- `JWT_SIGNING_KEY` - Use a secure, random key in production
- `DB_HOST` and `REDIS_HOST` - Update if using different hosts

### Database Setup

1. Create the database:
```sql
CREATE DATABASE los_auth;
```

2. Run migrations:
```bash
psql -h localhost -U postgres -d los_auth -f migrations/001_init.sql
```

### Running the Service

1. Install dependencies:
```bash
go mod download
```

2. Run the service:
```bash
go run cmd/main.go
```

3. The service will start on port 8080 (configurable via PORT env var)

### Docker Compose (Development)

```yaml
version: '3.8'
services:
  auth-service:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: los_auth
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

## Testing

### Run Tests
```bash
go test ./...
```

### Run Tests with Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Benchmarks
```bash
go test -bench=. ./...
```

## API Usage Examples

### Login
```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@lendingplatform.com",
    "password": "admin123!"
  }'
```

### Access Protected Endpoint
```bash
curl -X GET http://localhost:8080/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Refresh Token
```bash
curl -X POST http://localhost:8080/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

## Security Features

### Password Security
- bcrypt hashing with cost factor 12
- Password complexity requirements enforced at application level

### Token Security
- JWT tokens with 15-minute expiration
- Refresh tokens with 30-day expiration
- Automatic token rotation on refresh
- Token revocation support

### Rate Limiting
- 100 requests per hour per IP for authentication endpoints
- Account lockout after 5 failed login attempts
- 15-minute lockout duration

### HTTP Signature Authentication
- RFC-compliant HTTP signature validation
- HMAC-SHA256 signature algorithm
- Clock skew tolerance (5 minutes)
- Key rotation support

## Monitoring and Observability

### Health Checks
- `/health` endpoint for service health
- Database connectivity checks
- Redis connectivity checks

### Logging
- Structured JSON logging with Zap
- Authentication events logging
- Security events logging
- Request/response logging

### Metrics
- Authentication success/failure rates
- Token refresh rates
- Rate limiting events
- Session creation/cleanup

## Error Handling

The service uses standardized error codes:

- `AUTH_001` - Invalid credentials
- `AUTH_002` - Account locked
- `AUTH_003` - Account disabled
- `AUTH_004` - Invalid token
- `AUTH_005` - Token expired
- `AUTH_010` - Rate limit exceeded
- `AUTH_015` - Insufficient permissions

## Development

### Adding New Permissions

1. Define permission in `domain/models.go`:
```go
const PermissionNewFeature Permission = "feature:new_action"
```

2. Add to role permissions in `GetPermissions()` method

3. Use in middleware:
```go
router.Use(authMiddleware.RequirePermission(domain.PermissionNewFeature))
```

### Adding New Error Codes

1. Define constant in `domain/interfaces.go`:
```go
const AUTH_XXX = "AUTH_XXX" // Description
```

2. Add localized message in i18n files

3. Use in service methods:
```go
return domain.NewAuthError(domain.AUTH_XXX, "Message", "Description")
```

## Production Deployment

### Security Checklist
- [ ] Change default JWT signing key
- [ ] Use strong database passwords
- [ ] Enable TLS/SSL for all connections
- [ ] Configure firewall rules
- [ ] Set up log aggregation
- [ ] Configure monitoring and alerting
- [ ] Enable rate limiting
- [ ] Set appropriate CORS policies

### Performance Tuning
- Configure database connection pooling
- Tune Redis memory settings
- Set appropriate JWT token expiration
- Configure Go garbage collection
- Use connection pooling for HTTP clients

## Contributing

1. Follow Go coding standards
2. Write tests for new features
3. Update documentation
4. Use conventional commits
5. Ensure all tests pass

## License

Copyright (c) 2025 Lending Platform. All rights reserved.
